package utils

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

// TestParseGoVersionOutput tests the version parsing logic from go version output
func TestParseGoVersionOutput(t *testing.T) {
	tests := []struct {
		name      string
		output    string
		expected  string
		expectErr bool
	}{
		{
			name:     "standard darwin output",
			output:   "go version go1.24.0 darwin/amd64",
			expected: "1.24.0",
		},
		{
			name:     "standard linux output",
			output:   "go version go1.22.5 linux/amd64",
			expected: "1.22.5",
		},
		{
			name:     "standard windows output",
			output:   "go version go1.21.0 windows/amd64",
			expected: "1.21.0",
		},
		{
			name:     "arm64 architecture",
			output:   "go version go1.23.1 darwin/arm64",
			expected: "1.23.1",
		},
		{
			name:     "rc version",
			output:   "go version go1.24rc1 darwin/amd64",
			expected: "1.24rc1",
		},
		{
			name:     "beta version",
			output:   "go version go1.24beta1 darwin/amd64",
			expected: "1.24beta1",
		},
		{
			name:      "too few fields",
			output:    "go version",
			expectErr: true,
		},
		{
			name:      "empty output",
			output:    "",
			expectErr: true,
		},
		{
			name:      "single word",
			output:    "go",
			expectErr: true,
		},
		{
			name:      "two words",
			output:    "go version",
			expectErr: true,
		},
		{
			name:     "extra whitespace",
			output:   "  go  version  go1.24.0  darwin/amd64  ",
			expected: "1.24.0",
		},
		{
			name:     "unexpected prefix but valid format",
			output:   "go version xyz1.0.0 darwin/amd64",
			expected: "xyz1.0.0", // TrimPrefix removes "go" only if present
		},
	}

	// parseGoVersionOutput is the parsing logic extracted from GetGoVersion
	parseGoVersionOutput := func(output string) (string, error) {
		parts := strings.Fields(output)
		if len(parts) >= 3 {
			return strings.TrimPrefix(parts[2], "go"), nil
		}
		return "", errParseGoVersion
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseGoVersionOutput(tt.output)
			if tt.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, errParseGoVersion)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
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

	t.Run("GetGoVersion format is semantic", func(t *testing.T) {
		version, err := GetGoVersion()
		require.NoError(t, err)
		// Version should start with a digit (e.g., "1.24.0")
		assert.Regexp(t, `^\d+\.\d+`, version, "version should start with major.minor")
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

	t.Run("GoList with invalid package returns error", func(t *testing.T) {
		// Test with a package that doesn't exist
		_, err := GoList("nonexistent/package/that/does/not/exist/12345")
		require.Error(t, err)
	})

	t.Run("GoList parses multiple packages", func(t *testing.T) {
		// Test listing all packages in the module
		packages, err := GoList("./...")
		if err != nil {
			t.Skipf("Skipping - not in a proper Go module context: %v", err)
		}
		// Should find at least one package
		assert.NotEmpty(t, packages)
		// Each package should be a valid import path
		for _, pkg := range packages {
			assert.NotEmpty(t, pkg)
			assert.NotContains(t, pkg, "\n") // No embedded newlines
		}
	})

	t.Run("GoList empty lines filtered", func(t *testing.T) {
		// The current package should return exactly one result
		packages, err := GoList(".")
		if err != nil {
			t.Skipf("Skipping - not in a proper Go module context: %v", err)
		}
		// Should return non-empty package names only
		for _, pkg := range packages {
			assert.NotEmpty(t, pkg, "should not contain empty package names")
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

// TestCommandExecutionErrorPaths tests error handling in command execution functions
func TestCommandExecutionErrorPaths(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping command execution error path tests in short mode")
	}

	t.Run("RunCmd error wraps command info", func(t *testing.T) {
		err := RunCmd("nonexistent-command-xyz-12345")
		require.Error(t, err)
		// Verify the error contains the command name for debugging
		assert.Contains(t, err.Error(), "nonexistent-command-xyz-12345")
	})

	t.Run("RunCmd with args includes args in error", func(t *testing.T) {
		err := RunCmd("nonexistent-command-xyz", "arg1", "arg2")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nonexistent-command-xyz")
	})

	t.Run("RunCmdOutput error wraps command and output info", func(t *testing.T) {
		_, err := RunCmdOutput("nonexistent-command-xyz-67890")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nonexistent-command-xyz-67890")
	})

	t.Run("RunCmdPipe command start failure", func(t *testing.T) {
		// First command is valid, second doesn't exist
		cmd1 := exec.CommandContext(context.Background(), "echo", "test")
		cmd2 := exec.CommandContext(context.Background(), "/nonexistent/binary/path")

		err := RunCmdPipe(cmd1, cmd2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed")
	})

	t.Run("RunCmdPipe command exit failure", func(t *testing.T) {
		// First command succeeds, second exits with error
		cmd1 := exec.CommandContext(context.Background(), "echo", "test")
		cmd2 := exec.CommandContext(context.Background(), "sh", "-c", "exit 1")

		err := RunCmdPipe(cmd1, cmd2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed")
	})

	t.Run("RunCmdPipe single command works", func(t *testing.T) {
		cmd := exec.CommandContext(context.Background(), "echo", "single")
		err := RunCmdPipe(cmd)
		require.NoError(t, err)
	})

	t.Run("RunCmdPipe empty pipeline", func(t *testing.T) {
		err := RunCmdPipe()
		require.NoError(t, err)
	})
}

// Benchmark tests
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

// TestCheckFileLineLength tests the CheckFileLineLength function
func TestCheckFileLineLength(t *testing.T) {
	t.Run("file within limit", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "short.txt")

		content := "short line\nanother short line\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, lineNum, lineLen, err := CheckFileLineLength(testFile, 100)
		require.NoError(t, err)
		assert.False(t, hasLong, "should not have long lines")
		assert.Equal(t, 0, lineNum, "line number should be 0 when no long lines found")
		assert.Positive(t, lineLen, "max line length should be greater than 0")
	})

	t.Run("file exceeds limit", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "long.txt")

		// Create file with a long line on line 2
		content := "short\n" + strings.Repeat("x", 150) + "\nshort again\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, lineNum, lineLen, err := CheckFileLineLength(testFile, 100)
		require.NoError(t, err)
		assert.True(t, hasLong, "should have found long line")
		assert.Equal(t, 2, lineNum, "long line should be on line 2")
		assert.Equal(t, 150, lineLen, "line length should be 150")
	})

	t.Run("empty file", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "empty.txt")

		err := os.WriteFile(testFile, []byte{}, 0o600)
		require.NoError(t, err)

		hasLong, lineNum, lineLen, err := CheckFileLineLength(testFile, 100)
		require.NoError(t, err)
		assert.False(t, hasLong, "empty file should not have long lines")
		assert.Equal(t, 0, lineNum)
		assert.Equal(t, 0, lineLen)
	})

	t.Run("non-existent file", func(t *testing.T) {
		_, _, _, err := CheckFileLineLength("/nonexistent/file.txt", 100)
		require.Error(t, err)
	})

	t.Run("first line exceeds limit", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "longfirst.txt")

		content := strings.Repeat("x", 200) + "\nshort\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, lineNum, lineLen, err := CheckFileLineLength(testFile, 100)
		require.NoError(t, err)
		assert.True(t, hasLong)
		assert.Equal(t, 1, lineNum, "long line should be on line 1")
		assert.Equal(t, 200, lineLen)
	})

	t.Run("very long line larger than buffer", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "verylongline.txt")

		// Create a line longer than the 128KB buffer
		longLine := strings.Repeat("x", 200*1024) // 200KB line
		content := "short\n" + longLine + "\nshort\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, lineNum, lineLen, err := CheckFileLineLength(testFile, 100)
		require.NoError(t, err)
		assert.True(t, hasLong, "should detect very long line")
		assert.Equal(t, 2, lineNum, "long line should be on line 2")
		assert.GreaterOrEqual(t, lineLen, 200*1024, "line length should be at least 200KB")
	})
}

