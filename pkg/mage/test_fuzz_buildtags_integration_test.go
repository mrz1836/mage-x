//go:build integration
// +build integration

package mage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFuzzWithDiscoveredBuildTagsIntegration fuzzes a real module in which one
// package only builds with a build tag, the way magex test:fuzz does: discover
// the tags, list each package's targets with them, warm up, and fuzz.
func TestFuzzWithDiscoveredBuildTagsIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real fuzz run in short mode")
	}

	dir := t.TempDir()
	writeFuzzModuleFile(t, dir, ".mage.yaml", "test:\n  auto_discover_build_tags: true\n")
	writeFuzzModuleFile(t, dir, "go.mod", "module example.com/fuzzrepro\n\ngo 1.24\n")
	writeFuzzModuleFile(t, dir, "plain/plain.go", "package plain\n\nfunc Echo(s string) string { return s }\n")
	writeFuzzModuleFile(t, dir, "plain/plain_test.go", fuzzTargetSource("", "plain", "FuzzX"))
	writeFuzzModuleFile(t, dir, "tagged/tagged.go", "//go:build mage\n\npackage tagged\n\nfunc Echo(s string) string { return s }\n")
	writeFuzzModuleFile(t, dir, "tagged/tagged_test.go", fuzzTargetSource("//go:build mage\n\n", "tagged", "FuzzY"))
	t.Chdir(dir)

	// Load the module's .mage.yaml and run the real go commands. The magex test
	// runner exports its own discovery settings to `go test` (its exclude list
	// contains "mage"), so clear them to use only the module's config.
	t.Setenv("MAGE_X_AUTO_DISCOVER_BUILD_TAGS", "")
	t.Setenv("MAGE_X_AUTO_DISCOVER_BUILD_TAGS_EXCLUDE", "")
	t.Setenv("MAGE_X_AUTO_DISCOVER_BUILD_TAGS_COMBINE", "")
	TestResetConfig()
	t.Cleanup(TestResetConfig)
	original := GetRunner()
	require.NoError(t, SetRunner(NewSecureCommandRunner()))
	t.Cleanup(func() { require.NoError(t, SetRunner(original)) })

	config, err := GetConfig()
	require.NoError(t, err)
	require.True(t, config.Test.AutoDiscoverBuildTags)

	discovered := displayTestHeader("fuzz", config)
	require.Equal(t, []string{"mage"}, discovered)
	tagSets := testBuildTagPasses(config, discovered)
	require.Equal(t, []string{"mage"}, tagSets)

	packages := findFuzzPackages()
	require.ElementsMatch(t, []string{"example.com/fuzzrepro/plain", "example.com/fuzzrepro/tagged"}, packages)

	results, _ := runFuzzTestsWithResultsCI(config, time.Second, packages, tagSets, true)

	fuzzed := make(map[string]fuzzTestResult, len(results))
	for _, result := range results {
		fuzzed[result.Test] = result
	}
	require.Len(t, fuzzed, 2, "both FuzzX and FuzzY should run")
	for _, name := range []string{"FuzzX", "FuzzY"} {
		result, ok := fuzzed[name]
		require.True(t, ok, "%s should be fuzzed", name)
		require.NoError(t, result.Error, "%s output:\n%s", name, result.Output)
		assert.Contains(t, result.Output, "fuzz: elapsed", "%s should reach the fuzzing phase", name)
	}
	assert.Equal(t, "example.com/fuzzrepro/tagged", fuzzed["FuzzY"].Package)
}

// writeFuzzModuleFile writes a file of the temporary fuzz module.
func writeFuzzModuleFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(name))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o750))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
}

// fuzzTargetSource returns a test file with one fuzz target for package pkg.
func fuzzTargetSource(header, pkg, target string) string {
	return header + "package " + pkg + `

import "testing"

func ` + target + `(f *testing.F) {
	f.Add("seed")
	f.Fuzz(func(t *testing.T, s string) {
		if Echo(s) != s {
			t.Fatalf("Echo(%q) changed its input", s)
		}
	})
}
`
}
