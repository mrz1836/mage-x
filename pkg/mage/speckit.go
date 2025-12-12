package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/magefile/mage/mg"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for err113 compliance
var (
	errUVNotInstalled       = errors.New("uv package manager not installed")
	errSpecifyNotInstalled  = errors.New("specify CLI not installed")
	errConstitutionNotFound = errors.New("constitution file not found")
	errBackupFailed         = errors.New("failed to backup constitution")
	errVersionParseFailed   = errors.New("failed to parse spec-kit version")
	errSpeckitInstallFailed = errors.New("failed to install speckit prerequisites")
)

// Speckit namespace for spec-kit CLI management tasks
type Speckit mg.Namespace

// Install installs spec-kit prerequisites (uv, uvx, specify-cli) with verbose output
func (Speckit) Install() error {
	utils.Header("Installing Spec-Kit Prerequisites")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Step 1: Check for uv
	utils.Info("Step 1/4: Checking for uv package manager...")
	if !utils.CommandExists(CmdUV) {
		utils.Warn("uv package manager not found")
		utils.Info("Install uv with: curl -LsSf https://astral.sh/uv/install.sh | sh")
		utils.Info("Or visit: https://docs.astral.sh/uv/getting-started/installation/")
		return errUVNotInstalled
	}
	utils.Success("uv package manager found")

	// Step 2: Check for uvx (included with uv)
	utils.Info("Step 2/4: Checking for uvx command runner...")
	if !utils.CommandExists(CmdUVX) {
		utils.Warn("uvx command runner not found")
		utils.Info("uvx should be included with uv installation")
		utils.Info("Try reinstalling uv: curl -LsSf https://astral.sh/uv/install.sh | sh")
		return errUVNotInstalled
	}
	utils.Success("uvx command runner found")

	// Step 3: Install/update specify-cli
	utils.Info("Step 3/4: Installing/updating specify-cli...")
	if err := installSpeckitCLI(config); err != nil {
		utils.Error("Failed to install specify-cli: %v", err)
		return fmt.Errorf("%w: %w", errSpeckitInstallFailed, err)
	}
	utils.Success("specify-cli installed successfully")

	// Step 4: Initialize project configuration
	utils.Info("Step 4/4: Initializing project configuration...")
	if err := upgradeSpeckitProjectConfig(config); err != nil {
		utils.Error("Failed to initialize project: %v", err)
		return fmt.Errorf("failed to initialize speckit project: %w", err)
	}
	utils.Success("Project configuration initialized")

	utils.Success("Spec-Kit fully installed and configured!")
	utils.Info("")
	utils.Info("You can now use spec-kit commands:")
	utils.Info("  specify check        - Verify installation")
	utils.Info("  specify constitution - Manage project constitution")
	utils.Info("  magex speckit:check  - Check installation status")
	return nil
}

// Check verifies that spec-kit is installed and working correctly
func (Speckit) Check() error {
	utils.Header("Checking Spec-Kit Installation")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Check prerequisites
	utils.Info("Checking prerequisites...")
	if prereqErr := checkSpeckitPrerequisites(); prereqErr != nil {
		return prereqErr
	}
	utils.Success("All required tools are installed")

	// Get version
	utils.Info("Getting spec-kit version...")
	speckitVersion, versionErr := getSpeckitVersion(config)
	if versionErr != nil {
		utils.Error("Failed to get version: %v", versionErr)
		return versionErr
	}
	utils.Success("Spec-kit version: %s", speckitVersion)

	// Verify with specify check
	utils.Info("Running specify check...")
	if verifyErr := verifySpeckitInstallation(); verifyErr != nil {
		utils.Error("Verification failed: %v", verifyErr)
		return verifyErr
	}
	utils.Success("Spec-kit is working correctly")

	// Check constitution
	constitutionPath := getSpeckitConstitutionPath(config)
	if _, err := os.Stat(constitutionPath); err != nil {
		utils.Warn("Constitution file not found: %s", constitutionPath)
		utils.Info("Run 'specify init' to initialize spec-kit in your project")
	} else {
		utils.Success("Constitution file exists: %s", constitutionPath)
	}

	return nil
}

