# MAGE-X Quick Setup Guide

## Creating Your MAGE-X Repository

### 1. Initialize the repository

```bash
# Create and enter the directory
mkdir MAGE-X
cd MAGE-X

# Initialize git
git init

# Initialize go module (replace with your username/org)
go mod init github.com/mrz1836/mage-x
```

### 2. Create the directory structure

```bash
# Create directories
mkdir -p pkg/mage pkg/utils templates examples .github/workflows

# Create go.mod
go get github.com/magefile/mage@v1.15.0
go get gopkg.in/yaml.v3@v3.0.1
```

### 3. Add the library files

Copy all the artifact files to their respective locations:
- `pkg/mage/config.go`
- `pkg/mage/build.go`
- `pkg/mage/test.go`
- `pkg/mage/lint.go`
- `pkg/mage/tools.go`
- `pkg/mage/deps.go`
- `pkg/utils/utils.go` (combine all utility functions)
- `.github/AGENTS.md`
- `README.md`

### 4. Create a test magefile

Create `magefile.go` in the root:

```go
//go:build mage
// +build mage

package main

import (
    _ "github.com/mrz1836/mage-x/pkg/mage"
)

var Default = Test
```

### 5. Test the library

```bash
# Install mage
go install github.com/magefile/mage@latest

# List available tasks
mage -l

# Run tests
mage test
```

## Using MAGE-X in Your Projects

### 1. Add to your project

```bash
go get github.com/mrz1836/mage-x
```

### 2. Create magefile.go

```go
//go:build mage
// +build mage

package main

import (
    _ "github.com/mrz1836/mage-x/pkg/mage"
)

var Default = Build
```

### 3. (Optional) Add configuration

Create `.mage.yaml`:

```yaml
project:
  name: myapp
  binary: myapp

build:
  platforms:
    - linux/amd64
    - darwin/arm64
```

### 4. Run tasks

```bash
# Build (default)
mage

# Complete test suite with linting  
mage test

# Unit tests only (no linting)
mage testUnit

# Test with coverage analysis
mage testCover

# Run linter
mage lint

# Fix linting issues
mage lintFix

# Update dependencies (equivalent to "make update")
mage depsUpdate

# Install development tools
mage toolsInstall
```

## Key Features

### Zero Configuration
Works immediately with sensible defaults:
```bash
mage                # Build for current platform (default)
mage test           # Run complete test suite with linting
mage testUnit       # Run unit tests only
mage lint           # Run linter with golangci-lint
mage depsUpdate     # Update all dependencies
mage toolsInstall   # Install development tools
```

### Easy Customization
Add custom tasks alongside imported ones:
```go
func Deploy() error {
    mg.Deps(Build)
    // Your deployment logic
    return nil
}
```

### Comprehensive Command Set
All development operations covered:
```bash
# Build operations
mage buildDocker    # Build Docker containers  
mage buildClean     # Clean build artifacts

# Test operations
mage testRace       # Test with race detector
mage testBench      # Run benchmarks
mage testFuzz       # Run fuzz tests

# Development tools
mage toolsUpdate    # Update all tools
mage toolsCheck     # Verify tools available

# Dependency management
mage depsTidy       # Clean go.mod and go.sum
mage depsAudit      # Check for vulnerabilities

# Documentation
mage docsGenerate   # Generate documentation
mage docsServe      # Serve docs locally

# Version control
mage gitStatus      # Show git status
mage versionShow    # Display version info
```

### CI/CD Ready
```yaml
# .github/workflows/ci.yml
- name: Run CI
  run: |
    mage test          # Run complete test suite
    mage lint          # Run linter
    mage depsAudit     # Check for vulnerabilities
```

### Complete Command Reference

For a complete list of available commands, run:
```bash
mage -l
```

**Core Commands:**
- `mage` - Build for current platform
- `mage test` - Complete test suite with linting  
- `mage lint` - Run linter
- `mage depsUpdate` - Update dependencies
- `mage toolsInstall` - Install development tools

**All Available Commands:**
```bash
# Build
buildClean, buildDefault, buildDocker, buildGenerate

# Dependencies  
depsAudit, depsDownload, depsOutdated, depsTidy, depsUpdate

# Documentation
docsBuild, docsCheck, docsGenerate, docsServe

# Git Operations
gitCommit, gitPush, gitStatus, gitTag

# Installation
installBinary, installStdlib, installTools, uninstall

# Linting
lint, lintAll, lintFix

# Metrics
metricsComplexity, metricsCoverage, metricsLOC

# Module Management
modDownload, modTidy, modUpdate, modVerify

# Releases
release, releaseDefault

# Testing
test, testBench, testCover, testFuzz, testIntegration, testRace, testUnit

# Tools
toolsCheck, toolsInstall, toolsUpdate

# Version Management
versionBump, versionCheck, versionShow
```

## Best Practices

1. **Start Simple**: Use default configuration first
2. **Add Config As Needed**: Only configure what differs from defaults
3. **Keep Tasks Fast**: Use parallelism where possible
4. **Document Custom Tasks**: Add comments for clarity
5. **Version Your Tools**: Pin tool versions in .mage.yaml

## Common Patterns

### Multi-stage builds
```go
func Release() error {
    mg.SerialDeps(
        Test,
        mg.Namespace(Build{}).All,
        mg.Namespace(Release{}).Tag,
    )
    return nil
}
```

### Environment-specific builds
```go
func Production() error {
    os.Setenv("GO_BUILD_TAGS", "prod")
    return mg.Namespace(Build{}).Default()
}
```

### Custom tool integration
```yaml
tools:
  custom:
    migrate: github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Troubleshooting

### Tasks not showing up
- Ensure `//go:build mage` tag is present
- Check for compilation errors: `go build magefile.go`

### Import errors
- Run `go mod tidy`
- Ensure Go 1.24+ is installed

### Configuration not working
- Check file is named `.mage.yaml` (with dot)
- Validate YAML syntax

## Next Steps

1. Explore the [examples](../examples) directory
2. Read the full [README](../README.md)
3. Check [AGENTS.md](../.github/AGENTS.md) for AI assistant integration
4. Contribute improvements via pull requests

Happy building with Mage! 🪄
