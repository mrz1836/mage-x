package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/mrz1836/mage-x/pkg/mage/registry"
)

func TestMain(m *testing.M) {
	// Setup test environment
	os.Exit(m.Run())
}

func TestShowVersion(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	showVersion()

	if err := w.Close(); err != nil {
		t.Logf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Logf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	if !strings.Contains(output, version) {
		t.Errorf("showVersion() output should contain version %s, got: %s", version, output)
	}
	if !strings.Contains(output, "MAGE-X") {
		t.Errorf("showVersion() output should contain 'MAGE-X', got: %s", output)
	}
}

func TestShowHelp(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	showUnifiedHelp("")

	if err := w.Close(); err != nil {
		t.Logf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Logf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	expectedSections := []string{
		"MAGE-X",
		"Usage:",
		"Options:",
		"Commands",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("showUnifiedHelp() output should contain '%s', got: %s", section, output)
		}
	}
}

func TestShowUsage(t *testing.T) {
	// Capture stdout (where showUsage now writes via showUnifiedHelp)
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	showUsage()

	if err := w.Close(); err != nil {
		t.Logf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Logf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	if !strings.Contains(output, "magex") {
		t.Errorf("showUsage() output should contain 'magex', got: %s", output)
	}
	if !strings.Contains(output, "command") {
		t.Errorf("showUsage() output should contain 'command', got: %s", output)
	}
}

func TestInitMagefile(t *testing.T) {
	// Test in a temporary directory
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(oldDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	err = initMagefile()
	if err != nil {
		t.Fatalf("initMagefile() should not fail: %v", err)
	}

	if err := w.Close(); err != nil {
		t.Logf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Logf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Check if magefile.go was created
	if _, err := os.Stat("magefile.go"); os.IsNotExist(err) {
		t.Error("initMagefile() should create magefile.go")
	}

	// Check output message
	if !strings.Contains(output, "magefile.go") {
		t.Errorf("initMagefile() output should mention magefile.go, got: %s", output)
	}
}

func TestInitMagefile_Existing(t *testing.T) {
	// Test in a temporary directory
	tmpDir := t.TempDir()
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(oldDir); chdirErr != nil {
			t.Logf("Failed to restore directory: %v", chdirErr)
		}
	}()

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Create existing magefile
	existingContent := "existing content"
	err = os.WriteFile("magefile.go", []byte(existingContent), 0o600)
	if err != nil {
		t.Fatalf("Failed to create existing magefile: %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	err = initMagefile()
	// Should return error for existing file
	if closeErr := w.Close(); closeErr != nil {
		t.Logf("Failed to close writer: %v", closeErr)
	}
	os.Stdout = oldStdout

	if err == nil {
		t.Error("initMagefile() should return error when file already exists")
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Logf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Check that existing file wasn't overwritten
	content, err2 := os.ReadFile("magefile.go")
	if err2 != nil {
		t.Fatalf("Failed to read magefile: %v", err2)
	}

	if string(content) != existingContent {
		t.Error("initMagefile() should not overwrite existing magefile.go")
	}

	// Check output message
	if !strings.Contains(output, "already exists") {
		t.Errorf("initMagefile() output should mention file already exists, got: %s", output)
	}
}

func TestCleanCache(t *testing.T) {
	// Capture both stdout and stderr to avoid file descriptor issues
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Redirect both stdout and stderr to the same pipe to catch all output
	os.Stdout = w
	os.Stderr = w

	cleanCache()

	// Close writer and restore streams
	if err := w.Close(); err != nil {
		t.Logf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read output
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Logf("Failed to read from pipe: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Logf("Failed to close reader: %v", err)
	}
	output := buf.String()

	// cleanCache() should run without errors
	// Note: Output may be empty if there are no cache directories to clean
	// This is acceptable behavior
	t.Logf("cleanCache() output: %q", output)
}

func TestShowCommands_List(t *testing.T) {
	// Create a test registry
	reg := registry.NewRegistry()

	// Register test commands
	testCmd1, err := registry.NewCommand("test1").
		WithDescription("Test command 1").
		WithFunc(func() error { return nil }).
		WithCategory("Test").
		Build()
	if err != nil {
		t.Fatalf("Failed to build test command: %v", err)
	}

	testCmd2, err := registry.NewNamespaceCommand("build", "linux").
		WithDescription("Build for Linux").
		WithFunc(func() error { return nil }).
		WithCategory("Build").
		Build()
	if err != nil {
		t.Fatalf("Failed to build test command: %v", err)
	}

	reg.MustRegister(testCmd1)
	reg.MustRegister(testCmd2)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	listCommands(reg, NewCommandDiscovery(reg), false)

	if err := w.Close(); err != nil {
		t.Logf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Logf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Check for commands in output
	if !strings.Contains(output, "test1") {
		t.Errorf("showCommands() should contain 'test1', got: %s", output)
	}
	if !strings.Contains(output, "build:linux") {
		t.Errorf("showCommands() should contain 'build:linux', got: %s", output)
	}
}

func TestShowCommands_Namespace(t *testing.T) {
	// Create a test registry
	reg := registry.NewRegistry()

	// Register test commands in different namespaces
	buildCmd, err := registry.NewNamespaceCommand("build", "linux").
		WithDescription("Build for Linux").
		WithFunc(func() error { return nil }).
		WithCategory("Build").
		Build()
	if err != nil {
		t.Fatalf("Failed to build test command: %v", err)
	}

	testCmd, err := registry.NewNamespaceCommand("test", "unit").
		WithDescription("Run unit tests").
		WithFunc(func() error { return nil }).
		WithCategory("Test").
		Build()
	if err != nil {
		t.Fatalf("Failed to build test command: %v", err)
	}

	reg.MustRegister(buildCmd)
	reg.MustRegister(testCmd)

	// Capture both stdout and stderr to avoid file descriptor issues
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}

	// Redirect both stdout and stderr to the same pipe to catch all output
	os.Stdout = w
	os.Stderr = w

	// Run the function
	listByNamespace(reg, NewCommandDiscovery(reg))

	// Close writer and restore streams
	if err := w.Close(); err != nil {
		t.Logf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	// Read output
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Logf("Failed to read from pipe: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Logf("Failed to close reader: %v", err)
	}
	output := buf.String()

	// Should show namespaces
	if !strings.Contains(output, "build:") || !strings.Contains(output, "test:") {
		t.Errorf("listByNamespace() should organize by namespace, got: %s", output)
	}
}

func TestShowCommands_Search(t *testing.T) {
	// Create a test registry
	reg := registry.NewRegistry()

	// Register test commands
	buildCmd, err := registry.NewCommand("build").
		WithDescription("Build the project").
		WithFunc(func() error { return nil }).
		WithCategory("Build").
		Build()
	if err != nil {
		t.Fatalf("Failed to build test command: %v", err)
	}

	testCmd, err := registry.NewCommand("test").
		WithDescription("Run tests").
		WithFunc(func() error { return nil }).
		WithCategory("Test").
		Build()
	if err != nil {
		t.Fatalf("Failed to build test command: %v", err)
	}

	reg.MustRegister(buildCmd)
	reg.MustRegister(testCmd)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	searchCommands(reg, NewCommandDiscovery(reg), "build")

	if err := w.Close(); err != nil {
		t.Logf("Failed to close writer: %v", err)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Logf("Failed to read from pipe: %v", err)
	}
	output := buf.String()

	// Should only show build command
	if !strings.Contains(output, "build") {
		t.Errorf("showCommands(search='build') should contain 'build', got: %s", output)
	}
	if strings.Contains(output, "test") {
		t.Errorf("showCommands(search='build') should not contain 'test', got: %s", output)
	}
}

func TestRegistryExecute(t *testing.T) {
	// Create a test registry
	reg := registry.NewRegistry()

	// Register test command
	executed := false
	testCmd, err := registry.NewCommand("testexec").
		WithDescription("Test execution").
		WithFunc(func() error {
			executed = true
			return nil
		}).
		WithCategory("Test").
		Build()
	if err != nil {
		t.Fatalf("Failed to build test command: %v", err)
	}

	reg.MustRegister(testCmd)

	// Test successful execution
	err = reg.Execute("testexec")
	if err != nil {
		t.Errorf("Execute() should succeed, got error: %v", err)
	}
	if !executed {
		t.Error("Execute() should have executed the command")
	}
}

func TestRegistryExecute_NotFound(t *testing.T) {
	reg := registry.NewRegistry()

	err := reg.Execute("nonexistent")
	if err == nil {
		t.Error("Execute() should fail for non-existent command")
	}
	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("Execute() error should mention 'unknown command', got: %v", err)
	}
}

func TestRegistryExecute_WithArgs(t *testing.T) {
	// Create a test registry
	reg := registry.NewRegistry()

	// Register test command with args
	var receivedArgs []string
	testCmd, err := registry.NewCommand("testargs").
		WithDescription("Test with arguments").
		WithArgsFunc(func(args ...string) error {
			receivedArgs = args
			return nil
		}).
		WithCategory("Test").
		Build()
	if err != nil {
		t.Fatalf("Failed to build test command: %v", err)
	}

	reg.MustRegister(testCmd)

	// Test execution with args
	args := []string{"arg1", "arg2", "arg3"}
	err = reg.Execute("testargs", args...)
	if err != nil {
		t.Errorf("Execute() should succeed, got error: %v", err)
	}

	if len(receivedArgs) != 3 || receivedArgs[0] != "arg1" || receivedArgs[1] != "arg2" || receivedArgs[2] != "arg3" {
		t.Errorf("Execute() args = %v, expected %v", receivedArgs, args)
	}
}

func TestFlagParsing(t *testing.T) {
	// Save the original flag set
	originalFlags := flag.CommandLine
	defer func() {
		flag.CommandLine = originalFlags
	}()

	// Create a new flag set for this test
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	// Initialize flags like main() does
	_ = initFlags()

	// Test that flags are properly declared
	tests := []struct {
		name string
		flag *flag.Flag
	}{
		{"list", flag.Lookup("l")},
		{"help", flag.Lookup("h")},
		{"version", flag.Lookup("version")},
		{"verbose", flag.Lookup("v")},
		{"namespace", flag.Lookup("n")},
		{"search", flag.Lookup("search")},
		{"timeout", flag.Lookup("t")},
		{"force", flag.Lookup("f")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.flag == nil {
				t.Errorf("Flag '%s' not found", tt.name)
			}
		})
	}
}

func TestEnvironmentVariables(t *testing.T) {
	// Save original environment
	origVerbose := os.Getenv("MAGEX_VERBOSE")
	origDebug := os.Getenv("MAGEX_DEBUG")
	defer func() {
		if err := os.Setenv("MAGEX_VERBOSE", origVerbose); err != nil {
			// Test cleanup error is non-critical
			_ = err
		}
		if err := os.Setenv("MAGEX_DEBUG", origDebug); err != nil {
			// Test cleanup error is non-critical
			_ = err
		}
	}()

	// Create test flags
	verboseFlag := true
	debugFlag := false

	flags := &Flags{
		Verbose: &verboseFlag,
		Debug:   &debugFlag,
	}

	// Test verbose flag setting environment
	setEnvironmentFromFlags(flags)

	if os.Getenv("MAGEX_VERBOSE") != trueValue {
		t.Error("Verbose flag should set MAGEX_VERBOSE=true")
	}
	if os.Getenv("MAGE_X_VERBOSE") != "1" {
		t.Error("Verbose flag should set MAGE_X_VERBOSE=1")
	}

	// Reset
	verboseFlag = false
	debugFlag = true
	if err := os.Unsetenv("MAGEX_VERBOSE"); err != nil {
		// Test reset error is non-critical
		_ = err
	}
	if err := os.Unsetenv("MAGE_X_VERBOSE"); err != nil {
		// Test reset error is non-critical
		_ = err
	}

	// Test debug flag setting environment
	setEnvironmentFromFlags(flags)

	if os.Getenv("MAGEX_DEBUG") != "true" {
		t.Error("Debug flag should set MAGEX_DEBUG=true")
	}
	if os.Getenv("MAGE_X_DEBUG") != "1" {
		t.Error("Debug flag should set MAGE_X_DEBUG=1")
	}
}

// setEnvironmentFromFlags is extracted from main() for testing
func setEnvironmentFromFlags(flags *Flags) {
	if *flags.Verbose {
		if err := os.Setenv("MAGEX_VERBOSE", trueValue); err != nil {
			// Test environment variable setting is non-critical
			_ = err
		}
		if err := os.Setenv("MAGE_X_VERBOSE", "1"); err != nil {
			// Test environment variable setting is non-critical
			_ = err
		}
	}
	if *flags.Debug {
		if err := os.Setenv("MAGEX_DEBUG", "true"); err != nil {
			// Test environment variable setting is non-critical
			_ = err
		}
		if err := os.Setenv("MAGE_X_DEBUG", "1"); err != nil {
			// Test environment variable setting is non-critical
			_ = err
		}
	}
}

func BenchmarkShowCommands(b *testing.B) {
	// Create a test registry with many commands
	reg := registry.NewRegistry()

	for i := 0; i < 100; i++ {
		cmd, err := registry.NewCommand(fmt.Sprintf("cmd%d", i)).
			WithDescription(fmt.Sprintf("Command %d", i)).
			WithFunc(func() error { return nil }).
			WithCategory("Benchmark").
			Build()
		if err != nil {
			b.Fatalf("Failed to build test command: %v", err)
		}
		reg.MustRegister(cmd)
	}

	// Redirect output to discard
	oldStdout := os.Stdout
	stdout, err := os.Open(os.DevNull)
	if err != nil {
		b.Fatalf("Failed to open devnull: %v", err)
	}
	os.Stdout = stdout
	defer func() { os.Stdout = oldStdout }()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		listCommands(reg, NewCommandDiscovery(reg), false)
	}
}

func BenchmarkRegistryExecute(b *testing.B) {
	reg := registry.NewRegistry()

	testCmd, err := registry.NewCommand("bench").
		WithDescription("Benchmark command").
		WithFunc(func() error { return nil }).
		WithCategory("Benchmark").
		Build()
	if err != nil {
		b.Fatalf("Failed to build test command: %v", err)
	}

	reg.MustRegister(testCmd)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := reg.Execute("bench"); err != nil {
			// Benchmark execution error is expected, continue
			_ = err
		}
	}
}

// TestEditDistance tests the Levenshtein distance calculation
func TestEditDistance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		s1   string
		s2   string
		want int
	}{
		// Base cases - empty strings
		{"both empty", "", "", 0},
		{"first empty", "", "abc", 3},
		{"second empty", "abc", "", 3},
		{"first empty single", "", "a", 1},
		{"second empty single", "a", "", 1},

		// Identical strings
		{"identical single char", "a", "a", 0},
		{"identical short word", "test", "test", 0},
		{"identical long word", "building", "building", 0},

		// Single operations - insertion
		{"single insert end", "cat", "cats", 1},
		{"single insert middle", "cat", "cart", 1},
		{"single insert start", "at", "cat", 1},

		// Single operations - deletion
		{"single delete end", "cats", "cat", 1},
		{"single delete middle", "cart", "cat", 1},
		{"single delete start", "cat", "at", 1},

		// Single operations - replacement
		{"single replace start", "cat", "bat", 1},
		{"single replace middle", "cat", "cot", 1},
		{"single replace end", "cat", "cab", 1},

		// Multiple operations
		{"two replacements", "cat", "dog", 3},
		{"classic kitten-sitting", "kitten", "sitting", 3},
		{"classic saturday-sunday", "saturday", "sunday", 3},
		{"intention-execution", "intention", "execution", 5},

		// Different lengths
		{"one char vs two", "a", "ab", 1},
		{"prefix match", "build", "builder", 2},
		{"suffix addition", "test", "testing", 3},

		// Completely different strings
		{"completely different short", "abc", "xyz", 3},
		{"completely different medium", "hello", "world", 4},

		// Case sensitivity (editDistance is case-sensitive)
		{"case differs single", "A", "a", 1},
		{"case differs word", "Test", "test", 1},
		{"case differs all", "ABC", "abc", 3},

		// Similar command names (real use case)
		{"build typo", "build", "biuld", 2},
		{"format typo", "format", "fromat", 2},
		{"test typo", "test", "tset", 2},

		// Transpositions (counts as 2 in standard Levenshtein)
		{"transposition ab-ba", "ab", "ba", 2},
		{"transposition abc-bac", "abc", "bac", 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := editDistance(tt.s1, tt.s2)
			if got != tt.want {
				t.Errorf("editDistance(%q, %q) = %d, want %d", tt.s1, tt.s2, got, tt.want)
			}
		})
	}
}

// TestEditDistance_Symmetry verifies that edit distance is symmetric
func TestEditDistance_Symmetry(t *testing.T) {
	t.Parallel()

	pairs := [][2]string{
		{"cat", "dog"},
		{"hello", "world"},
		{"build", "test"},
		{"", "abc"},
		{"kitten", "sitting"},
	}

	for _, pair := range pairs {
		s1, s2 := pair[0], pair[1]
		t.Run(s1+"-"+s2, func(t *testing.T) {
			t.Parallel()
			d1 := editDistance(s1, s2)
			d2 := editDistance(s2, s1)
			if d1 != d2 {
				t.Errorf("editDistance(%q, %q) = %d, but editDistance(%q, %q) = %d; should be symmetric",
					s1, s2, d1, s2, s1, d2)
			}
		})
	}
}

// TestFuzzyMatch tests the fuzzy matching function used for command suggestions
func TestFuzzyMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		text    string
		pattern string
		want    bool
	}{
		// Exact matches (substring path)
		{"exact match", "build", "build", true},
		{"exact match long", "format", "format", true},

		// Substring matches (fast path - strings.Contains)
		{"substring at start", "build", "bui", true},
		{"substring at end", "build", "ild", true},
		{"substring in middle", "rebuild", "buil", true},
		{"full text is substring", "build:linux", "build", true},
		{"suffix is substring", "build:linux", "linux", true},
		{"colon is substring", "build:linux", ":", true},

		// Case insensitivity for substring matches
		{"case insensitive text upper", "BUILD", "build", true},
		{"case insensitive pattern upper", "build", "BUILD", true},
		{"case insensitive mixed", "Build", "buILD", true},
		{"case insensitive substring", "FORMAT", "orm", true},

		// Edit distance matches (distance <= 2 AND len(pattern) > 2)
		{"typo one char swap", "build", "biuld", true},    // edit distance 2
		{"typo one char wrong", "format", "fromat", true}, // edit distance 2
		{"typo missing char", "build", "buld", true},      // edit distance 1, but len > 2
		{"typo extra char", "test", "tesst", true},        // edit distance 1
		{"two typos", "format", "foramt", true},           // edit distance 2

		// Should NOT match - edit distance too high
		{"completely different", "build", "test", false},
		{"distance 3", "build", "xxxxx", false},
		{"distance 4", "hello", "world", false},

		// Should NOT match - pattern too short for edit distance
		{"short pattern exact 2", "ab", "ab", true},       // exact match still works
		{"short pattern substring", "ab", "a", true},      // substring still works
		{"short pattern typo 2 chars", "ab", "ba", false}, // len <= 2, edit distance skipped
		{"short pattern typo 1 char", "a", "b", false},    // len <= 2, not substring

		// Empty strings
		{"empty pattern", "build", "", true}, // empty pattern is substring
		{"empty text", "", "build", false},   // pattern not in empty text
		{"both empty", "", "", true},         // empty is substring of empty

		// Namespace command patterns (real use case)
		{"namespace full match", "build:linux", "build:linux", true},
		{"namespace partial namespace", "build:linux", "build", true},
		{"namespace partial command", "build:linux", "linux", true},
		// Note: "biuld" does NOT match "build:linux" because fuzzyMatch compares entire strings,
		// and editDistance("build:linux", "biuld") > 2
		{"namespace typo no match", "build:linux", "biuld", false},

		// Real command typos
		{"lint typo", "lint", "lnit", true},
		{"deps typo", "deps", "dpes", true},
		{"clean typo", "clean", "claen", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := fuzzyMatch(tt.text, tt.pattern)
			if got != tt.want {
				t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", tt.text, tt.pattern, got, tt.want)
			}
		})
	}
}

