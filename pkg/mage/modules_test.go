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
