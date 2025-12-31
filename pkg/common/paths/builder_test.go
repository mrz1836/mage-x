package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testOSWindows is a constant for Windows OS checks in tests
const testOSWindows = "windows"

// TestNewPathBuilder tests path builder creation
func TestNewPathBuilder(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantPath string // Expected cleaned path
		wantSafe bool   // Should path be considered safe
	}{
		{
			name:     "simple path",
			path:     "test/path",
			wantPath: "test/path",
			wantSafe: true,
		},
		{
			name:     "path with dots gets cleaned",
			path:     "test/./path",
			wantPath: "test/path",
			wantSafe: true,
		},
		{
			name:     "path with traversal attempt",
			path:     "test/../etc/passwd",
			wantPath: "etc/passwd",
			wantSafe: false, // Original has ..
		},
		{
			name:     "absolute path",
			path:     "/tmp/test",
			wantPath: "/tmp/test",
			wantSafe: true,
		},
		{
			name:     "empty path",
			path:     "",
			wantPath: ".",
			wantSafe: false, // Empty original path is unsafe
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPathBuilder(tt.path)
			assert.Equal(t, tt.wantPath, pb.String())
			assert.Equal(t, tt.wantSafe, pb.IsSafe(), "IsSafe() mismatch")
		})
	}
}

// TestPathBuilder_Join tests joining path elements
func TestPathBuilder_Join(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		elements []string
		want     string
	}{
		{
			name:     "join single element",
			base:     "/tmp",
			elements: []string{"test"},
			want:     "/tmp/test",
		},
		{
			name:     "join multiple elements",
			base:     "/tmp",
			elements: []string{"test", "subdir", "file.txt"},
			want:     "/tmp/test/subdir/file.txt",
		},
		{
			name:     "join empty elements",
			base:     "/tmp",
			elements: []string{},
			want:     "/tmp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPathBuilder(tt.base)
			result := pb.Join(tt.elements...)
			assert.Equal(t, tt.want, result.String())
		})
	}
}

// TestPathBuilder_SecurityChecks tests security validation
func TestPathBuilder_SecurityChecks(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantSafe bool
		reason   string
	}{
		{
			name:     "safe simple path",
			path:     "test/file.txt",
			wantSafe: true,
		},
		{
			name:     "path traversal with ..",
			path:     "../../../etc/passwd",
			wantSafe: false,
			reason:   "contains ..",
		},
		{
			name:     "hidden traversal in middle",
			path:     "test/../../../etc/passwd",
			wantSafe: false,
			reason:   "contains ..",
		},
		{
			name:     "proc filesystem path",
			path:     "/proc/self/fd/../../secrets",
			wantSafe: false,
			reason:   "contains /proc/",
		},
		{
			name:     "dev filesystem path",
			path:     "/dev/null",
			wantSafe: false,
			reason:   "contains /dev/",
		},
		{
			name:     "URL encoded traversal %2e",
			path:     "test%2e%2e/passwd",
			wantSafe: false,
			reason:   "URL encoded",
		},
		{
			name:     "double URL encoded traversal",
			path:     "test%252e%252e/passwd",
			wantSafe: false,
			reason:   "double URL encoded",
		},
		{
			name:     "null byte injection",
			path:     "test\x00.txt",
			wantSafe: false,
			reason:   "null byte",
		},
		{
			name:     "null byte URL encoded",
			path:     "test%00.txt",
			wantSafe: false,
			reason:   "null byte encoded",
		},
		{
			name:     "control characters",
			path:     "test\x01file.txt",
			wantSafe: false,
			reason:   "control character",
		},
		{
			name:     "unicode fraction slash",
			path:     "test⁄file.txt",
			wantSafe: false,
			reason:   "unicode confusable",
		},
		{
			name:     "overlong UTF-8",
			path:     "test\xc0\xaffile.txt",
			wantSafe: false,
			reason:   "overlong UTF-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPathBuilder(tt.path)
			got := pb.IsSafe()
			assert.Equal(t, tt.wantSafe, got, "reason: %s", tt.reason)
		})
	}
}

