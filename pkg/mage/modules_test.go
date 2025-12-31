package mage

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
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

// TestRunCommandInModuleWithRunner_DirRunner verifies that the DirRunner interface
// is used when available, avoiding the goroutine-unsafe os.Chdir().
func TestRunCommandInModuleWithRunner_DirRunner(t *testing.T) {
	t.Parallel()

	// Create a mock DirRunner that tracks the directory passed
	mock := &mockDirRunner{
		runCmdCalls:      make([]string, 0),
		runCmdInDirCalls: make([]mockDirCall, 0),
	}

	module := ModuleInfo{
		Path:     "/test/module/path",
		Module:   "github.com/test/module",
		Relative: "module",
	}

	// Run command - should use RunCmdInDir, not RunCmd
	err := runCommandInModuleWithRunner(module, mock, "go", "test", "./...")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify RunCmdInDir was called, not RunCmd
	if len(mock.runCmdInDirCalls) != 1 {
		t.Errorf("Expected 1 RunCmdInDir call, got %d", len(mock.runCmdInDirCalls))
	}
	if len(mock.runCmdCalls) != 0 {
		t.Errorf("Expected 0 RunCmd calls, got %d", len(mock.runCmdCalls))
	}

	// Verify correct directory was passed
	if len(mock.runCmdInDirCalls) > 0 {
		call := mock.runCmdInDirCalls[0]
		if call.dir != "/test/module/path" {
			t.Errorf("Expected dir '/test/module/path', got '%s'", call.dir)
		}
		if call.cmd != "go" {
			t.Errorf("Expected cmd 'go', got '%s'", call.cmd)
		}
	}
}

// TestRunCommandInModuleWithRunner_FallbackMutex verifies that a mutex is used
// for runners that don't implement DirRunner.
func TestRunCommandInModuleWithRunner_FallbackMutex(t *testing.T) {
	// Note: This test cannot truly verify mutex behavior without race detector,
	// but it validates that the fallback path works correctly.

	tmpDir := t.TempDir()

	// Create a simple runner that doesn't implement DirRunner
	mock := &simpleRunner{
		runCmdCalls: make([]string, 0),
	}

	module := ModuleInfo{
		Path:     tmpDir,
		Module:   "github.com/test/module",
		Relative: ".",
	}

	// Save original directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalDir) //nolint:errcheck // Best effort cleanup in test
	}()

	// Run command - should use RunCmd with directory change
	err = runCommandInModuleWithRunner(module, mock, "echo", "test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify RunCmd was called
	if len(mock.runCmdCalls) != 1 {
		t.Errorf("Expected 1 RunCmd call, got %d", len(mock.runCmdCalls))
	}

	// Verify we're back in the original directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory after test: %v", err)
	}
	if currentDir != originalDir {
		t.Errorf("Working directory not restored: expected %s, got %s", originalDir, currentDir)
	}
}

// mockDirCall records a call to RunCmdInDir
type mockDirCall struct {
	dir  string
	cmd  string
	args []string
}

// mockDirRunner implements both CommandRunner and DirRunner for testing
type mockDirRunner struct {
	runCmdCalls      []string
	runCmdInDirCalls []mockDirCall
}

func (m *mockDirRunner) RunCmd(name string, args ...string) error {
	m.runCmdCalls = append(m.runCmdCalls, name)
	return nil
}

func (m *mockDirRunner) RunCmdOutput(name string, _ ...string) (string, error) {
	return name, nil
}

func (m *mockDirRunner) RunCmdInDir(dir, name string, args ...string) error {
	m.runCmdInDirCalls = append(m.runCmdInDirCalls, mockDirCall{dir: dir, cmd: name, args: args})
	return nil
}

func (m *mockDirRunner) RunCmdOutputInDir(dir, name string, _ ...string) (string, error) {
	return dir + ":" + name, nil
}

// simpleRunner implements only CommandRunner (not DirRunner) for testing fallback
type simpleRunner struct {
	runCmdCalls []string
}

func (s *simpleRunner) RunCmd(name string, _ ...string) error {
	s.runCmdCalls = append(s.runCmdCalls, name)
	return nil
}

func (s *simpleRunner) RunCmdOutput(name string, _ ...string) (string, error) {
	return name, nil
}

