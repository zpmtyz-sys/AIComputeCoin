package hardware

import "time"

// HardwareReport contains GPU hardware fingerprint and attestation data.
type HardwareReport struct {
	GPUModel          string
	Serial            string
	PCISlot           string
	DriverVersion     string
	VRAM              uint64 // in MB
	ComputeCapability string
	Fingerprint       string
	AttestationReport []byte
	Timestamp         time.Time
}

// GPUInfo holds raw GPU device information from NVML.
type GPUInfo struct {
	Model             string
	Serial            string
	PCISlot           string
	DriverVersion     string
	VRAM              uint64
	ComputeCapability string
}

// GPUProfile defines expected specifications for a known GPU model.
type GPUProfile struct {
	Model             string
	MinVRAM           uint64
	MaxVRAM           uint64
	DriverPrefix      string
	ComputeCapability string
}

// KnownGPUProfiles contains expected specs for supported GPU models.
var KnownGPUProfiles = map[string]GPUProfile{
	"NVIDIA A100": {
		Model:             "NVIDIA A100",
		MinVRAM:           40960,
		MaxVRAM:           81920,
		DriverPrefix:      "535.",
		ComputeCapability: "8.0",
	},
	"NVIDIA H100": {
		Model:             "NVIDIA H100",
		MinVRAM:           81920,
		MaxVRAM:           81920,
		DriverPrefix:      "535.",
		ComputeCapability: "9.0",
	},
	"NVIDIA RTX 4090": {
		Model:             "NVIDIA RTX 4090",
		MinVRAM:           24576,
		MaxVRAM:           24576,
		DriverPrefix:      "535.",
		ComputeCapability: "8.9",
	},
}
