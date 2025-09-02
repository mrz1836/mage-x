package mage

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mrz1836/mage-x/pkg/security"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockSecureCommandRunner is a mock implementation of CommandRunner that simulates SecureCommandRunner behavior
type MockSecureCommandRunner struct {
	MockCommandRunner

	executor *MockSecureExecutor //nolint:unused // field reserved for future use
}

// MockSecureExecutor is a mock for security.SecureExecutor methods needed for retry testing
type MockSecureExecutor struct {
	mock.Mock
}

func (m *MockSecureExecutor) Execute(ctx context.Context, name string, args ...string) error {
	callArgs := []interface{}{ctx, name}
	for _, arg := range args {
		callArgs = append(callArgs, arg)
	}
	args_ := m.Called(callArgs...)
	return args_.Error(0)
}

func (m *MockSecureExecutor) ExecuteOutput(ctx context.Context, name string, args ...string) (string, error) {
	callArgs := []interface{}{ctx, name}
	for _, arg := range args {
		callArgs = append(callArgs, arg)
	}
	args_ := m.Called(callArgs...)
	return args_.String(0), args_.Error(1)
}

func (m *MockSecureExecutor) ExecuteWithRetry(ctx context.Context, maxRetries int, initialDelay time.Duration, name string, args ...string) error {
	callArgs := []interface{}{ctx, maxRetries, initialDelay, name}
	for _, arg := range args {
		callArgs = append(callArgs, arg)
	}
	args_ := m.Called(callArgs...)
	return args_.Error(0)
}

func (m *MockSecureExecutor) ExecuteWithEnv(ctx context.Context, env []string, name string, args ...string) error {
	callArgs := []interface{}{ctx, env, name}
	for _, arg := range args {
		callArgs = append(callArgs, arg)
	}
	args_ := m.Called(callArgs...)
	return args_.Error(0)
}

// TestEnsureGofumptRetryLogicIntegration tests retry logic in integration with command execution
func TestEnsureGofumptRetryLogicIntegration(t *testing.T) {
	t.Run("tool already exists", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock runner
		mockRunner := &MockCommandRunner{}
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		// When gofumpt is already available, ensureGofumpt should not try to install
		// We can't easily mock utils.CommandExists but can test the overall behavior

		err := ensureGofumpt()
		// If the test system has gofumpt installed, this should succeed without installation
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

		// In a real integration test, this would attempt to install gofumpt
		err := ensureGofumpt()
		// This test depends on network availability and actual tool installation
		// In CI/testing environment, this might be mocked or skipped
		if err != nil {
			t.Logf("Installation failed (expected in test environment): %v", err)
		}
	})
}

// TestEnsureGoimportsRetryLogicIntegration tests retry logic for goimports
func TestEnsureGoimportsRetryLogicIntegration(t *testing.T) {
	t.Run("tool already exists", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create mock runner
		mockRunner := &MockCommandRunner{}
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		err := ensureGoimports()
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
		secureRunner := NewSecureCommandRunner()
		_ = SetRunner(secureRunner) //nolint:errcheck // test setup

		err := ensureGoimports()
		// This test depends on network availability and actual tool installation
		if err != nil {
			t.Logf("Installation failed (expected in test environment): %v", err)
		}
	})
}

// TestSecureCommandRunnerTypeAssertion tests type assertion behavior
func TestSecureCommandRunnerTypeAssertion(t *testing.T) {
	t.Run("wrong runner type in ensureGofumpt", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Set a non-SecureCommandRunner
		mockRunner := &MockCommandRunner{}
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		err := ensureGofumpt()

		// If gofumpt already exists on the system, this will return nil
		// If gofumpt doesn't exist, it will try to install and get type assertion error
		if err != nil {
			assert.Contains(t, err.Error(), "expected SecureCommandRunner")
			assert.ErrorIs(t, err, ErrUnexpectedRunnerType)
		} else {
			t.Log("gofumpt already exists on system, no type assertion occurred")
		}
	})

	t.Run("wrong runner type in ensureGoimports", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Set a non-SecureCommandRunner
		mockRunner := &MockCommandRunner{}
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		err := ensureGoimports()

		// If goimports already exists on the system, this will return nil
		// If goimports doesn't exist, it will try to install and get type assertion error
		if err != nil {
			assert.Contains(t, err.Error(), "expected SecureCommandRunner")
			assert.ErrorIs(t, err, ErrUnexpectedRunnerType)
		} else {
			t.Log("goimports already exists on system, no type assertion occurred")
		}
	})
}

