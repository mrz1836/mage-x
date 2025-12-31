# 🗂️ MAGE-X Registry API Documentation

> **Registry System**: The heart of MAGE-X's hybrid binary+library architecture that enables command discovery, registration, and execution

## 🎯 Overview

The MAGE-X Registry API provides a sophisticated command management system that powers the `magex` binary and enables developers to create custom build tooling. It offers type-safe command registration, intelligent discovery, and flexible execution patterns.

### ✨ Core Features

- **🔧 Command Registration**: Type-safe command registration with validation
- **🔍 Discovery & Search**: Intelligent command discovery and fuzzy search
- **⚡ Execution Engine**: Unified command execution with dependency resolution
- **🎭 Flexible Architecture**: Support for both functional and method-based commands
- **🧠 Smart Suggestions**: Context-aware error messages and command suggestions
- **🔒 Thread Safety**: Concurrent access with mutex protection

## 📦 Installation

```bash
go get github.com/mrz1836/mage-x/pkg/mage/registry
```

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/mrz1836/mage-x/pkg/mage/registry"
)

func main() {
    // Create a new registry
    reg := registry.NewRegistry()

    // Register a simple command
    cmd, err := registry.NewCommand("hello").
        WithDescription("Say hello").
        WithFunc(func() error {
            fmt.Println("Hello, World!")
            return nil
        }).
        Build()

    if err != nil {
        panic(err)
    }

    reg.MustRegister(cmd)

    // Execute the command
    err = reg.Execute("hello")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Using the Global Registry

```go
package main

import (
    "fmt"
    "github.com/mrz1836/mage-x/pkg/mage/registry"
)

func main() {
    // Use the global registry (like magex does)
    registry.MustRegister(
        registry.NewCommand("build").
            WithDescription("Build the application").
            WithFunc(func() error {
                fmt.Println("Building...")
                return nil
            }).
            MustBuild(),
    )

    // Execute using global functions
    err := registry.Execute("build")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
    }
}
```

## 🏗️ Core Types

### Registry

The main registry type that manages command storage and execution:

```go
type Registry struct {
    // Thread-safe command storage
    mu       sync.RWMutex
    commands map[string]*Command
    aliases  map[string]string
    metadata CommandMetadata
}
```

#### Methods

```go
// Create new registry
func NewRegistry() *Registry

// Command management
func (r *Registry) Register(cmd *Command) error
func (r *Registry) MustRegister(cmd *Command)
func (r *Registry) Get(name string) (*Command, bool)
func (r *Registry) Clear()

// Discovery and listing
func (r *Registry) List() []*Command
func (r *Registry) ListByNamespace(namespace string) []*Command
func (r *Registry) ListByCategory(category string) []*Command
func (r *Registry) Namespaces() []string
func (r *Registry) Categories() []string
func (r *Registry) Search(query string) []*Command

// Execution
func (r *Registry) Execute(name string, args ...string) error

// Metadata
func (r *Registry) Metadata() CommandMetadata
```

### Command

Represents a single executable command:

```go
type Command struct {
    Name         string              // Command name
    Namespace    string              // Namespace (e.g., "build")
    Method       string              // Method name (e.g., "linux")
    Description  string              // Help description
    Aliases      []string            // Alternative names
    Func         CommandFunc         // Function without args
    FuncWithArgs CommandWithArgsFunc // Function with args
    Hidden       bool                // Hide from listings
    Deprecated   string              // Deprecation message
    Dependencies []string            // Required commands
    Category     string              // Grouping category
    Since        string              // Version introduced
}
```

#### Function Types

```go
type CommandFunc func() error
type CommandWithArgsFunc func(args ...string) error
```

#### Methods

```go
func (c *Command) FullName() string
func (c *Command) IsNamespace() bool
func (c *Command) Validate() error
func (c *Command) Execute(args ...string) error
```

## 🔧 Command Builder API

The fluent builder API for creating commands:

### Basic Builder

```go
// Create command with name
builder := registry.NewCommand("mycommand")

// Create namespace command
builder := registry.NewNamespaceCommand("build", "linux")
```

### Builder Methods

```go
func (b *CommandBuilder) WithDescription(desc string) *CommandBuilder
func (b *CommandBuilder) WithFunc(fn CommandFunc) *CommandBuilder
func (b *CommandBuilder) WithArgsFunc(fn CommandWithArgsFunc) *CommandBuilder
func (b *CommandBuilder) WithAliases(aliases ...string) *CommandBuilder
func (b *CommandBuilder) WithCategory(category string) *CommandBuilder
func (b *CommandBuilder) WithDependencies(deps ...string) *CommandBuilder
func (b *CommandBuilder) Hidden() *CommandBuilder
func (b *CommandBuilder) Deprecated(message string) *CommandBuilder
func (b *CommandBuilder) Since(version string) *CommandBuilder
func (b *CommandBuilder) Build() (*Command, error)
func (b *CommandBuilder) MustBuild() *Command
```

### Builder Examples

```go
// Simple command
cmd := registry.NewCommand("test").
    WithDescription("Run tests").
    WithFunc(func() error {
        return runTests()
    }).
    WithCategory("Testing").
    MustBuild()

// Complex namespace command
cmd := registry.NewNamespaceCommand("build", "multiplatform").
    WithDescription("Build for multiple platforms").
    WithArgsFunc(func(args ...string) error {
        platforms := args
        if len(platforms) == 0 {
            platforms = []string{"linux", "darwin", "windows"}
        }
        return buildForPlatforms(platforms)
    }).
    WithAliases("multi", "all").
    WithCategory("Build").
    WithDependencies("deps:check").
    Since("v1.0.0").
    MustBuild()

// Hidden/deprecated command
cmd := registry.NewCommand("oldcommand").
    WithDescription("Old command (use newcommand instead)").
    WithFunc(legacyFunction).
    Hidden().
    Deprecated("Use 'newcommand' instead").
    MustBuild()
```

## 🔍 Discovery & Search

### Listing Commands

```go
reg := registry.NewRegistry()
// ... register commands ...

// List all visible commands
commands := reg.List()
for _, cmd := range commands {
    fmt.Printf("%s - %s\n", cmd.FullName(), cmd.Description)
}

// List by namespace
buildCommands := reg.ListByNamespace("build")
for _, cmd := range buildCommands {
    fmt.Printf("  %s - %s\n", cmd.Method, cmd.Description)
}

// List by category
testCommands := reg.ListByCategory("Testing")
```

### Search Functionality

```go
// Search commands
results := reg.Search("build")
for _, cmd := range results {
    fmt.Printf("Found: %s\n", cmd.FullName())
}

// Search is case-insensitive and matches:
// - Command names
// - Namespace names
// - Method names
// - Descriptions
results = reg.Search("multi platform")  // Finds "build:multiplatform"
```

### Metadata Access

```go
meta := reg.Metadata()
fmt.Printf("Total commands: %d\n", meta.TotalCommands)
fmt.Printf("Namespaces: %v\n", meta.Namespaces)
fmt.Printf("Categories: %v\n", meta.Categories)
```

## ⚡ Command Execution

### Basic Execution

```go
// Execute command without arguments
err := reg.Execute("build")

// Execute command with arguments
err := reg.Execute("test", "pkg1", "pkg2")

// Execute namespace command
err := reg.Execute("build:linux")
```

### Execution Flow

1. **Lookup**: Find command by name or alias
2. **Dependencies**: Execute required dependencies first
3. **Validation**: Ensure command is valid
4. **Execution**: Call appropriate function
5. **Error Handling**: Return execution results

### Error Handling

```go
err := reg.Execute("nonexistent")
if err != nil {
    // Registry provides helpful error messages
    fmt.Printf("Error: %v\n", err)
    // Output: "unknown command 'nonexistent'. Did you mean: 'build'?"
}
```

### Command with Dependencies

```go
// Register commands with dependencies
registry.MustRegister(
    registry.NewCommand("test").
        WithDescription("Run tests").
        WithFunc(runTests).
        WithDependencies("build").  // Runs build first
        MustBuild(),
)

registry.MustRegister(
    registry.NewCommand("build").
        WithDescription("Build application").
        WithFunc(buildApp).
        MustBuild(),
)

// When you run test, it automatically runs build first
reg.Execute("test")  // Executes: build -> test
```

## 🎭 Advanced Patterns

### Namespace-Based Organization

```go
// Register commands in build namespace
registry.MustRegister(
    registry.NewNamespaceCommand("build", "linux").
        WithDescription("Build for Linux").
        WithFunc(buildLinux).
        MustBuild(),
)

registry.MustRegister(
    registry.NewNamespaceCommand("build", "windows").
        WithDescription("Build for Windows").
        WithFunc(buildWindows).
        MustBuild(),
)

// Access via namespace:method syntax
reg.Execute("build:linux")
reg.Execute("build:windows")

// List namespace commands
buildCmds := reg.ListByNamespace("build")
```

### Category-Based Grouping

```go
// Organize commands by category
registry.MustRegister(
    registry.NewCommand("unit").
        WithDescription("Run unit tests").
        WithFunc(runUnitTests).
        WithCategory("Testing").
        MustBuild(),
)

registry.MustRegister(
    registry.NewCommand("integration").
        WithDescription("Run integration tests").
        WithFunc(runIntegrationTests).
        WithCategory("Testing").
        MustBuild(),
)

// List by category
testingCmds := reg.ListByCategory("Testing")
```

### Aliases and Shortcuts

```go
// Create commands with multiple aliases
registry.MustRegister(
    registry.NewCommand("build").
        WithDescription("Build application").
        WithFunc(buildApp).
        WithAliases("b", "compile", "make").
        MustBuild(),
)

// All these work the same:
reg.Execute("build")
reg.Execute("b")
reg.Execute("compile")
reg.Execute("make")
```

### Command Arguments

```go
// Command that accepts arguments
registry.MustRegister(
    registry.NewCommand("test").
        WithDescription("Run specific tests").
        WithArgsFunc(func(args ...string) error {
            if len(args) == 0 {
                return runAllTests()
            }
            return runSpecificTests(args)
        }).
        MustBuild(),
)

// Usage:
reg.Execute("test")                    // Run all tests
reg.Execute("test", "pkg1", "pkg2")    // Run specific tests
```

### Hybrid Function Support

```go
// Command with both function types
registry.MustRegister(
    registry.NewCommand("serve").
        WithDescription("Start development server").
        WithFunc(func() error {
            return startServer(":8080")  // Default
        }).
        WithArgsFunc(func(args ...string) error {
            port := ":8080"
            if len(args) > 0 {
                port = ":" + args[0]
            }
            return startServer(port)
        }).
        MustBuild(),
)

// Usage:
reg.Execute("serve")        // Uses default port
reg.Execute("serve", "3000") // Uses custom port
```

## 🔧 Dynamic Loading

The registry supports dynamic command loading from user magefiles:

### Loader API

```go
type Loader struct {
    registry *Registry
    verbose  bool
}

func NewLoader(registry *Registry) *Loader
func (l *Loader) LoadUserMagefile(dir string) error
```

### Usage Example

```go
reg := registry.NewRegistry()
loader := registry.NewLoader(reg)

// Load commands from magefile.go in current directory
err := loader.LoadUserMagefile(".")
if err != nil {
    fmt.Printf("Failed to load magefile: %v\n", err)
}

// Now registry contains both built-in and user commands
commands := reg.List()
```

## 🛡️ Thread Safety

The registry is designed for concurrent access:

```go
reg := registry.NewRegistry()

// Safe to call from multiple goroutines
go reg.Execute("command1")
go reg.Execute("command2")
go reg.List()
go reg.Search("query")
```

All registry operations use read-write mutexes for thread safety.

## 🧪 Testing

### Testing Commands

```go
func TestMyCommand(t *testing.T) {
    reg := registry.NewRegistry()

    // Register test command
    called := false
    cmd := registry.NewCommand("test").
        WithDescription("Test command").
        WithFunc(func() error {
            called = true
            return nil
        }).
        MustBuild()

    reg.MustRegister(cmd)

    // Test execution
    err := reg.Execute("test")
    assert.NoError(t, err)
    assert.True(t, called)
}
```

### Mock Registry

```go
func TestWithMockRegistry(t *testing.T) {
    // Create isolated registry for testing
    reg := registry.NewRegistry()

    // Register mock commands
    reg.MustRegister(
        registry.NewCommand("mock").
            WithFunc(func() error {
                return nil
            }).
            MustBuild(),
    )

    // Test behavior
    commands := reg.List()
    assert.Len(t, commands, 1)
    assert.Equal(t, "mock", commands[0].Name)
}
```

## 📊 Performance

### Benchmarks

The registry is optimized for performance:

- **Registration**: ~1μs per command
- **Lookup**: ~100ns per command (O(1) hash map)
- **Search**: ~10μs for 100 commands (linear scan)
- **Execution**: ~50ns overhead per command

### Memory Usage

- **Base registry**: ~1KB
- **Per command**: ~200 bytes
- **215 commands**: ~45KB total

### Optimization Tips

```go
// Pre-allocate if you know command count
reg := registry.NewRegistry()
// ... register many commands ...

// Use Get() for repeated lookups (it's cached)
cmd, exists := reg.Get("frequently_used_command")
if exists {
    cmd.Execute()
}

// Batch register commands to reduce lock contention
commands := []*registry.Command{
    // ... many commands
}
for _, cmd := range commands {
    reg.Register(cmd)  // Each call locks/unlocks
}
```

## 🔮 Advanced Usage

### Custom Execution Logic

```go
// Implement custom execution wrapper
func ExecuteWithLogging(reg *registry.Registry, name string, args ...string) error {
    start := time.Now()
    fmt.Printf("Executing: %s %v\n", name, args)

    err := reg.Execute(name, args...)

    duration := time.Since(start)
    if err != nil {
        fmt.Printf("Failed in %v: %v\n", duration, err)
    } else {
        fmt.Printf("Completed in %v\n", duration)
    }

    return err
}
```

### Registry Middleware

```go
// Wrap registry with middleware
type MiddlewareRegistry struct {
    *registry.Registry
    middleware []func(string, []string) error
}

func (m *MiddlewareRegistry) Execute(name string, args ...string) error {
    // Run middleware
    for _, mw := range m.middleware {
        if err := mw(name, args); err != nil {
            return err
        }
    }

    // Execute original command
    return m.Registry.Execute(name, args...)
}
```

### Command Validation

```go
// Add custom validation
func ValidateCommand(cmd *registry.Command) error {
    if err := cmd.Validate(); err != nil {
        return err
    }

    // Custom validation rules
    if cmd.Namespace == "deploy" && cmd.Dependencies == nil {
        return fmt.Errorf("deploy commands must have dependencies")
    }

    return nil
}

// Use in registration
cmd, err := registry.NewCommand("deploy:prod").
    WithFunc(deployProduction).
    Build()

if err := ValidateCommand(cmd); err != nil {
    return err
}

reg.MustRegister(cmd)
```

## 📚 Best Practices

### Command Design
- Use **clear, descriptive names**
- Provide **meaningful descriptions**
- Follow **consistent naming** conventions
- Group related commands in **namespaces**

### Error Handling
- Return **specific error messages**
- Use **error wrapping** for context
- Provide **actionable suggestions**
- **Validate inputs** early

### Performance
- **Avoid expensive operations** in command registration
- **Cache results** when possible
- Use **appropriate data structures**
- **Profile** custom commands

### Testing
- **Test command registration**
- **Test command execution**
- **Test error conditions**
- Use **isolated registries** for tests

## 🔗 Integration Examples

### With magex Binary

```go
// How magex uses the registry
func main() {
    reg := registry.Global()

    // Register all built-in commands
    embed.RegisterAll(reg)

    // Load user magefile if present
    loader := registry.NewLoader(reg)
    loader.LoadUserMagefile(".")

    // Execute command from CLI args
    if len(os.Args) > 1 {
        err := reg.Execute(os.Args[1], os.Args[2:]...)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error: %v\n", err)
            os.Exit(1)
        }
    }
}
```

### Custom Build Tool

```go
// Create your own build tool using the registry
func main() {
    reg := registry.NewRegistry()

    // Register your custom commands
    registerProjectCommands(reg)

    // Add CLI interface
    if len(os.Args) < 2 {
        listCommands(reg)
        return
    }

    command := os.Args[1]
    args := os.Args[2:]

    err := reg.Execute(command, args...)
    if err != nil {
        log.Fatal(err)
    }
}
```

---

**Ready to build your own automation tools with the MAGE-X Registry?**

```bash
go get github.com/mrz1836/mage-x/pkg/mage/registry
```

The Registry API provides the foundation for creating powerful, flexible, and user-friendly build tooling. Whether you're extending `magex` or building your own custom solution, the registry system gives you the building blocks you need.
