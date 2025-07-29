package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/mrz1836/go-mage/pkg/mage/testutil"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// CommonTestSuite defines the test suite for common functions
type CommonTestSuite struct {
	suite.Suite
	env *testutil.TestEnvironment
}

// SetupTest runs before each test
func (ts *CommonTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
}

// TearDownTest runs after each test
func (ts *CommonTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestGetVersion tests the getVersion function
func (ts *CommonTestSuite) TestGetVersion() {
	ts.Run("version from git tag", func() {
		ts.env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("v1.2.3", nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) { SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				version := getVersion()
				require.Equal(ts.T(), "v1.2.3", version)
				return nil
			},
		)

		require.NoError(ts.T(), err)
	})

	ts.Run("version from VERSION file", func() {
		// Create a fresh test environment for this test
		env := testutil.NewTestEnvironment(ts.T())
		defer env.Cleanup()

		// Mock git command to fail
		env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("", fmt.Errorf("not a git repository"))

		// Create VERSION file
		versionPath := filepath.Join(env.TempDir, "VERSION")
		err := os.WriteFile(versionPath, []byte("2.0.0\n"), 0o644)
		require.NoError(ts.T(), err)

		// Change to temp directory
		oldDir, _ := os.Getwd()
		defer os.Chdir(oldDir)
		os.Chdir(env.TempDir)

		err = env.WithMockRunner(
			func(r interface{}) { SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				version := getVersion()
				require.Equal(ts.T(), "2.0.0", version)
				return nil
			},
		)

		require.NoError(ts.T(), err)
	})

	ts.Run("version from config", func() {
		// Create a fresh test environment for this test
		env := testutil.NewTestEnvironment(ts.T())
		defer env.Cleanup()

		// Mock git and file system to fail
		env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("", fmt.Errorf("not a git repository"))

		// Set global config
		originalCfg := cfg
		defer func() { cfg = originalCfg }()
		cfg = &Config{
			Project: ProjectConfig{
				Version: "3.0.0",
			},
		}

		err := env.WithMockRunner(
			func(r interface{}) { SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				version := getVersion()
				require.Equal(ts.T(), "3.0.0", version)
				return nil
			},
		)

		require.NoError(ts.T(), err)
	})

	ts.Run("default dev version", func() {
		// Create a fresh test environment for this test
		env := testutil.NewTestEnvironment(ts.T())
		defer env.Cleanup()

		// Mock git to fail
		env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("", fmt.Errorf("not a git repository"))

		// Ensure no config
		originalCfg := cfg
		defer func() { cfg = originalCfg }()
		cfg = nil

		err := env.WithMockRunner(
			func(r interface{}) { SetRunner(r.(CommandRunner)) },
			func() interface{} { return GetRunner() },
			func() error {
				version := getVersion()
				require.Equal(ts.T(), "dev", version)
				return nil
			},
		)

		require.NoError(ts.T(), err)
	})
}

// TestGetModuleName tests the getModuleName function
func (ts *CommonTestSuite) TestGetModuleName() {
	ts.Run("successful module name retrieval", func() {
		name, err := getModuleName()
		require.NoError(ts.T(), err)
		require.NotEmpty(ts.T(), name)
		// Should contain the test module name
		require.Contains(ts.T(), name, "test/module")
	})
}

// TestGetDirSize tests the getDirSize function
func (ts *CommonTestSuite) TestGetDirSize() {
	ts.Run("calculate directory size", func() {
		// Create test files
		testDir := filepath.Join(ts.env.TempDir, "testdir")
		err := os.MkdirAll(testDir, 0o755)
		require.NoError(ts.T(), err)

		// Create files with known sizes
		file1 := filepath.Join(testDir, "file1.txt")
		file2 := filepath.Join(testDir, "file2.txt")
		err = os.WriteFile(file1, []byte("hello"), 0o644) // 5 bytes
		require.NoError(ts.T(), err)
		err = os.WriteFile(file2, []byte("world123"), 0o644) // 8 bytes
		require.NoError(ts.T(), err)

		size, err := getDirSize(testDir)
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), int64(13), size) // 5 + 8 bytes
	})

	ts.Run("empty directory", func() {
		emptyDir := filepath.Join(ts.env.TempDir, "empty")
		err := os.MkdirAll(emptyDir, 0o755)
		require.NoError(ts.T(), err)

		size, err := getDirSize(emptyDir)
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), int64(0), size)
	})

	ts.Run("nonexistent directory", func() {
		size, err := getDirSize("/nonexistent/path")
		require.Error(ts.T(), err)
		require.Equal(ts.T(), int64(0), size)
	})

	ts.Run("nested directories", func() {
		// Create nested structure
		nestedDir := filepath.Join(ts.env.TempDir, "nested", "deep", "structure")
		err := os.MkdirAll(nestedDir, 0o755)
		require.NoError(ts.T(), err)

		// Add files at different levels
		file1 := filepath.Join(ts.env.TempDir, "nested", "top.txt")
		file2 := filepath.Join(ts.env.TempDir, "nested", "deep", "middle.txt")
		file3 := filepath.Join(nestedDir, "bottom.txt")

		err = os.WriteFile(file1, []byte("top"), 0o644)    // 3 bytes
		require.NoError(ts.T(), err)
		err = os.WriteFile(file2, []byte("middle"), 0o644) // 6 bytes
		require.NoError(ts.T(), err)
		err = os.WriteFile(file3, []byte("bottom"), 0o644) // 6 bytes
		require.NoError(ts.T(), err)

		size, err := getDirSize(filepath.Join(ts.env.TempDir, "nested"))
		require.NoError(ts.T(), err)
		require.Equal(ts.T(), int64(15), size) // 3 + 6 + 6 bytes
	})
}

