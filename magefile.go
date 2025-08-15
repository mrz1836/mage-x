//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Re-export all namespaces from mage package
type (
	Audit     = mage.Audit
	Build     = mage.Build
	Configure = mage.Configure
	Deps      = mage.Deps
	Docs      = mage.Docs
	Format    = mage.Format
	Generate  = mage.Generate
	Git       = mage.Git
	Help      = mage.Help
	Init      = mage.Init
	Install   = mage.Install
	Lint      = mage.Lint
	Metrics   = mage.Metrics
	Mod       = mage.Mod
	Recipes   = mage.Recipes
	Release   = mage.Release
	Test      = mage.Test
	Tools     = mage.Tools
	Update    = mage.Update
	Version   = mage.Version
	Vet       = mage.Vet
)

// Aliases provides short command names for common operations
//
//nolint:gochecknoglobals // Required by mage for command aliases
var Aliases = map[string]interface{}{
	"build":   BuildDefault,
	"docs":    DocsDefault,
	"help":    HelpDefault,
	"lint":    LintDefault,
	"loc":     MetricsLOC,
	"release": ReleaseDefault,
	"test":    TestDefault,
}

// Default target
//
//nolint:gochecknoglobals // Required by mage for default target
var Default = BuildDefault

// BuildDefault is the default build target
func BuildDefault() error {
	var b Build
	return b.Default()
}

// TestDefault runs the default test suite (unit tests only)
func TestDefault() error {
	var t Test
	return t.Default()
}

// TestFull runs the full test suite with linting
func TestFull() error {
	var t Test
	return t.Full()
}

// LintDefault runs the default linter
func LintDefault() error {
	var l Lint
	return l.Default()
}

// TestUnit runs unit tests without linting
func TestUnit() error {
	var t Test
	return t.Unit()
}

// TestRace runs tests with race detector
func TestRace() error {
	var t Test
	return t.Race()
}

// InstallStdlib installs Go standard library (for cross-compilation)
func InstallStdlib() error {
	utils.Header("Installing Go Standard Library")

	cfg, err := mage.GetConfig()
	if err != nil {
		return err
	}

	// Install standard library for all configured platforms
	for _, platform := range cfg.Build.Platforms {
		var p utils.Platform
		p, err = utils.ParsePlatform(platform)
		if err != nil {
			return err
		}

		utils.Info("Installing stdlib for %s/%s...", p.OS, p.Arch)

		// Set environment for the target platform
		if err := os.Setenv("GOOS", p.OS); err != nil {
			return fmt.Errorf("failed to set GOOS: %w", err)
		}
		if err := os.Setenv("GOARCH", p.Arch); err != nil {
			return fmt.Errorf("failed to set GOARCH: %w", err)
		}
		defer func() {
			if unsetErr := os.Unsetenv("GOOS"); unsetErr != nil {
				utils.Error("Failed to unset GOOS: %v", unsetErr)
			}
			if unsetErr := os.Unsetenv("GOARCH"); unsetErr != nil {
				utils.Error("Failed to unset GOARCH: %v", unsetErr)
			}
		}()

		// Install the standard library
		if err := mage.GetRunner().RunCmd("go", "install", "-a", "std"); err != nil {
			return fmt.Errorf("failed to install stdlib for %s: %w", platform, err)
		}
	}

	utils.Success("Standard library installed for all platforms")
	return nil
}

// Uninstall removes the installed binary
func Uninstall() error {
	var i Install
	return i.Uninstall()
}

// ReleaseDefault creates a new release (default)
func ReleaseDefault() error {
	var r Release
	return r.Default()
}

// DepsUpdate updates all dependencies (equivalent to "make update")
func DepsUpdate() error {
	var d Deps
	return d.Update()
}

// DepsTidy cleans up go.mod and go.sum
func DepsTidy() error {
	var d Deps
	return d.Tidy()
}

// DepsDownload downloads all dependencies
func DepsDownload() error {
	var d Deps
	return d.Download()
}

// DepsOutdated shows outdated dependencies
func DepsOutdated() error {
	var d Deps
	return d.Outdated()
}

// DepsAudit audits dependencies for vulnerabilities
func DepsAudit() error {
	var d Deps
	return d.Audit()
}

// ToolsUpdate updates all development tools
func ToolsUpdate() error {
	var t Tools
	return t.Update()
}

// ToolsInstall installs all required development tools
func ToolsInstall() error {
	var t Tools
	return t.Install()
}

// ToolsCheck checks if all required tools are available
func ToolsCheck() error {
	var t Tools
	return t.Check()
}

// ModUpdate updates go.mod file
func ModUpdate() error {
	var m Mod
	return m.Update()
}

// ModTidy tidies the go.mod file
func ModTidy() error {
	var m Mod
	return m.Tidy()
}

// ModVerify verifies module checksums
func ModVerify() error {
	var m Mod
	return m.Verify()
}

// ModDownload downloads modules
func ModDownload() error {
	var m Mod
	return m.Download()
}