// Upgrade safely upgrades spec-kit with automatic constitution backup and version tracking
//
//nolint:gocognit // Orchestration function with multiple sequential steps
func (Speckit) Upgrade() error {
	utils.Header("Spec-Kit Automated Upgrade")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Step 1: Check prerequisites
	utils.Info("Step 1/10: Checking prerequisites...")
	if prereqErr := checkSpeckitPrerequisites(); prereqErr != nil {
		return prereqErr
	}
	utils.Success("Prerequisites verified")

	// Step 2: Get current version
	utils.Info("Step 2/10: Getting current spec-kit version...")
	oldVersion, versionErr := getSpeckitVersion(config)
	if versionErr != nil {
		utils.Warn("Could not determine current version: %v", versionErr)
		oldVersion = statusUnknown
	} else {
		utils.Success("Current version: %s", oldVersion)
	}

	// Step 3: Backup constitution
	utils.Info("Step 3/10: Backing up constitution file...")
	backupPath, backupErr := backupSpeckitConstitution(config)
	if backupErr != nil {
		if errors.Is(backupErr, errConstitutionNotFound) {
			utils.Warn("Constitution file not found, skipping backup")
			backupPath = ""
		} else {
			utils.Error("Constitution backup failed: %v", backupErr)
			return backupErr
		}
	} else {
		utils.Success("Constitution backed up to: %s", backupPath)
	}

	// Steps 4-7: Perform upgrade actions
	newVersion, err := performSpeckitUpgradeActions(config)
	if err != nil {
		return err
	}

	// Step 8: Restore constitution (if we had a backup)
	if backupPath != "" {
		utils.Info("Step 8/10: Restoring constitution from backup...")
		if err := restoreSpeckitConstitution(config, backupPath); err != nil {
			utils.Error("Failed to restore constitution: %v", err)
			utils.Warn("Your constitution is safely backed up at: %s", backupPath)
			// Continue to version update
		} else {
			utils.Success("Constitution restored successfully")
		}
	} else {
		utils.Info("Step 8/10: Skipping restore (no backup needed)")
	}

	// Step 9: Update version file
	utils.Info("Step 9/10: Updating version tracking file...")
	if err := updateSpeckitVersionFile(config, oldVersion, newVersion, backupPath); err != nil {
		utils.Warn("Failed to update version file: %v", err)
	} else {
		utils.Success("Version tracking file updated")
	}

	// Step 10: Clean old backups
	utils.Info("Step 10/10: Cleaning old constitution backups...")
	if err := cleanOldSpeckitBackups(config); err != nil {
		utils.Warn("Failed to clean old backups: %v", err)
	} else {
		utils.Success("Old backups cleaned (keeping last %d)", config.Speckit.BackupsToKeep)
	}

	// Summary
	printSpeckitUpgradeSummary(oldVersion, newVersion, backupPath)

	return nil
}

// checkSpeckitPrerequisites verifies that required tools are installed
func checkSpeckitPrerequisites() error {
	if !utils.CommandExists(CmdUV) {
		utils.Error("uv package manager not found")
		utils.Info("Install with: curl -LsSf https://astral.sh/uv/install.sh | sh")
		return errUVNotInstalled
	}

	if !utils.CommandExists(CmdUVX) {
		utils.Error("uvx command runner not found")
		utils.Info("uvx should be included with uv installation")
		return errUVNotInstalled
	}

	if !utils.CommandExists(CmdSpecify) {
		utils.Error("specify CLI not found")
		utils.Info("Install with: uv tool install specify-cli")
		return errSpecifyNotInstalled
	}

	return nil
}

// getSpeckitVersion gets the current spec-kit version from uv tool list
func getSpeckitVersion(config *Config) (string, error) {
	output, err := GetRunner().RunCmdOutput(CmdUV, "tool", "list")
	if err != nil {
		return "", fmt.Errorf("failed to run uv tool list: %w", err)
	}

	cliName := config.Speckit.CLIName
	if cliName == "" {
		cliName = DefaultSpeckitCLIName
	}

	// Parse output: specify-cli v0.0.20
	re := regexp.MustCompile(regexp.QuoteMeta(cliName) + `\s+(v[\d.]+)`)
	matches := re.FindStringSubmatch(output)
	if matches == nil {
		return "", errVersionParseFailed
	}

	return matches[1], nil
}

// getSpeckitConstitutionPath returns the configured constitution path
func getSpeckitConstitutionPath(config *Config) string {
	if config.Speckit.ConstitutionPath != "" {
		return config.Speckit.ConstitutionPath
	}
	return DefaultSpeckitConstitutionPath
}

// installSpeckitCLI installs the specify-cli tool using uv
func installSpeckitCLI(config *Config) error {
	cliName := config.Speckit.CLIName
	if cliName == "" {
		cliName = DefaultSpeckitCLIName
	}

	utils.Info("Installing %s via uv...", cliName)
	return GetRunner().RunCmd(CmdUV, "tool", "install", cliName)
}

// verifySpeckitInstallation runs 'specify check' to verify the installation
func verifySpeckitInstallation() error {
	return GetRunner().RunCmd(CmdSpecify, "check")
}

