// ci_failure_scenarios_test.go
//
// TEMPORARY FILE - DELETE AFTER CI TESTING
//
// This file contains intentional test failures for validating CI workflow:
// 1. Nested subtest failure - tests hierarchical test name handling
// 2. Panic - tests panic detection and stack trace capture
// 3. Fuzz test failure - tests fuzz failure detection
//
// Run individually:
//   go test -run TestCI_NestedSubtestFailure ./pkg/mage/...
//   go test -run TestCI_PanicFailure ./pkg/mage/...
//   go test -fuzz=FuzzCI_IntentionalFailure -fuzztime=5s ./pkg/mage/...
//
// Run with CI mode:
//   CI=true mage test:unit ./pkg/mage/...

package mage

import (
	"strings"
	"testing"
)

// TestCI_NestedSubtestFailure creates a deeply nested subtest that fails.
// This tests CI's ability to handle hierarchical test names like:
// TestCI_NestedSubtestFailure/Level1/Level2/Level3_Fails
func TestCI_NestedSubtestFailure(t *testing.T) {
	t.Run("Level1", func(t *testing.T) {
		t.Run("Level2", func(t *testing.T) {
			t.Run("Level3_Fails", func(t *testing.T) {
				t.Errorf("intentional nested subtest failure for CI testing")
			})
		})
	})
}

// TestCI_PanicFailure triggers a panic via nil pointer dereference.
// This tests CI's panic detection via the "panic:" pattern and stack trace capture.
func TestCI_PanicFailure(t *testing.T) {
	var ptr *int
	_ = *ptr // nil pointer dereference - produces a panic with stack trace
}

// FuzzCI_IntentionalFailure is a fuzz test that fails on seed corpus.
// This tests CI's fuzz failure detection.
func FuzzCI_IntentionalFailure(f *testing.F) {
	// Seed corpus - "fail_now" will trigger immediate failure
	f.Add("valid")
	f.Add("fail_now")

	f.Fuzz(func(t *testing.T, input string) {
		if input == "fail_now" || strings.Contains(input, "fail") {
			t.Errorf("intentional fuzz failure: input contained failure trigger '%s'", input)
		}
	})
}
