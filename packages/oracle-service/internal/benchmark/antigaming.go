package benchmark

import (
	"fmt"
	"math"
	"time"
)

// AntiGamingChecker detects attempts to game benchmark results.
type AntiGamingChecker struct {
	// ZScoreThreshold is the maximum allowed z-score deviation.
	ZScoreThreshold float64
	// MinBenchmarkDuration is the minimum acceptable benchmark duration.
	MinBenchmarkDuration time.Duration
}

// NewAntiGamingChecker creates a new AntiGamingChecker with default settings.
func NewAntiGamingChecker() *AntiGamingChecker {
	return &AntiGamingChecker{
		ZScoreThreshold:      2.0,
		MinBenchmarkDuration: 100 * time.Millisecond,
	}
}

// ValidateResult checks a CU report for gaming indicators.
func (c *AntiGamingChecker) ValidateResult(report *CUReport, historicalScores []float64) error {
	// Check for pre-computed results (suspiciously fast execution)
	if report.Duration < c.MinBenchmarkDuration {
		return fmt.Errorf("benchmark completed too quickly (%v), possible pre-computed results", report.Duration)
	}

	// Statistical outlier detection against historical scores
	if len(historicalScores) >= 3 {
		if err := c.checkOutlier(report.CalibratedCU, historicalScores); err != nil {
			return err
		}
	}

	// Check for impossible scores (exceeding theoretical maximum)
	if report.MatMulScore > 5000 || report.InferenceScore > 5000 || report.BandwidthScore > 5000 {
		return fmt.Errorf("benchmark scores exceed theoretical maximum")
	}

	return nil
}

// checkOutlier checks if a value is a statistical outlier using z-score.
func (c *AntiGamingChecker) checkOutlier(value float64, historical []float64) error {
	if len(historical) == 0 {
		return nil
	}

	mean, stddev := meanAndStdDev(historical)
	if stddev == 0 {
		if value != mean {
			return fmt.Errorf("score %.2f deviates from constant historical value %.2f", value, mean)
		}
		return nil
	}

	zScore := math.Abs(value-mean) / stddev
	if zScore > c.ZScoreThreshold {
		return fmt.Errorf("score %.2f has z-score %.2f (threshold: %.2f), detected as outlier",
			value, zScore, c.ZScoreThreshold)
	}

	return nil
}

// meanAndStdDev computes mean and standard deviation.
func meanAndStdDev(values []float64) (float64, float64) {
	if len(values) == 0 {
		return 0, 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	sumSqDiff := 0.0
	for _, v := range values {
		diff := v - mean
		sumSqDiff += diff * diff
	}
	stddev := math.Sqrt(sumSqDiff / float64(len(values)))

	return mean, stddev
}
