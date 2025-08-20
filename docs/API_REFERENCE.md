# API Reference: Namespace Interface Architecture

## Core Interfaces

### BuildNamespace

The BuildNamespace interface provides methods for building and compilation tasks.

```go
type BuildNamespace interface {
    Default() error         // Default build
    All() error            // Build for all platforms
    Platform(string) error // Build for specific platform
    Linux() error          // Build for Linux
    Darwin() error         // Build for macOS
    Windows() error        // Build for Windows
    Docker() error         // Build Docker image
    Clean() error          // Clean build artifacts
    Install() error        // Install built binary
    Generate() error       // Generate build files
    PreBuild() error       // Pre-build tasks
}
```

**Factory Function**: `NewBuildNamespace() BuildNamespace`

**Usage**:
```go
build := mage.NewBuildNamespace()
err := build.Default()
```

### TestNamespace

The TestNamespace interface provides methods for testing operations.

```go
type TestNamespace interface {
    Default() error        // Run default tests
    Unit() error          // Run unit tests
    Integration() error   // Run integration tests
    Benchmark() error     // Run benchmarks
    Coverage() error      // Generate coverage report
    Race() error          // Run tests with race detection
    Short() error         // Run short tests only
    Verbose() error       // Run tests with verbose output
    All() error           // Run all tests
}
```

**Factory Function**: `NewTestNamespace() TestNamespace`

**Usage**:
```go
test := mage.NewTestNamespace()
err := test.Unit()
```

### LintNamespace

The LintNamespace interface provides methods for code linting and analysis.

```go
type LintNamespace interface {
    Default() error    // Run default linter
    All() error       // Run all linters
    Fix() error       // Fix linting issues automatically
    CI() error        // Run CI-specific linting
    Fast() error      // Run fast linting only
}
```

**Factory Function**: `NewLintNamespace() LintNamespace`

**Usage**:
```go
lint := mage.NewLintNamespace()
err := lint.Default()
```

### FormatNamespace

The FormatNamespace interface provides methods for code formatting.

```go
type FormatNamespace interface {
    Default() error    // Format with default settings
    Check() error     // Check formatting without changes
    Fix() error       // Fix formatting issues
    All() error       // Format all supported files
}
```

**Factory Function**: `NewFormatNamespace() FormatNamespace`

### DocNamespace

The DocNamespace interface provides methods for documentation generation.

```go
type DocNamespace interface {
    Default() error        // Generate default documentation
    Build() error         // Build documentation
    Serve() error         // Serve documentation locally
    Deploy() error        // Deploy documentation
    Clean() error         // Clean documentation artifacts
    Markdown() error      // Generate Markdown documentation
}
```

**Factory Function**: `NewDocNamespace() DocNamespace`

### ReleaseNamespace

The ReleaseNamespace interface provides methods for release management.

```go
type ReleaseNamespace interface {
    Default() error           // Create default release
    Major() error            // Create major version release
    Minor() error            // Create minor version release
    Patch() error            // Create patch version release
    Prerelease() error       // Create prerelease
    Changelog() error        // Generate changelog
    Tag() error              // Create release tag
    Publish() error          // Publish release
}
```

**Factory Function**: `NewReleaseNamespace() ReleaseNamespace`

## Registry API

### NamespaceRegistry

The NamespaceRegistry provides centralized namespace management.

```go
type NamespaceRegistry interface {
    Register(name string, namespace interface{}) error
    Get(name string) interface{}
    Available() []string
    Unregister(name string) error
}
```

**Constructor**: `NewNamespaceRegistry() *NamespaceRegistry`

**Methods**:

#### Register
```go
func (r *NamespaceRegistry) Register(name string, namespace interface{}) error
```
Registers a custom namespace implementation.

**Parameters**:
- `name`: Unique identifier for the namespace
- `namespace`: Implementation that satisfies the corresponding interface

**Returns**: Error if registration fails (e.g., duplicate name)

**Example**:
```go
registry := mage.NewNamespaceRegistry()
customBuild := &MyCustomBuild{}
err := registry.Register("build", customBuild)
```

