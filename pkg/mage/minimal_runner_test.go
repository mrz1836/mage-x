package mage

import (
	"errors"
	"testing"
	"time"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Static test errors to satisfy err113 linter
var (
	errMinimalCommandFailed       = errors.New("command failed")
	errMinimalCommandOutputFailed = errors.New("command output failed")
	errMinimalExecutionFailed     = errors.New("execution failed")
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
	ts.Require().NotNil(runner)

	// Verify it implements CommandRunner interface
	_ = runner

	// Verify it's a SecureCommandRunner
	secureRunner, ok := runner.(*SecureCommandRunner)
	ts.Require().True(ok)
	ts.Require().NotNil(secureRunner.executor)
}

// TestSecureCommandRunner_RunCmd tests the RunCmd method
func (ts *MinimalRunnerTestSuite) TestSecureCommandRunner_RunCmd() {
	ts.Run("successful command execution", func() {
		ts.env.Runner.On("RunCmd", "echo", []string{"hello", "world"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				return runner.RunCmd("echo", "hello", "world")
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("command execution failure", func() {
		expectedError := errMinimalCommandFailed
		ts.env.Runner.On("RunCmd", "false", []string(nil)).Return(expectedError)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				return runner.RunCmd("false")
			},
		)

		ts.Require().Error(err)
		ts.Require().Equal(expectedError, err)
	})

	ts.Run("command with multiple arguments", func() {
		ts.env.Runner.On("RunCmd", "git", []string{"status", "--porcelain", "--short"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				return runner.RunCmd("git", "status", "--porcelain", "--short")
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("command with no arguments", func() {
		// Mock expects nil slice when no args are passed
		ts.env.Runner.On("RunCmd", "pwd", []string(nil)).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				return runner.RunCmd("pwd")
			},
		)

		ts.Require().NoError(err)
	})
}

// TestSecureCommandRunner_RunCmdOutput tests the RunCmdOutput method
func (ts *MinimalRunnerTestSuite) TestSecureCommandRunner_RunCmdOutput() {
	ts.Run("successful command with output", func() {
		expectedOutput := "hello world"
		ts.env.Runner.On("RunCmdOutput", "echo", []string{"hello", "world"}).Return(expectedOutput, nil)

		var output string
		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				var err error
				output, err = runner.RunCmdOutput("echo", "hello", "world")
				return err
			},
		)

		ts.Require().NoError(err)
		ts.Require().Equal(expectedOutput, output)
	})

	ts.Run("command output with trailing whitespace", func() {
		// The RunCmdOutput method should NOT trim whitespace automatically
		// (that happens in the SecureCommandRunner implementation)
		rawOutput := "hello world\n\n\t "
		ts.env.Runner.On("RunCmdOutput", "echo", []string{"test"}).Return(rawOutput, nil)

		var output string
		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				var err error
				output, err = runner.RunCmdOutput("echo", "test")
				return err
			},
		)

		ts.Require().NoError(err)
		ts.Require().Equal(rawOutput, output) // Test should reflect mock's exact return
	})

	ts.Run("command output failure", func() {
		expectedError := errMinimalCommandOutputFailed
		ts.env.Runner.On("RunCmdOutput", "false", []string(nil)).Return("", expectedError)

		var output string
		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				var err error
				output, err = runner.RunCmdOutput("false")
				return err
			},
		)

		ts.Require().Error(err)
		ts.Require().Equal(expectedError, err)
		ts.Require().Empty(output)
	})

	ts.Run("empty command output", func() {
		ts.env.Runner.On("RunCmdOutput", "true", []string(nil)).Return("", nil)

		var output string
		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				var err error
				output, err = runner.RunCmdOutput("true")
				return err
			},
		)

		ts.Require().NoError(err)
		ts.Require().Empty(output)
	})
}

// TestGetRunner tests the GetRunner function
func (ts *MinimalRunnerTestSuite) TestGetRunner() {
	runner := GetRunner()
	ts.Require().NotNil(runner)

	// Should return the same instance (singleton pattern)
	runner2 := GetRunner()
	ts.Require().Same(runner, runner2)
}

// TestSetRunner tests the SetRunner function
func (ts *MinimalRunnerTestSuite) TestSetRunner() {
	// Get original runner
	originalRunner := GetRunner()
	ts.Require().NotNil(originalRunner)

	// Create a custom mock runner
	mockRunner := ts.env.Runner

	// Set the mock runner
	ts.Require().NoError(SetRunner(mockRunner))

	// Verify the runner was set
	currentRunner := GetRunner()
	ts.Require().Same(mockRunner, currentRunner)
	ts.Require().NotSame(originalRunner, currentRunner)

	// Restore original runner
	ts.Require().NoError(SetRunner(originalRunner))
	restoredRunner := GetRunner()
	ts.Require().Same(originalRunner, restoredRunner)
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
			ts.Require().Equal(tc.expected, result)
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
			ts.Require().Equal(tc.expected, result)
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
		ts.Require().NoError(SetRunner(mockRunner))

		// Verify the mock is active
		currentRunner := GetRunner()
		ts.Require().Same(mockRunner, currentRunner)

		// Set up mock expectations
		mockRunner.On("RunCmd", "test1", []string(nil)).Return(nil)
		mockRunner.On("RunCmdOutput", "test2", []string(nil)).Return("output", nil)

		// Execute commands through the runner
		err := currentRunner.RunCmd("test1")
		ts.Require().NoError(err)

		output, err := currentRunner.RunCmdOutput("test2")
		ts.Require().NoError(err)
		ts.Require().Equal("output", output)

		// Restore original runner
		ts.Require().NoError(SetRunner(runner1))
		restoredRunner := GetRunner()
		ts.Require().Same(runner1, restoredRunner)
	})

	ts.Run("error handling chain", func() {
		ts.env.Runner.On("RunCmd", "failing-command", []string(nil)).Return(errMinimalExecutionFailed)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				runner := GetRunner()
				return runner.RunCmd("failing-command")
			},
		)

		ts.Require().Error(err)
		ts.Require().Contains(err.Error(), "execution failed")
	})
}

