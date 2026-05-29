package consensus

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"
)

const (
	// MinOracles is the minimum number of oracles required for consensus.
	MinOracles = 3
	// OutlierSigmaThreshold is the z-score threshold for outlier removal.
	OutlierSigmaThreshold = 2.0
)

// Aggregator implements BFT consensus for oracle measurements.
type Aggregator struct {
	logger *slog.Logger
}

// NewAggregator creates a new consensus Aggregator.
func NewAggregator(logger *slog.Logger) *Aggregator {
	if logger == nil {
		logger = slog.Default()
	}
	return &Aggregator{logger: logger}
}

// Aggregate collects measurements and reaches consensus using BFT protocol.
// Requires minimum 3 oracles. Tolerates f < n/3 byzantine oracles.
func (a *Aggregator) Aggregate(measurements []Measurement) (*ConsensusResult, error) {
	if len(measurements) < MinOracles {
		return nil, fmt.Errorf("insufficient oracles: got %d, need at least %d",
			len(measurements), MinOracles)
	}

	// Check if we have enough honest oracles for BFT (n >= 3f + 1)
	n := len(measurements)
	maxByzantine := (n - 1) / 3

	a.logger.Info("starting consensus",
		"total_oracles", n,
		"max_byzantine", maxByzantine,
		"node_id", measurements[0].NodeID,
	)

	// Remove statistical outliers (>2 sigma from median)
	filtered := a.removeOutliers(measurements)

	// Check if too many were removed (byzantine > n/3)
	removed := n - len(filtered)
	if removed > maxByzantine {
		return nil, fmt.Errorf("too many outliers (%d/%d), exceeds byzantine tolerance (%d)",
			removed, n, maxByzantine)
	}

	// Need 2/3+1 threshold after filtering
	threshold := (2*n)/3 + 1
	if len(filtered) < threshold {
		return nil, fmt.Errorf("insufficient agreement: %d oracles remaining, need %d (2/3+1 of %d)",
			len(filtered), threshold, n)
	}

	// Final CU = median of remaining measurements
	finalCU := a.computeMedian(filtered)

	// Compute confidence based on agreement
	confidence := float64(len(filtered)) / float64(n)

	// Generate oracle signatures for the consensus result
	signatures := make([]OracleSignature, 0, len(filtered))
	for _, m := range filtered {
		sig := a.signResult(m.OracleID, measurements[0].NodeID, finalCU)
		signatures = append(signatures, OracleSignature{
			OracleID:  m.OracleID,
			Signature: sig,
		})
	}

	result := &ConsensusResult{
		NodeID:           measurements[0].NodeID,
		FinalCU:          finalCU,
		OracleSignatures: signatures,
		Timestamp:        time.Now(),
		Confidence:       confidence,
	}

	a.logger.Info("consensus reached",
		"node_id", result.NodeID,
		"final_cu", finalCU,
		"confidence", confidence,
		"signers", len(signatures),
	)

	return result, nil
}

// removeOutliers removes measurements that are >2 sigma from the median.
func (a *Aggregator) removeOutliers(measurements []Measurement) []Measurement {
	if len(measurements) <= 2 {
		return measurements
	}

	values := make([]float64, len(measurements))
	for i, m := range measurements {
		values[i] = m.CUValue
	}

	median := computeMedianValues(values)

	// Use median absolute deviation (MAD) for robust outlier detection
	absDeviations := make([]float64, len(values))
	for i, v := range values {
		absDeviations[i] = math.Abs(v - median)
	}
	sort.Float64s(absDeviations)
	var mad float64
	n := len(absDeviations)
	if n%2 == 0 {
		mad = (absDeviations[n/2-1] + absDeviations[n/2]) / 2
	} else {
		mad = absDeviations[n/2]
	}

	if mad == 0 {
		// If MAD is 0, use mean-based stddev as fallback
		stddev := computeStdDev(values, median)
		if stddev == 0 {
			return measurements
		}
		var filtered []Measurement
		for _, m := range measurements {
			zScore := math.Abs(m.CUValue-median) / stddev
			if zScore <= OutlierSigmaThreshold {
				filtered = append(filtered, m)
			} else {
				a.logger.Info("outlier removed",
					"oracle_id", m.OracleID,
					"cu_value", m.CUValue,
					"z_score", zScore,
				)
			}
		}
		return filtered
	}

	// Modified z-score using MAD: |value - median| / (1.4826 * MAD)
	// 1.4826 is the consistency constant for normal distribution
	scaledMAD := 1.4826 * mad

	var filtered []Measurement
	for _, m := range measurements {
		modifiedZ := math.Abs(m.CUValue-median) / scaledMAD
		if modifiedZ <= OutlierSigmaThreshold {
			filtered = append(filtered, m)
		} else {
			a.logger.Info("outlier removed",
				"oracle_id", m.OracleID,
				"cu_value", m.CUValue,
				"modified_z", modifiedZ,
			)
		}
	}

	return filtered
}

// computeMedian computes the median CU value from filtered measurements.
func (a *Aggregator) computeMedian(measurements []Measurement) float64 {
	values := make([]float64, len(measurements))
	for i, m := range measurements {
		values[i] = m.CUValue
	}
	return computeMedianValues(values)
}

// computeMedianValues computes the median of a slice of float64.
func computeMedianValues(values []float64) float64 {
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// computeStdDev computes standard deviation using the provided center point.
func computeStdDev(values []float64, center float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sumSqDiff := 0.0
	for _, v := range values {
		diff := v - center
		sumSqDiff += diff * diff
	}
	return math.Sqrt(sumSqDiff / float64(len(values)))
}

// signResult generates a mock signature for the consensus result.
func (a *Aggregator) signResult(oracleID, nodeID string, cu float64) []byte {
	data := fmt.Sprintf("%s|%s|%.6f", oracleID, nodeID, cu)
	cuBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(cuBytes, math.Float64bits(cu))
	data += string(cuBytes)
	hash := sha256.Sum256([]byte(data))
	return hash[:]
}
