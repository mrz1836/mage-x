package paths

import (
	"sort"
	"sync"
)

// DefaultPathSet implements PathSet using a map for efficient operations
type DefaultPathSet struct {
	mu    sync.RWMutex
	paths map[string]bool
}

// NewPathSet creates a new empty path set
func NewPathSet() *DefaultPathSet {
	return &DefaultPathSet{
		paths: make(map[string]bool),
	}
}

// NewPathSetFromPaths creates a new path set from a slice of paths
func NewPathSetFromPaths(paths []string) *DefaultPathSet {
	set := NewPathSet()
	for _, path := range paths {
		set.Add(path)
	}
	return set
}

// NewPathSetFromPathBuilders creates a new path set from PathBuilders
func NewPathSetFromPathBuilders(paths []PathBuilder) *DefaultPathSet {
	set := NewPathSet()
	for _, path := range paths {
		set.AddPath(path)
	}
	return set
}

// Set operations

// Add adds a path to the set
func (ps *DefaultPathSet) Add(path string) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.paths[path] {
		return false // Already exists
	}

	ps.paths[path] = true
	return true
}

// AddPath adds a PathBuilder to the set
func (ps *DefaultPathSet) AddPath(path PathBuilder) bool {
	return ps.Add(path.String())
}

// Remove removes a path from the set
func (ps *DefaultPathSet) Remove(path string) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if !ps.paths[path] {
		return false // Doesn't exist
	}

	delete(ps.paths, path)
	return true
}

// RemovePath removes a PathBuilder from the set
func (ps *DefaultPathSet) RemovePath(path PathBuilder) bool {
	return ps.Remove(path.String())
}

// Contains returns true if the path is in the set
func (ps *DefaultPathSet) Contains(path string) bool {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return ps.paths[path]
}

// ContainsPath returns true if the PathBuilder is in the set
func (ps *DefaultPathSet) ContainsPath(path PathBuilder) bool {
	return ps.Contains(path.String())
}

// Clear removes all paths from the set
func (ps *DefaultPathSet) Clear() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.paths = make(map[string]bool)
	return nil
}

// Set information

// Size returns the number of paths in the set
func (ps *DefaultPathSet) Size() int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	return len(ps.paths)
}

// IsEmpty returns true if the set is empty
func (ps *DefaultPathSet) IsEmpty() bool {
	return ps.Size() == 0
}

// Paths returns all paths as a sorted slice
func (ps *DefaultPathSet) Paths() []string {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	result := make([]string, 0, len(ps.paths))
	for path := range ps.paths {
		result = append(result, path)
	}

	sort.Strings(result)
	return result
}

// PathBuilders returns all paths as PathBuilders
func (ps *DefaultPathSet) PathBuilders() []PathBuilder {
	paths := ps.Paths()
	result := make([]PathBuilder, 0, len(paths))

	for _, path := range paths {
		result = append(result, NewPathBuilder(path))
	}

	return result
}

// Set operations

// Union returns a new set containing paths from both sets
func (ps *DefaultPathSet) Union(other PathSet) PathSet {
	result := NewPathSet()

	// Add all paths from this set
	for _, path := range ps.Paths() {
		result.Add(path)
	}

	// Add all paths from other set
	for _, path := range other.Paths() {
		result.Add(path)
	}

	return result
}

// Intersection returns a new set containing paths common to both sets
func (ps *DefaultPathSet) Intersection(other PathSet) PathSet {
	result := NewPathSet()

	// Find common paths
	for _, path := range ps.Paths() {
		if other.Contains(path) {
			result.Add(path)
		}
	}

	return result
}

// Difference returns a new set containing paths in this set but not in other
func (ps *DefaultPathSet) Difference(other PathSet) PathSet {
	result := NewPathSet()

	// Find paths only in this set
	for _, path := range ps.Paths() {
		if !other.Contains(path) {
			result.Add(path)
		}
	}

	return result
}

// SymmetricDifference returns a new set containing paths in either set but not both
func (ps *DefaultPathSet) SymmetricDifference(other PathSet) PathSet {
	result := NewPathSet()

	// Add paths from this set that are not in other
	for _, path := range ps.Paths() {
		if !other.Contains(path) {
			result.Add(path)
		}
	}

	// Add paths from other set that are not in this
	for _, path := range other.Paths() {
		if !ps.Contains(path) {
			result.Add(path)
		}
	}

	return result
}

// Filtering

// Filter returns a new set containing only paths that match the predicate
func (ps *DefaultPathSet) Filter(predicate func(string) bool) PathSet {
	result := NewPathSet()

	for _, path := range ps.Paths() {
		if predicate(path) {
			result.Add(path)
		}
	}

	return result
}

// FilterPaths returns a new set containing only PathBuilders that match the predicate
func (ps *DefaultPathSet) FilterPaths(predicate func(PathBuilder) bool) PathSet {
	result := NewPathSet()

	for _, path := range ps.PathBuilders() {
		if predicate(path) {
			result.Add(path.String())
		}
	}

	return result
}

// Iteration

// ForEach executes a function for each path in the set
func (ps *DefaultPathSet) ForEach(fn func(string) error) error {
	for _, path := range ps.Paths() {
		if err := fn(path); err != nil {
			return err
		}
	}
	return nil
}

// ForEachPath executes a function for each PathBuilder in the set
func (ps *DefaultPathSet) ForEachPath(fn func(PathBuilder) error) error {
	for _, path := range ps.PathBuilders() {
		if err := fn(path); err != nil {
			return err
		}
	}
	return nil
}

// Package-level convenience functions

// UnionSets returns the union of multiple path sets
func UnionSets(sets ...PathSet) PathSet {
	result := NewPathSet()

	for _, set := range sets {
		for _, path := range set.Paths() {
			result.Add(path)
		}
	}

	return result
}

// IntersectionSets returns the intersection of multiple path sets
func IntersectionSets(sets ...PathSet) PathSet {
	if len(sets) == 0 {
		return NewPathSet()
	}

	result := NewPathSet()

	// Start with paths from the first set
	firstSet := sets[0]
	for _, path := range firstSet.Paths() {
		// Check if path exists in all other sets
		inAll := true
		for i := 1; i < len(sets); i++ {
			if !sets[i].Contains(path) {
				inAll = false
				break
			}
		}

		if inAll {
			result.Add(path)
		}
	}

	return result
}

// PathsToSet converts a slice of paths to a PathSet
func PathsToSet(paths []string) PathSet {
	return NewPathSetFromPaths(paths)
}

// PathBuildersToSet converts a slice of PathBuilders to a PathSet
func PathBuildersToSet(paths []PathBuilder) PathSet {
	return NewPathSetFromPathBuilders(paths)
}
