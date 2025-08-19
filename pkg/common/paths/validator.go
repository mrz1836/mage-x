package paths

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// Error definitions for validator operations
var (
	ErrRuleCannotBeNil         = errors.New("rule cannot be nil")
	ErrRuleNotFound            = errors.New("rule not found")
	ErrPathMustBeAbsolute      = errors.New("path must be absolute")
	ErrPathMustBeRelative      = errors.New("path must be relative")
	ErrPathDoesNotExist        = errors.New("path does not exist")
	ErrPathAlreadyExists       = errors.New("path already exists")
	ErrPathTraversalDetected   = errors.New("invalid path: path traversal detected")
	errPathNotExecutable       = errors.New("path is not executable")
	errPathNotDirectory        = errors.New("path is not a directory")
	errPathNotFile             = errors.New("path is not a file")
	ErrInvalidExtensions       = errors.New("path must have one of the required extensions")
	ErrPathTooLong             = errors.New("path exceeds maximum length")
	ErrPatternMismatch         = errors.New("path does not match required pattern")
	ErrForbiddenPattern        = errors.New("path matches forbidden pattern")
	ErrPathContainsNullByte    = errors.New("path contains null byte")
	ErrPathContainsControlChar = errors.New("path contains control character")
	ErrWindowsReservedName     = errors.New("path uses Windows reserved device name")
	ErrUNCPathNotAllowed       = errors.New("UNC paths not allowed")
	ErrDrivePathNotAllowed     = errors.New("windows drive paths not allowed")
	ErrInvalidUTF8Bytes        = errors.New("path contains invalid UTF-8 bytes")
	ErrOverlongUTF8Sequence    = errors.New("path contains overlong UTF-8 sequence")
	ErrInvalidUTF8Continuation = errors.New("path contains invalid UTF-8 continuation byte")
)

// DefaultPathValidator implements PathValidator
type DefaultPathValidator struct {
	mu    sync.RWMutex
	rules []ValidationRule
}

// NewPathValidator creates a new path validator
func NewPathValidator() *DefaultPathValidator {
	return &DefaultPathValidator{
		rules: make([]ValidationRule, 0),
	}
}

// Validation rules

// AddRule adds a validation rule
func (pv *DefaultPathValidator) AddRule(rule ValidationRule) error {
	if rule == nil {
		return ErrRuleCannotBeNil
	}

	pv.mu.Lock()
	defer pv.mu.Unlock()

	pv.rules = append(pv.rules, rule)
	return nil
}

