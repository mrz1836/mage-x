// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Mod namespace for Go module management tasks
type Mod mg.Namespace

// Download downloads all module dependencies
func (Mod) Download() error {
	utils.Header("Downloading Module Dependencies")

	// Show current module
	module, err := utils.GetModuleName()
	if err == nil {
		utils.Info("Module: %s", module)
	}

	// Download dependencies
	utils.Info("Downloading dependencies...")
	if err := GetRunner().RunCmd("go", "mod", "download"); err != nil {
		return fmt.Errorf("failed to download dependencies: %w", err)
	}

	// Verify downloads
	utils.Info("Verifying downloads...")
	if err := GetRunner().RunCmd("go", "mod", "verify"); err != nil {
		utils.Warn("Verification failed: %v", err)
	}

	utils.Success("Dependencies downloaded successfully")
	return nil
}

// Tidy cleans up go.mod and go.sum
func (Mod) Tidy() error {
	utils.Header("Tidying Module Dependencies")

	// Run go mod tidy
	utils.Info("Running go mod tidy...")
	if err := GetRunner().RunCmd("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	// Check if anything changed
	output, err := GetRunner().RunCmdOutput("git", "status", "--porcelain", "go.mod", "go.sum")
	if err == nil && strings.TrimSpace(output) != "" {
		utils.Info("Module files were updated:")
		fmt.Println(output)
	} else {
		utils.Success("Module files are already tidy")
	}

	return nil
}

// Update updates all dependencies to latest versions
func (Mod) Update() error {
	utils.Header("Updating Dependencies")

	// List outdated dependencies first
	utils.Info("Checking for updates...")
	output, err := GetRunner().RunCmdOutput("go", "list", "-u", "-m", "all")
	if err != nil {
		return fmt.Errorf("failed to list dependencies: %w", err)
	}

	// Count updates available
	updates := 0
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "[") && strings.Contains(line, "]") {
			updates++
			utils.Info("  %s", line)
		}
	}

	if updates == 0 {
		utils.Success("All dependencies are up to date")
		return nil
	}

	utils.Info("\nFound %d updates available", updates)

	// Update dependencies
	utils.Info("Updating dependencies...")
	if err := GetRunner().RunCmd("go", "get", "-u", "./..."); err != nil {
		return fmt.Errorf("failed to update dependencies: %w", err)
	}

	// Run tidy after update
	utils.Info("Running go mod tidy...")
	if err := GetRunner().RunCmd("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	utils.Success("Dependencies updated successfully")
	return nil
}

// Clean removes the module cache
func (Mod) Clean() error {
	utils.Header("Cleaning Module Cache")

	utils.Warn("This will remove all cached modules!")
	utils.Info("Module cache location: %s", getModCache())

	// Check if FORCE is set
	if os.Getenv("FORCE") != "true" {
		utils.Error("Set FORCE=true to confirm module cache deletion")
		return fmt.Errorf("operation canceled")
	}

	// Clean module cache
	utils.Info("Cleaning module cache...")
	if err := GetRunner().RunCmd("go", "clean", "-modcache"); err != nil {
		return fmt.Errorf("failed to clean module cache: %w", err)
	}

	utils.Success("Module cache cleaned")
	return nil
}

// Graph generates a dependency graph
func (Mod) Graph() error {
	utils.Header("Generating Dependency Graph")

	// Get module graph
	utils.Info("Analyzing dependencies...")
	output, err := GetRunner().RunCmdOutput("go", "mod", "graph")
	if err != nil {
		return fmt.Errorf("failed to generate dependency graph: %w", err)
	}

	// Parse and display summary
	deps := make(map[string][]string)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) == 2 {
			parent := parts[0]
			dep := parts[1]
			deps[parent] = append(deps[parent], dep)
		}
	}

	// Show direct dependencies
	module, _ := utils.GetModuleName()
	if directDeps, ok := deps[module]; ok {
		utils.Info("\nDirect dependencies (%d):", len(directDeps))
		for _, dep := range directDeps {
			fmt.Printf("  - %s\n", dep)
		}
	}

	// Save full graph if requested
	if graphFile := os.Getenv("GRAPH_FILE"); graphFile != "" {
		fileOps := fileops.New()
		if err := fileOps.File.WriteFile(graphFile, []byte(output), 0o644); err != nil {
			return fmt.Errorf("failed to write graph file: %w", err)
		}
		utils.Success("Full dependency graph saved to: %s", graphFile)
	}

	utils.Info("\nTotal modules in graph: %d", len(deps))
	return nil
}

