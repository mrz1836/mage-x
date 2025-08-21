package registry

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	buildCmdName = "Build"
)

func TestNewLoader(t *testing.T) {
	r := NewRegistry()
	loader := NewLoader(r)

	if loader == nil {
		t.Fatal("NewLoader() returned nil")
	}
	if loader.registry != r {
		t.Error("Loader registry not set correctly")
	}
}

func TestNewLoader_NilRegistry(t *testing.T) {
	loader := NewLoader(nil)

	if loader == nil {
		t.Fatal("NewLoader() returned nil")
	}
	if loader.registry != Global() {
		t.Error("Loader should use global registry when nil passed")
	}
}

func TestLoader_parseMagefile(t *testing.T) {
	loader := NewLoader(NewRegistry())

	// Create a temporary magefile
	tmpDir := t.TempDir()
	magefilePath := filepath.Join(tmpDir, "magefile.go")

	magefileContent := `//go:build mage
package main

// Build builds the project
func Build() error {
	return nil
}

// Test runs the tests
func Test() error {
	return nil
}

// unexported function should be ignored
func helper() error {
	return nil
}

// Build namespace for advanced builds
type Build struct{}

// Linux builds for Linux
func (Build) Linux() error {
	return nil
}

// Windows builds for Windows
func (Build) Windows() error {
	return nil
}
`

	err := os.WriteFile(magefilePath, []byte(magefileContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test magefile: %v", err)
	}

	commands, err := loader.parseMagefile(magefilePath)
	if err != nil {
		t.Fatalf("parseMagefile() failed: %v", err)
	}

	if len(commands) < 4 {
		t.Errorf("Expected at least 4 commands, got %d", len(commands))
	}

	// Check for specific commands
	var foundBuild, foundTest, foundLinux, foundWindows bool
	for _, cmd := range commands {
		switch {
		case cmd.Name == buildCmdName && !cmd.IsNamespace:
			foundBuild = true
			if cmd.Description != buildCmdName+" builds the project" {
				t.Errorf("%s description = %q, expected '%s builds the project'", buildCmdName, cmd.Description, buildCmdName)
			}
		case cmd.Name == "Test":
			foundTest = true
			if cmd.Description != "Test runs the tests" {
				t.Errorf("Test description = %q, expected 'Test runs the tests'", cmd.Description)
			}
		case cmd.Method == "Linux" && cmd.IsNamespace:
			foundLinux = true
			if cmd.Namespace != buildCmdName {
				t.Errorf("Linux namespace = %q, expected '%s'", cmd.Namespace, buildCmdName)
			}
		case cmd.Method == "Windows" && cmd.IsNamespace:
			foundWindows = true
			if cmd.Namespace != buildCmdName {
				t.Errorf("Windows namespace = %q, expected '%s'", cmd.Namespace, buildCmdName)
			}
		case cmd.Name == "helper":
			t.Error("Unexported function 'helper' should not be included")
		}
	}

	if !foundBuild {
		t.Error("Build command not found")
	}
	if !foundTest {
		t.Error("Test command not found")
	}
	if !foundLinux {
		t.Error("Linux method not found")
	}
	if !foundWindows {
		t.Error("Windows method not found")
	}
}

func TestLoader_parseMagefile_EmptyFile(t *testing.T) {
	loader := NewLoader(NewRegistry())

	tmpDir := t.TempDir()
	magefilePath := filepath.Join(tmpDir, "magefile.go")

	emptyContent := `package main`
	err := os.WriteFile(magefilePath, []byte(emptyContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test magefile: %v", err)
	}

	commands, err := loader.parseMagefile(magefilePath)
	if err != nil {
		t.Fatalf("parseMagefile() failed: %v", err)
	}

	if len(commands) != 0 {
		t.Errorf("Expected 0 commands from empty file, got %d", len(commands))
	}
}

func TestCommandInfo(t *testing.T) {
	info := CommandInfo{
		Name:        "Build",
		IsNamespace: false,
		Namespace:   "",
		Method:      "",
		Description: "Build the project",
	}

	if info.Name != "Build" {
		t.Errorf("Name = %q, expected 'Build'", info.Name)
	}
	if info.IsNamespace {
		t.Error("IsNamespace should be false")
	}
	if info.Description != "Build the project" {
		t.Errorf("Description = %q, expected 'Build the project'", info.Description)
	}
}

func TestExtractDescription(t *testing.T) {
	tests := []struct {
		name     string
		comments []string
		expected string
	}{
		{
			name:     "single line comment",
			comments: []string{"// Build builds the project"},
			expected: buildCmdName + " builds the project",
		},
		{
			name:     "multi line comment",
			comments: []string{"// Build builds the project", "// with advanced features"},
			expected: "Build builds the project with advanced features",
		},
		{
			name:     "block comment",
			comments: []string{"/* Build builds the project */"},
			expected: buildCmdName + " builds the project",
		},
		{
			name:     "mixed comments",
			comments: []string{"// Build builds", "/* the project */"},
			expected: buildCmdName + " builds the project",
		},
		{
			name:     "empty comments",
			comments: []string{"//", "// ", "//   "},
			expected: "",
		},
		{
			name:     "no comments",
			comments: []string{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a comment group
			var comments []*ast.Comment
			for _, text := range tt.comments {
				comments = append(comments, &ast.Comment{Text: text})
			}

			var doc *ast.CommentGroup
			if len(comments) > 0 {
				doc = &ast.CommentGroup{List: comments}
			}

			result := extractDescription(doc)
			if result != tt.expected {
				t.Errorf("extractDescription() = %q, expected %q", result, tt.expected)
			}
		})
	}
}

func TestGetReceiverType(t *testing.T) {
	tests := []struct {
		name     string
		receiver string
		expected string
	}{
		{
			name:     "pointer receiver",
			receiver: "func (*Build) Linux() error { return nil }",
			expected: "Build",
		},
		{
			name:     "value receiver",
			receiver: "func (Build) Linux() error { return nil }",
			expected: "Build",
		},
		{
			name:     "named receiver",
			receiver: "func (b Build) Linux() error { return nil }",
			expected: "Build",
		},
		{
			name:     "named pointer receiver",
			receiver: "func (b *Build) Linux() error { return nil }",
			expected: "Build",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the method declaration
			code := `package main
			` + tt.receiver

			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, "", code, 0)
			if err != nil {
				t.Fatalf("Failed to parse code: %v", err)
			}

			// Find the function declaration
			for _, decl := range file.Decls {
				if fn, ok := decl.(*ast.FuncDecl); ok {
					result := getReceiverType(fn.Recv)
					if result != tt.expected {
						t.Errorf("getReceiverType() = %q, expected %q", result, tt.expected)
					}
					return
				}
			}

			t.Error("No function declaration found")
		})
	}
}

