//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Namespace types - these enable namespace:method syntax in mage
// Each type must be defined as mg.Namespace for mage to recognize it
type (
	Build mg.Namespace
	Test  mg.Namespace
	Lint  mg.Namespace
	Deps  mg.Namespace
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
	var b mage.Build
	return b.Default()
}

// TestDefault runs the default test suite (unit tests only)
func TestDefault() error {
	var t mage.Test
	return t.Default()
}

// TestFull runs the full test suite with linting
func TestFull() error {
	var t mage.Test
	return t.Full()
}

// LintDefault runs the default linter
func LintDefault() error {
	var l mage.Lint
	return l.Default()
}

// TestUnit runs unit tests without linting
func TestUnit() error {
	var t mage.Test
	return t.Unit()
}

// TestRace runs tests with race detector
func TestRace() error {
	var t mage.Test
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
	var i mage.Install
	return i.Uninstall()
}

// ReleaseDefault creates a new release (default)
func ReleaseDefault() error {
	var r mage.Release
	return r.Default()
}

// DepsUpdate updates all dependencies (equivalent to "make update")
func DepsUpdate() error {
	var d mage.Deps
	return d.Update()
}

// DepsTidy cleans up go.mod and go.sum
func DepsTidy() error {
	var d mage.Deps
	return d.Tidy()
}

// DepsDownload downloads all dependencies
func DepsDownload() error {
	var d mage.Deps
	return d.Download()
}

// DepsOutdated shows outdated dependencies
func DepsOutdated() error {
	var d mage.Deps
	return d.Outdated()
}

// DepsAudit audits dependencies for vulnerabilities
func DepsAudit() error {
	var d mage.Deps
	return d.Audit()
}

// ToolsUpdate updates all development tools
func ToolsUpdate() error {
	var t mage.Tools
	return t.Update()
}

// ToolsInstall installs all required development tools
func ToolsInstall() error {
	var t mage.Tools
	return t.Install()
}

// ToolsCheck checks if all required tools are available
func ToolsCheck() error {
	var t mage.Tools
	return t.Check()
}

// ModUpdate updates go.mod file
func ModUpdate() error {
	var m mage.Mod
	return m.Update()
}

// ModTidy tidies the go.mod file
func ModTidy() error {
	var m mage.Mod
	return m.Tidy()
}

// ModVerify verifies module checksums
func ModVerify() error {
	var m mage.Mod
	return m.Verify()
}

// ModDownload downloads modules
func ModDownload() error {
	var m mage.Mod
	return m.Download()
}

// DocsGenerate generates documentation
func DocsGenerate() error {
	var d mage.Docs
	return d.Generate()
}

// DocsServe serves documentation locally
func DocsServe() error {
	var d mage.Docs
	return d.Serve()
}

// DocsBuild builds static documentation
func DocsBuild() error {
	var d mage.Docs
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
	var d mage.Docs
	return d.Check()
}

// GitStatus shows git repository status
func GitStatus() error {
	var g mage.Git
	return g.Status()
}

// GitCommit commits changes
func GitCommit() error {
	var g mage.Git
	return g.Commit()
}

// GitTag creates and pushes a new tag
func GitTag() error {
	var g mage.Git
	return g.Tag()
}

// GitPush pushes changes to remote
func GitPush() error {
	// Get current branch name
	current, err := mage.GetRunner().RunCmdOutput("git", "branch", "--show-current")
	if err != nil {
		// Fallback to master if current branch detection fails
		var impl mage.Git
		return impl.Push("origin", "master")
	}

	branchName := strings.TrimSpace(current)
	if branchName == "" {
		// Fallback to master if branch name is empty
		branchName = "master"
	}

	var impl mage.Git
	return impl.Push("origin", branchName)
}

// VersionShow displays current version information
func VersionShow() error {
	var v mage.Version
	return v.Show()
}

// VersionBump bumps the version
func VersionBump() error {
	var v mage.Version
	return v.Bump()
}

// VersionCheck checks version information
func VersionCheck() error {
	var v mage.Version
	return v.Check()
}

// BuildDocker builds Docker containers
func BuildDocker() error {
	var b mage.Build
	return b.Docker()
}

