use super::*;
use crate::types::Address;

fn setup() -> GovernanceModule {
    let state = GovernanceState::default();
    GovernanceModule::new(state)
}

#[test]
fn test_create_proposal_with_deposit() {
    let mut module = setup();
    let proposer = Address::from_byte(1);

    let result = module.create_proposal(
        "Upgrade Network".to_string(),
        "Upgrade to v2".to_string(),
        ProposalType::SoftwareUpgrade,
        proposer,
        10_000,
        100,
    );

    assert!(result.is_ok());
    let id = result.unwrap();
    assert_eq!(id, 1);
    let proposal = module.state.proposals.get(&id).unwrap();
    assert_eq!(proposal.title, "Upgrade Network");
    assert_eq!(proposal.status, ProposalStatus::Voting);
    assert_eq!(proposal.deposit, 10_000);
    assert_eq!(proposal.voting_end, 100 + VOTING_PERIOD_BLOCKS);
}

#[test]
fn test_create_proposal_insufficient_deposit() {
    let mut module = setup();
    let proposer = Address::from_byte(1);

    let result = module.create_proposal(
        "Proposal".to_string(),
        "Desc".to_string(),
        ProposalType::Text,
        proposer,
        9_999,
        100,
    );

    assert_eq!(
        result,
        Err(GovernanceError::InsufficientDeposit(PROPOSAL_DEPOSIT))
    );
}

#[test]
fn test_voting_with_stake_weighted_power() {
    let mut module = setup();
    let proposer = Address::from_byte(1);
    let voter1 = Address::from_byte(2);
    let voter2 = Address::from_byte(3);

    let id = module
        .create_proposal(
            "Test".to_string(),
            "Desc".to_string(),
            ProposalType::Text,
            proposer,
            10_000,
            100,
        )
        .unwrap();

    // voter1 has 5000 staked, voter2 has 3000 staked
    module.vote(id, voter1, VoteOption::Yes, 5000, 150).unwrap();
    module.vote(id, voter2, VoteOption::No, 3000, 150).unwrap();

    let proposal = module.state.proposals.get(&id).unwrap();
    assert_eq!(proposal.votes_yes, 5000);
    assert_eq!(proposal.votes_no, 3000);
}

#[test]
fn test_double_vote_prevention() {
    let mut module = setup();
    let proposer = Address::from_byte(1);
    let voter = Address::from_byte(2);

    let id = module
        .create_proposal(
            "Test".to_string(),
            "Desc".to_string(),
            ProposalType::Text,
            proposer,
            10_000,
            100,
        )
        .unwrap();

    module.vote(id, voter, VoteOption::Yes, 1000, 150).unwrap();
    let result = module.vote(id, voter, VoteOption::No, 1000, 150);

    assert_eq!(result, Err(GovernanceError::AlreadyVoted));
}

#[test]
fn test_quorum_not_met_rejects() {
    let mut module = setup();
    let proposer = Address::from_byte(1);
    let voter = Address::from_byte(2);

    let id = module
        .create_proposal(
            "Test".to_string(),
            "Desc".to_string(),
            ProposalType::Text,
            proposer,
            10_000,
            100,
        )
        .unwrap();

    // Total voting power is 100_000, quorum is 33% = 33_000
    // Only vote with 1000 (not enough)
    module.vote(id, voter, VoteOption::Yes, 1000, 150).unwrap();

    let status = module
        .tally_votes(id, 100_000, 100 + VOTING_PERIOD_BLOCKS)
        .unwrap();
    assert_eq!(status, ProposalStatus::Rejected);
}

#[test]
fn test_majority_yes_passes() {
    let mut module = setup();
    let proposer = Address::from_byte(1);
    let voter1 = Address::from_byte(2);
    let voter2 = Address::from_byte(3);

    let id = module
        .create_proposal(
            "Test".to_string(),
            "Desc".to_string(),
            ProposalType::Text,
            proposer,
            10_000,
            100,
        )
        .unwrap();

    // Total voting power 100_000, need >33% participation and >50% Yes
    module
        .vote(id, voter1, VoteOption::Yes, 40_000, 150)
        .unwrap();
    module
        .vote(id, voter2, VoteOption::No, 10_000, 150)
        .unwrap();

    let status = module
        .tally_votes(id, 100_000, 100 + VOTING_PERIOD_BLOCKS)
        .unwrap();
    assert_eq!(status, ProposalStatus::Passed);
}

#[test]
fn test_veto_threshold_burns_deposit() {
    let mut module = setup();
    let proposer = Address::from_byte(1);
    let voter1 = Address::from_byte(2);
    let voter2 = Address::from_byte(3);

    let id = module
        .create_proposal(
            "Test".to_string(),
            "Desc".to_string(),
            ProposalType::Text,
            proposer,
            10_000,
            100,
        )
        .unwrap();

    // >33% NoWithVeto triggers veto
    module
        .vote(id, voter1, VoteOption::Yes, 20_000, 150)
        .unwrap();
    module
        .vote(id, voter2, VoteOption::NoWithVeto, 20_000, 150)
        .unwrap();

    let status = module
        .tally_votes(id, 100_000, 100 + VOTING_PERIOD_BLOCKS)
        .unwrap();
    assert_eq!(status, ProposalStatus::Vetoed);
    // Deposit pool should be empty (burned)
    assert_eq!(module.state.deposit_pool, 0);
}

#[test]
fn test_execution_after_timelock() {
    let mut module = setup();
    let proposer = Address::from_byte(1);
    let voter = Address::from_byte(2);
    let voting_start = 100;

    let id = module
        .create_proposal(
            "Test".to_string(),
            "Desc".to_string(),
            ProposalType::Text,
            proposer,
            10_000,
            voting_start,
        )
        .unwrap();

    module
        .vote(id, voter, VoteOption::Yes, 50_000, 150)
        .unwrap();

    let tally_height = voting_start + VOTING_PERIOD_BLOCKS;
    module.tally_votes(id, 100_000, tally_height).unwrap();

    // Try before timelock
    let result = module.execute_proposal(id, tally_height);
    assert_eq!(result, Err(GovernanceError::TimelockNotExpired));

    // Execute after timelock
    let exec_height = tally_height + TIMELOCK_PERIOD_BLOCKS;
    let result = module.execute_proposal(id, exec_height);
    assert!(result.is_ok());

    let proposal = module.state.proposals.get(&id).unwrap();
    assert_eq!(proposal.status, ProposalStatus::Executed);
}

#[test]
fn test_voting_after_period_ends_fails() {
    let mut module = setup();
    let proposer = Address::from_byte(1);
    let voter = Address::from_byte(2);

    let id = module
        .create_proposal(
            "Test".to_string(),
            "Desc".to_string(),
            ProposalType::Text,
            proposer,
            10_000,
            100,
        )
        .unwrap();

    // Vote after voting period
    let result = module.vote(id, voter, VoteOption::Yes, 1000, 100 + VOTING_PERIOD_BLOCKS);
    assert_eq!(result, Err(GovernanceError::VotingPeriodEnded));
}