func TestGetReceiverType_EdgeCases(t *testing.T) {
	// Test nil receiver
	result := getReceiverType(nil)
	if result != "" {
		t.Errorf("getReceiverType(nil) = %q, expected empty string", result)
	}

	// Test empty receiver list
	recv := &ast.FieldList{List: []*ast.Field{}}
	result = getReceiverType(recv)
	if result != "" {
		t.Errorf("getReceiverType(empty) = %q, expected empty string", result)
	}

	// Test receiver with nil type
	recv = &ast.FieldList{
		List: []*ast.Field{
			{Type: nil},
		},
	}
	result = getReceiverType(recv)
	if result != "" {
		t.Errorf("getReceiverType(nil type) = %q, expected empty string", result)
	}
}

func TestLoader_Verbose(t *testing.T) {
	// Test default (non-verbose)
	loader := NewLoader(NewRegistry())
	if loader.verbose {
		t.Error("Loader should not be verbose by default")
	}

	// Test verbose mode
	if err := os.Setenv("MAGE_X_VERBOSE", "true"); err != nil {
		// Test environment variable setting is non-critical
		_ = err
	}
	defer func() {
		if err := os.Unsetenv("MAGE_X_VERBOSE"); err != nil {
			// Test cleanup error is non-critical
			_ = err
		}
	}()

	loader = NewLoader(NewRegistry())
	if !loader.verbose {
		t.Error("Loader should be verbose when MAGE_X_VERBOSE=true")
	}
}

func BenchmarkLoader_ParseMagefile(b *testing.B) {
	loader := NewLoader(NewRegistry())

	// Create a test magefile
	tmpDir := b.TempDir()
	magefilePath := filepath.Join(tmpDir, "magefile.go")

	magefileContent := `//go:build mage
package main

func Build() error { return nil }
func Test() error { return nil }
func Clean() error { return nil }

type Build struct{}
func (Build) Linux() error { return nil }
func (Build) Windows() error { return nil }
func (Build) Darwin() error { return nil }
`

	err := os.WriteFile(magefilePath, []byte(magefileContent), 0o600)
	if err != nil {
		b.Fatalf("Failed to create test magefile: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.parseMagefile(magefilePath)
		if err != nil {
			b.Fatalf("parseMagefile() failed: %v", err)
		}
	}
}

func BenchmarkExtractDescription(b *testing.B) {
	// Create a comment group
	comments := []*ast.Comment{
		{Text: "// Build builds the project"},
		{Text: "// with advanced features"},
		{Text: "// and optimizations"},
	}
	doc := &ast.CommentGroup{List: comments}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		extractDescription(doc)
	}
}