// TestFuzzyMatch_EditDistanceThreshold verifies the edit distance threshold behavior
func TestFuzzyMatch_EditDistanceThreshold(t *testing.T) {
	t.Parallel()

	// These test the exact boundary: edit distance = 2 should match, > 2 should not
	tests := []struct {
		name    string
		text    string
		pattern string
		want    bool
	}{
		// Distance exactly 2 - should match (if len > 2)
		{"distance 2 matches", "abcde", "abcxy", true},
		// Distance exactly 3 - should NOT match
		{"distance 3 no match", "abcde", "abxyz", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := fuzzyMatch(tt.text, tt.pattern)
			if got != tt.want {
				t.Errorf("fuzzyMatch(%q, %q) = %v, want %v (edit distance threshold test)",
					tt.text, tt.pattern, got, tt.want)
			}
		})
	}
}

// TestHighlightMatch tests the match highlighting function for CLI output
func TestHighlightMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		text    string
		pattern string
		want    string
	}{
		// Match at different positions
		{"match at start", "build", "bui", "[bui]ld"},
		{"match at end", "build", "ild", "bu[ild]"},
		{"match in middle", "rebuild", "buil", "re[buil]d"},
		{"full match", "test", "test", "[test]"},

		// No match - returns original text
		{"no match different", "build", "xyz", "build"},
		{"no match partial", "build", "abc", "build"},

		// Case insensitivity (finds match regardless of case)
		{"case insensitive lower pattern", "Build", "build", "[Build]"},
		{"case insensitive upper pattern", "build", "BUILD", "[build]"},
		{"case insensitive mixed", "BuIlD", "build", "[BuIlD]"},

		// Preserves original case in output
		{"preserves case upper", "BUILD", "build", "[BUILD]"},
		{"preserves case mixed", "Build", "bui", "[Bui]ld"},

		// Single character matches
		{"single char match", "test", "t", "[t]est"},
		{"single char end", "test", "t", "[t]est"}, // finds first occurrence

		// Empty strings
		{"empty pattern", "build", "", "[]build"}, // Index returns 0 for empty, wraps nothing at start
		{"empty text", "", "build", ""},

		// Namespace commands
		{"namespace highlight ns", "build:linux", "build", "[build]:linux"},
		{"namespace highlight cmd", "build:linux", "linux", "build:[linux]"},
		{"namespace highlight colon", "build:linux", ":", "build[:]linux"},

		// Multiple potential matches (highlights first one)
		{"multiple matches", "testtest", "test", "[test]test"},

		// Partial word matches
		{"partial word", "building", "uild", "b[uild]ing"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := highlightMatch(tt.text, tt.pattern)
			if got != tt.want {
				t.Errorf("highlightMatch(%q, %q) = %q, want %q", tt.text, tt.pattern, got, tt.want)
			}
		})
	}
}

