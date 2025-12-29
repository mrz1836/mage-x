package env

import (
	"fmt"
	"strings"
)

// Validator functions for integer parsing.

// Positive returns true if v is greater than zero.
func Positive(v int) bool { return v > 0 }

// NonNegative returns true if v is zero or greater.
func NonNegative(v int) bool { return v >= 0 }

// PositiveFloat returns true if v is greater than zero.
func PositiveFloat(v float64) bool { return v > 0 }

// ParseInt parses an environment variable as an integer with optional validation.
// Returns (value, true) if the variable is set, parses successfully, and passes validation.
// Returns (0, false) if the variable is unset, empty, fails to parse, or fails validation.
func ParseInt(key string, validate func(int) bool) (int, bool) {
	v := MustGet(key)
	if v == "" {
		return 0, false
	}
	var value int
	if _, err := fmt.Sscanf(v, "%d", &value); err != nil {
		return 0, false
	}
	if validate != nil && !validate(value) {
		return 0, false
	}
	return value, true
}

// ParseFloat parses an environment variable as a float64 with optional validation.
// Returns (value, true) if the variable is set, parses successfully, and passes validation.
// Returns (0, false) if the variable is unset, empty, fails to parse, or fails validation.
func ParseFloat(key string, validate func(float64) bool) (float64, bool) {
	v := MustGet(key)
	if v == "" {
		return 0, false
	}
	var value float64
	if _, err := fmt.Sscanf(v, "%f", &value); err != nil {
		return 0, false
	}
	if validate != nil && !validate(value) {
		return 0, false
	}
	return value, true
}

// ParseStringSlice parses a comma-separated environment variable into a string slice.
// Each element is trimmed of whitespace. Empty elements are filtered out.
// Returns nil if the variable is unset or empty.
func ParseStringSlice(key string) []string {
	v := MustGet(key)
	if v == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// ParseBool parses an environment variable as a boolean.
// Recognizes true values: true, 1, yes, on (case-insensitive)
// Recognizes false values: false, 0, no, off (case-insensitive)
// Returns (value, true) if the variable is set and recognized as a boolean.
// Returns (false, false) if the variable is unset, empty, or not a recognized boolean value.
func ParseBool(key string) (value, found bool) {
	v := strings.ToLower(MustGet(key))
	if v == "" {
		return false, false
	}
	switch v {
	case "true", "1", "yes", "on":
		return true, true
	case "false", "0", "no", "off":
		return false, true
	default:
		return false, false
	}
}
