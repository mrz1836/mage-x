// Package errors provides error factory utilities to reduce duplication in error handling
package errors

import (
	"fmt"
	"strings"
	"sync"
)

// CommonErrorFactory provides standardized error creation patterns
type CommonErrorFactory struct{}

// NewCommonErrorFactory creates a new error factory
func NewCommonErrorFactory() *CommonErrorFactory {
	return &CommonErrorFactory{}
}

// createDefaultFactory creates a new factory instance
// Using a function instead of global variable to satisfy linters
func createDefaultFactory() *CommonErrorFactory {
	return NewCommonErrorFactory()
}

// Wrap creates a wrapped error with additional context
func (f *CommonErrorFactory) Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

// Wrapf creates a wrapped error with formatted message
func (f *CommonErrorFactory) Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// WithCode creates an error with a specific error code
func (f *CommonErrorFactory) WithCode(code ErrorCode, message string) MageError {
	return NewBuilder().WithCode(code).WithMessage(message).Build()
}

// WithCodef creates an error with a specific error code and formatted message
func (f *CommonErrorFactory) WithCodef(code ErrorCode, format string, args ...interface{}) MageError {
	return NewBuilder().WithCode(code).WithMessage(format, args...).Build()
}

// NotFound creates a standardized "not found" error
func (f *CommonErrorFactory) NotFound(resource, identifier string) error {
	return f.WithCodef(ErrNotFound, "%s not found: %s", resource, identifier)
}

// AlreadyExists creates a standardized "already exists" error
func (f *CommonErrorFactory) AlreadyExists(resource, identifier string) error {
	return f.WithCodef(ErrAlreadyExists, "%s already exists: %s", resource, identifier)
}

// InvalidArgument creates a standardized "invalid argument" error
func (f *CommonErrorFactory) InvalidArgument(field, value, reason string) error {
	return f.WithCodef(ErrInvalidArgument, "invalid %s '%s': %s", field, value, reason)
}

// OperationFailed creates a standardized "operation failed" error
func (f *CommonErrorFactory) OperationFailed(operation string, cause error) error {
	return f.WithCodef(ErrInternal, "%s failed", operation).WithCause(cause)
}

// FileNotFound creates a standardized file not found error
func (f *CommonErrorFactory) FileNotFound(path string) error {
	return f.WithCodef(ErrFileNotFound, "file not found: %s", path)
}

// FileExists creates a standardized file exists error
func (f *CommonErrorFactory) FileExists(path string) error {
	return f.WithCodef(ErrFileExists, "file already exists: %s", path)
}

// CommandFailed creates a standardized command failure error
func (f *CommonErrorFactory) CommandFailed(command string, args []string, cause error) error {
	cmdStr := command
	if len(args) > 0 {
		cmdStr = fmt.Sprintf("%s %s", command, strings.Join(args, " "))
	}
	return f.WithCodef(ErrCommandFailed, "command failed: %s", cmdStr).WithCause(cause)
}

// ValidationFailed creates a standardized validation failure error
func (f *CommonErrorFactory) ValidationFailed(field, value, constraint string) error {
	return f.WithCodef(ErrInvalidArgument, "validation failed for field '%s' with value '%s': %s", field, value, constraint)
}

// ConfigurationError creates a standardized configuration error
func (f *CommonErrorFactory) ConfigurationError(setting, issue string) error {
	return f.WithCodef(ErrConfigInvalid, "configuration error in '%s': %s", setting, issue)
}

// PermissionDenied creates a standardized permission denied error
func (f *CommonErrorFactory) PermissionDenied(resource, operation string) error {
	return f.WithCodef(ErrPermissionDenied, "permission denied: cannot %s %s", operation, resource)
}

// Timeout creates a standardized timeout error
func (f *CommonErrorFactory) Timeout(operation string, duration string) error {
	return f.WithCodef(ErrTimeout, "%s timed out after %s", operation, duration)
}