#### Get
```go
func (r *NamespaceRegistry) Get(name string) interface{}
```
Retrieves a registered namespace implementation.

**Parameters**:
- `name`: Identifier of the namespace to retrieve

**Returns**: Namespace implementation or nil if not found

**Example**:
```go
build := registry.Get("build").(mage.BuildNamespace)
```

#### Available
```go
func (r *NamespaceRegistry) Available() []string
```
Returns list of available namespace names.

**Returns**: Slice of registered namespace names

**Example**:
```go
namespaces := registry.Available()
fmt.Println("Available:", namespaces)
```

#### Unregister
```go
func (r *NamespaceRegistry) Unregister(name string) error
```
Removes a namespace from the registry.

**Parameters**:
- `name`: Identifier of the namespace to remove

**Returns**: Error if unregistration fails

## Factory Functions

All factory functions follow the pattern `New{Namespace}Namespace()` and return the corresponding interface type.

### Core Development
```go
func NewBuildNamespace() BuildNamespace
func NewTestNamespace() TestNamespace
func NewLintNamespace() LintNamespace
func NewFormatNamespace() FormatNamespace
func NewDocNamespace() DocNamespace
```

### Project Management
```go
func NewInitNamespace() InitNamespace
func NewGenerateNamespace() GenerateNamespace
func NewReleaseNamespace() ReleaseNamespace
func NewVersionNamespace() VersionNamespace
func NewUpdateNamespace() UpdateNamespace
```

### Dependencies & Tools
```go
func NewDepsNamespace() DepsNamespace
func NewToolsNamespace() ToolsNamespace
func NewModNamespace() ModNamespace
func NewVetNamespace() VetNamespace
```

### Deployment & Infrastructure
```go
func NewDockerNamespace() DockerNamespace
func NewK8sNamespace() K8sNamespace
func NewInstallNamespace() InstallNamespace
func NewConfigureNamespace() ConfigureNamespace
```

### Quality & Security
```go
func NewAuditNamespace() AuditNamespace
func NewSecurityNamespace() SecurityNamespace
func NewBenchNamespace() BenchNamespace
func NewMetricsNamespace() MetricsNamespace
```

### Integration & Workflow
```go
func NewIntegrationsNamespace() IntegrationsNamespace
func NewWorkflowNamespace() WorkflowNamespace
func NewReleaseManagerNamespace() ReleaseManagerNamespace
func NewOperationsNamespace() OperationsNamespace
```

### Enterprise Features
```go
func NewEnterpriseNamespace() EnterpriseNamespace
func NewEnterpriseConfigNamespace() EnterpriseConfigNamespace
func NewAnalytics() AnalyticsNamespace
func NewDatabaseNamespace() DatabaseNamespace
```

### Utilities
```go
func NewCLINamespace() CLINamespace
func NewHelpNamespace() HelpNamespace
func NewYAMLNamespace() YAMLNamespace
func NewWizardNamespace() WizardNamespace
func NewCommonNamespace() CommonNamespace
func NewRecipesNamespace() RecipesNamespace
```

## Error Types

### RegistrationError
```go
type RegistrationError struct {
    Name   string
    Reason string
}

func (e *RegistrationError) Error() string
```

Returned when namespace registration fails.

### NamespaceNotFoundError
```go
type NamespaceNotFoundError struct {
    Name string
}

func (e *NamespaceNotFoundError) Error() string
```

Returned when attempting to retrieve non-existent namespace.

## Type Safety

### Interface Compliance Verification

You can verify interface compliance at compile time:

```go
// Verify custom implementation satisfies interface
var _ mage.BuildNamespace = (*MyCustomBuild)(nil)
var _ mage.TestNamespace = (*MyCustomTest)(nil)
```

### Safe Type Assertions

Use safe type assertions to avoid panics:

```go
// Safe assertion
if build, ok := registry.Get("build").(mage.BuildNamespace); ok {
    err := build.Default()
} else {
    return fmt.Errorf("build namespace not available")
}

// Type switch for multiple types
switch ns := registry.Get(name).(type) {
case mage.BuildNamespace:
    return ns.Default()
case mage.TestNamespace:
    return ns.Unit()
default:
    return fmt.Errorf("unknown namespace type: %T", ns)
}
```

