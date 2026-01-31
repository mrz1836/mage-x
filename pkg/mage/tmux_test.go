//go:build unit
// +build unit

package mage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tmuxTestRunner is a mock runner for tmux tests
type tmuxTestRunner struct {
	runCmdFunc       func(cmd string, args ...string) error
	runCmdOutputFunc func(cmd string, args ...string) (string, error)
	capturedCmd      string
	capturedArgs     []string
}

func (r *tmuxTestRunner) RunCmd(name string, args ...string) error {
	r.capturedCmd = name
	r.capturedArgs = args
	if r.runCmdFunc != nil {
		return r.runCmdFunc(name, args...)
	}
	return nil
}

func (r *tmuxTestRunner) RunCmdOutput(name string, args ...string) (string, error) {
	r.capturedCmd = name
	r.capturedArgs = args
	if r.runCmdOutputFunc != nil {
		return r.runCmdOutputFunc(name, args...)
	}
	return "", nil
}

// TestTmux_List_NoSessions tests the List command with no sessions
func TestTmux_List_NoSessions(t *testing.T) {
	// Save original runner and restore after test
	originalRunner := GetRunner()
	defer func() { _ = SetRunner(originalRunner) }()

	// Create mock runner that simulates no sessions
	mockRunner := &tmuxTestRunner{
		runCmdOutputFunc: func(cmd string, args ...string) (string, error) {
			if cmd == "tmux" && len(args) > 0 && args[0] == "ls" {
				// Return error code 1 with "no server running" message
				return "no server running on /tmp/tmux-501/default", fmt.Errorf("exit status 1")
			}
			return "", nil
		},
	}
	_ = SetRunner(mockRunner)

	// Run List command
	var tmux Tmux
	err := tmux.List()

	// Should not return error (handled gracefully)
	require.NoError(t, err)
}

// TestTmux_List_WithSessions tests the List command with active sessions
func TestTmux_List_WithSessions(t *testing.T) {
	originalRunner := GetRunner()
	defer func() { _ = SetRunner(originalRunner) }()

	mockRunner := &tmuxTestRunner{
		runCmdOutputFunc: func(cmd string, args ...string) (string, error) {
			if cmd == "tmux" && len(args) > 0 && args[0] == "ls" {
				return "session1: 1 windows (created Sat Jan 31 15:30:00 2026) [80x24]\nsession2: 2 windows (created Sat Jan 31 14:00:00 2026) [120x40] (attached)", nil
			}
			return "", nil
		},
	}
	_ = SetRunner(mockRunner)

	var tmux Tmux
	err := tmux.List()
	require.NoError(t, err)
}

// TestCheckTmux tests the tmux installation check
func TestCheckTmux(t *testing.T) {
	t.Run("tmux installed", func(t *testing.T) {
		// This test will pass if tmux is actually installed
		// In a real environment, we'd mock exec.LookPath
		err := checkTmux()
		// Can't assert without mocking exec.LookPath, so just run it
		_ = err
	})
}

