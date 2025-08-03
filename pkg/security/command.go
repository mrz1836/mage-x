// Package security provides secure command execution and validation
package security

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	mageErrors "github.com/mrz1836/mage-x/pkg/common/errors"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for err113 compliance
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
	ErrCommandFailed          = errors.New("command failed")
)

// CommandExecutor defines the interface for executing commands
type CommandExecutor interface {
	// Execute runs a command with the given arguments
	Execute(ctx context.Context, name string, args ...string) error
	// ExecuteOutput runs a command and returns its output
	ExecuteOutput(ctx context.Context, name string, args ...string) (string, error)
	// ExecuteWithEnv runs a command with custom environment variables
	ExecuteWithEnv(ctx context.Context, env []string, name string, args ...string) error
}

// SecureExecutor implements CommandExecutor with security checks
type SecureExecutor struct {
	// AllowedCommands is a whitelist of allowed commands (empty means allow all)
	AllowedCommands map[string]bool
	// WorkingDir is the directory to execute commands in
	WorkingDir string
	// Timeout is the default timeout for commands
	Timeout time.Duration
	// DryRun if true, commands are not actually executed
	DryRun bool
	// EnvWhitelist maps command names to environment variables they're allowed to access
	// e.g. "goreleaser" -> ["GITHUB_TOKEN", "GITLAB_TOKEN"]
	EnvWhitelist map[string][]string
}

// NewSecureExecutor creates a new secure command executor
func NewSecureExecutor() *SecureExecutor {
	return &SecureExecutor{
		AllowedCommands: make(map[string]bool),
		Timeout:         5 * time.Minute,
		EnvWhitelist: map[string][]string{
			"goreleaser": {"GITHUB_TOKEN", "GITLAB_TOKEN", "GITEA_TOKEN"},
		},
	}
}

