package service

import (
	"context"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/config"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/events"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/model"
)

func newTestService() *TradingService {
	cfg := &config.Config{
		Port: "50052",
	}
	svc := NewTradingService(cfg)
	return svc
}

func TestTradingService_SubmitOrder_Success(t *testing.T) {
	tests := []struct {
		name    string
		order   *model.Order
		balance string
	}{
		{
			name: "valid limit buy order",
			order: &model.Order{
				UserID:    "user1",
				Pair:      "BTC-USDT",
				Side:      model.SideBuy,
				OrderType: model.OrderTypeLimit,
				Price:     decimal.RequireFromString("50000"),
				Quantity:  decimal.RequireFromString("0.1"),
			},
			balance: "10000",
		},
		{
			name: "valid limit sell order",
			order: &model.Order{
				UserID:    "user1",
				Pair:      "ETH-USDT",
				Side:      model.SideSell,
				OrderType: model.OrderTypeLimit,
				Price:     decimal.RequireFromString("3000"),
				Quantity:  decimal.RequireFromString("1"),
			},
			balance: "10000",
		},
		{
			name: "order with client order ID",
			order: &model.Order{
				UserID:        "user1",
				Pair:          "BTC-USDT",
				Side:          model.SideBuy,
				OrderType:     model.OrderTypeLimit,
				Price:         decimal.RequireFromString("50000"),
				Quantity:      decimal.RequireFromString("0.1"),
				ClientOrderID: "my-order-1",
			},
			balance: "10000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService()
			defer svc.Stop()
			svc.SetBalance(tt.order.UserID, decimal.RequireFromString(tt.balance))

			result, err := svc.SubmitOrder(context.Background(), tt.order)

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.NotEmpty(t, result.ID)
			assert.Equal(t, model.OrderStatusNew, result.Status)
			assert.False(t, result.CreatedAt.IsZero())

			// Verify event was published
			evts := svc.GetEventProducer().GetEvents()
			assert.NotEmpty(t, evts)
			assert.Equal(t, events.EventOrderSubmitted, evts[len(evts)-1].Type)
		})
	}
}

func TestTradingService_SubmitOrder_Rejection(t *testing.T) {
	tests := []struct {
		name    string
		order   *model.Order
		balance string
		wantErr string
	}{
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
			balance: "10000",
			wantErr: "pair is required",
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
			balance: "10000",
			wantErr: "quantity must be positive",
		},
		{
			name: "insufficient balance",
			order: &model.Order{
				UserID:    "user1",
				Pair:      "BTC-USDT",
				Side:      model.SideBuy,
				OrderType: model.OrderTypeLimit,
				Price:     decimal.RequireFromString("50000"),
				Quantity:  decimal.RequireFromString("10"),
			},
			balance: "100", // 50000 * 10 * 0.10 = 50000 required, only 100 available
			wantErr: "insufficient",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService()
			defer svc.Stop()
			svc.SetBalance(tt.order.UserID, decimal.RequireFromString(tt.balance))

			result, err := svc.SubmitOrder(context.Background(), tt.order)

			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
			assert.Equal(t, model.OrderStatusRejected, result.Status)
		})
	}
}

func TestTradingService_SubmitOrder_Idempotency(t *testing.T) {
	svc := newTestService()
	defer svc.Stop()
	svc.SetBalance("user1", decimal.RequireFromString("100000"))

	order := &model.Order{
		UserID:        "user1",
		Pair:          "BTC-USDT",
		Side:          model.SideBuy,
		OrderType:     model.OrderTypeLimit,
		Price:         decimal.RequireFromString("50000"),
		Quantity:      decimal.RequireFromString("0.1"),
		ClientOrderID: "idempotent-1",
	}

	// First submission
	result1, err := svc.SubmitOrder(context.Background(), order)
	require.NoError(t, err)

	// Second submission with same client order ID
	order2 := &model.Order{
		UserID:        "user1",
		Pair:          "BTC-USDT",
		Side:          model.SideBuy,
		OrderType:     model.OrderTypeLimit,
		Price:         decimal.RequireFromString("50000"),
		Quantity:      decimal.RequireFromString("0.1"),
		ClientOrderID: "idempotent-1",
	}
	result2, err := svc.SubmitOrder(context.Background(), order2)

	assert.ErrorIs(t, err, ErrDuplicateOrder)
	assert.Equal(t, result1.ID, result2.ID)
}

