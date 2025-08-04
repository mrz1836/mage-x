package mage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mrz1836/mage-x/pkg/common/providers"
	"github.com/mrz1836/mage-x/pkg/security"
)

// Static errors to satisfy err113 linter
var (
	errRunnerNil = errors.New("runner cannot be nil")
)

// SecureCommandRunner provides a secure implementation of CommandRunner using SecureExecutor
type SecureCommandRunner struct {
	executor *security.SecureExecutor
}

// NewSecureCommandRunner creates a new secure command runner
func NewSecureCommandRunner() CommandRunner {
	return &SecureCommandRunner{
		executor: security.NewSecureExecutor(),
	}
}

// RunCmd executes a command and returns an error if it fails
func (r *SecureCommandRunner) RunCmd(name string, args ...string) error {
	ctx := context.Background()
	// Use adaptive timeout based on command type
	timeout := r.getCommandTimeout(name, args)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return r.executor.Execute(ctx, name, args...)
}

// RunCmdOutput executes a command and returns its output
func (r *SecureCommandRunner) RunCmdOutput(name string, args ...string) (string, error) {
	ctx := context.Background()
	// Use adaptive timeout based on command type
	timeout := r.getCommandTimeout(name, args)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	output, err := r.executor.ExecuteOutput(ctx, name, args...)
	return strings.TrimSpace(output), err
}

// getCommandTimeout returns appropriate timeout based on command type
func (r *SecureCommandRunner) getCommandTimeout(name string, args []string) time.Duration {
	// For go commands, use longer timeouts
	if name == "go" {
		if len(args) > 0 {
			switch args[0] {
			case "test":
				// Allow 10 minutes for go test commands
				return 10 * time.Minute
			case "install", "get", "mod":
				// Allow 5 minutes for package operations
				return 5 * time.Minute
			case "build", "run":
				// Allow 3 minutes for build operations
				return 3 * time.Minute
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
	case "golangci-lint":
		return 5 * time.Minute
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
