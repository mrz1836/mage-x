package config

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

// Static errors to comply with err113 linter
var (
	errFieldCannotEmpty      = errors.New("field cannot be empty")
	errValidationAlwaysFails = errors.New("validation always fails")
	errMinLength3            = errors.New("minimum length is 3")
	errMustBePositive        = errors.New("must be positive")
	errMockValidationFailure = errors.New("mock validation failure")
	errMockSourceLoadFailure = errors.New("mock source load failure")
	errMockFieldValidation   = errors.New("mock field validation failure")
	errTooShort              = errors.New("too short")
)

// ManagerTestSuite tests the DefaultConfigManager and BasicValidator
type ManagerTestSuite struct {
	suite.Suite

	manager   *DefaultConfigManager
	validator *BasicValidator
}

func (s *ManagerTestSuite) SetupTest() {
	s.manager = NewDefaultConfigManager()
	s.validator = NewBasicValidator()
}

func (s *ManagerTestSuite) setupFreshManager() *DefaultConfigManager {
	return NewDefaultConfigManager()
}

func (s *ManagerTestSuite) TearDownTest() {
	// Stop any watching that might be active
	s.manager.StopWatching()
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}

// TestStopWatching tests the StopWatching method
func (s *ManagerTestSuite) TestStopWatching() {
	s.Run("StopWatchingWithoutWatching", func() {
		// Should not panic or error when called without watching
		s.NotPanics(func() {
			s.manager.StopWatching()
		})
	})

	s.Run("StopWatchingWithActiveWatching", func() {
		// Start watching first
		callback := func(interface{}) {}
		err := s.manager.Watch(callback)
		s.Require().NoError(err)

		// Verify watching is active
		s.True(s.manager.IsWatching(), "Manager should be watching")

		// Stop watching
		s.manager.StopWatching()

		// Verify watching is stopped (no need for sleep as StopWatching is synchronous now)
		s.False(s.manager.IsWatching(), "Manager should not be watching after stop")
	})

	s.Run("MultipleStopWatchingCalls", func() {
		// Start watching
		callback := func(interface{}) {}
		err := s.manager.Watch(callback)
		s.Require().NoError(err)

		// Call stop multiple times - should not panic
		s.NotPanics(func() {
			s.manager.StopWatching()
			s.manager.StopWatching()
			s.manager.StopWatching()
		})
	})
}

// TestGetActiveSources tests the GetActiveSources method
func (s *ManagerTestSuite) TestGetActiveSources() {
	s.Run("NoSources", func() {
		sources := s.manager.GetActiveSources()
		s.Empty(sources, "Should return empty slice when no sources added")
	})

	s.Run("AllSourcesAvailable", func() {
		// Create mock sources
		source1 := &mockSource{name: "source1", priority: 100, available: true}
		source2 := &mockSource{name: "source2", priority: 200, available: true}
		source3 := &mockSource{name: "source3", priority: 50, available: true}

		s.manager.AddSource(source1)
		s.manager.AddSource(source2)
		s.manager.AddSource(source3)

		sources := s.manager.GetActiveSources()
		s.Len(sources, 3, "Should return all available sources")

		// Verify all sources are present
		sourceNames := make([]string, len(sources))
		for i, source := range sources {
			sourceNames[i] = source.Name()
		}
		s.ElementsMatch([]string{"source1", "source2", "source3"}, sourceNames)
	})

	s.Run("MixedAvailability", func() {
		manager := s.setupFreshManager()

		source1 := &mockSource{name: "available1", priority: 100, available: true}
		source2 := &mockSource{name: "unavailable1", priority: 200, available: false}
		source3 := &mockSource{name: "available2", priority: 50, available: true}
		source4 := &mockSource{name: "unavailable2", priority: 150, available: false}

		manager.AddSource(source1)
		manager.AddSource(source2)
		manager.AddSource(source3)
		manager.AddSource(source4)

		sources := manager.GetActiveSources()
		s.Len(sources, 2, "Should return only available sources")

		sourceNames := make([]string, len(sources))
		for i, source := range sources {
			sourceNames[i] = source.Name()
		}
		s.ElementsMatch([]string{"available1", "available2"}, sourceNames)
	})

	s.Run("NoAvailableSources", func() {
		manager := s.setupFreshManager()

		source1 := &mockSource{name: "unavailable1", priority: 100, available: false}
		source2 := &mockSource{name: "unavailable2", priority: 200, available: false}

		manager.AddSource(source1)
		manager.AddSource(source2)

		sources := manager.GetActiveSources()
		s.Empty(sources, "Should return empty slice when no sources are available")
	})

	s.Run("ConcurrentAccess", func() {
		manager := s.setupFreshManager()

		// Test concurrent access to GetActiveSources
		var wg sync.WaitGroup
		const numGoroutines = 10

		// Add some sources
		for i := 0; i < 5; i++ {
			source := &mockSource{
				name:      "concurrent_source_" + string(rune('A'+i)),
				priority:  100 + i*10,
				available: i%2 == 0, // Every other source is available
			}
			manager.AddSource(source)
		}

		// Run concurrent GetActiveSources calls
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sources := manager.GetActiveSources()
				s.NotNil(sources, "Should return valid slice")
			}()
		}

		wg.Wait()
	})
}

