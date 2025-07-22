 # 🪄 MAGE-X
> Write Once, Mage Everywhere: Enterprise-Grade Build Automation for Go

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
        <a href="https://github.com/mrz1836/go-mage/releases">
          <img src="https://img.shields.io/github/release-pre/mrz1836/go-mage?logo=github&style=flat&v=1" alt="Latest Release">
        </a><br/>
        <a href="https://github.com/mrz1836/go-mage/actions">
          <img src="https://img.shields.io/github/actions/workflow/status/mrz1836/go-mage/ci.yml?branch=main&logo=github&style=flat" alt="Build Status">
        </a><br/>
        <a href="https://github.com/mrz1836/go-mage/commits/main">
          <img src="https://img.shields.io/github/last-commit/mrz1836/go-mage?style=flat&logo=clockify&logoColor=white" alt="Last commit">
        </a>
      </td>
      <td valign="top" align="left">
        <a href="https://goreportcard.com/report/github.com/mrz1836/go-mage">
          <img src="https://goreportcard.com/badge/github.com/mrz1836/go-mage?style=flat" alt="Go Report Card">
        </a><br/>
        <a href="https://codecov.io/gh/mrz1836/go-mage">
          <img src="https://codecov.io/gh/mrz1836/go-mage/branch/main/graph/badge.svg?style=flat" alt="Code Coverage">
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
          <img src="https://img.shields.io/github/go-mod/go-version/mrz1836/go-mage?style=flat" alt="Go version">
        </a><br/>
        <a href="https://pkg.go.dev/github.com/mrz1836/go-mage">
          <img src="https://pkg.go.dev/badge/github.com/mrz1836/go-mage.svg?style=flat" alt="Go docs">
        </a><br/>
        <a href=".github/AGENTS.md">
          <img src="https://img.shields.io/badge/AGENTS.md-found-40b814?style=flat&logo=openai" alt="AI Agent Rules">
        </a><br/>
        <a href="https://magefile.org/">
          <img src="https://img.shields.io/badge/mage-powered-brightgreen?style=flat&logo=probot&logoColor=white" alt="Mage Powered">
        </a>
      </td>
      <td valign="top" align="left">
        <a href="https://github.com/mrz1836/go-mage/graphs/contributors">
          <img src="https://img.shields.io/github/contributors/mrz1836/go-mage?style=flat&logo=contentful&logoColor=white" alt="Contributors">
        </a><br/>
        <a href="https://github.com/sponsors/mrz">
          <img src="https://img.shields.io/badge/sponsor-mrz-181717.svg?logo=github&style=flat" alt="Sponsor">
        </a><br/>
        <a href="https://github.com/mrz1836/go-mage/stargazers">
          <img src="https://img.shields.io/github/stars/mrz1836/go-mage?style=social" alt="Stars">
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
* [Enterprise Features](#-enterprise-features)
* [Configuration](#-configuration)
* [Documentation](#-documentation)
* [Examples & Tests](#-examples--tests)
* [Benchmarks](#-benchmarks)
* [Code Standards](#-code-standards)
* [AI Compliance](#-ai-compliance)
* [Maintainers](#-maintainers)
* [Contributing](#-contributing)
* [License](#-license)

<br/>

## 🧩 What's Inside

**MAGE-X** is a comprehensive build automation toolkit that transforms how you manage Go projects. Built on the philosophy of "Write Once, Mage Everywhere," it provides enterprise-grade development tools with a delightfully friendly user experience.

Perfect for managing 30+ repositories or your first Go project, MAGE-X eliminates build boilerplate and delivers consistency across your entire development workflow.

<br/>

- **🎯 Zero-Configuration Excellence**  
  _Works out of the box with intelligent defaults. No complex setup, no YAML hell—just add one import, and you're ready to build, test, and ship._

- **🎭 Interactive Experience**  
  _Friendly CLI with guided wizards, auto-completion, and contextual help. Building software should be enjoyable, not a chore._

- **🚀 Multi-Channel Releases**  
  _Stable, beta, and edge release channels with automated versioning, GitHub integration, and asset distribution._

- **🔧 Recipe System**  
  _Pre-built patterns for common development scenarios—from fresh project setup to CI/CD pipeline configuration._

- **🛡️ Security-First Architecture**  
  _Input validation, secure command execution, and minimal dependencies. Built for environments where security matters._

- **🌍 Cross-Platform Excellence**  
  _Full support for Linux, macOS, and Windows with optimized parallel execution and CPU-aware builds._

- **🤖 AI Agent Ready**  
  _Machine-readable guidelines for ChatGPT, Claude, Cursor, and other AI assistants. Your AI follows the same house rules._

- **📊 Enterprise Features**  
  _Audit logging, compliance reporting, and team management capabilities for organizations that need governance._

<br/>

### 🚀 Quick Wins

* **One-Command Setup**: From zero to a production-ready build system in under 30 seconds
* **Intelligent Defaults**: No configuration required, but infinitely customizable when you need it
* **Multi-Project Management**: Manage 30+ repositories with consistent tooling and workflows
* **Interactive Wizards**: Guided setup for new projects, releases, and complex operations
* **Recipe Library**: Common patterns and best practices built right in

> **Tip:** Run `mage interactive` after installation to explore features with the guided wizard.

<br/>

## ⚡ Quick Start

### 1. Install MAGE-X

```bash
# Install Mage (if not already installed)
go install github.com/magefile/mage@latest

# Add MAGE-X to your project
go get github.com/mrz1836/go-mage
```

### 2. Create Your Magefile

```go
//go:build mage

package main

import (
    // Import all MAGE-X tasks
    _ "github.com/mrz1836/go-mage/pkg/mage"
)

// Default task - just run `mage`
var Default = func() error {
    return Build{}.Default()
}
```

### 3. Start Building

```bash
# Interactive mode (recommended for first use)
mage interactive

# Or dive right in
mage build          # Build your project
mage test           # Run tests with linting
mage release        # Create a release
```

### 4. Advanced Setup (Optional)

```bash
# Interactive project initialization
mage initCLI --name=myapp --module=github.com/user/myapp

# Create configuration file
mage yamlInit

# Explore available recipes
mage recipesList
```

<br/>

## 🚀 Features

### Core Excellence
- **🔧 Command Execution**: Secure, interface-based command execution with validation
- **📝 Native Logging**: Colored output, progress indicators, and structured logging
- **🛠️ Complete Build System**: All essential build, test, lint, and release tasks
- **🔄 Version Management**: Automatic version detection and update infrastructure

### Developer Experience
- **🏗️ Project Templates**: CLI, library, web API, and microservice templates
- **🚀 Multi-Channel Releases**: Stable, beta, and edge release channels
- **⚙️ Configuration Management**: Flexible mage.yaml with smart defaults
- **📦 Asset Distribution**: Automated building and distribution of release assets

### Interactive Features
- **🎭 Interactive Mode**: Friendly CLI with guided operations
- **🧙 Interactive Wizard**: Step-by-step setup for complex operations
- **📖 Help System**: Comprehensive help with auto-completion
- **🔄 Recipe System**: Common patterns and best practices library

### Enterprise Features
- **📊 Audit Logging**: Comprehensive activity tracking and compliance reporting
- **🛡️ Security Scanning**: Vulnerability detection and security policy enforcement
- **👥 Team Management**: Role-based access and team collaboration features
- **📈 Analytics**: Build metrics, performance tracking, and optimization insights

<br/>

## 🏢 Enterprise Features

MAGE-X provides comprehensive enterprise capabilities for organizations that need governance, security, and compliance in their build automation workflows.

### Enterprise Configuration Setup

Initialize enterprise features in your project:

```bash
# Initialize enterprise configuration
mage configureEnterprise

# Or set up manually
mage configureInit
```

### Audit Logging & Compliance

Track all build activities with comprehensive audit logging:

```bash
# Enable audit logging
mage auditEnable

# View audit logs
mage auditShow

# Export compliance reports
mage auditExport --format=json --output=compliance-report.json
```

**Features:**
- **Activity Tracking**: Every command, build, test, and release is logged
- **User Attribution**: Track who performed what actions and when
- **Compliance Reporting**: Generate reports for security audits and compliance
- **Retention Policies**: Configurable log retention and archival
- **Integration Support**: Export to SIEM systems and compliance tools

### Security Scanning & Policy Enforcement

Advanced security scanning with policy enforcement:

```bash
# Run comprehensive security scan
mage securityScan

# Enable vulnerability monitoring
mage securityMonitor

# Generate security report
mage securityReport --format=sarif
```

**Features:**
- **Vulnerability Detection**: Scan dependencies for known vulnerabilities
- **Policy Enforcement**: Configurable security policies and gates
- **Compliance Frameworks**: Support for SOC2, ISO 27001, and other frameworks
- **Integration Support**: GitHub Security Advisories, Snyk, and other security tools
- **Automated Remediation**: Automatic dependency updates and security patches

### Team Management & Collaboration

Enterprise-grade team management and collaboration features:

```bash
# Set up team configuration
mage teamSetup

# Configure user roles and permissions
mage teamRoles --user=john@company.com --role=developer

# Team analytics and insights
mage teamAnalytics
```

**Features:**
- **Role-Based Access**: Fine-grained permissions for team members
- **Team Analytics**: Track team productivity and build metrics
- **Collaboration Tools**: Integration with Slack, Teams, and other platforms
- **Project Templates**: Standardized project templates for consistency
- **Onboarding Workflows**: Automated team member onboarding

### Analytics & Performance Metrics

Comprehensive analytics and performance tracking:

```bash
# View build analytics
mage analyticsDashboard

# Performance metrics
mage analyticsPerformance

# Export metrics data
mage analyticsExport --timeframe=30d --format=json
```

**Features:**
- **Build Metrics**: Track build times, success rates, and performance trends
- **Resource Utilization**: Monitor CPU, memory, and disk usage
- **Dependency Analysis**: Track dependency updates and security metrics
- **Performance Optimization**: Identify bottlenecks and optimization opportunities
- **Custom Dashboards**: Create custom dashboards for different stakeholders

### Advanced CLI Features

Enterprise-focused CLI capabilities:

```bash
# Bulk operations across multiple repositories
mage cliBulk --operation=build --repos=repo1,repo2,repo3

# Advanced querying and filtering
mage cliQuery --filter="status=failed" --last=24h

# Interactive dashboard
mage cliDashboard

# Batch processing
mage cliBatch --workflow=ci-cd --parallel=5
```

**Features:**
- **Bulk Operations**: Execute commands across multiple repositories
- **Advanced Querying**: Filter and search across projects and builds
- **Interactive Dashboard**: Real-time monitoring and management
- **Batch Processing**: Efficient processing of multiple operations
- **Workflow Management**: Define and execute complex workflows

### Enterprise Workflow Engine

Define and execute complex workflows:

```bash
# Create workflow from template
mage workflowCreate --name=ci-cd --template=enterprise

# Execute workflow
mage workflowExecute --workflow=ci-cd

# Schedule workflows
mage workflowSchedule --workflow=ci-cd --cron="0 2 * * *"

# View workflow history
mage workflowHistory --workflow=ci-cd --limit=10
```

**Features:**
- **Workflow Templates**: Pre-built templates for common enterprise scenarios
- **Dependency Management**: Define step dependencies and parallel execution
- **Conditional Execution**: Execute steps based on conditions and variables
- **Scheduling**: Cron-based scheduling for automated workflows
- **History Tracking**: Complete workflow execution history and analytics

### Enterprise Integration Hub

Connect with your existing enterprise tools:

```bash
# Configure integrations
mage integrationsSetup

# Test integrations
mage integrationsTest --service=slack

# Sync with external systems
mage integrationsSync --service=jira
```

**Supported Integrations:**
- **Communication**: Slack, Microsoft Teams, Discord
- **Project Management**: Jira, Asana, Linear, GitHub Issues
- **CI/CD**: Jenkins, GitLab CI, GitHub Actions, Azure DevOps
- **Infrastructure**: Docker, Kubernetes, AWS, Azure, GCP
- **Monitoring**: Prometheus, Grafana, DataDog, New Relic
- **Security**: Vault, SIEM systems, security scanners

### Enterprise Configuration File

Create `.mage.enterprise.yaml` for advanced configuration:

```yaml
metadata:
  version: "1.0.0"
  created_at: "2024-01-01T00:00:00Z"
  updated_at: "2024-01-01T00:00:00Z"

organization:
  name: "Acme Corporation"
  domain: "acme.com"
  departments:
    - name: "Engineering"
      teams: ["backend", "frontend", "devops"]
    - name: "Security"
      teams: ["security", "compliance"]

security:
  level: "enterprise"
  policies:
    - name: "dependency-scanning"
      enabled: true
      severity: "high"
  vault_integration:
    enabled: true
    address: "https://vault.acme.com"
  compliance_frameworks:
    - "SOC2"
    - "ISO27001"

analytics:
  enabled: true
  metrics_interval: "5m"
  retention_days: 90
  export_formats: ["json", "csv", "prometheus"]

integrations:
  slack:
    enabled: true
    webhook_url: "${SLACK_WEBHOOK_URL}"
    channels: ["#builds", "#alerts"]
  jira:
    enabled: true
    url: "https://acme.atlassian.net"
    project_key: "ENG"
  vault:
    enabled: true
    address: "https://vault.acme.com"
    auth_method: "jwt"

workflows:
  templates_dir: ".mage/workflows"
  execution_timeout: "1h"
  max_concurrent_jobs: 10
  notification_on_failure: true
```

<br/>

## ⚙️ Configuration

### Basic Configuration

Create `.mage.yaml` in your project root:

```yaml
project:
  name: myproject
  binary: myapp
  version: v1.0.0
  module: github.com/user/myproject

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
  golangci_version: v1.55.2
  timeout: 5m
  
tools:
  golangci_lint: v1.55.2
  fumpt: latest
  govulncheck: latest

release:
  channels:
    - stable
    - beta
    - edge
  github:
    owner: mrz
    repo: myproject
```

### Environment Variable Overrides

```bash
# Build configuration
export BINARY_NAME=myapp
export GO_BUILD_TAGS=prod,feature1
export GOOS=linux
export GOARCH=amd64

# Test configuration
export VERBOSE=true
export TEST_RACE=true
export TEST_TIMEOUT=15m

# Tool configuration
export PARALLEL=8
export LINT_TIMEOUT=10m
export FUZZ_TIME=30s
export BENCH_TIME=10s
```

<br/>

## 📚 Documentation

For comprehensive documentation, visit the [docs](docs) directory:

- **[Getting Started](docs/README.md)** - Complete documentation index
- **[Namespace Interface Architecture](docs/NAMESPACE_INTERFACES.md)** - Modern interface-based namespace system
- **[API Reference](docs/API_REFERENCE.md)** - Complete interface and API documentation
- **[Enterprise Features](docs/ENTERPRISE.md)** - Enterprise capabilities guide
- **[Quick Start](docs/QUICK_START.md)** - Get up and running in minutes
- **[Configuration Reference](docs/CONFIGURATION.md)** - Complete configuration guide

### Available Commands

MAGE-X provides a comprehensive set of commands organized by functionality. All commands are available through the `mage` CLI.

#### 📦 Build Commands
```bash
# Core Build Operations
mage build              # Build for current platform (default)
mage buildDocker        # Build Docker containers
mage buildClean         # Clean build artifacts
mage buildGenerate      # Generate code before building

# Build System Management
mage installStdlib      # Install Go standard library for cross-compilation
```

#### 🧪 Test Commands
```bash
# Test Execution
mage test               # Run complete test suite with linting (default)
mage testUnit           # Run unit tests only (no linting)
mage testRace           # Run tests with race detector
mage testCover          # Run tests with coverage analysis
mage testBench          # Run benchmark tests
mage testFuzz           # Run fuzz tests
mage testIntegration    # Run integration tests
```

#### 🔍 Code Quality & Linting
```bash
# Linting and Code Quality
mage lint               # Run default linter (default)
mage lintAll            # Run all linting checks
mage lintFix            # Automatically fix linting issues
```

#### 📊 Metrics & Analysis
```bash
# Code Analysis
mage metricsLOC         # Analyze lines of code
mage metricsCoverage    # Generate coverage reports
mage metricsComplexity  # Analyze code complexity
```

#### 📦 Dependency Management
```bash
# Dependency Operations
mage depsUpdate         # Update all dependencies (equivalent to "make update")
mage depsTidy           # Clean up go.mod and go.sum
mage depsDownload       # Download all dependencies
mage depsOutdated       # Show outdated dependencies
mage depsAudit          # Audit dependencies for vulnerabilities
```

#### 🔧 Development Tools
```bash
# Tool Management
mage toolsUpdate        # Update all development tools
mage toolsInstall       # Install all required development tools
mage toolsCheck         # Check if all required tools are available
mage installTools       # Install development tools
mage installBinary      # Install the project binary
mage uninstall          # Remove installed binary
```

#### 📝 Module Management
```bash
# Go Module Operations
mage modUpdate          # Update go.mod file
mage modTidy            # Tidy the go.mod file
mage modVerify          # Verify module checksums
mage modDownload        # Download modules
```

#### 📚 Documentation
```bash
# Documentation Generation
mage docsGenerate       # Generate documentation
mage docsServe          # Serve documentation locally
mage docsBuild          # Build static documentation
mage docsCheck          # Validate documentation
```

#### 🔀 Git Operations
```bash
# Git Workflow
mage gitStatus          # Show git repository status
mage gitCommit          # Commit changes (interactive)
mage gitTag             # Create and push a new tag
mage gitPush            # Push changes to remote
```

#### 🏷️ Version Management
```bash
# Version Control
mage versionShow        # Display current version information
mage versionBump        # Bump the version (interactive)
mage versionCheck       # Check version information
```

#### 🚀 Release Management
```bash
# Release Operations
mage release            # Create a new release (default)
mage releaseDefault     # Create default release
```

#### 🎯 Default Targets
```bash
# Quick Access Commands
mage                    # Run default build
mage build              # Same as above
mage test               # Run complete test suite
mage lint               # Run linter
```


### 📋 Command Discovery

Discover available commands using these built-in help features:

```bash
# List all available targets
mage -l

# Get help for a specific command (if available)
mage -h <command>

# Show mage version and build info
mage -version
```

### Recipe System

MAGE-X includes a comprehensive recipe system for common development patterns:

```bash
# List all available recipes
mage recipesList

# Show recipe details
mage recipesShow fresh-start

# Run a recipe
RECIPE=fresh-start mage recipesRun

# Search for recipes
TERM=docker mage recipesSearch
```

**Available Recipes:**
- `fresh-start` - Clean project setup with best practices
- `ci-setup` - GitHub Actions CI/CD configuration
- `docker-setup` - Docker and containerization setup
- `security-hardening` - Security best practices implementation
- `performance-optimization` - Performance tuning and optimization
- `documentation-boost` - Documentation generation and maintenance

<br/>

## 🧪 Examples & Tests

All examples and tests run via GitHub Actions using Go 1.21+. View the [examples directory](examples) for complete project demonstrations.

### Run Tests

```bash
# Quick test suite
mage test

# Comprehensive testing
mage testRace testCover testFuzz

# Performance benchmarks
mage testBench
```

### Example Projects

- **[Basic Project](examples/basic)** - Minimal MAGE-X setup
- **[CLI Application](examples/cli/)** - Command-line tool with sub-commands
- **[Library](examples/library/)** - Reusable Go library
- **[Web API](examples/webapi/)** - REST API with Docker
- **[Microservice](examples/microservice/)** - Cloud-native microservice

<br/>

## ⚡ Benchmarks

Performance benchmarks for core MAGE-X operations:

| Operation             | Time  | Memory | Notes                             |
|-----------------------|-------|--------|-----------------------------------|
| Build Detection       | 1.2ms | 256KB  | Project type and configuration    |
| Command Execution     | 0.8ms | 128KB  | Secure command validation         |
| Configuration Loading | 2.1ms | 512KB  | YAML parsing and validation       |
| Recipe Processing     | 3.5ms | 1MB    | Template expansion and validation |

> Benchmarks run on Apple M1 Pro (ARM64) with Go 1.21.5
> All operations show consistent sub-5ms performance with minimal memory allocation

<br/>

## 🛠️ Code Standards

MAGE-X follows strict coding standards and best practices:

- **Code Quality**: 100% test coverage, comprehensive linting, and security scanning
- **Go Best Practices**: Idiomatic Go code following community standards
- **Security First**: Input validation, secure command execution, minimal dependencies
- **Documentation**: Comprehensive godoc coverage and usage examples
- **AI Compliance**: Machine-readable guidelines for AI assistants

Read more about our [code standards](.github/CODE_STANDARDS.md) and [contribution guidelines](.github/CONTRIBUTING.md).

<br/>

## 🤖 AI Compliance

MAGE-X includes comprehensive AI assistant guidelines:

- **[AGENTS.md](.github/AGENTS.md)** — Complete rules for coding style, workflows, and best practices
- **[CLAUDE.md](docs/CLAUDE.md)** — Specific guidelines for Claude AI integration
- **[.cursorrules](.cursorrules)** — Machine-readable policies for Cursor and similar tools
- **[sweep.yaml](.github/sweep.yaml)** — Configuration for Sweep AI code review

These files ensure that AI assistants follow the same high standards as human contributors, maintaining code quality and consistency across all contributions.

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
- **💬 Join discussions** and help other users

### Quick Start for Contributors

```bash
# Clone the repository
git clone https://github.com/mrz1836/go-mage.git
cd go-mage

# Install dependencies
go mod download

# Run tests
mage test

# Run linter
mage lint

# Start interactive mode to explore features
mage interactive
```

[![Stars](https://img.shields.io/github/stars/mrz1836/go-mage?label=Please%20like%20us&style=social)](https://github.com/mrz1836/go-mage/stargazers)

<br/>

## 📝 License

[![License](https://img.shields.io/github/license/mrz1836/go-mage.svg?style=flat)](LICENSE)

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

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
