// Package embed provides registration of all built-in MAGE-X commands
// This is the heart of the zero-boilerplate solution - all commands are pre-registered
package embed

import (
	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

// MethodBinding links method names to actual functions
type MethodBinding struct {
	NoArgs   func() error          // for WithFunc
	WithArgs func(...string) error // for WithArgsFunc
}

// CommandDef represents a declarative command definition
type CommandDef struct {
	Method   string   // e.g., "default", "all", "linux"
	Desc     string   // Short description
	Aliases  []string // Optional aliases
	Usage    string   // Optional usage pattern
	Examples []string // Optional examples
}

// registerNamespaceCommands registers all commands for a namespace using data definitions
func registerNamespaceCommands(
	reg *registry.Registry,
	namespace, category string,
	commands []CommandDef,
	bindings map[string]MethodBinding,
) {
	for _, cmd := range commands {
		binding := bindings[cmd.Method]
		builder := registry.NewNamespaceCommand(namespace, cmd.Method).
			WithDescription(cmd.Desc).
			WithCategory(category)

		if binding.NoArgs != nil {
			builder = builder.WithFunc(binding.NoArgs)
		}
		if binding.WithArgs != nil {
			builder = builder.WithArgsFunc(binding.WithArgs)
		}
		if len(cmd.Aliases) > 0 {
			builder = builder.WithAliases(cmd.Aliases...)
		}
		if cmd.Usage != "" {
			builder = builder.WithUsage(cmd.Usage)
		}
		if len(cmd.Examples) > 0 {
			builder = builder.WithExamples(cmd.Examples...)
		}
		reg.MustRegister(builder.MustBuild())
	}
}

// ============================================================================
// Command Definitions - Data Tables (as functions to avoid global variables)
// ============================================================================

func getBuildCommands() []CommandDef {
	return []CommandDef{
		{Method: "default", Desc: "Build the application for the current platform", Aliases: []string{"build"}},
		{Method: "all", Desc: "Build for all configured platforms"},
		{Method: "linux", Desc: "Build for Linux (amd64)"},
		{Method: "darwin", Desc: "Build for macOS (amd64 and arm64)", Aliases: []string{"build:mac", "build:macos"}},
		{Method: "windows", Desc: "Build for Windows (amd64)"},
		{Method: "clean", Desc: "Remove build artifacts", Aliases: []string{"clean"}},
		{Method: "install", Desc: "Install the binary to $GOPATH/bin", Aliases: []string{"install"}},
		{Method: "dev", Desc: "Build and install development version (forced 'dev' version)", Aliases: []string{"dev"}},
		{Method: "generate", Desc: "Run go generate"},
		{Method: "prebuild", Desc: "Pre-build all packages to warm cache"},
	}
}

func getTestCommands() []CommandDef {
	return []CommandDef{
		{Method: "default", Desc: "Run standard test suite", Aliases: []string{"test"}, Usage: "magex test [flags]", Examples: []string{"magex test", "magex test -json", "magex test -race -v"}},
		{Method: "unit", Desc: "Run unit tests only", Usage: "magex test:unit [flags]", Examples: []string{"magex test:unit", "magex test:unit -json", "magex test:unit -v"}},
		{Method: "full", Desc: "Run full test suite with linting", Usage: "magex test:full [flags]", Examples: []string{"magex test:full", "magex test:full -json", "magex test:full -v"}},
		{Method: "short", Desc: "Run short tests", Usage: "magex test:short [flags]", Examples: []string{"magex test:short", "magex test:short -json"}},
		{Method: "race", Desc: "Run tests with race detector", Usage: "magex test:race [flags]", Examples: []string{"magex test:race", "magex test:race -json", "magex test:race -v"}},
		{Method: "cover", Desc: "Run tests with coverage", Usage: "magex test:cover [flags]", Examples: []string{"magex test:cover", "magex test:cover -json", "magex test:cover -v"}},
		{Method: "coverrace", Desc: "Run tests with coverage and race detector", Usage: "magex test:coverrace [flags]", Examples: []string{"magex test:coverrace", "magex test:coverrace -json", "magex test:coverrace -v"}},
		{Method: "coverreport", Desc: "Generate coverage report"},
		{Method: "coverhtml", Desc: "Generate HTML coverage report"},
		{Method: "fuzz", Desc: "Run fuzz tests with optional time parameter", Usage: "magex test:fuzz [time=<duration>]", Examples: []string{"magex test:fuzz", "magex test:fuzz time=5s"}},
		{Method: "fuzzshort", Desc: "Run short fuzz tests with optional time parameter", Usage: "magex test:fuzzshort [time=<duration>]", Examples: []string{"magex test:fuzzshort", "magex test:fuzzshort time=3s"}},
		{Method: "bench", Desc: "Run benchmarks", Aliases: []string{"benchmark"}},
		{Method: "benchshort", Desc: "Run short benchmarks with optional time parameter", Usage: "magex test:benchshort [time=<duration>]", Examples: []string{"magex test:benchshort", "magex test:benchshort time=1s"}},
		{Method: "integration", Desc: "Run integration tests"},
		{Method: "ci", Desc: "Run CI test suite"},
		{Method: "parallel", Desc: "Run tests in parallel"},
		{Method: "nolint", Desc: "Run tests without linting"},
		{Method: "cinorace", Desc: "Run CI tests without race detector"},
		{Method: "run", Desc: "Run a specific test pattern"},
		{Method: "coverage", Desc: "Run tests and generate coverage"},
		{Method: "vet", Desc: "Run go vet"},
	}
}

func getLintCommands() []CommandDef {
	return []CommandDef{
		{Method: "default", Desc: "Run default linting", Aliases: []string{"lint"}},
		{Method: "fix", Desc: "Fix auto-fixable lint issues"},
		{Method: "ci", Desc: "Run CI linting (strict)"},
		{Method: "fast", Desc: "Run fast linting checks"},
		{Method: "issues", Desc: "Scan for TODOs, FIXMEs, nolint directives, and test skips"},
	}
}

func getFormatCommands() []CommandDef {
	return []CommandDef{
		{Method: "default", Desc: "Format Go code", Aliases: []string{"format", "fmt"}},
		{Method: "check", Desc: "Check if code is formatted"},
		{Method: "fix", Desc: "Fix formatting issues"},
		{Method: "imports", Desc: "Fix import statements"},
	}
}

func getDepsCommands() []CommandDef {
	return []CommandDef{
		{Method: "default", Desc: "Manage dependencies", Aliases: []string{"deps"}},
		{Method: "update", Desc: "Update all dependencies", Usage: "magex deps:update [all-modules] [dry-run] [fail-fast] [allow-major] [stable-only] [verbose]", Examples: []string{"magex deps:update", "magex deps:update all-modules", "magex deps:update all-modules dry-run", "magex deps:update all-modules fail-fast verbose", "magex deps:update all-modules allow-major stable-only verbose"}},
		{Method: "tidy", Desc: "Run go mod tidy", Aliases: []string{"tidy"}},
		{Method: "download", Desc: "Download dependencies"},
		{Method: "vendor", Desc: "Vendor dependencies", Aliases: []string{"vendor"}},
		{Method: "verify", Desc: "Verify dependencies"},
		{Method: "outdated", Desc: "List outdated dependencies"},
		{Method: "graph", Desc: "Show dependency graph"},
		{Method: "licenses", Desc: "List dependency licenses"},
	}
}

func getGitCommands() []CommandDef {
	return []CommandDef{
		{Method: "status", Desc: "Show git status"},
		{Method: "diff", Desc: "Show git diff and check for uncommitted changes"},
		{Method: "tag", Desc: "Create a git tag with version parameter", Usage: "magex git:tag version=<X.Y.Z>", Examples: []string{"magex git:tag version=1.2.3"}},
		{Method: "tagremove", Desc: "Remove a git tag with version parameter", Usage: "magex git:tagremove version=<X.Y.Z>", Examples: []string{"magex git:tagremove version=1.2.3"}},
		{Method: "tagupdate", Desc: "Update a git tag with version parameter", Usage: "magex git:tagupdate version=<X.Y.Z>", Examples: []string{"magex git:tagupdate version=1.2.3"}},
		{Method: "log", Desc: "Show git log"},
		{Method: "branch", Desc: "Show git branches"},
		{Method: "pull", Desc: "Pull from remote"},
		{Method: "commit", Desc: "Create a git commit with message parameter", Usage: "magex git:commit message=\"<commit message>\"", Examples: []string{"magex git:commit message=\"fix: resolve bug in parser\""}},
		{Method: "init", Desc: "Initialize git repository"},
		{Method: "add", Desc: "Add files to git"},
		{Method: "clone", Desc: "Clone a repository"},
	}
}

func getReleaseCommands() []CommandDef {
	return []CommandDef{
		{Method: "default", Desc: "Create a release", Aliases: []string{"release"}},
		{Method: "test", Desc: "Test release process without publishing"},
		{Method: "snapshot", Desc: "Create a snapshot release"},
		{Method: "check", Desc: "Check release configuration"},
		{Method: "init", Desc: "Initialize release configuration"},
		{Method: "changelog", Desc: "Generate changelog"},
		{Method: "validate", Desc: "Comprehensive release readiness validation"},
		{Method: "clean", Desc: "Clean release artifacts and build cache"},
		{Method: "localinstall", Desc: "Build from latest tag and install locally"},
	}
}

func getDocsCommands() []CommandDef {
	return []CommandDef{
		{Method: "default", Desc: "Generate and serve documentation", Aliases: []string{"docs"}},
		{Method: "build", Desc: "Build documentation"},
		{Method: "generate", Desc: "Generate documentation from code"},
		{Method: "serve", Desc: "Serve documentation locally"},
		{Method: "check", Desc: "Check documentation quality"},
		{Method: "godocs", Desc: "Generate godocs"},
		{Method: "examples", Desc: "Generate example documentation"},
		{Method: "readme", Desc: "Generate README documentation"},
		{Method: "api", Desc: "Generate API documentation"},
		{Method: "clean", Desc: "Clean documentation artifacts"},
	}
}

func getToolsCommands() []CommandDef {
	return []CommandDef{
		{Method: "install", Desc: "Install development tools"},
		{Method: "update", Desc: "Update development tools"},
		{Method: "verify", Desc: "Verify installed tools"},
		{Method: "clean", Desc: "Clean tool installations"},
	}
}

func getGenerateCommands() []CommandDef {
	return []CommandDef{
		{Method: "default", Desc: "Run code generation", Aliases: []string{"generate"}},
		{Method: "all", Desc: "Generate all code"},
		{Method: "mocks", Desc: "Generate mock files"},
		{Method: "proto", Desc: "Generate from protobuf files"},
		{Method: "clean", Desc: "Clean generated files"},
	}
}

func getUpdateCommands() []CommandDef {
	return []CommandDef{
		{Method: "check", Desc: "Check for updates"},
		{Method: "install", Desc: "Install updates"},
	}
}

func getModCommands() []CommandDef {
	return []CommandDef{
		{Method: "tidy", Desc: "Tidy go.mod"},
		{Method: "download", Desc: "Download module dependencies"},
		{Method: "update", Desc: "Update module dependencies"},
		{Method: "clean", Desc: "Clean module cache"},
		{Method: "graph", Desc: "Visualize dependency graph with parameters", Usage: "magex mod:graph [depth=<n>] [format=<type>] [filter=<pattern>] [show_versions=<bool>]", Examples: []string{"magex mod:graph", "magex mod:graph depth=3", "magex mod:graph format=json", "magex mod:graph filter=github.com show_versions=false"}},
		{Method: "why", Desc: "Explain why packages are needed", Usage: "magex mod:why <module1> [module2] ...", Examples: []string{"magex mod:why github.com/stretchr/testify", "magex mod:why github.com/pkg/errors golang.org/x/sync"}},
		{Method: "vendor", Desc: "Create vendor directory"},
		{Method: "init", Desc: "Initialize go.mod"},
		{Method: "verify", Desc: "Verify module dependencies"},
	}
}

func getMetricsCommands() []CommandDef {
	return []CommandDef{
		{Method: "loc", Desc: "Count lines of code (use json for JSON output)"},
		{Method: "coverage", Desc: "Calculate test coverage metrics"},
		{Method: "complexity", Desc: "Analyze code complexity"},
		{Method: "size", Desc: "Calculate binary size metrics"},
		{Method: "quality", Desc: "Generate quality metrics report"},
		{Method: "imports", Desc: "Analyze import dependencies"},
		{Method: "mage", Desc: "Analyze magefiles and targets"},
	}
}

func getBenchCommands() []CommandDef {
	return []CommandDef{
		{Method: "default", Desc: "Run benchmarks with optional parameters (time=duration)", Examples: []string{"magex bench", "magex bench time=50ms", "magex bench time=10s count=3"}},
		{Method: "compare", Desc: "Compare benchmark results with optional file parameters", Examples: []string{"magex bench:compare", "magex bench:compare old=baseline.txt new=current.txt"}},
		{Method: "save", Desc: "Save benchmark results with optional parameters (time=duration, output=file)", Examples: []string{"magex bench:save", "magex bench:save time=1s output=results.txt"}},
		{Method: "cpu", Desc: "Run CPU benchmarks with optional parameters (time=duration, profile=file)", Examples: []string{"magex bench:cpu", "magex bench:cpu time=30s profile=cpu-profile.out"}},
		{Method: "mem", Desc: "Run memory benchmarks with optional parameters (time=duration, profile=file)", Examples: []string{"magex bench:mem", "magex bench:mem time=2s profile=mem-profile.out"}},
		{Method: "profile", Desc: "Generate benchmark profiles with optional parameters (time=duration, cpu-profile=file, mem-profile=file)", Examples: []string{"magex bench:profile", "magex bench:profile time=5s cpu-profile=cpu.prof mem-profile=mem.prof"}},
		{Method: "trace", Desc: "Generate execution traces with optional parameters (time=duration, trace=file)", Examples: []string{"magex bench:trace", "magex bench:trace time=10s trace=bench-trace.out"}},
		{Method: "regression", Desc: "Run regression benchmarks with optional parameters (time=duration, update-baseline=true)", Examples: []string{"magex bench:regression", "magex bench:regression time=5s update-baseline=true"}},
	}
}

func getVetCommands() []CommandDef {
	return []CommandDef{
		{Method: "default", Desc: "Run go vet", Aliases: []string{"vet"}},
	}
}

func getConfigureCommands() []CommandDef {
	return []CommandDef{
		{Method: "init", Desc: "Initialize configuration"},
		{Method: "show", Desc: "Show current configuration"},
		{Method: "update", Desc: "Update configuration"},
		{Method: "export", Desc: "Export configuration"},
		{Method: "import", Desc: "Import configuration"},
		{Method: "validate", Desc: "Validate configuration"},
		{Method: "schema", Desc: "Show configuration schema"},
	}
}

func getHelpCommands() []CommandDef {
	return []CommandDef{
		{Method: "default", Desc: "Show help", Aliases: []string{"help"}},
		{Method: "commands", Desc: "List all available commands"},
		{Method: "command", Desc: "Show help for a specific command"},
		{Method: "examples", Desc: "Show usage examples"},
		{Method: "gettingstarted", Desc: "Getting started guide"},
		{Method: "completions", Desc: "Setup shell completions"},
		{Method: "topics", Desc: "List help topics"},
	}
}

func getVersionCommands() []CommandDef {
	return []CommandDef{
		{Method: "show", Desc: "Display current version information"},
		{Method: "check", Desc: "Check version information and compare with latest"},
		{Method: "update", Desc: "Update to latest version"},
		{Method: "bump", Desc: "Bump version with parameters: bump=<major|minor|patch> branch=<branch-name> push dry-run force major-confirm", Usage: "magex version:bump [bump=<type>] [branch=<branch-name>] [push] [dry-run] [force] [major-confirm]", Examples: []string{"magex version:bump bump=patch branch=master push", "magex version:bump bump=minor branch=main", "magex version:bump bump=major major-confirm branch=master push", "magex version:bump bump=patch dry-run", "magex version:bump bump=patch"}},
		{Method: "changelog", Desc: "Generate changelog from git history with parameters: from=<tag> to=<tag>", Usage: "magex version:changelog [from=<tag>] [to=<tag>]", Examples: []string{"magex version:changelog", "magex version:changelog from=v1.0.0", "magex version:changelog from=v1.0.0 to=v1.1.0"}},
		{Method: "tag", Desc: "Create version tag"},
	}
}

func getInstallCommands() []CommandDef {
	return []CommandDef{
		{Method: "default", Desc: "Default installation"},
		{Method: "local", Desc: "Install locally"},
		{Method: "binary", Desc: "Install project binary"},
		{Method: "tools", Desc: "Install development tools"},
		{Method: "go", Desc: "Install Go"},
		{Method: "stdlib", Desc: "Install Go standard library"},
		{Method: "systemwide", Desc: "Install system-wide"},
		{Method: "deps", Desc: "Install dependencies"},
		{Method: "mage", Desc: "Install mage"},
		{Method: "githooks", Desc: "Install git hooks"},
		{Method: "ci", Desc: "Install CI components"},
		{Method: "certs", Desc: "Install certificates"},
		{Method: "package", Desc: "Install package"},
		{Method: "all", Desc: "Install everything"},
		{Method: "uninstall", Desc: "Remove installation"},
	}
}

func getYamlCommands() []CommandDef {
	return []CommandDef{
		{Method: "init", Desc: "Create mage.yaml configuration"},
		{Method: "validate", Desc: "Validate YAML configuration"},
		{Method: "show", Desc: "Show current YAML configuration"},
		{Method: "update", Desc: "Update YAML configuration"},
		{Method: "template", Desc: "Generate YAML templates"},
	}
}

func getBmadCommands() []CommandDef {
	return []CommandDef{
		{Method: "install", Desc: "Install BMAD prerequisites (npm, npx, bmad-method)"},
		{Method: "check", Desc: "Verify BMAD installation and version"},
		{Method: "upgrade", Desc: "Upgrade BMAD to latest version"},
	}
}

// ============================================================================
// Method Bindings - Type-Safe Function References
// ============================================================================

func buildMethodBindings(b mage.Build) map[string]MethodBinding {
	return map[string]MethodBinding{
		"default":  {NoArgs: b.Default},
		"all":      {NoArgs: b.All},
		"linux":    {NoArgs: b.Linux},
		"darwin":   {NoArgs: b.Darwin},
		"windows":  {NoArgs: b.Windows},
		"clean":    {NoArgs: b.Clean},
		"install":  {NoArgs: b.Install},
		"dev":      {NoArgs: b.Dev},
		"generate": {NoArgs: b.Generate},
		"prebuild": {NoArgs: b.PreBuild},
	}
}

func testMethodBindings(t mage.Test) map[string]MethodBinding {
	return map[string]MethodBinding{
		"default":     {WithArgs: t.Default},
		"unit":        {WithArgs: t.Unit},
		"full":        {WithArgs: t.Full},
		"short":       {WithArgs: t.Short},
		"race":        {WithArgs: t.Race},
		"cover":       {WithArgs: t.Cover},
		"coverrace":   {WithArgs: t.CoverRace},
		"coverreport": {NoArgs: t.CoverReport},
		"coverhtml":   {NoArgs: t.CoverHTML},
		"fuzz":        {WithArgs: t.Fuzz},
		"fuzzshort":   {WithArgs: t.FuzzShort},
		"bench":       {WithArgs: t.Bench},
		"benchshort":  {WithArgs: t.BenchShort},
		"integration": {NoArgs: t.Integration},
		"ci":          {NoArgs: t.CI},
		"parallel":    {NoArgs: t.Parallel},
		"nolint":      {NoArgs: t.NoLint},
		"cinorace":    {NoArgs: t.CINoRace},
		"run":         {NoArgs: t.Run},
		"coverage":    {WithArgs: t.Coverage},
		"vet":         {NoArgs: t.Vet},
	}
}

func lintMethodBindings(l mage.Lint) map[string]MethodBinding {
	return map[string]MethodBinding{
		"default": {NoArgs: l.Default},
		"fix":     {NoArgs: l.Fix},
		"ci":      {NoArgs: l.CI},
		"fast":    {NoArgs: l.Fast},
		"issues":  {NoArgs: l.Issues},
	}
}

func formatMethodBindings(f mage.Format) map[string]MethodBinding {
	return map[string]MethodBinding{
		"default": {NoArgs: f.Default},
		"check":   {NoArgs: f.Check},
		"fix":     {NoArgs: f.Fix},
		"imports": {NoArgs: f.Imports},
	}
}

func depsMethodBindings(d mage.Deps) map[string]MethodBinding {
	return map[string]MethodBinding{
		"default":  {NoArgs: d.Default},
		"update":   {WithArgs: d.UpdateWithArgs},
		"tidy":     {NoArgs: d.Tidy},
		"download": {NoArgs: d.Download},
		"vendor":   {NoArgs: d.Vendor},
		"verify":   {NoArgs: d.Verify},
		"outdated": {NoArgs: d.Outdated},
		"graph":    {NoArgs: d.Graph},
		"licenses": {NoArgs: d.Licenses},
	}
}

func gitMethodBindings(g mage.Git) map[string]MethodBinding {
	return map[string]MethodBinding{
		"status":    {NoArgs: g.Status},
		"diff":      {NoArgs: g.Diff},
		"tag":       {WithArgs: g.TagWithArgs},
		"tagremove": {WithArgs: g.TagRemoveWithArgs},
		"tagupdate": {WithArgs: g.TagUpdate},
		"log":       {NoArgs: g.Log},
		"branch":    {NoArgs: g.Branch},
		"pull":      {NoArgs: g.Pull},
		"commit":    {WithArgs: g.Commit},
		"init":      {NoArgs: g.Init},
		"add":       {WithArgs: g.Add},
		"clone":     {NoArgs: g.Clone},
	}
}

func releaseMethodBindings(r mage.Release) map[string]MethodBinding {
	return map[string]MethodBinding{
		"default":      {WithArgs: r.Default},
		"test":         {NoArgs: r.Test},
		"snapshot":     {NoArgs: r.Snapshot},
		"check":        {NoArgs: r.Check},
		"init":         {NoArgs: r.Init},
		"changelog":    {NoArgs: r.Changelog},
		"validate":     {NoArgs: r.Validate},
		"clean":        {NoArgs: r.Clean},
		"localinstall": {NoArgs: r.LocalInstall},
	}
}

func docsMethodBindings(d mage.Docs) map[string]MethodBinding {
	return map[string]MethodBinding{
		"default":  {NoArgs: d.Default},
		"build":    {NoArgs: d.Build},
		"generate": {NoArgs: d.Generate},
		"serve":    {NoArgs: d.Serve},
		"check":    {NoArgs: d.Check},
		"godocs":   {WithArgs: d.GoDocs},
		"examples": {NoArgs: d.Examples},
		"readme":   {NoArgs: d.Readme},
		"api":      {NoArgs: d.API},
		"clean":    {NoArgs: d.Clean},
	}
}

func toolsMethodBindings(t mage.Tools) map[string]MethodBinding {
	return map[string]MethodBinding{
		"install": {NoArgs: t.Install},
		"update":  {NoArgs: t.Update},
		"verify":  {NoArgs: t.Verify},
		"clean":   {NoArgs: t.Clean},
	}
}

func generateMethodBindings(g mage.Generate) map[string]MethodBinding {
	return map[string]MethodBinding{
		"default": {NoArgs: g.Default},
		"all":     {NoArgs: g.All},
		"mocks":   {NoArgs: g.Mocks},
		"proto":   {NoArgs: g.Proto},
		"clean":   {NoArgs: g.Clean},
	}
}

func updateMethodBindings(u mage.Update) map[string]MethodBinding {
	return map[string]MethodBinding{
		"check":   {NoArgs: u.Check},
		"install": {NoArgs: u.Install},
	}
}

func modMethodBindings(m mage.Mod) map[string]MethodBinding {
	return map[string]MethodBinding{
		"tidy":     {NoArgs: m.Tidy},
		"download": {NoArgs: m.Download},
		"update":   {NoArgs: m.Update},
		"clean":    {NoArgs: m.Clean},
		"graph":    {WithArgs: m.Graph},
		"why":      {WithArgs: m.Why}, // Mod.Why takes args ...string
		"vendor":   {NoArgs: m.Vendor},
		"init":     {NoArgs: m.Init},
		"verify":   {NoArgs: m.Verify},
	}
}

func metricsMethodBindings(m mage.Metrics) map[string]MethodBinding {
	return map[string]MethodBinding{
		"loc":        {WithArgs: m.LOC},
		"coverage":   {NoArgs: m.Coverage},
		"complexity": {NoArgs: m.Complexity},
		"size":       {NoArgs: m.Size},
		"quality":    {NoArgs: m.Quality},
		"imports":    {NoArgs: m.Imports},
		"mage":       {NoArgs: m.Mage},
	}
}

func benchMethodBindings(b mage.Bench) map[string]MethodBinding {
	return map[string]MethodBinding{
		"default":    {NoArgs: b.Default, WithArgs: b.DefaultWithArgs},
		"compare":    {NoArgs: b.Compare, WithArgs: b.CompareWithArgs},
		"save":       {NoArgs: b.Save, WithArgs: b.SaveWithArgs},
		"cpu":        {NoArgs: b.CPU, WithArgs: b.CPUWithArgs},
		"mem":        {NoArgs: b.Mem, WithArgs: b.MemWithArgs},
		"profile":    {NoArgs: b.Profile, WithArgs: b.ProfileWithArgs},
		"trace":      {NoArgs: b.Trace, WithArgs: b.TraceWithArgs},
		"regression": {NoArgs: b.Regression, WithArgs: b.RegressionWithArgs},
	}
}

func vetMethodBindings(v mage.Vet) map[string]MethodBinding {
	return map[string]MethodBinding{
		"default": {NoArgs: v.Default},
	}
}

func configureMethodBindings(c mage.Configure) map[string]MethodBinding {
	return map[string]MethodBinding{
		"init":     {NoArgs: c.Init},
		"show":     {NoArgs: c.Show},
		"update":   {NoArgs: c.Update},
		"export":   {NoArgs: c.Export},
		"import":   {NoArgs: c.Import},
		"validate": {NoArgs: c.Validate},
		"schema":   {NoArgs: c.Schema},
	}
}

func helpMethodBindings(h mage.Help) map[string]MethodBinding {
	return map[string]MethodBinding{
		"default":        {NoArgs: h.Default},
		"commands":       {NoArgs: h.Commands},
		"command":        {NoArgs: h.Command},
		"examples":       {NoArgs: h.Examples},
		"gettingstarted": {NoArgs: h.GettingStarted},
		"completions":    {NoArgs: h.Completions},
		"topics":         {NoArgs: h.Topics},
	}
}

func versionMethodBindings(v mage.Version) map[string]MethodBinding {
	return map[string]MethodBinding{
		"show":      {NoArgs: v.Show},
		"check":     {WithArgs: v.Check}, // Version.Check takes _ ...string
		"update":    {NoArgs: v.Update},
		"bump":      {WithArgs: v.Bump},
		"changelog": {WithArgs: v.Changelog},
		"tag":       {WithArgs: v.Tag}, // Version.Tag takes _ ...string
	}
}

func installMethodBindings(i mage.Install) map[string]MethodBinding {
	return map[string]MethodBinding{
		"default":    {NoArgs: i.Default},
		"local":      {NoArgs: i.Local},
		"binary":     {NoArgs: i.Binary},
		"tools":      {NoArgs: i.Tools},
		"go":         {NoArgs: i.Go},
		"stdlib":     {NoArgs: i.Stdlib},
		"systemwide": {NoArgs: i.SystemWide},
		"deps":       {NoArgs: i.Deps},
		"mage":       {NoArgs: i.Mage},
		"githooks":   {NoArgs: i.GitHooks},
		"ci":         {NoArgs: i.CI},
		"certs":      {NoArgs: i.Certs},
		"package":    {NoArgs: i.Package},
		"all":        {NoArgs: i.All},
		"uninstall":  {NoArgs: i.Uninstall},
	}
}

func yamlMethodBindings(y mage.Yaml) map[string]MethodBinding {
	return map[string]MethodBinding{
		"init":     {NoArgs: y.Init},
		"validate": {NoArgs: y.Validate},
		"show":     {NoArgs: y.Show},
		"update":   {NoArgs: y.Update},
		"template": {NoArgs: y.Template},
	}
}

func bmadMethodBindings(b mage.Bmad) map[string]MethodBinding {
	return map[string]MethodBinding{
		"install": {NoArgs: b.Install},
		"check":   {NoArgs: b.Check},
		"upgrade": {NoArgs: b.Upgrade},
	}
}

// ============================================================================
// Registration Functions
// ============================================================================

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

	// Register all namespace commands
	registerBuildCommands(reg)
	registerTestCommands(reg)
	registerLintCommands(reg)
	registerFormatCommands(reg)
	registerDepsCommands(reg)
	registerGitCommands(reg)
	registerReleaseCommands(reg)
	registerDocsCommands(reg)
	registerToolsCommands(reg)
	registerGenerateCommands(reg)
	registerUpdateCommands(reg)
	registerModCommands(reg)
	registerMetricsCommands(reg)
	registerBenchCommands(reg)
	registerVetCommands(reg)
	registerConfigureCommands(reg)
	registerHelpCommands(reg)
	registerVersionCommands(reg)
	registerInstallCommands(reg)
	registerYamlCommands(reg)
	registerBmadCommands(reg)
	registerTopLevelCommands(reg)
}

