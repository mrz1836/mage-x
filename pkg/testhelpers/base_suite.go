// Package testhelpers provides unified test utilities for reducing duplication
package testhelpers

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/common/env"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
)

// BaseSuite provides common test infrastructure with standardized setup/teardown patterns
type BaseSuite struct {
	suite.Suite

	// Test environment
	TestEnv *TestEnvironment

	// Common services
	FileOps    fileops.FileOperator
	EnvManager env.Environment

	// Configuration
	TmpDir       string
	OriginalDir  string
	OriginalEnv  map[string]string
	EnvVarsToSet map[string]string // Environment variables to set during test

	// Options
	Options BaseSuiteOptions
}

// BaseSuiteOptions configures the base suite behavior
type BaseSuiteOptions struct {
	// CreateTempDir controls whether to create a temporary directory
	CreateTempDir bool
	// ChangeToTempDir controls whether to change working directory to temp dir
	ChangeToTempDir bool
	// CreateGoModule controls whether to create a test Go module
	CreateGoModule bool
	// ModuleName specifies the module name (defaults to test/module)
	ModuleName string
	// PreserveEnv controls whether to preserve environment variables
	PreserveEnv bool
	// DisableCache controls whether to disable mage cache
	DisableCache bool
	// SetupGitRepo controls whether to initialize a git repository
	SetupGitRepo bool
}

// DefaultOptions returns sensible defaults for most tests
func DefaultOptions() BaseSuiteOptions {
	return BaseSuiteOptions{
		CreateTempDir:   true,
		ChangeToTempDir: true,
		CreateGoModule:  true,
		ModuleName:      "test/module",
		PreserveEnv:     true,
		DisableCache:    true,
		SetupGitRepo:    false,
	}
}

// SetupSuite runs once before all tests in the suite
func (bs *BaseSuite) SetupSuite() {
	bs.T().Helper()

	// Set default options if not configured
	if bs.Options == (BaseSuiteOptions{}) {
		bs.Options = DefaultOptions()
	}

	// Initialize services
	bs.FileOps = fileops.NewFileOperator()
	bs.EnvManager = env.NewEnvironment()
	bs.EnvVarsToSet = make(map[string]string)

	// Save original working directory
	if bs.Options.ChangeToTempDir {
		origDir, err := os.Getwd()
		bs.Require().NoError(err, "Failed to get working directory")
		bs.OriginalDir = origDir
	}

	// Save original environment if preserving
	if bs.Options.PreserveEnv {
		bs.OriginalEnv = make(map[string]string)
		for _, e := range os.Environ() {
			parts := strings.SplitN(e, "=", 2)
			if len(parts) == 2 {
				bs.OriginalEnv[parts[0]] = parts[1]
			}
		}
	}

	// Create temporary directory
	if bs.Options.CreateTempDir {
		tmpDir, err := os.MkdirTemp("", "mage-base-suite-*")
		bs.Require().NoError(err, "Failed to create temp dir")
		bs.TmpDir = tmpDir

		// Change to temp dir if requested
		if bs.Options.ChangeToTempDir {
			bs.Require().NoError(os.Chdir(tmpDir), "Failed to change to temp dir")
		}
	}

	// Set standard environment variables
	if bs.Options.DisableCache {
		bs.SetEnvVar("MAGE_CACHE_DISABLED", "true")
	}
}

// TearDownSuite runs once after all tests in the suite
func (bs *BaseSuite) TearDownSuite() {
	bs.T().Helper()

	// Restore original working directory
	if bs.OriginalDir != "" {
		if err := os.Chdir(bs.OriginalDir); err != nil {
			bs.T().Logf("Warning: failed to restore original directory: %v", err)
		}
	}

	// Restore original environment
	bs.restoreEnvironment()

	// Clean up temporary directory
	if bs.TmpDir != "" {
		if err := os.RemoveAll(bs.TmpDir); err != nil {
			bs.T().Logf("Warning: failed to remove temp dir: %v", err)
		}
	}
}

// SetupTest runs before each test method
func (bs *BaseSuite) SetupTest() {
	bs.T().Helper()

	// Create test environment if needed
	if bs.TestEnv == nil && (bs.Options.CreateGoModule || bs.Options.SetupGitRepo) {
		bs.TestEnv = NewTestEnvironment(bs.T())
	}

	// Create Go module
	if bs.Options.CreateGoModule {
		moduleName := bs.Options.ModuleName
		if moduleName == "" {
			moduleName = "test/module"
		}
		bs.TestEnv.CreateGoModule(moduleName)
	}

	// Setup git repository
	if bs.Options.SetupGitRepo {
		bs.TestEnv.SetupGitRepo()
	}

	// Apply any environment variables set by individual tests
	for key, value := range bs.EnvVarsToSet {
		bs.Require().NoError(os.Setenv(key, value), "Failed to set env var %s", key)
	}
}

// TearDownTest runs after each test method
func (bs *BaseSuite) TearDownTest() {
	bs.T().Helper()

	// Clear test environment variables
	for key := range bs.EnvVarsToSet {
		if origValue, exists := bs.OriginalEnv[key]; exists {
			bs.Require().NoError(os.Setenv(key, origValue), "Failed to restore env var %s", key)
		} else {
			bs.Require().NoError(os.Unsetenv(key), "Failed to unset env var %s", key)
		}
	}

	// Clear the set vars for next test
	bs.EnvVarsToSet = make(map[string]string)

	// TestEnv has its own cleanup via t.Cleanup, so we don't need to manually clean it
}

// Helper methods for common test operations

// SetEnvVar sets an environment variable for the current test
func (bs *BaseSuite) SetEnvVar(key, value string) {
	bs.T().Helper()
	bs.EnvVarsToSet[key] = value
}