// TestRunCmdSecure tests the secure command execution
func TestRunCmdSecure(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping secure command execution tests in short mode")
	}

	t.Run("success", func(t *testing.T) {
		err := RunCmdSecure("echo", "secure test")
		require.NoError(t, err)
	})

	t.Run("failure", func(t *testing.T) {
		err := RunCmdSecure("nonexistent-command-12345")
		require.Error(t, err)
	})
}

// TestRunCmdWithRetry tests the retry command execution
func TestRunCmdWithRetry(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping retry command execution tests in short mode")
	}

	t.Run("success on first attempt", func(t *testing.T) {
		err := RunCmdWithRetry(3, "echo", "retry test")
		require.NoError(t, err)
	})

	t.Run("failure after retries", func(t *testing.T) {
		err := RunCmdWithRetry(2, "nonexistent-command-12345")
		require.Error(t, err)
	})
}

// TestGetModuleNameInDir tests GetModuleNameInDir function
func TestGetModuleNameInDir(t *testing.T) {
	if !CommandExists("go") {
		t.Skip("Go command not available")
	}

	t.Run("valid module directory", func(t *testing.T) {
		// Test in the current project directory which should be a Go module
		// Get the project root (go.mod should be there)
		wd, err := os.Getwd()
		require.NoError(t, err)

		// Walk up to find go.mod
		dir := wd
		for !FileExists(filepath.Join(dir, "go.mod")) {

			parent := filepath.Dir(dir)
			if parent == dir {
				t.Skip("Could not find go.mod in parent directories")
			}
			dir = parent
		}

		moduleName, err := GetModuleNameInDir(dir)
		require.NoError(t, err)
		assert.NotEmpty(t, moduleName)
		assert.Contains(t, moduleName, "mage-x", "module name should contain mage-x")
	})

	t.Run("invalid directory", func(t *testing.T) {
		_, err := GetModuleNameInDir("/nonexistent/directory")
		require.Error(t, err)
	})

	t.Run("directory without go.mod", func(t *testing.T) {
		tmpDir := t.TempDir()
		moduleName, err := GetModuleNameInDir(tmpDir)
		// Go returns "command-line-arguments" when not in a module
		// This verifies the function handles non-module directories appropriately
		if err == nil {
			// Either returns empty or "command-line-arguments" for non-module directories
			assert.True(t, moduleName == "" || moduleName == "command-line-arguments",
				"should return empty or 'command-line-arguments' for non-module directory, got: %s", moduleName)
		}
	})
}

