// Package paths provides advanced path building and manipulation utilities with mockable interfaces
package paths

import (
	"io/fs"
	"sync"
	"time"
)

// Global default instances - thread-safe singleton pattern
var (
	// All default instances use getter functions for dependency injection
	defaultBuilderOnce sync.Once   //nolint:gochecknoglobals // Required for thread-safe singleton pattern
	defaultBuilderData PathBuilder //nolint:gochecknoglobals // Private data for singleton pattern

	defaultMatcherOnce sync.Once   //nolint:gochecknoglobals // Required for thread-safe singleton pattern
	defaultMatcherData PathMatcher //nolint:gochecknoglobals // Private data for singleton pattern

	defaultValidatorOnce sync.Once     //nolint:gochecknoglobals // Required for thread-safe singleton pattern
	defaultValidatorData PathValidator //nolint:gochecknoglobals // Private data for singleton pattern

	defaultSetOnce sync.Once //nolint:gochecknoglobals // Required for thread-safe singleton pattern
	defaultSetData PathSet   //nolint:gochecknoglobals // Private data for singleton pattern

	defaultWatcherOnce sync.Once   //nolint:gochecknoglobals // Required for thread-safe singleton pattern
	defaultWatcherData PathWatcher //nolint:gochecknoglobals // Private data for singleton pattern

	defaultCacheOnce sync.Once //nolint:gochecknoglobals // Required for thread-safe singleton pattern
	defaultCacheData PathCache //nolint:gochecknoglobals // Private data for singleton pattern
)

// GetDefaultBuilder returns the default PathBuilder instance using thread-safe initialization
func GetDefaultBuilder() PathBuilder {
	defaultBuilderOnce.Do(func() {
		defaultBuilderData = NewPathBuilder("")
	})
	return defaultBuilderData
}

// GetDefaultMatcher returns the default PathMatcher instance using thread-safe initialization
func GetDefaultMatcher() PathMatcher {
	defaultMatcherOnce.Do(func() {
		defaultMatcherData = NewPathMatcher()
	})
	return defaultMatcherData
}

// GetDefaultValidator returns the default PathValidator instance using thread-safe initialization
func GetDefaultValidator() PathValidator {
	defaultValidatorOnce.Do(func() {
		defaultValidatorData = NewPathValidator()
	})
	return defaultValidatorData
}

// GetDefaultSet returns the default PathSet instance using thread-safe initialization
func GetDefaultSet() PathSet {
	defaultSetOnce.Do(func() {
		defaultSetData = NewPathSet()
	})
	return defaultSetData
}

// GetDefaultWatcher returns the default PathWatcher instance using thread-safe initialization
func GetDefaultWatcher() PathWatcher {
	defaultWatcherOnce.Do(func() {
		defaultWatcherData = NewPathWatcher()
	})
	return defaultWatcherData
}

// GetDefaultCache returns the default PathCache instance using thread-safe initialization
func GetDefaultCache() PathCache {
	defaultCacheOnce.Do(func() {
		defaultCacheData = NewPathCache()
	})
	return defaultCacheData
}

// Lazy wrapper types for backward compatibility
type lazyPathBuilder struct{}

func (l *lazyPathBuilder) Join(elements ...string) PathBuilder {
	return GetDefaultBuilder().Join(elements...)
}

func (l *lazyPathBuilder) Dir() PathBuilder {
	return GetDefaultBuilder().Dir()
}

func (l *lazyPathBuilder) Base() string {
	return GetDefaultBuilder().Base()
}

func (l *lazyPathBuilder) Ext() string {
	return GetDefaultBuilder().Ext()
}

func (l *lazyPathBuilder) Clean() PathBuilder {
	return GetDefaultBuilder().Clean()
}

func (l *lazyPathBuilder) Abs() (PathBuilder, error) {
	return GetDefaultBuilder().Abs()
}

func (l *lazyPathBuilder) Append(suffix string) PathBuilder {
	return GetDefaultBuilder().Append(suffix)
}

func (l *lazyPathBuilder) Prepend(prefix string) PathBuilder {
	return GetDefaultBuilder().Prepend(prefix)
}

func (l *lazyPathBuilder) WithExt(ext string) PathBuilder {
	return GetDefaultBuilder().WithExt(ext)
}

