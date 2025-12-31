package errors

import (
	"context"
	stderrors "errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockMageErrorOperations tests all MockMageError methods for coverage.
func TestMockMageErrorOperations(t *testing.T) {
	t.Run("NewMockMageError", func(t *testing.T) {
		mock := NewMockMageError("test error")
		require.NotNil(t, mock)
		assert.Equal(t, "test error", mock.Error())
		assert.Equal(t, ErrUnknown, mock.Code())
		assert.Equal(t, SeverityError, mock.Severity())
	})

	t.Run("WithContext", func(t *testing.T) {
		mock := NewMockMageError("test")
		ctx := &ErrorContext{
			Operation: "test-op",
			Resource:  "test-resource",
			Fields:    map[string]interface{}{"key": "value"},
		}

		result := mock.WithContext(ctx)
		require.NotNil(t, result)
		assert.Equal(t, "test-op", result.Context().Operation)
	})

	t.Run("WithContext_nil", func(t *testing.T) {
		mock := NewMockMageError("test")
		result := mock.WithContext(nil)
		require.NotNil(t, result)
	})

	t.Run("WithField", func(t *testing.T) {
		mock := NewMockMageError("test")
		result := mock.WithField("key", "value")
		require.NotNil(t, result)
		assert.Equal(t, 1, mock.GetCallCount("WithField"))
	})

	t.Run("WithFields", func(t *testing.T) {
		mock := NewMockMageError("test")
		fields := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		}
		result := mock.WithFields(fields)
		require.NotNil(t, result)
		assert.Equal(t, 1, mock.GetCallCount("WithFields"))
	})

	t.Run("WithCause", func(t *testing.T) {
		mock := NewMockMageError("test")
		cause := New("cause error")
		result := mock.WithCause(cause)
		require.NotNil(t, result)
		assert.Equal(t, cause, result.Cause())
	})

	t.Run("WithOperation", func(t *testing.T) {
		mock := NewMockMageError("test")
		result := mock.WithOperation("test-operation")
		require.NotNil(t, result)
		assert.Equal(t, "test-operation", result.Context().Operation)
	})

	t.Run("WithResource", func(t *testing.T) {
		mock := NewMockMageError("test")
		result := mock.WithResource("test-resource")
		require.NotNil(t, result)
		assert.Equal(t, "test-resource", result.Context().Resource)
	})

	t.Run("Format", func(t *testing.T) {
		mock := NewMockMageError("formatted message")
		result := mock.Format(true)
		assert.Equal(t, "formatted message", result)
		assert.Equal(t, 1, mock.GetCallCount("Format"))
	})

	t.Run("Is", func(t *testing.T) {
		mock := NewMockMageError("test")
		result := mock.Is(New("other"))
		assert.False(t, result)
		assert.Equal(t, 1, mock.GetCallCount("Is"))
	})

	t.Run("As", func(t *testing.T) {
		mock := NewMockMageError("test")
		var target *MockMageError
		result := mock.As(&target)
		assert.False(t, result)
		assert.Equal(t, 1, mock.GetCallCount("As"))
	})

	t.Run("GetCallCount", func(t *testing.T) {
		mock := NewMockMageError("test")
		_ = mock.Error()
		_ = mock.Error()
		_ = mock.Code()
		assert.Equal(t, 2, mock.GetCallCount("Error"))
		assert.Equal(t, 1, mock.GetCallCount("Code"))
		assert.Equal(t, 0, mock.GetCallCount("NonExistent"))
	})
}

