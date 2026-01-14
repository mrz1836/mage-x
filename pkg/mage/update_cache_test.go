package mage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// UpdateCacheTestSuite tests the update cache functionality
type UpdateCacheTestSuite struct {
	suite.Suite

	tempDir string
	cache   *UpdateNotifyCache
}

func (suite *UpdateCacheTestSuite) SetupTest() {
	var err error
	suite.tempDir, err = os.MkdirTemp("", "update-cache-test-*")
	suite.Require().NoError(err)

	suite.cache = NewUpdateNotifyCacheWithOptions(suite.tempDir, 24*time.Hour)
}

func (suite *UpdateCacheTestSuite) TearDownTest() {
	if suite.tempDir != "" {
		_ = os.RemoveAll(suite.tempDir) //nolint:errcheck // Best effort cleanup in test
	}
}

func TestUpdateCacheSuite(t *testing.T) {
	suite.Run(t, new(UpdateCacheTestSuite))
}

// TestNewUpdateNotifyCache tests cache creation
func (suite *UpdateCacheTestSuite) TestNewUpdateNotifyCache() {
	cache := NewUpdateNotifyCache()
	suite.NotNil(cache)
	suite.NotEmpty(cache.GetCacheDir())
	suite.NotEmpty(cache.GetCacheFile())
	suite.Equal(getUpdateCheckInterval(), cache.GetTTL())
}

// TestNewUpdateNotifyCacheWithOptions tests cache creation with custom options
func (suite *UpdateCacheTestSuite) TestNewUpdateNotifyCacheWithOptions() {
	customTTL := 12 * time.Hour
	cache := NewUpdateNotifyCacheWithOptions(suite.tempDir, customTTL)

	suite.Equal(suite.tempDir, cache.GetCacheDir())
	suite.Equal(filepath.Join(suite.tempDir, "update-check.json"), cache.GetCacheFile())
	suite.Equal(customTTL, cache.GetTTL())
}

// TestStoreAndGet tests storing and retrieving cache data
func (suite *UpdateCacheTestSuite) TestStoreAndGet() {
	data := &UpdateCacheData{
		CurrentVersion:  "v1.0.0",
		LatestVersion:   "v1.1.0",
		UpdateAvailable: true,
		ReleaseNotes:    "Test release notes",
		ReleaseURL:      "https://github.com/example/repo/releases/v1.1.0",
	}

	// Store data
	err := suite.cache.Store(data)
	suite.Require().NoError(err)

	// Retrieve data
	retrieved, valid := suite.cache.Get()
	suite.True(valid)
	suite.NotNil(retrieved)
	suite.Equal(data.CurrentVersion, retrieved.CurrentVersion)
	suite.Equal(data.LatestVersion, retrieved.LatestVersion)
	suite.Equal(data.UpdateAvailable, retrieved.UpdateAvailable)
	suite.Equal(data.ReleaseNotes, retrieved.ReleaseNotes)
	suite.Equal(data.ReleaseURL, retrieved.ReleaseURL)
	suite.NotZero(retrieved.LastCheck)
	suite.NotEmpty(retrieved.CheckInterval)
}

// TestGetNonExistentCache tests getting data when no cache exists
func (suite *UpdateCacheTestSuite) TestGetNonExistentCache() {
	retrieved, valid := suite.cache.Get()
	suite.False(valid)
	suite.Nil(retrieved)
}

// TestGetExpiredCache tests that expired cache is not returned
func (suite *UpdateCacheTestSuite) TestGetExpiredCache() {
	// Create cache with very short TTL
	shortCache := NewUpdateNotifyCacheWithOptions(suite.tempDir, 1*time.Millisecond)

	data := &UpdateCacheData{
		CurrentVersion:  "v1.0.0",
		LatestVersion:   "v1.1.0",
		UpdateAvailable: true,
	}

	err := shortCache.Store(data)
	suite.Require().NoError(err)

	// Wait for cache to expire
	time.Sleep(10 * time.Millisecond)

	// Cache should be expired
	retrieved, valid := shortCache.Get()
	suite.False(valid)
	suite.Nil(retrieved)
}

