// Package mage provides core interfaces for extensible build operations
package mage

import (
	"context"
	"time"
)

// Core operation interfaces that allow developers to override default implementations

// Builder interface for customizable build operations
type Builder interface {
	// Build builds the application with the given options
	Build(ctx context.Context, opts BuildOptions) error

	// CrossCompile builds for multiple platforms
	CrossCompile(ctx context.Context, platforms []Platform) error

	// Package creates distribution packages
	Package(ctx context.Context, format PackageFormat) error

	// Clean removes build artifacts
	Clean(ctx context.Context) error

	// Install installs the built binary
	Install(ctx context.Context, target string) error
}

// Tester interface for customizable test operations
type Tester interface {
	// RunTests executes the test suite
	RunTests(ctx context.Context, opts TestOptions) (*TestResults, error)

	// RunBenchmarks executes benchmark tests
	RunBenchmarks(ctx context.Context, opts IBenchmarkOptions) (*IBenchmarkResults, error)

	// GenerateCoverage creates test coverage reports
	GenerateCoverage(ctx context.Context, opts CoverageOptions) (*CoverageReport, error)

	// RunUnit executes unit tests only
	RunUnit(ctx context.Context, opts TestOptions) (*TestResults, error)

	// RunIntegration executes integration tests
	RunIntegration(ctx context.Context, opts TestOptions) (*TestResults, error)
}

// Deployer interface for multi-platform deployments
type Deployer interface {
	// Deploy deploys the application to a target environment
	Deploy(ctx context.Context, target DeployTarget, opts DeployOptions) error

	// Rollback rolls back to a previous deployment version
	Rollback(ctx context.Context, target DeployTarget, version string) error

	// Status gets the current deployment status
	Status(ctx context.Context, target DeployTarget) (*DeployStatus, error)

	// Scale adjusts the deployment scale
	Scale(ctx context.Context, target DeployTarget, replicas int) error

	// Logs retrieves deployment logs
	Logs(ctx context.Context, target DeployTarget, opts LogOptions) ([]string, error)
}

// Releaser interface for release operations
type Releaser interface {
	// CreateRelease creates a new release
	CreateRelease(ctx context.Context, version string, opts ReleaseOptions) (*IRelease, error)

	// PublishRelease publishes a release to distribution channels
	PublishRelease(ctx context.Context, release *IRelease, channels []string) error

	// GenerateChangelog creates release notes and changelog
	GenerateChangelog(ctx context.Context, fromVersion, toVersion string) (*Changelog, error)

	// ValidateRelease validates release artifacts and metadata
	ValidateRelease(ctx context.Context, release *IRelease) error

	// ArchiveRelease creates release archives
	ArchiveRelease(ctx context.Context, release *IRelease, formats []ArchiveFormat) error
}

// ISecurityScanner interface for security operations
type ISecurityScanner interface {
	// ScanVulnerabilities scans for security vulnerabilities
	ScanVulnerabilities(ctx context.Context, opts ScanOptions) (*SecurityReport, error)

	// ScanDependencies checks dependency security
	ScanDependencies(ctx context.Context, opts ScanOptions) (*DependencyReport, error)

	// ScanSecrets detects exposed secrets
	ScanSecrets(ctx context.Context, opts ScanOptions) (*SecretReport, error)

	// GenerateSecurityReport creates comprehensive security report
	GenerateSecurityReport(ctx context.Context, opts ReportOptions) (*SecurityReport, error)

	// ValidateCompliance checks compliance with security policies
	ValidateCompliance(ctx context.Context, policies []ISecurityPolicy) (*IComplianceReport, error)
}

// Linter interface for code quality operations
type Linter interface {
	// Lint runs linting on the codebase
	Lint(ctx context.Context, opts LintOptions) (*LintReport, error)

	// Format formats code according to style guidelines
	Format(ctx context.Context, opts FormatOptions) error

	// Fix automatically fixes linting issues where possible
	Fix(ctx context.Context, opts LintOptions) (*LintReport, error)

	// ValidateStyle checks code style compliance
	ValidateStyle(ctx context.Context, opts StyleOptions) (*StyleReport, error)
}

// ContainerProvider interface for container operations
type ContainerProvider interface {
	// Build builds a container image
	Build(ctx context.Context, opts ContainerBuildOptions) (*ContainerImage, error)

	// Push pushes an image to a registry
	Push(ctx context.Context, image *ContainerImage, registry string) error

	// Pull pulls an image from a registry
	Pull(ctx context.Context, imageRef string) (*ContainerImage, error)

	// Run runs a container
	Run(ctx context.Context, image *ContainerImage, opts ContainerRunOptions) (*ContainerInstance, error)

	// Stop stops a running container
	Stop(ctx context.Context, instance *ContainerInstance) error
}

