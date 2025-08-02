package env

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

const (
	testOriginalValue = "original"
)

type EnvFunctionsTestSuite struct {
	suite.Suite

	originalEnv      Environment
	originalPaths    PathResolver
	originalManager  Manager
	originalValidate Validator
}

func (suite *EnvFunctionsTestSuite) SetupTest() {
	// Backup original instances
	suite.originalEnv = DefaultEnv
	suite.originalPaths = DefaultPaths
	suite.originalManager = DefaultManager
	suite.originalValidate = DefaultValidate
}

func (suite *EnvFunctionsTestSuite) TearDownTest() {
	// Restore original instances
	DefaultEnv = suite.originalEnv
	DefaultPaths = suite.originalPaths
	DefaultManager = suite.originalManager
	DefaultValidate = suite.originalValidate

	// Clean up any test environment variables
	testKeys := []string{
		"TEST_ENV_GET", "TEST_ENV_SET", "TEST_ENV_EXISTS", "TEST_ENV_UNSET",
		"TEST_ENV_STRING", "TEST_ENV_BOOL", "TEST_ENV_INT", "TEST_ENV_INT64",
		"TEST_ENV_FLOAT64", "TEST_ENV_DURATION", "TEST_ENV_SLICE",
		"TEST_PREFIX_VAR1", "TEST_PREFIX_VAR2", "TEST_PREFIX_VAR3",
		"TEST_REQUIRED_VAR1", "TEST_REQUIRED_VAR2",
	}

	for _, key := range testKeys {
		//nolint:errcheck // Best-effort cleanup in test teardown
		os.Unsetenv(key)
	}
}

func TestEnvFunctionsTestSuite(t *testing.T) {
	suite.Run(t, new(EnvFunctionsTestSuite))
}

// Test global environment convenience functions

func (suite *EnvFunctionsTestSuite) TestGet() {
	// Test with existing environment variable
	key := "TEST_ENV_GET"
	value := "test_value"
	suite.Require().NoError(os.Setenv(key, value))

	result := Get(key)
	suite.Equal(value, result)

	// Test with non-existing environment variable
	result = Get("NON_EXISTING_VAR")
	suite.Empty(result)
}

func (suite *EnvFunctionsTestSuite) TestSet() {
	key := "TEST_ENV_SET"
	value := "test_set_value"

	err := Set(key, value)
	suite.Require().NoError(err)

	// Verify it was set
	actual := os.Getenv(key)
	suite.Equal(value, actual)
}

func (suite *EnvFunctionsTestSuite) TestUnset() {
	key := "TEST_ENV_UNSET"
	value := "test_unset_value"

	// Set the variable first
	suite.Require().NoError(os.Setenv(key, value))
	suite.True(Exists(key))

	// Unset it
	err := Unset(key)
	suite.Require().NoError(err)

	// Verify it was unset
	suite.False(Exists(key))
}

func (suite *EnvFunctionsTestSuite) TestExists() {
	key := "TEST_ENV_EXISTS"

	// Test non-existing variable
	suite.False(Exists(key))

	// Set the variable
	suite.Require().NoError(os.Setenv(key, "value"))

	// Test existing variable
	suite.True(Exists(key))
}

func (suite *EnvFunctionsTestSuite) TestGetString() {
	key := "TEST_ENV_STRING"
	defaultValue := "default_string"
	testValue := "test_string"

	// Test with default value (variable not set)
	result := GetString(key, defaultValue)
	suite.Equal(defaultValue, result)

	// Test with actual value
	suite.Require().NoError(os.Setenv(key, testValue))
	result = GetString(key, defaultValue)
	suite.Equal(testValue, result)
}

func (suite *EnvFunctionsTestSuite) TestGetBool() {
	key := "TEST_ENV_BOOL"
	defaultValue := false

	// Test with default value (variable not set)
	result := GetBool(key, defaultValue)
	suite.Equal(defaultValue, result)

	// Test with true value
	suite.Require().NoError(os.Setenv(key, "true"))
	result = GetBool(key, defaultValue)
	suite.True(result)

	// Test with false value
	suite.Require().NoError(os.Setenv(key, "false"))
	result = GetBool(key, defaultValue)
	suite.False(result)

	// Test with "1" value
	suite.Require().NoError(os.Setenv(key, "1"))
	result = GetBool(key, defaultValue)
	suite.True(result)

	// Test with "0" value
	suite.Require().NoError(os.Setenv(key, "0"))
	result = GetBool(key, defaultValue)
	suite.False(result)

	// Test with invalid value (should return default)
	suite.Require().NoError(os.Setenv(key, "invalid"))
	result = GetBool(key, true)
	suite.True(result) // Should return default value
}

