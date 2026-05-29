use serde::{Deserialize, Serialize};

use crate::types::{Address, Balance, BlockHeight, EpochNumber};

/// Number of blocks per epoch.
pub const BLOCKS_PER_EPOCH: BlockHeight = 100;

/// Base block reward (in smallest units).
pub const BASE_REWARD: Balance = 1_000_000;

/// Variable reward pool (in smallest units).
pub const VARIABLE_REWARD: Balance = 4_000_000;

/// Oracle-signed Compute Unit report.
#[derive(Debug, Clone)]
pub struct CUReport {
    pub node_id: Address,
    pub cu_value: u64,
    pub oracle_signatures: Vec<[u8; 64]>,
    pub timestamp: u64,
}

/// Epoch information for consensus.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct EpochInfo {
    pub epoch_number: EpochNumber,
    pub start_block: BlockHeight,
    pub validators: Vec<Address>,
    pub total_cu: u64,
}

/// Block reward breakdown.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BlockReward {
    pub base: Balance,
    pub variable: Balance,
    pub total: Balance,
}

/// Consensus events.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum ConsensusEvent {
    EpochTransition {
        old_epoch: EpochNumber,
        new_epoch: EpochNumber,
    },
    BlockProduced {
        producer: Address,
        height: BlockHeight,
        reward: Balance,
    },
    CUReportVerified {
        node_id: Address,
        cu_value: u64,
    },
}

/// Consensus errors.
#[derive(Debug, thiserror::Error, PartialEq, Eq)]
pub enum ConsensusError {
    #[error("invalid CU report: no signatures")]
    NoSignatures,
    #[error("invalid CU report: value out of range")]
    CUValueOutOfRange,
    #[error("invalid signature")]
    InvalidSignature,
    #[error("no validators registered")]
    NoValidators,
    #[error("not a valid block producer")]
    InvalidBlockProducer,
    #[error("overflow")]
    Overflow,
}
