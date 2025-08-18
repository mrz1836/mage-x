# CLAUDE.md - AI Assistant Guidelines for mage-x

## Project Overview

**mage-x** is a comprehensive build automation framework built on top of Magefile, designed to provide a complete set of build tools, workflows, and development operations for Go projects. The project provides both a library of reusable build tasks and advanced tooling for enterprise-level development.

### AI Agent Ecosystem

MAGE-X features a sophisticated ecosystem of **19 specialized AI agents** designed for intelligent development workflows:

- **🔧 Core Development (5)**: mage-x-builder, mage-x-linter, mage-x-deps, mage-x-docs, mage-x-security
- **🧪 Testing & Quality (2)**: mage-x-test-finder, mage-x-test-writer
- **🚀 Release & CI/CD (3)**: mage-x-releaser, mage-x-git, mage-x-gh
- **🏗️ Architecture & Performance (4)**: mage-x-architect, mage-x-refactor, mage-x-analyzer, mage-x-benchmark
- **🏢 Enterprise & Workflow (3)**: mage-x-enterprise, mage-x-workflow, mage-x-wizard
- **⚙️ Infrastructure (2)**: mage-x-config, mage-x-tools

**Key Features:**
- **Strategic Collaboration**: Agents coordinate intelligently for complex workflows
- **Parallel Execution**: Multiple agents work simultaneously for maximum efficiency
- **Security Integration**: All agents follow mage-x security-first patterns
- **Enterprise Scale**: Built for managing 30+ repositories with governance

See **[docs/SUB_AGENTS.md](../docs/SUB_AGENTS.md)** for complete agent documentation and collaboration patterns.

## Architecture

### Namespace Architecture
- **30+ namespace interfaces**: The project contains over 30 namespaces including:
  - **Core Namespaces (13)**: Build, Test, Lint, Tools, Deps, Mod, Docs, Git, Release, Metrics, Version, Install, Audit
  - **Advanced Namespaces (17+)**: Workflow, Enterprise, Wizard, Init, Releases, Help, CLI, Recipes, Bench, Yaml, Integrations, Vet, Update, Format, Configure, Generate, EnterpriseConfig
- **Interface-based design**: Each namespace has a corresponding interface (e.g., `BuildNamespace`)
- **Factory functions**: `NewBuildNamespace()`, `NewTestNamespace()`, etc.
- **Registry pattern**: `DefaultNamespaceRegistry` for centralized access
- **Flexible usage**: Both `Build{}` struct and interface-based approaches are supported

### Core Features
- **Build**: Build operations for Go projects (Default, All, Docker, Clean, Generate)
- **Test**: Comprehensive testing suite (Unit, Integration, Race, Coverage, Fuzz, Benchmarks)
- **Lint**: Code quality checks (golangci-lint, go vet, gofumpt)
- **Tools**: Development tool installation and management
- **Deps**: Dependency management (Update, Tidy, Audit, Outdated)
- **Mod**: Go module management
- **Docs**: Documentation generation and serving
- **Git**: Git operations (Status, Commit, Tag, Push)
- **Release**: Release management
- **Metrics**: Code analysis (LOC, Coverage, Complexity)
- **Version**: Version management (Show, Bump, Check)
- **Install**: Installation utilities (Tools, Binary, Stdlib)
- **Audit**: Security and compliance auditing

### Project Structure

```
mage-x/
├── pkg/mage/                          # Core mage package
│   ├── namespace_interfaces.go        # 18 namespace interfaces
│   ├── namespace_wrappers.go          # Interface implementations
│   ├── minimal_runner.go             # Command runner implementation
│   ├── build.go, test.go, lint.go... # Individual namespace implementations
│   ├── namespace_architecture_test.go # Architecture test
│   └── ...                           # Additional files
├── pkg/common/                       # Shared utilities
│   ├── env/                          # Environment management
│   ├── fileops/                      # File operations
│   ├── paths/                        # Path utilities
│   └── ...
├── pkg/security/                     # Security utilities
│   ├── command.go                    # Command executor interface
│   └── validator.go                  # Input validation
├── pkg/utils/                        # General utilities
│   └── ...
├── docs/                             # Comprehensive documentation
│   ├── NAMESPACE_INTERFACES.md       # User guide
│   ├── API_REFERENCE.md             # API documentation
│   └── NAMESPACE_EXAMPLES.md        # Usage examples
├── examples/                         # Extensive examples
│   ├── basic/                        # Basic usage examples
│   ├── custom/                       # Custom implementations
│   └── testing/                      # Testing patterns
└── .github/CLAUDE.md                 # This file
```

