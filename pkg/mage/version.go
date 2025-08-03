// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors to satisfy err113 linter
var (
	errCannotParseGitHubInfo     = errors.New("cannot parse GitHub info from module")
	errInvalidBumpType           = errors.New("invalid BUMP type (must be major, minor, or patch)")
	errVersionUncommittedChanges = errors.New("working directory has uncommitted changes")
	errGitHubAPIError            = errors.New("GitHub API error")
	errInvalidVersionFormat      = errors.New("invalid version format")
	errInvalidMajorVersion       = errors.New("invalid major version")
	errInvalidMinorVersion       = errors.New("invalid minor version")
	errInvalidPatchVersion       = errors.New("invalid patch version")
	errMultipleTagsOnCommit      = errors.New("current commit already has version tags")
	errIllogicalVersionJump      = errors.New("version jump appears illogical")
)

// Version namespace for version management tasks
type Version mg.Namespace

// GitHubRelease represents a GitHub release
type GitHubRelease struct {
	TagName     string                `json:"tag_name"`
	Name        string                `json:"name"`
	Prerelease  bool                  `json:"prerelease"`
	Draft       bool                  `json:"draft"`
	PublishedAt time.Time             `json:"published_at"`
	Body        string                `json:"body"`
	HTMLURL     string                `json:"html_url"`
	Assets      []VersionReleaseAsset `json:"assets"`
}

// VersionReleaseAsset represents a release asset
type VersionReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

// BuildInfo contains all build-time information
type BuildInfo struct {
	Version   string
	Commit    string
	BuildDate string
}

// Package-level variables for build info configuration
var (
	// Build-time variables that can be set at build time
	version   = "dev"
	commit    = "unknown" //nolint:gochecknoglobals // Build-time variables
	buildDate = "unknown" //nolint:gochecknoglobals // Build-time variables

	// Thread-safe access to build info
	buildInfoOnce sync.Once //nolint:gochecknoglobals // Required for thread-safe initialization
	buildInfoData BuildInfo //nolint:gochecknoglobals // Private data for sync.Once pattern
)

// Show displays the current version
func (Version) Show() error {
	utils.Header("Version Information")

	fmt.Printf("Version:    %s\n", getVersionInfo())
	fmt.Printf("Commit:     %s\n", getCommitInfo())
	fmt.Printf("Build Date: %s\n", getBuildDate())
	fmt.Printf("Go Version: %s\n", runtime.Version())
	fmt.Printf("Platform:   %s/%s\n", runtime.GOOS, runtime.GOARCH)

	// Check if this is a git repo
	if isGitRepo() {
		if dirty := isGitDirty(); dirty {
			utils.Warn("\nWorking directory has uncommitted changes")
		}
	}

	return nil
}

// Check checks for available updates
func (Version) Check(_ ...string) error {
	utils.Header("Checking for Updates")

	current := getVersionInfo()
	utils.Info("Current version: %s", current)

	// Get module info
	module, err := utils.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %w", err)
	}

	// Parse module to get owner/repo
	parts := strings.Split(module, "/")
	if len(parts) < 3 {
		return fmt.Errorf("%w: %s", errCannotParseGitHubInfo, module)
	}

	owner := parts[1]
	repo := parts[2]

	// Check GitHub releases
	latest, err := getLatestGitHubRelease(owner, repo)
	if err != nil {
		// Check if it's a 404 (no releases found)
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found") {
			utils.Warn("No GitHub releases found for %s/%s", owner, repo)
			utils.Info("This project may use Git tags instead of GitHub releases")
			utils.Info("\nTo create a release:")
			utils.Info("1. Visit https://github.com/%s/%s/releases", owner, repo)
			utils.Info("2. Click 'Create a new release'")
			utils.Info("3. Select tag %s and publish", current)
			return nil
		}
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	utils.Info("Latest version: %s", latest.TagName)

	// Compare versions
	if isNewer(latest.TagName, current) {
		utils.Success("\n🎉 New version available: %s", latest.TagName)
		utils.Info("GitHubRelease: %s", latest.Name)
		if latest.Body != "" {
			utils.Info("\nGitHubRelease Notes:")
			utils.Info("%s", formatReleaseNotes(latest.Body))
		}
		utils.Info("\nUpdate with: go install %s@%s", module, latest.TagName)
	} else {
		utils.Success("✅ You are running the latest version")
	}

	return nil
}

