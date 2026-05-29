package reporter

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

const (
	// HeartbeatInterval is the interval between heartbeat messages.
	HeartbeatInterval = 10 * time.Minute
)

// Heartbeat sends periodic liveness proofs.
type Heartbeat struct {
	oracleID string
	interval time.Duration
	sequence uint64
	logger   *slog.Logger
	mu       sync.Mutex
	running  bool
	stopCh   chan struct{}
	onBeat   func(HeartbeatMessage)
}

// NewHeartbeat creates a new Heartbeat with the specified oracle ID.
func NewHeartbeat(oracleID string, logger *slog.Logger) *Heartbeat {
	if logger == nil {
		logger = slog.Default()
	}
	return &Heartbeat{
		oracleID: oracleID,
		interval: HeartbeatInterval,
		logger:   logger,
		stopCh:   make(chan struct{}),
	}
}

// SetInterval sets a custom heartbeat interval (useful for testing).
func (h *Heartbeat) SetInterval(d time.Duration) {
	h.interval = d
}

// SetOnBeat sets a callback function that is called on each heartbeat.
func (h *Heartbeat) SetOnBeat(fn func(HeartbeatMessage)) {
	h.onBeat = fn
}

// Start begins sending periodic heartbeats.
func (h *Heartbeat) Start(ctx context.Context) {
	h.mu.Lock()
	if h.running {
		h.mu.Unlock()
		return
	}
	h.running = true
	h.mu.Unlock()

	go h.run(ctx)
}

// Stop stops the heartbeat.
func (h *Heartbeat) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.running {
		close(h.stopCh)
		h.running = false
	}
}

// IsRunning returns whether the heartbeat is active.
func (h *Heartbeat) IsRunning() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.running
}

// Beat generates a single heartbeat message.
func (h *Heartbeat) Beat() HeartbeatMessage {
	h.mu.Lock()
	h.sequence++
	seq := h.sequence
	h.mu.Unlock()

	msg := HeartbeatMessage{
		OracleID:  h.oracleID,
		Timestamp: time.Now(),
		Sequence:  seq,
	}

	h.logger.Info("heartbeat",
		"oracle_id", h.oracleID,
		"sequence", seq,
	)

	if h.onBeat != nil {
		h.onBeat(msg)
	}

	return msg
}

func (h *Heartbeat) run(ctx context.Context) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	// Send initial heartbeat
	h.Beat()

	for {
		select {
		case <-ticker.C:
			h.Beat()
		case <-ctx.Done():
			h.mu.Lock()
			h.running = false
			h.mu.Unlock()
			return
		case <-h.stopCh:
			return
		}
	}
}