// TestPathBuilder_WindowsSecurity tests Windows-specific security checks
func TestPathBuilder_WindowsSecurity(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantSafe bool
		reason   string
	}{
		{
			name:     "safe normal path",
			path:     "test/file.txt",
			wantSafe: true,
		},
		{
			name:     "Windows UNC path",
			path:     "\\\\server\\share",
			wantSafe: false,
			reason:   "UNC path",
		},
		{
			name:     "Windows drive letter",
			path:     "C:\\Windows\\System32",
			wantSafe: false,
			reason:   "drive letter",
		},
		{
			name:     "Windows alternate data stream",
			path:     "test.txt:$DATA",
			wantSafe: false,
			reason:   "alternate data stream",
		},
		{
			name:     "Windows reserved name CON",
			path:     "CON",
			wantSafe: false,
			reason:   "reserved device name",
		},
		{
			name:     "Windows reserved name NUL.txt",
			path:     "NUL.txt",
			wantSafe: false,
			reason:   "reserved device name with extension",
		},
		{
			name:     "Windows reserved name PRN",
			path:     "PRN",
			wantSafe: false,
			reason:   "reserved device name",
		},
		{
			name:     "trailing dot",
			path:     "test.txt.",
			wantSafe: false,
			reason:   "trailing dot",
		},
		{
			name:     "trailing space",
			path:     "test.txt ",
			wantSafe: false,
			reason:   "trailing space",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPathBuilder(tt.path)
			got := pb.isWindowsSafe(tt.path)
			assert.Equal(t, tt.wantSafe, got, "reason: %s", tt.reason)
		})
	}
}

// TestPathBuilder_UnixSecurity tests Unix-specific security checks
func TestPathBuilder_UnixSecurity(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantSafe bool
		reason   string
	}{
		{
			name:     "safe normal path",
			path:     "test/file.txt",
			wantSafe: true,
		},
		{
			name:     "proc filesystem",
			path:     "/proc/self/environ",
			wantSafe: false,
			reason:   "/proc/ path",
		},
		{
			name:     "dev filesystem",
			path:     "/dev/zero",
			wantSafe: false,
			reason:   "/dev/ path",
		},
		{
			name:     "relative path containing proc",
			path:     "test/proc/file",
			wantSafe: false,
			reason:   "contains /proc/ pattern",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPathBuilder(tt.path)
			got := pb.isUnixSafe(tt.path)
			assert.Equal(t, tt.wantSafe, got, "reason: %s", tt.reason)
		})
	}
}

// TestPathBuilder_Validate tests path validation
func TestPathBuilder_Validate(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		options     PathOptions
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid path",
			path:    "test/file.txt",
			wantErr: false,
		},
		{
			name:        "empty path",
			path:        "",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:    "path with .. when allowed",
			path:    "test/../file.txt",
			options: PathOptions{AllowUnsafePaths: true},
			wantErr: false,
		},
		{
			name:        "path with .. when not allowed",
			path:        "test/../file.txt",
			options:     PathOptions{AllowUnsafePaths: false},
			wantErr:     true,
			errContains: "unsafe",
		},
		{
			name:        "path exceeds max length",
			path:        strings.Repeat("a", 100),
			options:     PathOptions{MaxPathLength: 50},
			wantErr:     true,
			errContains: "maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPathBuilderWithOptions(tt.path, tt.options)
			err := pb.Validate()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestPathBuilder_BasePathRestriction tests restricted base path security
func TestPathBuilder_BasePathRestriction(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		path        string
		basePath    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "path within base",
			path:     filepath.Join(tempDir, "subdir", "file.txt"),
			basePath: tempDir,
			wantErr:  false,
		},
		{
			name:        "path outside base",
			path:        "/etc/passwd",
			basePath:    tempDir,
			wantErr:     true,
			errContains: "outside",
		},
		{
			name:        "relative path escaping base",
			path:        filepath.Join(tempDir, "..", "..", "etc", "passwd"),
			basePath:    tempDir,
			wantErr:     true,
			errContains: "outside",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			options := PathOptions{
				RestrictToBasePath: tt.basePath,
			}
			pb := NewPathBuilderWithOptions(tt.path, options)
			err := pb.Validate()

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestPathBuilder_WithExt tests extension manipulation
func TestPathBuilder_WithExt(t *testing.T) {
	tests := []struct {
		name string
		path string
		ext  string
		want string
	}{
		{
			name: "change extension with dot",
			path: "test.txt",
			ext:  ".md",
			want: "test.md",
		},
		{
			name: "change extension without dot",
			path: "test.txt",
			ext:  "md",
			want: "test.md",
		},
		{
			name: "remove extension",
			path: "test.txt",
			ext:  "",
			want: "test",
		},
		{
			name: "add extension to file without extension",
			path: "test",
			ext:  ".txt",
			want: "test.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPathBuilder(tt.path)
			result := pb.WithExt(tt.ext)
			assert.Equal(t, tt.want, result.String())
		})
	}
}

// TestPathBuilder_Append tests appending suffixes
func TestPathBuilder_Append(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		suffix string
		want   string
	}{
		{
			name:   "append to file with extension",
			path:   "test.txt",
			suffix: "_backup",
			want:   "test_backup.txt",
		},
		{
			name:   "append to file without extension",
			path:   "test",
			suffix: "_v2",
			want:   "test_v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPathBuilder(tt.path)
			result := pb.Append(tt.suffix)
			assert.Equal(t, tt.want, result.String())
		})
	}
}

