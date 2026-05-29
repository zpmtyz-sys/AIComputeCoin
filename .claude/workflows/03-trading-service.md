# Phase 3: Trading Service - Order Lifecycle Management

## Overview

Build the trading service in Go. This service manages the complete order lifecycle from validation through settlement, handles position tracking, P&L calculation, margin requirements, and event sourcing for audit trail and downstream consumers.

## Prerequisites

- Go 1.25.1+
- Kafka (or Redpanda) for event streaming
- PostgreSQL for persistent state
- Working directory: `packages/trading-service/`
- Matching engine (Phase 1) and API gateway (Phase 2) should be operational

## Tasks

### Task 1: Order Validation and Risk Checks

**Files:** `packages/trading-service/internal/validation/`

Pre-trade risk management:
- Validate order parameters (price > 0, quantity within limits, valid pair)
- Check account balance sufficiency (available balance >= order cost + fees)
- Position limit checks (max position size per pair, max open orders)
- Price band validation (reject orders > 10% from last trade price)
- Rate limiting per user (max 10 orders/second)
- Duplicate order detection (idempotency key check)
- Interface: `ValidateOrder(ctx, order) error`

### Task 2: Position Tracking and P&L Calculation

**Files:** `packages/trading-service/internal/position/`

Real-time position management:
- Position struct: pair, side, entry_price (weighted average), quantity, unrealized_pnl, realized_pnl
- Update position on trade fill (adjust entry price, quantity)
- Mark-to-market unrealized P&L calculation using index price
- Realized P&L on position close (FIFO accounting)
- Position history with open/close timestamps
- Interface: `UpdatePosition(ctx, trade) (*Position, error)`

### Task 3: Settlement Engine

**Files:** `packages/trading-service/internal/settlement/`

Trade settlement and funding:
- Mark-to-market settlement every 8 hours (00:00, 08:00, 16:00 UTC)
- Funding rate calculation: premium_rate = (mark_price - index_price) / index_price
- Funding payment: position_size * funding_rate (longs pay shorts when positive)
- Fee calculation: maker fee 0.02%, taker fee 0.05%
- Balance updates: debit/credit accounts atomically
- Settlement event emission for downstream accounting
- Interface: `Settle(ctx, trades []Trade) error`

### Task 4: Margin System

**Files:** `packages/trading-service/internal/margin/`

Margin and liquidation logic:
- Initial margin requirement: 10% of position notional (10x max leverage)
- Maintenance margin: 5% of position notional
- Margin ratio calculation: (equity / position_notional) * 100
- Liquidation trigger: margin_ratio < maintenance_margin
- Liquidation process: market close at bankruptcy price, insurance fund covers shortfall
- ADL (Auto-Deleveraging): if insurance fund insufficient, deleverage profitable traders
- Real-time margin monitoring with alert thresholds (warning at 7.5%)
- Interface: `CheckMargin(ctx, userId) (*MarginStatus, error)`

### Task 5: Event Sourcing with Kafka

**Files:** `packages/trading-service/internal/events/`

Event-driven architecture:
- Kafka producer for domain events:
  - `order.submitted` - new order received
  - `order.filled` - order matched (partial or full)
  - `order.cancelled` - order cancelled by user or system
  - `trade.executed` - trade completed
  - `position.opened` - new position created
  - `position.closed` - position fully closed
  - `position.liquidated` - forced liquidation
  - `settlement.completed` - funding/settlement cycle done
- Event schema with: event_id, event_type, timestamp, aggregate_id, payload
- Guaranteed ordering per aggregate (partition by user_id)
- At-least-once delivery with idempotent consumers
- Dead letter queue for failed processing

## Verification

```bash
cd packages/trading-service
go fmt ./...
go vet ./...
golangci-lint run
go test ./... -cover
```

## Success Criteria

- Full order lifecycle from submission to settlement works end-to-end
- Position tracking accurately calculates P&L
- Margin system correctly identifies liquidation candidates
- Funding rate payments are calculated and applied correctly
- All events are published to Kafka with correct schemas
- Test coverage > 80%
- No race conditions (tests pass with -race flag)
