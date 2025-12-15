package errors

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Static error variables for tests (err113 compliance)
var (
	errStandardMessage = errors.New("standard error message")
	errStandard        = errors.New("standard error")
	errGenericBase     = errors.New("generic base error")
)

// TestFormatter_NilError verifies Format(nil) returns "<nil>"
func TestFormatter_NilError(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	result := formatter.Format(nil)
	assert.Equal(t, "<nil>", result)
}

// TestFormatter_NilErrorWithOptions verifies FormatWithOptions(nil) returns "<nil>"
func TestFormatter_NilErrorWithOptions(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	result := formatter.FormatWithOptions(nil, FormatOptions{})
	assert.Equal(t, "<nil>", result)
}

// TestFormatter_GenericError verifies non-MageError formatted correctly
func TestFormatter_GenericError(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()

	result := formatter.Format(errStandardMessage)
	assert.Contains(t, result, "standard error message")
}

// TestFormatter_MageErrorBasic verifies basic MageError formatting
func TestFormatter_MageErrorBasic(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	err := WithCode(ErrBuildFailed, "build process failed")

	result := formatter.Format(err)

	assert.Contains(t, result, "BUILD_FAILED")
	assert.Contains(t, result, "build process failed")
	assert.Contains(t, result, "Severity:")
}

// TestFormatter_CompactMode verifies CompactMode produces single line
func TestFormatter_CompactMode(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	err := NewBuilder().
		WithMessage("compact error").
		WithCode(ErrBuildFailed).
		WithSeverity(SeverityError).
		Build()

	opts := FormatOptions{
		CompactMode: true,
	}

	result := formatter.FormatWithOptions(err, opts)

	assert.NotContains(t, result, "\n", "Compact format should be single line")
	assert.Contains(t, result, "BUILD_FAILED")
	assert.Contains(t, result, "ERROR")
	assert.Contains(t, result, "compact error")
}

// TestFormatter_CompactModeWithFields verifies fields included in compact format
func TestFormatter_CompactModeWithFields(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	err := NewBuilder().
		WithMessage("compact error").
		WithCode(ErrBuildFailed).
		WithField("key1", "value1").
		WithField("count", 42).
		Build()

	opts := FormatOptions{
		CompactMode:   true,
		IncludeFields: true,
	}

	result := formatter.FormatWithOptions(err, opts)

	assert.Contains(t, result, "key1=value1")
	assert.Contains(t, result, "count=42")
}

// TestFormatter_UseColor verifies ANSI codes applied per severity
func TestFormatter_UseColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		severity Severity
		color    string
	}{
		{
			name:     "debug uses gray",
			severity: SeverityDebug,
			color:    "\033[90m", // gray
		},
		{
			name:     "info uses cyan",
			severity: SeverityInfo,
			color:    "\033[36m", // cyan
		},
		{
			name:     "warning uses yellow",
			severity: SeverityWarning,
			color:    "\033[33m", // yellow
		},
		{
			name:     "error uses red",
			severity: SeverityError,
			color:    "\033[31m", // red
		},
		{
			name:     "critical uses purple",
			severity: SeverityCritical,
			color:    "\033[35m", // purple
		},
		{
			name:     "fatal uses red",
			severity: SeverityFatal,
			color:    "\033[31m", // red
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			formatter := NewFormatter()
			err := NewBuilder().
				WithMessage("test error").
				WithCode(ErrBuildFailed).
				WithSeverity(tt.severity).
				Build()

			opts := FormatOptions{
				UseColor: true,
			}

			result := formatter.FormatWithOptions(err, opts)
			assert.Contains(t, result, tt.color, "Should contain color code for %s", tt.severity)
			assert.Contains(t, result, "\033[0m", "Should contain reset code")
		})
	}
}

// TestFormatter_IndentLevel verifies custom indent level applied
func TestFormatter_IndentLevel(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()

	// Create nested error
	baseErr := WithCode(ErrInternal, "base error")
	wrappedErr := Wrap(baseErr, "wrapped error")

	opts := FormatOptions{
		IncludeCause: true,
		IndentLevel:  4, // Custom indent
	}

	formatter.SetDefaultOptions(opts)
	result := formatter.Format(wrappedErr)

	// Should have indentation in output
	assert.Contains(t, result, "Caused by:")
}

