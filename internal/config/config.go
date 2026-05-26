package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	ServerPort         int    // HTTP server port
	BaseURL            string // Base URL for shortened links
	DBPath             string // Path to BadgerDB data directory
	ShortCodeLength    int    // Length of generated short codes (6-10)
	AnalyticsQueueSize int    // Size of analytics event queue
	RedirectStatusCode int    // HTTP status code for redirects (301 or 302)
}

// LoadConfig loads configuration from environment variables with sensible defaults
func LoadConfig() *Config {
	return &Config{
		ServerPort:         getEnvAsInt("SERVER_PORT", 8080),
		BaseURL:            getEnv("BASE_URL", "http://localhost:8080"),
		DBPath:             getEnv("DB_PATH", "./data/badger"),
		ShortCodeLength:    getEnvAsInt("SHORT_CODE_LENGTH", 7),
		AnalyticsQueueSize: getEnvAsInt("ANALYTICS_QUEUE_SIZE", 1000),
		RedirectStatusCode: getEnvAsInt("REDIRECT_STATUS_CODE", 302),
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
