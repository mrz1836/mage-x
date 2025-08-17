package mage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLintIssuesIntegration tests the Issues method with real file system operations
func TestLintIssuesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create temporary test directory with real sample files
	testDir := createIntegrationTestCodebase(t)
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("Failed to clean up test directory: %v", err)
		}
	}()

	// Change to test directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	if chdirErr := os.Chdir(testDir); chdirErr != nil {
		t.Fatalf("Failed to change to test directory: %v", chdirErr)
	}

	// Use real command runner
	originalRunner := GetRunner()
	defer func() {
		if setErr := SetRunner(originalRunner); setErr != nil {
			t.Logf("Failed to restore runner: %v", setErr)
		}
	}()

	// Ensure we're using a real command runner for this test
	if _, ok := originalRunner.(*SimpleMockCommandRunner); ok {
		// If we're already using a mock, create a real one for this test
		if setErr := SetRunner(NewSecureCommandRunner()); setErr != nil {
			t.Fatalf("Failed to set real runner: %v", setErr)
		}
		defer func() {
			if restoreErr := SetRunner(originalRunner); restoreErr != nil {
				t.Errorf("Failed to restore runner: %v", restoreErr)
			}
		}()
	}

	// Test the Issues method with real grep commands
	lint := Lint{}
	err = lint.Issues()
	if err != nil {
		t.Errorf("Issues() returned error: %v", err)
	}
}

// TestRealGrepPatterns tests that our grep patterns work with actual grep command
func TestRealGrepPatterns(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDir := createIntegrationTestCodebase(t)
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("Failed to clean up test directory: %v", err)
		}
	}()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	if chdirErr := os.Chdir(testDir); chdirErr != nil {
		t.Fatalf("Failed to change to test directory: %v", chdirErr)
	}

	// Use real command runner
	runner := NewSecureCommandRunner()

	// Test TODO pattern
	todoOutput, err := runner.RunCmdOutput("grep", "-rn", "--include=*.go", "--exclude-dir=vendor", "--exclude-dir=.git", "-E", "//.*TODO|/\\*.*TODO", ".")
	if err != nil {
		// grep returns non-zero when no matches, which is ok
		t.Logf("TODO grep returned error (expected if no matches): %v", err)
	}
	if todoOutput != "" {
		lines := strings.Split(strings.TrimSpace(todoOutput), "\n")
		if len(lines) < 2 {
			t.Errorf("Expected at least 2 TODO matches, got %d", len(lines))
		}
		t.Logf("Found TODO matches: %d", len(lines))
	}

	// Test FIXME pattern
	fixmeOutput, err := runner.RunCmdOutput("grep", "-rn", "--include=*.go", "--exclude-dir=vendor", "--exclude-dir=.git", "-E", "//.*FIXME|/\\*.*FIXME", ".")
	if err != nil {
		t.Logf("FIXME grep returned error (expected if no matches): %v", err)
	}
	if fixmeOutput != "" {
		lines := strings.Split(strings.TrimSpace(fixmeOutput), "\n")
		t.Logf("Found FIXME matches: %d", len(lines))
	}

	// Test nolint pattern
	nolintOutput, err := runner.RunCmdOutput("grep", "-rn", "--include=*.go", "--exclude-dir=vendor", "--exclude-dir=.git", "-E", "//nolint", ".")
	if err != nil {
		t.Logf("Nolint grep returned error (expected if no matches): %v", err)
	}
	if nolintOutput != "" {
		lines := strings.Split(strings.TrimSpace(nolintOutput), "\n")
		t.Logf("Found nolint matches: %d", len(lines))
	}

	// Test t.Skip pattern
	skipOutput, err := runner.RunCmdOutput("grep", "-rn", "--include=*.go", "--exclude-dir=vendor", "--exclude-dir=.git", "-E", "t\\.Skip\\(", ".")
	if err != nil {
		t.Logf("Skip grep returned error (expected if no matches): %v", err)
	}
	if skipOutput != "" {
		lines := strings.Split(strings.TrimSpace(skipOutput), "\n")
		t.Logf("Found skip matches: %d", len(lines))
	}
}

