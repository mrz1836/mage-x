package log

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// copyFields returns a deep copy of the fields map
func copyFields(src Fields) Fields {
	if src == nil {
		return make(Fields)
	}
	dst := make(Fields, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

// CLIAdapter wraps the utils.Logger to implement CLILogger
type CLIAdapter struct {
	mu       sync.RWMutex
	level    Level
	prefix   string
	fields   Fields
	useColor bool
	output   io.Writer
}

// NewCLIAdapter creates a new CLI logger adapter
func NewCLIAdapter() *CLIAdapter {
	return &CLIAdapter{
		level:    LevelInfo,
		useColor: shouldUseColor(),
		output:   os.Stdout,
		fields:   make(Fields),
	}
}

// Debug logs a debug message
func (a *CLIAdapter) Debug(format string, args ...interface{}) {
	a.log(LevelDebug, format, args...)
}

// Info logs an informational message
func (a *CLIAdapter) Info(format string, args ...interface{}) {
	a.log(LevelInfo, format, args...)
}

// Warn logs a warning message
func (a *CLIAdapter) Warn(format string, args ...interface{}) {
	a.log(LevelWarn, format, args...)
}

// Error logs an error message
func (a *CLIAdapter) Error(format string, args ...interface{}) {
	a.log(LevelError, format, args...)
}

// Success logs a success message with emoji
func (a *CLIAdapter) Success(format string, args ...interface{}) {
	a.logWithEmoji(LevelInfo, "✅", format, args...)
}

// Fail logs a failure message with emoji
func (a *CLIAdapter) Fail(format string, args ...interface{}) {
	a.logWithEmoji(LevelError, "❌", format, args...)
}

// Header prints a formatted header
func (a *CLIAdapter) Header(text string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	line := "============================================================"
	if a.useColor {
		colorBlue := "\033[34m"
		colorBold := "\033[1m"
		colorReset := "\033[0m"
		//nolint:errcheck // Log output errors are intentionally ignored
		fmt.Fprintf(a.output, "\n%s%s%s\n", colorBold+colorBlue, line, colorReset)
		//nolint:errcheck // Log output errors are intentionally ignored
		fmt.Fprintf(a.output, "%s%s %s %s%s\n", colorBold+colorBlue, "===", text, "===", colorReset)
		//nolint:errcheck // Log output errors are intentionally ignored
		fmt.Fprintf(a.output, "%s%s%s\n\n", colorBold+colorBlue, line, colorReset)
	} else {
		//nolint:errcheck // Log output errors are intentionally ignored
		fmt.Fprintf(a.output, "\n%s\n=== %s ===\n%s\n\n", line, text, line)
	}
}

// StartSpinner starts a progress spinner (no-op in this adapter)
func (a *CLIAdapter) StartSpinner(_ string) {
	// Spinner functionality delegated to utils.Logger
}

// StopSpinner stops the current spinner (no-op in this adapter)
func (a *CLIAdapter) StopSpinner() {
	// Spinner functionality delegated to utils.Logger
}

// UpdateSpinner updates the spinner message (no-op in this adapter)
func (a *CLIAdapter) UpdateSpinner(_ string) {
	// Spinner functionality delegated to utils.Logger
}

// SetLevel sets the minimum log level
func (a *CLIAdapter) SetLevel(level Level) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.level = level
}

// GetLevel returns the current log level
func (a *CLIAdapter) GetLevel() Level {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.level
}

// SetOutput sets the output writer
func (a *CLIAdapter) SetOutput(w io.Writer) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.output = w
}

// SetColorEnabled enables or disables color output
func (a *CLIAdapter) SetColorEnabled(enabled bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.useColor = enabled
}

// WithPrefix creates a new logger with a prefix
func (a *CLIAdapter) WithPrefix(prefix string) Logger {
	a.mu.RLock()
	defer a.mu.RUnlock()

	newAdapter := &CLIAdapter{
		level:    a.level,
		prefix:   prefix,
		useColor: a.useColor,
		output:   a.output,
		fields:   copyFields(a.fields),
	}
	return newAdapter
}