func TestLoader_CompilerDependency(t *testing.T) {
	// This test verifies that the loader handles cases where Go compiler
	// is not available or compilation fails gracefully

	loader := NewLoader(NewRegistry())

	// Create a valid magefile
	tmpDir := t.TempDir()
	magefilePath := filepath.Join(tmpDir, "magefile.go")

	// Create a magefile with import issues that would fail compilation
	magefileContent := `//go:build mage
package main

import "nonexistent/package"

func Build() error {
	return nil
}
`

	err := os.WriteFile(magefilePath, []byte(magefileContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test magefile: %v", err)
	}

	// This should succeed in parsing but fail in loading (compilation)
	commands, err := loader.parseMagefile(magefilePath)
	if err != nil {
		t.Fatalf("parseMagefile() failed: %v", err)
	}

	if len(commands) == 0 {
		t.Error("Expected at least one command from parsing")
	}

	// Test that DiscoverUserCommands can parse the file structure even with invalid imports
	// (since we only parse AST, import validity doesn't matter)
	commands, err = loader.DiscoverUserCommands(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverUserCommands should succeed even with import issues: %v", err)
	}

	if len(commands) == 0 {
		t.Error("Expected at least one command to be discovered despite import issues")
	}
}

func TestLoader_Integration(t *testing.T) {
	// Integration test that combines parsing and registration
	r := NewRegistry()
	loader := NewLoader(r)

	tmpDir := t.TempDir()
	magefilePath := filepath.Join(tmpDir, "magefile.go")

	// Create a valid, simple magefile that should compile
	magefileContent := `//go:build mage
package main

import "fmt"

// Build builds the project
func Build() error {
	fmt.Println("Building...")
	return nil
}

// Test runs tests
func Test() error {
	fmt.Println("Testing...")
	return nil
}
`

	err := os.WriteFile(magefilePath, []byte(magefileContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test magefile: %v", err)
	}

	// Test parsing works
	commands, err := loader.parseMagefile(magefilePath)
	if err != nil {
		t.Fatalf("parseMagefile() failed: %v", err)
	}

	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands))
	}

	// Verify command details
	for _, cmd := range commands {
		switch cmd.Name {
		case "Build":
			if cmd.Description != buildCmdName+" builds the project" {
				t.Errorf("%s description = %q, expected '%s builds the project'", buildCmdName, cmd.Description, buildCmdName)
			}
		case "Test":
			if cmd.Description != "Test runs tests" {
				t.Errorf("Test description = %q, expected 'Test runs tests'", cmd.Description)
			}
		default:
			t.Errorf("Unexpected command: %s", cmd.Name)
		}
	}

	// Note: We skip the actual LoadUserMagefile test here as it requires
	// Go toolchain and plugin compilation which may not be available
	// in all test environments. The parsing tests above cover the core
	// functionality.
}

func TestLoader_DiscoverUserCommands_NotFound(t *testing.T) {
	r := NewRegistry()
	loader := NewLoader(r)

	// Test with a directory that doesn't have magefile.go
	tmpDir := t.TempDir()
	commands, err := loader.DiscoverUserCommands(tmpDir)
	if err != nil {
		t.Errorf("DiscoverUserCommands() should not error when no magefile found: %v", err)
	}
	if len(commands) > 0 {
		t.Errorf("DiscoverUserCommands() should return empty slice when no magefile found, got %d commands", len(commands))
	}
}

func TestLoader_DiscoverUserCommands_InvalidGo(t *testing.T) {
	r := NewRegistry()
	loader := NewLoader(r)

	// Create a directory with invalid Go code
	tmpDir := t.TempDir()
	magefilePath := filepath.Join(tmpDir, "magefile.go")

	invalidGo := `package main
	this is not valid go syntax
	`

	err := os.WriteFile(magefilePath, []byte(invalidGo), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test magefile: %v", err)
	}

	commands, err := loader.DiscoverUserCommands(tmpDir)
	if err == nil {
		t.Error("DiscoverUserCommands() should error with invalid Go syntax")
	}
	if !strings.Contains(err.Error(), "failed to parse magefile") {
		t.Errorf("Expected parse error, got: %v", err)
	}
	if commands != nil {
		t.Error("DiscoverUserCommands() should return nil commands on parse error")
	}
}

