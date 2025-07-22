# Backward Compatibility Guide

Go-Mage provides comprehensive backward compatibility for existing magefiles, ensuring a smooth transition to the new interface-based architecture.

## Quick Start

To enable backward compatibility, simply import the compatibility package in your existing magefile:

```go
// +build mage

package main

import (
    _ "github.com/mrz1836/go-mage/compat"
)

// Your existing targets will continue to work
func Build() error {
    return sh("go", "build", "-o", "myapp", ".")
}
```

## Automatic Features

When you import the compatibility package, you automatically get:

### 1. Legacy Target Aliases
All common target names are automatically registered:
- `build` → `Build:Default`
- `test` → `Test:Unit`
- `lint` → `Lint:All`
- `clean` → `Clean:All`
- `docker` → `Docker:Build`
- `deps` → `Deps:Download`
- And many more...

### 2. Legacy Helper Functions
```go
// Shell command execution
err := Sh("go", "build", ".")

// Must helper (panics on error)
Must(err)

// Must run (combines Sh and Must)
MustRun("go", "test", "./...")

// File operations
if Exists("config.yaml") {
    // ...
}

files := Glob("*.go")
```

### 3. Legacy Environment Variables
The following environment variables are automatically supported:
- `MAGEFILE_VERBOSE` → `MAGE_VERBOSE`
- `MAGEFILE_DEBUG` → `MAGE_DEBUG`
- `MAGEFILE_GOCMD` → `MAGE_GOCMD`

### 4. Legacy Configuration Files
Automatically loads configuration from:
- `.mage.yml`
- `mage.yaml`
- `Magefile.yml`
- `mage.yml`

## Migration

### Automatic Migration

Set the environment variable to enable auto-migration:

```bash
MAGE_AUTO_MIGRATE=true mage
```

This will:
1. Backup your existing files
2. Convert legacy configuration to new format
3. Update target definitions
4. Show migration summary

### Manual Migration

Run the migration command:

```bash
mage migrate
```

Or check for compatibility issues:

```bash
mage compat:check
```

### Step-by-Step Migration

1. **Check Compatibility**
   ```bash
   mage compat:check
   ```

2. **Review Migration Plan**
   ```bash
   mage migrate --dry-run
   ```

3. **Run Migration**
   ```bash
   mage migrate
   ```

4. **Verify Results**
   ```bash
   mage build
   mage test
   ```

## Legacy Namespace Support

If you have custom namespaces, they're automatically wrapped:

```go
// Old style
type Build mg.Namespace

func (Build) Linux() error {
    // ...
}

// Automatically works with new system
```

## Configuration Migration

### Old Format (.mage.yml)
```yaml
project_name: myapp
project_version: 1.0.0
build_output: dist
build_flags:
  - -v
  - -ldflags
  - "-s -w"
test_coverage: true
lint_tools:
  - golangci-lint
  - golint
```

### New Format (.mage.yaml)
```yaml
project:
  name: myapp
  version: 1.0.0
  binary: myapp

build:
  output: dist
  flags:
    - -v
  ldflags:
    - -s
    - -w

test:
  coverage: true
  coverage_file: coverage.out

lint:
  tools:
    - name: golangci-lint
      command: golangci-lint
      args: [run]
```

## Common Migration Scenarios

### Scenario 1: Simple Build Target

**Before:**
```go
func Build() error {
    return sh("go", "build", "-o", "myapp", ".")
}
```

**After (automatic):**
Works as-is with compatibility layer.

**After (migrated):**
```go
func (Build) Default() error {
    return utils.RunCmd("go", "build", "-o", "myapp", ".")
}
```

### Scenario 2: Cross-Platform Builds

**Before:**
```go
func BuildLinux() error {
    os.Setenv("GOOS", "linux")
    return sh("go", "build", "-o", "myapp-linux", ".")
}
```

**After (migrated):**
```go
func (Build) Linux() error {
    return Build{}.Platform("linux/amd64")
}
```

### Scenario 3: Complex Dependencies

**Before:**
```go
func CI() {
    mg.Deps(Test, Lint, Build)
}
```

**After (automatic):**
Works as-is with compatibility layer.

**After (migrated):**
```go
func (CI) All() error {
    mg.Deps(Test{}.All, Lint{}.All, Build{}.Default)
    return nil
}
```

## Gradual Migration

You can migrate gradually by:

1. Start with compatibility import
2. Migrate configuration file
3. Update targets one by one
4. Remove compatibility import when done

## Troubleshooting

### Issue: Legacy targets not found
**Solution:** Ensure you're importing the compatibility package:
```go
import _ "github.com/mrz1836/go-mage/compat"
```

### Issue: Configuration not loading
**Solution:** Check file name and location. Run:
```bash
mage compat:check
```

### Issue: Environment variables not working
**Solution:** The compatibility layer maps common variables. For custom variables, update your code to use the new env package.

## Best Practices

1. **Use Auto-Migration for New Projects**
   - Faster and more reliable
   - Generates optimal configuration

2. **Test After Migration**
   - Run all targets to ensure they work
   - Check CI/CD pipelines

3. **Update Documentation**
   - Update README with new target names
   - Document any custom namespaces

4. **Gradual Rollout**
   - Migrate development first
   - Test thoroughly
   - Roll out to production

## Getting Help

If you encounter issues during migration:

1. Run `mage compat:check` for diagnostics
2. Check the [migration guide](https://github.com/mrz1836/go-mage/wiki/Migration)
3. Open an issue with your legacy magefile

The compatibility layer ensures your existing magefiles continue to work while providing a smooth path to adopt the new features and improvements.
