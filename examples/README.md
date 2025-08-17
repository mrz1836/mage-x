# MAGE-X Examples

This directory contains practical examples demonstrating MAGE-X's zero-boilerplate approach and extensibility.

## Directory Structure

```
examples/
├── README.md                 # This file
├── zero-config/             # Zero configuration example (no magefile needed!)
│   └── README.md           # Demonstrates instant magex usage
├── basic/                   # Basic usage example
│   └── README.md           # Simple project setup
├── with-config/             # Configuration example using .mage.yaml
│   ├── README.md           # Complete configuration guide
│   ├── .mage.yaml          # Comprehensive configuration example
│   ├── main.go             # Sample Go application
│   ├── main_test.go        # Test file with various test types
│   └── go.mod              # Go module file
├── with-custom/            # Custom commands alongside built-in commands
│   ├── README.md           # How to add custom commands
│   └── magefile.go         # Custom commands example
└── override-commands/       # Override built-in commands with custom logic
    ├── README.md           # Command override patterns and examples
    └── magefile.go         # Override examples
```

## Quick Start

### Zero Configuration (Recommended)

The easiest way to use MAGE-X - no setup required:

```bash
# Step 1: Install magex
go install github.com/mrz1836/mage-x/cmd/magex@latest

# Step 2: Auto-update to latest stable release
magex update:install

# Use it immediately in any Go project
magex build         # Build your project
magex test          # Run tests
magex lint:fix      # Fix linting issues
```

See [zero-config/](zero-config/) for the complete example.

### Adding Custom Commands

Want project-specific commands alongside the 343+ built-in ones?

```go
//go:build mage
package main

// Custom command - works with all MAGE-X built-ins!
func Deploy() error {
    // Your deployment logic
    return nil
}
```

Now you have both:
```bash
magex build    # Built-in MAGE-X command
magex deploy   # Your custom command
```

See [with-custom/](with-custom/) for the complete example.

### Configuration-Driven Development

Want to customize build settings, test configuration, or tool versions?

```yaml
# .mage.yaml
project:
  name: my-project
  binary: myapp

build:
  platforms:
    - linux/amd64
    - darwin/amd64
    - windows/amd64
  ldflags:
    - -s -w
    - -X main.version={{.Version}}

test:
  timeout: 10m
  race: true
  cover: true
```

Now all MAGE-X commands use your configuration:
```bash
magex build:all   # Builds for all configured platforms
magex test        # Uses timeout and race detection from config
```

See [with-config/](with-config/) for the complete example.

## Example Categories

### 1. Zero Configuration (`zero-config/`)
**Best for:** Getting started immediately
- No magefile.go required
- All 343+ commands available instantly
- Perfect for standard Go projects

### 2. Basic Usage (`basic/`)
**Best for:** Understanding MAGE-X fundamentals
- Simple project structure
- Common command usage
- Standard workflows

### 3. Configuration Example (`with-config/`)
**Best for:** Customizing MAGE-X behavior
- Complete `.mage.yaml` configuration reference
- Build customization and cross-platform compilation
- Test configuration and coverage settings
- Tool version management
- Environment variable overrides

### 4. With Custom Commands (`with-custom/`)
**Best for:** Projects needing custom automation
- Add project-specific commands
- Works alongside all built-in commands
- Progressive enhancement approach

### 5. Override Commands (`override-commands/`)
**Best for:** Customizing built-in command behavior
- Override specific built-in commands (like `magex lint`)
- Add custom pre/post processing
- Maintain access to original functionality
- Integrate with organization-specific workflows

## Running Examples

Each example includes its own README with specific instructions:

```bash
# Navigate to any example
cd examples/zero-config

# Follow the README instructions
cat README.md
```

## Key Concepts Demonstrated

### Zero Boilerplate
Traditional Mage requires writing wrapper functions:
```go
// OLD WAY - Not needed with magex!
func Build() error { var b mage.Build; return b.Default() }
func Test() error { var t mage.Test; return t.Default() }
// ... 90+ more wrapper functions
```

With magex:
```bash
# Just run commands directly!
magex build
magex test
# All 343+ commands work instantly
```

### Progressive Enhancement
Start with zero configuration, add custom commands only when needed:
1. Begin with `magex` for instant access to all commands
2. Add a `magefile.go` only when you need custom commands
3. Custom commands work seamlessly with built-ins

### Command Discovery
```bash
magex -l              # List all commands
magex -n              # List by namespace
magex -search test    # Search for commands
```

## Common Workflows

### CI/CD Pipeline
```bash
# Complete pipeline with zero setup
magex lint:ci
magex test:race
magex test:cover
magex build:all
magex release:multi
```

### Development Workflow
```bash
magex format:fix      # Format code
magex lint:fix        # Fix linting issues
magex test:unit       # Run unit tests
magex build           # Build project
```

### Release Process
```bash
magex test:all        # Run all tests
magex build:all       # Build for all platforms
magex release:multi   # Create release artifacts
```

## Getting Help

- Run `magex -h` for help
- Run `magex help:<namespace>` for namespace-specific help
- Check individual example READMEs for detailed instructions

## Philosophy

MAGE-X follows the principle of **"Write Once, Mage Everywhere"**:
- Zero boilerplate to get started
- All commands available immediately
- Add custom commands only when needed
- Full backward compatibility

---

**Note:** These examples demonstrate MAGE-X with the `magex` binary. This is the recommended way to use MAGE-X, providing instant access to all commands without any setup.
