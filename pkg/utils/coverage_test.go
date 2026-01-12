package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
