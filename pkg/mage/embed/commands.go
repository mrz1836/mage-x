// Package embed provides registration of all built-in MAGE-X commands
// This is the heart of the zero-boilerplate solution - all commands are pre-registered
package embed

import (
	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// RegisterAll registers all built-in MAGE-X commands with the registry
// This is called automatically by the magex binary
func RegisterAll(reg *registry.Registry) {
	if reg == nil {
		reg = registry.Global()
	}

	// Prevent double registration
	if reg.IsRegistered() {
		return
	}
	reg.SetRegistered(true)

	// Register Build namespace commands
	registerBuildCommands(reg)

	// Register Test namespace commands
	registerTestCommands(reg)

	// Register Lint namespace commands
	registerLintCommands(reg)

	// Register Format namespace commands
	registerFormatCommands(reg)

	// Register Deps namespace commands
	registerDepsCommands(reg)

	// Register Git namespace commands
	registerGitCommands(reg)

	// Register Release namespace commands
	registerReleaseCommands(reg)

	// Register Docs namespace commands
	registerDocsCommands(reg)

	// Register Tools namespace commands
	registerToolsCommands(reg)

	// Register Generate namespace commands
	registerGenerateCommands(reg)

	// Register CLI namespace commands
	registerCLICommands(reg)

	// Register Update namespace commands
	registerUpdateCommands(reg)

	// Register Mod namespace commands
	registerModCommands(reg)

	// Register Recipes namespace commands
	registerRecipesCommands(reg)

	// Register Metrics namespace commands
	registerMetricsCommands(reg)

	// Register Workflow namespace commands
	registerWorkflowCommands(reg)

	// Register Bench namespace commands
	registerBenchCommands(reg)

	// Register Vet namespace commands
	registerVetCommands(reg)

	// Register Configure namespace commands
	registerConfigureCommands(reg)

	// Register Init namespace commands
	registerInitCommands(reg)

	// Register Enterprise namespace commands
	registerEnterpriseCommands(reg)

	// Register Integrations namespace commands
	registerIntegrationsCommands(reg)

	// Register Wizard namespace commands
	registerWizardCommands(reg)

	// Register Help namespace commands
	registerHelpCommands(reg)

	// Register Version namespace commands
	registerVersionCommands(reg)

	// Register Install namespace commands
	registerInstallCommands(reg)

	// Register Audit namespace commands
	registerAuditCommands(reg)

	// Register Yaml namespace commands
	registerYamlCommands(reg)

	// Register Releases namespace commands
	registerReleasesCommands(reg)

	// Register EnterpriseConfig namespace commands
	registerEnterpriseConfigCommands(reg)

	// Register top-level convenience commands
	registerTopLevelCommands(reg)
}

// registerBuildCommands registers all Build namespace commands
func registerBuildCommands(reg *registry.Registry) {
	b := mage.Build{}

	reg.MustRegister(
		registry.NewNamespaceCommand("build", "default").
			WithDescription("Build the application for the current platform").
			WithFunc(func() error { return b.Default() }).
			WithCategory("Build").
			WithAliases("build").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("build", "all").
			WithDescription("Build for all configured platforms").
			WithFunc(func() error { return b.All() }).
			WithCategory("Build").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("build", "linux").
			WithDescription("Build for Linux (amd64)").
			WithFunc(func() error { return b.Linux() }).
			WithCategory("Build").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("build", "darwin").
			WithDescription("Build for macOS (amd64 and arm64)").
			WithFunc(func() error { return b.Darwin() }).
			WithCategory("Build").
			WithAliases("build:mac", "build:macos").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("build", "windows").
			WithDescription("Build for Windows (amd64)").
			WithFunc(func() error { return b.Windows() }).
			WithCategory("Build").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("build", "docker").
			WithDescription("Build a Docker image").
			WithFunc(func() error { return b.Docker() }).
			WithCategory("Build").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("build", "clean").
			WithDescription("Remove build artifacts").
			WithFunc(func() error { return b.Clean() }).
			WithCategory("Build").
			WithAliases("clean").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("build", "install").
			WithDescription("Install the binary to $GOPATH/bin").
			WithFunc(func() error { return b.Install() }).
			WithCategory("Build").
			WithAliases("install").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("build", "generate").
			WithDescription("Run go generate").
			WithFunc(func() error { return b.Generate() }).
			WithCategory("Build").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("build", "prebuild").
			WithDescription("Pre-build all packages to warm cache").
			WithFunc(func() error { return b.PreBuild() }).
			WithCategory("Build").
			MustBuild(),
	)
}

// registerTestCommands registers all Test namespace commands
func registerTestCommands(reg *registry.Registry) {
	t := mage.Test{}

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "default").
			WithDescription("Run standard test suite").
			WithFunc(func() error { return t.Default() }).
			WithCategory("Test").
			WithAliases("test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "unit").
			WithDescription("Run unit tests only").
			WithFunc(func() error { return t.Unit() }).
			WithCategory("Test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "short").
			WithDescription("Run short tests").
			WithFunc(func() error { return t.Short() }).
			WithCategory("Test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "race").
			WithDescription("Run tests with race detector").
			WithFunc(func() error { return t.Race() }).
			WithCategory("Test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "cover").
			WithDescription("Run tests with coverage").
			WithFunc(func() error { return t.Cover() }).
			WithCategory("Test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "coverrace").
			WithDescription("Run tests with coverage and race detector").
			WithFunc(func() error { return t.CoverRace() }).
			WithCategory("Test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "coverreport").
			WithDescription("Generate coverage report").
			WithFunc(func() error { return t.CoverReport() }).
			WithCategory("Test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "coverhtml").
			WithDescription("Generate HTML coverage report").
			WithFunc(func() error { return t.CoverHTML() }).
			WithCategory("Test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "fuzz").
			WithDescription("Run fuzz tests").
			WithFunc(func() error { return t.Fuzz() }).
			WithCategory("Test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "bench").
			WithDescription("Run benchmarks").
			WithArgsFunc(func(args ...string) error { return t.Bench(args...) }).
			WithCategory("Test").
			WithAliases("benchmark").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "integration").
			WithDescription("Run integration tests").
			WithFunc(func() error { return t.Integration() }).
			WithCategory("Test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "ci").
			WithDescription("Run CI test suite").
			WithFunc(func() error { return t.CI() }).
			WithCategory("Test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "parallel").
			WithDescription("Run tests in parallel").
			WithFunc(func() error { return t.Parallel() }).
			WithCategory("Test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "nolint").
			WithDescription("Run tests without linting").
			WithFunc(func() error { return t.NoLint() }).
			WithCategory("Test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "cinorace").
			WithDescription("Run CI tests without race detector").
			WithFunc(func() error { return t.CINoRace() }).
			WithCategory("Test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "run").
			WithDescription("Run a specific test pattern").
			WithFunc(func() error { return t.Run() }).
			WithCategory("Test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "coverage").
			WithDescription("Run tests and generate coverage").
			WithArgsFunc(func(args ...string) error { return t.Coverage(args...) }).
			WithCategory("Test").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "vet").
			WithDescription("Run go vet").
			WithFunc(func() error { return t.Vet() }).
			WithCategory("Test").
			MustBuild(),
	)
}

// registerLintCommands registers all Lint namespace commands
func registerLintCommands(reg *registry.Registry) {
	l := mage.Lint{}

	reg.MustRegister(
		registry.NewNamespaceCommand("lint", "default").
			WithDescription("Run default linting").
			WithFunc(func() error { return l.Default() }).
			WithCategory("Lint").
			WithAliases("lint").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("lint", "fix").
			WithDescription("Fix auto-fixable lint issues").
			WithFunc(func() error { return l.Fix() }).
			WithCategory("Lint").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("lint", "ci").
			WithDescription("Run CI linting (strict)").
			WithFunc(func() error { return l.CI() }).
			WithCategory("Lint").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("lint", "fast").
			WithDescription("Run fast linting checks").
			WithFunc(func() error { return l.Fast() }).
			WithCategory("Lint").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("lint", "issues").
			WithDescription("Scan for TODOs, FIXMEs, nolint directives, and test skips").
			WithFunc(func() error { return l.Issues() }).
			WithCategory("Lint").
			MustBuild(),
	)
}

// registerFormatCommands registers all Format namespace commands
func registerFormatCommands(reg *registry.Registry) {
	f := mage.Format{}

	reg.MustRegister(
		registry.NewNamespaceCommand("format", "default").
			WithDescription("Format Go code").
			WithFunc(func() error { return f.Default() }).
			WithCategory("Format").
			WithAliases("format", "fmt").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("format", "check").
			WithDescription("Check if code is formatted").
			WithFunc(func() error { return f.Check() }).
			WithCategory("Format").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("format", "fix").
			WithDescription("Fix formatting issues").
			WithFunc(func() error { return f.Fix() }).
			WithCategory("Format").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("format", "imports").
			WithDescription("Fix import statements").
			WithFunc(func() error { return f.Imports() }).
			WithCategory("Format").
			MustBuild(),
	)
}

// registerDepsCommands registers all Deps namespace commands
func registerDepsCommands(reg *registry.Registry) {
	d := mage.Deps{}

	reg.MustRegister(
		registry.NewNamespaceCommand("deps", "default").
			WithDescription("Manage dependencies").
			WithFunc(func() error { return d.Default() }).
			WithCategory("Dependencies").
			WithAliases("deps").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("deps", "update").
			WithDescription("Update all dependencies").
			WithFunc(func() error { return d.Update() }).
			WithCategory("Dependencies").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("deps", "tidy").
			WithDescription("Run go mod tidy").
			WithFunc(func() error { return d.Tidy() }).
			WithCategory("Dependencies").
			WithAliases("tidy").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("deps", "download").
			WithDescription("Download dependencies").
			WithFunc(func() error { return d.Download() }).
			WithCategory("Dependencies").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("deps", "vendor").
			WithDescription("Vendor dependencies").
			WithFunc(func() error { return d.Vendor() }).
			WithCategory("Dependencies").
			WithAliases("vendor").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("deps", "verify").
			WithDescription("Verify dependencies").
			WithFunc(func() error { return d.Verify() }).
			WithCategory("Dependencies").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("deps", "audit").
			WithDescription("Security audit of dependencies").
			WithFunc(func() error { return d.Audit() }).
			WithCategory("Dependencies").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("deps", "outdated").
			WithDescription("List outdated dependencies").
			WithFunc(func() error { return d.Outdated() }).
			WithCategory("Dependencies").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("deps", "graph").
			WithDescription("Show dependency graph").
			WithFunc(func() error { return d.Graph() }).
			WithCategory("Dependencies").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("deps", "licenses").
			WithDescription("List dependency licenses").
			WithFunc(func() error { return d.Licenses() }).
			WithCategory("Dependencies").
			MustBuild(),
	)
}

// Continue with other namespace registrations...
// For brevity, I'll add a few more key namespaces

// registerTopLevelCommands registers convenience top-level commands
func registerTopLevelCommands(reg *registry.Registry) {
	// These are aliases for the most common commands
	b := mage.Build{}
	t := mage.Test{}
	l := mage.Lint{}
	f := mage.Format{}

	reg.MustRegister(
		registry.NewCommand("build").
			WithDescription("Build the application").
			WithFunc(func() error { return b.Default() }).
			WithCategory("Common").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewCommand("test").
			WithDescription("Run tests").
			WithFunc(func() error { return t.Default() }).
			WithCategory("Common").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewCommand("lint").
			WithDescription("Run linter").
			WithFunc(func() error { return l.Default() }).
			WithCategory("Common").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewCommand("format").
			WithDescription("Format code").
			WithFunc(func() error { return f.Default() }).
			WithCategory("Common").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewCommand("clean").
			WithDescription("Clean build artifacts").
			WithFunc(func() error { return b.Clean() }).
			WithCategory("Common").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewCommand("install").
			WithDescription("Install the application").
			WithFunc(func() error { return b.Install() }).
			WithCategory("Common").
			MustBuild(),
	)
}

// Stub implementations for remaining namespaces
// These would be fully implemented following the same pattern

func registerGitCommands(reg *registry.Registry) {
	g := mage.Git{}

	reg.MustRegister(
		registry.NewNamespaceCommand("git", "status").
			WithDescription("Show git status").
			WithFunc(func() error { return g.Status() }).
			WithCategory("Git").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("git", "diff").
			WithDescription("Show git diff and check for uncommitted changes").
			WithFunc(func() error { return g.Diff() }).
			WithCategory("Git").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("git", "tag").
			WithDescription("Create a git tag from version").
			WithFunc(func() error { return g.Tag() }).
			WithCategory("Git").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("git", "tagremove").
			WithDescription("Remove a git tag").
			WithFunc(func() error { return g.TagRemove() }).
			WithCategory("Git").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("git", "tagupdate").
			WithDescription("Update a git tag").
			WithFunc(func() error { return g.TagUpdate() }).
			WithCategory("Git").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("git", "log").
			WithDescription("Show git log").
			WithFunc(func() error { return g.Log() }).
			WithCategory("Git").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("git", "branch").
			WithDescription("Show git branches").
			WithFunc(func() error { return g.Branch() }).
			WithCategory("Git").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("git", "pull").
			WithDescription("Pull from remote").
			WithFunc(func() error { return g.Pull() }).
			WithCategory("Git").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("git", "commit").
			WithDescription("Create a git commit").
			WithArgsFunc(func(args ...string) error { return g.Commit(args...) }).
			WithCategory("Git").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("git", "init").
			WithDescription("Initialize git repository").
			WithFunc(func() error { return g.Init() }).
			WithCategory("Git").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("git", "add").
			WithDescription("Add files to git").
			WithArgsFunc(func(args ...string) error { return g.Add(args...) }).
			WithCategory("Git").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("git", "clone").
			WithDescription("Clone a repository").
			WithFunc(func() error { return g.Clone() }).
			WithCategory("Git").
			MustBuild(),
	)
}

func registerReleaseCommands(reg *registry.Registry) {
	r := mage.Release{}

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "default").
			WithDescription("Create a release").
			WithFunc(func() error { return r.Default() }).
			WithCategory("Release").
			WithAliases("release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "test").
			WithDescription("Test release process without publishing").
			WithFunc(func() error { return r.Test() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "snapshot").
			WithDescription("Create a snapshot release").
			WithFunc(func() error { return r.Snapshot() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "install").
			WithDescription("Install GoReleaser").
			WithFunc(func() error { return r.Install() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "update").
			WithDescription("Update GoReleaser").
			WithFunc(func() error { return r.Update() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "check").
			WithDescription("Check release configuration").
			WithFunc(func() error { return r.Check() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "init").
			WithDescription("Initialize release configuration").
			WithFunc(func() error { return r.Init() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "changelog").
			WithDescription("Generate changelog").
			WithFunc(func() error { return r.Changelog() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "create").
			WithDescription("Create a new release").
			WithFunc(func() error { return r.Create() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "prepare").
			WithDescription("Prepare release").
			WithFunc(func() error { return r.Prepare() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "publish").
			WithDescription("Publish release").
			WithFunc(func() error { return r.Publish() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "notes").
			WithDescription("Generate release notes").
			WithFunc(func() error { return r.Notes() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "validate").
			WithDescription("Validate release").
			WithFunc(func() error { return r.Validate() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "clean").
			WithDescription("Clean release artifacts").
			WithFunc(func() error { return r.Clean() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "build").
			WithDescription("Build release artifacts").
			WithFunc(func() error { return r.Build() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "package").
			WithDescription("Package release").
			WithFunc(func() error { return r.Package() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "draft").
			WithDescription("Create draft release").
			WithFunc(func() error { return r.Draft() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "alpha").
			WithDescription("Create alpha release").
			WithFunc(func() error { return r.Alpha() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "beta").
			WithDescription("Create beta release").
			WithFunc(func() error { return r.Beta() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "rc").
			WithDescription("Create release candidate").
			WithFunc(func() error { return r.RC() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "final").
			WithDescription("Create final release").
			WithFunc(func() error { return r.Final() }).
			WithCategory("Release").
			MustBuild(),
	)
}

func registerDocsCommands(reg *registry.Registry) {
	d := mage.Docs{}

	reg.MustRegister(
		registry.NewNamespaceCommand("docs", "default").
			WithDescription("Generate and serve documentation").
			WithFunc(func() error { return d.Default() }).
			WithCategory("Documentation").
			WithAliases("docs").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("docs", "build").
			WithDescription("Build documentation").
			WithFunc(func() error { return d.Build() }).
			WithCategory("Documentation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("docs", "generate").
			WithDescription("Generate documentation from code").
			WithFunc(func() error { return d.Generate() }).
			WithCategory("Documentation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("docs", "serve").
			WithDescription("Serve documentation locally").
			WithFunc(func() error { return d.Serve() }).
			WithCategory("Documentation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("docs", "check").
			WithDescription("Check documentation quality").
			WithFunc(func() error { return d.Check() }).
			WithCategory("Documentation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("docs", "godocs").
			WithDescription("Generate godocs").
			WithFunc(func() error { return d.GoDocs() }).
			WithCategory("Documentation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("docs", "examples").
			WithDescription("Generate example documentation").
			WithFunc(func() error { return d.Examples() }).
			WithCategory("Documentation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("docs", "readme").
			WithDescription("Generate README documentation").
			WithFunc(func() error { return d.Readme() }).
			WithCategory("Documentation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("docs", "api").
			WithDescription("Generate API documentation").
			WithFunc(func() error { return d.API() }).
			WithCategory("Documentation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("docs", "clean").
			WithDescription("Clean documentation artifacts").
			WithFunc(func() error { return d.Clean() }).
			WithCategory("Documentation").
			MustBuild(),
	)
}

func registerToolsCommands(reg *registry.Registry) {
	t := mage.Tools{}

	reg.MustRegister(
		registry.NewNamespaceCommand("tools", "install").
			WithDescription("Install development tools").
			WithFunc(func() error { return t.Install() }).
			WithCategory("Tools").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("tools", "update").
			WithDescription("Update development tools").
			WithFunc(func() error { return t.Update() }).
			WithCategory("Tools").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("tools", "verify").
			WithDescription("Verify installed tools").
			WithFunc(func() error { return t.Verify() }).
			WithCategory("Tools").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("tools", "clean").
			WithDescription("Clean tool installations").
			WithFunc(func() error { return t.Clean() }).
			WithCategory("Tools").
			MustBuild(),
	)
}

func registerGenerateCommands(reg *registry.Registry) {
	g := mage.Generate{}

	reg.MustRegister(
		registry.NewNamespaceCommand("generate", "default").
			WithDescription("Run code generation").
			WithFunc(func() error { return g.Default() }).
			WithCategory("Generate").
			WithAliases("generate").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("generate", "all").
			WithDescription("Generate all code").
			WithFunc(func() error { return g.All() }).
			WithCategory("Generate").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("generate", "mocks").
			WithDescription("Generate mock files").
			WithFunc(func() error { return g.Mocks() }).
			WithCategory("Generate").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("generate", "proto").
			WithDescription("Generate from protobuf files").
			WithFunc(func() error { return g.Proto() }).
			WithCategory("Generate").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("generate", "clean").
			WithDescription("Clean generated files").
			WithFunc(func() error { return g.Clean() }).
			WithCategory("Generate").
			MustBuild(),
	)
}

func registerCLICommands(reg *registry.Registry) {
	c := mage.CLI{}

	reg.MustRegister(
		registry.NewNamespaceCommand("cli", "default").
			WithDescription("Default CLI operations").
			WithFunc(func() error { return c.Default() }).
			WithCategory("CLI").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("cli", "bulk").
			WithDescription("Execute commands across multiple repositories").
			WithFunc(func() error { return c.Bulk() }).
			WithCategory("CLI").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("cli", "query").
			WithDescription("Query project information and metadata").
			WithFunc(func() error { return c.Query() }).
			WithCategory("CLI").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("cli", "dashboard").
			WithDescription("Display project dashboard and status").
			WithFunc(func() error { return c.Dashboard() }).
			WithCategory("CLI").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("cli", "batch").
			WithDescription("Execute batch operations from file").
			WithFunc(func() error { return c.Batch() }).
			WithCategory("CLI").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("cli", "monitor").
			WithDescription("Monitor build and test execution").
			WithFunc(func() error { return c.Monitor() }).
			WithCategory("CLI").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("cli", "workspace").
			WithDescription("Manage workspace configuration").
			WithFunc(func() error { return c.Workspace() }).
			WithCategory("CLI").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("cli", "pipeline").
			WithDescription("Execute pipeline operations").
			WithFunc(func() error { return c.Pipeline() }).
			WithCategory("CLI").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("cli", "compliance").
			WithDescription("Run compliance checks and reports").
			WithFunc(func() error { return c.Compliance() }).
			WithCategory("CLI").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("cli", "help").
			WithDescription("Display help for CLI commands").
			WithFunc(func() error { return c.Help() }).
			WithCategory("CLI").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("cli", "version").
			WithDescription("Display CLI version information").
			WithFunc(func() error { return c.Version() }).
			WithCategory("CLI").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("cli", "completion").
			WithDescription("Generate shell completion scripts").
			WithFunc(func() error { return c.Completion() }).
			WithCategory("CLI").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("cli", "config").
			WithDescription("Manage CLI configuration").
			WithFunc(func() error { return c.Config() }).
			WithCategory("CLI").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("cli", "update").
			WithDescription("Update CLI to latest version").
			WithFunc(func() error { return c.Update() }).
			WithCategory("CLI").
			MustBuild(),
	)
}

func registerUpdateCommands(reg *registry.Registry) {
	u := mage.Update{}

	reg.MustRegister(
		registry.NewNamespaceCommand("update", "check").
			WithDescription("Check for updates").
			WithFunc(func() error { return u.Check() }).
			WithCategory("Update").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("update", "install").
			WithDescription("Install updates").
			WithFunc(func() error { return u.Install() }).
			WithCategory("Update").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("update", "auto").
			WithDescription("Enable automatic updates").
			WithFunc(func() error { return u.Auto() }).
			WithCategory("Update").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("update", "history").
			WithDescription("Show update history").
			WithFunc(func() error { return u.History() }).
			WithCategory("Update").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("update", "rollback").
			WithDescription("Rollback to previous version").
			WithFunc(func() error { return u.Rollback() }).
			WithCategory("Update").
			MustBuild(),
	)
}

func registerModCommands(reg *registry.Registry) {
	m := mage.Mod{}

	reg.MustRegister(
		registry.NewNamespaceCommand("mod", "tidy").
			WithDescription("Tidy go.mod").
			WithFunc(func() error { return m.Tidy() }).
			WithCategory("Module").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("mod", "download").
			WithDescription("Download module dependencies").
			WithFunc(func() error { return m.Download() }).
			WithCategory("Module").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("mod", "update").
			WithDescription("Update module dependencies").
			WithFunc(func() error { return m.Update() }).
			WithCategory("Module").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("mod", "clean").
			WithDescription("Clean module cache").
			WithFunc(func() error { return m.Clean() }).
			WithCategory("Module").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("mod", "graph").
			WithDescription("Show module dependency graph").
			WithFunc(func() error { return m.Graph() }).
			WithCategory("Module").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("mod", "why").
			WithDescription("Explain why packages are needed").
			WithFunc(func() error { return m.Why() }).
			WithCategory("Module").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("mod", "vendor").
			WithDescription("Create vendor directory").
			WithFunc(func() error { return m.Vendor() }).
			WithCategory("Module").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("mod", "init").
			WithDescription("Initialize go.mod").
			WithFunc(func() error { return m.Init() }).
			WithCategory("Module").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("mod", "verify").
			WithDescription("Verify module dependencies").
			WithFunc(func() error { return m.Verify() }).
			WithCategory("Module").
			MustBuild(),
	)
}

func registerRecipesCommands(reg *registry.Registry) {
	r := mage.Recipes{}

	reg.MustRegister(
		registry.NewNamespaceCommand("recipes", "default").
			WithDescription("Show recipes menu").
			WithFunc(func() error { return r.Default() }).
			WithCategory("Recipes").
			WithAliases("recipes").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("recipes", "list").
			WithDescription("List available recipes").
			WithFunc(func() error { return r.List() }).
			WithCategory("Recipes").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("recipes", "show").
			WithDescription("Show recipe details").
			WithFunc(func() error { return r.Show() }).
			WithCategory("Recipes").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("recipes", "run").
			WithDescription("Run a recipe").
			WithFunc(func() error { return r.Run() }).
			WithCategory("Recipes").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("recipes", "search").
			WithDescription("Search recipes").
			WithFunc(func() error { return r.Search() }).
			WithCategory("Recipes").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("recipes", "create").
			WithDescription("Create a new recipe").
			WithFunc(func() error { return r.Create() }).
			WithCategory("Recipes").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("recipes", "install").
			WithDescription("Install a recipe").
			WithFunc(func() error { return r.Install() }).
			WithCategory("Recipes").
			MustBuild(),
	)
}

func registerMetricsCommands(reg *registry.Registry) {
	m := mage.Metrics{}

	reg.MustRegister(
		registry.NewNamespaceCommand("metrics", "loc").
			WithDescription("Count lines of code").
			WithFunc(func() error { return m.LOC() }).
			WithCategory("Metrics").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("metrics", "coverage").
			WithDescription("Calculate test coverage metrics").
			WithFunc(func() error { return m.Coverage() }).
			WithCategory("Metrics").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("metrics", "complexity").
			WithDescription("Analyze code complexity").
			WithFunc(func() error { return m.Complexity() }).
			WithCategory("Metrics").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("metrics", "size").
			WithDescription("Calculate binary size metrics").
			WithFunc(func() error { return m.Size() }).
			WithCategory("Metrics").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("metrics", "quality").
			WithDescription("Generate quality metrics report").
			WithFunc(func() error { return m.Quality() }).
			WithCategory("Metrics").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("metrics", "imports").
			WithDescription("Analyze import dependencies").
			WithFunc(func() error { return m.Imports() }).
			WithCategory("Metrics").
			MustBuild(),
	)
}

func registerWorkflowCommands(reg *registry.Registry) {
	w := mage.Workflow{}

	reg.MustRegister(
		registry.NewNamespaceCommand("workflow", "execute").
			WithDescription("Execute a workflow").
			WithFunc(func() error { return w.Execute() }).
			WithCategory("Workflow").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("workflow", "list").
			WithDescription("List workflows").
			WithFunc(func() error { return w.List() }).
			WithCategory("Workflow").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("workflow", "status").
			WithDescription("Show workflow status").
			WithFunc(func() error { return w.Status() }).
			WithCategory("Workflow").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("workflow", "create").
			WithDescription("Create a new workflow").
			WithFunc(func() error { return w.Create() }).
			WithCategory("Workflow").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("workflow", "validate").
			WithDescription("Validate workflow configuration").
			WithFunc(func() error { return w.Validate() }).
			WithCategory("Workflow").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("workflow", "schedule").
			WithDescription("Schedule workflow execution").
			WithFunc(func() error { return w.Schedule() }).
			WithCategory("Workflow").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("workflow", "template").
			WithDescription("Create workflow from template").
			WithFunc(func() error { return w.Template() }).
			WithCategory("Workflow").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("workflow", "history").
			WithDescription("Show workflow execution history").
			WithFunc(func() error { return w.History() }).
			WithCategory("Workflow").
			MustBuild(),
	)
}

func registerBenchCommands(reg *registry.Registry) {
	b := mage.Bench{}

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "default").
			WithDescription("Run benchmarks").
			WithFunc(func() error { return b.Default() }).
			WithCategory("Benchmark").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "compare").
			WithDescription("Compare benchmark results").
			WithFunc(func() error { return b.Compare() }).
			WithCategory("Benchmark").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "save").
			WithDescription("Save benchmark results").
			WithFunc(func() error { return b.Save() }).
			WithCategory("Benchmark").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "cpu").
			WithDescription("Run CPU benchmarks").
			WithFunc(func() error { return b.CPU() }).
			WithCategory("Benchmark").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "mem").
			WithDescription("Run memory benchmarks").
			WithFunc(func() error { return b.Mem() }).
			WithCategory("Benchmark").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "profile").
			WithDescription("Generate benchmark profiles").
			WithFunc(func() error { return b.Profile() }).
			WithCategory("Benchmark").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "trace").
			WithDescription("Generate execution traces").
			WithFunc(func() error { return b.Trace() }).
			WithCategory("Benchmark").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "regression").
			WithDescription("Run regression benchmarks").
			WithFunc(func() error { return b.Regression() }).
			WithCategory("Benchmark").
			MustBuild(),
	)
}

func registerVetCommands(reg *registry.Registry) {
	v := mage.Vet{}

	reg.MustRegister(
		registry.NewNamespaceCommand("vet", "default").
			WithDescription("Run go vet").
			WithFunc(func() error { return v.Default() }).
			WithCategory("Vet").
			WithAliases("vet").
			MustBuild(),
	)

	// Add more vet commands...
}

func registerConfigureCommands(reg *registry.Registry) {
	c := mage.Configure{}

	reg.MustRegister(
		registry.NewNamespaceCommand("configure", "init").
			WithDescription("Initialize configuration").
			WithFunc(func() error { return c.Init() }).
			WithCategory("Configure").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("configure", "show").
			WithDescription("Show current configuration").
			WithFunc(func() error { return c.Show() }).
			WithCategory("Configure").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("configure", "update").
			WithDescription("Update configuration").
			WithFunc(func() error { return c.Update() }).
			WithCategory("Configure").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("configure", "enterprise").
			WithDescription("Configure enterprise features").
			WithFunc(func() error { return c.Enterprise() }).
			WithCategory("Configure").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("configure", "export").
			WithDescription("Export configuration").
			WithFunc(func() error { return c.Export() }).
			WithCategory("Configure").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("configure", "import").
			WithDescription("Import configuration").
			WithFunc(func() error { return c.Import() }).
			WithCategory("Configure").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("configure", "validate").
			WithDescription("Validate configuration").
			WithFunc(func() error { return c.Validate() }).
			WithCategory("Configure").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("configure", "schema").
			WithDescription("Show configuration schema").
			WithFunc(func() error { return c.Schema() }).
			WithCategory("Configure").
			MustBuild(),
	)
}

func registerInitCommands(reg *registry.Registry) {
	i := mage.Init{}

	reg.MustRegister(
		registry.NewNamespaceCommand("init", "default").
			WithDescription("Initialize new project with smart defaults").
			WithFunc(func() error { return i.Default() }).
			WithCategory("Init").
			WithAliases("init").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("init", "project").
			WithDescription("Initialize new project").
			WithFunc(func() error { return i.Project() }).
			WithCategory("Init").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("init", "library").
			WithDescription("Initialize library project").
			WithFunc(func() error { return i.Library() }).
			WithCategory("Init").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("init", "cli").
			WithDescription("Initialize CLI application").
			WithFunc(func() error { return i.CLI() }).
			WithCategory("Init").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("init", "webapi").
			WithDescription("Initialize web API project").
			WithFunc(func() error { return i.WebAPI() }).
			WithCategory("Init").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("init", "microservice").
			WithDescription("Initialize microservice project").
			WithFunc(func() error { return i.Microservice() }).
			WithCategory("Init").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("init", "config").
			WithDescription("Initialize configuration files").
			WithFunc(func() error { return i.Config() }).
			WithCategory("Init").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("init", "git").
			WithDescription("Initialize git repository").
			WithFunc(func() error { return i.Git() }).
			WithCategory("Init").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("init", "ci").
			WithDescription("Initialize CI/CD configuration").
			WithFunc(func() error { return i.CI() }).
			WithCategory("Init").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("init", "docker").
			WithDescription("Initialize Docker configuration").
			WithFunc(func() error { return i.Docker() }).
			WithCategory("Init").
			MustBuild(),
	)
}

func registerEnterpriseCommands(reg *registry.Registry) {
	e := mage.Enterprise{}

	reg.MustRegister(
		registry.NewNamespaceCommand("enterprise", "init").
			WithDescription("Initialize enterprise features").
			WithFunc(func() error { return e.Init() }).
			WithCategory("Enterprise").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("enterprise", "config").
			WithDescription("Configure enterprise settings").
			WithFunc(func() error { return e.Config() }).
			WithCategory("Enterprise").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("enterprise", "deploy").
			WithDescription("Enterprise deployment").
			WithFunc(func() error { return e.Deploy() }).
			WithCategory("Enterprise").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("enterprise", "rollback").
			WithDescription("Rollback enterprise deployment").
			WithFunc(func() error { return e.Rollback() }).
			WithCategory("Enterprise").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("enterprise", "promote").
			WithDescription("Promote between environments").
			WithFunc(func() error { return e.Promote() }).
			WithCategory("Enterprise").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("enterprise", "status").
			WithDescription("Show enterprise deployment status").
			WithFunc(func() error { return e.Status() }).
			WithCategory("Enterprise").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("enterprise", "backup").
			WithDescription("Backup enterprise data").
			WithFunc(func() error { return e.Backup() }).
			WithCategory("Enterprise").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("enterprise", "restore").
			WithDescription("Restore enterprise data").
			WithFunc(func() error { return e.Restore() }).
			WithCategory("Enterprise").
			MustBuild(),
	)
}

func registerIntegrationsCommands(reg *registry.Registry) {
	i := mage.Integrations{}

	reg.MustRegister(
		registry.NewNamespaceCommand("integrations", "setup").
			WithDescription("Setup integrations").
			WithFunc(func() error { return i.Setup() }).
			WithCategory("Integrations").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("integrations", "test").
			WithDescription("Test integrations").
			WithFunc(func() error { return i.Test() }).
			WithCategory("Integrations").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("integrations", "sync").
			WithDescription("Sync integration data").
			WithFunc(func() error { return i.Sync() }).
			WithCategory("Integrations").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("integrations", "notify").
			WithDescription("Send notifications").
			WithFunc(func() error { return i.Notify() }).
			WithCategory("Integrations").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("integrations", "status").
			WithDescription("Show integration status").
			WithFunc(func() error { return i.Status() }).
			WithCategory("Integrations").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("integrations", "webhook").
			WithDescription("Configure webhooks").
			WithFunc(func() error { return i.Webhook() }).
			WithCategory("Integrations").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("integrations", "export").
			WithDescription("Export integration configuration").
			WithFunc(func() error { return i.Export() }).
			WithCategory("Integrations").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("integrations", "import").
			WithDescription("Import integration configuration").
			WithFunc(func() error { return i.Import() }).
			WithCategory("Integrations").
			MustBuild(),
	)
}

func registerWizardCommands(reg *registry.Registry) {
	w := mage.Wizard{}

	reg.MustRegister(
		registry.NewNamespaceCommand("wizard", "setup").
			WithDescription("Interactive setup wizard").
			WithFunc(func() error { return w.Setup() }).
			WithCategory("Wizard").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("wizard", "project").
			WithDescription("Project configuration wizard").
			WithFunc(func() error { return w.Project() }).
			WithCategory("Wizard").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("wizard", "integration").
			WithDescription("Integration setup wizard").
			WithFunc(func() error { return w.Integration() }).
			WithCategory("Wizard").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("wizard", "security").
			WithDescription("Security configuration wizard").
			WithFunc(func() error { return w.Security() }).
			WithCategory("Wizard").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("wizard", "workflow").
			WithDescription("Workflow setup wizard").
			WithFunc(func() error { return w.Workflow() }).
			WithCategory("Wizard").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("wizard", "deployment").
			WithDescription("Deployment configuration wizard").
			WithFunc(func() error { return w.Deployment() }).
			WithCategory("Wizard").
			MustBuild(),
	)
}

func registerHelpCommands(reg *registry.Registry) {
	h := mage.Help{}

	reg.MustRegister(
		registry.NewNamespaceCommand("help", "default").
			WithDescription("Show help").
			WithFunc(func() error { return h.Default() }).
			WithCategory("Help").
			WithAliases("help").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("help", "commands").
			WithDescription("List all available commands").
			WithFunc(func() error { return h.Commands() }).
			WithCategory("Help").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("help", "command").
			WithDescription("Show help for a specific command").
			WithFunc(func() error { return h.Command() }).
			WithCategory("Help").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("help", "examples").
			WithDescription("Show usage examples").
			WithFunc(func() error { return h.Examples() }).
			WithCategory("Help").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("help", "gettingstarted").
			WithDescription("Getting started guide").
			WithFunc(func() error { return h.GettingStarted() }).
			WithCategory("Help").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("help", "completions").
			WithDescription("Setup shell completions").
			WithFunc(func() error { return h.Completions() }).
			WithCategory("Help").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("help", "topics").
			WithDescription("List help topics").
			WithFunc(func() error { return h.Topics() }).
			WithCategory("Help").
			MustBuild(),
	)
}

// registerVersionCommands registers all Version namespace commands
func registerVersionCommands(reg *registry.Registry) {
	var v mage.Version

	reg.MustRegister(
		registry.NewNamespaceCommand("version", "show").
			WithDescription("Display current version information").
			WithFunc(func() error { return v.Show() }).
			WithCategory("Version Management").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("version", "check").
			WithDescription("Check version information and compare with latest").
			WithFunc(func() error { return v.Check() }).
			WithCategory("Version Management").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("version", "update").
			WithDescription("Update to latest version").
			WithFunc(func() error { return v.Update() }).
			WithCategory("Version Management").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("version", "bump").
			WithDescription("Bump version (patch, minor, major)").
			WithFunc(func() error { return v.Bump() }).
			WithCategory("Version Management").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("version", "changelog").
			WithDescription("Generate changelog from git history").
			WithFunc(func() error { return v.Changelog() }).
			WithCategory("Version Management").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("version", "tag").
			WithDescription("Create version tag").
			WithFunc(func() error { return v.Tag() }).
			WithCategory("Version Management").
			MustBuild(),
	)
}

// registerInstallCommands registers all Install namespace commands
func registerInstallCommands(reg *registry.Registry) {
	var i mage.Install

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "default").
			WithDescription("Default installation").
			WithFunc(func() error { return i.Default() }).
			WithCategory("Installation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "local").
			WithDescription("Install locally").
			WithFunc(func() error { return i.Local() }).
			WithCategory("Installation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "binary").
			WithDescription("Install project binary").
			WithFunc(func() error { return i.Binary() }).
			WithCategory("Installation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "tools").
			WithDescription("Install development tools").
			WithFunc(func() error { return i.Tools() }).
			WithCategory("Installation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "go").
			WithDescription("Install Go").
			WithFunc(func() error { return i.Go() }).
			WithCategory("Installation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "stdlib").
			WithDescription("Install Go standard library").
			WithFunc(func() error { return i.Stdlib() }).
			WithCategory("Installation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "systemwide").
			WithDescription("Install system-wide").
			WithFunc(func() error { return i.SystemWide() }).
			WithCategory("Installation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "deps").
			WithDescription("Install dependencies").
			WithFunc(func() error { return i.Deps() }).
			WithCategory("Installation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "mage").
			WithDescription("Install mage").
			WithFunc(func() error { return i.Mage() }).
			WithCategory("Installation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "docker").
			WithDescription("Install Docker components").
			WithFunc(func() error { return i.Docker() }).
			WithCategory("Installation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "githooks").
			WithDescription("Install git hooks").
			WithFunc(func() error { return i.GitHooks() }).
			WithCategory("Installation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "ci").
			WithDescription("Install CI components").
			WithFunc(func() error { return i.CI() }).
			WithCategory("Installation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "certs").
			WithDescription("Install certificates").
			WithFunc(func() error { return i.Certs() }).
			WithCategory("Installation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "package").
			WithDescription("Install package").
			WithFunc(func() error { return i.Package() }).
			WithCategory("Installation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "all").
			WithDescription("Install everything").
			WithFunc(func() error { return i.All() }).
			WithCategory("Installation").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("install", "uninstall").
			WithDescription("Remove installation").
			WithFunc(func() error { return i.Uninstall() }).
			WithCategory("Installation").
			MustBuild(),
	)
}

// registerAuditCommands registers all Audit namespace commands
func registerAuditCommands(reg *registry.Registry) {
	var a mage.Audit

	reg.MustRegister(
		registry.NewNamespaceCommand("audit", "show").
			WithDescription("Display audit events with optional filtering").
			WithFunc(func() error { return a.Show() }).
			WithCategory("Audit & Security").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("audit", "stats").
			WithDescription("Show audit statistics and summaries").
			WithFunc(func() error { return a.Stats() }).
			WithCategory("Audit & Security").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("audit", "export").
			WithDescription("Export audit data to various formats").
			WithFunc(func() error { return a.Export() }).
			WithCategory("Audit & Security").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("audit", "cleanup").
			WithDescription("Clean up old audit entries").
			WithFunc(func() error { return a.Cleanup() }).
			WithCategory("Audit & Security").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("audit", "enable").
			WithDescription("Enable audit logging").
			WithFunc(func() error { return a.Enable() }).
			WithCategory("Audit & Security").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("audit", "disable").
			WithDescription("Disable audit logging").
			WithFunc(func() error { return a.Disable() }).
			WithCategory("Audit & Security").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("audit", "report").
			WithDescription("Generate comprehensive audit reports").
			WithFunc(func() error { return a.Report() }).
			WithCategory("Audit & Security").
			MustBuild(),
	)
}

// registerYamlCommands registers all Yaml namespace commands
func registerYamlCommands(reg *registry.Registry) {
	var y mage.Yaml

	reg.MustRegister(
		registry.NewNamespaceCommand("yaml", "init").
			WithDescription("Create mage.yaml configuration").
			WithFunc(func() error { return y.Init() }).
			WithCategory("Configuration").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("yaml", "validate").
			WithDescription("Validate YAML configuration").
			WithFunc(func() error { return y.Validate() }).
			WithCategory("Configuration").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("yaml", "show").
			WithDescription("Show current YAML configuration").
			WithFunc(func() error { return y.Show() }).
			WithCategory("Configuration").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("yaml", "update").
			WithDescription("Update YAML configuration").
			WithFunc(func() error { return y.Update() }).
			WithCategory("Configuration").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("yaml", "template").
			WithDescription("Generate YAML templates").
			WithFunc(func() error { return y.Template() }).
			WithCategory("Configuration").
			MustBuild(),
	)
}

// registerReleasesCommands registers all Releases namespace commands
func registerReleasesCommands(reg *registry.Registry) {
	var r mage.Releases

	reg.MustRegister(
		registry.NewNamespaceCommand("releases", "create").
			WithDescription("Create a new release").
			WithFunc(func() error { return r.Create() }).
			WithCategory("Release Management").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("releases", "publish").
			WithDescription("Publish a release").
			WithFunc(func() error { return r.Publish() }).
			WithCategory("Release Management").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("releases", "stable").
			WithDescription("Create stable releases").
			WithFunc(func() error { return r.Stable() }).
			WithCategory("Release Management").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("releases", "beta").
			WithDescription("Create beta releases").
			WithFunc(func() error { return r.Beta() }).
			WithCategory("Release Management").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("releases", "edge").
			WithDescription("Create edge releases").
			WithFunc(func() error { return r.Edge() }).
			WithCategory("Release Management").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("releases", "draft").
			WithDescription("Create draft releases").
			WithFunc(func() error { return r.Draft() }).
			WithCategory("Release Management").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("releases", "promote").
			WithDescription("Promote a release").
			WithFunc(func() error { return r.Promote() }).
			WithCategory("Release Management").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("releases", "status").
			WithDescription("Show release status").
			WithFunc(func() error { return r.Status() }).
			WithCategory("Release Management").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("releases", "channels").
			WithDescription("List available release channels").
			WithFunc(func() error { return r.Channels() }).
			WithCategory("Release Management").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("releases", "cleanup").
			WithDescription("Clean up old releases").
			WithFunc(func() error { return r.Cleanup() }).
			WithCategory("Release Management").
			MustBuild(),
	)
}

// registerEnterpriseConfigCommands registers all EnterpriseConfig namespace commands
func registerEnterpriseConfigCommands(reg *registry.Registry) {
	var ec mage.EnterpriseConfigNamespace

	reg.MustRegister(
		registry.NewNamespaceCommand("enterpriseconfig", "init").
			WithDescription("Initialize enterprise configuration").
			WithFunc(func() error { return ec.Init() }).
			WithCategory("Enterprise").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("enterpriseconfig", "validate").
			WithDescription("Validate enterprise configuration").
			WithFunc(func() error { return ec.Validate() }).
			WithCategory("Enterprise").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("enterpriseconfig", "update").
			WithDescription("Update enterprise configuration").
			WithFunc(func() error { return ec.Update() }).
			WithCategory("Enterprise").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("enterpriseconfig", "export").
			WithDescription("Export enterprise configuration").
			WithFunc(func() error { return ec.Export() }).
			WithCategory("Enterprise").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("enterpriseconfig", "import").
			WithDescription("Import enterprise configuration").
			WithFunc(func() error { return ec.Import() }).
			WithCategory("Enterprise").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("enterpriseconfig", "schema").
			WithDescription("Show enterprise configuration schema").
			WithFunc(func() error { return ec.Schema() }).
			WithCategory("Enterprise").
			MustBuild(),
	)
}
