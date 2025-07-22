// Package security provides secure command execution and validation
package security

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// Version regex for semantic versioning
	versionRegex = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9\-\.]+)?(\+[a-zA-Z0-9\-\.]+)?$`)

	// Git ref regex (branch names, tags)
	gitRefRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_\.\/]+$`)

	// Safe filename regex
	safeFilenameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_\.]+$`)
)

// ValidateVersion validates a semantic version string
func ValidateVersion(version string) error {
	// Check for dangerous patterns first
	if strings.Contains(version, "..") {
		return fmt.Errorf("version contains path traversal pattern: %s", version)
	}
	if strings.Contains(version, "$(") || strings.Contains(version, "`") {
		return fmt.Errorf("version contains command injection pattern: %s", version)
	}
	if strings.Contains(version, "\x00") {
		return fmt.Errorf("version contains null byte: %s", version)
	}
	if strings.Contains(version, "\n") || strings.Contains(version, "\r") {
		return fmt.Errorf("version contains control character: %s", version)
	}

	// Remove leading 'v' if present for validation
	cleanVersion := strings.TrimPrefix(version, "v")

	if !versionRegex.MatchString(cleanVersion) {
		return fmt.Errorf("invalid version format: %s (expected format: X.Y.Z or vX.Y.Z)", version)
	}

	return nil
}

// ValidateGitRef validates a git reference (branch or tag name)
func ValidateGitRef(ref string) error {
	if ref == "" {
		return fmt.Errorf("git ref cannot be empty")
	}

	if !gitRefRegex.MatchString(ref) {
		return fmt.Errorf("invalid git ref format: %s", ref)
	}

	// Check for dangerous patterns
	dangerous := []string{
		"..",
		"~",
		"^",
		":",
		"\\",
		"*",
		"?",
		"[",
		" ",
	}

	for _, pattern := range dangerous {
		if strings.Contains(ref, pattern) {
			return fmt.Errorf("git ref contains dangerous character: %s", pattern)
		}
	}

	return nil
}

// ValidateFilename validates a filename for safety
func ValidateFilename(filename string) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	// Check for dangerous patterns first
	if filename == ".." || filename == "." {
		return fmt.Errorf("filename cannot be directory reference: %s", filename)
	}
	if strings.Contains(filename, "\x00") {
		return fmt.Errorf("filename contains null byte: %s", filename)
	}
	if strings.TrimSpace(filename) != filename {
		return fmt.Errorf("filename cannot have leading/trailing whitespace: %s", filename)
	}

	// Check if it's just a filename (no path)
	if filename != filepath.Base(filename) {
		return fmt.Errorf("filename cannot contain path separators")
	}

	if !safeFilenameRegex.MatchString(filename) {
		return fmt.Errorf("filename contains invalid characters: %s", filename)
	}

	return nil
}

// ValidateURL validates a URL for safety
func ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Check for control characters and whitespace issues
	if strings.Contains(url, "\x00") {
		return fmt.Errorf("URL contains null byte")
	}
	if strings.Contains(url, "\n") || strings.Contains(url, "\r") {
		return fmt.Errorf("URL contains control character")
	}
	if strings.TrimSpace(url) != url {
		return fmt.Errorf("URL cannot have leading/trailing whitespace")
	}

	// Basic URL validation
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	// Check for suspicious patterns
	suspicious := []string{
		"javascript:",
		"data:",
		"vbscript:",
		"file:",
		"about:",
		"chrome:",
		"<script",
		"%3cscript", // lowercase version
		"%3Cscript",
		"onerror=",
		"onload=",
	}

	lowerURL := strings.ToLower(url)
	for _, pattern := range suspicious {
		if strings.Contains(lowerURL, pattern) {
			return fmt.Errorf("URL contains suspicious pattern: %s", pattern)
		}
	}

	return nil
}

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Basic email validation
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return fmt.Errorf("invalid email format")
	}

	if parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid email format")
	}

	// Check domain has at least one dot
	if !strings.Contains(parts[1], ".") {
		return fmt.Errorf("invalid email domain")
	}

	return nil
}

// ValidatePort validates a port number
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	// Warn about privileged ports
	if port < 1024 {
		// This is just a warning, not an error
		// The caller can decide what to do with privileged ports
	}

	return nil
}

// ValidateEnvVar validates an environment variable name
func ValidateEnvVar(name string) error {
	if name == "" {
		return fmt.Errorf("environment variable name cannot be empty")
	}

	// Must start with letter or underscore
	if !regexp.MustCompile(`^[a-zA-Z_]`).MatchString(name) {
		return fmt.Errorf("environment variable must start with letter or underscore")
	}

	// Can only contain alphanumeric and underscore
	if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(name) {
		return fmt.Errorf("environment variable contains invalid characters")
	}

	return nil
}
