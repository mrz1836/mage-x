package testhelpers

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Static test errors to satisfy err113 linter
var (
	errMockTest     = errors.New("mock test error")
	errMockCallback = errors.New("callback error")
	errMockDefault  = errors.New("default error")
	errMockSpecific = errors.New("specific error")
	errMockCommand  = errors.New("command error")
)

// mockTestingT implements TestingT for testing assertion methods
type mockTestingT struct {
	errors  []string
	helpers int
}

func (m *mockTestingT) Errorf(format string, args ...interface{}) {
	m.errors = append(m.errors, format)
}

func (m *mockTestingT) Helper() {
	m.helpers++
}

func (m *mockTestingT) hasError() bool {
	return len(m.errors) > 0
}

func (m *mockTestingT) reset() {
	m.errors = nil
	m.helpers = 0
}

// TestNewMockRunner verifies MockRunner initialization
func TestNewMockRunner(t *testing.T) {
	m := NewMockRunner()

	require.NotNil(t, m)
	require.NotNil(t, m.commands)
	require.NotNil(t, m.outputs)
	require.NotNil(t, m.errors)
	require.NotNil(t, m.callbacks)
	require.Empty(t, m.commands)
	require.Empty(t, m.outputs)
	require.Empty(t, m.errors)
	require.Empty(t, m.callbacks)
	require.NoError(t, m.defaultErr)
	require.Empty(t, m.defaultOut)
}

// TestMockRunnerRunCmd tests the RunCmd method
func TestMockRunnerRunCmd(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockRunner)
		cmd     string
		args    []string
		wantErr bool
	}{
		{
			name:    "successful command with no error configured",
			setup:   func(_ *MockRunner) {},
			cmd:     "echo",
			args:    []string{"hello"},
			wantErr: false,
		},
		{
			name: "command with configured error via SetError",
			setup: func(m *MockRunner) {
				m.SetError("echo", []string{"hello"}, errMockTest)
			},
			cmd:     "echo",
			args:    []string{"hello"},
			wantErr: true,
		},
		{
			name: "command with callback returning error",
			setup: func(m *MockRunner) {
				m.SetCallback("echo", []string{"hello"}, func(_ string, _ []string) (string, error) {
					return "", errMockCallback
				})
			},
			cmd:     "echo",
			args:    []string{"hello"},
			wantErr: true,
		},
		{
			name: "command with callback returning success",
			setup: func(m *MockRunner) {
				m.SetCallback("echo", []string{"hello"}, func(_ string, _ []string) (string, error) {
					return "output", nil
				})
			},
			cmd:     "echo",
			args:    []string{"hello"},
			wantErr: false,
		},
		{
			name: "command with default error",
			setup: func(m *MockRunner) {
				m.SetDefaultError(errMockDefault)
			},
			cmd:     "any",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "command with no args",
			setup:   func(_ *MockRunner) {},
			cmd:     "ls",
			args:    nil,
			wantErr: false,
		},
		{
			name: "command matches by name only when args not registered",
			setup: func(m *MockRunner) {
				m.SetErrorForCommand("git", errMockCommand)
			},
			cmd:     "git",
			args:    []string{"status"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMockRunner()
			tt.setup(m)

			err := m.RunCmd(tt.cmd, tt.args...)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Verify command was recorded
			cmds := m.GetCommands()
			require.Len(t, cmds, 1)
			require.Equal(t, tt.cmd, cmds[0].Name)
			require.Equal(t, tt.args, cmds[0].Args)
		})
	}
}

// TestMockRunnerRunCmdRecordsMultipleCommands verifies command recording
func TestMockRunnerRunCmdRecordsMultipleCommands(t *testing.T) {
	m := NewMockRunner()

	require.NoError(t, m.RunCmd("echo", "first"))
	require.NoError(t, m.RunCmd("echo", "second"))
	require.NoError(t, m.RunCmd("ls", "-la"))

	cmds := m.GetCommands()
	require.Len(t, cmds, 3)
	require.Equal(t, "echo", cmds[0].Name)
	require.Equal(t, []string{"first"}, cmds[0].Args)
	require.Equal(t, "echo", cmds[1].Name)
	require.Equal(t, []string{"second"}, cmds[1].Args)
	require.Equal(t, "ls", cmds[2].Name)
	require.Equal(t, []string{"-la"}, cmds[2].Args)
}

