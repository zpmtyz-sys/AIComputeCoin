package hardware

// NVMLClient provides an interface for NVIDIA Management Library operations.
type NVMLClient interface {
	// GetGPUInfo retrieves GPU device information.
	GetGPUInfo() (*GPUInfo, error)
}

// MockNVMLClient provides realistic GPU data for testing without hardware.
type MockNVMLClient struct {
	Profile string
}

// NewMockNVMLClient creates a MockNVMLClient with the specified GPU profile.
func NewMockNVMLClient(profile string) *MockNVMLClient {
	return &MockNVMLClient{Profile: profile}
}

// GetGPUInfo returns mock GPU information based on the configured profile.
func (m *MockNVMLClient) GetGPUInfo() (*GPUInfo, error) {
	switch m.Profile {
	case "A100":
		return &GPUInfo{
			Model:             "NVIDIA A100",
			Serial:            "GPU-a100-serial-001",
			PCISlot:           "0000:3b:00.0",
			DriverVersion:     "535.129.03",
			VRAM:              81920,
			ComputeCapability: "8.0",
		}, nil
	case "H100":
		return &GPUInfo{
			Model:             "NVIDIA H100",
			Serial:            "GPU-h100-serial-001",
			PCISlot:           "0000:4a:00.0",
			DriverVersion:     "535.154.05",
			VRAM:              81920,
			ComputeCapability: "9.0",
		}, nil
	case "RTX4090":
		return &GPUInfo{
			Model:             "NVIDIA RTX 4090",
			Serial:            "GPU-rtx4090-serial-001",
			PCISlot:           "0000:01:00.0",
			DriverVersion:     "535.183.01",
			VRAM:              24576,
			ComputeCapability: "8.9",
		}, nil
	default:
		return &GPUInfo{
			Model:             "NVIDIA A100",
			Serial:            "GPU-default-serial-001",
			PCISlot:           "0000:3b:00.0",
			DriverVersion:     "535.129.03",
			VRAM:              81920,
			ComputeCapability: "8.0",
		}, nil
	}
}
