// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Static errors for releases operations
var (
	errReleasesVersionRequired     = errors.New("VERSION environment variable is required")
	errGitHubTokenRequiredReleases = errors.New("GITHUB_TOKEN environment variable is required")
	errReleaseAlreadyExists        = errors.New("release already exists")
	errInvalidPlatformFormat       = errors.New("invalid platform format")
	errInvalidFromChannel          = errors.New("invalid from channel")
	errInvalidToChannel            = errors.New("invalid to channel")
)

// Releases namespace for multi-channel release management
type Releases mg.Namespace

// ReleaseChannel represents different release channels
type ReleaseChannel string

const (
	// StableRelease represents the stable release channel
	StableRelease ReleaseChannel = "stable"
	// BetaRelease represents the beta release channel
	BetaRelease ReleaseChannel = "beta"
	// EdgeRelease represents the edge release channel
	EdgeRelease ReleaseChannel = "edge"
)

// MultiChannelReleaseConfig contains release configuration for multi-channel releases
type MultiChannelReleaseConfig struct {
	Channel     ReleaseChannel
	Version     string
	Prerelease  bool
	Draft       bool
	Notes       string
	Assets      []string
	Platforms   []string
	GitHubToken string
}

// Create creates a new release (alias for Stable for compatibility)
func (Releases) Create() error {
	return Releases{}.Stable()
}

// Publish publishes a release (alias for Stable for compatibility)
func (Releases) Publish() error {
	return Releases{}.Stable()
}

// Stable creates a stable release
func (Releases) Stable() error {
	utils.Header("🚀 Creating Stable Release")

	config := &MultiChannelReleaseConfig{
		Channel:    StableRelease,
		Prerelease: false,
		Draft:      false,
		Platforms:  []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64"},
	}

	return createRelease(config)
}

// Beta creates a beta release
func (Releases) Beta() error {
	utils.Header("🧪 Creating Beta Release")

	config := &MultiChannelReleaseConfig{
		Channel:    BetaRelease,
		Prerelease: true,
		Draft:      false,
		Platforms:  []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64"},
	}

	return createRelease(config)
}

// Edge creates an edge release
func (Releases) Edge() error {
	utils.Header("⚡ Creating Edge Release")

	config := &MultiChannelReleaseConfig{
		Channel:    EdgeRelease,
		Prerelease: true,
		Draft:      false,
		Platforms:  []string{"linux/amd64", "darwin/amd64", "windows/amd64"},
	}

	return createRelease(config)
}

// Draft creates a draft release
func (Releases) Draft() error {
	utils.Header("📝 Creating Draft Release")

	config := &MultiChannelReleaseConfig{
		Channel:    StableRelease,
		Prerelease: false,
		Draft:      true,
		Platforms:  []string{"linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64", "windows/amd64"},
	}

	return createRelease(config)
}

// Promote promotes a release from one channel to another
func (Releases) Promote() error {
	utils.Header("📈 Promoting Release")

	fromChannel := utils.GetEnv("FROM_CHANNEL", "beta")
	toChannel := utils.GetEnv("TO_CHANNEL", "stable")
	version := utils.GetEnv("VERSION", "")

	if version == "" {
		return errReleasesVersionRequired
	}

	utils.Info("Promoting %s from %s to %s", version, fromChannel, toChannel)

	// Validate channels
	if err := validateChannels(fromChannel, toChannel); err != nil {
		return err
	}

	// Get existing release
	release := getExistingRelease(version)

	// Update release properties
	newConfig := &MultiChannelReleaseConfig{
		Channel:    ReleaseChannel(toChannel),
		Version:    version,
		Prerelease: toChannel != "stable",
		Draft:      false,
		Notes:      release.Notes,
		Assets:     release.Assets,
	}

	// Create promoted release
	return promoteRelease(newConfig)
}

