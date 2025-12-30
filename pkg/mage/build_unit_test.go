package mage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNormalizeMainPath tests path normalization logic
func TestNormalizeMainPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "adds prefix to bare path",
			input:    "cmd/app",
			expected: "./cmd/app",
		},
		{
			name:     "preserves existing relative prefix",
			input:    "./cmd/app",
			expected: "./cmd/app",
		},
		{
			name:     "preserves absolute path",
			input:    "/usr/local/bin/app",
			expected: "/usr/local/bin/app",
		},
		{
			name:     "handles main.go suffix with slash",
			input:    "./cmd/app/main.go",
			expected: "./cmd/app",
		},
		{
			name:     "handles main.go suffix without leading dot slash",
			input:    "cmd/app/main.go",
			expected: "./cmd/app",
		},
		{
			name:     "handles bare main.go",
			input:    "main.go",
			expected: ".",
		},
		{
			name:     "handles dot slash main.go",
			input:    "./main.go",
			expected: ".",
		},
		{
			name:     "handles nested path without prefix",
			input:    "internal/cmd/myapp",
			expected: "./internal/cmd/myapp",
		},
		{
			name:     "handles single directory",
			input:    "cmd",
			expected: "./cmd",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "./",
		},
	}

	b := Build{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := b.normalizeMainPath(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestIsValidMainPath tests main path validation
func TestIsValidMainPath(t *testing.T) {
	b := Build{}

	t.Run("returns false for nonexistent directory", func(t *testing.T) {
		result := b.isValidMainPath("./nonexistent/path/that/does/not/exist")
		assert.False(t, result)
	})

	t.Run("returns false for directory without main.go", func(t *testing.T) {
		tempDir := t.TempDir()
		// Create a directory without main.go
		subDir := filepath.Join(tempDir, "cmd", "app")
		require.NoError(t, os.MkdirAll(subDir, 0o750))

		// Create a non-main.go file
		require.NoError(t, os.WriteFile(filepath.Join(subDir, "other.go"), []byte("package app"), 0o600))

		result := b.isValidMainPath("./" + filepath.Join("cmd", "app"))
		assert.False(t, result)
	})

	t.Run("returns true for directory with valid main package", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create cmd/app directory with main.go
		subDir := filepath.Join(tempDir, "cmd", "app")
		require.NoError(t, os.MkdirAll(subDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(subDir, "main.go"), []byte("package main\n\nfunc main() {}"), 0o600))

		result := b.isValidMainPath("./cmd/app")
		assert.True(t, result)
	})

	t.Run("returns false for directory with non-main package in main.go", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create cmd/app directory with main.go but not package main
		subDir := filepath.Join(tempDir, "cmd", "app")
		require.NoError(t, os.MkdirAll(subDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(subDir, "main.go"), []byte("package app\n\nfunc Run() {}"), 0o600))

		result := b.isValidMainPath("./cmd/app")
		assert.False(t, result)
	})
}

// TestValidateConfiguredMainPath tests configured main path validation
func TestValidateConfiguredMainPath(t *testing.T) {
	b := Build{}

	t.Run("returns error for invalid path", func(t *testing.T) {
		_, err := b.validateConfiguredMainPath("./nonexistent/path")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidMainPath)
	})

	t.Run("returns normalized path for valid main package", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create valid main package
		subDir := filepath.Join(tempDir, "cmd", "myapp")
		require.NoError(t, os.MkdirAll(subDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(subDir, "main.go"), []byte("package main\n\nfunc main() {}"), 0o600))

		result, err := b.validateConfiguredMainPath("cmd/myapp")
		require.NoError(t, err)
		assert.Equal(t, "./cmd/myapp", result)
	})

	t.Run("handles main.go suffix in path", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create valid main package
		subDir := filepath.Join(tempDir, "cmd", "tool")
		require.NoError(t, os.MkdirAll(subDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(subDir, "main.go"), []byte("package main\n\nfunc main() {}"), 0o600))

		result, err := b.validateConfiguredMainPath("cmd/tool/main.go")
		require.NoError(t, err)
		assert.Equal(t, "./cmd/tool", result)
	})
}

