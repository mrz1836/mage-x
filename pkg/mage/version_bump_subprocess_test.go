package mage

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestVersionBumpMagexSubprocess tests the actual magex binary with subprocess execution
func TestVersionBumpMagexSubprocess(t *testing.T) {
	// Skip this test in short mode since it builds a binary
	if testing.Short() {
		t.Skip("Skipping subprocess test in short mode")
	}

	// Build the magex binary for testing
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	tempDir := t.TempDir()
	magexBinary := filepath.Join(tempDir, "magex")

	// Build the magex binary
	//nolint:gosec // G204: Subprocess launched with variable - safe in test context
	buildCmd := exec.CommandContext(ctx, "go", "build", "-o", magexBinary, "./cmd/magex")
	buildCmd.Dir = filepath.Join("..", "..") // Go up to the project root
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build magex binary: %v\nOutput: %s", err, buildOutput)
	}

	t.Run("MagexVersionBumpWithPushTrue", func(t *testing.T) {
		// Create a temporary git repository for testing
		testRepoDir := t.TempDir()

		// Initialize git repository
		initGit(t, testRepoDir)

		// Create initial commit and tag
		createInitialCommitAndTag(t, testRepoDir, "v1.0.0")

		// Create another commit so we have something to bump from
		createTestCommit(t, testRepoDir, "test change")

		// Change to the test directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			if chdirErr := os.Chdir(originalDir); chdirErr != nil {
				t.Logf("Failed to restore original directory: %v", chdirErr)
			}
		}()

		err = os.Chdir(testRepoDir)
		require.NoError(t, err)

		// Run magex version:bump with push=true parameter
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		//nolint:gosec // G204: Subprocess launched with variable - safe in test context
		magexCmd := exec.CommandContext(ctx, magexBinary, "version:bump", "push=true", "bump=patch")
		magexCmd.Dir = testRepoDir
		output, err := magexCmd.CombinedOutput()

		t.Logf("magex command: %s", strings.Join(magexCmd.Args, " "))
		t.Logf("magex output: %s", output)

		if err != nil {
			t.Logf("magex failed with error: %v", err)
		}

		// Check if the command mentioned pushing
		outputStr := string(output)

		// Look for indicators that push was attempted
		hasPushMessage := strings.Contains(outputStr, "Tag pushed to remote") ||
			strings.Contains(outputStr, "Pushing tag to remote")

		// Look for indicators that push was NOT attempted
		hasNoPushMessage := strings.Contains(outputStr, "To push the tag, run:") ||
			strings.Contains(outputStr, "Or add 'push' parameter")

		t.Logf("Has push message: %t", hasPushMessage)
		t.Logf("Has no-push message: %t", hasNoPushMessage)

		if hasNoPushMessage && !hasPushMessage {
			t.Errorf("magex version:bump with push=true parameter did not attempt to push the tag")
			t.Errorf("Expected to see push attempt, but got no-push message instead")
		}

		// Also check git tags to see if tag was created
		checkCmd := exec.CommandContext(ctx, "git", "tag", "--list")
		checkCmd.Dir = testRepoDir
		tagOutput, tagErr := checkCmd.CombinedOutput()
		if tagErr == nil {
			t.Logf("Git tags after version:bump: %s", tagOutput)
		}
	})

	t.Run("MagexVersionBumpWithPushFlag", func(t *testing.T) {
		// Create a temporary git repository for testing
		testRepoDir := t.TempDir()

		// Initialize git repository
		initGit(t, testRepoDir)

		// Create initial commit and tag
		createInitialCommitAndTag(t, testRepoDir, "v1.0.0")

		// Create another commit so we have something to bump from
		createTestCommit(t, testRepoDir, "test change")

		// Change to the test directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			if chdirErr := os.Chdir(originalDir); chdirErr != nil {
				t.Logf("Failed to restore original directory: %v", chdirErr)
			}
		}()

		err = os.Chdir(testRepoDir)
		require.NoError(t, err)

		// Run magex version:bump with push as boolean flag
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		//nolint:gosec // G204: Subprocess launched with variable - safe in test context
		magexCmd := exec.CommandContext(ctx, magexBinary, "version:bump", "push", "bump=patch")
		magexCmd.Dir = testRepoDir
		output, err := magexCmd.CombinedOutput()

		t.Logf("magex command: %s", strings.Join(magexCmd.Args, " "))
		t.Logf("magex output: %s", output)

		if err != nil {
			t.Logf("magex failed with error: %v", err)
		}

		// Check if the command mentioned pushing
		outputStr := string(output)

		// Look for indicators that push was attempted
		hasPushMessage := strings.Contains(outputStr, "Tag pushed to remote") ||
			strings.Contains(outputStr, "Pushing tag to remote")

		// Look for indicators that push was NOT attempted
		hasNoPushMessage := strings.Contains(outputStr, "To push the tag, run:") ||
			strings.Contains(outputStr, "Or add 'push' parameter")

		t.Logf("Has push message: %t", hasPushMessage)
		t.Logf("Has no-push message: %t", hasNoPushMessage)

		if hasNoPushMessage && !hasPushMessage {
			t.Errorf("magex version:bump with push flag did not attempt to push the tag")
			t.Errorf("Expected to see push attempt, but got no-push message instead")
		}
	})

	t.Run("MagexVersionBumpWithoutPush", func(t *testing.T) {
		// Create a temporary git repository for testing
		testRepoDir := t.TempDir()

		// Initialize git repository
		initGit(t, testRepoDir)

		// Create initial commit and tag
		createInitialCommitAndTag(t, testRepoDir, "v1.0.0")

		// Create another commit so we have something to bump from
		createTestCommit(t, testRepoDir, "test change")

		// Change to the test directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			if chdirErr := os.Chdir(originalDir); chdirErr != nil {
				t.Logf("Failed to restore original directory: %v", chdirErr)
			}
		}()

		err = os.Chdir(testRepoDir)
		require.NoError(t, err)

		// Run magex version:bump without push parameter
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		//nolint:gosec // G204: Subprocess launched with variable - safe in test context
		magexCmd := exec.CommandContext(ctx, magexBinary, "version:bump", "bump=patch")
		magexCmd.Dir = testRepoDir
		output, err := magexCmd.CombinedOutput()

		t.Logf("magex command: %s", strings.Join(magexCmd.Args, " "))
		t.Logf("magex output: %s", output)

		if err != nil {
			t.Logf("magex failed with error: %v", err)
		}

		// Check if the command mentioned NOT pushing (which is expected)
		outputStr := string(output)

		// Look for indicators that push was NOT attempted (expected behavior)
		hasNoPushMessage := strings.Contains(outputStr, "To push the tag, run:") ||
			strings.Contains(outputStr, "Or add 'push' parameter")

		t.Logf("Has no-push message: %t", hasNoPushMessage)

		if !hasNoPushMessage {
			t.Errorf("magex version:bump without push parameter should have shown no-push message")
		}
	})
}

