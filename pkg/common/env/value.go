package env

import "strings"

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