// TestMockRunnerRunCmdOutput tests the RunCmdOutput method
func TestMockRunnerRunCmdOutput(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*MockRunner)
		cmd        string
		args       []string
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "returns empty output with no configuration",
			setup:      func(_ *MockRunner) {},
			cmd:        "echo",
			args:       []string{"hello"},
			wantOutput: "",
			wantErr:    false,
		},
		{
			name: "returns configured output via SetOutput",
			setup: func(m *MockRunner) {
				m.SetOutput("echo", []string{"hello"}, "hello world")
			},
			cmd:        "echo",
			args:       []string{"hello"},
			wantOutput: "hello world",
			wantErr:    false,
		},
		{
			name: "returns configured output via SetOutputForCommand",
			setup: func(m *MockRunner) {
				m.SetOutputForCommand("git", "git version 2.0")
			},
			cmd:        "git",
			args:       []string{"--version"},
			wantOutput: "git version 2.0",
			wantErr:    false,
		},
		{
			name: "returns error via SetErrorForCommand",
			setup: func(m *MockRunner) {
				m.SetErrorForCommand("fail", errMockTest)
			},
			cmd:        "fail",
			args:       []string{},
			wantOutput: "",
			wantErr:    true,
		},
		{
			name: "returns error via SetError with matching args",
			setup: func(m *MockRunner) {
				m.SetError("fail", []string{"now"}, errMockTest)
			},
			cmd:        "fail",
			args:       []string{"now"},
			wantOutput: "",
			wantErr:    true,
		},
		{
			name: "returns output and error together with matching args",
			setup: func(m *MockRunner) {
				m.SetOutput("cmd", []string{"arg"}, "partial output")
				m.SetError("cmd", []string{"arg"}, errMockTest)
			},
			cmd:        "cmd",
			args:       []string{"arg"},
			wantOutput: "partial output",
			wantErr:    true,
		},
		{
			name: "callback takes precedence over SetOutput",
			setup: func(m *MockRunner) {
				m.SetOutput("echo", []string{"test"}, "from SetOutput")
				m.SetCallback("echo", []string{"test"}, func(_ string, _ []string) (string, error) {
					return "from callback", nil
				})
			},
			cmd:        "echo",
			args:       []string{"test"},
			wantOutput: "from callback",
			wantErr:    false,
		},
		{
			name: "returns default output when no specific match",
			setup: func(m *MockRunner) {
				m.SetDefaultOutput("default output")
			},
			cmd:        "any",
			args:       []string{"args"},
			wantOutput: "default output",
			wantErr:    false,
		},
		{
			name: "returns default error when no specific match",
			setup: func(m *MockRunner) {
				m.SetDefaultError(errMockDefault)
			},
			cmd:        "any",
			args:       []string{},
			wantOutput: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMockRunner()
			tt.setup(m)

			output, err := m.RunCmdOutput(tt.cmd, tt.args...)

			require.Equal(t, tt.wantOutput, output)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Verify command was recorded
			cmds := m.GetCommands()
			require.Len(t, cmds, 1)
			require.Equal(t, tt.cmd, cmds[0].Name)
		})
	}
}

// TestMockRunnerRunCmdInDir tests the RunCmdInDir method
func TestMockRunnerRunCmdInDir(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockRunner)
		dir     string
		cmd     string
		args    []string
		wantErr bool
	}{
		{
			name:    "successful command with directory",
			setup:   func(_ *MockRunner) {},
			dir:     "/home/user",
			cmd:     "ls",
			args:    []string{"-la"},
			wantErr: false,
		},
		{
			name: "command with error",
			setup: func(m *MockRunner) {
				m.SetError("git", []string{"status"}, errMockTest)
			},
			dir:     "/repo",
			cmd:     "git",
			args:    []string{"status"},
			wantErr: true,
		},
		{
			name: "command with callback",
			setup: func(m *MockRunner) {
				m.SetCallback("make", []string{"build"}, func(_ string, _ []string) (string, error) {
					return "", errMockCallback
				})
			},
			dir:     "/project",
			cmd:     "make",
			args:    []string{"build"},
			wantErr: true,
		},
		{
			name: "command with default error",
			setup: func(m *MockRunner) {
				m.SetDefaultError(errMockDefault)
			},
			dir:     "/any",
			cmd:     "any",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMockRunner()
			tt.setup(m)

			err := m.RunCmdInDir(tt.dir, tt.cmd, tt.args...)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Verify command was recorded with directory
			cmds := m.GetCommands()
			require.Len(t, cmds, 1)
			require.Equal(t, tt.cmd, cmds[0].Name)
			require.Equal(t, tt.args, cmds[0].Args)
			require.Equal(t, tt.dir, cmds[0].Dir)
		})
	}
}

