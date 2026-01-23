package mage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

const (
	cmdPush = "push"
)

// TestVersionBumpViaCLI tests version:bump command via the CLI registry mechanism
func TestVersionBumpViaCLI(t *testing.T) {
	// Save original runner
	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore original runner: %v", err)
		}
	}()

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

	t.Run("CLIExecutionWithPushTrue", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		if err := SetRunner(mockRunner); err != nil {
			t.Fatalf("Failed to set mock runner: %v", err)
		}

		// Clear MAGE_ARGS to ensure we're testing direct argument passing
		if err := os.Unsetenv("MAGE_ARGS"); err != nil {
			t.Logf("Failed to unset MAGE_ARGS: %v", err)
		}

		// Mock git commands for version bump
		mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
		mockRunner.SetOutput("git tag --sort=-version:refname", "v1.3.27\nv1.3.26")                                                        // Highest tag in repo
		mockRunner.SetOutput("git describe --tags --abbrev=0", "v1.3.27")                                                                  // Reachable tag
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")                                                                    // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility

		// Create a registry and register the version:bump command like magex does
		reg := registry.NewRegistry()
		var v Version
		reg.MustRegister(
			registry.NewNamespaceCommand("version", "bump").
				WithDescription("Bump version with parameters: bump=<major|minor|patch> push dry-run force major-confirm").
				WithArgsFunc(func(args ...string) error { return v.Bump(args...) }).
				WithCategory("Version Management").
				WithUsage("magex version:bump [bump=<type>] [push] [dry-run] [force] [major-confirm]").
				MustBuild(),
		)

		// Execute the command like magex would
		args := []string{"push=true", "bump=patch"}
		err := reg.Execute("version:bump", args...)
		require.NoError(t, err)

		// Verify git tag command was called
		expectedTagCmd := []string{"git", "tag", "-a", "v1.3.28", "-m", "GitHubRelease v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedTagCmd),
			"Expected git tag command not found. Commands: %v", mockRunner.GetCommands())

		// Verify git push command was called
		expectedPushCmd := []string{"git", cmdPush, "origin", "v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedPushCmd),
			"Expected git push command not found. Commands: %v", mockRunner.GetCommands())
	})

	t.Run("CLIExecutionWithPushFlag", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		if err := SetRunner(mockRunner); err != nil {
			t.Fatalf("Failed to set mock runner: %v", err)
		}

		// Clear MAGE_ARGS to ensure we're testing direct argument passing
		if err := os.Unsetenv("MAGE_ARGS"); err != nil {
			t.Logf("Failed to unset MAGE_ARGS: %v", err)
		}

		// Mock git commands for version bump
		mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
		mockRunner.SetOutput("git tag --sort=-version:refname", "v1.3.27\nv1.3.26")                                                        // Highest tag in repo
		mockRunner.SetOutput("git describe --tags --abbrev=0", "v1.3.27")                                                                  // Reachable tag
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")                                                                    // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility

		// Create a registry and register the version:bump command like magex does
		reg := registry.NewRegistry()
		var v Version
		reg.MustRegister(
			registry.NewNamespaceCommand("version", "bump").
				WithDescription("Bump version with parameters: bump=<major|minor|patch> push dry-run force major-confirm").
				WithArgsFunc(func(args ...string) error { return v.Bump(args...) }).
				WithCategory("Version Management").
				WithUsage("magex version:bump [bump=<type>] [push] [dry-run] [force] [major-confirm]").
				MustBuild(),
		)

		// Execute the command like magex would with push as boolean flag
		args := []string{cmdPush, "bump=patch"}
		err := reg.Execute("version:bump", args...)
		require.NoError(t, err)

		// Verify git tag command was called
		expectedTagCmd := []string{"git", "tag", "-a", "v1.3.28", "-m", "GitHubRelease v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedTagCmd),
			"Expected git tag command not found. Commands: %v", mockRunner.GetCommands())

		// Verify git push command was called
		expectedPushCmd := []string{"git", cmdPush, "origin", "v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedPushCmd),
			"Expected git push command not found. Commands: %v", mockRunner.GetCommands())
	})

	t.Run("CLIExecutionNoPush", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		if err := SetRunner(mockRunner); err != nil {
			t.Fatalf("Failed to set mock runner: %v", err)
		}

		// Clear MAGE_ARGS to ensure we're testing direct argument passing
		if err := os.Unsetenv("MAGE_ARGS"); err != nil {
			t.Logf("Failed to unset MAGE_ARGS: %v", err)
		}

		// Mock git commands for version bump
		mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
		mockRunner.SetOutput("git tag --sort=-version:refname", "v1.3.27\nv1.3.26")                                                        // Highest tag in repo
		mockRunner.SetOutput("git describe --tags --abbrev=0", "v1.3.27")                                                                  // Reachable tag
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")                                                                    // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility

		// Create a registry and register the version:bump command like magex does
		reg := registry.NewRegistry()
		var v Version
		reg.MustRegister(
			registry.NewNamespaceCommand("version", "bump").
				WithDescription("Bump version with parameters: bump=<major|minor|patch> push dry-run force major-confirm").
				WithArgsFunc(func(args ...string) error { return v.Bump(args...) }).
				WithCategory("Version Management").
				WithUsage("magex version:bump [bump=<type>] [push] [dry-run] [force] [major-confirm]").
				MustBuild(),
		)

		// Execute the command like magex would without push parameter
		args := []string{"bump=patch"}
		err := reg.Execute("version:bump", args...)
		require.NoError(t, err)

		// Verify git tag command was called
		expectedTagCmd := []string{"git", "tag", "-a", "v1.3.28", "-m", "GitHubRelease v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedTagCmd),
			"Expected git tag command not found. Commands: %v", mockRunner.GetCommands())

		// Verify git push command was NOT called
		expectedPushCmd := []string{"git", cmdPush, "origin", "v1.3.28"}
		require.False(t, mockRunner.HasCommand(expectedPushCmd),
			"Git push command should not be called without push parameter")
	})
}

