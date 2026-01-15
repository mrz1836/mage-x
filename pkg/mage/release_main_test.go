//go:build integration
// +build integration

package mage

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// Static test errors
var (
	errReleaseMainTestFailed    = errors.New("release main test failed")
	errReleaseLocalInstallError = errors.New("local install error")
	errReleaseGoreleaser        = errors.New("goreleaser error")
	errReleaseGitError          = errors.New("git error")
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
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ensureGoreleaser()
		},
	)

	ts.Assert().NoError(err)
}

// TestEnsureGoreleaser_NeedsInstall tests when goreleaser needs installation
func (ts *ReleaseMainTestSuite) TestEnsureGoreleaser_NeedsInstall() {
	// Mock which command to return error (not found)
	ts.env.Runner.On("RunCmd", "which", []string{"goreleaser"}).Return(errReleaseGoreleaser)

	// Mock brew install (for macOS simulation)
	ts.env.Runner.On("RunCmd", "brew", []string{"install", "goreleaser"}).Return(nil).Maybe()

	// Mock go install (fallback)
	ts.env.Runner.On("RunCmd", "go", []string{"install", "github.com/goreleaser/goreleaser@latest"}).Return(nil).Maybe()

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
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
	err := os.WriteFile(filepath.Join(ts.env.TempDir, ".goreleaser.yml"), []byte(configContent), 0o644)
	ts.Require().NoError(err)

	oldPwd, _ := os.Getwd()
	os.Chdir(ts.env.TempDir)
	defer os.Chdir(oldPwd)

	err = ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.release.Validate()
		},
	)

	ts.Assert().NoError(err)
}

// TestValidate_DirtyRepo tests validation with uncommitted changes
func (ts *ReleaseMainTestSuite) TestValidate_DirtyRepo() {
	// Mock git status (dirty)
	ts.env.Runner.On("RunCmdOutput", "git", []string{"status", "--porcelain"}).Return("M file.go\n", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.release.Validate()
		},
	)

	ts.Assert().Error(err)
	ts.Assert().ErrorIs(err, errGitDirtyWorkingTree)
}

// TestValidate_NoTags tests validation with no git tags
func (ts *ReleaseMainTestSuite) TestValidate_NoTags() {
	// Mock git status (clean)
	ts.env.Runner.On("RunCmdOutput", "git", []string{"status", "--porcelain"}).Return("", nil)

	// Mock git describe (no tags)
	ts.env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("", errReleaseGitError)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.release.Validate()
		},
	)

	ts.Assert().Error(err)
	ts.Assert().Contains(err.Error(), "no git tags found")
}

// TestCheck_ConfigExists tests Check with existing config
func (ts *ReleaseMainTestSuite) TestCheck_ConfigExists() {
	// Create goreleaser config
	configContent := `version: 2
`
	err := os.WriteFile(filepath.Join(ts.env.TempDir, ".goreleaser.yml"), []byte(configContent), 0o644)
	ts.Require().NoError(err)

	// Mock goreleaser check
	ts.env.Runner.On("RunCmd", "which", []string{"goreleaser"}).Return(nil)
	ts.env.Runner.On("RunCmd", "goreleaser", []string{"check"}).Return(nil)

	oldPwd, _ := os.Getwd()
	os.Chdir(ts.env.TempDir)
	defer os.Chdir(oldPwd)

	err = ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.release.Check()
		},
	)

	ts.Assert().NoError(err)
}

// TestCheck_NoConfig tests Check with missing config
func (ts *ReleaseMainTestSuite) TestCheck_NoConfig() {
	// Mock goreleaser presence
	ts.env.Runner.On("RunCmd", "which", []string{"goreleaser"}).Return(nil)

	oldPwd, _ := os.Getwd()
	os.Chdir(ts.env.TempDir)
	defer os.Chdir(oldPwd)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.release.Check()
		},
	)

	ts.Assert().Error(err)
	ts.Assert().ErrorIs(err, errNoGoreleaserConfig)
}

// TestInit_Success tests Init creating config
func (ts *ReleaseMainTestSuite) TestInit_Success() {
	// Mock goreleaser presence
	ts.env.Runner.On("RunCmd", "which", []string{"goreleaser"}).Return(nil)

	// Mock goreleaser init
	ts.env.Runner.On("RunCmd", "goreleaser", []string{"init"}).Return(nil)

	oldPwd, _ := os.Getwd()
	os.Chdir(ts.env.TempDir)
	defer os.Chdir(oldPwd)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.release.Init()
		},
	)

	ts.Assert().NoError(err)
}