// CloudProvider interface for cloud deployment operations
type CloudProvider interface {
	// Name returns the provider name (e.g., "aws", "gcp", "azure")
	Name() string

	// Deploy deploys an application to the cloud
	Deploy(ctx context.Context, app Application, env Environment) (*Deployment, error)

	// GetStatus retrieves deployment status
	GetStatus(ctx context.Context, deployment *Deployment) (*DeploymentStatus, error)

	// Scale adjusts application scale
	Scale(ctx context.Context, deployment *Deployment, replicas int) error

	// Rollback rolls back to a previous version
	Rollback(ctx context.Context, deployment *Deployment, version string) error

	// GetLogs retrieves application logs
	GetLogs(ctx context.Context, deployment *Deployment, opts LogOptions) ([]string, error)
}

// ToolProvider interface for external tool management
type ToolProvider interface {
	// Name returns the tool name
	Name() string

	// Install installs the tool
	Install(ctx context.Context, version string) error

	// Uninstall removes the tool
	Uninstall(ctx context.Context) error

	// GetVersion returns the installed version
	GetVersion(ctx context.Context) (string, error)

	// Update updates to the latest version
	Update(ctx context.Context) error

	// IsInstalled checks if the tool is installed
	IsInstalled(ctx context.Context) (bool, error)
}

// CommandRunner interface for executing commands
type CommandRunner interface {
	RunCmd(name string, args ...string) error
	RunCmdOutput(name string, args ...string) (string, error)
}

// ContextCommandRunner provides enhanced CommandRunner with context support
type ContextCommandRunner interface {
	CommandRunner // Embed existing interface for backward compatibility

	// RunCmdWithContext executes a command with context
	RunCmdWithContext(ctx context.Context, name string, args ...string) error

	// RunCmdOutputWithContext executes a command and returns output with context
	RunCmdOutputWithContext(ctx context.Context, name string, args ...string) (string, error)

	// RunCmdStreamWithContext executes a command with streaming output
	RunCmdStreamWithContext(ctx context.Context, name string, args ...string) error
}

// Supporting data structures

// BuildOptions contains build configuration
type BuildOptions struct {
	Tags       []string
	LDFlags    []string
	GoFlags    []string
	Output     string
	Verbose    bool
	TrimPath   bool
	CGOEnabled bool
}

// Platform represents a build target platform
type Platform struct {
	OS   string
	Arch string
}

// PackageFormat represents distribution package format
type PackageFormat string

const (
	PackageFormatTarGz PackageFormat = "tar.gz"
	PackageFormatZip   PackageFormat = "zip"
	PackageFormatDeb   PackageFormat = "deb"
	PackageFormatRPM   PackageFormat = "rpm"
	PackageFormatDMG   PackageFormat = "dmg"
)

// TestOptions contains test configuration
type TestOptions struct {
	Packages  []string
	Tags      []string
	Timeout   time.Duration
	Parallel  int
	Race      bool
	Cover     bool
	Verbose   bool
	ShortMode bool
	FailFast  bool
}

// TestResults contains test execution results
type TestResults struct {
	Passed   int
	Failed   int
	Skipped  int
	Duration time.Duration
	Coverage *CoverageInfo
	Failures []TestFailure
}

// CoverageInfo contains coverage statistics
type CoverageInfo struct {
	Percentage float64
	Lines      int
	Covered    int
	Uncovered  int
}

// TestFailure represents a failed test
type TestFailure struct {
	Package string
	Test    string
	Error   string
	Output  string
}

// BenchmarkOptions contains benchmark configuration
type IBenchmarkOptions struct {
	Pattern    string
	Count      int
	Duration   time.Duration
	CPUProfile string
	MemProfile string
	OutputFile string
}

// BenchmarkResults contains benchmark execution results
type IBenchmarkResults struct {
	Benchmarks []IBenchmarkResult
	Duration   time.Duration
}

// IBenchmarkResult represents a single benchmark result for the interface
type IBenchmarkResult struct {
	Name        string
	Runs        int
	NsPerOp     int64
	BytesPerOp  int64
	AllocsPerOp int64
}

// CoverageOptions contains coverage generation options
type CoverageOptions struct {
	Mode       string // set, count, atomic
	OutputFile string
	HTMLOutput string
	Packages   []string
}

// CoverageReport contains detailed coverage information
type CoverageReport struct {
	Overall  *CoverageInfo
	Packages []PackageCoverage
}

// PackageCoverage contains per-package coverage
type PackageCoverage struct {
	Package  string
	Coverage *CoverageInfo
	Files    []FileCoverage
}

// FileCoverage contains per-file coverage
type FileCoverage struct {
	File     string
	Coverage *CoverageInfo
	Lines    []LineCoverage
}

// LineCoverage contains per-line coverage
type LineCoverage struct {
	Line    int
	Covered bool
	Count   int
}

