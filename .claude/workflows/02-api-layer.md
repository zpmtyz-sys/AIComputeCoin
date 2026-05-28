# Phase 2: API Layer - Gateway Service

## Overview

Build the API gateway service in Node.js/TypeScript using Fastify. This service is the public-facing interface that handles authentication, rate limiting, REST endpoints, and WebSocket connections for real-time data streaming.

## Prerequisites

- Node.js 22+ with pnpm
- Protobuf compiler for gRPC client generation
- Working directory: `packages/api-gateway/`
- Matching engine (Phase 1) must be complete for gRPC integration

## Tasks (Parallel)

### Task 1: Fastify Server Setup

**Files:** `packages/api-gateway/src/server.ts`, `packages/api-gateway/src/plugins/`

Set up the Fastify application with plugins:
- `@fastify/cors` - CORS configuration for web app origins
- `@fastify/jwt` - JWT token verification
- `@fastify/rate-limit` - per-IP and per-user rate limiting (100 req/s default, 10 req/s for order placement)
- `@fastify/websocket` - WebSocket upgrade handling
- `@fastify/swagger` - OpenAPI documentation generation
- Structured JSON logging with pino
- Graceful shutdown handler
- Health check endpoint at GET /health

### Task 2: Auth Module

**Files:** `packages/api-gateway/src/modules/auth/`

Authentication and authorization:
- POST /api/v1/auth/register - email/password registration with validation
- POST /api/v1/auth/login - returns access token (15min) + refresh token (7d)
- POST /api/v1/auth/refresh - exchange refresh token for new access token
- POST /api/v1/auth/api-keys - create API key with permissions (read, trade, withdraw)
- DELETE /api/v1/auth/api-keys/:id - revoke API key
- Password hashing with argon2
- JWT payload: { sub: userId, permissions: string[], iat, exp }
- API key authentication via X-API-Key header

### Task 3: Market Data Endpoints

**Files:** `packages/api-gateway/src/modules/market/`

Public market data (no auth required):
- GET /api/v1/market/:pair/orderbook - current order book depth (default 20 levels)
- GET /api/v1/market/:pair/trades - recent trades (default 100, max 1000)
- GET /api/v1/market/:pair/klines - OHLCV candlestick data (intervals: 1m,5m,15m,1h,4h,1d)
- GET /api/v1/market/:pair/ticker - 24h price statistics
- GET /api/v1/market/pairs - list all available trading pairs
- Response caching with configurable TTL
- Pagination with cursor-based approach

### Task 4: Trading Endpoints

**Files:** `packages/api-gateway/src/modules/trading/`

Authenticated trading operations:
- POST /api/v1/orders - create new order (requires trade permission)
- DELETE /api/v1/orders/:id - cancel order
- GET /api/v1/orders - list open orders with filtering
- GET /api/v1/orders/history - list completed/cancelled orders
- GET /api/v1/positions - list open positions
- GET /api/v1/account/balance - account balance by asset
- Input validation with zod schemas
- Idempotency key support for order creation

### Task 5: WebSocket Server

**Files:** `packages/api-gateway/src/modules/ws/`

Real-time data streaming:
- Connection: ws://host/ws/v1
- Public channels (no auth):
  - `orderbook@{pair}` - real-time order book updates (diff format)
  - `trades@{pair}` - real-time trade feed
  - `ticker@{pair}` - price ticker updates
  - `kline@{pair}@{interval}` - candlestick updates
- Private channels (auth required, send JWT in first message):
  - `orders` - user order status updates
  - `positions` - position change notifications
  - `balance` - balance updates
- Subscribe/unsubscribe message protocol
- Heartbeat ping/pong every 30 seconds
- Connection limit per user (max 5)

## Tasks (Sequential)

### Task 6: gRPC Integration with Matching Engine

**Files:** `packages/api-gateway/src/grpc/`

Connect to matching engine:
- Generate TypeScript client from `packages/shared/proto/matching.proto`
- Connection pool with retry logic (exponential backoff)
- Request timeout (5s default)
- Circuit breaker pattern for fault tolerance
- Transform gRPC responses to REST format

### Task 7: End-to-End API Tests

**Files:** `packages/api-gateway/tests/`

Integration test suite:
- Auth flow: register -> login -> access protected endpoint
- Order flow: login -> place order -> verify in open orders -> cancel
- Market data: request orderbook, trades, klines
- WebSocket: connect -> subscribe -> receive updates
- Rate limiting: verify 429 responses after limit exceeded
- Error handling: invalid inputs, expired tokens, insufficient permissions

## Verification

```bash
cd packages/api-gateway
pnpm install
pnpm lint
pnpm test
pnpm build
```

## Success Criteria

- All REST endpoints respond with correct status codes and data format
- JWT authentication works end-to-end
- WebSocket connections stream real-time data
- Rate limiting prevents abuse
- gRPC client connects to matching engine
- All tests pass with >80% coverage
- API documentation generated at /docs
