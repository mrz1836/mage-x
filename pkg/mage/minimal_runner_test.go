package mage

import (
	"errors"
	"testing"

	"github.com/mrz1836/go-mage/pkg/mage/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// MinimalRunnerTestSuite defines the test suite for minimal runner functions
type MinimalRunnerTestSuite struct {
	suite.Suite
	env *testutil.TestEnvironment
}

// SetupTest runs before each test
func (ts *MinimalRunnerTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
}

// TearDownTest runs after each test
func (ts *MinimalRunnerTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestNewSecureCommandRunner tests creating a new secure command runner
func (ts *MinimalRunnerTestSuite) TestNewSecureCommandRunner() {
	runner := NewSecureCommandRunner()
	require.NotNil(ts.T(), runner)

	// Verify it implements CommandRunner interface
	var _ CommandRunner = runner

	// Verify it's a SecureCommandRunner
	secureRunner, ok := runner.(*SecureCommandRunner)
	require.True(ts.T(), ok)
	require.NotNil(ts.T(), secureRunner.executor)
}

// TestSecureCommandRunner_RunCmd tests the RunCmd method
func (ts *MinimalRunnerTestSuite) TestSecureCommandRunner_RunCmd() {
	ts.Run("successful command execution", func() {
		ts.env.Runner.On("RunCmd", "echo", []string{"hello", "world"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) { SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				return runner.RunCmd("echo", "hello", "world")
			},
		)

		require.NoError(ts.T(), err)
	})

	ts.Run("command execution failure", func() {
		expectedError := errors.New("command failed")
		ts.env.Runner.On("RunCmd", "false", []string(nil)).Return(expectedError)

		err := ts.env.WithMockRunner(
			func(r interface{}) { SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				return runner.RunCmd("false")
			},
		)

		require.Error(ts.T(), err)
		require.Equal(ts.T(), expectedError, err)
	})

	ts.Run("command with multiple arguments", func() {
		ts.env.Runner.On("RunCmd", "git", []string{"status", "--porcelain", "--short"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) { SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				return runner.RunCmd("git", "status", "--porcelain", "--short")
			},
		)

		require.NoError(ts.T(), err)
	})

	ts.Run("command with no arguments", func() {
		// Mock expects nil slice when no args are passed
		ts.env.Runner.On("RunCmd", "pwd", []string(nil)).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) { SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				return runner.RunCmd("pwd")
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestSecureCommandRunner_RunCmdOutput tests the RunCmdOutput method
func (ts *MinimalRunnerTestSuite) TestSecureCommandRunner_RunCmdOutput() {
	ts.Run("successful command with output", func() {
		expectedOutput := "hello world"
		ts.env.Runner.On("RunCmdOutput", "echo", []string{"hello", "world"}).Return(expectedOutput, nil)

		var output string
		err := ts.env.WithMockRunner(
			func(r interface{}) { SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				var err error
				output, err = runner.RunCmdOutput("echo", "hello", "world")
				return err
			},
		)

		require.NoError(ts.T(), err)
		require.Equal(ts.T(), expectedOutput, output)
	})

	ts.Run("command output with trailing whitespace", func() {
		// The RunCmdOutput method should NOT trim whitespace automatically
		// (that happens in the SecureCommandRunner implementation)
		rawOutput := "hello world\n\n\t "
		ts.env.Runner.On("RunCmdOutput", "echo", []string{"test"}).Return(rawOutput, nil)

		var output string
		err := ts.env.WithMockRunner(
			func(r interface{}) { SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				var err error
				output, err = runner.RunCmdOutput("echo", "test")
				return err
			},
		)

		require.NoError(ts.T(), err)
		require.Equal(ts.T(), rawOutput, output) // Test should reflect mock's exact return
	})

	ts.Run("command output failure", func() {
		expectedError := errors.New("command output failed")
		ts.env.Runner.On("RunCmdOutput", "false", []string(nil)).Return("", expectedError)

		var output string
		err := ts.env.WithMockRunner(
			func(r interface{}) { SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				var err error
				output, err = runner.RunCmdOutput("false")
				return err
			},
		)

		require.Error(ts.T(), err)
		require.Equal(ts.T(), expectedError, err)
		require.Empty(ts.T(), output)
	})

	ts.Run("empty command output", func() {
		ts.env.Runner.On("RunCmdOutput", "true", []string(nil)).Return("", nil)

		var output string
		err := ts.env.WithMockRunner(
			func(r interface{}) { SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				var err error
				output, err = runner.RunCmdOutput("true")
				return err
			},
		)

		require.NoError(ts.T(), err)
		require.Empty(ts.T(), output)
	})
}

// TestGetRunner tests the GetRunner function
func (ts *MinimalRunnerTestSuite) TestGetRunner() {
	runner := GetRunner()
	require.NotNil(ts.T(), runner)

	// Should return the same instance (singleton pattern)
	runner2 := GetRunner()
	require.Same(ts.T(), runner, runner2)
}

// TestSetRunner tests the SetRunner function
func (ts *MinimalRunnerTestSuite) TestSetRunner() {
	// Get original runner
	originalRunner := GetRunner()
	require.NotNil(ts.T(), originalRunner)

	// Create a custom mock runner
	mockRunner := ts.env.Runner

	// Set the mock runner
	SetRunner(mockRunner)

	// Verify the runner was set
	currentRunner := GetRunner()
	require.Same(ts.T(), mockRunner, currentRunner)
	require.NotSame(ts.T(), originalRunner, currentRunner)

	// Restore original runner
	SetRunner(originalRunner)
	restoredRunner := GetRunner()
	require.Same(ts.T(), originalRunner, restoredRunner)
}

// TestTruncateString tests the truncateString helper function
func (ts *MinimalRunnerTestSuite) TestTruncateString() {
	testCases := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "string shorter than max length",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "string equal to max length",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "string longer than max length",
			input:    "hello world",
			maxLen:   5,
			expected: "hello...",
		},
		{
			name:     "empty string",
			input:    "",
			maxLen:   5,
			expected: "",
		},
		{
			name:     "max length zero",
			input:    "hello",
			maxLen:   0,
			expected: "...",
		},
		{
			name:     "max length one",
			input:    "hello",
			maxLen:   1,
			expected: "h...",
		},
		{
			name:     "unicode string",
			input:    "café world",
			maxLen:   4,
			expected: "caf\xc3...", // truncateString is byte-based, not rune-based - cuts mid-UTF8 sequence
		},
		{
			name:     "very long string",
			input:    "This is a very long string that should be truncated",
			maxLen:   20,
			expected: "This is a very long ...",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			result := truncateString(tc.input, tc.maxLen)
			require.Equal(ts.T(), tc.expected, result)
		})
	}
}

// TestFormatBytes tests the formatBytes helper function
func (ts *MinimalRunnerTestSuite) TestFormatBytes() {
	testCases := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "zero bytes",
			bytes:    0,
			expected: "0 B",
		},
		{
			name:     "bytes less than 1KB",
			bytes:    512,
			expected: "512 B",
		},
		{
			name:     "exactly 1KB",
			bytes:    1024,
			expected: "1.0 KB",
		},
		{
			name:     "kilobytes",
			bytes:    2048,
			expected: "2.0 KB",
		},
		{
			name:     "kilobytes with decimal",
			bytes:    1536, // 1.5 KB
			expected: "1.5 KB",
		},
		{
			name:     "exactly 1MB",
			bytes:    1024 * 1024,
			expected: "1.0 MB",
		},
		{
			name:     "megabytes",
			bytes:    5 * 1024 * 1024,
			expected: "5.0 MB",
		},
		{
			name:     "megabytes with decimal",
			bytes:    1536 * 1024, // 1.5 MB
			expected: "1.5 MB",
		},
		{
			name:     "exactly 1GB",
			bytes:    1024 * 1024 * 1024,
			expected: "1.0 GB",
		},
		{
			name:     "gigabytes",
			bytes:    3 * 1024 * 1024 * 1024,
			expected: "3.0 GB",
		},
		{
			name:     "gigabytes with decimal",
			bytes:    1536 * 1024 * 1024, // 1.5 GB
			expected: "1.5 GB",
		},
		{
			name:     "exactly 1TB",
			bytes:    1024 * 1024 * 1024 * 1024,
			expected: "1.0 TB",
		},
		{
			name:     "terabytes",
			bytes:    2 * 1024 * 1024 * 1024 * 1024,
			expected: "2.0 TB",
		},
		{
			name:     "petabytes",
			bytes:    1024 * 1024 * 1024 * 1024 * 1024,
			expected: "1.0 PB",
		},
		{
			name:     "exabytes",
			bytes:    1024 * 1024 * 1024 * 1024 * 1024 * 1024,
			expected: "1.0 EB",
		},
		{
			name:     "large number",
			bytes:    9223372036854775807, // Max int64
			expected: "8.0 EB",
		},
		{
			name:     "mixed precision",
			bytes:    1234567890,
			expected: "1.1 GB",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			result := formatBytes(tc.bytes)
			require.Equal(ts.T(), tc.expected, result)
		})
	}
}