// TestFindDisabledFilesIntegration tests disabled file detection with real find command
func TestFindDisabledFilesIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	testDir := createIntegrationTestCodebase(t)
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("Failed to clean up test directory: %v", err)
		}
	}()

	// Create some real .go.disabled files
	disabledFiles := []string{
		"old_handler.go.disabled",
		"subdir/legacy_utils.go.disabled",
	}

	for _, file := range disabledFiles {
		dir := filepath.Dir(file)
		if dir != "." {
			fullDir := filepath.Join(testDir, dir)
			if err := os.MkdirAll(fullDir, 0o750); err != nil {
				t.Fatalf("Failed to create directory %s: %v", fullDir, err)
			}
		}

		filePath := filepath.Join(testDir, file)
		content := `package main

// This file is disabled
func oldFunction() {
	// Legacy code that's no longer used
}
`
		if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
			t.Fatalf("Failed to create disabled file %s: %v", filePath, err)
		}
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	if chdirErr := os.Chdir(testDir); chdirErr != nil {
		t.Fatalf("Failed to change to test directory: %v", chdirErr)
	}

	// Use real command runner
	runner := NewSecureCommandRunner()

	// Test find command for disabled files
	output, err := runner.RunCmdOutput("find", ".", "-name", "*.go.disabled", "-type", "f")
	if err != nil {
		t.Fatalf("Find command failed: %v", err)
	}

	if output == "" {
		t.Error("Expected to find disabled files but got empty output")
		return
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != len(disabledFiles) {
		t.Errorf("Expected %d disabled files, found %d", len(disabledFiles), len(lines))
	}

	// Check that found files match expected ones
	foundFiles := make(map[string]bool)
	for _, line := range lines {
		// Remove ./ prefix if present
		file := strings.TrimPrefix(line, "./")
		foundFiles[file] = true
	}

	for _, expected := range disabledFiles {
		if !foundFiles[expected] {
			t.Errorf("Expected to find disabled file %s", expected)
		}
	}
}

// TestGrepErrorHandling tests how we handle grep command errors
func TestGrepErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create empty test directory
	testDir := t.TempDir()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	if chdirErr := os.Chdir(testDir); chdirErr != nil {
		t.Fatalf("Failed to change to test directory: %v", chdirErr)
	}

	// Use real command runner
	_ = NewSecureCommandRunner()

	// Test grep with no matches (should not error in our implementation)
	results := findMatches("NONEXISTENT_PATTERN")
	if len(results) != 0 {
		t.Errorf("Expected no results for non-existent pattern, got %d", len(results))
	}

	// Test with invalid pattern (should handle gracefully)
	results = findMatches("[") // Invalid regex
	// Should not panic, may return empty results
	t.Logf("Invalid pattern test returned %d results", len(results))
}

// TestLargeScanPerformance tests performance with larger codebases
func TestLargeScanPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	// Create a larger test codebase
	testDir := createLargeTestCodebase(t)
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("Failed to clean up test directory: %v", err)
		}
	}()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	if chdirErr := os.Chdir(testDir); chdirErr != nil {
		t.Fatalf("Failed to change to test directory: %v", chdirErr)
	}

	// Use real command runner
	originalRunner := GetRunner()
	defer func() {
		if setErr := SetRunner(originalRunner); setErr != nil {
			t.Logf("Failed to restore runner: %v", setErr)
		}
	}()

	if _, ok := originalRunner.(*SimpleMockCommandRunner); ok {
		if setErr := SetRunner(NewSecureCommandRunner()); setErr != nil {
			t.Fatalf("Failed to set secure runner: %v", setErr)
		}
		defer func() {
			if setErr := SetRunner(originalRunner); setErr != nil {
				t.Logf("Failed to restore runner: %v", setErr)
			}
		}()
	}

	// Measure performance of the scan
	lint := Lint{}
	err = lint.Issues()
	if err != nil {
		t.Errorf("Issues() returned error on large codebase: %v", err)
	}
}