func registerBuildCommands(reg *registry.Registry) {
	b := mage.Build{}
	registerNamespaceCommands(reg, "build", "Build", getBuildCommands(), buildMethodBindings(b))
}

func registerTestCommands(reg *registry.Registry) {
	t := mage.Test{}
	registerNamespaceCommands(reg, "test", "Test", getTestCommands(), testMethodBindings(t))
}

func registerLintCommands(reg *registry.Registry) {
	l := mage.Lint{}
	registerNamespaceCommands(reg, "lint", "Lint", getLintCommands(), lintMethodBindings(l))
}

func registerFormatCommands(reg *registry.Registry) {
	f := mage.Format{}
	registerNamespaceCommands(reg, "format", "Format", getFormatCommands(), formatMethodBindings(f))
}

func registerDepsCommands(reg *registry.Registry) {
	d := mage.Deps{}
	registerNamespaceCommands(reg, "deps", "Dependencies", getDepsCommands(), depsMethodBindings(d))

	// Special case: deps:audit with Options and LongDescription (kept as explicit builder)
	reg.MustRegister(
		registry.NewNamespaceCommand("deps", "audit").
			WithDescription("Security audit of dependencies (govulncheck)").
			WithLongDescription("Run govulncheck to scan dependencies for known vulnerabilities.\n\n"+
				"Supports CVE exclusions for known/accepted vulnerabilities via:\n"+
				"  - Environment variable: MAGE_X_CVE_EXCLUDES (comma-separated CVE IDs)\n"+
				"  - Command parameter: exclude=CVE-2024-38513,CVE-2023-45142\n\n"+
				"When exclusions are specified, the scan will only fail if non-excluded\n"+
				"vulnerabilities are found. Excluded CVEs are reported as warnings.").
			WithArgsFunc(func(args ...string) error { return d.Audit(args...) }).
			WithCategory("Dependencies").
			WithUsage("magex deps:audit [exclude=CVE-ID,...]").
			WithExamples(
				"magex deps:audit",
				"magex deps:audit exclude=CVE-2024-38513",
				"magex deps:audit exclude=CVE-2024-38513,CVE-2023-45142",
				"MAGE_X_CVE_EXCLUDES=CVE-2024-38513 magex deps:audit",
			).
			WithOptions(
				registry.CommandOption{
					Name:        "exclude",
					Description: "Comma-separated CVE IDs to exclude (e.g., CVE-2024-38513,CVE-2023-45142)",
					Default:     "",
				},
				registry.CommandOption{
					Name:        "MAGE_X_CVE_EXCLUDES",
					Description: "Environment variable with comma-separated CVE IDs to exclude",
					Default:     "",
				},
			).
			MustBuild(),
	)
}

