use std::collections::BTreeMap;

use super::types::{Channel, ChannelState, IbcError, IbcEvent};

/// Channel manager for IBC channel lifecycle.
#[derive(Debug, Clone)]
pub struct ChannelManager {
    pub channels: BTreeMap<String, Channel>,
    pub events: Vec<IbcEvent>,
}

impl ChannelManager {
    pub fn new() -> Self {
        Self {
            channels: BTreeMap::new(),
            events: Vec::new(),
        }
    }

    /// Initiate channel opening (INIT state).
    pub fn open_init(
        &mut self,
        channel_id: String,
        port: String,
        counterparty_channel: String,
        counterparty_port: String,
    ) -> Result<(), IbcError> {
        let channel = Channel {
            id: channel_id.clone(),
            state: ChannelState::Init,
            counterparty_channel,
            counterparty_port,
            port: port.clone(),
            next_sequence_send: 1,
            next_sequence_recv: 1,
        };
        self.channels.insert(channel_id.clone(), channel);
        self.events
            .push(IbcEvent::ChannelOpenInit { channel_id, port });
        Ok(())
    }

    /// Process TryOpen from counterparty.
    pub fn open_try(&mut self, channel_id: &str) -> Result<(), IbcError> {
        let channel = self
            .channels
            .get_mut(channel_id)
            .ok_or_else(|| IbcError::ChannelNotFound(channel_id.to_string()))?;
        if channel.state != ChannelState::Init {
            return Err(IbcError::InvalidChannelState);
        }
        channel.state = ChannelState::TryOpen;
        self.events.push(IbcEvent::ChannelOpenTry {
            channel_id: channel_id.to_string(),
        });
        Ok(())
    }

    /// Acknowledge channel opening.
    pub fn open_ack(&mut self, channel_id: &str) -> Result<(), IbcError> {
        let channel = self
            .channels
            .get_mut(channel_id)
            .ok_or_else(|| IbcError::ChannelNotFound(channel_id.to_string()))?;
        if channel.state != ChannelState::TryOpen {
            return Err(IbcError::InvalidChannelState);
        }
        channel.state = ChannelState::Open;
        self.events.push(IbcEvent::ChannelOpenAck {
            channel_id: channel_id.to_string(),
        });
        Ok(())
    }

    /// Confirm channel opening.
    pub fn open_confirm(&mut self, channel_id: &str) -> Result<(), IbcError> {
        let channel = self
            .channels
            .get_mut(channel_id)
            .ok_or_else(|| IbcError::ChannelNotFound(channel_id.to_string()))?;
        if channel.state != ChannelState::Open {
            return Err(IbcError::InvalidChannelState);
        }
        self.events.push(IbcEvent::ChannelOpenConfirm {
            channel_id: channel_id.to_string(),
        });
        Ok(())
    }

    /// Initiate channel close.
    pub fn close_init(&mut self, channel_id: &str) -> Result<(), IbcError> {
        let channel = self
            .channels
            .get_mut(channel_id)
            .ok_or_else(|| IbcError::ChannelNotFound(channel_id.to_string()))?;
        if channel.state != ChannelState::Open {
            return Err(IbcError::InvalidChannelState);
        }
        channel.state = ChannelState::Closed;
        self.events.push(IbcEvent::ChannelClosed {
            channel_id: channel_id.to_string(),
        });
        Ok(())
    }

    /// Confirm channel close.
    pub fn close_confirm(&mut self, channel_id: &str) -> Result<(), IbcError> {
        let channel = self
            .channels
            .get_mut(channel_id)
            .ok_or_else(|| IbcError::ChannelNotFound(channel_id.to_string()))?;
        if channel.state == ChannelState::Closed {
            return Ok(());
        }
        channel.state = ChannelState::Closed;
        self.events.push(IbcEvent::ChannelClosed {
            channel_id: channel_id.to_string(),
        });
        Ok(())
    }

    /// Get a channel by ID.
    pub fn get_channel(&self, channel_id: &str) -> Option<&Channel> {
        self.channels.get(channel_id)
    }

    /// Check if a channel is open.
    pub fn is_channel_open(&self, channel_id: &str) -> bool {
        self.channels
            .get(channel_id)
            .is_some_and(|c| c.state == ChannelState::Open)
    }
}

impl Default for ChannelManager {
    fn default() -> Self {
        Self::new()
    }
}
