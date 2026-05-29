package benchmark

import "time"

// BenchConfig holds benchmark execution parameters.
type BenchConfig struct {
	MatrixSize     int
	SequenceLength int
	RandomSeed     int64
}

// DefaultBenchConfig returns the default benchmark configuration.
func DefaultBenchConfig() BenchConfig {
	return BenchConfig{
		MatrixSize:     4096,
		SequenceLength: 512,
		RandomSeed:     42,
	}
}

// CUReport contains the results of a compute unit benchmark.
type CUReport struct {
	MatMulScore    float64
	InferenceScore float64
	BandwidthScore float64
	RawCU          float64
	CalibratedCU   float64
	Timestamp      time.Time
	Duration       time.Duration
}

// GPUCapability provides simulated GPU performance characteristics.
type GPUCapability struct {
	Model         string
	PeakTFLOPS    float64 // FP32 peak TFLOPS
	MemBandwidth  float64 // GB/s
	TensorCoreMul float64 // Tensor core multiplier
}

// KnownGPUCapabilities provides performance baselines for known GPUs.
var KnownGPUCapabilities = map[string]GPUCapability{
	"NVIDIA A100": {
		Model:         "NVIDIA A100",
		PeakTFLOPS:    19.5,
		MemBandwidth:  2039.0,
		TensorCoreMul: 16.0,
	},
	"NVIDIA H100": {
		Model:         "NVIDIA H100",
		PeakTFLOPS:    51.2,
		MemBandwidth:  3350.0,
		TensorCoreMul: 32.0,
	},
	"NVIDIA RTX 4090": {
		Model:         "NVIDIA RTX 4090",
		PeakTFLOPS:    82.6,
		MemBandwidth:  1008.0,
		TensorCoreMul: 8.0,
	},
}
