package mage

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pkglog "github.com/mrz1836/mage-x/pkg/log"
)

// Static test errors to satisfy err113 linter
var (
	// errListBuildConstraints is a `go test -list` failure in the runner's
	// format: the command and exit status first, then the command's own output.
	errListBuildConstraints = errors.New("command failed [go test -list ^Fuzz example.com/fuzzrepro/tagged]: exit status 1\n" +
		"package example.com/fuzzrepro/tagged: build constraints exclude all Go files in /src/tagged\n" +
		"FAIL\texample.com/fuzzrepro/tagged [setup failed]")
	errListExitStatus = errors.New("exit status 1")
	// errListCompile is a `go test -list` failure caused by a compile error.
	errListCompile = errors.New("command failed [go test -list ^Fuzz example.com/fuzzrepro/broken]: exit status 1\n" +
		"# example.com/fuzzrepro/broken [example.com/fuzzrepro/broken.test]\n" +
		"broken/broken_test.go:6:40: undefined: undefinedCall\n" +
		"FAIL\texample.com/fuzzrepro/broken [build failed]")
)

func TestBuildTagPasses(t *testing.T) {
	tests := []struct {
		name       string
		testTags   string
		discovered []string
		combine    bool
		want       []string
	}{
		{name: "NoneNoTestTags", want: []string{""}},
		{name: "NonePerTagMode", combine: false, want: []string{""}},
		{name: "NoneWithTestTags", testTags: "integration", combine: true, want: []string{"integration"}},
		{name: "NoneKeepsTestTagsVerbatim", testTags: "integration e2e", want: []string{"integration e2e"}},
		{name: "Combined", discovered: []string{"unit", "e2e"}, combine: true, want: []string{"unit,e2e"}},
		{
			name: "CombinedWithTestTags", testTags: "integration", discovered: []string{"unit", "e2e"}, combine: true,
			want: []string{"unit,e2e,integration"},
		},
		{
			name: "CombinedDeduplicatesTestTags", testTags: "e2e integration", discovered: []string{"unit", "e2e"}, combine: true,
			want: []string{"unit,e2e,integration"},
		},
		{name: "PerTag", discovered: []string{"unit", "e2e"}, want: []string{"", "unit", "e2e"}},
		{
			name: "PerTagWithTestTags", testTags: "integration", discovered: []string{"unit", "e2e"},
			want: []string{"integration", "unit,integration", "e2e,integration"},
		},
		{
			// The "unit" pass would enable the same tags as the base pass.
			name: "PerTagDeduplicatesPasses", testTags: "unit", discovered: []string{"unit", "e2e"},
			want: []string{"unit", "e2e,unit"},
		},
		{
			// "b,a" is the same tag set as the base pass "a,b".
			name: "PerTagDeduplicatesReorderedSets", testTags: "a,b", discovered: []string{"b", "c"},
			want: []string{"a,b", "c,a,b"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, buildTagPasses(tt.testTags, tt.discovered, tt.combine))
		})
	}
}

func TestTestBuildTagPassesIgnoresDiscoveredTagsWhenDisabled(t *testing.T) {
	config := &Config{Test: TestConfig{AutoDiscoverBuildTags: false, CombineBuildTags: true, Tags: "integration"}}
	assert.Equal(t, []string{"integration"}, testBuildTagPasses(config, []string{"unit"}))

	config.Test.AutoDiscoverBuildTags = true
	assert.Equal(t, []string{"unit,integration"}, testBuildTagPasses(config, []string{"unit"}))
}

func TestMergeBuildTags(t *testing.T) {
	assert.Empty(t, mergeBuildTags())
	assert.Empty(t, mergeBuildTags("", " , "))
	assert.Equal(t, "unit", mergeBuildTags("unit", ""))
	assert.Equal(t, "unit,e2e,integration", mergeBuildTags("unit,e2e", "integration e2e", "unit"))
}

