// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"

	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for metrics operations
var (
	errQualityChecksFailed = errors.New("quality checks failed")
)

// Metrics namespace for code metrics and analysis tasks
type Metrics mg.Namespace

// LOCResult represents JSON output for LOC statistics
type LOCResult struct {
	// EXISTING FIELDS - DO NOT MODIFY OR REMOVE (fortress workflow dependency)
	TestFilesLOC    int      `json:"test_files_loc"`
	TestFilesCount  int      `json:"test_files_count"`
	GoFilesLOC      int      `json:"go_files_loc"`
	GoFilesCount    int      `json:"go_files_count"`
	TotalLOC        int      `json:"total_loc"`
	TotalFilesCount int      `json:"total_files_count"`
	Date            string   `json:"date"`
	ExcludedDirs    []string `json:"excluded_dirs"`

	// NEW FIELDS - ADDITIVE ONLY
	TestFilesSizeBytes  int64   `json:"test_files_size_bytes"`
	TestFilesSizeHuman  string  `json:"test_files_size_human"`
	GoFilesSizeBytes    int64   `json:"go_files_size_bytes"`
	GoFilesSizeHuman    string  `json:"go_files_size_human"`
	TotalSizeBytes      int64   `json:"total_size_bytes"`
	TotalSizeHuman      string  `json:"total_size_human"`
	AvgLinesPerFile     float64 `json:"avg_lines_per_file"`
	TestCoverageRatio   float64 `json:"test_coverage_ratio"`
	PackageCount        int     `json:"package_count"`
	TestAvgLinesPerFile float64 `json:"test_avg_lines_per_file"`
	GoAvgLinesPerFile   float64 `json:"go_avg_lines_per_file"`
	TestAvgSizeBytes    int64   `json:"test_avg_size_bytes"`
	GoAvgSizeBytes      int64   `json:"go_avg_size_bytes"`
}

// LOCStats holds line, file counts, and total size
type LOCStats struct {
	Lines      int
	Files      int
	TotalBytes int64 // Total size in bytes
}

// LOC displays lines of code statistics (use json for JSON output)
func (Metrics) LOC(args ...string) error {
	// Parse command-line parameters
	params := utils.ParseParams(args)
	jsonOutput := utils.IsParamTrue(params, "json")

	excludeDirs := []string{"vendor", "third_party"}

	// Count lines and files in test files
	testStats, err := countLinesWithStats("*_test.go", excludeDirs)
	if err != nil {
		if !jsonOutput {
			utils.Warn("Failed to count test files: %v", err)
		}
		testStats = LOCStats{}
	}

	// Count lines and files in non-test Go files
	goStats, err := countGoLinesWithStats(excludeDirs)
	if err != nil {
		if !jsonOutput {
			utils.Warn("Failed to count Go files: %v", err)
		}
		goStats = LOCStats{}
	}

	date := time.Now().Format("2006-01-02")
	totalLOC := testStats.Lines + goStats.Lines
	totalFiles := testStats.Files + goStats.Files

	// Count packages
	packageCount, err := countPackages(excludeDirs)
	if err != nil {
		if !jsonOutput {
			utils.Warn("Failed to count packages: %v", err)
		}
		packageCount = 0
	}

	// Calculate derived metrics
	totalBytes := testStats.TotalBytes + goStats.TotalBytes
	avgLinesPerFile := safeAverage(totalLOC, totalFiles)
	testCoverageRatio := safeAverage(testStats.Lines, goStats.Lines) * 100
	testAvgLines := safeAverage(testStats.Lines, testStats.Files)
	goAvgLines := safeAverage(goStats.Lines, goStats.Files)
	testAvgBytes := safeAverageBytes(testStats.TotalBytes, testStats.Files)
	goAvgBytes := safeAverageBytes(goStats.TotalBytes, goStats.Files)

	if jsonOutput {
		// JSON output - no headers, no success messages, just JSON
		result := LOCResult{
			// EXISTING FIELDS - UNCHANGED
			TestFilesLOC:    testStats.Lines,
			TestFilesCount:  testStats.Files,
			GoFilesLOC:      goStats.Lines,
			GoFilesCount:    goStats.Files,
			TotalLOC:        totalLOC,
			TotalFilesCount: totalFiles,
			Date:            date,
			ExcludedDirs:    excludeDirs,

			// NEW FIELDS
			TestFilesSizeBytes:  testStats.TotalBytes,
			TestFilesSizeHuman:  formatBytesMetrics(testStats.TotalBytes),
			GoFilesSizeBytes:    goStats.TotalBytes,
			GoFilesSizeHuman:    formatBytesMetrics(goStats.TotalBytes),
			TotalSizeBytes:      totalBytes,
			TotalSizeHuman:      formatBytesMetrics(totalBytes),
			AvgLinesPerFile:     avgLinesPerFile,
			TestCoverageRatio:   testCoverageRatio,
			PackageCount:        packageCount,
			TestAvgLinesPerFile: testAvgLines,
			GoAvgLinesPerFile:   goAvgLines,
			TestAvgSizeBytes:    testAvgBytes,
			GoAvgSizeBytes:      goAvgBytes,
		}

		jsonBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		utils.Println(string(jsonBytes))
		return nil
	}

	// Default markdown table output
	utils.Header("Lines of Code Statistics")

	utils.Println("")
	utils.Println("| Type       | Total Lines | File Count | Total Size | Avg Size    | Date       |")
	utils.Println("|------------|-------------|------------|------------|-------------|------------|")
	utils.Print("| Test Files | %-11s | %-10d | %-10s | %-11s | %s |\n",
		formatNumberWithCommas(testStats.Lines),
		testStats.Files,
		formatBytesMetrics(testStats.TotalBytes),
		formatBytesMetrics(testAvgBytes),
		date)
	utils.Print("| Go Files   | %-11s | %-10d | %-10s | %-11s | %s |\n",
		formatNumberWithCommas(goStats.Lines),
		goStats.Files,
		formatBytesMetrics(goStats.TotalBytes),
		formatBytesMetrics(goAvgBytes),
		date)
	utils.Println("")

	// Summary section - use markdown table for GitHub compatibility
	utils.Println("")
	utils.Println("Summary")
	utils.Println("")
	utils.Println("| Metric                  | Value                                 |")
	utils.Println("|-------------------------|---------------------------------------|")
	utils.Print("| Total Lines of Code     | %-37s |\n", formatNumberWithCommas(totalLOC))
	utils.Print("| Total Files             | %-37d |\n", totalFiles)
	utils.Print("| Total Size              | %-37s |\n", formatBytesMetrics(totalBytes))
	utils.Print("| Package/Directory Count | %-37d |\n", packageCount)
	utils.Print("| Average Lines per File  | %-37.1f |\n", avgLinesPerFile)
	utils.Print("| Test Coverage Ratio     | %-37s |\n", fmt.Sprintf("%.1f%% (test LOC / production LOC)", testCoverageRatio))
	utils.Println("")

	utils.Success("Analysis complete!")

	return nil
}