// DocsGenerate generates documentation
func DocsGenerate() error {
	var d Docs
	return d.Generate()
}

// DocsServe serves documentation locally
func DocsServe() error {
	var d Docs
	return d.Serve()
}

// DocsBuild builds static documentation
func DocsBuild() error {
	var d Docs
	return d.Build()
}

// DocsDefault generates and serves documentation (generate + serve in one command)
func DocsDefault() error {
	utils.Header("📚 MAGE-X Documentation Generator & Server")

	// First generate documentation
	utils.Info("Step 1: Generating documentation...")
	if err := DocsGenerate(); err != nil {
		return fmt.Errorf("failed to generate documentation: %w", err)
	}

	utils.Success("Documentation generated successfully!")
	utils.Info("Step 2: Starting documentation server...")

	// Then serve it
	return DocsServe()
}

// DocsCheck validates documentation
func DocsCheck() error {
	var d Docs
	return d.Check()
}

// GitStatus shows git repository status
func GitStatus() error {
	var g Git
	return g.Status()
}

// GitCommit commits changes
func GitCommit() error {
	var g Git
	return g.Commit()
}

// GitTag creates and pushes a new tag
func GitTag() error {
	var g Git
	return g.Tag()
}

// GitPush pushes changes to remote
func GitPush() error {
	var g Git

	// Get current branch name
	current, err := mage.GetRunner().RunCmdOutput("git", "branch", "--show-current")
	if err != nil {
		// Fallback to master if current branch detection fails
		return g.Push("origin", "master")
	}

	branchName := strings.TrimSpace(current)
	if branchName == "" {
		// Fallback to master if branch name is empty
		branchName = "master"
	}

	return g.Push("origin", branchName)
}

// VersionShow displays current version information
func VersionShow() error {
	var v Version
	return v.Show()
}

// VersionBump bumps the version
func VersionBump() error {
	var v Version
	return v.Bump()
}

// VersionCheck checks version information
func VersionCheck() error {
	var v Version
	return v.Check()
}

// BuildDocker builds Docker containers
func BuildDocker() error {
	var b Build
	return b.Docker()
}

// BuildClean cleans build artifacts
func BuildClean() error {
	var b Build
	return b.Clean()
}

// BuildGenerate generates code before building
func BuildGenerate() error {
	var b Build
	return b.Generate()
}

// TestCover runs tests with coverage analysis
func TestCover() error {
	var t Test
	return t.Cover()
}

// TestBench runs benchmark tests
func TestBench() error {
	var t Test
	return t.Bench()
}

// TestBenchShort runs benchmark tests with shorter duration for quick feedback
func TestBenchShort() error {
	var t Test
	return t.BenchShort()
}

// TestFuzz runs fuzz tests
func TestFuzz() error {
	var t Test
	return t.Fuzz()
}

// TestFuzzShort runs fuzz tests with shorter duration for quick feedback
func TestFuzzShort() error {
	var t Test
	return t.FuzzShort()
}

// TestIntegration runs integration tests
func TestIntegration() error {
	var t Test
	return t.Integration()
}

// LintAll runs all linting checks
func LintAll() error {
	var l Lint
	return l.All()
}

// LintFix automatically fixes linting issues
func LintFix() error {
	var l Lint
	return l.Fix()
}

// MetricsLOC analyzes lines of code
func MetricsLOC() error {
	var m Metrics
	return m.LOC()
}

// MetricsCoverage generates coverage reports
func MetricsCoverage() error {
	var m Metrics
	return m.Coverage()
}

// MetricsComplexity analyzes code complexity
func MetricsComplexity() error {
	var m Metrics
	return m.Complexity()
}

// InstallTools installs all development tools
func InstallTools() error {
	var i Install
	return i.Tools()
}

// InstallBinary installs the project binary
func InstallBinary() error {
	var i Install
	return i.Binary()
}

// TestShort runs short tests (excludes integration tests)
func TestShort() error {
	var t Test
	return t.Short()
}

// TestCoverRace runs tests with both coverage and race detector
func TestCoverRace() error {
	var t Test
	return t.CoverRace()
}

// LintVersion shows the linter version information
func LintVersion() error {
	var l Lint
	return l.Version()
}

// BuildAll builds for all configured platforms
func BuildAll() error {
	var b Build
	return b.All()
}

// ToolsVulnCheck runs vulnerability check using govulncheck
func ToolsVulnCheck() error {
	var t Tools
	return t.VulnCheck()
}

// LintVet runs go vet static analysis
func LintVet() error {
	var l Lint
	return l.Vet()
}

// LintFumpt runs gofumpt code formatting
func LintFumpt() error {
	var l Lint
	return l.Fumpt()
}

// AuditShow displays audit events with optional filtering
func AuditShow() error {
	var a Audit
	return a.Show()
}

// ConfigureInit initializes a new mage configuration
func ConfigureInit() error {
	var c Configure
	return c.Init()
}

// ConfigureShow displays the current configuration
func ConfigureShow() error {
	var c Configure
	return c.Show()
}