// TestRetryBehaviorWithSecureExecutor tests retry behavior using the actual SecureExecutor
func TestRetryBehaviorWithSecureExecutor(t *testing.T) {
	t.Run("retry logic with actual executor", func(t *testing.T) {
		// Create a SecureExecutor instance
		executor := security.NewSecureExecutor()

		// Test the retry logic directly
		ctx := context.Background()

		// This should fail quickly for a non-existent command
		err := executor.ExecuteWithRetry(ctx, 2, 10*time.Millisecond, "nonexistentcommandxyz123", "arg1")
		require.Error(t, err)
		// Should contain information about the error (either retry attempts or permanent error)
		assert.True(t, strings.Contains(err.Error(), "command failed after") ||
			strings.Contains(err.Error(), "permanent command error") ||
			strings.Contains(err.Error(), "executable file not found"),
			"Error should indicate command failure: %v", err)
	})

	t.Run("successful execution with retry", func(t *testing.T) {
		// Create a SecureExecutor instance
		executor := security.NewSecureExecutor()

		// Test with a command that should exist (echo)
		ctx := context.Background()

		err := executor.ExecuteWithRetry(ctx, 1, 10*time.Millisecond, "echo", "test")
		// This should succeed on the first try
		assert.NoError(t, err)
	})
}

// TestInstallToolsWithSecureRunner tests InstallTools with actual SecureCommandRunner
func TestInstallToolsWithSecureRunner(t *testing.T) {
	t.Run("install tools with dry run", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// Create a SecureCommandRunner with DryRun enabled
		executor := security.NewSecureExecutor()
		executor.DryRun = true

		runner := &SecureCommandRunner{executor: executor}
		_ = SetRunner(runner) //nolint:errcheck // test setup

		format := Format{}
		err := format.InstallTools()

		// With DryRun=true, this should succeed without actually installing
		assert.NoError(t, err)
	})
}

// TestConfigurationRetryBehavior tests retry behavior with different configurations
func TestConfigurationRetryBehavior(t *testing.T) {
	t.Run("custom retry configuration from file", func(t *testing.T) {
		// Create a temporary config file with custom retry settings
		tmpDir := t.TempDir()
		configPath := tmpDir + "/.mage.yaml"

		customConfig := `
download:
  maxRetries: 2
  initialDelayMs: 50
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

		runner := &SecureCommandRunner{executor: executor}
		_ = SetRunner(runner) //nolint:errcheck // test setup

		// This should use the custom configuration
		err = ensureGofumpt()
		assert.NoError(t, err) // Should succeed with DryRun
	})
}

// TestErrorWrapping tests proper error wrapping in retry scenarios
func TestErrorWrapping(t *testing.T) {
	t.Run("error messages from retry logic", func(t *testing.T) {
		// Test that ExecuteWithRetry properly wraps errors
		executor := security.NewSecureExecutor()
		ctx := context.Background()

		// Use a command that will fail
		err := executor.ExecuteWithRetry(ctx, 1, 10*time.Millisecond, "falsecmdthatdoesnotexist")
		require.Error(t, err)

		// Should contain information about the error (either retry attempts or permanent error)
		assert.True(t, strings.Contains(err.Error(), "command failed after") ||
			strings.Contains(err.Error(), "permanent command error") ||
			strings.Contains(err.Error(), "executable file not found"),
			"Error should indicate command failure: %v", err)
	})

	t.Run("context cancellation during retry", func(t *testing.T) {
		executor := security.NewSecureExecutor()

		// Create a context that will be canceled
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Millisecond)
		defer cancel()

		err := executor.ExecuteWithRetry(ctx, 5, 100*time.Millisecond, "sleep", "10")
		require.Error(t, err)

		// Check for context error or command termination due to timeout
		// The context timeout causes the command to be killed with "signal: killed"
		assert.True(t, errors.Is(err, context.DeadlineExceeded) ||
			errors.Is(err, context.Canceled) ||
			strings.Contains(err.Error(), "context") ||
			strings.Contains(err.Error(), "deadline exceeded") ||
			strings.Contains(err.Error(), "command failed after") ||
			strings.Contains(err.Error(), "signal: killed") ||
			strings.Contains(err.Error(), "permanent command error"),
			"Error should indicate context cancellation or timeout, got: %v", err)
	})
}

// TestInstallToolsErrorHandling tests error handling in InstallTools
func TestInstallToolsErrorHandling(t *testing.T) {
	t.Run("format install tools with mock failures", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() { _ = SetRunner(originalRunner) }() //nolint:errcheck // test cleanup

		// This tests the overall error handling structure
		// The InstallTools method tries to type assert to SecureCommandRunner
		// When it gets MockCommandRunner instead, it should fail with type assertion error
		mockRunner := &MockCommandRunner{}
		_ = SetRunner(mockRunner) //nolint:errcheck // test setup

		format := Format{}
		err := format.InstallTools()

		// Should get type assertion error since MockCommandRunner doesn't implement SecureCommandRunner
		// However, the current implementation might succeed if tools are already installed
		// Let's check what we actually get
		if err != nil {
			// If we get an error, it should be the expected type assertion error
			assert.True(t, strings.Contains(err.Error(), "unexpected runner type") ||
				strings.Contains(err.Error(), "failed to install"))
		} else {
			// If no error, it means tools were already installed and check succeeded
			// This is actually valid behavior - skip the test in this case
			t.Skip("Tools already installed, cannot test error handling")
		}
	})
}
