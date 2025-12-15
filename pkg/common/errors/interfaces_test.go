package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSeverity_String verifies all severity values have string representation
func TestSeverity_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityDebug, "DEBUG"},
		{SeverityInfo, "INFO"},
		{SeverityWarning, "WARNING"},
		{SeverityError, "ERROR"},
		{SeverityCritical, "CRITICAL"},
		{SeverityFatal, "FATAL"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			result := tt.severity.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestSeverity_UnknownValue verifies invalid severity has fallback
func TestSeverity_UnknownValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity Severity
	}{
		{"negative value", Severity(-1)},
		{"beyond fatal", Severity(100)},
		{"arbitrary value", Severity(42)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := tt.severity.String()
			// Should return "UNKNOWN" or similar fallback
			assert.NotEmpty(t, result, "Unknown severity should have some string representation")
		})
	}
}

// TestSeverity_MarshalText verifies MarshalText produces correct bytes
func TestSeverity_MarshalText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityDebug, "DEBUG"},
		{SeverityInfo, "INFO"},
		{SeverityWarning, "WARNING"},
		{SeverityError, "ERROR"},
		{SeverityCritical, "CRITICAL"},
		{SeverityFatal, "FATAL"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			data, err := tt.severity.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))
		})
	}
}

// TestSeverity_UnmarshalText verifies UnmarshalText parses correctly
func TestSeverity_UnmarshalText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected Severity
	}{
		{"DEBUG", SeverityDebug},
		{"INFO", SeverityInfo},
		{"WARNING", SeverityWarning},
		{"ERROR", SeverityError},
		{"CRITICAL", SeverityCritical},
		{"FATAL", SeverityFatal},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			var sev Severity
			err := sev.UnmarshalText([]byte(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, sev)
		})
	}
}

// TestSeverity_UnmarshalTextCaseInsensitive verifies case-insensitive parsing
func TestSeverity_UnmarshalTextCaseInsensitive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected Severity
	}{
		{"debug", SeverityDebug},
		{"Debug", SeverityDebug},
		{"info", SeverityInfo},
		{"Info", SeverityInfo},
		{"warning", SeverityWarning},
		{"Warning", SeverityWarning},
		{"error", SeverityError},
		{"Error", SeverityError},
		{"critical", SeverityCritical},
		{"Critical", SeverityCritical},
		{"fatal", SeverityFatal},
		{"Fatal", SeverityFatal},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			var sev Severity
			err := sev.UnmarshalText([]byte(tt.input))
			require.NoError(t, err)
			assert.Equal(t, tt.expected, sev)
		})
	}
}

// TestSeverity_UnmarshalInvalid verifies invalid text returns error
func TestSeverity_UnmarshalInvalid(t *testing.T) {
	t.Parallel()

	invalidInputs := []string{
		"INVALID",
		"unknown",
		"",
		"123",
		"WARN", // Not WARNING
		"ERR",  // Not ERROR
		"CRIT", // Not CRITICAL
	}

	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			var sev Severity
			err := sev.UnmarshalText([]byte(input))
			assert.Error(t, err, "Invalid input %q should return error", input)
		})
	}
}

// TestSeverity_RoundTrip verifies Marshal then Unmarshal preserves value
func TestSeverity_RoundTrip(t *testing.T) {
	t.Parallel()

	severities := []Severity{
		SeverityDebug,
		SeverityInfo,
		SeverityWarning,
		SeverityError,
		SeverityCritical,
		SeverityFatal,
	}

	for _, original := range severities {
		t.Run(original.String(), func(t *testing.T) {
			t.Parallel()

			// Marshal
			data, err := original.MarshalText()
			require.NoError(t, err)

			// Unmarshal
			var result Severity
			err = result.UnmarshalText(data)
			require.NoError(t, err)

			// Verify
			assert.Equal(t, original, result, "Round-trip should preserve value")
		})
	}
}

// TestSeverity_Comparison verifies severity comparison
func TestSeverity_Comparison(t *testing.T) {
	t.Parallel()

	// Severities should be orderable
	assert.Less(t, SeverityDebug, SeverityInfo)
	assert.Less(t, SeverityInfo, SeverityWarning)
	assert.Less(t, SeverityWarning, SeverityError)
	assert.Less(t, SeverityError, SeverityCritical)
	assert.Less(t, SeverityCritical, SeverityFatal)
}

// TestSeverity_Constants verifies severity constants have expected values
func TestSeverity_Constants(t *testing.T) {
	t.Parallel()

	// Verify the severity levels are distinct
	severities := []Severity{
		SeverityDebug,
		SeverityInfo,
		SeverityWarning,
		SeverityError,
		SeverityCritical,
		SeverityFatal,
	}

	// Check all are distinct
	seen := make(map[Severity]bool)
	for _, sev := range severities {
		assert.False(t, seen[sev], "Severity %v should be unique", sev)
		seen[sev] = true
	}
}

