// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors to satisfy err113 linter
var (
	errGoVetFailed        = errors.New("go vet failed")
	errStrictChecksFailed = errors.New("strict checks failed")
)

// ShadowToolManager manages shadow tool installation and ensures thread-safety
type ShadowToolManager interface {
	EnsureShadowTool() error
}

// shadowToolManager implements ShadowToolManager with thread-safe shadow tool installation
type shadowToolManager struct {
	mutex sync.Mutex
}

// NewShadowToolManager creates a new shadow tool manager
func NewShadowToolManager() ShadowToolManager {
	return &shadowToolManager{}
}

// GetDefaultShadowToolManager returns a default shadow tool manager instance
func GetDefaultShadowToolManager() ShadowToolManager {
	return NewShadowToolManager()
}

// Vet namespace for go vet tasks
type Vet mg.Namespace

// Default runs go vet on module packages only
func (Vet) Default() error {
	utils.Header("Running go vet")

	// Get module name
	module, err := utils.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %w", err)
	}

	utils.Info("Checking module: %s", module)

	// Get all packages in this module
	output, err := GetRunner().RunCmdOutput("go", "list", "./...")
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	packages := strings.Split(strings.TrimSpace(output), "\n")
	modulePackages := []string{}

	// Filter to only module packages
	for _, pkg := range packages {
		if strings.HasPrefix(pkg, module) {
			modulePackages = append(modulePackages, pkg)
		}
	}

	if len(modulePackages) == 0 {
		utils.Warn("No packages found in module")
		return nil
	}

	utils.Info("Vetting %d packages", len(modulePackages))

	// Build vet arguments
	args := []string{"vet"}

	// Add verbose flag if requested
	if verbose := utils.GetEnv("VERBOSE", ""); verbose == approvalTrue {
		args = append(args, "-v")
	}

	// Add build tags if specified
	if tags := utils.GetEnv("MAGE_X_BUILD_TAGS", ""); tags != "" {
		args = append(args, "-tags", tags)
	}

	// Add all module packages
	args = append(args, modulePackages...)

	start := time.Now()
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("go vet found issues: %w", err)
	}

	utils.Success("go vet passed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// All runs go vet on all packages (including dependencies)
func (Vet) All() error {
	utils.Header("Running go vet (All Packages)")

	args := []string{"vet"}

	// Add verbose flag if requested
	if verbose := utils.GetEnv("VERBOSE", ""); verbose == approvalTrue {
		args = append(args, "-v")
	}

	// Add build tags if specified
	if tags := utils.GetEnv("MAGE_X_BUILD_TAGS", ""); tags != "" {
		args = append(args, "-tags", tags)
	}

	args = append(args, "./...")

	start := time.Now()
	if err := GetRunner().RunCmd("go", args...); err != nil {
		return fmt.Errorf("go vet found issues: %w", err)
	}

	utils.Success("go vet passed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// Parallel runs go vet in parallel (faster for large repos)
func (Vet) Parallel() error {
	utils.Header("Running go vet in Parallel")

	// Get module name
	module, err := utils.GetModuleName()
	if err != nil {
		return fmt.Errorf("failed to get module name: %w", err)
	}

	// Get all packages in this module
	output, err := GetRunner().RunCmdOutput("go", "list", "./...")
	if err != nil {
		return fmt.Errorf("failed to list packages: %w", err)
	}

	packages := strings.Split(strings.TrimSpace(output), "\n")
	modulePackages := []string{}

	// Filter to only module packages
	for _, pkg := range packages {
		if strings.HasPrefix(pkg, module) {
			modulePackages = append(modulePackages, pkg)
		}
	}

	if len(modulePackages) == 0 {
		utils.Warn("No packages found in module")
		return nil
	}

	// Get CPU count for parallel execution
	parallel := getCPUCount()
	utils.Info("Vetting %d packages with %d workers", len(modulePackages), parallel)

	// Create error channel
	errChan := make(chan error, len(modulePackages))
	semaphore := make(chan struct{}, parallel)

	start := time.Now()

	// Vet packages in parallel
	for _, pkg := range modulePackages {
		go func(p string) {
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			args := []string{"vet"}

			// Add build tags if specified
			if tags := utils.GetEnv("MAGE_X_BUILD_TAGS", ""); tags != "" {
				args = append(args, "-tags", tags)
			}

			args = append(args, p)

			if err := GetRunner().RunCmd("go", args...); err != nil {
				errChan <- fmt.Errorf("vet failed for %s: %w", p, err)
			} else {
				errChan <- nil
			}
		}(pkg)
	}

	// Wait for all goroutines to complete
	var errors []error
	for i := 0; i < len(modulePackages); i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		utils.Error("go vet found issues in %d packages:", len(errors))
		for _, err := range errors {
			fmt.Printf("  - %v\n", err)
		}
		return errGoVetFailed
	}

	utils.Success("go vet passed in %s", utils.FormatDuration(time.Since(start)))
	return nil
}

// Shadow checks for shadowed variables
func (Vet) Shadow() error {
	utils.Header("Checking for Shadowed Variables")

	// Check if shadow tool exists, and install if needed
	if err := GetDefaultShadowToolManager().EnsureShadowTool(); err != nil {
		return fmt.Errorf("failed to ensure shadow tool is available: %w", err)
	}

	// Determine shadow command path
	shadowCmd := "shadow"
	if gopath, err := GetRunner().RunCmdOutput("go", "env", "GOPATH"); err == nil {
		shadowPath := strings.TrimSpace(gopath) + "/bin/shadow"
		if utils.FileExists(shadowPath) {
			shadowCmd = shadowPath
		}
	}

	// Build shadow command args - exclude test files to reduce noise
	args := []string{"-test=false", "./..."}

	// Add build tags if specified
	if tags := utils.GetEnv("MAGE_X_BUILD_TAGS", ""); tags != "" {
		args = append([]string{"-tags", tags}, args...)
	}

	// Run shadow directly
	utils.Info("Running shadow check...")
	if err := GetRunner().RunCmd(shadowCmd, args...); err != nil {
		return fmt.Errorf("shadow check found issues: %w", err)
	}

	utils.Success("No shadowed variables found")
	return nil
}

// EnsureShadowTool ensures the shadow tool is installed and available
func (stm *shadowToolManager) EnsureShadowTool() error {
	// Use mutex to prevent concurrent installations
	stm.mutex.Lock()
	defer stm.mutex.Unlock()

	// Try using go env GOPATH to check if shadow is in GOPATH/bin first
	if gopath, err := GetRunner().RunCmdOutput("go", "env", "GOPATH"); err == nil {
		shadowPath := strings.TrimSpace(gopath) + "/bin/shadow"
		if utils.FileExists(shadowPath) {
			return nil // Tool is available in GOPATH/bin
		}
	}

	// Check if shadow tool is already available by trying to run it
	if err := GetRunner().RunCmd("which", "shadow"); err == nil {
		return nil // Tool is already available
	}

	// Install shadow tool
	utils.Info("Installing shadow analyzer...")
	return GetRunner().RunCmd("go", "install", "golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest")
}

// Strict runs go vet with additional strict checks
func (Vet) Strict() error {
	utils.Header("Running Strict go vet")

	// Run multiple vet-like tools
	checks := []struct {
		name string
		fn   func() error
	}{
		{"Standard vet", Vet{}.Default},
		{"Shadow check", Vet{}.Shadow},
		{"Staticcheck", runStaticcheck},
		{"Ineffassign", runIneffassign},
		{"Misspell", runMisspell},
	}

	failed := 0
	for _, check := range checks {
		utils.Info("\nRunning %s...", check.name)
		if err := check.fn(); err != nil {
			utils.Error("Failed: %v", err)
			failed++
		}
	}

	if failed > 0 {
		return fmt.Errorf("%w: %d", errStrictChecksFailed, failed)
	}

	utils.Success("\nAll strict checks passed!")
	return nil
}

// Helper functions

// runStaticcheck runs the staticcheck tool
func runStaticcheck() error {
	staticcheckVersion := getLinterVersion("staticcheck")

	// Check if staticcheck is installed
	if !utils.CommandExists("staticcheck") {
		utils.Info("Installing staticcheck...")
		if err := GetRunner().RunCmd("go", "install", "honnef.co/go/tools/cmd/staticcheck@latest"); err != nil {
			return fmt.Errorf("failed to install staticcheck: %w", err)
		}
		staticcheckVersion = getLinterVersion("staticcheck")
	}

	// Run staticcheck
	utils.Info("Running staticcheck %s...", staticcheckVersion)
	if err := GetRunner().RunCmd("staticcheck", "./..."); err != nil {
		return fmt.Errorf("staticcheck found issues: %w", err)
	}

	utils.Success("staticcheck %s passed", staticcheckVersion)
	return nil
}

// runIneffassign checks for ineffectual assignments
func runIneffassign() error {
	ineffassignVersion := getLinterVersion("ineffassign")

	// Check if ineffassign is installed
	if !utils.CommandExists("ineffassign") {
		utils.Info("Installing ineffassign...")
		if err := GetRunner().RunCmd("go", "install", "github.com/gordonklaus/ineffassign@latest"); err != nil {
			return fmt.Errorf("failed to install ineffassign: %w", err)
		}
		ineffassignVersion = getLinterVersion("ineffassign")
	}

	// Run ineffassign
	utils.Info("Running ineffassign %s...", ineffassignVersion)
	if err := GetRunner().RunCmd("ineffassign", "./..."); err != nil {
		return fmt.Errorf("ineffassign found issues: %w", err)
	}

	utils.Success("ineffassign %s passed", ineffassignVersion)
	return nil
}

// runMisspell checks for misspelled words
func runMisspell() error {
	misspellVersion := getLinterVersion("misspell")

	// Check if misspell is installed
	if !utils.CommandExists("misspell") {
		utils.Info("Installing misspell...")
		if err := GetRunner().RunCmd("go", "install", "github.com/client9/misspell/cmd/misspell@latest"); err != nil {
			return fmt.Errorf("failed to install misspell: %w", err)
		}
		misspellVersion = getLinterVersion("misspell")
	}

	// Run misspell
	utils.Info("Running misspell %s...", misspellVersion)
	if err := GetRunner().RunCmd("misspell", "-w", "."); err != nil {
		return fmt.Errorf("misspell found issues: %w", err)
	}

	utils.Success("misspell %s passed", misspellVersion)
	return nil
}
