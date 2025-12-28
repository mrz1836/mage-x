package exec

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"

	"github.com/mrz1836/mage-x/pkg/log"
)

// AuditEvent represents a command execution audit event
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

// AuditLogger interface for logging audit events
type AuditLogger interface {
	LogEvent(event AuditEvent) error
}

// AuditingExecutor wraps an executor with audit logging
type AuditingExecutor struct {
	wrapped    Executor
	logger     AuditLogger
	workingDir string
	dryRun     bool
	metadata   map[string]string
}

// AuditingOption configures an AuditingExecutor
type AuditingOption func(*AuditingExecutor)

// WithAuditLogger sets a custom audit logger
func WithAuditLogger(logger AuditLogger) AuditingOption {
	return func(a *AuditingExecutor) {
		a.logger = logger
	}
}

// WithAuditWorkingDir sets the working directory for audit events
func WithAuditWorkingDir(dir string) AuditingOption {
	return func(a *AuditingExecutor) {
		a.workingDir = dir
	}
}

// WithAuditDryRun sets the dry run flag for metadata
func WithAuditDryRun(dryRun bool) AuditingOption {
	return func(a *AuditingExecutor) {
		a.dryRun = dryRun
	}
}

// WithAuditMetadata adds custom metadata to all audit events
func WithAuditMetadata(key, value string) AuditingOption {
	return func(a *AuditingExecutor) {
		if a.metadata == nil {
			a.metadata = make(map[string]string)
		}
		a.metadata[key] = value
	}
}

// NewAuditingExecutor creates a new auditing executor wrapper
func NewAuditingExecutor(wrapped Executor, opts ...AuditingOption) *AuditingExecutor {
	a := &AuditingExecutor{
		wrapped:  wrapped,
		metadata: make(map[string]string),
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Execute runs a command with audit logging
func (a *AuditingExecutor) Execute(ctx context.Context, name string, args ...string) error {
	startTime := time.Now()
	err := a.wrapped.Execute(ctx, name, args...)
	a.logAuditEvent(name, args, startTime, err)
	return err
}

// ExecuteOutput runs a command with audit logging and returns output
func (a *AuditingExecutor) ExecuteOutput(ctx context.Context, name string, args ...string) (string, error) {
	startTime := time.Now()
	output, err := a.wrapped.ExecuteOutput(ctx, name, args...)
	a.logAuditEvent(name, args, startTime, err)
	return output, err
}

// logAuditEvent creates and logs an audit event
func (a *AuditingExecutor) logAuditEvent(command string, args []string, startTime time.Time, err error) {
	// Skip if no logger configured
	if a.logger == nil {
		return
	}

	duration := time.Since(startTime)
	success := err == nil
	exitCode := 0

	if err != nil {
		exitError := &exec.ExitError{}
		if errors.As(err, &exitError) {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	// Get current user
	currentUser := "unknown"
	if usr, err := user.Current(); err == nil {
		currentUser = usr.Username
	}

	// Get working directory
	workingDir := a.workingDir
	if workingDir == "" {
		var dirErr error
		workingDir, dirErr = os.Getwd()
		if dirErr != nil {
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

	// Create metadata
	metadata := make(map[string]string)
	for k, v := range a.metadata {
		metadata[k] = v
	}
	metadata["executor_type"] = "AuditingExecutor"
	metadata["dry_run"] = fmt.Sprintf("%v", a.dryRun)

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
		Metadata:    metadata,
	}

	// Log the event (ignore errors to avoid breaking command execution)
	if logErr := a.logger.LogEvent(event); logErr != nil {
		log.Warn("failed to log audit event: %v", logErr)
	}
}

// DefaultAuditLogger is a simple logger that logs to stderr
type DefaultAuditLogger struct{}

// LogEvent logs an audit event to stderr
func (d *DefaultAuditLogger) LogEvent(event AuditEvent) error {
	fmt.Fprintf(os.Stderr, "[AUDIT] %s: %s %s (duration=%s, exit=%d, success=%v)\n",
		event.Timestamp.Format(time.RFC3339),
		event.Command,
		strings.Join(event.Args, " "),
		event.Duration,
		event.ExitCode,
		event.Success,
	)
	return nil
}

// Ensure AuditingExecutor implements Executor
var _ Executor = (*AuditingExecutor)(nil)
