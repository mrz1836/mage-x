//go:build mage
// +build mage

package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"

	// Import tasks from MAGE-X
	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// ErrDatabaseURLRequired is returned when DATABASE_URL environment variable is not set
var ErrDatabaseURLRequired = errors.New("DATABASE_URL environment variable is required")

// ErrEnvironmentRequired is returned when environment parameter is missing
var ErrEnvironmentRequired = errors.New("environment is required: dev, staging, or prod")

// ErrInvalidEnvironment is returned when environment parameter is invalid
var ErrInvalidEnvironment = errors.New("invalid environment: must be dev, staging, or prod")

// Re-export types
type (
	Build = mage.Build
	Test  = mage.Test
	Lint  = mage.Lint
	Deps  = mage.Deps
	Tools = mage.Tools
)

// Default target
func Default() error {
	var b Build
	return b.Default()
}

// Custom namespace for project-specific tasks
type Custom mg.Namespace

// Deploy deploys the application to the specified environment
func (Custom) Deploy(env string) error {
	if env == "" {
		return ErrEnvironmentRequired
	}

	// Validate environment
	validEnvs := []string{"dev", "staging", "prod"}
	valid := false
	for _, e := range validEnvs {
		if e == env {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("%w: %s", ErrInvalidEnvironment, env)
	}

	// Ensure we have a build
	var b Build
	mg.Deps(b.Default)

	fmt.Printf("🚀 Deploying to %s environment...\n", env)

	// Environment-specific deployment
	switch env {
	case "dev":
		return deployDev()
	case "staging":
		return deployStaging()
	case "prod":
		return deployProd()
	}

	return nil
}

// deployDev deploys to development environment
func deployDev() error {
	utils.Info("📦 Deploying to development server...")

	// Example: Copy binary to dev server
	if err := sh.Run("scp", "bin/myapp", "dev-user@dev-server:/opt/myapp/myapp-new"); err != nil {
		return err
	}

	// Restart service
	if err := sh.Run("ssh", "dev-user@dev-server", "sudo systemctl restart myapp"); err != nil {
		return err
	}

	utils.Info("✅ Development deployment completed!")
	return nil
}

// deployStaging deploys to staging environment
func deployStaging() error {
	utils.Info("📦 Deploying to staging server...")

	// Run tests first
	var t Test
	mg.Deps(t.Default)

	// Build Docker image
	if err := sh.Run("docker", "build", "-t", "myapp:staging", "."); err != nil {
		return err
	}

	// Push to registry
	if err := sh.Run("docker", "push", "myregistry.com/myapp:staging"); err != nil {
		return err
	}

	// Update staging cluster
	if err := sh.Run("kubectl", "set", "image", "deployment/myapp", "myapp=myregistry.com/myapp:staging", "-n", "staging"); err != nil {
		return err
	}

	utils.Info("✅ Staging deployment completed!")
	return nil
}

// deployProd deploys to production environment
func deployProd() error {
	utils.Info("🚨 Production deployment starting...")

	// Confirm production deployment
	fmt.Print("Are you sure you want to deploy to PRODUCTION? (yes/no): ")
	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		// Continue with empty response if user input fails
		response = ""
	}

	if response != "yes" {
		utils.Info("❌ Production deployment canceled")
		return nil
	}

	// Run full test suite
	var t Test
	mg.Deps(t.CI)

	// Build for production
	if err := os.Setenv("GO_BUILD_TAGS", "prod"); err != nil {
		return fmt.Errorf("failed to set GO_BUILD_TAGS: %w", err)
	}
	var b Build
	mg.Deps(b.All)

	// Create release tag
	version := time.Now().Format("v20060102-150405")
	if err := sh.Run("git", "tag", "-a", version, "-m", "Production release "+version); err != nil {
		return err
	}

	// Push tag
	if err := sh.Run("git", "push", "origin", version); err != nil {
		return err
	}

	fmt.Printf("✅ Production deployment completed! Version: %s\n", version)
	return nil
}

