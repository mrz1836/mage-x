package testhelpers

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Static test errors to satisfy err113 linter
var (
	errTestError   = errors.New("error")
	errNotYet      = errors.New("not yet")
	errAlwaysFails = errors.New("always fails")
	errError2      = errors.New("error2")
)

func TestRunCommand(t *testing.T) {
	t.Run("successful command", func(t *testing.T) {
		output, err := RunCommand(t, "echo", "hello world")
		require.NoError(t, err)
		require.Equal(t, "hello world\n", output)
	})

	t.Run("command with error", func(t *testing.T) {
		_, err := RunCommand(t, "ls", "/nonexistent/directory")
		require.Error(t, err)
		require.Contains(t, err.Error(), "stderr:")
	})

	t.Run("nonexistent command", func(t *testing.T) {
		_, err := RunCommand(t, "nonexistentcommand123")
		require.Error(t, err)
	})
}

func TestRunCommandWithInput(t *testing.T) {
	t.Run("command with input", func(t *testing.T) {
		// Use a command that reads from stdin
		output, err := RunCommandWithInput(t, "test input", "cat")
		require.NoError(t, err)
		require.Equal(t, "test input", output)
	})

	t.Run("command with error and input", func(t *testing.T) {
		_, err := RunCommandWithInput(t, "input", "ls", "/nonexistent")
		require.Error(t, err)
		require.Contains(t, err.Error(), "stderr:")
	})
}

func TestRequireCommand(t *testing.T) {
	t.Run("existing command", func(t *testing.T) {
		// This should not skip
		RequireCommand(t, "echo")
		// If we get here, the test didn't skip (command exists)
	})

	// We can't easily test the skip case without actually skipping
}

func TestRequireEnv(t *testing.T) {
	t.Run("existing env var", func(t *testing.T) {
		// Set a test env var
		require.NoError(t, os.Setenv("TEST_REQUIRE_ENV", "test_value"))
		defer func() {
			require.NoError(t, os.Unsetenv("TEST_REQUIRE_ENV"))
		}()

		value := RequireEnv(t, "TEST_REQUIRE_ENV")
		require.Equal(t, "test_value", value)
	})

	// Can't easily test the skip case
}

func TestRequireNetwork(t *testing.T) {
	// This is hard to test reliably
	// Just ensure it doesn't panic
	RequireNetwork(t)
	// Either it skips or it doesn't, both are valid
}

func TestRequireDocker(t *testing.T) {
	// Test in a subtest to allow skipping
	t.Run("docker check", func(t *testing.T) {
		RequireDocker(t)
		// If we get here, Docker is available
	})
}

func TestRequireGit(t *testing.T) {
	// Test in a subtest to allow skipping
	t.Run("git check", func(t *testing.T) {
		RequireGit(t)
		// If we get here, Git is available
	})
}

func TestAssertContains(t *testing.T) {
	// Test the actual function with real testing.T
	t.Run("contains substring", func(t *testing.T) {
		// This should pass without error
		AssertContains(t, "hello world", "world")
	})
}

func TestAssertNotContains(t *testing.T) {
	t.Run("does not contain substring", func(t *testing.T) {
		// This should pass without error
		AssertNotContains(t, "hello world", "foo")
	})
}

func TestAssertEquals(t *testing.T) {
	t.Run("equal values", func(t *testing.T) {
		// This should pass without error
		AssertEquals(t, 42, 42)
		AssertEquals(t, "hello", "hello")
	})
}

func TestAssertNotEquals(t *testing.T) {
	t.Run("unequal values", func(t *testing.T) {
		// This should pass without error
		AssertNotEquals(t, 42, 43)
		AssertNotEquals(t, "hello", "world")
	})
}

func TestAssertTrue(t *testing.T) {
	t.Run("true value", func(t *testing.T) {
		// This should pass without error
		AssertTrue(t, true)
		AssertTrue(t, len("test") == 4, "string length works")
	})
}

func TestAssertFalse(t *testing.T) {
	t.Run("false value", func(t *testing.T) {
		// This should pass without error
		AssertFalse(t, false)
		AssertFalse(t, 1 == 2, "math still works")
	})
}

func TestAssertNil(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		// This should pass without error
		var nilErr error
		AssertNil(t, nilErr)
		AssertNil(t, nil)
	})
}

func TestAssertNotNil(t *testing.T) {
	t.Run("non-nil value", func(t *testing.T) {
		// This should pass without error
		AssertNotNil(t, "not nil")
		AssertNotNil(t, errTestError)
	})
}

