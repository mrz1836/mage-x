package mage

import (
	"strings"
	"testing"
)

func TestTextStreamParser_ParseLine_FailLine(t *testing.T) {
	t.Parallel()

	t.Run("basic fuzz failure", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		err := parser.ParseLine("--- FAIL: FuzzGetToken (0.34s)")
		if err != nil {
			t.Fatalf("ParseLine() error = %v", err)
		}

		_, _, failed, _ := parser.GetStats()
		if failed != 1 {
			t.Errorf("GetStats().failed = %d, want 1", failed)
		}

		failures := parser.Flush()
		if len(failures) != 1 {
			t.Fatalf("Flush() returned %d failures, want 1", len(failures))
		}

		if failures[0].Test != "FuzzGetToken" {
			t.Errorf("failure.Test = %q, want %q", failures[0].Test, "FuzzGetToken")
		}
		if failures[0].Duration != "340ms" {
			t.Errorf("failure.Duration = %q, want %q", failures[0].Duration, "340ms")
		}
		if failures[0].Type != FailureTypeFuzz {
			t.Errorf("failure.Type = %q, want %q", failures[0].Type, FailureTypeFuzz)
		}
	})

	t.Run("failure with milliseconds", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		err := parser.ParseLine("--- FAIL: FuzzTest (50ms)")
		if err != nil {
			t.Fatalf("ParseLine() error = %v", err)
		}

		failures := parser.Flush()
		if len(failures) != 1 {
			t.Fatalf("got %d failures, want 1", len(failures))
		}

		// 50ms = 0.05s, which formats as "50ms"
		if failures[0].Duration != "50ms" {
			t.Errorf("failure.Duration = %q, want %q", failures[0].Duration, "50ms")
		}
	})

	t.Run("multiple failures", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, false) // No dedup

		lines := []string{
			"--- FAIL: FuzzOne (0.10s)",
			"    error output line 1",
			"--- FAIL: FuzzTwo (0.20s)",
			"    error output line 2",
		}

		for _, line := range lines {
			if err := parser.ParseLine(line); err != nil {
				t.Fatalf("ParseLine(%q) error = %v", line, err)
			}
		}

		failures := parser.Flush()
		if len(failures) != 2 {
			t.Fatalf("got %d failures, want 2", len(failures))
		}

		tests := map[string]bool{}
		for _, f := range failures {
			tests[f.Test] = true
		}

		if !tests["FuzzOne"] {
			t.Error("missing FuzzOne failure")
		}
		if !tests["FuzzTwo"] {
			t.Error("missing FuzzTwo failure")
		}
	})
}

func TestTextStreamParser_Deduplication(t *testing.T) {
	t.Parallel()

	t.Run("keeps parent test with longer elapsed", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		// Simulate nested test output where child fails first, then parent
		lines := []string{
			"--- FAIL: FuzzGetToken (0.01s)", // Child test (shorter elapsed)
			"    child error",
			"--- FAIL: FuzzGetToken (0.34s)", // Parent test (longer elapsed)
			"    parent error",
		}

		for _, line := range lines {
			if err := parser.ParseLine(line); err != nil {
				t.Fatalf("ParseLine(%q) error = %v", line, err)
			}
		}

		failures := parser.Flush()
		if len(failures) != 1 {
			t.Fatalf("got %d failures, want 1 (deduplicated)", len(failures))
		}

		// Should keep the parent (longer elapsed = 0.34s)
		if failures[0].Duration != "340ms" {
			t.Errorf("failure.Duration = %q, want %q (parent)", failures[0].Duration, "340ms")
		}
	})

	t.Run("no dedup when disabled", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, false) // Dedup disabled

		lines := []string{
			"--- FAIL: FuzzGetToken (0.01s)",
			"--- FAIL: FuzzGetToken (0.34s)",
		}

		for _, line := range lines {
			if err := parser.ParseLine(line); err != nil {
				t.Fatalf("ParseLine(%q) error = %v", line, err)
			}
		}

		failures := parser.Flush()
		if len(failures) != 2 {
			t.Errorf("got %d failures, want 2 (no dedup)", len(failures))
		}
	})
}

