// Package errors provides comprehensive error logging capabilities
package errors

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DefaultErrorLogger implements the ErrorLogger interface
type DefaultErrorLogger struct {
	mu        sync.RWMutex
	formatter ErrorFormatter
	logLevel  Severity
	output    io.Writer
	logger    *log.Logger
	enabled   bool
}

// NewErrorLogger creates a new error logger with default settings
func NewErrorLogger() ErrorLogger {
	return NewErrorLoggerWithOptions(ErrorLoggerOptions{
		Output:   os.Stderr,
		LogLevel: SeverityInfo,
		Enabled:  true,
	})
}

// NewErrorLoggerWithOptions creates an error logger with custom options
func NewErrorLoggerWithOptions(options ErrorLoggerOptions) ErrorLogger {
	logger := &DefaultErrorLogger{
		output:   options.Output,
		logLevel: options.LogLevel,
		enabled:  options.Enabled,
	}

	if logger.output == nil {
		logger.output = os.Stderr
	}

	logger.logger = log.New(logger.output, "", 0)

	if options.Formatter != nil {
		logger.formatter = options.Formatter
	} else {
		logger.formatter = NewFormatter()
	}

	return logger
}

// LogError logs a standard error
func (l *DefaultErrorLogger) LogError(err error) {
	l.LogErrorWithContext(context.Background(), err)
}

// LogErrorWithContext logs an error with context
func (l *DefaultErrorLogger) LogErrorWithContext(ctx context.Context, err error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if !l.enabled || err == nil {
		return
	}

	// Convert to MageError if possible
	var mageErr MageError
	if errors.As(err, &mageErr) {
		l.logMageError(ctx, mageErr)
	} else {
		// Create a MageError wrapper
		builder := NewErrorBuilder()
		builder.WithMessage("%s", err.Error())
		builder.WithCode(ErrUnknown)
		builder.WithSeverity(SeverityError)
		builder.WithField("logger", "standard_error")
		mageErr := builder.Build()
		l.logMageError(ctx, mageErr)
	}
}

// LogMageError logs a MageError directly
func (l *DefaultErrorLogger) LogMageError(err MageError) {
	l.LogMageErrorWithContext(context.Background(), err)
}

// LogMageErrorWithContext logs a MageError with context
func (l *DefaultErrorLogger) LogMageErrorWithContext(ctx context.Context, err MageError) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if !l.enabled || err == nil {
		return
	}

	l.logMageError(ctx, err)
}

// SetLogLevel sets the minimum log level
func (l *DefaultErrorLogger) SetLogLevel(severity Severity) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logLevel = severity
}

// SetFormatter sets the error formatter
func (l *DefaultErrorLogger) SetFormatter(formatter ErrorFormatter) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.formatter = formatter
}

// SetOutput sets the output writer
func (l *DefaultErrorLogger) SetOutput(output io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.output = output
	l.logger = log.New(output, "", 0)
}

// SetEnabled enables or disables logging
func (l *DefaultErrorLogger) SetEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.enabled = enabled
}

// GetLogLevel returns the current log level
func (l *DefaultErrorLogger) GetLogLevel() Severity {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.logLevel
}

// IsEnabled returns whether logging is enabled
func (l *DefaultErrorLogger) IsEnabled() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.enabled
}

// logMageError performs the actual logging
func (l *DefaultErrorLogger) logMageError(ctx context.Context, err MageError) {
	// Check if error severity meets the log level threshold
	if err.Severity() < l.logLevel {
		return
	}

	// Format the error
	formatted := l.formatter.Format(err)

	// Add timestamp and context information
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logLine := fmt.Sprintf("[%s] [%s] %s", timestamp, err.Severity().String(), formatted)

	// Add context information if available
	if requestID := getRequestIDFromContext(ctx); requestID != "" {
		logLine = fmt.Sprintf("[req:%s] %s", requestID, logLine)
	}

	// Log the formatted error
	l.logger.Println(logLine)
}

// getRequestIDFromContext extracts request ID from context
func getRequestIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	if value := ctx.Value("request_id"); value != nil {
		if requestID, ok := value.(string); ok {
			return requestID
		}
	}

	if value := ctx.Value("requestId"); value != nil {
		if requestID, ok := value.(string); ok {
			return requestID
		}
	}

	return ""
}

// ErrorLoggerOptions holds configuration options for error loggers
type ErrorLoggerOptions struct {
	Output    io.Writer
	LogLevel  Severity
	Formatter ErrorFormatter
	Enabled   bool
}

// StructuredErrorLogger provides structured logging capabilities
type StructuredErrorLogger struct {
	*DefaultErrorLogger
	fields map[string]interface{}
}

// NewStructuredErrorLogger creates a new structured error logger
func NewStructuredErrorLogger() *StructuredErrorLogger {
	return &StructuredErrorLogger{
		DefaultErrorLogger: func() *DefaultErrorLogger {
			logger := NewErrorLogger()
			if defaultLogger, ok := logger.(*DefaultErrorLogger); ok {
				return defaultLogger
			}
			return nil
		}(),
		fields: make(map[string]interface{}),
	}
}