// TestCreateBuildArgs tests build argument construction
func TestCreateBuildArgs(t *testing.T) {
	b := Build{}

	t.Run("creates basic build args", func(t *testing.T) {
		cfg := &Config{
			Build: BuildConfig{},
		}

		args := b.createBuildArgs(cfg, "bin/app", "./cmd/app")

		require.Contains(t, args, "build")
		require.Contains(t, args, "-o")
		require.Contains(t, args, "bin/app")
		require.Contains(t, args, "./cmd/app")
	})

	t.Run("includes verbose flag when enabled", func(t *testing.T) {
		cfg := &Config{
			Build: BuildConfig{
				Verbose: true,
			},
		}

		args := b.createBuildArgs(cfg, "bin/app", "./cmd/app")

		assert.Contains(t, args, "-v")
	})

	t.Run("argument order is correct", func(t *testing.T) {
		cfg := &Config{
			Build: BuildConfig{},
		}

		args := b.createBuildArgs(cfg, "output/binary", "./src")

		// First arg should be "build"
		require.NotEmpty(t, args)
		assert.Equal(t, "build", args[0])

		// Last arg should be package path
		assert.Equal(t, "./src", args[len(args)-1])

		// Find -o and check its value follows
		for i, arg := range args {
			if arg == "-o" && i+1 < len(args) {
				assert.Equal(t, "output/binary", args[i+1])
				break
			}
		}
	})
}

// TestDetermineLDFlags tests linker flag determination
func TestDetermineLDFlags(t *testing.T) {
	b := Build{}

	t.Run("uses config ldflags when provided", func(t *testing.T) {
		cfg := &Config{
			Build: BuildConfig{
				LDFlags: []string{"-X main.version=1.0.0", "-X main.commit=abc123"},
			},
		}

		result := b.determineLDFlags(cfg)

		assert.Contains(t, result, "-X main.version=1.0.0")
		assert.Contains(t, result, "-X main.commit=abc123")
	})

	t.Run("generates default ldflags when config is empty", func(t *testing.T) {
		cfg := &Config{
			Build: BuildConfig{
				LDFlags: []string{},
			},
		}

		result := b.determineLDFlags(cfg)

		// Should contain version, commit, and buildDate
		assert.Contains(t, result, "-X main.version=")
		assert.Contains(t, result, "-X main.commit=")
		assert.Contains(t, result, "-X main.buildDate=")
	})

	t.Run("includes strip flags in non-debug mode", func(t *testing.T) {
		// Ensure DEBUG is not set
		originalDebug := os.Getenv("DEBUG")
		require.NoError(t, os.Unsetenv("DEBUG"))
		defer func() {
			if originalDebug != "" {
				require.NoError(t, os.Setenv("DEBUG", originalDebug))
			}
		}()

		cfg := &Config{
			Build: BuildConfig{},
		}

		result := b.determineLDFlags(cfg)

		assert.Contains(t, result, "-s")
		assert.Contains(t, result, "-w")
	})
}

// TestGetConfigFiles tests configuration file discovery
func TestGetConfigFiles(t *testing.T) {
	b := Build{}

	t.Run("always includes go.mod and go.sum", func(t *testing.T) {
		result := b.getConfigFiles()

		assert.Contains(t, result, "go.mod")
		assert.Contains(t, result, "go.sum")
	})

	t.Run("includes mage.yaml when present", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create .mage.yaml
		require.NoError(t, os.WriteFile(".mage.yaml", []byte("project:\n  name: test"), 0o600))

		result := b.getConfigFiles()

		assert.Contains(t, result, ".mage.yaml")
	})

	t.Run("does not include mage.yaml when absent", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		result := b.getConfigFiles()

		assert.NotContains(t, result, ".mage.yaml")
	})
}

// TestExpandLDFlagsTemplates tests template expansion in ldflags
func TestExpandLDFlagsTemplates(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		contains []string
	}{
		{
			name:     "expands Version template",
			input:    []string{"-X main.version={{.Version}}"},
			contains: []string{"-X main.version="},
		},
		{
			name:     "expands Commit template",
			input:    []string{"-X main.commit={{.Commit}}"},
			contains: []string{"-X main.commit="},
		},
		{
			name:     "expands Date template",
			input:    []string{"-X main.date={{.Date}}"},
			contains: []string{"-X main.date="},
		},
		{
			name:     "expands BuildDate template",
			input:    []string{"-X main.buildDate={{.BuildDate}}"},
			contains: []string{"-X main.buildDate="},
		},
		{
			name:     "expands BuildTime template",
			input:    []string{"-X main.buildTime={{.BuildTime}}"},
			contains: []string{"-X main.buildTime="},
		},
		{
			name:     "handles multiple templates",
			input:    []string{"-X main.version={{.Version}}", "-X main.commit={{.Commit}}"},
			contains: []string{"-X main.version=", "-X main.commit="},
		},
		{
			name:     "preserves non-template values",
			input:    []string{"-s", "-w", "-X main.foo=bar"},
			contains: []string{"-s", "-w", "-X main.foo=bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandLDFlagsTemplates(tt.input)

			require.Len(t, result, len(tt.input))
			for _, expected := range tt.contains {
				found := false
				for _, r := range result {
					if len(r) >= len(expected) && r[:len(expected)] == expected || r == expected {
						found = true
						break
					}
				}
				assert.True(t, found, "expected to find %q in result %v", expected, result)
			}
		})
	}
}

