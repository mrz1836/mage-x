package errors

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Static error variables for tests (err113 compliance)
var (
	errMatcherStandard = errors.New("standard error")
	errMatcherBase     = errors.New("base")
)

// TestRealMatcher_NilError verifies Match(nil) behavior
func TestRealMatcher_NilError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() *RealDefaultErrorMatcher
		expected bool
	}{
		{
			name: "empty matcher returns false for nil",
			setup: func() *RealDefaultErrorMatcher {
				return NewErrorMatcher()
			},
			expected: false,
		},
		{
			name: "matcher with code returns false for nil",
			setup: func() *RealDefaultErrorMatcher {
				m := NewErrorMatcher()
				m.MatchCode(ErrBuildFailed)
				return m
			},
			expected: false,
		},
		{
			name: "inverted empty matcher returns true for nil",
			setup: func() *RealDefaultErrorMatcher {
				m := NewErrorMatcher()
				result, ok := m.Not().(*RealDefaultErrorMatcher)
				if !ok {
					return nil
				}
				return result
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			matcher := tt.setup()
			result := matcher.Match(nil)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestRealMatcher_NoMatchersMatchesAll verifies empty matcher matches any error
func TestRealMatcher_NoMatchersMatchesAll(t *testing.T) {
	t.Parallel()

	matcher := NewErrorMatcher()

	tests := []struct {
		name string
		err  error
	}{
		{
			name: "matches MageError",
			err:  WithCode(ErrBuildFailed, "build failed"),
		},
		{
			name: "matches standard error",
			err:  errMatcherStandard,
		},
		{
			name: "matches wrapped error",
			err:  Wrap(errMatcherBase, "wrapped"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.True(t, matcher.Match(tt.err), "Empty matcher should match any error")
		})
	}
}

// TestRealMatcher_MatchCode verifies MatchCode filters by error code
func TestRealMatcher_MatchCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		matcherCode ErrorCode
		errorCode   ErrorCode
		shouldMatch bool
	}{
		{
			name:        "matches same code",
			matcherCode: ErrBuildFailed,
			errorCode:   ErrBuildFailed,
			shouldMatch: true,
		},
		{
			name:        "does not match different code",
			matcherCode: ErrBuildFailed,
			errorCode:   ErrTestFailed,
			shouldMatch: false,
		},
		{
			name:        "matches custom code",
			matcherCode: "CUSTOM_CODE",
			errorCode:   "CUSTOM_CODE",
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			matcher := NewErrorMatcher().MatchCode(tt.matcherCode)
			err := WithCode(tt.errorCode, "test error")
			result := matcher.Match(err)
			assert.Equal(t, tt.shouldMatch, result)
		})
	}
}

// TestRealMatcher_MatchCodeNonMageError verifies non-MageError doesn't match code
func TestRealMatcher_MatchCodeNonMageError(t *testing.T) {
	t.Parallel()

	matcher := NewErrorMatcher().MatchCode(ErrBuildFailed)

	assert.False(t, matcher.Match(errMatcherStandard), "Non-MageError should not match code matcher")
}

// TestRealMatcher_MatchSeverity verifies MatchSeverity filters by severity
func TestRealMatcher_MatchSeverity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		matcherSeverity Severity
		errorSeverity   Severity
		shouldMatch     bool
	}{
		{
			name:            "matches same severity",
			matcherSeverity: SeverityError,
			errorSeverity:   SeverityError,
			shouldMatch:     true,
		},
		{
			name:            "does not match different severity",
			matcherSeverity: SeverityCritical,
			errorSeverity:   SeverityWarning,
			shouldMatch:     false,
		},
		{
			name:            "matches fatal severity",
			matcherSeverity: SeverityFatal,
			errorSeverity:   SeverityFatal,
			shouldMatch:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			matcher := NewErrorMatcher().MatchSeverity(tt.matcherSeverity)
			err := NewBuilder().
				WithMessage("test").
				WithSeverity(tt.errorSeverity).
				Build()
			result := matcher.Match(err)
			assert.Equal(t, tt.shouldMatch, result)
		})
	}
}

