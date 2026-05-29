# Phase 4: Oracle Network - Compute Verification

## Overview

Build the oracle network service in Go. This is the unique differentiator of ComputeCoin - a decentralized network of oracle nodes that verify GPU compute capacity, measure standardized Compute Units (CU), and report verified measurements to the blockchain for tokenization.

## Prerequisites

- Go 1.25.1+
- NVIDIA drivers / CUDA toolkit (for GPU detection in production; mocked in tests)
- Working directory: `packages/oracle-service/`
- Blockchain module (Phase 5) for on-chain reporting interface

## Tasks

### Task 1: Hardware Fingerprinting Module

**Files:** `packages/oracle-service/internal/hardware/`

GPU identification and attestation:
- Detect GPU model, VRAM, compute capability via NVML bindings (with mock for tests)
- Generate hardware fingerprint: hash(gpu_model + serial + pci_slot + driver_version)
- TEE attestation stub: simulate Trusted Execution Environment verification
  - Generate attestation report (mock SGX/SEV quote)
  - Verify attestation report signature
  - Extract hardware measurements from report
- Anti-spoofing checks:
  - PCI device tree validation
  - Driver version consistency check
  - VRAM capacity vs reported model cross-reference
- Interface: `Fingerprint(ctx) (*HardwareReport, error)`

### Task 2: Performance Benchmarking Module

**Files:** `packages/oracle-service/internal/benchmark/`

Standardized Compute Unit measurement:
- Matrix multiplication benchmark: 4096x4096 FP32 GEMM operations
- Transformer inference benchmark: forward pass of standard model (BERT-base equivalent)
- Memory bandwidth test: sequential and random access patterns
- Compute Unit formula: CU = weighted_sum(matmul_score * 0.4, inference_score * 0.4, bandwidth_score * 0.2)
- Calibration: normalize scores against reference hardware (A100 = 1000 CU)
- Anti-gaming measures:
  - Randomized benchmark parameters (matrix sizes, sequence lengths)
  - Statistical outlier detection across multiple runs
  - Timing analysis to detect pre-computed results
- Interface: `MeasureCU(ctx, config *BenchConfig) (*CUReport, error)`

### Task 3: Economic Verification

**Files:** `packages/oracle-service/internal/economics/`

Staking and slashing for oracle honesty:
- Operator stake requirement: minimum 10,000 CC tokens to run oracle node
- Stake slashing conditions:
  - Reporting measurements > 2 standard deviations from consensus: 5% slash
  - Failing to report within deadline: 1% slash per miss
  - Proven fraud (fabricated hardware): 100% slash + ban
- Reward distribution: oracles earn 0.1% of verified compute value per epoch
- Reputation scoring: weighted moving average of accuracy over last 100 reports
- Interface: `ValidateStake(ctx, operatorId) error`, `CalculateReward(ctx, epoch) (*Reward, error)`

### Task 4: Oracle Aggregation

**Files:** `packages/oracle-service/internal/consensus/`

BFT consensus among oracle nodes:
- Minimum 3 oracle nodes must measure same hardware
- Byzantine fault tolerance: tolerate f < n/3 malicious oracles
- Aggregation algorithm:
  1. Each oracle submits signed CU measurement
  2. Collect measurements until threshold (2/3 + 1 oracles)
  3. Remove statistical outliers (> 2 sigma from median)
  4. Final CU = median of remaining measurements
  5. All agreeing oracles sign final report
- Dispute resolution: if consensus fails, escalate to governance vote
- Interface: `Aggregate(ctx, measurements []Measurement) (*ConsensusResult, error)`

### Task 5: Reporting Interface

**Files:** `packages/oracle-service/internal/reporter/`

Submit verified CU measurements to blockchain:
- Construct on-chain transaction with: node_id, cu_value, measurement_timestamp, oracle_signatures
- Batch reporting: aggregate multiple node measurements per transaction (gas efficiency)
- Retry logic with exponential backoff for failed submissions
- Local cache of pending reports for crash recovery
- Periodic heartbeat: oracle liveness proof every 10 minutes
- Interface: `SubmitReport(ctx, report *ConsensusResult) (txHash, error)`

## Verification

```bash
cd packages/oracle-service
go fmt ./...
go vet ./...
golangci-lint run
go test ./... -cover -race
```

## Success Criteria

- Hardware fingerprinting correctly identifies GPU capabilities (with mocks)
- Benchmark module produces consistent CU measurements for same hardware
- Economic model correctly calculates rewards and applies slashing
- Oracle consensus converges with 3+ honest nodes
- Reports are correctly formatted for blockchain submission
- All tests pass including race condition detection
- Test coverage > 75%
