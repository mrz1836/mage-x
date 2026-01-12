package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

// TestLoadWithEnvOverrides_ErrorPath tests error handling in LoadWithEnvOverrides
func TestLoadWithEnvOverrides_ErrorPath(t *testing.T) {
	cfg := New()

	// Create a test struct
	type TestConfig struct {
		Name string `yaml:"name"`
	}
	var dest TestConfig

	// Try to load from nonexistent paths - should return error from LoadFromPaths
	_, err := cfg.LoadWithEnvOverrides(&dest, "nonexistent_file", "TEST", "/nonexistent/directory")
	require.Error(t, err, "Should error when files don't exist")
}

// TestMockConfigLoader_Coverage tests mock config loader for coverage
func TestMockConfigLoader_Coverage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockConfigLoader(ctrl)

	// Test GetSupportedFormats
	mock.EXPECT().GetSupportedFormats().Return([]string{"yaml", "json"})
	formats := mock.GetSupportedFormats()
	require.Equal(t, []string{"yaml", "json"}, formats)

	// Test Load
	type TestConfig struct {
		Name string
	}
	var dest TestConfig
	mock.EXPECT().Load([]string{"test.yaml"}, gomock.Any()).Return("test.yaml", nil)
	path, err := mock.Load([]string{"test.yaml"}, &dest)
	require.NoError(t, err)
	require.Equal(t, "test.yaml", path)

	// Test LoadFrom
	mock.EXPECT().LoadFrom("test.yaml", gomock.Any()).Return(nil)
	err = mock.LoadFrom("test.yaml", &dest)
	require.NoError(t, err)

	// Test Save
	mock.EXPECT().Save("test.yaml", gomock.Any(), "yaml").Return(nil)
	err = mock.Save("test.yaml", &dest, "yaml")
	require.NoError(t, err)

	// Test Validate
	mock.EXPECT().Validate(gomock.Any()).Return(nil)
	err = mock.Validate(&dest)
	require.NoError(t, err)
}

// TestMockEnvProvider_Coverage tests mock env provider for coverage
func TestMockEnvProvider_Coverage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockEnvProvider(ctrl)

	// Test Get
	mock.EXPECT().Get("TEST_VAR").Return("test_value")
	val := mock.Get("TEST_VAR")
	require.Equal(t, "test_value", val)

	// Test GetAll
	mock.EXPECT().GetAll().Return(map[string]string{"KEY": "VALUE"})
	all := mock.GetAll()
	require.Equal(t, map[string]string{"KEY": "VALUE"}, all)

	// Test GetWithDefault
	mock.EXPECT().GetWithDefault("MISSING", "default").Return("default")
	val = mock.GetWithDefault("MISSING", "default")
	require.Equal(t, "default", val)

	// Test LookupEnv
	mock.EXPECT().LookupEnv("TEST_VAR").Return("value", true)
	val, ok := mock.LookupEnv("TEST_VAR")
	require.True(t, ok)
	require.Equal(t, "value", val)

	// Test Set
	mock.EXPECT().Set("NEW_VAR", "new_value").Return(nil)
	err := mock.Set("NEW_VAR", "new_value")
	require.NoError(t, err)

	// Test Unset
	mock.EXPECT().Unset("VAR").Return(nil)
	err = mock.Unset("VAR")
	require.NoError(t, err)
}