func registerGitCommands(reg *registry.Registry) {
	g := mage.Git{}
	registerNamespaceCommands(reg, "git", "Git", getGitCommands(), gitMethodBindings(g))
}

func registerReleaseCommands(reg *registry.Registry) {
	r := mage.Release{}
	registerNamespaceCommands(reg, "release", "Release", getReleaseCommands(), releaseMethodBindings(r))
}

func registerDocsCommands(reg *registry.Registry) {
	d := mage.Docs{}
	registerNamespaceCommands(reg, "docs", "Documentation", getDocsCommands(), docsMethodBindings(d))
}

func registerToolsCommands(reg *registry.Registry) {
	t := mage.Tools{}
	registerNamespaceCommands(reg, "tools", "Tools", getToolsCommands(), toolsMethodBindings(t))
}

func registerGenerateCommands(reg *registry.Registry) {
	g := mage.Generate{}
	registerNamespaceCommands(reg, "generate", "Generate", getGenerateCommands(), generateMethodBindings(g))
}

func registerUpdateCommands(reg *registry.Registry) {
	u := mage.Update{}
	registerNamespaceCommands(reg, "update", "Update", getUpdateCommands(), updateMethodBindings(u))
}

func registerModCommands(reg *registry.Registry) {
	m := mage.Mod{}
	registerNamespaceCommands(reg, "mod", "Module", getModCommands(), modMethodBindings(m))
}