func TestRunTestsWithBuildTagDiscoveryTagsMergesTestTags(t *testing.T) {
	modules := []ModuleInfo{{Path: ".", Relative: ".", Name: "test-module", IsRoot: true}}
	discoveredTags := []string{"unit", "e2e"}

	tests := []struct {
		name    string
		combine bool
		want    []string
	}{
		{name: "Combined", combine: true, want: []string{"unit,e2e,integration"}},
		{name: "PerTag", combine: false, want: []string{"integration", "unit,integration", "e2e,integration"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := useRecordingRunner(t)
			config := &Config{Test: TestConfig{
				AutoDiscoverBuildTags: true,
				CombineBuildTags:      tt.combine,
				Tags:                  "integration",
			}}

			require.NoError(t, runTestsWithBuildTagDiscoveryTags(config, modules, []string{}, "unit", discoveredTags))

			require.Len(t, recorder.calls, len(tt.want))
			for i, want := range tt.want {
				assert.Equal(t, want, tagsForCall(recorder.calls[i]))
			}
		})
	}
}

func TestRunCoverageTestsWithBuildTagDiscoveryTagsMergesTestTags(t *testing.T) {
	recorder := useRecordingRunner(t)
	config := &Config{Test: TestConfig{
		AutoDiscoverBuildTags: true,
		CombineBuildTags:      true,
		CoverMode:             "atomic",
		Tags:                  "integration",
	}}
	modules := []ModuleInfo{{Path: ".", Relative: ".", Name: "test-module", IsRoot: true}}

	require.NoError(t, runCoverageTestsWithBuildTagDiscoveryTags(config, modules, false, []string{}, []string{"unit", "e2e"}))

	require.Len(t, recorder.calls, 1)
	assert.Equal(t, "unit,e2e,integration", tagsForCall(recorder.calls[0]))
}

func TestFuzzArgsCarryBuildTags(t *testing.T) {
	const pkg = "example.com/fuzzrepro/tagged"

	t.Run("List", func(t *testing.T) {
		assert.Equal(t, []string{"test", "-list", "^Fuzz", pkg}, fuzzListArgs(pkg, ""))
		assert.Equal(t, []string{"test", "-tags", "mage", "-list", "^Fuzz", pkg}, fuzzListArgs(pkg, "mage"))
	})

	t.Run("Warmup", func(t *testing.T) {
		assert.Equal(t, []string{"test", "-run=^$", "-fuzz=^FuzzY$", "-fuzztime=1x", pkg}, fuzzWarmupArgs(pkg, "FuzzY", ""))
		assert.Equal(t, []string{"test", "-tags", "mage,unit", "-run=^$", "-fuzz=^FuzzY$", "-fuzztime=1x", pkg},
			fuzzWarmupArgs(pkg, "FuzzY", "mage,unit"))
	})

	t.Run("Run", func(t *testing.T) {
		assert.Equal(t,
			[]string{"test", "-run=^$", "-fuzz=^FuzzY$", "-fuzztime", "3s", "-timeout", "1m30s", pkg},
			fuzzRunArgs(pkg, "FuzzY", "", "3s", 90*time.Second, false))
		assert.Equal(t,
			[]string{"test", "-tags", "mage", "-run=^$", "-fuzz=^FuzzY$", "-fuzztime", "3s", "-timeout", "1m30s", "-v", pkg},
			fuzzRunArgs(pkg, "FuzzY", "mage", "3s", 90*time.Second, true))
	})

	t.Run("WarmupMatchesRunBuildFlags", func(t *testing.T) {
		// The warm-up only helps when it builds with the run's -tags.
		run := fuzzRunArgs(pkg, "FuzzY", "mage", "3s", time.Minute, false)
		warmup := fuzzWarmupArgs(pkg, "FuzzY", "mage")
		assert.Equal(t, tagsForCall(run), tagsForCall(warmup))
	})

	t.Run("ResolvePkgDir", func(t *testing.T) {
		recorder := useRecordingRunner(t)
		resolvePkgDir(pkg, "mage")
		require.Len(t, recorder.calls, 1)
		assert.Equal(t, []string{"go", "list", "-tags", "mage", "-f", "{{.Dir}}", pkg}, recorder.calls[0])
	})
}

