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
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Static errors to satisfy err113 linter
var (
	errNoReleasesFound       = errors.New("no releases found")
	errUpdateVersionRequired = errors.New("VERSION environment variable is required")
	errNoBetaReleasesFound   = errors.New("no beta releases found")
)

// Update namespace for auto-update functionality
type Update mg.Namespace

// UpdateChannel represents a release channel
type UpdateChannel string

const (
	// StableChannel represents the stable release channel.
	StableChannel UpdateChannel = "stable"
	// BetaChannel represents the beta release channel.
	BetaChannel UpdateChannel = "beta"
	// EdgeChannel represents the edge release channel.
	EdgeChannel UpdateChannel = "edge"
)

// UpdateInfo contains update information
type UpdateInfo struct {
	Channel         UpdateChannel
	CurrentVersion  string
	LatestVersion   string
	UpdateAvailable bool
	ReleaseNotes    string
	DownloadURL     string
}

// Check checks for available updates in the specified channel
func (Update) Check() error {
	utils.Header("Checking for Updates")

	channel := getUpdateChannel()
	utils.Info("Update channel: %s", channel)

	info, err := checkForUpdates(channel)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	utils.Info("Current version: %s", info.CurrentVersion)
	utils.Info("Latest version:  %s", info.LatestVersion)

	if info.UpdateAvailable {
		utils.Success("\n🎉 Update available!")
		if info.ReleaseNotes != "" {
			utils.Info("\nRelease Notes:")
			utils.Info("%s", info.ReleaseNotes)
		}
		utils.Info("\nRun 'mage update:install' to update")
	} else {
		utils.Success("\n✅ You are running the latest version")
	}

	return nil
}

// Install installs the latest update
func (Update) Install() error {
	utils.Header("Installing Update")

	channel := getUpdateChannel()
	info, err := checkForUpdates(channel)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !info.UpdateAvailable {
		utils.Success("Already running the latest version: %s", info.CurrentVersion)
		return nil
	}

	utils.Info("Updating from %s to %s", info.CurrentVersion, info.LatestVersion)

	// Create update directory
	updateDir := filepath.Join(os.TempDir(), "mage-update")
	if err := os.MkdirAll(updateDir, 0o750); err != nil {
		return fmt.Errorf("failed to create update directory: %w", err)
	}
	defer func() {
		// Ignore error in defer cleanup
		if err := os.RemoveAll(updateDir); err != nil {
			// Best effort cleanup - ignore error
			log.Printf("failed to clean up update directory %s: %v", updateDir, err)
		}
	}()

	// Download update
	if err := downloadUpdate(info, updateDir); err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	// Install update
	if err := installUpdate(); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}

	utils.Success("Successfully updated to version %s", info.LatestVersion)
	utils.Info("Please restart your application to use the new version")

	// Save update record
	saveUpdateRecord(info)

	return nil
}

// Auto enables automatic update checking
func (Update) Auto() error {
	utils.Header("Configuring Automatic Updates")

	enabled := utils.GetEnv("ENABLED", approvalTrue) == approvalTrue
	interval := utils.GetEnv("INTERVAL", "24h")
	channel := getUpdateChannel()

	configPath := getUpdateConfigPath()

	config := map[string]interface{}{
		"enabled":   enabled,
		"interval":  interval,
		"channel":   channel,
		"lastCheck": time.Now().Format(time.RFC3339),
	}

	// Save configuration
	configData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, configData, 0o600); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	if enabled {
		utils.Success("Automatic updates enabled")
		utils.Info("Channel: %s", channel)
		utils.Info("Check interval: %s", interval)
		utils.Info("Configuration saved to: %s", configPath)
	} else {
		utils.Success("Automatic updates disabled")
	}

	return nil
}

// History shows update history
func (Update) History() error {
	utils.Header("Update History")

	historyPath := getUpdateHistoryPath()
	if !utils.FileExists(historyPath) {
		utils.Info("No update history found")
		return nil
	}

	fileOps := fileops.New()
	var history []map[string]string
	if err := fileOps.JSON.ReadJSON(historyPath, &history); err != nil {
		return fmt.Errorf("failed to read history: %w", err)
	}

	if len(history) == 0 {
		utils.Info("No updates recorded")
		return nil
	}

	utils.Info("\nUpdate History:")
	utils.Info("Date                  From Version    To Version      Channel")
	utils.Info("-------------------   -------------   -------------   -------")

	for _, record := range history {
		date, err := time.Parse(time.RFC3339, record["date"])
		if err != nil {
			// If date parsing fails, use current time as fallback
			date = time.Now()
		}
		fmt.Printf("%-19s   %-13s   %-13s   %s\n",
			date.Format("2006-01-02 15:04:05"),
			record["from"],
			record["to"],
			record["channel"],
		)
	}

	return nil
}

// Rollback rolls back to a previous version
func (Update) Rollback() error {
	utils.Header("Rolling Back Update")

	version := utils.GetEnv("VERSION", "")
	if version == "" {
		return errUpdateVersionRequired
	}

	utils.Info("Rolling back to version %s", version)

	// Get module info
	module, err := utils.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %w", err)
	}

	// Install specific version
	pkg := fmt.Sprintf("%s@%s", module, version)

	if err := GetRunner().RunCmd("go", "install", pkg); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	utils.Success("Successfully rolled back to version %s", version)
	utils.Info("Please restart your application")

	return nil
}

// Helper functions

// getUpdateChannel returns the configured update channel
func getUpdateChannel() UpdateChannel {
	channel := strings.ToLower(utils.GetEnv("UPDATE_CHANNEL", "stable"))

	switch channel {
	case "beta":
		return BetaChannel
	case "edge":
		return EdgeChannel
	default:
		return StableChannel
	}
}

