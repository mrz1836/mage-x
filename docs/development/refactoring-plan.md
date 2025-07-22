# Refactoring Plan: Eliminate Code Duplication in Go Mage

## Overview
Refactor the Go mage codebase to eliminate duplication by creating reusable, testable, and mockable packages while ensuring zero linter errors and all tests pass.

## Phase 1: Create Common Utility Packages with Interfaces

### 1.1 File Operations Package (`pkg/common/fileops`)

**Interface Design for Mockability:**
```go
type FileOperator interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte, perm os.FileMode) error
    Exists(path string) bool
    MkdirAll(path string, perm os.FileMode) error
    Remove(path string) error
}

type JSONOperator interface {
    Marshal(v interface{}) ([]byte, error)
    Unmarshal(data []byte, v interface{}) error
    WriteJSON(path string, v interface{}) error
    ReadJSON(path string, v interface{}) error
}

type YAMLOperator interface {
    Marshal(v interface{}) ([]byte, error)
    Unmarshal(data []byte, v interface{}) error
    WriteYAML(path string, v interface{}) error
    ReadYAML(path string, v interface{}) error
}
```

**Implementation:**
```go
- DefaultFileOperator (real implementation)
- MockFileOperator (for testing)
- SafeFileOps wrapper with retry logic
- Atomic write operations
```

### 1.2 Configuration Package (`pkg/common/config`)

**Interface Design:**
```go
type ConfigLoader interface {
    Load(paths []string, v interface{}) error
    Save(path string, v interface{}, format string) error
    Validate(v interface{}) error
}

type EnvProvider interface {
    Get(key string) string
    Set(key, value string) error
    LookupEnv(key string) (string, bool)
}
```

**Features:**
- Pluggable config sources (file, env, remote)
- Schema validation
- Hot-reload capability
- Migration support

### 1.3 Environment Package (`pkg/common/env`)

**Interface Design:**
```go
type Environment interface {
    GetString(key, defaultValue string) string
    GetBool(key string, defaultValue bool) bool
    GetInt(key string, defaultValue int) int
    GetDuration(key string, defaultValue time.Duration) time.Duration
    GetStringSlice(key string, defaultValue []string) []string
}

type PathResolver interface {
    GOPATH() string
    ConfigDir(appName string) string
    DataDir(appName string) string
    CacheDir(appName string) string
}
```

### 1.4 Path Builder Package (`pkg/common/paths`)

**Interface Design:**
```go
type PathBuilder interface {
    MageDir(subdir ...string) string
    ConfigPath(filename string) string
    DataPath(category, filename string) string
    BackupPath(id string) string
    TempPath(prefix string) string
}

type PathValidator interface {
    IsValid(path string) bool
    IsSecure(path string) bool
    Normalize(path string) string
}
```

### 1.5 Error Handling Package (`pkg/common/errors`)

**Enhanced Error Types:**
```go
type MageError interface {
    error
    Code() string
    Context() map[string]interface{}
    Wrap(message string) MageError
    WithContext(key string, value interface{}) MageError
}

type ErrorHandler interface {
    Handle(err error) error
    HandleWithRetry(fn func() error, maxRetries int) error
    IsRetryable(err error) bool
}
```

## Phase 2: Testing Infrastructure

### 2.1 Test Helpers Package (`pkg/testhelpers`)

```go
type TestEnvironment struct {
    FileOps    *MockFileOperator
    ConfigOps  *MockConfigLoader
    EnvOps     *MockEnvironment
    PathOps    *MockPathBuilder
}

func NewTestEnvironment(t testing.TB) *TestEnvironment
func (e *TestEnvironment) Cleanup()
```

### 2.2 Mock Generators
- Use mockgen for interface mocking
- Pre-generate common mocks
- Provide mock builders for complex scenarios

### 2.3 Test Fixtures
```go
pkg/testdata/
├── configs/
│   ├── valid/
│   └── invalid/
├── files/
└── scenarios/
```

## Phase 3: Refactoring Strategy

### 3.1 Dependency Injection Pattern
```go
// Before
func LoadConfig() (*Config, error) {
    data, err := os.ReadFile(".mage.yaml")
    // ...
}

// After
type ConfigService struct {
    fileOps FileOperator
    loader  ConfigLoader
}

func NewConfigService(fileOps FileOperator, loader ConfigLoader) *ConfigService

func (s *ConfigService) LoadConfig() (*Config, error) {
    // Uses injected dependencies
}
```

