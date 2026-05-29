package benchmark

import (
	"fmt"
)

// BandwidthBenchmark simulates memory bandwidth testing.
type BandwidthBenchmark struct {
	config BenchConfig
}

// NewBandwidthBenchmark creates a new BandwidthBenchmark.
func NewBandwidthBenchmark(config BenchConfig) *BandwidthBenchmark {
	return &BandwidthBenchmark{config: config}
}

// Run executes the bandwidth benchmark for the given GPU model.
// Tests sequential and random access patterns.
// Returns a score normalized where A100 = 1000.
func (b *BandwidthBenchmark) Run(gpuModel string) (float64, error) {
	cap, ok := KnownGPUCapabilities[gpuModel]
	if !ok {
		return 0, fmt.Errorf("unknown GPU model: %s", gpuModel)
	}

	// Sequential bandwidth test: large contiguous memory transfers
	seqEfficiency := 0.85 // Sequential access achieves ~85% of theoretical peak
	seqBandwidth := cap.MemBandwidth * seqEfficiency

	// Random access: significantly lower efficiency
	randEfficiency := 0.35
	randBandwidth := cap.MemBandwidth * randEfficiency

	// Combined score: 70% sequential, 30% random (weighted by typical workload patterns)
	combinedBandwidth := seqBandwidth*0.7 + randBandwidth*0.3

	// Normalize against A100
	a100Cap := KnownGPUCapabilities["NVIDIA A100"]
	a100Combined := a100Cap.MemBandwidth*seqEfficiency*0.7 + a100Cap.MemBandwidth*randEfficiency*0.3

	score := (combinedBandwidth / a100Combined) * 1000.0

	return score, nil
}
