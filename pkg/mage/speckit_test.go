//go:build integration
// +build integration

package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// SpeckitTestSuite defines the test suite for speckit functions
type SpeckitTestSuite struct {
	suite.Suite

	env     *testutil.TestEnvironment
	speckit Speckit
}

// SetupTest runs before each test
func (ts *SpeckitTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.speckit = Speckit{}
}

// TearDownTest runs after each test
func (ts *SpeckitTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestBackupSpeckitConstitution tests the constitution backup functionality
func (ts *SpeckitTestSuite) TestBackupSpeckitConstitution() {
	ts.setupSpeckitConfig()

	// Create test constitution file
	constitutionDir := ".specify/memory"
	err := os.MkdirAll(constitutionDir, 0o755)
	ts.Require().NoError(err)

	constitutionPath := filepath.Join(constitutionDir, "constitution.md")
	err = os.WriteFile(constitutionPath, []byte("# Test Constitution\n\nThis is a test constitution."), 0o644)
	ts.Require().NoError(err)

	config, err := GetConfig()
	ts.Require().NoError(err)

	backupPath, err := backupSpeckitConstitution(config)
	ts.Require().NoError(err)
	ts.Require().FileExists(backupPath)

	// Verify backup content
	content, err := os.ReadFile(backupPath)
	ts.Require().NoError(err)
	ts.Require().Contains(string(content), "# Test Constitution")

	// Cleanup
	_ = os.Remove(backupPath)
}

// TestBackupSpeckitConstitution_NotFound tests backup when constitution doesn't exist
func (ts *SpeckitTestSuite) TestBackupSpeckitConstitution_NotFound() {
	ts.setupSpeckitConfig()

	config, err := GetConfig()
	ts.Require().NoError(err)

	_, err = backupSpeckitConstitution(config)
	ts.Require().ErrorIs(err, errConstitutionNotFound)
}

// TestRestoreSpeckitConstitution tests the constitution restore functionality
func (ts *SpeckitTestSuite) TestRestoreSpeckitConstitution() {
	ts.setupSpeckitConfig()

	// Create backup file
	backupDir := ".specify/backups"
	err := os.MkdirAll(backupDir, 0o755)
	ts.Require().NoError(err)

	backupPath := filepath.Join(backupDir, "constitution.backup.20251212.120000.md")
	backupContent := "# Restored Constitution\n\nThis content should be restored."
	err = os.WriteFile(backupPath, []byte(backupContent), 0o644)
	ts.Require().NoError(err)

	config, err := GetConfig()
	ts.Require().NoError(err)

	err = restoreSpeckitConstitution(config, backupPath)
	ts.Require().NoError(err)

	// Verify restored content
	constitutionPath := getSpeckitConstitutionPath(config)
	content, err := os.ReadFile(constitutionPath)
	ts.Require().NoError(err)
	ts.Require().Equal(backupContent, string(content))

	// Cleanup
	_ = os.Remove(backupPath)
	_ = os.Remove(constitutionPath)
}

// TestUpdateSpeckitVersionFile tests the version file creation
func (ts *SpeckitTestSuite) TestUpdateSpeckitVersionFile() {
	ts.setupSpeckitConfig()

	config, err := GetConfig()
	ts.Require().NoError(err)

	oldVersion := "v0.0.19"
	newVersion := "v0.0.20"
	backupPath := ".specify/backups/constitution.backup.20251212.120000.md"

	err = updateSpeckitVersionFile(config, oldVersion, newVersion, backupPath)
	ts.Require().NoError(err)

	// Verify version file content
	versionFile := config.Speckit.VersionFile
	if versionFile == "" {
		versionFile = DefaultSpeckitVersionFile
	}

	content, err := os.ReadFile(versionFile)
	ts.Require().NoError(err)

	contentStr := string(content)
	ts.Require().Contains(contentStr, "version=v0.0.20")
	ts.Require().Contains(contentStr, "previous_version=v0.0.19")
	ts.Require().Contains(contentStr, "upgrade_method=automated")
	ts.Require().Contains(contentStr, "constitution_backup=")
	ts.Require().Contains(contentStr, "last_upgrade=")

	// Cleanup
	_ = os.Remove(versionFile)
}

// TestCleanOldSpeckitBackups tests the backup cleanup functionality
func (ts *SpeckitTestSuite) TestCleanOldSpeckitBackups() {
	ts.setupSpeckitConfig()

	// Create backup directory with multiple files
	backupDir := ".specify/backups"
	err := os.MkdirAll(backupDir, 0o755)
	ts.Require().NoError(err)

	// Create 10 mock backup files
	for i := 1; i <= 10; i++ {
		filename := filepath.Join(backupDir,
			fmt.Sprintf("constitution.backup.2025110%d.100000.md", i))
		err := os.WriteFile(filename, []byte("test backup content"), 0o644)
		ts.Require().NoError(err)
	}

	config, err := GetConfig()
	ts.Require().NoError(err)
	config.Speckit.BackupsToKeep = 5

	err = cleanOldSpeckitBackups(config)
	ts.Require().NoError(err)

	// Verify only 5 remain
	entries, err := os.ReadDir(backupDir)
	ts.Require().NoError(err)
	ts.Require().Len(entries, 5)

	// Cleanup
	_ = os.RemoveAll(backupDir)
}

// TestCleanOldSpeckitBackups_NoBackups tests cleanup when no backups exist
func (ts *SpeckitTestSuite) TestCleanOldSpeckitBackups_NoBackups() {
	ts.setupSpeckitConfig()

	config, err := GetConfig()
	ts.Require().NoError(err)

	// Should not error if backup directory doesn't exist
	err = cleanOldSpeckitBackups(config)
	ts.Require().NoError(err)
}

// TestGetSpeckitConstitutionPath tests the constitution path helper
func (ts *SpeckitTestSuite) TestGetSpeckitConstitutionPath() {
	// Test with default
	config := &Config{
		Speckit: SpeckitConfig{},
	}
	path := getSpeckitConstitutionPath(config)
	ts.Require().Equal(DefaultSpeckitConstitutionPath, path)

	// Test with custom path
	config.Speckit.ConstitutionPath = "custom/.specify/constitution.md"
	path = getSpeckitConstitutionPath(config)
	ts.Require().Equal("custom/.specify/constitution.md", path)
}

// TestSpeckitConstants tests that all required constants are defined
func (ts *SpeckitTestSuite) TestSpeckitConstants() {
	// Verify command constants
	ts.Require().NotEmpty(CmdUV, "CmdUV should be defined")
	ts.Require().NotEmpty(CmdUVX, "CmdUVX should be defined")
	ts.Require().NotEmpty(CmdSpecify, "CmdSpecify should be defined")

	// Verify default value constants
	ts.Require().NotEmpty(DefaultSpeckitConstitutionPath, "DefaultSpeckitConstitutionPath should be defined")
	ts.Require().NotEmpty(DefaultSpeckitVersionFile, "DefaultSpeckitVersionFile should be defined")
	ts.Require().NotEmpty(DefaultSpeckitBackupDir, "DefaultSpeckitBackupDir should be defined")
	ts.Require().NotEmpty(DefaultSpeckitCLIName, "DefaultSpeckitCLIName should be defined")
	ts.Require().NotEmpty(DefaultSpeckitGitHubRepo, "DefaultSpeckitGitHubRepo should be defined")
	ts.Require().NotEmpty(DefaultSpeckitAIProvider, "DefaultSpeckitAIProvider should be defined")
	ts.Require().Positive(DefaultSpeckitBackupsToKeep, "DefaultSpeckitBackupsToKeep should be positive")
}

// setupSpeckitConfig creates a test configuration with speckit enabled
func (ts *SpeckitTestSuite) setupSpeckitConfig() {
	TestSetConfig(&Config{
		Speckit: SpeckitConfig{
			Enabled:          true,
			ConstitutionPath: ".specify/memory/constitution.md",
			VersionFile:      ".specify/version.txt",
			BackupDir:        ".specify/backups",
			BackupsToKeep:    5,
			CLIName:          "specify-cli",
			GitHubRepo:       "git+https://github.com/github/spec-kit.git",
			AIProvider:       "claude",
		},
	})
}

// TestSpeckitTestSuite runs the test suite
func TestSpeckitTestSuite(t *testing.T) {
	suite.Run(t, new(SpeckitTestSuite))
}