// TestValidateModel tests model validation
func TestValidateModel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"known model - sonnet", "sonnet", "claude-sonnet-4-5"},
		{"known model - opus", "opus", "claude-opus-4-5"},
		{"known model - haiku", "haiku", "claude-haiku-4-5"},
		{"unknown model", "custom-model", "custom-model"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validateModel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetSessionName tests session name generation
func TestGetSessionName(t *testing.T) {
	t.Run("explicit name provided", func(t *testing.T) {
		params := map[string]string{
			"name": "my-session",
		}
		name, err := getSessionName(params)
		require.NoError(t, err)
		assert.Equal(t, "my-session", name)
	})

	t.Run("use directory basename", func(t *testing.T) {
		params := map[string]string{
			"dir": "/path/to/project",
		}
		name, err := getSessionName(params)
		require.NoError(t, err)
		assert.Equal(t, "project", name)
	})

	t.Run("use current directory", func(t *testing.T) {
		params := map[string]string{}
		name, err := getSessionName(params)
		require.NoError(t, err)
		// Should return current directory basename
		assert.NotEmpty(t, name)
	})
}

// TestGetExpandedDir tests directory expansion
func TestGetExpandedDir(t *testing.T) {
	t.Run("expand home directory", func(t *testing.T) {
		params := map[string]string{
			"dir": "~/projects",
		}
		expanded, err := getExpandedDir(params)
		require.NoError(t, err)
		assert.Contains(t, expanded, "projects")
		assert.NotContains(t, expanded, "~")
	})

	t.Run("absolute path unchanged", func(t *testing.T) {
		params := map[string]string{
			"dir": "/absolute/path",
		}
		expanded, err := getExpandedDir(params)
		require.NoError(t, err)
		assert.Equal(t, "/absolute/path", expanded)
	})

	t.Run("empty dir uses current", func(t *testing.T) {
		params := map[string]string{}
		expanded, err := getExpandedDir(params)
		require.NoError(t, err)
		assert.NotEmpty(t, expanded)
	})
}

// TestGetSupportedModelNames tests the model name listing
func TestGetSupportedModelNames(t *testing.T) {
	names := getSupportedModelNames()
	require.NotEmpty(t, names)
	assert.Contains(t, names, "opus")
	assert.Contains(t, names, "sonnet")
	assert.Contains(t, names, "haiku")
}

// TestGetSupportedModels tests the supported models map
func TestGetSupportedModels(t *testing.T) {
	models := getSupportedModels()
	require.NotEmpty(t, models)

	// Test Anthropic models
	assert.Equal(t, "claude-opus-4-5", models["opus"])
	assert.Equal(t, "claude-sonnet-4-5", models["sonnet"])
	assert.Equal(t, "claude-haiku-4-5", models["haiku"])

	// Test OpenAI models
	assert.Equal(t, "gpt-4o", models["gpt"])
	assert.Equal(t, "o1", models["o1"])
	assert.Equal(t, "o3-mini", models["o3"])

	// Test Google models
	assert.Equal(t, "gemini-3-flash", models["gemini"])
}

// TestTmux_Attach_NoName tests Attach with missing session name
func TestTmux_Attach_NoName(t *testing.T) {
	originalRunner := GetRunner()
	defer func() { _ = SetRunner(originalRunner) }()

	mockRunner := &tmuxTestRunner{}
	_ = SetRunner(mockRunner)

	var tmux Tmux
	err := tmux.Attach()

	// Should return error for missing session name
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

// TestTmux_Attach_SessionNotFound tests Attach with non-existent session
func TestTmux_Attach_SessionNotFound(t *testing.T) {
	originalRunner := GetRunner()
	defer func() { _ = SetRunner(originalRunner) }()

	mockRunner := &tmuxTestRunner{
		runCmdFunc: func(cmd string, args ...string) error {
			if cmd == "tmux" && args[0] == "has-session" {
				return fmt.Errorf("session not found")
			}
			return nil
		},
	}
	_ = SetRunner(mockRunner)

	var tmux Tmux
	err := tmux.Attach("name=nonexistent")

	// Should return error for non-existent session
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

// TestTmux_Kill_NoName tests Kill with missing session name
func TestTmux_Kill_NoName(t *testing.T) {
	originalRunner := GetRunner()
	defer func() { _ = SetRunner(originalRunner) }()

	mockRunner := &tmuxTestRunner{}
	_ = SetRunner(mockRunner)

	var tmux Tmux
	err := tmux.Kill()

	// Should return error for missing session name
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")
}

// TestTmux_Kill_SessionNotFound tests Kill with non-existent session
func TestTmux_Kill_SessionNotFound(t *testing.T) {
	originalRunner := GetRunner()
	defer func() { _ = SetRunner(originalRunner) }()

	mockRunner := &tmuxTestRunner{
		runCmdFunc: func(cmd string, args ...string) error {
			if cmd == "tmux" && args[0] == "has-session" {
				return fmt.Errorf("session not found")
			}
			return nil
		},
	}
	_ = SetRunner(mockRunner)

	var tmux Tmux
	err := tmux.Kill("name=nonexistent")

	// Should return error for non-existent session
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session not found")
}

// TestTmux_KillAll_NoSessions tests KillAll with no sessions
func TestTmux_KillAll_NoSessions(t *testing.T) {
	originalRunner := GetRunner()
	defer func() { _ = SetRunner(originalRunner) }()

	mockRunner := &tmuxTestRunner{
		runCmdOutputFunc: func(cmd string, args ...string) (string, error) {
			if cmd == "tmux" && args[0] == "ls" {
				return "no server running", fmt.Errorf("exit status 1")
			}
			return "", nil
		},
	}
	_ = SetRunner(mockRunner)

	var tmux Tmux
	err := tmux.KillAll()

	// Should not return error (handled gracefully)
	require.NoError(t, err)
}

// TestTmux_Status_NoSessions tests Status with no sessions
func TestTmux_Status_NoSessions(t *testing.T) {
	originalRunner := GetRunner()
	defer func() { _ = SetRunner(originalRunner) }()

	mockRunner := &tmuxTestRunner{
		runCmdOutputFunc: func(cmd string, args ...string) (string, error) {
			if cmd == "tmux" && args[0] == "ls" {
				return "no server running", fmt.Errorf("exit status 1")
			}
			return "", nil
		},
	}
	_ = SetRunner(mockRunner)

	var tmux Tmux
	err := tmux.Status()

	// Should not return error (handled gracefully)
	require.NoError(t, err)
}

// TestTmux_Status_WithSessions tests Status with active sessions
func TestTmux_Status_WithSessions(t *testing.T) {
	originalRunner := GetRunner()
	defer func() { _ = SetRunner(originalRunner) }()

	mockRunner := &tmuxTestRunner{
		runCmdOutputFunc: func(cmd string, args ...string) (string, error) {
			if cmd == "tmux" && args[0] == "ls" {
				return "session1: 1 windows (created Sat Jan 31 15:30:00 2026) [80x24]\nsession2: 2 windows (created Sat Jan 31 14:00:00 2026) [120x40] (attached)", nil
			}
			return "", nil
		},
	}
	_ = SetRunner(mockRunner)

	var tmux Tmux
	err := tmux.Status()
	require.NoError(t, err)
}
