package mage

import (
	"os"
	"testing"

	"github.com/mrz1836/mage-x/pkg/testhelpers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ReleaseSimpleTestSuite provides simplified release testing
type ReleaseSimpleTestSuite struct {
	suite.Suite

	release     Release
	mockRunner  *testhelpers.MockRunner
	originalEnv map[string]string
}

// SetupSuite prepares the test environment
func (suite *ReleaseSimpleTestSuite) SetupSuite() {
	// Save original environment
	suite.originalEnv = make(map[string]string)
	envVars := []string{"GITHUB_TOKEN", "github_token", "USER"}
	for _, env := range envVars {
		if val := os.Getenv(env); val != "" {
			suite.originalEnv[env] = val
		}
	}
}

// TearDownSuite cleans up the test environment
func (suite *ReleaseSimpleTestSuite) TearDownSuite() {
	// Restore original environment
	for env, val := range suite.originalEnv {
		if err := os.Setenv(env, val); err != nil {
			suite.T().Logf("Warning: failed to restore environment variable %s: %v", env, err)
		}
	}
}

// SetupTest prepares each test case
func (suite *ReleaseSimpleTestSuite) SetupTest() {
	suite.release = Release{}
	suite.mockRunner = testhelpers.NewMockRunner()

	// Clean environment for each test
	if err := os.Unsetenv("GITHUB_TOKEN"); err != nil {
		suite.T().Logf("Warning: failed to unset GITHUB_TOKEN: %v", err)
	}
	if err := os.Unsetenv("github_token"); err != nil {
		suite.T().Logf("Warning: failed to unset github_token: %v", err)
	}
	if err := os.Unsetenv("USER"); err != nil {
		suite.T().Logf("Warning: failed to unset USER: %v", err)
	}
}

// TearDownTest cleans up after each test
func (suite *ReleaseSimpleTestSuite) TearDownTest() {
	// Clean environment variables
	envVars := []string{"GITHUB_TOKEN", "github_token", "USER"}
	for _, env := range envVars {
		if err := os.Unsetenv(env); err != nil {
			suite.T().Logf("Warning: failed to unset environment variable %s: %v", env, err)
		}
	}
}

// TestReleaseDefaultSuccess tests successful default release
func (suite *ReleaseSimpleTestSuite) TestReleaseDefaultSuccess() {
	if err := os.Setenv("GITHUB_TOKEN", "test-token"); err != nil {
		suite.T().Fatalf("Failed to set GITHUB_TOKEN: %v", err)
	}

	// Mock goreleaser commands
	suite.mockRunner.SetOutput("which", []string{"goreleaser"}, "/usr/local/bin/goreleaser")
	suite.mockRunner.SetError("goreleaser", []string{"release", "--clean"}, nil)

	originalRunner := GetRunner()
	if err := SetRunner(suite.mockRunner); err != nil {
		suite.T().Fatalf("Failed to set mock runner: %v", err)
	}
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			suite.T().Logf("Warning: failed to restore original runner: %v", err)
		}
	}()

	err := suite.release.Default()
	suite.Require().NoError(err)

	// Verify commands were called
	suite.mockRunner.AssertCalledWith(suite.T(), "which", "goreleaser")
}

// TestReleaseDefaultFailsWithoutToken tests failure without GitHub token
func (suite *ReleaseSimpleTestSuite) TestReleaseDefaultFailsWithoutToken() {
	// No token set
	err := suite.release.Default()
	suite.Require().Error(err)
	suite.Contains(err.Error(), "GITHUB_TOKEN environment variable is required")
}

// TestReleaseTestSuccess tests successful test release
func (suite *ReleaseSimpleTestSuite) TestReleaseTestSuccess() {
	// Mock goreleaser commands
	suite.mockRunner.SetOutput("which", []string{"goreleaser"}, "/usr/local/bin/goreleaser")
	suite.mockRunner.SetError("goreleaser", []string{"release", "--skip=publish", "--clean"}, nil)

	originalRunner := GetRunner()
	if err := SetRunner(suite.mockRunner); err != nil {
		suite.T().Fatalf("Failed to set mock runner: %v", err)
	}
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			suite.T().Logf("Warning: failed to restore original runner: %v", err)
		}
	}()

	err := suite.release.Test()
	suite.Require().NoError(err)
}

