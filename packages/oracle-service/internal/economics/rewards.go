package economics

import (
	"log/slog"
	"time"

	"github.com/shopspring/decimal"
)

// RewardDistributor distributes rewards to oracles.
type RewardDistributor struct {
	logger *slog.Logger
}

// NewRewardDistributor creates a new RewardDistributor.
func NewRewardDistributor(logger *slog.Logger) *RewardDistributor {
	if logger == nil {
		logger = slog.Default()
	}
	return &RewardDistributor{logger: logger}
}

// DistributeRewards distributes 0.1% of verified compute value per epoch
// proportionally to oracle contributions.
func (d *RewardDistributor) DistributeRewards(
	epoch uint64,
	computeValue decimal.Decimal,
	contributions map[string]decimal.Decimal,
) []Reward {
	// Total reward pool = 0.1% of compute value
	rewardPool := computeValue.Mul(RewardRate)

	// Calculate total contributions
	totalContribution := decimal.Zero
	for _, c := range contributions {
		totalContribution = totalContribution.Add(c)
	}

	if totalContribution.IsZero() {
		return nil
	}

	var rewards []Reward
	for oracleID, contribution := range contributions {
		// Proportional reward
		proportion := contribution.Div(totalContribution)
		reward := rewardPool.Mul(proportion)

		rewards = append(rewards, Reward{
			OracleID:  oracleID,
			Amount:    reward,
			Epoch:     epoch,
			Reason:    "epoch_reward",
			Timestamp: time.Now(),
		})

		d.logger.Info("reward distributed",
			"oracle_id", oracleID,
			"amount", reward.String(),
			"epoch", epoch,
		)
	}

	return rewards
}
