package mage

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test-specific constants and errors
const cmdGit = "git"

var (
	errBranchNotFoundLocally = errors.New("branch not found locally")
	errTagAlreadyExists      = errors.New("tag already exists")
	errRemoteRejected        = errors.New("remote rejected")
	errNetworkUnreachable    = errors.New("network unreachable")
	errNotGitRepository      = errors.New("not a git repository")
)

// TestVersionBumpWithBranchParameter tests the version:bump command with branch parameter
func TestVersionBumpWithBranchParameter(t *testing.T) {
	// Save original runner
	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore original runner: %v", err)
		}
	}()

	t.Run("BranchParameterSwitchAndSwitchBack", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock git commands for version bump with branch switching
		mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
		mockRunner.SetOutput("git branch --show-current", "gitbutler/workspace")                                                           // Current branch
		mockRunner.SetOutput("git branch -a", "  master\n* gitbutler/workspace\n  remotes/origin/master\n  remotes/origin/develop")        // Available branches
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.27-3-gabcdef")                                                 // Previous tag with distance
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")                                                                    // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility
		mockRunner.SetOutput("git ls-remote --tags origin v1.3.28", "")                                                                    // Tag doesn't exist on remote yet

		version := Version{}

		// Test with branch parameter to switch to master
		err := version.Bump("branch=master", "bump=patch", "push")
		require.NoError(t, err)

		commands := mockRunner.GetCommands()

		// Check that we got current branch
		expectedCurrentBranchCmd := []string{"git", "branch", "--show-current"}
		require.True(t, mockRunner.HasCommand(expectedCurrentBranchCmd),
			"Expected current branch command not found. Commands: %v", commands)

		// Check that we listed branches to validate target branch
		expectedBranchListCmd := []string{"git", "branch", "-a"}
		require.True(t, mockRunner.HasCommand(expectedBranchListCmd),
			"Expected branch list command not found. Commands: %v", commands)

		// Check that we checked out the target branch
		expectedCheckoutCmd := []string{"git", "checkout", "master"}
		require.True(t, mockRunner.HasCommand(expectedCheckoutCmd),
			"Expected checkout command not found. Commands: %v", commands)

		// Check that we fetched and pulled latest changes (with --tags --force to get all remote tags)
		expectedFetchCmd := []string{"git", "fetch", "--tags", "--force", "origin"}
		require.True(t, mockRunner.HasCommand(expectedFetchCmd),
			"Expected fetch --tags --force command not found. Commands: %v", commands)

		expectedPullCmd := []string{"git", "pull", "--rebase", "origin"}
		require.True(t, mockRunner.HasCommand(expectedPullCmd),
			"Expected pull command not found. Commands: %v", commands)

		// Check version bump operations
		expectedTagCmd := []string{"git", "tag", "-a", "v1.3.28", "-m", "GitHubRelease v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedTagCmd),
			"Expected git tag command not found. Commands: %v", commands)

		expectedPushCmd := []string{"git", "push", "origin", "v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedPushCmd),
			"Expected git push command not found. Commands: %v", commands)

		// Check that we switched back to original branch (should be at the end)
		expectedSwitchBackCmd := []string{"git", "checkout", "gitbutler/workspace"}
		require.True(t, mockRunner.HasCommand(expectedSwitchBackCmd),
			"Expected switch back command not found. Commands: %v", commands)
	})

	t.Run("BranchParameterSameBranch", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock git commands - already on target branch
		mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
		mockRunner.SetOutput("git branch --show-current", "master")                                                                        // Already on master
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.27-3-gabcdef")                                                 // Previous tag with distance
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")                                                                    // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility

		version := Version{}

		// Test with branch parameter for same branch
		err := version.Bump("branch=master", "bump=patch")
		require.NoError(t, err)

		commands := mockRunner.GetCommands()

		// Should NOT have checkout command since we're already on the right branch
		expectedCheckoutCmd := []string{"git", "checkout", "master"}
		require.False(t, mockRunner.HasCommand(expectedCheckoutCmd),
			"Should not checkout when already on target branch. Commands: %v", commands)

		// Should still pull latest changes (with --tags --force to get all remote tags)
		expectedFetchCmd := []string{"git", "fetch", "--tags", "--force", "origin"}
		require.True(t, mockRunner.HasCommand(expectedFetchCmd),
			"Expected fetch --tags --force command not found. Commands: %v", commands)

		expectedPullCmd := []string{"git", "pull", "--rebase", "origin"}
		require.True(t, mockRunner.HasCommand(expectedPullCmd),
			"Expected pull command not found. Commands: %v", commands)
	})

	t.Run("NoBranchParameterWarning", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock git commands for normal version bump
		mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
		mockRunner.SetOutput("git branch --show-current", "gitbutler/workspace")                                                           // Current branch for warning
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.27-3-gabcdef")                                                 // Previous tag with distance
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")                                                                    // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility

		version := Version{}

		// Test without branch parameter - should show warning but still work
		err := version.Bump("bump=patch")
		require.NoError(t, err)

		commands := mockRunner.GetCommands()

		// Should get current branch for warning
		expectedCurrentBranchCmd := []string{"git", "branch", "--show-current"}
		require.True(t, mockRunner.HasCommand(expectedCurrentBranchCmd),
			"Expected current branch command not found. Commands: %v", commands)

		// Should NOT have checkout commands
		checkoutCommands := []string{}
		for _, cmd := range commands {
			if len(cmd) >= 2 && cmd[0] == cmdGit && cmd[1] == "checkout" {
				checkoutCommands = append(checkoutCommands, strings.Join(cmd, " "))
			}
		}
		require.Empty(t, checkoutCommands, "Should not have any checkout commands when no branch parameter provided")

		// Should still perform version bump
		expectedTagCmd := []string{"git", "tag", "-a", "v1.3.28", "-m", "GitHubRelease v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedTagCmd),
			"Expected git tag command not found. Commands: %v", commands)
	})

	t.Run("InvalidBranchParameter", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock git commands
		mockRunner.SetOutput("git status --porcelain", "")                                                        // Clean working directory
		mockRunner.SetOutput("git branch --show-current", "master")                                               // Current branch
		mockRunner.SetOutput("git branch -a", "  master\n* main\n  remotes/origin/master\n  remotes/origin/main") // Available branches (no 'nonexistent')

		version := Version{}

		// Test with invalid branch parameter
		err := version.Bump("branch=nonexistent", "bump=patch")
		require.Error(t, err)
		require.Contains(t, err.Error(), "branch does not exist locally or remotely: 'nonexistent'")
	})

	t.Run("UncommittedChangesWithBranch", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock uncommitted changes
		mockRunner.SetOutput("git status --porcelain", "M some-file.go\n?? new-file.go") // Uncommitted changes
		mockRunner.SetOutput("git branch --show-current", "master")                      // Current branch

		version := Version{}

		// Test with uncommitted changes - should fail before branch operations
		err := version.Bump("branch=develop", "bump=patch")
		require.Error(t, err)
		require.Equal(t, errVersionUncommittedChanges, err)

		commands := mockRunner.GetCommands()

		// Should NOT have any checkout commands since we failed early
		checkoutCommands := []string{}
		for _, cmd := range commands {
			if len(cmd) >= 2 && cmd[0] == cmdGit && cmd[1] == "checkout" {
				checkoutCommands = append(checkoutCommands, strings.Join(cmd, " "))
			}
		}
		require.Empty(t, checkoutCommands, "Should not have any checkout commands when uncommitted changes exist")
	})

	t.Run("DryRunWithBranchParameter", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock git commands for dry-run
		mockRunner.SetOutput("git status --porcelain", "")                                                                          // Clean working directory
		mockRunner.SetOutput("git branch --show-current", "gitbutler/workspace")                                                    // Current branch
		mockRunner.SetOutput("git branch -a", "  master\n* gitbutler/workspace\n  remotes/origin/master\n  remotes/origin/develop") // Available branches
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                        // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.27-3-gabcdef")                                          // Previous tag with distance
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")                                                             // Distance from tag

		version := Version{}

		// Test dry-run with branch parameter
		err := version.Bump("branch=master", "bump=patch", "dry-run")
		require.NoError(t, err)

		commands := mockRunner.GetCommands()

		// Should NOT have actual checkout, fetch, pull, tag, or push commands in dry-run
		forbiddenCommands := [][]string{
			{"git", "checkout", "master"},
			{"git", "fetch", "--tags", "--force", "origin"},
			{"git", "pull", "--rebase", "origin"},
			{"git", "tag", "-a", "v1.3.28", "-m", "GitHubRelease v1.3.28"},
			{"git", "push", "origin", "v1.3.28"},
			{"git", "checkout", "gitbutler/workspace"},
		}

		for _, forbiddenCmd := range forbiddenCommands {
			require.False(t, mockRunner.HasCommand(forbiddenCmd),
				"Dry-run should not execute command: %v. Commands: %v", forbiddenCmd, commands)
		}

		// Should still check branches for validation
		expectedCurrentBranchCmd := []string{"git", "branch", "--show-current"}
		require.True(t, mockRunner.HasCommand(expectedCurrentBranchCmd),
			"Expected current branch command not found. Commands: %v", commands)
	})

	t.Run("RemoteBranchCheckout", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock git commands where local branch doesn't exist but remote does
		mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
		mockRunner.SetOutput("git branch --show-current", "master")                                                                        // Current branch
		mockRunner.SetOutput("git branch -a", "  master\n* main\n  remotes/origin/master\n  remotes/origin/develop")                       // develop only exists remotely
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.27-3-gabcdef")                                                 // Previous tag with distance
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")                                                                    // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility

		// Make local checkout fail, so it tries remote checkout
		mockRunner.SetError("git checkout develop", errBranchNotFoundLocally)

		version := Version{}

		// Test with remote branch
		err := version.Bump("branch=develop", "bump=patch")
		require.NoError(t, err)

		commands := mockRunner.GetCommands()

		// Should try local checkout first
		expectedLocalCheckoutCmd := []string{"git", "checkout", "develop"}
		require.True(t, mockRunner.HasCommand(expectedLocalCheckoutCmd),
			"Expected local checkout command not found. Commands: %v", commands)

		// Should then try remote checkout
		expectedRemoteCheckoutCmd := []string{"git", "checkout", "-b", "develop", "origin/develop"}
		require.True(t, mockRunner.HasCommand(expectedRemoteCheckoutCmd),
			"Expected remote checkout command not found. Commands: %v", commands)
	})

	t.Run("TagCreationFailureAfterBranchSwitch", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock successful branch operations
		mockRunner.SetOutput("git status --porcelain", "")                                                                                    // Clean working directory
		mockRunner.SetOutput("git branch --show-current", "gitbutler/workspace")                                                              // Current branch
		mockRunner.SetOutput("git branch -a", "  master\\n* gitbutler/workspace\\n  remotes/origin/master\\n  remotes/origin/develop")        // Available branches
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                                  // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.27-3-gabcdef")                                                    // Previous tag with distance
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")                                                                       // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\\tgit@github.com:test/repo.git (fetch)\\norigin\\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\\trefs/heads/main")                                             // Mock remote accessibility

		// Mock tag creation failure
		mockRunner.SetError("git tag -a v1.3.28 -m GitHubRelease v1.3.28", errTagAlreadyExists)

		version := Version{}

		// Test that tag creation failure still switches back to original branch
		err := version.Bump("branch=master", "bump=patch")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to create tag")

		commands := mockRunner.GetCommands()

		// Should have switched to master
		expectedCheckoutCmd := []string{"git", "checkout", "master"}
		require.True(t, mockRunner.HasCommand(expectedCheckoutCmd),
			"Expected checkout command not found. Commands: %v", commands)

		// Should have attempted tag creation
		expectedTagCmd := []string{"git", "tag", "-a", "v1.3.28", "-m", "GitHubRelease v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedTagCmd),
			"Expected tag command not found. Commands: %v", commands)

		// Should have switched back to original branch (defer should execute)
		expectedSwitchBackCmd := []string{"git", "checkout", "gitbutler/workspace"}
		require.True(t, mockRunner.HasCommand(expectedSwitchBackCmd),
			"Expected switch back command not found. Commands: %v", commands)
	})

	t.Run("PushFailureAfterTagCreation", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock successful operations until push
		mockRunner.SetOutput("git status --porcelain", "")                                                                                    // Clean working directory
		mockRunner.SetOutput("git branch --show-current", "gitbutler/workspace")                                                              // Current branch
		mockRunner.SetOutput("git branch -a", "  master\\n* gitbutler/workspace\\n  remotes/origin/master\\n  remotes/origin/develop")        // Available branches
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                                  // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.27-3-gabcdef")                                                    // Previous tag with distance
		mockRunner.SetOutput("git rev-list --count v1.3.27..HEAD", "3")                                                                       // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\\tgit@github.com:test/repo.git (fetch)\\norigin\\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\\trefs/heads/main")                                             // Mock remote accessibility

		// Mock push failure
		mockRunner.SetError("git push origin v1.3.28", errRemoteRejected)

		version := Version{}

		// Test that push failure still switches back to original branch
		err := version.Bump("branch=master", "bump=patch", "push")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to push tag")

		commands := mockRunner.GetCommands()

		// Should have switched to master
		expectedCheckoutCmd := []string{"git", "checkout", "master"}
		require.True(t, mockRunner.HasCommand(expectedCheckoutCmd),
			"Expected checkout command not found. Commands: %v", commands)

		// Should have created tag successfully
		expectedTagCmd := []string{"git", "tag", "-a", "v1.3.28", "-m", "GitHubRelease v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedTagCmd),
			"Expected tag command not found. Commands: %v", commands)

		// Should have attempted push
		expectedPushCmd := []string{"git", "push", "origin", "v1.3.28"}
		require.True(t, mockRunner.HasCommand(expectedPushCmd),
			"Expected push command not found. Commands: %v", commands)

		// Should have switched back to original branch (defer should execute)
		expectedSwitchBackCmd := []string{"git", "checkout", "gitbutler/workspace"}
		require.True(t, mockRunner.HasCommand(expectedSwitchBackCmd),
			"Expected switch back command not found. Commands: %v", commands)
	})

	t.Run("NetworkFailureDuringPull", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock successful branch operations until pull
		mockRunner.SetOutput("git status --porcelain", "")                                                                             // Clean working directory
		mockRunner.SetOutput("git branch --show-current", "gitbutler/workspace")                                                       // Current branch
		mockRunner.SetOutput("git branch -a", "  master\\n* gitbutler/workspace\\n  remotes/origin/master\\n  remotes/origin/develop") // Available branches

		// Mock network failure during pull
		mockRunner.SetError("git pull --rebase origin", errNetworkUnreachable)

		version := Version{}

		// Test that pull failure returns error and doesn't continue
		err := version.Bump("branch=master", "bump=patch")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to pull latest changes")

		commands := mockRunner.GetCommands()

		// Should have switched to master
		expectedCheckoutCmd := []string{"git", "checkout", "master"}
		require.True(t, mockRunner.HasCommand(expectedCheckoutCmd),
			"Expected checkout command not found. Commands: %v", commands)

		// Should have attempted pull
		expectedPullCmd := []string{"git", "pull", "--rebase", "origin"}
		require.True(t, mockRunner.HasCommand(expectedPullCmd),
			"Expected pull command not found. Commands: %v", commands)

		// Should have switched back to original branch (defer should execute even on pull failure)
		expectedSwitchBackCmd := []string{"git", "checkout", "gitbutler/workspace"}
		require.True(t, mockRunner.HasCommand(expectedSwitchBackCmd),
			"Expected switch back command not found. Commands: %v", commands)

		// Should NOT have attempted tag creation after pull failure
		forbiddenTagCmd := []string{"git", "tag", "-a", "v1.3.28", "-m", "GitHubRelease v1.3.28"}
		require.False(t, mockRunner.HasCommand(forbiddenTagCmd),
			"Should not attempt tag creation after pull failure. Commands: %v", commands)
	})
}

