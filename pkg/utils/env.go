package utils

import (
	"fmt"
	"os"
	"strings"

	"github.com/mrz1836/mage-x/pkg/common/env"
)

// TrueValue is the string representation of true for environment variables
const TrueValue = "true"

// GetEnv returns an environment variable with a default value
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvBool returns a boolean environment variable
func GetEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == TrueValue || value == "1" || value == "yes"
}

// GetEnvInt returns an integer environment variable
func GetEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	var result int
	if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
		return result
	}
	return defaultValue
}

// GetEnvClean retrieves an environment variable with inline comment stripping
// It removes anything after " #" and trims whitespace from the value
// This is useful for environment files that contain inline comments like:
// VARIABLE_NAME=value  # comment here
func GetEnvClean(key string) string {
	value := os.Getenv(key)
	if value == "" {
		return ""
	}

	// Find inline comment marker (space followed by #)
	if idx := strings.Index(value, " #"); idx >= 0 {
		value = value[:idx]
	}

	// Trim any leading/trailing whitespace
	return strings.TrimSpace(value)
}

// IsVerbose checks if verbose mode is enabled via environment variables.
// It delegates to env.IsVerbose() which is the canonical source.
func IsVerbose() bool {
	return env.IsVerbose()
}

// IsCI checks if running in CI environment.
// It delegates to env.IsCI() which is the canonical source.
func IsCI() bool {
	return env.IsCI()
}
