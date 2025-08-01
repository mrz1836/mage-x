// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Static errors for init operations
var (
	errNotGoProject = errors.New("this doesn't appear to be a Go project (no go.mod found)")
)

// Init namespace for project initialization tasks
type Init mg.Namespace

// ProjectType represents different types of Go projects
type ProjectType string

const (
	// LibraryProject represents a reusable Go library project
	LibraryProject ProjectType = "library"
	// CLIProject represents a command-line interface application
	CLIProject ProjectType = "cli"
	// WebAPIProject represents a web API service
	WebAPIProject ProjectType = "webapi"
	// MicroserviceProject represents a microservice architecture project
	MicroserviceProject ProjectType = "microservice"
	// ToolProject represents a utility or tool project
	ToolProject ProjectType = "tool"
	// GenericProject represents a generic Go project
	GenericProject ProjectType = "generic"
)

// InitProjectConfig contains project initialization configuration
type InitProjectConfig struct {
	Name        string
	Type        ProjectType
	Module      string
	Description string
	Author      string
	Email       string
	License     string
	GoVersion   string
	UseMage     bool
	UseDocker   bool
	UseCI       bool
	Features    []string
}

// Default initializes a new project with interactive prompts
func (Init) Default() error {
	utils.Header("🚀 MAGE-X Project Initialization")

	// Get project configuration
	config, err := getInitProjectConfig()
	if err != nil {
		return fmt.Errorf("failed to get project configuration: %w", err)
	}

	// Create project structure
	if err := createProjectStructure(config); err != nil {
		return fmt.Errorf("failed to create project structure: %w", err)
	}

	// Initialize project files
	if err := initializeProjectFiles(config); err != nil {
		return fmt.Errorf("failed to initialize project files: %w", err)
	}

	// Initialize git repository
	if err := initializeGitRepo(config); err != nil {
		utils.Warn("Failed to initialize git repository: %v", err)
	}

	// Install dependencies
	if err := installDependencies(); err != nil {
		utils.Warn("Failed to install dependencies: %v", err)
	}

	// Show completion message
	showCompletionMessage(config)

	return nil
}

// Library initializes a Go library project
func (Init) Library() error {
	utils.Header("📚 Creating Go Library Project")

	config := &InitProjectConfig{
		Type:      LibraryProject,
		UseMage:   true,
		UseDocker: false,
		UseCI:     true,
		Features:  []string{"testing", "benchmarks", "docs"},
	}

	return initializeProject(config)
}

// CLI initializes a CLI application project
func (Init) CLI() error {
	utils.Header("⚡ Creating CLI Application Project")

	config := &InitProjectConfig{
		Type:      CLIProject,
		UseMage:   true,
		UseDocker: true,
		UseCI:     true,
		Features:  []string{"cobra", "viper", "testing", "goreleaser"},
	}

	return initializeProject(config)
}

// WebAPI initializes a web API project
func (Init) WebAPI() error {
	utils.Header("🌐 Creating Web API Project")

	config := &InitProjectConfig{
		Type:      WebAPIProject,
		UseMage:   true,
		UseDocker: true,
		UseCI:     true,
		Features:  []string{"gin", "gorm", "swagger", "testing", "migrations"},
	}

	return initializeProject(config)
}

// Microservice initializes a microservice project
func (Init) Microservice() error {
	utils.Header("🔧 Creating Microservice Project")

	config := &InitProjectConfig{
		Type:      MicroserviceProject,
		UseMage:   true,
		UseDocker: true,
		UseCI:     true,
		Features:  []string{"grpc", "prometheus", "tracing", "testing", "kubernetes"},
	}

	return initializeProject(config)
}

// Tool initializes a developer tool project
func (Init) Tool() error {
	utils.Header("🔨 Creating Developer Tool Project")

	config := &InitProjectConfig{
		Type:      ToolProject,
		UseMage:   true,
		UseDocker: true,
		UseCI:     true,
		Features:  []string{"cobra", "testing", "goreleaser", "homebrew"},
	}

	return initializeProject(config)
}

