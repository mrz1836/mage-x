package mage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	errNetworkError    = errors.New("network error")
	errWouldClobberTag = errors.New("! [rejected] v1.4.0 -> v1.4.0 (would clobber existing tag)")
)

// TestVersionBumpWithRemoteTagsNotFetched tests version bump when remote has tags not fetched locally
func TestVersionBumpWithRemoteTagsNotFetched(t *testing.T) {
	// Save original runner
	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore original runner: %v", err)
		}
	}()

	t.Run("RemoteHasNewerTagNotFetchedLocally", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Simulate scenario where remote has v1.4.0 but local doesn't
		// Before fetch: local only knows about v1.3.0
		// After fetch --tags: local should detect v1.4.0 exists on remote

		// Mock git commands for version bump
		mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.4.0-5-gabcdef")                                                  // After fetch, should find v1.4.0
		mockRunner.SetOutput("git rev-list --count v1.4.0..HEAD", "5")                                                                     // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility

		// Simulate branch checkout scenario (when branch parameter is used)
		mockRunner.SetOutput("git branch --show-current", "feature-branch")                                     // Current branch
		mockRunner.SetOutput("git branch -a", "* feature-branch\n  master\n  remotes/origin/master")            // Branch list
		mockRunner.SetOutput("git log --oneline -5 --no-decorate", "abc123 Recent commit\ndef456 Other commit") // Recent commits

		version := Version{}

		// Test version bump with branch parameter (which triggers pullLatestBranch with --tags)
		err := version.Bump("branch=master", "bump=patch")
		require.NoError(t, err)

		// Verify git fetch --tags was called (not just git fetch)
		expectedFetchCmd := []string{"git", "fetch", "--tags", "origin"}
		require.True(t, mockRunner.HasCommand(expectedFetchCmd),
			"Expected 'git fetch --tags origin' command to be called. Commands: %v", mockRunner.GetCommands())

		// Verify the version was bumped from v1.4.0 (remote tag) not v1.3.0 (old local tag)
		expectedTagCmd := []string{"git", "tag", "-a", "v1.4.1", "-m", "GitHubRelease v1.4.1"}
		require.True(t, mockRunner.HasCommand(expectedTagCmd),
			"Expected version to be bumped from v1.4.0 to v1.4.1 after fetching remote tags. Commands: %v", mockRunner.GetCommands())
	})
}

