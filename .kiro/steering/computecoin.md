# ComputeCoin Steering (for Kiro / AI Agents)

## Project Identity
- ComputeCoin (算力币): standardize delivered AI compute into CU -> transparent CC token -> spot exchange
- Compliant edition: no hidden mints, no secret founder backdoors, no unregistered securities
- Integrates with sub2api (HTTP adapter, process-isolated, MIT stays clean from LGPL)

## Critical Rules
1. Never use float for money/quantities. Always `Decimal`.
2. Every balance mutation goes through `ledger.service._apply` (double-entry).
3. Total CC minted must never exceed `tokenomics.MAX_SUPPLY`.
4. `source_attestation` is mandatory when registering capacity.
5. Cross-module calls only through service layers.
6. Run `make lint && make test` before any commit.

## Tech Stack
- Python 3.11, FastAPI, SQLAlchemy (async), Pydantic v2
- PostgreSQL (prod via docker-compose) / SQLite (dev/test)
- Redis (session, locks, hot data)
- Docker Compose for deployment (port 8788, not 8080)

## Port Convention
- 8080: reserved for sub2api (user's existing instance)
- 8788: ComputeCoin app (API + dashboard)
- 5433: PostgreSQL (host-mapped)
- 6380: Redis (host-mapped)

## Module Map
- `backend/app/modules/identity/` - auth, users, KYC hooks
- `backend/app/modules/relay/` - sub2api integration, capacity registration
- `backend/app/modules/compute/` - CU standardization, PoDC verification
- `backend/app/modules/ledger/` - double-entry ledger, genesis, emission
- `backend/app/modules/exchange/` - order book, matching engine
- `backend/app/modules/marketmaker/` - transparent quoting
- `backend/app/tokenomics.py` - ALL economic constants (single source of truth)

## When Modifying
- Parameters/constants: edit `tokenomics.py` or `core/config.py`, run tests
- Module logic: edit `modules/<x>/service.py`, keep router thin
- Shared zone (`core/`, `models/`, `tokenomics.py`): serialize changes, isolated PRs