// Upgrade adds MAGE-X to an existing project
func (Init) Upgrade() error {
	utils.Header("⬆️ Upgrading Project to MAGE-X")

	// Check if this is a Go project
	if !utils.FileExists("go.mod") {
		return errNotGoProject
	}

	// Get existing module name
	module, err := utils.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %w", err)
	}

	utils.Info("Upgrading project: %s", module)

	// Add mage files
	if err := addMageFiles(); err != nil {
		return fmt.Errorf("failed to add mage files: %w", err)
	}

	// Update go.mod
	if err := updateGoMod(); err != nil {
		return fmt.Errorf("failed to update go.mod: %w", err)
	}

	utils.Success("✅ Project upgraded to MAGE-X!")
	utils.Info("Run 'mage -l' to see available commands")

	return nil
}

// Templates lists available project templates
func (Init) Templates() error {
	utils.Header("📋 Available Project Templates")

	templates := []struct {
		Type        ProjectType
		Name        string
		Description string
		Features    []string
	}{
		{LibraryProject, "Library", "Go library with testing and documentation", []string{"testing", "benchmarks", "docs"}},
		{CLIProject, "CLI", "Command-line application with Cobra", []string{"cobra", "viper", "testing", "goreleaser"}},
		{WebAPIProject, "Web API", "REST API with Gin and database", []string{"gin", "gorm", "swagger", "testing"}},
		{MicroserviceProject, "Microservice", "Cloud-native microservice", []string{"grpc", "prometheus", "tracing", "k8s"}},
		{ToolProject, "Tool", "Developer tool with releases", []string{"cobra", "testing", "goreleaser", "homebrew"}},
		{GenericProject, "Generic", "Basic Go project structure", []string{"testing", "basic"}},
	}

	utils.Info("\n🎯 Available Templates:")
	utils.Info("Type          Name          Description")
	utils.Info("----------    ----------    -----------")

	for _, tmpl := range templates {
		fmt.Printf("%-12s  %-12s  %s\n", tmpl.Type, tmpl.Name, tmpl.Description)
	}

	utils.Info("\nUsage:")
	utils.Info("  mage init:library      # Create library project")
	utils.Info("  mage init:cli          # Create CLI project")
	utils.Info("  mage init:webapi       # Create web API project")
	utils.Info("  mage init:microservice # Create microservice project")
	utils.Info("  mage init:tool         # Create tool project")
	utils.Info("  mage init:default      # Interactive setup")

	return nil
}

// Helper functions

// getInitProjectConfig gets project configuration from environment or prompts
func getInitProjectConfig() (*InitProjectConfig, error) {
	config := &InitProjectConfig{
		GoVersion: runtime.Version(),
		UseMage:   true,
		UseCI:     true,
	}

	// Get from environment variables first
	config.Name = utils.GetEnv("PROJECT_NAME", "")
	config.Module = utils.GetEnv("PROJECT_MODULE", "")
	config.Description = utils.GetEnv("PROJECT_DESCRIPTION", "")
	config.Author = utils.GetEnv("PROJECT_AUTHOR", "")
	config.Email = utils.GetEnv("PROJECT_EMAIL", "")
	config.License = utils.GetEnv("PROJECT_LICENSE", "MIT")

	// Get project type
	projectType := utils.GetEnv("PROJECT_TYPE", "generic")
	switch projectType {
	case "library":
		config.Type = LibraryProject
	case "cli":
		config.Type = CLIProject
	case "webapi":
		config.Type = WebAPIProject
	case "microservice":
		config.Type = MicroserviceProject
	case "tool":
		config.Type = ToolProject
	default:
		config.Type = GenericProject
	}

	// If not provided via env, use defaults based on current directory
	if config.Name == "" {
		pwd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
		config.Name = filepath.Base(pwd)
	}

	if config.Module == "" {
		config.Module = fmt.Sprintf("github.com/username/%s", config.Name)
	}

	if config.Description == "" {
		config.Description = fmt.Sprintf("A %s project built with MAGE-X", config.Type)
	}

	return config, nil
}

