package errors

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

// Test sentinel errors for err113 compliance.
var (
	errMageUnderlyingCause = errors.New("underlying cause")
	errMageTheCause        = errors.New("the cause")
	errMageRootCause       = errors.New("root cause")
	errMageUnrelated       = errors.New("unrelated")
	errMageCauseFormat     = errors.New("the cause")
)

func TestNewMageError(t *testing.T) {
	t.Parallel()

	err := NewMageError("test message")
	if err == nil {
		t.Fatal("NewMageError() returned nil")
	}

	if err.message != "test message" {
		t.Errorf("message = %q, want %q", err.message, "test message")
	}

	if err.code != ErrUnknown {
		t.Errorf("code = %v, want %v", err.code, ErrUnknown)
	}

	if err.severity != SeverityError {
		t.Errorf("severity = %v, want %v", err.severity, SeverityError)
	}

	if err.context.Fields == nil {
		t.Error("context.Fields should be initialized")
	}
}

func TestDefaultMageError_Error(t *testing.T) {
	t.Parallel()

	t.Run("without cause", func(t *testing.T) {
		t.Parallel()
		err := NewMageError("test message")
		if err.Error() != "test message" {
			t.Errorf("Error() = %q, want %q", err.Error(), "test message")
		}
	})

	t.Run("with cause", func(t *testing.T) {
		t.Parallel()
		err, ok := NewMageError("test message").WithCause(errMageUnderlyingCause).(*DefaultMageError)
		if !ok {
			t.Fatal("expected *DefaultMageError")
		}

		errStr := err.Error()
		if !strings.Contains(errStr, "test message") {
			t.Errorf("Error() should contain message, got: %s", errStr)
		}
		if !strings.Contains(errStr, "underlying cause") {
			t.Errorf("Error() should contain cause, got: %s", errStr)
		}
	})
}

func TestDefaultMageError_Code(t *testing.T) {
	t.Parallel()

	err := NewMageError("test")
	if err.Code() != ErrUnknown {
		t.Errorf("Code() = %v, want %v", err.Code(), ErrUnknown)
	}

	withCode, ok := err.WithCode(ErrInternal).(*DefaultMageError)
	if !ok {
		t.Fatal("expected *DefaultMageError")
	}
	if withCode.Code() != ErrInternal {
		t.Errorf("WithCode().Code() = %v, want %v", withCode.Code(), ErrInternal)
	}

	// Original should be unchanged
	if err.Code() != ErrUnknown {
		t.Error("original error code should be unchanged")
	}
}

func TestDefaultMageError_Severity(t *testing.T) {
	t.Parallel()

	err := NewMageError("test")
	if err.Severity() != SeverityError {
		t.Errorf("Severity() = %v, want %v", err.Severity(), SeverityError)
	}

	withSeverity, ok := err.WithSeverity(SeverityCritical).(*DefaultMageError)
	if !ok {
		t.Fatal("expected *DefaultMageError")
	}
	if withSeverity.Severity() != SeverityCritical {
		t.Errorf("WithSeverity().Severity() = %v, want %v", withSeverity.Severity(), SeverityCritical)
	}

	// Original should be unchanged
	if err.Severity() != SeverityError {
		t.Error("original error severity should be unchanged")
	}
}

func TestDefaultMageError_Context(t *testing.T) {
	t.Parallel()

	err := NewMageError("test")
	ctx := err.Context()

	if ctx.Fields == nil {
		t.Error("Context().Fields should not be nil")
	}

	// Modify the returned context
	ctx.Fields["modified"] = true

	// Original should be unchanged (deep copy)
	originalCtx := err.Context()
	if _, exists := originalCtx.Fields["modified"]; exists {
		t.Error("Context() should return a copy, not the original")
	}
}

func TestDefaultMageError_WithField(t *testing.T) {
	t.Parallel()

	err := NewMageError("test")
	withField, ok := err.WithField("key", "value").(*DefaultMageError)
	if !ok {
		t.Fatal("expected *DefaultMageError")
	}

	ctx := withField.Context()
	if ctx.Fields["key"] != "value" {
		t.Errorf("WithField() field = %v, want %v", ctx.Fields["key"], "value")
	}

	// Original should be unchanged
	if _, exists := err.Context().Fields["key"]; exists {
		t.Error("original error should not have the new field")
	}
}

func TestDefaultMageError_WithFields(t *testing.T) {
	t.Parallel()

	err := NewMageError("test")
	withFields, ok := err.WithFields(map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}).(*DefaultMageError)
	if !ok {
		t.Fatal("expected *DefaultMageError")
	}

	ctx := withFields.Context()
	if len(ctx.Fields) != 2 {
		t.Errorf("WithFields() should add 2 fields, got %d", len(ctx.Fields))
	}

	// Original should be unchanged
	if len(err.Context().Fields) != 0 {
		t.Error("original error should not have the new fields")
	}
}

