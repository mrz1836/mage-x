package env

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestDefaultEnvScope_EmptyValue tests that modifying an env var with an empty
// value is tracked correctly. This documents a behavior where empty string values
// are treated as "not set" for change detection purposes.
func TestDefaultEnvScope_EmptyValue(t *testing.T) {
	manager := NewDefaultEnvManager()
	testKey := "SCOPE_EMPTY_VALUE_TEST"

	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv(testKey))
	})

	t.Run("empty_to_value_is_set_action", func(t *testing.T) {
		require.NoError(t, os.Unsetenv(testKey))

		scope := manager.PushScope()
		defer manager.PopScope() //nolint:errcheck // Cleanup in defer

		err := scope.Set(testKey, "new_value")
		require.NoError(t, err)

		changes := scope.Changes()
		change, exists := changes[testKey]
		require.True(t, exists)
		require.Equal(t, ActionSet, change.Action)
		require.Empty(t, change.OldValue)
		require.Equal(t, "new_value", change.NewValue)
	})

	t.Run("value_to_new_value_is_modify_action", func(t *testing.T) {
		require.NoError(t, os.Setenv(testKey, "original"))

		scope := manager.PushScope()
		defer manager.PopScope() //nolint:errcheck // Cleanup in defer

		err := scope.Set(testKey, "modified")
		require.NoError(t, err)

		changes := scope.Changes()
		change, exists := changes[testKey]
		require.True(t, exists)
		require.Equal(t, ActionModify, change.Action)
		require.Equal(t, "original", change.OldValue)
		require.Equal(t, "modified", change.NewValue)
	})

	t.Run("empty_string_treated_as_not_set", func(t *testing.T) {
		// Set env var to empty string
		require.NoError(t, os.Setenv(testKey, ""))

		scope := manager.PushScope()
		defer manager.PopScope() //nolint:errcheck // Cleanup in defer

		err := scope.Set(testKey, "value")
		require.NoError(t, err)

		changes := scope.Changes()
		change, exists := changes[testKey]
		require.True(t, exists)
		// Note: Empty string is treated as "not set", so action is Set not Modify
		// This documents the current behavior - empty values are not distinguished
		// from missing values for change tracking purposes
		require.Equal(t, ActionSet, change.Action,
			"empty string value is treated as 'not set' - action is Set, not Modify")
	})

	t.Run("unset_records_old_value", func(t *testing.T) {
		require.NoError(t, os.Setenv(testKey, "to_unset"))

		scope := manager.PushScope()
		defer manager.PopScope() //nolint:errcheck // Cleanup in defer

		err := scope.Unset(testKey)
		require.NoError(t, err)

		changes := scope.Changes()
		change, exists := changes[testKey]
		require.True(t, exists)
		require.Equal(t, ActionUnset, change.Action)
		require.Equal(t, "to_unset", change.OldValue)
		require.Empty(t, change.NewValue)
	})
}

// TestDefaultEnvScope_Rollback tests rollback functionality.
func TestDefaultEnvScope_Rollback(t *testing.T) {
	manager := NewDefaultEnvManager()

	t.Run("rollback_restores_original_values", func(t *testing.T) {
		testKey := "ROLLBACK_TEST_VAR"
		require.NoError(t, os.Setenv(testKey, "original"))
		t.Cleanup(func() {
			require.NoError(t, os.Unsetenv(testKey))
		})

		scope := manager.PushScope()
		defer manager.PopScope() //nolint:errcheck // Cleanup in defer

		// Make changes
		err := scope.Set(testKey, "modified")
		require.NoError(t, err)
		require.Equal(t, "modified", os.Getenv(testKey))

		// Rollback
		err = scope.Rollback()
		require.NoError(t, err)

		// Verify original value restored
		require.Equal(t, "original", os.Getenv(testKey))
		require.False(t, scope.HasChanges())
	})

	t.Run("rollback_removes_new_variables", func(t *testing.T) {
		testKey := "ROLLBACK_NEW_VAR"
		require.NoError(t, os.Unsetenv(testKey))
		t.Cleanup(func() {
			require.NoError(t, os.Unsetenv(testKey))
		})

		scope := manager.PushScope()
		defer manager.PopScope() //nolint:errcheck // Cleanup in defer

		// Set new variable
		err := scope.Set(testKey, "new_value")
		require.NoError(t, err)
		require.True(t, scope.HasChanges())

		// Rollback
		err = scope.Rollback()
		require.NoError(t, err)

		// Variable should be unset
		_, exists := os.LookupEnv(testKey)
		require.False(t, exists, "variable should be removed after rollback")
	})

	t.Run("rollback_restores_unset_variables", func(t *testing.T) {
		testKey := "ROLLBACK_UNSET_VAR"
		require.NoError(t, os.Setenv(testKey, "original"))
		t.Cleanup(func() {
			require.NoError(t, os.Unsetenv(testKey))
		})

		scope := manager.PushScope()
		defer manager.PopScope() //nolint:errcheck // Cleanup in defer

		// Unset the variable
		err := scope.Unset(testKey)
		require.NoError(t, err)
		_, exists := os.LookupEnv(testKey)
		require.False(t, exists)

		// Rollback
		err = scope.Rollback()
		require.NoError(t, err)

		// Variable should be restored
		require.Equal(t, "original", os.Getenv(testKey))
	})
}