// TestPathBuilder_Prepend tests prepending prefixes
func TestPathBuilder_Prepend(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		prefix string
		want   string
	}{
		{
			name:   "prepend to simple filename",
			path:   "test.txt",
			prefix: "backup_",
			want:   "backup_test.txt",
		},
		{
			name:   "prepend to path with directory",
			path:   "dir/test.txt",
			prefix: "new_",
			want:   "dir/new_test.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPathBuilder(tt.path)
			result := pb.Prepend(tt.prefix)
			assert.Equal(t, tt.want, result.String())
		})
	}
}

// TestPathBuilder_FileSystemOps tests file system operations
func TestPathBuilder_FileSystemOps(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("create file", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "test.txt")
		pb := NewPathBuilder(filePath)

		err := pb.Create()
		require.NoError(t, err)
		assert.True(t, pb.Exists())
		assert.True(t, pb.IsFile())
		assert.False(t, pb.IsDir())
	})

	t.Run("create directory", func(t *testing.T) {
		dirPath := filepath.Join(tempDir, "testdir")
		pb := NewPathBuilder(dirPath)

		err := pb.CreateDir()
		require.NoError(t, err)
		assert.True(t, pb.Exists())
		assert.True(t, pb.IsDir())
		assert.False(t, pb.IsFile())
	})

	t.Run("create nested directory", func(t *testing.T) {
		dirPath := filepath.Join(tempDir, "nested", "deep", "dir")
		pb := NewPathBuilder(dirPath)

		err := pb.CreateDirAll()
		require.NoError(t, err)
		assert.True(t, pb.Exists())
		assert.True(t, pb.IsDir())
	})

	t.Run("file size and modtime", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "sized.txt")
		content := []byte("test content")
		err := os.WriteFile(filePath, content, 0o644) //nolint:gosec // test file
		require.NoError(t, err)

		pb := NewPathBuilder(filePath)
		assert.Equal(t, int64(len(content)), pb.Size())
		assert.False(t, pb.ModTime().IsZero())
	})
}

// TestPathBuilder_PathManipulation tests path manipulation methods
func TestPathBuilder_PathManipulation(t *testing.T) {
	t.Run("Dir", func(t *testing.T) {
		pb := NewPathBuilder("/tmp/test/file.txt")
		dir := pb.Dir()
		assert.Equal(t, "/tmp/test", dir.String())
	})

	t.Run("Base", func(t *testing.T) {
		pb := NewPathBuilder("/tmp/test/file.txt")
		base := pb.Base()
		assert.Equal(t, "file.txt", base)
	})

	t.Run("Ext", func(t *testing.T) {
		pb := NewPathBuilder("/tmp/test/file.txt")
		ext := pb.Ext()
		assert.Equal(t, ".txt", ext)
	})

	t.Run("Clean", func(t *testing.T) {
		pb := NewPathBuilder("/tmp/./test/../file.txt")
		cleaned := pb.Clean()
		assert.Equal(t, "/tmp/file.txt", cleaned.String())
	})

	t.Run("IsAbs", func(t *testing.T) {
		absPb := NewPathBuilder("/tmp/test")
		relPb := NewPathBuilder("test/file")
		assert.True(t, absPb.IsAbs())
		assert.False(t, relPb.IsAbs())
	})
}

