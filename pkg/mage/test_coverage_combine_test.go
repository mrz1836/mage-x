package mage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errSimulatedCoverageRun is returned by coverageWritingRunner when configured to
// fail a coverage pass, so tests can verify error propagation.
var errSimulatedCoverageRun = errors.New("simulated coverage test failure")

// coverageWritingRunner simulates `go test ... -coverprofile=<file> ...` by writing
// a minimal but valid Go coverage profile to the path named in the flag, relative
// to the current working directory — exactly where `go test` would write it when
// invoked inside a module directory. Calls without a -coverprofile flag (e.g.
// `go tool cover`) are recorded and treated as successful no-ops.
//
// It deliberately does NOT implement DirRunner so that runInModuleDir exercises the
// chdir fallback, which is what places the profile in the module's directory.
type coverageWritingRunner struct {
	calls        [][]string
	failTestRuns bool // when true, any coverage pass returns an error without writing
}

func (r *coverageWritingRunner) RunCmd(name string, args ...string) error {
	r.calls = append(r.calls, append([]string{name}, args...))

	coverPath := coverProfilePath(args)
	if coverPath == "" {
		return nil // not a coverage pass (e.g. `go tool cover`)
	}
	if r.failTestRuns {
		return errSimulatedCoverageRun
	}

	// A unique package path per pass keeps merged output verifiable.
	line := fmt.Sprintf("mode: atomic\nexample.com/pkg/file_%d.go:1.1,2.1 1 1\n", len(r.calls))
	if err := os.WriteFile(coverPath, []byte(line), 0o600); err != nil {
		return fmt.Errorf("write simulated coverage profile: %w", err)
	}
	return nil
}

func (r *coverageWritingRunner) RunCmdOutput(name string, args ...string) (string, error) {
	return "", r.RunCmd(name, args...)
}

// coverProfilePath extracts the path from a -coverprofile=<path> argument, or "".
func coverProfilePath(args []string) string {
	for _, a := range args {
		if v, ok := strings.CutPrefix(a, "-coverprofile="); ok {
			return v
		}
	}
	return ""
}

// coverProfileCalls returns only the recorded calls that ran a coverage pass.
func coverProfileCalls(calls [][]string) [][]string {
	var out [][]string
	for _, c := range calls {
		if coverProfilePath(c) != "" {
			out = append(out, c)
		}
	}
	return out
}

// withCoverageHarness chdirs into a clean temp dir and installs a coverage-writing
// runner as the global runner, restoring both on cleanup. It returns the runner so
// callers can assert on the recorded passes.
func withCoverageHarness(t *testing.T, failTestRuns bool) *coverageWritingRunner {
	t.Helper()

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	tempDir := t.TempDir()
	require.NoError(t, os.Chdir(tempDir))
	t.Cleanup(func() { _ = os.Chdir(originalDir) }) //nolint:errcheck // test cleanup

	runner := &coverageWritingRunner{failTestRuns: failTestRuns}
	original := GetRunner()
	require.NoError(t, SetRunner(runner))
	t.Cleanup(func() { _ = SetRunner(original) }) //nolint:errcheck // test cleanup

	return runner
}

// TestCoverageCombinedModeProducesCanonicalFile is the regression guard for the CI
// failure where `magex test:coverrace` passed but never produced coverage.txt. In
// combined build-tag mode the single coverage pass runs with -tags set, but the
// merged profile must land in the canonical coverage.txt (not coverage_<tags>.txt),
// otherwise CI's Codecov upload finds nothing and the run fails despite green tests.
// race is true here to mirror the exact failing command, test:coverrace.
func TestCoverageCombinedModeProducesCanonicalFile(t *testing.T) {
	runner := withCoverageHarness(t, false)

	config := &Config{Test: TestConfig{
		AutoDiscoverBuildTags: true,
		CombineBuildTags:      true,
		CoverMode:             "atomic",
	}}
	modules := []ModuleInfo{{Path: ".", Relative: ".", Name: "test-module", IsRoot: true}}

	err := runCoverageTestsWithBuildTagDiscoveryTags(config, modules, true, nil, []string{"unit", "integration"})
	require.NoError(t, err)

	// The canonical file exists; the tag-named file must NOT.
	assert.FileExists(t, "coverage.txt")
	assert.NoFileExists(t, "coverage_unit_integration.txt")

	content, err := os.ReadFile("coverage.txt")
	require.NoError(t, err)
	assert.Contains(t, string(content), "mode: atomic")
	assert.Contains(t, string(content), "example.com/pkg/")

	// Exactly one coverage pass, carrying both tags combined.
	passes := coverProfileCalls(runner.calls)
	require.Len(t, passes, 1)
	assert.Equal(t, "unit,integration", tagsForCall(passes[0]))
}