// DeployTarget represents a deployment target
type DeployTarget struct {
	Name        string
	Environment string
	Provider    string
	Region      string
	Config      map[string]interface{}
}

// DeployOptions contains deployment configuration
type DeployOptions struct {
	Image       string
	Version     string
	Replicas    int
	Resources   ResourceRequirements
	Environment map[string]string
	Secrets     map[string]string
	Volumes     []VolumeMount
	HealthCheck *HealthCheck
}

// ResourceRequirements defines resource limits
type ResourceRequirements struct {
	CPU    string
	Memory string
	Disk   string
}

// VolumeMount represents a volume mount
type VolumeMount struct {
	Name      string
	MountPath string
	ReadOnly  bool
}

// HealthCheck defines health check configuration
type HealthCheck struct {
	Path     string
	Port     int
	Interval time.Duration
	Timeout  time.Duration
	Retries  int
}

// DeployStatus represents deployment status
type DeployStatus struct {
	State       string
	Ready       int
	Total       int
	Version     string
	LastUpdated time.Time
	Conditions  []StatusCondition
}

// StatusCondition represents a deployment condition
type StatusCondition struct {
	Type    string
	Status  string
	Reason  string
	Message string
}

// LogOptions contains log retrieval options
type LogOptions struct {
	Follow    bool
	Tail      int
	Since     time.Time
	Container string
	Previous  bool
}

// IRelease represents a software release for interfaces
type IRelease struct {
	Version     string
	Name        string
	Description string
	Assets      []ReleaseAsset
	Changelog   *Changelog
	CreatedAt   time.Time
	PublishedAt *time.Time
	Draft       bool
	Prerelease  bool
	TagName     string
}

// ReleaseAsset represents a release artifact
type ReleaseAsset struct {
	Name     string
	URL      string
	Size     int64
	Checksum string
	Platform *Platform
}

// ReleaseOptions contains release creation options
type ReleaseOptions struct {
	Draft      bool
	Prerelease bool
	Assets     []string
	Changelog  string
	Target     string
}

// Changelog represents release notes and changes
type Changelog struct {
	Version  string
	Date     time.Time
	Changes  []ChangeEntry
	Sections map[string][]ChangeEntry
}

// ChangeEntry represents a single change
type ChangeEntry struct {
	Type        string // feat, fix, docs, etc.
	Description string
	Scope       string
	Breaking    bool
	Author      string
	CommitHash  string
}

// ArchiveFormat represents archive file format
type ArchiveFormat string

const (
	ArchiveFormatTarGz ArchiveFormat = "tar.gz"
	ArchiveFormatZip   ArchiveFormat = "zip"
	ArchiveFormatTarXz ArchiveFormat = "tar.xz"
)

// ScanOptions contains security scan configuration
type ScanOptions struct {
	Paths      []string
	Exclude    []string
	Severity   []string
	Format     string
	OutputFile string
	Recursive  bool
	IncludeDev bool
}

// SecurityReport contains security scan results
type SecurityReport struct {
	Summary         *SecuritySummary
	Vulnerabilities []IVulnerability
	Dependencies    *DependencyReport
	Secrets         *SecretReport
	GeneratedAt     time.Time
}

// SecuritySummary contains scan summary
type SecuritySummary struct {
	Total    int
	Critical int
	High     int
	Medium   int
	Low      int
	Info     int
}

// Vulnerability represents a security vulnerability
type IVulnerability struct {
	ID          string
	Title       string
	Description string
	Severity    string
	CVSS        float64
	Package     string
	Version     string
	FixedIn     string
	References  []string
}

// DependencyReport contains dependency security information
type DependencyReport struct {
	Dependencies    []IDependency
	Vulnerabilities []IVulnerability
	Outdated        []IOutdatedDependency
}

// Dependency represents a project dependency
type IDependency struct {
	Name    string
	Version string
	Type    string
	Direct  bool
	License string
}

// OutdatedDependency represents an outdated dependency
type IOutdatedDependency struct {
	IDependency
	LatestVersion string
	UpdateType    string
}

// SecretReport contains secret scan results
type SecretReport struct {
	Secrets []DetectedSecret
	Files   []string
}

// DetectedSecret represents a detected secret
type DetectedSecret struct {
	Type        string
	File        string
	Line        int
	Description string
	Confidence  string
	Value       string // Redacted
}

// ReportOptions contains security report options
type ReportOptions struct {
	Format     string
	OutputFile string
	Template   string
	Include    []string
	Exclude    []string
}

// ISecurityPolicy represents a security policy for interfaces
type ISecurityPolicy struct {
	ID          string
	Name        string
	Description string
	Rules       []IPolicyRule
	Severity    string
	Enabled     bool
}

