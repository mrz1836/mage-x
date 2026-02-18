// Package mage provides fuzz test timing utilities for accurate timeout calculation.
//
// Go's fuzz tests have two distinct phases:
//  1. Baseline gathering - Runs ALL seed corpus entries before fuzzing starts
//  2. Fuzzing - The actual -fuzztime duration
//
// The -fuzztime flag only controls phase 2. Phase 1 runs before the timer starts,
// so total wall-clock time is: baseline_time + fuzztime.
//
// This package provides utilities to:
//   - Count seed corpus entries (f.Add() calls + testdata/fuzz/* files)
//   - Calculate accurate timeouts accounting for baseline gathering
//   - Report timing breakdown for debugging
package mage

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// FuzzSeedInfo contains information about a fuzz test's seed corpus
type FuzzSeedInfo struct {
	TestName       string // Name of the fuzz test function
	Package        string // Package path
	CodeSeeds      int    // Number of f.Add() calls in code
	CorpusSeeds    int    // Number of files in testdata/fuzz/<TestName>/
	TotalSeeds     int    // CodeSeeds + CorpusSeeds
	EstimatedTime  time.Duration
	SourceFile     string // Path to the test file containing this fuzz test
	HasLoopedSeeds bool   // True if f.Add is called inside a loop (count may be inaccurate)
}

// FuzzTimingConfig holds configuration for fuzz test timing calculations
type FuzzTimingConfig struct {
	BaselineOverheadPerSeed time.Duration // Time per seed during baseline gathering
	BaselineBuffer          time.Duration // Extra buffer time for safety margin (includes compilation)
	MaxTimeout              time.Duration // Maximum allowed timeout (cap)
	MinTimeout              time.Duration // Minimum timeout regardless of calculation
	WarmupTimeout           time.Duration // Timeout for pre-compiling fuzz test binary (0 disables warmup)
}

// DefaultFuzzTimingConfig returns the default fuzz timing configuration.
// Configuration can be overridden via environment variables:
//   - MAGE_X_FUZZ_BASELINE_BUFFER: Buffer time including compilation (default: "90s")
//   - MAGE_X_FUZZ_BASELINE_OVERHEAD_PER_SEED: Time per seed (default: "500ms")
//   - MAGE_X_FUZZ_MIN_TIMEOUT: Minimum timeout (default: "90s")
//   - MAGE_X_FUZZ_MAX_TIMEOUT: Maximum timeout cap (default: "30m")
//   - MAGE_X_FUZZ_WARMUP_TIMEOUT: Timeout for build cache warmup (default: "5m", "0s" disables)
func DefaultFuzzTimingConfig() FuzzTimingConfig {
	cfg := FuzzTimingConfig{
		BaselineOverheadPerSeed: 500 * time.Millisecond,
		BaselineBuffer:          90 * time.Second, // Increased to account for compilation overhead
		MaxTimeout:              30 * time.Minute,
		MinTimeout:              90 * time.Second, // Increased to account for compilation overhead
		WarmupTimeout:           5 * time.Minute,
	}

	// Allow environment variable overrides
	if envBuffer := os.Getenv("MAGE_X_FUZZ_BASELINE_BUFFER"); envBuffer != "" {
		if d, err := time.ParseDuration(envBuffer); err == nil && d >= 0 {
			cfg.BaselineBuffer = d
		}
	}

	if envOverhead := os.Getenv("MAGE_X_FUZZ_BASELINE_OVERHEAD_PER_SEED"); envOverhead != "" {
		if d, err := time.ParseDuration(envOverhead); err == nil && d > 0 {
			cfg.BaselineOverheadPerSeed = d
		}
	}

	if envMin := os.Getenv("MAGE_X_FUZZ_MIN_TIMEOUT"); envMin != "" {
		if d, err := time.ParseDuration(envMin); err == nil && d > 0 {
			cfg.MinTimeout = d
		}
	}

	if envMax := os.Getenv("MAGE_X_FUZZ_MAX_TIMEOUT"); envMax != "" {
		if d, err := time.ParseDuration(envMax); err == nil && d > 0 {
			cfg.MaxTimeout = d
		}
	}

	if envWarmup := os.Getenv("MAGE_X_FUZZ_WARMUP_TIMEOUT"); envWarmup != "" {
		if d, err := time.ParseDuration(envWarmup); err == nil && d >= 0 {
			cfg.WarmupTimeout = d
		}
	}

	return cfg
}