// TestRunInModuleDir_InvalidPath verifies error handling when module path doesn't exist.
func TestRunInModuleDir_InvalidPath(t *testing.T) {
	t.Parallel()

	// Use a simple runner (no DirRunner) to trigger the chdir fallback path
	mock := &simpleRunner{
		runCmdCalls: make([]string, 0),
	}

	module := ModuleInfo{
		Path:     "/nonexistent/path/that/definitely/does/not/exist",
		Module:   "github.com/test/module",
		Relative: "nonexistent",
	}

	// Try to run command - should fail because directory doesn't exist
	err := runCommandInModuleWithRunner(module, mock, "echo", "test")
	if err == nil {
		t.Fatal("Expected error for invalid path, got nil")
	}

	// Verify error message mentions the path
	if !strings.Contains(err.Error(), "failed to change to directory") {
		t.Errorf("Expected error about changing directory, got: %v", err)
	}

	// Verify command was not called
	if len(mock.runCmdCalls) != 0 {
		t.Errorf("Expected 0 RunCmd calls for invalid path, got %d", len(mock.runCmdCalls))
	}
}

// TestRunCommandInModuleOutputWithRunner_DirRunner verifies output capture uses DirRunner.
func TestRunCommandInModuleOutputWithRunner_DirRunner(t *testing.T) {
	t.Parallel()

	mock := &mockDirRunner{
		runCmdCalls:      make([]string, 0),
		runCmdInDirCalls: make([]mockDirCall, 0),
	}

	module := ModuleInfo{
		Path:     "/test/module/path",
		Module:   "github.com/test/module",
		Relative: "module",
	}

	// Run command with output capture
	output, err := runCommandInModuleOutputWithRunner(module, mock, "go", "list", "./...")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// mockDirRunner.RunCmdOutputInDir returns "dir:name"
	expectedOutput := "/test/module/path:go"
	if output != expectedOutput {
		t.Errorf("Expected output '%s', got '%s'", expectedOutput, output)
	}

	// Verify no RunCmd calls were made (only DirRunner methods should be used)
	if len(mock.runCmdCalls) != 0 {
		t.Errorf("Expected 0 RunCmd calls, got %d", len(mock.runCmdCalls))
	}
}

// TestRunInModuleDir_DirectoryRestored verifies directory is restored after command execution.
func TestRunInModuleDir_DirectoryRestored(t *testing.T) {
	// This test must not run in parallel since it changes the working directory
	tmpDir := t.TempDir()

	mock := &simpleRunner{
		runCmdCalls: make([]string, 0),
	}

	module := ModuleInfo{
		Path:     tmpDir,
		Module:   "github.com/test/module",
		Relative: ".",
	}

	// Get original directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get original directory: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalDir) //nolint:errcheck // Best effort cleanup
	}()

	// Run multiple commands to ensure directory is always restored
	for i := 0; i < 3; i++ {
		err = runCommandInModuleWithRunner(module, mock, "echo", "test")
		if err != nil {
			t.Fatalf("Iteration %d: Unexpected error: %v", i, err)
		}

		currentDir, err := os.Getwd()
		if err != nil {
			t.Fatalf("Iteration %d: Failed to get current directory: %v", i, err)
		}
		if currentDir != originalDir {
			t.Errorf("Iteration %d: Directory not restored: expected %s, got %s", i, originalDir, currentDir)
		}
	}
}

// TestRunInModuleDir_ConcurrentSafety verifies mutex protects concurrent chdir calls.
func TestRunInModuleDir_ConcurrentSafety(t *testing.T) {
	t.Parallel()

	// Create multiple temp directories
	dirs := make([]string, 5)
	for i := range dirs {
		dirs[i] = t.TempDir()
	}

	// Run concurrent commands to different directories
	var wg sync.WaitGroup
	errChan := make(chan error, len(dirs)*2)

	for _, dir := range dirs {
		for j := 0; j < 2; j++ {
			wg.Add(1)
			go func(path string) {
				defer wg.Done()

				module := ModuleInfo{
					Path:     path,
					Module:   "github.com/test/module",
					Relative: ".",
				}

				// Use runInModuleDir directly to test the generic function
				_, err := runInModuleDir(module, &simpleRunner{},
					func(_ DirRunner, _ string) (struct{}, error) {
						return struct{}{}, nil
					},
					func() (struct{}, error) {
						// Simulate some work in the fallback path
						time.Sleep(time.Millisecond)
						return struct{}{}, nil
					},
				)
				if err != nil {
					errChan <- err
				}
			}(dir)
		}
	}

	wg.Wait()
	close(errChan)

	// Check for any errors
	for err := range errChan {
		t.Errorf("Concurrent execution error: %v", err)
	}
}

func TestDisplayModuleCompletion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		module    ModuleInfo
		operation string
		err       error
	}{
		{
			name:      "success for root module",
			module:    ModuleInfo{Relative: "."},
			operation: "Linting",
			err:       nil,
		},
		{
			name:      "success for submodule",
			module:    ModuleInfo{Relative: "pkg/utils"},
			operation: "Tests",
			err:       nil,
		},
		{
			name:      "error for module",
			module:    ModuleInfo{Relative: "pkg/api"},
			operation: "Benchmarks",
			err:       os.ErrNotExist,
		},
		{
			name:      "error for root module",
			module:    ModuleInfo{Relative: "."},
			operation: "Vet",
			err:       os.ErrPermission,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Just verify it doesn't panic - output goes to utils logger
			displayModuleCompletion(tt.module, tt.operation, fakeStartTime(), tt.err)
		})
	}
}

