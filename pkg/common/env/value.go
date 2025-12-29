package env

import (
	"os"
	"strings"
)

// CleanValue removes trailing comments from environment variable values.
// It handles both space-prefixed (" #") and tab-prefixed ("\t#") comments.
// Returns empty string for comment-only lines (starting with "#").
//
// Examples:
//
//	CleanValue("value")           -> "value"
//	CleanValue("value #comment")  -> "value"
//	CleanValue("value\t#comment") -> "value"
//	CleanValue("#comment")        -> ""
//	CleanValue("  value  ")       -> "value"
func CleanValue(value string) string {
	if value == "" {
		return ""
	}

	// Trim leading/trailing whitespace first
	value = strings.TrimSpace(value)

	// If it's a comment-only line, return empty string
	if strings.HasPrefix(value, "#") {
		return ""
	}

	// Find inline comment markers (space or tab followed by #)
	if idx := strings.Index(value, " #"); idx >= 0 {
		value = value[:idx]
	}
	if idx := strings.Index(value, "\t#"); idx >= 0 {
		value = value[:idx]
	}

	// Trim any trailing whitespace after comment removal
	return strings.TrimSpace(value)
}

// GetOr retrieves an environment variable, cleans it with CleanValue,
// and returns the default if the result is empty.
// This combines os.Getenv, CleanValue, and default fallback in one call.
//
// Examples:
//
//	GetOr("MAGE_X_TIMEOUT", "30s")     // Returns cleaned env value or "30s"
//	GetOr("UNSET_VAR", "default")      // Returns "default" if not set
//	GetOr("VAR_WITH_COMMENT", "def")   // Cleans "value #comment" to "value"
func GetOr(key, defaultValue string) string {
	if value := CleanValue(os.Getenv(key)); value != "" {
		return value
	}
	return defaultValue
}

// MustGet retrieves an environment variable with CleanValue processing.
// Returns empty string if not set (no default). Use GetOr for default support.
func MustGet(key string) string {
	return CleanValue(os.Getenv(key))
}