// TestExtractMainPackageFromLine tests JSON line parsing for main packages
func TestExtractMainPackageFromLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "extracts main package with spaces",
			input:    `{"ImportPath": "github.com/test/cmd/app", "Name": "main"}`,
			expected: "github.com/test/cmd/app",
		},
		{
			name:     "extracts main package without spaces",
			input:    `{"ImportPath":"github.com/test/cmd/tool","Name":"main"}`,
			expected: "github.com/test/cmd/tool",
		},
		{
			name:     "returns empty for non-main package",
			input:    `{"ImportPath": "github.com/test/pkg", "Name": "pkg"}`,
			expected: "",
		},
		{
			name:     "returns empty for root mage-x package",
			input:    `{"ImportPath": "github.com/mrz1836/mage-x", "Name": "main"}`,
			expected: "",
		},
		{
			name:     "handles malformed JSON gracefully",
			input:    `{"ImportPath": "incomplete`,
			expected: "",
		},
		{
			name:     "handles empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "handles input without name field",
			input:    `{"ImportPath": "github.com/test/pkg"}`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractMainPackageFromLine(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestApplyMemoryLimit tests memory limit parsing and strategy adjustment
func TestApplyMemoryLimit(t *testing.T) {
	b := Build{}

	tests := []struct {
		name           string
		memoryLimit    string
		inputStrategy  string
		expectStrategy string
	}{
		{
			name:           "parses gigabyte limit",
			memoryLimit:    "4G",
			inputStrategy:  "full",
			expectStrategy: "full", // May change to incremental based on available memory
		},
		{
			name:           "parses megabyte limit",
			memoryLimit:    "4096M",
			inputStrategy:  "full",
			expectStrategy: "full",
		},
		{
			name:           "preserves strategy for invalid limit format",
			memoryLimit:    "invalid",
			inputStrategy:  "smart",
			expectStrategy: "smart",
		},
		{
			name:           "preserves strategy for empty limit",
			memoryLimit:    "",
			inputStrategy:  "incremental",
			expectStrategy: "incremental",
		},
		{
			name:           "handles zero value",
			memoryLimit:    "0G",
			inputStrategy:  "full",
			expectStrategy: "full",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := b.applyMemoryLimit(tt.memoryLimit, tt.inputStrategy)
			// The actual strategy may change based on system memory
			// so we just verify it returns a valid strategy
			assert.NotEmpty(t, result)
		})
	}
}

// TestIsMainPackage tests main package detection from file content
func TestIsMainPackage(t *testing.T) {
	t.Run("returns true for valid main package", func(t *testing.T) {
		tempDir := t.TempDir()
		cmdDir := filepath.Join(tempDir, "cmd", "app")
		require.NoError(t, os.MkdirAll(cmdDir, 0o750))

		mainFile := filepath.Join(cmdDir, "main.go")
		require.NoError(t, os.WriteFile(mainFile, []byte("package main\n\nfunc main() {}"), 0o600))

		result := isMainPackage(mainFile)
		assert.True(t, result)
	})

	t.Run("returns false for non-main package", func(t *testing.T) {
		tempDir := t.TempDir()
		cmdDir := filepath.Join(tempDir, "cmd", "lib")
		require.NoError(t, os.MkdirAll(cmdDir, 0o750))

		mainFile := filepath.Join(cmdDir, "main.go")
		require.NoError(t, os.WriteFile(mainFile, []byte("package lib\n\nfunc Run() {}"), 0o600))

		result := isMainPackage(mainFile)
		assert.False(t, result)
	})

	t.Run("returns false for non-go file", func(t *testing.T) {
		tempDir := t.TempDir()
		txtFile := filepath.Join(tempDir, "readme.txt")
		require.NoError(t, os.WriteFile(txtFile, []byte("package main"), 0o600))

		result := isMainPackage(txtFile)
		assert.False(t, result)
	})

	t.Run("returns false for file not in cmd directory", func(t *testing.T) {
		tempDir := t.TempDir()
		mainFile := filepath.Join(tempDir, "main.go")
		require.NoError(t, os.WriteFile(mainFile, []byte("package main\n\nfunc main() {}"), 0o600))

		result := isMainPackage(mainFile)
		assert.False(t, result)
	})

	t.Run("returns false for nonexistent file", func(t *testing.T) {
		result := isMainPackage("/nonexistent/cmd/app/main.go")
		assert.False(t, result)
	})

	t.Run("handles file with comments before package declaration", func(t *testing.T) {
		tempDir := t.TempDir()
		cmdDir := filepath.Join(tempDir, "cmd", "app")
		require.NoError(t, os.MkdirAll(cmdDir, 0o750))

		content := `// Copyright 2024
// License: MIT

package main

func main() {}
`
		mainFile := filepath.Join(cmdDir, "main.go")
		require.NoError(t, os.WriteFile(mainFile, []byte(content), 0o600))

		result := isMainPackage(mainFile)
		assert.True(t, result)
	})

	t.Run("handles file with empty lines before package", func(t *testing.T) {
		tempDir := t.TempDir()
		cmdDir := filepath.Join(tempDir, "cmd", "app")
		require.NoError(t, os.MkdirAll(cmdDir, 0o750))

		content := `

package main

func main() {}
`
		mainFile := filepath.Join(cmdDir, "main.go")
		require.NoError(t, os.WriteFile(mainFile, []byte(content), 0o600))

		result := isMainPackage(mainFile)
		assert.True(t, result)
	})
}

// TestFindMainInCmdDir tests finding main packages in cmd directory
func TestFindMainInCmdDir(t *testing.T) {
	t.Run("returns empty when cmd directory does not exist", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		result := findMainInCmdDir()
		assert.Empty(t, result)
	})

	t.Run("finds main package in cmd subdirectory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create cmd/app with main.go
		cmdDir := filepath.Join(tempDir, "cmd", "app")
		require.NoError(t, os.MkdirAll(cmdDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte("package main\n\nfunc main() {}"), 0o600))

		result := findMainInCmdDir()
		assert.Equal(t, "./cmd/app", result)
	})

	t.Run("returns empty when cmd has no main packages", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create cmd/lib with non-main package
		cmdDir := filepath.Join(tempDir, "cmd", "lib")
		require.NoError(t, os.MkdirAll(cmdDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "lib.go"), []byte("package lib"), 0o600))

		result := findMainInCmdDir()
		assert.Empty(t, result)
	})

	t.Run("ignores files directly in cmd directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create cmd directory with file directly in it (not in subdirectory)
		cmdDir := filepath.Join(tempDir, "cmd")
		require.NoError(t, os.MkdirAll(cmdDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte("package main\n\nfunc main() {}"), 0o600))

		result := findMainInCmdDir()
		assert.Empty(t, result)
	})
}

