package mage

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// VersionBumpMockRunner for testing version bump commands
type VersionBumpMockRunner struct {
	commands [][]string
	outputs  map[string]string
	errors   map[string]error
}

func NewVersionBumpMockRunner() *VersionBumpMockRunner {
	return &VersionBumpMockRunner{
		commands: make([][]string, 0),
		outputs:  make(map[string]string),
		errors:   make(map[string]error),
	}
}

func (mr *VersionBumpMockRunner) RunCmd(name string, args ...string) error {
	fullCmd := append([]string{name}, args...)
	mr.commands = append(mr.commands, fullCmd)

	cmdKey := strings.Join(fullCmd, " ")
	if err, exists := mr.errors[cmdKey]; exists {
		return err
	}
	return nil
}

func (mr *VersionBumpMockRunner) RunCmdOutput(name string, args ...string) (string, error) {
	fullCmd := append([]string{name}, args...)
	mr.commands = append(mr.commands, fullCmd)

	cmdKey := strings.Join(fullCmd, " ")
	if output, exists := mr.outputs[cmdKey]; exists {
		return output, nil
	}
	if err, exists := mr.errors[cmdKey]; exists {
		return "", err
	}
	return "", nil
}

func (mr *VersionBumpMockRunner) GetCommands() [][]string {
	return mr.commands
}

func (mr *VersionBumpMockRunner) SetOutput(command, output string) {
	mr.outputs[command] = output
}

func (mr *VersionBumpMockRunner) SetError(command string, err error) {
	mr.errors[command] = err
}

func (mr *VersionBumpMockRunner) HasCommand(expectedCmd []string) bool {
	expectedStr := strings.Join(expectedCmd, " ")
	for _, cmd := range mr.commands {
		if strings.Join(cmd, " ") == expectedStr {
			return true
		}
	}
	return false
}

// TestVersionBumpPushParameter tests the version:bump command with push parameter
func TestVersionBumpPushParameter(t *testing.T) {
	// Save original runner
	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore original runner: %v", err)
		}
	}()

	t.Run("PushParameterTrue", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock git commands for version bump
		mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.27-3-gabcdef")                                                 // Previous tag with distance
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")                                                                    // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility

		version := Version{}

		// Test with push=true parameter
		err := version.Bump("push=true", "bump=patch")
		require.NoError(t, err)

		// Verify git tag command was called
		expectedTagCmd := []string{"git", "tag", "-a", "v1.3.28", "-m", "GitHubRelease v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedTagCmd),
			"Expected git tag command not found. Commands: %v", mockRunner.GetCommands())

		// Verify git push command was called
		expectedPushCmd := []string{"git", "push", "origin", "v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedPushCmd),
			"Expected git push command not found. Commands: %v", mockRunner.GetCommands())
	})

	t.Run("PushParameterBooleanFlag", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock git commands for version bump
		mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.27-3-gabcdef")                                                 // Previous tag with distance
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")                                                                    // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility

		version := Version{}

		// Test with push as boolean flag
		err := version.Bump("push", "bump=patch")
		require.NoError(t, err)

		// Verify git tag command was called
		expectedTagCmd := []string{"git", "tag", "-a", "v1.3.28", "-m", "GitHubRelease v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedTagCmd),
			"Expected git tag command not found. Commands: %v", mockRunner.GetCommands())

		// Verify git push command was called
		expectedPushCmd := []string{"git", "push", "origin", "v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedPushCmd),
			"Expected git push command not found. Commands: %v", mockRunner.GetCommands())
	})

	t.Run("NoPushParameter", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock git commands for version bump
		mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.27-3-gabcdef")                                                 // Previous tag with distance
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")                                                                    // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility

		version := Version{}

		// Test without push parameter
		err := version.Bump("bump=patch")
		require.NoError(t, err)

		// Verify git tag command was called
		expectedTagCmd := []string{"git", "tag", "-a", "v1.3.28", "-m", "GitHubRelease v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedTagCmd),
			"Expected git tag command not found. Commands: %v", mockRunner.GetCommands())

		// Verify git push command was NOT called
		expectedPushCmd := []string{"git", "push", "origin", "v1.3.28"}
		require.False(t, mockRunner.HasCommand(expectedPushCmd),
			"Git push command should not be called without push parameter")
	})

	t.Run("DryRunWithPush", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock git commands for version bump
		mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.27-3-gabcdef")                                                 // Previous tag with distance
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")                                                                    // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility

		version := Version{}

		// Test with dry-run and push parameter
		err := version.Bump("push=true", "bump=patch", "dry-run")
		require.NoError(t, err)

		// In dry-run mode, no actual git tag creation or push commands should be executed
		commands := mockRunner.GetCommands()
		for _, cmd := range commands {
			// Check for actual tag creation command (with -a flag)
			if len(cmd) >= 4 && cmd[0] == "git" && cmd[1] == "tag" && cmd[2] == "-a" {
				t.Errorf("No git tag creation commands should be executed in dry-run mode, found: %v", cmd)
			}
			// Check for push commands
			if len(cmd) >= 2 && cmd[0] == "git" && cmd[1] == "push" {
				t.Errorf("No git push commands should be executed in dry-run mode, found: %v", cmd)
			}
		}
	})
}

