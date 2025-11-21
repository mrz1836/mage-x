package mage

// Module represents a Go module for operations
type Module = ModuleInfo

// ModuleError represents an error that occurred during module processing
type ModuleError = moduleError

// FindAllModules discovers all go.mod files in the project
func FindAllModules() ([]Module, error) {
	return findAllModules()
}

// RunCommandInModule executes a command in the module directory
func RunCommandInModule(module Module, command string, args ...string) error {
	return runCommandInModule(module, command, args...)
}

// FormatModuleErrors formats multiple module errors into a single error
func FormatModuleErrors(errors []ModuleError) error {
	return formatModuleErrors(errors)
}

// SortModulesByDependency performs a topological sort of modules based on their dependencies
func SortModulesByDependency(modules []Module) ([]Module, error) {
	return sortModulesByDependency(modules)
}

// ParseModuleDependencies reads a go.mod file and extracts local replace directives
func ParseModuleDependencies(module Module, allModulePaths map[string]bool) ([]string, error) {
	return parseModuleDependencies(module, allModulePaths)
}

// DisplayModuleSummary displays a summary of found modules in dependency order
func DisplayModuleSummary(modules []Module, moduleDeps map[string][]string) {
	displayModuleSummary(modules, moduleDeps)
}
