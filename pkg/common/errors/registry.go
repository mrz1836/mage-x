package errors

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// RealDefaultErrorRegistry is the actual implementation of ErrorRegistry
type RealDefaultErrorRegistry struct {
	mu          sync.RWMutex
	definitions map[ErrorCode]ErrorDefinition
}

// NewErrorRegistry creates a new error registry
func NewErrorRegistry() *RealDefaultErrorRegistry {
	registry := &RealDefaultErrorRegistry{
		definitions: make(map[ErrorCode]ErrorDefinition),
	}

	// Register common error codes
	registry.registerDefaultErrors()

	return registry
}

// Register registers a new error code
func (r *RealDefaultErrorRegistry) Register(code ErrorCode, description string) error {
	return r.RegisterWithSeverity(code, description, SeverityError)
}

// RegisterWithSeverity registers a new error code with severity
func (r *RealDefaultErrorRegistry) RegisterWithSeverity(code ErrorCode, description string, severity Severity) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.definitions[code]; exists {
		return fmt.Errorf("error code %s already registered", code)
	}

	r.definitions[code] = ErrorDefinition{
		Code:        code,
		Description: description,
		Severity:    severity,
		Category:    extractCategory(code),
		Retryable:   isRetryable(code),
		Tags:        extractTags(code),
	}

	return nil
}

// Unregister removes an error code from the registry
func (r *RealDefaultErrorRegistry) Unregister(code ErrorCode) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.definitions[code]; !exists {
		return fmt.Errorf("error code %s not found", code)
	}

	delete(r.definitions, code)
	return nil
}

// Get retrieves an error definition
func (r *RealDefaultErrorRegistry) Get(code ErrorCode) (ErrorDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	def, exists := r.definitions[code]
	return def, exists
}

// List returns all error definitions
func (r *RealDefaultErrorRegistry) List() []ErrorDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]ErrorDefinition, 0, len(r.definitions))
	for _, def := range r.definitions {
		result = append(result, def)
	}

	// Sort by code
	sort.Slice(result, func(i, j int) bool {
		return result[i].Code < result[j].Code
	})

	return result
}

// ListByPrefix returns error definitions matching a prefix
func (r *RealDefaultErrorRegistry) ListByPrefix(prefix string) []ErrorDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]ErrorDefinition, 0)
	for code, def := range r.definitions {
		if strings.HasPrefix(string(code), prefix) {
			result = append(result, def)
		}
	}

	// Sort by code
	sort.Slice(result, func(i, j int) bool {
		return result[i].Code < result[j].Code
	})

	return result
}

// ListBySeverity returns error definitions with a specific severity
func (r *RealDefaultErrorRegistry) ListBySeverity(severity Severity) []ErrorDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]ErrorDefinition, 0)
	for _, def := range r.definitions {
		if def.Severity == severity {
			result = append(result, def)
		}
	}

	// Sort by code
	sort.Slice(result, func(i, j int) bool {
		return result[i].Code < result[j].Code
	})

	return result
}

// Contains checks if an error code is registered
func (r *RealDefaultErrorRegistry) Contains(code ErrorCode) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.definitions[code]
	return exists
}

// Clear removes all error definitions
func (r *RealDefaultErrorRegistry) Clear() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.definitions = make(map[ErrorCode]ErrorDefinition)
	return nil
}

