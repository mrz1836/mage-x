//go:build mage
// +build mage

package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors
var (
	errRequiredFileMissing = errors.New("required file missing")
	errTODOCommentsFound   = errors.New("found TODO comments")
	errFmtPrintUsage       = errors.New("found fmt.Print* usage (use logging instead)")
)

// This magefile demonstrates how to override built-in MAGE-X commands
// with custom implementations while maintaining access to the original functionality

// Override the default lint command with custom pre-lint checks
// This completely replaces the built-in "magex lint" command
func LintDefault() error {
	utils.Header("🔍 Custom Lint Override")

	// Step 1: Custom pre-lint validation
	utils.Info("Running custom pre-lint checks...")

	// Example: Check for required files
	requiredFiles := []string{"go.mod", "README.md"}
	for _, file := range requiredFiles {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			return fmt.Errorf("❌ %w: %s", errRequiredFileMissing, file)
		}
		utils.Info("✅ Found required file: %s", file)
	}

	// Example: Check for TODO comments (custom rule)
	if err := checkForTODOs(); err != nil {
		utils.Warn("⚠️  Found TODO comments: %v", err)
		// Continue anyway - just a warning
	}

	// Step 2: Call the original MAGE-X lint command
	utils.Info("Running original MAGE-X lint...")
	var l mage.Lint
	if err := l.Default(); err != nil {
		return fmt.Errorf("original lint failed: %w", err)
	}

	// Step 3: Custom post-lint actions
	utils.Info("Running custom post-lint actions...")
	utils.Success("✅ Custom lint override completed successfully!")

	return nil
}

// Override the Build namespace to add custom build logic
// This replaces "magex build:default"
type Build struct{}

// Default overrides the default build command
func (Build) Default() error {
	utils.Header("🏗️ Custom Build Override")

	// Custom pre-build setup
	utils.Info("Setting up custom build environment...")

	// Example: Set custom build tags
	customTags := os.Getenv("CUSTOM_BUILD_TAGS")
	if customTags != "" {
		utils.Info("Using custom build tags: %s", customTags)
		if err := os.Setenv("BUILD_TAGS", customTags); err != nil {
			return fmt.Errorf("failed to set build tags: %w", err)
		}
	}

	// Example: Custom version injection
	if version := os.Getenv("CUSTOM_VERSION"); version != "" {
		utils.Info("Injecting custom version: %s", version)
		// Set ldflags for version injection
		ldflags := fmt.Sprintf("-X main.Version=%s", version)
		if err := os.Setenv("BUILD_LDFLAGS", ldflags); err != nil {
			return fmt.Errorf("failed to set ldflags: %w", err)
		}
	}

	// Call the original MAGE-X build
	utils.Info("Running original MAGE-X build...")
	var b mage.Build
	if err := b.Default(); err != nil {
		return fmt.Errorf("original build failed: %w", err)
	}

	// Custom post-build actions
	utils.Info("Running custom post-build actions...")
	if err := createBuildSummary(); err != nil {
		utils.Warn("Failed to create build summary: %v", err)
		// Continue anyway
	}

	utils.Success("✅ Custom build override completed!")
	return nil
}

// All builds for all platforms with custom logic
func (Build) All() error {
	utils.Header("🌍 Custom Build All Override")

	utils.Info("Custom multi-platform build preparation...")

	// Custom logic before building all platforms
	platforms := []string{"linux/amd64", "darwin/amd64", "windows/amd64"}
	for _, platform := range platforms {
		utils.Info("Preparing build for: %s", platform)
	}

	// Call original build all
	var b mage.Build
	if err := b.All(); err != nil {
		return fmt.Errorf("original build all failed: %w", err)
	}

	// Custom post-build all actions
	utils.Info("Custom post-build-all cleanup...")

	return nil
}

// Override Test namespace partially - only override specific commands
type Test struct {
	mage.Test // Embed original to keep other commands
}

