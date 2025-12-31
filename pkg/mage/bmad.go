package mage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/magefile/mage/mg"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for err113 compliance
var (
	errNpmNotInstalled   = errors.New("npm not installed")
	errNpxNotInstalled   = errors.New("npx not installed")
	errBmadNotInstalled  = errors.New("bmad-method not installed")
	errBmadInstallFailed = errors.New("failed to install bmad prerequisites")
	errBmadVersionParse  = errors.New("failed to parse bmad version")
)

// Bmad namespace for BMAD CLI management tasks
type Bmad mg.Namespace

// Install installs BMAD prerequisites (npm, npx, bmad-method) with verbose output
func (Bmad) Install() error {
	utils.Header("Installing BMAD Prerequisites")

	config, err := GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Step 1: Check for npm
	utils.Info("Step 1/3: Checking for npm...")
	if !utils.CommandExists(CmdNpm) {
		utils.Warn("npm not found")
		utils.Info("Install Node.js from: https://nodejs.org/")
		utils.Info("Or use a version manager like nvm: https://github.com/nvm-sh/nvm")
		return errNpmNotInstalled
	}
	utils.Success("npm found")

	// Step 2: Check for npx
	utils.Info("Step 2/3: Checking for npx...")
	if !utils.CommandExists(CmdNpx) {
		utils.Warn("npx not found")
		utils.Info("npx should be included with npm. Try reinstalling Node.js")
		return errNpxNotInstalled
	}
	utils.Success("npx found")

	// Step 3: Install/run bmad-method
	utils.Info("Step 3/3: Running BMAD installer...")
	if err := installBmadCLI(config); err != nil {
		utils.Error("Failed to install BMAD: %v", err)
		return fmt.Errorf("%w: %w", errBmadInstallFailed, err)
	}
	utils.Success("BMAD installed successfully")

	// Verify installation
	if err := verifyBmadInstallation(config); err != nil {
		utils.Warn("Installation verification: %v", err)
		utils.Info("This may be expected on first install - BMAD creates the %s folder interactively", config.Bmad.ProjectDir)
	} else {
		utils.Success("BMAD project directory verified: %s", config.Bmad.ProjectDir)
	}

	utils.Success("BMAD fully installed and configured!")
	utils.Info("")
	utils.Info("Next steps:")
	utils.Info("  1. Load an agent (e.g., Analyst) in your IDE")
	utils.Info("  2. Run '*workflow-init' to initialize your project")
	utils.Info("  3. Follow the guided workflow to plan and build")
	utils.Info("")
	utils.Info("For more info: https://github.com/bmad-code-org/BMAD-METHOD")
	return nil
}

// Check verifies that BMAD is installed and working correctly
func (Bmad) Check() error {
	utils.Header("Checking BMAD Installation")

	config, err := GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Check prerequisites
	utils.Info("Checking prerequisites...")
	if prereqErr := checkBmadPrerequisites(); prereqErr != nil {
		return prereqErr
	}
	utils.Success("All required tools are installed")

	// Get version
	utils.Info("Getting BMAD version...")
	bmadVersion, versionErr := getBmadVersion(config)
	if versionErr != nil {
		utils.Warn("Could not determine version: %v", versionErr)
		utils.Info("This is normal if bmad-method hasn't been cached locally yet")
	} else {
		utils.Success("BMAD version: %s", bmadVersion)
	}

	// Check project directory
	projectDir := getBmadProjectDir(config)
	if _, err := os.Stat(projectDir); err != nil {
		utils.Warn("BMAD project directory not found: %s", projectDir)
		utils.Info("Run 'magex bmad:install' to initialize BMAD in your project")
	} else {
		utils.Success("BMAD project directory exists: %s", projectDir)
	}

	return nil
}

// Upgrade upgrades BMAD to the latest version
func (Bmad) Upgrade() error {
	utils.Header("BMAD Upgrade")

	config, err := GetConfig()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Step 1: Check prerequisites
	utils.Info("Step 1/4: Checking prerequisites...")
	if prereqErr := checkBmadPrerequisites(); prereqErr != nil {
		return prereqErr
	}
	utils.Success("Prerequisites verified")

	// Step 2: Get current version (if available)
	utils.Info("Step 2/4: Getting current BMAD version...")
	oldVersion, versionErr := getBmadVersion(config)
	if versionErr != nil {
		utils.Warn("Could not determine current version: %v", versionErr)
		oldVersion = statusUnknown
	} else {
		utils.Success("Current version: %s", oldVersion)
	}

	// Step 3: Clear npx cache and reinstall
	utils.Info("Step 3/4: Upgrading BMAD...")
	if err := upgradeBmadCLI(config); err != nil {
		utils.Error("Failed to upgrade BMAD: %v", err)
		return fmt.Errorf("failed to upgrade BMAD CLI: %w", err)
	}
	utils.Success("BMAD upgraded")

	// Step 4: Get new version
	utils.Info("Step 4/4: Getting new BMAD version...")
	newVersion, versionErr := getBmadVersion(config)
	if versionErr != nil {
		utils.Warn("Could not determine new version: %v", versionErr)
		newVersion = statusUnknown
	} else {
		utils.Success("New version: %s", newVersion)
	}

	// Summary
	printBmadUpgradeSummary(oldVersion, newVersion)

	return nil
}

