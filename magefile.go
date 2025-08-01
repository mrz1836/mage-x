//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"

	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Re-export all namespaces from mage package
type (
	Build   = mage.Build
	Test    = mage.Test
	Lint    = mage.Lint
	Tools   = mage.Tools
	Deps    = mage.Deps
	Mod     = mage.Mod
	Docs    = mage.Docs
	Git     = mage.Git
	Release = mage.Release
	Metrics = mage.Metrics
	Version = mage.Version
	Install = mage.Install
	Audit   = mage.Audit
)

// Default target
var Default = BuildDefault //nolint:gochecknoglobals // Mage requires global Default target

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
		p, err := utils.ParsePlatform(platform)
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
			_ = os.Unsetenv("GOOS")   //nolint:errcheck // Cleanup, error not critical
			_ = os.Unsetenv("GOARCH") //nolint:errcheck // Cleanup, error not critical
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
	return g.Push("origin", "main")
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

// Help displays all available commands with beautiful MAGE-X formatting (clean output)
func Help() error {
	utils.Header("🎯 MAGE-X Commands")

	// Command categories with emoji icons and descriptions
	categories := map[string]struct {
		icon        string
		description string
		commands    []CommandInfo
	}{
		"core": {
			icon:        "🎯",
			description: "Essential Operations",
			commands: []CommandInfo{
				{"buildDefault", "Build the project for current platform", []string{"mage buildDefault", "mage build"}},
				{"testDefault", "Run the default test suite", []string{"mage testDefault", "mage test"}},
				{"lintDefault", "Run code quality checks", []string{"mage lintDefault", "mage lint"}},
				{"releaseDefault", "Create a new release", []string{"mage releaseDefault", "mage release"}},
			},
		},
		"build": {
			icon:        "🔨",
			description: "Build & Compilation",
			commands: []CommandInfo{
				{"buildAll", "Build for all configured platforms", []string{"mage buildAll"}},
				{"buildClean", "Clean build artifacts", []string{"mage buildClean"}},
				{"buildDocker", "Build Docker containers", []string{"mage buildDocker"}},
				{"buildGenerate", "Generate code before building", []string{"mage buildGenerate"}},
			},
		},
		"test": {
			icon:        "🧪",
			description: "Testing & Quality",
			commands: []CommandInfo{
				{"testFull", "Run full test suite with linting", []string{"mage testFull"}},
				{"testUnit", "Run unit tests only", []string{"mage testUnit"}},
				{"testRace", "Run tests with race detector", []string{"mage testRace"}},
				{"testCover", "Run tests with coverage", []string{"mage testCover"}},
				{"testBench", "Run benchmark tests", []string{"mage testBench"}},
				{"testBenchShort", "Run quick benchmark tests", []string{"mage testBenchShort"}},
				{"testFuzz", "Run fuzz tests", []string{"mage testFuzz"}},
				{"testFuzzShort", "Run quick fuzz tests (5s default)", []string{"mage testFuzzShort"}},
				{"testIntegration", "Run integration tests", []string{"mage testIntegration"}},
				{"testShort", "Run short tests only", []string{"mage testShort"}},
				{"testCoverRace", "Run tests with coverage and race detector", []string{"mage testCoverRace"}},
			},
		},
		"quality": {
			icon:        "✨",
			description: "Code Quality & Linting",
			commands: []CommandInfo{
				{"lintAll", "Run all linting checks", []string{"mage lintAll"}},
				{"lintFix", "Automatically fix linting issues", []string{"mage lintFix"}},
				{"lintVet", "Run go vet static analysis", []string{"mage lintVet"}},
				{"lintFumpt", "Run gofumpt code formatting", []string{"mage lintFumpt"}},
				{"lintVersion", "Show linter version", []string{"mage lintVersion"}},
			},
		},
		"deps": {
			icon:        "📦",
			description: "Dependency Management",
			commands: []CommandInfo{
				{"depsUpdate", "Update all dependencies", []string{"mage depsUpdate"}},
				{"depsTidy", "Clean up go.mod and go.sum", []string{"mage depsTidy"}},
				{"depsDownload", "Download all dependencies", []string{"mage depsDownload"}},
				{"depsOutdated", "Show outdated dependencies", []string{"mage depsOutdated"}},
				{"depsAudit", "Audit dependencies for vulnerabilities", []string{"mage depsAudit"}},
			},
		},
		"tools": {
			icon:        "🔧",
			description: "Development Tools",
			commands: []CommandInfo{
				{"toolsInstall", "Install development tools", []string{"mage toolsInstall"}},
				{"toolsUpdate", "Update development tools", []string{"mage toolsUpdate"}},
				{"toolsCheck", "Check if tools are available", []string{"mage toolsCheck"}},
				{"toolsVulnCheck", "Run vulnerability check", []string{"mage toolsVulnCheck"}},
				{"installTools", "Install all tools", []string{"mage installTools"}},
				{"installBinary", "Install project binary", []string{"mage installBinary"}},
				{"installStdlib", "Install Go standard library", []string{"mage installStdlib"}},
				{"uninstall", "Remove installed binary", []string{"mage uninstall"}},
			},
		},
		"modules": {
			icon:        "📋",
			description: "Go Module Operations",
			commands: []CommandInfo{
				{"modUpdate", "Update go.mod file", []string{"mage modUpdate"}},
				{"modTidy", "Tidy go.mod file", []string{"mage modTidy"}},
				{"modVerify", "Verify module checksums", []string{"mage modVerify"}},
				{"modDownload", "Download modules", []string{"mage modDownload"}},
			},
		},
		"docs": {
			icon:        "📚",
			description: "Documentation",
			commands: []CommandInfo{
				{"docsGenerate", "Generate documentation", []string{"mage docsGenerate"}},
				{"docsServe", "Serve documentation locally", []string{"mage docsServe"}},
				{"docsBuild", "Build static documentation", []string{"mage docsBuild"}},
				{"docsCheck", "Validate documentation", []string{"mage docsCheck"}},
			},
		},
		"git": {
			icon:        "🔀",
			description: "Git Operations",
			commands: []CommandInfo{
				{"gitStatus", "Show repository status", []string{"mage gitStatus"}},
				{"gitCommit", "Commit changes", []string{"mage gitCommit"}},
				{"gitTag", "Create and push tag", []string{"mage gitTag"}},
				{"gitPush", "Push changes to remote", []string{"mage gitPush"}},
			},
		},
		"version": {
			icon:        "🏷️",
			description: "Version Management",
			commands: []CommandInfo{
				{"versionShow", "Display version information", []string{"mage versionShow"}},
				{"versionBump", "Bump the version", []string{"mage versionBump"}},
				{"versionCheck", "Check version information", []string{"mage versionCheck"}},
			},
		},
		"metrics": {
			icon:        "📊",
			description: "Code Analysis & Metrics",
			commands: []CommandInfo{
				{"metricsLOC", "Analyze lines of code", []string{"mage metricsLOC"}},
				{"metricsCoverage", "Generate coverage reports", []string{"mage metricsCoverage"}},
				{"metricsComplexity", "Analyze code complexity", []string{"mage metricsComplexity"}},
			},
		},
		"audit": {
			icon:        "🛡️",
			description: "Security & Audit",
			commands: []CommandInfo{
				{"auditShow", "Display audit events", []string{"mage auditShow"}},
			},
		},
	}

	// Display each category with clean fmt.Printf (no logging prefixes)
	categoryOrder := []string{"core", "build", "test", "quality", "deps", "tools", "modules", "docs", "git", "version", "metrics", "audit"}

	for _, catKey := range categoryOrder {
		if category, exists := categories[catKey]; exists {
			fmt.Printf("\n%s %s:\n", category.icon, category.description)

			for _, cmd := range category.commands {
				fmt.Printf("  %-20s %s\n", cmd.name, cmd.description)
			}
		}
	}

	// Display usage information with clean fmt.Printf
	fmt.Printf("\n💡 Usage Tips:\n")
	fmt.Printf("  • Run any command: mage COMMAND\n")
	fmt.Printf("  • Beautiful list: mage help (this view)\n")
	fmt.Printf("  • Plain list: mage -l\n")
	fmt.Printf("  • Default target: mage (runs buildDefault)\n")
	fmt.Printf("  • Verbose output: VERBOSE=true mage COMMAND\n")
	fmt.Printf("  • Multiple commands: mage test lint build\n")

	fmt.Printf("\n🎯 Quick Start:\n")
	fmt.Printf("  mage build      # Build your project\n")
	fmt.Printf("  mage test       # Run tests\n")
	fmt.Printf("  mage lint       # Check code quality\n")
	fmt.Printf("  mage release    # Create a release\n")

	fmt.Printf("\n📖 More Help:\n")
	fmt.Printf("  • Interactive mode: mage interactive\n")
	fmt.Printf("  • Recipe system: mage recipes:list\n")
	fmt.Printf("  • Help system: mage help\n")

	return nil
}

