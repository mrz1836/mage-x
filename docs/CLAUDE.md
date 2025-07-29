# CLAUDE.md - AI Assistant Guidelines for go-mage

## Project Overview

**go-mage** is a comprehensive build automation framework built on top of Magefile, designed to provide a complete set of build tools, workflows, and development operations for Go projects. The project provides both a library of reusable build tasks and advanced tooling for enterprise-level development.

## Architecture

### Namespace Architecture
- **37 total namespaces**: Build, Test, Lint, Format, Deps, Git, Release, Docs, Deploy, Tools, Security, Generate, CLI, Update, Mod, Recipes, Metrics, Workflow, Database, Common, Operations, Version, Vault, Bench, Install, Audit, Analytics, VCS, Interactive, Vet, Init, Configure, YAML, Releases, Enterprise
- **Interface-based design**: Each namespace has a corresponding interface (e.g., `BuildNamespace`)
- **Factory functions**: `NewBuildNamespace()`, `NewTestNamespace()`, etc.
- **Registry pattern**: `DefaultNamespaceRegistry` for centralized access
- **Backward compatibility**: Existing `Build{}`, `Test{}` types still work

### Core Features
- **Build**: Build operations for Go projects
- **Test**: Comprehensive testing with linting and test suites  
- **Tools**: Development tool installation (golangci-lint, gofumpt, govulncheck, mockgen, swag)
- **Deps**: Dependency management
- **Mod**: Go module management
- **Metrics**: Code analysis and statistics
- **Generate**: Code generation
- **Lint**: Code quality checks
- **Update**: Update checking

### Project Structure

```
go-mage/
├── pkg/mage/                          # Core mage package
│   ├── namespace_interfaces.go        # 37 namespace interfaces
│   ├── namespace_wrappers.go          # Interface implementations
│   ├── minimal_runner.go             # Command runner implementation
│   ├── build.go, test.go, lint.go... # Individual namespace implementations
│   └── *.go.disabled                 # Temporarily disabled files
├── pkg/common/                       # Shared utilities
│   ├── env/                          # Environment management
│   ├── fileops/                      # File operations
│   ├── paths/                        # Path utilities
│   └── ...
├── docs/                             # Comprehensive documentation
│   ├── NAMESPACE_INTERFACES.md       # User guide
│   ├── API_REFERENCE.md             # API documentation
│   └── NAMESPACE_EXAMPLES.md        # Usage examples
├── examples/                         # Extensive examples
│   ├── basic/                        # Basic usage examples
│   ├── custom/                       # Custom implementations
│   └── testing/                      # Testing patterns
├── tests/                            # Test files
│   └── namespace_architecture_test.go # Main architecture test
├── scripts/                          # Shell scripts
│   ├── setup.sh                      # Setup script
│   └── test-runner.sh                # Test runner
└── docs/CLAUDE.md                    # This file
```

## Usage Patterns

### Basic Usage (Backward Compatible)
```go
//go:build mage

package main

import "github.com/mrz1836/go-mage/pkg/mage"

var Default = mage.Build{}.Default

func Test() error {
    return mage.Test{}.Default()
}
```

### Interface-Based Usage (New Pattern)
```go
//go:build mage

package main

import "github.com/mrz1836/go-mage/pkg/mage"

var Default = BuildDefault

func BuildDefault() error {
    build := mage.NewBuildNamespace()
    return build.Default()
}

func TestDefault() error {
    test := mage.NewTestNamespace()
    return test.Default()
}
```

### Custom Implementation
```go
//go:build mage

package main

import "github.com/mrz1836/go-mage/pkg/mage"

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

#### Safe Operations
- Read any file in the project
- Modify namespace implementations in `pkg/mage/`
- Add new tests
- Update documentation
- Fix compilation issues
- Add new namespace methods
- Create examples
- Run architecture tests

#### Best Practices
- Always test with `go test ./tests/namespace_architecture_test.go -v`
- Verify compilation with `go build ./pkg/mage`
- Use factory functions: `NewBuildNamespace()`, etc.
- Maintain interface contracts
- Preserve backward compatibility

#### Dangerous Operations (Avoid)
- Don't modify interface definitions without careful consideration
- Don't break backward compatibility
- Don't remove existing public methods
- Don't modify the registry pattern without testing
- Avoid touching disabled files

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
go test ./tests/namespace_architecture_test.go -v
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

## Architecture Features

The go-mage architecture supports:

- **Flexibility**: Interface-based design allows custom implementations
- **Testability**: Mock interfaces for unit testing  
- **Maintainability**: Clean separation of concerns
- **Backward Compatibility**: Existing code continues to work
- **Extensibility**: Easy to add new namespaces and methods
- **Production Ready**: Core functionality working with real tool execution

The project provides a modern, flexible build automation system for Go projects with comprehensive testing demonstrating successful execution of build operations, tool management, dependency handling, and code analysis across the entire interface architecture.