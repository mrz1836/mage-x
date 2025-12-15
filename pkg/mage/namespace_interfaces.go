// Package mage provides namespace interfaces for mage build operations
package mage

import (
	"sync"
	"time"

	"github.com/mrz1836/mage-x/pkg/common/providers"
)

// BuildNamespace interface defines the contract for build operations
type BuildNamespace interface {
	// Default builds the application for the current platform
	Default() error

	// All builds for all configured platforms
	All() error

	// Platform builds for a specific platform (e.g., "linux/amd64")
	Platform(platform string) error

	// Linux builds for Linux (amd64)
	Linux() error

	// Darwin builds for macOS (amd64 and arm64)
	Darwin() error

	// Windows builds for Windows (amd64)
	Windows() error

	// Clean removes build artifacts
	Clean() error

	// Install installs the binary to $GOPATH/bin
	Install() error

	// Generate runs go generate
	Generate() error

	// PreBuild pre-builds all packages to warm cache
	PreBuild() error
}

// TestNamespace interface defines the contract for test operations
type TestNamespace interface {
	// Default runs standard test suite
	Default(args ...string) error

	// Unit runs unit tests only
	Unit(args ...string) error

	// Short runs short tests
	Short(args ...string) error

	// Race runs tests with race detector
	Race(args ...string) error

	// Cover runs tests with coverage
	Cover(args ...string) error

	// CoverRace runs tests with coverage and race detector
	CoverRace(args ...string) error

	// Full runs full test suite with linting
	Full(args ...string) error

	// CoverReport generates coverage report
	CoverReport() error

	// CoverHTML generates HTML coverage report
	CoverHTML() error

	// Fuzz runs fuzz tests with configurable time
	Fuzz(params ...string) error

	// FuzzShort runs short fuzz tests with configurable time
	FuzzShort(params ...string) error

	// FuzzWithTime runs fuzz tests with specified duration (deprecated - use Fuzz with time parameter)
	FuzzWithTime(fuzzTime time.Duration) error

	// FuzzShortWithTime runs short fuzz tests with specified duration (deprecated - use FuzzShort with time parameter)
	FuzzShortWithTime(fuzzTime time.Duration) error

	// Bench runs benchmarks
	Bench(params ...string) error

	// Integration runs integration tests
	Integration() error

	// CI runs CI test suite
	CI() error

	// Parallel runs tests in parallel
	Parallel() error

	// NoLint runs tests without linting
	NoLint() error

	// CINoRace runs CI tests without race detector
	CINoRace() error

	// Run runs a specific test pattern
	Run() error

	// Coverage runs tests and generates coverage
	Coverage(args ...string) error

	// Vet runs go vet
	Vet() error

	// Lint runs linting
	Lint() error

	// Clean cleans test cache
	Clean() error

	// All runs all tests
	All() error
}

// LintNamespace interface defines the contract for linting operations
type LintNamespace interface {
	// Default runs default linters
	Default() error

	// All runs all configured linters
	All() error

	// Go runs Go linters
	Go() error

	// Yaml runs YAML linters
	Yaml() error

	// Fix attempts to fix linting issues
	Fix() error

	// CI runs linters for CI environment
	CI() error

	// Fast runs fast linters only
	Fast() error

	// Config validates linter configuration
	Config() error

	// Issues scans for TODOs, FIXMEs, nolint directives, and test skips
	Issues() error
}

// FormatNamespace interface defines the contract for formatting operations
type FormatNamespace interface {
	// Default formats all code
	Default() error

	// Check checks if code is formatted
	Check() error

	// Go formats Go code
	Go() error

	// Yaml formats YAML files
	Yaml() error

	// JSON formats JSON files
	JSON() error

	// All formats all supported file types
	All() error

	// Fix fixes formatting issues
	Fix() error
}

// DepsNamespace interface defines the contract for dependency operations
type DepsNamespace interface {
	// Default manages default dependencies
	Default() error

	// Download downloads dependencies
	Download() error

	// Update updates dependencies
	Update() error

	// Tidy runs go mod tidy
	Tidy() error

	// Vendor vendors dependencies
	Vendor() error

	// Verify verifies dependencies
	Verify() error

	// Clean cleans dependency cache
	Clean() error

	// Graph shows dependency graph
	Graph() error

	// Why shows why a dependency is needed
	Why(dep string) error

	// VulnCheck checks for vulnerabilities
	VulnCheck() error

	// List lists all dependencies
	List() error

	// Outdated shows outdated dependencies
	Outdated() error

	// Init initializes a new module
	Init(module string) error

	// Licenses shows dependency licenses
	Licenses() error

	// Check checks for updates
	Check() error

	// Audit performs dependency audit
	Audit() error
}