func TestTradingService_CancelOrder(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*TradingService) string
		userID  string
		wantErr bool
	}{
		{
			name: "cancel existing order",
			setup: func(svc *TradingService) string {
				svc.SetBalance("user1", decimal.RequireFromString("100000"))
				order := &model.Order{
					UserID:    "user1",
					Pair:      "BTC-USDT",
					Side:      model.SideBuy,
					OrderType: model.OrderTypeLimit,
					Price:     decimal.RequireFromString("50000"),
					Quantity:  decimal.RequireFromString("0.1"),
				}
				result, _ := svc.SubmitOrder(context.Background(), order)
				return result.ID
			},
			userID:  "user1",
			wantErr: false,
		},
		{
			name: "cancel nonexistent order",
			setup: func(svc *TradingService) string {
				return "nonexistent-id"
			},
			userID:  "user1",
			wantErr: true,
		},
		{
			name: "cancel order from different user",
			setup: func(svc *TradingService) string {
				svc.SetBalance("user1", decimal.RequireFromString("100000"))
				order := &model.Order{
					UserID:    "user1",
					Pair:      "BTC-USDT",
					Side:      model.SideBuy,
					OrderType: model.OrderTypeLimit,
					Price:     decimal.RequireFromString("50000"),
					Quantity:  decimal.RequireFromString("0.1"),
				}
				result, _ := svc.SubmitOrder(context.Background(), order)
				return result.ID
			},
			userID:  "user2",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newTestService()
			defer svc.Stop()
			orderID := tt.setup(svc)

			err := svc.CancelOrder(context.Background(), orderID, tt.userID)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify order is cancelled
				order, _ := svc.GetOrder(context.Background(), orderID)
				assert.Equal(t, model.OrderStatusCancelled, order.Status)

				// Verify cancel event was published
				evts := svc.GetEventProducer().GetEvents()
				lastEvt := evts[len(evts)-1]
				assert.Equal(t, events.EventOrderCancelled, lastEvt.Type)
			}
		})
	}
}

func TestTradingService_GetBalance(t *testing.T) {
	svc := newTestService()
	defer svc.Stop()
	svc.SetBalance("user1", decimal.RequireFromString("5000"))

	balance := svc.GetBalance(context.Background(), "user1")
	assert.True(t, balance.GetAvailable().Equal(decimal.RequireFromString("5000")))
	assert.True(t, balance.GetLocked().IsZero())

	// Submit order to lock margin
	order := &model.Order{
		UserID:    "user1",
		Pair:      "BTC-USDT",
		Side:      model.SideBuy,
		OrderType: model.OrderTypeLimit,
		Price:     decimal.RequireFromString("50000"),
		Quantity:  decimal.RequireFromString("0.1"),
	}
	_, err := svc.SubmitOrder(context.Background(), order)
	require.NoError(t, err)

	// Balance should now have locked portion (50000 * 0.1 * 0.10 = 500)
	balance = svc.GetBalance(context.Background(), "user1")
	assert.True(t, balance.GetAvailable().Equal(decimal.RequireFromString("4500")))
	assert.True(t, balance.GetLocked().Equal(decimal.RequireFromString("500")))
}

func TestTradingService_GetOpenOrders(t *testing.T) {
	svc := newTestService()
	defer svc.Stop()
	svc.SetBalance("user1", decimal.RequireFromString("100000"))

	// Submit multiple orders
	for i := 0; i < 3; i++ {
		order := &model.Order{
			UserID:    "user1",
			Pair:      "BTC-USDT",
			Side:      model.SideBuy,
			OrderType: model.OrderTypeLimit,
			Price:     decimal.RequireFromString("50000"),
			Quantity:  decimal.RequireFromString("0.01"),
		}
		_, err := svc.SubmitOrder(context.Background(), order)
		require.NoError(t, err)
	}

	orders := svc.GetOpenOrders(context.Background(), "user1")
	assert.Len(t, orders, 3)

	// Cancel one
	err := svc.CancelOrder(context.Background(), orders[0].ID, "user1")
	require.NoError(t, err)

	orders = svc.GetOpenOrders(context.Background(), "user1")
	assert.Len(t, orders, 2)
}

func TestTradingService_PositionLifecycle(t *testing.T) {
	svc := newTestService()
	defer svc.Stop()
	svc.SetBalance("user1", decimal.RequireFromString("100000"))

	// Open a position via the tracker
	tracker := svc.GetPositionTracker()
	pos := tracker.OpenPosition(
		"user1",
		"BTC-USDT",
		model.SideBuy,
		decimal.RequireFromString("1"),
		decimal.RequireFromString("50000"),
		decimal.RequireFromString("5000"),
	)

	// Verify position exists
	positions := svc.GetPositions(context.Background(), "user1")
	require.Len(t, positions, 1)
	assert.Equal(t, pos.ID, positions[0].ID)

	// Update mark price
	tracker.UpdateMarkPrice("BTC-USDT", decimal.RequireFromString("52000"))

	// Verify PnL
	positions = svc.GetPositions(context.Background(), "user1")
	require.Len(t, positions, 1)
	assert.True(t, positions[0].UnrealizedPnL.Equal(decimal.RequireFromString("2000")))

	// Close position
	pnl, err := tracker.ClosePosition("user1", pos.ID,
		decimal.RequireFromString("53000"), decimal.RequireFromString("1"))
	require.NoError(t, err)
	assert.True(t, pnl.Equal(decimal.RequireFromString("3000")))

	// No more open positions
	positions = svc.GetPositions(context.Background(), "user1")
	assert.Len(t, positions, 0)
}

func TestTradingService_GetBalance_NonexistentUser(t *testing.T) {
	svc := newTestService()
	defer svc.Stop()

	balance := svc.GetBalance(context.Background(), "unknown-user")
	assert.True(t, balance.GetAvailable().IsZero())
}
