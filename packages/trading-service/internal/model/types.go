package model

// Side represents the direction of an order or position.
type Side string

const (
	SideBuy  Side = "BUY"
	SideSell Side = "SELL"
)

// IsValid returns true if the side is a recognized value.
func (s Side) IsValid() bool {
	return s == SideBuy || s == SideSell
}

// OrderType represents the type of order.
type OrderType string

const (
	OrderTypeLimit     OrderType = "LIMIT"
	OrderTypeMarket    OrderType = "MARKET"
	OrderTypeStopLimit OrderType = "STOP_LIMIT"
	OrderTypeIOC       OrderType = "IOC"
	OrderTypeFOK       OrderType = "FOK"
)

// IsValid returns true if the order type is a recognized value.
func (ot OrderType) IsValid() bool {
	switch ot {
	case OrderTypeLimit, OrderTypeMarket, OrderTypeStopLimit, OrderTypeIOC, OrderTypeFOK:
		return true
	}
	return false
}

// OrderStatus represents the status of an order.
type OrderStatus string

const (
	OrderStatusNew             OrderStatus = "NEW"
	OrderStatusPartiallyFilled OrderStatus = "PARTIALLY_FILLED"
	OrderStatusFilled          OrderStatus = "FILLED"
	OrderStatusCancelled       OrderStatus = "CANCELLED"
	OrderStatusRejected        OrderStatus = "REJECTED"
)

// PositionStatus represents the status of a position.
type PositionStatus string

const (
	PositionStatusOpen       PositionStatus = "OPEN"
	PositionStatusClosed     PositionStatus = "CLOSED"
	PositionStatusLiquidated PositionStatus = "LIQUIDATED"
)

// MarginLevel represents the margin health level.
type MarginLevel string

const (
	MarginLevelHealthy     MarginLevel = "HEALTHY"
	MarginLevelWarning     MarginLevel = "WARNING"
	MarginLevelDanger      MarginLevel = "DANGER"
	MarginLevelLiquidation MarginLevel = "LIQUIDATION"
)

// TimeInForce represents the time-in-force policy for an order.
type TimeInForce string

const (
	TimeInForceGTC TimeInForce = "GTC" // Good til cancelled
	TimeInForceIOC TimeInForce = "IOC" // Immediate or cancel
	TimeInForceFOK TimeInForce = "FOK" // Fill or kill
)
