package hardware

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFingerprint_Determinism(t *testing.T) {
	tests := []struct {
		name    string
		profile string
	}{
		{"A100 produces consistent fingerprint", "A100"},
		{"H100 produces consistent fingerprint", "H100"},
		{"RTX4090 produces consistent fingerprint", "RTX4090"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nvml := NewMockNVMLClient(tt.profile)
			fp := NewFingerprinter(nvml, slog.Default())

			report1, err := fp.Fingerprint(context.Background())
			require.NoError(t, err)

			report2, err := fp.Fingerprint(context.Background())
			require.NoError(t, err)

			assert.Equal(t, report1.Fingerprint, report2.Fingerprint)
			assert.Len(t, report1.Fingerprint, 64) // SHA-256 hex
		})
	}
}

func TestFingerprint_DifferentGPUs(t *testing.T) {
	profiles := []string{"A100", "H100", "RTX4090"}
	fingerprints := make(map[string]string)

	for _, profile := range profiles {
		nvml := NewMockNVMLClient(profile)
		fp := NewFingerprinter(nvml, slog.Default())

		report, err := fp.Fingerprint(context.Background())
		require.NoError(t, err)
		fingerprints[profile] = report.Fingerprint
	}

	// All fingerprints should be unique
	assert.NotEqual(t, fingerprints["A100"], fingerprints["H100"])
	assert.NotEqual(t, fingerprints["A100"], fingerprints["RTX4090"])
	assert.NotEqual(t, fingerprints["H100"], fingerprints["RTX4090"])
}

func TestAntiSpoofChecker_ValidGPU(t *testing.T) {
	tests := []struct {
		name    string
		info    *GPUInfo
		wantErr bool
	}{
		{
			name: "valid A100",
			info: &GPUInfo{
				Model:         "NVIDIA A100",
				Serial:        "GPU-a100-001",
				PCISlot:       "0000:3b:00.0",
				DriverVersion: "535.129.03",
				VRAM:          81920,
			},
			wantErr: false,
		},
		{
			name: "valid H100",
			info: &GPUInfo{
				Model:         "NVIDIA H100",
				Serial:        "GPU-h100-001",
				PCISlot:       "0000:4a:00.0",
				DriverVersion: "535.154.05",
				VRAM:          81920,
			},
			wantErr: false,
		},
		{
			name: "valid RTX 4090",
			info: &GPUInfo{
				Model:         "NVIDIA RTX 4090",
				Serial:        "GPU-rtx4090-001",
				PCISlot:       "0000:01:00.0",
				DriverVersion: "535.183.01",
				VRAM:          24576,
			},
			wantErr: false,
		},
	}

	checker := NewAntiSpoofChecker()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checker.Validate(tt.info)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAntiSpoofChecker_DetectsSpoofing(t *testing.T) {
	tests := []struct {
		name       string
		info       *GPUInfo
		wantErrMsg string
	}{
		{
			name: "mismatched driver version",
			info: &GPUInfo{
				Model:         "NVIDIA A100",
				Serial:        "GPU-a100-001",
				PCISlot:       "0000:3b:00.0",
				DriverVersion: "450.51.06",
				VRAM:          81920,
			},
			wantErrMsg: "driver version",
		},
		{
			name: "mismatched VRAM for A100",
			info: &GPUInfo{
				Model:         "NVIDIA A100",
				Serial:        "GPU-a100-001",
				PCISlot:       "0000:3b:00.0",
				DriverVersion: "535.129.03",
				VRAM:          8192, // Too low for A100
			},
			wantErrMsg: "VRAM",
		},
		{
			name: "invalid PCI slot format",
			info: &GPUInfo{
				Model:         "NVIDIA A100",
				Serial:        "GPU-a100-001",
				PCISlot:       "invalid-pci",
				DriverVersion: "535.129.03",
				VRAM:          81920,
			},
			wantErrMsg: "invalid PCI",
		},
		{
			name: "PCI slot missing function",
			info: &GPUInfo{
				Model:         "NVIDIA A100",
				Serial:        "GPU-a100-001",
				PCISlot:       "0000:3b:00",
				DriverVersion: "535.129.03",
				VRAM:          81920,
			},
			wantErrMsg: "device.function",
		},
	}

	checker := NewAntiSpoofChecker()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checker.Validate(tt.info)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErrMsg)
		})
	}
}

func TestTEEAttestation(t *testing.T) {
	tests := []struct {
		name    string
		profile string
	}{
		{"A100 attestation", "A100"},
		{"H100 attestation", "H100"},
	}

	attestor := NewTEEAttestor()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nvml := NewMockNVMLClient(tt.profile)
			info, err := nvml.GetGPUInfo()
			require.NoError(t, err)

			attestation := attestor.GenerateAttestation(info)
			assert.Len(t, attestation, 106)

			valid := attestor.VerifyAttestation(attestation)
			assert.True(t, valid)
		})
	}
}

func TestTEEAttestation_InvalidData(t *testing.T) {
	attestor := NewTEEAttestor()

	tests := []struct {
		name        string
		attestation []byte
		wantValid   bool
	}{
		{
			name:        "too short",
			attestation: []byte{0x00, 0x01},
			wantValid:   false,
		},
		{
			name:        "empty",
			attestation: []byte{},
			wantValid:   false,
		},
		{
			name:        "wrong version",
			attestation: make([]byte, 106), // version 0
			wantValid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := attestor.VerifyAttestation(tt.attestation)
			assert.Equal(t, tt.wantValid, valid)
		})
	}
}

func TestMockNVMLClient_Profiles(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		model   string
		vram    uint64
	}{
		{"A100 profile", "A100", "NVIDIA A100", 81920},
		{"H100 profile", "H100", "NVIDIA H100", 81920},
		{"RTX4090 profile", "RTX4090", "NVIDIA RTX 4090", 24576},
		{"default profile", "unknown", "NVIDIA A100", 81920},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewMockNVMLClient(tt.profile)
			info, err := client.GetGPUInfo()
			require.NoError(t, err)
			assert.Equal(t, tt.model, info.Model)
			assert.Equal(t, tt.vram, info.VRAM)
		})
	}
}
