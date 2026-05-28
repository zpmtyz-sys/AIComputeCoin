use std::collections::HashMap;

use crate::types::{Address, Balance};

/// Staking operations trait.
pub trait Staking {
    fn stake(&mut self, account: Address, amount: Balance) -> Result<(), StakingError>;
    fn unstake(&mut self, account: Address, amount: Balance) -> Result<(), StakingError>;
    fn get_stake(&self, account: &Address) -> Balance;
    fn distribute_rewards(&mut self, total_reward: Balance) -> Result<(), StakingError>;
}

#[derive(Debug, thiserror::Error)]
pub enum StakingError {
    #[error("insufficient stake")]
    InsufficientStake,
    #[error("minimum stake not met")]
    MinimumStakeNotMet,
    #[error("overflow")]
    Overflow,
}

/// Simple staking pool implementation.
pub struct StakingPool {
    stakes: HashMap<Address, Balance>,
    total_staked: Balance,
    min_stake: Balance,
}

impl StakingPool {
    pub fn new(min_stake: Balance) -> Self {
        Self {
            stakes: HashMap::new(),
            total_staked: 0,
            min_stake,
        }
    }

    pub fn total_staked(&self) -> Balance {
        self.total_staked
    }
}

impl Staking for StakingPool {
    fn stake(&mut self, account: Address, amount: Balance) -> Result<(), StakingError> {
        if amount < self.min_stake {
            return Err(StakingError::MinimumStakeNotMet);
        }
        let current = self.stakes.entry(account).or_insert(0);
        *current = current.checked_add(amount).ok_or(StakingError::Overflow)?;
        self.total_staked = self
            .total_staked
            .checked_add(amount)
            .ok_or(StakingError::Overflow)?;
        Ok(())
    }

    fn unstake(&mut self, account: Address, amount: Balance) -> Result<(), StakingError> {
        let current = self.stakes.get_mut(&account).ok_or(StakingError::InsufficientStake)?;
        if *current < amount {
            return Err(StakingError::InsufficientStake);
        }
        *current -= amount;
        self.total_staked -= amount;
        Ok(())
    }

    fn get_stake(&self, account: &Address) -> Balance {
        self.stakes.get(account).copied().unwrap_or(0)
    }

    fn distribute_rewards(&mut self, total_reward: Balance) -> Result<(), StakingError> {
        if self.total_staked == 0 {
            return Ok(());
        }
        // Distribute proportionally to stake
        let stakers: Vec<(Address, Balance)> = self.stakes.iter().map(|(a, s)| (*a, *s)).collect();
        for (account, stake) in stakers {
            let reward = (stake as u128)
                .checked_mul(total_reward)
                .ok_or(StakingError::Overflow)?
                / self.total_staked;
            if reward > 0 {
                *self.stakes.entry(account).or_insert(0) += reward;
                self.total_staked += reward;
            }
        }
        Ok(())
    }
}
