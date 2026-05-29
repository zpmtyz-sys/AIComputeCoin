package model

import (
	"time"

	"github.com/shopspring/decimal"
)

// TradingPosition represents a leveraged trading position with margin tracking.
type TradingPosition struct {
	ID               string          `json:"id"`
	UserID           string          `json:"user_id"`
	Pair             string          `json:"pair"`
	Side             Side            `json:"side"`
	Size             decimal.Decimal `json:"size"`
	EntryPrice       decimal.Decimal `json:"entry_price"`
	MarkPrice        decimal.Decimal `json:"mark_price"`
	UnrealizedPnL    decimal.Decimal `json:"unrealized_pnl"`
	RealizedPnL      decimal.Decimal `json:"realized_pnl"`
	Margin           decimal.Decimal `json:"margin"`
	LiquidationPrice decimal.Decimal `json:"liquidation_price"`
	Leverage         decimal.Decimal `json:"leverage"`
	Status           PositionStatus  `json:"status"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

// MarkToMarket updates the position's mark price and recalculates unrealized P&L.
func (p *TradingPosition) MarkToMarket(markPrice decimal.Decimal) {
	p.MarkPrice = markPrice
	p.UnrealizedPnL = p.CalculatePnL()
	p.UpdatedAt = time.Now()
}

// CalculatePnL calculates the unrealized P&L based on mark price.
func (p *TradingPosition) CalculatePnL() decimal.Decimal {
	if p.Size.IsZero() {
		return decimal.Zero
	}
	if p.Side == SideBuy {
		// Long: PnL = (markPrice - entryPrice) * size
		return p.MarkPrice.Sub(p.EntryPrice).Mul(p.Size)
	}
	// Short: PnL = (entryPrice - markPrice) * size
	return p.EntryPrice.Sub(p.MarkPrice).Mul(p.Size)
}

// IsLiquidatable returns true if the position should be liquidated.
// A position is liquidatable when the margin ratio drops below the maintenance level (5%).
func (p *TradingPosition) IsLiquidatable() bool {
	if p.Margin.IsZero() || p.Status != PositionStatusOpen {
		return false
	}
	// Effective margin = initial margin + unrealized PnL
	effectiveMargin := p.Margin.Add(p.UnrealizedPnL)
	// Notional value at mark price
	notional := p.MarkPrice.Mul(p.Size)
	if notional.IsZero() {
		return false
	}
	// Margin ratio = effective margin / notional
	marginRatio := effectiveMargin.Div(notional)
	// Maintenance margin is 5%
	maintenanceRate := decimal.NewFromFloat(0.05)
	return marginRatio.LessThan(maintenanceRate)
}
