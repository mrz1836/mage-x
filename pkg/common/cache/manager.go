// Package cache provides cache management and integration utilities
package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mrz1836/go-mage/pkg/common/fileops"
)

// Manager coordinates different cache types and provides unified cache operations
type Manager struct {
	buildCache *BuildCache
	cacheDir   string
	fileOps    *fileops.FileOps
	config     *Config
}

// Config holds cache configuration
type Config struct {
	Enabled     bool          `yaml:"enabled"`
	Directory   string        `yaml:"directory"`
	MaxSize     int64         `yaml:"max_size"`
	TTL         time.Duration `yaml:"ttl"`
	Compression bool          `yaml:"compression"`
	Strategies  *Strategies   `yaml:"strategies"`
}

// Strategies defines caching strategies for different operations
type Strategies struct {
	Build string `yaml:"build"`
	Test  string `yaml:"test"`
	Lint  string `yaml:"lint"`
	Deps  string `yaml:"deps"`
}

// DefaultConfig returns a default cache configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:     true,
		Directory:   ".mage-cache",
		MaxSize:     5 * 1024 * 1024 * 1024, // 5GB
		TTL:         7 * 24 * time.Hour,     // 7 days
		Compression: true,
		Strategies: &Strategies{
			Build: "file_hash",
			Test:  "content_hash",
			Lint:  "file_hash+config_hash",
			Deps:  "version_hash",
		},
	}
}

// NewManager creates a new cache manager
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	cacheDir := config.Directory
	if !filepath.IsAbs(cacheDir) {
		if cwd, err := os.Getwd(); err == nil {
			cacheDir = filepath.Join(cwd, cacheDir)
		}
	}

	return &Manager{
		buildCache: NewBuildCache(cacheDir),
		cacheDir:   cacheDir,
		fileOps:    fileops.New(),
		config:     config,
	}
}