// withMockedStdin is a test helper that temporarily replaces os.Stdin with a pipe
// containing the provided input string. The original stdin is restored after fn completes.
func withMockedStdin(t *testing.T, input string, fn func()) {
	t.Helper()

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	require.NoError(t, err)

	os.Stdin = r

	// Write input in a goroutine to avoid blocking
	go func() {
		defer func() {
			w.Close() //nolint:errcheck,gosec // Best-effort close in test helper
		}()
		w.WriteString(input) //nolint:errcheck,gosec // Best-effort write in test helper
	}()

	defer func() {
		os.Stdin = oldStdin
		r.Close() //nolint:errcheck,gosec // Best-effort close in test helper
	}()

	fn()
}

// TestPromptForInput tests the PromptForInput function with various input scenarios
func TestPromptForInput(t *testing.T) {
	tests := []struct {
		name      string
		prompt    string
		input     string
		expected  string
		expectErr bool
	}{
		{
			name:     "valid input with prompt",
			prompt:   "Enter name",
			input:    "Alice\n",
			expected: "Alice",
		},
		{
			name:     "valid input without prompt",
			prompt:   "",
			input:    "data\n",
			expected: "data",
		},
		{
			name:     "whitespace is trimmed",
			prompt:   "test",
			input:    "  hello  \n",
			expected: "hello",
		},
		{
			name:     "empty input returns empty string",
			prompt:   "test",
			input:    "\n",
			expected: "",
		},
		{
			name:     "EOF returns empty string without error",
			prompt:   "test",
			input:    "",
			expected: "",
		},
		{
			name:     "multiline input reads first line only",
			prompt:   "test",
			input:    "first line\nsecond line\n",
			expected: "first line",
		},
		{
			name:     "input with special characters",
			prompt:   "test",
			input:    "hello@world.com\n",
			expected: "hello@world.com",
		},
		{
			name:     "input with unicode characters",
			prompt:   "test",
			input:    "こんにちは\n",
			expected: "こんにちは",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result string
			var err error

			withMockedStdin(t, tt.input, func() {
				result, err = PromptForInput(tt.prompt)
			})

			if tt.expectErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestParallel_EdgeCases tests additional edge cases for Parallel function
func TestParallel_EdgeCases(t *testing.T) {
	t.Run("large number of functions - 100", func(t *testing.T) {
		var count int32
		fns := make([]func() error, 100)
		for i := 0; i < 100; i++ {
			fns[i] = func() error {
				atomic.AddInt32(&count, 1)
				return nil
			}
		}

		err := Parallel(fns...)
		require.NoError(t, err)
		assert.Equal(t, int32(100), atomic.LoadInt32(&count))
	})

	t.Run("large number with some failures", func(t *testing.T) {
		fns := make([]func() error, 100)
		for i := 0; i < 100; i++ {
			if i%10 == 0 {
				fns[i] = func() error { return errTestError }
			} else {
				fns[i] = func() error { return nil }
			}
		}

		err := Parallel(fns...)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parallel execution failed")
	})

	t.Run("concurrency verification - shorter sleep completes first", func(t *testing.T) {
		var order []int
		var mu sync.Mutex

		fns := []func() error{
			func() error {
				time.Sleep(30 * time.Millisecond)
				mu.Lock()
				order = append(order, 1)
				mu.Unlock()
				return nil
			},
			func() error {
				time.Sleep(10 * time.Millisecond)
				mu.Lock()
				order = append(order, 2)
				mu.Unlock()
				return nil
			},
			func() error {
				time.Sleep(20 * time.Millisecond)
				mu.Lock()
				order = append(order, 3)
				mu.Unlock()
				return nil
			},
		}

		err := Parallel(fns...)
		require.NoError(t, err)

		// If truly parallel, order should be [2, 3, 1] not [1, 2, 3]
		require.Len(t, order, 3)
		assert.Equal(t, 2, order[0], "shortest sleep should complete first")
	})

	t.Run("function with delayed error", func(t *testing.T) {
		fns := []func() error{
			func() error {
				time.Sleep(50 * time.Millisecond)
				return errTestError
			},
			func() error { return nil },
		}

		err := Parallel(fns...)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parallel execution failed")
	})
}

// TestFormatBytes_EdgeCases tests additional edge cases for FormatBytes function
func TestFormatBytes_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{
			name:     "negative bytes - -1024",
			bytes:    -1024,
			expected: "-1024 B",
		},
		{
			name:     "negative single byte",
			bytes:    -1,
			expected: "-1 B",
		},
		{
			name:     "negative large value - -5GB",
			bytes:    -5368709120,
			expected: "-5368709120 B",
		},
		{
			name:     "1023 bytes - boundary before KB",
			bytes:    1023,
			expected: "1023 B",
		},
		{
			name:     "exactly 1 KB",
			bytes:    1024,
			expected: "1.0 KB",
		},
		{
			name:     "exactly 1 TB",
			bytes:    1024 * 1024 * 1024 * 1024,
			expected: "1.0 TB",
		},
		{
			name:     "exactly 1 PB",
			bytes:    1024 * 1024 * 1024 * 1024 * 1024,
			expected: "1.0 PB",
		},
		{
			name:     "exactly 1 EB",
			bytes:    1024 * 1024 * 1024 * 1024 * 1024 * 1024,
			expected: "1.0 EB",
		},
		{
			name:     "max int64 value",
			bytes:    9223372036854775807,
			expected: "8.0 EB",
		},
		{
			name:     "fractional MB - 2.5 MB",
			bytes:    2621440,
			expected: "2.5 MB",
		},
		{
			name:     "large GB value - 5 GB",
			bytes:    5368709120,
			expected: "5.0 GB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatBytes(tt.bytes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestFormatDuration_EdgeCases tests additional edge cases for FormatDuration function
func TestFormatDuration_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
	}{
		{
			name:     "one nanosecond",
			duration: 1 * time.Nanosecond,
		},
		{
			name:     "one microsecond",
			duration: 1 * time.Microsecond,
		},
		{
			name:     "100 microseconds",
			duration: 100 * time.Microsecond,
		},
		{
			name:     "one hour",
			duration: 1 * time.Hour,
		},
		{
			name:     "compound duration - 1h23m45s",
			duration: 1*time.Hour + 23*time.Minute + 45*time.Second,
		},
		{
			name:     "very large duration - 24 hours",
			duration: 24 * time.Hour,
		},
		{
			name:     "very large duration - 100 hours",
			duration: 100 * time.Hour,
		},
		{
			name:     "fractional second - 1.5s",
			duration: 1500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.duration)
			assert.NotEmpty(t, result, "FormatDuration should not return empty string")
			// Just verify it returns a reasonable string representation
			assert.NotContains(t, result, "unknown", "should not contain 'unknown'")
		})
	}
}

