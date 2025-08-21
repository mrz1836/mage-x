 # 🪄 MAGE-X
> Write Once, Mage Everywhere: Zero-Boilerplate Build Automation for Go

<table>
  <thead>
    <tr>
      <th>CI&nbsp;/&nbsp;CD</th>
      <th>Quality&nbsp;&amp;&nbsp;Security</th>
      <th>Docs&nbsp;&amp;&nbsp;Meta</th>
      <th>Community</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td valign="top" align="left">
        <a href="https://github.com/mrz1836/mage-x/releases">
          <img src="https://img.shields.io/github/release-pre/mrz1836/mage-x?logo=github&style=flat&v=1" alt="Latest Release">
        </a><br/>
        <a href="https://github.com/mrz1836/mage-x/actions">
          <img src="https://img.shields.io/github/actions/workflow/status/mrz1836/mage-x/fortress.yml?branch=master&logo=github&style=flat" alt="Build Status">
        </a><br/>
        <a href="https://github.com/mrz1836/mage-x/commits/master">
          <img src="https://img.shields.io/github/last-commit/mrz1836/mage-x?style=flat&logo=clockify&logoColor=white" alt="Last commit">
        </a>
      </td>
      <td valign="top" align="left">
        <a href="https://goreportcard.com/report/github.com/mrz1836/mage-x">
          <img src="https://goreportcard.com/badge/github.com/mrz1836/mage-x?style=flat&v=1" alt="Go Report Card">
        </a><br/>
        <a href="https://app.codecov.io/gh/mrz1836/mage-x/tree/master">
          <img src="https://codecov.io/gh/mrz1836/mage-x/branch/master/graph/badge.svg?style=flat" alt="Code Coverage">
        </a><br/>
        <a href="https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck">
          <img src="https://img.shields.io/badge/security-govulncheck-blue?style=flat&logo=springsecurity&logoColor=white" alt="Security Scanning">
        </a><br/>
        <a href=".github/SECURITY.md">
          <img src="https://img.shields.io/badge/security-policy-blue?style=flat&logo=springsecurity&logoColor=white" alt="Security Policy">
        </a>
      </td>
      <td valign="top" align="left">
        <a href="https://golang.org/">
          <img src="https://img.shields.io/github/go-mod/go-version/mrz1836/mage-x?style=flat" alt="Go version">
        </a><br/>
        <a href="https://pkg.go.dev/github.com/mrz1836/mage-x">
          <img src="https://pkg.go.dev/badge/github.com/mrz1836/mage-x.svg?style=flat" alt="Go docs">
        </a><br/>
        <a href=".github/AGENTS.md">
          <img src="https://img.shields.io/badge/AGENTS.md-found-40b814?style=flat&logo=openai" alt="AI Agent Rules">
        </a><br/>
        <a href="https://magefile.org/">
          <img src="https://img.shields.io/badge/mage-powered-brightgreen?style=flat&logo=probot&logoColor=white" alt="Mage Powered">
        </a>
      </td>
      <td valign="top" align="left">
        <a href="https://github.com/mrz1836/mage-x/graphs/contributors">
          <img src="https://img.shields.io/github/contributors/mrz1836/mage-x?style=flat&logo=contentful&logoColor=white" alt="Contributors">
        </a><br/>
        <a href="https://github.com/sponsors/mrz">
          <img src="https://img.shields.io/badge/sponsor-mrz-181717.svg?logo=github&style=flat" alt="Sponsor">
        </a><br/>
        <a href="https://github.com/mrz1836/mage-x/stargazers">
          <img src="https://img.shields.io/github/stars/mrz1836/mage-x?style=social?v=1" alt="Stars">
        </a>
      </td>
    </tr>
  </tbody>
</table>

<br/>

