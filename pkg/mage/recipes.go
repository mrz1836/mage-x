// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/common/fileops"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for recipe operations
var (
	errRecipeShowRequired         = errors.New("RECIPE environment variable is required. Usage: RECIPE=<name> mage recipes:show")
	errRecipeRunRequired          = errors.New("RECIPE environment variable is required. Usage: RECIPE=<name> mage recipes:run")
	errTermSearchRequired         = errors.New("TERM environment variable is required. Usage: TERM=<search> mage recipes:search")
	errRecipeCreateRequired       = errors.New("RECIPE environment variable is required. Usage: RECIPE=<name> mage recipes:create")
	errSourceInstallRequired      = errors.New("SOURCE environment variable is required. Usage: SOURCE=<url|file> mage recipes:install")
	errRecipeNotFound             = errors.New("recipe not found")
	errCustomRecipeFileNotFound   = errors.New("custom recipe file not found")
	errRequiredDependencyNotFound = errors.New("required dependency not found")
)

// Recipes namespace for common development patterns
type Recipes mg.Namespace

// Recipe represents a reusable development pattern
type Recipe struct {
	Name         string
	Description  string
	Category     string
	Dependencies []string
	Steps        []RecipeStep
	Templates    map[string]string
	Variables    map[string]string
	Tags         []string
}

// RecipeStep represents a single step in a recipe
type RecipeStep struct {
	Name        string
	Description string
	Command     string
	Args        []string
	Env         map[string]string
	Optional    bool
	Condition   string
}

// RecipeContext contains context for recipe execution
type RecipeContext struct {
	ProjectName string
	ModulePath  string
	Version     string
	GoVersion   string
	Variables   map[string]string
}

// Default lists available recipes
func (Recipes) Default() error {
	return (Recipes{}).List()
}

// List shows all available recipes
func (Recipes) List() error {
	utils.Header("📚 Available MAGE-X Recipes")

	recipes := getBuiltinRecipes()

	// Group by category
	categories := make(map[string][]Recipe)
	for i := range recipes {
		recipe := &recipes[i]
		categories[recipe.Category] = append(categories[recipe.Category], *recipe)
	}

	utils.Info("\n🎯 Recipe Categories:")

	for category, categoryRecipes := range categories {
		fmt.Printf("\n%s:\n", strings.ToUpper(category[:1])+category[1:])
		for i := range categoryRecipes {
			recipe := &categoryRecipes[i]
			fmt.Printf("  %-20s - %s\n", recipe.Name, recipe.Description)
		}
	}

	utils.Info("\nUsage:")
	utils.Info("  mage recipes:show <name>     # Show recipe details")
	utils.Info("  mage recipes:run <name>      # Run a recipe")
	utils.Info("  mage recipes:create <name>   # Create custom recipe")
	utils.Info("  mage recipes:search <term>   # Search recipes")

	return nil
}

// Show displays details of a specific recipe
func (Recipes) Show() error {
	recipeName := utils.GetEnv("RECIPE", "")
	if recipeName == "" {
		return errRecipeShowRequired
	}

	recipe, err := getRecipe(recipeName)
	if err != nil {
		return err
	}

	utils.Header("📖 Recipe: " + recipe.Name)

	fmt.Printf("Description: %s\n", recipe.Description)
	fmt.Printf("Category: %s\n", recipe.Category)
	fmt.Printf("Tags: %s\n", strings.Join(recipe.Tags, ", "))

	if len(recipe.Dependencies) > 0 {
		fmt.Printf("Dependencies: %s\n", strings.Join(recipe.Dependencies, ", "))
	}

	utils.Info("\nSteps:")
	for i, step := range recipe.Steps {
		fmt.Printf("  %d. %s\n", i+1, step.Name)
		if step.Description != "" {
			fmt.Printf("     %s\n", step.Description)
		}
		if step.Optional {
			fmt.Printf("     (Optional)\n")
		}
	}

	if len(recipe.Variables) > 0 {
		utils.Info("\nVariables:")
		for key, value := range recipe.Variables {
			fmt.Printf("  %s = %s\n", key, value)
		}
	}

	return nil
}

