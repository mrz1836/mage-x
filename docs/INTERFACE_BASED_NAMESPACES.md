# Interface-Based Namespaces in Go-Mage

## Overview

The Go-Mage project has been refactored to use an interface-based architecture for all namespace operations. This provides better testability, flexibility, and extensibility while maintaining backward compatibility with existing magefiles.

## Benefits

### 1. **Testability**
- All namespace operations can be easily mocked for unit testing
- Dependencies are injected, making tests isolated and fast
- No need for complex test setup or file system operations

### 2. **Flexibility**
- Custom implementations can be provided for any namespace
- Easy to extend or override default behavior
- Support for different environments (local, CI, cloud)

### 3. **Extensibility**
- Third-party plugins can implement namespace interfaces
- Easy to add new features without modifying core code
- Support for multiple implementations of the same interface

### 4. **Maintainability**
- Clear contracts defined by interfaces
- Separation of concerns between interface and implementation
- Easier to understand and modify code

## Architecture

### Core Components

1. **Namespace Interfaces** (`namespace_interfaces.go`)
   - Define contracts for each namespace (Build, Test, Lint, etc.)
   - Each interface method corresponds to a mage target

2. **Namespace Registry** (`namespace_interfaces.go`)
   - Central registry for all namespace implementations
   - Allows runtime replacement of implementations
   - Provides default implementations

3. **Namespace Wrappers** (`namespace_wrappers.go`)
   - Adapt existing `mg.Namespace` structs to interfaces
   - Maintain backward compatibility

4. **Common Packages** (`pkg/common/`)
   - Reusable, mockable interfaces for common operations
   - Config loading, environment, file operations, paths, errors

## Usage Examples

### Basic Usage (Backward Compatible)

Existing magefiles continue to work without modification:

```go
// +build mage

package main

import (
    "github.com/mrz1836/go-mage/pkg/mage"
)

// Build namespace
type Build mage.Build

// Test namespace  
type Test mage.Test
```

### Using Interfaces Directly

```go
// +build mage

package main

import (
    "github.com/mrz1836/go-mage/pkg/mage"
)

// Use the interface-based implementation
func Build() error {
    registry := mage.GetNamespaceRegistry()
    return registry.Build().Default()
}

// Cross-platform build
func BuildAll() error {
    registry := mage.GetNamespaceRegistry()
    return registry.Build().All()
}
```

### Custom Implementation

Create a custom build implementation with additional features:

```go
// +build mage

package main

import (
    "time"
    "github.com/mrz1836/go-mage/pkg/mage"
    "github.com/mrz1836/go-mage/pkg/common/config"
    "github.com/mrz1836/go-mage/pkg/common/env"
    "github.com/mrz1836/go-mage/pkg/common/fileops"
)

// customBuildImpl adds caching and notifications
type customBuildImpl struct {
    mage.BuildNamespace // Embed default implementation
    cache   *buildCache
    notifier *buildNotifier
}

func NewCustomBuild() mage.BuildNamespace {
    defaultBuild := mage.NewBuildImpl(
        config.NewFileLoader(".mage.yaml"),
        env.NewOSEnvironment(),
        fileops.NewLocalFileOperator(),
        mage.GetRunner(),
    )
    
    return &customBuildImpl{
        BuildNamespace: defaultBuild,
        cache:         newBuildCache(),
        notifier:      newBuildNotifier(),
    }
}

// Override Default to add caching
func (b *customBuildImpl) Default() error {
    // Check cache
    if b.cache.IsValid() {
        log.Println("Using cached build")
        return nil
    }
    
    // Notify build start
    b.notifier.NotifyStart()
    
    // Run actual build
    err := b.BuildNamespace.Default()
    
    // Update cache on success
    if err == nil {
        b.cache.Update()
    }
    
    // Notify completion
    b.notifier.NotifyComplete(err)
    
    return err
}

// Register custom implementation
func init() {
    registry := mage.GetNamespaceRegistry()
    registry.SetBuild(NewCustomBuild())
}
```

### Testing with Mocks

```go
package main

import (
    "testing"
    "github.com/mrz1836/go-mage/pkg/mage"
    "github.com/mrz1836/go-mage/pkg/common/config"
    "github.com/mrz1836/go-mage/pkg/common/env"
    "github.com/mrz1836/go-mage/pkg/common/fileops"
)

func TestBuildDefault(t *testing.T) {
    // Create mocks
    mockConfig := config.NewMockConfigLoader()
    mockEnv := env.NewMockEnvironment()
    mockFiles := fileops.NewMockFileOperator()
    mockRunner := &mockCommandRunner{}
    
    // Configure mocks
    mockConfig.SetData(&mage.Config{
        Project: mage.ProjectConfig{
            Binary: "test-app",
        },
        Build: mage.BuildConfig{
            Output: "dist",
        },
    })
    
    // Create build with mocks
    build := mage.NewBuildImpl(mockConfig, mockEnv, mockFiles, mockRunner)
    
    // Test build
    err := build.Default()
    if err != nil {
        t.Fatalf("Build failed: %v", err)
    }
    
    // Verify command was executed
    if len(mockRunner.commands) != 1 {
        t.Error("Expected one command to be executed")
    }
}
```

## Available Interfaces

### Build Operations
```go
type BuildNamespace interface {
    Default() error
    All() error
    Platform(platform string) error
    Linux() error
    Darwin() error
    Windows() error
    Docker() error
    Clean() error
    Install() error
    Generate() error
    PreBuild() error
}
```

