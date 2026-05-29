package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

// Order represents a trading order.
type Order struct {
	ID              string          `json:"id"`
	UserID          string          `json:"user_id"`
	Pair            string          `json:"pair"`
	Side            Side            `json:"side"`
	OrderType       OrderType       `json:"order_type"`
	Price           decimal.Decimal `json:"price"`
	StopPrice       decimal.Decimal `json:"stop_price"`
	Quantity        decimal.Decimal `json:"quantity"`
	FilledQuantity  decimal.Decimal `json:"filled_quantity"`
	Status          OrderStatus     `json:"status"`
	TimeInForce     TimeInForce     `json:"time_in_force"`
	ClientOrderID   string          `json:"client_order_id"`
	RejectionReason string          `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

// ValidateOrder validates the order fields and returns an error if invalid.
func (o *Order) ValidateOrder() error {
	var errs []string

	// Validate pair format (e.g., "BTC-USDT")
	if o.Pair == "" {
		errs = append(errs, "pair is required")
	} else if !isValidPair(o.Pair) {
		errs = append(errs, fmt.Sprintf("invalid pair format: %s (expected BASE-QUOTE)", o.Pair))
	}

	// Validate quantity
	if o.Quantity.IsZero() || o.Quantity.IsNegative() {
		errs = append(errs, "quantity must be positive")
	}

	// Validate price for limit orders
	if o.OrderType == OrderTypeLimit || o.OrderType == OrderTypeStopLimit {
		if o.Price.IsZero() || o.Price.IsNegative() {
			errs = append(errs, "price must be positive for limit orders")
		}
	}

	// Validate stop price for stop-limit orders
	if o.OrderType == OrderTypeStopLimit {
		if o.StopPrice.IsZero() || o.StopPrice.IsNegative() {
			errs = append(errs, "stop price must be positive for stop-limit orders")
		}
	}

	// Validate side
	if !o.Side.IsValid() {
		errs = append(errs, fmt.Sprintf("invalid side: %s", o.Side))
	}

	// Validate order type
	if !o.OrderType.IsValid() {
		errs = append(errs, fmt.Sprintf("invalid order type: %s", o.OrderType))
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// isValidPair checks if the pair follows the BASE-QUOTE format.
func isValidPair(pair string) bool {
	parts := strings.Split(pair, "-")
	if len(parts) != 2 {
		return false
	}
	return len(parts[0]) > 0 && len(parts[1]) > 0
}

// Position represents a user's open position.
type Position struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Pair       string    `json:"pair"`
	Side       string    `json:"side"`
	Quantity   string    `json:"quantity"`
	EntryPrice string    `json:"entry_price"`
	CreatedAt  time.Time `json:"created_at"`
}

// Trade represents an executed trade.
type Trade struct {
	ID           string          `json:"id"`
	MakerOrderID string          `json:"maker_order_id"`
	TakerOrderID string          `json:"taker_order_id"`
	Pair         string          `json:"pair"`
	Price        decimal.Decimal `json:"price"`
	Quantity     decimal.Decimal `json:"quantity"`
	Side         Side            `json:"side"`
	IsMaker      bool            `json:"is_maker"`
	TradedAt     time.Time       `json:"traded_at"`
}
