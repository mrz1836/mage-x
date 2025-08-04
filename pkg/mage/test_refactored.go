package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/mage/builders"
	"github.com/mrz1836/mage-x/pkg/mage/operations"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors to satisfy err113 linter
var (
	errNoCoverageFile         = errors.New("no coverage file found. Run 'mage test:cover' first")
	errNoCoverageFilesToMerge = errors.New("no coverage files to merge")
)

// Test namespace for test-related tasks
type Test mg.Namespace

// Default runs the default test suite (unit tests only)
func (Test) Default() error {
	return Test{}.Unit()
}

// Full runs the complete test suite with linting
func (Test) Full() error {
	ctx, err := operations.NewOperation("Running Full Test Suite (Lint + Tests)")
	if err != nil {
		return err
	}

	// Run lint first
	if err := (Lint{}).Default(); err != nil {
		return ctx.Complete(err)
	}

	// Then run unit tests
	return ctx.Complete(Test{}.Unit())
}

// Unit runs unit tests using the new common components
func (Test) Unit() error {
	ctx, err := operations.NewOperation("Running Unit Tests")
	if err != nil {
		return err
	}

	// Create test operation
	testOp := operations.NewTestOperation(ctx.Config(), builders.TestOptions{
		Short: true,
	})

	// Run tests on all modules
	runner := operations.NewModuleRunner(ctx.Config(), testOp)
	return ctx.Complete(runner.RunForAllModules())
}

// Cover runs tests with coverage analysis
func (Test) Cover() error {
	ctx, err := operations.NewOperation("Running Tests with Coverage")
	if err != nil {
		return err
	}

	// Check if we should open coverage in browser
	openCoverage := utils.GetEnv("OPEN_COVERAGE", "false") == "true"

	// Create coverage operation
	coverOp := operations.NewTestOperation(ctx.Config(), builders.TestOptions{
		Coverage:     true,
		CoverageFile: "coverage.out",
		CoverageMode: ctx.Config().Test.CoverageMode,
	})

	// Run coverage on all modules
	runner := operations.NewModuleRunner(ctx.Config(), coverOp)
	if err := runner.RunForAllModules(); err != nil {
		return ctx.Complete(err)
	}

	// Generate HTML report if requested
	if openCoverage {
		ctx.Info("Generating HTML coverage report...")
		if err := utils.RunCommand("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html"); err != nil {
			ctx.Warn("Failed to generate HTML coverage: %v", err)
		} else {
			ctx.Success("Coverage report saved to coverage.html")
			// Open in browser
			if err := utils.OpenInBrowser("coverage.html"); err != nil {
				ctx.Warn("Failed to open coverage in browser: %v", err)
			}
		}
	}

	return ctx.Complete(nil)
}

// Race runs tests with race detector
func (Test) Race() error {
	ctx, err := operations.NewOperation("Running Tests with Race Detector")
	if err != nil {
		return err
	}

	// Create race test operation
	raceOp := operations.NewTestOperation(ctx.Config(), builders.TestOptions{
		Race:  true,
		Short: true,
	})

	// Run tests on all modules
	runner := operations.NewModuleRunner(ctx.Config(), raceOp)
	return ctx.Complete(runner.RunForAllModules())
}

// Integration runs integration tests
func (Test) Integration() error {
	ctx, err := operations.NewOperation("Running Integration Tests")
	if err != nil {
		return err
	}

	// Create integration test operation
	integrationOp := operations.NewTestOperation(ctx.Config(), builders.TestOptions{
		Integration: true,
	})

	// Run tests on all modules
	runner := operations.NewModuleRunner(ctx.Config(), integrationOp)
	return ctx.Complete(runner.RunForAllModules())
}

// Bench runs benchmark tests
func (Test) Bench() error {
	ctx, err := operations.NewOperation("Running Benchmarks")
	if err != nil {
		return err
	}

	pattern := utils.GetEnv("BENCH_PATTERN", ".")

	// Create benchmark operation
	benchOp := &BenchmarkOperation{
		config:  ctx.Config(),
		pattern: pattern,
	}

	// Run benchmarks on all modules
	runner := operations.NewModuleRunner(ctx.Config(), benchOp)
	return ctx.Complete(runner.RunForAllModules())
}

// BenchmarkOperation implements ModuleOperation for benchmarks
type BenchmarkOperation struct {
	config  *Config
	pattern string
}

func (b *BenchmarkOperation) Name() string {
	return "Benchmarks"
}

func (b *BenchmarkOperation) Execute(module Module, config *Config) error {
	builder := builders.NewTestCommandBuilder(config)
	args := builder.BuildBenchmarkArgs(b.pattern)
	return RunCommandInModule(module, "go", args...)
}

// Short runs short tests (excludes integration tests)
func (Test) Short() error {
	ctx, err := operations.NewOperation("Running Short Tests")
	if err != nil {
		return err
	}

	// Create short test operation
	shortOp := operations.NewTestOperation(ctx.Config(), builders.TestOptions{
		Short: true,
	})

	// Run tests on all modules
	runner := operations.NewModuleRunner(ctx.Config(), shortOp)
	return ctx.Complete(runner.RunForAllModules())
}

