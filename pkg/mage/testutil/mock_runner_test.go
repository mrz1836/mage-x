package testutil_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// Static errors for testing
var (
	errCommandTest       = errors.New("command failed")
	errOutputCommandTest = errors.New("output command failed")
	errGitCommandTest    = errors.New("git command failed")
	errBuildTest         = errors.New("build failed")
	errAnyCommandTest    = errors.New("any command failed")
)

// MockRunnerTestSuite tests the MockRunner and related functionality
type MockRunnerTestSuite struct {
	suite.Suite

	runner  *testutil.MockRunner
	builder *testutil.MockBuilder
}

func (ts *MockRunnerTestSuite) SetupTest() {
	ts.runner, ts.builder = testutil.NewMockRunner()
}

func (ts *MockRunnerTestSuite) TestNewMockRunner() {
	runner, builder := testutil.NewMockRunner()
	ts.Require().NotNil(runner)
	ts.Require().NotNil(builder)
}

func (ts *MockRunnerTestSuite) TestRunCmd() {
	// Set up expectation
	ts.runner.On("RunCmd", "echo", []string{"hello"}).Return(nil)

	// Execute command
	err := ts.runner.RunCmd("echo", "hello")
	ts.Require().NoError(err)

	// Verify expectations
	ts.runner.AssertExpectations(ts.T())
}

func (ts *MockRunnerTestSuite) TestRunCmdWithError() {
	// Set up expectation with error
	expectedErr := errCommandTest
	ts.runner.On("RunCmd", "failing-command", []string{"arg1"}).Return(expectedErr)

	// Execute command
	err := ts.runner.RunCmd("failing-command", "arg1")
	ts.Require().Error(err)
	ts.Require().Equal(expectedErr, err)

	// Verify expectations
	ts.runner.AssertExpectations(ts.T())
}

func (ts *MockRunnerTestSuite) TestRunCmdOutput() {
	// Set up expectation
	expectedOutput := "command output"
	ts.runner.On("RunCmdOutput", "ls", []string{"-la"}).Return(expectedOutput, nil)

	// Execute command
	output, err := ts.runner.RunCmdOutput("ls", "-la")
	ts.Require().NoError(err)
	ts.Require().Equal(expectedOutput, output)

	// Verify expectations
	ts.runner.AssertExpectations(ts.T())
}

func (ts *MockRunnerTestSuite) TestRunCmdOutputWithError() {
	// Set up expectation with error
	expectedErr := errOutputCommandTest
	ts.runner.On("RunCmdOutput", "failing-command", []string{"arg1"}).Return("", expectedErr)

	// Execute command
	output, err := ts.runner.RunCmdOutput("failing-command", "arg1")
	ts.Require().Error(err)
	ts.Require().Equal(expectedErr, err)
	ts.Require().Empty(output)

	// Verify expectations
	ts.runner.AssertExpectations(ts.T())
}

func (ts *MockRunnerTestSuite) TestExpectGitCommand() {
	mocks := ts.builder.ExpectGitCommand("status", "clean working directory", nil)
	ts.Require().NotNil(mocks)

	// Test the expectation works
	output, err := ts.runner.RunCmdOutput("git", "status", "--porcelain")
	ts.Require().NoError(err)
	ts.Require().Equal("clean working directory", output)
}

func (ts *MockRunnerTestSuite) TestExpectGitCommandWithError() {
	expectedErr := errGitCommandTest
	ts.builder.ExpectGitCommand("push", "", expectedErr)

	// Test the expectation works
	output, err := ts.runner.RunCmdOutput("git", "push", "origin", "main")
	ts.Require().Error(err)
	ts.Require().Equal(expectedErr, err)
	ts.Require().Empty(output)
}

func (ts *MockRunnerTestSuite) TestExpectGoCommand() {
	mocks := ts.builder.ExpectGoCommand("build", nil)
	ts.Require().NotNil(mocks)

	// Test the expectation works
	err := ts.runner.RunCmd("go", "build", "./...")
	ts.Require().NoError(err)
}

func (ts *MockRunnerTestSuite) TestExpectGoCommandWithError() {
	expectedErr := errBuildTest
	ts.builder.ExpectGoCommand("build", expectedErr)

	// Test the expectation works
	err := ts.runner.RunCmd("go", "build", "./...")
	ts.Require().Error(err)
	ts.Require().Equal(expectedErr, err)
}

func (ts *MockRunnerTestSuite) TestExpectAnyCommand() {
	mocks := ts.builder.ExpectAnyCommand(nil)
	ts.Require().NotNil(mocks)

	// Test that any command works
	err := ts.runner.RunCmd("any-command", "arg1", "arg2")
	ts.Require().NoError(err)

	output, err := ts.runner.RunCmdOutput("another-command", "arg1")
	ts.Require().NoError(err)
	ts.Require().Empty(output)
}

func (ts *MockRunnerTestSuite) TestExpectAnyCommandWithError() {
	expectedErr := errAnyCommandTest
	ts.builder.ExpectAnyCommand(expectedErr)

	// Test that any command returns the error
	err := ts.runner.RunCmd("any-command", "arg1", "arg2")
	ts.Require().Error(err)
	ts.Require().Equal(expectedErr, err)

	output, err := ts.runner.RunCmdOutput("another-command", "arg1")
	ts.Require().Error(err)
	ts.Require().Equal(expectedErr, err)
	ts.Require().Empty(output)
}

