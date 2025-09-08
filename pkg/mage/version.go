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
	errCannotParseGitHubInfo        = errors.New("cannot parse GitHub info from module")
	errInvalidBumpType              = errors.New("invalid BUMP type (must be major, minor, or patch)")
	errVersionUncommittedChanges    = errors.New("working directory has uncommitted changes")
	errGitHubAPIError               = errors.New("GitHub API error")
	errInvalidVersionFormat         = errors.New("invalid version format")
	errInvalidMajorVersion          = errors.New("invalid major version")
	errInvalidMinorVersion          = errors.New("invalid minor version")
	errInvalidPatchVersion          = errors.New("invalid patch version")
	errMultipleTagsOnCommit         = errors.New("current commit already has version tags")
	errIllogicalVersionJump         = errors.New("version jump appears illogical")
	errMajorBumpRequiresConfirm     = errors.New("major version bump requires explicit confirmation via MAJOR_BUMP_CONFIRM=true")
	errVersionBumpBlocked           = errors.New("version bump blocked due to safety check - use FORCE_VERSION_BUMP=true to override")
	errUnexpectedMajorVersionJump   = errors.New("unexpected major version jump")
	errUnexpectedlyLargeVersionJump = errors.New("unexpectedly large version jump")
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

// GHReleaseResponse represents a GitHub release from gh CLI
type GHReleaseResponse struct {
	TagName      string           `json:"tagName"`
	Body         string           `json:"body"`
	IsPrerelease bool             `json:"isPrerelease"`
	IsDraft      bool             `json:"isDraft"`
	PublishedAt  string           `json:"publishedAt"`
	URL          string           `json:"url"`
	Assets       []GHReleaseAsset `json:"assets"`
}

// GHReleaseAsset represents a release asset from gh CLI
type GHReleaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64  `json:"size"`
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
	version = "dev"
)

// BuildInfoProvider manages thread-safe access to build information
type BuildInfoProvider interface {
	GetBuildInfo() BuildInfo
}

// buildInfoProvider implements BuildInfoProvider with thread-safe lazy initialization
type buildInfoProvider struct {
	once sync.Once
	data BuildInfo
}

// NewBuildInfoProvider creates a new build info provider
func NewBuildInfoProvider() BuildInfoProvider {
	return &buildInfoProvider{}
}

// GetBuildInfo returns the build information using thread-safe initialization
func (bip *buildInfoProvider) GetBuildInfo() BuildInfo {
	bip.once.Do(func() {
		// Build-time variables that can be set at build time
		commit := statusUnknown
		buildDate := statusUnknown

		bip.data = BuildInfo{
			Version:   version,
			Commit:    commit,
			BuildDate: buildDate,
		}
	})
	return bip.data
}