// RemoveRule removes a validation rule by name
func (pv *DefaultPathValidator) RemoveRule(name string) error {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	for i, rule := range pv.rules {
		if rule.Name() == name {
			pv.rules = append(pv.rules[:i], pv.rules[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("%w: %q", ErrRuleNotFound, name)
}

// ClearRules removes all validation rules
func (pv *DefaultPathValidator) ClearRules() error {
	pv.mu.Lock()
	defer pv.mu.Unlock()

	pv.rules = make([]ValidationRule, 0)
	return nil
}

// Rules returns all validation rules
func (pv *DefaultPathValidator) Rules() []ValidationRule {
	pv.mu.RLock()
	defer pv.mu.RUnlock()

	result := make([]ValidationRule, len(pv.rules))
	copy(result, pv.rules)
	return result
}

// Validation operations

// Validate validates a path against all rules
func (pv *DefaultPathValidator) Validate(path string) []ValidationError {
	pv.mu.RLock()
	defer pv.mu.RUnlock()

	var errors []ValidationError

	for _, rule := range pv.rules {
		if err := rule.Validate(path); err != nil {
			errors = append(errors, ValidationError{
				Path:    path,
				Rule:    rule.Name(),
				Message: err.Error(),
				Code:    "VALIDATION_FAILED",
			})
		}
	}

	return errors
}

// ValidatePath validates a PathBuilder against all rules
func (pv *DefaultPathValidator) ValidatePath(path PathBuilder) []ValidationError {
	pv.mu.RLock()
	defer pv.mu.RUnlock()

	var errors []ValidationError

	for _, rule := range pv.rules {
		if err := rule.ValidatePath(path); err != nil {
			errors = append(errors, ValidationError{
				Path:    path.String(),
				Rule:    rule.Name(),
				Message: err.Error(),
				Code:    "VALIDATION_FAILED",
			})
		}
	}

	return errors
}

// IsValid returns true if the path passes all validation rules
func (pv *DefaultPathValidator) IsValid(path string) bool {
	return len(pv.Validate(path)) == 0
}

// IsValidPath returns true if the PathBuilder passes all validation rules
func (pv *DefaultPathValidator) IsValidPath(path PathBuilder) bool {
	return len(pv.ValidatePath(path)) == 0
}

// Built-in validators

// RequireAbsolute requires the path to be absolute
func (pv *DefaultPathValidator) RequireAbsolute() PathValidator {
	if err := pv.AddRule(&AbsolutePathRule{}); err != nil {
		// This shouldn't fail for built-in rules, but we continue anyway
		log.Printf("failed to add absolute path rule: %v", err)
	}
	return pv
}

// RequireRelative requires the path to be relative
func (pv *DefaultPathValidator) RequireRelative() PathValidator {
	if err := pv.AddRule(&RelativePathRule{}); err != nil {
		// This shouldn't fail for built-in rules, but we continue anyway
		log.Printf("failed to add relative path rule: %v", err)
	}
	return pv
}

// RequireExists requires the path to exist
func (pv *DefaultPathValidator) RequireExists() PathValidator {
	if err := pv.AddRule(&ExistsRule{}); err != nil {
		// This shouldn't fail for built-in rules, but we continue anyway
		log.Printf("failed to add exists rule: %v", err)
	}
	return pv
}

// RequireNotExists requires the path to not exist
func (pv *DefaultPathValidator) RequireNotExists() PathValidator {
	if err := pv.AddRule(&NotExistsRule{}); err != nil {
		// This shouldn't fail for built-in rules, but we continue anyway
		log.Printf("failed to add not exists rule: %v", err)
	}
	return pv
}

// RequireReadable requires the path to be readable
func (pv *DefaultPathValidator) RequireReadable() PathValidator {
	if err := pv.AddRule(&ReadableRule{}); err != nil {
		// This shouldn't fail for built-in rules, but we continue anyway
		log.Printf("failed to add readable rule: %v", err)
	}
	return pv
}

// RequireWritable requires the path to be writable
func (pv *DefaultPathValidator) RequireWritable() PathValidator {
	if err := pv.AddRule(&WritableRule{}); err != nil {
		// This shouldn't fail for built-in rules, but we continue anyway
		log.Printf("failed to add writable rule: %v", err)
	}
	return pv
}

// RequireExecutable requires the path to be executable
func (pv *DefaultPathValidator) RequireExecutable() PathValidator {
	if err := pv.AddRule(&ExecutableRule{}); err != nil {
		// This shouldn't fail for built-in rules, but we continue anyway
		log.Printf("failed to add executable rule: %v", err)
	}
	return pv
}

// RequireDirectory requires the path to be a directory
func (pv *DefaultPathValidator) RequireDirectory() PathValidator {
	if err := pv.AddRule(&DirectoryRule{}); err != nil {
		// This shouldn't fail for built-in rules, but we continue anyway
		log.Printf("failed to add directory rule: %v", err)
	}
	return pv
}

// RequireFile requires the path to be a file
func (pv *DefaultPathValidator) RequireFile() PathValidator {
	if err := pv.AddRule(&FileRule{}); err != nil {
		// This shouldn't fail for built-in rules, but we continue anyway
		log.Printf("failed to add file rule: %v", err)
	}
	return pv
}

// RequireExtension requires the path to have one of the specified extensions
func (pv *DefaultPathValidator) RequireExtension(exts ...string) PathValidator {
	if err := pv.AddRule(&ExtensionRule{Extensions: exts}); err != nil {
		// This shouldn't fail for built-in rules, but we continue anyway
		log.Printf("failed to add extension rule: %v", err)
	}
	return pv
}

// RequireMaxLength requires the path to be shorter than the specified length
func (pv *DefaultPathValidator) RequireMaxLength(length int) PathValidator {
	if err := pv.AddRule(&MaxLengthRule{MaxLength: length}); err != nil {
		// This shouldn't fail for built-in rules, but we continue anyway
		log.Printf("failed to add max length rule: %v", err)
	}
	return pv
}

// RequirePattern requires the path to match a pattern
func (pv *DefaultPathValidator) RequirePattern(pattern string) PathValidator {
	if err := pv.AddRule(&PatternRule{Pattern: pattern, Required: true}); err != nil {
		// This shouldn't fail for built-in rules, but we continue anyway
		log.Printf("failed to add required pattern rule: %v", err)
	}
	return pv
}

// ForbidPattern forbids the path from matching a pattern
func (pv *DefaultPathValidator) ForbidPattern(pattern string) PathValidator {
	if err := pv.AddRule(&PatternRule{Pattern: pattern, Required: false}); err != nil {
		// This shouldn't fail for built-in rules, but we continue anyway
		log.Printf("failed to add forbidden pattern rule: %v", err)
	}
	return pv
}

// Built-in validation rules

// AbsolutePathRule validates that a path is absolute
type AbsolutePathRule struct{}

// Name returns the rule name
func (r *AbsolutePathRule) Name() string { return "absolute-path" }

// Description returns the rule description
func (r *AbsolutePathRule) Description() string { return "path must be absolute" }

// Validate checks if the given path is absolute
func (r *AbsolutePathRule) Validate(path string) error {
	if !filepath.IsAbs(path) {
		return ErrPathMustBeAbsolute
	}
	return nil
}

// ValidatePath validates a PathBuilder instance
func (r *AbsolutePathRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// RelativePathRule validates that a path is relative
type RelativePathRule struct{}

// Name returns the name of the RelativePathRule
func (r *RelativePathRule) Name() string { return "relative-path" }

// Description returns the description of the RelativePathRule
func (r *RelativePathRule) Description() string { return "path must be relative" }

// Validate validates that the given path is relative
func (r *RelativePathRule) Validate(path string) error {
	if filepath.IsAbs(path) {
		return ErrPathMustBeRelative
	}
	return nil
}

// ValidatePath validates that the given PathBuilder represents a relative path
func (r *RelativePathRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// ExistsRule validates that a path exists
type ExistsRule struct{}

// Name returns the name of the ExistsRule
func (r *ExistsRule) Name() string { return "exists" }

// Description returns the description of the ExistsRule
func (r *ExistsRule) Description() string { return "path must exist" }

// Validate validates that the given path exists
func (r *ExistsRule) Validate(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ErrPathDoesNotExist
	}
	return nil
}

// ValidatePath validates that the given PathBuilder exists
func (r *ExistsRule) ValidatePath(path PathBuilder) error {
	if !path.Exists() {
		return ErrPathDoesNotExist
	}
	return nil
}

// NotExistsRule validates that a path does not exist
type NotExistsRule struct{}

// Name returns the name of the NotExistsRule
func (r *NotExistsRule) Name() string { return "not-exists" }

// Description returns the description of the NotExistsRule
func (r *NotExistsRule) Description() string { return "path must not exist" }

// Validate validates that the given path does not exist
func (r *NotExistsRule) Validate(path string) error {
	if _, err := os.Stat(path); err == nil {
		return ErrPathAlreadyExists
	}
	return nil
}

// ValidatePath validates that the given PathBuilder does not exist
func (r *NotExistsRule) ValidatePath(path PathBuilder) error {
	if path.Exists() {
		return ErrPathAlreadyExists
	}
	return nil
}

// ReadableRule validates that a path is readable
type ReadableRule struct{}

// Name returns the name of the ReadableRule
func (r *ReadableRule) Name() string { return "readable" }

// Description returns the description of the ReadableRule
func (r *ReadableRule) Description() string { return "path must be readable" }

// Validate validates that the given path is readable
func (r *ReadableRule) Validate(path string) error {
	// Clean path to prevent directory traversal
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return ErrPathTraversalDetected
	}

	file, err := os.Open(cleanPath)
	if err != nil {
		return fmt.Errorf("path is not readable: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}

// ValidatePath validates that the given PathBuilder path is readable
func (r *ReadableRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// WritableRule validates that a path is writable
type WritableRule struct{}

// Name returns the name of the WritableRule
func (r *WritableRule) Name() string { return "writable" }

// Description returns the description of the WritableRule
func (r *WritableRule) Description() string { return "path must be writable" }

// Validate validates that the given path is writable
func (r *WritableRule) Validate(path string) error {
	// Clean path to prevent directory traversal
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return ErrPathTraversalDetected
	}

	// Check if path exists
	info, err := os.Stat(cleanPath)
	if err != nil {
		// Path doesn't exist, check if parent directory is writable
		parentDir := filepath.Dir(cleanPath)
		return r.checkDirWritable(parentDir)
	}

	// Path exists, check if it's writable
	if info.IsDir() {
		return r.checkDirWritable(cleanPath)
	}

	// It's a file, try to open for writing
	file, err := os.OpenFile(cleanPath, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("path is not writable: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}

func (r *WritableRule) checkDirWritable(dir string) error {
	// Try to create a temporary file in the directory
	tempFile, err := os.CreateTemp(dir, "write_test_*")
	if err != nil {
		return fmt.Errorf("directory is not writable: %w", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	if err := os.Remove(tempFile.Name()); err != nil {
		return fmt.Errorf("failed to remove temp file: %w", err)
	}
	return nil
}

// ValidatePath validates that the given PathBuilder path is writable
func (r *WritableRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// ExecutableRule validates that a path is executable
type ExecutableRule struct{}

// Name returns the name of the ExecutableRule
func (r *ExecutableRule) Name() string { return "executable" }

// Description returns the description of the ExecutableRule
func (r *ExecutableRule) Description() string { return "path must be executable" }

// Validate validates that the given path is executable
func (r *ExecutableRule) Validate(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}

	if info.Mode()&0o111 == 0 {
		return errPathNotExecutable
	}

	return nil
}

// ValidatePath validates that the given PathBuilder path is executable
func (r *ExecutableRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// DirectoryRule validates that a path is a directory
type DirectoryRule struct{}

// Name returns the name of the DirectoryRule
func (r *DirectoryRule) Name() string { return "directory" }

// Description returns the description of the DirectoryRule
func (r *DirectoryRule) Description() string { return "path must be a directory" }

// Validate validates that the given path is a directory
func (r *DirectoryRule) Validate(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}

	if !info.IsDir() {
		return errPathNotDirectory
	}

	return nil
}

// ValidatePath validates that the given PathBuilder path is a directory
func (r *DirectoryRule) ValidatePath(path PathBuilder) error {
	if !path.IsDir() {
		return errPathNotDirectory
	}
	return nil
}

// FileRule validates that a path is a file
type FileRule struct{}

// Name returns the name of the FileRule
func (r *FileRule) Name() string { return "file" }

// Description returns the description of the FileRule
func (r *FileRule) Description() string { return "path must be a file" }

// Validate validates that the given path is a file
func (r *FileRule) Validate(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}

	if info.IsDir() {
		return errPathNotFile
	}

	return nil
}

// ValidatePath validates that the given PathBuilder path is a file
func (r *FileRule) ValidatePath(path PathBuilder) error {
	if !path.IsFile() {
		return errPathNotFile
	}
	return nil
}

// ExtensionRule validates that a path has one of the allowed extensions
type ExtensionRule struct {
	Extensions []string
}

// Name returns the name of the ExtensionRule
func (r *ExtensionRule) Name() string { return "extension" }

// Description returns the description of the ExtensionRule
func (r *ExtensionRule) Description() string { return "path must have allowed extension" }

// Validate validates that the given path has an allowed extension
func (r *ExtensionRule) Validate(path string) error {
	ext := strings.ToLower(filepath.Ext(path))

	for _, allowedExt := range r.Extensions {
		if !strings.HasPrefix(allowedExt, ".") {
			allowedExt = "." + allowedExt
		}
		if strings.EqualFold(allowedExt, ext) {
			return nil
		}
	}

	return fmt.Errorf("%w: %v", ErrInvalidExtensions, r.Extensions)
}

// ValidatePath validates that the given PathBuilder path has an allowed extension
func (r *ExtensionRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// MaxLengthRule validates that a path is shorter than the maximum length
type MaxLengthRule struct {
	MaxLength int
}

// Name returns the name of the MaxLengthRule
func (r *MaxLengthRule) Name() string { return "max-length" }

// Description returns the description of the MaxLengthRule
func (r *MaxLengthRule) Description() string { return "path must not exceed maximum length" }

// Validate validates that the given path does not exceed maximum length
func (r *MaxLengthRule) Validate(path string) error {
	if len(path) > r.MaxLength {
		return fmt.Errorf("%w: %d exceeds maximum %d", ErrPathTooLong, len(path), r.MaxLength)
	}
	return nil
}

// ValidatePath validates that the given PathBuilder path does not exceed maximum length
func (r *MaxLengthRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// PatternRule validates that a path matches or doesn't match a pattern
type PatternRule struct {
	Pattern  string
	Required bool // true = must match, false = must not match
}

// Name returns the name of the pattern rule.
func (r *PatternRule) Name() string {
	if r.Required {
		return "require-pattern"
	}
	return "forbid-pattern"
}

// Description returns the description of the pattern rule.
func (r *PatternRule) Description() string {
	if r.Required {
		return "path must match pattern"
	}
	return "path must not match pattern"
}

// Validate validates that the given path matches or does not match the pattern
func (r *PatternRule) Validate(path string) error {
	matched, err := regexp.MatchString(r.Pattern, path)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	if r.Required && !matched {
		return fmt.Errorf("%w: %s", ErrPatternMismatch, r.Pattern)
	}

	if !r.Required && matched {
		return fmt.Errorf("%w: %s", ErrForbiddenPattern, r.Pattern)
	}

	return nil
}

// ValidatePath validates that the given PathBuilder path matches or does not match the pattern
func (r *PatternRule) ValidatePath(path PathBuilder) error {
	// For security patterns (like path traversal), check the original path
	// This prevents bypasses where cleaning removes dangerous patterns
	if r.isSecurityPattern() {
		originalErr := r.Validate(path.Original())
		if originalErr != nil {
			return originalErr
		}
	}

	// Always check the cleaned path as well
	return r.Validate(path.String())
}

// isSecurityPattern checks if this is a security-related pattern that should check original paths
func (r *PatternRule) isSecurityPattern() bool {
	// Common security patterns that attackers try to bypass through path cleaning
	securityPatterns := []string{
		`\.\.`,     // Path traversal
		`\.\./`,    // Path traversal with slash
		`/\.\./`,   // Path traversal with leading slash
		`%2e%2e`,   // URL encoded ..
		`%252e`,    // Double URL encoded
		`\x2e\x2e`, // Hex encoded ..
	}

	for _, pattern := range securityPatterns {
		if r.Pattern == pattern {
			return true
		}
	}
	return false
}

// Package-level convenience functions

// Validate validates a path with a validator
func Validate(path string, validator PathValidator) []ValidationError {
	return validator.Validate(path)
}

// IsValid returns true if a path is valid according to the validator
func IsValid(path string, validator PathValidator) bool {
	return validator.IsValid(path)
}

// ValidateExists returns an error if the path doesn't exist
func ValidateExists(path string) error {
	rule := &ExistsRule{}
	return rule.Validate(path)
}

// ValidateReadable returns an error if the path isn't readable
func ValidateReadable(path string) error {
	rule := &ReadableRule{}
	return rule.Validate(path)
}

// ValidateWritable returns an error if the path isn't writable
func ValidateWritable(path string) error {
	rule := &WritableRule{}
	return rule.Validate(path)
}

// ValidateExtension returns an error if the path doesn't have the required extension
func ValidateExtension(path string, extensions ...string) error {
	rule := &ExtensionRule{Extensions: extensions}
	return rule.Validate(path)
}

// Additional security validation rules

// PathTraversalRule validates that a path doesn't contain path traversal patterns
type PathTraversalRule struct{}

// Name returns the rule name
func (r *PathTraversalRule) Name() string { return "no-path-traversal" }

// Description returns the rule description
func (r *PathTraversalRule) Description() string {
	return "path must not contain path traversal patterns"
}

// Validate checks if the given path contains path traversal patterns
func (r *PathTraversalRule) Validate(path string) error {
	// Check for various path traversal patterns
	if strings.Contains(path, "..") {
		return ErrPathTraversalDetected
	}

	// Check URL encoded patterns
	if strings.Contains(path, "%2e%2e") || strings.Contains(path, "%252e%252e") {
		return ErrPathTraversalDetected
	}

	// Check Unicode encoded patterns
	if strings.Contains(path, "\u002e\u002e") {
		return ErrPathTraversalDetected
	}

	// Check hex encoded patterns
	if strings.Contains(path, "\x2e\x2e") {
		return ErrPathTraversalDetected
	}

	return nil
}

// ValidatePath validates a PathBuilder instance
func (r *PathTraversalRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// NullByteRule validates that a path doesn't contain null bytes
type NullByteRule struct{}

// Name returns the rule name
func (r *NullByteRule) Name() string { return "no-null-bytes" }

// Description returns the rule description
func (r *NullByteRule) Description() string { return "path must not contain null bytes" }

// Validate checks if the given path contains null bytes
func (r *NullByteRule) Validate(path string) error {
	if strings.Contains(path, "\x00") || strings.Contains(path, "%00") {
		return ErrPathContainsNullByte
	}
	return nil
}

// ValidatePath validates a PathBuilder instance
func (r *NullByteRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// ControlCharacterRule validates that a path doesn't contain control characters
type ControlCharacterRule struct{}

// Name returns the rule name
func (r *ControlCharacterRule) Name() string { return "no-control-chars" }

// Description returns the rule description
func (r *ControlCharacterRule) Description() string {
	return "path must not contain control characters"
}

// Validate checks if the given path contains control characters
func (r *ControlCharacterRule) Validate(path string) error {
	for _, char := range path {
		if char < 32 && char != '\t' { // Allow tabs
			return ErrPathContainsControlChar
		}
	}
	return nil
}

// ValidatePath validates a PathBuilder instance
func (r *ControlCharacterRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// WindowsReservedRule validates that a path doesn't use Windows reserved names
type WindowsReservedRule struct{}

// Name returns the rule name
func (r *WindowsReservedRule) Name() string { return "no-windows-reserved" }

// Description returns the rule description
func (r *WindowsReservedRule) Description() string {
	return "path must not use Windows reserved device names"
}

// Validate checks if the given path uses Windows reserved names
func (r *WindowsReservedRule) Validate(path string) error {
	reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9", "CONIN$", "CONOUT$"}
	baseName := strings.ToUpper(filepath.Base(path))
	// Remove extension for checking
	if idx := strings.LastIndex(baseName, "."); idx > 0 {
		baseName = baseName[:idx]
	}
	for _, reserved := range reservedNames {
		if baseName == reserved {
			return ErrWindowsReservedName
		}
	}
	return nil
}

// ValidatePath validates a PathBuilder instance
func (r *WindowsReservedRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// UNCPathRule validates that a path doesn't use UNC paths
type UNCPathRule struct{}

// Name returns the rule name
func (r *UNCPathRule) Name() string { return "no-unc-paths" }

// Description returns the rule description
func (r *UNCPathRule) Description() string { return "path must not use UNC paths" }

// Validate checks if the given path is a UNC path
func (r *UNCPathRule) Validate(path string) error {
	if strings.HasPrefix(path, "\\\\") {
		return ErrUNCPathNotAllowed
	}
	return nil
}

// ValidatePath validates a PathBuilder instance
func (r *UNCPathRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// DrivePathRule validates that a path doesn't use Windows drive paths
type DrivePathRule struct{}

// Name returns the rule name
func (r *DrivePathRule) Name() string { return "no-drive-paths" }

// Description returns the rule description
func (r *DrivePathRule) Description() string { return "path must not use Windows drive paths" }

// Validate checks if the given path is a Windows drive path
func (r *DrivePathRule) Validate(path string) error {
	if len(path) > 1 && path[1] == ':' {
		return ErrDrivePathNotAllowed
	}
	return nil
}

// ValidatePath validates a PathBuilder instance
func (r *DrivePathRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// ValidUTF8Rule validates that a path contains valid UTF-8
type ValidUTF8Rule struct{}

// Name returns the rule name
func (r *ValidUTF8Rule) Name() string { return "valid-utf8" }

// Description returns the rule description
func (r *ValidUTF8Rule) Description() string { return "path must contain valid UTF-8" }

// Validate checks if the given path contains valid UTF-8
func (r *ValidUTF8Rule) Validate(path string) error {
	// Check for invalid UTF-8 byte sequences
	for i := 0; i < len(path); i++ {
		b := path[i]
		// Check for invalid UTF-8 bytes
		if b == 0xff || b == 0xfe {
			return ErrInvalidUTF8Bytes
		}
		// Check for overlong UTF-8 sequences (like \xc0\xaf for "/")
		if b == 0xc0 || b == 0xc1 {
			return ErrOverlongUTF8Sequence
		}
		// Check for other suspicious byte sequences
		if b >= 0x80 && b <= 0xbf && (i == 0 || path[i-1] < 0x80) {
			return ErrInvalidUTF8Continuation
		}
	}
	return nil
}

// ValidatePath validates a PathBuilder instance
func (r *ValidUTF8Rule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// RequireSecure adds comprehensive security validation rules
func (pv *DefaultPathValidator) RequireSecure() PathValidator {
	// Add all security rules
	if err := pv.AddRule(&PathTraversalRule{}); err != nil {
		log.Printf("failed to add path traversal rule: %v", err)
	}
	if err := pv.AddRule(&NullByteRule{}); err != nil {
		log.Printf("failed to add null byte rule: %v", err)
	}
	if err := pv.AddRule(&ControlCharacterRule{}); err != nil {
		log.Printf("failed to add control character rule: %v", err)
	}
	if err := pv.AddRule(&WindowsReservedRule{}); err != nil {
		log.Printf("failed to add Windows reserved rule: %v", err)
	}
	if err := pv.AddRule(&UNCPathRule{}); err != nil {
		log.Printf("failed to add UNC path rule: %v", err)
	}
	if err := pv.AddRule(&DrivePathRule{}); err != nil {
		log.Printf("failed to add drive path rule: %v", err)
	}
	if err := pv.AddRule(&ValidUTF8Rule{}); err != nil {
		log.Printf("failed to add valid UTF-8 rule: %v", err)
	}
	return pv
}
