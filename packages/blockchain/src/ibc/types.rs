use serde::{Deserialize, Serialize};

use crate::types::{Address, Balance, BlockHeight};

/// Channel state machine states.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum ChannelState {
    Init,
    TryOpen,
    Open,
    Closed,
}

/// IBC Channel.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Channel {
    pub id: String,
    pub state: ChannelState,
    pub counterparty_channel: String,
    pub counterparty_port: String,
    pub port: String,
    pub next_sequence_send: u64,
    pub next_sequence_recv: u64,
}

/// IBC Packet.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub struct Packet {
    pub sequence: u64,
    pub source_port: String,
    pub source_channel: String,
    pub dest_port: String,
    pub dest_channel: String,
    pub data: Vec<u8>,
    pub timeout_height: BlockHeight,
    pub timeout_timestamp: u64,
}

/// Acknowledgement result.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum Acknowledgement {
    Success(Vec<u8>),
    Error(String),
}

/// ICS-20 Transfer message.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct TransferMessage {
    pub sender: Address,
    pub receiver: Address,
    pub amount: Balance,
    pub denom: String,
    pub source_channel: String,
}

/// IBC events.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum IbcEvent {
    ChannelOpenInit {
        channel_id: String,
        port: String,
    },
    ChannelOpenTry {
        channel_id: String,
    },
    ChannelOpenAck {
        channel_id: String,
    },
    ChannelOpenConfirm {
        channel_id: String,
    },
    ChannelClosed {
        channel_id: String,
    },
    PacketSent {
        sequence: u64,
        source_channel: String,
    },
    PacketReceived {
        sequence: u64,
        dest_channel: String,
    },
    PacketAcknowledged {
        sequence: u64,
        channel_id: String,
    },
    PacketTimeout {
        sequence: u64,
        channel_id: String,
    },
    TokensLocked {
        sender: Address,
        amount: Balance,
    },
    TokensMinted {
        receiver: Address,
        amount: Balance,
    },
    TokensRefunded {
        receiver: Address,
        amount: Balance,
    },
}

/// IBC errors.
#[derive(Debug, thiserror::Error, PartialEq, Eq)]
pub enum IbcError {
    #[error("channel not found: {0}")]
    ChannelNotFound(String),
    #[error("channel not open")]
    ChannelNotOpen,
    #[error("invalid channel state for operation")]
    InvalidChannelState,
    #[error("packet timeout")]
    PacketTimeout,
    #[error("packet commitment not found")]
    CommitmentNotFound,
    #[error("invalid sequence number")]
    InvalidSequence,
    #[error("insufficient balance")]
    InsufficientBalance,
    #[error("overflow")]
    Overflow,
    #[error("invalid packet data")]
    InvalidPacketData,
}
