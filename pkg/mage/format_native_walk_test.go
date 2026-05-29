package mage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// formatWalkTestEnv sets up an isolated temp working directory containing the given files
// (keyed by relative path), applies the given MAGE_X_* environment overrides, and restores
// the working directory and environment on cleanup. Unlike setupYAMLTestDir, it does not
// force a particular validation mode, so callers can exercise validation-enabled paths.
func formatWalkTestEnv(t *testing.T, files, envOverrides map[string]string) {
	t.Helper()

	// Snapshot and reset the format-related env vars to a clean baseline, restoring on cleanup.
	for _, key := range []string{"MAGE_X_YAML_VALIDATION", "MAGE_X_FORMAT_EXCLUDE_PATHS"} {

		orig, had := os.LookupEnv(key)
		t.Cleanup(func() {
			if had {
				_ = os.Setenv(key, orig) //nolint:errcheck // test cleanup
			} else {
				_ = os.Unsetenv(key) //nolint:errcheck // test cleanup
			}
		})
		_ = os.Unsetenv(key) //nolint:errcheck // clean baseline
	}
	for k, v := range envOverrides {
		require.NoError(t, os.Setenv(k, v))
	}

	origDir, err := os.Getwd()
	require.NoError(t, err)
	tmpDir := t.TempDir()
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() { _ = os.Chdir(origDir) }) //nolint:errcheck // test cleanup

	for path, content := range files {
		if dir := filepath.Dir(path); dir != "." {
			require.NoError(t, os.MkdirAll(dir, 0o750))
		}
		require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	}
}

// installMockRunner swaps in a fresh MockCommandRunner and restores the original on cleanup.
func installMockRunner(t *testing.T) *MockCommandRunner {
	t.Helper()
	originalRunner := GetRunner()
	t.Cleanup(func() { _ = SetRunner(originalRunner) }) //nolint:errcheck // test cleanup
	mockRunner := &MockCommandRunner{}
	require.NoError(t, SetRunner(mockRunner))
	return mockRunner
}

// TestChunkByArgBytes verifies the ARG_MAX-aware chunking logic.
func TestChunkByArgBytes(t *testing.T) {
	t.Run("empty input returns nil", func(t *testing.T) {
		assert.Nil(t, chunkByArgBytes(nil, 100))
		assert.Nil(t, chunkByArgBytes([]string{}, 100))
	})

	t.Run("single file is one chunk", func(t *testing.T) {
		got := chunkByArgBytes([]string{"a.yml"}, 100)
		assert.Equal(t, [][]string{{"a.yml"}}, got)
	})

	t.Run("files within budget stay in one chunk", func(t *testing.T) {
		got := chunkByArgBytes([]string{"a.yml", "b.yml", "c.yml"}, 100)
		assert.Equal(t, [][]string{{"a.yml", "b.yml", "c.yml"}}, got)
	})

	t.Run("splits at the budget boundary preserving order", func(t *testing.T) {
		// Each path costs len+1. With 5-char paths that is 6 bytes each.
		// Budget 12 fits exactly two per chunk.
		files := []string{"a.yml", "b.yml", "c.yml", "d.yml", "e.yml"}
		got := chunkByArgBytes(files, 12)
		assert.Equal(t, [][]string{{"a.yml", "b.yml"}, {"c.yml", "d.yml"}, {"e.yml"}}, got)
	})

	t.Run("over-budget single file becomes its own chunk", func(t *testing.T) {
		big := strings.Repeat("x", 50) + ".yml"
		got := chunkByArgBytes([]string{"a.yml", big, "b.yml"}, 12)
		assert.Equal(t, [][]string{{"a.yml"}, {big}, {"b.yml"}}, got)
	})
}