// GitNamespace interface defines the contract for git operations
type GitNamespace interface {
	// Diff shows git diff and checks for uncommitted changes
	Diff() error

	// Tag creates and pushes a new tag
	Tag() error

	// TagRemove removes local and remote tags
	TagRemove() error
}

// ReleaseNamespace interface defines the contract for release operations
type ReleaseNamespace interface {
	// Default creates a default release
	Default(args ...string) error

	// Test runs release dry-run (no publish)
	Test() error

	// Snapshot builds snapshot binaries
	Snapshot() error

	// LocalInstall builds from the latest tag and installs locally
	LocalInstall() error

	// Check validates .goreleaser.yml configuration
	Check() error

	// Init creates a .goreleaser.yml template
	Init() error

	// Changelog generates changelog
	Changelog() error

	// Validate validates release readiness
	Validate() error

	// Clean cleans release artifacts
	Clean() error
}

// DocsNamespace interface defines the contract for documentation operations
type DocsNamespace interface {
	// Default generates default documentation
	Default() error

	// Generate generates documentation
	Generate() error

	// Serve serves documentation locally
	Serve() error

	// Build builds documentation
	Build() error

	// Clean cleans documentation artifacts
	Clean() error

	// Check validates documentation
	Check() error

	// API generates API documentation
	API() error

	// Markdown generates Markdown documentation
	Markdown() error

	// GoDocs triggers GoDocs proxy sync
	GoDocs(args ...string) error

	// Update triggers GoDocs proxy sync (alias for GoDocs)
	Update() error
}

// DeployNamespace interface defines the contract for deployment operations
type DeployNamespace interface {
	// Dev deploys to development environment
	Dev() error

	// Staging deploys to staging environment
	Staging() error

	// Production deploys to production environment
	Production() error

	// Rollback rolls back deployment
	Rollback() error

	// Status shows deployment status
	Status() error

	// Logs shows deployment logs
	Logs() error

	// Scale scales deployment
	Scale() error
}

// ToolsNamespace interface defines the contract for tool management
type ToolsNamespace interface {
	// Default installs default tools
	Default() error

	// Install installs required tools
	Install() error

	// Update updates installed tools
	Update() error

	// Check checks tool versions
	Check() error

	// Clean removes tool installations
	Clean() error

	// List lists installed tools
	List() error
}

// SecurityNamespace interface defines the contract for security operations
type SecurityNamespace interface {
	// Default runs default security checks
	Default() error

	// Scan runs security scans
	Scan() error

	// Audit runs security audit
	Audit() error

	// Dependencies checks dependency vulnerabilities
	Dependencies() error

	// Secrets scans for exposed secrets
	Secrets() error

	// SAST runs static application security testing
	SAST() error

	// License checks license compliance
	License() error

	// Report generates security report
	Report() error
}

// GenerateNamespace interface defines the contract for code generation
type GenerateNamespace interface {
	// Default runs default generators
	Default() error

	// Code generates code
	Code() error

	// Mocks generates mocks
	Mocks() error

	// Proto generates from protobuf
	Proto() error

	// OpenAPI generates from OpenAPI spec
	OpenAPI() error

	// GraphQL generates GraphQL code
	GraphQL() error

	// Docs generates documentation
	Docs() error

	// Clean cleans generated files
	Clean() error
}

// UpdateNamespace interface defines the contract for update operations
type UpdateNamespace interface {
	// Check checks for available updates
	Check() error

	// Install installs the latest update
	Install() error
}

