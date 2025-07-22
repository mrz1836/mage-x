// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Bench namespace for benchmark-related tasks
type Bench mg.Namespace

// Default runs all benchmarks with memory profiling
func (Bench) Default() error {
	utils.Header("Running Benchmarks")

	args := []string{"test", "-bench=.", "-benchmem", "-run=^$"}

	// Add bench time
	benchTime := utils.GetEnv("BENCH_TIME", "10s")
	args = append(args, "-benchtime", benchTime)

	// Add count if specified
	if count := os.Getenv("BENCH_COUNT"); count != "" {
		args = append(args, "-count", count)
	}

	// Add CPU profile if requested
	if cpuProfile := os.Getenv("BENCH_CPU_PROFILE"); cpuProfile != "" {
		args = append(args, "-cpuprofile", cpuProfile)
		utils.Info("CPU profile will be saved to: %s", cpuProfile)
	}

	// Add memory profile if requested
	if memProfile := os.Getenv("BENCH_MEM_PROFILE"); memProfile != "" {
		args = append(args, "-memprofile", memProfile)
		utils.Info("Memory profile will be saved to: %s", memProfile)
	}

	args = append(args, "./...")

	start := time.Now()
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("benchmarks failed: %w", err)
	}

	utils.Success("Benchmarks completed in %s", utils.FormatDuration(time.Since(start)))

	// Show how to analyze profiles if created
	if cpuProfile := os.Getenv("BENCH_CPU_PROFILE"); cpuProfile != "" {
		utils.Info("\nAnalyze CPU profile with:")
		utils.Info("  go tool pprof %s", cpuProfile)
	}
	if memProfile := os.Getenv("BENCH_MEM_PROFILE"); memProfile != "" {
		utils.Info("\nAnalyze memory profile with:")
		utils.Info("  go tool pprof %s", memProfile)
	}

	return nil
}

// Compare compares benchmark results
func (Bench) Compare() error {
	utils.Header("Comparing Benchmark Results")

	// Check if benchstat is installed
	if !utils.CommandExists("benchstat") {
		utils.Info("Installing benchstat...")
		if err := GetRunner().RunCmd("go", "install", "golang.org/x/perf/cmd/benchstat@latest"); err != nil {
			return fmt.Errorf("failed to install benchstat: %w", err)
		}
	}

	// Get benchmark files
	old := utils.GetEnv("BENCH_OLD", "old.txt")
	new := utils.GetEnv("BENCH_NEW", "new.txt")

	// Check if files exist
	if !utils.FileExists(old) {
		return fmt.Errorf("old benchmark file not found: %s", old)
	}
	if !utils.FileExists(new) {
		return fmt.Errorf("new benchmark file not found: %s", new)
	}

	utils.Info("Comparing %s vs %s", old, new)

	// Run benchstat
	return GetRunner().RunCmd("benchstat", old, new)
}

// Save saves current benchmark results
func (Bench) Save() error {
	utils.Header("Saving Benchmark Results")

	// Determine output file
	output := os.Getenv("BENCH_FILE")
	if output == "" {
		// Generate filename with timestamp
		output = fmt.Sprintf("bench-%s.txt", time.Now().Format("20060102-150405"))
	}

	utils.Info("Saving results to: %s", output)

	// Create output directory if needed
	fileOps := fileops.New()
	if dir := filepath.Dir(output); dir != "." && dir != "" {
		if err := fileOps.File.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Run benchmarks and save output
	args := []string{"test", "-bench=.", "-benchmem", "-run=^$"}

	benchTime := utils.GetEnv("BENCH_TIME", "10s")
	args = append(args, "-benchtime", benchTime)

	if count := os.Getenv("BENCH_COUNT"); count != "" {
		args = append(args, "-count", count)
	}

	args = append(args, "./...")

	utils.Info("Running: go %s", strings.Join(args, " "))

	// Execute and capture output
	output2, err := GetRunner().RunCmdOutput("go", args...)
	if err != nil {
		return fmt.Errorf("benchmarks failed: %w", err)
	}

	// Write output
	if err := fileOps.File.WriteFile(output, []byte(output2), 0o644); err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}

	utils.Success("Benchmark results saved to: %s", output)
	return nil
}

