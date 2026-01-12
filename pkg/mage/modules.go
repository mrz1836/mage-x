package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// chdirMu protects os.Chdir calls when DirRunner interface is not available.
// This is a fallback safety mechanism - prefer using DirRunner implementations.
//
//nolint:gochecknoglobals // Required for process-wide directory change synchronization
var chdirMu sync.Mutex

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
			return fmt.Errorf("failed to access path %s: %w", path, err)
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
			var relPath string
			relPath, err = filepath.Rel(root, dir)
			if err != nil {
				relPath = dir
			}

			// Read module name from go.mod
			var moduleName string
			moduleName, err = getModuleNameFromFile(path)
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
		return "", fmt.Errorf("failed to read go.mod file %s: %w", goModPath, err)
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

// sortModules sorts modules with root module first, then alphabetically by path
func sortModules(modules []ModuleInfo) {
	sort.Slice(modules, func(i, j int) bool {
		// Root module (.) should always be first
		if modules[i].Relative == "." {
			return true
		}
		if modules[j].Relative == "." {
			return false
		}
		// Otherwise sort alphabetically by relative path
		return modules[i].Relative < modules[j].Relative
	})
}

// runInModuleDir handles directory management for module commands using generics.
// It prefers DirRunner if available, otherwise falls back to os.Chdir with mutex protection.
// This consolidates the directory switching logic used by both runCommandInModuleWithRunner
// and runCommandInModuleOutputWithRunner.
func runInModuleDir[T any](module ModuleInfo, runner CommandRunner,
	dirFn func(dirRunner DirRunner, path string) (T, error),
	fallbackFn func() (T, error),
) (T, error) {
	// Prefer DirRunner interface - it's goroutine-safe
	if dirRunner, ok := runner.(DirRunner); ok {
		return dirFn(dirRunner, module.Path)
	}

	// Fallback: Use os.Chdir with mutex protection
	// WARNING: This is not ideal for high parallelism but provides safety for legacy runners
	chdirMu.Lock()
	defer chdirMu.Unlock()

	var zero T

	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		return zero, fmt.Errorf("failed to get current directory: %w", err)
	}

	// Change to module directory
	if err = os.Chdir(module.Path); err != nil {
		return zero, fmt.Errorf("failed to change to directory %s: %w", module.Path, err)
	}

	// Ensure we change back to original directory
	defer func() {
		if chErr := os.Chdir(originalDir); chErr != nil {
			utils.Error("Failed to change back to original directory: %v", chErr)
		}
	}()

	return fallbackFn()
}

// runCommandInModule runs a command in a specific module directory
func runCommandInModule(module ModuleInfo, command string, args ...string) error {
	return runCommandInModuleWithRunner(module, GetRunner(), command, args...)
}

// runCommandInModuleOutput runs a command in a specific module directory and captures output.
func runCommandInModuleOutput(module ModuleInfo, command string, args ...string) (string, error) {
	return runCommandInModuleOutputWithRunner(module, GetRunner(), command, args...)
}

// runCommandInModuleOutputWithRunner runs a command in a specific module directory and captures output.
// If the runner implements DirRunner, it uses the goroutine-safe RunCmdOutputInDir method.
// Otherwise, falls back to os.Chdir with mutex protection for safety.
func runCommandInModuleOutputWithRunner(module ModuleInfo, runner CommandRunner, command string, args ...string) (string, error) {
	return runInModuleDir(module, runner,
		func(dirRunner DirRunner, path string) (string, error) {
			return dirRunner.RunCmdOutputInDir(path, command, args...)
		},
		func() (string, error) {
			return runner.RunCmdOutput(command, args...)
		},
	)
}

// runCommandInModuleWithRunner runs a command in a specific module directory using the provided runner.
// If the runner implements DirRunner, it uses the goroutine-safe RunCmdInDir method.
// Otherwise, falls back to os.Chdir with mutex protection for safety.
func runCommandInModuleWithRunner(module ModuleInfo, runner CommandRunner, command string, args ...string) error {
	_, err := runInModuleDir(module, runner,
		func(dirRunner DirRunner, path string) (struct{}, error) {
			return struct{}{}, dirRunner.RunCmdInDir(path, command, args...)
		},
		func() (struct{}, error) {
			return struct{}{}, runner.RunCmd(command, args...)
		},
	)
	if err != nil {
		return fmt.Errorf("failed to run command '%s' in module %s: %w", command, module.Relative, err)
	}
	return nil
}