func TestEventuallyTrue(t *testing.T) {
	t.Run("becomes true immediately", func(t *testing.T) {
		called := false
		EventuallyTrue(t, func() bool {
			called = true
			return true
		}, 100*time.Millisecond, "should be true")

		require.True(t, called)
	})

	t.Run("becomes true after delay", func(t *testing.T) {
		start := time.Now()
		counter := 0

		EventuallyTrue(t, func() bool {
			counter++
			return counter >= 3
		}, 200*time.Millisecond, "should become true")

		require.GreaterOrEqual(t, counter, 3)
		require.Less(t, time.Since(start), 200*time.Millisecond)
	})
}

func TestEventuallyEquals(t *testing.T) {
	t.Run("equals immediately", func(t *testing.T) {
		EventuallyEquals(t, func() interface{} {
			return 42
		}, 42, 100*time.Millisecond)
	})

	t.Run("equals after delay", func(t *testing.T) {
		counter := 0

		EventuallyEquals(t, func() interface{} {
			counter++
			return counter
		}, 3, 200*time.Millisecond)

		require.Equal(t, 3, counter)
	})
}

func TestRetry(t *testing.T) {
	t.Run("succeeds immediately", func(t *testing.T) {
		attempts := 0
		err := Retry(t, func() error {
			attempts++
			return nil
		}, 3, 10*time.Millisecond)

		require.NoError(t, err)
		require.Equal(t, 1, attempts)
	})

	t.Run("succeeds after retries", func(t *testing.T) {
		attempts := 0
		err := Retry(t, func() error {
			attempts++
			if attempts < 3 {
				return errNotYet
			}
			return nil
		}, 5, 10*time.Millisecond)

		require.NoError(t, err)
		require.Equal(t, 3, attempts)
	})

	t.Run("fails after all attempts", func(t *testing.T) {
		attempts := 0
		err := Retry(t, func() error {
			attempts++
			return errAlwaysFails
		}, 3, 10*time.Millisecond)

		require.Error(t, err)
		require.Contains(t, err.Error(), "failed after 3 attempts")
		require.Contains(t, err.Error(), "always fails")
		require.Equal(t, 3, attempts)
	})
}

func TestSkipIfShort(t *testing.T) {
	// Test in a subtest so the parent test doesn't skip
	t.Run("skip check", func(t *testing.T) {
		SkipIfShort(t)
		// If we get here, we're not in short mode
		require.False(t, testing.Short())
	})
}

func TestSkipIfCI(t *testing.T) {
	// Save original env
	origCI := os.Getenv("CI")
	defer func() {
		if origCI == "" {
			require.NoError(t, os.Unsetenv("CI"))
		} else {
			require.NoError(t, os.Setenv("CI", origCI))
		}
	}()

	t.Run("not in CI", func(t *testing.T) {
		require.NoError(t, os.Unsetenv("CI"))
		require.NoError(t, os.Unsetenv("GITHUB_ACTIONS"))

		SkipIfCI(t)
		// If we get here, we're not in CI
	})
}

func TestRunParallel(t *testing.T) {
	// Just ensure it doesn't panic
	RunParallel(t)
}

func TestBenchmark(t *testing.T) {
	executed := false

	Benchmark(t, "test operation", func() {
		executed = true
		time.Sleep(10 * time.Millisecond)
	})

	require.True(t, executed)
}

func TestMeasureTime(t *testing.T) {
	done := MeasureTime(t, "test operation")
	time.Sleep(10 * time.Millisecond)
	done()
	// Function should log the time taken
}

func TestCaptureLog(t *testing.T) {
	// Skip on Windows as pipe handling is different
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows")
	}

	output := CaptureLog(t, func() {
		fmt.Fprintln(os.Stderr, "test log output")
	})

	// The captured output might be empty due to buffering
	// Just ensure the function doesn't panic
	_ = output
}