// Run executes a recipe
func (Recipes) Run() error {
	recipeName := utils.GetEnv("RECIPE", "")
	if recipeName == "" {
		return errRecipeRunRequired
	}

	recipe, err := getRecipe(recipeName)
	if err != nil {
		return err
	}

	utils.Header("🚀 Running Recipe: " + recipe.Name)

	// Create context
	context := createRecipeContext(&recipe)

	// Check dependencies
	if err := checkDependencies(recipe.Dependencies); err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	// Execute steps
	for i, step := range recipe.Steps {
		utils.Info("\n📋 Step %d: %s", i+1, step.Name)

		if step.Description != "" {
			utils.Info("   %s", step.Description)
		}

		// Check condition
		if step.Condition != "" && !evaluateRecipeCondition(step.Condition) {
			utils.Info("   Skipping (condition not met)")
			continue
		}

		// Execute step
		if err := executeRecipeStep(&step, context); err != nil {
			if step.Optional {
				utils.Warn("   Optional step failed: %v", err)
				continue
			}
			return fmt.Errorf("step %d failed: %w", i+1, err)
		}

		utils.Success("   ✅ Completed")
	}

	utils.Success("\n🎉 Recipe '%s' completed successfully!", recipe.Name)
	return nil
}

// Search searches for recipes by name or description
func (Recipes) Search() error {
	searchTerm := utils.GetEnv("TERM", "")
	if searchTerm == "" {
		return errTermSearchRequired
	}

	utils.Header("🔍 Recipe Search Results")

	recipes := getBuiltinRecipes()
	searchTerm = strings.ToLower(searchTerm)

	var matches []Recipe
	for i := range recipes {
		recipe := &recipes[i]
		if strings.Contains(strings.ToLower(recipe.Name), searchTerm) ||
			strings.Contains(strings.ToLower(recipe.Description), searchTerm) ||
			containsTag(recipe.Tags, searchTerm) {
			matches = append(matches, *recipe)
		}
	}

	if len(matches) == 0 {
		utils.Info("No recipes found matching '%s'", searchTerm)
		return nil
	}

	utils.Info("Found %d recipes matching '%s':", len(matches), searchTerm)

	for i := range matches {
		recipe := &matches[i]
		fmt.Printf("\n📦 %s (%s)\n", recipe.Name, recipe.Category)
		fmt.Printf("   %s\n", recipe.Description)
		if len(recipe.Tags) > 0 {
			fmt.Printf("   Tags: %s\n", strings.Join(recipe.Tags, ", "))
		}
	}

	return nil
}

// Create creates a new custom recipe
func (Recipes) Create() error {
	recipeName := utils.GetEnv("RECIPE", "")
	if recipeName == "" {
		return errRecipeCreateRequired
	}

	utils.Header("📝 Creating Custom Recipe: " + recipeName)

	// Create recipe template
	recipe := Recipe{
		Name:        recipeName,
		Description: fmt.Sprintf("Custom recipe: %s", recipeName),
		Category:    "custom",
		Steps: []RecipeStep{
			{
				Name:        "Example Step",
				Description: "This is an example step",
				Command:     "echo",
				Args:        []string{"Hello from recipe!"},
			},
		},
		Variables: map[string]string{
			"EXAMPLE_VAR": "example_value",
		},
		Tags: []string{"custom", "example"},
	}

	// Create recipes directory
	recipesDir := ".mage/recipes"
	if err := os.MkdirAll(recipesDir, 0o750); err != nil {
		return fmt.Errorf("failed to create recipes directory: %w", err)
	}

	// Save recipe as YAML
	recipeFile := filepath.Join(recipesDir, recipeName+".yaml")
	if err := saveRecipeAsYAML(&recipe, recipeFile); err != nil {
		return fmt.Errorf("failed to save recipe: %w", err)
	}

	utils.Success("Created custom recipe: %s", recipeFile)
	utils.Info("Edit the file to customize your recipe")
	utils.Info("Run with: RECIPE=%s mage recipes:run", recipeName)

	return nil
}

