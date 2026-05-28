# ComputeCoin Iteration Plan

## Overview

Development follows a phased approach with clear success metrics and quality gates between phases. Each phase builds on the previous, with defined deliverables that can be independently validated.

## Sprint Methodology

- **Sprint duration:** 2 weeks
- **Ceremonies:** Planning (Monday), Daily standup, Review (Friday), Retro (Friday)
- **Estimation:** Story points (Fibonacci: 1, 2, 3, 5, 8, 13)
- **Velocity tracking:** After Phase 0, target sustainable velocity
- **Definition of Done:** Code reviewed, tests passing, documented, deployed to staging

---

## Phase 0: Foundation (Weeks 1-4)

### Objective
Build the core matching engine and basic trading API. Establish development infrastructure and prove the core technology works.

### Deliverables

| Sprint | Deliverable | Package |
|--------|-------------|---------|
| Sprint 1 | Order book data structure (red-black tree, price levels) | matching-engine |
| Sprint 1 | Basic order types (market, limit) | matching-engine |
| Sprint 1 | Monorepo setup, CI/CD pipeline | infrastructure |
| Sprint 2 | Price-time priority matching algorithm | matching-engine |
| Sprint 2 | REST API skeleton (place/cancel order, get orderbook) | api-gateway |
| Sprint 2 | Testnet CC token (basic ERC-20 equivalent) | blockchain |
| Sprint 2 | Docker Compose local development environment | infrastructure |

### Success Metrics

| Metric | Target |
|--------|--------|
| Matching throughput | 1,000 orders/sec |
| Matching latency (p99) | <100 microseconds |
| Test coverage | >80% |
| CI pipeline | Green on every merge |
| Order types supported | Market + Limit |

### Team Size: 5 Engineers
- 2 Rust (matching engine)
- 1 Go (service skeleton)
- 1 Node.js (API gateway)
- 1 DevOps (infrastructure)

### Quality Gate to Phase 1
- [ ] Matching engine passes 1000 orders/sec benchmark
- [ ] All unit tests passing (>80% coverage)
- [ ] CI/CD pipeline operational
- [ ] Local dev environment works with single command
- [ ] Architecture review completed

---

## Phase 1: MVP (Weeks 5-12)

### Objective
Launch invite-only spot trading with real users. Integrate initial GPU nodes and prove market demand.

### Deliverables

| Sprint | Deliverable | Package |
|--------|-------------|---------|
| Sprint 3 | WebSocket real-time market data | api-gateway |
| Sprint 3 | User authentication (JWT + API keys) | api-gateway |
| Sprint 3 | Account and balance management | trading-service |
| Sprint 4 | Order lifecycle (fills, partial fills, cancels) | trading-service |
| Sprint 4 | Basic risk checks (balance validation) | trading-service |
| Sprint 4 | Web trading UI (order entry, orderbook, chart) | web-app |
| Sprint 5 | Oracle v1 (basic compute verification) | oracle-service |
| Sprint 5 | 50 GPU node integrations | oracle-service |
| Sprint 5 | CU pricing (initial reference price from compute cost) | trading-service |
| Sprint 6 | Deposit/withdrawal (CC token) | blockchain |
| Sprint 6 | Trade history and account statements | trading-service |
| Sprint 6 | Invite system and onboarding flow | web-app |

### Success Metrics

| Metric | Target |
|--------|--------|
| Daily trading volume | $100K |
| Active users | 100 |
| Platform uptime | 99.9% |
| GPU nodes integrated | 50+ |
| Order-to-fill latency | <50ms (end-to-end) |
| Matching latency (p99) | <50 microseconds |

### Team Size: 10 Engineers
- 2 Rust (matching engine optimization)
- 2 Go (trading + oracle services)
- 2 Node.js (API gateway + WebSocket)
- 2 Frontend (web trading UI)
- 1 DevOps (deployment, monitoring)
- 1 QA (test automation)

### Quality Gate to Phase 2
- [ ] $100K daily volume sustained for 7 days
- [ ] 100 active users trading
- [ ] No critical bugs in production for 14 days
- [ ] 99.9% uptime over 30 days
- [ ] Security audit (Phase 1 scope) completed
- [ ] Oracle verifying 50+ nodes successfully

---

## Phase 2: Growth (Weeks 13-24)

### Objective
Launch futures trading, expand oracle network, launch mobile app, and open to public. Establish market maker partnerships.

### Deliverables

| Sprint | Deliverable | Package |
|--------|-------------|---------|
| Sprint 7 | Futures order types (perpetual, expiring) | matching-engine |
| Sprint 7 | Margin system (isolated and cross) | trading-service |
| Sprint 7 | Funding rate calculation and settlement | trading-service |
| Sprint 8 | Liquidation engine | trading-service |
| Sprint 8 | Insurance fund mechanics | trading-service |
| Sprint 8 | Advanced order types (stop-loss, take-profit, trailing stop) | matching-engine |
| Sprint 9 | Oracle v2 (TEE attestation, full 3-layer) | oracle-service |
| Sprint 9 | Independent oracle node operator program (10+ nodes) | oracle-service |
| Sprint 9 | Market maker API (co-location, FIX protocol) | api-gateway |
| Sprint 10 | Mobile app (iOS + Android, React Native) | mobile-app |
| Sprint 10 | Advanced charting (TradingView integration) | web-app |
| Sprint 10 | Public registration, KYC integration | api-gateway |
| Sprint 11 | Market maker onboarding (3+ firms) | partnerships |
| Sprint 11 | Referral and incentive programs | web-app |
| Sprint 11 | Multi-language support | web-app |
| Sprint 12 | Performance optimization (10K orders/sec) | matching-engine |
| Sprint 12 | Geographic expansion (EU, APAC nodes) | infrastructure |
| Sprint 12 | Circuit breakers and market surveillance | trading-service |

