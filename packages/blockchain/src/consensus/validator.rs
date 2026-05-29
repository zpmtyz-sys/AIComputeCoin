use std::collections::BTreeMap;

use crate::types::{Address, Balance};

/// Entry tracking a validator's participation in consensus.
#[derive(Debug, Clone)]
pub struct ValidatorEntry {
    pub address: Address,
    pub stake: Balance,
    pub cu_contribution: u64,
    pub active: bool,
}

/// Validator set management for consensus.
#[derive(Debug, Clone)]
pub struct ValidatorSet {
    pub validators: BTreeMap<Address, ValidatorEntry>,
    pub max_active: usize,
}

impl ValidatorSet {
    pub fn new() -> Self {
        Self {
            validators: BTreeMap::new(),
            max_active: 100,
        }
    }

    pub fn with_max_active(max_active: usize) -> Self {
        Self {
            validators: BTreeMap::new(),
            max_active,
        }
    }

    /// Add or update a validator.
    pub fn set_validator(&mut self, address: Address, stake: Balance, cu_contribution: u64) {
        self.validators.insert(
            address,
            ValidatorEntry {
                address,
                stake,
                cu_contribution,
                active: true,
            },
        );
    }

    /// Remove a validator.
    pub fn remove_validator(&mut self, address: &Address) {
        self.validators.remove(address);
    }

    /// Recalculate the active validator set based on combined score (stake * CU).
    pub fn recalculate(&mut self) {
        let mut scored: Vec<(Address, u128)> = self
            .validators
            .iter()
            .map(|(addr, entry)| {
                let score = entry
                    .stake
                    .checked_mul(entry.cu_contribution as u128)
                    .unwrap_or(0);
                (*addr, score)
            })
            .collect();
        scored.sort_by(|a, b| b.1.cmp(&a.1));
        let active_set: std::collections::BTreeSet<Address> = scored
            .iter()
            .take(self.max_active)
            .map(|(a, _)| *a)
            .collect();
        for (addr, entry) in self.validators.iter_mut() {
            entry.active = active_set.contains(addr);
        }
    }

    /// Get list of active validator addresses.
    pub fn get_active_validators(&self) -> Vec<Address> {
        self.validators
            .values()
            .filter(|v| v.active)
            .map(|v| v.address)
            .collect()
    }

    /// Get total CU of active validators.
    pub fn total_cu(&self) -> u64 {
        self.validators
            .values()
            .filter(|v| v.active)
            .map(|v| v.cu_contribution)
            .sum()
    }

    /// Select block producer weighted by stake * CU contribution.
    /// Uses a deterministic selection based on block height as seed.
    pub fn select_producer(&self, seed: u64) -> Option<Address> {
        let active: Vec<&ValidatorEntry> = self.validators.values().filter(|v| v.active).collect();
        if active.is_empty() {
            return None;
        }
        let total_weight: u128 = active
            .iter()
            .map(|v| v.stake.checked_mul(v.cu_contribution as u128).unwrap_or(0))
            .sum();
        if total_weight == 0 {
            // Fallback: select by index
            let idx = (seed as usize) % active.len();
            return Some(active[idx].address);
        }
        let target = (seed as u128) % total_weight;
        let mut cumulative: u128 = 0;
        for v in &active {
            let weight = v.stake.checked_mul(v.cu_contribution as u128).unwrap_or(0);
            cumulative = cumulative.saturating_add(weight);
            if cumulative > target {
                return Some(v.address);
            }
        }
        Some(active.last().unwrap().address)
    }
}

impl Default for ValidatorSet {
    fn default() -> Self {
        Self::new()
    }
}
