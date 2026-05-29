use serde::{Deserialize, Serialize};

use crate::types::{Address, Balance, BlockHeight, ProposalId};

/// Required deposit to create a proposal (10,000 CC).
pub const PROPOSAL_DEPOSIT: Balance = 10_000;

/// Voting period in blocks (~7 days at 6s block time).
pub const VOTING_PERIOD_BLOCKS: BlockHeight = 100_800;

/// Timelock period in blocks (~2 days at 6s block time).
pub const TIMELOCK_PERIOD_BLOCKS: BlockHeight = 28_800;

/// Quorum threshold: 33% of total voting power.
pub const QUORUM_THRESHOLD_BPS: u32 = 3300;

/// Pass threshold: >50% Yes of non-abstain votes.
pub const PASS_THRESHOLD_BPS: u32 = 5000;

/// Veto threshold: >33% NoWithVeto of total votes.
pub const VETO_THRESHOLD_BPS: u32 = 3300;

/// Proposal type.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum ProposalType {
    ParameterChange,
    SoftwareUpgrade,
    CommunitySpend,
    Text,
}

/// Proposal status.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum ProposalStatus {
    Voting,
    Passed,
    Rejected,
    Vetoed,
    Executed,
}

/// A governance proposal.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Proposal {
    pub id: ProposalId,
    pub title: String,
    pub description: String,
    pub proposal_type: ProposalType,
    pub proposer: Address,
    pub deposit: Balance,
    pub status: ProposalStatus,
    pub created_at: BlockHeight,
    pub voting_end: BlockHeight,
    pub execution_time: Option<BlockHeight>,
    pub votes_yes: Balance,
    pub votes_no: Balance,
    pub votes_abstain: Balance,
    pub votes_veto: Balance,
}

/// Vote option.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum VoteOption {
    Yes,
    No,
    Abstain,
    NoWithVeto,
}

/// Events emitted by governance operations.
#[derive(Debug, Clone, PartialEq, Eq, Serialize, Deserialize)]
pub enum GovernanceEvent {
    ProposalCreated {
        id: ProposalId,
        proposer: Address,
    },
    Voted {
        proposal_id: ProposalId,
        voter: Address,
        option: VoteOption,
        weight: Balance,
    },
    ProposalPassed {
        id: ProposalId,
    },
    ProposalRejected {
        id: ProposalId,
    },
    ProposalVetoed {
        id: ProposalId,
        deposit_burned: Balance,
    },
    ProposalExecuted {
        id: ProposalId,
    },
}

/// Governance operation errors.
#[derive(Debug, thiserror::Error, PartialEq, Eq)]
pub enum GovernanceError {
    #[error("insufficient deposit (requires {0} CC)")]
    InsufficientDeposit(Balance),
    #[error("proposal not found")]
    ProposalNotFound,
    #[error("voting period ended")]
    VotingPeriodEnded,
    #[error("voting period not ended")]
    VotingPeriodNotEnded,
    #[error("already voted")]
    AlreadyVoted,
    #[error("no voting power")]
    NoVotingPower,
    #[error("proposal not passed")]
    ProposalNotPassed,
    #[error("timelock not expired")]
    TimelockNotExpired,
    #[error("already executed")]
    AlreadyExecuted,
    #[error("quorum not reached")]
    QuorumNotReached,
    #[error("overflow")]
    Overflow,
}
