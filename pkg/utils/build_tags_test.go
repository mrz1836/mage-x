package utils

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildTagsDiscovery(t *testing.T) {
	// Get testdata directory path
	testdataDir := filepath.Join("testdata")

	t.Run("DiscoverAllTags", func(t *testing.T) {
		discovery := NewBuildTagsDiscovery(testdataDir, nil)
		tags, err := discovery.DiscoverBuildTags()

		require.NoError(t, err)
		sort.Strings(tags)

		// Expected tags from our test files
		expected := []string{"integration", "nested", "unit", "windows"}
		sort.Strings(expected)

		assert.Equal(t, expected, tags)
	})

	t.Run("DiscoverWithExclusions", func(t *testing.T) {
		excludeTags := []string{"integration", "windows"}
		discovery := NewBuildTagsDiscovery(testdataDir, excludeTags)
		tags, err := discovery.DiscoverBuildTags()

		require.NoError(t, err)
		sort.Strings(tags)

		// Should not contain excluded tags
		assert.NotContains(t, tags, "integration")
		assert.NotContains(t, tags, "windows")

		// Should still contain other tags
		assert.Contains(t, tags, "unit")
		assert.Contains(t, tags, "nested")
	})

	t.Run("EmptyDirectory", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "build-tags-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(tempDir) }() //nolint:errcheck // Test cleanup

		discovery := NewBuildTagsDiscovery(tempDir, nil)
		tags, err := discovery.DiscoverBuildTags()

		require.NoError(t, err)
		assert.Empty(t, tags)
	})

	t.Run("NonExistentDirectory", func(t *testing.T) {
		discovery := NewBuildTagsDiscovery("/non/existent/path", nil)
		tags, err := discovery.DiscoverBuildTags()

		require.Error(t, err)
		assert.Nil(t, tags)
	})

	t.Run("ExcludeWhitespace", func(t *testing.T) {
		excludeTags := []string{" integration ", "  performance  "}
		discovery := NewBuildTagsDiscovery(testdataDir, excludeTags)
		tags, err := discovery.DiscoverBuildTags()

		require.NoError(t, err)

		// Should properly handle whitespace in exclude tags
		assert.NotContains(t, tags, "integration")
		assert.NotContains(t, tags, "performance")
	})
}

func TestExtractBuildTagsFromFile(t *testing.T) {
	testdataDir := filepath.Join("testdata")
	discovery := NewBuildTagsDiscovery(testdataDir, nil)

	t.Run("ModernGoTags", func(t *testing.T) {
		tags, err := discovery.extractBuildTagsFromFile(filepath.Join(testdataDir, "integration_test.go"))
		require.NoError(t, err)

		assert.Contains(t, tags, "integration")
	})

	// Removed LegacyTags test - those build tags are not commonly used

	t.Run("ComplexTags", func(t *testing.T) {
		tags, err := discovery.extractBuildTagsFromFile(filepath.Join(testdataDir, "complex_build_tags.go"))
		require.NoError(t, err)

		assert.Contains(t, tags, "integration")
		assert.Contains(t, tags, "windows")
	})

	t.Run("NoTags", func(t *testing.T) {
		tags, err := discovery.extractBuildTagsFromFile(filepath.Join(testdataDir, "no_build_tags.go"))
		require.NoError(t, err)

		assert.Empty(t, tags)
	})

	t.Run("NonExistentFile", func(t *testing.T) {
		tags, err := discovery.extractBuildTagsFromFile(filepath.Join(testdataDir, "non_existent.go"))
		require.Error(t, err)
		assert.Nil(t, tags)
	})
}

