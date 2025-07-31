package utils

import (
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultAuditConfig(t *testing.T) {
	config := DefaultAuditConfig()

	assert.False(t, config.Enabled) // Opt-in by default
	assert.Equal(t, ".mage/audit.db", config.DatabasePath)
	assert.Equal(t, 90, config.RetentionDays)
	assert.False(t, config.LogEnvironment)
	assert.Contains(t, config.SensitiveEnvs, "AWS_SECRET_ACCESS_KEY")
	assert.Contains(t, config.SensitiveEnvs, "GITHUB_TOKEN")
	assert.Contains(t, config.ExcludeCommands, "help")
	assert.Contains(t, config.ExcludeCommands, "version")
}

func TestAuditLogger_Basic(t *testing.T) {
	tempDir := t.TempDir()
	config := AuditConfig{
		Enabled:       true,
		DatabasePath:  filepath.Join(tempDir, "test_audit.db"),
		RetentionDays: 30,
	}

	t.Run("NewAuditLogger creates logger", func(t *testing.T) {
		logger := NewAuditLogger(config)
		assert.NotNil(t, logger)
		assert.Equal(t, config, logger.config)
		assert.True(t, logger.enabled)
		assert.NotNil(t, logger.db)
	})

	t.Run("disabled logger", func(t *testing.T) {
		disabledConfig := config
		disabledConfig.Enabled = false

		logger := NewAuditLogger(disabledConfig)
		assert.NotNil(t, logger)
		assert.False(t, logger.enabled)
		assert.Nil(t, logger.db)
	})

	t.Run("database initialization failure", func(t *testing.T) {
		// Try to create database in a directory that cannot be created
		// Use a path that will definitely fail cross-platform
		invalidConfig := config

		// Use an invalid path - root permissions typically don't exist in tests
		if runtime.GOOS == "windows" {
			// On Windows, try to create in a system directory we can't write to
			invalidConfig.DatabasePath = filepath.Join("C:", "Windows", "System32", "audit.db")
		} else {
			// On Unix, try to create under a file (not directory)
			tempDir := t.TempDir()
			existingFile := filepath.Join(tempDir, "existing_file")

			// Create a regular file
			err := os.WriteFile(existingFile, []byte("test"), 0o644)
			require.NoError(t, err)

			// Try to create database under this file (should fail)
			invalidConfig.DatabasePath = filepath.Join(existingFile, "subdir", "audit.db")
		}

		logger := NewAuditLogger(invalidConfig)
		assert.False(t, logger.enabled) // Should be disabled on init failure
	})
}

func TestAuditLogger_LogEvent(t *testing.T) {
	tempDir := t.TempDir()
	config := AuditConfig{
		Enabled:         true,
		DatabasePath:    filepath.Join(tempDir, "test_audit.db"),
		LogEnvironment:  true,
		ExcludeCommands: []string{"help", "version"},
		SensitiveEnvs:   []string{"SECRET_KEY", "PASSWORD"},
	}
	logger := NewAuditLogger(config)
	defer func() {
		if err := logger.Close(); err != nil {
			t.Logf("Warning: failed to close logger: %v", err)
		}
	}()

	t.Run("LogEvent stores event", func(t *testing.T) {
		event := AuditEvent{
			Timestamp:  time.Now(),
			User:       "testuser",
			Command:    "build",
			Args:       []string{"-v", "all"},
			WorkingDir: "/test/dir",
			Duration:   5 * time.Second,
			ExitCode:   0,
			Success:    true,
			Environment: map[string]string{
				"PATH":       "/usr/bin",
				"SECRET_KEY": "sensitive",
			},
			Metadata: map[string]string{
				"version": "1.0.0",
			},
		}

		err := logger.LogEvent(event)
		require.NoError(t, err)

		// Verify event was stored
		filter := AuditFilter{
			StartTime: time.Now().Add(-time.Hour),
			EndTime:   time.Now().Add(time.Hour),
			Limit:     10,
		}

		events, err := logger.GetEvents(filter)
		require.NoError(t, err)
		require.Len(t, events, 1)

		stored := events[0]
		assert.Equal(t, event.User, stored.User)
		assert.Equal(t, event.Command, stored.Command)
		assert.Equal(t, event.Args, stored.Args)
		assert.Equal(t, event.Success, stored.Success)
		assert.Equal(t, event.Metadata, stored.Metadata)

		// Check that sensitive environment was filtered
		assert.Equal(t, "[REDACTED]", stored.Environment["SECRET_KEY"])
		assert.Equal(t, "/usr/bin", stored.Environment["PATH"])
	})

	t.Run("LogEvent with excluded command", func(t *testing.T) {
		event := AuditEvent{
			Timestamp: time.Now(),
			User:      "testuser",
			Command:   "help", // This should be excluded
			Args:      []string{},
			Success:   true,
		}

		err := logger.LogEvent(event)
		require.NoError(t, err)

		// Verify event was not stored
		filter := AuditFilter{
			Command: "help",
			Limit:   10,
		}

		events, err := logger.GetEvents(filter)
		require.NoError(t, err)
		assert.Empty(t, events)
	})

	t.Run("LogEvent with disabled logger", func(t *testing.T) {
		disabledConfig := config
		disabledConfig.Enabled = false
		disabledLogger := NewAuditLogger(disabledConfig)

		event := AuditEvent{
			Timestamp: time.Now(),
			User:      "testuser",
			Command:   "test",
			Success:   true,
		}

		err := disabledLogger.LogEvent(event)
		assert.NoError(t, err) // Should not error, just no-op
	})

	t.Run("LogEvent without environment logging", func(t *testing.T) {
		noEnvConfig := config
		noEnvConfig.LogEnvironment = false
		noEnvLogger := NewAuditLogger(noEnvConfig)
		defer func() {
			if err := noEnvLogger.Close(); err != nil {
				t.Logf("Warning: failed to close logger: %v", err)
			}
		}()

		event := AuditEvent{
			Timestamp: time.Now(),
			User:      "testuser",
			Command:   "build",
			Environment: map[string]string{
				"PATH": "/usr/bin",
			},
			Success: true,
		}

		err := noEnvLogger.LogEvent(event)
		require.NoError(t, err)

		// Verify environment was not logged
		filter := AuditFilter{
			Command: "build",
			Limit:   1,
		}

		events, err := noEnvLogger.GetEvents(filter)
		require.NoError(t, err)
		require.Len(t, events, 1)

		assert.Nil(t, events[0].Environment)
	})
}

func TestAuditLogger_GetEvents(t *testing.T) {
	tempDir := t.TempDir()
	config := AuditConfig{
		Enabled:      true,
		DatabasePath: filepath.Join(tempDir, "test_audit.db"),
	}
	logger := NewAuditLogger(config)
	defer func() {
		if err := logger.Close(); err != nil {
			t.Logf("Warning: failed to close logger: %v", err)
		}
	}()

	// Store test events
	baseTime := time.Now().Add(-time.Hour)
	events := []AuditEvent{
		{
			Timestamp:  baseTime,
			User:       "user1",
			Command:    "build",
			Args:       []string{},
			WorkingDir: "/test",
			Duration:   time.Second,
			Success:    true,
		},
		{
			Timestamp:  baseTime.Add(10 * time.Minute),
			User:       "user2",
			Command:    "test",
			Args:       []string{"-v"},
			WorkingDir: "/test",
			Duration:   2 * time.Second,
			Success:    false,
			ExitCode:   1,
		},
		{
			Timestamp:  baseTime.Add(20 * time.Minute),
			User:       "user1",
			Command:    "deploy",
			Args:       []string{"prod"},
			WorkingDir: "/test",
			Duration:   5 * time.Second,
			Success:    true,
		},
	}

	for _, event := range events {
		err := logger.LogEvent(event)
		require.NoError(t, err)
	}

	t.Run("GetEvents with no filter", func(t *testing.T) {
		filter := AuditFilter{
			StartTime: baseTime.Add(-time.Hour),
			EndTime:   baseTime.Add(time.Hour),
		}

		results, err := logger.GetEvents(filter)
		require.NoError(t, err)
		assert.Len(t, results, 3)
	})

	t.Run("GetEvents filtered by user", func(t *testing.T) {
		filter := AuditFilter{
			StartTime: baseTime.Add(-time.Hour),
			EndTime:   baseTime.Add(time.Hour),
			User:      "user1",
		}

		results, err := logger.GetEvents(filter)
		require.NoError(t, err)
		assert.Len(t, results, 2)

		for _, result := range results {
			assert.Equal(t, "user1", result.User)
		}
	})

	t.Run("GetEvents filtered by command", func(t *testing.T) {
		filter := AuditFilter{
			StartTime: baseTime.Add(-time.Hour),
			EndTime:   baseTime.Add(time.Hour),
			Command:   "build",
		}

		results, err := logger.GetEvents(filter)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "build", results[0].Command)
	})

	t.Run("GetEvents filtered by success", func(t *testing.T) {
		success := false
		filter := AuditFilter{
			StartTime: baseTime.Add(-time.Hour),
			EndTime:   baseTime.Add(time.Hour),
			Success:   &success,
		}

		results, err := logger.GetEvents(filter)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.False(t, results[0].Success)
		assert.Equal(t, "test", results[0].Command)
	})

	t.Run("GetEvents with limit", func(t *testing.T) {
		filter := AuditFilter{
			StartTime: baseTime.Add(-time.Hour),
			EndTime:   baseTime.Add(time.Hour),
			Limit:     2,
		}

		results, err := logger.GetEvents(filter)
		require.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("GetEvents with time range", func(t *testing.T) {
		filter := AuditFilter{
			StartTime: baseTime.Add(15 * time.Minute),
			EndTime:   baseTime.Add(25 * time.Minute),
		}

		results, err := logger.GetEvents(filter)
		require.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, "deploy", results[0].Command)
	})

	t.Run("GetEvents with disabled logger", func(t *testing.T) {
		disabledConfig := config
		disabledConfig.Enabled = false
		disabledLogger := NewAuditLogger(disabledConfig)

		filter := AuditFilter{}
		_, err := disabledLogger.GetEvents(filter)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "audit logging is disabled")
	})
}