// FuzzTimingConfigFromTestConfig creates a FuzzTimingConfig from TestConfig
func FuzzTimingConfigFromTestConfig(cfg *TestConfig) FuzzTimingConfig {
	ftc := DefaultFuzzTimingConfig()

	if cfg == nil {
		return ftc
	}

	// Parse overhead per seed
	if cfg.FuzzBaselineOverheadPerSeed != "" {
		if d, err := time.ParseDuration(cfg.FuzzBaselineOverheadPerSeed); err == nil && d > 0 {
			ftc.BaselineOverheadPerSeed = d
		}
	}

	// Parse baseline buffer
	if cfg.FuzzBaselineBuffer != "" {
		if d, err := time.ParseDuration(cfg.FuzzBaselineBuffer); err == nil && d >= 0 {
			ftc.BaselineBuffer = d
		}
	}

	return ftc
}

// CountFuzzSeeds counts the total number of seed corpus entries for a fuzz test.
// It combines:
//   - f.Add() calls found in the test source code
//   - Files in testdata/fuzz/<TestName>/ directory
//
// The pkgDir should be the directory containing the test files.
// Returns FuzzSeedInfo with details about the seed corpus.
func CountFuzzSeeds(pkgDir, testName string) (*FuzzSeedInfo, error) {
	info := &FuzzSeedInfo{
		TestName: testName,
		Package:  pkgDir,
	}

	// Count f.Add() calls in source code
	codeSeeds, sourceFile, hasLoop, err := countFAddCalls(pkgDir, testName)
	if err != nil {
		// Non-fatal: we can still check testdata
		utils.Debug("Failed to count f.Add() calls: %v", err)
	}
	info.CodeSeeds = codeSeeds
	info.SourceFile = sourceFile
	info.HasLoopedSeeds = hasLoop

	// Count corpus files in testdata/fuzz/<TestName>/
	corpusSeeds, err := countCorpusFiles(pkgDir, testName)
	if err != nil {
		// Non-fatal: testdata might not exist
		utils.Debug("Failed to count corpus files: %v", err)
	}
	info.CorpusSeeds = corpusSeeds

	info.TotalSeeds = info.CodeSeeds + info.CorpusSeeds

	return info, nil
}

// countFAddCalls parses test files to count f.Add() calls within a specific fuzz test.
// Returns the count, source file path, whether seeds are in a loop, and any error.
func countFAddCalls(pkgDir, testName string) (int, string, bool, error) {
	// Find all *_test.go files in the package directory
	pattern := filepath.Join(pkgDir, "*_test.go")
	testFiles, err := filepath.Glob(pattern)
	if err != nil {
		return 0, "", false, err
	}

	fset := token.NewFileSet()
	count := 0
	sourceFile := ""
	hasLoop := false

	for _, testFile := range testFiles {
		file, parseErr := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
		if parseErr != nil {
			utils.Debug("Failed to parse %s: %v", testFile, parseErr)
			continue
		}

		// Look for the specific fuzz test function
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if funcDecl.Name.Name != testName {
				continue
			}

			// Found the fuzz test function
			sourceFile = testFile

			// Count f.Add() calls and detect if they're in loops
			c, inLoop := countFAddInFunc(funcDecl)
			count += c
			if inLoop {
				hasLoop = true
			}
		}
	}

	return count, sourceFile, hasLoop, nil
}

