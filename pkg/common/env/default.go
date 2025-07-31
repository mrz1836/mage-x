package env

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	osWindows = "windows"
)

// DefaultEnvironment implements Environment using os package
type DefaultEnvironment struct {
	options Options
}

// NewDefaultEnvironment creates a new default environment
func NewDefaultEnvironment() *DefaultEnvironment {
	return &DefaultEnvironment{
		options: Options{
			CaseSensitive:  runtime.GOOS != osWindows,
			AllowOverwrite: true,
			AutoTrim:       true,
			ExpandVars:     false,
		},
	}
}

// NewDefaultEnvironmentWithOptions creates a new default environment with options
func NewDefaultEnvironmentWithOptions(options Options) *DefaultEnvironment {
	return &DefaultEnvironment{options: options}
}

// Get retrieves an environment variable
func (e *DefaultEnvironment) Get(key string) string {
	value := os.Getenv(key)
	if e.options.AutoTrim {
		value = strings.TrimSpace(value)
	}
	if e.options.ExpandVars {
		value = os.ExpandEnv(value)
	}
	return value
}

// Set sets an environment variable
func (e *DefaultEnvironment) Set(key, value string) error {
	if !e.options.AllowOverwrite && e.Exists(key) {
		return fmt.Errorf("environment variable %s already exists and overwrite is not allowed", key)
	}

	if e.options.AutoTrim {
		value = strings.TrimSpace(value)
	}

	return os.Setenv(key, value)
}

// Unset removes an environment variable
func (e *DefaultEnvironment) Unset(key string) error {
	return os.Unsetenv(key)
}

// Exists checks if an environment variable exists
func (e *DefaultEnvironment) Exists(key string) bool {
	_, exists := os.LookupEnv(key)
	return exists
}

// GetString retrieves a string environment variable with default
func (e *DefaultEnvironment) GetString(key, defaultValue string) string {
	if value := e.Get(key); value != "" {
		return value
	}
	return defaultValue
}

// GetWithDefault is an alias for GetString for compatibility
func (e *DefaultEnvironment) GetWithDefault(key, defaultValue string) string {
	return e.GetString(key, defaultValue)
}

// GetBool retrieves a boolean environment variable with default
func (e *DefaultEnvironment) GetBool(key string, defaultValue bool) bool {
	value := e.Get(key)
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

// GetInt retrieves an integer environment variable with default
func (e *DefaultEnvironment) GetInt(key string, defaultValue int) int {
	value := e.Get(key)
	if value == "" {
		return defaultValue
	}

	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}

	return defaultValue
}

// GetInt64 retrieves an int64 environment variable with default
func (e *DefaultEnvironment) GetInt64(key string, defaultValue int64) int64 {
	value := e.Get(key)
	if value == "" {
		return defaultValue
	}

	if int64Value, err := strconv.ParseInt(value, 10, 64); err == nil {
		return int64Value
	}

	return defaultValue
}

// GetFloat64 retrieves a float64 environment variable with default
func (e *DefaultEnvironment) GetFloat64(key string, defaultValue float64) float64 {
	value := e.Get(key)
	if value == "" {
		return defaultValue
	}

	if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
		return floatValue
	}

	return defaultValue
}

// GetDuration retrieves a duration environment variable with default
func (e *DefaultEnvironment) GetDuration(key string, defaultValue time.Duration) time.Duration {
	value := e.Get(key)
	if value == "" {
		return defaultValue
	}

	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}

	return defaultValue
}

// GetStringSlice retrieves a string slice environment variable (comma-separated)
func (e *DefaultEnvironment) GetStringSlice(key string, defaultValue []string) []string {
	value := e.Get(key)
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

// GetWithPrefix retrieves all environment variables with a given prefix
func (e *DefaultEnvironment) GetWithPrefix(prefix string) map[string]string {
	environ := os.Environ()
	result := make(map[string]string)

	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			if strings.HasPrefix(key, prefix) {
				value := parts[1]
				if e.options.AutoTrim {
					value = strings.TrimSpace(value)
				}
				result[key] = value
			}
		}
	}

	return result
}

