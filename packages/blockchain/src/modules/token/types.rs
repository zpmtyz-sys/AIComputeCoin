use serde::{Deserialize, Serialize};

use crate::types::{Address, Balance};

/// Token metadata.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TokenMetadata {
    pub name: String,
    pub symbol: String,
    pub decimals: u8,
}

impl Default for TokenMetadata {
    fn default() -> Self {
        Self {
            name: "ComputeCoin".to_string(),
            symbol: "CC".to_string(),
            decimals: 18,
        }
    }
}

/// Events emitted by token operations.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum TokenEvent {
    Transfer {
        from: Address,
        to: Address,
        amount: Balance,
    },
    Mint {
        to: Address,
        amount: Balance,
    },
    Burn {
        from: Address,
        amount: Balance,
    },
}

/// Token operation errors.
#[derive(Debug, thiserror::Error, PartialEq, Eq)]
pub enum TokenError {
    #[error("insufficient balance")]
    InsufficientBalance,
    #[error("overflow")]
    Overflow,
    #[error("unauthorized minter")]
    UnauthorizedMinter,
    #[error("zero amount")]
    ZeroAmount,
}