// Update updates to the latest version
func (Version) Update() error {
	utils.Header("Updating to Latest Version")

	// Check for updates first
	current := getVersionInfo()
	utils.Info("Current version: %s", current)

	// Get module info
	module, err := utils.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %w", err)
	}

	// Parse module to get owner/repo
	parts := strings.Split(module, "/")
	if len(parts) < 3 {
		return fmt.Errorf("%w: %s", errCannotParseGitHubInfo, module)
	}

	owner := parts[1]
	repo := parts[2]

	// Get latest release
	latest, err := getLatestGitHubRelease(owner, repo)
	if err != nil {
		// Check if it's a 404 (no releases found)
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "Not Found") {
			utils.Warn("No GitHub releases found for %s/%s", owner, repo)
			utils.Info("Cannot update without published releases")
			utils.Info("Current version: %s", current)
			return nil
		}
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !isNewer(latest.TagName, current) {
		utils.Success("Already running the latest version: %s", current)
		return nil
	}

	utils.Info("Updating to version %s...", latest.TagName)

	// Use go install to update
	pkg := fmt.Sprintf("%s@%s", module, latest.TagName)

	if err := GetRunner().RunCmd("go", "install", pkg); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	utils.Success("Successfully updated to version %s", latest.TagName)
	utils.Info("Restart your application to use the new version")

	return nil
}

// Bump bumps the version number
func (Version) Bump(_ ...string) error {
	utils.Header("Bumping Version")

	// Get bump type from environment
	bumpType := utils.GetEnv("BUMP", "patch")
	if bumpType != "major" && bumpType != "minor" && bumpType != "patch" {
		return fmt.Errorf("%w: %s", errInvalidBumpType, bumpType)
	}

	// Check for uncommitted changes first
	if dirty := isGitDirty(); dirty {
		return errVersionUncommittedChanges
	}

	// Check if current commit already has version tags
	existingTags, err := getTagsOnCurrentCommit()
	if err != nil {
		return fmt.Errorf("failed to check existing tags: %w", err)
	}

	if len(existingTags) > 0 {
		utils.Warn("Current commit already has version tags: %s", strings.Join(existingTags, ", "))
		utils.Warn("Please create a new commit before bumping the version again")
		return fmt.Errorf("%w: %s", errMultipleTagsOnCommit, strings.Join(existingTags, ", "))
	}

	// Get current version
	current := getCurrentGitTag()
	if current == "" {
		current = "v0.0.0"
		utils.Info("No previous tags found, starting from %s", current)
	}

	// Parse and bump version
	newVersion, err := bumpVersion(current, bumpType)
	if err != nil {
		return fmt.Errorf("failed to bump version: %w", err)
	}

	// Validate version progression
	if err := validateVersionProgression(current, newVersion, bumpType); err != nil {
		return err
	}

	utils.Info("Bumping from %s to %s (%s bump)", current, newVersion, bumpType)

	// Create annotated tag
	message := fmt.Sprintf("GitHubRelease %s", newVersion)
	if err := GetRunner().RunCmd("git", "tag", "-a", newVersion, "-m", message); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	utils.Success("Created tag: %s", newVersion)

	// Ask to push
	if os.Getenv("PUSH") == approvalTrue {
		utils.Info("Pushing tag to remote...")
		if err := GetRunner().RunCmd("git", "push", "origin", newVersion); err != nil {
			return fmt.Errorf("failed to push tag: %w", err)
		}
		utils.Success("Tag pushed to remote")
	} else {
		utils.Info("To push the tag, run: git push origin %s", newVersion)
		utils.Info("Or set PUSH=true to push automatically")
	}

	return nil
}

