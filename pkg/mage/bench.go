// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for err113 compliance
var (
	ErrOldBenchFileNotFound = errors.New("old benchmark file not found")
	ErrNewBenchFileNotFound = errors.New("new benchmark file not found")
)

// Bench namespace for benchmark-related tasks
type Bench mg.Namespace

// Default runs all benchmarks with memory profiling
func (Bench) Default() error {
	return Bench{}.DefaultWithArgs()
}

// DefaultWithArgs runs all benchmarks with memory profiling and accepts parameters
func (Bench) DefaultWithArgs(argsList ...string) error {
	utils.Header("Running Benchmarks")

	// Parse command-line parameters
	params := utils.ParseParams(argsList)

	args := []string{"test", "-bench=.", "-benchmem", "-run=^$"}

	// Add bench time from parameter or default
	benchTime := utils.GetParam(params, "time", "10s")
	args = append(args, "-benchtime", benchTime)

	// Get count from parameter
	count := utils.GetParam(params, "count", "")
	if count != "" {
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
		utils.Info("Analyze CPU profile with:")
		utils.Info("  go tool pprof %s", cpuProfile)
	}
	if memProfile := os.Getenv("BENCH_MEM_PROFILE"); memProfile != "" {
		utils.Info("Analyze memory profile with:")
		utils.Info("  go tool pprof %s", memProfile)
	}

	return nil
}

// Compare compares benchmark results
func (Bench) Compare() error {
	return Bench{}.CompareWithArgs()
}

// CompareWithArgs compares benchmark results with parameters
func (Bench) CompareWithArgs(argsList ...string) error {
	utils.Header("Comparing Benchmark Results")

	// Parse command-line parameters
	params := utils.ParseParams(argsList)

	// Check if benchstat is installed
	if !utils.CommandExists("benchstat") {
		utils.Info("Installing benchstat...")
		if err := GetRunner().RunCmd("go", "install", "golang.org/x/perf/cmd/benchstat@latest"); err != nil {
			return fmt.Errorf("failed to install benchstat: %w", err)
		}
	}

	// Get benchmark files from parameters or environment
	oldBenchFile := utils.GetParam(params, "old", "")
	if oldBenchFile == "" {
		oldBenchFile = utils.GetEnv("BENCH_OLD", "old.txt")
	}
	newBenchFile := utils.GetParam(params, "new", "")
	if newBenchFile == "" {
		newBenchFile = utils.GetEnv("BENCH_NEW", "new.txt")
	}

	// Check if files exist
	if !utils.FileExists(oldBenchFile) {
		return fmt.Errorf("%w: %s", ErrOldBenchFileNotFound, oldBenchFile)
	}
	if !utils.FileExists(newBenchFile) {
		return fmt.Errorf("%w: %s", ErrNewBenchFileNotFound, newBenchFile)
	}

	utils.Info("Comparing %s vs %s", oldBenchFile, newBenchFile)

	// Run benchstat
	return GetRunner().RunCmd("benchstat", oldBenchFile, newBenchFile)
}

// Save saves current benchmark results
func (Bench) Save() error {
	return Bench{}.SaveWithArgs()
}

