package testhelpers

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Static test errors to satisfy err113 linter
var (
	errSpecificTestError = errors.New("specific test error")
	errMockTestError     = errors.New("test error")
	errSetError          = errors.New("set error")
	errFieldRequired     = errors.New("field is required")
	errValidationSystem  = errors.New("validation system error")
	errNoArgsProvided    = errors.New("no arguments provided")
	errHandlerError      = errors.New("handler error")
	errInvalidArgType    = errors.New("invalid argument type")
	errCustomError       = errors.New("custom error")
)

func TestMockBase(t *testing.T) {
	t.Run("NewMockBase", func(t *testing.T) {
		mock := NewMockBase(t)

		require.NotNil(t, mock)
		assert.Equal(t, t, mock.t)
		assert.Empty(t, mock.calls)
		assert.Empty(t, mock.errorMap)
		assert.Empty(t, mock.callCount)
		assert.False(t, mock.shouldError)
	})

	t.Run("SetError and ShouldReturnError", func(t *testing.T) {
		mock := NewMockBase(t)

		// Initially should not error
		err := mock.ShouldReturnError("TestMethod")
		require.NoError(t, err)

		// Set global error flag
		mock.SetError(true)
		err = mock.ShouldReturnError("TestMethod")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "mock error for TestMethod")

		// Clear global error flag
		mock.SetError(false)
		err = mock.ShouldReturnError("TestMethod")
		assert.NoError(t, err)
	})

	t.Run("SetMethodError and ShouldReturnError", func(t *testing.T) {
		mock := NewMockBase(t)
		testErr := errSpecificTestError

		// Set method-specific error
		mock.SetMethodError("SpecificMethod", testErr)

		// Should return specific error for that method
		err := mock.ShouldReturnError("SpecificMethod")
		assert.Equal(t, testErr, err)

		// Should not error for other methods
		err = mock.ShouldReturnError("OtherMethod")
		require.NoError(t, err)

		// Clear method error
		mock.ClearMethodError("SpecificMethod")
		err = mock.ShouldReturnError("SpecificMethod")
		assert.NoError(t, err)
	})

	t.Run("ClearAllErrors", func(t *testing.T) {
		mock := NewMockBase(t)
		testErr := errMockTestError

		// Set both global and method-specific errors
		mock.SetError(true)
		mock.SetMethodError("TestMethod", testErr)

		// Verify errors are set
		err := mock.ShouldReturnError("TestMethod")
		assert.Equal(t, testErr, err) // Method-specific takes precedence

		err = mock.ShouldReturnError("OtherMethod")
		require.Error(t, err) // Global error

		// Clear all errors
		mock.ClearAllErrors()

		// Verify all errors are cleared
		err = mock.ShouldReturnError("TestMethod")
		require.NoError(t, err)

		err = mock.ShouldReturnError("OtherMethod")
		assert.NoError(t, err)
	})

	t.Run("RecordCall and GetCalls", func(t *testing.T) {
		mock := NewMockBase(t)

		// Record some calls
		args1 := []interface{}{"arg1", 42}
		result1 := []interface{}{"result1"}
		mock.RecordCall("Method1", args1, result1, nil)

		args2 := []interface{}{"arg2"}
		testErr := errMockTestError
		mock.RecordCall("Method2", args2, nil, testErr)

		// Get all calls
		calls := mock.GetCalls()
		require.Len(t, calls, 2)

		// Verify first call
		assert.Equal(t, "Method1", calls[0].Method)
		assert.Equal(t, args1, calls[0].Args)
		assert.Equal(t, result1, calls[0].Result)
		require.NoError(t, calls[0].Error)
		assert.WithinDuration(t, time.Now(), calls[0].Timestamp, time.Second)

		// Verify second call
		assert.Equal(t, "Method2", calls[1].Method)
		assert.Equal(t, args2, calls[1].Args)
		assert.Nil(t, calls[1].Result)
		assert.Equal(t, testErr, calls[1].Error)
	})

	t.Run("GetCallsForMethod", func(t *testing.T) {
		mock := NewMockBase(t)

		// Record calls for different methods
		mock.RecordCall("Method1", []interface{}{"call1"}, nil, nil)
		mock.RecordCall("Method2", []interface{}{"call2"}, nil, nil)
		mock.RecordCall("Method1", []interface{}{"call3"}, nil, nil)

		// Get calls for specific method
		method1Calls := mock.GetCallsForMethod("Method1")
		require.Len(t, method1Calls, 2)
		assert.Equal(t, "call1", method1Calls[0].Args[0])
		assert.Equal(t, "call3", method1Calls[1].Args[0])

		method2Calls := mock.GetCallsForMethod("Method2")
		require.Len(t, method2Calls, 1)
		assert.Equal(t, "call2", method2Calls[0].Args[0])

		// Non-existent method
		nonExistentCalls := mock.GetCallsForMethod("NonExistent")
		assert.Empty(t, nonExistentCalls)
	})

	t.Run("CallCount and TotalCalls", func(t *testing.T) {
		mock := NewMockBase(t)

		// Initially no calls
		assert.Equal(t, 0, mock.CallCount("Method1"))
		assert.Equal(t, 0, mock.TotalCalls())

		// Record some calls
		mock.RecordCall("Method1", nil, nil, nil)
		mock.RecordCall("Method1", nil, nil, nil)
		mock.RecordCall("Method2", nil, nil, nil)

		// Check counts
		assert.Equal(t, 2, mock.CallCount("Method1"))
		assert.Equal(t, 1, mock.CallCount("Method2"))
		assert.Equal(t, 0, mock.CallCount("NonExistent"))
		assert.Equal(t, 3, mock.TotalCalls())
	})

	t.Run("Reset", func(t *testing.T) {
		mock := NewMockBase(t)
		testErr := errMockTestError

		// Set up some state
		mock.SetError(true)
		mock.SetMethodError("TestMethod", testErr)
		mock.RecordCall("Method1", []interface{}{"arg"}, nil, nil)
		mock.RecordCall("Method2", []interface{}{"arg"}, nil, nil)

		// Verify state is set
		assert.True(t, mock.shouldError)
		assert.Len(t, mock.errorMap, 1)
		assert.Equal(t, 2, mock.TotalCalls())

		// Reset
		mock.Reset()

		// Verify state is cleared
		assert.False(t, mock.shouldError)
		assert.Empty(t, mock.errorMap)
		assert.Equal(t, 0, mock.TotalCalls())
		assert.Empty(t, mock.GetCalls())
	})

	t.Run("GetLastCall", func(t *testing.T) {
		mock := NewMockBase(t)

		// No calls initially
		lastCall := mock.GetLastCall("Method1")
		assert.Nil(t, lastCall)

		// Record some calls
		mock.RecordCall("Method1", []interface{}{"first"}, nil, nil)
		mock.RecordCall("Method2", []interface{}{"other"}, nil, nil)
		mock.RecordCall("Method1", []interface{}{"last"}, nil, nil)

		// Get last call for Method1
		lastCall = mock.GetLastCall("Method1")
		require.NotNil(t, lastCall)
		assert.Equal(t, "Method1", lastCall.Method)
		assert.Equal(t, "last", lastCall.Args[0])

		// Get last call for Method2
		lastCall = mock.GetLastCall("Method2")
		require.NotNil(t, lastCall)
		assert.Equal(t, "Method2", lastCall.Method)
		assert.Equal(t, "other", lastCall.Args[0])

		// Non-existent method
		lastCall = mock.GetLastCall("NonExistent")
		assert.Nil(t, lastCall)
	})
}