// CreateTempFile creates a temporary file with content
func (bs *BaseSuite) CreateTempFile(filename, content string) string {
	bs.T().Helper()

	var fullPath string
	if bs.TestEnv != nil {
		bs.TestEnv.WriteFile(filename, content)
		fullPath = bs.TestEnv.AbsPath(filename)
	} else {
		fullPath = filepath.Join(bs.TmpDir, filename)
		dir := filepath.Dir(fullPath)
		bs.Require().NoError(os.MkdirAll(dir, 0o750), "Failed to create directory")
		bs.Require().NoError(os.WriteFile(fullPath, []byte(content), 0o600), "Failed to write file")
	}

	return fullPath
}

// ReadTempFile reads content from a temporary file
func (bs *BaseSuite) ReadTempFile(filename string) string {
	bs.T().Helper()

	if bs.TestEnv != nil {
		return bs.TestEnv.ReadFile(filename)
	}

	fullPath := filepath.Join(bs.TmpDir, filename)
	// Security: Validate path is within temp directory
	cleanPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanPath, bs.TmpDir) {
		bs.T().Fatalf("Security violation: path %s is outside temp directory %s", cleanPath, bs.TmpDir)
	}
	content, err := os.ReadFile(cleanPath)
	bs.Require().NoError(err, "Failed to read file %s", filename)
	return string(content)
}

// AssertFileExists asserts that a file exists
func (bs *BaseSuite) AssertFileExists(filename string) {
	bs.T().Helper()

	if bs.TestEnv != nil {
		bs.TestEnv.AssertFileExists(filename)
		return
	}

	fullPath := filepath.Join(bs.TmpDir, filename)
	_, err := os.Stat(fullPath)
	bs.NoError(err, "Expected file %s to exist", filename)
}

// AssertFileNotExists asserts that a file does not exist
func (bs *BaseSuite) AssertFileNotExists(filename string) {
	bs.T().Helper()

	if bs.TestEnv != nil {
		bs.TestEnv.AssertFileNotExists(filename)
		return
	}

	fullPath := filepath.Join(bs.TmpDir, filename)
	_, err := os.Stat(fullPath)
	bs.Error(err, "Expected file %s to not exist", filename)
}

// AssertFileContains asserts that a file contains expected content
func (bs *BaseSuite) AssertFileContains(filename, expected string) {
	bs.T().Helper()

	content := bs.ReadTempFile(filename)
	bs.Contains(content, expected, "File %s should contain %s", filename, expected)
}

// AssertFileNotContains asserts that a file does not contain content
func (bs *BaseSuite) AssertFileNotContains(filename, unexpected string) {
	bs.T().Helper()

	content := bs.ReadTempFile(filename)
	bs.NotContains(content, unexpected, "File %s should not contain %s", filename, unexpected)
}

// RequireNoError is a convenience method for require.NoError with helper marking
func (bs *BaseSuite) RequireNoError(err error, msgAndArgs ...interface{}) {
	bs.T().Helper()
	bs.Require().NoError(err, msgAndArgs...)
}

// AssertNoError is a convenience method for assert.NoError with helper marking
func (bs *BaseSuite) AssertNoError(err error, msgAndArgs ...interface{}) {
	bs.T().Helper()
	bs.NoError(err, msgAndArgs...)
}

// AssertError is a convenience method for assert.Error with helper marking
func (bs *BaseSuite) AssertError(err error, msgAndArgs ...interface{}) {
	bs.T().Helper()
	bs.Error(err, msgAndArgs...)
}

// AssertErrorContains asserts that an error contains expected text
func (bs *BaseSuite) AssertErrorContains(err error, expected string, msgAndArgs ...interface{}) {
	bs.T().Helper()
	bs.Require().Error(err, "Expected an error")
	bs.Contains(err.Error(), expected, msgAndArgs...)
}

// CreateGoModFile creates a standard go.mod file
func (bs *BaseSuite) CreateGoModFile(moduleName string) {
	bs.T().Helper()

	content := `module ` + moduleName + `

go 1.24

require github.com/magefile/mage v1.15.0
`
	bs.CreateTempFile("go.mod", content)
}

// CreateMageConfig creates a standard mage configuration file
func (bs *BaseSuite) CreateMageConfig(configContent string) {
	bs.T().Helper()

	if configContent == "" {
		// Create default config
		configContent = `project:
  name: test-project
  version: 1.0.0

build:
  output: bin/
  flags:
    - -v

test:
  coverage: true
`
	}
	bs.CreateTempFile(".mage.yaml", configContent)
}

// WithTempDir returns the temporary directory path
func (bs *BaseSuite) WithTempDir() string {
	return bs.TmpDir
}

// WithTestEnv provides access to the test environment
func (bs *BaseSuite) WithTestEnv() *TestEnvironment {
	if bs.TestEnv == nil {
		bs.TestEnv = NewTestEnvironment(bs.T())
	}
	return bs.TestEnv
}

// restoreEnvironment restores the original environment variables
func (bs *BaseSuite) restoreEnvironment() {
	bs.T().Helper()

	if bs.OriginalEnv == nil {
		return
	}

	// Clear all environment variables that were set during tests
	for key := range bs.EnvVarsToSet {
		if origValue, exists := bs.OriginalEnv[key]; exists {
			if err := os.Setenv(key, origValue); err != nil {
				bs.T().Logf("Warning: failed to restore env var %s: %v", key, err)
			}
		} else {
			if err := os.Unsetenv(key); err != nil {
				bs.T().Logf("Warning: failed to unset env var %s: %v", key, err)
			}
		}
	}
}
