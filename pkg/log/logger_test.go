package log

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestLevel_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level Level
		want  string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{LevelFatal, "FATAL"},
		{Level(100), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := tt.level.String(); got != tt.want {
				t.Errorf("Level.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"WARN", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"fatal", LevelFatal},
		{"unknown", LevelInfo}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := ParseLevel(tt.input); got != tt.want {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestCLIAdapter_Logging(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	adapter := NewCLIAdapter()
	adapter.SetOutput(&buf)
	adapter.SetColorEnabled(false)
	adapter.SetLevel(LevelDebug)

	adapter.Debug("debug message %d", 1)
	adapter.Info("info message %s", "test")
	adapter.Warn("warn message")
	adapter.Error("error message")

	output := buf.String()

	if !strings.Contains(output, "[DEBUG] debug message 1") {
		t.Errorf("Expected debug message, got: %s", output)
	}
	if !strings.Contains(output, "[INFO] info message test") {
		t.Errorf("Expected info message, got: %s", output)
	}
	if !strings.Contains(output, "[WARN] warn message") {
		t.Errorf("Expected warn message, got: %s", output)
	}
	if !strings.Contains(output, "[ERROR] error message") {
		t.Errorf("Expected error message, got: %s", output)
	}
}

func TestCLIAdapter_LevelFiltering(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	adapter := NewCLIAdapter()
	adapter.SetOutput(&buf)
	adapter.SetColorEnabled(false)
	adapter.SetLevel(LevelWarn)

	adapter.Debug("should not appear")
	adapter.Info("should not appear")
	adapter.Warn("should appear")
	adapter.Error("should appear")

	output := buf.String()

	if strings.Contains(output, "should not appear") {
		t.Errorf("Debug/Info messages should be filtered, got: %s", output)
	}
	if !strings.Contains(output, "should appear") {
		t.Errorf("Warn/Error messages should appear, got: %s", output)
	}
}

func TestCLIAdapter_SuccessAndFail(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	adapter := NewCLIAdapter()
	adapter.SetOutput(&buf)
	adapter.SetColorEnabled(false)

	adapter.Success("success message")
	adapter.Fail("fail message")

	output := buf.String()

	if !strings.Contains(output, "✅ success message") {
		t.Errorf("Expected success emoji, got: %s", output)
	}
	if !strings.Contains(output, "❌ fail message") {
		t.Errorf("Expected fail emoji, got: %s", output)
	}
}

func TestCLIAdapter_Header(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	adapter := NewCLIAdapter()
	adapter.SetOutput(&buf)
	adapter.SetColorEnabled(false)

	adapter.Header("Test Header")

	output := buf.String()

	if !strings.Contains(output, "=== Test Header ===") {
		t.Errorf("Expected header, got: %s", output)
	}
}

func TestCLIAdapter_WithPrefix(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	adapter := NewCLIAdapter()
	adapter.SetOutput(&buf)
	adapter.SetColorEnabled(false)

	prefixed := adapter.WithPrefix("PREFIX")
	cliPrefixed, ok := prefixed.(*CLIAdapter)
	if !ok {
		t.Fatal("WithPrefix should return *CLIAdapter")
	}
	cliPrefixed.SetOutput(&buf)

	cliPrefixed.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "[PREFIX] test message") {
		t.Errorf("Expected prefix, got: %s", output)
	}
}

func TestCLIAdapter_WithField(t *testing.T) {
	t.Parallel()

	adapter := NewCLIAdapter()
	withField := adapter.WithField("key", "value")

	if withField == nil {
		t.Error("WithField should not return nil")
	}
}

func TestStructuredAdapter_Logging(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	adapter := NewStructuredAdapter()
	adapter.SetOutput(&buf)
	adapter.SetLevel(LevelDebug)

	adapter.Debug("debug message")
	adapter.Info("info message")
	adapter.Warn("warn message")
	adapter.Error("error message")

	output := buf.String()

	if !strings.Contains(output, "[DEBUG] debug message") {
		t.Errorf("Expected debug message, got: %s", output)
	}
	if !strings.Contains(output, "[INFO] info message") {
		t.Errorf("Expected info message, got: %s", output)
	}
}

func TestStructuredAdapter_ContextLogging(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	adapter := NewStructuredAdapter()
	adapter.SetOutput(&buf)

	// The adapter looks for string keys "request_id", "requestId", etc.
	// We use a string key here to test the context extraction
	//nolint:staticcheck // SA1029: Using string key intentionally to match getRequestIDFromContext
	ctx := context.WithValue(context.Background(), "request_id", "req-123")
	adapter.InfoContext(ctx, "context message")

	output := buf.String()

	if !strings.Contains(output, "[req:req-123]") {
		t.Errorf("Expected request ID in output, got: %s", output)
	}
	if !strings.Contains(output, "context message") {
		t.Errorf("Expected message, got: %s", output)
	}
}

func TestStructuredAdapter_WithFields(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	adapter := NewStructuredAdapter()
	adapter.SetOutput(&buf)

	withFields := adapter.WithFields(Fields{
		"user": "john",
		"id":   123,
	})

	structuredWithFields, ok := withFields.(*StructuredAdapter)
	if !ok {
		t.Fatal("WithFields should return *StructuredAdapter")
	}
	structuredWithFields.SetOutput(&buf)

	structuredWithFields.Info("test message")

	output := buf.String()

	if !strings.Contains(output, "user=john") {
		t.Errorf("Expected user field, got: %s", output)
	}
	if !strings.Contains(output, "id=123") {
		t.Errorf("Expected id field, got: %s", output)
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	// Not parallel: test modifies global default logger

	// Save and restore default logger
	oldDefault := Default()
	defer SetDefault(oldDefault)

	var buf bytes.Buffer
	adapter := NewCLIAdapter()
	adapter.SetOutput(&buf)
	adapter.SetColorEnabled(false)
	adapter.SetLevel(LevelDebug)
	SetDefault(adapter)

	Debug("debug %d", 1)
	Info("info %s", "test")
	Warn("warn")
	Error("error")
	Success("success")
	Fail("fail")
	Header("test header")

	output := buf.String()

	if !strings.Contains(output, "debug 1") {
		t.Errorf("Expected debug message, got: %s", output)
	}
	if !strings.Contains(output, "info test") {
		t.Errorf("Expected info message, got: %s", output)
	}
}

func TestGlobalLevelSetting(t *testing.T) {
	// Not parallel: test modifies global log level

	// Save and restore
	oldLevel := GetLevel()
	defer SetLevel(oldLevel)

	SetLevel(LevelError)
	if GetLevel() != LevelError {
		t.Errorf("GetLevel() = %v, want %v", GetLevel(), LevelError)
	}
}
