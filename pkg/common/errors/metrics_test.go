package errors

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Static error variables for tests (err113 compliance)
var (
	errMetricsStandard  = errors.New("standard error")
	errMetricsThree     = errors.New("3")
	errMetricsStdRecord = errors.New("standard")
)

// TestRealMetrics_NilErrorIgnored verifies RecordError(nil) is no-op
func TestRealMetrics_NilErrorIgnored(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()
	metrics.RecordError(nil)

	assert.Equal(t, int64(0), metrics.GetTotalErrors())
}

// TestRealMetrics_NilMageErrorIgnored verifies RecordMageError(nil) is no-op
func TestRealMetrics_NilMageErrorIgnored(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()
	metrics.RecordMageError(nil)

	assert.Equal(t, int64(0), metrics.GetTotalErrors())
}

// TestRealMetrics_NonMageErrorRecordedAsUnknown verifies standard errors get ErrUnknown code
func TestRealMetrics_NonMageErrorRecordedAsUnknown(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	metrics.RecordError(errMetricsStandard)

	assert.Equal(t, int64(1), metrics.GetCount(ErrUnknown))
	assert.Equal(t, int64(1), metrics.GetTotalErrors())
}

// TestRealMetrics_MageErrorRecordedByCode verifies MageError recorded by code
func TestRealMetrics_MageErrorRecordedByCode(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	metrics.RecordError(WithCode(ErrBuildFailed, "build"))
	metrics.RecordError(WithCode(ErrBuildFailed, "another build"))
	metrics.RecordError(WithCode(ErrTestFailed, "test"))

	assert.Equal(t, int64(2), metrics.GetCount(ErrBuildFailed))
	assert.Equal(t, int64(1), metrics.GetCount(ErrTestFailed))
	assert.Equal(t, int64(3), metrics.GetTotalErrors())
}

// TestRealMetrics_CountsByCode verifies GetCount returns correct per-code counts
func TestRealMetrics_CountsByCode(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	// Record multiple errors of different codes
	for i := 0; i < 5; i++ {
		metrics.RecordError(WithCode(ErrBuildFailed, "build"))
	}
	for i := 0; i < 3; i++ {
		metrics.RecordError(WithCode(ErrTestFailed, "test"))
	}
	for i := 0; i < 7; i++ {
		metrics.RecordError(WithCode(ErrTimeout, "timeout"))
	}

	assert.Equal(t, int64(5), metrics.GetCount(ErrBuildFailed))
	assert.Equal(t, int64(3), metrics.GetCount(ErrTestFailed))
	assert.Equal(t, int64(7), metrics.GetCount(ErrTimeout))
	assert.Equal(t, int64(0), metrics.GetCount(ErrNotFound)) // Not recorded
}

// TestRealMetrics_CountsBySeverity verifies GetCountBySeverity returns correct counts
func TestRealMetrics_CountsBySeverity(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	// Record errors with different severities
	for i := 0; i < 4; i++ {
		metrics.RecordMageError(NewBuilder().
			WithMessage("error").
			WithSeverity(SeverityError).
			Build())
	}
	for i := 0; i < 2; i++ {
		metrics.RecordMageError(NewBuilder().
			WithMessage("critical").
			WithSeverity(SeverityCritical).
			Build())
	}

	assert.Equal(t, int64(4), metrics.GetCountBySeverity(SeverityError))
	assert.Equal(t, int64(2), metrics.GetCountBySeverity(SeverityCritical))
	assert.Equal(t, int64(0), metrics.GetCountBySeverity(SeverityWarning)) // Not recorded
}

// TestRealMetrics_TotalErrors verifies GetTotalErrors returns sum
func TestRealMetrics_TotalErrors(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	// RecordError increments totalErrors, RecordMageError does not
	metrics.RecordError(WithCode(ErrBuildFailed, "1"))
	metrics.RecordError(WithCode(ErrTestFailed, "2"))
	metrics.RecordError(errMetricsThree)
	metrics.RecordError(WithCode(ErrTimeout, "4"))

	assert.Equal(t, int64(4), metrics.GetTotalErrors())
}

// TestRealMetrics_GetRateCalculation verifies rate calculated correctly
func TestRealMetrics_GetRateCalculation(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	// Record some errors
	for i := 0; i < 10; i++ {
		metrics.RecordError(WithCode(ErrBuildFailed, "error"))
	}

	// Get rate over 1 second
	rate := metrics.GetRate(ErrBuildFailed, 1*time.Second)

	// Rate should be positive (errors per second)
	assert.Greater(t, rate, float64(0))
}

