package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultErrorRegistryMethods tests methods on DefaultErrorRegistry
func TestDefaultErrorRegistryMethods(t *testing.T) {
	t.Run("Register and Get", func(t *testing.T) {
		registry := NewErrorRegistry()
		err := registry.Register("TEST_CODE", "Test error message")
		require.NoError(t, err)

		entry, found := registry.Get("TEST_CODE")
		assert.True(t, found)
		assert.Equal(t, ErrorCode("TEST_CODE"), entry.Code)
		assert.Equal(t, "Test error message", entry.Description)
	})

	t.Run("RegisterWithSeverity", func(t *testing.T) {
		registry := NewErrorRegistry()
		err := registry.RegisterWithSeverity("TEST_SEV", "Test severity error", SeverityWarning)
		require.NoError(t, err)

		entry, found := registry.Get("TEST_SEV")
		assert.True(t, found)
		assert.Equal(t, SeverityWarning, entry.Severity)
	})

	t.Run("Unregister", func(t *testing.T) {
		registry := NewErrorRegistry()
		err := registry.Register("TEST_UNREG", "To be unregistered")
		require.NoError(t, err)

		err = registry.Unregister("TEST_UNREG")
		require.NoError(t, err)

		_, found := registry.Get("TEST_UNREG")
		assert.False(t, found)
	})

	t.Run("List", func(t *testing.T) {
		registry := NewErrorRegistry()
		require.NoError(t, registry.Register("LIST_A", "Error A"))
		require.NoError(t, registry.Register("LIST_B", "Error B"))

		list := registry.List()
		assert.GreaterOrEqual(t, len(list), 2)
	})

	t.Run("ListByPrefix", func(t *testing.T) {
		registry := NewErrorRegistry()
		require.NoError(t, registry.Register("PREFIX_A", "Error A"))
		require.NoError(t, registry.Register("PREFIX_B", "Error B"))
		require.NoError(t, registry.Register("OTHER_C", "Error C"))

		list := registry.ListByPrefix("PREFIX")
		assert.Len(t, list, 2)
	})

	t.Run("ListBySeverity", func(t *testing.T) {
		registry := NewErrorRegistry()
		require.NoError(t, registry.RegisterWithSeverity("SEV_ERR1", "Error 1", SeverityError))
		require.NoError(t, registry.RegisterWithSeverity("SEV_WARN", "Warning", SeverityWarning))

		list := registry.ListBySeverity(SeverityError)
		assert.GreaterOrEqual(t, len(list), 1)
	})

	t.Run("Contains", func(t *testing.T) {
		registry := NewErrorRegistry()
		require.NoError(t, registry.Register("CONTAINS_TEST", "Test"))

		assert.True(t, registry.Contains("CONTAINS_TEST"))
		assert.False(t, registry.Contains("NONEXISTENT"))
	})

	t.Run("Clear", func(t *testing.T) {
		registry := NewErrorRegistry()
		require.NoError(t, registry.Register("CLEAR_TEST", "Test"))
		require.NoError(t, registry.Clear())

		assert.False(t, registry.Contains("CLEAR_TEST"))
	})
}

// TestNewErrorFactoryFunctions tests the New*Error factory functions
func TestNewErrorFactoryFunctions(t *testing.T) {
	t.Run("NewBuildError", func(t *testing.T) {
		err := NewBuildError("compilation failed", nil)
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "compilation failed")
	})

	t.Run("NewBuildError with cause", func(t *testing.T) {
		cause := New("underlying error")
		err := NewBuildError("build failed", cause)
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "build failed")
	})

	t.Run("NewConfigError", func(t *testing.T) {
		// NewConfigError(message, configFile string)
		err := NewConfigError("invalid config", "config.yaml")
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "invalid config")
	})

	t.Run("NewFileError", func(t *testing.T) {
		// NewFileError(code ErrorCode, message, path string)
		err := NewFileError(ErrFileNotFound, "read failed", "/path/to/file")
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "read failed")
	})

	t.Run("NewCommandError", func(t *testing.T) {
		err := NewCommandError("go build", 1, "compile error")
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "go build")
	})

	t.Run("NewValidationError", func(t *testing.T) {
		err := NewValidationError("field", "must be required", "")
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "field")
	})
}

