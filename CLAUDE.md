# CLAUDE.md - ComputeCoin Master Configuration

## Superpowers Methodology

This project uses the following superpowers skills for development:

- **brainstorming** - Collaborative ideation for architecture decisions and feature design
- **writing-plans** - Structured planning documents before implementation
- **subagent-driven-development** - Delegating complex tasks to specialized sub-agents
- **test-driven-development** - Write tests first, then implement to pass them
- **systematic-debugging** - Methodical approach to diagnosing and fixing issues

Reference `.claude/workflows/` for phase-specific orchestration and workflow definitions.

## Project Context

**ComputeCoin (算力币)** is the world's first compute power financial exchange - a global platform for compute power financialization and trading.

### Core Concept

**CU (Compute Unit) Standardization:**

```
1 CU = 1 TFLOPS x 1 hour FP16 computation
```

ComputeCoin combines:
- **CU Standardization** - Universal unit for measuring compute power
- **Crypto Liquidity** - Blockchain-native settlement and token economics
- **Commodity Derivatives** - Full derivatives stack (futures, options, structured products)

## Monorepo Structure

```
packages/
  matching-engine/    # Rust - Lock-free order book, <10us latency target
  api-gateway/        # Node.js/Fastify - REST/WebSocket API, auth, rate limiting
  trading-service/    # Go - Order lifecycle, position management, risk engine
  oracle-service/     # Go - 3-layer compute verification (TEE, benchmarks, staking)
  blockchain/         # Rust - Custom chain (Cosmos SDK compatible), CC token, PoC
  web-app/            # Next.js - Trading UI, dashboards, portfolio management
  shared/             # Protobuf definitions, shared types, generated code
```

## Build Commands

```bash
make build    # Build all packages
make test     # Run all tests
make lint     # Run all linters
make proto    # Regenerate protobuf types
make setup    # Install all dependencies
```

## Coding Standards

### Commits
- **Conventional commits**: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `perf:`, `test:`
- Scope optional: `feat(matching-engine): add limit order support`

### Development Principles
- **TDD** - Test-driven development: write tests first, implement to pass
- **YAGNI** - You Aren't Gonna Need It: don't build what isn't required now
- **DRY** - Don't Repeat Yourself: extract shared logic into packages/shared

### Per-Language Standards
- **Rust**: `cargo fmt`, `cargo clippy` (deny warnings), edition 2021
- **Go**: `golangci-lint`, `gofmt`, Go 1.25+
- **TypeScript**: ESLint + Prettier, strict mode, no any
- **Python**: Ruff, type hints required, Python 3.9+

### Testing
- Minimum 80% code coverage per package
- Unit tests co-located with source
- Integration tests in `tests/` directory per package
- E2E tests in `tests/e2e/` at repo root

## Key Architectural Decisions

1. **Rust for matching engine** - Zero-cost abstractions, guaranteed memory safety, predictable latency
2. **Lock-free data structures** - Avoid mutex contention in hot path
3. **Event sourcing** - All state changes are immutable events in Kafka
4. **CQRS** - Command/Query Responsibility Segregation for read/write optimization
5. **Protobuf for IPC** - Type-safe, language-agnostic inter-service communication
6. **Kubernetes-native** - All services containerized, horizontally scalable
7. **Multi-region** - Active-active deployment for low latency globally
8. **Proof of Compute (PoC)** - Novel consensus mechanism rewarding real computation

## Workflow Orchestration

See `.claude/workflows/` for phase-specific development workflows:
- `phase-0-foundation.yml` - Core engine and basic API
- `phase-1-mvp.yml` - Spot trading and user-facing features
- `phase-2-growth.yml` - Futures, oracle network, mobile
- `phase-3-scale.yml` - Options, cross-chain, institutional
