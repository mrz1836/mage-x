// Package security provides secure command execution and validation
package security

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/mrz1836/mage-x/pkg/log"
)

// Static errors for validation
var (
	ErrEmailEmpty              = errors.New("email cannot be empty")
	ErrEmailInvalidDomain      = errors.New("invalid email domain")
	ErrEmailInvalidFormat      = errors.New("invalid email format")
	ErrEnvVarInvalidChar       = errors.New("environment variable contains invalid characters")
	ErrEnvVarInvalidStart      = errors.New("environment variable must start with letter or underscore")
	ErrEnvVarNameEmpty         = errors.New("environment variable name cannot be empty")
	ErrFilenameDirRef          = errors.New("filename cannot be directory reference")
	ErrFilenameEmpty           = errors.New("filename cannot be empty")
	ErrFilenameInvalidChar     = errors.New("filename contains invalid characters")
	ErrFilenameNullByte        = errors.New("filename contains null byte")
	ErrFilenamePathSeparator   = errors.New("filename cannot contain path separators")
	ErrFilenameWhitespace      = errors.New("filename cannot have leading/trailing whitespace")
	ErrGitRefDangerousChar     = errors.New("git ref contains dangerous character")
	ErrGitRefEmpty             = errors.New("git ref cannot be empty")
	ErrGitRefInvalidFormat     = errors.New("invalid git ref format")
	ErrPortInvalidRange        = errors.New("port must be between 1 and 65535")
	ErrURLControlChar          = errors.New("URL contains control character")
	ErrURLEmpty                = errors.New("URL cannot be empty")
	ErrURLInvalidProtocol      = errors.New("invalid protocol - URL must use http or https")
	ErrURLNullByte             = errors.New("URL contains null byte")
	ErrURLSuspiciousPattern    = errors.New("suspicious pattern detected in URL")
	ErrURLWhitespace           = errors.New("URL cannot have leading/trailing whitespace")
	ErrVersionCommandInjection = errors.New("version contains command injection pattern")
	ErrVersionControlChar      = errors.New("version contains control character")
	ErrVersionDoubleV          = errors.New("double 'v' prefix not allowed")
	ErrVersionInvalidFormat    = errors.New("invalid version format")
	ErrVersionInvalidUTF8      = errors.New("version contains invalid UTF-8")
	ErrVersionNullByte         = errors.New("version contains null byte")
	ErrVersionPathTraversal    = errors.New("version contains path traversal pattern")
)

// Error format constants
const (
	errFormatWithContext = "%w: %s"
)

var (
	// Version regex for semantic versioning
	versionRegex = regexp.MustCompile(`^v?\d+\.\d+\.\d+(-[a-zA-Z0-9\-\.]+)?(\+[a-zA-Z0-9\-\.]+)?$`)

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
	if strings.Contains(version, "..") || strings.Contains(version, "./") || strings.Contains(version, "../") || strings.Contains(version, "\\") {
		return fmt.Errorf(errFormatWithContext, ErrVersionPathTraversal, version)
	}
	if strings.Contains(version, "$(") || strings.Contains(version, "`") || strings.Contains(version, "${") {
		return fmt.Errorf(errFormatWithContext, ErrVersionCommandInjection, version)
	}
	// Check for dangerous symbols that could be used for injection
	if strings.Contains(version, "<") || strings.Contains(version, ">") {
		return fmt.Errorf(errFormatWithContext, ErrVersionCommandInjection, version)
	}
	if strings.Contains(version, "\x00") {
		return fmt.Errorf(errFormatWithContext, ErrVersionNullByte, version)
	}
	if strings.Contains(version, "\n") || strings.Contains(version, "\r") || strings.Contains(version, "\t") || strings.Contains(version, "\x1b") {
		return fmt.Errorf(errFormatWithContext, ErrVersionControlChar, version)
	}

	// Remove leading 'v' if present for validation
	cleanVersion := strings.TrimPrefix(version, "v")

	// The cleaned version should not start with 'v' (no double 'v' allowed)
	if strings.HasPrefix(cleanVersion, "v") {
		return fmt.Errorf(errFormatWithContext, ErrVersionDoubleV, version)
	}

	// Check for empty prerelease or metadata sections
	if strings.HasSuffix(cleanVersion, "-") || strings.HasSuffix(cleanVersion, "+") ||
		strings.Contains(cleanVersion, "-+") {
		return fmt.Errorf("%w: %s (prerelease/metadata sections cannot be empty)", ErrVersionInvalidFormat, version)
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
		return fmt.Errorf(errFormatWithContext, ErrGitRefInvalidFormat, ref)
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
			return fmt.Errorf(errFormatWithContext, ErrGitRefDangerousChar, pattern)
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
		return fmt.Errorf(errFormatWithContext, ErrFilenameDirRef, filename)
	}
	if strings.Contains(filename, "\x00") {
		return fmt.Errorf(errFormatWithContext, ErrFilenameNullByte, filename)
	}
	if strings.TrimSpace(filename) != filename {
		return fmt.Errorf(errFormatWithContext, ErrFilenameWhitespace, filename)
	}

	// Check if it's just a filename (no path)
	if filename != filepath.Base(filename) {
		return ErrFilenamePathSeparator
	}

	if !safeFilenameRegex.MatchString(filename) {
		return fmt.Errorf(errFormatWithContext, ErrFilenameInvalidChar, filename)
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

	// Check for dangerous protocols first
	dangerousProtocols := []string{
		"javascript:",
		"file:",
		"ftp:",
	}

	lowerURL := strings.ToLower(url)
	for _, protocol := range dangerousProtocols {
		if strings.HasPrefix(lowerURL, protocol) {
			return ErrURLInvalidProtocol
		}
	}

	// Check for suspicious patterns
	suspicious := []string{
		"data:",
		"vbscript:",
		"about:",
		"chrome:",
		"<script",
		"%3cscript", // lowercase version
		"%3Cscript",
		"onerror=",
		"onload=",
	}

	for _, pattern := range suspicious {
		if strings.Contains(lowerURL, pattern) {
			return fmt.Errorf(errFormatWithContext, ErrURLSuspiciousPattern, pattern)
		}
	}

	// Basic URL validation - must use http or https
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return ErrURLInvalidProtocol
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

	// Check domain doesn't start or end with dot
	domain := parts[1]
	if strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
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
		log.Warn("port %d is a privileged port (< 1024)", port)
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
