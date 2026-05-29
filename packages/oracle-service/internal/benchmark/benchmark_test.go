package benchmark

import (
	"context"
	"log/slog"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCUCalculation_Weights(t *testing.T) {
	tests := []struct {
		name       string
		matmul     float64
		inference  float64
		bandwidth  float64
		expectedCU float64
	}{
		{
			name:       "equal scores of 1000",
			matmul:     1000,
			inference:  1000,
			bandwidth:  1000,
			expectedCU: 1000,
		},
		{
			name:       "only matmul",
			matmul:     1000,
			inference:  0,
			bandwidth:  0,
			expectedCU: 400,
		},
		{
			name:       "only inference",
			matmul:     0,
			inference:  1000,
			bandwidth:  0,
			expectedCU: 400,
		},
		{
			name:       "only bandwidth",
			matmul:     0,
			inference:  0,
			bandwidth:  1000,
			expectedCU: 200,
		},
		{
			name:       "mixed scores",
			matmul:     2000,
			inference:  1500,
			bandwidth:  800,
			expectedCU: 2000*0.4 + 1500*0.4 + 800*0.2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cu := CalculateCU(tt.matmul, tt.inference, tt.bandwidth)
			assert.InDelta(t, tt.expectedCU, cu, 0.001)
		})
	}
}

func TestA100Calibration(t *testing.T) {
	calc := NewCalculator(DefaultBenchConfig(), slog.Default())

	report, err := calc.RunBenchmarks(context.Background(), "NVIDIA A100")
	require.NoError(t, err)

	// A100 should produce exactly 1000 CU
	assert.InDelta(t, A100BaselineCU, report.CalibratedCU, 0.01)
	assert.InDelta(t, 1000.0, report.MatMulScore, 0.01)
	assert.InDelta(t, 1000.0, report.InferenceScore, 0.01)
	assert.InDelta(t, 1000.0, report.BandwidthScore, 0.01)
}

func TestDifferentGPUTiers(t *testing.T) {
	calc := NewCalculator(DefaultBenchConfig(), slog.Default())

	tests := []struct {
		name     string
		gpuModel string
		minCU    float64
		maxCU    float64
	}{
		{
			name:     "A100 baseline",
			gpuModel: "NVIDIA A100",
			minCU:    999,
			maxCU:    1001,
		},
		{
			name:     "H100 higher than A100",
			gpuModel: "NVIDIA H100",
			minCU:    1500,
			maxCU:    5000,
		},
		{
			name:     "RTX 4090 different from A100",
			gpuModel: "NVIDIA RTX 4090",
			minCU:    500,
			maxCU:    5000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := calc.RunBenchmarks(context.Background(), tt.gpuModel)
			require.NoError(t, err)
			assert.Greater(t, report.CalibratedCU, tt.minCU)
			assert.Less(t, report.CalibratedCU, tt.maxCU)
		})
	}
}

func TestConsistentResults(t *testing.T) {
	calc := NewCalculator(DefaultBenchConfig(), slog.Default())

	report1, err := calc.RunBenchmarks(context.Background(), "NVIDIA A100")
	require.NoError(t, err)

	report2, err := calc.RunBenchmarks(context.Background(), "NVIDIA A100")
	require.NoError(t, err)

	assert.InDelta(t, report1.CalibratedCU, report2.CalibratedCU, 0.001)
}

func TestUnknownGPU(t *testing.T) {
	calc := NewCalculator(DefaultBenchConfig(), slog.Default())

	_, err := calc.RunBenchmarks(context.Background(), "Unknown GPU")
	assert.Error(t, err)
}

func TestAntiGaming_OutlierDetection(t *testing.T) {
	checker := NewAntiGamingChecker()

	tests := []struct {
		name       string
		report     *CUReport
		historical []float64
		wantErr    bool
	}{
		{
			name: "normal result within historical range",
			report: &CUReport{
				CalibratedCU: 1000,
				Duration:     5 * time.Second,
			},
			historical: []float64{990, 1010, 995, 1005, 998},
			wantErr:    false,
		},
		{
			name: "outlier result far from historical",
			report: &CUReport{
				CalibratedCU: 5000,
				Duration:     5 * time.Second,
			},
			historical: []float64{990, 1010, 995, 1005, 998},
			wantErr:    true,
		},
		{
			name: "too fast execution",
			report: &CUReport{
				CalibratedCU: 1000,
				Duration:     10 * time.Millisecond,
			},
			historical: []float64{990, 1010, 995},
			wantErr:    true,
		},
		{
			name: "impossible scores",
			report: &CUReport{
				MatMulScore:  6000,
				CalibratedCU: 1000,
				Duration:     5 * time.Second,
			},
			historical: nil,
			wantErr:    true,
		},
		{
			name: "no historical data, normal result",
			report: &CUReport{
				CalibratedCU: 1000,
				Duration:     5 * time.Second,
			},
			historical: nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checker.ValidateResult(tt.report, tt.historical)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMeanAndStdDev(t *testing.T) {
	tests := []struct {
		name       string
		values     []float64
		wantMean   float64
		wantStdDev float64
	}{
		{
			name:       "simple values",
			values:     []float64{1, 2, 3, 4, 5},
			wantMean:   3.0,
			wantStdDev: math.Sqrt(2.0),
		},
		{
			name:       "identical values",
			values:     []float64{5, 5, 5},
			wantMean:   5.0,
			wantStdDev: 0.0,
		},
		{
			name:       "empty",
			values:     []float64{},
			wantMean:   0,
			wantStdDev: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mean, stddev := meanAndStdDev(tt.values)
			assert.InDelta(t, tt.wantMean, mean, 0.001)
			assert.InDelta(t, tt.wantStdDev, stddev, 0.001)
		})
	}
}
