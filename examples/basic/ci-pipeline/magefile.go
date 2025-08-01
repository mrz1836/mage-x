//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"

	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// CI runs the complete CI pipeline with all quality checks
func CI() error {
	utils.Info("🚀 Starting CI Pipeline...")

	// Step 1: Format check
	utils.Info("📝 Step 1: Checking code formatting...")
	format := mage.NewFormatNamespace()
	if err := format.Check(); err != nil {
		return fmt.Errorf("formatting check failed: %w", err)
	}
	utils.Info("✅ Code formatting is correct")

	// Step 2: Linting
	utils.Info("🔍 Step 2: Running linters...")
	lint := mage.NewLintNamespace()
	if err := lint.CI(); err != nil {
		return fmt.Errorf("linting failed: %w", err)
	}
	utils.Info("✅ Linting passed")

	// Step 3: Security scan
	utils.Info("🔒 Step 3: Running security scan...")
	security := mage.NewSecurityNamespace()
	if err := security.Scan(); err != nil {
		return fmt.Errorf("security scan failed: %w", err)
	}
	utils.Info("✅ Security scan passed")

	// Step 4: Tests with coverage
	utils.Info("🧪 Step 4: Running tests with coverage...")
	test := mage.NewTestNamespace()
	if err := test.Coverage(); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}
	utils.Info("✅ Tests passed with coverage")

	// Step 5: Race condition tests
	utils.Info("🏃 Step 5: Running race condition tests...")
	if err := test.Race(); err != nil {
		return fmt.Errorf("race condition tests failed: %w", err)
	}
	utils.Info("✅ Race condition tests passed")

	// Step 6: Build
	utils.Info("🔨 Step 6: Building application...")
	build := mage.NewBuildNamespace()
	if err := build.Default(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}
	utils.Info("✅ Build completed")

	// Step 7: Integration tests (if available)
	if hasIntegrationTests() {
		utils.Info("🔗 Step 7: Running integration tests...")
		if err := test.Integration(); err != nil {
			return fmt.Errorf("integration tests failed: %w", err)
		}
		utils.Info("✅ Integration tests passed")
	}

	utils.Info("🎉 CI Pipeline completed successfully!")
	return nil
}

// CIFast runs a faster CI pipeline for development
func CIFast() error {
	utils.Info("⚡ Starting Fast CI Pipeline...")

	// Quick lint
	utils.Info("🔍 Quick linting...")
	lint := mage.NewLintNamespace()
	if err := lint.Fast(); err != nil {
		return fmt.Errorf("fast linting failed: %w", err)
	}

	// Unit tests only
	utils.Info("🧪 Running unit tests...")
	test := mage.NewTestNamespace()
	if err := test.Unit(); err != nil {
		return fmt.Errorf("unit tests failed: %w", err)
	}

	// Quick build
	utils.Info("🔨 Quick build...")
	build := mage.NewBuildNamespace()
	if err := build.Default(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	utils.Info("✅ Fast CI completed!")
	return nil
}

// PreCommit runs pre-commit checks
func PreCommit() error {
	utils.Info("🔄 Running pre-commit checks...")

	// Format code automatically
	utils.Info("📝 Auto-formatting code...")
	format := mage.NewFormatNamespace()
	if err := format.Fix(); err != nil {
		return fmt.Errorf("auto-formatting failed: %w", err)
	}

	// Fix linting issues if possible
	utils.Info("🔧 Auto-fixing lint issues...")
	lint := mage.NewLintNamespace()
	if err := lint.Fix(); err != nil {
		utils.Warn("⚠️  Some lint issues need manual attention")
		// Don't fail on lint fix errors, just warn
	}

	// Run quick tests
	utils.Info("🧪 Running quick tests...")
	test := mage.NewTestNamespace()
	if err := test.Short(); err != nil {
		return fmt.Errorf("quick tests failed: %w", err)
	}

	utils.Info("✅ Pre-commit checks completed!")
	return nil
}

// Deploy builds and deploys the application
func Deploy() error {
	utils.Info("🚀 Starting deployment pipeline...")

	// Run full CI first
	if err := CI(); err != nil {
		return fmt.Errorf("CI pipeline failed: %w", err)
	}

	// Build for production
	utils.Info("🏭 Building for production...")
	build := mage.NewBuildNamespace()
	if err := build.All(); err != nil {
		return fmt.Errorf("production build failed: %w", err)
	}

	// Build Docker image if Dockerfile exists
	if hasDockerfile() {
		utils.Info("🐳 Building Docker image...")
		build := mage.NewBuildNamespace()
		if err := build.Docker(); err != nil {
			return fmt.Errorf("docker build failed: %w", err)
		}
	}

	// Generate docs
	utils.Info("📚 Generating documentation...")
	docs := mage.NewDocsNamespace()
	if err := docs.Build(); err != nil {
		utils.Warn("⚠️  Documentation generation failed, continuing...")
	}

	utils.Info("🎉 Deployment pipeline completed!")
	return nil
}

// Benchmark runs performance benchmarks
func Benchmark() error {
	utils.Info("📊 Running performance benchmarks...")

	test := mage.NewTestNamespace()
	if err := test.Bench(); err != nil {
		return fmt.Errorf("benchmarks failed: %w", err)
	}

	utils.Info("✅ Benchmarks completed!")
	return nil
}

// QualityGate runs all quality checks
func QualityGate() error {
	utils.Info("🎯 Running quality gate checks...")

	// Code metrics
	utils.Info("📊 Analyzing code metrics...")
	metrics := mage.NewMetricsNamespace()
	if err := metrics.LOC(); err != nil {
		utils.Warn("⚠️  Code metrics analysis failed")
	}

	// Complexity analysis
	if err := metrics.Complexity(); err != nil {
		utils.Warn("⚠️  Complexity analysis failed")
	}

	// Quality analysis
	if err := metrics.Quality(); err != nil {
		utils.Warn("⚠️  Quality analysis failed")
	}

	// Security audit
	utils.Info("🔒 Running security audit...")
	security := mage.NewSecurityNamespace()
	if err := security.Audit(); err != nil {
		utils.Warn("⚠️  Security audit failed")
	}

	// Dependency audit
	utils.Info("📦 Auditing dependencies...")
	deps := mage.NewDepsNamespace()
	if err := deps.Audit(); err != nil {
		return fmt.Errorf("dependency audit failed: %w", err)
	}

	utils.Info("✅ Quality gate passed!")
	return nil
}

// Helper functions

func hasIntegrationTests() bool {
	// Check if integration tests exist
	if _, err := os.Stat("tests/integration"); err == nil {
		return true
	}
	if _, err := os.Stat("integration_test.go"); err == nil {
		return true
	}
	return false
}

func hasDockerfile() bool {
	_, err := os.Stat("Dockerfile")
	return err == nil
}

// Individual namespace examples

// Build builds the application
func Build() error {
	build := mage.NewBuildNamespace()
	return build.Default()
}

// Test runs tests
func Test() error {
	test := mage.NewTestNamespace()
	return test.Default()
}

// Lint runs linting
func Lint() error {
	lint := mage.NewLintNamespace()
	return lint.Default()
}

// Format formats code
func Format() error {
	format := mage.NewFormatNamespace()
	return format.Default()
}

// Clean cleans build artifacts
func Clean() error {
	build := mage.NewBuildNamespace()
	return build.Clean()
}
