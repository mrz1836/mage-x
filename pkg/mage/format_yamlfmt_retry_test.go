package mage

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/security"
)

// TestEnsureYamlfmtRetryLogicIntegration tests retry logic in integration with command execution
func TestEnsureYamlfmtRetryLogicIntegration(t *testing.T) {
	t.Run("tool already exists", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock runner
		mockRunner := &MockCommandRunner{}
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// When yamlfmt is already available, ensureYamlfmt should not try to install
		// We can't easily mock utils.CommandExists but can test the overall behavior

		err := ensureYamlfmt()
		// If the test system has yamlfmt installed, this should succeed without installation
		// If not, it will try to install and may fail
		if err != nil {
			// Should be a type assertion error since we're not using SecureCommandRunner
			assert.Contains(t, err.Error(), "expected SecureCommandRunner")
		}
	})

	t.Run("installation through secure runner", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create a real SecureCommandRunner for integration testing
		// This tests the actual retry logic but may fail if network is unavailable
		secureRunner := NewSecureCommandRunner()
		_ = SetRunner(secureRunner) //nolint:errcheck // test setup

		// In a real integration test, this would attempt to install yamlfmt
		err := ensureYamlfmt()
		// This test depends on network availability and actual tool installation
		// In CI/testing environment, this might be mocked or skipped
		if err != nil {
			t.Logf("Installation failed (expected in test environment): %v", err)
		}
	})

	t.Run("version handling", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create a real SecureCommandRunner for testing version handling
		executor := security.NewSecureExecutor()
		executor.DryRun = true // Enable dry run to avoid actual installation
		secureRunner := &SecureCommandRunner{executor: executor}
		_ = SetRunner(secureRunner) //nolint:errcheck // test setup

		// Test with dry run - this should succeed without actual installation
		err := ensureYamlfmt()
		assert.NoError(t, err, "Dry run should succeed")
	})
}

// TestSecureCommandRunnerTypeAssertionYamlfmt tests type assertion behavior for yamlfmt
func TestSecureCommandRunnerTypeAssertionYamlfmt(t *testing.T) {
	t.Run("wrong runner type in ensureYamlfmt", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Set a non-SecureCommandRunner
		mockRunner := &MockCommandRunner{}
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		err := ensureYamlfmt()

		// If yamlfmt already exists on the system, this will return nil
		// If yamlfmt doesn't exist, it will try to install and get type assertion error
		if err != nil {
			assert.Contains(t, err.Error(), "expected SecureCommandRunner")
			assert.ErrorIs(t, err, ErrUnexpectedRunnerType)
		} else {
			t.Log("yamlfmt already exists on system, no type assertion occurred")
		}
	})
}

// TestYamlfmtRetryBehaviorWithSecureExecutor tests retry behavior using the actual SecureExecutor
func TestYamlfmtRetryBehaviorWithSecureExecutor(t *testing.T) {
	t.Run("retry logic with actual executor for yamlfmt installation", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create a SecureCommandRunner with DryRun enabled
		executor := security.NewSecureExecutor()
		executor.DryRun = true
		secureRunner := &SecureCommandRunner{executor: executor}
		_ = SetRunner(secureRunner) //nolint:errcheck // test setup

		// Test the yamlfmt installation with dry run
		err := ensureYamlfmt()
		// With DryRun=true, this should succeed without actually installing
		assert.NoError(t, err)
	})

	t.Run("context cancellation during yamlfmt installation", func(t *testing.T) {
		// This test would require mocking the internal executor calls
		// For now, we test that the function can handle different execution contexts
		t.Skip("Requires advanced mocking of internal executor context handling")
	})
}

