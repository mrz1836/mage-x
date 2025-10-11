package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for dependency management
var (
	errDependencyNameRequired = errors.New("dependency name required: mage deps:why github.com/pkg/errors")
	errModuleNameRequired     = errors.New("module name required: mage deps:init github.com/user/project")
	errGoModAlreadyExists     = errors.New("go.mod already exists")
	errNoGoModFound           = errors.New("no go.mod file found - govulncheck requires a Go module. Run 'go mod init <module-name>' to initialize a module, or navigate to a directory with a go.mod file")
	errNoGoModFoundSimple     = errors.New("no go.mod file found - govulncheck requires a Go module. Run 'go mod init <module-name>' to initialize a module")
	errNoGoPackagesFound      = errors.New("no Go packages found in current directory. Ensure you're in a directory with Go source files")
)

// Deps namespace for dependency management tasks
type Deps mg.Namespace

// Default manages default dependencies
func (Deps) Default() error {
	utils.Header("Managing Dependencies")
	runner := GetRunner()
	return runner.RunCmd("go", "mod", "download")
}

// Download downloads all dependencies
func (Deps) Download() error {
	utils.Header("Downloading Dependencies")

	start := time.Now()
	if err := GetRunner().RunCmd("go", "mod", "download"); err != nil {
		return fmt.Errorf("failed to download dependencies: %w", err)
	}

	utils.Success("Dependencies downloaded in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// Tidy cleans up go.mod and go.sum
func (Deps) Tidy() error {
	utils.Header("Tidying Dependencies")

	if err := GetRunner().RunCmd("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("failed to tidy dependencies: %w", err)
	}

	utils.Success("Dependencies tidied")
	return nil
}

// Update updates all dependencies
func (Deps) Update() error {
	return (Deps{}).UpdateWithArgs()
}

// UpdateWithArgs updates all dependencies with optional parameters
// Supports:
//   - allow-major: Allow major version updates (default: false)
//   - stable-only: Force downgrade from pre-release to stable versions (default: false)
//   - verbose: Show detailed output for all package updates (default: false)
func (Deps) UpdateWithArgs(argsList ...string) error {
	utils.Header("Updating Dependencies")

	// Parse command-line parameters
	params := utils.ParseParams(argsList)
	allowMajor := utils.IsParamTrue(params, "allow-major")
	stableOnly := utils.IsParamTrue(params, "stable-only")
	verbose := utils.IsParamTrue(params, "verbose")

	if allowMajor {
		utils.Info("Major version updates are ENABLED")
	} else {
		utils.Info("Major version updates are DISABLED (use 'allow-major' to enable)")
	}

	if stableOnly {
		utils.Info("Stable-only mode ENABLED - will downgrade pre-release versions to stable")
	} else {
		utils.Info("Pre-release versions will be preserved (use 'stable-only' to force downgrade to stable)")
	}

	// Capture initial state of all dependencies
	initialDeps, err := captureAllDependencies()
	if err != nil {
		return fmt.Errorf("failed to capture initial dependencies: %w", err)
	}

	// Get list of direct dependencies
	output, err := GetRunner().RunCmdOutput("go", "list", "-m", "-f", "{{if not .Indirect}}{{.Path}}{{end}}", "all")
	if err != nil {
		return fmt.Errorf("failed to list dependencies: %w", err)
	}

	deps := strings.Split(strings.TrimSpace(output), "\n")
	updatedCount := 0
	checkedCount := 0
	skippedMajorCount := 0
	var skippedMajorUpdates []string

	for _, dep := range deps {
		dep = strings.TrimSpace(dep)
		if dep == "" || strings.Contains(dep, "=>") {
			continue
		}

		utils.Info("Checking %s...", dep)
		checkedCount++

		// Get current version
		currentOutput, currentErr := GetRunner().RunCmdOutput("go", "list", "-m", dep)
		if currentErr != nil {
			utils.Warn("Failed to get current version for %s: %v", dep, currentErr)
			continue
		}

		currentParts := strings.Fields(currentOutput)
		if len(currentParts) < 2 {
			continue
		}
		currentVersion := currentParts[1]

		// Get latest version
		latestOutput, latestErr := GetRunner().RunCmdOutput("go", "list", "-m", "-versions", dep)
		if latestErr != nil {
			utils.Warn("Failed to check %s: %v", dep, latestErr)
			continue
		}

		parts := strings.Fields(latestOutput)
		if len(parts) < 2 {
			continue
		}

		latestVersion := parts[len(parts)-1]

		// Check if this would be a major version update
		isMajorUpdate := isMajorVersionUpdate(currentVersion, latestVersion)

		if isMajorUpdate && !allowMajor {
			// Skip major version update but notify user
			utils.Warn("Skipping major version update: %s %s → %s (use 'allow-major' to update)", dep, currentVersion, latestVersion)
			skippedMajorCount++
			skippedMajorUpdates = append(skippedMajorUpdates, fmt.Sprintf("%s %s → %s", dep, currentVersion, latestVersion))
			continue
		}

		// Check if current version is already newer than the "latest" stable version
		// This handles pre-release versions that are newer than stable versions
		if !stableOnly && isVersionNewer(currentVersion, latestVersion) {
			utils.Info("Skipping %s: current version %s is newer than latest stable %s", dep, currentVersion, latestVersion)
			continue
		} else if stableOnly && isVersionNewer(currentVersion, latestVersion) {
			utils.Warn("Downgrading %s: %s → %s (stable-only mode)", dep, currentVersion, latestVersion)
		}

		// Update to latest (or within major version if allow-major is false)
		targetVersion := latestVersion
		if !allowMajor && isMajorUpdate {
			// This shouldn't happen due to the check above, but be safe
			continue
		}

		if updateErr := GetRunner().RunCmd("go", "get", "-u", dep+"@"+targetVersion); updateErr != nil {
			utils.Warn("Failed to update %s: %v", dep, updateErr)
		} else {
			if currentVersion != targetVersion {
				utils.Info("Updated %s: %s → %s", dep, currentVersion, targetVersion)
				updatedCount++
			}
		}
	}

	// Update indirect dependencies
	utils.Info("Updating indirect dependencies...")
	if err = GetRunner().RunCmd("go", "get", "-u", "./..."); err != nil {
		utils.Warn("Failed to update indirect dependencies: %v", err)
	}

	// Tidy after updates
	if tidyErr := (Deps{}).Tidy(); tidyErr != nil {
		return tidyErr
	}

	// Capture final state and show comprehensive summary
	finalDeps, err := captureAllDependencies()
	if err != nil {
		utils.Warn("Failed to capture final dependencies for detailed summary: %v", err)
		// Fall back to simple summary
		if updatedCount == 0 && skippedMajorCount == 0 {
			utils.Success("Checked %d dependencies - all are up to date", checkedCount)
		} else {
			if updatedCount > 0 {
				utils.Success("Checked %d dependencies, updated %d", checkedCount, updatedCount)
			}
		}
	} else {
		// Show detailed summary with all changes
		showUpdateSummary(initialDeps, finalDeps, updatedCount, checkedCount, verbose)
	}

	if skippedMajorCount > 0 {
		utils.Info("\n📋 Skipped %d major version updates:", skippedMajorCount)
		for _, update := range skippedMajorUpdates {
			utils.Info("  • %s", update)
		}
		utils.Info("\nTo allow major version updates, run: magex deps:update allow-major")
	}

	return nil
}

// Clean cleans the module cache
func (Deps) Clean() error {
	utils.Header("Cleaning Module Cache")

	if err := GetRunner().RunCmd("go", "clean", "-modcache"); err != nil {
		return fmt.Errorf("failed to clean module cache: %w", err)
	}

	utils.Success("Module cache cleaned")
	return nil
}

// Graph shows the dependency graph
func (Deps) Graph() error {
	utils.Header("Dependency Graph")

	return GetRunner().RunCmd("go", "mod", "graph")
}

// Why shows why a dependency is needed
func (Deps) Why(dep string) error {
	if dep == "" {
		return errDependencyNameRequired
	}

	utils.Header(fmt.Sprintf("Why %s?", dep))

	return GetRunner().RunCmd("go", "mod", "why", dep)
}

// Verify verifies dependencies
func (Deps) Verify() error {
	utils.Header("Verifying Dependencies")

	if err := GetRunner().RunCmd("go", "mod", "verify"); err != nil {
		return fmt.Errorf("dependency verification failed: %w", err)
	}

	utils.Success("All dependencies verified")
	return nil
}

// VulnCheck checks for known vulnerabilities
func (Deps) VulnCheck() error {
	// Delegate to tools:vulncheck
	return Tools{}.VulnCheck()
}

// List lists all dependencies
func (Deps) List() error {
	utils.Header("Dependencies")

	// Direct dependencies
	utils.Info("Direct dependencies:")
	if err := GetRunner().RunCmd("go", "list", "-m", "-f", "{{if not .Indirect}}{{.Path}} {{.Version}}{{end}}", "all"); err != nil {
		return err
	}

	// Show count of indirect dependencies
	output, err := GetRunner().RunCmdOutput("go", "list", "-m", "-f", "{{if .Indirect}}1{{end}}", "all")
	if err == nil {
		count := strings.Count(output, "1")
		utils.Info("(%d indirect dependencies)", count)
	}

	return nil
}

// Outdated shows outdated dependencies
func (Deps) Outdated() error {
	utils.Header("Checking for Outdated Dependencies")

	// Get all direct dependencies
	output, err := GetRunner().RunCmdOutput("go", "list", "-m", "-u", "-f", "{{if and (not .Indirect) .Update}}{{.Path}} {{.Version}} → {{.Update.Version}}{{end}}", "all")
	if err != nil {
		return fmt.Errorf("failed to check outdated dependencies: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	outdated := []string{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			outdated = append(outdated, line)
		}
	}

	if len(outdated) == 0 {
		utils.Success("All dependencies are up to date!")
	} else {
		utils.Info("Found %d outdated dependencies:\n", len(outdated))
		for _, dep := range outdated {
			utils.Info("  %s", dep)
		}
		utils.Info("Run 'magex deps:update' to update all dependencies")
	}

	return nil
}

// Vendor vendors all dependencies
func (Deps) Vendor() error {
	utils.Header("Vendoring Dependencies")

	if err := GetRunner().RunCmd("go", "mod", "vendor"); err != nil {
		return fmt.Errorf("failed to vendor dependencies: %w", err)
	}

	utils.Success("Dependencies vendored to vendor/")
	return nil
}

// Init initializes a new go module
func (Deps) Init(module string) error {
	if module == "" {
		return errModuleNameRequired
	}

	utils.Header("Initializing Go Module")

	if utils.FileExists("go.mod") {
		return errGoModAlreadyExists
	}

	if err := GetRunner().RunCmd("go", "mod", "init", module); err != nil {
		return fmt.Errorf("failed to initialize module: %w", err)
	}

	utils.Success("Module initialized: %s", module)
	return nil
}

// Licenses shows dependency licenses
func (Deps) Licenses() error {
	utils.Header("Checking Dependency Licenses")
	runner := GetRunner()
	return runner.RunCmd("echo", "Checking dependency licenses")
}

// Check checks for updates
func (Deps) Check() error {
	utils.Header("Checking Dependencies")
	runner := GetRunner()
	return runner.RunCmd("go", "list", "-m", "-u", "all")
}

// isMajorVersionUpdate checks if updating from currentVersion to latestVersion
// would be a major version update (v1 to v2, v2 to v3, etc.)
func isMajorVersionUpdate(currentVersion, latestVersion string) bool {
	// Handle common Go version formats

	// Remove common prefixes
	current := strings.TrimPrefix(currentVersion, "v")
	latest := strings.TrimPrefix(latestVersion, "v")

	// Split by dots to get major.minor.patch
	currentParts := strings.Split(current, ".")
	latestParts := strings.Split(latest, ".")

	// Need at least major version in both
	if len(currentParts) == 0 || len(latestParts) == 0 {
		return false
	}

	// Extract major versions (first part before any non-digit)
	currentMajor := extractMajorVersion(currentParts[0])
	latestMajor := extractMajorVersion(latestParts[0])

	// Compare major versions
	return latestMajor > currentMajor
}

// extractMajorVersion extracts the major version number from a version string part
func extractMajorVersion(versionPart string) int {
	// Find the first non-digit character and extract everything before it
	majorStr := ""
	for _, r := range versionPart {
		if r >= '0' && r <= '9' {
			majorStr += string(r)
		} else {
			break
		}
	}

	if majorStr == "" {
		return 0
	}

	// Convert to integer
	major := 0
	for _, r := range majorStr {
		major = major*10 + int(r-'0')
	}

	return major
}

// compareVersions compares two version strings and returns:
// -1 if v1 < v2
//
//	0 if v1 == v2
//	1 if v1 > v2
//
// Handles pre-release versions according to semantic versioning rules
func compareVersions(v1, v2 string) int {
	// Normalize versions by removing 'v' prefix
	v1 = strings.TrimPrefix(v1, "v")
	v2 = strings.TrimPrefix(v2, "v")

	if v1 == v2 {
		return 0
	}

	// Split version into parts (major.minor.patch-prerelease)
	parts1 := parseVersion(v1)
	parts2 := parseVersion(v2)

	// Compare major.minor.patch
	for i := 0; i < 3; i++ {
		if parts1.numbers[i] > parts2.numbers[i] {
			return 1
		} else if parts1.numbers[i] < parts2.numbers[i] {
			return -1
		}
	}

	// If major.minor.patch are equal, compare pre-release
	// According to semver: pre-release versions have lower precedence than normal versions
	if parts1.prerelease == "" && parts2.prerelease != "" {
		return 1 // v1.2.3 > v1.2.3-alpha
	}
	if parts1.prerelease != "" && parts2.prerelease == "" {
		return -1 // v1.2.3-alpha < v1.2.3
	}
	if parts1.prerelease != "" && parts2.prerelease != "" {
		// Both have pre-release, compare lexicographically
		if parts1.prerelease > parts2.prerelease {
			return 1
		} else if parts1.prerelease < parts2.prerelease {
			return -1
		}
	}

	return 0
}

// versionParts represents parsed version components
type versionParts struct {
	numbers    [3]int // major, minor, patch
	prerelease string // everything after the first dash
}

// parseVersion parses a version string into components
func parseVersion(version string) versionParts {
	var parts versionParts

	// Split at first dash to separate version from pre-release
	mainPart := version
	if dashIndex := strings.Index(version, "-"); dashIndex != -1 {
		mainPart = version[:dashIndex]
		parts.prerelease = version[dashIndex+1:]
	}

	// Parse major.minor.patch
	numberParts := strings.Split(mainPart, ".")
	for i := 0; i < 3 && i < len(numberParts); i++ {
		if num, err := strconv.Atoi(numberParts[i]); err == nil {
			parts.numbers[i] = num
		}
	}

	return parts
}

// isVersionNewer checks if version v1 is newer than version v2
func isVersionNewer(v1, v2 string) bool {
	return compareVersions(v1, v2) > 0
}

// dependencyInfo holds information about a dependency
type dependencyInfo struct {
	Path    string
	Version string
}

// captureAllDependencies captures the current state of all dependencies (direct and indirect)
func captureAllDependencies() (map[string]dependencyInfo, error) {
	output, err := GetRunner().RunCmdOutput("go", "list", "-m", "-f", "{{.Path}} {{.Version}}", "all")
	if err != nil {
		return nil, err
	}

	deps := make(map[string]dependencyInfo)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			deps[parts[0]] = dependencyInfo{
				Path:    parts[0],
				Version: parts[1],
			}
		}
	}

	return deps, nil
}

