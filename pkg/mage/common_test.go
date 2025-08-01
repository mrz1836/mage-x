package mage

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
	"github.com/stretchr/testify/suite"
)

// Static test errors to satisfy err113 linter
var (
	errNotGitRepository = errors.New("not a git repository")
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
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				version := getVersion()
				ts.Require().Equal("v1.2.3", version)
				return nil
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("version from VERSION file", func() {
		// Create a fresh test environment for this test
		env := testutil.NewTestEnvironment(ts.T())
		defer env.Cleanup()

		// Mock git command to fail
		env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("", errNotGitRepository)

		// Create VERSION file
		versionPath := filepath.Join(env.TempDir, "VERSION")
		err := os.WriteFile(versionPath, []byte("2.0.0\n"), 0o600)
		ts.Require().NoError(err)

		// Change to temp directory
		oldDir, err := os.Getwd()
		ts.Require().NoError(err)
		defer func() {
			if chdirErr := os.Chdir(oldDir); chdirErr != nil {
				ts.T().Logf("Failed to restore working directory: %v", chdirErr)
			}
		}()
		err = os.Chdir(env.TempDir)
		ts.Require().NoError(err)

		err = env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				version := getVersion()
				ts.Require().Equal("2.0.0", version)
				return nil
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("version from config", func() {
		// Create a fresh test environment for this test
		env := testutil.NewTestEnvironment(ts.T())
		defer env.Cleanup()

		// Mock git and file system to fail
		env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("", errNotGitRepository)

		// Set global config for test
		TestResetConfig()
		cfg = &Config{
			Project: ProjectConfig{
				Version: "3.0.0",
			},
		}

		err := env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				version := getVersion()
				ts.Require().Equal("3.0.0", version)
				return nil
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("default dev version", func() {
		// Create a fresh test environment for this test
		env := testutil.NewTestEnvironment(ts.T())
		defer env.Cleanup()

		// Mock git to fail
		env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("", errNotGitRepository)

		// Reset config for test
		TestResetConfig()

		err := env.WithMockRunner(
			func(r interface{}) error {
				return SetRunner(r.(CommandRunner)) //nolint:errcheck // Test setup function returns error
			},
			func() interface{} { return GetRunner() },
			func() error {
				version := getVersion()
				ts.Require().Equal("dev", version)
				return nil
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGetModuleName tests the getModuleName function
func (ts *CommonTestSuite) TestGetModuleName() {
	ts.Run("successful module name retrieval", func() {
		name, err := getModuleName()
		ts.Require().NoError(err)
		ts.Require().NotEmpty(name)
		// Should contain the test module name
		ts.Require().Contains(name, "test/module")
	})
}

// TestGetDirSize tests the getDirSize function
func (ts *CommonTestSuite) TestGetDirSize() {
	ts.Run("calculate directory size", func() {
		// Create test files
		testDir := filepath.Join(ts.env.TempDir, "testdir")
		err := os.MkdirAll(testDir, 0o750)
		ts.Require().NoError(err)

		// Create files with known sizes
		file1 := filepath.Join(testDir, "file1.txt")
		file2 := filepath.Join(testDir, "file2.txt")
		err = os.WriteFile(file1, []byte("hello"), 0o600) // 5 bytes
		ts.Require().NoError(err)
		err = os.WriteFile(file2, []byte("world123"), 0o600) // 8 bytes
		ts.Require().NoError(err)

		size, err := getDirSize(testDir)
		ts.Require().NoError(err)
		ts.Require().Equal(int64(13), size) // 5 + 8 bytes
	})

	ts.Run("empty directory", func() {
		emptyDir := filepath.Join(ts.env.TempDir, "empty")
		err := os.MkdirAll(emptyDir, 0o750)
		ts.Require().NoError(err)

		size, err := getDirSize(emptyDir)
		ts.Require().NoError(err)
		ts.Require().Equal(int64(0), size)
	})

	ts.Run("nonexistent directory", func() {
		size, err := getDirSize("/nonexistent/path")
		ts.Require().Error(err)
		ts.Require().Equal(int64(0), size)
	})

	ts.Run("nested directories", func() {
		// Create nested structure
		nestedDir := filepath.Join(ts.env.TempDir, "nested", "deep", "structure")
		err := os.MkdirAll(nestedDir, 0o750)
		ts.Require().NoError(err)

		// Add files at different levels
		file1 := filepath.Join(ts.env.TempDir, "nested", "top.txt")
		file2 := filepath.Join(ts.env.TempDir, "nested", "deep", "middle.txt")
		file3 := filepath.Join(nestedDir, "bottom.txt")

		err = os.WriteFile(file1, []byte("top"), 0o600) // 3 bytes
		ts.Require().NoError(err)
		err = os.WriteFile(file2, []byte("middle"), 0o600) // 6 bytes
		ts.Require().NoError(err)
		err = os.WriteFile(file3, []byte("bottom"), 0o600) // 6 bytes
		ts.Require().NoError(err)

		size, err := getDirSize(filepath.Join(ts.env.TempDir, "nested"))
		ts.Require().NoError(err)
		ts.Require().Equal(int64(15), size) // 3 + 6 + 6 bytes
	})
}

// TestGetCPUCount tests the getCPUCount function
func (ts *CommonTestSuite) TestGetCPUCount() {
	count := getCPUCount()
	ts.Require().Positive(count)
	ts.Require().Equal(runtime.NumCPU(), count)
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
			ts.Require().Equal(tc.expected, result,
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
			ts.Require().Equal(tc.expected, result)
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
			ts.Require().Equal(tc.expected, result)
		})
	}
}

// TestCommonTestSuite runs the test suite
func TestCommonTestSuite(t *testing.T) {
	suite.Run(t, new(CommonTestSuite))
}
