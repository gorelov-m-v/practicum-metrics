// Package config provides helpers for reading environment variables with fallback defaults.
package config

import (
	"os"
	"strconv"
)

// GetEnvAsString returns the value of the environment variable or defaultValue if not set.
func GetEnvAsString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvAsInt returns the environment variable parsed as int, or defaultValue.
func GetEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvAsBool returns the environment variable parsed as bool, or defaultValue.
func GetEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
