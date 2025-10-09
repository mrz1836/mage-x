package mage

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

var errNetworkError = errors.New("network error")

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

		// Verify git fetch --tags --force was called (not just git fetch)
		expectedFetchCmd := []string{"git", "fetch", "--tags", "--force", "origin"}
		require.True(t, mockRunner.HasCommand(expectedFetchCmd),
			"Expected 'git fetch --tags --force origin' command to be called. Commands: %v", mockRunner.GetCommands())

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

	t.Run("FetchWithTagsAndForceFlag", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock successful fetch and pull
		mockRunner.SetOutput("git fetch --tags --force origin", "")
		mockRunner.SetOutput("git pull --rebase origin", "Already up to date.")

		// Call pullLatestBranch directly
		err := pullLatestBranch()
		require.NoError(t, err)

		// Verify git fetch --tags --force was called
		expectedFetchCmd := []string{"git", "fetch", "--tags", "--force", "origin"}
		require.True(t, mockRunner.HasCommand(expectedFetchCmd),
			"Expected 'git fetch --tags --force origin' to be called. Commands: %v", mockRunner.GetCommands())
	})

	t.Run("FetchFailure", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Mock fetch failure (even with --force)
		mockRunner.SetError("git fetch --tags --force origin", errNetworkError)

		// Call pullLatestBranch directly
		err := pullLatestBranch()
		require.Error(t, err, "Should return error when fetch fails")
		require.Contains(t, err.Error(), "failed to fetch from origin", "Error should indicate fetch failure")
	})
}