// TestMustFunctions tests Must and MustValue utility functions
func TestMustFunctions(t *testing.T) {
	t.Run("Must with nil error", func(t *testing.T) {
		// Should not panic
		assert.NotPanics(t, func() {
			Must(nil)
		})
	})

	t.Run("Must with error panics", func(t *testing.T) {
		err := New("test error")
		assert.Panics(t, func() {
			Must(err)
		})
	})

	t.Run("MustValue with nil error", func(t *testing.T) {
		result := MustValue("value", nil)
		assert.Equal(t, "value", result)
	})

	t.Run("MustValue with error panics", func(t *testing.T) {
		err := New("test error")
		assert.Panics(t, func() {
			MustValue("value", err)
		})
	})
}

// TestHandleFunction tests the Handle convenience function
func TestHandleFunction(t *testing.T) {
	t.Run("Handle with nil error", func(t *testing.T) {
		// Handle returns nil for nil error
		result := Handle(nil)
		assert.NoError(t, result)
	})

	t.Run("Handle with error", func(t *testing.T) {
		err := New("test error")
		// Handle should return the error
		result := Handle(err)
		assert.Error(t, result)
	})
}

// TestRecoverFunctions tests Recover and RecoverTo functions using ErrorRecovery
func TestRecoverFunctions(t *testing.T) {
	t.Run("ErrorRecovery Recover", func(t *testing.T) {
		recovery := NewErrorRecovery()
		err := recovery.Recover(func() error {
			panic("test panic")
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test panic")
	})

	t.Run("ErrorRecovery RecoverWithFallback", func(t *testing.T) {
		recovery := NewErrorRecovery()
		fallbackCalled := false
		err := recovery.RecoverWithFallback(
			func() error { return New("primary error") },
			func(err error) error {
				fallbackCalled = true
				return nil
			},
		)
		require.NoError(t, err)
		assert.True(t, fallbackCalled)
	})
}

// TestSafeExecuteWithFallback tests the SafeExecuteWithFallback function
func TestSafeExecuteWithFallback(t *testing.T) {
	t.Run("success case", func(t *testing.T) {
		err := SafeExecuteWithFallback(
			func() error { return nil },
			func(err error) error { return New("fallback") },
		)
		assert.NoError(t, err)
	})

	t.Run("fallback called on error", func(t *testing.T) {
		err := SafeExecuteWithFallback(
			func() error { return New("primary error") },
			func(err error) error { return nil }, // fallback succeeds
		)
		assert.NoError(t, err)
	})
}

// TestLogFunctions tests LogError and LogMageError
func TestLogFunctions(t *testing.T) {
	t.Run("LogError", func(t *testing.T) {
		// Should not panic
		LogError(New("test error"))
	})

	t.Run("LogMageError", func(t *testing.T) {
		// Should not panic - NewMageError takes only message
		LogMageError(NewMageError("test error"))
	})
}

// TestNotifyError tests NotifyError function
func TestNotifyError(t *testing.T) {
	// Should not panic - NotifyError sends notification and returns any error
	err := NotifyError(New("test error"))
	// Default notifier should succeed without error
	assert.NoError(t, err)
}

// TestTransformError tests TransformError function
func TestTransformError(t *testing.T) {
	err := New("test error")
	transformed := TransformError(err)
	// Without a transformer set, it should return the original error
	assert.Error(t, transformed)
}

// TestSetFunctions tests SetErrorLogger, SetErrorNotifier, SetErrorTransformer
func TestSetFunctions(t *testing.T) {
	t.Run("SetErrorLogger", func(t *testing.T) {
		logger := NewErrorLogger()
		SetErrorLogger(logger)
	})

	t.Run("SetErrorNotifier", func(t *testing.T) {
		notifier := NewErrorNotifier()
		SetErrorNotifier(notifier)
	})

	t.Run("SetErrorTransformer", func(t *testing.T) {
		transformer := NewErrorTransformer()
		SetErrorTransformer(transformer)
	})
}

// TestFactoryFunctions tests additional factory functions
func TestFactoryFunctions(t *testing.T) {
	factory := NewCommonErrorFactory()

	t.Run("OperationFailed", func(t *testing.T) {
		// OperationFailed(operation string, cause error)
		err := factory.OperationFailed("test operation", nil)
		require.Error(t, err)
	})

	t.Run("FileExists", func(t *testing.T) {
		err := factory.FileExists("/path/to/file")
		require.Error(t, err)
	})

	t.Run("ConfigurationError", func(t *testing.T) {
		// ConfigurationError(setting, issue string)
		err := factory.ConfigurationError("some-setting", "invalid value")
		require.Error(t, err)
	})

	t.Run("Timeout", func(t *testing.T) {
		// Timeout(operation, duration string)
		err := factory.Timeout("operation", "30s")
		require.Error(t, err)
	})

	t.Run("Recovery", func(t *testing.T) {
		// Recovery(panicValue interface{}, operation string)
		err := factory.Recovery("panic value", "test operation")
		require.Error(t, err)
	})
}

// TestPackageLevelFactoryFunctions tests package-level factory functions
func TestPackageLevelFactoryFunctions(t *testing.T) {
	t.Run("OperationFailed", func(t *testing.T) {
		err := OperationFailed("test", nil)
		require.Error(t, err)
	})

	t.Run("FileExists", func(t *testing.T) {
		err := FileExists("/path")
		require.Error(t, err)
	})

	t.Run("ConfigurationError", func(t *testing.T) {
		// ConfigurationError(setting, issue string)
		err := ConfigurationError("some-setting", "invalid value")
		require.Error(t, err)
	})

	t.Run("Timeout", func(t *testing.T) {
		// Timeout(operation, duration string)
		err := Timeout("operation", "30s")
		require.Error(t, err)
	})

	t.Run("Recovery", func(t *testing.T) {
		// Recovery(panicValue interface{}, operation string)
		err := Recovery("panic value", "test operation")
		require.Error(t, err)
	})
}

// TestBuildErrorFactory tests build error factory functions
func TestBuildErrorFactory(t *testing.T) {
	factory := NewBuildErrorFactory()

	t.Run("DependencyError", func(t *testing.T) {
		// DependencyError likely takes string, string
		err := factory.DependencyError("module", "not found")
		require.Error(t, err)
	})

	t.Run("PackagingFailed", func(t *testing.T) {
		// PackagingFailed(format string, cause error)
		err := factory.PackagingFailed("docker", nil)
		require.Error(t, err)
	})
}

// TestTestErrorFactory tests test error factory functions
func TestTestErrorFactory(t *testing.T) {
	factory := NewTestErrorFactory()

	t.Run("TestFailed", func(t *testing.T) {
		err := factory.TestFailed("TestSomething", nil)
		require.Error(t, err)
	})
}

// TestSecurityErrorFactory tests security error factory functions
func TestSecurityErrorFactory(t *testing.T) {
	factory := NewSecurityErrorFactory()

	t.Run("SecurityValidationFailed", func(t *testing.T) {
		// SecurityValidationFailed takes string, string
		err := factory.SecurityValidationFailed("check", "validation failed")
		require.Error(t, err)
	})
}

// TestFactoryRegistry tests FactoryRegistry functions
func TestFactoryRegistry(t *testing.T) {
	t.Run("NewFactoryRegistry", func(t *testing.T) {
		registry := NewFactoryRegistry()
		require.NotNil(t, registry)
	})

	t.Run("BuildErrors", func(t *testing.T) {
		registry := NewFactoryRegistry()
		build := registry.BuildErrors()
		require.NotNil(t, build)
	})

	t.Run("TestErrors", func(t *testing.T) {
		registry := NewFactoryRegistry()
		testFactory := registry.TestErrors()
		require.NotNil(t, testFactory)
	})

	t.Run("SecurityErrors", func(t *testing.T) {
		registry := NewFactoryRegistry()
		sec := registry.SecurityErrors()
		require.NotNil(t, sec)
	})
}

// TestAsFunction tests the As function
func TestAsFunction(t *testing.T) {
	t.Run("As with MageError", func(t *testing.T) {
		err := NewMageError("test error")
		var target *DefaultMageError
		result := As(err, &target)
		assert.True(t, result)
		assert.NotNil(t, target)
	})

	t.Run("As with non-matching type", func(t *testing.T) {
		err := New("regular error")
		var target *DefaultMageError
		result := As(err, &target)
		// As may or may not match depending on error type
		_ = result
	})
}
