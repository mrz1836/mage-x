# MAGE-X Configuration Example

This example demonstrates how to use `.mage.yaml` configuration files to customize MAGE-X behavior for your Go projects.

## 📋 What This Example Shows

- **Comprehensive Configuration**: A complete `.mage.yaml` file with all available options
- **Build Customization**: How to configure build settings, platforms, and ldflags
- **Test Configuration**: Custom test timeouts, coverage, and parallel execution
- **Tool Management**: Version control for development tools
- **Environment Overrides**: How environment variables can override file settings
- **Enterprise Features**: Configuration for advanced features

## 🗂️ Files in This Example

```
with-config/
├── README.md           # This documentation
├── .mage.yaml         # Complete configuration example
├── main.go            # Simple Go application
├── main_test.go       # Test file with various test types
└── go.mod             # Go module file
```

## 🚀 Quick Start

### 1. Navigate to This Example

```bash
cd examples/with-config
```

### 2. View Current Configuration

```bash
# Display the loaded configuration
magex configure:show

# Validate configuration syntax and settings
magex configure:validate
```

### 3. Try Basic Commands

```bash
# Build using configuration settings
magex build

# Run tests with configured settings
magex test

# Run tests with coverage (enabled in .mage.yaml)
magex test:cover

# Run benchmarks with configured duration
magex test:bench           # Uses config file bench_time setting
magex bench time=50ms      # Override with parameter (preferred)
magex bench time=10s count=3  # Multiple parameters
```

### 4. Cross-Platform Builds

The `.mage.yaml` file configures multiple platforms:

```bash
# Build for all configured platforms
magex build:all

# This will create binaries for:
# - Linux AMD64
# - macOS AMD64 
# - macOS ARM64 (Apple Silicon)
# - Windows AMD64
```

## ⚙️ Configuration Sections Explained

### Project Metadata

```yaml
project:
  name: config-example              # Project name
  binary: myapp                     # Output binary name
  version: v1.0.0                  # Project version
  module: github.com/example/config-demo  # Go module name
```

### Build Configuration

```yaml
build:
  output: bin                       # Output directory
  trimpath: true                   # Remove file paths from binary
  platforms:                       # Cross-compilation targets
    - linux/amd64
    - darwin/amd64
    - darwin/arm64
    - windows/amd64
  ldflags:                         # Linker flags for optimization
    - -s -w                        # Strip debug info
    - -X main.version={{.Version}} # Inject version
```

### Test Configuration

```yaml
test:
  parallel: 4                      # Parallel test processes
  timeout: 10m                     # Global test timeout
  race: true                       # Enable race detection
  cover: true                      # Enable coverage
  covermode: atomic               # Coverage mode
```

### Tool Versions

```yaml
tools:
  golangci_lint: v2.3.1           # Linter version
  fumpt: latest                    # Formatter version
  govulncheck: latest             # Security scanner
```

## 🌍 Environment Variable Overrides

Any configuration setting can be overridden with environment variables using the `MAGE_X_` prefix:

```bash
# Override binary name
MAGE_X_BINARY_NAME=custom-app magex build

# Override test timeout
MAGE_X_TEST_TIMEOUT=5m magex test

# Override build tags
MAGE_X_BUILD_TAGS=debug,local magex build

# Override parallel processes
MAGE_X_PARALLEL=8 magex test
```

## 📊 Configuration Priority

Settings are applied in this order (highest to lowest priority):

1. **Environment Variables** (`MAGE_X_*`)
2. **Configuration File** (`.mage.yaml`, `.mage.yml`, etc.)
3. **Smart Defaults** (built-in sensible defaults)

## 🔍 Configuration Discovery

MAGE-X automatically looks for configuration files in this order:

1. `.mage.yaml`
2. `.mage.yml`
3. `mage.yaml`
4. `mage.yml`

The first file found is used. If no file is found, smart defaults are applied.

## 🧪 Testing the Configuration

### Run Different Test Types

