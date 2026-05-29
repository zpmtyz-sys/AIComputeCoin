# Phase 1: Foundation - Matching Engine

## Overview

Build the core matching engine in Rust. This is the heart of the trading platform - a high-performance order book with price-time priority matching. The engine must handle limit, market, stop-limit, IOC, and FOK order types with sub-microsecond matching latency targets.

## Prerequisites

- Rust toolchain (1.92.0+) with cargo
- Protobuf compiler (protoc) for gRPC interface definitions
- Working directory: `packages/matching-engine/`

## Tasks (Parallel)

These tasks can be developed concurrently by multiple agents:

### Task 1: Order Book Data Structure

**File:** `packages/matching-engine/src/orderbook.rs`

Implement the core order book with:
- BTreeMap-based price levels (sorted bids descending, asks ascending)
- Price-time priority within each level (VecDeque of orders)
- O(log n) insertion and removal
- Snapshot generation for market data feeds
- Methods: `add_order`, `cancel_order`, `get_best_bid`, `get_best_ask`, `get_depth(levels)`

### Task 2: Order Types

**File:** `packages/matching-engine/src/types.rs`

Implement order type definitions and validation:
- Limit order: price + quantity, rests on book if not filled
- Market order: executes at best available price, no resting
- Stop-Limit order: activates as limit when stop price is triggered
- IOC (Immediate or Cancel): fill what you can, cancel remainder
- FOK (Fill or Kill): fill entirely or reject completely
- Order struct with: id, side (Buy/Sell), order_type, price, quantity, filled_quantity, timestamp, status

### Task 3: Matching Algorithm

**File:** `packages/matching-engine/src/matcher.rs`

Implement the matching engine logic:
- Price-time priority: best price first, then earliest order at same price
- Partial fills: generate two trade events (maker fill + taker fill)
- Trade event generation with: trade_id, maker_order_id, taker_order_id, price, quantity, timestamp
- Handle crossing orders (new buy >= best ask, or new sell <= best bid)
- Self-trade prevention (same user on both sides)

### Task 4: gRPC Server Interface

**Files:** `packages/matching-engine/src/server.rs`, `packages/shared/proto/matching.proto`

Implement tonic-based gRPC server:
- Service methods: SubmitOrder, CancelOrder, GetOrderBook, GetTrades
- Streaming: SubscribeOrderBook, SubscribeTrades
- Request/response types defined in protobuf
- Connection pooling and graceful shutdown
- Health check endpoint

## Tasks (Sequential)

These must be done after parallel tasks are complete:

### Task 5: Integration Tests

**File:** `packages/matching-engine/tests/integration_test.rs`

Full order lifecycle tests:
- Place limit buy, place limit sell at crossing price, verify trade
- Partial fill scenario (large order matched by multiple smaller orders)
- Cancel order and verify removal from book
- IOC order that partially fills
- FOK order that cannot fully fill (should reject)
- Stop-limit activation when market moves
- Order book state consistency after many operations

### Task 6: Benchmarks

**File:** `packages/matching-engine/benches/matching_bench.rs`

Performance benchmarks targeting <10us per match:
- Single order insertion benchmark
- Order matching (crossing) benchmark
- Bulk order processing (1000 orders)
- Order cancellation benchmark
- Order book depth query benchmark

## Verification

```bash
cd packages/matching-engine
cargo fmt --check
cargo clippy -- -D warnings
cargo test
cargo bench
```

## Success Criteria

- All unit and integration tests pass
- No clippy warnings
- Code is formatted with rustfmt
- Benchmark shows <100us per match operation (stretch goal: <10us)
- gRPC server starts and accepts connections
- Order book maintains consistency under concurrent access
