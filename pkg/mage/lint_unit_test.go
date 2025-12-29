package mage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGolangciLintArgs_ResolveConfigPath_ModuleConfig(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	moduleDir := filepath.Join(tmpDir, "module")
	require.NoError(t, os.MkdirAll(moduleDir, 0o750))

	// Create config file in module directory
	configPath := filepath.Join(moduleDir, ".golangci.json")
	require.NoError(t, os.WriteFile(configPath, []byte(`{"run":{}}`), 0o600))

	// Change to temp dir so root config check works correctly
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // test cleanup

	args := &golangciLintArgs{
		modulePath: moduleDir,
		config:     &Config{},
		withFix:    false,
	}

	resolvedPath, configArgs := args.resolveConfigPath()

	assert.Equal(t, configPath, resolvedPath)
	assert.Equal(t, []string{"--config", ".golangci.json"}, configArgs)
}

func TestGolangciLintArgs_ResolveConfigPath_RootConfig(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	moduleDir := filepath.Join(tmpDir, "module")
	require.NoError(t, os.MkdirAll(moduleDir, 0o750))

	// Create config file in root directory only
	rootConfig := filepath.Join(tmpDir, ".golangci.json")
	require.NoError(t, os.WriteFile(rootConfig, []byte(`{"run":{}}`), 0o600))

	// Change to temp dir so root config check works correctly
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // test cleanup

	args := &golangciLintArgs{
		modulePath: moduleDir,
		config:     &Config{},
		withFix:    false,
	}

	resolvedPath, configArgs := args.resolveConfigPath()

	// Should find the root config
	assert.NotEmpty(t, resolvedPath)
	assert.Len(t, configArgs, 2)
	assert.Equal(t, "--config", configArgs[0])
	// The second arg should be the absolute path to the root config
	assert.Contains(t, configArgs[1], ".golangci.json")
}

func TestGolangciLintArgs_ResolveConfigPath_NoConfig(t *testing.T) {
	// Create temp directory structure with no config files
	tmpDir := t.TempDir()
	moduleDir := filepath.Join(tmpDir, "module")
	require.NoError(t, os.MkdirAll(moduleDir, 0o750))

	// Change to temp dir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // test cleanup

	args := &golangciLintArgs{
		modulePath: moduleDir,
		config:     &Config{},
		withFix:    false,
	}

	resolvedPath, configArgs := args.resolveConfigPath()

	assert.Empty(t, resolvedPath)
	assert.Nil(t, configArgs)
}

func TestGolangciLintArgs_TimeoutArgs_FromFile(t *testing.T) {
	// Create temp directory with config file containing timeout
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".golangci.json")
	configContent := `{"run":{"timeout":"5m"}}`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o600))

	args := &golangciLintArgs{
		modulePath: tmpDir,
		config:     &Config{Lint: LintConfig{Timeout: "2m"}}, // mage config has different timeout
		withFix:    false,
	}

	timeoutArgs := args.timeoutArgs(configPath)

	// Should use timeout from config file (5m), not mage config (2m)
	assert.Equal(t, []string{"--timeout", "5m"}, timeoutArgs)
}

func TestGolangciLintArgs_TimeoutArgs_FromMageConfig(t *testing.T) {
	// Create temp directory with config file containing no timeout
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ".golangci.json")
	configContent := `{"run":{}}`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o600))

	args := &golangciLintArgs{
		modulePath: tmpDir,
		config:     &Config{Lint: LintConfig{Timeout: "3m"}},
		withFix:    false,
	}

	timeoutArgs := args.timeoutArgs(configPath)

	// Should use mage config timeout when file timeout is empty
	assert.Equal(t, []string{"--timeout", "3m"}, timeoutArgs)
}

func TestGolangciLintArgs_TimeoutArgs_NoTimeout(t *testing.T) {
	args := &golangciLintArgs{
		modulePath: "/nonexistent",
		config:     &Config{Lint: LintConfig{Timeout: ""}},
		withFix:    false,
	}

	timeoutArgs := args.timeoutArgs("")

	assert.Nil(t, timeoutArgs)
}