// Database namespace for database operations
type DB mg.Namespace

// Migrate runs database migrations
func (DB) Migrate() error {
	utils.Info("🗄️  Running database migrations...")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return ErrDatabaseURLRequired
	}

	// Example using golang-migrate
	return sh.Run("migrate",
		"-path", "./migrations",
		"-database", dbURL,
		"up")
}

// Rollback rolls back the last migration
func (DB) Rollback() error {
	utils.Info("⏪ Rolling back last migration...")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return ErrDatabaseURLRequired
	}

	return sh.Run("migrate",
		"-path", "./migrations",
		"-database", dbURL,
		"down", "1")
}

// Seed seeds the database with test data
func (DB) Seed() error {
	// Ensure migrations are run first
	mg.Deps(DB{}.Migrate)

	utils.Info("🌱 Seeding database...")
	return sh.Run("go", "run", "cmd/seed/main.go")
}

// Reset drops, recreates, and seeds the database
func (DB) Reset() error {
	utils.Info("🔄 Resetting database...")

	// Drop all tables
	utils.Info("💥 Dropping all tables...")
	if err := sh.Run("go", "run", "cmd/dbutil/main.go", "drop"); err != nil {
		return err
	}

	// Run migrations
	if err := (DB{}).Migrate(); err != nil {
		return err
	}

	// Seed data
	if err := (DB{}).Seed(); err != nil {
		return err
	}

	utils.Info("✅ Database reset completed!")
	return nil
}

// Docker namespace for Docker operations
type Docker mg.Namespace

// Build builds the Docker image
func (Docker) Build() error {
	utils.Info("🐳 Building Docker image...")

	// Get version from git or use "latest"
	version, err := sh.Output("git", "describe", "--tags", "--always", "--dirty")
	if err != nil {
		version = "latest"
	}

	return sh.Run("docker", "build",
		"-t", fmt.Sprintf("myapp:%s", version),
		"-t", "myapp:latest",
		".")
}

// Run runs the application in Docker
func (Docker) Run() error {
	// Build image first
	mg.Deps(Docker{}.Build)

	utils.Info("🏃 Running Docker container...")

	return sh.Run("docker", "run",
		"--rm",
		"-p", "8080:8080",
		"--env-file", ".env",
		"--name", "myapp",
		"myapp:latest")
}

// Push pushes the Docker image to registry
func (Docker) Push() error {
	utils.Info("📤 Pushing Docker image...")

	registry := os.Getenv("DOCKER_REGISTRY")
	if registry == "" {
		registry = "docker.io/myorg"
	}

	version, err := sh.Output("git", "describe", "--tags", "--always", "--dirty")
	if err != nil {
		version = "latest"
	}

	// Tag for registry
	localTag := fmt.Sprintf("myapp:%s", version)
	remoteTag := fmt.Sprintf("%s/myapp:%s", registry, version)

	if err := sh.Run("docker", "tag", localTag, remoteTag); err != nil {
		return err
	}

	// Push to registry
	return sh.Run("docker", "push", remoteTag)
}

// Dev runs the application in development mode with hot reload
func Dev() error {
	utils.Info("🔧 Starting development server with hot reload...")

	// Install air if not present
	if err := sh.Run("which", "air"); err != nil {
		utils.Info("Installing air for hot reload...")
		if err := sh.Run("go", "install", "github.com/cosmtrek/air@latest"); err != nil {
			return err
		}
	}

	// Run with air
	return sh.Run("air")
}

// Generate namespace for code generation
type Generate mg.Namespace

// Mocks generates mock files
func (Generate) Mocks() error {
	utils.Info("🎭 Generating mocks...")

	// Find all interfaces to mock
	interfaces := []struct {
		source string
		dest   string
		pkg    string
	}{
		{"./pkg/service/interfaces.go", "./pkg/service/mocks/mocks.go", "mocks"},
		{"./pkg/repository/interfaces.go", "./pkg/repository/mocks/mocks.go", "mocks"},
	}

	for _, i := range interfaces {
		if err := sh.Run("mockgen",
			"-source", i.source,
			"-destination", i.dest,
			"-package", i.pkg,
		); err != nil {
			return err
		}
	}

	utils.Info("✅ Mocks generated!")
	return nil
}

