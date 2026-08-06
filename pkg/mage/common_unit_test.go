package mage

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
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

// Note: isNewer, stripPrerelease, and formatReleaseNotes were removed with the
// in-house update engine; their behavior now lives in (and is tested by)
// github.com/mrz1836/go-selfupdate.

// withVersionMockRunner runs fn with the test environment's mock runner installed.
func withVersionMockRunner(t *testing.T, testEnv *testutil.TestEnvironment, fn func() error) error {
	t.Helper()

	return testEnv.WithMockRunner(
		func(r interface{}) error {
			return SetRunner(r.(CommandRunner)) //nolint:errcheck // type assertion is safe in test
		},
		func() interface{} { return GetRunner() },
		fn,
	)
}

// versionTestProjectTag is the most recent tag of the project being built in these tests.
const versionTestProjectTag = "v0.5.0"

// mockGitVersionCommands stubs the three git commands the version resolver invokes,
// against a clean working tree. describe is the `git describe --tags --always --dirty`
// output, which is what distinguishes a checkout sitting exactly on a tag from one
// carrying commits on top of it.
func mockGitVersionCommands(testEnv *testutil.TestEnvironment, describe string) {
	testEnv.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).
		Return(versionTestProjectTag, nil).Maybe()
	testEnv.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--always", "--dirty"}).
		Return(describe, nil).Maybe()
	testEnv.Runner.On("RunCmdOutput", "git", []string{"status", "--porcelain"}).
		Return("", nil).Maybe()
}

// TestGetVersionIgnoresMageXVersionEnv guards against the toolchain pin leaking into
// the version stamped onto a target binary.
//
// MAGE_X_VERSION selects which mage-x release CI installs. Projects declare it in their
// .github/env files, which are loaded into the process environment at startup, so any
// project built locally would otherwise be stamped with the pinned tool version instead
// of its own. The build version must be resolved from the project being built:
// MAGE_X_RELEASE_VERSION -> git tag -> VERSION file -> config -> "dev".
func TestGetVersionIgnoresMageXVersionEnv(t *testing.T) {
	const pinnedToolVersion = "v1.24.1"

	t.Run("exact clean tag resolves to the project tag", func(t *testing.T) {
		testEnv := testutil.NewTestEnvironment(t)
		defer testEnv.Cleanup()

		t.Setenv("MAGE_X_VERSION", pinnedToolVersion)
		t.Setenv("MAGE_X_RELEASE_VERSION", "")
		TestResetConfig()

		mockGitVersionCommands(testEnv, versionTestProjectTag)

		require.NoError(t, withVersionMockRunner(t, testEnv, func() error {
			version := getVersion()
			require.Equal(t, versionTestProjectTag, version)
			require.NotEqual(t, pinnedToolVersion, version)
			return nil
		}))
	})

	t.Run("commits ahead of the tag resolve to dev", func(t *testing.T) {
		testEnv := testutil.NewTestEnvironment(t)
		defer testEnv.Cleanup()

		t.Setenv("MAGE_X_VERSION", pinnedToolVersion)
		t.Setenv("MAGE_X_RELEASE_VERSION", "")
		TestResetConfig()

		mockGitVersionCommands(testEnv, "v0.5.0-1-g3b88731")

		require.NoError(t, withVersionMockRunner(t, testEnv, func() error {
			version := getVersion()
			require.Equal(t, versionDev, version)
			require.NotEqual(t, pinnedToolVersion, version)
			return nil
		}))
	})

	t.Run("release version override still wins", func(t *testing.T) {
		testEnv := testutil.NewTestEnvironment(t)
		defer testEnv.Cleanup()

		t.Setenv("MAGE_X_VERSION", pinnedToolVersion)
		t.Setenv("MAGE_X_RELEASE_VERSION", "v9.9.9")
		TestResetConfig()

		mockGitVersionCommands(testEnv, versionTestProjectTag)

		require.NoError(t, withVersionMockRunner(t, testEnv, func() error {
			require.Equal(t, "v9.9.9", getVersion())
			return nil
		}))
	})

	t.Run("default ldflags do not carry the pin", func(t *testing.T) {
		testEnv := testutil.NewTestEnvironment(t)
		defer testEnv.Cleanup()

		t.Setenv("MAGE_X_VERSION", pinnedToolVersion)
		t.Setenv("MAGE_X_RELEASE_VERSION", "")
		TestResetConfig()

		mockGitVersionCommands(testEnv, "v0.5.0-1-g3b88731")
		testEnv.Runner.On("RunCmdOutput", "git", []string{"rev-parse", "--short", "HEAD"}).
			Return("3b88731", nil).Maybe()

		require.NoError(t, withVersionMockRunner(t, testEnv, func() error {
			flags := defaultLDFlags()
			require.Contains(t, flags, "-X main.version="+versionDev)
			require.NotContains(t, flags, "-X main.version="+pinnedToolVersion)
			return nil
		}))
	})

	t.Run("ldflags template expansion does not carry the pin", func(t *testing.T) {
		testEnv := testutil.NewTestEnvironment(t)
		defer testEnv.Cleanup()

		t.Setenv("MAGE_X_VERSION", pinnedToolVersion)
		t.Setenv("MAGE_X_RELEASE_VERSION", "")
		TestResetConfig()

		mockGitVersionCommands(testEnv, "v0.5.0-1-g3b88731")
		testEnv.Runner.On("RunCmdOutput", "git", []string{"rev-parse", "--short", "HEAD"}).
			Return("3b88731", nil).Maybe()

		require.NoError(t, withVersionMockRunner(t, testEnv, func() error {
			expanded := expandLDFlagsTemplates([]string{"-X main.version={{.Version}}"})
			require.Equal(t, []string{"-X main.version=" + versionDev}, expanded)
			return nil
		}))
	})
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