// countFAddInFunc counts f.Add() calls within a function and detects if any are in loops.
// When f.Add is called in a loop (for, range), the static count is unreliable.
func countFAddInFunc(funcDecl *ast.FuncDecl) (int, bool) {
	if funcDecl.Body == nil {
		return 0, false
	}

	count := 0
	inLoop := false
	loopDepth := 0

	// Track the fuzz parameter name (usually "f")
	fuzzParamName := ""
	if funcDecl.Type.Params != nil && len(funcDecl.Type.Params.List) > 0 {
		// Fuzz test signature: func FuzzXxx(f *testing.F)
		for _, param := range funcDecl.Type.Params.List {
			if len(param.Names) > 0 {
				fuzzParamName = param.Names[0].Name
				break
			}
		}
	}

	if fuzzParamName == "" {
		fuzzParamName = "f" // Default assumption
	}

	// Walk the AST to find f.Add() calls
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.ForStmt, *ast.RangeStmt:
			loopDepth++
			// Continue inspection but track we're in a loop
			return true
		case *ast.CallExpr:
			if isAddCall(node, fuzzParamName) {
				count++
				if loopDepth > 0 {
					inLoop = true
				}
			}
		}
		return true
	})

	// Note: This simple approach doesn't properly handle loop exit,
	// but for our purposes (detecting ANY loop usage), it's sufficient.
	// A more complex solution would use a visitor pattern with proper scope tracking.

	return count, inLoop
}

// isAddCall checks if a call expression is f.Add() where f is the fuzz parameter
func isAddCall(call *ast.CallExpr, fuzzParamName string) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	// Check if it's <fuzzParam>.Add
	ident, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == fuzzParamName && sel.Sel.Name == "Add"
}

// countCorpusFiles counts files in testdata/fuzz/<TestName>/ directory
func countCorpusFiles(pkgDir, testName string) (int, error) {
	corpusDir := filepath.Join(pkgDir, "testdata", "fuzz", testName)

	entries, err := os.ReadDir(corpusDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // No corpus directory is fine
		}
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			count++
		}
	}

	return count, nil
}

// CalculateFuzzTimeout calculates the appropriate timeout for a fuzz test
// based on the fuzz time and estimated seed count.
//
// Formula: timeout = fuzzTime + (seedCount * overheadPerSeed) + buffer
//
// The calculation accounts for:
//   - The actual fuzzing duration (-fuzztime)
//   - Baseline gathering phase (runs all seeds before fuzzing starts)
//   - Safety buffer for setup/teardown overhead
func CalculateFuzzTimeout(fuzzTime time.Duration, seedCount int, cfg FuzzTimingConfig) time.Duration {
	// Base: the actual fuzz time
	timeout := fuzzTime

	// Add baseline gathering overhead
	if seedCount > 0 {
		baselineTime := time.Duration(seedCount) * cfg.BaselineOverheadPerSeed
		timeout += baselineTime
	}

	// Add safety buffer
	timeout += cfg.BaselineBuffer

	// Apply minimum
	if timeout < cfg.MinTimeout {
		timeout = cfg.MinTimeout
	}

	// Apply maximum cap
	if timeout > cfg.MaxTimeout {
		timeout = cfg.MaxTimeout
	}

	return timeout
}

// CalculateFuzzTimeoutWithArgs calculates timeout from command-line args.
// This is a convenience wrapper that parses -fuzztime from args.
func CalculateFuzzTimeoutWithArgs(args []string, seedCount int, cfg FuzzTimingConfig) time.Duration {
	fuzzTime := parseFuzzTimeFromArgs(args)
	return CalculateFuzzTimeout(fuzzTime, seedCount, cfg)
}

// parseFuzzTimeFromArgs extracts -fuzztime duration from command arguments
func parseFuzzTimeFromArgs(args []string) time.Duration {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "-fuzztime" {
			if d, err := time.ParseDuration(args[i+1]); err == nil {
				return d
			}
		}
	}

	// Also check for -fuzztime=value format
	for _, arg := range args {
		if strings.HasPrefix(arg, "-fuzztime=") {
			val := strings.TrimPrefix(arg, "-fuzztime=")
			if d, err := time.ParseDuration(val); err == nil {
				return d
			}
		}
	}

	// Default fuzz time if not specified
	return 10 * time.Second
}

// FuzzTestTiming captures timing information for a fuzz test run
type FuzzTestTiming struct {
	TestName         string
	Package          string
	SeedCount        int
	BaselineStart    time.Time
	BaselineEnd      time.Time
	BaselineDuration time.Duration
	FuzzingStart     time.Time
	FuzzingEnd       time.Time
	FuzzingDuration  time.Duration
	TotalDuration    time.Duration
}

