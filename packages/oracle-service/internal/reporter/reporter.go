package reporter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/oracle-service/internal/consensus"
)

// ChainSubmitter is the interface for submitting reports to the blockchain.
type ChainSubmitter interface {
	Submit(ctx context.Context, report *OnChainReport) (string, error)
}

// Reporter submits consensus results to the blockchain.
type Reporter struct {
	submitter ChainSubmitter
	cache     *Cache
	retry     *RetryConfig
	logger    *slog.Logger
}

// NewReporter creates a new Reporter.
func NewReporter(submitter ChainSubmitter, cachePath string, logger *slog.Logger) *Reporter {
	if logger == nil {
		logger = slog.Default()
	}
	return &Reporter{
		submitter: submitter,
		cache:     NewCache(cachePath),
		retry:     DefaultRetryConfig(),
		logger:    logger,
	}
}

// SubmitReport constructs and submits an on-chain report from a consensus result.
func (r *Reporter) SubmitReport(ctx context.Context, result *consensus.ConsensusResult) (string, error) {
	report := r.constructReport(result)

	r.logger.Info("submitting report",
		"node_id", report.NodeID,
		"cu_value", report.CUValue,
	)

	// Cache before submission for crash recovery
	if err := r.cache.Store(report); err != nil {
		r.logger.Warn("failed to cache report", "error", err)
	}

	// Submit with retry
	txHash, err := r.retry.Execute(ctx, func() (string, error) {
		return r.submitter.Submit(ctx, report)
	})
	if err != nil {
		r.logger.Error("report submission failed", "error", err, "node_id", report.NodeID)
		return "", fmt.Errorf("failed to submit report for node %s: %w", report.NodeID, err)
	}

	report.TxHash = txHash

	// Remove from cache after successful submission
	r.cache.Remove(report.NodeID)

	r.logger.Info("report submitted successfully",
		"node_id", report.NodeID,
		"tx_hash", txHash,
	)

	return txHash, nil
}

// constructReport builds an OnChainReport from a ConsensusResult.
func (r *Reporter) constructReport(result *consensus.ConsensusResult) *OnChainReport {
	signatures := make([]SignatureEntry, 0, len(result.OracleSignatures))
	for _, sig := range result.OracleSignatures {
		signatures = append(signatures, SignatureEntry{
			OracleID:  sig.OracleID,
			Signature: sig.Signature,
		})
	}

	return &OnChainReport{
		NodeID:           result.NodeID,
		CUValue:          result.FinalCU,
		Timestamp:        result.Timestamp,
		OracleSignatures: signatures,
	}
}

// MockChainSubmitter is a test implementation of ChainSubmitter.
type MockChainSubmitter struct {
	FailCount int
	calls     int
}

// NewMockChainSubmitter creates a MockChainSubmitter that fails the first n attempts.
func NewMockChainSubmitter(failCount int) *MockChainSubmitter {
	return &MockChainSubmitter{FailCount: failCount}
}

// Submit simulates blockchain submission, failing the first FailCount times.
func (m *MockChainSubmitter) Submit(_ context.Context, report *OnChainReport) (string, error) {
	m.calls++
	if m.calls <= m.FailCount {
		return "", fmt.Errorf("simulated chain submission failure (attempt %d)", m.calls)
	}

	// Generate a mock tx hash
	data := fmt.Sprintf("%s|%.6f|%s", report.NodeID, report.CUValue, uuid.New().String())
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16]), nil
}

// Calls returns the total number of submit calls made.
func (m *MockChainSubmitter) Calls() int {
	return m.calls
}
