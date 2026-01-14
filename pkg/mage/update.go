// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
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

	"github.com/mrz1836/mage-x/pkg/common/env"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/utils"
)

const (
	// magexModule is the module path for the magex binary
	magexModule = "github.com/mrz1836/mage-x"

	// maxUpdateFileSize is the maximum allowed size for extracted files (500MB)
	// This prevents zip bomb attacks during update extraction
	maxUpdateFileSize = 500 * 1024 * 1024

	// downloadTimeout is the maximum time allowed for downloading update files
	downloadTimeout = 5 * time.Minute
)

// Static errors to satisfy err113 linter
var (
	errNoReleasesFound     = errors.New("no releases found")
	errNoBetaReleasesFound = errors.New("no beta releases found")
	errNoTarGzFound        = errors.New("no tar.gz file found in update directory")
	errMagexBinaryNotFound = errors.New("magex binary not found in extracted files")
	errPathTraversal       = errors.New("path traversal attempt detected")
	errChecksumMismatch    = errors.New("checksum verification failed")
	errFileTooLarge        = errors.New("file exceeds maximum allowed size")
	errChecksumFetchFailed = errors.New("failed to fetch checksums file")
	errChecksumNotFound    = errors.New("checksum not found in checksums file")
	errDownloadFailed      = errors.New("download failed")
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
	ChecksumSHA256  string // Expected SHA256 checksum of the downloaded file (hex-encoded)
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
		utils.Success("🎉 Update available!")
		if info.ReleaseNotes != "" {
			utils.Info("Release Notes:")
			utils.Info("%s", info.ReleaseNotes)
		}
		utils.Info("Run 'magex update:install' to update")
	} else {
		utils.Success("✅ You are running the latest version")
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

	// Create update directory with random suffix for security
	// Using MkdirTemp prevents symlink race attacks on predictable paths
	updateDir, err := os.MkdirTemp("", "mage-update-*")
	if err != nil {
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
	if err := installUpdate(info, updateDir); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}

	utils.Success("Successfully updated to version %s", info.LatestVersion)

	// Clear the update cache so the user doesn't see stale "update available" notifications
	// The next run will perform a fresh check and correctly detect the new version
	cache := NewUpdateNotifyCache()
	if err := cache.Clear(); err != nil {
		// Non-fatal: log but don't fail the installation
		log.Printf("failed to clear update cache: %v", err)
	}

	utils.Info("Please restart your application to use the new version")

	return nil
}

// Helper functions