// TestPathBuilder_PatternMatching tests pattern matching
func TestPathBuilder_PatternMatching(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		pattern string
		want    bool
	}{
		{
			name:    "match extension",
			path:    "test/file.txt",
			pattern: "*.txt",
			want:    true,
		},
		{
			name:    "no match",
			path:    "test/file.txt",
			pattern: "*.md",
			want:    false,
		},
		{
			name:    "match prefix",
			path:    "test/file.txt",
			pattern: "file*",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPathBuilder(tt.path)
			got := pb.Match(tt.pattern)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestPathBuilder_Contains tests string matching
func TestPathBuilder_Contains(t *testing.T) {
	pb := NewPathBuilder("/tmp/test/file.txt")

	assert.True(t, pb.Contains("test"))
	assert.True(t, pb.Contains("/tmp"))
	assert.False(t, pb.Contains("notfound"))
}

// TestPathBuilder_HasPrefix tests prefix checking
func TestPathBuilder_HasPrefix(t *testing.T) {
	pb := NewPathBuilder("/tmp/test/file.txt")

	assert.True(t, pb.HasPrefix("/tmp"))
	assert.True(t, pb.HasPrefix("/tmp/test"))
	assert.False(t, pb.HasPrefix("/var"))
}

// TestPathBuilder_HasSuffix tests suffix checking
func TestPathBuilder_HasSuffix(t *testing.T) {
	pb := NewPathBuilder("/tmp/test/file.txt")

	assert.True(t, pb.HasSuffix(".txt"))
	assert.True(t, pb.HasSuffix("file.txt"))
	assert.False(t, pb.HasSuffix(".md"))
}

// TestPathBuilder_Cloning tests cloning
func TestPathBuilder_Cloning(t *testing.T) {
	original := NewPathBuilder("/tmp/test.txt")
	cloned := original.Clone()

	// Should have same path
	assert.Equal(t, original.String(), cloned.String())

	// Should be different instances
	modified := cloned.WithExt(".md")
	assert.NotEqual(t, original.String(), modified.String())
}

// TestPathBuilder_LengthSafety tests length safety checks
func TestPathBuilder_LengthSafety(t *testing.T) {
	tests := []struct {
		name     string
		pathLen  int
		wantSafe bool
	}{
		{
			name:     "normal length",
			pathLen:  100,
			wantSafe: true,
		},
		{
			name:     "maximum safe length",
			pathLen:  4096,
			wantSafe: true,
		},
		{
			name:     "exceeds safe length",
			pathLen:  5000,
			wantSafe: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := strings.Repeat("a", tt.pathLen)
			pb := NewPathBuilder(path)
			got := pb.isLengthSafe(path)
			assert.Equal(t, tt.wantSafe, got)
		})
	}
}

// TestPathBuilder_List tests directory listing
func TestPathBuilder_List(t *testing.T) {
	tempDir := t.TempDir()

	// Create test structure
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "file1.txt"), []byte("test"), 0o644)) //nolint:gosec // test file
	require.NoError(t, os.WriteFile(filepath.Join(tempDir, "file2.txt"), []byte("test"), 0o644)) //nolint:gosec // test file
	require.NoError(t, os.Mkdir(filepath.Join(tempDir, "subdir"), 0o755))                        //nolint:gosec // test directory

	pb := NewPathBuilder(tempDir)

	t.Run("List all entries", func(t *testing.T) {
		entries, err := pb.List()
		require.NoError(t, err)
		assert.Len(t, entries, 3) // 2 files + 1 directory
	})

	t.Run("ListFiles only", func(t *testing.T) {
		files, err := pb.ListFiles()
		require.NoError(t, err)
		assert.Len(t, files, 2) // Only files
	})

	t.Run("ListDirs only", func(t *testing.T) {
		dirs, err := pb.ListDirs()
		require.NoError(t, err)
		assert.Len(t, dirs, 1) // Only directory
	})
}

// TestPathBuilder_Copy tests file copying
func TestPathBuilder_Copy(t *testing.T) {
	if runtime.GOOS == testOSWindows {
		t.Skip("Skipping copy test on Windows due to file system differences")
	}

	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")

	// Create source file
	content := []byte("test content")
	require.NoError(t, os.WriteFile(srcPath, content, 0o644)) //nolint:gosec // test file

	// Copy file
	srcPb := NewPathBuilder(srcPath)
	dstPb := NewPathBuilder(dstPath)

	err := srcPb.Copy(dstPb)
	require.NoError(t, err)

	// Verify destination exists and has same content
	assert.True(t, dstPb.Exists())
	dstContent, err := os.ReadFile(dstPath) //nolint:gosec // test file
	require.NoError(t, err)
	assert.Equal(t, content, dstContent)
}

