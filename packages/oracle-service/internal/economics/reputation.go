package economics

import (
	"sync"
	"time"

	"github.com/shopspring/decimal"
)

const maxHistorySize = 100

// ReputationTracker tracks oracle reputation using weighted moving average.
type ReputationTracker struct {
	mu     sync.RWMutex
	scores map[string]*reputationState
}

type reputationState struct {
	history []decimal.Decimal
	current decimal.Decimal
}

// NewReputationTracker creates a new ReputationTracker.
func NewReputationTracker() *ReputationTracker {
	return &ReputationTracker{
		scores: make(map[string]*reputationState),
	}
}

// RecordReport records a new accuracy score (0-1) for an oracle.
func (r *ReputationTracker) RecordReport(oracleID string, accuracy decimal.Decimal) {
	r.mu.Lock()
	defer r.mu.Unlock()

	state, ok := r.scores[oracleID]
	if !ok {
		state = &reputationState{
			history: make([]decimal.Decimal, 0, maxHistorySize),
			current: decimal.Zero,
		}
		r.scores[oracleID] = state
	}

	// Add to history, keep only last 100
	state.history = append(state.history, accuracy)
	if len(state.history) > maxHistorySize {
		state.history = state.history[len(state.history)-maxHistorySize:]
	}

	// Calculate weighted moving average
	state.current = weightedMovingAverage(state.history)
}

// GetScore returns the current reputation score for an oracle.
func (r *ReputationTracker) GetScore(oracleID string) *ReputationScore {
	r.mu.RLock()
	defer r.mu.RUnlock()

	state, ok := r.scores[oracleID]
	if !ok {
		return &ReputationScore{
			OracleID:     oracleID,
			Score:        decimal.Zero,
			TotalReports: 0,
			LastUpdated:  time.Now(),
		}
	}

	return &ReputationScore{
		OracleID:     oracleID,
		Score:        state.current,
		TotalReports: len(state.history),
		LastUpdated:  time.Now(),
	}
}

// weightedMovingAverage calculates the weighted moving average.
// More recent reports have higher weight.
func weightedMovingAverage(history []decimal.Decimal) decimal.Decimal {
	if len(history) == 0 {
		return decimal.Zero
	}

	totalWeight := decimal.Zero
	weightedSum := decimal.Zero

	for i, score := range history {
		// Linear weight: position + 1 (most recent = highest weight)
		weight := decimal.NewFromInt(int64(i + 1))
		totalWeight = totalWeight.Add(weight)
		weightedSum = weightedSum.Add(score.Mul(weight))
	}

	return weightedSum.Div(totalWeight)
}
