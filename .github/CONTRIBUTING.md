# 🤝 Contributing to mage-x

Thank you for your interest in contributing to mage-x! This document provides guidelines to help you contribute effectively while maintaining code quality and consistency.

## 📋 Quick Reference

- **Before starting**: Check [existing issues](../../issues) and [pull requests](../../pulls)
- **Code style**: Use `make lint` and `make fmt` before submitting
- **Tests required**: All changes must include tests and pass `make test-race`
- **Documentation**: Update relevant docs and add examples for new features

## 📦 How to Contribute

1. **Fork the repo** and create a feature branch
2. **Install pre-commit hooks**:
   ```bash
   pip install pre-commit
   pre-commit install
   ```
3. **Development setup**:
   ```bash
   git clone https://github.com/mrz1836/mage-x.git
   cd mage-x
   go mod download
   ```
4. **Make changes following guidelines** (see sections below)
5. **Write comprehensive tests** and ensure they pass
6. **Open a pull request** with clear description of changes

More info on [pull requests](http://help.github.com/pull-requests/).

<br/>

## 🔄 Code Duplication Prevention

**Important**: mage-x follows strict anti-duplication practices. Before writing new code, always check for existing utilities and patterns.

### Shared Utilities

Before implementing common functionality, check these packages:

#### 🧪 Testing Utilities (`internal/testutil`)
- **Mock utilities**: `ValidateArgs()`, `ExtractResult()`, `ExtractError()`
- **File helpers**: `CreateTestFiles()`, `WriteTestFile()`, `CreateTestDirectory()`
- **Test patterns**: `TestCase[T,R]`, `RunTableTests()`, `BenchmarkCase`
- **Assertions**: `AssertNoError()`, `AssertError()`, `AssertEqual()`

```go
// ✅ Use shared test patterns
tests := []testutil.TestCase[string, int]{
    {Name: "valid input", Input: "42", Expected: 42, WantErr: false},
}
testutil.RunTableTests(t, tests, func(t *testing.T, tc testutil.TestCase[string, int]) {
    result, err := processInput(tc.Input)
    testutil.AssertNoError(t, err)
    testutil.AssertEqual(t, tc.Expected, result)
})

// ❌ Don't duplicate test patterns
for _, tt := range []struct{name string; input string; want int; wantErr bool}{...} {
    t.Run(tt.name, func(t *testing.T) {
        // repeated assertion patterns
    })
}
```

#### 🔧 Error Handling (`internal/errors`)
- **Wrapping**: `WrapWithContext()` for consistent error context
- **Validation**: `InvalidField()`, `EmptyField()`, `RequiredField()`
- **Commands**: `CommandFailed()` for command execution errors
- **Security**: `PathTraversal()` for path validation errors

```go
// ✅ Use standardized error utilities
if err := processFile(path); err != nil {
    return errors.WrapWithContext(err, "process config file")
}

if name == "" {
    return errors.EmptyField("repository name")
}

// ❌ Don't create ad-hoc errors
return fmt.Errorf("failed to process config file: %w", err)
return fmt.Errorf("repository name cannot be empty")
```

#### ✅ Validation (`internal/validation`)
- **Repository**: `ValidateRepository()` for owner/repo format
- **Paths**: `ValidatePath()` with security checks
- **Branches**: `ValidateBranch()` for Git branch names
- **Batch validation**: `ValidateFields()` for multiple checks

```go
// ✅ Use centralized validation
if err := validation.ValidateRepository(repo); err != nil {
    return err
}

// ❌ Don't duplicate validation logic
if !regexp.MustCompile(`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+$`).MatchString(repo) {
    return fmt.Errorf("invalid repository format")
}
```

#### 🏃 Benchmarking (`internal/benchmark`)
- **Memory tracking**: `WithMemoryTracking()` for consistent measurement
- **File setup**: `SetupBenchmarkFiles()`, `SetupBenchmarkRepo()`
- **Standardized sizes**: `StandardSizes()` for scaling tests
- **Data generation**: `CreateBenchmarkData()` for test data

```go
// ✅ Use shared benchmark patterns
func BenchmarkOperation(b *testing.B) {
    files := benchmark.SetupBenchmarkFiles(b, tempDir, 100)
    
    benchmark.WithMemoryTracking(b, func() {
        for i := 0; i < b.N; i++ {
            processFiles(files)
        }
    })
}

// ❌ Don't duplicate benchmark setup
func BenchmarkOperation(b *testing.B) {
    b.ReportAllocs()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // manual setup and teardown
    }
    b.StopTimer()
}
```

#### 📄 JSON Processing (`internal/jsonutil`)
- **Type-safe operations**: `MarshalJSON[T]()`, `UnmarshalJSON[T]()`
- **Test data**: `GenerateTestJSON()` for consistent test data
- **Formatting**: `PrettyPrint()`, `CompactJSON()`
- **Merging**: `MergeJSON()` for combining objects

```go
// ✅ Use type-safe JSON utilities
config, err := jsonutil.UnmarshalJSON[Config](data)
if err != nil {
    return err
}

// ❌ Don't use manual JSON handling
var config Config
if err := json.Unmarshal(data, &config); err != nil {
    return fmt.Errorf("failed to unmarshal config: %w", err)
}
```

### Pre-Contribution Checklist

Before writing new code, ask yourself:

1. **Does similar functionality already exist?**
   - Check `internal/` packages for utilities
   - Search for similar patterns: `grep -r "similar_pattern" internal/`

2. **Can existing utilities be extended?**
   - Add functions to existing packages vs. creating new ones
   - Follow established patterns and naming conventions

3. **Is this a one-off or reusable pattern?**
   - If used in 2+ places, extract to shared utility
   - Consider future use cases when designing APIs

4. **Does it follow security best practices?**
   - Use validation utilities for user input
   - Wrap errors with proper context
   - Follow path traversal prevention patterns

### Code Review Guidelines

When reviewing code, check for:

- **Duplication**: Could this use existing utilities?
- **Patterns**: Does this follow established conventions?
- **Testing**: Are shared test utilities used appropriately?
- **Errors**: Are errors wrapped with proper context?
- **Security**: Are inputs validated using shared validators?

## 🧪 Testing Requirements

### Test Coverage
All tests follow standard Go patterns. We love:

* ✅ [Go Tests](https://golang.org/pkg/testing/)
* 📘 [Go Examples](https://golang.org/pkg/testing/#hdr-Examples)
* ⚡ [Go Benchmarks](https://golang.org/pkg/testing/#hdr-Benchmarks)

- **Unit tests**: All new functions must have tests
- **Integration tests**: Add to `test/integration/` for end-to-end scenarios
- **Benchmarks**: Performance-critical code needs benchmarks
- **Fuzz tests**: Add fuzz tests for parser/validation code

Tests should be:
- Easy to understand
- Focused on one behavior
- Fast

Use `require` over `assert` where possible (we lint for this).

This project aims for >= **90% code coverage**. Every code path must be tested to keep the Codecov badge green and CI passing.

### Test Patterns
```go
// ✅ Table-driven tests with shared utilities
func TestNewFunction(t *testing.T) {
    tests := []testutil.TestCase[Input, Output]{
        {Name: "description", Input: input, Expected: output, WantErr: false},
    }
    
    testutil.RunTableTests(t, tests, func(t *testing.T, tc testutil.TestCase[Input, Output]) {
        result, err := NewFunction(tc.Input)
        if tc.WantErr {
            testutil.AssertError(t, err)
        } else {
            testutil.AssertNoError(t, err)
            testutil.AssertEqual(t, tc.Expected, result)
        }
    })
}
```

### Development Commands
```bash
make test          # Fast tests with linting
make test-race     # Tests with race detection
make test-cover    # Tests with coverage report
make lint          # Run all linters
make fmt           # Format code
make fumpt         # Advanced formatting
```

<br/>

## 🧹 Coding Conventions

We follow [Effective Go](https://golang.org/doc/effective_go.html), plus:

* 📖 [godoc](https://godoc.org/golang.org/x/tools/cmd/godoc)
* 🧼 [golangci-lint](https://golangci-lint.run/)
* 🧾 [Go Report Card](https://goreportcard.com/)

Format your code with `gofmt`, lint with `golangci-lint`, and keep your diffs minimal.

### Common Patterns

#### Error Handling
```go
// ✅ Consistent error patterns
func ProcessFile(path string) error {
    if err := validation.ValidatePath(path); err != nil {
        return err
    }
    
    data, err := os.ReadFile(path)
    if err != nil {
        return errors.WrapWithContext(err, "read file")
    }
    
    return processData(data)
}
```

#### File Operations in Tests
```go
// ✅ Use testutil for file operations
func TestFileProcessing(t *testing.T) {
    tempDir := testutil.CreateTestDirectory(t)
    files := testutil.CreateTestFiles(t, tempDir, 5)
    
    for _, file := range files {
        err := ProcessFile(file)
        testutil.AssertNoError(t, err)
    }
}
```

#### Configuration Validation
```go
// ✅ Batch validation pattern
func ValidateConfig(cfg *Config) error {
    return validation.ValidateFields(map[string]func() error{
        "repository": func() error { return validation.ValidateRepository(cfg.Repo) },
        "branch":     func() error { return validation.ValidateBranch(cfg.Branch) },
        "webhook":    func() error { return validation.ValidateURL(cfg.WebhookURL) },
    })
}
```

## 🛡️ Security Guidelines

1. **Input validation**: Always validate user input using shared validators
2. **Path traversal**: Use `validation.ValidatePath()` for file paths
3. **Error context**: Don't expose sensitive information in error messages
4. **Dependencies**: Keep dependencies up to date

<br/>

## 🚀 Pull Request Process

1. **Create feature branch**:
   ```bash
   git checkout -b feature/description
   ```

2. **Make changes following guidelines**:
   - Use existing utilities where possible
   - Add comprehensive tests
   - Update documentation

3. **Verify quality**:
   ```bash
   make test-race    # All tests pass
   make lint         # No lint errors
   make bench        # Benchmarks run successfully
   ```

4. **Submit PR**:
   - Clear title describing the change
   - Reference any related issues
   - Include examples of new functionality

### PR Template
```markdown
## Summary
Brief description of changes and motivation.

## Changes
- List of specific changes made
- New utilities added/used
- Tests added

## Testing
- [ ] Unit tests pass (`make test`)
- [ ] Race tests pass (`make test-race`)
- [ ] Lint checks pass (`make lint`)
- [ ] Benchmarks run successfully

## Documentation
- [ ] Updated relevant README files
- [ ] Added code examples
- [ ] Updated API documentation
```

## 🤝 Community Guidelines

- **Be respectful**: Follow the [Code of Conduct](../CODE_OF_CONDUCT.md)
- **Ask questions**: Use [GitHub Discussions](../../discussions) for questions
- **Report issues**: Use the issue template for bug reports
- **Share knowledge**: Help others with code reviews and discussions

## 📞 Getting Help

- **Documentation**: Check the [README](../README.md) and package docs
- **Issues**: Search [existing issues](../../issues) first
- **Discussions**: Use [GitHub Discussions](../../discussions) for questions
- **Code examples**: Check the [examples directory](../examples/)

## 📚 More Guidance

For detailed workflows, commit standards, branch naming, PR templates, and more—read [AGENTS.md](./AGENTS.md). It's the rulebook.

<br/>

Thank you for contributing to mage-x! Let's build something great. 💪 🎉
