package paths

import (
	"errors"
	"sort"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test errors for ForEach iteration tests
var (
	errStopIteration = errors.New("stop iteration")
	errStop          = errors.New("stop")
)

// =============================================================================
// Basic Set Operations Tests
// =============================================================================

// TestSet_BasicOperations tests Add, Remove, Contains, Clear, Size, and IsEmpty.
// These are the fundamental operations for managing paths in a set.
func TestSet_BasicOperations(t *testing.T) {
	t.Run("Add and Contains", func(t *testing.T) {
		set := NewPathSet()

		// Add returns true for new path
		added := set.Add("/path/to/file1.txt")
		assert.True(t, added, "Add should return true for new path")

		// Contains returns true for added path
		assert.True(t, set.Contains("/path/to/file1.txt"))
		assert.False(t, set.Contains("/path/to/nonexistent.txt"))
	})

	t.Run("Add duplicate returns false", func(t *testing.T) {
		set := NewPathSet()

		set.Add("/path/to/file.txt")
		added := set.Add("/path/to/file.txt")
		assert.False(t, added, "Add should return false for duplicate")
		assert.Equal(t, 1, set.Size(), "Size should remain 1 after duplicate add")
	})

	t.Run("Remove existing path", func(t *testing.T) {
		set := NewPathSet()
		set.Add("/path/to/file.txt")

		removed := set.Remove("/path/to/file.txt")
		assert.True(t, removed, "Remove should return true for existing path")
		assert.False(t, set.Contains("/path/to/file.txt"))
	})

	t.Run("Remove non-existent path returns false", func(t *testing.T) {
		set := NewPathSet()

		removed := set.Remove("/nonexistent.txt")
		assert.False(t, removed, "Remove should return false for non-existent path")
	})

	t.Run("Clear removes all paths", func(t *testing.T) {
		set := NewPathSet()
		set.Add("/a.txt")
		set.Add("/b.txt")
		set.Add("/c.txt")

		err := set.Clear()
		require.NoError(t, err)
		assert.Equal(t, 0, set.Size())
		assert.True(t, set.IsEmpty())
	})

	t.Run("Size returns correct count", func(t *testing.T) {
		set := NewPathSet()
		assert.Equal(t, 0, set.Size())

		set.Add("/a.txt")
		assert.Equal(t, 1, set.Size())

		set.Add("/b.txt")
		assert.Equal(t, 2, set.Size())

		set.Remove("/a.txt")
		assert.Equal(t, 1, set.Size())
	})

	t.Run("IsEmpty returns correct state", func(t *testing.T) {
		set := NewPathSet()
		assert.True(t, set.IsEmpty())

		set.Add("/file.txt")
		assert.False(t, set.IsEmpty())

		set.Remove("/file.txt")
		assert.True(t, set.IsEmpty())
	})
}

// TestSet_PathBuilderOperations tests AddPath, RemovePath, and ContainsPath.
// These are convenience methods that work with PathBuilder instead of strings.
func TestSet_PathBuilderOperations(t *testing.T) {
	t.Run("AddPath with PathBuilder", func(t *testing.T) {
		set := NewPathSet()
		pb := NewPathBuilder("/path/to/file.txt")

		added := set.AddPath(pb)
		assert.True(t, added)
		assert.True(t, set.Contains("/path/to/file.txt"))
	})

	t.Run("RemovePath with PathBuilder", func(t *testing.T) {
		set := NewPathSet()
		set.Add("/path/to/file.txt")

		pb := NewPathBuilder("/path/to/file.txt")
		removed := set.RemovePath(pb)
		assert.True(t, removed)
		assert.False(t, set.Contains("/path/to/file.txt"))
	})

	t.Run("ContainsPath with PathBuilder", func(t *testing.T) {
		set := NewPathSet()
		set.Add("/path/to/file.txt")

		pb := NewPathBuilder("/path/to/file.txt")
		assert.True(t, set.ContainsPath(pb))

		pb2 := NewPathBuilder("/nonexistent.txt")
		assert.False(t, set.ContainsPath(pb2))
	})
}

// =============================================================================
// Set Creation Tests
// =============================================================================

// TestSet_Creation tests various ways to create path sets.
func TestSet_Creation(t *testing.T) {
	t.Run("NewPathSet creates empty set", func(t *testing.T) {
		set := NewPathSet()
		assert.NotNil(t, set)
		assert.True(t, set.IsEmpty())
	})

	t.Run("NewPathSetFromPaths creates populated set", func(t *testing.T) {
		paths := []string{"/a.txt", "/b.txt", "/c.txt"}
		set := NewPathSetFromPaths(paths)

		assert.Equal(t, 3, set.Size())
		assert.True(t, set.Contains("/a.txt"))
		assert.True(t, set.Contains("/b.txt"))
		assert.True(t, set.Contains("/c.txt"))
	})

	t.Run("NewPathSetFromPaths handles duplicates", func(t *testing.T) {
		paths := []string{"/a.txt", "/b.txt", "/a.txt", "/b.txt"}
		set := NewPathSetFromPaths(paths)

		assert.Equal(t, 2, set.Size())
	})

	t.Run("NewPathSetFromPaths handles empty slice", func(t *testing.T) {
		set := NewPathSetFromPaths([]string{})
		assert.True(t, set.IsEmpty())
	})

	t.Run("NewPathSetFromPathBuilders creates populated set", func(t *testing.T) {
		pbs := []PathBuilder{
			NewPathBuilder("/a.txt"),
			NewPathBuilder("/b.txt"),
		}
		set := NewPathSetFromPathBuilders(pbs)

		assert.Equal(t, 2, set.Size())
		assert.True(t, set.Contains("/a.txt"))
		assert.True(t, set.Contains("/b.txt"))
	})
}

// =============================================================================
// Paths and PathBuilders Retrieval Tests
// =============================================================================

// TestSet_PathsRetrieval tests Paths and PathBuilders methods.
func TestSet_PathsRetrieval(t *testing.T) {
	t.Run("Paths returns sorted slice", func(t *testing.T) {
		set := NewPathSet()
		set.Add("/c.txt")
		set.Add("/a.txt")
		set.Add("/b.txt")

		paths := set.Paths()
		assert.Len(t, paths, 3)
		assert.True(t, sort.StringsAreSorted(paths), "Paths should be sorted")
		assert.Equal(t, []string{"/a.txt", "/b.txt", "/c.txt"}, paths)
	})

	t.Run("Paths returns empty slice for empty set", func(t *testing.T) {
		set := NewPathSet()
		paths := set.Paths()
		assert.Empty(t, paths)
		assert.NotNil(t, paths, "Should return empty slice, not nil")
	})

	t.Run("PathBuilders returns PathBuilder slice", func(t *testing.T) {
		set := NewPathSet()
		set.Add("/a.txt")
		set.Add("/b.txt")

		pbs := set.PathBuilders()
		assert.Len(t, pbs, 2)

		// Verify they're sorted
		assert.Equal(t, "/a.txt", pbs[0].String())
		assert.Equal(t, "/b.txt", pbs[1].String())
	})

	t.Run("PathBuilders returns empty slice for empty set", func(t *testing.T) {
		set := NewPathSet()
		pbs := set.PathBuilders()
		assert.Empty(t, pbs)
		assert.NotNil(t, pbs)
	})
}

// =============================================================================
// Union Tests
// =============================================================================

// TestSet_Union tests the Union method which combines two sets.
func TestSet_Union(t *testing.T) {
	t.Run("union of disjoint sets", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})
		set2 := NewPathSetFromPaths([]string{"/c.txt", "/d.txt"})

		union := set1.Union(set2)
		assert.Equal(t, 4, union.Size())
		assert.True(t, union.Contains("/a.txt"))
		assert.True(t, union.Contains("/b.txt"))
		assert.True(t, union.Contains("/c.txt"))
		assert.True(t, union.Contains("/d.txt"))
	})

	t.Run("union of overlapping sets", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt", "/c.txt"})
		set2 := NewPathSetFromPaths([]string{"/b.txt", "/c.txt", "/d.txt"})

		union := set1.Union(set2)
		assert.Equal(t, 4, union.Size(), "Union should not have duplicates")
	})

	t.Run("union with empty set", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})
		set2 := NewPathSet()

		union := set1.Union(set2)
		assert.Equal(t, 2, union.Size())
	})

	t.Run("union of empty sets", func(t *testing.T) {
		set1 := NewPathSet()
		set2 := NewPathSet()

		union := set1.Union(set2)
		assert.True(t, union.IsEmpty())
	})

	t.Run("self-union is idempotent", func(t *testing.T) {
		set := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})

		union := set.Union(set)
		assert.Equal(t, 2, union.Size())
	})

	t.Run("union does not modify original sets", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt"})
		set2 := NewPathSetFromPaths([]string{"/b.txt"})

		_ = set1.Union(set2)

		assert.Equal(t, 1, set1.Size())
		assert.Equal(t, 1, set2.Size())
	})
}