// TestRealMatcher_MatchSeverityNonMageError verifies non-MageError doesn't match severity
func TestRealMatcher_MatchSeverityNonMageError(t *testing.T) {
	t.Parallel()

	matcher := NewErrorMatcher().MatchSeverity(SeverityError)

	assert.False(t, matcher.Match(errMatcherStandard), "Non-MageError should not match severity matcher")
}

// TestRealMatcher_MatchMessageRegex verifies valid regex matches messages
func TestRealMatcher_MatchMessageRegex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		pattern     string
		message     string
		shouldMatch bool
	}{
		{
			name:        "simple regex match",
			pattern:     "build.*failed",
			message:     "build process failed",
			shouldMatch: true,
		},
		{
			name:        "case sensitive no match",
			pattern:     "BUILD",
			message:     "build failed",
			shouldMatch: false,
		},
		{
			name:        "word boundary match",
			pattern:     `\bfailed\b`,
			message:     "operation failed successfully",
			shouldMatch: true,
		},
		{
			name:        "no match",
			pattern:     "timeout",
			message:     "build failed",
			shouldMatch: false,
		},
		{
			name:        "anchored pattern",
			pattern:     "^build",
			message:     "build failed",
			shouldMatch: true,
		},
		{
			name:        "anchored pattern no match",
			pattern:     "^failed",
			message:     "build failed",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			matcher := NewErrorMatcher().MatchMessage(tt.pattern)
			err := New(tt.message)
			result := matcher.Match(err)
			assert.Equal(t, tt.shouldMatch, result)
		})
	}
}

// TestRealMatcher_MatchMessageInvalidRegex verifies invalid regex falls back to exact match
func TestRealMatcher_MatchMessageInvalidRegex(t *testing.T) {
	t.Parallel()

	// Invalid regex pattern (unclosed bracket)
	invalidPattern := "[invalid"

	matcher := NewErrorMatcher().MatchMessage(invalidPattern)

	// Should fall back to exact string match
	exactMatch := New(invalidPattern)
	assert.True(t, matcher.Match(exactMatch), "Should match exact string when regex is invalid")

	partialMatch := New("this contains [invalid but more")
	assert.False(t, matcher.Match(partialMatch), "Should not match partial string")
}

// TestRealMatcher_MatchType verifies MatchType uses errors.Is
func TestRealMatcher_MatchType(t *testing.T) {
	t.Parallel()

	targetErr := WithCode(ErrBuildFailed, "target")
	matcher := NewErrorMatcher().MatchType(targetErr)

	// Same error should match
	assert.True(t, matcher.Match(targetErr))

	// Error with same code should match (via Is)
	sameCode := WithCode(ErrBuildFailed, "different message")
	assert.True(t, matcher.Match(sameCode))

	// Different error should not match
	differentErr := WithCode(ErrTestFailed, "different")
	assert.False(t, matcher.Match(differentErr))
}

// TestRealMatcher_MatchField verifies MatchField checks context fields
func TestRealMatcher_MatchField(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		fieldKey    string
		fieldValue  interface{}
		errFields   map[string]interface{}
		shouldMatch bool
	}{
		{
			name:        "matches string field",
			fieldKey:    "env",
			fieldValue:  "production",
			errFields:   map[string]interface{}{"env": "production"},
			shouldMatch: true,
		},
		{
			name:        "matches int field",
			fieldKey:    "exitCode",
			fieldValue:  1,
			errFields:   map[string]interface{}{"exitCode": 1},
			shouldMatch: true,
		},
		{
			name:        "does not match different value",
			fieldKey:    "env",
			fieldValue:  "production",
			errFields:   map[string]interface{}{"env": "staging"},
			shouldMatch: false,
		},
		{
			name:        "does not match different type",
			fieldKey:    "count",
			fieldValue:  "1",
			errFields:   map[string]interface{}{"count": 1},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			matcher := NewErrorMatcher().MatchField(tt.fieldKey, tt.fieldValue)
			err := NewBuilder().
				WithMessage("test").
				WithFields(tt.errFields).
				Build()
			result := matcher.Match(err)
			assert.Equal(t, tt.shouldMatch, result)
		})
	}
}