func (suite *EnvFunctionsTestSuite) TestGetInt() {
	key := "TEST_ENV_INT"
	defaultValue := 42
	testValue := 123

	// Test with default value (variable not set)
	result := GetInt(key, defaultValue)
	suite.Equal(defaultValue, result)

	// Test with actual value
	suite.Require().NoError(os.Setenv(key, "123"))
	result = GetInt(key, defaultValue)
	suite.Equal(testValue, result)

	// Test with invalid value (should return default)
	suite.Require().NoError(os.Setenv(key, "invalid"))
	result = GetInt(key, defaultValue)
	suite.Equal(defaultValue, result)

	// Test with negative value
	suite.Require().NoError(os.Setenv(key, "-456"))
	result = GetInt(key, defaultValue)
	suite.Equal(-456, result)
}

func (suite *EnvFunctionsTestSuite) TestGetInt64() {
	key := "TEST_ENV_INT64"
	defaultValue := int64(42)
	testValue := int64(123456789)

	// Test with default value (variable not set)
	result := GetInt64(key, defaultValue)
	suite.Equal(defaultValue, result)

	// Test with actual value
	suite.Require().NoError(os.Setenv(key, "123456789"))
	result = GetInt64(key, defaultValue)
	suite.Equal(testValue, result)

	// Test with invalid value (should return default)
	suite.Require().NoError(os.Setenv(key, "invalid"))
	result = GetInt64(key, defaultValue)
	suite.Equal(defaultValue, result)
}

func (suite *EnvFunctionsTestSuite) TestGetFloat64() {
	key := "TEST_ENV_FLOAT64"
	defaultValue := 3.14
	testValue := 2.71828

	// Test with default value (variable not set)
	result := GetFloat64(key, defaultValue)
	suite.InDelta(defaultValue, result, 0.00001)

	// Test with actual value
	suite.Require().NoError(os.Setenv(key, "2.71828"))
	result = GetFloat64(key, defaultValue)
	suite.InDelta(testValue, result, 0.00001)

	// Test with invalid value (should return default)
	suite.Require().NoError(os.Setenv(key, "invalid"))
	result = GetFloat64(key, defaultValue)
	suite.InDelta(defaultValue, result, 0.00001)
}

func (suite *EnvFunctionsTestSuite) TestGetDuration() {
	key := "TEST_ENV_DURATION"
	defaultValue := 30 * time.Second
	testValue := 5 * time.Minute

	// Test with default value (variable not set)
	result := GetDuration(key, defaultValue)
	suite.Equal(defaultValue, result)

	// Test with actual value
	suite.Require().NoError(os.Setenv(key, "5m"))
	result = GetDuration(key, defaultValue)
	suite.Equal(testValue, result)

	// Test with seconds
	suite.Require().NoError(os.Setenv(key, "45s"))
	result = GetDuration(key, defaultValue)
	suite.Equal(45*time.Second, result)

	// Test with invalid value (should return default)
	suite.Require().NoError(os.Setenv(key, "invalid"))
	result = GetDuration(key, defaultValue)
	suite.Equal(defaultValue, result)
}

func (suite *EnvFunctionsTestSuite) TestGetStringSlice() {
	key := "TEST_ENV_SLICE"
	defaultValue := []string{"default1", "default2"}
	testValue := []string{"value1", "value2", "value3"}

	// Test with default value (variable not set)
	result := GetStringSlice(key, defaultValue)
	suite.Equal(defaultValue, result)

	// Test with actual value
	suite.Require().NoError(os.Setenv(key, "value1,value2,value3"))
	result = GetStringSlice(key, defaultValue)
	suite.Equal(testValue, result)

	// Test with single value
	suite.Require().NoError(os.Setenv(key, "single"))
	result = GetStringSlice(key, defaultValue)
	suite.Equal([]string{"single"}, result)

	// Test with empty value
	suite.Require().NoError(os.Setenv(key, ""))
	result = GetStringSlice(key, defaultValue)
	suite.Equal([]string{""}, result)
}

