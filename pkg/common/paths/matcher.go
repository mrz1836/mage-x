package paths

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// Error definitions for matcher operations
var (
	ErrPatternEmpty    = errors.New("pattern cannot be empty")
	ErrPatternNotFound = errors.New("pattern not found")
)

// DefaultPathMatcher implements PathMatcher
type DefaultPathMatcher struct {
	mu            sync.RWMutex
	patterns      []string
	compiledRegex []*regexp.Regexp
	caseSensitive bool
	recursive     bool
	maxDepth      int
}

// NewPathMatcher creates a new path matcher
func NewPathMatcher() *DefaultPathMatcher {
	return &DefaultPathMatcher{
		patterns:      make([]string, 0),
		compiledRegex: make([]*regexp.Regexp, 0),
		caseSensitive: true,
		recursive:     false,
		maxDepth:      -1,
	}
}

// NewPathMatcherWithPattern creates a new path matcher with an initial pattern
func NewPathMatcherWithPattern(pattern string) (*DefaultPathMatcher, error) {
	matcher := NewPathMatcher()
	if err := matcher.AddPattern(pattern); err != nil {
		return nil, err
	}
	return matcher, nil
}

// Pattern matching

// Match returns true if the path matches any pattern
func (pm *DefaultPathMatcher) Match(path string) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.patterns) == 0 {
		return false
	}

	// Try glob patterns first
	for _, pattern := range pm.patterns {
		if pm.matchGlob(path, pattern) {
			return true
		}
	}

	// Try regex patterns
	for _, regex := range pm.compiledRegex {
		if regex.MatchString(path) {
			return true
		}
	}

	return false
}

// MatchPath returns true if the PathBuilder matches any pattern
func (pm *DefaultPathMatcher) MatchPath(path PathBuilder) bool {
	return pm.Match(path.String())
}

// Compile compiles a pattern (supports both glob and regex)
func (pm *DefaultPathMatcher) Compile(pattern string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Clear existing patterns
	pm.patterns = make([]string, 0)
	pm.compiledRegex = make([]*regexp.Regexp, 0)

	return pm.addPatternUnsafe(pattern)
}

// Pattern returns the first pattern (for compatibility)
func (pm *DefaultPathMatcher) Pattern() string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if len(pm.patterns) > 0 {
		return pm.patterns[0]
	}
	return ""
}

// Multiple patterns

// AddPattern adds a new pattern to the matcher
func (pm *DefaultPathMatcher) AddPattern(pattern string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	return pm.addPatternUnsafe(pattern)
}

// addPatternUnsafe adds a pattern without locking (internal use)
func (pm *DefaultPathMatcher) addPatternUnsafe(pattern string) error {
	if pattern == "" {
		return ErrPatternEmpty
	}

	// Check if it's a regex pattern (starts with ^ or contains regex special chars)
	if pm.isRegexPattern(pattern) {
		return pm.addRegexPattern(pattern)
	}

	// Treat as glob pattern
	pm.addGlobPattern(pattern)

	return nil
}

// addRegexPattern compiles and adds a regex pattern
func (pm *DefaultPathMatcher) addRegexPattern(pattern string) error {
	flags := ""
	if !pm.caseSensitive {
		flags = "(?i)"
	}

	regex, err := regexp.Compile(flags + pattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern %q: %w", pattern, err)
	}

	pm.compiledRegex = append(pm.compiledRegex, regex)
	return nil
}

// addGlobPattern processes and adds a glob pattern
func (pm *DefaultPathMatcher) addGlobPattern(pattern string) {
	if !pm.caseSensitive {
		pattern = strings.ToLower(pattern)
	}
	pm.patterns = append(pm.patterns, pattern)
}

// RemovePattern removes a pattern from the matcher
func (pm *DefaultPathMatcher) RemovePattern(pattern string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Remove from glob patterns
	for i, p := range pm.patterns {
		if p == pattern || (!pm.caseSensitive && strings.EqualFold(p, pattern)) {
			pm.patterns = append(pm.patterns[:i], pm.patterns[i+1:]...)
			return nil
		}
	}

	// Remove from regex patterns
	for i, regex := range pm.compiledRegex {
		if regex.String() == pattern {
			pm.compiledRegex = append(pm.compiledRegex[:i], pm.compiledRegex[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("%w: %q", ErrPatternNotFound, pattern)
}

// ClearPatterns removes all patterns
func (pm *DefaultPathMatcher) ClearPatterns() error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.patterns = make([]string, 0)
	pm.compiledRegex = make([]*regexp.Regexp, 0)
	return nil
}

// Patterns returns all patterns
func (pm *DefaultPathMatcher) Patterns() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]string, 0, len(pm.patterns)+len(pm.compiledRegex))
	result = append(result, pm.patterns...)

	for _, regex := range pm.compiledRegex {
		result = append(result, regex.String())
	}

	return result
}

