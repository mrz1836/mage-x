package mage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for linting operations
var (
	errVetFailed = errors.New("go vet found issues")
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

	linter := Lint{}
	totalStart := time.Now()

	// Display linter configuration info
	displayLinterConfig()

	// Run golangci-lint
	utils.Info("Running golangci-lint...")
	if err := ensureGolangciLint(config); err != nil {
		return err
	}

	args := []string{"run", "./pkg/..."}

	// Check for config file
	if utils.FileExists(".golangci.json") {
		args = append(args, "--config", ".golangci.json")
	}

	if config.Lint.Timeout != "" {
		args = append(args, "--timeout", config.Lint.Timeout)
	}

	if config.Build.Verbose {
		args = append(args, "--verbose")
	}

	start := time.Now()
	if err := GetRunner().RunCmd("golangci-lint", args...); err != nil {
		return fmt.Errorf("golangci-lint failed: %w", err)
	}
	utils.Success("golangci-lint passed in %s", utils.FormatDuration(time.Since(start)))

	// Run go vet
	utils.Info("Running go vet...")
	if err := linter.Vet(); err != nil {
		return fmt.Errorf("go vet failed: %w", err)
	}

	utils.Success("Default linting passed in %s", utils.FormatDuration(time.Since(totalStart)))
	return nil
}

// Fix runs golangci-lint with auto-fix and applies code formatting
func (Lint) Fix() error {
	utils.Header("Running Linter with Auto-Fix and Formatting")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	linter := Lint{}
	totalStart := time.Now()

	// Display linter configuration info
	displayLinterConfig()

	// Run golangci-lint with auto-fix
	utils.Info("Running golangci-lint --fix...")
	if err := ensureGolangciLint(config); err != nil {
		return err
	}

	args := []string{"run", "--fix", "./pkg/..."}

	// Check for config file
	if utils.FileExists(".golangci.json") {
		args = append(args, "--config", ".golangci.json")
	}

	if config.Lint.Timeout != "" {
		args = append(args, "--timeout", config.Lint.Timeout)
	}

	start := time.Now()
	if err := GetRunner().RunCmd("golangci-lint", args...); err != nil {
		return fmt.Errorf("golangci-lint fix failed: %w", err)
	}
	utils.Success("golangci-lint fixes applied in %s", utils.FormatDuration(time.Since(start)))

	// Apply code formatting - prefer gofumpt, fallback to go fmt
	if err := applyCodeFormatting(linter); err != nil {
		return fmt.Errorf("formatting failed: %w", err)
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
	if err := GetRunner().RunCmd("gofumpt", "-w", "-extra", "."); err != nil {
		return fmt.Errorf("gofumpt failed: %w", err)
	}

	utils.Success("Code formatted with gofumpt")
	return nil
}

// Vet runs go vet
func (Lint) Vet() error {
	utils.Header("Running go vet")

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

	args := []string{"vet"}

	if config.Build.Verbose {
		args = append(args, "-v")
	}

	if len(config.Build.Tags) > 0 {
		args = append(args, "-tags", strings.Join(config.Build.Tags, ","))
	}

	args = append(args, modulePackages...)

	start := time.Now()
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("vet failed: %w", err)
	}

	utils.Success("Vet passed in %s", utils.FormatDuration(time.Since(start)))
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

// applyCodeFormatting applies the best available code formatting
func applyCodeFormatting(linter Lint) error {
	if utils.CommandExists("gofumpt") {
		utils.Info("Running gofumpt for stricter formatting...")
		start := time.Now()
		if err := linter.Fumpt(); err != nil {
			utils.Warn("gofumpt failed, falling back to go fmt: %v", err)
			return applyGoFmt(linter)
		}
		utils.Success("gofumpt formatting applied in %s", utils.FormatDuration(time.Since(start)))
		return nil
	}

	return applyGoFmt(linter)
}

// applyGoFmt applies basic go fmt formatting
func applyGoFmt(linter Lint) error {
	utils.Info("Running go fmt for basic formatting...")
	start := time.Now()
	if err := linter.Fmt(); err != nil {
		return err
	}
	utils.Success("go fmt formatting applied in %s", utils.FormatDuration(time.Since(start)))
	return nil
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
	if err := linter.Default(); err != nil {
		return fmt.Errorf("default linter failed: %w", err)
	}

	if err := linter.Vet(); err != nil {
		return fmt.Errorf("vet failed: %w", err)
	}

	if err := linter.Fmt(); err != nil {
		return fmt.Errorf("fmt failed: %w", err)
	}

	utils.Success("All linters passed")
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

	// Check if hadolint is available
	if !utils.CommandExists("hadolint") {
		utils.Info("hadolint not found, install it for Docker linting: brew install hadolint")
		return nil
	}

	if err := GetRunner().RunCmd("hadolint", "Dockerfile"); err != nil {
		return fmt.Errorf("docker linting failed: %w", err)
	}

	utils.Success("Docker linting passed")
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

	// Check if yamllint is available
	if !utils.CommandExists("yamllint") {
		utils.Info("yamllint not found, install it for YAML linting: pip install yamllint")
		return nil
	}

	if err := GetRunner().RunCmd("yamllint", "."); err != nil {
		return fmt.Errorf("yaml linting failed: %w", err)
	}

	utils.Success("YAML linting passed")
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

	// Check if markdownlint is available
	if !utils.CommandExists("markdownlint") {
		utils.Info("markdownlint not found, install it for Markdown linting: npm install -g markdownlint-cli")
		return nil
	}

	if err := GetRunner().RunCmd("markdownlint", "*.md"); err != nil {
		return fmt.Errorf("markdown linting failed: %w", err)
	}

	utils.Success("Markdown linting passed")
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

	// Check if shellcheck is available
	if !utils.CommandExists("shellcheck") {
		utils.Info("shellcheck not found, install it for shell linting: brew install shellcheck")
		return nil
	}

	if err := GetRunner().RunCmd("shellcheck", "**/*.sh"); err != nil {
		return fmt.Errorf("shell linting failed: %w", err)
	}

	utils.Success("Shell script linting passed")
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

	// Check if sqlfluff is available
	if !utils.CommandExists("sqlfluff") {
		utils.Info("sqlfluff not found, install it for SQL linting: pip install sqlfluff")
		return nil
	}

	if err := GetRunner().RunCmd("sqlfluff", "lint", "."); err != nil {
		return fmt.Errorf("sql linting failed: %w", err)
	}

	utils.Success("SQL linting passed")
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
func getLinterConfigInfo() (configFile string, enabledCount int, disabledCount int) {
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

	if configFile == "default (no config file found)" {
		utils.Info("Using golangci-lint defaults (no config file found)")
	} else {
		absPath, err := filepath.Abs(configFile)
		if err != nil {
			absPath = configFile
		}
		utils.Info("Config: %s", absPath)
		if enabledCount > 0 || disabledCount > 0 {
			utils.Info("Linters: %d enabled, %d disabled", enabledCount, disabledCount)
		}
	}
}