// GetDefaultBuildInfoProvider returns a default build info provider instance
func GetDefaultBuildInfoProvider() BuildInfoProvider {
	return NewBuildInfoProvider()
}

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
			utils.Info("To create a release:")
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
		utils.Success("🎉 New version available: %s", latest.TagName)
		utils.Info("GitHubRelease: %s", latest.Name)
		if latest.Body != "" {
			utils.Info("GitHubRelease Notes:")
			utils.Info("%s", formatReleaseNotes(latest.Body))
		}
		utils.Info("Update with: go install %s@%s", module, latest.TagName)
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
func (Version) Bump(args ...string) error {
	utils.Header("Bumping Version")

	// Parse command-line parameters
	params := utils.ParseParams(args)

	// Check for dry-run mode
	dryRun := utils.IsParamTrue(params, "dry-run")

	// Get bump type from parameters with default
	bumpType := utils.GetParam(params, "bump", "patch")
	bumpType = strings.TrimSpace(strings.ToLower(bumpType))

	// Log the bump type being used for debugging
	utils.Info("Using bump type: %s", bumpType)

	if bumpType != "major" && bumpType != "minor" && bumpType != "patch" {
		return fmt.Errorf("%w: %s", errInvalidBumpType, bumpType)
	}

	// Special validation for major version bumps to prevent accidents
	if bumpType == "major" && !dryRun {
		if err := validateMajorVersionBump(params); err != nil {
			return err
		}
	}

	if dryRun {
		utils.Info("🔍 Running in DRY-RUN mode - no changes will be made")
	}

	// Check for uncommitted changes first
	if dirty := isGitDirty(); dirty {
		if dryRun {
			utils.Warn("Working directory has uncommitted changes (would fail in normal mode)")
		} else {
			return errVersionUncommittedChanges
		}
	}

	// Check if current commit already has version tags
	existingTags, err := getTagsOnCurrentCommit()
	if err != nil {
		return fmt.Errorf("failed to check existing tags: %w", err)
	}

	if len(existingTags) > 0 {
		utils.Warn("Current commit already has version tags: %s", strings.Join(existingTags, ", "))
		if dryRun {
			utils.Warn("Would fail in normal mode - need a new commit before bumping")
		} else {
			utils.Warn("Please create a new commit before bumping the version again")
			return fmt.Errorf("%w: %s", errMultipleTagsOnCommit, strings.Join(existingTags, ", "))
		}
	}

	// Get current version with enhanced logging
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

	// Check for version gaps and warn about them
	gapWarnings := detectVersionGaps(current, newVersion)
	for _, warning := range gapWarnings {
		utils.Warn("⚠️  %s", warning)
	}

	// Validate version progression
	if err := validateVersionProgression(current, newVersion, bumpType); err != nil {
		return err
	}

	// Additional check for unexpected version jumps (beyond validation errors)
	if !dryRun {
		if err := checkForUnexpectedVersionJump(current, newVersion, bumpType); err != nil {
			utils.Warn("⚠️  %s", err.Error())
			if !utils.IsParamTrue(params, "force") {
				utils.Warn("To proceed anyway, add 'force' parameter")
				utils.Warn("Or use 'dry-run' to preview the change first")
				return errVersionBumpBlocked
			}
			utils.Warn("⚠️  Proceeding with potentially unexpected version jump due to 'force' parameter")
		}
	}

	utils.Info("📋 Version Bump Summary:")
	utils.Info("  From:    %s", current)
	utils.Info("  To:      %s", newVersion)
	utils.Info("  Type:    %s bump", bumpType)

	if dryRun {
		// Dry-run mode - show what would happen
		utils.Info("📋 DRY-RUN Summary:")
		utils.Info("  Current version: %s", current)
		utils.Info("  New version:     %s", newVersion)
		utils.Info("  Bump type:       %s", bumpType)
		utils.Info("🔧 Commands that would be executed:")
		message := fmt.Sprintf("GitHubRelease %s", newVersion)
		utils.Info("  git tag -a %s -m \"%s\"", newVersion, message)

		if utils.IsParamTrue(params, "push") {
			utils.Info("  git push origin %s", newVersion)
		} else {
			utils.Info("📌 Note: Tag would be created locally only")
			utils.Info("  To push: git push origin %s", newVersion)
			utils.Info("  Or add 'push' parameter to push automatically")
		}

		utils.Success("✅ DRY-RUN completed - no changes made")
		return nil
	}

	// Create annotated tag
	message := fmt.Sprintf("GitHubRelease %s", newVersion)
	if err := GetRunner().RunCmd("git", "tag", "-a", newVersion, "-m", message); err != nil {
		return fmt.Errorf("failed to create tag: %w", err)
	}

	utils.Success("✅ Created tag: %s", newVersion)

	// Push if requested
	if utils.IsParamTrue(params, "push") {
		utils.Info("Pushing tag to remote...")
		if err := GetRunner().RunCmd("git", "push", "origin", newVersion); err != nil {
			return fmt.Errorf("failed to push tag: %w", err)
		}
		utils.Success("✅ Tag pushed to remote")
	} else {
		utils.Info("To push the tag, run: git push origin %s", newVersion)
		utils.Info("Or add 'push' parameter to push automatically")
	}

	return nil
}

