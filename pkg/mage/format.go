// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"context"
	"encoding/json"
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
		utils.Info("Running %s...", formatter.name)
		if err := formatter.fn(); err != nil {
			utils.Warn("Failed: %v", err)
			// Continue with other formatters
		}
	}

	utils.Success("Formatting complete")
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
		{"yamlfmt", "github.com/google/yamlfmt/cmd/yamlfmt@" + GetDefaultYamlfmtVersion(), ensureYamlfmt},
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

// ensureYamlfmt checks if yamlfmt is installed with retry logic
func ensureYamlfmt() error {
	if utils.CommandExists("yamlfmt") {
		return nil
	}

	utils.Info("yamlfmt not found, installing with retry logic...")

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
	yamlfmtVersion := GetDefaultYamlfmtVersion()
	if yamlfmtVersion == "" || yamlfmtVersion == VersionLatest {
		yamlfmtVersion = VersionLatest
	}
	installCmd := "github.com/google/yamlfmt/cmd/yamlfmt@" + yamlfmtVersion

	err = executor.ExecuteWithRetry(ctx, maxRetries, initialDelay, "go", "install", installCmd)
	if err != nil {
		// Try with direct proxy as fallback
		utils.Warn("Installation failed: %v, trying direct proxy...", err)

		env := []string{"GOPROXY=direct"}
		if err := executor.ExecuteWithEnv(ctx, env, "go", "install", installCmd); err != nil {
			return fmt.Errorf("failed to install yamlfmt after %d retries and fallback: %w", maxRetries, err)
		}
	}

	utils.Success("yamlfmt installed successfully")
	return nil
}

// getFormatExcludePaths returns the list of paths to exclude from formatting
func getFormatExcludePaths() []string {
	excludePaths := utils.GetEnv("MAGE_X_FORMAT_EXCLUDE_PATHS", "vendor,node_modules,.git,.idea,.vscode")
	return strings.Split(excludePaths, ",")
}

// buildFindExcludeArgs builds the find command arguments for excluding paths
func buildFindExcludeArgs() []string {
	excludePaths := getFormatExcludePaths()
	var args []string
	for _, excludePath := range excludePaths {
		trimmed := strings.TrimSpace(excludePath)
		if trimmed != "" {
			// Add patterns for both direct matches and subdirectories
			args = append(args, "-not", "-path", "./"+trimmed+"/*")
			args = append(args, "-not", "-path", "./*"+trimmed+"*/*")
		}
	}
	return args
}

// findGoFiles finds all Go files in the project
func findGoFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories from environment variable
		if info.IsDir() {
			excludePaths := getFormatExcludePaths()
			for _, excludePath := range excludePaths {
				if info.Name() == strings.TrimSpace(excludePath) {
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
	findArgs := []string{".", "-name", "*.yml", "-o", "-name", "*.yaml"}
	findArgs = append(findArgs, buildFindExcludeArgs()...)
	yamlFiles, err := GetRunner().RunCmdOutput("find", findArgs...)
	if err != nil {
		return fmt.Errorf("failed to find YAML files: %w", err)
	}

	if yamlFiles == "" {
		utils.Info("No YAML files found")
		return nil
	}

	// Ensure yamlfmt is available for YAML formatting
	if err = ensureYamlfmt(); err != nil {
		return fmt.Errorf("failed to ensure yamlfmt is available: %w", err)
	}

	utils.Info("Formatting YAML files with yamlfmt...")

	// Use yamlfmt with config file
	configPath := ".github/.yamlfmt"
	if utils.FileExists(configPath) {
		if err := GetRunner().RunCmd("yamlfmt", "-conf", configPath, "."); err != nil {
			return fmt.Errorf("yamlfmt formatting failed: %w", err)
		}
	} else {
		// Fallback to default yamlfmt behavior
		if err := GetRunner().RunCmd("yamlfmt", "."); err != nil {
			return fmt.Errorf("yamlfmt formatting failed: %w", err)
		}
	}

	utils.Success("YAML files formatted")
	return nil
}

// Yaml formats YAML files (alias for interface compatibility)
func (Format) Yaml() error {
	formatter := Format{}
	return formatter.YAML()
}

// formatJSONFile formats a single JSON file using native Go
func formatJSONFile(file string) bool {
	utils.Info("Formatting %s", file)

	return formatJSONFileNative(file)
}

// formatJSONFileNative formats a single JSON file using Go's standard library
func formatJSONFileNative(file string) bool {
	// Read the JSON file
	data, err := os.ReadFile(file) //nolint:gosec // file path is user-provided via API
	if err != nil {
		utils.Warn("Failed to read %s: %v", file, err)
		return false
	}

	// Parse JSON to validate it and format it
	var jsonData interface{}
	if unmarshalErr := json.Unmarshal(data, &jsonData); unmarshalErr != nil {
		utils.Warn("Invalid JSON in %s: %v", file, unmarshalErr)
		return false
	}

	// Format JSON with 4-space indentation and disabled HTML escaping
	var formatted strings.Builder
	encoder := json.NewEncoder(&formatted)
	encoder.SetIndent("", "    ")
	encoder.SetEscapeHTML(false) // Disable HTML escaping to preserve & characters
	if err := encoder.Encode(jsonData); err != nil {
		utils.Warn("Failed to format JSON in %s: %v", file, err)
		return false
	}

	// Get formatted data as bytes
	formattedData := []byte(formatted.String())

	// Write to temporary file first for atomic operation
	tmpFile := file + ".tmp"
	if err := os.WriteFile(tmpFile, formattedData, 0o600); err != nil {
		utils.Warn("Failed to write temporary file %s: %v", tmpFile, err)
		return false
	}

	// Move temp file to original
	if err := os.Rename(tmpFile, file); err != nil {
		utils.Warn("Failed to replace %s: %v", file, err)
		// Clean up temp file on failure
		_ = os.Remove(tmpFile) //nolint:errcheck // cleanup on failure, ignore error
		return false
	}

	return true
}

// JSON formats JSON files
func (Format) JSON() error {
	utils.Header("Formatting JSON Files")

	// Find JSON files
	findArgs := []string{".", "-name", "*.json"}
	findArgs = append(findArgs, buildFindExcludeArgs()...)
	jsonFiles, err := GetRunner().RunCmdOutput("find", findArgs...)
	if err != nil {
		return fmt.Errorf("failed to find JSON files: %w", err)
	}

	if jsonFiles == "" {
		utils.Info("No JSON files found")
		return nil
	}

	// Format JSON files using native Go
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
