package benchmark

import (
	"fmt"
)

// InferenceBenchmark simulates a transformer inference benchmark.
type InferenceBenchmark struct {
	config BenchConfig
}

// NewInferenceBenchmark creates a new InferenceBenchmark.
func NewInferenceBenchmark(config BenchConfig) *InferenceBenchmark {
	return &InferenceBenchmark{config: config}
}

// Run executes the inference benchmark for the given GPU model.
// Simulates BERT-base equivalent forward pass timing.
// Returns a score normalized where A100 = 1000.
func (b *InferenceBenchmark) Run(gpuModel string) (float64, error) {
	cap, ok := KnownGPUCapabilities[gpuModel]
	if !ok {
		return 0, fmt.Errorf("unknown GPU model: %s", gpuModel)
	}

	// BERT-base parameters: 110M params, sequence length from config
	// Approximate FLOPS for forward pass: 2 * params * seq_length
	params := 110e6
	seqLen := float64(b.config.SequenceLength)
	totalFLOPS := 2 * params * seqLen

	// Inference heavily uses tensor cores
	effectiveTFLOPS := cap.PeakTFLOPS * cap.TensorCoreMul * 0.5 // 50% tensor core utilization

	// Simulated execution time
	execTime := totalFLOPS / (effectiveTFLOPS * 1e12)
	_ = execTime

	// Throughput score based on tensor core performance
	a100Effective := KnownGPUCapabilities["NVIDIA A100"].PeakTFLOPS *
		KnownGPUCapabilities["NVIDIA A100"].TensorCoreMul * 0.5
	score := (effectiveTFLOPS / a100Effective) * 1000.0

	return score, nil
}
