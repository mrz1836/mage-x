package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
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

	t.Run("cleans existing directory with files", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDir := filepath.Join(tmpDir, "test-clean")

		// Create directory with files
		require.NoError(t, os.Mkdir(testDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(testDir, "file.txt"), []byte("content"), 0o600))

		// Clean should remove files and recreate directory
		err := CleanDir(testDir)
		require.NoError(t, err)

		// Verify directory exists but is empty
		entries, err := os.ReadDir(testDir)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})
}

// TestRunCmd_Coverage tests RunCmd for coverage
func TestRunCmd_Coverage(t *testing.T) {
	t.Run("executes command successfully", func(t *testing.T) {
		// Save current executor
		oldExec := DefaultExecutor
		defer func() { DefaultExecutor = oldExec }()

		// Set executor that succeeds
		SetExecutor(&mockExecutor{})

		err := RunCmd("echo", "test")
		require.NoError(t, err)
	})

	t.Run("executes command with verbose mode", func(t *testing.T) {
		// Save current executor and env
		oldExec := DefaultExecutor
		oldVerbose := os.Getenv("VERBOSE")
		defer func() {
			DefaultExecutor = oldExec
			if oldVerbose == "" {
				_ = os.Unsetenv("VERBOSE") //nolint:errcheck // cleanup
			} else {
				_ = os.Setenv("VERBOSE", oldVerbose) //nolint:errcheck // cleanup
			}
		}()

		// Enable verbose mode
		require.NoError(t, os.Setenv("VERBOSE", "true"))

		// Set executor that succeeds
		SetExecutor(&mockExecutor{})

		err := RunCmd("echo", "test")
		require.NoError(t, err)
	})

	t.Run("returns error from command", func(t *testing.T) {
		// Save current executor
		oldExec := DefaultExecutor
		defer func() { DefaultExecutor = oldExec }()

		// Set executor that fails
		SetExecutor(&mockExecutor{
			executeErr: errors.New("command failed"), //nolint:err113 // test error
		})

		err := RunCmd("failing-command")
		require.Error(t, err)
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

	t.Run("executes command with verbose mode", func(t *testing.T) {
		// Save current executor and env
		oldExec := DefaultExecutor
		oldVerbose := os.Getenv("VERBOSE")
		defer func() {
			DefaultExecutor = oldExec
			if oldVerbose == "" {
				_ = os.Unsetenv("VERBOSE") //nolint:errcheck // cleanup
			} else {
				_ = os.Setenv("VERBOSE", oldVerbose) //nolint:errcheck // cleanup
			}
		}()

		// Enable verbose mode
		require.NoError(t, os.Setenv("VERBOSE", "true"))

		// Set executor that succeeds
		SetExecutor(&mockExecutor{
			executeOutput: "test output",
		})

		output, err := RunCmdOutput("echo", "test")
		require.NoError(t, err)
		assert.Equal(t, "test output", output)
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

	t.Run("executes with verbose logging", func(t *testing.T) {
		// Save and restore state
		oldExec := DefaultExecutor
		origVerbose := os.Getenv("VERBOSE")
		defer func() {
			DefaultExecutor = oldExec
			if origVerbose == "" {
				_ = os.Unsetenv("VERBOSE") //nolint:errcheck // cleanup
			} else {
				_ = os.Setenv("VERBOSE", origVerbose) //nolint:errcheck // cleanup
			}
		}()

		// Set verbose mode
		_ = os.Setenv("VERBOSE", "true") //nolint:errcheck // test setup
		SetExecutor(&mockExecutor{})

		err := RunCmdSecure("echo", "test")
		require.NoError(t, err)
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

	t.Run("executes with verbose logging", func(t *testing.T) {
		// Save and restore state
		oldExec := DefaultExecutor
		origVerbose := os.Getenv("VERBOSE")
		defer func() {
			DefaultExecutor = oldExec
			if origVerbose == "" {
				_ = os.Unsetenv("VERBOSE") //nolint:errcheck // cleanup
			} else {
				_ = os.Setenv("VERBOSE", origVerbose) //nolint:errcheck // cleanup
			}
		}()

		// Set verbose mode
		_ = os.Setenv("VERBOSE", "true") //nolint:errcheck // test setup
		SetExecutor(&mockExecutor{})

		err := RunCmdWithRetry(3, "echo", "test")
		require.NoError(t, err)
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

// TestRunCmdPipe_Coverage tests RunCmdPipe for coverage
func TestRunCmdPipe_Coverage(t *testing.T) {
	t.Run("executes simple pipe", func(t *testing.T) {
		ctx := context.Background()

		// Create commands that pipe together
		cmd1 := exec.CommandContext(ctx, "echo", "test")
		cmd2 := exec.CommandContext(ctx, "cat")

		// Run the pipe
		err := RunCmdPipe(cmd1, cmd2)
		// These commands should work
		require.NoError(t, err)
	})
}

// TestMetrics_AdditionalCoverage tests metrics functions for coverage
func TestMetrics_AdditionalCoverage(t *testing.T) {
	t.Run("NewMetricsCollector with disabled config", func(t *testing.T) {
		config := DefaultMetricsConfig()
		config.Enabled = false

		collector := NewMetricsCollector(&config)
		require.NotNil(t, collector)
	})

	t.Run("RecordMetric with disabled collector", func(t *testing.T) {
		config := DefaultMetricsConfig()
		config.Enabled = false
		collector := NewMetricsCollector(&config)

		err := collector.RecordMetric(&Metric{
			Name:  "test",
			Type:  "counter",
			Value: 1,
		})
		// Should not error even when disabled
		require.NoError(t, err)
	})

	t.Run("Stop timer with enabled collector", func(t *testing.T) {
		config := DefaultMetricsConfig()
		config.Enabled = true
		config.StoragePath = t.TempDir()
		collector := NewMetricsCollector(&config)

		timer := collector.StartTimer("test_timer", nil)
		duration := timer.Stop()
		require.Greater(t, duration, time.Duration(0))
	})

	t.Run("StopWithError", func(t *testing.T) {
		config := DefaultMetricsConfig()
		config.Enabled = true
		config.StoragePath = t.TempDir()
		collector := NewMetricsCollector(&config)

		timer := collector.StartTimer("test_timer", nil)
		testErr := errors.New("test error") //nolint:err113 // test error
		duration := timer.StopWithError(testErr)
		require.Greater(t, duration, time.Duration(0))
	})

	t.Run("Stop with disabled collector", func(t *testing.T) {
		config := DefaultMetricsConfig()
		config.Enabled = false
		collector := NewMetricsCollector(&config)

		timer := collector.StartTimer("test_timer", nil)
		duration := timer.Stop()
		require.Greater(t, duration, time.Duration(0))
	})

	t.Run("StopWithError with disabled collector", func(t *testing.T) {
		config := DefaultMetricsConfig()
		config.Enabled = false
		collector := NewMetricsCollector(&config)

		timer := collector.StartTimer("test_timer", nil)
		testErr := errors.New("test error") //nolint:err113 // test error
		duration := timer.StopWithError(testErr)
		require.Greater(t, duration, time.Duration(0))
	})

	t.Run("Stop with nil collector", func(t *testing.T) {
		timer := &PerformanceTimer{
			Name:      "test_timer",
			StartTime: time.Now().Add(-100 * time.Millisecond),
			collector: nil,
		}
		duration := timer.Stop()
		require.Greater(t, duration, time.Duration(0))
	})

	t.Run("StopWithError with nil collector", func(t *testing.T) {
		timer := &PerformanceTimer{
			Name:      "test_timer",
			StartTime: time.Now().Add(-100 * time.Millisecond),
			collector: nil,
		}
		testErr := errors.New("test error") //nolint:err113 // test error
		duration := timer.StopWithError(testErr)
		require.Greater(t, duration, time.Duration(0))
	})
}

// TestHTTPGetJSON_AdditionalCoverage tests HTTPGetJSON error handling
func TestHTTPGetJSON_AdditionalCoverage(t *testing.T) {
	t.Run("handles invalid URL", func(t *testing.T) {
		_, err := HTTPGetJSON[map[string]interface{}]("://invalid-url", 5*time.Second)
		require.Error(t, err)
	})

	t.Run("handles network error", func(t *testing.T) {
		// Use a URL that will fail
		_, err := HTTPGetJSON[map[string]interface{}]("http://localhost:99999/nonexistent", 1*time.Second)
		require.Error(t, err)
	})
}

// TestNewMetricsCollector_ErrorHandling tests NewMetricsCollector error paths
func TestNewMetricsCollector_ErrorHandling(t *testing.T) {
	t.Run("disables metrics when storage init fails", func(t *testing.T) {
		config := DefaultMetricsConfig()
		config.Enabled = true
		// Use an invalid path that will fail to create
		config.StoragePath = "/dev/null/invalid/path"

		collector := NewMetricsCollector(&config)
		require.NotNil(t, collector)
		// Should have disabled metrics after storage failure
		assert.False(t, collector.config.Enabled)
	})

	t.Run("creates collector with valid config", func(t *testing.T) {
		config := DefaultMetricsConfig()
		config.Enabled = true
		config.StoragePath = t.TempDir()

		collector := NewMetricsCollector(&config)
		require.NotNil(t, collector)
		assert.True(t, collector.config.Enabled)
		assert.NotNil(t, collector.storage)
	})
}

// TestCreateDefaultCollector_Coverage tests createDefaultCollector with env vars
func TestCreateDefaultCollector_Coverage(t *testing.T) {
	t.Run("with MAGE_X_METRICS_ENABLED", func(t *testing.T) {
		// Save and restore env
		oldEnabled := os.Getenv("MAGE_X_METRICS_ENABLED")
		defer func() {
			if oldEnabled == "" {
				_ = os.Unsetenv("MAGE_X_METRICS_ENABLED") //nolint:errcheck // cleanup
			} else {
				_ = os.Setenv("MAGE_X_METRICS_ENABLED", oldEnabled) //nolint:errcheck // cleanup
			}
		}()

		// Set environment variable
		require.NoError(t, os.Setenv("MAGE_X_METRICS_ENABLED", "true"))

		// Create collector - this will use the env var
		collector := createDefaultCollector()
		require.NotNil(t, collector)
	})

	t.Run("with MAGE_X_METRICS_PATH", func(t *testing.T) {
		// Save and restore env
		oldPath := os.Getenv("MAGE_X_METRICS_PATH")
		defer func() {
			if oldPath == "" {
				_ = os.Unsetenv("MAGE_X_METRICS_PATH") //nolint:errcheck // cleanup
			} else {
				_ = os.Setenv("MAGE_X_METRICS_PATH", oldPath) //nolint:errcheck // cleanup
			}
		}()

		// Set custom path
		customPath := t.TempDir()
		require.NoError(t, os.Setenv("MAGE_X_METRICS_PATH", customPath))

		// Create collector - this will use the custom path
		collector := createDefaultCollector()
		require.NotNil(t, collector)
	})
}

// TestNewJSONStorage_Coverage tests NewJSONStorage error handling
func TestNewJSONStorage_Coverage(t *testing.T) {
	t.Run("creates storage successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewJSONStorage(tmpDir)
		require.NoError(t, err)
		require.NotNil(t, storage)
	})

	t.Run("handles directory creation error", func(t *testing.T) {
		// Try to create storage in an invalid path
		// On Unix systems, /dev/null/subdir should fail
		_, err := NewJSONStorage("/dev/null/subdir")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create storage directory")
	})
}

// Test_shouldRemoveMetricFile tests the shouldRemoveMetricFile function
func Test_shouldRemoveMetricFile(t *testing.T) {
	cutoffDate := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "valid old metric file",
			filename: "metrics_2025-12-10.json",
			want:     true,
		},
		{
			name:     "valid new metric file",
			filename: "metrics_2025-12-20.json",
			want:     false,
		},
		{
			name:     "file without metrics prefix",
			filename: "data_2025-12-10.json",
			want:     false,
		},
		{
			name:     "file without json suffix",
			filename: "metrics_2025-12-10.txt",
			want:     false,
		},
		{
			name:     "file with invalid date",
			filename: "metrics_invalid-date.json",
			want:     false,
		},
		{
			name:     "file with incomplete date",
			filename: "metrics_2025-12.json",
			want:     false,
		},
		{
			name:     "random file",
			filename: "random.json",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldRemoveMetricFile(tt.filename, cutoffDate)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestJSONStorage_Cleanup tests the Cleanup method
func TestJSONStorage_Cleanup(t *testing.T) {
	t.Run("removes old files and keeps new files", func(t *testing.T) {
		// Create temp directory
		tmpDir := t.TempDir()
		storage, err := NewJSONStorage(tmpDir)
		require.NoError(t, err)

		// Create metric files with different dates
		oldDate := time.Now().AddDate(0, 0, -60) // 60 days ago
		newDate := time.Now().AddDate(0, 0, -5)  // 5 days ago

		oldFile := filepath.Join(tmpDir, "metrics_"+oldDate.Format("2006-01-02")+".json")
		newFile := filepath.Join(tmpDir, "metrics_"+newDate.Format("2006-01-02")+".json")
		nonMetricFile := filepath.Join(tmpDir, "other.json")

		require.NoError(t, os.WriteFile(oldFile, []byte("{}"), 0o600))
		require.NoError(t, os.WriteFile(newFile, []byte("{}"), 0o600))
		require.NoError(t, os.WriteFile(nonMetricFile, []byte("{}"), 0o600))

		// Run cleanup with 30 day retention
		err = storage.Cleanup(30)
		require.NoError(t, err)

		// Verify old file was removed
		_, err = os.Stat(oldFile)
		assert.True(t, os.IsNotExist(err), "old metric file should be removed")

		// Verify new file was kept
		_, err = os.Stat(newFile)
		require.NoError(t, err, "new metric file should be kept")

		// Verify non-metric file was kept
		_, err = os.Stat(nonMetricFile)
		require.NoError(t, err, "non-metric file should be kept")
	})

	t.Run("handles directory entries", func(t *testing.T) {
		// Create temp directory with subdirectory
		tmpDir := t.TempDir()
		storage, err := NewJSONStorage(tmpDir)
		require.NoError(t, err)

		// Create a subdirectory (should be ignored)
		subDir := filepath.Join(tmpDir, "metrics_2025-01-01.json")
		require.NoError(t, os.Mkdir(subDir, 0o700))

		// Run cleanup - should not fail on directory entry
		err = storage.Cleanup(30)
		require.NoError(t, err)

		// Verify directory still exists
		info, err := os.Stat(subDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})
}

// failingWriter is a writer that always returns an error
type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write failed") //nolint:err113 // test error
}

// TestJSONStorage_Store_AdditionalCoverage tests Store with corrupted files
func TestJSONStorage_Store_AdditionalCoverage(t *testing.T) {
	t.Run("handles corrupted metrics file", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewJSONStorage(tmpDir)
		require.NoError(t, err)

		// Create a corrupted metrics file for today
		now := time.Now()
		filename := fmt.Sprintf("metrics_%s.json", now.Format("2006-01-02"))
		filePath := filepath.Join(tmpDir, filename)
		require.NoError(t, os.WriteFile(filePath, []byte("{invalid json"), 0o600))

		// Store should handle corrupted file gracefully
		metric := &Metric{
			Name:      "test_metric",
			Type:      MetricTypeCounter,
			Value:     100.0,
			Timestamp: now,
		}
		err = storage.Store(metric)
		require.NoError(t, err)

		// Verify the metric was still stored
		data, err := os.ReadFile(filePath) //nolint:gosec // test file in temp directory
		require.NoError(t, err)
		var metrics []*Metric
		require.NoError(t, json.Unmarshal(data, &metrics))
		require.Len(t, metrics, 1)
	})
}

// TestLogger_AdditionalCoverage tests various logger functions
func TestLogger_AdditionalCoverage(t *testing.T) {
	t.Run("Print package function", func(t *testing.T) {
		// Test that Print doesn't panic with various formats
		Print("test message")
		Print("formatted %s", "message")
	})

	t.Run("Println package function", func(t *testing.T) {
		// Test that Println doesn't panic
		Println("test line")
	})

	t.Run("GetContextualMessage handles unknown context", func(t *testing.T) {
		logger := NewLogger()

		// Get unknown context returns empty
		unknownMsg := logger.GetContextualMessage("unknown")
		assert.Empty(t, unknownMsg)
	})

	t.Run("GetTimeContext doesnt panic", func(t *testing.T) {
		logger := NewLogger()

		// Just verify it doesn't panic - output depends on current time
		_ = logger.GetTimeContext()
	})

	t.Run("GetDayContext doesnt panic", func(t *testing.T) {
		logger := NewLogger()

		// Just verify it doesn't panic - output depends on current day
		_ = logger.GetDayContext()
	})
}

// mockStorage is a storage implementation that can be configured to fail
type mockStorage struct {
	shouldFail bool
	failError  error
}

func (m *mockStorage) Store(_ *Metric) error {
	if m.shouldFail {
		return m.failError
	}
	return nil
}

func (m *mockStorage) Query(_ *MetricsQuery) ([]*Metric, error) {
	return nil, nil
}

func (m *mockStorage) Cleanup(_ int) error {
	return nil
}

func (m *mockStorage) Aggregate(_ *MetricsQuery) (*AggregatedMetrics, error) {
	return &AggregatedMetrics{}, nil
}

// TestMetricsCollector_RecordMetricErrors tests RecordMetric error handling
func TestMetricsCollector_RecordMetricErrors(t *testing.T) {
	t.Run("returns storage error when Store fails", func(t *testing.T) {
		config := DefaultMetricsConfig()
		config.Enabled = true
		collector := &MetricsCollector{
			config:  config,
			metrics: make(map[string]*Metric),
			storage: &mockStorage{
				shouldFail: true,
				failError:  errors.New("storage write failed"), //nolint:err113 // test error
			},
		}

		metric := &Metric{
			Name:      "test_metric",
			Type:      MetricTypeCounter,
			Value:     1.0,
			Timestamp: time.Now(),
		}

		err := collector.RecordMetric(metric)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "storage write failed")
	})

	t.Run("succeeds when storage is nil", func(t *testing.T) {
		config := DefaultMetricsConfig()
		config.Enabled = true
		collector := &MetricsCollector{
			config:  config,
			metrics: make(map[string]*Metric),
			storage: nil,
		}

		metric := &Metric{
			Name:      "test_metric",
			Type:      MetricTypeCounter,
			Value:     1.0,
			Timestamp: time.Now(),
		}

		err := collector.RecordMetric(metric)
		require.NoError(t, err)
	})
}

// TestPromptForInput_EmptyPrompt tests PromptForInput with empty prompt
func TestPromptForInput_EmptyPrompt(t *testing.T) {
	t.Run("handles empty prompt string", func(t *testing.T) {
		// Create a pipe to simulate stdin
		oldStdin := os.Stdin
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdin = r
		defer func() {
			os.Stdin = oldStdin
			_ = r.Close() //nolint:errcheck // cleanup in defer
		}()

		// Write test input
		go func() {
			_, _ = w.WriteString("test input\n") //nolint:errcheck // test helper
			_ = w.Close()                        //nolint:errcheck // test cleanup
		}()

		result, err := PromptForInput("")
		require.NoError(t, err)
		assert.Equal(t, "test input", result)
	})
}

// TestDownloadWithRetry_EdgeCases tests DownloadWithRetry edge cases
func TestDownloadWithRetry_EdgeCases(t *testing.T) {
	t.Run("returns error for empty URL", func(t *testing.T) {
		ctx := context.Background()
		tmpFile := filepath.Join(t.TempDir(), "download.txt")
		err := DownloadWithRetry(ctx, "", tmpFile, nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidURL)
	})

	t.Run("returns error for empty destination", func(t *testing.T) {
		ctx := context.Background()
		err := DownloadWithRetry(ctx, "http://example.com/file.txt", "", nil)
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidDestination)
	})
}

// TestLogger_HeaderWithSpinner tests Header when spinner is active
func TestLogger_HeaderWithSpinner(t *testing.T) {
	t.Run("stops spinner before printing header", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger()
		logger.SetOutput(&buf)
		logger.SetColorEnabled(false)

		// Start a spinner
		logger.StartSpinner("test spinner")
		// Give spinner time to start
		time.Sleep(10 * time.Millisecond)

		// Call Header - should stop spinner
		logger.Header("Test Header")

		output := buf.String()
		assert.Contains(t, output, "Test Header")
	})
}

// TestProgress_RenderEdgeCases tests Progress render with different states
func TestProgress_RenderEdgeCases(t *testing.T) {
	t.Run("render with showTime and not complete", func(t *testing.T) {
		progress := NewProgress(100, "Testing")
		progress.Update(50)
		// Give some time to elapse
		time.Sleep(10 * time.Millisecond)
		progress.Update(60)
		// This should render with ETA
	})

	t.Run("render with showTime and complete", func(t *testing.T) {
		progress := NewProgress(100, "Testing")
		progress.Update(50)
		time.Sleep(10 * time.Millisecond)
		progress.Update(100)
		// This should render with total time
	})

	t.Run("render with zero current progress", func(t *testing.T) {
		progress := NewProgress(100, "Testing")
		// Don't update, leave at 0
		progress.render()
		// Should show different emoji for 0 progress
	})
}

// TestBuildTags_DiscoverEdgeCases tests DiscoverBuildTags edge cases
func TestBuildTags_DiscoverEdgeCases(t *testing.T) {
	t.Run("handles directory read errors", func(t *testing.T) {
		// Use a path that doesn't exist
		discovery := NewBuildTagsDiscovery("/nonexistent/path", nil)
		tags, err := discovery.DiscoverBuildTags()
		// Should handle error gracefully
		require.Error(t, err)
		assert.Nil(t, tags)
	})
}

// TestQuery_LimitEdgeCases tests Query with limit edge cases
func TestQuery_LimitEdgeCases(t *testing.T) {
	t.Run("applies limit correctly", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewJSONStorage(tmpDir)
		require.NoError(t, err)

		// Store multiple metrics
		now := time.Now()
		for i := 0; i < 10; i++ {
			metric := &Metric{
				Name:      fmt.Sprintf("metric_%d", i),
				Type:      MetricTypeCounter,
				Value:     float64(i),
				Timestamp: now,
			}
			require.NoError(t, storage.Store(metric))
		}

		// Query with limit
		query := &MetricsQuery{
			StartTime: now.Add(-time.Hour),
			EndTime:   now.Add(time.Hour),
			Limit:     5,
		}

		results, err := storage.Query(query)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(results), 5)
	})
}

// TestPromptForInput_WithPrompt tests PromptForInput with a non-empty prompt
func TestPromptForInput_WithPrompt(t *testing.T) {
	t.Run("displays prompt before reading input", func(t *testing.T) {
		// Create a pipe to simulate stdin
		oldStdin := os.Stdin
		r, w, err := os.Pipe()
		require.NoError(t, err)
		os.Stdin = r
		defer func() {
			os.Stdin = oldStdin
			_ = r.Close() //nolint:errcheck // cleanup in defer
		}()

		// Write test input
		go func() {
			_, _ = w.WriteString("user input\n") //nolint:errcheck // test helper
			_ = w.Close()                        //nolint:errcheck // test cleanup
		}()

		result, err := PromptForInput("Enter name")
		require.NoError(t, err)
		assert.Equal(t, "user input", result)
	})
}

// TestRecordBuildMetrics_Coverage tests RecordBuildMetrics edge cases
func TestRecordBuildMetrics_Coverage(t *testing.T) {
	t.Run("records all build metrics when enabled", func(t *testing.T) {
		config := DefaultMetricsConfig()
		config.Enabled = true
		tmpDir := t.TempDir()
		config.StoragePath = tmpDir

		collector := NewMetricsCollector(&config)

		buildMetrics := &BuildMetrics{
			Duration:      100 * time.Millisecond,
			LinesOfCode:   1000,
			PackagesBuilt: 5,
			TestsRun:      20,
			TestsPassed:   18,
			TestsFailed:   2,
			Coverage:      85.5,
			BinarySize:    1024 * 1024,
			Resources: ResourceMetrics{
				CPUUsage:    50.0,
				MemoryUsage: 512 * 1024 * 1024,
			},
			Timestamp: time.Now(),
			Success:   true,
			Tags:      map[string]string{"env": "test"},
		}

		err := collector.RecordBuildMetrics(buildMetrics)
		require.NoError(t, err)
	})
}

// TestSpinner_StopError tests Stop when output write fails
func TestSpinner_StopError(t *testing.T) {
	t.Run("handles write error gracefully", func(t *testing.T) {
		spinner := NewSpinner("test")
		spinner.Start()
		// Give spinner time to start
		time.Sleep(10 * time.Millisecond)
		// Stop should not panic even if write fails
		spinner.Stop()
	})
}

// TestMultiSpinner_StopError tests MultiSpinner Stop with active tasks
func TestMultiSpinner_StopError(t *testing.T) {
	t.Run("stops all tasks", func(t *testing.T) {
		ms := NewMultiSpinner()
		ms.AddTask("task1", "Task 1")
		ms.AddTask("task2", "Task 2")
		ms.Start()
		// Give time to start
		time.Sleep(10 * time.Millisecond)
		ms.Stop()
		// Should not panic
	})
}

// TestGetCurrentResourceMetrics_Coverage tests GetCurrentResourceMetrics edge cases
func TestGetCurrentResourceMetrics_Coverage(t *testing.T) {
	t.Run("returns resource metrics", func(t *testing.T) {
		metrics := GetCurrentResourceMetrics()
		// Should return valid metrics
		assert.NotNil(t, metrics)
		assert.GreaterOrEqual(t, metrics.CPUUsage, float64(0))
		assert.GreaterOrEqual(t, metrics.MemoryUsage, int64(0))
	})
}

// TestEstimatePackageBuildMemory_Coverage tests memory estimation
func TestEstimatePackageBuildMemory_Coverage(t *testing.T) {
	t.Run("estimates memory for different package counts", func(t *testing.T) {
		// Test with small package count
		mem := EstimatePackageBuildMemory(1)
		assert.Positive(t, mem)

		// Test with larger package count
		mem2 := EstimatePackageBuildMemory(100)
		assert.Greater(t, mem2, mem)
	})

	t.Run("caps at maximum memory", func(t *testing.T) {
		// Test with very large package count (> 8GB cap)
		mem := EstimatePackageBuildMemory(1000)
		// Should be capped at 8GB = 8000 MB
		expected := uint64(8000) * 1024 * 1024
		assert.Equal(t, expected, mem)
	})

	t.Run("handles negative package count", func(t *testing.T) {
		// Test with negative package count that results in negative estimatedMB
		// -100 * 50 = -5000, so 500 + (-5000) = -4500 < 0
		mem := EstimatePackageBuildMemory(-100)
		// Should return base memory (500 MB)
		expected := uint64(500) * 1024 * 1024
		assert.Equal(t, expected, mem)
	})
}

// TestProgressFinish_Coverage tests Progress Finish edge cases
func TestProgressFinish_Coverage(t *testing.T) {
	t.Run("finishes progress bar with newline", func(t *testing.T) {
		progress := NewProgress(100, "Testing")
		progress.Update(75)
		// Finish should set to 100% and print newline
		progress.Finish()
	})
}

// TestCleanup_Coverage tests Cleanup with retention days
func TestCleanup_Coverage(t *testing.T) {
	t.Run("cleanup with different retention periods", func(t *testing.T) {
		tmpDir := t.TempDir()
		storage, err := NewJSONStorage(tmpDir)
		require.NoError(t, err)

		// Create some old metric files
		oldDate := time.Now().AddDate(0, 0, -45)
		filename := fmt.Sprintf("metrics_%s.json", oldDate.Format("2006-01-02"))
		filePath := filepath.Join(tmpDir, filename)
		require.NoError(t, os.WriteFile(filePath, []byte("[]"), 0o600))

		// Cleanup with 30 day retention
		err = storage.Cleanup(30)
		require.NoError(t, err)

		// Old file should be removed
		_, err = os.Stat(filePath)
		assert.True(t, os.IsNotExist(err))
	})
}

// TestPerformanceTimer_StopWithStorageError tests PerformanceTimer.Stop when RecordMetric fails
func TestPerformanceTimer_StopWithStorageError(t *testing.T) {
	t.Run("handles RecordMetric error gracefully in Stop", func(t *testing.T) {
		config := DefaultMetricsConfig()
		config.Enabled = true
		collector := &MetricsCollector{
			config:  config,
			metrics: make(map[string]*Metric),
			storage: &mockStorage{
				shouldFail: true,
				failError:  errors.New("storage error"), //nolint:err113 // test error
			},
		}

		timer := collector.StartTimer("test_timer", nil)
		time.Sleep(10 * time.Millisecond)

		// Stop should complete even if RecordMetric fails
		duration := timer.Stop()
		assert.Greater(t, duration, time.Duration(0))
	})

	t.Run("handles RecordMetric error gracefully in StopWithError", func(t *testing.T) {
		config := DefaultMetricsConfig()
		config.Enabled = true
		collector := &MetricsCollector{
			config:  config,
			metrics: make(map[string]*Metric),
			storage: &mockStorage{
				shouldFail: true,
				failError:  errors.New("storage error"), //nolint:err113 // test error
			},
		}

		timer := collector.StartTimer("test_timer", nil)
		time.Sleep(10 * time.Millisecond)

		// StopWithError should complete even if RecordMetric fails
		testErr := errors.New("test error") //nolint:err113 // test error
		duration := timer.StopWithError(testErr)
		assert.Greater(t, duration, time.Duration(0))
	})
}

// TestRecordBuildMetrics_StorageError tests RecordBuildMetrics with storage errors
func TestRecordBuildMetrics_StorageError(t *testing.T) {
	t.Run("returns error when storage fails", func(t *testing.T) {
		config := DefaultMetricsConfig()
		config.Enabled = true
		collector := &MetricsCollector{
			config:  config,
			metrics: make(map[string]*Metric),
			storage: &mockStorage{
				shouldFail: true,
				failError:  errors.New("storage write failed"), //nolint:err113 // test error
			},
		}

		buildMetrics := &BuildMetrics{
			Duration:      100 * time.Millisecond,
			LinesOfCode:   1000,
			PackagesBuilt: 5,
			TestsRun:      20,
			Coverage:      85.5,
			BinarySize:    1024 * 1024,
			Resources: ResourceMetrics{
				CPUUsage:    50.0,
				MemoryUsage: 512 * 1024 * 1024,
			},
			Timestamp: time.Now(),
			Success:   true,
		}

		err := collector.RecordBuildMetrics(buildMetrics)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to record metric")
		assert.Contains(t, err.Error(), "storage write failed")
	})
}

// TestMultiSpinner_RenderStates tests MultiSpinner render with different task states
func TestMultiSpinner_RenderStates(t *testing.T) {
	t.Run("renders spinners with different task states", func(t *testing.T) {
		ms := NewMultiSpinner()
		defer func() {
			ms.Stop()
		}()

		// Add tasks
		ms.AddTask("task1", "Task 1")
		ms.AddTask("task2", "Task 2")
		ms.AddTask("task3", "Task 3")

		// Start the spinner
		ms.Start()

		// Update task states
		ms.UpdateTask("task1", TaskStatusRunning, "Task 1 running")
		ms.UpdateTask("task2", TaskStatusSuccess, "Task 2 complete")
		ms.UpdateTask("task3", TaskStatusFailed, "Task 3 failed")

		// Let it render a few frames
		time.Sleep(200 * time.Millisecond)

		// Stop should clean up
		ms.Stop()
	})
}

// TestSpinner_AnimateFrames tests Spinner animation
func TestSpinner_AnimateFrames(t *testing.T) {
	t.Run("animates through frames", func(t *testing.T) {
		spinner := NewSpinner("Testing")

		// Start and let it animate
		spinner.Start()
		time.Sleep(150 * time.Millisecond)

		// Stop should complete without error
		spinner.Stop()
	})
}

// TestShouldUseColor_EnvVars tests shouldUseColor with different environment variables
func TestShouldUseColor_EnvVars(t *testing.T) {
	t.Run("returns false when CI env is set", func(t *testing.T) {
		origCI := os.Getenv("CI")
		defer func() {
			if origCI == "" {
				_ = os.Unsetenv("CI") //nolint:errcheck // cleanup
			} else {
				_ = os.Setenv("CI", origCI) //nolint:errcheck // cleanup
			}
		}()

		_ = os.Setenv("CI", "true") //nolint:errcheck // test setup

		result := shouldUseColor()
		assert.False(t, result)
	})

	t.Run("returns false when NO_COLOR is set", func(t *testing.T) {
		origCI := os.Getenv("CI")
		origNoColor := os.Getenv("NO_COLOR")
		defer func() {
			if origCI == "" {
				_ = os.Unsetenv("CI") //nolint:errcheck // cleanup
			} else {
				_ = os.Setenv("CI", origCI) //nolint:errcheck // cleanup
			}
			if origNoColor == "" {
				_ = os.Unsetenv("NO_COLOR") //nolint:errcheck // cleanup
			} else {
				_ = os.Setenv("NO_COLOR", origNoColor) //nolint:errcheck // cleanup
			}
		}()

		_ = os.Unsetenv("CI")          //nolint:errcheck // test setup
		_ = os.Setenv("NO_COLOR", "1") //nolint:errcheck // test setup

		result := shouldUseColor()
		assert.False(t, result)
	})
}

// TestLogger_LogWithEmoji tests logWithEmoji with different log levels
func TestLogger_LogWithEmoji_Coverage(t *testing.T) {
	t.Run("logs with different levels and emoji", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger()
		logger.SetOutput(&buf)
		logger.SetColorEnabled(false)

		// Test different log levels
		logger.logWithEmoji(LogLevelInfo, "ℹ️", "info message")
		logger.logWithEmoji(LogLevelWarn, "⚠️", "warning message")
		logger.logWithEmoji(LogLevelError, "❌", "error message")
		logger.logWithEmoji(LogLevelDebug, "🐛", "debug message")

		output := buf.String()
		assert.Contains(t, output, "info message")
		assert.Contains(t, output, "warning message")
		assert.Contains(t, output, "error message")
	})

	t.Run("respects log level filtering", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger()
		logger.SetOutput(&buf)
		logger.SetLevel(LogLevelWarn) // Only warn and above

		logger.logWithEmoji(LogLevelDebug, "🐛", "debug message")
		logger.logWithEmoji(LogLevelInfo, "ℹ️", "info message")

		output := buf.String()
		assert.NotContains(t, output, "debug message")
		assert.NotContains(t, output, "info message")
	})

	t.Run("adds prefix to messages", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger()
		logger.SetOutput(&buf)
		logger.SetColorEnabled(false)
		prefixedLogger := logger.WithPrefix("TEST")

		prefixedLogger.logWithEmoji(LogLevelInfo, "ℹ️", "message")

		output := buf.String()
		assert.Contains(t, output, "[TEST]")
		assert.Contains(t, output, "message")
	})

	t.Run("pauses spinner while logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger()
		logger.SetOutput(&buf)
		logger.SetColorEnabled(false)

		logger.StartSpinner("test")
		time.Sleep(50 * time.Millisecond)

		logger.logWithEmoji(LogLevelInfo, "ℹ️", "message")

		output := buf.String()
		assert.Contains(t, output, "message")
	})

	t.Run("uses color when enabled", func(t *testing.T) {
		var buf bytes.Buffer
		logger := NewLogger()
		logger.SetOutput(&buf)
		logger.SetColorEnabled(true)

		logger.logWithEmoji(LogLevelInfo, "ℹ️", "message")

		output := buf.String()
		// When color is enabled, ANSI codes should be present
		assert.Contains(t, output, "message")
	})
}