// Status shows release status across channels
func (Releases) Status() error {
	utils.Header("📊 Release Status")

	channels := []ReleaseChannel{StableRelease, BetaRelease, EdgeRelease}

	utils.Info("\n📈 Multi-Channel Release Status:")
	utils.Info("Channel   Latest Version   Released        Downloads")
	utils.Info("--------  --------------   --------------- ----------")

	for _, channel := range channels {
		info := getChannelInfo()

		releasedAt := info.ReleasedAt.Format("2006-01-02 15:04")
		fmt.Printf("%-8s  %-14s   %-15s %d\n", channel, info.Version, releasedAt, info.Downloads)
	}

	return nil
}

// Channels lists available release channels
func (Releases) Channels() error {
	utils.Header("📺 Release Channels")

	channels := []struct {
		Name        ReleaseChannel
		Description string
		Stability   string
		Audience    string
	}{
		{StableRelease, "Production-ready releases", "High", "All users"},
		{BetaRelease, "Feature-complete pre-releases", "Medium", "Early adopters"},
		{EdgeRelease, "Latest development builds", "Low", "Developers"},
	}

	utils.Info("\n📋 Available Release Channels:")
	utils.Info("Channel   Description                     Stability   Audience")
	utils.Info("--------  -----------------------------   ---------   ----------------")

	for _, ch := range channels {
		fmt.Printf("%-8s  %-29s   %-9s   %s\n", ch.Name, ch.Description, ch.Stability, ch.Audience)
	}

	utils.Info("\nUsage:")
	utils.Info("  mage releases:stable    # Create stable release")
	utils.Info("  mage releases:beta      # Create beta release")
	utils.Info("  mage releases:edge      # Create edge release")
	utils.Info("  mage releases:status    # Show release status")

	return nil
}

// Cleanup removes old releases
func (Releases) Cleanup() error {
	utils.Header("🧹 Cleaning Up Old Releases")

	// Get cleanup configuration
	keepStable := utils.GetEnvInt("KEEP_STABLE", 5)
	keepBeta := utils.GetEnvInt("KEEP_BETA", 10)
	keepEdge := utils.GetEnvInt("KEEP_EDGE", 5)

	utils.Info("Cleanup policy:")
	utils.Info("  Stable releases to keep: %d", keepStable)
	utils.Info("  Beta releases to keep: %d", keepBeta)
	utils.Info("  Edge releases to keep: %d", keepEdge)

	// Cleanup each channel
	channels := map[ReleaseChannel]int{
		StableRelease: keepStable,
		BetaRelease:   keepBeta,
		EdgeRelease:   keepEdge,
	}

	totalCleaned := 0
	for channel, keep := range channels {
		cleaned, err := cleanupChannel(channel, keep)
		if err != nil {
			utils.Warn("Failed to cleanup %s channel: %v", channel, err)
			continue
		}
		totalCleaned += cleaned
		utils.Info("Cleaned %d releases from %s channel", cleaned, channel)
	}

	if totalCleaned > 0 {
		utils.Success("Cleaned up %d old releases", totalCleaned)
	} else {
		utils.Info("No releases to clean up")
	}

	return nil
}

// Helper functions

// createRelease creates a new release
func createRelease(config *MultiChannelReleaseConfig) error {
	// Get version
	if config.Version == "" {
		version := utils.GetEnv("VERSION", "")
		if version == "" {
			// Generate version based on channel
			version = generateVersion(config.Channel)
		}
		config.Version = version
	}

	// Get GitHub token
	config.GitHubToken = os.Getenv("GITHUB_TOKEN")
	if config.GitHubToken == "" {
		config.GitHubToken = os.Getenv("github_token")
	}

	if config.GitHubToken == "" {
		return errGitHubTokenRequiredReleases
	}

	utils.Info("Creating %s release: %s", config.Channel, config.Version)

	// Check if release already exists
	if releaseExists(config.Version) {
		return fmt.Errorf("%w: %s", errReleaseAlreadyExists, config.Version)
	}

	// Generate release notes
	if config.Notes == "" {
		notes, err := generateReleaseNotes(config)
		if err != nil {
			utils.Warn("Failed to generate release notes: %v", err)
			config.Notes = fmt.Sprintf("Release %s", config.Version)
		} else {
			config.Notes = notes
		}
	}

	// Build assets
	if err := buildReleaseAssets(config); err != nil {
		return fmt.Errorf("failed to build release assets: %w", err)
	}

	// Create git tag
	if err := createGitTag(config); err != nil {
		return fmt.Errorf("failed to create git tag: %w", err)
	}

	// Create GitHub release
	if err := createGitHubRelease(config); err != nil {
		return fmt.Errorf("failed to create GitHub release: %w", err)
	}

	utils.Success("Successfully created %s release: %s", config.Channel, config.Version)
	return nil
}

