use crate::types::{Address, Balance, BlockHeight, ProposalId};

use super::state::GovernanceState;
use super::types::{
    GovernanceError, GovernanceEvent, Proposal, ProposalStatus, ProposalType, VoteOption,
    PASS_THRESHOLD_BPS, PROPOSAL_DEPOSIT, QUORUM_THRESHOLD_BPS, TIMELOCK_PERIOD_BLOCKS,
    VETO_THRESHOLD_BPS, VOTING_PERIOD_BLOCKS,
};

/// Governance module providing proposal creation, voting, tallying, and execution.
pub struct GovernanceModule {
    pub state: GovernanceState,
}

impl GovernanceModule {
    pub fn new(state: GovernanceState) -> Self {
        Self { state }
    }

    /// Create a proposal. Requires deposit of PROPOSAL_DEPOSIT tokens.
    pub fn create_proposal(
        &mut self,
        title: String,
        description: String,
        proposal_type: ProposalType,
        proposer: Address,
        deposit: Balance,
        current_height: BlockHeight,
    ) -> Result<ProposalId, GovernanceError> {
        if deposit < PROPOSAL_DEPOSIT {
            return Err(GovernanceError::InsufficientDeposit(PROPOSAL_DEPOSIT));
        }
        let id = self.state.next_proposal_id;
        self.state.next_proposal_id = id.checked_add(1).ok_or(GovernanceError::Overflow)?;
        let voting_end = current_height
            .checked_add(VOTING_PERIOD_BLOCKS)
            .ok_or(GovernanceError::Overflow)?;
        let proposal = Proposal {
            id,
            title,
            description,
            proposal_type,
            proposer,
            deposit,
            status: ProposalStatus::Voting,
            created_at: current_height,
            voting_end,
            execution_time: None,
            votes_yes: 0,
            votes_no: 0,
            votes_abstain: 0,
            votes_veto: 0,
        };
        self.state.proposals.insert(id, proposal);
        self.state.deposit_pool = self
            .state
            .deposit_pool
            .checked_add(deposit)
            .ok_or(GovernanceError::Overflow)?;
        self.state
            .events
            .push(GovernanceEvent::ProposalCreated { id, proposer });
        Ok(id)
    }

    /// Vote on a proposal. Weight is the voter's staked token amount.
    pub fn vote(
        &mut self,
        proposal_id: ProposalId,
        voter: Address,
        option: VoteOption,
        voting_power: Balance,
        current_height: BlockHeight,
    ) -> Result<(), GovernanceError> {
        let proposal = self
            .state
            .proposals
            .get(&proposal_id)
            .ok_or(GovernanceError::ProposalNotFound)?;
        if current_height >= proposal.voting_end {
            return Err(GovernanceError::VotingPeriodEnded);
        }
        if proposal.status != ProposalStatus::Voting {
            return Err(GovernanceError::VotingPeriodEnded);
        }
        if voting_power == 0 {
            return Err(GovernanceError::NoVotingPower);
        }
        if self.state.votes.contains_key(&(proposal_id, voter)) {
            return Err(GovernanceError::AlreadyVoted);
        }
        self.state.votes.insert((proposal_id, voter), option);
        let proposal = self
            .state
            .proposals
            .get_mut(&proposal_id)
            .ok_or(GovernanceError::ProposalNotFound)?;
        match option {
            VoteOption::Yes => {
                proposal.votes_yes = proposal
                    .votes_yes
                    .checked_add(voting_power)
                    .ok_or(GovernanceError::Overflow)?;
            }
            VoteOption::No => {
                proposal.votes_no = proposal
                    .votes_no
                    .checked_add(voting_power)
                    .ok_or(GovernanceError::Overflow)?;
            }
            VoteOption::Abstain => {
                proposal.votes_abstain = proposal
                    .votes_abstain
                    .checked_add(voting_power)
                    .ok_or(GovernanceError::Overflow)?;
            }
            VoteOption::NoWithVeto => {
                proposal.votes_veto = proposal
                    .votes_veto
                    .checked_add(voting_power)
                    .ok_or(GovernanceError::Overflow)?;
            }
        }
        self.state.events.push(GovernanceEvent::Voted {
            proposal_id,
            voter,
            option,
            weight: voting_power,
        });
        Ok(())
    }