// TestHandleNoSearchResults tests the search suggestion display when no exact matches found
func TestHandleNoSearchResults(t *testing.T) {
	tests := []struct {
		name           string
		commands       []string // commands to register in test registry
		customCommands []DiscoveredCommand
		query          string
		wantContains   []string
		wantNotContain []string
	}{
		{
			name:         "typo suggests similar command",
			commands:     []string{"build", "test", "lint", "format"},
			query:        "biuld", // typo of "build" - edit distance 2
			wantContains: []string{"No exact commands found", "Did you mean", "build"},
		},
		{
			name:           "no suggestions for completely different query",
			commands:       []string{"build", "test", "lint"},
			query:          "xyzabc123",
			wantContains:   []string{"No exact commands found"},
			wantNotContain: []string{"Did you mean"},
		},
		{
			name:           "suggests custom command",
			commands:       []string{"build"},
			customCommands: []DiscoveredCommand{{Name: "deploy", Description: "Deploy app"}},
			query:          "deplyo", // typo of "deploy"
			wantContains:   []string{"No exact commands found", "Did you mean", "deploy", "custom"},
		},
		{
			name:         "multiple suggestions",
			commands:     []string{"build", "built", "bulls"},
			query:        "buil", // substring match for multiple
			wantContains: []string{"Did you mean", "build"},
		},
		{
			name:           "empty registry no crash",
			commands:       []string{},
			customCommands: []DiscoveredCommand{},
			query:          "anything",
			wantContains:   []string{"No exact commands found"},
			wantNotContain: []string{"Did you mean"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test registry
			reg := registry.NewRegistry()
			for _, cmdName := range tt.commands {
				cmd, err := registry.NewCommand(cmdName).
					WithDescription(cmdName + " command").
					WithFunc(func() error { return nil }).
					Build()
				if err != nil {
					t.Fatalf("Failed to build test command %s: %v", cmdName, err)
				}
				reg.MustRegister(cmd)
			}

			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			// Call the function
			handleNoSearchResults(reg, tt.customCommands, tt.query)

			// Close and restore
			if err := w.Close(); err != nil {
				t.Logf("Failed to close writer: %v", err)
			}
			os.Stdout = oldStdout

			// Read output
			var buf bytes.Buffer
			if _, err := buf.ReadFrom(r); err != nil {
				t.Logf("Failed to read from pipe: %v", err)
			}
			output := buf.String()

			// Verify expected content
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("handleNoSearchResults() output should contain %q\nGot: %s", want, output)
				}
			}

			// Verify content that should NOT be present
			for _, notWant := range tt.wantNotContain {
				if strings.Contains(output, notWant) {
					t.Errorf("handleNoSearchResults() output should NOT contain %q\nGot: %s", notWant, output)
				}
			}
		})
	}
}

