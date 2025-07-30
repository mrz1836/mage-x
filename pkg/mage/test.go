package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Test namespace for test-related tasks
type Test mg.Namespace

// Default runs the default test suite with linting
func (Test) Default() error {
	// Run lint first
	if err := (Lint{}).Default(); err != nil {
		return err
	}

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

	args := buildTestArgs(config, false, false)
	args = append(args, "./...")

	start := time.Now()
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	utils.Success("Tests passed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// Short runs short tests (excludes integration tests)
func (Test) Short() error {
	utils.Header("Running Short Tests")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	args := buildTestArgs(config, false, false)
	args = append(args, "-short")
	args = append(args, "./...")

	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("short tests failed: %w", err)
	}

	utils.Success("Short tests passed")
	return nil
}

// Race runs tests with race detector
func (Test) Race() error {
	utils.Header("Running Tests with Race Detector")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	args := buildTestArgs(config, true, false)
	args = append(args, "./...")

	start := time.Now()
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("race tests failed: %w", err)
	}

	utils.Success("Race tests passed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// Cover runs tests with coverage
func (Test) Cover() error {
	utils.Header("Running Tests with Coverage")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	args := buildTestArgs(config, false, true)
	args = append(args, "-coverprofile=coverage.txt")
	args = append(args, "-covermode="+config.Test.CoverMode)

	if len(config.Test.CoverPkg) > 0 {
		args = append(args, "-coverpkg="+strings.Join(config.Test.CoverPkg, ","))
	}

	args = append(args, "./...")

	start := time.Now()
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("coverage tests failed: %w", err)
	}

	utils.Success("Coverage tests passed in %s", utils.FormatDuration(time.Since(start)))

	// Show coverage summary
	return Test{}.CoverReport()
}

// CoverRace runs tests with both coverage and race detector
func (Test) CoverRace() error {
	utils.Header("Running Tests with Coverage and Race Detector")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	args := buildTestArgs(config, true, true)
	args = append(args, "-coverprofile=coverage.txt")
	args = append(args, "-covermode=atomic") // atomic is required with race

	if len(config.Test.CoverPkg) > 0 {
		args = append(args, "-coverpkg="+strings.Join(config.Test.CoverPkg, ","))
	}

	args = append(args, "./...")

	start := time.Now()
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("coverage race tests failed: %w", err)
	}

	utils.Success("Coverage race tests passed in %s", utils.FormatDuration(time.Since(start)))

	// Show coverage summary
	return Test{}.CoverReport()
}

// CoverReport shows coverage report
func (Test) CoverReport() error {
	if !utils.FileExists("coverage.txt") {
		utils.Warn("No coverage file found. Run 'mage test:cover' first.")
		return nil
	}

	utils.Info("Coverage Report:")
	return GetRunner().RunCmd("go", "tool", "cover", "-func=coverage.txt")
}

// CoverHTML generates HTML coverage report
func (Test) CoverHTML() error {
	if !utils.FileExists("coverage.txt") {
		return fmt.Errorf("no coverage file found. Run 'mage test:cover' first")
	}

	utils.Info("Generating HTML coverage report...")
	if err := GetRunner().RunCmd("go", "tool", "cover", "-html=coverage.txt", "-o=coverage.html"); err != nil {
		return err
	}

	utils.Success("Coverage report generated: coverage.html")

	// Try to open in browser
	if utils.CommandExists("open") {
		GetRunner().RunCmd("open", "coverage.html")
	} else if utils.CommandExists("xdg-open") {
		GetRunner().RunCmd("xdg-open", "coverage.html")
	}

	return nil
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
	packages, err := findFuzzPackages()
	if err != nil {
		return err
	}

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

// Bench runs benchmarks
func (Test) Bench(params ...string) error {
	utils.Header("Running Benchmarks")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	args := []string{"test", "-bench=.", "-benchmem", "-run=^$"}

	if config.Test.Verbose {
		args = append(args, "-v")
	}

	if len(config.Test.Tags) > 0 {
		args = append(args, "-tags", strings.Join(config.Test.Tags, ","))
	}

	benchTime := utils.GetEnv("BENCH_TIME", "10s")
	args = append(args, "-benchtime", benchTime)

	if count := utils.GetEnv("BENCH_COUNT", ""); count != "" {
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

// Integration runs integration tests
func (Test) Integration() error {
	utils.Header("Running Integration Tests")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Set integration test tag
	tags := append(config.Test.Tags, "integration")

	args := []string{"test"}
	args = append(args, "-tags", strings.Join(tags, ","))

	if config.Test.Verbose {
		args = append(args, "-v")
	}

	// Longer timeout for integration tests
	args = append(args, "-timeout", "30m")

	// Don't run short tests
	args = append(args, "-run", "Integration")

	args = append(args, "./...")

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
	os.Setenv("CI", "true")

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

	if len(config.Test.Tags) > 0 {
		args = append(args, "-tags", strings.Join(config.Test.Tags, ","))
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
	args = append(args, "-p", fmt.Sprintf("%d", parallel))

	args = append(args, "./...")

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
	os.Setenv("CI", "true")

	// Run tests with coverage only (no race)
	return Test{}.Cover()
}

// Helper functions

// buildTestArgs builds common test arguments
func buildTestArgs(cfg *Config, race, cover bool) []string {
	args := []string{"test"}

	if cfg.Test.Parallel {
		parallelCount := cfg.Build.Parallel
		if parallelCount <= 0 {
			parallelCount = runtime.NumCPU()
		}
		args = append(args, "-p", fmt.Sprintf("%d", parallelCount))
	}

	if cfg.Test.Verbose {
		args = append(args, "-v")
	}

	if cfg.Test.Timeout != "" {
		args = append(args, "-timeout", cfg.Test.Timeout)
	}

	if len(cfg.Test.Tags) > 0 {
		args = append(args, "-tags", strings.Join(cfg.Test.Tags, ","))
	}

	if race || cfg.Test.Race {
		args = append(args, "-race")
	}

	if cover || cfg.Test.Cover {
		args = append(args, "-cover")
	}

	return args
}

// findFuzzPackages finds packages containing fuzz tests
func findFuzzPackages() ([]string, error) {
	output, err := GetRunner().RunCmdOutput("grep", "-r", "-l", "^func Fuzz", "--include=*_test.go", ".")
	if err != nil {
		// grep returns error if no matches found
		return nil, nil
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

		module, _ := utils.GetModuleName()
		if module != "" && pkg != "" {
			packageMap[filepath.Join(module, pkg)] = true
		} else if module != "" {
			packageMap[module] = true
		}
	}

	var packages []string
	for pkg := range packageMap {
		packages = append(packages, pkg)
	}

	return packages, nil
}

// Additional methods for Test namespace required by tests

// Run runs tests
func (Test) Run() error {
	runner := GetRunner()
	return runner.RunCmd("go", "test", "./...")
}

// Coverage generates test coverage
func (Test) Coverage(args ...string) error {
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
