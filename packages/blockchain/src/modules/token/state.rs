use std::collections::{BTreeMap, BTreeSet};

use crate::types::{Address, Balance};

use super::types::{TokenEvent, TokenMetadata};

/// Token state holding balances, supply, and event log.
#[derive(Debug, Clone)]
pub struct TokenState {
    pub metadata: TokenMetadata,
    pub balances: BTreeMap<Address, Balance>,
    pub total_supply: Balance,
    pub authorized_minters: BTreeSet<Address>,
    pub events: Vec<TokenEvent>,
}

impl TokenState {
    pub fn new(metadata: TokenMetadata) -> Self {
        Self {
            metadata,
            balances: BTreeMap::new(),
            total_supply: 0,
            authorized_minters: BTreeSet::new(),
            events: Vec::new(),
        }
    }
}

impl Default for TokenState {
    fn default() -> Self {
        Self::new(TokenMetadata::default())
    }
}