// =============================================================================
// Intersection Tests
// =============================================================================

// TestSet_Intersection tests the Intersection method which finds common paths.
func TestSet_Intersection(t *testing.T) {
	t.Run("intersection of overlapping sets", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt", "/c.txt"})
		set2 := NewPathSetFromPaths([]string{"/b.txt", "/c.txt", "/d.txt"})

		intersection := set1.Intersection(set2)
		assert.Equal(t, 2, intersection.Size())
		assert.True(t, intersection.Contains("/b.txt"))
		assert.True(t, intersection.Contains("/c.txt"))
		assert.False(t, intersection.Contains("/a.txt"))
		assert.False(t, intersection.Contains("/d.txt"))
	})

	t.Run("intersection of disjoint sets is empty", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})
		set2 := NewPathSetFromPaths([]string{"/c.txt", "/d.txt"})

		intersection := set1.Intersection(set2)
		assert.True(t, intersection.IsEmpty())
	})

	t.Run("intersection with empty set is empty", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})
		set2 := NewPathSet()

		intersection := set1.Intersection(set2)
		assert.True(t, intersection.IsEmpty())
	})

	t.Run("self-intersection is identity", func(t *testing.T) {
		set := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})

		intersection := set.Intersection(set)
		assert.Equal(t, 2, intersection.Size())
		assert.True(t, intersection.Contains("/a.txt"))
		assert.True(t, intersection.Contains("/b.txt"))
	})

	t.Run("intersection of empty sets is empty", func(t *testing.T) {
		set1 := NewPathSet()
		set2 := NewPathSet()

		intersection := set1.Intersection(set2)
		assert.True(t, intersection.IsEmpty())
	})
}

