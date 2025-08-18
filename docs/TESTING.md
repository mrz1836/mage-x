# Testing Guide for MAGE-X

This guide covers testing best practices, patterns, and examples for the MAGE-X project.

## Table of Contents

1. [Testing Philosophy](#testing-philosophy)
2. [Test Structure](#test-structure)
3. [Testing Patterns](#testing-patterns)
4. [Running Tests](#running-tests)
5. [Writing Tests](#writing-tests)
6. [Mocking](#mocking)
7. [Integration Tests](#integration-tests)
8. [Benchmarks](#benchmarks)
9. [Coverage](#coverage)
10. [CI/CD](#cicd)

## Testing Philosophy

Following the guidelines from AGENTS.md:

- **Comprehensive Coverage**: Each module should have comprehensive tests
- **Table-Driven Tests**: Use table-driven tests for testing multiple scenarios
- **Mock External Commands**: Unit tests should mock external commands
- **Tagged Integration Tests**: Integration tests should be build-tagged
- **Clear Test Names**: Use descriptive test names that explain what is being tested

## Test Structure

```
pkg/
├── mage/
│   ├── build.go
│   ├── build_test.go         # Unit tests for build module
│   ├── config.go
│   ├── config_test.go        # Unit tests for config module
│   └── testhelpers_test.go   # Shared test helpers
└── utils/
    ├── utils.go
    └── utils_test.go         # Unit tests for utilities
```

## Testing Patterns

### 1. Table-Driven Tests

```go
func TestParsePlatform(t *testing.T) {
    tests := []struct {
        name        string
        input       string
        wantOS      string
        wantArch    string
        wantErr     bool
    }{
        {
            name:     "valid linux/amd64",
            input:    "linux/amd64",
            wantOS:   "linux",
            wantArch: "amd64",
            wantErr:  false,
        },
        {
            name:     "invalid format",
            input:    "linux-amd64",
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            platform, err := ParsePlatform(tt.input)

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.wantOS, platform.OS)
                assert.Equal(t, tt.wantArch, platform.Arch)
            }
        })
    }
}
```

### 2. Test Suites with Setup/Teardown

```go
type BuildTestSuite struct {
    suite.Suite
    tempDir string
    build   Build
}

func (suite *BuildTestSuite) SetupTest() {
    suite.tempDir = suite.T().TempDir()
    // Setup code
}

func (suite *BuildTestSuite) TearDownTest() {
    // Cleanup code
}

func TestBuildTestSuite(t *testing.T) {
    suite.Run(t, new(BuildTestSuite))
}
```

### 3. Test Helpers

```go
// Create a test project structure
func NewTestProject(t *testing.T) *TestProject {
    t.Helper()

    tp := &TestProject{
        Dir: t.TempDir(),
        t:   t,
    }

    tp.CreateGoModule("test/project")
    tp.CreateMainGo()

    return tp
}
```

### 4. Environment Variable Testing

```go
func TestEnvOverrides(t *testing.T) {
    // Save and restore env var
    oldVal := os.Getenv("BINARY_NAME")
    defer os.Setenv("BINARY_NAME", oldVal)

    os.Setenv("BINARY_NAME", "testbin")

    cfg, _ := GetConfig()
    assert.Equal(t, "testbin", cfg.Project.Binary)
}
```

## Running Tests

### Basic Test Commands

```bash
# Run all tests
mage test

# Run with coverage
mage test:cover

# Run with race detector
mage test:race

# Run short tests only
mage test:short

# Run specific package tests
go test ./pkg/mage

# Run specific test
go test -run TestBuildDefault ./pkg/mage

# Run tests verbosely
VERBOSE=1 mage test
```

### Integration Tests

```bash
# Run integration tests
mage test:integration

# Run with build tag
go test -tags integration ./...
```

### Benchmarks

```bash
# Run all benchmarks
mage test:bench

# Run specific benchmark
go test -bench=BenchmarkBuildFlags ./pkg/mage

# Run with custom time
mage test:bench time=30s

# Short
mage test:benchshort
```

### Fuzz Tests

```bash
# Run fuzz tests
mage test:fuzz

# Run quick fuzz tests (5s default)
mage test:fuzzshort

# Run specific fuzz test
go test -fuzz=FuzzReverse -fuzztime=10s ./pkg/utils

# Override fuzz time
FUZZ_TIME=30s mage test:fuzzshort
```

## Writing Tests

### Unit Test Example

```go
func TestBuild_Default(t *testing.T) {
    // Create test project
    tp := NewTestProject(t)
    defer tp.Cleanup()

    // Setup project files
    tp.CreateGoModule("test/project")
    tp.CreateMainGo()
    tp.CreateMageConfig(DefaultMageConfig())

    // Execute build
    build := Build{}
    err := build.Default()

    // Assert results
    assert.NoError(t, err)
    tp.AssertFileExists("bin/testapp")
}
```

### Mock Example

```go
type MockRunner struct {
    mock.Mock
}

func (m *MockRunner) RunCmd(name string, args ...string) error {
    arguments := m.Called(name, args)
    return arguments.Error(0)
}

func TestWithMock(t *testing.T) {
    mockRunner := new(MockRunner)
    mockRunner.On("RunCmd", "go", mock.Anything).Return(nil)

    // Use mock in test
    err := mockRunner.RunCmd("go", "build")

    assert.NoError(t, err)
    mockRunner.AssertExpectations(t)
}
```

### Integration Test Example

```go
//go:build integration
// +build integration

func TestBuildIntegration(t *testing.T) {
    SkipIfShort(t)
    SkipIfNoMage(t)

    tp := NewTestProject(t)
    defer tp.Cleanup()

    // Setup real project
    tp.CreatePackageStructure()
    tp.GitInit()

    // Run real mage command
    err := tp.RunMage("build")

    assert.NoError(t, err)
    tp.AssertFileExists("bin/testapp")
}
```

## Mocking

### Command Execution Mocking

```go
func TestCommandExecution(t *testing.T) {
    runner := NewMockCommandRunner(t)

    // Setup expectations
    runner.AddCommand("go version", "go version go1.24.0 linux/amd64", nil)
    runner.AddCommand("go build -o bin/app", "", nil)

    // Execute test
    version, _ := runner.RunCmdOutput("go", "version")
    assert.Contains(t, version, "go1.24.0")

    err := runner.RunCmd("go", "build", "-o", "bin/app")
    assert.NoError(t, err)

    // Verify calls
    runner.AssertCalled(t, "go version")
    runner.AssertCalled(t, "go build -o bin/app")
}
```

### File System Mocking

```go
func TestFileOperations(t *testing.T) {
    // Use temp directory for isolation
    tempDir := t.TempDir()

    // Create test files
    testFile := filepath.Join(tempDir, "test.txt")
    err := os.WriteFile(testFile, []byte("content"), 0644)
    require.NoError(t, err)

    // Test file operations
    assert.True(t, FileExists(testFile))
}
```

## Coverage

### Generate Coverage Report

```bash
# Generate coverage file
mage test:cover

# View coverage in terminal
mage test:coverreport

# Generate HTML report
mage test:coverhtml

# Check coverage threshold
go test -coverprofile=coverage.out ./... && \
  go tool cover -func=coverage.out | \
  grep total | \
  awk '{print $3}' | \
  sed 's/%//'
```

### Coverage Requirements

- Aim for >80% coverage for core modules
- Critical paths should have >90% coverage
- Utility functions should have >95% coverage

## CI/CD

### GitHub Actions Example

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go: ['1.24', '1.22']

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}

      - name: Install Mage
        run: go install github.com/magefile/mage@latest

      - name: Run Tests
        run: mage test:ci

      - name: Upload Coverage
        uses: codecov/codecov-action@v5
        with:
          file: ./coverage.txt
```

## Best Practices

### 1. Test Naming

```go
// Good: Descriptive test names
func TestConfig_LoadFromFile_WithValidYAML(t *testing.T)
func TestBuild_Platform_LinuxAMD64(t *testing.T)

// Bad: Unclear names
func TestConfig1(t *testing.T)
func TestBuildLinux(t *testing.T)
```

### 2. Assertions

```go
// Use testify for clear assertions
assert.Equal(t, expected, actual, "Binary name should match")
assert.NoError(t, err, "Build should succeed")
require.NotNil(t, cfg, "Config must not be nil")

// Prefer require for critical checks
require.NoError(t, err) // Test stops here if error
assert.Equal(t, "value", result) // Test continues even if fails
```

### 3. Test Data

```go
// Use test fixtures for common data
var testConfig = `
project:
  name: test
  binary: testapp
`

// Or use helper functions
func defaultTestConfig() *Config {
    return &Config{
        Project: ProjectConfig{
            Name:   "test",
            Binary: "testapp",
        },
    }
}
```

### 4. Parallel Tests

```go
func TestParallelSafe(t *testing.T) {
    t.Parallel() // Mark test as parallel-safe

    // Don't use shared state
    // Don't change global variables
    // Use unique temp directories
}
```

### 5. Skip Conditions

```go
func TestRequiresDocker(t *testing.T) {
    if !utils.CommandExists("docker") {
        t.Skip("Docker not available")
    }
    // Test docker functionality
}
```

## Debugging Tests

### Verbose Output

```bash
# Run with verbose output
go test -v ./pkg/mage

# Run with mage verbose
VERBOSE=1 mage test
```

### Debug Single Test

```bash
# Run single test with output
go test -run TestBuildDefault -v ./pkg/mage

# With Delve debugger
dlv test ./pkg/mage -- -test.run TestBuildDefault
```

### Test Timing

```bash
# Show test timing
go test -v ./... | grep -E "^--- (PASS|FAIL)"

# Profile slow tests
go test -cpuprofile cpu.prof -memprofile mem.prof ./pkg/mage
```

## Troubleshooting

### Common Issues

1. **Tests fail on CI but pass locally**
   - Check for platform-specific code
   - Verify environment variables
   - Check file path separators

2. **Flaky tests**
   - Add proper synchronization
   - Use deterministic test data
   - Avoid time-dependent tests

3. **Slow tests**
   - Use t.Parallel() where possible
   - Mock expensive operations
   - Use test short mode

### Test Maintenance

- Review and update tests when changing functionality
- Remove obsolete tests
- Keep test code clean and readable
- Document complex test scenarios

## Summary

Good tests in MAGE-X:
- Are comprehensive and cover edge cases
- Use clear, descriptive names
- Follow consistent patterns
- Mock external dependencies appropriately
- Run quickly and reliably
- Provide helpful failure messages

Remember: Tests are documentation for how your code should work!
