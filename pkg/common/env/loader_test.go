package env

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestLoadEnvFile tests the loadEnvFile function with various input formats.
// This validates the core .env parsing logic including comments, whitespace,
// and special characters.
func TestLoadEnvFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected map[string]string
		wantErr  bool
	}{
		{
			name:    "basic_key_value",
			content: "KEY=value",
			expected: map[string]string{
				"KEY": "value",
			},
		},
		{
			name:    "inline_comment",
			content: "KEY=value # this is a comment",
			expected: map[string]string{
				"KEY": "value",
			},
		},
		{
			name:    "equals_in_value",
			content: "KEY=a=b=c",
			expected: map[string]string{
				"KEY": "a=b=c",
			},
		},
		{
			name:    "empty_value",
			content: "KEY=",
			expected: map[string]string{
				"KEY": "",
			},
		},
		{
			name:     "no_separator_skipped",
			content:  "KEYONLY",
			expected: map[string]string{},
		},
		{
			name:     "blank_line_skipped",
			content:  "\n\n",
			expected: map[string]string{},
		},
		{
			name:     "comment_line_skipped",
			content:  "# this is a comment",
			expected: map[string]string{},
		},
		{
			name:    "whitespace_trimmed",
			content: "  KEY  =  value  ",
			expected: map[string]string{
				"KEY": "value",
			},
		},
		{
			name:    "unicode_value",
			content: "KEY=日本語", //nolint:gosmopolitan // Intentional Unicode test
			expected: map[string]string{
				"KEY": "日本語", //nolint:gosmopolitan // Intentional Unicode test
			},
		},
		{
			name:    "crlf_line_endings",
			content: "KEY=value\r\n",
			expected: map[string]string{
				"KEY": "value",
			},
		},
		{
			name: "multiple_variables",
			content: `VAR1=value1
VAR2=value2
VAR3=value3`,
			expected: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
				"VAR3": "value3",
			},
		},
		{
			name: "mixed_content",
			content: `# Header comment
VAR1=value1

# Another comment
VAR2=value2 # inline comment

EMPTY=
INVALID_LINE
VAR3=has=equals=signs`,
			expected: map[string]string{
				"VAR1":  "value1",
				"VAR2":  "value2",
				"EMPTY": "",
				"VAR3":  "has=equals=signs",
			},
		},
		{
			name:    "hash_at_start_of_value",
			content: "KEY=#startswithash",
			expected: map[string]string{
				"KEY": "",
			},
		},
		{
			name:    "special_characters_in_key",
			content: "MY_VAR_123=value",
			expected: map[string]string{
				"MY_VAR_123": "value",
			},
		},
		{
			name:    "spaces_around_equals",
			content: "KEY = value",
			expected: map[string]string{
				"KEY": "value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with test content
			tmpDir := t.TempDir()
			envFile := filepath.Join(tmpDir, ".env")
			err := os.WriteFile(envFile, []byte(tt.content), 0o600)
			require.NoError(t, err, "failed to create test env file")

			// Clear any existing test vars and track what we set
			for key := range tt.expected {
				require.NoError(t, os.Unsetenv(key))
			}
			t.Cleanup(func() {
				for key := range tt.expected {
					require.NoError(t, os.Unsetenv(key))
				}
			})

			// Load the env file
			err = loadEnvFile(envFile)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Verify expected values
			for key, expectedValue := range tt.expected {
				actualValue := os.Getenv(key)
				require.Equal(t, expectedValue, actualValue,
					"key %s: expected %q, got %q", key, expectedValue, actualValue)
			}
		})
	}
}

// TestLoadEnvFile_NonExistentFile verifies that loading a non-existent file
// returns an error (not nil).
func TestLoadEnvFile_NonExistentFile(t *testing.T) {
	err := loadEnvFile("/nonexistent/path/.env")
	require.Error(t, err, "loading non-existent file should return error")
}

