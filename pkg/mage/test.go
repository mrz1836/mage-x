package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors to satisfy err113 linter
var (
	errNoCoverageFile         = errors.New("no coverage file found. Run 'magex test:cover' first")
	errNoCoverageFilesToMerge = errors.New("no coverage files to merge")
)

// Test namespace for test-related tasks
type Test mg.Namespace

// Default runs the default test suite (unit tests only)
func (Test) Default() error {
	utils.Header("Running Default Test Suite")

	// Run unit tests only - no linting
	return Test{}.Unit()
}

// Full runs the complete test suite with linting
func (Test) Full() error {
	utils.Header("Running Full Test Suite (Lint + Tests)")

	fmt.Printf("\n📋 Step 1/2: Running linters...\n")
	fmt.Printf("════════════════════════════════════════════════\n\n")

	// Run lint first
	if err := (Lint{}).Default(); err != nil {
		return err
	}

	fmt.Printf("\n📋 Step 2/2: Running unit tests...\n")
	fmt.Printf("════════════════════════════════════════════════\n\n")

	// Then run unit tests
	return Test{}.Unit()
}

// Unit runs unit tests
func (Test) Unit() error {
	utils.Header("Running Unit Tests")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Discover all modules
	modules, err := findAllModules()
	if err != nil {
		return fmt.Errorf("failed to find modules: %w", err)
	}

	if len(modules) == 0 {
		utils.Warn("No Go modules found")
		return nil
	}

	// Show modules found
	if len(modules) > 1 {
		utils.Info("Found %d Go modules", len(modules))
	}

	totalStart := time.Now()
	var moduleErrors []moduleError

	// Run tests for each module
	for _, module := range modules {
		displayModuleHeader(module, "Testing")

		moduleStart := time.Now()

		// Build test args
		args := buildTestArgs(config, false, false)
		args = append(args, "-short", "./...")

		// Run tests in module directory
		err := runCommandInModule(module, "go", args...)

		if err != nil {
			moduleErrors = append(moduleErrors, moduleError{Module: module, Error: err})
			utils.Error("Tests failed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		} else {
			utils.Success("Tests passed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		}
	}

	// Report overall results
	if len(moduleErrors) > 0 {
		utils.Error("Unit tests failed in %d/%d modules", len(moduleErrors), len(modules))
		return formatModuleErrors(moduleErrors)
	}

	utils.Success("All unit tests passed in %s", utils.FormatDuration(time.Since(totalStart)))
	return nil
}

// Short runs short tests (excludes integration tests)
func (Test) Short() error {
	utils.Header("Running Short Tests")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Discover all modules
	modules, err := findAllModules()
	if err != nil {
		return fmt.Errorf("failed to find modules: %w", err)
	}

	if len(modules) == 0 {
		utils.Warn("No Go modules found")
		return nil
	}

	// Show modules found
	if len(modules) > 1 {
		utils.Info("Found %d Go modules", len(modules))
	}

	totalStart := time.Now()
	var moduleErrors []moduleError

	// Run tests for each module
	for _, module := range modules {
		displayModuleHeader(module, "Running short tests in")

		moduleStart := time.Now()

		// Explicitly disable race and coverage for short tests to keep them fast
		raceDisabled := false
		coverDisabled := false
		args := buildTestArgsWithOverrides(config, &raceDisabled, &coverDisabled)
		args = append(args, "-short", "./...")

		// Run tests in module directory
		err := runCommandInModule(module, "go", args...)

		if err != nil {
			moduleErrors = append(moduleErrors, moduleError{Module: module, Error: err})
			utils.Error("Short tests failed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		} else {
			utils.Success("Short tests passed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		}
	}

	// Report overall results
	if len(moduleErrors) > 0 {
		utils.Error("Short tests failed in %d/%d modules", len(moduleErrors), len(modules))
		return formatModuleErrors(moduleErrors)
	}

	utils.Success("All short tests passed in %s", utils.FormatDuration(time.Since(totalStart)))
	return nil
}

// Race runs tests with race detector
func (Test) Race() error {
	utils.Header("Running Tests with Race Detector")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Discover all modules
	modules, err := findAllModules()
	if err != nil {
		return fmt.Errorf("failed to find modules: %w", err)
	}

	if len(modules) == 0 {
		utils.Warn("No Go modules found")
		return nil
	}

	// Show modules found
	if len(modules) > 1 {
		utils.Info("Found %d Go modules", len(modules))
	}

	totalStart := time.Now()
	var moduleErrors []moduleError

	// Run tests for each module
	for _, module := range modules {
		displayModuleHeader(module, "Running race tests in")

		moduleStart := time.Now()

		args := buildTestArgs(config, true, false)
		args = append(args, "./...")

		// Run tests in module directory
		err := runCommandInModule(module, "go", args...)

		if err != nil {
			moduleErrors = append(moduleErrors, moduleError{Module: module, Error: err})
			utils.Error("Race tests failed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		} else {
			utils.Success("Race tests passed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		}
	}

	// Report overall results
	if len(moduleErrors) > 0 {
		utils.Error("Race tests failed in %d/%d modules", len(moduleErrors), len(modules))
		return formatModuleErrors(moduleErrors)
	}

	utils.Success("All race tests passed in %s", utils.FormatDuration(time.Since(totalStart)))
	return nil
}

// Cover runs tests with coverage
func (Test) Cover() error {
	utils.Header("Running Tests with Coverage")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Discover all modules
	modules, err := findAllModules()
	if err != nil {
		return fmt.Errorf("failed to find modules: %w", err)
	}

	if len(modules) == 0 {
		utils.Warn("No Go modules found")
		return nil
	}

	// Show modules found
	if len(modules) > 1 {
		utils.Info("Found %d Go modules", len(modules))
	}

	totalStart := time.Now()
	var moduleErrors []moduleError
	var coverageFiles []string

	// Run tests for each module
	for i, module := range modules {
		displayModuleHeader(module, "Running coverage tests in")

		moduleStart := time.Now()

		// Create unique coverage file name for each module
		coverFile := fmt.Sprintf("coverage_%d.txt", i)
		if module.Relative != "." {
			// Use sanitized path for coverage file name
			sanitized := strings.ReplaceAll(module.Relative, "/", "_")
			coverFile = fmt.Sprintf("coverage_%s.txt", sanitized)
		}

		args := buildTestArgs(config, false, true)
		args = append(args, "-coverprofile="+coverFile, "-covermode="+config.Test.CoverMode)

		if len(config.Test.CoverPkg) > 0 {
			args = append(args, "-coverpkg="+strings.Join(config.Test.CoverPkg, ","))
		}

		args = append(args, "./...")

		// Run tests in module directory
		err := runCommandInModule(module, "go", args...)

		if err != nil {
			moduleErrors = append(moduleErrors, moduleError{Module: module, Error: err})
			utils.Error("Coverage tests failed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		} else {
			// Move coverage file to root directory
			srcPath := filepath.Join(module.Path, coverFile)
			dstPath := coverFile
			if utils.FileExists(srcPath) {
				if err := os.Rename(srcPath, dstPath); err != nil {
					utils.Warn("Failed to move coverage file for %s: %v", module.Relative, err)
				} else {
					coverageFiles = append(coverageFiles, coverFile)
				}
			}
			utils.Success("Coverage tests passed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		}
	}

	// Handle coverage files
	handleCoverageFiles(coverageFiles)

	// Report overall results
	if len(moduleErrors) > 0 {
		utils.Error("Coverage tests failed in %d/%d modules", len(moduleErrors), len(modules))
		return formatModuleErrors(moduleErrors)
	}

	utils.Success("All coverage tests passed in %s", utils.FormatDuration(time.Since(totalStart)))

	// Show coverage summary
	if utils.FileExists("coverage.txt") {
		return Test{}.CoverReport()
	}
	return nil
}

// CoverRace runs tests with both coverage and race detector
func (Test) CoverRace() error {
	utils.Header("Running Tests with Coverage and Race Detector")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Discover all modules
	modules, err := findAllModules()
	if err != nil {
		return fmt.Errorf("failed to find modules: %w", err)
	}

	if len(modules) == 0 {
		utils.Warn("No Go modules found")
		return nil
	}

	// Show modules found
	if len(modules) > 1 {
		utils.Info("Found %d Go modules", len(modules))
	}

	totalStart := time.Now()
	var moduleErrors []moduleError
	var coverageFiles []string

	// Run tests for each module
	for i, module := range modules {
		displayModuleHeader(module, "Running coverage+race tests in")

		moduleStart := time.Now()

		// Create unique coverage file name for each module
		coverFile := fmt.Sprintf("coverage_%d.txt", i)
		if module.Relative != "." {
			// Use sanitized path for coverage file name
			sanitized := strings.ReplaceAll(module.Relative, "/", "_")
			coverFile = fmt.Sprintf("coverage_%s.txt", sanitized)
		}

		args := buildTestArgs(config, true, true)
		args = append(args, "-coverprofile="+coverFile, "-covermode=atomic") // atomic is required with race

		if len(config.Test.CoverPkg) > 0 {
			args = append(args, "-coverpkg="+strings.Join(config.Test.CoverPkg, ","))
		}

		args = append(args, "./...")

		// Run tests in module directory
		err := runCommandInModule(module, "go", args...)

		if err != nil {
			moduleErrors = append(moduleErrors, moduleError{Module: module, Error: err})
			utils.Error("Coverage+race tests failed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		} else {
			// Move coverage file to root directory
			srcPath := filepath.Join(module.Path, coverFile)
			dstPath := coverFile
			if utils.FileExists(srcPath) {
				if err := os.Rename(srcPath, dstPath); err != nil {
					utils.Warn("Failed to move coverage file for %s: %v", module.Relative, err)
				} else {
					coverageFiles = append(coverageFiles, coverFile)
				}
			}
			utils.Success("Coverage+race tests passed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		}
	}

	// Handle coverage files
	handleCoverageFiles(coverageFiles)

	// Report overall results
	if len(moduleErrors) > 0 {
		utils.Error("Coverage+race tests failed in %d/%d modules", len(moduleErrors), len(modules))
		return formatModuleErrors(moduleErrors)
	}

	utils.Success("All coverage+race tests passed in %s", utils.FormatDuration(time.Since(totalStart)))

	// Show coverage summary
	if utils.FileExists("coverage.txt") {
		return Test{}.CoverReport()
	}
	return nil
}

// CoverReport shows coverage report
func (Test) CoverReport() error {
	if !utils.FileExists("coverage.txt") {
		utils.Warn("No coverage file found. Run 'magex test:cover' first.")
		return nil
	}

	// Check if this is a multi-module coverage file
	if isMultiModuleCoverage("coverage.txt") {
		utils.Info("Multi-module coverage file detected. Coverage files generated successfully.")
		utils.Info("Note: Individual module coverage reports cannot be displayed with 'go tool cover' for multi-module projects.")
		return nil
	}

	utils.Info("Coverage Report:")
	return GetRunner().RunCmd("go", "tool", "cover", "-func=coverage.txt")
}

// CoverHTML generates HTML coverage report
func (Test) CoverHTML() error {
	if !utils.FileExists("coverage.txt") {
		return errNoCoverageFile
	}

	// Check if this is a multi-module coverage file
	if isMultiModuleCoverage("coverage.txt") {
		utils.Warn("Cannot generate HTML coverage report for multi-module projects.")
		utils.Info("Coverage data is available in coverage.txt for external tools.")
		return nil
	}

	utils.Info("Generating HTML coverage report...")
	if err := GetRunner().RunCmd("go", "tool", "cover", "-html=coverage.txt", "-o=coverage.html"); err != nil {
		return err
	}

	utils.Success("Coverage report generated: coverage.html")

	// Try to open in browser
	if utils.CommandExists("open") {
		// Ignore error - best effort
		err := GetRunner().RunCmd("open", "coverage.html")
		_ = err // Best effort - ignore error
	} else if utils.CommandExists("xdg-open") {
		// Ignore error - best effort
		err := GetRunner().RunCmd("xdg-open", "coverage.html")
		_ = err // Best effort - ignore error
	}

	return nil
}

// handleCoverageFiles processes coverage files by merging multiple files or renaming single file
func handleCoverageFiles(coverageFiles []string) {
	if len(coverageFiles) > 1 {
		utils.Info("\nMerging coverage files...")
		handleMultipleCoverageFiles(coverageFiles)
	} else if len(coverageFiles) == 1 {
		handleSingleCoverageFile(coverageFiles[0])
	}
}

// handleMultipleCoverageFiles merges multiple coverage files
func handleMultipleCoverageFiles(coverageFiles []string) {
	if err := mergeCoverageFiles(coverageFiles, "coverage.txt"); err != nil {
		utils.Warn("Failed to merge coverage files: %v", err)
		return
	}

	// Clean up individual coverage files
	for _, file := range coverageFiles {
		if err := os.Remove(file); err != nil {
			utils.Warn("Failed to remove coverage file %s: %v", file, err)
		}
	}
}

// handleSingleCoverageFile renames single coverage file to standard name
func handleSingleCoverageFile(coverageFile string) {
	if err := os.Rename(coverageFile, "coverage.txt"); err != nil {
		utils.Warn("Failed to rename coverage file: %v", err)
	}
}

// isMultiModuleCoverage checks if a coverage file contains packages from multiple modules
func isMultiModuleCoverage(coverageFile string) bool {
	content, err := os.ReadFile(coverageFile) // #nosec G304 -- coverage file path is controlled
	if err != nil {
		utils.Debug("Failed to read coverage file: %v", err)
		return false
	}

	// Get the current module name
	currentModule, err := utils.GetModuleName()
	if err != nil {
		utils.Debug("Failed to get module name: %v", err)
		return false
	}
	utils.Debug("Current module: %s", currentModule)

	// Check if any line contains a package path from a submodule
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if line == "" || strings.HasPrefix(line, "mode:") {
			continue
		}
		// Coverage lines start with the package path
		if idx := strings.Index(line, ":"); idx > 0 {
			pkg := line[:idx]
			// Check if this package is from a submodule by looking for paths that contain
			// the main module but have their own go.mod (like .github/test-module)
			if strings.HasPrefix(pkg, currentModule+"/") {
				// Extract the relative path after the module name
				relativePath := strings.TrimPrefix(pkg, currentModule+"/")
				// Check for common submodule patterns
				if strings.HasPrefix(relativePath, ".github/test-module/") ||
					strings.HasPrefix(relativePath, "tools/cli-helper/") {
					utils.Debug("Found package from submodule: %s", pkg)
					return true
				}
			} else if !strings.HasPrefix(pkg, currentModule) && strings.Contains(pkg, "/") {
				// Package from completely different module
				utils.Debug("Found package from different module: %s", pkg)
				return true
			}
		}
	}
	return false
}

// Fuzz runs fuzz tests
func (Test) Fuzz() error {
	utils.Header("Running Fuzz Tests")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	if config.Test.SkipFuzz {
		utils.Info("Fuzz tests skipped")
		return nil
	}

	// Find packages with fuzz tests
	packages := findFuzzPackages()

	if len(packages) == 0 {
		utils.Info("No fuzz tests found")
		return nil
	}

	fuzzTime := utils.GetEnv("FUZZ_TIME", "10s")

	for _, pkg := range packages {
		// List fuzz tests in package
		output, err := GetRunner().RunCmdOutput("go", "test", "-list", "^Fuzz", pkg)
		if err != nil {
			continue
		}

		fuzzTests := strings.Split(strings.TrimSpace(output), "\n")
		for _, test := range fuzzTests {
			if !strings.HasPrefix(test, "Fuzz") {
				continue
			}

			utils.Info("Fuzzing %s.%s", pkg, test)

			args := []string{"test", "-run=^$", fmt.Sprintf("-fuzz=^%s$", test)}
			args = append(args, "-fuzztime", fuzzTime)

			if config.Test.Verbose {
				args = append(args, "-v")
			}

			args = append(args, pkg)

			if err := GetRunner().RunCmd("go", args...); err != nil {
				return fmt.Errorf("fuzz test %s failed: %w", test, err)
			}
		}
	}

	utils.Success("Fuzz tests completed")
	return nil
}

// FuzzShort runs fuzz tests with shorter duration for quick feedback
func (Test) FuzzShort() error {
	utils.Header("Running Short Fuzz Tests")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	if config.Test.SkipFuzz {
		utils.Info("Fuzz tests skipped")
		return nil
	}

	// Find packages with fuzz tests
	packages := findFuzzPackages()

	if len(packages) == 0 {
		utils.Info("No fuzz tests found")
		return nil
	}

	// Use shorter fuzz time for quick feedback (5s instead of 10s)
	fuzzTime := utils.GetEnv("FUZZ_TIME", "5s")

	for _, pkg := range packages {
		// List fuzz tests in package
		output, err := GetRunner().RunCmdOutput("go", "test", "-list", "^Fuzz", pkg)
		if err != nil {
			continue
		}

		fuzzTests := strings.Split(strings.TrimSpace(output), "\n")
		for _, test := range fuzzTests {
			if !strings.HasPrefix(test, "Fuzz") {
				continue
			}

			utils.Info("Fuzzing %s.%s", pkg, test)

			args := []string{"test", "-run=^$", fmt.Sprintf("-fuzz=^%s$", test)}
			args = append(args, "-fuzztime", fuzzTime)

			if config.Test.Verbose {
				args = append(args, "-v")
			}

			args = append(args, pkg)

			if err := GetRunner().RunCmd("go", args...); err != nil {
				return fmt.Errorf("short fuzz test %s failed: %w", test, err)
			}
		}
	}

	utils.Success("Short fuzz tests completed")
	return nil
}

// Bench runs benchmarks
func (Test) Bench(argsList ...string) error {
	utils.Header("Running Benchmarks")

	// Parse command-line parameters
	params := utils.ParseParams(argsList)

	config, err := GetConfig()
	if err != nil {
		return err
	}

	args := []string{"test", "-bench=.", "-benchmem", "-run=^$"}

	if config.Test.Verbose {
		args = append(args, "-v")
	}

	if config.Test.Tags != "" {
		args = append(args, "-tags", config.Test.Tags)
	}

	// Get benchmark time from parameter, fallback to environment, then default
	benchTime := utils.GetParam(params, "time", "")
	if benchTime == "" {
		benchTime = utils.GetEnv("BENCH_TIME", "10s")
	}
	args = append(args, "-benchtime", benchTime)

	// Get count from parameter or environment
	count := utils.GetParam(params, "count", "")
	if count == "" {
		count = utils.GetEnv("BENCH_COUNT", "")
	}
	if count != "" {
		args = append(args, "-count", count)
	}

	args = append(args, "./...")

	start := time.Now()
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("benchmarks failed: %w", err)
	}

	utils.Success("Benchmarks completed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// BenchShort runs benchmarks with shorter duration for quick feedback
func (Test) BenchShort(argsList ...string) error {
	utils.Header("Running Short Benchmarks")

	// Parse command-line parameters
	params := utils.ParseParams(argsList)

	config, err := GetConfig()
	if err != nil {
		return err
	}

	args := []string{"test", "-bench=.", "-benchmem", "-run=^$"}

	if config.Test.Verbose {
		args = append(args, "-v")
	}

	if config.Test.Tags != "" {
		args = append(args, "-tags", config.Test.Tags)
	}

	// Get benchmark time from parameter, fallback to environment, then default (1s for short)
	benchTime := utils.GetParam(params, "time", "")
	if benchTime == "" {
		benchTime = utils.GetEnv("BENCH_TIME", "1s")
	}
	args = append(args, "-benchtime", benchTime)

	// Get count from parameter or environment
	count := utils.GetParam(params, "count", "")
	if count == "" {
		count = utils.GetEnv("BENCH_COUNT", "")
	}
	if count != "" {
		args = append(args, "-count", count)
	}

	args = append(args, "./...")

	start := time.Now()
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("short benchmarks failed: %w", err)
	}

	utils.Success("Short benchmarks completed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// Integration runs integration tests
func (Test) Integration() error {
	utils.Header("Running Integration Tests")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Set integration test tag
	tags := config.Test.Tags
	if tags != "" {
		tags += ",integration"
	} else {
		tags = "integration"
	}

	args := []string{"test"}
	args = append(args, "-tags", tags)

	if config.Test.Verbose {
		args = append(args, "-v")
	}

	// Longer timeout for integration tests and don't run short tests
	args = append(args, "-timeout", "30m", "-run", "Integration", "./...")

	start := time.Now()
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("integration tests failed: %w", err)
	}

	utils.Success("Integration tests passed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// CI runs the CI test suite
func (Test) CI() error {
	utils.Header("Running CI Test Suite")

	// Set CI environment
	if err := os.Setenv("CI", "true"); err != nil {
		return fmt.Errorf("failed to set CI environment: %w", err)
	}

	// Run tests with race detector and coverage
	return Test{}.CoverRace()
}

// Parallel runs tests in parallel (faster for large repos)
func (Test) Parallel() error {
	utils.Header("Running Tests in Parallel")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Force parallel execution
	parallel := getCPUCount()

	args := []string{"test", "-p", fmt.Sprintf("%d", parallel)}

	if config.Test.Verbose {
		args = append(args, "-v")
	}

	if config.Test.Tags != "" {
		args = append(args, "-tags", config.Test.Tags)
	}

	args = append(args, "./...")

	start := time.Now()
	utils.Info("Running tests with %d parallel workers", parallel)

	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("parallel tests failed: %w", err)
	}

	utils.Success("Parallel tests passed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// NoLint runs tests without linting
func (Test) NoLint() error {
	utils.Header("Running Tests (No Lint)")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Run unit tests without linting
	args := buildTestArgs(config, false, false)

	// Force parallel execution
	parallel := getCPUCount()
	args = append(args, "-p", fmt.Sprintf("%d", parallel), "./...")

	start := time.Now()
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	utils.Success("Tests passed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// CINoRace runs CI tests without race detector
func (Test) CINoRace() error {
	utils.Header("Running CI Test Suite (No Race)")

	// Set CI environment
	if err := os.Setenv("CI", "true"); err != nil {
		return fmt.Errorf("failed to set CI environment: %w", err)
	}

	// Run tests with coverage only (no race)
	return Test{}.Cover()
}

// Helper functions

// buildTestArgs builds common test arguments
func buildTestArgs(cfg *Config, race, cover bool) []string {
	args := []string{"test"}

	if cfg.Test.Parallel > 0 {
		args = append(args, "-p", fmt.Sprintf("%d", cfg.Test.Parallel))
	}

	if cfg.Test.Verbose {
		args = append(args, "-v")
	}

	if cfg.Test.Timeout != "" {
		args = append(args, "-timeout", cfg.Test.Timeout)
	}

	if cfg.Test.Tags != "" {
		args = append(args, "-tags", cfg.Test.Tags)
	}

	if race || cfg.Test.Race {
		args = append(args, "-race")
	}

	if cover || cfg.Test.Cover {
		args = append(args, "-cover")
	}

	return args
}

// buildTestArgsWithOverrides builds test arguments with explicit overrides for race and cover
// When raceOverride or coverOverride are not nil, they take precedence over config defaults
func buildTestArgsWithOverrides(cfg *Config, raceOverride, coverOverride *bool) []string {
	args := []string{"test"}

	if cfg.Test.Parallel > 0 {
		args = append(args, "-p", fmt.Sprintf("%d", cfg.Test.Parallel))
	}

	if cfg.Test.Verbose {
		args = append(args, "-v")
	}

	if cfg.Test.Timeout != "" {
		args = append(args, "-timeout", cfg.Test.Timeout)
	}

	if cfg.Test.Tags != "" {
		args = append(args, "-tags", cfg.Test.Tags)
	}

	// Handle race flag with explicit override
	useRace := cfg.Test.Race
	if raceOverride != nil {
		useRace = *raceOverride
	}
	if useRace {
		args = append(args, "-race")
	}

	// Handle cover flag with explicit override
	useCover := cfg.Test.Cover
	if coverOverride != nil {
		useCover = *coverOverride
	}
	if useCover {
		args = append(args, "-cover")
	}

	return args
}

// findFuzzPackages finds packages containing fuzz tests
func findFuzzPackages() []string {
	output, err := GetRunner().RunCmdOutput("grep", "-r", "-l", "^func Fuzz", "--include=*_test.go", ".")
	if err != nil {
		// grep returns error if no matches found
		return nil
	}

	files := strings.Split(strings.TrimSpace(output), "\n")
	packageMap := make(map[string]bool)

	for _, file := range files {
		if file == "" {
			continue
		}

		dir := filepath.Dir(file)
		pkg := strings.TrimPrefix(dir, "./")
		if pkg == "." {
			pkg = ""
		}

		module, err := utils.GetModuleName()
		if err != nil {
			// If we can't get module name, skip this package
			continue
		}
		if module != "" && pkg != "" {
			packageMap[filepath.Join(module, pkg)] = true
		} else if module != "" {
			packageMap[module] = true
		}
	}

	packages := make([]string, 0, len(packageMap))
	for pkg := range packageMap {
		packages = append(packages, pkg)
	}

	return packages
}

// Additional methods for Test namespace required by tests

// Run runs tests
func (Test) Run() error {
	runner := GetRunner()
	return runner.RunCmd("go", "test", "./...")
}

// Coverage generates test coverage
func (Test) Coverage(_ ...string) error {
	runner := GetRunner()
	return runner.RunCmd("go", "test", "-cover", "./...")
}

// Vet runs go vet
func (Test) Vet() error {
	runner := GetRunner()
	return runner.RunCmd("go", "vet", "./...")
}

// Lint runs test linting
func (Test) Lint() error {
	runner := GetRunner()
	if err := runner.RunCmd("go", "vet", "./..."); err != nil {
		return fmt.Errorf("go vet failed: %w", err)
	}
	return nil
}

// Clean cleans test artifacts
func (Test) Clean() error {
	runner := GetRunner()
	if err := runner.RunCmd("go", "clean", "-testcache"); err != nil {
		return fmt.Errorf("failed to clean test cache: %w", err)
	}
	return nil
}

// All runs all tests
func (Test) All() error {
	runner := GetRunner()
	return runner.RunCmd("go", "test", "./...")
}

// mergeCoverageFiles merges multiple coverage files into a single file
func mergeCoverageFiles(files []string, output string) error {
	if len(files) == 0 {
		return errNoCoverageFilesToMerge
	}

	// Read all coverage files
	var allLines []string
	modeSet := false

	for _, file := range files {
		content, err := os.ReadFile(file) // #nosec G304 -- coverage file path from controlled list
		if err != nil {
			return fmt.Errorf("failed to read coverage file %s: %w", file, err)
		}

		lines := strings.Split(string(content), "\n")
		for i, line := range lines {
			if i == 0 && strings.HasPrefix(line, "mode:") {
				if !modeSet {
					modeSet = true
					allLines = append(allLines, line)
				}
				// Skip mode line for subsequent files
				continue
			}
			if line != "" {
				allLines = append(allLines, line)
			}
		}
	}

	// Write merged coverage file
	mergedContent := strings.Join(allLines, "\n")
	if err := os.WriteFile(output, []byte(mergedContent), 0o600); err != nil { // #nosec G306 -- coverage output file permissions
		return fmt.Errorf("failed to write merged coverage file: %w", err)
	}

	return nil
}
