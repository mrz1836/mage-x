//go:build integration
// +build integration

package integration

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModGraphDepthParameter tests the depth parameter extensively
func TestModGraphDepthParameter(t *testing.T) {
	// Build magex once for all tests
	magexBinary := buildMagexForTesting(t)
	defer os.Remove(magexBinary)

	magexPath, err := filepath.Abs(magexBinary)
	require.NoError(t, err)

	// Setup a test module with dependencies for depth testing
	testDir := setupGoModuleWithDependencies(t)

	t.Run("DepthZero", func(t *testing.T) {
		// Test depth=0 - means unlimited depth (show all dependencies)
		cmd := exec.Command(magexPath, "mod:graph", "depth=0")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		assert.NoError(t, err, "Command failed: %s", outputStr)

		// With depth=0 (unlimited), should show full dependency tree
		maxDepth := countTreeDepth(outputStr)
		assert.GreaterOrEqual(t, maxDepth, 1, "Depth=0 (unlimited) should show dependencies, but saw depth %d", maxDepth)

		// Should contain tree symbols for dependencies since it's unlimited
		assert.True(t, hasTreeSymbols(outputStr), "Depth=0 (unlimited) should show dependency tree symbols")
	})

	t.Run("DepthOne", func(t *testing.T) {
		// Test depth=1 - should show root + direct dependencies
		cmd := exec.Command(magexPath, "mod:graph", "depth=1")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		assert.NoError(t, err, "Command failed: %s", outputStr)

		// With depth=1, should see at most depth 1
		maxDepth := countTreeDepth(outputStr)
		assert.LessOrEqual(t, maxDepth, 1, "Depth=1 should show at most depth 1, but saw depth %d", maxDepth)
		assert.Greater(t, maxDepth, 0, "Depth=1 should show some dependencies, but saw depth %d", maxDepth)

		// Should contain tree symbols for direct dependencies
		assert.True(t, strings.Contains(outputStr, "├──") || strings.Contains(outputStr, "└──"),
			"Depth=1 should show dependency tree symbols")
	})

	t.Run("DepthThree", func(t *testing.T) {
		// Test depth=3 - should show 3 levels of dependencies
		cmd := exec.Command(magexPath, "mod:graph", "depth=3")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		assert.NoError(t, err, "Command failed: %s", outputStr)

		// With depth=3, should see at most depth 3
		maxDepth := countTreeDepth(outputStr)
		assert.LessOrEqual(t, maxDepth, 3, "Depth=3 should show at most depth 3, but saw depth %d", maxDepth)

		// Should show more dependencies than depth=1
		assert.Greater(t, maxDepth, 0, "Depth=3 should show some dependencies")
	})

	t.Run("DepthComparison", func(t *testing.T) {
		// Compare depth=1 vs depth=3 to ensure they're different

		// Get output for depth=1
		cmd1 := exec.Command(magexPath, "mod:graph", "depth=1")
		cmd1.Dir = testDir
		output1, err1 := cmd1.CombinedOutput()
		require.NoError(t, err1, "Depth=1 command failed")

		depth1 := countTreeDepth(string(output1))
		lines1 := len(strings.Split(string(output1), "\n"))

		// Get output for depth=3
		cmd3 := exec.Command(magexPath, "mod:graph", "depth=3")
		cmd3.Dir = testDir
		output3, err3 := cmd3.CombinedOutput()
		require.NoError(t, err3, "Depth=3 command failed")

		depth3 := countTreeDepth(string(output3))
		lines3 := len(strings.Split(string(output3), "\n"))

		// Depth=3 should show deeper nesting or more lines
		if depth3 > depth1 {
			t.Logf("✓ Depth parameter working: depth=1 shows max depth %d, depth=3 shows max depth %d", depth1, depth3)
		} else if lines3 > lines1 {
			t.Logf("✓ Depth parameter working: depth=1 shows %d lines, depth=3 shows %d lines", lines1, lines3)
		} else {
			t.Errorf("Depth parameter not working correctly: depth=1 (depth=%d, lines=%d) vs depth=3 (depth=%d, lines=%d)",
				depth1, lines1, depth3, lines3)
		}

		// At minimum, depth should be properly limited
		assert.LessOrEqual(t, depth1, 1, "Depth=1 exceeded limit")
		assert.LessOrEqual(t, depth3, 3, "Depth=3 exceeded limit")
	})

	t.Run("InvalidDepth", func(t *testing.T) {
		// Test invalid depth values
		invalidDepths := []string{"-1", "abc", "999999"}

		for _, invalidDepth := range invalidDepths {
			t.Run("Invalid_"+invalidDepth, func(t *testing.T) {
				cmd := exec.Command(magexPath, "mod:graph", "depth="+invalidDepth)
				cmd.Dir = testDir

				output, err := cmd.CombinedOutput()
				outputStr := string(output)

				// Should either error or default to reasonable behavior
				if err == nil {
					// If it doesn't error, it should at least produce some output
					assert.NotEmpty(t, outputStr, "Should produce some output even with invalid depth")

					// And the depth should be reasonable (not unlimited)
					maxDepth := countTreeDepth(outputStr)
					assert.LessOrEqual(t, maxDepth, 10, "Invalid depth should not produce unlimited depth %d", maxDepth)
				}
			})
		}
	})

	t.Run("LargeDepth", func(t *testing.T) {
		// Test very large depth value - should show all dependencies but not crash
		cmd := exec.Command(magexPath, "mod:graph", "depth=100")
		cmd.Dir = testDir

		output, err := cmd.CombinedOutput()
		outputStr := string(output)

		assert.NoError(t, err, "Large depth should not crash: %s", outputStr)
		assert.NotEmpty(t, outputStr, "Should produce output with large depth")

		// Should show dependencies but be reasonable
		maxDepth := countTreeDepth(outputStr)
		assert.LessOrEqual(t, maxDepth, 20, "Large depth should be capped at reasonable level, saw %d", maxDepth)
	})
}