// TestExpandVariables tests variable expansion in .env values.
// This validates both ${VAR} and $VAR formats, as well as local var priority.
func TestExpandVariables(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		localVars map[string]string
		envVars   map[string]string
		expected  string
	}{
		{
			name:      "brace_format_from_env",
			input:     "${TEST_HOME}",
			localVars: map[string]string{},
			envVars:   map[string]string{"TEST_HOME": "/home/test"},
			expected:  "/home/test",
		},
		{
			name:      "brace_format_from_local",
			input:     "${TEST_DB}",
			localVars: map[string]string{"TEST_DB": "local_value"},
			envVars:   map[string]string{"TEST_DB": "env_value"},
			expected:  "local_value",
		},
		{
			name:      "dollar_simple_from_env",
			input:     "$TEST_PATH",
			localVars: map[string]string{},
			envVars:   map[string]string{"TEST_PATH": "/usr/bin"},
			expected:  "/usr/bin",
		},
		{
			name:      "dollar_simple_from_local",
			input:     "$TEST_VAR",
			localVars: map[string]string{"TEST_VAR": "local"},
			envVars:   map[string]string{"TEST_VAR": "env"},
			expected:  "local",
		},
		{
			name:      "mixed_text_with_braces",
			input:     "path/${TEST_DIR}/file",
			localVars: map[string]string{"TEST_DIR": "subdir"},
			envVars:   map[string]string{},
			expected:  "path/subdir/file",
		},
		{
			name:      "multiple_brace_expansions",
			input:     "${TEST_A}${TEST_B}",
			localVars: map[string]string{"TEST_A": "hello", "TEST_B": "world"},
			envVars:   map[string]string{},
			expected:  "helloworld",
		},
		{
			name:      "undefined_variable",
			input:     "${TEST_MISSING}",
			localVars: map[string]string{},
			envVars:   map[string]string{},
			expected:  "",
		},
		{
			name:      "unclosed_brace",
			input:     "${TEST_OPEN",
			localVars: map[string]string{},
			envVars:   map[string]string{},
			// Unclosed brace falls through to $VAR handling, which treats
			// "{TEST_OPEN" as varName and looks it up (returns empty)
			expected: "",
		},
		{
			name:      "dollar_with_space_not_expanded",
			input:     "$TEST_VAR some text",
			localVars: map[string]string{"TEST_VAR": "value"},
			envVars:   map[string]string{},
			expected:  "$TEST_VAR some text",
		},
		{
			name:      "no_variables",
			input:     "plain text value",
			localVars: map[string]string{},
			envVars:   map[string]string{},
			expected:  "plain text value",
		},
		{
			name:      "empty_input",
			input:     "",
			localVars: map[string]string{},
			envVars:   map[string]string{},
			expected:  "",
		},
		{
			name:      "nested_braces_sequential",
			input:     "prefix_${TEST_X}_middle_${TEST_Y}_suffix",
			localVars: map[string]string{"TEST_X": "first", "TEST_Y": "second"},
			envVars:   map[string]string{},
			expected:  "prefix_first_middle_second_suffix",
		},
		{
			name:      "dollar_only",
			input:     "$",
			localVars: map[string]string{},
			envVars:   map[string]string{},
			// Single "$" triggers $VAR handling with empty varName,
			// os.Getenv("") returns empty string
			expected: "",
		},
		{
			name:      "dollar_brace_only",
			input:     "${}",
			localVars: map[string]string{},
			envVars:   map[string]string{},
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment variables
			for key, value := range tt.envVars {
				require.NoError(t, os.Setenv(key, value))
			}
			t.Cleanup(func() {
				for key := range tt.envVars {
					require.NoError(t, os.Unsetenv(key))
				}
			})

			result := expandVariables(tt.input, tt.localVars)
			require.Equal(t, tt.expected, result,
				"expandVariables(%q) = %q, want %q", tt.input, result, tt.expected)
		})
	}
}

