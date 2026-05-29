package benchmark

import (
	"fmt"
	"math"
)

// MatMulBenchmark simulates a matrix multiplication benchmark.
type MatMulBenchmark struct {
	config BenchConfig
}

// NewMatMulBenchmark creates a new MatMulBenchmark.
func NewMatMulBenchmark(config BenchConfig) *MatMulBenchmark {
	return &MatMulBenchmark{config: config}
}

// Run executes the matmul benchmark for the given GPU model.
// Returns a score normalized where A100 = 1000.
func (b *MatMulBenchmark) Run(gpuModel string) (float64, error) {
	cap, ok := KnownGPUCapabilities[gpuModel]
	if !ok {
		return 0, fmt.Errorf("unknown GPU model: %s", gpuModel)
	}

	// Simulate 4096x4096 FP32 GEMM
	// FLOPS = 2 * N^3 for matrix multiplication
	n := float64(b.config.MatrixSize)
	totalFLOPS := 2 * math.Pow(n, 3)

	// Simulated execution time based on GPU capability
	// Time = FLOPS / (peak_tflops * 1e12 * efficiency)
	efficiency := 0.75 // Typical achieved efficiency
	execTime := totalFLOPS / (cap.PeakTFLOPS * 1e12 * efficiency)

	// Achieved TFLOPS
	achievedTFLOPS := totalFLOPS / execTime / 1e12

	// Normalize: A100 baseline TFLOPS at 75% efficiency = 14.625
	a100Baseline := KnownGPUCapabilities["NVIDIA A100"].PeakTFLOPS * efficiency
	score := (achievedTFLOPS / a100Baseline) * 1000.0

	return score, nil
}
