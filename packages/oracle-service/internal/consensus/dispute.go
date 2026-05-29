package consensus

import "time"

// DisputeResolver handles escalation when consensus fails.
type DisputeResolver struct{}

// NewDisputeResolver creates a new DisputeResolver.
func NewDisputeResolver() *DisputeResolver {
	return &DisputeResolver{}
}

// Escalate creates a dispute escalation record when consensus cannot be reached.
func (d *DisputeResolver) Escalate(measurements []Measurement, reason string) *DisputeEscalation {
	nodeID := ""
	if len(measurements) > 0 {
		nodeID = measurements[0].NodeID
	}

	return &DisputeEscalation{
		NodeID:       nodeID,
		Reason:       reason,
		Measurements: measurements,
		Timestamp:    time.Now(),
	}
}