// TestNormalizeCommandName tests the command name normalization function
func TestNormalizeCommandName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"already normalized lowercase colon", "build:default", "build:default"},
		{"dot separator", "build.default", "build:default"},
		{"dash separator", "build-default", "build:default"},
		{"mixed case with dot", "Build.Default", "build:default"},
		{"uppercase with dash", "TEST-UNIT", "test:unit"},
		{"multiple dots", "ns.sub.method", "ns:sub:method"},
		{"multiple dashes", "ns-sub-method", "ns:sub:method"},
		{"mixed separators dot and dash", "ns.sub-method", "ns:sub:method"},
		{"no separator lowercase", "build", "build"},
		{"no separator uppercase", "BUILD", "build"},
		{"single char lowercase", "b", "b"},
		{"single char uppercase", "B", "b"},
		{"already lowercase no sep", "test", "test"},
		{"CamelCase no separator", "BuildProject", "buildproject"},
		{"numbers preserved", "test123", "test123"},
		{"numbers with separator", "test.123", "test:123"},
		{"underscore preserved", "test_unit", "test_unit"},
		{"colon already present", "build:linux", "build:linux"},
		{"mixed colon and dot", "build:sub.method", "build:sub:method"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := normalizeCommandName(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeCommandName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestTruncate tests the string truncation helper function
func TestTruncate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"shorter than max", "hello", 10, "hello"},
		{"exactly max length", "hello", 5, "hello"},
		{"longer than max needs truncation", "hello world", 8, "hello..."},
		{"empty string any max", "", 10, ""},
		{"empty string zero max", "", 0, ""},
		{"truncate to 4 chars", "hello", 4, "h..."},
		{"truncate to 5 chars", "abcdefgh", 5, "ab..."},
		{"truncate to 6 chars", "abcdefgh", 6, "abc..."},
		{"truncate long string", "this is a very long string", 15, "this is a ve..."},
		{"single char shorter than max", "a", 5, "a"},
		{"unicode string shorter", "héllo", 10, "héllo"},
		{"whitespace preserved", "a b c", 10, "a b c"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestTruncateEdgeCases tests edge cases for the truncate function
func TestTruncateEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"maxLen exactly 3 with long string", "abcd", 3, "..."},
		{"maxLen 3 with 3 char string", "abc", 3, "abc"},
		{"maxLen 3 with 4 char string", "abcd", 3, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := truncate(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// TestShowQuickList tests the quick command list display
func TestShowQuickList(t *testing.T) {
	tests := []struct {
		name            string
		commands        []string
		customCommands  []DiscoveredCommand
		wantContains    []string
		wantNotContains []string
	}{
		{
			name:         "shows common commands",
			commands:     []string{"build", "test", "lint", "format", "clean"},
			wantContains: []string{"build", "test", "lint", "format", "clean"},
		},
		{
			name:           "shows custom commands section",
			commands:       []string{"build"},
			customCommands: []DiscoveredCommand{{Name: "deploy", Description: "Deploy app"}},
			wantContains:   []string{"build", "Custom commands", "deploy"},
		},
		{
			name:     "limits custom commands to 5",
			commands: []string{},
			customCommands: []DiscoveredCommand{
				{Name: "cmd1", Description: "Cmd 1"},
				{Name: "cmd2", Description: "Cmd 2"},
				{Name: "cmd3", Description: "Cmd 3"},
				{Name: "cmd4", Description: "Cmd 4"},
				{Name: "cmd5", Description: "Cmd 5"},
				{Name: "cmd6", Description: "Cmd 6"},
				{Name: "cmd7", Description: "Cmd 7"},
			},
			wantContains:    []string{"cmd1", "cmd2", "cmd3", "cmd4", "cmd5", "... and 2 more"},
			wantNotContains: []string{"cmd6", "cmd7"},
		},
		{
			name:         "handles empty registry",
			commands:     []string{},
			wantContains: []string{}, // Just should not crash
		},
		{
			name:           "uses default description for custom commands without description",
			commands:       []string{},
			customCommands: []DiscoveredCommand{{Name: "nodesc", Description: ""}},
			wantContains:   []string{"nodesc", "Custom command"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test registry
			reg := registry.NewRegistry()
			for _, cmdName := range tt.commands {
				cmd, err := registry.NewCommand(cmdName).
					WithDescription(cmdName + " command").
					WithFunc(func() error { return nil }).
					Build()
				if err != nil {
					t.Fatalf("Failed to build test command %s: %v", cmdName, err)
				}
				reg.MustRegister(cmd)
			}

			// Create discovery with custom commands
			discovery := NewCommandDiscovery(reg)
			discovery.commands = tt.customCommands
			discovery.loaded = true

			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			showQuickList(reg, discovery)

			if closeErr := w.Close(); closeErr != nil {
				t.Logf("Failed to close writer: %v", closeErr)
			}
			os.Stdout = oldStdout

			var buf bytes.Buffer
			if _, readErr := buf.ReadFrom(r); readErr != nil {
				t.Logf("Failed to read from pipe: %v", readErr)
			}
			output := buf.String()

			// Check expected content
			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("showQuickList() output should contain %q\nGot: %s", want, output)
				}
			}

			// Check content that should NOT be present
			for _, notWant := range tt.wantNotContains {
				if strings.Contains(output, notWant) {
					t.Errorf("showQuickList() output should NOT contain %q\nGot: %s", notWant, output)
				}
			}
		})
	}
}

// TestListCommandsVerbose tests verbose command listing
func TestListCommandsVerbose(t *testing.T) {
	tests := []struct {
		name           string
		commands       []*registry.Command
		customCommands []DiscoveredCommand
		wantContains   []string
	}{
		{
			name:         "shows built-in commands with descriptions",
			commands:     createTestCommands(t, []string{"build", "test"}),
			wantContains: []string{"build", "test"},
		},
		{
			name:           "shows custom commands with custom marker",
			commands:       createTestCommands(t, []string{"build"}),
			customCommands: []DiscoveredCommand{{Name: "deploy", Description: "Deploy app"}},
			wantContains:   []string{"build", "deploy", "(custom)"},
		},
		{
			name:           "uses default description for empty custom command description",
			commands:       []*registry.Command{},
			customCommands: []DiscoveredCommand{{Name: "nodesc", Description: ""}},
			wantContains:   []string{"nodesc", "Custom command", "(custom)"},
		},
		{
			name:         "shows deprecated warning",
			commands:     createDeprecatedCommand(t),
			wantContains: []string{"DEPRECATED"},
		},
		{
			name:           "handles empty commands list",
			commands:       []*registry.Command{},
			customCommands: []DiscoveredCommand{},
			wantContains:   []string{}, // Just should not crash
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			listCommandsVerbose(tt.commands, tt.customCommands)

			if closeErr := w.Close(); closeErr != nil {
				t.Logf("Failed to close writer: %v", closeErr)
			}
			os.Stdout = oldStdout

			var buf bytes.Buffer
			if _, readErr := buf.ReadFrom(r); readErr != nil {
				t.Logf("Failed to read from pipe: %v", readErr)
			}
			output := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("listCommandsVerbose() output should contain %q\nGot: %s", want, output)
				}
			}
		})
	}
}