// Unit overrides just the unit test command
func (Test) Unit() error {
	utils.Header("🧪 Custom Unit Test Override")

	// Custom pre-test setup
	utils.Info("Setting up custom test environment...")

	// Example: Set test database
	if err := os.Setenv("TEST_DB", "memory"); err != nil {
		return fmt.Errorf("failed to set test database: %w", err)
	}

	// Call original unit tests
	utils.Info("Running original unit tests...")
	var t mage.Test
	if err := t.Unit(); err != nil {
		return fmt.Errorf("unit tests failed: %w", err)
	}

	// Custom post-test actions
	utils.Info("Generating custom test report...")

	utils.Success("✅ Custom unit tests completed!")
	return nil
}

// Example of completely custom command (not overriding anything)
func CustomLintStrict() error {
	utils.Header("🔒 Strict Custom Lint")

	utils.Info("Running extra-strict linting rules...")

	// First run the overridden default lint
	if err := LintDefault(); err != nil {
		return fmt.Errorf("default lint failed: %w", err)
	}

	// Then add strict rules
	utils.Info("Checking strict coding standards...")

	// Example: No fmt.Print* allowed
	if err := checkForFmtPrint(); err != nil {
		return fmt.Errorf("strict lint failed: %w", err)
	}

	utils.Success("✅ Strict lint passed!")
	return nil
}

// Helper functions for custom logic

func checkForTODOs() error {
	var todos []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor and hidden directories
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor") {
			return filepath.SkipDir
		}

		// Only check .go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Validate path for security
		if !strings.HasPrefix(path, ".") || strings.Contains(path, "..") {
			return nil
		}

		content, readErr := os.ReadFile(path) //nolint:gosec // path validated above
		if readErr != nil {
			return readErr
		}

		if strings.Contains(string(content), "TODO") {
			todos = append(todos, path)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if len(todos) > 0 {
		return fmt.Errorf("%w in: %s", errTODOCommentsFound, strings.Join(todos, ", "))
	}

	return nil
}

func checkForFmtPrint() error {
	var violations []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor") {
			return filepath.SkipDir
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Validate path for security
		if !strings.HasPrefix(path, ".") || strings.Contains(path, "..") {
			return nil
		}

		content, readErr := os.ReadFile(path) //nolint:gosec // path validated above
		if readErr != nil {
			return readErr
		}

		if strings.Contains(string(content), "fmt.Print") {
			violations = append(violations, path)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if len(violations) > 0 {
		return fmt.Errorf("%w in: %s", errFmtPrintUsage, strings.Join(violations, ", "))
	}

	return nil
}

func createBuildSummary() error {
	summary := `# Build Summary

Build completed successfully at: %s
Custom build tags: %s
Custom version: %s

## Build Artifacts
- Binary created with custom configuration
- Version information injected
- Custom build tags applied

## Next Steps
- Test the binary: ./binary-name
- Run tests: magex test
- Deploy: magex deploy (if available)
`

	content := fmt.Sprintf(summary,
		time.Now().Format("2006-01-02 15:04:05"),
		os.Getenv("CUSTOM_BUILD_TAGS"),
		os.Getenv("CUSTOM_VERSION"),
	)

	return os.WriteFile("BUILD_SUMMARY.md", []byte(content), 0o600)
}

// Note: This magefile demonstrates several override patterns:
//
// 1. Complete Function Override (LintDefault):
//    - Replaces the entire built-in command
//    - Can call original command within override
//    - Adds pre/post processing
//
// 2. Namespace Override with Embedding (Build):
//    - Replaces specific namespace commands
//    - Can selectively override methods
//    - Maintains access to originals
//
// 3. Partial Override with Embedding (Test):
//    - Embeds original namespace
//    - Only overrides specific commands
//    - Other commands remain unchanged
//
// 4. Custom Commands (CustomLintStrict):
//    - Completely new functionality
//    - Can build on overridden commands
//    - Extends the available command set
//
// Usage Examples:
//   magex lint              # Runs LintDefault() override
//   magex build             # Runs Build.Default() override
//   magex build:all         # Runs Build.All() override
//   magex test:unit         # Runs Test.Unit() override
//   magex test:race         # Runs original mage.Test.Race() (embedded)
//   magex customlintstrict  # Runs custom command
