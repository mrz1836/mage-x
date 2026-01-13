package mage

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Static error for mock fetcher (satisfies err113 linter)
var errMockNetwork = errors.New("network error")

// MockReleaseFetcher is a mock implementation of GitHubReleaseFetcher
type MockReleaseFetcher struct {
	Release *GitHubRelease
	Error   error
	Called  bool
}

func (m *MockReleaseFetcher) FetchLatestRelease(_ context.Context) (*GitHubRelease, error) {
	m.Called = true
	if m.Error != nil {
		return nil, m.Error
	}
	return m.Release, nil
}

// UpdateNotifyTestSuite tests the update notification functionality
type UpdateNotifyTestSuite struct {
	suite.Suite

	tempDir string
	cache   *UpdateNotifyCache
}

func (suite *UpdateNotifyTestSuite) SetupTest() {
	var err error
	suite.tempDir, err = os.MkdirTemp("", "update-notify-test-*")
	suite.Require().NoError(err)

	suite.cache = NewUpdateNotifyCacheWithOptions(suite.tempDir, 24*time.Hour)
}

func (suite *UpdateNotifyTestSuite) TearDownTest() {
	if suite.tempDir != "" {
		_ = os.RemoveAll(suite.tempDir) //nolint:errcheck // Best effort cleanup in test
	}
}

func TestUpdateNotifySuite(t *testing.T) {
	suite.Run(t, new(UpdateNotifyTestSuite))
}

// TestNewUpdateNotifier tests creating a new notifier
func (suite *UpdateNotifyTestSuite) TestNewUpdateNotifier() {
	notifier := NewUpdateNotifier()
	suite.NotNil(notifier)
	suite.NotNil(notifier.cache)
	suite.NotEmpty(notifier.currentVersion)
	suite.Equal(updateCheckTimeout, notifier.timeout)
}

// TestNewUpdateNotifierWithOptions tests creating a notifier with options
func (suite *UpdateNotifyTestSuite) TestNewUpdateNotifierWithOptions() {
	mockFetcher := &MockReleaseFetcher{}
	customTimeout := 10 * time.Second

	notifier := NewUpdateNotifier(
		WithTimeout(customTimeout),
		WithChannel(BetaChannel),
		WithCache(suite.cache),
		WithFetcher(mockFetcher),
		WithCurrentVersion("v1.0.0"),
	)

	suite.Equal(customTimeout, notifier.timeout)
	suite.Equal(BetaChannel, notifier.channel)
	suite.Equal(suite.cache, notifier.cache)
	suite.Equal(mockFetcher, notifier.fetcher)
	suite.Equal("v1.0.0", notifier.currentVersion)
}

// TestCheckWithCacheHit tests that cached results are returned
func (suite *UpdateNotifyTestSuite) TestCheckWithCacheHit() {
	// Pre-populate cache
	cacheData := &UpdateCacheData{
		CurrentVersion:  "v1.0.0",
		LatestVersion:   "v1.1.0",
		UpdateAvailable: true,
		ReleaseNotes:    "Cached release notes",
		ReleaseURL:      "https://github.com/test/cached",
		LastCheck:       time.Now(),
	}
	err := suite.cache.Store(cacheData)
	suite.Require().NoError(err)

	mockFetcher := &MockReleaseFetcher{
		Release: &GitHubRelease{TagName: "v2.0.0"},
	}

	notifier := NewUpdateNotifier(
		WithCache(suite.cache),
		WithFetcher(mockFetcher),
		WithCurrentVersion("v1.0.0"),
	)

	result, err := notifier.Check(context.Background())
	suite.Require().NoError(err)
	suite.NotNil(result)

	// Should return cached data, not fetch new
	suite.False(mockFetcher.Called, "Fetcher should not be called when cache is valid")
	suite.True(result.FromCache)
	suite.Equal("v1.1.0", result.LatestVersion) // From cache, not fetcher
}

// TestCheckWithCacheMiss tests that API is called when cache is empty
func (suite *UpdateNotifyTestSuite) TestCheckWithCacheMiss() {
	mockFetcher := &MockReleaseFetcher{
		Release: &GitHubRelease{
			TagName: "v1.2.0",
			Body:    "New release notes",
			HTMLURL: "https://github.com/test/release",
		},
	}

	notifier := NewUpdateNotifier(
		WithCache(suite.cache),
		WithFetcher(mockFetcher),
		WithCurrentVersion("v1.0.0"),
	)

	result, err := notifier.Check(context.Background())
	suite.Require().NoError(err)
	suite.NotNil(result)

	suite.True(mockFetcher.Called, "Fetcher should be called when cache is empty")
	suite.False(result.FromCache)
	suite.Equal("v1.2.0", result.LatestVersion)
	suite.True(result.UpdateAvailable)
}

