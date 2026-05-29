package model

import (
	"sync"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBalance_Lock(t *testing.T) {
	tests := []struct {
		name      string
		available string
		lockAmt   string
		wantErr   bool
		errMsg    string
		wantAvail string
		wantLock  string
	}{
		{
			name:      "successful lock",
			available: "1000",
			lockAmt:   "100",
			wantErr:   false,
			wantAvail: "900",
			wantLock:  "100",
		},
		{
			name:      "lock entire balance",
			available: "500",
			lockAmt:   "500",
			wantErr:   false,
			wantAvail: "0",
			wantLock:  "500",
		},
		{
			name:      "insufficient funds",
			available: "100",
			lockAmt:   "200",
			wantErr:   true,
			errMsg:    "insufficient funds",
			wantAvail: "100",
			wantLock:  "0",
		},
		{
			name:      "zero amount",
			available: "1000",
			lockAmt:   "0",
			wantErr:   true,
			errMsg:    "lock amount must be positive",
			wantAvail: "1000",
			wantLock:  "0",
		},
		{
			name:      "negative amount",
			available: "1000",
			lockAmt:   "-50",
			wantErr:   true,
			errMsg:    "lock amount must be positive",
			wantAvail: "1000",
			wantLock:  "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bal := NewBalance("user1", decimal.RequireFromString(tt.available))
			err := bal.Lock(decimal.RequireFromString(tt.lockAmt))

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
			assert.True(t, bal.GetAvailable().Equal(decimal.RequireFromString(tt.wantAvail)))
			assert.True(t, bal.GetLocked().Equal(decimal.RequireFromString(tt.wantLock)))
		})
	}
}

func TestBalance_Unlock(t *testing.T) {
	tests := []struct {
		name       string
		available  string
		locked     string
		unlockAmt  string
		wantErr    bool
		errMsg     string
		wantAvail  string
		wantLocked string
	}{
		{
			name:       "successful unlock",
			available:  "500",
			locked:     "500",
			unlockAmt:  "200",
			wantErr:    false,
			wantAvail:  "700",
			wantLocked: "300",
		},
		{
			name:       "unlock all",
			available:  "0",
			locked:     "1000",
			unlockAmt:  "1000",
			wantErr:    false,
			wantAvail:  "1000",
			wantLocked: "0",
		},
		{
			name:       "unlock exceeds locked",
			available:  "500",
			locked:     "100",
			unlockAmt:  "200",
			wantErr:    true,
			errMsg:     "unlock amount exceeds locked balance",
			wantAvail:  "500",
			wantLocked: "100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bal := NewBalance("user1", decimal.RequireFromString(tt.available))
			// Set up locked balance
			if locked := decimal.RequireFromString(tt.locked); locked.IsPositive() {
				bal.Available = bal.Available.Add(locked)
				_ = bal.Lock(locked)
			}

			err := bal.Unlock(decimal.RequireFromString(tt.unlockAmt))
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
			assert.True(t, bal.GetAvailable().Equal(decimal.RequireFromString(tt.wantAvail)))
			assert.True(t, bal.GetLocked().Equal(decimal.RequireFromString(tt.wantLocked)))
		})
	}
}

func TestBalance_Credit(t *testing.T) {
	tests := []struct {
		name      string
		available string
		creditAmt string
		wantErr   bool
		wantAvail string
	}{
		{
			name:      "successful credit",
			available: "1000",
			creditAmt: "500",
			wantErr:   false,
			wantAvail: "1500",
		},
		{
			name:      "zero amount",
			available: "1000",
			creditAmt: "0",
			wantErr:   true,
			wantAvail: "1000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bal := NewBalance("user1", decimal.RequireFromString(tt.available))
			err := bal.Credit(decimal.RequireFromString(tt.creditAmt))
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.True(t, bal.GetAvailable().Equal(decimal.RequireFromString(tt.wantAvail)))
		})
	}
}

func TestBalance_Debit(t *testing.T) {
	tests := []struct {
		name      string
		available string
		debitAmt  string
		wantErr   bool
		errMsg    string
		wantAvail string
	}{
		{
			name:      "successful debit",
			available: "1000",
			debitAmt:  "300",
			wantErr:   false,
			wantAvail: "700",
		},
		{
			name:      "insufficient funds",
			available: "100",
			debitAmt:  "200",
			wantErr:   true,
			errMsg:    "insufficient funds",
			wantAvail: "100",
		},
		{
			name:      "debit exact amount",
			available: "500",
			debitAmt:  "500",
			wantErr:   false,
			wantAvail: "0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bal := NewBalance("user1", decimal.RequireFromString(tt.available))
			err := bal.Debit(decimal.RequireFromString(tt.debitAmt))
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
			assert.True(t, bal.GetAvailable().Equal(decimal.RequireFromString(tt.wantAvail)))
		})
	}
}

func TestBalance_Total(t *testing.T) {
	bal := NewBalance("user1", decimal.RequireFromString("1000"))
	_ = bal.Lock(decimal.RequireFromString("300"))
	assert.True(t, bal.Total().Equal(decimal.RequireFromString("1000")))
}

func TestBalance_ConcurrentAccess(t *testing.T) {
	bal := NewBalance("user1", decimal.RequireFromString("10000"))
	var wg sync.WaitGroup

	// Run concurrent operations
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = bal.Lock(decimal.RequireFromString("10"))
		}()
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = bal.Credit(decimal.RequireFromString("5"))
		}()
	}

	wg.Wait()

	// Total should be 10000 + 50*5 = 10250 (credits added)
	// Available = 10000 - (successful locks * 10) + 50*5
	// All 100 locks should succeed since we had 10000 initially
	total := bal.Total()
	assert.True(t, total.Equal(decimal.RequireFromString("10250")))
}
