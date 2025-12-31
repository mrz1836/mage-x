package env

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockEnvironmentBasicOperations tests basic mock environment operations.
func TestMockEnvironmentBasicOperations(t *testing.T) {
	mock := NewMockEnvironment()

	t.Run("get_set_unset", func(t *testing.T) {
		// Initially empty
		require.Empty(t, mock.Get("KEY"))
		require.False(t, mock.Exists("KEY"))

		// Set a value
		require.NoError(t, mock.Set("KEY", "value"))
		require.Equal(t, "value", mock.Get("KEY"))
		require.True(t, mock.Exists("KEY"))

		// Unset
		require.NoError(t, mock.Unset("KEY"))
		require.Empty(t, mock.Get("KEY"))
		require.False(t, mock.Exists("KEY"))
	})

	t.Run("get_string_with_default", func(t *testing.T) {
		require.NoError(t, mock.Set("EXISTING", "exists"))
		defer func() { require.NoError(t, mock.Unset("EXISTING")) }()

		require.Equal(t, "exists", mock.GetString("EXISTING", "default"))
		require.Equal(t, "default", mock.GetString("NONEXISTENT", "default"))
	})

	t.Run("get_with_default_alias", func(t *testing.T) {
		require.NoError(t, mock.Set("ALIASED", "value"))
		defer func() { require.NoError(t, mock.Unset("ALIASED")) }()

		require.Equal(t, "value", mock.GetWithDefault("ALIASED", "fallback"))
		require.Equal(t, "fallback", mock.GetWithDefault("MISSING", "fallback"))
	})

	t.Run("get_bool", func(t *testing.T) {
		require.NoError(t, mock.Set("TRUE_VAL", "true"))
		require.NoError(t, mock.Set("ONE_VAL", "1"))
		require.NoError(t, mock.Set("FALSE_VAL", "false"))
		defer func() {
			require.NoError(t, mock.Unset("TRUE_VAL"))
			require.NoError(t, mock.Unset("ONE_VAL"))
			require.NoError(t, mock.Unset("FALSE_VAL"))
		}()

		require.True(t, mock.GetBool("TRUE_VAL", false))
		require.True(t, mock.GetBool("ONE_VAL", false))
		require.False(t, mock.GetBool("FALSE_VAL", true))
		require.True(t, mock.GetBool("MISSING", true))
	})

	t.Run("get_int_returns_default", func(t *testing.T) {
		// Mock GetInt always returns default (simplified implementation)
		require.Equal(t, 42, mock.GetInt("ANY", 42))
	})

	t.Run("get_int64_returns_default", func(t *testing.T) {
		require.Equal(t, int64(100), mock.GetInt64("ANY", 100))
	})

	t.Run("get_float64_returns_default", func(t *testing.T) {
		require.InDelta(t, 3.14, mock.GetFloat64("ANY", 3.14), 0.001)
	})

	t.Run("get_duration_returns_default", func(t *testing.T) {
		require.Equal(t, 5*time.Second, mock.GetDuration("ANY", 5*time.Second))
	})

	t.Run("get_string_slice_returns_default", func(t *testing.T) {
		expected := []string{"a", "b", "c"}
		require.Equal(t, expected, mock.GetStringSlice("ANY", expected))
	})
}

// TestMockEnvironmentGetWithPrefix tests prefix filtering.
func TestMockEnvironmentGetWithPrefix(t *testing.T) {
	mock := NewMockEnvironment()

	// Set some variables with and without prefix
	require.NoError(t, mock.Set("PREFIX_VAR1", "val1"))
	require.NoError(t, mock.Set("PREFIX_VAR2", "val2"))
	require.NoError(t, mock.Set("OTHER_VAR", "other"))
	defer func() {
		require.NoError(t, mock.Unset("PREFIX_VAR1"))
		require.NoError(t, mock.Unset("PREFIX_VAR2"))
		require.NoError(t, mock.Unset("OTHER_VAR"))
	}()

	result := mock.GetWithPrefix("PREFIX_")
	require.Len(t, result, 2)
	require.Equal(t, "val1", result["PREFIX_VAR1"])
	require.Equal(t, "val2", result["PREFIX_VAR2"])
	require.NotContains(t, result, "OTHER_VAR")
}