// TestDepthParameterPrecision tests that depth parameter is precisely respected
func TestDepthParameterPrecision(t *testing.T) {
	magexBinary := buildMagexForTesting(t)
	defer os.Remove(magexBinary)

	magexPath, err := filepath.Abs(magexBinary)
	require.NoError(t, err)

	testDir := setupGoModuleWithDependencies(t)

	// Test each depth value precisely
	depthTests := []struct {
		depth    int
		maxDepth int
		name     string
	}{
		{0, 20, "unlimited"}, // depth=0 means unlimited, so allow up to 20 levels
		{1, 1, "one"},
		{2, 2, "two"},
		{3, 3, "three"},
		{5, 5, "five"},
	}

	for _, tt := range depthTests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(magexPath, "mod:graph", fmt.Sprintf("depth=%d", tt.depth))
			cmd.Dir = testDir

			output, err := cmd.CombinedOutput()
			outputStr := string(output)

			assert.NoError(t, err, "Command failed: %s", outputStr)

			actualDepth := countTreeDepth(outputStr)
			assert.LessOrEqual(t, actualDepth, tt.maxDepth,
				"Depth=%d should show at most depth %d, but saw depth %d", tt.depth, tt.maxDepth, actualDepth)

			if tt.depth > 0 {
				// Should show some tree structure unless depth=0
				hasTreeSymbols := strings.Contains(outputStr, "├──") || strings.Contains(outputStr, "└──")
				assert.True(t, hasTreeSymbols || actualDepth == 0, "Depth=%d should show tree symbols", tt.depth)
			}
		})
	}
}

// Helper function to setup a Go module with dependencies for testing
func setupGoModuleWithDependencies(t *testing.T) string {
	t.Helper()

	testDir := t.TempDir()

	// Create go.mod with multiple dependencies for better depth testing
	goModContent := `module testmodule

go 1.21

require (
    github.com/stretchr/testify v1.8.4
    github.com/gorilla/mux v1.8.0
    github.com/spf13/cobra v1.7.0
)
`
	err := os.WriteFile(filepath.Join(testDir, "go.mod"), []byte(goModContent), 0o644)
	require.NoError(t, err)

	// Run go mod download to populate the module graph
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = testDir
	_ = cmd.Run() // Ignore errors as it might fail in CI

	return testDir
}
