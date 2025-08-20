# AGENTS.md

## 🎯 Purpose & Scope

This file defines the **baseline standards, workflows, and architecture** for *all contributors and AI agents* operating within the MAGE-X repository. It serves as the definitive authority for engineering conduct, coding conventions, security practices, and collaborative norms.

MAGE-X is a comprehensive build automation toolkit that follows the philosophy of "Write Once, Mage Everywhere." This document will help AI assistants (Claude, Cursor, ChatGPT, Sweep AI) and human developers understand our architecture, contribute clean and secure code, and navigate the complex multi-phase implementation effectively.

> Whether reading, writing, testing, or extending MAGE-X, **you must adhere to the principles and patterns in this document.**

Additional documentation files complement this guide:
- `CODE_STANDARDS.md` — Detailed coding style and conventions
- `CONTRIBUTING.md` — Contribution workflow and guidelines
- `SECURITY.md` — Security policies and vulnerability reporting
- `README.md` — Project overview and quick start guide

<br/>

---

<br/>

## 🎭 Project Overview: MAGE-X

**MAGE-X** is an enterprise-grade build automation toolkit for Go projects that transforms how developers manage build, test, and release workflows. Built on the philosophy of "Write Once, Mage Everywhere," it provides a comprehensive solution for managing 30+ repositories with consistent tooling and zero-configuration ease.

### Core Philosophy
- **Zero-Configuration Excellence**: Works out of the box with intelligent defaults
- **Security-First Architecture**: Every command execution is validated and sandboxed
- **Interactive Experience**: Friendly CLI with guided wizards and contextual help
- **Multi-Repository Management**: Designed for enterprise-scale repository management
- **AI Agent Ready**: Machine-readable guidelines for modern development workflows

### Key Capabilities
- **Comprehensive Build System**: All essential build, test, lint, and release tasks
- **Multi-Channel Releases**: Stable, beta, and edge release channels with automation
- **Interactive Mode**: Guided wizards and friendly CLI for complex operations
- **Recipe System**: Pre-built patterns for common development scenarios
- **Security Validation**: Command injection prevention and input sanitization
- **Enterprise Features**: Audit logging, compliance reporting, team management

<br/>

---

<br/>

## 🏗️ Architecture Overview

MAGE-X follows a layered architecture designed for security, extensibility, and maintainability:

### Layer 1: Security Foundation (`pkg/security/`)
- **CommandExecutor Interface**: Secure command execution with validation
- **Input Validation**: Prevents shell injection and path traversal attacks
- **Environment Filtering**: Automatic removal of sensitive environment variables
- **Timeout Management**: Prevents runaway processes and resource exhaustion

### Layer 2: Core Infrastructure (`pkg/utils/`)
- **Native Logging**: Colored output with progress indicators and spinners
- **Platform Abstraction**: Cross-platform compatibility utilities
- **Configuration Management**: YAML parsing with environment overrides
- **Error Handling**: Consistent error formatting and reporting

### Layer 3: Task Implementation (`pkg/mage/`)
- **Namespace Organization**: Logical grouping of related tasks
- **Build System**: Cross-platform builds with optimization
- **Test Framework**: Comprehensive testing with coverage and race detection
- **Quality Assurance**: Linting, formatting, and security scanning
- **Release Management**: Multi-channel releases with automation

### Layer 4: User Experience
- **Interactive Mode**: Guided CLI experience with wizards
- **Recipe System**: Common patterns and best practices
- **Help System**: Comprehensive documentation and usage examples
- **Configuration**: Flexible YAML-based configuration with smart defaults

<br/>

---

<br/>

## 📁 Directory Structure

| Directory               | Description                                                                  |
|-------------------------|------------------------------------------------------------------------------|
| `.github/`              | GitHub workflows, issue templates, and AI agent documentation               |
| `pkg/mage/`             | Core mage task implementations (25+ task files)                            |
| `pkg/security/`         | Security-first command execution and validation                             |
| `pkg/utils/`            | Utility functions, logging, and platform abstractions                      |
| `examples/`             | Example projects demonstrating various usage patterns                       |
| `templates/`            | Project templates for different Go project types                            |
| `project-plan.md`       | Complete implementation roadmap and feature specifications                   |

