# CLAUDE.md - Quick Reference for Claude AI

## 📋 Essential Context

This is **MAGE-X** - a comprehensive build automation toolkit following the philosophy "Write Once, Mage Everywhere." For complete understanding, **read [AGENTS.md](AGENTS.md) first** - it contains the full architecture, security model, and development guidelines.

## 🎯 Quick Start for Claude

### Project Understanding
- **Architecture**: Security-first, layered design (Security → Utils → Mage → UX)
- **Current Status**: Phase 3 complete (Interactive Experience), Phase 4 pending (Enterprise)
- **Key Principle**: Every command execution must use the `CommandExecutor` interface

### Core Patterns to Follow

#### 1. Security-First Command Execution
```go
// ✅ ALWAYS use this pattern
func (namespace Namespace) Task() error {
    ctx := context.Background()
    
    // Use secure executor from common.go
    if err := executor.Execute(ctx, "command", "args"); err != nil {
        return fmt.Errorf("operation failed: %w", err)
    }
    
    return nil
}

// ❌ NEVER do this - security vulnerability!
func (namespace Namespace) Task() error {
    cmd := exec.Command("command", "args") // Direct execution forbidden!
    return cmd.Run()
}
```

#### 2. Consistent User Experience
```go
// ✅ Use utils functions for all user interaction
func (namespace Namespace) Task() error {
    utils.Header("🎯 Task Description")
    utils.Info("Processing...")
    
    // ... do work ...
    
    utils.Success("Task completed successfully")
    return nil
}
```

#### 3. Context-First Design
```go
// ✅ Always pass context as first parameter
func ProcessData(ctx context.Context, data string) error {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
    }
    
    // Continue processing...
    return nil
}
```

## 🏗️ Architecture Layers

1. **Security Layer** (`pkg/security/`): Command validation, input sanitization
2. **Utils Layer** (`pkg/utils/`): Logging, platform abstraction, configuration
3. **Mage Layer** (`pkg/mage/`): 25+ task files with namespace organization
4. **UX Layer**: Interactive mode, wizards, help system

## 🔐 Security Requirements

- **Always validate inputs** with `security.ValidateCommandArg()`
- **Never use exec.Command directly** - use `executor` from common.go
- **Filter sensitive env vars** automatically (already implemented)
- **Implement timeout handling** via context
- **Support dry-run mode** for testing

## 🎨 New Task Template

```go
package mage

import (
    "context"
    "fmt"
    "github.com/magefile/mage/mg"
    "github.com/mrz1836/go-mage/pkg/utils"
)

// NewTask namespace for new operations
type NewTask mg.Namespace

// Default performs the main operation
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

// Sub-task example
func (NewTask) SubTask() error {
    utils.Header("🔧 Sub Task Operation")
    
    ctx := context.Background()
    
    // Configuration access
    cfg, err := LoadConfig()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    // Use configuration
    if cfg.NewTask.Enabled {
        if err := executor.Execute(ctx, "command", cfg.NewTask.Options...); err != nil {
            return fmt.Errorf("sub task failed: %w", err)
        }
    }
    
    utils.Success("Sub task completed successfully")
    return nil
}
```

## 🧪 Testing Pattern

```go
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

## 📊 Configuration Extension

```go
// 1. Define config struct
type NewTaskConfig struct {
    Enabled   bool     `yaml:"enabled"`
    Timeout   string   `yaml:"timeout"`
    Options   []string `yaml:"options"`
}

// 2. Add to main Config struct
type Config struct {
    // ... existing fields
    NewTask NewTaskConfig `yaml:"new_task"`
}

// 3. Provide defaults
func defaultNewTaskConfig() NewTaskConfig {
    return NewTaskConfig{
        Enabled: true,
        Timeout: "5m",
        Options: []string{},
    }
}

// 4. Add environment overrides in applyEnvOverrides()
if enabled := os.Getenv("NEW_TASK_ENABLED"); enabled != "" {
    cfg.NewTask.Enabled = enabled == "true"
}
```

## 🚨 Critical Don'ts

- **Never use `exec.Command` directly** - always use `executor`
- **Never ignore context cancellation** - check `ctx.Done()` in loops
- **Never hardcode paths** - use `filepath.Join()` and validate with `security.ValidatePath()`
- **Never log sensitive data** - environment filtering is automatic but be careful
- **Never skip input validation** - use `security.ValidateCommandArg()`

## 🎯 Common Scenarios

### Adding Build Step
```go
func (Build) NewStep() error {
    utils.Header("🔨 New Build Step")
    
    ctx := context.Background()
    cfg, _ := LoadConfig()
    
    args := []string{"build", "-o", cfg.Build.Output}
    if cfg.Build.Verbose {
        args = append(args, "-v")
    }
    
    if err := executor.Execute(ctx, "go", args...); err != nil {
        return fmt.Errorf("build step failed: %w", err)
    }
    
    utils.Success("Build step completed")
    return nil
}
```

### Adding Test Command
```go
func (Test) NewType() error {
    utils.Header("🧪 New Test Type")
    
    ctx := context.Background()
    cfg, _ := LoadConfig()
    
    args := []string{"test", "-run", "TestPattern"}
    if cfg.Test.Verbose {
        args = append(args, "-v")
    }
    
    if err := executor.Execute(ctx, "go", args...); err != nil {
        return fmt.Errorf("tests failed: %w", err)
    }
    
    utils.Success("Tests completed")
    return nil
}
```

## 🔗 Complete Documentation

For comprehensive guidance, architectural details, and advanced patterns:

→ **[Read AGENTS.md](AGENTS.md)** ← Complete documentation

This includes:
- Full architecture overview
- Security model details
- Implementation phase status
- AI agent specific guidelines
- Code review checklist
- Troubleshooting guide
- Multi-repository management patterns

## 🎉 Remember

MAGE-X is about **"Write Once, Mage Everywhere"** - focus on:
- **Security first** - validate everything
- **User experience** - friendly, helpful messages
- **Consistency** - follow established patterns
- **Testing** - comprehensive coverage
- **Documentation** - clear, actionable guidance

When in doubt, check existing code patterns in `pkg/mage/` or refer to the comprehensive [AGENTS.md](AGENTS.md) guide.