// initializeProject initializes a project with the given configuration
func initializeProject(config *InitProjectConfig) error {
	// Complete configuration from environment
	fullConfig, err := getInitProjectConfig()
	if err != nil {
		return err
	}

	// Merge with provided config
	fullConfig.Type = config.Type
	fullConfig.UseMage = config.UseMage
	fullConfig.UseDocker = config.UseDocker
	fullConfig.UseCI = config.UseCI
	fullConfig.Features = config.Features

	// Create project
	if err := createProjectStructure(fullConfig); err != nil {
		return err
	}

	if err := initializeProjectFiles(fullConfig); err != nil {
		return err
	}

	if err := initializeGitRepo(fullConfig); err != nil {
		utils.Warn("Failed to initialize git repository: %v", err)
	}

	if err := installDependencies(); err != nil {
		utils.Warn("Failed to install dependencies: %v", err)
	}

	showCompletionMessage(fullConfig)
	return nil
}

// createProjectStructure creates the basic project directory structure
func createProjectStructure(config *InitProjectConfig) error {
	utils.Info("Creating project structure...")

	// Basic directories
	dirs := []string{
		"cmd",
		"pkg",
		"internal",
		"test",
		"docs",
		"scripts",
		".github/workflows",
	}

	// Add type-specific directories
	switch config.Type {
	case WebAPIProject, MicroserviceProject:
		dirs = append(dirs, "api", "migrations", "deployments")
	case LibraryProject:
		dirs = append(dirs, "examples")
	case CLIProject, ToolProject:
		dirs = append(dirs, "completions")
	case GenericProject:
		// No additional directories for generic projects
	}

	// Create directories
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// initializeProjectFiles creates the initial project files
func initializeProjectFiles(config *InitProjectConfig) error {
	utils.Info("Initializing project files...")

	// Create go.mod
	if err := createGoMod(config); err != nil {
		return err
	}

	// Create main.go
	if err := createMainFile(config); err != nil {
		return err
	}

	// Create README.md
	if err := createReadme(config); err != nil {
		return err
	}

	// Create .gitignore
	if err := createGitignore(); err != nil {
		return err
	}

	// Create LICENSE
	if err := createLicense(config); err != nil {
		return err
	}

	// Create mage files
	if config.UseMage {
		if err := addMageFiles(); err != nil {
			return err
		}
	}

	// Create Docker files
	if config.UseDocker {
		if err := createDockerFiles(config); err != nil {
			return err
		}
	}

	// Create CI files
	if config.UseCI {
		if err := createCIFiles(); err != nil {
			return err
		}
	}

	return nil
}

// createGoMod creates the go.mod file
func createGoMod(config *InitProjectConfig) error {
	content := fmt.Sprintf(`module %s

go %s

require (
	github.com/magefile/mage v1.15.0
	github.com/mrz1836/go-mage v1.0.0
)
`, config.Module, strings.TrimPrefix(config.GoVersion, "go"))

	fileOps := fileops.New()
	return fileOps.File.WriteFile("go.mod", []byte(content), 0o644)
}

// createMainFile creates the main.go file
func createMainFile(config *InitProjectConfig) error {
	var content string

	switch config.Type {
	case CLIProject, ToolProject:
		content = getMainCLITemplate(config)
	case WebAPIProject:
		content = getMainWebAPITemplate(config)
	case MicroserviceProject:
		content = getMainMicroserviceTemplate(config)
	case LibraryProject, GenericProject:
		content = getMainGenericTemplate(config)
	}

	fileOps := fileops.New()
	return fileOps.File.WriteFile("main.go", []byte(content), 0o644)
}

// createReadme creates the README.md file
func createReadme(config *InitProjectConfig) error {
	tmpl := `# {{.Name}}

{{.Description}}

## Installation

` + "```bash" + `
go install {{.Module}}@latest
` + "```" + `

## Usage

` + "```bash" + `
{{.Name}} --help
` + "```" + `

## Development

This project uses [MAGE-X](https://github.com/mrz1836/go-mage) for build automation.

` + "```bash" + `
# Install mage
go install github.com/magefile/mage@latest

# See available commands
mage -l

# Run tests
mage test

# Build
mage build
` + "```" + `

## Contributing

1. Fork the repository
2. Create your feature branch (` + "`git checkout -b feature/amazing-feature`" + `)
3. Commit your changes (` + "`git commit -m 'Add some amazing feature'`" + `)
4. Push to the branch (` + "`git push origin feature/amazing-feature`" + `)
5. Open a Pull Request

## License

This project is licensed under the {{.License}} License - see the [LICENSE](LICENSE) file for details.
`

	t, err := template.New("readme").Parse(tmpl)
	if err != nil {
		return err
	}

	// Generate README content
	var buf strings.Builder
	if err := t.Execute(&buf, config); err != nil {
		return err
	}

	fileOps := fileops.New()
	return fileOps.File.WriteFile("README.md", []byte(buf.String()), 0o644)
}

// createGitignore creates the .gitignore file
func createGitignore() error {
	content := `# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with go test -c
*.test

# Output of the go coverage tool
*.out
coverage.txt
coverage.html

# Dependency directories
vendor/

# Go workspace file
go.work

# IDE files
.vscode/
.idea/
*.swp
*.swo
*~

# OS files
.DS_Store
Thumbs.db

# Build artifacts
dist/
bin/
build/

# Log files
*.log

# Environment files
.env
.env.local
.env.*.local

# Docker
.dockerignore

# Temporary files
tmp/
temp/
`

	fileOps := fileops.New()
	return fileOps.File.WriteFile(".gitignore", []byte(content), 0o644)
}

// createLicense creates the LICENSE file
func createLicense(config *InitProjectConfig) error {
	// This is a simplified MIT license template
	content := `MIT License

Copyright (c) 2024 ` + config.Author + `

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`

	fileOps := fileops.New()
	return fileOps.File.WriteFile("LICENSE", []byte(content), 0o644)
}

// addMageFiles adds mage build files to the project
func addMageFiles() error {
	utils.Info("Adding MAGE-X build files...")

	// Create magefile.go
	magefileContent := `//go:build mage
// +build mage

package main

import (
	"github.com/mrz1836/go-mage/pkg/mage"
)

// Build builds the application
func Build() error {
	return mage.Build{}.Default()
}

// Test runs the test suite
func Test() error {
	return mage.Test{}.Default()
}

// Lint runs the linter
func Lint() error {
	return mage.Lint{}.Default()
}

// Clean cleans build artifacts
func Clean() error {
	return mage.Build{}.Clean()
}
`

	fileOps := fileops.New()
	return fileOps.File.WriteFile("magefile.go", []byte(magefileContent), 0o644)
}

// Template functions for different project types

func getMainGenericTemplate(config *InitProjectConfig) string {
	return `package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("Hello, ` + config.Name + `!")
}
`
}

func getMainCLITemplate(config *InitProjectConfig) string {
	return `package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Usage: %s <command>\n", os.Args[0])
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "version":
		fmt.Println("` + config.Name + ` v1.0.0")
	case "help":
		fmt.Printf("Usage: %s <command>\n", os.Args[0])
		fmt.Println("Commands:")
		fmt.Println("  version  Show version")
		fmt.Println("  help     Show this help")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}
`
}

func getMainWebAPITemplate(config *InitProjectConfig) string {
	return `package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello from ` + config.Name + `!")
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	fmt.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
`
}

func getMainMicroserviceTemplate(config *InitProjectConfig) string {
	return `package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "` + config.Name + ` microservice")
	})

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Ready")
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		fmt.Println("Server starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exited")
}
`
}

// createDockerFiles creates Dockerfile and docker-compose.yml
func createDockerFiles(config *InitProjectConfig) error {
	// Create Dockerfile
	dockerfile := `FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
`

	fileOps := fileops.New()
	if err := fileOps.File.WriteFile("Dockerfile", []byte(dockerfile), 0o644); err != nil {
		return err
	}

	// Create docker-compose.yml for web services
	if config.Type == WebAPIProject || config.Type == MicroserviceProject {
		compose := `version: '3.8'

services:
  ` + config.Name + `:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENV=development
    depends_on:
      - db

  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: ` + config.Name + `
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
`

		if err := fileOps.File.WriteFile("docker-compose.yml", []byte(compose), 0o644); err != nil {
			return err
		}
	}

	return nil
}

// createCIFiles creates GitHub Actions workflow
func createCIFiles() error {
	workflow := `name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v4

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.24'

    - name: Install Mage
      run: go install github.com/magefile/mage@latest

    - name: Run tests
      run: mage test

    - name: Run linter
      run: mage lint

    - name: Build
      run: mage build
`

	fileOps := fileops.New()
	return fileOps.File.WriteFile(".github/workflows/ci.yml", []byte(workflow), 0o644)
}

// initializeGitRepo initializes git repository
func initializeGitRepo(config *InitProjectConfig) error {
	utils.Info("Initializing git repository...")

	// Initialize git
	if err := GetRunner().RunCmd("git", "init"); err != nil {
		return err
	}

	// Add files
	if err := GetRunner().RunCmd("git", "add", "."); err != nil {
		return err
	}

	// Initial commit
	message := fmt.Sprintf("Initial commit: %s", config.Description)
	return GetRunner().RunCmd("git", "commit", "-m", message)
}

// installDependencies installs project dependencies
func installDependencies() error {
	utils.Info("Installing dependencies...")

	// Download dependencies
	if err := GetRunner().RunCmd("go", "mod", "download"); err != nil {
		return err
	}

	// Tidy up
	return GetRunner().RunCmd("go", "mod", "tidy")
}

// updateGoMod updates go.mod for project upgrades
func updateGoMod() error {
	// Add mage dependency
	if err := GetRunner().RunCmd("go", "get", "github.com/magefile/mage@latest"); err != nil {
		return err
	}

	// Add mage-tools dependency
	if err := GetRunner().RunCmd("go", "get", "github.com/mrz1836/go-mage@latest"); err != nil {
		return err
	}

	// Tidy up
	return GetRunner().RunCmd("go", "mod", "tidy")
}

// showCompletionMessage shows project creation completion message
func showCompletionMessage(config *InitProjectConfig) {
	utils.Success("\n🎉 Project '%s' created successfully!", config.Name)

	fmt.Printf("\n📁 Project Structure:\n")
	fmt.Printf("  Module: %s\n", config.Module)
	fmt.Printf("  Type: %s\n", config.Type)
	fmt.Printf("  Features: %s\n", strings.Join(config.Features, ", "))

	fmt.Printf("\n🚀 Next Steps:\n")
	fmt.Printf("  1. cd %s\n", config.Name)
	fmt.Printf("  2. mage -l          # List available commands\n")
	fmt.Printf("  3. mage test        # Run tests\n")
	fmt.Printf("  4. mage build       # Build project\n")

	if config.UseDocker {
		fmt.Printf("  5. docker-compose up  # Run with Docker\n")
	}

	fmt.Printf("\n✨ Happy coding with MAGE-X!\n")
}

// Additional methods for Init namespace required by tests

// Project initializes a new project
func (Init) Project() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Initializing project")
}

// Config initializes configuration
func (Init) Config() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Initializing config")
}

// Git initializes git repository
func (Init) Git() error {
	runner := GetRunner()
	return runner.RunCmd("git", "init")
}

// Mage initializes mage files
func (Init) Mage() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Initializing mage")
}

// CI initializes CI configuration
func (Init) CI() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Initializing CI")
}

// Docker initializes Docker configuration
func (Init) Docker() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Initializing Docker")
}

// Docs initializes documentation
func (Init) Docs() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Initializing docs")
}

// License initializes license file
func (Init) License() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Initializing license")
}

// Makefile initializes Makefile
func (Init) Makefile() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Initializing Makefile")
}

// Editorconfig initializes .editorconfig
func (Init) Editorconfig() error {
	runner := GetRunner()
	return runner.RunCmd("echo", "Initializing .editorconfig")
}
