package exec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

// Validation errors
var (
	ErrCommandNotAllowed      = errors.New("command is not in allowed list")
	ErrPathTraversal          = errors.New("path traversal detected")
	ErrCommandPathTraversal   = errors.New("command name contains path traversal")
	ErrInvalidUTF8            = errors.New("argument contains invalid UTF-8")
	ErrDangerousPipePattern   = errors.New("potentially dangerous pattern '|' detected")
	ErrDangerousPattern       = errors.New("potentially dangerous pattern detected")
	ErrPathContainsNull       = errors.New("path contains null byte")
	ErrPathContainsControl    = errors.New("path contains control character")
	ErrAbsolutePathNotAllowed = errors.New("absolute paths not allowed outside of /tmp")
)

// ValidatingExecutor wraps an executor with security validation
type ValidatingExecutor struct {
	wrapped FullExecutor

	// AllowedCommands is a whitelist of allowed commands (empty means allow all)
	AllowedCommands map[string]bool
}

// ValidatingOption configures a ValidatingExecutor
type ValidatingOption func(*ValidatingExecutor)

// WithAllowedCommands sets the command whitelist
func WithAllowedCommands(commands []string) ValidatingOption {
	return func(v *ValidatingExecutor) {
		v.AllowedCommands = make(map[string]bool)
		for _, cmd := range commands {
			v.AllowedCommands[cmd] = true
		}
	}
}

// NewValidatingExecutor creates a new validating executor
func NewValidatingExecutor(wrapped FullExecutor, opts ...ValidatingOption) *ValidatingExecutor {
	v := &ValidatingExecutor{
		wrapped:         wrapped,
		AllowedCommands: make(map[string]bool),
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// Execute validates and runs a command
func (v *ValidatingExecutor) Execute(ctx context.Context, name string, args ...string) error {
	if err := v.validate(name, args); err != nil {
		return fmt.Errorf("command validation failed: %w", err)
	}
	return v.wrapped.Execute(ctx, name, args...)
}

// ExecuteOutput validates and runs a command, returning output
func (v *ValidatingExecutor) ExecuteOutput(ctx context.Context, name string, args ...string) (string, error) {
	if err := v.validate(name, args); err != nil {
		return "", fmt.Errorf("command validation failed: %w", err)
	}
	return v.wrapped.ExecuteOutput(ctx, name, args...)
}

// ExecuteWithEnv validates and runs a command with additional environment variables
func (v *ValidatingExecutor) ExecuteWithEnv(ctx context.Context, env []string, name string, args ...string) error {
	if err := v.validate(name, args); err != nil {
		return fmt.Errorf("command validation failed: %w", err)
	}
	return v.wrapped.ExecuteWithEnv(ctx, env, name, args...)
}

// ExecuteInDir validates and runs a command in the specified directory
func (v *ValidatingExecutor) ExecuteInDir(ctx context.Context, dir, name string, args ...string) error {
	if err := v.validate(name, args); err != nil {
		return fmt.Errorf("command validation failed: %w", err)
	}
	return v.wrapped.ExecuteInDir(ctx, dir, name, args...)
}

// ExecuteOutputInDir validates and runs a command in the specified directory, returning output
func (v *ValidatingExecutor) ExecuteOutputInDir(ctx context.Context, dir, name string, args ...string) (string, error) {
	if err := v.validate(name, args); err != nil {
		return "", fmt.Errorf("command validation failed: %w", err)
	}
	return v.wrapped.ExecuteOutputInDir(ctx, dir, name, args...)
}

// ExecuteStreaming validates and runs a command with custom stdout/stderr
func (v *ValidatingExecutor) ExecuteStreaming(ctx context.Context, stdout, stderr io.Writer, name string, args ...string) error {
	if err := v.validate(name, args); err != nil {
		return fmt.Errorf("command validation failed: %w", err)
	}
	return v.wrapped.ExecuteStreaming(ctx, stdout, stderr, name, args...)
}

// validate checks if a command is allowed to run
func (v *ValidatingExecutor) validate(name string, args []string) error {
	// Check if command is in allowed list (if list is not empty)
	if len(v.AllowedCommands) > 0 {
		if !v.AllowedCommands[name] {
			return fmt.Errorf("%w: '%s'", ErrCommandNotAllowed, name)
		}
	}

	// Validate command name doesn't contain path traversal
	if strings.Contains(name, "..") {
		return ErrCommandPathTraversal
	}

	// Validate arguments don't contain dangerous patterns
	for _, arg := range args {
		if err := ValidateCommandArg(arg); err != nil {
			return fmt.Errorf("invalid argument '%s': %w", arg, err)
		}
	}

	return nil
}

// ValidateCommandArg validates a command argument for security issues
func ValidateCommandArg(arg string) error {
	// Check for valid UTF-8
	if !utf8.ValidString(arg) {
		return ErrInvalidUTF8
	}

	// Check for shell injection attempts
	// Order matters - check multi-char patterns (like ||, &&) before single pipe
	dangerousPatterns := []string{
		"$(",     // Command substitution
		"`",      // Command substitution
		"&&",     // Command chaining
		"||",     // Command chaining (must be checked before single |)
		";",      // Command separator
		">",      // Redirect
		"<",      // Redirect
		"$(echo", // Common injection pattern
		"${IFS}", // Shell variable manipulation
	}

	// Check dangerous patterns first (before single pipe check)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(arg, pattern) {
			return fmt.Errorf("%w: '%s'", ErrDangerousPattern, pattern)
		}
	}

	// Special cases where single pipe is dangerous (not in regex or URLs)
	// This runs AFTER the pattern check, so || is caught as a pattern, not a pipe
	if strings.Contains(arg, "|") {
		isRegex := strings.ContainsAny(arg, "^$[]()+*?.{}\\")
		isURL := strings.HasPrefix(arg, "http://") || strings.HasPrefix(arg, "https://")

		if !isRegex && !isURL {
			return ErrDangerousPipePattern
		}

		// Additional check: even with regex chars, reject suspicious command patterns
		if isRegex && containsSuspiciousPipeCommand(arg) {
			return ErrDangerousPipePattern
		}
	}

	return nil
}

