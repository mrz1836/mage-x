package mage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for linting operations
var (
	errVetFailed     = errors.New("go vet found issues")
	errLintingFailed = errors.New("linting failed")
	errFixFailed     = errors.New("fix failed")
)

// Lint namespace for linting and formatting tasks
type Lint mg.Namespace

// Default runs the default linter (golangci-lint + go vet)
func (Lint) Default() error {
	utils.Header("Running Default Linters")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Discover all modules
	modules, err := findAllModules()
	if err != nil {
		return fmt.Errorf("failed to find modules: %w", err)
	}

	if len(modules) == 0 {
		utils.Warn("No Go modules found")
		return nil
	}

	// Show modules found
	if len(modules) > 1 {
		fmt.Printf("📦 Found %d Go modules\n", len(modules))
	}

	totalStart := time.Now()
	var moduleErrors []moduleError

	// Display linter configuration info
	displayLinterConfig()

	// Ensure golangci-lint is installed
	utils.Info("Checking golangci-lint installation...")
	if err := ensureGolangciLint(config); err != nil {
		return err
	}

	// Run linters for each module
	for _, module := range modules {
		displayModuleHeader(module, "Linting")

		moduleStart := time.Now()
		hasError := false

		// Run golangci-lint
		golangciVersion := getLinterVersion("golangci-lint")
		utils.Info("Running golangci-lint %s...", golangciVersion)
		args := []string{"run", "./..."}

		// Check for config file in root directory
		configPath := filepath.Join(module.Path, ".golangci.json")
		if !utils.FileExists(configPath) {
			// Check in root directory
			rootConfig := ".golangci.json"
			if utils.FileExists(rootConfig) {
				// Use absolute path to root config
				absPath, err := filepath.Abs(rootConfig)
				if err != nil {
					utils.Warn("Failed to get absolute path for config: %v", err)
					absPath = rootConfig
				}
				args = append(args, "--config", absPath)
			}
		} else {
			args = append(args, "--config", ".golangci.json")
		}

		if config.Lint.Timeout != "" {
			args = append(args, "--timeout", config.Lint.Timeout)
		}

		if config.Build.Verbose {
			args = append(args, "--verbose")
		}

		err := runCommandInModule(module, "golangci-lint", args...)
		if err != nil {
			hasError = true
			utils.Error("golangci-lint failed for %s", module.Relative)
		} else {
			utils.Success("golangci-lint passed for %s", module.Relative)
		}

		// Run go vet
		goVersion := getLinterVersion("go", "version")
		utils.Info("Running go vet (%s)...", goVersion)
		if err := runVetInModule(module, config); err != nil {
			hasError = true
			utils.Error("go vet failed for %s", module.Relative)
		} else {
			utils.Success("go vet passed for %s", module.Relative)
		}

		if hasError {
			moduleErrors = append(moduleErrors, moduleError{
				Module: module,
				Error:  errLintingFailed,
			})
			utils.Error("Linting failed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		} else {
			utils.Success("Linting passed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		}
	}

	// Report overall results
	if len(moduleErrors) > 0 {
		utils.Error("\nLinting failed in %d/%d modules", len(moduleErrors), len(modules))
		return formatModuleErrors(moduleErrors)
	}

	utils.Success("\nAll linting passed in %s", utils.FormatDuration(time.Since(totalStart)))
	return nil
}

// Fix runs golangci-lint with auto-fix and applies code formatting
func (Lint) Fix() error {
	utils.Header("Running Linter with Auto-Fix and Formatting")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Discover all modules
	modules, err := findAllModules()
	if err != nil {
		return fmt.Errorf("failed to find modules: %w", err)
	}

	if len(modules) == 0 {
		utils.Warn("No Go modules found")
		return nil
	}

	// Show modules found
	if len(modules) > 1 {
		fmt.Printf("📦 Found %d Go modules\n", len(modules))
	}

	totalStart := time.Now()
	var moduleErrors []moduleError

	// Display linter configuration info
	displayLinterConfig()

	// Ensure golangci-lint is installed
	utils.Info("Checking golangci-lint installation...")
	if err := ensureGolangciLint(config); err != nil {
		return err
	}

	// Run fix for each module
	for _, module := range modules {
		displayModuleHeader(module, "Fixing lint issues in")

		moduleStart := time.Now()
		hasError := false

		// Run golangci-lint with auto-fix
		golangciVersion := getLinterVersion("golangci-lint")
		utils.Info("Running golangci-lint %s --fix...", golangciVersion)
		args := []string{"run", "--fix", "./..."}

		// Check for config file
		configPath := filepath.Join(module.Path, ".golangci.json")
		if !utils.FileExists(configPath) {
			// Check in root directory
			rootConfig := ".golangci.json"
			if utils.FileExists(rootConfig) {
				// Use absolute path to root config
				absPath, err := filepath.Abs(rootConfig)
				if err != nil {
					utils.Warn("Failed to get absolute path for config: %v", err)
					absPath = rootConfig
				}
				args = append(args, "--config", absPath)
			}
		} else {
			args = append(args, "--config", ".golangci.json")
		}

		if config.Lint.Timeout != "" {
			args = append(args, "--timeout", config.Lint.Timeout)
		}

		err := runCommandInModule(module, "golangci-lint", args...)
		if err != nil {
			hasError = true
			utils.Error("golangci-lint fix failed for %s", module.Relative)
		} else {
			utils.Success("golangci-lint fixes applied for %s", module.Relative)
		}

		// Apply code formatting in module
		utils.Info("Applying code formatting...")
		if err := applyCodeFormattingInModule(module, Lint{}); err != nil {
			hasError = true
			utils.Error("Formatting failed for %s: %v", module.Relative, err)
		} else {
			utils.Success("Code formatted for %s", module.Relative)
		}

		if hasError {
			moduleErrors = append(moduleErrors, moduleError{
				Module: module,
				Error:  errFixFailed,
			})
			utils.Error("Fix failed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		} else {
			utils.Success("All issues fixed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		}
	}

	// Report overall results
	if len(moduleErrors) > 0 {
		utils.Error("\nFix failed in %d/%d modules", len(moduleErrors), len(modules))
		return formatModuleErrors(moduleErrors)
	}

	utils.Success("\nAll lint issues fixed and code formatted in %s", utils.FormatDuration(time.Since(totalStart)))
	return nil
}

// Fmt runs go fmt
func (Lint) Fmt() error {
	utils.Header("Formatting Code")

	// Get list of packages
	packages, err := utils.GoList("./...")
	if err != nil {
		return err
	}

	// Check formatting
	unformatted := []string{}
	for _, pkg := range packages {
		output, err := GetRunner().RunCmdOutput("gofmt", "-l", pkg)
		if err != nil {
			continue
		}
		if output != "" {
			unformatted = append(unformatted, strings.Split(output, "\n")...)
		}
	}

	if len(unformatted) > 0 {
		utils.Warn("Found unformatted files:")
		for _, file := range unformatted {
			if file != "" {
				fmt.Printf("  - %s\n", file)
			}
		}

		// Fix formatting
		utils.Info("Fixing formatting...")
		if err := GetRunner().RunCmd("gofmt", "-w", "."); err != nil {
			return fmt.Errorf("formatting failed: %w", err)
		}
	}

	utils.Success("Code formatted")
	return nil
}

// Fumpt runs gofumpt for stricter formatting
func (Lint) Fumpt() error {
	utils.Header("Running gofumpt")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	gofumptVersion := getLinterVersion("gofumpt")

	// Ensure gofumpt is installed
	if !utils.CommandExists("gofumpt") {
		utils.Info("Installing gofumpt...")
		version := config.Tools.Fumpt
		if version == "" || version == DefaultGoVulnCheckVersion {
			version = VersionAtLatest
		} else if !strings.HasPrefix(version, "@") {
			version = "@" + version
		}

		if err := GetRunner().RunCmd("go", "install", "mvdan.cc/gofumpt"+version); err != nil {
			return fmt.Errorf("failed to install gofumpt: %w", err)
		}
	}

	// Run gofumpt
	utils.Info("Running gofumpt %s...", gofumptVersion)
	if err := GetRunner().RunCmd("gofumpt", "-w", "-extra", "."); err != nil {
		return fmt.Errorf("gofumpt failed: %w", err)
	}

	utils.Success("Code formatted with gofumpt %s", gofumptVersion)
	return nil
}

// Vet runs go vet
func (Lint) Vet() error {
	utils.Header("Running go vet")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Discover all modules
	modules, err := findAllModules()
	if err != nil {
		return fmt.Errorf("failed to find modules: %w", err)
	}

	if len(modules) == 0 {
		utils.Warn("No Go modules found")
		return nil
	}

	// Show modules found
	if len(modules) > 1 {
		fmt.Printf("📦 Found %d Go modules\n", len(modules))
	}

	totalStart := time.Now()
	var moduleErrors []moduleError

	// Run vet for each module
	for _, module := range modules {
		displayModuleHeader(module, "Running go vet in")

		moduleStart := time.Now()

		err := runVetInModule(module, config)
		if err != nil {
			moduleErrors = append(moduleErrors, moduleError{Module: module, Error: err})
			utils.Error("Vet failed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		} else {
			utils.Success("Vet passed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		}
	}

	// Report overall results
	if len(moduleErrors) > 0 {
		utils.Error("\nVet failed in %d/%d modules", len(moduleErrors), len(modules))
		return formatModuleErrors(moduleErrors)
	}

	utils.Success("\nAll vet checks passed in %s", utils.FormatDuration(time.Since(totalStart)))
	return nil
}

// VetParallel runs go vet in parallel
func (Lint) VetParallel() error {
	utils.Header("Running go vet in parallel")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	module, err := utils.GetModuleName()
	if err != nil {
		return err
	}

	// Get packages in current module
	packages, err := utils.GoList("./...")
	if err != nil {
		return err
	}

	// Filter to only module packages
	var modulePackages []string
	for _, pkg := range packages {
		if strings.HasPrefix(pkg, module) {
			modulePackages = append(modulePackages, pkg)
		}
	}

	if len(modulePackages) == 0 {
		utils.Warn("No packages to vet")
		return nil
	}

	// Create channel for errors
	errors := make(chan error, len(modulePackages))
	semaphore := make(chan struct{}, config.Build.Parallel)

	start := time.Now()

	// Vet packages in parallel
	for _, pkg := range modulePackages {
		go func(p string) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			args := []string{"vet"}
			if len(config.Build.Tags) > 0 {
				args = append(args, "-tags", strings.Join(config.Build.Tags, ","))
			}
			args = append(args, p)

			if err := GetRunner().RunCmd("go", args...); err != nil {
				errors <- fmt.Errorf("vet %s: %w", p, err)
			} else {
				errors <- nil
			}
		}(pkg)
	}

	// Collect results
	var vetErrors []string
	for i := 0; i < len(modulePackages); i++ {
		if err := <-errors; err != nil {
			vetErrors = append(vetErrors, err.Error())
		}
	}

	if len(vetErrors) > 0 {
		return fmt.Errorf("%w:\n%s", errVetFailed, strings.Join(vetErrors, "\n"))
	}

	utils.Success("Parallel vet passed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// Version shows golangci-lint version
func (Lint) Version() error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Display configuration information
	configFile, enabledCount, disabledCount := getLinterConfigInfo()

	fmt.Printf("Configuration:\n")
	if configFile == "default (no config file found)" {
		fmt.Printf("  Config file: %s\n", configFile)
	} else {
		absPath, err := filepath.Abs(configFile)
		if err != nil {
			absPath = configFile
		}
		fmt.Printf("  Config file: %s\n", absPath)
		if enabledCount > 0 || disabledCount > 0 {
			fmt.Printf("  Linters: %d enabled, %d disabled\n", enabledCount, disabledCount)
		}
	}

	fmt.Printf("\nConfigured version: %s\n", config.Lint.GolangciVersion)

	if utils.CommandExists("golangci-lint") {
		utils.Info("\nInstalled version:")
		return GetRunner().RunCmd("golangci-lint", "--version")
	}

	utils.Warn("golangci-lint is not installed")
	return nil
}

// Helper functions

// getLinterVersion gets the version of a linter command
func getLinterVersion(command string, versionArgs ...string) string {
	if !utils.CommandExists(command) {
		return "not installed"
	}

	// Default version arguments if none provided
	if len(versionArgs) == 0 {
		versionArgs = []string{"--version"}
	}

	// Try to get version output
	output, err := GetRunner().RunCmdOutput(command, versionArgs...)
	if err != nil {
		// Try alternative version flags
		alternatives := [][]string{
			{"-version"},
			{"version"},
			{"-V"},
		}

		for _, alt := range alternatives {
			if output, err = GetRunner().RunCmdOutput(command, alt...); err == nil {
				break
			}
		}

		if err != nil {
			return statusUnknown
		}
	}

	return parseVersionFromOutput(output)
}

// parseVersionFromOutput extracts version information from command output
func parseVersionFromOutput(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return statusUnknown
	}

	// Take the first line which usually contains version info
	firstLine := strings.TrimSpace(lines[0])

	// Common version patterns to extract
	patterns := []string{
		// Matches "v1.2.3" or "version 1.2.3"
		`v?\d+\.\d+\.\d+(?:\-[a-zA-Z0-9\-\.]+)?`,
		// Matches "1.2" format
		`\d+\.\d+`,
	}

	for _, pattern := range patterns {
		if match := regexp.MustCompile(pattern).FindString(firstLine); match != "" {
			return match
		}
	}

	// If no pattern matches, return the first word that looks like a version
	words := strings.Fields(firstLine)
	for _, word := range words {
		if regexp.MustCompile(`[0-9]`).MatchString(word) {
			return word
		}
	}

	return strings.TrimSpace(firstLine)
}

// ensureGolangciLint ensures golangci-lint is installed
func ensureGolangciLint(cfg *Config) error {
	if utils.CommandExists("golangci-lint") {
		return nil
	}

	utils.Info("golangci-lint not found, installing...")

	// Try brew on macOS
	if runtime.GOOS == "darwin" && utils.CommandExists("brew") {
		utils.Info("Installing via Homebrew...")
		if err := GetRunner().RunCmd("brew", "install", "golangci-lint"); err == nil {
			return nil
		}
	}

	// Install via curl
	utils.Info("Installing via curl...")

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
	}
	binPath := filepath.Join(gopath, "bin")

	// Ensure bin directory exists
	if err := utils.EnsureDir(binPath); err != nil {
		return err
	}

	installScript := "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh"
	cmd := fmt.Sprintf("curl -sSfL %s | sh -s -- -b %s %s", installScript, binPath, cfg.Lint.GolangciVersion)

	shell, shellArgs := utils.GetShell()
	if err := GetRunner().RunCmd(shell, append(shellArgs, cmd)...); err != nil {
		return fmt.Errorf("failed to install golangci-lint: %w", err)
	}

	utils.Success("golangci-lint installed successfully")
	return nil
}

// Additional methods for Lint namespace required by tests

// All runs all linters
func (Lint) All() error {
	utils.Header("Running All Linters")

	// Display linter configuration info at the start
	displayLinterConfig()

	linter := Lint{}

	// Run each linter in sequence
	fmt.Printf("🔍 Running default linters...\n")
	if err := linter.Default(); err != nil {
		return fmt.Errorf("default linter failed: %w", err)
	}

	fmt.Printf("\n🔍 Running go vet...\n")
	if err := linter.Vet(); err != nil {
		return fmt.Errorf("vet failed: %w", err)
	}

	fmt.Printf("\n🔍 Running go fmt...\n")
	if err := linter.Fmt(); err != nil {
		return fmt.Errorf("fmt failed: %w", err)
	}

	fmt.Printf("\n✅ All linters passed\n")
	return nil
}

// Go runs Go-specific linters
func (Lint) Go() error {
	utils.Header("Running Go Linters")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Display linter configuration info
	displayLinterConfig()

	// Ensure golangci-lint is installed
	if err := ensureGolangciLint(config); err != nil {
		return err
	}

	if err := GetRunner().RunCmd("golangci-lint", "run"); err != nil {
		return fmt.Errorf("go linting failed: %w", err)
	}

	utils.Success("Go linting passed")
	return nil
}

// Docker runs Docker linters
func (Lint) Docker() error {
	utils.Header("Running Docker Linters")

	// Check if Dockerfile exists
	if !utils.FileExists("Dockerfile") {
		utils.Warn("No Dockerfile found, skipping Docker linting")
		return nil
	}

	hadolintVersion := getLinterVersion("hadolint")

	// Check if hadolint is available
	if !utils.CommandExists("hadolint") {
		utils.Info("hadolint not found, install it for Docker linting: brew install hadolint")
		return nil
	}

	utils.Info("Running hadolint %s...", hadolintVersion)
	if err := GetRunner().RunCmd("hadolint", "Dockerfile"); err != nil {
		return fmt.Errorf("docker linting failed: %w", err)
	}

	utils.Success("Docker linting passed with hadolint %s", hadolintVersion)
	return nil
}

// YAML runs YAML linters
func (Lint) YAML() error {
	utils.Header("Running YAML Linters")

	// Find YAML files
	yamlFiles, err := GetRunner().RunCmdOutput("find", ".", "-name", "*.yml", "-o", "-name", "*.yaml")
	if err != nil {
		return fmt.Errorf("failed to find YAML files: %w", err)
	}
	if yamlFiles == "" {
		utils.Info("No YAML files found")
		return nil
	}

	yamllintVersion := getLinterVersion("yamllint")

	// Check if yamllint is available
	if !utils.CommandExists("yamllint") {
		utils.Info("yamllint not found, install it for YAML linting: pip install yamllint")
		return nil
	}

	utils.Info("Running yamllint %s...", yamllintVersion)
	if err := GetRunner().RunCmd("yamllint", "."); err != nil {
		return fmt.Errorf("yaml linting failed: %w", err)
	}

	utils.Success("YAML linting passed with yamllint %s", yamllintVersion)
	return nil
}

// Yaml runs YAML linters (alias for interface compatibility)
func (Lint) Yaml() error {
	return Lint{}.YAML()
}

// Markdown runs Markdown linters
func (Lint) Markdown() error {
	utils.Header("Running Markdown Linters")

	// Find Markdown files
	mdFiles, err := GetRunner().RunCmdOutput("find", ".", "-name", "*.md", "-not", "-path", "./vendor/*")
	if err != nil {
		return fmt.Errorf("failed to find Markdown files: %w", err)
	}
	if mdFiles == "" {
		utils.Info("No Markdown files found")
		return nil
	}

	markdownlintVersion := getLinterVersion("markdownlint")

	// Check if markdownlint is available
	if !utils.CommandExists("markdownlint") {
		utils.Info("markdownlint not found, install it for Markdown linting: npm install -g markdownlint-cli")
		return nil
	}

	utils.Info("Running markdownlint %s...", markdownlintVersion)
	if err := GetRunner().RunCmd("markdownlint", "*.md"); err != nil {
		return fmt.Errorf("markdown linting failed: %w", err)
	}

	utils.Success("Markdown linting passed with markdownlint %s", markdownlintVersion)
	return nil
}

// Shell runs shell script linters
func (Lint) Shell() error {
	utils.Header("Running Shell Script Linters")

	// Find shell script files
	shellFiles, err := GetRunner().RunCmdOutput("find", ".", "-name", "*.sh", "-o", "-name", "*.bash")
	if err != nil {
		return fmt.Errorf("failed to find shell script files: %w", err)
	}
	if shellFiles == "" {
		utils.Info("No shell script files found")
		return nil
	}

	shellcheckVersion := getLinterVersion("shellcheck")

	// Check if shellcheck is available
	if !utils.CommandExists("shellcheck") {
		utils.Info("shellcheck not found, install it for shell linting: brew install shellcheck")
		return nil
	}

	utils.Info("Running shellcheck %s...", shellcheckVersion)
	if err := GetRunner().RunCmd("shellcheck", "**/*.sh"); err != nil {
		return fmt.Errorf("shell linting failed: %w", err)
	}

	utils.Success("Shell script linting passed with shellcheck %s", shellcheckVersion)
	return nil
}

// JSON runs JSON linters
func (Lint) JSON() error {
	utils.Header("Running JSON Linters")

	// Find JSON files
	jsonFiles, err := GetRunner().RunCmdOutput("find", ".", "-name", "*.json", "-not", "-path", "./vendor/*")
	if err != nil {
		return fmt.Errorf("failed to find JSON files: %w", err)
	}
	if jsonFiles == "" {
		utils.Info("No JSON files found")
		return nil
	}

	// Validate JSON syntax using built-in tools
	files := strings.Split(strings.TrimSpace(jsonFiles), "\n")
	for _, file := range files {
		if file != "" {
			if err := GetRunner().RunCmd("python3", "-m", "json.tool", file); err != nil {
				return fmt.Errorf("json validation failed for %s: %w", file, err)
			}
		}
	}

	utils.Success("JSON linting passed")
	return nil
}

// SQL runs SQL linters
func (Lint) SQL() error {
	utils.Header("Running SQL Linters")

	// Find SQL files
	sqlFiles, err := GetRunner().RunCmdOutput("find", ".", "-name", "*.sql")
	if err != nil {
		return fmt.Errorf("failed to find SQL files: %w", err)
	}
	if sqlFiles == "" {
		utils.Info("No SQL files found")
		return nil
	}

	sqlfluffVersion := getLinterVersion("sqlfluff")

	// Check if sqlfluff is available
	if !utils.CommandExists("sqlfluff") {
		utils.Info("sqlfluff not found, install it for SQL linting: pip install sqlfluff")
		return nil
	}

	utils.Info("Running sqlfluff %s...", sqlfluffVersion)
	if err := GetRunner().RunCmd("sqlfluff", "lint", "."); err != nil {
		return fmt.Errorf("sql linting failed: %w", err)
	}

	utils.Success("SQL linting passed with sqlfluff %s", sqlfluffVersion)
	return nil
}

// Config runs configuration linters
func (Lint) Config() error {
	utils.Header("Running Configuration Linters")

	// Check common config files
	configFiles := []string{".golangci.json", ".golangci.yml", "config.yaml", "config.json"}
	found := false

	for _, file := range configFiles {
		if utils.FileExists(file) {
			found = true
			utils.Info("Validating %s", file)
			// Basic validation - check if file is readable
			if strings.HasSuffix(file, ".json") {
				if err := GetRunner().RunCmd("python3", "-m", "json.tool", file); err != nil {
					return fmt.Errorf("config validation failed for %s: %w", file, err)
				}
			}
		}
	}

	if !found {
		utils.Info("No configuration files found to lint")
		return nil
	}

	utils.Success("Configuration linting passed")
	return nil
}

// CI runs linters for CI environment
func (Lint) CI() error {
	utils.Header("Running CI Linters")

	linter := Lint{}

	// Run the essential linters for CI
	if err := linter.Default(); err != nil {
		return fmt.Errorf("default linter failed: %w", err)
	}

	if err := linter.Vet(); err != nil {
		return fmt.Errorf("vet failed: %w", err)
	}

	// Check CI-specific files
	ciFiles := []string{".github/workflows/*.yml", ".github/workflows/*.yaml", ".gitlab-ci.yml", ".travis.yml"}
	for _, pattern := range ciFiles {
		files, err := GetRunner().RunCmdOutput("ls", "-la", pattern)
		if err == nil && files != "" {
			utils.Info("Found CI configuration files")
			break
		}
	}

	utils.Success("CI linting passed")
	return nil
}

// Fast runs fast linters only
func (Lint) Fast() error {
	utils.Header("Running Fast Linters")

	linter := Lint{}

	// Run only the fastest linters
	if err := linter.Fmt(); err != nil {
		return fmt.Errorf("fmt failed: %w", err)
	}

	if err := linter.Vet(); err != nil {
		return fmt.Errorf("vet failed: %w", err)
	}

	utils.Success("Fast linting passed")
	return nil
}

// Helper functions for config information

// getLinterConfigInfo returns information about the golangci-lint configuration
func getLinterConfigInfo() (configFile string, enabledCount, disabledCount int) {
	// Check for config files in order of precedence
	configFiles := []string{".golangci.json", ".golangci.yml", ".golangci.yaml", "golangci.yml", "golangci.yaml"}

	for _, file := range configFiles {
		if utils.FileExists(file) {
			configFile = file
			break
		}
	}

	if configFile == "" {
		return "default (no config file found)", 0, 0
	}

	// For JSON config, parse and count linters
	if strings.HasSuffix(configFile, ".json") {
		// Use filepath.Clean to sanitize the config file path
		cleanPath := filepath.Clean(configFile)
		data, err := os.ReadFile(cleanPath)
		if err != nil {
			return configFile, 0, 0
		}

		var config struct {
			Linters struct {
				Enable  []string `json:"enable"`
				Disable []string `json:"disable"`
			} `json:"linters"`
		}

		if err := json.Unmarshal(data, &config); err == nil {
			enabledCount = len(config.Linters.Enable)
			disabledCount = len(config.Linters.Disable)
		}
	}

	return configFile, enabledCount, disabledCount
}

// displayLinterConfig displays linter configuration information
func displayLinterConfig() {
	configFile, enabledCount, disabledCount := getLinterConfigInfo()
	golangciVersion := getLinterVersion("golangci-lint")

	if configFile == "default (no config file found)" {
		utils.Info("Using golangci-lint %s defaults (no config file found)", golangciVersion)
	} else {
		absPath, err := filepath.Abs(configFile)
		if err != nil {
			absPath = configFile
		}
		utils.Info("Config: %s", absPath)
		if enabledCount > 0 || disabledCount > 0 {
			utils.Info("Linters: %d enabled, %d disabled (golangci-lint %s)", enabledCount, disabledCount, golangciVersion)
		} else {
			utils.Info("Using golangci-lint %s", golangciVersion)
		}
	}
}

// runVetInModule runs go vet in a specific module directory
func runVetInModule(module ModuleInfo, config *Config) error {
	args := []string{"vet"}

	if config.Build.Verbose {
		args = append(args, "-v")
	}

	if len(config.Build.Tags) > 0 {
		args = append(args, "-tags", strings.Join(config.Build.Tags, ","))
	}

	args = append(args, "./...")

	return runCommandInModule(module, "go", args...)
}

// applyCodeFormattingInModule applies code formatting in a specific module
func applyCodeFormattingInModule(module ModuleInfo, linter Lint) error {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Change to module directory
	if err := os.Chdir(module.Path); err != nil {
		return fmt.Errorf("failed to change to directory %s: %w", module.Path, err)
	}

	// Ensure we change back to original directory
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			utils.Error("Failed to change back to original directory: %v", err)
		}
	}()

	// Apply formatting
	if utils.CommandExists("gofumpt") {
		return applyGofumptFormatting()
	}
	return applyGofmtFormatting()
}

// applyGofumptFormatting applies gofumpt formatting with fallback to gofmt
func applyGofumptFormatting() error {
	start := time.Now()
	if err := GetRunner().RunCmd("gofumpt", "-w", "-extra", "."); err != nil {
		utils.Warn("gofumpt failed, falling back to go fmt: %v", err)
		return applyGofmtFallback(start)
	}
	utils.Success("gofumpt formatting applied in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// applyGofmtFormatting applies standard gofmt formatting
func applyGofmtFormatting() error {
	start := time.Now()
	if err := GetRunner().RunCmd("gofmt", "-w", "."); err != nil {
		return fmt.Errorf("go fmt failed: %w", err)
	}
	utils.Success("go fmt formatting applied in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// applyGofmtFallback applies gofmt as a fallback when gofumpt fails
func applyGofmtFallback(start time.Time) error {
	if err := GetRunner().RunCmd("gofmt", "-w", "."); err != nil {
		return fmt.Errorf("go fmt failed: %w", err)
	}
	utils.Success("go fmt formatting applied in %s", utils.FormatDuration(time.Since(start)))
	return nil
}