// TestGetCPUCount tests the getCPUCount function
func (ts *CommonTestSuite) TestGetCPUCount() {
	count := getCPUCount()
	require.Greater(ts.T(), count, 0)
	require.Equal(ts.T(), runtime.NumCPU(), count)
}

// TestIsNewer tests the isNewer version comparison function
func (ts *CommonTestSuite) TestIsNewer() {
	testCases := []struct {
		name     string
		versionA string
		versionB string
		expected bool
	}{
		{
			name:     "newer version",
			versionA: "2.0.0",
			versionB: "1.0.0",
			expected: true,
		},
		{
			name:     "older version",
			versionA: "1.0.0",
			versionB: "2.0.0",
			expected: false,
		},
		{
			name:     "same version",
			versionA: "1.0.0",
			versionB: "1.0.0",
			expected: false,
		},
		{
			name:     "with v prefix",
			versionA: "v2.0.0",
			versionB: "v1.0.0",
			expected: true,
		},
		{
			name:     "mixed v prefix",
			versionA: "2.0.0",
			versionB: "v1.0.0",
			expected: true,
		},
		{
			name:     "compare to dev",
			versionA: "1.0.0",
			versionB: "dev",
			expected: true,
		},
		{
			name:     "patch version newer",
			versionA: "1.0.1",
			versionB: "1.0.0",
			expected: true,
		},
		{
			name:     "minor version newer",
			versionA: "1.1.0",
			versionB: "1.0.9",
			expected: true,
		},
		{
			name:     "longer version newer",
			versionA: "1.0.0.1",
			versionB: "1.0.0",
			expected: true,
		},
		{
			name:     "shorter version older",
			versionA: "1.0",
			versionB: "1.0.0",
			expected: false,
		},
		{
			name:     "pre-release versions",
			versionA: "2.0.0-alpha",
			versionB: "1.0.0",
			expected: true,
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			result := isNewer(tc.versionA, tc.versionB)
			require.Equal(ts.T(), tc.expected, result,
				"isNewer(%q, %q) expected %v, got %v", tc.versionA, tc.versionB, tc.expected, result)
		})
	}
}

// TestFormatReleaseNotes tests the formatReleaseNotes function
func (ts *CommonTestSuite) TestFormatReleaseNotes() {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple release notes",
			input:    "Added new feature\nFixed bug\nImproved performance",
			expected: "  Added new feature\n  Fixed bug\n  Improved performance",
		},
		{
			name:     "with empty lines",
			input:    "Added feature\n\nFixed bug\n\n\nImproved performance",
			expected: "  Added feature\n  Fixed bug\n  Improved performance",
		},
		{
			name:     "with whitespace lines",
			input:    "Added feature\n   \nFixed bug\n\t\nImproved performance",
			expected: "  Added feature\n  Fixed bug\n  Improved performance",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "only empty lines",
			input:    "\n\n\n",
			expected: "",
		},
		{
			name:     "single line",
			input:    "Single feature",
			expected: "  Single feature",
		},
		{
			name:     "already indented",
			input:    "  Already indented\n    More indented",
			expected: "    Already indented\n      More indented",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			result := formatReleaseNotes(tc.input)
			require.Equal(ts.T(), tc.expected, result)
		})
	}
}

// TestFormatDuration tests the formatDuration function
func (ts *CommonTestSuite) TestFormatDuration() {
	testCases := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "milliseconds",
			duration: 500 * time.Millisecond,
			expected: "500ms",
		},
		{
			name:     "less than millisecond",
			duration: 500 * time.Microsecond,
			expected: "0ms",
		},
		{
			name:     "seconds",
			duration: 5*time.Second + 500*time.Millisecond,
			expected: "5.5s",
		},
		{
			name:     "minutes",
			duration: 5*time.Minute + 30*time.Second,
			expected: "5.5m",
		},
		{
			name:     "hours",
			duration: 2*time.Hour + 30*time.Minute,
			expected: "2.5h",
		},
		{
			name:     "exact second",
			duration: 10 * time.Second,
			expected: "10.0s",
		},
		{
			name:     "exact minute",
			duration: 3 * time.Minute,
			expected: "3.0m",
		},
		{
			name:     "exact hour",
			duration: 1 * time.Hour,
			expected: "1.0h",
		},
		{
			name:     "zero duration",
			duration: 0,
			expected: "0ms",
		},
		{
			name:     "large duration",
			duration: 25*time.Hour + 30*time.Minute,
			expected: "25.5h",
		},
	}

	for _, tc := range testCases {
		ts.Run(tc.name, func() {
			result := formatDuration(tc.duration)
			require.Equal(ts.T(), tc.expected, result)
		})
	}
}

// TestCommonTestSuite runs the test suite
func TestCommonTestSuite(t *testing.T) {
	suite.Run(t, new(CommonTestSuite))
}