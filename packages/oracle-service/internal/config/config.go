package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds the oracle service configuration.
type Config struct {
	// Server
	Port string

	// Chain
	ChainRPC string

	// Staking
	MinStake string

	// Benchmarks
	MatrixSize     int
	SequenceLength int

	// Reporting
	BatchSize         int
	HeartbeatInterval time.Duration
	CachePath         string

	// Retry
	RetryBaseDelay   time.Duration
	RetryMaxDelay    time.Duration
	RetryMaxAttempts int

	// Oracle Identity
	OracleID string
}

// NewConfig creates a new Config from environment variables with defaults.
func NewConfig() *Config {
	return &Config{
		Port:              getEnv("PORT", "50053"),
		ChainRPC:          getEnv("CHAIN_RPC_URL", "http://localhost:26657"),
		MinStake:          getEnv("MIN_STAKE", "10000"),
		MatrixSize:        getEnvInt("MATRIX_SIZE", 4096),
		SequenceLength:    getEnvInt("SEQUENCE_LENGTH", 512),
		BatchSize:         getEnvInt("BATCH_SIZE", 10),
		HeartbeatInterval: getEnvDuration("HEARTBEAT_INTERVAL", 10*time.Minute),
		CachePath:         getEnv("CACHE_PATH", "/tmp/oracle-cache.json"),
		RetryBaseDelay:    getEnvDuration("RETRY_BASE_DELAY", 1*time.Second),
		RetryMaxDelay:     getEnvDuration("RETRY_MAX_DELAY", 60*time.Second),
		RetryMaxAttempts:  getEnvInt("RETRY_MAX_ATTEMPTS", 5),
		OracleID:          getEnv("ORACLE_ID", "oracle-default"),
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultValue
}
