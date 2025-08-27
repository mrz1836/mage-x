package mage

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestLintIssues tests the main Issues method
func TestLintIssues(t *testing.T) {
	// Create temporary test directory with sample files
	testDir := createTestCodebase(t)
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Errorf("Failed to remove test directory: %v", err)
		}
	}()

	// Change to test directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if restoreErr := os.Chdir(originalDir); restoreErr != nil {
			t.Errorf("Failed to restore original directory: %v", restoreErr)
		}
	}()

	if chdirErr := os.Chdir(testDir); chdirErr != nil {
		t.Fatalf("Failed to change to test directory: %v", chdirErr)
	}

	// Mock the command runner to return our test data
	originalRunner := GetRunner()
	defer func() {
		if restoreErr := SetRunner(originalRunner); restoreErr != nil {
			t.Errorf("Failed to restore original runner: %v", restoreErr)
		}
	}()

	mockRunner := &SimpleMockCommandRunner{
		outputs: make(map[string]string),
	}
	if setErr := SetRunner(mockRunner); setErr != nil {
		t.Fatalf("Failed to set mock runner: %v", setErr)
	}

	// Test the Issues method
	lint := Lint{}
	err = lint.Issues()
	if err != nil {
		t.Errorf("Issues() returned error: %v", err)
	}
}

// TestScanForComments tests the comment scanning functionality
func TestScanForComments(t *testing.T) {
	testDir := createTestCodebase(t)
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
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	// Mock the grep output
	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore runner: %v", err)
		}
	}()

	mockRunner := &SimpleMockCommandRunner{
		outputs: map[string]string{
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //.*TODO|/\\*.*TODO .": `main.go:5:	// TODO: implement database connection
utils.go:10:	// TODO: add error handling
handlers.go:15:	/* TODO: refactor this function */`,
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //.*FIXME|/\\*.*FIXME .": `main.go:8:	// FIXME: memory leak in this function
utils.go:20:	// FIXME: optimize performance`,
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //.*HACK|/\\*.*HACK .": `handlers.go:25:	// HACK: temporary workaround`,
		},
	}
	if err := SetRunner(mockRunner); err != nil {
		t.Fatalf("Failed to set mock runner: %v", err)
	}

	results := scanForComments()

	// Check TODO results
	todoItems, todoExists := results["TODO"]
	if !todoExists {
		t.Error("Expected TODO results but got none")
		return
	}
	if len(todoItems) == 0 {
		t.Error("Expected TODO items but got none")
		return
	}

	// Should have grouped by message
	found := false
	for _, item := range todoItems {
		if strings.Contains(item.Message, "implement database connection") {
			found = true
			if item.Count != 1 {
				t.Errorf("Expected count 1 for 'implement database connection', got %d", item.Count)
			}
		}
	}
	if !found {
		t.Error("Expected to find 'implement database connection' TODO")
	}

	// Check FIXME results
	if fixmeItems, exists := results["FIXME"]; exists {
		if len(fixmeItems) == 0 {
			t.Error("Expected FIXME items but got none")
		}
	} else {
		t.Error("Expected FIXME results but got none")
	}

	// Check HACK results
	if hackItems, exists := results["HACK"]; exists {
		if len(hackItems) == 0 {
			t.Error("Expected HACK items but got none")
		}
	} else {
		t.Error("Expected HACK results but got none")
	}
}