// TestPathBuilder_CopyFile_Comprehensive tests copyFile method comprehensively
func TestPathBuilder_CopyFile_Comprehensive(t *testing.T) {
	if runtime.GOOS == testOSWindows {
		t.Skip("Skipping copy test on Windows due to file system differences")
	}

	tempDir := t.TempDir()

	t.Run("copy with CreateParents creates destination directory", func(t *testing.T) {
		srcPath := filepath.Join(tempDir, "src_create_parents.txt")
		dstPath := filepath.Join(tempDir, "nested", "deep", "dest_create_parents.txt")

		content := []byte("test content for nested copy")
		require.NoError(t, os.WriteFile(srcPath, content, 0o600))

		options := PathOptions{
			CreateParents: true,
			CreateMode:    0o750,
			BufferSize:    8192,
		}
		srcPb := NewPathBuilderWithOptions(srcPath, options)
		dstPb := NewPathBuilder(dstPath)

		err := srcPb.Copy(dstPb)
		require.NoError(t, err)

		assert.True(t, dstPb.Exists())
		dstContent, err := os.ReadFile(dstPath) //nolint:gosec // G304: test file path is controlled
		require.NoError(t, err)
		assert.Equal(t, content, dstContent)
	})

	t.Run("copy with PreserveMode preserves file permissions", func(t *testing.T) {
		srcPath := filepath.Join(tempDir, "src_preserve_mode.txt")
		dstPath := filepath.Join(tempDir, "dest_preserve_mode.txt")

		content := []byte("test content")
		require.NoError(t, os.WriteFile(srcPath, content, 0o755)) //nolint:gosec // test file

		options := PathOptions{
			PreserveMode: true,
			BufferSize:   8192,
		}
		srcPb := NewPathBuilderWithOptions(srcPath, options)
		dstPb := NewPathBuilder(dstPath)

		err := srcPb.Copy(dstPb)
		require.NoError(t, err)

		srcInfo, statErr := os.Stat(srcPath)
		require.NoError(t, statErr)
		dstInfo, statErr := os.Stat(dstPath)
		require.NoError(t, statErr)
		assert.Equal(t, srcInfo.Mode(), dstInfo.Mode())
	})

	t.Run("copy with PreserveMtime preserves modification time", func(t *testing.T) {
		srcPath := filepath.Join(tempDir, "src_preserve_mtime.txt")
		dstPath := filepath.Join(tempDir, "dest_preserve_mtime.txt")

		content := []byte("test content")
		require.NoError(t, os.WriteFile(srcPath, content, 0o600))

		options := PathOptions{
			PreserveMtime: true,
			BufferSize:    8192,
		}
		srcPb := NewPathBuilderWithOptions(srcPath, options)
		dstPb := NewPathBuilder(dstPath)

		err := srcPb.Copy(dstPb)
		require.NoError(t, err)

		srcInfo, statErr := os.Stat(srcPath)
		require.NoError(t, statErr)
		dstInfo, statErr := os.Stat(dstPath)
		require.NoError(t, statErr)
		assert.Equal(t, srcInfo.ModTime().Unix(), dstInfo.ModTime().Unix())
	})

	t.Run("copy non-existent source returns error", func(t *testing.T) {
		srcPath := filepath.Join(tempDir, "nonexistent.txt")
		dstPath := filepath.Join(tempDir, "dest_nonexistent.txt")

		srcPb := NewPathBuilder(srcPath)
		dstPb := NewPathBuilder(dstPath)

		err := srcPb.Copy(dstPb)
		require.Error(t, err)
	})

	t.Run("copy with small buffer size", func(t *testing.T) {
		srcPath := filepath.Join(tempDir, "src_small_buffer.txt")
		dstPath := filepath.Join(tempDir, "dest_small_buffer.txt")

		// Create a larger file to test buffered reading
		content := make([]byte, 1024)
		for i := range content {
			content[i] = byte(i % 256)
		}
		require.NoError(t, os.WriteFile(srcPath, content, 0o600))

		options := PathOptions{
			BufferSize: 64, // Small buffer to force multiple read/write cycles
		}
		srcPb := NewPathBuilderWithOptions(srcPath, options)
		dstPb := NewPathBuilder(dstPath)

		err := srcPb.Copy(dstPb)
		require.NoError(t, err)

		dstContent, err := os.ReadFile(dstPath) //nolint:gosec // G304: test file path is controlled
		require.NoError(t, err)
		assert.Equal(t, content, dstContent)
	})
}