// IPolicyRule represents a policy rule for interfaces
type IPolicyRule struct {
	Type        string `json:"type" yaml:"type"`
	Pattern     string `json:"pattern" yaml:"pattern"`
	Action      string `json:"action" yaml:"action"`
	Severity    string `json:"severity" yaml:"severity"`
	Description string `json:"description" yaml:"description"`
	Message     string `json:"message" yaml:"message"`
	Remediation string `json:"remediation" yaml:"remediation"`
}

// IComplianceReport contains compliance check results for interfaces
type IComplianceReport struct {
	Policies []IPolicyResult
	Summary  *IComplianceSummary
}

// IPolicyResult represents policy check result for interfaces
type IPolicyResult struct {
	Policy   *ISecurityPolicy
	Status   string
	Findings []PolicyFinding
}

// PolicyFinding represents a policy violation
type PolicyFinding struct {
	Rule        *IPolicyRule
	Location    string
	Message     string
	Severity    string
	Remediation string
}

// ComplianceSummary contains compliance summary
type IComplianceSummary struct {
	Total   int
	Passed  int
	Failed  int
	Skipped int
	Score   float64
}

// LintOptions contains linting configuration
type LintOptions struct {
	Paths      []string
	Exclude    []string
	Rules      []string
	Config     string
	Format     string
	OutputFile string
	Fix        bool
}

// LintReport contains linting results
type LintReport struct {
	Issues   []LintIssue
	Summary  *LintSummary
	Duration time.Duration
}

// LintIssue represents a single lint issue
type LintIssue struct {
	File     string
	Line     int
	Column   int
	Severity string
	Message  string
	Rule     string
}

// LintSummary provides a summary of lint results
type LintSummary struct {
	TotalIssues int
	Errors      int
	Warnings    int
	Info        int
}

// LintIssue represents a linting issue
type ILintIssue struct {
	File       string
	Line       int
	Column     int
	Rule       string
	Severity   string
	Message    string
	Suggestion string
	Fixable    bool
}

// LintSummary contains linting summary
type ILintSummary struct {
	Total    int
	Errors   int
	Warnings int
	Info     int
	Fixable  int
}

// FormatOptions contains code formatting options
type FormatOptions struct {
	Paths     []string
	Exclude   []string
	Config    string
	DryRun    bool
	Recursive bool
}

// StyleOptions contains style checking options
type StyleOptions struct {
	Paths   []string
	Exclude []string
	Config  string
	Strict  bool
}

// StyleReport contains style check results
type StyleReport struct {
	Issues   []StyleIssue
	Summary  *StyleSummary
	Duration time.Duration
}

// StyleIssue represents a style issue
type StyleIssue struct {
	File     string
	Line     int
	Column   int
	Rule     string
	Message  string
	Expected string
	Actual   string
}

// StyleSummary contains style check summary
type StyleSummary struct {
	Total  int
	Files  int
	Passed int
	Failed int
}

// Container-related types

// ContainerBuildOptions contains container build configuration
type ContainerBuildOptions struct {
	Dockerfile string
	Context    string
	Tags       []string
	BuildArgs  map[string]string
	Target     string
	Platform   string
	NoCache    bool
}

// ContainerImage represents a container image
type ContainerImage struct {
	ID       string
	Tags     []string
	Size     int64
	Created  time.Time
	Platform string
}

// ContainerRunOptions contains container run configuration
type ContainerRunOptions struct {
	Name        string
	Ports       []PortMapping
	Volumes     []VolumeMount
	Environment map[string]string
	WorkingDir  string
	Command     []string
	Args        []string
	Detach      bool
	Remove      bool
}

// PortMapping represents a port mapping
type PortMapping struct {
	HostPort      int
	ContainerPort int
	Protocol      string
}

// ContainerInstance represents a running container
type ContainerInstance struct {
	ID     string
	Name   string
	Image  string
	Status string
	Ports  []PortMapping
}

// Cloud provider types

// Application represents a deployable application
type Application struct {
	Name        string
	Version     string
	Image       string
	Config      map[string]interface{}
	Secrets     map[string]string
	Environment map[string]string
}

// IEnvironment represents a deployment environment for interfaces
type IEnvironment struct {
	Name   string
	Region string
	Config map[string]interface{}
}

// Deployment represents a cloud deployment
type Deployment struct {
	ID          string
	Application *Application
	Environment *IEnvironment
	Status      string
	Version     string
	Replicas    int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// DeploymentStatus represents detailed deployment status
type DeploymentStatus struct {
	State       string
	Ready       int
	Total       int
	Version     string
	LastUpdated time.Time
	Conditions  []StatusCondition
	Events      []DeploymentEvent
}

// DeploymentEvent represents a deployment event
type DeploymentEvent struct {
	Type      string
	Reason    string
	Message   string
	Timestamp time.Time
}
