package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/mrz1836/go-mage/pkg/common/fileops"
)

// Static errors for configuration loading
var (
	errEnvConfigNotImplemented = errors.New("environment configuration loading not implemented yet")
)

var (
	// ErrNoValidConfigFile is returned when no valid configuration file is found in any of the provided paths
	ErrNoValidConfigFile = errors.New("no valid configuration file found in paths")
	// ErrConfigFileNotExists is returned when a configuration file does not exist
	ErrConfigFileNotExists = errors.New("configuration file does not exist")
	// ErrUnsupportedFormat is returned when an unsupported format is encountered
	ErrUnsupportedFormat = errors.New("unsupported format")
	// ErrConfigDataNil is returned when configuration data is nil
	ErrConfigDataNil = errors.New("configuration data is nil")
)

// DefaultConfigLoader implements ConfigLoader using fileops
type DefaultConfigLoader struct {
	fileOps fileops.FileOperator
	jsonOps fileops.JSONOperator
	yamlOps fileops.YAMLOperator
}

// NewDefaultConfigLoader creates a new default config loader
func NewDefaultConfigLoader() *DefaultConfigLoader {
	fileOp := fileops.NewDefaultFileOperator()
	return &DefaultConfigLoader{
		fileOps: fileOp,
		jsonOps: fileops.NewDefaultJSONOperator(fileOp),
		yamlOps: fileops.NewDefaultYAMLOperator(fileOp),
	}
}

// Load loads configuration from multiple file paths with fallback
func (d *DefaultConfigLoader) Load(paths []string, dest interface{}) (string, error) {
	for _, path := range paths {
		if !d.fileOps.Exists(path) {
			continue
		}

		if err := d.LoadFrom(path, dest); err != nil {
			continue // Try next path
		}

		return path, nil
	}

	return "", fmt.Errorf("%w: %v", ErrNoValidConfigFile, paths)
}

// LoadFrom loads configuration from a specific file
func (d *DefaultConfigLoader) LoadFrom(path string, dest interface{}) error {
	if !d.fileOps.Exists(path) {
		return fmt.Errorf("%w: %s", ErrConfigFileNotExists, path)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		return d.yamlOps.ReadYAML(path, dest)
	case ".json":
		return d.jsonOps.ReadJSON(path, dest)
	default:
		// Try YAML first, then JSON
		if err := d.yamlOps.ReadYAML(path, dest); err == nil {
			return nil
		}
		return d.jsonOps.ReadJSON(path, dest)
	}
}

// Save saves configuration to a file in the specified format
func (d *DefaultConfigLoader) Save(path string, data interface{}, format string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := d.fileOps.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	switch strings.ToLower(format) {
	case "yaml", "yml":
		return d.yamlOps.WriteYAML(path, data)
	case "json":
		jsonData, err := d.jsonOps.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		return d.fileOps.WriteFile(path, jsonData, 0o644)
	default:
		return fmt.Errorf("%w: %s", ErrUnsupportedFormat, format)
	}
}

// Validate validates configuration data (basic implementation)
func (d *DefaultConfigLoader) Validate(data interface{}) error {
	if data == nil {
		return ErrConfigDataNil
	}
	// Basic validation - can be extended with more sophisticated validation
	return nil
}

// GetSupportedFormats returns list of supported file formats
func (d *DefaultConfigLoader) GetSupportedFormats() []string {
	return []string{"yaml", "yml", "json"}
}

// DefaultEnvProvider implements TypedEnvProvider using os package
type DefaultEnvProvider struct{}

// NewDefaultEnvProvider creates a new default environment provider
func NewDefaultEnvProvider() *DefaultEnvProvider {
	return &DefaultEnvProvider{}
}

// Get retrieves an environment variable
func (d *DefaultEnvProvider) Get(key string) string {
	return os.Getenv(key)
}