// String returns a human-readable timing breakdown
func (t *FuzzTestTiming) String() string {
	if t.BaselineDuration > 0 {
		return fmt.Sprintf("Baseline: %s (%d seeds) | Fuzzing: %s | Total: %s",
			utils.FormatDuration(t.BaselineDuration),
			t.SeedCount,
			utils.FormatDuration(t.FuzzingDuration),
			utils.FormatDuration(t.TotalDuration))
	}
	return fmt.Sprintf("Total: %s", utils.FormatDuration(t.TotalDuration))
}

// ParseFuzzOutputTiming extracts timing information from fuzz test output.
// It looks for patterns like "gathering baseline coverage: N/M completed"
// to determine when baseline gathering started and ended.
func ParseFuzzOutputTiming(output string, startTime time.Time) *FuzzTestTiming {
	timing := &FuzzTestTiming{}

	// Pattern: "gathering baseline coverage: 0/8 completed"
	// This appears at the start of baseline gathering
	baselineStartPattern := regexp.MustCompile(`gathering baseline coverage:\s*0/(\d+)\s*completed`)
	// Pattern: "fuzz: elapsed:" indicates fuzzing phase has started (baseline complete)
	// Note: Can't use backreference to match "N/N completed" in Go regex, so we use a simpler pattern
	baselineEndPattern := regexp.MustCompile(`fuzz:\s*elapsed:`)

	lines := strings.Split(output, "\n")

	baselineStarted := false
	var baselineEndTime time.Time

	for _, line := range lines {
		if !baselineStarted {
			if matches := baselineStartPattern.FindStringSubmatch(line); matches != nil {
				baselineStarted = true
				timing.BaselineStart = startTime
				// Try to extract total seed count from "0/N"
				// The regex captured it in matches[1]
			}
		} else if baselineEndPattern.MatchString(line) {
			// Baseline completed, fuzzing phase started
			baselineEndTime = time.Now() // Approximate
			timing.BaselineEnd = baselineEndTime
			timing.FuzzingStart = baselineEndTime
			break
		}
	}

	return timing
}

// EstimateFuzzTestDuration provides an estimate of how long a fuzz test will take
// based on seed count and fuzz time. Useful for pre-flight planning.
func EstimateFuzzTestDuration(fuzzTime time.Duration, seedInfo *FuzzSeedInfo, cfg FuzzTimingConfig) time.Duration {
	if seedInfo == nil {
		return fuzzTime + cfg.BaselineBuffer
	}

	baselineEstimate := time.Duration(seedInfo.TotalSeeds) * cfg.BaselineOverheadPerSeed
	return fuzzTime + baselineEstimate + cfg.BaselineBuffer
}

// WarnIfHighSeedCount logs a warning if a fuzz test has an unusually high seed count
// that might cause timeout issues.
func WarnIfHighSeedCount(seedInfo *FuzzSeedInfo, fuzzTime time.Duration, cfg FuzzTimingConfig) {
	if seedInfo == nil {
		return
	}

	// Threshold: if baseline might take longer than fuzz time itself, warn
	baselineEstimate := time.Duration(seedInfo.TotalSeeds) * cfg.BaselineOverheadPerSeed

	if baselineEstimate > fuzzTime {
		utils.Warn("Fuzz test %s has %d seeds - baseline gathering (~%s) may exceed fuzz time (%s)",
			seedInfo.TestName,
			seedInfo.TotalSeeds,
			utils.FormatDuration(baselineEstimate),
			utils.FormatDuration(fuzzTime))
	}

	if seedInfo.HasLoopedSeeds {
		utils.Warn("Fuzz test %s has f.Add() calls in a loop - actual seed count may be higher than detected (%d)",
			seedInfo.TestName,
			seedInfo.CodeSeeds)
	}
}

// FuzzTestDiagnosticInfo contains information about a completed fuzz test run
// used for diagnosing context deadline exceeded failures.
type FuzzTestDiagnosticInfo struct {
	TestName     string        // Name of the fuzz test function
	Package      string        // Package path
	TestErr      error         // Error returned by the test, if any
	TestOutput   string        // Captured test output (may be empty in non-CI mode)
	TestDuration time.Duration // Wall-clock duration of the test run
	FuzzTime     time.Duration // The -fuzztime value used
}