func TestGolangciLintArgs_CommonFlags_WithTags(t *testing.T) {
	args := &golangciLintArgs{
		modulePath: "/test",
		config: &Config{
			Build: BuildConfig{
				Tags: []string{"integration", "e2e"},
			},
		},
		withFix: false,
	}

	flags := args.commonFlags()

	assert.Contains(t, flags, "--build-tags")
	// Find the index of --build-tags and check the next element
	for i, flag := range flags {
		if flag == "--build-tags" && i+1 < len(flags) {
			assert.Equal(t, "integration,e2e", flags[i+1])
			break
		}
	}
}

func TestGolangciLintArgs_CommonFlags_NoTags(t *testing.T) {
	args := &golangciLintArgs{
		modulePath: "/test",
		config: &Config{
			Build: BuildConfig{
				Tags: []string{},
			},
		},
		withFix: false,
	}

	flags := args.commonFlags()

	assert.NotContains(t, flags, "--build-tags")
}

func TestGolangciLintArgs_BuildArgs_Default(t *testing.T) {
	// Create temp directory with no config file
	tmpDir := t.TempDir()
	moduleDir := filepath.Join(tmpDir, "module")
	require.NoError(t, os.MkdirAll(moduleDir, 0o750))

	// Change to temp dir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // test cleanup

	args := &golangciLintArgs{
		modulePath: moduleDir,
		config:     &Config{},
		withFix:    false,
	}

	builtArgs := args.buildArgs()

	// Should start with "run" and "./..."
	assert.Equal(t, "run", builtArgs[0])
	assert.Contains(t, builtArgs, "./...")
	// Should NOT contain --fix
	assert.NotContains(t, builtArgs, "--fix")
}

func TestGolangciLintArgs_BuildArgs_WithFix(t *testing.T) {
	// Create temp directory with no config file
	tmpDir := t.TempDir()
	moduleDir := filepath.Join(tmpDir, "module")
	require.NoError(t, os.MkdirAll(moduleDir, 0o750))

	// Change to temp dir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // test cleanup

	args := &golangciLintArgs{
		modulePath: moduleDir,
		config:     &Config{},
		withFix:    true,
	}

	builtArgs := args.buildArgs()

	// Should start with "run" and contain "--fix"
	assert.Equal(t, "run", builtArgs[0])
	assert.Contains(t, builtArgs, "--fix")
	assert.Contains(t, builtArgs, "./...")
}

func TestGolangciLintArgs_BuildArgs_FullIntegration(t *testing.T) {
	// Create temp directory with config file
	tmpDir := t.TempDir()
	moduleDir := filepath.Join(tmpDir, "module")
	require.NoError(t, os.MkdirAll(moduleDir, 0o750))

	// Create config file in module directory
	configPath := filepath.Join(moduleDir, ".golangci.json")
	configContent := `{"run":{"timeout":"10m"}}`
	require.NoError(t, os.WriteFile(configPath, []byte(configContent), 0o600))

	// Change to temp dir
	origDir, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origDir) }() //nolint:errcheck // test cleanup

	args := &golangciLintArgs{
		modulePath: moduleDir,
		config: &Config{
			Build: BuildConfig{
				Tags: []string{"test"},
			},
			Lint: LintConfig{
				Timeout: "5m", // Will be overridden by config file
			},
		},
		withFix: true,
	}

	builtArgs := args.buildArgs()

	// Verify all expected arguments are present
	assert.Equal(t, "run", builtArgs[0])
	assert.Contains(t, builtArgs, "--fix")
	assert.Contains(t, builtArgs, "./...")
	assert.Contains(t, builtArgs, "--config")
	assert.Contains(t, builtArgs, "--timeout")
	assert.Contains(t, builtArgs, "--build-tags")

	// Verify timeout is from config file (10m), not mage config (5m)
	for i, arg := range builtArgs {
		if arg == "--timeout" && i+1 < len(builtArgs) {
			assert.Equal(t, "10m", builtArgs[i+1])
			break
		}
	}
}