// checkForUpdates checks for available updates
func checkForUpdates(channel UpdateChannel) (*UpdateInfo, error) {
	current := getVersionInfo()

	// Get module info
	module, err := utils.GetModuleName()
	if err != nil {
		return nil, fmt.Errorf("failed to get module name: %w", err)
	}

	// Parse module to get owner/repo
	parts := strings.Split(module, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("%w: %s", errCannotParseGitHubInfo, module)
	}

	owner := parts[1]
	repo := parts[2]

	// Get releases based on channel
	release, err := getReleaseForChannel(owner, repo, channel)
	if err != nil {
		return nil, err
	}

	info := &UpdateInfo{
		Channel:         channel,
		CurrentVersion:  current,
		LatestVersion:   release.TagName,
		UpdateAvailable: isNewer(release.TagName, current),
		ReleaseNotes:    formatReleaseNotes(release.Body),
	}

	// Find appropriate asset
	assetName := fmt.Sprintf("%s-%s-%s", repo, runtime.GOOS, runtime.GOARCH)
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, assetName) {
			info.DownloadURL = asset.BrowserDownloadURL
			break
		}
	}

	return info, nil
}

// getReleaseForChannel gets the appropriate release for a channel
func getReleaseForChannel(owner, repo string, channel UpdateChannel) (*GitHubRelease, error) {
	switch channel {
	case StableChannel:
		return getLatestStableRelease(owner, repo)
	case BetaChannel:
		return getLatestBetaRelease(owner, repo)
	case EdgeChannel:
		return getLatestEdgeRelease(owner, repo)
	default:
		return getLatestStableRelease(owner, repo)
	}
}

// getLatestStableRelease gets the latest stable release
func getLatestStableRelease(owner, repo string) (*GitHubRelease, error) {
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
		// Ignore error in defer cleanup
		if err := resp.Body.Close(); err != nil {
			// Best effort cleanup - ignore error
			log.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response: %w", err)
		}
		return nil, fmt.Errorf("%w: %s", errGitHubAPIError, body)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	return &release, nil
}

// getLatestBetaRelease gets the latest beta release
func getLatestBetaRelease(owner, repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)

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
		// Ignore error in defer cleanup
		if err := resp.Body.Close(); err != nil {
			// Best effort cleanup - ignore error
			log.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response: %w", err)
		}
		return nil, fmt.Errorf("%w: %s", errGitHubAPIError, body)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	// Find latest beta or stable
	for _, release := range releases {
		if !release.Draft && (release.Prerelease || !release.Prerelease) {
			return &release, nil
		}
	}

	return nil, errNoBetaReleasesFound
}

// getLatestEdgeRelease gets the latest edge release (any release including pre-release)
func getLatestEdgeRelease(owner, repo string) (*GitHubRelease, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)

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
		// Ignore error in defer cleanup
		if err := resp.Body.Close(); err != nil {
			// Best effort cleanup - ignore error
			log.Printf("failed to close response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read error response: %w", err)
		}
		return nil, fmt.Errorf("%w: %s", errGitHubAPIError, body)
	}

	var releases []GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return nil, err
	}

	if len(releases) > 0 {
		return &releases[0], nil
	}

	return nil, errNoReleasesFound
}

// downloadUpdate downloads the update
func downloadUpdate(info *UpdateInfo, dir string) error {
	if info.DownloadURL == "" {
		// No binary asset, use go install
		return nil
	}

	utils.Info("Downloading update...")

	// Download binary
	req, err := http.NewRequestWithContext(context.Background(), "GET", info.DownloadURL, http.NoBody)
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		// Ignore error in defer cleanup
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Best effort cleanup - ignore error
			log.Printf("failed to close response body: %v", closeErr)
		}
	}()

	// Save to file
	filename := filepath.Base(info.DownloadURL)
	targetPath := filepath.Join(dir, filename)

	// Save response body to file
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fileOps := fileops.New()
	return fileOps.File.WriteFile(targetPath, data, 0o644)
}

// installUpdate installs the downloaded update
func installUpdate() error {
	// This is simplified - in reality would handle binary replacement
	// For now, we use go install

	module, err := utils.GetModuleName()
	if err != nil {
		return err
	}

	return GetRunner().RunCmd("go", "install", module+"@latest")
}

// getUpdateConfigPath returns the update configuration path
func getUpdateConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), ".config", "mage", "update.json")
	}
	return filepath.Join(home, ".config", "mage", "update.json")
}

// getUpdateHistoryPath returns the update history path
func getUpdateHistoryPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), ".config", "mage", "history.json")
	}
	return filepath.Join(home, ".config", "mage", "history.json")
}

// saveUpdateRecord saves an update record to history
func saveUpdateRecord(info *UpdateInfo) {
	historyPath := getUpdateHistoryPath()
	fileOps := fileops.New()

	var history []map[string]string
	if err := fileOps.JSON.ReadJSON(historyPath, &history); err != nil {
		// If file doesn't exist, start with empty history
		history = []map[string]string{}
	}

	record := map[string]string{
		"date":    time.Now().Format(time.RFC3339),
		"from":    info.CurrentVersion,
		"to":      info.LatestVersion,
		"channel": string(info.Channel),
	}

	history = append(history, record)

	// Keep only last 50 records
	if len(history) > 50 {
		history = history[len(history)-50:]
	}

	// Ensure directory exists and save
	if err := fileOps.File.MkdirAll(filepath.Dir(historyPath), 0o755); err != nil {
		// Best effort - ignore error in cleanup
		log.Printf("failed to create history directory: %v", err)
	}
	if err := fileOps.JSON.WriteJSONIndent(historyPath, history, "", "  "); err != nil {
		// Best effort - ignore error in cleanup
		log.Printf("failed to write update history: %v", err)
	}
}
