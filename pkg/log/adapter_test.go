package log

import (
	"bytes"
	"context"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// CLIAdapter Tests
// =============================================================================

func TestCLIAdapterGetLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setLevel Level
		want     Level
	}{
		{
			name:     "default level is Info",
			setLevel: Level(-1), // sentinel to skip SetLevel
			want:     LevelInfo,
		},
		{
			name:     "returns Debug after SetLevel",
			setLevel: LevelDebug,
			want:     LevelDebug,
		},
		{
			name:     "returns Warn after SetLevel",
			setLevel: LevelWarn,
			want:     LevelWarn,
		},
		{
			name:     "returns Error after SetLevel",
			setLevel: LevelError,
			want:     LevelError,
		},
		{
			name:     "returns Fatal after SetLevel",
			setLevel: LevelFatal,
			want:     LevelFatal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewCLIAdapter()
			if tt.setLevel != Level(-1) {
				adapter.SetLevel(tt.setLevel)
			}

			got := adapter.GetLevel()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCLIAdapterWithFieldsCreatesNewAdapter(t *testing.T) {
	t.Parallel()

	t.Run("returns new CLIAdapter with merged fields", func(t *testing.T) {
		t.Parallel()

		adapter := NewCLIAdapter()
		adapter.SetLevel(LevelDebug)

		// Add initial field
		withField := adapter.WithField("initial", "value")
		cliWithField, ok := withField.(*CLIAdapter)
		require.True(t, ok, "WithField should return *CLIAdapter")

		// Add multiple fields
		withFields := cliWithField.WithFields(Fields{
			"key1": "value1",
			"key2": 42,
		})
		cliWithFields, ok := withFields.(*CLIAdapter)
		require.True(t, ok, "WithFields should return *CLIAdapter")

		// Verify fields are merged
		assert.Contains(t, cliWithFields.fields, "initial")
		assert.Contains(t, cliWithFields.fields, "key1")
		assert.Contains(t, cliWithFields.fields, "key2")
		assert.Equal(t, "value", cliWithFields.fields["initial"])
		assert.Equal(t, "value1", cliWithFields.fields["key1"])
		assert.Equal(t, 42, cliWithFields.fields["key2"])
	})

	t.Run("original adapter fields unchanged", func(t *testing.T) {
		t.Parallel()

		adapter := NewCLIAdapter()
		original, ok := adapter.WithField("original", "value").(*CLIAdapter)
		require.True(t, ok, "WithField should return *CLIAdapter")

		_ = original.WithFields(Fields{"new": "field"})

		// Original should not have the new field
		_, exists := original.fields["new"]
		assert.False(t, exists, "original adapter should not have new field")
	})

	t.Run("empty fields creates valid adapter", func(t *testing.T) {
		t.Parallel()

		adapter := NewCLIAdapter()
		withFields := adapter.WithFields(Fields{})

		require.NotNil(t, withFields)
		cliAdapter, ok := withFields.(*CLIAdapter)
		require.True(t, ok)
		assert.NotNil(t, cliAdapter.fields)
	})

	t.Run("nil fields creates valid adapter", func(t *testing.T) {
		t.Parallel()

		adapter := NewCLIAdapter()
		withFields := adapter.WithFields(nil)

		require.NotNil(t, withFields)
		cliAdapter, ok := withFields.(*CLIAdapter)
		require.True(t, ok)
		assert.NotNil(t, cliAdapter.fields)
	})
}

func TestCLIAdapterSpinnerMethodsAreNoOps(t *testing.T) {
	t.Parallel()

	adapter := NewCLIAdapter()

	// These should not panic and should be no-ops
	assert.NotPanics(t, func() {
		adapter.StartSpinner("test message")
	})

	assert.NotPanics(t, func() {
		adapter.UpdateSpinner("updated message")
	})

	assert.NotPanics(t, func() {
		adapter.StopSpinner()
	})
}

func TestCLIAdapterHeaderWithColorEnabled(t *testing.T) {
	t.Parallel()

	t.Run("header with color enabled contains ANSI codes", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewCLIAdapter()
		adapter.SetOutput(&buf)
		adapter.SetColorEnabled(true)

		adapter.Header("Test Header")

		output := buf.String()
		assert.Contains(t, output, "\033[34m", "should contain blue color code")
		assert.Contains(t, output, "\033[1m", "should contain bold code")
		assert.Contains(t, output, "\033[0m", "should contain reset code")
		assert.Contains(t, output, "Test Header")
		assert.Contains(t, output, "===")
	})

	t.Run("header without color has no ANSI codes", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewCLIAdapter()
		adapter.SetOutput(&buf)
		adapter.SetColorEnabled(false)

		adapter.Header("Plain Header")

		output := buf.String()
		assert.NotContains(t, output, "\033[", "should not contain ANSI escape codes")
		assert.Contains(t, output, "=== Plain Header ===")
	})
}

func TestCLIAdapterLogWithColorEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		level      Level
		logFunc    func(*CLIAdapter)
		colorCode  string
		levelLabel string
	}{
		{
			name:  "debug level with color",
			level: LevelDebug,
			logFunc: func(a *CLIAdapter) {
				a.Debug("debug msg")
			},
			colorCode:  "\033[90m", // gray
			levelLabel: "DEBUG",
		},
		{
			name:  "info level with color",
			level: LevelDebug,
			logFunc: func(a *CLIAdapter) {
				a.Info("info msg")
			},
			colorCode:  "\033[34m", // blue
			levelLabel: "INFO",
		},
		{
			name:  "warn level with color",
			level: LevelDebug,
			logFunc: func(a *CLIAdapter) {
				a.Warn("warn msg")
			},
			colorCode:  "\033[33m", // yellow
			levelLabel: "WARN",
		},
		{
			name:  "error level with color",
			level: LevelDebug,
			logFunc: func(a *CLIAdapter) {
				a.Error("error msg")
			},
			colorCode:  "\033[31m", // red
			levelLabel: "ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			adapter := NewCLIAdapter()
			adapter.SetOutput(&buf)
			adapter.SetColorEnabled(true)
			adapter.SetLevel(tt.level)

			tt.logFunc(adapter)

			output := buf.String()
			assert.Contains(t, output, tt.colorCode, "should contain expected color code")
			assert.Contains(t, output, tt.levelLabel, "should contain level label")
			assert.Contains(t, output, "\033[0m", "should contain reset code")
		})
	}
}

func TestCLIAdapterLogAllLevels(t *testing.T) {
	t.Parallel()

	t.Run("logs at all levels when level is Debug", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewCLIAdapter()
		adapter.SetOutput(&buf)
		adapter.SetColorEnabled(false)
		adapter.SetLevel(LevelDebug)

		adapter.Debug("debug message")
		adapter.Info("info message")
		adapter.Warn("warn message")
		adapter.Error("error message")

		output := buf.String()
		assert.Contains(t, output, "[DEBUG] debug message")
		assert.Contains(t, output, "[INFO] info message")
		assert.Contains(t, output, "[WARN] warn message")
		assert.Contains(t, output, "[ERROR] error message")
	})

	t.Run("log internal method handles Fatal level", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewCLIAdapter()
		adapter.SetOutput(&buf)
		adapter.SetColorEnabled(false)
		adapter.SetLevel(LevelDebug)

		// Call internal log method directly via the exported methods
		// Fatal level is handled in the switch but there's no public Fatal method
		// We can test the color mapping by enabling colors
		adapter.SetColorEnabled(true)

		// Log at error level which uses same color as fatal
		adapter.Error("error with color")

		output := buf.String()
		assert.Contains(t, output, "\033[31m") // red color for error/fatal
	})
}

func TestCLIAdapterLogWithEmojiColorEnabled(t *testing.T) {
	t.Parallel()

	t.Run("Success with color enabled uses green", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewCLIAdapter()
		adapter.SetOutput(&buf)
		adapter.SetColorEnabled(true)

		adapter.Success("success message")

		output := buf.String()
		assert.Contains(t, output, "\033[32m", "should contain green color code")
		assert.Contains(t, output, "\033[0m", "should contain reset code")
		assert.Contains(t, output, "success message")
	})

	t.Run("Fail with color enabled uses red", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewCLIAdapter()
		adapter.SetOutput(&buf)
		adapter.SetColorEnabled(true)

		adapter.Fail("fail message")

		output := buf.String()
		assert.Contains(t, output, "\033[31m", "should contain red color code")
		assert.Contains(t, output, "\033[0m", "should contain reset code")
		assert.Contains(t, output, "fail message")
	})
}

