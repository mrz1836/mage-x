package paths

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/testhelpers"
)

// PathsRefactoredTestSuite demonstrates using BaseSuite for path testing
type PathsRefactoredTestSuite struct {
	testhelpers.BaseSuite
}

// SetupSuite configures the test suite options
func (ts *PathsRefactoredTestSuite) SetupSuite() {
	// Configure options for path testing
	ts.Options = testhelpers.BaseSuiteOptions{
		CreateTempDir:   true,
		ChangeToTempDir: true,
		CreateGoModule:  false, // Path tests don't need Go modules
		PreserveEnv:     false, // Path tests don't need env management
		DisableCache:    false, // Path tests don't use mage cache
		SetupGitRepo:    false, // Path tests don't need git
	}

	// Call parent setup
	ts.BaseSuite.SetupSuite()
}

// TestPathBuilder demonstrates the refactored approach to testing
func (ts *PathsRefactoredTestSuite) TestPathBuilder() {
	ts.Run("build simple path", func() {
		builder := Build("test", "path")
		result := builder.String()
		ts.Require().Contains(result, "test")
		ts.Require().Contains(result, "path")
	})

	ts.Run("build from current dir", func() {
		builder := Current()
		result := builder.String()
		ts.Require().NotEmpty(result)
	})

	ts.Run("build file path", func() {
		builder := File("test.txt")
		result := builder.String()
		ts.Require().Equal("test.txt", result)
	})
}

// TestFileOperations demonstrates file operations with the refactored suite
func (ts *PathsRefactoredTestSuite) TestFileOperations() {
	// Create test files using the consolidated helper
	ts.CreateTempFile("test1.txt", "content1")
	ts.CreateTempFile("subdir/test2.txt", "content2")

	// Assert files exist using the consolidated helper
	ts.AssertFileExists("test1.txt")
	ts.AssertFileExists("subdir/test2.txt")

	// Verify content using the consolidated helper
	ts.AssertFileContains("test1.txt", "content1")
	ts.AssertFileContains("subdir/test2.txt", "content2")

	// Test file doesn't exist
	ts.AssertFileNotExists("nonexistent.txt")
}

// TestPathValidation demonstrates validation testing with the suite
func (ts *PathsRefactoredTestSuite) TestPathValidation() {
	validator := NewPathValidator()

	ts.Run("valid path", func() {
		// This tests that the validator exists and can be created
		ts.Require().NotNil(validator)
	})

	ts.Run("path set operations", func() {
		pathSet := NewPathSet()
		pathSet.Add("test/path1")
		pathSet.Add("test/path2")

		ts.Require().True(pathSet.Contains("test/path1"))
		ts.Require().True(pathSet.Contains("test/path2"))
		ts.Require().False(pathSet.Contains("nonexistent"))
	})
}

// TestPathsRefactoredTestSuite runs the refactored test suite
func TestPathsRefactoredTestSuite(t *testing.T) {
	suite.Run(t, new(PathsRefactoredTestSuite))
}
