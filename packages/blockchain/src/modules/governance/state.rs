use std::collections::BTreeMap;

use crate::types::{Address, Balance, ProposalId};

use super::types::{GovernanceEvent, Proposal, VoteOption};

/// Governance state.
#[derive(Debug, Clone)]
pub struct GovernanceState {
    /// All proposals by ID.
    pub proposals: BTreeMap<ProposalId, Proposal>,
    /// Votes: (proposal_id, voter) -> vote option
    pub votes: BTreeMap<(ProposalId, Address), VoteOption>,
    /// Deposit pool tracking (burned on veto)
    pub deposit_pool: Balance,
    /// Next proposal ID counter
    pub next_proposal_id: ProposalId,
    /// Event log
    pub events: Vec<GovernanceEvent>,
}

impl GovernanceState {
    pub fn new() -> Self {
        Self {
            proposals: BTreeMap::new(),
            votes: BTreeMap::new(),
            deposit_pool: 0,
            next_proposal_id: 1,
            events: Vec::new(),
        }
    }
}

impl Default for GovernanceState {
    fn default() -> Self {
        Self::new()
    }
}