    /// Tally votes and determine proposal outcome.
    pub fn tally_votes(
        &mut self,
        proposal_id: ProposalId,
        total_voting_power: Balance,
        current_height: BlockHeight,
    ) -> Result<ProposalStatus, GovernanceError> {
        let proposal = self
            .state
            .proposals
            .get(&proposal_id)
            .ok_or(GovernanceError::ProposalNotFound)?;
        if current_height < proposal.voting_end {
            return Err(GovernanceError::VotingPeriodNotEnded);
        }
        if proposal.status != ProposalStatus::Voting {
            return Err(GovernanceError::AlreadyExecuted);
        }
        let total_voted = proposal
            .votes_yes
            .checked_add(proposal.votes_no)
            .and_then(|s| s.checked_add(proposal.votes_abstain))
            .and_then(|s| s.checked_add(proposal.votes_veto))
            .ok_or(GovernanceError::Overflow)?;

        // Check quorum (33% of total voting power must participate)
        let quorum_required = total_voting_power
            .checked_mul(QUORUM_THRESHOLD_BPS as u128)
            .ok_or(GovernanceError::Overflow)?
            / 10_000;
        if total_voted < quorum_required {
            let proposal = self
                .state
                .proposals
                .get_mut(&proposal_id)
                .ok_or(GovernanceError::ProposalNotFound)?;
            proposal.status = ProposalStatus::Rejected;
            self.state
                .events
                .push(GovernanceEvent::ProposalRejected { id: proposal_id });
            return Ok(ProposalStatus::Rejected);
        }

        // Check veto threshold (>33% NoWithVeto of total voted)
        let veto_threshold = total_voted
            .checked_mul(VETO_THRESHOLD_BPS as u128)
            .ok_or(GovernanceError::Overflow)?
            / 10_000;
        let proposal = self
            .state
            .proposals
            .get(&proposal_id)
            .ok_or(GovernanceError::ProposalNotFound)?;
        let deposit = proposal.deposit;

        if proposal.votes_veto > veto_threshold {
            let proposal = self
                .state
                .proposals
                .get_mut(&proposal_id)
                .ok_or(GovernanceError::ProposalNotFound)?;
            proposal.status = ProposalStatus::Vetoed;
            // Burn deposit
            self.state.deposit_pool = self
                .state
                .deposit_pool
                .checked_sub(deposit)
                .ok_or(GovernanceError::Overflow)?;
            self.state.events.push(GovernanceEvent::ProposalVetoed {
                id: proposal_id,
                deposit_burned: deposit,
            });
            return Ok(ProposalStatus::Vetoed);
        }

        // Check pass threshold (>50% of non-abstain votes)
        let non_abstain = proposal
            .votes_yes
            .checked_add(proposal.votes_no)
            .and_then(|s| s.checked_add(proposal.votes_veto))
            .ok_or(GovernanceError::Overflow)?;
        let pass_threshold = non_abstain
            .checked_mul(PASS_THRESHOLD_BPS as u128)
            .ok_or(GovernanceError::Overflow)?
            / 10_000;

        if proposal.votes_yes > pass_threshold {
            let proposal = self
                .state
                .proposals
                .get_mut(&proposal_id)
                .ok_or(GovernanceError::ProposalNotFound)?;
            let execution_time = current_height
                .checked_add(TIMELOCK_PERIOD_BLOCKS)
                .ok_or(GovernanceError::Overflow)?;
            proposal.status = ProposalStatus::Passed;
            proposal.execution_time = Some(execution_time);
            self.state
                .events
                .push(GovernanceEvent::ProposalPassed { id: proposal_id });
            Ok(ProposalStatus::Passed)
        } else {
            let proposal = self
                .state
                .proposals
                .get_mut(&proposal_id)
                .ok_or(GovernanceError::ProposalNotFound)?;
            proposal.status = ProposalStatus::Rejected;
            self.state
                .events
                .push(GovernanceEvent::ProposalRejected { id: proposal_id });
            Ok(ProposalStatus::Rejected)
        }
    }

    /// Execute a passed proposal after the timelock period.
    pub fn execute_proposal(
        &mut self,
        proposal_id: ProposalId,
        current_height: BlockHeight,
    ) -> Result<(), GovernanceError> {
        let proposal = self
            .state
            .proposals
            .get(&proposal_id)
            .ok_or(GovernanceError::ProposalNotFound)?;
        if proposal.status == ProposalStatus::Executed {
            return Err(GovernanceError::AlreadyExecuted);
        }
        if proposal.status != ProposalStatus::Passed {
            return Err(GovernanceError::ProposalNotPassed);
        }
        let execution_time = proposal
            .execution_time
            .ok_or(GovernanceError::TimelockNotExpired)?;
        if current_height < execution_time {
            return Err(GovernanceError::TimelockNotExpired);
        }
        let proposal = self
            .state
            .proposals
            .get_mut(&proposal_id)
            .ok_or(GovernanceError::ProposalNotFound)?;
        proposal.status = ProposalStatus::Executed;
        // Return deposit
        let deposit = proposal.deposit;
        self.state.deposit_pool = self
            .state
            .deposit_pool
            .checked_sub(deposit)
            .ok_or(GovernanceError::Overflow)?;
        self.state
            .events
            .push(GovernanceEvent::ProposalExecuted { id: proposal_id });
        Ok(())
    }

    /// Get events.
    pub fn events(&self) -> &[GovernanceEvent] {
        &self.state.events
    }
}