// TestBranchHelperFunctions tests the individual branch helper functions
func TestBranchHelperFunctions(t *testing.T) {
	// Save original runner
	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore original runner: %v", err)
		}
	}()

	t.Run("GetCurrentBranch", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		mockRunner.SetOutput("git branch --show-current", "feature/test-branch")

		branch, err := getCurrentBranch()
		require.NoError(t, err)
		require.Equal(t, "feature/test-branch", branch)
	})

	t.Run("GetCurrentBranchError", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		mockRunner.SetError("git branch --show-current", errNotGitRepository)

		_, err := getCurrentBranch()
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get current branch")
	})

	t.Run("IsValidBranchExists", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		mockRunner.SetOutput("git branch -a", "  master\n* main\n  feature/test\n  remotes/origin/master\n  remotes/origin/develop")

		// Test local branch
		err := isValidBranch("master")
		require.NoError(t, err)

		// Test current branch
		err = isValidBranch("main")
		require.NoError(t, err)

		// Test remote branch
		err = isValidBranch("develop")
		require.NoError(t, err)
	})

	t.Run("IsValidBranchNotExists", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		mockRunner.SetOutput("git branch -a", "  master\n* main\n  remotes/origin/master")

		err := isValidBranch("nonexistent")
		require.Error(t, err)
		require.Contains(t, err.Error(), "branch does not exist locally or remotely: 'nonexistent'")
	})
}