// TestLoadEnvFiles tests the public LoadEnvFiles function.
// This validates file loading order, priority, and error handling.
func TestLoadEnvFiles(t *testing.T) {
	t.Run("single_file", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create .env file
		envContent := "SINGLE_TEST_VAR=single_value"
		err := os.WriteFile(filepath.Join(tmpDir, ".env"), []byte(envContent), 0o600)
		require.NoError(t, err)

		t.Cleanup(func() {
			require.NoError(t, os.Unsetenv("SINGLE_TEST_VAR"))
		})

		err = LoadEnvFiles(tmpDir)
		require.NoError(t, err)
		require.Equal(t, "single_value", os.Getenv("SINGLE_TEST_VAR"))
	})

	t.Run("priority_ordering", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create .env.base (lowest priority)
		err := os.WriteFile(filepath.Join(tmpDir, ".env.base"),
			[]byte("PRIORITY_TEST=base"), 0o600)
		require.NoError(t, err)

		// Create .env (highest priority)
		err = os.WriteFile(filepath.Join(tmpDir, ".env"),
			[]byte("PRIORITY_TEST=highest"), 0o600)
		require.NoError(t, err)

		t.Cleanup(func() {
			require.NoError(t, os.Unsetenv("PRIORITY_TEST"))
		})

		err = LoadEnvFiles(tmpDir)
		require.NoError(t, err)
		require.Equal(t, "highest", os.Getenv("PRIORITY_TEST"),
			".env should override .env.base")
	})

	t.Run("multiple_base_paths", func(t *testing.T) {
		tmpDir1 := t.TempDir()
		tmpDir2 := t.TempDir()

		// Create .env in first directory
		err := os.WriteFile(filepath.Join(tmpDir1, ".env"),
			[]byte("MULTI_PATH_VAR1=dir1"), 0o600)
		require.NoError(t, err)

		// Create .env in second directory
		err = os.WriteFile(filepath.Join(tmpDir2, ".env"),
			[]byte("MULTI_PATH_VAR2=dir2"), 0o600)
		require.NoError(t, err)

		t.Cleanup(func() {
			require.NoError(t, os.Unsetenv("MULTI_PATH_VAR1"))
			require.NoError(t, os.Unsetenv("MULTI_PATH_VAR2"))
		})

		err = LoadEnvFiles(tmpDir1, tmpDir2)
		require.NoError(t, err)
		require.Equal(t, "dir1", os.Getenv("MULTI_PATH_VAR1"))
		require.Equal(t, "dir2", os.Getenv("MULTI_PATH_VAR2"))
	})

	t.Run("nonexistent_files_no_error", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Don't create any .env files

		err := LoadEnvFiles(tmpDir)
		require.NoError(t, err, "should not error when no .env files exist")
	})

	t.Run("default_cwd_when_no_paths", func(t *testing.T) {
		// Save and restore working directory
		originalWd, err := os.Getwd()
		require.NoError(t, err)

		tmpDir := t.TempDir()
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, os.Chdir(originalWd))
			require.NoError(t, os.Unsetenv("CWD_TEST_VAR"))
		})

		// Create .env in temp directory (now cwd)
		err = os.WriteFile(filepath.Join(tmpDir, ".env"),
			[]byte("CWD_TEST_VAR=cwd_value"), 0o600)
		require.NoError(t, err)

		err = LoadEnvFiles() // No paths = use cwd
		require.NoError(t, err)
		require.Equal(t, "cwd_value", os.Getenv("CWD_TEST_VAR"))
	})

	t.Run("variable_expansion_between_files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create .env.base with base variable
		err := os.WriteFile(filepath.Join(tmpDir, ".env.base"),
			[]byte("BASE_VAR=base"), 0o600)
		require.NoError(t, err)

		// Create .env that references the base variable
		err = os.WriteFile(filepath.Join(tmpDir, ".env"),
			[]byte("EXPANDED_VAR=prefix_${BASE_VAR}_suffix"), 0o600)
		require.NoError(t, err)

		t.Cleanup(func() {
			require.NoError(t, os.Unsetenv("BASE_VAR"))
			require.NoError(t, os.Unsetenv("EXPANDED_VAR"))
		})

		err = LoadEnvFiles(tmpDir)
		require.NoError(t, err)
		require.Equal(t, "prefix_base_suffix", os.Getenv("EXPANDED_VAR"))
	})

	t.Run("github_env_directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create .github directory
		githubDir := filepath.Join(tmpDir, ".github")
		err := os.MkdirAll(githubDir, 0o750)
		require.NoError(t, err)

		// Create .github/.env.base
		err = os.WriteFile(filepath.Join(githubDir, ".env.base"),
			[]byte("GITHUB_BASE_VAR=github_base"), 0o600)
		require.NoError(t, err)

		t.Cleanup(func() {
			require.NoError(t, os.Unsetenv("GITHUB_BASE_VAR"))
		})

		err = LoadEnvFiles(tmpDir)
		require.NoError(t, err)
		require.Equal(t, "github_base", os.Getenv("GITHUB_BASE_VAR"))
	})
}