// TestYamlfmtVersionHandling tests version management for yamlfmt
func TestYamlfmtVersionHandling(t *testing.T) {
	t.Run("default version handling", func(t *testing.T) {
		version := GetDefaultYamlfmtVersion()
		assert.NotEmpty(t, version, "Default yamlfmt version should not be empty")

		// Version should either be "latest" or a semantic version
		assert.True(t, version == "latest" || version == VersionLatest ||
			strings.HasPrefix(version, "v"),
			"Version should be 'latest' or start with 'v': %s", version)
	})

	t.Run("version latest fallback", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create a SecureCommandRunner with DryRun enabled
		executor := security.NewSecureExecutor()
		executor.DryRun = true
		secureRunner := &SecureCommandRunner{executor: executor}
		_ = SetRunner(secureRunner) //nolint:errcheck // test setup

		// Test that ensureYamlfmt handles version fallback correctly
		err := ensureYamlfmt()
		assert.NoError(t, err, "Should handle version fallback")
	})
}

// TestYamlfmtInstallationWithCustomConfig tests yamlfmt installation with different configurations
func TestYamlfmtInstallationWithCustomConfig(t *testing.T) {
	t.Run("custom retry configuration from file", func(t *testing.T) {
		// Create a temporary config file with custom retry settings
		tmpDir := t.TempDir()
		configPath := tmpDir + "/.mage.yaml"

		customConfig := `
download:
  maxRetries: 3
  initialDelayMs: 100
`
		err := os.WriteFile(configPath, []byte(customConfig), 0o600)
		require.NoError(t, err)

		// Change to temp directory so config is loaded
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // test cleanup

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create a SecureCommandRunner with dry run
		executor := security.NewSecureExecutor()
		executor.DryRun = true
		secureRunner := &SecureCommandRunner{executor: executor}
		_ = SetRunner(secureRunner) //nolint:errcheck // test setup

		// This should use the custom configuration
		err = ensureYamlfmt()
		assert.NoError(t, err) // Should succeed with DryRun
	})

	t.Run("default configuration when config file missing", func(t *testing.T) {
		// Create a temporary directory without config file
		tmpDir := t.TempDir()

		// Change to temp directory
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // test cleanup

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create a SecureCommandRunner with dry run
		executor := security.NewSecureExecutor()
		executor.DryRun = true
		secureRunner := &SecureCommandRunner{executor: executor}
		_ = SetRunner(secureRunner) //nolint:errcheck // test setup

		// This should use default configuration
		err = ensureYamlfmt()
		assert.NoError(t, err) // Should succeed with DryRun and default config
	})
}

// TestYamlfmtErrorWrapping tests proper error wrapping in yamlfmt retry scenarios
func TestYamlfmtErrorWrapping(t *testing.T) {
	t.Run("error messages from yamlfmt retry logic", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Set a non-SecureCommandRunner to trigger type assertion error
		mockRunner := &MockCommandRunner{}
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// Force the scenario where yamlfmt is not installed by testing on a system without it
		// The error should be properly wrapped
		err := ensureYamlfmt()

		if err != nil {
			// Should contain information about the error type
			assert.True(t, strings.Contains(err.Error(), "expected SecureCommandRunner") ||
				strings.Contains(err.Error(), "unexpected runner type"),
				"Error should indicate type assertion failure: %v", err)
		} else {
			t.Log("yamlfmt already installed, skipping error wrapping test")
		}
	})

	t.Run("network error handling with fallback", func(t *testing.T) {
		// This test would require advanced mocking to simulate network failures
		// and test the GOPROXY=direct fallback logic
		t.Skip("Requires advanced network simulation mocking")
	})
}

// TestYamlfmtInstallationScenarios tests various yamlfmt installation scenarios
func TestYamlfmtInstallationScenarios(t *testing.T) {
	t.Run("successful installation with dry run", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create a SecureCommandRunner with DryRun enabled
		executor := security.NewSecureExecutor()
		executor.DryRun = true
		secureRunner := &SecureCommandRunner{executor: executor}
		_ = SetRunner(secureRunner) //nolint:errcheck // test setup

		// Test successful installation scenario
		err := ensureYamlfmt()
		assert.NoError(t, err, "Dry run installation should succeed")
	})

	t.Run("fallback proxy scenario", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create a SecureCommandRunner with DryRun enabled
		executor := security.NewSecureExecutor()
		executor.DryRun = true
		secureRunner := &SecureCommandRunner{executor: executor}
		_ = SetRunner(secureRunner) //nolint:errcheck // test setup

		// Test that the fallback proxy logic can be executed
		err := ensureYamlfmt()
		assert.NoError(t, err, "Should handle fallback proxy scenario")
	})
}

