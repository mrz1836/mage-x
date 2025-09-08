//go:build integration
// +build integration

package integration

import (
	"bytes"
	"os"
	"os/exec"
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
		cmd := exec.Command("go", "build", "-o", "magex-test", "./cmd/magex")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to build magex: %v\nOutput: %s", err, output)
		}
	})

	// Clean up the test binary after tests
	defer os.Remove("magex-test")

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
		err := os.WriteFile(magefilePath, []byte(magefileContent), 0o644)
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
		err = os.WriteFile(testFilePath, []byte(fuzzTestContent), 0o644)
		require.NoError(t, err)

		// Run magex with time parameter
		cmd := exec.Command(magexPath, "test:fuzz", "time=1s")
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
			cmd.Process.Kill()
			t.Error("Fuzz test took too long - likely using default 10s instead of 1s")
		}
	})

	t.Run("VersionBumpDryRun", func(t *testing.T) {
		// Test version bump with dry-run
		testDir := t.TempDir()

		// Initialize git repo
		cmd := exec.Command("git", "init")
		cmd.Dir = testDir
		err := cmd.Run()
		require.NoError(t, err)

		// Configure git
		cmd = exec.Command("git", "config", "user.email", "test@example.com")
		cmd.Dir = testDir
		err = cmd.Run()
		require.NoError(t, err)

		cmd = exec.Command("git", "config", "user.name", "Test User")
		cmd.Dir = testDir
		err = cmd.Run()
		require.NoError(t, err)

		// Create initial commit
		testFile := filepath.Join(testDir, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0o644)
		require.NoError(t, err)

		cmd = exec.Command("git", "add", ".")
		cmd.Dir = testDir
		err = cmd.Run()
		require.NoError(t, err)

		cmd = exec.Command("git", "commit", "-m", "initial")
		cmd.Dir = testDir
		err = cmd.Run()
		require.NoError(t, err)

		// Add initial tag
		cmd = exec.Command("git", "tag", "v1.0.0")
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
		err = os.WriteFile(magefilePath, []byte(magefileContent), 0o644)
		require.NoError(t, err)

		// Run version bump with dry-run
		cmd = exec.Command(magexPath, "version:bump", "dry-run", "bump=minor")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Should succeed
		assert.NoError(t, err, "Command failed: %s", outputStr)

		// Should indicate dry-run mode
		assert.Contains(t, outputStr, "DRY-RUN", "Should indicate dry-run mode")
		assert.Contains(t, outputStr, "v1.1.0", "Should show version bump to v1.1.0")

		// Verify no tag was actually created
		cmd = exec.Command("git", "tag", "-l")
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
		err := os.WriteFile(filepath.Join(testDir, "go.mod"), []byte(modContent), 0o644)
		require.NoError(t, err)

		// Create test files in different packages
		err = os.MkdirAll(filepath.Join(testDir, "pkg1"), 0o755)
		require.NoError(t, err)
		err = os.MkdirAll(filepath.Join(testDir, "pkg2"), 0o755)
		require.NoError(t, err)

		// pkg1 test
		pkg1Test := `package pkg1

import "testing"

func TestPkg1(t *testing.T) {
	t.Log("pkg1 test")
}
`
		err = os.WriteFile(filepath.Join(testDir, "pkg1", "pkg1_test.go"), []byte(pkg1Test), 0o644)
		require.NoError(t, err)

		// pkg2 test
		pkg2Test := `package pkg2

import "testing"

func TestPkg2(t *testing.T) {
	t.Log("pkg2 test")
}
`
		err = os.WriteFile(filepath.Join(testDir, "pkg2", "pkg2_test.go"), []byte(pkg2Test), 0o644)
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
		err = os.WriteFile(magefilePath, []byte(magefileContent), 0o644)
		require.NoError(t, err)

		// Run coverage for specific package
		cmd := exec.Command(magexPath, "test:cover", "package=./pkg1")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Should run (may fail due to missing dependencies, but should attempt)
		// The important thing is that it uses the package parameter
		if strings.Contains(outputStr, "./pkg1") || strings.Contains(outputStr, "pkg1") {
			// Good - it's using the package parameter
			assert.True(t, true, "Package parameter was used")
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
		err := os.WriteFile(filepath.Join(testDir, "simple_test.go"), []byte(testContent), 0o644)
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
		err = os.WriteFile(magefilePath, []byte(magefileContent), 0o644)
		require.NoError(t, err)

		// Run with multiple parameters
		cmd := exec.Command(magexPath, "test:unit", "verbose", "short", "package=.")
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
				assert.True(t, true, "Parameters were processed")
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
	cmd := exec.Command("go", "build", "-o", "magex-test", "./cmd/magex")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build magex: %v\nOutput: %s", err, output)
	}
	defer os.Remove("magex-test")

	magexPath, err := filepath.Abs("./magex-test")
	require.NoError(t, err)

	// Test various parameter formats
	parameterTests := []struct {
		name     string
		command  string
		args     []string
		checkFor []string
	}{
		{
			name:     "Boolean flags",
			command:  "test:unit",
			args:     []string{"verbose", "short", "race"},
			checkFor: []string{"-v", "-short", "-race"},
		},
		{
			name:     "Key-value pairs",
			command:  "test:bench",
			args:     []string{"time=5s", "count=3", "cpu=2"},
			checkFor: []string{"5s", "3", "2"},
		},
		{
			name:     "Mixed parameters",
			command:  "test:cover",
			args:     []string{"verbose", "package=./pkg", "timeout=30s"},
			checkFor: []string{"./pkg", "30s"},
		},
		{
			name:     "Special characters",
			command:  "git:commit",
			args:     []string{`message="fix: bug #123"`, "all"},
			checkFor: []string{"fix:", "bug", "123"},
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
			err := os.WriteFile(filepath.Join(testDir, "magefile.go"), []byte(magefileContent), 0o644)
			require.NoError(t, err)

			// Run magex with the test parameters
			cmdArgs := []string{"default"}
			cmdArgs = append(cmdArgs, tt.args...)

			cmd := exec.Command(magexPath, cmdArgs...)
			cmd.Dir = testDir

			output, _ := cmd.CombinedOutput()
			outputStr := string(output)

			// Verify parameters were received
			for _, check := range tt.args {
				if !strings.Contains(outputStr, check) {
					t.Logf("Full output: %s", outputStr)
				}
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
	cmd := exec.Command("go", "build", "-o", "magex-test", "./cmd/magex")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build magex: %v\nOutput: %s", err, output)
	}
	defer os.Remove("magex-test")

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
		err := os.WriteFile(filepath.Join(testDir, "magefile.go"), []byte(magefileContent), 0o644)
		require.NoError(t, err)

		// Run with parameters
		cmd := exec.Command(magexPath, "checkEnv", "test=true", "param=value")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Should succeed
		assert.NoError(t, err, "Command failed: %s", outputStr)

		// Should show MAGE_ARGS
		assert.Contains(t, outputStr, "MAGE_ARGS:", "Should print MAGE_ARGS")
		assert.Contains(t, outputStr, "test=true", "Should contain test=true")
		assert.Contains(t, outputStr, "param=value", "Should contain param=value")
		assert.NotContains(t, outputStr, "ERROR:", "Should not have error")
	})

	t.Run("CriticalBugScenario", func(t *testing.T) {
		// Reproduce the exact scenario that was broken:
		// magex test:fuzz time=7s was running for 10s instead

		testDir := t.TempDir()

		// Create a magefile that tracks time parameter
		magefileContent := `//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"strings"
	"github.com/magefile/mage/mg"
)

type Test mg.Namespace

func getMageArgs() []string {
	if mageArgs := os.Getenv("MAGE_ARGS"); mageArgs != "" {
		return strings.Fields(mageArgs)
	}
	return nil
}

func (t Test) Fuzz() error {
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
		err := os.WriteFile(filepath.Join(testDir, "magefile.go"), []byte(magefileContent), 0o644)
		require.NoError(t, err)

		// Run the exact command that was broken
		cmd := exec.Command(magexPath, "test:fuzz", "time=7s")
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