// backupSpeckitConstitution creates a timestamped backup of the constitution file
func backupSpeckitConstitution(config *Config) (string, error) {
	constitutionPath := getSpeckitConstitutionPath(config)

	// Check if constitution exists
	if _, err := os.Stat(constitutionPath); err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%w: %s", errConstitutionNotFound, constitutionPath)
		}
		return "", fmt.Errorf("failed to stat constitution: %w", err)
	}

	// Get backup directory
	backupDir := config.Speckit.BackupDir
	if backupDir == "" {
		backupDir = DefaultSpeckitBackupDir
	}

	// Create backup directory if it doesn't exist
	//nolint:gosec // G301: Backup directory needs standard permissions for user access
	if err := os.MkdirAll(backupDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Generate backup filename with timestamp
	timestamp := time.Now().Format("20060102.150405")
	backupFilename := fmt.Sprintf("constitution.backup.%s.md", timestamp)
	backupPath := filepath.Join(backupDir, backupFilename)

	// Read constitution
	//nolint:gosec // G304: constitutionPath comes from config, not user input
	data, err := os.ReadFile(constitutionPath)
	if err != nil {
		return "", fmt.Errorf("failed to read constitution: %w", err)
	}

	// Write backup
	//nolint:gosec // G306: Constitution backups need to be readable for manual restoration
	if err := os.WriteFile(backupPath, data, 0o644); err != nil {
		return "", fmt.Errorf("%w: %w", errBackupFailed, err)
	}

	return backupPath, nil
}

// restoreSpeckitConstitution copies the backup file back to the constitution path
func restoreSpeckitConstitution(config *Config, backupPath string) error {
	constitutionPath := getSpeckitConstitutionPath(config)

	// Read backup
	//nolint:gosec // G304: Backup path is generated internally
	data, err := os.ReadFile(backupPath)
	if err != nil {
		return fmt.Errorf("failed to read backup: %w", err)
	}

	// Ensure the directory exists
	constitutionDir := filepath.Dir(constitutionPath)
	//nolint:gosec // G301: Constitution directory needs standard permissions
	if err := os.MkdirAll(constitutionDir, 0o755); err != nil {
		return fmt.Errorf("failed to create constitution directory: %w", err)
	}

	// Write to constitution path
	//nolint:gosec // G306: Constitution needs to be readable
	if err := os.WriteFile(constitutionPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write constitution: %w", err)
	}

	return nil
}

// performSpeckitUpgradeActions handles the core upgrade steps (CLI upgrade, verification, config upgrade)
func performSpeckitUpgradeActions(config *Config) (string, error) {
	// Step 4: Upgrade uv tool
	utils.Info("Step 4/10: Upgrading spec-kit CLI...")
	if err := upgradeSpeckitUVTool(config); err != nil {
		utils.Error("Failed to upgrade spec-kit CLI: %v", err)
		return "", err
	}
	utils.Success("Spec-kit CLI upgraded")

	// Step 5: Verify installation
	utils.Info("Step 5/10: Verifying spec-kit installation...")
	if err := verifySpeckitInstallation(); err != nil {
		utils.Warn("Verification warning: %v", err)
	} else {
		utils.Success("Spec-kit installation verified")
	}

	// Step 6: Get new version
	utils.Info("Step 6/10: Getting new spec-kit version...")
	newVersion, versionErr := getSpeckitVersion(config)
	if versionErr != nil {
		utils.Warn("Could not determine new version: %v", versionErr)
		newVersion = statusUnknown
	} else {
		utils.Success("New version: %s", newVersion)
	}

	// Step 7: Upgrade project configuration
	utils.Info("Step 7/10: Upgrading project configuration...")
	if configErr := upgradeSpeckitProjectConfig(config); configErr != nil {
		utils.Error("Failed to upgrade project configuration: %v", configErr)
		return "", configErr
	}
	utils.Success("Project configuration upgraded")

	return newVersion, nil
}

// upgradeSpeckitUVTool upgrades the spec-kit CLI using uv tool upgrade
func upgradeSpeckitUVTool(config *Config) error {
	cliName := config.Speckit.CLIName
	if cliName == "" {
		cliName = DefaultSpeckitCLIName
	}

	return GetRunner().RunCmd(CmdUV, "tool", "upgrade", cliName)
}