// Install installs a recipe from a URL or file
func (Recipes) Install() error {
	source := utils.GetEnv("SOURCE", "")
	if source == "" {
		return errSourceInstallRequired
	}

	utils.Header("📥 Installing Recipe")

	utils.Info("Installing recipe from: %s", source)

	// For now, just show what would be done
	// In a real implementation, this would download and install recipes

	utils.Info("Recipe installation is not yet implemented")
	utils.Info("This would download and install a recipe from the specified source")

	return nil
}

// Builtin recipes

// getBuiltinRecipes returns all built-in recipes
func getBuiltinRecipes() []Recipe {
	return []Recipe{
		// Setup recipes
		{
			Name:        "fresh-start",
			Description: "Set up a fresh Go project with best practices",
			Category:    "setup",
			Tags:        []string{"setup", "init", "best-practices"},
			Steps: []RecipeStep{
				{Name: "Initialize Go module", Command: "go", Args: []string{"mod", "init", "{{.ModulePath}}"}},
				{Name: "Create project structure", Command: "mkdir", Args: []string{"-p", "cmd", "pkg", "internal", "test"}},
				{Name: "Add .gitignore", Command: "create-gitignore"},
				{Name: "Add README.md", Command: "create-readme"},
				{Name: "Initialize git repository", Command: "git", Args: []string{"init"}},
				{Name: "Add mage build files", Command: "add-mage-files"},
			},
		},

		{
			Name:         "cli-app",
			Description:  "Create a CLI application with Cobra",
			Category:     "setup",
			Tags:         []string{"cli", "cobra", "application"},
			Dependencies: []string{"github.com/spf13/cobra"},
			Steps: []RecipeStep{
				{Name: "Install Cobra", Command: "go", Args: []string{"get", "github.com/spf13/cobra/cobra"}},
				{Name: "Initialize Cobra app", Command: "cobra", Args: []string{"init", "--pkg-name", "{{.ModulePath}}"}},
				{Name: "Add version command", Command: "cobra", Args: []string{"add", "version"}},
				{Name: "Configure build", Command: "setup-cli-build"},
			},
		},

		{
			Name:         "web-api",
			Description:  "Create a REST API with Gin and database",
			Category:     "setup",
			Tags:         []string{"web", "api", "gin", "database"},
			Dependencies: []string{"github.com/gin-gonic/gin", "gorm.io/gorm"},
			Steps: []RecipeStep{
				{Name: "Install Gin", Command: "go", Args: []string{"get", "github.com/gin-gonic/gin"}},
				{Name: "Install GORM", Command: "go", Args: []string{"get", "gorm.io/gorm"}},
				{Name: "Create API structure", Command: "create-api-structure"},
				{Name: "Add database models", Command: "create-db-models"},
				{Name: "Add handlers", Command: "create-handlers"},
				{Name: "Add middleware", Command: "create-middleware"},
				{Name: "Add Docker files", Command: "create-docker-files"},
			},
		},

		// Development recipes
		{
			Name:        "ci-setup",
			Description: "Set up CI/CD pipeline with GitHub Actions",
			Category:    "development",
			Tags:        []string{"ci", "github", "actions", "pipeline"},
			Steps: []RecipeStep{
				{Name: "Create .github directory", Command: "mkdir", Args: []string{"-p", ".github/workflows"}},
				{Name: "Add CI workflow", Command: "create-ci-workflow"},
				{Name: "Add release workflow", Command: "create-release-workflow"},
				{Name: "Add dependabot config", Command: "create-dependabot-config"},
			},
		},

		{
			Name:        "docker-setup",
			Description: "Set up Docker and Docker Compose",
			Category:    "development",
			Tags:        []string{"docker", "compose", "containerization"},
			Steps: []RecipeStep{
				{Name: "Create Dockerfile", Command: "create-dockerfile"},
				{Name: "Create docker-compose.yml", Command: "create-docker-compose"},
				{Name: "Add .dockerignore", Command: "create-dockerignore"},
				{Name: "Add Docker scripts", Command: "create-docker-scripts"},
			},
		},

		{
			Name:        "testing-setup",
			Description: "Set up comprehensive testing infrastructure",
			Category:    "development",
			Tags:        []string{"testing", "coverage", "benchmarks"},
			Steps: []RecipeStep{
				{Name: "Create test directories", Command: "mkdir", Args: []string{"-p", "test/unit", "test/integration"}},
				{Name: "Add test helpers", Command: "create-test-helpers"},
				{Name: "Add benchmark tests", Command: "create-benchmark-tests"},
				{Name: "Configure coverage", Command: "setup-coverage"},
				{Name: "Add testing scripts", Command: "create-testing-scripts"},
			},
		},

		// Quality recipes
		{
			Name:        "quality-gates",
			Description: "Set up code quality gates and standards",
			Category:    "quality",
			Tags:        []string{"quality", "linting", "formatting"},
			Steps: []RecipeStep{
				{Name: "Add .golangci.json", Command: "create-golangci-config"},
				{Name: "Add pre-commit hooks", Command: "create-pre-commit-hooks"},
				{Name: "Configure formatting", Command: "setup-formatting"},
				{Name: "Add quality scripts", Command: "create-quality-scripts"},
			},
		},

		{
			Name:        "security-scan",
			Description: "Set up security scanning and vulnerability checks",
			Category:    "quality",
			Tags:        []string{"security", "vulnerability", "scanning"},
			Steps: []RecipeStep{
				{Name: "Install govulncheck", Command: "go", Args: []string{"install", "golang.org/x/vuln/cmd/govulncheck@latest"}},
				{Name: "Run vulnerability scan", Command: "govulncheck", Args: []string{"./..."}},
				{Name: "Add security workflow", Command: "create-security-workflow"},
				{Name: "Configure security scanning", Command: "setup-security-scanning"},
			},
		},

		// Release recipes
		{
			Name:        "release-prep",
			Description: "Prepare project for release",
			Category:    "release",
			Tags:        []string{"release", "preparation", "versioning"},
			Steps: []RecipeStep{
				{Name: "Update version", Command: "update-version"},
				{Name: "Generate changelog", Command: "generate-changelog"},
				{Name: "Update documentation", Command: "update-docs"},
				{Name: "Run full test suite", Command: "run-full-tests"},
				{Name: "Build release artifacts", Command: "build-release-artifacts"},
			},
		},

		{
			Name:        "goreleaser-setup",
			Description: "Set up GoReleaser for automated releases",
			Category:    "release",
			Tags:        []string{"goreleaser", "release", "automation"},
			Steps: []RecipeStep{
				{Name: "Install GoReleaser", Command: "install-goreleaser"},
				{Name: "Initialize .goreleaser.yml", Command: "goreleaser", Args: []string{"init"}},
				{Name: "Configure release settings", Command: "configure-goreleaser"},
				{Name: "Add release workflow", Command: "create-release-workflow"},
			},
		},

		// Maintenance recipes
		{
			Name:        "dependency-update",
			Description: "Update all dependencies safely",
			Category:    "maintenance",
			Tags:        []string{"dependencies", "update", "maintenance"},
			Steps: []RecipeStep{
				{Name: "Check for updates", Command: "go", Args: []string{"list", "-u", "-m", "all"}},
				{Name: "Update dependencies", Command: "go", Args: []string{"get", "-u", "./..."}},
				{Name: "Tidy modules", Command: "go", Args: []string{"mod", "tidy"}},
				{Name: "Run tests", Command: "go", Args: []string{"test", "./..."}},
				{Name: "Check for vulnerabilities", Command: "govulncheck", Args: []string{"./..."}},
			},
		},

		{
			Name:        "cleanup",
			Description: "Clean up project files and caches",
			Category:    "maintenance",
			Tags:        []string{"cleanup", "cache", "maintenance"},
			Steps: []RecipeStep{
				{Name: "Clean build cache", Command: "go", Args: []string{"clean", "-cache"}},
				{Name: "Clean module cache", Command: "go", Args: []string{"clean", "-modcache"}, Optional: true},
				{Name: "Remove build artifacts", Command: "rm", Args: []string{"-rf", "dist", "bin"}},
				{Name: "Clean temporary files", Command: "find", Args: []string{".", "-name", "*.tmp", "-delete"}},
			},
		},
	}
}

