package reporter

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/oracle-service/internal/consensus"
)

// BatchReporter aggregates multiple consensus results for gas-efficient submission.
type BatchReporter struct {
	reporter  *Reporter
	batchSize int
	pending   []*consensus.ConsensusResult
	logger    *slog.Logger
}

// NewBatchReporter creates a new BatchReporter with the specified batch size.
func NewBatchReporter(reporter *Reporter, batchSize int, logger *slog.Logger) *BatchReporter {
	if logger == nil {
		logger = slog.Default()
	}
	if batchSize < 1 {
		batchSize = 10
	}
	return &BatchReporter{
		reporter:  reporter,
		batchSize: batchSize,
		pending:   make([]*consensus.ConsensusResult, 0, batchSize),
		logger:    logger,
	}
}

// Add adds a consensus result to the batch.
// Returns the batch report if the batch is full, nil otherwise.
func (b *BatchReporter) Add(ctx context.Context, result *consensus.ConsensusResult) (*BatchReport, error) {
	b.pending = append(b.pending, result)

	if len(b.pending) >= b.batchSize {
		return b.Flush(ctx)
	}

	return nil, nil
}

// Flush submits all pending results as a batch.
func (b *BatchReporter) Flush(ctx context.Context) (*BatchReport, error) {
	if len(b.pending) == 0 {
		return nil, nil
	}

	b.logger.Info("flushing batch", "count", len(b.pending))

	batch := &BatchReport{
		Reports:   make([]OnChainReport, 0, len(b.pending)),
		BatchID:   uuid.New().String(),
		Timestamp: time.Now(),
	}

	for _, result := range b.pending {
		txHash, err := b.reporter.SubmitReport(ctx, result)
		if err != nil {
			return nil, fmt.Errorf("batch submission failed for node %s: %w", result.NodeID, err)
		}
		batch.Reports = append(batch.Reports, OnChainReport{
			NodeID:    result.NodeID,
			CUValue:   result.FinalCU,
			Timestamp: result.Timestamp,
			TxHash:    txHash,
		})
	}

	b.pending = b.pending[:0]

	b.logger.Info("batch submitted",
		"batch_id", batch.BatchID,
		"reports", len(batch.Reports),
	)

	return batch, nil
}

// PendingCount returns the number of pending results.
func (b *BatchReporter) PendingCount() int {
	return len(b.pending)
}
