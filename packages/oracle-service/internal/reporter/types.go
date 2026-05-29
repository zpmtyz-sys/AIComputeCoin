package reporter

import "time"

// OnChainReport represents a report to be submitted to the blockchain.
type OnChainReport struct {
	NodeID           string
	CUValue          float64
	Timestamp        time.Time
	OracleSignatures []SignatureEntry
	TxHash           string
}

// SignatureEntry pairs an oracle ID with its signature.
type SignatureEntry struct {
	OracleID  string
	Signature []byte
}

// BatchReport aggregates multiple reports for gas-efficient submission.
type BatchReport struct {
	Reports   []OnChainReport
	BatchID   string
	Timestamp time.Time
	TxHash    string
}

// HeartbeatMessage is a periodic liveness proof.
type HeartbeatMessage struct {
	OracleID  string
	Timestamp time.Time
	Sequence  uint64
}

// ReportStatus tracks the state of a submitted report.
type ReportStatus int

const (
	StatusPending ReportStatus = iota
	StatusSubmitted
	StatusConfirmed
	StatusFailed
)
