package errors

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
)

// Test sentinel errors for err113 compliance.
var (
	errLoggerTest            = errors.New("test error message")
	errLoggerTestGeneric     = errors.New("test error")
	errLoggerFirst           = errors.New("first error")
	errLoggerSecond          = errors.New("second error")
	errLoggerTestPlaceholder = errors.New("test")
	errLoggerConcurrent      = errors.New("concurrent error")
	errLoggerTestTracking    = errors.New("test error")
)

func TestNewErrorLogger(t *testing.T) {
	t.Parallel()

	logger := NewErrorLogger()
	if logger == nil {
		t.Fatal("NewErrorLogger() returned nil")
	}

	defaultLogger, ok := logger.(*DefaultErrorLogger)
	if !ok {
		t.Fatal("NewErrorLogger() should return *DefaultErrorLogger")
	}

	if !defaultLogger.enabled {
		t.Error("NewErrorLogger() should be enabled by default")
	}

	if defaultLogger.logLevel != SeverityInfo {
		t.Errorf("NewErrorLogger() logLevel = %v, want %v", defaultLogger.logLevel, SeverityInfo)
	}
}

func TestNewErrorLoggerWithOptions(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := NewErrorLoggerWithOptions(ErrorLoggerOptions{
		Output:   &buf,
		LogLevel: SeverityWarning,
		Enabled:  true,
	})

	defaultLogger, ok := logger.(*DefaultErrorLogger)
	if !ok {
		t.Fatal("NewErrorLoggerWithOptions() should return *DefaultErrorLogger")
	}

	if defaultLogger.logLevel != SeverityWarning {
		t.Errorf("logLevel = %v, want %v", defaultLogger.logLevel, SeverityWarning)
	}

	if defaultLogger.output != &buf {
		t.Error("output should match provided writer")
	}
}

func TestNewErrorLoggerWithOptions_NilOutput(t *testing.T) {
	t.Parallel()

	logger := NewErrorLoggerWithOptions(ErrorLoggerOptions{
		Output:   nil,
		LogLevel: SeverityInfo,
		Enabled:  true,
	})

	defaultLogger, ok := logger.(*DefaultErrorLogger)
	if !ok {
		t.Fatal("NewErrorLoggerWithOptions() should return *DefaultErrorLogger")
	}

	if defaultLogger.output == nil {
		t.Error("output should default to os.Stderr when nil provided")
	}
}

func TestDefaultErrorLogger_LogError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := NewErrorLoggerWithOptions(ErrorLoggerOptions{
		Output:   &buf,
		LogLevel: SeverityInfo,
		Enabled:  true,
	})

	logger.LogError(errLoggerTest)

	output := buf.String()
	if !strings.Contains(output, "test error message") {
		t.Errorf("LogError() output should contain error message, got: %s", output)
	}
	if !strings.Contains(output, "ERROR") {
		t.Errorf("LogError() output should contain ERROR severity, got: %s", output)
	}
}

func TestDefaultErrorLogger_LogError_Nil(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := NewErrorLoggerWithOptions(ErrorLoggerOptions{
		Output:   &buf,
		LogLevel: SeverityInfo,
		Enabled:  true,
	})

	logger.LogError(nil)

	if buf.Len() > 0 {
		t.Errorf("LogError(nil) should not produce output, got: %s", buf.String())
	}
}

func TestDefaultErrorLogger_LogError_Disabled(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := NewErrorLoggerWithOptions(ErrorLoggerOptions{
		Output:   &buf,
		LogLevel: SeverityInfo,
		Enabled:  false,
	})

	logger.LogError(errLoggerTestGeneric)

	if buf.Len() > 0 {
		t.Errorf("LogError() with disabled logger should not produce output, got: %s", buf.String())
	}
}

func TestDefaultErrorLogger_LogErrorWithContext(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := NewErrorLoggerWithOptions(ErrorLoggerOptions{
		Output:   &buf,
		LogLevel: SeverityInfo,
		Enabled:  true,
	})

	//nolint:staticcheck // SA1029: Using string key intentionally
	ctx := context.WithValue(context.Background(), "request_id", "req-123")
	logger.LogErrorWithContext(ctx, errTestGeneric)

	output := buf.String()
	if !strings.Contains(output, "[req:req-123]") {
		t.Errorf("LogErrorWithContext() should include request ID, got: %s", output)
	}
}

func TestDefaultErrorLogger_LogMageError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := NewErrorLoggerWithOptions(ErrorLoggerOptions{
		Output:   &buf,
		LogLevel: SeverityInfo,
		Enabled:  true,
	})

	mageErr := NewMageError("mage error message")
	logger.LogMageError(mageErr)

	output := buf.String()
	if !strings.Contains(output, "mage error message") {
		t.Errorf("LogMageError() should contain error message, got: %s", output)
	}
}

