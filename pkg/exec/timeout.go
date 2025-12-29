package exec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"
)

// TimeoutExecutor wraps an executor with timeout support
type TimeoutExecutor struct {
	wrapped FullExecutor

	// DefaultTimeout is the timeout for commands without a specified timeout
	DefaultTimeout time.Duration

	// TimeoutResolver optionally provides per-command timeouts
	TimeoutResolver TimeoutResolver
}

// TimeoutResolver determines the timeout for a specific command
type TimeoutResolver interface {
	// GetTimeout returns the timeout for a command
	GetTimeout(name string, args []string) time.Duration
}

// TimeoutResolverFunc is a function adapter for TimeoutResolver
type TimeoutResolverFunc func(name string, args []string) time.Duration

// GetTimeout implements TimeoutResolver
func (f TimeoutResolverFunc) GetTimeout(name string, args []string) time.Duration {
	return f(name, args)
}

// TimeoutOption configures a TimeoutExecutor
type TimeoutOption func(*TimeoutExecutor)

// WithDefaultTimeout sets the default timeout
func WithDefaultTimeout(timeout time.Duration) TimeoutOption {
	return func(t *TimeoutExecutor) {
		t.DefaultTimeout = timeout
	}
}

// WithTimeoutResolver sets a custom timeout resolver
func WithTimeoutResolver(resolver TimeoutResolver) TimeoutOption {
	return func(t *TimeoutExecutor) {
		t.TimeoutResolver = resolver
	}
}

// NewTimeoutExecutor creates a new timeout executor
func NewTimeoutExecutor(wrapped FullExecutor, opts ...TimeoutOption) *TimeoutExecutor {
	t := &TimeoutExecutor{
		wrapped:        wrapped,
		DefaultTimeout: 5 * time.Minute, // Default timeout
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// Execute runs a command with timeout
func (t *TimeoutExecutor) Execute(ctx context.Context, name string, args ...string) error {
	timeout := t.getTimeout(name, args)
	ctx, cancel := t.contextWithTimeout(ctx, timeout)
	defer cancel()

	err := t.wrapped.Execute(ctx, name, args...)
	return t.wrapTimeoutError(err, name, timeout)
}

// ExecuteOutput runs a command with timeout and returns output
func (t *TimeoutExecutor) ExecuteOutput(ctx context.Context, name string, args ...string) (string, error) {
	timeout := t.getTimeout(name, args)
	ctx, cancel := t.contextWithTimeout(ctx, timeout)
	defer cancel()

	output, err := t.wrapped.ExecuteOutput(ctx, name, args...)
	return output, t.wrapTimeoutError(err, name, timeout)
}

// ExecuteWithEnv runs a command with additional environment variables and timeout
func (t *TimeoutExecutor) ExecuteWithEnv(ctx context.Context, env []string, name string, args ...string) error {
	timeout := t.getTimeout(name, args)
	ctx, cancel := t.contextWithTimeout(ctx, timeout)
	defer cancel()

	err := t.wrapped.ExecuteWithEnv(ctx, env, name, args...)
	return t.wrapTimeoutError(err, name, timeout)
}

// ExecuteInDir runs a command in the specified directory with timeout
func (t *TimeoutExecutor) ExecuteInDir(ctx context.Context, dir, name string, args ...string) error {
	timeout := t.getTimeout(name, args)
	ctx, cancel := t.contextWithTimeout(ctx, timeout)
	defer cancel()

	err := t.wrapped.ExecuteInDir(ctx, dir, name, args...)
	return t.wrapTimeoutError(err, name, timeout)
}

// ExecuteOutputInDir runs a command in the specified directory with timeout and returns output
func (t *TimeoutExecutor) ExecuteOutputInDir(ctx context.Context, dir, name string, args ...string) (string, error) {
	timeout := t.getTimeout(name, args)
	ctx, cancel := t.contextWithTimeout(ctx, timeout)
	defer cancel()

	output, err := t.wrapped.ExecuteOutputInDir(ctx, dir, name, args...)
	return output, t.wrapTimeoutError(err, name, timeout)
}

// ExecuteStreaming runs a command with custom stdout/stderr and timeout
func (t *TimeoutExecutor) ExecuteStreaming(ctx context.Context, stdout, stderr io.Writer, name string, args ...string) error {
	timeout := t.getTimeout(name, args)
	ctx, cancel := t.contextWithTimeout(ctx, timeout)
	defer cancel()

	err := t.wrapped.ExecuteStreaming(ctx, stdout, stderr, name, args...)
	return t.wrapTimeoutError(err, name, timeout)
}

// getTimeout returns the timeout for a command
func (t *TimeoutExecutor) getTimeout(name string, args []string) time.Duration {
	if t.TimeoutResolver != nil {
		return t.TimeoutResolver.GetTimeout(name, args)
	}
	return t.DefaultTimeout
}

// contextWithTimeout creates a context with timeout if not already present
func (t *TimeoutExecutor) contextWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if _, ok := ctx.Deadline(); ok {
		// Context already has a deadline, wrap with cancel capability
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}

// wrapTimeoutError wraps context errors with more descriptive messages
func (t *TimeoutExecutor) wrapTimeoutError(err error, name string, timeout time.Duration) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("command '%s' exceeded timeout of %s: %w", name, timeout, err)
	}
	if errors.Is(err, context.Canceled) {
		return fmt.Errorf("command '%s' was canceled: %w", name, err)
	}
	return err
}

// AdaptiveTimeoutResolver provides per-command timeouts based on command type
type AdaptiveTimeoutResolver struct {
	// CommandTimeouts maps command names to base timeouts
	CommandTimeouts map[string]time.Duration

	// SubcommandTimeouts maps "command subcommand" to timeouts
	SubcommandTimeouts map[string]time.Duration

	// DefaultTimeout is used when no specific timeout is configured
	DefaultTimeout time.Duration
}

// NewAdaptiveTimeoutResolver creates a resolver with sensible defaults
func NewAdaptiveTimeoutResolver() *AdaptiveTimeoutResolver {
	return &AdaptiveTimeoutResolver{
		CommandTimeouts: map[string]time.Duration{
			"golangci-lint": 20 * time.Minute,
			"goreleaser":    30 * time.Minute,
			"staticcheck":   3 * time.Minute,
			"gosec":         3 * time.Minute,
			"govulncheck":   3 * time.Minute,
			"mage":          3 * time.Minute,
		},
		SubcommandTimeouts: map[string]time.Duration{
			"go test":    10 * time.Minute,
			"go install": 5 * time.Minute,
			"go get":     5 * time.Minute,
			"go mod":     5 * time.Minute,
			"go build":   3 * time.Minute,
			"go run":     3 * time.Minute,
			"go vet":     1 * time.Minute,
			"go list":    1 * time.Minute,
		},
		DefaultTimeout: 30 * time.Second,
	}
}

// GetTimeout returns the appropriate timeout for a command
func (r *AdaptiveTimeoutResolver) GetTimeout(name string, args []string) time.Duration {
	// Check for subcommand-specific timeout
	if len(args) > 0 {
		key := name + " " + args[0]
		if timeout, ok := r.SubcommandTimeouts[key]; ok {
			return timeout
		}
	}

	// Check for command-specific timeout
	if timeout, ok := r.CommandTimeouts[name]; ok {
		return timeout
	}

	return r.DefaultTimeout
}

// Ensure TimeoutExecutor implements FullExecutor
var _ FullExecutor = (*TimeoutExecutor)(nil)