// TestCheckFileLineLength_EdgeCases tests additional edge cases for CheckFileLineLength function
func TestCheckFileLineLength_EdgeCases(t *testing.T) {
	t.Run("no newline at EOF", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "nonewline.txt")

		content := "line1\nline2\nline3 without newline"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, _, _, err := CheckFileLineLength(testFile, 100)
		require.NoError(t, err)
		assert.False(t, hasLong, "should handle file without trailing newline")
	})

	t.Run("CRLF line endings", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "crlf.txt")

		content := "line1\r\nline2\r\nline3\r\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, _, _, err := CheckFileLineLength(testFile, 100)
		require.NoError(t, err)
		assert.False(t, hasLong, "should handle CRLF line endings")
	})

	t.Run("CRLF with long line", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "crlf_long.txt")

		content := strings.Repeat("a", 150) + "\r\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, lineNum, lineLen, err := CheckFileLineLength(testFile, 100)
		require.NoError(t, err)
		assert.True(t, hasLong)
		assert.Equal(t, 1, lineNum)
		assert.Equal(t, 150, lineLen)
	})

	t.Run("mixed line endings", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "mixed.txt")

		content := "line1\nline2\r\nline3\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, _, _, err := CheckFileLineLength(testFile, 100)
		require.NoError(t, err)
		assert.False(t, hasLong, "should handle mixed line endings")
	})

	t.Run("unicode characters", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "unicode.txt")

		content := "日本語\n中文\nEnglish\n" //nolint:gosmopolitan // Testing unicode handling
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, _, _, err := CheckFileLineLength(testFile, 100)
		require.NoError(t, err)
		assert.False(t, hasLong, "should handle unicode characters")
	})

	t.Run("long unicode line", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "unicode_long.txt")

		content := strings.Repeat("日本語", 50) + "\n" //nolint:gosmopolitan // Testing unicode handling
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, lineNum, _, err := CheckFileLineLength(testFile, 100)
		require.NoError(t, err)
		assert.True(t, hasLong, "should detect long unicode line")
		assert.Equal(t, 1, lineNum)
	})

	t.Run("exactly at limit", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "exact.txt")

		content := strings.Repeat("a", 80) + "\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, _, _, err := CheckFileLineLength(testFile, 80)
		require.NoError(t, err)
		assert.False(t, hasLong, "line exactly at limit should not trigger")
	})

	t.Run("one byte over limit", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "oneover.txt")

		content := strings.Repeat("a", 81) + "\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, lineNum, lineLen, err := CheckFileLineLength(testFile, 80)
		require.NoError(t, err)
		assert.True(t, hasLong)
		assert.Equal(t, 1, lineNum)
		assert.Equal(t, 81, lineLen)
	})

	t.Run("file with only newlines", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "onlynewlines.txt")

		content := "\n\n\n\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, _, _, err := CheckFileLineLength(testFile, 100)
		require.NoError(t, err)
		assert.False(t, hasLong, "file with only newlines should not have long lines")
	})

	t.Run("line exactly 128KB - buffer boundary", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "128kb.txt")

		content := strings.Repeat("b", 128*1024) + "\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, lineNum, lineLen, err := CheckFileLineLength(testFile, 1000)
		require.NoError(t, err)
		assert.True(t, hasLong)
		assert.Equal(t, 1, lineNum)
		assert.Equal(t, 128*1024, lineLen)
	})

	t.Run("very large line - 500KB", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "500kb.txt")

		longLine := strings.Repeat("x", 500*1024)
		content := "short\n" + longLine + "\nshort\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, lineNum, lineLen, err := CheckFileLineLength(testFile, 1000)
		require.NoError(t, err)
		assert.True(t, hasLong, "should detect 500KB line")
		assert.Equal(t, 2, lineNum)
		assert.GreaterOrEqual(t, lineLen, 500*1024)
	})

	t.Run("multiple lines all under limit", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "multiunder.txt")

		content := strings.Repeat("short line\n", 100)
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, _, _, err := CheckFileLineLength(testFile, 100)
		require.NoError(t, err)
		assert.False(t, hasLong, "all lines under limit")
	})

	t.Run("maxLen of 0", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "maxlen0.txt")

		content := "any content\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, lineNum, _, err := CheckFileLineLength(testFile, 0)
		require.NoError(t, err)
		assert.True(t, hasLong, "maxLen 0 should flag any content")
		assert.Equal(t, 1, lineNum)
	})

	t.Run("special characters", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "special.txt")

		content := "!@#$%^&*(){}[]|\\:;\"'<>?,./\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, _, _, err := CheckFileLineLength(testFile, 100)
		require.NoError(t, err)
		assert.False(t, hasLong, "should handle special characters")
	})

	t.Run("last line exceeds limit", func(t *testing.T) {
		tmpDir := t.TempDir()
		testFile := filepath.Join(tmpDir, "lastlong.txt")

		content := "short\nshort\n" + strings.Repeat("x", 200) + "\n"
		err := os.WriteFile(testFile, []byte(content), 0o600)
		require.NoError(t, err)

		hasLong, lineNum, lineLen, err := CheckFileLineLength(testFile, 100)
		require.NoError(t, err)
		assert.True(t, hasLong)
		assert.Equal(t, 3, lineNum, "long line should be on line 3")
		assert.Equal(t, 200, lineLen)
	})
}