## Advanced Patterns

### Interface Composition

Create higher-level interfaces by composing namespace interfaces:

```go
type BuildTestNamespace interface {
    mage.BuildNamespace
    mage.TestNamespace
}

type CompletePipeline struct {
    build mage.BuildNamespace
    test  mage.TestNamespace
}

func (c *CompletePipeline) Default() error {
    if err := c.build.Default(); err != nil {
        return err
    }
    return c.test.Unit()
}

// Verify composition
var _ BuildTestNamespace = (*CompletePipeline)(nil)
```

### Namespace Middleware

Implement middleware patterns using interface wrapping:

```go
type LoggingBuild struct {
    mage.BuildNamespace
    logger *log.Logger
}

func (l *LoggingBuild) Default() error {
    l.logger.Println("Starting build...")
    err := l.BuildNamespace.Default()
    if err != nil {
        l.logger.Printf("Build failed: %v", err)
    } else {
        l.logger.Println("Build completed successfully")
    }
    return err
}

// Usage
build := &LoggingBuild{
    BuildNamespace: mage.NewBuildNamespace(),
    logger:         log.New(os.Stdout, "[BUILD] ", log.LstdFlags),
}
```

### Dynamic Namespace Loading

Load namespaces dynamically based on configuration:

```go
type Config struct {
    BuildProvider string `yaml:"build_provider"`
    TestProvider  string `yaml:"test_provider"`
}

func LoadNamespaces(config *Config) (*mage.NamespaceRegistry, error) {
    registry := mage.NewNamespaceRegistry()

    // Load build namespace based on config
    switch config.BuildProvider {
    case "fast":
        registry.Register("build", &FastBuild{})
    case "secure":
        registry.Register("build", &SecureBuild{})
    default:
        // Use default implementation
    }

    return registry, nil
}
```

## Available Mage Targets

The following mage targets are available through the exposed wrapper functions in `magefile.go`:

### Build Targets

| Target | Description | Equivalent Command |
|--------|-------------|-------------------|
| `mage build` | Build for current platform (default) | `go build` |
| `mage buildDefault` | Same as above | `go build` |
| `mage buildDocker` | Build Docker containers | `docker build` |
| `mage buildClean` | Clean build artifacts | `rm -rf bin/` |
| `mage buildGenerate` | Generate code before building | `go generate` |

### Test Targets

| Target | Description | Equivalent Command |
|--------|-------------|-------------------|
| `mage test` | Run complete test suite with linting | `make test` |
| `mage testDefault` | Same as above | `make test` |
| `mage testUnit` | Run unit tests only (no linting) | `go test ./...` |
| `mage testRace` | Run tests with race detector | `go test -race ./...` |
| `mage testCover` | Run tests with coverage analysis | `go test -cover ./...` |
| `mage testBench` | Run benchmark tests | `go test -bench=. ./...` |
| `mage testFuzz` | Run fuzz tests | `go test -fuzz=. ./...` |
| `mage testFuzzShort` | Run quick fuzz tests (5s) | `go test -fuzz=. -fuzztime=5s ./...` |
| `mage testIntegration` | Run integration tests | `go test -tags=integration ./...` |

### Dependency Management Targets

| Target | Description | Equivalent Command |
|--------|-------------|-------------------|
| `mage depsUpdate` | Update all dependencies | `make update` |
| `mage depsTidy` | Clean up go.mod and go.sum | `go mod tidy` |
| `mage depsDownload` | Download all dependencies | `go mod download` |
| `mage depsOutdated` | Show outdated dependencies | `go list -u -m all` |
| `mage depsAudit` | Audit dependencies for vulnerabilities | `govulncheck ./...` |

### Tools Management Targets

| Target | Description | Equivalent Command |
|--------|-------------|-------------------|
| `mage toolsUpdate` | Update all development tools | `make tools-update` |
| `mage toolsInstall` | Install all required development tools | `make tools` |
| `mage toolsCheck` | Check if all required tools are available | Tool availability check |

### Module Management Targets

