package config

import (
	"fmt"
	"sort"
	"sync"
)

// DefaultConfigManager implements ConfigManager
type DefaultConfigManager struct {
	mu        sync.RWMutex
	sources   []ConfigSource
	validator Validator
	watching  bool
	stopWatch chan bool
}

// NewDefaultConfigManager creates a new default config manager
func NewDefaultConfigManager() *DefaultConfigManager {
	return &DefaultConfigManager{
		sources:   make([]ConfigSource, 0),
		stopWatch: make(chan bool, 1),
	}
}

// AddSource adds a configuration source
func (m *DefaultConfigManager) AddSource(source ConfigSource) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sources = append(m.sources, source)

	// Sort sources by priority (highest first)
	sort.Slice(m.sources, func(i, j int) bool {
		return m.sources[i].Priority() > m.sources[j].Priority()
	})
}

// LoadConfig loads configuration from all sources in priority order
func (m *DefaultConfigManager) LoadConfig(dest interface{}) error {
	m.mu.RLock()
	sources := make([]ConfigSource, len(m.sources))
	copy(sources, m.sources)
	m.mu.RUnlock()

	var errors []error
	loaded := false

	for _, source := range sources {
		if !source.IsAvailable() {
			continue
		}

		if err := source.Load(dest); err != nil {
			errors = append(errors, fmt.Errorf("failed to load from %s: %w", source.Name(), err))
			continue
		}

		loaded = true
		// In a real implementation, we might merge configurations instead of using the first successful one
		break
	}

	if !loaded {
		if len(errors) > 0 {
			return fmt.Errorf("failed to load configuration from any source: %v", errors)
		}
		return fmt.Errorf("no configuration sources available")
	}

	// Validate if validator is set
	if m.validator != nil {
		if err := m.validator.Validate(dest); err != nil {
			return fmt.Errorf("configuration validation failed: %w", err)
		}
	}

	return nil
}

// Reload reloads configuration from all sources
func (m *DefaultConfigManager) Reload(dest interface{}) error {
	return m.LoadConfig(dest)
}

// Watch watches for configuration changes (basic implementation)
func (m *DefaultConfigManager) Watch(callback func(interface{})) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.watching {
		return fmt.Errorf("already watching for configuration changes")
	}

	m.watching = true

	// This is a simplified implementation
	// In a real implementation, this would use filesystem watchers and other mechanisms
	go func() {
		<-m.stopWatch
		m.mu.Lock()
		m.watching = false
		m.mu.Unlock()
	}()

	return nil
}

// StopWatching stops watching for configuration changes
func (m *DefaultConfigManager) StopWatching() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.watching {
		select {
		case m.stopWatch <- true:
		default:
		}
	}
}

// GetActiveSources returns list of currently active sources
func (m *DefaultConfigManager) GetActiveSources() []ConfigSource {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := make([]ConfigSource, 0)
	for _, source := range m.sources {
		if source.IsAvailable() {
			active = append(active, source)
		}
	}

	return active
}

// SetValidator sets the configuration validator
func (m *DefaultConfigManager) SetValidator(validator Validator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.validator = validator
}

// BasicValidator implements Validator with basic validation rules
type BasicValidator struct {
	rules map[string]interface{}
}

// NewBasicValidator creates a new basic validator
func NewBasicValidator() *BasicValidator {
	return &BasicValidator{
		rules: make(map[string]interface{}),
	}
}

// Validate validates configuration data
func (v *BasicValidator) Validate(data interface{}) error {
	if data == nil {
		return fmt.Errorf("configuration data cannot be nil")
	}

	// Basic validation - can be extended with more sophisticated rules
	return nil
}

// ValidateField validates a specific field
func (v *BasicValidator) ValidateField(fieldName string, value interface{}) error {
	rule, exists := v.rules[fieldName]
	if !exists {
		return nil // No rule for this field
	}

	// Simple validation based on rule type
	switch r := rule.(type) {
	case func(interface{}) error:
		return r(value)
	default:
		return nil
	}
}

// GetValidationRules returns current validation rules
func (v *BasicValidator) GetValidationRules() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range v.rules {
		result[k] = v
	}
	return result
}

// SetValidationRules sets validation rules
func (v *BasicValidator) SetValidationRules(rules map[string]interface{}) {
	v.rules = make(map[string]interface{})
	for k, val := range rules {
		v.rules[k] = val
	}
}