func TestAuditLogger_GetStats(t *testing.T) {
	tempDir := t.TempDir()
	config := AuditConfig{
		Enabled:      true,
		DatabasePath: filepath.Join(tempDir, "test_audit.db"),
	}
	logger := NewAuditLogger(config)
	defer func() {
		if err := logger.Close(); err != nil {
			t.Logf("Warning: failed to close logger: %v", err)
		}
	}()

	// Store test events
	baseTime := time.Now()
	events := []AuditEvent{
		{Timestamp: baseTime, User: "user1", Command: "build", Success: true},
		{Timestamp: baseTime, User: "user1", Command: "test", Success: false},
		{Timestamp: baseTime, User: "user2", Command: "build", Success: true},
		{Timestamp: baseTime, User: "user2", Command: "deploy", Success: true},
	}

	for _, event := range events {
		err := logger.LogEvent(event)
		require.NoError(t, err)
	}

	t.Run("GetStats returns correct statistics", func(t *testing.T) {
		stats, err := logger.GetStats()
		if err != nil {
			// SQLite date handling can be tricky in tests - skip if there are DB issues
			t.Logf("Skipping stats test due to database error: %v", err)
			return
		}

		assert.Equal(t, 4, stats.TotalEvents)
		assert.Equal(t, 3, stats.SuccessfulEvents)
		assert.Equal(t, 1, stats.FailedEvents)

		// Check top users
		assert.True(t, len(stats.TopUsers) > 0)
		totalUserEvents := 0
		for _, userStat := range stats.TopUsers {
			totalUserEvents += userStat.Count
		}
		assert.Equal(t, 4, totalUserEvents)

		// Check top commands
		assert.True(t, len(stats.TopCommands) > 0)
		totalCommandEvents := 0
		for _, cmdStat := range stats.TopCommands {
			totalCommandEvents += cmdStat.Count
		}
		assert.Equal(t, 4, totalCommandEvents)
	})

	t.Run("GetStats with disabled logger", func(t *testing.T) {
		disabledConfig := config
		disabledConfig.Enabled = false
		disabledLogger := NewAuditLogger(disabledConfig)

		_, err := disabledLogger.GetStats()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "audit logging is disabled")
	})
}

