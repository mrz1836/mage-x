package env

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DefaultEnvManager implements EnvManager
type DefaultEnvManager struct {
	mu      sync.RWMutex
	scopes  []EnvScope
	baseEnv Environment
}

// NewDefaultEnvManager creates a new environment manager
func NewDefaultEnvManager() *DefaultEnvManager {
	return &DefaultEnvManager{
		scopes:  make([]EnvScope, 0),
		baseEnv: NewDefaultEnvironment(),
	}
}

// PushScope creates and pushes a new environment scope
func (m *DefaultEnvManager) PushScope() EnvScope {
	m.mu.Lock()
	defer m.mu.Unlock()

	scope := NewDefaultEnvScope(m.baseEnv)
	m.scopes = append(m.scopes, scope)
	return scope
}

// PopScope removes the top scope
func (m *DefaultEnvManager) PopScope() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.scopes) == 0 {
		return fmt.Errorf("no scopes to pop")
	}

	// Remove the last scope
	m.scopes = m.scopes[:len(m.scopes)-1]
	return nil
}

// WithScope executes a function within a new scope
func (m *DefaultEnvManager) WithScope(fn func(EnvScope) error) error {
	scope := m.PushScope()
	defer func() {
		if err := m.PopScope(); err != nil {
			// Log the error but don't fail the operation
			// as this is in a defer block
		}
	}()

	return fn(scope)
}

// SaveContext saves the current environment state
func (m *DefaultEnvManager) SaveContext() (EnvContext, error) {
	vars := m.baseEnv.GetAll()
	return &DefaultEnvContext{
		timestamp: time.Now(),
		variables: vars,
	}, nil
}

// RestoreContext restores a saved environment state
func (m *DefaultEnvManager) RestoreContext(ctx EnvContext) error {
	// Clear current environment
	current := m.baseEnv.GetAll()
	for key := range current {
		if err := m.baseEnv.Unset(key); err != nil {
			return fmt.Errorf("failed to unset %s: %w", key, err)
		}
	}

	// Restore saved variables
	saved := ctx.Export()
	return m.baseEnv.SetMultiple(saved)
}

// Isolate runs a function with isolated environment variables
func (m *DefaultEnvManager) Isolate(vars map[string]string, fn func() error) error {
	// Save current context
	saved, err := m.SaveContext()
	if err != nil {
		return fmt.Errorf("failed to save context: %w", err)
	}

	// Set isolated variables
	if setErr := m.baseEnv.SetMultiple(vars); setErr != nil {
		return fmt.Errorf("failed to set isolated variables: %w", setErr)
	}

	// Execute function
	err = fn()

	// Restore original context
	if restoreErr := m.RestoreContext(saved); restoreErr != nil {
		return fmt.Errorf("failed to restore context: %w (original error: %w)", restoreErr, err)
	}

	return err
}

// Fork creates a new manager with the same base environment
func (m *DefaultEnvManager) Fork() EnvManager {
	return NewDefaultEnvManager()
}

// DefaultEnvScope implements EnvScope
type DefaultEnvScope struct {
	Environment
	mu       sync.RWMutex
	changes  map[string]EnvChange
	original map[string]string
	baseEnv  Environment
}

// NewDefaultEnvScope creates a new environment scope
func NewDefaultEnvScope(baseEnv Environment) *DefaultEnvScope {
	return &DefaultEnvScope{
		Environment: baseEnv,
		changes:     make(map[string]EnvChange),
		original:    baseEnv.GetAll(),
		baseEnv:     baseEnv,
	}
}

// Set overrides Environment.Set to track changes
func (s *DefaultEnvScope) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldValue := s.baseEnv.Get(key)

	if err := s.baseEnv.Set(key, value); err != nil {
		return err
	}

	var action ChangeAction
	if oldValue == "" {
		action = ActionSet
	} else {
		action = ActionModify
	}

	s.changes[key] = EnvChange{
		Key:      key,
		OldValue: oldValue,
		NewValue: value,
		Action:   action,
	}

	return nil
}

// Unset overrides Environment.Unset to track changes
func (s *DefaultEnvScope) Unset(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldValue := s.baseEnv.Get(key)

	if err := s.baseEnv.Unset(key); err != nil {
		return err
	}

	s.changes[key] = EnvChange{
		Key:      key,
		OldValue: oldValue,
		NewValue: "",
		Action:   ActionUnset,
	}

	return nil
}

// Commit applies all changes permanently
func (s *DefaultEnvScope) Commit() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Changes are already applied to the base environment
	// Clear the changes to mark them as committed
	s.changes = make(map[string]EnvChange)
	return nil
}

