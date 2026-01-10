package utils

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockExecutor implements pkgexec.Executor for testing
type mockExecutor struct {
	executeErr       error
	executeOutputErr error
	executeOutput    string
	executeCalls     [][]string
}

func (m *mockExecutor) Execute(_ context.Context, name string, args ...string) error {
	m.executeCalls = append(m.executeCalls, append([]string{name}, args...))
	return m.executeErr
}

func (m *mockExecutor) ExecuteOutput(_ context.Context, name string, args ...string) (string, error) {
	m.executeCalls = append(m.executeCalls, append([]string{name}, args...))
	return m.executeOutput, m.executeOutputErr
}

// errMockCommand is a static error for testing command failures
var errMockCommand = errors.New("command failed")

func TestSetExecutor(t *testing.T) {
	// Save original executor
	original := DefaultExecutor
	defer func() {
		DefaultExecutor = original
	}()

	// Create and set mock executor
	mock := &mockExecutor{}
	SetExecutor(mock)

	// Verify executor was set
	assert.Same(t, mock, DefaultExecutor)
}

func TestResetExecutor(t *testing.T) {
	// Save original executor
	original := DefaultExecutor
	defer func() {
		DefaultExecutor = original
	}()

	// Set a mock executor
	mock := &mockExecutor{}
	SetExecutor(mock)
	assert.Same(t, mock, DefaultExecutor)

	// Reset to default
	ResetExecutor()

	// Verify it's no longer the mock
	assert.NotSame(t, mock, DefaultExecutor)
	assert.NotNil(t, DefaultExecutor)
}

func TestRunCmdWithMockExecutor(t *testing.T) {
	// Save original executor
	original := DefaultExecutor
	defer func() {
		DefaultExecutor = original
	}()

	// Create mock executor
	mock := &mockExecutor{}
	SetExecutor(mock)

	// Run command
	err := RunCmd("echo", "hello", "world")
	require.NoError(t, err)

	// Verify the command was called
	require.Len(t, mock.executeCalls, 1)
	assert.Equal(t, []string{"echo", "hello", "world"}, mock.executeCalls[0])
}

func TestRunCmdWithMockExecutorError(t *testing.T) {
	// Save original executor
	original := DefaultExecutor
	defer func() {
		DefaultExecutor = original
	}()

	// Create mock executor that returns an error
	mock := &mockExecutor{executeErr: errMockCommand}
	SetExecutor(mock)

	// Run command
	err := RunCmd("failing-command", "arg1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "command failed")
}

func TestRunCmdOutputWithMockExecutor(t *testing.T) {
	// Save original executor
	original := DefaultExecutor
	defer func() {
		DefaultExecutor = original
	}()

	// Create mock executor with output
	mock := &mockExecutor{executeOutput: "test output"}
	SetExecutor(mock)

	// Run command
	output, err := RunCmdOutput("echo", "test")
	require.NoError(t, err)
	assert.Equal(t, "test output", output)

	// Verify the command was called
	require.Len(t, mock.executeCalls, 1)
	assert.Equal(t, []string{"echo", "test"}, mock.executeCalls[0])
}
