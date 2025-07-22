// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"fmt"
	"os"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/security"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Git namespace for git-related tasks
type Git mg.Namespace

// Diff shows git diff and fails if uncommitted changes exist
func (Git) Diff() error {
	utils.Header("Checking for Uncommitted Changes")

	// Check git diff
	output, err := GetRunner().RunCmdOutput("git", "diff", "--exit-code")
	if err != nil {
		utils.Error("Uncommitted changes detected in working directory")
		if output != "" {
			fmt.Println(output)
		}
		return fmt.Errorf("uncommitted changes in working directory")
	}

	// Check git status for untracked files
	output, err = GetRunner().RunCmdOutput("git", "status", "--porcelain")
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if strings.TrimSpace(output) != "" {
		utils.Error("Uncommitted changes detected:")
		fmt.Println(output)
		return fmt.Errorf("uncommitted changes found")
	}

	utils.Success("Working directory is clean")
	return nil
}

// Tag creates and pushes a new tag (use version=X.Y.Z)
func (Git) Tag() error {
	version := os.Getenv("version")
	if version == "" {
		return fmt.Errorf("version variable is required. Use: version=X.Y.Z mage git:tag")
	}

	// Validate version format
	if err := security.ValidateVersion(version); err != nil {
		return fmt.Errorf("invalid version format: %w", err)
	}

	utils.Header(fmt.Sprintf("Creating Tag v%s", version))

	tagName := fmt.Sprintf("v%s", version)

	// Create annotated tag
	if err := GetRunner().RunCmd("git", "tag", "-a", tagName, "-m", "Pending full release..."); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	// Push tag to origin
	if err := GetRunner().RunCmd("git", "push", "origin", tagName); err != nil {
		return fmt.Errorf("failed to push tag: %w", err)
	}

	// Fetch tags to ensure local is in sync
	if err := GetRunner().RunCmd("git", "fetch", "--tags", "-f"); err != nil {
		utils.Warn("Failed to fetch tags: %v", err)
	}

	utils.Success("Tag %s created and pushed successfully", tagName)
	return nil
}

// TagRemove removes local and remote tag (use version=X.Y.Z)
func (Git) TagRemove() error {
	version := os.Getenv("version")
	if version == "" {
		return fmt.Errorf("version variable is required. Use: version=X.Y.Z mage git:tagremove")
	}

	// Validate version format
	if err := security.ValidateVersion(version); err != nil {
		return fmt.Errorf("invalid version format: %w", err)
	}

	utils.Header(fmt.Sprintf("Removing Tag v%s", version))

	tagName := fmt.Sprintf("v%s", version)

	// Delete local tag
	if err := GetRunner().RunCmd("git", "tag", "-d", tagName); err != nil {
		utils.Warn("Failed to delete local tag (might not exist): %v", err)
	}

	// Delete remote tag
	if err := GetRunner().RunCmd("git", "push", "--delete", "origin", tagName); err != nil {
		utils.Warn("Failed to delete remote tag (might not exist): %v", err)
	}

	// Fetch tags to ensure local is in sync
	if err := GetRunner().RunCmd("git", "fetch", "--tags"); err != nil {
		utils.Warn("Failed to fetch tags: %v", err)
	}

	utils.Success("Tag %s removed from local and remote", tagName)
	return nil
}

// TagUpdate force-updates tag to current commit (use version=X.Y.Z)
func (Git) TagUpdate() error {
	version := os.Getenv("version")
	if version == "" {
		return fmt.Errorf("version variable is required. Use: version=X.Y.Z mage git:tagupdate")
	}

	// Validate version format
	if err := security.ValidateVersion(version); err != nil {
		return fmt.Errorf("invalid version format: %w", err)
	}

	utils.Header(fmt.Sprintf("Force Updating Tag v%s", version))

	tagName := fmt.Sprintf("v%s", version)

	// Force push current HEAD to tag
	refSpec := fmt.Sprintf("HEAD:refs/tags/%s", tagName)
	if err := GetRunner().RunCmd("git", "push", "--force", "origin", refSpec); err != nil {
		return fmt.Errorf("failed to force update tag: %w", err)
	}

	// Fetch tags to ensure local is in sync
	if err := GetRunner().RunCmd("git", "fetch", "--tags", "-f"); err != nil {
		utils.Warn("Failed to fetch tags: %v", err)
	}

	utils.Success("Tag %s force-updated to current commit", tagName)
	return nil
}

