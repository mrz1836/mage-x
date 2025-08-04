package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// Static errors for module operations
var (
	errModuleNameNotFound   = errors.New("module name not found")
	errMultipleModuleErrors = errors.New("multiple module errors")
)

// ModuleInfo represents information about a Go module
type ModuleInfo struct {
	Path     string // Absolute path to the directory containing go.mod
	Module   string // Module name from go.mod
	Relative string // Relative path from root
	IsRoot   bool   // Whether this is the root module
	Name     string // Short name for display (last part of module path)
}

// GetPath returns the module path (for builder interface compatibility)
func (m *ModuleInfo) GetPath() string {
	return m.Path
}

// findAllModules discovers all go.mod files in the project, excluding vendor directories
func findAllModules() ([]ModuleInfo, error) {
	// Get the root directory
	root, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	var modules []ModuleInfo

	// Walk the directory tree looking for go.mod files
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor directories
		if info.IsDir() && info.Name() == "vendor" {
			return filepath.SkipDir
		}

		// Skip hidden directories (like .git)
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") && info.Name() != "." {
			// Allow .github directory as mentioned in requirements
			if info.Name() != ".github" {
				return filepath.SkipDir
			}
		}

		// Look for go.mod files
		if !info.IsDir() && info.Name() == "go.mod" {
			dir := filepath.Dir(path)
			relPath, err := filepath.Rel(root, dir)
			if err != nil {
				relPath = dir
			}

			// Read module name from go.mod
			moduleName, err := getModuleNameFromFile(path)
			if err != nil {
				utils.Warn("Failed to read module name from %s: %v", path, err)
				moduleName = relPath
			}

			// Extract the name (last part of module path)
			name := moduleName
			if idx := strings.LastIndex(moduleName, "/"); idx >= 0 {
				name = moduleName[idx+1:]
			}

			modules = append(modules, ModuleInfo{
				Path:     dir,
				Module:   moduleName,
				Relative: relPath,
				IsRoot:   relPath == ".",
				Name:     name,
			})
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk directory tree: %w", err)
	}

	// Sort modules by path for consistent output
	// Root module should come first
	sortModules(modules)

	return modules, nil
}

// getModuleNameFromFile reads the module name from a go.mod file
func getModuleNameFromFile(goModPath string) (string, error) {
	content, err := os.ReadFile(goModPath) // #nosec G304 -- go.mod path from controlled module discovery
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			moduleName := strings.TrimPrefix(line, "module ")
			moduleName = strings.TrimSpace(moduleName)
			// Remove any comments
			if idx := strings.Index(moduleName, "//"); idx >= 0 {
				moduleName = strings.TrimSpace(moduleName[:idx])
			}
			return moduleName, nil
		}
	}

	return "", fmt.Errorf("%w in %s", errModuleNameNotFound, goModPath)
}

// sortModules sorts modules with root module first, then by path
func sortModules(modules []ModuleInfo) {
	// Simple bubble sort - root module first, then alphabetical by path
	for i := 0; i < len(modules); i++ {
		for j := i + 1; j < len(modules); j++ {
			// Root module (.) should always be first
			if modules[j].Relative == "." {
				modules[i], modules[j] = modules[j], modules[i]
			} else if modules[i].Relative != "." && modules[i].Relative > modules[j].Relative {
				modules[i], modules[j] = modules[j], modules[i]
			}
		}
	}
}

// runCommandInModule runs a command in a specific module directory
func runCommandInModule(module ModuleInfo, command string, args ...string) error {
	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Change to module directory
	if err := os.Chdir(module.Path); err != nil {
		return fmt.Errorf("failed to change to directory %s: %w", module.Path, err)
	}

	// Ensure we change back to original directory
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			utils.Error("Failed to change back to original directory: %v", err)
		}
	}()

	// Run the command
	return GetRunner().RunCmd(command, args...)
}

// displayModuleHeader displays a header for the current module being processed
func displayModuleHeader(module ModuleInfo, operation string) {
	if module.Relative == "." {
		utils.Info("\n%s main module...", operation)
	} else {
		utils.Info("\n%s module in %s...", operation, module.Relative)
	}
}

// collectModuleErrors collects errors from multiple modules
type moduleError struct {
	Module ModuleInfo
	Error  error
}

func formatModuleErrors(errors []moduleError) error {
	if len(errors) == 0 {
		return nil
	}

	messages := make([]string, 0, len(errors))
	for _, me := range errors {
		location := "main module"
		if me.Module.Relative != "." {
			location = me.Module.Relative
		}
		messages = append(messages, fmt.Sprintf("  - %s: %v", location, me.Error))
	}

	return fmt.Errorf("%w: errors in %d module(s):\n%s", errMultipleModuleErrors, len(errors), strings.Join(messages, "\n"))
}
