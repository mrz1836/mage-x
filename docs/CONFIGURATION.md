# 🔧 MAGE-X Configuration Reference

This document provides a comprehensive guide to configuring MAGE-X for your Go projects.

## 📋 Table of Contents

- [Overview](#overview)
- [Configuration File](#configuration-file)
- [Project Configuration](#project-configuration)
- [Build Configuration](#build-configuration)
- [Test Configuration](#test-configuration)
- [Analytics Configuration](#analytics-configuration)
- [Security Configuration](#security-configuration)
- [Deployment Configuration](#deployment-configuration)
- [Environment Variables](#environment-variables)
- [Advanced Configuration](#advanced-configuration)
- [Configuration Validation](#configuration-validation)
- [Best Practices](#best-practices)

## 📖 Overview

MAGE-X uses a flexible configuration system that supports:

- **YAML Configuration Files**: `.mage.yaml` or `.mage.yml` in your project root
- **Environment Variable Overrides**: Override any configuration via environment variables
- **Smart Defaults**: Zero-configuration setup with intelligent defaults
- **Validation**: Built-in configuration validation and error reporting

## 📄 Configuration File

Create a `.mage.yaml` file in your project root:

```yaml
# .mage.yaml - MAGE-X Configuration File
project:
  name: my-project
  version: v1.0.0
  description: "My awesome Go project"
  authors:
    - "Your Name <your.email@example.com>"
  license: MIT
  homepage: "https://github.com/user/my-project"
  repository: "https://github.com/user/my-project.git"

build:
  go_version: "1.24"
  platform: "linux/amd64"
  tags:
    - prod
    - feature1
  ldflags: "-s -w -X main.version={{.Version}} -X main.commit={{.Commit}}"
  gcflags: ""
  cgo_enabled: false
  output_dir: "bin"
  binary: "myapp"

test:
  timeout: 600         # seconds
  coverage: true
  verbose: false
  race: false
  parallel: 4
  tags:
    - integration
  output_dir: "test-results"
  bench_time: "10s"
  mem_profile: false
  cpu_profile: false

analytics:
  enabled: false
  sample_rate: 0.1
  retention_days: 30
  export_formats:
    - json
    - csv
  endpoints:
    metrics: "https://metrics.example.com"
    events: "https://events.example.com"
  batch_size: 100
  flush_interval: 60   # seconds

security:
  enable_vuln_check: true
  skip_vuln_check:
    - "GO-2021-0001"
  required_checks:
    - gosec
    - govulncheck
  policy_file: ".security-policy.yaml"
  enable_code_scan: true
  enable_secret_scan: true

deploy:
  strategy: "rolling"
  environment: "production"
  variables:
    ENV: "prod"
    LOG_LEVEL: "info"
  hooks:
    pre_deploy:
      - "mage test:default"
      - "mage lint:default"
    post_deploy:
      - "mage health:check"
    on_failure:
      - "mage rollback"
    on_success:
      - "mage notify:success"
  rollback:
    enabled: true
    max_versions: 5
    auto_rollback: true
    health_check_url: "https://api.example.com/health"
    health_check_retry: 3
```

## 🏗️ Project Configuration

Configure project metadata and information:

```yaml
project:
  name: my-project              # Required: Project name
  version: v1.0.0             # Required: Semantic version
  description: "Description"   # Optional: Project description
  authors:                    # Optional: List of authors
    - "Name <email@example.com>"
  license: MIT                # Optional: License identifier
  homepage: "https://..."     # Optional: Project homepage
  repository: "https://..."   # Optional: Repository URL
```

### Validation Rules

- `name`: Must be a valid Go module name
- `version`: Must follow semantic versioning (v1.2.3)
- `authors`: Must include valid email addresses
- `homepage` and `repository`: Must be valid URLs

## 🔨 Build Configuration

Control how your project is built:

```yaml
build:
  go_version: "1.24"           # Go version to use
  platform: "linux/amd64"     # Target platform
  tags:                       # Build tags
    - prod
    - feature1
  ldflags: "-s -w"           # Linker flags
  gcflags: ""                # Compiler flags
  cgo_enabled: false         # Enable/disable CGO
  output_dir: "bin"          # Output directory
  binary: "myapp"            # Binary name
```

### Platform Options

Support for multiple platforms:

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`
- `windows/amd64`
- `windows/arm64`

### Common LDFlags

```yaml
ldflags: |
  -s -w
  -X main.version={{.Version}}
  -X main.commit={{.Commit}}
  -X main.buildTime={{.Date}}
  -X main.goVersion={{.GoVersion}}
```

Template variables available:
- `{{.Version}}`: Project version
- `{{.Commit}}`: Git commit hash
- `{{.Date}}`: Build timestamp
- `{{.GoVersion}}`: Go version used

## 🧪 Test Configuration

Configure testing behavior:

```yaml
test:
  timeout: 600               # Test timeout in seconds
  coverage: true             # Enable code coverage
  verbose: false             # Verbose test output
  race: false               # Enable race detector
  parallel: 4               # Number of parallel tests
  tags:                     # Test-specific build tags
    - integration
    - e2e
  output_dir: "test-results" # Test output directory
  bench_time: "10s"         # Benchmark duration
  mem_profile: false        # Enable memory profiling
  cpu_profile: false        # Enable CPU profiling
```

### Test Types

Different test configurations for different scenarios:

```yaml
# Unit tests only
test:
  timeout: 300
  coverage: true
  tags: []

# Integration tests
test:
  timeout: 1200
  coverage: false
  tags: ["integration"]

# Performance tests
test:
  timeout: 600
  bench_time: "30s"      # Default for environment variable fallback
  mem_profile: true
  cpu_profile: true

# With parameters (preferred approach):
# magex bench time=30s cpu-profile=cpu.prof mem-profile=mem.prof
```

## ⚡ Benchmark Configuration

MAGE-X supports benchmark timing through both configuration files and command-line parameters. **Parameters are the preferred approach** for better CLI usability:

### Parameter-Based Timing (Recommended)

```bash
# Quick benchmarks for CI/fast feedback
magex bench time=50ms

# Standard benchmarks  
magex bench time=100ms

# Comprehensive benchmarks
magex bench time=10s

# Multiple parameters
magex bench time=5s count=3

# Profiling benchmarks
magex bench:cpu time=30s profile=cpu-profile.out
magex bench:mem time=2s profile=mem-profile.out
magex bench:profile time=10s cpu-profile=cpu.prof mem-profile=mem.prof

# Benchmark comparison
magex bench:compare old=baseline.txt new=current.txt

# Regression testing
magex bench:regression time=5s update-baseline=true
```

### Configuration File Fallback

For backward compatibility, you can still configure default timing in `.mage.yaml`:

```yaml
test:
  bench_time: "10s"  # Default when no time parameter is provided
```

### Environment Variable Fallback

Also supports environment variables for backward compatibility:

```bash
export BENCH_TIME="10s"  # Used when no parameter or config is set
```

**Priority Order**: Parameter > Config File > Environment Variable > Default (10s)

## 📊 Analytics Configuration

Configure analytics and metrics collection:

```yaml
analytics:
  enabled: true              # Enable analytics
  sample_rate: 0.1          # Sample rate (0.0-1.0)
  retention_days: 30        # Data retention period
  export_formats:           # Export formats
    - json
    - csv
    - prometheus
  endpoints:               # Analytics endpoints
    metrics: "https://metrics.example.com/api/v1"
    events: "https://events.example.com/api/v1"
  batch_size: 100          # Batch size for exports
  flush_interval: 60       # Flush interval in seconds
```

### Privacy and Compliance

Analytics configuration supports privacy-focused settings:

```yaml
analytics:
  enabled: true
  sample_rate: 0.01          # Low sampling for privacy
  retention_days: 7          # Short retention
  anonymize_data: true       # Anonymize personal data
  respect_dnt: true          # Respect Do Not Track
```

## 🛡️ Security Configuration

Configure security scanning and policies:

```yaml
security:
  enable_vuln_check: true    # Enable vulnerability checking
  skip_vuln_check:          # Skip specific vulnerabilities
    - "GO-2021-0001"
    - "GO-2022-0001"
  required_checks:          # Required security checks
    - gosec
    - govulncheck
    - staticcheck
  policy_file: ".security-policy.yaml"  # Security policy file
  enable_code_scan: true    # Enable code scanning
  enable_secret_scan: true  # Enable secret scanning
```

### Security Policy File

Create a `.security-policy.yaml` file:

```yaml
# .security-policy.yaml
vulnerabilities:
  severity_threshold: "MEDIUM"  # Minimum severity to fail
  max_age_days: 30             # Maximum age for vulnerabilities

secrets:
  patterns:
    - "api[_-]?key"
    - "secret[_-]?key"
    - "password"
  exclude_files:
    - "*.test"
    - "testdata/**"

code_quality:
  complexity_threshold: 10
  duplication_threshold: 0.1
```

## 🚀 Deployment Configuration

Configure deployment strategies and hooks:

```yaml
deploy:
  strategy: "rolling"        # Deployment strategy
  environment: "production"  # Target environment
  variables:                # Environment variables
    ENV: "prod"
    LOG_LEVEL: "info"
    DATABASE_URL: "postgres://..."
  hooks:                    # Deployment hooks
    pre_deploy:
      - "mage test:default"
      - "mage lint:default"
      - "mage security:check"
    post_deploy:
      - "mage health:check"
      - "mage smoke:test"
    on_failure:
      - "mage rollback"
      - "mage notify:failure"
    on_success:
      - "mage notify:success"
      - "mage update:metrics"
  rollback:                 # Rollback configuration
    enabled: true
    max_versions: 5
    auto_rollback: true
    health_check_url: "https://api.example.com/health"
    health_check_retry: 3
```

### Deployment Strategies

Available deployment strategies:

- `rolling`: Rolling deployment (default)
- `blue_green`: Blue-green deployment
- `canary`: Canary deployment
- `immediate`: Immediate deployment

## 🌍 Environment Variables

Override any configuration using environment variables:

### Project Variables
```bash
export MAGE_PROJECT_NAME="my-project"
export MAGE_PROJECT_VERSION="v1.0.0"
export MAGE_PROJECT_DESCRIPTION="My project"
```

### Build Variables
```bash
export MAGE_BUILD_GO_VERSION="1.24"
export MAGE_BUILD_PLATFORM="linux/amd64"
export MAGE_BUILD_CGO_ENABLED="false"
export MAGE_BUILD_OUTPUT_DIR="dist"
export MAGE_BUILD_BINARY="myapp"
```

### Test Variables
```bash
export MAGE_TEST_TIMEOUT="600"
export MAGE_TEST_COVERAGE="true"
export MAGE_TEST_RACE="false"
export MAGE_TEST_PARALLEL="4"
```

### Security Variables
```bash
export MAGE_SECURITY_ENABLE_VULN_CHECK="true"
export MAGE_SECURITY_ENABLE_CODE_SCAN="true"
export MAGE_SECURITY_POLICY_FILE=".security-policy.yaml"
```

### Legacy Environment Variables

For backward compatibility, these variables are also supported:

```bash
# Build configuration
export BINARY_NAME="myapp"
export GO_BUILD_TAGS="prod,feature1"
export GOOS="linux"
export GOARCH="amd64"

# Test configuration
export VERBOSE="true"
export TEST_RACE="true"
export TEST_TIMEOUT="15m"

# Tool configuration
export PARALLEL="8"
export LINT_TIMEOUT="10m"

# Benchmark timing (for backward compatibility)
export BENCH_TIME="10s"

# Preferred: Use parameters instead
magex bench time=50ms         # Quick benchmarks
magex bench time=10s          # Standard benchmarks
magex bench:cpu time=30s      # CPU profiling with custom duration
```

## ⚙️ Advanced Configuration

### Multi-Environment Configuration

Support for environment-specific configurations:

```yaml
# .mage.yaml
project:
  name: my-project
  version: v1.0.0

# Environment-specific overrides
environments:
  development:
    build:
      tags: ["dev", "debug"]
    test:
      verbose: true
      race: true
    security:
      enable_vuln_check: false

  staging:
    build:
      tags: ["staging"]
    analytics:
      enabled: true
      sample_rate: 0.5

  production:
    build:
      tags: ["prod", "release"]
      ldflags: "-s -w -X main.version={{.Version}}"
    security:
      enable_vuln_check: true
      required_checks: ["gosec", "govulncheck", "staticcheck"]
```

### Configuration Includes

Include external configuration files:

```yaml
# .mage.yaml
includes:
  - "configs/build.yaml"
  - "configs/security.yaml"
  - "configs/deploy-${ENVIRONMENT}.yaml"

project:
  name: my-project
  version: v1.0.0
```

### Dynamic Configuration

Use templates and functions in configuration:

```yaml
project:
  name: "{{ .Env.PROJECT_NAME | default \"my-project\" }}"
  version: "{{ .Git.Tag | default \"v0.0.0-dev\" }}"

build:
  ldflags: |
    -s -w
    -X main.version={{ .Project.Version }}
    -X main.commit={{ .Git.Commit }}
    -X main.buildTime={{ .Time.RFC3339 }}
```

## ✅ Configuration Validation

MAGE-X validates configuration automatically:

### Validation Rules

- **Project**: Name and version are required
- **Build**: Go version format validation
- **Test**: Timeout must be positive
- **Analytics**: Sample rate must be 0.0-1.0
- **Security**: Policy file must exist if specified

### Custom Validation

Add custom validation rules:

```yaml
validation:
  rules:
    - name: "version_format"
      condition: "project.version matches '^v\\d+\\.\\d+\\.\\d+$'"
      message: "Version must follow semantic versioning"
    - name: "required_tags"
      condition: "len(build.tags) > 0"
      message: "At least one build tag is required"
```

## 📚 Best Practices

### 1. Use Environment-Specific Configurations

```yaml
# Base configuration
project:
  name: my-project

# Environment overrides
environments:
  development:
    security:
      enable_vuln_check: false
  production:
    security:
      enable_vuln_check: true
      required_checks: ["gosec", "govulncheck"]
```

### 2. Secure Sensitive Data

Never store secrets in configuration files:

```yaml
# ❌ Don't do this
deploy:
  variables:
    DATABASE_PASSWORD: "secret123"

# ✅ Do this instead
deploy:
  variables:
    DATABASE_PASSWORD: "${DATABASE_PASSWORD}"
```

### 3. Use Meaningful Defaults

Provide sensible defaults that work for most projects:

```yaml
test:
  timeout: 600     # 10 minutes - enough for most tests
  coverage: true   # Enable by default
  parallel: 4      # Good balance of speed vs resources
```

### 4. Document Custom Configuration

Add comments to explain custom settings:

```yaml
build:
  # Custom ldflags for embedded version info
  ldflags: "-s -w -X main.version={{.Version}}"

  # Enable CGO for database drivers
  cgo_enabled: true

  # Production-specific build tags
  tags:
    - prod        # Production optimizations
    - netgo       # Pure Go networking
    - osusergo    # Pure Go user lookup
```

### 5. Validate Configuration

Always validate configuration in CI/CD:

```bash
# In your CI pipeline
mage config:validate

# Or check specific aspects
mage config:check --security
mage config:check --build
```

### 6. Version Your Configuration

Track configuration changes with your code:

```yaml
# .mage.yaml
config_version: "2.1.0"  # Track configuration schema version

project:
  name: my-project
  version: v1.0.0
```

## 🔗 Related Documentation

- [Quick Start Guide](QUICK_START.md)
- [Enterprise Features](ENTERPRISE.md)
- [API Reference](API_REFERENCE.md)
- [Namespace Interfaces](NAMESPACE_INTERFACES.md)

---

For more information and examples, visit the [MAGE-X GitHub repository](https://github.com/mrz1836/mage-x).
