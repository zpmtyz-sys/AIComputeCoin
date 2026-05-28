use std::collections::HashMap;

use crate::types::{Address, Balance};

/// Token operations trait.
pub trait Token {
    fn mint(&mut self, to: Address, amount: Balance) -> Result<(), TokenError>;
    fn burn(&mut self, from: Address, amount: Balance) -> Result<(), TokenError>;
    fn transfer(&mut self, from: Address, to: Address, amount: Balance) -> Result<(), TokenError>;
    fn balance_of(&self, account: &Address) -> Balance;
}

#[derive(Debug, thiserror::Error)]
pub enum TokenError {
    #[error("insufficient balance")]
    InsufficientBalance,
    #[error("overflow")]
    Overflow,
}

/// Simple in-memory token implementation.
pub struct SimpleToken {
    balances: HashMap<Address, Balance>,
    total_supply: Balance,
}

impl SimpleToken {
    pub fn new() -> Self {
        Self {
            balances: HashMap::new(),
            total_supply: 0,
        }
    }

    pub fn total_supply(&self) -> Balance {
        self.total_supply
    }
}

impl Default for SimpleToken {
    fn default() -> Self {
        Self::new()
    }
}

impl Token for SimpleToken {
    fn mint(&mut self, to: Address, amount: Balance) -> Result<(), TokenError> {
        let balance = self.balances.entry(to).or_insert(0);
        *balance = balance.checked_add(amount).ok_or(TokenError::Overflow)?;
        self.total_supply = self
            .total_supply
            .checked_add(amount)
            .ok_or(TokenError::Overflow)?;
        Ok(())
    }

    fn burn(&mut self, from: Address, amount: Balance) -> Result<(), TokenError> {
        let balance = self.balances.entry(from).or_insert(0);
        if *balance < amount {
            return Err(TokenError::InsufficientBalance);
        }
        *balance -= amount;
        self.total_supply -= amount;
        Ok(())
    }

    fn transfer(&mut self, from: Address, to: Address, amount: Balance) -> Result<(), TokenError> {
        let from_balance = self.balances.get(&from).copied().unwrap_or(0);
        if from_balance < amount {
            return Err(TokenError::InsufficientBalance);
        }

        *self.balances.entry(from).or_insert(0) -= amount;
        let to_balance = self.balances.entry(to).or_insert(0);
        *to_balance = to_balance.checked_add(amount).ok_or(TokenError::Overflow)?;
        Ok(())
    }

    fn balance_of(&self, account: &Address) -> Balance {
        self.balances.get(account).copied().unwrap_or(0)
    }
}