// packageGroup represents a group of related packages
type packageGroup struct {
	Name    string
	Updates []dependencyInfo
	IsGroup bool
}

// groupRelatedPackages groups packages by common prefixes (e.g., AWS SDK packages)
func groupRelatedPackages(changes []dependencyInfo) []packageGroup {
	// Common package prefixes that should be grouped
	commonPrefixes := []string{
		"github.com/aws/aws-sdk-go-v2",
		"github.com/aws/smithy-go",
		"google.golang.org/genproto",
		"golang.org/x",
	}

	groups := make(map[string][]dependencyInfo)
	ungrouped := []dependencyInfo{}

	// Sort changes into groups
	for _, change := range changes {
		grouped := false
		for _, prefix := range commonPrefixes {
			if strings.HasPrefix(change.Path, prefix) {
				groups[prefix] = append(groups[prefix], change)
				grouped = true
				break
			}
		}
		if !grouped {
			ungrouped = append(ungrouped, change)
		}
	}

	result := []packageGroup{}

	// Add grouped packages
	for prefix, packageList := range groups {
		if len(packageList) > 1 {
			// Multiple packages in group - show as a group
			result = append(result, packageGroup{
				Name:    prefix,
				Updates: packageList,
				IsGroup: true,
			})
		} else {
			// Single package - show ungrouped
			ungrouped = append(ungrouped, packageList...)
		}
	}

	// Add ungrouped packages
	for _, pkg := range ungrouped {
		result = append(result, packageGroup{
			Name:    pkg.Path,
			Updates: []dependencyInfo{pkg},
			IsGroup: false,
		})
	}

	return result
}

