// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Metrics namespace for code metrics and analysis tasks
type Metrics mg.Namespace

// LOC displays lines of code statistics
func (Metrics) LOC() error {
	utils.Header("Lines of Code Statistics")

	// Count lines in test files
	testCount, err := countLines("*_test.go", []string{"vendor", "third_party"})
	if err != nil {
		utils.Warn("Failed to count test files: %v", err)
		testCount = 0
	}

	// Count lines in non-test Go files
	goCount, err := countGoLines([]string{"vendor", "third_party"})
	if err != nil {
		utils.Warn("Failed to count Go files: %v", err)
		goCount = 0
	}

	// Display table
	date := time.Now().Format("2006-01-02")

	fmt.Println()
	fmt.Println("| Type       | Total Lines | Date        |")
	fmt.Println("|------------|-------------|-------------|")
	fmt.Printf("| Test Files | %-11d | %s |\n", testCount, date)
	fmt.Printf("| Go Files   | %-11d | %s |\n", goCount, date)
	fmt.Println()

	total := testCount + goCount
	utils.Success("Total lines of code: %d", total)

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
	utils.Info("\nPackage coverage:")
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
		// This is expected behavior
	}

	utils.Info("\nTop 10 most complex functions:")
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
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
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

	utils.Info("Binary size: %s", formatBytes(stat.Size()))

	// Analyze binary sections
	utils.Info("\nBinary composition:")
	if err := GetRunner().RunCmd("go", "tool", "nm", "-size", binaryName); err != nil {
		utils.Warn("Failed to analyze binary sections")
	}

	// Check module size
	utils.Info("\nModule dependencies size:")
	if modCache := os.Getenv("GOMODCACHE"); modCache != "" {
		if info, err := getDirSize(modCache); err == nil {
			utils.Info("Module cache size: %s", formatBytes(info))
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
		{"Lines of Code", Metrics{}.LOC},
		{"Test Coverage", Metrics{}.Coverage},
		{"Complexity", Metrics{}.Complexity},
	}

	failed := 0
	for _, check := range checks {
		utils.Info("\nRunning %s...", check.name)
		if err := check.fn(); err != nil {
			utils.Error("Failed: %v", err)
			failed++
		}
	}

	if failed > 0 {
		return fmt.Errorf("%d quality checks failed", failed)
	}

	utils.Success("\nAll quality checks passed!")
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

// countLines counts lines in files matching pattern
func countLines(pattern string, excludeDirs []string) (int, error) {
	count := 0

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
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
			// Log error but continue to check other files
			utils.Debug("line count: failed to match pattern for %s: %v", path, err)
			return nil
		}
		if !matched {
			return nil
		}

		// Count lines
		fileOps := fileops.New()
		content, err := fileOps.File.ReadFile(path)
		if err != nil {
			// Log error but continue to count other files
			utils.Debug("line count: failed to read file %s: %v", path, err)
			return nil
		}

		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			// Skip empty lines and full-line comments
			if line != "" && !strings.HasPrefix(line, "//") {
				count++
			}
		}

		return nil
	})

	return count, err
}

// countGoLines counts lines in non-test Go files
func countGoLines(excludeDirs []string) (int, error) {
	count := 0

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
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

		// Count lines
		fileOps := fileops.New()
		content, err := fileOps.File.ReadFile(path)
		if err != nil {
			// Log error but continue to count other files
			utils.Debug("line count: failed to read file %s: %v", path, err)
			return nil
		}

		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			// Skip empty lines and full-line comments
			if line != "" && !strings.HasPrefix(line, "//") {
				count++
			}
		}

		return nil
	})

	return count, err
}