// TestMinimalRunnerIntegration tests integration scenarios
func (ts *MinimalRunnerTestSuite) TestMinimalRunnerIntegration() {
	ts.Run("runner persistence across operations", func() {
		// Get initial runner
		runner1 := GetRunner()

		// Set a mock runner
		mockRunner := ts.env.Runner
		SetRunner(mockRunner)

		// Verify the mock is active
		currentRunner := GetRunner()
		require.Same(ts.T(), mockRunner, currentRunner)

		// Set up mock expectations
		mockRunner.On("RunCmd", "test1", []string(nil)).Return(nil)
		mockRunner.On("RunCmdOutput", "test2", []string(nil)).Return("output", nil)

		// Execute commands through the runner
		err := currentRunner.RunCmd("test1")
		require.NoError(ts.T(), err)

		output, err := currentRunner.RunCmdOutput("test2")
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), "output", output)

		// Restore original runner
		SetRunner(runner1)
		restoredRunner := GetRunner()
		require.Same(ts.T(), runner1, restoredRunner)
	})

	ts.Run("error handling chain", func() {
		ts.env.Runner.On("RunCmd", "failing-command", []string(nil)).Return(errors.New("execution failed"))

		err := ts.env.WithMockRunner(
			func(r interface{}) { SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				return runner.RunCmd("failing-command")
			},
		)

		require.Error(ts.T(), err)
		require.Contains(ts.T(), err.Error(), "execution failed")
	})
}

// TestMinimalRunnerTestSuite runs the test suite
func TestMinimalRunnerTestSuite(t *testing.T) {
	suite.Run(t, new(MinimalRunnerTestSuite))
}