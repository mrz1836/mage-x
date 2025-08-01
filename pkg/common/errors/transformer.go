// Package errors provides comprehensive error transformation capabilities
package errors

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Static errors to comply with err113 linter
var (
	errMockTransform = errors.New("mock transform error")
	errSanitized     = errors.New("sanitized error")
)

// DefaultErrorTransformer implements the ErrorTransformer interface
type DefaultErrorTransformer struct {
	mu                 sync.RWMutex
	codeTransforms     map[ErrorCode]ErrorCode
	severityTransforms map[Severity]Severity
	transformers       []NamedTransformer
	enabled            bool
}

// NamedTransformer represents a transformer function with a name
type NamedTransformer struct {
	Name        string
	Function    func(error) error
	Description string
	Priority    int
}

// NewErrorTransformer creates a new error transformer
func NewErrorTransformer() ErrorTransformer {
	return &DefaultErrorTransformer{
		codeTransforms:     make(map[ErrorCode]ErrorCode),
		severityTransforms: make(map[Severity]Severity),
		transformers:       make([]NamedTransformer, 0),
		enabled:            true,
	}
}

// Transform applies all configured transformations to an error
func (t *DefaultErrorTransformer) Transform(err error) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if !t.enabled || err == nil {
		return err
	}

	currentErr := err

	// Apply custom transformers first (in priority order)
	for _, transformer := range t.transformers {
		if transformed := transformer.Function(currentErr); transformed != nil {
			currentErr = transformed
		}
	}

	// Apply code and severity transformations for MageErrors
	var mageErr MageError
	if errors.As(currentErr, &mageErr) {
		currentErr = t.transformMageError(mageErr)
	}

	return currentErr
}

// TransformCode adds a code transformation rule
func (t *DefaultErrorTransformer) TransformCode(from, to ErrorCode) ErrorTransformer {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.codeTransforms[from] = to
	return t
}

// TransformSeverity adds a severity transformation rule
func (t *DefaultErrorTransformer) TransformSeverity(from, to Severity) ErrorTransformer {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.severityTransforms[from] = to
	return t
}

// AddTransformer adds a custom transformation function
func (t *DefaultErrorTransformer) AddTransformer(fn func(error) error) ErrorTransformer {
	return t.AddNamedTransformer(fmt.Sprintf("transformer_%d", len(t.transformers)), fn, 0)
}

