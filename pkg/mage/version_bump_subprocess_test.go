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

	// Initialize git repo
	ctx := context.Background()
	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to init git repo: %v\nOutput: %s", err, output)
	}

	// Configure git user (required for commits)
	configCommands := [][]string{
		{"git", "config", "user.name", "Test User"},
		{"git", "config", "user.email", "test@example.com"},
		{"git", "config", "init.defaultBranch", "main"},
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