// TestSetValidator tests the SetValidator method
func (s *ManagerTestSuite) TestSetValidator() {
	s.Run("SetNilValidator", func() {
		s.manager.SetValidator(nil)

		// Try to load config - should succeed without validation
		source := &mockSource{name: "test", priority: 100, available: true}
		s.manager.AddSource(source)

		var config map[string]interface{}
		err := s.manager.LoadConfig(&config)
		s.NoError(err, "Should succeed without validator")
	})

	s.Run("SetValidValidator", func() {
		validator := NewBasicValidator()
		s.manager.SetValidator(validator)

		// Access validator field using reflection or by testing behavior
		// We test behavior by loading config that will trigger validation
		source := &mockSource{name: "test", priority: 100, available: true}
		s.manager.AddSource(source)

		var config map[string]interface{}
		err := s.manager.LoadConfig(&config)
		s.NoError(err, "Should succeed with valid validator")
	})

	s.Run("ConcurrentSetValidator", func() {
		var wg sync.WaitGroup
		const numGoroutines = 10

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				validator := NewBasicValidator()
				s.NotPanics(func() {
					s.manager.SetValidator(validator)
				})
			}()
		}

		wg.Wait()
	})
}

// TestNewBasicValidator tests the NewBasicValidator constructor
func (s *ManagerTestSuite) TestNewBasicValidator() {
	s.Run("CreateValidator", func() {
		validator := NewBasicValidator()
		s.NotNil(validator, "Should create non-nil validator")
		s.NotNil(validator.rules, "Should initialize rules map")
		s.Empty(validator.rules, "Should start with empty rules")
	})

	s.Run("MultipleValidators", func() {
		validator1 := NewBasicValidator()
		validator2 := NewBasicValidator()

		s.NotNil(validator1)
		s.NotNil(validator2)
		s.NotSame(validator1, validator2, "Should create separate instances")
	})
}