// TestMockEnvironmentClear tests the clear functionality.
func TestMockEnvironmentClear(t *testing.T) {
	mock := NewMockEnvironment()

	// Set some variables
	require.NoError(t, mock.Set("VAR1", "val1"))
	require.NoError(t, mock.Set("VAR2", "val2"))
	require.True(t, mock.Exists("VAR1"))
	require.True(t, mock.Exists("VAR2"))

	// Clear all
	require.NoError(t, mock.Clear())

	// Verify all cleared
	require.False(t, mock.Exists("VAR1"))
	require.False(t, mock.Exists("VAR2"))
	require.Empty(t, mock.GetAll())
}

// TestMockEnvironmentGetAll tests getting all variables.
func TestMockEnvironmentGetAll(t *testing.T) {
	mock := NewMockEnvironment()

	require.NoError(t, mock.Set("A", "1"))
	require.NoError(t, mock.Set("B", "2"))
	defer func() {
		require.NoError(t, mock.Unset("A"))
		require.NoError(t, mock.Unset("B"))
	}()

	all := mock.GetAll()
	require.Len(t, all, 2)
	require.Equal(t, "1", all["A"])
	require.Equal(t, "2", all["B"])
}

// TestMockEnvironmentSetMultiple tests setting multiple variables.
func TestMockEnvironmentSetMultiple(t *testing.T) {
	mock := NewMockEnvironment()

	vars := map[string]string{
		"MULTI_A": "valueA",
		"MULTI_B": "valueB",
	}

	require.NoError(t, mock.SetMultiple(vars))
	defer func() {
		require.NoError(t, mock.Unset("MULTI_A"))
		require.NoError(t, mock.Unset("MULTI_B"))
	}()

	require.Equal(t, "valueA", mock.Get("MULTI_A"))
	require.Equal(t, "valueB", mock.Get("MULTI_B"))
}

// TestMockEnvironmentValidate tests the validate function.
func TestMockEnvironmentValidate(t *testing.T) {
	mock := NewMockEnvironment()

	require.NoError(t, mock.Set("VALIDATABLE", "valid_value"))
	defer func() { require.NoError(t, mock.Unset("VALIDATABLE")) }()

	// Test with passing validator
	result := mock.Validate("VALIDATABLE", func(s string) bool {
		return s == "valid_value"
	})
	require.True(t, result)

	// Test with failing validator
	result = mock.Validate("VALIDATABLE", func(s string) bool {
		return s == "invalid"
	})
	require.False(t, result)
}

// TestMockEnvironmentRequired tests required variable checking.
func TestMockEnvironmentRequired(t *testing.T) {
	mock := NewMockEnvironment()

	require.NoError(t, mock.Set("REQUIRED_VAR", "value"))
	defer func() { require.NoError(t, mock.Unset("REQUIRED_VAR")) }()

	// Should pass for existing variable
	require.NoError(t, mock.Required("REQUIRED_VAR"))

	// Should fail for missing variable
	err := mock.Required("MISSING_VAR")
	require.Error(t, err)
}

// TestMockEnvironmentGetCallCount tests call tracking.
func TestMockEnvironmentGetCallCount(t *testing.T) {
	mock := NewMockEnvironment()

	// Initial counts should be zero
	require.Equal(t, 0, mock.GetCallCount("Get"))
	require.Equal(t, 0, mock.GetCallCount("Set"))

	// Make some calls
	_ = mock.Get("KEY")
	_ = mock.Get("KEY")
	require.NoError(t, mock.Set("KEY", "val"))

	// Verify counts
	require.Equal(t, 2, mock.GetCallCount("Get"))
	require.Equal(t, 1, mock.GetCallCount("Set"))
}

// TestMockPathResolverBasicPaths tests basic path resolver operations.
func TestMockPathResolverBasicPaths(t *testing.T) {
	mock := NewMockPathResolver()

	t.Run("home_default", func(t *testing.T) {
		home := mock.Home()
		require.Equal(t, "/mock/home", home)
	})

	t.Run("config_dir", func(t *testing.T) {
		configDir := mock.ConfigDir("myapp")
		require.Equal(t, "/mock/config/myapp", configDir)
	})

	t.Run("cache_dir", func(t *testing.T) {
		cacheDir := mock.CacheDir("myapp")
		require.Equal(t, "/mock/cache/myapp", cacheDir)
	})

	t.Run("data_dir", func(t *testing.T) {
		dataDir := mock.DataDir("myapp")
		require.Equal(t, "/mock/data/myapp", dataDir)
	})

	t.Run("temp_dir", func(t *testing.T) {
		tempDir := mock.TempDir()
		require.Equal(t, "/mock/tmp", tempDir)
	})

	t.Run("working_dir", func(t *testing.T) {
		workDir := mock.WorkingDir()
		require.Equal(t, "/mock/work", workDir)
	})

	t.Run("gopath", func(t *testing.T) {
		gopath := mock.GOPATH()
		require.Equal(t, "/mock/go", gopath)
	})

	t.Run("goroot", func(t *testing.T) {
		goroot := mock.GOROOT()
		require.Equal(t, "/mock/goroot", goroot)
	})

	t.Run("gocache", func(t *testing.T) {
		gocache := mock.GOCACHE()
		require.Equal(t, "/mock/gocache", gocache)
	})

	t.Run("gomodcache", func(t *testing.T) {
		gomodcache := mock.GOMODCACHE()
		require.Equal(t, "/mock/gomodcache", gomodcache)
	})
}