// Helper functions for setting up test git repositories

func initGit(t *testing.T, dir string) {
	t.Helper()

	// Initialize git repo with explicit branch name for consistency
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "git", "init", "--initial-branch=master")
	cmd.Dir = dir
	_, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback for older git versions that don't support --initial-branch
		cmd = exec.CommandContext(ctx, "git", "init")
		cmd.Dir = dir
		output, fallbackErr := cmd.CombinedOutput()
		if fallbackErr != nil {
			t.Fatalf("Failed to init git repo: %v\nOutput: %s", fallbackErr, output)
		}
	}

	// Configure git user (required for commits)
	configCommands := [][]string{
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
	}

	for _, configCmd := range configCommands {
		//nolint:gosec // G204: Subprocess launched with a potential tainted input - safe in test context
		cmd := exec.CommandContext(ctx, configCmd[0], configCmd[1:]...)
		cmd.Dir = dir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to configure git: %v\nCommand: %v\nOutput: %s", err, configCmd, output)
		}
	}
}

func createInitialCommitAndTag(t *testing.T, dir, tag string) {
	t.Helper()

	// Create a test file
	testFile := filepath.Join(dir, "test.txt")
	err := os.WriteFile(testFile, []byte("initial content"), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add and commit
	ctx := context.Background()
	commands := [][]string{
		{"git", "add", "test.txt"},
		{"git", "commit", "-m", "Initial commit"},
		{"git", "tag", "-a", tag, "-m", fmt.Sprintf("Release %s", tag)},
	}

	for _, cmdArgs := range commands {
		//nolint:gosec // G204: Subprocess launched with a potential tainted input - safe in test context
		cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = dir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to execute git command %v: %v\nOutput: %s", cmdArgs, err, output)
		}
	}
}