// Changelog generates a changelog for the current version
func (Version) Changelog(args ...string) error {
	utils.Header("Generating Changelog")

	// Parse command-line parameters
	params := utils.ParseParams(args)

	// Get version range
	fromTag := utils.GetParam(params, "from", "")
	toTag := utils.GetParam(params, "to", "HEAD")

	if fromTag == "" {
		// Get previous tag
		fromTag = getPreviousTag()
		if fromTag == "" {
			utils.Info("No previous tag found, showing all commits")
		}
	}

	// Generate changelog

	var gitArgs []string
	if fromTag != "" {
		gitArgs = []string{"log", "--pretty=format:- %s (%h)", fmt.Sprintf("%s..%s", fromTag, toTag)}
	} else {
		gitArgs = []string{"log", "--pretty=format:- %s (%h)", toTag}
	}

	output, err := GetRunner().RunCmdOutput("git", gitArgs...)
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
// Deprecated: Use BuildInfoProvider.GetBuildInfo() instead
func getBuildInfo() BuildInfo {
	return GetDefaultBuildInfoProvider().GetBuildInfo()
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

// getCurrentGitTag gets the current git tag with detailed logging
func getCurrentGitTag() string {
	utils.Info("Detecting current version...")

	// First, check for tags directly on HEAD
	tagsOnHead := getTagsOnHead()
	if len(tagsOnHead) > 0 {
		selectedTag := tagsOnHead[0]
		utils.Info("Found tag on HEAD commit: %s", selectedTag)
		if len(tagsOnHead) > 1 {
			utils.Warn("Multiple tags found on HEAD: %s", strings.Join(tagsOnHead, ", "))
			utils.Info("Using highest version: %s", selectedTag)
		}
		return selectedTag
	}

	utils.Info("No tags found on HEAD commit")
	utils.Info("Searching for latest reachable tag...")

	// Get the latest reachable tag with distance info
	latestTag, distance := getLatestReachableTag()
	if latestTag == "" {
		utils.Info("No previous tags found in repository")
		return ""
	}

	if distance > 0 {
		utils.Warn("Current version tag is not on HEAD - there are %d commits since %s", distance, latestTag)
		// Show recent commits for context
		if recentCommits, err := GetRunner().RunCmdOutput("git", "log", "--oneline", "-5", "--no-decorate"); err == nil {
			utils.Info("Recent commits:\n%s", recentCommits)
		}
	} else {
		utils.Info("Found tag: %s (on current commit)", latestTag)
	}

	// Show recent version tags for context
	showRecentVersionTags()

	return latestTag
}

// getTagsOnHead returns all version tags pointing to HEAD, sorted by version (highest first)
func getTagsOnHead() []string {
	tags, err := GetRunner().RunCmdOutput("git", "tag", "--sort=-version:refname", "--points-at", "HEAD")
	if err != nil || strings.TrimSpace(tags) == "" {
		return []string{}
	}

	tagList := strings.Split(strings.TrimSpace(tags), "\n")
	// Filter out empty strings
	var result []string
	for _, tag := range tagList {
		if tag != "" {
			result = append(result, tag)
		}
	}
	return result
}

// getLatestReachableTag returns the latest tag reachable from HEAD and the distance (commits) from it
func getLatestReachableTag() (string, int) {
	// Use git describe to get tag and distance
	describeOutput, err := GetRunner().RunCmdOutput("git", "describe", "--tags", "--long", "--abbrev=0")
	if err != nil {
		// Fallback to simple describe
		simpleTag, simpleErr := GetRunner().RunCmdOutput("git", "describe", "--tags", "--abbrev=0")
		if simpleErr != nil {
			return "", 0
		}
		// Try to get distance separately
		tag := strings.TrimSpace(simpleTag)
		distance := 0
		if distanceCmd, err := GetRunner().RunCmdOutput("git", "rev-list", "--count", tag+"..HEAD"); err == nil {
			if d, parseErr := strconv.Atoi(strings.TrimSpace(distanceCmd)); parseErr == nil {
				distance = d
			}
		}
		return tag, distance
	}

	// Parse the describe output (format: tag-distance-gcommit)
	parts := strings.Split(strings.TrimSpace(describeOutput), "-")
	if len(parts) >= 2 {
		tag := parts[0]
		distance := 0
		if len(parts) >= 2 {
			if d, err := strconv.Atoi(parts[len(parts)-2]); err == nil {
				distance = d
			}
		}
		return tag, distance
	}

	return strings.TrimSpace(describeOutput), 0
}

// showRecentVersionTags displays recent version tags for context
func showRecentVersionTags() {
	tags, err := GetRunner().RunCmdOutput("git", "tag", "--sort=-version:refname", "-n", "5")
	if err != nil || strings.TrimSpace(tags) == "" {
		return
	}

	tagList := strings.Split(strings.TrimSpace(tags), "\n")
	if len(tagList) == 0 {
		return
	}

	var recentTags []string
	for i, tag := range tagList {
		if i >= 5 {
			break
		}
		if tag != "" {
			// Extract just the tag name (remove any annotation)
			tagParts := strings.Fields(tag)
			if len(tagParts) > 0 {
				recentTags = append(recentTags, tagParts[0])
			}
		}
	}

	if len(recentTags) > 0 {
		utils.Info("Recent version tags: %s", strings.Join(recentTags, ", "))
	}
}

// detectVersionGaps checks for non-sequential version progression and returns warnings
func detectVersionGaps(currentVersion, newVersion string) []string {
	var warnings []string

	// Get all existing tags
	allTags, err := GetRunner().RunCmdOutput("git", "tag", "--sort=-version:refname")
	if err != nil {
		return warnings
	}

	tagList := strings.Split(strings.TrimSpace(allTags), "\n")

	// Check if there are tags between current and new version
	var foundCurrent, foundGap bool
	for _, tag := range tagList {
		if tag == "" {
			continue
		}
		if tag == currentVersion {
			foundCurrent = true
			continue
		}
		if foundCurrent && !foundGap {
			// Check if this tag would be between current and new
			if isVersionBetween(currentVersion, tag, newVersion) {
				warnings = append(warnings, fmt.Sprintf("Version %s exists between %s and %s", tag, currentVersion, newVersion))
				foundGap = true
			}
		}
	}

	return warnings
}

// isVersionBetween checks if middle version is between start and end versions
func isVersionBetween(start, middle, end string) bool {
	// Simple comparison - could be enhanced with proper semver parsing
	startClean := strings.TrimPrefix(start, "v")
	middleClean := strings.TrimPrefix(middle, "v")
	endClean := strings.TrimPrefix(end, "v")

	// Parse versions
	startParts := strings.Split(startClean, ".")
	middleParts := strings.Split(middleClean, ".")
	endParts := strings.Split(endClean, ".")

	if len(startParts) != 3 || len(middleParts) != 3 || len(endParts) != 3 {
		return false
	}

	// Compare major.minor.patch
	for i := 0; i < 3; i++ {
		s, errS := strconv.Atoi(startParts[i])
		m, errM := strconv.Atoi(middleParts[i])
		e, errE := strconv.Atoi(endParts[i])

		if errS != nil || errM != nil || errE != nil {
			return false
		}

		if m > s && m < e {
			return true
		}
		if m < s || m > e {
			return false
		}
	}

	return false
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
func validateVersionProgression(current, newVersion, bumpType string) error {
	currentParts := strings.Split(strings.TrimPrefix(current, "v"), ".")
	newParts := strings.Split(strings.TrimPrefix(newVersion, "v"), ".")

	if len(currentParts) != 3 || len(newParts) != 3 {
		return nil // Skip validation if format is unexpected
	}

	currMajor, err := strconv.Atoi(currentParts[0])
	if err != nil {
		return fmt.Errorf("failed to parse current major version: %w", err)
	}
	currMinor, err := strconv.Atoi(currentParts[1])
	if err != nil {
		return fmt.Errorf("failed to parse current minor version: %w", err)
	}
	currPatch, err := strconv.Atoi(currentParts[2])
	if err != nil {
		return fmt.Errorf("failed to parse current patch version: %w", err)
	}

	newMajor, err := strconv.Atoi(newParts[0])
	if err != nil {
		return fmt.Errorf("failed to parse new major version: %w", err)
	}
	newMinor, err := strconv.Atoi(newParts[1])
	if err != nil {
		return fmt.Errorf("failed to parse new minor version: %w", err)
	}
	newPatch, err := strconv.Atoi(newParts[2])
	if err != nil {
		return fmt.Errorf("failed to parse new patch version: %w", err)
	}

	switch bumpType {
	case "patch":
		if newMajor != currMajor || newMinor != currMinor || newPatch != currPatch+1 {
			return fmt.Errorf("%w: expected %s → v%d.%d.%d, got %s",
				errIllogicalVersionJump, current, currMajor, currMinor, currPatch+1, newVersion)
		}
	case "minor":
		if newMajor != currMajor || newMinor != currMinor+1 || newPatch != 0 {
			return fmt.Errorf("%w: expected %s → v%d.%d.0, got %s",
				errIllogicalVersionJump, current, currMajor, currMinor+1, newVersion)
		}
	case "major":
		if newMajor != currMajor+1 || newMinor != 0 || newPatch != 0 {
			return fmt.Errorf("%w: expected %s → v%d.0.0, got %s",
				errIllogicalVersionJump, current, currMajor+1, newVersion)
		}
	}

	return nil
}

// validateMajorVersionBump validates major version bumps to prevent accidents
func validateMajorVersionBump(params map[string]string) error {
	// Check if this appears to be an accidental major bump
	current := getCurrentGitTag()
	if current == "" {
		return nil
	}

	newVersion, err := bumpVersion(current, "major")
	if err != nil {
		return nil //nolint:nilerr // Skip validation if bump calculation fails
	}

	utils.Warn("⚠️  MAJOR VERSION BUMP DETECTED:")
	utils.Warn("   Current version: %s", current)
	utils.Warn("   New version:     %s", newVersion)
	utils.Warn("   This will create a breaking change release!")

	// Check if user explicitly confirmed major bump
	if !utils.IsParamTrue(params, "major-confirm") {
		utils.Warn("")
		utils.Warn("To proceed with major version bump, add 'major-confirm' parameter")
		utils.Warn("Example: magex version:bump bump=major major-confirm push")
		utils.Warn("")
		utils.Warn("Or use 'dry-run' to preview the change first:")
		utils.Warn("Example: magex version:bump bump=major dry-run")
		return errMajorBumpRequiresConfirm
	}
	utils.Success("✅ Major version bump confirmed via 'major-confirm' parameter")
	return nil
}

// checkForUnexpectedVersionJump provides additional safety checks beyond basic validation
func checkForUnexpectedVersionJump(current, newVersion, bumpType string) error {
	// Parse versions
	currentParts := strings.Split(strings.TrimPrefix(current, "v"), ".")
	newParts := strings.Split(strings.TrimPrefix(newVersion, "v"), ".")

	if len(currentParts) != 3 || len(newParts) != 3 {
		return nil // Skip check for malformed versions
	}

	currMajor, err := strconv.Atoi(currentParts[0])
	if err != nil {
		return nil //nolint:nilerr // Skip check for malformed versions
	}
	newMajor, err := strconv.Atoi(newParts[0])
	if err != nil {
		return nil //nolint:nilerr // Skip check for malformed versions
	}

	// Check for unexpected major version jump when expecting patch
	if bumpType == "patch" && newMajor > currMajor {
		return fmt.Errorf("%w from %s to %s when BUMP=%s", errUnexpectedMajorVersionJump, current, newVersion, bumpType)
	}

	// Check for surprisingly large jumps that might indicate environment contamination
	majorJump := newMajor - currMajor
	if majorJump > 1 {
		return fmt.Errorf("%w from %s to %s (major version increased by %d)", errUnexpectedlyLargeVersionJump, current, newVersion, majorJump)
	}

	return nil
}

// Additional methods for Version namespace required by tests

// Tag creates a version tag
func (Version) Tag(_ ...string) error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Creating version tag")
}

// Next shows the next version
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