// getUpdateChannel returns the configured update channel
func getUpdateChannel() UpdateChannel {
	channel := strings.ToLower(env.GetString("UPDATE_CHANNEL", "stable"))

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
	current := getVersionInfoForUpdate()

	// Always use the magex module for updates, regardless of current working directory
	module := magexModule

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
		return nil, fmt.Errorf("failed to get release for channel %s: %w", channel, err)
	}

	info := &UpdateInfo{
		Channel:         channel,
		CurrentVersion:  current,
		LatestVersion:   release.TagName,
		UpdateAvailable: isNewer(release.TagName, current),
		ReleaseNotes:    formatReleaseNotes(release.Body),
	}

	// Find appropriate asset - pattern: mage-x_VERSION_OS_ARCH.tar.gz
	assetPattern := fmt.Sprintf("%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	var checksumURL string

	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, assetPattern) {
			info.DownloadURL = asset.BrowserDownloadURL
		}
		// Look for checksums file (goreleaser convention: checksums.txt)
		if strings.HasSuffix(asset.Name, "checksums.txt") {
			checksumURL = asset.BrowserDownloadURL
		}
	}

	// Try to fetch and parse the checksum for our asset
	if checksumURL != "" && info.DownloadURL != "" {
		assetName := filepath.Base(info.DownloadURL)
		if checksum, err := fetchChecksumForAsset(checksumURL, assetName); err == nil {
			info.ChecksumSHA256 = checksum
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
	// Try gh CLI first if available
	if utils.CommandExists("gh") {
		if release, err := getLatestStableReleaseViaGH(owner, repo); err == nil {
			return release, nil
		}
		utils.Info("gh CLI failed, falling back to GitHub API...")
	}

	// Fallback to direct GitHub API
	return getLatestStableReleaseViaAPI(owner, repo)
}

// convertGHReleaseToGitHubRelease converts gh CLI response to GitHub API format
func convertGHReleaseToGitHubRelease(ghRelease *GHReleaseResponse) (*GitHubRelease, error) {
	// Parse the publishedAt time
	publishedAt, err := time.Parse(time.RFC3339, ghRelease.PublishedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse publishedAt time: %w", err)
	}

	// Convert assets
	assets := make([]VersionReleaseAsset, len(ghRelease.Assets))
	for i, asset := range ghRelease.Assets {
		assets[i] = VersionReleaseAsset{
			Name:               asset.Name,
			BrowserDownloadURL: asset.URL,
			Size:               asset.Size,
		}
	}

	return &GitHubRelease{
		TagName:     ghRelease.TagName,
		Name:        ghRelease.TagName, // gh CLI doesn't return name, use tagName
		Prerelease:  ghRelease.IsPrerelease,
		Draft:       ghRelease.IsDraft,
		PublishedAt: publishedAt,
		Body:        ghRelease.Body,
		HTMLURL:     ghRelease.URL,
		Assets:      assets,
	}, nil
}

// getLatestStableReleaseViaGH gets the latest stable release using gh CLI
func getLatestStableReleaseViaGH(owner, repo string) (*GitHubRelease, error) {
	repoName := fmt.Sprintf("%s/%s", owner, repo)

	// Get the latest release using gh CLI
	output, err := utils.RunCmdOutput("gh", "release", "view", "--repo", repoName, "--json", "tagName,assets,body,isPrerelease,isDraft,publishedAt,url")
	if err != nil {
		return nil, fmt.Errorf("gh CLI command failed: %w", err)
	}

	var ghRelease GHReleaseResponse
	if err := json.Unmarshal([]byte(output), &ghRelease); err != nil {
		return nil, fmt.Errorf("failed to parse gh CLI response: %w", err)
	}

	return convertGHReleaseToGitHubRelease(&ghRelease)
}

// getLatestStableReleaseViaAPI gets the latest stable release using GitHub API
func getLatestStableReleaseViaAPI(owner, repo string) (*GitHubRelease, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	return utils.HTTPGetJSON[GitHubRelease](ctx, url)
}

// getLatestBetaRelease gets the latest beta release
func getLatestBetaRelease(owner, repo string) (*GitHubRelease, error) {
	// Try gh CLI first if available
	if utils.CommandExists("gh") {
		if release, err := getLatestBetaReleaseViaGH(owner, repo); err == nil {
			return release, nil
		}
		utils.Info("gh CLI failed, falling back to GitHub API...")
	}

	// Fallback to direct GitHub API
	return getLatestBetaReleaseViaAPI(owner, repo)
}

// getLatestBetaReleaseViaGH gets the latest beta release using gh CLI.
// Beta channel prioritizes prereleases but falls back to stable if none exist.
func getLatestBetaReleaseViaGH(owner, repo string) (*GitHubRelease, error) {
	repoName := fmt.Sprintf("%s/%s", owner, repo)

	// Get all releases using gh CLI
	output, err := utils.RunCmdOutput("gh", "release", "list", "--repo", repoName, "--json", "tagName,assets,body,isPrerelease,isDraft,publishedAt,url", "--limit", "20")
	if err != nil {
		return nil, fmt.Errorf("gh CLI command failed: %w", err)
	}

	var ghReleases []GHReleaseResponse
	if err := json.Unmarshal([]byte(output), &ghReleases); err != nil {
		return nil, fmt.Errorf("failed to parse gh CLI response: %w", err)
	}

	// First pass: look for the latest prerelease (beta/alpha/rc)
	for _, ghRelease := range ghReleases {
		if !ghRelease.IsDraft && ghRelease.IsPrerelease {
			return convertGHReleaseToGitHubRelease(&ghRelease)
		}
	}

	// Fallback: if no prerelease found, return the latest stable
	for _, ghRelease := range ghReleases {
		if !ghRelease.IsDraft {
			return convertGHReleaseToGitHubRelease(&ghRelease)
		}
	}

	return nil, errNoBetaReleasesFound
}

// getLatestBetaReleaseViaAPI gets the latest beta release using GitHub API.
// Beta channel prioritizes prereleases but falls back to stable if none exist.
func getLatestBetaReleaseViaAPI(owner, repo string) (*GitHubRelease, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)
	releases, err := utils.HTTPGetJSON[[]GitHubRelease](ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases from GitHub API: %w", err)
	}

	// First pass: look for the latest prerelease (beta/alpha/rc)
	for i := range *releases {
		release := &(*releases)[i]
		if !release.Draft && release.Prerelease {
			return release, nil
		}
	}

	// Fallback: if no prerelease found, return the latest stable
	for i := range *releases {
		release := &(*releases)[i]
		if !release.Draft {
			return release, nil
		}
	}

	return nil, errNoBetaReleasesFound
}

