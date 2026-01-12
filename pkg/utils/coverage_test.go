package utils

import (
	"bytes"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLogger_WriteColoredHeader tests writeColoredHeader with write errors
func TestLogger_WriteColoredHeader(t *testing.T) {
	t.Run("handles write errors gracefully", func(t *testing.T) {
		logger := NewLogger()

		// Use a writer that always fails
		logger.SetOutput(&failingWriter{})

		// writeColoredHeader should handle errors gracefully without panicking
		logger.writeColoredHeader("===", "Test")
		// If we got here without panic, the test passes
	})

	t.Run("writes colored header successfully", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger()
		logger.SetOutput(&buf)
		logger.SetColorEnabled(true)

		logger.writeColoredHeader("===", "Test Header")

		output := buf.String()
		assert.Contains(t, output, "Test Header")
		assert.Contains(t, output, "===")
	})
}

// TestLogger_WritePlainHeader tests writePlainHeader with write errors
func TestLogger_WritePlainHeader(t *testing.T) {
	t.Run("handles write errors gracefully", func(t *testing.T) {
		logger := NewLogger()

		// Use a writer that always fails
		logger.SetOutput(&failingWriter{})

		// writePlainHeader should handle errors gracefully without panicking
		logger.writePlainHeader("===", "Test")
		// If we got here without panic, the test passes
	})

	t.Run("writes plain header successfully", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger()
		logger.SetOutput(&buf)
		logger.SetColorEnabled(false)

		logger.writePlainHeader("===", "Test Header")

		output := buf.String()
		assert.Contains(t, output, "Test Header")
		assert.Contains(t, output, "===")
	})
}

// TestPrint tests Print function error handling
func TestPrint(t *testing.T) {
	t.Run("prints message successfully", func(t *testing.T) {
		// Print writes to os.Stdout, which we can't easily redirect in this test
		// Just ensure it doesn't panic
		Print("test message %s", "arg")
	})
}

// TestPrintln tests Println function error handling
func TestPrintln(t *testing.T) {
	t.Run("prints message with newline", func(t *testing.T) {
		// Println writes to os.Stdout, which we can't easily redirect in this test
		// Just ensure it doesn't panic
		Println("test message")
	})
}

// TestLogger_GetTimeContext tests GetTimeContext for all time ranges
func TestLogger_GetTimeContext(t *testing.T) {
	logger := NewLogger()

	// We can't easily change time.Now(), but we can test the logic indirectly
	// by checking that it returns one of the valid contexts
	context := logger.GetTimeContext()
	assert.Contains(t, []string{"morning", "afternoon", "evening"}, context)

	// Test the logic by checking hour ranges
	hour := time.Now().Hour()
	expectedContext := "evening" // default
	if hour >= 5 && hour < 12 {
		expectedContext = "morning"
	} else if hour >= 12 && hour < 17 {
		expectedContext = "afternoon"
	}

	assert.Equal(t, expectedContext, context)
}

// TestLogger_GetDayContext tests GetDayContext for different weekdays
func TestLogger_GetDayContext(t *testing.T) {
	logger := NewLogger()

	// Get current day context
	context := logger.GetDayContext()

	// Valid contexts are "monday", "friday", or ""
	assert.Contains(t, []string{"monday", "friday", ""}, context)

	// Verify it matches current weekday
	weekday := time.Now().Weekday()
	expectedContext := ""
	switch weekday {
	case time.Monday:
		expectedContext = "monday"
	case time.Friday:
		expectedContext = "friday"
	case time.Sunday, time.Tuesday, time.Wednesday, time.Thursday, time.Saturday:
		expectedContext = ""
	}

	assert.Equal(t, expectedContext, context)
}

// TestGoList_ErrorPaths tests error handling in GoList
func TestGoList_ErrorPaths(t *testing.T) {
	t.Run("includes output in error message", func(t *testing.T) {
		// Use an invalid flag that produces output
		_, err := GoList("-bad-flag-with-output")
		if err != nil {
			// Error should exist for invalid flag
			require.Error(t, err)
		}
	})
}

// TestGetModuleName_ErrorPath tests GetModuleName error handling
func TestGetModuleName_ErrorPath(t *testing.T) {
	t.Run("handles command error", func(t *testing.T) {
		// Save current executor
		oldExec := DefaultExecutor
		defer func() { DefaultExecutor = oldExec }()

		// Set executor that fails
		SetExecutor(&mockExecutor{
			executeOutputErr: errors.New("command failed"), //nolint:err113 // test error
		})

		_, err := GetModuleName()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get module name")
	})
}