// SaveWithArgs saves current benchmark results with parameters
func (Bench) SaveWithArgs(argsList ...string) error {
	utils.Header("Saving Benchmark Results")

	// Parse command-line parameters
	params := utils.ParseParams(argsList)

	// Determine output file
	output := utils.GetParam(params, "output", "")
	if output == "" {
		output = os.Getenv("BENCH_FILE")
	}
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

	benchTime := utils.GetParam(params, "time", "10s")
	args = append(args, "-benchtime", benchTime)

	// Get count from parameter
	count := utils.GetParam(params, "count", "")
	if count != "" {
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
	return Bench{}.CPUWithArgs()
}

// CPUWithArgs runs benchmarks with CPU profiling and accepts parameters
func (Bench) CPUWithArgs(argsList ...string) error {
	utils.Header("Running Benchmarks with CPU Profiling")

	// Parse command-line parameters
	params := utils.ParseParams(argsList)

	profile := utils.GetParam(params, "profile", "cpu.prof")
	if profile == "" {
		profile = utils.GetEnv("CPU_PROFILE", "cpu.prof")
	}
	if err := os.Setenv("BENCH_CPU_PROFILE", profile); err != nil {
		return fmt.Errorf("failed to set BENCH_CPU_PROFILE: %w", err)
	}

	var b Bench
	if err := b.DefaultWithArgs(argsList...); err != nil {
		return err
	}

	// Analyze the profile
	utils.Info("Analyzing CPU profile...")

	// Show top functions
	utils.Info("Top CPU consuming functions:")
	if err := GetRunner().RunCmd("go", "tool", "pprof", "-top", profile); err != nil {
		// Error expected and acceptable - analysis command may fail
		utils.Warn("CPU profile analysis failed: %v", err)
	}

	return nil
}

// Mem runs benchmarks with memory profiling
func (Bench) Mem() error {
	return Bench{}.MemWithArgs()
}

// MemWithArgs runs benchmarks with memory profiling and accepts parameters
func (Bench) MemWithArgs(argsList ...string) error {
	utils.Header("Running Benchmarks with Memory Profiling")

	// Parse command-line parameters
	params := utils.ParseParams(argsList)

	profile := utils.GetParam(params, "profile", "mem.prof")
	if profile == "" {
		profile = utils.GetEnv("MEM_PROFILE", "mem.prof")
	}
	if err := os.Setenv("BENCH_MEM_PROFILE", profile); err != nil {
		return fmt.Errorf("failed to set BENCH_MEM_PROFILE: %w", err)
	}

	var b Bench
	if err := b.DefaultWithArgs(argsList...); err != nil {
		return err
	}

	// Analyze the profile
	utils.Info("Analyzing memory profile...")

	// Show top memory allocations
	utils.Info("Top memory allocating functions:")
	if err := GetRunner().RunCmd("go", "tool", "pprof", "-top", "-alloc_space", profile); err != nil {
		// Error expected and acceptable - analysis command may fail
		utils.Warn("Memory profile analysis failed: %v", err)
	}

	return nil
}

// Profile runs benchmarks with both CPU and memory profiling
func (Bench) Profile() error {
	return Bench{}.ProfileWithArgs()
}

// ProfileWithArgs runs benchmarks with both CPU and memory profiling and accepts parameters
func (Bench) ProfileWithArgs(argsList ...string) error {
	utils.Header("Running Benchmarks with Full Profiling")

	// Parse command-line parameters
	params := utils.ParseParams(argsList)

	cpuProfile := utils.GetParam(params, "cpu-profile", "cpu.prof")
	memProfile := utils.GetParam(params, "mem-profile", "mem.prof")

	if err := os.Setenv("BENCH_CPU_PROFILE", cpuProfile); err != nil {
		return fmt.Errorf("failed to set BENCH_CPU_PROFILE: %w", err)
	}
	if err := os.Setenv("BENCH_MEM_PROFILE", memProfile); err != nil {
		return fmt.Errorf("failed to set BENCH_MEM_PROFILE: %w", err)
	}

	var b Bench
	if err := b.DefaultWithArgs(argsList...); err != nil {
		return err
	}

	utils.Success("Profiling data saved:")
	utils.Info("  CPU profile: cpu.prof")
	utils.Info("  Memory profile: mem.prof")
	utils.Info("Use 'go tool pprof' to analyze the profiles")

	return nil
}

// Trace runs benchmarks with execution tracing
func (Bench) Trace() error {
	return Bench{}.TraceWithArgs()
}

// TraceWithArgs runs benchmarks with execution tracing and accepts parameters
func (Bench) TraceWithArgs(argsList ...string) error {
	utils.Header("Running Benchmarks with Execution Trace")

	// Parse command-line parameters
	params := utils.ParseParams(argsList)

	trace := utils.GetParam(params, "trace", "trace.out")
	if trace == "" {
		trace = utils.GetEnv("TRACE_FILE", "trace.out")
	}

	args := []string{"test", "-bench=.", "-benchmem", "-run=^$", "-trace", trace}

	benchTime := utils.GetParam(params, "time", "10s")
	args = append(args, "-benchtime", benchTime, "./...")

	utils.Info("Trace will be saved to: %s", trace)

	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("benchmarks failed: %w", err)
	}

	utils.Success("Trace saved to: %s", trace)
	utils.Info("Analyze trace with:")
	utils.Info("  go tool trace %s", trace)

	return nil
}

// Regression checks for performance regressions
func (Bench) Regression() error {
	return Bench{}.RegressionWithArgs()
}

// RegressionWithArgs checks for performance regressions with parameters
func (Bench) RegressionWithArgs(argsList ...string) error {
	utils.Header("Checking for Performance Regressions")

	// Parse command-line parameters
	params := utils.ParseParams(argsList)

	// Save current results
	currentFile := "bench-current.txt"
	if err := os.Setenv("BENCH_FILE", currentFile); err != nil {
		return fmt.Errorf("failed to set BENCH_FILE: %w", err)
	}
	var b Bench
	if err := b.SaveWithArgs(argsList...); err != nil {
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
	if err := os.Setenv("BENCH_OLD", baseline); err != nil {
		return fmt.Errorf("failed to set BENCH_OLD: %w", err)
	}
	if err := os.Setenv("BENCH_NEW", currentFile); err != nil {
		return fmt.Errorf("failed to set BENCH_NEW: %w", err)
	}

	utils.Info("Comparing with baseline...")
	compareArgs := []string{"old=" + baseline, "new=" + currentFile}
	if err := b.CompareWithArgs(compareArgs...); err != nil {
		return err
	}

	// Ask about updating baseline
	updateBaseline := utils.GetParam(params, "update-baseline", "")
	if updateBaseline == "" {
		utils.Info("Update baseline with current results? (set UPDATE_BASELINE=true or use update-baseline=true parameter)")
		updateBaseline = os.Getenv("UPDATE_BASELINE")
	}
	if updateBaseline == approvalTrue || utils.IsParamTrue(params, "update-baseline") {
		if err := os.Rename(currentFile, baseline); err != nil {
			return fmt.Errorf("failed to update baseline: %w", err)
		}
		utils.Success("Baseline updated")
	}

	return nil
}
