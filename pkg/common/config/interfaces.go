// Package config provides configuration management utilities with mockable interfaces
package config

import (
	"time"
)

// ConfigLoader provides an interface for loading and saving configurations
type ConfigLoader interface {
	// Load loads configuration from multiple file paths with fallback
	Load(paths []string, dest interface{}) (string, error)

	// LoadFrom loads configuration from a specific file
	LoadFrom(path string, dest interface{}) error

	// Save saves configuration to a file in the specified format
	Save(path string, data interface{}, format string) error

	// Validate validates configuration data
	Validate(data interface{}) error

	// GetSupportedFormats returns list of supported file formats
	GetSupportedFormats() []string
}

// EnvProvider provides an interface for environment variable operations
type EnvProvider interface {
	// Get retrieves an environment variable
	Get(key string) string

	// GetWithDefault retrieves an environment variable with a default value
	GetWithDefault(key, defaultValue string) string

	// Set sets an environment variable
	Set(key, value string) error

	// LookupEnv retrieves an environment variable and reports whether it was found
	LookupEnv(key string) (string, bool)

	// Unset removes an environment variable
	Unset(key string) error

	// GetAll returns all environment variables as a map
	GetAll() map[string]string
}

// TypedEnvProvider provides type-safe environment variable operations
type TypedEnvProvider interface {
	EnvProvider

	// GetBool retrieves a boolean environment variable
	GetBool(key string, defaultValue bool) bool

	// GetInt retrieves an integer environment variable
	GetInt(key string, defaultValue int) int

	// GetInt64 retrieves an int64 environment variable
	GetInt64(key string, defaultValue int64) int64

	// GetFloat64 retrieves a float64 environment variable
	GetFloat64(key string, defaultValue float64) float64

	// GetDuration retrieves a duration environment variable
	GetDuration(key string, defaultValue time.Duration) time.Duration

	// GetStringSlice retrieves a string slice environment variable (comma-separated)
	GetStringSlice(key string, defaultValue []string) []string
}

// ConfigSource represents a configuration source
type ConfigSource interface {
	// Name returns the name of the configuration source
	Name() string

	// Load loads configuration data from the source
	Load(dest interface{}) error

	// IsAvailable checks if the source is available
	IsAvailable() bool

	// Priority returns the priority of this source (higher = more important)
	Priority() int
}

// ConfigManager manages multiple configuration sources
type ConfigManager interface {
	// AddSource adds a configuration source
	AddSource(source ConfigSource)

	// LoadConfig loads configuration from all sources in priority order
	LoadConfig(dest interface{}) error

	// Reload reloads configuration from all sources
	Reload(dest interface{}) error

	// Watch watches for configuration changes (if supported)
	Watch(callback func(interface{})) error

	// StopWatching stops watching for configuration changes
	StopWatching()

	// GetActiveSources returns list of currently active sources
	GetActiveSources() []ConfigSource
}

// Validator provides configuration validation capabilities
type Validator interface {
	// Validate validates configuration data
	Validate(data interface{}) error

	// ValidateField validates a specific field
	ValidateField(fieldName string, value interface{}) error

	// GetValidationRules returns current validation rules
	GetValidationRules() map[string]interface{}

	// SetValidationRules sets validation rules
	SetValidationRules(rules map[string]interface{})
}

// ConfigOptions holds configuration loading options
type ConfigOptions struct {
	// Paths to search for configuration files
	Paths []string

	// Formats to try when loading (e.g., "yaml", "json")
	Formats []string

	// StrictMode fails on unknown fields
	StrictMode bool

	// AllowMissingFiles doesn't fail if config files don't exist
	AllowMissingFiles bool

	// EnvPrefix for environment variable overrides
	EnvPrefix string

	// ValidateOnLoad validates configuration after loading
	ValidateOnLoad bool

	// AutoReload enables automatic reloading when files change
	AutoReload bool

	// MergeStrategy defines how to merge multiple config sources
	MergeStrategy MergeStrategy
}

// MergeStrategy defines how configuration values are merged
type MergeStrategy int

const (
	// MergeOverwrite replaces values completely
	MergeOverwrite MergeStrategy = iota

	// MergeDeep performs deep merging of objects
	MergeDeep

	// MergeAppend appends to arrays/slices
	MergeAppend
)

// ConfigFormat represents supported configuration formats
type ConfigFormat string

const (
	// FormatYAML represents YAML configuration format
	FormatYAML ConfigFormat = "yaml"
	// FormatJSON represents JSON configuration format
	FormatJSON ConfigFormat = "json"
	// FormatTOML represents TOML configuration format
	FormatTOML ConfigFormat = "toml"
	// FormatINI represents INI configuration format
	FormatINI ConfigFormat = "ini"
	// FormatEnv represents environment variable configuration format
	FormatEnv ConfigFormat = "env"
)

// Type aliases to avoid stuttering while maintaining backwards compatibility.
type (
	// Source represents a configuration source.
	Source = ConfigSource
	// Manager manages multiple configuration sources.
	Manager = ConfigManager
	// Options holds configuration loading options.
	Options = ConfigOptions
	// Format represents supported configuration formats.
	Format = ConfigFormat
)