func (l *lazyPathBuilder) WithoutExt() PathBuilder {
	return GetDefaultBuilder().WithoutExt()
}

func (l *lazyPathBuilder) WithName(name string) PathBuilder {
	return GetDefaultBuilder().WithName(name)
}

func (l *lazyPathBuilder) Rel(basepath string) (PathBuilder, error) {
	return GetDefaultBuilder().Rel(basepath)
}

func (l *lazyPathBuilder) RelTo(target PathBuilder) (PathBuilder, error) {
	return GetDefaultBuilder().RelTo(target)
}

func (l *lazyPathBuilder) String() string {
	return GetDefaultBuilder().String()
}

func (l *lazyPathBuilder) IsAbs() bool {
	return GetDefaultBuilder().IsAbs()
}

func (l *lazyPathBuilder) IsDir() bool {
	return GetDefaultBuilder().IsDir()
}

func (l *lazyPathBuilder) IsFile() bool {
	return GetDefaultBuilder().IsFile()
}

func (l *lazyPathBuilder) Exists() bool {
	return GetDefaultBuilder().Exists()
}

func (l *lazyPathBuilder) Size() int64 {
	return GetDefaultBuilder().Size()
}

func (l *lazyPathBuilder) ModTime() time.Time {
	return GetDefaultBuilder().ModTime()
}

func (l *lazyPathBuilder) Mode() fs.FileMode {
	return GetDefaultBuilder().Mode()
}

func (l *lazyPathBuilder) Walk(fn WalkFunc) error {
	return GetDefaultBuilder().Walk(fn)
}

func (l *lazyPathBuilder) List() ([]PathBuilder, error) {
	return GetDefaultBuilder().List()
}

func (l *lazyPathBuilder) ListFiles() ([]PathBuilder, error) {
	return GetDefaultBuilder().ListFiles()
}

func (l *lazyPathBuilder) ListDirs() ([]PathBuilder, error) {
	return GetDefaultBuilder().ListDirs()
}

func (l *lazyPathBuilder) Glob(pattern string) ([]PathBuilder, error) {
	return GetDefaultBuilder().Glob(pattern)
}

func (l *lazyPathBuilder) Validate() error {
	return GetDefaultBuilder().Validate()
}

func (l *lazyPathBuilder) IsValid() bool {
	return GetDefaultBuilder().IsValid()
}

func (l *lazyPathBuilder) IsEmpty() bool {
	return GetDefaultBuilder().IsEmpty()
}

func (l *lazyPathBuilder) IsSafe() bool {
	return GetDefaultBuilder().IsSafe()
}

func (l *lazyPathBuilder) Create() error {
	return GetDefaultBuilder().Create()
}

func (l *lazyPathBuilder) CreateDir() error {
	return GetDefaultBuilder().CreateDir()
}

func (l *lazyPathBuilder) CreateDirAll() error {
	return GetDefaultBuilder().CreateDirAll()
}

func (l *lazyPathBuilder) Remove() error {
	return GetDefaultBuilder().Remove()
}

func (l *lazyPathBuilder) RemoveAll() error {
	return GetDefaultBuilder().RemoveAll()
}

func (l *lazyPathBuilder) Copy(dest PathBuilder) error {
	return GetDefaultBuilder().Copy(dest)
}

func (l *lazyPathBuilder) Move(dest PathBuilder) error {
	return GetDefaultBuilder().Move(dest)
}

func (l *lazyPathBuilder) Readlink() (PathBuilder, error) {
	return GetDefaultBuilder().Readlink()
}

func (l *lazyPathBuilder) Symlink(target PathBuilder) error {
	return GetDefaultBuilder().Symlink(target)
}

func (l *lazyPathBuilder) Match(pattern string) bool {
	return GetDefaultBuilder().Match(pattern)
}

func (l *lazyPathBuilder) Contains(sub string) bool {
	return GetDefaultBuilder().Contains(sub)
}

func (l *lazyPathBuilder) HasPrefix(prefix string) bool {
	return GetDefaultBuilder().HasPrefix(prefix)
}

func (l *lazyPathBuilder) HasSuffix(suffix string) bool {
	return GetDefaultBuilder().HasSuffix(suffix)
}

func (l *lazyPathBuilder) Clone() PathBuilder {
	return GetDefaultBuilder().Clone()
}