// TestMockRunnerRunCmdOutputInDir tests the RunCmdOutputInDir method
func TestMockRunnerRunCmdOutputInDir(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*MockRunner)
		dir        string
		cmd        string
		args       []string
		wantOutput string
		wantErr    bool
	}{
		{
			name:       "returns empty output with no configuration",
			setup:      func(_ *MockRunner) {},
			dir:        "/home",
			cmd:        "pwd",
			args:       nil,
			wantOutput: "",
			wantErr:    false,
		},
		{
			name: "returns configured output",
			setup: func(m *MockRunner) {
				m.SetOutput("git", []string{"log"}, "commit abc123")
			},
			dir:        "/repo",
			cmd:        "git",
			args:       []string{"log"},
			wantOutput: "commit abc123",
			wantErr:    false,
		},
		{
			name: "returns error",
			setup: func(m *MockRunner) {
				m.SetError("git", []string{"push"}, errMockTest)
			},
			dir:        "/repo",
			cmd:        "git",
			args:       []string{"push"},
			wantOutput: "",
			wantErr:    true,
		},
		{
			name: "callback returns output",
			setup: func(m *MockRunner) {
				m.SetCallback("cat", []string{"file.txt"}, func(_ string, _ []string) (string, error) {
					return "file contents", nil
				})
			},
			dir:        "/home",
			cmd:        "cat",
			args:       []string{"file.txt"},
			wantOutput: "file contents",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMockRunner()
			tt.setup(m)

			output, err := m.RunCmdOutputInDir(tt.dir, tt.cmd, tt.args...)

			require.Equal(t, tt.wantOutput, output)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Verify command was recorded with directory
			cmds := m.GetCommands()
			require.Len(t, cmds, 1)
			require.Equal(t, tt.dir, cmds[0].Dir)
		})
	}
}

// TestMockRunnerRunCmdWithEnv tests the RunCmdWithEnv method
func TestMockRunnerRunCmdWithEnv(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*MockRunner)
		env     []string
		cmd     string
		args    []string
		wantErr bool
	}{
		{
			name:    "successful command with environment",
			setup:   func(_ *MockRunner) {},
			env:     []string{"PATH=/usr/bin", "HOME=/home/user"},
			cmd:     "env",
			args:    nil,
			wantErr: false,
		},
		{
			name: "command with error",
			setup: func(m *MockRunner) {
				m.SetError("docker", []string{"build"}, errMockTest)
			},
			env:     []string{"DOCKER_HOST=tcp://localhost:2375"},
			cmd:     "docker",
			args:    []string{"build"},
			wantErr: true,
		},
		{
			name: "command with callback",
			setup: func(m *MockRunner) {
				m.SetCallback("npm", []string{"install"}, func(_ string, _ []string) (string, error) {
					return "", errMockCallback
				})
			},
			env:     []string{"NODE_ENV=production"},
			cmd:     "npm",
			args:    []string{"install"},
			wantErr: true,
		},
		{
			name:    "empty environment",
			setup:   func(_ *MockRunner) {},
			env:     []string{},
			cmd:     "echo",
			args:    []string{"hello"},
			wantErr: false,
		},
		{
			name:    "nil environment",
			setup:   func(_ *MockRunner) {},
			env:     nil,
			cmd:     "echo",
			args:    []string{"hello"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMockRunner()
			tt.setup(m)

			err := m.RunCmdWithEnv(tt.env, tt.cmd, tt.args...)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			// Verify command was recorded with environment
			cmds := m.GetCommands()
			require.Len(t, cmds, 1)
			require.Equal(t, tt.cmd, cmds[0].Name)
			require.Equal(t, tt.args, cmds[0].Args)
			require.Equal(t, tt.env, cmds[0].Env)
		})
	}
}

// TestMockRunnerSetOutput tests the SetOutput method
func TestMockRunnerSetOutput(t *testing.T) {
	m := NewMockRunner()

	m.SetOutput("echo", []string{"hello"}, "hello world")

	output, err := m.RunCmdOutput("echo", "hello")
	require.NoError(t, err)
	require.Equal(t, "hello world", output)

	// Different args should not match
	output, err = m.RunCmdOutput("echo", "goodbye")
	require.NoError(t, err)
	require.Empty(t, output)
}

// TestMockRunnerSetError tests the SetError method
func TestMockRunnerSetError(t *testing.T) {
	m := NewMockRunner()

	m.SetError("fail", []string{"now"}, errMockTest)

	err := m.RunCmd("fail", "now")
	require.Error(t, err)
	require.Equal(t, errMockTest, err)

	// Different args should not match
	err = m.RunCmd("fail", "later")
	require.NoError(t, err)
}

