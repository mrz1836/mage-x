// Package paths provides advanced path building and manipulation utilities with mockable interfaces
package paths

import (
	"io/fs"
	"time"
)

// PathBuilder provides advanced path construction and manipulation
type PathBuilder interface {
	// Basic operations
	Join(elements ...string) PathBuilder
	Dir() PathBuilder
	Base() string
	Ext() string
	Clean() PathBuilder
	Abs() (PathBuilder, error)

	// Path operations
	Append(suffix string) PathBuilder
	Prepend(prefix string) PathBuilder
	WithExt(ext string) PathBuilder
	WithoutExt() PathBuilder
	WithName(name string) PathBuilder

	// Relative operations
	Rel(basepath string) (PathBuilder, error)
	RelTo(target PathBuilder) (PathBuilder, error)

	// Information
	String() string
	IsAbs() bool
	IsDir() bool
	IsFile() bool
	Exists() bool
	Size() int64
	ModTime() time.Time
	Mode() fs.FileMode

	// Directory operations
	Walk(fn WalkFunc) error
	List() ([]PathBuilder, error)
	ListFiles() ([]PathBuilder, error)
	ListDirs() ([]PathBuilder, error)
	Glob(pattern string) ([]PathBuilder, error)

	// Validation
	Validate() error
	IsValid() bool
	IsEmpty() bool
	IsSafe() bool

	// Modification
	Create() error
	CreateDir() error
	CreateDirAll() error
	Remove() error
	RemoveAll() error
	Copy(dest PathBuilder) error
	Move(dest PathBuilder) error

	// Links
	Readlink() (PathBuilder, error)
	Symlink(target PathBuilder) error

	// Matching
	Match(pattern string) bool
	Contains(sub string) bool
	HasPrefix(prefix string) bool
	HasSuffix(suffix string) bool

	// Cloning
	Clone() PathBuilder
}

// PathMatcher provides pattern matching capabilities
type PathMatcher interface {
	// Pattern matching
	Match(path string) bool
	MatchPath(path PathBuilder) bool
	Compile(pattern string) error
	Pattern() string

	// Multiple patterns
	AddPattern(pattern string) error
	RemovePattern(pattern string) error
	ClearPatterns() error
	Patterns() []string

	// Matching options
	SetCaseSensitive(sensitive bool) PathMatcher
	SetRecursive(recursive bool) PathMatcher
	SetMaxDepth(depth int) PathMatcher

	// Advanced matching
	MatchAny(paths ...string) bool
	MatchAll(paths ...string) bool
	Filter(paths []string) []string
	FilterPaths(paths []PathBuilder) []PathBuilder
}

// PathWatcher provides file system watching capabilities
type PathWatcher interface {
	// Watching operations
	Watch(path string, events EventMask) error
	WatchPath(path PathBuilder, events EventMask) error
	Unwatch(path string) error
	UnwatchPath(path PathBuilder) error

	// Event handling
	Events() <-chan *PathEvent
	Errors() <-chan error
	Close() error

	// Configuration
	SetBufferSize(size int) PathWatcher
	SetRecursive(recursive bool) PathWatcher
	SetDebounce(duration time.Duration) PathWatcher

	// Status
	IsWatching(path string) bool
	WatchedPaths() []string
}

// PathCache provides caching for path operations
type PathCache interface {
	// Cache operations
	Get(key string) (PathBuilder, bool)
	Set(key string, path PathBuilder) error
	Delete(key string) error
	Clear() error

	// Cache information
	Size() int
	Keys() []string
	Stats() CacheStats

	// Configuration
	SetMaxSize(size int) PathCache
	SetTTL(ttl time.Duration) PathCache
	SetEvictionPolicy(policy EvictionPolicy) PathCache

	// Validation
	Validate(key string) error
	Refresh(key string) error
	RefreshAll() error
}

// PathValidator provides path validation capabilities
type PathValidator interface {
	// Validation rules
	AddRule(rule ValidationRule) error
	RemoveRule(name string) error
	ClearRules() error
	Rules() []ValidationRule

	// Validation operations
	Validate(path string) []ValidationError
	ValidatePath(path PathBuilder) []ValidationError
	IsValid(path string) bool
	IsValidPath(path PathBuilder) bool

	// Built-in validators
	RequireAbsolute() PathValidator
	RequireRelative() PathValidator
	RequireExists() PathValidator
	RequireNotExists() PathValidator
	RequireReadable() PathValidator
	RequireWritable() PathValidator
	RequireExecutable() PathValidator
	RequireDirectory() PathValidator
	RequireFile() PathValidator
	RequireExtension(exts ...string) PathValidator
	RequireMaxLength(length int) PathValidator
	RequirePattern(pattern string) PathValidator
	ForbidPattern(pattern string) PathValidator
}

