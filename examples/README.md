# Go-Mage Namespace Interface Examples

This directory contains practical examples demonstrating the namespace interface architecture in various real-world scenarios.

## Directory Structure

```
examples/
├── README.md                          # This file
├── basic/                             # Basic usage examples
│   ├── simple-build/                  # Simple build script
│   └── ci-pipeline/                   # Complete CI pipeline
├── custom/                            # Custom middleware examples
│   └── logging-middleware/            # Logging middleware pattern
├── custom-implementations/            # Custom namespace implementations
│   ├── analytics/                     # Analytics implementation
│   ├── audit/                         # Audit implementation
│   ├── bench/                         # Benchmark implementation
│   ├── build/                         # Build implementation
│   ├── cli/                           # CLI implementation
│   ├── configure/                     # Configuration implementation
│   ├── deps/                          # Dependencies implementation
│   ├── docker/                        # Docker implementation
│   ├── docs/                          # Documentation implementation
│   ├── enterprise/                    # Enterprise implementation
│   ├── enterprise-config/             # Enterprise config implementation
│   ├── format/                        # Format implementation
│   ├── generate/                      # Generate implementation
│   ├── help/                          # Help implementation
│   ├── init/                          # Init implementation
│   ├── install/                       # Install implementation
│   ├── integrations/                  # Integrations implementation
│   ├── lint/                          # Lint implementation
│   ├── metrics/                       # Metrics implementation
│   ├── mod/                           # Module implementation
│   ├── recipes/                       # Recipes implementation
│   ├── release/                       # Release implementation
│   ├── releases/                      # Releases implementation
│   ├── security/                      # Security implementation
│   ├── test/                          # Test implementation
│   ├── tools/                         # Tools implementation
│   ├── version/                       # Version implementation
│   ├── vet/                           # Vet implementation
│   ├── wizard/                        # Wizard implementation
│   ├── workflow/                      # Workflow implementation
│   └── yaml/                          # YAML implementation
├── provider-pattern/                  # Provider pattern example
│   └── main.go                        # Provider usage demonstration
└── testing/                           # Testing examples
    └── unit-testing/                  # Unit test patterns
```

## Quick Start Examples

### Basic Build Script

```go
//go:build mage
package main

import "github.com/mrz1836/go-mage/pkg/mage"

func Build() error {
    build := mage.NewBuildNamespace()
    return build.Default()
}

func Test() error {
    test := mage.NewTestNamespace()
    return test.Unit()
}
```

### CI Pipeline

```go
//go:build mage
package main

import (
    "fmt"
    "github.com/mrz1836/go-mage/pkg/mage"
)

func CI() error {
    lint := mage.NewLintNamespace()
    if err := lint.Default(); err != nil {
        return fmt.Errorf("linting failed: %w", err)
    }
    
    test := mage.NewTestNamespace()
    if err := test.Coverage(); err != nil {
        return fmt.Errorf("tests failed: %w", err)
    }
    
    build := mage.NewBuildNamespace()
    if err := build.Default(); err != nil {
        return fmt.Errorf("build failed: %w", err)
    }
    
    return nil
}
```

### Custom Implementation

```go
//go:build mage
package main

import (
    "fmt"
    "time"
    "github.com/mrz1836/go-mage/pkg/mage"
)

type LoggingBuild struct {
    mage.BuildNamespace
}

func (l *LoggingBuild) Default() error {
    fmt.Println("🔨 Starting build...")
    start := time.Now()
    
    err := l.BuildNamespace.Default()
    
    duration := time.Since(start)
    if err != nil {
        fmt.Printf("❌ Build failed in %v: %v\n", duration, err)
    } else {
        fmt.Printf("✅ Build completed in %v\n", duration)
    }
    
    return err
}

func Build() error {
    baseBuild := mage.NewBuildNamespace()
    loggingBuild := &LoggingBuild{BuildNamespace: baseBuild}
    return loggingBuild.Default()
}
```

## Running Examples

Each example directory contains a complete, runnable example:

```bash
# Navigate to any example
cd examples/basic/simple-build

# Run the example
go run magefile.go build
go run magefile.go test
```

### Prerequisites

- Go 1.19 or later
- Mage installed (`go install github.com/magefile/mage@latest`)
- Go-Mage package available