// TestMockRunnerSetOutputForCommand tests command-level output
func TestMockRunnerSetOutputForCommand(t *testing.T) {
	m := NewMockRunner()

	m.SetOutputForCommand("git", "git output")

	// Should match regardless of args
	output, err := m.RunCmdOutput("git", "status")
	require.NoError(t, err)
	require.Equal(t, "git output", output)

	output, err = m.RunCmdOutput("git", "log", "--oneline")
	require.NoError(t, err)
	require.Equal(t, "git output", output)

	// Other commands should not match
	output, err = m.RunCmdOutput("make")
	require.NoError(t, err)
	require.Empty(t, output)
}

// TestMockRunnerSetErrorForCommand tests command-level errors
func TestMockRunnerSetErrorForCommand(t *testing.T) {
	m := NewMockRunner()

	m.SetErrorForCommand("docker", errMockCommand)

	// Should match regardless of args
	err := m.RunCmd("docker", "build")
	require.Error(t, err)
	require.Equal(t, errMockCommand, err)

	err = m.RunCmd("docker", "run", "-it", "ubuntu")
	require.Error(t, err)
	require.Equal(t, errMockCommand, err)

	// Other commands should not match
	err = m.RunCmd("make", "build")
	require.NoError(t, err)
}

// TestMockRunnerSetCallback tests callback configuration
func TestMockRunnerSetCallback(t *testing.T) {
	m := NewMockRunner()

	callCount := 0
	m.SetCallback("echo", []string{"count"}, func(cmd string, args []string) (string, error) {
		callCount++
		require.Equal(t, "echo", cmd)
		require.Equal(t, []string{"count"}, args)
		return "called", nil
	})

	output, err := m.RunCmdOutput("echo", "count")
	require.NoError(t, err)
	require.Equal(t, "called", output)
	require.Equal(t, 1, callCount)

	// Call again
	output, err = m.RunCmdOutput("echo", "count")
	require.NoError(t, err)
	require.Equal(t, "called", output)
	require.Equal(t, 2, callCount)
}

// TestMockRunnerSetCallbackForCommand tests command-level callbacks
func TestMockRunnerSetCallbackForCommand(t *testing.T) {
	m := NewMockRunner()

	callArgs := [][]string{}
	m.SetCallbackForCommand("git", func(_ string, args []string) (string, error) {
		callArgs = append(callArgs, args)
		return "git output", nil
	})

	output, err := m.RunCmdOutput("git", "status")
	require.NoError(t, err)
	require.Equal(t, "git output", output)

	output, err = m.RunCmdOutput("git", "log", "--oneline")
	require.NoError(t, err)
	require.Equal(t, "git output", output)

	require.Len(t, callArgs, 2)
	require.Equal(t, []string{"status"}, callArgs[0])
	require.Equal(t, []string{"log", "--oneline"}, callArgs[1])
}

// TestMockRunnerSetDefaultOutput tests default output configuration
func TestMockRunnerSetDefaultOutput(t *testing.T) {
	m := NewMockRunner()

	m.SetDefaultOutput("default")

	output, err := m.RunCmdOutput("any", "command")
	require.NoError(t, err)
	require.Equal(t, "default", output)

	// Specific output via SetOutputForCommand should take precedence
	m.SetOutputForCommand("specific", "specific output")
	output, err = m.RunCmdOutput("specific")
	require.NoError(t, err)
	require.Equal(t, "specific output", output)

	// Specific output via SetOutput with args should also take precedence
	m.SetOutput("echo", []string{"hello"}, "echo output")
	output, err = m.RunCmdOutput("echo", "hello")
	require.NoError(t, err)
	require.Equal(t, "echo output", output)
}

// TestMockRunnerSetDefaultError tests default error configuration
func TestMockRunnerSetDefaultError(t *testing.T) {
	m := NewMockRunner()

	m.SetDefaultError(errMockDefault)

	err := m.RunCmd("any", "command")
	require.Error(t, err)
	require.Equal(t, errMockDefault, err)

	// Specific error via SetErrorForCommand should take precedence
	m.SetErrorForCommand("specific", errMockSpecific)
	err = m.RunCmd("specific")
	require.Error(t, err)
	require.Equal(t, errMockSpecific, err)

	// Specific error via SetError with args should also take precedence
	m.SetError("fail", []string{"now"}, errMockTest)
	err = m.RunCmd("fail", "now")
	require.Error(t, err)
	require.Equal(t, errMockTest, err)
}