<br/>

---

<br/>

## 🎯 Go Essentials for MAGE-X

These are the non-negotiable practices that define professional Go development in the MAGE-X ecosystem. Every function, package, and design decision should reflect these principles.

### 🌐 Context-First Design

Context flows through the entire MAGE-X architecture for proper cancellation and timeout handling.

* **Always pass `context.Context` as the first parameter** for operations that can be canceled or have timeouts
* **Never store context in structs**—pass it explicitly through function calls
* **Use `context.Background()` only at the top level** (main functions, test setup, service initialization)
* **Derive child contexts** using `context.WithTimeout()`, `context.WithCancel()`, or `context.WithValue()`
* **Respect context cancellation** by checking `ctx.Done()` in long-running operations

```go
// ✅ Correct: Context as first parameter in MAGE-X patterns
func (e *SecureExecutor) Execute(ctx context.Context, name string, args ...string) error {
    // Create command with timeout context
    ctx, cancel := e.contextWithTimeout(ctx)
    defer cancel()

    // Validate before execution
    if err := e.validateCommand(name, args); err != nil {
        return fmt.Errorf("command validation failed: %w", err)
    }

    // Execute with context
    cmd := exec.CommandContext(ctx, name, args...)
    return cmd.Run()
}
```

### 🔒 Security-First Development

Security is paramount in MAGE-X. Every command execution must be validated and sandboxed.

* **Always use the `CommandExecutor` interface** for any external command execution
* **Never use `exec.Command` directly** - use `security.NewSecureExecutor()`
* **Validate all inputs** using `security.ValidateCommandArg()` and `security.ValidatePath()`
* **Filter sensitive environment variables** automatically
* **Implement dry-run mode** for testing command execution without side effects

```go
// ✅ Correct: Security-first command execution
func (b Build) Default() error {
    ctx := context.Background()

    // Use secure executor from common.go
    if err := executor.Execute(ctx, "go", "build", "-o", cfg.Project.Binary); err != nil {
        return fmt.Errorf("build failed: %w", err)
    }

    utils.Success("Build completed successfully")
    return nil
}

// ❌ Incorrect: Direct command execution
func (b Build) Default() error {
    cmd := exec.Command("go", "build") // Security vulnerability!
    return cmd.Run()
}
```

### 🎨 Consistent User Experience

MAGE-X provides a unified, friendly user experience across all operations.

* **Use utils.Header()** for section titles and major operations
* **Use utils.Success()** for successful completion messages
* **Use utils.Error()** for error messages with actionable context
* **Use utils.Info()** for informational messages during operations
* **Use utils.Warning()** for non-fatal issues that need attention

```go
// ✅ Correct: Consistent messaging patterns
func (t Test) Cover() error {
    utils.Header("🧪 Running Tests with Coverage")

    ctx := context.Background()
    utils.Info("Generating coverage report...")

    if err := executor.Execute(ctx, "go", "test", "-cover", "./..."); err != nil {
        return utils.Error("Tests failed: %v", err)
    }

    utils.Success("Coverage report generated successfully")
    return nil
}
```

### 🏗️ Interface-Based Architecture

MAGE-X uses interfaces for testability, flexibility, and security.

* **Define interfaces for all external dependencies** (command execution, file operations)
* **Use dependency injection** to provide implementations
* **Create mock implementations** for testing
* **Keep interfaces small and focused** on single responsibilities

```go
// ✅ Correct: Interface-based design
type CommandExecutor interface {
    Execute(ctx context.Context, name string, args ...string) error
    ExecuteOutput(ctx context.Context, name string, args ...string) (string, error)
    ExecuteWithEnv(ctx context.Context, env []string, name string, args ...string) error
}

// Implementation uses interface
var executor CommandExecutor = security.NewSecureExecutor()
```