// displayModuleHeader displays a header for the current module being processed
func displayModuleHeader(module ModuleInfo, operation string) {
	if module.Relative == "." {
		utils.Info("\n%s main module...", operation)
	} else {
		utils.Info("\n%s module in %s...", operation, module.Relative)
	}
}

// ModuleCompletionOptions configures module completion messages.
type ModuleCompletionOptions struct {
	Module      ModuleInfo
	Operation   string // e.g., "Linting", "Coverage tests"
	Suffix      string // e.g., " [integration]" for build tags
	StartTime   time.Time
	Err         error
	SuccessVerb string // Custom success verb (default: "passed")
}

// displayModuleCompletion displays a completion message for a module operation.
// Uses "passed" for success and "failed" for error by default.
func displayModuleCompletion(module ModuleInfo, operation string, startTime time.Time, err error) {
	displayModuleCompletionWithOptions(ModuleCompletionOptions{
		Module:    module,
		Operation: operation,
		StartTime: startTime,
		Err:       err,
	})
}

// displayModuleCompletionWithSuffix handles cases with build tags or other suffixes.
func displayModuleCompletionWithSuffix(module ModuleInfo, operation, suffix string, startTime time.Time, err error) {
	displayModuleCompletionWithOptions(ModuleCompletionOptions{
		Module:    module,
		Operation: operation,
		Suffix:    suffix,
		StartTime: startTime,
		Err:       err,
	})
}

// displayModuleCompletionWithOptions provides full control over completion messages.
// Use this for edge cases like "All issues fixed" instead of "passed".
func displayModuleCompletionWithOptions(opts ModuleCompletionOptions) {
	duration := utils.FormatDuration(time.Since(opts.StartTime))
	location := opts.Module.Relative + opts.Suffix
	if opts.Err != nil {
		utils.Error("%s failed for %s in %s", opts.Operation, location, duration)
	} else {
		verb := opts.SuccessVerb
		if verb == "" {
			verb = "passed"
		}
		utils.Success("%s %s for %s in %s", opts.Operation, verb, location, duration)
	}
}

// OverallCompletionOptions configures overall completion messages.
type OverallCompletionOptions struct {
	Operation string // e.g., "linting", "coverage tests"
	Prefix    string // e.g., "Unit " for "All Unit tests passed"
	Suffix    string // e.g., " [integration]" for build tags
	Verb      string // "passed" or "completed"
	StartTime time.Time
}

// displayOverallCompletion displays overall completion for batch operations.
func displayOverallCompletion(operation, verb string, startTime time.Time) {
	displayOverallCompletionWithOptions(OverallCompletionOptions{
		Operation: operation,
		Verb:      verb,
		StartTime: startTime,
	})
}

// displayOverallCompletionWithOptions provides full control over overall completion messages.
// Handles patterns like "All Unit tests [integration] passed in X"
func displayOverallCompletionWithOptions(opts OverallCompletionOptions) {
	duration := utils.FormatDuration(time.Since(opts.StartTime))
	utils.Success("All %s%s%s %s in %s", opts.Prefix, opts.Operation, opts.Suffix, opts.Verb, duration)
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

// ModuleDependencies represents a module with its local dependencies
type ModuleDependencies struct {
	Module       ModuleInfo
	Dependencies []string // Module paths this module depends on
}

// parseModuleDependencies reads a go.mod file and extracts local replace directives
// to identify dependencies on other modules in the workspace
func parseModuleDependencies(module ModuleInfo, allModulePaths map[string]bool) ([]string, error) {
	goModPath := filepath.Join(module.Path, "go.mod")
	content, err := os.ReadFile(goModPath) // #nosec G304 -- go.mod path from controlled module discovery
	if err != nil {
		return nil, fmt.Errorf("failed to read go.mod: %w", err)
	}

	var dependencies []string
	lines := strings.Split(string(content), "\n")
	inReplaceBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Handle replace block
		if strings.HasPrefix(line, "replace (") || strings.HasPrefix(line, "replace(") {
			inReplaceBlock = true
			continue
		}
		if inReplaceBlock && line == ")" {
			inReplaceBlock = false
			continue
		}

		// Parse replace directives (both single-line and block format)
		var replaceLine string
		if strings.HasPrefix(line, "replace ") && !inReplaceBlock {
			replaceLine = strings.TrimPrefix(line, "replace ")
		} else if inReplaceBlock && strings.Contains(line, "=>") {
			replaceLine = line
		}

		if replaceLine != "" {
			dep := parseReplaceDirective(replaceLine, allModulePaths)
			if dep != "" {
				dependencies = append(dependencies, dep)
			}
		}

		// Also check require directives for direct dependencies on workspace modules
		if strings.HasPrefix(line, "require ") || inRequireBlock(line, lines) {
			dep := parseRequireDirective(line, allModulePaths)
			if dep != "" && !contains(dependencies, dep) {
				dependencies = append(dependencies, dep)
			}
		}
	}

	return dependencies, nil
}

