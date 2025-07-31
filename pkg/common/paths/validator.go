package paths

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
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
		return fmt.Errorf("rule cannot be nil")
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

	return fmt.Errorf("rule %q not found", name)
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
		return fmt.Errorf("path must be absolute")
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
		return fmt.Errorf("path must be relative")
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
		return fmt.Errorf("path does not exist")
	}
	return nil
}

// ValidatePath validates that the given PathBuilder exists
func (r *ExistsRule) ValidatePath(path PathBuilder) error {
	if !path.Exists() {
		return fmt.Errorf("path does not exist")
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
		return fmt.Errorf("path already exists")
	}
	return nil
}

// ValidatePath validates that the given PathBuilder does not exist
func (r *NotExistsRule) ValidatePath(path PathBuilder) error {
	if path.Exists() {
		return fmt.Errorf("path already exists")
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
		return fmt.Errorf("invalid path: path traversal detected")
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

func (r *ReadableRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// WritableRule validates that a path is writable
type WritableRule struct{}

func (r *WritableRule) Name() string        { return "writable" }
func (r *WritableRule) Description() string { return "path must be writable" }

func (r *WritableRule) Validate(path string) error {
	// Clean path to prevent directory traversal
	cleanPath := filepath.Clean(path)
	if strings.Contains(cleanPath, "..") {
		return fmt.Errorf("invalid path: path traversal detected")
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

func (r *WritableRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// ExecutableRule validates that a path is executable
type ExecutableRule struct{}

func (r *ExecutableRule) Name() string        { return "executable" }
func (r *ExecutableRule) Description() string { return "path must be executable" }

func (r *ExecutableRule) Validate(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}

	if info.Mode()&0o111 == 0 {
		return fmt.Errorf("path is not executable")
	}

	return nil
}

func (r *ExecutableRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// DirectoryRule validates that a path is a directory
type DirectoryRule struct{}

func (r *DirectoryRule) Name() string        { return "directory" }
func (r *DirectoryRule) Description() string { return "path must be a directory" }

func (r *DirectoryRule) Validate(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	return nil
}

func (r *DirectoryRule) ValidatePath(path PathBuilder) error {
	if !path.IsDir() {
		return fmt.Errorf("path is not a directory")
	}
	return nil
}

// FileRule validates that a path is a file
type FileRule struct{}

func (r *FileRule) Name() string        { return "file" }
func (r *FileRule) Description() string { return "path must be a file" }

func (r *FileRule) Validate(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("path does not exist: %w", err)
	}

	if info.IsDir() {
		return fmt.Errorf("path is not a file")
	}

	return nil
}

func (r *FileRule) ValidatePath(path PathBuilder) error {
	if !path.IsFile() {
		return fmt.Errorf("path is not a file")
	}
	return nil
}

// ExtensionRule validates that a path has one of the allowed extensions
type ExtensionRule struct {
	Extensions []string
}

func (r *ExtensionRule) Name() string        { return "extension" }
func (r *ExtensionRule) Description() string { return "path must have allowed extension" }

func (r *ExtensionRule) Validate(path string) error {
	ext := strings.ToLower(filepath.Ext(path))

	for _, allowedExt := range r.Extensions {
		if !strings.HasPrefix(allowedExt, ".") {
			allowedExt = "." + allowedExt
		}
		if strings.ToLower(allowedExt) == ext {
			return nil
		}
	}

	return fmt.Errorf("path must have one of these extensions: %v", r.Extensions)
}

func (r *ExtensionRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// MaxLengthRule validates that a path is shorter than the maximum length
type MaxLengthRule struct {
	MaxLength int
}

func (r *MaxLengthRule) Name() string        { return "max-length" }
func (r *MaxLengthRule) Description() string { return "path must not exceed maximum length" }

func (r *MaxLengthRule) Validate(path string) error {
	if len(path) > r.MaxLength {
		return fmt.Errorf("path length %d exceeds maximum %d", len(path), r.MaxLength)
	}
	return nil
}

func (r *MaxLengthRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
}

// PatternRule validates that a path matches or doesn't match a pattern
type PatternRule struct {
	Pattern  string
	Required bool // true = must match, false = must not match
}

func (r *PatternRule) Name() string {
	if r.Required {
		return "require-pattern"
	}
	return "forbid-pattern"
}

func (r *PatternRule) Description() string {
	if r.Required {
		return "path must match pattern"
	}
	return "path must not match pattern"
}

func (r *PatternRule) Validate(path string) error {
	matched, err := regexp.MatchString(r.Pattern, path)
	if err != nil {
		return fmt.Errorf("invalid pattern: %w", err)
	}

	if r.Required && !matched {
		return fmt.Errorf("path does not match required pattern %s", r.Pattern)
	}

	if !r.Required && matched {
		return fmt.Errorf("path matches forbidden pattern %s", r.Pattern)
	}

	return nil
}

func (r *PatternRule) ValidatePath(path PathBuilder) error {
	return r.Validate(path.String())
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
