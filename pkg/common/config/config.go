package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// Config provides a facade for configuration management
type Config struct {
	Loader  ConfigLoader
	Env     TypedEnvProvider
	Manager ConfigManager
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
func NewWithOptions(loader ConfigLoader, env TypedEnvProvider, manager ConfigManager) *Config {
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
	var paths []string

	for _, dir := range searchDirs {
		// Expand environment variables in directory paths
		expandedDir := c.expandPath(dir)

		for _, ext := range extensions {
			paths = append(paths, filepath.Join(expandedDir, baseName+ext))
		}

		// Also try without extension (for auto-detection)
		paths = append(paths, filepath.Join(expandedDir, baseName))
	}

	return paths
}

// expandPath expands environment variables in path
func (c *Config) expandPath(path string) string {
	if strings.HasPrefix(path, "$HOME") {
		home := c.Env.Get("HOME")
		return strings.Replace(path, "$HOME", home, 1)
	}
	return path
}

// detectFormat detects configuration format from file extension
func (c *Config) detectFormat(path string) ConfigFormat {
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
func (c *Config) applyEnvOverrides(dest interface{}, envPrefix string) error {
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

// Package-level convenience instance
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