// inRequireBlock is a helper that would need proper state tracking
// For simplicity, we focus on replace directives which are the primary way
// to link local modules
func inRequireBlock(_ string, _ []string) bool {
	return false // Replace directives are sufficient for local module ordering
}

// parseReplaceDirective extracts module path from a replace directive if it points to a local path
// Format: "module/path => ../relative/path" or "module/path => /absolute/path"
func parseReplaceDirective(line string, allModulePaths map[string]bool) string {
	parts := strings.Split(line, "=>")
	if len(parts) != 2 {
		return ""
	}

	// Get the original module path (left side)
	originalModule := strings.TrimSpace(strings.Fields(parts[0])[0])

	// Get the replacement (right side)
	replacement := strings.TrimSpace(parts[1])
	replacementParts := strings.Fields(replacement)
	if len(replacementParts) == 0 {
		return ""
	}
	replacementPath := replacementParts[0]

	// Check if this is a local path replacement (starts with . or /)
	if strings.HasPrefix(replacementPath, ".") || strings.HasPrefix(replacementPath, "/") {
		// This module depends on a local module
		// Check if the original module is in our workspace
		if allModulePaths[originalModule] {
			return originalModule
		}
	}

	return ""
}

// parseRequireDirective extracts module path from a require directive if it's a workspace module
func parseRequireDirective(line string, allModulePaths map[string]bool) string {
	line = strings.TrimPrefix(line, "require ")
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return ""
	}

	modulePath := parts[0]
	if allModulePaths[modulePath] {
		return modulePath
	}
	return ""
}

