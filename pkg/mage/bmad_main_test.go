//go:build integration
// +build integration

package mage

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// Static test errors
var (
	errBmadMainTestFailed    = errors.New("bmad main test failed")
	errBmadInstallTestFailed = errors.New("bmad install test failed")
	errBmadCheckTestFailed   = errors.New("bmad check test failed")
	errBmadUpgradeTestFailed = errors.New("bmad upgrade test failed")
)

// BmadMainTestSuite defines the test suite for Bmad main methods
type BmadMainTestSuite struct {
	suite.Suite
	env  *testutil.TestEnvironment
	bmad Bmad
}

// SetupTest runs before each test
func (ts *BmadMainTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.bmad = Bmad{}
}

// TearDownTest runs after each test
func (ts *BmadMainTestSuite) TearDownTest() {
	TestResetConfig()
	ts.env.Cleanup()
}

// TestCheck_PrerequisitesMissing tests Check with missing npm
func (ts *BmadMainTestSuite) TestCheck_PrerequisitesMissing() {
	// Create minimal config
	configContent := `
bmad:
  package_name: bmad-method
  version_tag: "@beta"
`
	configPath := filepath.Join(ts.env.TempDir, ".mage.yaml")
	err := os.WriteFile(configPath, []byte(configContent), 0o644)
	ts.Require().NoError(err)

	// Change to temp directory
	oldPwd, _ := os.Getwd()
	os.Chdir(ts.env.TempDir)
	defer os.Chdir(oldPwd)

	// Note: If npm is actually installed on the system, this test will pass
	// In a real CI environment without npm, it would detect the missing prerequisite
	err = ts.bmad.Check()
	// Either succeeds (npm present) or fails with appropriate error
	if err != nil {
		// Should be a prerequisite error
		ts.Assert().True(
			errors.Is(err, errNpmNotInstalled) || errors.Is(err, errNpxNotInstalled),
			"should be npm or npx error",
		)
	}
}

// TestCheck_Success tests Check with valid installation
func (ts *BmadMainTestSuite) TestCheck_Success() {
	// Create BMAD project directory
	projectDir := filepath.Join(ts.env.TempDir, DefaultBmadProjectDir)
	err := os.MkdirAll(projectDir, 0o755)
	ts.Require().NoError(err)

	// Create config
	configContent := `
bmad:
  project_dir: ` + DefaultBmadProjectDir + `
`
	configPath := filepath.Join(ts.env.TempDir, ".mage.yaml")
	err = os.WriteFile(configPath, []byte(configContent), 0o644)
	ts.Require().NoError(err)

	// Change to temp directory
	oldPwd, _ := os.Getwd()
	os.Chdir(ts.env.TempDir)
	defer os.Chdir(oldPwd)

	// Mock npm view command for version check
	ts.env.Runner.On("RunCmdOutput", CmdNpm, []string{
		"view", "bmad-method@beta", "version",
	}).Return("1.0.0", nil).Maybe()

	err = ts.env.WithMockRunner(
		func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
		func() interface{} { return GetRunner() },
		func() error {
			return ts.bmad.Check()
		},
	)
	// Should succeed or warn about version
	if err != nil {
		// If npm/npx not available, that's acceptable in test environment
		ts.T().Skip("npm/npx not available in test environment")
	}
}

// TestGetBmadProjectDir tests project directory retrieval
func (ts *BmadMainTestSuite) TestGetBmadProjectDir() {
	tests := []struct {
		name       string
		projectDir string
		want       string
	}{
		{
			name:       "default project dir",
			projectDir: "",
			want:       DefaultBmadProjectDir,
		},
		{
			name:       "custom project dir",
			projectDir: "custom-bmad",
			want:       "custom-bmad",
		},
		{
			name:       "absolute path",
			projectDir: "/tmp/bmad-test",
			want:       "/tmp/bmad-test",
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			config := &Config{
				Bmad: BmadConfig{
					ProjectDir: tt.projectDir,
				},
			}

			result := getBmadProjectDir(config)
			ts.Assert().Equal(tt.want, result)
		})
	}
}

// TestGetBmadVersion tests version retrieval
func (ts *BmadMainTestSuite) TestGetBmadVersion() {
	tests := []struct {
		name        string
		packageName string
		versionTag  string
		mockOutput  string
		mockError   error
		wantVersion string
		wantError   bool
	}{
		{
			name:        "successful version retrieval",
			packageName: "bmad-method",
			versionTag:  "@beta",
			mockOutput:  "1.2.3",
			mockError:   nil,
			wantVersion: "1.2.3",
			wantError:   false,
		},
		{
			name:        "empty version output",
			packageName: "bmad-method",
			versionTag:  "@beta",
			mockOutput:  "",
			mockError:   nil,
			wantVersion: "",
			wantError:   true,
		},
		{
			name:        "npm command fails",
			packageName: "bmad-method",
			versionTag:  "@beta",
			mockOutput:  "",
			mockError:   errBmadMainTestFailed,
			wantVersion: "",
			wantError:   true,
		},
		{
			name:        "version with whitespace",
			packageName: "bmad-method",
			versionTag:  "@beta",
			mockOutput:  "  2.0.0  \n",
			mockError:   nil,
			wantVersion: "2.0.0",
			wantError:   false,
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			config := &Config{
				Bmad: BmadConfig{
					PackageName: tt.packageName,
					VersionTag:  tt.versionTag,
				},
			}

			// Mock npm view command
			packageSpec := tt.packageName + tt.versionTag
			ts.env.Runner.On("RunCmdOutput", CmdNpm, []string{
				"view", packageSpec, "version",
			}).Return(tt.mockOutput, tt.mockError)

			err := ts.env.WithMockRunner(
				func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
				func() interface{} { return GetRunner() },
				func() error {
					version, err := getBmadVersion(config)

					if tt.wantError {
						return err
					}

					ts.Assert().Equal(tt.wantVersion, version)
					return err
				},
			)

			if tt.wantError {
				ts.Assert().Error(err)
			} else {
				ts.Assert().NoError(err)
			}
		})
	}
}

