package position

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/model"
)

func TestPositionTracker_OpenPosition(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		pair       string
		side       model.Side
		size       string
		entryPrice string
		margin     string
		wantSide   model.Side
	}{
		{
			name:       "open long position",
			userID:     "user1",
			pair:       "BTC-USDT",
			side:       model.SideBuy,
			size:       "0.5",
			entryPrice: "50000",
			margin:     "2500",
			wantSide:   model.SideBuy,
		},
		{
			name:       "open short position",
			userID:     "user1",
			pair:       "ETH-USDT",
			side:       model.SideSell,
			size:       "10",
			entryPrice: "3000",
			margin:     "3000",
			wantSide:   model.SideSell,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewPositionTracker()
			pos := tracker.OpenPosition(
				tt.userID,
				tt.pair,
				tt.side,
				decimal.RequireFromString(tt.size),
				decimal.RequireFromString(tt.entryPrice),
				decimal.RequireFromString(tt.margin),
			)

			require.NotNil(t, pos)
			assert.NotEmpty(t, pos.ID)
			assert.Equal(t, tt.userID, pos.UserID)
			assert.Equal(t, tt.pair, pos.Pair)
			assert.Equal(t, tt.wantSide, pos.Side)
			assert.True(t, pos.Size.Equal(decimal.RequireFromString(tt.size)))
			assert.True(t, pos.EntryPrice.Equal(decimal.RequireFromString(tt.entryPrice)))
			assert.Equal(t, model.PositionStatusOpen, pos.Status)
			assert.True(t, pos.UnrealizedPnL.IsZero())

			// Verify position is retrievable
			positions := tracker.GetPositions(tt.userID)
			assert.Len(t, positions, 1)
			assert.Equal(t, pos.ID, positions[0].ID)
		})
	}
}

func TestPositionTracker_ClosePosition(t *testing.T) {
	tests := []struct {
		name       string
		side       model.Side
		size       string
		entryPrice string
		closePrice string
		closeQty   string
		wantPnL    string
		wantStatus model.PositionStatus
		wantSize   string
	}{
		{
			name:       "close full long position with profit",
			side:       model.SideBuy,
			size:       "1",
			entryPrice: "50000",
			closePrice: "55000",
			closeQty:   "1",
			wantPnL:    "5000",
			wantStatus: model.PositionStatusClosed,
			wantSize:   "0",
		},
		{
			name:       "close full short position with profit",
			side:       model.SideSell,
			size:       "2",
			entryPrice: "3000",
			closePrice: "2800",
			closeQty:   "2",
			wantPnL:    "400",
			wantStatus: model.PositionStatusClosed,
			wantSize:   "0",
		},
		{
			name:       "close partial long position",
			side:       model.SideBuy,
			size:       "10",
			entryPrice: "100",
			closePrice: "120",
			closeQty:   "4",
			wantPnL:    "80",
			wantStatus: model.PositionStatusOpen,
			wantSize:   "6",
		},
		{
			name:       "close full long position with loss",
			side:       model.SideBuy,
			size:       "1",
			entryPrice: "50000",
			closePrice: "48000",
			closeQty:   "1",
			wantPnL:    "-2000",
			wantStatus: model.PositionStatusClosed,
			wantSize:   "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewPositionTracker()
			pos := tracker.OpenPosition(
				"user1",
				"BTC-USDT",
				tt.side,
				decimal.RequireFromString(tt.size),
				decimal.RequireFromString(tt.entryPrice),
				decimal.RequireFromString("5000"),
			)

			pnl, err := tracker.ClosePosition(
				"user1",
				pos.ID,
				decimal.RequireFromString(tt.closePrice),
				decimal.RequireFromString(tt.closeQty),
			)

			require.NoError(t, err)
			assert.True(t, pnl.Equal(decimal.RequireFromString(tt.wantPnL)),
				"expected PnL %s, got %s", tt.wantPnL, pnl.String())

			// Check position state
			updatedPos := tracker.GetPositionByID("user1", pos.ID)
			assert.Equal(t, tt.wantStatus, updatedPos.Status)
			assert.True(t, updatedPos.Size.Equal(decimal.RequireFromString(tt.wantSize)))
		})
	}
}

func TestPositionTracker_FIFO(t *testing.T) {
	tracker := NewPositionTracker()

	// Open two positions for the same user
	pos1 := tracker.OpenPosition("user1", "BTC-USDT", model.SideBuy,
		decimal.RequireFromString("1"), decimal.RequireFromString("50000"), decimal.RequireFromString("5000"))
	pos2 := tracker.OpenPosition("user1", "BTC-USDT", model.SideBuy,
		decimal.RequireFromString("1"), decimal.RequireFromString("52000"), decimal.RequireFromString("5200"))

	// Close the first one (FIFO)
	pnl, err := tracker.ClosePosition("user1", pos1.ID,
		decimal.RequireFromString("55000"), decimal.RequireFromString("1"))
	require.NoError(t, err)
	assert.True(t, pnl.Equal(decimal.RequireFromString("5000"))) // 55000 - 50000

	// Second position should still be open
	positions := tracker.GetPositions("user1")
	assert.Len(t, positions, 1)
	assert.Equal(t, pos2.ID, positions[0].ID)
}

func TestPositionTracker_UpdateMarkPrice(t *testing.T) {
	tests := []struct {
		name       string
		side       model.Side
		entryPrice string
		markPrice  string
		size       string
		wantPnL    string
	}{
		{
			name:       "long position mark up",
			side:       model.SideBuy,
			entryPrice: "50000",
			markPrice:  "52000",
			size:       "1",
			wantPnL:    "2000",
		},
		{
			name:       "long position mark down",
			side:       model.SideBuy,
			entryPrice: "50000",
			markPrice:  "48000",
			size:       "1",
			wantPnL:    "-2000",
		},
		{
			name:       "short position mark down (profit)",
			side:       model.SideSell,
			entryPrice: "3000",
			markPrice:  "2800",
			size:       "5",
			wantPnL:    "1000",
		},
		{
			name:       "short position mark up (loss)",
			side:       model.SideSell,
			entryPrice: "3000",
			markPrice:  "3200",
			size:       "5",
			wantPnL:    "-1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewPositionTracker()
			tracker.OpenPosition("user1", "BTC-USDT", tt.side,
				decimal.RequireFromString(tt.size),
				decimal.RequireFromString(tt.entryPrice),
				decimal.RequireFromString("5000"),
			)

			tracker.UpdateMarkPrice("BTC-USDT", decimal.RequireFromString(tt.markPrice))

			positions := tracker.GetPositions("user1")
			require.Len(t, positions, 1)
			assert.True(t, positions[0].UnrealizedPnL.Equal(decimal.RequireFromString(tt.wantPnL)),
				"expected PnL %s, got %s", tt.wantPnL, positions[0].UnrealizedPnL.String())
		})
	}
}

func TestPositionTracker_ClosePosition_NotFound(t *testing.T) {
	tracker := NewPositionTracker()
	_, err := tracker.ClosePosition("user1", "nonexistent",
		decimal.RequireFromString("50000"), decimal.RequireFromString("1"))
	assert.ErrorIs(t, err, ErrPositionNotFound)
}
