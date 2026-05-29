package settlement

import (
	"sync"
	"time"

	"github.com/shopspring/decimal"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/trading-service/internal/model"
)

var (
	makerFeeRate = decimal.NewFromFloat(0.0002) // 0.02%
	takerFeeRate = decimal.NewFromFloat(0.0005) // 0.05%
)

// FeeResult holds the calculated fee for a trade.
type FeeResult struct {
	Fee  decimal.Decimal
	Rate decimal.Decimal
}

// FundingResult holds the result of a funding rate calculation.
type FundingResult struct {
	Pair         string
	Rate         decimal.Decimal
	CalculatedAt time.Time
}

// SettlementResult holds the results of a settlement cycle.
type SettlementResult struct {
	ProcessedPositions int
	TotalFunding       decimal.Decimal
	SocializedLoss     decimal.Decimal
	SettledAt          time.Time
}

// SettlementEngine handles mark-to-market settlement, funding rate calculation, and fee processing.
type SettlementEngine struct {
	mu           sync.RWMutex
	markPrices   map[string]decimal.Decimal // pair -> mark price
	indexPrices  map[string]decimal.Decimal // pair -> index price
	fundingRates map[string]decimal.Decimal // pair -> current funding rate
}

// NewSettlementEngine creates a new SettlementEngine.
func NewSettlementEngine() *SettlementEngine {
	return &SettlementEngine{
		markPrices:   make(map[string]decimal.Decimal),
		indexPrices:  make(map[string]decimal.Decimal),
		fundingRates: make(map[string]decimal.Decimal),
	}
}

// SetMarkPrice sets the mark price for a pair.
func (e *SettlementEngine) SetMarkPrice(pair string, price decimal.Decimal) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.markPrices[pair] = price
}

// SetIndexPrice sets the index price for a pair.
func (e *SettlementEngine) SetIndexPrice(pair string, price decimal.Decimal) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.indexPrices[pair] = price
}

// GetMarkPrice returns the mark price for a pair.
func (e *SettlementEngine) GetMarkPrice(pair string) decimal.Decimal {
	e.mu.RLock()
	defer e.mu.RUnlock()
	price, ok := e.markPrices[pair]
	if !ok {
		return decimal.Zero
	}
	return price
}

// CalculateFundingRate calculates the funding rate for a pair based on mark-index price spread.
// Funding rate = (mark price - index price) / index price, clamped to [-0.05%, 0.05%] per period.
func (e *SettlementEngine) CalculateFundingRate(pair string) FundingResult {
	e.mu.Lock()
	defer e.mu.Unlock()

	markPrice := e.markPrices[pair]
	indexPrice := e.indexPrices[pair]

	if indexPrice.IsZero() {
		return FundingResult{Pair: pair, Rate: decimal.Zero, CalculatedAt: time.Now()}
	}

	// Funding rate = (mark - index) / index
	rate := markPrice.Sub(indexPrice).Div(indexPrice)

	// Clamp to [-0.0005, 0.0005] (0.05%)
	maxRate := decimal.NewFromFloat(0.0005)
	minRate := maxRate.Neg()
	if rate.GreaterThan(maxRate) {
		rate = maxRate
	} else if rate.LessThan(minRate) {
		rate = minRate
	}

	e.fundingRates[pair] = rate
	return FundingResult{Pair: pair, Rate: rate, CalculatedAt: time.Now()}
}

// CalculateFees calculates trading fees for a trade.
func (e *SettlementEngine) CalculateFees(trade *model.Trade) FeeResult {
	notional := trade.Price.Mul(trade.Quantity)
	if trade.IsMaker {
		return FeeResult{
			Fee:  notional.Mul(makerFeeRate),
			Rate: makerFeeRate,
		}
	}
	return FeeResult{
		Fee:  notional.Mul(takerFeeRate),
		Rate: takerFeeRate,
	}
}

// ProcessSettlement runs settlement on a set of positions: applies funding rates and marks to market.
// Returns the settlement result.
func (e *SettlementEngine) ProcessSettlement(positions []*model.TradingPosition, balances map[string]*model.Balance) SettlementResult {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := SettlementResult{
		SettledAt: time.Now(),
	}

	for _, pos := range positions {
		if pos.Status != model.PositionStatusOpen {
			continue
		}

		// Apply mark-to-market
		markPrice, exists := e.markPrices[pos.Pair]
		if exists {
			pos.MarkToMarket(markPrice)
		}

		// Apply funding rate
		fundingRate, exists := e.fundingRates[pos.Pair]
		if exists && !fundingRate.IsZero() {
			notional := pos.MarkPrice.Mul(pos.Size)
			funding := notional.Mul(fundingRate)

			// Longs pay shorts when rate is positive, shorts pay longs when negative
			if pos.Side == model.SideBuy {
				// Long pays funding
				funding = funding.Neg()
			}

			// Apply to balance if available
			if bal, ok := balances[pos.UserID]; ok {
				if funding.IsPositive() {
					_ = bal.Credit(funding)
				} else {
					if err := bal.Debit(funding.Abs()); err != nil {
						// User cannot cover funding payment; track as socialized loss
						result.SocializedLoss = result.SocializedLoss.Add(funding.Abs())
					}
				}
			}

			result.TotalFunding = result.TotalFunding.Add(funding.Abs())
		}

		result.ProcessedPositions++
	}

	return result
}

// RunSettlement performs a full settlement cycle (mark-to-market every 8h).
func (e *SettlementEngine) RunSettlement(positions []*model.TradingPosition, balances map[string]*model.Balance) SettlementResult {
	return e.ProcessSettlement(positions, balances)
}
