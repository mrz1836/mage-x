// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for err113 compliance
var (
	ErrCodeNotFormatted     = errors.New("code is not properly formatted")
	ErrUnexpectedRunnerType = errors.New("expected SecureCommandRunner")
)

// Format namespace for code formatting tasks
type Format mg.Namespace

// Default runs all formatters
func (Format) Default() error {
	utils.Header("Running Code Formatters")

	// Run formatters in sequence
	formatters := []struct {
		name string
		fn   func() error
	}{
		{"gofmt", Format{}.Gofmt},
		{"gofumpt", Format{}.Fumpt},
		{"goimports", Format{}.Imports},
	}

	for _, formatter := range formatters {
		utils.Info("\nRunning %s...", formatter.name)
		if err := formatter.fn(); err != nil {
			utils.Warn("Failed: %v", err)
			// Continue with other formatters
		}
	}

	utils.Success("\nFormatting complete")
	return nil
}

// Gofmt runs standard Go formatting
func (Format) Gofmt() error {
	utils.Header("Running gofmt")

	// Find all Go files
	files, err := findGoFiles()
	if err != nil {
		return fmt.Errorf("failed to find Go files: %w", err)
	}

	if len(files) == 0 {
		utils.Info("No Go files found")
		return nil
	}

	// Check if any files need formatting
	output, err := GetRunner().RunCmdOutput("gofmt", "-l", ".")
	if err != nil {
		return fmt.Errorf("gofmt check failed: %w", err)
	}

	unformatted := strings.Split(strings.TrimSpace(output), "\n")
	unformatted = filterEmpty(unformatted)

	if len(unformatted) == 0 {
		utils.Success("All files are properly formatted")
		return nil
	}

	// Format files
	utils.Info("Formatting %d files...", len(unformatted))
	if err := GetRunner().RunCmd("gofmt", "-w", "."); err != nil {
		return fmt.Errorf("gofmt failed: %w", err)
	}

	utils.Success("Formatted %d files", len(unformatted))
	return nil
}

// Fumpt runs stricter gofumpt formatting
func (Format) Fumpt() error {
	utils.Header("Running gofumpt")

	// Ensure gofumpt is installed
	if err := ensureGofumpt(); err != nil {
		return err
	}

	// Run gofumpt with extra rules
	utils.Info("Running gofumpt with extra formatting rules...")

	if err := GetRunner().RunCmd("gofumpt", "-w", "-extra", "."); err != nil {
		return fmt.Errorf("gofumpt failed: %w", err)
	}

	utils.Success("gofumpt formatting complete")
	return nil
}

// Imports organizes imports with goimports
func (Format) Imports() error {
	utils.Header("Organizing Imports")

	// Ensure goimports is installed
	if err := ensureGoimports(); err != nil {
		return err
	}

	// Run goimports
	utils.Info("Running goimports...")

	if err := GetRunner().RunCmd("goimports", "-w", "."); err != nil {
		return fmt.Errorf("goimports failed: %w", err)
	}

	utils.Success("Import organization complete")
	return nil
}

// Check verifies formatting without making changes
func (Format) Check() error {
	utils.Header("Checking Code Format")

	issues := []string{}

	// Check gofmt
	output, err := GetRunner().RunCmdOutput("gofmt", "-l", ".")
	if err != nil {
		return fmt.Errorf("gofmt check failed: %w", err)
	}

	if strings.TrimSpace(output) != "" {
		issues = append(issues, "gofmt: files need formatting")
		for _, file := range strings.Split(output, "\n") {
			if file != "" {
				utils.Info("  - %s", file)
			}
		}
	}

	// Check gofumpt if available
	if utils.CommandExists("gofumpt") {
		output, err = GetRunner().RunCmdOutput("gofumpt", "-l", "-extra", ".")
		if err == nil && strings.TrimSpace(output) != "" {
			issues = append(issues, "gofumpt: files need formatting")
		}
	}

	// Check goimports if available
	if utils.CommandExists("goimports") {
		output, err = GetRunner().RunCmdOutput("goimports", "-l", ".")
		if err == nil && strings.TrimSpace(output) != "" {
			issues = append(issues, "goimports: imports need organizing")
		}
	}

	if len(issues) > 0 {
		utils.Error("Format check failed:")
		for _, issue := range issues {
			fmt.Printf("  - %s\n", issue)
		}
		return ErrCodeNotFormatted
	}

	utils.Success("All files are properly formatted")
	return nil
}

