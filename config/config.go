package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds NAT Management application configuration
type Config struct {
	// Server configuration
	ServerPort string `json:"server_port"`
	ServerHost string `json:"server_host"`
	Debug      bool   `json:"debug"`
}

// Load loads configuration from environment variables
func Load() *Config {
	// Simple config for NAT Management app
	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerHost: getEnv("SERVER_HOST", "localhost"),
		Debug:      getEnvBool("DEBUG", true),
	}

	log.Printf("🚀 NAT Management App Configuration Loaded")
	log.Printf("   Server: %s:%s", cfg.ServerHost, cfg.ServerPort)
	log.Printf("   Debug Mode: %v", cfg.Debug)

	return cfg
}

// Helper functions to get environment variables
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