// containsSuspiciousPipeCommand checks if an argument containing a pipe has suspicious
// command patterns that suggest shell injection rather than legitimate regex use.
func containsSuspiciousPipeCommand(arg string) bool {
	pipeIdx := strings.Index(arg, "|")
	if pipeIdx == -1 {
		return false
	}

	// Get content after the pipe
	after := arg[pipeIdx+1:]
	afterTrimmed := strings.TrimSpace(after)

	// Check for common command names after pipe
	suspiciousCommands := []string{
		"cat", "rm", "wget", "curl", "bash", "sh", "nc", "python", "perl", "ruby",
		"chmod", "chown", "mv", "cp", "dd", "head", "tail", "grep", "awk", "sed",
		"xargs", "find", "exec", "eval", "source", "env", "sudo",
	}
	for _, cmd := range suspiciousCommands {
		// Check for "| cmd" or "|cmd" followed by space or end
		if strings.HasPrefix(afterTrimmed, cmd+" ") || afterTrimmed == cmd {
			return true
		}
	}

	return false
}

// ValidatePath validates a file path for security issues
func ValidatePath(path string) error {
	// Check for control characters and dangerous patterns first
	if strings.Contains(path, "\x00") {
		return ErrPathContainsNull
	}
	if strings.Contains(path, "\n") || strings.Contains(path, "\r") {
		return ErrPathContainsControl
	}

	// Check for path traversal BEFORE cleaning (Unix and Windows styles)
	if strings.Contains(path, "../") || strings.Contains(path, "..\\") ||
		strings.HasSuffix(path, "..") || path == ".." {
		return ErrPathTraversal
	}

	// Clean the path
	cleaned := filepath.Clean(path)

	// Check for Windows-style paths (which should be rejected on Unix systems)
	if strings.Contains(path, ":") && len(path) > 1 && path[1] == ':' {
		return ErrAbsolutePathNotAllowed
	}

	// Check for UNC paths
	if strings.HasPrefix(path, "\\\\") {
		return ErrAbsolutePathNotAllowed
	}

	// Check if path is absolute when it shouldn't be
	if filepath.IsAbs(cleaned) && !strings.HasPrefix(cleaned, "/tmp") {
		return ErrAbsolutePathNotAllowed
	}

	return nil
}

// Ensure ValidatingExecutor implements FullExecutor
var _ FullExecutor = (*ValidatingExecutor)(nil)
