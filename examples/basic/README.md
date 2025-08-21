# Basic MAGE-X Example

This example demonstrates basic MAGE-X usage with the `magex` binary - **ZERO SETUP REQUIRED!**

## 🎯 Key Point: NO magefile.go Needed!

Unlike traditional Mage which requires a magefile.go with wrapper functions, MAGE-X's `magex` binary provides all commands instantly.

## Quick Start

```bash
# NO magefile.go needed! Just run magex commands directly:

magex build         # Build the project
magex test          # Run tests
magex lint          # Run linter
magex format:fix    # Format code
```

## Common Commands

### Building
```bash
magex build           # Default build
magex build:linux     # Linux build
magex build:darwin    # macOS build
magex build:windows   # Windows build
magex build:all       # All platforms
```

### Testing
```bash
magex test            # Run tests
magex test:unit       # Unit tests only
magex test:race       # With race detector
magex test:cover      # With coverage
magex test:bench      # Run benchmarks
magex bench time=50ms # Quick benchmarks with parameter
```

### Code Quality
```bash
magex lint            # Run linter
magex lint:fix        # Auto-fix issues
magex format          # Check formatting
magex format:fix      # Fix formatting
```

### Dependencies
```bash
magex deps            # Check dependencies
magex deps:update     # Update dependencies (safe - no major version bumps)
magex deps:update allow-major  # Update including major versions (v1→v2, etc)
magex deps:tidy       # Clean up go.mod
magex deps:audit      # Security audit
```

## Complete Workflow Example

Here's a typical development workflow:

```bash
# 1. Format your code
magex format:fix

# 2. Run linter and fix issues
magex lint:fix

# 3. Run tests
magex test:race

# 4. Build the project
magex build

# 5. Build for all platforms
magex build:all
```

## CI Pipeline Example

```bash
#!/bin/bash
# ci.sh - Complete CI pipeline

set -e  # Exit on error

echo "🔍 Checking format..."
magex format:check

echo "🔍 Running linter..."
magex lint:ci

echo "🧪 Running tests..."
magex test:race
magex test:cover

echo "🔨 Building..."
magex build:all

echo "✅ CI Pipeline Complete!"
```

## Discovering Commands

```bash
# List all available commands
magex -l

# List commands by namespace
magex -n

# Search for specific commands
magex -search test
magex -search build

# Get help
magex -h
```

## Key Benefits

1. **Zero Setup** - No magefile.go needed
2. **241+ Commands** - All available instantly
3. **Consistent Interface** - Same commands across all projects
4. **No Maintenance** - Updates automatically with MAGE-X

## Sample Project Structure

```
my-project/
├── main.go
├── go.mod
├── go.sum
└── (NO magefile.go needed - this is the magic!)
```

**Just run `magex` commands in your project directory - that's it!**

## Why This Works

- **Traditional Mage**: Requires writing 200+ lines of wrapper functions
- **MAGE-X magex**: Zero setup, all 241+ commands work instantly
- **True zero-boilerplate**: No configuration files, no imports, no wrappers

## Try It Now

```bash
# In this directory, try:
magex -l              # See all available commands
magex build           # Build (if there's Go code)
magex test            # Run tests (if there are test files)
```

---

**Remember:** With MAGE-X's `magex` binary, you get all the power of make/rake/gradle but with zero configuration!
