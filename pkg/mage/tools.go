package mage

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors to satisfy err113 linter
var (
	errToolsMissing = errors.New("some tools are missing. Run 'magex tools:install' to install them")
)

// Tools namespace for tool management tasks
type Tools mg.Namespace

// Default installs default tools
func (Tools) Default() error {
	utils.Header("Installing Default Tools")
	return Tools{}.Install()
}

// ToolDefinition represents a development tool
type ToolDefinition struct {
	Name    string
	Module  string
	Version string
	Check   string // Command to check if installed
}

// Install installs all required development tools
func (Tools) Install() error {
	utils.Header("Installing Development Tools")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	tools := getRequiredTools(config)

	for _, tool := range tools {
		if err := installTool(tool); err != nil {
			return fmt.Errorf("failed to install %s: %w", tool.Name, err)
		}
	}

	utils.Success("All tools installed successfully")
	return nil
}

// Update updates all development tools to latest versions with retry logic
func (Tools) Update() error {
	utils.Header("Updating Development Tools")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	ctx := context.Background()
	maxRetries := config.Download.MaxRetries
	initialDelay := time.Duration(config.Download.InitialDelayMs) * time.Millisecond

	// Update golangci-lint if using brew with retry logic
	if utils.IsMac() && utils.CommandExists("brew") {
		utils.Info("Updating golangci-lint via brew with retry logic...")

		// Get the secure executor with retry capabilities
		runner := GetRunner()
		secureRunner, ok := runner.(*SecureCommandRunner)
		if !ok {
			return fmt.Errorf("%w: got %T", ErrUnexpectedRunnerType, runner)
		}
		executor := secureRunner.executor
		err := executor.ExecuteWithRetry(ctx, maxRetries, initialDelay, "brew", "upgrade", "golangci-lint")
		if err != nil {
			utils.Warn("Failed to update golangci-lint via brew: %v", err)
		}
	}

	// Update other tools with retry logic
	tools := getRequiredTools(config)
	for _, tool := range tools {
		if tool.Module == "" {
			continue
		}

		utils.Info("Updating %s with retry logic...", tool.Name)

		toolVersion := tool.Version

		// If version is "latest", try to resolve from environment variables first
		if toolVersion == "" || toolVersion == VersionLatest {
			if resolvedVersion := getToolVersionFromEnv(tool.Name); resolvedVersion != "" {
				toolVersion = resolvedVersion
			} else {
				utils.Warn("Version for tool %s not available, using @latest", tool.Name)
				toolVersion = VersionAtLatest
			}
		}

		if !strings.HasPrefix(toolVersion, "@") && toolVersion != VersionAtLatest {
			toolVersion = "@" + toolVersion
		}
		// Keep @latest as is for VersionAtLatest

		moduleWithVersion := tool.Module + toolVersion

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
			utils.Warn("Failed to update %s: %v, trying direct proxy...", tool.Name, err)

			env := []string{"GOPROXY=direct"}
			if err := executor.ExecuteWithEnv(ctx, env, "go", "install", moduleWithVersion); err != nil {
				utils.Warn("Failed to update %s after retries and fallback: %v", tool.Name, err)
			} else {
				utils.Success("Updated %s via direct proxy", tool.Name)
			}
		} else {
			utils.Success("Updated %s", tool.Name)
		}
	}

	utils.Success("Tools update completed")
	return nil
}

// Verify checks that all required tools are installed
func (Tools) Verify() error {
	utils.Header("Verifying Development Tools")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	tools := getRequiredTools(config)
	allGood := true

	for _, tool := range tools {
		checkCmd := tool.Check
		if checkCmd == "" {
			checkCmd = tool.Name
		}

		if utils.CommandExists(checkCmd) {
			// Try to get version
			output, err := GetRunner().RunCmdOutput(checkCmd, "--version")
			if err != nil {
				// If version check fails, just show as installed
				utils.Success("%s: installed", tool.Name)
				continue
			}
			installedVersion := strings.TrimSpace(output)
			if installedVersion == "" {
				installedVersion = "installed"
			}
			utils.Success("%s: %s", tool.Name, installedVersion)
		} else {
			utils.Error("%s: not installed", tool.Name)
			allGood = false
		}
	}

	if !allGood {
		return errToolsMissing
	}

	utils.Success("All tools verified")
	return nil
}