// TestReleaseSnapshotSuccess tests successful snapshot build
func (suite *ReleaseSimpleTestSuite) TestReleaseSnapshotSuccess() {
	// Mock goreleaser commands
	suite.mockRunner.SetOutput("which", []string{"goreleaser"}, "/usr/local/bin/goreleaser")
	suite.mockRunner.SetError("goreleaser", []string{"release", "--snapshot", "--skip=publish", "--clean"}, nil)

	originalRunner := GetRunner()
	if err := SetRunner(suite.mockRunner); err != nil {
		suite.T().Fatalf("Failed to set mock runner: %v", err)
	}
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			suite.T().Logf("Warning: failed to restore original runner: %v", err)
		}
	}()

	err := suite.release.Snapshot()
	suite.Require().NoError(err)
}

// TestReleaseCheckWithConfig tests configuration validation
func (suite *ReleaseSimpleTestSuite) TestReleaseCheckWithConfig() {
	// Mock goreleaser commands
	suite.mockRunner.SetOutput("which", []string{"goreleaser"}, "/usr/local/bin/goreleaser")
	suite.mockRunner.SetError("goreleaser", []string{"check"}, nil)

	originalRunner := GetRunner()
	if err := SetRunner(suite.mockRunner); err != nil {
		suite.T().Fatalf("Failed to set mock runner: %v", err)
	}
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			suite.T().Logf("Warning: failed to restore original runner: %v", err)
		}
	}()

	// Create a mock config file
	err := os.WriteFile(".goreleaser.yml", []byte("# test config"), 0o600)
	suite.Require().NoError(err)
	defer func() {
		if removeErr := os.Remove(".goreleaser.yml"); removeErr != nil {
			suite.T().Logf("Warning: failed to remove .goreleaser.yml: %v", removeErr)
		}
	}()

	err = suite.release.Check()
	suite.Require().NoError(err)
}

// TestReleaseCheckWithoutConfig tests failure when no config file found
func (suite *ReleaseSimpleTestSuite) TestReleaseCheckWithoutConfig() {
	// Mock goreleaser commands
	suite.mockRunner.SetOutput("which", []string{"goreleaser"}, "/usr/local/bin/goreleaser")

	originalRunner := GetRunner()
	if err := SetRunner(suite.mockRunner); err != nil {
		suite.T().Fatalf("Failed to set mock runner: %v", err)
	}
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			suite.T().Logf("Warning: failed to restore original runner: %v", err)
		}
	}()

	// Ensure no config files exist
	configFiles := []string{".goreleaser.yml", ".goreleaser.yaml", "goreleaser.yml", "goreleaser.yaml"}
	for _, file := range configFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			suite.T().Logf("Warning: failed to remove config file %s: %v", file, err)
		}
	}

	err := suite.release.Check()
	suite.Require().Error(err)
	suite.Contains(err.Error(), "no goreleaser configuration file found")
}

// TestReleaseInitSuccess tests successful configuration initialization
func (suite *ReleaseSimpleTestSuite) TestReleaseInitSuccess() {
	// Mock goreleaser commands
	suite.mockRunner.SetOutput("which", []string{"goreleaser"}, "/usr/local/bin/goreleaser")
	suite.mockRunner.SetError("goreleaser", []string{"init"}, nil)

	originalRunner := GetRunner()
	if err := SetRunner(suite.mockRunner); err != nil {
		suite.T().Fatalf("Failed to set mock runner: %v", err)
	}
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			suite.T().Logf("Warning: failed to restore original runner: %v", err)
		}
	}()

	// Ensure no existing config files
	configFiles := []string{".goreleaser.yml", ".goreleaser.yaml", "goreleaser.yml", "goreleaser.yaml"}
	for _, file := range configFiles {
		if err := os.Remove(file); err != nil && !os.IsNotExist(err) {
			suite.T().Logf("Warning: failed to remove config file %s: %v", file, err)
		}
	}

	err := suite.release.Init()
	suite.Require().NoError(err)
}