// InstallTools installs all formatting tools
func (Format) InstallTools() error {
	utils.Header("Installing Formatting Tools")

	tools := []struct {
		name    string
		pkg     string
		install func() error
	}{
		{"gofumpt", "mvdan.cc/gofumpt@latest", ensureGofumpt},
		{"goimports", "golang.org/x/tools/cmd/goimports@latest", ensureGoimports},
	}

	for _, tool := range tools {
		utils.Info("Installing %s...", tool.name)
		if err := tool.install(); err != nil {
			return fmt.Errorf("failed to install %s: %w", tool.name, err)
		}
	}

	utils.Success("All formatting tools installed")
	return nil
}

// Helper functions

// ensureGofumpt checks if gofumpt is installed with retry logic
func ensureGofumpt() error {
	if utils.CommandExists("gofumpt") {
		return nil
	}

	utils.Info("gofumpt not found, installing with retry logic...")

	config, err := GetConfig()
	if err != nil {
		// Use default config if loading fails
		config = defaultConfig()
	}

	ctx := context.Background()
	maxRetries := config.Download.MaxRetries
	initialDelay := time.Duration(config.Download.InitialDelayMs) * time.Millisecond

	// Get the secure executor with retry capabilities
	runner := GetRunner()
	secureRunner, ok := runner.(*SecureCommandRunner)
	if !ok {
		return fmt.Errorf("%w: got %T", ErrUnexpectedRunnerType, runner)
	}
	executor := secureRunner.executor
	err = executor.ExecuteWithRetry(ctx, maxRetries, initialDelay, "go", "install", "mvdan.cc/gofumpt@latest")
	if err != nil {
		// Try with direct proxy as fallback
		utils.Warn("Installation failed: %v, trying direct proxy...", err)

		env := []string{"GOPROXY=direct"}
		if err := executor.ExecuteWithEnv(ctx, env, "go", "install", "mvdan.cc/gofumpt@latest"); err != nil {
			return fmt.Errorf("failed to install gofumpt after %d retries and fallback: %w", maxRetries, err)
		}
	}

	utils.Success("gofumpt installed successfully")
	return nil
}

// ensureGoimports checks if goimports is installed with retry logic
func ensureGoimports() error {
	if utils.CommandExists("goimports") {
		return nil
	}

	utils.Info("goimports not found, installing with retry logic...")

	config, err := GetConfig()
	if err != nil {
		// Use default config if loading fails
		config = defaultConfig()
	}

	ctx := context.Background()
	maxRetries := config.Download.MaxRetries
	initialDelay := time.Duration(config.Download.InitialDelayMs) * time.Millisecond

	// Get the secure executor with retry capabilities
	runner := GetRunner()
	secureRunner, ok := runner.(*SecureCommandRunner)
	if !ok {
		return fmt.Errorf("%w: got %T", ErrUnexpectedRunnerType, runner)
	}
	executor := secureRunner.executor
	err = executor.ExecuteWithRetry(ctx, maxRetries, initialDelay, "go", "install", "golang.org/x/tools/cmd/goimports@latest")
	if err != nil {
		// Try with direct proxy as fallback
		utils.Warn("Installation failed: %v, trying direct proxy...", err)

		env := []string{"GOPROXY=direct"}
		if err := executor.ExecuteWithEnv(ctx, env, "go", "install", "golang.org/x/tools/cmd/goimports@latest"); err != nil {
			return fmt.Errorf("failed to install goimports after %d retries and fallback: %w", maxRetries, err)
		}
	}

	utils.Success("goimports installed successfully")
	return nil
}