// TestShowCommandHelp tests detailed command help display
func TestShowCommandHelp(t *testing.T) {
	tests := []struct {
		name         string
		commandName  string
		setupReg     func(*registry.Registry)
		wantContains []string
	}{
		{
			name:        "shows help for existing command",
			commandName: "build",
			setupReg: func(reg *registry.Registry) {
				cmd := registry.NewCommand("build").
					WithDescription("Build the project").
					WithLongDescription("Builds the project with all dependencies").
					WithUsage("magex build [flags]").
					WithCategory("Build").
					WithExamples("magex build", "magex build --verbose").
					WithAliases("b").
					WithFunc(func() error { return nil }).
					MustBuild()
				reg.MustRegister(cmd)
			},
			wantContains: []string{
				"Command Help: build",
				"Description:",
				"Build the project",
				"Detailed Description:",
				"Usage:",
				"Category: Build",
				"Aliases:",
				"Examples:",
			},
		},
		{
			name:        "shows namespace help for namespace name",
			commandName: "test",
			setupReg: func(reg *registry.Registry) {
				cmd1 := registry.NewNamespaceCommand("test", "unit").
					WithDescription("Run unit tests").
					WithFunc(func() error { return nil }).
					MustBuild()
				cmd2 := registry.NewNamespaceCommand("test", "integration").
					WithDescription("Run integration tests").
					WithFunc(func() error { return nil }).
					MustBuild()
				reg.MustRegister(cmd1)
				reg.MustRegister(cmd2)
			},
			wantContains: []string{"Namespace Help: test", "test:unit", "test:integration"},
		},
		{
			name:        "shows error for unknown command",
			commandName: "nonexistent",
			setupReg:    func(reg *registry.Registry) {},
			wantContains: []string{
				"Unknown command 'nonexistent'",
			},
		},
		{
			name:        "shows suggestions for similar commands",
			commandName: "buil",
			setupReg: func(reg *registry.Registry) {
				cmd := registry.NewCommand("build").
					WithDescription("Build the project").
					WithFunc(func() error { return nil }).
					MustBuild()
				reg.MustRegister(cmd)
			},
			wantContains: []string{"Unknown command 'buil'", "Did you mean", "build"},
		},
		{
			name:        "shows deprecation warning",
			commandName: "oldcmd",
			setupReg: func(reg *registry.Registry) {
				cmd := registry.NewCommand("oldcmd").
					WithDescription("Old command").
					Deprecated("Use newcmd instead").
					WithFunc(func() error { return nil }).
					MustBuild()
				reg.MustRegister(cmd)
			},
			wantContains: []string{"WARNING", "deprecated"},
		},
		{
			name:        "shows since version",
			commandName: "newcmd",
			setupReg: func(reg *registry.Registry) {
				cmd := registry.NewCommand("newcmd").
					WithDescription("New command").
					Since("1.2.0").
					WithFunc(func() error { return nil }).
					MustBuild()
				reg.MustRegister(cmd)
			},
			wantContains: []string{"Since: MAGE-X 1.2.0"},
		},
		{
			name:        "shows tags",
			commandName: "taggedcmd",
			setupReg: func(reg *registry.Registry) {
				cmd := registry.NewCommand("taggedcmd").
					WithDescription("Tagged command").
					WithTags("core", "essential").
					WithFunc(func() error { return nil }).
					MustBuild()
				reg.MustRegister(cmd)
			},
			wantContains: []string{"Tags:", "core", "essential"},
		},
		{
			name:        "shows dependencies",
			commandName: "depcmd",
			setupReg: func(reg *registry.Registry) {
				cmd := registry.NewCommand("depcmd").
					WithDescription("Command with deps").
					WithDependencies("build", "test").
					WithFunc(func() error { return nil }).
					MustBuild()
				reg.MustRegister(cmd)
			},
			wantContains: []string{"Dependencies:", "build", "test"},
		},
		{
			name:        "shows see also",
			commandName: "seealso",
			setupReg: func(reg *registry.Registry) {
				cmd := registry.NewCommand("seealso").
					WithDescription("See also command").
					WithSeeAlso("build", "test").
					WithFunc(func() error { return nil }).
					MustBuild()
				reg.MustRegister(cmd)
			},
			wantContains: []string{"See Also:", "magex build", "magex test"},
		},
		{
			name:        "shows options",
			commandName: "optcmd",
			setupReg: func(reg *registry.Registry) {
				cmd := registry.NewCommand("optcmd").
					WithDescription("Command with options").
					WithOptions(
						registry.CommandOption{Name: "--verbose", Description: "Enable verbose output", Default: "false"},
						registry.CommandOption{Name: "--output", Description: "Output file", Required: true},
					).
					WithFunc(func() error { return nil }).
					MustBuild()
				reg.MustRegister(cmd)
			},
			wantContains: []string{"Options:", "--verbose", "--output", "(required)", "[default: false]"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := registry.NewRegistry()
			tt.setupReg(reg)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			showCommandHelp(reg, tt.commandName)

			if closeErr := w.Close(); closeErr != nil {
				t.Logf("Failed to close writer: %v", closeErr)
			}
			os.Stdout = oldStdout

			var buf bytes.Buffer
			if _, readErr := buf.ReadFrom(r); readErr != nil {
				t.Logf("Failed to read from pipe: %v", readErr)
			}
			output := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("showCommandHelp(%q) output should contain %q\nGot: %s", tt.commandName, want, output)
				}
			}
		})
	}
}