func (suite *EnvFunctionsTestSuite) TestGetWithPrefix() {
	// Set up test variables
	suite.Require().NoError(os.Setenv("TEST_PREFIX_VAR1", "value1"))
	suite.Require().NoError(os.Setenv("TEST_PREFIX_VAR2", "value2"))
	suite.Require().NoError(os.Setenv("OTHER_VAR", "other"))

	result := GetWithPrefix("TEST_PREFIX_")

	expected := map[string]string{
		"TEST_PREFIX_VAR1": "value1",
		"TEST_PREFIX_VAR2": "value2",
	}

	suite.Equal(expected, result)
	suite.NotContains(result, "OTHER_VAR")

	// Clean up
	suite.Require().NoError(os.Unsetenv("OTHER_VAR"))
}

func (suite *EnvFunctionsTestSuite) TestSetMultiple() {
	vars := map[string]string{
		"TEST_MULTI_VAR1": "value1",
		"TEST_MULTI_VAR2": "value2",
		"TEST_MULTI_VAR3": "value3",
	}

	err := SetMultiple(vars)
	suite.Require().NoError(err)

	// Verify all variables were set
	for key, expectedValue := range vars {
		actualValue := os.Getenv(key)
		suite.Equal(expectedValue, actualValue)

		// Clean up
		suite.Require().NoError(os.Unsetenv(key))
	}
}

func (suite *EnvFunctionsTestSuite) TestRequired() {
	// Test with all required variables set
	suite.Require().NoError(os.Setenv("TEST_REQUIRED_VAR1", "value1"))
	suite.Require().NoError(os.Setenv("TEST_REQUIRED_VAR2", "value2"))

	err := Required("TEST_REQUIRED_VAR1", "TEST_REQUIRED_VAR2")
	suite.Require().NoError(err)

	// Test with missing required variable
	suite.Require().NoError(os.Unsetenv("TEST_REQUIRED_VAR2"))
	err = Required("TEST_REQUIRED_VAR1", "TEST_REQUIRED_VAR2")
	suite.Require().Error(err)
	suite.Contains(err.Error(), "TEST_REQUIRED_VAR2")
}

// Test path convenience functions

func (suite *EnvFunctionsTestSuite) TestHome() {
	result := Home()
	suite.NotEmpty(result)
	suite.True(filepath.IsAbs(result))
}

func (suite *EnvFunctionsTestSuite) TestConfigDir() {
	appName := "testapp"
	result := ConfigDir(appName)
	suite.NotEmpty(result)
	suite.Contains(result, appName)
}

func (suite *EnvFunctionsTestSuite) TestDataDir() {
	appName := "testapp"
	result := DataDir(appName)
	suite.NotEmpty(result)
	suite.Contains(result, appName)
}

func (suite *EnvFunctionsTestSuite) TestCacheDir() {
	appName := "testapp"
	result := CacheDir(appName)
	suite.NotEmpty(result)
	suite.Contains(result, appName)
}

func (suite *EnvFunctionsTestSuite) TestTempDir() {
	result := TempDir()
	suite.NotEmpty(result)
	suite.True(filepath.IsAbs(result))
}

func (suite *EnvFunctionsTestSuite) TestWorkingDir() {
	result := WorkingDir()
	suite.NotEmpty(result)
	suite.True(filepath.IsAbs(result))
}

func (suite *EnvFunctionsTestSuite) TestGOPATH() {
	result := GOPATH()
	suite.NotEmpty(result)
	// GOPATH should either be set explicitly or have a default value
}

func (suite *EnvFunctionsTestSuite) TestGOROOT() {
	result := GOROOT()
	suite.NotEmpty(result)
	suite.True(filepath.IsAbs(result))
}

func (suite *EnvFunctionsTestSuite) TestGOCACHE() {
	result := GOCACHE()
	suite.NotEmpty(result)
	suite.True(filepath.IsAbs(result))
}

func (suite *EnvFunctionsTestSuite) TestGOMODCACHE() {
	result := GOMODCACHE()
	suite.NotEmpty(result)
	suite.True(filepath.IsAbs(result))
}

func (suite *EnvFunctionsTestSuite) TestExpand() {
	// Test with environment variable expansion
	suite.Require().NoError(os.Setenv("TEST_EXPAND_VAR", "expanded"))

	result := Expand("$TEST_EXPAND_VAR/path")
	suite.Contains(result, "expanded")

	// Clean up
	suite.Require().NoError(os.Unsetenv("TEST_EXPAND_VAR"))
}

func (suite *EnvFunctionsTestSuite) TestResolve() {
	// Test with current directory
	result, err := Resolve(".")
	suite.Require().NoError(err)
	suite.True(filepath.IsAbs(result))

	// Test with relative path
	result, err = Resolve("test/path")
	suite.Require().NoError(err)
	suite.True(filepath.IsAbs(result))
}