// TestVersionBumpTagExistsOnRemote tests detection of tags that already exist on remote
func TestVersionBumpTagExistsOnRemote(t *testing.T) {
	// Save original runner
	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore original runner: %v", err)
		}
	}()

	t.Run("TagAlreadyExistsOnRemote", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Simulate scenario where:
		// 1. Local has v1.3.0
		// 2. Version bump creates v1.4.0
		// 3. But v1.4.0 already exists on remote (from another branch)

		// Mock git commands for version bump
		mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.0-22-gabcdef")                                                 // Current version v1.3.0
		mockRunner.SetOutput("git rev-list --count v1.3.0..HEAD", "22")                                                                    // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility
		mockRunner.SetOutput("git log --oneline -5 --no-decorate", "abc123 Recent commit\ndef456 Other commit")                            // Recent commits

		// CRITICAL: Mock that v1.4.0 already exists on remote
		mockRunner.SetOutput("git ls-remote --tags origin v1.4.0", "98abdabb5b928ada967550c3218ea0faf7cc40b7\trefs/tags/v1.4.0")

		version := Version{}

		// Test version bump with push parameter
		err := version.Bump("push=true", "bump=minor")
		require.Error(t, err, "Expected error when tag already exists on remote")
		require.Contains(t, err.Error(), "already exists on remote", "Error should mention tag already exists on remote")

		// Verify git ls-remote check was performed
		expectedCheckCmd := []string{"git", "ls-remote", "--tags", "origin", "v1.4.0"}
		require.True(t, mockRunner.HasCommand(expectedCheckCmd),
			"Expected 'git ls-remote --tags origin v1.4.0' check to be performed. Commands: %v", mockRunner.GetCommands())

		// Verify git push was NOT called (because check failed)
		expectedPushCmd := []string{"git", "push", "origin", "v1.4.0"}
		require.False(t, mockRunner.HasCommand(expectedPushCmd),
			"Git push should not be called when tag already exists on remote. Commands: %v", mockRunner.GetCommands())
	})

	t.Run("TagDoesNotExistOnRemote", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Simulate normal scenario where tag doesn't exist on remote

		// Mock git commands for version bump
		mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.0-22-gabcdef")                                                 // Current version v1.3.0
		mockRunner.SetOutput("git rev-list --count v1.3.0..HEAD", "22")                                                                    // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility
		mockRunner.SetOutput("git log --oneline -5 --no-decorate", "abc123 Recent commit\ndef456 Other commit")                            // Recent commits

		// Mock that v1.4.0 does NOT exist on remote (empty response)
		mockRunner.SetOutput("git ls-remote --tags origin v1.4.0", "")

		version := Version{}

		// Test version bump with push parameter
		err := version.Bump("push=true", "bump=minor")
		require.NoError(t, err, "Should succeed when tag doesn't exist on remote")

		// Verify git ls-remote check was performed
		expectedCheckCmd := []string{"git", "ls-remote", "--tags", "origin", "v1.4.0"}
		require.True(t, mockRunner.HasCommand(expectedCheckCmd),
			"Expected 'git ls-remote --tags origin v1.4.0' check to be performed. Commands: %v", mockRunner.GetCommands())

		// Verify git push WAS called (because check passed)
		expectedPushCmd := []string{"git", "push", "origin", "v1.4.0"}
		require.True(t, mockRunner.HasCommand(expectedPushCmd),
			"Git push should be called when tag doesn't exist on remote. Commands: %v", mockRunner.GetCommands())
	})
}

// TestPullLatestBranchFetchesTags tests that pullLatestBranch fetches all tags
func TestPullLatestBranchFetchesTags(t *testing.T) {
	// Save original runner
	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore original runner: %v", err)
		}
	}()

	t.Run("FetchWithTagsFlag", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock successful fetch and pull
		mockRunner.SetOutput("git fetch --tags origin", "")
		mockRunner.SetOutput("git pull --rebase origin", "Already up to date.")

		// Call pullLatestBranch directly
		err := pullLatestBranch()
		require.NoError(t, err)

		// Verify git fetch --tags was called
		expectedFetchCmd := []string{"git", "fetch", "--tags", "origin"}
		require.True(t, mockRunner.HasCommand(expectedFetchCmd),
			"Expected 'git fetch --tags origin' to be called, not 'git fetch origin'. Commands: %v", mockRunner.GetCommands())
	})

	t.Run("FetchFailure", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock fetch failure
		mockRunner.SetError("git fetch --tags origin", errNetworkError)

		// Call pullLatestBranch directly
		err := pullLatestBranch()
		require.Error(t, err, "Should return error when fetch fails")
		require.Contains(t, err.Error(), "failed to fetch from origin", "Error should indicate fetch failure")
	})

	t.Run("TagConflictRetryWithForce", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock initial fetch failure with "would clobber" error
		mockRunner.SetError("git fetch --tags origin", errWouldClobberTag)

		// Mock successful force fetch
		mockRunner.SetOutput("git fetch --tags --force origin", "")
		mockRunner.SetOutput("git pull --rebase origin", "Already up to date.")

		// Call pullLatestBranch directly
		err := pullLatestBranch()
		require.NoError(t, err, "Should succeed when retrying with --force")

		// Verify both commands were called in sequence
		commands := mockRunner.GetCommands()

		// Should have tried normal fetch first
		expectedFetchCmd := []string{"git", "fetch", "--tags", "origin"}
		require.True(t, mockRunner.HasCommand(expectedFetchCmd),
			"Expected 'git fetch --tags origin' to be tried first. Commands: %v", commands)

		// Should have retried with force
		expectedForceFetchCmd := []string{"git", "fetch", "--tags", "--force", "origin"}
		require.True(t, mockRunner.HasCommand(expectedForceFetchCmd),
			"Expected 'git fetch --tags --force origin' to be called after failure. Commands: %v", commands)
	})
}