func TestLoader_DiscoverUserCommands_Success(t *testing.T) {
	loader := NewLoader(NewRegistry())

	// Create a temporary magefile
	tmpDir := t.TempDir()
	magefilePath := filepath.Join(tmpDir, "magefile.go")

	magefileContent := `//go:build mage
package main

// Build builds the project
func Build() error {
	return nil
}

// Test runs the tests
func Test() error {
	return nil
}
`

	err := os.WriteFile(magefilePath, []byte(magefileContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test magefile: %v", err)
	}

	commands, err := loader.DiscoverUserCommands(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverUserCommands() failed: %v", err)
	}

	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands))
	}

	// Check for specific commands
	var foundBuild, foundTest bool
	for _, cmd := range commands {
		switch cmd.Name {
		case "Build":
			foundBuild = true
			if cmd.IsNamespace {
				t.Error("Build should not be a namespace command")
			}
			if cmd.Description != "Build builds the project" {
				t.Errorf("Build description = %q, expected 'Build builds the project'", cmd.Description)
			}
		case "Test":
			foundTest = true
			if cmd.IsNamespace {
				t.Error("Test should not be a namespace command")
			}
			if cmd.Description != "Test runs the tests" {
				t.Errorf("Test description = %q, expected 'Test runs the tests'", cmd.Description)
			}
		}
	}

	if !foundBuild {
		t.Error("Build command not found")
	}
	if !foundTest {
		t.Error("Test command not found")
	}
}

func TestLoader_DiscoverUserCommands_Namespaces(t *testing.T) {
	loader := NewLoader(NewRegistry())

	// Create a temporary magefile with namespace methods
	tmpDir := t.TempDir()
	magefilePath := filepath.Join(tmpDir, "magefile.go")

	magefileContent := `//go:build mage
package main

// Build namespace for advanced builds
type Build struct{}

// Linux builds for Linux
func (Build) Linux() error {
	return nil
}

// Windows builds for Windows
func (Build) Windows() error {
	return nil
}

// Named receiver test
func (b Build) Darwin() error {
	return nil
}

// Pointer receiver test
func (b *Build) FreeBSD() error {
	return nil
}
`

	err := os.WriteFile(magefilePath, []byte(magefileContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test magefile: %v", err)
	}

	commands, err := loader.DiscoverUserCommands(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverUserCommands() failed: %v", err)
	}

	// Should find 5 commands: Build type + 4 methods
	if len(commands) != 5 {
		t.Errorf("Expected 5 commands, got %d", len(commands))
	}

	// Check namespace methods
	var foundLinux, foundWindows, foundDarwin, foundFreeBSD bool
	for _, cmd := range commands {
		if cmd.IsNamespace && cmd.Namespace == buildCmdName {
			switch cmd.Method {
			case "Linux":
				foundLinux = true
			case "Windows":
				foundWindows = true
			case "Darwin":
				foundDarwin = true
			case "FreeBSD":
				foundFreeBSD = true
			}
		}
	}

	if !foundLinux {
		t.Error("Linux method not found")
	}
	if !foundWindows {
		t.Error("Windows method not found")
	}
	if !foundDarwin {
		t.Error("Darwin method not found")
	}
	if !foundFreeBSD {
		t.Error("FreeBSD method not found")
	}
}

func TestLoader_DiscoverUserCommands_IgnoresUnexported(t *testing.T) {
	loader := NewLoader(NewRegistry())

	// Create a temporary magefile with unexported functions
	tmpDir := t.TempDir()
	magefilePath := filepath.Join(tmpDir, "magefile.go")

	magefileContent := `//go:build mage
package main

import "github.com/magefile/mage/mg"

// Build builds the project (exported)
func Build() error {
	return nil
}

// helper is an unexported function (should be ignored)
func helper() error {
	return nil
}

// internalBuild is unexported (should be ignored)
func internalBuild() error {
	return nil
}

// Namespace type alias (should be ignored)
type NS = mg.Namespace

// Custom namespace (should be included)
type Deploy struct{}

// Deploy method (should be included)
func (Deploy) Production() error {
	return nil
}
`

	err := os.WriteFile(magefilePath, []byte(magefileContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create test magefile: %v", err)
	}

	commands, err := loader.DiscoverUserCommands(tmpDir)
	if err != nil {
		t.Fatalf("DiscoverUserCommands() failed: %v", err)
	}

	// Should find: Build function + Deploy type + Production method = 3 commands
	if len(commands) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(commands))
		for _, cmd := range commands {
			t.Logf("Found command: %s (namespace: %v)", cmd.Name, cmd.IsNamespace)
		}
	}

	// Verify unexported functions are not included
	for _, cmd := range commands {
		if cmd.Name == "helper" || cmd.Name == "internalBuild" || cmd.Name == "NS" {
			t.Errorf("Unexported function/type %s should not be included", cmd.Name)
		}
	}

	// Verify exported functions are included
	var foundBuild, foundDeploy, foundProduction bool
	for _, cmd := range commands {
		switch {
		case cmd.Name == "Build" && !cmd.IsNamespace:
			foundBuild = true
		case cmd.Name == "Deploy" && cmd.IsNamespace:
			foundDeploy = true
		case cmd.Method == "Production" && cmd.IsNamespace:
			foundProduction = true
		}
	}

	if !foundBuild {
		t.Error("Build function should be included")
	}
	if !foundDeploy {
		t.Error("Deploy type should be included")
	}
	if !foundProduction {
		t.Error("Production method should be included")
	}
}
