use std::collections::BTreeMap;

use crate::types::{Address, Balance};

use super::types::{IbcError, IbcEvent, TransferMessage};

/// ICS-20 token transfer using lock/mint pattern.
#[derive(Debug, Clone)]
pub struct TokenTransfer {
    /// Locked tokens on the source chain: (sender, denom) -> amount
    pub locked: BTreeMap<(Address, String), Balance>,
    /// Minted tokens on the destination chain: (receiver, denom) -> amount
    pub minted: BTreeMap<(Address, String), Balance>,
    /// Pending transfers: channel -> sequence -> transfer message
    pub pending: BTreeMap<(String, u64), TransferMessage>,
    pub events: Vec<IbcEvent>,
}

impl TokenTransfer {
    pub fn new() -> Self {
        Self {
            locked: BTreeMap::new(),
            minted: BTreeMap::new(),
            pending: BTreeMap::new(),
            events: Vec::new(),
        }
    }

    /// Send a token transfer (locks tokens on source chain).
    pub fn send_transfer(
        &mut self,
        msg: TransferMessage,
        sender_balance: Balance,
        channel: &str,
        sequence: u64,
    ) -> Result<Vec<u8>, IbcError> {
        if sender_balance < msg.amount {
            return Err(IbcError::InsufficientBalance);
        }
        if msg.amount == 0 {
            return Err(IbcError::InvalidPacketData);
        }
        // Lock tokens
        let key = (msg.sender, msg.denom.clone());
        let locked = self.locked.entry(key).or_insert(0);
        *locked = locked.checked_add(msg.amount).ok_or(IbcError::Overflow)?;
        self.events.push(IbcEvent::TokensLocked {
            sender: msg.sender,
            amount: msg.amount,
        });
        // Construct packet data
        let packet_data = serde_json::to_vec(&msg).map_err(|_| IbcError::InvalidPacketData)?;
        // Store pending transfer for potential refund
        self.pending.insert((channel.to_string(), sequence), msg);
        Ok(packet_data)
    }

    /// Receive a token transfer (mints tokens on destination chain).
    pub fn recv_transfer(&mut self, data: &[u8]) -> Result<Address, IbcError> {
        let msg: TransferMessage =
            serde_json::from_slice(data).map_err(|_| IbcError::InvalidPacketData)?;
        if msg.amount == 0 {
            return Err(IbcError::InvalidPacketData);
        }
        let key = (msg.receiver, msg.denom);
        let minted = self.minted.entry(key).or_insert(0);
        *minted = minted.checked_add(msg.amount).ok_or(IbcError::Overflow)?;
        self.events.push(IbcEvent::TokensMinted {
            receiver: msg.receiver,
            amount: msg.amount,
        });
        Ok(msg.receiver)
    }

    /// Refund a failed/timed-out transfer (unlocks tokens).
    pub fn refund_transfer(&mut self, channel: &str, sequence: u64) -> Result<(), IbcError> {
        let msg = self
            .pending
            .remove(&(channel.to_string(), sequence))
            .ok_or(IbcError::CommitmentNotFound)?;
        let key = (msg.sender, msg.denom);
        let locked = self
            .locked
            .get_mut(&key)
            .ok_or(IbcError::InsufficientBalance)?;
        *locked = locked
            .checked_sub(msg.amount)
            .ok_or(IbcError::InsufficientBalance)?;
        self.events.push(IbcEvent::TokensRefunded {
            receiver: msg.sender,
            amount: msg.amount,
        });
        Ok(())
    }

    /// Get locked amount for an address and denom.
    pub fn get_locked(&self, address: &Address, denom: &str) -> Balance {
        self.locked
            .get(&(*address, denom.to_string()))
            .copied()
            .unwrap_or(0)
    }

    /// Get minted amount for an address and denom.
    pub fn get_minted(&self, address: &Address, denom: &str) -> Balance {
        self.minted
            .get(&(*address, denom.to_string()))
            .copied()
            .unwrap_or(0)
    }
}

impl Default for TokenTransfer {
    fn default() -> Self {
        Self::new()
    }
}
