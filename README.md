<div align="center">

# 🪄&nbsp;&nbsp;MAGE-X

**Write Once, Mage Everywhere: Zero-Boilerplate Build Automation for Go.**

<br/>

<a href="https://github.com/mrz1836/mage-x/releases"><img src="https://img.shields.io/github/release-pre/mrz1836/mage-x?include_prereleases&style=flat-square&logo=github&color=black" alt="Release"></a>
<a href="https://golang.org/"><img src="https://img.shields.io/github/go-mod/go-version/mrz1836/mage-x?style=flat-square&logo=go&color=00ADD8" alt="Go Version"></a>
<a href="https://github.com/mrz1836/mage-x/blob/master/LICENSE"><img src="https://img.shields.io/github/license/mrz1836/mage-x?style=flat-square&color=blue" alt="License"></a>

<br/>

<table align="center" border="0">
  <tr>
    <td align="right">
       <code>CI / CD</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://github.com/mrz1836/mage-x/actions"><img src="https://img.shields.io/github/actions/workflow/status/mrz1836/mage-x/fortress.yml?branch=master&label=build&logo=github&style=flat-square" alt="Build"></a>
       <a href="https://github.com/mrz1836/mage-x/commits/master"><img src="https://img.shields.io/github/last-commit/mrz1836/mage-x?style=flat-square&logo=git&logoColor=white&label=last%20update" alt="Last Commit"></a>
    </td>
    <td align="right">
       &nbsp;&nbsp;&nbsp;&nbsp; <code>Quality</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://goreportcard.com/report/github.com/mrz1836/mage-x"><img src="https://goreportcard.com/badge/github.com/mrz1836/mage-x?style=flat-square" alt="Go Report"></a>
       <a href="https://app.codecov.io/gh/mrz1836/mage-x/tree/master"><img src="https://codecov.io/gh/mrz1836/mage-x/branch/master/graph/badge.svg?style=flat-square" alt="Coverage"></a>
    </td>
  </tr>

  <tr>
    <td align="right">
       <code>Security</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://scorecard.dev/viewer/?uri=github.com/mrz1836/mage-x"><img src="https://api.scorecard.dev/projects/github.com/mrz1836/mage-x/badge?style=flat-square" alt="Scorecard"></a>
       <a href=".github/SECURITY.md"><img src="https://img.shields.io/badge/policy-active-success?style=flat-square&logo=security&logoColor=white" alt="Security"></a>
    </td>
    <td align="right">
       &nbsp;&nbsp;&nbsp;&nbsp; <code>Community</code> &nbsp;&nbsp;
    </td>
    <td align="left">
       <a href="https://github.com/mrz1836/mage-x/graphs/contributors"><img src="https://img.shields.io/github/contributors/mrz1836/mage-x?style=flat-square&color=orange" alt="Contributors"></a>
       <a href="https://github.com/sponsors/mrz"><img src="https://img.shields.io/badge/sponsor-mrz-181717.svg?logo=github&style=flat-square" alt="Sponsor"></a>
    </td>
  </tr>
</table>

</div>

<br/>
<br/>

<div align="center">

### <code>Project Navigation</code>

</div>

<table align="center">
  <tr>
    <td align="center" width="25%">
       🧩&nbsp;<a href="#-whats-inside"><code>What's&nbsp;Inside</code></a>
    </td>
    <td align="center" width="25%">
       ⚡&nbsp;<a href="#-quick-start"><code>Quick&nbsp;Start</code></a>
    </td>
    <td align="center" width="25%">
       🚀&nbsp;<a href="#-features"><code>Features</code></a>
    </td>
    <td align="center" width="25%">
       ⚙️&nbsp;<a href="#-configuration"><code>Configuration</code></a>
    </td>
  </tr>
  <tr>
    <td align="center">
       📚&nbsp;<a href="#-documentation"><code>Documentation</code></a>
    </td>
    <td align="center">
       📐&nbsp;<a href="#-spec-driven-development"><code>SDD</code></a>
    </td>
    <td align="center">
       🧪&nbsp;<a href="#-examples--tests"><code>Examples&nbsp;&&nbsp;Tests</code></a>
    </td>
    <td align="center">
       ⚡&nbsp;<a href="#-benchmarks"><code>Benchmarks</code></a>
    </td>
  </tr>
  <tr>
    <td align="center">
       🛠️&nbsp;<a href="#-code-standards"><code>Code&nbsp;Standards</code></a>
    </td>
    <td align="center">
      🤖&nbsp;<a href="#-ai-usage--assistant-guidelines"><code>AI&nbsp;Usage</code></a>
    </td>
    <td align="center">
       🤝&nbsp;<a href="#-contributing"><code>Contributing</code></a>
    </td>
    <td align="center">
       📝&nbsp;<a href="#-license"><code>License</code></a>
    </td>
  </tr>
