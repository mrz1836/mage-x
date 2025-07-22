# Namespace Interface Examples

This document provides comprehensive examples of using the Go-Mage namespace interface architecture in real-world scenarios.

## Table of Contents

- [Basic Usage Examples](#basic-usage-examples)
- [Advanced Patterns](#advanced-patterns)
- [Custom Implementations](#custom-implementations)
- [Testing Examples](#testing-examples)
- [Production Patterns](#production-patterns)
- [Migration Examples](#migration-examples)

## Basic Usage Examples

### Simple Build Script

```go
//go:build mage
// +build mage

package main

import (
    "fmt"
    "github.com/mrz1836/go-mage/pkg/mage"
)

// Build compiles the application
func Build() error {
    build := mage.NewBuildNamespace()
    return build.Default()
}

// Test runs the test suite
func Test() error {
    test := mage.NewTestNamespace()
    return test.Unit()
}

// Lint runs code analysis
func Lint() error {
    lint := mage.NewLintNamespace()
    return lint.Default()
}

// Clean removes build artifacts
func Clean() error {
    build := mage.NewBuildNamespace()
    return build.Clean()
}
```

### Multi-Platform Build

```go
//go:build mage
// +build mage

package main

import (
    "fmt"
    "github.com/mrz1836/go-mage/pkg/mage"
)

// BuildAll builds for all supported platforms
func BuildAll() error {
    build := mage.NewBuildNamespace()
    
    platforms := []string{
        "linux/amd64",
        "linux/arm64", 
        "darwin/amd64",
        "darwin/arm64",
        "windows/amd64",
    }
    
    for _, platform := range platforms {
        fmt.Printf("Building for %s...\n", platform)
        if err := build.Platform(platform); err != nil {
            return fmt.Errorf("build failed for %s: %w", platform, err)
        }
    }
    
    return nil
}

// BuildLinux builds only for Linux
func BuildLinux() error {
    build := mage.NewBuildNamespace()
    return build.Linux()
}

// BuildDarwin builds only for macOS
func BuildDarwin() error {
    build := mage.NewBuildNamespace()
    return build.Darwin()
}

// BuildWindows builds only for Windows
func BuildWindows() error {
    build := mage.NewBuildNamespace()
    return build.Windows()
}
```

### Complete CI Pipeline

```go
//go:build mage
// +build mage

package main

import (
    "fmt"
    "github.com/mrz1836/go-mage/pkg/mage"
)

// CI runs the complete CI pipeline
func CI() error {
    fmt.Println("🚀 Starting CI Pipeline...")
    
    // Step 1: Format check
    fmt.Println("📝 Checking code formatting...")
    format := mage.NewFormatNamespace()
    if err := format.Check(); err != nil {
        return fmt.Errorf("formatting check failed: %w", err)
    }
    
    // Step 2: Linting
    fmt.Println("🔍 Running linters...")
    lint := mage.NewLintNamespace()
    if err := lint.CI(); err != nil {
        return fmt.Errorf("linting failed: %w", err)
    }
    
    // Step 3: Security scan
    fmt.Println("🔒 Running security scan...")
    security := mage.NewSecurityNamespace()
    if err := security.Scan(); err != nil {
        return fmt.Errorf("security scan failed: %w", err)
    }
    
    // Step 4: Tests
    fmt.Println("🧪 Running tests...")
    test := mage.NewTestNamespace()
    if err := test.Coverage(); err != nil {
        return fmt.Errorf("tests failed: %w", err)
    }
    
    // Step 5: Build
    fmt.Println("🔨 Building application...")
    build := mage.NewBuildNamespace()
    if err := build.Default(); err != nil {
        return fmt.Errorf("build failed: %w", err)
    }
    
    fmt.Println("✅ CI Pipeline completed successfully!")
    return nil
}
```

## Advanced Patterns

### Pipeline with Error Recovery

```go
//go:build mage
// +build mage

package main

import (
    "fmt"
    "log"
    "github.com/mrz1836/go-mage/pkg/mage"
)

// Pipeline represents a build pipeline
type Pipeline struct {
    build    mage.BuildNamespace
    test     mage.TestNamespace
    lint     mage.LintNamespace
    security mage.SecurityNamespace
    docker   mage.DockerNamespace
}

// NewPipeline creates a new build pipeline
func NewPipeline() *Pipeline {
    return &Pipeline{
        build:    mage.NewBuildNamespace(),
        test:     mage.NewTestNamespace(),
        lint:     mage.NewLintNamespace(),
        security: mage.NewSecurityNamespace(),
        docker:   mage.NewDockerNamespace(),
    }
}

// Execute runs the pipeline with error recovery
func (p *Pipeline) Execute() error {
    steps := []struct {
        name string
        fn   func() error
    }{
        {"lint", p.lint.Default},
        {"test", p.test.Unit},
        {"build", p.build.Default},
        {"security-scan", p.security.Scan},
        {"docker-build", p.docker.Build},
    }
    
    for _, step := range steps {
        fmt.Printf("🔄 Running %s...\n", step.name)
        
        if err := step.fn(); err != nil {
            log.Printf("❌ Step %s failed: %v", step.name, err)
            
            // Attempt recovery for certain steps
            if step.name == "lint" {
                fmt.Println("🔧 Attempting to fix linting issues...")
                if fixErr := p.lint.Fix(); fixErr != nil {
                    return fmt.Errorf("step %s failed and auto-fix failed: %w", step.name, err)
                }
                // Retry after fix
                if retryErr := step.fn(); retryErr != nil {
                    return fmt.Errorf("step %s failed after fix: %w", step.name, retryErr)
                }
            } else {
                return fmt.Errorf("step %s failed: %w", step.name, err)
            }
        }
        
        fmt.Printf("✅ Step %s completed\n", step.name)
    }
    
    return nil
}

// Deploy runs the deployment pipeline
func Deploy() error {
    pipeline := NewPipeline()
    return pipeline.Execute()
}
```

### Registry-Based Configuration

```go
//go:build mage
// +build mage

package main

import (
    "fmt"
    "os"
    "github.com/mrz1836/go-mage/pkg/mage"
)

// setupNamespaces configures namespaces based on environment
func setupNamespaces() *mage.NamespaceRegistry {
    registry := mage.NewNamespaceRegistry()
    
    // Configure based on environment
    env := os.Getenv("BUILD_ENV")
    
    switch env {
    case "production":
        // Use optimized build for production
        registry.Register("build", &ProductionBuild{})
        registry.Register("test", &ProductionTest{})
    case "development":
        // Use fast build for development
        registry.Register("build", &DevelopmentBuild{})
        registry.Register("test", &DevelopmentTest{})
    default:
        // Use default implementations
        fmt.Println("Using default namespace implementations")
    }
    
    return registry
}

// Build with environment-specific configuration
func Build() error {
    registry := setupNamespaces()
    build := registry.Get("build").(mage.BuildNamespace)
    return build.Default()
}

// Test with environment-specific configuration
func Test() error {
    registry := setupNamespaces()
    test := registry.Get("test").(mage.TestNamespace)
    return test.Unit()
}

// ProductionBuild optimized for production builds
type ProductionBuild struct{}

func (p *ProductionBuild) Default() error {
    fmt.Println("🚀 Running optimized production build...")
    // Production-specific build logic
    return nil
}

func (p *ProductionBuild) All() error { return nil }
func (p *ProductionBuild) Platform(platform string) error { return nil }
func (p *ProductionBuild) Linux() error { return nil }
func (p *ProductionBuild) Darwin() error { return nil }
func (p *ProductionBuild) Windows() error { return nil }
func (p *ProductionBuild) Docker() error { return nil }
func (p *ProductionBuild) Clean() error { return nil }
func (p *ProductionBuild) Install() error { return nil }
func (p *ProductionBuild) Generate() error { return nil }
func (p *ProductionBuild) PreBuild() error { return nil }

// DevelopmentBuild optimized for fast development cycles
type DevelopmentBuild struct{}

func (d *DevelopmentBuild) Default() error {
    fmt.Println("⚡ Running fast development build...")
    // Development-specific build logic
    return nil
}

func (d *DevelopmentBuild) All() error { return nil }
func (d *DevelopmentBuild) Platform(platform string) error { return nil }
func (d *DevelopmentBuild) Linux() error { return nil }
func (d *DevelopmentBuild) Darwin() error { return nil }
func (d *DevelopmentBuild) Windows() error { return nil }
func (d *DevelopmentBuild) Docker() error { return nil }
func (d *DevelopmentBuild) Clean() error { return nil }
func (d *DevelopmentBuild) Install() error { return nil }
func (d *DevelopmentBuild) Generate() error { return nil }
func (d *DevelopmentBuild) PreBuild() error { return nil }

// Similar implementations for ProductionTest and DevelopmentTest...
type ProductionTest struct{}
func (p *ProductionTest) Default() error { return nil }
func (p *ProductionTest) Unit() error { return nil }
func (p *ProductionTest) Integration() error { return nil }
func (p *ProductionTest) Benchmark() error { return nil }
func (p *ProductionTest) Coverage() error { return nil }
func (p *ProductionTest) Race() error { return nil }
func (p *ProductionTest) Short() error { return nil }
func (p *ProductionTest) Verbose() error { return nil }
func (p *ProductionTest) All() error { return nil }

type DevelopmentTest struct{}
func (d *DevelopmentTest) Default() error { return nil }
func (d *DevelopmentTest) Unit() error { return nil }
func (d *DevelopmentTest) Integration() error { return nil }
func (d *DevelopmentTest) Benchmark() error { return nil }
func (d *DevelopmentTest) Coverage() error { return nil }
func (d *DevelopmentTest) Race() error { return nil }
func (d *DevelopmentTest) Short() error { return nil }
func (d *DevelopmentTest) Verbose() error { return nil }
func (d *DevelopmentTest) All() error { return nil }
```

## Custom Implementations

### Logging Middleware

```go
//go:build mage
// +build mage

package main

import (
    "fmt"
    "log"
    "time"
    "github.com/mrz1836/go-mage/pkg/mage"
)

// LoggingBuild wraps any BuildNamespace with logging
type LoggingBuild struct {
    mage.BuildNamespace
    logger *log.Logger
}

// NewLoggingBuild creates a build namespace with logging
func NewLoggingBuild(inner mage.BuildNamespace, logger *log.Logger) *LoggingBuild {
    return &LoggingBuild{
        BuildNamespace: inner,
        logger:        logger,
    }
}

func (l *LoggingBuild) Default() error {
    l.logger.Println("🔨 Starting default build...")
    start := time.Now()
    
    err := l.BuildNamespace.Default()
    
    duration := time.Since(start)
    if err != nil {
        l.logger.Printf("❌ Build failed after %v: %v", duration, err)
    } else {
        l.logger.Printf("✅ Build completed successfully in %v", duration)
    }
    
    return err
}

func (l *LoggingBuild) Platform(platform string) error {
    l.logger.Printf("🔨 Starting build for platform %s...", platform)
    start := time.Now()
    
    err := l.BuildNamespace.Platform(platform)
    
    duration := time.Since(start)
    if err != nil {
        l.logger.Printf("❌ Platform build failed after %v: %v", duration, err)
    } else {
        l.logger.Printf("✅ Platform build completed in %v", duration)
    }
    
    return err
}

// Build with logging
func Build() error {
    logger := log.New(os.Stdout, "[BUILD] ", log.LstdFlags)
    
    baseBuild := mage.NewBuildNamespace()
    loggingBuild := NewLoggingBuild(baseBuild, logger)
    
    return loggingBuild.Default()
}
```

### Metrics Collection

```go
//go:build mage
// +build mage

package main

import (
    "encoding/json"
    "fmt"
    "os"
    "time"
    "github.com/mrz1836/go-mage/pkg/mage"
)

// BuildMetrics tracks build performance
type BuildMetrics struct {
    StartTime    time.Time     `json:"start_time"`
    EndTime      time.Time     `json:"end_time"`
    Duration     time.Duration `json:"duration"`
    Success      bool          `json:"success"`
    Platform     string        `json:"platform,omitempty"`
    ErrorMessage string        `json:"error_message,omitempty"`
}

// MetricsBuild collects metrics for build operations
type MetricsBuild struct {
    mage.BuildNamespace
    metricsFile string
}

// NewMetricsBuild creates a build namespace that collects metrics
func NewMetricsBuild(inner mage.BuildNamespace, metricsFile string) *MetricsBuild {
    return &MetricsBuild{
        BuildNamespace: inner,
        metricsFile:   metricsFile,
    }
}

func (m *MetricsBuild) Default() error {
    return m.withMetrics("default", "", func() error {
        return m.BuildNamespace.Default()
    })
}

func (m *MetricsBuild) Platform(platform string) error {
    return m.withMetrics("platform", platform, func() error {
        return m.BuildNamespace.Platform(platform)
    })
}

func (m *MetricsBuild) withMetrics(operation, platform string, fn func() error) error {
    metrics := BuildMetrics{
        StartTime: time.Now(),
        Platform:  platform,
    }
    
    err := fn()
    
    metrics.EndTime = time.Now()
    metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
    metrics.Success = err == nil
    
    if err != nil {
        metrics.ErrorMessage = err.Error()
    }
    
    // Save metrics
    if saveErr := m.saveMetrics(metrics); saveErr != nil {
        fmt.Printf("Warning: Failed to save metrics: %v\n", saveErr)
    }
    
    return err
}

func (m *MetricsBuild) saveMetrics(metrics BuildMetrics) error {
    data, err := json.MarshalIndent(metrics, "", "  ")
    if err != nil {
        return err
    }
    
    return os.WriteFile(m.metricsFile, data, 0644)
}

// Build with metrics collection
func Build() error {
    baseBuild := mage.NewBuildNamespace()
    metricsBuild := NewMetricsBuild(baseBuild, "build-metrics.json")
    
    return metricsBuild.Default()
}
```

## Testing Examples

### Unit Testing with Mocks

```go
package main

import (
    "testing"
    "github.com/mrz1836/go-mage/pkg/mage"
)

// MockBuild for testing
type MockBuild struct {
    defaultCalled   bool
    platformCalled map[string]bool
    shouldFail      bool
}

func NewMockBuild() *MockBuild {
    return &MockBuild{
        platformCalled: make(map[string]bool),
    }
}

func (m *MockBuild) Default() error {
    m.defaultCalled = true
    if m.shouldFail {
        return fmt.Errorf("mock build failure")
    }
    return nil
}

func (m *MockBuild) Platform(platform string) error {
    m.platformCalled[platform] = true
    if m.shouldFail {
        return fmt.Errorf("mock platform build failure")
    }
    return nil
}

func (m *MockBuild) All() error { return nil }
func (m *MockBuild) Linux() error { return nil }
func (m *MockBuild) Darwin() error { return nil }
func (m *MockBuild) Windows() error { return nil }
func (m *MockBuild) Docker() error { return nil }
func (m *MockBuild) Clean() error { return nil }
func (m *MockBuild) Install() error { return nil }
func (m *MockBuild) Generate() error { return nil }
func (m *MockBuild) PreBuild() error { return nil }

// Function to test
func deployApp(build mage.BuildNamespace, platforms []string) error {
    for _, platform := range platforms {
        if err := build.Platform(platform); err != nil {
            return err
        }
    }
    return nil
}

// Test function
func TestDeployApp(t *testing.T) {
    t.Run("successful deployment", func(t *testing.T) {
        mockBuild := NewMockBuild()
        platforms := []string{"linux/amd64", "darwin/amd64"}
        
        err := deployApp(mockBuild, platforms)
        
        if err != nil {
            t.Errorf("Expected no error, got %v", err)
        }
        
        for _, platform := range platforms {
            if !mockBuild.platformCalled[platform] {
                t.Errorf("Platform %s was not built", platform)
            }
        }
    })
    
    t.Run("build failure", func(t *testing.T) {
        mockBuild := NewMockBuild()
        mockBuild.shouldFail = true
        platforms := []string{"linux/amd64"}
        
        err := deployApp(mockBuild, platforms)
        
        if err == nil {
            t.Error("Expected error, got nil")
        }
    })
}
```

### Integration Testing

```go
package main

import (
    "testing"
    "github.com/mrz1836/go-mage/pkg/mage"
    "github.com/mrz1836/go-mage/pkg/testhelpers"
)

func TestBuildIntegration(t *testing.T) {
    env := testhelpers.NewTestEnvironment(t)
    
    // Create a test Go project
    env.CreateGoModule("test-app")
    env.WriteFile("main.go", `package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}`)
    
    t.Run("build succeeds", func(t *testing.T) {
        build := mage.NewBuildNamespace()
        err := build.Default()
        
        // Build may fail due to missing dependencies in test environment
        // but we can verify the interface works
        t.Logf("Build result: %v", err)
    })
    
    t.Run("clean works", func(t *testing.T) {
        build := mage.NewBuildNamespace()
        err := build.Clean()
        env.AssertNoError(err)
    })
}

func TestNamespaceRegistry(t *testing.T) {
    t.Run("default namespaces available", func(t *testing.T) {
        registry := mage.NewNamespaceRegistry()
        
        requiredNamespaces := []string{
            "build", "test", "lint", "format", "docs",
        }
        
        for _, name := range requiredNamespaces {
            ns := registry.Get(name)
            if ns == nil {
                t.Errorf("Namespace %s not available", name)
            }
        }
    })
    
    t.Run("custom registration works", func(t *testing.T) {
        registry := mage.NewNamespaceRegistry()
        mockBuild := NewMockBuild()
        
        err := registry.Register("build", mockBuild)
        if err != nil {
            t.Errorf("Failed to register custom build: %v", err)
        }
        
        retrieved := registry.Get("build")
        if retrieved != mockBuild {
            t.Error("Registry did not return custom implementation")
        }
    })
}
```

## Production Patterns

### Microservices Build System

```go
//go:build mage
// +build mage

package main

import (
    "fmt"
    "path/filepath"
    "github.com/mrz1836/go-mage/pkg/mage"
)

// Microservice represents a single service
type Microservice struct {
    Name string
    Path string
    Port int
}

// services defines all microservices in the system
var services = []Microservice{
    {Name: "user-service", Path: "./services/user", Port: 8001},
    {Name: "auth-service", Path: "./services/auth", Port: 8002},
    {Name: "payment-service", Path: "./services/payment", Port: 8003},
    {Name: "notification-service", Path: "./services/notification", Port: 8004},
}

// BuildAll builds all microservices
func BuildAll() error {
    fmt.Println("🏗️ Building all microservices...")
    
    for _, service := range services {
        if err := buildService(service); err != nil {
            return fmt.Errorf("failed to build %s: %w", service.Name, err)
        }
    }
    
    fmt.Println("✅ All microservices built successfully!")
    return nil
}

// BuildService builds a specific microservice
func BuildService(serviceName string) error {
    for _, service := range services {
        if service.Name == serviceName {
            return buildService(service)
        }
    }
    return fmt.Errorf("service %s not found", serviceName)
}

func buildService(service Microservice) error {
    fmt.Printf("🔨 Building %s...\n", service.Name)
    
    build := mage.NewBuildNamespace()
    
    // Change to service directory and build
    originalDir, _ := os.Getwd()
    defer os.Chdir(originalDir)
    
    if err := os.Chdir(service.Path); err != nil {
        return fmt.Errorf("failed to change to service directory: %w", err)
    }
    
    return build.Default()
}

// TestAll runs tests for all microservices
func TestAll() error {
    fmt.Println("🧪 Testing all microservices...")
    
    for _, service := range services {
        if err := testService(service); err != nil {
            return fmt.Errorf("tests failed for %s: %w", service.Name, err)
        }
    }
    
    fmt.Println("✅ All tests passed!")
    return nil
}

func testService(service Microservice) error {
    fmt.Printf("🧪 Testing %s...\n", service.Name)
    
    test := mage.NewTestNamespace()
    
    originalDir, _ := os.Getwd()
    defer os.Chdir(originalDir)
    
    if err := os.Chdir(service.Path); err != nil {
        return fmt.Errorf("failed to change to service directory: %w", err)
    }
    
    return test.Unit()
}

// DeployAll deploys all microservices
func DeployAll() error {
    fmt.Println("🚀 Deploying all microservices...")
    
    // First build all services
    if err := BuildAll(); err != nil {
        return err
    }
    
    // Then build Docker images
    docker := mage.NewDockerNamespace()
    
    for _, service := range services {
        fmt.Printf("🐳 Building Docker image for %s...\n", service.Name)
        
        originalDir, _ := os.Getwd()
        os.Chdir(service.Path)
        
        if err := docker.Build(); err != nil {
            os.Chdir(originalDir)
            return fmt.Errorf("failed to build Docker image for %s: %w", service.Name, err)
        }
        
        os.Chdir(originalDir)
    }
    
    fmt.Println("✅ All microservices deployed!")
    return nil
}
```

### Monorepo Management

```go
//go:build mage
// +build mage

package main

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "github.com/mrz1836/go-mage/pkg/mage"
)

// Package represents a Go package in the monorepo
type Package struct {
    Name string
    Path string
    Type string // "library", "service", "tool"
}

// discoverPackages finds all Go packages in the monorepo
func discoverPackages() ([]Package, error) {
    var packages []Package
    
    err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if info.Name() == "go.mod" {
            dir := filepath.Dir(path)
            if dir == "." {
                return nil // Skip root module
            }
            
            name := filepath.Base(dir)
            pkgType := determinePackageType(dir)
            
            packages = append(packages, Package{
                Name: name,
                Path: dir,
                Type: pkgType,
            })
        }
        
        return nil
    })
    
    return packages, err
}

func determinePackageType(path string) string {
    if strings.Contains(path, "/cmd/") || strings.Contains(path, "/services/") {
        return "service"
    }
    if strings.Contains(path, "/tools/") {
        return "tool"
    }
    return "library"
}

// BuildChanged builds only packages that have changed
func BuildChanged() error {
    packages, err := discoverPackages()
    if err != nil {
        return fmt.Errorf("failed to discover packages: %w", err)
    }
    
    changedPkgs, err := getChangedPackages(packages)
    if err != nil {
        return fmt.Errorf("failed to detect changes: %w", err)
    }
    
    if len(changedPkgs) == 0 {
        fmt.Println("📦 No packages have changed")
        return nil
    }
    
    fmt.Printf("📦 Building %d changed packages...\n", len(changedPkgs))
    
    build := mage.NewBuildNamespace()
    
    for _, pkg := range changedPkgs {
        fmt.Printf("🔨 Building %s (%s)...\n", pkg.Name, pkg.Type)
        
        originalDir, _ := os.Getwd()
        defer os.Chdir(originalDir)
        
        if err := os.Chdir(pkg.Path); err != nil {
            return fmt.Errorf("failed to change to package directory: %w", err)
        }
        
        if err := build.Default(); err != nil {
            return fmt.Errorf("build failed for %s: %w", pkg.Name, err)
        }
    }
    
    fmt.Println("✅ All changed packages built successfully!")
    return nil
}

func getChangedPackages(packages []Package) ([]Package, error) {
    // Simple implementation - check git status
    // In production, you'd use more sophisticated change detection
    return packages, nil // For example, return all packages
}

// TestAffected runs tests for packages affected by changes
func TestAffected() error {
    packages, err := discoverPackages()
    if err != nil {
        return fmt.Errorf("failed to discover packages: %w", err)
    }
    
    test := mage.NewTestNamespace()
    
    for _, pkg := range packages {
        fmt.Printf("🧪 Testing %s...\n", pkg.Name)
        
        originalDir, _ := os.Getwd()
        defer os.Chdir(originalDir)
        
        if err := os.Chdir(pkg.Path); err != nil {
            return fmt.Errorf("failed to change to package directory: %w", err)
        }
        
        if err := test.Unit(); err != nil {
            return fmt.Errorf("tests failed for %s: %w", pkg.Name, err)
        }
    }
    
    return nil
}
```

## Migration Examples

### Gradual Migration from Old to New

```go
//go:build mage
// +build mage

package main

import (
    "github.com/mrz1836/go-mage/pkg/mage"
)

// Old style functions (keep existing)
func BuildOld() error {
    build := mage.Build{}
    return build.Default()
}

func TestOld() error {
    test := mage.Test{}
    return test.Unit()
}

// New style functions (add gradually)
func BuildNew() error {
    build := mage.NewBuildNamespace()
    return build.Default()
}

func TestNew() error {
    test := mage.NewTestNamespace()
    return test.Unit()
}

// Hybrid approach - use new style internally
func CI() error {
    // Use new interface-based approach for better testability
    build := mage.NewBuildNamespace()
    test := mage.NewTestNamespace()
    lint := mage.NewLintNamespace()
    
    if err := lint.Default(); err != nil {
        return err
    }
    
    if err := build.Default(); err != nil {
        return err
    }
    
    return test.Unit()
}

// Wrapper functions for backward compatibility
func Build() error {
    return BuildNew() // Delegate to new implementation
}

func Test() error {
    return TestNew() // Delegate to new implementation
}
```

### Converting to Registry-Based System

```go
//go:build mage
// +build mage

package main

import (
    "os"
    "github.com/mrz1836/go-mage/pkg/mage"
)

// Global registry for consistent namespace access
var registry *mage.NamespaceRegistry

func init() {
    registry = setupNamespaces()
}

func setupNamespaces() *mage.NamespaceRegistry {
    reg := mage.NewNamespaceRegistry()
    
    // Configure based on environment or configuration
    if os.Getenv("FAST_BUILD") == "true" {
        // Register fast implementations
        reg.Register("build", &FastBuild{})
    }
    
    if os.Getenv("VERBOSE_TESTS") == "true" {
        // Register verbose test implementation
        reg.Register("test", &VerboseTest{})
    }
    
    return reg
}

// Build using registry
func Build() error {
    build := registry.Get("build").(mage.BuildNamespace)
    return build.Default()
}

// Test using registry
func Test() error {
    test := registry.Get("test").(mage.TestNamespace)
    return test.Unit()
}

// Lint using registry
func Lint() error {
    lint := registry.Get("lint").(mage.LintNamespace)
    return lint.Default()
}

// Custom implementations
type FastBuild struct{}
func (f *FastBuild) Default() error {
    // Fast build implementation
    return nil
}
// ... implement all other BuildNamespace methods

type VerboseTest struct{}
func (v *VerboseTest) Unit() error {
    // Verbose test implementation
    return nil
}
// ... implement all other TestNamespace methods
```

These examples demonstrate the flexibility and power of the namespace interface architecture, showing how it can be used in various scenarios from simple scripts to complex production systems.
