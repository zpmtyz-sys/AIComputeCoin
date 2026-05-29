package reporter

import (
	"context"
	"fmt"
	"time"
)

// RetryConfig holds retry configuration with exponential backoff.
type RetryConfig struct {
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	MaxAttempts int
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		BaseDelay:   1 * time.Second,
		MaxDelay:    60 * time.Second,
		MaxAttempts: 5,
	}
}

// Execute runs the operation with exponential backoff retry.
func (r *RetryConfig) Execute(ctx context.Context, op func() (string, error)) (string, error) {
	var lastErr error
	delay := r.BaseDelay

	for attempt := 1; attempt <= r.MaxAttempts; attempt++ {
		result, err := op()
		if err == nil {
			return result, nil
		}

		lastErr = err

		if attempt == r.MaxAttempts {
			break
		}

		// Check context before sleeping
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("retry cancelled: %w", ctx.Err())
		default:
		}

		// Exponential backoff
		time.Sleep(delay)
		delay *= 2
		if delay > r.MaxDelay {
			delay = r.MaxDelay
		}
	}

	return "", fmt.Errorf("operation failed after %d attempts: %w", r.MaxAttempts, lastErr)
}