### 📊 Configuration Management

MAGE-X supports flexible configuration with sensible defaults.

* **Support both YAML configuration and environment variables**
* **Environment variables override YAML settings**
* **Provide comprehensive defaults** for zero-configuration usage
* **Validate configuration** before use
* **Use typed configuration structs** with YAML tags

```go
// ✅ Correct: Configuration pattern
type BuildConfig struct {
    Tags      []string `yaml:"tags"`
    Platforms []string `yaml:"platforms"`
    Output    string   `yaml:"output"`
    Parallel  int      `yaml:"parallel"`
    Verbose   bool     `yaml:"verbose"`
}

// Apply environment overrides
func applyEnvOverrides(cfg *Config) {
    if tags := os.Getenv("MAGE_X_BUILD_TAGS"); tags != "" {
        cfg.Build.Tags = strings.Split(tags, ",")
    }
    if verbose := os.Getenv("MAGE_X_VERBOSE"); verbose != "" {
        cfg.Build.Verbose = verbose == "true"
    }
}
```

<br/>

---

<br/>

## 🔐 Security Architecture

MAGE-X implements a comprehensive security model to prevent common vulnerabilities in build automation tools.

### Command Execution Security

All external commands go through the `CommandExecutor` interface with multiple layers of validation:

1. **Command Whitelist**: Optional allowlist of permitted commands
2. **Argument Validation**: Detection of shell injection patterns
3. **Path Validation**: Prevention of directory traversal attacks
4. **Environment Filtering**: Automatic removal of sensitive variables
5. **Timeout Management**: Prevents resource exhaustion

```go
// Security validation patterns
func (e *SecureExecutor) validateCommand(name string, args []string) error {
    // Check command whitelist
    if len(e.AllowedCommands) > 0 && !e.AllowedCommands[name] {
        return fmt.Errorf("command '%s' is not in allowed list", name)
    }

    // Validate arguments for injection attacks
    for _, arg := range args {
        if err := ValidateCommandArg(arg); err != nil {
            return fmt.Errorf("invalid argument '%s': %w", arg, err)
        }
    }

    return nil
}
```

### Input Validation

Critical for preventing injection attacks:

```go
// Dangerous patterns detected by ValidateCommandArg
var dangerousPatterns = []string{
    "$(",      // Command substitution
    "`",       // Command substitution
    "&&",      // Command chaining
    "||",      // Command chaining
    ";",       // Command separator
    "|",       // Pipe operations
    ">", "<",  // Redirections
}
```

### Environment Security

Automatic filtering of sensitive environment variables:

```go
// Sensitive prefixes automatically filtered
var sensitivePrefix = []string{
    "AWS_SECRET", "GITHUB_TOKEN", "GITLAB_TOKEN",
    "NPM_TOKEN", "DOCKER_PASSWORD", "DATABASE_PASSWORD",
    "API_KEY", "SECRET", "PRIVATE_KEY",
}
```

<br/>

---

<br/>

## 🚀 Implementation Phases

MAGE-X follows a structured development approach with clear phases:

### Phase 1: Core Excellence ✅ **COMPLETE**
- **Security Infrastructure**: CommandExecutor interface with validation
- **Native Logging**: Colored output with progress indicators
- **Core Build System**: All essential build, test, lint, and release tasks
- **Version Management**: Automatic version detection and update infrastructure

### Phase 2: Zero-to-Hero Experience ✅ **COMPLETE**
- **Project Templates**: CLI, library, web API, and microservice templates
- **Multi-Channel Releases**: Stable, beta, and edge release channels
- **Configuration Management**: Flexible mage.yaml with smart defaults
- **Asset Distribution**: Automated building and distribution

### Phase 3: Interactive Experience ✅ **COMPLETE**
- **Interactive Mode**: Friendly CLI with guided operations
- **Interactive Wizard**: Step-by-step setup for complex operations
- **Help System**: Comprehensive help and usage examples
- **Recipe System**: Common patterns and best practices library

### Phase 4: Enterprise Features 🔄 **PENDING**
- **Audit Logging**: Comprehensive activity tracking and compliance reporting
- **Security Scanning**: Vulnerability detection and security policy enforcement
- **Team Management**: Role-based access and team collaboration features
- **Analytics**: Build metrics, performance tracking, and optimization insights

<br/>

---

<br/>

## 🎨 User Experience Patterns

MAGE-X provides a consistent, friendly user experience across all operations.

### Interactive Mode

The interactive mode provides a guided experience for complex operations:

```go
// Interactive session structure
type InteractiveSession struct {
    reader   *bufio.Reader
    history  []string
    commands map[string]InteractiveCommand
    running  bool
}