```bash
# Unit tests only
magex test:unit

# Integration tests (uses integration_timeout from config)
magex test:integration

# Short tests (skips long-running tests)
magex test:short

# Race detection tests (enabled in config)
magex test:race

# Coverage tests (configured for atomic mode)
magex test:cover

# Benchmark tests (uses bench_time from config)
magex test:bench              # Uses config: bench_time: "10s"
magex bench time=1s           # Override with parameter
```

### Verify Configuration

```bash
# Check if configuration is valid
magex configure:validate

# Show configuration in different formats
magex configure:show

# Initialize a new configuration (if needed)
magex configure:init
```

## 🔧 Advanced Features

### Enterprise Configuration

For advanced features, you can also create a `.mage.enterprise.yaml` file:

```bash
# Initialize enterprise configuration
magex configure:enterprise

# This creates a separate enterprise config file with:
# - Audit logging
# - Analytics
# - Security policies
# - Integration settings
```

### Docker Integration

The configuration includes Docker settings:

```bash
# Build Docker image using config settings
magex build:docker

# Uses registry, repository, and build_args from .mage.yaml
```

### Release Management

```bash
# Create release using configured platforms and formats
magex release:default

# This uses the github owner/repo and platforms from .mage.yaml
```

## 📝 Customizing for Your Project

### 1. Copy the Configuration

```bash
# Copy this example's .mage.yaml to your project
cp .mage.yaml /path/to/your/project/

# Edit it for your needs
```

### 2. Essential Settings to Update

- `project.name`: Your project name
- `project.binary`: Your binary name
- `project.module`: Your Go module path
- `project.repo_owner` and `project.repo_name`: Your GitHub repository
- `build.platforms`: Platforms you want to support
- `tools.*`: Tool versions you prefer

### 3. Optional Settings

Many settings can be left as-is or removed entirely:

```yaml
# Minimal configuration example
project:
  name: my-project
  binary: myapp
  module: github.com/me/my-project

build:
  platforms:
    - linux/amd64
    - darwin/amd64
```

## 🏆 Best Practices

### Configuration Management

1. **Start Simple**: Begin with minimal configuration and add settings as needed
2. **Version Control**: Always commit your `.mage.yaml` file
3. **Environment Specific**: Use environment variables for environment-specific settings
4. **Validation**: Regularly run `magex configure:validate` to check configuration
5. **Documentation**: Comment your configuration file for team members

### Security Considerations

1. **No Secrets**: Never put secrets or API keys in configuration files
2. **Environment Variables**: Use environment variables for sensitive settings
3. **Tool Versions**: Pin tool versions for reproducible builds
4. **Trimpath**: Enable `trimpath` to remove file system paths from binaries

### Performance Optimization

1. **Parallel Builds**: Set appropriate parallel values for your CI/CD system
2. **Build Cache**: Use build cache in CI/CD environments
3. **Test Parallelism**: Configure test parallelism based on your test suite
4. **Coverage Exclusions**: Exclude generated files from coverage

## 🔗 Related Examples

- **[Zero Config](../zero-config/)**: Using MAGE-X without any configuration
- **[Basic Usage](../basic/)**: Simple MAGE-X usage patterns
- **[With Custom Commands](../with-custom/)**: Adding custom commands alongside configuration
- **[Override Commands](../override-commands/)**: Overriding built-in commands

## 📚 Additional Resources

- [Configuration Reference](../../docs/CONFIGURATION.md): Complete configuration documentation
- [MAGE-X Documentation](../../docs/): Full project documentation
- [Environment Variables](../../docs/CONFIGURATION.md#environment-variables): Complete list of environment overrides

## 💡 Tips and Tricks

### Quick Configuration Check

```bash
# See what configuration is actually being used
magex configure:show | grep -A 5 "Build Configuration"
```

### Testing Configuration Changes

```bash
# Validate before applying changes
magex configure:validate

# Test with different settings temporarily
MAGE_X_VERBOSE=true magex build
```

### Debugging Configuration Issues

```bash
# Enable verbose output to see configuration loading
MAGE_X_VERBOSE=true magex configure:show

# Check which config file is being used
magex configure:show | head -10
```

---

This example demonstrates the full power of MAGE-X configuration while maintaining the zero-boilerplate philosophy. You can use as much or as little configuration as needed for your project!