func TestMockBaseAssertions(t *testing.T) {
	t.Run("AssertCalled", func(t *testing.T) {
		mock := NewMockBase(t)

		// Method not called - would fail in real test
		// We can't easily test assertion failures without complex setup

		// Record a call
		mock.RecordCall("TestMethod", nil, nil, nil)

		// This should pass (no assertion failure)
		mock.AssertCalled("TestMethod")

		// Verify call count increased
		assert.Equal(t, 1, mock.CallCount("TestMethod"))
	})

	t.Run("AssertCalledTimes", func(t *testing.T) {
		mock := NewMockBase(t)

		// Record some calls
		mock.RecordCall("TestMethod", nil, nil, nil)
		mock.RecordCall("TestMethod", nil, nil, nil)

		// This should pass
		mock.AssertCalledTimes("TestMethod", 2)

		// Verify call count
		assert.Equal(t, 2, mock.CallCount("TestMethod"))
	})

	t.Run("AssertCalledWith", func(t *testing.T) {
		mock := NewMockBase(t)

		// Record calls with different arguments
		mock.RecordCall("TestMethod", []interface{}{"arg1", 42}, nil, nil)
		mock.RecordCall("TestMethod", []interface{}{"arg2", 24}, nil, nil)

		// This should pass
		mock.AssertCalledWith("TestMethod", "arg1", 42)
		mock.AssertCalledWith("TestMethod", "arg2", 24)
	})

	t.Run("AssertNotCalled", func(t *testing.T) {
		mock := NewMockBase(t)

		// This should pass
		mock.AssertNotCalled("UnusedMethod")

		// Record a call
		mock.RecordCall("TestMethod", nil, nil, nil)

		// This should still pass for a different method
		mock.AssertNotCalled("UnusedMethod")
	})

	t.Run("argsMatch", func(t *testing.T) {
		mock := NewMockBase(t)

		// Test argument matching
		assert.True(t, mock.argsMatch([]interface{}{"a", 1}, []interface{}{"a", 1}))
		assert.False(t, mock.argsMatch([]interface{}{"a", 1}, []interface{}{"b", 1}))
		assert.False(t, mock.argsMatch([]interface{}{"a"}, []interface{}{"a", 1}))
		assert.True(t, mock.argsMatch([]interface{}{}, []interface{}{}))
	})

	t.Run("argEquals", func(t *testing.T) {
		mock := NewMockBase(t)

		// Test individual argument equality
		assert.True(t, mock.argEquals("test", "test"))
		assert.True(t, mock.argEquals(42, 42))
		assert.False(t, mock.argEquals("test", "different"))
		assert.False(t, mock.argEquals(42, 24))
	})
}

