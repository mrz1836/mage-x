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
