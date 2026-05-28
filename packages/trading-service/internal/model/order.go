package model

import "time"

// Order represents a trading order.
type Order struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	Pair           string    `json:"pair"`
	Side           string    `json:"side"`
	OrderType      string    `json:"order_type"`
	Price          string    `json:"price"`
	Quantity       string    `json:"quantity"`
	FilledQuantity string    `json:"filled_quantity"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Position represents a user's open position.
type Position struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Pair      string    `json:"pair"`
	Side      string    `json:"side"`
	Quantity  string    `json:"quantity"`
	EntryPrice string   `json:"entry_price"`
	CreatedAt time.Time `json:"created_at"`
}

// Trade represents an executed trade.
type Trade struct {
	ID           string    `json:"id"`
	MakerOrderID string    `json:"maker_order_id"`
	TakerOrderID string    `json:"taker_order_id"`
	Pair         string    `json:"pair"`
	Price        string    `json:"price"`
	Quantity     string    `json:"quantity"`
	Side         string    `json:"side"`
	TradedAt     time.Time `json:"traded_at"`
}