## Usage Patterns

### Basic Usage
```go
//go:build mage

package main

import "github.com/mrz1836/mage-x/pkg/mage"

var Default = mage.Build{}.Default

func Test() error {
    return mage.Test{}.Default()
}
```

### Interface-Based Usage
```go
//go:build mage

package main

import "github.com/mrz1836/mage-x/pkg/mage"

var Default = BuildDefault

func BuildDefault() error {
    build := mage.NewBuildNamespace()
    return build.Default()
}

func TestDefault() error {
    test := mage.NewTestNamespace()
    return test.Default()
}

// With aliases, users can now run:
// mage test (alias for mage testDefault)
// mage build (alias for mage buildDefault)
// mage lint (alias for mage lintDefault)
```

### Custom Implementation
```go
//go:build mage

package main

import "github.com/mrz1836/mage-x/pkg/mage"

type CustomBuild struct {
    mage.BuildNamespace
    // Custom fields
}

func (c CustomBuild) Default() error {
    // Custom build logic
    return nil
}

var Default = CustomBuild{}.Default
```

## Development Guidelines

### For AI Assistants

#### AI Agent Integration
Claude Code users have access to 19 specialized mage-x agents for comprehensive development workflows:
- **Single-Agent Tasks**: Use specialized agents like `mage-x-linter` for code quality or `mage-x-security` for security scanning
- **Multi-Agent Workflows**: Coordinate multiple agents for complex operations (e.g., "analyze code quality and write tests" triggers parallel execution)
- **Strategic Collaboration**: Agents automatically coordinate when their domains overlap
- **Parallel Optimization**: Multiple agents work simultaneously for maximum efficiency

#### Safe Operations
- Read any file in the project
- Modify namespace implementations in `pkg/mage/`
- Add new tests
- Update documentation
- Fix compilation issues
- Add new namespace methods
- Create examples
- Run architecture tests
- Coordinate with specialized mage-x agents for comprehensive workflows

#### Best Practices
- Always test with `go test ./pkg/mage/namespace_architecture_test.go -v`
- Verify compilation with `go build ./pkg/mage`
- Use factory functions: `NewBuildNamespace()`, etc.
- Maintain interface contracts
- Follow existing code patterns
- Use proper error handling with `fmt.Errorf`
- Add comments for exported functions
- Ensure cross-platform compatibility

#### Security Considerations
- The project includes security utilities in `pkg/security/`
- Command execution should be validated when implementing new features
- Input validation is available via security validators
- Be mindful of path traversal and command injection risks

#### Dangerous Operations (Avoid)
- Don't modify interface definitions without careful consideration
- Don't break existing public methods
- Don't remove existing public methods
- Don't modify the registry pattern without testing
- Test changes thoroughly before committing

#### Adding New Namespaces
1. Add interface to `namespace_interfaces.go`
2. Create implementation file (e.g., `mynew.go`)
3. Add wrapper in `namespace_wrappers.go`
4. Add factory function `NewMyNewNamespace()`
5. Add to registry in `namespace_interfaces.go`
6. Add tests
7. Update documentation

## Quick Commands

### Run Tests
```bash
go test ./pkg/mage/namespace_architecture_test.go -v
```

### Check Compilation
```bash
go build ./pkg/mage
```

### Run Specific Namespace
```bash
# Through factory function
go run -tags mage . build
```

### Clean Imports
```bash
goimports -w pkg/mage/*.go
```

### Run Core Package Tests
```bash
go test ./pkg/common/env ./pkg/common/fileops ./pkg/common/paths -v
```

### Run All Tests
```bash
go test ./... -v
```

### Run Tests with Race Detection
```bash
go test ./... -race
```

## Architecture Features

The mage-x architecture supports:

- **Flexibility**: Interface-based design allows custom implementations
- **Testability**: Mock interfaces for unit testing
- **Maintainability**: Clean separation of concerns
- **Dual Approach**: Both struct-based and interface-based usage patterns
- **Extensibility**: Easy to add new namespaces and methods
- **Production Ready**: Core functionality working with real tool execution
- **Security**: Built-in security utilities for command validation
- **Cross-Platform**: Works on Windows, macOS, and Linux

## Common Patterns

