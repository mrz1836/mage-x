package mage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/magefile/mage/mg"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors to satisfy err113 linter
var (
	errNoCoverageFile         = errors.New("no coverage file found. Run 'magex test:cover' first")
	errNoCoverageFilesToMerge = errors.New("no coverage files to merge")
	errFlagNotAllowed         = errors.New("flag not allowed for security reasons")
	errFuzzTestFailed         = errors.New("fuzz test(s) failed")
)

// fuzzTestResult captures the result of a single fuzz test run
type fuzzTestResult struct {
	Package  string
	Test     string
	Duration time.Duration
	Error    error
	Output   string // Captured stdout output for text parsing
}

// getCIParams extracts CI-related parameters from args and returns the remaining args
func getCIParams(args []string) (params map[string]string, remainingArgs []string) {
	params = make(map[string]string)
	for _, arg := range args {
		// Check for ci parameter (ci=true, ci=false, etc.)
		if strings.HasPrefix(arg, "ci=") {
			params["ci"] = strings.TrimPrefix(arg, "ci=")
		} else if arg == "ci" {
			params["ci"] = trueValue
		} else {
			remainingArgs = append(remainingArgs, arg)
		}
	}
	return params, remainingArgs
}

// getTestRunner returns a CI-wrapped runner if CI mode is enabled, otherwise the standard runner
func getTestRunner(params map[string]string, config *Config) CommandRunner {
	return GetCIRunner(GetRunner(), params, config)
}

// generateCIReport generates the final CI report if using a CI runner
func generateCIReport(runner CommandRunner) {
	if ciRunner, ok := runner.(CIRunner); ok {
		if err := ciRunner.GenerateReport(); err != nil {
			utils.Warn("Failed to generate CI report: %v", err)
		}
	}
}

// shouldExcludeModule checks if a module should be excluded from processing
// based on the configured exclusion list. By default, "magefiles" modules are excluded
// because they have special build constraints (//go:build mage).
func shouldExcludeModule(module ModuleInfo, config *Config) bool {
	if config == nil || len(config.Test.ExcludeModules) == 0 {
		return false
	}
	for _, excludedName := range config.Test.ExcludeModules {
		if module.Name == excludedName {
			return true
		}
	}
	return false
}

// filterModulesForProcessing filters modules based on the exclusion configuration.
// This is used by test, lint, and deps commands to skip modules like "magefiles"
// that have special build constraints.
func filterModulesForProcessing(modules []ModuleInfo, config *Config, operation string) []ModuleInfo {
	if config == nil || len(config.Test.ExcludeModules) == 0 {
		return modules
	}
	filtered := make([]ModuleInfo, 0, len(modules))
	for _, m := range modules {
		if shouldExcludeModule(m, config) {
			utils.Info("Skipping module %s (excluded from %s)", m.Name, operation)
			continue
		}
		filtered = append(filtered, m)
	}
	return filtered
}

// Test namespace for test-related tasks
type Test mg.Namespace

// Default runs the default test suite (unit tests only)
func (Test) Default(args ...string) error {
	utils.Header("Running Default Test Suite")

	// Run unit tests only - no linting
	return Test{}.Unit(args...)
}

// Full runs the complete test suite with linting
func (Test) Full(args ...string) error {
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
	return Test{}.Unit(args...)
}

// Unit runs unit tests
func (Test) Unit(args ...string) error {
	// Extract CI params from args
	ciParams, remainingArgs := getCIParams(args)

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Get CI-aware runner
	runner := getTestRunner(ciParams, config)
	defer generateCIReport(runner)

	// Display header with build tag information
	discoveredTags := displayTestHeader("unit", config)

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

	// Use new build tag discovery if enabled, passing pre-discovered tags
	return runTestsWithBuildTagDiscoveryTagsWithRunner(config, modules, false, remainingArgs, "unit", discoveredTags, runner)
}

// Short runs short tests (excludes integration tests)
func (Test) Short(args ...string) error {
	// Extract CI params from args
	ciParams, remainingArgs := getCIParams(args)

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Get CI-aware runner
	runner := getTestRunner(ciParams, config)
	defer generateCIReport(runner)

	// Display header with build tag information
	discoveredTags := displayTestHeader("short", config)

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

	// Use new build tag discovery if enabled (explicitly disable race for speed)
	return runTestsWithBuildTagDiscoveryTagsWithRunner(config, modules, false, remainingArgs, "short", discoveredTags, runner)
}

// Race runs tests with race detector
func (Test) Race(args ...string) error {
	// Extract CI params from args
	ciParams, remainingArgs := getCIParams(args)

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Get CI-aware runner
	runner := getTestRunner(ciParams, config)
	defer generateCIReport(runner)

	// Display header with build tag information
	discoveredTags := displayTestHeader("race", config)

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

	// Use new build tag discovery if enabled (with race detector)
	return runTestsWithBuildTagDiscoveryTagsWithRunner(config, modules, true, remainingArgs, "race", discoveredTags, runner)
}

// Cover runs tests with coverage
func (Test) Cover(args ...string) error {
	// Extract CI params from args
	ciParams, remainingArgs := getCIParams(args)

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Get CI-aware runner
	runner := getTestRunner(ciParams, config)
	defer generateCIReport(runner)

	// Display header with build tag information
	discoveredTags := displayTestHeader("coverage", config)

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

	// Use build tag discovery for coverage tests
	return runCoverageTestsWithBuildTagDiscoveryTagsWithRunner(config, modules, false, remainingArgs, discoveredTags, runner)
}

