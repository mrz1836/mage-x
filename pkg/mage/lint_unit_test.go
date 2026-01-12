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

// Tests for parseVersionFromOutput and containsDigit (pre-compiled regex optimization)

func TestParseVersionFromOutput_PreCompiledRegex(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "semantic version with v prefix",
			input:    "golangci-lint has version v1.55.2 built with go1.21.0",
			expected: "v1.55.2",
		},
		{
			name:     "semantic version without v prefix",
			input:    "version 1.55.2",
			expected: "1.55.2",
		},
		{
			name:     "semantic version with prerelease",
			input:    "v1.55.2-beta.1",
			expected: "v1.55.2-beta.1",
		},
		{
			name:     "major.minor only",
			input:    "go version go1.21 darwin/arm64",
			expected: "1.21",
		},
		{
			name:     "version on second word",
			input:    "golangci-lint 1.55.2",
			expected: "1.55.2",
		},
		{
			name:     "version in complex string",
			input:    "Built from: v1.55.2-abc123",
			expected: "v1.55.2-abc123",
		},
		{
			name:     "empty input returns empty string",
			input:    "",
			expected: "", // strings.Split("", "\n") returns [""], which has length 1
		},
		{
			name:     "whitespace only returns empty string",
			input:    "   \n   ",
			expected: "", // trimmed to empty
		},
		{
			name:     "multiline with version on first line",
			input:    "v2.0.0\nsome other info\nmore stuff",
			expected: "v2.0.0",
		},
		{
			name:     "no version pattern fallback to word with digit",
			input:    "tool abc123def version",
			expected: "abc123def",
		},
		{
			name:     "no digits returns first line",
			input:    "unknown",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseVersionFromOutput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsDigit(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "contains single digit",
			input:    "abc1def",
			expected: true,
		},
		{
			name:     "contains multiple digits",
			input:    "v1.2.3",
			expected: true,
		},
		{
			name:     "digit at start",
			input:    "1abc",
			expected: true,
		},
		{
			name:     "digit at end",
			input:    "abc9",
			expected: true,
		},
		{
			name:     "only digits",
			input:    "12345",
			expected: true,
		},
		{
			name:     "no digits",
			input:    "abcdef",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "special characters only",
			input:    "!@#$%",
			expected: false,
		},
		{
			name:     "letters only no digits",
			input:    "abcxyz",
			expected: false,
		},
		{
			name:     "mixed characters with digit",
			input:    "abc1xyz",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsDigit(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark to verify pre-compiled regex performance improvement
func BenchmarkParseVersionFromOutput(b *testing.B) {
	input := "golangci-lint has version v1.55.2 built with go1.21.0"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = parseVersionFromOutput(input)
	}
}

func BenchmarkContainsDigit(b *testing.B) {
	input := "golangci-lint"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = containsDigit(input)
	}
}