// TestIsExpired tests cache expiration checking
func (suite *UpdateCacheTestSuite) TestIsExpired() {
	// No cache file - should be expired
	suite.True(suite.cache.IsExpired())

	// Store data
	err := suite.cache.Store(&UpdateCacheData{
		CurrentVersion: "v1.0.0",
		LatestVersion:  "v1.0.0",
	})
	suite.Require().NoError(err)

	// Should not be expired immediately
	suite.False(suite.cache.IsExpired())
}

// TestIsExpiredWithShortTTL tests expiration with short TTL
func (suite *UpdateCacheTestSuite) TestIsExpiredWithShortTTL() {
	shortCache := NewUpdateNotifyCacheWithOptions(suite.tempDir, 1*time.Millisecond)

	err := shortCache.Store(&UpdateCacheData{
		CurrentVersion: "v1.0.0",
		LatestVersion:  "v1.0.0",
	})
	suite.Require().NoError(err)

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	suite.True(shortCache.IsExpired())
}

// TestClear tests cache clearing
func (suite *UpdateCacheTestSuite) TestClear() {
	// Store data first
	err := suite.cache.Store(&UpdateCacheData{
		CurrentVersion: "v1.0.0",
		LatestVersion:  "v1.0.0",
	})
	suite.Require().NoError(err)

	// Verify it exists
	_, valid := suite.cache.Get()
	suite.True(valid)

	// Clear the cache
	err = suite.cache.Clear()
	suite.Require().NoError(err)

	// Verify it's gone
	_, valid = suite.cache.Get()
	suite.False(valid)
}

// TestClearNonExistent tests clearing when no cache exists
func (suite *UpdateCacheTestSuite) TestClearNonExistent() {
	err := suite.cache.Clear()
	suite.Require().NoError(err) // Should not error when file doesn't exist
}

// TestStoreCreatesDirectory tests that Store creates the cache directory
func (suite *UpdateCacheTestSuite) TestStoreCreatesDirectory() {
	newDir := filepath.Join(suite.tempDir, "nested", "cache", "dir")
	cache := NewUpdateNotifyCacheWithOptions(newDir, 24*time.Hour)

	err := cache.Store(&UpdateCacheData{
		CurrentVersion: "v1.0.0",
		LatestVersion:  "v1.0.0",
	})
	suite.Require().NoError(err)

	// Directory should exist
	_, err = os.Stat(newDir)
	suite.Require().NoError(err)
}

// TestConcurrentAccess tests thread-safe cache access
func (suite *UpdateCacheTestSuite) TestConcurrentAccess() {
	// Run multiple goroutines accessing the cache
	done := make(chan bool, 10)
	errChan := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			// Store
			err := suite.cache.Store(&UpdateCacheData{
				CurrentVersion: "v1.0.0",
				LatestVersion:  "v1.0.0",
			})
			if err != nil {
				errChan <- err
			}

			// Get
			suite.cache.Get()

			// Check expiration
			suite.cache.IsExpired()

			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check for any errors
	close(errChan)
	for err := range errChan {
		suite.Require().NoError(err)
	}
}

// TestCorruptedCacheFile tests handling of corrupted cache file
func (suite *UpdateCacheTestSuite) TestCorruptedCacheFile() {
	// Write invalid JSON to cache file
	cacheFile := suite.cache.GetCacheFile()
	err := os.MkdirAll(filepath.Dir(cacheFile), 0o700)
	suite.Require().NoError(err)
	err = os.WriteFile(cacheFile, []byte("not valid json"), 0o600)
	suite.Require().NoError(err)

	// Get should return invalid
	data, valid := suite.cache.Get()
	suite.False(valid)
	suite.Nil(data)
}