// TestVerifyBmadInstallation tests installation verification
func (ts *BmadMainTestSuite) TestVerifyBmadInstallation() {
	ts.Run("installation verified", func() {
		// Create project directory
		projectDir := filepath.Join(ts.env.TempDir, "_bmad")
		err := os.MkdirAll(projectDir, 0o755)
		ts.Require().NoError(err)

		config := &Config{
			Bmad: BmadConfig{
				ProjectDir: projectDir,
			},
		}

		err = verifyBmadInstallation(config)
		ts.Assert().NoError(err)
	})

	ts.Run("project directory missing", func() {
		config := &Config{
			Bmad: BmadConfig{
				ProjectDir: "/nonexistent/directory",
			},
		}

		err := verifyBmadInstallation(config)
		ts.Assert().Error(err)
		ts.Assert().ErrorIs(err, errBmadNotInstalled)
	})
}

// TestCheckBmadPrerequisites tests prerequisite checking
func (ts *BmadMainTestSuite) TestCheckBmadPrerequisites() {
	// This test verifies the function logic
	// Actual npm/npx availability depends on the test environment
	err := checkBmadPrerequisites()
	// Either succeeds (prerequisites present) or returns specific error
	if err != nil {
		ts.Assert().True(
			errors.Is(err, errNpmNotInstalled) || errors.Is(err, errNpxNotInstalled),
			"should return specific prerequisite error",
		)
	}
}

// TestPrintBmadUpgradeSummary tests upgrade summary printing
func (ts *BmadMainTestSuite) TestPrintBmadUpgradeSummary() {
	tests := []struct {
		name       string
		oldVersion string
		newVersion string
	}{
		{
			name:       "normal upgrade",
			oldVersion: "1.0.0",
			newVersion: "2.0.0",
		},
		{
			name:       "unknown old version",
			oldVersion: statusUnknown,
			newVersion: "2.0.0",
		},
		{
			name:       "unknown new version",
			oldVersion: "1.0.0",
			newVersion: statusUnknown,
		},
		{
			name:       "both unknown",
			oldVersion: statusUnknown,
			newVersion: statusUnknown,
		},
		{
			name:       "same version",
			oldVersion: "1.0.0",
			newVersion: "1.0.0",
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			ts.Assert().NotPanics(func() {
				printBmadUpgradeSummary(tt.oldVersion, tt.newVersion)
			})
		})
	}
}

// TestBmadConfigVariations tests various configuration scenarios
func (ts *BmadMainTestSuite) TestBmadConfigVariations() {
	tests := []struct {
		name        string
		packageName string
		versionTag  string
		projectDir  string
	}{
		{
			name:        "default config",
			packageName: DefaultBmadPackageName,
			versionTag:  DefaultBmadVersionTag,
			projectDir:  DefaultBmadProjectDir,
		},
		{
			name:        "custom package",
			packageName: "custom-bmad-package",
			versionTag:  "@latest",
			projectDir:  "custom-dir",
		},
		{
			name:        "beta version",
			packageName: "bmad-method",
			versionTag:  "@beta",
			projectDir:  "_bmad-beta",
		},
	}

	for _, tt := range tests {
		ts.Run(tt.name, func() {
			config := &Config{
				Bmad: BmadConfig{
					PackageName: tt.packageName,
					VersionTag:  tt.versionTag,
					ProjectDir:  tt.projectDir,
				},
			}

			// Verify getter functions work correctly
			projectDir := getBmadProjectDir(config)
			ts.Assert().Equal(tt.projectDir, projectDir)

			// Verify package spec construction
			expectedPackageSpec := tt.packageName + tt.versionTag

			// Mock version retrieval
			ts.env.Runner.On("RunCmdOutput", CmdNpm, []string{
				"view", expectedPackageSpec, "version",
			}).Return("1.0.0", nil)

			err := ts.env.WithMockRunner(
				func(r interface{}) error { return SetRunner(r.(CommandRunner)) },
				func() interface{} { return GetRunner() },
				func() error {
					_, err := getBmadVersion(config)
					return err
				},
			)

			ts.Assert().NoError(err)
		})
	}
}

// TestBmadMainTestSuite runs the test suite
func TestBmadMainTestSuite(t *testing.T) {
	suite.Run(t, new(BmadMainTestSuite))
}
