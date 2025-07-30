package testutil

import (
	"fmt"
	"strings"

	"github.com/stretchr/testify/mock"
)

// MockRunner provides enhanced mocking capabilities for command execution
type MockRunner struct {
	mock.Mock
}

// RunCmd executes a command and returns an error
func (m *MockRunner) RunCmd(name string, args ...string) error {
	arguments := m.Called(name, args)
	return arguments.Error(0)
}

// RunCmdOutput executes a command and returns output and error
func (m *MockRunner) RunCmdOutput(name string, args ...string) (string, error) {
	arguments := m.Called(name, args)
	return arguments.String(0), arguments.Error(1)
}

// MockBuilder provides a fluent interface for setting up common mock scenarios
type MockBuilder struct {
	runner *MockRunner
}

// NewMockRunner creates a new enhanced mock runner with builder
func NewMockRunner() (*MockRunner, *MockBuilder) {
	runner := &MockRunner{}
	builder := &MockBuilder{runner: runner}
	return runner, builder
}

// ExpectGitCommand sets up expectations for git commands
func (b *MockBuilder) ExpectGitCommand(subcommand string, output string, err error) *MockBuilder {
	b.runner.On("RunCmdOutput", "git", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == subcommand
	})).Return(output, err).Maybe()
	return b
}

// ExpectGoCommand sets up expectations for go commands
func (b *MockBuilder) ExpectGoCommand(subcommand string, err error) *MockBuilder {
	b.runner.On("RunCmd", "go", mock.MatchedBy(func(args []string) bool {
		return len(args) > 0 && args[0] == subcommand
	})).Return(err).Maybe()
	return b
}

// ExpectAnyCommand sets up a catch-all expectation for any command
func (b *MockBuilder) ExpectAnyCommand(err error) *MockBuilder {
	b.runner.On("RunCmd", mock.AnythingOfType("string"), mock.Anything).Return(err).Maybe()
	b.runner.On("RunCmdOutput", mock.AnythingOfType("string"), mock.Anything).Return("", err).Maybe()
	return b
}

// ExpectVersion sets up standard version-related git commands
func (b *MockBuilder) ExpectVersion(version string) *MockBuilder {
	b.ExpectGitCommand("describe", version, nil)
	b.ExpectGitCommand("rev-parse", "abc123", nil)
	return b
}

// ExpectSuccess sets up successful command execution expectations
func (b *MockBuilder) ExpectSuccess() *MockBuilder {
	return b.ExpectAnyCommand(nil)
}

// ExpectFailure sets up failing command execution expectations
func (b *MockBuilder) ExpectFailure(message string) *MockBuilder {
	return b.ExpectAnyCommand(fmt.Errorf("%s", message))
}

// Build returns the configured mock runner
func (b *MockBuilder) Build() *MockRunner {
	return b.runner
}

// CommandMatcher provides utilities for matching command arguments
type CommandMatcher struct{}

// ContainsArg returns a matcher that checks if args contain a specific argument
func (CommandMatcher) ContainsArg(arg string) func([]string) bool {
	return func(args []string) bool {
		for _, a := range args {
			if a == arg {
				return true
			}
		}
		return false
	}
}

// HasSubcommand returns a matcher that checks if the first arg matches
func (CommandMatcher) HasSubcommand(subcommand string) func([]string) bool {
	return func(args []string) bool {
		return len(args) > 0 && args[0] == subcommand
	}
}

// ContainsFlag returns a matcher that checks if args contain a flag
func (CommandMatcher) ContainsFlag(flag string) func([]string) bool {
	return func(args []string) bool {
		for _, a := range args {
			if strings.HasPrefix(a, flag) {
				return true
			}
		}
		return false
	}
}

// Global command matcher instance
var Cmd = CommandMatcher{} //nolint:gochecknoglobals // Test utility global
