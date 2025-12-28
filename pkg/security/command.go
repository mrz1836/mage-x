// Package security provides secure command execution and validation
package security

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	mageExec "github.com/mrz1836/mage-x/pkg/exec"
	"github.com/mrz1836/mage-x/pkg/retry"
)

// Re-export validation errors from exec package for backwards compatibility
var (
	ErrCommandNotAllowed      = mageExec.ErrCommandNotAllowed
	ErrPathTraversal          = mageExec.ErrPathTraversal
	ErrCommandPathTraversal   = mageExec.ErrCommandPathTraversal
	ErrInvalidUTF8            = mageExec.ErrInvalidUTF8
	ErrDangerousPipePattern   = mageExec.ErrDangerousPipePattern
	ErrDangerousPattern       = mageExec.ErrDangerousPattern
	ErrPathContainsNull       = mageExec.ErrPathContainsNull
	ErrPathContainsControl    = mageExec.ErrPathContainsControl
	ErrAbsolutePathNotAllowed = mageExec.ErrAbsolutePathNotAllowed
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

// SecureExecutor implements CommandExecutor with security checks.
// This is a thin wrapper around pkg/exec that provides a familiar API.
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

// getExecutor returns the underlying executor, building it if necessary
func (e *SecureExecutor) getExecutor() mageExec.Executor {
	// Build executor on demand based on current configuration
	builder := mageExec.NewBuilder()

	// Set working directory
	if e.WorkingDir != "" {
		builder = builder.WithWorkingDirectory(e.WorkingDir)
	}

	// Set dry run mode
	builder = builder.WithDryRun(e.DryRun)

	// Add validation if allowed commands are set
	if len(e.AllowedCommands) > 0 {
		allowedList := make([]string, 0, len(e.AllowedCommands))
		for cmd := range e.AllowedCommands {
			allowedList = append(allowedList, cmd)
		}
		builder = builder.WithValidation(mageExec.WithAllowedCommands(allowedList))
	} else {
		// Still add validation for argument checking, just without command whitelist
		builder = builder.WithValidation()
	}

	// Add environment filtering
	builder = builder.WithEnvFiltering(mageExec.WithEnvWhitelist(e.EnvWhitelist))

	// Add timeout
	builder = builder.WithTimeout(e.Timeout)

	return builder.Build()
}

// Execute runs a command with security checks
func (e *SecureExecutor) Execute(ctx context.Context, name string, args ...string) error {
	return e.getExecutor().Execute(ctx, name, args...)
}

// ExecuteOutput runs a command and returns its output
func (e *SecureExecutor) ExecuteOutput(ctx context.Context, name string, args ...string) (string, error) {
	return e.getExecutor().ExecuteOutput(ctx, name, args...)
}

// ExecuteWithEnv runs a command with custom environment variables
func (e *SecureExecutor) ExecuteWithEnv(ctx context.Context, env []string, name string, args ...string) error {
	executor := e.getExecutor()

	// Try to use ExecutorWithEnv interface
	if envExecutor, ok := executor.(mageExec.ExecutorWithEnv); ok {
		return envExecutor.ExecuteWithEnv(ctx, env, name, args...)
	}

	// Fallback to regular execute (env will be ignored, but this shouldn't happen)
	return executor.Execute(ctx, name, args...)
}

// ExecuteWithRetry executes a command with retry logic for network-related commands
func (e *SecureExecutor) ExecuteWithRetry(ctx context.Context, maxRetries int, initialDelay time.Duration,
	name string, args ...string,
) error {
	return mageExec.ExecuteWithRetry(ctx, e.getExecutor(), maxRetries, initialDelay, name, args...)
}

// ExecuteOutputWithRetry executes a command with retry logic and returns output
func (e *SecureExecutor) ExecuteOutputWithRetry(ctx context.Context, maxRetries int, initialDelay time.Duration,
	name string, args ...string,
) (string, error) {
	return mageExec.ExecuteOutputWithRetry(ctx, e.getExecutor(), maxRetries, initialDelay, name, args...)
}

// validateCommand checks if a command is allowed to run (for testing backwards compatibility)
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

	// Validate arguments
	for _, arg := range args {
		if err := ValidateCommandArg(arg); err != nil {
			return fmt.Errorf("invalid argument '%s': %w", arg, err)
		}
	}

	return nil
}

// FilterEnvironment removes sensitive environment variables (exported for testing)
func (e *SecureExecutor) FilterEnvironment(env []string, commandName string) []string {
	return e.filterEnvironment(env, commandName)
}

// filterEnvironment removes sensitive environment variables
func (e *SecureExecutor) filterEnvironment(env []string, commandName string) []string {
	// Use exec package's env filter logic
	filter := mageExec.NewEnvFilteringExecutor(nil, mageExec.WithEnvWhitelist(e.EnvWhitelist))
	return filter.FilterEnvironment(env, commandName)
}

// ValidateCommandArg validates a command argument for security issues
// Re-exported from exec package for backwards compatibility
func ValidateCommandArg(arg string) error {
	return mageExec.ValidateCommandArg(arg)
}

// ValidatePath validates a file path for security issues
// Re-exported from exec package for backwards compatibility
func ValidatePath(path string) error {
	return mageExec.ValidatePath(path)
}

// isRetriableCommandError determines if a command error should trigger a retry
// Re-exported for backwards compatibility
func isRetriableCommandError(err error) bool {
	if err == nil {
		return false
	}

	// First try the retry package's command classifier
	if retry.NewCommandClassifier().IsRetriable(err) {
		return true
	}

	// Also check error string for patterns (backwards compatibility)
	errorStr := strings.ToLower(err.Error())
	retriablePatterns := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"timeout",
		"temporary failure in name resolution",
		"no such host",
		"no route to host",
		"network is unreachable",
		"host is unreachable",
		"i/o timeout",
		"context deadline exceeded",
		"unexpected eof",
		"tls handshake timeout",
		"dial tcp",
		"proxyconnect tcp",
		"go: downloading",
		"go: module",
		"verifying module",
		"getting requirements",
		"sumdb verification",
		"network timeout",
	}

	for _, pattern := range retriablePatterns {
		if strings.Contains(errorStr, pattern) {
			return true
		}
	}

	return false
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
