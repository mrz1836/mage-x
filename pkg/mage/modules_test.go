package mage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindAllModules(t *testing.T) {
	// Get current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			// Log the error but don't fail the test cleanup
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	// Create a temporary test directory structure
	tmpDir := t.TempDir()

	// Create test directory structure with multiple modules
	testDirs := []struct {
		path       string
		hasGoMod   bool
		moduleName string
	}{
		{path: ".", hasGoMod: true, moduleName: "github.com/test/main"},
		{path: ".github/actions/test", hasGoMod: true, moduleName: "github.com/test/action"},
		{path: "tools/helper", hasGoMod: true, moduleName: "github.com/test/helper"},
		{path: "vendor/example", hasGoMod: true, moduleName: "github.com/vendor/example"}, // Should be excluded
		{path: "pkg/utils", hasGoMod: false, moduleName: ""},                              // No go.mod
		{path: ".hidden/module", hasGoMod: true, moduleName: "github.com/test/hidden"},    // Hidden dir (not .github)
	}

	// Create directories and go.mod files
	for _, td := range testDirs {
		dir := filepath.Join(tmpDir, td.path)
		if mkdirErr := os.MkdirAll(dir, 0o750); mkdirErr != nil { // #nosec G301 -- test directory permissions
			t.Fatalf("Failed to create directory %s: %v", dir, mkdirErr)
		}

		if td.hasGoMod {
			goModPath := filepath.Join(dir, "go.mod")
			content := "module " + td.moduleName + "\n\ngo 1.24\n"
			if writeErr := os.WriteFile(goModPath, []byte(content), 0o600); writeErr != nil { // #nosec G306 -- test file permissions
				t.Fatalf("Failed to create go.mod in %s: %v", dir, writeErr)
			}
		}
	}

	// Change to temp directory for testing
	if chdirErr := os.Chdir(tmpDir); chdirErr != nil {
		t.Fatalf("Failed to change to temp directory: %v", chdirErr)
	}

	// Test findAllModules
	modules, err := findAllModules()
	if err != nil {
		t.Fatalf("findAllModules() error = %v", err)
	}

	// Expected modules (vendor and hidden dirs except .github should be excluded)
	expectedCount := 3 // main, .github/actions/test, tools/helper
	if len(modules) != expectedCount {
		t.Errorf("Expected %d modules, got %d", expectedCount, len(modules))
		for _, m := range modules {
			t.Logf("Found module: %s (path: %s)", m.Module, m.Relative)
		}
	}

	// Verify specific modules
	moduleMap := make(map[string]ModuleInfo)
	for _, m := range modules {
		moduleMap[m.Relative] = m
	}

	// Check root module is first
	if len(modules) > 0 && modules[0].Relative != "." {
		t.Errorf("Expected root module (.) to be first, got %s", modules[0].Relative)
	}

	// Check specific modules exist
	expectedModules := map[string]string{
		".":                    "github.com/test/main",
		".github/actions/test": "github.com/test/action",
		"tools/helper":         "github.com/test/helper",
	}

	for path, expectedModule := range expectedModules {
		if module, ok := moduleMap[path]; ok {
			if module.Module != expectedModule {
				t.Errorf("Module at %s: expected %s, got %s", path, expectedModule, module.Module)
			}
		} else {
			t.Errorf("Expected module at %s not found", path)
		}
	}

	// Verify vendor is excluded
	if _, ok := moduleMap["vendor/example"]; ok {
		t.Error("Vendor directory should be excluded")
	}

	// Verify hidden directories (except .github) are excluded
	if _, ok := moduleMap[".hidden/module"]; ok {
		t.Error("Hidden directories (except .github) should be excluded")
	}
}

