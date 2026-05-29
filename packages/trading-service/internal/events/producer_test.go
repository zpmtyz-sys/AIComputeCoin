package events

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryProducer_Publish(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		payload   any
	}{
		{
			name:      "order submitted event",
			eventType: EventOrderSubmitted,
			payload: OrderSubmittedPayload{
				OrderID:       "order-1",
				UserID:        "user-1",
				Pair:          "BTC-USDT",
				Side:          "BUY",
				OrderType:     "LIMIT",
				Price:         decimal.RequireFromString("50000"),
				Quantity:      decimal.RequireFromString("1"),
				ClientOrderID: "client-1",
			},
		},
		{
			name:      "order cancelled event",
			eventType: EventOrderCancelled,
			payload: OrderCancelledPayload{
				OrderID: "order-1",
				UserID:  "user-1",
				Reason:  "user requested",
			},
		},
		{
			name:      "trade executed event",
			eventType: EventTradeExecuted,
			payload: TradeExecutedPayload{
				TradeID:      "trade-1",
				MakerOrderID: "order-1",
				TakerOrderID: "order-2",
				Pair:         "BTC-USDT",
				Price:        decimal.RequireFromString("50000"),
				Quantity:     decimal.RequireFromString("0.5"),
				Side:         "BUY",
			},
		},
		{
			name:      "position liquidated event",
			eventType: EventPositionLiquidated,
			payload: PositionLiquidatedPayload{
				PositionID:     "pos-1",
				UserID:         "user-1",
				LiquidatedSize: decimal.RequireFromString("2"),
				MarkPrice:      decimal.RequireFromString("45000"),
			},
		},
		{
			name:      "settlement completed event",
			eventType: EventSettlementCompleted,
			payload: SettlementCompletedPayload{
				ProcessedPositions: 10,
				TotalFunding:       decimal.RequireFromString("500"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer := NewInMemoryProducer()

			err := producer.Publish(tt.eventType, tt.payload)
			require.NoError(t, err)

			events := producer.GetEvents()
			require.Len(t, events, 1)

			event := events[0]
			assert.NotEmpty(t, event.ID)
			assert.Equal(t, tt.eventType, event.Type)
			assert.False(t, event.Timestamp.IsZero())
			assert.Equal(t, tt.payload, event.Payload)
		})
	}
}

func TestInMemoryProducer_EventOrdering(t *testing.T) {
	producer := NewInMemoryProducer()

	eventTypes := []EventType{
		EventOrderSubmitted,
		EventTradeExecuted,
		EventPositionOpened,
		EventSettlementCompleted,
	}

	for _, et := range eventTypes {
		err := producer.Publish(et, nil)
		require.NoError(t, err)
	}

	events := producer.GetEvents()
	require.Len(t, events, 4)

	for i, et := range eventTypes {
		assert.Equal(t, et, events[i].Type)
	}
}

func TestInMemoryProducer_GetEventsByType(t *testing.T) {
	producer := NewInMemoryProducer()

	_ = producer.Publish(EventOrderSubmitted, OrderSubmittedPayload{OrderID: "o1"})
	_ = producer.Publish(EventTradeExecuted, TradeExecutedPayload{TradeID: "t1"})
	_ = producer.Publish(EventOrderSubmitted, OrderSubmittedPayload{OrderID: "o2"})
	_ = producer.Publish(EventOrderCancelled, OrderCancelledPayload{OrderID: "o1"})

	submitted := producer.GetEventsByType(EventOrderSubmitted)
	assert.Len(t, submitted, 2)

	trades := producer.GetEventsByType(EventTradeExecuted)
	assert.Len(t, trades, 1)

	cancelled := producer.GetEventsByType(EventOrderCancelled)
	assert.Len(t, cancelled, 1)
}

func TestInMemoryProducer_Clear(t *testing.T) {
	producer := NewInMemoryProducer()

	_ = producer.Publish(EventOrderSubmitted, nil)
	_ = producer.Publish(EventTradeExecuted, nil)

	assert.Len(t, producer.GetEvents(), 2)

	producer.Clear()
	assert.Len(t, producer.GetEvents(), 0)
}

func TestInMemoryProducer_PayloadCorrectness(t *testing.T) {
	producer := NewInMemoryProducer()

	payload := OrderSubmittedPayload{
		OrderID:       "order-123",
		UserID:        "user-456",
		Pair:          "ETH-USDT",
		Side:          "SELL",
		OrderType:     "MARKET",
		Price:         decimal.RequireFromString("3000.50"),
		Quantity:      decimal.RequireFromString("2.5"),
		ClientOrderID: "client-789",
	}

	err := producer.Publish(EventOrderSubmitted, payload)
	require.NoError(t, err)

	events := producer.GetEvents()
	require.Len(t, events, 1)

	retrieved := events[0].Payload.(OrderSubmittedPayload)
	assert.Equal(t, "order-123", retrieved.OrderID)
	assert.Equal(t, "user-456", retrieved.UserID)
	assert.Equal(t, "ETH-USDT", retrieved.Pair)
	assert.Equal(t, "SELL", retrieved.Side)
	assert.Equal(t, "MARKET", retrieved.OrderType)
	assert.True(t, retrieved.Price.Equal(decimal.RequireFromString("3000.50")))
	assert.True(t, retrieved.Quantity.Equal(decimal.RequireFromString("2.5")))
	assert.Equal(t, "client-789", retrieved.ClientOrderID)
}
