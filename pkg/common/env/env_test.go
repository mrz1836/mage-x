package env

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultEnvironment(t *testing.T) {
	env := NewDefaultEnvironment()

	// Test Set and Get
	key := "TEST_VAR_123"
	value := "test_value"

	// Clean up at the end
	defer func() {
		if err := env.Unset(key); err != nil {
			t.Errorf("Failed to unset %s: %v", key, err)
		}
	}()

	if err := env.Set(key, value); err != nil {
		t.Fatalf("Failed to set environment variable: %v", err)
	}

	if got := env.Get(key); got != value {
		t.Errorf("Get() = %v, want %v", got, value)
	}

	// Test Exists
	if !env.Exists(key) {
		t.Error("Exists() should return true for existing variable")
	}

	// Test Unset
	if err := env.Unset(key); err != nil {
		t.Fatalf("Failed to unset environment variable: %v", err)
	}

	if env.Exists(key) {
		t.Error("Exists() should return false after unset")
	}
}

func TestDefaultEnvironment_TypedGetters(t *testing.T) {
	env := NewDefaultEnvironment()

	// Clean up at the end
	defer func() {
		for _, key := range []string{"TEST_BOOL", "TEST_INT", "TEST_INT64", "TEST_FLOAT64", "TEST_DURATION", "TEST_SLICE"} {
			if err := env.Unset(key); err != nil {
				t.Errorf("Failed to unset %s: %v", key, err)
			}
		}
	}()

	// Test GetBool
	require.NoError(t, env.Set("TEST_BOOL", "true"))
	if !env.GetBool("TEST_BOOL", false) {
		t.Error("GetBool() should return true for 'true'")
	}

	// Test GetInt
	require.NoError(t, env.Set("TEST_INT", "42"))
	if got := env.GetInt("TEST_INT", 0); got != 42 {
		t.Errorf("GetInt() = %v, want 42", got)
	}

	// Test GetInt64
	require.NoError(t, env.Set("TEST_INT64", "9223372036854775807"))
	if got := env.GetInt64("TEST_INT64", 0); got != 9223372036854775807 {
		t.Errorf("GetInt64() = %v, want 9223372036854775807", got)
	}

	// Test GetFloat64
	require.NoError(t, env.Set("TEST_FLOAT64", "3.14159"))
	if got := env.GetFloat64("TEST_FLOAT64", 0.0); got != 3.14159 {
		t.Errorf("GetFloat64() = %v, want 3.14159", got)
	}

	// Test GetDuration
	require.NoError(t, env.Set("TEST_DURATION", "5m30s"))
	expected := 5*time.Minute + 30*time.Second
	if got := env.GetDuration("TEST_DURATION", 0); got != expected {
		t.Errorf("GetDuration() = %v, want %v", got, expected)
	}

	// Test GetStringSlice
	require.NoError(t, env.Set("TEST_SLICE", "one,two,three"))
	expected_slice := []string{"one", "two", "three"}
	if got := env.GetStringSlice("TEST_SLICE", nil); !equalStringSlices(got, expected_slice) {
		t.Errorf("GetStringSlice() = %v, want %v", got, expected_slice)
	}
}

func TestDefaultEnvironment_Advanced(t *testing.T) {
	env := NewDefaultEnvironment()

	// Test SetMultiple
	vars := map[string]string{
		"TEST_VAR1": "value1",
		"TEST_VAR2": "value2",
		"TEST_VAR3": "value3",
	}

	defer func() {
		for key := range vars {
			if err := env.Unset(key); err != nil {
				t.Errorf("Failed to unset %s: %v", key, err)
			}
		}
	}()

	if err := env.SetMultiple(vars); err != nil {
		t.Fatalf("SetMultiple() failed: %v", err)
	}

	// Test GetWithPrefix
	prefixed := env.GetWithPrefix("TEST_VAR")
	if len(prefixed) != 3 {
		t.Errorf("GetWithPrefix() returned %d variables, want 3", len(prefixed))
	}

	// Test GetAll
	all := env.GetAll()
	if len(all) == 0 {
		t.Error("GetAll() should return environment variables")
	}

	// Test Required
	if err := env.Required("TEST_VAR1", "TEST_VAR2"); err != nil {
		t.Errorf("Required() failed for existing variables: %v", err)
	}

	if err := env.Required("NONEXISTENT_VAR"); err == nil {
		t.Error("Required() should fail for non-existent variable")
	}
}

func TestDefaultPathResolver(t *testing.T) {
	resolver := NewDefaultPathResolver()

	// Test Home
	home := resolver.Home()
	if home == "" {
		t.Error("Home() should return a valid path")
	}

	// Test standard directories
	appName := "test-app"

	configDir := resolver.ConfigDir(appName)
	if configDir == "" || !strings.Contains(configDir, appName) {
		t.Errorf("ConfigDir() = %v, should contain app name", configDir)
	}

	dataDir := resolver.DataDir(appName)
	if dataDir == "" || !strings.Contains(dataDir, appName) {
		t.Errorf("DataDir() = %v, should contain app name", dataDir)
	}

	cacheDir := resolver.CacheDir(appName)
	if cacheDir == "" || !strings.Contains(cacheDir, appName) {
		t.Errorf("CacheDir() = %v, should contain app name", cacheDir)
	}

	// Test temp directory
	tempDir := resolver.TempDir()
	if tempDir == "" {
		t.Error("TempDir() should return a valid path")
	}

	// Test working directory
	workingDir := resolver.WorkingDir()
	if workingDir == "" {
		t.Error("WorkingDir() should return a valid path")
	}
}