// TestMockRunnerGetCommands tests command retrieval
func TestMockRunnerGetCommands(t *testing.T) {
	m := NewMockRunner()

	// Initially empty
	cmds := m.GetCommands()
	require.Empty(t, cmds)

	// After running commands
	require.NoError(t, m.RunCmd("echo", "hello"))
	require.NoError(t, m.RunCmd("ls", "-la"))

	cmds = m.GetCommands()
	require.Len(t, cmds, 2)
	require.Equal(t, "echo", cmds[0].Name)
	require.Equal(t, "ls", cmds[1].Name)

	// Verify it returns a copy (modifying returned slice doesn't affect internal state)
	cmds[0].Name = "modified"
	newCmds := m.GetCommands()
	require.Equal(t, "echo", newCmds[0].Name)
}

// TestMockRunnerGetCommandCount tests command counting
func TestMockRunnerGetCommandCount(t *testing.T) {
	m := NewMockRunner()

	require.Equal(t, 0, m.GetCommandCount())

	require.NoError(t, m.RunCmd("echo", "1"))
	require.Equal(t, 1, m.GetCommandCount())

	require.NoError(t, m.RunCmd("echo", "2"))
	require.Equal(t, 2, m.GetCommandCount())

	require.NoError(t, m.RunCmd("echo", "3"))
	require.Equal(t, 3, m.GetCommandCount())
}

// TestMockRunnerGetLastCommand tests last command retrieval
func TestMockRunnerGetLastCommand(t *testing.T) {
	m := NewMockRunner()

	// Empty runner returns nil
	last := m.GetLastCommand()
	require.Nil(t, last)

	// After running commands
	require.NoError(t, m.RunCmd("echo", "first"))
	last = m.GetLastCommand()
	require.NotNil(t, last)
	require.Equal(t, "echo", last.Name)
	require.Equal(t, []string{"first"}, last.Args)

	require.NoError(t, m.RunCmd("ls", "-la"))
	last = m.GetLastCommand()
	require.NotNil(t, last)
	require.Equal(t, "ls", last.Name)
	require.Equal(t, []string{"-la"}, last.Args)

	// Verify it returns a copy
	last.Name = "modified"
	newLast := m.GetLastCommand()
	require.Equal(t, "ls", newLast.Name)
}

// TestMockRunnerFindCommand tests finding commands by name
func TestMockRunnerFindCommand(t *testing.T) {
	m := NewMockRunner()

	// Not found in empty runner
	cmd := m.FindCommand("echo")
	require.Nil(t, cmd)

	// After running commands
	require.NoError(t, m.RunCmd("echo", "first"))
	require.NoError(t, m.RunCmd("ls", "-la"))
	require.NoError(t, m.RunCmd("echo", "second"))

	// Finds first match
	cmd = m.FindCommand("echo")
	require.NotNil(t, cmd)
	require.Equal(t, "echo", cmd.Name)
	require.Equal(t, []string{"first"}, cmd.Args)

	// Finds other commands
	cmd = m.FindCommand("ls")
	require.NotNil(t, cmd)
	require.Equal(t, "ls", cmd.Name)

	// Returns nil for not found
	cmd = m.FindCommand("notfound")
	require.Nil(t, cmd)

	// Verify it returns a copy
	cmd = m.FindCommand("echo")
	cmd.Name = "modified"
	newCmd := m.FindCommand("echo")
	require.Equal(t, "echo", newCmd.Name)
}

// TestMockRunnerFindCommandWithArgs tests finding commands by name and args
func TestMockRunnerFindCommandWithArgs(t *testing.T) {
	m := NewMockRunner()

	// Not found in empty runner
	cmd := m.FindCommandWithArgs("echo", "hello")
	require.Nil(t, cmd)

	// After running commands
	require.NoError(t, m.RunCmd("echo", "hello"))
	require.NoError(t, m.RunCmd("echo", "world"))
	require.NoError(t, m.RunCmd("ls", "-la", "/home"))

	// Finds exact match
	cmd = m.FindCommandWithArgs("echo", "hello")
	require.NotNil(t, cmd)
	require.Equal(t, "echo", cmd.Name)
	require.Equal(t, []string{"hello"}, cmd.Args)

	cmd = m.FindCommandWithArgs("ls", "-la", "/home")
	require.NotNil(t, cmd)
	require.Equal(t, "ls", cmd.Name)

	// Returns nil for partial match
	cmd = m.FindCommandWithArgs("echo")
	require.Nil(t, cmd)

	cmd = m.FindCommandWithArgs("echo", "hello", "extra")
	require.Nil(t, cmd)

	// Returns nil for not found
	cmd = m.FindCommandWithArgs("notfound", "arg")
	require.Nil(t, cmd)
}