// TestCheckUpdateAvailable tests version comparison for update availability
func (suite *UpdateNotifyTestSuite) TestCheckUpdateAvailable() {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		expectedAvail  bool
	}{
		{
			name:           "newer version available",
			currentVersion: "v1.0.0",
			latestVersion:  "v1.1.0",
			expectedAvail:  true,
		},
		{
			name:           "same version",
			currentVersion: "v1.0.0",
			latestVersion:  "v1.0.0",
			expectedAvail:  false,
		},
		{
			name:           "older version on server",
			currentVersion: "v2.0.0",
			latestVersion:  "v1.0.0",
			expectedAvail:  false,
		},
		{
			name:           "dev version always shows update",
			currentVersion: "dev",
			latestVersion:  "v1.0.0",
			expectedAvail:  true,
		},
		{
			name:           "major version bump",
			currentVersion: "v1.9.9",
			latestVersion:  "v2.0.0",
			expectedAvail:  true,
		},
		{
			name:           "minor version bump",
			currentVersion: "v1.0.9",
			latestVersion:  "v1.1.0",
			expectedAvail:  true,
		},
		{
			name:           "patch version bump",
			currentVersion: "v1.0.0",
			latestVersion:  "v1.0.1",
			expectedAvail:  true,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			mockFetcher := &MockReleaseFetcher{
				Release: &GitHubRelease{
					TagName: tt.latestVersion,
				},
			}

			notifier := NewUpdateNotifier(
				WithCache(NewUpdateNotifyCacheWithOptions(suite.tempDir, 0)), // No caching
				WithFetcher(mockFetcher),
				WithCurrentVersion(tt.currentVersion),
			)

			result, err := notifier.Check(context.Background())
			suite.Require().NoError(err)
			suite.Equal(tt.expectedAvail, result.UpdateAvailable,
				"Expected UpdateAvailable=%v for current=%s, latest=%s",
				tt.expectedAvail, tt.currentVersion, tt.latestVersion)
		})
	}
}

// TestCheckWithFetchError tests error handling when fetch fails
func (suite *UpdateNotifyTestSuite) TestCheckWithFetchError() {
	mockFetcher := &MockReleaseFetcher{
		Error: errMockNetwork,
	}

	notifier := NewUpdateNotifier(
		WithCache(suite.cache),
		WithFetcher(mockFetcher),
		WithCurrentVersion("v1.0.0"),
	)

	result, err := notifier.Check(context.Background())
	suite.Require().Error(err)
	suite.NotNil(result) // Result is still returned with error info
	suite.NotEmpty(result.Error)
}

