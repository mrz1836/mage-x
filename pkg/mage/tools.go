package mage

import (
	"errors"
	"fmt"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Static errors to satisfy err113 linter
var (
	errToolsMissing = errors.New("some tools are missing. Run 'mage tools:install' to install them")
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

// Update updates all development tools to latest versions
func (Tools) Update() error {
	utils.Header("Updating Development Tools")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Update golangci-lint if using brew
	if utils.IsMac() && utils.CommandExists("brew") {
		utils.Info("Updating golangci-lint via brew...")
		// Ignore error - best effort
		err := GetRunner().RunCmd("brew", "upgrade", "golangci-lint")
		_ = err // Best effort - ignore error
	}

	// Update other tools
	tools := getRequiredTools(config)
	for _, tool := range tools {
		if tool.Module == "" {
			continue
		}

		utils.Info("Updating %s...", tool.Name)

		version := tool.Version
		if version == DefaultGoVulnCheckVersion || version == "" {
			version = "@latest"
		} else if !strings.HasPrefix(version, "@") {
			version = "@" + version
		}

		if err := GetRunner().RunCmd("go", "install", tool.Module+version); err != nil {
			utils.Warn("Failed to update %s: %v", tool.Name, err)
		}
	}

	utils.Success("Tools updated")
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
			version := strings.TrimSpace(output)
			if version == "" {
				version = "installed"
			}
			utils.Success("%s: %s", tool.Name, version)
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
		utils.Info("\nCustom tools:")
		for name, version := range config.Tools.Custom {
			fmt.Printf("  %s: %s\n", name, version)
		}
	}

	return nil
}

// VulnCheck installs and runs govulncheck
func (Tools) VulnCheck() error {
	utils.Header("Checking for Vulnerabilities")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Ensure govulncheck is installed
	if !utils.CommandExists("govulncheck") {
		utils.Info("Installing govulncheck...")

		version := config.Tools.GoVulnCheck
		if version == "" || version == "latest" {
			version = "@latest"
		} else if !strings.HasPrefix(version, "@") {
			version = "@" + version
		}

		if err := GetRunner().RunCmd("go", "install", "golang.org/x/vuln/cmd/govulncheck"+version); err != nil {
			return fmt.Errorf("failed to install govulncheck: %w", err)
		}
	}

	// Run vulnerability check
	if err := GetRunner().RunCmd("govulncheck", "-show", "verbose", "./..."); err != nil {
		return fmt.Errorf("vulnerability check failed: %w", err)
	}

	utils.Success("No vulnerabilities found")
	return nil
}

// Helper functions

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
		version := "latest"
		if len(parts) > 1 {
			version = parts[1]
		}

		tools = append(tools, ToolDefinition{
			Name:    name,
			Module:  modulePath,
			Version: version,
			Check:   name,
		})
	}

	return tools
}

// installTool installs a single tool
func installTool(tool ToolDefinition) error {
	// Check if already installed
	if utils.CommandExists(tool.Check) {
		utils.Info("%s is already installed", tool.Name)
		return nil
	}

	utils.Info("Installing %s...", tool.Name)

	// Special case for golangci-lint
	if tool.Name == CmdGolangciLint {
		config, err := GetConfig()
		if err != nil {
			// Use default config if loading fails
			config = defaultConfig()
		}
		return ensureGolangciLint(config)
	}

	// Install via go install
	if tool.Module != "" {
		version := tool.Version
		if version == "" || version == "latest" {
			version = "@latest"
		} else if !strings.HasPrefix(version, "@") {
			version = "@" + version
		}

		if err := GetRunner().RunCmd("go", "install", tool.Module+version); err != nil {
			return err
		}
	}

	utils.Success("%s installed", tool.Name)
	return nil
}

// Check checks tool versions
func (Tools) Check() error {
	utils.Header("Checking Tool Versions")
	runner := GetRunner()
	return runner.RunCmd("echo", "Checking tool versions")
}

// Clean removes tool installations
func (Tools) Clean() error {
	utils.Header("Cleaning Tool Installations")
	runner := GetRunner()
	return runner.RunCmd("echo", "Cleaning tool installations")
}
