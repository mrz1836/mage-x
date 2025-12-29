package errors

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConsoleChannel(t *testing.T) {
	channel := NewConsoleChannel("test-channel")

	require.NotNil(t, channel)
	assert.Equal(t, "test-channel", channel.Name())
	assert.True(t, channel.IsEnabled())
}

func TestConsoleChannel_Name(t *testing.T) {
	channel := NewConsoleChannel("my-channel")
	assert.Equal(t, "my-channel", channel.Name())
}

func TestConsoleChannel_EnabledState(t *testing.T) {
	channel := NewConsoleChannel("test")

	// Default enabled
	assert.True(t, channel.IsEnabled())

	// Disable
	channel.SetEnabled(false)
	assert.False(t, channel.IsEnabled())

	// Re-enable
	channel.SetEnabled(true)
	assert.True(t, channel.IsEnabled())
}

func TestConsoleChannel_Send_WhenDisabled(t *testing.T) {
	channel := NewConsoleChannel("test")
	channel.SetEnabled(false)

	notification := &ErrorNotification{
		Error:     NewMageError("test error"),
		Timestamp: time.Now(),
	}

	err := channel.Send(context.Background(), notification)
	assert.NoError(t, err)
}

func TestConsoleChannel_Send_NilNotification(t *testing.T) {
	channel := NewConsoleChannel("test")

	err := channel.Send(context.Background(), nil)
	assert.NoError(t, err)
}

func TestConsoleChannel_Send_ValidNotification(t *testing.T) {
	channel := NewConsoleChannel("test")

	notification := &ErrorNotification{
		Error:       NewMageError("test error message"),
		Timestamp:   time.Now(),
		Environment: "test",
		Hostname:    "localhost",
		Service:     "test-service",
	}

	// Should not error - logging happens internally
	err := channel.Send(context.Background(), notification)
	assert.NoError(t, err)
}

func TestFormatErrorContext(t *testing.T) {
	t.Run("nil context", func(t *testing.T) {
		result := formatErrorContext(nil)
		assert.Equal(t, "No additional context", result)
	})

	t.Run("empty fields", func(t *testing.T) {
		ctx := &ErrorContext{Fields: map[string]interface{}{}}
		result := formatErrorContext(ctx)
		assert.Equal(t, "No additional context", result)
	})

	t.Run("with fields", func(t *testing.T) {
		ctx := &ErrorContext{
			Fields: map[string]interface{}{
				"key": "value",
			},
		}
		result := formatErrorContext(ctx)
		assert.Contains(t, result, "key: value")
	})
}

func TestEscapeJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"no escaping", "hello world", "hello world"},
		{"backslash", `a\b`, `a\\b`},
		{"quote", `a"b`, `a\"b`},
		{"newline", "a\nb", `a\nb`},
		{"carriage return", "a\rb", `a\rb`},
		{"tab", "a\tb", `a\tb`},
		{"combined", "a\"\n\\b", `a\"\n\\b`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := escapeJSON(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