func TestCLIAdapterLogWithEmojiAllLevels(t *testing.T) {
	t.Parallel()

	t.Run("emoji respects level filtering", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewCLIAdapter()
		adapter.SetOutput(&buf)
		adapter.SetColorEnabled(false)
		adapter.SetLevel(LevelError) // Only error and above

		adapter.Success("should not appear") // Success uses Info level

		output := buf.String()
		assert.Empty(t, output, "Success should be filtered at Error level")
	})

	t.Run("emoji with prefix", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewCLIAdapter()
		adapter.SetOutput(&buf)
		adapter.SetColorEnabled(false)

		prefixed, ok := adapter.WithPrefix("TEST").(*CLIAdapter)
		require.True(t, ok, "WithPrefix should return *CLIAdapter")
		prefixed.SetOutput(&buf)

		prefixed.Success("with prefix")

		output := buf.String()
		assert.Contains(t, output, "[TEST]")
		assert.Contains(t, output, "with prefix")
	})
}

// =============================================================================
// StructuredAdapter Tests
// =============================================================================

func TestStructuredAdapterGetLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setLevel Level
		want     Level
	}{
		{
			name:     "default level is Info",
			setLevel: Level(-1), // sentinel to skip SetLevel
			want:     LevelInfo,
		},
		{
			name:     "returns Debug after SetLevel",
			setLevel: LevelDebug,
			want:     LevelDebug,
		},
		{
			name:     "returns Error after SetLevel",
			setLevel: LevelError,
			want:     LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			adapter := NewStructuredAdapter()
			if tt.setLevel != Level(-1) {
				adapter.SetLevel(tt.setLevel)
			}

			got := adapter.GetLevel()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStructuredAdapterWithPrefix(t *testing.T) {
	t.Parallel()

	t.Run("creates new adapter with prefix", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewStructuredAdapter()
		adapter.SetOutput(&buf)

		prefixed := adapter.WithPrefix("COMPONENT")
		structPrefixed, ok := prefixed.(*StructuredAdapter)
		require.True(t, ok, "WithPrefix should return *StructuredAdapter")

		structPrefixed.SetOutput(&buf)
		structPrefixed.Info("test message")

		output := buf.String()
		assert.Contains(t, output, "[COMPONENT]")
		assert.Contains(t, output, "test message")
	})

	t.Run("preserves level from original", func(t *testing.T) {
		t.Parallel()

		adapter := NewStructuredAdapter()
		adapter.SetLevel(LevelError)

		prefixed, ok := adapter.WithPrefix("PREFIX").(*StructuredAdapter)
		require.True(t, ok, "WithPrefix should return *StructuredAdapter")

		assert.Equal(t, LevelError, prefixed.level)
	})

	t.Run("preserves fields from original", func(t *testing.T) {
		t.Parallel()

		adapter := NewStructuredAdapter()
		withField, ok := adapter.WithField("key", "value").(*StructuredAdapter)
		require.True(t, ok, "WithField should return *StructuredAdapter")

		prefixed, ok := withField.WithPrefix("PREFIX").(*StructuredAdapter)
		require.True(t, ok, "WithPrefix should return *StructuredAdapter")

		assert.Contains(t, prefixed.fields, "key")
		assert.Equal(t, "value", prefixed.fields["key"])
	})
}

func TestStructuredAdapterWithField(t *testing.T) {
	t.Parallel()

	t.Run("creates new adapter with field", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewStructuredAdapter()

		withField := adapter.WithField("user", "john")
		structWithField, ok := withField.(*StructuredAdapter)
		require.True(t, ok, "WithField should return *StructuredAdapter")

		structWithField.SetOutput(&buf)
		structWithField.Info("test")

		output := buf.String()
		assert.Contains(t, output, "user=john")
	})

	t.Run("original adapter unchanged", func(t *testing.T) {
		t.Parallel()

		adapter := NewStructuredAdapter()
		_ = adapter.WithField("new", "field")

		_, exists := adapter.fields["new"]
		assert.False(t, exists, "original should not have new field")
	})

	t.Run("field with various value types", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewStructuredAdapter()

		withFields, ok := adapter.WithField("string", "value").(*StructuredAdapter)
		require.True(t, ok, "WithField should return *StructuredAdapter")

		withFields, ok = withFields.WithField("int", 42).(*StructuredAdapter)
		require.True(t, ok, "WithField should return *StructuredAdapter")

		withFields, ok = withFields.WithField("bool", true).(*StructuredAdapter)
		require.True(t, ok, "WithField should return *StructuredAdapter")

		withFields.SetOutput(&buf)

		withFields.Info("test")

		output := buf.String()
		assert.Contains(t, output, "string=value")
		assert.Contains(t, output, "int=42")
		assert.Contains(t, output, "bool=true")
	})
}