func registerMetricsCommands(reg *registry.Registry) {
	m := mage.Metrics{}
	registerNamespaceCommands(reg, "metrics", "Metrics", getMetricsCommands(), metricsMethodBindings(m))
}

func registerBenchCommands(reg *registry.Registry) {
	b := mage.Bench{}
	registerNamespaceCommands(reg, "bench", "Benchmark", getBenchCommands(), benchMethodBindings(b))
}

func registerVetCommands(reg *registry.Registry) {
	v := mage.Vet{}
	registerNamespaceCommands(reg, "vet", "Vet", getVetCommands(), vetMethodBindings(v))
}

func registerConfigureCommands(reg *registry.Registry) {
	c := mage.Configure{}
	registerNamespaceCommands(reg, "configure", "Configure", getConfigureCommands(), configureMethodBindings(c))
}

func registerHelpCommands(reg *registry.Registry) {
	h := mage.Help{}
	registerNamespaceCommands(reg, "help", "Help", getHelpCommands(), helpMethodBindings(h))
}

func registerVersionCommands(reg *registry.Registry) {
	v := mage.Version{}
	registerNamespaceCommands(reg, "version", "Version Management", getVersionCommands(), versionMethodBindings(v))
}

func registerInstallCommands(reg *registry.Registry) {
	i := mage.Install{}
	registerNamespaceCommands(reg, "install", "Installation", getInstallCommands(), installMethodBindings(i))
}