// TestMockPathResolverSetPath tests custom path setting.
func TestMockPathResolverSetPath(t *testing.T) {
	mock := NewMockPathResolver()

	// Set custom home path
	mock.SetPath("home", "/custom/home")
	require.Equal(t, "/custom/home", mock.Home())
}

// TestMockPathResolverPathOperations tests path manipulation operations.
func TestMockPathResolverPathOperations(t *testing.T) {
	mock := NewMockPathResolver()

	t.Run("expand_passthrough", func(t *testing.T) {
		// Mock just returns the path as-is
		require.Equal(t, "~/path", mock.Expand("~/path"))
	})

	t.Run("resolve_passthrough", func(t *testing.T) {
		resolved, err := mock.Resolve("/some/path")
		require.NoError(t, err)
		require.Equal(t, "/some/path", resolved)
	})

	t.Run("is_absolute", func(t *testing.T) {
		require.True(t, mock.IsAbsolute("/absolute"))
		require.False(t, mock.IsAbsolute("relative"))
	})

	t.Run("make_absolute", func(t *testing.T) {
		// Absolute path returned as-is
		abs, err := mock.MakeAbsolute("/already/absolute")
		require.NoError(t, err)
		require.Equal(t, "/already/absolute", abs)

		// Relative path prefixed with mock working dir
		abs, err = mock.MakeAbsolute("relative")
		require.NoError(t, err)
		require.Equal(t, "/mock/work/relative", abs)
	})

	t.Run("clean_passthrough", func(t *testing.T) {
		require.Equal(t, "dirty/path", mock.Clean("dirty/path"))
	})

	t.Run("ensure_dir", func(t *testing.T) {
		require.NoError(t, mock.EnsureDir("/any/path"))
	})

	t.Run("ensure_dir_with_mode", func(t *testing.T) {
		require.NoError(t, mock.EnsureDirWithMode("/any/path", 0o755))
	})
}

// TestMockPathResolverGetCallCount tests call tracking.
func TestMockPathResolverGetCallCount(t *testing.T) {
	mock := NewMockPathResolver()

	require.Equal(t, 0, mock.GetCallCount("Home"))

	_ = mock.Home()
	_ = mock.Home()

	require.Equal(t, 2, mock.GetCallCount("Home"))
}

// TestMockEnvManagerScopes tests scope management.
func TestMockEnvManagerScopes(t *testing.T) {
	mock := NewMockEnvManager()

	t.Run("push_pop_scope", func(t *testing.T) {
		// Pop on empty should fail
		err := mock.PopScope()
		require.Error(t, err)

		// Push creates a scope
		scope := mock.PushScope()
		require.NotNil(t, scope)

		// Pop should succeed
		require.NoError(t, mock.PopScope())
	})

	t.Run("with_scope", func(t *testing.T) {
		called := false
		err := mock.WithScope(func(s Scope) error {
			called = true
			require.NotNil(t, s)
			return nil
		})
		require.NoError(t, err)
		require.True(t, called)
	})

	t.Run("save_restore_context", func(t *testing.T) {
		ctx, err := mock.SaveContext()
		require.NoError(t, err)
		require.NotNil(t, ctx)

		require.NoError(t, mock.RestoreContext(ctx))
	})

	t.Run("isolate", func(t *testing.T) {
		called := false
		err := mock.Isolate(map[string]string{"KEY": "VALUE"}, func() error {
			called = true
			return nil
		})
		require.NoError(t, err)
		require.True(t, called)
	})

	t.Run("fork", func(t *testing.T) {
		forked := mock.Fork()
		require.NotNil(t, forked)
	})

	t.Run("get_call_count", func(t *testing.T) {
		// Reset mock
		newMock := NewMockEnvManager()
		require.Equal(t, 0, newMock.GetCallCount("PushScope"))

		_ = newMock.PushScope()
		require.Equal(t, 1, newMock.GetCallCount("PushScope"))
	})
}