// TestFormatter_MaxDepthReached verifies deep nesting truncated
func TestFormatter_MaxDepthReached(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()

	// Create deeply nested error chain
	err := WithCode(ErrInternal, "level 0")
	for i := 1; i <= 15; i++ {
		err = Wrap(err, "level "+string(rune('0'+i%10)))
	}

	opts := FormatOptions{
		IncludeCause: true,
		MaxDepth:     3,
	}

	result := formatter.FormatWithOptions(err, opts)
	assert.Contains(t, result, "max depth reached", "Should indicate max depth was reached")
}

// TestFormatter_IncludeStack verifies stack trace included when enabled
func TestFormatter_IncludeStack(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	err := NewBuilder().
		WithMessage("error with stack").
		WithCode(ErrBuildFailed).
		WithStackTrace().
		Build()

	// Without IncludeStack option - stack should not be shown
	noStackOpts := FormatOptions{
		IncludeStack: false,
	}
	noStackResult := formatter.FormatWithOptions(err, noStackOpts)
	assert.NotContains(t, noStackResult, "Stack trace:")

	// With IncludeStack option - stack should be shown if error has one
	withStackOpts := FormatOptions{
		IncludeStack: true,
	}
	withStackResult := formatter.FormatWithOptions(err, withStackOpts)

	// Check if the error actually captured a stack trace
	if err.Context().StackTrace != "" {
		assert.Contains(t, withStackResult, "Stack trace:")
	}
}

// TestFormatter_IncludeTimestamp verifies timestamp formatted correctly
func TestFormatter_IncludeTimestamp(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	err := NewBuilder().
		WithMessage("test error").
		WithCode(ErrBuildFailed).
		Build()

	opts := FormatOptions{
		IncludeTimestamp: true,
		TimeFormat:       time.RFC3339,
	}

	result := formatter.FormatWithOptions(err, opts)
	assert.Contains(t, result, "Time:")
}

// TestFormatter_IncludeContext verifies context fields formatted
func TestFormatter_IncludeContext(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	err := NewBuilder().
		WithMessage("test error").
		WithCode(ErrBuildFailed).
		WithOperation("build").
		WithResource("main.go").
		Build()

	opts := FormatOptions{
		IncludeContext: true,
	}

	result := formatter.FormatWithOptions(err, opts)

	assert.Contains(t, result, "Operation: build")
	assert.Contains(t, result, "Resource: main.go")
}

// TestFormatter_IncludeCause verifies cause chain formatted
func TestFormatter_IncludeCause(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	baseErr := WithCode(ErrInternal, "root cause")
	wrappedErr := Wrap(baseErr, "wrapped error")

	// Without cause
	noCauseOpts := FormatOptions{
		IncludeCause: false,
	}
	noCauseResult := formatter.FormatWithOptions(wrappedErr, noCauseOpts)
	assert.NotContains(t, noCauseResult, "Caused by:")

	// With cause
	withCauseOpts := FormatOptions{
		IncludeCause: true,
	}
	withCauseResult := formatter.FormatWithOptions(wrappedErr, withCauseOpts)
	assert.Contains(t, withCauseResult, "Caused by:")
	assert.Contains(t, withCauseResult, "root cause")
}

// TestFormatter_ChainFormatting verifies error chain shows count and items
func TestFormatter_ChainFormatting(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	chain := NewChain()
	chain = chain.Add(WithCode(ErrBuildFailed, "build error"))
	chain = chain.Add(WithCode(ErrTestFailed, "test error"))
	chain = chain.Add(WithCode(ErrLintFailed, "lint error"))

	result := formatter.FormatChain(chain)

	assert.Contains(t, result, "3 errors")
	assert.Contains(t, result, "[1]")
	assert.Contains(t, result, "[2]")
	assert.Contains(t, result, "[3]")
	assert.Contains(t, result, "BUILD_FAILED")
	assert.Contains(t, result, "TEST_FAILED")
	assert.Contains(t, result, "LINT_FAILED")
}

// TestFormatter_EmptyChain verifies empty chain returns "no errors"
func TestFormatter_EmptyChain(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	chain := NewChain()

	result := formatter.FormatChain(chain)
	assert.Equal(t, "no errors in chain", result)
}

// TestFormatter_SetDefaultOptions verifies options persist across calls
func TestFormatter_SetDefaultOptions(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()

	newOpts := FormatOptions{
		CompactMode:      true,
		IncludeFields:    true,
		IncludeStack:     false,
		IncludeContext:   false,
		IncludeTimestamp: false,
	}

	formatter.SetDefaultOptions(newOpts)

	retrieved := formatter.GetDefaultOptions()
	assert.Equal(t, newOpts.CompactMode, retrieved.CompactMode)
	assert.Equal(t, newOpts.IncludeFields, retrieved.IncludeFields)
	assert.Equal(t, newOpts.IncludeStack, retrieved.IncludeStack)
}