// TestDefaultEnvValidator_MultipleErrors tests that ValidateAll returns all
// errors for a key with multiple rules, not just the first error.
func TestDefaultEnvValidator_MultipleErrors(t *testing.T) {
	t.Run("multiple_rules_same_key", func(t *testing.T) {
		validator := NewDefaultEnvValidator()
		testKey := "MULTI_RULE_VAR"

		require.NoError(t, os.Unsetenv(testKey))
		t.Cleanup(func() {
			require.NoError(t, os.Unsetenv(testKey))
		})

		// Add multiple rules that will all fail
		validator.Required(testKey)
		validator.Pattern(testKey, `^[a-z]+$`)

		// With empty value, both rules should fail
		errors := validator.ValidateAll()
		require.Len(t, errors, 2, "should return errors from both rules")

		// Verify both error types are present
		hasRequired := false
		hasPattern := false
		for _, err := range errors {
			if err.Rule == "required" {
				hasRequired = true
			}
			// Pattern rule Description includes the pattern itself
			if err.Rule == "pattern: ^[a-z]+$" {
				hasPattern = true
			}
		}
		require.True(t, hasRequired, "should have required rule error")
		require.True(t, hasPattern, "should have pattern rule error")
	})

	t.Run("all_rules_pass", func(t *testing.T) {
		validator := NewDefaultEnvValidator()
		testKey := "ALL_PASS_VAR"

		require.NoError(t, os.Setenv(testKey, "validvalue"))
		t.Cleanup(func() {
			require.NoError(t, os.Unsetenv(testKey))
		})

		validator.Required(testKey)
		validator.Pattern(testKey, `^[a-z]+$`)

		errors := validator.ValidateAll()
		require.Empty(t, errors, "should return no errors when all rules pass")
	})

	t.Run("partial_rule_failure", func(t *testing.T) {
		validator := NewDefaultEnvValidator()
		testKey := "PARTIAL_FAIL_VAR"

		// Set a numeric value that passes Required but fails pattern
		require.NoError(t, os.Setenv(testKey, "12345"))
		t.Cleanup(func() {
			require.NoError(t, os.Unsetenv(testKey))
		})

		validator.Required(testKey)
		validator.Pattern(testKey, `^[a-z]+$`)

		errors := validator.ValidateAll()
		require.Len(t, errors, 1, "should return only the failing rule error")
		require.Equal(t, "pattern: ^[a-z]+$", errors[0].Rule)
	})

	t.Run("multiple_keys_multiple_rules", func(t *testing.T) {
		validator := NewDefaultEnvValidator()

		require.NoError(t, os.Setenv("KEY1", ""))    // Will fail required
		require.NoError(t, os.Setenv("KEY2", "123")) // Will fail pattern
		t.Cleanup(func() {
			require.NoError(t, os.Unsetenv("KEY1"))
			require.NoError(t, os.Unsetenv("KEY2"))
		})

		validator.Required("KEY1")
		validator.Pattern("KEY2", `^[a-z]+$`)

		errors := validator.ValidateAll()
		require.Len(t, errors, 2)
	})
}