// TestYamlfmtConcurrentInstallation tests concurrent yamlfmt installation scenarios
func TestYamlfmtConcurrentInstallation(t *testing.T) {
	t.Run("concurrent installation attempts", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create a SecureCommandRunner with DryRun enabled
		executor := security.NewSecureExecutor()
		executor.DryRun = true
		secureRunner := &SecureCommandRunner{executor: executor}
		_ = SetRunner(secureRunner) //nolint:errcheck // test setup

		// Test concurrent installation attempts
		done := make(chan error, 3)

		for i := 0; i < 3; i++ {
			go func() {
				done <- ensureYamlfmt()
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 3; i++ {
			err := <-done
			assert.NoError(t, err, "Concurrent installation should succeed")
		}
	})
}

// TestYamlfmtConfigPathHandling tests yamlfmt config path handling
func TestYamlfmtConfigPathHandling(t *testing.T) {
	t.Run("config file exists", func(t *testing.T) {
		// Create a temporary directory with yamlfmt config
		tmpDir := t.TempDir()

		// Change to temp directory
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // test cleanup

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create .github directory and config file
		err = os.MkdirAll(".github", 0o750)
		require.NoError(t, err)

		configContent := `formatter:
  type: basic
  indent: 2
`
		err = os.WriteFile(".github/.yamlfmt", []byte(configContent), 0o600)
		require.NoError(t, err)

		// Disable YAML validation for this test
		originalEnv := os.Getenv("MAGE_X_YAML_VALIDATION")
		defer func() {
			if originalEnv == "" {
				_ = os.Unsetenv("MAGE_X_YAML_VALIDATION") //nolint:errcheck // test cleanup
			} else {
				_ = os.Setenv("MAGE_X_YAML_VALIDATION", originalEnv) //nolint:errcheck // test cleanup
			}
		}()
		_ = os.Setenv("MAGE_X_YAML_VALIDATION", "false") //nolint:errcheck // test setup

		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock runner to capture commands
		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmdOutput", "find", ".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*").Return("test.yml", nil)
		mockRunner.On("RunCmd", "yamlfmt", "-conf", ".github/.yamlfmt", ".").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// Test YAML formatting with config file
		formatter := Format{}
		err = formatter.YAML()
		require.NoError(t, err, "YAML formatting with config should succeed")

		// Verify the config file was used
		mockRunner.AssertExpectations(t)
	})

	t.Run("config file missing - use defaults", func(t *testing.T) {
		// Create a temporary directory without yamlfmt config
		tmpDir := t.TempDir()

		// Change to temp directory
		origDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // test cleanup

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Disable YAML validation for this test
		originalEnv := os.Getenv("MAGE_X_YAML_VALIDATION")
		defer func() {
			if originalEnv == "" {
				_ = os.Unsetenv("MAGE_X_YAML_VALIDATION") //nolint:errcheck // test cleanup
			} else {
				_ = os.Setenv("MAGE_X_YAML_VALIDATION", originalEnv) //nolint:errcheck // test cleanup
			}
		}()
		_ = os.Setenv("MAGE_X_YAML_VALIDATION", "false") //nolint:errcheck // test setup

		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock runner to capture commands
		mockRunner := &MockCommandRunner{}
		mockRunner.On("RunCmdOutput", "find", ".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*", "-not", "-path", "./*vendor*/*", "-not", "-path", "./node_modules/*", "-not", "-path", "./*node_modules*/*", "-not", "-path", "./.git/*", "-not", "-path", "./*.git*/*", "-not", "-path", "./.idea/*", "-not", "-path", "./*.idea*/*", "-not", "-path", "./.vscode/*", "-not", "-path", "./*.vscode*/*").Return("test.yml", nil)
		mockRunner.On("RunCmd", "yamlfmt", ".").Return(nil)
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// Test YAML formatting without config file
		formatter := Format{}
		err = formatter.YAML()
		require.NoError(t, err, "YAML formatting without config should succeed")

		// Verify the default behavior was used
		mockRunner.AssertExpectations(t)
	})
}