// TestCoverageCombinedModeMergesAllModules verifies that when combined mode runs
// across multiple modules, every module's profile is merged into the single
// canonical coverage.txt and the per-module temporaries are cleaned up.
func TestCoverageCombinedModeMergesAllModules(t *testing.T) {
	runner := withCoverageHarness(t, false)
	require.NoError(t, os.Mkdir("sub", 0o750))

	config := &Config{Test: TestConfig{
		AutoDiscoverBuildTags: true,
		CombineBuildTags:      true,
		CoverMode:             "atomic",
	}}
	modules := []ModuleInfo{
		{Path: ".", Relative: ".", Name: "root", IsRoot: true},
		{Path: "sub", Relative: "sub", Name: "sub"},
	}

	err := runCoverageTestsWithBuildTagDiscoveryTags(config, modules, false, nil, []string{"unit", "integration"})
	require.NoError(t, err)

	assert.FileExists(t, "coverage.txt")
	// Per-module temporaries are removed after merge.
	assert.NoFileExists(t, "coverage_0_unit_integration.txt")
	assert.NoFileExists(t, "coverage_sub_unit_integration.txt")
	assert.NoFileExists(t, "coverage_unit_integration.txt")

	content, err := os.ReadFile("coverage.txt")
	require.NoError(t, err)
	// Merged output keeps a single mode header and both modules' coverage lines.
	assert.Equal(t, 1, strings.Count(string(content), "mode: atomic"))
	assert.Contains(t, string(content), "file_1.go")
	assert.Contains(t, string(content), "file_2.go")

	require.Len(t, coverProfileCalls(runner.calls), 2)
}

// TestCoveragePerTagModeProducesBaseAndTaggedFiles verifies the non-combined path is
// unaffected by the fix: it runs a base pass plus one pass per tag, writing
// coverage.txt for the base suite and coverage_<tag>.txt for each isolated tag.
func TestCoveragePerTagModeProducesBaseAndTaggedFiles(t *testing.T) {
	runner := withCoverageHarness(t, false)

	config := &Config{Test: TestConfig{
		AutoDiscoverBuildTags: true,
		CombineBuildTags:      false,
		CoverMode:             "atomic",
	}}
	modules := []ModuleInfo{{Path: ".", Relative: ".", Name: "test-module", IsRoot: true}}

	err := runCoverageTestsWithBuildTagDiscoveryTags(config, modules, false, nil, []string{"unit", "integration"})
	require.NoError(t, err)

	assert.FileExists(t, "coverage.txt")             // base (untagged) pass
	assert.FileExists(t, "coverage_unit.txt")        // isolated unit pass
	assert.FileExists(t, "coverage_integration.txt") // isolated integration pass

	// Base + one pass per tag = three coverage passes.
	passes := coverProfileCalls(runner.calls)
	require.Len(t, passes, 3)
	assert.Empty(t, tagsForCall(passes[0]))
	assert.Equal(t, "unit", tagsForCall(passes[1]))
	assert.Equal(t, "integration", tagsForCall(passes[2]))
}

