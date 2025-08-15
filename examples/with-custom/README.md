# Custom Commands Example

This example shows how to combine custom project-specific commands with MAGE-X's built-in commands.

## 🎯 The Best of Both Worlds

With `magex`, you get:
1. **All 90+ built-in MAGE-X commands** - No boilerplate needed
2. **Your custom commands** - For project-specific needs
3. **Perfect integration** - They work together seamlessly

## 📁 Project Structure

```
your-project/
├── magefile.go    # Your custom commands (optional)
├── main.go        # Your application
└── go.mod         # Go modules
```

## 🚀 How It Works

1. **Built-in commands are always available:**
```bash
magex build         # MAGE-X built-in
magex test          # MAGE-X built-in
magex lint:fix      # MAGE-X built-in
magex release:multi # MAGE-X built-in
```

2. **Your custom commands from magefile.go:**
```bash
magex deploy        # Your custom command
magex stage         # Your custom command
magex rollback      # Your custom command
magex pipeline:ci   # Your custom namespace command
```

3. **See all commands (built-in + custom):**
```bash
magex -l
```

## 📝 Writing Custom Commands

Your `magefile.go` only needs to contain YOUR custom logic:

```go
//go:build mage
package main

// Deploy - your custom deployment
func Deploy() error {
    // Your deployment logic
    return nil
}

// Custom namespace for organized commands
type Pipeline mg.Namespace

func (Pipeline) CI() error {
    // Your CI logic
    return nil
}
```

That's it! No need to wrap MAGE-X commands - they're already available!

## 🔄 Migration Path

### From Standard Mage

**Before (Standard Mage):**
```go
// 200+ lines of boilerplate
func Build() error {
    return sh.Run("go", "build", "./...")
}
func Test() error {
    return sh.Run("go", "test", "./...")
}
// ... many more wrapper functions
```

**After (MAGE-X with magex):**
```go
// Just your custom commands!
func Deploy() error {
    // Your deployment logic
}
// All standard commands available via magex
```

### From MAGE-X Library Mode

**Before (MAGE-X Library):**
```go
// Still needed wrappers for each command
func Build() error { var b mage.Build; return b.Default() }
func Test() error { var t mage.Test; return t.Default() }
// ... 90+ wrapper functions
```

**After (MAGE-X Binary):**
```go
// No wrappers needed! Just custom commands
func Deploy() error {
    // Your deployment logic
}
```

## 🎨 Advanced: Calling MAGE-X from Custom Commands

You can compose MAGE-X commands in your custom logic:

```go
func Deploy() error {
    // First, ensure tests pass
    if err := exec.Command("magex", "test:ci").Run(); err != nil {
        return fmt.Errorf("tests failed: %w", err)
    }

    // Build for production
    if err := exec.Command("magex", "build:all").Run(); err != nil {
        return fmt.Errorf("build failed: %w", err)
    }

    // Your deployment logic
    fmt.Println("Deploying...")

    return nil
}
```

## 🌟 Benefits

1. **Minimal Code** - Only write what's unique to your project
2. **Maximum Power** - Access to all MAGE-X commands
3. **Easy Maintenance** - No boilerplate to maintain
4. **Team Friendly** - New developers can use standard commands immediately
5. **Incremental Adoption** - Start with built-ins, add custom as needed

## 📊 Comparison

| What You Write | Standard Mage | MAGE-X Library | MAGE-X Binary |
|---------------|---------------|----------------|---------------|
| Build command | ✍️ Write it | ✍️ Write wrapper | ✅ Built-in |
| Test command | ✍️ Write it | ✍️ Write wrapper | ✅ Built-in |
| Lint command | ✍️ Write it | ✍️ Write wrapper | ✅ Built-in |
| ... 87 more | ✍️ Write all | ✍️ Write wrappers | ✅ Built-in |
| Custom Deploy | ✍️ Write it | ✍️ Write it | ✍️ Write it |

## 🎯 Try It

```bash
# Install magex
go install github.com/mrz1836/mage-x/cmd/magex@latest

# See all commands (built-in + custom)
magex -l

# Run built-in commands
magex build
magex test:race
magex lint:fix

# Run custom commands
magex deploy
magex pipeline:ci

# Get help
magex -h
```

## 💡 Pro Tips

1. **Start with zero config** - Use built-in commands only
2. **Add custom gradually** - Create magefile.go when you need project-specific logic
3. **Keep it simple** - Only write what's unique to your project
4. **Compose commands** - Call magex from your custom commands for complex workflows

---

**MAGE-X + Your Code = Perfect Build Automation** 🚀
