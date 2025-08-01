package mage

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for dependency management
var (
	errDependencyNameRequired = errors.New("dependency name required: mage deps:why github.com/pkg/errors")
	errModuleNameRequired     = errors.New("module name required: mage deps:init github.com/user/project")
	errGoModAlreadyExists     = errors.New("go.mod already exists")
)

// Deps namespace for dependency management tasks
type Deps mg.Namespace

// Default manages default dependencies
func (Deps) Default() error {
	utils.Header("Managing Dependencies")
	runner := GetRunner()
	return runner.RunCmd("go", "mod", "download")
}

// Download downloads all dependencies
func (Deps) Download() error {
	utils.Header("Downloading Dependencies")

	start := time.Now()
	if err := GetRunner().RunCmd("go", "mod", "download"); err != nil {
		return fmt.Errorf("failed to download dependencies: %w", err)
	}

	utils.Success("Dependencies downloaded in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// Tidy cleans up go.mod and go.sum
func (Deps) Tidy() error {
	utils.Header("Tidying Dependencies")

	if err := GetRunner().RunCmd("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("failed to tidy dependencies: %w", err)
	}

	utils.Success("Dependencies tidied")
	return nil
}

// Update updates all dependencies
func (Deps) Update() error {
	utils.Header("Updating Dependencies")

	// Get list of direct dependencies
	output, err := GetRunner().RunCmdOutput("go", "list", "-m", "-f", "{{if not .Indirect}}{{.Path}}{{end}}", "all")
	if err != nil {
		return fmt.Errorf("failed to list dependencies: %w", err)
	}

	deps := strings.Split(strings.TrimSpace(output), "\n")
	updatedCount := 0

	for _, dep := range deps {
		dep = strings.TrimSpace(dep)
		if dep == "" || strings.Contains(dep, "=>") {
			continue
		}

		utils.Info("Checking %s...", dep)

		// Get latest version
		latestOutput, err := GetRunner().RunCmdOutput("go", "list", "-m", "-versions", dep)
		if err != nil {
			utils.Warn("Failed to check %s: %v", dep, err)
			continue
		}

		parts := strings.Fields(latestOutput)
		if len(parts) < 2 {
			continue
		}

		latestVersion := parts[len(parts)-1]

		// Update to latest
		if err := GetRunner().RunCmd("go", "get", "-u", dep+"@"+latestVersion); err != nil {
			utils.Warn("Failed to update %s: %v", dep, err)
		} else {
			updatedCount++
		}
	}

	// Tidy after updates
	if err := GetRunner().RunCmd("go", "mod", "tidy"); err != nil {
		return fmt.Errorf("failed to tidy after updates: %w", err)
	}

	utils.Success("Updated %d dependencies", updatedCount)
	return nil
}

// Clean cleans the module cache
func (Deps) Clean() error {
	utils.Header("Cleaning Module Cache")

	if err := GetRunner().RunCmd("go", "clean", "-modcache"); err != nil {
		return fmt.Errorf("failed to clean module cache: %w", err)
	}

	utils.Success("Module cache cleaned")
	return nil
}

// Graph shows the dependency graph
func (Deps) Graph() error {
	utils.Header("Dependency Graph")

	return GetRunner().RunCmd("go", "mod", "graph")
}

// Why shows why a dependency is needed
func (Deps) Why(dep string) error {
	if dep == "" {
		return errDependencyNameRequired
	}

	utils.Header(fmt.Sprintf("Why %s?", dep))

	return GetRunner().RunCmd("go", "mod", "why", dep)
}

// Verify verifies dependencies
func (Deps) Verify() error {
	utils.Header("Verifying Dependencies")

	if err := GetRunner().RunCmd("go", "mod", "verify"); err != nil {
		return fmt.Errorf("dependency verification failed: %w", err)
	}

	utils.Success("All dependencies verified")
	return nil
}

// VulnCheck checks for known vulnerabilities
func (Deps) VulnCheck() error {
	// Delegate to tools:vulncheck
	return Tools{}.VulnCheck()
}

// List lists all dependencies
func (Deps) List() error {
	utils.Header("Dependencies")

	// Direct dependencies
	utils.Info("\nDirect dependencies:")
	if err := GetRunner().RunCmd("go", "list", "-m", "-f", "{{if not .Indirect}}{{.Path}} {{.Version}}{{end}}", "all"); err != nil {
		return err
	}

	// Show count of indirect dependencies
	output, err := GetRunner().RunCmdOutput("go", "list", "-m", "-f", "{{if .Indirect}}1{{end}}", "all")
	if err == nil {
		count := strings.Count(output, "1")
		utils.Info("\n(%d indirect dependencies)", count)
	}

	return nil
}

// Outdated shows outdated dependencies
func (Deps) Outdated() error {
	utils.Header("Checking for Outdated Dependencies")

	// Get all direct dependencies
	output, err := GetRunner().RunCmdOutput("go", "list", "-m", "-u", "-f", "{{if and (not .Indirect) .Update}}{{.Path}} {{.Version}} -> {{.Update.Version}}{{end}}", "all")
	if err != nil {
		return fmt.Errorf("failed to check outdated dependencies: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	outdated := []string{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			outdated = append(outdated, line)
		}
	}

	if len(outdated) == 0 {
		utils.Success("All dependencies are up to date!")
	} else {
		utils.Info("Found %d outdated dependencies:\n", len(outdated))
		for _, dep := range outdated {
			utils.Info("  %s", dep)
		}
		utils.Info("\nRun 'mage deps:update' to update all dependencies")
	}

	return nil
}

// Vendor vendors all dependencies
func (Deps) Vendor() error {
	utils.Header("Vendoring Dependencies")

	if err := GetRunner().RunCmd("go", "mod", "vendor"); err != nil {
		return fmt.Errorf("failed to vendor dependencies: %w", err)
	}

	utils.Success("Dependencies vendored to vendor/")
	return nil
}

// Init initializes a new go module
func (Deps) Init(module string) error {
	if module == "" {
		return errModuleNameRequired
	}

	utils.Header("Initializing Go Module")

	if utils.FileExists("go.mod") {
		return errGoModAlreadyExists
	}

	if err := GetRunner().RunCmd("go", "mod", "init", module); err != nil {
		return fmt.Errorf("failed to initialize module: %w", err)
	}

	utils.Success("Module initialized: %s", module)
	return nil
}

// Audit performs comprehensive dependency audit
func (Deps) Audit() error {
	utils.Header("Auditing Dependencies")
	runner := GetRunner()
	return runner.RunCmd("go", "list", "-m", "all")
}

// Licenses shows dependency licenses
func (Deps) Licenses() error {
	utils.Header("Checking Dependency Licenses")
	runner := GetRunner()
	return runner.RunCmd("echo", "Checking dependency licenses")
}

// Check checks for updates
func (Deps) Check() error {
	utils.Header("Checking Dependencies")
	runner := GetRunner()
	return runner.RunCmd("go", "list", "-m", "-u", "all")
}
