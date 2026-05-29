package margin

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/model"
)

func TestMarginSystem_CalculateInitialMargin(t *testing.T) {
	tests := []struct {
		name       string
		orderType  model.OrderType
		price      string
		quantity   string
		wantMargin string
	}{
		{
			name:       "limit order BTC",
			orderType:  model.OrderTypeLimit,
			price:      "50000",
			quantity:   "1",
			wantMargin: "5000", // 50000 * 1 * 0.10
		},
		{
			name:       "limit order small quantity",
			orderType:  model.OrderTypeLimit,
			price:      "3000",
			quantity:   "0.5",
			wantMargin: "150", // 3000 * 0.5 * 0.10
		},
		{
			name:       "market order uses quantity",
			orderType:  model.OrderTypeMarket,
			price:      "0",
			quantity:   "1000",
			wantMargin: "100", // 1000 * 0.10 (uses quantity for market)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewMarginSystem()
			order := &model.Order{
				OrderType: tt.orderType,
				Price:     decimal.RequireFromString(tt.price),
				Quantity:  decimal.RequireFromString(tt.quantity),
			}

			margin := ms.CalculateInitialMargin(order)
			assert.True(t, margin.Equal(decimal.RequireFromString(tt.wantMargin)),
				"expected margin %s, got %s", tt.wantMargin, margin.String())
		})
	}
}

func TestMarginSystem_CalculateMaintenanceMargin(t *testing.T) {
	tests := []struct {
		name       string
		markPrice  string
		size       string
		wantMargin string
	}{
		{
			name:       "standard position",
			markPrice:  "50000",
			size:       "1",
			wantMargin: "2500", // 50000 * 1 * 0.05
		},
		{
			name:       "small position",
			markPrice:  "100",
			size:       "10",
			wantMargin: "50", // 100 * 10 * 0.05
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewMarginSystem()
			pos := &model.TradingPosition{
				MarkPrice: decimal.RequireFromString(tt.markPrice),
				Size:      decimal.RequireFromString(tt.size),
			}

			margin := ms.CalculateMaintenanceMargin(pos)
			assert.True(t, margin.Equal(decimal.RequireFromString(tt.wantMargin)),
				"expected margin %s, got %s", tt.wantMargin, margin.String())
		})
	}
}

func TestMarginSystem_CheckLiquidation(t *testing.T) {
	tests := []struct {
		name           string
		side           model.Side
		entryPrice     string
		size           string
		margin         string
		markPrice      string
		wantLiquidated bool
	}{
		{
			name:           "healthy long position",
			side:           model.SideBuy,
			entryPrice:     "50000",
			size:           "1",
			margin:         "5000",
			markPrice:      "50000",
			wantLiquidated: false,
		},
		{
			name:           "long position near liquidation",
			side:           model.SideBuy,
			entryPrice:     "50000",
			size:           "1",
			margin:         "5000",  // 10% initial margin
			markPrice:      "45500", // loss = 4500, effective margin = 500, maintenance = 45500*0.05 = 2275
			wantLiquidated: true,
		},
		{
			name:           "short position near liquidation",
			side:           model.SideSell,
			entryPrice:     "50000",
			size:           "1",
			margin:         "5000",
			markPrice:      "54500", // loss = 4500, effective margin = 500, maintenance = 54500*0.05 = 2725
			wantLiquidated: true,
		},
		{
			name:           "healthy short position",
			side:           model.SideSell,
			entryPrice:     "50000",
			size:           "1",
			margin:         "5000",
			markPrice:      "49000",
			wantLiquidated: false,
		},
		{
			name:           "closed position not liquidatable",
			side:           model.SideBuy,
			entryPrice:     "50000",
			size:           "1",
			margin:         "100",
			markPrice:      "10000",
			wantLiquidated: false, // because we'll set status to closed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := NewMarginSystem()
			pos := &model.TradingPosition{
				Side:       tt.side,
				EntryPrice: decimal.RequireFromString(tt.entryPrice),
				Size:       decimal.RequireFromString(tt.size),
				Margin:     decimal.RequireFromString(tt.margin),
				Status:     model.PositionStatusOpen,
			}

			if tt.name == "closed position not liquidatable" {
				pos.Status = model.PositionStatusClosed
			}

			result := ms.CheckLiquidation(pos, decimal.RequireFromString(tt.markPrice))
			assert.Equal(t, tt.wantLiquidated, result)
		})
	}
}