// TestDebugArgumentPassing helps debug how arguments are passed through the system
func TestDebugArgumentPassing(t *testing.T) {
	// This test is to understand exactly how arguments flow from CLI to Version.Bump

	t.Run("DirectFunctionCall", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() {
			if err := SetRunner(originalRunner); err != nil {
				t.Logf("Failed to restore original runner: %v", err)
			}
		}()

		mockRunner := NewVersionBumpMockRunner()
		if err := SetRunner(mockRunner); err != nil {
			t.Fatalf("Failed to set mock runner: %v", err)
		}

		// Mock git commands
		mockRunner.SetOutput("git status --porcelain", "")
		mockRunner.SetOutput("git tag --points-at HEAD", "")
		mockRunner.SetOutput("git tag --sort=-version:refname", "v1.3.27\nv1.3.26")
		mockRunner.SetOutput("git describe --tags --abbrev=0", "v1.3.27")
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)")
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")

		version := Version{}

		// Print arguments for debugging
		args := []string{"push=true", "bump=patch"}
		t.Logf("Calling Version.Bump with args: %v", args)

		err := version.Bump(args...)
		require.NoError(t, err)

		// Check commands executed
		commands := mockRunner.GetCommands()
		t.Logf("Commands executed: %v", commands)

		// Check specifically for push command
		hasPushCommand := false
		for _, cmd := range commands {
			const cmdGit = "git"
			if len(cmd) >= 3 && cmd[0] == cmdGit && cmd[1] == cmdPush {
				hasPushCommand = true
				t.Logf("Found push command: %v", cmd)
				break
			}
		}

		require.True(t, hasPushCommand, "Expected git push command to be executed")
	})

	t.Run("ViaRegistry", func(t *testing.T) {
		// Save original runner
		originalRunner := GetRunner()
		defer func() {
			if err := SetRunner(originalRunner); err != nil {
				t.Logf("Failed to restore original runner: %v", err)
			}
		}()

		mockRunner := NewVersionBumpMockRunner()
		if err := SetRunner(mockRunner); err != nil {
			t.Fatalf("Failed to set mock runner: %v", err)
		}

		// Mock git commands
		mockRunner.SetOutput("git status --porcelain", "")
		mockRunner.SetOutput("git tag --points-at HEAD", "")
		mockRunner.SetOutput("git tag --sort=-version:refname", "v1.3.27\nv1.3.26")
		mockRunner.SetOutput("git describe --tags --abbrev=0", "v1.3.27")
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)")
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")

		// Create a registry and register the version:bump command
		reg := registry.NewRegistry()
		var v Version

		reg.MustRegister(
			registry.NewNamespaceCommand("version", "bump").
				WithArgsFunc(func(args ...string) error {
					t.Logf("Registry calling Version.Bump with args: %v", args)
					return v.Bump(args...)
				}).
				MustBuild(),
		)

		// Execute via registry like magex does
		args := []string{"push=true", "bump=patch"}
		t.Logf("Registry.Execute called with command: version:bump, args: %v", args)

		err := reg.Execute("version:bump", args...)
		require.NoError(t, err)

		// Check commands executed
		commands := mockRunner.GetCommands()
		t.Logf("Commands executed via registry: %v", commands)

		// Check specifically for push command
		hasPushCommand := false
		for _, cmd := range commands {
			const cmdGit = "git"
			if len(cmd) >= 3 && cmd[0] == cmdGit && cmd[1] == cmdPush {
				hasPushCommand = true
				t.Logf("Found push command via registry: %v", cmd)
				break
			}
		}

		require.True(t, hasPushCommand, "Expected git push command to be executed via registry")
	})
}