// Mage scans for magefiles and reports found targets
func (Metrics) Mage() error {
	utils.Header("Magefile Analysis")

	// Check for magefiles directory
	magefilesDir := "magefiles"
	hasMagefilesDir := false
	if info, err := os.Stat(magefilesDir); err == nil && info.IsDir() {
		hasMagefilesDir = true
	}

	utils.Println("")
	utils.Print("Magefiles Directory Found: ")
	if hasMagefilesDir {
		utils.Success("Yes")
	} else {
		utils.Warn("No")
	}
	utils.Println("")

	// Find magefile sources
	type mageFile struct {
		Path      string
		Functions int
		Targets   []string
	}
	var files []mageFile
	totalFuncs := 0

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk path %s: %w", path, err)
		}

		// Skip hidden directories and common vendor dirs
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
				return filepath.SkipDir
			}
			if info.Name() == "vendor" || info.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Check if file is a magefile
		isMage := strings.HasPrefix(path, magefilesDir+string(filepath.Separator))

		// 1. Check strict location (magefiles directory)

		// 2. Check build tags if not already confirmed
		// Also parse to get function counts regardless
		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			utils.Debug("Failed to parse file %s: %v", path, err)
			return nil
		}

		if !isMage {
			// Check build tags in comments
			for _, comment := range node.Comments {
				if strings.Contains(comment.Text(), "+build mage") ||
					strings.Contains(comment.Text(), "go:build mage") {
					isMage = true
					break
				}
			}
		}

		if isMage {
			funcCount := 0
			var targets []string

			for _, decl := range node.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					if fn.Name.IsExported() {
						funcCount++
						targets = append(targets, fn.Name.Name)
					}
				}
			}

			if funcCount > 0 || isMage {
				files = append(files, mageFile{
					Path:      path,
					Functions: funcCount,
					Targets:   targets,
				})
				totalFuncs += funcCount
			}
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to scan for magefiles: %w", err)
	}

	// Display results
	if len(files) > 0 {
		utils.Println("| File Path                                        | Functions |")
		utils.Println("|--------------------------------------------------|-----------|")
		for _, f := range files {
			// Truncate path if too long
			displayPath := f.Path
			if len(displayPath) > 48 {
				displayPath = "..." + displayPath[len(displayPath)-45:]
			}
			utils.Print("| %-48s | %-9d |\n", displayPath, f.Functions)
		}
		utils.Println("")
		utils.Success("Total functions found: %d", totalFuncs)
	} else {
		utils.Warn("No magefiles or mage targets found.")
	}

	return nil
}