</table>

<br/>

## 🧩 What's Inside

**MAGE-X** revolutionizes Go build automation with TRUE zero-boilerplate. Unlike traditional Mage which requires writing wrapper functions, MAGE-X provides all commands instantly through the `magex` binary.

<br/>

- **Truly Zero-Configuration**<br/>
  _No magefile.go needed. No imports. No wrappers. Just install `magex` and all built-in and custom commands work immediately in any Go project._
  <br/><br/>
- **Drop-in Mage Replacement**<br/>
  _`magex` is a superset of `mage` with all MAGE-X commands built-in. Your existing magefiles still work, now enhanced with 190+ professional commands._
  <br/><br/>
- **Cross-Platform Excellence**<br/>
  _Full support for Linux, macOS, and Windows with multi-architecture builds, parallel execution, and CPU-aware optimization._
  <br/><br/>
- **Security-First Architecture**<br/>
  _Input validation, secure command execution, and minimal dependencies. Built for environments where security matters._
  <br/><br/>
- **Professional Release Management**<br/>
  _Automated versioning, GitHub integration, GoDocs proxy sync, and release automation for production deployments._
  <br/><br/>
- **Smart Documentation System**<br/>
  _Hybrid pkgsite/godoc support with auto-detection, port management, and cross-platform browser integration._
  <br/><br/>

<br/>

## ⚡ Quick Start

### Which Path Should I Take?

MAGE-X automatically detects your project structure and **just works**:

| Your Project Structure              | What MAGE-X Does                  |
|-------------------------------------|-----------------------------------|
| **Single Binary** `main.go` in root | Builds to `bin/projectname`       |
| **Multi-Binary** `cmd/*/main.go`    | Auto-detects & builds first found |
| **Library** No main package         | Verifies compilation with `./...` |

> 📖 See the [Complete Quick Start Guide](docs/QUICK_START.md) for project-specific examples and troubleshooting.

<br>

### Zero Boilerplate Installation

```bash
# Install magex (production branch)
go install github.com/mrz1836/mage-x/cmd/magex@latest

# Auto-update to latest stable release (with proper version info)
magex update:install

# Now use it in ANY Go project (no setup!)
# Use either 'magex' or its shorter alias 'mgx'
magex build         # Automatically detects & builds your project
magex test          # Run your tests
magex bench         # Run your benchmarks
magex lint:fix      # Fix any linting issues
magex format:fix    # Format your code

# That's it! No magefile.go needed! 🚀
```

<br>

### Quick Project Check
```bash
# See what MAGE-X detects in your project:
magex build --dry-run  # Shows what will be built without building

# or use the shorter alias:
mgx build --dry-run    # Same command, shorter to type
```

<br>

### Full Command List

```bash
magex help          # List all commands & get help
magex -n            # Commands by namespace
magex -search test  # Find specific commands
# Remember: you can use 'magex' or 'mgx' interchangeably
```

<br>

### _Optional:_ Add Custom Commands

Only create a `magefile.go` if you need project-specific commands:

```go
//go:build mage
package main

// Your custom command (works alongside all MAGE-X commands!)
func Deploy() error {
    // Custom deployment logic
    return nil
}
```

Now you have both:
```bash
magex build    # Built-in MAGE-X command
magex deploy   # Your custom command
```

### Hybrid Execution Model

MAGE-X uses a smart hybrid approach that provides the best of both worlds:

- **Built-in commands** execute directly for speed and consistency (`magex build`, `magex test`, etc.)
- **Custom commands** in your magefile.go are automatically discovered and delegated to `mage`
- **Unified experience** - one command line interface for everything

This means:
- `magex build` (or `mgx build`) always behaves consistently across all projects
- Your custom commands work seamlessly without any setup
- No plugin compilation or platform-specific issues
- Zero configuration required

<br>

## 🚀 Features

### Core Excellence
- **Command Execution**: Secure, interface-based command execution with validation
- **Native Logging**: Colored output, progress indicators, and structured logging
- **Complete Build System**: All essential build, test, lint, and release tasks
- **Version Management**: Automatic version detection and update infrastructure

### Developer Experience
- **Release Automation**: Multi-platform asset building with GitHub integration
- **Configuration Management**: Flexible mage.yaml with smart defaults
- **Command Discovery**: Comprehensive CLI with intuitive command structure
- **Help System**: Built-in documentation and usage examples

### Production Features
- **Security Scanning**: Vulnerability detection with govulncheck integration
- **Configuration Management**: Centralized project configuration and validation
- **Extensible Architecture**: Plugin-ready foundation for custom production needs

<br/>

## ⚙️ Configuration

