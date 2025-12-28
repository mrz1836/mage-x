package log

import (
	"context"
	"io"
	"sync"
)

// Logger is the unified logging interface.
// All logging in mage-x should go through this interface.
type Logger interface {
	// Basic logging methods
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Warn(format string, args ...interface{})
	Error(format string, args ...interface{})

	// Level control
	SetLevel(level Level)
	GetLevel() Level

	// Output control
	SetOutput(w io.Writer)

	// Create a child logger with a prefix
	WithPrefix(prefix string) Logger

	// Create a child logger with a field
	WithField(key string, value interface{}) Logger

	// Create a child logger with multiple fields
	WithFields(fields Fields) Logger
}

// CLILogger extends Logger with CLI-specific features.
// Use this for user-facing output that needs colors, spinners, etc.
type CLILogger interface {
	Logger

	// CLI-specific methods
	Success(format string, args ...interface{})
	Fail(format string, args ...interface{})
	Header(text string)

	// Spinner control
	StartSpinner(message string)
	StopSpinner()
	UpdateSpinner(message string)

	// Color control
	SetColorEnabled(enabled bool)
}

// StructuredLogger extends Logger with structured logging features.
// Use this for machine-readable logs and error tracking.
type StructuredLogger interface {
	Logger

	// Context-aware logging
	DebugContext(ctx context.Context, format string, args ...interface{})
	InfoContext(ctx context.Context, format string, args ...interface{})
	WarnContext(ctx context.Context, format string, args ...interface{})
	ErrorContext(ctx context.Context, format string, args ...interface{})
}

// Fields is a map of key-value pairs for structured logging
type Fields map[string]interface{}

// loggerManager manages the default loggers
type loggerManager struct {
	mu         sync.RWMutex
	cli        CLILogger
	structured StructuredLogger
	level      Level
}

// manager is the package-level singleton
//
//nolint:gochecknoglobals // Required for package-level singleton pattern
var manager = &loggerManager{
	level: LevelInfo,
}

// SetDefault sets the default CLI logger
func SetDefault(l CLILogger) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	manager.cli = l
}

// SetStructured sets the default structured logger
func SetStructured(l StructuredLogger) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	manager.structured = l
}

// Default returns the default CLI logger
func Default() CLILogger {
	ensureInitialized()
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return manager.cli
}

// Structured returns the default structured logger
func Structured() StructuredLogger {
	ensureInitialized()
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return manager.structured
}

// SetLevel sets the global log level
func SetLevel(level Level) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	manager.level = level

	// Propagate to loggers if set
	if manager.cli != nil {
		manager.cli.SetLevel(level)
	}
	if manager.structured != nil {
		manager.structured.SetLevel(level)
	}
}

// GetLevel returns the global log level
func GetLevel() Level {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return manager.level
}

// Package-level convenience functions that use the default CLI logger

// Debug logs a debug message
//
//nolint:goprintffuncname // Domain-specific API for cleaner logging interface
func Debug(format string, args ...interface{}) {
	if l := Default(); l != nil {
		l.Debug(format, args...)
	}
}

// Info logs an informational message
//
//nolint:goprintffuncname // Domain-specific API for cleaner logging interface
func Info(format string, args ...interface{}) {
	if l := Default(); l != nil {
		l.Info(format, args...)
	}
}

// Warn logs a warning message
//
//nolint:goprintffuncname // Domain-specific API for cleaner logging interface
func Warn(format string, args ...interface{}) {
	if l := Default(); l != nil {
		l.Warn(format, args...)
	}
}

// Error logs an error message
//
//nolint:goprintffuncname // Domain-specific API for cleaner logging interface
func Error(format string, args ...interface{}) {
	if l := Default(); l != nil {
		l.Error(format, args...)
	}
}

// Success logs a success message
//
//nolint:goprintffuncname // Domain-specific API for cleaner logging interface
func Success(format string, args ...interface{}) {
	if l := Default(); l != nil {
		l.Success(format, args...)
	}
}

// Fail logs a failure message
//
//nolint:goprintffuncname // Domain-specific API for cleaner logging interface
func Fail(format string, args ...interface{}) {
	if l := Default(); l != nil {
		l.Fail(format, args...)
	}
}

// Header prints a formatted header
func Header(text string) {
	if l := Default(); l != nil {
		l.Header(text)
	}
}

// StartSpinner starts a progress spinner
func StartSpinner(message string) {
	if l := Default(); l != nil {
		l.StartSpinner(message)
	}
}

// StopSpinner stops the current spinner
func StopSpinner() {
	if l := Default(); l != nil {
		l.StopSpinner()
	}
}

// UpdateSpinner updates the spinner message
func UpdateSpinner(message string) {
	if l := Default(); l != nil {
		l.UpdateSpinner(message)
	}
}

// WithField creates a logger with an additional field
func WithField(key string, value interface{}) Logger {
	if l := Default(); l != nil {
		return l.WithField(key, value)
	}
	return nil
}

// WithFields creates a logger with additional fields
func WithFields(fields Fields) Logger {
	if l := Default(); l != nil {
		return l.WithFields(fields)
	}
	return nil
}

// WithPrefix creates a logger with a prefix
func WithPrefix(prefix string) Logger {
	if l := Default(); l != nil {
		return l.WithPrefix(prefix)
	}
	return nil
}
