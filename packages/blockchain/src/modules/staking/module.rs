use crate::types::{Address, Balance, BlockHeight};

use super::state::StakingState;
use super::types::{
    StakingError, StakingEvent, UnbondingEntry, ValidatorInfo, MINIMUM_DELEGATION,
    MINIMUM_VALIDATOR_STAKE, UNBONDING_PERIOD_BLOCKS,
};

/// Staking module providing delegation, unbonding, rewards, and validator management.
pub struct StakingModule {
    pub state: StakingState,
}

impl StakingModule {
    pub fn new(state: StakingState) -> Self {
        Self { state }
    }

    /// Register a new validator.
    pub fn register_validator(
        &mut self,
        address: Address,
        commission_rate: u32,
    ) -> Result<(), StakingError> {
        if self.state.validators.contains_key(&address) {
            return Err(StakingError::ValidatorAlreadyExists);
        }
        self.state.validators.insert(
            address,
            ValidatorInfo {
                address,
                total_stake: 0,
                cu_contribution: 0,
                commission_rate,
                active: true,
            },
        );
        self.state
            .events
            .push(StakingEvent::ValidatorAdded { address });
        Ok(())
    }

    /// Delegate tokens to a validator.
    pub fn delegate(
        &mut self,
        delegator: Address,
        validator: Address,
        amount: Balance,
    ) -> Result<(), StakingError> {
        if amount < MINIMUM_DELEGATION {
            return Err(StakingError::MinimumDelegationNotMet(MINIMUM_DELEGATION));
        }
        let val = self
            .state
            .validators
            .get_mut(&validator)
            .ok_or(StakingError::ValidatorNotFound)?;
        val.total_stake = val
            .total_stake
            .checked_add(amount)
            .ok_or(StakingError::Overflow)?;
        let delegation = self
            .state
            .delegations
            .entry((delegator, validator))
            .or_insert(0);
        *delegation = delegation
            .checked_add(amount)
            .ok_or(StakingError::Overflow)?;
        self.state.total_staked = self
            .state
            .total_staked
            .checked_add(amount)
            .ok_or(StakingError::Overflow)?;
        self.state.events.push(StakingEvent::Delegated {
            delegator,
            validator,
            amount,
        });
        Ok(())
    }

    /// Undelegate tokens from a validator (starts unbonding period).
    pub fn undelegate(
        &mut self,
        delegator: Address,
        validator: Address,
        amount: Balance,
        current_height: BlockHeight,
    ) -> Result<(), StakingError> {
        let delegation = self
            .state
            .delegations
            .get_mut(&(delegator, validator))
            .ok_or(StakingError::NoDelegationFound)?;
        if *delegation < amount {
            return Err(StakingError::InsufficientBalance);
        }
        *delegation = delegation
            .checked_sub(amount)
            .ok_or(StakingError::Overflow)?;
        let val = self
            .state
            .validators
            .get_mut(&validator)
            .ok_or(StakingError::ValidatorNotFound)?;
        val.total_stake = val
            .total_stake
            .checked_sub(amount)
            .ok_or(StakingError::Overflow)?;
        self.state.total_staked = self
            .state
            .total_staked
            .checked_sub(amount)
            .ok_or(StakingError::Overflow)?;
        let completion_height = current_height
            .checked_add(UNBONDING_PERIOD_BLOCKS)
            .ok_or(StakingError::Overflow)?;
        let entry = UnbondingEntry {
            delegator,
            validator,
            amount,
            completion_height,
        };
        self.state
            .unbonding_queue
            .entry(completion_height)
            .or_default()
            .push(entry);
        self.state.events.push(StakingEvent::Undelegated {
            delegator,
            validator,
            amount,
            completion_height,
        });
        Ok(())
    }

    /// Distribute rewards proportional to stake * CU contribution.
    pub fn distribute_rewards(&mut self, total_reward: Balance) -> Result<(), StakingError> {
        if self.state.total_staked == 0 {
            return Ok(());
        }
        // Calculate total weighted contribution: sum of (delegation * validator_cu)
        let mut total_weighted: u128 = 0;
        let mut weights: Vec<(Address, u128)> = Vec::new();

        for (&(delegator, validator), &amount) in &self.state.delegations {
            if amount == 0 {
                continue;
            }
            let cu = self
                .state
                .validators
                .get(&validator)
                .map(|v| v.cu_contribution as u128)
                .unwrap_or(0);
            let weight = amount.checked_mul(cu).ok_or(StakingError::Overflow)?;
            if weight > 0 {
                total_weighted = total_weighted
                    .checked_add(weight)
                    .ok_or(StakingError::Overflow)?;
                weights.push((delegator, weight));
            }
        }

        if total_weighted == 0 {
            return Ok(());
        }

        let mut distributed: Balance = 0;
        let count = weights.len();
        for (i, (delegator, weight)) in weights.iter().enumerate() {
            let reward = if i == count - 1 {
                total_reward
                    .checked_sub(distributed)
                    .ok_or(StakingError::Overflow)?
            } else {
                weight
                    .checked_mul(total_reward)
                    .ok_or(StakingError::Overflow)?
                    / total_weighted
            };
            distributed = distributed
                .checked_add(reward)
                .ok_or(StakingError::Overflow)?;
            if reward > 0 {
                let pending = self.state.pending_rewards.entry(*delegator).or_insert(0);
                *pending = pending.checked_add(reward).ok_or(StakingError::Overflow)?;
            }
        }
        Ok(())
    }