// TestBasicValidatorValidate tests the BasicValidator.Validate method
func (s *ManagerTestSuite) TestBasicValidatorValidate() {
	s.Run("ValidateNilData", func() {
		err := s.validator.Validate(nil)
		s.Require().Error(err, "Should error on nil data")
		s.Equal(errConfigDataCannotNil, err, "Should return specific error")
	})

	s.Run("ValidateValidData", func() {
		testCases := []struct {
			name string
			data interface{}
		}{
			{"String data", "test string"},
			{"Integer data", 42},
			{"Boolean data", true},
			{"Map data", map[string]interface{}{"key": "value"}},
			{"Slice data", []string{"a", "b", "c"}},
			{"Struct data", struct{ Name string }{Name: "test"}},
			{"Empty map", map[string]interface{}{}},
			{"Empty slice", []string{}},
		}

		for _, tc := range testCases {
			s.Run(tc.name, func() {
				err := s.validator.Validate(tc.data)
				s.NoError(err, "Should not error on valid data: %v", tc.data)
			})
		}
	})

	s.Run("ValidateMapPointerData", func() {
		// Test with valid *map[string]interface{} pointer
		data := map[string]interface{}{"key": "value"}
		err := s.validator.Validate(&data)
		s.NoError(err, "Should not error on valid map pointer data")
	})

	s.Run("ValidateNilMapPointerData", func() {
		// Test with nil *map[string]interface{} pointer
		var nilMapPtr *map[string]interface{}
		err := s.validator.Validate(nilMapPtr)
		s.Require().Error(err, "Should error on nil map pointer")
		s.Equal(errConfigDataCannotNil, err, "Should return specific error for nil map pointer")
	})

	s.Run("ValidateMapWithFailingFieldValidation", func() {
		// Set up a validation rule that will fail
		validator := NewBasicValidator()
		validator.rules["badField"] = func(value interface{}) error {
			return errMockFieldValidation
		}

		// Create data with a field that will fail validation
		data := map[string]interface{}{
			"badField": "some value",
		}

		err := validator.Validate(data)
		s.Require().Error(err, "Should error when field validation fails")
		s.Contains(err.Error(), "validation failed for field 'badField'")
	})

	s.Run("ValidateMapPointerWithFailingFieldValidation", func() {
		// Set up a validation rule that will fail
		validator := NewBasicValidator()
		validator.rules["invalidField"] = func(value interface{}) error {
			return errMockValidationFailure
		}

		// Create data with a field that will fail validation
		data := map[string]interface{}{
			"invalidField": "test",
		}

		err := validator.Validate(&data)
		s.Require().Error(err, "Should error when field validation fails via pointer")
		s.Contains(err.Error(), "validation failed for field 'invalidField'")
	})
}

// TestBasicValidatorValidateField tests the BasicValidator.ValidateField method
func (s *ManagerTestSuite) TestBasicValidatorValidateField() {
	s.Run("ValidateFieldWithoutRule", func() {
		err := s.validator.ValidateField("nonexistent", "value")
		s.NoError(err, "Should not error when no rule exists for field")
	})

	s.Run("ValidateFieldWithFunctionRule", func() {
		// Add a validation rule that checks if value is not empty
		rule := func(value interface{}) error {
			if str, ok := value.(string); ok && str == "" {
				return errFieldCannotEmpty
			}
			return nil
		}
		s.validator.rules["testField"] = rule

		// Test with valid value
		err := s.validator.ValidateField("testField", "valid value")
		s.Require().NoError(err, "Should pass validation with valid value")

		// Test with invalid value
		err = s.validator.ValidateField("testField", "")
		s.Require().Error(err, "Should fail validation with empty value")
		s.Contains(err.Error(), "cannot be empty")
	})

	s.Run("ValidateFieldWithNonFunctionRule", func() {
		// Add a non-function rule
		s.validator.rules["testField"] = "not a function"

		err := s.validator.ValidateField("testField", "value")
		s.NoError(err, "Should not error with non-function rule")
	})

	s.Run("ValidateFieldWithErroringRule", func() {
		rule := func(value interface{}) error {
			return errValidationAlwaysFails
		}
		s.validator.rules["failField"] = rule

		err := s.validator.ValidateField("failField", "any value")
		s.Require().Error(err, "Should return error from validation rule")
		s.Contains(err.Error(), "validation always fails")
	})

	s.Run("ValidateFieldTableDriven", func() {
		// Set up multiple validation rules
		rules := map[string]interface{}{
			"minLength": func(value interface{}) error {
				if str, ok := value.(string); ok && len(str) < 3 {
					return errMinLength3
				}
				return nil
			},
			"positiveNumber": func(value interface{}) error {
				if num, ok := value.(int); ok && num <= 0 {
					return errMustBePositive
				}
				return nil
			},
			"nonFunction": "this is not a function",
		}

		for field, rule := range rules {
			s.validator.rules[field] = rule
		}

		testCases := []struct {
			field       string
			value       interface{}
			expectError bool
			errorText   string
		}{
			{"minLength", "hello", false, ""},
			{"minLength", "hi", true, "minimum length is 3"},
			{"minLength", "", true, "minimum length is 3"},
			{"positiveNumber", 5, false, ""},
			{"positiveNumber", -1, true, "must be positive"},
			{"positiveNumber", 0, true, "must be positive"},
			{"nonFunction", "any value", false, ""},
			{"nonExistentField", "any value", false, ""},
		}

		for _, tc := range testCases {
			s.Run(fmt.Sprintf("%s_%s", tc.field, tc.errorText), func() {
				err := s.validator.ValidateField(tc.field, tc.value)
				if tc.expectError {
					s.Require().Error(err, "Expected error for field %s with value %v", tc.field, tc.value)
					if tc.errorText != "" {
						s.Contains(err.Error(), tc.errorText)
					}
				} else {
					s.NoError(err, "Expected no error for field %s with value %v", tc.field, tc.value)
				}
			})
		}
	})
}