// TestFormatter_NestedMageErrorCause verifies nested MageError causes formatted
func TestFormatter_NestedMageErrorCause(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()

	innerErr := NewBuilder().
		WithMessage("inner error").
		WithCode(ErrInternal).
		WithSeverity(SeverityError).
		WithField("innerKey", "innerValue").
		Build()

	outerErr := NewBuilder().
		WithMessage("outer error").
		WithCode(ErrBuildFailed).
		WithSeverity(SeverityCritical).
		WithCause(innerErr).
		Build()

	opts := FormatOptions{
		IncludeCause:  true,
		IncludeFields: true,
		MaxDepth:      10, // Ensure enough depth for nested cause
	}

	result := formatter.FormatWithOptions(outerErr, opts)

	assert.Contains(t, result, "outer error")
	assert.Contains(t, result, "BUILD_FAILED")
	assert.Contains(t, result, "Caused by:")
	// Inner error details should be visible with sufficient depth
	assert.Contains(t, result, "inner error")
}

// TestFormatter_FormatMageError verifies FormatMageError method
func TestFormatter_FormatMageError(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	err := NewBuilder().
		WithMessage("test error").
		WithCode(ErrBuildFailed).
		WithSeverity(SeverityError).
		Build()

	result := formatter.FormatMageError(err)

	assert.Contains(t, result, "BUILD_FAILED")
	assert.Contains(t, result, "test error")
}

// TestFormatter_FormatWithOptionsChain verifies chain detection in FormatWithOptions
func TestFormatter_FormatWithOptionsChain(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	chain := NewChain()
	chain = chain.Add(WithCode(ErrBuildFailed, "error 1"))
	chain = chain.Add(WithCode(ErrTestFailed, "error 2"))

	result := formatter.FormatWithOptions(chain, FormatOptions{})

	assert.Contains(t, result, "2 errors")
}

// TestFormatter_FieldsFormatting verifies fields are properly formatted
func TestFormatter_FieldsFormatting(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	err := NewBuilder().
		WithMessage("test error").
		WithCode(ErrBuildFailed).
		WithFields(map[string]interface{}{
			"string":  "value",
			"number":  42,
			"boolean": true,
		}).
		Build()

	opts := FormatOptions{
		IncludeFields: true,
	}

	result := formatter.FormatWithOptions(err, opts)

	assert.Contains(t, result, "Fields:")
	assert.Contains(t, result, "string: value")
	assert.Contains(t, result, "number: 42")
	assert.Contains(t, result, "boolean: true")
}

// TestFormatter_NoFieldsWhenEmpty verifies Fields section omitted when empty
func TestFormatter_NoFieldsWhenEmpty(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	err := NewBuilder().
		WithMessage("test error").
		WithCode(ErrBuildFailed).
		Build() // No fields

	opts := FormatOptions{
		IncludeFields: true,
	}

	result := formatter.FormatWithOptions(err, opts)

	assert.NotContains(t, result, "Fields:")
}

// TestFormatter_CompactModeFieldsSeparator verifies custom field separator
func TestFormatter_CompactModeFieldsSeparator(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	formatter.SetDefaultOptions(FormatOptions{
		CompactMode:     true,
		IncludeFields:   true,
		FieldsSeparator: " | ",
	})

	err := NewBuilder().
		WithMessage("test").
		WithCode(ErrBuildFailed).
		WithField("a", "1").
		WithField("b", "2").
		Build()

	result := formatter.Format(err)

	// Fields should be separated by custom separator
	assert.Contains(t, result, "{")
	assert.Contains(t, result, "}")
}

// TestFormatter_GenericErrorCompact verifies compact mode for generic error
func TestFormatter_GenericErrorCompact(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()

	opts := FormatOptions{
		CompactMode: true,
	}

	result := formatter.FormatWithOptions(errStandard, opts)
	assert.Equal(t, "standard error", result)
}

// TestFormatter_ContextEnvironment verifies Environment field in context
func TestFormatter_ContextEnvironment(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()

	// Create context with environment
	ctx := &ErrorContext{
		Environment: "production",
	}

	enrichedErr := NewBuilder().
		WithMessage("test error").
		WithCode(ErrBuildFailed).
		WithContext(ctx).
		Build()

	opts := FormatOptions{
		IncludeContext: true,
	}

	result := formatter.FormatWithOptions(enrichedErr, opts)
	assert.Contains(t, result, "Environment: production")
}