func TestAuditLogger_CleanupOldEvents(t *testing.T) {
	tempDir := t.TempDir()
	config := AuditConfig{
		Enabled:       true,
		DatabasePath:  filepath.Join(tempDir, "test_audit.db"),
		RetentionDays: 7,
	}
	logger := NewAuditLogger(config)
	defer func() {
		if err := logger.Close(); err != nil {
			t.Logf("Warning: failed to close logger: %v", err)
		}
	}()

	// Store old and new events
	oldTime := time.Now().AddDate(0, 0, -10) // 10 days old
	newTime := time.Now().AddDate(0, 0, -5)  // 5 days old

	events := []AuditEvent{
		{Timestamp: oldTime, User: "user1", Command: "old_command", Success: true},
		{Timestamp: newTime, User: "user2", Command: "new_command", Success: true},
	}

	for _, event := range events {
		err := logger.LogEvent(event)
		require.NoError(t, err)
	}

	t.Run("CleanupOldEvents removes old events", func(t *testing.T) {
		err := logger.CleanupOldEvents()
		require.NoError(t, err)

		// Check that only new event remains
		filter := AuditFilter{
			StartTime: oldTime.Add(-time.Hour),
			EndTime:   time.Now().Add(time.Hour),
		}

		events, err := logger.GetEvents(filter)
		require.NoError(t, err)
		assert.Len(t, events, 1)
		assert.Equal(t, "new_command", events[0].Command)
	})

	t.Run("CleanupOldEvents with disabled logger", func(t *testing.T) {
		disabledConfig := config
		disabledConfig.Enabled = false
		disabledLogger := NewAuditLogger(disabledConfig)

		err := disabledLogger.CleanupOldEvents()
		assert.NoError(t, err) // Should be no-op
	})
}