// getLatestEdgeRelease gets the latest edge release (any release including pre-release)
func getLatestEdgeRelease(owner, repo string) (*GitHubRelease, error) {
	// Try gh CLI first if available
	if utils.CommandExists("gh") {
		if release, err := getLatestEdgeReleaseViaGH(owner, repo); err == nil {
			return release, nil
		}
		utils.Info("gh CLI failed, falling back to GitHub API...")
	}

	// Fallback to direct GitHub API
	return getLatestEdgeReleaseViaAPI(owner, repo)
}

// getLatestEdgeReleaseViaGH gets the latest edge release using gh CLI
func getLatestEdgeReleaseViaGH(owner, repo string) (*GitHubRelease, error) {
	repoName := fmt.Sprintf("%s/%s", owner, repo)

	// Get all releases using gh CLI (edge means the very latest, including prereleases)
	output, err := utils.RunCmdOutput("gh", "release", "list", "--repo", repoName, "--json", "tagName,assets,body,isPrerelease,isDraft,publishedAt,url", "--limit", "1")
	if err != nil {
		return nil, fmt.Errorf("gh CLI command failed: %w", err)
	}

	var ghReleases []GHReleaseResponse
	if err := json.Unmarshal([]byte(output), &ghReleases); err != nil {
		return nil, fmt.Errorf("failed to parse gh CLI response: %w", err)
	}

	if len(ghReleases) > 0 {
		return convertGHReleaseToGitHubRelease(&ghReleases[0])
	}

	return nil, errNoReleasesFound
}

// getLatestEdgeReleaseViaAPI gets the latest edge release using GitHub API
func getLatestEdgeReleaseViaAPI(owner, repo string) (*GitHubRelease, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases", owner, repo)
	releases, err := utils.HTTPGetJSON[[]GitHubRelease](ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch releases from GitHub API: %w", err)
	}
	if len(*releases) > 0 {
		return &(*releases)[0], nil
	}
	return nil, errNoReleasesFound
}

// fetchChecksumForAsset fetches the checksums file and extracts the checksum for the specified asset.
// Returns the hex-encoded SHA256 checksum or an error if not found.
func fetchChecksumForAsset(checksumURL, assetName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", checksumURL, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("failed to create checksum request: %w", err)
	}

	resp, err := utils.DefaultHTTPClient().Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch checksums file: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("failed to close checksum response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: status %d", errChecksumFetchFailed, resp.StatusCode)
	}

	// Limit read to 1MB (checksums file should be small)
	data, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read checksums data: %w", err)
	}

	// Parse checksums file (goreleaser format: "checksum  filename")
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == assetName {
			// Validate it looks like a hex SHA256 (64 chars)
			if len(parts[0]) == 64 {
				return parts[0], nil
			}
		}
	}

	return "", fmt.Errorf("%w: %s", errChecksumNotFound, assetName)
}

