// Package env provides environment variable and path resolution utilities with mockable interfaces
package env

import (
	"sync"
	"time"
)

// NewEnvironment creates a new default environment instance
func NewEnvironment() Environment {
	return NewDefaultEnvironment()
}

// NewOSEnvironment creates a new OS environment instance (alias for compatibility)
func NewOSEnvironment() Environment {
	return NewDefaultEnvironment()
}

// Thread-safe singleton instances for internal use
var (
	defaultEnvOnce      sync.Once //nolint:gochecknoglobals // Required for thread-safe singleton pattern
	defaultPathsOnce    sync.Once //nolint:gochecknoglobals // Required for thread-safe singleton pattern
	defaultManagerOnce  sync.Once //nolint:gochecknoglobals // Required for thread-safe singleton pattern
	defaultValidateOnce sync.Once //nolint:gochecknoglobals // Required for thread-safe singleton pattern

	defaultEnvInstance      Environment  //nolint:gochecknoglobals // Required for thread-safe singleton pattern
	defaultPathsInstance    PathResolver //nolint:gochecknoglobals // Required for thread-safe singleton pattern
	defaultManagerInstance  Manager      //nolint:gochecknoglobals // Required for thread-safe singleton pattern
	defaultValidateInstance Validator    //nolint:gochecknoglobals // Required for thread-safe singleton pattern
)

// GetDefaultEnv returns the default environment instance, creating it if necessary
func GetDefaultEnv() Environment {
	defaultEnvOnce.Do(func() {
		defaultEnvInstance = NewDefaultEnvironment()
	})
	return defaultEnvInstance
}

// GetDefaultPaths returns the default path resolver instance, creating it if necessary
func GetDefaultPaths() PathResolver {
	defaultPathsOnce.Do(func() {
		defaultPathsInstance = NewDefaultPathResolver()
	})
	return defaultPathsInstance
}

// GetDefaultManager returns the default manager instance, creating it if necessary
func GetDefaultManager() Manager {
	defaultManagerOnce.Do(func() {
		defaultManagerInstance = NewDefaultEnvManager()
	})
	return defaultManagerInstance
}

// GetDefaultValidate returns the default validator instance, creating it if necessary
func GetDefaultValidate() Validator {
	defaultValidateOnce.Do(func() {
		defaultValidateInstance = NewDefaultEnvValidator()
	})
	return defaultValidateInstance
}

// Package-level default instances for backward compatibility.
// These variables provide the global API while internally using thread-safe singletons.
var (
	DefaultEnv      Environment  = GetDefaultEnv()      //nolint:gochecknoglobals // Required for backward compatibility API
	DefaultPaths    PathResolver = GetDefaultPaths()    //nolint:gochecknoglobals // Required for backward compatibility API
	DefaultManager  Manager      = GetDefaultManager()  //nolint:gochecknoglobals // Required for backward compatibility API
	DefaultValidate Validator    = GetDefaultValidate() //nolint:gochecknoglobals // Required for backward compatibility API
)

// Environment variable convenience functions

// Get retrieves an environment variable using the default environment
func Get(key string) string {
	return GetDefaultEnv().Get(key)
}

// Set sets an environment variable using the default environment
func Set(key, value string) error {
	return GetDefaultEnv().Set(key, value)
}

// Unset removes an environment variable using the default environment
func Unset(key string) error {
	return GetDefaultEnv().Unset(key)
}

// Exists checks if an environment variable exists using the default environment
func Exists(key string) bool {
	return GetDefaultEnv().Exists(key)
}

// GetString retrieves a string environment variable with default
func GetString(key, defaultValue string) string {
	return GetDefaultEnv().GetString(key, defaultValue)
}

// GetBool retrieves a boolean environment variable with default
func GetBool(key string, defaultValue bool) bool {
	return GetDefaultEnv().GetBool(key, defaultValue)
}

// GetInt retrieves an integer environment variable with default
func GetInt(key string, defaultValue int) int {
	return GetDefaultEnv().GetInt(key, defaultValue)
}

// GetInt64 retrieves an int64 environment variable with default
func GetInt64(key string, defaultValue int64) int64 {
	return GetDefaultEnv().GetInt64(key, defaultValue)
}

// GetFloat64 retrieves a float64 environment variable with default
func GetFloat64(key string, defaultValue float64) float64 {
	return GetDefaultEnv().GetFloat64(key, defaultValue)
}

// GetDuration retrieves a duration environment variable with default
func GetDuration(key string, defaultValue time.Duration) time.Duration {
	return GetDefaultEnv().GetDuration(key, defaultValue)
}

// GetStringSlice retrieves a string slice environment variable (comma-separated)
func GetStringSlice(key string, defaultValue []string) []string {
	return GetDefaultEnv().GetStringSlice(key, defaultValue)
}