func TestGetModuleNameFromFile(t *testing.T) {
	// Create a temporary go.mod file
	tmpFile := filepath.Join(t.TempDir(), "go.mod")

	tests := []struct {
		name        string
		content     string
		expected    string
		shouldError bool
	}{
		{
			name:        "simple module",
			content:     "module github.com/test/project\n\ngo 1.24\n",
			expected:    "github.com/test/project",
			shouldError: false,
		},
		{
			name:        "module with comment",
			content:     "module github.com/test/project // this is a comment\n\ngo 1.24\n",
			expected:    "github.com/test/project",
			shouldError: false,
		},
		{
			name:        "module with spaces",
			content:     "module    github.com/test/project    \n\ngo 1.24\n",
			expected:    "github.com/test/project",
			shouldError: false,
		},
		{
			name:        "no module statement",
			content:     "go 1.24\n",
			expected:    "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write test content
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0o600); err != nil { // #nosec G306 -- test file permissions
				t.Fatalf("Failed to write test file: %v", err)
			}

			// Test getModuleNameFromFile
			result, err := getModuleNameFromFile(tmpFile)

			if tt.shouldError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected module name %s, got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestSortModules(t *testing.T) {
	modules := []ModuleInfo{
		{Relative: "tools/helper", Module: "github.com/test/helper"},
		{Relative: ".github/actions/test", Module: "github.com/test/action"},
		{Relative: ".", Module: "github.com/test/main"},
		{Relative: "pkg/utils", Module: "github.com/test/utils"},
	}

	sortModules(modules)

	// Check that root module is first
	if modules[0].Relative != "." {
		t.Errorf("Expected root module (.) to be first, got %s", modules[0].Relative)
	}

	// Check that other modules are sorted alphabetically
	for i := 1; i < len(modules)-1; i++ {
		if modules[i].Relative > modules[i+1].Relative {
			t.Errorf("Modules not sorted: %s > %s", modules[i].Relative, modules[i+1].Relative)
		}
	}
}

func TestFormatModuleErrors(t *testing.T) {
	errors := []moduleError{
		{
			Module: ModuleInfo{Relative: ".", Module: "github.com/test/main"},
			Error:  os.ErrNotExist,
		},
		{
			Module: ModuleInfo{Relative: "tools/helper", Module: "github.com/test/helper"},
			Error:  os.ErrPermission,
		},
	}

	err := formatModuleErrors(errors)
	if err == nil {
		t.Error("Expected error, got nil")
	}

	errStr := err.Error()
	if !strings.Contains(errStr, "errors in 2 module(s)") {
		t.Errorf("Expected error message to contain module count, got: %s", errStr)
	}
	if !strings.Contains(errStr, "main module") {
		t.Errorf("Expected error message to contain 'main module', got: %s", errStr)
	}
	if !strings.Contains(errStr, "tools/helper") {
		t.Errorf("Expected error message to contain 'tools/helper', got: %s", errStr)
	}
}

func TestParseReplaceDirective(t *testing.T) {
	allModulePaths := map[string]bool{
		"github.com/test/core":    true,
		"github.com/test/utils":   true,
		"github.com/test/api":     true,
		"github.com/external/lib": true,
	}

	tests := []struct {
		name     string
		line     string
		expected string
	}{
		{
			name:     "local replace with relative path",
			line:     "github.com/test/core => ../core",
			expected: "github.com/test/core",
		},
		{
			name:     "local replace with absolute path",
			line:     "github.com/test/utils => /path/to/utils",
			expected: "github.com/test/utils",
		},
		{
			name:     "local replace with version",
			line:     "github.com/test/api v1.0.0 => ./api",
			expected: "github.com/test/api",
		},
		{
			name:     "external replace (not local path)",
			line:     "github.com/external/lib => github.com/fork/lib v1.2.3",
			expected: "",
		},
		{
			name:     "module not in workspace",
			line:     "github.com/other/pkg => ../other",
			expected: "",
		},
		{
			name:     "invalid format",
			line:     "invalid line without arrow",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseReplaceDirective(tt.line, allModulePaths)
			if result != tt.expected {
				t.Errorf("parseReplaceDirective(%q) = %q, want %q", tt.line, result, tt.expected)
			}
		})
	}
}

func TestParseModuleDependencies(t *testing.T) {
	tmpDir := t.TempDir()

	// Create module directories
	coreDir := filepath.Join(tmpDir, "core")
	utilsDir := filepath.Join(tmpDir, "utils")
	apiDir := filepath.Join(tmpDir, "api")

	for _, dir := range []string{coreDir, utilsDir, apiDir} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
	}

	// Create go.mod files
	// core has no dependencies
	coreGoMod := `module github.com/test/core

go 1.24
`
	if err := os.WriteFile(filepath.Join(coreDir, "go.mod"), []byte(coreGoMod), 0o600); err != nil {
		t.Fatalf("Failed to write core go.mod: %v", err)
	}

	// utils has no dependencies
	utilsGoMod := `module github.com/test/utils

go 1.24
`
	if err := os.WriteFile(filepath.Join(utilsDir, "go.mod"), []byte(utilsGoMod), 0o600); err != nil {
		t.Fatalf("Failed to write utils go.mod: %v", err)
	}

	// api depends on core and utils
	apiGoMod := `module github.com/test/api

go 1.24

require (
	github.com/test/core v0.0.0
	github.com/test/utils v0.0.0
)

replace (
	github.com/test/core => ../core
	github.com/test/utils => ../utils
)
`
	if err := os.WriteFile(filepath.Join(apiDir, "go.mod"), []byte(apiGoMod), 0o600); err != nil {
		t.Fatalf("Failed to write api go.mod: %v", err)
	}

	allModulePaths := map[string]bool{
		"github.com/test/core":  true,
		"github.com/test/utils": true,
		"github.com/test/api":   true,
	}

	// Test core module (no dependencies)
	coreModule := ModuleInfo{Path: coreDir, Module: "github.com/test/core"}
	coreDeps, err := parseModuleDependencies(coreModule, allModulePaths)
	if err != nil {
		t.Errorf("parseModuleDependencies(core) error = %v", err)
	}
	if len(coreDeps) != 0 {
		t.Errorf("Expected core to have 0 dependencies, got %d: %v", len(coreDeps), coreDeps)
	}

	// Test api module (depends on core and utils)
	apiModule := ModuleInfo{Path: apiDir, Module: "github.com/test/api"}
	apiDeps, err := parseModuleDependencies(apiModule, allModulePaths)
	if err != nil {
		t.Errorf("parseModuleDependencies(api) error = %v", err)
	}
	if len(apiDeps) != 2 {
		t.Errorf("Expected api to have 2 dependencies, got %d: %v", len(apiDeps), apiDeps)
	}

	// Verify both dependencies are found
	depsMap := make(map[string]bool)
	for _, dep := range apiDeps {
		depsMap[dep] = true
	}
	if !depsMap["github.com/test/core"] {
		t.Error("Expected api to depend on github.com/test/core")
	}
	if !depsMap["github.com/test/utils"] {
		t.Error("Expected api to depend on github.com/test/utils")
	}
}