// downloadUpdate downloads the update with security verification
func downloadUpdate(info *UpdateInfo, dir string) error {
	if info.DownloadURL == "" {
		// No binary asset, use go install
		return nil
	}

	utils.Info("Downloading update...")

	// Create HTTP client with explicit timeout to prevent hanging downloads
	ctx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", info.DownloadURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create download request: %w", err)
	}

	client := &http.Client{
		Timeout: downloadTimeout,
		Transport: &http.Transport{
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 30 * time.Second,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer func() {
		// Ignore error in defer cleanup
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Best effort cleanup - ignore error
			log.Printf("failed to close response body: %v", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", errDownloadFailed, resp.StatusCode)
	}

	// Save to file
	filename := filepath.Base(info.DownloadURL)
	targetPath := filepath.Join(dir, filename)

	// Save response body to file
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read download response: %w", err)
	}

	// Verify checksum if provided (security: prevent MITM attacks)
	if info.ChecksumSHA256 != "" {
		actualHash := sha256.Sum256(data)
		actualHashHex := hex.EncodeToString(actualHash[:])

		if !strings.EqualFold(actualHashHex, info.ChecksumSHA256) {
			return fmt.Errorf("%w: expected %s, got %s",
				errChecksumMismatch, info.ChecksumSHA256, actualHashHex)
		}
		utils.Info("Checksum verified: %s", actualHashHex[:16]+"...")
	} else {
		utils.Warn("No checksum available for verification - proceeding with caution")
	}

	fileOps := fileops.New()
	return fileOps.File.WriteFile(targetPath, data, fileops.PermFile)
}

// validateExtractPath validates that a file path stays within the destination directory
// and prevents directory traversal attacks (Zip Slip vulnerability)
func validateExtractPath(destDir, tarPath string) (string, error) {
	// Reject absolute paths in tar entries (Zip Slip defense)
	if filepath.IsAbs(tarPath) {
		return "", fmt.Errorf("%w: absolute path not allowed: %s", errPathTraversal, tarPath)
	}

	// Clean the destination directory path
	destDir = filepath.Clean(destDir)

	// Join and clean the target path
	targetPath := filepath.Join(destDir, tarPath)
	targetPath = filepath.Clean(targetPath)

	// Check if the target path is within the destination directory
	relPath, err := filepath.Rel(destDir, targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to determine relative path: %w", err)
	}

	// Reject paths that escape the destination directory
	if strings.HasPrefix(relPath, "..") || strings.Contains(relPath, string(filepath.Separator)+"..") {
		return "", fmt.Errorf("%w: %s", errPathTraversal, tarPath)
	}

	return targetPath, nil
}

// extractTarGz extracts a tar.gz file to the specified directory
func extractTarGz(src, dest string) error {
	// Open the tar.gz file
	//nolint:gosec // G304: src path validated by caller
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("failed to close tar.gz file: %v", closeErr)
		}
	}()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer func() {
		if closeErr := gzipReader.Close(); closeErr != nil {
			log.Printf("failed to close gzip reader: %v", closeErr)
		}
	}()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)

	// Extract files
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Only extract regular files
		if header.Typeflag != tar.TypeReg {
			continue
		}

		// Validate and create secure destination file path
		destPath, err := validateExtractPath(dest, header.Name)
		if err != nil {
			utils.Info("Skipping malicious file path: %s (%v)", header.Name, err)
			continue
		}

		// Ensure the destination directory exists
		if dirErr := os.MkdirAll(filepath.Dir(destPath), fileops.PermDirSensitive); dirErr != nil {
			return fmt.Errorf("failed to create destination directory for %s: %w", destPath, dirErr)
		}

		// Normalize file permissions for security
		// Executables get 0o755, regular files get 0o644
		// This prevents malicious tar files from setting dangerous permissions (e.g., 0o777, setuid)
		// Mask to standard permission bits to avoid overflow when converting int64 to FileMode
		//nolint:gosec // G115: Mode is intentionally masked to valid permission bits
		mode := normalizeFileMode(os.FileMode(header.Mode & 0o7777))

		// Create the file with normalized permissions
		//nolint:gosec // G304: destPath validated by validateExtractPath function
		destFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
		if err != nil {
			return fmt.Errorf("failed to create destination file %s: %w", destPath, err)
		}

		// Copy file content with size limit to prevent zip bomb attacks
		limitedReader := io.LimitReader(tarReader, maxUpdateFileSize)
		n, copyErr := io.Copy(destFile, limitedReader)
		closeErr := destFile.Close()

		// Check if file exceeded size limit
		if n >= maxUpdateFileSize {
			// Clean up the oversized file
			if rmErr := os.Remove(destPath); rmErr != nil {
				log.Printf("failed to remove oversized file %s: %v", destPath, rmErr)
			}
			return fmt.Errorf("%w: %s exceeds %d bytes", errFileTooLarge, header.Name, maxUpdateFileSize)
		}

		if copyErr != nil {
			return fmt.Errorf("failed to extract file %s: %w", destPath, copyErr)
		}
		if closeErr != nil {
			return fmt.Errorf("failed to close destination file %s: %w", destPath, closeErr)
		}

		utils.Info("Extracted: %s", filepath.Base(destPath))
	}

	return nil
}

