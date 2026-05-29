use super::*;
use crate::types::Address;
use sha2::{Digest, Sha256};

fn make_oracle_key_and_sig(node_id: &Address, cu_value: u64) -> ([u8; 32], [u8; 64]) {
    // Create a signature
    let mut sig = [0u8; 64];
    sig[0] = 0xAB;
    sig[1] = 0xCD;
    // Compute the sig hash the same way the PoC does
    let mut hasher = Sha256::new();
    hasher.update(sig);
    hasher.update(node_id.0);
    hasher.update(cu_value.to_le_bytes());
    let result = hasher.finalize();
    let mut key_hash = [0u8; 32];
    // The first 16 bytes must match for verification
    key_hash[..16].copy_from_slice(&result[..16]);
    (key_hash, sig)
}

#[test]
fn test_block_production_weight_calculation() {
    let weight = ProofOfCompute::block_production_weight(10_000, 500);
    assert_eq!(weight, 5_000_000);

    let weight = ProofOfCompute::block_production_weight(0, 1000);
    assert_eq!(weight, 0);

    let weight = ProofOfCompute::block_production_weight(1000, 0);
    assert_eq!(weight, 0);
}

#[test]
fn test_verify_cu_report_valid() {
    let node_id = Address::from_byte(1);
    let (key_hash, sig) = make_oracle_key_and_sig(&node_id, 500);

    let mut poc = ProofOfCompute::new(1, 10_000);
    poc.add_oracle_key(key_hash);

    let report = CUReport {
        node_id,
        cu_value: 500,
        oracle_signatures: vec![sig],
        timestamp: 1000,
    };

    assert!(poc.verify_cu_report(&report).is_ok());
}

#[test]
fn test_verify_cu_report_no_signatures() {
    let poc = ProofOfCompute::new(1, 10_000);
    let report = CUReport {
        node_id: Address::from_byte(1),
        cu_value: 500,
        oracle_signatures: vec![],
        timestamp: 1000,
    };

    assert_eq!(
        poc.verify_cu_report(&report),
        Err(ConsensusError::NoSignatures)
    );
}

#[test]
fn test_verify_cu_report_cu_out_of_range() {
    let node_id = Address::from_byte(1);
    let (key_hash, sig) = make_oracle_key_and_sig(&node_id, 500);

    let mut poc = ProofOfCompute::new(1, 100);
    poc.add_oracle_key(key_hash);

    let report = CUReport {
        node_id,
        cu_value: 500,
        oracle_signatures: vec![sig],
        timestamp: 1000,
    };

    assert_eq!(
        poc.verify_cu_report(&report),
        Err(ConsensusError::CUValueOutOfRange)
    );
}

#[test]
fn test_verify_cu_report_invalid_signature() {
    let mut poc = ProofOfCompute::new(1, 10_000);
    poc.add_oracle_key([0xFF; 32]); // Key that won't match any signature

    let report = CUReport {
        node_id: Address::from_byte(1),
        cu_value: 500,
        oracle_signatures: vec![[0x00; 64]],
        timestamp: 1000,
    };

    assert_eq!(
        poc.verify_cu_report(&report),
        Err(ConsensusError::InvalidSignature)
    );
}

#[test]
fn test_block_reward_proportional_to_cu() {
    // Node contributes 25% of total CU
    let reward = ProofOfCompute::calculate_block_reward(250, 1000).unwrap();
    assert_eq!(reward.base, 1_000_000);
    assert_eq!(reward.variable, 1_000_000); // 250/1000 * 4_000_000
    assert_eq!(reward.total, 2_000_000);

    // Node contributes 100% of total CU
    let reward = ProofOfCompute::calculate_block_reward(1000, 1000).unwrap();
    assert_eq!(reward.variable, 4_000_000);
    assert_eq!(reward.total, 5_000_000);

    // Zero total CU
    let reward = ProofOfCompute::calculate_block_reward(100, 0).unwrap();
    assert_eq!(reward.variable, 0);
    assert_eq!(reward.total, 1_000_000);
}

#[test]
fn test_epoch_transition() {
    let mut poc = ProofOfCompute::new(1, 10_000);
    poc.validator_set
        .set_validator(Address::from_byte(1), 10_000, 500);

    // No transition at block 0
    assert!(poc.check_epoch_transition(0).is_none());

    // Transition at block 100
    let new_epoch = poc.check_epoch_transition(100);
    assert_eq!(new_epoch, Some(1));
    assert_eq!(poc.current_epoch.epoch_number, 1);
    assert_eq!(poc.current_epoch.start_block, 100);
    assert_eq!(poc.current_epoch.total_cu, 500);
}

#[test]
fn test_validator_selection_by_combined_score() {
    let mut vs = ValidatorSet::with_max_active(10);
    let v1 = Address::from_byte(1);
    let v2 = Address::from_byte(2);
    let v3 = Address::from_byte(3);

    vs.set_validator(v1, 10_000, 100); // weight = 1_000_000
    vs.set_validator(v2, 5_000, 200); // weight = 1_000_000
    vs.set_validator(v3, 1_000, 50); // weight = 50_000
                                     // Total weight = 2,050,000

    // Seed that maps to v1 range (targets 0..999_999 select v1)
    // seed % 2_050_000 < 1_000_000 -> v1
    let producer = vs.select_producer(500_000);
    assert_eq!(producer, Some(v1));

    // Seed that maps to v2 range (targets 1_000_000..1_999_999 select v2)
    let producer = vs.select_producer(1_500_000);
    assert_eq!(producer, Some(v2));

    // Seed that maps to v3 range (targets 2_000_000..2_049_999 select v3)
    let producer = vs.select_producer(2_000_001);
    assert_eq!(producer, Some(v3));

    // Verify weight proportionality: v1 and v2 have equal weight, both > v3
    let mut v1_count = 0u64;
    let mut v2_count = 0u64;
    let mut v3_count = 0u64;
    // Use seeds spread across the full weight range
    for i in 0..2_050u64 {
        let seed = i * 1000;
        match vs.select_producer(seed) {
            Some(addr) if addr == v1 => v1_count += 1,
            Some(addr) if addr == v2 => v2_count += 1,
            Some(addr) if addr == v3 => v3_count += 1,
            _ => {}
        }
    }
    // v1 and v2 should each have ~48.8% of selections, v3 ~2.4%
    assert!(v1_count > v3_count);
    assert!(v2_count > v3_count);
}

#[test]
fn test_validator_set_recalculate() {
    let mut vs = ValidatorSet::with_max_active(2);
    let v1 = Address::from_byte(1);
    let v2 = Address::from_byte(2);
    let v3 = Address::from_byte(3);

    vs.set_validator(v1, 10_000, 100); // score = 1_000_000
    vs.set_validator(v2, 5_000, 300); // score = 1_500_000
    vs.set_validator(v3, 1_000, 50); // score = 50_000

    vs.recalculate();

    let active = vs.get_active_validators();
    assert!(active.contains(&v1));
    assert!(active.contains(&v2));
    assert!(!active.contains(&v3));
}