// TestRealMatcher_MatchFieldMissing verifies missing field returns false
func TestRealMatcher_MatchFieldMissing(t *testing.T) {
	t.Parallel()

	matcher := NewErrorMatcher().MatchField("nonexistent", "value")
	err := NewBuilder().
		WithMessage("test").
		WithField("other", "value").
		Build()

	assert.False(t, matcher.Match(err), "Missing field should not match")
}

// TestRealMatcher_MatchFieldNonMageError verifies non-MageError doesn't match field
func TestRealMatcher_MatchFieldNonMageError(t *testing.T) {
	t.Parallel()

	matcher := NewErrorMatcher().MatchField("key", "value")

	assert.False(t, matcher.Match(errMatcherStandard), "Non-MageError should not match field matcher")
}

// TestRealMatcher_MatchAny verifies any submatcher matching passes
func TestRealMatcher_MatchAny(t *testing.T) {
	t.Parallel()

	buildMatcher := NewErrorMatcher().MatchCode(ErrBuildFailed)
	testMatcher := NewErrorMatcher().MatchCode(ErrTestFailed)

	anyMatcher := NewErrorMatcher().MatchAny(buildMatcher, testMatcher)

	// Should match build error
	buildErr := WithCode(ErrBuildFailed, "build")
	assert.True(t, anyMatcher.Match(buildErr))

	// Should match test error
	testErr := WithCode(ErrTestFailed, "test")
	assert.True(t, anyMatcher.Match(testErr))

	// Should not match config error
	configErr := WithCode(ErrConfigInvalid, "config")
	assert.False(t, anyMatcher.Match(configErr))
}

// TestRealMatcher_MatchAnyEmpty verifies empty MatchAny returns false
func TestRealMatcher_MatchAnyEmpty(t *testing.T) {
	t.Parallel()

	anyMatcher := NewErrorMatcher().MatchAny() // No matchers
	err := WithCode(ErrBuildFailed, "test")

	assert.False(t, anyMatcher.Match(err), "MatchAny with no matchers should return false")
}

// TestRealMatcher_MatchAll verifies all submatchers must match
func TestRealMatcher_MatchAll(t *testing.T) {
	t.Parallel()

	codeMatcher := NewErrorMatcher().MatchCode(ErrBuildFailed)
	severityMatcher := NewErrorMatcher().MatchSeverity(SeverityCritical)

	allMatcher := NewErrorMatcher().MatchAll(codeMatcher, severityMatcher)

	// Should match error with both conditions
	matchBoth := NewBuilder().
		WithMessage("test").
		WithCode(ErrBuildFailed).
		WithSeverity(SeverityCritical).
		Build()
	assert.True(t, allMatcher.Match(matchBoth))

	// Should not match error with only code
	codeOnly := NewBuilder().
		WithMessage("test").
		WithCode(ErrBuildFailed).
		WithSeverity(SeverityWarning).
		Build()
	assert.False(t, allMatcher.Match(codeOnly))

	// Should not match error with only severity
	severityOnly := NewBuilder().
		WithMessage("test").
		WithCode(ErrTestFailed).
		WithSeverity(SeverityCritical).
		Build()
	assert.False(t, allMatcher.Match(severityOnly))
}

// TestRealMatcher_MatchAllEmpty verifies empty MatchAll returns true
func TestRealMatcher_MatchAllEmpty(t *testing.T) {
	t.Parallel()

	allMatcher := NewErrorMatcher().MatchAll() // No matchers
	err := WithCode(ErrBuildFailed, "test")

	assert.True(t, allMatcher.Match(err), "MatchAll with no matchers should return true")
}

