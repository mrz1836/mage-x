// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/mrz1836/mage-x/pkg/common/env"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Error constants for update checking
var (
	// ErrEmptyTagName is returned when the GitHub API returns a release with no tag name
	ErrEmptyTagName = errors.New("GitHub release has empty tag name")

	// ErrRateLimited is returned when GitHub API rate limit is exceeded
	ErrRateLimited = errors.New("GitHub API rate limit exceeded, try again later")
)

// Update notification constants
const (
	// updateCheckTimeout is the maximum time for an update check API call
	updateCheckTimeout = 5 * time.Second

	// magexGitHubOwner is the GitHub owner for mage-x releases
	magexGitHubOwner = "mrz1836"

	// magexGitHubRepo is the GitHub repository for mage-x releases
	magexGitHubRepo = "mage-x"
)

// UpdateCheckResult contains the result of an update check
type UpdateCheckResult struct {
	CurrentVersion  string    `json:"current_version"`
	LatestVersion   string    `json:"latest_version"`
	UpdateAvailable bool      `json:"update_available"`
	ReleaseNotes    string    `json:"release_notes,omitempty"`
	ReleaseURL      string    `json:"release_url,omitempty"`
	CheckedAt       time.Time `json:"checked_at"`
	FromCache       bool      `json:"from_cache"`
	Error           string    `json:"error,omitempty"`
}

// UpdateChecker interface for testing
type UpdateChecker interface {
	Check(ctx context.Context) (*UpdateCheckResult, error)
}

// GitHubReleaseFetcher interface for testing
type GitHubReleaseFetcher interface {
	FetchLatestRelease(ctx context.Context) (*GitHubRelease, error)
}

// UpdateNotifier handles update checking and notification
type UpdateNotifier struct {
	cache          *UpdateNotifyCache
	currentVersion string
	channel        UpdateChannel
	timeout        time.Duration
	fetcher        GitHubReleaseFetcher
}

// UpdateNotifierOption configures the UpdateNotifier
type UpdateNotifierOption func(*UpdateNotifier)

// WithTimeout sets the HTTP timeout for update checks
func WithTimeout(d time.Duration) UpdateNotifierOption {
	return func(n *UpdateNotifier) {
		n.timeout = d
	}
}

// WithChannel sets the release channel
func WithChannel(ch UpdateChannel) UpdateNotifierOption {
	return func(n *UpdateNotifier) {
		n.channel = ch
	}
}

// WithCache sets a custom cache (for testing)
func WithCache(cache *UpdateNotifyCache) UpdateNotifierOption {
	return func(n *UpdateNotifier) {
		n.cache = cache
	}
}

// WithFetcher sets a custom fetcher (for testing)
func WithFetcher(fetcher GitHubReleaseFetcher) UpdateNotifierOption {
	return func(n *UpdateNotifier) {
		n.fetcher = fetcher
	}
}

// WithCurrentVersion sets a custom current version (for testing)
func WithCurrentVersion(version string) UpdateNotifierOption {
	return func(n *UpdateNotifier) {
		n.currentVersion = version
	}
}

// registeredBinaryVersion holds the version registered by main at startup
// This allows the binary's actual embedded version to be used for update checks
// without requiring ldflags to be set for the pkg/mage package
var registeredBinaryVersion string //nolint:gochecknoglobals // Required for version registration from main

// RegisterBinaryVersion registers the binary's version with the mage package.
// This should be called by main at startup to ensure the update checker
// uses the correct version that was embedded via ldflags in main.
// If not called, getBuildInfoVersion() falls back to pkg/mage's default "dev".
func RegisterBinaryVersion(version string) {
	registeredBinaryVersion = version
}

// getBuildInfoVersion returns the version embedded in the binary
// This is used for update checking to ensure "dev" builds always see updates
// It prioritizes the registered version from main over the pkg default
func getBuildInfoVersion() string {
	// Use registered version from main if available
	if registeredBinaryVersion != "" {
		return registeredBinaryVersion
	}
	// Fall back to pkg/mage's BuildInfo (will be "dev" in dev builds)
	return getBuildInfo().Version
}

// NewUpdateNotifier creates a new update notifier
func NewUpdateNotifier(opts ...UpdateNotifierOption) *UpdateNotifier {
	n := &UpdateNotifier{
		cache:          NewUpdateNotifyCache(),
		currentVersion: getBuildInfoVersion(),
		channel:        getUpdateChannel(),
		timeout:        updateCheckTimeout,
	}

	for _, opt := range opts {
		opt(n)
	}

	return n
}

// Check performs an update check, using cache if available
func (n *UpdateNotifier) Check(ctx context.Context) (*UpdateCheckResult, error) {
	// Check cache first
	if cached, valid := n.cache.Get(); valid {
		return &UpdateCheckResult{
			CurrentVersion:  n.currentVersion,
			LatestVersion:   cached.LatestVersion,
			UpdateAvailable: cached.UpdateAvailable,
			ReleaseNotes:    cached.ReleaseNotes,
			ReleaseURL:      cached.ReleaseURL,
			CheckedAt:       cached.LastCheck,
			FromCache:       true,
		}, nil
	}

	// Perform actual check
	result, err := n.checkForUpdate(ctx)
	if err != nil {
		// Return the result even on error (it contains error info)
		return result, err
	}

	// Store in cache (best effort, ignore errors)
	_ = n.cache.Store(&UpdateCacheData{ //nolint:errcheck // Best effort cache storage
		CurrentVersion:  result.CurrentVersion,
		LatestVersion:   result.LatestVersion,
		UpdateAvailable: result.UpdateAvailable,
		ReleaseNotes:    result.ReleaseNotes,
		ReleaseURL:      result.ReleaseURL,
	})

	return result, nil
}