// Command structure for interactive mode
type InteractiveCommand struct {
    Name        string
    Description string
    Usage       string
    Handler     func(args []string) error
    Category    string
}
```

### Recipe System

Recipes provide reusable patterns for common development scenarios:

```go
// Recipe structure
type Recipe struct {
    Name         string
    Description  string
    Category     string
    Dependencies []string
    Steps        []RecipeStep
    Templates    map[string]string
    Variables    map[string]string
}

// Available recipe categories
var recipeCategories = []string{
    "project-setup",    // Project initialization patterns
    "ci-cd",           // CI/CD pipeline configurations
    "security",        // Security hardening patterns
    "performance",     // Performance optimization
    "documentation",   // Documentation generation
}
```

### Progress Indicators

Native progress tracking for long-running operations:

```go
// Spinner for indefinite operations
spinner := utils.NewSpinner("Building project...")
spinner.Start()
defer spinner.Stop()

// Progress bars for definite operations
progress := utils.NewProgressBar("Compiling packages", totalPackages)
for i, pkg := range packages {
    // Compile package
    progress.Update(i + 1)
}
```

<br/>

---

<br/>

## 🤖 AI Agent Guidelines

Specific guidelines for different AI assistants working with MAGE-X:

### For Claude AI

When working with MAGE-X:

1. **Architecture Awareness**: Understand the layered architecture (Security → Utils → Mage → UX)
2. **Security First**: Always use the CommandExecutor interface, never direct exec.Command
3. **Context Propagation**: Pass context through all function calls
4. **Error Handling**: Wrap errors with context using fmt.Errorf
5. **User Experience**: Use utils functions for consistent messaging
6. **Testing**: Create comprehensive tests with both unit and integration coverage

```go
// Claude-preferred pattern for new mage tasks
func (namespace Namespace) NewTask() error {
    utils.Header("🎯 Starting New Task")

    ctx := context.Background()

    // Use secure executor
    if err := executor.Execute(ctx, "command", "args"); err != nil {
        return fmt.Errorf("task failed: %w", err)
    }

    utils.Success("Task completed successfully")
    return nil
}
```

### For Cursor IDE

When using Cursor with MAGE-X:

1. **Code Completion**: Leverage the comprehensive type system and interfaces
2. **Refactoring**: Use the consistent patterns across all mage tasks
3. **Navigation**: Follow the namespace organization in pkg/mage/
4. **Testing**: Use the existing test patterns and mock implementations

### For ChatGPT

When generating MAGE-X code:

1. **Follow Patterns**: Use existing code patterns from similar tasks
2. **Security Validation**: Always include input validation and security checks
3. **Configuration**: Support both YAML and environment variable configuration
4. **Documentation**: Include comprehensive godoc comments
5. **Error Messages**: Provide actionable error messages with context

### For Sweep AI

When reviewing MAGE-X PRs:

1. **Security Review**: Check for proper use of CommandExecutor interface
2. **Pattern Consistency**: Ensure new code follows established patterns
3. **Test Coverage**: Verify comprehensive test coverage including edge cases
4. **Documentation**: Ensure all public functions have proper godoc comments
5. **Performance**: Check for proper resource management and cleanup

<br/>

---

<br/>

## 📋 Common Development Patterns

### Adding New Mage Tasks

1. **Create Task File**: Follow naming pattern (e.g., `pkg/mage/newtask.go`)
2. **Define Namespace**: Use mage namespace pattern
3. **Implement Interface**: Use CommandExecutor for any external commands
4. **Add Configuration**: Extend Config struct if needed
5. **Write Tests**: Comprehensive unit and integration tests
6. **Update Documentation**: Add to help system and examples

```go
// Template for new mage task
package mage

