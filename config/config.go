package config

import (
	"os"
)

// Config holds the application configuration
type Config struct {
	FirebaseProjectID string
}

// Load loads configuration from environment variables
func Load() *Config {
	return &Config{
		FirebaseProjectID: getEnv("FIREBASE_PROJECT_ID", ""),
	}
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