// =============================================================================
// Difference Tests
// =============================================================================

// TestSet_Difference tests the Difference method (A - B).
func TestSet_Difference(t *testing.T) {
	t.Run("difference of overlapping sets", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt", "/c.txt"})
		set2 := NewPathSetFromPaths([]string{"/b.txt", "/c.txt", "/d.txt"})

		diff := set1.Difference(set2)
		assert.Equal(t, 1, diff.Size())
		assert.True(t, diff.Contains("/a.txt"))
	})

	t.Run("difference of disjoint sets is first set", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})
		set2 := NewPathSetFromPaths([]string{"/c.txt", "/d.txt"})

		diff := set1.Difference(set2)
		assert.Equal(t, 2, diff.Size())
		assert.True(t, diff.Contains("/a.txt"))
		assert.True(t, diff.Contains("/b.txt"))
	})

	t.Run("difference with superset is empty", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})
		set2 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt", "/c.txt"})

		diff := set1.Difference(set2)
		assert.True(t, diff.IsEmpty())
	})

	t.Run("self-difference is empty", func(t *testing.T) {
		set := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})

		diff := set.Difference(set)
		assert.True(t, diff.IsEmpty())
	})

	t.Run("difference with empty set is original set", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})
		set2 := NewPathSet()

		diff := set1.Difference(set2)
		assert.Equal(t, 2, diff.Size())
	})

	t.Run("empty set difference is empty", func(t *testing.T) {
		set1 := NewPathSet()
		set2 := NewPathSetFromPaths([]string{"/a.txt"})

		diff := set1.Difference(set2)
		assert.True(t, diff.IsEmpty())
	})
}

