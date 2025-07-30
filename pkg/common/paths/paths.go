// Package paths provides advanced path building and manipulation utilities with mockable interfaces
package paths

import (
	"io/fs"
	"time"
)

// Global default instances
var (
	DefaultBuilder   = NewPathBuilder("") //nolint:gochecknoglobals // Package-level default
	DefaultMatcher   = NewPathMatcher()   //nolint:gochecknoglobals // Package-level default
	DefaultValidator = NewPathValidator() //nolint:gochecknoglobals // Package-level default
	DefaultSet       = NewPathSet()       //nolint:gochecknoglobals // Package-level default
	DefaultWatcher   = NewPathWatcher()   //nolint:gochecknoglobals // Package-level default
	DefaultCache     = NewPathCache()     //nolint:gochecknoglobals // Package-level default
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

// Dir returns the directory component
func DirOf(path string) PathBuilder {
	return NewPathBuilder(path).Dir()
}

// Join joins path elements
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
	err := root.Walk(func(path PathBuilder, info fs.FileInfo, err error) error {
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
	err := root.Walk(func(path PathBuilder, info fs.FileInfo, err error) error {
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

// SetDefaultOptions sets default options for new PathBuilders
func SetDefaultOptions(options PathOptions) {
	// Update global default options that would be used by new PathBuilders
	// This implementation stores options for future PathBuilder creations
	globalPathOptions = options
}

// Global options storage
var globalPathOptions = PathOptions{ //nolint:gochecknoglobals // Package-level configuration
	CreateMode:    0o755,
	CreateParents: true,
	BufferSize:    8192,
}

// GetDefaultOptions returns the current default options
func GetDefaultOptions() PathOptions {
	return globalPathOptions
}

// Utility functions for testing and mocking

// SetDefaultBuilder sets the global default builder (for testing)
func SetDefaultBuilder(builder PathBuilder) {
	if builder != nil {
		if defaultBuilder, ok := builder.(*DefaultPathBuilder); ok {
			DefaultBuilder = defaultBuilder
		}
	}
}

// SetDefaultMatcher sets the global default matcher (for testing)
func SetDefaultMatcher(matcher PathMatcher) {
	if defaultMatcher, ok := matcher.(*DefaultPathMatcher); ok {
		DefaultMatcher = defaultMatcher
	}
}

// SetDefaultValidator sets the global default validator (for testing)
func SetDefaultValidator(validator PathValidator) {
	if defaultValidator, ok := validator.(*DefaultPathValidator); ok {
		DefaultValidator = defaultValidator
	}
}

// SetDefaultSet sets the global default set (for testing)
func SetDefaultSet(set PathSet) {
	if defaultSet, ok := set.(*DefaultPathSet); ok {
		DefaultSet = defaultSet
	}
}

// SetDefaultWatcher sets the global default watcher (for testing)
func SetDefaultWatcher(watcher PathWatcher) {
	if watcher != nil {
		DefaultWatcher = watcher
	}
}

// SetDefaultCache sets the global default cache (for testing)
func SetDefaultCache(cache PathCache) {
	if cache != nil {
		DefaultCache = cache
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
	return DefaultWatcher.Watch(path, events)
}

// UnwatchPath stops watching a path using the default watcher
func UnwatchPath(path string) error {
	return DefaultWatcher.Unwatch(path)
}

// IsWatchingPath checks if a path is being watched using the default watcher
func IsWatchingPath(path string) bool {
	return DefaultWatcher.IsWatching(path)
}

// GetAllWatchedPaths returns all watched paths using the default watcher
func GetAllWatchedPaths() []string {
	return DefaultWatcher.WatchedPaths()
}

// PathCache convenience functions

// CachePath stores a value in the default cache
func CachePath(key string, path PathBuilder) error {
	return DefaultCache.Set(key, path)
}

// GetCachedPath retrieves a value from the default cache
func GetCachedPath(key string) (PathBuilder, bool) {
	return DefaultCache.Get(key)
}

// DeleteCachedPath removes a value from the default cache
func DeleteCachedPath(key string) error {
	return DefaultCache.Delete(key)
}

// ClearCache clears all cached values
func ClearCache() error {
	return DefaultCache.Clear()
}

// GetCacheStats returns cache statistics
func GetCacheStats() CacheStats {
	return DefaultCache.Stats()
}

// ExpireCachedPaths removes expired cache entries
func ExpireCachedPaths() int {
	if defaultCache, ok := DefaultCache.(*DefaultPathCache); ok {
		return defaultCache.Expire()
	}
	return 0
}
