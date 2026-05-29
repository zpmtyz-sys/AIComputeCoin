use super::*;
use crate::types::Address;

fn setup() -> StakingModule {
    let state = StakingState::new(10);
    let mut module = StakingModule::new(state);
    // Register a validator
    let validator = Address::from_byte(1);
    module.register_validator(validator, 1000).unwrap();
    module
}

#[test]
fn test_delegate_below_minimum_fails() {
    let mut module = setup();
    let delegator = Address::from_byte(2);
    let validator = Address::from_byte(1);

    let result = module.delegate(delegator, validator, 999);

    assert_eq!(
        result,
        Err(StakingError::MinimumDelegationNotMet(MINIMUM_DELEGATION))
    );
}

#[test]
fn test_delegate_at_minimum_succeeds() {
    let mut module = setup();
    let delegator = Address::from_byte(2);
    let validator = Address::from_byte(1);

    let result = module.delegate(delegator, validator, 1000);

    assert!(result.is_ok());
    assert_eq!(
        module.state.delegations.get(&(delegator, validator)),
        Some(&1000)
    );
    assert_eq!(module.state.total_staked, 1000);
}

#[test]
fn test_delegate_to_unknown_validator_fails() {
    let mut module = setup();
    let delegator = Address::from_byte(2);
    let unknown = Address::from_byte(99);

    let result = module.delegate(delegator, unknown, 1000);

    assert_eq!(result, Err(StakingError::ValidatorNotFound));
}

#[test]
fn test_undelegate_creates_unbonding_entry() {
    let mut module = setup();
    let delegator = Address::from_byte(2);
    let validator = Address::from_byte(1);
    let current_height: u64 = 1000;

    module.delegate(delegator, validator, 5000).unwrap();
    module
        .undelegate(delegator, validator, 2000, current_height)
        .unwrap();

    let completion = current_height + UNBONDING_PERIOD_BLOCKS;
    let entries = module.state.unbonding_queue.get(&completion).unwrap();
    assert_eq!(entries.len(), 1);
    assert_eq!(entries[0].amount, 2000);
    assert_eq!(entries[0].delegator, delegator);
    assert_eq!(
        module.state.delegations.get(&(delegator, validator)),
        Some(&3000)
    );
}

#[test]
fn test_undelegate_more_than_delegated_fails() {
    let mut module = setup();
    let delegator = Address::from_byte(2);
    let validator = Address::from_byte(1);

    module.delegate(delegator, validator, 1000).unwrap();
    let result = module.undelegate(delegator, validator, 2000, 100);

    assert_eq!(result, Err(StakingError::InsufficientBalance));
}

#[test]
fn test_rewards_proportional_to_stake_times_cu() {
    let mut module = setup();
    let validator = Address::from_byte(1);
    let alice = Address::from_byte(2);
    let bob = Address::from_byte(3);

    module.set_cu_contribution(validator, 100).unwrap();
    module.delegate(alice, validator, 3000).unwrap();
    module.delegate(bob, validator, 7000).unwrap();

    module.distribute_rewards(10_000).unwrap();

    let alice_rewards = module
        .state
        .pending_rewards
        .get(&alice)
        .copied()
        .unwrap_or(0);
    let bob_rewards = module.state.pending_rewards.get(&bob).copied().unwrap_or(0);

    // Alice has 30% of stake, Bob has 70%
    assert_eq!(alice_rewards, 3000);
    assert_eq!(bob_rewards, 7000);
}

#[test]
fn test_claim_rewards() {
    let mut module = setup();
    let validator = Address::from_byte(1);
    let alice = Address::from_byte(2);

    module.set_cu_contribution(validator, 100).unwrap();
    module.delegate(alice, validator, 5000).unwrap();
    module.distribute_rewards(1000).unwrap();

    let claimed = module.claim_rewards(alice).unwrap();
    assert_eq!(claimed, 1000);

    // Second claim should fail
    let result = module.claim_rewards(alice);
    assert_eq!(result, Err(StakingError::NoRewardsAvailable));
}

#[test]
fn test_validator_set_selects_top_n() {
    let state = StakingState::new(2); // Only top 2
    let mut module = StakingModule::new(state);

    let v1 = Address::from_byte(1);
    let v2 = Address::from_byte(2);
    let v3 = Address::from_byte(3);
    let delegator = Address::from_byte(10);

    module.register_validator(v1, 1000).unwrap();
    module.register_validator(v2, 1000).unwrap();
    module.register_validator(v3, 1000).unwrap();

    // Give them different stakes (all above MINIMUM_VALIDATOR_STAKE)
    module.delegate(delegator, v1, 500_000).unwrap();
    module.delegate(delegator, v2, 300_000).unwrap();
    module.delegate(delegator, v3, 100_000).unwrap();

    module.update_validator_set();

    assert!(module.state.validators.get(&v1).unwrap().active);
    assert!(module.state.validators.get(&v2).unwrap().active);
    assert!(!module.state.validators.get(&v3).unwrap().active);
}

#[test]
fn test_slashing_reduces_stake() {
    let mut module = setup();
    let validator = Address::from_byte(1);
    let delegator = Address::from_byte(2);

    module.delegate(delegator, validator, 10_000).unwrap();
    // Slash 5% (500 basis points)
    let slashed = module.slash(validator, 500).unwrap();

    assert_eq!(slashed, 500);
    assert_eq!(
        module.state.validators.get(&validator).unwrap().total_stake,
        9500
    );
    assert_eq!(module.state.total_staked, 9500);
    // Delegator's stake should also be reduced
    assert_eq!(
        module.state.delegations.get(&(delegator, validator)),
        Some(&9500)
    );
}

#[test]
fn test_unbonding_completes_after_period() {
    let mut module = setup();
    let delegator = Address::from_byte(2);
    let validator = Address::from_byte(1);

    module.delegate(delegator, validator, 5000).unwrap();
    module.undelegate(delegator, validator, 2000, 1000).unwrap();

    // Before completion
    let completed = module.process_unbonding(1000 + UNBONDING_PERIOD_BLOCKS - 1);
    assert!(completed.is_empty());

    // At completion
    let completed = module.process_unbonding(1000 + UNBONDING_PERIOD_BLOCKS);
    assert_eq!(completed.len(), 1);
    assert_eq!(completed[0].amount, 2000);
}

#[test]
fn test_delegate_emits_event() {
    let mut module = setup();
    let delegator = Address::from_byte(2);
    let validator = Address::from_byte(1);

    module.delegate(delegator, validator, 5000).unwrap();

    let events = module.events();
    assert!(events.contains(&StakingEvent::Delegated {
        delegator,
        validator,
        amount: 5000,
    }));
}
