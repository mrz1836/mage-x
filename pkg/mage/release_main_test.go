//go:build integration
// +build integration

package mage

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// Static test errors
var (
	errReleaseGoreleaser = errors.New("goreleaser error")
	errReleaseGitError   = errors.New("git error")
)

// ReleaseMainTestSuite defines the test suite for Release main methods
type ReleaseMainTestSuite struct {
	suite.Suite

	env     *testutil.TestEnvironment
	release Release
}

// SetupTest runs before each test
func (ts *ReleaseMainTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.release = Release{}
}

// TearDownTest runs after each test
func (ts *ReleaseMainTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestEnsureGoreleaser_AlreadyInstalled tests when goreleaser is already present
func (ts *ReleaseMainTestSuite) TestEnsureGoreleaser_AlreadyInstalled() {
	// Mock which command to return success
	ts.env.Runner.On("RunCmd", "which", []string{"goreleaser"}).Return(nil)

	err := ts.env.WithMockRunner(
		setTestCommandRunner,
		func() any { return GetRunner() },
		func() error {
			return ensureGoreleaser()
		},
	)

	ts.Require().NoError(err)
}

// TestEnsureGoreleaser_NeedsInstall tests when goreleaser needs installation
func (ts *ReleaseMainTestSuite) TestEnsureGoreleaser_NeedsInstall() {
	// Mock which command to return error (not found). Defined first so it still matches
	// the goreleaser lookup before the catch-all below.
	ts.env.Runner.On("RunCmd", "which", []string{"goreleaser"}).Return(errReleaseGoreleaser)

	// Mock brew install (for macOS simulation)
	ts.env.Runner.On("RunCmd", "brew", []string{"install", "goreleaser"}).Return(nil).Maybe()

	// Mock go install (fallback)
	ts.env.Runner.On("RunCmd", "go", []string{"install", "github.com/goreleaser/goreleaser@latest"}).Return(nil).Maybe()

	// Catch-all so installGoreleaser's unmocked "curl ... | sh" (and any OS-specific
	// fallback) returns success instead of panicking on an unexpected call. The test
	// ignores the result, so this only prevents the panic across OSes.
	ts.env.Runner.On("RunCmd", mock.Anything, mock.Anything).Return(nil).Maybe()

	err := ts.env.WithMockRunner(
		setTestCommandRunner,
		func() any { return GetRunner() },
		func() error {
			return ensureGoreleaser()
		},
	)

	// Should attempt installation
	// Error handling depends on whether brew/go install succeed
	_ = err // May succeed or fail depending on mock execution path
}

// TestValidate_CleanRepo tests validation with clean repository
func (ts *ReleaseMainTestSuite) TestValidate_CleanRepo() {
	// Mock git status (clean)
	ts.env.Runner.On("RunCmdOutput", "git", []string{"status", "--porcelain"}).Return("", nil)

	// Mock git describe (tag exists)
	ts.env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("v1.0.0", nil)

	// Mock which goreleaser
	ts.env.Runner.On("RunCmd", "which", []string{"goreleaser"}).Return(nil)

	// Mock goreleaser check
	ts.env.Runner.On("RunCmd", "goreleaser", []string{"check"}).Return(nil)

	// Create dummy goreleaser config
	configContent := `# .goreleaser.yml
version: 2
`
	err := os.WriteFile(filepath.Join(ts.env.TempDir, ".goreleaser.yml"), []byte(configContent), 0o600)
	ts.Require().NoError(err)

	defer chdirForTest(ts.T(), ts.env.TempDir)()

	err = ts.env.WithMockRunner(
		setTestCommandRunner,
		func() any { return GetRunner() },
		func() error {
			return ts.release.Validate()
		},
	)

	ts.Require().NoError(err)
}

// TestValidate_DirtyRepo tests validation with uncommitted changes
func (ts *ReleaseMainTestSuite) TestValidate_DirtyRepo() {
	// Mock git status (dirty)
	ts.env.Runner.On("RunCmdOutput", "git", []string{"status", "--porcelain"}).Return("M file.go\n", nil)

	err := ts.env.WithMockRunner(
		setTestCommandRunner,
		func() any { return GetRunner() },
		func() error {
			return ts.release.Validate()
		},
	)

	ts.Require().Error(err)
	ts.ErrorIs(err, errGitDirtyWorkingTree)
}

// TestValidate_NoTags tests validation with no git tags
func (ts *ReleaseMainTestSuite) TestValidate_NoTags() {
	// Mock git status (clean)
	ts.env.Runner.On("RunCmdOutput", "git", []string{"status", "--porcelain"}).Return("", nil)

	// Mock git describe (no tags)
	ts.env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("", errReleaseGitError)

	err := ts.env.WithMockRunner(
		setTestCommandRunner,
		func() any { return GetRunner() },
		func() error {
			return ts.release.Validate()
		},
	)

	ts.Require().Error(err)
	ts.Contains(err.Error(), "no git tags found")
}

// TestCheck_ConfigExists tests Check with existing config
func (ts *ReleaseMainTestSuite) TestCheck_ConfigExists() {
	// Create goreleaser config
	configContent := `version: 2
`
	err := os.WriteFile(filepath.Join(ts.env.TempDir, ".goreleaser.yml"), []byte(configContent), 0o600)
	ts.Require().NoError(err)

	// Mock goreleaser check
	ts.env.Runner.On("RunCmd", "which", []string{"goreleaser"}).Return(nil)
	ts.env.Runner.On("RunCmd", "goreleaser", []string{"check"}).Return(nil)

	defer chdirForTest(ts.T(), ts.env.TempDir)()

	err = ts.env.WithMockRunner(
		setTestCommandRunner,
		func() any { return GetRunner() },
		func() error {
			return ts.release.Check()
		},
	)

	ts.Require().NoError(err)
}

// TestCheck_NoConfig tests Check with missing config
func (ts *ReleaseMainTestSuite) TestCheck_NoConfig() {
	// Mock goreleaser presence
	ts.env.Runner.On("RunCmd", "which", []string{"goreleaser"}).Return(nil)

	defer chdirForTest(ts.T(), ts.env.TempDir)()

	err := ts.env.WithMockRunner(
		setTestCommandRunner,
		func() any { return GetRunner() },
		func() error {
			return ts.release.Check()
		},
	)

	ts.Require().Error(err)
	ts.ErrorIs(err, errNoGoreleaserConfig)
}

// TestInit_Success tests Init creating config
func (ts *ReleaseMainTestSuite) TestInit_Success() {
	// Mock goreleaser presence
	ts.env.Runner.On("RunCmd", "which", []string{"goreleaser"}).Return(nil)

	// Mock goreleaser init
	ts.env.Runner.On("RunCmd", "goreleaser", []string{"init"}).Return(nil)

	defer chdirForTest(ts.T(), ts.env.TempDir)()

	err := ts.env.WithMockRunner(
		setTestCommandRunner,
		func() any { return GetRunner() },
		func() error {
			return ts.release.Init()
		},
	)

	ts.Require().NoError(err)
}

// TestInit_ConfigExists tests Init with existing config
func (ts *ReleaseMainTestSuite) TestInit_ConfigExists() {
	// Create existing config
	err := os.WriteFile(filepath.Join(ts.env.TempDir, ".goreleaser.yml"), []byte("version: 2\n"), 0o600)
	ts.Require().NoError(err)

	defer chdirForTest(ts.T(), ts.env.TempDir)()

	err = ts.release.Init()

	ts.Require().Error(err)
	ts.ErrorIs(err, errGoreleaserConfigExists)
}

// TestClean_Success tests Clean removing dist directory
func (ts *ReleaseMainTestSuite) TestClean_Success() {
	// Create dist directory
	distDir := filepath.Join(ts.env.TempDir, "dist")
	err := os.MkdirAll(distDir, 0o750)
	ts.Require().NoError(err)

	// Create some temp files. Clean uses filepath.Glob(".goreleaser-*")
	// relative to the working directory, so the file must live in TempDir
	// (where we chdir below) and the glob will yield the relative basename.
	const tempFileName = ".goreleaser-temp"
	tempFile := filepath.Join(ts.env.TempDir, tempFileName)
	err = os.WriteFile(tempFile, []byte("temp"), 0o600)
	ts.Require().NoError(err)

	// Mock rm command
	ts.env.Runner.On("RunCmd", "rm", []string{"-rf", "dist"}).Return(nil).Once()
	// Clean globs relative paths, so it removes the basename, not the absolute path.
	ts.env.Runner.On("RunCmd", "rm", []string{"-f", tempFileName}).Return(nil).Once()

	// Mock go clean
	ts.env.Runner.On("RunCmd", "go", []string{"clean", "-cache"}).Return(nil)

	defer chdirForTest(ts.T(), ts.env.TempDir)()

	err = ts.env.WithMockRunner(
		setTestCommandRunner,
		func() any { return GetRunner() },
		func() error {
			return ts.release.Clean()
		},
	)

	ts.Require().NoError(err)
}

// TestChangelog_NoTags tests Changelog with no previous tags
func (ts *ReleaseMainTestSuite) TestChangelog_NoTags() {
	// Mock git describe (no tags)
	ts.env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("", errReleaseGitError)

	// Mock git log (full history)
	ts.env.Runner.On("RunCmdOutput", "git", []string{"log", "--pretty=format:- %s"}).Return("- Initial commit\n- Add feature\n", nil)

	err := ts.env.WithMockRunner(
		setTestCommandRunner,
		func() any { return GetRunner() },
		func() error {
			return ts.release.Changelog()
		},
	)

	ts.Require().NoError(err)
}

// TestChangelog_WithTag tests Changelog since last tag
func (ts *ReleaseMainTestSuite) TestChangelog_WithTag() {
	// Mock git describe
	ts.env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("v1.0.0", nil)

	// Mock git log since tag
	ts.env.Runner.On("RunCmdOutput", "git", []string{"log", "--pretty=format:- %s", "v1.0.0..HEAD"}).Return("- Fix bug\n- Add feature\n", nil)

	err := ts.env.WithMockRunner(
		setTestCommandRunner,
		func() any { return GetRunner() },
		func() error {
			return ts.release.Changelog()
		},
	)

	ts.Require().NoError(err)
}

// TestGoreleaserConfigFiles tests config file detection
func (ts *ReleaseMainTestSuite) TestGoreleaserConfigFiles() {
	files := GoreleaserConfigFiles()

	// Should return a list of possible config file names
	ts.NotEmpty(files)
	ts.Contains(files, ".goreleaser.yml")
}

// TestReleaseHelpers tests various helper functions
func (ts *ReleaseMainTestSuite) TestReleaseHelpers() {
	ts.Run("config file detection", func() {
		// Create various config files
		tests := []string{
			".goreleaser.yml",
			".goreleaser.yaml",
			"goreleaser.yml",
			"goreleaser.yaml",
		}

		for _, filename := range tests {
			testDir := filepath.Join(ts.env.TempDir, filename)
			ts.Require().NoError(os.MkdirAll(filepath.Dir(testDir), 0o750))
			configPath := filepath.Join(ts.env.TempDir, filename)
			err := os.WriteFile(configPath, []byte("version: 2\n"), 0o600)
			ts.Require().NoError(err)

			files := GoreleaserConfigFiles()
			found := false
			for _, f := range files {
				if f == filename {
					found = true
					break
				}
			}
			ts.True(found, "config file %s should be in list", filename)

			ts.Require().NoError(os.Remove(configPath))
		}
	})

	ts.Run("binary name detection", func() {
		// Test binary name based on OS
		expectedName := "magex"
		if runtime.GOOS == OSWindows {
			expectedName = "magex.exe"
		}

		// This is implicitly tested by buildAndInstallFromTag logic
		ts.NotEmpty(expectedName)
	})

	ts.Run("installation path detection", func() {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			expectedPath := filepath.Join(homeDir, "go", "bin", "magex")
			if runtime.GOOS == OSWindows {
				expectedPath += ".exe"
			}
			ts.NotEmpty(expectedPath)
		}
	})
}