// =============================================================================
// Symmetric Difference Tests
// =============================================================================

// TestSet_SymmetricDifference tests paths in either set but not both.
func TestSet_SymmetricDifference(t *testing.T) {
	t.Run("symmetric difference of overlapping sets", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt", "/c.txt"})
		set2 := NewPathSetFromPaths([]string{"/b.txt", "/c.txt", "/d.txt"})

		symDiff := set1.SymmetricDifference(set2)
		assert.Equal(t, 2, symDiff.Size())
		assert.True(t, symDiff.Contains("/a.txt"))
		assert.True(t, symDiff.Contains("/d.txt"))
		assert.False(t, symDiff.Contains("/b.txt"))
		assert.False(t, symDiff.Contains("/c.txt"))
	})

	t.Run("symmetric difference of disjoint sets is union", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})
		set2 := NewPathSetFromPaths([]string{"/c.txt", "/d.txt"})

		symDiff := set1.SymmetricDifference(set2)
		assert.Equal(t, 4, symDiff.Size())
	})

	t.Run("self-symmetric difference is empty", func(t *testing.T) {
		set := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})

		symDiff := set.SymmetricDifference(set)
		assert.True(t, symDiff.IsEmpty())
	})

	t.Run("symmetric difference with empty set is original set", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})
		set2 := NewPathSet()

		symDiff := set1.SymmetricDifference(set2)
		assert.Equal(t, 2, symDiff.Size())
	})
}

// =============================================================================
// Filter Tests
// =============================================================================

// TestSet_Filter tests filtering paths by a predicate function.
func TestSet_Filter(t *testing.T) {
	t.Run("filter with matching predicate", func(t *testing.T) {
		set := NewPathSetFromPaths([]string{
			"/file.txt",
			"/file.go",
			"/file.md",
			"/other.txt",
		})

		// Filter for .txt files
		filtered := set.Filter(func(path string) bool {
			return len(path) > 4 && path[len(path)-4:] == ".txt"
		})

		assert.Equal(t, 2, filtered.Size())
		assert.True(t, filtered.Contains("/file.txt"))
		assert.True(t, filtered.Contains("/other.txt"))
	})

	t.Run("filter with no matches", func(t *testing.T) {
		set := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})

		filtered := set.Filter(func(path string) bool {
			return false // Never match
		})

		assert.True(t, filtered.IsEmpty())
	})

	t.Run("filter with all matches", func(t *testing.T) {
		set := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})

		filtered := set.Filter(func(path string) bool {
			return true // Always match
		})

		assert.Equal(t, 2, filtered.Size())
	})

	t.Run("filter empty set", func(t *testing.T) {
		set := NewPathSet()

		filtered := set.Filter(func(path string) bool {
			return true
		})

		assert.True(t, filtered.IsEmpty())
	})

	t.Run("filter does not modify original", func(t *testing.T) {
		set := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})

		_ = set.Filter(func(path string) bool {
			return path == "/a.txt"
		})

		assert.Equal(t, 2, set.Size())
	})
}

// TestSet_FilterPaths tests filtering using PathBuilder predicate.
func TestSet_FilterPaths(t *testing.T) {
	set := NewPathSetFromPaths([]string{
		"/home/user/file.go",
		"/home/user/file.txt",
		"/var/log/app.log",
	})

	// Filter paths containing "user"
	filtered := set.FilterPaths(func(pb PathBuilder) bool {
		return pb.String() != "" && len(pb.String()) > 6 &&
			pb.String()[:11] == "/home/user/"
	})

	assert.Equal(t, 2, filtered.Size())
}