// lazyPathMatcher provides lazy initialization for the default PathMatcher instance
type lazyPathMatcher struct{}

func (l *lazyPathMatcher) Match(path string) bool {
	return GetDefaultMatcher().Match(path)
}

func (l *lazyPathMatcher) MatchPath(path PathBuilder) bool {
	return GetDefaultMatcher().MatchPath(path)
}

func (l *lazyPathMatcher) Compile(pattern string) error {
	return GetDefaultMatcher().Compile(pattern)
}

func (l *lazyPathMatcher) Pattern() string {
	return GetDefaultMatcher().Pattern()
}

func (l *lazyPathMatcher) AddPattern(pattern string) error {
	return GetDefaultMatcher().AddPattern(pattern)
}

func (l *lazyPathMatcher) RemovePattern(pattern string) error {
	return GetDefaultMatcher().RemovePattern(pattern)
}

func (l *lazyPathMatcher) ClearPatterns() error {
	return GetDefaultMatcher().ClearPatterns()
}

func (l *lazyPathMatcher) Patterns() []string {
	return GetDefaultMatcher().Patterns()
}

func (l *lazyPathMatcher) SetCaseSensitive(sensitive bool) PathMatcher {
	return GetDefaultMatcher().SetCaseSensitive(sensitive)
}

func (l *lazyPathMatcher) SetRecursive(recursive bool) PathMatcher {
	return GetDefaultMatcher().SetRecursive(recursive)
}

func (l *lazyPathMatcher) SetMaxDepth(depth int) PathMatcher {
	return GetDefaultMatcher().SetMaxDepth(depth)
}

func (l *lazyPathMatcher) MatchAny(paths ...string) bool {
	return GetDefaultMatcher().MatchAny(paths...)
}

func (l *lazyPathMatcher) MatchAll(paths ...string) bool {
	return GetDefaultMatcher().MatchAll(paths...)
}

func (l *lazyPathMatcher) Filter(paths []string) []string {
	return GetDefaultMatcher().Filter(paths)
}

func (l *lazyPathMatcher) FilterPaths(paths []PathBuilder) []PathBuilder {
	return GetDefaultMatcher().FilterPaths(paths)
}

// lazyPathValidator provides lazy initialization for the default PathValidator instance
type lazyPathValidator struct{}

func (l *lazyPathValidator) AddRule(rule ValidationRule) error {
	return GetDefaultValidator().AddRule(rule)
}

func (l *lazyPathValidator) RemoveRule(name string) error {
	return GetDefaultValidator().RemoveRule(name)
}

func (l *lazyPathValidator) ClearRules() error {
	return GetDefaultValidator().ClearRules()
}

func (l *lazyPathValidator) Rules() []ValidationRule {
	return GetDefaultValidator().Rules()
}

func (l *lazyPathValidator) Validate(path string) []ValidationError {
	return GetDefaultValidator().Validate(path)
}

func (l *lazyPathValidator) ValidatePath(path PathBuilder) []ValidationError {
	return GetDefaultValidator().ValidatePath(path)
}

func (l *lazyPathValidator) IsValid(path string) bool {
	return GetDefaultValidator().IsValid(path)
}

func (l *lazyPathValidator) IsValidPath(path PathBuilder) bool {
	return GetDefaultValidator().IsValidPath(path)
}

func (l *lazyPathValidator) RequireAbsolute() PathValidator {
	return GetDefaultValidator().RequireAbsolute()
}

func (l *lazyPathValidator) RequireRelative() PathValidator {
	return GetDefaultValidator().RequireRelative()
}

func (l *lazyPathValidator) RequireExists() PathValidator {
	return GetDefaultValidator().RequireExists()
}

func (l *lazyPathValidator) RequireNotExists() PathValidator {
	return GetDefaultValidator().RequireNotExists()
}

func (l *lazyPathValidator) RequireReadable() PathValidator {
	return GetDefaultValidator().RequireReadable()
}

func (l *lazyPathValidator) RequireWritable() PathValidator {
	return GetDefaultValidator().RequireWritable()
}

func (l *lazyPathValidator) RequireExecutable() PathValidator {
	return GetDefaultValidator().RequireExecutable()
}

func (l *lazyPathValidator) RequireDirectory() PathValidator {
	return GetDefaultValidator().RequireDirectory()
}

