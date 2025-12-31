package testhelpers

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

// errTest is a static test error for linter compliance
var errTest = testError{"test error"}

// errTestWithMsg is a static test error with message for linter compliance
var errTestWithMsg = testError{"test error with message"}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e testError) Error() string {
	return e.msg
}

// BaseSuiteTestSuite tests the BaseSuite functionality
type BaseSuiteTestSuite struct {
	BaseSuite
}

func (s *BaseSuiteTestSuite) SetupSuite() {
	s.Options = DefaultOptions()
	s.BaseSuite.SetupSuite()
}

func (s *BaseSuiteTestSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *BaseSuiteTestSuite) SetupTest() {
	s.BaseSuite.SetupTest()
}

func (s *BaseSuiteTestSuite) TearDownTest() {
	s.BaseSuite.TearDownTest()
}

func (s *BaseSuiteTestSuite) TestDefaultOptions() {
	opts := DefaultOptions()
	s.True(opts.CreateTempDir)
	s.True(opts.ChangeToTempDir)
	s.True(opts.CreateGoModule)
	s.Equal("test/module", opts.ModuleName)
	s.True(opts.PreserveEnv)
	s.True(opts.DisableCache)
	s.False(opts.SetupGitRepo)
}

func (s *BaseSuiteTestSuite) TestSetEnvVar() {
	s.SetEnvVar("TEST_VAR", "test_value")
	s.Equal("test_value", s.EnvVarsToSet["TEST_VAR"])
}

func (s *BaseSuiteTestSuite) TestCreateTempFile() {
	fullPath := s.CreateTempFile("test_file.txt", "test content")
	s.FileExists(fullPath)

	content, err := os.ReadFile(fullPath) //nolint:gosec // test file
	s.Require().NoError(err)
	s.Equal("test content", string(content))
}

func (s *BaseSuiteTestSuite) TestCreateTempFileNested() {
	fullPath := s.CreateTempFile("nested/dir/test_file.txt", "nested content")
	s.FileExists(fullPath)

	content, err := os.ReadFile(fullPath) //nolint:gosec // test file
	s.Require().NoError(err)
	s.Equal("nested content", string(content))
}

func (s *BaseSuiteTestSuite) TestReadTempFile() {
	s.CreateTempFile("read_test.txt", "read content")
	content := s.ReadTempFile("read_test.txt")
	s.Equal("read content", content)
}

func (s *BaseSuiteTestSuite) TestAssertFileExists() {
	s.CreateTempFile("exists_test.txt", "I exist")
	s.AssertFileExists("exists_test.txt")
}

func (s *BaseSuiteTestSuite) TestAssertFileNotExists() {
	s.AssertFileNotExists("nonexistent_file.txt")
}

func (s *BaseSuiteTestSuite) TestAssertFileContains() {
	s.CreateTempFile("contains_test.txt", "hello world foo bar")
	s.AssertFileContains("contains_test.txt", "foo bar")
}

func (s *BaseSuiteTestSuite) TestAssertFileNotContains() {
	s.CreateTempFile("not_contains_test.txt", "hello world")
	s.AssertFileNotContains("not_contains_test.txt", "foo bar")
}

func (s *BaseSuiteTestSuite) TestRequireNoError() {
	s.RequireNoError(nil)
}

func (s *BaseSuiteTestSuite) TestAssertNoError() {
	s.AssertNoError(nil)
}

func (s *BaseSuiteTestSuite) TestAssertError() {
	s.AssertError(errTest)
}

func (s *BaseSuiteTestSuite) TestAssertErrorContains() {
	s.AssertErrorContains(errTestWithMsg, "with message")
}

func (s *BaseSuiteTestSuite) TestCreateGoModFile() {
	s.CreateGoModFile("test/mymodule")
	s.AssertFileExists("go.mod")
	s.AssertFileContains("go.mod", "module test/mymodule")
	s.AssertFileContains("go.mod", "go 1.24")
}

func (s *BaseSuiteTestSuite) TestCreateMageConfig() {
	s.CreateMageConfig("")
	s.AssertFileExists(".mage.yaml")
	s.AssertFileContains(".mage.yaml", "test-project")
	s.AssertFileContains(".mage.yaml", "coverage: true")
}

func (s *BaseSuiteTestSuite) TestCreateMageConfigCustom() {
	customConfig := `project:
  name: custom-project
build:
  output: dist/
`
	s.CreateMageConfig(customConfig)
	s.AssertFileExists(".mage.yaml")
	s.AssertFileContains(".mage.yaml", "custom-project")
	s.AssertFileContains(".mage.yaml", "dist/")
}

func (s *BaseSuiteTestSuite) TestWithTempDir() {
	tmpDir := s.WithTempDir()
	s.NotEmpty(tmpDir)
	s.DirExists(tmpDir)
}