// TestSecureCommandRunner_getCommandTimeout tests the timeout logic
func TestSecureCommandRunner_getCommandTimeout(t *testing.T) {
	runner := &SecureCommandRunner{}

	tests := []struct {
		name     string
		cmd      string
		args     []string
		expected time.Duration
	}{
		// Go commands
		{
			name:     "go test gets 10 minutes",
			cmd:      "go",
			args:     []string{"test", "./..."},
			expected: 10 * time.Minute,
		},
		{
			name:     "go install gets 5 minutes",
			cmd:      "go",
			args:     []string{"install", "github.com/tool"},
			expected: 5 * time.Minute,
		},
		{
			name:     "go build gets 3 minutes",
			cmd:      "go",
			args:     []string{"build", "./cmd/app"},
			expected: 3 * time.Minute,
		},
		{
			name:     "go fmt gets 2 minutes",
			cmd:      "go",
			args:     []string{"fmt", "./..."},
			expected: 2 * time.Minute,
		},
		// Mage commands
		{
			name:     "mage testDefault gets 10 minutes",
			cmd:      "mage",
			args:     []string{"testDefault"},
			expected: 10 * time.Minute,
		},
		{
			name:     "mage test gets 10 minutes",
			cmd:      "mage",
			args:     []string{"test"},
			expected: 10 * time.Minute,
		},
		{
			name:     "mage test:cover gets 10 minutes",
			cmd:      "mage",
			args:     []string{"test:cover"},
			expected: 10 * time.Minute,
		},
		{
			name:     "mage test:race gets 10 minutes",
			cmd:      "mage",
			args:     []string{"test:race"},
			expected: 10 * time.Minute,
		},
		{
			name:     "mage test:ci gets 10 minutes",
			cmd:      "mage",
			args:     []string{"test:ci"},
			expected: 10 * time.Minute,
		},
		{
			name:     "mage test:full gets 10 minutes",
			cmd:      "mage",
			args:     []string{"test:full"},
			expected: 10 * time.Minute,
		},
		{
			name:     "mage build gets 3 minutes",
			cmd:      "mage",
			args:     []string{"build"},
			expected: 3 * time.Minute,
		},
		{
			name:     "mage lint gets 3 minutes",
			cmd:      "mage",
			args:     []string{"lint"},
			expected: 3 * time.Minute,
		},
		// Other tools
		{
			name:     "goreleaser gets 30 minutes",
			cmd:      "goreleaser",
			args:     []string{"release", "--clean"},
			expected: 30 * time.Minute,
		},
		{
			name:     "golangci-lint gets 5 minutes",
			cmd:      "golangci-lint",
			args:     []string{"run", "./..."},
			expected: 5 * time.Minute,
		},
		{
			name:     "staticcheck gets 3 minutes",
			cmd:      "staticcheck",
			args:     []string{"./..."},
			expected: 3 * time.Minute,
		},
		{
			name:     "gosec gets 3 minutes",
			cmd:      "gosec",
			args:     []string{"./..."},
			expected: 3 * time.Minute,
		},
		{
			name:     "govulncheck gets 3 minutes",
			cmd:      "govulncheck",
			args:     []string{"./..."},
			expected: 3 * time.Minute,
		},
		{
			name:     "echo gets 30 seconds",
			cmd:      "echo",
			args:     []string{"hello"},
			expected: 30 * time.Second,
		},
		{
			name:     "git gets 30 seconds",
			cmd:      "git",
			args:     []string{"status"},
			expected: 30 * time.Second,
		},
		// Edge cases
		{
			name:     "mage with no args gets 3 minutes",
			cmd:      "mage",
			args:     []string{},
			expected: 3 * time.Minute,
		},
		{
			name:     "go with no args gets 2 minutes",
			cmd:      "go",
			args:     []string{},
			expected: 2 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			timeout := runner.getCommandTimeout(tt.cmd, tt.args)
			assert.Equal(t, tt.expected, timeout)
		})
	}
}

// TestMinimalRunnerTestSuite runs the test suite
func TestMinimalRunnerTestSuite(t *testing.T) {
	suite.Run(t, new(MinimalRunnerTestSuite))
}
