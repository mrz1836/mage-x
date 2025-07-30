# CLAUDE.md - AI Assistant Guidelines for go-mage

## Project Overview

**go-mage** is a comprehensive build automation framework built on top of Magefile, designed to provide a complete set of build tools, workflows, and development operations for Go projects. The project provides both a library of reusable build tasks and advanced tooling for enterprise-level development.

## Architecture

### Namespace Architecture
- **18 namespace interfaces**: Build, Test, Lint, Format, Deps, Git, Release, Docs, Deploy, Tools, Security, Generate, CLI, Update, Mod, Recipes, Metrics, Workflow
- **Interface-based design**: Each namespace has a corresponding interface (e.g., `BuildNamespace`)
- **Factory functions**: `NewBuildNamespace()`, `NewTestNamespace()`, etc.
- **Registry pattern**: `DefaultNamespaceRegistry` for centralized access
- **Flexible usage**: Both `Build{}` struct and interface-based approaches are supported

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
- **Security**: Security scanning and validation
- **Workflow**: Advanced workflow operations

### Project Structure

```
go-mage/
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
├── scripts/                          # Shell scripts
│   ├── setup.sh                      # Setup script
│   └── test-runner.sh                # Test runner
└── .github/CLAUDE.md                 # This file
```

## Usage Patterns

### Basic Usage
```go
//go:build mage

package main

import "github.com/mrz1836/go-mage/pkg/mage"

var Default = mage.Build{}.Default

func Test() error {
    return mage.Test{}.Default()
}
```

### Interface-Based Usage
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

The go-mage architecture supports:

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
    cfg, err := LoadConfig()
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

## Contributing

When contributing to go-mage:

1. Follow the existing code style
2. Add tests for new functionality
3. Update documentation as needed
4. Ensure all tests pass
5. Run linting before submitting

## Additional Resources

- See examples in `examples/` directory for practical usage
- Check `docs/` for detailed documentation
- Review existing namespace implementations for patterns
- Use the test files as references for testing approaches

The project provides a modern, flexible build automation system for Go projects with comprehensive testing demonstrating successful execution of build operations, tool management, dependency handling, and code analysis across the entire interface architecture.