// TestMockErrorBuilderOperations tests all MockErrorBuilder methods for coverage.
func TestMockErrorBuilderOperations(t *testing.T) {
	t.Run("NewMockErrorBuilder", func(t *testing.T) {
		builder := NewMockErrorBuilder()
		require.NotNil(t, builder)
	})

	t.Run("WithMessage", func(t *testing.T) {
		builder := NewMockErrorBuilder()
		result := builder.WithMessage("test %s", "message")
		assert.Equal(t, builder, result)
	})

	t.Run("WithCode", func(t *testing.T) {
		builder := NewMockErrorBuilder()
		result := builder.WithCode(ErrConfigInvalid)
		assert.Equal(t, builder, result)
	})

	t.Run("WithSeverity", func(t *testing.T) {
		builder := NewMockErrorBuilder()
		result := builder.WithSeverity(SeverityCritical)
		assert.Equal(t, builder, result)
	})

	t.Run("WithContext", func(t *testing.T) {
		builder := NewMockErrorBuilder()
		ctx := &ErrorContext{Operation: "test"}
		result := builder.WithContext(ctx)
		assert.Equal(t, builder, result)
	})

	t.Run("WithContext_nil", func(t *testing.T) {
		builder := NewMockErrorBuilder()
		result := builder.WithContext(nil)
		assert.Equal(t, builder, result)
	})

	t.Run("WithField", func(t *testing.T) {
		builder := NewMockErrorBuilder()
		result := builder.WithField("key", "value")
		assert.Equal(t, builder, result)
	})

	t.Run("WithFields", func(t *testing.T) {
		builder := NewMockErrorBuilder()
		result := builder.WithFields(map[string]interface{}{"key": "value"})
		assert.Equal(t, builder, result)
	})

	t.Run("WithCause", func(t *testing.T) {
		builder := NewMockErrorBuilder()
		result := builder.WithCause(New("cause"))
		assert.Equal(t, builder, result)
	})

	t.Run("WithOperation", func(t *testing.T) {
		builder := NewMockErrorBuilder()
		result := builder.WithOperation("test-op")
		assert.Equal(t, builder, result)
	})

	t.Run("WithResource", func(t *testing.T) {
		builder := NewMockErrorBuilder()
		result := builder.WithResource("test-resource")
		assert.Equal(t, builder, result)
	})

	t.Run("WithStackTrace", func(t *testing.T) {
		builder := NewMockErrorBuilder()
		result := builder.WithStackTrace()
		assert.Equal(t, builder, result)
	})

	t.Run("Build", func(t *testing.T) {
		builder := NewMockErrorBuilder()
		builder.WithCode(ErrBuildFailed)
		builder.WithSeverity(SeverityWarning)
		result := builder.Build()
		require.NotNil(t, result)
	})

	t.Run("GetCallCount", func(t *testing.T) {
		builder := NewMockErrorBuilder()
		builder.WithMessage("test")
		builder.WithMessage("test2")
		builder.WithCode(ErrBuildFailed)
		assert.Equal(t, 2, builder.GetCallCount("WithMessage"))
		assert.Equal(t, 1, builder.GetCallCount("WithCode"))
	})
}