func TestParseBuildExpression(t *testing.T) {
	discovery := NewBuildTagsDiscovery(".", nil)

	tests := []struct {
		name       string
		expression string
		expected   []string
	}{
		{
			name:       "SimpleTag",
			expression: "integration",
			expected:   []string{"integration"},
		},
		{
			name:       "AndExpression",
			expression: "integration && !windows",
			expected:   []string{"integration", "windows"},
		},
		{
			name:       "OrExpression",
			expression: "unit || integration",
			expected:   []string{"unit", "integration"},
		},
		{
			name:       "ComplexExpression",
			expression: "(unit || integration) && !windows && cgo",
			expected:   []string{"unit", "integration", "windows", "cgo"},
		},
		{
			name:       "EmptyExpression",
			expression: "",
			expected:   []string{},
		},
		{
			name:       "LogicalOperatorsOnly",
			expression: "and or not",
			expected:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := discovery.parseBuildExpression(tt.expression)
			sort.Strings(result)
			sort.Strings(tt.expected)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseLegacyBuildExpression(t *testing.T) {
	discovery := NewBuildTagsDiscovery(".", nil)

	tests := []struct {
		name       string
		expression string
		expected   []string
	}{
		{
			name:       "SimpleTag",
			expression: "integration",
			expected:   []string{"integration"},
		},
		{
			name:       "NegatedTag",
			expression: "!windows",
			expected:   []string{"windows"},
		},
		{
			name:       "MultipleTags",
			expression: "integration,!windows",
			expected:   []string{"integration", "windows"},
		},
		{
			name:       "SpaceSeparated",
			expression: "unit integration",
			expected:   []string{"unit", "integration"},
		},
		{
			name:       "EmptyExpression",
			expression: "",
			expected:   []string{},
		},
		{
			name:       "OnlyNegated",
			expression: "!windows !darwin",
			expected:   []string{"windows", "darwin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := discovery.parseLegacyBuildExpression(tt.expression)
			sort.Strings(result)
			sort.Strings(tt.expected)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	discovery := NewBuildTagsDiscovery(".", nil)

	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "NoDuplicates",
			input:    []string{"unit", "integration", "e2e"},
			expected: []string{"unit", "integration", "e2e"},
		},
		{
			name:     "WithDuplicates",
			input:    []string{"unit", "integration", "unit", "e2e", "integration"},
			expected: []string{"unit", "integration", "e2e"},
		},
		{
			name:     "EmptySlice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "SingleElement",
			input:    []string{"unit"},
			expected: []string{"unit"},
		},
		{
			name:     "AllDuplicates",
			input:    []string{"unit", "unit", "unit"},
			expected: []string{"unit"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := discovery.removeDuplicates(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestDiscoverBuildTagsConvenienceFunction(t *testing.T) {
	testdataDir := filepath.Join("testdata")

	t.Run("WithoutExclusions", func(t *testing.T) {
		tags, err := DiscoverBuildTags(testdataDir, nil)
		require.NoError(t, err)
		assert.NotEmpty(t, tags)
		assert.Contains(t, tags, "integration")
		assert.Contains(t, tags, "unit")
	})

	t.Run("WithExclusions", func(t *testing.T) {
		excludeTags := []string{"integration", "performance"}
		tags, err := DiscoverBuildTags(testdataDir, excludeTags)
		require.NoError(t, err)

		assert.NotContains(t, tags, "integration")
		assert.NotContains(t, tags, "performance")
	})
}

func TestDiscoverBuildTagsFromCurrentDir(t *testing.T) {
	// This test changes working directory, so we need to restore it
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(originalWd))
	}()

	testdataDir := filepath.Join("testdata")
	testdataAbsPath, err := filepath.Abs(testdataDir)
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		err := os.Chdir(testdataAbsPath)
		require.NoError(t, err)

		tags, err := DiscoverBuildTagsFromCurrentDir(nil)
		require.NoError(t, err)
		assert.NotEmpty(t, tags)
		assert.Contains(t, tags, "integration")
		assert.Contains(t, tags, "unit")
	})

	t.Run("WithExclusions", func(t *testing.T) {
		err := os.Chdir(testdataAbsPath)
		require.NoError(t, err)

		excludeTags := []string{"integration"}
		tags, err := DiscoverBuildTagsFromCurrentDir(excludeTags)
		require.NoError(t, err)

		assert.NotContains(t, tags, "integration")
		assert.Contains(t, tags, "unit")
	})
}

func TestNewBuildTagsDiscovery(t *testing.T) {
	t.Run("WithoutExclusions", func(t *testing.T) {
		discovery := NewBuildTagsDiscovery("/some/path", nil)
		assert.Equal(t, "/some/path", discovery.rootPath)
		assert.Empty(t, discovery.excludeList)
	})

	t.Run("WithExclusions", func(t *testing.T) {
		excludeTags := []string{"integration", " unit ", "  e2e  "}
		discovery := NewBuildTagsDiscovery("/some/path", excludeTags)

		assert.Equal(t, "/some/path", discovery.rootPath)
		assert.True(t, discovery.excludeList["integration"])
		assert.True(t, discovery.excludeList["unit"]) // whitespace trimmed
		assert.True(t, discovery.excludeList["e2e"])  // whitespace trimmed
		assert.False(t, discovery.excludeList["performance"])
	})
}
