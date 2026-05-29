package economics

import (
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
)

// Slasher applies slashing penalties to oracle stakes.
type Slasher struct {
	stakeStore *InMemoryStakeStore
	logger     *slog.Logger
}

// NewSlasher creates a new Slasher.
func NewSlasher(stakeStore *InMemoryStakeStore, logger *slog.Logger) *Slasher {
	if logger == nil {
		logger = slog.Default()
	}
	return &Slasher{
		stakeStore: stakeStore,
		logger:     logger,
	}
}

// Slash applies a slashing penalty to an oracle.
func (s *Slasher) Slash(oracleID string, slashType SlashType, reason string) (*SlashEvent, error) {
	stake, err := s.stakeStore.GetStake(oracleID)
	if err != nil {
		return nil, err
	}

	var rate decimal.Decimal
	var banned bool

	switch slashType {
	case SlashTypeDeviation:
		rate = SlashDeviation
	case SlashTypeMissed:
		rate = SlashMissed
	case SlashTypeFraud:
		rate = SlashFraud
		banned = true
	}

	slashAmount := stake.Mul(rate)
	newStake := stake.Sub(slashAmount)
	if newStake.IsNegative() {
		newStake = decimal.Zero
	}

	s.stakeStore.SetStake(oracleID, newStake)

	event := &SlashEvent{
		OracleID:  oracleID,
		Amount:    slashAmount,
		Type:      slashType,
		Reason:    reason,
		Banned:    banned,
		Timestamp: time.Now(),
	}

	s.logger.Info("oracle slashed",
		"oracle_id", oracleID,
		"amount", slashAmount.String(),
		"type", slashType,
		"reason", reason,
		"banned", banned,
	)

	return event, nil
}
