// Package cache provides build caching capabilities for improved performance
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// BuildCache provides caching for build operations
type BuildCache struct {
	cacheDir string
	fileOps  *fileops.FileOps
	enabled  bool
	maxSize  int64 // in bytes
	ttl      time.Duration
}

// BuildResult represents a cached build result
type BuildResult struct {
	Hash        string            `json:"hash"`
	Binary      string            `json:"binary"`
	Platform    string            `json:"platform"`
	BuildFlags  []string          `json:"build_flags"`
	Environment map[string]string `json:"environment"`
	Timestamp   time.Time         `json:"timestamp"`
	Success     bool              `json:"success"`
	Error       string            `json:"error,omitempty"`
	Metrics     BuildMetrics      `json:"metrics"`
}

// TestResult represents a cached test result
type TestResult struct {
	Hash      string        `json:"hash"`
	Package   string        `json:"package"`
	Success   bool          `json:"success"`
	Output    string        `json:"output"`
	Coverage  float64       `json:"coverage,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
	Metrics   TestMetrics   `json:"metrics"`
}

// LintResult represents a cached lint result
type LintResult struct {
	Hash      string        `json:"hash"`
	Files     []string      `json:"files"`
	Success   bool          `json:"success"`
	Issues    []LintIssue   `json:"issues"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// DependencyResult represents a cached dependency operation
type DependencyResult struct {
	Hash        string            `json:"hash"`
	ModFile     string            `json:"mod_file"`
	SumFile     string            `json:"sum_file"`
	Modules     []ModuleInfo      `json:"modules"`
	Success     bool              `json:"success"`
	Timestamp   time.Time         `json:"timestamp"`
	Environment map[string]string `json:"environment"`
}

// BuildMetrics contains build performance metrics
type BuildMetrics struct {
	CompileTime time.Duration `json:"compile_time"`
	LinkTime    time.Duration `json:"link_time"`
	BinarySize  int64         `json:"binary_size"`
	SourceFiles int           `json:"source_files"`
}

// TestMetrics contains test performance metrics
type TestMetrics struct {
	TestCount   int           `json:"test_count"`
	PassCount   int           `json:"pass_count"`
	FailCount   int           `json:"fail_count"`
	SkipCount   int           `json:"skip_count"`
	CompileTime time.Duration `json:"compile_time"`
	ExecuteTime time.Duration `json:"execute_time"`
}

// LintIssue represents a linting issue
type LintIssue struct {
	File     string `json:"file"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Rule     string `json:"rule"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// ModuleInfo represents module information
type ModuleInfo struct {
	Path    string    `json:"path"`
	Version string    `json:"version"`
	Hash    string    `json:"hash"`
	Time    time.Time `json:"time"`
}

// Stats provides cache statistics
type Stats struct {
	Hits        int64     `json:"hits"`
	Misses      int64     `json:"misses"`
	HitRate     float64   `json:"hit_rate"`
	TotalSize   int64     `json:"total_size"`
	EntryCount  int       `json:"entry_count"`
	LastCleanup time.Time `json:"last_cleanup"`
}

// NewBuildCache creates a new build cache
func NewBuildCache(cacheDir string) *BuildCache {
	return &BuildCache{
		cacheDir: cacheDir,
		fileOps:  fileops.New(),
		enabled:  true,
		maxSize:  5 * 1024 * 1024 * 1024, // 5GB default
		ttl:      7 * 24 * time.Hour,     // 7 days default
	}
}

// SetOptions configures cache options
func (c *BuildCache) SetOptions(enabled bool, maxSize int64, ttl time.Duration) {
	c.enabled = enabled
	c.maxSize = maxSize
	c.ttl = ttl
}

// Init initializes the cache directory structure
func (c *BuildCache) Init() error {
	if !c.enabled {
		return nil
	}

	dirs := []string{
		"builds",
		"tests",
		"lint",
		"deps",
		"tools",
		"meta",
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(c.cacheDir, dir)
		if err := c.fileOps.File.MkdirAll(dirPath, 0o755); err != nil {
			return fmt.Errorf("failed to create cache directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetBuildResult retrieves a cached build result
func (c *BuildCache) GetBuildResult(hash string) (*BuildResult, bool) {
	if !c.enabled {
		return nil, false
	}

	path := filepath.Join(c.cacheDir, "builds", hash+".json")
	if !c.fileOps.File.Exists(path) {
		return nil, false
	}

	data, err := c.fileOps.File.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var result BuildResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, false
	}

	// Check if cache entry is expired
	if time.Since(result.Timestamp) > c.ttl {
		if err := c.removeCacheEntry(path); err != nil {
			utils.Warn("Failed to remove expired cache entry %s: %v", path, err)
		}
		return nil, false
	}

	// Verify binary still exists
	if result.Success && !c.fileOps.File.Exists(result.Binary) {
		if err := c.removeCacheEntry(path); err != nil {
			utils.Warn("Failed to remove cache entry with missing binary %s: %v", path, err)
		}
		return nil, false
	}

	return &result, true
}

// StoreBuildResult stores a build result in cache
func (c *BuildCache) StoreBuildResult(hash string, result *BuildResult) error {
	if !c.enabled {
		return nil
	}

	result.Hash = hash
	result.Timestamp = time.Now()

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(c.cacheDir, "builds", hash+".json")
	return c.fileOps.File.WriteFile(path, data, 0o644)
}

// GetTestResult retrieves a cached test result
func (c *BuildCache) GetTestResult(hash string) (*TestResult, bool) {
	if !c.enabled {
		return nil, false
	}

	path := filepath.Join(c.cacheDir, "tests", hash+".json")
	if !c.fileOps.File.Exists(path) {
		return nil, false
	}

	data, err := c.fileOps.File.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var result TestResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, false
	}

	// Check if cache entry is expired
	if time.Since(result.Timestamp) > c.ttl {
		if err := c.removeCacheEntry(path); err != nil {
			utils.Warn("Failed to remove expired test cache entry %s: %v", path, err)
		}
		return nil, false
	}

	return &result, true
}

// StoreTestResult stores a test result in cache
func (c *BuildCache) StoreTestResult(hash string, result *TestResult) error {
	if !c.enabled {
		return nil
	}

	result.Hash = hash
	result.Timestamp = time.Now()

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(c.cacheDir, "tests", hash+".json")
	return c.fileOps.File.WriteFile(path, data, 0o644)
}

// GetLintResult retrieves a cached lint result
func (c *BuildCache) GetLintResult(hash string) (*LintResult, bool) {
	if !c.enabled {
		return nil, false
	}

	path := filepath.Join(c.cacheDir, "lint", hash+".json")
	if !c.fileOps.File.Exists(path) {
		return nil, false
	}

	data, err := c.fileOps.File.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var result LintResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, false
	}

	// Check if cache entry is expired
	if time.Since(result.Timestamp) > c.ttl {
		if err := c.removeCacheEntry(path); err != nil {
			utils.Warn("Failed to remove expired lint cache entry %s: %v", path, err)
		}
		return nil, false
	}

	return &result, true
}

// StoreLintResult stores a lint result in cache
func (c *BuildCache) StoreLintResult(hash string, result *LintResult) error {
	if !c.enabled {
		return nil
	}

	result.Hash = hash
	result.Timestamp = time.Now()

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(c.cacheDir, "lint", hash+".json")
	return c.fileOps.File.WriteFile(path, data, 0o644)
}

// GetDependencyResult retrieves a cached dependency result
func (c *BuildCache) GetDependencyResult(hash string) (*DependencyResult, bool) {
	if !c.enabled {
		return nil, false
	}

	path := filepath.Join(c.cacheDir, "deps", hash+".json")
	if !c.fileOps.File.Exists(path) {
		return nil, false
	}

	data, err := c.fileOps.File.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var result DependencyResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, false
	}

	// Check if cache entry is expired
	if time.Since(result.Timestamp) > c.ttl {
		if err := c.removeCacheEntry(path); err != nil {
			utils.Warn("Failed to remove expired dependency cache entry %s: %v", path, err)
		}
		return nil, false
	}

	return &result, true
}

// StoreDependencyResult stores a dependency result in cache
func (c *BuildCache) StoreDependencyResult(hash string, result *DependencyResult) error {
	if !c.enabled {
		return nil
	}

	result.Hash = hash
	result.Timestamp = time.Now()

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(c.cacheDir, "deps", hash+".json")
	return c.fileOps.File.WriteFile(path, data, 0o644)
}

// GenerateHash creates a hash for cache keys
func (c *BuildCache) GenerateHash(inputs ...string) string {
	h := sha256.New()
	for _, input := range inputs {
		h.Write([]byte(input))
	}
	return hex.EncodeToString(h.Sum(nil))[:16] // Use first 16 chars for shorter keys
}

// GenerateFileHash creates a hash based on file contents
func (c *BuildCache) GenerateFileHash(filePaths []string) (string, error) {
	h := sha256.New()

	for _, path := range filePaths {
		if !c.fileOps.File.Exists(path) {
			continue
		}

		// Add file path to hash
		h.Write([]byte(path))

		// Add file modification time and size
		info, err := c.fileOps.File.Stat(path)
		if err != nil {
			continue
		}
		if _, err := fmt.Fprintf(h, "%d-%d", info.ModTime().Unix(), info.Size()); err != nil {
			continue // Skip if we can't write to hash
		}

		// For small files, include content hash
		if info.Size() < 1024*1024 { // 1MB limit
			content, err := c.fileOps.File.ReadFile(path)
			if err == nil {
				h.Write(content)
			}
		}
	}

	return hex.EncodeToString(h.Sum(nil))[:16], nil
}

// GetStats returns cache statistics
func (c *BuildCache) GetStats() (*Stats, error) {
	if !c.enabled {
		return &Stats{}, nil
	}

	stats := &Stats{
		LastCleanup: time.Now(),
	}

	// Calculate cache size and entry count
	err := filepath.WalkDir(c.cacheDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Log error but continue walking to gather partial stats
			utils.Debug("cache stats: skipping path %s due to error: %v", path, err)
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				stats.TotalSize += info.Size()
				stats.EntryCount++
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Calculate hit rate if we have both hits and misses
	if stats.Hits+stats.Misses > 0 {
		stats.HitRate = float64(stats.Hits) / float64(stats.Hits+stats.Misses)
	}

	return stats, nil
}

// Cleanup removes expired cache entries
func (c *BuildCache) Cleanup() error {
	if !c.enabled {
		return nil
	}

	removed := 0
	err := filepath.WalkDir(c.cacheDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Log error but continue cleanup to remove what we can
			utils.Debug("cache cleanup: skipping path %s due to error: %v", path, err)
			return nil
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			info, err := d.Info()
			if err == nil && time.Since(info.ModTime()) > c.ttl {
				if c.removeCacheEntry(path) == nil {
					removed++
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	utils.Info("Cache cleanup completed: removed %d expired entries", removed)
	return nil
}

// Clear removes all cache entries
func (c *BuildCache) Clear() error {
	if !c.enabled {
		return nil
	}

	return c.fileOps.File.RemoveAll(c.cacheDir)
}

// removeCacheEntry removes a single cache entry
func (c *BuildCache) removeCacheEntry(path string) error {
	return c.fileOps.File.Remove(path)
}

// IsEnabled returns whether caching is enabled
func (c *BuildCache) IsEnabled() bool {
	return c.enabled
}

// GetCacheDir returns the cache directory path
func (c *BuildCache) GetCacheDir() string {
	return c.cacheDir
}