// contains checks if a slice contains a string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// sortModulesByDependency performs a topological sort of modules based on their dependencies
// Modules that depend on others are processed after their dependencies
// Falls back to root-first ordering if no inter-module dependencies exist
func sortModulesByDependency(modules []ModuleInfo) ([]ModuleInfo, error) {
	if len(modules) <= 1 {
		return modules, nil
	}

	// Build a map of module paths for quick lookup
	allModulePaths := make(map[string]bool)
	moduleByPath := make(map[string]ModuleInfo)
	for _, m := range modules {
		allModulePaths[m.Module] = true
		moduleByPath[m.Module] = m
	}

	// Parse dependencies for each module
	moduleDeps := make(map[string][]string)
	hasAnyDeps := false
	for _, m := range modules {
		deps, err := parseModuleDependencies(m, allModulePaths)
		if err != nil {
			utils.Warn("Failed to parse dependencies for %s: %v", m.Module, err)
			deps = []string{}
		}
		moduleDeps[m.Module] = deps
		if len(deps) > 0 {
			hasAnyDeps = true
		}
	}

	// If no inter-module dependencies, use root-first ordering
	if !hasAnyDeps {
		sortModules(modules)
		return modules, nil
	}

	// Perform topological sort using Kahn's algorithm
	// Calculate in-degree (number of dependencies) for each module
	inDegree := make(map[string]int)
	for _, m := range modules {
		if _, exists := inDegree[m.Module]; !exists {
			inDegree[m.Module] = 0
		}
		for _, dep := range moduleDeps[m.Module] {
			// Ensure dep is initialized in the map (zero value if not present)
			if _, exists := inDegree[dep]; !exists {
				inDegree[dep] = 0
			}
		}
	}

	// For topological sort, we need reverse dependency tracking
	// If A depends on B, B should come before A
	// So we track: for each module, which modules depend on it
	dependents := make(map[string][]string)
	for modulePath, deps := range moduleDeps {
		for _, dep := range deps {
			dependents[dep] = append(dependents[dep], modulePath)
		}
	}

	// Recalculate in-degree based on dependencies
	for modulePath := range inDegree {
		inDegree[modulePath] = len(moduleDeps[modulePath])
	}

	// Start with modules that have no dependencies
	var queue []string
	for _, m := range modules {
		if inDegree[m.Module] == 0 {
			queue = append(queue, m.Module)
		}
	}

	var sorted []ModuleInfo
	for len(queue) > 0 {
		// Sort queue to ensure deterministic order (root first within same level)
		sortQueue(queue, moduleByPath)

		// Pop first element
		current := queue[0]
		queue = queue[1:]

		sorted = append(sorted, moduleByPath[current])

		// Reduce in-degree for modules that depend on this one
		for _, dependent := range dependents[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Check for cycles
	if len(sorted) != len(modules) {
		utils.Warn("Circular dependencies detected, falling back to root-first ordering")
		sortModules(modules)
		return modules, nil
	}

	return sorted, nil
}

// sortQueue sorts the queue with root module first, then alphabetically
func sortQueue(queue []string, moduleByPath map[string]ModuleInfo) {
	sort.Slice(queue, func(i, j int) bool {
		mi := moduleByPath[queue[i]]
		mj := moduleByPath[queue[j]]
		if mi.IsRoot != mj.IsRoot {
			return mi.IsRoot // Root module first
		}
		return queue[i] < queue[j] // Then alphabetically
	})
}

// displayModuleSummary displays a summary of found modules in dependency order
func displayModuleSummary(modules []ModuleInfo, moduleDeps map[string][]string) {
	utils.Info("\nFound %d modules (dependency order):", len(modules))

	for i, m := range modules {
		location := m.Relative
		if location == "." {
			location = "./"
		} else {
			location = "./" + location
		}

		deps := moduleDeps[m.Module]
		if len(deps) > 0 {
			// Show short dependency names
			shortDeps := make([]string, len(deps))
			for j, dep := range deps {
				if idx := strings.LastIndex(dep, "/"); idx >= 0 {
					shortDeps[j] = dep[idx+1:]
				} else {
					shortDeps[j] = dep
				}
			}
			utils.Info("  %d. %-20s (%s) → depends on: %s", i+1, location, m.Module, strings.Join(shortDeps, ", "))
		} else {
			utils.Info("  %d. %-20s (%s)", i+1, location, m.Module)
		}
	}
	utils.Info("")
}

// shouldExcludeModule checks if a module should be excluded from processing
// based on the configured exclusion list. By default, "magefiles" modules are excluded
// because they have special build constraints (//go:build mage).
func shouldExcludeModule(module ModuleInfo, config *Config) bool {
	if config == nil || len(config.Test.ExcludeModules) == 0 {
		return false
	}
	for _, excludedName := range config.Test.ExcludeModules {
		if module.Name == excludedName {
			return true
		}
	}
	return false
}

// filterModulesForProcessing filters modules based on the exclusion configuration.
// This is used by test, lint, bench, and deps commands to skip modules like "magefiles"
// that have special build constraints.
func filterModulesForProcessing(modules []ModuleInfo, config *Config, operation string) []ModuleInfo {
	if config == nil || len(config.Test.ExcludeModules) == 0 {
		return modules
	}
	filtered := make([]ModuleInfo, 0, len(modules))
	for _, m := range modules {
		if shouldExcludeModule(m, config) {
			utils.Info("Skipping module %s (excluded from %s)", m.Name, operation)
			continue
		}
		filtered = append(filtered, m)
	}
	return filtered
}

// ModuleDiscoveryOptions configures module discovery and filtering behavior.
type ModuleDiscoveryOptions struct {
	Operation string // Description for logging (e.g., "benchmarks", "linting")
	Quiet     bool   // If true, suppress "Found X modules" message
}

// ModuleDiscoveryResult contains the result of module discovery and filtering.
type ModuleDiscoveryResult struct {
	Modules []ModuleInfo
	Empty   bool // True if no modules found at all
	Skipped bool // True if all modules were filtered out
}

// discoverAndFilterModules discovers all modules and filters them based on configuration.
// This consolidates the repeated pattern of findAllModules + filterModulesForProcessing.
// Returns (result, nil) for success, even if result.Empty or result.Skipped is true.
// Only returns error on actual discovery failure.
func discoverAndFilterModules(config *Config, opts ModuleDiscoveryOptions) (*ModuleDiscoveryResult, error) {
	// Discover all modules
	modules, err := findAllModules()
	if err != nil {
		return nil, fmt.Errorf("failed to find modules: %w", err)
	}

	if len(modules) == 0 {
		utils.Warn("No Go modules found")
		return &ModuleDiscoveryResult{Empty: true}, nil
	}

	// Show modules found (unless quiet mode)
	if !opts.Quiet && len(modules) > 1 {
		utils.Info("Found %d Go modules", len(modules))
	}

	// Filter modules based on exclusion configuration
	filtered := filterModulesForProcessing(modules, config, opts.Operation)
	if len(filtered) == 0 {
		utils.Warn("No modules to %s after exclusions", opts.Operation)
		return &ModuleDiscoveryResult{Skipped: true}, nil
	}

	return &ModuleDiscoveryResult{Modules: filtered}, nil
}

// ModuleCommandConfig configures a module-iterating command's initialization.
type ModuleCommandConfig struct {
	Header    string // Header message for utils.Header()
	Operation string // Operation name for discoverAndFilterModules
}

// ModuleCommandContext holds initialized command state after setup.
// A nil context with nil error indicates no modules to process (Empty or Skipped).
type ModuleCommandContext struct {
	Config  *Config
	Modules []ModuleInfo
}

// prepareModuleCommand consolidates command setup boilerplate.
// It performs: utils.Header(), GetConfig(), discoverAndFilterModules(), and Empty/Skipped check.
// Returns (nil, nil) if no modules to process (Empty or Skipped).
// Returns (nil, error) if setup failed.
// Returns (context, nil) on success with modules ready to process.
func prepareModuleCommand(cfg ModuleCommandConfig) (*ModuleCommandContext, error) {
	utils.Header(cfg.Header)

	config, err := GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get config: %w", err)
	}

	result, err := discoverAndFilterModules(config, ModuleDiscoveryOptions{
		Operation: cfg.Operation,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to discover modules: %w", err)
	}
	if result.Empty || result.Skipped {
		return nil, nil //nolint:nilnil // Intentional: nil context with nil error signals no modules to process
	}

	return &ModuleCommandContext{
		Config:  config,
		Modules: result.Modules,
	}, nil
}

// ModuleIteratorOptions configures the module iteration behavior.
type ModuleIteratorOptions struct {
	Operation string // Description for display (e.g., "Running benchmarks for")
	Verb      string // Verb for completion message (e.g., "completed", "passed")
}

// ModuleAction is the function signature for per-module operations.
// It receives the module and returns an error if the operation failed.
type ModuleAction func(module ModuleInfo) error

// forEachModule iterates over modules, running the action and collecting errors.
// It handles displayModuleHeader, timing, displayModuleCompletion, and error aggregation.
// Returns nil if all modules succeeded, or a formatted error summarizing failures.
func forEachModule(modules []ModuleInfo, opts ModuleIteratorOptions, action ModuleAction) error {
	totalStart := time.Now()
	var moduleErrors []moduleError

	for _, module := range modules {
		displayModuleHeader(module, opts.Operation)

		moduleStart := time.Now()
		err := action(module)
		if err != nil {
			moduleErrors = append(moduleErrors, moduleError{Module: module, Error: err})
		}
		displayModuleCompletion(module, opts.Operation, moduleStart, err)
	}

	if len(moduleErrors) > 0 {
		utils.Error("%s failed in %d/%d modules", opts.Operation, len(moduleErrors), len(modules))
		return formatModuleErrors(moduleErrors)
	}

	displayOverallCompletion(opts.Operation, opts.Verb, totalStart)
	return nil
}
