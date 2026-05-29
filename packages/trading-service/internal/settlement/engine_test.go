package settlement

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/model"
)

func TestSettlementEngine_CalculateFundingRate(t *testing.T) {
	tests := []struct {
		name       string
		markPrice  string
		indexPrice string
		wantRate   string
	}{
		{
			name:       "positive funding rate (mark > index)",
			markPrice:  "50010",
			indexPrice: "50000",
			wantRate:   "0.0002", // (50010-50000)/50000 = 0.0002
		},
		{
			name:       "negative funding rate (mark < index)",
			markPrice:  "49990",
			indexPrice: "50000",
			wantRate:   "-0.0002", // (49990-50000)/50000 = -0.0002
		},
		{
			name:       "zero funding rate (mark == index)",
			markPrice:  "50000",
			indexPrice: "50000",
			wantRate:   "0",
		},
		{
			name:       "clamped positive rate",
			markPrice:  "51000",
			indexPrice: "50000",
			wantRate:   "0.0005", // (51000-50000)/50000 = 0.02, clamped to 0.0005
		},
		{
			name:       "clamped negative rate",
			markPrice:  "49000",
			indexPrice: "50000",
			wantRate:   "-0.0005", // (49000-50000)/50000 = -0.02, clamped to -0.0005
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewSettlementEngine()
			engine.SetMarkPrice("BTC-USDT", decimal.RequireFromString(tt.markPrice))
			engine.SetIndexPrice("BTC-USDT", decimal.RequireFromString(tt.indexPrice))

			result := engine.CalculateFundingRate("BTC-USDT")

			assert.Equal(t, "BTC-USDT", result.Pair)
			assert.True(t, result.Rate.Equal(decimal.RequireFromString(tt.wantRate)),
				"expected rate %s, got %s", tt.wantRate, result.Rate.String())
		})
	}
}

func TestSettlementEngine_CalculateFees(t *testing.T) {
	tests := []struct {
		name     string
		price    string
		quantity string
		isMaker  bool
		wantFee  string
		wantRate string
	}{
		{
			name:     "maker fee",
			price:    "50000",
			quantity: "1",
			isMaker:  true,
			wantFee:  "10",     // 50000 * 1 * 0.0002 = 10
			wantRate: "0.0002", // 0.02%
		},
		{
			name:     "taker fee",
			price:    "50000",
			quantity: "1",
			isMaker:  false,
			wantFee:  "25",     // 50000 * 1 * 0.0005 = 25
			wantRate: "0.0005", // 0.05%
		},
		{
			name:     "maker fee small trade",
			price:    "3000",
			quantity: "0.5",
			isMaker:  true,
			wantFee:  "0.3", // 3000 * 0.5 * 0.0002 = 0.3
			wantRate: "0.0002",
		},
		{
			name:     "taker fee large trade",
			price:    "100",
			quantity: "1000",
			isMaker:  false,
			wantFee:  "50", // 100 * 1000 * 0.0005 = 50
			wantRate: "0.0005",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine := NewSettlementEngine()
			trade := &model.Trade{
				Price:    decimal.RequireFromString(tt.price),
				Quantity: decimal.RequireFromString(tt.quantity),
				IsMaker:  tt.isMaker,
			}

			result := engine.CalculateFees(trade)

			assert.True(t, result.Fee.Equal(decimal.RequireFromString(tt.wantFee)),
				"expected fee %s, got %s", tt.wantFee, result.Fee.String())
			assert.True(t, result.Rate.Equal(decimal.RequireFromString(tt.wantRate)),
				"expected rate %s, got %s", tt.wantRate, result.Rate.String())
		})
	}
}

func TestSettlementEngine_ProcessSettlement(t *testing.T) {
	engine := NewSettlementEngine()
	engine.SetMarkPrice("BTC-USDT", decimal.RequireFromString("52000"))
	engine.SetIndexPrice("BTC-USDT", decimal.RequireFromString("50000"))

	// Calculate funding rate first
	engine.CalculateFundingRate("BTC-USDT")

	// Create positions
	positions := []*model.TradingPosition{
		{
			ID:         "pos1",
			UserID:     "user1",
			Pair:       "BTC-USDT",
			Side:       model.SideBuy,
			Size:       decimal.RequireFromString("1"),
			EntryPrice: decimal.RequireFromString("50000"),
			MarkPrice:  decimal.RequireFromString("50000"),
			Margin:     decimal.RequireFromString("5000"),
			Status:     model.PositionStatusOpen,
		},
		{
			ID:         "pos2",
			UserID:     "user2",
			Pair:       "BTC-USDT",
			Side:       model.SideSell,
			Size:       decimal.RequireFromString("1"),
			EntryPrice: decimal.RequireFromString("51000"),
			MarkPrice:  decimal.RequireFromString("51000"),
			Margin:     decimal.RequireFromString("5100"),
			Status:     model.PositionStatusOpen,
		},
	}

	balances := map[string]*model.Balance{
		"user1": model.NewBalance("user1", decimal.RequireFromString("10000")),
		"user2": model.NewBalance("user2", decimal.RequireFromString("10000")),
	}

	result := engine.ProcessSettlement(positions, balances)

	require.Equal(t, 2, result.ProcessedPositions)
	assert.False(t, result.TotalFunding.IsZero())
	assert.False(t, result.SettledAt.IsZero())

	// Verify mark-to-market was applied
	// pos1 (long): PnL = (52000 - 50000) * 1 = 2000
	assert.True(t, positions[0].UnrealizedPnL.Equal(decimal.RequireFromString("2000")),
		"expected PnL 2000, got %s", positions[0].UnrealizedPnL.String())
	// pos2 (short): PnL = (51000 - 52000) * 1 = -1000
	assert.True(t, positions[1].UnrealizedPnL.Equal(decimal.RequireFromString("-1000")),
		"expected PnL -1000, got %s", positions[1].UnrealizedPnL.String())
}

func TestSettlementEngine_FundingRate_ZeroIndex(t *testing.T) {
	engine := NewSettlementEngine()
	engine.SetMarkPrice("NEW-USDT", decimal.RequireFromString("100"))
	// No index price set

	result := engine.CalculateFundingRate("NEW-USDT")
	assert.True(t, result.Rate.IsZero())
}