// TestErrorCode_String verifies ErrorCode string conversion
func TestErrorCode_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		code     ErrorCode
		expected string
	}{
		{ErrUnknown, "UNKNOWN"},
		{ErrBuildFailed, "BUILD_FAILED"},
		{ErrTestFailed, "TEST_FAILED"},
		{ErrorCode("CUSTOM"), "CUSTOM"},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, string(tt.code))
		})
	}
}

// TestErrorContext_ZeroValue verifies ErrorContext zero value is usable
func TestErrorContext_ZeroValue(t *testing.T) {
	t.Parallel()

	var ctx ErrorContext

	assert.Empty(t, ctx.Operation)
	assert.Empty(t, ctx.Resource)
	assert.Empty(t, ctx.User)
	assert.Empty(t, ctx.RequestID)
	assert.Empty(t, ctx.Environment)
	assert.Empty(t, ctx.StackTrace)
	assert.True(t, ctx.Timestamp.IsZero())
	assert.Nil(t, ctx.Fields)
}

// TestErrorContext_WithFields verifies Fields map initialization
func TestErrorContext_WithFields(t *testing.T) {
	t.Parallel()

	ctx := ErrorContext{
		Fields: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	assert.Equal(t, "value1", ctx.Fields["key1"])
	assert.Equal(t, 42, ctx.Fields["key2"])
}

// TestFormatOptions_ZeroValue verifies FormatOptions zero value
func TestFormatOptions_ZeroValue(t *testing.T) {
	t.Parallel()

	var opts FormatOptions

	// All booleans should be false
	assert.False(t, opts.IncludeStack)
	assert.False(t, opts.IncludeContext)
	assert.False(t, opts.IncludeCause)
	assert.False(t, opts.IncludeTimestamp)
	assert.False(t, opts.IncludeFields)
	assert.False(t, opts.UseColor)
	assert.False(t, opts.CompactMode)

	// Numbers should be zero
	assert.Equal(t, 0, opts.IndentLevel)
	assert.Equal(t, 0, opts.MaxDepth)

	// Strings should be empty
	assert.Empty(t, opts.TimeFormat)
	assert.Empty(t, opts.FieldsSeparator)
}

// TestErrorDefinition_Fields verifies ErrorDefinition structure
func TestErrorDefinition_Fields(t *testing.T) {
	t.Parallel()

	def := ErrorDefinition{
		Code:        ErrBuildFailed,
		Description: "Build process failed",
		Severity:    SeverityError,
		Category:    "build",
		Retryable:   true,
		Tags:        []string{"build", "failure"},
	}

	assert.Equal(t, ErrBuildFailed, def.Code)
	assert.Equal(t, "Build process failed", def.Description)
	assert.Equal(t, SeverityError, def.Severity)
	assert.Equal(t, "build", def.Category)
	assert.True(t, def.Retryable)
	assert.Contains(t, def.Tags, "build")
	assert.Contains(t, def.Tags, "failure")
}

// TestErrorStat_Fields verifies ErrorStat structure
func TestErrorStat_Fields(t *testing.T) {
	t.Parallel()

	stat := ErrorStat{
		Code:        ErrTimeout,
		Count:       42,
		AverageRate: 3.14,
	}

	assert.Equal(t, ErrTimeout, stat.Code)
	assert.Equal(t, int64(42), stat.Count)
	assert.InDelta(t, 3.14, stat.AverageRate, 0.001)
}

// TestBackoffConfig_ZeroValue verifies BackoffConfig zero value
func TestBackoffConfig_ZeroValue(t *testing.T) {
	t.Parallel()

	var config BackoffConfig

	assert.Zero(t, config.InitialDelay)
	assert.Zero(t, config.MaxDelay)
	assert.Zero(t, config.Multiplier)
	assert.Zero(t, config.MaxRetries)
	assert.Nil(t, config.RetryIf)
}

// TestBackoffConfig_CustomValues verifies BackoffConfig custom values
func TestBackoffConfig_CustomValues(t *testing.T) {
	t.Parallel()

	called := false
	predicate := func(_ error) bool {
		called = true
		return true
	}

	config := BackoffConfig{
		InitialDelay: 100,
		MaxDelay:     1000,
		Multiplier:   2.5,
		MaxRetries:   5,
		RetryIf:      predicate,
	}

	assert.Equal(t, 100, int(config.InitialDelay))
	assert.Equal(t, 1000, int(config.MaxDelay))
	assert.InDelta(t, 2.5, config.Multiplier, 0.001)
	assert.Equal(t, 5, config.MaxRetries)

	// Test predicate works
	assert.NotNil(t, config.RetryIf)
	config.RetryIf(nil)
	assert.True(t, called)
}