// GetWithDefault retrieves an environment variable with a default value
func (d *DefaultEnvProvider) GetWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Set sets an environment variable
func (d *DefaultEnvProvider) Set(key, value string) error {
	return os.Setenv(key, value)
}

// LookupEnv retrieves an environment variable and reports whether it was found
func (d *DefaultEnvProvider) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

// Unset removes an environment variable
func (d *DefaultEnvProvider) Unset(key string) error {
	return os.Unsetenv(key)
}

// GetAll returns all environment variables as a map
func (d *DefaultEnvProvider) GetAll() map[string]string {
	environ := os.Environ()
	result := make(map[string]string, len(environ))

	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			result[parts[0]] = parts[1]
		}
	}

	return result
}

// GetBool retrieves a boolean environment variable
func (d *DefaultEnvProvider) GetBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	switch strings.ToLower(value) {
	case "true", "1", "yes", "on", "enabled":
		return true
	case "false", "0", "no", "off", "disabled":
		return false
	default:
		return defaultValue
	}
}

// GetInt retrieves an integer environment variable
func (d *DefaultEnvProvider) GetInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}

	return defaultValue
}

// GetInt64 retrieves an int64 environment variable
func (d *DefaultEnvProvider) GetInt64(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	if int64Value, err := strconv.ParseInt(value, 10, 64); err == nil {
		return int64Value
	}

	return defaultValue
}

// GetFloat64 retrieves a float64 environment variable
func (d *DefaultEnvProvider) GetFloat64(key string, defaultValue float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}

	return defaultValue
}

// GetDuration retrieves a duration environment variable
func (d *DefaultEnvProvider) GetDuration(key string, defaultValue time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}

	return defaultValue
}

// GetStringSlice retrieves a string slice environment variable (comma-separated)
func (d *DefaultEnvProvider) GetStringSlice(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}

	if len(result) == 0 {
		return defaultValue
	}

	return result
}

// FileConfigSource implements ConfigSource for file-based configuration
type FileConfigSource struct {
	path     string
	format   Format
	priority int
	loader   Loader
}

// NewFileConfigSource creates a new file configuration source
func NewFileConfigSource(path string, format Format, priority int) *FileConfigSource {
	return &FileConfigSource{
		path:     path,
		format:   format,
		priority: priority,
		loader:   NewDefaultConfigLoader(),
	}
}

// Name returns the name of the configuration source
func (f *FileConfigSource) Name() string {
	return fmt.Sprintf("file:%s", f.path)
}

// Load loads configuration data from the source
func (f *FileConfigSource) Load(dest interface{}) error {
	return f.loader.LoadFrom(f.path, dest)
}

// IsAvailable checks if the source is available
func (f *FileConfigSource) IsAvailable() bool {
	fileOps := fileops.NewDefaultFileOperator()
	return fileOps.Exists(f.path)
}

// Priority returns the priority of this source (higher = more important)
func (f *FileConfigSource) Priority() int {
	return f.priority
}

// EnvConfigSource implements ConfigSource for environment variables
type EnvConfigSource struct {
	prefix   string
	priority int
	envOps   TypedEnvProvider
}

// NewEnvConfigSource creates a new environment configuration source
func NewEnvConfigSource(prefix string, priority int) *EnvConfigSource {
	return &EnvConfigSource{
		prefix:   prefix,
		priority: priority,
		envOps:   NewDefaultEnvProvider(),
	}
}

// Name returns the name of the configuration source
func (e *EnvConfigSource) Name() string {
	return fmt.Sprintf("env:%s", e.prefix)
}

// Load loads configuration data from environment variables
func (e *EnvConfigSource) Load(_ interface{}) error {
	// This is a simplified implementation
	// In a real implementation, this would use reflection to map env vars to struct fields
	return errEnvConfigNotImplemented
}

// IsAvailable checks if the source is available
func (e *EnvConfigSource) IsAvailable() bool {
	return true // Environment is always available
}

// Priority returns the priority of this source
func (e *EnvConfigSource) Priority() int {
	return e.priority
}
