package economics

import (
	"fmt"
	"sync"

	"github.com/shopspring/decimal"
)

// StakeValidator validates oracle staking requirements.
type StakeValidator interface {
	// ValidateStake checks if an oracle has sufficient stake.
	ValidateStake(oracleID string) (bool, error)
	// GetStake returns the stake amount for an oracle.
	GetStake(oracleID string) (decimal.Decimal, error)
}

// InMemoryStakeStore is an in-memory implementation of stake storage for testing.
type InMemoryStakeStore struct {
	mu     sync.RWMutex
	stakes map[string]decimal.Decimal
}

// NewInMemoryStakeStore creates a new InMemoryStakeStore.
func NewInMemoryStakeStore() *InMemoryStakeStore {
	return &InMemoryStakeStore{
		stakes: make(map[string]decimal.Decimal),
	}
}

// SetStake sets the stake for an oracle.
func (s *InMemoryStakeStore) SetStake(oracleID string, amount decimal.Decimal) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stakes[oracleID] = amount
}

// ValidateStake checks if an oracle has the minimum required stake (10,000 CC).
func (s *InMemoryStakeStore) ValidateStake(oracleID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stake, ok := s.stakes[oracleID]
	if !ok {
		return false, fmt.Errorf("oracle %s not found in stake store", oracleID)
	}

	return stake.GreaterThanOrEqual(MinimumStake), nil
}

// GetStake returns the stake amount for an oracle.
func (s *InMemoryStakeStore) GetStake(oracleID string) (decimal.Decimal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stake, ok := s.stakes[oracleID]
	if !ok {
		return decimal.Zero, fmt.Errorf("oracle %s not found in stake store", oracleID)
	}

	return stake, nil
}
