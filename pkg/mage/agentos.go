package mage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/magefile/mage/mg"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for err113 compliance
var (
	errAgentOSNotInstalled       = errors.New("agent-os base not installed")
	errAgentOSProjectNotFound    = errors.New("agent-os project directory not found")
	errAgentOSInstallFailed      = errors.New("failed to install agent-os")
	errAgentOSUpgradeFailed      = errors.New("failed to upgrade agent-os")
	errAgentOSVersionParseFailed = errors.New("failed to parse agent-os version")
	errAgentOSAlreadyInstalled   = errors.New("agent-os already installed - use upgrade command")
	errAgentOSScriptNotFound     = errors.New("agent-os script not found")
	errAgentOSStandardsNotFound  = errors.New("standards directory not found")
	errCurlNotInstalled          = errors.New("curl not installed")
	errBashNotInstalled          = errors.New("bash not installed")
)

// AgentOS namespace for Agent OS CLI management tasks
// Agent OS is a system for spec-driven agentic development that provides
// structured workflows for AI coding agents with Claude Code integration.
type AgentOS mg.Namespace

// Install installs Agent OS (base + project) with verbose output.
// If the base installation (~/.agent-os) doesn't exist, it will be installed automatically.
// Then the project-level installation is performed to set up the current project.
func (AgentOS) Install() error {
	utils.Header("Installing Agent OS")

	config, err := GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Check for existing project installation - install is for fresh projects only
	projectDir := getAgentOSProjectDir(config)
	if _, statErr := os.Stat(projectDir); statErr == nil {
		utils.Error("Existing Agent OS installation detected: %s", projectDir)
		utils.Info("The install command is intended for fresh projects only")
		utils.Info("Use 'magex agentos:upgrade' to upgrade an existing installation")
		return errAgentOSAlreadyInstalled
	}

	// Step 1: Check prerequisites
	utils.Info("Step 1/4: Checking prerequisites...")
	if prereqErr := checkAgentOSPrerequisites(); prereqErr != nil {
		return prereqErr
	}
	utils.Success("Prerequisites verified (curl, bash)")

	// Step 2: Check/install base
	utils.Info("Step 2/4: Checking base installation...")
	if !isAgentOSBaseInstalled(config) {
		utils.Info("Base installation not found, installing...")
		if baseErr := installAgentOSBase(); baseErr != nil {
			utils.Error("Failed to install Agent OS base: %v", baseErr)
			return fmt.Errorf("%w: %w", errAgentOSInstallFailed, baseErr)
		}
		utils.Success("Agent OS base installed to ~/agent-os")
	} else {
		utils.Success("Agent OS base already installed")
	}

	// Step 3: Run project install
	utils.Info("Step 3/4: Installing Agent OS in current project...")
	if projectErr := runAgentOSProjectInstall(config); projectErr != nil {
		utils.Error("Failed to install Agent OS project: %v", projectErr)
		return fmt.Errorf("%w: %w", errAgentOSInstallFailed, projectErr)
	}
	utils.Success("Agent OS project installation complete")

	// Step 4: Verify installation
	utils.Info("Step 4/4: Verifying installation...")
	if verifyErr := verifyAgentOSInstallation(config); verifyErr != nil {
		utils.Warn("Verification warning: %v", verifyErr)
		utils.Info("This may be expected if some optional components were skipped")
	} else {
		utils.Success("Installation verified")
	}

	printAgentOSInstallSuccess(config)
	return nil
}

