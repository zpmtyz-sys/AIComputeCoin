package validation

import (
	"sync"
	"time"
)

// Tier represents a user's rate limit tier.
type Tier string

const (
	TierDefault     Tier = "DEFAULT"
	TierMarketMaker Tier = "MARKET_MAKER"
)

// tokenBucket implements a token bucket rate limiter for a single user.
type tokenBucket struct {
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

// RateLimiter provides per-user rate limiting using token bucket algorithm.
type RateLimiter struct {
	mu      sync.Mutex
	buckets sync.Map // userID -> *tokenBucket
	rates   map[Tier]float64
	now     func() time.Time // injectable clock for testing
}

// NewRateLimiter creates a new RateLimiter with default rates.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		rates: map[Tier]float64{
			TierDefault:     10.0,  // 10 orders/second
			TierMarketMaker: 100.0, // 100 orders/second
		},
		now: time.Now,
	}
}

// Allow checks if the user is allowed to submit an order based on rate limits.
func (rl *RateLimiter) Allow(userID string, tier Tier) bool {
	rate, ok := rl.rates[tier]
	if !ok {
		rate = rl.rates[TierDefault]
	}

	now := rl.now()
	val, _ := rl.buckets.LoadOrStore(userID, &tokenBucket{
		tokens:     rate,
		maxTokens:  rate,
		refillRate: rate,
		lastRefill: now,
	})

	bucket := val.(*tokenBucket)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens based on elapsed time
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	bucket.tokens += elapsed * bucket.refillRate
	if bucket.tokens > bucket.maxTokens {
		bucket.tokens = bucket.maxTokens
	}
	bucket.lastRefill = now

	// Update rate if tier changed
	if bucket.maxTokens != rate {
		bucket.maxTokens = rate
		bucket.refillRate = rate
	}

	// Try to consume a token
	if bucket.tokens >= 1.0 {
		bucket.tokens--
		return true
	}
	return false
}
