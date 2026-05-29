package validation

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/model"
)

func TestOrderValidator_ValidateOrder(t *testing.T) {
	tests := []struct {
		name       string
		order      *model.Order
		balance    *model.Balance
		lastPrice  decimal.Decimal
		openOrders int
		wantValid  bool
		wantErrMsg string
	}{
		{
			name: "valid limit order",
			order: &model.Order{
				UserID:    "user1",
				Pair:      "BTC-USDT",
				Side:      model.SideBuy,
				OrderType: model.OrderTypeLimit,
				Price:     decimal.RequireFromString("50000"),
				Quantity:  decimal.RequireFromString("0.1"),
			},
			balance:   model.NewBalance("user1", decimal.RequireFromString("10000")),
			wantValid: true,
		},
		{
			name: "valid market order",
			order: &model.Order{
				UserID:    "user1",
				Pair:      "ETH-USDT",
				Side:      model.SideSell,
				OrderType: model.OrderTypeMarket,
				Quantity:  decimal.RequireFromString("1"),
			},
			balance:   model.NewBalance("user1", decimal.RequireFromString("1000")),
			wantValid: true,
		},
		{
			name: "missing pair",
			order: &model.Order{
				UserID:    "user1",
				Pair:      "",
				Side:      model.SideBuy,
				OrderType: model.OrderTypeLimit,
				Price:     decimal.RequireFromString("50000"),
				Quantity:  decimal.RequireFromString("0.1"),
			},
			balance:    model.NewBalance("user1", decimal.RequireFromString("10000")),
			wantValid:  false,
			wantErrMsg: "pair is required",
		},
		{
			name: "zero quantity",
			order: &model.Order{
				UserID:    "user1",
				Pair:      "BTC-USDT",
				Side:      model.SideBuy,
				OrderType: model.OrderTypeLimit,
				Price:     decimal.RequireFromString("50000"),
				Quantity:  decimal.Zero,
			},
			balance:    model.NewBalance("user1", decimal.RequireFromString("10000")),
			wantValid:  false,
			wantErrMsg: "quantity must be positive",
		},
		{
			name: "negative price for limit order",
			order: &model.Order{
				UserID:    "user1",
				Pair:      "BTC-USDT",
				Side:      model.SideBuy,
				OrderType: model.OrderTypeLimit,
				Price:     decimal.RequireFromString("-100"),
				Quantity:  decimal.RequireFromString("0.1"),
			},
			balance:    model.NewBalance("user1", decimal.RequireFromString("10000")),
			wantValid:  false,
			wantErrMsg: "price must be positive for limit orders",
		},
		{
			name: "insufficient balance",
			order: &model.Order{
				UserID:    "user1",
				Pair:      "BTC-USDT",
				Side:      model.SideBuy,
				OrderType: model.OrderTypeLimit,
				Price:     decimal.RequireFromString("50000"),
				Quantity:  decimal.RequireFromString("1"),
			},
			balance:    model.NewBalance("user1", decimal.RequireFromString("100")),
			wantValid:  false,
			wantErrMsg: "insufficient balance",
		},
		{
			name: "position limit exceeded",
			order: &model.Order{
				UserID:    "user1",
				Pair:      "BTC-USDT",
				Side:      model.SideBuy,
				OrderType: model.OrderTypeLimit,
				Price:     decimal.RequireFromString("50000"),
				Quantity:  decimal.RequireFromString("0.1"),
			},
			balance:    model.NewBalance("user1", decimal.RequireFromString("100000")),
			openOrders: 100,
			wantValid:  false,
			wantErrMsg: "position limit exceeded",
		},
		{
			name: "price band violation",
			order: &model.Order{
				UserID:    "user1",
				Pair:      "BTC-USDT",
				Side:      model.SideBuy,
				OrderType: model.OrderTypeLimit,
				Price:     decimal.RequireFromString("60000"),
				Quantity:  decimal.RequireFromString("0.1"),
			},
			balance:    model.NewBalance("user1", decimal.RequireFromString("100000")),
			lastPrice:  decimal.RequireFromString("50000"),
			wantValid:  false,
			wantErrMsg: "price band violation",
		},
		{
			name: "invalid side",
			order: &model.Order{
				UserID:    "user1",
				Pair:      "BTC-USDT",
				Side:      model.Side("INVALID"),
				OrderType: model.OrderTypeLimit,
				Price:     decimal.RequireFromString("50000"),
				Quantity:  decimal.RequireFromString("0.1"),
			},
			balance:    model.NewBalance("user1", decimal.RequireFromString("10000")),
			wantValid:  false,
			wantErrMsg: "invalid side",
		},
		{
			name: "invalid order type",
			order: &model.Order{
				UserID:    "user1",
				Pair:      "BTC-USDT",
				Side:      model.SideBuy,
				OrderType: model.OrderType("INVALID"),
				Price:     decimal.RequireFromString("50000"),
				Quantity:  decimal.RequireFromString("0.1"),
			},
			balance:    model.NewBalance("user1", decimal.RequireFromString("10000")),
			wantValid:  false,
			wantErrMsg: "invalid order type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewOrderValidator()

			if tt.openOrders > 0 {
				validator.SetOpenOrderCount(tt.order.UserID, tt.openOrders)
			}
			if !tt.lastPrice.IsZero() {
				validator.SetLastPrice(tt.order.Pair, tt.lastPrice)
			}

			result := validator.ValidateOrder(tt.order, tt.balance)

			assert.Equal(t, tt.wantValid, result.Valid)
			if !tt.wantValid {
				assert.NotEmpty(t, result.Errors)
				found := false
				for _, e := range result.Errors {
					if contains(e, tt.wantErrMsg) {
						found = true
						break
					}
				}
				assert.True(t, found, "expected error containing %q, got %v", tt.wantErrMsg, result.Errors)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