## 🗂️ Table of Contents
* [What's Inside](#-whats-inside)
* [Quick Start](#-quick-start)
* [Features](#-features)
* [Advanced Features](#-advanced-features)
* [Configuration](#-configuration)
* [Documentation](#-documentation)
* [Examples & Tests](#-examples--tests)
* [Benchmarks](#-benchmarks)
* [Code Standards](#-code-standards)
* [AI Agent Ecosystem](#-ai-agent-ecosystem)
* [Maintainers](#-maintainers)
* [Contributing](#-contributing)
* [License](#-license)

<br/>

## 🧩 What's Inside

**MAGE-X** revolutionizes Go build automation with TRUE zero-boilerplate. Unlike traditional Mage which requires writing wrapper functions, MAGE-X provides all commands instantly through the `magex` binary.

<br/>

- **Truly Zero-Configuration**<br/>
  _No magefile.go needed. No imports. No wrappers. Just install `magex` and all built-in and custom commands work immediately in any Go project._
  <br/><br/>
- **Drop-in Mage Replacement**<br/>
  _`magex` is a superset of `mage` with all MAGE-X commands built-in. Your existing magefiles still work, now enhanced with 241+ professional commands._
  <br/><br/>
- **Cross-Platform Excellence**<br/>
  _Full support for Linux, macOS, and Windows with multi-architecture builds, parallel execution, and CPU-aware optimization._
  <br/><br/>
- **Security-First Architecture**<br/>
  _Input validation, secure command execution, and minimal dependencies. Built for environments where security matters._
  <br/><br/>
- **Advanced Recipe System**<br/>
  _Pre-built development patterns with templates and automation—from fresh project setup to complete CI/CD workflows._
  <br/><br/>
- **Professional Release Management**<br/>
  _Multi-platform asset building, automated versioning, GitHub integration, GoDocs proxy sync, and release automation for production deployments._
  <br/><br/>
- **Smart Documentation System**<br/>
  _Hybrid pkgsite/godoc support with auto-detection, port management, and cross-platform browser integration._
  <br/><br/>
- **AI Development Integration**<br/>
  _19 specialized AI agents and 13 Claude Code slash commands for intelligent workflows, code analysis, and automated development tasks._
  <br/><br/>
- **Enterprise Foundation**<br/>
  _Audit logging, configuration management, and extensible architecture ready for organizational governance and compliance needs._

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
```

<br>

### Full Command List

```bash
magex help          # List all commands & get help
magex -n            # Commands by namespace
magex -search test  # Find specific commands
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
- `magex build` always behaves consistently across all projects
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
- **Project Templates**: CLI, library, web API, and microservice templates
- **Release Automation**: Multi-platform asset building with GitHub integration
- **Configuration Management**: Flexible mage.yaml with smart defaults
- **Recipe System**: Pre-built patterns and templates for common scenarios

### User Experience Features
- **Command Discovery**: Comprehensive CLI with intuitive command structure
- **Help System**: Built-in documentation and usage examples
- **Recipe System**: Common patterns and best practices library
- **Project Templates**: Ready-made configurations for different project types

### Enterprise Features
- **Audit Logging**: Activity tracking and compliance reporting foundation
- **Security Scanning**: Vulnerability detection with govulncheck integration
- **Configuration Management**: Centralized project configuration and validation
- **Extensible Architecture**: Plugin-ready foundation for custom enterprise needs

<br/>

## 🏢 Advanced Features

MAGE-X includes enterprise-ready capabilities and extensible architecture for organizations.

### Enterprise Capabilities

Available enterprise functionality:

```bash
# Audit logging and compliance
magex audit:show          # View audit events
magex audit:stats         # Audit statistics
magex audit:export        # Export audit data

# Configuration management
magex configure:show      # Display configuration
magex configure:validate  # Validate configuration
```

**Production-Ready Features:**
- **Audit System**: Track build activities with export capabilities
- **Configuration Validation**: Schema-based configuration management
- **Security Integration**: govulncheck and policy enforcement
- **Extensible Architecture**: Plugin-ready foundation for custom enterprise needs

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

# Tool configuration
export MAGE_X_PARALLEL=8
export MAGE_X_LINT_TIMEOUT=10m

# Benchmark timing with parameters
magex bench time=50ms         # Quick benchmarks
magex bench time=10s          # Full benchmarks
magex bench:cpu time=30s      # CPU profiling
magex bench:save time=5s output=results.txt
magex bench time=2s count=5   # With custom count parameter

# Version management examples with new parameter format
magex version:bump bump=patch              # Bump patch version
magex version:bump bump=minor push         # Bump minor and push
magex git:tag version=1.2.3                # Create git tag
magex git:commit message="fix: bug fix"    # Commit with message
```

</details>

<br/>

## 📚 Documentation

For comprehensive documentation, visit the [docs](docs) directory:

- **[Getting Started](docs/README.md)** - Complete documentation index
- **[AI Agent Ecosystem](docs/SUB_AGENTS.md)** - 19 specialized agents for intelligent development workflows
- **[Claude Code Commands](docs/CLAUDE_COMMANDS.md)** - 13 optimized slash commands for agent orchestration
- **[Namespace Interface Architecture](docs/NAMESPACE_INTERFACES.md)** - Modern interface-based namespace system
- **[API Reference](docs/API_REFERENCE.md)** - Complete interface and API documentation
- **[Enterprise Features](docs/ENTERPRISE.md)** - Enterprise capabilities guide
- **[Quick Start](docs/QUICK_START.md)** - Get up and running in minutes
- **[Configuration Reference](docs/CONFIGURATION.md)** - Complete configuration guide

<br>

### Available Commands

MAGE-X provides 241+ commands organized by functionality. All commands work instantly through the `magex` CLI.

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
magex release:default    # Create a new release
```

</details>

<details>
<summary>🏗️ <strong>Project Setup & Initialization</strong></summary>

```bash
# Code Generation
magex generate:default    # Run go generate
magex generate:clean      # Remove generated files

# Project Initialization
magex init:default        # Default initialization
magex init:project        # Initialize a new project
magex init:library        # Initialize a library project
magex init:cli            # Initialize a CLI project
magex init:webapi         # Initialize a web API project
magex init:microservice   # Initialize a microservice project
magex init:tool           # Initialize a tool project
magex init:upgrade        # Upgrade existing project
magex init:templates      # Show available templates
magex init:config         # Initialize configuration
magex init:git            # Initialize git repository
magex init:mage           # Initialize mage files
magex init:ci             # Initialize CI/CD files
magex init:docker         # Initialize Docker files
magex init:docs           # Initialize documentation
magex init:license        # Add license file
magex init:makefile       # Create Makefile
magex init:editorconfig   # Create .editorconfig

# Recipe Management
magex recipes:list                    # List available recipes
magex recipes:show recipe=fresh-start # Show recipe details
magex recipes:run recipe=fresh-start  # Run the fresh-start recipe
magex recipes:search term=docker      # Search for recipes
magex recipes:create recipe=my-recipe # Create a custom recipe
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

# Module Management
magex mod:update          # Update go.mod file
magex mod:tidy            # Tidy the go.mod file
magex mod:verify          # Verify module checksums
magex mod:download        # Download modules
```

</details>

<details>
<summary>📦 <strong>Build & Compilation</strong></summary>

```bash
magex build:default       # Build for current platform
magex build:all           # Build for all configured platforms
magex build:docker        # Build Docker containers
magex build:clean         # Clean build artifacts
magex build:generate      # Generate code before building
magex build:linux         # Build for Linux amd64
magex build:darwin        # Build for macOS (amd64 and arm64)
magex build:windows       # Build for Windows amd64
magex build:install       # Install binary to $GOPATH/bin
magex build:prebuild      # Pre-build all packages to warm cache
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

# Code Quality & Linting
magex lint                     # Run essential linters
magex lint:fix                 # Auto-fix linting issues + apply formatting
magex lint:issues              # Scan for TODOs, FIXMEs, nolint directives, and test skips
magex test:vet                 # Run go vet static analysis
magex format:fix               # Run code formatting
magex tools:verify             # Show tool version information
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
magex format:markdown     # Format Markdown files
magex format:sql          # Format SQL files
magex format:dockerfile   # Format Dockerfiles
magex format:shell        # Format shell scripts
magex format:fix          # Fix formatting issues automatically
magex format:check        # Check if files are properly formatted
```

</details>

<details>
<summary>📊 <strong>Analysis & Metrics</strong></summary>

```bash
magex metrics:loc         # Analyze lines of code
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
magex git:status                           # Show git repository status
magex git:commit message="fix: commit message"  # Commit changes with message parameter
magex git:tag version=1.2.3                # Create and push a new tag with version parameter
magex git:tagremove version=1.2.3          # Remove a tag
magex git:tagupdate version=1.2.3          # Force update a tag

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

# Dry-run mode (preview changes without making them)
magex version:bump dry-run                 # Preview patch bump
magex version:bump bump=minor dry-run      # Preview minor bump
magex version:bump bump=major major-confirm push dry-run  # Preview major bump with push
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
magex install:stdlib      # Install Go standard library for cross-compilation
magex uninstall           # Remove installed binary
```

</details>

<details>
<summary>⚙️ <strong>Configuration Management</strong></summary>

```bash
magex configure:init      # Initialize a new mage configuration
magex configure:show      # Display the current configuration
magex configure:update    # Update configuration values interactively
magex configure:enterprise # Configure enterprise settings
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
<summary>🛡️ <strong>Security & Audit</strong></summary>

```bash
magex audit:show          # Display audit events with optional filtering
magex audit:stats         # Show audit statistics and summaries
magex audit:export        # Export audit data to various formats
magex audit:cleanup       # Clean up old audit entries
magex audit:enable        # Enable audit logging
magex audit:disable       # Disable audit logging
magex audit:report        # Generate comprehensive audit reports
```

</details>

<details>
<summary>🏢 <strong>Enterprise & Advanced Features</strong></summary>

```bash
# Enterprise Management
magex enterprise:init     # Initialize enterprise features
magex enterprise:config   # Configure enterprise settings
magex enterprise:deploy   # Deploy enterprise version
magex enterprise:rollback # Rollback deployment
magex enterprise:promote  # Promote deployment
magex enterprise:status   # Show deployment status
magex enterprise:backup   # Backup enterprise data
magex enterprise:restore  # Restore enterprise data

# Workflow Automation
magex workflow:execute    # Execute workflow
magex workflow:list       # List workflows
magex workflow:status     # Show workflow status
magex workflow:create     # Create workflow
magex workflow:validate   # Validate workflow
magex workflow:schedule   # Schedule workflow
magex workflow:template   # Workflow templates
magex workflow:history    # Show workflow history

# Integration Management
magex integrations:setup    # Setup integrations
magex integrations:test     # Test integrations
magex integrations:sync     # Sync integrations
magex integrations:notify   # Send notifications
magex integrations:status   # Show integration status
magex integrations:webhook  # Webhook operations
magex integrations:export   # Export integration data
magex integrations:import   # Import integration data

# CLI Operations
magex cli:default         # Default CLI operation
magex cli:help            # CLI help
magex cli:version         # Show CLI version
magex cli:completion      # Generate completions
magex cli:config          # CLI configuration
magex cli:update          # Update CLI
magex cli:bulk            # Bulk operations
magex cli:query           # Query operations
magex cli:dashboard       # CLI dashboard
magex cli:batch           # Batch operations
magex cli:monitor         # Monitoring
magex cli:workspace       # Workspace management
magex cli:pipeline        # Pipeline operations
magex cli:compliance      # Compliance checking

# Interactive Wizards
magex wizard:setup         # Interactive setup wizard
magex wizard:config        # Configuration wizard
magex wizard:project       # Project setup wizard
magex wizard:deploy        # Deployment wizard
magex wizard:troubleshoot  # Troubleshooting wizard
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
magex install:docker      # Install Docker components
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

Run `magex -l` to see a plain list of all available commands (241+ commands), or use `magex help` for a beautiful categorized view with descriptions and usage tips.

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
- **Static Building**: Enhanced markdown with metadata and navigation
- **Build Artifacts**: JSON metadata and organized output structure

</details>

<br/>

### 🏢 Enterprise Namespaces

MAGE-X includes specialized namespaces for power users. These are opt-in to keep the core installation lightweight.

<details>
<summary>🔧 <strong>Advanced Namespace Configuration</strong></summary>

**Available Specialized Namespaces:**
- **CLI** - Advanced command-line operations and bulk processing
- **Wizard** - Interactive setup and configuration wizards
- **Enterprise** - Audit logging and enterprise management
- **Workflow** - Build automation and pipeline orchestration
- **Bench** - Performance benchmarking and profiling
- **Releases** - Release creation and asset distribution
- **Yaml** - YAML configuration management and validation

**To enable enterprise features in your magefile:**
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

// Add enterprise namespaces as needed
type (
    CLI          = mage.CLI
    Wizard       = mage.Wizard
    Enterprise   = mage.Enterprise
    Workflow     = mage.Workflow
    Integrations = mage.Integrations
    Bench        = mage.Bench
    Releases     = mage.Releases
)
```

This approach keeps the default installation lightweight while allowing power users to access advanced features when needed.

**Example: Enterprise Operations**
```go
// Enable audit logging and compliance
func AuditReport() error {
    var enterprise Enterprise
    return enterprise.Audit()
}

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

## 🧪 Examples & Tests

All examples and tests run via GitHub Actions using Go 1.24+. View the [examples directory](examples) for complete project demonstrations.

### Run Tests

```bash
# Quick test suite
magex test

# Comprehensive testing
magex test:race test:cover test:fuzz

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
| Recipe Processing     | 3.5ms | 1MB    | Template expansion and validation |

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

## 🤖 AI Agent Ecosystem

MAGE-X features a comprehensive ecosystem of 19 specialized Claude Code AI agents designed for intelligent development workflows:

### Agent Categories
- **Core Development (5)**: Build, lint, deps, docs, security specialists
- **Testing & Quality (2)**: Coverage analysis and comprehensive test implementation
- **Release & CI/CD (3)**: Version management, git operations, GitHub automation
- **Architecture & Performance (4)**: Code architecture, refactoring, analysis, benchmarking
- **Enterprise & Workflow (3)**: Governance, automation pipelines, interactive guidance
- **Infrastructure (2)**: Configuration management, development environment setup

<br/>

### Key Features
- **Strategic Collaboration**: Agents coordinate intelligently for complex workflows
- **Parallel Execution**: Multiple agents work simultaneously for maximum efficiency
- **Security Integration**: All agents follow mage-x security-first patterns
- **Enterprise Scale**: Built for managing 30+ repositories with governance and compliance

<br/>

### Claude Code Integration
MAGE-X includes 13 optimized Claude Code slash commands that leverage the agent ecosystem:
- **Testing & Quality**: `/test`, `/fix`, `/quality`
- **Code Improvement**: `/dedupe`, `/explain`, `/optimize`
- **Documentation**: `/docs-update`, `/docs-review`
- **Development Workflow**: `/ci-diagnose`, `/release`
- **Architecture & Security**: `/architect`, `/security`
- **Planning**: `/prd`

See [Claude Code Commands](docs/CLAUDE_COMMANDS.md) for complete documentation.

<br/>

### Documentation & Guidelines
- **[AI Agent Ecosystem](docs/SUB_AGENTS.md)** — Complete agent directory and collaboration patterns
- **[Claude Code Commands](docs/CLAUDE_COMMANDS.md)** — Optimized slash commands for agent orchestration
- **[AGENTS.md](.github/AGENTS.md)** — Development rules and architectural guidelines
- **[CLAUDE.md](.github/CLAUDE.md)** — AI assistant integration guidelines
- **[.cursorrules](.cursorrules)** — Machine-readable policies for AI development tools

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
