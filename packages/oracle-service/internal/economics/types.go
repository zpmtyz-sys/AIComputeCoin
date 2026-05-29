package economics

import (
	"time"

	"github.com/shopspring/decimal"
)

// MinimumStake is the minimum required stake in CC tokens.
var MinimumStake = decimal.RequireFromString("10000")

// Slashing percentages.
var (
	SlashDeviation = decimal.RequireFromString("0.05") // 5% for >2 sigma deviation
	SlashMissed    = decimal.RequireFromString("0.01") // 1% per missed deadline
	SlashFraud     = decimal.RequireFromString("1.00") // 100% for proven fraud
)

// RewardRate is 0.1% of verified compute value per epoch.
var RewardRate = decimal.RequireFromString("0.001")

// SlashType represents the reason for slashing.
type SlashType int

const (
	SlashTypeDeviation SlashType = iota
	SlashTypeMissed
	SlashTypeFraud
)

// Reward represents a reward distribution event.
type Reward struct {
	OracleID  string
	Amount    decimal.Decimal
	Epoch     uint64
	Reason    string
	Timestamp time.Time
}

// SlashEvent represents a slashing event.
type SlashEvent struct {
	OracleID  string
	Amount    decimal.Decimal
	Type      SlashType
	Reason    string
	Banned    bool
	Timestamp time.Time
}

// ReputationScore tracks an oracle's reputation.
type ReputationScore struct {
	OracleID     string
	Score        decimal.Decimal
	TotalReports int
	LastUpdated  time.Time
}

// StakeInfo holds staking information for an oracle.
type StakeInfo struct {
	OracleID string
	Amount   decimal.Decimal
	Locked   bool
}
