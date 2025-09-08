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

// getMageArgs reads and parses arguments from the MAGE_ARGS environment variable
// This is how magex passes parameters to namespace methods
func getMageArgs() []string {
	if mageArgs := os.Getenv("MAGE_ARGS"); mageArgs != "" {
		return strings.Fields(mageArgs)
	}
	return nil
}

// Namespace types - these enable namespace:method syntax in mage
// Each type must be defined as mg.Namespace for mage to recognize it
type (
	Audit            mg.Namespace
	Bench            mg.Namespace
	Build            mg.Namespace
	CLI              mg.Namespace
	Configure        mg.Namespace
	Deps             mg.Namespace
	Docs             mg.Namespace
	Enterprise       mg.Namespace
	EnterpriseConfig mg.Namespace
	Format           mg.Namespace
	Generate         mg.Namespace
	Git              mg.Namespace
	Help             mg.Namespace
	Init             mg.Namespace
	Install          mg.Namespace
	Integrations     mg.Namespace
	Lint             mg.Namespace
	Metrics          mg.Namespace
	Mod              mg.Namespace
	Recipes          mg.Namespace
	Release          mg.Namespace
	Test             mg.Namespace
	Tools            mg.Namespace
	Update           mg.Namespace
	Version          mg.Namespace
	Vet              mg.Namespace
	Wizard           mg.Namespace
	Workflow         mg.Namespace
	Yaml             mg.Namespace
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
	// Get arguments from MAGE_ARGS environment variable if available
	// This is set by magex when delegating to mage
	var args []string
	if mageArgs := os.Getenv("MAGE_ARGS"); mageArgs != "" {
		args = strings.Fields(mageArgs)
	}
	return v.Bump(args...)
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

func (b Build) Install() error {
	var impl mage.Build
	return impl.Install()
}

func (b Build) Dev() error {
	var impl mage.Build
	return impl.Dev()
}

// Test namespace methods
func (t Test) Default() error {
	var impl mage.Test
	return impl.Default(getMageArgs()...)
}

func (t Test) Full() error {
	var impl mage.Test
	return impl.Full(getMageArgs()...)
}

func (t Test) Unit() error {
	var impl mage.Test
	return impl.Unit(getMageArgs()...)
}

func (t Test) Short() error {
	var impl mage.Test
	return impl.Short(getMageArgs()...)
}

func (t Test) Race() error {
	var impl mage.Test
	return impl.Race(getMageArgs()...)
}

func (t Test) Cover() error {
	var impl mage.Test
	return impl.Cover(getMageArgs()...)
}

func (t Test) CoverRace() error {
	var impl mage.Test
	return impl.CoverRace(getMageArgs()...)
}

func (t Test) CoverReport() error {
	var impl mage.Test
	return impl.CoverReport()
}

func (t Test) CoverHTML() error {
	var impl mage.Test
	return impl.CoverHTML()
}

func (t Test) Fuzz() error {
	var impl mage.Test
	return impl.Fuzz(getMageArgs()...)
}

func (t Test) FuzzShort() error {
	var impl mage.Test
	return impl.FuzzShort(getMageArgs()...)
}

func (t Test) Bench() error {
	var impl mage.Test
	return impl.Bench(getMageArgs()...)
}

func (t Test) BenchShort() error {
	var impl mage.Test
	return impl.BenchShort(getMageArgs()...)
}

func (t Test) Integration() error {
	var impl mage.Test
	return impl.Integration()
}

func (t Test) CI() error {
	var impl mage.Test
	return impl.CI()
}

func (t Test) Parallel() error {
	var impl mage.Test
	return impl.Parallel()
}

func (t Test) NoLint() error {
	var impl mage.Test
	return impl.NoLint()
}

func (t Test) CINoRace() error {
	var impl mage.Test
	return impl.CINoRace()
}

func (t Test) Run() error {
	var impl mage.Test
	return impl.Run()
}

func (t Test) Coverage() error {
	var impl mage.Test
	return impl.Coverage(getMageArgs()...)
}

// Lint namespace methods
func (l Lint) Default() error {
	var impl mage.Lint
	return impl.Default()
}

func (l Lint) Fix() error {
	var impl mage.Lint
	return impl.Fix()
}

func (l Lint) Fmt() error {
	var impl mage.Lint
	return impl.Fmt()
}

func (l Lint) Fumpt() error {
	var impl mage.Lint
	return impl.Fumpt()
}

func (l Lint) Vet() error {
	var impl mage.Lint
	return impl.Vet()
}

func (l Lint) VetParallel() error {
	var impl mage.Lint
	return impl.VetParallel()
}

func (l Lint) Version() error {
	var impl mage.Lint
	return impl.Version()
}

func (l Lint) All() error {
	var impl mage.Lint
	return impl.All()
}

func (l Lint) Go() error {
	var impl mage.Lint
	return impl.Go()
}

func (l Lint) Docker() error {
	var impl mage.Lint
	return impl.Docker()
}

func (l Lint) YAML() error {
	var impl mage.Lint
	return impl.YAML()
}

func (l Lint) Markdown() error {
	var impl mage.Lint
	return impl.Markdown()
}

func (l Lint) Shell() error {
	var impl mage.Lint
	return impl.Shell()
}

func (l Lint) JSON() error {
	var impl mage.Lint
	return impl.JSON()
}

func (l Lint) SQL() error {
	var impl mage.Lint
	return impl.SQL()
}

func (l Lint) Config() error {
	var impl mage.Lint
	return impl.Config()
}

func (l Lint) CI() error {
	var impl mage.Lint
	return impl.CI()
}

func (l Lint) Fast() error {
	var impl mage.Lint
	return impl.Fast()
}

// Deps namespace methods
func (d Deps) Default() error {
	var impl mage.Deps
	return impl.Default()
}

func (d Deps) Download() error {
	var impl mage.Deps
	return impl.Download()
}

func (d Deps) Tidy() error {
	var impl mage.Deps
	return impl.Tidy()
}

func (d Deps) Update() error {
	var impl mage.Deps
	return impl.Update()
}

func (d Deps) Clean() error {
	var impl mage.Deps
	return impl.Clean()
}

func (d Deps) Graph() error {
	var impl mage.Deps
	return impl.Graph()
}

// Why method requires parameters, not suitable for namespace syntax
func (d Deps) Verify() error {
	var impl mage.Deps
	return impl.Verify()
}

func (d Deps) VulnCheck() error {
	var impl mage.Deps
	return impl.VulnCheck()
}

func (d Deps) List() error {
	var impl mage.Deps
	return impl.List()
}

func (d Deps) Outdated() error {
	var impl mage.Deps
	return impl.Outdated()
}

func (d Deps) Vendor() error {
	var impl mage.Deps
	return impl.Vendor()
}

// Init method requires parameters, not suitable for namespace syntax
func (d Deps) Audit() error {
	var impl mage.Deps
	return impl.Audit()
}

func (d Deps) Licenses() error {
	var impl mage.Deps
	return impl.Licenses()
}

func (d Deps) Check() error {
	var impl mage.Deps
	return impl.Check()
}

// Tools namespace methods
func (t Tools) Default() error {
	var impl mage.Tools
	return impl.Default()
}

func (t Tools) Install() error {
	var impl mage.Tools
	return impl.Install()
}

func (t Tools) Update() error {
	var impl mage.Tools
	return impl.Update()
}

func (t Tools) Verify() error {
	var impl mage.Tools
	return impl.Verify()
}

func (t Tools) List() error {
	var impl mage.Tools
	return impl.List()
}

func (t Tools) VulnCheck() error {
	var impl mage.Tools
	return impl.VulnCheck()
}

func (t Tools) Check() error {
	var impl mage.Tools
	return impl.Check()
}

func (t Tools) Clean() error {
	var impl mage.Tools
	return impl.Clean()
}

// Mod namespace methods
func (m Mod) Download() error {
	var impl mage.Mod
	return impl.Download()
}

func (m Mod) Tidy() error {
	var impl mage.Mod
	return impl.Tidy()
}

func (m Mod) Update() error {
	var impl mage.Mod
	return impl.Update()
}

func (m Mod) Clean() error {
	var impl mage.Mod
	return impl.Clean()
}

func (m Mod) Graph() error {
	var impl mage.Mod
	return impl.Graph(getMageArgs()...)
}

func (m Mod) Why() error {
	var impl mage.Mod
	return impl.Why(getMageArgs()...)
}

func (m Mod) Vendor() error {
	var impl mage.Mod
	return impl.Vendor()
}

func (m Mod) Init() error {
	var impl mage.Mod
	return impl.Init()
}

func (m Mod) Verify() error {
	var impl mage.Mod
	return impl.Verify()
}

func (m Mod) Edit() error {
	var impl mage.Mod
	return impl.Edit(getMageArgs()...)
}

func (m Mod) Get() error {
	var impl mage.Mod
	return impl.Get(getMageArgs()...)
}

func (m Mod) List() error {
	var impl mage.Mod
	return impl.List(getMageArgs()...)
}

// Docs namespace methods
func (d Docs) Default() error {
	var impl mage.Docs
	return impl.Default()
}

func (d Docs) Generate() error {
	var impl mage.Docs
	return impl.Generate()
}

func (d Docs) Serve() error {
	var impl mage.Docs
	return impl.Serve()
}

func (d Docs) ServeDefault() error {
	var impl mage.Docs
	return impl.ServeDefault()
}

func (d Docs) ServePkgsite() error {
	var impl mage.Docs
	return impl.ServePkgsite()
}

func (d Docs) ServeGodoc() error {
	var impl mage.Docs
	return impl.ServeGodoc()
}

func (d Docs) ServeStdlib() error {
	var impl mage.Docs
	return impl.ServeStdlib()
}

func (d Docs) ServeProject() error {
	var impl mage.Docs
	return impl.ServeProject()
}

func (d Docs) ServeBoth() error {
	var impl mage.Docs
	return impl.ServeBoth()
}

func (d Docs) Check() error {
	var impl mage.Docs
	return impl.Check()
}

func (d Docs) Examples() error {
	var impl mage.Docs
	return impl.Examples()
}

func (d Docs) Build() error {
	var impl mage.Docs
	return impl.Build()
}

func (d Docs) Lint() error {
	var impl mage.Docs
	return impl.Lint()
}

func (d Docs) Spell() error {
	var impl mage.Docs
	return impl.Spell()
}

func (d Docs) Links() error {
	var impl mage.Docs
	return impl.Links()
}

func (d Docs) API() error {
	var impl mage.Docs
	return impl.API()
}

func (d Docs) Markdown() error {
	var impl mage.Docs
	return impl.Markdown()
}

func (d Docs) Readme() error {
	var impl mage.Docs
	return impl.Readme()
}

// Changelog method requires parameters, not suitable for namespace syntax
func (d Docs) GoDocs() error {
	var impl mage.Docs
	return impl.GoDocs(getMageArgs()...)
}

// Git namespace methods
func (g Git) Diff() error {
	var impl mage.Git
	return impl.Diff()
}

func (g Git) Tag() error {
	var impl mage.Git
	return impl.Tag()
}

func (g Git) TagRemove() error {
	var impl mage.Git
	return impl.TagRemove()
}

func (g Git) TagUpdate() error {
	var impl mage.Git
	return impl.TagUpdate(getMageArgs()...)
}

func (g Git) Status() error {
	var impl mage.Git
	return impl.Status()
}

func (g Git) Log() error {
	var impl mage.Git
	return impl.Log()
}

func (g Git) Branch() error {
	var impl mage.Git
	return impl.Branch()
}

func (g Git) Pull() error {
	var impl mage.Git
	return impl.Pull()
}

func (g Git) Commit() error {
	var impl mage.Git
	return impl.Commit(getMageArgs()...)
}

func (g Git) Init() error {
	var impl mage.Git
	return impl.Init()
}

func (g Git) Add() error {
	var impl mage.Git
	return impl.Add(getMageArgs()...)
}

func (g Git) Clone() error {
	var impl mage.Git
	return impl.Clone()
}

// Push method requires parameters, not suitable for namespace syntax
// PullWithRemote method requires parameters, not suitable for namespace syntax
// TagWithMessage method requires parameters, not suitable for namespace syntax
// BranchWithName method requires parameters, not suitable for namespace syntax
// CloneRepo method requires parameters, not suitable for namespace syntax
// LogWithCount method requires parameters, not suitable for namespace syntax

// Version namespace methods
func (v Version) Show() error {
	var impl mage.Version
	return impl.Show()
}

func (v Version) Check() error {
	var impl mage.Version
	return impl.Check(getMageArgs()...)
}

func (v Version) Update() error {
	var impl mage.Version
	return impl.Update()
}

func (v Version) Bump() error {
	var impl mage.Version
	return impl.Bump(getMageArgs()...)
}

func (v Version) Changelog() error {
	var impl mage.Version
	return impl.Changelog(getMageArgs()...)
}

func (v Version) Tag() error {
	var impl mage.Version
	return impl.Tag(getMageArgs()...)
}

// Next method requires parameters and returns multiple values, not suitable for namespace syntax
// Compare method requires parameters, not suitable for namespace syntax
// Validate method requires parameters, not suitable for namespace syntax
// Parse method requires parameters and returns multiple values, not suitable for namespace syntax
// Format method requires parameters and returns wrong type, not suitable for namespace syntax

// Release namespace methods
func (r Release) Default() error {
	var impl mage.Release
	return impl.Default(getMageArgs()...)
}

func (r Release) Test() error {
	var impl mage.Release
	return impl.Test()
}

func (r Release) Snapshot() error {
	var impl mage.Release
	return impl.Snapshot()
}

func (r Release) LocalInstall() error {
	var impl mage.Release
	return impl.LocalInstall()
}

func (r Release) Check() error {
	var impl mage.Release
	return impl.Check()
}

func (r Release) Init() error {
	var impl mage.Release
	return impl.Init()
}

func (r Release) Changelog() error {
	var impl mage.Release
	return impl.Changelog()
}

func (r Release) Validate() error {
	var impl mage.Release
	return impl.Validate()
}

func (r Release) Clean() error {
	var impl mage.Release
	return impl.Clean()
}

// Metrics namespace methods
func (m Metrics) LOC() error {
	var impl mage.Metrics
	return impl.LOC()
}

func (m Metrics) Coverage() error {
	var impl mage.Metrics
	return impl.Coverage()
}

func (m Metrics) Complexity() error {
	var impl mage.Metrics
	return impl.Complexity()
}

func (m Metrics) Size() error {
	var impl mage.Metrics
	return impl.Size()
}

func (m Metrics) Quality() error {
	var impl mage.Metrics
	return impl.Quality()
}

func (m Metrics) Imports() error {
	var impl mage.Metrics
	return impl.Imports()
}

// Install namespace methods
func (i Install) Local() error {
	var impl mage.Install
	return impl.Local()
}

func (i Install) Uninstall() error {
	var impl mage.Install
	return impl.Uninstall()
}

func (i Install) Default() error {
	var impl mage.Install
	return impl.Default()
}

func (i Install) Go() error {
	var impl mage.Install
	return impl.Go()
}

func (i Install) Stdlib() error {
	var impl mage.Install
	return impl.Stdlib()
}

func (i Install) Tools() error {
	var impl mage.Install
	return impl.Tools()
}

func (i Install) SystemWide() error {
	var impl mage.Install
	return impl.SystemWide()
}

func (i Install) Binary() error {
	var impl mage.Install
	return impl.Binary()
}

func (i Install) Deps() error {
	var impl mage.Install
	return impl.Deps()
}

func (i Install) Mage() error {
	var impl mage.Install
	return impl.Mage()
}

func (i Install) Docker() error {
	var impl mage.Install
	return impl.Docker()
}

func (i Install) GitHooks() error {
	var impl mage.Install
	return impl.GitHooks()
}

func (i Install) CI() error {
	var impl mage.Install
	return impl.CI()
}

func (i Install) Certs() error {
	var impl mage.Install
	return impl.Certs()
}

func (i Install) Package() error {
	var impl mage.Install
	return impl.Package()
}

func (i Install) All() error {
	var impl mage.Install
	return impl.All()
}

// Audit namespace methods
func (a Audit) Show() error {
	var impl mage.Audit
	return impl.Show()
}

func (a Audit) Stats() error {
	var impl mage.Audit
	return impl.Stats()
}

func (a Audit) Export() error {
	var impl mage.Audit
	return impl.Export()
}

func (a Audit) Cleanup() error {
	var impl mage.Audit
	return impl.Cleanup()
}

func (a Audit) Enable() error {
	var impl mage.Audit
	return impl.Enable()
}

func (a Audit) Disable() error {
	var impl mage.Audit
	return impl.Disable()
}

func (a Audit) Report() error {
	var impl mage.Audit
	return impl.Report()
}

// Configure namespace methods
func (c Configure) Init() error {
	var impl mage.Configure
	return impl.Init()
}

func (c Configure) Show() error {
	var impl mage.Configure
	return impl.Show()
}

func (c Configure) Update() error {
	var impl mage.Configure
	return impl.Update()
}

func (c Configure) Enterprise() error {
	var impl mage.Configure
	return impl.Enterprise()
}

func (c Configure) Export() error {
	var impl mage.Configure
	return impl.Export()
}

func (c Configure) Import() error {
	var impl mage.Configure
	return impl.Import()
}

func (c Configure) Validate() error {
	var impl mage.Configure
	return impl.Validate()
}

func (c Configure) Schema() error {
	var impl mage.Configure
	return impl.Schema()
}

// Format namespace methods
func (f Format) Default() error {
	var impl mage.Format
	return impl.Default()
}

func (f Format) All() error {
	var impl mage.Format
	return impl.All()
}

func (f Format) Go() error {
	var impl mage.Format
	return impl.Go()
}

func (f Format) YAML() error {
	var impl mage.Format
	return impl.YAML()
}

func (f Format) JSON() error {
	var impl mage.Format
	return impl.JSON()
}

func (f Format) Markdown() error {
	var impl mage.Format
	return impl.Markdown()
}

func (f Format) SQL() error {
	var impl mage.Format
	return impl.SQL()
}

func (f Format) Dockerfile() error {
	var impl mage.Format
	return impl.Dockerfile()
}

func (f Format) Shell() error {
	var impl mage.Format
	return impl.Shell()
}

func (f Format) Fix() error {
	var impl mage.Format
	return impl.Fix()
}

func (f Format) Gofmt() error {
	var impl mage.Format
	return impl.Gofmt()
}

func (f Format) Fumpt() error {
	var impl mage.Format
	return impl.Fumpt()
}

func (f Format) Imports() error {
	var impl mage.Format
	return impl.Imports()
}

func (f Format) Check() error {
	var impl mage.Format
	return impl.Check()
}

func (f Format) InstallTools() error {
	var impl mage.Format
	return impl.InstallTools()
}

// Generate namespace methods
func (g Generate) Default() error {
	var impl mage.Generate
	return impl.Default()
}

func (g Generate) All() error {
	var impl mage.Generate
	return impl.All()
}

func (g Generate) Mocks() error {
	var impl mage.Generate
	return impl.Mocks()
}

func (g Generate) Proto() error {
	var impl mage.Generate
	return impl.Proto()
}

func (g Generate) Clean() error {
	var impl mage.Generate
	return impl.Clean()
}

func (g Generate) Check() error {
	var impl mage.Generate
	return impl.Check()
}

func (g Generate) Code() error {
	var impl mage.Generate
	return impl.Code()
}

func (g Generate) Docs() error {
	var impl mage.Generate
	return impl.Docs()
}

func (g Generate) Swagger() error {
	var impl mage.Generate
	return impl.Swagger()
}

func (g Generate) OpenAPI() error {
	var impl mage.Generate
	return impl.OpenAPI()
}

func (g Generate) GraphQL() error {
	var impl mage.Generate
	return impl.GraphQL()
}

func (g Generate) SQL() error {
	var impl mage.Generate
	return impl.SQL()
}

func (g Generate) Wire() error {
	var impl mage.Generate
	return impl.Wire()
}

func (g Generate) Config() error {
	var impl mage.Generate
	return impl.Config()
}

// Help namespace methods
func (h Help) Default() error {
	var impl mage.Help
	return impl.Default()
}

func (h Help) Commands() error {
	var impl mage.Help
	return impl.Commands()
}

func (h Help) Command() error {
	var impl mage.Help
	return impl.Command()
}

func (h Help) Examples() error {
	var impl mage.Help
	return impl.Examples()
}

func (h Help) GettingStarted() error {
	var impl mage.Help
	return impl.GettingStarted()
}

func (h Help) Completions() error {
	var impl mage.Help
	return impl.Completions()
}

func (h Help) Topics() error {
	var impl mage.Help
	return impl.Topics()
}

// Init namespace methods
func (i Init) Default() error {
	var impl mage.Init
	return impl.Default()
}

func (i Init) Library() error {
	var impl mage.Init
	return impl.Library()
}

func (i Init) CLI() error {
	var impl mage.Init
	return impl.CLI()
}

func (i Init) WebAPI() error {
	var impl mage.Init
	return impl.WebAPI()
}

func (i Init) Microservice() error {
	var impl mage.Init
	return impl.Microservice()
}

func (i Init) Tool() error {
	var impl mage.Init
	return impl.Tool()
}

func (i Init) Upgrade() error {
	var impl mage.Init
	return impl.Upgrade()
}

func (i Init) Templates() error {
	var impl mage.Init
	return impl.Templates()
}

func (i Init) Project() error {
	var impl mage.Init
	return impl.Project()
}

func (i Init) Config() error {
	var impl mage.Init
	return impl.Config()
}

func (i Init) Git() error {
	var impl mage.Init
	return impl.Git()
}

func (i Init) Mage() error {
	var impl mage.Init
	return impl.Mage()
}

func (i Init) CI() error {
	var impl mage.Init
	return impl.CI()
}

func (i Init) Docker() error {
	var impl mage.Init
	return impl.Docker()
}

func (i Init) Docs() error {
	var impl mage.Init
	return impl.Docs()
}

func (i Init) License() error {
	var impl mage.Init
	return impl.License()
}

func (i Init) Makefile() error {
	var impl mage.Init
	return impl.Makefile()
}

func (i Init) Editorconfig() error {
	var impl mage.Init
	return impl.Editorconfig()
}

// Integrations namespace methods
func (i Integrations) Setup() error {
	var impl mage.Integrations
	return impl.Setup()
}

func (i Integrations) Test() error {
	var impl mage.Integrations
	return impl.Test()
}

func (i Integrations) Sync() error {
	var impl mage.Integrations
	return impl.Sync()
}

func (i Integrations) Notify() error {
	var impl mage.Integrations
	return impl.Notify()
}

func (i Integrations) Status() error {
	var impl mage.Integrations
	return impl.Status()
}

func (i Integrations) Webhook() error {
	var impl mage.Integrations
	return impl.Webhook()
}

func (i Integrations) Export() error {
	var impl mage.Integrations
	return impl.Export()
}

func (i Integrations) Import() error {
	var impl mage.Integrations
	return impl.Import()
}

// Recipes namespace methods
func (r Recipes) Default() error {
	var impl mage.Recipes
	return impl.Default()
}

func (r Recipes) List() error {
	var impl mage.Recipes
	return impl.List()
}

func (r Recipes) Show() error {
	var impl mage.Recipes
	return impl.Show()
}

func (r Recipes) Run() error {
	var impl mage.Recipes
	return impl.Run()
}

func (r Recipes) Search() error {
	var impl mage.Recipes
	return impl.Search()
}

func (r Recipes) Create() error {
	var impl mage.Recipes
	return impl.Create()
}

func (r Recipes) Install() error {
	var impl mage.Recipes
	return impl.Install()
}

// Update namespace methods
func (u Update) Check() error {
	var impl mage.Update
	return impl.Check()
}

func (u Update) Install() error {
	var impl mage.Update
	return impl.Install()
}

// Vet namespace methods
func (v Vet) Default() error {
	var impl mage.Vet
	return impl.Default()
}

func (v Vet) All() error {
	var impl mage.Vet
	return impl.All()
}

func (v Vet) Parallel() error {
	var impl mage.Vet
	return impl.Parallel()
}

func (v Vet) Shadow() error {
	var impl mage.Vet
	return impl.Shadow()
}

func (v Vet) Strict() error {
	var impl mage.Vet
	return impl.Strict()
}

// Bench namespace methods
func (b Bench) Default() error {
	var impl mage.Bench
	return impl.Default()
}

func (b Bench) Compare() error {
	var impl mage.Bench
	return impl.Compare()
}

func (b Bench) Save() error {
	var impl mage.Bench
	return impl.Save()
}

func (b Bench) CPU() error {
	var impl mage.Bench
	return impl.CPU()
}

func (b Bench) Mem() error {
	var impl mage.Bench
	return impl.Mem()
}

func (b Bench) Profile() error {
	var impl mage.Bench
	return impl.Profile()
}

func (b Bench) Trace() error {
	var impl mage.Bench
	return impl.Trace()
}

func (b Bench) Regression() error {
	var impl mage.Bench
	return impl.Regression()
}

// CLI namespace methods
func (c CLI) Bulk() error {
	var impl mage.CLI
	return impl.Bulk()
}

func (c CLI) Query() error {
	var impl mage.CLI
	return impl.Query()
}

func (c CLI) Dashboard() error {
	var impl mage.CLI
	return impl.Dashboard()
}

func (c CLI) Batch() error {
	var impl mage.CLI
	return impl.Batch()
}

func (c CLI) Monitor() error {
	var impl mage.CLI
	return impl.Monitor()
}

func (c CLI) Workspace() error {
	var impl mage.CLI
	return impl.Workspace()
}

func (c CLI) Pipeline() error {
	var impl mage.CLI
	return impl.Pipeline()
}

func (c CLI) Compliance() error {
	var impl mage.CLI
	return impl.Compliance()
}

func (c CLI) Default() error {
	var impl mage.CLI
	return impl.Default()
}

func (c CLI) Help() error {
	var impl mage.CLI
	return impl.Help()
}

func (c CLI) Version() error {
	var impl mage.CLI
	return impl.Version()
}

func (c CLI) Completion() error {
	var impl mage.CLI
	return impl.Completion()
}

func (c CLI) Config() error {
	var impl mage.CLI
	return impl.Config()
}

func (c CLI) Update() error {
	var impl mage.CLI
	return impl.Update()
}

// Enterprise namespace methods
func (e Enterprise) Init() error {
	var impl mage.Enterprise
	return impl.Init()
}

func (e Enterprise) Config() error {
	var impl mage.Enterprise
	return impl.Config()
}

func (e Enterprise) Deploy() error {
	var impl mage.Enterprise
	return impl.Deploy()
}

func (e Enterprise) Rollback() error {
	var impl mage.Enterprise
	return impl.Rollback()
}

func (e Enterprise) Promote() error {
	var impl mage.Enterprise
	return impl.Promote()
}

func (e Enterprise) Status() error {
	var impl mage.Enterprise
	return impl.Status()
}

func (e Enterprise) Backup() error {
	var impl mage.Enterprise
	return impl.Backup()
}

func (e Enterprise) Restore() error {
	var impl mage.Enterprise
	return impl.Restore()
}

// EnterpriseConfig namespace methods
func (e EnterpriseConfig) Init() error {
	var impl mage.EnterpriseConfigNamespace
	return impl.Init()
}

func (e EnterpriseConfig) Validate() error {
	var impl mage.EnterpriseConfigNamespace
	return impl.Validate()
}

func (e EnterpriseConfig) Update() error {
	var impl mage.EnterpriseConfigNamespace
	return impl.Update()
}

func (e EnterpriseConfig) Export() error {
	var impl mage.EnterpriseConfigNamespace
	return impl.Export()
}

func (e EnterpriseConfig) Import() error {
	var impl mage.EnterpriseConfigNamespace
	return impl.Import()
}

func (e EnterpriseConfig) Schema() error {
	var impl mage.EnterpriseConfigNamespace
	return impl.Schema()
}

// Wizard namespace methods
func (w Wizard) Setup() error {
	var impl mage.Wizard
	return impl.Setup()
}

func (w Wizard) Project() error {
	var impl mage.Wizard
	return impl.Project()
}

func (w Wizard) Integration() error {
	var impl mage.Wizard
	return impl.Integration()
}

func (w Wizard) Security() error {
	var impl mage.Wizard
	return impl.Security()
}

func (w Wizard) Workflow() error {
	var impl mage.Wizard
	return impl.Workflow()
}

func (w Wizard) Deployment() error {
	var impl mage.Wizard
	return impl.Deployment()
}

// Run method not available in wizard namespace
// GetName method not available in wizard namespace
// GetDescription method not available in wizard namespace

// Workflow namespace methods
func (w Workflow) Execute() error {
	var impl mage.Workflow
	return impl.Execute()
}

func (w Workflow) List() error {
	var impl mage.Workflow
	return impl.List()
}

func (w Workflow) Status() error {
	var impl mage.Workflow
	return impl.Status()
}

func (w Workflow) Create() error {
	var impl mage.Workflow
	return impl.Create()
}

func (w Workflow) Validate() error {
	var impl mage.Workflow
	return impl.Validate()
}

func (w Workflow) Schedule() error {
	var impl mage.Workflow
	return impl.Schedule()
}

func (w Workflow) Template() error {
	var impl mage.Workflow
	return impl.Template()
}

func (w Workflow) History() error {
	var impl mage.Workflow
	return impl.History()
}

// Yaml namespace methods
func (y Yaml) Init() error {
	var impl mage.Yaml
	return impl.Init()
}

func (y Yaml) Validate() error {
	var impl mage.Yaml
	return impl.Validate()
}

func (y Yaml) Show() error {
	var impl mage.Yaml
	return impl.Show()
}

func (y Yaml) Update() error {
	var impl mage.Yaml
	return impl.Update()
}

func (y Yaml) Template() error {
	var impl mage.Yaml
	return impl.Template()
}