// Rollback reverts all changes
func (s *DefaultEnvScope) Rollback() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Revert all changes
	for key, change := range s.changes {
		switch change.Action {
		case ActionSet, ActionModify:
			if change.OldValue == "" {
				if err := s.baseEnv.Unset(key); err != nil {
					// Log error but continue reverting other changes
				}
			} else {
				if err := s.baseEnv.Set(key, change.OldValue); err != nil {
					// Log error but continue reverting other changes
				}
			}
		case ActionUnset:
			if err := s.baseEnv.Set(key, change.OldValue); err != nil {
				// Log error but continue reverting other changes
			}
		}
	}

	s.changes = make(map[string]EnvChange)
	return nil
}

// Changes returns all changes made in this scope
func (s *DefaultEnvScope) Changes() map[string]EnvChange {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]EnvChange)
	for k, v := range s.changes {
		result[k] = v
	}
	return result
}

// HasChanges returns true if there are uncommitted changes
func (s *DefaultEnvScope) HasChanges() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.changes) > 0
}

// DefaultEnvContext implements EnvContext
type DefaultEnvContext struct {
	timestamp time.Time
	variables map[string]string
}

// Timestamp returns when the context was created
func (c *DefaultEnvContext) Timestamp() time.Time {
	return c.timestamp
}

// Variables returns all variables in the context
func (c *DefaultEnvContext) Variables() map[string]string {
	result := make(map[string]string)
	for k, v := range c.variables {
		result[k] = v
	}
	return result
}

// Count returns the number of variables
func (c *DefaultEnvContext) Count() int {
	return len(c.variables)
}

// Diff compares this context with another
func (c *DefaultEnvContext) Diff(other EnvContext) map[string]EnvChange {
	otherVars := other.Variables()
	changes := make(map[string]EnvChange)

	// Check for additions and modifications
	for key, value := range c.variables {
		if otherValue, exists := otherVars[key]; !exists {
			changes[key] = EnvChange{
				Key:      key,
				OldValue: "",
				NewValue: value,
				Action:   ActionSet,
			}
		} else if otherValue != value {
			changes[key] = EnvChange{
				Key:      key,
				OldValue: otherValue,
				NewValue: value,
				Action:   ActionModify,
			}
		}
	}

	// Check for deletions
	for key, value := range otherVars {
		if _, exists := c.variables[key]; !exists {
			changes[key] = EnvChange{
				Key:      key,
				OldValue: value,
				NewValue: "",
				Action:   ActionUnset,
			}
		}
	}

	return changes
}

// Merge merges this context with another
func (c *DefaultEnvContext) Merge(other EnvContext) EnvContext {
	merged := make(map[string]string)

	// Start with this context
	for k, v := range c.variables {
		merged[k] = v
	}

	// Add/override with other context
	for k, v := range other.Variables() {
		merged[k] = v
	}

	return &DefaultEnvContext{
		timestamp: time.Now(),
		variables: merged,
	}
}

// Export returns all variables as a map
func (c *DefaultEnvContext) Export() map[string]string {
	return c.Variables()
}

// DefaultEnvValidator implements EnvValidator
type DefaultEnvValidator struct {
	mu    sync.RWMutex
	rules map[string][]ValidationRule
}

// NewDefaultEnvValidator creates a new environment validator
func NewDefaultEnvValidator() *DefaultEnvValidator {
	return &DefaultEnvValidator{
		rules: make(map[string][]ValidationRule),
	}
}

// AddRule adds a validation rule
func (v *DefaultEnvValidator) AddRule(key string, rule ValidationRule) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if v.rules[key] == nil {
		v.rules[key] = make([]ValidationRule, 0)
	}

	v.rules[key] = append(v.rules[key], rule)
	return nil
}

// RemoveRule removes validation rules for a key
func (v *DefaultEnvValidator) RemoveRule(key string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	delete(v.rules, key)
	return nil
}

// ValidateAll validates all environment variables with rules
func (v *DefaultEnvValidator) ValidateAll() []ValidationError {
	v.mu.RLock()
	defer v.mu.RUnlock()

	var errors []ValidationError
	env := NewDefaultEnvironment()

	for key, rules := range v.rules {
		value := env.Get(key)
		for _, rule := range rules {
			if err := rule.Validate(value); err != nil {
				errors = append(errors, ValidationError{
					Key:     key,
					Value:   value,
					Rule:    rule.Description(),
					Message: err.Error(),
				})
			}
		}
	}

	return errors
}