// TestVersionBumpParameterParsing tests parameter parsing specifically for version bump
func TestVersionBumpParameterParsing(t *testing.T) {
	// Test the utils.ParseParams function with version bump arguments
	args := []string{"push=true", "bump=patch"}

	// Use the actual utils.ParseParams function
	params := utils.ParseParams(args)

	// Verify push parameter is parsed correctly
	require.Equal(t, "true", params["push"], "push parameter should be 'true'")
	require.Equal(t, "patch", params["bump"], "bump parameter should be 'patch'")

	// Test IsParamTrue function
	require.True(t, utils.IsParamTrue(params, "push"), "push parameter should be evaluated as true")
}

// TestVersionBumpDebugOutput tests the version bump with debug output to understand what's happening
func TestVersionBumpDebugOutput(t *testing.T) {
	// This test is designed to help debug the issue by capturing output
	// Save original environment variables
	origMageArgs := os.Getenv("MAGE_ARGS")
	defer func() {
		if origMageArgs == "" {
			if err := os.Unsetenv("MAGE_ARGS"); err != nil {
				t.Logf("Failed to unset MAGE_ARGS: %v", err)
			}
		} else {
			if err := os.Setenv("MAGE_ARGS", origMageArgs); err != nil {
				t.Logf("Failed to restore MAGE_ARGS: %v", err)
			}
		}
	}()

	// Save original runner
	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore original runner: %v", err)
		}
	}()

	mockRunner := NewVersionBumpMockRunner()
	require.NoError(t, SetRunner(mockRunner))

	// Mock git commands for version bump
	mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
	mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
	mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.27-3-gabcdef")                                                 // Previous tag with distance
	mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")                                                                    // Distance from tag
	mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
	mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility

	version := Version{}

	t.Run("DebugParameterHandling", func(t *testing.T) {
		// Clear MAGE_ARGS to ensure we're testing direct argument passing
		if err := os.Unsetenv("MAGE_ARGS"); err != nil {
			t.Logf("Failed to unset MAGE_ARGS: %v", err)
		}

		// Test the exact command that's failing
		args := []string{"push=true", "bump=patch"}

		fmt.Printf("Testing with args: %v\n", args)

		err := version.Bump(args...)
		require.NoError(t, err)

		// Print all commands executed for debugging
		commands := mockRunner.GetCommands()
		fmt.Printf("Commands executed: %v\n", commands)

		// Check if push command was executed
		hasPushCommand := false
		for _, cmd := range commands {
			if len(cmd) >= 3 && cmd[0] == "git" && cmd[1] == "push" {
				hasPushCommand = true
				fmt.Printf("Found push command: %v\n", cmd)
			}
		}

		if !hasPushCommand {
			t.Errorf("Expected git push command to be executed but it wasn't. All commands: %v", commands)
		}
	})
}