### Test Operations
```go
type TestNamespace interface {
    Default() error
    Unit() error
    Short() error
    Race() error
    Cover() error
    CoverRace() error
    CoverReport() error
    CoverHTML() error
    Fuzz() error
    Bench(params ...string) error
    Integration() error
    CI() error
    Parallel() error
    NoLint() error
    CINoRace() error
    Run() error
    Coverage(args ...string) error
    Vet() error
    Lint() error
    Clean() error
    All() error
}
```

### Other Namespaces
- `LintNamespace` - Code linting operations
- `FormatNamespace` - Code formatting operations
- `DepsNamespace` - Dependency management
- `GitNamespace` - Git operations
- `ReleaseNamespace` - Release management
- `DocsNamespace` - Documentation generation
- `DeployNamespace` - Deployment operations
- `ToolsNamespace` - Tool management
- `SecurityNamespace` - Security scanning
- `GenerateNamespace` - Code generation
- `CLINamespace` - CLI operations

## Common Package Interfaces

### Config Loading
```go
type ConfigLoader interface {
    Load() (interface{}, error)
    LoadInto(target interface{}) error
    Watch(callback func(interface{})) error
    Reload() error
    GetPath() string
    SetPath(path string)
    Validate() error
}
```

### Environment
```go
type Environment interface {
    Get(key string, defaultValue string) string
    GetRequired(key string) (string, error)
    Set(key, value string) error
    Unset(key string) error
    All() map[string]string
    Expand(value string) string
    Has(key string) bool
    GetBool(key string, defaultValue bool) bool
    GetInt(key string, defaultValue int) int
    GetDuration(key string, defaultValue time.Duration) time.Duration
}
```

### File Operations
```go
type FileOperator interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte, perm os.FileMode) error
    AppendFile(path string, data []byte, perm os.FileMode) error
    Exists(path string) bool
    IsDir(path string) bool
    IsFile(path string) bool
    MkdirAll(path string, perm os.FileMode) error
    RemoveAll(path string) error
    Copy(src, dst string) error
    Move(src, dst string) error
    Walk(root string, fn filepath.WalkFunc) error
    Glob(pattern string) ([]string, error)
    Stat(path string) (os.FileInfo, error)
    Chmod(path string, mode os.FileMode) error
    CreateTemp(dir, pattern string) (*os.File, error)
    CreateTempDir(dir, pattern string) (string, error)
}
```

### Path Building
```go
type PathBuilder interface {
    Join(elements ...string) PathBuilder
    Dir() PathBuilder
    Base() string
    Ext() string
    Clean() PathBuilder
    Abs() (PathBuilder, error)
    String() string
    Exists() bool
    IsDir() bool
    IsFile() bool
    Create() error
    CreateDir() error
    Remove() error
    // ... and many more
}
```

### Error Handling
```go
type MageError interface {
    error
    Code() ErrorCode
    Severity() Severity
    Context() ErrorContext
    Cause() error
    WithCode(code ErrorCode) MageError
    WithSeverity(severity Severity) MageError
    WithContext(ctx ErrorContext) MageError
    WithField(key string, value interface{}) MageError
    Format(includeStack bool) string
    Is(target error) bool
    As(target interface{}) bool
}
```

## Migration Guide

### Step 1: Update Imports
```go
// Old
import "github.com/mrz1836/go-mage/pkg/utils"

// New - also import common packages
import (
    "github.com/mrz1836/go-mage/pkg/utils"
    "github.com/mrz1836/go-mage/pkg/common/config"
    "github.com/mrz1836/go-mage/pkg/common/env"
    "github.com/mrz1836/go-mage/pkg/common/errors"
)
```

### Step 2: Use Interfaces for Testing
```go
// Instead of direct file operations
data, err := os.ReadFile("config.yaml")

// Use fileops interface
files := fileops.NewLocalFileOperator()
data, err := files.ReadFile("config.yaml")
```

### Step 3: Enhanced Error Handling
```go
// Old
if err != nil {
    return fmt.Errorf("build failed: %w", err)
}

// New - structured errors
if err != nil {
    return errors.WithCode(errors.ErrBuildFailed, "build failed").
        WithCause(err).
        WithField("platform", platform)
}
```

### Step 4: Configuration Management
```go
// Old
cfg, err := GetConfig()

// New - with interface
loader := config.NewFileLoader(".mage.yaml")
cfg := &Config{}
err := loader.LoadInto(cfg)
```

## Best Practices

1. **Use Interfaces for New Code**
   - Define interfaces for new functionality
   - Implement with testability in mind

2. **Mock Dependencies in Tests**
   - Use mock implementations for all external dependencies
   - Test business logic in isolation

3. **Leverage Common Packages**
   - Use the common packages instead of direct stdlib calls
   - This improves testability and consistency

4. **Custom Implementations**
   - Create custom implementations for specific needs
   - Embed default implementation to inherit standard behavior

5. **Error Handling**
   - Use structured errors with proper codes
   - Include context information for debugging

## Future Enhancements

1. **Plugin System**
   - Dynamic loading of namespace implementations
   - Third-party namespace plugins

2. **Remote Execution**
   - Implementations that execute on remote systems
   - Cloud-native build implementations

3. **Caching Layer**
   - Built-in caching for expensive operations
   - Distributed cache support

4. **Metrics and Observability**
   - Built-in metrics collection
   - OpenTelemetry integration

5. **Advanced Error Recovery**
   - Automatic retry with backoff
   - Circuit breaker patterns
