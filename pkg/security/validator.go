// Package security provides secure command execution and validation
package security

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Static errors for validation
var (
	ErrVersionInvalidUTF8      = errors.New("version contains invalid UTF-8")
	ErrVersionPathTraversal    = errors.New("version contains path traversal pattern")
	ErrVersionCommandInjection = errors.New("version contains command injection pattern")
	ErrVersionNullByte         = errors.New("version contains null byte")
	ErrVersionControlChar      = errors.New("version contains control character")
	ErrVersionDoubleV          = errors.New("double 'v' prefix not allowed")
	ErrVersionInvalidFormat    = errors.New("invalid version format")
	ErrGitRefEmpty             = errors.New("git ref cannot be empty")
	ErrGitRefInvalidFormat     = errors.New("invalid git ref format")
	ErrGitRefDangerousChar     = errors.New("git ref contains dangerous character")
	ErrFilenameEmpty           = errors.New("filename cannot be empty")
	ErrFilenameInvalidChar     = errors.New("filename contains invalid characters")
	ErrFilenamePathTraversal   = errors.New("filename contains path traversal")
	ErrPathEmpty               = errors.New("path cannot be empty")
	ErrPathInvalidChar         = errors.New("path contains invalid characters")
	ErrPathTraversalDetected   = errors.New("path traversal detected")
	ErrURLEmpty                = errors.New("URL cannot be empty")
	ErrURLInvalidScheme        = errors.New("URL must use https scheme")
	ErrURLDangerousChar        = errors.New("URL contains dangerous characters")
	ErrEmailInvalidFormat      = errors.New("invalid email format")
	ErrEmailInvalidDomain      = errors.New("invalid email domain")
	ErrPortInvalidRange        = errors.New("port must be between 1 and 65535")
	ErrEnvVarNameEmpty         = errors.New("environment variable name cannot be empty")
	ErrEnvVarInvalidStart      = errors.New("environment variable must start with letter or underscore")
	ErrEnvVarInvalidChar       = errors.New("environment variable contains invalid characters")
	ErrFilenameDirRef          = errors.New("filename cannot be directory reference")
	ErrFilenameNullByte        = errors.New("filename contains null byte")
	ErrFilenameWhitespace      = errors.New("filename cannot have leading/trailing whitespace")
	ErrFilenamePathSeparator   = errors.New("filename cannot contain path separators")
	ErrURLNullByte             = errors.New("URL contains null byte")
	ErrURLControlChar          = errors.New("URL contains control character")
	ErrURLWhitespace           = errors.New("URL cannot have leading/trailing whitespace")
	ErrURLInvalidProtocol      = errors.New("URL must start with http:// or https://")
	ErrURLSuspiciousPattern    = errors.New("URL contains suspicious pattern")
	ErrEmailEmpty              = errors.New("email cannot be empty")
)

var (
	// Version regex for semantic versioning
	versionRegex = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9\-.]+)?(\+[a-zA-Z0-9\-.]+)?$`)

	// Git ref regex (branch names, tags)
	gitRefRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_./]+$`)

	// Safe filename regex
	safeFilenameRegex = regexp.MustCompile(`^[a-zA-Z0-9\-_.]+$`)
)

// ValidateVersion validates a semantic version string
func ValidateVersion(version string) error {
	// Check for valid UTF-8
	if !utf8.ValidString(version) {
		return ErrVersionInvalidUTF8
	}

	// Check for dangerous patterns first
	if strings.Contains(version, "..") {
		return fmt.Errorf("%w: %s", ErrVersionPathTraversal, version)
	}
	if strings.Contains(version, "$(") || strings.Contains(version, "`") {
		return fmt.Errorf("%w: %s", ErrVersionCommandInjection, version)
	}
	if strings.Contains(version, "\x00") {
		return fmt.Errorf("%w: %s", ErrVersionNullByte, version)
	}
	if strings.Contains(version, "\n") || strings.Contains(version, "\r") {
		return fmt.Errorf("%w: %s", ErrVersionControlChar, version)
	}

	// Remove leading 'v' if present for validation
	cleanVersion := strings.TrimPrefix(version, "v")

	// The cleaned version should not start with 'v' (no double 'v' allowed)
	if strings.HasPrefix(cleanVersion, "v") {
		return fmt.Errorf("%w: %s", ErrVersionDoubleV, version)
	}

	if !versionRegex.MatchString(cleanVersion) {
		return fmt.Errorf("%w: %s (expected format: X.Y.Z or vX.Y.Z)", ErrVersionInvalidFormat, version)
	}

	return nil
}