func TestDefaultErrorLogger_SeverityFiltering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		logLevel      Severity
		errorSeverity Severity
		shouldLog     bool
	}{
		{
			name:          "info level logs error",
			logLevel:      SeverityInfo,
			errorSeverity: SeverityError,
			shouldLog:     true,
		},
		{
			name:          "error level logs error",
			logLevel:      SeverityError,
			errorSeverity: SeverityError,
			shouldLog:     true,
		},
		{
			name:          "critical level skips error",
			logLevel:      SeverityCritical,
			errorSeverity: SeverityError,
			shouldLog:     false,
		},
		{
			name:          "debug level logs warning",
			logLevel:      SeverityDebug,
			errorSeverity: SeverityWarning,
			shouldLog:     true,
		},
		{
			name:          "warning level skips info",
			logLevel:      SeverityWarning,
			errorSeverity: SeverityInfo,
			shouldLog:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := NewErrorLoggerWithOptions(ErrorLoggerOptions{
				Output:   &buf,
				LogLevel: tt.logLevel,
				Enabled:  true,
			})

			builder := NewErrorBuilder()
			builder.WithMessage("test message")
			builder.WithSeverity(tt.errorSeverity)
			mageErr := builder.Build()

			logger.LogMageError(mageErr)

			hasOutput := buf.Len() > 0
			if hasOutput != tt.shouldLog {
				t.Errorf("severity filtering: shouldLog = %v, got hasOutput = %v", tt.shouldLog, hasOutput)
			}
		})
	}
}

func TestDefaultErrorLogger_SetLogLevel(t *testing.T) {
	t.Parallel()

	logger := NewErrorLogger()
	defaultLogger, ok := logger.(*DefaultErrorLogger)
	if !ok {
		t.Fatal("expected *DefaultErrorLogger")
	}

	logger.SetLogLevel(SeverityCritical)

	if defaultLogger.GetLogLevel() != SeverityCritical {
		t.Errorf("SetLogLevel() did not update level, got %v", defaultLogger.GetLogLevel())
	}
}

func TestDefaultErrorLogger_SetEnabled(t *testing.T) {
	t.Parallel()

	logger := NewErrorLogger()
	defaultLogger, ok := logger.(*DefaultErrorLogger)
	if !ok {
		t.Fatal("expected *DefaultErrorLogger")
	}

	if !defaultLogger.IsEnabled() {
		t.Error("logger should be enabled by default")
	}

	defaultLogger.SetEnabled(false)

	if defaultLogger.IsEnabled() {
		t.Error("SetEnabled(false) should disable logger")
	}
}

func TestDefaultErrorLogger_SetOutput(t *testing.T) {
	t.Parallel()

	var buf1, buf2 bytes.Buffer
	logger := NewErrorLoggerWithOptions(ErrorLoggerOptions{
		Output:   &buf1,
		LogLevel: SeverityInfo,
		Enabled:  true,
	})

	logger.LogError(errLoggerFirst)
	if buf1.Len() == 0 {
		t.Error("first log should write to buf1")
	}

	// Use concrete type to access SetOutput
	defaultLogger, ok := logger.(*DefaultErrorLogger)
	if !ok {
		t.Fatal("expected *DefaultErrorLogger")
	}
	defaultLogger.SetOutput(&buf2)
	logger.LogError(errLoggerSecond)

	if buf2.Len() == 0 {
		t.Error("second log should write to buf2 after SetOutput")
	}
}

func TestDefaultErrorLogger_SetFormatter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := NewErrorLoggerWithOptions(ErrorLoggerOptions{
		Output:   &buf,
		LogLevel: SeverityInfo,
		Enabled:  true,
	})

	// Create a default formatter
	customFormatter := NewFormatter()

	// Use concrete type to access SetFormatter
	defaultLogger, ok := logger.(*DefaultErrorLogger)
	if !ok {
		t.Fatal("expected *DefaultErrorLogger")
	}
	defaultLogger.SetFormatter(customFormatter)
	logger.LogError(errLoggerTestPlaceholder)

	output := buf.String()
	if !strings.Contains(output, "UNKNOWN") { // Error code should be present
		t.Errorf("formatter should be applied, got: %s", output)
	}
}

func TestDefaultErrorLogger_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	logger := NewErrorLoggerWithOptions(ErrorLoggerOptions{
		Output:   &buf,
		LogLevel: SeverityDebug,
		Enabled:  true,
	})

	var wg sync.WaitGroup
	const numGoroutines = 50

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Mix of operations
			logger.LogError(errLoggerConcurrent)
			logger.SetLogLevel(SeverityInfo)
			if defaultLogger, ok := logger.(*DefaultErrorLogger); ok {
				_ = defaultLogger.GetLogLevel()
				_ = defaultLogger.IsEnabled()
			}
		}()
	}

	wg.Wait()
	// If we get here without deadlock or panic, the test passes
}