// checkBmadPrerequisites verifies that required tools are installed
func checkBmadPrerequisites() error {
	if !utils.CommandExists(CmdNpm) {
		utils.Error("npm not found")
		utils.Info("Install Node.js from: https://nodejs.org/")
		return errNpmNotInstalled
	}

	if !utils.CommandExists(CmdNpx) {
		utils.Error("npx not found")
		utils.Info("npx should be included with npm. Try reinstalling Node.js")
		return errNpxNotInstalled
	}

	return nil
}

// getBmadVersion gets the current BMAD version using npm view
func getBmadVersion(config *Config) (string, error) {
	packageName := config.Bmad.PackageName
	if packageName == "" {
		packageName = DefaultBmadPackageName
	}

	versionTag := config.Bmad.VersionTag
	if versionTag == "" {
		versionTag = DefaultBmadVersionTag
	}

	// Use npm view to get the version of the package
	packageSpec := packageName + versionTag
	output, err := GetRunner().RunCmdOutput(CmdNpm, "view", packageSpec, "version")
	if err != nil {
		return "", fmt.Errorf("failed to get bmad version: %w", err)
	}

	version := strings.TrimSpace(output)
	if version == "" {
		return "", errBmadVersionParse
	}

	return version, nil
}

// getBmadProjectDir returns the configured project directory
func getBmadProjectDir(config *Config) string {
	if config.Bmad.ProjectDir != "" {
		return config.Bmad.ProjectDir
	}
	return DefaultBmadProjectDir
}

// installBmadCLI runs the BMAD installer using npx (interactive)
func installBmadCLI(config *Config) error {
	packageName := config.Bmad.PackageName
	if packageName == "" {
		packageName = DefaultBmadPackageName
	}

	versionTag := config.Bmad.VersionTag
	if versionTag == "" {
		versionTag = DefaultBmadVersionTag
	}

	// Run: npx bmad-method@alpha install (interactive - requires user input)
	packageSpec := packageName + versionTag
	utils.Info("Running: npx %s install", packageSpec)
	utils.Info("Note: This is an interactive installer - please follow the prompts")
	utils.Info("")

	return runInteractiveCmd(CmdNpx, packageSpec, "install")
}

// runInteractiveCmd executes a command with stdin/stdout/stderr connected for interactive use
func runInteractiveCmd(name string, args ...string) error {
	// Use background context - interactive commands shouldn't have a timeout

	cmd := exec.CommandContext(context.Background(), name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("command failed: %s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

// verifyBmadInstallation checks if the BMAD project directory exists
func verifyBmadInstallation(config *Config) error {
	projectDir := getBmadProjectDir(config)
	if _, err := os.Stat(projectDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: project directory %s not found", errBmadNotInstalled, projectDir)
		}
		return fmt.Errorf("failed to stat project directory: %w", err)
	}
	return nil
}

// upgradeBmadCLI runs the BMAD install command to upgrade an existing installation
// The bmad-method install command auto-detects existing installations and offers upgrade options
func upgradeBmadCLI(config *Config) error {
	packageName := config.Bmad.PackageName
	if packageName == "" {
		packageName = DefaultBmadPackageName
	}

	versionTag := config.Bmad.VersionTag
	if versionTag == "" {
		versionTag = DefaultBmadVersionTag
	}

	// Use bmad install command which detects existing installations
	packageSpec := packageName + versionTag
	utils.Info("Running: npx %s install", packageSpec)
	utils.Info("Note: The installer will detect your existing installation and offer upgrade options")
	utils.Info("")

	return runInteractiveCmd(CmdNpx, packageSpec, "install")
}

// printBmadUpgradeSummary prints the upgrade summary
func printBmadUpgradeSummary(oldVersion, newVersion string) {
	utils.Info("")
	utils.Success("BMAD upgrade complete!")
	utils.Info("")
	utils.Info("Version: %s -> %s", oldVersion, newVersion)
	utils.Info("")
	utils.Info("Note: BMAD configuration files in _bmad/ are preserved during upgrades")
	utils.Info("")
}