// TestBasicValidatorGetValidationRules tests the BasicValidator.GetValidationRules method
func (s *ManagerTestSuite) TestBasicValidatorGetValidationRules() {
	s.Run("GetEmptyRules", func() {
		rules := s.validator.GetValidationRules()
		s.NotNil(rules, "Should return non-nil map")
		s.Empty(rules, "Should return empty map when no rules set")
	})

	s.Run("GetRulesWithData", func() {
		// Add some rules
		testRules := map[string]interface{}{
			"rule1": func(interface{}) error { return nil },
			"rule2": "string rule",
			"rule3": 42,
		}

		for k, v := range testRules {
			s.validator.rules[k] = v
		}

		rules := s.validator.GetValidationRules()
		s.Len(rules, 3, "Should return all rules")
		s.Contains(rules, "rule1")
		s.Contains(rules, "rule2")
		s.Contains(rules, "rule3")
		s.Equal("string rule", rules["rule2"])
		s.Equal(42, rules["rule3"])
	})

	s.Run("GetRulesReturnsCopy", func() {
		// Add a rule
		s.validator.rules["original"] = "value"

		// Get rules and modify the returned map
		rules := s.validator.GetValidationRules()
		rules["new"] = "should not affect original"
		delete(rules, "original")

		// Verify original rules are unchanged
		originalRules := s.validator.GetValidationRules()
		s.Contains(originalRules, "original", "Original rules should be unchanged")
		s.NotContains(originalRules, "new", "Original rules should not contain modifications")
	})
}

// TestBasicValidatorSetValidationRules tests the BasicValidator.SetValidationRules method
func (s *ManagerTestSuite) TestBasicValidatorSetValidationRules() {
	s.Run("SetEmptyRules", func() {
		// First add some rules
		s.validator.rules["existing"] = "value"

		// Set empty rules
		s.validator.SetValidationRules(map[string]interface{}{})

		rules := s.validator.GetValidationRules()
		s.Empty(rules, "Should clear all existing rules")
	})

	s.Run("SetNilRules", func() {
		// First add some rules
		s.validator.rules["existing"] = "value"

		// Set nil rules
		s.validator.SetValidationRules(nil)

		rules := s.validator.GetValidationRules()
		s.Empty(rules, "Should clear all existing rules when nil is passed")
	})

	s.Run("SetValidRules", func() {
		newRules := map[string]interface{}{
			"rule1": func(interface{}) error { return nil },
			"rule2": "string validation",
			"rule3": 123,
		}

		s.validator.SetValidationRules(newRules)

		rules := s.validator.GetValidationRules()
		s.Len(rules, 3, "Should set all provided rules")
		s.Equal("string validation", rules["rule2"])
		s.Equal(123, rules["rule3"])
	})

	s.Run("SetRulesOverwritesExisting", func() {
		// Set initial rules
		s.validator.rules["old1"] = "old value 1"
		s.validator.rules["old2"] = "old value 2"

		// Set new rules
		newRules := map[string]interface{}{
			"new1": "new value 1",
			"new2": "new value 2",
		}
		s.validator.SetValidationRules(newRules)

		rules := s.validator.GetValidationRules()
		s.Len(rules, 2, "Should only contain new rules")
		s.Contains(rules, "new1")
		s.Contains(rules, "new2")
		s.NotContains(rules, "old1", "Should not contain old rules")
		s.NotContains(rules, "old2", "Should not contain old rules")
	})

	s.Run("SetRulesCreatesCopy", func() {
		originalRules := map[string]interface{}{
			"rule1": "value1",
			"rule2": "value2",
		}

		s.validator.SetValidationRules(originalRules)

		// Modify the original map
		originalRules["rule3"] = "value3"
		delete(originalRules, "rule1")

		// Verify validator rules are unchanged
		rules := s.validator.GetValidationRules()
		s.Len(rules, 2, "Should not be affected by modifications to original map")
		s.Contains(rules, "rule1", "Should still contain rule1")
		s.NotContains(rules, "rule3", "Should not contain new rule added to original")
	})
}

