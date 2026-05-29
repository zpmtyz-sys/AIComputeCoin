# Contributing to ComputeCoin

Thank you for your interest in contributing to ComputeCoin! This document provides guidelines and instructions for contributing.

## Code of Conduct

This project follows the [Contributor Covenant v2.1](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you agree to uphold this code. Report violations to conduct@computecoin.io.

## Getting Started

### Prerequisites

- **Rust** 1.92+ (with `rustfmt` and `clippy`)
- **Go** 1.25+
- **Node.js** 22+ with **pnpm** 9+
- **Python** 3.9+ with **poetry**
- **Docker** and **Docker Compose**
- **Protocol Buffers** compiler (`protoc`)
- **Make**

### Development Setup

```bash
# Fork and clone the repository
git clone https://github.com/YOUR_USERNAME/AIComputeCoin.git
cd AIComputeCoin

# Install all dependencies
make setup

# Build all packages
make build

# Run all tests
make test

# Run linters
make lint

# Start local development services (databases, Kafka, etc.)
docker compose up -d

# Run the full stack locally
make dev
```

### Verifying Your Setup

```bash
# This should pass with no errors
make build && make test && make lint
```

## Branch Naming

Use descriptive branch names with a type prefix:

| Prefix | Use For |
|--------|---------|
| `feat/` | New features |
| `fix/` | Bug fixes |
| `docs/` | Documentation changes |
| `chore/` | Maintenance tasks |
| `refactor/` | Code refactoring |
| `perf/` | Performance improvements |
| `test/` | Adding or fixing tests |

Examples:
- `feat/limit-order-matching`
- `fix/websocket-reconnection`
- `docs/api-authentication-guide`
- `refactor/order-book-structure`

## Commit Messages

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `chore` | Maintenance, deps, CI |
| `refactor` | Code change without feature or fix |
| `perf` | Performance improvement |
| `test` | Adding or correcting tests |

### Scope (optional)

Use the package name: `matching-engine`, `api-gateway`, `trading-service`, `oracle-service`, `blockchain`, `web-app`, `shared`

### Examples

```
feat(matching-engine): implement price-time priority matching

Add limit order matching with price-time priority algorithm.
Uses lock-free order book for concurrent access.

Closes #123
```

```
fix(api-gateway): handle WebSocket reconnection on token refresh

Previously, WebSocket connections dropped when JWT expired.
Now the gateway gracefully re-authenticates active connections.
```

## Pull Request Process

### Before Opening a PR

1. Ensure your branch is up to date with `main`
2. All tests pass locally: `make test`
3. Linting passes: `make lint`
4. New code has tests with adequate coverage
5. Documentation is updated if behavior changes

### PR Requirements

- Fill out the PR template completely
- Title follows conventional commit format
- All CI checks pass (build, test, lint, security scan)
- Minimum 2 approving reviews from maintainers
- No unresolved review comments
- Squash merge into `main`

### Review Process

- Reviews are expected within 2 business days
- Address all review comments (resolve or discuss)
- Re-request review after making changes
- Maintainers may request changes or reject PRs

### CI Checks

Every PR runs:
- Build all packages
- Run all tests (unit + integration)
- Lint (per-language linters)
- Security scan (cargo audit, npm audit, govulncheck)
- Coverage check (80% minimum)
- Docker image build

## Code Style

### Rust (packages/matching-engine, packages/blockchain)

```bash
# Format
cargo fmt --all

# Lint (warnings are errors)
cargo clippy --all-targets --all-features -- -D warnings
```

Key conventions:
- Use `Result<T, E>` for fallible operations, never panic in library code
- Prefer iterators over loops
- Document public APIs with `///` doc comments
- Use `#[cfg(test)]` modules for unit tests

### Go (packages/trading-service, packages/oracle-service)

```bash
# Format
gofmt -w .

# Lint
golangci-lint run ./...
```

Key conventions:
- Use standard library patterns (interfaces, error wrapping)
- Table-driven tests
- Context propagation through call chains
- Structured logging with `slog`

### TypeScript/Node.js (packages/api-gateway, packages/web-app)

```bash
# Format and lint
pnpm lint
pnpm format
```

Key conventions:
- Strict TypeScript mode (no `any`)
- Prefer `const` over `let`, never `var`
- Use async/await over raw promises
- Functional style where appropriate

### Python (data services, scripts)

```bash
# Lint and format
ruff check .
ruff format .
```

Key conventions:
- Type hints on all function signatures
- Docstrings on public functions (Google style)
- Dataclasses or Pydantic for data models
- pytest for testing

## Testing Requirements

### Coverage

- **Minimum 80% line coverage** per package
- Critical paths (matching engine, settlement) target 95%+
- Coverage checked in CI, PRs that drop coverage are blocked

### Test Types

| Type | Location | Description |
|------|----------|-------------|
| Unit | Co-located with source | Test individual functions/methods |
| Integration | `tests/` per package | Test component interactions |
| E2E | `tests/e2e/` at root | Full system flows |
| Benchmark | `benches/` per package | Performance regression tests |

### Test-Driven Development

We encourage TDD:
1. Write a failing test that defines the desired behavior
2. Implement the minimum code to make the test pass
3. Refactor while keeping tests green

### Running Tests

```bash
# All tests
make test

# Specific package
make test-matching-engine
make test-api-gateway
make test-trading-service

# With coverage
make coverage
```

## Issue Workflow

### Creating Issues

- Use issue templates (bug report, feature request, improvement)
- Include reproduction steps for bugs
- Reference related issues or PRs

### Labels

| Label | Description |
|-------|-------------|
| `good first issue` | Suitable for new contributors |
| `help wanted` | Extra attention needed |
| `bug` | Something is broken |
| `enhancement` | New feature or improvement |
| `documentation` | Documentation updates |
| `priority: critical` | Must fix ASAP |
| `priority: high` | Fix this sprint |
| `priority: medium` | Fix this milestone |
| `priority: low` | Nice to have |

## Security Vulnerabilities

**Do NOT open public issues for security vulnerabilities.**

Please report security issues to: **security@computecoin.io**

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact assessment
- Suggested fix (if any)

We follow responsible disclosure:
- Acknowledge receipt within 24 hours
- Provide assessment within 72 hours
- Target fix within 7 days for critical issues
- Credit reporter in security advisory (unless anonymity requested)

## Contributor License Agreement (CLA)

Before your first contribution can be merged, you must sign the Contributor License Agreement. The CLA bot will comment on your PR with instructions.

The CLA ensures:
- You have the right to submit the contribution
- You grant the project a perpetual license to use the contribution
- You are not required to provide support for the contribution

## Architecture Decision Records (ADRs)

Significant architectural changes require an ADR:

1. Create `docs/adr/NNNN-title.md` using the template in `docs/adr/template.md`
2. Include: context, decision, consequences, alternatives considered
3. ADR must be approved before implementation begins
4. Reference the ADR in related PRs

When to write an ADR:
- New service or major component
- Technology choice (language, framework, database)
- Significant API change
- Security-sensitive decisions
- Changes affecting more than 3 packages

## Good First Issues

Look for issues labeled `good first issue` for a great starting point. These are typically:
- Well-scoped with clear acceptance criteria
- Touching a single package
- Having existing test patterns to follow
- Not requiring deep domain knowledge

## Getting Help

- **GitHub Discussions** - Questions, ideas, general discussion
- **Discord** - Real-time chat with contributors
- **Weekly Office Hours** - Video call with maintainers (Thursdays 10am UTC)

## Recognition

All contributors are recognized in:
- `CONTRIBUTORS.md` file
- Release notes for significant contributions
- Annual contributor spotlight blog posts