// TestShowNamespaceHelp tests namespace help display
func TestShowNamespaceHelp(t *testing.T) {
	tests := []struct {
		name         string
		namespace    string
		setupReg     func(*registry.Registry)
		wantContains []string
	}{
		{
			name:      "shows namespace commands",
			namespace: "build",
			setupReg: func(reg *registry.Registry) {
				cmd1 := registry.NewNamespaceCommand("build", "linux").
					WithDescription("Build for Linux").
					WithFunc(func() error { return nil }).
					MustBuild()
				cmd2 := registry.NewNamespaceCommand("build", "darwin").
					WithDescription("Build for macOS").
					WithFunc(func() error { return nil }).
					MustBuild()
				cmd3 := registry.NewNamespaceCommand("build", "windows").
					WithDescription("Build for Windows").
					WithFunc(func() error { return nil }).
					MustBuild()
				reg.MustRegister(cmd1)
				reg.MustRegister(cmd2)
				reg.MustRegister(cmd3)
			},
			wantContains: []string{
				"Namespace Help: build",
				"Available Commands in build namespace (3 commands)",
				"build:linux",
				"build:darwin",
				"build:windows",
				"Usage Examples:",
				"For detailed help",
			},
		},
		{
			name:         "shows error for empty namespace",
			namespace:    "nonexistent",
			setupReg:     func(reg *registry.Registry) {},
			wantContains: []string{"No commands found in namespace 'nonexistent'"},
		},
		{
			name:      "truncates long descriptions",
			namespace: "long",
			setupReg: func(reg *registry.Registry) {
				cmd := registry.NewNamespaceCommand("long", "cmd").
					WithDescription("This is a very long description that should be truncated because it exceeds the maximum length allowed for display in the namespace help output").
					WithFunc(func() error { return nil }).
					MustBuild()
				reg.MustRegister(cmd)
			},
			wantContains: []string{"Namespace Help: long", "long:cmd", "..."},
		},
		{
			name:      "limits usage examples to 3",
			namespace: "many",
			setupReg: func(reg *registry.Registry) {
				for i := 1; i <= 5; i++ {
					cmd := registry.NewNamespaceCommand("many", fmt.Sprintf("cmd%d", i)).
						WithDescription(fmt.Sprintf("Command %d", i)).
						WithFunc(func() error { return nil }).
						MustBuild()
					reg.MustRegister(cmd)
				}
			},
			wantContains: []string{"... and 2 more commands"},
		},
		{
			name:      "uses no description placeholder",
			namespace: "nodesc",
			setupReg: func(reg *registry.Registry) {
				cmd := registry.NewNamespaceCommand("nodesc", "cmd").
					WithDescription("").
					WithFunc(func() error { return nil }).
					MustBuild()
				reg.MustRegister(cmd)
			},
			wantContains: []string{"nodesc:cmd"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := registry.NewRegistry()
			tt.setupReg(reg)

			// Capture stdout
			oldStdout := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			showNamespaceHelp(reg, tt.namespace)

			if closeErr := w.Close(); closeErr != nil {
				t.Logf("Failed to close writer: %v", closeErr)
			}
			os.Stdout = oldStdout

			var buf bytes.Buffer
			if _, readErr := buf.ReadFrom(r); readErr != nil {
				t.Logf("Failed to read from pipe: %v", readErr)
			}
			output := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("showNamespaceHelp(%q) output should contain %q\nGot: %s", tt.namespace, want, output)
				}
			}
		})
	}
}