func TestSortModulesByDependency_NoDependencies(t *testing.T) {
	// Create temp directory with modules that have no inter-dependencies
	tmpDir := t.TempDir()

	modules := []ModuleInfo{
		{Path: filepath.Join(tmpDir, "b"), Module: "github.com/test/b", Relative: "b"},
		{Path: filepath.Join(tmpDir, "a"), Module: "github.com/test/a", Relative: "a"},
		{Path: tmpDir, Module: "github.com/test/root", Relative: ".", IsRoot: true},
	}

	// Create go.mod files with no replace directives
	for _, m := range modules {
		if err := os.MkdirAll(m.Path, 0o750); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		goModContent := "module " + m.Module + "\n\ngo 1.24\n"
		if err := os.WriteFile(filepath.Join(m.Path, "go.mod"), []byte(goModContent), 0o600); err != nil {
			t.Fatalf("Failed to write go.mod: %v", err)
		}
	}

	sorted, err := sortModulesByDependency(modules)
	if err != nil {
		t.Errorf("sortModulesByDependency() error = %v", err)
	}

	// Root module should be first (falls back to root-first ordering)
	if len(sorted) > 0 && !sorted[0].IsRoot {
		t.Errorf("Expected root module to be first, got %s", sorted[0].Relative)
	}
}