// TestMockRunnerAssertCalled tests the AssertCalled method
func TestMockRunnerAssertCalled(t *testing.T) {
	m := NewMockRunner()
	mt := &mockTestingT{}

	// Assert on empty runner should error
	m.AssertCalled(mt, "echo")
	require.True(t, mt.hasError())
	mt.reset()

	// After running command
	require.NoError(t, m.RunCmd("echo", "hello"))

	// Should not error when called
	m.AssertCalled(mt, "echo")
	require.False(t, mt.hasError())

	// Should error when not called
	m.AssertCalled(mt, "notcalled")
	require.True(t, mt.hasError())
}

// TestMockRunnerAssertCalledWith tests the AssertCalledWith method
func TestMockRunnerAssertCalledWith(t *testing.T) {
	m := NewMockRunner()
	mt := &mockTestingT{}

	require.NoError(t, m.RunCmd("echo", "hello", "world"))

	// Exact match should not error
	m.AssertCalledWith(mt, "echo", "hello", "world")
	require.False(t, mt.hasError())

	// Different args should error
	mt.reset()
	m.AssertCalledWith(mt, "echo", "hello")
	require.True(t, mt.hasError())

	// Different command should error
	mt.reset()
	m.AssertCalledWith(mt, "notcalled", "hello", "world")
	require.True(t, mt.hasError())
}

// TestMockRunnerAssertNotCalled tests the AssertNotCalled method
func TestMockRunnerAssertNotCalled(t *testing.T) {
	m := NewMockRunner()
	mt := &mockTestingT{}

	// Assert on empty runner should not error
	m.AssertNotCalled(mt, "echo")
	require.False(t, mt.hasError())

	// After running command
	require.NoError(t, m.RunCmd("echo", "hello"))

	// Should error when called
	m.AssertNotCalled(mt, "echo")
	require.True(t, mt.hasError())

	// Should not error for other commands
	mt.reset()
	m.AssertNotCalled(mt, "ls")
	require.False(t, mt.hasError())
}

// TestMockRunnerAssertCallCount tests the AssertCallCount method
func TestMockRunnerAssertCallCount(t *testing.T) {
	m := NewMockRunner()
	mt := &mockTestingT{}

	// Zero calls should match
	m.AssertCallCount(mt, "echo", 0)
	require.False(t, mt.hasError())

	// After running commands
	require.NoError(t, m.RunCmd("echo", "1"))
	require.NoError(t, m.RunCmd("echo", "2"))
	require.NoError(t, m.RunCmd("ls"))

	// Correct count should not error
	m.AssertCallCount(mt, "echo", 2)
	require.False(t, mt.hasError())

	m.AssertCallCount(mt, "ls", 1)
	require.False(t, mt.hasError())

	// Wrong count should error
	mt.reset()
	m.AssertCallCount(mt, "echo", 3)
	require.True(t, mt.hasError())

	mt.reset()
	m.AssertCallCount(mt, "echo", 1)
	require.True(t, mt.hasError())
}

// TestMockRunnerReset tests the Reset method
func TestMockRunnerReset(t *testing.T) {
	m := NewMockRunner()

	// Configure and run commands (without default error to allow commands to succeed)
	m.SetOutput("echo", []string{"hello"}, "output")
	m.SetErrorForCommand("fail", errMockTest)
	m.SetCallback("cb", []string{"arg"}, func(_ string, _ []string) (string, error) {
		return "", nil
	})
	m.SetDefaultOutput("default")
	require.NoError(t, m.RunCmd("echo", "hello"))
	require.NoError(t, m.RunCmd("ls"))

	require.Equal(t, 2, m.GetCommandCount())

	// Now set default error to verify it gets cleared
	m.SetDefaultError(errMockDefault)

	// Verify default error is set
	err := m.RunCmd("anycommand")
	require.Error(t, err)
	require.Equal(t, errMockDefault, err)

	// Reset
	m.Reset()

	// Verify everything is cleared
	require.Equal(t, 0, m.GetCommandCount())
	require.Empty(t, m.GetCommands())

	// Configured outputs/errors should be cleared
	output, err := m.RunCmdOutput("echo", "hello")
	require.NoError(t, err)
	require.Empty(t, output)

	// Default error should be cleared
	err = m.RunCmd("anycommand")
	require.NoError(t, err)

	// Specific error should be cleared
	err = m.RunCmd("fail")
	require.NoError(t, err)
}

