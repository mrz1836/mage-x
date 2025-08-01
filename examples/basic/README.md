# Basic MAGE-X Example

This is the simplest possible example of using MAGE-X.

## Setup

1. Add MAGE-X to your project:
```bash
go get github.com/mrz1836/mage-x
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
mage buildAll
```

## What You Get

With just 3 lines of code, you get:

### Build Tasks
- `build` - Build for current platform
- `buildAll` - Build for all platforms
- `buildLinux`, `buildDarwin`, `buildWindows` - Platform-specific builds
- `buildDocker` - Build Docker image
- `buildClean` - Clean build artifacts
- `buildInstall` - Install to $GOPATH/bin

### Test Tasks
- `test` - Run tests with linting
- `testUnit` - Unit tests only
- `testRace` - Tests with race detector
- `testCover` - Tests with coverage
- `testBench` - Run benchmarks
- `testFuzz` - Run fuzz tests

### Lint Tasks
- `lint` - Run golangci-lint
- `lintFix` - Auto-fix issues
- `lintFmt` - Format code
- `lintVet` - Run go vet

### Dependency Tasks
- `depsDownload` - Download dependencies
- `depsUpdate` - Update all dependencies
- `depsTidy` - Clean up go.mod
- `depsVulncheck` - Check for vulnerabilities

### Tool Tasks
- `toolsInstall` - Install dev tools
- `toolsVerify` - Verify tools are installed

## Customization

To add custom tasks, see the [custom tasks example](../custom).