MAGE-X works without any configuration, but you can customize behavior with `.mage.yaml` or environment variables.

<details>
<summary>📋 <strong>Configuration File (.mage.yaml)</strong></summary>

Create `.mage.yaml` in your project root for custom settings:

```yaml
project:
  name: my-project
  binary: myapp
  version: v1.0.0
  module: github.com/user/my-project

build:
  output: bin
  trimpath: true
  platforms:
    - linux/amd64
    - darwin/amd64
    - darwin/arm64
    - windows/amd64
  tags:
    - prod
  ldflags:
    - -s -w
    - -X main.version={{.Version}}
    - -X main.commit={{.Commit}}

test:
  parallel: true
  timeout: 10m
  race: false
  cover: true
  covermode: atomic

lint:
  golangci_version: v2.3.1
  timeout: 5m

tools:
  golangci_lint: v2.3.1
  fumpt: latest
  govulncheck: latest

release:
  github:
    owner: mrz
    repo: my-project
  platforms:
    - linux/amd64
    - darwin/amd64
    - darwin/arm64
    - windows/amd64
```

</details>

<details>
<summary>🌍 <strong>Environment Variable Overrides</strong></summary>

Override any configuration with environment variables:

```bash
# Build configuration
export MAGE_X_BINARY_NAME=myapp
export MAGE_X_BUILD_TAGS=prod,feature1
export GOOS=linux
export GOARCH=amd64

# Test configuration
export MAGE_X_VERBOSE=true
export MAGE_X_TEST_RACE=true
export MAGE_X_TEST_TIMEOUT=15m

# Linting configuration
export MAGE_X_LINT_VERBOSE=true   # Enable verbose linting output
export MAGE_X_LINT_TIMEOUT=10m

# Tool configuration
export MAGE_X_PARALLEL=8

# Benchmark timing with parameters
magex bench time=50ms         # Quick benchmarks
magex bench time=10s          # Full benchmarks
magex bench:cpu time=30s      # CPU profiling
magex bench:save time=5s output=results.txt
magex bench time=2s count=5   # With custom count parameter

# Version management examples with new parameter format
magex version:bump bump=patch              # Bump patch version
magex version:bump bump=minor push         # Bump minor and push
magex version:bump bump=patch branch=master push  # Switch to master, bump, and push
magex git:tag version=1.2.3                # Create git tag
magex git:commit message="fix: bug fix"    # Commit with message
```

</details>

<br/>

## 📚 Documentation

For comprehensive documentation, visit the [docs](docs) directory:

- **[Getting Started](docs/README.md)** - Complete documentation index
- **[Claude Code Commands](docs/CLAUDE_COMMANDS.md)** - 13 optimized slash commands for agent orchestration
- **[Namespace Interface Architecture](docs/NAMESPACE_INTERFACES.md)** - Modern interface-based namespace system
- **[API Reference](docs/API_REFERENCE.md)** - Complete interface and API documentation
- **[Quick Start](docs/QUICK_START.md)** - Get up and running in minutes
- **[Configuration Reference](docs/CONFIGURATION.md)** - Complete configuration guide

<br>

### Available Commands

MAGE-X provides 190+ commands organized by functionality. All commands work instantly through the `magex` CLI.

<details>
<summary>🎯 <strong>Essential Commands</strong></summary>

**Quick Discovery:**
```bash
magex help:default       # Beautiful command listing with categories and emojis
magex -l                 # Plain list of all available targets
magex -search test       # Find specific commands
```

**Most Used Commands:**
```bash
magex                    # Run default build
magex build              # Build for current platform
magex test               # Run complete test suite
magex bench              # Run benchmarks
magex lint:fix           # Auto-fix linting issues
magex format:fix		 # Format code automatically
magex version:bump 	     # Bump version (patch, minor, major)
```

</details>

<details>
<summary>🏗️ <strong>Code Generation</strong></summary>

```bash
magex generate:default    # Run go generate
magex generate:clean      # Remove generated files
```

</details>

<details>
<summary>📦 <strong>Dependencies & Modules</strong></summary>

