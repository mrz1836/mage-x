//go:build integration
// +build integration

package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mrz1836/mage-x/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBuildTagAutoDiscoveryIntegration tests the complete build tag auto-discovery workflow
func TestBuildTagAutoDiscoveryIntegration(t *testing.T) {
	// Create a temporary project directory
	tempDir, err := os.MkdirTemp("", "mage-buildtag-integration-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	require.NoError(t, os.Chdir(tempDir))

	// Reset config for clean test
	originalProvider := GetConfigProvider()
	defer func() {
		SetConfigProvider(originalProvider)
		TestResetConfig()
	}()
	TestResetConfig()

	t.Run("FullWorkflowWithRealFiles", func(t *testing.T) {
		// Create a realistic Go project structure with build tagged files
		createIntegrationProjectStructure(t, tempDir)

		// Create .mage.yaml config with build tag auto-discovery enabled
		createMageConfig(t, tempDir, true, []string{"performance"})

		// Test that GetConfig() loads the auto-discovery settings correctly
		config, err := GetConfig()
		require.NoError(t, err)
		assert.True(t, config.Test.AutoDiscoverBuildTags)
		assert.Contains(t, config.Test.AutoDiscoverBuildTagsExclude, "performance")

		// Test build tag discovery
		tags, err := GetDiscoveredBuildTags(config)
		require.NoError(t, err)

		// Verify discovered tags
		assert.Contains(t, tags, "unit")
		assert.Contains(t, tags, "integration")
		assert.Contains(t, tags, "e2e")
		assert.NotContains(t, tags, "performance") // Should be excluded

		// Test displayTestHeader with auto-discovery
		discoveredTags := displayTestHeader("unit", config)
		assert.Contains(t, discoveredTags, "unit")
		assert.Contains(t, discoveredTags, "integration")
		assert.Contains(t, discoveredTags, "e2e")
	})

	t.Run("EnvironmentVariableOverrides", func(t *testing.T) {
		// Create basic project structure
		createIntegrationProjectStructure(t, tempDir)

		// Test environment variable overrides
		require.NoError(t, os.Setenv("MAGE_X_AUTO_DISCOVER_BUILD_TAGS", "true"))
		require.NoError(t, os.Setenv("MAGE_X_AUTO_DISCOVER_BUILD_TAGS_EXCLUDE", "integration,e2e"))
		defer func() {
			os.Unsetenv("MAGE_X_AUTO_DISCOVER_BUILD_TAGS")
			os.Unsetenv("MAGE_X_AUTO_DISCOVER_BUILD_TAGS_EXCLUDE")
		}()

		// Reset config to pick up environment variables
		TestResetConfig()

		config, err := GetConfig()
		require.NoError(t, err)

		assert.True(t, config.Test.AutoDiscoverBuildTags)
		assert.Contains(t, config.Test.AutoDiscoverBuildTagsExclude, "integration")
		assert.Contains(t, config.Test.AutoDiscoverBuildTagsExclude, "e2e")

		// Test discovery with exclusions
		tags, err := GetDiscoveredBuildTags(config)
		require.NoError(t, err)

		assert.Contains(t, tags, "unit")
		assert.Contains(t, tags, "performance")
		assert.NotContains(t, tags, "integration")
		assert.NotContains(t, tags, "e2e")
	})

	t.Run("DisabledAutoDiscovery", func(t *testing.T) {
		// Create project structure
		createIntegrationProjectStructure(t, tempDir)

		// Create config with auto-discovery disabled
		createMageConfig(t, tempDir, false, nil)

		config, err := GetConfig()
		require.NoError(t, err)
		assert.False(t, config.Test.AutoDiscoverBuildTags)

		// Build tag discovery should return empty when disabled
		tags, err := GetDiscoveredBuildTags(config)
		require.NoError(t, err)
		assert.Empty(t, tags)
	})

	t.Run("ComplexBuildExpressions", func(t *testing.T) {
		// Create files with complex build expressions
		createComplexBuildTagFiles(t, tempDir)

		// Create config with auto-discovery enabled
		createMageConfig(t, tempDir, true, nil)

		config, err := GetConfig()
		require.NoError(t, err)

		tags, err := GetDiscoveredBuildTags(config)
		require.NoError(t, err)

		// Should discover all individual tags from complex expressions
		assert.Contains(t, tags, "linux")
		assert.Contains(t, tags, "windows")
		assert.Contains(t, tags, "cgo")
		assert.Contains(t, tags, "debug")
		assert.Contains(t, tags, "fast")
	})

	t.Run("CoverageTestsWithBuildTags", func(t *testing.T) {
		// Create basic project structure
		createIntegrationProjectStructure(t, tempDir)

		// Create config with auto-discovery enabled
		createMageConfig(t, tempDir, true, nil)

		config, err := GetConfig()
		require.NoError(t, err)
		assert.True(t, config.Test.AutoDiscoverBuildTags) // Use config

		// Test coverage file naming with build tags
		modules := []ModuleInfo{{Path: ".", Relative: ".", Name: "test-module", IsRoot: true}}
		assert.Len(t, modules, 1) // Use modules

		// Create mock coverage files
		coverageFiles := []string{"coverage_1.txt", "coverage_2.txt"}
		for _, file := range coverageFiles {
			content := "mode: atomic\ntest/pkg coverage 1\n"
			err := os.WriteFile(file, []byte(content), 0o644)
			require.NoError(t, err)
		}

		// Test handleCoverageFilesWithBuildTag
		handleCoverageFilesWithBuildTag(coverageFiles, "")
		assert.FileExists(t, "coverage.txt")

		// Clean up
		os.Remove("coverage.txt")

		// Test with build tag
		for _, file := range coverageFiles {
			content := "mode: atomic\ntest/pkg coverage 1\n"
			err := os.WriteFile(file, []byte(content), 0o644)
			require.NoError(t, err)
		}

		handleCoverageFilesWithBuildTag(coverageFiles, "unit")
		assert.FileExists(t, "coverage_unit.txt")

		// Clean up
		os.Remove("coverage_unit.txt")
	})
}

// TestBuildTagConfigurationEdgeCases tests edge cases in build tag configuration
func TestBuildTagConfigurationEdgeCases(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "mage-buildtag-edge-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)
	require.NoError(t, os.Chdir(tempDir))

	// Reset config for clean test
	originalProvider := GetConfigProvider()
	defer func() {
		SetConfigProvider(originalProvider)
		TestResetConfig()
	}()
	TestResetConfig()

	t.Run("EmptyExclusionList", func(t *testing.T) {
		createIntegrationProjectStructure(t, tempDir)
		createMageConfig(t, tempDir, true, []string{})

		config, err := GetConfig()
		require.NoError(t, err)
		assert.True(t, config.Test.AutoDiscoverBuildTags)
		assert.Empty(t, config.Test.AutoDiscoverBuildTagsExclude)

		tags, err := GetDiscoveredBuildTags(config)
		require.NoError(t, err)
		assert.Greater(t, len(tags), 0)
	})

	t.Run("WhitespaceInExclusions", func(t *testing.T) {
		require.NoError(t, os.Setenv("MAGE_X_AUTO_DISCOVER_BUILD_TAGS_EXCLUDE", "  unit  ,  integration  ,  e2e  "))
		defer os.Unsetenv("MAGE_X_AUTO_DISCOVER_BUILD_TAGS_EXCLUDE")

		TestResetConfig()
		config, err := GetConfig()
		require.NoError(t, err)

		// Whitespace should be trimmed
		assert.Contains(t, config.Test.AutoDiscoverBuildTagsExclude, "unit")
		assert.Contains(t, config.Test.AutoDiscoverBuildTagsExclude, "integration")
		assert.Contains(t, config.Test.AutoDiscoverBuildTagsExclude, "e2e")

		// Should not contain whitespace
		for _, tag := range config.Test.AutoDiscoverBuildTagsExclude {
			assert.Equal(t, tag, strings.TrimSpace(tag), "Tag should not contain whitespace: %q", tag)
		}
	})

	t.Run("InvalidConfigurationValues", func(t *testing.T) {
		// Test with invalid boolean values
		require.NoError(t, os.Setenv("MAGE_X_AUTO_DISCOVER_BUILD_TAGS", "maybe"))
		defer os.Unsetenv("MAGE_X_AUTO_DISCOVER_BUILD_TAGS")

		TestResetConfig()
		config, err := GetConfig()
		require.NoError(t, err)

		// Invalid values should not change the default
		assert.False(t, config.Test.AutoDiscoverBuildTags)
	})
}

// Helper functions for integration tests

func createIntegrationProjectStructure(t *testing.T, baseDir string) {
	// Create main.go
	mainContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
	err := os.WriteFile(filepath.Join(baseDir, "main.go"), []byte(mainContent), 0o644)
	require.NoError(t, err)

	// Create go.mod
	goModContent := `module test-project

go 1.21
`
	err = os.WriteFile(filepath.Join(baseDir, "go.mod"), []byte(goModContent), 0o644)
	require.NoError(t, err)

	// Create unit test file
	unitTestContent := `//go:build unit
// +build unit

package main

import "testing"

func TestUnit(t *testing.T) {
	t.Log("Unit test")
}
`
	err = os.WriteFile(filepath.Join(baseDir, "unit_test.go"), []byte(unitTestContent), 0o644)
	require.NoError(t, err)

	// Create integration test file
	integrationTestContent := `//go:build integration
// +build integration

package main

import "testing"

func TestIntegration(t *testing.T) {
	t.Log("Integration test")
}
`
	err = os.WriteFile(filepath.Join(baseDir, "integration_test.go"), []byte(integrationTestContent), 0o644)
	require.NoError(t, err)

	// Create e2e test file
	e2eTestContent := `//go:build e2e
// +build e2e

package main

import "testing"

func TestE2E(t *testing.T) {
	t.Log("E2E test")
}
`
	err = os.WriteFile(filepath.Join(baseDir, "e2e_test.go"), []byte(e2eTestContent), 0o644)
	require.NoError(t, err)

	// Create performance test file
	performanceTestContent := `//go:build performance
// +build performance

package main

import "testing"

func TestPerformance(t *testing.T) {
	t.Log("Performance test")
}
`
	err = os.WriteFile(filepath.Join(baseDir, "performance_test.go"), []byte(performanceTestContent), 0o644)
	require.NoError(t, err)

	// Create subdirectory with more test files
	subDir := filepath.Join(baseDir, "pkg", "util")
	err = os.MkdirAll(subDir, 0o755)
	require.NoError(t, err)

	subTestContent := `//go:build unit
// +build unit

package util

import "testing"

func TestUtilUnit(t *testing.T) {
	t.Log("Util unit test")
}
`
	err = os.WriteFile(filepath.Join(subDir, "util_test.go"), []byte(subTestContent), 0o644)
	require.NoError(t, err)
}

func createComplexBuildTagFiles(t *testing.T, baseDir string) {
	// File with complex AND expression
	andContent := `//go:build linux && cgo && !debug
// +build linux,cgo,!debug

package main

import "testing"

func TestLinuxCgo(t *testing.T) {
	t.Log("Linux CGO test")
}
`
	err := os.WriteFile(filepath.Join(baseDir, "linux_cgo_test.go"), []byte(andContent), 0o644)
	require.NoError(t, err)

	// File with complex OR expression
	orContent := `//go:build (windows || darwin) && fast
// +build windows darwin
// +build fast

package main

import "testing"

func TestWindowsDarwinFast(t *testing.T) {
	t.Log("Windows/Darwin fast test")
}
`
	err = os.WriteFile(filepath.Join(baseDir, "windows_darwin_test.go"), []byte(orContent), 0o644)
	require.NoError(t, err)

	// File with parentheses and negation
	complexContent := `//go:build !(windows && debug) && (cgo || fast)
// +build !windows !debug
// +build cgo fast

package main

import "testing"

func TestComplex(t *testing.T) {
	t.Log("Complex build tag test")
}
`
	err = os.WriteFile(filepath.Join(baseDir, "complex_test.go"), []byte(complexContent), 0o644)
	require.NoError(t, err)
}

func createMageConfig(t *testing.T, baseDir string, autoDiscover bool, excludeTags []string) {
	configContent := fmt.Sprintf(`test:
  auto_discover_build_tags: %t`, autoDiscover)

	if len(excludeTags) > 0 {
		configContent += "\n  auto_discover_build_tags_exclude: ["
		for i, tag := range excludeTags {
			if i > 0 {
				configContent += ", "
			}
			configContent += fmt.Sprintf(`"%s"`, tag)
		}
		configContent += "]"
	}

	err := os.WriteFile(filepath.Join(baseDir, ".mage.yaml"), []byte(configContent), 0o644)
	require.NoError(t, err)
}

// GetDiscoveredBuildTags is a helper function to get discovered build tags for testing
func GetDiscoveredBuildTags(config *Config) ([]string, error) {
	if !config.Test.AutoDiscoverBuildTags {
		return nil, nil
	}

	tags, err := utils.DiscoverBuildTagsFromCurrentDir(config.Test.AutoDiscoverBuildTagsExclude)
	if err != nil {
		return nil, fmt.Errorf("failed to discover build tags: %w", err)
	}

	return tags, nil
}
