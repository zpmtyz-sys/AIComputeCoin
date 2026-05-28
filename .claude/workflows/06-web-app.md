# Phase 6: Web Application - Trading Interface

## Overview

Build the web trading interface using Next.js 14 with App Router. This is the primary user-facing application featuring real-time charts, order book visualization, order entry, portfolio management, and wallet connection for the ComputeCoin trading platform.

## Prerequisites

- Node.js 22+ with pnpm
- Working directory: `packages/web-app/`
- API gateway (Phase 2) must be running for data
- Familiarity with TradingView lightweight-charts library

## Tasks (Parallel)

### Task 1: Layout and Navigation

**Files:** `packages/web-app/src/app/layout.tsx`, `packages/web-app/src/components/layout/`

Application shell:
- Root layout with dark theme (trading platforms are dark by default)
- Header: logo, market selector dropdown, account menu, notifications bell
- Sidebar: navigation links (Trade, Portfolio, Markets, Compute Nodes, Governance)
- Responsive design: desktop (full layout), tablet (collapsible sidebar), mobile (bottom nav)
- Theme: CSS variables for colors, consistent spacing scale
- Font: Inter for UI, JetBrains Mono for numbers/prices
- Loading states with skeleton screens

### Task 2: TradingView Chart Integration

**Files:** `packages/web-app/src/components/chart/`

Price chart with technical analysis:
- Use `lightweight-charts` library (not full TradingView widget)
- Candlestick chart as default view
- Time intervals: 1m, 5m, 15m, 1h, 4h, 1D, 1W
- Overlay indicators: MA(7), MA(25), MA(99), Bollinger Bands
- Volume histogram below price chart
- Crosshair with price/time tooltip
- Drawing tools: trend line, horizontal line, fibonacci retracement
- Fetch kline data from API: GET /api/v1/market/:pair/klines
- Real-time updates via WebSocket `kline@{pair}@{interval}` channel

### Task 3: Order Book Component

**Files:** `packages/web-app/src/components/orderbook/`

Real-time depth visualization:
- Dual-column layout: bids (green, left) and asks (red, right)
- Each row: price, quantity, cumulative total
- Background bar showing depth percentage
- Spread indicator between best bid and best ask
- Grouping selector: 0.01, 0.1, 1, 10 price increments
- Click on price to populate order form
- Depth chart visualization (area chart of cumulative bids/asks)
- WebSocket subscription: `orderbook@{pair}` for real-time diff updates
- Local orderbook state management with diff application

### Task 4: Order Entry Form

**Files:** `packages/web-app/src/components/order-form/`

Order placement interface:
- Tabs: Limit, Market, Stop-Limit
- Limit order: price input, quantity input, total display
- Market order: quantity input only, estimated fill price
- Stop-Limit: stop price, limit price, quantity
- Buy/Sell toggle with color indication (green buy, red sell)
- Slider for quick quantity selection (25%, 50%, 75%, 100% of balance)
- Available balance display
- Fee estimate display
- Confirmation modal for large orders (> 10% of balance)
- Form validation with clear error messages
- Submit via POST /api/v1/orders

### Task 5: Portfolio Dashboard

**Files:** `packages/web-app/src/components/portfolio/`

Account overview and positions:
- Total equity display with 24h change
- Asset breakdown table: token, available, in_orders, total, USD_value
- Open positions table: pair, side, size, entry_price, mark_price, unrealized_pnl, margin
- Open orders table: pair, side, type, price, quantity, filled%, time, cancel button
- Trade history: time, pair, side, price, quantity, fee, realized_pnl
- P&L chart: daily/weekly/monthly equity curve
- Filtering and sorting on all tables
- Export to CSV functionality

### Task 6: Authentication Flow

**Files:** `packages/web-app/src/components/auth/`, `packages/web-app/src/lib/auth.ts`

User authentication:
- Login page: email + password form
- Register page: email, password, confirm password, terms checkbox
- Wallet connect: MetaMask / WalletConnect integration for Web3 login
- Token management: store JWT in httpOnly cookie, auto-refresh before expiry
- Protected routes: redirect to login if unauthenticated
- Auth context provider: expose user state to all components
- Logout: clear tokens, redirect to home
- 2FA setup page (TOTP with QR code)

## Tasks (Sequential)

### Task 7: WebSocket Integration for Real-Time Updates

**Files:** `packages/web-app/src/lib/websocket.ts`, `packages/web-app/src/hooks/useWebSocket.ts`

Centralized WebSocket management:
- Single connection manager (reconnect on disconnect, exponential backoff)
- Subscription manager: subscribe/unsubscribe to channels
- Message routing: dispatch updates to appropriate components via context/store
- Connection status indicator in UI (green dot = connected, red = disconnected)
- Queue messages during reconnection, replay on restore
- React hook: `useWebSocket(channel)` returns latest data + connection status
- Automatic resubscription on reconnect

### Task 8: End-to-End Tests with Playwright

**Files:** `packages/web-app/tests/e2e/`

Browser automation tests:
- Login flow: enter credentials, verify redirect to trading page
- Place limit order: fill form, submit, verify appears in open orders
- Cancel order: click cancel, confirm, verify removal
- Chart interaction: change interval, verify candles update
- Order book: verify real-time updates, click price fills order form
- Responsive: test mobile layout, verify bottom navigation appears
- Error states: invalid login, network failure, insufficient balance

## Verification

```bash
cd packages/web-app
pnpm install
pnpm lint
pnpm test
pnpm build
```

## Success Criteria

- Trading interface renders correctly at all breakpoints
- Charts display real-time price data with technical indicators
- Order book updates in real-time with sub-second latency
- Orders can be placed and cancelled through the UI
- Portfolio shows accurate position and balance data
- Authentication flow works end-to-end including token refresh
- WebSocket reconnection is seamless to user
- All E2E tests pass
- Lighthouse score > 90 for performance