// checkForUpdate performs the actual update check against GitHub
func (n *UpdateNotifier) checkForUpdate(ctx context.Context) (*UpdateCheckResult, error) {
	// Create timeout context
	checkCtx, cancel := context.WithTimeout(ctx, n.timeout)
	defer cancel()

	// Fetch latest release
	var release *GitHubRelease
	var err error

	if n.fetcher != nil {
		release, err = n.fetcher.FetchLatestRelease(checkCtx)
	} else {
		release, err = fetchLatestReleaseWithAuth(checkCtx)
	}

	if err != nil {
		// Check for rate limiting (429) in error message
		errStr := err.Error()
		if strings.Contains(errStr, "429") || strings.Contains(errStr, "rate limit") {
			return &UpdateCheckResult{
				CurrentVersion: n.currentVersion,
				CheckedAt:      time.Now(),
				Error:          ErrRateLimited.Error(),
			}, ErrRateLimited
		}
		return &UpdateCheckResult{
			CurrentVersion: n.currentVersion,
			CheckedAt:      time.Now(),
			Error:          err.Error(),
		}, err
	}

	// Validate the response has required data
	if release == nil || strings.TrimSpace(release.TagName) == "" {
		return &UpdateCheckResult{
			CurrentVersion: n.currentVersion,
			CheckedAt:      time.Now(),
			Error:          ErrEmptyTagName.Error(),
		}, ErrEmptyTagName
	}

	result := &UpdateCheckResult{
		CurrentVersion:  n.currentVersion,
		LatestVersion:   release.TagName,
		UpdateAvailable: isNewer(release.TagName, n.currentVersion),
		ReleaseNotes:    formatReleaseNotes(release.Body),
		ReleaseURL:      release.HTMLURL,
		CheckedAt:       time.Now(),
		FromCache:       false,
	}

	return result, nil
}

// StartBackgroundUpdateCheck starts an asynchronous update check
// Returns a channel that receives the result when complete
// The check is non-blocking and runs in a goroutine
func StartBackgroundUpdateCheck(ctx context.Context) <-chan *UpdateCheckResult {
	resultChan := make(chan *UpdateCheckResult, 1)

	go func() {
		defer close(resultChan)

		// Recover from any panics to prevent crashing the CLI
		defer func() {
			if r := recover(); r != nil {
				utils.Debug("Update check recovered from panic: %v", r)
			}
		}()

		// Skip if update checking is disabled
		if isUpdateCheckDisabled() {
			return
		}

		notifier := NewUpdateNotifier()
		result, err := notifier.Check(ctx)
		if err != nil {
			utils.Debug("Update check failed: %v", err)
			return
		}

		resultChan <- result
	}()

	return resultChan
}

// fetchLatestReleaseWithAuth fetches the latest release with authentication fallback
func fetchLatestReleaseWithAuth(ctx context.Context) (*GitHubRelease, error) {
	token := getGitHubTokenForUpdateCheck()
	url := "https://api.github.com/repos/" + magexGitHubOwner + "/" + magexGitHubRepo + "/releases/latest"

	// Try authenticated request first (higher rate limit: 5000/hr vs 60/hr)
	if token != "" {
		if release, err := utils.HTTPGetJSONWithAuth[GitHubRelease](ctx, url, token); err == nil {
			return release, nil
		}
		// Fall through to public API if authenticated request fails
		utils.Debug("Authenticated GitHub request failed, falling back to public API")
	}

	// Fallback to public API (60 requests/hour limit)
	return utils.HTTPGetJSON[GitHubRelease](ctx, url)
}

// getGitHubTokenForUpdateCheck returns the GitHub token for update checks
// Priority: MAGE_X_GITHUB_TOKEN > GITHUB_TOKEN > GH_TOKEN
func getGitHubTokenForUpdateCheck() string {
	if token := env.Get("MAGE_X_GITHUB_TOKEN"); token != "" {
		return token
	}
	if token := env.Get("GITHUB_TOKEN"); token != "" {
		return token
	}
	return env.Get("GH_TOKEN")
}

// formatReleaseNotes formats release notes for display
// This is already defined in version.go, so we use it from there
// Keeping this comment for documentation purposes

// DefaultReleaseFetcher implements GitHubReleaseFetcher using the GitHub API
type DefaultReleaseFetcher struct{}

// FetchLatestRelease fetches the latest release from GitHub
func (f *DefaultReleaseFetcher) FetchLatestRelease(ctx context.Context) (*GitHubRelease, error) {
	return fetchLatestReleaseWithAuth(ctx)
}