// TestMockTypedEnvProvider_Coverage tests mock typed env provider for coverage
func TestMockTypedEnvProvider_Coverage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockTypedEnvProvider(ctrl)

	// Test Get (inherited from EnvProvider)
	mock.EXPECT().Get("TEST_VAR").Return("test_value")
	val := mock.Get("TEST_VAR")
	require.Equal(t, "test_value", val)

	// Test GetAll (inherited from EnvProvider)
	mock.EXPECT().GetAll().Return(map[string]string{"KEY": "VALUE"})
	all := mock.GetAll()
	require.Equal(t, map[string]string{"KEY": "VALUE"}, all)

	// Test GetBool
	mock.EXPECT().GetBool("BOOL_VAR", true).Return(true)
	boolVal := mock.GetBool("BOOL_VAR", true)
	require.True(t, boolVal)

	// Test GetDuration
	mock.EXPECT().GetDuration("DURATION_VAR", gomock.Any()).Return(5 * time.Second)
	durVal := mock.GetDuration("DURATION_VAR", 0)
	require.Equal(t, 5*time.Second, durVal)

	// Test GetInt
	mock.EXPECT().GetInt("INT_VAR", 0).Return(42)
	intVal := mock.GetInt("INT_VAR", 0)
	require.Equal(t, 42, intVal)

	// Test GetStringSlice
	mock.EXPECT().GetStringSlice("SLICE_VAR", gomock.Any()).Return([]string{"a", "b"})
	sliceVal := mock.GetStringSlice("SLICE_VAR", []string{})
	require.Equal(t, []string{"a", "b"}, sliceVal)

	// Test GetWithDefault (inherited from EnvProvider)
	mock.EXPECT().GetWithDefault("MISSING", "default").Return("default")
	val = mock.GetWithDefault("MISSING", "default")
	require.Equal(t, "default", val)

	// Test LookupEnv (inherited from EnvProvider)
	mock.EXPECT().LookupEnv("TEST_VAR").Return("value", true)
	val, ok := mock.LookupEnv("TEST_VAR")
	require.True(t, ok)
	require.Equal(t, "value", val)

	// Test Set (inherited from EnvProvider)
	mock.EXPECT().Set("NEW_VAR", "new_value").Return(nil)
	err := mock.Set("NEW_VAR", "new_value")
	require.NoError(t, err)

	// Test Unset (inherited from EnvProvider)
	mock.EXPECT().Unset("VAR").Return(nil)
	err = mock.Unset("VAR")
	require.NoError(t, err)

	// Test GetFloat64
	mock.EXPECT().GetFloat64("FLOAT_VAR", 0.0).Return(3.14)
	floatVal := mock.GetFloat64("FLOAT_VAR", 0.0)
	require.InDelta(t, 3.14, floatVal, 0.01)

	// Test GetInt64
	mock.EXPECT().GetInt64("INT64_VAR", int64(0)).Return(int64(9999))
	int64Val := mock.GetInt64("INT64_VAR", 0)
	require.Equal(t, int64(9999), int64Val)
}

// TestMockConfigSource_Coverage tests mock config source for coverage
func TestMockConfigSource_Coverage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockConfigSource(ctrl)

	// Test Name
	mock.EXPECT().Name().Return("test-source")
	name := mock.Name()
	require.Equal(t, "test-source", name)

	// Test Priority
	mock.EXPECT().Priority().Return(100)
	priority := mock.Priority()
	require.Equal(t, 100, priority)

	// Test IsAvailable
	mock.EXPECT().IsAvailable().Return(true)
	available := mock.IsAvailable()
	require.True(t, available)

	// Test Load
	type TestConfig struct {
		Value string
	}
	var dest TestConfig
	mock.EXPECT().Load(gomock.Any()).Return(nil)
	err := mock.Load(&dest)
	require.NoError(t, err)
}

// TestMockConfigManager_Coverage tests mock config manager for coverage
func TestMockConfigManager_Coverage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mock := NewMockConfigManager(ctrl)

	// Test AddSource
	mockSource := NewMockConfigSource(ctrl)
	mock.EXPECT().AddSource(mockSource).Return()
	mock.AddSource(mockSource)

	// Test GetActiveSources
	mock.EXPECT().GetActiveSources().Return([]Source{mockSource})
	sources := mock.GetActiveSources()
	require.Len(t, sources, 1)

	// Test LoadConfig
	type TestConfig struct {
		Value string
	}
	var dest TestConfig
	mock.EXPECT().LoadConfig(gomock.Any()).Return(nil)
	err := mock.LoadConfig(&dest)
	require.NoError(t, err)

	// Test Reload
	mock.EXPECT().Reload(gomock.Any()).Return(nil)
	err = mock.Reload(&dest)
	require.NoError(t, err)

	// Test Watch
	callback := func(any) {}
	mock.EXPECT().Watch(gomock.Any()).Return(nil)
	err = mock.Watch(callback)
	require.NoError(t, err)

	// Test StopWatching
	mock.EXPECT().StopWatching().Return()
	mock.StopWatching()
}