// TestScanForNolintDirectives tests nolint directive scanning
func TestScanForNolintDirectives(t *testing.T) {
	testDir := createTestCodebase(t)
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
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore runner: %v", err)
		}
	}()

	mockRunner := &SimpleMockCommandRunner{
		outputs: map[string]string{
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //nolint .": `main.go:12:	func badCode() { //nolint:gocyclo,unused
utils.go:25:	var unused string //nolint:deadcode
handlers.go:30:	fmt.Println("debug") //nolint:forbidigo,gocyclo // Test data: intentional linting issues`,
		},
	}
	if err := SetRunner(mockRunner); err != nil {
		t.Fatalf("Failed to set mock runner: %v", err)
	}

	results := scanForNolintDirectives()

	if nolintItems, exists := results["NOLINT"]; exists {
		if len(nolintItems) == 0 {
			t.Error("Expected NOLINT items but got none")
		}

		// Check for specific tags
		tagCounts := make(map[string]int)
		for _, item := range nolintItems {
			tagCounts[item.Message] = item.Count
		}

		expectedTags := []string{"gocyclo", "unused", "deadcode", "forbidigo"}
		for _, tag := range expectedTags {
			if count, exists := tagCounts[tag]; !exists {
				t.Errorf("Expected to find nolint tag '%s' but didn't", tag)
			} else if count == 0 {
				t.Errorf("Expected count > 0 for tag '%s', got %d", tag, count)
			}
		}
	} else {
		t.Error("Expected NOLINT results but got none")
	}
}

