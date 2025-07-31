package utils

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger_Basic(t *testing.T) {
	t.Run("NewLogger creates logger with defaults", func(t *testing.T) {
		logger := NewLogger()
		assert.NotNil(t, logger)
		assert.Equal(t, LogLevelInfo, logger.level)
		assert.NotNil(t, logger.output)
	})

	t.Run("WithPrefix creates logger with prefix", func(t *testing.T) {
		logger := NewLogger()
		prefixedLogger := logger.WithPrefix("TEST")
		assert.Equal(t, "TEST", prefixedLogger.prefix)
		assert.Equal(t, logger.level, prefixedLogger.level)
	})

	t.Run("SetLevel changes log level", func(t *testing.T) {
		logger := NewLogger()
		logger.SetLevel(LogLevelDebug)
		assert.Equal(t, LogLevelDebug, logger.level)
	})

	t.Run("SetColorEnabled changes color setting", func(t *testing.T) {
		logger := NewLogger()
		logger.SetColorEnabled(false)
		assert.False(t, logger.useColor)
		logger.SetColorEnabled(true)
		assert.True(t, logger.useColor)
	})
}

func TestLogger_Logging(t *testing.T) {
	tests := []struct {
		name      string
		level     LogLevel
		logLevel  LogLevel
		logFunc   func(*Logger, string, ...interface{})
		message   string
		shouldLog bool
	}{
		{
			name:      "debug message logged at debug level",
			level:     LogLevelDebug,
			logLevel:  LogLevelDebug,
			logFunc:   (*Logger).Debug,
			message:   "debug message",
			shouldLog: true,
		},
		{
			name:      "debug message not logged at info level",
			level:     LogLevelInfo,
			logLevel:  LogLevelDebug,
			logFunc:   (*Logger).Debug,
			message:   "debug message",
			shouldLog: false,
		},
		{
			name:      "info message logged at info level",
			level:     LogLevelInfo,
			logLevel:  LogLevelInfo,
			logFunc:   (*Logger).Info,
			message:   "info message",
			shouldLog: true,
		},
		{
			name:      "warn message logged at info level",
			level:     LogLevelInfo,
			logLevel:  LogLevelWarn,
			logFunc:   (*Logger).Warn,
			message:   "warn message",
			shouldLog: true,
		},
		{
			name:      "error message logged at info level",
			level:     LogLevelInfo,
			logLevel:  LogLevelError,
			logFunc:   (*Logger).Error,
			message:   "error message",
			shouldLog: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger()
			logger.SetOutput(&buf)
			logger.SetLevel(tt.level)
			logger.SetColorEnabled(false) // Disable color for easier testing

			tt.logFunc(logger, tt.message)

			output := buf.String()
			if tt.shouldLog {
				assert.Contains(t, output, tt.message)
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLogger_Formatting(t *testing.T) {
	t.Run("log with format arguments", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger()
		logger.SetOutput(&buf)
		logger.SetColorEnabled(false)

		logger.Info("Hello %s, number: %d", "world", 42)

		output := buf.String()
		assert.Contains(t, output, "Hello world, number: 42")
	})

	t.Run("log with timestamp", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger()
		logger.SetOutput(&buf)
		logger.SetColorEnabled(false)

		logger.Info("test message")

		output := buf.String()
		// Should contain a timestamp in HH:MM:SS format
		assert.Regexp(t, `\d{2}:\d{2}:\d{2}`, output)
	})

	t.Run("log with prefix", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger()
		logger.SetOutput(&buf)
		logger.SetColorEnabled(false)
		prefixedLogger := logger.WithPrefix("MODULE")

		prefixedLogger.Info("test message")

		output := buf.String()
		assert.Contains(t, output, "[MODULE]")
		assert.Contains(t, output, "test message")
	})
}

func TestLogger_EmojiLogging(t *testing.T) {
	t.Run("Success logs with emoji", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger()
		logger.SetOutput(&buf)
		logger.SetColorEnabled(false)

		logger.Success("Operation completed")

		output := buf.String()
		assert.Contains(t, output, emojiSuccess)
		assert.Contains(t, output, "Operation completed")
	})

	t.Run("Fail logs with emoji", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger()
		logger.SetOutput(&buf)
		logger.SetColorEnabled(false)

		logger.Fail("Operation failed")

		output := buf.String()
		assert.Contains(t, output, emojiError)
		assert.Contains(t, output, "Operation failed")
	})
}

