use std::collections::BTreeMap;

use crate::types::{Address, Balance, BlockHeight};

use super::types::{StakingEvent, UnbondingEntry, ValidatorInfo};

/// Staking state.
#[derive(Debug, Clone)]
pub struct StakingState {
    /// Delegations: (delegator, validator) -> amount
    pub delegations: BTreeMap<(Address, Address), Balance>,
    /// Validator set
    pub validators: BTreeMap<Address, ValidatorInfo>,
    /// Unbonding queue: completion_height -> entries
    pub unbonding_queue: BTreeMap<BlockHeight, Vec<UnbondingEntry>>,
    /// Accumulated rewards: delegator -> pending rewards
    pub pending_rewards: BTreeMap<Address, Balance>,
    /// Total staked across all validators
    pub total_staked: Balance,
    /// Maximum active validators
    pub max_validators: usize,
    /// Event log
    pub events: Vec<StakingEvent>,
}

impl StakingState {
    pub fn new(max_validators: usize) -> Self {
        Self {
            delegations: BTreeMap::new(),
            validators: BTreeMap::new(),
            unbonding_queue: BTreeMap::new(),
            pending_rewards: BTreeMap::new(),
            total_staked: 0,
            max_validators,
            events: Vec::new(),
        }
    }
}

impl Default for StakingState {
    fn default() -> Self {
        Self::new(100)
    }
}