func TestListFuzzTargets(t *testing.T) {
	const pkg = "example.com/fuzzrepro/tagged"

	t.Run("PerTagModeRunsEachTargetOnceUnderFirstTagSet", func(t *testing.T) {
		runner := useFuzzListRunner(t, map[string]string{
			"":     "FuzzX\nok  \texample.com/fuzzrepro/tagged\t0.1s",
			"mage": "FuzzX\nFuzzY\nok  \texample.com/fuzzrepro/tagged\t0.1s",
		}, nil)

		groups := listFuzzTargets(pkg, []string{"", "mage"})

		assert.Equal(t, []fuzzTargetGroup{
			{Tags: "", Tests: []string{"FuzzX"}},
			{Tags: "mage", Tests: []string{"FuzzY"}},
		}, groups)
		require.Len(t, runner.calls, 2)
		assert.Equal(t, []string{"go", "test", "-list", "^Fuzz", pkg}, runner.calls[0])
		assert.Equal(t, []string{"go", "test", "-tags", "mage", "-list", "^Fuzz", pkg}, runner.calls[1])
	})

	t.Run("TaggedPackageListsUnderItsTagWithoutWarning", func(t *testing.T) {
		useFuzzListRunner(t, map[string]string{"mage": "FuzzY"}, map[string]error{"": errListBuildConstraints})

		var groups []fuzzTargetGroup
		logged := captureCLILog(t, func() { groups = listFuzzTargets(pkg, []string{"", "mage"}) })

		assert.Equal(t, []fuzzTargetGroup{{Tags: "mage", Tests: []string{"FuzzY"}}}, groups)
		assert.NotContains(t, logged, "Skipping fuzz tests")
	})

	t.Run("WarnsWhenListFailsUnderEveryTagSet", func(t *testing.T) {
		useFuzzListRunner(t, nil, map[string]error{"": errListBuildConstraints})

		var groups []fuzzTargetGroup
		logged := captureCLILog(t, func() { groups = listFuzzTargets(pkg, []string{""}) })

		assert.Empty(t, groups)
		assert.Contains(t, logged, "Skipping fuzz tests in "+pkg)
		assert.Contains(t, logged, "build constraints exclude all Go files in /src/tagged")
		assert.NotContains(t, logged, "setup failed", "only the first line of the failure is shown")
	})

	t.Run("NoWarningWhenListFindsNoTargets", func(t *testing.T) {
		useFuzzListRunner(t, map[string]string{"": "ok  \texample.com/fuzzrepro/tagged\t0.1s"}, nil)

		var groups []fuzzTargetGroup
		logged := captureCLILog(t, func() { groups = listFuzzTargets(pkg, []string{""}) })

		assert.Empty(t, groups)
		assert.NotContains(t, logged, "Skipping fuzz tests")
	})
}

func TestFuzzListFailureLine(t *testing.T) {
	assert.Equal(t,
		"package example.com/fuzzrepro/tagged: build constraints exclude all Go files in /src/tagged",
		fuzzListFailureLine("", errListBuildConstraints))
	assert.Equal(t,
		"broken/broken_test.go:6:40: undefined: undefinedCall",
		fuzzListFailureLine("", errListCompile))
	assert.Equal(t, "listed output", fuzzListFailureLine("\nlisted output\nmore", errListBuildConstraints))
	assert.Equal(t, "exit status 1", fuzzListFailureLine("", errListExitStatus))
}

func TestBenchmarkArgsCarryBuildTags(t *testing.T) {
	assert.Equal(t, []string{"test", "-bench=.", "-benchmem", "-run=^$", "-benchtime", "1s", "./..."},
		benchmarkArgs(false, "", "1s", ""))
	assert.Equal(t, []string{"test", "-bench=.", "-benchmem", "-run=^$", "-v", "-tags", "mage", "-benchtime", "1s", "-count", "2", "./..."},
		benchmarkArgs(true, "mage", "1s", "2"))
}

