//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/testhelpers"
)

const integrationCommandTimeout = 2 * time.Minute

func TestMain(m *testing.M) {
	os.Exit(runTestMain(m))
}

func runTestMain(m *testing.M) int {
	cacheRoot, err := os.MkdirTemp("", "magex-integration-cache-*")
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create integration cache root: %v\n", err)
		return 1
	}
	defer func() {
		if removeErr := os.RemoveAll(cacheRoot); removeErr != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to remove integration cache root: %v\n", removeErr)
		}
	}()

	cacheDirs := map[string]string{
		"GOCACHE":        filepath.Join(cacheRoot, "go-build"),
		"MAGEFILE_CACHE": filepath.Join(cacheRoot, "magefile"),
	}
	for key, dir := range cacheDirs {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to create %s directory: %v\n", key, err)
			return 1
		}
		if err := os.Setenv(key, dir); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to set %s: %v\n", key, err)
			return 1
		}
	}
	if err := os.Setenv("GOTELEMETRY", "off"); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to disable Go telemetry: %v\n", err)
		return 1
	}

	// Integration tests shell out to magex, go and git; keep all of that off the
	// network so a test run never reaches the internet.
	if err := testhelpers.BlockExternalNetwork(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to block external network: %v\n", err)
		return 1
	}

	return m.Run()
}

// fixtureDepVersion is the version every fixture dependency is required at. The
// modules only exist on disk, so the exact value does not matter - it just has
// to be valid semver for the require directives.
const fixtureDepVersion = "v1.0.0"

// writeModuleGraphFixture creates a Go module whose dependency graph resolves
// entirely from disk: the module requires the head of a chain, each chain entry
// requires the next, and the module replaces every one of them with a directory
// written next to it. Nothing is ever fetched, so `go mod graph` succeeds with
// GOPROXY=off and renders the same tree on every machine - a chain of length n
// produces exactly n levels below the root. The returned path is the main
// module's directory.
//
// The chain modules live in a sibling directory rather than under the main
// module so magex's multi-module discovery, which walks down from the working
// directory, does not pick them up as extra modules to build and test.
//
// They declare go 1.16 deliberately: module graph pruning (go 1.17 and later)
// drops the requirements of indirect modules, which would flatten the tree to a
// single level and make the depth assertions meaningless.
func writeModuleGraphFixture(t *testing.T, length int) string {
	t.Helper()

	fixtureRoot := t.TempDir()
	moduleDir := filepath.Join(fixtureRoot, "testmodule")
	require.NoError(t, os.MkdirAll(moduleDir, 0o750))

	deps := make([]string, length)
	for i := range deps {
		deps[i] = fmt.Sprintf("example.com/dep/d%d", i+1)
	}

	var mainMod strings.Builder
	fmt.Fprintf(&mainMod, "module testmodule\n\ngo 1.21\n\nrequire %s %s\n\n", deps[0], fixtureDepVersion)

	for i, dep := range deps {
		fmt.Fprintf(&mainMod, "replace %s => ../deps/d%d\n", dep, i+1)

		depDir := filepath.Join(fixtureRoot, "deps", fmt.Sprintf("d%d", i+1))
		require.NoError(t, os.MkdirAll(depDir, 0o750))

		var depMod strings.Builder
		fmt.Fprintf(&depMod, "module %s\n\ngo 1.16\n", dep)
		if i+1 < len(deps) {
			fmt.Fprintf(&depMod, "\nrequire %s %s\n", deps[i+1], fixtureDepVersion)
		}
		require.NoError(t, os.WriteFile(filepath.Join(depDir, "go.mod"), []byte(depMod.String()), 0o600))
	}

	require.NoError(t, os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte(mainMod.String()), 0o600))

	return moduleDir
}

func testCommand(t *testing.T, name string, args ...string) *exec.Cmd {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), integrationCommandTimeout)
	t.Cleanup(cancel)

	return exec.CommandContext(ctx, name, args...) //nolint:gosec // integration tests execute fixed, test-controlled commands.
}

func cleanupFile(t *testing.T, path string) {
	t.Helper()

	t.Cleanup(func() {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			t.Errorf("remove %s: %v", path, err)
		}
	})
}

// countTreeDepth accurately counts the maximum depth in a dependency tree output
// Returns the maximum depth found in the tree structure
func countTreeDepth(output string) int {
	maxDepth := 0

	for _, line := range strings.Split(output, "\n") {
		if !isDependencyTreeLine(line) {
			continue
		}

		if depth := treeLineDepth(line); depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth
}

func isDependencyTreeLine(line string) bool {
	return strings.Contains(line, "├──") || strings.Contains(line, "└──")
}

func treeLineDepth(line string) int {
	depth := 0
	runes := []rune(line)

	for i := 0; i < len(runes); {
		switch {
		case i+4 <= len(runes) && string(runes[i:i+4]) == "│   ":
			depth++
			i += 4
		case i+3 <= len(runes) && string(runes[i:i+3]) == "│  ":
			depth++
			i += 3
		case runes[i] == '├' || runes[i] == '└':
			return depth + 1
		case runes[i] == ' ' || runes[i] == '\t':
			i++
		default:
			return depth
		}
	}

	return depth
}

// hasTreeSymbols checks if the output contains dependency tree symbols
func hasTreeSymbols(output string) bool {
	return strings.Contains(output, "├──") ||
		strings.Contains(output, "└──") ||
		strings.Contains(output, "│")
}

// countDependencyLines counts how many lines contain actual dependencies (not headers)
func countDependencyLines(output string) int {
	lines := strings.Split(output, "\n")
	count := 0

	for _, line := range lines {
		// Count lines that contain dependency names but aren't headers
		if (strings.Contains(line, "├──") || strings.Contains(line, "└──")) &&
			!strings.Contains(line, "Dependency Tree:") &&
			!strings.Contains(line, "testmodule") {
			count++
		}
	}

	return count
}

// compareTreeOutputs compares two tree outputs and returns analysis
func compareTreeOutputs(output1, output2 string) (deeper, moreLines bool, analysis string) {
	depth1 := countTreeDepth(output1)
	depth2 := countTreeDepth(output2)

	lines1 := len(strings.Split(output1, "\n"))
	lines2 := len(strings.Split(output2, "\n"))

	deps1 := countDependencyLines(output1)
	deps2 := countDependencyLines(output2)

	deeper = depth2 > depth1
	moreLines = lines2 > lines1

	analysis = fmt.Sprintf("Comparison:\n  Output1: depth=%d, lines=%d, deps=%d\n  Output2: depth=%d, lines=%d, deps=%d",
		depth1, lines1, deps1, depth2, lines2, deps2)

	return deeper, moreLines, analysis
}