// generateVersion generates a version number based on channel
func generateVersion(channel ReleaseChannel) string {
	baseVersion := getVersion()
	if baseVersion == versionDev {
		baseVersion = "0.1.0"
	}

	// Remove 'v' prefix if present
	baseVersion = strings.TrimPrefix(baseVersion, "v")

	switch channel {
	case StableRelease:
		return fmt.Sprintf("v%s", baseVersion)
	case BetaRelease:
		timestamp := time.Now().Format("20060102")
		return fmt.Sprintf("v%s-beta.%s", baseVersion, timestamp)
	case EdgeRelease:
		timestamp := time.Now().Format("20060102.150405")
		return fmt.Sprintf("v%s-edge.%s", baseVersion, timestamp)
	default:
		return fmt.Sprintf("v%s", baseVersion)
	}
}

// releaseExists checks if a release already exists
func releaseExists(version string) bool {
	// Check if git tag exists
	err := GetRunner().RunCmd("git", "rev-parse", version)
	return err == nil
}

// generateReleaseNotes generates release notes
func generateReleaseNotes(config *MultiChannelReleaseConfig) (string, error) {
	// Get previous release
	prevRelease := getPreviousRelease(config.Channel)

	// Generate changelog
	var args []string
	if prevRelease != "" {
		args = []string{"log", "--pretty=format:- %s", fmt.Sprintf("%s..HEAD", prevRelease)}
	} else {
		args = []string{"log", "--pretty=format:- %s", "--max-count=20"}
	}

	output, err := GetRunner().RunCmdOutput("git", args...)
	if err != nil {
		return "", err
	}

	// Format release notes
	notes := fmt.Sprintf("# %s Release %s\n\n", config.Channel, config.Version)

	switch config.Channel {
	case StableRelease:
		notes += "🚀 **Stable Release** - Production ready\n\n"
	case BetaRelease:
		notes += "🧪 **Beta Release** - Feature complete, testing phase\n\n"
	case EdgeRelease:
		notes += "⚡ **Edge Release** - Latest development build\n\n"
	}

	if strings.TrimSpace(output) != "" {
		notes += "## Changes\n\n"
		notes += output
	}

	notes += "\n\n---\n"
	notes += "Generated by MAGE-X Release System"

	return notes, nil
}

// buildReleaseAssets builds release assets for all platforms
func buildReleaseAssets(config *MultiChannelReleaseConfig) error {
	utils.Info("Building release assets...")

	// Clean dist directory
	distDir := "dist"
	if err := utils.CleanDir(distDir); err != nil {
		return fmt.Errorf("failed to clean dist directory: %w", err)
	}

	// Use existing release tools
	if utils.CommandExists("goreleaser") {
		return buildWithGoreleaser(config)
	}

	// Fallback to manual build
	return buildManually(config)
}

