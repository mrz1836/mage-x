# Contributing to MAGE-X

Thank you for your interest in contributing to MAGE-X! This guide will help you get started.

## Table of Contents

1. [Code of Conduct](#code-of-conduct)
2. [Getting Started](#getting-started)
3. [Development Setup](#development-setup)
4. [Making Changes](#making-changes)
5. [Testing](#testing)
6. [Submitting Changes](#submitting-changes)
7. [Style Guidelines](#style-guidelines)
8. [Adding New Features](#adding-new-features)

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct:

- Be respectful and inclusive
- Welcome newcomers and help them get started
- Focus on constructive criticism
- Accept feedback gracefully
- Prioritize the project's best interests

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/mrz1836/go-mage.git
   cd MAGE-X
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/originalowner/MAGE-X.git
   ```

## Development Setup

### Prerequisites

- Go 1.21 or later
- Mage (`go install github.com/magefile/mage@latest`)
- Git

### Initial Setup

```bash
# Install dependencies
go mod download

# Install development tools
mage tools:install

# Run tests to verify setup
mage test
```

## Making Changes

### 1. Create a Branch

```bash
# Update main branch
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feature/your-feature-name
```

### 2. Make Your Changes

- Write clear, concise code
- Follow the existing code style
- Add tests for new functionality
- Update documentation as needed

### 3. Test Your Changes

```bash
# Run all tests
mage test

# Run with race detector
mage test:race

# Run with coverage
mage test:cover

# Run linter
mage lint

# Fix linting issues
mage lint:fix
```

### 4. Commit Your Changes

Follow conventional commit format:

```bash
git add .
git commit -m "feat: add new awesome feature"
```

Commit types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `test`: Test additions or changes
- `chore`: Build process or auxiliary tool changes

## Testing

### Writing Tests

1. **Unit Tests**: Test individual functions
   ```go
   func TestMyFunction(t *testing.T) {
       result := MyFunction(input)
       assert.Equal(t, expected, result)
   }
   ```

2. **Table-Driven Tests**: Test multiple scenarios
   ```go
   tests := []struct {
       name     string
       input    string
       expected string
   }{
       {"basic test", "input", "expected"},
   }
   
   for _, tt := range tests {
       t.Run(tt.name, func(t *testing.T) {
           result := MyFunction(tt.input)
           assert.Equal(t, tt.expected, result)
       })
   }
   ```

3. **Integration Tests**: Test complete workflows
   ```go
   //go:build integration
   // +build integration
   
   func TestIntegration(t *testing.T) {
       // Integration test code
   }
   ```

### Coverage Requirements

- New code should have >80% test coverage
- Run `mage test:cover` to check coverage
- View detailed report with `mage test:coverhtml`

## Submitting Changes

### 1. Push to Your Fork

```bash
git push origin feature/your-feature-name
```

### 2. Create Pull Request

1. Go to GitHub and create a pull request
2. Fill out the PR template
3. Link any related issues
4. Ensure all CI checks pass

### 3. PR Guidelines

- **Title**: Clear and descriptive
- **Description**: Explain what and why
- **Testing**: Describe how you tested
- **Screenshots**: Include if relevant
- **Breaking Changes**: Clearly marked

### 4. Review Process

- Maintainers will review your PR
- Address any requested changes
- Once approved, it will be merged

## Style Guidelines

### Go Code Style

1. **Format**: Use `gofmt` (run `mage lint:fmt`)
2. **Imports**: Group standard library, external, and internal imports
3. **Comments**: Export functions must have comments
4. **Error Handling**: Always handle errors explicitly
5. **Names**: Use clear, descriptive names

### Documentation Style

1. **README**: Update when adding features
2. **Code Comments**: Explain why, not what
3. **Examples**: Provide usage examples
4. **Godoc**: Follow Go documentation conventions

## Adding New Features

### 1. New Task Module

Create a new file in `pkg/mage/`:

```go
package mage

import (
    "github.com/magefile/mage/mg"
    "github.com/mrz1836/go-mage/pkg/utils"
)

// NewFeature namespace for new feature tasks
type NewFeature mg.Namespace

// Task performs the new task
func (NewFeature) Task() error {
    utils.PrintHeader("Running New Task")
    // Implementation
    return nil
}
```

### 2. Add Configuration

Update `pkg/mage/config.go`:

```go
type Config struct {
    // ... existing fields
    NewFeature NewFeatureConfig `yaml:"new_feature"`
}

type NewFeatureConfig struct {
    Enabled bool   `yaml:"enabled"`
    Options string `yaml:"options"`
}
```

### 3. Add Tests

Create `pkg/mage/newfeature_test.go`:

```go
func TestNewFeature(t *testing.T) {
    // Test implementation
}
```

### 4. Update Documentation

- Add to README.md
- Update examples
- Add to AGENTS.md if relevant

## Common Tasks

### Update Dependencies

```bash
mage deps:update
mage deps:tidy
```

### Run Full CI Locally

```bash
mage ci
```

### Generate Mocks

```bash
go generate ./...
```

## Getting Help

- **Issues**: Check existing issues or create a new one
- **Discussions**: Use GitHub Discussions for questions
- **Documentation**: Read the docs thoroughly
- **Examples**: Look at existing code for patterns

## Recognition

Contributors will be:
- Listed in CONTRIBUTORS.md
- Mentioned in release notes
- Given credit in commit messages

Thank you for contributing to MAGE-X! 🎉