// Validate validates a specific key-value pair
func (v *DefaultEnvValidator) Validate(key, value string) error {
	v.mu.RLock()
	defer v.mu.RUnlock()

	rules, exists := v.rules[key]
	if !exists {
		return nil // No rules for this key
	}

	for _, rule := range rules {
		if err := rule.Validate(value); err != nil {
			return err
		}
	}

	return nil
}

// Required adds required validation for keys
func (v *DefaultEnvValidator) Required(keys ...string) EnvValidator {
	for _, key := range keys {
		if err := v.AddRule(key, &RequiredRule{}); err != nil {
			// Log error but continue with other keys
		}
	}
	return v
}

// NotEmpty adds not-empty validation for keys
func (v *DefaultEnvValidator) NotEmpty(keys ...string) EnvValidator {
	for _, key := range keys {
		if err := v.AddRule(key, &NotEmptyRule{}); err != nil {
			// Log error but continue with other keys
		}
	}
	return v
}

// Pattern adds pattern validation for a key
func (v *DefaultEnvValidator) Pattern(key, pattern string) EnvValidator {
	if err := v.AddRule(key, &PatternRule{Pattern: pattern}); err != nil {
		// Log error but continue
	}
	return v
}

// Range adds range validation for a key
func (v *DefaultEnvValidator) Range(key string, minValue, maxValue interface{}) EnvValidator {
	if err := v.AddRule(key, &RangeRule{Min: minValue, Max: maxValue}); err != nil {
		// Log error but continue
	}
	return v
}

// OneOf adds one-of validation for a key
func (v *DefaultEnvValidator) OneOf(key string, values ...string) EnvValidator {
	if err := v.AddRule(key, &OneOfRule{Values: values}); err != nil {
		// Log error but continue
	}
	return v
}

// Built-in validation rules

// RequiredRule validates that a value is present
type RequiredRule struct{}

func (r *RequiredRule) Validate(value string) error {
	if value == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

func (r *RequiredRule) Description() string {
	return "required"
}

// NotEmptyRule validates that a value is not empty
type NotEmptyRule struct{}

func (r *NotEmptyRule) Validate(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("value cannot be empty")
	}
	return nil
}

func (r *NotEmptyRule) Description() string {
	return "not-empty"
}

// PatternRule validates that a value matches a regex pattern
type PatternRule struct {
	Pattern string
	regex   *regexp.Regexp
}

func (r *PatternRule) Validate(value string) error {
	if r.regex == nil {
		var err error
		r.regex, err = regexp.Compile(r.Pattern)
		if err != nil {
			return fmt.Errorf("invalid pattern: %w", err)
		}
	}

	if !r.regex.MatchString(value) {
		return fmt.Errorf("value does not match pattern %s", r.Pattern)
	}

	return nil
}

func (r *PatternRule) Description() string {
	return fmt.Sprintf("pattern: %s", r.Pattern)
}

// RangeRule validates that a numeric value is within range
type RangeRule struct {
	Min interface{}
	Max interface{}
}

func (r *RangeRule) Validate(value string) error {
	if value == "" {
		return nil // Allow empty values, use Required rule to enforce non-empty
	}

	// Try to parse as int
	if intVal, err := strconv.Atoi(value); err == nil {
		if minInt, ok := r.Min.(int); ok && intVal < minInt {
			return fmt.Errorf("value %d is less than minimum %d", intVal, minInt)
		}
		if maxInt, ok := r.Max.(int); ok && intVal > maxInt {
			return fmt.Errorf("value %d is greater than maximum %d", intVal, maxInt)
		}
		return nil
	}

	// Try to parse as float
	if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
		if minFloat, ok := r.Min.(float64); ok && floatVal < minFloat {
			return fmt.Errorf("value %f is less than minimum %f", floatVal, minFloat)
		}
		if maxFloat, ok := r.Max.(float64); ok && floatVal > maxFloat {
			return fmt.Errorf("value %f is greater than maximum %f", floatVal, maxFloat)
		}
		return nil
	}

	return fmt.Errorf("value is not numeric")
}

func (r *RangeRule) Description() string {
	return fmt.Sprintf("range: %v-%v", r.Min, r.Max)
}

// OneOfRule validates that a value is one of the allowed values
type OneOfRule struct {
	Values []string
}

func (r *OneOfRule) Validate(value string) error {
	for _, allowed := range r.Values {
		if value == allowed {
			return nil
		}
	}

	sort.Strings(r.Values) // For consistent error messages
	return fmt.Errorf("value must be one of: %v", r.Values)
}

func (r *OneOfRule) Description() string {
	return fmt.Sprintf("one-of: %v", r.Values)
}
