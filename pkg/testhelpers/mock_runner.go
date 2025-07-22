package testhelpers

import (
	"fmt"
	"strings"
	"sync"
)

// MockRunner implements CommandRunner interface for testing
type MockRunner struct {
	mu         sync.Mutex
	commands   []ExecutedCommand
	outputs    map[string]string
	errors     map[string]error
	callbacks  map[string]func(cmd string, args []string) (string, error)
	defaultErr error
	defaultOut string
}

// ExecutedCommand represents a command that was executed
type ExecutedCommand struct {
	Name string
	Args []string
	Dir  string
	Env  []string
}

// NewMockRunner creates a new mock runner
func NewMockRunner() *MockRunner {
	return &MockRunner{
		commands:  []ExecutedCommand{},
		outputs:   make(map[string]string),
		errors:    make(map[string]error),
		callbacks: make(map[string]func(cmd string, args []string) (string, error)),
	}
}

// RunCmd implements CommandRunner.RunCmd
func (m *MockRunner) RunCmd(name string, args ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.commands = append(m.commands, ExecutedCommand{
		Name: name,
		Args: args,
	})

	key := m.commandKey(name, args)

	// Check for callback
	if cb, ok := m.callbacks[key]; ok {
		_, err := cb(name, args)
		return err
	}

	// Check for specific error
	if err, ok := m.errors[key]; ok {
		return err
	}

	// Return default error
	return m.defaultErr
}

// RunCmdOutput implements CommandRunner.RunCmdOutput
func (m *MockRunner) RunCmdOutput(name string, args ...string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.commands = append(m.commands, ExecutedCommand{
		Name: name,
		Args: args,
	})

	key := m.commandKey(name, args)

	// Check for callback
	if cb, ok := m.callbacks[key]; ok {
		return cb(name, args)
	}

	// Check for specific output and error
	output, hasOutput := m.outputs[key]
	err, hasError := m.errors[key]

	if hasOutput || hasError {
		return output, err
	}

	// Return defaults
	return m.defaultOut, m.defaultErr
}

// RunCmdInDir implements CommandRunner.RunCmdInDir
func (m *MockRunner) RunCmdInDir(dir, name string, args ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.commands = append(m.commands, ExecutedCommand{
		Name: name,
		Args: args,
		Dir:  dir,
	})

	key := m.commandKey(name, args)

	// Check for callback
	if cb, ok := m.callbacks[key]; ok {
		_, err := cb(name, args)
		return err
	}

	// Check for specific error
	if err, ok := m.errors[key]; ok {
		return err
	}

	// Return default error
	return m.defaultErr
}

// RunCmdOutputInDir implements CommandRunner.RunCmdOutputInDir
func (m *MockRunner) RunCmdOutputInDir(dir, name string, args ...string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.commands = append(m.commands, ExecutedCommand{
		Name: name,
		Args: args,
		Dir:  dir,
	})

	key := m.commandKey(name, args)

	// Check for callback
	if cb, ok := m.callbacks[key]; ok {
		return cb(name, args)
	}

	// Check for specific output and error
	output, hasOutput := m.outputs[key]
	err, hasError := m.errors[key]

	if hasOutput || hasError {
		return output, err
	}

	// Return defaults
	return m.defaultOut, m.defaultErr
}

// RunCmdWithEnv implements CommandRunner.RunCmdWithEnv
func (m *MockRunner) RunCmdWithEnv(env []string, name string, args ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.commands = append(m.commands, ExecutedCommand{
		Name: name,
		Args: args,
		Env:  env,
	})

	key := m.commandKey(name, args)

	// Check for callback
	if cb, ok := m.callbacks[key]; ok {
		_, err := cb(name, args)
		return err
	}

	// Check for specific error
	if err, ok := m.errors[key]; ok {
		return err
	}

	// Return default error
	return m.defaultErr
}

// SetOutput sets the output for a specific command
func (m *MockRunner) SetOutput(cmd string, args []string, output string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.commandKeyWithArgs(cmd, args)
	m.outputs[key] = output
}

// SetError sets the error for a specific command
func (m *MockRunner) SetError(cmd string, args []string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.commandKeyWithArgs(cmd, args)
	m.errors[key] = err
}

// SetOutputForCommand sets output for any invocation of a command
func (m *MockRunner) SetOutputForCommand(cmd string, output string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.outputs[cmd] = output
}

// SetErrorForCommand sets error for any invocation of a command
func (m *MockRunner) SetErrorForCommand(cmd string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.errors[cmd] = err
}

// SetCallback sets a callback function for a specific command
func (m *MockRunner) SetCallback(cmd string, args []string, callback func(cmd string, args []string) (string, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := m.commandKeyWithArgs(cmd, args)
	m.callbacks[key] = callback
}

// SetCallbackForCommand sets a callback for any invocation of a command
func (m *MockRunner) SetCallbackForCommand(cmd string, callback func(cmd string, args []string) (string, error)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.callbacks[cmd] = callback
}

// SetDefaultOutput sets the default output for all commands
func (m *MockRunner) SetDefaultOutput(output string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.defaultOut = output
}

// SetDefaultError sets the default error for all commands
func (m *MockRunner) SetDefaultError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.defaultErr = err
}

// GetCommands returns all executed commands
func (m *MockRunner) GetCommands() []ExecutedCommand {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make([]ExecutedCommand, len(m.commands))
	copy(result, m.commands)
	return result
}