// CPU runs benchmarks with CPU profiling
func (Bench) CPU() error {
	utils.Header("Running Benchmarks with CPU Profiling")

	profile := utils.GetEnv("CPU_PROFILE", "cpu.prof")
	os.Setenv("BENCH_CPU_PROFILE", profile)

	var b Bench
	if err := b.Default(); err != nil {
		return err
	}

	// Analyze the profile
	utils.Info("\nAnalyzing CPU profile...")

	// Show top functions
	utils.Info("\nTop CPU consuming functions:")
	GetRunner().RunCmd("go", "tool", "pprof", "-top", profile)

	return nil
}

// Mem runs benchmarks with memory profiling
func (Bench) Mem() error {
	utils.Header("Running Benchmarks with Memory Profiling")

	profile := utils.GetEnv("MEM_PROFILE", "mem.prof")
	os.Setenv("BENCH_MEM_PROFILE", profile)

	var b Bench
	if err := b.Default(); err != nil {
		return err
	}

	// Analyze the profile
	utils.Info("\nAnalyzing memory profile...")

	// Show top memory allocations
	utils.Info("\nTop memory allocating functions:")
	GetRunner().RunCmd("go", "tool", "pprof", "-top", "-alloc_space", profile)

	return nil
}

// Profile runs benchmarks with both CPU and memory profiling
func (Bench) Profile() error {
	utils.Header("Running Benchmarks with Full Profiling")

	os.Setenv("BENCH_CPU_PROFILE", "cpu.prof")
	os.Setenv("BENCH_MEM_PROFILE", "mem.prof")

	var b Bench
	if err := b.Default(); err != nil {
		return err
	}

	utils.Success("\nProfiling data saved:")
	utils.Info("  CPU profile: cpu.prof")
	utils.Info("  Memory profile: mem.prof")
	utils.Info("\nUse 'go tool pprof' to analyze the profiles")

	return nil
}

// Trace runs benchmarks with execution tracing
func (Bench) Trace() error {
	utils.Header("Running Benchmarks with Execution Trace")

	trace := utils.GetEnv("TRACE_FILE", "trace.out")

	args := []string{"test", "-bench=.", "-benchmem", "-run=^$", "-trace", trace}

	benchTime := utils.GetEnv("BENCH_TIME", "10s")
	args = append(args, "-benchtime", benchTime)

	args = append(args, "./...")

	utils.Info("Trace will be saved to: %s", trace)

	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("benchmarks failed: %w", err)
	}

	utils.Success("Trace saved to: %s", trace)
	utils.Info("\nAnalyze trace with:")
	utils.Info("  go tool trace %s", trace)

	return nil
}

// Regression checks for performance regressions
func (Bench) Regression() error {
	utils.Header("Checking for Performance Regressions")

	// Save current results
	os.Setenv("BENCH_FILE", "bench-current.txt")
	var b Bench
	if err := b.Save(); err != nil {
		return err
	}

	// Check if we have a baseline
	baseline := utils.GetEnv("BENCH_BASELINE", "bench-baseline.txt")
	if !utils.FileExists(baseline) {
		utils.Warn("No baseline found at %s", baseline)
		utils.Info("Creating baseline from current results...")

		if err := os.Rename("bench-current.txt", baseline); err != nil {
			return fmt.Errorf("failed to create baseline: %w", err)
		}

		utils.Success("Baseline created. Run benchmarks again to compare.")
		return nil
	}

	// Compare with baseline
	os.Setenv("BENCH_OLD", baseline)
	os.Setenv("BENCH_NEW", "bench-current.txt")

	utils.Info("\nComparing with baseline...")
	if err := b.Compare(); err != nil {
		return err
	}

	// Ask about updating baseline
	utils.Info("\nUpdate baseline with current results? (set UPDATE_BASELINE=true)")
	if os.Getenv("UPDATE_BASELINE") == "true" {
		if err := os.Rename("bench-current.txt", baseline); err != nil {
			return fmt.Errorf("failed to update baseline: %w", err)
		}
		utils.Success("Baseline updated")
	}

	return nil
}