// Coverage analyzes test coverage across the codebase
func (Metrics) Coverage() error {
	utils.Header("Test Coverage Analysis")

	// Run tests with coverage
	utils.Info("Generating coverage data...")
	if err := GetRunner().RunCmd("go", "test", "-coverprofile=coverage.tmp", "./..."); err != nil {
		return fmt.Errorf("failed to generate coverage: %w", err)
	}

	// Get coverage percentage
	output, err := GetRunner().RunCmdOutput("go", "tool", "cover", "-func=coverage.tmp")
	if err != nil {
		return fmt.Errorf("failed to analyze coverage: %w", err)
	}

	// Parse total coverage
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "total:") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				utils.Success("Total coverage: %s", fields[len(fields)-1])
				break
			}
		}
	}

	// Clean up temp file
	if err := os.Remove("coverage.tmp"); err != nil {
		// Log but don't fail - this is cleanup
		utils.Debug("Failed to remove coverage temp file: %v", err)
	}

	// Show detailed coverage by package
	utils.Info("Package coverage:")
	if err := GetRunner().RunCmd("go", "test", "-cover", "./..."); err != nil {
		utils.Warn("Failed to show package coverage")
	}

	return nil
}

// Complexity analyzes code complexity
func (Metrics) Complexity() error {
	utils.Header("Code Complexity Analysis")

	// Check if gocyclo is installed
	if !utils.CommandExists("gocyclo") {
		utils.Info("Installing gocyclo...")
		if err := GetRunner().RunCmd("go", "install", "github.com/fzipp/gocyclo/cmd/gocyclo@latest"); err != nil {
			return fmt.Errorf("failed to install gocyclo: %w", err)
		}
	}

	// Run complexity analysis
	utils.Info("Analyzing cyclomatic complexity...")
	utils.Info("Functions with complexity > 10:")

	if err := GetRunner().RunCmd("gocyclo", "-over", "10", "."); err != nil {
		// gocyclo returns error if it finds complex functions
		// This is expected behavior, we continue to show results
		utils.Debug("%s", "gocyclo found complex functions (expected): "+err.Error())
	}

	utils.Info("Top 10 most complex functions:")
	if err := GetRunner().RunCmd("gocyclo", "-top", "10", "."); err != nil {
		return fmt.Errorf("failed to analyze complexity: %w", err)
	}

	return nil
}

// Size analyzes binary and module sizes
func (Metrics) Size() error {
	utils.Header("Size Analysis")

	// Build binary to check size
	utils.Info("Building binary...")
	binaryName := "temp-size-check"
	if runtime.GOOS == OSWindows {
		binaryName += windowsExeExt
	}

	if err := GetRunner().RunCmd("go", "build", "-o", binaryName); err != nil {
		return fmt.Errorf("failed to build binary: %w", err)
	}
	defer func() {
		if err := os.Remove(binaryName); err != nil {
			// Log but don't fail - this is cleanup
			utils.Debug("Failed to remove binary %s: %v", binaryName, err)
		}
	}()

	// Get binary size
	stat, err := os.Stat(binaryName)
	if err != nil {
		return fmt.Errorf("failed to stat binary: %w", err)
	}

	utils.Info("Binary size: %s", formatBytesMetrics(stat.Size()))

	// Analyze binary sections
	utils.Info("Binary composition:")
	if err := GetRunner().RunCmd("go", "tool", "nm", "-size", binaryName); err != nil {
		utils.Warn("Failed to analyze binary sections")
	}

	// Check module size
	utils.Info("Module dependencies size:")
	if modCache := os.Getenv("GOMODCACHE"); modCache != "" {
		if info, err := getDirSize(modCache); err == nil {
			utils.Info("Module cache size: %s", formatBytesMetrics(info))
		}
	}

	return nil
}

// Quality runs various code quality metrics
func (Metrics) Quality() error {
	utils.Header("Code Quality Metrics")

	// Run multiple quality checks
	checks := []struct {
		name string
		fn   func() error
	}{
		{"Lines of Code", func() error { return Metrics{}.LOC() }},
		{"Test Coverage", Metrics{}.Coverage},
		{"Complexity", Metrics{}.Complexity},
	}

	failed := 0
	for _, check := range checks {
		utils.Info("Running %s...", check.name)
		if err := check.fn(); err != nil {
			utils.Error("Failed: %v", err)
			failed++
		}
	}

	if failed > 0 {
		return fmt.Errorf("%w: %d checks", errQualityChecksFailed, failed)
	}

	utils.Success("All quality checks passed!")
	return nil
}