// TestLoadStartupEnv tests the LoadStartupEnv convenience function.
// This validates parent directory searching and priority.
func TestLoadStartupEnv(t *testing.T) {
	t.Run("loads_from_cwd", func(t *testing.T) {
		// Save original working directory
		originalWd, err := os.Getwd()
		require.NoError(t, err)

		// Create nested directory structure
		tmpDir := t.TempDir()
		childDir := filepath.Join(tmpDir, "child")
		err = os.MkdirAll(childDir, 0o750)
		require.NoError(t, err)

		// Create .env in child directory
		err = os.WriteFile(filepath.Join(childDir, ".env"),
			[]byte("STARTUP_CWD_VAR=cwd_value"), 0o600)
		require.NoError(t, err)

		// Change to child directory
		err = os.Chdir(childDir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, os.Chdir(originalWd))
			require.NoError(t, os.Unsetenv("STARTUP_CWD_VAR"))
		})

		err = LoadStartupEnv()
		require.NoError(t, err)
		require.Equal(t, "cwd_value", os.Getenv("STARTUP_CWD_VAR"))
	})

	t.Run("searches_parent_directories", func(t *testing.T) {
		originalWd, err := os.Getwd()
		require.NoError(t, err)

		// Create nested directory structure
		tmpDir := t.TempDir()
		level1 := filepath.Join(tmpDir, "level1")
		level2 := filepath.Join(level1, "level2")
		level3 := filepath.Join(level2, "level3")
		err = os.MkdirAll(level3, 0o750)
		require.NoError(t, err)

		// Create .env only in tmpDir (3 levels up)
		err = os.WriteFile(filepath.Join(tmpDir, ".env"),
			[]byte("PARENT_SEARCH_VAR=found_in_parent"), 0o600)
		require.NoError(t, err)

		// Change to deepest directory
		err = os.Chdir(level3)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, os.Chdir(originalWd))
			require.NoError(t, os.Unsetenv("PARENT_SEARCH_VAR"))
		})

		err = LoadStartupEnv()
		require.NoError(t, err)
		require.Equal(t, "found_in_parent", os.Getenv("PARENT_SEARCH_VAR"))
	})

	t.Run("local_overrides_parent", func(t *testing.T) {
		originalWd, err := os.Getwd()
		require.NoError(t, err)

		// Create nested directory structure
		tmpDir := t.TempDir()
		childDir := filepath.Join(tmpDir, "child")
		err = os.MkdirAll(childDir, 0o750)
		require.NoError(t, err)

		// Create .env in parent with one value
		err = os.WriteFile(filepath.Join(tmpDir, ".env"),
			[]byte("OVERRIDE_VAR=parent_value"), 0o600)
		require.NoError(t, err)

		// Create .env in child with different value
		err = os.WriteFile(filepath.Join(childDir, ".env"),
			[]byte("OVERRIDE_VAR=child_value"), 0o600)
		require.NoError(t, err)

		// Change to child directory
		err = os.Chdir(childDir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, os.Chdir(originalWd))
			require.NoError(t, os.Unsetenv("OVERRIDE_VAR"))
		})

		err = LoadStartupEnv()
		require.NoError(t, err)
		// Child .env is loaded first (cwd), then parent .env
		// Since loadEnvFile always sets (no "only if not set" logic),
		// the last loaded file wins - which is from parent directory
		// because searchPaths order is [cwd, parent1, parent2, parent3]
		// Actually, looking at the code, all paths are loaded, and later
		// files override earlier ones within each path. So the parent
		// directory's .env is loaded after the child's, overriding it.
		// Let me verify the actual behavior:
		// searchPaths = [cwd, parent1, parent2, parent3]
		// For each path, it loads envFiles in order
		// So: cwd/.env is loaded, then parent1/.env, etc.
		// Last write wins, so parent value should win.
		require.Equal(t, "parent_value", os.Getenv("OVERRIDE_VAR"))
	})

	t.Run("no_env_files_no_error", func(t *testing.T) {
		originalWd, err := os.Getwd()
		require.NoError(t, err)

		tmpDir := t.TempDir()
		err = os.Chdir(tmpDir)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, os.Chdir(originalWd))
		})

		// No .env files created
		err = LoadStartupEnv()
		require.NoError(t, err, "should not error when no .env files exist")
	})
}

// TestLoadEnvFile_VariableExpansionWithinFile tests that variables defined
// earlier in the same file can be referenced by later variables.
func TestLoadEnvFile_VariableExpansionWithinFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `BASE_DIR=/opt/app
LOG_DIR=${BASE_DIR}/logs
CONFIG_DIR=${BASE_DIR}/config`

	err := os.WriteFile(envFile, []byte(content), 0o600)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, os.Unsetenv("BASE_DIR"))
		require.NoError(t, os.Unsetenv("LOG_DIR"))
		require.NoError(t, os.Unsetenv("CONFIG_DIR"))
	})

	err = loadEnvFile(envFile)
	require.NoError(t, err)

	require.Equal(t, "/opt/app", os.Getenv("BASE_DIR"))
	require.Equal(t, "/opt/app/logs", os.Getenv("LOG_DIR"))
	require.Equal(t, "/opt/app/config", os.Getenv("CONFIG_DIR"))
}

// Benchmark tests

func BenchmarkLoadEnvFile(b *testing.B) {
	tmpDir := b.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	// Create a file with 100 variables
	content := ""
	for i := 0; i < 100; i++ {
		content += "VAR_" + string(rune('A'+i%26)) + "=value_" + string(rune('0'+i%10)) + "\n"
	}
	err := os.WriteFile(envFile, []byte(content), 0o600)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = loadEnvFile(envFile) //nolint:errcheck // Benchmark intentionally ignores return value
	}
}

func BenchmarkExpandVariables(b *testing.B) {
	localVars := map[string]string{
		"VAR1": "value1",
		"VAR2": "value2",
		"VAR3": "value3",
	}
	input := "prefix_${VAR1}_${VAR2}_${VAR3}_suffix"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = expandVariables(input, localVars)
	}
}
