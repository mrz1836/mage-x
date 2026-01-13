// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mrz1836/mage-x/pkg/common/env"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
)

// Update check cache constants
const (
	// updateCacheDir is the directory name under home for magex cache
	updateCacheDir = ".magex"

	// updateCacheFile is the filename for the update check cache
	updateCacheFile = "update-check.json"

	// defaultUpdateCheckInterval is the default time between version checks
	defaultUpdateCheckInterval = 24 * time.Hour

	// minUpdateCheckInterval is the minimum allowed check interval to prevent API abuse
	minUpdateCheckInterval = 1 * time.Hour
)

// UpdateCacheData represents the cached update check data persisted to disk
type UpdateCacheData struct {
	LastCheck       time.Time `json:"last_check"`
	CurrentVersion  string    `json:"current_version"`
	LatestVersion   string    `json:"latest_version"`
	UpdateAvailable bool      `json:"update_available"`
	ReleaseNotes    string    `json:"release_notes,omitempty"`
	ReleaseURL      string    `json:"release_url,omitempty"`
	CheckInterval   string    `json:"check_interval,omitempty"` // Stored for debugging
}

// UpdateNotifyCache manages the update check cache in ~/.magex/
type UpdateNotifyCache struct {
	cacheDir  string
	cacheFile string
	ttl       time.Duration
	mu        sync.RWMutex
}

// CacheProvider interface for testing
type CacheProvider interface {
	Get() (*UpdateCacheData, bool)
	Store(data *UpdateCacheData) error
	IsExpired() bool
	Clear() error
	GetCacheDir() string
}

// NewUpdateNotifyCache creates a new update cache instance
func NewUpdateNotifyCache() *UpdateNotifyCache {
	homeDir := env.Home()
	cacheDir := filepath.Join(homeDir, updateCacheDir)

	cache := &UpdateNotifyCache{
		cacheDir:  cacheDir,
		cacheFile: filepath.Join(cacheDir, updateCacheFile),
		ttl:       getUpdateCheckInterval(),
	}

	// Clean up any orphaned temp files from previous crashed runs
	cache.cleanupOrphanedTempFiles()

	return cache
}

// NewUpdateNotifyCacheWithOptions creates a cache with custom options (for testing)
func NewUpdateNotifyCacheWithOptions(cacheDir string, ttl time.Duration) *UpdateNotifyCache {
	return &UpdateNotifyCache{
		cacheDir:  cacheDir,
		cacheFile: filepath.Join(cacheDir, updateCacheFile),
		ttl:       ttl,
	}
}

// Get retrieves the cached update check data
// Returns the data and true if valid cache exists, nil and false otherwise
func (c *UpdateNotifyCache) Get() (*UpdateCacheData, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Read the cache file
	data, err := os.ReadFile(c.cacheFile)
	if err != nil {
		return nil, false
	}

	var cached UpdateCacheData
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, false
	}

	// Check if cache is expired
	if time.Since(cached.LastCheck) > c.ttl {
		return nil, false
	}

	return &cached, true
}

// Store saves the update check data to the cache file
func (c *UpdateNotifyCache) Store(data *UpdateCacheData) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Ensure cache directory exists
	if err := os.MkdirAll(c.cacheDir, fileops.PermDirPrivate); err != nil {
		return err
	}

	// Set the check time and interval for debugging
	data.LastCheck = time.Now()
	data.CheckInterval = c.ttl.String()

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	// Write atomically by writing to temp file first, then renaming
	tempFile := c.cacheFile + ".tmp"
	if err := os.WriteFile(tempFile, jsonData, fileops.PermFileSensitive); err != nil {
		return err
	}

	return os.Rename(tempFile, c.cacheFile)
}

// IsExpired checks if the cache is expired by reading the JSON LastCheck field
// This ensures consistency with Get() which also uses the JSON timestamp
func (c *UpdateNotifyCache) IsExpired() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Read the cache file
	data, err := os.ReadFile(c.cacheFile)
	if err != nil {
		return true
	}

	var cached UpdateCacheData
	if err := json.Unmarshal(data, &cached); err != nil {
		return true
	}

	return time.Since(cached.LastCheck) > c.ttl
}

// Clear removes the cache file
func (c *UpdateNotifyCache) Clear() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := os.Remove(c.cacheFile); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// GetCacheDir returns the cache directory path
func (c *UpdateNotifyCache) GetCacheDir() string {
	return c.cacheDir
}

// GetCacheFile returns the cache file path
func (c *UpdateNotifyCache) GetCacheFile() string {
	return c.cacheFile
}

// GetTTL returns the cache TTL
func (c *UpdateNotifyCache) GetTTL() time.Duration {
	return c.ttl
}

// cleanupOrphanedTempFiles removes any .tmp files left over from crashed writes
// This prevents accumulation of orphaned temp files over time
func (c *UpdateNotifyCache) cleanupOrphanedTempFiles() {
	tempFile := c.cacheFile + ".tmp"
	// Best effort cleanup - ignore errors as this is not critical
	_ = os.Remove(tempFile) //nolint:errcheck // Best effort cleanup
}

// getUpdateCheckInterval returns the configured update check interval
func getUpdateCheckInterval() time.Duration {
	intervalStr := env.GetString("MAGEX_UPDATE_CHECK_INTERVAL", "24h")

	duration, err := time.ParseDuration(intervalStr)
	if err != nil {
		return defaultUpdateCheckInterval
	}

	// Enforce minimum interval to prevent API abuse
	if duration < minUpdateCheckInterval {
		return minUpdateCheckInterval
	}

	return duration
}

// isUpdateCheckDisabled returns true if update checking is disabled
func isUpdateCheckDisabled() bool {
	// Check explicit disable flag
	if env.GetBool("MAGEX_DISABLE_UPDATE_CHECK", false) {
		return true
	}

	// Disable in CI environments
	if env.IsCI() {
		return true
	}

	return false
}