// TestReleaseInitFailsWhenConfigExists tests failure when config already exists
func (suite *ReleaseSimpleTestSuite) TestReleaseInitFailsWhenConfigExists() {
	// Create existing config file
	err := os.WriteFile(".goreleaser.yml", []byte("# existing config"), 0o600)
	suite.Require().NoError(err)
	defer func() {
		if removeErr := os.Remove(".goreleaser.yml"); removeErr != nil && !os.IsNotExist(removeErr) {
			suite.T().Logf("Warning: failed to remove .goreleaser.yml: %v", removeErr)
		}
	}()

	err = suite.release.Init()
	suite.Require().Error(err)
	suite.Contains(err.Error(), "goreleaser configuration already exists")
}

// TestReleaseChangelog tests changelog generation
func (suite *ReleaseSimpleTestSuite) TestReleaseChangelog() {
	// Mock git commands
	suite.mockRunner.SetOutput("git", []string{"describe", "--tags", "--abbrev=0"}, "v1.0.0")
	suite.mockRunner.SetOutput("git", []string{"log", "--pretty=format:- %s", "v1.0.0..HEAD"}, "- feat: new feature\n- fix: bug fix")

	originalRunner := GetRunner()
	if err := SetRunner(suite.mockRunner); err != nil {
		suite.T().Fatalf("Failed to set mock runner: %v", err)
	}
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			suite.T().Logf("Warning: failed to restore original runner: %v", err)
		}
	}()

	err := suite.release.Changelog()
	suite.Require().NoError(err)
}

// TestReleaseValidateSuccess tests successful release validation
func (suite *ReleaseSimpleTestSuite) TestReleaseValidateSuccess() {
	// Mock commands for successful validation
	suite.mockRunner.SetOutput("git", []string{"status", "--porcelain"}, "")
	suite.mockRunner.SetOutput("git", []string{"describe", "--tags", "--abbrev=0"}, "v1.0.0")
	suite.mockRunner.SetOutput("which", []string{"goreleaser"}, "/usr/local/bin/goreleaser")
	suite.mockRunner.SetError("goreleaser", []string{"check"}, nil)

	originalRunner := GetRunner()
	if err := SetRunner(suite.mockRunner); err != nil {
		suite.T().Fatalf("Failed to set mock runner: %v", err)
	}
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			suite.T().Logf("Warning: failed to restore original runner: %v", err)
		}
	}()

	// Create a mock config file
	err := os.WriteFile(".goreleaser.yml", []byte("# test config"), 0o600)
	suite.Require().NoError(err)
	defer func() {
		if removeErr := os.Remove(".goreleaser.yml"); removeErr != nil {
			suite.T().Logf("Warning: failed to remove .goreleaser.yml: %v", removeErr)
		}
	}()

	err = suite.release.Validate()
	suite.Require().NoError(err)
}

// TestReleaseCleanSuccess tests successful artifact cleanup
func (suite *ReleaseSimpleTestSuite) TestReleaseCleanSuccess() {
	// Mock commands for clean operation
	suite.mockRunner.SetError("rm", []string{"-rf", "dist"}, nil)
	suite.mockRunner.SetError("rm", []string{"-f"}, nil) // for temp files
	suite.mockRunner.SetError("go", []string{"clean", "-cache"}, nil)

	originalRunner := GetRunner()
	if err := SetRunner(suite.mockRunner); err != nil {
		suite.T().Fatalf("Failed to set mock runner: %v", err)
	}
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			suite.T().Logf("Warning: failed to restore original runner: %v", err)
		}
	}()

	// Create a mock dist directory
	err := os.MkdirAll("dist", 0o750)
	suite.Require().NoError(err)
	defer func() {
		if removeErr := os.RemoveAll("dist"); removeErr != nil {
			suite.T().Logf("Warning: failed to remove dist: %v", removeErr)
		}
	}()

	err = suite.release.Clean()
	suite.Require().NoError(err)
}

