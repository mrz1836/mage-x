package mage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mrz1836/mage-x/pkg/common/providers"
	"github.com/mrz1836/mage-x/pkg/exec"
)

// Static errors to satisfy err113 linter
var (
	errRunnerNil = errors.New("runner cannot be nil")
)

// SecureCommandRunner provides a secure implementation of CommandRunner using pkg/exec
type SecureCommandRunner struct {
	executor  exec.FullExecutor // Base executor for dir-specific operations
	validated exec.Executor     // Validated executor for regular operations
}

// NewSecureCommandRunner creates a new secure command runner
func NewSecureCommandRunner() CommandRunner {
	base := exec.NewBase(
		exec.WithVerbose(false), // Set via environment if needed
	)
	return &SecureCommandRunner{
		executor:  base,
		validated: exec.NewValidatingExecutor(base),
	}
}

// RunCmd executes a command and returns an error if it fails
func (r *SecureCommandRunner) RunCmd(name string, args ...string) error {
	ctx := context.Background()
	// Use adaptive timeout based on command type
	timeout := r.getCommandTimeout(name, args)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := r.validated.Execute(ctx, name, args...)
	if err != nil {
		// Check if the error is due to context cancellation/timeout
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("command '%s' exceeded timeout of %s (context deadline exceeded): %w",
				name, timeout, err)
		}
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf("command '%s' was canceled after %s: %w",
				name, timeout, err)
		}
	}
	return err
}

// RunCmdOutput executes a command and returns its output
func (r *SecureCommandRunner) RunCmdOutput(name string, args ...string) (string, error) {
	ctx := context.Background()
	// Use adaptive timeout based on command type
	timeout := r.getCommandTimeout(name, args)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	output, err := r.validated.ExecuteOutput(ctx, name, args...)
	if err != nil {
		// Check if the error is due to context cancellation/timeout
		if errors.Is(err, context.DeadlineExceeded) {
			return strings.TrimSpace(output), fmt.Errorf("command '%s' exceeded timeout of %s (context deadline exceeded): %w",
				name, timeout, err)
		}
		if errors.Is(err, context.Canceled) {
			return strings.TrimSpace(output), fmt.Errorf("command '%s' was canceled after %s: %w",
				name, timeout, err)
		}
	}
	return strings.TrimSpace(output), err
}

// RunCmdInDir executes a command in the specified working directory.
// This is goroutine-safe unlike os.Chdir() - each command runs with its own cmd.Dir.
func (r *SecureCommandRunner) RunCmdInDir(dir, name string, args ...string) error {
	ctx := context.Background()
	timeout := r.getCommandTimeout(name, args)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	err := r.executor.ExecuteInDir(ctx, dir, name, args...)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("command '%s' in '%s' exceeded timeout of %s: %w",
				name, dir, timeout, err)
		}
		if errors.Is(err, context.Canceled) {
			return fmt.Errorf("command '%s' in '%s' was canceled: %w",
				name, dir, err)
		}
	}
	return err
}

// RunCmdOutputInDir executes a command in the specified directory and returns output.
// This is goroutine-safe unlike os.Chdir() - each command runs with its own cmd.Dir.
func (r *SecureCommandRunner) RunCmdOutputInDir(dir, name string, args ...string) (string, error) {
	ctx := context.Background()
	timeout := r.getCommandTimeout(name, args)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	output, err := r.executor.ExecuteOutputInDir(ctx, dir, name, args...)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return strings.TrimSpace(output), fmt.Errorf("command '%s' in '%s' exceeded timeout: %w",
				name, dir, err)
		}
		if errors.Is(err, context.Canceled) {
			return strings.TrimSpace(output), fmt.Errorf("command '%s' in '%s' was canceled: %w",
				name, dir, err)
		}
	}
	return strings.TrimSpace(output), err
}

// getCommandTimeout returns appropriate timeout based on command type
func (r *SecureCommandRunner) getCommandTimeout(name string, args []string) time.Duration {
	// For golangci-lint, check if --timeout flag is provided in args
	if name == "golangci-lint" {
		for i := 0; i < len(args)-1; i++ {
			if args[i] == "--timeout" {
				// Parse the timeout value from the next argument
				if duration, err := time.ParseDuration(args[i+1]); err == nil {
					// Add 5 minutes buffer to the configured timeout to prevent context cancellation
					// before golangci-lint's own timeout
					return duration + 5*time.Minute
				}
			}
		}
		// Default golangci-lint timeout if not specified (increased from 5m to 20m)
		return 20 * time.Minute
	}

	// For go commands, use longer timeouts
	if name == "go" {
		if len(args) > 0 {
			switch args[0] {
			case CmdGoTest:
				// Allow 10 minutes for go test commands
				return 10 * time.Minute
			case "install", "get", "mod":
				// Allow 5 minutes for package operations
				return 5 * time.Minute
			case "build", "run":
				// Allow 3 minutes for build operations
				return 3 * time.Minute
			case "vet", "list":
				// Allow 1 minute for vet and list operations
				return 1 * time.Minute
			default:
				// Default go command timeout
				return 2 * time.Minute
			}
		}
		// Default go command timeout (when no args)
		return 2 * time.Minute
	}

	// For mage commands, check if it's a test-related command
	if name == "mage" {
		if len(args) > 0 {
			// Check for test-related mage commands
			if args[0] == "testDefault" || args[0] == "test" ||
				args[0] == "test:default" || args[0] == "test:unit" ||
				args[0] == "test:cover" || args[0] == "test:race" ||
				args[0] == "test:ci" || args[0] == "test:full" {
				// Allow 10 minutes for mage test commands
				return 10 * time.Minute
			}
		}
		// Default mage command timeout (3 minutes for general tasks)
		return 3 * time.Minute
	}

	// For other tools that might take longer
	switch name {
	case "goreleaser":
		// Allow 30 minutes for goreleaser (builds, tests, uploads)
		return 30 * time.Minute
	case "staticcheck", "gosec", "govulncheck":
		return 3 * time.Minute
	default:
		// Default timeout for other commands
		return 30 * time.Second
	}
}

// packageCommandRunnerProvider provides a generic package-level command runner provider using the generic framework
//
//nolint:gochecknoglobals // Required for package-level singleton access pattern
var packageCommandRunnerProvider = providers.NewPackageProvider(func() CommandRunner {
	return NewSecureCommandRunner()
})

// GetRunner returns the secure command runner with thread-safe lazy initialization
func GetRunner() CommandRunner {
	return packageCommandRunnerProvider.Get()
}

// SetRunner allows setting a custom runner (mainly for testing)
func SetRunner(r CommandRunner) error {
	if r == nil {
		return errRunnerNil
	}
	packageCommandRunnerProvider.Set(r)
	return nil
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// formatBytes formats bytes into human readable format
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