func (suite *EnvFunctionsTestSuite) TestIsAbsolute() {
	// Test absolute path
	suite.True(IsAbsolute("/absolute/path"))

	// Test relative path
	suite.False(IsAbsolute("relative/path"))

	// Test current directory
	suite.False(IsAbsolute("."))
}

func (suite *EnvFunctionsTestSuite) TestMakeAbsolute() {
	// Test relative path
	result, err := MakeAbsolute("relative/path")
	suite.Require().NoError(err)
	suite.True(filepath.IsAbs(result))

	// Test already absolute path
	absolutePath := "/already/absolute"
	result, err = MakeAbsolute(absolutePath)
	suite.Require().NoError(err)
	suite.Equal(absolutePath, result)
}

func (suite *EnvFunctionsTestSuite) TestClean() {
	// Test path cleaning
	dirtyPath := "path/../to/./clean"
	result := Clean(dirtyPath)
	suite.Equal("to/clean", result)

	// Test already clean path
	cleanPath := "clean/path"
	result = Clean(cleanPath)
	suite.Equal(cleanPath, result)
}

func (suite *EnvFunctionsTestSuite) TestEnsureDir() {
	tempDir := suite.T().TempDir()
	testDir := filepath.Join(tempDir, "test", "nested", "directory")

	err := EnsureDir(testDir)
	suite.Require().NoError(err)

	// Verify directory exists
	stat, err := os.Stat(testDir)
	suite.Require().NoError(err)
	suite.True(stat.IsDir())
}

func (suite *EnvFunctionsTestSuite) TestEnsureDirWithMode() {
	tempDir := suite.T().TempDir()
	testDir := filepath.Join(tempDir, "test", "mode", "directory")

	err := EnsureDirWithMode(testDir, 0o755)
	suite.Require().NoError(err)

	// Verify directory exists
	stat, err := os.Stat(testDir)
	suite.Require().NoError(err)
	suite.True(stat.IsDir())
}

// Test manager convenience functions

func (suite *EnvFunctionsTestSuite) TestWithScope() {
	testVar := "TEST_SCOPE_VAR"
	originalValue := testOriginalValue
	scopeValue := "scoped"

	// Set original value
	suite.Require().NoError(os.Setenv(testVar, originalValue))

	err := WithScope(func(scope Scope) error {
		// Set value within scope
		return scope.Set(testVar, scopeValue)
	})
	suite.Require().NoError(err)

	// Verify original value is restored
	actualValue := os.Getenv(testVar)
	suite.Equal(originalValue, actualValue)

	// Clean up
	suite.Require().NoError(os.Unsetenv(testVar))
}

func (suite *EnvFunctionsTestSuite) TestSaveAndRestoreContext() {
	testVar := "TEST_CONTEXT_VAR"
	originalValue := testOriginalValue
	newValue := "modified"

	// Set original value
	suite.Require().NoError(os.Setenv(testVar, originalValue))

	// Save context
	ctx, err := SaveContext()
	suite.Require().NoError(err)
	suite.NotNil(ctx)

	// Modify value
	suite.Require().NoError(os.Setenv(testVar, newValue))
	suite.Equal(newValue, os.Getenv(testVar))

	// Restore context
	err = RestoreContext(ctx)
	suite.Require().NoError(err)

	// Verify original value is restored
	actualValue := os.Getenv(testVar)
	suite.Equal(originalValue, actualValue)

	// Clean up
	suite.Require().NoError(os.Unsetenv(testVar))
}

func (suite *EnvFunctionsTestSuite) TestIsolate() {
	testVar := "TEST_ISOLATE_VAR"
	originalValue := testOriginalValue
	isolatedValue := "isolated"

	// Set original value
	suite.Require().NoError(os.Setenv(testVar, originalValue))

	vars := map[string]string{
		testVar: isolatedValue,
	}

	var valueInIsolation string
	err := Isolate(vars, func() error {
		valueInIsolation = os.Getenv(testVar)
		return nil
	})
	suite.Require().NoError(err)

	// Verify isolated value was set within isolation
	suite.Equal(isolatedValue, valueInIsolation)

	// Verify original value is restored
	actualValue := os.Getenv(testVar)
	suite.Equal(originalValue, actualValue)

	// Clean up
	suite.Require().NoError(os.Unsetenv(testVar))
}

// Test configuration functions

func (suite *EnvFunctionsTestSuite) TestSetEnvironment() {
	mockEnv := NewMockEnvironment()

	SetEnvironment(mockEnv)
	suite.Equal(mockEnv, DefaultEnv)

	// Reset to original
	DefaultEnv = suite.originalEnv
}

