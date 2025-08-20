# Override Commands Example

This example demonstrates how to override built-in MAGE-X commands with custom implementations while maintaining access to the original functionality.

## 🎯 When to Use Command Overrides

Override commands when you need to:
- Add custom validation or setup before/after built-in commands
- Integrate with organization-specific tools or processes
- Enforce custom coding standards or policies
- Add logging, metrics, or reporting to standard operations
- Modify behavior for specific environments or workflows

## 🔧 Override Patterns

### 1. Complete Function Override

Replace an entire command with custom logic:

```go
// Override "magex lint" completely
func LintDefault() error {
    // Custom pre-processing
    utils.Info("Running custom checks...")

    // Call original MAGE-X command
    var l mage.Lint
    if err := l.Default(); err != nil {
        return err
    }

    // Custom post-processing
    utils.Info("Custom cleanup...")
    return nil
}
```

**Usage**: `magex lint` now runs your custom version

### 2. Namespace Override

Override specific methods within a namespace:

```go
// Override build namespace
type Build struct{}

// Only override the default build
func (Build) Default() error {
    // Custom build logic
    var b mage.Build
    return b.Default() // Call original
}
// build:all, build:race, etc. are NOT available
```

**Usage**: `magex build` uses custom logic, but `magex build:all` fails

### 3. Partial Override with Embedding

Override some methods while keeping others:

```go
// Override test namespace partially
type Test struct {
    mage.Test // Embed original - keeps other commands
}

// Only override unit tests
func (Test) Unit() error {
    // Custom unit test logic
    var t mage.Test
    return t.Unit()
}
// test:race, test:cover, etc. still work from embedded original
```

**Usage**: `magex test:unit` uses custom logic, `magex test:race` uses original

## 📝 Example Commands

This example provides several override patterns:

### Overridden Commands
```bash
# Custom lint with pre/post checks
magex lint

# Custom build with environment setup
magex build

# Custom build all platforms with preparation
magex build:all

# Custom unit tests with test environment
magex test:unit

# All other test commands use originals
magex test:race     # Original mage.Test.Race()
magex test:cover    # Original mage.Test.Cover()
```

### Custom Commands
```bash
# Completely custom strict linting
magex customlintstrict
```

## 🚀 Quick Start

1. **Copy this example**:
   ```bash
   cp -r examples/override-commands/ my-project/
   cd my-project/
   ```

2. **Try the overridden commands**:
   ```bash
   # See custom lint in action
   magex lint

   # See custom build with environment variables
   MAGE_X_CUSTOM_VERSION=v2.0.0 MAGE_X_CUSTOM_BUILD_TAGS=debug magex build

   # Try custom unit tests
   magex test:unit
   ```

3. **Compare with originals**:
   ```bash
   # Test commands that use embedded originals
   magex test:race
   magex test:cover
   ```

## 💡 Key Concepts

### Override Precedence
When you define a function in your magefile.go:
1. **Local function wins**: Your `LintDefault()` overrides built-in `mage.Lint.Default()`
2. **Namespace methods win**: Your `Build.Default()` overrides `mage.Build.Default()`
3. **Embedded methods preserved**: With embedding, only overridden methods change

### Calling Originals
Always available within overrides:
```go
func LintDefault() error {
    // Your custom logic

    // Call original MAGE-X command
    var l mage.Lint
    return l.Default()
}
```

### Environment Integration
Override commands can:
- Read environment variables for configuration
- Set up custom environments
- Integrate with CI/CD specific requirements
- Apply organization-specific policies

## ⚠️ Important Considerations

### Compatibility
- **API Changes**: MAGE-X updates might change internal APIs
- **Testing**: Always test overrides after MAGE-X updates
- **Documentation**: Keep override behavior documented

### Best Practices
- **Call originals when possible**: Leverage existing functionality
- **Clear naming**: Use descriptive function names
- **Error handling**: Properly handle and propagate errors
- **Logging**: Use utils.Info/Warning/Error for consistent output

### When NOT to Override
- **Simple customization**: Use environment variables or config files instead
- **Minor changes**: Consider contributing improvements to MAGE-X
- **Complex logic**: Create separate custom commands instead

## 📚 Testing Your Overrides

Create test files to verify your custom behavior:

```bash
# Test the custom lint checks
echo "// TODO: fix this" > test.go
magex lint  # Should warn about TODOs

# Test custom build
MAGE_X_CUSTOM_VERSION=test magex build
cat BUILD_SUMMARY.md  # Should show custom version

# Test strict lint
echo 'import "fmt"; func main() { fmt.Println("test") }' > bad.go
magex customlintstrict  # Should fail on fmt.Print usage
```

## 🔗 Integration Examples

### CI/CD Integration
```go
func LintDefault() error {
    // CI-specific setup
    if os.Getenv("CI") != "" {
        setupCIEnvironment()
    }

    // Run original lint
    var l mage.Lint
    if err := l.Default(); err != nil {
        // CI-specific error reporting
        reportToCISystem(err)
        return err
    }

    return nil
}
```

### Organization Standards
```go
func (Build) Default() error {
    // Enforce organization build standards
    if err := validateOrgStandards(); err != nil {
        return err
    }

    // Apply organization-specific build flags
    applyOrgBuildConfig()

    // Run original build
    var b mage.Build
    return b.Default()
}
```

## 🎮 Advanced Usage

### Conditional Overrides
```go
func LintDefault() error {
    // Use override logic only in development
    if os.Getenv("GO_ENV") == "development" {
        return customDevLint()
    }

    // Use original in production
    var l mage.Lint
    return l.Default()
}
```

### Chain Multiple Overrides
```go
func CustomLintStrict() error {
    // Call your overridden lint first
    if err := LintDefault(); err != nil {
        return err
    }

    // Add extra strict checks
    return runStrictChecks()
}
```

---

## 🎯 Summary

Command overrides in MAGE-X provide powerful customization while maintaining access to the robust built-in functionality. Use them to:

- **Enhance** existing commands with custom logic
- **Integrate** with organization-specific workflows
- **Enforce** custom standards and policies
- **Extend** functionality without losing built-in features

Start with simple overrides and gradually add complexity as your needs grow. The key is balancing customization with maintainability and compatibility.

---

**Next**: Explore other examples in the [examples/](../) directory or check out the [with-custom/](../with-custom/) example for adding completely new commands alongside built-ins.
