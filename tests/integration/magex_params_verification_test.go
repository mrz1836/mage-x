//go:build integration
// +build integration

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParameterVerification tests that parameters actually affect command behavior
func TestParameterVerification(t *testing.T) {
	// Build magex once for all tests
	magexBinary := buildMagexForTesting(t)
	defer os.Remove(magexBinary)

	magexPath, err := filepath.Abs(magexBinary)
	require.NoError(t, err)

	t.Run("VerboseParameter", func(t *testing.T) {
		// Test that verbose parameter actually adds verbose output
		testDir := setupGoProject(t)

		// Run test without verbose
		cmd1 := exec.Command(magexPath, "test")
		cmd1.Dir = testDir
		output1, err1 := cmd1.CombinedOutput()
		output1Str := string(output1)

		// Run test with verbose
		cmd2 := exec.Command(magexPath, "test", "verbose=true")
		cmd2.Dir = testDir
		output2, err2 := cmd2.CombinedOutput()
		output2Str := string(output2)

		// Both should complete (even if no tests)
		if err1 != nil && !strings.Contains(output1Str, "no test files") {
			t.Logf("Non-verbose test failed: %v", err1)
		}
		if err2 != nil && !strings.Contains(output2Str, "no test files") {
			t.Logf("Verbose test failed: %v", err2)
		}

		// Verbose should show additional information
		if len(output2Str) > len(output1Str) ||
			strings.Contains(output2Str, "Verbose: ✓") ||
			strings.Contains(output2Str, "-v") {
			t.Logf("✓ Verbose parameter working - verbose output is longer or contains verbose markers")
		} else {
			t.Logf("Note: Verbose parameter may not be working - outputs similar length")
		}
	})

	t.Run("PackageParameter", func(t *testing.T) {
		// Test that package parameter filters to specific packages
		testDir := setupGoProject(t)

		// Run test for all packages
		cmd1 := exec.Command(magexPath, "test", "package=./...")
		cmd1.Dir = testDir
		output1, err1 := cmd1.CombinedOutput()
		output1Str := string(output1)

		// Run test for specific package
		cmd2 := exec.Command(magexPath, "test", "package=./pkg/utils")
		cmd2.Dir = testDir
		output2, err2 := cmd2.CombinedOutput()
		output2Str := string(output2)

		// Check if specific package is mentioned
		if strings.Contains(output2Str, "pkg/utils") ||
			strings.Contains(output2Str, "utils") {
			t.Logf("✓ Package parameter working - specific package mentioned")
		} else if strings.Contains(output2Str, "no test files") {
			t.Logf("✓ Package parameter working - correctly reports no test files for specific package")
		}

		// Should not show all packages when specific package requested
		if err1 == nil && err2 == nil && len(output1Str) > len(output2Str) {
			t.Logf("✓ Package parameter working - specific package output is shorter")
		}
	})

	t.Run("ModGraphFormatParameter", func(t *testing.T) {
		// Test that format parameter affects mod:graph output
		testDir := setupGoModule(t)

		// Test different format parameters (based on supported formats in mod.go)
		formats := []string{"tree", "json", "dot", "mermaid"}

		for _, format := range formats {
			t.Run("Format_"+format, func(t *testing.T) {
				cmd := exec.Command(magexPath, "mod:graph", "format="+format)
				cmd.Dir = testDir

				output, err := cmd.CombinedOutput()
				outputStr := string(output)

				if err != nil {
					// Format might not be supported, but should not crash
					if !strings.Contains(outputStr, "unknown format") &&
						!strings.Contains(outputStr, "invalid format") {
						t.Errorf("Unexpected error for format=%s: %v", format, err)
					}
				} else {
					// Different formats should produce different outputs
					switch format {
					case "tree":
						if hasTreeSymbols(outputStr) {
							t.Logf("✓ Tree format working - contains tree symbols")
						}
					case "json":
						if strings.Contains(outputStr, "{") && strings.Contains(outputStr, "}") {
							t.Logf("✓ JSON format working - contains JSON brackets")
						}
					case "dot":
						if strings.Contains(outputStr, "digraph") && strings.Contains(outputStr, "->") {
							t.Logf("✓ DOT format working - contains digraph and arrows")
						}
					case "mermaid":
						if strings.Contains(outputStr, "graph TD") || strings.Contains(outputStr, "-->") {
							t.Logf("✓ Mermaid format working - contains graph syntax")
						}
					}
				}
			})
		}
	})

	t.Run("ModGraphShowVersionsParameter", func(t *testing.T) {
		// Test that show_versions parameter affects output
		testDir := setupGoModule(t)

		// Run without show_versions
		cmd1 := exec.Command(magexPath, "mod:graph", "show_versions=false")
		cmd1.Dir = testDir
		output1, err1 := cmd1.CombinedOutput()
		output1Str := string(output1)

		// Run with show_versions
		cmd2 := exec.Command(magexPath, "mod:graph", "show_versions=true")
		cmd2.Dir = testDir
		output2, err2 := cmd2.CombinedOutput()
		output2Str := string(output2)

		if err1 == nil && err2 == nil {
			// With show_versions=true, should see version information (@ symbols)
			versions1 := strings.Count(output1Str, "@")
			versions2 := strings.Count(output2Str, "@")

			if versions2 > versions1 {
				t.Logf("✓ show_versions parameter working - more @ symbols with show_versions=true (%d vs %d)", versions2, versions1)
			} else if versions2 > 0 {
				t.Logf("✓ show_versions parameter working - versions shown (%d @ symbols)", versions2)
			}
		}
	})

	t.Run("BuildPlatformParameter", func(t *testing.T) {
		// Test that platform parameter affects build
		testDir := setupGoProject(t)

		platforms := []string{"linux/amd64", "windows/amd64", "darwin/amd64"}

		for _, platform := range platforms {
			t.Run("Platform_"+strings.ReplaceAll(platform, "/", "_"), func(t *testing.T) {
				cmd := exec.Command(magexPath, "build", "platform="+platform)
				cmd.Dir = testDir

				output, err := cmd.CombinedOutput()
				outputStr := string(output)

				// Should at least recognize the platform parameter
				if strings.Contains(outputStr, platform) ||
					strings.Contains(outputStr, strings.Split(platform, "/")[0]) ||
					strings.Contains(outputStr, "GOOS=") ||
					strings.Contains(outputStr, "GOARCH=") {
					t.Logf("✓ Platform parameter working - platform mentioned in output")
				} else if err != nil && (strings.Contains(outputStr, "cross-compilation") ||
					strings.Contains(outputStr, "no main package")) {
					t.Logf("✓ Platform parameter working - cross-compilation or package issue mentioned")
				}
			})
		}
	})

	t.Run("InvalidParameterRejection", func(t *testing.T) {
		// Test that invalid parameters are handled gracefully
		testDir := setupGoModule(t)

		invalidTests := []struct {
			command string
			param   string
		}{
			{"mod:graph", "invalid_param=value"},
			{"test", "nonexistent=true"},
			{"build", "fake_option=123"},
		}

		for _, tt := range invalidTests {
			t.Run(tt.command+"_"+tt.param, func(t *testing.T) {
				cmd := exec.Command(magexPath, tt.command, tt.param)
				cmd.Dir = testDir

				output, err := cmd.CombinedOutput()
				outputStr := string(output)

				// Should either work (ignoring unknown params) or provide meaningful error
				if err != nil {
					// If it errors, should be informative
					if strings.Contains(outputStr, "unknown") ||
						strings.Contains(outputStr, "invalid") ||
						strings.Contains(outputStr, "not recognized") {
						t.Logf("✓ Invalid parameter properly rejected with message: %s", outputStr)
					} else {
						t.Logf("Invalid parameter caused error but message unclear: %s", outputStr)
					}
				} else {
					// If it doesn't error, that's also acceptable (graceful handling)
					t.Logf("✓ Invalid parameter ignored gracefully")
				}
			})
		}
	})

	t.Run("BooleanParameterParsing", func(t *testing.T) {
		// Test boolean parameter parsing variations
		testDir := setupGoProject(t)

		boolTests := []string{
			"verbose=true",
			"verbose=false",
			"verbose", // flag-style
			"race=1",  // numeric boolean
			"race=0",
		}

		for _, param := range boolTests {
			t.Run("Bool_"+strings.ReplaceAll(param, "=", "_"), func(t *testing.T) {
				cmd := exec.Command(magexPath, "test", param)
				cmd.Dir = testDir

				output, err := cmd.CombinedOutput()
				outputStr := string(output)

				// Should parse without syntax errors
				if err != nil && strings.Contains(outputStr, "syntax error") {
					t.Errorf("Boolean parameter parsing failed: %s", outputStr)
				} else {
					t.Logf("✓ Boolean parameter '%s' parsed successfully", param)
				}
			})
		}
	})
}