import (
    "context"
    "github.com/magefile/mage/mg"
    "github.com/mrz1836/mage-x/pkg/utils"
)

// NewTask namespace for new task operations
type NewTask mg.Namespace

// Default performs the main task operation
func (NewTask) Default() error {
    utils.Header("🎯 New Task Operation")

    ctx := context.Background()

    // Use secure executor
    if err := executor.Execute(ctx, "command", "args"); err != nil {
        return fmt.Errorf("new task failed: %w", err)
    }

    utils.Success("New task completed successfully")
    return nil
}
```

### Configuration Extension

1. **Add Config Struct**: Define configuration structure
2. **Add to Main Config**: Include in main Config struct
3. **Add Defaults**: Provide sensible defaults
4. **Add Environment Overrides**: Support environment variable overrides
5. **Add Validation**: Validate configuration values

```go
// Configuration pattern
type NewTaskConfig struct {
    Enabled   bool     `yaml:"enabled"`
    Timeout   string   `yaml:"timeout"`
    Options   []string `yaml:"options"`
    Parallel  int      `yaml:"parallel"`
}

// Add to main config
type Config struct {
    // ... existing fields
    NewTask NewTaskConfig `yaml:"new_task"`
}

// Provide defaults
func defaultNewTaskConfig() NewTaskConfig {
    return NewTaskConfig{
        Enabled:  true,
        Timeout:  "5m",
        Options:  []string{},
        Parallel: runtime.NumCPU(),
    }
}
```

### Testing Patterns

1. **Unit Tests**: Test individual functions with mocked dependencies
2. **Integration Tests**: Test complete workflows with real dependencies
3. **Security Tests**: Test input validation and security features
4. **Performance Tests**: Test resource usage and performance characteristics

```go
// Test pattern for mage tasks
func TestNewTask_Default(t *testing.T) {
    // Setup mock executor
    mockExecutor := security.NewMockExecutor()
    mockExecutor.SetResponse("command args", "", nil)

    // Replace global executor for testing
    originalExecutor := executor
    executor = mockExecutor
    defer func() { executor = originalExecutor }()

    // Test the task
    task := NewTask{}
    err := task.Default()

    // Verify results
    assert.NoError(t, err)
    assert.Len(t, mockExecutor.ExecuteCalls, 1)
    assert.Equal(t, "command", mockExecutor.ExecuteCalls[0].Name)
}
```

<br/>

---

<br/>

## 🔍 Code Review Checklist

When reviewing MAGE-X code changes:

### Security Review
- [ ] Uses CommandExecutor interface for all external commands
- [ ] Includes proper input validation for all user inputs
- [ ] Filters sensitive environment variables
- [ ] Implements proper timeout handling
- [ ] Validates file paths for traversal attacks

### Architecture Review
- [ ] Follows layered architecture (Security → Utils → Mage → UX)
- [ ] Uses proper error handling with context
- [ ] Implements consistent user experience patterns
- [ ] Uses appropriate logging and messaging
- [ ] Follows interface-based design

### Testing Review
- [ ] Includes comprehensive unit tests
- [ ] Includes integration tests where appropriate
- [ ] Uses mock implementations for external dependencies
- [ ] Tests error conditions and edge cases
- [ ] Includes security-focused tests

### Documentation Review
- [ ] All public functions have godoc comments
- [ ] Complex logic is properly commented
- [ ] Examples are provided for new features
- [ ] Configuration options are documented
- [ ] Updates help system and interactive mode

<br/>

---

<br/>

## 🎯 Multi-Repository Management

MAGE-X is designed to manage 30+ repositories with consistent tooling:

### Repository Patterns
- **Consistent Magefile**: Same import pattern across all repositories
- **Shared Configuration**: Common mage.yaml patterns with local overrides
- **Centralized Updates**: Update MAGE-X once, propagate to all repositories
- **Automated Testing**: Consistent testing patterns across all projects

### Update Distribution
- **Semantic Versioning**: Proper version management with compatibility guarantees
- **Channel Management**: Stable, beta, and edge channels for different needs
- **Automated Notifications**: Optional notifications for available updates
- **Migration Support**: Automated migration tools for breaking changes

### Best Practices
- **Configuration Management**: Consistent patterns across repositories
- **Security Policies**: Centralized security configuration
- **Team Workflows**: Standardized development and release processes
- **Monitoring**: Centralized metrics and performance tracking

<br/>

---

<br/>

## 🚨 Troubleshooting Guide

### Common Issues

**1. Command Execution Failures**
- Check command is in PATH or use full path
- Verify command arguments are properly escaped
- Check for environment variable conflicts
- Enable dry-run mode for debugging

**2. Configuration Problems**
- Validate YAML syntax in mage.yaml
- Check environment variable naming
- Verify configuration file permissions
- Use `mage yaml:validate` for validation

**3. Security Validation Errors**
- Review command arguments for dangerous patterns
- Check file paths for traversal attempts
- Verify environment variable filtering
- Use `mage security:check` for analysis

**4. Performance Issues**
- Check parallel execution settings
- Monitor resource usage during builds
- Optimize build caching strategies
- Use `mage metrics:show` for analysis

### Debugging Techniques

**1. Verbose Mode**
```bash
VERBOSE=true mage build
```

**2. Dry Run Mode**
```bash
magex version:bump dry-run  # For version management
```

**3. Debug Logging**
```bash
DEBUG=true mage build
```

**4. Configuration Validation**
```bash
mage yaml:validate
mage yaml:show
```

<br/>

---

<br/>

## 📚 Resources and References

### Core Documentation
- [MAGE-X README](../README.md) - Project overview and quick start
- [Project Plan](../project-plan.md) - Complete implementation roadmap
- [Contributing Guidelines](CONTRIBUTING.md) - Contribution workflow
- [Security Policy](SECURITY.md) - Security practices and reporting

### External Resources
- [Mage Documentation](https://magefile.org) - Official Mage documentation
- [Go Security Guidelines](https://go.dev/doc/security) - Go security best practices
- [Semantic Versioning](https://semver.org) - Version management standards

### Community
- [GitHub Issues](https://github.com/mrz1836/mage-x/issues) - Bug reports and feature requests
- [GitHub Discussions](https://github.com/mrz1836/mage-x/discussions) - Community discussion
- [Contributing Guide](CONTRIBUTING.md) - How to contribute to the project

<br/>

---

<br/>

## 🎉 Conclusion

MAGE-X represents a new approach to build automation that prioritizes security, user experience, and enterprise-scale management. By following the guidelines in this document, contributors can maintain the high standards of quality, security, and consistency that make MAGE-X a joy to use.

Whether you're an AI assistant generating code or a human developer extending functionality, remember the core principles:
- **Security First**: Every command execution must be validated
- **User Experience**: Consistent, friendly, and helpful interface
- **Architecture**: Layered, interface-based design for maintainability
- **Testing**: Comprehensive coverage including security and performance
- **Documentation**: Clear, actionable guidance for all users

Together, we're building the future of build automation. **Write Once, Mage Everywhere!** 🪄

---

*This document is version-controlled and should be updated whenever architectural changes or new patterns are introduced to the MAGE-X project.*
