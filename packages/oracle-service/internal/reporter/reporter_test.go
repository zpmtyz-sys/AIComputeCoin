package reporter

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/oracle-service/internal/consensus"
)

func makeConsensusResult(nodeID string, cu float64) *consensus.ConsensusResult {
	return &consensus.ConsensusResult{
		NodeID:  nodeID,
		FinalCU: cu,
		OracleSignatures: []consensus.OracleSignature{
			{OracleID: "oracle-1", Signature: []byte("sig-1")},
			{OracleID: "oracle-2", Signature: []byte("sig-2")},
			{OracleID: "oracle-3", Signature: []byte("sig-3")},
		},
		Timestamp:  time.Now(),
		Confidence: 1.0,
	}
}

func TestReporter_SubmitReport(t *testing.T) {
	submitter := NewMockChainSubmitter(0)
	reporter := NewReporter(submitter, "", slog.Default())

	result := makeConsensusResult("node-1", 1000)
	txHash, err := reporter.SubmitReport(context.Background(), result)
	require.NoError(t, err)
	assert.NotEmpty(t, txHash)
	assert.Equal(t, 1, submitter.Calls())
}

func TestReporter_SubmitWithRetry(t *testing.T) {
	// Fails first 2 attempts, succeeds on 3rd
	submitter := NewMockChainSubmitter(2)
	reporter := NewReporter(submitter, "", slog.Default())
	reporter.retry = &RetryConfig{
		BaseDelay:   1 * time.Millisecond, // Fast for tests
		MaxDelay:    10 * time.Millisecond,
		MaxAttempts: 5,
	}

	result := makeConsensusResult("node-1", 1000)
	txHash, err := reporter.SubmitReport(context.Background(), result)
	require.NoError(t, err)
	assert.NotEmpty(t, txHash)
	assert.Equal(t, 3, submitter.Calls())
}

func TestReporter_SubmitExhaustsRetries(t *testing.T) {
	// Fails all 5 attempts
	submitter := NewMockChainSubmitter(10)
	reporter := NewReporter(submitter, "", slog.Default())
	reporter.retry = &RetryConfig{
		BaseDelay:   1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
		MaxAttempts: 5,
	}

	result := makeConsensusResult("node-1", 1000)
	_, err := reporter.SubmitReport(context.Background(), result)
	assert.Error(t, err)
	assert.Equal(t, 5, submitter.Calls())
}

func TestRetryConfig_ExponentialBackoff(t *testing.T) {
	tests := []struct {
		name        string
		failCount   int
		maxAttempts int
		wantErr     bool
	}{
		{
			name:        "succeeds immediately",
			failCount:   0,
			maxAttempts: 3,
			wantErr:     false,
		},
		{
			name:        "succeeds after 1 retry",
			failCount:   1,
			maxAttempts: 3,
			wantErr:     false,
		},
		{
			name:        "fails all attempts",
			failCount:   5,
			maxAttempts: 3,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			op := func() (string, error) {
				calls++
				if calls <= tt.failCount {
					return "", assert.AnError
				}
				return "success", nil
			}

			config := &RetryConfig{
				BaseDelay:   1 * time.Millisecond,
				MaxDelay:    5 * time.Millisecond,
				MaxAttempts: tt.maxAttempts,
			}

			result, err := config.Execute(context.Background(), op)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, "success", result)
			}
		})
	}
}

func TestRetryConfig_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := &RetryConfig{
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    1 * time.Second,
		MaxAttempts: 5,
	}

	calls := 0
	op := func() (string, error) {
		calls++
		return "", assert.AnError
	}

	_, err := config.Execute(ctx, op)
	assert.Error(t, err)
	// Should have stopped early due to context cancellation
	assert.LessOrEqual(t, calls, 2)
}

func TestCache_StoreAndRecover(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.json")

	// Store a report
	cache1 := NewCache(cachePath)
	report := &OnChainReport{
		NodeID:    "node-1",
		CUValue:   1000,
		Timestamp: time.Now(),
	}
	err := cache1.Store(report)
	require.NoError(t, err)
	assert.Equal(t, 1, cache1.Count())

	// Recover from a new cache instance (simulates crash recovery)
	cache2 := NewCache(cachePath)
	pending := cache2.GetPending()
	assert.Len(t, pending, 1)
	assert.Equal(t, "node-1", pending[0].NodeID)
	assert.Equal(t, 1000.0, pending[0].CUValue)
}

func TestCache_Remove(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "cache.json")

	cache := NewCache(cachePath)
	cache.Store(&OnChainReport{NodeID: "node-1", CUValue: 1000})
	cache.Store(&OnChainReport{NodeID: "node-2", CUValue: 2000})
	assert.Equal(t, 2, cache.Count())

	cache.Remove("node-1")
	assert.Equal(t, 1, cache.Count())

	pending := cache.GetPending()
	assert.Len(t, pending, 1)
	assert.Equal(t, "node-2", pending[0].NodeID)
}