// ValidateGitRef validates a git reference (branch or tag name)
func ValidateGitRef(ref string) error {
	if ref == "" {
		return ErrGitRefEmpty
	}

	if !gitRefRegex.MatchString(ref) {
		return fmt.Errorf("%w: %s", ErrGitRefInvalidFormat, ref)
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
			return fmt.Errorf("%w: %s", ErrGitRefDangerousChar, pattern)
		}
	}

	return nil
}

// ValidateFilename validates a filename for safety
func ValidateFilename(filename string) error {
	if filename == "" {
		return ErrFilenameEmpty
	}

	// Check for dangerous patterns first
	if filename == ".." || filename == "." {
		return fmt.Errorf("%w: %s", ErrFilenameDirRef, filename)
	}
	if strings.Contains(filename, "\x00") {
		return fmt.Errorf("%w: %s", ErrFilenameNullByte, filename)
	}
	if strings.TrimSpace(filename) != filename {
		return fmt.Errorf("%w: %s", ErrFilenameWhitespace, filename)
	}

	// Check if it's just a filename (no path)
	if filename != filepath.Base(filename) {
		return ErrFilenamePathSeparator
	}

	if !safeFilenameRegex.MatchString(filename) {
		return fmt.Errorf("%w: %s", ErrFilenameInvalidChar, filename)
	}

	return nil
}

// ValidateURL validates a URL for safety
func ValidateURL(url string) error {
	if url == "" {
		return ErrURLEmpty
	}

	// Check for control characters and whitespace issues
	if strings.Contains(url, "\x00") {
		return ErrURLNullByte
	}
	if strings.Contains(url, "\n") || strings.Contains(url, "\r") {
		return ErrURLControlChar
	}
	if strings.TrimSpace(url) != url {
		return ErrURLWhitespace
	}

	// Basic URL validation
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return ErrURLInvalidProtocol
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
			return fmt.Errorf("%w: %s", ErrURLSuspiciousPattern, pattern)
		}
	}

	return nil
}

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if email == "" {
		return ErrEmailEmpty
	}

	// Basic email validation
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return ErrEmailInvalidFormat
	}

	if parts[0] == "" || parts[1] == "" {
		return ErrEmailInvalidFormat
	}

	// Check domain has at least one dot
	if !strings.Contains(parts[1], ".") {
		return ErrEmailInvalidDomain
	}

	return nil
}

// ValidatePort validates a port number
func ValidatePort(port int) error {
	if port < 1 || port > 65535 {
		return ErrPortInvalidRange
	}

	// Warn about privileged ports
	if port < 1024 {
		// This is just a warning, not an error
		// The caller can decide what to do with privileged ports
		log.Printf("warning: port %d is a privileged port (< 1024)", port)
	}

	return nil
}

// ValidateEnvVar validates an environment variable name
func ValidateEnvVar(name string) error {
	if name == "" {
		return ErrEnvVarNameEmpty
	}

	// Must start with letter or underscore
	if !regexp.MustCompile(`^[a-zA-Z_]`).MatchString(name) {
		return ErrEnvVarInvalidStart
	}

	// Can only contain alphanumeric and underscore
	if !regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`).MatchString(name) {
		return ErrEnvVarInvalidChar
	}

	return nil
}