// Matching options

// SetCaseSensitive sets case sensitivity for matching
func (pm *DefaultPathMatcher) SetCaseSensitive(sensitive bool) PathMatcher {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.caseSensitive = sensitive
	return pm
}

// SetRecursive sets whether to match recursively
func (pm *DefaultPathMatcher) SetRecursive(recursive bool) PathMatcher {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.recursive = recursive
	return pm
}

// SetMaxDepth sets the maximum depth for recursive matching
func (pm *DefaultPathMatcher) SetMaxDepth(depth int) PathMatcher {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.maxDepth = depth
	return pm
}

// Advanced matching

// MatchAny returns true if any of the paths match
func (pm *DefaultPathMatcher) MatchAny(paths ...string) bool {
	for _, path := range paths {
		if pm.Match(path) {
			return true
		}
	}
	return false
}

// MatchAll returns true if all paths match
func (pm *DefaultPathMatcher) MatchAll(paths ...string) bool {
	if len(paths) == 0 {
		return false
	}

	for _, path := range paths {
		if !pm.Match(path) {
			return false
		}
	}
	return true
}

// Filter returns only the paths that match
func (pm *DefaultPathMatcher) Filter(paths []string) []string {
	result := make([]string, 0)

	for _, path := range paths {
		if pm.Match(path) {
			result = append(result, path)
		}
	}

	return result
}

// FilterPaths returns only the PathBuilders that match
func (pm *DefaultPathMatcher) FilterPaths(paths []PathBuilder) []PathBuilder {
	result := make([]PathBuilder, 0)

	for _, path := range paths {
		if pm.MatchPath(path) {
			result = append(result, path)
		}
	}

	return result
}

// Helper methods

// isRegexPattern determines if a pattern should be treated as regex
func (pm *DefaultPathMatcher) isRegexPattern(pattern string) bool {
	// Simple heuristic: if it starts with ^ or contains regex metacharacters
	// that are not glob metacharacters, treat as regex
	if strings.HasPrefix(pattern, "^") || strings.HasSuffix(pattern, "$") {
		return true
	}

	// Check for regex-specific characters
	regexChars := []string{"(", ")", "[", "]", "{", "}", "|", "+", "\\d", "\\w", "\\s"}
	for _, char := range regexChars {
		if strings.Contains(pattern, char) {
			return true
		}
	}

	return false
}

// matchGlob performs glob matching with case sensitivity consideration
func (pm *DefaultPathMatcher) matchGlob(path, pattern string) bool {
	checkPath := path
	checkPattern := pattern

	if !pm.caseSensitive {
		checkPath = strings.ToLower(path)
		checkPattern = strings.ToLower(pattern)
	}

	// Use filepath.Match for basic glob matching
	matched, err := filepath.Match(checkPattern, filepath.Base(checkPath))
	if err == nil && matched {
		return true
	}

	// For recursive matching, check against the full path
	if pm.recursive {
		matched, err = filepath.Match(checkPattern, checkPath)
		if err == nil && matched {
			return true
		}

		// Check each directory component
		parts := strings.Split(checkPath, string(filepath.Separator))
		for i := range parts {
			subPath := strings.Join(parts[i:], string(filepath.Separator))
			matched, err = filepath.Match(checkPattern, subPath)
			if err == nil && matched {
				return true
			}
		}
	}

	return false
}

// Package-level convenience functions

// Match returns true if path matches pattern
func Match(pattern, path string) bool {
	matcher := NewPathMatcher()
	if err := matcher.AddPattern(pattern); err != nil {
		return false
	}
	return matcher.Match(path)
}

// MatchAny returns true if path matches any of the patterns
func MatchAny(path string, patterns ...string) bool {
	matcher := NewPathMatcher()
	for _, pattern := range patterns {
		if err := matcher.AddPattern(pattern); err != nil {
			return false
		}
	}
	return matcher.Match(path)
}

// FilterByPattern returns only paths that match the pattern
func FilterByPattern(pattern string, paths []string) ([]string, error) {
	matcher, err := NewPathMatcherWithPattern(pattern)
	if err != nil {
		return nil, err
	}
	return matcher.Filter(paths), nil
}

// FilterPathsByPattern returns only PathBuilders that match the pattern
func FilterPathsByPattern(pattern string, paths []PathBuilder) ([]PathBuilder, error) {
	matcher, err := NewPathMatcherWithPattern(pattern)
	if err != nil {
		return nil, err
	}
	return matcher.FilterPaths(paths), nil
}