// showUpdateSummary displays a comprehensive summary of all dependency changes
func showUpdateSummary(initial, final map[string]dependencyInfo, directUpdates, checkedCount int, verbose bool) {
	// Find all changes
	changes := []dependencyInfo{}

	for path, finalDep := range final {
		if initialDep, exists := initial[path]; exists {
			if initialDep.Version != finalDep.Version {
				changes = append(changes, dependencyInfo{
					Path:    path,
					Version: fmt.Sprintf("%s → %s", initialDep.Version, finalDep.Version),
				})
			}
		}
	}

	totalChanges := len(changes)

	if totalChanges == 0 {
		utils.Success("Checked %d dependencies - all are up to date", checkedCount)
		return
	}

	// Show summary
	if directUpdates == totalChanges {
		utils.Success("Checked %d dependencies, updated %d", checkedCount, totalChanges)
	} else {
		utils.Success("Checked %d dependencies, updated %d direct dependencies (%d total packages affected)",
			checkedCount, directUpdates, totalChanges)
	}

	if verbose {
		// Verbose mode: show all changes
		utils.Info("\nDetailed changes:")
		for _, change := range changes {
			utils.Info("  • %s: %s", change.Path, change.Version)
		}
	} else {
		// Compact mode: group related packages
		groups := groupRelatedPackages(changes)

		if len(groups) > 0 {
			utils.Info("\nUpdated packages:")
			for _, group := range groups {
				if group.IsGroup {
					utils.Info("  • %s (%d packages)", group.Name, len(group.Updates))
					for _, update := range group.Updates {
						utils.Info("    - %s: %s",
							strings.TrimPrefix(update.Path, group.Name+"/"),
							update.Version)
					}
				} else {
					utils.Info("  • %s: %s", group.Updates[0].Path, group.Updates[0].Version)
				}
			}
		}
	}
}