func TestDefaultMageError_WithCause(t *testing.T) {
	t.Parallel()

	err := NewMageError("test")
	withCause, ok := err.WithCause(errMageTheCause).(*DefaultMageError)
	if !ok {
		t.Fatal("expected *DefaultMageError")
	}

	if !errors.Is(withCause.Cause(), errMageTheCause) {
		t.Errorf("WithCause().Cause() should return the cause")
	}

	if !errors.Is(withCause.Unwrap(), errMageTheCause) {
		t.Errorf("Unwrap() should return the cause")
	}

	// Original should be unchanged
	if err.Cause() != nil {
		t.Error("original error should not have a cause")
	}
}

func TestDefaultMageError_WithOperation(t *testing.T) {
	t.Parallel()

	err := NewMageError("test")
	withOp, ok := err.WithOperation("build").(*DefaultMageError)
	if !ok {
		t.Fatal("expected *DefaultMageError")
	}

	ctx := withOp.Context()
	if ctx.Operation != "build" {
		t.Errorf("WithOperation().Context().Operation = %q, want %q", ctx.Operation, "build")
	}
}

func TestDefaultMageError_WithResource(t *testing.T) {
	t.Parallel()

	err := NewMageError("test")
	withRes, ok := err.WithResource("/path/to/file").(*DefaultMageError)
	if !ok {
		t.Fatal("expected *DefaultMageError")
	}

	ctx := withRes.Context()
	if ctx.Resource != "/path/to/file" {
		t.Errorf("WithResource().Context().Resource = %q, want %q", ctx.Resource, "/path/to/file")
	}
}

func TestDefaultMageError_WithContext(t *testing.T) {
	t.Parallel()

	err := NewMageError("test")
	newCtx := &ErrorContext{
		Operation: "test-op",
		Resource:  "test-resource",
		Fields:    map[string]interface{}{"custom": "field"},
	}

	withCtx, ok := err.WithContext(newCtx).(*DefaultMageError)
	if !ok {
		t.Fatal("expected *DefaultMageError")
	}
	ctx := withCtx.Context()

	if ctx.Operation != "test-op" {
		t.Errorf("Operation = %q, want %q", ctx.Operation, "test-op")
	}
	if ctx.Resource != "test-resource" {
		t.Errorf("Resource = %q, want %q", ctx.Resource, "test-resource")
	}
}

func TestDefaultMageError_Clone(t *testing.T) {
	t.Parallel()

	original := NewMageError("original message")
	original.code = ErrInternal
	original.severity = SeverityCritical
	original.context.Operation = "test-op"
	original.context.Fields["key"] = "value"

	cloned := original.clone()

	// Verify cloned values
	if cloned.message != original.message {
		t.Error("cloned message should match")
	}
	if cloned.code != original.code {
		t.Error("cloned code should match")
	}
	if cloned.severity != original.severity {
		t.Error("cloned severity should match")
	}
	if cloned.context.Operation != original.context.Operation {
		t.Error("cloned operation should match")
	}

	// Modify cloned fields
	cloned.context.Fields["key"] = "modified"
	cloned.context.Fields["new"] = "added"

	// Original should be unchanged (deep copy)
	if original.context.Fields["key"] != "value" {
		t.Error("original field should be unchanged")
	}
	if _, exists := original.context.Fields["new"]; exists {
		t.Error("original should not have new field")
	}
}

func TestDefaultMageError_Is(t *testing.T) {
	t.Parallel()

	err, ok := NewMageError("test").WithCause(errMageRootCause).(*DefaultMageError)
	if !ok {
		t.Fatal("expected *DefaultMageError")
	}

	// Is should return true for the cause
	if !errors.Is(err, errMageRootCause) {
		t.Error("errors.Is(err, cause) should return true")
	}

	// Is should work with MageError with same code
	sameCodeErr, ok := NewMageError("other").WithCode(ErrUnknown).(*DefaultMageError)
	if !ok {
		t.Fatal("expected *DefaultMageError")
	}
	if !err.Is(sameCodeErr) {
		t.Error("Is() should return true for MageError with same code")
	}

	// Is should return false for unrelated errors
	if errors.Is(err, errMageUnrelated) {
		t.Error("errors.Is(err, unrelated) should return false")
	}
}

func TestDefaultMageError_As(t *testing.T) {
	t.Parallel()

	err := NewMageError("test")

	// As should work with *DefaultMageError
	var target *DefaultMageError
	if !errors.As(err, &target) {
		t.Error("errors.As(err, *DefaultMageError) should return true")
	}
	if target != err {
		t.Error("target should be the error itself")
	}

	// As should work with MageError interface
	var mageTarget MageError
	if !errors.As(err, &mageTarget) {
		t.Error("errors.As(err, MageError) should return true")
	}
}

