use serde::{Deserialize, Serialize};

use crate::types::{Address, BlockHeight, Timestamp};

/// A governance proposal.
#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Proposal {
    pub id: u64,
    pub title: String,
    pub description: String,
    pub proposer: Address,
    pub created_at: Timestamp,
    pub voting_end: BlockHeight,
    pub yes_votes: u64,
    pub no_votes: u64,
    pub abstain_votes: u64,
    pub executed: bool,
}

/// Vote options.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize, Deserialize)]
pub enum Vote {
    Yes,
    No,
    Abstain,
}

/// Governance operations trait.
pub trait Governance {
    fn create_proposal(
        &mut self,
        title: String,
        description: String,
        proposer: Address,
        voting_period: BlockHeight,
    ) -> Result<u64, GovernanceError>;

    fn vote(
        &mut self,
        proposal_id: u64,
        voter: Address,
        vote: Vote,
    ) -> Result<(), GovernanceError>;

    fn execute(&mut self, proposal_id: u64) -> Result<(), GovernanceError>;
}

#[derive(Debug, thiserror::Error)]
pub enum GovernanceError {
    #[error("proposal not found")]
    ProposalNotFound,
    #[error("voting period ended")]
    VotingPeriodEnded,
    #[error("proposal not passed")]
    ProposalNotPassed,
    #[error("already executed")]
    AlreadyExecuted,
}