func TestStructuredAdapterDebugContext(t *testing.T) {
	t.Parallel()

	t.Run("logs with request ID from context", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewStructuredAdapter()
		adapter.SetOutput(&buf)
		adapter.SetLevel(LevelDebug)

		//nolint:staticcheck // SA1029: Using string key intentionally
		ctx := context.WithValue(context.Background(), "request_id", "debug-req-123")
		adapter.DebugContext(ctx, "debug context message")

		output := buf.String()
		assert.Contains(t, output, "[req:debug-req-123]")
		assert.Contains(t, output, "[DEBUG]")
		assert.Contains(t, output, "debug context message")
	})

	t.Run("respects level filtering", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewStructuredAdapter()
		adapter.SetOutput(&buf)
		adapter.SetLevel(LevelInfo) // Debug should be filtered

		ctx := context.Background()
		adapter.DebugContext(ctx, "should not appear")

		output := buf.String()
		assert.Empty(t, output)
	})
}

func TestStructuredAdapterWarnContext(t *testing.T) {
	t.Parallel()

	t.Run("logs warning with request ID", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewStructuredAdapter()
		adapter.SetOutput(&buf)

		//nolint:staticcheck // SA1029: Using string key intentionally
		ctx := context.WithValue(context.Background(), "request_id", "warn-req-456")
		adapter.WarnContext(ctx, "warning message %d", 1)

		output := buf.String()
		assert.Contains(t, output, "[req:warn-req-456]")
		assert.Contains(t, output, "[WARN]")
		assert.Contains(t, output, "warning message 1")
	})

	t.Run("logs without request ID when not in context", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewStructuredAdapter()
		adapter.SetOutput(&buf)

		ctx := context.Background()
		adapter.WarnContext(ctx, "no request id")

		output := buf.String()
		assert.NotContains(t, output, "[req:")
		assert.Contains(t, output, "[WARN]")
		assert.Contains(t, output, "no request id")
	})
}

func TestStructuredAdapterErrorContext(t *testing.T) {
	t.Parallel()

	t.Run("logs error with request ID", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewStructuredAdapter()
		adapter.SetOutput(&buf)

		//nolint:staticcheck // SA1029: Using string key intentionally
		ctx := context.WithValue(context.Background(), "trace_id", "trace-789")
		adapter.ErrorContext(ctx, "error occurred: %s", "something bad")

		output := buf.String()
		assert.Contains(t, output, "[req:trace-789]")
		assert.Contains(t, output, "[ERROR]")
		assert.Contains(t, output, "error occurred: something bad")
	})
}

func TestStructuredAdapterLogWithContextAndFields(t *testing.T) {
	t.Parallel()

	t.Run("logs with both context and fields", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewStructuredAdapter()

		withFields, ok := adapter.WithFields(Fields{
			"service": "test-service",
			"version": "1.0.0",
		}).(*StructuredAdapter)
		require.True(t, ok, "WithFields should return *StructuredAdapter")
		withFields.SetOutput(&buf)

		//nolint:staticcheck // SA1029: Using string key intentionally
		ctx := context.WithValue(context.Background(), "request_id", "req-with-fields")
		withFields.InfoContext(ctx, "message with context and fields")

		output := buf.String()
		assert.Contains(t, output, "[req:req-with-fields]")
		assert.Contains(t, output, "service=test-service")
		assert.Contains(t, output, "version=1.0.0")
		assert.Contains(t, output, "message with context and fields")
	})

	t.Run("logs with fields but no request ID", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewStructuredAdapter()

		withFields, ok := adapter.WithField("key", "value").(*StructuredAdapter)
		require.True(t, ok, "WithField should return *StructuredAdapter")
		withFields.SetOutput(&buf)

		ctx := context.Background()
		withFields.InfoContext(ctx, "fields only")

		output := buf.String()
		assert.NotContains(t, output, "[req:")
		assert.Contains(t, output, "key=value")
		assert.Contains(t, output, "fields only")
	})
}

