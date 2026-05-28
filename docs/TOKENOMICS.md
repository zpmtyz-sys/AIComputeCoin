# ComputeCoin (CC) Token Economics

## Token Overview

| Property | Value |
|----------|-------|
| Token Name | ComputeCoin |
| Symbol | CC |
| Total Supply | 2,100,000,000 (2.1 billion) |
| Decimals | 18 |
| Type | Deflationary (burn + halving) |
| Chain | ComputeCoin L1 (Cosmos SDK compatible) |
| Standard | Native + IBC (cross-chain wrapped) |

## Emission Schedule

CC follows a halving model inspired by Bitcoin but applied to compute:

| Epoch | Years | Block Reward | Annual Emission | Cumulative |
|-------|-------|-------------|-----------------|------------|
| 1 | 0-4 | 100 CC | ~262.8M CC | 1,051.2M |
| 2 | 4-8 | 50 CC | ~131.4M CC | 1,576.8M |
| 3 | 8-12 | 25 CC | ~65.7M CC | 1,839.6M |
| 4 | 12-16 | 12.5 CC | ~32.85M CC | 1,971.0M |
| 5 | 16-20 | 6.25 CC | ~16.4M CC | 2,036.6M |
| ... | ... | ... | ... | ... |

**Block time:** 6 seconds
**Blocks per year:** ~5,256,000
**Halving interval:** Every 4 years (~21,024,000 blocks)

## Token Distribution

| Allocation | Percentage | Amount | Vesting |
|-----------|-----------|--------|---------|
| Mining Rewards | 50% | 1,050M CC | Emitted per block via PoC |
| Team & Advisors | 15% | 315M CC | 4-year vest, 1-year cliff |
| Investors (Seed) | 5% | 105M CC | 2-year vest, 6-month cliff |
| Investors (Series A) | 7% | 147M CC | 2-year vest, 6-month cliff |
| Investors (Public) | 3% | 63M CC | 25% at TGE, 75% over 12 months |
| Treasury | 10% | 210M CC | Governed by DAO, 5-year unlock |
| Ecosystem Grants | 5% | 105M CC | Milestone-based disbursement |
| Insurance Fund | 5% | 105M CC | Locked, governance-controlled |

## Proof of Compute (PoC) Mining

### How Mining Works

Instead of solving hash puzzles (PoW) or locking capital (PoS), PoC miners earn CC by providing verified useful computation:

1. **Register**: Miner registers hardware with TEE attestation
2. **Verify**: Oracle network validates compute capacity (CU rating)
3. **Contribute**: Miner serves compute requests on the marketplace
4. **Earn**: Block rewards proportional to verified CU contribution

### Mining Reward Formula

```
Miner Reward = Block Reward x (Miner CU Contribution / Total Network CU) x Uptime Factor
```

Where:
- `Miner CU Contribution` = Verified CU-hours provided in the epoch
- `Total Network CU` = Sum of all verified CU-hours in the epoch
- `Uptime Factor` = 1.0 for 99.9%+, scales linearly down to 0.5 at 95%

### Mining Requirements

| Requirement | Minimum |
|-------------|---------|
| Compute | 10 CU capacity (10 TFLOPS FP16) |
| Stake | 10,000 CC |
| Uptime | 95% |
| Verification | Pass oracle check every epoch |
| Network | 1 Gbps symmetric |

## Staking Tiers

### Validator Staking

| Tier | Stake Required | Annual Yield | Benefits |
|------|---------------|--------------|----------|
| Bronze | 100,000 CC | 8% | Basic validation, governance voting |
| Silver | 500,000 CC | 12% | Priority order routing, reduced fees |
| Gold | 2,000,000 CC | 16% | Market maker benefits, proposal creation |
| Platinum | 10,000,000 CC | 20% | Governance council seat, treasury access |

### Delegator Staking

- Delegate to any validator
- Earn proportional rewards minus validator commission (5-20%)
- 21-day unbonding period
- Slashing risk shared with validator (capped at 5%)

### Staking Benefits

| Benefit | Bronze | Silver | Gold | Platinum |
|---------|--------|--------|------|----------|
| Fee Discount | 10% | 25% | 40% | 60% |
| Governance Weight | 1x | 1.5x | 2x | 3x |
| API Rate Limit Boost | 1.5x | 2x | 3x | 5x |
| Insurance Coverage | Base | +25% | +50% | +100% |
| Early Feature Access | No | Yes | Yes | Yes |

## Fee Structure

### Spot Trading

| Fee Type | Standard | VIP (>$1M/month) | Market Maker |
|----------|----------|-------------------|--------------|
| Maker | 0.10% | 0.05% | 0.00% |
| Taker | 0.20% | 0.10% | 0.05% |

### Futures Trading