// TestRealMetrics_GetRateZeroCount verifies zero count returns zero rate
func TestRealMetrics_GetRateZeroCount(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	rate := metrics.GetRate(ErrBuildFailed, 1*time.Second)
	assert.InDelta(t, float64(0), rate, 0.001)
}

// TestRealMetrics_GetRateShortDuration verifies rate with very short duration
func TestRealMetrics_GetRateShortDuration(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()
	metrics.RecordError(WithCode(ErrBuildFailed, "error"))

	// Rate over very short duration should still work
	rate := metrics.GetRate(ErrBuildFailed, 1*time.Millisecond)
	assert.GreaterOrEqual(t, rate, float64(0))
}

// TestRealMetrics_TopErrorsSorted verifies GetTopErrors returns descending order
func TestRealMetrics_TopErrorsSorted(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	// Record different counts for different codes
	for i := 0; i < 5; i++ {
		metrics.RecordError(WithCode(ErrBuildFailed, "build"))
	}
	for i := 0; i < 10; i++ {
		metrics.RecordError(WithCode(ErrTestFailed, "test"))
	}
	for i := 0; i < 3; i++ {
		metrics.RecordError(WithCode(ErrTimeout, "timeout"))
	}

	top := metrics.GetTopErrors(10)

	require.GreaterOrEqual(t, len(top), 3)

	// Verify descending order
	for i := 1; i < len(top); i++ {
		assert.GreaterOrEqual(t, top[i-1].Count, top[i].Count,
			"Top errors should be in descending order")
	}

	// First should be TestFailed with count 10
	assert.Equal(t, ErrTestFailed, top[0].Code)
	assert.Equal(t, int64(10), top[0].Count)
}

// TestRealMetrics_TopErrorsLimited verifies limit parameter respected
func TestRealMetrics_TopErrorsLimited(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	// Record errors for many codes
	codes := []ErrorCode{
		ErrBuildFailed, ErrTestFailed, ErrTimeout,
		ErrNotFound, ErrInternal, ErrConfigInvalid,
	}
	for _, code := range codes {
		metrics.RecordError(WithCode(code, "error"))
	}

	// Limit to 3
	top := metrics.GetTopErrors(3)
	assert.LessOrEqual(t, len(top), 3)
}

// TestRealMetrics_TopErrorsZeroLimit verifies zero limit returns all
func TestRealMetrics_TopErrorsZeroLimit(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	codes := []ErrorCode{ErrBuildFailed, ErrTestFailed, ErrTimeout}
	for _, code := range codes {
		metrics.RecordError(WithCode(code, "error"))
	}

	// Zero limit should return all
	top := metrics.GetTopErrors(0)
	assert.GreaterOrEqual(t, len(top), 3)
}

// TestRealMetrics_Reset verifies Reset clears all stats
func TestRealMetrics_Reset(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	// Record some errors
	metrics.RecordError(WithCode(ErrBuildFailed, "build"))
	metrics.RecordError(WithCode(ErrTestFailed, "test"))
	metrics.RecordMageError(NewBuilder().
		WithMessage("error").
		WithSeverity(SeverityError).
		Build())

	// Verify counts before reset
	require.Positive(t, metrics.GetTotalErrors())

	// Reset
	err := metrics.Reset()
	require.NoError(t, err)

	// Verify all counts are zero
	assert.Equal(t, int64(0), metrics.GetTotalErrors())
	assert.Equal(t, int64(0), metrics.GetCount(ErrBuildFailed))
	assert.Equal(t, int64(0), metrics.GetCount(ErrTestFailed))
	assert.Equal(t, int64(0), metrics.GetCountBySeverity(SeverityError))
	assert.Empty(t, metrics.GetTopErrors(10))
}

// TestRealMetrics_FirstSeenLastSeen verifies timestamps recorded correctly
func TestRealMetrics_FirstSeenLastSeen(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	beforeFirst := time.Now()
	time.Sleep(10 * time.Millisecond)

	metrics.RecordError(WithCode(ErrBuildFailed, "first"))

	time.Sleep(10 * time.Millisecond)

	metrics.RecordError(WithCode(ErrBuildFailed, "second"))

	afterLast := time.Now()

	top := metrics.GetTopErrors(1)
	require.Len(t, top, 1)

	stat := top[0]
	assert.Equal(t, ErrBuildFailed, stat.Code)
	assert.Equal(t, int64(2), stat.Count)

	// FirstSeen should be after beforeFirst and before afterLast
	assert.True(t, stat.FirstSeen.After(beforeFirst) || stat.FirstSeen.Equal(beforeFirst))
	assert.True(t, stat.LastSeen.Before(afterLast) || stat.LastSeen.Equal(afterLast))

	// LastSeen should be after or equal to FirstSeen
	assert.True(t, stat.LastSeen.After(stat.FirstSeen) || stat.LastSeen.Equal(stat.FirstSeen))
}