// Check verifies that Agent OS is installed and working correctly.
// It checks both the base installation and the project-level installation.
func (AgentOS) Check() error {
	utils.Header("Checking Agent OS Installation")

	config, err := GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Check prerequisites
	utils.Info("Checking prerequisites...")
	if prereqErr := checkAgentOSPrerequisites(); prereqErr != nil {
		return prereqErr
	}
	utils.Success("All required tools are installed (curl, bash)")

	// Check base installation
	utils.Info("Checking base installation...")
	if !isAgentOSBaseInstalled(config) {
		utils.Error("Agent OS base not installed")
		utils.Info("Run 'magex agentos:install' to install Agent OS")
		return errAgentOSNotInstalled
	}
	utils.Success("Base installation found: %s", getAgentOSHomePath(config))

	// Get version
	utils.Info("Getting Agent OS version...")
	version, versionErr := getAgentOSVersion(config)
	if versionErr != nil {
		utils.Warn("Could not determine version: %v", versionErr)
	} else {
		utils.Success("Agent OS version: %s", version)
	}

	// Check project installation
	projectDir := getAgentOSProjectDir(config)
	if _, statErr := os.Stat(projectDir); statErr != nil {
		utils.Warn("Project directory not found: %s", projectDir)
		utils.Info("Run 'magex agentos:install' to install Agent OS in this project")
	} else {
		utils.Success("Project directory exists: %s", projectDir)

		// Check for preserved directories
		specsDir := filepath.Join(projectDir, "specs")
		if _, err := os.Stat(specsDir); err == nil {
			utils.Success("Specs directory exists: %s", specsDir)
		}

		productDir := filepath.Join(projectDir, "product")
		if _, err := os.Stat(productDir); err == nil {
			utils.Success("Product directory exists: %s", productDir)
		}
	}

	// Check Claude Code integration
	claudeCommandsDir := filepath.Join(".claude", "commands", "agent-os")
	if _, err := os.Stat(claudeCommandsDir); err == nil {
		utils.Success("Claude Code commands installed: %s", claudeCommandsDir)
	} else if config.AgentOS.ClaudeCodeCommands {
		utils.Warn("Claude Code commands not found: %s", claudeCommandsDir)
	}

	claudeAgentsDir := filepath.Join(".claude", "agents", "agent-os")
	if _, err := os.Stat(claudeAgentsDir); err == nil {
		utils.Success("Claude Code agents installed: %s", claudeAgentsDir)
	} else if config.AgentOS.UseClaudeCodeSubagents {
		utils.Warn("Claude Code agents not found: %s", claudeAgentsDir)
	}

	return nil
}

// Upgrade safely upgrades Agent OS with automatic preservation of user data.
// The specs/ and product/ directories are preserved during upgrade.
//
//nolint:gocognit // Orchestration function with multiple sequential steps
func (AgentOS) Upgrade() error {
	utils.Header("Agent OS Automated Upgrade")

	config, err := GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Step 1: Check prerequisites
	utils.Info("Step 1/6: Checking prerequisites...")
	if prereqErr := checkAgentOSPrerequisites(); prereqErr != nil {
		return prereqErr
	}
	utils.Success("Prerequisites verified")

	// Step 2: Verify base installation
	utils.Info("Step 2/6: Verifying base installation...")
	if !isAgentOSBaseInstalled(config) {
		utils.Error("Agent OS base not installed")
		utils.Info("Run 'magex agentos:install' first")
		return errAgentOSNotInstalled
	}
	utils.Success("Base installation verified")

	// Step 3: Get current version
	utils.Info("Step 3/6: Getting current Agent OS version...")
	oldVersion, versionErr := getAgentOSVersion(config)
	if versionErr != nil {
		utils.Warn("Could not determine current version: %v", versionErr)
		oldVersion = statusUnknown
	} else {
		utils.Success("Current version: %s", oldVersion)
	}

	// Step 4: Update base installation
	utils.Info("Step 4/6: Updating base installation...")
	if err := updateAgentOSBase(); err != nil {
		utils.Error("Failed to update base: %v", err)
		return fmt.Errorf("%w: %w", errAgentOSUpgradeFailed, err)
	}
	utils.Success("Base installation updated")

	// Step 5: Run project update (preserves specs/ and product/)
	utils.Info("Step 5/6: Updating project installation...")
	utils.Info("Note: Your specs/ and product/ directories will be preserved")
	if err := runAgentOSProjectUpdate(config); err != nil {
		utils.Error("Failed to update project: %v", err)
		return fmt.Errorf("%w: %w", errAgentOSUpgradeFailed, err)
	}
	utils.Success("Project installation updated")

	// Step 6: Get new version
	utils.Info("Step 6/6: Getting new Agent OS version...")
	newVersion, versionErr := getAgentOSVersion(config)
	if versionErr != nil {
		utils.Warn("Could not determine new version: %v", versionErr)
		newVersion = statusUnknown
	} else {
		utils.Success("New version: %s", newVersion)
	}

	// Summary
	printAgentOSUpgradeSummary(oldVersion, newVersion)

	return nil
}

// checkAgentOSPrerequisites verifies that required tools are installed
func checkAgentOSPrerequisites() error {
	if !utils.CommandExists(CmdCurl) {
		utils.Error("curl not found")
		utils.Info("Install curl using your system package manager")
		utils.Info("  macOS: brew install curl")
		utils.Info("  Ubuntu/Debian: sudo apt install curl")
		return errCurlNotInstalled
	}

	if !utils.CommandExists(CmdBash) {
		utils.Error("bash not found")
		utils.Info("Install bash using your system package manager")
		return errBashNotInstalled
	}

	return nil
}

