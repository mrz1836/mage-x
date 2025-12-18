# CLAUDE.md - AI Assistant Guidelines for mage-x

## Project Overview
**mage-x** is a comprehensive build automation framework for Go projects, providing 400+ built-in commands through the `magex` binary. It features 19 specialized AI agents for intelligent development workflows, interface-based architecture with 30+ namespaces, and zero-configuration usage.

## Quick Start
```bash
magex build          # Build for current platform
magex test           # Run complete test suite
magex lint           # Check code quality
magex format:fix     # Auto-fix code formatting
magex -l             # List all 400+ commands
```

## AI Agent Ecosystem (19 Specialized Agents)
- **Core Development**: builder, linter, deps, docs, security
- **Testing & Quality**: test-finder, test-writer, benchmark
- **Release & CI/CD**: releaser, git, gh (GitHub operations)
- **Architecture**: architect, refactor, analyzer, tools
- **Infrastructure**: config, workflow, wizard

Agents coordinate automatically and execute in parallel when domains overlap. See `docs/SUB_AGENTS.md` for details.

## Architecture

### Project Structure
```
mage-x/
├── pkg/mage/                # Core with 30+ namespace implementations
├── pkg/common/              # Shared utilities (env, fileops, paths)
├── pkg/security/            # Command validation and security
├── pkg/utils/               # General utilities
├── docs/                    # Documentation
└── examples/                # Usage examples
```

### Key Namespaces (30+)
bench, build, configure, deps, docs, format, generate, git, help, install, lint, metrics, mod, release, test, tools, version, vet, yaml

Each namespace has an interface (e.g., `BuildNamespace`) and factory function (e.g., `NewBuildNamespace()`).

## Essential Commands

### Build & Development
```bash
magex build                  # Build for current platform
magex build:all             # Build for all platforms
magex build:prebuild        # Memory-efficient cache warming
magex clean                 # Remove build artifacts
```

### Testing
```bash
magex test                  # Run full test suite with linting
magex test:unit            # Unit tests only
magex test:cover           # Tests with coverage
magex test:race            # Tests with race detector
magex test:fuzz            # Run fuzz tests
```

### Code Quality
```bash
magex lint                  # Run essential linters
magex lint:fix             # Auto-fix linting issues
magex format:fix           # Fix code formatting
magex vet                  # Static analysis
```

### Dependencies
```bash
magex deps:update          # Update all dependencies
magex deps:tidy           # Clean go.mod and go.sum
magex deps:audit          # Security vulnerability check
```

### Documentation
```bash
magex docs                 # Generate and serve docs
magex docs:serve          # Serve with auto-detection (pkgsite/godoc)
magex docs:check          # Validate documentation
```

### Git & Release
```bash
magex git:status          # Repository status
magex git:commit          # Commit changes (needs message="...")
magex git:tag            # Create tag (needs version="...")
magex release            # Create new release
```

### Tools & Metrics
```bash
magex tools:install       # Install dev tools
magex metrics:loc        # Lines of code analysis
magex metrics:coverage   # Coverage reports
```

## Development Guidelines

### For AI Assistants

#### Safe Operations
- Read/modify files in `pkg/mage/`
- Add tests and examples
- Fix compilation issues
- Run `magex` commands
- Create custom magefiles
- Use specialized agents for complex tasks

#### Command Format
- Use kebab-case: `magex test:unit` NOT `magex testUnit`
- Simple aliases available: `test`, `build`, `lint`, `clean`
- Full format: `namespace:action`

#### Testing Commands
```bash
go test ./pkg/mage/namespace_architecture_test.go -v  # Architecture test
go build ./pkg/mage                                   # Check compilation
go test ./... -v                                      # All tests
```

#### Security
- Command validation in `pkg/security/`
- Never expose secrets or credentials
- Validate all inputs

## Quick Examples

### Custom Implementation
```go
//go:build mage
package main

import "github.com/mrz1836/mage-x/pkg/mage"

// Use built-in namespace
var Default = mage.Build{}.Default

// Custom command
func Deploy() error {
    // Custom logic
    return nil
}
```

### Interface-Based Usage
```go
func BuildDefault() error {
    build := mage.NewBuildNamespace()
    return build.Default()
}
```

## Enhanced Pre-Build Strategies

### Memory-Efficient Build Cache Warming
```bash
magex build:prebuild                     # Smart auto-detection
magex build:prebuild strategy=incremental batch=5  # Small batches
magex build:prebuild strategy=mains-first # Prioritize main packages
```

Configuration via `.mage.yaml`:
```yaml
build:
  prebuild:
    strategy: incremental
    batch_size: 10
    memory_limit: "4G"
```

## Additional Resources
- `docs/SUB_AGENTS.md` - AI agent documentation
- `docs/NAMESPACE_INTERFACES.md` - Architecture guide
- `examples/` - Usage examples
- Run `magex -l` or `magex help` for all available commands