// TestDefaultEnvValidator_Validate tests single key validation.
func TestDefaultEnvValidator_Validate(t *testing.T) {
	t.Run("stops_on_first_error", func(t *testing.T) {
		validator := NewDefaultEnvValidator()
		testKey := "VALIDATE_SINGLE"

		// Add multiple rules
		validator.Required(testKey)
		validator.NotEmpty(testKey)
		validator.Pattern(testKey, `^[a-z]+$`)

		// Validate empty value - should return first error only
		err := validator.Validate(testKey, "")
		require.Error(t, err)
		require.ErrorIs(t, err, errValueRequired)
	})

	t.Run("no_rules_returns_nil", func(t *testing.T) {
		validator := NewDefaultEnvValidator()
		err := validator.Validate("NO_RULES_KEY", "any_value")
		require.NoError(t, err)
	})
}

// TestDefaultEnvContext_Diff_EdgeCases tests edge cases in context diffing.
func TestDefaultEnvContext_Diff_EdgeCases(t *testing.T) {
	t.Run("identical_contexts", func(t *testing.T) {
		ctx1 := &DefaultEnvContext{
			timestamp: time.Now(),
			variables: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
			},
		}

		ctx2 := &DefaultEnvContext{
			timestamp: time.Now(),
			variables: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
			},
		}

		diff := ctx1.Diff(ctx2)
		require.Empty(t, diff, "identical contexts should have empty diff")
	})

	t.Run("empty_context_vs_populated", func(t *testing.T) {
		emptyCtx := &DefaultEnvContext{
			timestamp: time.Now(),
			variables: map[string]string{},
		}

		populatedCtx := &DefaultEnvContext{
			timestamp: time.Now(),
			variables: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
			},
		}

		// Diff empty against populated: empty has nothing that populated doesn't
		// But populated has things that empty doesn't → deletions
		diff := emptyCtx.Diff(populatedCtx)
		require.Len(t, diff, 2, "should detect 2 deletions")
		for _, change := range diff {
			require.Equal(t, ActionUnset, change.Action)
		}
	})

	t.Run("populated_vs_empty_context", func(t *testing.T) {
		populatedCtx := &DefaultEnvContext{
			timestamp: time.Now(),
			variables: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
			},
		}

		emptyCtx := &DefaultEnvContext{
			timestamp: time.Now(),
			variables: map[string]string{},
		}

		// Diff populated against empty: populated has additions
		diff := populatedCtx.Diff(emptyCtx)
		require.Len(t, diff, 2, "should detect 2 additions")
		for _, change := range diff {
			require.Equal(t, ActionSet, change.Action)
		}
	})

	t.Run("large_context", func(t *testing.T) {
		// Create two contexts with 1000 variables
		vars1 := make(map[string]string)
		vars2 := make(map[string]string)
		for i := 0; i < 1000; i++ {
			key := "LARGE_VAR_" + string(rune('A'+i%26)) + "_" + string(rune('0'+i%10))
			vars1[key] = "value_v1"
			vars2[key] = "value_v1"
		}

		// Modify 10 variables in vars2
		for i := 0; i < 10; i++ {
			key := "LARGE_VAR_" + string(rune('A'+i%26)) + "_" + string(rune('0'+i%10))
			vars2[key] = "value_v2_modified"
		}

		ctx1 := &DefaultEnvContext{timestamp: time.Now(), variables: vars1}
		ctx2 := &DefaultEnvContext{timestamp: time.Now(), variables: vars2}

		diff := ctx1.Diff(ctx2)
		require.Len(t, diff, 10, "should detect 10 modifications")
	})
}