// =============================================================================
// ForEach Tests
// =============================================================================

// TestSet_ForEach tests iteration over paths.
func TestSet_ForEach(t *testing.T) {
	t.Run("ForEach iterates all paths", func(t *testing.T) {
		set := NewPathSetFromPaths([]string{"/a.txt", "/b.txt", "/c.txt"})

		var visited []string
		err := set.ForEach(func(path string) error {
			visited = append(visited, path)
			return nil
		})

		require.NoError(t, err)
		assert.Len(t, visited, 3)
		sort.Strings(visited)
		assert.Equal(t, []string{"/a.txt", "/b.txt", "/c.txt"}, visited)
	})

	t.Run("ForEach stops on error", func(t *testing.T) {
		set := NewPathSetFromPaths([]string{"/a.txt", "/b.txt", "/c.txt"})

		count := 0

		err := set.ForEach(func(path string) error {
			count++
			if count == 2 {
				return errStopIteration
			}
			return nil
		})

		require.Error(t, err)
		require.ErrorIs(t, err, errStopIteration)
		assert.Equal(t, 2, count)
	})

	t.Run("ForEach on empty set", func(t *testing.T) {
		set := NewPathSet()

		called := false
		err := set.ForEach(func(path string) error {
			called = true
			return nil
		})

		require.NoError(t, err)
		assert.False(t, called)
	})
}

// TestSet_ForEachPath tests iteration with PathBuilder.
func TestSet_ForEachPath(t *testing.T) {
	t.Run("ForEachPath iterates all paths", func(t *testing.T) {
		set := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})

		var visited []string
		err := set.ForEachPath(func(pb PathBuilder) error {
			visited = append(visited, pb.String())
			return nil
		})

		require.NoError(t, err)
		assert.Len(t, visited, 2)
	})

	t.Run("ForEachPath stops on error", func(t *testing.T) {
		set := NewPathSetFromPaths([]string{"/a.txt", "/b.txt", "/c.txt"})

		count := 0

		err := set.ForEachPath(func(pb PathBuilder) error {
			count++
			return errStop
		})

		require.Error(t, err)
		require.ErrorIs(t, err, errStop)
		assert.Equal(t, 1, count)
	})
}

// =============================================================================
// Package-Level Functions Tests
// =============================================================================

// TestSet_PackageFunctions tests UnionSets, IntersectionSets, ToSet, PathBuildersToSet.
func TestSet_PackageFunctions(t *testing.T) {
	t.Run("UnionSets with multiple sets", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt"})
		set2 := NewPathSetFromPaths([]string{"/b.txt"})
		set3 := NewPathSetFromPaths([]string{"/c.txt"})

		union := UnionSets(set1, set2, set3)
		assert.Equal(t, 3, union.Size())
	})

	t.Run("UnionSets with empty slice", func(t *testing.T) {
		union := UnionSets()
		assert.True(t, union.IsEmpty())
	})

	t.Run("UnionSets with single set", func(t *testing.T) {
		set := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})
		union := UnionSets(set)
		assert.Equal(t, 2, union.Size())
	})

	t.Run("UnionSets with overlapping sets", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt"})
		set2 := NewPathSetFromPaths([]string{"/b.txt", "/c.txt"})
		set3 := NewPathSetFromPaths([]string{"/c.txt", "/d.txt"})

		union := UnionSets(set1, set2, set3)
		assert.Equal(t, 4, union.Size())
	})

	t.Run("IntersectionSets with multiple sets", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt", "/b.txt", "/c.txt"})
		set2 := NewPathSetFromPaths([]string{"/b.txt", "/c.txt", "/d.txt"})
		set3 := NewPathSetFromPaths([]string{"/c.txt", "/d.txt", "/e.txt"})

		intersection := IntersectionSets(set1, set2, set3)
		assert.Equal(t, 1, intersection.Size())
		assert.True(t, intersection.Contains("/c.txt"))
	})

	t.Run("IntersectionSets with empty slice", func(t *testing.T) {
		intersection := IntersectionSets()
		assert.True(t, intersection.IsEmpty())
	})

	t.Run("IntersectionSets with no common elements", func(t *testing.T) {
		set1 := NewPathSetFromPaths([]string{"/a.txt"})
		set2 := NewPathSetFromPaths([]string{"/b.txt"})
		set3 := NewPathSetFromPaths([]string{"/c.txt"})

		intersection := IntersectionSets(set1, set2, set3)
		assert.True(t, intersection.IsEmpty())
	})

	t.Run("ToSet creates PathSet from strings", func(t *testing.T) {
		set := ToSet([]string{"/a.txt", "/b.txt"})
		assert.Equal(t, 2, set.Size())
		assert.True(t, set.Contains("/a.txt"))
	})

	t.Run("PathBuildersToSet creates PathSet from PathBuilders", func(t *testing.T) {
		pbs := []PathBuilder{
			NewPathBuilder("/a.txt"),
			NewPathBuilder("/b.txt"),
		}
		set := PathBuildersToSet(pbs)
		assert.Equal(t, 2, set.Size())
	})
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