func (l *lazyPathValidator) RequireFile() PathValidator {
	return GetDefaultValidator().RequireFile()
}

func (l *lazyPathValidator) RequireExtension(exts ...string) PathValidator {
	return GetDefaultValidator().RequireExtension(exts...)
}

func (l *lazyPathValidator) RequireMaxLength(length int) PathValidator {
	return GetDefaultValidator().RequireMaxLength(length)
}

func (l *lazyPathValidator) RequirePattern(pattern string) PathValidator {
	return GetDefaultValidator().RequirePattern(pattern)
}

func (l *lazyPathValidator) ForbidPattern(pattern string) PathValidator {
	return GetDefaultValidator().ForbidPattern(pattern)
}

// lazyPathSet provides lazy initialization for the default PathSet instance
type lazyPathSet struct{}

func (l *lazyPathSet) Add(path string) bool {
	return GetDefaultSet().Add(path)
}

func (l *lazyPathSet) AddPath(path PathBuilder) bool {
	return GetDefaultSet().AddPath(path)
}

func (l *lazyPathSet) Remove(path string) bool {
	return GetDefaultSet().Remove(path)
}

func (l *lazyPathSet) RemovePath(path PathBuilder) bool {
	return GetDefaultSet().RemovePath(path)
}

func (l *lazyPathSet) Contains(path string) bool {
	return GetDefaultSet().Contains(path)
}

func (l *lazyPathSet) ContainsPath(path PathBuilder) bool {
	return GetDefaultSet().ContainsPath(path)
}

func (l *lazyPathSet) Clear() error {
	return GetDefaultSet().Clear()
}

func (l *lazyPathSet) Size() int {
	return GetDefaultSet().Size()
}

func (l *lazyPathSet) IsEmpty() bool {
	return GetDefaultSet().IsEmpty()
}

func (l *lazyPathSet) Paths() []string {
	return GetDefaultSet().Paths()
}

func (l *lazyPathSet) PathBuilders() []PathBuilder {
	return GetDefaultSet().PathBuilders()
}

func (l *lazyPathSet) Union(other PathSet) PathSet {
	return GetDefaultSet().Union(other)
}

func (l *lazyPathSet) Intersection(other PathSet) PathSet {
	return GetDefaultSet().Intersection(other)
}

func (l *lazyPathSet) Difference(other PathSet) PathSet {
	return GetDefaultSet().Difference(other)
}

func (l *lazyPathSet) SymmetricDifference(other PathSet) PathSet {
	return GetDefaultSet().SymmetricDifference(other)
}

func (l *lazyPathSet) Filter(predicate func(string) bool) PathSet {
	return GetDefaultSet().Filter(predicate)
}

func (l *lazyPathSet) FilterPaths(predicate func(PathBuilder) bool) PathSet {
	return GetDefaultSet().FilterPaths(predicate)
}

func (l *lazyPathSet) ForEach(fn func(string) error) error {
	return GetDefaultSet().ForEach(fn)
}

func (l *lazyPathSet) ForEachPath(fn func(PathBuilder) error) error {
	return GetDefaultSet().ForEachPath(fn)
}

// lazyPathWatcher provides lazy initialization for the default PathWatcher instance
type lazyPathWatcher struct{}

func (l *lazyPathWatcher) Watch(path string, events EventMask) error {
	return GetDefaultWatcher().Watch(path, events)
}

func (l *lazyPathWatcher) WatchPath(path PathBuilder, events EventMask) error {
	return GetDefaultWatcher().WatchPath(path, events)
}

func (l *lazyPathWatcher) Unwatch(path string) error {
	return GetDefaultWatcher().Unwatch(path)
}

func (l *lazyPathWatcher) UnwatchPath(path PathBuilder) error {
	return GetDefaultWatcher().UnwatchPath(path)
}

func (l *lazyPathWatcher) Events() <-chan *PathEvent {
	return GetDefaultWatcher().Events()
}

func (l *lazyPathWatcher) Errors() <-chan error {
	return GetDefaultWatcher().Errors()
}

func (l *lazyPathWatcher) Close() error {
	return GetDefaultWatcher().Close()
}

func (l *lazyPathWatcher) SetBufferSize(size int) PathWatcher {
	return GetDefaultWatcher().SetBufferSize(size)
}