func (ts *MockRunnerTestSuite) TestExpectVersion() {
	version := "v1.2.3"
	mocks := ts.builder.ExpectVersion(version)
	ts.Require().NotNil(mocks)

	// Test version commands
	output, err := ts.runner.RunCmdOutput("git", "describe", "--tags")
	ts.Require().NoError(err)
	ts.Require().Equal(version, output)

	output, err = ts.runner.RunCmdOutput("git", "rev-parse", "HEAD")
	ts.Require().NoError(err)
	ts.Require().Equal("abc123", output)
}

func (ts *MockRunnerTestSuite) TestExpectSuccess() {
	mocks := ts.builder.ExpectSuccess()
	ts.Require().NotNil(mocks)

	// Test that any command succeeds
	err := ts.runner.RunCmd("successful-command", "arg1")
	ts.Require().NoError(err)

	output, err := ts.runner.RunCmdOutput("output-command", "arg1")
	ts.Require().NoError(err)
	ts.Require().Empty(output)
}

func (ts *MockRunnerTestSuite) TestExpectFailure() {
	message := "operation failed"
	mocks := ts.builder.ExpectFailure(message)
	ts.Require().NotNil(mocks)

	// Test that any command fails with the expected message
	err := ts.runner.RunCmd("failing-command", "arg1")
	ts.Require().Error(err)
	ts.Require().Contains(err.Error(), message)
	ts.Require().Contains(err.Error(), "command failed")

	output, err := ts.runner.RunCmdOutput("output-command", "arg1")
	ts.Require().Error(err)
	ts.Require().Contains(err.Error(), message)
	ts.Require().Empty(output)
}

func (ts *MockRunnerTestSuite) TestBuild() {
	mocks := ts.builder.ExpectSuccess()
	runner := mocks.Build()
	ts.Require().NotNil(runner)
	ts.Require().Equal(ts.runner, runner)
}

func (ts *MockRunnerTestSuite) TestCommandMatcherContainsArg() {
	matcher := testutil.NewCommandMatcher()
	ts.Require().NotNil(matcher)

	// Test ContainsArg matcher
	containsArg := matcher.ContainsArg("--verbose")

	ts.Require().True(containsArg([]string{"command", "--verbose", "arg"}))
	ts.Require().True(containsArg([]string{"--verbose"}))
	ts.Require().False(containsArg([]string{"command", "arg"}))
	ts.Require().False(containsArg([]string{}))
}

func (ts *MockRunnerTestSuite) TestCommandMatcherHasSubcommand() {
	matcher := testutil.NewCommandMatcher()

	// Test HasSubcommand matcher
	hasSubcommand := matcher.HasSubcommand("build")

	ts.Require().True(hasSubcommand([]string{"build", "arg1", "arg2"}))
	ts.Require().True(hasSubcommand([]string{"build"}))
	ts.Require().False(hasSubcommand([]string{"test", "arg1"}))
	ts.Require().False(hasSubcommand([]string{}))
}

func (ts *MockRunnerTestSuite) TestCommandMatcherContainsFlag() {
	matcher := testutil.NewCommandMatcher()

	// Test ContainsFlag matcher
	containsFlag := matcher.ContainsFlag("--output")

	ts.Require().True(containsFlag([]string{"command", "--output=file.txt"}))
	ts.Require().True(containsFlag([]string{"--output", "file.txt"}))
	ts.Require().True(containsFlag([]string{"--output-dir", "dir"}))
	ts.Require().False(containsFlag([]string{"command", "-o", "file.txt"}))
	ts.Require().False(containsFlag([]string{}))
}

func (ts *MockRunnerTestSuite) TestNewCommandMatcher() {
	matcher := testutil.NewCommandMatcher()
	ts.Require().NotNil(matcher)
}

func (ts *MockRunnerTestSuite) TestChainedExpectations() {
	// Test chaining multiple expectations
	ts.builder.
		ExpectVersion("v1.0.0").
		ExpectGoCommand("build", nil).
		ExpectGoCommand("test", nil).
		ExpectGitCommand("status", "clean", nil).
		ExpectSuccess()

	// Test all expectations work
	output, err := ts.runner.RunCmdOutput("git", "describe", "--tags")
	ts.Require().NoError(err)
	ts.Require().Equal("v1.0.0", output)

	err = ts.runner.RunCmd("go", "build", "./...")
	ts.Require().NoError(err)

	err = ts.runner.RunCmd("go", "test", "./...")
	ts.Require().NoError(err)

	output, err = ts.runner.RunCmdOutput("git", "status", "--porcelain")
	ts.Require().NoError(err)
	ts.Require().Equal("clean", output)

	err = ts.runner.RunCmd("any-other-command")
	ts.Require().NoError(err)
}

func (ts *MockRunnerTestSuite) TestMultipleArguments() {
	// Test with multiple arguments of various types
	ts.runner.On("RunCmd", "command", []string{"arg1", "arg2", "arg3"}).Return(nil)

	err := ts.runner.RunCmd("command", "arg1", "arg2", "arg3")
	ts.Require().NoError(err)

	ts.runner.AssertExpectations(ts.T())
}

func (ts *MockRunnerTestSuite) TestEmptyArguments() {
	// Test with no arguments - use mock.MatchedBy to handle empty slice matching
	ts.runner.On("RunCmd", "command", mock.MatchedBy(func(args []string) bool {
		return len(args) == 0
	})).Return(nil)

	err := ts.runner.RunCmd("command")
	ts.Require().NoError(err)

	ts.runner.AssertExpectations(ts.T())
}

func TestMockRunnerTestSuite(t *testing.T) {
	suite.Run(t, new(MockRunnerTestSuite))
}