// DiagnoseFuzzContextDeadline detects when a fuzz test failed due to Go's
// internal fuzztime context deadline rather than mage-x's -timeout flag.
//
// Go's fuzzer has two independent timeout systems:
//  1. The -timeout flag (test-level) — controlled by mage-x, correctly calculated
//  2. A fuzztime-internal context — created by Go's fuzzer for worker coordination,
//     tied to -fuzztime. Mage-x cannot control this.
//
// When fuzz test functions perform expensive operations on large fuzzer-generated
// inputs (up to 1MB+), this internal deadline can expire. The fix on the user side
// is adding t.Skip() guards for oversized inputs.
//
// Returns true if the pattern was detected and a diagnostic was emitted.
func DiagnoseFuzzContextDeadline(info FuzzTestDiagnosticInfo) bool {
	if info.TestErr == nil {
		return false
	}

	// Check if "context deadline exceeded" appears in the error or output
	errStr := info.TestErr.Error()
	hasDeadlineErr := strings.Contains(errStr, "context deadline exceeded") ||
		strings.Contains(info.TestOutput, "context deadline exceeded")

	if !hasDeadlineErr {
		return false
	}

	// Check if the test duration is within a tolerance window of fuzztime.
	// If the test failed near the fuzztime boundary, it's likely the internal
	// fuzztime context that expired, not mage-x's -timeout.
	// Window: [fuzzTime - 1s, fuzzTime + 2s]
	lowerBound := info.FuzzTime - 1*time.Second
	if lowerBound < 0 {
		lowerBound = 0
	}
	upperBound := info.FuzzTime + 2*time.Second

	if info.TestDuration < lowerBound || info.TestDuration > upperBound {
		return false
	}

	utils.Warn("Fuzz test %s.%s failed with 'context deadline exceeded' near the fuzztime boundary (%s elapsed, fuzztime=%s)",
		info.Package, info.TestName,
		utils.FormatDuration(info.TestDuration),
		utils.FormatDuration(info.FuzzTime))
	utils.Warn("This is likely caused by Go's internal fuzztime context, not mage-x's -timeout flag")
	utils.Warn("The fuzzer generates inputs up to ~1MB — if your fuzz test does expensive work on large inputs, the internal deadline expires")
	utils.Warn("")
	utils.Warn("SOLUTION: Add input size guards + context timeouts to prevent expensive operations:")
	utils.Warn("")
	utils.Warn("  f.Fuzz(func(t *testing.T, input []byte) {")
	utils.Warn("      // 1. INPUT SIZE GUARD - adjust limit based on operation cost")
	utils.Warn("      if len(input) > 10000 { t.Skipf(\"Input too large: %%d bytes\", len(input)) }")
	utils.Warn("")
	utils.Warn("      // 2. CONTEXT TIMEOUT - prevent expensive operations from hanging")
	utils.Warn("      ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)")
	utils.Warn("      defer cancel()")
	utils.Warn("")
	utils.Warn("      // 3. Your expensive operation here")
	utils.Warn("      processInput(ctx, input)")
	utils.Warn("  })")
	utils.Warn("")
	utils.Warn("RECOMMENDED LIMITS by operation type:")
	utils.Warn("  • Simple validation: 5000-10000 bytes")
	utils.Warn("  • String operations: 2000-5000 bytes")
	utils.Warn("  • JSON/YAML parsing: 3000-5000 bytes, 2-3s timeout")
	utils.Warn("  • Regex operations: 1000 bytes MAX, 2s timeout")
	utils.Warn("  • Multiple expensive operations: 300-500 bytes, 1.5s timeout")

	if !isFuzzStrictDeadline() {
		utils.Info("[INFO] This failure is being tolerated (not counted as a CI failure)")
		utils.Info("[INFO] Set MAGE_X_FUZZ_STRICT_DEADLINE=true to treat this as a real failure")
	}

	return true
}

// isFuzzStrictDeadline returns true if fuzztime context deadline tolerance is disabled.
// When MAGE_X_FUZZ_STRICT_DEADLINE=true, fuzztime boundary failures are treated as real failures.
func isFuzzStrictDeadline() bool {
	v := os.Getenv("MAGE_X_FUZZ_STRICT_DEADLINE")
	return strings.EqualFold(v, "true")
}