// ModNamespace interface defines the contract for Go module operations
type ModNamespace interface {
	// Download downloads all module dependencies
	Download() error

	// Tidy cleans up go.mod and go.sum
	Tidy() error

	// Update updates all dependencies to latest versions
	Update() error

	// Clean removes the module cache
	Clean() error

	// Graph generates a dependency graph
	Graph(args ...string) error

	// Why shows why a module is needed
	Why(args ...string) error

	// Vendor vendors all dependencies
	Vendor() error

	// Init initializes a new Go module
	Init() error

	// Verify verifies module dependencies
	Verify() error

	// Edit edits go.mod from tools or scripts
	Edit(args ...string) error

	// Get adds dependencies
	Get(packages ...string) error

	// List lists modules
	List(pattern ...string) error
}

// MetricsNamespace interface defines the contract for metrics operations
type MetricsNamespace interface {
	// LOC displays lines of code statistics (use json for JSON output)
	LOC(args ...string) error

	// Coverage analyzes test coverage across the codebase
	Coverage() error

	// Complexity analyzes code complexity
	Complexity() error

	// Size analyzes binary and module sizes
	Size() error

	// Quality runs various code quality metrics
	Quality() error

	// Imports analyzes import dependencies
	Imports() error
}

// NamespaceRegistry provides access to all namespace implementations
type NamespaceRegistry interface {
	// Build returns the build namespace
	Build() BuildNamespace

	// Test returns the test namespace
	Test() TestNamespace

	// Lint returns the lint namespace
	Lint() LintNamespace

	// Format returns the format namespace
	Format() FormatNamespace

	// Deps returns the dependencies namespace
	Deps() DepsNamespace

	// Git returns the git namespace
	Git() GitNamespace

	// Release returns the release namespace
	Release() ReleaseNamespace

	// Docs returns the documentation namespace
	Docs() DocsNamespace

	// Deploy returns the deployment namespace
	Deploy() DeployNamespace

	// Tools returns the tools namespace
	Tools() ToolsNamespace

	// Security returns the security namespace
	Security() SecurityNamespace

	// Generate returns the generate namespace
	Generate() GenerateNamespace

	// Update returns the update namespace
	Update() UpdateNamespace

	// Mod returns the mod namespace
	Mod() ModNamespace

	// Metrics returns the metrics namespace
	Metrics() MetricsNamespace
}

// DefaultNamespaceRegistry provides the default namespace implementations
type DefaultNamespaceRegistry struct {
	mu       sync.RWMutex
	build    BuildNamespace
	buildSet bool // Track whether build was explicitly set
	test     TestNamespace
	lint     LintNamespace
	format   FormatNamespace
	deps     DepsNamespace
	git      GitNamespace
	release  ReleaseNamespace
	docs     DocsNamespace
	deploy   DeployNamespace
	tools    ToolsNamespace
	security SecurityNamespace
	generate GenerateNamespace
	update   UpdateNamespace
	mod      ModNamespace
	metrics  MetricsNamespace
}

// NewNamespaceRegistry creates a new namespace registry with default implementations
func NewNamespaceRegistry() *DefaultNamespaceRegistry {
	return &DefaultNamespaceRegistry{
		// These will be populated with wrapper implementations
	}
}