func TestTextStreamParser_OutputCapture(t *testing.T) {
	t.Parallel()

	t.Run("captures error output", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		lines := []string{
			"--- FAIL: FuzzTest (0.10s)",
			"    fuzz_test.go:42: expected 1, got 2",
			"    additional context line",
			"FAIL github.com/example/pkg 0.123s",
		}

		for _, line := range lines {
			if err := parser.ParseLine(line); err != nil {
				t.Fatalf("ParseLine error = %v", err)
			}
		}

		failures := parser.Flush()
		if len(failures) != 1 {
			t.Fatalf("got %d failures, want 1", len(failures))
		}

		if !strings.Contains(failures[0].Output, "expected 1, got 2") {
			t.Error("output should contain error message")
		}
		if !strings.Contains(failures[0].Output, "additional context") {
			t.Error("output should contain context line")
		}
	})

	t.Run("extracts file:line location", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		lines := []string{
			"--- FAIL: FuzzTest (0.10s)",
			"    token_test.go:42: assertion failed",
		}

		for _, line := range lines {
			if err := parser.ParseLine(line); err != nil {
				t.Fatalf("ParseLine error = %v", err)
			}
		}

		failures := parser.Flush()
		if len(failures) != 1 {
			t.Fatalf("got %d failures, want 1", len(failures))
		}

		if failures[0].File != "token_test.go" {
			t.Errorf("failure.File = %q, want %q", failures[0].File, "token_test.go")
		}
		if failures[0].Line != 42 {
			t.Errorf("failure.Line = %d, want %d", failures[0].Line, 42)
		}
	})

	t.Run("extracts error message", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		lines := []string{
			"--- FAIL: FuzzTest (0.10s)",
			"    test.go:10: expected true, got false",
		}

		for _, line := range lines {
			if err := parser.ParseLine(line); err != nil {
				t.Fatalf("ParseLine error = %v", err)
			}
		}

		failures := parser.Flush()
		if len(failures) != 1 {
			t.Fatalf("got %d failures, want 1", len(failures))
		}

		if failures[0].Error != "expected true, got false" {
			t.Errorf("failure.Error = %q, want %q", failures[0].Error, "expected true, got false")
		}
	})
}

