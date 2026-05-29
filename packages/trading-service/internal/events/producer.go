package events

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// EventProducer is the interface for producing domain events.
type EventProducer interface {
	Publish(eventType EventType, payload any) error
	GetEvents() []Event
}

// InMemoryProducer is an in-memory implementation of EventProducer for testing.
type InMemoryProducer struct {
	mu     sync.RWMutex
	events []Event
}

// NewInMemoryProducer creates a new InMemoryProducer.
func NewInMemoryProducer() *InMemoryProducer {
	return &InMemoryProducer{
		events: make([]Event, 0),
	}
}

// Publish stores an event in memory.
func (p *InMemoryProducer) Publish(eventType EventType, payload any) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	event := Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now(),
		Payload:   payload,
	}

	p.events = append(p.events, event)
	return nil
}

// GetEvents returns all stored events.
func (p *InMemoryProducer) GetEvents() []Event {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]Event, len(p.events))
	copy(result, p.events)
	return result
}

// GetEventsByType returns events filtered by type.
func (p *InMemoryProducer) GetEventsByType(eventType EventType) []Event {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []Event
	for _, e := range p.events {
		if e.Type == eventType {
			result = append(result, e)
		}
	}
	return result
}

// Clear removes all stored events.
func (p *InMemoryProducer) Clear() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = p.events[:0]
}