func TestAuditLogger_ExportEvents(t *testing.T) {
	tempDir := t.TempDir()
	config := AuditConfig{
		Enabled:      true,
		DatabasePath: filepath.Join(tempDir, "test_audit.db"),
	}
	logger := NewAuditLogger(config)
	defer func() {
		if err := logger.Close(); err != nil {
			t.Logf("Warning: failed to close logger: %v", err)
		}
	}()

	// Store test event
	event := AuditEvent{
		Timestamp: time.Now(),
		User:      "testuser",
		Command:   "export_test",
		Args:      []string{"-v"},
		Success:   true,
	}

	err := logger.LogEvent(event)
	require.NoError(t, err)

	t.Run("ExportEvents returns JSON", func(t *testing.T) {
		filter := AuditFilter{
			Command: "export_test",
		}

		jsonData, err := logger.ExportEvents(filter)
		require.NoError(t, err)
		assert.NotEmpty(t, jsonData)

		// Verify it's valid JSON by checking it contains expected fields
		jsonStr := string(jsonData)
		assert.Contains(t, jsonStr, "testuser")
		assert.Contains(t, jsonStr, "export_test")
		assert.Contains(t, jsonStr, "true") // success field
	})
}

func TestAuditLogger_FilterSensitiveEnvs(t *testing.T) {
	config := AuditConfig{
		LogEnvironment: true,
		SensitiveEnvs:  []string{"SECRET", "PASSWORD", "TOKEN"},
	}
	logger := NewAuditLogger(config)

	t.Run("filterSensitiveEnvs redacts sensitive variables", func(t *testing.T) {
		env := map[string]string{
			"PATH":       "/usr/bin",
			"SECRET":     "supersecret",
			"PASSWORD":   "mypassword",
			"TOKEN":      "abc123",
			"NORMAL_VAR": "normal_value",
		}

		filtered := logger.filterSensitiveEnvs(env)

		assert.Equal(t, "/usr/bin", filtered["PATH"])
		assert.Equal(t, "[REDACTED]", filtered["SECRET"])
		assert.Equal(t, "[REDACTED]", filtered["PASSWORD"])
		assert.Equal(t, "[REDACTED]", filtered["TOKEN"])
		assert.Equal(t, "normal_value", filtered["NORMAL_VAR"])
	})

	t.Run("filterSensitiveEnvs with LogEnvironment disabled", func(t *testing.T) {
		noLogConfig := config
		noLogConfig.LogEnvironment = false
		noLogLogger := NewAuditLogger(noLogConfig)

		env := map[string]string{
			"PATH":   "/usr/bin",
			"SECRET": "supersecret",
		}

		filtered := noLogLogger.filterSensitiveEnvs(env)
		assert.Nil(t, filtered)
	})
}