// TestScanForTestSkips tests test skip scanning
func TestScanForTestSkips(t *testing.T) {
	testDir := createTestCodebase(t)
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
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore runner: %v", err)
		}
	}()

	mockRunner := &SimpleMockCommandRunner{
		outputs: map[string]string{
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E t\\.Skip\\( .": `main_test.go:20:	t.Skip("database not available")
utils_test.go:35:	t.Skip("requires external service")
handlers_test.go:40:	t.Skip("database not available")`,
		},
	}
	if err := SetRunner(mockRunner); err != nil {
		t.Fatalf("Failed to set mock runner: %v", err)
	}

	results := scanForTestSkips()

	skipItems, skipExists := results["TEST_SKIP"]
	if !skipExists {
		t.Error("Expected TEST_SKIP results but got none")
		return
	}
	if len(skipItems) == 0 {
		t.Error("Expected TEST_SKIP items but got none")
		return
	}

	// Check for specific skip messages
	messageCounts := make(map[string]int)
	for _, item := range skipItems {
		messageCounts[item.Message] = item.Count
	}

	if count, exists := messageCounts["database not available"]; !exists {
		t.Error("Expected to find 'database not available' skip message")
	} else if count != 2 {
		t.Errorf("Expected count 2 for 'database not available', got %d", count)
	}

	if count, exists := messageCounts["requires external service"]; !exists {
		t.Error("Expected to find 'requires external service' skip message")
	} else if count != 1 {
		t.Errorf("Expected count 1 for 'requires external service', got %d", count)
	}
}

// TestScanForDisabledFiles tests disabled file scanning
func TestScanForDisabledFiles(t *testing.T) {
	testDir := createTestCodebase(t)
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			t.Logf("Failed to clean up test directory: %v", err)
		}
	}()

	// Create some .go.disabled files
	disabledFiles := []string{
		"old_handler.go.disabled",
		"legacy/old_utils.go.disabled",
		"deprecated/auth.go.disabled",
	}

	for _, file := range disabledFiles {
		dir := filepath.Dir(file)
		if dir != "." {
			if err := os.MkdirAll(filepath.Join(testDir, dir), 0o750); err != nil {
				t.Fatalf("Failed to create directory %s: %v", dir, err)
			}
		}
		filePath := filepath.Join(testDir, file)
		if err := os.WriteFile(filePath, []byte("// disabled code"), 0o600); err != nil {
			t.Fatalf("Failed to create disabled file %s: %v", filePath, err)
		}
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore runner: %v", err)
		}
	}()

	mockRunner := &SimpleMockCommandRunner{
		outputs: map[string]string{
			"find . -name *.go.disabled -type f": `./old_handler.go.disabled
./legacy/old_utils.go.disabled
./deprecated/auth.go.disabled`,
		},
	}
	if err := SetRunner(mockRunner); err != nil {
		t.Fatalf("Failed to set mock runner: %v", err)
	}

	results := scanForDisabledFiles()

	if len(results) != 3 {
		t.Errorf("Expected 3 disabled files, got %d", len(results))
	}

	expectedFiles := []string{
		"old_handler.go.disabled",
		"legacy/old_utils.go.disabled",
		"deprecated/auth.go.disabled",
	}

	for _, expected := range expectedFiles {
		found := false
		for _, result := range results {
			if result == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected to find disabled file '%s' but didn't", expected)
		}
	}
}

// TestGroupByMessage tests the message grouping logic
func TestGroupByMessage(t *testing.T) {
	matches := []string{
		"main.go:5:	// TODO: implement database connection",
		"utils.go:10:	// TODO: add error handling",
		"handlers.go:15:	// TODO: implement database connection", // duplicate message
		"service.go:20:	// TODO: optimize performance",
	}

	results := groupByMessage(matches, "TODO")

	if len(results) != 3 {
		t.Errorf("Expected 3 unique messages, got %d", len(results))
	}

	// Check for the duplicate message grouping
	found := false
	for _, result := range results {
		if strings.Contains(result.Message, "implement database connection") {
			found = true
			if result.Count != 2 {
				t.Errorf("Expected count 2 for 'implement database connection', got %d", result.Count)
			}
			if len(result.Files) != 2 {
				t.Errorf("Expected 2 files for 'implement database connection', got %d", len(result.Files))
			}
		}
	}
	if !found {
		t.Error("Expected to find grouped 'implement database connection' messages")
	}
}

// TestGroupNolintByTag tests nolint tag grouping
func TestGroupNolintByTag(t *testing.T) {
	matches := []string{
		"main.go:12:	func badCode() { //nolint:gocyclo,unused",
		"utils.go:25:	var unused string //nolint:deadcode",
		"handlers.go:30:	fmt.Println(debug) //nolint:gocyclo,forbidigo", // gocyclo appears again - test data
	}

	results := groupNolintByTag(matches)

	// Should have 4 unique tags: gocyclo (appears twice), unused, deadcode, forbidigo
	if len(results) != 4 {
		t.Errorf("Expected 4 unique tags, got %d", len(results))
	}

	// Check for gocyclo appearing twice
	found := false
	for _, result := range results {
		if result.Message == "gocyclo" {
			found = true
			if result.Count != 2 {
				t.Errorf("Expected count 2 for 'gocyclo', got %d", result.Count)
			}
		}
	}
	if !found {
		t.Error("Expected to find 'gocyclo' tag with count 2")
	}
}

// TestGroupSkipsByMessage tests test skip message grouping
func TestGroupSkipsByMessage(t *testing.T) {
	matches := []string{
		`main_test.go:20:	t.Skip("database not available")`,
		`utils_test.go:35:	t.Skip("requires external service")`,
		`handlers_test.go:40:	t.Skip("database not available")`, // duplicate message
	}

	results := groupSkipsByMessage(matches)

	if len(results) != 2 {
		t.Errorf("Expected 2 unique skip messages, got %d", len(results))
	}

	// Check for the duplicate message grouping
	found := false
	for _, result := range results {
		if result.Message == "database not available" {
			found = true
			if result.Count != 2 {
				t.Errorf("Expected count 2 for 'database not available', got %d", result.Count)
			}
		}
	}
	if !found {
		t.Error("Expected to find 'database not available' message with count 2")
	}
}

// TestCountTotalIssues tests the total counting functionality
func TestCountTotalIssues(t *testing.T) {
	issues := map[string][]IssueCount{
		"TODO": {
			{Message: "message1", Count: 3},
			{Message: "message2", Count: 2},
		},
		"FIXME": {
			{Message: "message3", Count: 1},
		},
	}

	total := countTotalIssues(issues)
	expected := 6 // 3 + 2 + 1
	if total != expected {
		t.Errorf("Expected total %d, got %d", expected, total)
	}
}

// TestPluralize tests the pluralization helper
func TestPluralize(t *testing.T) {
	tests := []struct {
		count    int
		expected string
	}{
		{0, "s"},
		{1, ""},
		{2, "s"},
		{10, "s"},
	}

	for _, test := range tests {
		result := pluralize(test.count)
		if result != test.expected {
			t.Errorf("pluralize(%d) = %q, expected %q", test.count, result, test.expected)
		}
	}
}

// TestFindMatches tests the grep pattern matching
func TestFindMatches(t *testing.T) {
	testDir := createTestCodebase(t)
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
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(testDir); err != nil {
		t.Fatalf("Failed to change to test directory: %v", err)
	}

	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			t.Logf("Failed to restore runner: %v", err)
		}
	}()

	mockRunner := &SimpleMockCommandRunner{
		outputs: map[string]string{
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //.*TODO|/\\*.*TODO .": `main.go:5:	// TODO: test message
utils.go:10:	/* TODO: another message */`,
		},
	}
	if err := SetRunner(mockRunner); err != nil {
		t.Fatalf("Failed to set mock runner: %v", err)
	}

	results := findMatches("TODO")

	if len(results) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(results))
	}

	if !strings.Contains(results[0], "main.go:5") {
		t.Errorf("Expected first result to contain 'main.go:5', got %q", results[0])
	}

	if !strings.Contains(results[1], "utils.go:10") {
		t.Errorf("Expected second result to contain 'utils.go:10', got %q", results[1])
	}
}

// TestDisplayIssueResults tests the display formatting (this would normally print to stdout)
func TestDisplayIssueResults(t *testing.T) {
	todoIssues := map[string][]IssueCount{
		"TODO": {
			{Message: "implement feature", Count: 3, Files: []string{"a.go", "b.go", "c.go"}},
		},
	}

	nolintIssues := map[string][]IssueCount{
		"NOLINT": {
			{Message: "gocyclo", Count: 5, Files: []string{"x.go", "y.go"}},
		},
	}

	skipIssues := map[string][]IssueCount{
		"TEST_SKIP": {
			{Message: "database not available", Count: 2, Files: []string{"test1.go", "test2.go"}},
		},
	}

	disabledFiles := []string{"old.go.disabled", "legacy.go.disabled"}

	// This function primarily prints to stdout, so we're mainly testing it doesn't panic
	displayIssueResults(todoIssues, nolintIssues, skipIssues, disabledFiles)
}

// BenchmarkIssuesScanning benchmarks the Issues method performance
func BenchmarkIssuesScanning(b *testing.B) {
	testDir := createTestCodebase(b)
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			b.Logf("Failed to clean up test directory: %v", err)
		}
	}()

	originalDir, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			b.Logf("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(testDir); err != nil {
		b.Fatalf("Failed to change to test directory: %v", err)
	}

	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			b.Logf("Failed to restore runner: %v", err)
		}
	}()

	mockRunner := &SimpleMockCommandRunner{
		outputs: map[string]string{
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //.*TODO|/\\*.*TODO .":   generateLargeMockOutput("TODO", 1000),
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //.*FIXME|/\\*.*FIXME .": generateLargeMockOutput("FIXME", 500),
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //.*HACK|/\\*.*HACK .":   generateLargeMockOutput("HACK", 100),
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //nolint .":              generateNolintMockOutput(2000),
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E t\\.Skip\\( .":           generateSkipMockOutput(300),
			"find . -name *.go.disabled -type f":                                                         generateDisabledFilesMock(50),
		},
	}
	if err := SetRunner(mockRunner); err != nil {
		b.Fatalf("Failed to set mock runner: %v", err)
	}

	lint := Lint{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := lint.Issues()
		if err != nil {
			b.Fatalf("Issues() returned error: %v", err)
		}
	}
}

// BenchmarkScanForComments benchmarks comment scanning
func BenchmarkScanForComments(b *testing.B) {
	testDir := createTestCodebase(b)
	defer func() {
		if err := os.RemoveAll(testDir); err != nil {
			b.Logf("Failed to clean up test directory: %v", err)
		}
	}()

	originalDir, err := os.Getwd()
	if err != nil {
		b.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			b.Logf("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(testDir); err != nil {
		b.Fatalf("Failed to change to test directory: %v", err)
	}

	originalRunner := GetRunner()
	defer func() {
		if err := SetRunner(originalRunner); err != nil {
			b.Logf("Failed to restore runner: %v", err)
		}
	}()

	mockRunner := &SimpleMockCommandRunner{
		outputs: map[string]string{
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //.*TODO|/\\*.*TODO .":   generateLargeMockOutput("TODO", 1000),
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //.*FIXME|/\\*.*FIXME .": generateLargeMockOutput("FIXME", 500),
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //.*HACK|/\\*.*HACK .":   generateLargeMockOutput("HACK", 100),
		},
	}
	if err := SetRunner(mockRunner); err != nil {
		b.Fatalf("Failed to set mock runner: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = scanForComments()
	}
}

// Helper functions for testing

// SimpleMockCommandRunner implements CommandRunner for testing
type SimpleMockCommandRunner struct {
	outputs map[string]string
	errors  map[string]error
}

func (m *SimpleMockCommandRunner) RunCmd(name string, args ...string) error {
	cmd := strings.Join(append([]string{name}, args...), " ")
	if err, exists := m.errors[cmd]; exists {
		return err
	}
	return nil
}

func (m *SimpleMockCommandRunner) RunCmdOutput(name string, args ...string) (string, error) {
	cmd := strings.Join(append([]string{name}, args...), " ")
	if err, exists := m.errors[cmd]; exists {
		return "", err
	}
	if output, exists := m.outputs[cmd]; exists {
		return output, nil
	}
	return "", nil
}

func (m *SimpleMockCommandRunner) RunCmdBackground(name string, args ...string) error {
	return m.RunCmd(name, args...)
}

func (m *SimpleMockCommandRunner) RunCmdWithTimeout(timeout time.Duration, name string, args ...string) error {
	return m.RunCmd(name, args...)
}

func (m *SimpleMockCommandRunner) RunCmdOutputWithTimeout(timeout time.Duration, name string, args ...string) (string, error) {
	return m.RunCmdOutput(name, args...)
}

// createTestCodebase creates a temporary directory with sample Go files for testing
func createTestCodebase(t interface{}) string {
	var tmpDir string
	var err error

	switch tb := t.(type) {
	case *testing.T:
		tmpDir = tb.TempDir()
	case *testing.B:
		tmpDir, err = os.MkdirTemp("", "mage_test_*")
		if err != nil {
			tb.Fatalf("Failed to create temp dir: %v", err)
		}
	default:
		panic("unsupported testing type")
	}

	// Create sample Go files with various issues
	files := map[string]string{
		"main.go": `package main

import "fmt"

// TODO: implement database connection
func main() {
	// FIXME: memory leak in this function
	fmt.Println("Hello World")
}
`,
		"utils.go": `package main

// TODO: add error handling
func helper() {
	// Some code
	var unused string //nolint:deadcode
}
`,
		"handlers.go": `package main

// TODO: refactor this function
/* HACK: temporary workaround */
func handler() {
	fmt.Println("debug") //nolint:forbidigo,gocyclo // Test data: intentional linting issues for testing
}
`,
		"main_test.go": `package main

import "testing"

func TestSomething(t *testing.T) {
	t.Skip("database not available")
}
`,
		"utils_test.go": `package main

import "testing"

func TestUtils(t *testing.T) {
	t.Skip("requires external service")
}
`,
	}

	for filename, content := range files {
		filePath := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
			switch tb := t.(type) {
			case *testing.T:
				tb.Fatalf("Failed to create file %s: %v", filename, err)
			case *testing.B:
				tb.Fatalf("Failed to create file %s: %v", filename, err)
			}
		}
	}

	return tmpDir
}

// generateLargeMockOutput generates mock grep output for performance testing
func generateLargeMockOutput(pattern string, count int) string {
	var lines []string
	for i := 0; i < count; i++ {
		line := ""
		switch pattern {
		case "TODO":
			line = "file%d.go:%d:	// TODO: implement feature %d"
		case "FIXME":
			line = "file%d.go:%d:	// FIXME: fix issue %d"
		case "HACK":
			line = "file%d.go:%d:	// HACK: temporary solution %d"
		}
		lines = append(lines, strings.ReplaceAll(line, "%d", strings.Repeat("1", len(line)/10+1)[:3]))
	}
	return strings.Join(lines, "\n")
}

// generateNolintMockOutput generates mock nolint directive output
func generateNolintMockOutput(count int) string {
	var lines []string
	tags := []string{"gocyclo", "unused", "deadcode", "forbidigo", "govet"}
	for i := 0; i < count; i++ {
		tag := tags[i%len(tags)]
		line := strings.ReplaceAll("file%d.go:%d:	code() //nolint:%s", "%d", strings.Repeat("1", 3))
		line = strings.ReplaceAll(line, "%s", tag)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// generateSkipMockOutput generates mock test skip output
func generateSkipMockOutput(count int) string {
	var lines []string
	messages := []string{"database not available", "requires external service", "slow test", "flaky test"}
	for i := 0; i < count; i++ {
		message := messages[i%len(messages)]
		line := strings.ReplaceAll(`file%d_test.go:%d:	t.Skip("%s")`, "%d", strings.Repeat("1", 3))
		line = strings.ReplaceAll(line, "%s", message)
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// generateDisabledFilesMock generates mock disabled files output
func generateDisabledFilesMock(count int) string {
	var lines []string
	for i := 0; i < count; i++ {
		line := strings.ReplaceAll("./file%d.go.disabled", "%d", strings.Repeat("1", 3))
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

// TestIssueCountStruct tests the IssueCount struct
func TestIssueCountStruct(t *testing.T) {
	issue := IssueCount{
		Message: "test message",
		Count:   5,
		Files:   []string{"file1.go", "file2.go"},
	}

	if issue.Message != "test message" {
		t.Errorf("Expected message 'test message', got %q", issue.Message)
	}

	if issue.Count != 5 {
		t.Errorf("Expected count 5, got %d", issue.Count)
	}

	if len(issue.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(issue.Files))
	}

	expectedFiles := []string{"file1.go", "file2.go"}
	if !reflect.DeepEqual(issue.Files, expectedFiles) {
		t.Errorf("Expected files %v, got %v", expectedFiles, issue.Files)
	}
}

// TestEmptyResults tests behavior with no issues found
func TestEmptyResults(t *testing.T) {
	testDir := createTestCodebase(t)
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

	originalRunner := GetRunner()
	defer func() {
		if setErr := SetRunner(originalRunner); setErr != nil {
			t.Logf("Failed to restore runner: %v", setErr)
		}
	}()

	// Mock empty results
	mockRunner := &SimpleMockCommandRunner{
		outputs: map[string]string{
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //.*TODO|/\\*.*TODO .":   "",
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //.*FIXME|/\\*.*FIXME .": "",
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //.*HACK|/\\*.*HACK .":   "",
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E //nolint .":              "",
			"grep -rn --include=*.go --exclude-dir=vendor --exclude-dir=.git -E t\\.Skip\\( .":           "",
			"find . -name *.go.disabled -type f":                                                         "",
		},
	}
	if setErr := SetRunner(mockRunner); setErr != nil {
		t.Fatalf("Failed to set mock runner: %v", setErr)
	}

	// Test individual functions
	todoResults := scanForComments()
	if len(todoResults) != 0 {
		t.Errorf("Expected no TODO results, got %d categories", len(todoResults))
	}

	nolintResults := scanForNolintDirectives()
	if len(nolintResults) != 0 {
		t.Errorf("Expected no nolint results, got %d categories", len(nolintResults))
	}

	skipResults := scanForTestSkips()
	if len(skipResults) != 0 {
		t.Errorf("Expected no skip results, got %d categories", len(skipResults))
	}

	disabledResults := scanForDisabledFiles()
	if len(disabledResults) != 0 {
		t.Errorf("Expected no disabled files, got %d", len(disabledResults))
	}

	// Test main Issues method
	lint := Lint{}
	err = lint.Issues()
	if err != nil {
		t.Errorf("Issues() should not return error with empty results: %v", err)
	}
}