// Helper functions

// getRecipe gets a recipe by name
func getRecipe(name string) (Recipe, error) {
	recipes := getBuiltinRecipes()

	for i := range recipes {
		recipe := &recipes[i]
		if recipe.Name == name {
			return *recipe, nil
		}
	}

	// Try to load from custom recipes
	customRecipe, err := loadCustomRecipe(name)
	if err == nil {
		return customRecipe, nil
	}

	return Recipe{}, fmt.Errorf("%w: %s", errRecipeNotFound, name)
}

// loadCustomRecipe loads a custom recipe from file
func loadCustomRecipe(name string) (Recipe, error) {
	recipeFile := filepath.Join(".mage", "recipes", name+".yaml")

	if !utils.FileExists(recipeFile) {
		return Recipe{}, fmt.Errorf("%w: %s", errCustomRecipeFileNotFound, recipeFile)
	}

	// In a real implementation, this would parse YAML
	// For now, return a placeholder
	return Recipe{
		Name:        name,
		Description: "Custom recipe",
		Category:    "custom",
		Steps:       []RecipeStep{},
	}, nil
}

// createRecipeContext creates execution context for a recipe
func createRecipeContext(recipe *Recipe) *RecipeContext {
	context := &RecipeContext{
		Variables: make(map[string]string),
	}

	// Get project information
	if module, err := getModuleName(); err == nil {
		context.ModulePath = module
		parts := strings.Split(module, "/")
		if len(parts) > 0 {
			context.ProjectName = parts[len(parts)-1]
		}
	}

	context.Version = getVersion()
	if goVersion, err := utils.GetGoVersion(); err == nil {
		context.GoVersion = goVersion
	}

	// Add recipe variables
	for key, value := range recipe.Variables {
		context.Variables[key] = value
	}

	// Add environment variables
	for key, value := range recipe.Variables {
		if envValue := os.Getenv(key); envValue != "" {
			context.Variables[key] = envValue
		} else {
			context.Variables[key] = value
		}
	}

	return context
}

