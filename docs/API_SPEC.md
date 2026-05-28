# ComputeCoin API Specification

## Overview

ComputeCoin provides three interface layers:

1. **REST API** - Standard HTTP endpoints for trading and account management
2. **WebSocket API** - Real-time streaming for market data and user events
3. **gRPC Internal** - High-performance inter-service communication

**Base URL:** `https://api.computecoin.io/api/v1/`
**WebSocket URL:** `wss://stream.computecoin.io/ws`

## Authentication

### JWT Bearer Token

Used for web and mobile clients:

```
Authorization: Bearer <jwt_token>
```

- Access token expiry: 15 minutes
- Refresh token expiry: 7 days
- Refresh token rotation on each use

### API Key + HMAC Signature

Used for programmatic access:

```
X-CC-API-KEY: <api_key>
X-CC-TIMESTAMP: <unix_ms>
X-CC-SIGNATURE: <hmac_sha256(timestamp + method + path + body, secret)>
```

## Rate Limiting

| Tier | Requests/sec | Burst | WebSocket Streams |
|------|-------------|-------|-------------------|
| Public (no auth) | 10/s | 20 | 1 connection |
| Authenticated | 50/s | 100 | 5 connections |
| Market Maker | 200/s | 500 | 20 connections |
| Institutional | 500/s | 1000 | 50 connections |

Rate limit headers returned on every response:
```
X-RateLimit-Limit: 50
X-RateLimit-Remaining: 47
X-RateLimit-Reset: 1700000000
```

## Error Format

All errors follow a consistent format:

```json
{
  "code": "ORDER_INSUFFICIENT_BALANCE",
  "message": "Insufficient balance to place order",
  "details": {
    "required": "1500.00",
    "available": "1200.50",
    "currency": "USDT"
  }
}
```

### Error Codes

| Code | HTTP Status | Description |
|------|------------|-------------|
| AUTH_INVALID_TOKEN | 401 | Token expired or malformed |
| AUTH_INSUFFICIENT_PERMISSIONS | 403 | Insufficient permissions |
| RATE_LIMIT_EXCEEDED | 429 | Too many requests |
| ORDER_INSUFFICIENT_BALANCE | 400 | Not enough funds |
| ORDER_INVALID_PRICE | 400 | Price outside allowed range |
| ORDER_NOT_FOUND | 404 | Order does not exist |
| MARKET_NOT_FOUND | 404 | Trading pair not available |
| MARKET_HALTED | 503 | Trading halted for this market |
| INTERNAL_ERROR | 500 | Unexpected server error |

---

## REST API Endpoints

### Authentication

#### POST /auth/register

Register a new user account.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecureP@ss123!",
  "referral_code": "ABC123"
}
```

**Response (201):**
```json
{
  "user_id": "usr_abc123def456",
  "email": "user@example.com",
  "created_at": "2024-01-15T10:30:00Z",
  "verification_sent": true
}
```

#### POST /auth/login

Authenticate and receive tokens.

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecureP@ss123!",
  "totp_code": "123456"
}
```

**Response (200):**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "refresh_token": "rt_abc123...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

#### POST /auth/refresh

Refresh an expired access token.

**Request:**
```json
{
  "refresh_token": "rt_abc123..."
}
```