// upgradeSpeckitProjectConfig upgrades the project configuration using uvx
// It attempts to use gh CLI for authentication first to avoid rate limits
func upgradeSpeckitProjectConfig(config *Config) error {
	gitHubRepo := config.Speckit.GitHubRepo
	if gitHubRepo == "" {
		gitHubRepo = DefaultSpeckitGitHubRepo
	}

	aiProvider := config.Speckit.AIProvider
	if aiProvider == "" {
		aiProvider = DefaultSpeckitAIProvider
	}

	args := []string{
		"--from",
		gitHubRepo,
		CmdSpecify,
		"init",
		"--here",
		"--ai",
		aiProvider,
		"--force",
	}

	// Try to get GitHub token from gh CLI for higher rate limits
	// Only set if GH_TOKEN/GITHUB_TOKEN not already set
	if os.Getenv("GH_TOKEN") == "" && os.Getenv("GITHUB_TOKEN") == "" {
		if ghToken := getGitHubTokenFromGH(); ghToken != "" {
			utils.Info("Using GitHub CLI authentication for higher rate limits")
			os.Setenv("GH_TOKEN", ghToken) //nolint:errcheck,gosec // Best effort, G104 acknowledged
			defer os.Unsetenv("GH_TOKEN")  //nolint:errcheck // Cleanup
		} else {
			utils.Warn("GitHub CLI not available - using unauthenticated access (60 requests/hour)")
			utils.Info("Install gh CLI and run 'gh auth login' for higher rate limits")
		}
	}

	return GetRunner().RunCmd(CmdUVX, args...)
}

// getGitHubTokenFromGH attempts to get a GitHub token from the gh CLI
func getGitHubTokenFromGH() string {
	// Check if gh is available
	if !utils.CommandExists("gh") {
		return ""
	}

	// Try to get the auth token
	output, err := GetRunner().RunCmdOutput("gh", "auth", "token")
	if err != nil {
		return ""
	}

	return strings.TrimSpace(output)
}

// updateSpeckitVersionFile writes version tracking information to the version file
func updateSpeckitVersionFile(config *Config, oldVersion, newVersion, backupPath string) error {
	versionFile := config.Speckit.VersionFile
	if versionFile == "" {
		versionFile = DefaultSpeckitVersionFile
	}

	// Ensure the directory exists
	versionDir := filepath.Dir(versionFile)
	//nolint:gosec // G301: Version file directory needs standard permissions
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		return fmt.Errorf("failed to create version file directory: %w", err)
	}

	timestamp := time.Now().Format(time.RFC3339)
	content := fmt.Sprintf(`version=%s
last_upgrade=%s
constitution_backup=%s
previous_version=%s
upgrade_method=automated
`, newVersion, timestamp, backupPath, oldVersion)

	//nolint:gosec // G306: Version file needs to be readable for user inspection
	if err := os.WriteFile(versionFile, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write version file: %w", err)
	}

	return nil
}

// cleanOldSpeckitBackups removes old constitution backups, keeping only the most recent N
//
//nolint:gocognit,gocyclo // File management with multiple conditional checks
func cleanOldSpeckitBackups(config *Config) error {
	backupDir := config.Speckit.BackupDir
	if backupDir == "" {
		backupDir = DefaultSpeckitBackupDir
	}

	backupsToKeep := config.Speckit.BackupsToKeep
	if backupsToKeep <= 0 {
		backupsToKeep = DefaultSpeckitBackupsToKeep
	}

	// Check if backup directory exists
	if _, err := os.Stat(backupDir); err != nil {
		if os.IsNotExist(err) {
			return nil // No backups to clean
		}
		return fmt.Errorf("failed to stat backup directory: %w", err)
	}

	// Read backup directory
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	// Filter for constitution backup files
	var backups []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), "constitution.backup.") && strings.HasSuffix(entry.Name(), ".md") {
			backups = append(backups, entry.Name())
		}
	}

	// Sort backups by name (timestamp-based names sort chronologically)
	sort.Strings(backups)

	// Delete old backups if we have more than backupsToKeep
	if len(backups) <= backupsToKeep {
		return nil // Nothing to clean
	}

	toDelete := len(backups) - backupsToKeep
	for i := 0; i < toDelete; i++ {
		backupPath := filepath.Join(backupDir, backups[i])
		if err := os.Remove(backupPath); err != nil {
			utils.Warn("Failed to delete old backup: %s", backups[i])
		}
	}

	return nil
}

// printSpeckitUpgradeSummary prints the upgrade summary
func printSpeckitUpgradeSummary(oldVersion, newVersion, backupPath string) {
	utils.Info("")
	utils.Success("Spec-Kit upgrade complete!")
	utils.Info("")
	utils.Info("Version: %s -> %s", oldVersion, newVersion)
	if backupPath != "" {
		utils.Info("Constitution backup: %s", backupPath)
	}
	utils.Warn("Review .specify/memory/constitution.md for any changes")
	utils.Info("")
}