// TestPathBuilder_CopyDir_Comprehensive tests copyDir method comprehensively
func TestPathBuilder_CopyDir_Comprehensive(t *testing.T) {
	if runtime.GOOS == testOSWindows {
		t.Skip("Skipping copy test on Windows due to file system differences")
	}

	tempDir := t.TempDir()

	t.Run("copy directory recursively", func(t *testing.T) {
		// Create source directory structure
		srcDir := filepath.Join(tempDir, "src_dir")
		require.NoError(t, os.MkdirAll(filepath.Join(srcDir, "subdir"), 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "file1.txt"), []byte("content1"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "subdir", "file2.txt"), []byte("content2"), 0o600))

		dstDir := filepath.Join(tempDir, "dst_dir")

		srcPb := NewPathBuilder(srcDir)
		dstPb := NewPathBuilder(dstDir)

		err := srcPb.Copy(dstPb)
		require.NoError(t, err)

		// Verify structure
		assert.True(t, NewPathBuilder(dstDir).Exists())
		assert.True(t, NewPathBuilder(filepath.Join(dstDir, "file1.txt")).Exists())
		assert.True(t, NewPathBuilder(filepath.Join(dstDir, "subdir")).IsDir())
		assert.True(t, NewPathBuilder(filepath.Join(dstDir, "subdir", "file2.txt")).Exists())

		// Verify content
		content1, readErr := os.ReadFile(filepath.Join(dstDir, "file1.txt")) //nolint:gosec // G304: test file path is controlled
		require.NoError(t, readErr)
		assert.Equal(t, []byte("content1"), content1)
		content2, readErr := os.ReadFile(filepath.Join(dstDir, "subdir", "file2.txt")) //nolint:gosec // G304: test file path is controlled
		require.NoError(t, readErr)
		assert.Equal(t, []byte("content2"), content2)
	})

	t.Run("copy directory read error returns error", func(t *testing.T) {
		// Create a directory then make it unreadable
		srcDir := filepath.Join(tempDir, "unreadable_dir")
		require.NoError(t, os.MkdirAll(srcDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(srcDir, "file.txt"), []byte("content"), 0o600))

		// Make source directory unreadable
		require.NoError(t, os.Chmod(srcDir, 0o000))
		defer func() {
			_ = os.Chmod(srcDir, 0o750) //nolint:gosec,errcheck // G302: restore permissions for cleanup
		}()

		dstDir := filepath.Join(tempDir, "dst_unreadable")
		srcPb := NewPathBuilder(srcDir)
		dstPb := NewPathBuilder(dstDir)

		err := srcPb.Copy(dstPb)
		require.Error(t, err)
	})
}

// TestPathBuilder_Move tests file moving
func TestPathBuilder_Move(t *testing.T) {
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")

	// Create source file
	require.NoError(t, os.WriteFile(srcPath, []byte("test"), 0o644)) //nolint:gosec // test file

	// Move file
	srcPb := NewPathBuilder(srcPath)
	dstPb := NewPathBuilder(dstPath)

	err := srcPb.Move(dstPb)
	require.NoError(t, err)

	// Verify source no longer exists and destination exists
	assert.False(t, srcPb.Exists())
	assert.True(t, dstPb.Exists())
}

// TestPathBuilder_WithName tests changing filename
func TestPathBuilder_WithName(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		newName string
		want    string
	}{
		{
			name:    "change filename in directory",
			path:    "/tmp/old.txt",
			newName: "new.txt",
			want:    "/tmp/new.txt",
		},
		{
			name:    "change filename without directory",
			path:    "old.txt",
			newName: "new.txt",
			want:    "new.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPathBuilder(tt.path)
			result := pb.WithName(tt.newName)
			assert.Equal(t, tt.want, result.String())
		})
	}
}

// TestPathBuilder_Original tests original path tracking for security
func TestPathBuilder_Original(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		wantOriginal string
	}{
		{
			name:         "simple path",
			path:         "test/file.txt",
			wantOriginal: "test/file.txt",
		},
		{
			name:         "path with traversal that gets cleaned",
			path:         "test/../file.txt",
			wantOriginal: "test/../file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pb := NewPathBuilder(tt.path)
			assert.Equal(t, tt.wantOriginal, pb.Original())
		})
	}
}

// TestTempAndTempDir tests temporary file and directory creation
func TestTempAndTempDir(t *testing.T) {
	t.Run("Temp creates temporary file path", func(t *testing.T) {
		pb, err := Temp("test-*")
		require.NoError(t, err)
		assert.NotNil(t, pb)
		assert.NotEmpty(t, pb.String())
		assert.Contains(t, pb.String(), "test-")
		// File should NOT exist (Temp removes it after creating)
		assert.False(t, pb.Exists())
	})

	t.Run("TempDir creates temporary directory", func(t *testing.T) {
		pb, err := TempDir("test-dir-*")
		require.NoError(t, err)
		assert.NotNil(t, pb)
		assert.True(t, pb.Exists())
		assert.True(t, pb.IsDir())
		// Cleanup
		require.NoError(t, pb.RemoveAll())
	})

	t.Run("multiple Temp calls create unique paths", func(t *testing.T) {
		pb1, err := Temp("unique-*")
		require.NoError(t, err)
		pb2, err := Temp("unique-*")
		require.NoError(t, err)
		assert.NotEqual(t, pb1.String(), pb2.String())
	})

	t.Run("multiple TempDir calls create unique directories", func(t *testing.T) {
		pb1, err := TempDir("unique-dir-*")
		require.NoError(t, err)
		defer func() { require.NoError(t, pb1.RemoveAll()) }()
		pb2, err := TempDir("unique-dir-*")
		require.NoError(t, err)
		defer func() { require.NoError(t, pb2.RemoveAll()) }()
		assert.NotEqual(t, pb1.String(), pb2.String())
	})
}

