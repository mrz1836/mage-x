// Package utils provides common utilities for MAGE-X
package utils

import (
	"strings"
)

// ParseParams parses command-line arguments into a map of key-value pairs
// Supports both key=value and boolean flags (key without value means key=true)
//
// Examples:
//
//	ParseParams([]string{"bump=patch", "push", "dry-run=false"})
//	Returns: map[string]string{"bump": "patch", "push": "true", "dry-run": "false"}
//
//	ParseParams([]string{"verbose", "output=json", "count=5"})
//	Returns: map[string]string{"verbose": "true", "output": "json", "count": "5"}
func ParseParams(args []string) map[string]string {
	params := make(map[string]string)

	for _, arg := range args {
		if strings.Contains(arg, "=") {
			// Key=value format
			parts := strings.SplitN(arg, "=", 2)
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key != "" {
				params[key] = value
			}
		} else {
			// Boolean flag format (key without value means true)
			key := strings.TrimSpace(arg)
			if key != "" {
				params[key] = trueValue
			}
		}
	}

	return params
}

// GetParam retrieves a parameter value with a default fallback
func GetParam(params map[string]string, key, defaultValue string) string {
	if value, exists := params[key]; exists {
		return value
	}
	return defaultValue
}

// HasParam checks if a parameter exists (useful for boolean flags)
func HasParam(params map[string]string, key string) bool {
	_, exists := params[key]
	return exists
}

// IsParamTrue checks if a parameter is set to a truthy value
// Considers "true", "1", "yes", "on" as true (case-insensitive)
func IsParamTrue(params map[string]string, key string) bool {
	value, exists := params[key]
	if !exists {
		return false
	}

	value = strings.ToLower(strings.TrimSpace(value))
	return value == "true" || value == "1" || value == "yes" || value == "on"
}

// IsParamFalse checks if a parameter is explicitly set to a falsy value
// Considers "false", "0", "no", "off" as false (case-insensitive)
func IsParamFalse(params map[string]string, key string) bool {
	value, exists := params[key]
	if !exists {
		return false
	}

	value = strings.ToLower(strings.TrimSpace(value))
	return value == "false" || value == "0" || value == "no" || value == "off"
}