// BuildClean cleans build artifacts
func BuildClean() error {
	var b mage.Build
	return b.Clean()
}

// BuildGenerate generates code before building
func BuildGenerate() error {
	var b mage.Build
	return b.Generate()
}

// TestCover runs tests with coverage analysis
func TestCover() error {
	var t mage.Test
	return t.Cover()
}

// TestBench runs benchmark tests
func TestBench() error {
	var t mage.Test
	return t.Bench()
}

// TestBenchShort runs benchmark tests with shorter duration for quick feedback
func TestBenchShort() error {
	var t mage.Test
	return t.BenchShort()
}

// TestFuzz runs fuzz tests
func TestFuzz() error {
	var t mage.Test
	return t.Fuzz()
}

// TestFuzzShort runs fuzz tests with shorter duration for quick feedback
func TestFuzzShort() error {
	var t mage.Test
	return t.FuzzShort()
}

// TestIntegration runs integration tests
func TestIntegration() error {
	var t mage.Test
	return t.Integration()
}

// LintAll runs all linting checks
func LintAll() error {
	var l mage.Lint
	return l.All()
}

// LintFix automatically fixes linting issues
func LintFix() error {
	var l mage.Lint
	return l.Fix()
}

// MetricsLOC analyzes lines of code
func MetricsLOC() error {
	var m mage.Metrics
	return m.LOC()
}

// MetricsCoverage generates coverage reports
func MetricsCoverage() error {
	var m mage.Metrics
	return m.Coverage()
}

// MetricsComplexity analyzes code complexity
func MetricsComplexity() error {
	var m mage.Metrics
	return m.Complexity()
}

// InstallTools installs all development tools
func InstallTools() error {
	var i mage.Install
	return i.Tools()
}

// InstallBinary installs the project binary
func InstallBinary() error {
	var i mage.Install
	return i.Binary()
}

// TestShort runs short tests (excludes integration tests)
func TestShort() error {
	var t mage.Test
	return t.Short()
}

// TestCoverRace runs tests with both coverage and race detector
func TestCoverRace() error {
	var t mage.Test
	return t.CoverRace()
}

// LintVersion shows the linter version information
func LintVersion() error {
	var l mage.Lint
	return l.Version()
}

// BuildAll builds for all configured platforms
func BuildAll() error {
	var b mage.Build
	return b.All()
}

// ToolsVulnCheck runs vulnerability check using govulncheck
func ToolsVulnCheck() error {
	var t mage.Tools
	return t.VulnCheck()
}

// LintVet runs go vet static analysis
func LintVet() error {
	var l mage.Lint
	return l.Vet()
}

// LintFumpt runs gofumpt code formatting
func LintFumpt() error {
	var l mage.Lint
	return l.Fumpt()
}

// AuditShow displays audit events with optional filtering
func AuditShow() error {
	var a mage.Audit
	return a.Show()
}

// ConfigureInit initializes a new mage configuration
func ConfigureInit() error {
	var c mage.Configure
	return c.Init()
}

// ConfigureShow displays the current configuration
func ConfigureShow() error {
	var c mage.Configure
	return c.Show()
}

// ConfigureUpdate updates configuration values interactively
func ConfigureUpdate() error {
	var c mage.Configure
	return c.Update()
}

// FormatAll formats all Go source files
func FormatAll() error {
	var f mage.Format
	return f.All()
}

// FormatCheck checks if Go source files are properly formatted
func FormatCheck() error {
	var f mage.Format
	return f.Check()
}

// GenerateDefault runs go generate
func GenerateDefault() error {
	var g mage.Generate
	return g.Default()
}

// GenerateClean removes generated files
func GenerateClean() error {
	var g mage.Generate
	return g.Clean()
}

// InitProject initializes a new project
func InitProject() error {
	var i mage.Init
	return i.Project()
}

// InitCLI initializes a CLI project
func InitCLI() error {
	var i mage.Init
	return i.CLI()
}

// InitLibrary initializes a library project
func InitLibrary() error {
	var i mage.Init
	return i.Library()
}