// Why shows why a module is needed
func (Mod) Why() error {
	module := os.Getenv("MODULE")
	if module == "" {
		return fmt.Errorf("MODULE environment variable is required. Usage: MODULE=example.com/pkg mage mod:why")
	}

	utils.Header("Module Dependency Analysis")
	utils.Info("Analyzing why %s is needed...", module)

	// Run go mod why
	output, err := GetRunner().RunCmdOutput("go", "mod", "why", module)
	if err != nil {
		return fmt.Errorf("failed to analyze module: %w", err)
	}

	fmt.Println("\nDependency path:")
	fmt.Println(output)

	// Also check if it's a direct dependency
	directDeps, err := GetRunner().RunCmdOutput("go", "list", "-m", "-f", "{{.Require}}", "all")
	if err == nil && strings.Contains(directDeps, module) {
		utils.Info("This is a DIRECT dependency")
	} else {
		utils.Info("This is an INDIRECT dependency")
	}

	return nil
}

// Vendor vendors all dependencies
func (Mod) Vendor() error {
	utils.Header("Vendoring Dependencies")

	// Check if vendor directory exists
	vendorExists := utils.DirExists("vendor")

	// Run go mod vendor
	utils.Info("Vendoring dependencies...")
	if err := GetRunner().RunCmd("go", "mod", "vendor"); err != nil {
		return fmt.Errorf("vendoring failed: %w", err)
	}

	// Show vendor directory size
	if size, err := getDirSize("vendor"); err == nil {
		utils.Info("Vendor directory size: %s", formatBytes(size))
	}

	if !vendorExists {
		utils.Success("Dependencies vendored successfully")
		utils.Info("Remember to add vendor/ to your .gitignore")
	} else {
		utils.Success("Vendor directory updated")
	}

	return nil
}

// Init initializes a new Go module
func (Mod) Init() error {
	utils.Header("Initializing Go Module")

	// Check if go.mod already exists
	if utils.FileExists("go.mod") {
		return fmt.Errorf("go.mod already exists")
	}

	// Get module name
	moduleName := os.Getenv("MODULE")
	if moduleName == "" {
		// Try to infer from git remote
		if remote, err := GetRunner().RunCmdOutput("git", "remote", "get-url", "origin"); err == nil {
			remote = strings.TrimSpace(remote)
			// Convert git URL to module path
			moduleName = gitURLToModulePath(remote)
		}
	}

	if moduleName == "" {
		return fmt.Errorf("MODULE environment variable is required. Usage: MODULE=github.com/user/repo mage mod:init")
	}

	// Initialize module
	utils.Info("Initializing module: %s", moduleName)
	if err := GetRunner().RunCmd("go", "mod", "init", moduleName); err != nil {
		return fmt.Errorf("module initialization failed: %w", err)
	}

	utils.Success("Module initialized successfully")
	utils.Info("Created go.mod for %s", moduleName)

	return nil
}

// Helper functions

// getModCache returns the module cache directory
func getModCache() string {
	if cache := os.Getenv("GOMODCACHE"); cache != "" {
		return cache
	}

	// Default locations
	if gopath := os.Getenv("GOPATH"); gopath != "" {
		return filepath.Join(gopath, "pkg", "mod")
	}

	home, _ := os.UserHomeDir()
	return filepath.Join(home, "go", "pkg", "mod")
}

// getDirSize calculates directory size

// gitURLToModulePath converts a git URL to a Go module path
func gitURLToModulePath(gitURL string) string {
	// Remove protocol
	gitURL = strings.TrimPrefix(gitURL, "https://")
	gitURL = strings.TrimPrefix(gitURL, "http://")
	gitURL = strings.TrimPrefix(gitURL, "git@")
	gitURL = strings.TrimPrefix(gitURL, "ssh://git@")

	// Convert git@github.com:user/repo.git to github.com/user/repo
	gitURL = strings.Replace(gitURL, ":", "/", 1)

	// Remove .git suffix
	gitURL = strings.TrimSuffix(gitURL, ".git")

	return gitURL
}

// Additional methods for Mod namespace required by tests

// Verify verifies module dependencies
func (Mod) Verify() error {
	runner := GetRunner()
	return runner.RunCmd("go", "mod", "verify")
}

// Edit edits go.mod from tools or scripts
func (Mod) Edit(args ...string) error {
	runner := GetRunner()
	cmdArgs := append([]string{"mod", "edit"}, args...)
	return runner.RunCmd("go", cmdArgs...)
}

// Get adds dependencies
func (Mod) Get(packages ...string) error {
	runner := GetRunner()
	cmdArgs := append([]string{"get"}, packages...)
	return runner.RunCmd("go", cmdArgs...)
}

// List lists modules
func (Mod) List(pattern ...string) error {
	runner := GetRunner()
	cmdArgs := []string{"list", "-m"}
	if len(pattern) > 0 {
		cmdArgs = append(cmdArgs, pattern...)
	} else {
		cmdArgs = append(cmdArgs, "all")
	}
	return runner.RunCmd("go", cmdArgs...)
}
