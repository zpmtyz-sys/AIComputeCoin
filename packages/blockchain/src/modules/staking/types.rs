use serde::{Deserialize, Serialize};

use crate::types::{Address, Balance, BlockHeight};

/// Minimum CC to delegate.
pub const MINIMUM_DELEGATION: Balance = 1_000;

/// Minimum CC to operate a validator.
pub const MINIMUM_VALIDATOR_STAKE: Balance = 100_000;

/// Unbonding period in blocks (~21 days at 6s block time).
pub const UNBONDING_PERIOD_BLOCKS: BlockHeight = 302_400;

/// Delegation record.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Delegation {
    pub delegator: Address,
    pub validator: Address,
    pub amount: Balance,
}

/// Entry in the unbonding queue.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct UnbondingEntry {
    pub delegator: Address,
    pub validator: Address,
    pub amount: Balance,
    pub completion_height: BlockHeight,
}

/// Validator information.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct ValidatorInfo {
    pub address: Address,
    pub total_stake: Balance,
    pub cu_contribution: u64,
    pub commission_rate: u32, // basis points (e.g. 1000 = 10%)
    pub active: bool,
}

/// Events emitted by staking operations.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum StakingEvent {
    Delegated {
        delegator: Address,
        validator: Address,
        amount: Balance,
    },
    Undelegated {
        delegator: Address,
        validator: Address,
        amount: Balance,
        completion_height: BlockHeight,
    },
    RewardsClaimed {
        delegator: Address,
        amount: Balance,
    },
    Slashed {
        validator: Address,
        amount: Balance,
    },
    ValidatorAdded {
        address: Address,
    },
    ValidatorRemoved {
        address: Address,
    },
}

/// Staking operation errors.
#[derive(Debug, thiserror::Error, PartialEq, Eq)]
pub enum StakingError {
    #[error("insufficient balance")]
    InsufficientBalance,
    #[error("minimum delegation not met (requires {0} CC)")]
    MinimumDelegationNotMet(Balance),
    #[error("minimum validator stake not met (requires {0} CC)")]
    MinimumValidatorStakeNotMet(Balance),
    #[error("validator not found")]
    ValidatorNotFound,
    #[error("validator already exists")]
    ValidatorAlreadyExists,
    #[error("no delegation found")]
    NoDelegationFound,
    #[error("no rewards available")]
    NoRewardsAvailable,
    #[error("overflow")]
    Overflow,
}