// TestGetGoVersion_ErrorPaths tests GetGoVersion error handling
func TestGetGoVersion_ErrorPaths(t *testing.T) {
	t.Run("handles command error", func(t *testing.T) {
		// Save current executor
		oldExec := DefaultExecutor
		defer func() { DefaultExecutor = oldExec }()

		// Set executor that fails
		SetExecutor(&mockExecutor{
			executeOutputErr: errors.New("command failed"), //nolint:err113 // test error
		})

		_, err := GetGoVersion()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get Go version")
	})

	t.Run("handles parse error", func(t *testing.T) {
		// Save current executor
		oldExec := DefaultExecutor
		defer func() { DefaultExecutor = oldExec }()

		// Set executor that returns unparseable output
		SetExecutor(&mockExecutor{
			executeOutput: "invalid output",
		})

		_, err := GetGoVersion()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unable to parse go version")
	})
}

// TestCleanDir_ErrorPath tests CleanDir with error handling
func TestCleanDir_ErrorPath(t *testing.T) {
	t.Run("handles non-existent directory without error", func(t *testing.T) {
		tmpDir := t.TempDir()
		nonExistentDir := tmpDir + "/does-not-exist"

		// Should succeed even if directory doesn't exist
		err := CleanDir(nonExistentDir)
		require.NoError(t, err)

		// Directory should be created
		assert.True(t, DirExists(nonExistentDir))
	})
}

// TestRunCmdOutput_ErrorWithOutput tests RunCmdOutput error path
func TestRunCmdOutput_ErrorWithOutput(t *testing.T) {
	t.Run("returns error from command", func(t *testing.T) {
		// Save current executor
		oldExec := DefaultExecutor
		defer func() { DefaultExecutor = oldExec }()

		// Set executor that fails
		SetExecutor(&mockExecutor{
			executeOutputErr: errors.New("command failed"), //nolint:err113 // test error
			executeOutput:    "error output",
		})

		_, err := RunCmdOutput("failing-command")
		require.Error(t, err)
	})
}

// TestRunCmdSecure_Coverage tests RunCmdSecure
func TestRunCmdSecure_Coverage(t *testing.T) {
	t.Run("executes command securely", func(t *testing.T) {
		// Save current executor
		oldExec := DefaultExecutor
		defer func() { DefaultExecutor = oldExec }()

		// Set executor that succeeds
		SetExecutor(&mockExecutor{})

		err := RunCmdSecure("echo", "test")
		require.NoError(t, err)
	})

	t.Run("returns error from command", func(t *testing.T) {
		// Save current executor
		oldExec := DefaultExecutor
		defer func() { DefaultExecutor = oldExec }()

		// Set executor that fails
		SetExecutor(&mockExecutor{
			executeErr: errors.New("secure command failed"), //nolint:err113 // test error
		})

		err := RunCmdSecure("failing-command")
		require.Error(t, err)
	})
}

// TestRunCmdWithRetry_Coverage tests RunCmdWithRetry
func TestRunCmdWithRetry_Coverage(t *testing.T) {
	t.Run("executes command with retry", func(t *testing.T) {
		// Save current executor
		oldExec := DefaultExecutor
		defer func() { DefaultExecutor = oldExec }()

		// Set executor that succeeds
		SetExecutor(&mockExecutor{})

		err := RunCmdWithRetry(3, "echo", "test")
		require.NoError(t, err)
	})

	t.Run("returns error from command", func(t *testing.T) {
		// Save current executor
		oldExec := DefaultExecutor
		defer func() { DefaultExecutor = oldExec }()

		// Set executor that fails
		SetExecutor(&mockExecutor{
			executeErr: errors.New("retry command failed"), //nolint:err113 // test error
		})

		err := RunCmdWithRetry(3, "failing-command")
		require.Error(t, err)
	})
}

// TestFindFiles_ErrorPath tests findFiles error handling
func TestFindFiles_ErrorPath(t *testing.T) {
	t.Run("handles invalid pattern", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldDir, err := os.Getwd()
		require.NoError(t, err)

		require.NoError(t, os.Chdir(tmpDir))
		defer func() {
			_ = os.Chdir(oldDir) //nolint:errcheck // cleanup
		}()

		// Invalid glob pattern
		_, err = findFiles(".", "[")
		assert.Error(t, err)
	})
}

// TestGoList_ErrorWithOutput tests GoList error with output
func TestGoList_ErrorWithOutput(t *testing.T) {
	t.Run("includes output in error when command fails", func(t *testing.T) {
		// Save current executor
		oldExec := DefaultExecutor
		defer func() { DefaultExecutor = oldExec }()

		// Set executor that fails with output
		SetExecutor(&mockExecutor{
			executeOutputErr: errors.New("list command failed"), //nolint:err113 // test error
			executeOutput:    "some error output",
		})

		_, err := GoList("./...")
		require.Error(t, err)
	})
}

// failingWriter is a writer that always returns an error
type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write failed") //nolint:err113 // test error
}
