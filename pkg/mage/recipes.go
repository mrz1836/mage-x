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
	errRecipeShowRequired         = errors.New("recipe parameter is required. Usage: magex recipes:show recipe=<name>")
	errRecipeRunRequired          = errors.New("recipe parameter is required. Usage: magex recipes:run recipe=<name>")
	errTermSearchRequired         = errors.New("term parameter is required. Usage: magex recipes:search term=<search>")
	errRecipeCreateRequired       = errors.New("recipe parameter is required. Usage: magex recipes:create recipe=<name>")
	errSourceInstallRequired      = errors.New("source parameter is required. Usage: magex recipes:install source=<url|file>")
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

	utils.Info("🎯 Recipe Categories:")

	for category, categoryRecipes := range categories {
		fmt.Printf("\n%s:\n", strings.ToUpper(category[:1])+category[1:])
		for i := range categoryRecipes {
			recipe := &categoryRecipes[i]
			fmt.Printf("  %-20s - %s\n", recipe.Name, recipe.Description)
		}
	}

	utils.Info("Usage:")
	utils.Info("  magex recipes:show <name>     # Show recipe details")
	utils.Info("  magex recipes:run <name>      # Run a recipe")
	utils.Info("  magex recipes:create <name>   # Create custom recipe")
	utils.Info("  magex recipes:search <term>   # Search recipes")

	return nil
}

// Show displays details of all available recipes
func (Recipes) Show() error {
	utils.Header("📖 All Recipe Details")

	recipes := getBuiltinRecipes()

	if len(recipes) == 0 {
		utils.Info("No recipes available")
		return nil
	}

	// Display each recipe in detail
	for i := range recipes {
		recipe := &recipes[i]

		if i > 0 {
			utils.Info("")
		}

		fmt.Printf("📋 %s\n", recipe.Name)
		fmt.Printf("   Category: %s\n", recipe.Category)
		fmt.Printf("   Description: %s\n", recipe.Description)

		if len(recipe.Tags) > 0 {
			fmt.Printf("   Tags: %s\n", strings.Join(recipe.Tags, ", "))
		}

		if len(recipe.Dependencies) > 0 {
			fmt.Printf("   Dependencies: %s\n", strings.Join(recipe.Dependencies, ", "))
		}

		fmt.Printf("   Steps: %d\n", len(recipe.Steps))
	}

	utils.Info("\nUse 'magex recipes:show recipe=<name>' to see detailed steps for a specific recipe")

	return nil
}

// ShowWithArgs displays details of a specific recipe (use recipe=<name>)
func (Recipes) ShowWithArgs(args ...string) error {
	// Parse command-line parameters
	params := utils.ParseParams(args)

	recipeName := utils.GetParam(params, "recipe", "")
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

	utils.Info("Steps:")
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
		utils.Info("Variables:")
		for key, value := range recipe.Variables {
			fmt.Printf("  %s = %s\n", key, value)
		}
	}

	return nil
}

// Run shows available recipes and prompts for selection
func (Recipes) Run() error {
	utils.Header("🚀 Recipe Execution Menu")

	recipes := getBuiltinRecipes()

	if len(recipes) == 0 {
		utils.Info("No recipes available")
		return nil
	}

	// Display recipes with numbers
	utils.Info("Available recipes:")
	for i, recipe := range recipes {
		fmt.Printf("  %d. %-20s - %s\n", i+1, recipe.Name, recipe.Description)
	}

	// Prompt for selection
	utils.Info("\nEnter recipe number or name:")
	choice, err := utils.PromptForInput("Recipe choice")
	if err != nil {
		return fmt.Errorf("failed to get recipe choice: %w", err)
	}

	if choice == "" {
		utils.Info("No recipe selected")
		return nil
	}

	var selectedRecipe *Recipe

	// Try to parse as number first
	if num := parseInt(choice); num > 0 && num <= len(recipes) {
		selectedRecipe = &recipes[num-1]
	} else {
		// Try to find by name
		for i := range recipes {
			recipe := &recipes[i]
			if strings.EqualFold(recipe.Name, choice) || strings.Contains(strings.ToLower(recipe.Name), strings.ToLower(choice)) {
				selectedRecipe = recipe
				break
			}
		}
	}

	if selectedRecipe == nil {
		return fmt.Errorf("%w: %s", errRecipeNotFound, choice)
	}

	utils.Info("🎯 Selected recipe: %s", selectedRecipe.Name)
	utils.Info("📝 Description: %s", selectedRecipe.Description)

	// Confirm execution
	confirm, err := utils.PromptForInput("Execute this recipe? (y/N)")
	if err != nil {
		return fmt.Errorf("failed to get confirmation: %w", err)
	}

	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		utils.Info("Recipe execution canceled")
		return nil
	}

	// Execute the selected recipe
	return executeRecipe(selectedRecipe)
}

