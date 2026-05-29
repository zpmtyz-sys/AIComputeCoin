package economics

import (
	"log/slog"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStakeValidation(t *testing.T) {
	tests := []struct {
		name      string
		oracleID  string
		stake     decimal.Decimal
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "stake above minimum",
			oracleID:  "oracle-1",
			stake:     decimal.RequireFromString("15000"),
			wantValid: true,
		},
		{
			name:      "stake at minimum",
			oracleID:  "oracle-2",
			stake:     decimal.RequireFromString("10000"),
			wantValid: true,
		},
		{
			name:      "stake below minimum",
			oracleID:  "oracle-3",
			stake:     decimal.RequireFromString("9999"),
			wantValid: false,
		},
		{
			name:      "zero stake",
			oracleID:  "oracle-4",
			stake:     decimal.Zero,
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewInMemoryStakeStore()
			store.SetStake(tt.oracleID, tt.stake)

			valid, err := store.ValidateStake(tt.oracleID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantValid, valid)
			}
		})
	}
}

func TestStakeValidation_OracleNotFound(t *testing.T) {
	store := NewInMemoryStakeStore()
	_, err := store.ValidateStake("nonexistent")
	assert.Error(t, err)
}

func TestSlashing_Deviation(t *testing.T) {
	store := NewInMemoryStakeStore()
	store.SetStake("oracle-1", decimal.RequireFromString("20000"))
	slasher := NewSlasher(store, slog.Default())

	event, err := slasher.Slash("oracle-1", SlashTypeDeviation, "exceeded 2 sigma")
	require.NoError(t, err)

	// 5% of 20000 = 1000
	assert.True(t, event.Amount.Equal(decimal.RequireFromString("1000")))
	assert.False(t, event.Banned)

	// Remaining stake should be 19000
	remaining, _ := store.GetStake("oracle-1")
	assert.True(t, remaining.Equal(decimal.RequireFromString("19000")))
}

func TestSlashing_Missed(t *testing.T) {
	store := NewInMemoryStakeStore()
	store.SetStake("oracle-1", decimal.RequireFromString("10000"))
	slasher := NewSlasher(store, slog.Default())

	event, err := slasher.Slash("oracle-1", SlashTypeMissed, "missed deadline")
	require.NoError(t, err)

	// 1% of 10000 = 100
	assert.True(t, event.Amount.Equal(decimal.RequireFromString("100")))
	assert.False(t, event.Banned)

	remaining, _ := store.GetStake("oracle-1")
	assert.True(t, remaining.Equal(decimal.RequireFromString("9900")))
}

func TestSlashing_Fraud(t *testing.T) {
	store := NewInMemoryStakeStore()
	store.SetStake("oracle-1", decimal.RequireFromString("50000"))
	slasher := NewSlasher(store, slog.Default())

	event, err := slasher.Slash("oracle-1", SlashTypeFraud, "proven fraud")
	require.NoError(t, err)

	// 100% of 50000 = 50000
	assert.True(t, event.Amount.Equal(decimal.RequireFromString("50000")))
	assert.True(t, event.Banned)

	remaining, _ := store.GetStake("oracle-1")
	assert.True(t, remaining.Equal(decimal.Zero))
}

func TestSlashing_OracleNotFound(t *testing.T) {
	store := NewInMemoryStakeStore()
	slasher := NewSlasher(store, slog.Default())

	_, err := slasher.Slash("nonexistent", SlashTypeDeviation, "test")
	assert.Error(t, err)
}