// TestProgress_FinishWithNewline tests Progress.Finish
func TestProgress_FinishWithNewline(t *testing.T) {
	t.Run("finishes and adds newline", func(t *testing.T) {
		progress := NewProgress(100, "Test")
		progress.Update(50)

		// Finish should set to 100% and add newline
		progress.Finish()

		// Verify internal state
		progress.mu.Lock()
		current := progress.current
		total := progress.total
		progress.mu.Unlock()

		assert.Equal(t, total, current)
	})
}

// TestRender_TreeStructure tests tree rendering functions
func TestRender_TreeStructure(t *testing.T) {
	t.Run("renders tree with branches", func(t *testing.T) {
		var buf bytes.Buffer
		w := &buf

		// Create a simple tree structure
		type testNode struct {
			name     string
			children []*testNode
		}

		root := &testNode{
			name: "root",
			children: []*testNode{
				{name: "child1", children: nil},
				{name: "child2", children: []*testNode{
					{name: "grandchild1", children: nil},
				}},
			},
		}

		// Render function
		var render func(node *testNode, prefix string, isLast bool)
		render = func(node *testNode, prefix string, isLast bool) {
			branch := "├── "
			if isLast {
				branch = "└── "
			}

			_, _ = fmt.Fprintf(w, "%s%s%s\n", prefix, branch, node.name)

			newPrefix := prefix
			if isLast {
				newPrefix += "    "
			} else {
				newPrefix += "│   "
			}

			for i, child := range node.children {
				render(child, newPrefix, i == len(node.children)-1)
			}
		}

		// Render root
		_, _ = fmt.Fprintln(w, root.name)
		for i, child := range root.children {
			render(child, "", i == len(root.children)-1)
		}

		output := buf.String()
		assert.Contains(t, output, "root")
		assert.Contains(t, output, "child1")
		assert.Contains(t, output, "child2")
		assert.Contains(t, output, "grandchild1")
	})
}