// findGoFiles finds all Go files in the project
func findGoFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor and other common directories
		if info.IsDir() {
			skip := []string{"vendor", ".git", "node_modules", ".idea", ".vscode"}
			for _, s := range skip {
				if info.Name() == s {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check if it's a Go file
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, ".pb.go") {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// filterEmpty removes empty strings from a slice
func filterEmpty(s []string) []string {
	var result []string
	for _, str := range s {
		if str != "" {
			result = append(result, str)
		}
	}
	return result
}

// Additional methods for Format namespace required by tests

// All formats all files
func (Format) All() error {
	utils.Header("Formatting All Files")

	formatter := Format{}

	// Run all formatters
	if err := formatter.Default(); err != nil {
		return fmt.Errorf("formatting failed: %w", err)
	}

	utils.Success("All files formatted")
	return nil
}

// Go formats Go files
func (Format) Go() error {
	runner := GetRunner()
	return runner.RunCmd("gofmt", "-w", ".")
}

// YAML formats YAML files
func (Format) YAML() error {
	utils.Header("Formatting YAML Files")

	// Find YAML files
	yamlFiles, err := GetRunner().RunCmdOutput("find", ".", "-name", "*.yml", "-o", "-name", "*.yaml", "-not", "-path", "./vendor/*")
	if err != nil {
		return fmt.Errorf("failed to find YAML files: %w", err)
	}

	if yamlFiles == "" {
		utils.Info("No YAML files found")
		return nil
	}

	// Check if prettier is available for YAML formatting
	if utils.CommandExists("prettier") {
		utils.Info("Formatting YAML files with prettier...")
		if err := GetRunner().RunCmd("prettier", "--write", "**/*.{yml,yaml}"); err != nil {
			return fmt.Errorf("prettier formatting failed: %w", err)
		}
	} else {
		utils.Info("prettier not found, install with: npm install -g prettier")
		utils.Info("Skipping YAML formatting")
		return nil
	}

	utils.Success("YAML files formatted")
	return nil
}

// Yaml formats YAML files (alias for interface compatibility)
func (Format) Yaml() error {
	formatter := Format{}
	return formatter.YAML()
}

// formatJSONFile formats a single JSON file using prettier or python3
func formatJSONFile(file string) bool {
	utils.Info("Formatting %s", file)

	// Use prettier if available, otherwise python
	if utils.CommandExists("prettier") {
		if err := GetRunner().RunCmd("prettier", "--write", file); err != nil {
			utils.Warn("Failed to format %s with prettier: %v", file, err)
			return false
		}
		return true
	}

	if utils.CommandExists("python3") {
		// Read, format, and write back
		if err := GetRunner().RunCmd("python3", "-m", "json.tool", file, file+".tmp"); err != nil {
			utils.Warn("Failed to format %s: %v", file, err)
			return false
		}
		// Move temp file to original
		if err := os.Rename(file+".tmp", file); err != nil {
			utils.Warn("Failed to replace %s: %v", file, err)
			return false
		}
		return true
	}

	// This should not happen since we check availability upfront
	return false
}

// JSON formats JSON files
func (Format) JSON() error {
	utils.Header("Formatting JSON Files")

	// Check if formatter is available
	if !utils.CommandExists("prettier") && !utils.CommandExists("python3") {
		utils.Warn("No JSON formatter available, install prettier or python3")
		return nil
	}

	// Find JSON files
	jsonFiles, err := GetRunner().RunCmdOutput("find", ".", "-name", "*.json", "-not", "-path", "./vendor/*")
	if err != nil {
		return fmt.Errorf("failed to find JSON files: %w", err)
	}

	if jsonFiles == "" {
		utils.Info("No JSON files found")
		return nil
	}

	// Format JSON files using python's json.tool or prettier
	files := strings.Split(strings.TrimSpace(jsonFiles), "\n")
	formatted := 0

	for _, file := range files {
		if file != "" && formatJSONFile(file) {
			formatted++
		}
	}

	if formatted > 0 {
		utils.Success("Formatted %d JSON files", formatted)
	} else {
		utils.Info("No JSON files formatted")
	}
	return nil
}

// Markdown formats Markdown files
func (Format) Markdown() error {
	utils.Header("Formatting Markdown Files")

	// Find Markdown files
	mdFiles, err := GetRunner().RunCmdOutput("find", ".", "-name", "*.md", "-not", "-path", "./vendor/*")
	if err != nil {
		return fmt.Errorf("failed to find Markdown files: %w", err)
	}

	if mdFiles == "" {
		utils.Info("No Markdown files found")
		return nil
	}

	// Check if prettier is available for Markdown formatting
	if utils.CommandExists("prettier") {
		utils.Info("Formatting Markdown files with prettier...")
		if err := GetRunner().RunCmd("prettier", "--write", "**/*.md"); err != nil {
			return fmt.Errorf("prettier formatting failed: %w", err)
		}
	} else {
		utils.Info("prettier not found, install with: npm install -g prettier")
		utils.Info("Skipping Markdown formatting")
		return nil
	}

	utils.Success("Markdown files formatted")
	return nil
}

// SQL formats SQL files
func (Format) SQL() error {
	utils.Header("Formatting SQL Files")

	// Find SQL files
	sqlFiles, err := GetRunner().RunCmdOutput("find", ".", "-name", "*.sql")
	if err != nil {
		return fmt.Errorf("failed to find SQL files: %w", err)
	}

	if sqlFiles == "" {
		utils.Info("No SQL files found")
		return nil
	}

	// Check if sqlfluff is available for SQL formatting
	if utils.CommandExists("sqlfluff") {
		utils.Info("Formatting SQL files with sqlfluff...")
		if err := GetRunner().RunCmd("sqlfluff", "format", "."); err != nil {
			return fmt.Errorf("sqlfluff formatting failed: %w", err)
		}
	} else {
		utils.Info("sqlfluff not found, install with: pip install sqlfluff")
		utils.Info("Skipping SQL formatting")
		return nil
	}

	utils.Success("SQL files formatted")
	return nil
}

// Dockerfile formats Dockerfile
func (Format) Dockerfile() error {
	utils.Header("Formatting Dockerfile")

	// Check if Dockerfile exists
	if !utils.FileExists("Dockerfile") {
		utils.Info("No Dockerfile found")
		return nil
	}

	// Check if dockerfile_lint is available
	if utils.CommandExists("dockerfile_lint") {
		utils.Info("Linting Dockerfile...")
		if err := GetRunner().RunCmd("dockerfile_lint", "Dockerfile"); err != nil {
			utils.Warn("Dockerfile linting failed: %v", err)
		}
	} else {
		utils.Info("dockerfile_lint not found, install with: npm install -g dockerfile_lint")
	}

	// Basic formatting suggestions
	utils.Info("Dockerfile formatting suggestions:")
	utils.Info("  - Use multi-stage builds when possible")
	utils.Info("  - Minimize layers by combining RUN commands")
	utils.Info("  - Use .dockerignore to exclude unnecessary files")
	utils.Info("  - Sort multi-line arguments alphabetically")

	utils.Success("Dockerfile formatting guidance provided")
	return nil
}

// Shell formats shell scripts
func (Format) Shell() error {
	utils.Header("Formatting Shell Scripts")

	// Find shell script files
	shellFiles, err := GetRunner().RunCmdOutput("find", ".", "-name", "*.sh", "-o", "-name", "*.bash")
	if err != nil {
		return fmt.Errorf("failed to find shell script files: %w", err)
	}

	if shellFiles == "" {
		utils.Info("No shell script files found")
		return nil
	}

	// Check if shfmt is available for shell formatting
	if utils.CommandExists("shfmt") {
		utils.Info("Formatting shell scripts with shfmt...")
		// Format with 2-space indentation and simplified output
		if err := GetRunner().RunCmd("shfmt", "-i", "2", "-w", "."); err != nil {
			return fmt.Errorf("shfmt formatting failed: %w", err)
		}
	} else {
		utils.Info("shfmt not found, install with: go install mvdan.cc/sh/v3/cmd/shfmt@latest")
		utils.Info("Skipping shell script formatting")
		return nil
	}

	utils.Success("Shell scripts formatted")
	return nil
}

// Fix fixes formatting issues
func (Format) Fix() error {
	utils.Header("Fixing Formatting Issues")

	formatter := Format{}

	// Run all formatters to fix issues
	utils.Info("Running Go formatters...")
	if err := formatter.Gofmt(); err != nil {
		utils.Warn("gofmt failed: %v", err)
	}

	if err := formatter.Fumpt(); err != nil {
		utils.Warn("gofumpt failed: %v", err)
	}

	if err := formatter.Imports(); err != nil {
		utils.Warn("goimports failed: %v", err)
	}

	// Run other formatters
	utils.Info("Running other formatters...")
	if err := formatter.JSON(); err != nil {
		utils.Warn("JSON formatting failed: %v", err)
	}

	if err := formatter.YAML(); err != nil {
		utils.Warn("YAML formatting failed: %v", err)
	}

	utils.Success("Formatting fixes completed")
	return nil
}