// Status shows the current git status
func (Git) Status() error {
	utils.Header("Git Status")

	return GetRunner().RunCmd("git", "status")
}

// Log shows recent commit history
func (Git) Log() error {
	utils.Header("Recent Commits")

	return GetRunner().RunCmd("git", "log", "--oneline", "-10")
}

// Branch shows current branch and all branches
func (Git) Branch() error {
	utils.Header("Git Branches")

	// Show current branch
	current, err := GetRunner().RunCmdOutput("git", "branch", "--show-current")
	if err == nil {
		utils.Info("Current branch: %s", strings.TrimSpace(current))
	}

	// List all branches
	return GetRunner().RunCmd("git", "branch", "-a")
}

// Pull pulls latest changes from origin
func (Git) Pull() error {
	utils.Header("Pulling Latest Changes")

	// Fetch first
	if err := GetRunner().RunCmd("git", "fetch", "--all", "--prune"); err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	// Pull with rebase
	if err := GetRunner().RunCmd("git", "pull", "--rebase"); err != nil {
		return fmt.Errorf("failed to pull: %w", err)
	}

	utils.Success("Successfully pulled latest changes")
	return nil
}

// Commit creates a commit with a message
func (Git) Commit(message ...string) error {
	var commitMessage string
	if len(message) > 0 {
		commitMessage = message[0]
	} else {
		commitMessage = os.Getenv("message")
		if commitMessage == "" {
			return fmt.Errorf("message parameter or environment variable is required")
		}
	}

	utils.Header("Creating Commit")

	// Add all changes
	if err := GetRunner().RunCmd("git", "add", "-A"); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Create commit
	if err := GetRunner().RunCmd("git", "commit", "-m", commitMessage); err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}

	utils.Success("Commit created successfully")
	return nil
}

// Additional methods for Git namespace required by tests

// Init initializes a git repository
func (Git) Init() error {
	runner := GetRunner()
	return runner.RunCmd("git", "init")
}

// Add adds files to staging area
func (Git) Add(files ...string) error {
	runner := GetRunner()
	if len(files) == 0 {
		return runner.RunCmd("git", "add", ".")
	}
	args := append([]string{"add"}, files...)
	return runner.RunCmd("git", args...)
}

// Clone clones a repository
func (Git) Clone() error {
	runner := GetRunner()
	return runner.RunCmd("git", "clone")
}

// Push pushes to remote with branch
func (Git) Push(remote, branch string) error {
	runner := GetRunner()
	return runner.RunCmd("git", "push", remote, branch)
}

// PullWithRemote pulls from specific remote and branch
func (Git) PullWithRemote(remote, branch string) error {
	runner := GetRunner()
	return runner.RunCmd("git", "pull", remote, branch)
}

// TagWithMessage creates a tag with message
func (Git) TagWithMessage(name, message string) error {
	runner := GetRunner()
	return runner.RunCmd("git", "tag", "-a", name, "-m", message)
}

// BranchWithName creates or switches to branch
func (Git) BranchWithName(name string) error {
	runner := GetRunner()
	return runner.RunCmd("git", "branch", name)
}

// CloneRepo clones a repository
func (Git) CloneRepo(url, directory string) error {
	runner := GetRunner()
	return runner.RunCmd("git", "clone", url, directory)
}

// LogWithCount shows log with limit
func (Git) LogWithCount(count int) error {
	runner := GetRunner()
	return runner.RunCmd("git", "log", fmt.Sprintf("-%d", count))
}
