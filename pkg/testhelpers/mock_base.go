// Package testhelpers provides generic mock base framework to reduce duplication
package testhelpers

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// CallRecord represents a recorded method call for verification
type CallRecord struct {
	Method    string
	Args      []any
	Result    []any
	Timestamp time.Time
	Error     error
}

// MockBase provides common mock functionality to reduce duplication
type MockBase struct {
	mu          sync.RWMutex
	t           *testing.T
	calls       []CallRecord
	shouldError bool
	errorMap    map[string]error // method name -> error to return
	callCount   map[string]int   // method name -> call count
}

// NewMockBase creates a new mock base with common functionality
func NewMockBase(t *testing.T) *MockBase {
	return &MockBase{
		t:         t,
		calls:     make([]CallRecord, 0),
		errorMap:  make(map[string]error),
		callCount: make(map[string]int),
	}
}

// SetError configures the mock to return errors for all methods
func (m *MockBase) SetError(shouldError bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = shouldError
}

// SetMethodError configures the mock to return a specific error for a method
func (m *MockBase) SetMethodError(method string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.errorMap[method] = err
}

// ClearMethodError removes error configuration for a method
func (m *MockBase) ClearMethodError(method string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.errorMap, method)
}

// ClearAllErrors removes all error configurations
func (m *MockBase) ClearAllErrors() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.shouldError = false
	m.errorMap = make(map[string]error)
}

// RecordCall records a method call for later verification
func (m *MockBase) RecordCall(method string, args, result []any, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	record := CallRecord{
		Method:    method,
		Args:      args,
		Result:    result,
		Timestamp: time.Now(),
		Error:     err,
	}

	m.calls = append(m.calls, record)
	m.callCount[method]++
}

// ShouldReturnError checks if the mock should return an error for the given method
func (m *MockBase) ShouldReturnError(method string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check for method-specific error first
	if err, exists := m.errorMap[method]; exists {
		return err
	}

	// Check for global error flag
	if m.shouldError {
		return fmt.Errorf("mock error for %s", method) //nolint:err113 // Dynamic errors are appropriate for test mocks
	}

	return nil
}

// GetCalls returns all recorded calls (thread-safe copy)
func (m *MockBase) GetCalls() []CallRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	calls := make([]CallRecord, len(m.calls))
	copy(calls, m.calls)
	return calls
}

// GetCallsForMethod returns calls for a specific method
func (m *MockBase) GetCallsForMethod(method string) []CallRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var methodCalls []CallRecord
	for _, call := range m.calls {
		if call.Method == method {
			methodCalls = append(methodCalls, call)
		}
	}
	return methodCalls
}

// CallCount returns the number of times a method was called
func (m *MockBase) CallCount(method string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.callCount[method]
}

// TotalCalls returns the total number of method calls
func (m *MockBase) TotalCalls() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.calls)
}

// Reset clears all recorded calls and error configurations
func (m *MockBase) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls = make([]CallRecord, 0)
	m.callCount = make(map[string]int)
	m.shouldError = false
	m.errorMap = make(map[string]error)
}

// AssertCalled verifies that a method was called
func (m *MockBase) AssertCalled(method string) {
	if m.t == nil {
		return
	}
	m.t.Helper()

	count := m.CallCount(method)
	if count == 0 {
		m.t.Errorf("Expected method %s to be called, but it wasn't", method)
	}
}

// AssertCalledTimes verifies that a method was called a specific number of times
func (m *MockBase) AssertCalledTimes(method string, expectedTimes int) {
	if m.t == nil {
		return
	}
	m.t.Helper()

	count := m.CallCount(method)
	if count != expectedTimes {
		m.t.Errorf("Expected method %s to be called %d times, but was called %d times", method, expectedTimes, count)
	}
}

// AssertCalledWith verifies that a method was called with specific arguments
func (m *MockBase) AssertCalledWith(method string, expectedArgs ...any) {
	if m.t == nil {
		return
	}
	m.t.Helper()

	calls := m.GetCallsForMethod(method)
	if len(calls) == 0 {
		m.t.Errorf("Expected method %s to be called, but it wasn't", method)
		return
	}

	// Check if any call matches the expected arguments
	for _, call := range calls {
		if m.argsMatch(call.Args, expectedArgs) {
			return // Found a matching call
		}
	}

	m.t.Errorf("Expected method %s to be called with args %v, but no matching call found", method, expectedArgs)
}

// AssertNotCalled verifies that a method was not called
func (m *MockBase) AssertNotCalled(method string) {
	if m.t == nil {
		return
	}
	m.t.Helper()

	count := m.CallCount(method)
	if count > 0 {
		m.t.Errorf("Expected method %s to not be called, but it was called %d times", method, count)
	}
}

// GetLastCall returns the most recent call for a method
func (m *MockBase) GetLastCall(method string) *CallRecord {
	calls := m.GetCallsForMethod(method)
	if len(calls) == 0 {
		return nil
	}
	return &calls[len(calls)-1]
}