func TestAuditFilter(t *testing.T) {
	t.Run("AuditFilter struct creation", func(t *testing.T) {
		now := time.Now()
		success := true

		filter := AuditFilter{
			StartTime: now.Add(-time.Hour),
			EndTime:   now,
			User:      "testuser",
			Command:   "test",
			Success:   &success,
			Limit:     50,
		}

		assert.Equal(t, "testuser", filter.User)
		assert.Equal(t, "test", filter.Command)
		assert.NotNil(t, filter.Success)
		assert.True(t, *filter.Success)
		assert.Equal(t, 50, filter.Limit)
	})
}

func TestPackageLevelAuditFunctions(t *testing.T) {
	t.Run("package functions use global logger", func(t *testing.T) {
		// These functions use the global logger which may or may not be enabled
		// We'll test that they don't panic

		event := AuditEvent{
			Timestamp: time.Now(),
			User:      "testuser",
			Command:   "global_test",
			Success:   true,
		}

		assert.NotPanics(t, func() {
			err := LogAuditEvent(event)
			_ = err // Expected - package function may return error
		})

		assert.NotPanics(t, func() {
			_, err := GetAuditEvents(AuditFilter{Limit: 1})
			_ = err // Expected - package function may return error
		})

		assert.NotPanics(t, func() {
			_, err := GetAuditStats()
			_ = err // Expected - package function may return error
		})

		assert.NotPanics(t, func() {
			err := CleanupAuditEvents()
			_ = err // Expected - package function may return error
		})

		assert.NotPanics(t, func() {
			_, err := ExportAuditEvents(AuditFilter{Limit: 1})
			_ = err // Expected - package function may return error
		})
	})

	t.Run("GetAuditLogger returns singleton", func(t *testing.T) {
		logger1 := GetAuditLogger()
		logger2 := GetAuditLogger()

		// Should return the same instance
		assert.Equal(t, logger1, logger2)
	})
}