// getAgentOSHomePath returns the full path to the Agent OS base installation directory
func getAgentOSHomePath(config *Config) string {
	homeDir := config.AgentOS.HomeDir
	if homeDir == "" {
		homeDir = DefaultAgentOSHomeDir
	}

	// Expand home directory
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("~", homeDir)
	}
	return filepath.Join(home, homeDir)
}

// getAgentOSProjectDir returns the configured project directory
func getAgentOSProjectDir(config *Config) string {
	if config.AgentOS.BaseDir != "" {
		return config.AgentOS.BaseDir
	}
	return DefaultAgentOSBaseDir
}

// isAgentOSBaseInstalled checks if the base installation exists
func isAgentOSBaseInstalled(config *Config) bool {
	homePath := getAgentOSHomePath(config)
	configFile := filepath.Join(homePath, DefaultAgentOSConfigFile)
	_, err := os.Stat(configFile)
	return err == nil
}

// installAgentOSBase downloads and runs the base-install.sh script
func installAgentOSBase() error {
	utils.Info("Downloading and running Agent OS base installer...")
	utils.Info("Note: This is an interactive installer - please follow any prompts")
	utils.Info("")

	// Run: curl -sSL https://raw.githubusercontent.com/buildermethods/agent-os/main/scripts/base-install.sh | bash
	installURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/scripts/base-install.sh",
		DefaultAgentOSGitHubRepo, DefaultAgentOSBranch)

	return runInteractivePipedCmd(CmdCurl, []string{"-sSL", installURL}, CmdBash, []string{})
}

// updateAgentOSBase updates the base installation by re-running the installer
func updateAgentOSBase() error {
	utils.Info("Updating Agent OS base installation...")
	utils.Info("Note: This will update scripts and profiles")
	utils.Info("")

	// The base-install.sh script detects existing installations and offers update options
	installURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/scripts/base-install.sh",
		DefaultAgentOSGitHubRepo, DefaultAgentOSBranch)

	return runInteractivePipedCmd(CmdCurl, []string{"-sSL", installURL}, CmdBash, []string{})
}

// runAgentOSProjectInstall runs the project-install.sh script with config options
func runAgentOSProjectInstall(config *Config) error {
	homePath := getAgentOSHomePath(config)
	scriptPath := filepath.Join(homePath, "scripts", "project-install.sh")

	// Check if script exists
	if _, err := os.Stat(scriptPath); err != nil {
		return fmt.Errorf("%w: %s", errAgentOSScriptNotFound, scriptPath)
	}

	// Build arguments based on config
	args := buildAgentOSInstallArgs(config)

	utils.Info("Running: bash %s %s", scriptPath, strings.Join(args, " "))
	utils.Info("Note: This is an interactive installer - please follow any prompts")
	utils.Info("")

	return runInteractiveCmd(CmdBash, append([]string{scriptPath}, args...)...)
}

// runAgentOSProjectUpdate runs the project-update.sh script
func runAgentOSProjectUpdate(config *Config) error {
	homePath := getAgentOSHomePath(config)
	scriptPath := filepath.Join(homePath, "scripts", "project-update.sh")

	// Check if script exists
	if _, err := os.Stat(scriptPath); err != nil {
		return fmt.Errorf("%w: %s", errAgentOSScriptNotFound, scriptPath)
	}

	utils.Info("Running: bash %s", scriptPath)
	utils.Info("Note: This is an interactive updater - please follow any prompts")
	utils.Info("")

	return runInteractiveCmd(CmdBash, scriptPath)
}

// buildAgentOSInstallArgs builds command-line arguments for the install script
func buildAgentOSInstallArgs(config *Config) []string {
	var args []string

	// Profile selection
	if config.AgentOS.Profile != "" && config.AgentOS.Profile != "default" {
		args = append(args, "--profile", config.AgentOS.Profile)
	}

	// Claude Code commands
	if !config.AgentOS.ClaudeCodeCommands {
		args = append(args, "--no-claude-code-commands")
	}

	// Agent OS commands
	if config.AgentOS.AgentOSCommands {
		args = append(args, "--agent-os-commands")
	}

	// Subagents
	if !config.AgentOS.UseClaudeCodeSubagents {
		args = append(args, "--no-subagents")
	}

	// Standards as skills
	if config.AgentOS.StandardsAsSkills {
		args = append(args, "--standards-as-skills")
	}

	return args
}