func (l *lazyPathWatcher) SetRecursive(recursive bool) PathWatcher {
	return GetDefaultWatcher().SetRecursive(recursive)
}

func (l *lazyPathWatcher) SetDebounce(duration time.Duration) PathWatcher {
	return GetDefaultWatcher().SetDebounce(duration)
}

func (l *lazyPathWatcher) IsWatching(path string) bool {
	return GetDefaultWatcher().IsWatching(path)
}

func (l *lazyPathWatcher) WatchedPaths() []string {
	return GetDefaultWatcher().WatchedPaths()
}

// lazyPathCache provides lazy initialization for the default PathCache instance
type lazyPathCache struct{}

func (l *lazyPathCache) Get(key string) (PathBuilder, bool) {
	return GetDefaultCache().Get(key)
}

func (l *lazyPathCache) Set(key string, path PathBuilder) error {
	return GetDefaultCache().Set(key, path)
}

func (l *lazyPathCache) Delete(key string) error {
	return GetDefaultCache().Delete(key)
}

func (l *lazyPathCache) Clear() error {
	return GetDefaultCache().Clear()
}

func (l *lazyPathCache) Size() int {
	return GetDefaultCache().Size()
}

func (l *lazyPathCache) Keys() []string {
	return GetDefaultCache().Keys()
}

func (l *lazyPathCache) Stats() CacheStats {
	return GetDefaultCache().Stats()
}

func (l *lazyPathCache) SetMaxSize(size int) PathCache {
	return GetDefaultCache().SetMaxSize(size)
}

func (l *lazyPathCache) SetTTL(ttl time.Duration) PathCache {
	return GetDefaultCache().SetTTL(ttl)
}

func (l *lazyPathCache) SetEvictionPolicy(policy EvictionPolicy) PathCache {
	return GetDefaultCache().SetEvictionPolicy(policy)
}

func (l *lazyPathCache) Validate(key string) error {
	return GetDefaultCache().Validate(key)
}

func (l *lazyPathCache) Refresh(key string) error {
	return GetDefaultCache().Refresh(key)
}

func (l *lazyPathCache) RefreshAll() error {
	return GetDefaultCache().RefreshAll()
}

// Backward compatibility: provide the original global variable names as lazy-initialized instances
// These will be initialized on first access, maintaining the same interface
var (
	DefaultBuilder   PathBuilder   = &lazyPathBuilder{}   //nolint:gochecknoglobals // Backward compatibility wrapper
	DefaultMatcher   PathMatcher   = &lazyPathMatcher{}   //nolint:gochecknoglobals // Backward compatibility wrapper
	DefaultValidator PathValidator = &lazyPathValidator{} //nolint:gochecknoglobals // Backward compatibility wrapper
	DefaultSet       PathSet       = &lazyPathSet{}       //nolint:gochecknoglobals // Backward compatibility wrapper
	DefaultWatcher   PathWatcher   = &lazyPathWatcher{}   //nolint:gochecknoglobals // Backward compatibility wrapper
	DefaultCache     PathCache     = &lazyPathCache{}     //nolint:gochecknoglobals // Backward compatibility wrapper
)

// Path Builder convenience functions

// Build creates a new PathBuilder from elements
func Build(elements ...string) PathBuilder {
	if len(elements) == 0 {
		return NewPathBuilder("")
	}
	if len(elements) == 1 {
		return NewPathBuilder(elements[0])
	}
	return NewPathBuilder(elements[0]).Join(elements[1:]...)
}

// File creates a PathBuilder for a file path
func File(path string) PathBuilder {
	return NewPathBuilder(path)
}

// Dir creates a PathBuilder for a directory path
func Dir(path string) PathBuilder {
	return NewPathBuilder(path)
}

// Current returns a PathBuilder for the current working directory
func Current() PathBuilder {
	return NewPathBuilder(".")
}

// Home returns a PathBuilder for the user's home directory
func Home() PathBuilder {
	// This could integrate with the env package for home directory resolution
	return NewPathBuilder("~")
}

// Root returns a PathBuilder for the root directory
func Root() PathBuilder {
	return NewPathBuilder("/")
}

// Path manipulation convenience functions

// Clean cleans and returns a new PathBuilder
func Clean(path string) PathBuilder {
	return NewPathBuilder(path).Clean()
}