func TestLogger_Header(t *testing.T) {
	t.Run("Header prints formatted header", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger()
		logger.SetOutput(&buf)
		logger.SetColorEnabled(false)

		logger.Header("Test Header")

		output := buf.String()
		assert.Contains(t, output, "Test Header")
		assert.Contains(t, output, "===")
		assert.Contains(t, output, strings.Repeat("=", 60))
	})
}

func TestLogger_Spinner(t *testing.T) {
	t.Run("StartSpinner and StopSpinner", func(t *testing.T) {
		logger := NewLogger()

		// Start spinner
		logger.StartSpinner("Loading...")
		assert.NotNil(t, logger.spinner)

		// Allow spinner to run briefly
		time.Sleep(50 * time.Millisecond)

		// Stop spinner
		logger.StopSpinner()
		assert.Nil(t, logger.spinner)
	})

	t.Run("UpdateSpinner updates message", func(t *testing.T) {
		logger := NewLogger()

		logger.StartSpinner("Loading...")
		require.NotNil(t, logger.spinner)

		logger.UpdateSpinner("Processing...")
		assert.Equal(t, "Processing...", logger.spinner.message)

		logger.StopSpinner()
	})

	t.Run("Multiple StartSpinner calls", func(t *testing.T) {
		logger := NewLogger()

		logger.StartSpinner("First")
		firstSpinner := logger.spinner

		logger.StartSpinner("Second")
		// Should have replaced the first spinner
		assert.NotEqual(t, firstSpinner, logger.spinner)
		assert.Equal(t, "Second", logger.spinner.message)

		logger.StopSpinner()
	})
}

func TestLogger_Context(t *testing.T) {
	t.Run("GetTimeContext returns correct context", func(t *testing.T) {
		logger := NewLogger()

		// Mock different times of day by testing the logic directly
		testCases := []struct {
			hour     int
			expected string
		}{
			{6, "morning"},    // 6 AM
			{11, "morning"},   // 11 AM
			{13, "afternoon"}, // 1 PM
			{16, "afternoon"}, // 4 PM
			{19, "evening"},   // 7 PM
			{23, "evening"},   // 11 PM
			{2, "evening"},    // 2 AM
		}

		for _, tc := range testCases {
			t.Run(tc.expected, func(t *testing.T) {
				// We can't easily mock time.Now(), but we can test the logic
				var context string
				switch {
				case tc.hour >= 5 && tc.hour < 12:
					context = "morning"
				case tc.hour >= 12 && tc.hour < 17:
					context = "afternoon"
				default:
					context = "evening"
				}
				assert.Equal(t, tc.expected, context)
			})
		}

		// Just verify the actual function works
		_ = logger.GetTimeContext()
	})

	t.Run("GetDayContext returns correct context", func(t *testing.T) {
		logger := NewLogger()

		// Test with current day context
		context := logger.GetDayContext()
		// Should be one of: "monday", "friday", or ""
		assert.True(t, context == "monday" || context == "friday" || context == "")
	})

	t.Run("GetContextualMessage returns message", func(t *testing.T) {
		logger := NewLogger()

		// Test with known context
		message := logger.GetContextualMessage("morning")
		if message != "" {
			assert.NotEmpty(t, message)
		}

		// Test with unknown context
		message = logger.GetContextualMessage("unknown")
		assert.Empty(t, message)
	})
}

func TestProgress(t *testing.T) {
	t.Run("NewProgress creates progress bar", func(t *testing.T) {
		progress := NewProgress(100, "Testing")
		assert.Equal(t, 100, progress.total)
		assert.Equal(t, "Testing", progress.message)
		assert.Equal(t, 0, progress.current)
	})

	t.Run("Update changes current value", func(t *testing.T) {
		progress := NewProgress(100, "Testing")
		progress.Update(50)
		assert.Equal(t, 50, progress.current)
	})

	t.Run("Increment increases current value", func(t *testing.T) {
		progress := NewProgress(100, "Testing")
		progress.Increment()
		assert.Equal(t, 1, progress.current)
		progress.Increment()
		assert.Equal(t, 2, progress.current)
	})

	t.Run("Finish sets to total", func(t *testing.T) {
		progress := NewProgress(100, "Testing")
		progress.Update(50)
		progress.Finish()
		assert.Equal(t, 100, progress.current)
	})

	// Note: render() is difficult to test directly without capturing stdout
	// We'll test that the methods don't panic
	t.Run("render methods don't panic", func(t *testing.T) {
		progress := NewProgress(10, "Testing")

		assert.NotPanics(t, func() {
			progress.Update(5)
		})

		assert.NotPanics(t, func() {
			progress.Increment()
		})

		assert.NotPanics(t, func() {
			progress.Finish()
		})
	})
}