func TestDisplayModuleCompletionWithSuffix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		module    ModuleInfo
		operation string
		suffix    string
		err       error
	}{
		{
			name:      "with build tag suffix",
			module:    ModuleInfo{Relative: "pkg/api"},
			operation: "Coverage tests",
			suffix:    " [integration]",
			err:       nil,
		},
		{
			name:      "with empty suffix",
			module:    ModuleInfo{Relative: "."},
			operation: "Tests",
			suffix:    "",
			err:       nil,
		},
		{
			name:      "error with suffix",
			module:    ModuleInfo{Relative: "pkg/core"},
			operation: "Unit tests",
			suffix:    " (tag: unit)",
			err:       os.ErrNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Just verify it doesn't panic - output goes to utils logger
			displayModuleCompletionWithSuffix(tt.module, tt.operation, tt.suffix, fakeStartTime(), tt.err)
		})
	}
}

func TestDisplayModuleCompletionWithOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts ModuleCompletionOptions
	}{
		{
			name: "default success verb",
			opts: ModuleCompletionOptions{
				Module:    ModuleInfo{Relative: "pkg/api"},
				Operation: "Linting",
				StartTime: fakeStartTime(),
				Err:       nil,
			},
		},
		{
			name: "custom success verb - fixed",
			opts: ModuleCompletionOptions{
				Module:      ModuleInfo{Relative: "pkg/api"},
				Operation:   "All issues",
				StartTime:   fakeStartTime(),
				Err:         nil,
				SuccessVerb: "fixed",
			},
		},
		{
			name: "with suffix and custom verb",
			opts: ModuleCompletionOptions{
				Module:      ModuleInfo{Relative: "."},
				Operation:   "Format",
				Suffix:      " [yaml]",
				StartTime:   fakeStartTime(),
				Err:         nil,
				SuccessVerb: "applied",
			},
		},
		{
			name: "error ignores success verb",
			opts: ModuleCompletionOptions{
				Module:      ModuleInfo{Relative: "pkg/core"},
				Operation:   "Build",
				StartTime:   fakeStartTime(),
				Err:         os.ErrNotExist,
				SuccessVerb: "completed", // Should be ignored for errors
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Just verify it doesn't panic - output goes to utils logger
			displayModuleCompletionWithOptions(tt.opts)
		})
	}
}

func TestDisplayOverallCompletion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation string
		verb      string
	}{
		{name: "linting passed", operation: "linting", verb: "passed"},
		{name: "benchmarks completed", operation: "benchmarks", verb: "completed"},
		{name: "vet checks passed", operation: "vet checks", verb: "passed"},
		{name: "tests completed", operation: "tests", verb: "completed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Just verify it doesn't panic - output goes to utils logger
			displayOverallCompletion(tt.operation, tt.verb, fakeStartTime())
		})
	}
}

func TestDisplayOverallCompletionWithOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		opts OverallCompletionOptions
	}{
		{
			name: "simple operation",
			opts: OverallCompletionOptions{
				Operation: "linting",
				Verb:      "passed",
				StartTime: fakeStartTime(),
			},
		},
		{
			name: "with prefix",
			opts: OverallCompletionOptions{
				Prefix:    "Unit ",
				Operation: "tests",
				Verb:      "passed",
				StartTime: fakeStartTime(),
			},
		},
		{
			name: "with suffix",
			opts: OverallCompletionOptions{
				Operation: "coverage tests",
				Suffix:    " [integration]",
				Verb:      "passed",
				StartTime: fakeStartTime(),
			},
		},
		{
			name: "with prefix and suffix",
			opts: OverallCompletionOptions{
				Prefix:    "Short ",
				Operation: "tests",
				Suffix:    " (tag: unit)",
				Verb:      "completed",
				StartTime: fakeStartTime(),
			},
		},
		{
			name: "lint issues fixed with suffix",
			opts: OverallCompletionOptions{
				Operation: "lint issues",
				Suffix:    " and code formatted",
				Verb:      "fixed",
				StartTime: fakeStartTime(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Just verify it doesn't panic - output goes to utils logger
			displayOverallCompletionWithOptions(tt.opts)
		})
	}
}

// fakeStartTime returns a time in the past for duration formatting tests
func fakeStartTime() time.Time {
	return time.Now().Add(-100 * time.Millisecond)
}

