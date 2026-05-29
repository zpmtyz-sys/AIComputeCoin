package position

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/model"
)

// PositionTracker manages positions for all users with FIFO accounting.
type PositionTracker struct {
	mu        sync.RWMutex
	positions map[string][]*model.TradingPosition // userID -> positions (ordered by creation time for FIFO)
}

// NewPositionTracker creates a new PositionTracker.
func NewPositionTracker() *PositionTracker {
	return &PositionTracker{
		positions: make(map[string][]*model.TradingPosition),
	}
}

// OpenPosition opens a new position from a trade.
func (pt *PositionTracker) OpenPosition(userID, pair string, side model.Side, size, entryPrice, margin decimal.Decimal) *model.TradingPosition {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	leverage := decimal.Zero
	if !margin.IsZero() {
		notional := entryPrice.Mul(size)
		leverage = notional.Div(margin)
	}

	pos := &model.TradingPosition{
		ID:            uuid.New().String(),
		UserID:        userID,
		Pair:          pair,
		Side:          side,
		Size:          size,
		EntryPrice:    entryPrice,
		MarkPrice:     entryPrice,
		UnrealizedPnL: decimal.Zero,
		RealizedPnL:   decimal.Zero,
		Margin:        margin,
		Leverage:      leverage,
		Status:        model.PositionStatusOpen,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Calculate liquidation price
	pos.LiquidationPrice = calculateLiquidationPrice(pos)

	pt.positions[userID] = append(pt.positions[userID], pos)
	return pos
}

// ClosePosition closes a position partially or fully using FIFO accounting.
// Returns the realized PnL from closing.
func (pt *PositionTracker) ClosePosition(userID, positionID string, closePrice, closeQuantity decimal.Decimal) (decimal.Decimal, error) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	positions := pt.positions[userID]
	for i, pos := range positions {
		if pos.ID == positionID && pos.Status == model.PositionStatusOpen {
			if closeQuantity.GreaterThanOrEqual(pos.Size) {
				// Full close
				closeQuantity = pos.Size
				pos.Status = model.PositionStatusClosed
			}

			// Calculate realized PnL for the closed portion
			var pnl decimal.Decimal
			if pos.Side == model.SideBuy {
				pnl = closePrice.Sub(pos.EntryPrice).Mul(closeQuantity)
			} else {
				pnl = pos.EntryPrice.Sub(closePrice).Mul(closeQuantity)
			}

			pos.RealizedPnL = pos.RealizedPnL.Add(pnl)
			pos.Size = pos.Size.Sub(closeQuantity)
			pos.UpdatedAt = time.Now()

			if pos.Size.IsZero() {
				pos.Status = model.PositionStatusClosed
			}

			pt.positions[userID][i] = pos
			return pnl, nil
		}
	}
	return decimal.Zero, ErrPositionNotFound
}

// UpdateMarkPrice updates the mark price for all positions in a given pair.
func (pt *PositionTracker) UpdateMarkPrice(pair string, markPrice decimal.Decimal) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	for _, positions := range pt.positions {
		for _, pos := range positions {
			if pos.Pair == pair && pos.Status == model.PositionStatusOpen {
				pos.MarkToMarket(markPrice)
			}
		}
	}
}

// GetPositions returns all open positions for a user.
func (pt *PositionTracker) GetPositions(userID string) []*model.TradingPosition {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	var result []*model.TradingPosition
	for _, pos := range pt.positions[userID] {
		if pos.Status == model.PositionStatusOpen {
			result = append(result, pos)
		}
	}
	return result
}

// GetAllPositions returns all positions across all users (used for settlement).
func (pt *PositionTracker) GetAllPositions() []*model.TradingPosition {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	var result []*model.TradingPosition
	for _, positions := range pt.positions {
		for _, pos := range positions {
			if pos.Status == model.PositionStatusOpen {
				result = append(result, pos)
			}
		}
	}
	return result
}

// GetPositionByID returns a position by ID.
func (pt *PositionTracker) GetPositionByID(userID, positionID string) *model.TradingPosition {
	pt.mu.RLock()
	defer pt.mu.RUnlock()

	for _, pos := range pt.positions[userID] {
		if pos.ID == positionID {
			return pos
		}
	}
	return nil
}

// calculateLiquidationPrice computes the liquidation price for a position.
func calculateLiquidationPrice(pos *model.TradingPosition) decimal.Decimal {
	if pos.Size.IsZero() || pos.Margin.IsZero() {
		return decimal.Zero
	}
	// Maintenance margin rate is 5%
	maintenanceRate := decimal.NewFromFloat(0.05)
	notional := pos.EntryPrice.Mul(pos.Size)
	maintenanceMargin := notional.Mul(maintenanceRate)

	if pos.Side == model.SideBuy {
		// Long liquidation: entry - (margin - maintenance) / size
		marginBuffer := pos.Margin.Sub(maintenanceMargin)
		return pos.EntryPrice.Sub(marginBuffer.Div(pos.Size))
	}
	// Short liquidation: entry + (margin - maintenance) / size
	marginBuffer := pos.Margin.Sub(maintenanceMargin)
	return pos.EntryPrice.Add(marginBuffer.Div(pos.Size))
}