```bash
# Dependency Management
magex deps:update         # Update dependencies (safe - no major version bumps)
magex deps:update allow-major  # Update including major versions (v1→v2, etc)
magex deps:tidy           # Clean up go.mod and go.sum
magex deps:download       # Download all dependencies
magex deps:outdated       # Show outdated dependencies
magex deps:audit          # Audit dependencies for vulnerabilities

# Multi-Module Update (for monorepos with multiple go.mod files)
magex deps:update all-modules                    # Update all modules in workspace
magex deps:update all-modules dry-run            # Preview modules without updating
magex deps:update all-modules fail-fast          # Stop on first module error
magex deps:update all-modules verbose            # Show detailed update info
magex deps:update all-modules allow-major stable-only verbose fail-fast  # All options

# Module Management
magex mod:update          # Update go.mod file
magex mod:tidy            # Tidy the go.mod file
magex mod:verify          # Verify module checksums
magex mod:download        # Download modules
magex mod:graph           # Visualize dependency graph as tree with relationships
magex mod:why             # Show why specific modules are needed

# Dependency Graph Examples
magex mod:graph                                  # Default tree view with versions
magex mod:graph depth=3                          # Limit depth to 3 levels
magex mod:graph show_versions=false              # Hide version numbers
magex mod:graph format=json                      # JSON output format
magex mod:graph format=dot                       # DOT format for graphviz
magex mod:graph format=mermaid                   # Mermaid diagram format
magex mod:graph filter=github.com depth=2        # Filter + depth combined
magex mod:graph show_versions=false format=tree  # Clean tree view

# Module Dependency Analysis Examples
magex mod:why github.com/stretchr/testify        # Show why testify is needed
magex mod:why github.com/pkg/errors golang.org/x/sync  # Analyze multiple modules
```

</details>

<details>
<summary>📦 <strong>Build & Compilation</strong></summary>

```bash
magex build:default       # Build for current platform
magex build:all           # Build for all configured platforms
magex build:clean         # Clean build artifacts
magex build:generate      # Generate code before building
magex build:linux         # Build for Linux amd64
magex build:darwin        # Build for macOS (amd64 and arm64)
magex build:windows       # Build for Windows amd64
magex build:install       # Install binary to $GOPATH/bin
magex build:dev           # Build and install development version (forced 'dev' version)
magex build:prebuild      # Pre-build all packages to warm cache
magex build:prebuild parallel=2  # Pre-build with 2 parallel processes
magex build:prebuild p=4         # Pre-build with 4 parallel processes (short form)

# Memory-efficient prebuild strategies (great for CI/CD environments)
magex build:prebuild strategy=incremental batch=10  # Build in batches of 10 packages
magex build:prebuild strategy=mains-first           # Build main packages first, then dependencies
magex build:prebuild strategy=smart                 # Auto-select best strategy based on available memory
magex build:prebuild strategy=full p=8              # Traditional full build with 8 parallel jobs

# Advanced prebuild options
magex build:prebuild strategy=incremental batch=5 delay=100   # 100ms delay between batches
magex build:prebuild strategy=mains-first mains-only=true     # Only build main packages
magex build:prebuild exclude=test verbose=true                # Exclude test packages, verbose output
```

</details>

<details>
<summary>🧪 <strong>Testing & Quality</strong></summary>

```bash
magex test                     # Run complete test suite with linting
magex test:unit                # Run unit tests only
magex test:short               # Run short tests (excludes integration tests)
magex test:race                # Run tests with race detector
magex test:cover               # Run tests with coverage analysis
magex test:coverrace           # Run tests with both coverage and race detector
magex test:bench               # Run benchmark tests
magex bench                    # Run benchmarks with default timing
magex bench time=50ms          # Run quick benchmarks (50ms duration)
magex bench time=10s count=3   # Run benchmarks with custom time and count
magex test:fuzz                # Run fuzz tests (default: 10s)
magex test:fuzz time=30s       # Run fuzz tests with custom duration
magex test:fuzzShort time=1s   # Run short fuzz tests with custom duration
magex test:integration         # Run integration tests

# JSON Output Support for Test Commands
magex test -json               # Run tests with JSON output for tooling
magex test:unit -json          # Unit tests with JSON output
magex test:race -json          # Race detection with JSON output
magex test:cover -json         # Coverage tests with JSON output
magex test:coverrace -json     # Coverage + race with JSON output
magex test:short -json         # Short tests with JSON output

# Advanced Test Options (all commands support these flags)
magex test -v                  # Verbose output
magex test:cover -count=3      # Run tests 3 times
magex test -failfast           # Stop at first failure
magex test -shuffle=on         # Randomize test execution
magex test -parallel=4         # Set parallel test execution

# Code Quality & Linting
magex format:fix               # Run code formatting
magex lint                     # Run essential linters
magex lint verbose=true        # Run linters with verbose output
magex lint:fix                 # Auto-fix linting issues + apply formatting
magex lint:issues              # Scan for TODOs, FIXMEs, nolint directives, and test skips
magex lint:verbose             # Alternative: dedicated verbose linting command
magex lint:version 		       # Show golangci-lint version
magex test:vet                 # Run go vet static analysis
magex tools:verify             # Show tool version information
```

</details>

<details>
<summary>🤖 <strong>CI Test Output Mode</strong></summary>