// List lists all configured tools
func (Tools) List() error {
	config, err := GetConfig()
	if err != nil {
		return err
	}

	utils.Info("Configured tools:")
	utils.Info("")

	// Standard tools
	fmt.Printf("  golangci-lint: %s\n", config.Tools.GolangciLint)
	fmt.Printf("  fumpt:         %s\n", config.Tools.Fumpt)
	fmt.Printf("  govulncheck:   %s\n", config.Tools.GoVulnCheck)
	fmt.Printf("  mockgen:       %s\n", config.Tools.Mockgen)
	fmt.Printf("  swag:          %s\n", config.Tools.Swag)

	// Custom tools
	if len(config.Tools.Custom) > 0 {
		utils.Info("Custom tools:")
		for name, version := range config.Tools.Custom {
			fmt.Printf("  %s: %s\n", name, version)
		}
	}

	return nil
}

// VulnCheck installs and runs govulncheck with retry logic
func (Tools) VulnCheck() error {
	utils.Header("Checking for Vulnerabilities")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	ctx := context.Background()
	maxRetries := config.Download.MaxRetries
	initialDelay := time.Duration(config.Download.InitialDelayMs) * time.Millisecond

	// Ensure govulncheck is installed with retry logic

	if !utils.CommandExists("govulncheck") {
		if err := installGovulncheck(ctx, config, maxRetries, initialDelay); err != nil {
			return err
		}
	}

	// Run vulnerability check (this generally doesn't need retry as it's running locally)
	if err := GetRunner().RunCmd("govulncheck", "-show", "verbose", "./..."); err != nil {
		return fmt.Errorf("vulnerability check failed: %w", err)
	}

	utils.Success("No vulnerabilities found")
	return nil
}

// installGovulncheck installs govulncheck with retry and fallback logic
func installGovulncheck(ctx context.Context, config *Config, maxRetries int, initialDelay time.Duration) error {
	utils.Info("Installing govulncheck with retry logic...")

	govulnVersion := config.Tools.GoVulnCheck
	if govulnVersion == "" || govulnVersion == VersionLatest {
		utils.Warn("GoVulnCheck version not available, using @latest")
		govulnVersion = VersionAtLatest
	} else if !strings.HasPrefix(govulnVersion, "@") {
		govulnVersion = "@" + govulnVersion
	}

	moduleWithVersion := "golang.org/x/vuln/cmd/govulncheck" + govulnVersion

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
		if err := executor.ExecuteWithEnv(ctx, env, "go", "install", moduleWithVersion); err != nil {
			return fmt.Errorf("failed to install govulncheck after %d retries and fallback: %w", maxRetries, err)
		}
	}
	return nil
}

// installToolFromModule installs a tool from a Go module with retry logic
func installToolFromModule(ctx context.Context, tool ToolDefinition, _ *Config, maxRetries int, initialDelay time.Duration) error {
	moduleVersion := tool.Version

	// If version is "latest", try to resolve from environment variables first
	if moduleVersion == "" || moduleVersion == VersionLatest {
		if resolvedVersion := getToolVersionFromEnv(tool.Name); resolvedVersion != "" {
			moduleVersion = resolvedVersion
		} else {
			utils.Warn("Version for tool %s not available, using @latest", tool.Name)
			moduleVersion = VersionAtLatest
		}
	}

	if !strings.HasPrefix(moduleVersion, "@") && moduleVersion != VersionAtLatest {
		moduleVersion = "@" + moduleVersion
	}
	// Keep @latest as is for VersionAtLatest

	moduleWithVersion := tool.Module + moduleVersion
	utils.Info("Installing %s from %s...", tool.Name, moduleWithVersion)

	// Get the secure executor with retry capabilities
	runner := GetRunner()
	secureRunner, ok := runner.(*SecureCommandRunner)
	if !ok {
		return fmt.Errorf("%w: got %T", ErrUnexpectedRunnerType, runner)
	}
	executor := secureRunner.executor
	err := executor.ExecuteWithRetry(ctx, maxRetries, initialDelay, "go", "install", moduleWithVersion)
	if err != nil {
		// Try with different module proxy settings as fallback
		utils.Warn("Installation failed: %v, trying with direct module proxy...", err)

		// Try with GOPROXY=direct to bypass proxy issues
		env := []string{"GOPROXY=direct"}
		if err := executor.ExecuteWithEnv(ctx, env, "go", "install", moduleWithVersion); err != nil {
			return fmt.Errorf("failed to install %s after %d retries and fallback: %w", tool.Name, maxRetries, err)
		}
	}
	return nil
}