// checkDependencies checks if required dependencies are available
func checkDependencies(deps []string) error {
	for _, dep := range deps {
		if !isDependencyAvailable(dep) {
			return fmt.Errorf("%w: %s", errRequiredDependencyNotFound, dep)
		}
	}
	return nil
}

// isDependencyAvailable checks if a dependency is available
func isDependencyAvailable(dep string) bool {
	// Check if it's a command
	if utils.CommandExists(dep) {
		return true
	}

	// Check if it's a Go module
	if strings.Contains(dep, ".") {
		// This would check if the module is in go.mod
		return true // Simplified for now
	}

	return false
}

// evaluateRecipeCondition evaluates a recipe condition
func evaluateRecipeCondition(condition string) bool {
	// Simple condition evaluation
	switch condition {
	case "git_available":
		return utils.CommandExists("git")
	case "docker_available":
		return utils.CommandExists("docker")
	case "is_module":
		return utils.FileExists("go.mod")
	case "has_tests":
		return utils.FileExists("*_test.go")
	default:
		return true
	}
}

// executeRecipeStep executes a single recipe step
func executeRecipeStep(step *RecipeStep, context *RecipeContext) error {
	// Expand templates in command and args
	command := expandTemplate(step.Command, context)

	args := make([]string, 0, len(step.Args))
	for _, arg := range step.Args {
		args = append(args, expandTemplate(arg, context))
	}

	// Set environment variables
	for key, value := range step.Env {
		if err := os.Setenv(key, expandTemplate(value, context)); err != nil {
			return fmt.Errorf("failed to set environment variable %s: %w", key, err)
		}
	}

	// Handle special commands
	switch command {
	case "create-gitignore":
		return createGitignoreForRecipe()
	case "create-readme":
		return createReadmeForRecipe(context)
	case "add-mage-files":
		return addMageFilesForRecipe()
	case "setup-cli-build":
		return setupCLIBuild()
	case "create-api-structure":
		return createAPIStructure()
	case "create-dockerfile":
		return createDockerfile(context)
	case "create-ci-workflow":
		return createCIWorkflow()
	default:
		// Execute as regular command
		return GetRunner().RunCmd(command, args...)
	}
}