MAGE-X automatically detects CI environments and produces structured test output with precise file:line locations for failures. This eliminates complex bash/jq parsing in CI workflows.

**Automatic Detection** (Zero Configuration):
```bash
# In GitHub Actions (or any CI with CI=true) - works automatically!
magex test:unit    # Produces GitHub annotations + JSONL output

# Locally - unchanged behavior
magex test:unit    # Standard terminal output
```

**Explicit CI Mode**:
```bash
magex test:unit ci          # Force CI mode locally (preview CI output)
magex test:unit ci=false    # Disable CI mode in CI environment
```

**All test commands support CI mode**:
- `magex test:unit ci`
- `magex test:race ci`
- `magex test:cover ci`
- `magex test:fuzz ci`

**Output in CI**:
- **GitHub Annotations**: Clickable file:line links in PR sidebar
- **Step Summary**: Markdown table in `$GITHUB_STEP_SUMMARY`
- **Structured Output**: `.mage-x/ci-results.jsonl` for automation

**Configuration** (optional in `.mage.yaml`):
```yaml
test:
  ci_mode:
    enabled: auto          # auto (default), on, or off
    format: github         # auto, github, or json
    context_lines: 20      # Lines of code context around failures
    output_path: ".mage-x/ci-results.jsonl"
```

**Environment Variables**:
```bash
export MAGE_X_CI_MODE=auto      # auto/on/off
export MAGE_X_CI_FORMAT=github  # github/json/auto
export MAGE_X_CI_CONTEXT=20     # Context lines (0-100)
```

</details>

<details>
<summary>✅ <strong>Code Validation & Formatting</strong></summary>

```bash
# Code Validation
magex vet:default         # Run go vet
magex vet:all             # Run go vet with all checks
magex vet:parallel        # Run go vet in parallel
magex vet:shadow          # Check for variable shadowing
magex vet:strict          # Run strict vet checks

# Code Formatting
magex format:default      # Default formatting (gofmt)
magex format:all          # Format all supported file types
magex format:gofmt        # Run gofmt
magex format:fumpt        # Run gofumpt (stricter formatting)
magex format:imports      # Format imports
magex format:go           # Format Go files
magex format:yaml         # Format YAML files
magex format:json         # Format JSON files
magex format:fix          # Fix formatting issues automatically
magex format:check        # Check if files are properly formatted
```

</details>

<details>
<summary>📊 <strong>Analysis & Metrics</strong></summary>

```bash
magex metrics:loc         # Analyze lines of code
magex metrics:loc json    # Output LOC metrics as JSON
magex metrics:mage        # Analyze magefiles and targets
magex metrics:coverage    # Generate coverage reports
magex metrics:complexity  # Analyze code complexity

# Performance & Benchmarking
magex bench               # Default benchmark operations
magex bench time=50ms     # Quick benchmarks (50ms runs)
magex bench time=10s      # Comprehensive benchmarks (10s runs)
magex bench count=3       # Run benchmarks 3 times
magex bench:profile       # Profile application performance
magex bench:compare       # Compare benchmark results
magex bench:regression    # Check for performance regressions
magex bench:cpu time=30s  # CPU usage benchmarks with custom duration
magex bench:mem time=2s   # Memory usage benchmarks
```

</details>

<details>
<summary>📚 <strong>Documentation</strong></summary>

```bash
magex docs               # Generate and serve documentation (generate + serve in one command)
magex docs:generate      # Generate Go package documentation from source code
magex docs:serve         # Serve documentation locally with hybrid pkgsite/godoc support
magex docs:build         # Build enhanced static documentation files with metadata
magex docs:check         # Validate documentation completeness and quality
magex docs:update        # Update GoDocs proxy (trigger pkg.go.dev sync)
magex docs:godocs        # Update GoDocs proxy (alias for docs:update)
```

</details>

<details>
<summary>🔀 <strong>Git & Version Control</strong></summary>

