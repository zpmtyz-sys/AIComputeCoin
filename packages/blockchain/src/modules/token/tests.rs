use super::*;
use crate::types::Address;

fn setup() -> TokenModule {
    let state = TokenState::default();
    let mut module = TokenModule::new(state);
    module.add_minter(Address::from_byte(1));
    module
}

#[test]
fn test_mint_increases_supply_and_balance() {
    let mut module = setup();
    let minter = Address::from_byte(1);
    let recipient = Address::from_byte(2);

    module.mint(&minter, recipient, 1000).unwrap();

    assert_eq!(module.balance_of(&recipient), 1000);
    assert_eq!(module.total_supply(), 1000);
}

#[test]
fn test_mint_emits_event() {
    let mut module = setup();
    let minter = Address::from_byte(1);
    let recipient = Address::from_byte(2);

    module.mint(&minter, recipient, 500).unwrap();

    assert_eq!(
        module.events(),
        &[TokenEvent::Mint {
            to: recipient,
            amount: 500,
        }]
    );
}

#[test]
fn test_unauthorized_mint_fails() {
    let mut module = setup();
    let unauthorized = Address::from_byte(99);
    let recipient = Address::from_byte(2);

    let result = module.mint(&unauthorized, recipient, 1000);

    assert_eq!(result, Err(TokenError::UnauthorizedMinter));
    assert_eq!(module.total_supply(), 0);
}

#[test]
fn test_mint_zero_amount_fails() {
    let mut module = setup();
    let minter = Address::from_byte(1);
    let recipient = Address::from_byte(2);

    let result = module.mint(&minter, recipient, 0);

    assert_eq!(result, Err(TokenError::ZeroAmount));
}

#[test]
fn test_burn_reduces_supply_and_balance() {
    let mut module = setup();
    let minter = Address::from_byte(1);
    let account = Address::from_byte(2);

    module.mint(&minter, account, 1000).unwrap();
    module.burn(account, 400).unwrap();

    assert_eq!(module.balance_of(&account), 600);
    assert_eq!(module.total_supply(), 600);
}

#[test]
fn test_burn_emits_event() {
    let mut module = setup();
    let minter = Address::from_byte(1);
    let account = Address::from_byte(2);

    module.mint(&minter, account, 1000).unwrap();
    module.burn(account, 300).unwrap();

    assert_eq!(module.events().len(), 2);
    assert_eq!(
        module.events()[1],
        TokenEvent::Burn {
            from: account,
            amount: 300,
        }
    );
}

#[test]
fn test_burn_insufficient_balance_fails() {
    let mut module = setup();
    let minter = Address::from_byte(1);
    let account = Address::from_byte(2);

    module.mint(&minter, account, 100).unwrap();
    let result = module.burn(account, 200);

    assert_eq!(result, Err(TokenError::InsufficientBalance));
    assert_eq!(module.balance_of(&account), 100);
    assert_eq!(module.total_supply(), 100);
}

#[test]
fn test_transfer_moves_funds() {
    let mut module = setup();
    let minter = Address::from_byte(1);
    let alice = Address::from_byte(2);
    let bob = Address::from_byte(3);

    module.mint(&minter, alice, 1000).unwrap();
    module.transfer(alice, bob, 300).unwrap();

    assert_eq!(module.balance_of(&alice), 700);
    assert_eq!(module.balance_of(&bob), 300);
    assert_eq!(module.total_supply(), 1000);
}

#[test]
fn test_transfer_emits_event() {
    let mut module = setup();
    let minter = Address::from_byte(1);
    let alice = Address::from_byte(2);
    let bob = Address::from_byte(3);

    module.mint(&minter, alice, 1000).unwrap();
    module.transfer(alice, bob, 200).unwrap();

    assert_eq!(
        module.events()[1],
        TokenEvent::Transfer {
            from: alice,
            to: bob,
            amount: 200,
        }
    );
}

#[test]
fn test_transfer_insufficient_balance_fails() {
    let mut module = setup();
    let minter = Address::from_byte(1);
    let alice = Address::from_byte(2);
    let bob = Address::from_byte(3);

    module.mint(&minter, alice, 100).unwrap();
    let result = module.transfer(alice, bob, 200);

    assert_eq!(result, Err(TokenError::InsufficientBalance));
    assert_eq!(module.balance_of(&alice), 100);
    assert_eq!(module.balance_of(&bob), 0);
}

#[test]
fn test_total_supply_tracks_across_operations() {
    let mut module = setup();
    let minter = Address::from_byte(1);
    let alice = Address::from_byte(2);
    let bob = Address::from_byte(3);

    module.mint(&minter, alice, 5000).unwrap();
    assert_eq!(module.total_supply(), 5000);

    module.mint(&minter, bob, 3000).unwrap();
    assert_eq!(module.total_supply(), 8000);

    module.burn(alice, 2000).unwrap();
    assert_eq!(module.total_supply(), 6000);

    module.transfer(alice, bob, 1000).unwrap();
    assert_eq!(module.total_supply(), 6000); // transfer doesn't change supply
}

#[test]
fn test_metadata_defaults() {
    let module = setup();
    assert_eq!(module.state.metadata.name, "ComputeCoin");
    assert_eq!(module.state.metadata.symbol, "CC");
    assert_eq!(module.state.metadata.decimals, 18);
}
