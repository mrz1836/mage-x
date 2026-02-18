package exec

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Base is the foundational command executor.
// It provides simple command execution without any decorators.
type Base struct {
	// WorkingDir is the directory to execute commands in
	WorkingDir string

	// Env is additional environment variables to set
	Env []string

	// Verbose enables command logging
	Verbose bool

	// DryRun logs commands without executing them
	DryRun bool

	// logger is the function used for verbose logging
	logger func(format string, args ...interface{})
}

// Option configures a Base executor
type Option func(*Base)

// WithWorkingDir sets the working directory
func WithWorkingDir(dir string) Option {
	return func(b *Base) {
		b.WorkingDir = dir
	}
}

// WithEnv sets additional environment variables
func WithEnv(env []string) Option {
	return func(b *Base) {
		b.Env = env
	}
}

// WithVerbose enables verbose command logging
func WithVerbose(verbose bool) Option {
	return func(b *Base) {
		b.Verbose = verbose
	}
}

// WithDryRun enables dry run mode (log commands without executing)
func WithDryRun(dryRun bool) Option {
	return func(b *Base) {
		b.DryRun = dryRun
	}
}

// WithLogger sets a custom logger for verbose output
func WithLogger(logger func(format string, args ...interface{})) Option {
	return func(b *Base) {
		b.logger = logger
	}
}

// NewBase creates a new base executor with optional configuration
func NewBase(opts ...Option) *Base {
	b := &Base{
		logger: func(format string, args ...interface{}) {
			fmt.Printf(format+"\n", args...)
		},
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

// logVerbose logs a message if verbose mode is enabled and logger is set.
func (b *Base) logVerbose(format string, args ...interface{}) {
	if b.Verbose && b.logger != nil {
		b.logger(format, args...)
	}
}

// checkDryRun checks if dry-run mode is enabled. If so, it logs the message
// and returns true (caller should return early). Otherwise returns false.
func (b *Base) checkDryRun(format string, args ...interface{}) bool {
	if b.DryRun {
		if b.logger != nil {
			b.logger(format, args...)
		}
		return true
	}
	return false
}

// Execute runs a command with stdout/stderr connected to os.Stdout/os.Stderr
func (b *Base) Execute(ctx context.Context, name string, args ...string) error {
	return b.ExecuteStreaming(ctx, os.Stdout, os.Stderr, name, args...)
}

// ExecuteOutput runs a command and returns its combined output
func (b *Base) ExecuteOutput(ctx context.Context, name string, args ...string) (string, error) {
	b.logVerbose("➤ %s %s", name, strings.Join(args, " "))
	if b.checkDryRun("[DRY RUN] Would execute: %s %s", name, strings.Join(args, " ")) {
		return "", nil
	}

	cmd := exec.CommandContext(ctx, name, args...) // #nosec G204 -- executor library, callers control command name
	b.configureCommand(cmd)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", CommandErrorWithOutput(name, args, err, string(output))
	}
	return string(output), nil
}

// ExecuteWithEnv runs a command with additional environment variables
func (b *Base) ExecuteWithEnv(ctx context.Context, env []string, name string, args ...string) error {
	b.logVerbose("➤ %s %s (with env)", name, strings.Join(args, " "))
	if b.checkDryRun("[DRY RUN] Would execute: %s %s (with env)", name, strings.Join(args, " ")) {
		return nil
	}

	cmd := exec.CommandContext(ctx, name, args...) // #nosec G204 -- executor library, callers control command name
	b.configureCommand(cmd)

	// Add additional environment variables
	cmd.Env = append(cmd.Env, env...)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return CommandError(name, args, err)
	}
	return nil
}

// ExecuteInDir runs a command in the specified directory
func (b *Base) ExecuteInDir(ctx context.Context, dir, name string, args ...string) error {
	b.logVerbose("➤ [%s] %s %s", dir, name, strings.Join(args, " "))
	if b.checkDryRun("[DRY RUN] Would execute in %s: %s %s", dir, name, strings.Join(args, " ")) {
		return nil
	}

	cmd := exec.CommandContext(ctx, name, args...) // #nosec G204 -- executor library, callers control command name
	b.configureCommand(cmd)
	cmd.Dir = dir // Override working directory

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return CommandErrorInDir(name, args, dir, err)
	}
	return nil
}

// ExecuteOutputInDir runs a command in the specified directory and returns output
func (b *Base) ExecuteOutputInDir(ctx context.Context, dir, name string, args ...string) (string, error) {
	b.logVerbose("➤ [%s] %s %s", dir, name, strings.Join(args, " "))
	if b.checkDryRun("[DRY RUN] Would execute in %s: %s %s", dir, name, strings.Join(args, " ")) {
		return "", nil
	}

	cmd := exec.CommandContext(ctx, name, args...) // #nosec G204 -- executor library, callers control command name
	b.configureCommand(cmd)
	cmd.Dir = dir // Override working directory

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", CommandErrorInDirWithOutput(name, args, dir, err, string(output))
	}
	return string(output), nil
}

// ExecuteStreaming runs a command with custom stdout/stderr
func (b *Base) ExecuteStreaming(ctx context.Context, stdout, stderr io.Writer, name string, args ...string) error {
	b.logVerbose("➤ %s %s", name, strings.Join(args, " "))
	if b.checkDryRun("[DRY RUN] Would execute: %s %s", name, strings.Join(args, " ")) {
		return nil
	}

	cmd := exec.CommandContext(ctx, name, args...) // #nosec G204,G702 -- executor library, callers control command name
	b.configureCommand(cmd)

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err := cmd.Run(); err != nil {
		return CommandError(name, args, err)
	}
	return nil
}

// configureCommand applies common configuration to a command
func (b *Base) configureCommand(cmd *exec.Cmd) {
	if b.WorkingDir != "" {
		cmd.Dir = b.WorkingDir
	}

	// Start with current environment
	cmd.Env = os.Environ()

	// Add any additional environment variables
	if len(b.Env) > 0 {
		cmd.Env = append(cmd.Env, b.Env...)
	}
}

// Ensure Base implements all interfaces
var (
	_ Executor          = (*Base)(nil)
	_ ExecutorWithEnv   = (*Base)(nil)
	_ ExecutorWithDir   = (*Base)(nil)
	_ StreamingExecutor = (*Base)(nil)
	_ FullExecutor      = (*Base)(nil)
)