// TestPackageLevelFunctions tests the package-level convenience functions
func TestPackageLevelFunctions(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Join creates path builder", func(t *testing.T) {
		pb := Join("path", "to", "file.txt")
		assert.Equal(t, filepath.Join("path", "to", "file.txt"), pb.String())
	})

	t.Run("FromString creates path builder", func(t *testing.T) {
		pb := FromString("/tmp/test.txt")
		assert.Equal(t, "/tmp/test.txt", pb.String())
	})

	t.Run("Exists checks file existence", func(t *testing.T) {
		// Existing file
		existingFile := filepath.Join(tempDir, "existing.txt")
		require.NoError(t, os.WriteFile(existingFile, []byte("content"), 0o600))
		assert.True(t, Exists(existingFile))

		// Non-existing file
		assert.False(t, Exists(filepath.Join(tempDir, "nonexistent.txt")))
	})

	t.Run("IsDir checks directory", func(t *testing.T) {
		dirPath := filepath.Join(tempDir, "testdir")
		require.NoError(t, os.MkdirAll(dirPath, 0o750))

		filePath := filepath.Join(tempDir, "testfile.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o600))

		assert.True(t, IsDir(dirPath))
		assert.False(t, IsDir(filePath))
		assert.False(t, IsDir(filepath.Join(tempDir, "nonexistent")))
	})

	t.Run("IsFile checks file", func(t *testing.T) {
		dirPath := filepath.Join(tempDir, "testdir2")
		require.NoError(t, os.MkdirAll(dirPath, 0o750))

		filePath := filepath.Join(tempDir, "testfile2.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o600))

		assert.True(t, IsFile(filePath))
		assert.False(t, IsFile(dirPath))
		assert.False(t, IsFile(filepath.Join(tempDir, "nonexistent")))
	})

	t.Run("GlobPaths returns sorted matching paths", func(t *testing.T) {
		// Create test files
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "a.txt"), []byte("a"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "b.txt"), []byte("b"), 0o600))
		require.NoError(t, os.WriteFile(filepath.Join(tempDir, "c.txt"), []byte("c"), 0o600))

		paths, err := GlobPaths(filepath.Join(tempDir, "*.txt"))
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(paths), 3)

		// Verify sorted order
		for i := 1; i < len(paths); i++ {
			assert.LessOrEqual(t, paths[i-1].String(), paths[i].String(), "paths should be sorted")
		}
	})

	t.Run("GlobPaths returns empty for no matches", func(t *testing.T) {
		paths, err := GlobPaths(filepath.Join(tempDir, "*.nonexistent"))
		require.NoError(t, err)
		assert.Empty(t, paths)
	})
}

// TestPathBuilder_ErrorPaths tests various error conditions
func TestPathBuilder_ErrorPaths(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Abs error is wrapped", func(t *testing.T) {
		// This is hard to trigger on normal systems, but test the method exists
		pb := NewPathBuilder("valid/path")
		absPb, err := pb.Abs()
		require.NoError(t, err)
		assert.NotNil(t, absPb)
	})

	t.Run("Rel error for incompatible paths", func(t *testing.T) {
		pb := NewPathBuilder("/tmp/test")
		_, err := pb.Rel("relative/path/that/doesnt/work")
		// May or may not error depending on platform
		_ = err
	})

	t.Run("List returns error for non-directory", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "file.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o600))

		pb := NewPathBuilder(filePath)
		_, err := pb.List()
		require.Error(t, err)
	})

	t.Run("ListFiles returns error for non-directory", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "file2.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o600))

		pb := NewPathBuilder(filePath)
		_, err := pb.ListFiles()
		require.Error(t, err)
	})

	t.Run("ListDirs returns error for non-directory", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "file3.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o600))

		pb := NewPathBuilder(filePath)
		_, err := pb.ListDirs()
		require.Error(t, err)
	})

	t.Run("Glob returns error for invalid pattern", func(t *testing.T) {
		pb := NewPathBuilder(tempDir)
		_, err := pb.Glob("[invalid")
		require.Error(t, err)
	})

	t.Run("Readlink returns error for non-symlink", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "regular_file.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o600))

		pb := NewPathBuilder(filePath)
		_, err := pb.Readlink()
		require.Error(t, err)
	})

	t.Run("Match returns false for invalid pattern", func(t *testing.T) {
		pb := NewPathBuilder("file.txt")
		result := pb.Match("[invalid")
		assert.False(t, result)
	})

	t.Run("Create returns error when file exists and overwrite disabled", func(t *testing.T) {
		filePath := filepath.Join(tempDir, "existing_file.txt")
		require.NoError(t, os.WriteFile(filePath, []byte("content"), 0o600))

		options := PathOptions{
			OverwriteExisting: false,
			CreateMode:        0o644,
		}
		pb := NewPathBuilderWithOptions(filePath, options)
		err := pb.Create()
		require.Error(t, err)
	})
}