// Imports analyzes import dependencies
func (Metrics) Imports() error {
	utils.Header("Import Analysis")

	// Get all imports
	output, err := GetRunner().RunCmdOutput("go", "list", "-f", "{{join .Imports \"\\n\"}}", "./...")
	if err != nil {
		return fmt.Errorf("failed to list imports: %w", err)
	}

	// Count unique imports
	imports := make(map[string]int)
	for _, line := range strings.Split(output, "\n") {
		imp := strings.TrimSpace(line)
		if imp != "" {
			imports[imp]++
		}
	}

	// Categorize imports
	var stdlib, internal, external []string
	for imp := range imports {
		if !strings.Contains(imp, ".") {
			stdlib = append(stdlib, imp)
		} else if module, err := getModuleName(); err == nil && strings.HasPrefix(imp, module) {
			internal = append(internal, imp)
		} else {
			external = append(external, imp)
		}
	}

	// Display statistics
	fmt.Printf("\nImport Statistics:\n")
	fmt.Printf("  Standard library: %d packages\n", len(stdlib))
	fmt.Printf("  Internal:         %d packages\n", len(internal))
	fmt.Printf("  External:         %d packages\n", len(external))
	fmt.Printf("  Total unique:     %d packages\n", len(imports))

	// Show most used imports
	fmt.Printf("\nMost frequently imported:\n")
	for imp, count := range imports {
		if count > 3 {
			fmt.Printf("  %s: %d times\n", imp, count)
		}
	}

	return nil
}

// Helper functions

// formatNumberWithCommas formats an integer with thousand separators
func formatNumberWithCommas(n int) string {
	str := fmt.Sprintf("%d", n)
	if len(str) <= 3 {
		return str
	}

	result := ""
	for i, digit := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(digit)
	}
	return result
}

// countLinesWithStats counts lines and files matching pattern
//
//nolint:unparam // pattern is always "*_test.go" in production but varies in tests
func countLinesWithStats(pattern string, excludeDirs []string) (LOCStats, error) {
	stats := LOCStats{}

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk path %s: %w", path, err)
		}

		// Skip excluded directories
		for _, exclude := range excludeDirs {
			if strings.Contains(path, exclude) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Check if file matches pattern
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			utils.Debug("line count: failed to match pattern for %s: %v", path, err)
			return nil
		}
		if !matched {
			return nil
		}

		// Count this file
		stats.Files++
		stats.TotalBytes += info.Size()

		// Count lines
		fileOps := fileops.New()
		content, err := fileOps.File.ReadFile(path)
		if err != nil {
			utils.Debug("line count: failed to read file %s: %v", path, err)
			return nil
		}

		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "//") {
				stats.Lines++
			}
		}

		return nil
	})

	return stats, err
}

// countGoLinesWithStats counts lines and files in non-test Go files
func countGoLinesWithStats(excludeDirs []string) (LOCStats, error) {
	stats := LOCStats{}

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk path %s: %w", path, err)
		}

		// Skip excluded directories
		for _, exclude := range excludeDirs {
			if strings.Contains(path, exclude) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// Check if it's a Go file (not test)
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Count this file
		stats.Files++
		stats.TotalBytes += info.Size()

		// Count lines
		fileOps := fileops.New()
		content, err := fileOps.File.ReadFile(path)
		if err != nil {
			utils.Debug("line count: failed to read file %s: %v", path, err)
			return nil
		}

		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "//") {
				stats.Lines++
			}
		}

		return nil
	})

	return stats, err
}

// countPackages counts unique Go packages (directories with .go files)
func countPackages(excludeDirs []string) (int, error) {
	packages := make(map[string]bool)

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk path %s: %w", path, err)
		}

		// Skip excluded directories
		for _, exclude := range excludeDirs {
			if strings.Contains(path, exclude) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		// If it's a .go file, mark its directory as a package
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			dir := filepath.Dir(path)
			packages[dir] = true
		}

		return nil
	})

	return len(packages), err
}

// safeAverage calculates average safely, returning 0 if denominator is 0
func safeAverage(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0.0
	}
	return float64(numerator) / float64(denominator)
}

// safeAverageBytes calculates average bytes safely
func safeAverageBytes(totalBytes int64, count int) int64 {
	if count == 0 {
		return 0
	}
	return totalBytes / int64(count)
}

// formatBytesMetrics formats bytes as human-readable size string
func formatBytesMetrics(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