### Error Handling
```go
func (namespace Namespace) Task() error {
    if err := someOperation(); err != nil {
        return fmt.Errorf("task failed: %w", err)
    }
    return nil
}
```

### Configuration Access
```go
func (namespace Namespace) Task() error {
    cfg, err := GetConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }

    // Use configuration
    if cfg.SomeOption {
        // Do something
    }

    return nil
}
```

### Command Execution
```go
func (namespace Namespace) Task() error {
    runner := GetRunner()
    if err := runner.RunCmd("go", "build", "./..."); err != nil {
        return fmt.Errorf("build failed: %w", err)
    }
    return nil
}
```

## Complete Command Reference

### Exposed Commands (59 total)

The main `magefile.go` exposes 59 commands that can be called directly via `mage`:

#### Default & Aliases
- `mage` - Default command (runs buildDefault)
- `Default` - Maps to BuildDefault

**Available Aliases:**
- `mage lint` - Alias for `mage lintDefault`
- `mage test` - Alias for `mage testDefault`
- `mage build` - Alias for `mage buildDefault`
- `mage release` - Alias for `mage releaseDefault`
- `mage loc` - Alias for `mage metricsLOC`
- `mage docs` - Alias for `mage docsDefault`

#### Build Commands (5)
- `buildDefault` - Build for current platform
- `buildAll` - Build for all configured platforms
- `buildClean` - Clean build artifacts
- `buildDocker` - Build Docker containers
- `buildGenerate` - Generate code before building

#### Test Commands (12)
- `testDefault` - Run complete test suite with linting
- `testFull` - Run full test suite with linting
- `testUnit` - Run unit tests only
- `testShort` - Run short tests
- `testRace` - Run tests with race detector
- `testCover` - Run tests with coverage
- `testCoverRace` - Run tests with coverage and race detector
- `testBench` - Run benchmark tests
- `testBenchShort` - Run short benchmark tests
- `testFuzz` - Run fuzz tests
- `testFuzzShort` - Run quick fuzz tests
- `testIntegration` - Run integration tests

#### Lint Commands (6)
- `lintDefault` - Run essential linters
- `lintAll` - Run all linting checks
- `lintFix` - Auto-fix linting issues
- `lintVet` - Run go vet static analysis
- `lintFumpt` - Run gofumpt code formatting
- `lintVersion` - Show linter version

#### Dependency Commands (5)
- `depsUpdate` - Update all dependencies
- `depsTidy` - Clean up go.mod and go.sum
- `depsDownload` - Download all dependencies
- `depsOutdated` - Show outdated dependencies
- `depsAudit` - Audit dependencies for vulnerabilities

#### Tools Commands (4)
- `toolsUpdate` - Update development tools
- `toolsInstall` - Install required tools
- `toolsCheck` - Check if tools are available
- `toolsVulnCheck` - Run vulnerability check

#### Module Commands (4)
- `modUpdate` - Update go.mod file
- `modTidy` - Tidy go.mod file
- `modVerify` - Verify module checksums
- `modDownload` - Download modules

#### Documentation Commands (5)
- `docsDefault` - Generate and serve documentation (generate + serve combined)
- `docsGenerate` - Generate Go package documentation from source code
- `docsServe` - Serve documentation locally with hybrid pkgsite/godoc support
- `docsBuild` - Build enhanced static documentation files with metadata
- `docsCheck` - Validate documentation completeness and quality

#### Git Commands (4)
- `gitStatus` - Show repository status
- `gitCommit` - Commit changes (requires `message="commit message"` env var)
- `gitTag` - Create and push tag (requires `version="1.2.3"` env var)
- `gitPush` - Push changes to remote

#### Version Commands (3)
- `versionShow` - Display version information
- `versionBump` - Bump the version
- `versionCheck` - Check version information

#### Metrics Commands (3)
- `metricsLOC` - Analyze lines of code
- `metricsCoverage` - Generate coverage reports
- `metricsComplexity` - Analyze code complexity

#### Install Commands (4)
- `installTools` - Install development tools
- `installBinary` - Install project binary
- `installStdlib` - Install Go standard library
- `uninstall` - Remove installed binary

#### Other Commands (4)
- `releaseDefault` - Create a new release
- `auditShow` - Display audit events
- `help` - Beautiful command listing
- `list` - Alternative command listing

### Additional Namespace Methods

While the main magefile exposes 59 commands, the pkg/mage package contains 30+ namespaces with hundreds of additional methods. These can be accessed by creating custom magefiles. Some examples:

