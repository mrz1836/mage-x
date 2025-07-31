// Package env provides environment variable and path resolution utilities with mockable interfaces
package env

import (
	"time"
)

// Environment provides advanced environment variable operations
type Environment interface {
	// Basic operations
	Get(key string) string
	Set(key, value string) error
	Unset(key string) error
	Exists(key string) bool

	// Type-safe operations
	GetString(key, defaultValue string) string
	GetWithDefault(key, defaultValue string) string // Alias for GetString for compatibility
	GetBool(key string, defaultValue bool) bool
	GetInt(key string, defaultValue int) int
	GetInt64(key string, defaultValue int64) int64
	GetFloat64(key string, defaultValue float64) float64
	GetDuration(key string, defaultValue time.Duration) time.Duration
	GetStringSlice(key string, defaultValue []string) []string

	// Advanced operations
	GetWithPrefix(prefix string) map[string]string
	SetMultiple(vars map[string]string) error
	GetAll() map[string]string
	Clear() error

	// Validation
	Validate(key string, validator func(string) bool) bool
	Required(keys ...string) error
}

// PathResolver provides path resolution and manipulation utilities
type PathResolver interface {
	// Standard paths
	Home() string
	ConfigDir(appName string) string
	DataDir(appName string) string
	CacheDir(appName string) string
	TempDir() string
	WorkingDir() string

	// Go-specific paths
	GOPATH() string
	GOROOT() string
	GOCACHE() string
	GOMODCACHE() string

	// Path operations
	Expand(path string) string
	Resolve(path string) (string, error)
	IsAbsolute(path string) bool
	MakeAbsolute(path string) (string, error)
	Clean(path string) string

	// Directory operations
	EnsureDir(path string) error
	EnsureDirWithMode(path string, mode uint32) error
}

// EnvManager manages environment variable scopes and contexts
type EnvManager interface {
	// Scope management
	PushScope() EnvScope
	PopScope() error
	WithScope(fn func(EnvScope) error) error

	// Context operations
	SaveContext() (EnvContext, error)
	RestoreContext(ctx EnvContext) error

	// Isolation
	Isolate(vars map[string]string, fn func() error) error
	Fork() EnvManager
}

// EnvScope represents a scoped environment context
type EnvScope interface {
	Environment

	// Scope-specific operations
	Commit() error
	Rollback() error
	Changes() map[string]EnvChange
	HasChanges() bool
}

// EnvContext represents a saved environment state
type EnvContext interface {
	// Metadata
	Timestamp() time.Time
	Variables() map[string]string
	Count() int

	// Operations
	Diff(other EnvContext) map[string]EnvChange
	Merge(other EnvContext) EnvContext
	Export() map[string]string
}

// EnvValidator provides environment variable validation
type EnvValidator interface {
	// Validation rules
	AddRule(key string, rule ValidationRule) error
	RemoveRule(key string) error
	ValidateAll() []ValidationError
	Validate(key, value string) error

	// Built-in validators
	Required(keys ...string) EnvValidator
	NotEmpty(keys ...string) EnvValidator
	Pattern(key, pattern string) EnvValidator
	Range(key string, minValue, maxValue interface{}) EnvValidator
	OneOf(key string, values ...string) EnvValidator
}

// Supporting types

// EnvChange represents a change to an environment variable
type EnvChange struct {
	Key      string
	OldValue string
	NewValue string
	Action   ChangeAction
}

// ChangeAction represents the type of change
type ChangeAction int

const (
	ActionSet ChangeAction = iota
	ActionUnset
	ActionModify
)

// ValidationRule represents a validation rule for environment variables
type ValidationRule interface {
	Validate(value string) error
	Description() string
}

// ValidationError represents a validation error
type ValidationError struct {
	Key     string
	Value   string
	Rule    string
	Message string
}

// PathOptions contains options for path operations
type PathOptions struct {
	// CreateMissing creates missing directories
	CreateMissing bool

	// Mode is the file mode for created directories
	Mode uint32

	// FollowSymlinks follows symbolic links when resolving
	FollowSymlinks bool

	// ResolveEnvVars expands environment variables in paths
	ResolveEnvVars bool
}

// EnvOptions contains options for environment operations
type EnvOptions struct {
	// CaseSensitive determines if variable names are case sensitive
	CaseSensitive bool

	// AllowOverwrite allows overwriting existing variables
	AllowOverwrite bool

	// AutoTrim trims whitespace from values
	AutoTrim bool

	// ExpandVars expands variables in values (e.g., $HOME/path)
	ExpandVars bool

	// Prefix for scoped operations
	Prefix string
}