// CoverRace runs tests with both coverage and race detector
func (Test) CoverRace(args ...string) error {
	// Extract CI params from args
	ciParams, remainingArgs := getCIParams(args)

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Get CI-aware runner
	runner := getTestRunner(ciParams, config)
	defer generateCIReport(runner)

	// Display header with build tag information
	discoveredTags := displayTestHeader("coverage+race", config)

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

	// Use build tag discovery for coverage+race tests
	return runCoverageTestsWithBuildTagDiscoveryTagsWithRunner(config, modules, true, remainingArgs, discoveredTags, runner)
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
				if strings.HasPrefix(relativePath, ".github/test-module/") {
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

// Fuzz runs fuzz tests with configurable time
func (Test) Fuzz(argsList ...string) error {
	// Parse command-line parameters
	params := utils.ParseParams(argsList)

	// Extract CI params from arguments
	ciParams, _ := getCIParams(argsList)

	config, err := GetConfig()
	if err != nil {
		return err
	}

	displayTestHeader("fuzz", config)

	// Print CI banner if CI mode is enabled (fuzz tests can't use full CI runner due to non-JSON output)
	PrintCIBannerIfEnabled(ciParams, config)

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

	// Get fuzz time from parameter or use default (10s)
	fuzzTimeStr := utils.GetParam(params, "time", "10s")
	fuzzTime, err := time.ParseDuration(fuzzTimeStr)
	if err != nil {
		return fmt.Errorf("invalid time parameter: %w", err)
	}

	// Run all fuzz tests and collect results (continues on failure)
	// In CI mode, capture output for text parsing to extract detailed failure info
	ciEnabled := IsCIEnabled(ciParams, config)
	results, totalDuration := runFuzzTestsWithResultsCI(config, fuzzTime, packages, ciEnabled)

	// Print CI summary if enabled
	printCIFuzzSummaryIfEnabled(ciParams, config, results, totalDuration)

	// Write CI results to JSONL format if enabled (standardized output for validation)
	writeFuzzCIResultsIfEnabled(ciParams, config, results, totalDuration)

	// Return aggregated error if any failures
	if err := fuzzResultsToError(results); err != nil {
		return err
	}

	utils.Success("Fuzz tests completed")
	return nil
}

// FuzzWithTime runs fuzz tests with specified duration
func (Test) FuzzWithTime(fuzzTime time.Duration) error {
	utils.Header("Running Fuzz Tests")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Print CI banner if CI mode is enabled (fuzz tests can't use full CI runner due to non-JSON output)
	PrintCIBannerIfEnabled(nil, config)

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

	// Run all fuzz tests and collect results (continues on failure)
	// In CI mode, capture output for text parsing to extract detailed failure info
	ciEnabled := IsCIEnabled(nil, config)
	results, totalDuration := runFuzzTestsWithResultsCI(config, fuzzTime, packages, ciEnabled)

	// Print CI summary if enabled (nil params relies on config/environment detection)
	printCIFuzzSummaryIfEnabled(nil, config, results, totalDuration)

	// Write CI results to JSONL format if enabled (standardized output for validation)
	writeFuzzCIResultsIfEnabled(nil, config, results, totalDuration)

	// Return aggregated error if any failures
	if err := fuzzResultsToError(results); err != nil {
		return err
	}

	utils.Success("Fuzz tests completed")
	return nil
}

// FuzzShort runs fuzz tests with configurable time (default: 5s for quick feedback)
func (Test) FuzzShort(argsList ...string) error {
	// Parse command-line parameters
	params := utils.ParseParams(argsList)

	// Extract CI params from arguments
	ciParams, _ := getCIParams(argsList)

	config, err := GetConfig()
	if err != nil {
		return err
	}

	displayTestHeader("fuzz-short", config)

	// Print CI banner if CI mode is enabled (fuzz tests can't use full CI runner due to non-JSON output)
	PrintCIBannerIfEnabled(ciParams, config)

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

	// Get fuzz time from parameter or use default (5s for short)
	fuzzTimeStr := utils.GetParam(params, "time", "5s")
	fuzzTime, err := time.ParseDuration(fuzzTimeStr)
	if err != nil {
		return fmt.Errorf("invalid time parameter: %w", err)
	}

	// Run all fuzz tests and collect results (continues on failure)
	// In CI mode, capture output for text parsing to extract detailed failure info
	ciEnabled := IsCIEnabled(ciParams, config)
	results, totalDuration := runFuzzTestsWithResultsCI(config, fuzzTime, packages, ciEnabled)

	// Print CI summary if enabled
	printCIFuzzSummaryIfEnabled(ciParams, config, results, totalDuration)

	// Write CI results to JSONL format if enabled (standardized output for validation)
	writeFuzzCIResultsIfEnabled(ciParams, config, results, totalDuration)

	// Return aggregated error if any failures
	if err := fuzzResultsToError(results); err != nil {
		return err
	}

	utils.Success("Short fuzz tests completed")
	return nil
}

// FuzzShortWithTime runs fuzz tests with specified duration (optimized for quick feedback)
func (Test) FuzzShortWithTime(fuzzTime time.Duration) error {
	utils.Header("Running Short Fuzz Tests")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Print CI banner if CI mode is enabled (fuzz tests can't use full CI runner due to non-JSON output)
	PrintCIBannerIfEnabled(nil, config)

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

	// Run all fuzz tests and collect results (continues on failure)
	// In CI mode, capture output for text parsing to extract detailed failure info
	ciEnabled := IsCIEnabled(nil, config)
	results, totalDuration := runFuzzTestsWithResultsCI(config, fuzzTime, packages, ciEnabled)

	// Print CI summary if enabled (nil params relies on config/environment detection)
	printCIFuzzSummaryIfEnabled(nil, config, results, totalDuration)

	// Write CI results to JSONL format if enabled (standardized output for validation)
	writeFuzzCIResultsIfEnabled(nil, config, results, totalDuration)

	// Return aggregated error if any failures
	if err := fuzzResultsToError(results); err != nil {
		return err
	}

	utils.Success("Short fuzz tests completed")
	return nil
}

// Bench runs benchmarks
func (Test) Bench(argsList ...string) error {
	// Parse command-line parameters
	params := utils.ParseParams(argsList)

	config, err := GetConfig()
	if err != nil {
		return err
	}

	displayTestHeader("benchmark", config)

	// Discover all modules
	modules, err := findAllModules()
	if err != nil {
		return fmt.Errorf("failed to find modules: %w", err)
	}

	if len(modules) == 0 {
		utils.Warn("No Go modules found")
		return nil
	}

	// Filter modules based on exclusion configuration
	benchModules := filterModulesForProcessing(modules, config, "benchmarks")
	if len(benchModules) == 0 {
		utils.Warn("No modules to benchmark after exclusions")
		return nil
	}

	// Build base benchmark args
	args := []string{"test", "-bench=.", "-benchmem", "-run=^$"}

	if config.Test.Verbose {
		args = append(args, "-v")
	}

	if config.Test.Tags != "" {
		args = append(args, "-tags", config.Test.Tags)
	}

	// Get benchmark time from parameter or use default
	benchTime := utils.GetParam(params, "time", "10s")
	args = append(args, "-benchtime", benchTime)

	// Get count from parameter
	count := utils.GetParam(params, "count", "")
	if count != "" {
		args = append(args, "-count", count)
	}

	args = append(args, "./...")

	totalStart := time.Now()
	var moduleErrors []moduleError

	// Run benchmarks for each module
	for _, module := range benchModules {
		displayModuleHeader(module, "Running benchmarks for")

		moduleStart := time.Now()
		err := runCommandInModule(module, "go", args...)
		if err != nil {
			moduleErrors = append(moduleErrors, moduleError{Module: module, Error: err})
			utils.Error("Benchmarks failed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		} else {
			utils.Success("Benchmarks passed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		}
	}

	// Report overall results
	if len(moduleErrors) > 0 {
		utils.Error("Benchmarks failed in %d/%d modules", len(moduleErrors), len(benchModules))
		return formatModuleErrors(moduleErrors)
	}

	utils.Success("All benchmarks completed in %s", utils.FormatDuration(time.Since(totalStart)))
	return nil
}

// BenchShort runs benchmarks with shorter duration for quick feedback
func (Test) BenchShort(argsList ...string) error {
	// Parse command-line parameters
	params := utils.ParseParams(argsList)

	config, err := GetConfig()
	if err != nil {
		return err
	}

	displayTestHeader("benchmark-short", config)

	// Discover all modules
	modules, err := findAllModules()
	if err != nil {
		return fmt.Errorf("failed to find modules: %w", err)
	}

	if len(modules) == 0 {
		utils.Warn("No Go modules found")
		return nil
	}

	// Filter modules based on exclusion configuration
	benchModules := filterModulesForProcessing(modules, config, "benchmarks")
	if len(benchModules) == 0 {
		utils.Warn("No modules to benchmark after exclusions")
		return nil
	}

	// Build base benchmark args
	args := []string{"test", "-bench=.", "-benchmem", "-run=^$"}

	if config.Test.Verbose {
		args = append(args, "-v")
	}

	if config.Test.Tags != "" {
		args = append(args, "-tags", config.Test.Tags)
	}

	// Get benchmark time from parameter or use default (1s for short)
	benchTime := utils.GetParam(params, "time", "1s")
	args = append(args, "-benchtime", benchTime)

	// Get count from parameter
	count := utils.GetParam(params, "count", "")
	if count != "" {
		args = append(args, "-count", count)
	}

	args = append(args, "./...")

	totalStart := time.Now()
	var moduleErrors []moduleError

	// Run benchmarks for each module
	for _, module := range benchModules {
		displayModuleHeader(module, "Running short benchmarks for")

		moduleStart := time.Now()
		err := runCommandInModule(module, "go", args...)
		if err != nil {
			moduleErrors = append(moduleErrors, moduleError{Module: module, Error: err})
			utils.Error("Short benchmarks failed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		} else {
			utils.Success("Short benchmarks passed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		}
	}

	// Report overall results
	if len(moduleErrors) > 0 {
		utils.Error("Short benchmarks failed in %d/%d modules", len(moduleErrors), len(benchModules))
		return formatModuleErrors(moduleErrors)
	}

	utils.Success("All short benchmarks completed in %s", utils.FormatDuration(time.Since(totalStart)))
	return nil
}

// Integration runs integration tests
func (Test) Integration() error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	displayTestHeader("integration", config)

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

// displayTestHeader displays test header with build tag information if auto-discovery is enabled
// formatBool returns a checkmark or X for boolean values
func formatBool(value bool) string {
	if value {
		return "✓"
	}
	return "✗"
}

// getEnvWithDefault gets an environment variable with a fallback value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// formatDuration formats timeout values nicely
func formatDuration(duration string) string {
	if duration == "" {
		return "default"
	}
	return duration
}

// truncateList truncates a string slice to a maximum length with ellipsis
func truncateList(items []string, maxLength int) string {
	if len(items) == 0 {
		return "none"
	}

	joined := strings.Join(items, ", ")
	if len(joined) <= maxLength {
		return joined
	}

	// Find how many items we can fit
	result := ""
	for _, item := range items {
		candidate := result
		if candidate != "" {
			candidate += ", "
		}
		candidate += item

		if len(candidate)+4 > maxLength { // +4 for " ..."
			if result == "" {
				// Even first item is too long, truncate it
				return item[:maxLength-3] + "..."
			}
			return result + " ..."
		}
		result = candidate
	}
	return result
}

func displayTestHeader(testType string, config *Config) []string {
	// Discover build tags early for display
	var discoveredTags []string
	var buildTagsInfo string

	if config.Test.AutoDiscoverBuildTags {
		discoveredTags, buildTagsInfo = processBuildTagAutoDiscovery(config)
	} else {
		buildTagsInfo = fmt.Sprintf("Auto-Discovery: %s • Manual Tags: %s", formatBool(false),
			getEnvWithDefault("MAGE_X_BUILD_TAGS", "none"))
	}

	// Get module information
	modules, err := findAllModules()
	var moduleList string
	if err != nil {
		moduleList = "Error loading modules"
	} else {
		moduleNames := make([]string, len(modules))
		for i, mod := range modules {
			if mod.IsRoot {
				moduleNames[i] = fmt.Sprintf("%s (main)", mod.Name)
			} else {
				moduleNames[i] = mod.Name
			}
		}
		moduleList = truncateList(moduleNames, 80)
	}

	// Get timeout from config only
	timeout := formatDuration(config.Test.Timeout)

	// Calculate effective values based on test type
	effectiveRace := config.Test.Race
	effectiveCover := config.Test.Cover
	effectiveShort := config.Test.Short

	switch testType {
	case "race":
		effectiveRace = true // Race tests always enable race detection
	case "coverage":
		effectiveCover = true // Coverage tests always enable coverage
	case "coverage+race":
		effectiveRace = true // Coverage+race enables both
		effectiveCover = true
	case "short":
		effectiveShort = true // Short tests enable short mode
	}

	// Coverage information from config
	coverMode := config.Test.CoverMode
	if coverMode == "" {
		coverMode = "atomic"
	}

	// Display clean header using utils.Print to avoid [INFO] prefixes
	utils.Println(fmt.Sprintf("Running %s Tests", titleCase(testType)))
	utils.Println(strings.Repeat("━", 60))

	utils.Println("\nTest Configuration:")
	utils.Print("  Timeout: %s • Race: %s • Cover: %s • Verbose: %s • Parallel: %d\n",
		timeout, formatBool(effectiveRace), formatBool(effectiveCover),
		formatBool(config.Test.Verbose), config.Test.Parallel)
	utils.Print("  Coverage Mode: %s • Shuffle: %s • Short: %s\n",
		coverMode, formatBool(config.Test.Shuffle), formatBool(effectiveShort))

	utils.Println("\nBuild Tags:")
	utils.Print("  %s\n", buildTagsInfo)

	utils.Println("\nModules:")
	if len(modules) == 0 {
		utils.Println("  No modules found")
	} else {
		utils.Print("  %d found: %s\n", len(modules), moduleList)
	}

	utils.Println(strings.Repeat("━", 60))

	return discoveredTags
}

// processBuildTagAutoDiscovery handles build tag auto-discovery logic
func processBuildTagAutoDiscovery(config *Config) ([]string, string) {
	tags, err := utils.DiscoverBuildTagsFromCurrentDir(config.Test.AutoDiscoverBuildTagsExclude)
	if err != nil {
		return nil, fmt.Sprintf("Auto-Discovery: %s | Error: %v", formatBool(true), err)
	}

	if len(tags) == 0 {
		return nil, fmt.Sprintf("Auto-Discovery: %s | Found: none", formatBool(true))
	}

	excluded := truncateList(config.Test.AutoDiscoverBuildTagsExclude, 40)
	testTargets := []string{"base"}
	testTargets = append(testTargets, tags...)

	// Show count and full list of discovered tags (or reasonable truncation)
	var foundDisplay string
	if len(tags) <= 10 {
		// Show all tags if there are 10 or fewer
		foundDisplay = fmt.Sprintf("%d tags: [%s]", len(tags), strings.Join(tags, ", "))
	} else {
		// Show count and truncated list for larger sets
		foundDisplay = fmt.Sprintf("%d tags: [%s]", len(tags), truncateList(tags, 80))
	}

	buildTagsInfo := fmt.Sprintf("Auto-Discovery: %s | Found: %s", formatBool(true), foundDisplay)

	if len(config.Test.AutoDiscoverBuildTagsExclude) > 0 {
		buildTagsInfo += fmt.Sprintf("\n  Excluded: [%s]", excluded)
		buildTagsInfo += fmt.Sprintf("\n  Testing: [%s]", truncateList(testTargets, 80))
	} else {
		buildTagsInfo += fmt.Sprintf("\n  Testing: [%s]", truncateList(testTargets, 80))
	}

	return tags, buildTagsInfo
}

// titleCase converts the first character to uppercase (replacement for deprecated strings.Title)
func titleCase(s string) string {
	if s == "" {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// getTagInfo returns formatted tag information string
func getTagInfo(buildTag string) string {
	if buildTag != "" {
		return fmt.Sprintf(" with tag '%s'", buildTag)
	}
	return ""
}

// handleCoverageFileMove moves coverage file to root directory and updates coverage files list
func handleCoverageFileMove(module ModuleInfo, coverFile string, coverageFiles *[]string) {
	srcPath := filepath.Join(module.Path, coverFile)
	dstPath := coverFile
	if utils.FileExists(srcPath) {
		if err := os.Rename(srcPath, dstPath); err != nil {
			utils.Warn("Failed to move coverage file for %s: %v", module.Relative, err)
		} else {
			*coverageFiles = append(*coverageFiles, coverFile)
		}
	}
}

// handleCoverageFilesWithBuildTag handles coverage files based on build tag
func handleCoverageFilesWithBuildTag(coverageFiles []string, buildTag string) {
	if buildTag == "" {
		// For base coverage (no tags), use standard coverage.txt
		handleCoverageFiles(coverageFiles)
		return
	}

	// For build tag coverage, create a separate merged coverage file
	taggedCoverageFile := fmt.Sprintf("coverage_%s.txt", buildTag)
	handleTaggedCoverageFiles(coverageFiles, taggedCoverageFile, buildTag)
}

// handleTaggedCoverageFiles processes coverage files for a specific build tag
func handleTaggedCoverageFiles(coverageFiles []string, taggedCoverageFile, buildTag string) {
	switch len(coverageFiles) {
	case 0:
		return
	case 1:
		if err := os.Rename(coverageFiles[0], taggedCoverageFile); err != nil {
			utils.Warn("Failed to rename coverage file: %v", err)
		}
	default:
		mergeAndCleanupCoverageFiles(coverageFiles, taggedCoverageFile, buildTag)
	}
}

// mergeAndCleanupCoverageFiles merges multiple coverage files and cleans up
func mergeAndCleanupCoverageFiles(coverageFiles []string, taggedCoverageFile, buildTag string) {
	if err := mergeCoverageFiles(coverageFiles, taggedCoverageFile); err != nil {
		utils.Warn("Failed to merge coverage files for tag '%s': %v", buildTag, err)
		return
	}
	// Clean up individual coverage files
	for _, file := range coverageFiles {
		if err := os.Remove(file); err != nil {
			utils.Warn("Failed to remove coverage file %s: %v", file, err)
		}
	}
}

// runTestsWithBuildTagDiscoveryTags runs tests with pre-discovered build tags
//
//nolint:unparam // testType kept for API consistency with WithRunner variant, even though currently only called with "unit"
func runTestsWithBuildTagDiscoveryTags(config *Config, modules []ModuleInfo, additionalArgs []string, testType string, discoveredTags []string) error {
	return runTestsWithBuildTagDiscoveryTagsWithRunner(config, modules, false, additionalArgs, testType, discoveredTags, GetRunner())
}

// runTestsWithBuildTagDiscoveryTagsWithRunner runs tests with pre-discovered build tags using provided runner
func runTestsWithBuildTagDiscoveryTagsWithRunner(config *Config, modules []ModuleInfo, race bool, additionalArgs []string, testType string, discoveredTags []string, runner CommandRunner) error {
	if !config.Test.AutoDiscoverBuildTags || len(discoveredTags) == 0 {
		// Run tests normally without build tag discovery
		return runTestsForModulesWithRunner(config, modules, race, false, additionalArgs, testType, "", runner)
	}

	// Run base tests (no build tags)
	utils.Info("Running %s tests without build tags", testType)
	if err := runTestsForModulesWithRunner(config, modules, race, false, additionalArgs, testType, "", runner); err != nil {
		return fmt.Errorf("base tests failed: %w", err)
	}

	// Run tests for each discovered build tag
	for _, tag := range discoveredTags {
		utils.Info("Running %s tests with build tag: %s", testType, tag)
		if err := runTestsForModulesWithRunner(config, modules, race, false, additionalArgs, testType, tag, runner); err != nil {
			return fmt.Errorf("tests with tag '%s' failed: %w", tag, err)
		}
	}

	return nil
}

// runCoverageTestsWithBuildTagDiscoveryTags runs coverage tests with pre-discovered build tags
func runCoverageTestsWithBuildTagDiscoveryTags(config *Config, modules []ModuleInfo, race bool, additionalArgs, discoveredTags []string) error {
	return runCoverageTestsWithBuildTagDiscoveryTagsWithRunner(config, modules, race, additionalArgs, discoveredTags, GetRunner())
}

// runCoverageTestsWithBuildTagDiscoveryTagsWithRunner runs coverage tests with pre-discovered build tags using provided runner
func runCoverageTestsWithBuildTagDiscoveryTagsWithRunner(config *Config, modules []ModuleInfo, race bool, additionalArgs, discoveredTags []string, runner CommandRunner) error {
	if !config.Test.AutoDiscoverBuildTags || len(discoveredTags) == 0 {
		// Run coverage tests normally without build tag discovery
		return runCoverageTestsForModulesWithRunner(config, modules, race, additionalArgs, "", runner)
	}

	// Run base coverage tests (no build tags)
	utils.Info("Running coverage tests without build tags")
	if err := runCoverageTestsForModulesWithRunner(config, modules, race, additionalArgs, "", runner); err != nil {
		return fmt.Errorf("base coverage tests failed: %w", err)
	}

	// Run coverage tests for each discovered build tag
	for _, tag := range discoveredTags {
		utils.Info("Running coverage tests with build tag: %s", tag)
		if err := runCoverageTestsForModulesWithRunner(config, modules, race, additionalArgs, tag, runner); err != nil {
			return fmt.Errorf("coverage tests with tag '%s' failed: %w", tag, err)
		}
	}

	// Show final coverage summary
	if utils.FileExists("coverage.txt") {
		return Test{}.CoverReport()
	}
	return nil
}

// runCoverageTestsForModulesWithRunner runs coverage tests for all modules with the specified build tag using provided runner
func runCoverageTestsForModulesWithRunner(config *Config, modules []ModuleInfo, race bool, additionalArgs []string, buildTag string, runner CommandRunner) error {
	totalStart := time.Now()
	var moduleErrors []moduleError
	var coverageFiles []string

	tagSuffix := ""
	if buildTag != "" {
		tagSuffix = fmt.Sprintf("_%s", buildTag)
	}

	// Filter modules based on exclusion configuration
	filteredModules := filterModulesForProcessing(modules, config, "coverage tests")
	if len(filteredModules) == 0 {
		utils.Warn("No modules to test after exclusions")
		return nil
	}

	// Run tests for each module
	for i, module := range filteredModules {
		tagInfo := ""
		if buildTag != "" {
			tagInfo = fmt.Sprintf(" (tag: %s)", buildTag)
		}
		displayModuleHeader(module, fmt.Sprintf("Running coverage tests%s", tagInfo))

		moduleStart := time.Now()

		// Create unique coverage file name for each module
		coverFile := fmt.Sprintf("coverage_%d%s.txt", i, tagSuffix)
		if module.Relative != "." {
			// Use sanitized path for coverage file name
			sanitized := strings.ReplaceAll(module.Relative, "/", "_")
			coverFile = fmt.Sprintf("coverage_%s%s.txt", sanitized, tagSuffix)
		}

		// Build test args with build tag override if specified
		var testArgs []string
		if buildTag != "" {
			// Override config tags with discovered build tag
			tempConfig := *config
			tempConfig.Test.Tags = buildTag
			testArgs = buildTestArgs(&tempConfig, race, true, additionalArgs...)
		} else {
			testArgs = buildTestArgs(config, race, true, additionalArgs...)
		}

		testArgs = append(testArgs, "-coverprofile="+coverFile)
		if race {
			testArgs = append(testArgs, "-covermode=atomic") // atomic is required with race
		} else {
			testArgs = append(testArgs, "-covermode="+config.Test.CoverMode)
		}

		if len(config.Test.CoverPkg) > 0 {
			testArgs = append(testArgs, "-coverpkg="+strings.Join(config.Test.CoverPkg, ","))
		}

		testArgs = append(testArgs, "./...")

		// Run tests in module directory using provided runner
		err := runCommandInModuleWithRunner(module, runner, "go", testArgs...)

		if err != nil {
			moduleErrors = append(moduleErrors, moduleError{Module: module, Error: err})
			tagInfo := getTagInfo(buildTag)
			utils.Error("Coverage tests failed for %s%s in %s", module.Relative, tagInfo, utils.FormatDuration(time.Since(moduleStart)))
		} else {
			handleCoverageFileMove(module, coverFile, &coverageFiles)
			tagInfo := getTagInfo(buildTag)
			utils.Success("Coverage tests passed for %s%s in %s", module.Relative, tagInfo, utils.FormatDuration(time.Since(moduleStart)))
		}
	}

	// Handle coverage files
	handleCoverageFilesWithBuildTag(coverageFiles, buildTag)

	// Report overall results
	if len(moduleErrors) > 0 {
		tagInfo := getTagInfo(buildTag)
		utils.Error("Coverage tests%s failed in %d/%d modules", tagInfo, len(moduleErrors), len(filteredModules))
		return formatModuleErrors(moduleErrors)
	}

	tagInfo := getTagInfo(buildTag)
	utils.Success("All coverage tests%s passed in %s", tagInfo, utils.FormatDuration(time.Since(totalStart)))
	return nil
}

// runTestsForModulesWithRunner runs tests for all modules with the specified configuration using provided runner
func runTestsForModulesWithRunner(config *Config, modules []ModuleInfo, race, cover bool, additionalArgs []string, testType, buildTag string, runner CommandRunner) error {
	var moduleErrors []moduleError
	totalStart := time.Now()

	// Filter modules based on exclusion configuration
	filteredModules := filterModulesForProcessing(modules, config, testType+" tests")
	if len(filteredModules) == 0 {
		utils.Warn("No modules to test after exclusions")
		return nil
	}

	for _, module := range filteredModules {
		tagSuffix := ""
		if buildTag != "" {
			tagSuffix = fmt.Sprintf(" (tag: %s)", buildTag)
		}
		displayModuleHeader(module, fmt.Sprintf("Running %s tests%s", testType, tagSuffix))

		moduleStart := time.Now()

		// Build test args with build tag override if specified
		var testArgs []string
		if buildTag != "" {
			// Override config tags with discovered build tag
			tempConfig := *config
			tempConfig.Test.Tags = buildTag
			testArgs = buildTestArgs(&tempConfig, race, cover, additionalArgs...)
		} else {
			testArgs = buildTestArgs(config, race, cover, additionalArgs...)
		}

		if testType == "unit" || testType == "short" {
			testArgs = append(testArgs, "-short")
		}
		testArgs = append(testArgs, "./...")

		// Run tests in module directory using provided runner
		err := runCommandInModuleWithRunner(module, runner, "go", testArgs...)

		if err != nil {
			tagInfo := getTagInfo(buildTag)
			moduleErrors = append(moduleErrors, moduleError{Module: module, Error: err})
			utils.Error("%s tests failed for %s%s in %s", titleCase(testType), module.Relative, tagInfo, utils.FormatDuration(time.Since(moduleStart)))
		} else {
			tagInfo := getTagInfo(buildTag)
			utils.Success("%s tests passed for %s%s in %s", titleCase(testType), module.Relative, tagInfo, utils.FormatDuration(time.Since(moduleStart)))
		}
	}

	// Report overall results
	if len(moduleErrors) > 0 {
		tagInfo := getTagInfo(buildTag)
		utils.Error("%s tests%s failed in %d/%d modules", titleCase(testType), tagInfo, len(moduleErrors), len(filteredModules))
		return formatModuleErrors(moduleErrors)
	}

	tagInfo := getTagInfo(buildTag)
	utils.Success("All %s tests%s passed in %s", testType, tagInfo, utils.FormatDuration(time.Since(totalStart)))
	return nil
}

// parseSafeTestArgs parses and validates safe test arguments
func parseSafeTestArgs(args []string) ([]string, error) {
	var safeArgs []string

	// Define allowed flags for security
	allowedFlags := map[string]bool{
		"-json":      true,
		"-v":         true,
		"-count":     true,
		"-cpu":       true,
		"-parallel":  true,
		"-shuffle":   true,
		"-failfast":  true,
		"-vet":       true,
		"-run":       true,
		"-bench":     true,
		"-benchmem":  true,
		"-benchtime": true,
		"-short":     true,
		"-timeout":   true,
		"-race":      true,
		"-cover":     true,
		"-covermode": true,
		"-coverpkg":  true,
		"-tags":      true,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Handle flags with values (like -count=5 or -count 5)
		if strings.Contains(arg, "=") {
			parts := strings.SplitN(arg, "=", 2)
			flagName := parts[0]
			if !allowedFlags[flagName] {
				return nil, fmt.Errorf("%w: %s", errFlagNotAllowed, flagName)
			}
			safeArgs = append(safeArgs, arg)
		} else if allowedFlags[arg] {
			safeArgs = append(safeArgs, arg)

			// Check if this flag expects a value (next argument)
			flagsWithValues := map[string]bool{
				"-count": true, "-cpu": true, "-parallel": true, "-vet": true,
				"-run": true, "-bench": true, "-benchtime": true, "-timeout": true,
				"-covermode": true, "-coverpkg": true, "-tags": true,
			}

			if flagsWithValues[arg] && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++ // Skip the next argument (it's the value)
				safeArgs = append(safeArgs, args[i])
			}
		} else if strings.HasPrefix(arg, "-") {
			return nil, fmt.Errorf("%w: %s", errFlagNotAllowed, arg)
		} else {
			// Non-flag arguments are allowed (like package paths)
			safeArgs = append(safeArgs, arg)
		}
	}

	return safeArgs, nil
}

// buildTestArgs builds common test arguments with optional additional args
func buildTestArgs(cfg *Config, race, cover bool, additionalArgs ...string) []string {
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

	// Add additional arguments if provided
	if len(additionalArgs) > 0 {
		safeArgs, err := parseSafeTestArgs(additionalArgs)
		if err != nil {
			// Log warning but continue - better to run tests without unsafe args than fail
			fmt.Printf("Warning: %v. Ignoring unsafe arguments.\n", err)
		} else {
			// Check if user provided timeout in additional args to avoid duplicates
			userHasTimeout := hasTimeoutFlag(safeArgs)
			if userHasTimeout {
				// Remove config timeout if user provided one
				args = removeTimeoutFlag(args)
			}
			args = append(args, safeArgs...)
		}
	}

	return args
}

// buildTestArgsWithOverrides builds test arguments with explicit overrides for race and cover
// When raceOverride or coverOverride are not nil, they take precedence over config defaults
func buildTestArgsWithOverrides(cfg *Config, raceOverride, coverOverride *bool, additionalArgs ...string) []string {
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

	// Add additional arguments if provided
	if len(additionalArgs) > 0 {
		safeArgs, err := parseSafeTestArgs(additionalArgs)
		if err != nil {
			// Log warning but continue - better to run tests without unsafe args than fail
			fmt.Printf("Warning: %v. Ignoring unsafe arguments.\n", err)
		} else {
			// Check if user provided timeout in additional args to avoid duplicates
			userHasTimeout := hasTimeoutFlag(safeArgs)
			if userHasTimeout {
				// Remove config timeout if user provided one
				args = removeTimeoutFlag(args)
			}
			args = append(args, safeArgs...)
		}
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

// runFuzzTestsWithResultsCI runs fuzz tests with optional CI output capture.
// When ciEnabled is true, stdout is captured for text parsing to extract detailed failure info.
func runFuzzTestsWithResultsCI(config *Config, fuzzTime time.Duration, packages []string, ciEnabled bool) ([]fuzzTestResult, time.Duration) {
	startTime := time.Now()
	var results []fuzzTestResult

	fuzzTimeStr := fuzzTime.String()

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

			testStart := time.Now()
			args := []string{"test", "-run=^$", fmt.Sprintf("-fuzz=^%s$", test)}
			args = append(args, "-fuzztime", fuzzTimeStr)

			if config.Test.Verbose {
				args = append(args, "-v")
			}
			args = append(args, pkg)

			var testErr error
			var testOutput string

			if ciEnabled {
				// Capture output for CI text parsing
				testOutput, testErr = runFuzzTestWithOutput("go", args...)
			} else {
				// Standard execution without output capture
				testErr = GetRunner().RunCmd("go", args...)
			}
			testDuration := time.Since(testStart)

			results = append(results, fuzzTestResult{
				Package:  pkg,
				Test:     test,
				Duration: testDuration,
				Error:    testErr,
				Output:   testOutput,
			})
		}
	}

	return results, time.Since(startTime)
}

// runFuzzTestWithOutput executes a fuzz test command and captures stdout/stderr.
// This is used in CI mode to capture output for text parsing.
func runFuzzTestWithOutput(name string, args ...string) (string, error) {
	ctx := context.Background()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = os.Environ()

	// Capture both stdout and stderr
	var outputBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &outputBuf) // Tee to stdout for live display
	cmd.Stderr = io.MultiWriter(os.Stderr, &outputBuf) // Capture stderr too

	err := cmd.Run()
	return outputBuf.String(), err
}

// printCIFuzzSummaryIfEnabled prints the fuzz summary if CI mode is enabled
func printCIFuzzSummaryIfEnabled(params map[string]string, config *Config, results []fuzzTestResult, totalDuration time.Duration) {
	if IsCIEnabled(params, config) {
		printCIFuzzSummary(results, totalDuration)
	}
}

// writeFuzzCIResultsIfEnabled writes fuzz test results to JSONL format when CI mode is enabled.
// This produces output in the same format as regular tests, allowing the validation workflow
// to process fuzz results uniformly with unit test results.
//
// When fuzz test output is captured, it uses TextStreamParser to extract detailed failure info
// including file:line locations, error messages, and fuzz corpus paths.
func writeFuzzCIResultsIfEnabled(params map[string]string, config *Config, results []fuzzTestResult, totalDuration time.Duration) {
	if !IsCIEnabled(params, config) {
		return
	}

	detector := NewCIDetector()
	mode := detector.GetConfig(params, config)

	// Use a separate file for fuzz results to avoid overwriting regular test results
	// The path is based on mode.OutputPath but with "-fuzz" suffix
	fuzzOutputPath := strings.TrimSuffix(mode.OutputPath, ".jsonl") + "-fuzz.jsonl"

	reporter, err := NewJSONReporter(fuzzOutputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to create fuzz CI reporter: %v\n", err)
		return
	}
	defer func() {
		if closeErr := reporter.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to close fuzz CI reporter: %v\n", closeErr)
		}
	}()

	// Also add GitHub reporter if running in GitHub Actions
	var reporters []CIReporter
	reporters = append(reporters, reporter)
	if mode.Format == CIFormatGitHub || (mode.Format == CIFormatAuto && detector.Platform() == string(CIFormatGitHub)) {
		reporters = append(reporters, NewGitHubReporter())
	}
	multiReporter := NewMultiReporter(reporters...)

	// Write metadata
	if d, ok := detector.(*ciDetector); ok {
		if startErr := multiReporter.Start(d.GetMetadata()); startErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to write fuzz CI start: %v\n", startErr)
		}
	}

	// Parse fuzz output with text parser to extract detailed failure info
	failures := parseFuzzResultsWithTextParser(results, mode.ContextLines, mode.Dedup)

	// Count results
	total, passed, failed := len(results), 0, 0
	for _, r := range results {
		if r.Error != nil {
			failed++
		} else {
			passed++
		}
	}

	// Report failures
	for _, failure := range failures {
		if reportErr := multiReporter.ReportFailure(failure); reportErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: Failed to report fuzz failure: %v\n", reportErr)
		}
	}

	// Write summary
	status := TestStatusPassed
	if failed > 0 {
		status = TestStatusFailed
	}

	result := &CIResult{
		Summary: CISummary{
			Status:   status,
			Total:    total,
			Passed:   passed,
			Failed:   len(failures), // Use deduplicated count from parser
			Skipped:  0,
			Duration: formatDurationForSummary(totalDuration),
		},
		Failures:  failures,
		Timestamp: time.Now(),
		Duration:  totalDuration,
	}

	if summaryErr := multiReporter.WriteSummary(result); summaryErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to write fuzz CI summary: %v\n", summaryErr)
	}

	if closeErr := multiReporter.Close(); closeErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to close fuzz CI multi reporter: %v\n", closeErr)
	}
}

// parseFuzzResultsWithTextParser extracts detailed failure information from fuzz test output
// using the TextStreamParser. This provides file:line locations, error messages, and fuzz corpus paths.
func parseFuzzResultsWithTextParser(results []fuzzTestResult, contextLines int, dedup bool) []CITestFailure {
	parser := NewTextStreamParser(contextLines, dedup)

	// Combine all output and parse
	for _, r := range results {
		if r.Output == "" {
			// No output captured, create basic failure from error
			if r.Error != nil {
				// Parse the basic failure line
				failLine := fmt.Sprintf("--- FAIL: %s (%.2fs)", r.Test, r.Duration.Seconds())
				_ = parser.ParseLine(failLine) //nolint:errcheck // Best effort parsing

				// Add package info
				pkgLine := fmt.Sprintf("FAIL %s %.3fs", r.Package, r.Duration.Seconds())
				_ = parser.ParseLine(pkgLine) //nolint:errcheck // Best effort parsing
			}
			continue
		}

		// Parse the captured output line by line
		lines := strings.Split(r.Output, "\n")
		for _, line := range lines {
			_ = parser.ParseLine(line) //nolint:errcheck // Best effort parsing
		}
	}

	return parser.Flush()
}

// fuzzResultsToError converts fuzz test results to an aggregated error
func fuzzResultsToError(results []fuzzTestResult) error {
	var failures []string
	for _, r := range results {
		if r.Error != nil {
			failures = append(failures, fmt.Sprintf("%s.%s", r.Package, r.Test))
		}
	}

	if len(failures) == 0 {
		return nil
	}

	if len(failures) == 1 {
		return fmt.Errorf("%w: %s", errFuzzTestFailed, failures[0])
	}
	return fmt.Errorf("%w: %d tests (%s)", errFuzzTestFailed, len(failures), strings.Join(failures, ", "))
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

// hasTimeoutFlag checks if a timeout flag exists in the arguments
func hasTimeoutFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-timeout" {
			return true
		}
		if strings.HasPrefix(arg, "-timeout=") {
			return true
		}
	}
	return false
}

// removeTimeoutFlag removes timeout flags from the arguments
func removeTimeoutFlag(args []string) []string {
	var result []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-timeout" {
			// Skip this flag and its value
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				i++ // Skip the value too
			}
		} else if strings.HasPrefix(arg, "-timeout=") {
			// Skip this flag with embedded value
			continue
		} else {
			result = append(result, arg)
		}
	}
	return result
}