### Success Metrics

| Metric | Target |
|--------|--------|
| Daily trading volume | $10M |
| Active users | 5,000 |
| Market makers | 3+ |
| Oracle nodes (independent) | 10+ |
| Platform uptime | 99.95% |
| Matching throughput | 10,000 orders/sec |
| Futures open interest | $5M |
| Geographic regions | 3 (NA, EU, APAC) |

### Team Size: 20 Engineers
- 3 Rust (matching engine, blockchain)
- 4 Go (trading, oracle, settlement)
- 3 Node.js (API gateway, WebSocket scaling)
- 3 Frontend (web + mobile)
- 2 Quant (risk models, pricing)
- 2 DevOps/SRE
- 2 QA
- 1 Security

### Quality Gate to Phase 3
- [ ] $10M daily volume sustained for 14 days
- [ ] 3+ independent market makers active
- [ ] Oracle network: 10+ independent nodes with 99% agreement
- [ ] No liquidation failures (insurance fund intact)
- [ ] Futures funding rate stable and fair
- [ ] Mobile app: 4+ star rating
- [ ] Security audit (full platform) completed
- [ ] Regulatory counsel engaged in 2+ jurisdictions

---

## Phase 3: Scale (Weeks 25-48)

### Objective
Launch options and structured products, achieve cross-chain settlement, secure regulatory licenses, and onboard institutional clients.

### Deliverables

| Sprint | Deliverable | Package |
|--------|-------------|---------|
| Sprint 13-14 | Options engine (European, American) | matching-engine |
| Sprint 13-14 | Options pricing (Black-Scholes adapted for compute) | trading-service |
| Sprint 13-14 | Greeks calculation and display | web-app |
| Sprint 15-16 | Structured products (CU bonds, yield vaults) | trading-service |
| Sprint 15-16 | Cross-chain bridge (IBC to Cosmos ecosystem) | blockchain |
| Sprint 15-16 | Ethereum bridge (wrapped CC on EVM) | blockchain |
| Sprint 17-18 | Institutional API (FIX 4.4, dedicated endpoints) | api-gateway |
| Sprint 17-18 | Sub-account management | trading-service |
| Sprint 17-18 | Regulatory reporting and compliance tools | compliance |
| Sprint 19-20 | L2 Rollup deployment (optimistic) | blockchain |
| Sprint 19-20 | Compute index products (CU-50, CU-100 indices) | trading-service |
| Sprint 19-20 | Advanced analytics dashboard | web-app |
| Sprint 21-22 | Regulatory license applications (2 jurisdictions) | legal |
| Sprint 21-22 | Fiat on/off ramp integration | api-gateway |
| Sprint 21-22 | Prime brokerage features | trading-service |
| Sprint 23-24 | Solana bridge | blockchain |
| Sprint 23-24 | Portfolio margin (cross-asset) | trading-service |
| Sprint 23-24 | Social trading and copy-trading | web-app |

### Success Metrics

| Metric | Target |
|--------|--------|
| Daily trading volume | $100M |
| Active users | 50,000 |
| Institutional accounts | 50+ |
| Options open interest | $20M |
| Cross-chain volume | $10M/day |
| Regulatory licenses | 2 jurisdictions |
| Platform uptime | 99.99% |
| Matching throughput | 100,000 orders/sec |
| Oracle nodes | 50+ (global) |

### Team Size: 30+ Engineers
- 4 Rust (matching engine, blockchain, bridges)
- 5 Go (trading, oracle, settlement, compliance)
- 4 Node.js (API, integrations)
- 4 Frontend (web, mobile, analytics)
- 3 Quant (options pricing, risk, indices)
- 3 Blockchain (bridges, L2, consensus)
- 3 DevOps/SRE
- 2 QA
- 2 Security

---

## Risk Register

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|-----------|
| Matching engine latency regression | Medium | High | Continuous benchmarking in CI, performance budgets |
| Oracle manipulation | Low | Critical | 3-layer verification, economic penalties, redundant nodes |
| Regulatory action | Medium | High | Proactive engagement, multi-jurisdiction, compliance-first |
| Market maker withdrawal | Medium | High | 3+ MMs, internal market making capacity, incentives |
| Smart contract exploit | Low | Critical | Multiple audits, formal verification, insurance fund |
| Key person dependency | Medium | Medium | Documentation, knowledge sharing, redundant expertise |
| Scaling bottleneck | Medium | Medium | Load testing, horizontal scaling, L2 for overflow |
| Competition entry | Low | Medium | First-mover advantage, network effects, fast iteration |

## Team Scaling Plan

| Phase | Headcount | Key Additions |
|-------|-----------|---------------|
| Phase 0 | 5 | Core founding engineers |
| Phase 1 | 10 | Frontend, QA, additional backend |
| Phase 2 | 20 | Mobile, quant, security, DevOps |
| Phase 3 | 30+ | Institutional, compliance, blockchain |

### Hiring Priorities by Phase

**Phase 0:** Senior Rust engineers, Go backend, DevOps
**Phase 1:** Frontend specialists, QA automation, additional Go engineers
**Phase 2:** Quant researchers, mobile developers, security engineer, SRE
**Phase 3:** Blockchain specialists, compliance officers, institutional sales engineers

## Definition of Done (Global)

Every deliverable must meet:
1. Code reviewed and approved by 2+ engineers
2. Unit tests passing with >80% coverage
3. Integration tests passing
4. Performance benchmarks not regressed
5. Documentation updated
6. Deployed to staging and smoke-tested
7. Security review for sensitive changes
8. Product sign-off for user-facing changes