// normalizeFileMode normalizes file permissions for security.
// Executables get 0o755, regular files get 0o644.
// This strips dangerous bits like setuid/setgid and prevents overly permissive modes.
func normalizeFileMode(mode os.FileMode) os.FileMode {
	// Clear setuid, setgid, and sticky bits for security
	mode &^= os.ModeSetuid | os.ModeSetgid | os.ModeSticky

	// If file has any execute bit, make it 0o755, otherwise 0o644
	if mode&0o111 != 0 {
		return fileops.PermFileExecutable
	}
	return fileops.PermFile
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	//nolint:gosec // G304: src path validated by caller
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() {
		if closeErr := srcFile.Close(); closeErr != nil {
			log.Printf("failed to close source file: %v", closeErr)
		}
	}()

	//nolint:gosec // G304: dst path validated by caller
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() {
		if closeErr := dstFile.Close(); closeErr != nil {
			log.Printf("failed to close destination file: %v", closeErr)
		}
	}()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// installUpdate installs the downloaded update
func installUpdate(info *UpdateInfo, updateDir string) error {
	// Get GOPATH for installation location
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
	}
	outputPath := filepath.Join(gopath, "bin", "magex")

	// If no binary was downloaded, fall back to go install
	if info.DownloadURL == "" {
		utils.Info("No binary asset found, using go install...")
		return GetRunner().RunCmd("go", "install", fmt.Sprintf("%s@%s", magexModule, info.LatestVersion))
	}

	utils.Info("Installing downloaded binary...")

	// Find the downloaded tar.gz file
	files, err := os.ReadDir(updateDir)
	if err != nil {
		return fmt.Errorf("failed to read update directory: %w", err)
	}

	var tarGzPath string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".tar.gz") {
			tarGzPath = filepath.Join(updateDir, file.Name())
			break
		}
	}

	if tarGzPath == "" {
		return errNoTarGzFound
	}

	// Create temporary extraction directory
	extractDir := filepath.Join(updateDir, "extract")
	if mkdirErr := os.MkdirAll(extractDir, fileops.PermDirSensitive); mkdirErr != nil {
		return fmt.Errorf("failed to create extraction directory: %w", mkdirErr)
	}

	// Extract the tar.gz
	if extractErr := extractTarGz(tarGzPath, extractDir); extractErr != nil {
		return fmt.Errorf("failed to extract binary: %w", extractErr)
	}

	// Find the magex binary in extracted files
	extractedFiles, err := os.ReadDir(extractDir)
	if err != nil {
		return fmt.Errorf("failed to read extraction directory: %w", err)
	}

	var binaryPath string
	for _, file := range extractedFiles {
		if file.Name() == "magex" || (runtime.GOOS == "windows" && file.Name() == "magex.exe") {
			binaryPath = filepath.Join(extractDir, file.Name())
			break
		}
	}

	if binaryPath == "" {
		return errMagexBinaryNotFound
	}

	// Move binary to final location
	if err := os.Rename(binaryPath, outputPath); err != nil {
		// Try copy + delete if rename fails (cross-filesystem moves)
		if copyErr := copyFile(binaryPath, outputPath); copyErr != nil {
			return fmt.Errorf("failed to install binary: %w", copyErr)
		}
		if removeErr := os.Remove(binaryPath); removeErr != nil {
			log.Printf("failed to remove temporary binary: %v", removeErr)
		}
	}

	// Ensure binary is executable
	if err := os.Chmod(outputPath, fileops.PermFileExecutable); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	utils.Success("Binary installed to: %s", outputPath)
	return nil
}

// getVersionInfoForUpdate returns version specifically for update checking
// This prioritizes the binary version and always returns "dev" to force updates when needed
func getVersionInfoForUpdate() string {
	buildInfo := getBuildInfo()

	// If we have a proper version in the binary, use it
	if buildInfo.Version != versionDev {
		return buildInfo.Version
	}

	// Binary shows "dev" - provide helpful context but always return "dev" to force update
	utils.Info("Detecting current version...")
	if module, err := utils.GetModuleName(); err == nil && strings.Contains(module, "mage-x") {
		// We're in the mage-x development environment - show git context
		if tag := getCurrentGitTag(); tag != "" {
			utils.Info("Found tag on HEAD commit: %s", tag)
		}
	}

	// Always return "dev" when binary shows "dev" to ensure update happens
	// This forces the comparison "dev" < "v1.x.x" = true, triggering update
	return versionDev
}