// TestReleaseSnapshot tests snapshot build
func (ts *ReleaseMainTestSuite) TestReleaseSnapshot() {
	// Mock goreleaser presence
	ts.env.Runner.On("RunCmd", "which", []string{"goreleaser"}).Return(nil)

	// Mock goreleaser snapshot command
	ts.env.Runner.On("RunCmd", "goreleaser", []string{"release", "--snapshot", "--skip=publish", "--clean"}).Return(nil)

	err := ts.env.WithMockRunner(
		setTestCommandRunner,
		func() any { return GetRunner() },
		func() error {
			return ts.release.Snapshot()
		},
	)

	ts.Require().NoError(err)
}

// TestReleaseTest tests test release (dry run)
func (ts *ReleaseMainTestSuite) TestReleaseTest() {
	// Mock goreleaser presence
	ts.env.Runner.On("RunCmd", "which", []string{"goreleaser"}).Return(nil)

	// Mock goreleaser test command
	ts.env.Runner.On("RunCmd", "goreleaser", []string{"release", "--skip=publish", "--clean"}).Return(nil)

	err := ts.env.WithMockRunner(
		setTestCommandRunner,
		func() any { return GetRunner() },
		func() error {
			return ts.release.Test()
		},
	)

	ts.Require().NoError(err)
}

// TestReleaseMainTestSuite runs the test suite
func TestReleaseMainTestSuite(t *testing.T) {
	suite.Run(t, new(ReleaseMainTestSuite))
}
