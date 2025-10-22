package config

import (
	"os"
)

// Server configuration
type Config struct {
	ServerAddr   string
	DBPath       string
	LogLevel     string
	WebhookURL   string
	WebhookToken string
}

// Load configuration from environment variables or use defaults
func Load() *Config {
	return &Config{
		ServerAddr:   getEnv("SERVER_ADDR", ":8285"),
		DBPath:       getEnv("DB_PATH", "orbit_app.db"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		WebhookURL:   getEnv("WEBHOOK_URL", ""),
		WebhookToken: getEnv("WEBHOOK_TOKEN", ""),
	}
}

// Helper function to get environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
