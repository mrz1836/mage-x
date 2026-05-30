// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"

	"github.com/mrz1836/mage-x/pkg/common/env"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
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
		{"gci", Format{}.Gci},
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

	unformatted := utils.ParseNonEmptyLines(output)

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
		return fmt.Errorf("failed to ensure gofumpt is installed: %w", err)
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
		return fmt.Errorf("failed to ensure goimports is installed: %w", err)
	}

	// Run goimports
	utils.Info("Running goimports...")

	if err := GetRunner().RunCmd("goimports", "-w", "."); err != nil {
		return fmt.Errorf("goimports failed: %w", err)
	}

	utils.Success("Import organization complete")
	return nil
}

// Gci organizes imports with gci (Go import organizer)
func (Format) Gci() error {
	utils.Header("Organizing Imports with gci")

	// Ensure gci is installed
	if err := ensureGci(); err != nil {
		return fmt.Errorf("failed to ensure gci is installed: %w", err)
	}

	// Build gci arguments
	args := []string{"write", "--skip-generated"}

	// Try to read gci configuration from .golangci.json
	gciSections := getGciSectionsFromConfig()
	if len(gciSections) > 0 {
		utils.Info("Using gci sections from .golangci.json: %v", gciSections)
		for _, section := range gciSections {
			args = append(args, "-s", section)
		}
	} else {
		// Default sections if no config found
		utils.Info("Using default gci sections")
		args = append(args, "-s", "standard", "-s", "default")
	}

	// Add current directory
	args = append(args, ".")

	// Run gci
	utils.Info("Running gci...")
	if err := GetRunner().RunCmd("gci", args...); err != nil {
		return fmt.Errorf("gci failed: %w", err)
	}

	utils.Success("Import organization with gci complete")
	return nil
}

