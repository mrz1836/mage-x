package mage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCIReporter is a test double that tracks method calls and can return configured errors
type mockCIReporter struct {
	startCalls         []CIMetadata
	reportFailureCalls []CITestFailure
	writeSummaryCalls  []*CIResult
	closeCalled        bool

	startErr         error
	reportFailureErr error
	writeSummaryErr  error
	closeErr         error
}

func newMockCIReporter() *mockCIReporter {
	return &mockCIReporter{}
}

func (m *mockCIReporter) Start(metadata CIMetadata) error {
	m.startCalls = append(m.startCalls, metadata)
	return m.startErr
}

func (m *mockCIReporter) ReportFailure(failure CITestFailure) error {
	m.reportFailureCalls = append(m.reportFailureCalls, failure)
	return m.reportFailureErr
}

func (m *mockCIReporter) WriteSummary(result *CIResult) error {
	m.writeSummaryCalls = append(m.writeSummaryCalls, result)
	return m.writeSummaryErr
}

func (m *mockCIReporter) Close() error {
	m.closeCalled = true
	return m.closeErr
}

// Test errors
var (
	errMockStart         = errors.New("mock start error")
	errMockReportFailure = errors.New("mock report failure error")
	errMockWriteSummary  = errors.New("mock write summary error")
	errMockClose         = errors.New("mock close error")
	errMockClose2        = errors.New("mock close error 2")
)

// TestNewMultiReporter tests the constructor
func TestNewMultiReporter(t *testing.T) {
	t.Parallel()

	t.Run("creates reporter with no reporters", func(t *testing.T) {
		t.Parallel()
		mr := NewMultiReporter()
		require.NotNil(t, mr)
		assert.Empty(t, mr.reporters)
	})

	t.Run("creates reporter with single reporter", func(t *testing.T) {
		t.Parallel()
		mock := newMockCIReporter()
		mr := NewMultiReporter(mock)
		require.NotNil(t, mr)
		assert.Len(t, mr.reporters, 1)
	})

	t.Run("creates reporter with multiple reporters", func(t *testing.T) {
		t.Parallel()
		mock1 := newMockCIReporter()
		mock2 := newMockCIReporter()
		mock3 := newMockCIReporter()
		mr := NewMultiReporter(mock1, mock2, mock3)
		require.NotNil(t, mr)
		assert.Len(t, mr.reporters, 3)
	})
}

// TestMultiReporter_Start tests the Start method
func TestMultiReporter_Start(t *testing.T) {
	t.Parallel()

	metadata := CIMetadata{
		Branch:    "main",
		Commit:    "abc123",
		Platform:  "linux",
		GoVersion: "1.21",
	}

	t.Run("success with no reporters", func(t *testing.T) {
		t.Parallel()
		mr := NewMultiReporter()
		err := mr.Start(metadata)
		assert.NoError(t, err)
	})

	t.Run("success with single reporter", func(t *testing.T) {
		t.Parallel()
		mock := newMockCIReporter()
		mr := NewMultiReporter(mock)

		err := mr.Start(metadata)

		require.NoError(t, err)
		require.Len(t, mock.startCalls, 1)
		assert.Equal(t, metadata, mock.startCalls[0])
	})

	t.Run("success with multiple reporters", func(t *testing.T) {
		t.Parallel()
		mock1 := newMockCIReporter()
		mock2 := newMockCIReporter()
		mock3 := newMockCIReporter()
		mr := NewMultiReporter(mock1, mock2, mock3)

		err := mr.Start(metadata)

		require.NoError(t, err)
		assert.Len(t, mock1.startCalls, 1)
		assert.Len(t, mock2.startCalls, 1)
		assert.Len(t, mock3.startCalls, 1)
	})

	t.Run("error from first reporter stops iteration", func(t *testing.T) {
		t.Parallel()
		mock1 := newMockCIReporter()
		mock1.startErr = errMockStart
		mock2 := newMockCIReporter()
		mr := NewMultiReporter(mock1, mock2)

		err := mr.Start(metadata)

		require.ErrorIs(t, err, errMockStart)
		assert.Len(t, mock1.startCalls, 1)
		assert.Empty(t, mock2.startCalls) // Second reporter not called
	})

	t.Run("error from middle reporter stops iteration", func(t *testing.T) {
		t.Parallel()
		mock1 := newMockCIReporter()
		mock2 := newMockCIReporter()
		mock2.startErr = errMockStart
		mock3 := newMockCIReporter()
		mr := NewMultiReporter(mock1, mock2, mock3)

		err := mr.Start(metadata)

		require.ErrorIs(t, err, errMockStart)
		assert.Len(t, mock1.startCalls, 1)
		assert.Len(t, mock2.startCalls, 1)
		assert.Empty(t, mock3.startCalls) // Third reporter not called
	})
}