// SetMultiple sets multiple environment variables
func (e *DefaultEnvironment) SetMultiple(vars map[string]string) error {
	for key, value := range vars {
		if err := e.Set(key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}
	return nil
}

// GetAll returns all environment variables
func (e *DefaultEnvironment) GetAll() map[string]string {
	environ := os.Environ()
	result := make(map[string]string, len(environ))

	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]
			if e.options.AutoTrim {
				value = strings.TrimSpace(value)
			}
			result[key] = value
		}
	}

	return result
}

// Clear removes all environment variables (dangerous - use with caution)
func (e *DefaultEnvironment) Clear() error {
	environ := os.Environ()
	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) >= 1 {
			if err := os.Unsetenv(parts[0]); err != nil {
				return err
			}
		}
	}
	return nil
}

// Validate validates an environment variable value
func (e *DefaultEnvironment) Validate(key string, validator func(string) bool) bool {
	value := e.Get(key)
	return validator(value)
}

// Required checks that required environment variables are set
func (e *DefaultEnvironment) Required(keys ...string) error {
	var missing []string

	for _, key := range keys {
		if !e.Exists(key) || e.Get(key) == "" {
			missing = append(missing, key)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("required environment variables not set: %v", missing)
	}

	return nil
}

// DefaultPathResolver implements PathResolver using os and filepath packages
type DefaultPathResolver struct {
	options PathOptions
}

// NewDefaultPathResolver creates a new default path resolver
func NewDefaultPathResolver() *DefaultPathResolver {
	return &DefaultPathResolver{
		options: PathOptions{
			CreateMissing:  false,
			Mode:           0o755,
			FollowSymlinks: true,
			ResolveEnvVars: true,
		},
	}
}

// NewDefaultPathResolverWithOptions creates a new path resolver with options
func NewDefaultPathResolverWithOptions(options PathOptions) *DefaultPathResolver {
	return &DefaultPathResolver{options: options}
}

// Home returns the user's home directory
func (p *DefaultPathResolver) Home() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}

	// Windows fallback
	if home := os.Getenv("USERPROFILE"); home != "" {
		return home
	}

	// Another Windows fallback
	if drive := os.Getenv("HOMEDRIVE"); drive != "" {
		if path := os.Getenv("HOMEPATH"); path != "" {
			return drive + path
		}
	}

	return ""
}

// ConfigDir returns the configuration directory for an application
func (p *DefaultPathResolver) ConfigDir(appName string) string {
	if configDir := os.Getenv("XDG_CONFIG_HOME"); configDir != "" {
		return filepath.Join(configDir, appName)
	}

	home := p.Home()
	if home == "" {
		return ""
	}

	switch runtime.GOOS {
	case osWindows:
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, appName)
		}
		return filepath.Join(home, "AppData", "Roaming", appName)
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", appName)
	default:
		return filepath.Join(home, ".config", appName)
	}
}

// DataDir returns the data directory for an application
func (p *DefaultPathResolver) DataDir(appName string) string {
	if dataDir := os.Getenv("XDG_DATA_HOME"); dataDir != "" {
		return filepath.Join(dataDir, appName)
	}

	home := p.Home()
	if home == "" {
		return ""
	}

	switch runtime.GOOS {
	case osWindows:
		return p.ConfigDir(appName) // Same as config on Windows
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", appName)
	default:
		return filepath.Join(home, ".local", "share", appName)
	}
}

// CacheDir returns the cache directory for an application
func (p *DefaultPathResolver) CacheDir(appName string) string {
	if cacheDir := os.Getenv("XDG_CACHE_HOME"); cacheDir != "" {
		return filepath.Join(cacheDir, appName)
	}

	home := p.Home()
	if home == "" {
		return ""
	}

	switch runtime.GOOS {
	case osWindows:
		if temp := os.Getenv("TEMP"); temp != "" {
			return filepath.Join(temp, appName)
		}
		return filepath.Join(home, "AppData", "Local", "Temp", appName)
	case "darwin":
		return filepath.Join(home, "Library", "Caches", appName)
	default:
		return filepath.Join(home, ".cache", appName)
	}
}