// Audit performs comprehensive dependency audit
func (Deps) Audit() error {
	utils.Header("Auditing Dependencies for Vulnerabilities")

	config, err := GetConfig()
	if err != nil {
		return err
	}

	// Ensure govulncheck is installed
	if !utils.CommandExists("govulncheck") {
		utils.Info("Installing govulncheck...")

		vulnVersion := config.Tools.GoVulnCheck
		if vulnVersion == "" || vulnVersion == VersionLatest {
			vulnVersion = VersionAtLatest
		} else if !strings.HasPrefix(vulnVersion, "@") {
			vulnVersion = "@" + vulnVersion
		}

		if err := GetRunner().RunCmd("go", "install", "golang.org/x/vuln/cmd/govulncheck"+vulnVersion); err != nil {
			return fmt.Errorf("failed to install govulncheck: %w", err)
		}
	}

	// Check if we're in a Go module directory
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		return errNoGoModFound
	}

	// Run vulnerability check on dependencies
	utils.Info("Scanning dependencies for known vulnerabilities...")

	// Try to use govulncheck from PATH first, then fall back to GOPATH/bin
	govulncheckCmd := findGovulncheckCommand()

	if err := GetRunner().RunCmd(govulncheckCmd, "-show", "verbose", "./..."); err != nil {
		// Provide more helpful error messages for common failures
		errMsg := err.Error()
		if strings.Contains(errMsg, "no go.mod file") {
			return errNoGoModFoundSimple
		}
		if strings.Contains(errMsg, "no packages") {
			return errNoGoPackagesFound
		}
		return fmt.Errorf("vulnerability check failed: %w", err)
	}

	utils.Success("Dependency audit completed")
	return nil
}

// findGovulncheckCommand finds the govulncheck command, trying PATH first, then GOPATH/bin
func findGovulncheckCommand() string {
	if utils.CommandExists("govulncheck") {
		return "govulncheck"
	}

	// Check if it's in GOPATH/bin
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		// Default GOPATH
		home, err := os.UserHomeDir()
		if err == nil {
			gopath = filepath.Join(home, "go")
		}
	}

	govulncheckPath := filepath.Join(gopath, "bin", "govulncheck")
	if _, err := os.Stat(govulncheckPath); err == nil {
		return govulncheckPath
	}

	return "govulncheck" // Fall back to default
}