func TestDefaultMageError_Format(t *testing.T) {
	t.Parallel()

	t.Run("basic format", func(t *testing.T) {
		t.Parallel()
		err := NewMageError("test message")
		formatted := err.Format(false)

		if !strings.Contains(formatted, "UNKNOWN") {
			t.Errorf("Format should contain error code, got: %s", formatted)
		}
		if !strings.Contains(formatted, "test message") {
			t.Errorf("Format should contain message, got: %s", formatted)
		}
	})

	t.Run("with operation", func(t *testing.T) {
		t.Parallel()
		err, ok := NewMageError("test").WithOperation("build").(*DefaultMageError)
		if !ok {
			t.Fatal("expected *DefaultMageError")
		}
		formatted := err.Format(false)

		if !strings.Contains(formatted, "Operation: build") {
			t.Errorf("Format should contain operation, got: %s", formatted)
		}
	})

	t.Run("with resource", func(t *testing.T) {
		t.Parallel()
		err, ok := NewMageError("test").WithResource("/path/file").(*DefaultMageError)
		if !ok {
			t.Fatal("expected *DefaultMageError")
		}
		formatted := err.Format(false)

		if !strings.Contains(formatted, "Resource: /path/file") {
			t.Errorf("Format should contain resource, got: %s", formatted)
		}
	})

	t.Run("with fields", func(t *testing.T) {
		t.Parallel()
		err, ok := NewMageError("test").WithField("key", "value").(*DefaultMageError)
		if !ok {
			t.Fatal("expected *DefaultMageError")
		}
		formatted := err.Format(false)

		if !strings.Contains(formatted, "key: value") {
			t.Errorf("Format should contain field, got: %s", formatted)
		}
	})

	t.Run("with cause", func(t *testing.T) {
		t.Parallel()
		err, ok := NewMageError("test").WithCause(errMageCauseFormat).(*DefaultMageError)
		if !ok {
			t.Fatal("expected *DefaultMageError")
		}
		formatted := err.Format(false)

		if !strings.Contains(formatted, "Caused by: the cause") {
			t.Errorf("Format should contain cause, got: %s", formatted)
		}
	})
}

func TestDefaultMageError_ConcurrentAccess(t *testing.T) {
	t.Parallel()

	err := NewMageError("concurrent test")

	var wg sync.WaitGroup
	const numGoroutines = 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Mix of read and write operations - testing concurrent access safety
			// Error return values are intentionally discarded in this concurrent safety test
			_ = err.Error()
			_ = err.Code()
			_ = err.Severity()
			_ = err.Context()
			_ = err.Cause() //nolint:errcheck // Testing concurrent safety
			_ = err.Format(false)

			// These create new instances, so should be safe
			_ = err.WithField("key", idx)         //nolint:errcheck // Testing concurrent safety
			_ = err.WithCode(ErrInternal)         //nolint:errcheck // Testing concurrent safety
			_ = err.WithSeverity(SeverityWarning) //nolint:errcheck // Testing concurrent safety
		}(i)
	}

	wg.Wait()
	// If we get here without panic, the test passes
}

func TestSeverity_MarshalUnmarshal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		severity Severity
		text     string
	}{
		{SeverityDebug, "DEBUG"},
		{SeverityInfo, "INFO"},
		{SeverityWarning, "WARNING"},
		{SeverityError, "ERROR"},
		{SeverityCritical, "CRITICAL"},
		{SeverityFatal, "FATAL"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()

			// Marshal
			data, err := tt.severity.MarshalText()
			if err != nil {
				t.Errorf("MarshalText() error = %v", err)
			}
			if string(data) != tt.text {
				t.Errorf("MarshalText() = %q, want %q", string(data), tt.text)
			}

			// Unmarshal
			var s Severity
			err = s.UnmarshalText([]byte(tt.text))
			if err != nil {
				t.Errorf("UnmarshalText() error = %v", err)
			}
			if s != tt.severity {
				t.Errorf("UnmarshalText() = %v, want %v", s, tt.severity)
			}
		})
	}
}

func TestSeverity_UnmarshalText_CaseInsensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected Severity
	}{
		{"debug", SeverityDebug},
		{"DEBUG", SeverityDebug},
		{"Debug", SeverityDebug},
		{"info", SeverityInfo},
		{"INFO", SeverityInfo},
		{"warning", SeverityWarning},
		{"WARNING", SeverityWarning},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()

			var s Severity
			err := s.UnmarshalText([]byte(tt.input))
			if err != nil {
				t.Errorf("UnmarshalText(%q) error = %v", tt.input, err)
			}
			if s != tt.expected {
				t.Errorf("UnmarshalText(%q) = %v, want %v", tt.input, s, tt.expected)
			}
		})
	}
}

func TestSeverity_UnmarshalText_Invalid(t *testing.T) {
	t.Parallel()

	var s Severity
	err := s.UnmarshalText([]byte("INVALID"))

	if err == nil {
		t.Error("UnmarshalText(INVALID) should return error")
	}
	if !errors.Is(err, ErrUnknownSeverity) {
		t.Errorf("error should be ErrUnknownSeverity, got: %v", err)
	}
}
