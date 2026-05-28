package config

import "os"

// Config holds the oracle service configuration.
type Config struct {
	ChainRPC        string
	StakeAmount     string
	MeasureInterval string
	Port            string
}

// NewConfig creates a new Config from environment variables with defaults.
func NewConfig() *Config {
	return &Config{
		ChainRPC:        getEnv("CHAIN_RPC_URL", "http://localhost:26657"),
		StakeAmount:     getEnv("STAKE_AMOUNT", "1000"),
		MeasureInterval: getEnv("MEASURE_INTERVAL", "60s"),
		Port:            getEnv("PORT", "50053"),
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
