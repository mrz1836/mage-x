//go:build integration
// +build integration

package mage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// BmadTestSuite defines the test suite for bmad functions
type BmadTestSuite struct {
	suite.Suite

	env  *testutil.TestEnvironment
	bmad Bmad
}

// SetupTest runs before each test
func (ts *BmadTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.bmad = Bmad{}
}

// TearDownTest runs after each test
func (ts *BmadTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestGetBmadProjectDir tests the project directory helper
func (ts *BmadTestSuite) TestGetBmadProjectDir() {
	// Test with default
	config := &Config{
		Bmad: BmadConfig{},
	}
	path := getBmadProjectDir(config)
	ts.Require().Equal(DefaultBmadProjectDir, path)

	// Test with custom path
	config.Bmad.ProjectDir = "custom/_bmad"
	path = getBmadProjectDir(config)
	ts.Require().Equal("custom/_bmad", path)
}

// TestVerifyBmadInstallation tests the installation verification
func (ts *BmadTestSuite) TestVerifyBmadInstallation() {
	ts.setupBmadConfig()

	config, err := GetConfig()
	ts.Require().NoError(err)

	// Should fail when project dir doesn't exist
	err = verifyBmadInstallation(config)
	ts.Require().ErrorIs(err, errBmadNotInstalled)
}

// TestVerifyBmadInstallation_Success tests successful installation verification
func (ts *BmadTestSuite) TestVerifyBmadInstallation_Success() {
	ts.setupBmadConfig()

	// Create the project directory
	projectDir := DefaultBmadProjectDir
	err := os.MkdirAll(projectDir, 0o755)
	ts.Require().NoError(err)
	defer os.RemoveAll(projectDir)

	config, err := GetConfig()
	ts.Require().NoError(err)

	// Should succeed when project dir exists
	err = verifyBmadInstallation(config)
	ts.Require().NoError(err)
}

// TestBmadConstants tests that all required constants are defined
func (ts *BmadTestSuite) TestBmadConstants() {
	// Verify command constants
	ts.Require().NotEmpty(CmdNpm, "CmdNpm should be defined")
	ts.Require().NotEmpty(CmdNpx, "CmdNpx should be defined")
	ts.Require().NotEmpty(CmdBmad, "CmdBmad should be defined")

	// Verify default value constants
	ts.Require().NotEmpty(DefaultBmadProjectDir, "DefaultBmadProjectDir should be defined")
	ts.Require().NotEmpty(DefaultBmadVersionTag, "DefaultBmadVersionTag should be defined")
	ts.Require().NotEmpty(DefaultBmadPackageName, "DefaultBmadPackageName should be defined")
}

// TestBmadConfigDefaults tests that BMAD config defaults are properly set
func (ts *BmadTestSuite) TestBmadConfigDefaults() {
	ts.setupBmadConfig()

	config, err := GetConfig()
	ts.Require().NoError(err)

	// Verify default values are set correctly
	ts.Require().Equal("_bmad", config.Bmad.ProjectDir)
	ts.Require().Equal("@alpha", config.Bmad.VersionTag)
	ts.Require().Equal("bmad-method", config.Bmad.PackageName)
}

// TestBmadCustomConfig tests that custom BMAD config is properly loaded
func (ts *BmadTestSuite) TestBmadCustomConfig() {
	// Set custom config
	TestSetConfig(&Config{
		Bmad: BmadConfig{
			Enabled:     true,
			ProjectDir:  "custom_bmad",
			VersionTag:  "@latest",
			PackageName: "custom-bmad",
		},
	})

	config, err := GetConfig()
	ts.Require().NoError(err)

	ts.Require().True(config.Bmad.Enabled)
	ts.Require().Equal("custom_bmad", config.Bmad.ProjectDir)
	ts.Require().Equal("@latest", config.Bmad.VersionTag)
	ts.Require().Equal("custom-bmad", config.Bmad.PackageName)
}

// TestBmadProjectDirCreation tests project directory creation and detection
func (ts *BmadTestSuite) TestBmadProjectDirCreation() {
	ts.setupBmadConfig()

	projectDir := "_bmad"

	// Directory should not exist initially
	_, err := os.Stat(projectDir)
	ts.Require().True(os.IsNotExist(err))

	// Create the directory
	err = os.MkdirAll(projectDir, 0o755)
	ts.Require().NoError(err)
	defer os.RemoveAll(projectDir)

	// Create a sample file to simulate BMAD installation
	sampleFile := filepath.Join(projectDir, "agents")
	err = os.MkdirAll(sampleFile, 0o755)
	ts.Require().NoError(err)

	// Verify directory exists
	info, err := os.Stat(projectDir)
	ts.Require().NoError(err)
	ts.Require().True(info.IsDir())
}

// setupBmadConfig creates a test configuration with bmad enabled
func (ts *BmadTestSuite) setupBmadConfig() {
	TestSetConfig(&Config{
		Bmad: BmadConfig{
			Enabled:     true,
			ProjectDir:  "_bmad",
			VersionTag:  "@alpha",
			PackageName: "bmad-method",
		},
	})
}

// TestBmadTestSuite runs the test suite
func TestBmadTestSuite(t *testing.T) {
	suite.Run(t, new(BmadTestSuite))
}
