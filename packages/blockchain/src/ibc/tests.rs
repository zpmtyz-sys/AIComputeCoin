use super::*;
use crate::types::Address;

fn setup_channel_manager() -> ChannelManager {
    let mut cm = ChannelManager::new();
    cm.open_init(
        "channel-0".to_string(),
        "transfer".to_string(),
        "channel-1".to_string(),
        "transfer".to_string(),
    )
    .unwrap();
    cm.open_try("channel-0").unwrap();
    cm.open_ack("channel-0").unwrap();
    cm
}

#[test]
fn test_channel_handshake_full_lifecycle() {
    let mut cm = ChannelManager::new();

    // Init
    cm.open_init(
        "ch-0".to_string(),
        "transfer".to_string(),
        "ch-1".to_string(),
        "transfer".to_string(),
    )
    .unwrap();
    assert_eq!(cm.get_channel("ch-0").unwrap().state, ChannelState::Init);

    // TryOpen
    cm.open_try("ch-0").unwrap();
    assert_eq!(cm.get_channel("ch-0").unwrap().state, ChannelState::TryOpen);

    // Ack
    cm.open_ack("ch-0").unwrap();
    assert_eq!(cm.get_channel("ch-0").unwrap().state, ChannelState::Open);

    // Confirm
    cm.open_confirm("ch-0").unwrap();
    // Still open after confirm
    assert_eq!(cm.get_channel("ch-0").unwrap().state, ChannelState::Open);

    // Close
    cm.close_init("ch-0").unwrap();
    assert_eq!(cm.get_channel("ch-0").unwrap().state, ChannelState::Closed);
}

#[test]
fn test_invalid_channel_state_transitions() {
    let mut cm = ChannelManager::new();
    cm.open_init(
        "ch-0".to_string(),
        "transfer".to_string(),
        "ch-1".to_string(),
        "transfer".to_string(),
    )
    .unwrap();

    // Cannot ack from Init (must be TryOpen)
    assert_eq!(cm.open_ack("ch-0"), Err(IbcError::InvalidChannelState));

    // Cannot close from Init
    assert_eq!(cm.close_init("ch-0"), Err(IbcError::InvalidChannelState));
}

#[test]
fn test_packet_send_receive_acknowledge() {
    let cm = setup_channel_manager();
    let mut relay = PacketRelay::new(cm);

    // Also need the dest channel
    relay
        .channel_manager
        .open_init(
            "channel-1".to_string(),
            "transfer".to_string(),
            "channel-0".to_string(),
            "transfer".to_string(),
        )
        .unwrap();
    relay.channel_manager.open_try("channel-1").unwrap();
    relay.channel_manager.open_ack("channel-1").unwrap();

    // Send packet
    let packet = relay
        .send_packet(
            "transfer".to_string(),
            "channel-0".to_string(),
            "transfer".to_string(),
            "channel-1".to_string(),
            b"hello".to_vec(),
            1000,
            0,
        )
        .unwrap();
    assert_eq!(packet.sequence, 1);

    // Receive packet
    let ack = relay.recv_packet(&packet).unwrap();
    assert_eq!(ack, Acknowledgement::Success(b"ok".to_vec()));

    // Acknowledge packet
    relay.acknowledge_packet("channel-0", 1, &ack).unwrap();
    // Commitment should be removed
    assert!(relay.commitments.is_empty());
}

#[test]
fn test_packet_send_on_closed_channel_fails() {
    let mut cm = setup_channel_manager();
    cm.close_init("channel-0").unwrap();
    let mut relay = PacketRelay::new(cm);

    let result = relay.send_packet(
        "transfer".to_string(),
        "channel-0".to_string(),
        "transfer".to_string(),
        "channel-1".to_string(),
        b"hello".to_vec(),
        1000,
        0,
    );
    assert_eq!(result, Err(IbcError::ChannelNotOpen));
}

#[test]
fn test_token_transfer_lock_and_mint() {
    let mut tt = TokenTransfer::new();
    let sender = Address::from_byte(1);
    let receiver = Address::from_byte(2);

    let msg = TransferMessage {
        sender,
        receiver,
        amount: 5000,
        denom: "CC".to_string(),
        source_channel: "channel-0".to_string(),
    };

    // Send (locks tokens)
    let packet_data = tt.send_transfer(msg, 10_000, "channel-0", 1).unwrap();

    assert_eq!(tt.get_locked(&sender, "CC"), 5000);

    // Receive on dest chain (mints tokens)
    tt.recv_transfer(&packet_data).unwrap();
    assert_eq!(tt.get_minted(&receiver, "CC"), 5000);
}

#[test]
fn test_token_transfer_insufficient_balance() {
    let mut tt = TokenTransfer::new();
    let sender = Address::from_byte(1);
    let receiver = Address::from_byte(2);

    let msg = TransferMessage {
        sender,
        receiver,
        amount: 5000,
        denom: "CC".to_string(),
        source_channel: "channel-0".to_string(),
    };

    let result = tt.send_transfer(msg, 1000, "channel-0", 1);
    assert_eq!(result, Err(IbcError::InsufficientBalance));
}

#[test]
fn test_timeout_triggers_refund() {
    let mut tt = TokenTransfer::new();
    let sender = Address::from_byte(1);
    let receiver = Address::from_byte(2);

    let msg = TransferMessage {
        sender,
        receiver,
        amount: 3000,
        denom: "CC".to_string(),
        source_channel: "channel-0".to_string(),
    };

    tt.send_transfer(msg, 10_000, "channel-0", 1).unwrap();
    assert_eq!(tt.get_locked(&sender, "CC"), 3000);

    // Refund on timeout
    tt.refund_transfer("channel-0", 1).unwrap();
    assert_eq!(tt.get_locked(&sender, "CC"), 0);
    assert!(tt.events.iter().any(
        |e| matches!(e, IbcEvent::TokensRefunded { receiver: r, amount: 3000 } if *r == sender)
    ));
}

#[test]
fn test_packet_timeout_processing() {
    let cm = setup_channel_manager();
    let mut relay = PacketRelay::new(cm);

    // Send packet with timeout at block 500
    let _packet = relay
        .send_packet(
            "transfer".to_string(),
            "channel-0".to_string(),
            "transfer".to_string(),
            "channel-1".to_string(),
            b"data".to_vec(),
            500,
            0,
        )
        .unwrap();

    // Cannot timeout before timeout height
    let result = relay.timeout_packet("channel-0", 1, 499, 500);
    assert_eq!(result, Err(IbcError::PacketTimeout));

    // Can timeout at or after timeout height
    relay.timeout_packet("channel-0", 1, 500, 500).unwrap();
    assert!(relay.commitments.is_empty());
}

#[test]
fn test_channel_not_found() {
    let cm = ChannelManager::new();
    let mut relay = PacketRelay::new(cm);

    let result = relay.send_packet(
        "transfer".to_string(),
        "nonexistent".to_string(),
        "transfer".to_string(),
        "channel-1".to_string(),
        b"data".to_vec(),
        1000,
        0,
    );
    assert_eq!(
        result,
        Err(IbcError::ChannelNotFound("nonexistent".to_string()))
    );
}
