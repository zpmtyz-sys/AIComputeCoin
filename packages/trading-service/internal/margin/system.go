package margin

import (
	"sort"
	"sync"

	"github.com/shopspring/decimal"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/model"
)

var (
	initialMarginRate     = decimal.NewFromFloat(0.10) // 10%
	maintenanceMarginRate = decimal.NewFromFloat(0.05) // 5%
)

// LiquidationResult holds the result of a liquidation.
type LiquidationResult struct {
	PositionID     string
	UserID         string
	LiquidatedSize decimal.Decimal
	LiquidatedAt   decimal.Decimal // price at liquidation
	InsurancePaid  decimal.Decimal
}

// ADLRanking holds a position's ranking for auto-deleveraging.
type ADLRanking struct {
	PositionID  string
	UserID      string
	ProfitRatio decimal.Decimal
}

// MarginSystem handles margin calculations, liquidation, and insurance fund.
type MarginSystem struct {
	mu            sync.RWMutex
	insuranceFund decimal.Decimal
}

// NewMarginSystem creates a new MarginSystem.
func NewMarginSystem() *MarginSystem {
	return &MarginSystem{
		insuranceFund: decimal.Zero,
	}
}

// CalculateInitialMargin calculates the initial margin required for an order.
// Initial margin = 10% of notional value.
func (ms *MarginSystem) CalculateInitialMargin(order *model.Order) decimal.Decimal {
	var notional decimal.Decimal
	if order.OrderType == model.OrderTypeMarket {
		notional = order.Quantity // simplified: uses quantity for market orders
	} else {
		notional = order.Price.Mul(order.Quantity)
	}
	return notional.Mul(initialMarginRate)
}

// CalculateMaintenanceMargin calculates the maintenance margin for a position.
// Maintenance margin = 5% of notional value at mark price.
func (ms *MarginSystem) CalculateMaintenanceMargin(position *model.TradingPosition) decimal.Decimal {
	notional := position.MarkPrice.Mul(position.Size)
	return notional.Mul(maintenanceMarginRate)
}

// CheckLiquidation returns true if the position should be liquidated.
// Liquidation occurs when effective margin < maintenance margin.
func (ms *MarginSystem) CheckLiquidation(position *model.TradingPosition, markPrice decimal.Decimal) bool {
	if position.Status != model.PositionStatusOpen || position.Size.IsZero() {
		return false
	}

	// Update PnL at the given mark price
	var unrealizedPnL decimal.Decimal
	if position.Side == model.SideBuy {
		unrealizedPnL = markPrice.Sub(position.EntryPrice).Mul(position.Size)
	} else {
		unrealizedPnL = position.EntryPrice.Sub(markPrice).Mul(position.Size)
	}

	// Effective margin = initial margin + unrealized PnL
	effectiveMargin := position.Margin.Add(unrealizedPnL)

	// Maintenance margin = 5% of notional at mark price
	notional := markPrice.Mul(position.Size)
	maintenanceMargin := notional.Mul(maintenanceMarginRate)

	return effectiveMargin.LessThan(maintenanceMargin)
}

// TriggerLiquidation processes a liquidation for a position.
func (ms *MarginSystem) TriggerLiquidation(position *model.TradingPosition) LiquidationResult {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	result := LiquidationResult{
		PositionID:     position.ID,
		UserID:         position.UserID,
		LiquidatedSize: position.Size,
		LiquidatedAt:   position.MarkPrice,
	}

	// Calculate remaining margin after liquidation
	remainingMargin := position.Margin.Add(position.UnrealizedPnL)

	// If remaining margin is negative, insurance fund covers the loss
	if remainingMargin.IsNegative() {
		insurancePayment := remainingMargin.Abs()
		if ms.insuranceFund.GreaterThanOrEqual(insurancePayment) {
			ms.insuranceFund = ms.insuranceFund.Sub(insurancePayment)
			result.InsurancePaid = insurancePayment
		} else {
			result.InsurancePaid = ms.insuranceFund
			ms.insuranceFund = decimal.Zero
		}
	} else {
		// Remaining margin goes to insurance fund
		ms.insuranceFund = ms.insuranceFund.Add(remainingMargin)
	}

	position.Status = model.PositionStatusLiquidated
	position.Size = decimal.Zero

	return result
}

// GetInsuranceFund returns the current insurance fund balance.
func (ms *MarginSystem) GetInsuranceFund() decimal.Decimal {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.insuranceFund
}

// AddToInsuranceFund adds funds to the insurance fund.
func (ms *MarginSystem) AddToInsuranceFund(amount decimal.Decimal) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.insuranceFund = ms.insuranceFund.Add(amount)
}

// CalculateADLRanking ranks positions for auto-deleveraging by profit ratio (highest first).
func (ms *MarginSystem) CalculateADLRanking(positions []*model.TradingPosition) []ADLRanking {
	var rankings []ADLRanking

	for _, pos := range positions {
		if pos.Status != model.PositionStatusOpen || pos.Margin.IsZero() {
			continue
		}
		profitRatio := pos.UnrealizedPnL.Div(pos.Margin)
		rankings = append(rankings, ADLRanking{
			PositionID:  pos.ID,
			UserID:      pos.UserID,
			ProfitRatio: profitRatio,
		})
	}

	// Sort by profit ratio descending (most profitable first for ADL)
	sort.Slice(rankings, func(i, j int) bool {
		return rankings[i].ProfitRatio.GreaterThan(rankings[j].ProfitRatio)
	})

	return rankings
}
