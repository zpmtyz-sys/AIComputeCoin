package hardware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"
)

// Fingerprinter generates hardware fingerprints from GPU metadata.
type Fingerprinter interface {
	// Fingerprint generates a hardware report with a unique fingerprint.
	Fingerprint(ctx context.Context) (*HardwareReport, error)
}

// DefaultFingerprinter implements Fingerprinter using an NVMLClient.
type DefaultFingerprinter struct {
	nvml   NVMLClient
	tee    *TEEAttestor
	spoof  *AntiSpoofChecker
	logger *slog.Logger
}

// NewFingerprinter creates a new DefaultFingerprinter.
func NewFingerprinter(nvml NVMLClient, logger *slog.Logger) *DefaultFingerprinter {
	if logger == nil {
		logger = slog.Default()
	}
	return &DefaultFingerprinter{
		nvml:   nvml,
		tee:    NewTEEAttestor(),
		spoof:  NewAntiSpoofChecker(),
		logger: logger,
	}
}

// Fingerprint generates a hardware report with SHA-256 fingerprint.
func (f *DefaultFingerprinter) Fingerprint(ctx context.Context) (*HardwareReport, error) {
	info, err := f.nvml.GetGPUInfo()
	if err != nil {
		return nil, fmt.Errorf("failed to get GPU info: %w", err)
	}

	f.logger.Info("generating hardware fingerprint",
		"gpu_model", info.Model,
		"serial", info.Serial,
	)

	// Anti-spoofing validation
	if err := f.spoof.Validate(info); err != nil {
		f.logger.Warn("anti-spoofing check failed", "error", err)
		return nil, fmt.Errorf("anti-spoofing validation failed: %w", err)
	}

	// Generate fingerprint from key hardware attributes
	fingerprint := generateFingerprint(info)

	// Generate TEE attestation
	attestation := f.tee.GenerateAttestation(info)

	report := &HardwareReport{
		GPUModel:          info.Model,
		Serial:            info.Serial,
		PCISlot:           info.PCISlot,
		DriverVersion:     info.DriverVersion,
		VRAM:              info.VRAM,
		ComputeCapability: info.ComputeCapability,
		Fingerprint:       fingerprint,
		AttestationReport: attestation,
		Timestamp:         time.Now(),
	}

	f.logger.Info("hardware fingerprint generated",
		"fingerprint", fingerprint,
		"gpu_model", info.Model,
	)

	return report, nil
}

// generateFingerprint creates a SHA-256 hash from GPU attributes.
func generateFingerprint(info *GPUInfo) string {
	data := fmt.Sprintf("%s|%s|%s|%s", info.Model, info.Serial, info.PCISlot, info.DriverVersion)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}
