package log

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStructured tests the Structured() function
func TestStructured(t *testing.T) {
	// Initialize the logger first
	Initialize()

	// Get the structured logger
	structured := Structured()
	require.NotNil(t, structured, "Structured logger should not be nil")

	// Verify it works
	structured.SetLevel(LevelInfo)
	assert.Equal(t, LevelInfo, structured.GetLevel())
}

// TestStartStopSpinner tests spinner functionality
func TestStartStopSpinner(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// Initialize
	Initialize()

	// Start spinner
	StartSpinner("Loading...")

	// Update spinner
	UpdateSpinner("Still loading...")

	// Stop spinner
	StopSpinner()

	// Restore stdout
	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	// Test passes if no panic
}

// TestWithField tests WithField function
func TestWithField(t *testing.T) {
	Initialize()

	logger := WithField("key", "value")
	require.NotNil(t, logger, "WithField should return a logger")

	// Use the logger to ensure it works
	logger.Info("test message")
}

// TestWithFields tests WithFields function
func TestWithFields(t *testing.T) {
	Initialize()

	fields := Fields{
		"key1": "value1",
		"key2": "value2",
	}

	logger := WithFields(fields)
	require.NotNil(t, logger, "WithFields should return a logger")

	// Use the logger to ensure it works
	logger.Info("test message")
}

// TestWithPrefix tests WithPrefix function
func TestWithPrefix(t *testing.T) {
	Initialize()

	logger := WithPrefix("[TEST]")
	require.NotNil(t, logger, "WithPrefix should return a logger")

	// Use the logger to ensure it works
	logger.Info("test message")
}

// TestShouldUseColor_EdgeCases tests color detection through initialization
func TestShouldUseColor_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		noColor string
		term    string
		ciEnv   string
	}{
		{
			name:    "NO_COLOR set",
			noColor: "1",
			term:    "xterm-256color",
			ciEnv:   "",
		},
		{
			name:    "dumb terminal",
			noColor: "",
			term:    "dumb",
			ciEnv:   "",
		},
		{
			name:    "CI environment",
			noColor: "",
			term:    "xterm",
			ciEnv:   "true",
		},
		{
			name:    "normal terminal",
			noColor: "",
			term:    "xterm-256color",
			ciEnv:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			if tt.noColor != "" {
				t.Setenv("NO_COLOR", tt.noColor)
			}
			if tt.term != "" {
				t.Setenv("TERM", tt.term)
			}
			if tt.ciEnv != "" {
				t.Setenv("CI", tt.ciEnv)
			}

			// Create adapter which calls shouldUseColor internally
			adapter := NewCLIAdapter()
			require.NotNil(t, adapter)
		})
	}
}

// TestLogWithEmoji_EdgeCases tests emoji logging through exported methods
func TestLogWithEmoji_EdgeCases(t *testing.T) {
	// Capture output
	var buf bytes.Buffer

	// Initialize and redirect output
	Initialize()
	adapter := Default()
	adapter.SetOutput(&buf)

	// Test Success which uses logWithEmoji internally
	adapter.Success("test success")
	assert.Contains(t, buf.String(), "test success")

	buf.Reset()

	// Test Fail which uses logWithEmoji internally
	adapter.Fail("test fail")
	assert.Contains(t, buf.String(), "test fail")
}

// TestLog_EdgeCases tests log function through exported methods
func TestLog_EdgeCases(t *testing.T) {
	// Capture output
	var buf bytes.Buffer

	// Initialize and redirect output
	Initialize()
	adapter := Default()
	adapter.SetOutput(&buf)
	adapter.SetLevel(LevelDebug)

	// Test with empty format
	adapter.Info("")

	// Test with format and args
	adapter.Info("test %s %d", "message", 42)

	// Test passes if no panic
	output := buf.String()
	assert.Contains(t, output, "test message 42")
}

// TestInitialize_EdgeCases tests Initialize edge cases
func TestInitialize_EdgeCases(t *testing.T) {
	t.Run("verbose mode with color detection", func(t *testing.T) {
		// Set terminal and verbose
		t.Setenv("TERM", "xterm-256color")
		t.Setenv("VERBOSE", "true")

		// Initialize
		Initialize()

		// Verify initialized
		assert.NotNil(t, Default())
		assert.Equal(t, LevelDebug, GetLevel())
	})

	t.Run("debug mode", func(t *testing.T) {
		// Set DEBUG env
		t.Setenv("DEBUG", "true")

		// Initialize
		Initialize()

		// Verify initialized
		assert.NotNil(t, Default())
		assert.Equal(t, LevelDebug, GetLevel())
	})
}

// TestSpinner_MultipleCalls tests spinner with multiple calls
func TestSpinner_MultipleCalls(t *testing.T) {
	// Capture output
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	Initialize()

	// Start multiple spinners
	StartSpinner("First")
	StartSpinner("Second") // Should replace first

	// Update multiple times
	UpdateSpinner("Update 1")
	UpdateSpinner("Update 2")

	// Stop
	StopSpinner()

	// Try to stop again (should be safe)
	StopSpinner()

	// Restore stdout
	_ = w.Close() //nolint:errcheck // test cleanup
	os.Stdout = oldStdout
	_ = r.Close() //nolint:errcheck // test cleanup

	// Test passes if no panic
}

// TestWithMethods_NilDefault tests With* methods when Default is nil
func TestWithMethods_NilDefault(t *testing.T) {
	// Reset manager to test nil case
	manager.mu.Lock()
	oldCLI := manager.cli
	manager.cli = nil
	manager.mu.Unlock()

	// Restore after test
	defer func() {
		manager.mu.Lock()
		manager.cli = oldCLI
		manager.mu.Unlock()
	}()

	// These should return nil without panicking
	assert.Nil(t, WithField("key", "value"))
	assert.Nil(t, WithFields(Fields{"key": "value"}))
	assert.Nil(t, WithPrefix("[TEST]"))

	// Spinner operations should not panic
	StartSpinner("test")
	UpdateSpinner("test")
	StopSpinner()
}
