package mockrec

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// Simple mock receiver for testing - just needs to be a valid receiver
type mockReceiver struct{}

// TestReporter implements gomock.TestReporter but doesn't enforce expectations
type TestReporter struct{}

func (tr TestReporter) Errorf(format string, args ...interface{}) {}
func (tr TestReporter) Fatalf(format string, args ...interface{}) {}
func (tr TestReporter) Helper()                                   {}

// Create a test controller that won't enforce missing call expectations
func newTestController() *gomock.Controller {
	return gomock.NewController(TestReporter{})
}

// Test helper to create a simple receiver
func newMockReceiver() *mockReceiver {
	return &mockReceiver{}
}

func TestRecordCall(t *testing.T) {
	t.Run("RecordCall_CreatesValidCall_DoesNotPanic", func(t *testing.T) {
		ctrl := newTestController()
		receiver := newMockReceiver()
		methodType := reflect.TypeOf((func(string, int) error)(nil))

		// Test that RecordCall doesn't panic and returns a valid Call
		call := RecordCall(ctrl, receiver, "TestMethod", methodType, "arg1", 42)

		require.NotNil(t, call, "RecordCall should return a valid call")
		assert.IsType(t, &gomock.Call{}, call, "Should return a gomock.Call")
	})

	t.Run("RecordCall_HandlesVariousArgTypes_DoesNotPanic", func(t *testing.T) {
		ctrl := newTestController()
		receiver := newMockReceiver()

		tests := []struct {
			name string
			args []interface{}
		}{
			{"StringAndInt", []interface{}{"test", 42}},
			{"NilValue", []interface{}{nil}},
			{"EmptyString", []interface{}{""}},
			{"MapValue", []interface{}{map[string]string{"key": "value"}}},
			{"NoArgs", []interface{}{}},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				methodType := reflect.TypeOf((func(...interface{}) interface{})(nil))
				call := RecordCall(ctrl, receiver, "TestMethod", methodType, test.args...)
				assert.NotNil(t, call, "RecordCall should work with %s", test.name)
			})
		}
	})
}

func TestRecordNoArgsCall(t *testing.T) {
	t.Run("RecordNoArgsCall_CreatesValidCall_DoesNotPanic", func(t *testing.T) {
		ctrl := newTestController()
		receiver := newMockReceiver()
		methodType := reflect.TypeOf((func() string)(nil))

		// Test that RecordNoArgsCall doesn't panic and returns a valid Call
		call := RecordNoArgsCall(ctrl, receiver, "TestMethod", methodType)

		require.NotNil(t, call, "RecordNoArgsCall should return a valid call")
		assert.IsType(t, &gomock.Call{}, call, "Should return a gomock.Call")
	})

	t.Run("RecordNoArgsCall_HandlesVariousReturnTypes_DoesNotPanic", func(t *testing.T) {
		ctrl := newTestController()
		receiver := newMockReceiver()

		tests := []struct {
			name       string
			methodType reflect.Type
		}{
			{"StringReturn", reflect.TypeOf((func() string)(nil))},
			{"ErrorReturn", reflect.TypeOf((func() error)(nil))},
			{"MultipleReturns", reflect.TypeOf((func() (string, error))(nil))},
			{"InterfaceReturn", reflect.TypeOf((func() interface{})(nil))},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				call := RecordNoArgsCall(ctrl, receiver, "TestMethod", test.methodType)
				assert.NotNil(t, call, "RecordNoArgsCall should work with %s", test.name)
			})
		}
	})
}

func TestHelperFunctionality(t *testing.T) {
	t.Run("BothHelpers_CallControllerTHelper_DoNotPanic", func(t *testing.T) {
		ctrl := newTestController()
		receiver := newMockReceiver()

		// Test RecordCall
		methodType1 := reflect.TypeOf((func(string, int) error)(nil))
		call1 := RecordCall(ctrl, receiver, "MethodWithArgs", methodType1, "test", 123)
		assert.NotNil(t, call1, "RecordCall should work")

		// Test RecordNoArgsCall
		methodType2 := reflect.TypeOf((func() string)(nil))
		call2 := RecordNoArgsCall(ctrl, receiver, "MethodNoArgs", methodType2)
		assert.NotNil(t, call2, "RecordNoArgsCall should work")

		// Both should return proper gomock.Call instances
		assert.IsType(t, &gomock.Call{}, call1, "RecordCall should return gomock.Call")
		assert.IsType(t, &gomock.Call{}, call2, "RecordNoArgsCall should return gomock.Call")
	})
}