// TestVersionBumpCompleteScenario tests the complete scenario from the user's bug report
func TestVersionBumpCompleteScenario(t *testing.T) {
	// Save original runner
	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore original runner: %v", err)
		}
	}()

	t.Run("UserReportedBugScenario", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Simulate EXACT scenario from user's bug report:
		// - On master branch
		// - Local has v1.3.0
		// - 22 commits since v1.3.0
		// - Remote has v1.4.0 (created from another branch, not fetched locally)
		// - User runs: magex version:bump push=true bump=minor branch=master

		// Mock git commands
		mockRunner.SetOutput("git status --porcelain", "")                                                                                                                                                                                                                                                       // Clean working directory
		mockRunner.SetOutput("git branch --show-current", "gitbutler/workspace")                                                                                                                                                                                                                                 // Current branch (GitButler workspace)
		mockRunner.SetOutput("git branch -a", "* gitbutler/workspace\n  master\n  remotes/origin/master")                                                                                                                                                                                                        // Branch list
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                                                                                                                                                                                                     // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.0-22-g78baa5e")                                                                                                                                                                                                                      // v1.3.0 with 22 commits ahead
		mockRunner.SetOutput("git rev-list --count v1.3.0..HEAD", "22")                                                                                                                                                                                                                                          // Distance from tag
		mockRunner.SetOutput("git log --oneline -5 --no-decorate", "78baa5e docs(examples): add P2PKH validation example (#38)\n4918f74 fix(tests): reuse genesis validation formats (#37)\ne7c2455 sync: update 9 files from source repository (#36)\n5b8a1ec fix: minor fixes\nb284f69 feat: upgraded readme") // Recent commits
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:bsv-blockchain/go-chaincfg.git (fetch)\norigin\tgit@github.com:bsv-blockchain/go-chaincfg.git (push)")                                                                                                                                     // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "78baa5e3d7bd355019db6caaa11286779aba6832\trefs/heads/master")                                                                                                                                                                             // Mock remote accessibility

		// CRITICAL: After fetch --tags, local should discover v1.4.0 exists on remote
		mockRunner.SetOutput("git ls-remote --tags origin v1.4.0", "98abdabb5b928ada967550c3218ea0faf7cc40b7\trefs/tags/v1.4.0")

		version := Version{}

		// Run the exact command from user's bug report
		err := version.Bump("push=true", "bump=minor", "branch=master")

		// Should get error about tag already existing on remote
		require.Error(t, err, "Should fail when trying to push tag that exists on remote")
		require.Contains(t, err.Error(), "already exists on remote", "Error should mention tag already exists")

		// Verify the correct sequence of commands
		commands := mockRunner.GetCommands()

		// Should have fetched with --tags
		expectedFetchCmd := []string{"git", "fetch", "--tags", "origin"}
		require.True(t, mockRunner.HasCommand(expectedFetchCmd),
			"Should fetch tags with --tags flag. Commands: %v", commands)

		// Should have checked if tag exists on remote
		expectedCheckCmd := []string{"git", "ls-remote", "--tags", "origin", "v1.4.0"}
		require.True(t, mockRunner.HasCommand(expectedCheckCmd),
			"Should check if tag exists on remote before pushing. Commands: %v", commands)

		// Should NOT have pushed (because check detected existing tag)
		expectedPushCmd := []string{"git", "push", "origin", "v1.4.0"}
		require.False(t, mockRunner.HasCommand(expectedPushCmd),
			"Should not push when tag already exists on remote. Commands: %v", commands)
	})
}