// registerDefaultErrors registers the default error codes
func (r *RealDefaultErrorRegistry) registerDefaultErrors() {
	// General errors
	r.definitions[ErrUnknown] = ErrorDefinition{
		Code:        ErrUnknown,
		Description: "Unknown error occurred",
		Severity:    SeverityError,
		Category:    "general",
		Retryable:   false,
	}

	r.definitions[ErrInternal] = ErrorDefinition{
		Code:        ErrInternal,
		Description: "Internal server error",
		Severity:    SeverityError,
		Category:    "general",
		Retryable:   true,
	}

	r.definitions[ErrInvalidArgument] = ErrorDefinition{
		Code:        ErrInvalidArgument,
		Description: "Invalid argument provided",
		Severity:    SeverityWarning,
		Category:    "general",
		Retryable:   false,
	}

	r.definitions[ErrNotFound] = ErrorDefinition{
		Code:        ErrNotFound,
		Description: "Resource not found",
		Severity:    SeverityWarning,
		Category:    "general",
		Retryable:   false,
	}

	r.definitions[ErrTimeout] = ErrorDefinition{
		Code:        ErrTimeout,
		Description: "Operation timed out",
		Severity:    SeverityError,
		Category:    "general",
		Retryable:   true,
	}

	// Build errors
	r.definitions[ErrBuildFailed] = ErrorDefinition{
		Code:        ErrBuildFailed,
		Description: "Build process failed",
		Severity:    SeverityError,
		Category:    "build",
		Retryable:   true,
	}

	r.definitions[ErrTestFailed] = ErrorDefinition{
		Code:        ErrTestFailed,
		Description: "Test execution failed",
		Severity:    SeverityError,
		Category:    "build",
		Retryable:   true,
	}

	// File errors
	r.definitions[ErrFileNotFound] = ErrorDefinition{
		Code:        ErrFileNotFound,
		Description: "File not found",
		Severity:    SeverityError,
		Category:    "file",
		Retryable:   false,
	}

	r.definitions[ErrFileAccessDenied] = ErrorDefinition{
		Code:        ErrFileAccessDenied,
		Description: "File access denied",
		Severity:    SeverityError,
		Category:    "file",
		Retryable:   false,
	}

	// Config errors
	r.definitions[ErrConfigNotFound] = ErrorDefinition{
		Code:        ErrConfigNotFound,
		Description: "Configuration not found",
		Severity:    SeverityError,
		Category:    "config",
		Retryable:   false,
	}

	r.definitions[ErrConfigInvalid] = ErrorDefinition{
		Code:        ErrConfigInvalid,
		Description: "Invalid configuration",
		Severity:    SeverityError,
		Category:    "config",
		Retryable:   false,
	}
}

// Helper functions

// extractCategory extracts the category from an error code
func extractCategory(code ErrorCode) string {
	str := string(code)
	parts := strings.Split(str, "_")
	if len(parts) > 1 {
		return strings.ToLower(parts[1])
	}
	return "general"
}

// isRetryable determines if an error is retryable
func isRetryable(code ErrorCode) bool {
	retryableCodes := map[ErrorCode]bool{
		ErrInternal:          true,
		ErrTimeout:           true,
		ErrResourceExhausted: true,
		ErrUnavailable:       true,
		ErrBuildFailed:       true,
		ErrTestFailed:        true,
		ErrCommandTimeout:    true,
	}

	return retryableCodes[code]
}

// extractTags extracts tags from an error code
func extractTags(code ErrorCode) []string {
	tags := []string{}

	// Add category as a tag
	category := extractCategory(code)
	tags = append(tags, category)

	// Add severity-based tags
	str := string(code)
	if strings.Contains(str, "FAILED") {
		tags = append(tags, "failure")
	}
	if strings.Contains(str, "TIMEOUT") {
		tags = append(tags, "timeout")
	}
	if strings.Contains(str, "INVALID") {
		tags = append(tags, "validation")
	}
	if strings.Contains(str, "NOT_FOUND") {
		tags = append(tags, "missing")
	}

	return tags
}

// Update DefaultErrorRegistry methods to use the real implementation
func (r *DefaultErrorRegistry) Register(code ErrorCode, description string) error {
	registry := NewErrorRegistry()
	return registry.Register(code, description)
}

func (r *DefaultErrorRegistry) RegisterWithSeverity(code ErrorCode, description string, severity Severity) error {
	registry := NewErrorRegistry()
	return registry.RegisterWithSeverity(code, description, severity)
}

func (r *DefaultErrorRegistry) Unregister(code ErrorCode) error {
	delete(r.definitions, code)
	return nil
}

func (r *DefaultErrorRegistry) Get(code ErrorCode) (ErrorDefinition, bool) {
	def, exists := r.definitions[code]
	return def, exists
}

func (r *DefaultErrorRegistry) List() []ErrorDefinition {
	result := make([]ErrorDefinition, 0, len(r.definitions))
	for _, def := range r.definitions {
		result = append(result, def)
	}
	return result
}

func (r *DefaultErrorRegistry) ListByPrefix(prefix string) []ErrorDefinition {
	result := make([]ErrorDefinition, 0)
	for code, def := range r.definitions {
		if strings.HasPrefix(string(code), prefix) {
			result = append(result, def)
		}
	}
	return result
}

func (r *DefaultErrorRegistry) ListBySeverity(severity Severity) []ErrorDefinition {
	result := make([]ErrorDefinition, 0)
	for _, def := range r.definitions {
		if def.Severity == severity {
			result = append(result, def)
		}
	}
	return result
}

func (r *DefaultErrorRegistry) Contains(code ErrorCode) bool {
	_, exists := r.definitions[code]
	return exists
}

func (r *DefaultErrorRegistry) Clear() error {
	r.definitions = make(map[ErrorCode]ErrorDefinition)
	return nil
}