- **Audit**: Stats(), Export(), Cleanup(), Enable(), Disable(), Report()
- **Build**: Linux(), Darwin(), Windows(), Platform(), PreBuild(), Install()
- **Test**: CI(), CINoRace(), CoverReport(), CoverHTML(), Parallel(), NoLint()
- **Format**: All(), Check(), Fix(), Go(), JSON(), Markdown(), SQL(), Shell(), YAML()
- **Workflow**: Create(), Execute(), History(), Schedule(), Status(), Template()
- **Enterprise**: Backup(), Deploy(), Promote(), Restore(), Rollback()
- **Recipes**: Create(), Install(), List(), Run(), Search(), Show()
- **Generate**: All(), Code(), Config(), Docs(), GraphQL(), Mocks(), OpenAPI()

## Contributing

When contributing to mage-x:

1. Follow the existing code style
2. Add tests for new functionality
3. Update documentation as needed
4. Ensure all tests pass
5. Run linting before submitting

## Additional Resources

- **[AI Agent Ecosystem](../docs/SUB_AGENTS.md)** - Complete documentation of all 19 specialized agents
- See examples in `examples/` directory for practical usage
- Check `docs/` for detailed documentation
- Review existing namespace implementations for patterns
- Use the test files as references for testing approaches

The project provides a modern, flexible build automation system for Go projects with comprehensive testing demonstrating successful execution of build operations, tool management, dependency handling, and code analysis across the entire interface architecture.

### AI Agent Workflow Examples

#### Multi-Agent Parallel Execution
```bash
# Quality assurance workflow - triggers 3 agents in parallel
"Analyze code quality, identify test gaps, and validate security"
# Activates: mage-x-linter + mage-x-test-finder + mage-x-security

# Performance optimization workflow
"Optimize code performance and validate with benchmarks"
# Activates: mage-x-refactor + mage-x-benchmark + mage-x-analyzer

# Release preparation workflow
"Prepare production release with full validation"
# Activates: mage-x-security + mage-x-docs + mage-x-releaser (coordinated sequence)
```

#### Strategic Agent Coordination
Agents automatically coordinate when their domains overlap:
- **mage-x-builder** automatically invokes **mage-x-linter** and **mage-x-security** for pre-build validation
- **mage-x-test-writer** coordinates with **mage-x-benchmark** for performance test creation
- **mage-x-releaser** works with **mage-x-git** and **mage-x-gh** for complete release workflows

## Documentation System

### Hybrid Documentation Server

MAGE-X includes a sophisticated hybrid documentation system with enterprise-grade capabilities:

- **Smart Tool Detection**: Automatically detects and uses the best available documentation tool (`pkgsite` or `godoc`)
- **Auto-Installation**: Automatically installs missing documentation tools when needed
- **Multiple Serving Modes**: Supports project-only, stdlib-only, or comprehensive documentation serving
- **Port Management**: Automatic port detection and conflict resolution
- **Cross-Platform Browser Opening**: Native browser launching on macOS, Linux, and Windows
- **CI/CD Ready**: Detects CI environments and adjusts behavior accordingly
- **Configuration Control**: Override tool selection with `docs.tool: "pkgsite"` and `docs.port: 8080` in .mage.yaml

### Available Documentation Commands

- `mage docs` - **NEW**: Generate and serve documentation in one command (generate + serve)
- `mage docsGenerate` - Generate comprehensive Go package documentation
- `mage docsServe` - Serve documentation with smart tool detection and fallback
- `mage docsBuild` - Build enhanced static documentation files with metadata and navigation
- `mage docsCheck` - Validate documentation completeness and quality

### Advanced Serving Options

The documentation system provides specialized serving methods (not exposed as top-level commands):

- `ServePkgsite()` - Force modern pkgsite server (project-focused)
- `ServeGodoc()` - Force classic godoc server (comprehensive)
- `ServeStdlib()` - Standard library documentation only
- `ServeProject()` - Project documentation only
- `ServeBoth()` - Both project and standard library documentation

### Documentation Generation Features

- **Package Discovery**: Automatically discovers all Go packages in the project
- **Categorization**: Groups packages by type (Core, Common, Providers, Security, Utils, Commands, Other)
- **Comprehensive Coverage**: Documents all public APIs, functions, types, and methods
- **Markdown Output**: Generates well-formatted markdown documentation
- **Build System**: Creates enhanced static documentation with metadata, navigation, and cross-references