// TestMultiReporter_ReportFailure tests the ReportFailure method
func TestMultiReporter_ReportFailure(t *testing.T) {
	t.Parallel()

	failure := CITestFailure{
		Package: "github.com/test/pkg",
		Test:    "TestExample",
		Output:  "test failed",
	}

	t.Run("success with no reporters", func(t *testing.T) {
		t.Parallel()
		mr := NewMultiReporter()
		err := mr.ReportFailure(failure)
		assert.NoError(t, err)
	})

	t.Run("success with single reporter", func(t *testing.T) {
		t.Parallel()
		mock := newMockCIReporter()
		mr := NewMultiReporter(mock)

		err := mr.ReportFailure(failure)

		require.NoError(t, err)
		require.Len(t, mock.reportFailureCalls, 1)
		assert.Equal(t, failure, mock.reportFailureCalls[0])
	})

	t.Run("success with multiple reporters", func(t *testing.T) {
		t.Parallel()
		mock1 := newMockCIReporter()
		mock2 := newMockCIReporter()
		mr := NewMultiReporter(mock1, mock2)

		err := mr.ReportFailure(failure)

		require.NoError(t, err)
		assert.Len(t, mock1.reportFailureCalls, 1)
		assert.Len(t, mock2.reportFailureCalls, 1)
	})

	t.Run("error from first reporter stops iteration", func(t *testing.T) {
		t.Parallel()
		mock1 := newMockCIReporter()
		mock1.reportFailureErr = errMockReportFailure
		mock2 := newMockCIReporter()
		mr := NewMultiReporter(mock1, mock2)

		err := mr.ReportFailure(failure)

		require.ErrorIs(t, err, errMockReportFailure)
		assert.Len(t, mock1.reportFailureCalls, 1)
		assert.Empty(t, mock2.reportFailureCalls)
	})

	t.Run("multiple failures reported correctly", func(t *testing.T) {
		t.Parallel()
		mock := newMockCIReporter()
		mr := NewMultiReporter(mock)

		failure1 := CITestFailure{Test: "Test1"}
		failure2 := CITestFailure{Test: "Test2"}

		require.NoError(t, mr.ReportFailure(failure1))
		require.NoError(t, mr.ReportFailure(failure2))

		require.Len(t, mock.reportFailureCalls, 2)
		assert.Equal(t, "Test1", mock.reportFailureCalls[0].Test)
		assert.Equal(t, "Test2", mock.reportFailureCalls[1].Test)
	})
}