// TestRealMatcher_Not verifies Not() inverts match result
func TestRealMatcher_Not(t *testing.T) {
	t.Parallel()

	buildMatcher := NewErrorMatcher().MatchCode(ErrBuildFailed)
	notBuildMatcher := buildMatcher.Not()

	buildErr := WithCode(ErrBuildFailed, "build")
	testErr := WithCode(ErrTestFailed, "test")

	// Original matcher behavior
	assert.True(t, buildMatcher.Match(buildErr))
	assert.False(t, buildMatcher.Match(testErr))

	// Inverted matcher behavior
	assert.False(t, notBuildMatcher.Match(buildErr))
	assert.True(t, notBuildMatcher.Match(testErr))
}

// TestRealMatcher_NotNot verifies double negation
func TestRealMatcher_NotNot(t *testing.T) {
	t.Parallel()

	buildMatcher := NewErrorMatcher().MatchCode(ErrBuildFailed)
	doubleNot := buildMatcher.Not().Not()

	buildErr := WithCode(ErrBuildFailed, "build")
	testErr := WithCode(ErrTestFailed, "test")

	// Double negation should match original behavior
	assert.True(t, doubleNot.Match(buildErr))
	assert.False(t, doubleNot.Match(testErr))
}

// TestRealMatcher_ChainedMatchers verifies multiple matchers AND together
func TestRealMatcher_ChainedMatchers(t *testing.T) {
	t.Parallel()

	matcher := NewErrorMatcher().
		MatchCode(ErrBuildFailed).
		MatchSeverity(SeverityCritical).
		MatchMessage("critical")

	// Should match all conditions
	matchAll := NewBuilder().
		WithMessage("critical build error").
		WithCode(ErrBuildFailed).
		WithSeverity(SeverityCritical).
		Build()
	assert.True(t, matcher.Match(matchAll))

	// Should not match if any condition fails
	wrongCode := NewBuilder().
		WithMessage("critical build error").
		WithCode(ErrTestFailed).
		WithSeverity(SeverityCritical).
		Build()
	assert.False(t, matcher.Match(wrongCode))

	wrongSeverity := NewBuilder().
		WithMessage("critical build error").
		WithCode(ErrBuildFailed).
		WithSeverity(SeverityWarning).
		Build()
	assert.False(t, matcher.Match(wrongSeverity))

	wrongMessage := NewBuilder().
		WithMessage("normal build error").
		WithCode(ErrBuildFailed).
		WithSeverity(SeverityCritical).
		Build()
	assert.False(t, matcher.Match(wrongMessage))
}

// TestRealMatcher_CodeMatcher verifies CodeMatcher convenience function
func TestRealMatcher_CodeMatcher(t *testing.T) {
	t.Parallel()

	// Single code
	singleMatcher := CodeMatcher(ErrBuildFailed)
	assert.True(t, singleMatcher.Match(WithCode(ErrBuildFailed, "build")))
	assert.False(t, singleMatcher.Match(WithCode(ErrTestFailed, "test")))

	// Multiple codes
	multiMatcher := CodeMatcher(ErrBuildFailed, ErrTestFailed, ErrLintFailed)
	assert.True(t, multiMatcher.Match(WithCode(ErrBuildFailed, "build")))
	assert.True(t, multiMatcher.Match(WithCode(ErrTestFailed, "test")))
	assert.True(t, multiMatcher.Match(WithCode(ErrLintFailed, "lint")))
	assert.False(t, multiMatcher.Match(WithCode(ErrConfigInvalid, "config")))
}