// Execute runs a command with security checks
func (e *SecureExecutor) Execute(ctx context.Context, name string, args ...string) error {
	// Start timing for audit
	startTime := time.Now()
	var exitCode int
	var success bool

	// Defer audit logging
	defer func() {
		duration := time.Since(startTime)
		e.logAuditEvent(name, args, startTime, duration, exitCode, success)
	}()

	// Validate the command
	if err := e.validateCommand(name, args); err != nil {
		return mageErrors.WrapError(err, "command validation failed")
	}

	// Create command with timeout
	ctx, cancel := e.contextWithTimeout(ctx)
	defer cancel()

	// Log what we're doing (dry run mode)
	if e.DryRun {
		utils.Info("[DRY RUN] Would execute: %s %s", name, strings.Join(args, " "))
		success = true
		return nil
	}

	// Create the command
	cmd := exec.CommandContext(ctx, name, args...)

	// Set working directory if specified
	if e.WorkingDir != "" {
		cmd.Dir = e.WorkingDir
	}

	// Inherit environment but filter sensitive variables
	cmd.Env = e.filterEnvironment(os.Environ(), name)

	// Connect output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute
	if err := cmd.Run(); err != nil {
		exitError := &exec.ExitError{}
		if errors.As(err, &exitError) {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
		success = false
		return mageErrors.CommandFailed(name, args, err)
	}

	exitCode = 0
	success = true
	return nil
}

// ExecuteOutput runs a command and returns its output
func (e *SecureExecutor) ExecuteOutput(ctx context.Context, name string, args ...string) (string, error) {
	// Start timing for audit
	startTime := time.Now()
	var exitCode int
	var success bool

	// Defer audit logging
	defer func() {
		duration := time.Since(startTime)
		e.logAuditEvent(name, args, startTime, duration, exitCode, success)
	}()

	// Validate the command
	if err := e.validateCommand(name, args); err != nil {
		return "", mageErrors.WrapError(err, "command validation failed")
	}

	// Create command with timeout
	ctx, cancel := e.contextWithTimeout(ctx)
	defer cancel()

	// Log what we're doing (dry run mode)
	if e.DryRun {
		success = true
		return fmt.Sprintf("[DRY RUN] Would execute: %s %s", name, strings.Join(args, " ")), nil
	}

	// Create the command
	cmd := exec.CommandContext(ctx, name, args...)

	// Set working directory if specified
	if e.WorkingDir != "" {
		cmd.Dir = e.WorkingDir
	}

	// Inherit environment but filter sensitive variables
	cmd.Env = e.filterEnvironment(os.Environ(), name)

	// Execute and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		exitError := &exec.ExitError{}
		if errors.As(err, &exitError) {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
		success = false
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	exitCode = 0
	success = true
	return string(output), nil
}

// ExecuteWithEnv runs a command with custom environment variables
func (e *SecureExecutor) ExecuteWithEnv(ctx context.Context, env []string, name string, args ...string) error {
	// Start timing for audit
	startTime := time.Now()
	var exitCode int
	var success bool

	// Defer audit logging
	defer func() {
		duration := time.Since(startTime)
		e.logAuditEvent(name, args, startTime, duration, exitCode, success)
	}()

	// Validate the command
	if err := e.validateCommand(name, args); err != nil {
		return mageErrors.WrapError(err, "command validation failed")
	}

	// Create command with timeout
	ctx, cancel := e.contextWithTimeout(ctx)
	defer cancel()

	// Log what we're doing (dry run mode)
	if e.DryRun {
		utils.Info("[DRY RUN] Would execute with env: %s %s", name, strings.Join(args, " "))
		success = true
		return nil
	}

	// Create the command
	cmd := exec.CommandContext(ctx, name, args...)

	// Set working directory if specified
	if e.WorkingDir != "" {
		cmd.Dir = e.WorkingDir
	}

	// Merge provided environment with filtered base environment
	baseEnv := e.filterEnvironment(os.Environ(), name)
	cmd.Env = make([]string, 0, len(baseEnv)+len(env))
	cmd.Env = append(cmd.Env, baseEnv...)
	cmd.Env = append(cmd.Env, env...)

	// Connect output
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute
	if err := cmd.Run(); err != nil {
		exitError := &exec.ExitError{}
		if errors.As(err, &exitError) {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
		success = false
		return mageErrors.CommandFailed(name, args, err)
	}

	exitCode = 0
	success = true
	return nil
}

// validateCommand checks if a command is allowed to run
func (e *SecureExecutor) validateCommand(name string, args []string) error {
	// Check if command is in allowed list (if list is not empty)
	if len(e.AllowedCommands) > 0 {
		if !e.AllowedCommands[name] {
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

// contextWithTimeout creates a context with timeout if not already present
func (e *SecureExecutor) contextWithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		// Context already has a deadline
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, e.Timeout)
}

// filterEnvironment removes sensitive environment variables
func (se *SecureExecutor) filterEnvironment(env []string, commandName string) []string {
	// List of environment variable prefixes to filter out
	sensitivePrefix := []string{
		"AWS_SECRET",
		"GITHUB_TOKEN",
		"GITLAB_TOKEN",
		"NPM_TOKEN",
		"DOCKER_PASSWORD",
		"DATABASE_PASSWORD",
		"API_KEY",
		"SECRET",
		"PRIVATE_KEY",
	}

	filtered := make([]string, 0, len(env))
	for _, e := range env {
		keep := true

		// Extract the variable name part (before =)
		parts := strings.SplitN(e, "=", 2)
		if len(parts) < 2 {
			// Malformed env var (no =), include it as is
			filtered = append(filtered, e)
			continue
		}

		varName := strings.ToUpper(parts[0])

		// Check if this environment variable contains sensitive data
		for _, prefix := range sensitivePrefix {
			if !strings.HasPrefix(varName, prefix) {
				continue
			}
			if len(varName) != len(prefix) && varName[len(prefix)] != '_' {
				continue
			}

			// Check if this variable is whitelisted for this command
			if whitelistedVars, ok := se.EnvWhitelist[commandName]; ok {
				for _, whitelistedVar := range whitelistedVars {
					if varName == whitelistedVar {
						// This sensitive variable is whitelisted for this command
						keep = true
						goto nextVar
					}
				}
			}
			// Only filter if it's an exact match or followed by underscore
			keep = false
			break
		}
	nextVar:

		if keep {
			filtered = append(filtered, e)
		}
	}

	return filtered
}

// ValidateCommandArg validates a command argument for security issues
func ValidateCommandArg(arg string) error {
	// Check for valid UTF-8
	if !utf8.ValidString(arg) {
		return ErrInvalidUTF8
	}

	// Check for shell injection attempts
	dangerousPatterns := []string{
		"$(",     // Command substitution
		"`",      // Command substitution
		"&&",     // Command chaining
		"||",     // Command chaining
		";",      // Command separator
		">",      // Redirect
		"<",      // Redirect
		"$(echo", // Common injection pattern
		"${IFS}", // Shell variable manipulation
	}

	// Special cases where pipe is dangerous (not in regex or URLs)
	if strings.Contains(arg, "|") {
		// Allow pipe in regex patterns (contains regex metacharacters)
		if !strings.ContainsAny(arg, "^$[]()+*?.{}\\") {
			// Allow pipe in URLs
			if !strings.HasPrefix(arg, "http://") && !strings.HasPrefix(arg, "https://") {
				return ErrDangerousPipePattern
			}
		}
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(arg, pattern) {
			return fmt.Errorf("%w: '%s'", ErrDangerousPattern, pattern)
		}
	}

	return nil
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
		// This looks like a Windows drive path (C:, D:, etc.)
		return ErrAbsolutePathNotAllowed
	}

	// Check for UNC paths
	if strings.HasPrefix(path, "\\\\") {
		return ErrAbsolutePathNotAllowed
	}

	// Check if path is absolute when it shouldn't be
	if filepath.IsAbs(cleaned) && !strings.HasPrefix(cleaned, "/tmp") {
		// Allow absolute paths only in /tmp for now
		return ErrAbsolutePathNotAllowed
	}

	return nil
}

// MockExecutor implements CommandExecutor for testing
type MockExecutor struct {
	// ExecuteCalls records all Execute calls
	ExecuteCalls []CommandCall
	// ExecuteOutputCalls records all ExecuteOutput calls
	ExecuteOutputCalls []CommandCall
	// Responses maps command strings to responses
	Responses map[string]CommandResponse
}

// CommandCall records a command execution call
type CommandCall struct {
	Name string
	Args []string
	Env  []string
}

// CommandResponse defines a mocked command response
type CommandResponse struct {
	Output string
	Error  error
}

// NewMockExecutor creates a new mock executor for testing
func NewMockExecutor() *MockExecutor {
	return &MockExecutor{
		ExecuteCalls:       make([]CommandCall, 0),
		ExecuteOutputCalls: make([]CommandCall, 0),
		Responses:          make(map[string]CommandResponse),
	}
}

// Execute records the call and returns a mocked response
func (m *MockExecutor) Execute(_ context.Context, name string, args ...string) error {
	m.ExecuteCalls = append(m.ExecuteCalls, CommandCall{
		Name: name,
		Args: args,
	})

	key := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
	if resp, ok := m.Responses[key]; ok {
		return resp.Error
	}

	return nil
}

// ExecuteOutput records the call and returns a mocked response
func (m *MockExecutor) ExecuteOutput(_ context.Context, name string, args ...string) (string, error) {
	m.ExecuteOutputCalls = append(m.ExecuteOutputCalls, CommandCall{
		Name: name,
		Args: args,
	})

	key := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
	if resp, ok := m.Responses[key]; ok {
		return resp.Output, resp.Error
	}

	return "", nil
}

// ExecuteWithEnv records the call and returns a mocked response
func (m *MockExecutor) ExecuteWithEnv(_ context.Context, env []string, name string, args ...string) error {
	m.ExecuteCalls = append(m.ExecuteCalls, CommandCall{
		Name: name,
		Args: args,
		Env:  env,
	})

	key := fmt.Sprintf("%s %s", name, strings.Join(args, " "))
	if resp, ok := m.Responses[key]; ok {
		return resp.Error
	}

	return nil
}

// SetResponse sets a mocked response for a command
func (m *MockExecutor) SetResponse(command, output string, err error) {
	m.Responses[command] = CommandResponse{
		Output: output,
		Error:  err,
	}
}

// logAuditEvent logs command execution for audit purposes
func (e *SecureExecutor) logAuditEvent(command string, args []string, startTime time.Time, duration time.Duration, exitCode int, success bool) {
	// Skip audit logging if not available (to avoid import cycles)
	// This will be handled by the audit package when imported
	auditLogger := getAuditLogger()
	if auditLogger == nil {
		return
	}

	// Get current user
	currentUser := "unknown"
	if usr, err := user.Current(); err == nil {
		currentUser = usr.Username
	}

	// Get current working directory
	workingDir := e.WorkingDir
	if workingDir == "" {
		var err error
		workingDir, err = os.Getwd()
		if err != nil {
			workingDir = "."
		}
	}

	// Create filtered environment map
	env := make(map[string]string)
	for _, envVar := range []string{"MAGE_VERBOSE", "MAGE_AUDIT_ENABLED", "GO_VERSION", "GOOS", "GOARCH"} {
		if value := os.Getenv(envVar); value != "" {
			env[envVar] = value
		}
	}

	// Create audit event
	event := AuditEvent{
		Timestamp:   startTime,
		User:        currentUser,
		Command:     command,
		Args:        args,
		WorkingDir:  workingDir,
		Duration:    duration,
		ExitCode:    exitCode,
		Success:     success,
		Environment: env,
		Metadata: map[string]string{
			"executor_type": "SecureExecutor",
			"dry_run":       fmt.Sprintf("%v", e.DryRun),
		},
	}

	// Log the event (ignore errors to avoid breaking command execution)
	if err := auditLogger.LogEvent(event); err != nil {
		// Audit logging failure should not break command execution
		log.Printf("failed to log audit event: %v", err)
	}
}

// AuditEvent represents an audit event (minimal definition to avoid circular imports)
type AuditEvent struct {
	Timestamp   time.Time
	User        string
	Command     string
	Args        []string
	WorkingDir  string
	Duration    time.Duration
	ExitCode    int
	Success     bool
	Environment map[string]string
	Metadata    map[string]string
}

// AuditLogger interface (minimal definition to avoid circular imports)
type AuditLogger interface {
	LogEvent(event AuditEvent) error
}

// getAuditLogger returns the audit logger if available
func getAuditLogger() AuditLogger {
	// This is a placeholder - the actual implementation will be provided
	// by the utils package when it's imported
	return nil
}
