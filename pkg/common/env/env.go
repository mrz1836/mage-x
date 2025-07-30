// Package env provides environment variable and path resolution utilities with mockable interfaces
package env

import "time"

// NewEnvironment creates a new default environment instance
func NewEnvironment() Environment {
	return NewDefaultEnvironment()
}

// NewOSEnvironment creates a new OS environment instance (alias for compatibility)
func NewOSEnvironment() Environment {
	return NewDefaultEnvironment()
}

// Global default instances
var (
	DefaultEnv      Environment  = NewDefaultEnvironment()  //nolint:gochecknoglobals // Package-level default
	DefaultPaths    PathResolver = NewDefaultPathResolver() //nolint:gochecknoglobals // Package-level default
	DefaultManager  EnvManager   = NewDefaultEnvManager()   //nolint:gochecknoglobals // Package-level default
	DefaultValidate EnvValidator = NewDefaultEnvValidator() //nolint:gochecknoglobals // Package-level default
)

// Environment variable convenience functions

// Get retrieves an environment variable using the default environment
func Get(key string) string {
	return DefaultEnv.Get(key)
}

// Set sets an environment variable using the default environment
func Set(key, value string) error {
	return DefaultEnv.Set(key, value)
}

// Unset removes an environment variable using the default environment
func Unset(key string) error {
	return DefaultEnv.Unset(key)
}

// Exists checks if an environment variable exists using the default environment
func Exists(key string) bool {
	return DefaultEnv.Exists(key)
}

// GetString retrieves a string environment variable with default
func GetString(key, defaultValue string) string {
	return DefaultEnv.GetString(key, defaultValue)
}

// GetBool retrieves a boolean environment variable with default
func GetBool(key string, defaultValue bool) bool {
	return DefaultEnv.GetBool(key, defaultValue)
}

// GetInt retrieves an integer environment variable with default
func GetInt(key string, defaultValue int) int {
	return DefaultEnv.GetInt(key, defaultValue)
}

// GetInt64 retrieves an int64 environment variable with default
func GetInt64(key string, defaultValue int64) int64 {
	return DefaultEnv.GetInt64(key, defaultValue)
}

// GetFloat64 retrieves a float64 environment variable with default
func GetFloat64(key string, defaultValue float64) float64 {
	return DefaultEnv.GetFloat64(key, defaultValue)
}

// GetDuration retrieves a duration environment variable with default
func GetDuration(key string, defaultValue time.Duration) time.Duration {
	return DefaultEnv.GetDuration(key, defaultValue)
}

// GetStringSlice retrieves a string slice environment variable (comma-separated)
func GetStringSlice(key string, defaultValue []string) []string {
	return DefaultEnv.GetStringSlice(key, defaultValue)
}

// GetWithPrefix retrieves all environment variables with a given prefix
func GetWithPrefix(prefix string) map[string]string {
	return DefaultEnv.GetWithPrefix(prefix)
}

// SetMultiple sets multiple environment variables
func SetMultiple(vars map[string]string) error {
	return DefaultEnv.SetMultiple(vars)
}

// Required checks that required environment variables are set
func Required(keys ...string) error {
	return DefaultEnv.Required(keys...)
}

// Path convenience functions

// Home returns the user's home directory
func Home() string {
	return DefaultPaths.Home()
}

// ConfigDir returns the configuration directory for an application
func ConfigDir(appName string) string {
	return DefaultPaths.ConfigDir(appName)
}

// DataDir returns the data directory for an application
func DataDir(appName string) string {
	return DefaultPaths.DataDir(appName)
}

// CacheDir returns the cache directory for an application
func CacheDir(appName string) string {
	return DefaultPaths.CacheDir(appName)
}

// TempDir returns the temporary directory
func TempDir() string {
	return DefaultPaths.TempDir()
}

// WorkingDir returns the current working directory
func WorkingDir() string {
	return DefaultPaths.WorkingDir()
}

