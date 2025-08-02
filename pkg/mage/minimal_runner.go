package mage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

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
	return r.executor.Execute(ctx, name, args...)
}

// RunCmdOutput executes a command and returns its output
func (r *SecureCommandRunner) RunCmdOutput(name string, args ...string) (string, error) {
	ctx := context.Background()
	output, err := r.executor.ExecuteOutput(ctx, name, args...)
	return strings.TrimSpace(output), err
}

// RunnerProvider manages the default command runner instance
type RunnerProvider struct {
	runner CommandRunner
	once   sync.Once
}

// defaultProvider holds the package-level runner provider
var defaultProvider = &RunnerProvider{} //nolint:gochecknoglobals // Package-level runner provider for singleton pattern

// GetRunner returns the secure command runner with thread-safe lazy initialization
func GetRunner() CommandRunner {
	defaultProvider.once.Do(func() {
		defaultProvider.runner = NewSecureCommandRunner()
	})
	return defaultProvider.runner
}

// SetRunner allows setting a custom runner (mainly for testing)
func SetRunner(r CommandRunner) error {
	if r == nil {
		return errRunnerNil
	}
	defaultProvider.runner = r
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