// Helper method to compare arguments
func (m *MockBase) argsMatch(actual, expected []any) bool {
	if len(actual) != len(expected) {
		return false
	}

	for i, actualArg := range actual {
		expectedArg := expected[i]
		if !m.argEquals(actualArg, expectedArg) {
			return false
		}
	}

	return true
}

// Helper method to compare individual arguments
func (m *MockBase) argEquals(actual, expected any) bool {
	// For now, use simple equality. This could be enhanced with deep equality
	return fmt.Sprintf("%v", actual) == fmt.Sprintf("%v", expected)
}

// MockStore provides a generic in-memory store mock for testing
type MockStore struct {
	*MockBase

	data map[string]any // key -> value storage
}

// NewMockStore creates a new generic mock store
func NewMockStore(t *testing.T) *MockStore {
	return &MockStore{
		MockBase: NewMockBase(t),
		data:     make(map[string]any),
	}
}

// Set stores a value in the mock store
func (m *MockStore) Set(key string, value any) error {
	if err := m.ShouldReturnError("Set"); err != nil {
		m.RecordCall("Set", []any{key, value}, nil, err)
		return err
	}

	m.data[key] = value
	m.RecordCall("Set", []any{key, value}, []any{}, nil)
	return nil
}

// Get retrieves a value from the mock store
func (m *MockStore) Get(key string) (any, error) {
	if err := m.ShouldReturnError("Get"); err != nil {
		m.RecordCall("Get", []any{key}, nil, err)
		return nil, err
	}

	value, exists := m.data[key]
	if !exists {
		err := fmt.Errorf("key %s not found", key) //nolint:err113 // Dynamic errors are appropriate for test mocks
		m.RecordCall("Get", []any{key}, nil, err)
		return nil, err
	}

	m.RecordCall("Get", []any{key}, []any{value}, nil)
	return value, nil
}

// Delete removes a value from the mock store
func (m *MockStore) Delete(key string) error {
	if err := m.ShouldReturnError("Delete"); err != nil {
		m.RecordCall("Delete", []any{key}, nil, err)
		return err
	}

	delete(m.data, key)
	m.RecordCall("Delete", []any{key}, []any{}, nil)
	return nil
}

// Exists checks if a key exists in the mock store
func (m *MockStore) Exists(key string) bool {
	_, exists := m.data[key]
	m.RecordCall("Exists", []any{key}, []any{exists}, nil)
	return exists
}

// Count returns the number of items in the store
func (m *MockStore) Count() int {
	count := len(m.data)
	m.RecordCall("Count", []any{}, []any{count}, nil)
	return count
}

// Clear removes all items from the store
func (m *MockStore) Clear() {
	m.data = make(map[string]any)
	m.RecordCall("Clear", []any{}, []any{}, nil)
}

// MockValidator provides a generic validator mock for testing
type MockValidator struct {
	*MockBase

	validationRules map[string]func(any) error
}

// NewMockValidator creates a new generic mock validator
func NewMockValidator(t *testing.T) *MockValidator {
	return &MockValidator{
		MockBase:        NewMockBase(t),
		validationRules: make(map[string]func(any) error),
	}
}

// SetValidationRule sets a validation rule for a specific field
func (m *MockValidator) SetValidationRule(field string, rule func(any) error) {
	m.validationRules[field] = rule
}

// Validate validates a value using configured rules
func (m *MockValidator) Validate(field string, value any) error {
	if err := m.ShouldReturnError("Validate"); err != nil {
		m.RecordCall("Validate", []any{field, value}, nil, err)
		return err
	}

	if rule, exists := m.validationRules[field]; exists {
		if err := rule(value); err != nil {
			m.RecordCall("Validate", []any{field, value}, nil, err)
			return err
		}
	}

	m.RecordCall("Validate", []any{field, value}, []any{}, nil)
	return nil
}

// MockHandler provides a generic handler mock for testing
type MockHandler struct {
	*MockBase

	handlers map[string]func(...any) (any, error)
}

// NewMockHandler creates a new generic mock handler
func NewMockHandler(t *testing.T) *MockHandler {
	return &MockHandler{
		MockBase: NewMockBase(t),
		handlers: make(map[string]func(...any) (any, error)),
	}
}

// SetHandler sets a handler function for a specific method
func (m *MockHandler) SetHandler(method string, handler func(...any) (any, error)) {
	m.handlers[method] = handler
}

// Handle executes a handler with the given arguments
func (m *MockHandler) Handle(method string, args ...any) (any, error) {
	if err := m.ShouldReturnError(method); err != nil {
		m.RecordCall(method, args, nil, err)
		return nil, err
	}

	if handler, exists := m.handlers[method]; exists {
		result, err := handler(args...)
		m.RecordCall(method, args, []any{result}, err)
		return result, err
	}

	// Default behavior: return nil, nil
	m.RecordCall(method, args, []any{nil}, nil)
	return nil, nil //nolint:nilnil // Returning nil,nil is appropriate for mock handlers with no configured behavior
}