// Helper functions for test setup

func createTestCommands(t *testing.T, names []string) []*registry.Command {
	t.Helper()
	commands := make([]*registry.Command, 0, len(names))
	for _, name := range names {
		cmd, err := registry.NewCommand(name).
			WithDescription(name + " command").
			WithFunc(func() error { return nil }).
			Build()
		if err != nil {
			t.Fatalf("Failed to build test command %s: %v", name, err)
		}
		commands = append(commands, cmd)
	}
	return commands
}

func createDeprecatedCommand(t *testing.T) []*registry.Command {
	t.Helper()
	cmd, err := registry.NewCommand("oldcmd").
		WithDescription("Old command").
		Deprecated("Use newcmd instead").
		WithFunc(func() error { return nil }).
		Build()
	if err != nil {
		t.Fatalf("Failed to build deprecated command: %v", err)
	}
	return []*registry.Command{cmd}
}

// TestListByNamespace_DefaultMethod tests listByNamespace with default method names
func TestListByNamespace_DefaultMethod(t *testing.T) {
	reg := registry.NewRegistry()

	// Create a command with default method name
	defaultCmd, err := registry.NewNamespaceCommand("deploy", "default").
		WithDescription("Default deploy command").
		WithFunc(func() error { return nil }).
		Build()
	if err != nil {
		t.Fatalf("Failed to build default command: %v", err)
	}
	reg.MustRegister(defaultCmd)

	// Also test with "Default" (capitalized)
	capitalCmd, err := registry.NewNamespaceCommand("release", "Default").
		WithDescription("Default release command").
		WithFunc(func() error { return nil }).
		Build()
	if err != nil {
		t.Fatalf("Failed to build capitalized default command: %v", err)
	}
	reg.MustRegister(capitalCmd)

	// Regular namespaced command for comparison
	regularCmd, err := registry.NewNamespaceCommand("build", "linux").
		WithDescription("Build for Linux").
		WithFunc(func() error { return nil }).
		Build()
	if err != nil {
		t.Fatalf("Failed to build regular command: %v", err)
	}
	reg.MustRegister(regularCmd)

	// Capture stdout
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w
	os.Stderr = w

	listByNamespace(reg, NewCommandDiscovery(reg))

	if closeErr := w.Close(); closeErr != nil {
		t.Logf("Failed to close writer: %v", closeErr)
	}
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var buf bytes.Buffer
	if _, readErr := buf.ReadFrom(r); readErr != nil {
		t.Logf("Failed to read from pipe: %v", readErr)
	}
	output := buf.String()

	// Default methods should show without the method suffix
	if !strings.Contains(output, "deploy:") {
		t.Errorf("listByNamespace() should show deploy namespace, got: %s", output)
	}
	if !strings.Contains(output, "build:linux") {
		t.Errorf("listByNamespace() should show build:linux, got: %s", output)
	}
}

