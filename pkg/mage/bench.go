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
	"github.com/mrz1836/mage-x/pkg/mage/runtimectx"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for err113 compliance
var (
	ErrOldBenchFileNotFound = errors.New("old benchmark file not found")
	ErrNewBenchFileNotFound = errors.New("new benchmark file not found")
)

// Benchmark default constants
const (
	defaultBenchTime    = "3s"
	defaultCPUProfile   = "cpu.prof"
	defaultMemProfile   = "mem.prof"
	defaultTraceFile    = "trace.out"
	defaultCurrentFile  = "bench-current.txt"
	defaultBaselineFile = "bench-baseline.txt"
	defaultOldFile      = "old.txt"
	defaultNewFile      = "new.txt"
	defaultPackage      = "./..."
)

// Bench namespace for benchmark-related tasks
type Bench mg.Namespace

// Default runs all benchmarks with memory profiling
func (Bench) Default() error {
	return Bench{}.DefaultWithArgs()
}

// DefaultWithArgs runs all benchmarks with memory profiling and accepts parameters
func (Bench) DefaultWithArgs(argsList ...string) error {
	_, err := runBenchDefault(argsList...)
	return err
}

// runBenchDefault runs the benchmarks for DefaultWithArgs, one pass per build
// tag set, and returns the passes it ran so CPU, Mem, and Profile can find each
// pass's profile. It returns no passes when there are no modules to benchmark.
func runBenchDefault(argsList ...string) ([]string, error) {
	ctx, err := prepareModuleCommand(ModuleCommandConfig{
		Header:    "Running Benchmarks",
		Operation: "benchmarks",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to prepare benchmark command: %w", err)
	}
	if ctx == nil {
		return nil, nil
	}

	// Parse command-line parameters from argsList
	params := utils.ParseParams(argsList)

	// Profiles requested by CPU, Mem, or Profile
	cpuProfile := GetMageXEnv("BENCH_CPU_PROFILE")
	memProfile := GetMageXEnv("BENCH_MEM_PROFILE")

	passes := testBuildTagPasses(ctx.Config, discoverTestBuildTags(ctx.Config))
	multiPass := len(passes) > 1

	for _, tags := range passes {
		passCPUProfile := benchPassOutput(cpuProfile, tags, multiPass)
		if passCPUProfile != "" {
			utils.Info("CPU profile will be saved to: %s", passCPUProfile)
		}
		passMemProfile := benchPassOutput(memProfile, tags, multiPass)
		if passMemProfile != "" {
			utils.Info("Memory profile will be saved to: %s", passMemProfile)
		}

		// Run benchmarks for each module
		args := benchArgs(params, tags, passCPUProfile, passMemProfile)
		err = forEachModule(ctx.Modules, ModuleIteratorOptions{
			Operation: "Benchmarks",
			Verb:      "completed",
		}, func(module ModuleInfo) error {
			// Show the command being run for transparency
			utils.Info("Running: go %s", strings.Join(args, " "))
			return runCommandInModule(module, "go", args...)
		})
		if err != nil {
			return passes, fmt.Errorf("failed to run benchmarks%s: %w", benchPassInfo(tags, multiPass), err)
		}
	}

	// Show how to analyze profiles if created
	showProfileHints("CPU", benchPassOutputs(cpuProfile, passes))
	showProfileHints("memory", benchPassOutputs(memProfile, passes))

	return passes, nil
}

// benchArgs returns the go test arguments DefaultWithArgs and SaveWithArgs run
// for one build tag pass. An empty profile name adds no profiling flag.
func benchArgs(params map[string]string, tags, cpuProfile, memProfile string) []string {
	args := []string{"test", "-bench=.", "-benchmem", "-run=^$"}
	args = append(args, tagsFlag(tags)...)

	// Add bench time from parameter or default
	args = append(args, "-benchtime", utils.GetParam(params, "time", defaultBenchTime))

	// Add verbose flag if requested
	if utils.IsParamTrue(params, "verbose") {
		args = append(args, "-v")
	}

	// Get count from parameter
	if count := utils.GetParam(params, "count", ""); count != "" {
		args = append(args, "-count", count)
	}

	// Add skip pattern if specified
	if skip := utils.GetParam(params, "skip", ""); skip != "" {
		args = append(args, "-skip", skip)
	}

	if cpuProfile != "" {
		args = append(args, "-cpuprofile", cpuProfile)
	}
	if memProfile != "" {
		args = append(args, "-memprofile", memProfile)
	}

	// Add package filter - use ./... for each module
	return append(args, utils.GetParam(params, "pkg", defaultPackage))
}

// benchPassOutput returns the file one build tag pass writes for the output
// named name. A lone pass keeps name as is. With several passes each pass
// writes its own file, so a later pass cannot overwrite an earlier one: the
// pass's tags, joined by underscores, go before the extension, and "base"
// names the pass without tags (cpu.prof becomes cpu.base.prof, cpu.unit.prof,
// and cpu.unit_e2e.prof).
func benchPassOutput(name, tags string, multiPass bool) string {
	if name == "" || !multiPass {
		return name
	}
	token := buildTagFileToken(mergeBuildTags(tags))
	if token == "" {
		token = "base"
	}
	ext := filepath.Ext(name)
	return strings.TrimSuffix(name, ext) + "." + token + ext
}

// benchPassOutputs returns the files the given passes write for the output
// named name, or nil when name is empty. No passes (nothing was benchmarked)
// counts as one pass.
func benchPassOutputs(name string, passes []string) []string {
	if name == "" {
		return nil
	}
	if len(passes) <= 1 {
		return []string{name}
	}
	outputs := make([]string, 0, len(passes))
	for _, tags := range passes {
		outputs = append(outputs, benchPassOutput(name, tags, true))
	}
	return outputs
}

// benchPassInfo describes a build tag pass in messages. It is empty for a lone
// pass, so single-pass messages read as they always have.
func benchPassInfo(tags string, multiPass bool) string {
	if !multiPass {
		return ""
	}
	return buildTagPassInfo(tags)
}

// showProfileHints prints how to analyze the given profiles.
func showProfileHints(kind string, profiles []string) {
	if len(profiles) == 0 {
		return
	}
	utils.Info("Analyze %s profile with:", kind)
	for _, profile := range profiles {
		utils.Info("  go tool pprof %s", profile)
	}
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
		return fmt.Errorf("failed to install benchstat tool: %w", err)
	}

	// Get benchmark files from parameters or environment
	oldBenchFile := utils.GetParam(params, "old", "")
	if oldBenchFile == "" {
		oldBenchFile = env.GetString("BENCH_OLD", defaultOldFile)
	}
	newBenchFile := utils.GetParam(params, "new", "")
	if newBenchFile == "" {
		newBenchFile = env.GetString("BENCH_NEW", defaultNewFile)
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
	ctx, err := prepareModuleCommand(ModuleCommandConfig{
		Header:    "Saving Benchmark Results",
		Operation: "benchmarks",
	})
	if err != nil {
		return fmt.Errorf("failed to prepare benchmark save command: %w", err)
	}
	if ctx == nil {
		return nil
	}

	// Parse command-line parameters from argsList
	params := utils.ParseParams(argsList)

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

	passes := testBuildTagPasses(ctx.Config, discoverTestBuildTags(ctx.Config))
	multiPass := len(passes) > 1

	// Collect all benchmark outputs. With several build tag passes, each pass's
	// results are appended to the same file under a "# Build tags:" line.
	// Note: This function needs output capture, so we use runCommandInModuleOutput
	// instead of forEachModule
	var allOutputs []string

	for _, tags := range passes {
		if multiPass {
			label := tags
			if label == "" {
				label = "none"
			}
			allOutputs = append(allOutputs, "# Build tags: "+label)
		}

		args := benchArgs(params, tags, "", "")
		for _, module := range ctx.Modules {
			if err := runtimectx.CheckCanceled(); err != nil {
				return fmt.Errorf("benchmarks canceled: %w", err)
			}
			displayModuleHeader(module, "Running benchmarks for")

			utils.Info("Running: go %s", strings.Join(args, " "))

			// Execute and capture output
			moduleOutput, err := runCommandInModuleOutput(module, "go", args...)
			if err != nil {
				return fmt.Errorf("benchmarks failed for %s%s: %w", module.Relative, benchPassInfo(tags, multiPass), err)
			}

			// Add module header to output
			if module.Relative == "." {
				allOutputs = append(allOutputs, fmt.Sprintf("# Module: main\n%s", moduleOutput))
			} else {
				allOutputs = append(allOutputs, fmt.Sprintf("# Module: %s\n%s", module.Relative, moduleOutput))
			}
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

	profile := utils.GetParam(params, "profile", defaultCPUProfile)
	if profile == "" {
		profile = env.GetString("CPU_PROFILE", defaultCPUProfile)
	}
	if err := os.Setenv("MAGE_X_BENCH_CPU_PROFILE", profile); err != nil {
		return fmt.Errorf("failed to set MAGE_X_BENCH_CPU_PROFILE: %w", err)
	}

	passes, err := runBenchDefault(argsList...)
	if err != nil {
		return fmt.Errorf("failed to run CPU benchmarks: %w", err)
	}

	// Analyze the profile
	utils.Info("Analyzing CPU profile...")

	// Show top functions, once per build tag pass's profile
	for _, passProfile := range benchPassOutputs(profile, passes) {
		if len(passes) > 1 {
			utils.Info("Top CPU consuming functions in %s:", passProfile)
		} else {
			utils.Info("Top CPU consuming functions:")
		}
		if pprofErr := GetRunner().RunCmd("go", "tool", "pprof", "-top", passProfile); pprofErr != nil {
			// Error expected and acceptable - analysis command may fail
			utils.Warn("CPU profile analysis failed: %v", pprofErr)
		}
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

	profile := utils.GetParam(params, "profile", defaultMemProfile)
	if profile == "" {
		profile = env.GetString("MEM_PROFILE", defaultMemProfile)
	}
	if err := os.Setenv("MAGE_X_BENCH_MEM_PROFILE", profile); err != nil {
		return fmt.Errorf("failed to set MAGE_X_BENCH_MEM_PROFILE: %w", err)
	}

	passes, err := runBenchDefault(argsList...)
	if err != nil {
		return fmt.Errorf("failed to run memory benchmarks: %w", err)
	}

	// Analyze the profile
	utils.Info("Analyzing memory profile...")

	// Show top memory allocations, once per build tag pass's profile
	for _, passProfile := range benchPassOutputs(profile, passes) {
		if len(passes) > 1 {
			utils.Info("Top memory allocating functions in %s:", passProfile)
		} else {
			utils.Info("Top memory allocating functions:")
		}
		if pprofErr := GetRunner().RunCmd("go", "tool", "pprof", "-top", "-alloc_space", passProfile); pprofErr != nil {
			// Error expected and acceptable - analysis command may fail
			utils.Warn("Memory profile analysis failed: %v", pprofErr)
		}
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

	cpuProfile := utils.GetParam(params, "cpu-profile", defaultCPUProfile)
	memProfile := utils.GetParam(params, "mem-profile", defaultMemProfile)

	if err := os.Setenv("MAGE_X_BENCH_CPU_PROFILE", cpuProfile); err != nil {
		return fmt.Errorf("failed to set MAGE_X_BENCH_CPU_PROFILE: %w", err)
	}
	if err := os.Setenv("MAGE_X_BENCH_MEM_PROFILE", memProfile); err != nil {
		return fmt.Errorf("failed to set MAGE_X_BENCH_MEM_PROFILE: %w", err)
	}

	passes, err := runBenchDefault(argsList...)
	if err != nil {
		return fmt.Errorf("failed to run profiling benchmarks: %w", err)
	}

	utils.Success("Profiling data saved:")
	for _, passProfile := range benchPassOutputs(cpuProfile, passes) {
		utils.Info("  CPU profile: %s", passProfile)
	}
	for _, passProfile := range benchPassOutputs(memProfile, passes) {
		utils.Info("  Memory profile: %s", passProfile)
	}
	utils.Info("Use 'go tool pprof' to analyze the profiles")

	return nil
}

// Trace runs benchmarks with execution tracing
func (Bench) Trace() error {
	return Bench{}.TraceWithArgs()
}

// TraceWithArgs runs benchmarks with execution tracing and accepts parameters
func (Bench) TraceWithArgs(argsList ...string) error {
	ctx, err := prepareModuleCommand(ModuleCommandConfig{
		Header:    "Running Benchmarks with Execution Trace",
		Operation: "benchmarks",
	})
	if err != nil {
		return fmt.Errorf("failed to prepare benchmark trace command: %w", err)
	}
	if ctx == nil {
		return nil
	}

	// Parse command-line parameters from argsList
	params := utils.ParseParams(argsList)

	trace := utils.GetParam(params, "trace", defaultTraceFile)
	if trace == "" {
		trace = env.GetString("TRACE_FILE", defaultTraceFile)
	}

	passes := testBuildTagPasses(ctx.Config, discoverTestBuildTags(ctx.Config))
	multiPass := len(passes) > 1

	totalStart := time.Now()
	moduleErrors := make([]moduleError, 0, len(ctx.Modules)*len(passes))

	// Run benchmarks for each build tag pass and module
	// Note: Custom loop because we need per-module trace file naming
	for _, tags := range passes {
		for i, module := range ctx.Modules {
			displayModuleHeader(module, "Running benchmarks with trace for")

			// Create unique trace file for each module
			moduleTrace := trace
			if len(ctx.Modules) > 1 {
				ext := filepath.Ext(trace)
				base := strings.TrimSuffix(trace, ext)
				if module.Relative == "." {
					moduleTrace = fmt.Sprintf("%s_main%s", base, ext)
				} else {
					sanitized := strings.ReplaceAll(module.Relative, "/", "_")
					moduleTrace = fmt.Sprintf("%s_%s%s", base, sanitized, ext)
				}
			}
			// And for each build tag pass
			moduleTrace = benchPassOutput(moduleTrace, tags, multiPass)

			args := benchTraceArgs(params, tags, moduleTrace)

			utils.Info("Trace will be saved to: %s", moduleTrace)

			moduleStart := time.Now()
			err := runCommandInModule(module, "go", args...)
			if err != nil {
				if multiPass {
					err = fmt.Errorf("benchmarks%s: %w", buildTagPassInfo(tags), err)
				}
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
	}

	// Report overall results
	if len(moduleErrors) > 0 {
		utils.Error("Benchmarks with trace failed in %d/%d modules", len(moduleErrors), len(ctx.Modules)*len(passes))
		return formatModuleErrors(moduleErrors)
	}

	displayOverallCompletion("benchmarks with trace", "completed", totalStart)
	return nil
}

// benchTraceArgs returns the go test arguments TraceWithArgs runs for one
// module and build tag pass.
func benchTraceArgs(params map[string]string, tags, traceFile string) []string {
	args := []string{"test", "-bench=.", "-benchmem", "-run=^$"}
	args = append(args, tagsFlag(tags)...)
	args = append(args, "-trace", traceFile, "-benchtime", utils.GetParam(params, "time", defaultBenchTime))

	// Add verbose flag if requested
	if utils.IsParamTrue(params, "verbose") {
		args = append(args, "-v")
	}

	// Add skip pattern if specified
	if skip := utils.GetParam(params, "skip", ""); skip != "" {
		args = append(args, "-skip", skip)
	}

	// Add package filter - use ./... for each module
	return append(args, utils.GetParam(params, "pkg", defaultPackage))
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
	currentFile := defaultCurrentFile
	if err := os.Setenv("MAGE_X_BENCH_FILE", currentFile); err != nil {
		return fmt.Errorf("failed to set MAGE_X_BENCH_FILE: %w", err)
	}
	var b Bench
	if err := b.SaveWithArgs(argsList...); err != nil {
		return fmt.Errorf("failed to save benchmark results: %w", err)
	}

	// Check if we have a baseline
	baseline := env.GetString("BENCH_BASELINE", defaultBaselineFile)
	if !utils.FileExists(baseline) {
		utils.Warn("No baseline found at %s", baseline)
		utils.Info("Creating baseline from current results...")

		if err := os.Rename(currentFile, baseline); err != nil {
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
		return fmt.Errorf("failed to compare benchmark results: %w", err)
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
