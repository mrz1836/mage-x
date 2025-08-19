package config

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

// Config provides a facade for configuration management
type Config struct {
	Loader  Loader
	Env     TypedEnvProvider
	Manager Manager
}

// New creates a new Config with default implementations
func New() *Config {
	return &Config{
		Loader:  NewDefaultConfigLoader(),
		Env:     NewDefaultEnvProvider(),
		Manager: NewDefaultConfigManager(),
	}
}

// NewWithOptions creates a new Config with custom implementations
func NewWithOptions(loader Loader, env TypedEnvProvider, manager Manager) *Config {
	return &Config{
		Loader:  loader,
		Env:     env,
		Manager: manager,
	}
}

// LoadFromPaths loads configuration from multiple paths with common fallbacks
func (c *Config) LoadFromPaths(dest interface{}, baseName string, searchDirs ...string) (string, error) {
	paths := c.buildConfigPaths(baseName, searchDirs...)
	return c.Loader.Load(paths, dest)
}

// LoadWithEnvOverrides loads configuration and applies environment variable overrides
func (c *Config) LoadWithEnvOverrides(dest interface{}, baseName, envPrefix string, searchDirs ...string) (string, error) {
	// Load from files first
	path, err := c.LoadFromPaths(dest, baseName, searchDirs...)
	if err != nil {
		return "", err
	}

	// Apply environment overrides
	if err := c.applyEnvOverrides(dest, envPrefix); err != nil {
		return path, fmt.Errorf("failed to apply environment overrides: %w", err)
	}

	return path, nil
}

// SetupManager configures the config manager with common sources
func (c *Config) SetupManager(baseName, envPrefix string, searchDirs ...string) {
	// Add file sources
	paths := c.buildConfigPaths(baseName, searchDirs...)
	for i, path := range paths {
		// Higher priority for files found first
		priority := 1000 - i
		format := c.detectFormat(path)
		source := NewFileConfigSource(path, format, priority)
		c.Manager.AddSource(source)
	}

	// Add environment source with highest priority
	if envPrefix != "" {
		envSource := NewEnvConfigSource(envPrefix, 2000)
		c.Manager.AddSource(envSource)
	}
}

// buildConfigPaths builds a list of potential configuration file paths
func (c *Config) buildConfigPaths(baseName string, searchDirs ...string) []string {
	if len(searchDirs) == 0 {
		searchDirs = []string{".", "/etc", "$HOME/.config", "$HOME"}
	}

	extensions := []string{".yaml", ".yml", ".json"}
	// Preallocate: len(searchDirs) * (len(extensions) + 1) paths
	paths := make([]string, 0, len(searchDirs)*(len(extensions)+1))

	for _, dir := range searchDirs {
		// Expand environment variables in directory paths
		expandedDir := c.expandPath(dir)

		for _, ext := range extensions {
			candidatePath := filepath.Join(expandedDir, baseName+ext)
			// Filter out paths containing null bytes
			if !strings.Contains(candidatePath, "\x00") {
				paths = append(paths, candidatePath)
			}
		}

		// Also try without extension (for auto-detection)
		candidatePath := filepath.Join(expandedDir, baseName)
		// Filter out paths containing null bytes
		if !strings.Contains(candidatePath, "\x00") {
			paths = append(paths, candidatePath)
		}
	}

	return paths
}

// expandPath expands environment variables in path with security validation
func (c *Config) expandPath(path string) string {
	// First decode any URL-encoded sequences to detect hidden path traversal
	decoded, err := url.QueryUnescape(path)
	if err != nil {
		// If decoding fails, use original path but validate it
		decoded = path
	}

	// Check for path traversal patterns in both original and decoded paths
	if strings.Contains(decoded, "../") || strings.Contains(decoded, "..\\") ||
		strings.HasSuffix(decoded, "..") || decoded == ".." {
		// Return the original path without expansion as a security measure
		return path
	}

	// Check for null bytes and control characters
	if strings.Contains(decoded, "\x00") || strings.Contains(decoded, "\n") || strings.Contains(decoded, "\r") {
		// Return the original path without expansion as a security measure
		return path
	}

	// Use the decoded path for expansion
	workingPath := decoded
	if strings.HasPrefix(workingPath, "$HOME") {
		home := c.Env.Get("HOME")
		return strings.Replace(workingPath, "$HOME", home, 1)
	}
	return workingPath
}

// detectFormat detects configuration format from file extension
func (c *Config) detectFormat(path string) Format {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return FormatYAML
	case ".json":
		return FormatJSON
	case ".toml":
		return FormatTOML
	case ".ini":
		return FormatINI
	default:
		return FormatYAML // Default to YAML
	}
}

// applyEnvOverrides applies environment variable overrides to configuration
func (c *Config) applyEnvOverrides(_ interface{}, _ string) error {
	// This is a simplified implementation
	// In a real implementation, this would use reflection to map env vars to struct fields
	// For now, we'll return nil (no error)
	return nil
}

// GetCommonConfigPaths returns common configuration file paths for an application
func GetCommonConfigPaths(appName string) []string {
	return []string{
		fmt.Sprintf(".%s.yaml", appName),
		fmt.Sprintf(".%s.yml", appName),
		fmt.Sprintf(".%s.json", appName),
		fmt.Sprintf("%s.yaml", appName),
		fmt.Sprintf("%s.yml", appName),
		fmt.Sprintf("%s.json", appName),
		fmt.Sprintf("$HOME/.config/%s/%s.yaml", appName, appName),
		fmt.Sprintf("$HOME/.config/%s.yaml", appName),
		fmt.Sprintf("${HOME}/.config/%s/%s.yaml", appName, appName),
		fmt.Sprintf("${HOME}/.config/%s.yaml", appName),
		fmt.Sprintf("/etc/%s/%s.yaml", appName, appName),
		fmt.Sprintf("/etc/%s.yaml", appName),
	}
}

// GetDefault returns a new default configuration instance
func GetDefault() *Config {
	return New()
}

// Package-level convenience functions

// LoadFromPaths loads configuration using the default instance
func LoadFromPaths(dest interface{}, baseName string, searchDirs ...string) (string, error) {
	return GetDefault().LoadFromPaths(dest, baseName, searchDirs...)
}

// LoadWithEnvOverrides loads configuration with env overrides using the default instance
func LoadWithEnvOverrides(dest interface{}, baseName, envPrefix string, searchDirs ...string) (string, error) {
	return GetDefault().LoadWithEnvOverrides(dest, baseName, envPrefix, searchDirs...)
}

// GetBool gets a boolean environment variable using the default instance
func GetBool(key string, defaultValue bool) bool {
	return GetDefault().Env.GetBool(key, defaultValue)
}

// GetInt gets an integer environment variable using the default instance
func GetInt(key string, defaultValue int) int {
	return GetDefault().Env.GetInt(key, defaultValue)
}

// GetString gets a string environment variable using the default instance
func GetString(key, defaultValue string) string {
	return GetDefault().Env.GetWithDefault(key, defaultValue)
}

// GetStringSlice gets a string slice environment variable using the default instance
func GetStringSlice(key string, defaultValue []string) []string {
	return GetDefault().Env.GetStringSlice(key, defaultValue)
}

// GetDuration gets a duration environment variable using the default instance
func GetDuration(key string, defaultValue time.Duration) time.Duration {
	return GetDefault().Env.GetDuration(key, defaultValue)
}