// getGciSectionsFromConfig reads gci sections from .golangci.json
// and ensures required sections (standard, default) are present
func getGciSectionsFromConfig() []string {
	configPath := ".golangci.json"
	if !utils.FileExists(configPath) {
		return nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		utils.Debug("Could not read gci config from %s: %v", configPath, err)
		return nil
	}

	var config struct {
		Formatters struct {
			Settings struct {
				Gci struct {
					Sections []string `json:"sections"`
				} `json:"gci"`
			} `json:"settings"`
		} `json:"formatters"`
	}

	if err := json.Unmarshal(data, &config); err != nil {
		utils.Debug("Could not parse gci config from %s: %v", configPath, err)
		return nil
	}

	sections := config.Formatters.Settings.Gci.Sections
	if len(sections) == 0 {
		return nil
	}

	// Check which required sections are already present
	hasStandard := false
	hasDefault := false
	for _, section := range sections {
		if section == "standard" {
			hasStandard = true
		}
		if section == "default" {
			hasDefault = true
		}
	}

	// Build complete section list by prepending missing required sections
	// This ensures gci can categorize all import types properly
	var result []string
	if !hasStandard {
		result = append(result, "standard")
	}
	if !hasDefault {
		result = append(result, "default")
	}
	result = append(result, sections...)

	return result
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
		{"gci", "github.com/daixiang0/gci@latest", ensureGci},
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

// ensureGofumpt checks if gofumpt is installed, installing it if needed
func ensureGofumpt() error {
	return installTool(ToolDefinition{
		Name:   "gofumpt",
		Module: "mvdan.cc/gofumpt",
		Check:  "gofumpt",
	})
}

// ensureGoimports checks if goimports is installed, installing it if needed
func ensureGoimports() error {
	return installTool(ToolDefinition{
		Name:   "goimports",
		Module: "golang.org/x/tools/cmd/goimports",
		Check:  "goimports",
	})
}

// ensureGci checks if gci is installed, installing it if needed
func ensureGci() error {
	return installTool(ToolDefinition{
		Name:   "gci",
		Module: "github.com/daixiang0/gci",
		Check:  "gci",
	})
}

// ensureYamlfmt checks if yamlfmt is installed, installing it if needed
func ensureYamlfmt() error {
	return installTool(ToolDefinition{
		Name:   "yamlfmt",
		Module: "github.com/google/yamlfmt/cmd/yamlfmt",
		Check:  "yamlfmt",
	})
}

// getFormatExcludePaths returns the list of paths to exclude from formatting
func getFormatExcludePaths() []string {
	excludePaths := env.GetString("MAGE_X_FORMAT_EXCLUDE_PATHS", "vendor,node_modules,.git,.idea,.vscode")
	return strings.Split(excludePaths, ",")
}

// newExcludeDirPredicate returns a predicate reporting whether a directory should be
// skipped during a format walk, based on MAGE_X_FORMAT_EXCLUDE_PATHS. A directory is
// skipped when its base name exactly matches one of the (trimmed, non-empty) exclude
// entries, at any depth. This mirrors findGoFiles' historical behavior and unifies
// exclusion semantics across Go, YAML, and JSON discovery so a directory like ci-tester
// is pruned regardless of file extension.
func newExcludeDirPredicate() func(name string) bool {
	excludePaths := getFormatExcludePaths()
	excluded := make(map[string]struct{}, len(excludePaths))
	for _, excludePath := range excludePaths {
		if trimmed := strings.TrimSpace(excludePath); trimmed != "" {
			excluded[trimmed] = struct{}{}
		}
	}
	return func(name string) bool {
		_, ok := excluded[name]
		return ok
	}
}

// findFilesByExt walks the current directory tree collecting files whose extension
// (case-insensitive) is in exts, skipping any directory for which skipDir(baseName)
// returns true. Paths are returned in filepath.Walk's deterministic lexical order.
// Each entry of exts must include the leading dot, e.g. ".yml".
func findFilesByExt(exts []string, skipDir func(name string) bool) ([]string, error) {
	var files []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("failed to walk path %s: %w", path, err)
		}

		if info.IsDir() {
			if skipDir != nil && skipDir(info.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		for _, want := range exts {
			if ext == want {
				files = append(files, path)
				break
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find files: %w", err)
	}

	return files, nil
}

// findGoFiles finds all Go files in the project, excluding generated protobuf files.
func findGoFiles() ([]string, error) {
	candidates, err := findFilesByExt([]string{".go"}, newExcludeDirPredicate())
	if err != nil {
		return nil, fmt.Errorf("failed to find Go files: %w", err)
	}

	var files []string
	for _, file := range candidates {
		if !strings.HasSuffix(file, ".pb.go") {
			files = append(files, file)
		}
	}

	return files, nil
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

// validateYAMLFiles checks YAML files for potential issues before formatting
// Returns lists of safe and problematic files
func validateYAMLFiles(files []string, maxLineLen int) (safeFiles, problematicFiles []string) {
	for _, file := range files {
		hasLongLines, lineNum, lineLen, err := utils.CheckFileLineLength(file, maxLineLen)
		if err != nil {
			utils.Warn("Error checking %s: %v (skipping)", file, err)
			problematicFiles = append(problematicFiles, fmt.Sprintf("%s (error: %v)", file, err))
			continue
		}

		if hasLongLines {
			utils.Warn("Skipping %s: line %d exceeds safe length (%s bytes)",
				file, lineNum, utils.FormatBytes(int64(lineLen)))
			problematicFiles = append(problematicFiles,
				fmt.Sprintf("%s (line %d: %s)", file, lineNum, utils.FormatBytes(int64(lineLen))))
		} else {
			safeFiles = append(safeFiles, file)
		}
	}
	return safeFiles, problematicFiles
}

// formatYAMLFile formats a single YAML file with yamlfmt
func formatYAMLFile(file, configPath string) error {
	if utils.FileExists(configPath) {
		return GetRunner().RunCmd("yamlfmt", "-conf", configPath, file)
	}
	return GetRunner().RunCmd("yamlfmt", file)
}

// formatYAMLFilesIndividually formats each YAML file individually
func formatYAMLFilesIndividually(files []string, configPath string) error {
	var lastErr error
	for _, file := range files {
		if err := formatYAMLFile(file, configPath); err != nil {
			utils.Warn("Failed to format %s: %v", file, err)
			lastErr = err
		}
	}
	return lastErr
}

// maxYAMLArgBytes caps the cumulative bytes of file-path arguments passed to a single
// yamlfmt invocation, staying well under the OS argument limit (macOS ARG_MAX is 262144)
// to avoid E2BIG while leaving room for the config flag and environment.
const maxYAMLArgBytes = 100_000

// chunkByArgBytes splits files into chunks whose joined path bytes stay within budget.
// Each chunk holds at least one file, so a single over-budget path becomes its own chunk
// rather than being dropped. Returns nil for an empty input.
func chunkByArgBytes(files []string, budget int) [][]string {
	if len(files) == 0 {
		return nil
	}

	var (
		chunks       [][]string
		current      []string
		currentBytes int
	)
	for _, file := range files {
		cost := len(file) + 1 // +1 for the inter-argument separator
		if len(current) > 0 && currentBytes+cost > budget {
			chunks = append(chunks, current)
			current = nil
			currentBytes = 0
		}
		current = append(current, file)
		currentBytes += cost
	}
	if len(current) > 0 {
		chunks = append(chunks, current)
	}

	return chunks
}

// formatYAMLFilesBatched formats the given files with yamlfmt in as few invocations as
// possible, chunked to respect the OS argument-length limit. yamlfmt is all-or-nothing
// per invocation (a single unparseable file fails the whole call and leaves the rest
// unformatted), so callers should fall back to per-file formatting on error to isolate
// the offending file and still format everything else.
func formatYAMLFilesBatched(files []string, configPath string) error {
	hasConfig := utils.FileExists(configPath)
	for _, chunk := range chunkByArgBytes(files, maxYAMLArgBytes) {
		args := make([]string, 0, len(chunk)+2)
		if hasConfig {
			args = append(args, "-conf", configPath)
		}
		args = append(args, chunk...)
		if err := GetRunner().RunCmd("yamlfmt", args...); err != nil {
			return err
		}
	}
	return nil
}

// YAML formats YAML files
func (Format) YAML() error {
	utils.Header("Formatting YAML Files")

	// Find YAML files with a native walk so directory exclusions
	// (MAGE_X_FORMAT_EXCLUDE_PATHS) apply uniformly to .yml and .yaml.
	yamlFiles, err := findFilesByExt([]string{".yml", ".yaml"}, newExcludeDirPredicate())
	if err != nil {
		return fmt.Errorf("failed to find YAML files: %w", err)
	}

	if len(yamlFiles) == 0 {
		utils.Info("No YAML files found")
		return nil
	}

	utils.Info("Found %d YAML files", len(yamlFiles))

	// Ensure yamlfmt is available for YAML formatting
	if err = ensureYamlfmt(); err != nil {
		return fmt.Errorf("failed to ensure yamlfmt is available: %w", err)
	}

	// Check if validation is enabled (default: true)
	validationEnabled := env.GetBool(EnvYAMLValidation, true)

	var safeFiles, problematicFiles []string

	if validationEnabled {
		utils.Info("Pre-validating YAML files for line length issues...")
		safeFiles, problematicFiles = validateYAMLFiles(yamlFiles, MaxYAMLLineLength)

		if len(problematicFiles) > 0 {
			utils.Warn("Skipped %d files with line length issues:", len(problematicFiles))
			for _, file := range problematicFiles {
				utils.Warn("  - %s", file)
			}
			utils.Info("Suggestion: Split long lines, use yamlfmt exclude list, or set %s=false", EnvYAMLValidation)
		}

		if len(safeFiles) == 0 {
			utils.Warn("No safe YAML files to format")
			return nil
		}

		utils.Info("Formatting %d safe YAML files...", len(safeFiles))
	} else {
		utils.Info("YAML validation disabled - formatting all files with yamlfmt...")
		safeFiles = yamlFiles
	}

	// Use yamlfmt with config file when present.
	configPath := ".github/.yamlfmt"

	// Format the explicit safe-file list in batched invocations. Because yamlfmt is
	// all-or-nothing per call, fall back to per-file formatting on failure so a single
	// unparseable file is isolated and everything else still gets formatted.
	yamlfmtErr := formatYAMLFilesBatched(safeFiles, configPath)
	if yamlfmtErr != nil {
		utils.Warn("Batch yamlfmt run failed (%v); retrying file-by-file to isolate unparseable files", yamlfmtErr)
		yamlfmtErr = formatYAMLFilesIndividually(safeFiles, configPath)
	}

	if yamlfmtErr != nil {
		return fmt.Errorf("yamlfmt formatting failed: %w", yamlfmtErr)
	}

	if len(problematicFiles) > 0 {
		utils.Success("YAML files formatted (%d formatted, %d skipped)", len(safeFiles), len(problematicFiles))
	} else {
		utils.Success("YAML files formatted (%d files)", len(safeFiles))
	}

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
	var jsonData any
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
	if err := os.WriteFile(tmpFile, formattedData, fileops.PermFileSensitive); err != nil {
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

	// Find JSON files with a native walk that honors MAGE_X_FORMAT_EXCLUDE_PATHS.
	files, err := findFilesByExt([]string{".json"}, newExcludeDirPredicate())
	if err != nil {
		return fmt.Errorf("failed to find JSON files: %w", err)
	}

	if len(files) == 0 {
		utils.Info("No JSON files found")
		return nil
	}

	// Format JSON files using native Go
	formatted := 0

	for _, file := range files {
		if formatJSONFile(file) {
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

	// Run gci BEFORE goimports to establish correct import order
	// gci sets the order, goimports adds/removes imports while preserving it
	if err := formatter.Gci(); err != nil {
		utils.Warn("gci failed: %v", err)
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
