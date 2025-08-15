# Zero Configuration Example

This example demonstrates MAGE-X's revolutionary zero-boilerplate approach.
No magefile.go required - all commands work immediately!

## ✨ The Magic

Traditional Mage requires you to write wrapper functions for every command:
```go
// magefile.go - The OLD way (90+ functions needed!)
func Build() error { var b mage.Build; return b.Default() }
func Test() error { var t mage.Test; return t.Default() }
func Lint() error { var l mage.Lint; return l.Default() }
// ... 87 more wrapper functions 😱
```

With MAGE-X's `magex` binary:
```bash
# No magefile.go needed! Just run:
magex build
magex test
magex lint
# All 90+ commands available instantly! 🎉
```

## 🚀 Quick Start

1. **Install MAGE-X**
```bash
go install github.com/mrz1836/mage-x/cmd/magex@latest
```

2. **Use it immediately in any Go project**
```bash
# No setup, no configuration, just run:
magex build         # Build your project
magex test          # Run tests
magex lint:fix      # Fix linting issues
magex release:multi # Multi-platform release
```

## 📋 Available Commands

Run `magex -l` to see all 90+ built-in commands:

```bash
$ magex -l

🎯 Available Commands (94 total):

  build              build:all          build:linux
  build:darwin       build:windows      build:docker
  build:clean        build:install      build:generate
  build:prebuild     test               test:unit
  test:short         test:race          test:cover
  test:coverrace     test:coverreport   test:coverhtml
  test:fuzz          test:bench         test:integration
  test:ci            test:parallel      lint
  lint:fix           lint:ci            lint:fast
  format             format:check       format:fix
  deps               deps:update        deps:tidy
  deps:audit         git:status         git:commit
  release            release:multi      docs:build
  tools:install      ... and many more!
```

## 🎨 Namespace Commands

Commands are organized in logical namespaces:

```bash
# Build namespace
magex build           # Default build
magex build:linux     # Linux build
magex build:docker    # Docker build

# Test namespace
magex test            # Default tests
magex test:unit       # Unit tests only
magex test:cover      # With coverage

# Lint namespace
magex lint            # Run linter
magex lint:fix        # Auto-fix issues
```

## 🔧 Adding Custom Commands

Want to add project-specific commands? Just create a `magefile.go`:

```go
//go:build mage
package main

// Your custom command - works alongside all MAGE-X commands!
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

## 🎯 Real-World Example

Here's a complete CI pipeline with zero setup:

```bash
#!/bin/bash
# ci.sh - Complete CI pipeline, no magefile needed!

# Install magex (if not already installed)
go install github.com/mrz1836/mage-x/cmd/magex@latest

# Run complete CI pipeline
magex format:check    # Check formatting
magex lint:ci        # Strict linting
magex test:race      # Tests with race detector
magex test:cover     # Coverage report
magex build:all      # Multi-platform builds
magex release:multi  # Create release artifacts
```

## 📊 Comparison

| Aspect | Traditional Mage | MAGE-X |
|--------|-----------------|---------|
| Setup Required | Write 90+ wrapper functions | None |
| Lines of Code | 200+ lines of boilerplate | 0 lines |
| Time to Start | 30+ minutes | 0 seconds |
| Maintenance | Update wrappers for new commands | Automatic |
| Command Discovery | Read documentation | `magex -l` |

## 🌟 Key Benefits

1. **Instant Productivity** - Start building immediately
2. **Zero Maintenance** - New MAGE-X commands automatically available
3. **Progressive Enhancement** - Add custom commands when needed
4. **Full Compatibility** - Works with existing magefiles
5. **Intelligent Help** - Built-in command discovery and documentation

## 🎉 Try It Now!

```bash
# In any Go project directory:
magex build && magex test

# That's it! You're using MAGE-X with zero configuration!
```

## 📖 Learn More

- Run `magex -h` for help
- Run `magex -l` for all commands
- Run `magex -n` for commands by namespace
- Run `magex -search <term>` to find specific commands

---

**MAGE-X: Write Once, Mage Everywhere** - Now truly living up to its promise!