// TestDefaultErrorMetricsOperations tests DefaultErrorMetrics methods for coverage.
func TestDefaultErrorMetricsOperations(t *testing.T) {
	t.Run("RecordError", func(t *testing.T) {
		metrics := &DefaultErrorMetrics{
			counts: make(map[ErrorCode]*ErrorStat),
		}
		err := New("test error")
		// Just ensure it doesn't panic
		metrics.RecordError(err)
	})

	t.Run("RecordMageError", func(t *testing.T) {
		metrics := &DefaultErrorMetrics{
			counts: make(map[ErrorCode]*ErrorStat),
		}
		mageErr := NewMageError("test")
		// Just ensure it doesn't panic
		metrics.RecordMageError(mageErr)
	})

	t.Run("GetCount", func(t *testing.T) {
		metrics := &DefaultErrorMetrics{
			counts: map[ErrorCode]*ErrorStat{
				ErrBuildFailed: {Code: ErrBuildFailed, Count: 5},
			},
		}
		assert.Equal(t, int64(5), metrics.GetCount(ErrBuildFailed))
		assert.Equal(t, int64(0), metrics.GetCount(ErrConfigInvalid))
	})

	t.Run("GetCountBySeverity", func(t *testing.T) {
		metrics := &DefaultErrorMetrics{
			counts: map[ErrorCode]*ErrorStat{
				ErrBuildFailed:   {Code: ErrBuildFailed, Count: 3},
				ErrConfigInvalid: {Code: ErrConfigInvalid, Count: 2},
			},
		}
		count := metrics.GetCountBySeverity(SeverityError)
		assert.Equal(t, int64(2), count) // Counts number of entries, not sum
	})

	t.Run("GetRate", func(t *testing.T) {
		now := time.Now()
		metrics := &DefaultErrorMetrics{
			counts: map[ErrorCode]*ErrorStat{
				ErrBuildFailed: {
					Code:      ErrBuildFailed,
					Count:     10,
					FirstSeen: now.Add(-10 * time.Second),
				},
			},
		}
		rate := metrics.GetRate(ErrBuildFailed, time.Minute)
		assert.Greater(t, rate, float64(0))

		// Test non-existent code
		zeroRate := metrics.GetRate(ErrConfigInvalid, time.Minute)
		assert.InDelta(t, float64(0), zeroRate, 0.001)
	})

	t.Run("GetRate_zero_count", func(t *testing.T) {
		metrics := &DefaultErrorMetrics{
			counts: map[ErrorCode]*ErrorStat{
				ErrBuildFailed: {Code: ErrBuildFailed, Count: 0, FirstSeen: time.Now()},
			},
		}
		rate := metrics.GetRate(ErrBuildFailed, time.Minute)
		assert.InDelta(t, float64(0), rate, 0.001)
	})

	t.Run("GetTopErrors", func(t *testing.T) {
		metrics := &DefaultErrorMetrics{
			counts: map[ErrorCode]*ErrorStat{
				ErrBuildFailed:   {Code: ErrBuildFailed, Count: 10},
				ErrConfigInvalid: {Code: ErrConfigInvalid, Count: 5},
				ErrUnknown:       {Code: ErrUnknown, Count: 3},
			},
		}

		// Get all
		all := metrics.GetTopErrors(0)
		assert.Len(t, all, 3)
		assert.Equal(t, int64(10), all[0].Count) // Should be sorted by count desc

		// Get top 2
		top2 := metrics.GetTopErrors(2)
		assert.Len(t, top2, 2)
	})

	t.Run("Reset", func(t *testing.T) {
		metrics := &DefaultErrorMetrics{
			counts: map[ErrorCode]*ErrorStat{
				ErrBuildFailed: {Code: ErrBuildFailed, Count: 10},
			},
		}
		err := metrics.Reset()
		require.NoError(t, err)
		assert.Empty(t, metrics.counts)
	})
}

// TestStructuredErrorLoggerLogMageErrorWithContext tests StructuredErrorLogger.LogMageErrorWithContext.
func TestStructuredErrorLoggerLogMageErrorWithContext(t *testing.T) {
	t.Run("disabled_logger", func(t *testing.T) {
		logger := &StructuredErrorLogger{
			DefaultErrorLogger: &DefaultErrorLogger{enabled: false},
			fields:             make(map[string]interface{}),
		}
		mageErr := NewMageError("test")

		// Should not panic when disabled
		logger.LogMageErrorWithContext(context.Background(), mageErr)
	})

	t.Run("nil_error", func(t *testing.T) {
		logger := &StructuredErrorLogger{
			DefaultErrorLogger: &DefaultErrorLogger{enabled: true},
			fields:             make(map[string]interface{}),
		}

		// Should not panic with nil error
		logger.LogMageErrorWithContext(context.Background(), nil)
	})

	t.Run("with_fields", func(t *testing.T) {
		// Create a logger that writes to a temp file
		tmpDir := t.TempDir()
		logPath := filepath.Join(tmpDir, "test.log")
		file, err := os.Create(logPath) //nolint:gosec // G304: Test file - path from t.TempDir()
		require.NoError(t, err)
		defer func() {
			if closeErr := file.Close(); closeErr != nil {
				t.Logf("failed to close file: %v", closeErr)
			}
		}()

		baseLogger := NewErrorLoggerWithOptions(ErrorLoggerOptions{
			Output:  file,
			Enabled: true,
		})
		defaultLogger, ok := baseLogger.(*DefaultErrorLogger)
		require.True(t, ok, "expected *DefaultErrorLogger")
		logger := &StructuredErrorLogger{
			DefaultErrorLogger: defaultLogger,
			fields: map[string]interface{}{
				"custom_field": "custom_value",
			},
		}

		mageErr := NewMageError("test error")
		logger.LogMageErrorWithContext(context.Background(), mageErr)

		// Verify file was written
		require.NoError(t, file.Sync())
		info, err := os.Stat(logPath)
		require.NoError(t, err)
		assert.Positive(t, info.Size())
	})
}

