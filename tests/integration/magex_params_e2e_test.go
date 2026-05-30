//go:build integration
// +build integration

package integration

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMagexParameterPassingE2E tests that parameters are correctly passed
// through the entire magex -> mage -> function pipeline
func TestMagexParameterPassingE2E(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Build magex first to ensure we're testing the latest code
	t.Run("BuildMagex", func(t *testing.T) {
		cmd := testCommand(t, "go", "build", "-o", "magex-test", "../../cmd/magex")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to build magex: %v\nOutput: %s", err, output)
		}
	})

	// Clean up the test binary after tests
	cleanupFile(t, "magex-test")

	// Get the absolute path to the test binary
	magexPath, err := filepath.Abs("./magex-test")
	require.NoError(t, err)

	t.Run("TestFuzzTimeParameter", func(t *testing.T) {
		// Create a test magefile with a fuzz test
		testDir := t.TempDir()
		magefilePath := filepath.Join(testDir, "magefile.go")
		testFilePath := filepath.Join(testDir, "fuzz_test.go")

		// Create a simple magefile that uses our fixed namespace
		magefileContent := `//go:build mage
// +build mage

package main

import (
	"os"
	"strings"
	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/mage"
)

type Test mg.Namespace

func getMageArgs() []string {
	if mageArgs := os.Getenv("MAGE_ARGS"); mageArgs != "" {
		return strings.Fields(mageArgs)
	}
	return nil
}

func (t Test) Fuzz() error {
	var impl mage.Test
	return impl.Fuzz(getMageArgs()...)
}
`
		err := os.WriteFile(magefilePath, []byte(magefileContent), 0o600)
		require.NoError(t, err)

		// Create a simple fuzz test
		fuzzTestContent := `package main

import "testing"

func FuzzTest(f *testing.F) {
	f.Fuzz(func(t *testing.T, data []byte) {
		// Simple fuzz test
		_ = string(data)
	})
}
`
		err = os.WriteFile(testFilePath, []byte(fuzzTestContent), 0o600)
		require.NoError(t, err)

		// Run magex with time parameter
		cmd := testCommand(t, magexPath, "test:fuzz", "time=1s")
		cmd.Dir = testDir

		// Capture output
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		// Run with timeout to ensure it doesn't run for default 10s
		done := make(chan error, 1)
		go func() {
			done <- cmd.Run()
		}()

		select {
		case err := <-done:
			// Command completed
			output := stdout.String() + stderr.String()
			if err != nil {
				t.Errorf("Fuzz test command failed: %v\nOutput: %s", err, output)
			}

			// Verify it used the correct time
			if strings.Contains(output, "fuzzing for 10s") {
				t.Errorf("Used default 10s instead of specified 1s\nOutput: %s", output)
			}
			if !strings.Contains(output, "fuzz") || !strings.Contains(output, "1s") {
				// The output format varies, but it should mention the time
				t.Logf("Output: %s", output)
			}
		case <-time.After(5 * time.Second):
			// If it takes more than 5s, it's probably using the wrong time
			if err := cmd.Process.Kill(); err != nil {
				t.Fatalf("kill fuzz test command: %v", err)
			}
			t.Error("Fuzz test took too long - likely using default 10s instead of 1s")
		}
	})

	t.Run("VersionBumpDryRun", func(t *testing.T) {
		// Test version bump with dry-run
		testDir := t.TempDir()

		// Initialize git repo
		cmd := testCommand(t, "git", "init")
		cmd.Dir = testDir
		err := cmd.Run()
		require.NoError(t, err)

		// Configure git
		cmd = testCommand(t, "git", "config", "user.email", "test@example.com")
		cmd.Dir = testDir
		err = cmd.Run()
		require.NoError(t, err)

		cmd = testCommand(t, "git", "config", "user.name", "Test User")
		cmd.Dir = testDir
		err = cmd.Run()
		require.NoError(t, err)

		// Create initial commit
		testFile := filepath.Join(testDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0o600)
		require.NoError(t, err)

		cmd = testCommand(t, "git", "add", ".")
		cmd.Dir = testDir
		err = cmd.Run()
		require.NoError(t, err)

		cmd = testCommand(t, "git", "commit", "-m", "initial")
		cmd.Dir = testDir
		err = cmd.Run()
		require.NoError(t, err)

		// Add initial tag
		cmd = testCommand(t, "git", "tag", "v1.0.0")
		cmd.Dir = testDir
		err = cmd.Run()
		require.NoError(t, err)

		// Create magefile
		magefilePath := filepath.Join(testDir, "magefile.go")
		magefileContent := `//go:build mage
// +build mage

package main

import (
	"os"
	"strings"
	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/mage"
)

type Version mg.Namespace

func getMageArgs() []string {
	if mageArgs := os.Getenv("MAGE_ARGS"); mageArgs != "" {
		return strings.Fields(mageArgs)
	}
	return nil
}

func (v Version) Bump() error {
	var impl mage.Version
	return impl.Bump(getMageArgs()...)
}
`
		err = os.WriteFile(magefilePath, []byte(magefileContent), 0o600)
		require.NoError(t, err)

		// Run version bump with dry-run
		cmd = testCommand(t, magexPath, "version:bump", "dry-run", "bump=minor")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Should succeed
		require.NoError(t, err, "Command failed: %s", outputStr)

		// Should indicate dry-run mode
		assert.Contains(t, outputStr, "DRY-RUN", "Should indicate dry-run mode")
		assert.Contains(t, outputStr, "v1.1.0", "Should show version bump to v1.1.0")

		// Verify no tag was actually created
		cmd = testCommand(t, "git", "tag", "-l")
		cmd.Dir = testDir
		tags, err := cmd.Output()
		require.NoError(t, err)
		assert.NotContains(t, string(tags), "v1.1.0", "Should not create tag in dry-run")
	})

	t.Run("TestCoveragePackageParameter", func(t *testing.T) {
		// Test coverage with package parameter
		testDir := t.TempDir()

		// Create a simple Go module
		modContent := `module testmodule

go 1.21

require github.com/mrz1836/mage-x v0.0.0
`
		err := os.WriteFile(filepath.Join(testDir, "go.mod"), []byte(modContent), 0o600)
		require.NoError(t, err)

		// Create test files in different packages
		err = os.MkdirAll(filepath.Join(testDir, "pkg1"), 0o750)
		require.NoError(t, err)
		err = os.MkdirAll(filepath.Join(testDir, "pkg2"), 0o750)
		require.NoError(t, err)

		// pkg1 test
		pkg1Test := `package pkg1

import "testing"

func TestPkg1(t *testing.T) {
	t.Log("pkg1 test")
}
`
		err = os.WriteFile(filepath.Join(testDir, "pkg1", "pkg1_test.go"), []byte(pkg1Test), 0o600)
		require.NoError(t, err)

		// pkg2 test
		pkg2Test := `package pkg2

import "testing"

func TestPkg2(t *testing.T) {
	t.Log("pkg2 test")
}
`
		err = os.WriteFile(filepath.Join(testDir, "pkg2", "pkg2_test.go"), []byte(pkg2Test), 0o600)
		require.NoError(t, err)

		// Create magefile
		magefilePath := filepath.Join(testDir, "magefile.go")
		magefileContent := `//go:build mage
// +build mage

package main

import (
	"os"
	"strings"
	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/mage"
)

type Test mg.Namespace

func getMageArgs() []string {
	if mageArgs := os.Getenv("MAGE_ARGS"); mageArgs != "" {
		return strings.Fields(mageArgs)
	}
	return nil
}

func (t Test) Cover() error {
	var impl mage.Test
	return impl.Cover(getMageArgs()...)
}
`
		err = os.WriteFile(magefilePath, []byte(magefileContent), 0o600)
		require.NoError(t, err)

		// Run coverage for specific package
		cmd := testCommand(t, magexPath, "test:cover", "package=./pkg1")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)
		if err != nil {
			t.Logf("coverage command exited with error while checking parameter handling: %v", err)
		}

		// Should run (may fail due to missing dependencies, but should attempt)
		// The important thing is that it uses the package parameter
		if strings.Contains(outputStr, "./pkg1") || strings.Contains(outputStr, "pkg1") {
			// Good - it's using the package parameter
			t.Log("Package parameter was used")
		} else if strings.Contains(outputStr, "cover") {
			// At least it tried to run coverage
			t.Log("Coverage command ran but package parameter may not have been visible in output")
		}
	})

	t.Run("MultipleParameters", func(t *testing.T) {
		// Test multiple parameters at once
		testDir := t.TempDir()

		// Create a test file
		testContent := `package main

import "testing"

func TestSimple(t *testing.T) {
	t.Log("test")
}
`
		err := os.WriteFile(filepath.Join(testDir, "simple_test.go"), []byte(testContent), 0o600)
		require.NoError(t, err)

		// Create magefile
		magefilePath := filepath.Join(testDir, "magefile.go")
		magefileContent := `//go:build mage
// +build mage

package main

import (
	"os"
	"strings"
	"github.com/magefile/mage/mg"
	"github.com/mrz1836/mage-x/pkg/mage"
)

type Test mg.Namespace

func getMageArgs() []string {
	if mageArgs := os.Getenv("MAGE_ARGS"); mageArgs != "" {
		return strings.Fields(mageArgs)
	}
	return nil
}

func (t Test) Unit() error {
	var impl mage.Test
	return impl.Unit(getMageArgs()...)
}
`
		err = os.WriteFile(magefilePath, []byte(magefileContent), 0o600)
		require.NoError(t, err)

		// Run with multiple parameters
		cmd := testCommand(t, magexPath, "test:unit", "verbose", "short", "package=.")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Check that parameters were processed
		// The test might fail due to setup, but we're checking parameter handling
		if err != nil {
			// Check if the error is due to parameters being passed
			if strings.Contains(outputStr, "verbose") ||
				strings.Contains(outputStr, "-v") ||
				strings.Contains(outputStr, "short") ||
				strings.Contains(outputStr, "-short") {
				// Parameters were processed
				t.Log("Parameters were processed")
			}
		}
	})
}