// TempDir returns the temporary directory
func (p *DefaultPathResolver) TempDir() string {
	return os.TempDir()
}

// WorkingDir returns the current working directory
func (p *DefaultPathResolver) WorkingDir() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return ""
}

// GOPATH returns the GOPATH environment variable or default
func (p *DefaultPathResolver) GOPATH() string {
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		return gopath
	}

	home := p.Home()
	if home == "" {
		return ""
	}

	return filepath.Join(home, "go")
}

// GOROOT returns the GOROOT environment variable
func (p *DefaultPathResolver) GOROOT() string {
	if goroot := os.Getenv("GOROOT"); goroot != "" {
		return goroot
	}
	// Use 'go env GOROOT' instead of deprecated runtime.GOROOT
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "env", "GOROOT")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}
	// Fallback to empty string if go command fails
	return ""
}

// GOCACHE returns the Go build cache directory
func (p *DefaultPathResolver) GOCACHE() string {
	if cache := os.Getenv("GOCACHE"); cache != "" {
		return cache
	}

	// Default cache location varies by OS
	switch runtime.GOOS {
	case osWindows:
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			return filepath.Join(localAppData, "go-build")
		}
	case "darwin":
		home := p.Home()
		if home != "" {
			return filepath.Join(home, "Library", "Caches", "go-build")
		}
	default:
		home := p.Home()
		if home != "" {
			return filepath.Join(home, ".cache", "go-build")
		}
	}

	return ""
}

// GOMODCACHE returns the Go module cache directory
func (p *DefaultPathResolver) GOMODCACHE() string {
	if cache := os.Getenv("GOMODCACHE"); cache != "" {
		return cache
	}

	return filepath.Join(p.GOPATH(), "pkg", "mod")
}

// Expand expands environment variables in a path
func (p *DefaultPathResolver) Expand(path string) string {
	if !p.options.ResolveEnvVars {
		return path
	}

	// Handle tilde expansion
	if strings.HasPrefix(path, "~/") {
		if home := p.Home(); home != "" {
			path = filepath.Join(home, path[2:])
		}
	}

	// Expand environment variables
	return os.ExpandEnv(path)
}

// Resolve resolves a path to its absolute form
func (p *DefaultPathResolver) Resolve(path string) (string, error) {
	expanded := p.Expand(path)

	if p.options.FollowSymlinks {
		if resolved, err := filepath.EvalSymlinks(expanded); err == nil {
			expanded = resolved
		}
	}

	abs, err := filepath.Abs(expanded)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path %s: %w", path, err)
	}

	return abs, nil
}

// IsAbsolute checks if a path is absolute
func (p *DefaultPathResolver) IsAbsolute(path string) bool {
	return filepath.IsAbs(path)
}

// MakeAbsolute converts a relative path to absolute
func (p *DefaultPathResolver) MakeAbsolute(path string) (string, error) {
	if p.IsAbsolute(path) {
		return path, nil
	}

	return p.Resolve(path)
}

// Clean cleans a path
func (p *DefaultPathResolver) Clean(path string) string {
	return filepath.Clean(path)
}

// EnsureDir ensures a directory exists
func (p *DefaultPathResolver) EnsureDir(path string) error {
	return p.EnsureDirWithMode(path, p.options.Mode)
}

// EnsureDirWithMode ensures a directory exists with specific mode
func (p *DefaultPathResolver) EnsureDirWithMode(path string, mode uint32) error {
	expanded := p.Expand(path)

	if _, err := os.Stat(expanded); os.IsNotExist(err) {
		return os.MkdirAll(expanded, os.FileMode(mode))
	}

	return nil
}