func TestTextStreamParser_FailureTypeDetection(t *testing.T) {
	t.Parallel()

	t.Run("detects panic", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		lines := []string{
			"--- FAIL: FuzzTest (0.10s)",
			"panic: runtime error: index out of range",
			"goroutine 1 [running]:",
		}

		for _, line := range lines {
			if err := parser.ParseLine(line); err != nil {
				t.Fatalf("ParseLine error = %v", err)
			}
		}

		failures := parser.Flush()
		if len(failures) != 1 {
			t.Fatalf("got %d failures, want 1", len(failures))
		}

		if failures[0].Type != FailureTypePanic {
			t.Errorf("failure.Type = %q, want %q", failures[0].Type, FailureTypePanic)
		}
		if failures[0].Error != "runtime error: index out of range" {
			t.Errorf("failure.Error = %q, want panic message", failures[0].Error)
		}
	})

	t.Run("detects data race", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		lines := []string{
			"--- FAIL: FuzzTest (0.10s)",
			"WARNING: DATA RACE",
			"Write at 0x00c000123456 by goroutine 7:",
		}

		for _, line := range lines {
			if err := parser.ParseLine(line); err != nil {
				t.Fatalf("ParseLine error = %v", err)
			}
		}

		failures := parser.Flush()
		if len(failures) != 1 {
			t.Fatalf("got %d failures, want 1", len(failures))
		}

		if failures[0].Type != FailureTypeRace {
			t.Errorf("failure.Type = %q, want %q", failures[0].Type, FailureTypeRace)
		}
	})

	t.Run("detects fuzz corpus", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		lines := []string{
			"--- FAIL: FuzzGetToken (0.34s)",
			"    Failing input written to testdata/fuzz/FuzzGetToken/582528ddfad69eb5",
		}

		for _, line := range lines {
			if err := parser.ParseLine(line); err != nil {
				t.Fatalf("ParseLine error = %v", err)
			}
		}

		failures := parser.Flush()
		if len(failures) != 1 {
			t.Fatalf("got %d failures, want 1", len(failures))
		}

		if failures[0].Type != FailureTypeFuzz {
			t.Errorf("failure.Type = %q, want %q", failures[0].Type, FailureTypeFuzz)
		}
		if failures[0].FuzzInfo == nil {
			t.Fatal("failure.FuzzInfo should not be nil")
		}
		expectedCorpus := "testdata/fuzz/FuzzGetToken/582528ddfad69eb5"
		if failures[0].FuzzInfo.CorpusPath != expectedCorpus {
			t.Errorf("FuzzInfo.CorpusPath = %q, want %q", failures[0].FuzzInfo.CorpusPath, expectedCorpus)
		}
	})

	t.Run("detects fatal error", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		lines := []string{
			"--- FAIL: FuzzTest (0.10s)",
			"fatal error: concurrent map writes",
		}

		for _, line := range lines {
			if err := parser.ParseLine(line); err != nil {
				t.Fatalf("ParseLine error = %v", err)
			}
		}

		failures := parser.Flush()
		if len(failures) != 1 {
			t.Fatalf("got %d failures, want 1", len(failures))
		}

		if failures[0].Type != FailureTypeFatal {
			t.Errorf("failure.Type = %q, want %q", failures[0].Type, FailureTypeFatal)
		}
	})

	t.Run("detects runtime error", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		lines := []string{
			"--- FAIL: FuzzTest (0.10s)",
			"runtime error: slice bounds out of range",
		}

		for _, line := range lines {
			if err := parser.ParseLine(line); err != nil {
				t.Fatalf("ParseLine error = %v", err)
			}
		}

		failures := parser.Flush()
		if len(failures) != 1 {
			t.Fatalf("got %d failures, want 1", len(failures))
		}

		if failures[0].Type != FailureTypePanic {
			t.Errorf("failure.Type = %q, want %q (runtime errors are panic type)", failures[0].Type, FailureTypePanic)
		}
	})
}

func TestTextStreamParser_PackageResult(t *testing.T) {
	t.Parallel()

	t.Run("assigns package to failures", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		lines := []string{
			"--- FAIL: FuzzTest (0.10s)",
			"    error output",
			"FAIL github.com/example/pkg 0.123s",
		}

		for _, line := range lines {
			if err := parser.ParseLine(line); err != nil {
				t.Fatalf("ParseLine error = %v", err)
			}
		}

		failures := parser.Flush()
		if len(failures) != 1 {
			t.Fatalf("got %d failures, want 1", len(failures))
		}

		if failures[0].Package != "github.com/example/pkg" {
			t.Errorf("failure.Package = %q, want %q", failures[0].Package, "github.com/example/pkg")
		}
	})

	t.Run("handles ok package line", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		err := parser.ParseLine("ok github.com/example/pkg 0.123s")
		if err != nil {
			t.Fatalf("ParseLine error = %v", err)
		}

		// Should not create any failures
		failures := parser.Flush()
		if len(failures) != 0 {
			t.Errorf("got %d failures, want 0 for ok package", len(failures))
		}
	})
}

func TestTextStreamParser_PassLine(t *testing.T) {
	t.Parallel()

	t.Run("counts passed tests", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		lines := []string{
			"--- PASS: TestOne (0.01s)",
			"--- PASS: TestTwo (0.02s)",
		}

		for _, line := range lines {
			if err := parser.ParseLine(line); err != nil {
				t.Fatalf("ParseLine error = %v", err)
			}
		}

		total, passed, failed, _ := parser.GetStats()
		if total != 2 {
			t.Errorf("total = %d, want 2", total)
		}
		if passed != 2 {
			t.Errorf("passed = %d, want 2", passed)
		}
		if failed != 0 {
			t.Errorf("failed = %d, want 0", failed)
		}
	})
}