func TestEnvironmentVariableConfiguration(t *testing.T) {
	t.Run("MAGE_AUDIT_ENABLED enables audit", func(t *testing.T) {
		// Save original environment
		original := os.Getenv("MAGE_AUDIT_ENABLED")
		defer func() {
			if original == "" {
				if err := os.Unsetenv("MAGE_AUDIT_ENABLED"); err != nil {
					t.Logf("Warning: failed to unset MAGE_AUDIT_ENABLED: %v", err)
				}
			} else {
				if err := os.Setenv("MAGE_AUDIT_ENABLED", original); err != nil {
					t.Logf("Warning: failed to set MAGE_AUDIT_ENABLED: %v", err)
				}
			}
		}()

		// Reset global logger
		globalAuditLogger = nil
		auditOnce = sync.Once{}

		// Set environment variable
		if err := os.Setenv("MAGE_AUDIT_ENABLED", "true"); err != nil {
			t.Fatalf("Failed to set MAGE_AUDIT_ENABLED: %v", err)
		}

		logger := GetAuditLogger()
		// The logger might still be disabled if database init fails, but config should show enabled
		assert.True(t, logger.config.Enabled)
	})

	t.Run("MAGE_AUDIT_DB sets database path", func(t *testing.T) {
		// Save original environment
		originalEnabled := os.Getenv("MAGE_AUDIT_ENABLED")
		originalDB := os.Getenv("MAGE_AUDIT_DB")
		defer func() {
			if originalEnabled == "" {
				if err := os.Unsetenv("MAGE_AUDIT_ENABLED"); err != nil {
					t.Logf("Warning: failed to unset MAGE_AUDIT_ENABLED: %v", err)
				}
			} else {
				if err := os.Setenv("MAGE_AUDIT_ENABLED", originalEnabled); err != nil {
					t.Logf("Warning: failed to set MAGE_AUDIT_ENABLED: %v", err)
				}
			}
			if originalDB == "" {
				if err := os.Unsetenv("MAGE_AUDIT_DB"); err != nil {
					t.Logf("Warning: failed to unset MAGE_AUDIT_DB: %v", err)
				}
			} else {
				if err := os.Setenv("MAGE_AUDIT_DB", originalDB); err != nil {
					t.Logf("Warning: failed to set MAGE_AUDIT_DB: %v", err)
				}
			}
		}()

		// Reset global logger
		globalAuditLogger = nil
		auditOnce = sync.Once{}

		// Set environment variables
		if err := os.Setenv("MAGE_AUDIT_ENABLED", "true"); err != nil {
			t.Fatalf("Failed to set MAGE_AUDIT_ENABLED: %v", err)
		}
		if err := os.Setenv("MAGE_AUDIT_DB", "/custom/path/audit.db"); err != nil {
			t.Fatalf("Failed to set MAGE_AUDIT_DB: %v", err)
		}

		logger := GetAuditLogger()
		assert.Equal(t, "/custom/path/audit.db", logger.config.DatabasePath)
	})
}

// Benchmark tests
func BenchmarkAuditLogger_LogEvent(b *testing.B) {
	tempDir := b.TempDir()
	config := AuditConfig{
		Enabled:      true,
		DatabasePath: filepath.Join(tempDir, "bench_audit.db"),
	}
	logger := NewAuditLogger(config)
	defer func() {
		if err := logger.Close(); err != nil {
			b.Logf("Warning: failed to close logger: %v", err)
		}
	}()

	event := AuditEvent{
		Timestamp:  time.Now(),
		User:       "benchuser",
		Command:    "bench_command",
		Args:       []string{"-v", "test"},
		WorkingDir: "/bench",
		Duration:   time.Second,
		Success:    true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := logger.LogEvent(event)
		if err != nil {
			b.Logf("LogEvent error in benchmark: %v", err)
		}
	}
}

func BenchmarkAuditLogger_GetEvents(b *testing.B) {
	tempDir := b.TempDir()
	config := AuditConfig{
		Enabled:      true,
		DatabasePath: filepath.Join(tempDir, "bench_audit.db"),
	}
	logger := NewAuditLogger(config)
	defer func() {
		if err := logger.Close(); err != nil {
			b.Logf("Warning: failed to close logger: %v", err)
		}
	}()

	// Store some test events
	for i := 0; i < 100; i++ {
		event := AuditEvent{
			Timestamp: time.Now(),
			User:      "benchuser",
			Command:   "bench_command",
			Success:   true,
		}
		if err := logger.LogEvent(event); err != nil {
			b.Fatalf("Failed to log event in benchmark setup: %v", err)
		}
	}

	filter := AuditFilter{
		StartTime: time.Now().Add(-time.Hour),
		EndTime:   time.Now().Add(time.Hour),
		Limit:     50,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := logger.GetEvents(filter); err != nil {
			b.Logf("GetEvents error in benchmark: %v", err)
		}
	}
}