// RecipesList lists available recipes
func RecipesList() error {
	var r mage.Recipes
	return r.List()
}

// RecipesRun runs a specific recipe
func RecipesRun() error {
	var r mage.Recipes
	return r.Run()
}

// UpdateCheck checks for updates
func UpdateCheck() error {
	var u mage.Update
	return u.Check()
}

// UpdateInstall installs the latest update
func UpdateInstall() error {
	var u mage.Update
	return u.Install()
}

// VetDefault runs go vet
func VetDefault() error {
	var v mage.Vet
	return v.Default()
}

// VetAll runs go vet with all checks
func VetAll() error {
	var v mage.Vet
	return v.All()
}

// HelpDefault displays all available commands with beautiful MAGE-X formatting (clean output)
func HelpDefault() error {
	var h mage.Help
	return h.Default()
}

// HelpCommands lists all available commands with descriptions
func HelpCommands() error {
	var h mage.Help
	return h.Commands()
}

// HelpExamples shows usage examples
func HelpExamples() error {
	var h mage.Help
	return h.Examples()
}

// HelpGettingStarted shows the getting started guide
func HelpGettingStarted() error {
	var h mage.Help
	return h.GettingStarted()
}

// HelpCompletions generates shell completions
func HelpCompletions() error {
	var h mage.Help
	return h.Completions()
}

