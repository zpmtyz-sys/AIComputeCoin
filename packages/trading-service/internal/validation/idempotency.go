package validation

import (
	"sync"
	"time"
)

// idempotencyEntry stores a mapping from client order ID to order ID with expiry.
type idempotencyEntry struct {
	orderID   string
	expiresAt time.Time
}

// IdempotencyStore tracks client order IDs to prevent duplicate submissions.
type IdempotencyStore struct {
	entries sync.Map
	ttl     time.Duration
	stopCh  chan struct{}
	once    sync.Once
}

// NewIdempotencyStore creates a new IdempotencyStore with 24h TTL and starts cleanup.
func NewIdempotencyStore() *IdempotencyStore {
	store := &IdempotencyStore{
		ttl:    24 * time.Hour,
		stopCh: make(chan struct{}),
	}
	go store.cleanup()
	return store
}

// Check returns whether the client order ID already exists and the associated order ID.
func (s *IdempotencyStore) Check(clientOrderID string) (exists bool, orderID string) {
	if clientOrderID == "" {
		return false, ""
	}
	val, ok := s.entries.Load(clientOrderID)
	if !ok {
		return false, ""
	}
	entry := val.(*idempotencyEntry)
	if time.Now().After(entry.expiresAt) {
		s.entries.Delete(clientOrderID)
		return false, ""
	}
	return true, entry.orderID
}

// Store saves a mapping from client order ID to order ID.
func (s *IdempotencyStore) Store(clientOrderID, orderID string) {
	if clientOrderID == "" {
		return
	}
	s.entries.Store(clientOrderID, &idempotencyEntry{
		orderID:   orderID,
		expiresAt: time.Now().Add(s.ttl),
	})
}

// Stop stops the background cleanup goroutine.
func (s *IdempotencyStore) Stop() {
	s.once.Do(func() {
		close(s.stopCh)
	})
}

// cleanup periodically removes expired entries.
func (s *IdempotencyStore) cleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			now := time.Now()
			s.entries.Range(func(key, value any) bool {
				entry := value.(*idempotencyEntry)
				if now.After(entry.expiresAt) {
					s.entries.Delete(key)
				}
				return true
			})
		}
	}
}