func TestReflectionFunctionality(t *testing.T) {
	t.Run("Helpers_WorkWithReflectionTypes_DoNotPanic", func(t *testing.T) {
		ctrl := newTestController()
		receiver := newMockReceiver()

		// Test various method signatures with RecordCall
		methodTypes := []reflect.Type{
			reflect.TypeOf((func(string, int) error)(nil)),
			reflect.TypeOf((func(interface{}) bool)(nil)),
			reflect.TypeOf((func(...string) interface{})(nil)),
		}

		for i, methodType := range methodTypes {
			call := RecordCall(ctrl, receiver, "TestMethod", methodType, "arg1", i)
			assert.NotNil(t, call, "RecordCall should work with method type %d", i)
		}

		// Test various return types with RecordNoArgsCall
		returnTypes := []reflect.Type{
			reflect.TypeOf((func() string)(nil)),
			reflect.TypeOf((func() (string, error))(nil)),
			reflect.TypeOf((func() interface{})(nil)),
		}

		for i, returnType := range returnTypes {
			call := RecordNoArgsCall(ctrl, receiver, "TestMethod", returnType)
			assert.NotNil(t, call, "RecordNoArgsCall should work with return type %d", i)
		}
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("RecordCall_WithVariadicArgs_HandlesCorrectly", func(t *testing.T) {
		ctrl := newTestController()
		receiver := newMockReceiver()

		// Test with multiple individual args (simulating variadic expansion)
		methodType := reflect.TypeOf((func(string, int) error)(nil))
		args := []interface{}{"variadic_test", 42}

		call := RecordCall(ctrl, receiver, "MethodWithArgs", methodType, args...)
		assert.NotNil(t, call, "Call should handle variadic args correctly")
		assert.IsType(t, &gomock.Call{}, call, "Should return a gomock.Call")
	})

	t.Run("RecordCall_WithNoArgs_WorksCorrectly", func(t *testing.T) {
		ctrl := newTestController()
		receiver := newMockReceiver()

		// Test RecordCall with no arguments
		methodType := reflect.TypeOf((func() string)(nil))
		call := RecordCall(ctrl, receiver, "MethodNoArgs", methodType)

		assert.NotNil(t, call, "RecordCall with no args should work")
		assert.IsType(t, &gomock.Call{}, call, "Should return a gomock.Call")
	})
}

func TestIntegrationWithGomockController(t *testing.T) {
	t.Run("HelperFunctions_IntegrateWithController_WorkCorrectly", func(t *testing.T) {
		ctrl := newTestController()
		receiver := newMockReceiver()

		// Test that both helper functions work with the same controller
		methodType1 := reflect.TypeOf((func(string, int) error)(nil))
		call1 := RecordCall(ctrl, receiver, "MethodWithArgs", methodType1, "integration", 100)

		methodType2 := reflect.TypeOf((func() string)(nil))
		call2 := RecordNoArgsCall(ctrl, receiver, "MethodNoArgs", methodType2)

		// Both should succeed without panicking
		assert.NotNil(t, call1, "RecordCall should work")
		assert.NotNil(t, call2, "RecordNoArgsCall should work")
		assert.IsType(t, &gomock.Call{}, call1, "RecordCall should return gomock.Call")
		assert.IsType(t, &gomock.Call{}, call2, "RecordNoArgsCall should return gomock.Call")
	})
}

// Benchmark tests to ensure performance is acceptable
func BenchmarkRecordCall(b *testing.B) {
	ctrl := newTestController()
	receiver := newMockReceiver()
	methodType := reflect.TypeOf((func(string, int) error)(nil))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RecordCall(ctrl, receiver, "MethodWithArgs", methodType, "benchmark", i)
	}
}

func BenchmarkRecordNoArgsCall(b *testing.B) {
	ctrl := newTestController()
	receiver := newMockReceiver()
	methodType := reflect.TypeOf((func() string)(nil))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = RecordNoArgsCall(ctrl, receiver, "MethodNoArgs", methodType)
	}
}

// Fuzz tests for robustness
func FuzzRecordCall(f *testing.F) {
	// Add seed corpus
	f.Add("test", 0)
	f.Add("", -1)
	f.Add("fuzz_method", 999999)
	f.Add("special!@#chars", 42)

	f.Fuzz(func(t *testing.T, str string, num int) {
		ctrl := newTestController()
		receiver := newMockReceiver()
		methodType := reflect.TypeOf((func(string, int) error)(nil))

		// Test that RecordCall doesn't panic with various inputs
		call := RecordCall(ctrl, receiver, "MethodWithArgs", methodType, str, num)
		assert.NotNil(t, call, "RecordCall should always return a valid call")
	})
}

func FuzzRecordNoArgsCall(f *testing.F) {
	// Add seed corpus for method names
	f.Add("MethodNoArgs")
	f.Add("MethodMultipleReturns")
	f.Add("")
	f.Add("fuzz!@#method")

	f.Fuzz(func(t *testing.T, methodName string) {
		ctrl := newTestController()
		receiver := newMockReceiver()

		// Use a valid method type regardless of methodName to avoid reflection panics
		methodType := reflect.TypeOf((func() string)(nil))

		// Test that RecordNoArgsCall doesn't panic with various method names
		call := RecordNoArgsCall(ctrl, receiver, methodName, methodType)
		assert.NotNil(t, call, "RecordNoArgsCall should always return a valid call")
	})
}
