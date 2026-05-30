# AGENTS.md

This repository uses a single source of truth for agent behavior: **[`CLAUDE.md`](CLAUDE.md)**.

Any coding agent (Claude Code, Hermes, Cursor, generic `/workflow` sub-agents, etc.) working in this
repo MUST read and follow `CLAUDE.md` first. It defines:

- The repository map (`backend/app/` modular monolith).
- The **non-negotiable invariants** (ledger conservation, supply cap, deterministic halving,
  transparent genesis, matching fairness, Decimal-only money, source attestation, no backdoors).
- Coding conventions and module-boundary rules that keep the codebase stable while many agents work
  in parallel.
- The compliance red lines.

## Quick pointers

| You want to... | Read |
|----------------|------|
| Understand the whole system | [`docs/architecture.md`](docs/architecture.md) |
| Change token/emission parameters | [`backend/app/tokenomics.py`](backend/app/tokenomics.py) + [`docs/tokenomics.md`](docs/tokenomics.md) |
| Understand a module's job & interface | [`docs/modules.md`](docs/modules.md) |
| Know how we (don't) touch sub2api | [`docs/sub2api-integration.md`](docs/sub2api-integration.md) + [`docs/compliance.md`](docs/compliance.md) |
| Pick up parallelizable work | [`docs/workflow.md`](docs/workflow.md) |

## Hermes / multi-agent note

- Treat `backend/app/core/`, `backend/app/models/`, and `backend/app/tokenomics.py` as a **serialized
  hot zone**: only one open change at a time, isolated PR. Everything under `backend/app/modules/<x>/`
  can be developed in parallel per-module.
- Definition of Done for every task: `make lint && make test` is green and no invariant in `CLAUDE.md`
  is violated.