// Abs returns an absolute PathBuilder
func Abs(path string) (PathBuilder, error) {
	return NewPathBuilder(path).Abs()
}

// Base returns the base name of a path
func Base(path string) string {
	return NewPathBuilder(path).Base()
}

// Ext returns the extension of a path
func Ext(path string) string {
	return NewPathBuilder(path).Ext()
}

// DirOf returns the directory component of the given path.
func DirOf(path string) PathBuilder {
	return NewPathBuilder(path).Dir()
}

// JoinPaths joins multiple path elements into a single path.
func JoinPaths(elements ...string) PathBuilder {
	return Join(elements...)
}

// Matcher convenience functions

// MatchPattern returns true if path matches pattern
func MatchPattern(pattern, path string) bool {
	return Match(pattern, path)
}

// MatchAnyPattern returns true if path matches any pattern
func MatchAnyPattern(path string, patterns ...string) bool {
	return MatchAny(path, patterns...)
}

// CreateMatcher creates a new matcher with patterns
func CreateMatcher(patterns ...string) (PathMatcher, error) {
	matcher := NewPathMatcher()
	for _, pattern := range patterns {
		if err := matcher.AddPattern(pattern); err != nil {
			return nil, err
		}
	}
	return matcher, nil
}

// Validator convenience functions

// ValidatePath validates a path with common rules
func ValidatePath(path string) []ValidationError {
	validator := NewPathValidator()
	return validator.Validate(path)
}

// ValidatePathExists validates that a path exists
func ValidatePathExists(path string) error {
	return ValidateExists(path)
}

// ValidatePathReadable validates that a path is readable
func ValidatePathReadable(path string) error {
	return ValidateReadable(path)
}

// ValidatePathWritable validates that a path is writable
func ValidatePathWritable(path string) error {
	return ValidateWritable(path)
}

// CreateValidator creates a validator with common rules
func CreateValidator() PathValidator {
	return NewPathValidator()
}

// CreateStrictValidator creates a validator with strict rules
func CreateStrictValidator() PathValidator {
	validator := NewPathValidator()
	validator.RequireAbsolute()
	validator.RequireExists()
	validator.RequireReadable()
	return validator
}

// Set convenience functions

// CreateSet creates a new PathSet from paths
func CreateSet(paths ...string) PathSet {
	return NewPathSetFromPaths(paths)
}

// CreateSetFromBuilders creates a new PathSet from PathBuilders
func CreateSetFromBuilders(paths ...PathBuilder) PathSet {
	return NewPathSetFromPathBuilders(paths)
}

// UnionAll returns the union of multiple sets
func UnionAll(sets ...PathSet) PathSet {
	return UnionSets(sets...)
}

// IntersectionAll returns the intersection of multiple sets
func IntersectionAll(sets ...PathSet) PathSet {
	return IntersectionSets(sets...)
}

// File system operation shortcuts

// CreateFile creates a file at the given path
func CreateFile(path string) error {
	return NewPathBuilder(path).Create()
}

// CreateDirectory creates a directory at the given path
func CreateDirectory(path string) error {
	return NewPathBuilder(path).CreateDirAll()
}

// RemovePath removes a path (file or directory)
func RemovePath(path string) error {
	return NewPathBuilder(path).Remove()
}

// RemoveAll removes a path and all its contents
func RemoveAll(path string) error {
	return NewPathBuilder(path).RemoveAll()
}

// CopyPath copies a path to another location
func CopyPath(src, dst string) error {
	srcPath := NewPathBuilder(src)
	dstPath := NewPathBuilder(dst)
	return srcPath.Copy(dstPath)
}

// MovePath moves a path to another location
func MovePath(src, dst string) error {
	srcPath := NewPathBuilder(src)
	dstPath := NewPathBuilder(dst)
	return srcPath.Move(dstPath)
}

// Information shortcuts

// PathExists checks if a path exists
func PathExists(path string) bool {
	return Exists(path)
}

// PathIsDir checks if a path is a directory
func PathIsDir(path string) bool {
	return IsDir(path)
}

// PathIsFile checks if a path is a file
func PathIsFile(path string) bool {
	return IsFile(path)
}

// PathSize returns the size of a file
func PathSize(path string) int64 {
	return NewPathBuilder(path).Size()
}

