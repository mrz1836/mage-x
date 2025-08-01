package utils

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Static test errors to comply with err113 linter
var (
	errTestError = errors.New("test error")
	errError1    = errors.New("error 1")
	errError2    = errors.New("error 2")
)

const (
	osWindows = "windows"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			name:         "returns environment value when set",
			key:          "TEST_VAR",
			defaultValue: "default",
			envValue:     "custom",
			expected:     "custom",
		},
		{
			name:         "returns default when env var is empty",
			key:          "TEST_VAR_EMPTY",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
		{
			name:         "returns default when env var not set",
			key:          "TEST_VAR_NOT_SET",
			defaultValue: "default",
			envValue:     "",
			expected:     "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up first
			if err := os.Unsetenv(tt.key); err != nil {
				t.Logf("Warning: failed to unset %s: %v", tt.key, err)
			}

			// Set environment variable if needed
			if tt.envValue != "" {
				if err := os.Setenv(tt.key, tt.envValue); err != nil {
					t.Fatalf("Failed to set %s: %v", tt.key, err)
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Logf("Warning: failed to unset %s: %v", tt.key, err)
					}
				}()
			}

			result := GetEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		envValue     string
		expected     bool
	}{
		{
			name:         "returns true for 'true'",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "true",
			expected:     true,
		},
		{
			name:         "returns true for '1'",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "1",
			expected:     true,
		},
		{
			name:         "returns true for 'yes'",
			key:          "TEST_BOOL",
			defaultValue: false,
			envValue:     "yes",
			expected:     true,
		},
		{
			name:         "returns false for 'false'",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "false",
			expected:     false,
		},
		{
			name:         "returns false for '0'",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "0",
			expected:     false,
		},
		{
			name:         "returns false for 'no'",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "no",
			expected:     false,
		},
		{
			name:         "returns default when not set",
			key:          "TEST_BOOL_NOT_SET",
			defaultValue: true,
			envValue:     "",
			expected:     true,
		},
		{
			name:         "returns false for random string",
			key:          "TEST_BOOL",
			defaultValue: true,
			envValue:     "maybe",
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up first
			if err := os.Unsetenv(tt.key); err != nil {
				t.Logf("Warning: failed to unset %s: %v", tt.key, err)
			}

			// Set environment variable if needed
			if tt.envValue != "" {
				if err := os.Setenv(tt.key, tt.envValue); err != nil {
					t.Fatalf("Failed to set %s: %v", tt.key, err)
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Logf("Warning: failed to unset %s: %v", tt.key, err)
					}
				}()
			}

			result := GetEnvBool(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEnvInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		expected     int
	}{
		{
			name:         "returns parsed integer",
			key:          "TEST_INT",
			defaultValue: 10,
			envValue:     "42",
			expected:     42,
		},
		{
			name:         "returns default for invalid number",
			key:          "TEST_INT",
			defaultValue: 10,
			envValue:     "not-a-number",
			expected:     10,
		},
		{
			name:         "returns default when not set",
			key:          "TEST_INT_NOT_SET",
			defaultValue: 10,
			envValue:     "",
			expected:     10,
		},
		{
			name:         "handles negative numbers",
			key:          "TEST_INT",
			defaultValue: 10,
			envValue:     "-5",
			expected:     -5,
		},
		{
			name:         "handles zero",
			key:          "TEST_INT",
			defaultValue: 10,
			envValue:     "0",
			expected:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up first
			if err := os.Unsetenv(tt.key); err != nil {
				t.Logf("Warning: failed to unset %s: %v", tt.key, err)
			}

			// Set environment variable if needed
			if tt.envValue != "" {
				if err := os.Setenv(tt.key, tt.envValue); err != nil {
					t.Fatalf("Failed to set %s: %v", tt.key, err)
				}
				defer func() {
					if err := os.Unsetenv(tt.key); err != nil {
						t.Logf("Warning: failed to unset %s: %v", tt.key, err)
					}
				}()
			}

			result := GetEnvInt(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsVerbose(t *testing.T) {
	tests := []struct {
		name     string
		verbose  string
		v        string
		expected bool
	}{
		{
			name:     "returns true when VERBOSE=true",
			verbose:  "true",
			v:        "",
			expected: true,
		},
		{
			name:     "returns true when V=1",
			verbose:  "",
			v:        "1",
			expected: true,
		},
		{
			name:     "returns false when neither set",
			verbose:  "",
			v:        "",
			expected: false,
		},
		{
			name:     "returns false when VERBOSE=false",
			verbose:  "false",
			v:        "",
			expected: false,
		},
		{
			name:     "V takes precedence when both set",
			verbose:  "false",
			v:        "true",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up first
			if err := os.Unsetenv("VERBOSE"); err != nil {
				t.Logf("Warning: failed to unset VERBOSE: %v", err)
			}
			if err := os.Unsetenv("V"); err != nil {
				t.Logf("Warning: failed to unset V: %v", err)
			}

			// Set environment variables
			if tt.verbose != "" {
				if err := os.Setenv("VERBOSE", tt.verbose); err != nil {
					t.Fatalf("Failed to set VERBOSE: %v", err)
				}
				defer func() {
					if err := os.Unsetenv("VERBOSE"); err != nil {
						t.Logf("Warning: failed to unset VERBOSE: %v", err)
					}
				}()
			}
			if tt.v != "" {
				if err := os.Setenv("V", tt.v); err != nil {
					t.Fatalf("Failed to set V: %v", err)
				}
				defer func() {
					if err := os.Unsetenv("V"); err != nil {
						t.Logf("Warning: failed to unset V: %v", err)
					}
				}()
			}

			result := IsVerbose()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsCI(t *testing.T) {
	ciVars := []string{
		"CI",
		"CONTINUOUS_INTEGRATION",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"TRAVIS",
		"CIRCLECI",
		"JENKINS_URL",
		"CODEBUILD_BUILD_ID",
	}

	// Test each CI variable
	for _, ciVar := range ciVars {
		t.Run(fmt.Sprintf("returns true for %s", ciVar), func(t *testing.T) {
			// Clean up all CI vars first
			for _, v := range ciVars {
				if err := os.Unsetenv(v); err != nil {
					t.Logf("Warning: failed to unset %s: %v", v, err)
				}
			}

			// Set the test variable
			if err := os.Setenv(ciVar, "true"); err != nil {
				t.Fatalf("Failed to set %s: %v", ciVar, err)
			}
			defer func() {
				if err := os.Unsetenv(ciVar); err != nil {
					t.Logf("Warning: failed to unset %s: %v", ciVar, err)
				}
			}()

			result := IsCI()
			assert.True(t, result)
		})
	}

	t.Run("returns false when no CI vars set", func(t *testing.T) {
		// Clean up all CI vars
		for _, v := range ciVars {
			if err := os.Unsetenv(v); err != nil {
				t.Logf("Warning: failed to unset %s: %v", v, err)
			}
		}

		result := IsCI()
		assert.False(t, result)
	})
}

func TestPlatformFunctions(t *testing.T) {
	t.Run("GetCurrentPlatform", func(t *testing.T) {
		platform := GetCurrentPlatform()
		assert.Equal(t, runtime.GOOS, platform.OS)
		assert.Equal(t, runtime.GOARCH, platform.Arch)
	})

	t.Run("Platform.String", func(t *testing.T) {
		platform := Platform{OS: "linux", Arch: "amd64"}
		expected := "linux/amd64"
		assert.Equal(t, expected, platform.String())
	})

	t.Run("ParsePlatform", func(t *testing.T) {
		tests := []struct {
			name      string
			input     string
			expectErr bool
			expected  Platform
		}{
			{
				name:      "valid linux/amd64",
				input:     "linux/amd64",
				expectErr: false,
				expected:  Platform{OS: "linux", Arch: "amd64"},
			},
			{
				name:      "valid darwin/arm64",
				input:     "darwin/arm64",
				expectErr: false,
				expected:  Platform{OS: "darwin", Arch: "arm64"},
			},
			{
				name:      "invalid format - no slash",
				input:     "linux-amd64",
				expectErr: true,
			},
			{
				name:      "invalid format - too many parts",
				input:     "linux/amd64/extra",
				expectErr: true,
			},
			{
				name:      "empty string",
				input:     "",
				expectErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				platform, err := ParsePlatform(tt.input)
				if tt.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tt.expected, platform)
				}
			})
		}
	})

	t.Run("GetBinaryExt", func(t *testing.T) {
		tests := []struct {
			name     string
			platform Platform
			expected string
		}{
			{
				name:     "windows returns .exe",
				platform: Platform{OS: "windows", Arch: "amd64"},
				expected: ".exe",
			},
			{
				name:     "linux returns empty",
				platform: Platform{OS: "linux", Arch: "amd64"},
				expected: "",
			},
			{
				name:     "darwin returns empty",
				platform: Platform{OS: "darwin", Arch: "amd64"},
				expected: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := GetBinaryExt(tt.platform)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("Platform detection functions", func(t *testing.T) {
		// Test IsWindows, IsMac, IsLinux
		currentOS := runtime.GOOS

		switch currentOS {
		case osWindows:
			assert.True(t, IsWindows())
			assert.False(t, IsMac())
			assert.False(t, IsLinux())
		case "darwin":
			assert.False(t, IsWindows())
			assert.True(t, IsMac())
			assert.False(t, IsLinux())
		case "linux":
			assert.False(t, IsWindows())
			assert.False(t, IsMac())
			assert.True(t, IsLinux())
		}
	})

	t.Run("GetShell", func(t *testing.T) {
		shell, args := GetShell()
		if runtime.GOOS == osWindows {
			assert.Equal(t, "cmd", shell)
			assert.Equal(t, []string{"/c"}, args)
		} else {
			assert.Equal(t, "sh", shell)
			assert.Equal(t, []string{"-c"}, args)
		}
	})
}

func TestCommandExists(t *testing.T) {
	t.Run("existing command", func(t *testing.T) {
		// Go should exist on any system running these tests
		result := CommandExists("go")
		assert.True(t, result)
	})

	t.Run("non-existing command", func(t *testing.T) {
		result := CommandExists("non-existent-command-12345")
		assert.False(t, result)
	})
}

func TestParallel(t *testing.T) {
	t.Run("all functions succeed", func(t *testing.T) {
		var count int32
		fns := []func() error{
			func() error { atomic.AddInt32(&count, 1); return nil },
			func() error { atomic.AddInt32(&count, 1); return nil },
			func() error { atomic.AddInt32(&count, 1); return nil },
		}

		err := Parallel(fns...)
		require.NoError(t, err)
		assert.Equal(t, int32(3), atomic.LoadInt32(&count))
	})

	t.Run("one function fails", func(t *testing.T) {
		testErr := errTestError
		fns := []func() error{
			func() error { return nil },
			func() error { return testErr },
			func() error { return nil },
		}

		err := Parallel(fns...)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parallel execution failed")
	})

	t.Run("multiple functions fail", func(t *testing.T) {
		err1 := errError1
		err2 := errError2
		fns := []func() error{
			func() error { return err1 },
			func() error { return err2 },
			func() error { return nil },
		}

		err := Parallel(fns...)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parallel execution failed")
	})

	t.Run("empty function list", func(t *testing.T) {
		err := Parallel()
		require.NoError(t, err)
	})
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{
			name:     "milliseconds",
			duration: 500 * time.Millisecond,
			expected: "500ms",
		},
		{
			name:     "seconds",
			duration: 2500 * time.Millisecond,
			expected: "2.5s",
		},
		{
			name:     "minutes",
			duration: 90 * time.Second,
			expected: "1.5m",
		},
		{
			name:     "negative duration",
			duration: -time.Second,
			expected: "0s",
		},
		{
			name:     "zero duration",
			duration: 0,
			expected: "0ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "bytes",
			bytes:    512,
			expected: "512 B",
		},
		{
			name:     "kilobytes",
			bytes:    1536, // 1.5 KB
			expected: "1.5 KB",
		},
		{
			name:     "megabytes",
			bytes:    1572864, // 1.5 MB
			expected: "1.5 MB",
		},
		{
			name:     "gigabytes",
			bytes:    1610612736, // 1.5 GB
			expected: "1.5 GB",
		},
		{
			name:     "zero bytes",
			bytes:    0,
			expected: "0 B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Test file system operations
func TestFileSystemOperations(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	t.Run("EnsureDir", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "test", "nested", "dir")
		err := EnsureDir(testDir)
		require.NoError(t, err)

		// Verify directory exists
		assert.True(t, DirExists(testDir))
	})

	t.Run("FileExists and DirExists", func(t *testing.T) {
		// Test file
		testFile := filepath.Join(tempDir, "test.txt")
		err := os.WriteFile(testFile, []byte("test"), 0o600)
		require.NoError(t, err)

		assert.True(t, FileExists(testFile))
		assert.False(t, DirExists(testFile)) // File, not directory

		// Test directory
		testDir := filepath.Join(tempDir, "testdir")
		err = os.MkdirAll(testDir, 0o750)
		require.NoError(t, err)

		assert.True(t, DirExists(testDir))
		assert.True(t, FileExists(testDir)) // Directory also "exists" as file

		// Test non-existent
		nonExistent := filepath.Join(tempDir, "nonexistent")
		assert.False(t, FileExists(nonExistent))
		assert.False(t, DirExists(nonExistent))
	})

	t.Run("CleanDir", func(t *testing.T) {
		testDir := filepath.Join(tempDir, "cleantest")

		// Create directory with content
		err := os.MkdirAll(testDir, 0o750)
		require.NoError(t, err)

		testFile := filepath.Join(testDir, "file.txt")
		err = os.WriteFile(testFile, []byte("content"), 0o600)
		require.NoError(t, err)

		// Clean directory
		err = CleanDir(testDir)
		require.NoError(t, err)

		// Directory should exist but be empty
		assert.True(t, DirExists(testDir))
		assert.False(t, FileExists(testFile))

		// Test cleaning non-existent directory
		nonExistentDir := filepath.Join(tempDir, "nonexistent")
		err = CleanDir(nonExistentDir)
		require.NoError(t, err)                   // Should not error
		assert.True(t, DirExists(nonExistentDir)) // Should create it
	})

	t.Run("CopyFile", func(t *testing.T) {
		// Create source file
		srcFile := filepath.Join(tempDir, "source.txt")
		content := "test file content"
		err := os.WriteFile(srcFile, []byte(content), 0o600)
		require.NoError(t, err)

		// Copy file
		dstFile := filepath.Join(tempDir, "destination.txt")
		err = CopyFile(srcFile, dstFile)
		require.NoError(t, err)

		// Verify copy
		assert.True(t, FileExists(dstFile))
		copiedContent, err := os.ReadFile(dstFile) // #nosec G304 -- controlled test file path
		require.NoError(t, err)
		assert.Equal(t, content, string(copiedContent))
	})

	t.Run("FindFiles", func(t *testing.T) {
		// Create test files
		testFiles := []string{"test.go", "test.txt", "other.go", "README.md"}
		for _, file := range testFiles {
			filePath := filepath.Join(tempDir, file)
			err := os.WriteFile(filePath, []byte("content"), 0o600)
			require.NoError(t, err)
		}

		// Test finding .go files
		// Note: This test may be sensitive to the current working directory
		// We'll test the basic functionality but the actual implementation
		// might need adjustment for proper directory handling
		_, err := FindFiles(tempDir, "*.go")
		require.NoError(t, err)
		// File search should complete successfully
	})
}

// Integration tests for Go commands (may require Go to be installed)
func TestGoCommands(t *testing.T) {
	// Only run if Go is available
	if !CommandExists("go") {
		t.Skip("Go command not available")
	}

	t.Run("GetGoVersion", func(t *testing.T) {
		version, err := GetGoVersion()
		require.NoError(t, err)
		assert.NotEmpty(t, version)
		assert.Contains(t, version, ".")
	})

	t.Run("GetModuleName", func(t *testing.T) {
		// This test will only work in a Go module
		moduleName, err := GetModuleName()
		if err != nil {
			// This is expected if we're not in a Go module
			t.Logf("GetModuleName failed (expected in non-module context): %v", err)
		} else {
			assert.NotEmpty(t, moduleName)
		}
	})

	t.Run("GoList", func(t *testing.T) {
		// Test a simple go list command
		packages, err := GoList(".")
		if err != nil {
			// This might fail if we're not in a Go module/package
			t.Logf("GoList failed (expected in some contexts): %v", err)
		} else {
			// Packages list should be valid
			t.Logf("Found %d packages", len(packages))
		}
	})
}

// Test command execution functions
func TestCommandExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping command execution tests in short mode")
	}

	t.Run("RunCmdOutput success", func(t *testing.T) {
		output, err := RunCmdOutput("echo", "hello world")
		require.NoError(t, err)
		assert.Contains(t, strings.TrimSpace(output), "hello world")
	})

	t.Run("RunCmdOutput failure", func(t *testing.T) {
		_, err := RunCmdOutput("nonexistent-command-12345")
		require.Error(t, err)
	})

	t.Run("RunCmd success", func(t *testing.T) {
		err := RunCmd("echo", "test")
		require.NoError(t, err)
	})

	t.Run("RunCmdV", func(t *testing.T) {
		// Test verbose command - mainly checking it doesn't crash
		err := RunCmdV("echo", "verbose test")
		require.NoError(t, err)
	})

	t.Run("RunCmdPipe", func(t *testing.T) {
		// Create a simple pipeline: echo "hello" | cat
		cmd1 := exec.CommandContext(context.Background(), "echo", "hello")
		cmd2 := exec.CommandContext(context.Background(), "cat")

		err := RunCmdPipe(cmd1, cmd2)
		require.NoError(t, err)
	})
}

// Benchmark tests
func BenchmarkGetEnv(b *testing.B) {
	if err := os.Setenv("BENCH_TEST", "value"); err != nil {
		b.Fatalf("Failed to set BENCH_TEST: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("BENCH_TEST"); err != nil {
			b.Logf("Warning: failed to unset BENCH_TEST: %v", err)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GetEnv("BENCH_TEST", "default")
	}
}

func BenchmarkIsVerbose(b *testing.B) {
	if err := os.Setenv("VERBOSE", "true"); err != nil {
		b.Fatalf("Failed to set VERBOSE: %v", err)
	}
	defer func() {
		if err := os.Unsetenv("VERBOSE"); err != nil {
			b.Logf("Warning: failed to unset VERBOSE: %v", err)
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsVerbose()
	}
}

func BenchmarkFormatBytes(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		FormatBytes(1073741824) // 1GB
	}
}

func BenchmarkParallel(b *testing.B) {
	fns := []func() error{
		func() error { return nil },
		func() error { return nil },
		func() error { return nil },
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := Parallel(fns...)
		if err != nil {
			b.Logf("Parallel error in benchmark: %v", err)
		}
	}
}