// RunWithArgs executes a recipe (use recipe=<name>)
func (Recipes) RunWithArgs(args ...string) error {
	// Parse command-line parameters
	params := utils.ParseParams(args)

	recipeName := utils.GetParam(params, "recipe", "")
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
		utils.Info("📋 Step %d: %s", i+1, step.Name)

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

	utils.Success("🎉 Recipe '%s' completed successfully!", recipe.Name)
	return nil
}

// promptToRunRecipe prompts user to run one of the found recipes
func promptToRunRecipe(matches []Recipe) error {
	runChoice, err := utils.PromptForInput("Run one of these recipes? (y/N)")
	if err != nil {
		return err
	}
	if !isYesResponse(runChoice) {
		return nil
	}

	if len(matches) == 1 {
		return executeRecipe(&matches[0])
	}

	return promptAndSelectRecipe(matches)
}

// isYesResponse checks if the response is a yes
func isYesResponse(response string) bool {
	lower := strings.ToLower(response)
	return lower == "y" || lower == "yes"
}

// promptAndSelectRecipe prompts user to select a recipe from matches
func promptAndSelectRecipe(matches []Recipe) error {
	utils.Info("\nEnter recipe name to run:")
	recipeName, err := utils.PromptForInput("Recipe name")
	if err != nil {
		return err
	}
	if recipeName == "" {
		return nil
	}

	for i := range matches {
		if strings.EqualFold(matches[i].Name, recipeName) {
			return executeRecipe(&matches[i])
		}
	}

	utils.Warn("Recipe '%s' not found in the results", recipeName)
	return nil
}

// Search launches interactive recipe search
func (Recipes) Search() error {
	utils.Header("🔍 Interactive Recipe Search")

	// Prompt for search term
	searchTerm, err := utils.PromptForInput("Enter search term")
	if err != nil {
		return fmt.Errorf("failed to get search term: %w", err)
	}

	if searchTerm == "" {
		utils.Info("No search term entered")
		return nil
	}

	utils.Info("Searching for recipes matching '%s'...", searchTerm)

	recipes := getBuiltinRecipes()
	searchTermLower := strings.ToLower(searchTerm)

	var matches []Recipe
	for i := range recipes {
		recipe := &recipes[i]
		if strings.Contains(strings.ToLower(recipe.Name), searchTermLower) ||
			strings.Contains(strings.ToLower(recipe.Description), searchTermLower) ||
			containsTag(recipe.Tags, searchTermLower) {
			matches = append(matches, *recipe)
		}
	}

	if len(matches) == 0 {
		utils.Info("No recipes found matching '%s'", searchTerm)
		utils.Info("\nTry searching for:")
		utils.Info("  - Project types: 'cli', 'web', 'library'")
		utils.Info("  - Technologies: 'docker', 'git', 'ci'")
		utils.Info("  - Actions: 'setup', 'build', 'deploy'")
		return nil
	}

	utils.Success("Found %d recipes matching '%s':", len(matches), searchTerm)

	for i := range matches {
		recipe := &matches[i]
		fmt.Printf("\n📦 %s (%s)\n", recipe.Name, recipe.Category)
		fmt.Printf("   %s\n", recipe.Description)
		if len(recipe.Tags) > 0 {
			fmt.Printf("   Tags: %s\n", strings.Join(recipe.Tags, ", "))
		}
	}

	// Ask if user wants to run any of the found recipes
	if len(matches) > 0 {
		return promptToRunRecipe(matches)
	}

	return nil
}