### 3.2 Factory Pattern for Commands
```go
type CommandFactory interface {
    CreateRunner(config RunnerConfig) CommandRunner
    CreateBuilder(config BuildConfig) Builder
    CreateTester(config TestConfig) Tester
}
```

## Phase 4: Implementation Plan

### Week 1: Foundation
1. Create all interface definitions
2. Implement core packages with default implementations
3. Generate mocks for all interfaces
4. Write comprehensive unit tests

### Week 2: Migration Wave 1
1. Migrate file operations (est. 300+ call sites)
2. Migrate configuration loading (est. 50+ call sites)
3. Update affected tests

### Week 3: Migration Wave 2
1. Migrate command runners
2. Migrate error handling
3. Performance testing

### Week 4: Polish & Documentation
1. Add integration tests
2. Create migration guide
3. Update all documentation
4. Performance optimization

## Phase 5: Quality Assurance

### 5.1 Linting Configuration (`.golangci.yml`)
```yaml
run:
  timeout: 5m
  tests: true

linters:
  enable:
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck
    - gocritic
    - gocyclo
    - gosec
    - dupl
    - misspell
    - unparam
    - gochecknoinits
    - goconst
    - gofmt
    - goimports
    - revive

linters-settings:
  govet:
    check-shadowing: true
  gocyclo:
    min-complexity: 15
  dupl:
    threshold: 100
  goconst:
    min-len: 2
    min-occurrences: 3
```

### 5.2 Test Coverage Requirements
- Minimum 80% coverage for new packages
- 100% coverage for error paths
- Integration tests for all major workflows

### 5.3 Benchmarking
- Benchmark before and after refactoring
- Ensure no performance regression
- Document any performance improvements

## Phase 6: Rollout Strategy

### 6.1 Feature Flags
```go
type FeatureFlags interface {
    IsEnabled(feature string) bool
    Enable(feature string)
    Disable(feature string)
}

// Gradual rollout
if flags.IsEnabled("use-new-fileops") {
    // New implementation
} else {
    // Legacy implementation
}
```

### 6.2 Compatibility Layer
- Maintain backward compatibility
- Deprecation warnings for old patterns
- Migration tools for existing projects

## Success Metrics

1. **Code Quality**
   - ✅ Zero linter errors
   - ✅ 85%+ test coverage
   - ✅ All tests pass with -race flag
   - ✅ Benchmarks show no regression

2. **Code Reduction**
   - ✅ 40-50% reduction in duplicated code
   - ✅ 30% reduction in total LOC
   - ✅ 60% improvement in cyclomatic complexity

3. **Developer Experience**
   - ✅ Easy to mock all external dependencies
   - ✅ Clear separation of concerns
   - ✅ Intuitive API design
   - ✅ Comprehensive documentation

4. **Maintainability**
   - ✅ Single source of truth for common operations
   - ✅ Consistent error handling
   - ✅ Standardized testing patterns

## Risk Mitigation

1. **Breaking Changes**
   - Use interfaces for all public APIs
   - Maintain compatibility layer
   - Extensive testing before release

2. **Performance Impact**
   - Benchmark critical paths
   - Use sync.Pool for frequent allocations
   - Profile memory usage

3. **Learning Curve**
   - Provide examples for all patterns
   - Create video tutorials
   - Pair programming sessions

## File Tracking

This plan will be saved to:
- `refactoring-plan.md` - Main plan document
- `refactoring-progress.md` - Progress tracking
- `refactoring-decisions.md` - Architecture decisions record

The plan includes testability and mockability as core requirements, ensuring easy testing for developers and maintaining high code quality throughout the refactoring process.

## Duplication Analysis Summary

Based on analysis of the codebase, the following duplication patterns were identified:

1. **File Operations** - 20+ instances of similar file I/O patterns
2. **JSON/YAML Operations** - 70+ instances of marshal/unmarshal patterns  
3. **Configuration Loading** - Multiple implementations of config loading with fallback
4. **Error Handling** - 522+ instances of similar error formatting
5. **Environment Variables** - Repeated GOPATH resolution and env var patterns
6. **Path Building** - 25+ instances of similar filepath.Join patterns
7. **Logging** - 765 instances that could benefit from structured logging

This refactoring will address all these patterns systematically while maintaining backward compatibility and improving testability.