    /// Claim pending rewards for a delegator.
    pub fn claim_rewards(&mut self, delegator: Address) -> Result<Balance, StakingError> {
        let pending = self
            .state
            .pending_rewards
            .get_mut(&delegator)
            .ok_or(StakingError::NoRewardsAvailable)?;
        if *pending == 0 {
            return Err(StakingError::NoRewardsAvailable);
        }
        let amount = *pending;
        *pending = 0;
        self.state
            .events
            .push(StakingEvent::RewardsClaimed { delegator, amount });
        Ok(amount)
    }

    /// Update the active validator set, keeping top N by total stake.
    pub fn update_validator_set(&mut self) {
        let max = self.state.max_validators;
        let mut validators: Vec<(Address, Balance)> = self
            .state
            .validators
            .iter()
            .filter(|(_, v)| v.total_stake >= MINIMUM_VALIDATOR_STAKE)
            .map(|(addr, v)| (*addr, v.total_stake))
            .collect();
        // Sort descending by stake
        validators.sort_by(|a, b| b.1.cmp(&a.1));
        let active_set: std::collections::BTreeSet<Address> =
            validators.iter().take(max).map(|(addr, _)| *addr).collect();
        for (addr, val) in self.state.validators.iter_mut() {
            val.active = active_set.contains(addr);
        }
    }

    /// Slash a validator (percentage in basis points, e.g. 500 = 5%).
    pub fn slash(
        &mut self,
        validator: Address,
        basis_points: u32,
    ) -> Result<Balance, StakingError> {
        let val = self
            .state
            .validators
            .get_mut(&validator)
            .ok_or(StakingError::ValidatorNotFound)?;
        let slash_amount = val
            .total_stake
            .checked_mul(basis_points as u128)
            .ok_or(StakingError::Overflow)?
            / 10_000;
        val.total_stake = val
            .total_stake
            .checked_sub(slash_amount)
            .ok_or(StakingError::Overflow)?;
        self.state.total_staked = self
            .state
            .total_staked
            .checked_sub(slash_amount)
            .ok_or(StakingError::Overflow)?;

        // Slash individual delegators proportionally
        let delegator_keys: Vec<(Address, Address)> = self
            .state
            .delegations
            .keys()
            .filter(|(_, v)| *v == validator)
            .cloned()
            .collect();
        for key in delegator_keys {
            if let Some(delegation) = self.state.delegations.get_mut(&key) {
                let delegator_slash = (*delegation)
                    .checked_mul(basis_points as u128)
                    .ok_or(StakingError::Overflow)?
                    / 10_000;
                *delegation = delegation
                    .checked_sub(delegator_slash)
                    .ok_or(StakingError::Overflow)?;
            }
        }

        self.state.events.push(StakingEvent::Slashed {
            validator,
            amount: slash_amount,
        });
        Ok(slash_amount)
    }

    /// Set the CU contribution for a validator.
    pub fn set_cu_contribution(&mut self, validator: Address, cu: u64) -> Result<(), StakingError> {
        let val = self
            .state
            .validators
            .get_mut(&validator)
            .ok_or(StakingError::ValidatorNotFound)?;
        val.cu_contribution = cu;
        Ok(())
    }

    /// Process completed unbonding entries at the given block height.
    pub fn process_unbonding(&mut self, current_height: BlockHeight) -> Vec<UnbondingEntry> {
        let mut completed = Vec::new();
        let heights_to_remove: Vec<BlockHeight> = self
            .state
            .unbonding_queue
            .range(..=current_height)
            .map(|(h, _)| *h)
            .collect();
        for h in heights_to_remove {
            if let Some(entries) = self.state.unbonding_queue.remove(&h) {
                completed.extend(entries);
            }
        }
        completed
    }

    /// Get events.
    pub fn events(&self) -> &[StakingEvent] {
        &self.state.events
    }
}