// getAgentOSVersion reads the version from the config.yml file
func getAgentOSVersion(config *Config) (string, error) {
	homePath := getAgentOSHomePath(config)
	configFile := filepath.Join(homePath, DefaultAgentOSConfigFile)

	// Read config file
	//nolint:gosec // G304: configFile is derived from config, not user input
	data, err := os.ReadFile(configFile)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse version from YAML (simple regex approach for version field)
	re := regexp.MustCompile(`version:\s*["']?([^"'\s\n]+)["']?`)
	matches := re.FindSubmatch(data)
	if matches == nil {
		return "", errAgentOSVersionParseFailed
	}

	return string(matches[1]), nil
}

// verifyAgentOSInstallation checks if the installation was successful
func verifyAgentOSInstallation(config *Config) error {
	projectDir := getAgentOSProjectDir(config)

	// Check project directory exists
	if _, err := os.Stat(projectDir); err != nil {
		return fmt.Errorf("%w: %s", errAgentOSProjectNotFound, projectDir)
	}

	// Check for standards directory (should exist after install)
	standardsDir := filepath.Join(projectDir, "standards")
	if _, err := os.Stat(standardsDir); err != nil {
		return fmt.Errorf("%w: %s", errAgentOSStandardsNotFound, standardsDir)
	}

	return nil
}

// runInteractivePipedCmd runs a command piped to another command with stdin/stdout connected
func runInteractivePipedCmd(cmd1 string, args1 []string, cmd2 string, args2 []string) error {
	// Create the curl command
	// #nosec G204 -- cmd1/cmd2 are internal function parameters for piped commands
	curlCmd := exec.CommandContext(context.Background(), cmd1, args1...)

	// #nosec G204 -- cmd1/cmd2 are internal function parameters for piped commands
	bashCmd := exec.CommandContext(context.Background(), cmd2, args2...)

	// Pipe curl's stdout to bash's stdin
	pipe, err := curlCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create pipe: %w", err)
	}
	bashCmd.Stdin = pipe

	// Connect bash to terminal for interactive use
	bashCmd.Stdout = os.Stdout
	bashCmd.Stderr = os.Stderr

	// Start curl
	if err := curlCmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", cmd1, err)
	}

	// Start bash
	if err := bashCmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", cmd2, err)
	}

	// Wait for curl to finish
	if err := curlCmd.Wait(); err != nil {
		return fmt.Errorf("%s failed: %w", cmd1, err)
	}

	// Wait for bash to finish
	if err := bashCmd.Wait(); err != nil {
		return fmt.Errorf("%s failed: %w", cmd2, err)
	}

	return nil
}

// printAgentOSInstallSuccess prints success message and next steps
func printAgentOSInstallSuccess(config *Config) {
	utils.Info("")
	utils.Success("Agent OS fully installed and configured!")
	utils.Info("")
	utils.Info("Your project now has:")
	utils.Info("  - %s/           Agent OS configuration and standards", getAgentOSProjectDir(config))
	if config.AgentOS.ClaudeCodeCommands {
		utils.Info("  - .claude/commands/agent-os/  Claude Code commands")
	}
	if config.AgentOS.UseClaudeCodeSubagents {
		utils.Info("  - .claude/agents/agent-os/    Claude Code agents")
	}
	utils.Info("")
	utils.Info("Next steps:")
	utils.Info("  1. Start Claude Code in this project")
	utils.Info("  2. Run /write-spec to write your first specification")
	utils.Info("  3. Follow the Agent OS workflow:")
	utils.Info("     /write-spec -> /shape-spec -> /plan-product -> /create-tasks -> /implement-tasks")
	utils.Info("")
	utils.Info("For more info: https://github.com/buildermethods/agent-os")
}

// printAgentOSUpgradeSummary prints the upgrade summary
func printAgentOSUpgradeSummary(oldVersion, newVersion string) {
	utils.Info("")
	utils.Success("Agent OS upgrade complete!")
	utils.Info("")
	utils.Info("Version: %s -> %s", oldVersion, newVersion)
	utils.Info("")
	utils.Info("Note: Your specs/ and product/ directories were preserved during upgrade")
	utils.Info("")
}