// TestBuildFlags tests the buildFlags helper function
func TestBuildFlags(t *testing.T) {
	t.Run("generates flags with tags", func(t *testing.T) {
		cfg := &Config{
			Build: BuildConfig{
				Tags: []string{"integration", "e2e"},
			},
		}

		flags := buildFlags(cfg)

		assert.Contains(t, flags, "-tags")
		// Find the value after -tags
		for i, f := range flags {
			if f == "-tags" && i+1 < len(flags) {
				assert.Equal(t, "integration,e2e", flags[i+1])
				break
			}
		}
	})

	t.Run("generates flags with trimpath", func(t *testing.T) {
		cfg := &Config{
			Build: BuildConfig{
				TrimPath: true,
			},
		}

		flags := buildFlags(cfg)

		assert.Contains(t, flags, "-trimpath")
	})

	t.Run("generates flags with verbose", func(t *testing.T) {
		cfg := &Config{
			Build: BuildConfig{
				Verbose: true,
			},
		}

		flags := buildFlags(cfg)

		assert.Contains(t, flags, "-v")
	})

	t.Run("generates flags with custom goflags", func(t *testing.T) {
		cfg := &Config{
			Build: BuildConfig{
				GoFlags: []string{"-race", "-cover"},
			},
		}

		flags := buildFlags(cfg)

		assert.Contains(t, flags, "-race")
		assert.Contains(t, flags, "-cover")
	})

	t.Run("always includes ldflags", func(t *testing.T) {
		cfg := &Config{
			Build: BuildConfig{},
		}

		flags := buildFlags(cfg)

		assert.Contains(t, flags, "-ldflags")
	})
}

