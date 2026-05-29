use sha2::{Digest, Sha256};

use crate::types::{Address, Balance, BlockHeight, EpochNumber};

use super::types::{
    BlockReward, CUReport, ConsensusError, ConsensusEvent, EpochInfo, BASE_REWARD,
    BLOCKS_PER_EPOCH, VARIABLE_REWARD,
};
use super::validator::ValidatorSet;

/// Proof-of-Compute consensus engine.
pub struct ProofOfCompute {
    pub validator_set: ValidatorSet,
    pub current_epoch: EpochInfo,
    pub events: Vec<ConsensusEvent>,
    /// Minimum required oracle signatures.
    pub min_signatures: usize,
    /// Maximum allowed CU value.
    pub max_cu_value: u64,
    /// Known oracle public key hashes (simplified for verification).
    pub oracle_key_hashes: Vec<[u8; 32]>,
}

impl ProofOfCompute {
    pub fn new(min_signatures: usize, max_cu_value: u64) -> Self {
        Self {
            validator_set: ValidatorSet::new(),
            current_epoch: EpochInfo {
                epoch_number: 0,
                start_block: 0,
                validators: Vec::new(),
                total_cu: 0,
            },
            events: Vec::new(),
            min_signatures,
            max_cu_value,
            oracle_key_hashes: Vec::new(),
        }
    }

    /// Add an oracle key hash for signature verification.
    pub fn add_oracle_key(&mut self, key_hash: [u8; 32]) {
        self.oracle_key_hashes.push(key_hash);
    }

    /// Calculate block production weight based on stake and compute contribution.
    pub fn block_production_weight(stake: Balance, compute_contribution: u64) -> u128 {
        stake.saturating_mul(compute_contribution as u128)
    }

    /// Verify an oracle-signed CU report.
    pub fn verify_cu_report(&self, report: &CUReport) -> Result<(), ConsensusError> {
        if report.oracle_signatures.is_empty() {
            return Err(ConsensusError::NoSignatures);
        }
        if report.oracle_signatures.len() < self.min_signatures {
            return Err(ConsensusError::NoSignatures);
        }
        if report.cu_value > self.max_cu_value {
            return Err(ConsensusError::CUValueOutOfRange);
        }
        // Verify each signature by computing a hash and checking it matches an oracle key
        for sig in &report.oracle_signatures {
            let sig_hash = self.hash_signature(sig, &report.node_id, report.cu_value);
            if !self
                .oracle_key_hashes
                .iter()
                .any(|key| self.verify_sig_hash(&sig_hash, key))
            {
                return Err(ConsensusError::InvalidSignature);
            }
        }
        Ok(())
    }

    /// Calculate block reward based on CU contribution.
    pub fn calculate_block_reward(
        cu_contribution: u64,
        total_cu: u64,
    ) -> Result<BlockReward, ConsensusError> {
        let base = BASE_REWARD;
        let variable = if total_cu > 0 {
            (cu_contribution as u128)
                .checked_mul(VARIABLE_REWARD)
                .ok_or(ConsensusError::Overflow)?
                / (total_cu as u128)
        } else {
            0
        };
        let total = base.checked_add(variable).ok_or(ConsensusError::Overflow)?;
        Ok(BlockReward {
            base,
            variable,
            total,
        })
    }

    /// Check if an epoch transition should occur and process it.
    pub fn check_epoch_transition(&mut self, current_height: BlockHeight) -> Option<EpochNumber> {
        let blocks_since_start = current_height.saturating_sub(self.current_epoch.start_block);
        if blocks_since_start >= BLOCKS_PER_EPOCH && current_height > 0 {
            let old_epoch = self.current_epoch.epoch_number;
            let new_epoch = old_epoch.saturating_add(1);
            self.current_epoch = EpochInfo {
                epoch_number: new_epoch,
                start_block: current_height,
                validators: self.validator_set.get_active_validators(),
                total_cu: self.validator_set.total_cu(),
            };
            self.events.push(ConsensusEvent::EpochTransition {
                old_epoch,
                new_epoch,
            });
            Some(new_epoch)
        } else {
            None
        }
    }

    /// Record block production.
    pub fn record_block_production(
        &mut self,
        producer: Address,
        height: BlockHeight,
        cu_contribution: u64,
    ) -> Result<BlockReward, ConsensusError> {
        let reward = Self::calculate_block_reward(cu_contribution, self.current_epoch.total_cu)?;
        self.events.push(ConsensusEvent::BlockProduced {
            producer,
            height,
            reward: reward.total,
        });
        Ok(reward)
    }

    /// Hash a signature for verification (simplified).
    fn hash_signature(&self, sig: &[u8; 64], node_id: &Address, cu_value: u64) -> [u8; 32] {
        let mut hasher = Sha256::new();
        hasher.update(sig);
        hasher.update(node_id.0);
        hasher.update(cu_value.to_le_bytes());
        let result = hasher.finalize();
        let mut hash = [0u8; 32];
        hash.copy_from_slice(&result);
        hash
    }

    /// Verify a signature hash against a known oracle key (simplified).
    fn verify_sig_hash(&self, sig_hash: &[u8; 32], oracle_key: &[u8; 32]) -> bool {
        // Simplified verification: check the first 16 bytes match
        // (in production, this would use ed25519 verification)
        sig_hash[..16] == oracle_key[..16]
    }

    /// Get events.
    pub fn events(&self) -> &[ConsensusEvent] {
        &self.events
    }
}
