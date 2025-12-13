package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Static test errors to satisfy err113 linter
var (
	errBaseTestsFailed   = errors.New("base tests failed")
	errTaggedTestsFailed = errors.New("tagged tests failed")
)

// MockBuildTagRunner implements CommandRunner for testing
type MockBuildTagRunner struct {
	mock.Mock
}

func (m *MockBuildTagRunner) RunCmd(cmd string, args ...string) error {
	allArgs := append([]string{cmd}, args...)
	argsInterface := make([]interface{}, len(allArgs))
	for i, v := range allArgs {
		argsInterface[i] = v
	}

	callArgs := m.Called(argsInterface...)
	return callArgs.Error(0)
}

func (m *MockBuildTagRunner) RunCmdOutput(cmd string, args ...string) (string, error) {
	allArgs := append([]string{cmd}, args...)
	argsInterface := make([]interface{}, len(allArgs))
	for i, v := range allArgs {
		argsInterface[i] = v
	}

	callArgs := m.Called(argsInterface...)
	return callArgs.String(0), callArgs.Error(1)
}

func TestDisplayTestHeader(t *testing.T) {
	// Reset config for clean test
	TestResetConfig()

	t.Run("WithAutoDiscoveryEnabled", func(t *testing.T) {
		config := &Config{
			Test: TestConfig{
				AutoDiscoverBuildTags:        true,
				AutoDiscoverBuildTagsExclude: []string{"integration", "performance"},
				Parallel:                     2,
				Timeout:                      "5m",
				Race:                         false,
				Cover:                        true,
				Verbose:                      true,
				CoverMode:                    "atomic",
				Shuffle:                      false,
				Short:                        false,
			},
		}

		// Create a temporary directory with test files for discovery
		testDir, err := os.MkdirTemp("", "build-tags-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(testDir) }() //nolint:errcheck // Test cleanup

		// Create test files with build tags
		createTestFile(t, testDir, "unit_test.go", "//go:build unit")
		createTestFile(t, testDir, "e2e_test.go", "//go:build e2e")
		createTestFile(t, testDir, "integration_test.go", "//go:build integration")

		// Change to test directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(originalDir) }() //nolint:errcheck // Test cleanup
		require.NoError(t, os.Chdir(testDir))

		tags := displayTestHeader("unit", config)

		// Should discover tags excluding the configured exclusions
		assert.Contains(t, tags, "unit")
		assert.Contains(t, tags, "e2e")
		assert.NotContains(t, tags, "integration") // excluded
	})

	t.Run("WithAutoDiscoveryDisabled", func(t *testing.T) {
		config := &Config{
			Test: TestConfig{
				AutoDiscoverBuildTags: false,
				Parallel:              4,
				Timeout:               "10m",
				Race:                  true,
				Cover:                 false,
				Verbose:               false,
				CoverMode:             "count",
				Shuffle:               true,
				Short:                 true,
			},
		}

		tags := displayTestHeader("integration", config)

		// Should return nil when auto-discovery is disabled
		assert.Nil(t, tags)
	})
}

func TestProcessBuildTagAutoDiscovery(t *testing.T) {
	t.Run("SuccessfulDiscovery", func(t *testing.T) {
		// Create a temporary directory with test files
		testDir, err := os.MkdirTemp("", "build-tags-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(testDir) }() //nolint:errcheck // Test cleanup

		// Create test files with different build tags
		createTestFile(t, testDir, "unit_test.go", "//go:build unit")
		createTestFile(t, testDir, "e2e_test.go", "//go:build e2e")
		createTestFile(t, testDir, "performance_test.go", "//go:build performance")

		// Change to test directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(originalDir) }() //nolint:errcheck // Test cleanup
		require.NoError(t, os.Chdir(testDir))

		config := &Config{
			Test: TestConfig{
				AutoDiscoverBuildTags:        true,
				AutoDiscoverBuildTagsExclude: []string{"performance"},
			},
		}

		tags, info := processBuildTagAutoDiscovery(config)

		assert.Contains(t, tags, "unit")
		assert.Contains(t, tags, "e2e")
		assert.NotContains(t, tags, "performance")

		assert.Contains(t, info, "Auto-Discovery: ✓")
		assert.Contains(t, info, "Found:")
		assert.Contains(t, info, "unit")
		assert.Contains(t, info, "e2e")
	})

	t.Run("NoTagsFound", func(t *testing.T) {
		// Create empty directory
		testDir, err := os.MkdirTemp("", "build-tags-test-*")
		require.NoError(t, err)
		defer func() { _ = os.RemoveAll(testDir) }() //nolint:errcheck // Test cleanup

		// Change to test directory
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(originalDir) }() //nolint:errcheck // Test cleanup
		require.NoError(t, os.Chdir(testDir))

		config := &Config{
			Test: TestConfig{
				AutoDiscoverBuildTags: true,
			},
		}

		tags, info := processBuildTagAutoDiscovery(config)

		assert.Nil(t, tags)
		assert.Contains(t, info, "Auto-Discovery: ✓")
		assert.Contains(t, info, "Found: none")
	})

	t.Run("DiscoveryError", func(t *testing.T) {
		// Change to non-existent directory to cause error
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() { _ = os.Chdir(originalDir) }() //nolint:errcheck // Test cleanup

		config := &Config{
			Test: TestConfig{
				AutoDiscoverBuildTags: true,
			},
		}

		// This should cause an error when trying to discover from non-existent directory
		// Since we can't actually change to a non-existent directory,
		// we'll skip this test and comment it for future implementation
		t.Skip("Cannot test directory change error scenario without mocking filesystem operations")

		tags, info := processBuildTagAutoDiscovery(config)

		assert.Nil(t, tags)
		assert.Contains(t, info, "Auto-Discovery: ✓")
		assert.Contains(t, info, "Error:")
	})
}

func TestRunTestsWithBuildTagDiscoveryTags(t *testing.T) {
	// Mock runner setup
	mockRunner := &MockBuildTagRunner{}
	originalRunner := GetRunner()
	err := SetRunner(mockRunner)
	require.NoError(t, err)
	defer func() {
		_ = SetRunner(originalRunner) //nolint:errcheck // cleanup in defer is acceptable to ignore
	}()

	// Reset config
	TestResetConfig()

	config := &Config{
		Test: TestConfig{
			AutoDiscoverBuildTags: true,
			Parallel:              2,
			Timeout:               "5m",
			Verbose:               false,
		},
	}

	modules := []ModuleInfo{
		{Path: ".", Relative: ".", Name: "test-module", IsRoot: true},
	}

	t.Run("WithAutoDiscoveryEnabled", func(t *testing.T) {
		// Mock successful test runs - expect multiple calls with "go" as first argument
		// The actual calls have 9 additional arguments beyond the command (10 total)
		mockRunner.On("RunCmd", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(3) // Base + 2 discovered tags

		discoveredTags := []string{"unit", "integration"}
		err := runTestsWithBuildTagDiscoveryTags(config, modules, []string{}, "unit", discoveredTags)

		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("WithAutoDiscoveryDisabled", func(t *testing.T) {
		mockRunner.ExpectedCalls = nil // Clear previous expectations

		// Mock single test run (no auto-discovery) - expect 8 arguments
		mockRunner.On("RunCmd", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		config.Test.AutoDiscoverBuildTags = false
		err := runTestsWithBuildTagDiscoveryTags(config, modules, []string{}, "unit", nil)

		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("BaseTestsFailure", func(t *testing.T) {
		mockRunner.ExpectedCalls = nil // Clear previous expectations

		// Mock failing base tests - expect 8 arguments
		mockRunner.On("RunCmd", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errBaseTestsFailed).Once()

		config.Test.AutoDiscoverBuildTags = true
		discoveredTags := []string{"unit"}
		err := runTestsWithBuildTagDiscoveryTags(config, modules, []string{}, "unit", discoveredTags)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "base tests failed")
		mockRunner.AssertExpectations(t)
	})

	t.Run("TaggedTestsFailure", func(t *testing.T) {
		mockRunner.ExpectedCalls = nil // Clear previous expectations

		// Mock successful base tests, but failing tagged tests
		mockRunner.On("RunCmd", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once() // Base tests succeed - expect 8 arguments

		mockRunner.On("RunCmd", mock.AnythingOfType("string"), mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(errTaggedTestsFailed).Once() // Tagged tests fail - expect 10 arguments

		config.Test.AutoDiscoverBuildTags = true
		discoveredTags := []string{"unit"}
		err := runTestsWithBuildTagDiscoveryTags(config, modules, []string{}, "unit", discoveredTags)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "tests with tag 'unit' failed")
		mockRunner.AssertExpectations(t)
	})
}

func TestRunCoverageTestsWithBuildTagDiscoveryTags(t *testing.T) {
	// Mock runner setup
	mockRunner := &MockBuildTagRunner{}
	originalRunner := GetRunner()
	err := SetRunner(mockRunner)
	require.NoError(t, err)
	defer func() {
		_ = SetRunner(originalRunner) //nolint:errcheck // cleanup in defer is acceptable to ignore
	}()

	config := &Config{
		Test: TestConfig{
			AutoDiscoverBuildTags: true,
			CoverMode:             "atomic",
		},
	}

	modules := []ModuleInfo{
		{Path: ".", Relative: ".", Name: "test-module", IsRoot: true},
	}

	t.Run("WithAutoDiscoveryEnabled", func(t *testing.T) {
		// Mock successful coverage run without tags: go test -cover -coverprofile=coverage_0.txt -covermode=atomic ./...
		mockRunner.On("RunCmd", "go", "test", "-cover", mock.MatchedBy(func(arg string) bool {
			return strings.HasPrefix(arg, "-coverprofile=coverage_") && !strings.Contains(arg, "_unit.txt") && !strings.Contains(arg, "_integration.txt")
		}), "-covermode=atomic", "./...").Return(nil).Once()

		// Mock successful coverage runs with tags: go test -tags unit -cover -coverprofile=coverage_0_unit.txt -covermode=atomic ./...
		mockRunner.On("RunCmd", "go", "test", "-tags", mock.AnythingOfType("string"), "-cover", mock.MatchedBy(func(arg string) bool {
			return strings.HasPrefix(arg, "-coverprofile=coverage_") && (strings.Contains(arg, "_unit.txt") || strings.Contains(arg, "_integration.txt"))
		}), "-covermode=atomic", "./...").Return(nil).Times(2)

		// Note: CoverReport is only called if coverage.txt exists, which it doesn't in this test

		discoveredTags := []string{"unit", "integration"}
		err := runCoverageTestsWithBuildTagDiscoveryTags(config, modules, false, []string{}, discoveredTags)
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})

	t.Run("WithAutoDiscoveryDisabled", func(t *testing.T) {
		mockRunner.ExpectedCalls = nil // Clear previous expectations

		// Mock single coverage run - expect coverage test arguments
		mockRunner.On("RunCmd", "go", "test", "-cover", mock.MatchedBy(func(arg string) bool {
			return strings.HasPrefix(arg, "-coverprofile=coverage_")
		}), "-covermode=atomic", "./...").Return(nil).Once()

		// Note: CoverReport is only called if coverage.txt exists, which it doesn't in this test

		config.Test.AutoDiscoverBuildTags = false
		err := runCoverageTestsWithBuildTagDiscoveryTags(config, modules, false, []string{}, nil)
		require.NoError(t, err)
		mockRunner.AssertExpectations(t)
	})
}

func TestGetTagInfo(t *testing.T) {
	tests := []struct {
		name     string
		buildTag string
		expected string
	}{
		{
			name:     "EmptyTag",
			buildTag: "",
			expected: "",
		},
		{
			name:     "NonEmptyTag",
			buildTag: "integration",
			expected: " with tag 'integration'",
		},
		{
			name:     "ComplexTag",
			buildTag: "unit-tests",
			expected: " with tag 'unit-tests'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTagInfo(tt.buildTag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestHandleCoverageFilesWithBuildTag(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "coverage-test-*")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tempDir) }() //nolint:errcheck // Test cleanup

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }() //nolint:errcheck // Test cleanup
	require.NoError(t, os.Chdir(tempDir))

	t.Run("EmptyTag", func(t *testing.T) {
		// Create test coverage files
		coverageFiles := []string{"coverage_1.txt", "coverage_2.txt"}
		for _, file := range coverageFiles {
			err := os.WriteFile(file, []byte("mode: atomic\ntest coverage"), 0o600)
			require.NoError(t, err)
		}

		handleCoverageFilesWithBuildTag(coverageFiles, "")

		// Should create coverage.txt (standard name for base coverage)
		assert.FileExists(t, "coverage.txt")

		// Original files should be cleaned up
		for _, file := range coverageFiles {
			assert.NoFileExists(t, file)
		}

		// Clean up for next test
		_ = os.Remove("coverage.txt") //nolint:errcheck // Test cleanup
	})

	t.Run("WithBuildTag", func(t *testing.T) {
		coverageFiles := []string{"coverage_unit.txt"}
		err := os.WriteFile(coverageFiles[0], []byte("mode: atomic\ntest coverage"), 0o600)
		require.NoError(t, err)

		handleCoverageFilesWithBuildTag(coverageFiles, "unit")

		// Should create coverage_unit.txt
		assert.FileExists(t, "coverage_unit.txt")

		// Original file should be renamed, not exist under old name
		// (but file will exist under new name, which is the same in this case)
	})
}

func TestTitleCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "EmptyString",
			input:    "",
			expected: "",
		},
		{
			name:     "SingleChar",
			input:    "a",
			expected: "A",
		},
		{
			name:     "LowercaseWord",
			input:    "unit",
			expected: "Unit",
		},
		{
			name:     "MixedCase",
			input:    "unitTest",
			expected: "UnitTest",
		},
		{
			name:     "AlreadyTitleCase",
			input:    "Unit",
			expected: "Unit",
		},
		{
			name:     "SpecialChars",
			input:    "unit-test",
			expected: "Unit-test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := titleCase(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions for tests

func createTestFile(t *testing.T, dir, filename, buildTag string) {
	content := fmt.Sprintf(`%s

package testdata

import "testing"

func TestExample(t *testing.T) {
    t.Log("Example test")
}
`, buildTag)

	filePath := filepath.Join(dir, filename)
	err := os.WriteFile(filePath, []byte(content), 0o600)
	require.NoError(t, err)
}