// TestMultiReporter_WriteSummary tests the WriteSummary method
func TestMultiReporter_WriteSummary(t *testing.T) {
	t.Parallel()

	result := &CIResult{
		Summary: CISummary{
			Total:  10,
			Passed: 8,
			Failed: 2,
		},
	}

	t.Run("success with no reporters", func(t *testing.T) {
		t.Parallel()
		mr := NewMultiReporter()
		err := mr.WriteSummary(result)
		assert.NoError(t, err)
	})

	t.Run("success with single reporter", func(t *testing.T) {
		t.Parallel()
		mock := newMockCIReporter()
		mr := NewMultiReporter(mock)

		err := mr.WriteSummary(result)

		require.NoError(t, err)
		require.Len(t, mock.writeSummaryCalls, 1)
		assert.Equal(t, result, mock.writeSummaryCalls[0])
	})

	t.Run("success with multiple reporters", func(t *testing.T) {
		t.Parallel()
		mock1 := newMockCIReporter()
		mock2 := newMockCIReporter()
		mr := NewMultiReporter(mock1, mock2)

		err := mr.WriteSummary(result)

		require.NoError(t, err)
		assert.Len(t, mock1.writeSummaryCalls, 1)
		assert.Len(t, mock2.writeSummaryCalls, 1)
	})

	t.Run("error from first reporter stops iteration", func(t *testing.T) {
		t.Parallel()
		mock1 := newMockCIReporter()
		mock1.writeSummaryErr = errMockWriteSummary
		mock2 := newMockCIReporter()
		mr := NewMultiReporter(mock1, mock2)

		err := mr.WriteSummary(result)

		require.ErrorIs(t, err, errMockWriteSummary)
		assert.Len(t, mock1.writeSummaryCalls, 1)
		assert.Empty(t, mock2.writeSummaryCalls)
	})

	t.Run("nil result handled correctly", func(t *testing.T) {
		t.Parallel()
		mock := newMockCIReporter()
		mr := NewMultiReporter(mock)

		err := mr.WriteSummary(nil)

		require.NoError(t, err)
		require.Len(t, mock.writeSummaryCalls, 1)
		assert.Nil(t, mock.writeSummaryCalls[0])
	})
}

// TestMultiReporter_Close tests the Close method
func TestMultiReporter_Close(t *testing.T) {
	t.Parallel()

	t.Run("success with no reporters", func(t *testing.T) {
		t.Parallel()
		mr := NewMultiReporter()
		err := mr.Close()
		assert.NoError(t, err)
	})

	t.Run("success with single reporter", func(t *testing.T) {
		t.Parallel()
		mock := newMockCIReporter()
		mr := NewMultiReporter(mock)

		err := mr.Close()

		require.NoError(t, err)
		assert.True(t, mock.closeCalled)
	})

	t.Run("success with multiple reporters", func(t *testing.T) {
		t.Parallel()
		mock1 := newMockCIReporter()
		mock2 := newMockCIReporter()
		mock3 := newMockCIReporter()
		mr := NewMultiReporter(mock1, mock2, mock3)

		err := mr.Close()

		require.NoError(t, err)
		assert.True(t, mock1.closeCalled)
		assert.True(t, mock2.closeCalled)
		assert.True(t, mock3.closeCalled)
	})

	t.Run("error from first reporter continues to others", func(t *testing.T) {
		t.Parallel()
		mock1 := newMockCIReporter()
		mock1.closeErr = errMockClose
		mock2 := newMockCIReporter()
		mock3 := newMockCIReporter()
		mr := NewMultiReporter(mock1, mock2, mock3)

		err := mr.Close()

		// Returns the error but continues calling other reporters
		require.ErrorIs(t, err, errMockClose)
		assert.True(t, mock1.closeCalled)
		assert.True(t, mock2.closeCalled) // Second reporter still called
		assert.True(t, mock3.closeCalled) // Third reporter still called
	})

	t.Run("returns last error when multiple errors occur", func(t *testing.T) {
		t.Parallel()
		mock1 := newMockCIReporter()
		mock1.closeErr = errMockClose
		mock2 := newMockCIReporter()
		mock2.closeErr = errMockClose2
		mock3 := newMockCIReporter()
		mr := NewMultiReporter(mock1, mock2, mock3)

		err := mr.Close()

		// Should return the last error (from mock2)
		require.ErrorIs(t, err, errMockClose2)
		assert.True(t, mock1.closeCalled)
		assert.True(t, mock2.closeCalled)
		assert.True(t, mock3.closeCalled)
	})

	t.Run("returns last error from last reporter", func(t *testing.T) {
		t.Parallel()
		mock1 := newMockCIReporter()
		mock1.closeErr = errMockClose
		mock2 := newMockCIReporter()
		mock3 := newMockCIReporter()
		mock3.closeErr = errMockClose2
		mr := NewMultiReporter(mock1, mock2, mock3)

		err := mr.Close()

		// Should return the last error (from mock3)
		require.ErrorIs(t, err, errMockClose2)
	})
}

