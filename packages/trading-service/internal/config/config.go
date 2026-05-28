package config

import "os"

// Config holds the service configuration.
type Config struct {
	DatabaseURL      string
	KafkaURL         string
	MatchingEngineURL string
	Port             string
}

// NewConfig creates a new Config from environment variables with defaults.
func NewConfig() *Config {
	return &Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgresql://computecoin:computecoin_dev@localhost:5432/computecoin"),
		KafkaURL:          getEnv("KAFKA_BROKERS", "localhost:9092"),
		MatchingEngineURL: getEnv("MATCHING_ENGINE_URL", "localhost:50051"),
		Port:              getEnv("PORT", "50052"),
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}
