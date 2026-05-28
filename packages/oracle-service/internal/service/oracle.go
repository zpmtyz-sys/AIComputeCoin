package service

import (
	"log"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/oracle-service/internal/config"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/oracle-service/internal/verifier"
)

// OracleService combines hardware, performance, and economic verification.
type OracleService struct {
	config      *config.Config
	hardware    verifier.HardwareVerifier
	performance verifier.PerformanceVerifier
	economic    verifier.EconomicVerifier
}

// NewOracleService creates a new OracleService with all verifiers.
func NewOracleService(cfg *config.Config) *OracleService {
	return &OracleService{
		config:      cfg,
		hardware:    verifier.NewHardwareVerifier(),
		performance: verifier.NewPerformanceVerifier(),
		economic:    verifier.NewEconomicVerifier(cfg.StakeAmount),
	}
}

// MeasureAndReport performs a complete measurement cycle and reports results.
func (s *OracleService) MeasureAndReport(nodeID string) error {
	// Step 1: Verify hardware
	fingerprint, err := s.hardware.GetGPUFingerprint()
	if err != nil {
		return err
	}
	log.Printf("Hardware fingerprint: %s", fingerprint)

	// Step 2: Measure performance
	if err := s.performance.RunBenchmark(); err != nil {
		return err
	}
	tflops, err := s.performance.MeasureCU()
	if err != nil {
		return err
	}
	log.Printf("Measured TFLOPS: %.2f", tflops)

	// Step 3: Economic verification
	stakeOk, err := s.economic.CheckStake(nodeID)
	if err != nil {
		return err
	}
	if !stakeOk {
		log.Printf("Insufficient stake for node %s", nodeID)
		return nil
	}

	confidence, err := s.economic.CalculateConfidence(tflops, fingerprint)
	if err != nil {
		return err
	}
	log.Printf("Confidence score: %.4f", confidence)

	// Step 4: Submit report (stub - would submit to blockchain)
	log.Printf("Report submitted for node %s: %.2f TFLOPS, confidence %.4f", nodeID, tflops, confidence)
	return nil
}