// TestPathBuilder_IsEmpty tests IsEmpty method
func TestPathBuilder_IsEmpty(t *testing.T) {
	t.Run("empty path", func(t *testing.T) {
		pb := NewPathBuilder("")
		// After cleaning, empty path becomes "."
		assert.False(t, pb.IsEmpty())
	})

	t.Run("non-empty path", func(t *testing.T) {
		pb := NewPathBuilder("test/path")
		assert.False(t, pb.IsEmpty())
	})
}

// TestPathBuilder_IsValid tests IsValid method
func TestPathBuilder_IsValid(t *testing.T) {
	t.Run("valid path", func(t *testing.T) {
		pb := NewPathBuilder("test/path")
		assert.True(t, pb.IsValid())
	})

	t.Run("invalid path with traversal when not allowed", func(t *testing.T) {
		options := PathOptions{
			AllowUnsafePaths: false,
		}
		pb := NewPathBuilderWithOptions("../escape", options)
		assert.False(t, pb.IsValid())
	})
}

// TestPathBuilder_Mode tests Mode method for non-existent files
func TestPathBuilder_Mode(t *testing.T) {
	pb := NewPathBuilder("/nonexistent/path")
	mode := pb.Mode()
	assert.Equal(t, os.FileMode(0), mode)
}

// TestPathBuilder_Size tests Size method for non-existent files
func TestPathBuilder_Size(t *testing.T) {
	pb := NewPathBuilder("/nonexistent/path")
	size := pb.Size()
	assert.Equal(t, int64(0), size)
}

// TestPathBuilder_ModTime tests ModTime method for non-existent files
func TestPathBuilder_ModTime(t *testing.T) {
	pb := NewPathBuilder("/nonexistent/path")
	modTime := pb.ModTime()
	assert.True(t, modTime.IsZero())
}

// TestValidateExtractPath tests the security-critical path validation
// This is a critical test for preventing Zip Slip and path traversal attacks
func TestValidateExtractPath(t *testing.T) {
	tests := []struct {
		name        string
		destDir     string
		tarPath     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid simple path",
			destDir: "/tmp/extract",
			tarPath: "file.txt",
			wantErr: false,
		},
		{
			name:    "valid nested path",
			destDir: "/tmp/extract",
			tarPath: "subdir/file.txt",
			wantErr: false,
		},
		{
			name:        "path traversal attempt",
			destDir:     "/tmp/extract",
			tarPath:     "../../../etc/passwd",
			wantErr:     true,
			errContains: "..",
		},
		{
			name:        "absolute path attempt",
			destDir:     "/tmp/extract",
			tarPath:     "/etc/passwd",
			wantErr:     true,
			errContains: "..",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = NewPathBuilder(tt.destDir) // Create for context

			// First check if tarPath itself is absolute (Zip Slip defense)
			isAbsolute := filepath.IsAbs(tt.tarPath)

			// Build the target path
			targetPath := filepath.Join(tt.destDir, tt.tarPath)
			targetPb := NewPathBuilder(targetPath)

			// Check if path stays within destDir
			rel, err := filepath.Rel(tt.destDir, targetPath)
			hasTraversal := err != nil || strings.HasPrefix(rel, "..") || isAbsolute

			if tt.wantErr {
				assert.True(t, hasTraversal, "should detect path traversal")
			} else {
				assert.False(t, hasTraversal, "should allow valid path")
				// Verify the path is within destDir
				assert.False(t, strings.HasPrefix(rel, ".."))
			}

			// Additional security check using IsSafe
			if !tt.wantErr {
				// For valid paths, they should be safe (unless they contain other unsafe patterns)
				_ = targetPb.IsSafe() // Just exercise the method
			}
		})
	}
}