func createTestCommit(t *testing.T, dir, message string) {
	t.Helper()

	// Modify the test file
	testFile := filepath.Join(dir, "test.txt")
	err := os.WriteFile(testFile, []byte("modified content - "+message), 0o600)
	if err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Add and commit
	ctx := context.Background()
	commands := [][]string{
		{"git", "add", "test.txt"},
		{"git", "commit", "-m", message},
	}

	for _, cmdArgs := range commands {
		//nolint:gosec // G204: Subprocess launched with a potential tainted input - safe in test context
		cmd := exec.CommandContext(ctx, cmdArgs[0], cmdArgs[1:]...)
		cmd.Dir = dir
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to execute git command %v: %v\nOutput: %s", cmdArgs, err, output)
		}
	}
}

// TestVersionBumpWithSquashMergedTag tests that version bump correctly detects tags on squash-merged branches
func TestVersionBumpWithSquashMergedTag(t *testing.T) {
	// Skip this test in short mode since it creates a real git repo
	if testing.Short() {
		t.Skip("Skipping squash-merge test in short mode")
	}

	// Save and restore the original runner to avoid test pollution
	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore original runner: %v", err)
		}
	}()
	// Reset to default runner for real git operations
	require.NoError(t, SetRunner(NewSecureCommandRunner()))

	// Create a temporary git repository for testing
	testRepoDir := t.TempDir()

	// Initialize git repository
	initGit(t, testRepoDir)

	// Create initial commit and v0.0.1 tag on main
	createInitialCommitAndTag(t, testRepoDir, "v0.0.1")

	ctx := context.Background()

	// Create a feature branch

	checkoutCmd := exec.CommandContext(ctx, "git", "checkout", "-b", "feature")
	checkoutCmd.Dir = testRepoDir
	if output, err := checkoutCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create feature branch: %v\nOutput: %s", err, output)
	}

	// Create a commit on the feature branch and tag it with v0.1.0
	createTestCommit(t, testRepoDir, "feature work")

	tagCmd := exec.CommandContext(ctx, "git", "tag", "-a", "v0.1.0", "-m", "Release v0.1.0")
	tagCmd.Dir = testRepoDir
	if output, err := tagCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create tag: %v\nOutput: %s", err, output)
	}

	// Go back to master branch (initGit sets up master as default)

	checkoutMainCmd := exec.CommandContext(ctx, "git", "checkout", "master")
	checkoutMainCmd.Dir = testRepoDir
	if output, err := checkoutMainCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to checkout master: %v\nOutput: %s", err, output)
	}

	// Squash merge the feature branch (this creates a new commit that is NOT an ancestor of v0.1.0)

	squashCmd := exec.CommandContext(ctx, "git", "merge", "--squash", "feature")
	squashCmd.Dir = testRepoDir
	if output, err := squashCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to squash merge: %v\nOutput: %s", err, output)
	}

	commitCmd := exec.CommandContext(ctx, "git", "commit", "-m", "Squash merge feature branch")
	commitCmd.Dir = testRepoDir
	if output, err := commitCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to commit squash merge: %v\nOutput: %s", err, output)
	}

	// At this point:
	// - v0.1.0 exists in the repo (on the feature branch commit)
	// - v0.0.1 is reachable from main (git describe can find it)
	// - v0.1.0 is NOT reachable from main (git describe cannot find it)
	// - The bug would cause version:bump to create v0.0.2 instead of v0.1.1

	// Verify git describe cannot find v0.1.0

	describeCmd := exec.CommandContext(ctx, "git", "describe", "--tags", "--abbrev=0")
	describeCmd.Dir = testRepoDir
	describeOutput, describeErr := describeCmd.CombinedOutput()
	t.Logf("git describe output: %s (err: %v)", strings.TrimSpace(string(describeOutput)), describeErr)

	// Verify git tag --sort can find v0.1.0

	tagListCmd := exec.CommandContext(ctx, "git", "tag", "--sort=-version:refname")
	tagListCmd.Dir = testRepoDir
	tagListOutput, tagListErr := tagListCmd.CombinedOutput()
	t.Logf("git tag list output: %s (err: %v)", strings.TrimSpace(string(tagListOutput)), tagListErr)

	// Test our getHighestVersionTag function
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore original directory: %v", chdirErr)
		}
	}()
	require.NoError(t, os.Chdir(testRepoDir))

	// The fix should find v0.1.0 as the highest tag
	highestTag := getHighestVersionTag()
	t.Logf("getHighestVersionTag returned: %s", highestTag)
	require.Equal(t, "v0.1.0", highestTag, "Should find v0.1.0 even though it's not reachable from HEAD")

	// getCurrentGitTag should also return v0.1.0
	currentTag := getCurrentGitTag()
	t.Logf("getCurrentGitTag returned: %s", currentTag)
	require.Equal(t, "v0.1.0", currentTag, "getCurrentGitTag should return v0.1.0 to prevent version regression")
}

