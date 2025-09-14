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

	// Register Update namespace commands
	registerUpdateCommands(reg)

	// Register Mod namespace commands
	registerModCommands(reg)

	// Register Metrics namespace commands
	registerMetricsCommands(reg)

	// Register Bench namespace commands
	registerBenchCommands(reg)

	// Register Vet namespace commands
	registerVetCommands(reg)

	// Register Configure namespace commands
	registerConfigureCommands(reg)

	// Register Init namespace commands
	registerInitCommands(reg)

	// Register Help namespace commands
	registerHelpCommands(reg)

	// Register Version namespace commands
	registerVersionCommands(reg)

	// Register Install namespace commands
	registerInstallCommands(reg)

	// Register Yaml namespace commands
	registerYamlCommands(reg)

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
		registry.NewNamespaceCommand("build", "dev").
			WithDescription("Build and install development version (forced 'dev' version)").
			WithFunc(func() error { return b.Dev() }).
			WithCategory("Build").
			WithAliases("dev").
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
			WithArgsFunc(func(args ...string) error { return t.Default(args...) }).
			WithCategory("Test").
			WithAliases("test").
			WithUsage("magex test [flags]").
			WithExamples("magex test", "magex test -json", "magex test -race -v").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "unit").
			WithDescription("Run unit tests only").
			WithArgsFunc(func(args ...string) error { return t.Unit(args...) }).
			WithCategory("Test").
			WithUsage("magex test:unit [flags]").
			WithExamples("magex test:unit", "magex test:unit -json", "magex test:unit -v").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "full").
			WithDescription("Run full test suite with linting").
			WithArgsFunc(func(args ...string) error { return t.Full(args...) }).
			WithCategory("Test").
			WithUsage("magex test:full [flags]").
			WithExamples("magex test:full", "magex test:full -json", "magex test:full -v").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "short").
			WithDescription("Run short tests").
			WithArgsFunc(func(args ...string) error { return t.Short(args...) }).
			WithCategory("Test").
			WithUsage("magex test:short [flags]").
			WithExamples("magex test:short", "magex test:short -json").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "race").
			WithDescription("Run tests with race detector").
			WithArgsFunc(func(args ...string) error { return t.Race(args...) }).
			WithCategory("Test").
			WithUsage("magex test:race [flags]").
			WithExamples("magex test:race", "magex test:race -json", "magex test:race -v").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "cover").
			WithDescription("Run tests with coverage").
			WithArgsFunc(func(args ...string) error { return t.Cover(args...) }).
			WithCategory("Test").
			WithUsage("magex test:cover [flags]").
			WithExamples("magex test:cover", "magex test:cover -json", "magex test:cover -v").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "coverrace").
			WithDescription("Run tests with coverage and race detector").
			WithArgsFunc(func(args ...string) error { return t.CoverRace(args...) }).
			WithCategory("Test").
			WithUsage("magex test:coverrace [flags]").
			WithExamples("magex test:coverrace", "magex test:coverrace -json", "magex test:coverrace -v").
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
			WithDescription("Run fuzz tests with optional time parameter").
			WithArgsFunc(func(args ...string) error { return t.Fuzz(args...) }).
			WithCategory("Test").
			WithUsage("magex test:fuzz [time=<duration>]").
			WithExamples("magex test:fuzz", "magex test:fuzz time=5s").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("test", "fuzzshort").
			WithDescription("Run short fuzz tests with optional time parameter").
			WithArgsFunc(func(args ...string) error { return t.FuzzShort(args...) }).
			WithCategory("Test").
			WithUsage("magex test:fuzzshort [time=<duration>]").
			WithExamples("magex test:fuzzshort", "magex test:fuzzshort time=3s").
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
		registry.NewNamespaceCommand("test", "benchshort").
			WithDescription("Run short benchmarks with optional time parameter").
			WithArgsFunc(func(args ...string) error { return t.BenchShort(args...) }).
			WithCategory("Test").
			WithUsage("magex test:benchshort [time=<duration>]").
			WithExamples("magex test:benchshort", "magex test:benchshort time=1s").
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

	reg.MustRegister(
		registry.NewNamespaceCommand("deps", "audit").
			WithDescription("Security audit of dependencies").
			WithFunc(func() error { return d.Audit() }).
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

	// Bench is an alias for test:bench
	bench := mage.Test{}
	reg.MustRegister(
		registry.NewCommand("bench").
			WithDescription("Run benchmarks").
			WithArgsFunc(func(args ...string) error { return bench.Bench(args...) }).
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
			WithDescription("Create a git tag with version parameter").
			WithArgsFunc(func(args ...string) error { return g.TagWithArgs(args...) }).
			WithCategory("Git").
			WithUsage("magex git:tag version=<X.Y.Z>").
			WithExamples("magex git:tag version=1.2.3").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("git", "tagremove").
			WithDescription("Remove a git tag with version parameter").
			WithArgsFunc(func(args ...string) error { return g.TagRemoveWithArgs(args...) }).
			WithCategory("Git").
			WithUsage("magex git:tagremove version=<X.Y.Z>").
			WithExamples("magex git:tagremove version=1.2.3").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("git", "tagupdate").
			WithDescription("Update a git tag with version parameter").
			WithArgsFunc(func(args ...string) error { return g.TagUpdate(args...) }).
			WithCategory("Git").
			WithUsage("magex git:tagupdate version=<X.Y.Z>").
			WithExamples("magex git:tagupdate version=1.2.3").
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
			WithDescription("Create a git commit with message parameter").
			WithArgsFunc(func(args ...string) error { return g.Commit(args...) }).
			WithCategory("Git").
			WithUsage("magex git:commit message=\"<commit message>\"").
			WithExamples("magex git:commit message=\"fix: resolve bug in parser\"").
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
			WithArgsFunc(func(args ...string) error { return r.Default(args...) }).
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
		registry.NewNamespaceCommand("release", "validate").
			WithDescription("Comprehensive release readiness validation").
			WithFunc(func() error { return r.Validate() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "clean").
			WithDescription("Clean release artifacts and build cache").
			WithFunc(func() error { return r.Clean() }).
			WithCategory("Release").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("release", "localinstall").
			WithDescription("Build from latest tag and install locally").
			WithFunc(func() error { return r.LocalInstall() }).
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
			WithDescription("Visualize dependency graph with parameters").
			WithArgsFunc(func(args ...string) error { return m.Graph(args...) }).
			WithCategory("Module").
			WithUsage("magex mod:graph [depth=<n>] [format=<type>] [filter=<pattern>] [show_versions=<bool>]").
			WithExamples(
				"magex mod:graph",
				"magex mod:graph depth=3",
				"magex mod:graph format=json",
				"magex mod:graph filter=github.com show_versions=false",
			).
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("mod", "why").
			WithDescription("Explain why packages are needed").
			WithFunc(func() error { return m.Why() }).
			WithArgsFunc(func(args ...string) error { return m.Why(args...) }).
			WithCategory("Module").
			WithUsage("magex mod:why <module1> [module2] ...").
			WithExamples(
				"magex mod:why github.com/stretchr/testify",
				"magex mod:why github.com/pkg/errors golang.org/x/sync",
			).
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

func registerBenchCommands(reg *registry.Registry) {
	b := mage.Bench{}

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "default").
			WithDescription("Run benchmarks with optional parameters (time=duration)").
			WithFunc(func() error { return b.Default() }).
			WithArgsFunc(func(args ...string) error { return b.DefaultWithArgs(args...) }).
			WithCategory("Benchmark").
			WithExamples("magex bench", "magex bench time=50ms", "magex bench time=10s count=3").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "compare").
			WithDescription("Compare benchmark results with optional file parameters").
			WithFunc(func() error { return b.Compare() }).
			WithArgsFunc(func(args ...string) error { return b.CompareWithArgs(args...) }).
			WithCategory("Benchmark").
			WithExamples("magex bench:compare", "magex bench:compare old=baseline.txt new=current.txt").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "save").
			WithDescription("Save benchmark results with optional parameters (time=duration, output=file)").
			WithFunc(func() error { return b.Save() }).
			WithArgsFunc(func(args ...string) error { return b.SaveWithArgs(args...) }).
			WithCategory("Benchmark").
			WithExamples("magex bench:save", "magex bench:save time=1s output=results.txt").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "cpu").
			WithDescription("Run CPU benchmarks with optional parameters (time=duration, profile=file)").
			WithFunc(func() error { return b.CPU() }).
			WithArgsFunc(func(args ...string) error { return b.CPUWithArgs(args...) }).
			WithCategory("Benchmark").
			WithExamples("magex bench:cpu", "magex bench:cpu time=30s profile=cpu-profile.out").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "mem").
			WithDescription("Run memory benchmarks with optional parameters (time=duration, profile=file)").
			WithFunc(func() error { return b.Mem() }).
			WithArgsFunc(func(args ...string) error { return b.MemWithArgs(args...) }).
			WithCategory("Benchmark").
			WithExamples("magex bench:mem", "magex bench:mem time=2s profile=mem-profile.out").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "profile").
			WithDescription("Generate benchmark profiles with optional parameters (time=duration, cpu-profile=file, mem-profile=file)").
			WithFunc(func() error { return b.Profile() }).
			WithArgsFunc(func(args ...string) error { return b.ProfileWithArgs(args...) }).
			WithCategory("Benchmark").
			WithExamples("magex bench:profile", "magex bench:profile time=5s cpu-profile=cpu.prof mem-profile=mem.prof").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "trace").
			WithDescription("Generate execution traces with optional parameters (time=duration, trace=file)").
			WithFunc(func() error { return b.Trace() }).
			WithArgsFunc(func(args ...string) error { return b.TraceWithArgs(args...) }).
			WithCategory("Benchmark").
			WithExamples("magex bench:trace", "magex bench:trace time=10s trace=bench-trace.out").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("bench", "regression").
			WithDescription("Run regression benchmarks with optional parameters (time=duration, update-baseline=true)").
			WithFunc(func() error { return b.Regression() }).
			WithArgsFunc(func(args ...string) error { return b.RegressionWithArgs(args...) }).
			WithCategory("Benchmark").
			WithExamples("magex bench:regression", "magex bench:regression time=5s update-baseline=true").
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
			WithDescription("Bump version with parameters: bump=<major|minor|patch> push dry-run force major-confirm").
			WithArgsFunc(func(args ...string) error { return v.Bump(args...) }).
			WithCategory("Version Management").
			WithUsage("magex version:bump [bump=<type>] [push] [dry-run] [force] [major-confirm]").
			WithExamples(
				"magex version:bump bump=patch push",
				"magex version:bump bump=minor dry-run",
				"magex version:bump bump=major major-confirm push",
			).
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewNamespaceCommand("version", "changelog").
			WithDescription("Generate changelog from git history with parameters: from=<tag> to=<tag>").
			WithArgsFunc(func(args ...string) error { return v.Changelog(args...) }).
			WithCategory("Version Management").
			WithUsage("magex version:changelog [from=<tag>] [to=<tag>]").
			WithExamples(
				"magex version:changelog",
				"magex version:changelog from=v1.0.0",
				"magex version:changelog from=v1.0.0 to=v1.1.0",
			).
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