// buildWithGoreleaser builds using goreleaser
func buildWithGoreleaser(config *MultiChannelReleaseConfig) error {
	args := []string{"release"}

	if config.Draft {
		args = append(args, "--skip=publish")
	}

	args = append(args, "--clean")

	// Set environment variables temporarily
	oldToken := os.Getenv("GITHUB_TOKEN")
	oldChannel := os.Getenv("RELEASE_CHANNEL")

	if err := os.Setenv("GITHUB_TOKEN", config.GitHubToken); err != nil {
		return fmt.Errorf("failed to set GITHUB_TOKEN: %w", err)
	}
	if err := os.Setenv("RELEASE_CHANNEL", string(config.Channel)); err != nil {
		return fmt.Errorf("failed to set RELEASE_CHANNEL: %w", err)
	}

	defer func() {
		if oldToken == "" {
			if err := os.Unsetenv("GITHUB_TOKEN"); err != nil {
				// Log error but don't fail - this is cleanup
				log.Printf("failed to unset GITHUB_TOKEN during cleanup: %v", err)
			}
		} else {
			if err := os.Setenv("GITHUB_TOKEN", oldToken); err != nil {
				// Log error but don't fail - this is cleanup
				log.Printf("failed to restore GITHUB_TOKEN during cleanup: %v", err)
			}
		}
		if oldChannel == "" {
			if err := os.Unsetenv("RELEASE_CHANNEL"); err != nil {
				// Log error but don't fail - this is cleanup
				log.Printf("failed to unset RELEASE_CHANNEL during cleanup: %v", err)
			}
		} else {
			if err := os.Setenv("RELEASE_CHANNEL", oldChannel); err != nil {
				// Log error but don't fail - this is cleanup
				log.Printf("failed to restore RELEASE_CHANNEL during cleanup: %v", err)
			}
		}
	}()

	return GetRunner().RunCmd("goreleaser", args...)
}

// buildManually builds assets manually
func buildManually(config *MultiChannelReleaseConfig) error {
	binaryName, err := getBinaryName()
	if err != nil {
		return err
	}

	for _, platform := range config.Platforms {
		outputPath, err := buildForPlatform(binaryName, platform)
		if err != nil {
			return err
		}
		config.Assets = append(config.Assets, outputPath)
	}

	return nil
}

func getBinaryName() (string, error) {
	module, err := getModuleName()
	if err != nil {
		return "", err
	}
	return filepath.Base(module), nil
}

func buildForPlatform(binaryName, platform string) (string, error) {
	goos, goarch, err := parsePlatform(platform)
	if err != nil {
		return "", err
	}

	outputPath := generateOutputPath(binaryName, goos, goarch)
	utils.Info("Building %s", filepath.Base(outputPath))

	// Build with proper environment
	err = withBuildEnvironment(goos, goarch, func() error {
		args := []string{"build", "-o", outputPath, "-ldflags", "-s -w"}
		return GetRunner().RunCmd("go", args...)
	})
	if err != nil {
		return "", fmt.Errorf("failed to build %s: %w", filepath.Base(outputPath), err)
	}

	return outputPath, nil
}

func parsePlatform(platform string) (goos, arch string, err error) {
	parts := strings.Split(platform, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("%w: %s", errInvalidPlatformFormat, platform)
	}
	return parts[0], parts[1], nil
}

func generateOutputPath(binaryName, goos, goarch string) string {
	outputName := fmt.Sprintf("%s-%s-%s", binaryName, goos, goarch)
	if goos == "windows" {
		outputName += ".exe"
	}
	return filepath.Join("dist", outputName)
}

func withBuildEnvironment(goos, goarch string, fn func() error) error {
	// Save current environment
	env := buildEnvironment{
		goos:   os.Getenv("GOOS"),
		goarch: os.Getenv("GOARCH"),
		cgo:    os.Getenv("CGO_ENABLED"),
	}

	// Set build environment
	if err := os.Setenv("GOOS", goos); err != nil {
		return fmt.Errorf("failed to set GOOS: %w", err)
	}
	if err := os.Setenv("GOARCH", goarch); err != nil {
		return fmt.Errorf("failed to set GOARCH: %w", err)
	}
	if err := os.Setenv("CGO_ENABLED", "0"); err != nil {
		return fmt.Errorf("failed to set CGO_ENABLED: %w", err)
	}

	// Restore environment on exit
	defer env.restore()

	return fn()
}

type buildEnvironment struct {
	goos   string
	goarch string
	cgo    string
}