```bash
# Git Operations
magex git:status                                # Show git repository status
magex git:commit message="fix: commit message"  # Commit changes with message parameter
magex git:tag version=1.2.3                     # Create and push a new tag with version parameter
magex git:tagremove version=1.2.3               # Remove a tag
magex git:tagupdate version=1.2.3               # Force update a tag

# Version Management
magex version:show         # Display current version information
magex version:check        # Check version information and compare with latest
magex version:update       # Update to latest version
magex version:bump         # Bump version (patch, minor, major)
magex version:changelog    # Generate changelog from git history
magex version:tag          # Create version tag
magex version:compare      # Compare two versions
magex version:validate     # Validate version format

# Version Bump Examples (now using parameters)
magex version:bump                          # Bump patch version (default)
magex version:bump bump=minor               # Bump minor version
magex version:bump bump=major major-confirm # Bump major version with confirmation
magex version:bump bump=minor push          # Bump minor and push to remote

# Branch Parameter Examples (recommended for GitButler users)
magex version:bump bump=patch branch=master push    # Switch to master, bump patch, and push
magex version:bump bump=minor branch=main           # Switch to main branch before bumping
magex version:bump bump=major branch=master major-confirm push  # Major bump on master with push

# Dry-run mode (preview changes without making them)
magex version:bump dry-run                 # Preview patch bump
magex version:bump bump=minor dry-run      # Preview minor bump
magex version:bump bump=major major-confirm push dry-run  # Preview major bump with push
magex version:bump bump=patch branch=master dry-run      # Preview branch switch and bump

# Important Notes:
# - Uncommitted changes will block version bump operation (safety check)
# - When using branch parameter, you will ALWAYS return to your original branch after completion
# - Branch switch-back happens even if tag creation or push fails (guaranteed cleanup)
# - If network issues occur during pull, operation stops but still switches back to original branch
```

</details>

<details>
<summary>🚀 <strong>Release Management</strong></summary>

```bash
# Core Release Operations
magex release              # Create a new release from the latest tag
magex release godocs       # Create a release and update GoDocs proxy
magex release:default      # Create a new release from the latest tag
magex release:test         # Dry-run release without publishing
magex release:snapshot     # Build release artifacts without git tag
magex release:localinstall # Build from latest tag and install locally

# Release Setup & Validation
magex release:init         # Initialize .goreleaser.yml configuration
magex release:check        # Validate .goreleaser.yml configuration
magex release:validate     # Comprehensive release readiness validation
magex release:changelog    # Generate changelog from git history
magex release:clean        # Clean release artifacts and build cache
```

</details>

<details>
<summary>🔧 <strong>Development Tools</strong></summary>

```bash
magex tools:update        # Update all development tools
magex tools:install       # Install all required development tools
magex tools:verify        # Check if all required tools are available
magex deps:audit          # Run vulnerability check using govulncheck
magex build:install       # Install the project binary
magex build:dev           # Build and install development version (forced 'dev' version)
magex install:stdlib      # Install Go standard library for cross-compilation
magex uninstall           # Remove installed binary

# Spec-Kit CLI Management (requires speckit.enabled: true in .mage.yaml)
magex speckit:install     # Install spec-kit prerequisites (uv, uvx, specify-cli)
magex speckit:check       # Verify spec-kit installation and report version info
magex speckit:upgrade     # Upgrade spec-kit with automatic constitution backup/restore
```

</details>

<details>
<summary>⚙️ <strong>Configuration Management</strong></summary>

```bash
magex configure:init      # Initialize a new mage configuration
magex configure:show      # Display the current configuration
magex configure:update    # Update configuration values interactively
magex configure:export    # Export configuration to file
magex configure:import    # Import configuration from file
magex configure:validate  # Validate configuration integrity
magex configure:schema    # Show configuration schema

# YAML Configuration
magex yaml:init           # Create mage.yaml configuration
magex yaml:validate       # Validate YAML configuration
magex yaml:show           # Show current YAML configuration
magex yaml:template       # Generate YAML templates
magex yaml:format         # Format YAML files
magex yaml:check          # Check YAML syntax
magex yaml:merge          # Merge YAML configurations
magex yaml:convert        # Convert between YAML formats
magex yaml:schema         # Show YAML schema
```

</details>


<details>
<summary>📖 <strong>Help & Updates</strong></summary>

```bash
# Help System
magex help:default        # Show general help
magex help:commands       # List all available commands
magex help:examples       # Show usage examples
magex help:gettingstarted # Getting started guide
magex help:completions    # Generate shell completions

# Update Management
magex update:check        # Check for updates
magex update:install      # Install the latest update

# Installation Management
magex install:default     # Default installation
magex install:local       # Install locally
magex install:binary      # Install project binary
magex install:tools       # Install development tools
magex install:go          # Install Go
magex install:stdlib      # Install Go standard library
magex install:systemwide  # Install system-wide
magex install:deps        # Install dependencies
magex install:mage        # Install mage
magex install:githooks    # Install git hooks
magex install:ci          # Install CI components
magex install:certs       # Install certificates
magex install:package     # Install package
magex install:all         # Install everything
magex uninstall           # Remove installation
```

</details>

<br/>

### 📋 Complete Command List

Run `magex -l` to see a plain list of all available commands (190+ commands), or use `magex help` for a beautiful categorized view with descriptions and usage tips.

<br/>

### 📚 Documentation System

MAGE-X includes a hybrid documentation system with auto-detection and cross-platform browser integration.

<details>
<summary>📖 <strong>Documentation Features & Capabilities</strong></summary>