// WithField adds a field to the structured logger
func (l *StructuredErrorLogger) WithField(key string, value interface{}) *StructuredErrorLogger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	newFields[key] = value

	return &StructuredErrorLogger{
		DefaultErrorLogger: l.DefaultErrorLogger,
		fields:             newFields,
	}
}

// WithFields adds multiple fields to the structured logger
func (l *StructuredErrorLogger) WithFields(fields map[string]interface{}) *StructuredErrorLogger {
	l.mu.Lock()
	defer l.mu.Unlock()

	newFields := make(map[string]interface{})
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	return &StructuredErrorLogger{
		DefaultErrorLogger: l.DefaultErrorLogger,
		fields:             newFields,
	}
}

// LogMageErrorWithContext logs a MageError with additional structured fields
func (l *StructuredErrorLogger) LogMageErrorWithContext(ctx context.Context, err MageError) {
	if !l.enabled || err == nil {
		return
	}

	// Create a new error with additional fields
	builder := NewErrorBuilder()
	builder.WithMessage("%s", err.Error())
	builder.WithCode(err.Code())
	builder.WithSeverity(err.Severity())
	for key, value := range l.fields {
		builder.WithField(key, value)
	}

	enhancedErr := builder.Build()
	l.DefaultErrorLogger.LogMageErrorWithContext(ctx, enhancedErr)
}

// FileRotatingLogger provides file-based logging with rotation
type FileRotatingLogger struct {
	*DefaultErrorLogger
	filepath    string
	maxSize     int64 // Maximum size in bytes
	maxFiles    int   // Maximum number of files to keep
	currentSize int64
}

// NewFileRotatingLogger creates a new file-based logger with rotation
func NewFileRotatingLogger(logPath string, maxSize int64, maxFiles int) (*FileRotatingLogger, error) {
	// Validate and clean the file path to prevent directory traversal
	cleanPath := filepath.Clean(logPath)
	if strings.Contains(cleanPath, "..") {
		return nil, fmt.Errorf("invalid file path: path traversal detected")
	}

	file, err := os.OpenFile(cleanPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Get current file size
	info, err := file.Stat()
	if err != nil {
		if closeErr := file.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to close file: %v\n", closeErr)
		}
		return nil, fmt.Errorf("failed to stat log file: %w", err)
	}

	return &FileRotatingLogger{
		DefaultErrorLogger: func() *DefaultErrorLogger {
			logger := NewErrorLoggerWithOptions(ErrorLoggerOptions{
				Output:  file,
				Enabled: true,
			})
			if defaultLogger, ok := logger.(*DefaultErrorLogger); ok {
				return defaultLogger
			}
			return nil
		}(),
		filepath:    cleanPath,
		maxSize:     maxSize,
		maxFiles:    maxFiles,
		currentSize: info.Size(),
	}, nil
}

// MockErrorLogger implements ErrorLogger for testing
type MockErrorLogger struct {
	LogErrorCalls            []error
	LogErrorWithContextCalls []MockLogErrorWithContextCall
	LogMageErrorCalls        []MageError
	SetLogLevelCalls         []Severity
	SetFormatterCalls        []ErrorFormatter
	ShouldError              bool
	LogLevel                 Severity
	Enabled                  bool
}

type MockLogErrorWithContextCall struct {
	Error error
}

func NewMockErrorLogger() *MockErrorLogger {
	return &MockErrorLogger{
		LogErrorCalls:            make([]error, 0),
		LogErrorWithContextCalls: make([]MockLogErrorWithContextCall, 0),
		LogMageErrorCalls:        make([]MageError, 0),
		SetLogLevelCalls:         make([]Severity, 0),
		SetFormatterCalls:        make([]ErrorFormatter, 0),
		LogLevel:                 SeverityInfo,
		Enabled:                  true,
	}
}

func (m *MockErrorLogger) LogError(err error) {
	m.LogErrorCalls = append(m.LogErrorCalls, err)
}

func (m *MockErrorLogger) LogErrorWithContext(ctx context.Context, err error) {
	m.LogErrorWithContextCalls = append(m.LogErrorWithContextCalls, MockLogErrorWithContextCall{
		Error: err,
	})
}

func (m *MockErrorLogger) LogMageError(err MageError) {
	m.LogMageErrorCalls = append(m.LogMageErrorCalls, err)
}

func (m *MockErrorLogger) SetLogLevel(severity Severity) {
	m.SetLogLevelCalls = append(m.SetLogLevelCalls, severity)
	m.LogLevel = severity
}

func (m *MockErrorLogger) SetFormatter(formatter ErrorFormatter) {
	m.SetFormatterCalls = append(m.SetFormatterCalls, formatter)
}