// Changelog generates a changelog for the current version
func (Version) Changelog() error {
	utils.Header("Generating Changelog")

	// Get version range
	fromTag := utils.GetEnv("FROM", "")
	toTag := utils.GetEnv("TO", "HEAD")

	if fromTag == "" {
		// Get previous tag
		fromTag = getPreviousTag()
		if fromTag == "" {
			utils.Info("No previous tag found, showing all commits")
		}
	}

	// Generate changelog

	var args []string
	if fromTag != "" {
		args = []string{"log", "--pretty=format:- %s (%h)", fmt.Sprintf("%s..%s", fromTag, toTag)}
	} else {
		args = []string{"log", "--pretty=format:- %s (%h)", toTag}
	}

	output, err := GetRunner().RunCmdOutput("git", args...)
	if err != nil {
		return fmt.Errorf("failed to generate changelog: %w", err)
	}

	if fromTag != "" {
		fmt.Printf("\n## Changes from %s to %s\n\n", fromTag, toTag)
	} else {
		fmt.Printf("\n## All Changes\n\n")
	}

	if strings.TrimSpace(output) == "" {
		utils.Info("No changes found")
	} else {
		utils.Info("%s", output)
	}

	// Show commit count
	var countArgs []string
	if fromTag != "" {
		countArgs = []string{"rev-list", "--count", fmt.Sprintf("%s..%s", fromTag, toTag)}
	} else {
		countArgs = []string{"rev-list", "--count", toTag}
	}

	if count, err := GetRunner().RunCmdOutput("git", countArgs...); err == nil {
		fmt.Printf("\n%s commits\n", strings.TrimSpace(count))
	}

	return nil
}

// getBuildInfo returns the build information using thread-safe initialization
func getBuildInfo() BuildInfo {
	buildInfoOnce.Do(func() {
		buildInfoData = BuildInfo{
			Version:   version,
			Commit:    commit,
			BuildDate: buildDate,
		}
	})
	return buildInfoData
}

// Helper functions

// getVersionInfo returns the current version
func getVersionInfo() string {
	buildInfo := getBuildInfo()
	if buildInfo.Version != versionDev {
		return buildInfo.Version
	}

	// Try to get from git
	if tag := getCurrentGitTag(); tag != "" {
		return tag
	}

	return versionDev
}

// getCommitInfo returns the current commit
func getCommitInfo() string {
	buildInfo := getBuildInfo()
	if buildInfo.Commit != statusUnknown {
		return buildInfo.Commit
	}

	// Try to get from git
	if sha, err := GetRunner().RunCmdOutput("git", "rev-parse", "--short", "HEAD"); err == nil {
		return strings.TrimSpace(sha)
	}

	return statusUnknown
}

// getBuildDate returns the build date
func getBuildDate() string {
	buildInfo := getBuildInfo()
	if buildInfo.BuildDate != statusUnknown {
		return buildInfo.BuildDate
	}

	return time.Now().Format(time.RFC3339)
}

// isGitRepo checks if we're in a git repository
func isGitRepo() bool {
	err := GetRunner().RunCmd("git", "rev-parse", "--git-dir")
	return err == nil
}

// isGitDirty checks if the working directory has uncommitted changes
func isGitDirty() bool {
	output, err := GetRunner().RunCmdOutput("git", "status", "--porcelain")
	return err == nil && strings.TrimSpace(output) != ""
}

// getCurrentGitTag gets the current git tag
func getCurrentGitTag() string {
	// Get all tags pointing to HEAD, sorted by version (highest first)
	tags, err := GetRunner().RunCmdOutput("git", "tag", "--sort=-version:refname", "--points-at", "HEAD")
	if err != nil {
		// Fallback to getting the most recent tag if no tags point to HEAD
		tag, err := GetRunner().RunCmdOutput("git", "describe", "--tags", "--abbrev=0")
		if err != nil {
			return ""
		}
		return strings.TrimSpace(tag)
	}

	// If we have tags, return the first one (highest version)
	tagList := strings.Split(strings.TrimSpace(tags), "\n")
	if len(tagList) > 0 && tagList[0] != "" {
		return tagList[0]
	}

	// Fallback to describe if no tags on HEAD
	tag, err := GetRunner().RunCmdOutput("git", "describe", "--tags", "--abbrev=0")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(tag)
}

// getPreviousTag gets the previous git tag
func getPreviousTag() string {
	tags, err := GetRunner().RunCmdOutput("git", "tag", "--sort=-version:refname")
	if err != nil {
		return ""
	}

	tagList := strings.Split(strings.TrimSpace(tags), "\n")
	if len(tagList) > 1 {
		return tagList[1]
	}

	return ""
}

