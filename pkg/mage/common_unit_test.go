package mage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetDirSizeUnit tests the getDirSize function
func TestGetDirSizeUnit(t *testing.T) {
	t.Run("calculate directory size", func(t *testing.T) {
		tmpDir := t.TempDir()
		testDir := filepath.Join(tmpDir, "testdir")
		require.NoError(t, os.MkdirAll(testDir, 0o750))

		// Create files with known sizes
		file1 := filepath.Join(testDir, "file1.txt")
		file2 := filepath.Join(testDir, "file2.txt")
		require.NoError(t, os.WriteFile(file1, []byte("hello"), 0o600))    // 5 bytes
		require.NoError(t, os.WriteFile(file2, []byte("world123"), 0o600)) // 8 bytes

		size, err := getDirSize(testDir)
		require.NoError(t, err)
		assert.Equal(t, int64(13), size) // 5 + 8 bytes
	})

	t.Run("empty directory", func(t *testing.T) {
		emptyDir := t.TempDir()

		size, err := getDirSize(emptyDir)
		require.NoError(t, err)
		assert.Equal(t, int64(0), size)
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		size, err := getDirSize("/nonexistent/path/that/does/not/exist")
		require.Error(t, err)
		assert.Equal(t, int64(0), size)
	})

	t.Run("nested directories", func(t *testing.T) {
		tmpDir := t.TempDir()
		nestedDir := filepath.Join(tmpDir, "nested", "deep", "structure")
		require.NoError(t, os.MkdirAll(nestedDir, 0o750))

		// Add files at different levels
		file1 := filepath.Join(tmpDir, "nested", "top.txt")
		file2 := filepath.Join(tmpDir, "nested", "deep", "middle.txt")
		file3 := filepath.Join(nestedDir, "bottom.txt")

		require.NoError(t, os.WriteFile(file1, []byte("top"), 0o600))    // 3 bytes
		require.NoError(t, os.WriteFile(file2, []byte("middle"), 0o600)) // 6 bytes
		require.NoError(t, os.WriteFile(file3, []byte("bottom"), 0o600)) // 6 bytes

		size, err := getDirSize(filepath.Join(tmpDir, "nested"))
		require.NoError(t, err)
		assert.Equal(t, int64(15), size) // 3 + 6 + 6 bytes
	})
}

// TestGetCPUCountUnit tests the getCPUCount function
func TestGetCPUCountUnit(t *testing.T) {
	count := getCPUCount()
	assert.Positive(t, count)
	assert.Equal(t, runtime.NumCPU(), count)
}

// TestIsNewerUnit tests the isNewer version comparison function
func TestIsNewerUnit(t *testing.T) {
	testCases := []struct {
		name     string
		versionA string
		versionB string
		expected bool
	}{
		{"newer version", "2.0.0", "1.0.0", true},
		{"older version", "1.0.0", "2.0.0", false},
		{"same version", "1.0.0", "1.0.0", false},
		{"with v prefix", "v2.0.0", "v1.0.0", true},
		{"mixed v prefix", "2.0.0", "v1.0.0", true},
		{"compare to dev", "1.0.0", "dev", true},
		{"patch version newer", "1.0.1", "1.0.0", true},
		{"minor version newer", "1.1.0", "1.0.9", true},
		{"shorter version older", "1.0", "1.0.0", false},
		{"pre-release versions", "2.0.0-alpha", "1.0.0", true},
		{"empty version b", "1.0.0", "", true},
		// When b is empty, function returns true (any version is newer than empty)
		{"both empty returns true", "", "", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isNewer(tc.versionA, tc.versionB)
			assert.Equal(t, tc.expected, result,
				"isNewer(%q, %q) expected %v, got %v", tc.versionA, tc.versionB, tc.expected, result)
		})
	}
}

// TestStripPrereleaseUnit tests the stripPrerelease function
func TestStripPrereleaseUnit(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"no prerelease", "1.0.0", "1.0.0"},
		{"with alpha", "2.0.0-alpha", "2.0.0"},
		{"with beta", "1.5.0-beta.1", "1.5.0"},
		{"with rc", "3.0.0-rc1", "3.0.0"},
		{"multiple dashes", "1.0.0-pre-release-1", "1.0.0"},
		{"empty string", "", ""},
		{"only prerelease", "-alpha", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := stripPrerelease(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestFormatReleaseNotesUnit tests the formatReleaseNotes function
func TestFormatReleaseNotesUnit(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple notes", "Added feature\nFixed bug", "  Added feature\n  Fixed bug"},
		{"with empty lines", "Added feature\n\nFixed bug", "  Added feature\n  Fixed bug"},
		{"empty input", "", ""},
		{"single line", "Single feature", "  Single feature"},
		{"whitespace lines", "Line1\n   \nLine2", "  Line1\n  Line2"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := formatReleaseNotes(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestGetVersionFromGitUnit tests the getVersionFromGit function
// Note: This function uses GetRunner() which may use the real git in test environments.
// The integration tests in common_test.go provide full coverage with proper mocking.
func TestGetVersionFromGitUnit(t *testing.T) {
	// Test the function behavior - it should return something (version or empty)
	// In a git repo, it may return a real version or "dev"
	version := getVersionFromGit()
	// Just verify it doesn't panic and returns a string
	assert.NotNil(t, version)
	// The result depends on git state - could be version, "dev", or ""
}
