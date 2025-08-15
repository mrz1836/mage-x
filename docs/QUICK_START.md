# 🚀 MAGE-X Quick Start Guide

> **MAGE-X**: Write Once, Mage Everywhere - True Zero-Boilerplate Build Automation

## The New Way: Zero Configuration! 🎉

MAGE-X now provides **true zero-boilerplate** automation through the `magex` binary. No more wrapper functions or complex setup!

### ⚡ 30-Second Setup

```bash
# 1. Install magex binary (once)
go install github.com/mrz1836/mage-x/cmd/magex@latest

# 2. Use immediately in ANY Go project!
cd your-go-project
magex build         # Build your project
magex test          # Run tests
magex lint:fix      # Fix linting issues
magex release:multi # Multi-platform release

# That's it! No magefile.go needed! 🚀
```

### 📋 See What's Available

```bash
magex -l            # List all 215+ commands
magex -n            # Commands organized by namespace
magex -search test  # Find specific commands
magex -h            # Get help
```

### 🔍 Command Examples

```bash
# Building
magex build                 # Build for current platform
magex build:linux          # Build for Linux
magex build:multiplatform  # Build for all platforms

# Testing
magex test                  # Run all tests
magex test:unit            # Run unit tests only
magex test:integration     # Run integration tests
magex test:coverage        # Generate coverage reports

# Linting & Quality
magex lint                 # Run linter
magex lint:fix            # Auto-fix linting issues
magex format              # Format code
magex vet                 # Run go vet

# Dependencies
magex deps:update         # Update dependencies
magex deps:audit          # Security audit
magex mod:tidy            # Clean up go.mod

# Release Management
magex release:patch       # Create patch release
magex release:minor       # Create minor release
magex release:major       # Create major release
```

## Optional: Custom Commands

Only create a `magefile.go` if you need **project-specific** commands:

### 1. Initialize Custom Magefile

```bash
magex -init              # Creates magefile.go template
```

### 2. Add Your Custom Commands

```go
//go:build mage
package main

// Deploy deploys your application (custom command)
func Deploy() error {
    // Your custom deployment logic
    fmt.Println("Deploying application...")
    return nil
}

// Custom build with special flags
func BuildSpecial() error {
    // Custom build logic
    return nil
}
```

### 3. Use Both Built-in and Custom Commands

```bash
magex build         # Built-in MAGE-X command
magex test          # Built-in MAGE-X command
magex deploy        # Your custom command
magex buildspecial  # Your custom command
```

## 📂 Project Structure Examples

### Zero-Config Project (Recommended)
```
my-go-project/
├── go.mod
├── go.sum
├── main.go
└── pkg/
    └── ...
# No magefile.go needed! Just use: magex build
```

### With Custom Commands
```
my-go-project/
├── go.mod
├── go.sum
├── main.go
├── magefile.go      # Only for custom commands
└── pkg/
    └── ...
```

## 🏗️ Real-World Examples

### New Go CLI Project
```bash
mkdir my-cli
cd my-cli
go mod init github.com/user/my-cli

# Create main.go
cat > main.go << EOF
package main

import "fmt"

func main() {
    fmt.Println("Hello, CLI!")
}
EOF

# Build and test immediately!
magex build
magex test
magex lint
```

### Existing Project Migration
```bash
cd existing-project

# No changes needed! Just start using magex
magex build         # Works immediately
magex test          # Works immediately

# Optional: Remove old Mage wrapper functions from magefile.go
# Keep only your project-specific custom commands
```

## 🆚 Migration from Traditional Mage

### Before (Traditional Mage)
```go
//go:build mage
package main

import (
    "github.com/mrz1836/mage-x/pkg/mage"
)

// Build wrapper function (boilerplate!)
func Build() error {
    b := mage.NewBuildNamespace()
    return b.Build()
}

// Test wrapper function (boilerplate!)
func Test() error {
    t := mage.NewTestNamespace()
    return t.Test()
}

// 90+ more wrapper functions... 😓
```

### After (MAGE-X with magex)
```bash
# No magefile.go needed!
magex build    # Just works!
magex test     # Just works!
```

**Or** with custom commands:
```go
//go:build mage
package main

// Keep only YOUR custom commands!
func Deploy() error {
    // Your project-specific logic
    return nil
}
```

## 🎛️ Advanced Configuration

### Project-Specific Configuration
Create `.magex.yaml` for project-specific settings:

```yaml
# .magex.yaml (optional)
build:
  binary_name: "my-app"
  output_dir: "./dist"

test:
  timeout: "5m"
  tags: ["integration", "unit"]

lint:
  enable: ["golint", "gofmt", "govet"]
```

### Global Configuration
Configure magex globally in `~/.magex/config.yaml`:

```yaml
# ~/.magex/config.yaml
defaults:
  verbose: true
  parallel: true

integrations:
  github:
    token: "${GITHUB_TOKEN}"
```

## 📖 Next Steps

1. **Explore Commands**: `magex -l` to see all available commands
2. **Namespace Deep-Dive**: `magex -n` to see commands by category
3. **Search & Discovery**: `magex -search <term>` to find specific commands
4. **Advanced Features**: Check out [enterprise features](ENTERPRISE.md)
5. **Customization**: Learn about [configuration options](CONFIGURATION.md)

## 🚀 Pro Tips

- **Aliases**: Use `alias mx=magex` for shorter commands
- **Discovery**: Use `magex -search` to find exactly what you need
- **Help**: Every command has help: `magex help <command>`
- **Parallel**: Most commands support parallel execution automatically
- **CI/CD**: Use `magex` in GitHub Actions, Jenkins, etc.

## ❓ Need Help?

- 📚 [Full Documentation](../README.md)
- 🏗️ [Architecture Guide](NAMESPACE_INTERFACES.md)
- 🏢 [Enterprise Features](ENTERPRISE.md)
- 🤖 [AI Agent System](SUB_AGENTS.md)
- 💬 [GitHub Issues](https://github.com/mrz1836/mage-x/issues)