// TestCheckWithTimeout tests context timeout handling
func (suite *UpdateNotifyTestSuite) TestCheckWithTimeout() {
	// Create a fetcher that blocks
	blockingFetcher := &slowFetcher{delay: 5 * time.Second}

	notifier := NewUpdateNotifier(
		WithCache(suite.cache),
		WithFetcher(blockingFetcher),
		WithCurrentVersion("v1.0.0"),
		WithTimeout(100*time.Millisecond),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	_, err := notifier.Check(ctx)
	suite.Require().Error(err) // Should timeout
}

// slowFetcher simulates a slow network response
type slowFetcher struct {
	delay time.Duration
}

func (f *slowFetcher) FetchLatestRelease(ctx context.Context) (*GitHubRelease, error) {
	select {
	case <-time.After(f.delay):
		return &GitHubRelease{TagName: "v1.0.0"}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// TestStartBackgroundUpdateCheck tests the async check
func (suite *UpdateNotifyTestSuite) TestStartBackgroundUpdateCheck() {
	// Clear CI env vars that would disable the check
	oldCI := os.Getenv("CI")
	oldGH := os.Getenv("GITHUB_ACTIONS")
	oldDisable := os.Getenv("MAGEX_DISABLE_UPDATE_CHECK")
	_ = os.Unsetenv("CI")                         //nolint:errcheck // Test env setup
	_ = os.Unsetenv("GITHUB_ACTIONS")             //nolint:errcheck // Test env setup
	_ = os.Unsetenv("MAGEX_DISABLE_UPDATE_CHECK") //nolint:errcheck // Test env setup
	defer func() {
		if oldCI != "" {
			_ = os.Setenv("CI", oldCI) //nolint:errcheck // Test env restore
		}
		if oldGH != "" {
			_ = os.Setenv("GITHUB_ACTIONS", oldGH) //nolint:errcheck // Test env restore
		}
		if oldDisable != "" {
			_ = os.Setenv("MAGEX_DISABLE_UPDATE_CHECK", oldDisable) //nolint:errcheck // Test env restore
		}
	}()

	ctx := context.Background()
	resultChan := StartBackgroundUpdateCheck(ctx)

	// Should receive a result or the channel should close
	select {
	case result := <-resultChan:
		// Result can be nil if the check failed (e.g., no network)
		// or can contain actual data
		_ = result // Just verify we got something back
	case <-time.After(10 * time.Second):
		suite.Fail("Background check timed out")
	}
}

// TestStartBackgroundUpdateCheckDisabled tests that check is skipped when disabled
func (suite *UpdateNotifyTestSuite) TestStartBackgroundUpdateCheckDisabled() {
	// Set disable env var
	_ = os.Setenv("MAGEX_DISABLE_UPDATE_CHECK", "true") //nolint:errcheck // Test env setup
	defer func() {
		_ = os.Unsetenv("MAGEX_DISABLE_UPDATE_CHECK") //nolint:errcheck // Test env cleanup
	}()

	ctx := context.Background()
	resultChan := StartBackgroundUpdateCheck(ctx)

	// Channel should close without sending a result
	select {
	case result := <-resultChan:
		suite.Nil(result, "Should not receive a result when disabled")
	case <-time.After(1 * time.Second):
		suite.Fail("Channel should close immediately when disabled")
	}
}

// TestStartBackgroundUpdateCheckCI tests that check is skipped in CI
func (suite *UpdateNotifyTestSuite) TestStartBackgroundUpdateCheckCI() {
	// Set CI env var
	_ = os.Setenv("CI", "true") //nolint:errcheck // Test env setup
	defer func() {
		_ = os.Unsetenv("CI") //nolint:errcheck // Test env cleanup
	}()

	ctx := context.Background()
	resultChan := StartBackgroundUpdateCheck(ctx)

	// Channel should close without sending a result
	select {
	case result := <-resultChan:
		suite.Nil(result, "Should not receive a result in CI")
	case <-time.After(1 * time.Second):
		suite.Fail("Channel should close immediately in CI")
	}
}

// TestGetGitHubTokenForUpdateCheck tests token discovery
func TestGetGitHubTokenForUpdateCheck(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected string
	}{
		{
			name:     "no token set",
			envVars:  map[string]string{},
			expected: "",
		},
		{
			name: "MAGE_X_GITHUB_TOKEN set",
			envVars: map[string]string{
				"MAGE_X_GITHUB_TOKEN": "mage_x_token",
			},
			expected: "mage_x_token",
		},
		{
			name: "GITHUB_TOKEN set",
			envVars: map[string]string{
				"GITHUB_TOKEN": "github_token",
			},
			expected: "github_token",
		},
		{
			name: "GH_TOKEN set",
			envVars: map[string]string{
				"GH_TOKEN": "gh_token",
			},
			expected: "gh_token",
		},
		{
			name: "MAGE_X_GITHUB_TOKEN takes priority",
			envVars: map[string]string{
				"MAGE_X_GITHUB_TOKEN": "mage_x_token",
				"GITHUB_TOKEN":        "github_token",
				"GH_TOKEN":            "gh_token",
			},
			expected: "mage_x_token",
		},
		{
			name: "GITHUB_TOKEN over GH_TOKEN",
			envVars: map[string]string{
				"GITHUB_TOKEN": "github_token",
				"GH_TOKEN":     "gh_token",
			},
			expected: "github_token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore env vars
			savedEnv := map[string]string{}
			for key := range tt.envVars {
				savedEnv[key] = os.Getenv(key)
			}
			// Also save all token vars
			for _, key := range []string{"MAGE_X_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"} {
				if _, exists := savedEnv[key]; !exists {
					savedEnv[key] = os.Getenv(key)
				}
			}

			defer func() {
				for key, val := range savedEnv {
					if val != "" {
						_ = os.Setenv(key, val) //nolint:errcheck // Test env restore
					} else {
						_ = os.Unsetenv(key) //nolint:errcheck // Test env restore
					}
				}
			}()

			// Clear all token env vars first
			_ = os.Unsetenv("MAGE_X_GITHUB_TOKEN") //nolint:errcheck // Test env setup
			_ = os.Unsetenv("GITHUB_TOKEN")        //nolint:errcheck // Test env setup
			_ = os.Unsetenv("GH_TOKEN")            //nolint:errcheck // Test env setup

			// Set test env vars
			for key, val := range tt.envVars {
				_ = os.Setenv(key, val) //nolint:errcheck // Test env setup
			}

			result := getGitHubTokenForUpdateCheck()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestUpdateCheckResult tests the UpdateCheckResult struct
func TestUpdateCheckResult(t *testing.T) {
	result := &UpdateCheckResult{
		CurrentVersion:  "v1.0.0",
		LatestVersion:   "v1.1.0",
		UpdateAvailable: true,
		ReleaseNotes:    "Test notes",
		ReleaseURL:      "https://example.com",
		CheckedAt:       time.Now(),
		FromCache:       false,
	}

	assert.Equal(t, "v1.0.0", result.CurrentVersion)
	assert.Equal(t, "v1.1.0", result.LatestVersion)
	assert.True(t, result.UpdateAvailable)
	assert.Equal(t, "Test notes", result.ReleaseNotes)
	assert.Equal(t, "https://example.com", result.ReleaseURL)
	assert.False(t, result.FromCache)
}

// TestDefaultReleaseFetcher tests the default fetcher implementation
func TestDefaultReleaseFetcher(t *testing.T) {
	fetcher := &DefaultReleaseFetcher{}

	// This will make a real API call, so we use a short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// We don't assert on the result because it depends on network
	// Just verify it doesn't panic
	//nolint:errcheck // We intentionally ignore the result in this test
	_, _ = fetcher.FetchLatestRelease(ctx)
}

// TestUpdateNotifierCachesResult tests that results are cached after check
func (suite *UpdateNotifyTestSuite) TestUpdateNotifierCachesResult() {
	mockFetcher := &MockReleaseFetcher{
		Release: &GitHubRelease{
			TagName: "v1.2.0",
			Body:    "Release notes",
			HTMLURL: "https://github.com/test/release",
		},
	}

	notifier := NewUpdateNotifier(
		WithCache(suite.cache),
		WithFetcher(mockFetcher),
		WithCurrentVersion("v1.0.0"),
	)

	// First check should call fetcher
	_, err := notifier.Check(context.Background())
	suite.Require().NoError(err)
	suite.True(mockFetcher.Called)

	// Reset called flag
	mockFetcher.Called = false

	// Second check should use cache
	result, err := notifier.Check(context.Background())
	suite.Require().NoError(err)
	suite.False(mockFetcher.Called, "Fetcher should not be called on second check")
	suite.True(result.FromCache)
}

// TestGetBuildInfoVersion tests that getBuildInfoVersion returns the binary version
func TestGetBuildInfoVersion(t *testing.T) {
	// getBuildInfoVersion should return the binary's embedded version
	// In tests, this will be "dev" since we don't set ldflags
	version := getBuildInfoVersion()

	// Should return a non-empty string
	assert.NotEmpty(t, version)

	// In test environment without ldflags, it should be "dev"
	// This is the expected behavior - dev builds get "dev" version
	assert.Equal(t, "dev", version)
}

// TestNewUpdateNotifierUsesBinaryVersion tests that NewUpdateNotifier uses the binary version
func TestNewUpdateNotifierUsesBinaryVersion(t *testing.T) {
	// Create a notifier without WithCurrentVersion option
	// It should use the binary's embedded version
	notifier := NewUpdateNotifier()

	// The current version should match getBuildInfoVersion()
	assert.Equal(t, getBuildInfoVersion(), notifier.currentVersion)

	// In test environment, this should be "dev"
	assert.Equal(t, "dev", notifier.currentVersion)
}

// TestDevVersionShowsUpdateAvailable tests that "dev" versions always show update available
func TestDevVersionShowsUpdateAvailable(t *testing.T) {
	// This test verifies the core fix: dev builds should always see updates
	mockFetcher := &MockReleaseFetcher{
		Release: &GitHubRelease{
			TagName: "v1.16.1",
			Body:    "Release notes",
			HTMLURL: "https://github.com/test/release",
		},
	}

	// Create temp cache directory
	tempDir, err := os.MkdirTemp("", "update-test-*")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir) //nolint:errcheck // Best effort cleanup
	}()

	cache := NewUpdateNotifyCacheWithOptions(tempDir, 0) // No caching

	// Create notifier with "dev" version (simulating a dev build)
	notifier := NewUpdateNotifier(
		WithCache(cache),
		WithFetcher(mockFetcher),
		WithCurrentVersion("dev"),
	)

	result, err := notifier.Check(context.Background())
	require.NoError(t, err)

	// Dev version should always show update available
	assert.True(t, result.UpdateAvailable, "Dev builds should always show update available")
	assert.Equal(t, "dev", result.CurrentVersion)
	assert.Equal(t, "v1.16.1", result.LatestVersion)
}

// TestCheckWithEmptyTagName tests handling of empty tag name from GitHub
func (suite *UpdateNotifyTestSuite) TestCheckWithEmptyTagName() {
	mockFetcher := &MockReleaseFetcher{
		Release: &GitHubRelease{
			TagName: "", // Empty tag name
			Body:    "Release notes",
			HTMLURL: "https://github.com/test/release",
		},
	}

	notifier := NewUpdateNotifier(
		WithCache(suite.cache),
		WithFetcher(mockFetcher),
		WithCurrentVersion("v1.0.0"),
	)

	result, err := notifier.Check(context.Background())

	// Should return error for empty tag name
	suite.Require().Error(err)
	suite.Equal(ErrEmptyTagName, err)
	suite.NotNil(result)
	suite.Contains(result.Error, "empty tag name")
}

// TestCheckWithWhitespaceTagName tests handling of whitespace-only tag name
func (suite *UpdateNotifyTestSuite) TestCheckWithWhitespaceTagName() {
	mockFetcher := &MockReleaseFetcher{
		Release: &GitHubRelease{
			TagName: "   \t\n", // Whitespace only
			Body:    "Release notes",
			HTMLURL: "https://github.com/test/release",
		},
	}

	notifier := NewUpdateNotifier(
		WithCache(suite.cache),
		WithFetcher(mockFetcher),
		WithCurrentVersion("v1.0.0"),
	)

	result, err := notifier.Check(context.Background())

	// Should return error for whitespace-only tag name
	suite.Require().Error(err)
	suite.Equal(ErrEmptyTagName, err)
	suite.NotNil(result)
}

// TestCheckWithNilRelease tests handling of nil release from fetcher
func (suite *UpdateNotifyTestSuite) TestCheckWithNilRelease() {
	mockFetcher := &MockReleaseFetcher{
		Release: nil, // Nil release
		Error:   nil, // No error
	}

	notifier := NewUpdateNotifier(
		WithCache(suite.cache),
		WithFetcher(mockFetcher),
		WithCurrentVersion("v1.0.0"),
	)

	result, err := notifier.Check(context.Background())

	// Should return error for nil release
	suite.Require().Error(err)
	suite.Equal(ErrEmptyTagName, err)
	suite.NotNil(result)
}

// errRateLimit is a static error for testing rate limit scenarios
var errRateLimit = errors.New("HTTP API error: GET https://api.github.com/repos/mrz1836/mage-x/releases/latest returned status 429: rate limit exceeded")

// rateLimitFetcher simulates a rate limit error
type rateLimitFetcher struct{}

func (f *rateLimitFetcher) FetchLatestRelease(_ context.Context) (*GitHubRelease, error) {
	return nil, errRateLimit
}

// TestCheckWithRateLimitError tests handling of rate limit errors
func (suite *UpdateNotifyTestSuite) TestCheckWithRateLimitError() {
	notifier := NewUpdateNotifier(
		WithCache(suite.cache),
		WithFetcher(&rateLimitFetcher{}),
		WithCurrentVersion("v1.0.0"),
	)

	result, err := notifier.Check(context.Background())

	// Should return rate limit error
	suite.Require().Error(err)
	suite.Equal(ErrRateLimited, err)
	suite.NotNil(result)
	suite.Contains(result.Error, "rate limit")
}

// TestCheckContextCancellation tests context cancellation during check
func (suite *UpdateNotifyTestSuite) TestCheckContextCancellation() {
	// Create a fetcher that blocks until context is canceled
	blockingFetcher := &slowFetcher{delay: 10 * time.Second}

	notifier := NewUpdateNotifier(
		WithCache(suite.cache),
		WithFetcher(blockingFetcher),
		WithCurrentVersion("v1.0.0"),
		WithTimeout(50*time.Millisecond), // Short timeout
	)

	// Create a context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	// Check should fail due to canceled context
	_, err := notifier.Check(ctx)
	suite.Require().Error(err)
}

// TestUpdateErrorConstants tests that update error constants are properly defined
func TestUpdateErrorConstants(t *testing.T) {
	// Verify error constants exist and have meaningful messages
	// Using NotNil instead of Error because we're testing the constants exist, not error flow
	assert.NotNil(t, ErrEmptyTagName, "ErrEmptyTagName should be defined") //nolint:testifylint // testing constant existence, not error flow
	assert.NotNil(t, ErrRateLimited, "ErrRateLimited should be defined")   //nolint:testifylint // testing constant existence, not error flow

	assert.Contains(t, ErrEmptyTagName.Error(), "empty")
	assert.Contains(t, ErrRateLimited.Error(), "rate limit")
}