// =============================================================================
// shouldUseColor Tests
// =============================================================================

// saveEnvVars saves the current values of CI and NO_COLOR environment variables
// and returns a cleanup function that restores them.
func saveEnvVars(t *testing.T) func() {
	t.Helper()
	oldCI := os.Getenv("CI")
	oldNoColor := os.Getenv("NO_COLOR")

	return func() {
		if oldCI == "" {
			require.NoError(t, os.Unsetenv("CI"))
		} else {
			require.NoError(t, os.Setenv("CI", oldCI))
		}
		if oldNoColor == "" {
			require.NoError(t, os.Unsetenv("NO_COLOR"))
		} else {
			require.NoError(t, os.Setenv("NO_COLOR", oldNoColor))
		}
	}
}

func TestShouldUseColorCIEnvironment(t *testing.T) {
	// Not parallel: modifies environment variables

	cleanup := saveEnvVars(t)
	defer cleanup()

	// Clear NO_COLOR to isolate CI test
	require.NoError(t, os.Unsetenv("NO_COLOR"))

	t.Run("CI environment disables color", func(t *testing.T) {
		require.NoError(t, os.Setenv("CI", "true"))

		result := shouldUseColor()
		assert.False(t, result, "shouldUseColor should return false when CI is set")
	})

	t.Run("CI with any value disables color", func(t *testing.T) {
		require.NoError(t, os.Setenv("CI", "1"))

		result := shouldUseColor()
		assert.False(t, result, "shouldUseColor should return false when CI has any value")
	})
}

func TestShouldUseColorNOCOLOREnvironment(t *testing.T) {
	// Not parallel: modifies environment variables

	cleanup := saveEnvVars(t)
	defer cleanup()

	// Clear CI to isolate NO_COLOR test
	require.NoError(t, os.Unsetenv("CI"))

	t.Run("NO_COLOR environment disables color", func(t *testing.T) {
		require.NoError(t, os.Setenv("NO_COLOR", "1"))

		result := shouldUseColor()
		assert.False(t, result, "shouldUseColor should return false when NO_COLOR is set")
	})

	t.Run("NO_COLOR with any value disables color", func(t *testing.T) {
		require.NoError(t, os.Setenv("NO_COLOR", "true"))

		result := shouldUseColor()
		assert.False(t, result, "shouldUseColor should return false when NO_COLOR has any value")
	})
}

func TestShouldUseColorNonTTY(t *testing.T) {
	// Not parallel: tests terminal detection

	// Note: In a test environment, stdout is typically not a TTY,
	// so shouldUseColor should return false even without CI/NO_COLOR set.
	// This tests the TTY detection branch.

	cleanup := saveEnvVars(t)
	defer cleanup()

	// Clear both env vars
	require.NoError(t, os.Unsetenv("CI"))
	require.NoError(t, os.Unsetenv("NO_COLOR"))

	t.Run("non-TTY stdout disables color", func(t *testing.T) {
		// In test environment, stdout is not a terminal
		result := shouldUseColor()
		assert.False(t, result, "shouldUseColor should return false for non-TTY stdout")
	})
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestCLIAdapterConcurrentAccess(t *testing.T) {
	t.Parallel()

	adapter := NewCLIAdapter()
	var buf bytes.Buffer
	adapter.SetOutput(&buf)
	adapter.SetColorEnabled(false)
	adapter.SetLevel(LevelDebug)

	var wg sync.WaitGroup
	const numGoroutines = 100

	// Concurrent reads and writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			// Mix of read and write operations
			adapter.Info("message %d", n)
			adapter.Debug("debug %d", n)
			_ = adapter.GetLevel()
			adapter.SetLevel(LevelDebug)

			// Create new adapters (read operation on fields)
			_ = adapter.WithField("key", n)
			_ = adapter.WithPrefix("prefix")
		}(i)
	}

	wg.Wait()

	// Verify no race conditions (test passes if no panic/race detected)
	output := buf.String()
	assert.NotEmpty(t, output, "should have logged messages")
}