// SearchWithArgs searches for recipes by keyword (use term=<search>)
func (Recipes) SearchWithArgs(args ...string) error {
	// Parse command-line parameters
	params := utils.ParseParams(args)

	searchTerm := utils.GetParam(params, "term", "")
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

// Create launches interactive recipe creation wizard
func (Recipes) Create() error {
	utils.Header("📝 Interactive Recipe Creation Wizard")

	// Prompt for recipe name
	recipeName, err := utils.PromptForInput("Enter recipe name")
	if err != nil {
		return fmt.Errorf("failed to get recipe name: %w", err)
	}

	if recipeName == "" {
		return errValueEmpty
	}

	// Prompt for description
	description, err := utils.PromptForInput("Enter recipe description")
	if err != nil {
		return fmt.Errorf("failed to get description: %w", err)
	}
	if description == "" {
		description = fmt.Sprintf("Custom recipe: %s", recipeName)
	}

	// Prompt for category
	utils.Info("Available categories:")
	utils.Info("  1. setup - Project setup and initialization")
	utils.Info("  2. build - Build and compilation tasks")
	utils.Info("  3. test - Testing and quality assurance")
	utils.Info("  4. deploy - Deployment and release tasks")
	utils.Info("  5. custom - Custom category")

	categoryChoice, err := utils.PromptForInput("Select category (1-5)")
	if err != nil {
		return fmt.Errorf("failed to get category: %w", err)
	}

	var category string
	switch categoryChoice {
	case "1":
		category = "setup"
	case "2":
		category = "build"
	case "3":
		category = "test"
	case "4":
		category = "deploy"
	case "5", "":
		category = "custom"
	default:
		category = "custom"
	}

	// Create recipe
	recipe := Recipe{
		Name:        recipeName,
		Description: description,
		Category:    category,
		Steps: []RecipeStep{
			{
				Name:        "Example Step",
				Description: "This is an example step - edit to customize",
				Command:     "echo",
				Args:        []string{"Hello from " + recipeName + " recipe!"},
			},
		},
		Variables: map[string]string{
			"PROJECT_NAME": "{{.ProjectName}}",
			"MODULE_PATH":  "{{.ModulePath}}",
		},
		Tags: []string{category, "custom"},
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

	utils.Success("✅ Created custom recipe: %s", recipeFile)
	utils.Info("📝 Category: %s", category)
	utils.Info("📁 Location: %s", recipeFile)
	utils.Info("✏️  Edit the file to customize your recipe steps and variables")

	return nil
}

// CreateWithArgs creates a new custom recipe (use recipe=<name>)
func (Recipes) CreateWithArgs(args ...string) error {
	// Parse command-line parameters
	params := utils.ParseParams(args)

	recipeName := utils.GetParam(params, "recipe", "")
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
	utils.Info("Run with: magex recipes:run recipe=%s", recipeName)

	return nil
}

// Install shows available recipe sources and installation options
func (Recipes) Install() error {
	utils.Header("📥 Recipe Installation Options")

	utils.Info("Available recipe sources:")
	utils.Info("  • Official MAGE-X Recipe Repository")
	utils.Info("  • GitHub repositories")
	utils.Info("  • Local recipe files")
	utils.Info("  • Community recipe collections")

	utils.Info("\nSupported source formats:")
	utils.Info("  • GitHub URLs: https://github.com/user/repo")
	utils.Info("  • Direct files: /path/to/recipe.yaml")
	utils.Info("  • Git repos: git@github.com:user/repo.git")

	utils.Info("\nExample usage:")
	utils.Info("  magex recipes:install source=https://github.com/mage-recipes/go-web")
	utils.Info("  magex recipes:install source=./my-recipe.yaml")
	utils.Info("  magex recipes:install source=https://raw.githubusercontent.com/user/repo/main/recipe.yaml")

	// Show currently available built-in recipes
	utils.Info("\nCurrently available built-in recipes:")
	recipes := getBuiltinRecipes()
	for _, recipe := range recipes {
		fmt.Printf("  %-20s - %s\n", recipe.Name, recipe.Description)
	}

	// Prompt for interactive installation
	installChoice, err := utils.PromptForInput("Install a recipe from source? (y/N)")
	if err == nil && (strings.ToLower(installChoice) == "y" || strings.ToLower(installChoice) == "yes") {
		source, err := utils.PromptForInput("Enter recipe source (URL or file path)")
		if err != nil {
			return fmt.Errorf("failed to get source: %w", err)
		}

		if source != "" {
			// Call the WithArgs version to handle the installation
			return (Recipes{}).InstallWithArgs("source=" + source)
		}
	}

	utils.Info("\nFor more recipe sources, visit: https://github.com/mage-x/recipes")

	return nil
}

// InstallWithArgs installs a recipe from a repository (use source=<url|file>)
func (Recipes) InstallWithArgs(args ...string) error {
	// Parse command-line parameters
	params := utils.ParseParams(args)

	source := utils.GetParam(params, "source", "")
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

// executeRecipe executes a complete recipe
func executeRecipe(recipe *Recipe) error {
	utils.Header("🚀 Running Recipe: " + recipe.Name)

	// Create context
	context := createRecipeContext(recipe)

	// Check dependencies
	if err := checkDependencies(recipe.Dependencies); err != nil {
		return fmt.Errorf("dependency check failed: %w", err)
	}

	// Execute steps
	for i, step := range recipe.Steps {
		utils.Info("📋 Step %d: %s", i+1, step.Name)

		if step.Description != "" {
			utils.Info("   %s", step.Description)
		}

		// Check condition if specified
		if step.Condition != "" && !evaluateRecipeCondition(step.Condition) {
			utils.Info("   ⏭️  Skipped (condition not met)")
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

	utils.Success("🎉 Recipe '%s' completed successfully!", recipe.Name)
	return nil
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

# See available commands (beautiful format)
mage help

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