func TestMarginSystem_TriggerLiquidation(t *testing.T) {
	ms := NewMarginSystem()

	pos := &model.TradingPosition{
		ID:            "pos1",
		UserID:        "user1",
		Side:          model.SideBuy,
		EntryPrice:    decimal.RequireFromString("50000"),
		Size:          decimal.RequireFromString("1"),
		Margin:        decimal.RequireFromString("5000"),
		MarkPrice:     decimal.RequireFromString("44000"),
		UnrealizedPnL: decimal.RequireFromString("-6000"), // loss exceeds margin
		Status:        model.PositionStatusOpen,
	}

	result := ms.TriggerLiquidation(pos)

	assert.Equal(t, "pos1", result.PositionID)
	assert.Equal(t, "user1", result.UserID)
	assert.True(t, result.LiquidatedSize.Equal(decimal.RequireFromString("1")))
	assert.Equal(t, model.PositionStatusLiquidated, pos.Status)
	assert.True(t, pos.Size.IsZero())

	// Since margin (5000) + PnL (-6000) = -1000, insurance should have paid
	assert.True(t, result.InsurancePaid.IsZero()) // insurance fund was 0, so nothing to pay
}

func TestMarginSystem_InsuranceFund(t *testing.T) {
	ms := NewMarginSystem()
	ms.AddToInsuranceFund(decimal.RequireFromString("10000"))

	assert.True(t, ms.GetInsuranceFund().Equal(decimal.RequireFromString("10000")))

	// Liquidate a position with deficit
	pos := &model.TradingPosition{
		ID:            "pos1",
		UserID:        "user1",
		Side:          model.SideBuy,
		EntryPrice:    decimal.RequireFromString("50000"),
		Size:          decimal.RequireFromString("1"),
		Margin:        decimal.RequireFromString("5000"),
		MarkPrice:     decimal.RequireFromString("44000"),
		UnrealizedPnL: decimal.RequireFromString("-6000"),
		Status:        model.PositionStatusOpen,
	}

	result := ms.TriggerLiquidation(pos)

	// Deficit = 5000 - 6000 = -1000, insurance pays 1000
	assert.True(t, result.InsurancePaid.Equal(decimal.RequireFromString("1000")))
	assert.True(t, ms.GetInsuranceFund().Equal(decimal.RequireFromString("9000")))
}

func TestMarginSystem_ADLRanking(t *testing.T) {
	ms := NewMarginSystem()

	positions := []*model.TradingPosition{
		{
			ID:            "pos1",
			UserID:        "user1",
			Margin:        decimal.RequireFromString("1000"),
			UnrealizedPnL: decimal.RequireFromString("500"), // ratio = 0.5
			Status:        model.PositionStatusOpen,
		},
		{
			ID:            "pos2",
			UserID:        "user2",
			Margin:        decimal.RequireFromString("1000"),
			UnrealizedPnL: decimal.RequireFromString("2000"), // ratio = 2.0
			Status:        model.PositionStatusOpen,
		},
		{
			ID:            "pos3",
			UserID:        "user3",
			Margin:        decimal.RequireFromString("1000"),
			UnrealizedPnL: decimal.RequireFromString("-200"), // ratio = -0.2
			Status:        model.PositionStatusOpen,
		},
	}

	rankings := ms.CalculateADLRanking(positions)

	assert.Len(t, rankings, 3)
	// Should be sorted by profit ratio descending
	assert.Equal(t, "pos2", rankings[0].PositionID) // ratio 2.0
	assert.Equal(t, "pos1", rankings[1].PositionID) // ratio 0.5
	assert.Equal(t, "pos3", rankings[2].PositionID) // ratio -0.2
}