// TestMockEnvScopeCommitRollback tests scope commit and rollback.
func TestMockEnvScopeCommitRollback(t *testing.T) {
	scope := NewMockEnvScope()

	t.Run("initial_no_changes", func(t *testing.T) {
		require.False(t, scope.HasChanges())
		require.Empty(t, scope.Changes())
	})

	t.Run("commit", func(t *testing.T) {
		require.NoError(t, scope.Commit())
		require.False(t, scope.HasChanges())
	})

	t.Run("rollback", func(t *testing.T) {
		require.NoError(t, scope.Rollback())
		require.False(t, scope.HasChanges())
	})
}

// TestMockEnvContextOperations tests context operations.
func TestMockEnvContextOperations(t *testing.T) {
	ctx := NewMockEnvContext()

	t.Run("timestamp", func(t *testing.T) {
		ts := ctx.Timestamp()
		require.False(t, ts.IsZero())
	})

	t.Run("variables_initially_empty", func(t *testing.T) {
		vars := ctx.Variables()
		require.Empty(t, vars)
	})

	t.Run("count", func(t *testing.T) {
		require.Equal(t, 0, ctx.Count())
	})

	t.Run("export", func(t *testing.T) {
		exported := ctx.Export()
		require.NotNil(t, exported)
	})

	t.Run("diff", func(t *testing.T) {
		other := NewMockEnvContext()
		diff := ctx.Diff(other)
		require.NotNil(t, diff)
	})

	t.Run("merge", func(t *testing.T) {
		other := NewMockEnvContext()
		merged := ctx.Merge(other)
		require.NotNil(t, merged)
	})
}

// mockValidationRule implements ValidationRule for testing.
type mockValidationRule struct {
	validateFunc func(string) error
	description  string
}

func (r *mockValidationRule) Validate(value string) error {
	if r.validateFunc != nil {
		return r.validateFunc(value)
	}
	return nil
}

func (r *mockValidationRule) Description() string {
	return r.description
}

// TestMockEnvValidatorRules tests validator rule management.
func TestMockEnvValidatorRules(t *testing.T) {
	validator := NewMockEnvValidator()

	t.Run("add_rule", func(t *testing.T) {
		rule := &mockValidationRule{
			validateFunc: func(s string) error { return nil },
			description:  "test rule",
		}
		err := validator.AddRule("KEY", rule)
		require.NoError(t, err)
	})

	t.Run("remove_rule", func(t *testing.T) {
		require.NoError(t, validator.RemoveRule("KEY"))
	})

	t.Run("validate", func(t *testing.T) {
		require.NoError(t, validator.Validate("KEY", "value"))
	})

	t.Run("validate_all", func(t *testing.T) {
		errs := validator.ValidateAll()
		require.Empty(t, errs)
	})

	t.Run("required_chaining", func(t *testing.T) {
		result := validator.Required("KEY1", "KEY2")
		require.Equal(t, validator, result)
	})

	t.Run("not_empty_chaining", func(t *testing.T) {
		result := validator.NotEmpty("KEY1")
		require.Equal(t, validator, result)
	})

	t.Run("pattern_chaining", func(t *testing.T) {
		result := validator.Pattern("KEY", ".*")
		require.Equal(t, validator, result)
	})

	t.Run("range_chaining", func(t *testing.T) {
		result := validator.Range("KEY", 0, 100)
		require.Equal(t, validator, result)
	})

	t.Run("one_of_chaining", func(t *testing.T) {
		result := validator.OneOf("KEY", "a", "b", "c")
		require.Equal(t, validator, result)
	})

	t.Run("get_call_count", func(t *testing.T) {
		newValidator := NewMockEnvValidator()
		err := newValidator.Validate("KEY", "value")
		require.NoError(t, err)
		require.Equal(t, 1, newValidator.GetCallCount("Validate"))
	})
}

// TestValidationErrorInterface tests the ValidationError error interface.
func TestValidationErrorInterface(t *testing.T) {
	t.Run("with_message", func(t *testing.T) {
		err := &ValidationError{Message: "custom message"}
		assert.Equal(t, "custom message", err.Error())
	})

	t.Run("without_message", func(t *testing.T) {
		err := &ValidationError{}
		assert.Equal(t, "validation error", err.Error())
	})

	t.Run("with_key", func(t *testing.T) {
		err := &ValidationError{Key: "MY_VAR", Message: "invalid"}
		assert.Equal(t, "invalid", err.Error())
	})
}