// TestCoverageCombinedModeFailingPassPropagatesError confirms that when the combined
// coverage pass fails, the error surfaces and no stale coverage.txt is produced —
// the run must not look "green but empty".
func TestCoverageCombinedModeFailingPassPropagatesError(t *testing.T) {
	withCoverageHarness(t, true)

	config := &Config{Test: TestConfig{
		AutoDiscoverBuildTags: true,
		CombineBuildTags:      true,
		CoverMode:             "atomic",
	}}
	modules := []ModuleInfo{{Path: ".", Relative: ".", Name: "test-module", IsRoot: true}}

	err := runCoverageTestsWithBuildTagDiscoveryTags(config, modules, false, nil, []string{"unit", "integration"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "combined tags")
	assert.NoFileExists(t, "coverage.txt")
}

func TestCoverageOutputForTag(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "coverage.txt", coverageOutputForTag(""))
	assert.Equal(t, "coverage_unit.txt", coverageOutputForTag("unit"))
	assert.Equal(t, "coverage_unit_integration.txt", coverageOutputForTag("unit,integration"))
}

func TestFinalizeCoverageProfiles(t *testing.T) {
	t.Parallel()

	const body = "mode: atomic\nexample.com/pkg/a.go:1.1,2.1 1 1\n"

	t.Run("empty list is a no-op", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		out := filepath.Join(dir, "coverage.txt")
		finalizeCoverageProfiles(nil, out)
		assert.NoFileExists(t, out)
	})

	t.Run("single profile is renamed to output", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		src := filepath.Join(dir, "coverage_0.txt")
		out := filepath.Join(dir, "coverage.txt")
		require.NoError(t, os.WriteFile(src, []byte(body), 0o600))

		finalizeCoverageProfiles([]string{src}, out)

		assert.NoFileExists(t, src)
		assert.FileExists(t, out)
		got, err := os.ReadFile(out) //nolint:gosec // controlled test path
		require.NoError(t, err)
		assert.Equal(t, body, string(got))
	})

	t.Run("single profile already at output is left untouched", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		out := filepath.Join(dir, "coverage.txt")
		require.NoError(t, os.WriteFile(out, []byte(body), 0o600))

		finalizeCoverageProfiles([]string{out}, out)

		assert.FileExists(t, out)
		got, err := os.ReadFile(out) //nolint:gosec // controlled test path
		require.NoError(t, err)
		assert.Equal(t, body, string(got))
	})

	t.Run("multiple profiles merge and clean up inputs", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		a := filepath.Join(dir, "coverage_0.txt")
		b := filepath.Join(dir, "coverage_1.txt")
		out := filepath.Join(dir, "coverage.txt")
		require.NoError(t, os.WriteFile(a, []byte("mode: atomic\nexample.com/pkg/a.go:1.1,2.1 1 1\n"), 0o600))
		require.NoError(t, os.WriteFile(b, []byte("mode: atomic\nexample.com/pkg/b.go:1.1,2.1 1 1\n"), 0o600))

		finalizeCoverageProfiles([]string{a, b}, out)

		assert.NoFileExists(t, a)
		assert.NoFileExists(t, b)
		assert.FileExists(t, out)
		got, err := os.ReadFile(out) //nolint:gosec // controlled test path
		require.NoError(t, err)
		assert.Equal(t, 1, strings.Count(string(got), "mode: atomic"))
		assert.Contains(t, string(got), "a.go")
		assert.Contains(t, string(got), "b.go")
	})

	t.Run("output that is also an input is preserved", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		a := filepath.Join(dir, "coverage_0.txt")
		out := filepath.Join(dir, "coverage.txt")
		require.NoError(t, os.WriteFile(a, []byte("mode: atomic\nexample.com/pkg/a.go:1.1,2.1 1 1\n"), 0o600))
		require.NoError(t, os.WriteFile(out, []byte("mode: atomic\nexample.com/pkg/out.go:1.1,2.1 1 1\n"), 0o600))

		finalizeCoverageProfiles([]string{a, out}, out)

		assert.NoFileExists(t, a)    // the other input is cleaned up
		assert.FileExists(t, out)    // the destination is never deleted
		got, err := os.ReadFile(out) //nolint:gosec // controlled test path
		require.NoError(t, err)
		assert.Contains(t, string(got), "a.go")
		assert.Contains(t, string(got), "out.go")
	})
}