func TestCache_EmptyPath(t *testing.T) {
	cache := NewCache("")
	err := cache.Store(&OnChainReport{NodeID: "node-1", CUValue: 1000})
	assert.NoError(t, err)
	assert.Equal(t, 1, cache.Count())
}

func TestCache_NonexistentPath(t *testing.T) {
	cache := NewCache("/nonexistent/path/cache.json")
	// Should not panic, just skip loading
	assert.Equal(t, 0, cache.Count())
}

func TestBatchReporter_BatchAccumulation(t *testing.T) {
	submitter := NewMockChainSubmitter(0)
	reporter := NewReporter(submitter, "", slog.Default())
	batch := NewBatchReporter(reporter, 3, slog.Default())

	// Add 2 reports - should not flush
	_, err := batch.Add(context.Background(), makeConsensusResult("node-1", 1000))
	require.NoError(t, err)
	_, err = batch.Add(context.Background(), makeConsensusResult("node-2", 1100))
	require.NoError(t, err)
	assert.Equal(t, 2, batch.PendingCount())

	// Add 3rd report - should trigger flush
	batchReport, err := batch.Add(context.Background(), makeConsensusResult("node-3", 900))
	require.NoError(t, err)
	require.NotNil(t, batchReport)
	assert.Len(t, batchReport.Reports, 3)
	assert.Equal(t, 0, batch.PendingCount())
}

func TestBatchReporter_ManualFlush(t *testing.T) {
	submitter := NewMockChainSubmitter(0)
	reporter := NewReporter(submitter, "", slog.Default())
	batch := NewBatchReporter(reporter, 10, slog.Default())

	batch.Add(context.Background(), makeConsensusResult("node-1", 1000))
	batch.Add(context.Background(), makeConsensusResult("node-2", 1100))
	assert.Equal(t, 2, batch.PendingCount())

	batchReport, err := batch.Flush(context.Background())
	require.NoError(t, err)
	require.NotNil(t, batchReport)
	assert.Len(t, batchReport.Reports, 2)
	assert.Equal(t, 0, batch.PendingCount())
}

func TestBatchReporter_EmptyFlush(t *testing.T) {
	submitter := NewMockChainSubmitter(0)
	reporter := NewReporter(submitter, "", slog.Default())
	batch := NewBatchReporter(reporter, 10, slog.Default())

	batchReport, err := batch.Flush(context.Background())
	require.NoError(t, err)
	assert.Nil(t, batchReport)
}

func TestHeartbeat_Beat(t *testing.T) {
	hb := NewHeartbeat("oracle-1", slog.Default())

	msg := hb.Beat()
	assert.Equal(t, "oracle-1", msg.OracleID)
	assert.Equal(t, uint64(1), msg.Sequence)

	msg2 := hb.Beat()
	assert.Equal(t, uint64(2), msg2.Sequence)
}

func TestHeartbeat_Callback(t *testing.T) {
	hb := NewHeartbeat("oracle-1", slog.Default())

	var received []HeartbeatMessage
	var mu sync.Mutex
	hb.SetOnBeat(func(msg HeartbeatMessage) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, msg)
	})

	hb.Beat()
	hb.Beat()

	mu.Lock()
	assert.Len(t, received, 2)
	mu.Unlock()
}

func TestHeartbeat_StartStop(t *testing.T) {
	hb := NewHeartbeat("oracle-1", slog.Default())
	hb.SetInterval(10 * time.Millisecond)

	var beats int
	var mu sync.Mutex
	hb.SetOnBeat(func(msg HeartbeatMessage) {
		mu.Lock()
		defer mu.Unlock()
		beats++
	})

	ctx, cancel := context.WithCancel(context.Background())
	hb.Start(ctx)
	assert.True(t, hb.IsRunning())

	// Wait for a few beats
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Give goroutine time to stop
	time.Sleep(20 * time.Millisecond)

	mu.Lock()
	assert.Greater(t, beats, 1)
	mu.Unlock()
}

func TestReportConstruction(t *testing.T) {
	submitter := NewMockChainSubmitter(0)
	reporter := NewReporter(submitter, "", slog.Default())

	result := makeConsensusResult("node-42", 1500.5)
	txHash, err := reporter.SubmitReport(context.Background(), result)
	require.NoError(t, err)
	assert.NotEmpty(t, txHash)
}

func TestCache_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "test_cache.json")

	// Write some reports
	cache := NewCache(cachePath)
	cache.Store(&OnChainReport{NodeID: "node-1", CUValue: 100})
	cache.Store(&OnChainReport{NodeID: "node-2", CUValue: 200})

	// Verify file exists
	_, err := os.Stat(cachePath)
	assert.NoError(t, err)

	// Load from file
	cache2 := NewCache(cachePath)
	assert.Equal(t, 2, cache2.Count())
}