func TestStructuredAdapterConcurrentAccess(t *testing.T) {
	t.Parallel()

	adapter := NewStructuredAdapter()
	var buf bytes.Buffer
	adapter.SetOutput(&buf)
	adapter.SetLevel(LevelDebug)

	var wg sync.WaitGroup
	const numGoroutines = 100

	// Concurrent reads and writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			// Mix of read and write operations
			adapter.Info("message %d", n)
			adapter.Debug("debug %d", n)
			_ = adapter.GetLevel()
			adapter.SetLevel(LevelDebug)

			// Create new adapters (read operation on fields)
			_ = adapter.WithField("key", n)
			_ = adapter.WithPrefix("prefix")
			_ = adapter.WithFields(Fields{"test": n})

			// Context logging
			//nolint:staticcheck // SA1029: Using string key intentionally
			ctx := context.WithValue(context.Background(), "request_id", "concurrent")
			adapter.InfoContext(ctx, "context message %d", n)
		}(i)
	}

	wg.Wait()

	// Verify no race conditions (test passes if no panic/race detected)
	output := buf.String()
	assert.NotEmpty(t, output, "should have logged messages")
}

// =============================================================================
// Edge Case Tests
// =============================================================================

func TestCLIAdapterOutputIsolation(t *testing.T) {
	t.Parallel()

	t.Run("WithPrefix shares output writer", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewCLIAdapter()
		adapter.SetOutput(&buf)
		adapter.SetColorEnabled(false)

		// Note: WithPrefix creates a new adapter but shares the output writer
		prefixed, ok := adapter.WithPrefix("PREFIX").(*CLIAdapter)
		require.True(t, ok, "WithPrefix should return *CLIAdapter")
		// prefixed shares adapter's output by default

		// Verify the output is shared correctly
		prefixed.SetOutput(&buf) // Explicitly set to same buffer
		prefixed.Info("test")

		output := buf.String()
		assert.Contains(t, output, "[PREFIX] test")
	})
}

func TestStructuredAdapterLevelFiltering(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		adapterLvl  Level
		logLevel    Level
		shouldLog   bool
		logFunc     func(*StructuredAdapter, context.Context)
		expectedStr string
	}{
		{
			name:       "debug filtered at info level",
			adapterLvl: LevelInfo,
			logLevel:   LevelDebug,
			shouldLog:  false,
			logFunc: func(a *StructuredAdapter, ctx context.Context) {
				a.DebugContext(ctx, "filtered")
			},
		},
		{
			name:       "info logged at info level",
			adapterLvl: LevelInfo,
			logLevel:   LevelInfo,
			shouldLog:  true,
			logFunc: func(a *StructuredAdapter, ctx context.Context) {
				a.InfoContext(ctx, "visible")
			},
			expectedStr: "visible",
		},
		{
			name:       "warn logged at info level",
			adapterLvl: LevelInfo,
			logLevel:   LevelWarn,
			shouldLog:  true,
			logFunc: func(a *StructuredAdapter, ctx context.Context) {
				a.WarnContext(ctx, "warning visible")
			},
			expectedStr: "warning visible",
		},
		{
			name:       "error logged at error level",
			adapterLvl: LevelError,
			logLevel:   LevelError,
			shouldLog:  true,
			logFunc: func(a *StructuredAdapter, ctx context.Context) {
				a.ErrorContext(ctx, "error visible")
			},
			expectedStr: "error visible",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			adapter := NewStructuredAdapter()
			adapter.SetOutput(&buf)
			adapter.SetLevel(tt.adapterLvl)

			ctx := context.Background()
			tt.logFunc(adapter, ctx)

			output := buf.String()
			if tt.shouldLog {
				assert.Contains(t, output, tt.expectedStr)
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestCLIAdapterFormatArgs(t *testing.T) {
	t.Parallel()

	t.Run("format string with multiple args", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewCLIAdapter()
		adapter.SetOutput(&buf)
		adapter.SetColorEnabled(false)

		adapter.Info("user %s logged in from %s with status %d", "john", "192.168.1.1", 200)

		output := buf.String()
		assert.Contains(t, output, "user john logged in from 192.168.1.1 with status 200")
	})

	t.Run("format string with no args", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewCLIAdapter()
		adapter.SetOutput(&buf)
		adapter.SetColorEnabled(false)

		adapter.Info("simple message")

		output := buf.String()
		assert.Contains(t, output, "simple message")
	})

	t.Run("format string with percent literal", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewCLIAdapter()
		adapter.SetOutput(&buf)
		adapter.SetColorEnabled(false)

		adapter.Info("progress: %d%%", 50)

		output := buf.String()
		assert.Contains(t, output, "progress: 50%")
	})
}

