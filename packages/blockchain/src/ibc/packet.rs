use std::collections::BTreeMap;

use crate::types::BlockHeight;

use super::channel::ChannelManager;
use super::types::{Acknowledgement, ChannelState, IbcError, IbcEvent, Packet};

/// Packet relay for IBC packet send/receive/acknowledge/timeout.
#[derive(Debug, Clone)]
pub struct PacketRelay {
    pub channel_manager: ChannelManager,
    /// Packet commitments: (channel_id, sequence) -> packet hash
    pub commitments: BTreeMap<(String, u64), Vec<u8>>,
    /// Received acknowledgements: (channel_id, sequence) -> ack
    pub acknowledgements: BTreeMap<(String, u64), Acknowledgement>,
    pub events: Vec<IbcEvent>,
}

impl PacketRelay {
    pub fn new(channel_manager: ChannelManager) -> Self {
        Self {
            channel_manager,
            commitments: BTreeMap::new(),
            acknowledgements: BTreeMap::new(),
            events: Vec::new(),
        }
    }

    /// Send a packet on an open channel.
    #[allow(clippy::too_many_arguments)]
    pub fn send_packet(
        &mut self,
        source_port: String,
        source_channel: String,
        dest_port: String,
        dest_channel: String,
        data: Vec<u8>,
        timeout_height: BlockHeight,
        timeout_timestamp: u64,
    ) -> Result<Packet, IbcError> {
        let channel = self
            .channel_manager
            .channels
            .get_mut(&source_channel)
            .ok_or_else(|| IbcError::ChannelNotFound(source_channel.clone()))?;
        if channel.state != ChannelState::Open {
            return Err(IbcError::ChannelNotOpen);
        }
        let sequence = channel.next_sequence_send;
        channel.next_sequence_send = sequence.checked_add(1).ok_or(IbcError::Overflow)?;
        let packet = Packet {
            sequence,
            source_port,
            source_channel: source_channel.clone(),
            dest_port,
            dest_channel,
            data,
            timeout_height,
            timeout_timestamp,
        };
        // Store commitment
        let commitment = self.compute_commitment(&packet);
        self.commitments
            .insert((source_channel.clone(), sequence), commitment);
        self.events.push(IbcEvent::PacketSent {
            sequence,
            source_channel,
        });
        Ok(packet)
    }

    /// Receive a packet on the destination chain.
    pub fn recv_packet(&mut self, packet: &Packet) -> Result<Acknowledgement, IbcError> {
        let channel = self
            .channel_manager
            .channels
            .get_mut(&packet.dest_channel)
            .ok_or_else(|| IbcError::ChannelNotFound(packet.dest_channel.clone()))?;
        if channel.state != ChannelState::Open {
            return Err(IbcError::ChannelNotOpen);
        }
        if packet.sequence != channel.next_sequence_recv {
            return Err(IbcError::InvalidSequence);
        }
        channel.next_sequence_recv = packet.sequence.checked_add(1).ok_or(IbcError::Overflow)?;
        self.events.push(IbcEvent::PacketReceived {
            sequence: packet.sequence,
            dest_channel: packet.dest_channel.clone(),
        });
        let ack = Acknowledgement::Success(b"ok".to_vec());
        self.acknowledgements
            .insert((packet.dest_channel.clone(), packet.sequence), ack.clone());
        Ok(ack)
    }

    /// Acknowledge a packet (on the source chain after receiving ack from dest).
    pub fn acknowledge_packet(
        &mut self,
        channel_id: &str,
        sequence: u64,
        _ack: &Acknowledgement,
    ) -> Result<(), IbcError> {
        let key = (channel_id.to_string(), sequence);
        if self.commitments.remove(&key).is_none() {
            return Err(IbcError::CommitmentNotFound);
        }
        self.events.push(IbcEvent::PacketAcknowledged {
            sequence,
            channel_id: channel_id.to_string(),
        });
        Ok(())
    }

    /// Process a packet timeout (refund).
    pub fn timeout_packet(
        &mut self,
        channel_id: &str,
        sequence: u64,
        current_height: BlockHeight,
        timeout_height: BlockHeight,
    ) -> Result<(), IbcError> {
        if current_height < timeout_height {
            return Err(IbcError::PacketTimeout);
        }
        let key = (channel_id.to_string(), sequence);
        if self.commitments.remove(&key).is_none() {
            return Err(IbcError::CommitmentNotFound);
        }
        self.events.push(IbcEvent::PacketTimeout {
            sequence,
            channel_id: channel_id.to_string(),
        });
        Ok(())
    }

    fn compute_commitment(&self, packet: &Packet) -> Vec<u8> {
        use sha2::{Digest, Sha256};
        let mut hasher = Sha256::new();
        hasher.update(packet.sequence.to_le_bytes());
        hasher.update(packet.source_channel.as_bytes());
        hasher.update(packet.dest_channel.as_bytes());
        hasher.update(&packet.data);
        hasher.update(packet.timeout_height.to_le_bytes());
        hasher.finalize().to_vec()
    }
}
