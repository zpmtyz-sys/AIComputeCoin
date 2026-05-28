use serde::{Deserialize, Serialize};

/// On-chain address represented as 32 bytes.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Serialize, Deserialize)]
pub struct Address(pub [u8; 32]);

/// Token balance type (supports up to 2^128 - 1 smallest units).
pub type Balance = u128;

/// Block height.
pub type BlockHeight = u64;

/// Unix timestamp in seconds.
pub type Timestamp = u64;

impl Address {
    /// Create an address from a hex string.
    pub fn from_hex(s: &str) -> Result<Self, hex::FromHexError> {
        let bytes = hex::decode(s)?;
        let mut addr = [0u8; 32];
        let len = bytes.len().min(32);
        addr[..len].copy_from_slice(&bytes[..len]);
        Ok(Address(addr))
    }

    /// Convert address to hex string.
    pub fn to_hex(&self) -> String {
        hex::encode(self.0)
    }
}