// Supporting types and interfaces

// WalkFunc is the type of function called for each file or directory visited by Walk
type WalkFunc func(path PathBuilder, info fs.FileInfo, err error) error

// PathEvent represents a file system event
type PathEvent struct {
	Path   string
	Op     EventMask
	Time   time.Time
	Info   fs.FileInfo
	Source string
}

// EventMask represents the type of file system events to watch
type EventMask uint32

const (
	// EventCreate represents a file creation event
	EventCreate EventMask = 1 << iota
	// EventWrite represents a file write event
	EventWrite
	// EventRemove represents a file removal event
	EventRemove
	// EventRename represents a file rename event
	EventRename
	// EventChmod represents a file permission change event
	EventChmod
	// EventAll represents all file system events combined
	EventAll = EventCreate | EventWrite | EventRemove | EventRename | EventChmod
)

// ValidationRule represents a path validation rule
type ValidationRule interface {
	Name() string
	Description() string
	Validate(path string) error
	ValidatePath(path PathBuilder) error
}

// ValidationError represents a path validation error
type ValidationError struct {
	Path    string
	Rule    string
	Message string
	Code    string
}

// Error implements the error interface
func (ve *ValidationError) Error() string {
	if ve.Message != "" {
		return ve.Message
	}
	return "path validation error"
}

// CacheStats provides cache statistics
type CacheStats struct {
	Size      int
	Hits      int64
	Misses    int64
	Evictions int64
	MaxSize   int
	TTL       time.Duration
}

// EvictionPolicy represents cache eviction policies
type EvictionPolicy int

const (
	// EvictLRU represents Least Recently Used eviction policy
	EvictLRU EvictionPolicy = iota
	// EvictLFU represents Least Frequently Used eviction policy
	EvictLFU
	// EvictFIFO represents First In First Out eviction policy
	EvictFIFO
	// EvictRandom represents random eviction policy
	EvictRandom
	// EvictTTL represents Time To Live eviction policy
	EvictTTL
)

// PathOptions contains options for path operations
type PathOptions struct {
	// Creation options
	CreateMode        fs.FileMode
	CreateParents     bool
	OverwriteExisting bool

	// Copy/Move options
	PreserveMtime     bool
	PreserveMode      bool
	PreserveOwnership bool
	FollowSymlinks    bool

	// Validation options
	ValidateOnCreate bool
	ValidateOnAccess bool
	StrictValidation bool

	// Performance options
	UseCache   bool
	CacheSize  int
	BufferSize int

	// Security options
	RestrictToBasePath string
	AllowUnsafePaths   bool
	MaxPathLength      int
}

// PathInfo provides extended information about a path
type PathInfo struct {
	Path         string
	AbsPath      string
	Size         int64
	Mode         fs.FileMode
	ModTime      time.Time
	IsDir        bool
	IsFile       bool
	IsSymlink    bool
	IsExecutable bool
	IsReadable   bool
	IsWritable   bool
	Owner        string
	Group        string
	Checksum     string
	MimeType     string
}

// PathSet provides set operations for paths
type PathSet interface {
	// Set operations
	Add(path string) bool
	AddPath(path PathBuilder) bool
	Remove(path string) bool
	RemovePath(path PathBuilder) bool
	Contains(path string) bool
	ContainsPath(path PathBuilder) bool
	Clear() error

	// Set information
	Size() int
	IsEmpty() bool
	Paths() []string
	PathBuilders() []PathBuilder

	// Set operations
	Union(other PathSet) PathSet
	Intersection(other PathSet) PathSet
	Difference(other PathSet) PathSet
	SymmetricDifference(other PathSet) PathSet

	// Filtering
	Filter(predicate func(string) bool) PathSet
	FilterPaths(predicate func(PathBuilder) bool) PathSet

	// Iteration
	ForEach(fn func(string) error) error
	ForEachPath(fn func(PathBuilder) error) error
}