// TestLoadConfigEdgeCases tests additional edge cases for LoadConfig method
func (s *ManagerTestSuite) TestLoadConfigEdgeCases() {
	s.Run("LoadConfigWithNoSources", func() {
		manager := s.setupFreshManager()

		var config map[string]interface{}
		err := manager.LoadConfig(&config)
		s.Require().Error(err, "Should fail when no sources are added")
		s.Equal(errNoConfigSources, err, "Should return specific error for no sources")
	})

	s.Run("LoadConfigWithUnavailableSources", func() {
		manager := s.setupFreshManager()

		// Add sources that are all unavailable
		source1 := &mockSource{name: "unavailable1", priority: 100, available: false}
		source2 := &mockSource{name: "unavailable2", priority: 200, available: false}

		manager.AddSource(source1)
		manager.AddSource(source2)

		var config map[string]interface{}
		err := manager.LoadConfig(&config)
		s.Require().Error(err, "Should fail when all sources are unavailable")
		s.Equal(errNoConfigSources, err, "Should return no sources error when all are unavailable")
	})

	s.Run("LoadConfigWithValidatorFailure", func() {
		// Create a validator that always fails
		failingValidator := NewBasicValidator()
		failingValidator.SetValidationRules(map[string]interface{}{
			"test": func(interface{}) error {
				return errMockValidationFailure
			},
		})

		// Override the validator's Validate method behavior by setting it
		s.manager.SetValidator(&alwaysFailValidator{})

		// Add a source that will succeed
		source := &mockSource{name: "test", priority: 100, available: true}
		s.manager.AddSource(source)

		var config map[string]interface{}
		err := s.manager.LoadConfig(&config)
		s.Require().Error(err, "Should fail when validator fails")
		s.Contains(err.Error(), "configuration validation failed")
	})

	s.Run("LoadConfigWithMultipleFailingSources", func() {
		manager := s.setupFreshManager()

		// Add multiple sources that will fail to load
		source1 := &mockSource{name: "fail1", priority: 300, available: true, shouldFail: true}
		source2 := &mockSource{name: "fail2", priority: 200, available: true, shouldFail: true}
		source3 := &mockSource{name: "fail3", priority: 100, available: true, shouldFail: true}

		manager.AddSource(source1)
		manager.AddSource(source2)
		manager.AddSource(source3)

		var config map[string]interface{}
		err := manager.LoadConfig(&config)
		s.Require().Error(err, "Should fail when all sources fail")
		s.Contains(err.Error(), "failed to load configuration from any source")
	})

	s.Run("LoadConfigConcurrentAccess", func() {
		manager := s.setupFreshManager()

		// Add some sources
		for i := 0; i < 3; i++ {
			source := &mockSource{
				name:      "concurrent_" + string(rune('A'+i)),
				priority:  100 + i*10,
				available: true,
			}
			manager.AddSource(source)
		}

		var wg sync.WaitGroup
		const numGoroutines = 10
		errors := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				var config map[string]interface{}
				err := manager.LoadConfig(&config)
				errors <- err
			}()
		}

		wg.Wait()
		close(errors)

		// Check that all operations succeeded
		for err := range errors {
			s.NoError(err, "Concurrent LoadConfig should succeed")
		}
	})
}