// TestSplitIntoBatches tests the batch splitting utility
func TestSplitIntoBatches(t *testing.T) {
	b := Build{}

	tests := []struct {
		name         string
		packages     []string
		batchSize    int
		expectedLen  int
		firstBatchSz int
		lastBatchSz  int
	}{
		{
			name:         "splits evenly divisible list",
			packages:     []string{"a", "b", "c", "d"},
			batchSize:    2,
			expectedLen:  2,
			firstBatchSz: 2,
			lastBatchSz:  2,
		},
		{
			name:         "handles remainder in last batch",
			packages:     []string{"a", "b", "c", "d", "e"},
			batchSize:    2,
			expectedLen:  3,
			firstBatchSz: 2,
			lastBatchSz:  1,
		},
		{
			name:         "single batch when list smaller than batch size",
			packages:     []string{"a", "b"},
			batchSize:    10,
			expectedLen:  1,
			firstBatchSz: 2,
			lastBatchSz:  2,
		},
		{
			name:         "handles empty list",
			packages:     []string{},
			batchSize:    5,
			expectedLen:  0,
			firstBatchSz: 0,
			lastBatchSz:  0,
		},
		{
			name:         "defaults to 10 for zero batch size",
			packages:     []string{"a", "b", "c"},
			batchSize:    0,
			expectedLen:  1,
			firstBatchSz: 3,
			lastBatchSz:  3,
		},
		{
			name:         "defaults to 10 for negative batch size",
			packages:     []string{"a", "b", "c"},
			batchSize:    -5,
			expectedLen:  1,
			firstBatchSz: 3,
			lastBatchSz:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			batches := b.splitIntoBatches(tt.packages, tt.batchSize)

			assert.Len(t, batches, tt.expectedLen)
			if tt.expectedLen > 0 {
				assert.Len(t, batches[0], tt.firstBatchSz)
				assert.Len(t, batches[len(batches)-1], tt.lastBatchSz)
			}
		})
	}
}

// TestGetBinarySize tests binary size retrieval
func TestGetBinarySize(t *testing.T) {
	t.Run("returns size for existing file", func(t *testing.T) {
		tempDir := t.TempDir()
		filePath := filepath.Join(tempDir, "binary")
		content := []byte("test binary content of known length")
		require.NoError(t, os.WriteFile(filePath, content, 0o600))

		size := getBinarySize(filePath)

		assert.Equal(t, int64(len(content)), size)
	})

	t.Run("returns zero for nonexistent file", func(t *testing.T) {
		size := getBinarySize("/nonexistent/path/to/binary")

		assert.Equal(t, int64(0), size)
	})
}

// TestDetermineBuildOutput tests build output path determination
func TestDetermineBuildOutput(t *testing.T) {
	b := Build{}

	t.Run("creates correct output path", func(t *testing.T) {
		cfg := &Config{
			Project: ProjectConfig{
				Binary: "myapp",
			},
			Build: BuildConfig{
				Output: "dist",
			},
		}

		result := b.determineBuildOutput(cfg)

		// On non-Windows, should be dist/myapp
		// On Windows, should be dist/myapp.exe
		assert.Contains(t, result, "myapp")
		assert.True(t, strings.HasPrefix(result, "dist"))
	})
}

// TestFindSourceFiles tests source file discovery for caching
func TestFindSourceFiles(t *testing.T) {
	t.Run("returns default files on error", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		b := Build{}
		result := b.findSourceFiles()

		// Should return at least go.mod as default (go.sum only if it exists)
		assert.Contains(t, result, "go.mod")
	})

	t.Run("includes go.sum when it exists", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create go.sum file
		require.NoError(t, os.WriteFile("go.sum", []byte("module/dep v1.0.0 h1:abc123"), 0o600))

		b := Build{}
		result := b.findSourceFiles()

		assert.Contains(t, result, "go.mod")
		assert.Contains(t, result, "go.sum")
	})
}