| Fee Type | Standard | VIP (>$10M/month) | Market Maker |
|----------|----------|---------------------|--------------|
| Maker | 0.02% | 0.01% | -0.01% (rebate) |
| Taker | 0.05% | 0.03% | 0.02% |

### Other Fees

| Fee | Amount |
|-----|--------|
| Withdrawal (CC) | 1 CC |
| Withdrawal (CU settlement) | 0.1% |
| Liquidation penalty | 1.5% of position |
| Funding rate | Variable (capped at 0.1%/8hr) |
| Options exercise | 0.03% of notional |

## Treasury

### Allocation (10% of supply = 210M CC)

| Purpose | Percentage | Amount |
|---------|-----------|--------|
| Development Funding | 40% | 84M CC |
| Liquidity Incentives | 25% | 52.5M CC |
| Strategic Partnerships | 15% | 31.5M CC |
| Emergency Reserve | 10% | 21M CC |
| Community Rewards | 10% | 21M CC |

### Governance

- Treasury controlled by DAO governance
- Spending proposals require 10% quorum + 66% supermajority
- Maximum single disbursement: 5M CC without time-lock
- Disbursements >5M CC have 7-day time-lock for community review

## Governance Model

### Voting Power

```
Voting Power = CC Staked x Time Multiplier x Tier Multiplier
```

| Lock Duration | Time Multiplier |
|--------------|-----------------|
| No lock (liquid) | 1.0x |
| 3 months | 1.25x |
| 6 months | 1.5x |
| 12 months | 2.0x |
| 24 months | 3.0x |

### Quadratic Voting

For contentious proposals, quadratic voting is activated:
```
Effective Votes = sqrt(CC Committed to Vote)
```

This prevents plutocratic capture by large holders.

### Proposal Types

| Type | Quorum | Threshold | Time-lock |
|------|--------|-----------|-----------|
| Parameter Change | 5% | 50%+1 | 3 days |
| Treasury Spend | 10% | 66% | 7 days |
| Protocol Upgrade | 20% | 75% | 14 days |
| Emergency Action | 33% | 80% | None |

## Vesting Schedules

### Team & Advisors (15% - 315M CC)

```
Month 0-12:  Cliff (0% vested)
Month 13:    25% unlocked (78.75M CC)
Month 14-48: Linear monthly vesting (remaining 75%)
```

### Investors - Seed (5% - 105M CC)

```
Month 0-6:   Cliff (0% vested)
Month 7:     20% unlocked (21M CC)
Month 7-24:  Linear monthly vesting (remaining 80%)
```

### Investors - Series A (7% - 147M CC)

```
Month 0-6:   Cliff (0% vested)
Month 7:     15% unlocked (22.05M CC)
Month 7-24:  Linear monthly vesting (remaining 85%)
```

### Public Sale (3% - 63M CC)

```
TGE:         25% unlocked (15.75M CC)
Month 1-12:  Linear monthly vesting (remaining 75%)
```

## Burn Mechanism

### Fee Burns

- **20% of all trading fees** are burned quarterly
- Burn events are on-chain and publicly verifiable
- Estimated annual burn: 50-200M CC at scale (depends on volume)

### Deflationary Pressure

```
Annual Supply Change = Mining Emission - Fee Burns - Slashing Burns
```

At sufficient trading volume, CC becomes net-deflationary:
- Break-even: ~$50M daily trading volume
- Net-deflationary: >$100M daily trading volume

### Burn History (Projected)

| Year | Emission | Burns (Est.) | Net Change |
|------|----------|-------------|------------|
| 1 | 262.8M | 10M | +252.8M |
| 2 | 262.8M | 100M | +162.8M |
| 3 | 262.8M | 300M | -37.2M (deflationary) |
| 5 | 131.4M | 500M | -368.6M (deflationary) |

## Economic Security

### Attack Cost Analysis

| Attack Vector | Cost to Execute | Deterrent |
|--------------|-----------------|-----------|
| 51% Compute Attack | >$500M in hardware | Slashing + reputation loss |
| Oracle Manipulation | >$100M stake | 3-layer verification + slashing |
| Governance Capture | >$200M in CC | Quadratic voting + time-locks |
| Market Manipulation | >$50M liquidity | Circuit breakers + surveillance |

### Minimum Stake Requirements

| Role | Minimum Stake | Slashing Risk |
|------|--------------|---------------|
| Compute Miner | 10,000 CC | Up to 100% for fraud |
| Validator | 100,000 CC | Up to 10% for downtime |
| Oracle Node | 50,000 CC | Up to 50% for false reports |
| Market Maker | 500,000 CC | Up to 20% for manipulation |

### Insurance Fund

- 5% of total supply (105M CC) dedicated to insurance
- Covers: liquidation shortfalls, oracle failures, smart contract bugs
- Replenished by: 5% of liquidation penalties, 2% of trading fees
- Maximum single payout: 10% of fund balance
- Governed by insurance committee (elected by stakers)
