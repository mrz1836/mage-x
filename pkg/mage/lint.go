package mage

import (
	"context"
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

// Constants for common string values
const (
	trueValue = "true"
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
	var modules []ModuleInfo
	modules, err = findAllModules()
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
	if err = ensureGolangciLint(config); err != nil {
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
				var absPath string
				absPath, err = filepath.Abs(rootConfig)
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

		// Add build tags if configured
		if len(config.Build.Tags) > 0 {
			args = append(args, "--build-tags", strings.Join(config.Build.Tags, ","))
		}

		// Add verbose flag if enabled via parameter, environment, or config
		if shouldUseVerboseMode(config) {
			args = append(args, "--verbose")
		}

		err = runCommandInModule(module, "golangci-lint", args...)
		if err != nil {
			hasError = true
			utils.Error("golangci-lint failed for %s", module.Relative)
		} else {
			utils.Success("golangci-lint passed for %s", module.Relative)
		}

		// Run go vet
		goVersion := getLinterVersion("go", "version")
		utils.Info("Running go vet (%s)...", goVersion)
		if err = runVetInModule(module, config); err != nil {
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
		utils.Error("Linting failed in %d/%d modules", len(moduleErrors), len(modules))
		return formatModuleErrors(moduleErrors)
	}

	utils.Success("All linting passed in %s", utils.FormatDuration(time.Since(totalStart)))
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
	var modules []ModuleInfo
	modules, err = findAllModules()
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
	if err = ensureGolangciLint(config); err != nil {
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
				var absPath string
				absPath, err = filepath.Abs(rootConfig)
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

		// Add build tags if configured
		if len(config.Build.Tags) > 0 {
			args = append(args, "--build-tags", strings.Join(config.Build.Tags, ","))
		}

		// Add verbose flag if enabled via parameter, environment, or config
		if shouldUseVerboseMode(config) {
			args = append(args, "--verbose")
		}

		err = runCommandInModule(module, "golangci-lint", args...)
		if err != nil {
			hasError = true
			utils.Error("golangci-lint fix failed for %s", module.Relative)
		} else {
			utils.Success("golangci-lint fixes applied for %s", module.Relative)
		}

		// Apply code formatting in module
		utils.Info("Applying code formatting...")
		if err = applyCodeFormattingInModule(module, Lint{}); err != nil {
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
		utils.Error("Fix failed in %d/%d modules", len(moduleErrors), len(modules))
		return formatModuleErrors(moduleErrors)
	}

	utils.Success("All lint issues fixed and code formatted in %s", utils.FormatDuration(time.Since(totalStart)))
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
	var unformatted []string
	for _, pkg := range packages {
		var output string
		output, err = GetRunner().RunCmdOutput("gofmt", "-l", pkg)
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
		if err = GetRunner().RunCmd("gofmt", "-w", "."); err != nil {
			return fmt.Errorf("formatting failed: %w", err)
		}
	}

	utils.Success("Code formatted")
	return nil
}

// Fumpt runs gofumpt for stricter formatting with retry logic
func (Lint) Fumpt() error {
	utils.Header("Running gofumpt")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	gofumptVersion := getLinterVersion("gofumpt")

	// Ensure gofumpt is installed with retry logic

	if !utils.CommandExists("gofumpt") {
		if err = installGofumpt(config); err != nil {
			return err
		}
	}

	// Run gofumpt
	utils.Info("Running gofumpt %s...", gofumptVersion)
	if err = GetRunner().RunCmd("gofumpt", "-w", "-extra", "."); err != nil {
		return fmt.Errorf("gofumpt failed: %w", err)
	}

	utils.Success("Code formatted with gofumpt %s", gofumptVersion)
	return nil
}

// installGofumpt installs gofumpt with retry and fallback logic
func installGofumpt(config *Config) error {
	utils.Info("Installing gofumpt with retry logic...")

	ctx := context.Background()
	maxRetries := config.Download.MaxRetries
	initialDelay := time.Duration(config.Download.InitialDelayMs) * time.Millisecond

	fumptVersion := config.Tools.Fumpt
	if fumptVersion == "" {
		utils.Warn("Gofumpt version not available, using @latest")
		fumptVersion = VersionAtLatest
	} else if !strings.HasPrefix(fumptVersion, "@") {
		fumptVersion = "@" + fumptVersion
	}

	moduleWithVersion := "mvdan.cc/gofumpt" + fumptVersion

	// Get the secure executor with retry capabilities
	runner := GetRunner()
	secureRunner, ok := runner.(*SecureCommandRunner)
	if !ok {
		return fmt.Errorf("%w: got %T", ErrUnexpectedRunnerType, runner)
	}
	executor := secureRunner.executor
	err := executor.ExecuteWithRetry(ctx, maxRetries, initialDelay, "go", "install", moduleWithVersion)
	if err != nil {
		// Try with direct proxy as fallback
		utils.Warn("Installation failed: %v, trying direct proxy...", err)

		env := []string{"GOPROXY=direct"}
		if err = executor.ExecuteWithEnv(ctx, env, "go", "install", moduleWithVersion); err != nil {
			return fmt.Errorf("failed to install gofumpt after %d retries and fallback: %w", maxRetries, err)
		}
	}
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

		err = runVetInModule(module, config)
		if err != nil {
			moduleErrors = append(moduleErrors, moduleError{Module: module, Error: err})
			utils.Error("Vet failed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		} else {
			utils.Success("Vet passed for %s in %s", module.Relative, utils.FormatDuration(time.Since(moduleStart)))
		}
	}

	// Report overall results
	if len(moduleErrors) > 0 {
		utils.Error("Vet failed in %d/%d modules", len(moduleErrors), len(modules))
		return formatModuleErrors(moduleErrors)
	}

	utils.Success("All vet checks passed in %s", utils.FormatDuration(time.Since(totalStart)))
	return nil
}

// VetParallel runs go vet in parallel
func (Lint) VetParallel() error {
	utils.Header("Running go vet in parallel")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	var module string
	module, err = utils.GetModuleName()
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
	errs := make(chan error, len(modulePackages))
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

			if runErr := GetRunner().RunCmd("go", args...); runErr != nil {
				errs <- fmt.Errorf("vet %s: %w", p, runErr)
			} else {
				errs <- nil
			}
		}(pkg)
	}

	// Collect results
	var vetErrors []string
	for i := 0; i < len(modulePackages); i++ {
		if runErr := <-errs; runErr != nil {
			vetErrors = append(vetErrors, runErr.Error())
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
		var absPath string
		absPath, err = filepath.Abs(configFile)
		if err != nil {
			absPath = configFile
		}
		fmt.Printf("  Config file: %s\n", absPath)
		if enabledCount > 0 || disabledCount > 0 {
			fmt.Printf("  Linters: %d enabled, %d disabled\n", enabledCount, disabledCount)
		}
	}

	// Get versions
	configuredEnvVersion := GetToolVersion("golangci-lint")
	installedVersion := getLinterVersion("golangci-lint")

	fmt.Printf("\nVersion Information:\n")
	if configuredEnvVersion != "" {
		fmt.Printf("  Configured (.env.base): %s\n", configuredEnvVersion)
	} else {
		fmt.Printf("  Configured (.env.base): not set (source .github/.env.base)\n")
	}

	fmt.Printf("  Configured (mage.yaml): %s\n", config.Lint.GolangciVersion)

	if utils.CommandExists("golangci-lint") {
		fmt.Printf("  Installed: %s\n", installedVersion)

		// Check for version mismatch
		checkLinterVersionMismatch(installedVersion)

		fmt.Printf("\nFull version details:\n")
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

// ensureGolangciLint ensures golangci-lint is installed with retry logic
func ensureGolangciLint(cfg *Config) error {
	if utils.CommandExists("golangci-lint") {
		return nil
	}

	utils.Info("golangci-lint not found, installing...")

	ctx := context.Background()
	maxRetries := cfg.Download.MaxRetries
	initialDelay := time.Duration(cfg.Download.InitialDelayMs) * time.Millisecond

	// Try brew on macOS with retry logic
	if runtime.GOOS == "darwin" && utils.CommandExists("brew") {
		utils.Info("Installing via Homebrew with retry logic...")

		// Get the secure executor with retry capabilities
		runner := GetRunner()
		secureRunner, ok := runner.(*SecureCommandRunner)
		if !ok {
			return fmt.Errorf("%w: got %T", ErrUnexpectedRunnerType, runner)
		}
		executor := secureRunner.executor
		err := executor.ExecuteWithRetry(ctx, maxRetries, initialDelay, "brew", "install", "golangci-lint")

		if err == nil {
			utils.Success("golangci-lint installed successfully via Homebrew")
			return nil
		}

		utils.Warn("Homebrew installation failed: %v, trying curl method...", err)
	}

	// Install via curl with download retry
	utils.Info("Installing via curl with retry logic...")

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
	}
	binPath := filepath.Join(gopath, "bin")

	// Ensure bin directory exists
	if err := utils.EnsureDir(binPath); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Use the new download utility with retry logic
	downloadConfig := &utils.DownloadConfig{
		MaxRetries:        cfg.Download.MaxRetries,
		InitialDelay:      time.Duration(cfg.Download.InitialDelayMs) * time.Millisecond,
		MaxDelay:          time.Duration(cfg.Download.MaxDelayMs) * time.Millisecond,
		Timeout:           time.Duration(cfg.Download.TimeoutMs) * time.Millisecond,
		BackoffMultiplier: cfg.Download.BackoffMultiplier,
		EnableResume:      cfg.Download.EnableResume,
		UserAgent:         cfg.Download.UserAgent,
	}

	installScriptURL := "https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh"
	scriptArgs := fmt.Sprintf("-- -b %s %s", binPath, cfg.Lint.GolangciVersion)

	// Download and execute the installation script with retry logic
	err := utils.DownloadScript(ctx, installScriptURL, scriptArgs, downloadConfig)
	if err != nil {
		// Fallback to direct command execution with retry
		utils.Warn("Script download failed: %v, trying direct curl command...", err)

		cmd := fmt.Sprintf("curl -sSfL %s | sh -s -- -b %s %s", installScriptURL, binPath, cfg.Lint.GolangciVersion)
		shell, shellArgs := utils.GetShell()

		// Get the secure executor with retry capabilities
		runner := GetRunner()
		secureRunner, ok := runner.(*SecureCommandRunner)
		if !ok {
			return fmt.Errorf("%w: got %T", ErrUnexpectedRunnerType, runner)
		}
		executor := secureRunner.executor
		if err = executor.ExecuteWithRetry(ctx, maxRetries, initialDelay, shell, append(shellArgs, cmd)...); err != nil {
			return fmt.Errorf("failed to install golangci-lint after %d retries: %w", maxRetries, err)
		}
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
	if err = ensureGolangciLint(config); err != nil {
		return err
	}

	if err = GetRunner().RunCmd("golangci-lint", "run"); err != nil {
		return fmt.Errorf("go linting failed: %w", err)
	}

	utils.Success("Go linting passed")
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
	if err = GetRunner().RunCmd("yamllint", "."); err != nil {
		return fmt.Errorf("yaml linting failed: %w", err)
	}

	utils.Success("YAML linting passed with yamllint %s", yamllintVersion)
	return nil
}

// Yaml runs YAML linters (alias for interface compatibility)
func (Lint) Yaml() error {
	return Lint{}.YAML()
}

// validateJSONFile validates a single JSON file using native Go
func validateJSONFile(file string) error {
	// Read the JSON file
	data, err := os.ReadFile(file) //nolint:gosec // file path is user-provided via API
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse JSON to validate syntax
	var jsonData interface{}
	if err = json.Unmarshal(data, &jsonData); err != nil {
		return fmt.Errorf("invalid JSON syntax: %w", err)
	}

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

	// Validate JSON syntax using native Go
	files := strings.Split(strings.TrimSpace(jsonFiles), "\n")
	for _, file := range files {
		if file != "" {
			if err := validateJSONFile(file); err != nil {
				return fmt.Errorf("json validation failed for %s: %w", file, err)
			}
		}
	}

	utils.Success("JSON linting passed")
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
				if err := validateJSONFile(file); err != nil {
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

// Fast will run fast linters only
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

		if err = json.Unmarshal(data, &config); err == nil {
			enabledCount = len(config.Linters.Enable)
			disabledCount = len(config.Linters.Disable)
		}
	}

	return configFile, enabledCount, disabledCount
}

// checkLinterVersionMismatch checks if the installed linter version matches the configured version
func checkLinterVersionMismatch(installedVersion string) {
	// Get configured version from environment
	configuredVersion := GetToolVersion("golangci-lint")

	// Skip check if no configured version is set
	if configuredVersion == "" {
		return
	}

	// Skip check if linter is not installed
	if installedVersion == "not installed" || installedVersion == statusUnknown {
		return
	}

	// Normalize versions for comparison (remove 'v' prefix if present)
	normalizedConfigured := strings.TrimPrefix(configuredVersion, "v")
	normalizedInstalled := strings.TrimPrefix(installedVersion, "v")

	// Compare versions
	if normalizedConfigured != normalizedInstalled {
		utils.Warn("⚠️  Version mismatch detected:")
		utils.Warn("    Configured version (MAGE_X_GOLANGCI_LINT_VERSION): %s", configuredVersion)
		utils.Warn("    Installed version: %s", installedVersion)
		utils.Warn("    💡 To fix: source .github/.env.base or reinstall golangci-lint")
	}
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

	// Check for version mismatch
	checkLinterVersionMismatch(golangciVersion)

	// Display build tags information and verbose mode status
	config, err := GetConfig()
	if err == nil {
		if len(config.Build.Tags) > 0 {
			utils.Info("Build tags: %s", strings.Join(config.Build.Tags, ", "))
		}

		// Show verbose mode status
		verboseStatus := "disabled"
		if shouldUseVerboseMode(config) {
			verboseStatus = "enabled"
		}
		utils.Info("Verbose output: %s", verboseStatus)
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
func applyCodeFormattingInModule(module ModuleInfo, _ Lint) error {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Change to module directory
	if err = os.Chdir(module.Path); err != nil {
		return fmt.Errorf("failed to change to directory %s: %w", module.Path, err)
	}

	// Ensure we change back to original directory
	defer func() {
		if err = os.Chdir(originalDir); err != nil {
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

// IssueCount represents counts for a specific issue type
type IssueCount struct {
	Message string
	Count   int
	Files   []string
}

// Issues scans the codebase for TODOs, FIXMEs, HACKs, nolint directives, test skips, and disabled files
func (Lint) Issues() error {
	utils.Header("Scanning Codebase for Issues")

	start := time.Now()

	// Scan for TODOs, FIXMEs, HACKs
	todoIssues := scanForComments()

	// Scan for nolint directives
	nolintIssues := scanForNolintDirectives()

	// Scan for test skips
	skipIssues := scanForTestSkips()

	// Scan for disabled files
	disabledFiles := scanForDisabledFiles()

	// Display results
	displayIssueResults(todoIssues, nolintIssues, skipIssues, disabledFiles)

	// Calculate totals
	totalTodos := countTotalIssues(todoIssues)
	totalNolints := countTotalIssues(nolintIssues)
	totalSkips := countTotalIssues(skipIssues)

	// Display summary
	utils.Info("📊 Summary (scanned in %s):", utils.FormatDuration(time.Since(start)))
	fmt.Printf("  • Code comments: %d issues\n", totalTodos)
	fmt.Printf("  • Nolint directives: %d issues\n", totalNolints)
	fmt.Printf("  • Test skips: %d issues\n", totalSkips)
	fmt.Printf("  • Disabled files: %d files\n", len(disabledFiles))

	grandTotal := totalTodos + totalNolints + totalSkips + len(disabledFiles)
	if grandTotal == 0 {
		utils.Success("✨ No issues found!")
	} else {
		utils.Warn("Total: %d issues found", grandTotal)
	}

	return nil
}

// scanForComments scans for TODO, FIXME, and HACK comments
func scanForComments() map[string][]IssueCount {
	patterns := map[string]string{
		"TODO":  "TODO",
		"FIXME": "FIXME",
		"HACK":  "HACK",
	}

	results := make(map[string][]IssueCount)

	for category, pattern := range patterns {
		matches := findMatches(pattern)
		counts := groupByMessage(matches, pattern)
		if len(counts) > 0 {
			results[category] = counts
		}
	}

	return results
}

// scanForNolintDirectives scans for //nolint directives
func scanForNolintDirectives() map[string][]IssueCount {
	matches := findMatches("//nolint")
	results := make(map[string][]IssueCount)
	if len(matches) > 0 {
		counts := groupNolintByTag(matches)
		results["NOLINT"] = counts
	}

	return results
}

// scanForTestSkips scans for t.Skip() usage in tests
func scanForTestSkips() map[string][]IssueCount {
	matches := findMatches(`t\.Skip\(`)
	results := make(map[string][]IssueCount)
	if len(matches) > 0 {
		counts := groupSkipsByMessage(matches)
		results["TEST_SKIP"] = counts
	}

	return results
}

// scanForDisabledFiles scans for .go.disabled files
func scanForDisabledFiles() []string {
	output, err := GetRunner().RunCmdOutput("find", ".", "-name", "*.go.disabled", "-type", "f")
	if err != nil {
		// find might fail if no files found, which is okay for this use case
		return []string{}
	}

	if output == "" {
		return []string{}
	}

	files := strings.Split(strings.TrimSpace(output), "\n")
	var cleanFiles []string
	for _, file := range files {
		if file != "" {
			cleanFiles = append(cleanFiles, strings.TrimPrefix(file, "./"))
		}
	}

	return cleanFiles
}

// findMatches uses grep to find pattern matches in Go files
func findMatches(pattern string) []string {
	// Use word boundaries for comment patterns and exclude binary files
	var cmd []string
	if pattern == "TODO" || pattern == "FIXME" || pattern == "HACK" {
		// Look for comment patterns: // TODO or /* TODO
		grepPattern := `//.*` + pattern + `|/\*.*` + pattern
		cmd = []string{"grep", "-rn", "--include=*.go", "--exclude-dir=vendor", "--exclude-dir=.git", "-E", grepPattern, "."}
	} else {
		// For other patterns, use them directly
		cmd = []string{"grep", "-rn", "--include=*.go", "--exclude-dir=vendor", "--exclude-dir=.git", "-E", pattern, "."}
	}

	output, err := GetRunner().RunCmdOutput(cmd[0], cmd[1:]...)
	if err != nil {
		// grep returns non-zero exit code when no matches found, which is expected
		return []string{}
	}

	if output == "" {
		return []string{}
	}

	return strings.Split(strings.TrimSpace(output), "\n")
}

// groupByMessage groups matches by the comment message
func groupByMessage(matches []string, pattern string) []IssueCount {
	messageCounts := make(map[string]*IssueCount)

	for _, match := range matches {
		if match == "" {
			continue
		}

		parts := strings.SplitN(match, ":", 3)
		if len(parts) < 3 {
			continue
		}

		file := parts[0]
		content := parts[2]

		// Simple approach: find the pattern and extract what comes after it
		// Look for comment start first
		if !strings.Contains(content, "//") && !strings.Contains(content, "/*") {
			continue
		}

		// Skip lines that are just describing these patterns (like in our code)
		lowerContent := strings.ToLower(content)
		if strings.Contains(lowerContent, "scan") && (strings.Contains(lowerContent, "todo") || strings.Contains(lowerContent, "fixme") || strings.Contains(lowerContent, "hack")) {
			continue
		}
		if strings.Contains(lowerContent, "extract") || strings.Contains(lowerContent, "pattern") {
			continue
		}

		// Find the pattern in the content
		idx := strings.Index(content, pattern)
		if idx == -1 {
			continue
		}

		// Extract everything after the pattern
		afterPattern := content[idx+len(pattern):]

		// Remove common separators
		message := strings.TrimLeft(afterPattern, ": ")
		message = strings.TrimSpace(message)

		// Remove trailing comment markers
		message = strings.TrimRight(message, "*/")
		message = strings.TrimSpace(message)

		if message == "" {
			message = "(no message)"
		}

		// Limit message length for readability
		if len(message) > 100 {
			message = message[:97] + "..."
		}

		key := pattern + ": " + message
		if existing, ok := messageCounts[key]; ok {
			existing.Count++
			existing.Files = append(existing.Files, file)
		} else {
			messageCounts[key] = &IssueCount{
				Message: message,
				Count:   1,
				Files:   []string{file},
			}
		}
	}

	// Convert to slice and sort by count (descending)
	results := make([]IssueCount, 0, len(messageCounts))
	for _, count := range messageCounts {
		results = append(results, *count)
	}

	// Sort by count (descending), then by message
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Count < results[j].Count ||
				(results[i].Count == results[j].Count && results[i].Message > results[j].Message) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

// groupNolintByTag groups nolint directives by their tags
func groupNolintByTag(matches []string) []IssueCount {
	tagCounts := make(map[string]*IssueCount)

	re := regexp.MustCompile(`//nolint:([a-zA-Z0-9,_-]+)`)

	for _, match := range matches {
		if match == "" {
			continue
		}

		parts := strings.SplitN(match, ":", 3)
		if len(parts) < 3 {
			continue
		}

		file := parts[0]
		content := parts[2]

		// Extract nolint tags
		submatches := re.FindAllStringSubmatch(content, -1)
		for _, submatch := range submatches {
			if len(submatch) < 2 {
				continue
			}

			tags := strings.Split(submatch[1], ",")
			for _, tag := range tags {
				tag = strings.TrimSpace(tag)
				if tag == "" {
					continue
				}

				if existing, ok := tagCounts[tag]; ok {
					existing.Count++
					existing.Files = append(existing.Files, file)
				} else {
					tagCounts[tag] = &IssueCount{
						Message: tag,
						Count:   1,
						Files:   []string{file},
					}
				}
			}
		}
	}

	// Convert to slice and sort
	results := make([]IssueCount, 0, len(tagCounts))
	for _, count := range tagCounts {
		results = append(results, *count)
	}

	// Sort by count (descending), then by tag name
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Count < results[j].Count ||
				(results[i].Count == results[j].Count && results[i].Message > results[j].Message) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

// groupSkipsByMessage groups test skips by their skip message
func groupSkipsByMessage(matches []string) []IssueCount {
	messageCounts := make(map[string]*IssueCount)

	re := regexp.MustCompile(`t\.Skip\("([^"]+)"\)`)

	for _, match := range matches {
		if match == "" {
			continue
		}

		parts := strings.SplitN(match, ":", 3)
		if len(parts) < 3 {
			continue
		}

		file := parts[0]
		content := parts[2]

		// Extract skip message
		submatches := re.FindStringSubmatch(content)
		var message string
		if len(submatches) >= 2 {
			message = submatches[1]
		} else {
			message = "(no message)"
		}

		if existing, ok := messageCounts[message]; ok {
			existing.Count++
			existing.Files = append(existing.Files, file)
		} else {
			messageCounts[message] = &IssueCount{
				Message: message,
				Count:   1,
				Files:   []string{file},
			}
		}
	}

	// Convert to slice and sort
	results := make([]IssueCount, 0, len(messageCounts))
	for _, count := range messageCounts {
		results = append(results, *count)
	}

	// Sort by count (descending), then by message
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Count < results[j].Count ||
				(results[i].Count == results[j].Count && results[i].Message > results[j].Message) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

// displayIssueResults displays the results in a formatted way
func displayIssueResults(todoIssues, nolintIssues, skipIssues map[string][]IssueCount, disabledFiles []string) {
	// Display TODO/FIXME/HACK results
	for category, issues := range todoIssues {
		if len(issues) > 0 {
			fmt.Printf("\n📝 %s Comments:\n", category)
			for _, issue := range issues {
				fmt.Printf("  • %s (%d occurrence%s)\n", issue.Message, issue.Count, pluralize(issue.Count))
			}
		}
	}

	// Display nolint directives
	for category, issues := range nolintIssues {
		if len(issues) > 0 {
			fmt.Printf("\n🚫 %s Directives:\n", category)
			for _, issue := range issues {
				fmt.Printf("  • %s (%d occurrence%s)\n", issue.Message, issue.Count, pluralize(issue.Count))
			}
		}
	}

	// Display test skips
	for category, issues := range skipIssues {
		if len(issues) > 0 {
			fmt.Printf("\n⏭️  %s Usage:\n", strings.ReplaceAll(category, "_", " "))
			for _, issue := range issues {
				fmt.Printf("  • %s (%d occurrence%s)\n", issue.Message, issue.Count, pluralize(issue.Count))
			}
		}
	}

	// Display disabled files
	if len(disabledFiles) > 0 {
		fmt.Printf("\n🚫 Disabled Files:\n")
		for _, file := range disabledFiles {
			fmt.Printf("  • %s\n", file)
		}
	}
}

// countTotalIssues counts total issues across all categories
func countTotalIssues(issues map[string][]IssueCount) int {
	total := 0
	for _, categoryIssues := range issues {
		for _, issue := range categoryIssues {
			total += issue.Count
		}
	}
	return total
}

// pluralize returns "s" for counts != 1, empty string otherwise
func pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

// shouldUseVerboseMode checks if verbose mode should be enabled based on:
// 1. Command-line parameters (highest priority)
// 2. Environment variables
// 3. Config settings
func shouldUseVerboseMode(config *Config) bool {
	// Parse command-line parameters from os.Args
	var targetArgs []string
	for i, arg := range os.Args {
		if strings.Contains(arg, "lint") {
			targetArgs = os.Args[i+1:]
			break
		}
	}

	// Check command-line parameter first (highest priority)
	params := utils.ParseParams(targetArgs)
	if verboseParam := utils.GetParam(params, "verbose", ""); verboseParam != "" {
		return verboseParam == trueValue || verboseParam == "1"
	}

	// Check environment variable
	if verboseEnv := utils.GetEnv("MAGE_X_LINT_VERBOSE", ""); verboseEnv != "" {
		return verboseEnv == trueValue || verboseEnv == "1"
	}

	// Default: use config setting or false
	return config.Build.Verbose
}

// Verbose runs the default linting with explicit verbose output control
func (Lint) Verbose() error {
	utils.Header("Running Linters (Verbose)")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Force verbose mode for this command
	originalVerbose := config.Build.Verbose
	config.Build.Verbose = true

	// Restore original setting after execution
	defer func() {
		config.Build.Verbose = originalVerbose
	}()

	// Run the default linting logic
	linter := Lint{}
	return linter.Default()
}
