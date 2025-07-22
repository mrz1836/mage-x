# Basic MAGE-X Example

This is the simplest possible example of using MAGE-X.

## Setup

1. Add MAGE-X to your project:
```bash
go get github.com/mrz1836/go-mage
```

2. Create a `magefile.go` (see the example in this directory)

3. Run tasks:
```bash
# List all available tasks
mage -l

# Build (default)
mage

# Run tests
mage test

# Run linter
mage lint

# Build for all platforms
mage build:all
```

## What You Get

With just 3 lines of code, you get:

### Build Tasks
- `build` - Build for current platform
- `build:all` - Build for all platforms
- `build:linux`, `build:darwin`, `build:windows` - Platform-specific builds
- `build:docker` - Build Docker image
- `build:clean` - Clean build artifacts
- `build:install` - Install to $GOPATH/bin

### Test Tasks
- `test` - Run tests with linting
- `test:unit` - Unit tests only
- `test:race` - Tests with race detector
- `test:cover` - Tests with coverage
- `test:bench` - Run benchmarks
- `test:fuzz` - Run fuzz tests

### Lint Tasks
- `lint` - Run golangci-lint
- `lint:fix` - Auto-fix issues
- `lint:fmt` - Format code
- `lint:vet` - Run go vet

### Dependency Tasks
- `deps:download` - Download dependencies
- `deps:update` - Update all dependencies
- `deps:tidy` - Clean up go.mod
- `deps:vulncheck` - Check for vulnerabilities

### Tool Tasks
- `tools:install` - Install dev tools
- `tools:verify` - Verify tools are installed

## Customization

To add custom tasks, see the [custom tasks example](../custom).
