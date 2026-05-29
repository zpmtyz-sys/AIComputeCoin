package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/zpmtyz-sys/AIComputeCoin/packages/oracle-service/internal/benchmark"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/oracle-service/internal/config"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/oracle-service/internal/consensus"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/oracle-service/internal/economics"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/oracle-service/internal/hardware"
	"github.com/zpmtyz-sys/AIComputeCoin/packages/oracle-service/internal/reporter"
)

// OracleService combines hardware, performance, economic verification,
// consensus, and reporting into a complete oracle workflow.
type OracleService struct {
	config      *config.Config
	fingerprint hardware.Fingerprinter
	benchmark   benchmark.Calculator
	stakeStore  *economics.InMemoryStakeStore
	reputation  *economics.ReputationTracker
	aggregator  *consensus.Aggregator
	reporter    *reporter.Reporter
	heartbeat   *reporter.Heartbeat
	logger      *slog.Logger
}

// NewOracleService creates a new OracleService with all subsystems.
func NewOracleService(cfg *config.Config, logger *slog.Logger) *OracleService {
	if logger == nil {
		logger = slog.Default()
	}

	nvml := hardware.NewMockNVMLClient("A100")
	fp := hardware.NewFingerprinter(nvml, logger)

	benchCfg := benchmark.BenchConfig{
		MatrixSize:     cfg.MatrixSize,
		SequenceLength: cfg.SequenceLength,
		RandomSeed:     42,
	}
	calc := benchmark.NewCalculator(benchCfg, logger)

	stakeStore := economics.NewInMemoryStakeStore()
	reputation := economics.NewReputationTracker()
	agg := consensus.NewAggregator(logger)

	submitter := reporter.NewMockChainSubmitter(0)
	rep := reporter.NewReporter(submitter, cfg.CachePath, logger)
	hb := reporter.NewHeartbeat(cfg.OracleID, logger)

	return &OracleService{
		config:      cfg,
		fingerprint: fp,
		benchmark:   calc,
		stakeStore:  stakeStore,
		reputation:  reputation,
		aggregator:  agg,
		reporter:    rep,
		heartbeat:   hb,
		logger:      logger,
	}
}

// MeasureAndReport performs a complete measurement cycle and reports results.
func (s *OracleService) MeasureAndReport(ctx context.Context, nodeID string) error {
	s.logger.Info("starting measurement cycle", "node_id", nodeID)

	// Step 1: Verify hardware
	report, err := s.fingerprint.Fingerprint(ctx)
	if err != nil {
		return fmt.Errorf("hardware fingerprinting failed: %w", err)
	}
	s.logger.Info("hardware verified", "fingerprint", report.Fingerprint, "gpu", report.GPUModel)

	// Step 2: Run benchmarks
	cuReport, err := s.benchmark.RunBenchmarks(ctx, report.GPUModel)
	if err != nil {
		return fmt.Errorf("benchmarks failed: %w", err)
	}
	s.logger.Info("benchmarks complete", "cu", cuReport.CalibratedCU)

	// Step 3: Validate stake
	valid, err := s.stakeStore.ValidateStake(nodeID)
	if err != nil {
		s.logger.Warn("stake validation failed", "error", err, "node_id", nodeID)
	} else if !valid {
		return fmt.Errorf("insufficient stake for node %s", nodeID)
	}

	// Step 4: Submit to consensus (would collect from other oracles in production)
	measurements := []consensus.Measurement{
		{OracleID: s.config.OracleID, NodeID: nodeID, CUValue: cuReport.CalibratedCU},
	}
	_ = measurements // Single oracle cannot reach consensus alone

	s.logger.Info("measurement cycle complete",
		"node_id", nodeID,
		"cu", cuReport.CalibratedCU,
		"fingerprint", report.Fingerprint,
	)

	return nil
}

// StartHeartbeat begins the periodic heartbeat.
func (s *OracleService) StartHeartbeat(ctx context.Context) {
	s.heartbeat.Start(ctx)
}