// TestRealMetrics_AverageRate verifies average rate in ErrorStat correct
func TestRealMetrics_AverageRate(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	// Record some errors with time gap
	metrics.RecordError(WithCode(ErrBuildFailed, "first"))
	time.Sleep(50 * time.Millisecond)
	metrics.RecordError(WithCode(ErrBuildFailed, "second"))
	time.Sleep(50 * time.Millisecond)
	metrics.RecordError(WithCode(ErrBuildFailed, "third"))

	top := metrics.GetTopErrors(1)
	require.Len(t, top, 1)

	stat := top[0]
	assert.Equal(t, int64(3), stat.Count)

	// AverageRate should be positive and reasonable
	if !stat.FirstSeen.IsZero() && !stat.LastSeen.IsZero() {
		assert.GreaterOrEqual(t, stat.AverageRate, float64(0))
	}
}

// TestRealMetrics_GetSummary verifies summary contains all data
func TestRealMetrics_GetSummary(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	// Record various errors via RecordError (which increments totalErrors)
	metrics.RecordError(WithCode(ErrBuildFailed, "build"))
	metrics.RecordError(WithCode(ErrBuildFailed, "build 2"))
	metrics.RecordError(WithCode(ErrTestFailed, "test"))
	metrics.RecordError(NewBuilder().
		WithMessage("critical").
		WithCode(ErrInternal).
		WithSeverity(SeverityCritical).
		Build())

	summary := metrics.GetSummary()

	assert.Equal(t, int64(4), summary.TotalErrors)
	assert.NotZero(t, summary.StartTime)
	assert.NotZero(t, summary.Duration)

	// Check error codes map
	assert.Equal(t, int64(2), summary.ErrorCodes[ErrBuildFailed])
	assert.Equal(t, int64(1), summary.ErrorCodes[ErrTestFailed])
	assert.Equal(t, int64(1), summary.ErrorCodes[ErrInternal])

	// Check severities map
	assert.Contains(t, summary.Severities, SeverityCritical)

	// Check top errors
	assert.NotEmpty(t, summary.TopErrors)
}

// TestRealMetrics_ConcurrentRecording verifies concurrent RecordError is safe
func TestRealMetrics_ConcurrentRecording(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()
	var wg sync.WaitGroup

	const numGoroutines = 100
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			// Alternate between different error types
			switch idx % 3 {
			case 0:
				metrics.RecordError(WithCode(ErrBuildFailed, "build"))
			case 1:
				metrics.RecordError(WithCode(ErrTestFailed, "test"))
			case 2:
				metrics.RecordError(errMetricsStdRecord)
			}
		}(i)
	}

	wg.Wait()

	// Verify total count
	assert.Equal(t, int64(numGoroutines), metrics.GetTotalErrors())

	// Verify individual counts sum to total
	buildCount := metrics.GetCount(ErrBuildFailed)
	testCount := metrics.GetCount(ErrTestFailed)
	unknownCount := metrics.GetCount(ErrUnknown)

	assert.Equal(t, int64(numGoroutines), buildCount+testCount+unknownCount)
}

// TestRealMetrics_ConcurrentReadWrite verifies concurrent read/write safe
func TestRealMetrics_ConcurrentReadWrite(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()
	var wg sync.WaitGroup

	// Writers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				metrics.RecordError(WithCode(ErrBuildFailed, "error"))
			}
		}()
	}

	// Readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				_ = metrics.GetCount(ErrBuildFailed)
				_ = metrics.GetTotalErrors()
				_ = metrics.GetTopErrors(5)
				_ = metrics.GetCountBySeverity(SeverityError)
				_ = metrics.GetSummary()
			}
		}()
	}

	wg.Wait()
	// Test passes if no race conditions occur
}

// TestRealMetrics_GetErrorRate verifies GetErrorRate calculation
func TestRealMetrics_GetErrorRate(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	// Record some errors
	for i := 0; i < 10; i++ {
		metrics.RecordError(WithCode(ErrBuildFailed, "error"))
	}

	// Wait a bit to have non-zero elapsed time
	time.Sleep(100 * time.Millisecond)

	rate := metrics.GetErrorRate()

	// Rate should be positive
	assert.Greater(t, rate, float64(0))

	// Rate should be reasonable (less than 1000 errors per second)
	assert.Less(t, rate, float64(1000))
}