// TestRealMatcher_SeverityMatcher verifies SeverityMatcher convenience function
func TestRealMatcher_SeverityMatcher(t *testing.T) {
	t.Parallel()

	// Single severity
	singleMatcher := SeverityMatcher(SeverityCritical)
	criticalErr := NewBuilder().WithMessage("test").WithSeverity(SeverityCritical).Build()
	warningErr := NewBuilder().WithMessage("test").WithSeverity(SeverityWarning).Build()

	assert.True(t, singleMatcher.Match(criticalErr))
	assert.False(t, singleMatcher.Match(warningErr))

	// Multiple severities
	multiMatcher := SeverityMatcher(SeverityCritical, SeverityFatal)
	fatalErr := NewBuilder().WithMessage("test").WithSeverity(SeverityFatal).Build()

	assert.True(t, multiMatcher.Match(criticalErr))
	assert.True(t, multiMatcher.Match(fatalErr))
	assert.False(t, multiMatcher.Match(warningErr))
}

// TestRealMatcher_MessageMatcher verifies MessageMatcher convenience function
func TestRealMatcher_MessageMatcher(t *testing.T) {
	t.Parallel()

	matcher := MessageMatcher("timeout")

	assert.True(t, matcher.Match(New("connection timeout")))
	assert.True(t, matcher.Match(New("timeout occurred")))
	assert.False(t, matcher.Match(New("build failed")))
}

// TestRealMatcher_TypeMatcher verifies TypeMatcher convenience function
func TestRealMatcher_TypeMatcher(t *testing.T) {
	t.Parallel()

	targetErr := WithCode(ErrTimeout, "target")
	matcher := TypeMatcher(targetErr)

	assert.True(t, matcher.Match(WithCode(ErrTimeout, "any timeout")))
	assert.False(t, matcher.Match(WithCode(ErrBuildFailed, "build")))
}

// TestRealMatcher_FieldMatcher verifies FieldMatcher convenience function
func TestRealMatcher_FieldMatcher(t *testing.T) {
	t.Parallel()

	matcher := FieldMatcher("env", "production")

	prodErr := NewBuilder().WithMessage("test").WithField("env", "production").Build()
	devErr := NewBuilder().WithMessage("test").WithField("env", "development").Build()

	assert.True(t, matcher.Match(prodErr))
	assert.False(t, matcher.Match(devErr))
}

// TestRealMatcher_CriticalMatcher verifies CriticalMatcher matches Critical/Fatal
func TestRealMatcher_CriticalMatcher(t *testing.T) {
	t.Parallel()

	matcher := CriticalMatcher()

	criticalErr := NewBuilder().WithMessage("test").WithSeverity(SeverityCritical).Build()
	fatalErr := NewBuilder().WithMessage("test").WithSeverity(SeverityFatal).Build()
	errorErr := NewBuilder().WithMessage("test").WithSeverity(SeverityError).Build()
	warningErr := NewBuilder().WithMessage("test").WithSeverity(SeverityWarning).Build()

	assert.True(t, matcher.Match(criticalErr), "Should match Critical")
	assert.True(t, matcher.Match(fatalErr), "Should match Fatal")
	assert.False(t, matcher.Match(errorErr), "Should not match Error")
	assert.False(t, matcher.Match(warningErr), "Should not match Warning")
}

// TestRealMatcher_RetryableMatcher verifies RetryableMatcher matches retryable codes
func TestRealMatcher_RetryableMatcher(t *testing.T) {
	t.Parallel()

	matcher := RetryableMatcher()

	retryableErrors := []ErrorCode{
		ErrInternal,
		ErrTimeout,
		ErrResourceExhausted,
		ErrUnavailable,
		ErrBuildFailed,
		ErrTestFailed,
		ErrCommandTimeout,
	}

	nonRetryableErrors := []ErrorCode{
		ErrNotFound,
		ErrPermissionDenied,
		ErrInvalidArgument,
	}

	for _, code := range retryableErrors {
		err := WithCode(code, "test")
		assert.True(t, matcher.Match(err), "%s should be matched as retryable", code)
	}

	for _, code := range nonRetryableErrors {
		err := WithCode(code, "test")
		assert.False(t, matcher.Match(err), "%s should not be matched as retryable", code)
	}
}