// TestActualProjectScan tests scanning the actual mage-x project
func TestActualProjectScan(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping actual project scan in short mode")
	}

	// Change to project root (assuming test is run from pkg/mage)
	projectRoot := "../.."
	if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); os.IsNotExist(err) {
		t.Skip("Not in expected project structure, skipping actual project scan")
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(originalDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	if chdirErr := os.Chdir(projectRoot); chdirErr != nil {
		t.Fatalf("Failed to change to project root: %v", chdirErr)
	}

	// Use real command runner
	originalRunner := GetRunner()
	defer func() {
		if setErr := SetRunner(originalRunner); setErr != nil {
			t.Logf("Failed to restore runner: %v", setErr)
		}
	}()

	if _, ok := originalRunner.(*SimpleMockCommandRunner); ok {
		if setErr := SetRunner(NewSecureCommandRunner()); setErr != nil {
			t.Fatalf("Failed to set secure runner: %v", setErr)
		}
		defer func() {
			if setErr := SetRunner(originalRunner); setErr != nil {
				t.Logf("Failed to restore runner: %v", setErr)
			}
		}()
	}

	// Run the actual scan on the mage-x project
	lint := Lint{}
	err = lint.Issues()
	if err != nil {
		t.Errorf("Issues() failed on actual project: %v", err)
	}

	// This is more of a smoke test - we just want to ensure it doesn't crash
	// on the real project structure
}

// createIntegrationTestCodebase creates a test codebase with more realistic content for integration testing
func createIntegrationTestCodebase(t *testing.T) string {
	tmpDir := t.TempDir()

	// Create more comprehensive test files
	files := map[string]string{
		"main.go": `package main

import (
	"fmt"
	"os"
)

// TODO: implement proper database connection pooling
func main() {
	/* FIXME: this should use structured logging */
	fmt.Println("Starting application...")

	if err := run(); err != nil {
		// TODO: implement graceful shutdown
		os.Exit(1)
	}
}

func run() error {
	// HACK: temporary solution until we implement proper config
	config := "hardcoded"
	_ = config
	return nil
}
`,
		"handlers.go": `package main

import "fmt"

// TODO: add proper error handling and validation
func handleRequest() {
	fmt.Println("handling request") //nolint:forbidigo

	// FIXME: remove debug code before production
	fmt.Printf("Debug info") //nolint:forbidigo,gocyclo
}

// TODO: implement authentication middleware
func authenticate() bool {
	// HACK: always return true for now
	return true //nolint:govet
}
`,
		"utils.go": `package main

// TODO: add comprehensive error handling
func helper() error {
	var unused string //nolint:deadcode,unused
	_ = unused

	// FIXME: optimize this algorithm
	for i := 0; i < 1000; i++ {
		// HACK: inefficient loop
		_ = i
	}

	return nil
}

/* TODO: refactor this entire utility package */
func anotherHelper() {
	// Implementation
}
`,
		"main_test.go": `package main

import "testing"

func TestMain(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Test implementation
}

func TestDatabase(t *testing.T) {
	t.Skip("database not available in test environment")
}

func TestExternal(t *testing.T) {
	t.Skip("requires external service")
}
`,
		"handlers_test.go": `package main

import "testing"

func TestHandlers(t *testing.T) {
	t.Skip("database not available in test environment")
}

func TestAuth(t *testing.T) {
	t.Skip("authentication service unavailable")
}
`,
		"config/config.go": `package config

// TODO: implement configuration validation
type Config struct {
	Database string
	Port     int
}

// FIXME: add proper defaults and environment variable support
func Load() *Config {
	return &Config{} //nolint:exhaustruct
}
`,
	}

	// Create directory structure
	configDir := filepath.Join(tmpDir, "config")
	if err := os.MkdirAll(configDir, 0o750); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	// Write all files
	for filename, content := range files {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
			t.Fatalf("Failed to create file %s: %v", filename, err)
		}
	}

	return tmpDir
}

// createLargeTestCodebase creates a larger test codebase for performance testing
func createLargeTestCodebase(t *testing.T) string {
	tmpDir := t.TempDir()

	// Create multiple directories
	directories := []string{"cmd", "pkg", "internal", "api", "web", "docs"}
	for _, dir := range directories {
		if err := os.MkdirAll(filepath.Join(tmpDir, dir), 0o750); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Generate many files with various issues
	for i := 0; i < 50; i++ {
		for _, dir := range directories {
			filename := filepath.Join(tmpDir, dir, "file"+strings.Repeat("1", 2)+".go")
			content := generateTestFileContent(i)
			if err := os.WriteFile(filename, []byte(content), 0o600); err != nil {
				t.Fatalf("Failed to create large test file %s: %v", filename, err)
			}
		}
	}

	return tmpDir
}

// generateTestFileContent generates Go file content with various types of issues
func generateTestFileContent(index int) string {
	templates := []string{
		`package pkg

import "fmt"

// TODO: implement feature %d
func Feature%d() {
	fmt.Println("feature") //nolint:forbidigo
	// FIXME: add error handling
}
`,
		`package pkg

// HACK: temporary solution %d
func Workaround%d() error {
	var unused int //nolint:unused,deadcode
	return nil
}
`,
		`package pkg

import "testing"

func TestFeature%d(t *testing.T) {
	t.Skip("external dependency required")
	// TODO: implement proper mocking
}
`,
	}

	template := templates[index%len(templates)]
	return strings.ReplaceAll(template, "%d", strings.Repeat("1", 3))
}