// TestFindFilesByExt verifies native-walk discovery, extension matching, and dir exclusion.
func TestFindFilesByExt(t *testing.T) {
	t.Run("matches extensions case-insensitively and skips others", func(t *testing.T) {
		formatWalkTestEnv(t, map[string]string{
			"a.yml":      "x: 1\n",
			"b.yaml":     "y: 2\n",
			"c.YML":      "z: 3\n",
			"d.json":     "{}",
			"notes.txt":  "ignore me",
			"sub/e.yaml": "w: 4\n",
		}, nil)

		got, err := findFilesByExt([]string{".yml", ".yaml"}, newExcludeDirPredicate())
		require.NoError(t, err)
		// Deterministic lexical order from filepath.Walk.
		assert.Equal(t, []string{"a.yml", "b.yaml", "c.YML", "sub/e.yaml"}, got)
	})

	t.Run("skips excluded directories at any depth (the .yml precedence bug)", func(t *testing.T) {
		formatWalkTestEnv(t, map[string]string{
			"keep.yml":                       "a: 1\n",
			"vendor/dep.yml":                 "v: 1\n",
			"deep/nested/ci-tester/bad.yml":  "broken",
			"deep/nested/ci-tester/bad.yaml": "broken",
			"deep/nested/keep2.yaml":         "k: 2\n",
			"node_modules/pkg/thing.yaml":    "n: 1\n",
		}, map[string]string{"MAGE_X_FORMAT_EXCLUDE_PATHS": "vendor,node_modules,ci-tester"})

		got, err := findFilesByExt([]string{".yml", ".yaml"}, newExcludeDirPredicate())
		require.NoError(t, err)
		assert.Equal(t, []string{"deep/nested/keep2.yaml", "keep.yml"}, got,
			"both .yml and .yaml under excluded dirs must be pruned, at any depth")
	})

	t.Run("nil predicate disables exclusion", func(t *testing.T) {
		formatWalkTestEnv(t, map[string]string{
			"keep.yml":       "a: 1\n",
			"vendor/dep.yml": "v: 1\n",
		}, nil)

		got, err := findFilesByExt([]string{".yml"}, nil)
		require.NoError(t, err)
		assert.Equal(t, []string{"keep.yml", "vendor/dep.yml"}, got)
	})

	t.Run("empty tree yields no files", func(t *testing.T) {
		formatWalkTestEnv(t, nil, nil)
		got, err := findFilesByExt([]string{".yml", ".yaml"}, newExcludeDirPredicate())
		require.NoError(t, err)
		assert.Empty(t, got)
	})

	t.Run("multiple distinct extensions", func(t *testing.T) {
		formatWalkTestEnv(t, map[string]string{
			"a.json": "{}",
			"b.yml":  "x: 1\n",
			"c.go":   "package x",
		}, nil)
		got, err := findFilesByExt([]string{".json"}, newExcludeDirPredicate())
		require.NoError(t, err)
		assert.Equal(t, []string{"a.json"}, got)
	})
}

// TestFormatYAMLFilesBatched verifies the explicit-list yamlfmt invocation and chunking.
func TestFormatYAMLFilesBatched(t *testing.T) {
	t.Run("no config: explicit list in one call", func(t *testing.T) {
		formatWalkTestEnv(t, nil, nil)
		mockRunner := installMockRunner(t)
		mockRunner.On("RunCmd", "yamlfmt", "a.yml", "b.yaml").Return(nil)

		require.NoError(t, formatYAMLFilesBatched([]string{"a.yml", "b.yaml"}, ".github/.yamlfmt"))
		mockRunner.AssertExpectations(t)
	})

	t.Run("config present: prepends -conf", func(t *testing.T) {
		formatWalkTestEnv(t, map[string]string{".github/.yamlfmt": "formatter:\n  type: basic\n"}, nil)
		mockRunner := installMockRunner(t)
		mockRunner.On("RunCmd", "yamlfmt", "-conf", ".github/.yamlfmt", "a.yml").Return(nil)

		require.NoError(t, formatYAMLFilesBatched([]string{"a.yml"}, ".github/.yamlfmt"))
		mockRunner.AssertExpectations(t)
	})

	t.Run("splits into multiple invocations beyond ARG_MAX budget", func(t *testing.T) {
		formatWalkTestEnv(t, nil, nil)
		mockRunner := installMockRunner(t)

		// Two ~40KB names fit in one chunk (~80KB < 100KB); the third forces a new chunk.
		name := func(c string) string { return strings.Repeat(c, 40_000) + ".yml" }
		f1, f2, f3 := name("a"), name("b"), name("c")
		mockRunner.On("RunCmd", "yamlfmt", f1, f2).Return(nil)
		mockRunner.On("RunCmd", "yamlfmt", f3).Return(nil)

		require.NoError(t, formatYAMLFilesBatched([]string{f1, f2, f3}, "nonexistent-config"))
		mockRunner.AssertExpectations(t)
	})

	t.Run("propagates the chunk error", func(t *testing.T) {
		formatWalkTestEnv(t, nil, nil)
		mockRunner := installMockRunner(t)
		mockRunner.On("RunCmd", "yamlfmt", "bad.yml").Return(ErrYamlfmtExecutionFailed)

		err := formatYAMLFilesBatched([]string{"bad.yml"}, "nonexistent-config")
		require.ErrorIs(t, err, ErrYamlfmtExecutionFailed)
	})
}

