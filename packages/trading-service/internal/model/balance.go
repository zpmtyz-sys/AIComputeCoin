package model

import (
	"errors"
	"sync"

	"github.com/shopspring/decimal"
)

// ErrInsufficientFunds is returned when an operation would result in negative available balance.
var ErrInsufficientFunds = errors.New("insufficient funds")

// Balance represents a user's account balance.
type Balance struct {
	mu        sync.RWMutex
	UserID    string          `json:"user_id"`
	Available decimal.Decimal `json:"available"`
	Locked    decimal.Decimal `json:"locked"`
}

// NewBalance creates a new Balance with the given initial available amount.
func NewBalance(userID string, available decimal.Decimal) *Balance {
	return &Balance{
		UserID:    userID,
		Available: available,
		Locked:    decimal.Zero,
	}
}

// Total returns the total balance (available + locked).
func (b *Balance) Total() decimal.Decimal {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Available.Add(b.Locked)
}

// Lock moves the specified amount from available to locked.
func (b *Balance) Lock(amount decimal.Decimal) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if amount.IsNegative() || amount.IsZero() {
		return errors.New("lock amount must be positive")
	}
	if b.Available.LessThan(amount) {
		return ErrInsufficientFunds
	}
	b.Available = b.Available.Sub(amount)
	b.Locked = b.Locked.Add(amount)
	return nil
}

// Unlock moves the specified amount from locked to available.
func (b *Balance) Unlock(amount decimal.Decimal) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if amount.IsNegative() || amount.IsZero() {
		return errors.New("unlock amount must be positive")
	}
	if b.Locked.LessThan(amount) {
		return errors.New("unlock amount exceeds locked balance")
	}
	b.Locked = b.Locked.Sub(amount)
	b.Available = b.Available.Add(amount)
	return nil
}

// Credit adds the specified amount to available balance.
func (b *Balance) Credit(amount decimal.Decimal) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if amount.IsNegative() || amount.IsZero() {
		return errors.New("credit amount must be positive")
	}
	b.Available = b.Available.Add(amount)
	return nil
}

// Debit removes the specified amount from available balance.
func (b *Balance) Debit(amount decimal.Decimal) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if amount.IsNegative() || amount.IsZero() {
		return errors.New("debit amount must be positive")
	}
	if b.Available.LessThan(amount) {
		return ErrInsufficientFunds
	}
	b.Available = b.Available.Sub(amount)
	return nil
}

// GetAvailable returns the available balance (thread-safe read).
func (b *Balance) GetAvailable() decimal.Decimal {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Available
}

// GetLocked returns the locked balance (thread-safe read).
func (b *Balance) GetLocked() decimal.Decimal {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Locked
}