// Build returns the build namespace
func (r *DefaultNamespaceRegistry) Build() BuildNamespace {
	r.mu.RLock()
	if r.buildSet {
		// Build was explicitly set (could be nil or non-nil)
		defer r.mu.RUnlock()
		return r.build
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.buildSet {
		r.build = NewBuildNamespace()
		r.buildSet = true
	}
	return r.build
}

// Test returns the test namespace
func (r *DefaultNamespaceRegistry) Test() TestNamespace {
	r.mu.RLock()
	if r.test != nil {
		defer r.mu.RUnlock()
		return r.test
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.test == nil {
		r.test = NewTestNamespace()
	}
	return r.test
}

// Lint returns the lint namespace
func (r *DefaultNamespaceRegistry) Lint() LintNamespace {
	r.mu.RLock()
	if r.lint != nil {
		defer r.mu.RUnlock()
		return r.lint
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.lint == nil {
		r.lint = NewLintNamespace()
	}
	return r.lint
}

// Format returns the format namespace
func (r *DefaultNamespaceRegistry) Format() FormatNamespace {
	r.mu.RLock()
	if r.format != nil {
		defer r.mu.RUnlock()
		return r.format
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.format == nil {
		r.format = NewFormatNamespace()
	}
	return r.format
}

// Deps returns the dependencies namespace
func (r *DefaultNamespaceRegistry) Deps() DepsNamespace {
	r.mu.RLock()
	if r.deps != nil {
		defer r.mu.RUnlock()
		return r.deps
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.deps == nil {
		r.deps = NewDepsNamespace()
	}
	return r.deps
}

// Git returns the git namespace
func (r *DefaultNamespaceRegistry) Git() GitNamespace {
	r.mu.RLock()
	if r.git != nil {
		defer r.mu.RUnlock()
		return r.git
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.git == nil {
		r.git = NewGitNamespace()
	}
	return r.git
}

// Release returns the release namespace
func (r *DefaultNamespaceRegistry) Release() ReleaseNamespace {
	r.mu.RLock()
	if r.release != nil {
		defer r.mu.RUnlock()
		return r.release
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.release == nil {
		r.release = NewReleaseNamespace()
	}
	return r.release
}

// Docs returns the documentation namespace
func (r *DefaultNamespaceRegistry) Docs() DocsNamespace {
	r.mu.RLock()
	if r.docs != nil {
		defer r.mu.RUnlock()
		return r.docs
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.docs == nil {
		r.docs = NewDocsNamespace()
	}
	return r.docs
}

// Deploy returns the deployment namespace
func (r *DefaultNamespaceRegistry) Deploy() DeployNamespace {
	// Deploy might not exist in current implementation
	return r.deploy
}

// Tools returns the tools namespace
func (r *DefaultNamespaceRegistry) Tools() ToolsNamespace {
	r.mu.RLock()
	if r.tools != nil {
		defer r.mu.RUnlock()
		return r.tools
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.tools == nil {
		r.tools = NewToolsNamespace()
	}
	return r.tools
}

// Security returns the security namespace
func (r *DefaultNamespaceRegistry) Security() SecurityNamespace {
	// Temporarily disabled - would use lazyInitNamespace(&r.security, NewSecurityNamespace)
	return nil
}

// Generate returns the generate namespace
func (r *DefaultNamespaceRegistry) Generate() GenerateNamespace {
	r.mu.RLock()
	if r.generate != nil {
		defer r.mu.RUnlock()
		return r.generate
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.generate == nil {
		r.generate = NewGenerateNamespace()
	}
	return r.generate
}

// SetBuild sets a custom build namespace implementation
func (r *DefaultNamespaceRegistry) SetBuild(build BuildNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.build = build
	r.buildSet = true
}

// SetTest sets a custom test namespace implementation
func (r *DefaultNamespaceRegistry) SetTest(test TestNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.test = test
}

// SetLint sets a custom lint namespace implementation
func (r *DefaultNamespaceRegistry) SetLint(lint LintNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lint = lint
}

// SetFormat sets a custom format namespace implementation
func (r *DefaultNamespaceRegistry) SetFormat(format FormatNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.format = format
}

// SetDeps sets a custom dependencies namespace implementation
func (r *DefaultNamespaceRegistry) SetDeps(deps DepsNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deps = deps
}

// SetGit sets a custom git namespace implementation
func (r *DefaultNamespaceRegistry) SetGit(git GitNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.git = git
}

// SetRelease sets a custom release namespace implementation
func (r *DefaultNamespaceRegistry) SetRelease(release ReleaseNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.release = release
}

// SetDocs sets a custom documentation namespace implementation
func (r *DefaultNamespaceRegistry) SetDocs(docs DocsNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.docs = docs
}

// SetDeploy sets a custom deployment namespace implementation
func (r *DefaultNamespaceRegistry) SetDeploy(deploy DeployNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deploy = deploy
}

// SetTools sets a custom tools namespace implementation
func (r *DefaultNamespaceRegistry) SetTools(tools ToolsNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools = tools
}

// SetSecurity sets a custom security namespace implementation
func (r *DefaultNamespaceRegistry) SetSecurity(security SecurityNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.security = security
}

// SetGenerate sets a custom generate namespace implementation
func (r *DefaultNamespaceRegistry) SetGenerate(generate GenerateNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.generate = generate
}

// Update returns the update namespace
func (r *DefaultNamespaceRegistry) Update() UpdateNamespace {
	r.mu.RLock()
	if r.update != nil {
		defer r.mu.RUnlock()
		return r.update
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.update == nil {
		r.update = NewUpdateNamespace()
	}
	return r.update
}

// SetUpdate sets a custom update namespace implementation
func (r *DefaultNamespaceRegistry) SetUpdate(update UpdateNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.update = update
}

// Mod returns the mod namespace
func (r *DefaultNamespaceRegistry) Mod() ModNamespace {
	r.mu.RLock()
	if r.mod != nil {
		defer r.mu.RUnlock()
		return r.mod
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.mod == nil {
		r.mod = NewModNamespace()
	}
	return r.mod
}

// SetMod sets a custom mod namespace implementation
func (r *DefaultNamespaceRegistry) SetMod(mod ModNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mod = mod
}

// Metrics returns the metrics namespace
func (r *DefaultNamespaceRegistry) Metrics() MetricsNamespace {
	r.mu.RLock()
	if r.metrics != nil {
		defer r.mu.RUnlock()
		return r.metrics
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()
	if r.metrics == nil {
		r.metrics = NewMetricsNamespace()
	}
	return r.metrics
}

// SetMetrics sets a custom metrics namespace implementation
func (r *DefaultNamespaceRegistry) SetMetrics(metrics MetricsNamespace) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics = metrics
}

// GetNamespaceRegistry returns the package-level namespace registry with thread-safe lazy initialization
func GetNamespaceRegistry() *DefaultNamespaceRegistry {
	return getNamespaceRegistryInstance()
}

// SetNamespaceRegistry sets the package-level namespace registry (primarily for testing)
// Deprecated: Use SetNamespaceRegistryProvider for better testability
func SetNamespaceRegistry(registry *DefaultNamespaceRegistry) {
	// Create a custom provider that returns the provided registry
	provider := &customNamespaceRegistryProvider{registry: registry}
	setNamespaceRegistryProvider(provider)
}

// customNamespaceRegistryProvider is a simple provider for the legacy SetNamespaceRegistry function
type customNamespaceRegistryProvider struct {
	registry *DefaultNamespaceRegistry
}

// GetNamespaceRegistry returns the pre-set registry
func (p *customNamespaceRegistryProvider) GetNamespaceRegistry() *DefaultNamespaceRegistry {
	return p.registry
}

// NamespaceRegistryProvider defines the interface for providing namespace registry instances
type NamespaceRegistryProvider interface {
	GetNamespaceRegistry() *DefaultNamespaceRegistry
}

// DefaultNamespaceRegistryProvider provides a thread-safe singleton namespace registry using generic provider
type DefaultNamespaceRegistryProvider struct {
	*providers.Provider[*DefaultNamespaceRegistry]
}

// NewDefaultNamespaceRegistryProvider creates a new default namespace registry provider using generic framework
func NewDefaultNamespaceRegistryProvider() *DefaultNamespaceRegistryProvider {
	factory := func() *DefaultNamespaceRegistry {
		return NewNamespaceRegistry()
	}

	return &DefaultNamespaceRegistryProvider{
		Provider: providers.NewProvider(factory),
	}
}

// GetNamespaceRegistry returns a namespace registry instance using thread-safe singleton pattern
func (p *DefaultNamespaceRegistryProvider) GetNamespaceRegistry() *DefaultNamespaceRegistry {
	return p.Get()
}

// packageNamespaceRegistryProvider provides a generic package-level namespace registry provider using the generic framework
//
//nolint:gochecknoglobals // Required for package-level singleton access pattern
var packageNamespaceRegistryProvider = providers.NewPackageProvider(func() NamespaceRegistryProvider {
	return NewDefaultNamespaceRegistryProvider()
})

// getNamespaceRegistryInstance returns the namespace registry using the generic package provider
func getNamespaceRegistryInstance() *DefaultNamespaceRegistry {
	return packageNamespaceRegistryProvider.Get().GetNamespaceRegistry()
}

// setNamespaceRegistryProvider sets a custom namespace registry provider using the generic package provider
func setNamespaceRegistryProvider(provider NamespaceRegistryProvider) {
	packageNamespaceRegistryProvider.Set(provider)
}

// SetNamespaceRegistryProvider sets a custom namespace registry provider
// This allows for dependency injection and testing with mock providers
func SetNamespaceRegistryProvider(provider NamespaceRegistryProvider) {
	setNamespaceRegistryProvider(provider)
}