// GOPATH returns the GOPATH environment variable or default
func GOPATH() string {
	return DefaultPaths.GOPATH()
}

// GOROOT returns the GOROOT environment variable
func GOROOT() string {
	return DefaultPaths.GOROOT()
}

// GOCACHE returns the Go build cache directory
func GOCACHE() string {
	return DefaultPaths.GOCACHE()
}

// GOMODCACHE returns the Go module cache directory
func GOMODCACHE() string {
	return DefaultPaths.GOMODCACHE()
}

// Expand expands environment variables in a path
func Expand(path string) string {
	return DefaultPaths.Expand(path)
}

// Resolve resolves a path to its absolute form
func Resolve(path string) (string, error) {
	return DefaultPaths.Resolve(path)
}

// IsAbsolute checks if a path is absolute
func IsAbsolute(path string) bool {
	return DefaultPaths.IsAbsolute(path)
}

// MakeAbsolute converts a relative path to absolute
func MakeAbsolute(path string) (string, error) {
	return DefaultPaths.MakeAbsolute(path)
}

// Clean cleans a path
func Clean(path string) string {
	return DefaultPaths.Clean(path)
}

// EnsureDir ensures a directory exists
func EnsureDir(path string) error {
	return DefaultPaths.EnsureDir(path)
}

// EnsureDirWithMode ensures a directory exists with specific mode
func EnsureDirWithMode(path string, mode uint32) error {
	return DefaultPaths.EnsureDirWithMode(path, mode)
}

// Manager convenience functions

// WithScope executes a function within a new environment scope
func WithScope(fn func(EnvScope) error) error {
	return DefaultManager.WithScope(fn)
}

// SaveContext saves the current environment state
func SaveContext() (EnvContext, error) {
	return DefaultManager.SaveContext()
}

// RestoreContext restores a saved environment state
func RestoreContext(ctx EnvContext) error {
	return DefaultManager.RestoreContext(ctx)
}

// Isolate runs a function with isolated environment variables
func Isolate(vars map[string]string, fn func() error) error {
	return DefaultManager.Isolate(vars, fn)
}

// Configuration functions

// SetEnvironment sets the global environment instance (for testing/mocking)
func SetEnvironment(env Environment) {
	DefaultEnv = env
}

// SetPathResolver sets the global path resolver instance (for testing/mocking)
func SetPathResolver(resolver PathResolver) {
	DefaultPaths = resolver
}

// SetManager sets the global manager instance (for testing/mocking)
func SetManager(manager EnvManager) {
	DefaultManager = manager
}

// SetValidator sets the global validator instance (for testing/mocking)
func SetValidator(validator EnvValidator) {
	DefaultValidate = validator
}

// Advanced utility functions

// LoadFromFile loads environment variables from a file
func LoadFromFile(filename string) error {
	// This would integrate with the config package for file loading
	// For now, return a basic implementation
	return nil
}

// SaveToFile saves current environment variables to a file
func SaveToFile(filename string, prefix string) error {
	// This would integrate with the config package for file saving
	// For now, return a basic implementation
	return nil
}

// Backup creates a backup of the current environment
func Backup() EnvContext {
	ctx, _ := SaveContext()
	return ctx
}

// Restore restores from a backup
func Restore(backup EnvContext) error {
	return RestoreContext(backup)
}

// GetAllWithPrefix returns all environment variables with a prefix
func GetAllWithPrefix(prefix string) map[string]string {
	return GetWithPrefix(prefix)
}

// SetFromMap sets environment variables from a map
func SetFromMap(vars map[string]string) error {
	return SetMultiple(vars)
}

// ClearPrefix removes all environment variables with a prefix
func ClearPrefix(prefix string) error {
	vars := GetWithPrefix(prefix)
	for key := range vars {
		if err := Unset(key); err != nil {
			return err
		}
	}
	return nil
}