// TestDefaultEnvContext_Merge_EdgeCases tests edge cases in context merging.
func TestDefaultEnvContext_Merge_EdgeCases(t *testing.T) {
	t.Run("merge_empty_contexts", func(t *testing.T) {
		ctx1 := &DefaultEnvContext{
			timestamp: time.Now(),
			variables: map[string]string{},
		}
		ctx2 := &DefaultEnvContext{
			timestamp: time.Now(),
			variables: map[string]string{},
		}

		merged := ctx1.Merge(ctx2)
		require.Equal(t, 0, merged.Count())
	})

	t.Run("merge_preserves_first_adds_second", func(t *testing.T) {
		ctx1 := &DefaultEnvContext{
			timestamp: time.Now(),
			variables: map[string]string{
				"ONLY_IN_FIRST": "first_value",
				"SHARED":        "first_shared",
			},
		}
		ctx2 := &DefaultEnvContext{
			timestamp: time.Now(),
			variables: map[string]string{
				"ONLY_IN_SECOND": "second_value",
				"SHARED":         "second_shared",
			},
		}

		merged := ctx1.Merge(ctx2)
		vars := merged.Variables()

		require.Equal(t, "first_value", vars["ONLY_IN_FIRST"])
		require.Equal(t, "second_value", vars["ONLY_IN_SECOND"])
		require.Equal(t, "second_shared", vars["SHARED"],
			"second context should override first")
	})

	t.Run("merge_timestamp_is_now", func(t *testing.T) {
		oldTime := time.Now().Add(-time.Hour)
		ctx1 := &DefaultEnvContext{
			timestamp: oldTime,
			variables: map[string]string{"VAR": "value"},
		}
		ctx2 := &DefaultEnvContext{
			timestamp: oldTime,
			variables: map[string]string{},
		}

		before := time.Now()
		merged := ctx1.Merge(ctx2)
		after := time.Now()

		mergedTime := merged.Timestamp()
		require.True(t, mergedTime.After(before) || mergedTime.Equal(before))
		require.True(t, mergedTime.Before(after) || mergedTime.Equal(after))
	})
}

// TestDefaultEnvManager_PopScope_Errors tests error cases for PopScope.
func TestDefaultEnvManager_PopScope_Errors(t *testing.T) {
	t.Run("pop_empty_stack", func(t *testing.T) {
		manager := NewDefaultEnvManager()

		err := manager.PopScope()
		require.Error(t, err)
		require.ErrorIs(t, err, errNoScopesToPop)
	})

	t.Run("pop_after_pop", func(t *testing.T) {
		manager := NewDefaultEnvManager()

		// Push and pop one scope
		manager.PushScope()
		err := manager.PopScope()
		require.NoError(t, err)

		// Try to pop again
		err = manager.PopScope()
		require.Error(t, err)
		require.ErrorIs(t, err, errNoScopesToPop)
	})

	t.Run("nested_scopes", func(t *testing.T) {
		manager := NewDefaultEnvManager()

		// Push multiple scopes
		scope1 := manager.PushScope()
		scope2 := manager.PushScope()
		require.NotNil(t, scope1)
		require.NotNil(t, scope2)

		// Pop both scopes
		err := manager.PopScope()
		require.NoError(t, err)
		err = manager.PopScope()
		require.NoError(t, err)

		// Third pop should fail
		err = manager.PopScope()
		require.ErrorIs(t, err, errNoScopesToPop)
	})
}

// TestDefaultEnvScope_Commit tests commit functionality.
func TestDefaultEnvScope_Commit(t *testing.T) {
	manager := NewDefaultEnvManager()
	testKey := "COMMIT_TEST_VAR"

	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv(testKey))
	})

	t.Run("commit_clears_changes", func(t *testing.T) {
		require.NoError(t, os.Unsetenv(testKey))

		scope := manager.PushScope()
		defer manager.PopScope() //nolint:errcheck // Cleanup in defer

		// Make changes
		err := scope.Set(testKey, "new_value")
		require.NoError(t, err)
		require.True(t, scope.HasChanges())

		// Commit
		err = scope.Commit()
		require.NoError(t, err)
		require.False(t, scope.HasChanges())

		// Value should still be set
		require.Equal(t, "new_value", os.Getenv(testKey))
	})
}

// TestDefaultEnvContext_Export tests the Export method.
func TestDefaultEnvContext_Export(t *testing.T) {
	ctx := &DefaultEnvContext{
		timestamp: time.Now(),
		variables: map[string]string{
			"VAR1": "value1",
			"VAR2": "value2",
		},
	}

	exported := ctx.Export()
	require.Len(t, exported, 2)
	require.Equal(t, "value1", exported["VAR1"])
	require.Equal(t, "value2", exported["VAR2"])

	// Verify Export returns a copy (modifying export doesn't affect original)
	exported["VAR3"] = "value3"
	require.Equal(t, 2, ctx.Count())
}