// TestMockRunnerArgsMatch tests the argsMatch helper method
func TestMockRunnerArgsMatch(t *testing.T) {
	m := NewMockRunner()

	tests := []struct {
		name  string
		a     []string
		b     []string
		match bool
	}{
		{
			name:  "empty slices match",
			a:     []string{},
			b:     []string{},
			match: true,
		},
		{
			name:  "nil slices match",
			a:     nil,
			b:     nil,
			match: true,
		},
		{
			name:  "same elements match",
			a:     []string{"a", "b", "c"},
			b:     []string{"a", "b", "c"},
			match: true,
		},
		{
			name:  "different lengths don't match",
			a:     []string{"a"},
			b:     []string{"a", "b"},
			match: false,
		},
		{
			name:  "different elements don't match",
			a:     []string{"a", "b"},
			b:     []string{"a", "c"},
			match: false,
		},
		{
			name:  "empty vs non-empty don't match",
			a:     []string{},
			b:     []string{"a"},
			match: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := m.argsMatch(tt.a, tt.b)
			require.Equal(t, tt.match, result)
		})
	}
}

// TestMockRunnerCommandKey tests the commandKey helper method
func TestMockRunnerCommandKey(t *testing.T) {
	m := NewMockRunner()

	// Without configured outputs/errors, returns command name
	key := m.commandKey("echo", []string{"hello"})
	require.Equal(t, "echo", key)

	// With configured output for specific args, returns full key
	m.SetOutput("echo", []string{"hello"}, "output")
	key = m.commandKey("echo", []string{"hello"})
	require.Equal(t, "echo hello", key)

	// With configured error for specific args
	m.SetError("fail", []string{"now"}, errMockTest)
	key = m.commandKey("fail", []string{"now"})
	require.Equal(t, "fail now", key)

	// With configured callback for specific args
	m.SetCallback("cb", []string{"arg"}, func(_ string, _ []string) (string, error) {
		return "", nil
	})
	key = m.commandKey("cb", []string{"arg"})
	require.Equal(t, "cb arg", key)

	// Without args
	key = m.commandKey("ls", nil)
	require.Equal(t, "ls", key)
}

// TestMockRunnerExpect tests the Expect method
func TestMockRunnerExpect(t *testing.T) {
	m := NewMockRunner()
	mt := &mockTestingT{}

	matcher := m.Expect(mt)
	require.NotNil(t, matcher)
	require.Equal(t, m, matcher.runner)
	require.Equal(t, mt, matcher.t)
}

// TestCommandMatcherCommand tests the Command method
func TestCommandMatcherCommand(t *testing.T) {
	m := NewMockRunner()
	mt := &mockTestingT{}

	matcher := m.Expect(mt)
	assertion := matcher.Command("echo")

	require.NotNil(t, assertion)
	require.Equal(t, "echo", assertion.name)
	require.Nil(t, assertion.args)
}

// TestCommandAssertionWithArgs tests the WithArgs method
func TestCommandAssertionWithArgs(t *testing.T) {
	m := NewMockRunner()
	mt := &mockTestingT{}

	assertion := m.Expect(mt).Command("echo").WithArgs("hello", "world")

	require.NotNil(t, assertion)
	require.Equal(t, "echo", assertion.name)
	require.Equal(t, []string{"hello", "world"}, assertion.args)
}

// TestCommandAssertionCalled tests the Called method
func TestCommandAssertionCalled(t *testing.T) {
	m := NewMockRunner()
	mt := &mockTestingT{}

	// Not called - should error
	m.Expect(mt).Command("echo").Called()
	require.True(t, mt.hasError())

	// After running command
	mt.reset()
	require.NoError(t, m.RunCmd("echo", "hello"))
	m.Expect(mt).Command("echo").Called()
	require.False(t, mt.hasError())

	// With args - exact match
	mt.reset()
	m.Expect(mt).Command("echo").WithArgs("hello").Called()
	require.False(t, mt.hasError())

	// With args - no match
	mt.reset()
	m.Expect(mt).Command("echo").WithArgs("world").Called()
	require.True(t, mt.hasError())
}

