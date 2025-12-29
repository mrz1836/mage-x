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

	"github.com/mrz1836/mage-x/pkg/common/env"
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

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Parse command-line parameters from os.Args
	// Find arguments after the target name
	var targetArgs []string
	for i, arg := range os.Args {
		if strings.Contains(arg, "bench:default") || strings.Contains(arg, "bench") {
			targetArgs = os.Args[i+1:]
			break
		}
	}
	params := utils.ParseParams(targetArgs)

	// Discover and filter modules
	result, err := discoverAndFilterModules(config, ModuleDiscoveryOptions{
		Operation: "benchmarks",
	})
	if err != nil {
		return err
	}
	if result.Empty || result.Skipped {
		return nil
	}

	// Build base benchmark args
	args := []string{"test", "-bench=.", "-benchmem", "-run=^$"}

	// Add bench time from parameter or default
	benchTime := utils.GetParam(params, "time", "3s")
	args = append(args, "-benchtime", benchTime)

	// Add verbose flag if requested
	if utils.IsParamTrue(params, "verbose") {
		args = append(args, "-v")
	}

	// Get count from parameter
	count := utils.GetParam(params, "count", "")
	if count != "" {
		args = append(args, "-count", count)
	}

	// Add skip pattern if specified
	skip := utils.GetParam(params, "skip", "")
	if skip != "" {
		args = append(args, "-skip", skip)
	}

	// Add CPU profile if requested
	if cpuProfile := GetMageXEnv("BENCH_CPU_PROFILE"); cpuProfile != "" {
		args = append(args, "-cpuprofile", cpuProfile)
		utils.Info("CPU profile will be saved to: %s", cpuProfile)
	}

	// Add memory profile if requested
	if memProfile := GetMageXEnv("BENCH_MEM_PROFILE"); memProfile != "" {
		args = append(args, "-memprofile", memProfile)
		utils.Info("Memory profile will be saved to: %s", memProfile)
	}

	// Add package filter - use ./... for each module
	pkg := utils.GetParam(params, "pkg", "./...")
	args = append(args, pkg)

	// Run benchmarks for each module
	err = forEachModule(result.Modules, ModuleIteratorOptions{
		Operation: "Benchmarks",
		Verb:      "completed",
	}, func(module ModuleInfo) error {
		// Show the command being run for transparency
		utils.Info("Running: go %s", strings.Join(args, " "))
		return runCommandInModule(module, "go", args...)
	})
	if err != nil {
		return err
	}

	// Show how to analyze profiles if created
	if cpuProfile := GetMageXEnv("BENCH_CPU_PROFILE"); cpuProfile != "" {
		utils.Info("Analyze CPU profile with:")
		utils.Info("  go tool pprof %s", cpuProfile)
	}
	if memProfile := GetMageXEnv("BENCH_MEM_PROFILE"); memProfile != "" {
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

	// Ensure benchstat is installed using consolidated helper
	if err := installTool(ToolDefinition{
		Name:   "benchstat",
		Module: "golang.org/x/perf/cmd/benchstat",
		Check:  "benchstat",
	}); err != nil {
		return err
	}

	// Get benchmark files from parameters or environment
	oldBenchFile := utils.GetParam(params, "old", "")
	if oldBenchFile == "" {
		oldBenchFile = env.GetString("BENCH_OLD", "old.txt")
	}
	newBenchFile := utils.GetParam(params, "new", "")
	if newBenchFile == "" {
		newBenchFile = env.GetString("BENCH_NEW", "new.txt")
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

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Parse command-line parameters from os.Args
	// Find arguments after the target name
	var targetArgs []string
	for i, arg := range os.Args {
		if strings.Contains(arg, "bench:save") {
			targetArgs = os.Args[i+1:]
			break
		}
	}
	params := utils.ParseParams(targetArgs)

	// Determine output file
	output := utils.GetParam(params, "output", "")
	if output == "" {
		output = GetMageXEnv("BENCH_FILE")
	}
	if output == "" {
		// Generate filename with timestamp
		output = fmt.Sprintf("bench-%s.txt", time.Now().Format("20060102-150405"))
	}

	utils.Info("Saving results to: %s", output)

	// Create output directory if needed
	fileOps := fileops.New()
	if dir := filepath.Dir(output); dir != "." && dir != "" {
		if mkdirErr := fileOps.File.MkdirAll(dir, fileops.PermDir); mkdirErr != nil {
			return fmt.Errorf("failed to create directory: %w", mkdirErr)
		}
	}

	// Discover and filter modules
	result, err := discoverAndFilterModules(config, ModuleDiscoveryOptions{
		Operation: "benchmarks",
	})
	if err != nil {
		return err
	}
	if result.Empty || result.Skipped {
		return nil
	}

	// Build base benchmark args
	args := []string{"test", "-bench=.", "-benchmem", "-run=^$"}

	benchTime := utils.GetParam(params, "time", "3s")
	args = append(args, "-benchtime", benchTime)

	// Add verbose flag if requested
	if utils.IsParamTrue(params, "verbose") {
		args = append(args, "-v")
	}

	// Get count from parameter
	count := utils.GetParam(params, "count", "")
	if count != "" {
		args = append(args, "-count", count)
	}

	// Add skip pattern if specified
	skip := utils.GetParam(params, "skip", "")
	if skip != "" {
		args = append(args, "-skip", skip)
	}

	// Add package filter - use ./... for each module
	pkg := utils.GetParam(params, "pkg", "./...")
	args = append(args, pkg)

	// Collect all benchmark outputs
	// Note: This function needs output capture, so we use runCommandInModuleOutput
	// instead of forEachModule
	var allOutputs []string

	for _, module := range result.Modules {
		displayModuleHeader(module, "Running benchmarks for")

		utils.Info("Running: go %s", strings.Join(args, " "))

		// Execute and capture output
		moduleOutput, err := runCommandInModuleOutput(module, "go", args...)
		if err != nil {
			return fmt.Errorf("benchmarks failed for %s: %w", module.Relative, err)
		}

		// Add module header to output
		if module.Relative == "." {
			allOutputs = append(allOutputs, fmt.Sprintf("# Module: main\n%s", moduleOutput))
		} else {
			allOutputs = append(allOutputs, fmt.Sprintf("# Module: %s\n%s", module.Relative, moduleOutput))
		}
	}

	// Combine all outputs
	combinedOutput := strings.Join(allOutputs, "\n")

	// Write output
	if err := fileOps.File.WriteFile(output, []byte(combinedOutput), fileops.PermFile); err != nil {
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
		profile = env.GetString("CPU_PROFILE", "cpu.prof")
	}
	if err := os.Setenv("MAGE_X_BENCH_CPU_PROFILE", profile); err != nil {
		return fmt.Errorf("failed to set MAGE_X_BENCH_CPU_PROFILE: %w", err)
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
		profile = env.GetString("MEM_PROFILE", "mem.prof")
	}
	if err := os.Setenv("MAGE_X_BENCH_MEM_PROFILE", profile); err != nil {
		return fmt.Errorf("failed to set MAGE_X_BENCH_MEM_PROFILE: %w", err)
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

	if err := os.Setenv("MAGE_X_BENCH_CPU_PROFILE", cpuProfile); err != nil {
		return fmt.Errorf("failed to set MAGE_X_BENCH_CPU_PROFILE: %w", err)
	}
	if err := os.Setenv("MAGE_X_BENCH_MEM_PROFILE", memProfile); err != nil {
		return fmt.Errorf("failed to set MAGE_X_BENCH_MEM_PROFILE: %w", err)
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

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Parse command-line parameters from os.Args
	// Find arguments after the target name
	var targetArgs []string
	for i, arg := range os.Args {
		if strings.Contains(arg, "bench:trace") {
			targetArgs = os.Args[i+1:]
			break
		}
	}
	params := utils.ParseParams(targetArgs)

	trace := utils.GetParam(params, "trace", "trace.out")
	if trace == "" {
		trace = env.GetString("TRACE_FILE", "trace.out")
	}

	// Discover and filter modules
	result, err := discoverAndFilterModules(config, ModuleDiscoveryOptions{
		Operation: "benchmarks",
	})
	if err != nil {
		return err
	}
	if result.Empty || result.Skipped {
		return nil
	}

	benchTime := utils.GetParam(params, "time", "3s")

	// Add skip pattern if specified
	skip := utils.GetParam(params, "skip", "")

	// Add package filter - use ./... for each module
	pkg := utils.GetParam(params, "pkg", "./...")

	totalStart := time.Now()
	var moduleErrors []moduleError

	// Run benchmarks for each module
	// Note: Custom loop because we need per-module trace file naming
	for i, module := range result.Modules {
		displayModuleHeader(module, "Running benchmarks with trace for")

		// Create unique trace file for each module
		moduleTrace := trace
		if len(result.Modules) > 1 {
			ext := filepath.Ext(trace)
			base := strings.TrimSuffix(trace, ext)
			if module.Relative == "." {
				moduleTrace = fmt.Sprintf("%s_main%s", base, ext)
			} else {
				sanitized := strings.ReplaceAll(module.Relative, "/", "_")
				moduleTrace = fmt.Sprintf("%s_%s%s", base, sanitized, ext)
			}
		}

		args := []string{"test", "-bench=.", "-benchmem", "-run=^$", "-trace", moduleTrace}
		args = append(args, "-benchtime", benchTime)

		// Add verbose flag if requested
		if utils.IsParamTrue(params, "verbose") {
			args = append(args, "-v")
		}

		if skip != "" {
			args = append(args, "-skip", skip)
		}

		args = append(args, pkg)

		utils.Info("Trace will be saved to: %s", moduleTrace)

		moduleStart := time.Now()
		err := runCommandInModule(module, "go", args...)
		if err != nil {
			moduleErrors = append(moduleErrors, moduleError{Module: module, Error: err})
			displayModuleCompletion(module, "Benchmarks with trace", moduleStart, err)
		} else {
			// Custom message includes trace file path - keep as-is
			utils.Success("Trace saved for %s to: %s in %s", module.Relative, moduleTrace, utils.FormatDuration(time.Since(moduleStart)))
		}

		// Show trace analysis hint for first successful module
		if err == nil && i == 0 {
			utils.Info("Analyze trace with:")
			utils.Info("  go tool trace %s", moduleTrace)
		}
	}

	// Report overall results
	if len(moduleErrors) > 0 {
		utils.Error("Benchmarks with trace failed in %d/%d modules", len(moduleErrors), len(result.Modules))
		return formatModuleErrors(moduleErrors)
	}

	displayOverallCompletion("benchmarks with trace", "completed", totalStart)
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
	if err := os.Setenv("MAGE_X_BENCH_FILE", currentFile); err != nil {
		return fmt.Errorf("failed to set MAGE_X_BENCH_FILE: %w", err)
	}
	var b Bench
	if err := b.SaveWithArgs(argsList...); err != nil {
		return err
	}

	// Check if we have a baseline
	baseline := env.GetString("BENCH_BASELINE", "bench-baseline.txt")
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
		updateBaseline = GetMageXEnv("UPDATE_BASELINE")
	}
	if updateBaseline == trueValue || utils.IsParamTrue(params, "update-baseline") {
		if err := os.Rename(currentFile, baseline); err != nil {
			return fmt.Errorf("failed to update baseline: %w", err)
		}
		utils.Success("Baseline updated")
	}

	return nil
}