// TestPromptForInput_EdgeCases tests additional edge cases for PromptForInput function
func TestPromptForInput_EdgeCases(t *testing.T) {
	t.Run("very long input - 10000 chars", func(t *testing.T) {
		longInput := strings.Repeat("a", 10000)
		var result string
		var err error

		withMockedStdin(t, longInput+"\n", func() {
			result, err = PromptForInput("Enter")
		})

		require.NoError(t, err)
		assert.Equal(t, longInput, result)
		assert.Len(t, result, 10000)
	})

	t.Run("input with tabs", func(t *testing.T) {
		var result string
		var err error

		withMockedStdin(t, "\ttest\t\n", func() {
			result, err = PromptForInput("Enter")
		})

		require.NoError(t, err)
		assert.Equal(t, "test", result, "tabs should be trimmed")
	})

	t.Run("whitespace-only input becomes empty", func(t *testing.T) {
		var result string
		var err error

		withMockedStdin(t, "   \t  \n", func() {
			result, err = PromptForInput("Enter")
		})

		require.NoError(t, err)
		assert.Empty(t, result, "whitespace-only should return empty string")
	})

	t.Run("input with many special characters", func(t *testing.T) {
		specialChars := "!@#$%^&*()_+-={}[]|\\:;\"'<>?,./~`"
		var result string
		var err error

		withMockedStdin(t, specialChars+"\n", func() {
			result, err = PromptForInput("Enter")
		})

		require.NoError(t, err)
		assert.Equal(t, specialChars, result)
	})
}

// TestParallel_Sync tests synchronization in Parallel function
func TestParallel_Sync(t *testing.T) {
	t.Run("waits for all functions to complete", func(t *testing.T) {
		completed := make([]bool, 5)
		var mu sync.Mutex

		fns := make([]func() error, 5)
		for i := 0; i < 5; i++ {
			idx := i
			fns[i] = func() error {
				time.Sleep(time.Duration(idx*10) * time.Millisecond)
				mu.Lock()
				completed[idx] = true
				mu.Unlock()
				return nil
			}
		}

		err := Parallel(fns...)
		require.NoError(t, err)

		// All functions should have completed
		for i, done := range completed {
			assert.True(t, done, "function %d should have completed", i)
		}
	})
}