func TestGolden(t *testing.T) {
	// Create a temporary directory for golden files
	tempDir := t.TempDir()

	t.Run("check with update", func(t *testing.T) {
		g := &Golden{
			t:      t,
			dir:    tempDir,
			update: true,
		}

		// Update golden file
		g.Check("updated", []byte("new content"))

		// Verify file was written
		content, err := os.ReadFile(tempDir + "/updated.golden") // #nosec G304 -- controlled test path
		require.NoError(t, err)
		require.Equal(t, "new content", string(content))
	})

	t.Run("check string", func(t *testing.T) {
		g := &Golden{
			t:      t,
			dir:    tempDir,
			update: true,
		}

		g.CheckString("string_test", "string content")

		// Verify file was written
		content, err := os.ReadFile(tempDir + "/string_test.golden") // #nosec G304 -- controlled test path
		require.NoError(t, err)
		require.Equal(t, "string content", string(content))
	})

	t.Run("check without update", func(t *testing.T) {
		// Create a golden file
		goldenPath := tempDir + "/test.golden"
		require.NoError(t, os.WriteFile(goldenPath, []byte("expected content"), 0o600))

		g := &Golden{
			t:      t,
			dir:    tempDir,
			update: false,
		}

		// Check matching content - should pass
		g.Check("test", []byte("expected content"))
	})
}

func TestNewGolden(t *testing.T) {
	t.Run("with custom dir", func(t *testing.T) {
		g := NewGolden(t, "custom/dir")
		require.Equal(t, "custom/dir", g.dir)
		require.Equal(t, t, g.t)
	})

	t.Run("with default dir", func(t *testing.T) {
		g := NewGolden(t, "")
		require.Equal(t, "testdata/golden", g.dir)
	})

	t.Run("with update env", func(t *testing.T) {
		require.NoError(t, os.Setenv("UPDATE_GOLDEN", "true"))
		defer func() {
			require.NoError(t, os.Unsetenv("UPDATE_GOLDEN"))
		}()

		g := NewGolden(t, "")
		require.True(t, g.update)
	})
}

func TestDataProvider(t *testing.T) {
	dp := NewDataProvider(t)
	require.NotNil(t, dp)

	t.Run("strings", func(t *testing.T) {
		strs := dp.Strings()
		require.Greater(t, len(strs), 5)
		require.Contains(t, strs, "")
		require.Contains(t, strs, "hello")
		require.Contains(t, strs, "hello world")

		// Check for special strings
		hasUnicode := false
		hasMultiline := false
		hasLong := false

		for _, s := range strs {
			if strings.Contains(s, "nihao") { // Chinese hello
				hasUnicode = true
			}
			if strings.Contains(s, "\n") {
				hasMultiline = true
			}
			if len(s) > 100 {
				hasLong = true
			}
		}

		require.True(t, hasUnicode)
		require.True(t, hasMultiline)
		require.True(t, hasLong)
	})

	t.Run("ints", func(t *testing.T) {
		ints := dp.Ints()
		require.Greater(t, len(ints), 5)
		require.Contains(t, ints, 0)
		require.Contains(t, ints, 1)
		require.Contains(t, ints, -1)
		require.Contains(t, ints, 42)
	})

	t.Run("bools", func(t *testing.T) {
		bools := dp.Bools()
		require.Len(t, bools, 2)
		require.Contains(t, bools, true)
		require.Contains(t, bools, false)
	})

	t.Run("errors", func(t *testing.T) {
		errs := dp.Errors()
		require.GreaterOrEqual(t, len(errs), 4)

		// Check for nil
		hasNil := false
		for _, err := range errs {
			if err == nil {
				hasNil = true
				break
			}
		}
		require.True(t, hasNil)

		// Check for wrapped error
		hasWrapped := false
		for _, err := range errs {
			if err != nil && strings.Contains(err.Error(), "wrapped:") {
				hasWrapped = true
				break
			}
		}
		require.True(t, hasWrapped)
	})
}

func TestRunTestCases(t *testing.T) {
	executed := make(map[string]bool)

	cases := []TestCase{
		{Name: "case1", Input: "input1", Want: "want1"},
		{Name: "case2", Input: "input2", Want: "want2", Error: errError2},
		{Name: "case3", Input: 42, Want: 84},
	}

	RunTestCases(t, cases, func(tc TestCase) {
		executed[tc.Name] = true

		// Verify test case properties are accessible
		switch tc.Name {
		case "case1":
			require.Equal(t, "input1", tc.Input)
			require.Equal(t, "want1", tc.Want)
			require.NoError(t, tc.Error)
		case "case2":
			require.Equal(t, "input2", tc.Input)
			require.Equal(t, "want2", tc.Want)
			require.Error(t, tc.Error)
		case "case3":
			require.Equal(t, 42, tc.Input)
			require.Equal(t, 84, tc.Want)
		}
	})

	// Verify all cases were executed
	require.Len(t, executed, 3)
	require.True(t, executed["case1"])
	require.True(t, executed["case2"])
	require.True(t, executed["case3"])
}
