package hardware

import (
	"crypto/sha256"
	"encoding/binary"
	"time"
)

// TEEAttestor generates mock SGX/SEV attestation quotes.
type TEEAttestor struct{}

// NewTEEAttestor creates a new TEEAttestor.
func NewTEEAttestor() *TEEAttestor {
	return &TEEAttestor{}
}

// GenerateAttestation creates a mock TEE attestation report for the given GPU info.
// In production, this would interact with actual SGX/SEV hardware.
func (t *TEEAttestor) GenerateAttestation(info *GPUInfo) []byte {
	// Construct mock attestation quote
	// Format: [version(2) | timestamp(8) | gpu_hash(32) | mock_signature(64)]
	quote := make([]byte, 106)

	// Version
	binary.BigEndian.PutUint16(quote[0:2], 1)

	// Timestamp
	binary.BigEndian.PutUint64(quote[2:10], uint64(time.Now().UnixNano()))

	// GPU info hash
	gpuData := []byte(info.Model + info.Serial + info.PCISlot)
	gpuHash := sha256.Sum256(gpuData)
	copy(quote[10:42], gpuHash[:])

	// Mock signature (in production this would be an SGX/SEV signature)
	sigData := append(quote[0:42], []byte("mock-tee-signing-key")...)
	sig := sha256.Sum256(sigData)
	copy(quote[42:74], sig[:])
	sig2 := sha256.Sum256(sig[:])
	copy(quote[74:106], sig2[:])

	return quote
}

// VerifyAttestation verifies a TEE attestation report.
// In production, this would verify against Intel/AMD attestation services.
func (t *TEEAttestor) VerifyAttestation(attestation []byte) bool {
	if len(attestation) < 106 {
		return false
	}

	// Check version
	version := binary.BigEndian.Uint16(attestation[0:2])
	if version != 1 {
		return false
	}

	// Check timestamp is not too old (within 24 hours)
	ts := binary.BigEndian.Uint64(attestation[2:10])
	reportTime := time.Unix(0, int64(ts))
	if time.Since(reportTime) > 24*time.Hour {
		return false
	}

	return true
}