**Response (200):**
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIs...",
  "refresh_token": "rt_def456...",
  "expires_in": 900,
  "token_type": "Bearer"
}
```

---

### Trading

#### POST /orders

Place a new order.

**Request:**
```json
{
  "pair": "CU-USDT",
  "side": "buy",
  "type": "limit",
  "price": "25.50",
  "quantity": "100.000",
  "time_in_force": "GTC",
  "client_order_id": "my-order-001"
}
```

**Response (201):**
```json
{
  "order_id": "ord_789xyz",
  "client_order_id": "my-order-001",
  "pair": "CU-USDT",
  "side": "buy",
  "type": "limit",
  "price": "25.50",
  "quantity": "100.000",
  "filled_quantity": "0.000",
  "status": "open",
  "time_in_force": "GTC",
  "created_at": "2024-01-15T10:30:00.123Z"
}
```

**Order Types:** `market`, `limit`, `stop_limit`, `stop_market`
**Time in Force:** `GTC` (Good Til Cancelled), `IOC` (Immediate or Cancel), `FOK` (Fill or Kill)

#### DELETE /orders/:id

Cancel an open order.

**Response (200):**
```json
{
  "order_id": "ord_789xyz",
  "status": "cancelled",
  "cancelled_at": "2024-01-15T10:31:00.456Z"
}
```

#### GET /orders

List orders with filtering.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| pair | string | all | Filter by trading pair |
| status | string | open | open, filled, cancelled, all |
| side | string | all | buy, sell |
| limit | int | 50 | Max results (1-500) |
| offset | int | 0 | Pagination offset |

**Response (200):**
```json
{
  "orders": [
    {
      "order_id": "ord_789xyz",
      "pair": "CU-USDT",
      "side": "buy",
      "type": "limit",
      "price": "25.50",
      "quantity": "100.000",
      "filled_quantity": "45.000",
      "status": "open",
      "created_at": "2024-01-15T10:30:00.123Z"
    }
  ],
  "total": 156,
  "limit": 50,
  "offset": 0
}
```

#### GET /orders/:id

Get a specific order by ID.

**Response (200):**
```json
{
  "order_id": "ord_789xyz",
  "client_order_id": "my-order-001",
  "pair": "CU-USDT",
  "side": "buy",
  "type": "limit",
  "price": "25.50",
  "quantity": "100.000",
  "filled_quantity": "100.000",
  "average_fill_price": "25.48",
  "status": "filled",
  "time_in_force": "GTC",
  "created_at": "2024-01-15T10:30:00.123Z",
  "updated_at": "2024-01-15T10:30:01.789Z",
  "fills": [
    {
      "fill_id": "fil_001",
      "price": "25.50",
      "quantity": "60.000",
      "fee": "0.0306",
      "fee_currency": "USDT",
      "timestamp": "2024-01-15T10:30:00.456Z"
    },
    {
      "fill_id": "fil_002",
      "price": "25.45",
      "quantity": "40.000",
      "fee": "0.02036",
      "fee_currency": "USDT",
      "timestamp": "2024-01-15T10:30:01.789Z"
    }
  ]
}
```

---

### Market Data

#### GET /markets

List all available trading pairs.

**Response (200):**
```json
{
  "markets": [
    {
      "pair": "CU-USDT",
      "base_asset": "CU",
      "quote_asset": "USDT",
      "status": "active",
      "min_order_size": "0.001",
      "max_order_size": "1000000.000",
      "price_precision": 2,
      "quantity_precision": 3,
      "maker_fee": "0.001",
      "taker_fee": "0.002",
      "last_price": "25.50",
      "volume_24h": "1523456.789",
      "change_24h": "3.25"
    }
  ]
}
```

#### GET /markets/:pair/orderbook

Get current order book for a trading pair.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| depth | int | 20 | Number of levels (1-100) |

**Response (200):**
```json
{
  "pair": "CU-USDT",
  "timestamp": 1705312200123,
  "bids": [
    ["25.48", "150.000"],
    ["25.47", "320.500"],
    ["25.46", "89.200"]
  ],
  "asks": [
    ["25.50", "200.000"],
    ["25.51", "175.300"],
    ["25.52", "410.000"]
  ]
}
```

#### GET /markets/:pair/trades

Get recent trades for a trading pair.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| limit | int | 50 | Max results (1-1000) |
| from_id | string | - | Fetch trades after this ID |

**Response (200):**
```json
{
  "trades": [
    {
      "trade_id": "trd_001",
      "price": "25.50",
      "quantity": "12.500",
      "side": "buy",
      "timestamp": 1705312200123
    },
    {
      "trade_id": "trd_002",
      "price": "25.49",
      "quantity": "8.300",
      "side": "sell",
      "timestamp": 1705312199456
    }
  ]
}
```

#### GET /markets/:pair/klines

Get OHLCV candlestick data.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| interval | string | 1h | 1m, 5m, 15m, 1h, 4h, 1d, 1w |
| start | int | - | Start time (unix ms) |
| end | int | - | End time (unix ms) |
| limit | int | 500 | Max results (1-1500) |

**Response (200):**
```json
{
  "pair": "CU-USDT",
  "interval": "1h",
  "klines": [
    {
      "open_time": 1705309200000,
      "open": "25.30",
      "high": "25.55",
      "low": "25.28",
      "close": "25.50",
      "volume": "45230.500",
      "close_time": 1705312799999,
      "trades": 1523
    }
  ]
}
```

---

### Account

#### GET /account/balance

Get account balances for all assets.

**Response (200):**
```json
{
  "balances": [
    {
      "asset": "USDT",
      "free": "15230.50",
      "locked": "2500.00",
      "total": "17730.50"
    },
    {
      "asset": "CU",
      "free": "450.000",
      "locked": "100.000",
      "total": "550.000"
    },
    {
      "asset": "CC",
      "free": "50000.000000",
      "locked": "10000.000000",
      "total": "60000.000000"
    }
  ]
}
```

#### GET /account/positions

Get open positions (futures/margin).

**Response (200):**
```json
{
  "positions": [
    {
      "position_id": "pos_001",
      "pair": "CU-USDT",
      "side": "long",
      "size": "500.000",
      "entry_price": "24.80",
      "mark_price": "25.50",
      "liquidation_price": "20.15",
      "leverage": 5,
      "unrealized_pnl": "350.00",
      "margin": "2480.00",
      "margin_ratio": "0.15",
      "opened_at": "2024-01-14T08:00:00Z"
    }
  ]
}
```

#### GET /account/history

Get account transaction history.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| type | string | all | trade, deposit, withdrawal, fee, funding, liquidation |
| asset | string | all | Filter by asset |
| start | int | - | Start time (unix ms) |
| end | int | - | End time (unix ms) |
| limit | int | 50 | Max results (1-1000) |

**Response (200):**
```json
{
  "history": [
    {
      "id": "txn_001",
      "type": "trade",
      "asset": "USDT",
      "amount": "-2550.00",
      "balance_after": "12680.50",
      "reference": "ord_789xyz",
      "timestamp": "2024-01-15T10:30:00.123Z"
    }
  ],
  "total": 2345,
  "limit": 50,
  "offset": 0
}
```

---

## WebSocket API

### Connection

```
wss://stream.computecoin.io/ws
```

### Authentication

Send auth message after connecting:
```json
{
  "method": "auth",
  "params": {
    "token": "eyJhbGciOiJSUzI1NiIs..."
  }
}
```

### Subscribe to Channels

```json
{
  "method": "subscribe",
  "params": {
    "channels": [
      "orderbook@CU-USDT",
      "trades@CU-USDT",
      "kline@CU-USDT@1m"
    ]
  }
}
```

### Channel Types

#### orderbook@{pair}

Real-time order book updates (diff stream):

```json
{
  "channel": "orderbook@CU-USDT",
  "data": {
    "timestamp": 1705312200123,
    "bids": [
      ["25.48", "150.000"],
      ["25.45", "0.000"]
    ],
    "asks": [
      ["25.52", "410.000"]
    ]
  }
}
```

A quantity of `"0.000"` means the price level has been removed.

#### trades@{pair}

Real-time trade stream:

```json
{
  "channel": "trades@CU-USDT",
  "data": {
    "trade_id": "trd_001",
    "price": "25.50",
    "quantity": "12.500",
    "side": "buy",
    "timestamp": 1705312200123
  }
}
```

#### kline@{pair}@{interval}

Candlestick updates (emitted every second during active trading):

```json
{
  "channel": "kline@CU-USDT@1m",
  "data": {
    "open_time": 1705312200000,
    "open": "25.48",
    "high": "25.52",
    "low": "25.47",
    "close": "25.50",
    "volume": "234.500",
    "closed": false
  }
}
```

#### user@{token}

Private user events (requires authentication):

```json
{
  "channel": "user",
  "data": {
    "event": "order_filled",
    "order_id": "ord_789xyz",
    "fill_price": "25.50",
    "fill_quantity": "60.000",
    "remaining": "40.000",
    "timestamp": 1705312200456
  }
}
```

User events include: `order_accepted`, `order_filled`, `order_cancelled`, `position_updated`, `balance_updated`, `liquidation_warning`

### Heartbeat

Server sends ping every 30 seconds. Client must respond with pong within 10 seconds or connection is terminated.

```json
{"method": "ping", "id": 1}
{"method": "pong", "id": 1}
```

---

## gRPC Internal Services

### MatchingService

```protobuf
service MatchingService {
  rpc PlaceOrder(PlaceOrderRequest) returns (PlaceOrderResponse);
  rpc CancelOrder(CancelOrderRequest) returns (CancelOrderResponse);
  rpc GetOrderBook(GetOrderBookRequest) returns (OrderBookSnapshot);
  rpc StreamMatches(StreamMatchesRequest) returns (stream MatchEvent);
}