// PathModTime returns the modification time of a path
func PathModTime(path string) time.Time {
	return NewPathBuilder(path).ModTime()
}

// Advanced operations

// FindFiles finds all files matching a pattern
func FindFiles(rootPath, pattern string) ([]PathBuilder, error) {
	root := NewPathBuilder(rootPath)
	if !root.IsDir() {
		return nil, &ValidationError{
			Path:    rootPath,
			Rule:    "directory",
			Message: "root path must be a directory",
		}
	}

	var result []PathBuilder
	err := root.Walk(func(path PathBuilder, _ fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path.IsFile() && path.Match(pattern) {
			result = append(result, path)
		}

		return nil
	})

	return result, err
}

// FindDirectories finds all directories matching a pattern
func FindDirectories(rootPath, pattern string) ([]PathBuilder, error) {
	root := NewPathBuilder(rootPath)
	if !root.IsDir() {
		return nil, &ValidationError{
			Path:    rootPath,
			Rule:    "directory",
			Message: "root path must be a directory",
		}
	}

	var result []PathBuilder
	err := root.Walk(func(path PathBuilder, _ fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path.IsDir() && path.Match(pattern) {
			result = append(result, path)
		}

		return nil
	})

	return result, err
}

// SafePath creates a safe PathBuilder with validation
func SafePath(path string) (PathBuilder, error) {
	pb := NewPathBuilder(path)
	if err := pb.Validate(); err != nil {
		return nil, err
	}
	return pb, nil
}

// SecurePath creates a PathBuilder with security restrictions
func SecurePath(path, basePath string) (PathBuilder, error) {
	options := PathOptions{
		RestrictToBasePath: basePath,
		AllowUnsafePaths:   false,
		MaxPathLength:      1024,
	}

	pb := NewPathBuilderWithOptions(path, options)
	if err := pb.Validate(); err != nil {
		return nil, err
	}

	return pb, nil
}

// Configuration functions

// Package-level variables for path options configuration
var (
	defaultPathOptionsOnce sync.Once    //nolint:gochecknoglobals // Required for thread-safe initialization
	defaultPathOptionsData PathOptions  //nolint:gochecknoglobals // Private data for sync.Once pattern
	defaultPathOptionsMu   sync.RWMutex //nolint:gochecknoglobals // Required for thread-safe read/write access
)

// getDefaultPathOptions returns the default path options using thread-safe initialization
func getDefaultPathOptions() PathOptions {
	defaultPathOptionsOnce.Do(func() {
		defaultPathOptionsData = PathOptions{
			CreateMode:    0o755,
			CreateParents: true,
			BufferSize:    8192,
		}
	})
	defaultPathOptionsMu.RLock()
	defer defaultPathOptionsMu.RUnlock()
	return defaultPathOptionsData
}

// SetDefaultOptions sets default options for new PathBuilders
func SetDefaultOptions(options PathOptions) {
	// Ensure initialization has occurred
	getDefaultPathOptions()

	// Update the options with write lock for thread safety
	defaultPathOptionsMu.Lock()
	defaultPathOptionsData = options
	defaultPathOptionsMu.Unlock()
}

// GetDefaultOptions returns the current default options
func GetDefaultOptions() PathOptions {
	return getDefaultPathOptions()
}

// Utility functions for testing and mocking

// SetDefaultBuilder sets the global default builder (for testing)
func SetDefaultBuilder(builder PathBuilder) {
	if builder != nil {
		// Force initialization and then replace the instance
		GetDefaultBuilder()
		defaultBuilderData = builder
	}
}

// SetDefaultMatcher sets the global default matcher (for testing)
func SetDefaultMatcher(matcher PathMatcher) {
	if matcher != nil {
		// Force initialization and then replace the instance
		GetDefaultMatcher()
		defaultMatcherData = matcher
	}
}

// SetDefaultValidator sets the global default validator (for testing)
func SetDefaultValidator(validator PathValidator) {
	if validator != nil {
		// Force initialization and then replace the instance
		GetDefaultValidator()
		defaultValidatorData = validator
	}
}

// SetDefaultSet sets the global default set (for testing)
func SetDefaultSet(set PathSet) {
	if set != nil {
		// Force initialization and then replace the instance
		GetDefaultSet()
		defaultSetData = set
	}
}