func TestTestBenchRunsOnePassPerBuildTagSet(t *testing.T) {
	tests := []struct {
		name    string
		combine bool
		want    []string
	}{
		{name: "Combined", combine: true, want: []string{"mage,integration"}},
		{name: "PerTag", combine: false, want: []string{"integration", "mage,integration"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := setupBuildTagModule(t, &Config{Test: TestConfig{
				AutoDiscoverBuildTags: true,
				CombineBuildTags:      tt.combine,
				Tags:                  "integration",
			}})

			require.NoError(t, Test{}.BenchShort("time=10ms"))

			require.Len(t, recorder.calls, len(tt.want))
			for i, want := range tt.want {
				assert.Equal(t, want, tagsForCall(recorder.calls[i]))
				assert.Contains(t, recorder.calls[i], "-bench=.")
			}
		})
	}
}

// useRecordingRunner installs a combineRecordingRunner for the rest of the test.
func useRecordingRunner(t *testing.T) *combineRecordingRunner {
	t.Helper()
	recorder := &combineRecordingRunner{}
	original := GetRunner()
	require.NoError(t, SetRunner(recorder))
	t.Cleanup(func() { require.NoError(t, SetRunner(original)) })
	return recorder
}

// fuzzListRunner answers `go test -list` with the listing or error configured
// for the call's -tags value, and records every call.
type fuzzListRunner struct {
	combineRecordingRunner

	listings map[string]string
	failures map[string]error
}

func (r *fuzzListRunner) RunCmdOutput(name string, args ...string) (string, error) {
	call := append([]string{name}, args...)
	r.calls = append(r.calls, call)
	tags := tagsForCall(call)
	if err, ok := r.failures[tags]; ok {
		return "", err
	}
	return r.listings[tags], nil
}

// useFuzzListRunner installs a fuzzListRunner for the rest of the test.
func useFuzzListRunner(t *testing.T, listings map[string]string, failures map[string]error) *fuzzListRunner {
	t.Helper()
	runner := &fuzzListRunner{listings: listings, failures: failures}
	original := GetRunner()
	require.NoError(t, SetRunner(runner))
	t.Cleanup(func() { require.NoError(t, SetRunner(original)) })
	return runner
}

// captureCLILog redirects the shared CLI logger into a buffer while fn runs and
// returns what it wrote.
func captureCLILog(t *testing.T, fn func()) string {
	t.Helper()
	var buf bytes.Buffer
	logger := pkglog.Default()
	logger.SetOutput(&buf)
	t.Cleanup(func() { logger.SetOutput(os.Stdout) })
	fn()
	return buf.String()
}

// setupBuildTagModule changes into a new module whose only build tag is
// "mage", and installs cfg and a recording runner for the rest of the test.
func setupBuildTagModule(t *testing.T, cfg *Config) *combineRecordingRunner {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/fuzzrepro\n\ngo 1.24\n"), 0o600))
	createTestFile(t, dir, "plain_test.go", "")
	createTestFile(t, dir, "tagged_test.go", "//go:build mage")
	t.Chdir(dir)

	TestSetConfig(cfg)
	t.Cleanup(TestResetConfig)

	recorder := useRecordingRunner(t)
	require.Empty(t, recorder.calls)
	return recorder
}

// TestSetupBuildTagModuleDiscoversMage guards the fixture the bench tests rely on.
func TestSetupBuildTagModuleDiscoversMage(t *testing.T) {
	setupBuildTagModule(t, &Config{Test: TestConfig{AutoDiscoverBuildTags: true}})
	config, err := GetConfig()
	require.NoError(t, err)
	assert.Equal(t, []string{"mage"}, discoverTestBuildTags(config))
}