func TestTextStreamParser_Parse(t *testing.T) {
	t.Parallel()

	t.Run("parses full output", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		input := `=== RUN   FuzzGetToken
=== RUN   FuzzGetToken/seed#0
--- FAIL: FuzzGetToken (0.34s)
    --- FAIL: FuzzGetToken/seed#0 (0.00s)
        token_test.go:42: expected valid token
    Failing input written to testdata/fuzz/FuzzGetToken/582528ddfad69eb5
FAIL
FAIL github.com/example/pkg 0.345s
`
		err := parser.Parse(strings.NewReader(input))
		if err != nil {
			t.Fatalf("Parse error = %v", err)
		}

		failures := parser.Flush()
		// With deduplication, should only have 1 failure (parent with longer elapsed)
		if len(failures) != 1 {
			t.Fatalf("got %d failures, want 1", len(failures))
		}

		if failures[0].Test != "FuzzGetToken" {
			t.Errorf("failure.Test = %q, want %q", failures[0].Test, "FuzzGetToken")
		}
		if failures[0].Package != "github.com/example/pkg" {
			t.Errorf("failure.Package = %q, want %q", failures[0].Package, "github.com/example/pkg")
		}
	})
}

func TestTextStreamParser_GetStats(t *testing.T) {
	t.Parallel()

	t.Run("returns correct stats", func(t *testing.T) {
		t.Parallel()
		parser := NewTextStreamParser(20, true)

		lines := []string{
			"--- PASS: TestOne (0.01s)",
			"--- FAIL: FuzzTwo (0.10s)",
			"--- PASS: TestThree (0.02s)",
		}

		for _, line := range lines {
			if err := parser.ParseLine(line); err != nil {
				t.Fatalf("ParseLine error = %v", err)
			}
		}

		total, passed, failed, skipped := parser.GetStats()
		if total != 3 {
			t.Errorf("total = %d, want 3", total)
		}
		if passed != 2 {
			t.Errorf("passed = %d, want 2", passed)
		}
		if failed != 1 {
			t.Errorf("failed = %d, want 1", failed)
		}
		if skipped != 0 {
			t.Errorf("skipped = %d, want 0", skipped)
		}
	})
}

func TestParseElapsed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		elapsed  string
		unit     string
		expected float64
	}{
		{"seconds", "0.34", "s", 0.34},
		{"seconds no unit", "0.34", "", 0.34},
		{"milliseconds", "50", "ms", 0.05},
		{"minutes", "1", "m", 60.0},
		{"invalid", "abc", "s", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := parseElapsed(tc.elapsed, tc.unit)
			if result != tc.expected {
				t.Errorf("parseElapsed(%q, %q) = %v, want %v", tc.elapsed, tc.unit, result, tc.expected)
			}
		})
	}
}

func TestCITestFailure_ElapsedSeconds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration string
		expected float64
	}{
		{"empty", "", 0},
		{"milliseconds", "340ms", 0.34},
		{"seconds", "1.5s", 1.5},
		{"duration format", "1m30s", 90.0},
		{"plain number", "2.5", 2.5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			f := CITestFailure{Duration: tc.duration}
			result := f.elapsedSeconds()
			if result != tc.expected {
				t.Errorf("elapsedSeconds() = %v, want %v", result, tc.expected)
			}
		})
	}
}

// Benchmarks

func BenchmarkTextStreamParser_ParseLine(b *testing.B) {
	parser := NewTextStreamParser(20, true)
	line := "--- FAIL: FuzzGetToken (0.34s)"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		//nolint:errcheck,gosec // Benchmark - ignore errors
		parser.ParseLine(line)
	}
}

func BenchmarkTextStreamParser_Parse(b *testing.B) {
	input := `=== RUN   FuzzGetToken
=== RUN   FuzzGetToken/seed#0
--- FAIL: FuzzGetToken (0.34s)
    --- FAIL: FuzzGetToken/seed#0 (0.00s)
        token_test.go:42: expected valid token
    Failing input written to testdata/fuzz/FuzzGetToken/582528ddfad69eb5
FAIL
FAIL github.com/example/pkg 0.345s
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parser := NewTextStreamParser(20, true)
		//nolint:errcheck,gosec // Benchmark - ignore errors
		parser.Parse(strings.NewReader(input))
		parser.Flush()
	}
}