// TestRealMatcher_BuildErrorMatcher verifies BuildErrorMatcher matches build codes
func TestRealMatcher_BuildErrorMatcher(t *testing.T) {
	t.Parallel()

	matcher := BuildErrorMatcher()

	buildErrors := []ErrorCode{
		ErrBuildFailed,
		ErrCompileFailed,
		ErrTestFailed,
		ErrLintFailed,
		ErrPackageFailed,
		ErrDependencyError,
	}

	nonBuildErrors := []ErrorCode{
		ErrNotFound,
		ErrTimeout,
		ErrConfigInvalid,
	}

	for _, code := range buildErrors {
		err := WithCode(code, "test")
		assert.True(t, matcher.Match(err), "%s should be matched as build error", code)
	}

	for _, code := range nonBuildErrors {
		err := WithCode(code, "test")
		assert.False(t, matcher.Match(err), "%s should not be matched as build error", code)
	}
}

// TestRealMatcher_Concurrent verifies concurrent matcher operations are safe
func TestRealMatcher_Concurrent(t *testing.T) {
	t.Parallel()

	matcher := NewErrorMatcher()
	var wg sync.WaitGroup

	// Concurrent MatchCode calls
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			matcher.MatchCode(ErrBuildFailed)
		}()
	}

	// Concurrent Match calls
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := WithCode(ErrBuildFailed, "test")
			_ = matcher.Match(err)
		}()
	}

	wg.Wait()
	// Test passes if no race conditions occur
}

// TestRealMatcher_FluentInterface verifies method chaining returns matcher
func TestRealMatcher_FluentInterface(t *testing.T) {
	t.Parallel()

	matcher := NewErrorMatcher()

	result := matcher.
		MatchCode(ErrBuildFailed).
		MatchSeverity(SeverityError).
		MatchMessage("test").
		MatchField("key", "value")

	// Should return ErrorMatcher interface (verified at compile time)
	assert.NotNil(t, result)
	// Verify it can match errors (proves it's a valid ErrorMatcher)
	assert.True(t, result.Match(NewBuilder().
		WithMessage("test").
		WithCode(ErrBuildFailed).
		WithSeverity(SeverityError).
		WithField("key", "value").
		Build()), "Chained methods should return working ErrorMatcher")
}

// TestRealMatcher_MatchAnyNested verifies nested MatchAny operations
func TestRealMatcher_MatchAnyNested(t *testing.T) {
	t.Parallel()

	// Match (build OR test) AND critical
	buildOrTest := NewErrorMatcher().MatchAny(
		NewErrorMatcher().MatchCode(ErrBuildFailed),
		NewErrorMatcher().MatchCode(ErrTestFailed),
	)

	criticalBuildOrTest := NewErrorMatcher().MatchAll(
		buildOrTest,
		NewErrorMatcher().MatchSeverity(SeverityCritical),
	)

	// Critical build error - should match
	criticalBuild := NewBuilder().
		WithMessage("test").
		WithCode(ErrBuildFailed).
		WithSeverity(SeverityCritical).
		Build()
	assert.True(t, criticalBuildOrTest.Match(criticalBuild))

	// Critical test error - should match
	criticalTest := NewBuilder().
		WithMessage("test").
		WithCode(ErrTestFailed).
		WithSeverity(SeverityCritical).
		Build()
	assert.True(t, criticalBuildOrTest.Match(criticalTest))

	// Warning build error - should not match (not critical)
	warningBuild := NewBuilder().
		WithMessage("test").
		WithCode(ErrBuildFailed).
		WithSeverity(SeverityWarning).
		Build()
	assert.False(t, criticalBuildOrTest.Match(warningBuild))

	// Critical config error - should not match (not build or test)
	criticalConfig := NewBuilder().
		WithMessage("test").
		WithCode(ErrConfigInvalid).
		WithSeverity(SeverityCritical).
		Build()
	assert.False(t, criticalBuildOrTest.Match(criticalConfig))
}
