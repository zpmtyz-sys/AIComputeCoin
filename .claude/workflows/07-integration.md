# Phase 7: Integration and Deployment

## Overview

Bring all services together into a cohesive platform. This phase focuses on Docker Compose orchestration, end-to-end testing across service boundaries, load testing, security hardening, monitoring, and deployment infrastructure.

## Prerequisites

- Docker and Docker Compose
- All services from Phases 1-6 implemented
- k6 for load testing
- Trivy for security scanning
- Helm for Kubernetes manifests

## Tasks

### Task 1: Docker Compose Full Stack Orchestration

**Files:** `docker-compose.yml`, `docker-compose.override.yml`, `packages/*/Dockerfile`

Local development environment:
- Service definitions for all packages:
  - matching-engine: Rust binary, exposes gRPC port 50051
  - api-gateway: Node.js, exposes HTTP 3000, WS 3001
  - trading-service: Go binary, exposes gRPC 50052
  - oracle-service: Go binary, exposes gRPC 50053
  - blockchain: Rust binary, exposes RPC 26657, P2P 26656
  - web-app: Next.js, exposes HTTP 8080
- Infrastructure services:
  - PostgreSQL 16 (trading-service, api-gateway)
  - Redis 7 (caching, rate limiting, sessions)
  - Kafka (Redpanda) for event streaming
  - Prometheus for metrics collection
  - Grafana for dashboards
- Network configuration: internal service mesh
- Volume mounts for persistent data
- Health checks on all services
- Environment variable management via .env files

### Task 2: End-to-End Test Suite

**Files:** `tests/e2e/`

Cross-service integration tests:
- Full trading flow: register -> deposit -> place order -> match -> check balance
- Order lifecycle: submit limit order -> market order crosses -> both filled -> positions updated
- WebSocket flow: connect -> subscribe orderbook -> place order -> receive update
- Settlement flow: open position -> wait for funding -> verify payment
- Oracle flow: submit compute proof -> oracle verifies -> CU reported on chain
- Token flow: stake CC -> earn rewards -> claim -> transfer
- Error scenarios: insufficient balance, invalid order, service timeout

### Task 3: Load Testing with k6

**Files:** `tests/load/`

Performance validation:
- Scenario 1: Order submission throughput (target: 10,000 orders/second)
- Scenario 2: Order book queries under load (target: 50,000 req/s)
- Scenario 3: WebSocket connections (target: 100,000 concurrent)
- Scenario 4: Mixed workload (realistic trading pattern)
- Metrics to capture: p50, p95, p99 latency, throughput, error rate
- Baseline comparison: track performance across releases
- Soak test: 1 hour sustained load at 80% capacity

### Task 4: Security Checklist

**Files:** `docs/SECURITY.md`, `scripts/security-scan.sh`

OWASP Top 10 and crypto-specific security:
- SQL injection: parameterized queries everywhere
- XSS: CSP headers, input sanitization, output encoding
- Authentication: secure password storage (argon2), JWT rotation
- Rate limiting: per-IP and per-user limits on all endpoints
- Input validation: strict schemas on all API inputs
- Dependency audit: cargo audit, npm audit, go vuln
- Container scanning: Trivy on all Docker images
- Secrets management: no hardcoded secrets, use env vars or vault
- API key security: hashed storage, prefix-based lookup
- WebSocket security: auth on connect, message size limits
- Private key management: encrypted at rest, HSM for production

### Task 5: Monitoring Setup

**Files:** `infrastructure/monitoring/`

Observability stack:
- Prometheus metrics:
  - matching_engine_orders_total (counter)
  - matching_engine_match_latency_us (histogram)
  - api_gateway_requests_total (counter by endpoint, status)
  - api_gateway_response_time_ms (histogram)
  - trading_service_positions_open (gauge)
  - oracle_service_cu_measurements_total (counter)
  - blockchain_block_height (gauge)
- Grafana dashboards:
  - Trading overview: orders/sec, matches/sec, latency
  - System health: CPU, memory, disk, network per service
  - Business metrics: volume, open interest, unique users
- Alerting rules:
  - Matching latency > 1ms for 5 minutes
  - API error rate > 1% for 2 minutes
  - Service down for > 30 seconds
  - Disk usage > 80%

### Task 6: Deployment Manifests

**Files:** `infrastructure/k8s/`, `infrastructure/helm/`

Kubernetes deployment:
- Helm chart per service with configurable values
- Horizontal Pod Autoscaler based on CPU/custom metrics
- Resource requests and limits per service
- ConfigMaps for non-sensitive configuration
- Secrets for sensitive data (DB passwords, API keys)
- Ingress configuration with TLS termination
- Network policies: restrict inter-service communication
- PodDisruptionBudgets for high availability
- Rolling update strategy with readiness probes

## Verification

```bash
docker compose build
docker compose up -d
sleep 30  # wait for services to be ready
cd tests/e2e && go test ./...
cd tests/load && k6 run mixed-workload.js
docker compose down
```

## Success Criteria

- All services start and communicate correctly via Docker Compose
- Full trading flow works end-to-end in containers
- Load test achieves 10K orders/second with p99 < 50ms
- No critical or high severity vulnerabilities in security scan
- Monitoring dashboards show all key metrics
- Helm charts deploy successfully to test cluster
- System recovers gracefully from single service failure