// AddNamedTransformer adds a named transformation function with priority
func (t *DefaultErrorTransformer) AddNamedTransformer(name string, fn func(error) error, priority int) ErrorTransformer {
	t.mu.Lock()
	defer t.mu.Unlock()

	transformer := NamedTransformer{
		Name:     name,
		Function: fn,
		Priority: priority,
	}

	// Insert in priority order (higher priority first)
	inserted := false
	for i, existing := range t.transformers {
		if priority > existing.Priority {
			// Insert at position i
			t.transformers = append(t.transformers[:i], append([]NamedTransformer{transformer}, t.transformers[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		t.transformers = append(t.transformers, transformer)
	}

	return t
}

// RemoveTransformer removes a named transformation function
func (t *DefaultErrorTransformer) RemoveTransformer(name string) ErrorTransformer {
	t.mu.Lock()
	defer t.mu.Unlock()

	for i, transformer := range t.transformers {
		if transformer.Name == name {
			t.transformers = append(t.transformers[:i], t.transformers[i+1:]...)
			break
		}
	}

	return t
}

// SetEnabled enables or disables transformation
func (t *DefaultErrorTransformer) SetEnabled(enabled bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.enabled = enabled
}

// IsEnabled returns whether transformation is enabled
func (t *DefaultErrorTransformer) IsEnabled() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return t.enabled
}

// GetTransformers returns all registered transformers
func (t *DefaultErrorTransformer) GetTransformers() []NamedTransformer {
	t.mu.RLock()
	defer t.mu.RUnlock()

	transformers := make([]NamedTransformer, len(t.transformers))
	copy(transformers, t.transformers)
	return transformers
}

// ClearTransformers removes all transformation rules
func (t *DefaultErrorTransformer) ClearTransformers() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.codeTransforms = make(map[ErrorCode]ErrorCode)
	t.severityTransforms = make(map[Severity]Severity)
	t.transformers = make([]NamedTransformer, 0)
}

// transformMageError applies code and severity transformations to a MageError
func (t *DefaultErrorTransformer) transformMageError(err MageError) error {
	// Create a new builder with the original error's information
	builder := NewErrorBuilder()
	builder.WithMessage("%s", err.Error())
	builder.WithCode(err.Code())
	builder.WithSeverity(err.Severity())

	// Copy fields from the original error's context
	errCtx := err.Context()
	if len(errCtx.Fields) > 0 {
		for key, value := range errCtx.Fields {
			builder.WithField(key, value)
		}
	}

	// Transform error code if rule exists
	if newCode, exists := t.codeTransforms[err.Code()]; exists {
		builder.WithCode(newCode)
	}

	// Transform severity if rule exists
	if newSeverity, exists := t.severityTransforms[err.Severity()]; exists {
		builder.WithSeverity(newSeverity)
	}

	return builder.Build()
}

// Predefined transformation functions

// SanitizeTransformer removes sensitive information from error messages
func SanitizeTransformer(err error) error {
	if err == nil {
		return nil
	}

	message := err.Error()

	// Remove common sensitive patterns
	patterns := []struct {
		pattern     string
		replacement string
	}{
		{"password=\\S+", "password=***"},
		{"token=\\S+", "token=***"},
		{"key=\\S+", "key=***"},
		{"secret=\\S+", "secret=***"},
		{"api[_-]?key=\\S+", "api_key=***"},
		{"auth[_-]?token=\\S+", "auth_token=***"},
	}

	sanitized := message
	for _, p := range patterns {
		// Simple string replacement (in production, use regex)
		if idx := strings.Index(p.pattern, "="); idx != -1 && strings.Contains(sanitized, p.pattern[:idx]) {
			parts := strings.Split(sanitized, " ")
			for i, part := range parts {
				if idx := strings.Index(p.pattern, "="); idx != -1 && strings.Contains(part, p.pattern[:idx]) {
					parts[i] = p.replacement
				}
			}
			sanitized = strings.Join(parts, " ")
		}
	}

	var mageErr MageError
	if errors.As(err, &mageErr) {
		builder := NewErrorBuilder()
		builder.WithMessage("%s", sanitized)
		builder.WithCode(mageErr.Code())
		builder.WithSeverity(mageErr.Severity())
		return builder.Build()
	}

	return fmt.Errorf("%w: %s", errSanitized, sanitized)
}

// EnrichTransformer adds additional context to errors
func EnrichTransformer(additionalContext map[string]interface{}) func(error) error {
	return func(err error) error {
		if err == nil {
			return nil
		}

		var mageErr MageError
		if errors.As(err, &mageErr) {
			builder := NewErrorBuilder()
			builder.WithMessage("%s", mageErr.Error())
			builder.WithCode(mageErr.Code())
			builder.WithSeverity(mageErr.Severity())
			for key, value := range additionalContext {
				builder.WithField(key, value)
			}
			return builder.Build()
		}

		return err
	}
}

// RetryableTransformer marks errors as retryable based on patterns
func RetryableTransformer(retryablePatterns []string) func(error) error {
	return func(err error) error {
		if err == nil {
			return nil
		}

		message := err.Error()
		isRetryable := false

		for _, pattern := range retryablePatterns {
			if strings.Contains(strings.ToLower(message), strings.ToLower(pattern)) {
				isRetryable = true
				break
			}
		}

		var mageErr MageError
		if errors.As(err, &mageErr) {
			builder := NewErrorBuilder()
			builder.WithMessage("%s", mageErr.Error())
			builder.WithCode(mageErr.Code())
			builder.WithSeverity(mageErr.Severity())
			builder.WithField("retryable", isRetryable)
			return builder.Build()
		}

		return err
	}
}

// NewConditionalTransformer creates a transformer that only applies when condition is met
func NewConditionalTransformer(condition func(error) bool, transformer func(error) error) func(error) error {
	return func(err error) error {
		if err == nil || !condition(err) {
			return err
		}
		return transformer(err)
	}
}

// NewChainTransformer creates a new chain transformer
func NewChainTransformer(transformers ...func(error) error) func(error) error {
	return func(err error) error {
		if err == nil {
			return nil
		}

		current := err
		for _, transformer := range transformers {
			if transformed := transformer(current); transformed != nil {
				current = transformed
			}
		}
		return current
	}
}

// Common transformation presets

// NewSecurityTransformer creates a transformer optimized for security
func NewSecurityTransformer() ErrorTransformer {
	transformer := NewErrorTransformer()

	// Add sanitization
	transformer.AddTransformer(SanitizeTransformer)

	// Transform security-related error codes to generic ones
	transformer.TransformCode(ErrNotFound, ErrUnknown)
	transformer.TransformCode(ErrPermissionDenied, ErrUnknown)

	// Reduce severity of some security errors to avoid information leakage
	transformer.TransformSeverity(SeverityError, SeverityWarning)

	return transformer
}

// NewDevelopmentTransformer creates a transformer optimized for development
func NewDevelopmentTransformer() ErrorTransformer {
	transformer := NewErrorTransformer()

	// Add context enrichment
	transformer.AddTransformer(EnrichTransformer(map[string]interface{}{
		"environment": "development",
		"debug":       true,
	}))

	// Mark common errors as retryable
	transformer.AddTransformer(RetryableTransformer([]string{
		"connection refused",
		"timeout",
		"temporary failure",
		"network",
	}))

	return transformer
}

// NewProductionTransformer creates a transformer optimized for production
func NewProductionTransformer() ErrorTransformer {
	transformer := NewErrorTransformer()

	// Add sanitization
	transformer.AddTransformer(SanitizeTransformer)

	// Add production context
	transformer.AddTransformer(EnrichTransformer(map[string]interface{}{
		"environment": "production",
		"debug":       false,
	}))

	// Transform debug errors to info level
	transformer.TransformSeverity(SeverityDebug, SeverityInfo)

	return transformer
}

// MockErrorTransformer implements ErrorTransformer for testing
type MockErrorTransformer struct {
	TransformCalls         []error
	TransformCodeCalls     []MockTransformCodeCall
	TransformSeverityCalls []MockTransformSeverityCall
	AddTransformerCalls    []func(error) error
	RemoveTransformerCalls []string
	ShouldError            bool
	TransformResult        error
	Enabled                bool
}

// MockTransformCodeCall represents a call to TransformCode on MockErrorTransformer
type MockTransformCodeCall struct {
	From ErrorCode
	To   ErrorCode
}

// MockTransformSeverityCall represents a call to TransformSeverity on MockErrorTransformer
type MockTransformSeverityCall struct {
	From Severity
	To   Severity
}

// NewMockErrorTransformer creates a new MockErrorTransformer instance
func NewMockErrorTransformer() *MockErrorTransformer {
	return &MockErrorTransformer{
		TransformCalls:         make([]error, 0),
		TransformCodeCalls:     make([]MockTransformCodeCall, 0),
		TransformSeverityCalls: make([]MockTransformSeverityCall, 0),
		AddTransformerCalls:    make([]func(error) error, 0),
		RemoveTransformerCalls: make([]string, 0),
		Enabled:                true,
	}
}

// Transform transforms an error using the MockErrorTransformer
func (m *MockErrorTransformer) Transform(err error) error {
	m.TransformCalls = append(m.TransformCalls, err)
	if m.ShouldError {
		return errMockTransform
	}
	if m.TransformResult != nil {
		return m.TransformResult
	}
	return err
}

// TransformCode adds a code transformation rule to the MockErrorTransformer
func (m *MockErrorTransformer) TransformCode(from, to ErrorCode) ErrorTransformer {
	m.TransformCodeCalls = append(m.TransformCodeCalls, MockTransformCodeCall{
		From: from,
		To:   to,
	})
	return m
}

// TransformSeverity adds a severity transformation rule to the MockErrorTransformer
func (m *MockErrorTransformer) TransformSeverity(from, to Severity) ErrorTransformer {
	m.TransformSeverityCalls = append(m.TransformSeverityCalls, MockTransformSeverityCall{
		From: from,
		To:   to,
	})
	return m
}

// AddTransformer adds a transformer function to the MockErrorTransformer
func (m *MockErrorTransformer) AddTransformer(fn func(error) error) ErrorTransformer {
	m.AddTransformerCalls = append(m.AddTransformerCalls, fn)
	return m
}

// RemoveTransformer removes a transformer by name from the MockErrorTransformer
func (m *MockErrorTransformer) RemoveTransformer(name string) ErrorTransformer {
	m.RemoveTransformerCalls = append(m.RemoveTransformerCalls, name)
	return m
}
