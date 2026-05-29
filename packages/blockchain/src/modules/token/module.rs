use crate::types::{Address, Balance};

use super::state::TokenState;
use super::types::{TokenError, TokenEvent};

/// Token module providing mint, burn, transfer operations.
pub struct TokenModule {
    pub state: TokenState,
}

impl TokenModule {
    pub fn new(state: TokenState) -> Self {
        Self { state }
    }

    /// Add an authorized minter.
    pub fn add_minter(&mut self, minter: Address) {
        self.state.authorized_minters.insert(minter);
    }

    /// Remove an authorized minter.
    pub fn remove_minter(&mut self, minter: Address) {
        self.state.authorized_minters.remove(&minter);
    }

    /// Mint tokens to an address. Only authorized minters can call this.
    pub fn mint(
        &mut self,
        minter: &Address,
        to: Address,
        amount: Balance,
    ) -> Result<(), TokenError> {
        if amount == 0 {
            return Err(TokenError::ZeroAmount);
        }
        if !self.state.authorized_minters.contains(minter) {
            return Err(TokenError::UnauthorizedMinter);
        }
        let balance = self.state.balances.entry(to).or_insert(0);
        *balance = balance.checked_add(amount).ok_or(TokenError::Overflow)?;
        self.state.total_supply = self
            .state
            .total_supply
            .checked_add(amount)
            .ok_or(TokenError::Overflow)?;
        self.state.events.push(TokenEvent::Mint { to, amount });
        Ok(())
    }

    /// Burn tokens from an address.
    pub fn burn(&mut self, from: Address, amount: Balance) -> Result<(), TokenError> {
        if amount == 0 {
            return Err(TokenError::ZeroAmount);
        }
        let balance = self.state.balances.entry(from).or_insert(0);
        if *balance < amount {
            return Err(TokenError::InsufficientBalance);
        }
        *balance = balance.checked_sub(amount).ok_or(TokenError::Overflow)?;
        self.state.total_supply = self
            .state
            .total_supply
            .checked_sub(amount)
            .ok_or(TokenError::Overflow)?;
        self.state.events.push(TokenEvent::Burn { from, amount });
        Ok(())
    }

    /// Transfer tokens between addresses.
    pub fn transfer(
        &mut self,
        from: Address,
        to: Address,
        amount: Balance,
    ) -> Result<(), TokenError> {
        if amount == 0 {
            return Err(TokenError::ZeroAmount);
        }
        let from_balance = self.state.balances.get(&from).copied().unwrap_or(0);
        if from_balance < amount {
            return Err(TokenError::InsufficientBalance);
        }
        // Deduct from sender
        let sender_bal = self.state.balances.entry(from).or_insert(0);
        *sender_bal = sender_bal.checked_sub(amount).ok_or(TokenError::Overflow)?;
        // Credit to receiver
        let receiver_bal = self.state.balances.entry(to).or_insert(0);
        *receiver_bal = receiver_bal
            .checked_add(amount)
            .ok_or(TokenError::Overflow)?;
        self.state
            .events
            .push(TokenEvent::Transfer { from, to, amount });
        Ok(())
    }

    /// Get the balance of an address.
    pub fn balance_of(&self, account: &Address) -> Balance {
        self.state.balances.get(account).copied().unwrap_or(0)
    }

    /// Get the total supply.
    pub fn total_supply(&self) -> Balance {
        self.state.total_supply
    }

    /// Get emitted events.
    pub fn events(&self) -> &[TokenEvent] {
        &self.state.events
    }
}