// Init initializes all cache components
func (m *Manager) Init() error {
	if !m.config.Enabled {
		return nil
	}

	// Create main cache directory
	if err := m.fileOps.File.MkdirAll(m.cacheDir, 0o755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Configure build cache
	m.buildCache.SetOptions(m.config.Enabled, m.config.MaxSize, m.config.TTL)

	// Initialize build cache
	if err := m.buildCache.Init(); err != nil {
		return fmt.Errorf("failed to initialize build cache: %w", err)
	}

	return nil
}

// GetBuildCache returns the build cache instance
func (m *Manager) GetBuildCache() *BuildCache {
	return m.buildCache
}

// GenerateBuildHash creates a hash for build operations
func (m *Manager) GenerateBuildHash(platform, ldflags string, sourceFiles []string, configFiles []string) (string, error) {
	strategy := m.config.Strategies.Build

	switch strategy {
	case "file_hash":
		return m.generateFileBasedHash(sourceFiles, configFiles, platform, ldflags)
	case "content_hash":
		return m.generateContentBasedHash(sourceFiles, configFiles, platform, ldflags)
	case "timestamp_hash":
		return m.generateTimestampBasedHash(sourceFiles, configFiles, platform, ldflags)
	default:
		return m.generateFileBasedHash(sourceFiles, configFiles, platform, ldflags)
	}
}

// GenerateTestHash creates a hash for test operations
func (m *Manager) GenerateTestHash(pkg string, testFiles []string, buildFlags []string) (string, error) {
	allFiles := append(testFiles, "go.mod", "go.sum")

	// Add build flags to hash
	flagsStr := strings.Join(buildFlags, "|")

	return m.buildCache.GenerateHash(
		pkg,
		flagsStr,
		runtime.GOOS,
		runtime.GOARCH,
		func() string {
			hash, _ := m.buildCache.GenerateFileHash(allFiles)
			return hash
		}(),
	), nil
}

// GenerateLintHash creates a hash for lint operations
func (m *Manager) GenerateLintHash(files []string, configFiles []string, lintConfig string) (string, error) {
	allFiles := append(files, configFiles...)

	fileHash, err := m.buildCache.GenerateFileHash(allFiles)
	if err != nil {
		return "", err
	}

	return m.buildCache.GenerateHash(
		fileHash,
		lintConfig,
		"lint",
	), nil
}

// GenerateDependencyHash creates a hash for dependency operations
func (m *Manager) GenerateDependencyHash(modFile, sumFile string, environment map[string]string) (string, error) {
	files := []string{modFile, sumFile}

	fileHash, err := m.buildCache.GenerateFileHash(files)
	if err != nil {
		return "", err
	}

	// Include relevant environment variables
	envStr := ""
	for k, v := range environment {
		if strings.HasPrefix(k, "GO") || k == "PATH" {
			envStr += fmt.Sprintf("%s=%s|", k, v)
		}
	}

	return m.buildCache.GenerateHash(
		fileHash,
		envStr,
		runtime.GOOS,
		runtime.GOARCH,
		"deps",
	), nil
}

// Cleanup performs cache maintenance
func (m *Manager) Cleanup() error {
	if !m.config.Enabled {
		return nil
	}

	// Cleanup build cache
	if err := m.buildCache.Cleanup(); err != nil {
		return fmt.Errorf("failed to cleanup build cache: %w", err)
	}

	// Check cache size and perform size-based cleanup if needed
	stats, err := m.buildCache.GetStats()
	if err != nil {
		return fmt.Errorf("failed to get cache stats: %w", err)
	}

	if stats.TotalSize > m.config.MaxSize {
		return m.performSizeBasedCleanup(stats.TotalSize - m.config.MaxSize)
	}

	return nil
}

// GetStats returns comprehensive cache statistics
func (m *Manager) GetStats() (*CacheStats, error) {
	if !m.config.Enabled {
		return &CacheStats{}, nil
	}

	return m.buildCache.GetStats()
}

// Clear removes all cache data
func (m *Manager) Clear() error {
	if !m.config.Enabled {
		return nil
	}

	return m.buildCache.Clear()
}

// IsEnabled returns whether caching is enabled
func (m *Manager) IsEnabled() bool {
	return m.config.Enabled
}

// GetCacheDir returns the cache directory
func (m *Manager) GetCacheDir() string {
	return m.cacheDir
}

// WarmCache pre-populates cache with common operations
func (m *Manager) WarmCache(operations []string) error {
	if !m.config.Enabled {
		return nil
	}

	// This would trigger common operations to populate cache
	// Implementation depends on specific warming strategies
	fmt.Printf("Warming cache for operations: %v\n", operations)
	return nil
}

// generateFileBasedHash creates hash based on file modification times and sizes
func (m *Manager) generateFileBasedHash(sourceFiles, configFiles []string, platform, ldflags string) (string, error) {
	allFiles := append(sourceFiles, configFiles...)
	allFiles = append(allFiles, "go.mod", "go.sum")

	fileHash, err := m.buildCache.GenerateFileHash(allFiles)
	if err != nil {
		return "", err
	}

	return m.buildCache.GenerateHash(
		fileHash,
		platform,
		ldflags,
		runtime.Version(),
	), nil
}

// generateContentBasedHash creates hash based on actual file contents
func (m *Manager) generateContentBasedHash(sourceFiles, configFiles []string, platform, ldflags string) (string, error) {
	// This is more expensive but more accurate for content changes
	allFiles := append(sourceFiles, configFiles...)
	allFiles = append(allFiles, "go.mod", "go.sum")

	contentHash := ""
	for _, file := range allFiles {
		if m.fileOps.File.Exists(file) {
			content, err := m.fileOps.File.ReadFile(file)
			if err == nil {
				contentHash += m.buildCache.GenerateHash(string(content))
			}
		}
	}

	return m.buildCache.GenerateHash(
		contentHash,
		platform,
		ldflags,
		runtime.Version(),
	), nil
}

// generateTimestampBasedHash creates hash based on timestamps (fastest but least accurate)
func (m *Manager) generateTimestampBasedHash(sourceFiles, configFiles []string, platform, ldflags string) (string, error) {
	allFiles := append(sourceFiles, configFiles...)
	allFiles = append(allFiles, "go.mod", "go.sum")

	var latestTime time.Time
	for _, file := range allFiles {
		if m.fileOps.File.Exists(file) {
			if info, err := m.fileOps.File.Stat(file); err == nil {
				if info.ModTime().After(latestTime) {
					latestTime = info.ModTime()
				}
			}
		}
	}

	return m.buildCache.GenerateHash(
		fmt.Sprintf("%d", latestTime.Unix()),
		platform,
		ldflags,
		runtime.Version(),
	), nil
}

// performSizeBasedCleanup removes old cache entries to free space
func (m *Manager) performSizeBasedCleanup(bytesToFree int64) error {
	// Implementation would remove oldest cache entries until target size is reached
	fmt.Printf("Performing size-based cleanup: freeing %d bytes\n", bytesToFree)

	// This is a simplified implementation - in practice, you'd:
	// 1. Collect all cache entries with their timestamps
	// 2. Sort by access time (oldest first)
	// 3. Remove entries until enough space is freed

	return m.buildCache.Cleanup()
}