// HelpCommand shows help for a specific command
func HelpCommand() error {
	var h mage.Help
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

// ============================================================================
// NAMESPACE FORWARDING METHODS
// These methods enable namespace:method syntax in mage (e.g., build:default)
// ============================================================================

// Build namespace methods
func (b Build) Default() error {
	var impl mage.Build
	return impl.Default()
}

func (b Build) All() error {
	var impl mage.Build
	return impl.All()
}

// Platform method requires parameters, not suitable for namespace syntax
func (b Build) Linux() error {
	var impl mage.Build
	return impl.Linux()
}

func (b Build) Darwin() error {
	var impl mage.Build
	return impl.Darwin()
}

func (b Build) Windows() error {
	var impl mage.Build
	return impl.Windows()
}

func (b Build) Docker() error {
	var impl mage.Build
	return impl.Docker()
}

func (b Build) Clean() error {
	var impl mage.Build
	return impl.Clean()
}

func (b Build) Generate() error {
	var impl mage.Build
	return impl.Generate()
}

// Test namespace methods
func (t Test) Default() error {
	var impl mage.Test
	return impl.Default()
}

func (t Test) Full() error {
	var impl mage.Test
	return impl.Full()
}

func (t Test) Unit() error {
	var impl mage.Test
	return impl.Unit()
}

func (t Test) Short() error {
	var impl mage.Test
	return impl.Short()
}

func (t Test) Race() error {
	var impl mage.Test
	return impl.Race()
}

func (t Test) Cover() error {
	var impl mage.Test
	return impl.Cover()
}

func (t Test) CoverRace() error {
	var impl mage.Test
	return impl.CoverRace()
}

func (t Test) CoverReport() error {
	var impl mage.Test
	return impl.CoverReport()
}

func (t Test) CoverHTML() error {
	var impl mage.Test
	return impl.CoverHTML()
}

/* Temporarily commented out remaining methods to test basic namespace functionality
func (t Test) Fuzz() error { return t.Test.Fuzz() }
func (t Test) FuzzShort() error { return t.Test.FuzzShort() }
func (t Test) Bench() error { return t.Test.Bench() }
func (t Test) BenchShort() error { return t.Test.BenchShort() }
func (t Test) Integration() error { return t.Test.Integration() }
func (t Test) CI() error { return t.Test.CI() }
func (t Test) Parallel() error { return t.Test.Parallel() }
func (t Test) NoLint() error { return t.Test.NoLint() }
func (t Test) CINoRace() error { return t.Test.CINoRace() }
func (t Test) Run() error { return t.Test.Run() }
func (t Test) Coverage() error { return t.Test.Coverage() }

// Lint namespace methods
func (l Lint) Default() error { return l.Lint.Default() }
func (l Lint) Fix() error { return l.Lint.Fix() }
func (l Lint) Fmt() error { return l.Lint.Fmt() }
func (l Lint) Fumpt() error { return l.Lint.Fumpt() }
func (l Lint) Vet() error { return l.Lint.Vet() }
func (l Lint) VetParallel() error { return l.Lint.VetParallel() }
func (l Lint) Version() error { return l.Lint.Version() }
func (l Lint) All() error { return l.Lint.All() }
func (l Lint) Go() error { return l.Lint.Go() }
func (l Lint) Docker() error { return l.Lint.Docker() }
func (l Lint) YAML() error { return l.Lint.YAML() }
func (l Lint) Yaml() error { return l.Lint.Yaml() }
func (l Lint) Markdown() error { return l.Lint.Markdown() }
func (l Lint) Shell() error { return l.Lint.Shell() }
func (l Lint) JSON() error { return l.Lint.JSON() }
func (l Lint) SQL() error { return l.Lint.SQL() }
func (l Lint) Config() error { return l.Lint.Config() }
func (l Lint) CI() error { return l.Lint.CI() }
func (l Lint) Fast() error { return l.Lint.Fast() }

// Deps namespace methods
func (d Deps) Default() error { return d.Deps.Default() }
func (d Deps) Download() error { return d.Deps.Download() }
func (d Deps) Tidy() error { return d.Deps.Tidy() }
func (d Deps) Update() error { return d.Deps.Update() }
func (d Deps) Clean() error { return d.Deps.Clean() }
func (d Deps) Graph() error { return d.Deps.Graph() }
// Why method requires parameters, not suitable for namespace syntax
func (d Deps) Verify() error { return d.Deps.Verify() }
func (d Deps) VulnCheck() error { return d.Deps.VulnCheck() }
func (d Deps) List() error { return d.Deps.List() }
func (d Deps) Outdated() error { return d.Deps.Outdated() }
func (d Deps) Vendor() error { return d.Deps.Vendor() }
// Init method requires parameters, not suitable for namespace syntax
func (d Deps) Audit() error { return d.Deps.Audit() }
func (d Deps) Licenses() error { return d.Deps.Licenses() }
func (d Deps) Check() error { return d.Deps.Check() }

// Tools namespace methods
func (t Tools) Default() error { return t.Tools.Default() }
func (t Tools) Install() error { return t.Tools.Install() }
func (t Tools) Update() error { return t.Tools.Update() }
func (t Tools) Verify() error { return t.Tools.Verify() }
func (t Tools) List() error { return t.Tools.List() }
func (t Tools) VulnCheck() error { return t.Tools.VulnCheck() }
func (t Tools) Check() error { return t.Tools.Check() }
func (t Tools) Clean() error { return t.Tools.Clean() }

// Mod namespace methods
func (m Mod) Download() error { return m.Mod.Download() }
func (m Mod) Tidy() error { return m.Mod.Tidy() }
func (m Mod) Update() error { return m.Mod.Update() }
func (m Mod) Clean() error { return m.Mod.Clean() }
func (m Mod) Graph() error { return m.Mod.Graph() }
func (m Mod) Why() error { return m.Mod.Why() }
func (m Mod) Vendor() error { return m.Mod.Vendor() }
func (m Mod) Init() error { return m.Mod.Init() }
func (m Mod) Verify() error { return m.Mod.Verify() }
func (m Mod) Edit() error { return m.Mod.Edit() }
func (m Mod) Get() error { return m.Mod.Get() }
func (m Mod) List() error { return m.Mod.List() }

// Docs namespace methods
func (d Docs) Default() error { return d.Docs.Default() }
func (d Docs) Generate() error { return d.Docs.Generate() }
func (d Docs) Serve() error { return d.Docs.Serve() }
func (d Docs) ServeDefault() error { return d.Docs.ServeDefault() }
func (d Docs) ServePkgsite() error { return d.Docs.ServePkgsite() }
func (d Docs) ServeGodoc() error { return d.Docs.ServeGodoc() }
func (d Docs) ServeStdlib() error { return d.Docs.ServeStdlib() }
func (d Docs) ServeProject() error { return d.Docs.ServeProject() }
func (d Docs) ServeBoth() error { return d.Docs.ServeBoth() }
func (d Docs) Check() error { return d.Docs.Check() }
func (d Docs) Examples() error { return d.Docs.Examples() }
func (d Docs) Build() error { return d.Docs.Build() }
func (d Docs) Lint() error { return d.Docs.Lint() }
func (d Docs) Spell() error { return d.Docs.Spell() }
func (d Docs) Links() error { return d.Docs.Links() }
func (d Docs) API() error { return d.Docs.API() }
func (d Docs) Markdown() error { return d.Docs.Markdown() }
func (d Docs) Readme() error { return d.Docs.Readme() }
// Changelog method requires parameters, not suitable for namespace syntax
func (d Docs) GoDocs() error { return d.Docs.GoDocs() }

// Git namespace methods
func (g Git) Diff() error { return g.Git.Diff() }
func (g Git) Tag() error { return g.Git.Tag() }
func (g Git) TagRemove() error { return g.Git.TagRemove() }
func (g Git) TagUpdate() error { return g.Git.TagUpdate() }
func (g Git) Status() error { return g.Git.Status() }
func (g Git) Log() error { return g.Git.Log() }
func (g Git) Branch() error { return g.Git.Branch() }
func (g Git) Pull() error { return g.Git.Pull() }
func (g Git) Commit() error { return g.Git.Commit() }
func (g Git) Init() error { return g.Git.Init() }
func (g Git) Add() error { return g.Git.Add() }
func (g Git) Clone() error { return g.Git.Clone() }
// Push method requires parameters, not suitable for namespace syntax
// PullWithRemote method requires parameters, not suitable for namespace syntax
// TagWithMessage method requires parameters, not suitable for namespace syntax
// BranchWithName method requires parameters, not suitable for namespace syntax
// CloneRepo method requires parameters, not suitable for namespace syntax
// LogWithCount method requires parameters, not suitable for namespace syntax

// Version namespace methods
func (v Version) Show() error { return v.Version.Show() }
func (v Version) Check() error { return v.Version.Check() }
func (v Version) Update() error { return v.Version.Update() }
func (v Version) Bump() error { return v.Version.Bump() }
func (v Version) Changelog() error { return v.Version.Changelog() }
func (v Version) Tag() error { return v.Version.Tag() }
// Next method requires parameters and returns multiple values, not suitable for namespace syntax
// Compare method requires parameters, not suitable for namespace syntax
// Validate method requires parameters, not suitable for namespace syntax
// Parse method requires parameters and returns multiple values, not suitable for namespace syntax
// Format method requires parameters and returns wrong type, not suitable for namespace syntax

// Release namespace methods
func (r Release) Default() error { return r.Release.Default() }
func (r Release) Test() error { return r.Release.Test() }
func (r Release) Snapshot() error { return r.Release.Snapshot() }
func (r Release) Install() error { return r.Release.Install() }
func (r Release) Update() error { return r.Release.Update() }
func (r Release) Check() error { return r.Release.Check() }
func (r Release) Init() error { return r.Release.Init() }
func (r Release) Changelog() error { return r.Release.Changelog() }
func (r Release) Create() error { return r.Release.Create() }
func (r Release) Prepare() error { return r.Release.Prepare() }
func (r Release) Publish() error { return r.Release.Publish() }
func (r Release) Notes() error { return r.Release.Notes() }
func (r Release) Validate() error { return r.Release.Validate() }
func (r Release) Clean() error { return r.Release.Clean() }
func (r Release) Archive() error { return r.Release.Archive() }
// Upload method requires parameters, not suitable for namespace syntax
func (r Release) List() error { return r.Release.List() }
func (r Release) Build() error { return r.Release.Build() }
func (r Release) Package() error { return r.Release.Package() }
func (r Release) Draft() error { return r.Release.Draft() }

// Metrics namespace methods
func (m Metrics) LOC() error { return m.Metrics.LOC() }
func (m Metrics) Coverage() error { return m.Metrics.Coverage() }
func (m Metrics) Complexity() error { return m.Metrics.Complexity() }
func (m Metrics) Size() error { return m.Metrics.Size() }
func (m Metrics) Quality() error { return m.Metrics.Quality() }
func (m Metrics) Imports() error { return m.Metrics.Imports() }

// Install namespace methods
func (i Install) Local() error { return i.Install.Local() }
func (i Install) Uninstall() error { return i.Install.Uninstall() }
func (i Install) Default() error { return i.Install.Default() }
func (i Install) Go() error { return i.Install.Go() }
func (i Install) Stdlib() error { return i.Install.Stdlib() }
func (i Install) Tools() error { return i.Install.Tools() }
func (i Install) SystemWide() error { return i.Install.SystemWide() }
func (i Install) Binary() error { return i.Install.Binary() }
func (i Install) Deps() error { return i.Install.Deps() }
func (i Install) Mage() error { return i.Install.Mage() }
func (i Install) Docker() error { return i.Install.Docker() }
func (i Install) GitHooks() error { return i.Install.GitHooks() }
func (i Install) CI() error { return i.Install.CI() }
func (i Install) Certs() error { return i.Install.Certs() }
func (i Install) Package() error { return i.Install.Package() }
func (i Install) All() error { return i.Install.All() }

// Audit namespace methods
func (a Audit) Show() error { return a.Audit.Show() }
func (a Audit) Stats() error { return a.Audit.Stats() }
func (a Audit) Export() error { return a.Audit.Export() }
func (a Audit) Cleanup() error { return a.Audit.Cleanup() }
func (a Audit) Enable() error { return a.Audit.Enable() }
func (a Audit) Disable() error { return a.Audit.Disable() }
func (a Audit) Report() error { return a.Audit.Report() }

// Configure namespace methods
func (c Configure) Init() error { return c.Configure.Init() }
func (c Configure) Show() error { return c.Configure.Show() }
func (c Configure) Update() error { return c.Configure.Update() }
func (c Configure) Enterprise() error { return c.Configure.Enterprise() }
func (c Configure) Export() error { return c.Configure.Export() }
func (c Configure) Import() error { return c.Configure.Import() }
func (c Configure) Validate() error { return c.Configure.Validate() }
func (c Configure) Schema() error { return c.Configure.Schema() }

// Format namespace methods
func (f Format) Default() error { return f.Format.Default() }
func (f Format) All() error { return f.Format.All() }
func (f Format) Go() error { return f.Format.Go() }
func (f Format) YAML() error { return f.Format.YAML() }
func (f Format) Yaml() error { return f.Format.Yaml() }
func (f Format) JSON() error { return f.Format.JSON() }
func (f Format) Markdown() error { return f.Format.Markdown() }
func (f Format) SQL() error { return f.Format.SQL() }
func (f Format) Dockerfile() error { return f.Format.Dockerfile() }
func (f Format) Shell() error { return f.Format.Shell() }
func (f Format) Fix() error { return f.Format.Fix() }
func (f Format) Gofmt() error { return f.Format.Gofmt() }
func (f Format) Fumpt() error { return f.Format.Fumpt() }
func (f Format) Imports() error { return f.Format.Imports() }
func (f Format) Check() error { return f.Format.Check() }
func (f Format) InstallTools() error { return f.Format.InstallTools() }

// Generate namespace methods
func (g Generate) Default() error { return g.Generate.Default() }
func (g Generate) All() error { return g.Generate.All() }
func (g Generate) Mocks() error { return g.Generate.Mocks() }
func (g Generate) Proto() error { return g.Generate.Proto() }
func (g Generate) Clean() error { return g.Generate.Clean() }
func (g Generate) Check() error { return g.Generate.Check() }
func (g Generate) Code() error { return g.Generate.Code() }
func (g Generate) Docs() error { return g.Generate.Docs() }
func (g Generate) Swagger() error { return g.Generate.Swagger() }
func (g Generate) OpenAPI() error { return g.Generate.OpenAPI() }
func (g Generate) GraphQL() error { return g.Generate.GraphQL() }
func (g Generate) SQL() error { return g.Generate.SQL() }
func (g Generate) Wire() error { return g.Generate.Wire() }
func (g Generate) Config() error { return g.Generate.Config() }

// Help namespace methods
func (h Help) Default() error { return h.Help.Default() }
func (h Help) Commands() error { return h.Help.Commands() }
func (h Help) Command() error { return h.Help.Command() }
func (h Help) Examples() error { return h.Help.Examples() }
func (h Help) GettingStarted() error { return h.Help.GettingStarted() }
func (h Help) Completions() error { return h.Help.Completions() }
func (h Help) Topics() error { return h.Help.Topics() }

// Init namespace methods
func (i Init) Default() error { return i.Init.Default() }
func (i Init) Library() error { return i.Init.Library() }
func (i Init) CLI() error { return i.Init.CLI() }
func (i Init) WebAPI() error { return i.Init.WebAPI() }
func (i Init) Microservice() error { return i.Init.Microservice() }
func (i Init) Tool() error { return i.Init.Tool() }
func (i Init) Upgrade() error { return i.Init.Upgrade() }
func (i Init) Templates() error { return i.Init.Templates() }
func (i Init) Project() error { return i.Init.Project() }
func (i Init) Config() error { return i.Init.Config() }
func (i Init) Git() error { return i.Init.Git() }
func (i Init) Mage() error { return i.Init.Mage() }
func (i Init) CI() error { return i.Init.CI() }
func (i Init) Docker() error { return i.Init.Docker() }
func (i Init) Docs() error { return i.Init.Docs() }
func (i Init) License() error { return i.Init.License() }
func (i Init) Makefile() error { return i.Init.Makefile() }
func (i Init) Editorconfig() error { return i.Init.Editorconfig() }

// Integrations namespace methods
func (i Integrations) Setup() error { return i.Integrations.Setup() }
func (i Integrations) Test() error { return i.Integrations.Test() }
func (i Integrations) Sync() error { return i.Integrations.Sync() }
func (i Integrations) Notify() error { return i.Integrations.Notify() }
func (i Integrations) Status() error { return i.Integrations.Status() }
func (i Integrations) Webhook() error { return i.Integrations.Webhook() }
func (i Integrations) Export() error { return i.Integrations.Export() }
func (i Integrations) Import() error { return i.Integrations.Import() }

// Recipes namespace methods
func (r Recipes) Default() error { return r.Recipes.Default() }
func (r Recipes) List() error { return r.Recipes.List() }
func (r Recipes) Show() error { return r.Recipes.Show() }
func (r Recipes) Run() error { return r.Recipes.Run() }
func (r Recipes) Search() error { return r.Recipes.Search() }
func (r Recipes) Create() error { return r.Recipes.Create() }
func (r Recipes) Install() error { return r.Recipes.Install() }

// Update namespace methods
func (u Update) Check() error { return u.Update.Check() }
func (u Update) Install() error { return u.Update.Install() }
func (u Update) Auto() error { return u.Update.Auto() }
func (u Update) History() error { return u.Update.History() }
func (u Update) Rollback() error { return u.Update.Rollback() }

// Vet namespace methods
func (v Vet) Default() error { return v.Vet.Default() }
func (v Vet) All() error { return v.Vet.All() }
func (v Vet) Parallel() error { return v.Vet.Parallel() }
func (v Vet) Shadow() error { return v.Vet.Shadow() }
func (v Vet) Strict() error { return v.Vet.Strict() }

// Bench namespace methods
func (b Bench) Default() error { return b.Bench.Default() }
func (b Bench) Compare() error { return b.Bench.Compare() }
func (b Bench) Save() error { return b.Bench.Save() }
func (b Bench) CPU() error { return b.Bench.CPU() }
func (b Bench) Mem() error { return b.Bench.Mem() }
func (b Bench) Profile() error { return b.Bench.Profile() }
func (b Bench) Trace() error { return b.Bench.Trace() }
func (b Bench) Regression() error { return b.Bench.Regression() }

// CLI namespace methods
func (c CLI) Bulk() error { return c.CLI.Bulk() }
func (c CLI) Query() error { return c.CLI.Query() }
func (c CLI) Dashboard() error { return c.CLI.Dashboard() }
func (c CLI) Batch() error { return c.CLI.Batch() }
func (c CLI) Monitor() error { return c.CLI.Monitor() }
func (c CLI) Workspace() error { return c.CLI.Workspace() }
func (c CLI) Pipeline() error { return c.CLI.Pipeline() }
func (c CLI) Compliance() error { return c.CLI.Compliance() }
func (c CLI) Default() error { return c.CLI.Default() }
func (c CLI) Help() error { return c.CLI.Help() }
func (c CLI) Version() error { return c.CLI.Version() }
func (c CLI) Completion() error { return c.CLI.Completion() }
func (c CLI) Config() error { return c.CLI.Config() }
func (c CLI) Update() error { return c.CLI.Update() }

// Enterprise namespace methods
func (e Enterprise) Init() error { return e.Enterprise.Init() }
func (e Enterprise) Config() error { return e.Enterprise.Config() }
func (e Enterprise) Deploy() error { return e.Enterprise.Deploy() }
func (e Enterprise) Rollback() error { return e.Enterprise.Rollback() }
func (e Enterprise) Promote() error { return e.Enterprise.Promote() }
func (e Enterprise) Status() error { return e.Enterprise.Status() }
func (e Enterprise) Backup() error { return e.Enterprise.Backup() }
func (e Enterprise) Restore() error { return e.Enterprise.Restore() }

// EnterpriseConfig namespace methods
func (e EnterpriseConfig) Init() error { return e.EnterpriseConfigNamespace.Init() }
func (e EnterpriseConfig) Validate() error { return e.EnterpriseConfigNamespace.Validate() }
func (e EnterpriseConfig) Update() error { return e.EnterpriseConfigNamespace.Update() }
func (e EnterpriseConfig) Export() error { return e.EnterpriseConfigNamespace.Export() }
func (e EnterpriseConfig) Import() error { return e.EnterpriseConfigNamespace.Import() }
func (e EnterpriseConfig) Schema() error { return e.EnterpriseConfigNamespace.Schema() }

// Releases namespace methods
func (r Releases) Create() error { return r.Releases.Create() }
func (r Releases) Publish() error { return r.Releases.Publish() }
func (r Releases) Stable() error { return r.Releases.Stable() }
func (r Releases) Beta() error { return r.Releases.Beta() }
func (r Releases) Edge() error { return r.Releases.Edge() }
func (r Releases) Draft() error { return r.Releases.Draft() }
func (r Releases) Promote() error { return r.Releases.Promote() }
func (r Releases) Status() error { return r.Releases.Status() }
func (r Releases) Channels() error { return r.Releases.Channels() }
func (r Releases) Cleanup() error { return r.Releases.Cleanup() }

// Wizard namespace methods
func (w Wizard) Setup() error { return w.Wizard.Setup() }
func (w Wizard) Project() error { return w.Wizard.Project() }
func (w Wizard) Integration() error { return w.Wizard.Integration() }
func (w Wizard) Security() error { return w.Wizard.Security() }
func (w Wizard) Workflow() error { return w.Wizard.Workflow() }
func (w Wizard) Deployment() error { return w.Wizard.Deployment() }
// Run method not available in wizard namespace
// GetName method not available in wizard namespace
// GetDescription method not available in wizard namespace

// Workflow namespace methods
func (w Workflow) Execute() error { return w.Workflow.Execute() }
func (w Workflow) List() error { return w.Workflow.List() }
func (w Workflow) Status() error { return w.Workflow.Status() }
func (w Workflow) Create() error { return w.Workflow.Create() }
func (w Workflow) Validate() error { return w.Workflow.Validate() }
func (w Workflow) Schedule() error { return w.Workflow.Schedule() }
func (w Workflow) Template() error { return w.Workflow.Template() }
func (w Workflow) History() error { return w.Workflow.History() }

// Yaml namespace methods
func (y Yaml) Init() error { return y.Yaml.Init() }
func (y Yaml) Validate() error { return y.Yaml.Validate() }
func (y Yaml) Show() error { return y.Yaml.Show() }
func (y Yaml) Update() error { return y.Yaml.Update() }
func (y Yaml) Template() error { return y.Yaml.Template() }
*/
