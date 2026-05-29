# Phase 5: Blockchain Layer - Custom Chain Modules

## Overview

Build the blockchain layer in Rust. This implements the custom chain modules for ComputeCoin including the CC token, staking, governance, and Proof-of-Compute (PoC) consensus integration. The chain validates compute proofs from oracle nodes and rewards miners accordingly.

## Prerequisites

- Rust 1.92.0+ with cargo
- Protobuf compiler for IBC message definitions
- Working directory: `packages/blockchain/`
- Oracle network (Phase 4) must define reporting format

## Tasks

### Task 1: CC Token Module

**Files:** `packages/blockchain/src/modules/token/`

Core token functionality:
- Mint: create new CC tokens (only callable by authorized minters - block rewards, governance)
- Burn: destroy tokens (fee burns, slashing)
- Transfer: move tokens between accounts with balance validation
- Balance query: get balance for any account
- Total supply tracking and query
- Token metadata: name="ComputeCoin", symbol="CC", decimals=18
- Events: Transfer { from, to, amount }, Mint { to, amount }, Burn { from, amount }
- Storage: account_balances (HashMap<Address, u128>), total_supply (u128)
- Interface trait: `TokenModule { mint, burn, transfer, balance_of, total_supply }`

### Task 2: Staking Module

**Files:** `packages/blockchain/src/modules/staking/`

Delegated Proof-of-Compute staking:
- Delegate: stake CC tokens to a compute node operator
- Undelegate: initiate unstaking (21-day unbonding period)
- Claim rewards: withdraw accumulated staking rewards
- Rewards distribution: proportional to delegated stake * node CU contribution
- Validator set management: top N staked operators are active validators
- Slashing: reduce stake on oracle-reported misbehavior
- Minimum stake: 1000 CC to delegate, 100,000 CC to operate validator
- Annual staking yield: 12% base + variable based on network utilization
- Storage: delegations, unbonding_queue, rewards_accumulator, validator_set

### Task 3: Governance Module

**Files:** `packages/blockchain/src/modules/governance/`

On-chain governance:
- Create proposal: requires 10,000 CC token deposit (returned if passes)
- Proposal types: parameter_change, software_upgrade, community_spend, text
- Voting: Yes/No/Abstain/NoWithVeto, weighted by staked tokens
- Voting period: 7 days
- Quorum: 33% of staked tokens must vote
- Pass threshold: >50% Yes votes (excluding Abstain)
- Veto threshold: >33% NoWithVeto triggers proposal failure and deposit burn
- Execution: approved proposals execute automatically after 2-day timelock
- Storage: proposals, votes, deposit_pool

### Task 4: Proof-of-Compute Consensus Integration

**Files:** `packages/blockchain/src/consensus/`

Custom consensus that rewards compute contribution:
- Block production: validators selected weighted by (stake * compute_contribution)
- Compute proof validation: verify oracle-signed CU reports included in block
- Block reward calculation: base_reward + (cu_contribution / total_cu) * variable_reward
- Epoch management: recalculate validator weights every 100 blocks
- Fork choice rule: heaviest chain (most accumulated compute proof weight)
- Finality: 2/3 validator signatures on block for finality

### Task 5: IBC-Compatible Message Passing

**Files:** `packages/blockchain/src/ibc/`

Inter-Blockchain Communication:
- IBC channel management: open, close, handshake
- Packet relay: send/receive messages to other IBC-enabled chains
- Token transfer: ICS-20 compatible cross-chain token transfers
- Compute attestation relay: share CU measurements with partner chains
- Timeout handling: refund on expired packets
- Acknowledgement processing: confirm receipt of cross-chain messages

## Verification

```bash
cd packages/blockchain
cargo fmt --check
cargo clippy -- -D warnings
cargo test
```

## Success Criteria

- Token transfers work correctly with proper balance updates
- Staking rewards calculate correctly based on delegation and CU contribution
- Governance proposals progress through full lifecycle (create, vote, execute)
- PoC consensus correctly weights validators by compute contribution
- IBC messages serialize/deserialize correctly
- All tests pass with no clippy warnings
- No unsafe code without explicit justification