func (suite *EnvFunctionsTestSuite) TestSetPathResolver() {
	mockPaths := NewMockPathResolver()

	SetPathResolver(mockPaths)
	suite.Equal(mockPaths, DefaultPaths)

	// Reset to original
	DefaultPaths = suite.originalPaths
}

func (suite *EnvFunctionsTestSuite) TestSetManager() {
	mockManager := NewMockEnvManager()

	SetManager(mockManager)
	suite.Equal(mockManager, DefaultManager)

	// Reset to original
	DefaultManager = suite.originalManager
}

func (suite *EnvFunctionsTestSuite) TestSetValidator() {
	mockValidator := NewMockEnvValidator()

	SetValidator(mockValidator)
	suite.Equal(mockValidator, DefaultValidate)

	// Reset to original
	DefaultValidate = suite.originalValidate
}

// Test advanced utility functions

func (suite *EnvFunctionsTestSuite) TestLoadFromFile() {
	// This is a placeholder implementation - should not error
	err := LoadFromFile("nonexistent.env")
	suite.Require().NoError(err)
}

func (suite *EnvFunctionsTestSuite) TestSaveToFile() {
	// This is a placeholder implementation - should not error
	err := SaveToFile("output.env", "TEST_")
	suite.Require().NoError(err)
}

func (suite *EnvFunctionsTestSuite) TestBackupAndRestore() {
	testVar := "TEST_BACKUP_VAR"
	originalValue := testOriginalValue
	newValue := "modified"

	// Set original value
	suite.Require().NoError(os.Setenv(testVar, originalValue))

	// Create backup
	backup := Backup()
	suite.NotNil(backup)

	// Modify value
	suite.Require().NoError(os.Setenv(testVar, newValue))
	suite.Equal(newValue, os.Getenv(testVar))

	// Restore from backup
	err := Restore(backup)
	suite.Require().NoError(err)

	// Verify original value is restored
	actualValue := os.Getenv(testVar)
	suite.Equal(originalValue, actualValue)

	// Clean up
	suite.Require().NoError(os.Unsetenv(testVar))
}

func (suite *EnvFunctionsTestSuite) TestGetAllWithPrefix() {
	// This should be equivalent to GetWithPrefix
	suite.Require().NoError(os.Setenv("TEST_PREFIX_VAR1", "value1"))
	suite.Require().NoError(os.Setenv("TEST_PREFIX_VAR2", "value2"))

	result1 := GetAllWithPrefix("TEST_PREFIX_")
	result2 := GetWithPrefix("TEST_PREFIX_")

	suite.Equal(result2, result1)
}

func (suite *EnvFunctionsTestSuite) TestSetFromMap() {
	vars := map[string]string{
		"TEST_MAP_VAR1": "value1",
		"TEST_MAP_VAR2": "value2",
	}

	err := SetFromMap(vars)
	suite.Require().NoError(err)

	// Verify variables were set
	for key, expectedValue := range vars {
		actualValue := os.Getenv(key)
		suite.Equal(expectedValue, actualValue)

		// Clean up
		suite.Require().NoError(os.Unsetenv(key))
	}
}

func (suite *EnvFunctionsTestSuite) TestClearPrefix() {
	// Set up test variables
	suite.Require().NoError(os.Setenv("TEST_CLEAR_VAR1", "value1"))
	suite.Require().NoError(os.Setenv("TEST_CLEAR_VAR2", "value2"))
	suite.Require().NoError(os.Setenv("OTHER_VAR", "other"))

	err := ClearPrefix("TEST_CLEAR_")
	suite.Require().NoError(err)

	// Verify prefixed variables were cleared
	suite.Empty(os.Getenv("TEST_CLEAR_VAR1"))
	suite.Empty(os.Getenv("TEST_CLEAR_VAR2"))

	// Verify other variables were not affected
	suite.Equal("other", os.Getenv("OTHER_VAR"))

	// Clean up
	suite.Require().NoError(os.Unsetenv("OTHER_VAR"))
}

// Test constructor functions

func (suite *EnvFunctionsTestSuite) TestNewEnvironment() {
	env := NewEnvironment()
	suite.NotNil(env)
	suite.IsType(&DefaultEnvironment{}, env)
}

func (suite *EnvFunctionsTestSuite) TestNewOSEnvironment() {
	env := NewOSEnvironment()
	suite.NotNil(env)
	suite.IsType(&DefaultEnvironment{}, env)
}
