package consensus

import "time"

// Measurement represents a single oracle's CU measurement for a node.
type Measurement struct {
	OracleID  string
	NodeID    string
	CUValue   float64
	Timestamp time.Time
	Signature []byte
}

// ConsensusResult represents the agreed-upon CU value after BFT consensus.
type ConsensusResult struct {
	NodeID           string
	FinalCU          float64
	OracleSignatures []OracleSignature
	Timestamp        time.Time
	Confidence       float64
}

// OracleSignature is a signature from an oracle agreeing to the consensus result.
type OracleSignature struct {
	OracleID  string
	Signature []byte
}

// DisputeEscalation is returned when consensus fails.
type DisputeEscalation struct {
	NodeID       string
	Reason       string
	Measurements []Measurement
	Timestamp    time.Time
}