message PlaceOrderRequest {
  string pair = 1;
  OrderSide side = 2;
  OrderType type = 3;
  string price = 4;
  string quantity = 5;
  string user_id = 6;
  string client_order_id = 7;
  TimeInForce time_in_force = 8;
}

message MatchEvent {
  string match_id = 1;
  string maker_order_id = 2;
  string taker_order_id = 3;
  string price = 4;
  string quantity = 5;
  int64 timestamp_ns = 6;
}
```

### SettlementService

```protobuf
service SettlementService {
  rpc SettleTrade(SettleTradeRequest) returns (SettleTradeResponse);
  rpc BatchSettle(BatchSettleRequest) returns (BatchSettleResponse);
  rpc GetSettlementStatus(GetSettlementStatusRequest) returns (SettlementStatus);
}

message SettleTradeRequest {
  string trade_id = 1;
  string buyer_id = 2;
  string seller_id = 3;
  string base_asset = 4;
  string quote_asset = 5;
  string quantity = 6;
  string price = 7;
  string maker_fee = 8;
  string taker_fee = 9;
}
```

### OracleService

```protobuf
service OracleService {
  rpc VerifyCompute(VerifyComputeRequest) returns (VerifyComputeResponse);
  rpc GetNodeStatus(GetNodeStatusRequest) returns (NodeStatus);
  rpc SubmitBenchmark(SubmitBenchmarkRequest) returns (BenchmarkResult);
  rpc StreamVerifications(StreamVerificationsRequest) returns (stream VerificationEvent);
}

message VerifyComputeRequest {
  string node_id = 1;
  string tee_attestation = 2;
  BenchmarkResult benchmark = 3;
  string stake_proof = 4;
}

message NodeStatus {
  string node_id = 1;
  double cu_rating = 2;
  double uptime = 3;
  int64 last_verified = 4;
  string stake_amount = 5;
  VerificationStatus status = 6;
}
```

---

## Versioning Strategy

- API versioned via URL prefix: `/api/v1/`, `/api/v2/`
- Breaking changes require new major version
- Deprecated endpoints receive `Sunset` header with removal date
- Minimum 6-month deprecation notice before removal
- WebSocket channels versioned separately: `v2/orderbook@CU-USDT`

## SDK Support

Official SDKs planned for:
- Python (`pip install computecoin`)
- JavaScript/TypeScript (`npm install @computecoin/sdk`)
- Go (`go get github.com/computecoin/sdk-go`)
- Rust (`computecoin-sdk` crate)
