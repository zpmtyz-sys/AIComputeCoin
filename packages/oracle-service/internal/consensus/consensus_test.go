package consensus

import (
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeMeasurement(oracleID, nodeID string, cu float64) Measurement {
	return Measurement{
		OracleID:  oracleID,
		NodeID:    nodeID,
		CUValue:   cu,
		Timestamp: time.Now(),
		Signature: []byte("test-sig-" + oracleID),
	}
}

func TestConsensus_Successful(t *testing.T) {
	tests := []struct {
		name         string
		measurements []Measurement
		expectedCU   float64
	}{
		{
			name: "3 honest oracles with similar values",
			measurements: []Measurement{
				makeMeasurement("oracle-1", "node-1", 1000),
				makeMeasurement("oracle-2", "node-1", 1005),
				makeMeasurement("oracle-3", "node-1", 998),
			},
			expectedCU: 1000, // median of [998, 1000, 1005]
		},
		{
			name: "5 honest oracles",
			measurements: []Measurement{
				makeMeasurement("oracle-1", "node-1", 1000),
				makeMeasurement("oracle-2", "node-1", 1010),
				makeMeasurement("oracle-3", "node-1", 990),
				makeMeasurement("oracle-4", "node-1", 1005),
				makeMeasurement("oracle-5", "node-1", 995),
			},
			expectedCU: 1000, // median of [990, 995, 1000, 1005, 1010]
		},
		{
			name: "4 oracles, even count median",
			measurements: []Measurement{
				makeMeasurement("oracle-1", "node-1", 900),
				makeMeasurement("oracle-2", "node-1", 1000),
				makeMeasurement("oracle-3", "node-1", 1100),
				makeMeasurement("oracle-4", "node-1", 1050),
			},
			expectedCU: 1025, // median of [900, 1000, 1050, 1100] = (1000+1050)/2
		},
	}

	agg := NewAggregator(slog.Default())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := agg.Aggregate(tt.measurements)
			require.NoError(t, err)
			assert.InDelta(t, tt.expectedCU, result.FinalCU, 0.01)
			assert.Equal(t, "node-1", result.NodeID)
			assert.Greater(t, result.Confidence, 0.0)
			assert.NotEmpty(t, result.OracleSignatures)
		})
	}
}

func TestConsensus_ByzantineOutlierRemoved(t *testing.T) {
	agg := NewAggregator(slog.Default())

	// 4 oracles: 3 honest (near 1000), 1 byzantine (reports 5000)
	// With n=4, max byzantine = (4-1)/3 = 1
	measurements := []Measurement{
		makeMeasurement("oracle-1", "node-1", 1000),
		makeMeasurement("oracle-2", "node-1", 1005),
		makeMeasurement("oracle-3", "node-1", 998),
		makeMeasurement("byzantine", "node-1", 5000), // Outlier
	}

	result, err := agg.Aggregate(measurements)
	require.NoError(t, err)

	// The byzantine oracle should be removed, median of remaining [998, 1000, 1005] = 1000
	assert.InDelta(t, 1000, result.FinalCU, 1.0)

	// Byzantine oracle should not be in signatures
	for _, sig := range result.OracleSignatures {
		assert.NotEqual(t, "byzantine", sig.OracleID)
	}
}

func TestConsensus_InsufficientOracles(t *testing.T) {
	agg := NewAggregator(slog.Default())

	tests := []struct {
		name         string
		measurements []Measurement
	}{
		{
			name: "only 1 oracle",
			measurements: []Measurement{
				makeMeasurement("oracle-1", "node-1", 1000),
			},
		},
		{
			name: "only 2 oracles",
			measurements: []Measurement{
				makeMeasurement("oracle-1", "node-1", 1000),
				makeMeasurement("oracle-2", "node-1", 1005),
			},
		},
		{
			name:         "zero oracles",
			measurements: []Measurement{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := agg.Aggregate(tt.measurements)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "insufficient oracles")
		})
	}
}

func TestConsensus_TooManyByzantines(t *testing.T) {
	agg := NewAggregator(slog.Default())

	// 3 oracles: max byzantine tolerance = (3-1)/3 = 0
	// If even one is removed as outlier, it exceeds tolerance
	// Two close values and one extreme outlier: the outlier gets removed,
	// but removing 1 from 3 exceeds max_byzantine=0
	measurements := []Measurement{
		makeMeasurement("honest-1", "node-1", 1000),
		makeMeasurement("honest-2", "node-1", 1002),
		makeMeasurement("byzantine", "node-1", 50000),
	}

	_, err := agg.Aggregate(measurements)
	assert.Error(t, err)
}

func TestConsensus_OutlierRemoval(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "no outliers in tight cluster",
			values:   []float64{100, 101, 99, 100, 102},
			expected: 100, // All within 2 sigma
		},
	}

	agg := NewAggregator(slog.Default())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			measurements := make([]Measurement, len(tt.values))
			for i, v := range tt.values {
				measurements[i] = makeMeasurement(
					fmt.Sprintf("oracle-%d", i),
					"node-1", v,
				)
			}

			result, err := agg.Aggregate(measurements)
			require.NoError(t, err)
			assert.InDelta(t, tt.expected, result.FinalCU, 1.0)
		})
	}
}

func TestDispute_Escalation(t *testing.T) {
	resolver := NewDisputeResolver()

	measurements := []Measurement{
		makeMeasurement("oracle-1", "node-1", 1000),
		makeMeasurement("oracle-2", "node-1", 5000),
	}

	escalation := resolver.Escalate(measurements, "consensus failed")
	assert.Equal(t, "node-1", escalation.NodeID)
	assert.Equal(t, "consensus failed", escalation.Reason)
	assert.Len(t, escalation.Measurements, 2)
}

func TestMedianComputation(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{
			name:     "odd count",
			values:   []float64{1, 3, 5},
			expected: 3,
		},
		{
			name:     "even count",
			values:   []float64{1, 3, 5, 7},
			expected: 4, // (3+5)/2
		},
		{
			name:     "single value",
			values:   []float64{42},
			expected: 42,
		},
		{
			name:     "two values",
			values:   []float64{10, 20},
			expected: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeMedianValues(tt.values)
			assert.InDelta(t, tt.expected, result, 0.001)
		})
	}
}

func TestStdDevComputation(t *testing.T) {
	tests := []struct {
		name    string
		values  []float64
		center  float64
		wantGT0 bool
	}{
		{
			name:    "identical values",
			values:  []float64{5, 5, 5},
			center:  5,
			wantGT0: false,
		},
		{
			name:    "spread values",
			values:  []float64{1, 2, 3, 4, 5},
			center:  3,
			wantGT0: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sd := computeStdDev(tt.values, tt.center)
			if tt.wantGT0 {
				assert.Greater(t, sd, 0.0)
			} else {
				assert.InDelta(t, 0, sd, 0.001)
			}
		})
	}
}
