# ComputeCoin -- Investor One-Pager

> **Contact**: 水镜先生 / Mr. Shuijing -- zpmtyz@gmail.com
>
> This document is for informational purposes only and does not constitute a securities offering,
> investment advice, or guarantee of returns. Potential investors should conduct their own due
> diligence and consult qualified legal counsel.

---

## The Problem

AI companies face an **unpredictable, illiquid compute market**:
- GPU capacity is scarce, fragmented across cloud providers, mining farms, and individuals.
- There is no standardized unit of measure, no transparent pricing, no hedging tools.
- Idle compute (nights, weekends, between training runs) is wasted value.

## The Solution: ComputeCoin

**Standardize delivered compute into a tradable token, then build a transparent exchange on top.**

```
Real compute delivered  -->  Measured in CU (1 TFLOPS * 1h FP16)
                        -->  Minted as CC via Proof-of-Delivered-Compute
                        -->  Traded on a transparent spot exchange
```

## Why It Works

| Pillar | Detail |
|--------|--------|
| **Intrinsic utility** | CC buys real compute; it is not an empty "consensus token" |
| **Transparent supply** | 2.1B cap, deterministic halving, public genesis allocation |
| **Compliance-first** | KYC/AML hooks, utility-token design, no hidden founder benefits |
| **Network effects** | More suppliers -> cheaper compute -> more demand -> more suppliers |
| **sub2api integration** | Leverages existing open-source AI API gateway (4000+ GitHub stars) |

## Business Model (Revenue to the Platform)

1. **Trading fees** (taker 0.10%, maker 0.02%) -- all to a public treasury
2. **Derivatives** (perpetuals, futures, options) -- backlog phases 2-3
3. **Market-making spread** -- transparent, public parameters
4. **Enterprise API access** -- institutional OTC desk, co-location
5. **Data products** -- compute pricing index, analytics

## Token Distribution (Transparent, Public)

| Bucket | % of Genesis | Lock |
|--------|-------------|------|
| Ecosystem & Liquidity | 30% | Unlocked (bootstrap liquidity) |
| Supplier Incentive Reserve | 25% | Linear 3y release |
| Team | 18% | 12-month cliff + 36-month linear |
| Investors (VC) | 15% | 6-month cliff + 24-month linear |
| Treasury (Governance) | 10% | Multi-sig governed |
| Community | 2% | Batch airdrops |

## Competitive Landscape

| Project | What they do | Why ComputeCoin is different |
|---------|-------------|------------------------------|
| Render / io.net / Akash | Compute marketplace | We add **financial infrastructure** (exchange, derivatives, standardized unit) |
| Bitcoin / Ethereum | Store of value / smart contracts | We have **intrinsic compute utility** backing the token |
| Binance / Coinbase | Crypto exchange | We are **vertically integrated** with compute supply |

## Roadmap

- **Phase 1 (0-6 mo)**: MVP live, seed users, 50+ compute nodes, daily volume $100K
- **Phase 2 (6-12 mo)**: Mainnet, perpetual contracts, 500+ nodes, $10M daily volume
- **Phase 3 (12-24 mo)**: Futures, options, ETF, enterprise OTC, regulatory licenses
- **Phase 4 (24-36 mo)**: Compute ABS, global pricing authority

## Ask

Seeking **seed / Series A** investment. Contact **zpmtyz@gmail.com**.

---

*This repository contains the working, tested codebase. Run `make up` and visit
`http://localhost:8788` to see it live.*