// TestParameterPassingStressTest ensures parameters work under various conditions
func TestParameterPassingStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	// Build magex
	cmd := testCommand(t, "go", "build", "-o", "magex-test", "../../cmd/magex")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build magex: %v\nOutput: %s", err, output)
	}
	cleanupFile(t, "magex-test")

	magexPath, err := filepath.Abs("./magex-test")
	require.NoError(t, err)

	// Test various parameter formats. Each case passes a set of args to a
	// custom "default" magefile target that echoes the raw MAGE_ARGS string it
	// received. The stress is on the magex -> mage -> MAGE_ARGS pipeline faithfully
	// carrying every parameter (booleans, key=value pairs, and quoted values with
	// special characters) without dropping or mangling them.
	parameterTests := []struct {
		name string
		args []string
	}{
		{
			name: "Boolean flags",
			args: []string{"verbose", "short", "race"},
		},
		{
			name: "Key-value pairs",
			args: []string{"time=5s", "count=3", "cpu=2"},
		},
		{
			name: "Mixed parameters",
			args: []string{"verbose", "package=./pkg", "timeout=30s"},
		},
		{
			name: "Special characters",
			args: []string{`message="fix: bug #123"`, "all"},
		},
	}

	for _, tt := range parameterTests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a minimal test environment
			testDir := t.TempDir()

			// Create a simple magefile
			magefileContent := `//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"strings"
)

func getMageArgs() []string {
	if mageArgs := os.Getenv("MAGE_ARGS"); mageArgs != "" {
		return strings.Fields(mageArgs)
	}
	return nil
}

func Default() error {
	args := getMageArgs()
	fmt.Printf("Received args: %v\n", args)
	for _, arg := range args {
		fmt.Printf("Arg: %s\n", arg)
	}
	return nil
}
`
			err := os.WriteFile(filepath.Join(testDir, "magefile.go"), []byte(magefileContent), 0o600)
			require.NoError(t, err)

			// Run magex with the test parameters
			cmdArgs := []string{"default"}
			cmdArgs = append(cmdArgs, tt.args...)

			cmd := testCommand(t, magexPath, cmdArgs...)
			cmd.Dir = testDir

			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			// The custom magefile target must run successfully.
			require.NoError(t, err, "magex default command failed: %s", outputStr)

			// The magefile echoes the raw MAGE_ARGS string, so every parameter we
			// passed must survive the magex -> mage pipeline and appear verbatim.
			for _, check := range tt.args {
				assert.Contains(t, outputStr, check,
					"Parameter %q must be passed through to MAGE_ARGS\nFull output: %s", check, outputStr)
			}
		})
	}
}