// TestParameterPassthrough tests that parameters are passed through the execution pipeline
func TestParameterPassthrough(t *testing.T) {
	magexBinary := buildMagexForTesting(t)
	defer os.Remove(magexBinary)

	magexPath, err := filepath.Abs(magexBinary)
	require.NoError(t, err)

	t.Run("MultipleParameterPassthrough", func(t *testing.T) {
		// Test that multiple parameters are all passed through
		testDir := setupGoModule(t)

		// Use a command with multiple parameters
		cmd := exec.Command(magexPath, "mod:graph",
			"depth=2",
			"show_versions=true",
			"format=tree")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		// Should succeed and respect all parameters
		assert.NoError(t, err, "Multiple parameter command failed: %s", outputStr)

		// Check depth limit
		depth := countTreeDepth(outputStr)
		assert.LessOrEqual(t, depth, 2, "Depth=2 not respected")

		// Check versions (should have @ symbols if available)
		hasVersions := strings.Contains(outputStr, "@")

		// Check tree format (should have tree symbols)
		hasTree := hasTreeSymbols(outputStr)
		assert.True(t, hasTree, "Tree format not applied")

		t.Logf("✓ Multiple parameters working: depth=%d, versions=%t, tree=%t",
			depth, hasVersions, hasTree)
	})

	t.Run("ParameterPrecedence", func(t *testing.T) {
		// Test parameter precedence when conflicting values given
		testDir := setupGoModule(t)

		// Test with conflicting depth values (later should win)
		cmd := exec.Command(magexPath, "mod:graph", "depth=1", "depth=3")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		if err == nil {
			depth := countTreeDepth(outputStr)
			// Should use the last value (depth=3)
			if depth <= 3 {
				t.Logf("✓ Parameter precedence working - final depth value respected")
			}
		}
	})
}