// ConfigureUpdate updates configuration values interactively
func ConfigureUpdate() error {
	var c Configure
	return c.Update()
}

// FormatAll formats all Go source files
func FormatAll() error {
	var f Format
	return f.All()
}

// FormatCheck checks if Go source files are properly formatted
func FormatCheck() error {
	var f Format
	return f.Check()
}

// GenerateDefault runs go generate
func GenerateDefault() error {
	var g Generate
	return g.Default()
}

// GenerateClean removes generated files
func GenerateClean() error {
	var g Generate
	return g.Clean()
}

// InitProject initializes a new project
func InitProject() error {
	var i Init
	return i.Project()
}

// InitCLI initializes a CLI project
func InitCLI() error {
	var i Init
	return i.CLI()
}

// InitLibrary initializes a library project
func InitLibrary() error {
	var i Init
	return i.Library()
}

// RecipesList lists available recipes
func RecipesList() error {
	var r Recipes
	return r.List()
}

// RecipesRun runs a specific recipe
func RecipesRun() error {
	var r Recipes
	return r.Run()
}

// UpdateCheck checks for updates
func UpdateCheck() error {
	var u Update
	return u.Check()
}

// UpdateInstall installs the latest update
func UpdateInstall() error {
	var u Update
	return u.Install()
}

// VetDefault runs go vet
func VetDefault() error {
	var v Vet
	return v.Default()
}

// VetAll runs go vet with all checks
func VetAll() error {
	var v Vet
	return v.All()
}

// HelpDefault displays all available commands with beautiful MAGE-X formatting (clean output)
func HelpDefault() error {
	var h Help
	return h.Default()
}

// HelpCommands lists all available commands with descriptions
func HelpCommands() error {
	var h Help
	return h.Commands()
}

// HelpExamples shows usage examples
func HelpExamples() error {
	var h Help
	return h.Examples()
}

// HelpGettingStarted shows the getting started guide
func HelpGettingStarted() error {
	var h Help
	return h.GettingStarted()
}

// HelpCompletions generates shell completions
func HelpCompletions() error {
	var h Help
	return h.Completions()
}

// HelpCommand shows help for a specific command
func HelpCommand() error {
	var h Help
	return h.Command()
}

// ShowHelp displays all available commands with beautiful MAGE-X formatting (clean output)
// This function is preserved for backward compatibility with existing magefiles
// The actual help functionality has been unified in the magex binary
func ShowHelp() {
	utils.Header("🎯 MAGE-X Commands")
	fmt.Printf("\n📚 For comprehensive help with all commands, use:\n")
	fmt.Printf("  magex -h                 # Full help with categories and examples\n")
	fmt.Printf("  magex -h <command>       # Detailed help for specific command\n")
	fmt.Printf("  magex -search <term>     # Search for commands\n")
	fmt.Printf("  magex -l                 # Simple list of all commands\n")

	fmt.Printf("\n🎯 Quick Commands:\n")
	fmt.Printf("  mage build               # Build your project\n")
	fmt.Printf("  mage test                # Run tests\n")
	fmt.Printf("  mage lint                # Check code quality\n")
	fmt.Printf("  mage release             # Create a release\n")

	fmt.Printf("\n📖 Note: This is the legacy help view. For the complete unified help system\n")
	fmt.Printf("   with 215+ commands, search, and detailed documentation, use 'magex -h'\n")
}

// List displays all available commands with beautiful MAGE-X formatting
// This function is preserved for backward compatibility with existing magefiles
// The actual help functionality has been unified in the magex binary
func List() {
	utils.Header("🎯 MAGE-X Commands")
	fmt.Printf("\n📚 For comprehensive command listing with all 215+ commands, use:\n")
	fmt.Printf("  magex -l                 # Simple list of all commands\n")
	fmt.Printf("  magex -n                 # Commands organized by namespace\n")
	fmt.Printf("  magex -h                 # Full help with categories and descriptions\n")
	fmt.Printf("  magex -search <term>     # Search for specific commands\n")

	fmt.Printf("\n🎯 Available in This Magefile:\n")
	fmt.Printf("  • All MAGE-X commands through mage binary\n")
	fmt.Printf("  • Custom functions defined in this file\n")
	fmt.Printf("  • Aliases: build, test, lint, release, docs, help, loc\n")

	fmt.Printf("\n💡 Quick Commands:\n")
	fmt.Printf("  mage build               # Build your project\n")
	fmt.Printf("  mage test                # Run tests\n")
	fmt.Printf("  mage lint                # Check code quality\n")
	fmt.Printf("  mage release             # Create a release\n")

	fmt.Printf("\n📖 Note: This is the legacy list view. For the complete unified system\n")
	fmt.Printf("   with all commands, categories, and search capabilities, use 'magex -l'\n")
}

// CommandInfo represents command information for the List function
// Fields commented to avoid unused variable warnings - reserved for future use
type CommandInfo struct {
	// name        string
	// description string
	// examples    []string
}