// List displays all available commands with beautiful MAGE-X formatting
func List() error {
	utils.Header("🎯 MAGE-X Commands")

	// Command categories with emoji icons and descriptions
	categories := map[string]struct {
		icon        string
		description string
		commands    []CommandInfo
	}{
		"core": {
			icon:        "🎯",
			description: "Essential Operations",
			commands: []CommandInfo{
				{"buildDefault", "Build the project for current platform", []string{"mage buildDefault", "mage build"}},
				{"testDefault", "Run the default test suite", []string{"mage testDefault", "mage test"}},
				{"lintDefault", "Run code quality checks", []string{"mage lintDefault", "mage lint"}},
				{"releaseDefault", "Create a new release", []string{"mage releaseDefault", "mage release"}},
			},
		},
		"build": {
			icon:        "🔨",
			description: "Build & Compilation",
			commands: []CommandInfo{
				{"buildAll", "Build for all configured platforms", []string{"mage buildAll"}},
				{"buildClean", "Clean build artifacts", []string{"mage buildClean"}},
				{"buildDocker", "Build Docker containers", []string{"mage buildDocker"}},
				{"buildGenerate", "Generate code before building", []string{"mage buildGenerate"}},
			},
		},
		"test": {
			icon:        "🧪",
			description: "Testing & Quality",
			commands: []CommandInfo{
				{"testFull", "Run full test suite with linting", []string{"mage testFull"}},
				{"testUnit", "Run unit tests only", []string{"mage testUnit"}},
				{"testRace", "Run tests with race detector", []string{"mage testRace"}},
				{"testCover", "Run tests with coverage", []string{"mage testCover"}},
				{"testBench", "Run benchmark tests", []string{"mage testBench"}},
				{"testBenchShort", "Run quick benchmark tests", []string{"mage testBenchShort"}},
				{"testFuzz", "Run fuzz tests", []string{"mage testFuzz"}},
				{"testFuzzShort", "Run quick fuzz tests (5s default)", []string{"mage testFuzzShort"}},
				{"testIntegration", "Run integration tests", []string{"mage testIntegration"}},
				{"testShort", "Run short tests only", []string{"mage testShort"}},
				{"testCoverRace", "Run tests with coverage and race detector", []string{"mage testCoverRace"}},
			},
		},
		"quality": {
			icon:        "✨",
			description: "Code Quality & Linting",
			commands: []CommandInfo{
				{"lintAll", "Run all linting checks", []string{"mage lintAll"}},
				{"lintFix", "Automatically fix linting issues", []string{"mage lintFix"}},
				{"lintVet", "Run go vet static analysis", []string{"mage lintVet"}},
				{"lintFumpt", "Run gofumpt code formatting", []string{"mage lintFumpt"}},
				{"lintVersion", "Show linter version", []string{"mage lintVersion"}},
			},
		},
		"deps": {
			icon:        "📦",
			description: "Dependency Management",
			commands: []CommandInfo{
				{"depsUpdate", "Update all dependencies", []string{"mage depsUpdate"}},
				{"depsTidy", "Clean up go.mod and go.sum", []string{"mage depsTidy"}},
				{"depsDownload", "Download all dependencies", []string{"mage depsDownload"}},
				{"depsOutdated", "Show outdated dependencies", []string{"mage depsOutdated"}},
				{"depsAudit", "Audit dependencies for vulnerabilities", []string{"mage depsAudit"}},
			},
		},
		"tools": {
			icon:        "🔧",
			description: "Development Tools",
			commands: []CommandInfo{
				{"toolsInstall", "Install development tools", []string{"mage toolsInstall"}},
				{"toolsUpdate", "Update development tools", []string{"mage toolsUpdate"}},
				{"toolsCheck", "Check if tools are available", []string{"mage toolsCheck"}},
				{"toolsVulnCheck", "Run vulnerability check", []string{"mage toolsVulnCheck"}},
				{"installTools", "Install all tools", []string{"mage installTools"}},
				{"installBinary", "Install project binary", []string{"mage installBinary"}},
				{"installStdlib", "Install Go standard library", []string{"mage installStdlib"}},
				{"uninstall", "Remove installed binary", []string{"mage uninstall"}},
			},
		},
		"modules": {
			icon:        "📋",
			description: "Go Module Operations",
			commands: []CommandInfo{
				{"modUpdate", "Update go.mod file", []string{"mage modUpdate"}},
				{"modTidy", "Tidy go.mod file", []string{"mage modTidy"}},
				{"modVerify", "Verify module checksums", []string{"mage modVerify"}},
				{"modDownload", "Download modules", []string{"mage modDownload"}},
			},
		},
		"docs": {
			icon:        "📚",
			description: "Documentation",
			commands: []CommandInfo{
				{"docsGenerate", "Generate documentation", []string{"mage docsGenerate"}},
				{"docsServe", "Serve documentation locally", []string{"mage docsServe"}},
				{"docsBuild", "Build static documentation", []string{"mage docsBuild"}},
				{"docsCheck", "Validate documentation", []string{"mage docsCheck"}},
			},
		},
		"git": {
			icon:        "🔀",
			description: "Git Operations",
			commands: []CommandInfo{
				{"gitStatus", "Show repository status", []string{"mage gitStatus"}},
				{"gitCommit", "Commit changes", []string{"mage gitCommit"}},
				{"gitTag", "Create and push tag", []string{"mage gitTag"}},
				{"gitPush", "Push changes to remote", []string{"mage gitPush"}},
			},
		},
		"version": {
			icon:        "🏷️",
			description: "Version Management",
			commands: []CommandInfo{
				{"versionShow", "Display version information", []string{"mage versionShow"}},
				{"versionBump", "Bump the version", []string{"mage versionBump"}},
				{"versionCheck", "Check version information", []string{"mage versionCheck"}},
			},
		},
		"metrics": {
			icon:        "📊",
			description: "Code Analysis & Metrics",
			commands: []CommandInfo{
				{"metricsLOC", "Analyze lines of code", []string{"mage metricsLOC"}},
				{"metricsCoverage", "Generate coverage reports", []string{"mage metricsCoverage"}},
				{"metricsComplexity", "Analyze code complexity", []string{"mage metricsComplexity"}},
			},
		},
		"audit": {
			icon:        "🛡️",
			description: "Security & Audit",
			commands: []CommandInfo{
				{"auditShow", "Display audit events", []string{"mage auditShow"}},
			},
		},
	}

	// Display each category
	categoryOrder := []string{"core", "build", "test", "quality", "deps", "tools", "modules", "docs", "git", "version", "metrics", "audit"}

	for _, catKey := range categoryOrder {
		if category, exists := categories[catKey]; exists {
			utils.Info("\n%s %s:", category.icon, category.description)

			for _, cmd := range category.commands {
				utils.Info("  %-20s %s", cmd.name, cmd.description)
			}
		}
	}

	// Display usage information
	utils.Info("\n💡 Usage Tips:")
	utils.Info("  • Run any command: mage COMMAND")
	utils.Info("  • Beautiful list: mage help (this view)")
	utils.Info("  • Plain list: mage -l")
	utils.Info("  • Default target: mage (runs buildDefault)")
	utils.Info("  • Verbose output: VERBOSE=true mage COMMAND")
	utils.Info("  • Multiple commands: mage test lint build")

	utils.Info("\n🎯 Quick Start:")
	utils.Info("  mage build      # Build your project")
	utils.Info("  mage test       # Run tests")
	utils.Info("  mage lint       # Check code quality")
	utils.Info("  mage release    # Create a release")

	utils.Info("\n📖 More Help:")
	utils.Info("  • Interactive mode: mage interactive")
	utils.Info("  • Recipe system: mage recipes:list")
	utils.Info("  • Help system: mage help")

	return nil
}

// CommandInfo represents command information for the List function
type CommandInfo struct {
	name        string
	description string
	examples    []string
}