// CoverRace runs tests with both coverage and race detector
func (Test) CoverRace() error {
	ctx, err := operations.NewOperation("Running Tests with Coverage and Race Detector")
	if err != nil {
		return err
	}

	// Create coverage + race operation
	coverRaceOp := operations.NewTestOperation(ctx.Config(), builders.TestOptions{
		Coverage:     true,
		Race:         true,
		CoverageFile: "coverage.out",
		CoverageMode: ctx.Config().Test.CoverageMode,
	})

	// Run tests on all modules
	runner := operations.NewModuleRunner(ctx.Config(), coverRaceOp)
	return ctx.Complete(runner.RunForAllModules())
}

// BenchShort runs benchmarks with shorter duration
func (Test) BenchShort() error {
	// Override bench time for quick feedback
	os.Setenv("BENCH_TIME", "1s")
	defer os.Unsetenv("BENCH_TIME")

	return Test{}.Bench()
}

// Fuzz runs fuzz tests
func (Test) Fuzz() error {
	ctx, err := operations.NewOperation("Running Fuzz Tests")
	if err != nil {
		return err
	}

	if ctx.Config().Test.SkipFuzz {
		ctx.Info("Fuzz tests are disabled in configuration")
		return ctx.CompleteWithoutTiming(nil)
	}

	// Create fuzz operation
	fuzzOp := &FuzzOperation{
		config:   ctx.Config(),
		duration: utils.GetEnv("FUZZ_TIME", "10s"),
	}

	// Run fuzz tests on all modules
	runner := operations.NewModuleRunner(ctx.Config(), fuzzOp)
	return ctx.Complete(runner.RunForAllModules())
}

// FuzzOperation implements ModuleOperation for fuzz tests
type FuzzOperation struct {
	config   *Config
	duration string
}

func (f *FuzzOperation) Name() string {
	return "Fuzz tests"
}

func (f *FuzzOperation) Execute(module Module, config *Config) error {
	args := []string{"test", "-fuzz=.", "-fuzztime", f.duration}

	if config.Build.Verbose {
		args = append(args, "-v")
	}

	// Add test tags if configured
	if config.Test.Tags != "" {
		args = append(args, "-tags", config.Test.Tags)
	}

	args = append(args, "./...")

	return RunCommandInModule(module, "go", args...)
}

// FuzzShort runs fuzz tests with shorter duration
func (Test) FuzzShort() error {
	// Override fuzz time for quick feedback
	os.Setenv("FUZZ_TIME", "1s")
	defer os.Unsetenv("FUZZ_TIME")

	return Test{}.Fuzz()
}

// HTML generates and opens HTML coverage report
func (Test) HTML() error {
	ctx, err := operations.NewOperation("Generating HTML Coverage Report")
	if err != nil {
		return err
	}

	// Check if coverage file exists
	if !utils.FileExists("coverage.out") {
		return ctx.Complete(errNoCoverageFile)
	}

	// Generate HTML report
	if err := utils.RunCommand("go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html"); err != nil {
		return ctx.Complete(fmt.Errorf("failed to generate HTML coverage: %w", err))
	}

	ctx.Success("Coverage report saved to coverage.html")

	// Open in browser
	if err := utils.OpenInBrowser("coverage.html"); err != nil {
		ctx.Warn("Failed to open coverage in browser: %v", err)
	}

	return ctx.Complete(nil)
}

// CoverageMerge merges coverage profiles from multiple modules
func (Test) CoverageMerge() error {
	ctx, err := operations.NewOperation("Merging Coverage Profiles")
	if err != nil {
		return err
	}

	// Find all coverage files
	coverageFiles, err := findCoverageFiles()
	if err != nil {
		return ctx.Complete(fmt.Errorf("failed to find coverage files: %w", err))
	}

	if len(coverageFiles) == 0 {
		return ctx.Complete(errNoCoverageFilesToMerge)
	}

	ctx.Info("Found %d coverage files to merge", len(coverageFiles))

	// Merge coverage files using gocovmerge if available
	if utils.CommandExists("gocovmerge") {
		args := append([]string{"-o", "coverage.out"}, coverageFiles...)
		if err := utils.RunCommand("gocovmerge", args...); err != nil {
			return ctx.Complete(fmt.Errorf("failed to merge coverage files: %w", err))
		}
	} else {
		ctx.Warn("gocovmerge not found, using simple concatenation")
		if err := simpleMergeCoverage(coverageFiles, "coverage.out"); err != nil {
			return ctx.Complete(fmt.Errorf("failed to merge coverage files: %w", err))
		}
	}

	ctx.Success("Merged coverage saved to coverage.out")
	return ctx.Complete(nil)
}

// Helper functions

func findCoverageFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor and hidden directories
		if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || info.Name() == "vendor") {
			return filepath.SkipDir
		}

		// Look for coverage files
		if strings.HasSuffix(path, ".cover") || strings.HasSuffix(path, ".coverage") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

func simpleMergeCoverage(inputFiles []string, outputFile string) error {
	// This is a simple merge that may not be 100% accurate
	// For production use, gocovmerge is recommended

	output, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer output.Close()

	// Write mode line
	fmt.Fprintln(output, "mode: set")

	// Concatenate all files, skipping mode lines
	for _, file := range inputFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if line != "" && !strings.HasPrefix(line, "mode:") {
				fmt.Fprintln(output, line)
			}
		}
	}

	return nil
}