// TestNullReporter_AllMethods tests the NullReporter implementation
func TestNullReporter_AllMethods(t *testing.T) {
	t.Parallel()

	t.Run("Start returns nil", func(t *testing.T) {
		t.Parallel()
		nr := NullReporter{}
		err := nr.Start(CIMetadata{Branch: "main", Platform: "linux", GoVersion: "1.21"})
		assert.NoError(t, err)
	})

	t.Run("ReportFailure returns nil", func(t *testing.T) {
		t.Parallel()
		nr := NullReporter{}
		err := nr.ReportFailure(CITestFailure{Test: "Test"})
		assert.NoError(t, err)
	})

	t.Run("WriteSummary returns nil", func(t *testing.T) {
		t.Parallel()
		nr := NullReporter{}
		err := nr.WriteSummary(&CIResult{Summary: CISummary{Total: 10}})
		assert.NoError(t, err)
	})

	t.Run("WriteSummary with nil result returns nil", func(t *testing.T) {
		t.Parallel()
		nr := NullReporter{}
		err := nr.WriteSummary(nil)
		assert.NoError(t, err)
	})

	t.Run("Close returns nil", func(t *testing.T) {
		t.Parallel()
		nr := NullReporter{}
		err := nr.Close()
		assert.NoError(t, err)
	})
}

// TestCIReporter_InterfaceCompliance verifies that types implement the CIReporter interface
func TestCIReporter_InterfaceCompliance(t *testing.T) {
	t.Parallel()

	t.Run("MultiReporter implements CIReporter", func(t *testing.T) {
		t.Parallel()
		var _ CIReporter = (*MultiReporter)(nil)
	})

	t.Run("NullReporter implements CIReporter", func(t *testing.T) {
		t.Parallel()
		var _ CIReporter = NullReporter{}
	})

	t.Run("mockCIReporter implements CIReporter", func(t *testing.T) {
		t.Parallel()
		var _ CIReporter = (*mockCIReporter)(nil)
	})
}

// TestMultiReporter_FullWorkflow tests a complete workflow through all methods
func TestMultiReporter_FullWorkflow(t *testing.T) {
	t.Parallel()

	mock1 := newMockCIReporter()
	mock2 := newMockCIReporter()
	mr := NewMultiReporter(mock1, mock2)

	// Start
	metadata := CIMetadata{Branch: "main", Platform: "linux", GoVersion: "1.21"}
	require.NoError(t, mr.Start(metadata))

	// Report multiple failures
	failure1 := CITestFailure{Test: "Test1", Package: "pkg1"}
	failure2 := CITestFailure{Test: "Test2", Package: "pkg2"}
	require.NoError(t, mr.ReportFailure(failure1))
	require.NoError(t, mr.ReportFailure(failure2))

	// Write summary
	result := &CIResult{Summary: CISummary{Total: 10, Failed: 2, Passed: 8}}
	require.NoError(t, mr.WriteSummary(result))

	// Close
	require.NoError(t, mr.Close())

	// Verify all calls were made to both reporters
	for _, mock := range []*mockCIReporter{mock1, mock2} {
		assert.Len(t, mock.startCalls, 1)
		assert.Len(t, mock.reportFailureCalls, 2)
		assert.Len(t, mock.writeSummaryCalls, 1)
		assert.True(t, mock.closeCalled)
	}
}

// TestMultiReporter_WithNullReporter tests MultiReporter containing NullReporter
func TestMultiReporter_WithNullReporter(t *testing.T) {
	t.Parallel()

	mock := newMockCIReporter()
	null := NullReporter{}
	mr := NewMultiReporter(mock, null)

	metadata := CIMetadata{Branch: "main", Platform: "linux", GoVersion: "1.21"}
	require.NoError(t, mr.Start(metadata))
	require.NoError(t, mr.ReportFailure(CITestFailure{Test: "Test"}))
	require.NoError(t, mr.WriteSummary(&CIResult{}))
	require.NoError(t, mr.Close())

	// Mock should have received all calls
	assert.Len(t, mock.startCalls, 1)
	assert.Len(t, mock.reportFailureCalls, 1)
	assert.Len(t, mock.writeSummaryCalls, 1)
	assert.True(t, mock.closeCalled)
}