func TestRewardDistribution(t *testing.T) {
	distributor := NewRewardDistributor(slog.Default())

	tests := []struct {
		name          string
		computeValue  decimal.Decimal
		contributions map[string]decimal.Decimal
		wantCount     int
		wantTotal     decimal.Decimal
	}{
		{
			name:         "single oracle",
			computeValue: decimal.RequireFromString("1000000"),
			contributions: map[string]decimal.Decimal{
				"oracle-1": decimal.RequireFromString("100"),
			},
			wantCount: 1,
			wantTotal: decimal.RequireFromString("1000"), // 0.1% of 1M
		},
		{
			name:         "two oracles equal contribution",
			computeValue: decimal.RequireFromString("1000000"),
			contributions: map[string]decimal.Decimal{
				"oracle-1": decimal.RequireFromString("50"),
				"oracle-2": decimal.RequireFromString("50"),
			},
			wantCount: 2,
			wantTotal: decimal.RequireFromString("1000"),
		},
		{
			name:          "no contributions",
			computeValue:  decimal.RequireFromString("1000000"),
			contributions: map[string]decimal.Decimal{},
			wantCount:     0,
			wantTotal:     decimal.Zero,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rewards := distributor.DistributeRewards(1, tt.computeValue, tt.contributions)
			assert.Len(t, rewards, tt.wantCount)

			total := decimal.Zero
			for _, r := range rewards {
				total = total.Add(r.Amount)
			}
			assert.True(t, total.Equal(tt.wantTotal),
				"expected total %s, got %s", tt.wantTotal.String(), total.String())
		})
	}
}

func TestRewardDistribution_Proportional(t *testing.T) {
	distributor := NewRewardDistributor(slog.Default())

	contributions := map[string]decimal.Decimal{
		"oracle-1": decimal.RequireFromString("75"),
		"oracle-2": decimal.RequireFromString("25"),
	}

	rewards := distributor.DistributeRewards(1, decimal.RequireFromString("100000"), contributions)
	require.Len(t, rewards, 2)

	// Find each oracle's reward
	rewardMap := make(map[string]decimal.Decimal)
	for _, r := range rewards {
		rewardMap[r.OracleID] = r.Amount
	}

	// oracle-1 should get 75% of pool, oracle-2 gets 25%
	// Pool = 100000 * 0.001 = 100
	assert.True(t, rewardMap["oracle-1"].Equal(decimal.RequireFromString("75")))
	assert.True(t, rewardMap["oracle-2"].Equal(decimal.RequireFromString("25")))
}

func TestReputation_WeightedMovingAverage(t *testing.T) {
	tracker := NewReputationTracker()

	// Add reports with increasing accuracy
	for i := 0; i < 10; i++ {
		accuracy := decimal.NewFromFloat(float64(i+1) / 10.0) // 0.1, 0.2, ..., 1.0
		tracker.RecordReport("oracle-1", accuracy)
	}

	score := tracker.GetScore("oracle-1")
	assert.Equal(t, 10, score.TotalReports)

	// Weighted average should be skewed toward recent (higher) values
	// With linear weights 1..10, the weighted average of 0.1..1.0 is:
	// sum(i * (i/10)) / sum(i) for i=1..10 = (1*0.1 + 2*0.2 + ... + 10*1.0) / 55
	// = (0.1 + 0.4 + 0.9 + 1.6 + 2.5 + 3.6 + 4.9 + 6.4 + 8.1 + 10.0) / 55
	// = 38.5 / 55 = 0.7
	assert.True(t, score.Score.Equal(decimal.RequireFromString("0.7")),
		"expected 0.7, got %s", score.Score.String())
}

func TestReputation_MaxHistory(t *testing.T) {
	tracker := NewReputationTracker()

	// Add 150 reports (should keep only last 100)
	for i := 0; i < 150; i++ {
		tracker.RecordReport("oracle-1", decimal.RequireFromString("0.9"))
	}

	score := tracker.GetScore("oracle-1")
	assert.Equal(t, 100, score.TotalReports)
}

func TestReputation_NewOracle(t *testing.T) {
	tracker := NewReputationTracker()

	score := tracker.GetScore("unknown-oracle")
	assert.True(t, score.Score.IsZero())
	assert.Equal(t, 0, score.TotalReports)
}

func TestReputation_Convergence(t *testing.T) {
	tracker := NewReputationTracker()

	// Start with bad reports, then consistently good
	for i := 0; i < 20; i++ {
		tracker.RecordReport("oracle-1", decimal.RequireFromString("0.2"))
	}
	scoreBad := tracker.GetScore("oracle-1")

	for i := 0; i < 80; i++ {
		tracker.RecordReport("oracle-1", decimal.RequireFromString("0.95"))
	}
	scoreGood := tracker.GetScore("oracle-1")

	// After many good reports, score should be much higher
	assert.True(t, scoreGood.Score.GreaterThan(scoreBad.Score))
	assert.True(t, scoreGood.Score.GreaterThan(decimal.RequireFromString("0.8")))
}