// Chain creates an error chain from multiple errors
func (f *CommonErrorFactory) Chain(errors ...error) ErrorChain {
	chain := NewChain()
	for _, err := range errors {
		if err != nil {
			if addErr := chain.Add(err); addErr != nil {
				// Log or handle the add error if needed
				// For now, we continue to avoid breaking the chain
				continue
			}
		}
	}
	return chain
}

// Recovery creates an error indicating recovery from a panic
func (f *CommonErrorFactory) Recovery(panicValue interface{}, operation string) error {
	return f.WithCodef(ErrInternal, "recovered from panic in %s: %v", operation, panicValue)
}

// Package-level convenience functions using the default factory

// WrapError creates a wrapped error with additional context
func WrapError(err error, message string) error {
	return createDefaultFactory().Wrap(err, message)
}

// WrapErrorf creates a wrapped error with formatted message
func WrapErrorf(err error, format string, args ...interface{}) error {
	return createDefaultFactory().Wrapf(err, format, args...)
}

// ErrorWithCode creates an error with a specific error code
func ErrorWithCode(code ErrorCode, message string) MageError {
	return createDefaultFactory().WithCode(code, message)
}

// ErrorWithCodef creates an error with a specific error code and formatted message
func ErrorWithCodef(code ErrorCode, format string, args ...interface{}) MageError {
	return createDefaultFactory().WithCodef(code, format, args...)
}

// NotFound creates a standardized "not found" error
func NotFound(resource, identifier string) error {
	return createDefaultFactory().NotFound(resource, identifier)
}

// AlreadyExists creates a standardized "already exists" error
func AlreadyExists(resource, identifier string) error {
	return createDefaultFactory().AlreadyExists(resource, identifier)
}

// InvalidArgument creates a standardized "invalid argument" error
func InvalidArgument(field, value, reason string) error {
	return createDefaultFactory().InvalidArgument(field, value, reason)
}

// OperationFailed creates a standardized "operation failed" error
func OperationFailed(operation string, cause error) error {
	return createDefaultFactory().OperationFailed(operation, cause)
}

// FileNotFound creates a standardized file not found error
func FileNotFound(path string) error {
	return createDefaultFactory().FileNotFound(path)
}

// FileExists creates a standardized file exists error
func FileExists(path string) error {
	return createDefaultFactory().FileExists(path)
}

// CommandFailed creates a standardized command failure error
func CommandFailed(command string, args []string, cause error) error {
	return createDefaultFactory().CommandFailed(command, args, cause)
}

// ValidationFailed creates a standardized validation failure error
func ValidationFailed(field, value, constraint string) error {
	return createDefaultFactory().ValidationFailed(field, value, constraint)
}

// ConfigurationError creates a standardized configuration error
func ConfigurationError(setting, issue string) error {
	return createDefaultFactory().ConfigurationError(setting, issue)
}

// PermissionDenied creates a standardized permission denied error
func PermissionDenied(resource, operation string) error {
	return createDefaultFactory().PermissionDenied(resource, operation)
}

// Timeout creates a standardized timeout error
func Timeout(operation string, duration string) error {
	return createDefaultFactory().Timeout(operation, duration)
}

// Chain creates an error chain from multiple errors
func Chain(errors ...error) ErrorChain {
	return createDefaultFactory().Chain(errors...)
}

// Recovery creates an error indicating recovery from a panic
func Recovery(panicValue interface{}, operation string) error {
	return createDefaultFactory().Recovery(panicValue, operation)
}

// Common error creation patterns for specific domains

// BuildErrorFactory provides build-specific error creation
type BuildErrorFactory struct {
	*CommonErrorFactory
}

// NewBuildErrorFactory creates a build-specific error factory
func NewBuildErrorFactory() *BuildErrorFactory {
	return &BuildErrorFactory{
		CommonErrorFactory: NewCommonErrorFactory(),
	}
}

// CompilationFailed creates a compilation failure error
func (f *BuildErrorFactory) CompilationFailed(file string, cause error) error {
	return f.WithCodef(ErrCompileFailed, "compilation failed for %s", file).WithCause(cause)
}