func (s *BaseSuiteTestSuite) TestWithTestEnv() {
	env := s.WithTestEnv()
	s.NotNil(env)
	s.Equal(env, s.TestEnv)
}

func TestBaseSuiteTestSuite(t *testing.T) {
	suite.Run(t, new(BaseSuiteTestSuite))
}

// MinimalBaseSuiteTestSuite tests BaseSuite with minimal options (no temp dir change)
type MinimalBaseSuiteTestSuite struct {
	BaseSuite
}

func (s *MinimalBaseSuiteTestSuite) SetupSuite() {
	s.Options = BaseSuiteOptions{
		CreateTempDir:   true,
		ChangeToTempDir: false, // Don't change directory
		CreateGoModule:  false,
		PreserveEnv:     true,
		DisableCache:    false,
	}
	s.BaseSuite.SetupSuite()
}

func (s *MinimalBaseSuiteTestSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *MinimalBaseSuiteTestSuite) SetupTest() {
	s.BaseSuite.SetupTest()
}

func (s *MinimalBaseSuiteTestSuite) TearDownTest() {
	s.BaseSuite.TearDownTest()
}

func (s *MinimalBaseSuiteTestSuite) TestCreateTempFileWithoutTestEnv() {
	// Without TestEnv, it should use TmpDir directly
	fullPath := s.CreateTempFile("minimal_test.txt", "minimal content")
	s.FileExists(fullPath)

	expectedPath := filepath.Join(s.TmpDir, "minimal_test.txt")
	s.Equal(expectedPath, fullPath)
}

func (s *MinimalBaseSuiteTestSuite) TestReadTempFileWithoutTestEnv() {
	s.CreateTempFile("read_minimal.txt", "read minimal content")
	content := s.ReadTempFile("read_minimal.txt")
	s.Equal("read minimal content", content)
}

func (s *MinimalBaseSuiteTestSuite) TestAssertFileExistsWithoutTestEnv() {
	s.CreateTempFile("exists_minimal.txt", "I exist")
	s.AssertFileExists("exists_minimal.txt")
}

func (s *MinimalBaseSuiteTestSuite) TestAssertFileNotExistsWithoutTestEnv() {
	s.AssertFileNotExists("nonexistent_minimal.txt")
}

func TestMinimalBaseSuiteTestSuite(t *testing.T) {
	suite.Run(t, new(MinimalBaseSuiteTestSuite))
}

// GitRepoBaseSuiteTestSuite tests BaseSuite with git repo setup
type GitRepoBaseSuiteTestSuite struct {
	BaseSuite
}

func (s *GitRepoBaseSuiteTestSuite) SetupSuite() {
	s.Options = BaseSuiteOptions{
		CreateTempDir:   true,
		ChangeToTempDir: true,
		CreateGoModule:  false,
		PreserveEnv:     true,
		DisableCache:    false,
		SetupGitRepo:    true,
	}
	s.BaseSuite.SetupSuite()
}

func (s *GitRepoBaseSuiteTestSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *GitRepoBaseSuiteTestSuite) SetupTest() {
	s.BaseSuite.SetupTest()
}

func (s *GitRepoBaseSuiteTestSuite) TearDownTest() {
	s.BaseSuite.TearDownTest()
}

func (s *GitRepoBaseSuiteTestSuite) TestGitRepoSetup() {
	// Git repo should be set up
	s.DirExists(filepath.Join(s.TestEnv.WorkDir(), ".git"))
}

func TestGitRepoBaseSuiteTestSuite(t *testing.T) {
	suite.Run(t, new(GitRepoBaseSuiteTestSuite))
}

// EmptyOptionsBaseSuiteTestSuite tests BaseSuite with zero-value options (uses defaults)
type EmptyOptionsBaseSuiteTestSuite struct {
	BaseSuite
}

func (s *EmptyOptionsBaseSuiteTestSuite) SetupSuite() {
	// Don't set Options - let SetupSuite use defaults
	s.BaseSuite.SetupSuite()
}

func (s *EmptyOptionsBaseSuiteTestSuite) TearDownSuite() {
	s.BaseSuite.TearDownSuite()
}

func (s *EmptyOptionsBaseSuiteTestSuite) SetupTest() {
	s.BaseSuite.SetupTest()
}

func (s *EmptyOptionsBaseSuiteTestSuite) TearDownTest() {
	s.BaseSuite.TearDownTest()
}

func (s *EmptyOptionsBaseSuiteTestSuite) TestDefaultOptionsApplied() {
	// Default options should be applied
	s.True(s.Options.CreateTempDir)
	s.True(s.Options.ChangeToTempDir)
	s.True(s.Options.CreateGoModule)
	s.NotEmpty(s.TmpDir)
}

func TestEmptyOptionsBaseSuiteTestSuite(t *testing.T) {
	suite.Run(t, new(EmptyOptionsBaseSuiteTestSuite))
}