#### Smart Tool Detection
- **Auto-detection**: Automatically detects and uses the best available documentation tool
- **Hybrid Support**: Supports both `pkgsite` (modern) and `godoc` (classic) with smart fallback
- **Auto-installation**: Automatically installs missing tools when needed
- **Configuration Control**: Override tool selection with `docs.tool` in .mage.yaml

#### Multiple Serving Modes
```bash
# Documentation serving
magex docs:serve         # Serve documentation locally
magex docs:godocs        # Serve with godoc format
magex docs:api           # API documentation
magex docs:examples      # Example documentation
magex docs:readme        # README documentation
```

#### Advanced Features
- **Port Management**: Automatic port detection and conflict resolution
- **Cross-Platform**: Browser auto-opening on macOS, Linux, and Windows
- **CI/CD Ready**: Detects CI environments and disables browser opening
- **Comprehensive Generation**: Documents all packages with categorization
- **Static Building**: Enhanced Markdown with metadata and navigation
- **Build Artifacts**: JSON metadata and organized output structure

</details>

<br/>

### 🔧 Advanced Namespaces

MAGE-X includes specialized namespaces for power users. These are opt-in to keep the core installation lightweight.

<details>
<summary>🔧 <strong>Advanced Namespace Configuration</strong></summary>

**Available Specialized Namespaces:**
- **Bench** - Performance benchmarking and profiling
- **Releases** - Release creation and asset distribution
- **Yaml** - YAML configuration management and validation

**To enable advanced features in your magefile:**
```go
//go:build mage

package main

import "github.com/mrz1836/mage-x/pkg/mage"

// Export core namespaces (already included by default)
type (
    Build = mage.Build
    Test  = mage.Test
    // ... other default namespaces
)

// Add advanced namespaces as needed
type (
    Integrations = mage.Integrations
    Bench        = mage.Bench
    Releases     = mage.Releases
)
```

This approach keeps the default installation lightweight while allowing power users to access advanced features when needed.

**Example: Production Operations**
```go
// Run benchmarking and performance analysis
func PerformanceTest() error {
    var bench Bench
    return bench.Default()
}

// Run quick benchmarks with custom timing
func QuickBench() error {
    var bench Bench
    return bench.DefaultWithArgs("time=50ms")
}
```

</details>

See the [examples directory](examples) for more custom magefile implementations.

<br/>

## 📐 Spec-Driven Development