// TestGetUpdateCheckInterval tests the interval parsing
func TestGetUpdateCheckInterval(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected time.Duration
	}{
		{
			name:     "default value",
			envValue: "",
			expected: 24 * time.Hour,
		},
		{
			name:     "valid duration",
			envValue: "12h",
			expected: 12 * time.Hour,
		},
		{
			name:     "short duration enforces minimum",
			envValue: "30m",
			expected: 1 * time.Hour, // Minimum
		},
		{
			name:     "invalid duration uses default",
			envValue: "invalid",
			expected: 24 * time.Hour,
		},
		{
			name:     "very long duration",
			envValue: "168h", // 1 week
			expected: 168 * time.Hour,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore env
			oldVal := os.Getenv("MAGEX_UPDATE_CHECK_INTERVAL")
			defer func() {
				if oldVal != "" {
					_ = os.Setenv("MAGEX_UPDATE_CHECK_INTERVAL", oldVal) //nolint:errcheck // Test env restore
				} else {
					_ = os.Unsetenv("MAGEX_UPDATE_CHECK_INTERVAL") //nolint:errcheck // Test env restore
				}
			}()

			if tt.envValue != "" {
				_ = os.Setenv("MAGEX_UPDATE_CHECK_INTERVAL", tt.envValue) //nolint:errcheck // Test env setup
			} else {
				_ = os.Unsetenv("MAGEX_UPDATE_CHECK_INTERVAL") //nolint:errcheck // Test env setup
			}

			result := getUpdateCheckInterval()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsUpdateCheckDisabled tests the disable flag
func TestIsUpdateCheckDisabled(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name:     "not disabled by default",
			envVars:  map[string]string{},
			expected: false,
		},
		{
			name: "disabled via MAGEX_DISABLE_UPDATE_CHECK",
			envVars: map[string]string{
				"MAGEX_DISABLE_UPDATE_CHECK": "true",
			},
			expected: true,
		},
		{
			name: "disabled in CI",
			envVars: map[string]string{
				"CI": "true",
			},
			expected: true,
		},
		{
			name: "disabled in GitHub Actions",
			envVars: map[string]string{
				"GITHUB_ACTIONS": "true",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore env vars
			savedEnv := map[string]string{}
			for key := range tt.envVars {
				savedEnv[key] = os.Getenv(key)
			}
			// Also save CI-related vars
			for _, key := range []string{"CI", "GITHUB_ACTIONS", "MAGEX_DISABLE_UPDATE_CHECK"} {
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

			// Clear all CI-related env vars first
			_ = os.Unsetenv("CI")                         //nolint:errcheck // Test env setup
			_ = os.Unsetenv("GITHUB_ACTIONS")             //nolint:errcheck // Test env setup
			_ = os.Unsetenv("MAGEX_DISABLE_UPDATE_CHECK") //nolint:errcheck // Test env setup

			// Set test env vars
			for key, val := range tt.envVars {
				_ = os.Setenv(key, val) //nolint:errcheck // Test env setup
			}

			result := isUpdateCheckDisabled()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCacheFilePermissions tests that cache file has correct permissions
func (suite *UpdateCacheTestSuite) TestCacheFilePermissions() {
	err := suite.cache.Store(&UpdateCacheData{
		CurrentVersion: "v1.0.0",
		LatestVersion:  "v1.0.0",
	})
	suite.Require().NoError(err)

	info, err := os.Stat(suite.cache.GetCacheFile())
	suite.Require().NoError(err)

	// File should be readable/writable by owner only (0o600)
	// On some systems, the mode might have extra bits, so we mask it
	mode := info.Mode().Perm()
	suite.Equal(os.FileMode(0o600), mode, "Cache file should have 0o600 permissions")
}

// TestCacheDirPermissions tests that cache directory has correct permissions
func (suite *UpdateCacheTestSuite) TestCacheDirPermissions() {
	newDir := filepath.Join(suite.tempDir, "test-perms")
	cache := NewUpdateNotifyCacheWithOptions(newDir, 24*time.Hour)

	err := cache.Store(&UpdateCacheData{
		CurrentVersion: "v1.0.0",
		LatestVersion:  "v1.0.0",
	})
	suite.Require().NoError(err)

	info, err := os.Stat(newDir)
	suite.Require().NoError(err)

	// Directory should be accessible by owner only (0o700)
	mode := info.Mode().Perm()
	suite.Equal(os.FileMode(0o700), mode, "Cache directory should have 0o700 permissions")
}

// TestOrphanedTempFileCleanup tests that orphaned temp files are cleaned up
func (suite *UpdateCacheTestSuite) TestOrphanedTempFileCleanup() {
	// Create an orphaned temp file
	tempFile := suite.cache.GetCacheFile() + ".tmp"
	err := os.MkdirAll(filepath.Dir(tempFile), 0o700)
	suite.Require().NoError(err)
	err = os.WriteFile(tempFile, []byte("orphaned data"), 0o600)
	suite.Require().NoError(err)

	// Verify temp file exists
	_, err = os.Stat(tempFile)
	suite.Require().NoError(err, "Temp file should exist before cleanup")

	// Create a new cache - this should clean up the orphaned temp file
	newCache := NewUpdateNotifyCacheWithOptions(suite.tempDir, 24*time.Hour)
	newCache.cleanupOrphanedTempFiles()

	// Verify temp file is gone
	_, err = os.Stat(tempFile)
	suite.True(os.IsNotExist(err), "Temp file should be removed after cleanup")
}

// TestIsExpiredConsistency tests that IsExpired and Get use consistent timing
func (suite *UpdateCacheTestSuite) TestIsExpiredConsistency() {
	// Create cache with 1 hour TTL
	cache := NewUpdateNotifyCacheWithOptions(suite.tempDir, 1*time.Hour)

	// Store data
	err := cache.Store(&UpdateCacheData{
		CurrentVersion: "v1.0.0",
		LatestVersion:  "v1.0.0",
	})
	suite.Require().NoError(err)

	// Both Get and IsExpired should agree the cache is valid
	_, valid := cache.Get()
	expired := cache.IsExpired()

	// If Get says valid, IsExpired should say not expired (and vice versa)
	suite.Equal(valid, !expired,
		"Get() valid=%v should be opposite of IsExpired()=%v", valid, expired)
}

// TestIsExpiredWithCorruptedCache tests IsExpired behavior with corrupted cache
func (suite *UpdateCacheTestSuite) TestIsExpiredWithCorruptedCache() {
	// Write invalid JSON to cache file
	cacheFile := suite.cache.GetCacheFile()
	err := os.MkdirAll(filepath.Dir(cacheFile), 0o700)
	suite.Require().NoError(err)
	err = os.WriteFile(cacheFile, []byte("not valid json"), 0o600)
	suite.Require().NoError(err)

	// IsExpired should return true for corrupted cache
	suite.True(suite.cache.IsExpired(), "Corrupted cache should be considered expired")
}

// TestCacheStoreOverwritesExisting tests that Store properly overwrites existing cache
func (suite *UpdateCacheTestSuite) TestCacheStoreOverwritesExisting() {
	// Store initial data
	err := suite.cache.Store(&UpdateCacheData{
		CurrentVersion:  "v1.0.0",
		LatestVersion:   "v1.0.0",
		UpdateAvailable: false,
	})
	suite.Require().NoError(err)

	// Store new data
	err = suite.cache.Store(&UpdateCacheData{
		CurrentVersion:  "v1.0.0",
		LatestVersion:   "v2.0.0",
		UpdateAvailable: true,
	})
	suite.Require().NoError(err)

	// Get should return the new data
	data, valid := suite.cache.Get()
	suite.True(valid)
	suite.Equal("v2.0.0", data.LatestVersion)
	suite.True(data.UpdateAvailable)
}

// TestClearAfterInstallScenario tests the expected behavior after update:install completes
// This simulates the UX flow: user has stale "update available" cache, installs update,
// cache is cleared, next run doesn't show stale notification
func (suite *UpdateCacheTestSuite) TestClearAfterInstallScenario() {
	// Simulate pre-install state: cache shows update available
	err := suite.cache.Store(&UpdateCacheData{
		CurrentVersion:  "v1.0.0",
		LatestVersion:   "v1.1.0",
		UpdateAvailable: true,
		ReleaseNotes:    "Bug fixes and improvements",
	})
	suite.Require().NoError(err)

	// Verify cache exists and shows update available
	data, valid := suite.cache.Get()
	suite.True(valid, "Cache should exist before install")
	suite.True(data.UpdateAvailable, "Cache should show update available")

	// Simulate post-install: cache is cleared (this is what Install() now does)
	err = suite.cache.Clear()
	suite.Require().NoError(err)

	// Verify cache is gone - next run will perform fresh check
	_, valid = suite.cache.Get()
	suite.False(valid, "Cache should be cleared after install")

	// Verify file doesn't exist
	_, err = os.Stat(suite.cache.GetCacheFile())
	suite.True(os.IsNotExist(err), "Cache file should not exist after clear")
}