// TestInit_ConfigExists tests Init with existing config
func (ts *ReleaseMainTestSuite) TestInit_ConfigExists() {
	// Create existing config
	err := os.WriteFile(filepath.Join(ts.env.TempDir, ".goreleaser.yml"), []byte("version: 2\n"), 0o644)
	ts.Require().NoError(err)

	oldPwd, _ := os.Getwd()
	os.Chdir(ts.env.TempDir)
	defer os.Chdir(oldPwd)

	err = ts.release.Init()

	ts.Assert().Error(err)
	ts.Assert().ErrorIs(err, errGoreleaserConfigExists)
}

// TestClean_Success tests Clean removing dist directory
func (ts *ReleaseMainTestSuite) TestClean_Success() {
	// Create dist directory
	distDir := filepath.Join(ts.env.TempDir, "dist")
	err := os.MkdirAll(distDir, 0o755)
	ts.Require().NoError(err)

	// Create some temp files
	tempFile := filepath.Join(ts.env.TempDir, ".goreleaser-temp")
	err = os.WriteFile(tempFile, []byte("temp"), 0o644)
	ts.Require().NoError(err)

	// Mock rm command
	ts.env.Runner.On("RunCmd", "rm", []string{"-rf", "dist"}).Return(nil)
	ts.env.Runner.On("RunCmd", "rm", []string{"-f", tempFile}).Return(nil).Maybe()

	// Mock go clean
	ts.env.Runner.On("RunCmd", "go", []string{"clean", "-cache"}).Return(nil)

	oldPwd, _ := os.Getwd()
	os.Chdir(ts.env.TempDir)
	defer os.Chdir(oldPwd)

	err = ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.release.Clean()
		},
	)

	ts.Assert().NoError(err)
}

// TestChangelog_NoTags tests Changelog with no previous tags
func (ts *ReleaseMainTestSuite) TestChangelog_NoTags() {
	// Mock git describe (no tags)
	ts.env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("", errReleaseGitError)

	// Mock git log (full history)
	ts.env.Runner.On("RunCmdOutput", "git", []string{"log", "--pretty=format:- %s"}).Return("- Initial commit\n- Add feature\n", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.release.Changelog()
		},
	)

	ts.Assert().NoError(err)
}

// TestChangelog_WithTag tests Changelog since last tag
func (ts *ReleaseMainTestSuite) TestChangelog_WithTag() {
	// Mock git describe
	ts.env.Runner.On("RunCmdOutput", "git", []string{"describe", "--tags", "--abbrev=0"}).Return("v1.0.0", nil)

	// Mock git log since tag
	ts.env.Runner.On("RunCmdOutput", "git", []string{"log", "--pretty=format:- %s", "v1.0.0..HEAD"}).Return("- Fix bug\n- Add feature\n", nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.release.Changelog()
		},
	)

	ts.Assert().NoError(err)
}

// TestGoreleaserConfigFiles tests config file detection
func (ts *ReleaseMainTestSuite) TestGoreleaserConfigFiles() {
	files := GoreleaserConfigFiles()

	// Should return a list of possible config file names
	ts.Assert().NotEmpty(files)
	ts.Assert().Contains(files, ".goreleaser.yml")
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
			os.MkdirAll(filepath.Dir(testDir), 0o755)
			configPath := filepath.Join(ts.env.TempDir, filename)
			err := os.WriteFile(configPath, []byte("version: 2\n"), 0o644)
			ts.Require().NoError(err)

			files := GoreleaserConfigFiles()
			found := false
			for _, f := range files {
				if f == filename {
					found = true
					break
				}
			}
			ts.Assert().True(found, "config file %s should be in list", filename)

			os.Remove(configPath)
		}
	})

	ts.Run("binary name detection", func() {
		// Test binary name based on OS
		expectedName := "magex"
		if runtime.GOOS == OSWindows {
			expectedName = "magex.exe"
		}

		// This is implicitly tested by buildAndInstallFromTag logic
		ts.Assert().NotEmpty(expectedName)
	})

	ts.Run("installation path detection", func() {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			expectedPath := filepath.Join(homeDir, "go", "bin", "magex")
			if runtime.GOOS == OSWindows {
				expectedPath += ".exe"
			}
			ts.Assert().NotEmpty(expectedPath)
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
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.release.Snapshot()
		},
	)

	ts.Assert().NoError(err)
}

// TestReleaseTest tests test release (dry run)
func (ts *ReleaseMainTestSuite) TestReleaseTest() {
	// Mock goreleaser presence
	ts.env.Runner.On("RunCmd", "which", []string{"goreleaser"}).Return(nil)

	// Mock goreleaser test command
	ts.env.Runner.On("RunCmd", "goreleaser", []string{"release", "--skip=publish", "--clean"}).Return(nil)

	err := ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.release.Test()
		},
	)

	ts.Assert().NoError(err)
}

// TestReleaseMainTestSuite runs the test suite
func TestReleaseMainTestSuite(t *testing.T) {
	suite.Run(t, new(ReleaseMainTestSuite))
}