// TestReleaseTokenHandling tests token environment variable handling
func (suite *ReleaseSimpleTestSuite) TestReleaseTokenHandling() {
	suite.Run("uses github_token when GITHUB_TOKEN not set", func() {
		if err := os.Setenv("github_token", "test-token"); err != nil {
			suite.T().Fatalf("Failed to set github_token: %v", err)
		}

		// Mock goreleaser commands
		suite.mockRunner.SetOutput("which", []string{"goreleaser"}, "/usr/local/bin/goreleaser")
		suite.mockRunner.SetError("goreleaser", []string{"release", "--clean"}, nil)

		originalRunner := GetRunner()
		if err := SetRunner(suite.mockRunner); err != nil {
			suite.T().Fatalf("Failed to set mock runner: %v", err)
		}
		defer func() {
			if err := SetRunner(originalRunner); err != nil {
				suite.T().Logf("Warning: failed to restore original runner: %v", err)
			}
		}()

		err := suite.release.Default()
		suite.Require().NoError(err)

		// Verify GITHUB_TOKEN was set from github_token
		suite.Equal("test-token", os.Getenv("GITHUB_TOKEN"))
	})

	suite.Run("prefers github_token over GITHUB_TOKEN", func() {
		if err := os.Setenv("GITHUB_TOKEN", "old-token"); err != nil {
			suite.T().Fatalf("Failed to set GITHUB_TOKEN: %v", err)
		}
		if err := os.Setenv("github_token", "new-token"); err != nil {
			suite.T().Fatalf("Failed to set github_token: %v", err)
		}

		// Mock goreleaser commands
		suite.mockRunner.SetOutput("which", []string{"goreleaser"}, "/usr/local/bin/goreleaser")
		suite.mockRunner.SetError("goreleaser", []string{"release", "--clean"}, nil)

		originalRunner := GetRunner()
		if err := SetRunner(suite.mockRunner); err != nil {
			suite.T().Fatalf("Failed to set mock runner: %v", err)
		}
		defer func() {
			if err := SetRunner(originalRunner); err != nil {
				suite.T().Logf("Warning: failed to restore original runner: %v", err)
			}
		}()

		err := suite.release.Default()
		suite.Require().NoError(err)

		// Verify GITHUB_TOKEN was updated to github_token value
		suite.Equal("new-token", os.Getenv("GITHUB_TOKEN"))
	})
}

// TestRunReleaseSimpleTestSuite runs the simplified release test suite
func TestRunReleaseSimpleTestSuite(t *testing.T) {
	suite.Run(t, new(ReleaseSimpleTestSuite))
}

// TestReleaseBasicFunctionality tests basic release functionality without mocking
func TestReleaseBasicFunctionality(t *testing.T) {
	t.Parallel()

	t.Run("release namespace creation", func(t *testing.T) {
		t.Parallel()
		release := Release{}
		assert.NotNil(t, release)
	})

	t.Run("factory function", func(t *testing.T) {
		t.Parallel()
		release := NewReleaseNamespace()
		assert.NotNil(t, release)

		// Test interface compliance
		_ = release
	})
}

// BenchmarkReleaseOperations benchmarks release operations
func BenchmarkReleaseOperations(b *testing.B) {
	release := Release{}
	mockRunner := testhelpers.NewMockRunner()
	mockRunner.SetDefaultError(nil)

	originalRunner := GetRunner()
	if err := SetRunner(mockRunner); err != nil {
		b.Fatalf("Failed to set mock runner: %v", err)
	}
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			b.Logf("Warning: failed to restore original runner: %v", err)
		}
	}()

	b.Run("Changelog", func(b *testing.B) {
		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				if err := release.Changelog(); err != nil {
					// Expected - this is a benchmark, just consume the error
					_ = err
				}
			}
		})
	})

	b.Run("Check", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if err := release.Check(); err != nil {
				// Expected - this is a benchmark, just consume the error
				_ = err
			}
		}
	})
}