func (e *buildEnvironment) restore() {
	restoreEnv("GOOS", e.goos)
	restoreEnv("GOARCH", e.goarch)
	restoreEnv("CGO_ENABLED", e.cgo)
}

func restoreEnv(key, value string) {
	if value == "" {
		if err := os.Unsetenv(key); err != nil {
			// Log error but don't fail - this is cleanup
			log.Printf("failed to unset environment variable %s during cleanup: %v", key, err)
		}
	} else {
		if err := os.Setenv(key, value); err != nil {
			// Log error but don't fail - this is cleanup
			log.Printf("failed to set environment variable %s during cleanup: %v", key, err)
		}
	}
}

// createGitTag creates a git tag
func createGitTag(config *MultiChannelReleaseConfig) error {
	// Create annotated tag
	message := fmt.Sprintf("%s release %s", config.Channel, config.Version)

	if err := GetRunner().RunCmd("git", "tag", "-a", config.Version, "-m", message); err != nil {
		return err
	}

	// Push tag
	return GetRunner().RunCmd("git", "push", "origin", config.Version)
}

// createGitHubRelease creates a GitHub release
func createGitHubRelease(config *MultiChannelReleaseConfig) error {
	// Use GitHub CLI if available
	if utils.CommandExists("gh") {
		return createGitHubReleaseWithCLI(config)
	}

	// Otherwise, provide instructions
	utils.Info("GitHub CLI not found. Create release manually:")
	utils.Info("  1. Go to your repository on GitHub")
	utils.Info("  2. Click 'Releases' -> 'Create a new release'")
	utils.Info("  3. Tag: %s", config.Version)
	utils.Info("  4. Title: %s", config.Version)
	utils.Info("  5. Description: %s", config.Notes)
	utils.Info("  6. Upload assets from dist/ directory")

	return nil
}

// createGitHubReleaseWithCLI creates release using GitHub CLI
func createGitHubReleaseWithCLI(config *MultiChannelReleaseConfig) error {
	args := []string{"release", "create", config.Version}

	// Add flags
	if config.Prerelease {
		args = append(args, "--prerelease")
	}

	if config.Draft {
		args = append(args, "--draft")
	}

	// Add title and notes
	args = append(args, "--title", config.Version, "--notes", config.Notes)

	// Add assets
	args = append(args, config.Assets...)

	return GetRunner().RunCmd("gh", args...)
}

// Placeholder implementations for other helper functions
func validateChannels(from, to string) error {
	validChannels := []string{"stable", "beta", "edge"}

	fromValid := false
	toValid := false

	for _, ch := range validChannels {
		if from == ch {
			fromValid = true
		}
		if to == ch {
			toValid = true
		}
	}

	if !fromValid {
		return fmt.Errorf("%w: %s", errInvalidFromChannel, from)
	}
	if !toValid {
		return fmt.Errorf("%w: %s", errInvalidToChannel, to)
	}

	return nil
}

func getExistingRelease(version string) *ReleaseInfo {
	// Placeholder implementation
	return &ReleaseInfo{
		Version: version,
		Notes:   fmt.Sprintf("Release %s", version),
		Assets:  []string{},
	}
}

func promoteRelease(config *MultiChannelReleaseConfig) error {
	// Placeholder implementation
	utils.Info("Promoting release %s to %s channel", config.Version, config.Channel)
	return nil
}

func getChannelInfo() *ChannelInfo {
	// Placeholder implementation
	return &ChannelInfo{
		Version:    "v1.0.0",
		ReleasedAt: time.Now().Add(-24 * time.Hour),
		Downloads:  1234,
	}
}

func cleanupChannel(_ ReleaseChannel, _ int) (int, error) {
	// Placeholder implementation
	return 0, nil
}

func getPreviousRelease(_ ReleaseChannel) string {
	// Placeholder implementation
	return ""
}

// ReleaseInfo contains release information
type ReleaseInfo struct {
	Version string
	Notes   string
	Assets  []string
}

// ChannelInfo contains release channel information
type ChannelInfo struct {
	Version    string
	ReleasedAt time.Time
	Downloads  int
}