func TestDefaultPathResolver_Go(t *testing.T) {
	resolver := NewDefaultPathResolver()

	// Test GOPATH
	gopath := resolver.GOPATH()
	if gopath == "" {
		t.Error("GOPATH() should return a valid path")
	}

	// Test GOROOT
	goroot := resolver.GOROOT()
	if goroot == "" {
		t.Error("GOROOT() should return a valid path")
	}

	// Test GOCACHE
	gocache := resolver.GOCACHE()
	if gocache == "" {
		t.Error("GOCACHE() should return a valid path")
	}

	// Test GOMODCACHE
	gomodcache := resolver.GOMODCACHE()
	if gomodcache == "" {
		t.Error("GOMODCACHE() should return a valid path")
	}
}

func TestDefaultPathResolver_Operations(t *testing.T) {
	resolver := NewDefaultPathResolver()

	// Test path expansion
	testPath := "~/test"
	expanded := resolver.Expand(testPath)
	if runtime.GOOS != "windows" && strings.HasPrefix(expanded, "~/") {
		t.Error("Expand() should resolve tilde")
	}

	// Test absolute path checking
	if runtime.GOOS == "windows" {
		if !resolver.IsAbsolute("C:\\test") {
			t.Error("IsAbsolute() should return true for Windows absolute path")
		}
	} else {
		if !resolver.IsAbsolute("/test") {
			t.Error("IsAbsolute() should return true for Unix absolute path")
		}
	}

	// Test path cleaning
	dirty := filepath.Join("test", "..", "clean")
	clean := resolver.Clean(dirty)
	if clean != "clean" {
		t.Errorf("Clean() = %v, want 'clean'", clean)
	}
}

func TestDefaultEnvManager(t *testing.T) {
	manager := NewDefaultEnvManager()

	// Test scope management
	scope := manager.PushScope()
	if scope == nil {
		t.Fatal("PushScope() should return a valid scope")
	}

	// Test setting in scope
	testKey := "SCOPE_TEST_VAR"
	testValue := "scope_value"

	defer func() {
		// Clean up
		if err := scope.Unset(testKey); err != nil {
			t.Errorf("Failed to unset %s in scope: %v", testKey, err)
		}
		if err := manager.PopScope(); err != nil {
			t.Errorf("Failed to pop scope: %v", err)
		}
	}()

	if err := scope.Set(testKey, testValue); err != nil {
		t.Fatalf("Failed to set variable in scope: %v", err)
	}

	if !scope.HasChanges() {
		t.Error("Scope should have changes after setting variable")
	}

	changes := scope.Changes()
	if len(changes) != 1 {
		t.Errorf("Expected 1 change, got %d", len(changes))
	}

	// Test rollback
	if err := scope.Rollback(); err != nil {
		t.Fatalf("Failed to rollback scope: %v", err)
	}

	if scope.HasChanges() {
		t.Error("Scope should not have changes after rollback")
	}
}

func TestDefaultEnvManager_Context(t *testing.T) {
	manager := NewDefaultEnvManager()

	// Set up test environment
	testKey := "CONTEXT_TEST_VAR"
	testValue := "context_value"

	defer func() {
		if err := manager.RestoreContext(&DefaultEnvContext{variables: make(map[string]string)}); err != nil {
			t.Errorf("Failed to restore context: %v", err)
		}
	}()

	require.NoError(t, manager.baseEnv.Set(testKey, testValue))

	// Save context
	ctx, err := manager.SaveContext()
	if err != nil {
		t.Fatalf("Failed to save context: %v", err)
	}

	if ctx.Count() == 0 {
		t.Error("Saved context should contain variables")
	}

	// Modify environment
	require.NoError(t, manager.baseEnv.Set(testKey, "modified_value"))

	// Restore context
	if err := manager.RestoreContext(ctx); err != nil {
		t.Fatalf("Failed to restore context: %v", err)
	}

	// Verify restoration
	if got := manager.baseEnv.Get(testKey); got != testValue {
		t.Errorf("After restore, got %v, want %v", got, testValue)
	}
}

