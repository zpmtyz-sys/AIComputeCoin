package validation

import (
	"fmt"
	"sync"

	"github.com/shopspring/decimal"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/model"
)

// ValidationResult holds the result of order validation.
type ValidationResult struct {
	Valid  bool
	Errors []string
}

// OrderValidator validates orders against business rules.
type OrderValidator struct {
	mu             sync.RWMutex
	openOrderCount map[string]int             // userID -> count of open orders
	lastPrices     map[string]decimal.Decimal // pair -> last traded price
	maxOpenOrders  int
	maxPriceBand   decimal.Decimal // max deviation from last price (e.g., 0.10 = 10%)
}

// NewOrderValidator creates a new OrderValidator with default limits.
func NewOrderValidator() *OrderValidator {
	return &OrderValidator{
		openOrderCount: make(map[string]int),
		lastPrices:     make(map[string]decimal.Decimal),
		maxOpenOrders:  100,
		maxPriceBand:   decimal.NewFromFloat(0.10),
	}
}

// ValidateOrder validates an order against business rules including balance sufficiency.
func (v *OrderValidator) ValidateOrder(order *model.Order, balance *model.Balance) ValidationResult {
	result := ValidationResult{Valid: true}

	// Basic order validation
	if err := order.ValidateOrder(); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, err.Error())
		return result
	}

	// Check balance sufficiency
	requiredMargin := v.calculateRequiredMargin(order)
	if balance != nil {
		available := balance.GetAvailable()
		if available.LessThan(requiredMargin) {
			result.Valid = false
			result.Errors = append(result.Errors, fmt.Sprintf(
				"insufficient balance: required %s, available %s",
				requiredMargin.String(), available.String(),
			))
		}
	}

	// Check position limits
	v.mu.RLock()
	openCount := v.openOrderCount[order.UserID]
	v.mu.RUnlock()
	if openCount >= v.maxOpenOrders {
		result.Valid = false
		result.Errors = append(result.Errors, fmt.Sprintf(
			"position limit exceeded: max %d open orders per user", v.maxOpenOrders,
		))
	}

	// Check price bands for limit orders
	if order.OrderType == model.OrderTypeLimit || order.OrderType == model.OrderTypeStopLimit {
		v.mu.RLock()
		lastPrice, exists := v.lastPrices[order.Pair]
		v.mu.RUnlock()
		if exists && !lastPrice.IsZero() {
			deviation := order.Price.Sub(lastPrice).Abs().Div(lastPrice)
			if deviation.GreaterThan(v.maxPriceBand) {
				result.Valid = false
				result.Errors = append(result.Errors, fmt.Sprintf(
					"price band violation: price deviates %.1f%% from last price (max %.1f%%)",
					deviation.Mul(decimal.NewFromInt(100)).InexactFloat64(),
					v.maxPriceBand.Mul(decimal.NewFromInt(100)).InexactFloat64(),
				))
			}
		}
	}

	if len(result.Errors) > 0 {
		result.Valid = false
	}
	return result
}

// IncrementOpenOrders increments the open order count for a user.
func (v *OrderValidator) IncrementOpenOrders(userID string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.openOrderCount[userID]++
}

// DecrementOpenOrders decrements the open order count for a user.
func (v *OrderValidator) DecrementOpenOrders(userID string) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.openOrderCount[userID] > 0 {
		v.openOrderCount[userID]--
	}
}

// SetLastPrice updates the last traded price for a pair.
func (v *OrderValidator) SetLastPrice(pair string, price decimal.Decimal) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.lastPrices[pair] = price
}

// GetOpenOrderCount returns the open order count for a user.
func (v *OrderValidator) GetOpenOrderCount(userID string) int {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.openOrderCount[userID]
}

// SetOpenOrderCount sets the open order count directly (for testing).
func (v *OrderValidator) SetOpenOrderCount(userID string, count int) {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.openOrderCount[userID] = count
}

// calculateRequiredMargin calculates the initial margin required for the order.
// Initial margin is 10% of notional value.
func (v *OrderValidator) calculateRequiredMargin(order *model.Order) decimal.Decimal {
	initialMarginRate := decimal.NewFromFloat(0.10)
	var notional decimal.Decimal
	if order.OrderType == model.OrderTypeMarket {
		// For market orders, use quantity as a rough estimate (price unknown)
		// In production, use a reference price
		notional = order.Quantity
	} else {
		notional = order.Price.Mul(order.Quantity)
	}
	return notional.Mul(initialMarginRate)
}