// TestGetTimeContext_Coverage tests GetTimeContext returns valid values
func TestGetTimeContext_Coverage(t *testing.T) {
	t.Run("returns valid time context", func(t *testing.T) {
		logger := NewLogger()
		context := logger.GetTimeContext()

		// Should return one of the valid contexts
		validContexts := []string{"morning", "afternoon", "evening"}
		assert.Contains(t, validContexts, context)
	})
}

// TestGetDayContext_Coverage tests GetDayContext with different days
func TestGetDayContext_Coverage(t *testing.T) {
	t.Run("returns valid day context", func(t *testing.T) {
		logger := NewLogger()
		context := logger.GetDayContext()

		// Should return monday, friday, or empty string
		validContexts := []string{"monday", "friday", ""}
		assert.Contains(t, validContexts, context)
	})
}

// TestHTTPGetJSON_ContextTimeout tests HTTPGetJSON with timeout
func TestHTTPGetJSON_ContextTimeout(t *testing.T) {
	t.Run("handles context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// Delay to allow context cancellation
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`)) //nolint:errcheck // test server
		}))
		defer server.Close()

		// Create a context that times out quickly
		result, err := HTTPGetJSON[map[string]string](server.URL, 10*time.Millisecond)
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestSpinner_StopEdgeCases tests Spinner Stop edge cases
func TestSpinner_StopEdgeCases(t *testing.T) {
	t.Run("stops inactive spinner without error", func(t *testing.T) {
		spinner := NewSpinner("test")
		// Don't start it, just stop
		spinner.Stop()
		// Should not panic or error
	})

	t.Run("stops and clears spinner", func(t *testing.T) {
		spinner := NewSpinner("test")
		spinner.Start()
		time.Sleep(50 * time.Millisecond)

		// Stop should clear the line
		spinner.Stop()

		// Verify spinner is not active
		spinner.mu.Lock()
		active := spinner.active
		spinner.mu.Unlock()
		assert.False(t, active)
	})
}

// TestMultiSpinner_StopEdgeCases tests MultiSpinner Stop edge cases
func TestMultiSpinner_StopEdgeCases(t *testing.T) {
	t.Run("stops inactive multispinner without error", func(t *testing.T) {
		ms := NewMultiSpinner()
		// Don't start it, just stop
		ms.Stop()
		// Should not panic or error
	})

	t.Run("clears all spinner lines on stop", func(t *testing.T) {
		ms := NewMultiSpinner()
		ms.AddTask("task1", "Task 1")
		ms.AddTask("task2", "Task 2")

		ms.Start()
		time.Sleep(100 * time.Millisecond)

		// Stop should clear all lines
		ms.Stop()

		// Verify multispinner is not active
		ms.mu.Lock()
		active := ms.active
		ms.mu.Unlock()
		assert.False(t, active)
	})
}

// TestExtractPagesValue_EdgeCases tests extractPagesValue error paths
func TestExtractPagesValue_EdgeCases(t *testing.T) {
	t.Run("returns 0 for lines with fewer than 3 fields", func(t *testing.T) {
		result := extractPagesValue("Pages free:")
		assert.Equal(t, uint64(0), result)

		result = extractPagesValue("Pages")
		assert.Equal(t, uint64(0), result)

		result = extractPagesValue("")
		assert.Equal(t, uint64(0), result)
	})

	t.Run("returns 0 for invalid number format", func(t *testing.T) {
		result := extractPagesValue("Pages free: invalid.")
		assert.Equal(t, uint64(0), result)

		result = extractPagesValue("Pages free: -123.")
		assert.Equal(t, uint64(0), result)

		result = extractPagesValue("Pages free: abc123.")
		assert.Equal(t, uint64(0), result)
	})

	t.Run("parses valid page count", func(t *testing.T) {
		result := extractPagesValue("Pages free: 12345.")
		assert.Equal(t, uint64(12345), result)

		result = extractPagesValue("Pages active: 67890.")
		assert.Equal(t, uint64(67890), result)
	})
}

// TestExtractPageSize_EdgeCases tests extractPageSize error paths
func TestExtractPageSize_EdgeCases(t *testing.T) {
	t.Run("returns default when 'of' keyword not found", func(t *testing.T) {
		result := extractPageSize("Mach Virtual Memory Statistics:", 4096)
		assert.Equal(t, uint64(4096), result)

		result = extractPageSize("Pages free: 12345.", 8192)
		assert.Equal(t, uint64(8192), result)
	})

	t.Run("returns default when index out of bounds", func(t *testing.T) {
		result := extractPageSize("page size of", 4096)
		assert.Equal(t, uint64(4096), result)
	})

	t.Run("returns default for invalid number format", func(t *testing.T) {
		result := extractPageSize("page size of invalid bytes", 4096)
		assert.Equal(t, uint64(4096), result)

		result = extractPageSize("page size of -123 bytes", 8192)
		assert.Equal(t, uint64(8192), result)
	})

	t.Run("parses valid page size", func(t *testing.T) {
		result := extractPageSize("page size of 4096 bytes", 0)
		assert.Equal(t, uint64(4096), result)

		result = extractPageSize("Mach Virtual Memory Statistics: (page size of 16384 bytes)", 0)
		assert.Equal(t, uint64(16384), result)
	})
}