// TestLoggerSetFormatter tests DefaultErrorLogger.SetFormatter.
func TestLoggerSetFormatter(t *testing.T) {
	t.Run("set_formatter", func(t *testing.T) {
		logger, ok := NewErrorLogger().(*DefaultErrorLogger)
		require.True(t, ok, "expected *DefaultErrorLogger")
		formatter := NewFormatter()

		logger.SetFormatter(formatter)
		assert.Equal(t, formatter, logger.formatter)
	})
}

// TestMatcherFunctions tests matcher functions for coverage.
func TestMatcherFunctions(t *testing.T) {
	t.Run("CodeMatcher", func(t *testing.T) {
		matcher := CodeMatcher(ErrBuildFailed)
		mageErr := NewMageError("test").WithCode(ErrBuildFailed)

		result := matcher.Match(mageErr)
		assert.True(t, result)

		nonMatchErr := NewMageError("test").WithCode(ErrConfigInvalid)
		result = matcher.Match(nonMatchErr)
		assert.False(t, result)
	})

	t.Run("CodeMatcher_with_standard_error", func(t *testing.T) {
		matcher := CodeMatcher(ErrBuildFailed)
		err := New("standard error")

		result := matcher.Match(err)
		assert.False(t, result)
	})

	t.Run("SeverityMatcher", func(t *testing.T) {
		matcher := SeverityMatcher(SeverityError)
		mageErr := NewMageError("test").WithSeverity(SeverityError)

		result := matcher.Match(mageErr)
		assert.True(t, result)

		wrongSeverityErr := NewMageError("test").WithSeverity(SeverityWarning)
		result = matcher.Match(wrongSeverityErr)
		assert.False(t, result)
	})

	t.Run("MatchAll_via_matcher", func(t *testing.T) {
		codeMatcher := CodeMatcher(ErrBuildFailed)
		severityMatcher := SeverityMatcher(SeverityError)

		allMatcher := NewMatcher().MatchAll(codeMatcher, severityMatcher)

		mageErr := NewMageError("test").WithCode(ErrBuildFailed).WithSeverity(SeverityError)
		result := allMatcher.Match(mageErr)
		assert.True(t, result)

		wrongSeverityErr := NewMageError("test").WithCode(ErrBuildFailed).WithSeverity(SeverityWarning)
		result = allMatcher.Match(wrongSeverityErr)
		assert.False(t, result)
	})

	t.Run("Not_via_matcher", func(t *testing.T) {
		matcher := NewMatcher().MatchCode(ErrBuildFailed)
		notMatcher := matcher.Not()

		buildErr := NewMageError("test").WithCode(ErrBuildFailed)
		configErr := NewMageError("test").WithCode(ErrConfigInvalid)

		assert.False(t, notMatcher.Match(buildErr))
		assert.True(t, notMatcher.Match(configErr))
	})
}

// TestDefaultErrorMatcherMatch tests DefaultErrorMatcher.Match coverage.
func TestDefaultErrorMatcherMatch(t *testing.T) {
	t.Run("Match", func(t *testing.T) {
		matcher := &DefaultErrorMatcher{
			matchers: []func(error) bool{
				func(err error) bool {
					var mErr MageError
					if stderrors.As(err, &mErr) {
						return mErr.Code() == ErrBuildFailed
					}
					return false
				},
			},
		}

		mageErr := NewMageError("test").WithCode(ErrBuildFailed)
		result := matcher.Match(mageErr)
		assert.True(t, result)
	})
}