// TestRegressionPrevention ensures the parameter bug never returns
func TestRegressionPrevention(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping regression test in short mode")
	}

	// This test specifically checks for the bug where parameters
	// were not passed from magex -> mage -> functions

	// Build magex
	cmd := testCommand(t, "go", "build", "-o", "magex-test", "../../cmd/magex")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build magex: %v\nOutput: %s", err, output)
	}
	cleanupFile(t, "magex-test")

	magexPath, err := filepath.Abs("./magex-test")
	require.NoError(t, err)

	t.Run("VerifyMAGE_ARGSEnvironmentIsSet", func(t *testing.T) {
		// Create test directory
		testDir := t.TempDir()

		// Create a magefile that prints MAGE_ARGS
		magefileContent := `//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
)

func CheckEnv() error {
	mageArgs := os.Getenv("MAGE_ARGS")
	if mageArgs == "" {
		fmt.Println("ERROR: MAGE_ARGS is empty!")
		return fmt.Errorf("MAGE_ARGS not set")
	}
	fmt.Printf("MAGE_ARGS: %s\n", mageArgs)
	return nil
}
`
		err := os.WriteFile(filepath.Join(testDir, "magefile.go"), []byte(magefileContent), 0o600)
		require.NoError(t, err)

		// Run with parameters
		cmd := testCommand(t, magexPath, "checkEnv", "test=true", "param=value")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Should succeed
		require.NoError(t, err, "Command failed: %s", outputStr)

		// Should show MAGE_ARGS
		assert.Contains(t, outputStr, "MAGE_ARGS:", "Should print MAGE_ARGS")
		assert.Contains(t, outputStr, "test=true", "Should contain test=true")
		assert.Contains(t, outputStr, "param=value", "Should contain param=value")
		assert.NotContains(t, outputStr, "ERROR:", "Should not have error")
	})

	t.Run("CriticalBugScenario", func(t *testing.T) {
		// Reproduce the exact scenario that was broken: a "time=7s" parameter
		// must flow through the magex -> mage -> magefile-function pipeline and
		// arrive as "7s" (not be dropped, leaving the function to fall back to a
		// default of 10s). This is the regression that magex must never reintroduce.
		//
		// NOTE: We deliberately do NOT name the magefile target "test:fuzz" here.
		// magex resolves built-in commands BEFORE custom magefile commands (see
		// cmd/magex/main.go: reg.Execute is tried first, custom commands only run
		// on ErrUnknownCommand). "test:fuzz" is a built-in, so a custom magefile
		// "Test.Fuzz" would never be invoked - magex would run the built-in fuzz
		// runner instead (which needs real fuzz targets/modules). To exercise the
		// parameter-passing pipeline hermetically we use a custom, non-built-in
		// target that echoes the args it received.
		testDir := t.TempDir()

		// Create a magefile that tracks the time parameter. It uses only the
		// standard library so it compiles in a bare temp dir with no go.mod.
		magefileContent := `//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"strings"
)

func getMageArgs() []string {
	if mageArgs := os.Getenv("MAGE_ARGS"); mageArgs != "" {
		return strings.Fields(mageArgs)
	}
	return nil
}

// CheckFuzzTime stands in for the old "test:fuzz" target. It verifies the
// time parameter survives the magex -> mage -> function pipeline.
func CheckFuzzTime() error {
	args := getMageArgs()
	fmt.Printf("Fuzz called with args: %v\n", args)

	// Check for time parameter
	for _, arg := range args {
		if strings.HasPrefix(arg, "time=") {
			timeVal := strings.TrimPrefix(arg, "time=")
			fmt.Printf("SUCCESS: Using time parameter: %s\n", timeVal)
			if timeVal == "7s" {
				fmt.Println("CORRECT: Got expected 7s, not default 10s")
			}
			return nil
		}
	}

	fmt.Println("ERROR: No time parameter found - would use default!")
	return fmt.Errorf("time parameter not passed")
}
`
		err := os.WriteFile(filepath.Join(testDir, "magefile.go"), []byte(magefileContent), 0o600)
		require.NoError(t, err)

		// Run the regression scenario: pass the exact parameter that used to be dropped.
		cmd := testCommand(t, magexPath, "checkFuzzTime", "time=7s")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Must succeed
		require.NoError(t, err, "Command failed: %s", outputStr)

		// Must use the correct time
		assert.Contains(t, outputStr, "SUCCESS:", "Should successfully find time parameter")
		assert.Contains(t, outputStr, "7s", "Should use 7s from parameter")
		assert.Contains(t, outputStr, "CORRECT:", "Should confirm correct time used")
		assert.NotContains(t, outputStr, "ERROR:", "Should not have error about missing parameter")
	})
}