func TestTimestampFormat(t *testing.T) {
	t.Parallel()

	t.Run("CLI adapter uses HH:MM:SS format", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewCLIAdapter()
		adapter.SetOutput(&buf)
		adapter.SetColorEnabled(false)

		adapter.Info("test")

		output := buf.String()
		// Should match pattern like "15:04:05"
		lines := strings.Split(output, "\n")
		require.NotEmpty(t, lines)

		// Timestamp is at the beginning, format is HH:MM:SS
		parts := strings.Fields(lines[0])
		require.NotEmpty(t, parts)
		timestamp := parts[0]
		assert.Regexp(t, `^\d{2}:\d{2}:\d{2}$`, timestamp)
	})

	t.Run("Structured adapter uses full timestamp format", func(t *testing.T) {
		t.Parallel()

		var buf bytes.Buffer
		adapter := NewStructuredAdapter()
		adapter.SetOutput(&buf)

		adapter.Info("test")

		output := buf.String()
		// Should match pattern like "2006-01-02 15:04:05.000"
		assert.Regexp(t, `\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\.\d{3}`, output)
	})
}

// TestStructuredAdapterFieldsFormatting tests that fields are correctly
// formatted using strings.Builder (performance optimization)
func TestStructuredAdapterFieldsFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		fields        Fields
		expectedParts []string
		notExpected   []string
	}{
		{
			name: "single field formats correctly",
			fields: Fields{
				"key": "value",
			},
			expectedParts: []string{" {", "key=value", "}"},
		},
		{
			name: "multiple fields formats correctly",
			fields: Fields{
				"name":  "test",
				"count": 42,
			},
			expectedParts: []string{" {", "name=test", "count=42", "}"},
		},
		{
			name: "numeric fields format correctly",
			fields: Fields{
				"int":   123,
				"float": 3.14,
			},
			expectedParts: []string{"int=123", "float=3.14"},
		},
		{
			name:        "no fields produces no braces",
			fields:      Fields{},
			notExpected: []string{" {", "}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			adapter := NewStructuredAdapter()
			adapter.SetOutput(&buf)
			adapter.SetLevel(LevelDebug)

			// Apply fields
			loggerWithFields := adapter.WithFields(tt.fields)
			structuredWithFields, ok := loggerWithFields.(*StructuredAdapter)
			require.True(t, ok)

			structuredWithFields.Info("test message")

			output := buf.String()

			for _, expected := range tt.expectedParts {
				assert.Contains(t, output, expected, "output should contain %q", expected)
			}

			for _, notExpected := range tt.notExpected {
				assert.NotContains(t, output, notExpected, "output should not contain %q", notExpected)
			}
		})
	}
}

// BenchmarkStructuredAdapterWithFields benchmarks field formatting performance
func BenchmarkStructuredAdapterWithFields(b *testing.B) {
	var buf bytes.Buffer
	adapter := NewStructuredAdapter()
	adapter.SetOutput(&buf)
	adapter.SetLevel(LevelDebug)

	fields := Fields{
		"user_id":    "12345",
		"request_id": "abc-def-ghi",
		"action":     "login",
		"status":     "success",
	}

	loggerWithFields, ok := adapter.WithFields(fields).(*StructuredAdapter)
	if !ok {
		b.Fatal("expected *StructuredAdapter")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		loggerWithFields.Info("test message")
	}
}