// GetCommandCount returns the number of executed commands
func (m *MockRunner) GetCommandCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	return len(m.commands)
}

// GetLastCommand returns the last executed command
func (m *MockRunner) GetLastCommand() *ExecutedCommand {
	m.mu.Lock()
	defer m.mu.Unlock()

	if len(m.commands) == 0 {
		return nil
	}

	cmd := m.commands[len(m.commands)-1]
	return &cmd
}

// FindCommand finds a command by name
func (m *MockRunner) FindCommand(name string) *ExecutedCommand {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cmd := range m.commands {
		if cmd.Name == name {
			cmdCopy := cmd
			return &cmdCopy
		}
	}

	return nil
}

// FindCommandWithArgs finds a command by name and args
func (m *MockRunner) FindCommandWithArgs(name string, args ...string) *ExecutedCommand {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cmd := range m.commands {
		if cmd.Name == name && m.argsMatch(cmd.Args, args) {
			cmdCopy := cmd
			return &cmdCopy
		}
	}

	return nil
}

// AssertCalled asserts that a command was called
func (m *MockRunner) AssertCalled(t TestingT, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cmd := range m.commands {
		if cmd.Name == name {
			return
		}
	}

	t.Errorf("Expected command '%s' to be called, but it wasn't", name)
}

// AssertCalledWith asserts that a command was called with specific args
func (m *MockRunner) AssertCalledWith(t TestingT, name string, args ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cmd := range m.commands {
		if cmd.Name == name && m.argsMatch(cmd.Args, args) {
			return
		}
	}

	t.Errorf("Expected command '%s %s' to be called, but it wasn't",
		name, strings.Join(args, " "))
}

// AssertNotCalled asserts that a command was not called
func (m *MockRunner) AssertNotCalled(t TestingT, name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, cmd := range m.commands {
		if cmd.Name == name {
			t.Errorf("Expected command '%s' to not be called, but it was", name)
			return
		}
	}
}

// AssertCallCount asserts the number of times a command was called
func (m *MockRunner) AssertCallCount(t TestingT, name string, expected int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	for _, cmd := range m.commands {
		if cmd.Name == name {
			count++
		}
	}

	if count != expected {
		t.Errorf("Expected command '%s' to be called %d times, but was called %d times",
			name, expected, count)
	}
}

// Reset clears all recorded commands and settings
func (m *MockRunner) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.commands = []ExecutedCommand{}
	m.outputs = make(map[string]string)
	m.errors = make(map[string]error)
	m.callbacks = make(map[string]func(cmd string, args []string) (string, error))
	m.defaultErr = nil
	m.defaultOut = ""
}

// commandKey generates a key for command lookup
func (m *MockRunner) commandKey(name string, args []string) string {
	// First check for exact match with args
	if len(args) > 0 {
		key := m.commandKeyWithArgs(name, args)
		if _, hasOutput := m.outputs[key]; hasOutput {
			return key
		}
		if _, hasError := m.errors[key]; hasError {
			return key
		}
		if _, hasCallback := m.callbacks[key]; hasCallback {
			return key
		}
	}

	// Fall back to command name only
	return name
}

// commandKeyWithArgs generates a key including args
func (m *MockRunner) commandKeyWithArgs(name string, args []string) string {
	return fmt.Sprintf("%s %s", name, strings.Join(args, " "))
}

// argsMatch checks if two arg slices match
func (m *MockRunner) argsMatch(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// TestingT is a subset of testing.T for assertions
type TestingT interface {
	Errorf(format string, args ...interface{})
	Helper()
}

// CommandMatcher provides fluent assertions for commands
type CommandMatcher struct {
	runner *MockRunner
	t      TestingT
}

// Expect returns a CommandMatcher for fluent assertions
func (m *MockRunner) Expect(t TestingT) *CommandMatcher {
	return &CommandMatcher{runner: m, t: t}
}

// Command starts a command assertion
func (cm *CommandMatcher) Command(name string) *CommandAssertion {
	return &CommandAssertion{
		matcher: cm,
		name:    name,
	}
}

// CommandAssertion provides assertions for a specific command
type CommandAssertion struct {
	matcher *CommandMatcher
	name    string
	args    []string
}

// WithArgs sets expected args
func (ca *CommandAssertion) WithArgs(args ...string) *CommandAssertion {
	ca.args = args
	return ca
}

// Called asserts the command was called
func (ca *CommandAssertion) Called() {
	ca.matcher.t.Helper()

	if ca.args != nil {
		ca.matcher.runner.AssertCalledWith(ca.matcher.t, ca.name, ca.args...)
	} else {
		ca.matcher.runner.AssertCalled(ca.matcher.t, ca.name)
	}
}

// NotCalled asserts the command was not called
func (ca *CommandAssertion) NotCalled() {
	ca.matcher.t.Helper()
	ca.matcher.runner.AssertNotCalled(ca.matcher.t, ca.name)
}

// CalledTimes asserts the command was called n times
func (ca *CommandAssertion) CalledTimes(n int) {
	ca.matcher.t.Helper()
	ca.matcher.runner.AssertCallCount(ca.matcher.t, ca.name, n)
}

// CalledOnce asserts the command was called exactly once
func (ca *CommandAssertion) CalledOnce() {
	ca.CalledTimes(1)
}

// CalledTwice asserts the command was called exactly twice
func (ca *CommandAssertion) CalledTwice() {
	ca.CalledTimes(2)
}