func TestDefaultEnvManager_Isolation(t *testing.T) {
	manager := NewDefaultEnvManager()

	testKey := "ISOLATION_TEST_VAR"
	originalValue := "original"
	isolatedValue := "isolated"

	// Set original value
	require.NoError(t, manager.baseEnv.Set(testKey, originalValue))

	defer func() {
		if err := manager.baseEnv.Unset(testKey); err != nil {
			t.Errorf("Failed to unset %s: %v", testKey, err)
		}
	}()

	// Test isolation
	err := manager.Isolate(map[string]string{
		testKey: isolatedValue,
	}, func() error {
		// Inside isolation, should see isolated value
		if got := manager.baseEnv.Get(testKey); got != isolatedValue {
			return fmt.Errorf("expected %v, got %v", isolatedValue, got)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Isolation failed: %v", err)
	}

	// After isolation, should see original value
	if got := manager.baseEnv.Get(testKey); got != originalValue {
		t.Errorf("After isolation, expected %v, got %v", originalValue, got)
	}
}

func TestDefaultEnvValidator(t *testing.T) {
	validator := NewDefaultEnvValidator()

	// Test required validation
	validator.Required("REQUIRED_VAR")

	// Should fail for empty value
	if err := validator.Validate("REQUIRED_VAR", ""); err == nil {
		t.Error("Required validation should fail for empty value")
	}

	// Should pass for non-empty value
	if err := validator.Validate("REQUIRED_VAR", "value"); err != nil {
		t.Errorf("Required validation should pass for non-empty value: %v", err)
	}

	// Test pattern validation
	validator.Pattern("EMAIL_VAR", `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	if err := validator.Validate("EMAIL_VAR", "invalid-email"); err == nil {
		t.Error("Pattern validation should fail for invalid email")
	}

	if err := validator.Validate("EMAIL_VAR", "test@example.com"); err != nil {
		t.Errorf("Pattern validation should pass for valid email: %v", err)
	}

	// Test range validation
	validator.Range("PORT_VAR", 1, 65535)

	if err := validator.Validate("PORT_VAR", "0"); err == nil {
		t.Error("Range validation should fail for value below minimum")
	}

	if err := validator.Validate("PORT_VAR", "8080"); err != nil {
		t.Errorf("Range validation should pass for valid port: %v", err)
	}

	// Test one-of validation
	validator.OneOf("ENV_VAR", "dev", "staging", "prod")

	if err := validator.Validate("ENV_VAR", "invalid"); err == nil {
		t.Error("OneOf validation should fail for invalid value")
	}

	if err := validator.Validate("ENV_VAR", "prod"); err != nil {
		t.Errorf("OneOf validation should pass for valid value: %v", err)
	}
}

func TestEnvContext_Operations(t *testing.T) {
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
			"VAR2": "modified2",
			"VAR3": "value3",
		},
	}

	// Test diff
	diff := ctx1.Diff(ctx2)
	if len(diff) == 0 {
		t.Error("Diff should detect changes between contexts")
	}

	// Test merge
	merged := ctx1.Merge(ctx2)
	mergedVars := merged.Variables()

	if mergedVars["VAR1"] != "value1" {
		t.Error("Merge should preserve VAR1 from ctx1")
	}

	if mergedVars["VAR2"] != "modified2" {
		t.Error("Merge should use VAR2 from ctx2")
	}

	if mergedVars["VAR3"] != "value3" {
		t.Error("Merge should include VAR3 from ctx2")
	}
}

func TestEnvOptions(t *testing.T) {
	// Test with auto-trim enabled
	options := EnvOptions{
		AutoTrim:       true,
		AllowOverwrite: true,
	}

	env := NewDefaultEnvironmentWithOptions(options)

	testKey := "TRIM_TEST_VAR"
	testValue := "  trimmed  "

	defer func() {
		if err := env.Unset(testKey); err != nil {
			t.Errorf("Failed to unset %s: %v", testKey, err)
		}
	}()

	if err := env.Set(testKey, testValue); err != nil {
		t.Fatalf("Failed to set variable: %v", err)
	}

	if got := env.Get(testKey); got != "trimmed" {
		t.Errorf("Expected trimmed value 'trimmed', got '%v'", got)
	}
}

// Helper functions

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Benchmark tests

func BenchmarkEnvironment_Get(b *testing.B) {
	env := NewDefaultEnvironment()
	if err := env.Set("BENCH_VAR", "benchmark_value"); err != nil {
		b.Fatalf("Failed to set BENCH_VAR: %v", err)
	}
	defer func() {
		if err := env.Unset("BENCH_VAR"); err != nil {
			b.Logf("Failed to unset BENCH_VAR: %v", err)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.Get("BENCH_VAR")
	}
}

func BenchmarkEnvironment_Set(b *testing.B) {
	env := NewDefaultEnvironment()
	defer func() {
		if err := env.Unset("BENCH_SET_VAR"); err != nil {
			b.Logf("Failed to unset BENCH_SET_VAR: %v", err)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := env.Set("BENCH_SET_VAR", fmt.Sprintf("value_%d", i)); err != nil {
			b.Fatalf("Failed to set BENCH_SET_VAR: %v", err)
		}
	}
}

func BenchmarkPathResolver_ConfigDir(b *testing.B) {
	resolver := NewDefaultPathResolver()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resolver.ConfigDir("test-app")
	}
}

func BenchmarkEnvManager_WithScope(b *testing.B) {
	manager := NewDefaultEnvManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := manager.WithScope(func(scope EnvScope) error {
			if err := scope.Set("BENCH_SCOPE_VAR", "value"); err != nil {
				return err
			}
			return nil
		}); err != nil {
			b.Fatalf("WithScope failed: %v", err)
		}
	}
}