// TestMAGE_ARGSEnvironmentVariable tests if MAGE_ARGS affects argument parsing
func TestMAGE_ARGSEnvironmentVariable(t *testing.T) {
	// Save original environment variable
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
	if err := SetRunner(mockRunner); err != nil {
		t.Fatalf("Failed to set mock runner: %v", err)
	}

	// Mock git commands
	mockRunner.SetOutput("git status --porcelain", "")
	mockRunner.SetOutput("git tag --points-at HEAD", "")
	mockRunner.SetOutput("git tag --sort=-version:refname", "v1.3.27\nv1.3.26")
	mockRunner.SetOutput("git describe --tags --abbrev=0", "v1.3.27")
	mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")
	mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)")
	mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")

	t.Run("WithMAGE_ARGS_Set", func(t *testing.T) {
		// Set MAGE_ARGS environment variable
		if err := os.Setenv("MAGE_ARGS", "push=false bump=minor"); err != nil {
			t.Fatalf("Failed to set MAGE_ARGS: %v", err)
		}

		version := Version{}

		// Call with empty args - should use MAGE_ARGS
		err := version.Bump()
		require.NoError(t, err)

		// Should NOT have push command because MAGE_ARGS has push=false
		commands := mockRunner.GetCommands()
		hasPushCommand := false
		for _, cmd := range commands {
			const cmdGit = "git"
			if len(cmd) >= 3 && cmd[0] == cmdGit && cmd[1] == cmdPush {
				hasPushCommand = true
				break
			}
		}
		require.False(t, hasPushCommand, "Should not have push command when MAGE_ARGS has push=false")
	})

	t.Run("WithArgsProvidedIgnoresMAGE_ARGS", func(t *testing.T) {
		// Reset mock runner
		mockRunner = NewVersionBumpMockRunner()
		if err := SetRunner(mockRunner); err != nil {
			t.Fatalf("Failed to set mock runner: %v", err)
		}
		mockRunner.SetOutput("git status --porcelain", "")
		mockRunner.SetOutput("git tag --points-at HEAD", "")
		mockRunner.SetOutput("git tag --sort=-version:refname", "v1.3.27\nv1.3.26")
		mockRunner.SetOutput("git describe --tags --abbrev=0", "v1.3.27")
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)")
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")

		// Set MAGE_ARGS environment variable
		if err := os.Setenv("MAGE_ARGS", "push=false bump=minor"); err != nil {
			t.Fatalf("Failed to set MAGE_ARGS: %v", err)
		}

		version := Version{}

		// Call with args provided - should ignore MAGE_ARGS
		err := version.Bump("push=true", "bump=patch")
		require.NoError(t, err)

		// Should HAVE push command because we provided push=true directly
		commands := mockRunner.GetCommands()
		hasPushCommand := false
		for _, cmd := range commands {
			const cmdGit = "git"
			if len(cmd) >= 3 && cmd[0] == cmdGit && cmd[1] == cmdPush {
				hasPushCommand = true
				break
			}
		}
		require.True(t, hasPushCommand, "Should have push command when args are provided directly")
	})
}