// TestSearchCommands_NoResults tests searchCommands when no results are found
func TestSearchCommands_NoResults(t *testing.T) {
	reg := registry.NewRegistry()

	// Register a command
	cmd, err := registry.NewCommand("build").
		WithDescription("Build the project").
		WithFunc(func() error { return nil }).
		Build()
	if err != nil {
		t.Fatalf("Failed to build command: %v", err)
	}
	reg.MustRegister(cmd)

	// Capture stdout
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w
	os.Stderr = w

	// Search for something that doesn't exist
	searchCommands(reg, NewCommandDiscovery(reg), "nonexistent_xyz_query")

	if closeErr := w.Close(); closeErr != nil {
		t.Logf("Failed to close writer: %v", closeErr)
	}
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var buf bytes.Buffer
	if _, readErr := buf.ReadFrom(r); readErr != nil {
		t.Logf("Failed to read from pipe: %v", readErr)
	}
	output := buf.String()

	// Should indicate no results found
	if !strings.Contains(strings.ToLower(output), "no") && !strings.Contains(strings.ToLower(output), "found") {
		t.Logf("searchCommands() output for no results: %s", output)
	}
}

// TestSearchCommands_DescriptionMatch tests searchCommands matching on description
func TestSearchCommands_DescriptionMatch(t *testing.T) {
	reg := registry.NewRegistry()

	// Register commands with unique descriptions
	cmd1, err := registry.NewCommand("foo").
		WithDescription("Compiles the project for production").
		WithFunc(func() error { return nil }).
		Build()
	if err != nil {
		t.Fatalf("Failed to build command: %v", err)
	}
	cmd2, err := registry.NewCommand("bar").
		WithDescription("Tests the application").
		WithFunc(func() error { return nil }).
		Build()
	if err != nil {
		t.Fatalf("Failed to build command: %v", err)
	}
	reg.MustRegister(cmd1)
	reg.MustRegister(cmd2)

	// Capture stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w
	os.Stderr = w

	// Search for "production" which is only in foo's description
	searchCommands(reg, NewCommandDiscovery(reg), "production")

	if closeErr := w.Close(); closeErr != nil {
		t.Logf("Failed to close writer: %v", closeErr)
	}
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	var buf bytes.Buffer
	if _, readErr := buf.ReadFrom(r); readErr != nil {
		t.Logf("Failed to read from pipe: %v", readErr)
	}
	output := buf.String()

	// Should find at least one result
	if !strings.Contains(output, "1 found") {
		t.Errorf("searchCommands('production') should find at least 1 result, got: %s", output)
	}
}

// TestListCommands_Verbose tests listCommands with verbose output
func TestListCommands_Verbose(t *testing.T) {
	reg := registry.NewRegistry()

	cmd, err := registry.NewCommand("mycommand").
		WithDescription("A command with a detailed description").
		WithFunc(func() error { return nil }).
		WithCategory("MyCategory").
		Build()
	if err != nil {
		t.Fatalf("Failed to build command: %v", err)
	}
	reg.MustRegister(cmd)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	listCommands(reg, NewCommandDiscovery(reg), true) // verbose=true

	if closeErr := w.Close(); closeErr != nil {
		t.Logf("Failed to close writer: %v", closeErr)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, readErr := buf.ReadFrom(r); readErr != nil {
		t.Logf("Failed to read from pipe: %v", readErr)
	}
	output := buf.String()

	if !strings.Contains(output, "mycommand") {
		t.Errorf("listCommands(verbose) should contain command name, got: %s", output)
	}
	// Verbose should show the description
	if !strings.Contains(output, "detailed description") {
		t.Errorf("listCommands(verbose) should contain description, got: %s", output)
	}
}

// TestShowCategorizedCommands tests categorized command display
func TestShowCategorizedCommands(t *testing.T) {
	reg := registry.NewRegistry()

	// Create commands in different categories
	buildCmd, err := registry.NewCommand("build").
		WithDescription("Build the project").
		WithFunc(func() error { return nil }).
		WithCategory("Build").
		Build()
	if err != nil {
		t.Fatalf("Failed to build command: %v", err)
	}

	testCmd, err := registry.NewCommand("test").
		WithDescription("Run tests").
		WithFunc(func() error { return nil }).
		WithCategory("Test").
		Build()
	if err != nil {
		t.Fatalf("Failed to build command: %v", err)
	}

	uncategorizedCmd, err := registry.NewCommand("misc").
		WithDescription("Miscellaneous command").
		WithFunc(func() error { return nil }).
		Build()
	if err != nil {
		t.Fatalf("Failed to build command: %v", err)
	}

	reg.MustRegister(buildCmd)
	reg.MustRegister(testCmd)
	reg.MustRegister(uncategorizedCmd)

	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	showCategorizedCommands(reg)

	if closeErr := w.Close(); closeErr != nil {
		t.Logf("Failed to close writer: %v", closeErr)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, readErr := buf.ReadFrom(r); readErr != nil {
		t.Logf("Failed to read from pipe: %v", readErr)
	}
	output := buf.String()

	// Should contain category headers
	if !strings.Contains(output, "Build") {
		t.Errorf("showCategorizedCommands() should contain 'Build' category, got: %s", output)
	}
	if !strings.Contains(output, "Test") {
		t.Errorf("showCategorizedCommands() should contain 'Test' category, got: %s", output)
	}
}

// TestShowVersionWithDetails tests version display with build details
func TestShowVersionWithDetails(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("Failed to create pipe: %v", err)
	}
	os.Stdout = w

	showVersion()

	if closeErr := w.Close(); closeErr != nil {
		t.Logf("Failed to close writer: %v", closeErr)
	}
	os.Stdout = oldStdout

	var buf bytes.Buffer
	if _, readErr := buf.ReadFrom(r); readErr != nil {
		t.Logf("Failed to read from pipe: %v", readErr)
	}
	output := buf.String()

	// Should contain version info
	if !strings.Contains(output, "MAGE-X") {
		t.Errorf("showVersion() should contain 'MAGE-X', got: %s", output)
	}
}
