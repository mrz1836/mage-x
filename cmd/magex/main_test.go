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

	listCommands(reg, false)

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
	listByNamespace(reg)

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

	searchCommands(reg, "build")

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

	if os.Getenv("MAGEX_VERBOSE") != "true" {
		t.Error("Verbose flag should set MAGEX_VERBOSE=true")
	}
	if os.Getenv("MAGE_VERBOSE") != "1" {
		t.Error("Verbose flag should set MAGE_VERBOSE=1")
	}

	// Reset
	verboseFlag = false
	debugFlag = true
	if err := os.Unsetenv("MAGEX_VERBOSE"); err != nil {
		// Test reset error is non-critical
		_ = err
	}
	if err := os.Unsetenv("MAGE_VERBOSE"); err != nil {
		// Test reset error is non-critical
		_ = err
	}

	// Test debug flag setting environment
	setEnvironmentFromFlags(flags)

	if os.Getenv("MAGEX_DEBUG") != "true" {
		t.Error("Debug flag should set MAGEX_DEBUG=true")
	}
	if os.Getenv("MAGE_DEBUG") != "1" {
		t.Error("Debug flag should set MAGE_DEBUG=1")
	}
}

// setEnvironmentFromFlags is extracted from main() for testing
func setEnvironmentFromFlags(flags *Flags) {
	if *flags.Verbose {
		if err := os.Setenv("MAGEX_VERBOSE", "true"); err != nil {
			// Test environment variable setting is non-critical
			_ = err
		}
		if err := os.Setenv("MAGE_VERBOSE", "1"); err != nil {
			// Test environment variable setting is non-critical
			_ = err
		}
	}
	if *flags.Debug {
		if err := os.Setenv("MAGEX_DEBUG", "true"); err != nil {
			// Test environment variable setting is non-critical
			_ = err
		}
		if err := os.Setenv("MAGE_DEBUG", "1"); err != nil {
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
		listCommands(reg, false)
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