## Example Categories

### Basic Examples
Perfect for getting started with the namespace interface architecture:
- **Simple Build**: Basic build, test, and lint operations
- **CI Pipeline**: Complete CI/CD pipeline implementation

### Custom Middleware Examples
Demonstrate how to extend functionality with middleware:
- **Logging Middleware**: Adding cross-cutting concerns like logging

### Custom Implementations
Complete implementations of each namespace interface:
- **Analytics**: Custom analytics namespace implementation
- **Audit**: Audit and compliance tracking implementation
- **Build**: Custom build process implementation
- **Test**: Custom test runner implementation
- **Docker**: Docker operations implementation
- And many more...

### Provider Pattern
Demonstrates the provider pattern for cloud integrations:
- **Provider Usage**: Shows how to use AWS and Azure providers

### Testing Examples
Testing patterns and strategies:
- **Unit Testing**: Testing functions that use namespaces

## Best Practices Demonstrated

### Type Safety
```go
// Compile-time verification
var _ mage.BuildNamespace = (*CustomBuild)(nil)

// Safe type assertion
if build, ok := registry.Get("build").(mage.BuildNamespace); ok {
    // Use build
}
```

### Error Handling
```go
func CI() error {
    if err := lint.Default(); err != nil {
        return fmt.Errorf("linting failed: %w", err)
    }
    // Continue with other steps...
}
```

### Dependency Injection
```go
func DeployApp(build mage.BuildNamespace, test mage.TestNamespace) error {
    // Function can work with any implementation
}
```

### Testing with Mocks
```go
func TestDeploy(t *testing.T) {
    mockBuild := &MockBuild{}
    mockTest := &MockTest{}
    
    err := DeployApp(mockBuild, mockTest)
    // Verify behavior
}
```

## Contributing Examples

To contribute a new example:

1. Create a directory under the appropriate category
2. Include a complete, runnable `magefile.go`
3. Add a `README.md` explaining the example
4. Include any necessary supporting files
5. Update this main README.md

### Example Template

```
examples/category/example-name/
├── README.md           # Explanation of the example
├── magefile.go         # Complete mage file
├── go.mod             # Go module (if needed)
└── main.go            # Sample application (if needed)
```

## Learning Path

1. **Start with Basic Examples**: Get familiar with the interface syntax
2. **Explore Custom Middleware**: Learn how to add cross-cutting concerns
3. **Study Custom Implementations**: Understand how to create full namespace implementations
4. **Practice with Provider Pattern**: Learn cloud provider integration
5. **Practice Testing**: Learn testing patterns with namespaces

## Getting Help

- **Documentation**: See [../docs/NAMESPACE_INTERFACES.md](../docs/NAMESPACE_INTERFACES.md)
- **API Reference**: See [../docs/API_REFERENCE.md](../docs/API_REFERENCE.md)
- **Migration Guide**: See [../docs/MIGRATION_GUIDE.md](../docs/MIGRATION_GUIDE.md)
- **Issues**: Report problems or request examples

## Example Index

| Example | Category | Complexity | Description |
|---------|----------|------------|-------------|
| [Simple Build](basic/simple-build/) | Basic | Beginner | Basic build, test, lint |
| [CI Pipeline](basic/ci-pipeline/) | Basic | Intermediate | Complete CI/CD pipeline |
| [Logging Middleware](custom/logging-middleware/) | Custom | Intermediate | Cross-cutting concerns |
| [Analytics](custom-implementations/analytics/) | Custom Implementation | Advanced | Custom analytics namespace |
| [Audit](custom-implementations/audit/) | Custom Implementation | Advanced | Audit and compliance tracking |
| [Build](custom-implementations/build/) | Custom Implementation | Advanced | Custom build process |
| [Test](custom-implementations/test/) | Custom Implementation | Advanced | Custom test runner |
| [Docker](custom-implementations/docker/) | Custom Implementation | Advanced | Docker operations |
| [Lint](custom-implementations/lint/) | Custom Implementation | Advanced | Custom linting implementation |
| [Provider Pattern](provider-pattern/) | Provider | Intermediate | AWS and Azure provider usage |
| [Unit Testing](testing/unit-testing/) | Testing | Intermediate | Testing with namespaces |