// TestVersionBumpWithPrereleaseTag tests that version bump correctly handles pre-release tags
func TestVersionBumpWithPrereleaseTag(t *testing.T) {
	// Skip this test in short mode since it creates a real git repo
	if testing.Short() {
		t.Skip("Skipping prerelease test in short mode")
	}

	// Save and restore the original runner to avoid test pollution
	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore original runner: %v", err)
		}
	}()
	// Reset to default runner for real git operations
	require.NoError(t, SetRunner(NewSecureCommandRunner()))

	// Create a temporary git repository for testing
	testRepoDir := t.TempDir()

	// Initialize git repository
	initGit(t, testRepoDir)

	// Create initial commit and v1.0.0 tag
	createInitialCommitAndTag(t, testRepoDir, "v1.0.0")

	ctx := context.Background()

	// Create another commit and tag it with v1.1.0-beta
	createTestCommit(t, testRepoDir, "beta release")

	tagCmd := exec.CommandContext(ctx, "git", "tag", "-a", "v1.1.0-beta", "-m", "Release v1.1.0-beta")
	tagCmd.Dir = testRepoDir
	if output, err := tagCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to create tag: %v\nOutput: %s", err, output)
	}

	// Create another commit (so we're not on the tagged commit)
	createTestCommit(t, testRepoDir, "post beta work")

	// Test version detection
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore original directory: %v", chdirErr)
		}
	}()
	require.NoError(t, os.Chdir(testRepoDir))

	// The highest tag should be v1.1.0-beta (according to git's version sort)
	highestTag := getHighestVersionTag()
	t.Logf("getHighestVersionTag returned: %s", highestTag)

	// Parse both versions and compare
	highestV, err := ParseSemanticVersion(highestTag)
	require.NoError(t, err)

	stableV, err := ParseSemanticVersion("v1.0.0")
	require.NoError(t, err)

	// According to semver, 1.0.0 > 1.0.0-beta, but 1.1.0-beta > 1.0.0
	// So v1.1.0-beta should be higher than v1.0.0
	require.True(t, highestV.Compare(stableV) > 0 || highestTag == "v1.1.0-beta",
		"v1.1.0-beta should be higher than v1.0.0 or be detected as highest tag")
}