// DependencyError creates a dependency-related error
func (f *BuildErrorFactory) DependencyError(dependency, issue string) error {
	return f.WithCodef(ErrDependencyError, "dependency error for %s: %s", dependency, issue)
}

// PackagingFailed creates a packaging failure error
func (f *BuildErrorFactory) PackagingFailed(format string, cause error) error {
	return f.WithCodef(ErrPackageFailed, "packaging failed for format %s", format).WithCause(cause)
}

// TestErrorFactory provides test-specific error creation
type TestErrorFactory struct {
	*CommonErrorFactory
}

// NewTestErrorFactory creates a test-specific error factory
func NewTestErrorFactory() *TestErrorFactory {
	return &TestErrorFactory{
		CommonErrorFactory: NewCommonErrorFactory(),
	}
}

// TestFailed creates a test failure error
func (f *TestErrorFactory) TestFailed(testName string, cause error) error {
	return f.WithCodef(ErrTestFailed, "test failed: %s", testName).WithCause(cause)
}

// CoverageBelowThreshold creates a coverage threshold error
func (f *TestErrorFactory) CoverageBelowThreshold(actual, required float64) error {
	return f.WithCodef(ErrTestFailed, "coverage %.2f%% below threshold %.2f%%", actual, required)
}

// SecurityErrorFactory provides security-specific error creation
type SecurityErrorFactory struct {
	*CommonErrorFactory
}

// NewSecurityErrorFactory creates a security-specific error factory
func NewSecurityErrorFactory() *SecurityErrorFactory {
	return &SecurityErrorFactory{
		CommonErrorFactory: NewCommonErrorFactory(),
	}
}

// SecurityValidationFailed creates a security validation failure error
func (f *SecurityErrorFactory) SecurityValidationFailed(check, details string) error {
	return f.WithCodef(ErrSecurityFailed, "security validation failed for %s: %s", check, details)
}

// UnauthorizedAccess creates an unauthorized access error
func (f *SecurityErrorFactory) UnauthorizedAccess(resource, user string) error {
	return f.WithCodef(ErrUnauthorized, "unauthorized access to %s by user %s", resource, user)
}

// createBuildErrors creates a new build error factory
func createBuildErrors() *BuildErrorFactory {
	return NewBuildErrorFactory()
}

// createTestErrors creates a new test error factory
func createTestErrors() *TestErrorFactory {
	return NewTestErrorFactory()
}

// createSecurityErrors creates a new security error factory
func createSecurityErrors() *SecurityErrorFactory {
	return NewSecurityErrorFactory()
}

// Domain-specific factory instances using sync.Once for thread-safe lazy initialization
var (
	buildErrorsOnce    sync.Once //nolint:gochecknoglobals // Required for singleton pattern
	testErrorsOnce     sync.Once //nolint:gochecknoglobals // Required for singleton pattern
	securityErrorsOnce sync.Once //nolint:gochecknoglobals // Required for singleton pattern

	buildErrorsInstance    *BuildErrorFactory    //nolint:gochecknoglobals // Required for singleton pattern
	testErrorsInstance     *TestErrorFactory     //nolint:gochecknoglobals // Required for singleton pattern
	securityErrorsInstance *SecurityErrorFactory //nolint:gochecknoglobals // Required for singleton pattern
)

// GetBuildErrors returns the singleton build error factory
func GetBuildErrors() *BuildErrorFactory {
	buildErrorsOnce.Do(func() {
		buildErrorsInstance = createBuildErrors()
	})
	return buildErrorsInstance
}

// GetTestErrors returns the singleton test error factory
func GetTestErrors() *TestErrorFactory {
	testErrorsOnce.Do(func() {
		testErrorsInstance = createTestErrors()
	})
	return testErrorsInstance
}

// GetSecurityErrors returns the singleton security error factory
func GetSecurityErrors() *SecurityErrorFactory {
	securityErrorsOnce.Do(func() {
		securityErrorsInstance = createSecurityErrors()
	})
	return securityErrorsInstance
}