// TestVersionBumpAutoIncrement tests auto-increment when calculated tag exists from different branch
func TestVersionBumpAutoIncrement(t *testing.T) {
	// Save original runner
	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore original runner: %v", err)
		}
	}()

	t.Run("AutoIncrementWhenTagExistsFromDifferentBranch", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Simulate scenario:
		// - Current reachable tag: v1.3.0
		// - Want to bump minor: v1.3.0 -> v1.4.0
		// - But v1.4.0 exists locally (from different branch)
		// - Should auto-increment to v1.5.0

		// Mock git commands for version bump
		mockRunner.SetOutput("git status --porcelain", "")                                                                                 // Clean working directory
		mockRunner.SetOutput("git tag --points-at HEAD", "")                                                                               // No tags on HEAD
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.0-22-gabcdef")                                                 // Current version v1.3.0
		mockRunner.SetOutput("git rev-list --count v1.3.0..HEAD", "22")                                                                    // Distance from tag
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)") // Mock git remote
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")                                           // Mock remote accessibility

		// Mock tag existence checks
		mockRunner.SetOutput("git tag -l v1.4.0", "v1.4.0")                       // v1.4.0 exists locally
		mockRunner.SetOutput("git rev-parse v1.4.0", "different123")              // Points to different commit
		mockRunner.SetOutput("git rev-parse HEAD", "abc123")                      // Current HEAD
		mockRunner.SetOutput("git tag -l v1.5.0", "")                             // v1.5.0 doesn't exist
		mockRunner.SetOutput("git ls-remote --tags origin v1.5.0", "")            // v1.5.0 doesn't exist on remote
		mockRunner.SetOutput("git log --oneline -5 --no-decorate", "abc123 Test") // Recent commits

		version := Version{}

		// Test version bump with push parameter
		err := version.Bump("push=true", "bump=minor")
		require.NoError(t, err, "Should succeed by auto-incrementing to next available version")

		// Verify v1.4.0 was checked and skipped
		expectedCheckCmd := []string{"git", "tag", "-l", "v1.4.0"}
		require.True(t, mockRunner.HasCommand(expectedCheckCmd),
			"Should check if v1.4.0 exists locally. Commands: %v", mockRunner.GetCommands())

		// Verify v1.5.0 was created (not v1.4.0)
		expectedTagCmd := []string{"git", "tag", "-a", "v1.5.0", "-m", "GitHubRelease v1.5.0"}
		require.True(t, mockRunner.HasCommand(expectedTagCmd),
			"Should create v1.5.0 after skipping v1.4.0. Commands: %v", mockRunner.GetCommands())

		// Verify v1.5.0 was pushed
		expectedPushCmd := []string{"git", "push", "origin", "v1.5.0"}
		require.True(t, mockRunner.HasCommand(expectedPushCmd),
			"Should push v1.5.0. Commands: %v", mockRunner.GetCommands())
	})

	t.Run("MultipleVersionsSkipped", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Simulate scenario:
		// - Current reachable tag: v1.3.0
		// - Want to bump minor: v1.3.0 -> v1.4.0
		// - But v1.4.0 and v1.5.0 exist locally (from different branches)
		// - Should auto-increment to v1.6.0

		// Mock git commands for version bump
		mockRunner.SetOutput("git status --porcelain", "")
		mockRunner.SetOutput("git tag --points-at HEAD", "")
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.0-22-gabcdef")
		mockRunner.SetOutput("git rev-list --count v1.3.0..HEAD", "22")
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)")
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")
		mockRunner.SetOutput("git log --oneline -5 --no-decorate", "abc123 Test")

		// Mock tag existence checks - v1.4.0 and v1.5.0 exist, v1.6.0 doesn't
		mockRunner.SetOutput("git tag -l v1.4.0", "v1.4.0")
		mockRunner.SetOutput("git rev-parse v1.4.0", "different123")
		mockRunner.SetOutput("git tag -l v1.5.0", "v1.5.0")
		mockRunner.SetOutput("git rev-parse v1.5.0", "different456")
		mockRunner.SetOutput("git tag -l v1.6.0", "")
		mockRunner.SetOutput("git rev-parse HEAD", "abc123")
		mockRunner.SetOutput("git ls-remote --tags origin v1.6.0", "")

		version := Version{}

		err := version.Bump("bump=minor")
		require.NoError(t, err)

		// Verify v1.6.0 was created (skipped v1.4.0 and v1.5.0)
		expectedTagCmd := []string{"git", "tag", "-a", "v1.6.0", "-m", "GitHubRelease v1.6.0"}
		require.True(t, mockRunner.HasCommand(expectedTagCmd),
			"Should create v1.6.0 after skipping v1.4.0 and v1.5.0. Commands: %v", mockRunner.GetCommands())
	})

	t.Run("TagAlreadyExistsOnHEAD", func(t *testing.T) {
		mockRunner := NewVersionBumpMockRunner()
		require.NoError(t, SetRunner(mockRunner))

		// Simulate scenario where calculated tag already exists on HEAD

		mockRunner.SetOutput("git status --porcelain", "")
		mockRunner.SetOutput("git tag --points-at HEAD", "")
		mockRunner.SetOutput("git describe --tags --long --abbrev=0", "v1.3.0-22-gabcdef")
		mockRunner.SetOutput("git rev-list --count v1.3.0..HEAD", "22")
		mockRunner.SetOutput("git remote -v", "origin\tgit@github.com:test/repo.git (fetch)\norigin\tgit@github.com:test/repo.git (push)")
		mockRunner.SetOutput("git ls-remote --exit-code origin HEAD", "abc123\trefs/heads/main")
		mockRunner.SetOutput("git log --oneline -5 --no-decorate", "abc123 Test")

		// Mock tag exists and points to HEAD
		mockRunner.SetOutput("git tag -l v1.4.0", "v1.4.0")
		mockRunner.SetOutput("git rev-parse v1.4.0", "abc123") // Same as HEAD
		mockRunner.SetOutput("git rev-parse HEAD", "abc123")

		version := Version{}

		err := version.Bump("bump=minor")
		require.NoError(t, err, "Should succeed when tag already exists on HEAD")

		// Verify tag creation was NOT attempted (tag already exists on HEAD)
		expectedTagCmd := []string{"git", "tag", "-a", "v1.4.0", "-m", "GitHubRelease v1.4.0"}
		require.False(t, mockRunner.HasCommand(expectedTagCmd),
			"Should not create tag when it already exists on HEAD. Commands: %v", mockRunner.GetCommands())
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
		// - Remote has v1.4.0 (created from another branch)
		// - After fetch --tags --force, v1.4.0 exists locally
		// - Auto-increment should skip v1.4.0 and create v1.5.0

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

		// Mock tag checks for auto-increment logic
		mockRunner.SetOutput("git tag -l v1.4.0", "v1.4.0")                                      // v1.4.0 exists locally (fetched from remote)
		mockRunner.SetOutput("git rev-parse v1.4.0", "98abdabb5b928ada967550c3218ea0faf7cc40b7") // Points to different commit
		mockRunner.SetOutput("git rev-parse HEAD", "78baa5e3d7bd355019db6caaa11286779aba6832")   // Current HEAD
		mockRunner.SetOutput("git tag -l v1.5.0", "")                                            // v1.5.0 doesn't exist
		mockRunner.SetOutput("git ls-remote --tags origin v1.5.0", "")                           // v1.5.0 doesn't exist on remote

		version := Version{}

		// Run the exact command from user's bug report
		err := version.Bump("push=true", "bump=minor", "branch=master")

		// Should succeed by auto-incrementing to v1.5.0
		require.NoError(t, err, "Should succeed by auto-incrementing to next available version")

		// Verify the correct sequence of commands
		commands := mockRunner.GetCommands()

		// Should have fetched with --tags --force
		expectedFetchCmd := []string{"git", "fetch", "--tags", "--force", "origin"}
		require.True(t, mockRunner.HasCommand(expectedFetchCmd),
			"Should fetch tags with --tags --force flag. Commands: %v", commands)

		// Should have checked if v1.4.0 exists locally
		expectedCheckCmd := []string{"git", "tag", "-l", "v1.4.0"}
		require.True(t, mockRunner.HasCommand(expectedCheckCmd),
			"Should check if v1.4.0 exists locally. Commands: %v", commands)

		// Should have created v1.5.0 (not v1.4.0)
		expectedTagCmd := []string{"git", "tag", "-a", "v1.5.0", "-m", "GitHubRelease v1.5.0"}
		require.True(t, mockRunner.HasCommand(expectedTagCmd),
			"Should create v1.5.0 after skipping v1.4.0. Commands: %v", commands)

		// Should have pushed v1.5.0
		expectedPushCmd := []string{"git", "push", "origin", "v1.5.0"}
		require.True(t, mockRunner.HasCommand(expectedPushCmd),
			"Should push v1.5.0. Commands: %v", commands)

		// Should NOT have tried to push v1.4.0
		unexpectedPushCmd := []string{"git", "push", "origin", "v1.4.0"}
		require.False(t, mockRunner.HasCommand(unexpectedPushCmd),
			"Should not push v1.4.0. Commands: %v", commands)
	})
}