// SetDefaultWatcher sets the global default watcher (for testing)
func SetDefaultWatcher(watcher PathWatcher) {
	if watcher != nil {
		// Force initialization and then replace the instance
		GetDefaultWatcher()
		defaultWatcherData = watcher
	}
}

// SetDefaultCache sets the global default cache (for testing)
func SetDefaultCache(cache PathCache) {
	if cache != nil {
		// Force initialization and then replace the instance
		GetDefaultCache()
		defaultCacheData = cache
	}
}

// Batch operations

// ProcessPaths processes multiple paths with a function
func ProcessPaths(paths []string, fn func(PathBuilder) error) error {
	for _, path := range paths {
		pb := NewPathBuilder(path)
		if err := fn(pb); err != nil {
			return err
		}
	}
	return nil
}

// FilterPaths filters paths based on a predicate
func FilterPaths(paths []string, predicate func(PathBuilder) bool) []PathBuilder {
	var result []PathBuilder
	for _, path := range paths {
		pb := NewPathBuilder(path)
		if predicate(pb) {
			result = append(result, pb)
		}
	}
	return result
}

// MapPaths transforms paths using a function
func MapPaths(paths []string, fn func(PathBuilder) PathBuilder) []PathBuilder {
	result := make([]PathBuilder, len(paths))
	for i, path := range paths {
		pb := NewPathBuilder(path)
		result[i] = fn(pb)
	}
	return result
}

// SortPaths sorts PathBuilders by their string representation
func SortPaths(paths []PathBuilder) []PathBuilder {
	// Create a copy to avoid modifying the original slice
	result := make([]PathBuilder, len(paths))
	copy(result, paths)

	// Simple bubble sort for demonstration
	// In a real implementation, you'd use sort.Slice
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].String() > result[j].String() {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// Advanced path analysis

// AnalyzePath returns detailed information about a path
func AnalyzePath(path string) (*PathInfo, error) {
	pb := NewPathBuilder(path)

	abs, err := pb.Abs()
	if err != nil {
		return nil, err
	}

	info := &PathInfo{
		Path:         pb.String(),
		AbsPath:      abs.String(),
		Size:         pb.Size(),
		Mode:         pb.Mode(),
		ModTime:      pb.ModTime(),
		IsDir:        pb.IsDir(),
		IsFile:       pb.IsFile(),
		IsSymlink:    false, // Would need additional logic to detect
		IsExecutable: (pb.Mode() & 0o111) != 0,
		IsReadable:   ValidateReadable(pb.String()) == nil,
		IsWritable:   ValidateWritable(pb.String()) == nil,
	}

	return info, nil
}

// PathWatcher convenience functions

// WatchPath starts watching a path using the default watcher
func WatchPath(path string, events EventMask) error {
	return GetDefaultWatcher().Watch(path, events)
}

// UnwatchPath stops watching a path using the default watcher
func UnwatchPath(path string) error {
	return GetDefaultWatcher().Unwatch(path)
}

// IsWatchingPath checks if a path is being watched using the default watcher
func IsWatchingPath(path string) bool {
	return GetDefaultWatcher().IsWatching(path)
}

// GetAllWatchedPaths returns all watched paths using the default watcher
func GetAllWatchedPaths() []string {
	return GetDefaultWatcher().WatchedPaths()
}

// PathCache convenience functions

// CachePath stores a value in the default cache
func CachePath(key string, path PathBuilder) error {
	return GetDefaultCache().Set(key, path)
}

// GetCachedPath retrieves a value from the default cache
func GetCachedPath(key string) (PathBuilder, bool) {
	return GetDefaultCache().Get(key)
}

// DeleteCachedPath removes a value from the default cache
func DeleteCachedPath(key string) error {
	return GetDefaultCache().Delete(key)
}

// ClearCache clears all cached values
func ClearCache() error {
	return GetDefaultCache().Clear()
}

// GetCacheStats returns cache statistics
func GetCacheStats() CacheStats {
	return GetDefaultCache().Stats()
}

// ExpireCachedPaths removes expired cache entries
func ExpireCachedPaths() int {
	cache := GetDefaultCache()
	if defaultCache, ok := cache.(*DefaultPathCache); ok {
		return defaultCache.Expire()
	}
	return 0
}