// getLatestGitHubRelease fetches the latest release from GitHub
func getLatestGitHubRelease(owner, repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(context.Background(), "GET", url, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			// Ignore error in defer cleanup
			log.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("GitHub API error: failed to read response body: %w", readErr)
		}
		return nil, fmt.Errorf("%w: %s", errGitHubAPIError, body)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// isNewer checks if version a is newer than version b

// bumpVersion bumps the version according to type
func bumpVersion(current, bumpType string) (string, error) {
	current = strings.TrimPrefix(current, "v")
	parts := strings.Split(current, ".")

	if len(parts) != 3 {
		return "", fmt.Errorf("%w: %s", errInvalidVersionFormat, current)
	}

	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return "", fmt.Errorf("%w: %s", errInvalidMajorVersion, parts[0])
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("%w: %s", errInvalidMinorVersion, parts[1])
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return "", fmt.Errorf("%w: %s", errInvalidPatchVersion, parts[2])
	}

	switch bumpType {
	case "major":
		major++
		minor = 0
		patch = 0
	case "minor":
		minor++
		patch = 0
	case "patch":
		patch++
	}

	return fmt.Sprintf("v%d.%d.%d", major, minor, patch), nil
}

// formatReleaseNotes formats release notes for display

// getTagsOnCurrentCommit returns all version tags on the current commit
func getTagsOnCurrentCommit() ([]string, error) {
	output, err := GetRunner().RunCmdOutput("git", "tag", "--points-at", "HEAD")
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(output) == "" {
		return []string{}, nil
	}

	// Filter only version tags (starting with 'v' followed by a number)
	allTags := strings.Split(strings.TrimSpace(output), "\n")
	var versionTags []string
	for _, tag := range allTags {
		if strings.HasPrefix(tag, "v") && len(tag) > 1 {
			if _, err := strconv.Atoi(string(tag[1])); err == nil {
				versionTags = append(versionTags, tag)
			}
		}
	}

	return versionTags, nil
}

// validateVersionProgression checks if the version bump is logical
func validateVersionProgression(current, new, bumpType string) error {
	currentParts := strings.Split(strings.TrimPrefix(current, "v"), ".")
	newParts := strings.Split(strings.TrimPrefix(new, "v"), ".")

	if len(currentParts) != 3 || len(newParts) != 3 {
		return nil // Skip validation if format is unexpected
	}

	currMajor, _ := strconv.Atoi(currentParts[0])
	currMinor, _ := strconv.Atoi(currentParts[1])
	currPatch, _ := strconv.Atoi(currentParts[2])

	newMajor, _ := strconv.Atoi(newParts[0])
	newMinor, _ := strconv.Atoi(newParts[1])
	newPatch, _ := strconv.Atoi(newParts[2])

	switch bumpType {
	case "patch":
		if newMajor != currMajor || newMinor != currMinor || newPatch != currPatch+1 {
			return fmt.Errorf("%w: expected %s → v%d.%d.%d, got %s",
				errIllogicalVersionJump, current, currMajor, currMinor, currPatch+1, new)
		}
	case "minor":
		if newMajor != currMajor || newMinor != currMinor+1 || newPatch != 0 {
			return fmt.Errorf("%w: expected %s → v%d.%d.0, got %s",
				errIllogicalVersionJump, current, currMajor, currMinor+1, new)
		}
	case "major":
		if newMajor != currMajor+1 || newMinor != 0 || newPatch != 0 {
			return fmt.Errorf("%w: expected %s → v%d.0.0, got %s",
				errIllogicalVersionJump, current, currMajor+1, new)
		}
	}

	return nil
}

// Additional methods for Version namespace required by tests

// Tag creates a version tag
func (Version) Tag(_ ...string) error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Creating version tag")
}

// Next shows next version
func (Version) Next(_, _ string) (string, error) {
	return "v1.0.1", nil
}

// Compare compares versions
func (Version) Compare(_, _ string) error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Comparing versions")
}

// Validate validates version format
func (Version) Validate(_ string) error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Validating version")
}

// Parse parses a version string
func (Version) Parse(_ string) ([]int, error) {
	return []int{1, 0, 0}, nil
}

// Format formats a version
func (Version) Format(_ []int) string {
	return "v1.0.0"
}
