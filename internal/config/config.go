package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	// Redpanda configuration
	RedpandaBrokers string
	ContentTopic    string
	CreatorTopic    string

	// Simulation configuration
	NumCreators         int
	IntervalMs          int
	AbnormalProbability float64
}

// Load loads configuration from environment variables with fallbacks
func Load() (*Config, error) {
	config := &Config{
		RedpandaBrokers:     getEnv("REDPANDA_BROKERS", "redpanda-1:9092,redpanda-2:9092"),
		ContentTopic:        getEnv("CONTENT_TOPIC", "content"),
		CreatorTopic:        getEnv("CREATOR_TOPIC", "creator"),
		NumCreators:         getEnvAsInt("NUM_CREATORS", 10),
		IntervalMs:          getEnvAsInt("INTERVAL_MS", 1000),
		AbnormalProbability: getEnvAsFloat("ABNORMAL_PROBABILITY", 0.8),
	}

	// Validate configuration
	if config.NumCreators <= 0 {
		return nil, fmt.Errorf("NUM_CREATORS must be greater than 0")
	}

	if config.IntervalMs < 100 {
		return nil, fmt.Errorf("INTERVAL_MS must be at least 100ms")
	}

	if config.AbnormalProbability < 0 || config.AbnormalProbability > 1 {
		return nil, fmt.Errorf("ABNORMAL_PROBABILITY must be between 0 and 1")
	}

	if config.ContentTopic == "" {
		return nil, fmt.Errorf("CONTENT_TOPIC cannot be empty")
	}

	if config.CreatorTopic == "" {
		return nil, fmt.Errorf("CREATOR_TOPIC cannot be empty")
	}

	return config, nil
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as an integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

// getEnvAsFloat gets an environment variable as a float with a fallback value
func getEnvAsFloat(key string, fallback float64) float64 {
	if value, exists := os.LookupEnv(key); exists {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return fallback
}