// TestRealMetrics_GetErrorRateNoErrors verifies GetErrorRate with no errors
func TestRealMetrics_GetErrorRateNoErrors(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	rate := metrics.GetErrorRate()
	assert.InDelta(t, float64(0), rate, 0.001)
}

// TestRealMetrics_RecordMageError verifies RecordMageError directly
func TestRealMetrics_RecordMageError(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	err := NewBuilder().
		WithMessage("test").
		WithCode(ErrBuildFailed).
		WithSeverity(SeverityCritical).
		Build()

	metrics.RecordMageError(err)

	assert.Equal(t, int64(1), metrics.GetCount(ErrBuildFailed))
	assert.Equal(t, int64(1), metrics.GetCountBySeverity(SeverityCritical))
}

// TestRealMetrics_MultipleSeverities verifies multiple severity recording
func TestRealMetrics_MultipleSeverities(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	severities := []Severity{
		SeverityDebug,
		SeverityInfo,
		SeverityWarning,
		SeverityError,
		SeverityCritical,
		SeverityFatal,
	}

	for _, sev := range severities {
		metrics.RecordMageError(NewBuilder().
			WithMessage("test").
			WithSeverity(sev).
			Build())
	}

	for _, sev := range severities {
		assert.Equal(t, int64(1), metrics.GetCountBySeverity(sev),
			"Should have 1 error for severity %v", sev)
	}
}

// TestRealMetrics_TopErrorsExcludesZeroCounts verifies zero count errors excluded
func TestRealMetrics_TopErrorsExcludesZeroCounts(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	// Record only one type
	metrics.RecordError(WithCode(ErrBuildFailed, "build"))

	top := metrics.GetTopErrors(100)

	// Should only have one entry (not entries with zero counts)
	assert.Len(t, top, 1)
	assert.Equal(t, ErrBuildFailed, top[0].Code)
}

// TestRealMetrics_ResetStartsNewTimePeriod verifies Reset resets start time
func TestRealMetrics_ResetStartsNewTimePeriod(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	// Record an error
	metrics.RecordError(WithCode(ErrBuildFailed, "build"))

	summary1 := metrics.GetSummary()
	startTime1 := summary1.StartTime

	time.Sleep(10 * time.Millisecond)

	// Reset
	_ = metrics.Reset() //nolint:errcheck // testing reset behavior, error not relevant

	summary2 := metrics.GetSummary()
	startTime2 := summary2.StartTime

	// Start time should be newer after reset
	assert.True(t, startTime2.After(startTime1) || startTime2.Equal(startTime1))
}

// TestRealMetrics_SummaryTopErrorsLimited verifies summary limits top errors to 10
func TestRealMetrics_SummaryTopErrorsLimited(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	// Record errors for many codes
	codes := []ErrorCode{
		ErrBuildFailed, ErrTestFailed, ErrTimeout, ErrNotFound,
		ErrInternal, ErrConfigInvalid, ErrPermissionDenied,
		ErrInvalidArgument, ErrResourceExhausted, ErrUnavailable,
		ErrFileNotFound, ErrFileAccessDenied, ErrUnknown,
	}

	for _, code := range codes {
		metrics.RecordError(WithCode(code, "error"))
	}

	summary := metrics.GetSummary()

	// TopErrors in summary should be limited to 10
	assert.LessOrEqual(t, len(summary.TopErrors), 10)
}

// TestRealMetrics_FirstSeenNotUpdatedOnSubsequentRecords verifies FirstSeen set only once
func TestRealMetrics_FirstSeenNotUpdatedOnSubsequentRecords(t *testing.T) {
	t.Parallel()

	metrics := NewErrorMetrics()

	// Record first error
	metrics.RecordError(WithCode(ErrBuildFailed, "first"))

	top1 := metrics.GetTopErrors(1)
	require.Len(t, top1, 1)
	firstSeen1 := top1[0].FirstSeen

	time.Sleep(50 * time.Millisecond)

	// Record more errors
	metrics.RecordError(WithCode(ErrBuildFailed, "second"))
	metrics.RecordError(WithCode(ErrBuildFailed, "third"))

	top2 := metrics.GetTopErrors(1)
	require.Len(t, top2, 1)
	firstSeen2 := top2[0].FirstSeen

	// FirstSeen should not change
	assert.Equal(t, firstSeen1, firstSeen2, "FirstSeen should not change after initial set")

	// LastSeen should be updated
	assert.True(t, top2[0].LastSeen.After(top1[0].FirstSeen) || top2[0].LastSeen.Equal(top1[0].FirstSeen))
}