This project uses [Spec-Kit](https://github.com/github/spec-kit) for spec-driven development, enabling executable specifications that directly generate working implementations. All feature specifications live in the [`specs/`](specs) directory.

<details>
<summary><strong><code>Creating a New Feature Spec</code></strong></summary>
<br/>

Follow the spec-kit workflow to create and implement new features:

### 1. **Constitution** (One-time setup)
```bash
/speckit.constitution
```
Establish project principles and development guidelines.

### 2. **Specify** (Define requirements)
```bash
/speckit.specify
```
Describe *what* and *why*, not the technology. Focus on requirements and user stories.

### 3. **Clarify** (Optional but recommended)
```bash
/speckit.clarify
```
Address underspecified areas with targeted questions.

### 4. **Plan** (Design implementation)
```bash
/speckit.plan
```
Create technical implementation strategy with your chosen tech stack.

### 5. **Tasks** (Generate action items)
```bash
/speckit.tasks
```
Generate actionable, dependency-ordered task lists.

### 6. **Analyze** (Validate consistency)
```bash
/speckit.analyze
```
Validate cross-artifact consistency before implementation.

### 7. **Implement** (Build the feature)
```bash
/speckit.implement
```
Execute all tasks to build the feature according to the plan.

**Pro tip:** Run `/speckit.checklist` anytime to generate custom validation criteria.

</details>

<details>
<summary><strong><code>Upgrading Spec-Kit</code></strong></summary>
<br/>

### Automated Commands (Recommended)

MAGE-X provides a `speckit` namespace for managing the spec-kit CLI:

```bash
# Install spec-kit prerequisites (uv, uvx, specify-cli)
magex speckit:install

# Check spec-kit installation and version
magex speckit:check

# Upgrade spec-kit with automatic backup/restore
magex speckit:upgrade
```

The `speckit:upgrade` command automatically:
- ✅ Backs up your constitution to `.specify/backups/` with timestamp
- ✅ Upgrades the spec-kit CLI using `uv tool upgrade`
- ✅ Updates project configuration with `--force` flag
- ✅ Restores your constitution from backup
- ✅ Tracks version history in `.specify/version.txt`
- ✅ Cleans old backups (keeps last 5)
- ✅ Verifies the upgrade with `specify check`

**Configuration** (in `.mage.yaml`):
```yaml
speckit:
  enabled: true                                    # Enable speckit commands (opt-in)
  constitution_path: ".specify/memory/constitution.md"
  backup_dir: ".specify/backups"
  backups_to_keep: 5
  ai_provider: "claude"
```

### Manual Upgrade (Alternative)

If you prefer manual control, follow these steps:

#### Step 1: Backup Your Constitution
```bash
cp .specify/memory/constitution.md ~/constitution.backup.md
```

#### Step 2: Upgrade the CLI

**If using persistent installation:**
```bash
uv tool upgrade specify-cli
```

**Verify the upgrade:**
```bash
specify check
```

#### Step 3: Upgrade the Project Configuration
```bash
uvx --from git+https://github.com/github/spec-kit.git specify init --here --ai claude --force
```

The `--force` flag safely merges updated configuration into your existing project.

#### Step 4: Restore Custom Constitution (if needed)
If you have custom constitution changes, carefully merge them back from your backup.

</details>

<br/>

## 🧪 Examples & Tests

All examples and tests run via GitHub Actions using Go 1.24+. View the [examples directory](examples) for complete project demonstrations.

### Run Tests

```bash
# Quick test suite
magex test

# Comprehensive testing
magex test:race test:cover test:fuzz

# JSON output for tooling integration
magex test -json               # JSON output for CI/CD tools
magex test:coverrace -json     # Coverage + race with JSON output

# Performance benchmarks
magex bench                    # Default benchmarks (10s duration)
magex bench time=50ms          # Quick benchmarks for CI
magex bench time=10s count=3   # Comprehensive benchmarks with count parameter
magex test:bench               # Via test namespace
magex test:bench time=30s      # Test namespace with custom timing
```

### Example Projects

- **[Basic Project](examples/basic)** - Zero-configuration MAGE-X usage
- **[With Configuration](examples/with-config)** - Using .mage.yaml for project customization
- **[With Custom Commands](examples/with-custom)** - Adding project-specific commands
- **[Zero Config](examples/zero-config)** - Instant productivity with magex

<br/>

## ⚡ Benchmarks

Performance benchmarks for core MAGE-X operations:

| Operation             | Time  | Memory | Notes                             |
|-----------------------|-------|--------|-----------------------------------|
| Build Detection       | 1.2ms | 256KB  | Project type and configuration    |
| Command Execution     | 0.8ms | 128KB  | Secure command validation         |
| Configuration Loading | 2.1ms | 512KB  | YAML parsing and validation       |

> Benchmarks run on Apple M1 Pro (ARM64) with Go 1.24+
> All operations show consistent sub-5ms performance with minimal memory allocation

<br/>

## 🛠️ Code Standards

MAGE-X follows strict coding standards and best practices:

- **Code Quality**: >80% test coverage, comprehensive linting, and security scanning
- **Go Best Practices**: Idiomatic Go code following community standards
- **Security First**: Input validation, secure command execution, minimal dependencies
- **Documentation**: Comprehensive godoc coverage and usage examples
- **AI Compliance**: Machine-readable guidelines for AI assistants

Read more about our [code standards](.github/CODE_STANDARDS.md) and [contribution guidelines](.github/CONTRIBUTING.md).

<br/>

## 🤖 AI Usage & Assistant Guidelines
Read the [AI Usage & Assistant Guidelines](.github/tech-conventions/ai-compliance.md) for details on how AI is used in this project and how to interact with the AI assistants.

<br/>

## 👥 Maintainers

| [<img src="https://github.com/mrz1836.png" height="50" width="50" alt="Maintainer" />](https://github.com/mrz1836) |
|:------------------------------------------------------------------------------------------------------------------:|
|                                         [mrz](https://github.com/mrz1836)                                          |

<br/>

## 🤝 Contributing

We welcome contributions from the community! Please read our [contributing guidelines](.github/CONTRIBUTING.md) and [code of conduct](.github/CODE_OF_CONDUCT.md).

### How Can I Help?

All kinds of contributions are welcome! :raised_hands:

- **⭐ Star the project** to show your support
- **🐛 Report bugs** through GitHub issues
- **💡 Suggest features** with detailed use cases
- **📝 Improve documentation** with examples and clarity
- **🔧 Submit pull requests** with bug fixes or new features


[![Stars](https://img.shields.io/github/stars/mrz1836/mage-x?label=Please%20like%20us&style=social)](https://github.com/mrz1836/mage-x/stargazers)

<br/>

## 📝 License

[![License](https://img.shields.io/github/license/mrz1836/mage-x.svg?style=flat&v=1)](LICENSE)

<br/>

---

<div align="center">
  <p>
    <strong>Built with ❤️ by the Go community</strong>
  </p>
  <p>
    <em>MAGE-X: Write Once, Mage Everywhere</em>
  </p>
</div>