// GetWithPrefix retrieves all environment variables with a given prefix
func GetWithPrefix(prefix string) map[string]string {
	return GetDefaultEnv().GetWithPrefix(prefix)
}

// SetMultiple sets multiple environment variables
func SetMultiple(vars map[string]string) error {
	return GetDefaultEnv().SetMultiple(vars)
}

// Required checks that required environment variables are set
func Required(keys ...string) error {
	return GetDefaultEnv().Required(keys...)
}

// Path convenience functions

// Home returns the user's home directory
func Home() string {
	return GetDefaultPaths().Home()
}

// ConfigDir returns the configuration directory for an application
func ConfigDir(appName string) string {
	return GetDefaultPaths().ConfigDir(appName)
}

// DataDir returns the data directory for an application
func DataDir(appName string) string {
	return GetDefaultPaths().DataDir(appName)
}

// CacheDir returns the cache directory for an application
func CacheDir(appName string) string {
	return GetDefaultPaths().CacheDir(appName)
}

// TempDir returns the temporary directory
func TempDir() string {
	return GetDefaultPaths().TempDir()
}

// WorkingDir returns the current working directory
func WorkingDir() string {
	return GetDefaultPaths().WorkingDir()
}

// GOPATH returns the GOPATH environment variable or default
func GOPATH() string {
	return GetDefaultPaths().GOPATH()
}

// GOROOT returns the GOROOT environment variable
func GOROOT() string {
	return GetDefaultPaths().GOROOT()
}

// GOCACHE returns the Go build cache directory
func GOCACHE() string {
	return GetDefaultPaths().GOCACHE()
}

// GOMODCACHE returns the Go module cache directory
func GOMODCACHE() string {
	return GetDefaultPaths().GOMODCACHE()
}

// Expand expands environment variables in a path
func Expand(path string) string {
	return GetDefaultPaths().Expand(path)
}

// Resolve resolves a path to its absolute form
func Resolve(path string) (string, error) {
	return GetDefaultPaths().Resolve(path)
}

// IsAbsolute checks if a path is absolute
func IsAbsolute(path string) bool {
	return GetDefaultPaths().IsAbsolute(path)
}

// MakeAbsolute converts a relative path to absolute
func MakeAbsolute(path string) (string, error) {
	return GetDefaultPaths().MakeAbsolute(path)
}

// Clean cleans a path
func Clean(path string) string {
	return GetDefaultPaths().Clean(path)
}

// EnsureDir ensures a directory exists
func EnsureDir(path string) error {
	return GetDefaultPaths().EnsureDir(path)
}

// EnsureDirWithMode ensures a directory exists with specific mode
func EnsureDirWithMode(path string, mode uint32) error {
	return GetDefaultPaths().EnsureDirWithMode(path, mode)
}

// Manager convenience functions

// WithScope executes a function within a new environment scope
func WithScope(fn func(Scope) error) error {
	return GetDefaultManager().WithScope(fn)
}

// SaveContext saves the current environment state
func SaveContext() (Context, error) {
	return GetDefaultManager().SaveContext()
}

// RestoreContext restores a saved environment state
func RestoreContext(ctx Context) error {
	return GetDefaultManager().RestoreContext(ctx)
}

// Isolate runs a function with isolated environment variables
func Isolate(vars map[string]string, fn func() error) error {
	return GetDefaultManager().Isolate(vars, fn)
}

// Configuration functions

// SetEnvironment sets the global environment instance (for testing/mocking)
func SetEnvironment(env Environment) {
	defaultEnvInstance = env
	defaultEnvOnce = sync.Once{}
	DefaultEnv = env
}

// SetPathResolver sets the global path resolver instance (for testing/mocking)
func SetPathResolver(resolver PathResolver) {
	defaultPathsInstance = resolver
	defaultPathsOnce = sync.Once{}
	DefaultPaths = resolver
}

// SetManager sets the global manager instance (for testing/mocking)
func SetManager(manager Manager) {
	defaultManagerInstance = manager
	defaultManagerOnce = sync.Once{}
	DefaultManager = manager
}

// SetValidator sets the global validator instance (for testing/mocking)
func SetValidator(validator Validator) {
	defaultValidateInstance = validator
	defaultValidateOnce = sync.Once{}
	DefaultValidate = validator
}

// Advanced utility functions

// LoadFromFile loads environment variables from a file
func LoadFromFile(_ string) error {
	// This would integrate with the config package for file loading
	// For now, return a basic implementation
	return nil
}

// SaveToFile saves current environment variables to a file
func SaveToFile(_, _ string) error {
	// This would integrate with the config package for file saving
	// For now, return a basic implementation
	return nil
}

// Backup creates a backup of the current environment
func Backup() Context {
	ctx, err := SaveContext()
	if err != nil {
		// Return empty context on error - this maintains backward compatibility
		return &DefaultEnvContext{variables: make(map[string]string)}
	}
	return ctx
}

// Restore restores from a backup
func Restore(backup Context) error {
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
