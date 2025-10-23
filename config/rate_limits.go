package config

import (
	"os"
	"strconv"
	"strings"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// Environment (development, staging, production)
	Environment string

	// General API rate limiting
	RequestsPerMinute int

	// Login-specific rate limiting (stricter)
	LoginAttemptsPerMinute int

	// Admin IP Whitelist (bypass rate limiting)
	AdminIPWhitelist []string
}

// LoadRateLimitConfig loads rate limit configuration from environment
func LoadRateLimitConfig() *RateLimitConfig {
	env := getEnv("ENVIRONMENT", "development")

	// Default values based on environment
	defaultRequestsPerMinute := 100  // Development default
	defaultLoginAttempts := 10        // Development default

	if env == "production" {
		defaultRequestsPerMinute = 60  // Production: More restrictive
		defaultLoginAttempts = 5       // Production: Stricter login limits
	} else if env == "staging" {
		defaultRequestsPerMinute = 80  // Staging: Middle ground
		defaultLoginAttempts = 7
	}

	// Load from environment with fallback to defaults
	requestsPerMinute := getEnvInt("RATE_LIMIT_REQUESTS_PER_MINUTE", defaultRequestsPerMinute)
	loginAttempts := getEnvInt("LOGIN_RATE_LIMIT", defaultLoginAttempts)

	// Parse admin IP whitelist
	whitelist := []string{}
	whitelistStr := os.Getenv("ADMIN_IP_WHITELIST")
	if whitelistStr != "" {
		whitelist = strings.Split(whitelistStr, ",")
		// Trim whitespace
		for i, ip := range whitelist {
			whitelist[i] = strings.TrimSpace(ip)
		}
	}

	return &RateLimitConfig{
		Environment:            env,
		RequestsPerMinute:      requestsPerMinute,
		LoginAttemptsPerMinute: loginAttempts,
		AdminIPWhitelist:       whitelist,
	}
}

// IsWhitelisted checks if an IP is in the admin whitelist
func (c *RateLimitConfig) IsWhitelisted(ip string) bool {
	for _, whitelistedIP := range c.AdminIPWhitelist {
		if whitelistedIP == ip {
			return true
		}
	}
	return false
}

// Helper function to get int from environment
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