func TestPrepareModuleCommand(t *testing.T) {
	// Get current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	t.Run("returns context when modules found", func(t *testing.T) {
		// Create a temporary test directory with a go.mod
		tmpDir := t.TempDir()
		goModPath := filepath.Join(tmpDir, "go.mod")
		content := "module github.com/test/project\n\ngo 1.24\n"
		if writeErr := os.WriteFile(goModPath, []byte(content), 0o600); writeErr != nil {
			t.Fatalf("Failed to create go.mod: %v", writeErr)
		}

		// Change to temp directory
		if chdirErr := os.Chdir(tmpDir); chdirErr != nil {
			t.Fatalf("Failed to change to temp directory: %v", chdirErr)
		}

		ctx, err := prepareModuleCommand(ModuleCommandConfig{
			Header:    "Test Header",
			Operation: "testing",
		})
		if err != nil {
			t.Errorf("prepareModuleCommand() error = %v", err)
		}
		if ctx == nil {
			t.Error("prepareModuleCommand() returned nil context, expected non-nil")
		}
		if ctx != nil && len(ctx.Modules) == 0 {
			t.Error("prepareModuleCommand() returned context with no modules")
		}
		if ctx != nil && ctx.Config == nil {
			t.Error("prepareModuleCommand() returned context with nil Config")
		}
	})

	t.Run("returns nil context when no modules found", func(t *testing.T) {
		// Create an empty temp directory (no go.mod)
		tmpDir := t.TempDir()

		// Change to temp directory
		if chdirErr := os.Chdir(tmpDir); chdirErr != nil {
			t.Fatalf("Failed to change to temp directory: %v", chdirErr)
		}

		ctx, err := prepareModuleCommand(ModuleCommandConfig{
			Header:    "Test Header",
			Operation: "testing",
		})
		if err != nil {
			t.Errorf("prepareModuleCommand() error = %v, expected nil", err)
		}
		if ctx != nil {
			t.Errorf("prepareModuleCommand() returned non-nil context, expected nil for empty directory")
		}
	})

	t.Run("context contains correct config and modules", func(t *testing.T) {
		// Create a temp directory with a go.mod
		tmpDir := t.TempDir()
		goModPath := filepath.Join(tmpDir, "go.mod")
		content := "module github.com/test/context\n\ngo 1.24\n"
		if writeErr := os.WriteFile(goModPath, []byte(content), 0o600); writeErr != nil {
			t.Fatalf("Failed to create go.mod: %v", writeErr)
		}

		// Change to temp directory
		if chdirErr := os.Chdir(tmpDir); chdirErr != nil {
			t.Fatalf("Failed to change to temp directory: %v", chdirErr)
		}

		ctx, err := prepareModuleCommand(ModuleCommandConfig{
			Header:    "Verify Context",
			Operation: "verification",
		})
		if err != nil {
			t.Fatalf("prepareModuleCommand() error = %v", err)
		}
		if ctx == nil {
			t.Fatal("prepareModuleCommand() returned nil context")
		}

		// Verify modules contain the expected module
		foundModule := false
		for _, m := range ctx.Modules {
			if m.Module == "github.com/test/context" {
				foundModule = true
				break
			}
		}
		if !foundModule {
			t.Error("Expected to find module 'github.com/test/context' in context.Modules")
		}

		// Verify config is populated (should have defaults at minimum)
		if ctx.Config == nil {
			t.Error("Expected Config to be non-nil")
		}
	})
}

func TestModuleCommandConfig(t *testing.T) {
	t.Run("config struct has required fields", func(t *testing.T) {
		t.Parallel()
		cfg := ModuleCommandConfig{
			Header:    "Test Header",
			Operation: "testing",
		}

		if cfg.Header != "Test Header" {
			t.Errorf("Header = %q, want %q", cfg.Header, "Test Header")
		}
		if cfg.Operation != "testing" {
			t.Errorf("Operation = %q, want %q", cfg.Operation, "testing")
		}
	})
}

func TestModuleCommandContext(t *testing.T) {
	t.Run("context struct holds config and modules", func(t *testing.T) {
		t.Parallel()
		modules := []ModuleInfo{
			{Path: "/test/path", Module: "github.com/test/module"},
		}
		config := &Config{}

		ctx := &ModuleCommandContext{
			Config:  config,
			Modules: modules,
		}

		if ctx.Config != config {
			t.Error("Config not stored correctly in context")
		}
		if len(ctx.Modules) != 1 {
			t.Errorf("Modules length = %d, want 1", len(ctx.Modules))
		}
		if ctx.Modules[0].Module != "github.com/test/module" {
			t.Errorf("Module = %q, want %q", ctx.Modules[0].Module, "github.com/test/module")
		}
	})
}