func registerYamlCommands(reg *registry.Registry) {
	y := mage.Yaml{}
	registerNamespaceCommands(reg, "yaml", "Configuration", getYamlCommands(), yamlMethodBindings(y))
}

func registerBmadCommands(reg *registry.Registry) {
	b := mage.Bmad{}
	registerNamespaceCommands(reg, "bmad", "AI/ML", getBmadCommands(), bmadMethodBindings(b))
}

func registerTopLevelCommands(reg *registry.Registry) {
	b := mage.Build{}
	t := mage.Test{}
	l := mage.Lint{}
	f := mage.Format{}

	// Top-level convenience commands
	// Note: Some commands need wrappers because the underlying method takes variadic args
	// but we want the top-level command to work without args too
	reg.MustRegister(
		registry.NewCommand("build").
			WithDescription("Build the application").
			WithFunc(b.Default).
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
			WithFunc(l.Default).
			WithCategory("Common").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewCommand("format").
			WithDescription("Format code").
			WithFunc(f.Default).
			WithCategory("Common").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewCommand("clean").
			WithDescription("Clean build artifacts").
			WithFunc(b.Clean).
			WithCategory("Common").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewCommand("install").
			WithDescription("Install the application").
			WithFunc(b.Install).
			WithCategory("Common").
			MustBuild(),
	)

	reg.MustRegister(
		registry.NewCommand("bench").
			WithDescription("Run benchmarks").
			WithArgsFunc(func(args ...string) error { return t.Bench(args...) }).
			WithCategory("Common").
			MustBuild(),
	)
}
