package benchmark

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

const (
	// WeightMatMul is the weight for matrix multiplication score.
	WeightMatMul = 0.4
	// WeightInference is the weight for inference score.
	WeightInference = 0.4
	// WeightBandwidth is the weight for bandwidth score.
	WeightBandwidth = 0.2
	// A100BaselineCU is the calibration target for A100.
	A100BaselineCU = 1000.0
)

// Calculator computes calibrated compute units from benchmark results.
type Calculator interface {
	// RunBenchmarks executes all benchmarks and returns a CU report.
	RunBenchmarks(ctx context.Context, gpuModel string) (*CUReport, error)
}

// DefaultCalculator implements the Calculator interface.
type DefaultCalculator struct {
	config    BenchConfig
	matmul    *MatMulBenchmark
	inference *InferenceBenchmark
	bandwidth *BandwidthBenchmark
	logger    *slog.Logger
}

// NewCalculator creates a new DefaultCalculator.
func NewCalculator(config BenchConfig, logger *slog.Logger) *DefaultCalculator {
	if logger == nil {
		logger = slog.Default()
	}
	return &DefaultCalculator{
		config:    config,
		matmul:    NewMatMulBenchmark(config),
		inference: NewInferenceBenchmark(config),
		bandwidth: NewBandwidthBenchmark(config),
		logger:    logger,
	}
}

// RunBenchmarks executes all benchmarks and computes the calibrated CU value.
func (c *DefaultCalculator) RunBenchmarks(ctx context.Context, gpuModel string) (*CUReport, error) {
	start := time.Now()

	c.logger.Info("starting benchmarks", "gpu_model", gpuModel)

	matmulScore, err := c.matmul.Run(gpuModel)
	if err != nil {
		return nil, fmt.Errorf("matmul benchmark failed: %w", err)
	}

	inferenceScore, err := c.inference.Run(gpuModel)
	if err != nil {
		return nil, fmt.Errorf("inference benchmark failed: %w", err)
	}

	bandwidthScore, err := c.bandwidth.Run(gpuModel)
	if err != nil {
		return nil, fmt.Errorf("bandwidth benchmark failed: %w", err)
	}

	// Calculate weighted CU
	rawCU := matmulScore*WeightMatMul + inferenceScore*WeightInference + bandwidthScore*WeightBandwidth

	// Calibrate: A100 should produce exactly 1000 CU
	// Since each sub-benchmark normalizes A100=1000, the weighted sum for A100
	// should be 1000*0.4 + 1000*0.4 + 1000*0.2 = 1000
	calibratedCU := rawCU

	duration := time.Since(start)

	report := &CUReport{
		MatMulScore:    matmulScore,
		InferenceScore: inferenceScore,
		BandwidthScore: bandwidthScore,
		RawCU:          rawCU,
		CalibratedCU:   calibratedCU,
		Timestamp:      time.Now(),
		Duration:       duration,
	}

	c.logger.Info("benchmarks completed",
		"gpu_model", gpuModel,
		"matmul_score", matmulScore,
		"inference_score", inferenceScore,
		"bandwidth_score", bandwidthScore,
		"calibrated_cu", calibratedCU,
		"duration", duration,
	)

	return report, nil
}

// CalculateCU computes the weighted CU from individual scores.
func CalculateCU(matmulScore, inferenceScore, bandwidthScore float64) float64 {
	return matmulScore*WeightMatMul + inferenceScore*WeightInference + bandwidthScore*WeightBandwidth
}