// WithField creates a new logger with an additional field
func (a *CLIAdapter) WithField(key string, value interface{}) Logger {
	a.mu.RLock()
	defer a.mu.RUnlock()

	newAdapter := &CLIAdapter{
		level:    a.level,
		prefix:   a.prefix,
		useColor: a.useColor,
		output:   a.output,
		fields:   copyFields(a.fields),
	}
	newAdapter.fields[key] = value
	return newAdapter
}

// WithFields creates a new logger with additional fields
func (a *CLIAdapter) WithFields(fields Fields) Logger {
	a.mu.RLock()
	defer a.mu.RUnlock()

	newAdapter := &CLIAdapter{
		level:    a.level,
		prefix:   a.prefix,
		useColor: a.useColor,
		output:   a.output,
		fields:   copyFields(a.fields),
	}
	for k, v := range fields {
		newAdapter.fields[k] = v
	}
	return newAdapter
}

// log is the internal logging function
func (a *CLIAdapter) log(level Level, format string, args ...interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if level < a.level {
		return
	}

	msg := fmt.Sprintf(format, args...)
	if a.prefix != "" {
		msg = fmt.Sprintf("[%s] %s", a.prefix, msg)
	}

	timestamp := time.Now().Format("15:04:05")

	var levelStr, color string
	colorReset := "\033[0m"
	switch level {
	case LevelDebug:
		levelStr = "DEBUG"
		color = "\033[90m" // gray
	case LevelInfo:
		levelStr = "INFO"
		color = "\033[34m" // blue
	case LevelWarn:
		levelStr = "WARN"
		color = "\033[33m" // yellow
	case LevelError:
		levelStr = "ERROR"
		color = "\033[31m" // red
	case LevelFatal:
		levelStr = "FATAL"
		color = "\033[31m" // red
	}

	if a.useColor {
		//nolint:errcheck // Log output errors are intentionally ignored
		fmt.Fprintf(a.output, "%s%s [%s]%s %s\n", color, timestamp, levelStr, colorReset, msg)
	} else {
		//nolint:errcheck // Log output errors are intentionally ignored
		fmt.Fprintf(a.output, "%s [%s] %s\n", timestamp, levelStr, msg)
	}
}

// logWithEmoji logs a message with an emoji prefix
func (a *CLIAdapter) logWithEmoji(level Level, emoji, format string, args ...interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if level < a.level {
		return
	}

	msg := fmt.Sprintf(format, args...)
	if a.prefix != "" {
		msg = fmt.Sprintf("[%s] %s", a.prefix, msg)
	}

	var color string
	colorReset := "\033[0m"
	//nolint:exhaustive // LevelDebug and LevelFatal use default color
	switch level {
	case LevelInfo:
		color = "\033[32m" // green
	case LevelWarn:
		color = "\033[33m" // yellow
	case LevelError, LevelFatal:
		color = "\033[31m" // red
	default:
		color = colorReset
	}

	if a.useColor {
		//nolint:errcheck // Log output errors are intentionally ignored
		fmt.Fprintf(a.output, "%s %s%s%s\n", emoji, color, msg, colorReset)
	} else {
		//nolint:errcheck // Log output errors are intentionally ignored
		fmt.Fprintf(a.output, "%s %s\n", emoji, msg)
	}
}

// StructuredAdapter provides structured logging capabilities
type StructuredAdapter struct {
	mu     sync.RWMutex
	level  Level
	prefix string
	fields Fields
	output io.Writer
}

// NewStructuredAdapter creates a new structured logger adapter
func NewStructuredAdapter() *StructuredAdapter {
	return &StructuredAdapter{
		level:  LevelInfo,
		output: os.Stderr,
		fields: make(Fields),
	}
}

// Debug logs a debug message
func (a *StructuredAdapter) Debug(format string, args ...interface{}) {
	a.log(LevelDebug, format, args...)
}

// Info logs an informational message
func (a *StructuredAdapter) Info(format string, args ...interface{}) {
	a.log(LevelInfo, format, args...)
}

// Warn logs a warning message
func (a *StructuredAdapter) Warn(format string, args ...interface{}) {
	a.log(LevelWarn, format, args...)
}

// Error logs an error message
func (a *StructuredAdapter) Error(format string, args ...interface{}) {
	a.log(LevelError, format, args...)
}