| Target | Description | Equivalent Command |
|--------|-------------|-------------------|
| `mage modUpdate` | Update go.mod file | `go get -u ./...` |
| `mage modTidy` | Tidy the go.mod file | `go mod tidy` |
| `mage modVerify` | Verify module checksums | `go mod verify` |
| `mage modDownload` | Download modules | `go mod download` |

### Linting Targets

| Target | Description | Equivalent Command |
|--------|-------------|-------------------|
| `mage lint` | Run essential linters | `golangci-lint run && go vet` |
| `mage lintDefault` | Same as above | `golangci-lint run && go vet` |
| `mage lintAll` | Run all linting checks | `golangci-lint run && go vet && go fmt` |
| `mage lintFix` | Auto-fix linting issues + formatting | `golangci-lint run --fix && gofumpt/go fmt` |

### Documentation Targets

| Target | Description | Equivalent Command |
|--------|-------------|-------------------|
| `mage docsGenerate` | Generate documentation | `go doc` / `godoc` |
| `mage docsServe` | Serve documentation locally | `godoc -http=:6060` |
| `mage docsBuild` | Build static documentation | Static doc generation |
| `mage docsCheck` | Validate documentation | Doc validation |

### Git Workflow Targets

| Target | Description | Equivalent Command |
|--------|-------------|-------------------|
| `mage gitStatus` | Show git repository status | `git status` |
| `mage gitCommit` | Commit changes (interactive) | `git commit` |
| `mage gitTag` | Create and push a new tag | `git tag && git push --tags` |
| `mage gitPush` | Push changes to remote | `git push origin main` |

### Version Management Targets

| Target | Description | Equivalent Command |
|--------|-------------|-------------------|
| `mage versionShow` | Display current version information | Version display |
| `mage versionBump` | Bump the version (interactive) | Version increment |
| `mage versionCheck` | Check version information | Version validation |

### Release Targets

| Target | Description | Equivalent Command |
|--------|-------------|-------------------|
| `mage release` | Create a new release (default) | Release creation |
| `mage releaseDefault` | Same as above | Release creation |

### Installation Targets

| Target | Description | Equivalent Command |
|--------|-------------|-------------------|
| `mage installTools` | Install development tools | Tool installation |
| `mage installBinary` | Install the project binary | `go install` |
| `mage installStdlib` | Install Go stdlib for cross-compilation | `go install -a std` |
| `mage uninstall` | Remove installed binary | Binary removal |

### Metrics Targets

| Target | Description | Equivalent Command |
|--------|-------------|-------------------|
| `mage metricsLOC` | Analyze lines of code | Code analysis |
| `mage metricsCoverage` | Generate coverage reports | Coverage analysis |
| `mage metricsComplexity` | Analyze code complexity | Complexity analysis |

### Default Targets

| Target | Description | Notes |
|--------|-------------|-------|
| `mage` | Run default build | Same as `mage build` |
| `Default` | Default mage target | Points to `BuildDefault` |

### Command Discovery

| Command | Description |
|---------|-------------|
| `mage help` | List all available targets with beautiful formatting |
| `mage -l` | List all available targets (plain text) |
| `mage -h <target>` | Get help for specific target (if available) |
| `mage -version` | Show mage version |

## Performance Characteristics

| Operation | Time Complexity | Memory | Notes |
|-----------|----------------|--------|-------|
| Factory function call | O(1) | Constant | ~2ns overhead |
| Registry registration | O(1) | Per namespace | ~10ns |
| Registry lookup | O(1) | Constant | ~12ns |
| Interface method call | O(1) | None | Inlined by compiler |

## Versioning and Compatibility

### Interface Versioning

Interfaces follow semantic versioning principles:

- **Patch**: Implementation bug fixes (no interface changes)
- **Minor**: Adding new methods to interfaces (backward compatible)
- **Major**: Removing or changing existing methods (breaking changes)

### Backward Compatibility

The architecture maintains backward compatibility:

```go
// v1 style (always supported)
build := mage.Build{}
err := build.Default()

// v2 style (recommended)
build := mage.NewBuildNamespace()
err := build.Default()
```

### Forward Compatibility

New namespace methods are added to interfaces in a backward-compatible way using interface embedding or optional interfaces.