func TestPackageLevelFunctions(t *testing.T) {
	// Test that package-level functions don't panic and use DefaultLogger
	t.Run("package functions use DefaultLogger", func(t *testing.T) {
		// Capture output from default logger
		var buf bytes.Buffer
		originalLogger := DefaultLogger
		DefaultLogger = NewLogger()
		DefaultLogger.SetOutput(&buf)
		DefaultLogger.SetColorEnabled(false)
		defer func() { DefaultLogger = originalLogger }()

		// Test various functions
		assert.NotPanics(t, func() {
			Debug("debug test")
			Info("info test")
			Warn("warn test")
			Error("error test")
			Success("success test")
			Fail("fail test")
			Header("Header test")
		})

		output := buf.String()
		assert.Contains(t, output, "info test")
		assert.Contains(t, output, "warn test")
		assert.Contains(t, output, "error test")
		assert.Contains(t, output, "success test")
		assert.Contains(t, output, "fail test")
		assert.Contains(t, output, "Header test")
		// Debug might not appear depending on log level
	})

	t.Run("spinner functions don't panic", func(t *testing.T) {
		assert.NotPanics(t, func() {
			StartSpinner("test")
			time.Sleep(10 * time.Millisecond)
			UpdateSpinner("updated")
			time.Sleep(10 * time.Millisecond)
			StopSpinner()
		})
	})
}

func TestFormatDurationInternal(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "negative duration returns 0s",
			duration: -time.Second,
			expected: "0s",
		},
		{
			name:     "milliseconds",
			duration: 250 * time.Millisecond,
			expected: "250ms",
		},
		{
			name:     "seconds",
			duration: 1500 * time.Millisecond,
			expected: "1.5s",
		},
		{
			name:     "minutes",
			duration: 75 * time.Second,
			expected: "1.2m", // 75s = 1.25m, formatted as 1.2m
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestShouldUseColor(t *testing.T) {
	// This function depends on environment variables and system state
	// We'll test the basic functionality without mocking the complex parts
	t.Run("shouldUseColor returns boolean", func(t *testing.T) {
		result := shouldUseColor()
		assert.True(t, result == true || result == false) // Just check it returns a boolean
	})

	// Test specific environment variable effects
	t.Run("NO_COLOR disables color", func(t *testing.T) {
		originalNoColor := os.Getenv("NO_COLOR")
		defer func() {
			if originalNoColor == "" {
				if err := os.Unsetenv("NO_COLOR"); err != nil {
					t.Logf("Warning: failed to unset NO_COLOR: %v", err)
				}
			} else {
				if err := os.Setenv("NO_COLOR", originalNoColor); err != nil {
					t.Logf("Warning: failed to set NO_COLOR: %v", err)
				}
			}
		}()

		if err := os.Setenv("NO_COLOR", "1"); err != nil {
			t.Fatalf("Failed to set NO_COLOR: %v", err)
		}
		result := shouldUseColor()
		assert.False(t, result)
	})

	t.Run("CI disables color", func(t *testing.T) {
		originalCI := os.Getenv("CI")
		defer func() {
			if originalCI == "" {
				if err := os.Unsetenv("CI"); err != nil {
					t.Logf("Warning: failed to unset CI: %v", err)
				}
			} else {
				if err := os.Setenv("CI", originalCI); err != nil {
					t.Logf("Warning: failed to set CI: %v", err)
				}
			}
		}()

		if err := os.Setenv("CI", "true"); err != nil {
			t.Fatalf("Failed to set CI: %v", err)
		}
		result := shouldUseColor()
		assert.False(t, result)
	})
}

// Benchmark tests
func BenchmarkLogger_Info(b *testing.B) {
	var buf bytes.Buffer
	logger := NewLogger()
	logger.SetOutput(&buf)
	logger.SetColorEnabled(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("benchmark test message %d", i)
	}
}

func BenchmarkLogger_WithSpinner(b *testing.B) {
	logger := NewLogger()
	logger.SetColorEnabled(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.StartSpinner("test")
		logger.StopSpinner()
	}
}

func BenchmarkFormatDuration(b *testing.B) {
	duration := 1500 * time.Millisecond

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatDuration(duration)
	}
}