// DebugContext logs a debug message with context
func (a *StructuredAdapter) DebugContext(ctx context.Context, format string, args ...interface{}) {
	a.logWithContext(ctx, LevelDebug, format, args...)
}

// InfoContext logs an info message with context
func (a *StructuredAdapter) InfoContext(ctx context.Context, format string, args ...interface{}) {
	a.logWithContext(ctx, LevelInfo, format, args...)
}

// WarnContext logs a warning message with context
func (a *StructuredAdapter) WarnContext(ctx context.Context, format string, args ...interface{}) {
	a.logWithContext(ctx, LevelWarn, format, args...)
}

// ErrorContext logs an error message with context
func (a *StructuredAdapter) ErrorContext(ctx context.Context, format string, args ...interface{}) {
	a.logWithContext(ctx, LevelError, format, args...)
}

// SetLevel sets the minimum log level
func (a *StructuredAdapter) SetLevel(level Level) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.level = level
}

// GetLevel returns the current log level
func (a *StructuredAdapter) GetLevel() Level {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.level
}

// SetOutput sets the output writer
func (a *StructuredAdapter) SetOutput(w io.Writer) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.output = w
}

// WithPrefix creates a new logger with a prefix
func (a *StructuredAdapter) WithPrefix(prefix string) Logger {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return &StructuredAdapter{
		level:  a.level,
		prefix: prefix,
		output: a.output,
		fields: copyFields(a.fields),
	}
}

// WithField creates a new logger with an additional field
func (a *StructuredAdapter) WithField(key string, value interface{}) Logger {
	a.mu.RLock()
	defer a.mu.RUnlock()

	newAdapter := &StructuredAdapter{
		level:  a.level,
		prefix: a.prefix,
		output: a.output,
		fields: copyFields(a.fields),
	}
	newAdapter.fields[key] = value
	return newAdapter
}

// WithFields creates a new logger with additional fields
func (a *StructuredAdapter) WithFields(fields Fields) Logger {
	a.mu.RLock()
	defer a.mu.RUnlock()

	newAdapter := &StructuredAdapter{
		level:  a.level,
		prefix: a.prefix,
		output: a.output,
		fields: copyFields(a.fields),
	}
	for k, v := range fields {
		newAdapter.fields[k] = v
	}
	return newAdapter
}

// log is the internal logging function
func (a *StructuredAdapter) log(level Level, format string, args ...interface{}) {
	a.logWithContext(context.Background(), level, format, args...)
}

// logWithContext logs with context information
func (a *StructuredAdapter) logWithContext(ctx context.Context, level Level, format string, args ...interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if level < a.level {
		return
	}

	msg := fmt.Sprintf(format, args...)
	if a.prefix != "" {
		msg = fmt.Sprintf("[%s] %s", a.prefix, msg)
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// Build structured log line
	logLine := fmt.Sprintf("[%s] [%s] %s", timestamp, level.String(), msg)

	// Add context information if available
	if requestID := GetRequestIDFromContext(ctx); requestID != "" {
		logLine = fmt.Sprintf("[req:%s] %s", requestID, logLine)
	}

	// Add fields if present
	if len(a.fields) > 0 {
		logLine += " {"
		first := true
		for k, v := range a.fields {
			if !first {
				logLine += ", "
			}
			logLine += fmt.Sprintf("%s=%v", k, v)
			first = false
		}
		logLine += "}"
	}

	//nolint:errcheck // Log output errors are intentionally ignored
	fmt.Fprintln(a.output, logLine)
}

// shouldUseColor determines if color output should be enabled
func shouldUseColor() bool {
	// Disable color in CI environments
	if os.Getenv("CI") != "" {
		return false
	}

	// Disable color if NO_COLOR is set
	if os.Getenv("NO_COLOR") != "" {
		return false
	}

	// Disable color if not a terminal
	fileInfo, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	if (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		return false
	}

	return true
}

// Ensure adapters implement interfaces
var (
	_ CLILogger        = (*CLIAdapter)(nil)
	_ Logger           = (*CLIAdapter)(nil)
	_ StructuredLogger = (*StructuredAdapter)(nil)
	_ Logger           = (*StructuredAdapter)(nil)
)