func TestMockStore(t *testing.T) {
	t.Run("NewMockStore", func(t *testing.T) {
		store := NewMockStore(t)

		require.NotNil(t, store)
		require.NotNil(t, store.MockBase)
		assert.NotNil(t, store.data)
		assert.Empty(t, store.data)
	})

	t.Run("Set and Get", func(t *testing.T) {
		store := NewMockStore(t)

		// Set a value
		err := store.Set("key1", "value1")
		require.NoError(t, err)

		// Get the value
		value, err := store.Get("key1")
		require.NoError(t, err)
		assert.Equal(t, "value1", value)

		// Verify call recording
		assert.Equal(t, 1, store.CallCount("Set"))
		assert.Equal(t, 1, store.CallCount("Get"))
	})

	t.Run("Get non-existent key", func(t *testing.T) {
		store := NewMockStore(t)

		// Get non-existent key
		value, err := store.Get("nonexistent")
		require.Error(t, err)
		assert.Nil(t, value)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("Delete", func(t *testing.T) {
		store := NewMockStore(t)

		// Set and then delete
		err := store.Set("key1", "value1")
		require.NoError(t, err)

		err = store.Delete("key1")
		require.NoError(t, err)

		// Verify deletion
		value, err := store.Get("key1")
		require.Error(t, err)
		assert.Nil(t, value)

		// Verify call counts
		assert.Equal(t, 1, store.CallCount("Set"))
		assert.Equal(t, 1, store.CallCount("Delete"))
		assert.Equal(t, 1, store.CallCount("Get"))
	})

	t.Run("Exists", func(t *testing.T) {
		store := NewMockStore(t)

		// Initially doesn't exist
		exists := store.Exists("key1")
		assert.False(t, exists)

		// Set a value
		err := store.Set("key1", "value1")
		require.NoError(t, err)

		// Now exists
		exists = store.Exists("key1")
		assert.True(t, exists)
	})

	t.Run("Count", func(t *testing.T) {
		store := NewMockStore(t)

		// Initially empty
		count := store.Count()
		assert.Equal(t, 0, count)

		// Add some items
		err := store.Set("key1", "value1")
		require.NoError(t, err)
		err = store.Set("key2", "value2")
		require.NoError(t, err)

		// Check count
		count = store.Count()
		assert.Equal(t, 2, count)
	})

	t.Run("Clear", func(t *testing.T) {
		store := NewMockStore(t)

		// Add some items
		err := store.Set("key1", "value1")
		require.NoError(t, err)
		err = store.Set("key2", "value2")
		require.NoError(t, err)

		// Clear
		store.Clear()

		// Verify cleared
		assert.Equal(t, 0, store.Count())
		exists := store.Exists("key1")
		assert.False(t, exists)
	})

	t.Run("Error simulation", func(t *testing.T) {
		store := NewMockStore(t)

		// Set error for Set method
		testErr := errSetError
		store.SetMethodError("Set", testErr)

		// Should return error
		err := store.Set("key1", "value1")
		assert.Equal(t, testErr, err)

		// Clear error and try again
		store.ClearMethodError("Set")
		err = store.Set("key1", "value1")
		assert.NoError(t, err)
	})
}

func TestMockValidator(t *testing.T) {
	t.Run("NewMockValidator", func(t *testing.T) {
		validator := NewMockValidator(t)

		require.NotNil(t, validator)
		require.NotNil(t, validator.MockBase)
		assert.NotNil(t, validator.validationRules)
		assert.Empty(t, validator.validationRules)
	})

	t.Run("Validate without rules", func(t *testing.T) {
		validator := NewMockValidator(t)

		// Should pass without any rules
		err := validator.Validate("field1", "value1")
		require.NoError(t, err)

		// Verify call recording
		assert.Equal(t, 1, validator.CallCount("Validate"))
	})

	t.Run("Validate with rules", func(t *testing.T) {
		validator := NewMockValidator(t)

		// Set validation rule
		validator.SetValidationRule("required_field", func(value interface{}) error {
			if value == nil || value == "" {
				return errFieldRequired
			}
			return nil
		})

		// Test valid value
		err := validator.Validate("required_field", "valid_value")
		require.NoError(t, err)

		// Test invalid value
		err = validator.Validate("required_field", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "field is required")

		// Test field without rule
		err = validator.Validate("other_field", "anything")
		assert.NoError(t, err)
	})

	t.Run("Error simulation", func(t *testing.T) {
		validator := NewMockValidator(t)

		// Set global error
		testErr := errValidationSystem
		validator.SetMethodError("Validate", testErr)

		// Should return error
		err := validator.Validate("field1", "value1")
		assert.Equal(t, testErr, err)
	})
}

func TestMockHandler(t *testing.T) {
	t.Run("NewMockHandler", func(t *testing.T) {
		handler := NewMockHandler(t)

		require.NotNil(t, handler)
		require.NotNil(t, handler.MockBase)
		assert.NotNil(t, handler.handlers)
		assert.Empty(t, handler.handlers)
	})

	t.Run("Handle without handler function", func(t *testing.T) {
		handler := NewMockHandler(t)

		// Should return nil, nil for unknown methods
		result, err := handler.Handle("unknown_method", "arg1", "arg2")
		require.NoError(t, err)
		assert.Nil(t, result)

		// Verify call recording
		assert.Equal(t, 1, handler.CallCount("unknown_method"))
	})

	t.Run("Handle with handler function", func(t *testing.T) {
		handler := NewMockHandler(t)

		// Set handler function
		handler.SetHandler("process", func(args ...interface{}) (interface{}, error) {
			if len(args) == 0 {
				return nil, errNoArgsProvided
			}
			str, ok := args[0].(string)
			if !ok {
				return nil, errInvalidArgType
			}
			return "processed: " + str, nil
		})

		// Test with valid arguments
		result, err := handler.Handle("process", "test_data")
		require.NoError(t, err)
		assert.Equal(t, "processed: test_data", result)

		// Test with no arguments
		result, err = handler.Handle("process")
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "no arguments provided")
	})

	t.Run("Error simulation", func(t *testing.T) {
		handler := NewMockHandler(t)

		// Set handler function
		handler.SetHandler("test_method", func(args ...interface{}) (interface{}, error) {
			return "success", nil
		})

		// Set error for method
		testErr := errHandlerError
		handler.SetMethodError("test_method", testErr)

		// Should return error instead of calling handler
		result, err := handler.Handle("test_method", "arg1")
		assert.Equal(t, testErr, err)
		assert.Nil(t, result)
	})
}

// Example of how the MockBase can be used to reduce duplication
func TestMockBaseUsageExample(t *testing.T) {
	// This demonstrates how our generic MockBase reduces duplication

	t.Run("Comparison with existing mock patterns", func(t *testing.T) {
		// Before: Each test file creates its own mock with similar patterns
		// After: Use our generic MockBase to reduce duplication

		// Create a custom mock using our base
		type CustomMock struct {
			*MockBase
		}

		custom := &CustomMock{
			MockBase: NewMockBase(t),
		}

		// Add custom method
		customMethod := func(input string) (string, error) {
			if err := custom.ShouldReturnError("CustomMethod"); err != nil {
				custom.RecordCall("CustomMethod", []interface{}{input}, nil, err)
				return "", err
			}

			result := "processed: " + input
			custom.RecordCall("CustomMethod", []interface{}{input}, []interface{}{result}, nil)
			return result, nil
		}

		// Test the custom method
		result, err := customMethod("test")
		require.NoError(t, err)
		assert.Equal(t, "processed: test", result)

		// Use built-in assertion methods
		custom.AssertCalled("CustomMethod")
		custom.AssertCalledWith("CustomMethod", "test")
		custom.AssertCalledTimes("CustomMethod", 1)

		// Test error simulation
		custom.SetMethodError("CustomMethod", errCustomError)
		result, err = customMethod("test2")
		require.Error(t, err)
		assert.Empty(t, result)
		assert.Contains(t, err.Error(), "custom error")

		// Verify call recording continues to work
		assert.Equal(t, 2, custom.CallCount("CustomMethod"))
	})
}