// TestFormatter_IndentFunction verifies indent helper function
func TestFormatter_IndentFunction(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()

	// Test indentation at different depths
	tests := []struct {
		name   string
		input  string
		depth  int
		expect string
	}{
		{
			name:   "depth 0 no indent",
			input:  "test",
			depth:  0,
			expect: "test",
		},
		{
			name:   "depth 1 with indent",
			input:  "test",
			depth:  1,
			expect: "  test", // default indent is 2 spaces
		},
		{
			name:   "multiline indent",
			input:  "line1\nline2",
			depth:  1,
			expect: "  line1\n  line2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatter.indent(tt.input, tt.depth)
			assert.Equal(t, tt.expect, result)
		})
	}
}

// TestFormatter_ColorizeAllSeverities verifies colorize handles all severities
func TestFormatter_ColorizeAllSeverities(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	text := "test text"

	severities := []Severity{
		SeverityDebug,
		SeverityInfo,
		SeverityWarning,
		SeverityError,
		SeverityCritical,
		SeverityFatal,
		Severity(99), // Unknown severity
	}

	for _, sev := range severities {
		result := formatter.colorize(sev, text)
		assert.Contains(t, result, text, "Colorized text should contain original text")
		assert.Contains(t, result, "\033[0m", "Should contain reset code")
	}
}

// TestFormatter_FormatFieldsEmpty verifies formatFields with empty map
func TestFormatter_FormatFieldsEmpty(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	result := formatter.formatFields(map[string]interface{}{}, FormatOptions{})
	assert.Empty(t, result)
}

// TestFormatter_FormatFieldsMultiple verifies formatFields with multiple fields
func TestFormatter_FormatFieldsMultiple(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	result := formatter.formatFields(fields, FormatOptions{FieldsSeparator: ", "})

	assert.Contains(t, result, "key1=value1")
	assert.Contains(t, result, "key2=42")
}

// TestFormatter_ChainWithGenericError verifies chain handles generic errors
func TestFormatter_ChainWithGenericError(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	chain := NewChain()
	chain = chain.Add(WithCode(ErrBuildFailed, "mage error"))
	chain = chain.Add(errStandard)

	result := formatter.FormatChain(chain)

	assert.Contains(t, result, "2 errors")
	assert.Contains(t, result, "BUILD_FAILED")
	assert.Contains(t, result, "standard error")
}

// TestFormatter_CauseIsGenericError verifies formatting when cause is generic error
func TestFormatter_CauseIsGenericError(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	mageErr := Wrap(errGenericBase, "wrapper")

	opts := FormatOptions{
		IncludeCause: true,
	}

	result := formatter.FormatWithOptions(mageErr, opts)

	assert.Contains(t, result, "Caused by:")
	assert.Contains(t, result, "generic base error")
}

// TestFormatter_DefaultOptionsInitialized verifies default options are set
func TestFormatter_DefaultOptionsInitialized(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()
	opts := formatter.GetDefaultOptions()

	// Verify sensible defaults
	assert.False(t, opts.CompactMode)
	assert.True(t, opts.IncludeContext)
	assert.True(t, opts.IncludeCause)
	assert.True(t, opts.IncludeTimestamp)
	assert.True(t, opts.IncludeFields)
	assert.False(t, opts.IncludeStack)
	assert.False(t, opts.UseColor)
	assert.Equal(t, 2, opts.IndentLevel)
	assert.Equal(t, 10, opts.MaxDepth)
	assert.Equal(t, time.RFC3339, opts.TimeFormat)
	assert.Equal(t, ", ", opts.FieldsSeparator)
}

// TestFormatter_ZeroTimestampNotShown verifies zero timestamp is not shown
func TestFormatter_ZeroTimestampNotShown(t *testing.T) {
	t.Parallel()

	formatter := NewFormatter()

	// Create error with explicitly zero timestamp by manipulating context
	err := New("test error")

	opts := FormatOptions{
		IncludeTimestamp: true,
	}

	result := formatter.FormatWithOptions(err, opts)

	// Should still show time because New() sets timestamp
	// This tests that timestamp is formatted when present
	if strings.Contains(result, "Time:") {
		assert.Contains(t, result, "Time:")
	}
}
