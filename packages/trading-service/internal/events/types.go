package events

import (
	"time"

	"github.com/shopspring/decimal"
)

// EventType represents the type of domain event.
type EventType string

const (
	EventOrderSubmitted      EventType = "ORDER_SUBMITTED"
	EventOrderCancelled      EventType = "ORDER_CANCELLED"
	EventOrderRejected       EventType = "ORDER_REJECTED"
	EventTradeExecuted       EventType = "TRADE_EXECUTED"
	EventPositionOpened      EventType = "POSITION_OPENED"
	EventPositionClosed      EventType = "POSITION_CLOSED"
	EventPositionLiquidated  EventType = "POSITION_LIQUIDATED"
	EventSettlementCompleted EventType = "SETTLEMENT_COMPLETED"
	EventFundingApplied      EventType = "FUNDING_APPLIED"
)

// Event represents a domain event.
type Event struct {
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload"`
}

// OrderSubmittedPayload is the payload for an ORDER_SUBMITTED event.
type OrderSubmittedPayload struct {
	OrderID       string          `json:"order_id"`
	UserID        string          `json:"user_id"`
	Pair          string          `json:"pair"`
	Side          string          `json:"side"`
	OrderType     string          `json:"order_type"`
	Price         decimal.Decimal `json:"price"`
	Quantity      decimal.Decimal `json:"quantity"`
	ClientOrderID string          `json:"client_order_id"`
}

// OrderCancelledPayload is the payload for an ORDER_CANCELLED event.
type OrderCancelledPayload struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
	Reason  string `json:"reason"`
}

// OrderRejectedPayload is the payload for an ORDER_REJECTED event.
type OrderRejectedPayload struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
	Reason  string `json:"reason"`
}

// TradeExecutedPayload is the payload for a TRADE_EXECUTED event.
type TradeExecutedPayload struct {
	TradeID      string          `json:"trade_id"`
	MakerOrderID string          `json:"maker_order_id"`
	TakerOrderID string          `json:"taker_order_id"`
	Pair         string          `json:"pair"`
	Price        decimal.Decimal `json:"price"`
	Quantity     decimal.Decimal `json:"quantity"`
	Side         string          `json:"side"`
}

// PositionOpenedPayload is the payload for a POSITION_OPENED event.
type PositionOpenedPayload struct {
	PositionID string          `json:"position_id"`
	UserID     string          `json:"user_id"`
	Pair       string          `json:"pair"`
	Side       string          `json:"side"`
	Size       decimal.Decimal `json:"size"`
	EntryPrice decimal.Decimal `json:"entry_price"`
	Margin     decimal.Decimal `json:"margin"`
}

// PositionClosedPayload is the payload for a POSITION_CLOSED event.
type PositionClosedPayload struct {
	PositionID  string          `json:"position_id"`
	UserID      string          `json:"user_id"`
	ClosePrice  decimal.Decimal `json:"close_price"`
	RealizedPnL decimal.Decimal `json:"realized_pnl"`
}

// PositionLiquidatedPayload is the payload for a POSITION_LIQUIDATED event.
type PositionLiquidatedPayload struct {
	PositionID     string          `json:"position_id"`
	UserID         string          `json:"user_id"`
	LiquidatedSize decimal.Decimal `json:"liquidated_size"`
	MarkPrice      decimal.Decimal `json:"mark_price"`
}

// SettlementCompletedPayload is the payload for a SETTLEMENT_COMPLETED event.
type SettlementCompletedPayload struct {
	ProcessedPositions int             `json:"processed_positions"`
	TotalFunding       decimal.Decimal `json:"total_funding"`
}

// FundingAppliedPayload is the payload for a FUNDING_APPLIED event.
type FundingAppliedPayload struct {
	Pair        string          `json:"pair"`
	FundingRate decimal.Decimal `json:"funding_rate"`
}