func TestSortModulesByDependency_LinearChain(t *testing.T) {
	// Create temp directory with linear dependency chain: root -> api -> core
	tmpDir := t.TempDir()

	coreDir := filepath.Join(tmpDir, "core")
	apiDir := filepath.Join(tmpDir, "api")

	for _, dir := range []string{coreDir, apiDir} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
	}

	// core has no dependencies
	coreGoMod := "module github.com/test/core\n\ngo 1.24\n"
	if err := os.WriteFile(filepath.Join(coreDir, "go.mod"), []byte(coreGoMod), 0o600); err != nil {
		t.Fatalf("Failed to write core go.mod: %v", err)
	}

	// api depends on core
	apiGoMod := `module github.com/test/api

go 1.24

replace github.com/test/core => ../core
`
	if err := os.WriteFile(filepath.Join(apiDir, "go.mod"), []byte(apiGoMod), 0o600); err != nil {
		t.Fatalf("Failed to write api go.mod: %v", err)
	}

	// root depends on api
	rootGoMod := `module github.com/test/root

go 1.24

replace github.com/test/api => ./api
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(rootGoMod), 0o600); err != nil {
		t.Fatalf("Failed to write root go.mod: %v", err)
	}

	modules := []ModuleInfo{
		{Path: tmpDir, Module: "github.com/test/root", Relative: ".", IsRoot: true},
		{Path: apiDir, Module: "github.com/test/api", Relative: "api"},
		{Path: coreDir, Module: "github.com/test/core", Relative: "core"},
	}

	sorted, err := sortModulesByDependency(modules)
	if err != nil {
		t.Errorf("sortModulesByDependency() error = %v", err)
	}

	if len(sorted) != 3 {
		t.Fatalf("Expected 3 modules, got %d", len(sorted))
	}

	// Core should be first (no dependencies)
	// API should be second (depends on core)
	// Root should be last (depends on api)
	expectedOrder := []string{"github.com/test/core", "github.com/test/api", "github.com/test/root"}
	for i, expected := range expectedOrder {
		if sorted[i].Module != expected {
			t.Errorf("Position %d: expected %s, got %s", i, expected, sorted[i].Module)
		}
	}
}

func TestSortModulesByDependency_CyclicDependency(t *testing.T) {
	// Create temp directory with cyclic dependencies: a -> b -> a
	tmpDir := t.TempDir()

	aDir := filepath.Join(tmpDir, "a")
	bDir := filepath.Join(tmpDir, "b")

	for _, dir := range []string{aDir, bDir} {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
	}

	// a depends on b
	aGoMod := `module github.com/test/a

go 1.24

replace github.com/test/b => ../b
`
	if err := os.WriteFile(filepath.Join(aDir, "go.mod"), []byte(aGoMod), 0o600); err != nil {
		t.Fatalf("Failed to write a go.mod: %v", err)
	}

	// b depends on a (creating a cycle)
	bGoMod := `module github.com/test/b

go 1.24

replace github.com/test/a => ../a
`
	if err := os.WriteFile(filepath.Join(bDir, "go.mod"), []byte(bGoMod), 0o600); err != nil {
		t.Fatalf("Failed to write b go.mod: %v", err)
	}

	modules := []ModuleInfo{
		{Path: aDir, Module: "github.com/test/a", Relative: "a"},
		{Path: bDir, Module: "github.com/test/b", Relative: "b"},
	}

	// Should fall back to root-first ordering due to cycle
	sorted, err := sortModulesByDependency(modules)
	if err != nil {
		t.Errorf("sortModulesByDependency() error = %v", err)
	}

	// Should still return all modules (fallback to simple sort)
	if len(sorted) != 2 {
		t.Errorf("Expected 2 modules even with cycle, got %d", len(sorted))
	}
}

func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}

	if !contains(slice, "a") {
		t.Error("Expected contains(slice, 'a') to be true")
	}
	if !contains(slice, "c") {
		t.Error("Expected contains(slice, 'c') to be true")
	}
	if contains(slice, "d") {
		t.Error("Expected contains(slice, 'd') to be false")
	}
	if contains(nil, "a") {
		t.Error("Expected contains(nil, 'a') to be false")
	}
}