// expandTemplate expands template variables
func expandTemplate(text string, context *RecipeContext) string {
	tmpl, err := template.New("recipe").Parse(text)
	if err != nil {
		return text
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, context); err != nil {
		return text
	}

	return result.String()
}

// containsTag checks if tags contain a search term
func containsTag(tags []string, term string) bool {
	for _, tag := range tags {
		if strings.Contains(strings.ToLower(tag), term) {
			return true
		}
	}
	return false
}

// saveRecipeAsYAML saves a recipe as YAML (simplified)
func saveRecipeAsYAML(recipe *Recipe, filename string) error {
	// In a real implementation, this would use yaml.Marshal
	content := fmt.Sprintf(`name: %s
description: %s
category: %s
tags: %s
variables:
  EXAMPLE_VAR: example_value
steps:
  - name: Example Step
    description: This is an example step
    command: echo
    args: ["Hello from recipe!"]
`, recipe.Name, recipe.Description, recipe.Category, strings.Join(recipe.Tags, ", "))

	fileOps := fileops.New()
	return fileOps.File.WriteFile(filename, []byte(content), 0o644)
}

// Special command implementations

func setupCLIBuild() error {
	utils.Info("Setting up CLI build configuration...")
	return nil
}

func createAPIStructure() error {
	dirs := []string{
		"api/handlers",
		"api/middleware",
		"api/routes",
		"internal/models",
		"internal/services",
		"internal/database",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}

	return nil
}

func createDockerfile(context *RecipeContext) error {
	content := fmt.Sprintf(`FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o %s .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/%s .
CMD ["./%s"]
`, context.ProjectName, context.ProjectName, context.ProjectName)

	return os.WriteFile("Dockerfile", []byte(content), 0o600)
}

func createCIWorkflow() error {
	content := `name: CI

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

	return os.WriteFile(".github/workflows/ci.yml", []byte(content), 0o600)
}

// Recipe-specific implementations that don't conflict with init.go functions

func createGitignoreForRecipe() error {
	content := `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary
*.test

# Output of the go coverage tool
*.out

# Go workspace file
go.work

# Build artifacts
dist/
bin/

# IDE files
.vscode/
.idea/
*.swp
*.swo

# OS files
.DS_Store
Thumbs.db
`
	return os.WriteFile(".gitignore", []byte(content), 0o600)
}

func createReadmeForRecipe(context *RecipeContext) error {
	content := fmt.Sprintf(`# %s

A Go project built with MAGE-X.

## Installation

`+"```bash"+`
go install %s@latest
`+"```"+`

## Usage

`+"```bash"+`
%s --help
`+"```"+`

## Development

This project uses [MAGE-X](https://github.com/mrz1836/mage-x) for build automation.

`+"```bash"+`
# Install mage
go install github.com/magefile/mage@latest

# See available commands
mage -l

# Build and test
mage build
mage test
`+"```"+`

## License

MIT License
`, context.ProjectName, context.ModulePath, context.ProjectName)

	return os.WriteFile("README.md", []byte(content), 0o600)
}

func addMageFilesForRecipe() error {
	content := `//go:build mage
// +build mage

package main

import (
	"github.com/mrz1836/mage-x/pkg/mage"
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
	return os.WriteFile("magefile.go", []byte(content), 0o600)
}