// TestCommandAssertionNotCalled tests the NotCalled method
func TestCommandAssertionNotCalled(t *testing.T) {
	m := NewMockRunner()
	mt := &mockTestingT{}

	// Not called - should not error
	m.Expect(mt).Command("echo").NotCalled()
	require.False(t, mt.hasError())

	// After running command
	require.NoError(t, m.RunCmd("echo", "hello"))

	// Should error
	mt.reset()
	m.Expect(mt).Command("echo").NotCalled()
	require.True(t, mt.hasError())

	// Other commands should not error
	mt.reset()
	m.Expect(mt).Command("ls").NotCalled()
	require.False(t, mt.hasError())
}

// TestCommandAssertionCalledTimes tests the CalledTimes method
func TestCommandAssertionCalledTimes(t *testing.T) {
	m := NewMockRunner()
	mt := &mockTestingT{}

	// Zero times
	m.Expect(mt).Command("echo").CalledTimes(0)
	require.False(t, mt.hasError())

	// After running commands
	require.NoError(t, m.RunCmd("echo", "1"))
	require.NoError(t, m.RunCmd("echo", "2"))
	require.NoError(t, m.RunCmd("echo", "3"))

	// Correct count
	m.Expect(mt).Command("echo").CalledTimes(3)
	require.False(t, mt.hasError())

	// Wrong count
	mt.reset()
	m.Expect(mt).Command("echo").CalledTimes(2)
	require.True(t, mt.hasError())
}

// TestCommandAssertionCalledOnce tests the CalledOnce method
func TestCommandAssertionCalledOnce(t *testing.T) {
	m := NewMockRunner()
	mt := &mockTestingT{}

	// Not called - should error
	m.Expect(mt).Command("echo").CalledOnce()
	require.True(t, mt.hasError())

	// Called once
	mt.reset()
	require.NoError(t, m.RunCmd("echo"))
	m.Expect(mt).Command("echo").CalledOnce()
	require.False(t, mt.hasError())

	// Called twice - should error
	mt.reset()
	require.NoError(t, m.RunCmd("echo"))
	m.Expect(mt).Command("echo").CalledOnce()
	require.True(t, mt.hasError())
}

// TestCommandAssertionCalledTwice tests the CalledTwice method
func TestCommandAssertionCalledTwice(t *testing.T) {
	m := NewMockRunner()
	mt := &mockTestingT{}

	// Not called - should error
	m.Expect(mt).Command("echo").CalledTwice()
	require.True(t, mt.hasError())

	// Called once - should error
	mt.reset()
	require.NoError(t, m.RunCmd("echo"))
	m.Expect(mt).Command("echo").CalledTwice()
	require.True(t, mt.hasError())

	// Called twice
	mt.reset()
	require.NoError(t, m.RunCmd("echo"))
	m.Expect(mt).Command("echo").CalledTwice()
	require.False(t, mt.hasError())

	// Called three times - should error
	mt.reset()
	require.NoError(t, m.RunCmd("echo"))
	m.Expect(mt).Command("echo").CalledTwice()
	require.True(t, mt.hasError())
}

// TestMockRunnerConcurrentAccess tests thread safety
func TestMockRunnerConcurrentAccess(t *testing.T) {
	m := NewMockRunner()
	m.SetDefaultOutput("concurrent")

	var wg sync.WaitGroup
	const goroutines = 10
	const iterations = 100

	// Run commands concurrently
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Mix of different operations
				output, err := m.RunCmdOutput("echo", "test")
				assert.NoError(t, err)
				assert.Equal(t, "concurrent", output)

				assert.NoError(t, m.RunCmd("cmd", "arg"))
				_ = m.GetCommands()
				_ = m.GetCommandCount()
				_ = m.GetLastCommand()
				_ = m.FindCommand("echo")
			}
		}()
	}

	wg.Wait()

	// Verify command count
	expectedCount := goroutines * iterations * 2 // RunCmdOutput + RunCmd per iteration
	require.Equal(t, expectedCount, m.GetCommandCount())
}

// TestMockRunnerConcurrentConfiguration tests thread safety during configuration
func TestMockRunnerConcurrentConfiguration(t *testing.T) {
	m := NewMockRunner()

	var wg sync.WaitGroup
	const goroutines = 5

	// Configure concurrently
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cmd := "cmd"
			m.SetOutput(cmd, []string{"arg"}, "output")
			m.SetError(cmd, []string{"err"}, errMockTest)
			m.SetCallback(cmd, []string{"cb"}, func(_ string, _ []string) (string, error) {
				return "cb", nil
			})
			m.SetDefaultOutput("default")
			m.SetDefaultError(errMockDefault)
		}()
	}

	wg.Wait()

	// Should not panic and should have consistent state
	output, err := m.RunCmdOutput("cmd", "arg")
	require.NoError(t, err)
	require.Equal(t, "output", output)
}