// TestSet_ConcurrentAccess tests thread-safety of the set.
func TestSet_ConcurrentAccess(t *testing.T) {
	set := NewPathSet()
	var wg sync.WaitGroup

	// Concurrent adds
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			set.Add("/path/" + string(rune('a'+n%26)) + ".txt")
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = set.Size()
			_ = set.Paths()
			_ = set.Contains("/path/a.txt")
		}()
	}

	wg.Wait()

	// Verify set is in a valid state
	assert.True(t, set.Size() > 0 && set.Size() <= 26, "Size should be between 1 and 26")
}

// =============================================================================
// Edge Cases Tests
// =============================================================================

// TestSet_EdgeCases tests edge cases and boundary conditions.
func TestSet_EdgeCases(t *testing.T) {
	t.Run("empty string path", func(t *testing.T) {
		set := NewPathSet()
		added := set.Add("")
		assert.True(t, added, "Should allow empty string path")
		assert.True(t, set.Contains(""))
	})

	t.Run("path with special characters", func(t *testing.T) {
		set := NewPathSet()
		specialPaths := []string{
			"/path with spaces/file.txt",
			"/path/to/file name.txt",
			"/unicode/путь/файл.txt",
			"/symbols/!@#$%^&()_+.txt",
		}

		for _, path := range specialPaths {
			set.Add(path)
		}

		assert.Equal(t, len(specialPaths), set.Size())
		for _, path := range specialPaths {
			assert.True(t, set.Contains(path), "Should contain: %s", path)
		}
	})

	t.Run("very long path", func(t *testing.T) {
		set := NewPathSet()

		// Create a very long path
		longPath := "/"
		for i := 0; i < 100; i++ {
			longPath += "very_long_directory_name/"
		}
		longPath += "file.txt"

		added := set.Add(longPath)
		assert.True(t, added)
		assert.True(t, set.Contains(longPath))
	})

	t.Run("add and remove same path repeatedly", func(t *testing.T) {
		set := NewPathSet()
		path := "/test.txt"

		for i := 0; i < 10; i++ {
			set.Add(path)
			assert.True(t, set.Contains(path))
			set.Remove(path)
			assert.False(t, set.Contains(path))
		}

		assert.True(t, set.IsEmpty())
	})

	t.Run("paths that differ only in trailing slash", func(t *testing.T) {
		set := NewPathSet()
		set.Add("/path/to/dir")
		set.Add("/path/to/dir/")

		// These are treated as different paths
		assert.Equal(t, 2, set.Size())
		assert.True(t, set.Contains("/path/to/dir"))
		assert.True(t, set.Contains("/path/to/dir/"))
	})
}