// TestWatchStopWatchingLifecycle tests the complete lifecycle of watching
func (s *ManagerTestSuite) TestWatchStopWatchingLifecycle() {
	s.Run("WatchStopWatchingSequence", func() {
		// Start watching
		callback := func(interface{}) {
			// Callback implementation for testing
		}

		err := s.manager.Watch(callback)
		s.Require().NoError(err, "Watch should succeed")

		// Verify watching state
		s.True(s.manager.IsWatching(), "Should be watching")

		// Stop watching
		s.manager.StopWatching()

		// Verify watching state (no sleep needed - StopWatching is synchronous)
		s.False(s.manager.IsWatching(), "Should not be watching after stop")

		// Should be able to watch again
		err = s.manager.Watch(callback)
		s.NoError(err, "Should be able to watch again after stopping")
	})

	s.Run("MultipleWatchCallsError", func() {
		manager := s.setupFreshManager()
		callback := func(interface{}) {}

		err := manager.Watch(callback)
		s.Require().NoError(err, "First watch should succeed")

		err = manager.Watch(callback)
		s.Require().Error(err, "Second watch should fail")
		s.Equal(errAlreadyWatching, err, "Should return specific error")

		// Clean up
		manager.StopWatching()
	})
}

// Mock source implementation for testing
type mockSource struct {
	name       string
	priority   int
	available  bool
	shouldFail bool
	loadData   map[string]interface{}
}

func (m *mockSource) Name() string {
	return m.name
}

func (m *mockSource) Priority() int {
	return m.priority
}

func (m *mockSource) IsAvailable() bool {
	return m.available
}

func (m *mockSource) Load(dest interface{}) error {
	if m.shouldFail {
		return errMockSourceLoadFailure
	}

	// Simple mock loading - just set some data
	if m.loadData != nil {
		if mapDest, ok := dest.(*map[string]interface{}); ok {
			*mapDest = make(map[string]interface{})
			for k, v := range m.loadData {
				(*mapDest)[k] = v
			}
		}
	}

	return nil
}

// Mock validator that always fails
type alwaysFailValidator struct{}

func (a *alwaysFailValidator) Validate(data interface{}) error {
	return errMockValidationFailure
}

func (a *alwaysFailValidator) ValidateField(fieldName string, value interface{}) error {
	return errMockFieldValidation
}

func (a *alwaysFailValidator) GetValidationRules() map[string]interface{} {
	return map[string]interface{}{}
}

func (a *alwaysFailValidator) SetValidationRules(rules map[string]interface{}) {
	// No-op for mock
}

// Benchmark tests for performance validation
func BenchmarkGetActiveSources(b *testing.B) {
	manager := NewDefaultConfigManager()

	// Add many sources
	for i := 0; i < 100; i++ {
		source := &mockSource{
			name:      "bench_source_" + string(rune('A'+i%26)),
			priority:  i,
			available: i%3 != 0, // Most sources are available
		}
		manager.AddSource(source)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetActiveSources()
	}
}

func BenchmarkBasicValidatorValidateField(b *testing.B) {
	validator := NewBasicValidator()

	// Set up validation rules
	rules := make(map[string]interface{})
	for i := 0; i < 10; i++ {
		fieldName := "field" + string(rune('A'+i))
		rules[fieldName] = func(value interface{}) error {
			if str, ok := value.(string); ok && len(str) < 3 {
				return errTooShort
			}
			return nil
		}
	}
	validator.SetValidationRules(rules)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		field := "field" + string(rune('A'+i%10))
		err := validator.ValidateField(field, "test_value")
		if err != nil {
			b.Errorf("Unexpected error: %v", err)
		}
	}
}

func BenchmarkSetValidatorConcurrent(b *testing.B) {
	manager := NewDefaultConfigManager()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			validator := NewBasicValidator()
			manager.SetValidator(validator)
		}
	})
}