func TestStructuredErrorLogger_New(t *testing.T) {
	t.Parallel()

	logger := NewStructuredErrorLogger()
	if logger == nil {
		t.Fatal("NewStructuredErrorLogger() returned nil")
	}

	if logger.DefaultErrorLogger == nil {
		t.Error("StructuredErrorLogger should have embedded DefaultErrorLogger")
	}

	if logger.fields == nil {
		t.Error("StructuredErrorLogger fields should be initialized")
	}
}

func TestStructuredErrorLogger_WithField(t *testing.T) {
	t.Parallel()

	logger := NewStructuredErrorLogger()
	withField := logger.WithField("key", "value")

	if withField == nil {
		t.Fatal("WithField() returned nil")
	}

	if withField == logger {
		t.Error("WithField() should return a new logger instance")
	}

	if len(withField.fields) != 1 {
		t.Errorf("WithField() should add one field, got %d", len(withField.fields))
	}

	if withField.fields["key"] != "value" {
		t.Errorf("WithField() field value = %v, want %v", withField.fields["key"], "value")
	}

	// Original should be unchanged
	if len(logger.fields) != 0 {
		t.Error("original logger fields should be unchanged")
	}
}

func TestStructuredErrorLogger_WithFields(t *testing.T) {
	t.Parallel()

	logger := NewStructuredErrorLogger()
	withFields := logger.WithFields(map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	})

	if withFields == nil {
		t.Fatal("WithFields() returned nil")
	}

	if len(withFields.fields) != 2 {
		t.Errorf("WithFields() should add two fields, got %d", len(withFields.fields))
	}

	// Original should be unchanged
	if len(logger.fields) != 0 {
		t.Error("original logger fields should be unchanged")
	}
}

func TestStructuredErrorLogger_FieldChaining(t *testing.T) {
	t.Parallel()

	logger := NewStructuredErrorLogger()
	chainedLogger := logger.
		WithField("a", 1).
		WithField("b", 2).
		WithField("c", 3)

	if len(chainedLogger.fields) != 3 {
		t.Errorf("chained fields should have 3 entries, got %d", len(chainedLogger.fields))
	}
}

func TestFileRotatingLogger_PathTraversal(t *testing.T) {
	t.Parallel()

	// Note: filepath.Clean resolves absolute paths, so /tmp/../etc/passwd becomes /etc/passwd
	// The path traversal check only catches relative paths that still contain ".." after cleaning
	tests := []struct {
		name        string
		path        string
		shouldError bool
	}{
		{
			name:        "relative path traversing beyond root",
			path:        "../../../../../../../etc/passwd",
			shouldError: true,
		},
		{
			name:        "deep relative path with ..",
			path:        "a/b/../../../../etc/passwd",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := NewFileRotatingLogger(tt.path, 1024*1024, 5)

			if tt.shouldError {
				if err == nil {
					t.Error("expected path traversal error, got nil")
				}
				if !errors.Is(err, errPathTraversalDetected) {
					t.Errorf("expected errPathTraversalDetected, got: %v", err)
				}
			}
		})
	}
}

func TestMockErrorLogger_Tracking(t *testing.T) {
	t.Parallel()

	mock := NewMockErrorLogger()

	mock.LogError(errLoggerTestTracking)

	if len(mock.LogErrorCalls) != 1 {
		t.Errorf("LogErrorCalls = %d, want 1", len(mock.LogErrorCalls))
	}
	if !errors.Is(mock.LogErrorCalls[0], errLoggerTestTracking) {
		t.Error("LogErrorCalls should contain the logged error")
	}

	mock.LogErrorWithContext(context.Background(), errLoggerTestTracking)
	if len(mock.LogErrorWithContextCalls) != 1 {
		t.Errorf("LogErrorWithContextCalls = %d, want 1", len(mock.LogErrorWithContextCalls))
	}

	mageErr := NewMageError("mage error")
	mock.LogMageError(mageErr)
	if len(mock.LogMageErrorCalls) != 1 {
		t.Errorf("LogMageErrorCalls = %d, want 1", len(mock.LogMageErrorCalls))
	}

	mock.SetLogLevel(SeverityCritical)
	if len(mock.SetLogLevelCalls) != 1 {
		t.Errorf("SetLogLevelCalls = %d, want 1", len(mock.SetLogLevelCalls))
	}
	if mock.LogLevel != SeverityCritical {
		t.Errorf("LogLevel = %v, want %v", mock.LogLevel, SeverityCritical)
	}
}