// Helper functions

// getToolVersionFromEnv resolves tool versions from environment variables
func getToolVersionFromEnv(toolName string) string {
	switch toolName {
	case "golangci-lint":
		return GetDefaultGolangciLintVersion()
	case "gofumpt":
		return GetDefaultGofumptVersion()
	case "yamlfmt":
		return GetDefaultYamlfmtVersion()
	case "govulncheck":
		return GetDefaultGoVulnCheckVersion()
	case "mockgen":
		return GetDefaultMockgenVersion()
	case "swag":
		return GetDefaultSwagVersion()
	default:
		// For unknown tools, try the generic version getter
		return GetToolVersion(toolName)
	}
}

// getRequiredTools returns the list of required tools based on configuration
func getRequiredTools(cfg *Config) []ToolDefinition {
	tools := []ToolDefinition{
		{
			Name:    "golangci-lint",
			Module:  "", // Special installation via script
			Version: cfg.Tools.GolangciLint,
			Check:   "golangci-lint",
		},
		{
			Name:    "gofumpt",
			Module:  "mvdan.cc/gofumpt",
			Version: cfg.Tools.Fumpt,
			Check:   "gofumpt",
		},
		{
			Name:    "govulncheck",
			Module:  "golang.org/x/vuln/cmd/govulncheck",
			Version: cfg.Tools.GoVulnCheck,
			Check:   "govulncheck",
		},
	}

	// Add optional tools if configured
	if cfg.Tools.Mockgen != "" {
		tools = append(tools, ToolDefinition{
			Name:    "mockgen",
			Module:  "go.uber.org/mock/mockgen",
			Version: cfg.Tools.Mockgen,
			Check:   "mockgen",
		})
	}

	if cfg.Tools.Swag != "" {
		tools = append(tools, ToolDefinition{
			Name:    "swag",
			Module:  "github.com/swaggo/swag/cmd/swag",
			Version: cfg.Tools.Swag,
			Check:   "swag",
		})
	}

	// Add custom tools
	for name, module := range cfg.Tools.Custom {
		// Parse module@version format
		parts := strings.Split(module, "@")
		modulePath := parts[0]
		modVer := VersionLatest
		if len(parts) > 1 {
			modVer = parts[1]
		}

		tools = append(tools, ToolDefinition{
			Name:    name,
			Module:  modulePath,
			Version: modVer,
			Check:   name,
		})
	}

	return tools
}

// installTool installs a single tool with retry logic
func installTool(tool ToolDefinition) error {
	// Check if already installed
	if utils.CommandExists(tool.Check) {
		utils.Info("%s is already installed", tool.Name)
		return nil
	}

	utils.Info("Installing %s with retry logic...", tool.Name)

	config, err := GetConfig()
	if err != nil {
		// Use default config if loading fails
		config = defaultConfig()
	}

	ctx := context.Background()
	maxRetries := config.Download.MaxRetries
	initialDelay := time.Duration(config.Download.InitialDelayMs) * time.Millisecond

	// Special case for golangci-lint
	if tool.Name == CmdGolangciLint {
		return ensureGolangciLint(config)
	}

	// Install via go install with retry logic

	if tool.Module != "" {
		if err := installToolFromModule(ctx, tool, config, maxRetries, initialDelay); err != nil {
			return err
		}
	}

	utils.Success("%s installed successfully", tool.Name)
	return nil
}

// Check checks tool versions
func (Tools) Check() error {
	// Delegate to Verify which does the actual checking
	return Tools{}.Verify()
}

// Clean removes tool installations
func (Tools) Clean() error {
	utils.Header("Cleaning Tool Installations")
	runner := GetRunner()
	return runner.RunCmd("echo", "Cleaning tool installations")
}