// Swagger generates Swagger documentation
func (Generate) Swagger() error {
	utils.Info("📚 Generating Swagger documentation...")

	return sh.Run("swag", "init",
		"-g", "./cmd/api/main.go",
		"-o", "./docs",
		"--parseDependency",
		"--parseInternal")
}

// Proto generates code from proto files
func (Generate) Proto() error {
	utils.Info("📡 Generating protobuf code...")

	return sh.Run("buf", "generate")
}

// All runs all code generation tasks
func (Generate) All() error {
	mg.Deps(
		Generate{}.Mocks,
		Generate{}.Swagger,
		Generate{}.Proto,
	)
	return nil
}

// Setup installs all dependencies and prepares the development environment
func Setup() error {
	utils.Info("🔧 Setting up development environment...")

	// Install Go dependencies
	var d Deps
	mg.Deps(d.Download)

	// Install tools
	var t Tools
	mg.Deps(t.Install)

	// Install custom tools
	customTools := []string{
		"github.com/cosmtrek/air@latest",
		"github.com/golang-migrate/migrate/v4/cmd/migrate@latest",
		"go.uber.org/mock/mockgen@latest",
		"github.com/swaggo/swag/cmd/swag@latest",
	}

	for _, tool := range customTools {
		fmt.Printf("Installing %s...\n", tool)
		if err := sh.Run("go", "install", tool); err != nil {
			return err
		}
	}

	// Create necessary directories
	dirs := []string{
		"bin",
		"tmp",
		"logs",
		"migrations",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o750); err != nil { // #nosec G301 -- example directory permissions
			return err
		}
	}

	// Copy example env file if needed
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		if err := sh.Run("cp", ".env.example", ".env"); err != nil {
			utils.Info("⚠️  No .env.example file found")
		}
	}

	utils.Info("✅ Setup completed! Run 'mage dev' to start development server")
	return nil
}

// CI runs the complete CI pipeline
func CI() error {
	utils.Info("🏗️  Running CI pipeline...")

	start := time.Now()

	// Set CI environment
	if err := os.Setenv("CI", "true"); err != nil {
		return fmt.Errorf("failed to set CI env var: %w", err)
	}

	// Run all checks in order
	var l Lint
	var t Test
	var b Build
	mg.SerialDeps(
		l.Default,
		t.CoverRace,
		b.All,
		Generate{}.All,
		Docker{}.Build,
	)

	duration := time.Since(start)
	fmt.Printf("✅ CI pipeline completed in %s\n", duration.Round(time.Second))

	return nil
}

// Clean removes all generated files and artifacts
func Clean() error {
	utils.Info("🧹 Cleaning up...")

	// Use MAGE-X clean
	var b Build
	mg.Deps(b.Clean)

	return cleanupArtifacts()
}

// cleanupArtifacts performs additional cleanup of build artifacts
func cleanupArtifacts() error {
	// Clean additional project files
	toRemove := []string{
		"tmp/",
		"logs/",
		"*.log",
		"coverage.*",
		"docs/swagger.json",
		"docs/swagger.yaml",
		"**/*_mock.go",
		"dist/",
	}

	for _, pattern := range toRemove {
		fmt.Printf("Removing %s\n", pattern)
		// Ignore errors for cleanup - files may not exist
		if err := sh.Run("rm", "-rf", pattern); err != nil {
			fmt.Printf("Warning: failed to remove %s: %v\n", pattern, err)
		}
	}

	_, _ = os.Stdout.WriteString("✅ Cleanup completed!\n") //nolint:errcheck // OK to ignore stdout write errors
	return nil
}
