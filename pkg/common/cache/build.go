// Package cache provides build caching capabilities for improved performance
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/mrz1836/mage-x/pkg/common/fileops"
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

// getCacheResult is a generic helper for retrieving cache results
func (c *BuildCache) getCacheResult(hash, subdir, entryType string, result interface{}) (interface{}, bool) {
	if !c.enabled {
		return nil, false
	}

	path := filepath.Join(c.cacheDir, subdir, hash+".json")
	if !c.fileOps.File.Exists(path) {
		return nil, false
	}

	data, err := c.fileOps.File.ReadFile(path)
	if err != nil {
		return nil, false
	}

	if err := json.Unmarshal(data, result); err != nil {
		return nil, false
	}

	// Check if cache entry is expired using reflection
	var timestamp time.Time
	switch r := result.(type) {
	case *BuildResult:
		timestamp = r.Timestamp
	case *TestResult:
		timestamp = r.Timestamp
	case *LintResult:
		timestamp = r.Timestamp
	case *DependencyResult:
		timestamp = r.Timestamp
	}

	if time.Since(timestamp) > c.ttl {
		if err := c.removeCacheEntry(path); err != nil {
			log.Printf("[WARN] Failed to remove expired %s cache entry %s: %v", entryType, path, err)
		}
		return nil, false
	}

	return result, true
}

// GetBuildResult retrieves a cached build result
func (c *BuildCache) GetBuildResult(hash string) (*BuildResult, bool) {
	var result BuildResult
	if cached, found := c.getCacheResult(hash, "builds", "", &result); found {
		buildResult, ok := cached.(*BuildResult)
		if !ok {
			return nil, false
		}
		// Verify binary still exists
		if buildResult.Success && !c.fileOps.File.Exists(buildResult.Binary) {
			path := filepath.Join(c.cacheDir, "builds", hash+".json")
			if err := c.removeCacheEntry(path); err != nil {
				log.Printf("[WARN] Failed to remove cache entry with missing binary %s: %v", path, err)
			}
			return nil, false
		}
		return buildResult, true
	}
	return nil, false
}

// storeCacheResult is a generic helper for storing cache results
func (c *BuildCache) storeCacheResult(hash, subdir string, result interface{}) error {
	if !c.enabled {
		return nil
	}

	// Set hash and timestamp using reflection for interface{} types
	switch r := result.(type) {
	case *BuildResult:
		r.Hash = hash
		r.Timestamp = time.Now()
	case *TestResult:
		r.Hash = hash
		r.Timestamp = time.Now()
	case *LintResult:
		r.Hash = hash
		r.Timestamp = time.Now()
	case *DependencyResult:
		r.Hash = hash
		r.Timestamp = time.Now()
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(c.cacheDir, subdir, hash+".json")
	return c.fileOps.File.WriteFile(path, data, 0o644)
}

// StoreBuildResult stores a build result in cache
func (c *BuildCache) StoreBuildResult(hash string, result *BuildResult) error {
	return c.storeCacheResult(hash, "builds", result)
}

// GetTestResult retrieves a cached test result
func (c *BuildCache) GetTestResult(hash string) (*TestResult, bool) {
	var result TestResult
	if cached, found := c.getCacheResult(hash, "tests", "test", &result); found {
		if testResult, ok := cached.(*TestResult); ok {
			return testResult, true
		}
	}
	return nil, false
}

// StoreTestResult stores a test result in cache
func (c *BuildCache) StoreTestResult(hash string, result *TestResult) error {
	return c.storeCacheResult(hash, "tests", result)
}

// GetLintResult retrieves a cached lint result
func (c *BuildCache) GetLintResult(hash string) (*LintResult, bool) {
	var result LintResult
	if cached, found := c.getCacheResult(hash, "lint", "lint", &result); found {
		if lintResult, ok := cached.(*LintResult); ok {
			return lintResult, true
		}
	}
	return nil, false
}

// StoreLintResult stores a lint result in cache
func (c *BuildCache) StoreLintResult(hash string, result *LintResult) error {
	return c.storeCacheResult(hash, "lint", result)
}

// GetDependencyResult retrieves a cached dependency result
func (c *BuildCache) GetDependencyResult(hash string) (*DependencyResult, bool) {
	var result DependencyResult
	if cached, found := c.getCacheResult(hash, "deps", "dependency", &result); found {
		if depResult, ok := cached.(*DependencyResult); ok {
			return depResult, true
		}
	}
	return nil, false
}

// StoreDependencyResult stores a dependency result in cache
func (c *BuildCache) StoreDependencyResult(hash string, result *DependencyResult) error {
	return c.storeCacheResult(hash, "deps", result)
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
			log.Printf("[DEBUG] cache stats: skipping path %s due to error: %v", path, err)
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
			log.Printf("[DEBUG] cache cleanup: skipping path %s due to error: %v", path, err)
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

	log.Printf("[INFO] Cache cleanup completed: removed %d expired entries", removed)
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