// TestFormatYAMLBatchFallback verifies the batch-then-per-file robustness behavior, the core
// guarantee that a single unparseable file cannot block formatting of the others.
func TestFormatYAMLBatchFallback(t *testing.T) {
	t.Run("batch success does not fall back to per-file", func(t *testing.T) {
		formatWalkTestEnv(t, map[string]string{
			"a.yml": "a: 1\n",
			"b.yml": "b: 2\n",
		}, nil)
		mockRunner := installMockRunner(t)
		// Only the batch call is expected; no per-file calls.
		mockRunner.On("RunCmd", "yamlfmt", "a.yml", "b.yml").Return(nil)

		require.NoError(t, Format{}.YAML())
		mockRunner.AssertExpectations(t)
		mockRunner.AssertNotCalled(t, "RunCmd", "yamlfmt", "a.yml")
	})

	t.Run("batch failure falls back per-file and isolates the bad file", func(t *testing.T) {
		formatWalkTestEnv(t, map[string]string{
			"a.yml": "a: 1\n",
			"b.yml": "b: 2\n",
			"c.yml": "c: 3\n",
		}, nil)
		mockRunner := installMockRunner(t)
		// Batch fails (c.yml is unparseable for yamlfmt)...
		mockRunner.On("RunCmd", "yamlfmt", "a.yml", "b.yml", "c.yml").Return(ErrYamlfmtExecutionFailed)
		// ...then each file is retried individually; only c.yml still fails.
		mockRunner.On("RunCmd", "yamlfmt", "a.yml").Return(nil)
		mockRunner.On("RunCmd", "yamlfmt", "b.yml").Return(nil)
		mockRunner.On("RunCmd", "yamlfmt", "c.yml").Return(ErrYamlfmtExecutionFailed)

		err := Format{}.YAML()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "yamlfmt formatting failed")
		// All four calls must have happened: the good files were still attempted.
		mockRunner.AssertExpectations(t)
	})

	t.Run("batch failure with all-good fallback succeeds", func(t *testing.T) {
		formatWalkTestEnv(t, map[string]string{
			"a.yml": "a: 1\n",
			"b.yml": "b: 2\n",
		}, nil)
		mockRunner := installMockRunner(t)
		mockRunner.On("RunCmd", "yamlfmt", "a.yml", "b.yml").Return(ErrYamlfmtExecutionFailed)
		mockRunner.On("RunCmd", "yamlfmt", "a.yml").Return(nil)
		mockRunner.On("RunCmd", "yamlfmt", "b.yml").Return(nil)

		require.NoError(t, Format{}.YAML())
		mockRunner.AssertExpectations(t)
	})
}

// TestFormatYAMLValidationSkipsLongLines verifies that line-length validation removes
// problematic files from the set handed to yamlfmt while still formatting the safe ones.
func TestFormatYAMLValidationSkipsLongLines(t *testing.T) {
	longLine := strings.Repeat("a", MaxYAMLLineLength+1)
	formatWalkTestEnv(t, map[string]string{
		"good.yml": "a: 1\n",
		"bad.yml":  "key: " + longLine + "\n",
	}, map[string]string{"MAGE_X_YAML_VALIDATION": "true"})

	mockRunner := installMockRunner(t)
	// Only the safe file reaches yamlfmt; bad.yml is skipped (no expectation set for it).
	mockRunner.On("RunCmd", "yamlfmt", "good.yml").Return(nil)

	require.NoError(t, Format{}.YAML())
	mockRunner.AssertExpectations(t)
	mockRunner.AssertNotCalled(t, "RunCmd", "yamlfmt", "bad.yml")
}

// TestFormatJSONNativeWalk verifies JSON discovery + native formatting honor exclusions and
// never invoke an external command.
func TestFormatJSONNativeWalk(t *testing.T) {
	t.Run("formats valid files and skips excluded directories", func(t *testing.T) {
		formatWalkTestEnv(t, map[string]string{
			"a.json":           `{"name":"test","value":123}`,
			"sub/b.JSON":       `{"x":1}`,
			"vendor/skip.json": `this is not valid json`,
		}, map[string]string{"MAGE_X_FORMAT_EXCLUDE_PATHS": "vendor"})

		// JSON formatting is pure Go; the runner must never be called.
		mockRunner := installMockRunner(t)

		require.NoError(t, Format{}.JSON())

		// Valid files reformatted with 4-space indentation.
		assert.Contains(t, readTestFile(t, "a.json"), "\n    ")
		assert.Contains(t, readTestFile(t, "sub/b.JSON"), "\n    ")
		// Invalid file under an excluded directory is left untouched.
		assert.Equal(t, `this is not valid json`, readTestFile(t, "vendor/skip.json"))

		mockRunner.AssertNotCalled(t, "RunCmd")
		mockRunner.AssertNotCalled(t, "RunCmdOutput")
	})

	t.Run("no JSON files is a no-op", func(t *testing.T) {
		formatWalkTestEnv(t, map[string]string{"notes.txt": "hi"}, nil)
		installMockRunner(t)
		require.NoError(t, Format{}.JSON())
	})
}

func readTestFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path) //nolint:gosec // test-controlled path
	require.NoError(t, err)
	return string(data)
}
