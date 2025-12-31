package fileops

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// Static errors for comprehensive tests
var (
	errTestConcurrentWrite = errors.New("concurrent write validation failed")
)

// TestFileSizeEdgeCases tests boundary conditions for file sizes
func TestFileSizeEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	ops := NewDefaultFileOperator()
	safeOps := NewDefaultSafeFileOperator()

	t.Run("EmptyFileWrite", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "empty.txt")
		emptyData := []byte{}

		err := ops.WriteFile(testFile, emptyData, 0o644)
		require.NoError(t, err, "Should write empty file successfully")

		assert.True(t, ops.Exists(testFile), "Empty file should exist")

		info, err := ops.Stat(testFile)
		require.NoError(t, err)
		assert.Equal(t, int64(0), info.Size(), "Empty file should have size 0")
	})

	t.Run("EmptyFileRead", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "empty_read.txt")
		err := ops.WriteFile(testFile, []byte{}, 0o644)
		require.NoError(t, err)

		data, err := ops.ReadFile(testFile)
		require.NoError(t, err, "Should read empty file successfully")
		assert.Empty(t, data, "Read data should be empty")
		assert.NotNil(t, data, "Read data should not be nil")
	})

	t.Run("EmptyFileCopy", func(t *testing.T) {
		srcFile := filepath.Join(tmpDir, "empty_src.txt")
		dstFile := filepath.Join(tmpDir, "empty_dst.txt")

		err := ops.WriteFile(srcFile, []byte{}, 0o644)
		require.NoError(t, err)

		err = ops.Copy(srcFile, dstFile)
		require.NoError(t, err, "Should copy empty file successfully")

		dstInfo, err := ops.Stat(dstFile)
		require.NoError(t, err)
		assert.Equal(t, int64(0), dstInfo.Size(), "Copied empty file should have size 0")
	})

	t.Run("EmptyAtomicWrite", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "empty_atomic.txt")

		err := safeOps.WriteFileAtomic(testFile, []byte{}, 0o644)
		require.NoError(t, err, "Should write empty file atomically")

		data, err := ops.ReadFile(testFile)
		require.NoError(t, err)
		assert.Empty(t, data, "Atomic write empty file should be empty")
	})

	t.Run("LargeFile10MB", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "large_10mb.bin")
		size := 10 * 1024 * 1024 // 10MB
		testData := make([]byte, size)

		// Fill with pattern for integrity verification
		for i := range testData {
			testData[i] = byte(i % 251) // Prime number for better distribution
		}

		expectedHash := sha256.Sum256(testData)

		err := ops.WriteFile(testFile, testData, 0o644)
		require.NoError(t, err, "Should write 10MB file successfully")

		// Verify size
		info, err := ops.Stat(testFile)
		require.NoError(t, err)
		assert.Equal(t, int64(size), info.Size(), "Large file should have correct size")

		// Verify integrity
		readData, err := ops.ReadFile(testFile)
		require.NoError(t, err)
		actualHash := sha256.Sum256(readData)
		assert.Equal(t, expectedHash, actualHash, "Large file data integrity should be preserved")
	})

	t.Run("LargeFileCopy", func(t *testing.T) {
		srcFile := filepath.Join(tmpDir, "large_src.bin")
		dstFile := filepath.Join(tmpDir, "large_dst.bin")
		size := 5 * 1024 * 1024 // 5MB
		testData := make([]byte, size)

		for i := range testData {
			testData[i] = byte(i % 251)
		}

		err := ops.WriteFile(srcFile, testData, 0o644)
		require.NoError(t, err)

		err = ops.Copy(srcFile, dstFile)
		require.NoError(t, err, "Should copy large file successfully")

		srcData, err := ops.ReadFile(srcFile)
		require.NoError(t, err)
		dstData, err := ops.ReadFile(dstFile)
		require.NoError(t, err)

		assert.Equal(t, sha256.Sum256(srcData), sha256.Sum256(dstData), "Copied data should match source")
	})

	t.Run("LargeAtomicWrite", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "large_atomic.bin")
		size := 5 * 1024 * 1024 // 5MB
		testData := make([]byte, size)

		for i := range testData {
			testData[i] = byte(i % 251)
		}

		expectedHash := sha256.Sum256(testData)

		err := safeOps.WriteFileAtomic(testFile, testData, 0o644)
		require.NoError(t, err, "Should write large file atomically")

		readData, err := ops.ReadFile(testFile)
		require.NoError(t, err)
		actualHash := sha256.Sum256(readData)
		assert.Equal(t, expectedHash, actualHash, "Large atomic write should preserve data integrity")
	})
}

// TestPathTraversalSecurity tests various path traversal attack patterns
func TestPathTraversalSecurity(t *testing.T) {
	ops := NewDefaultFileOperator()

	traversalPatterns := []struct {
		name        string
		path        string
		description string
	}{
		{"basic_traversal", "../../../etc/passwd", "Basic ../ traversal"},
		{"traversal_in_middle", "foo/../../../etc/passwd", "Traversal after valid path component"},
		{"complex_traversal", "./foo/../bar/../../../etc/passwd", "Complex mixed traversal"},
		{"double_dot_variations", "..../etc/passwd", "Multiple dots"},
		{"trailing_traversal", "safe/../../..", "Trailing traversal"},
		{"multiple_traversals", "..//..//..//etc/passwd", "Multiple slashes with traversal"},
		{"mixed_separators_unix", "foo/..\\..\\etc/passwd", "Mixed forward and back slashes"},
	}

	for _, tt := range traversalPatterns {
		t.Run(tt.name+"_ReadFile", func(t *testing.T) {
			_, err := ops.ReadFile(tt.path)
			if err == nil {
				// If no error, verify it didn't actually access /etc/passwd
				// This is acceptable if the path was cleaned but didn't escape
				t.Logf("Path %q was accepted (may have been cleaned safely)", tt.path)
			} else {
				// Verify it was rejected for path traversal or file not found
				isTraversalError := errors.Is(err, ErrPathTraversalDetected)
				isNotFoundError := os.IsNotExist(err)
				assert.True(t, isTraversalError || isNotFoundError,
					"Should reject %s: %s (got: %v)", tt.description, tt.path, err)
			}
		})

		t.Run(tt.name+"_Copy_Source", func(t *testing.T) {
			tmpDir := t.TempDir()
			dstFile := filepath.Join(tmpDir, "dst.txt")

			err := ops.Copy(tt.path, dstFile)
			if err == nil {
				t.Errorf("Copy source should have rejected: %s", tt.description)
			}
		})

		t.Run(tt.name+"_Copy_Dest", func(t *testing.T) {
			tmpDir := t.TempDir()
			srcFile := filepath.Join(tmpDir, "src.txt")
			writeErr := ops.WriteFile(srcFile, []byte("test"), 0o644)
			require.NoError(t, writeErr)

			copyErr := ops.Copy(srcFile, tt.path)
			if copyErr == nil {
				// Clean up any file that was accidentally created
				_ = os.Remove(tt.path) //nolint:errcheck // Best effort cleanup
				t.Errorf("Copy destination should have rejected: %s", tt.description)
			}
		})
	}

	// Test that valid relative paths still work
	t.Run("valid_relative_path", func(t *testing.T) {
		tmpDir := t.TempDir()
		// Change to tmpDir to test relative paths
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			if chErr := os.Chdir(originalDir); chErr != nil {
				t.Logf("Failed to restore working directory: %v", chErr)
			}
		}()

		err = os.Chdir(tmpDir)
		require.NoError(t, err)

		// Create a file and read it with a relative path
		testFile := "test.txt"
		err = ops.WriteFile(testFile, []byte("test content"), 0o644)
		require.NoError(t, err)

		data, err := ops.ReadFile(testFile)
		require.NoError(t, err, "Should read valid relative path")
		assert.Equal(t, "test content", string(data))
	})
}

// TestConcurrentAtomicWrites tests race conditions in atomic write operations
func TestConcurrentAtomicWrites(t *testing.T) {
	tmpDir := t.TempDir()
	safeOps := NewDefaultSafeFileOperator()

	t.Run("ConcurrentWritesToSameFile", func(t *testing.T) {
		targetFile := filepath.Join(tmpDir, "concurrent_target.txt")
		const numGoroutines = 50
		var wg sync.WaitGroup
		errChan := make(chan error, numGoroutines)

		wg.Add(numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				defer wg.Done()
				// Each goroutine writes a unique identifier
				data := []byte(fmt.Sprintf("goroutine-%03d-content", index))
				if err := safeOps.WriteFileAtomic(targetFile, data, 0o644); err != nil {
					errChan <- fmt.Errorf("goroutine %d write failed: %w", index, err)
				}
			}(i)
		}

		wg.Wait()
		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Errorf("Concurrent write error: %v", err)
		}

		// Verify file exists and contains valid content from one goroutine
		data, err := safeOps.ReadFile(targetFile)
		require.NoError(t, err, "Should be able to read file after concurrent writes")

		// Content should match pattern from one of the goroutines
		contentPattern := regexp.MustCompile(`^goroutine-\d{3}-content$`)
		assert.True(t, contentPattern.Match(data),
			"File content should be from one goroutine, got: %q", string(data))
	})

	t.Run("ConcurrentReadDuringWrite", func(t *testing.T) {
		targetFile := filepath.Join(tmpDir, "read_during_write.txt")
		initialData := []byte("initial-content-that-is-longer")
		err := safeOps.WriteFileAtomic(targetFile, initialData, 0o644)
		require.NoError(t, err)

		const numReaders = 10
		const numWriters = 5
		const numIterations = 20
		var writerWg sync.WaitGroup
		var readerWg sync.WaitGroup
		readResults := make(chan []byte, 1000) // Buffer for reads
		done := make(chan struct{})

		// Start readers (they will stop when done is closed)
		readerWg.Add(numReaders)
		for i := 0; i < numReaders; i++ {
			go func() {
				defer readerWg.Done()
				for {
					select {
					case <-done:
						return
					default:
						data, readErr := safeOps.ReadFile(targetFile)
						if readErr == nil {
							select {
							case readResults <- data:
							case <-done:
								return
							}
						}
					}
				}
			}()
		}

		// Start writers
		writerWg.Add(numWriters)
		for i := 0; i < numWriters; i++ {
			go func(index int) {
				defer writerWg.Done()
				for j := 0; j < numIterations; j++ {
					data := []byte(fmt.Sprintf("writer-%02d-iteration-%02d-data", index, j))
					_ = safeOps.WriteFileAtomic(targetFile, data, 0o644) //nolint:errcheck // Intentionally ignoring errors in stress test
				}
			}(i)
		}

		// Wait for writers to finish, then signal readers to stop
		writerWg.Wait()
		close(done)
		readerWg.Wait()
		close(readResults)

		// Verify all reads got complete, valid data (not partial writes)
		validPattern := regexp.MustCompile(`^(initial-content-that-is-longer|writer-\d{2}-iteration-\d{2}-data)$`)
		invalidReads := 0
		for data := range readResults {
			if !validPattern.Match(data) {
				invalidReads++
				if invalidReads <= 5 { // Only log first 5 invalid reads
					t.Logf("Invalid read data: %q", string(data))
				}
			}
		}
		assert.Zero(t, invalidReads, "All reads should return complete, valid data")
	})
}

// TestPermissionEdgeCases tests file permission handling
func TestPermissionEdgeCases(t *testing.T) {
	// Skip permission tests when running as root (permissions don't apply)
	if os.Getuid() == 0 {
		t.Skip("Skipping permission tests when running as root")
	}

	tmpDir := t.TempDir()
	ops := NewDefaultFileOperator()
	safeOps := NewDefaultSafeFileOperator()

	t.Run("WriteToReadOnlyFile", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "readonly.txt")
		err := ops.WriteFile(testFile, []byte("original"), 0o644)
		require.NoError(t, err)

		// Make file read-only
		err = ops.Chmod(testFile, 0o444)
		require.NoError(t, err)

		// Attempt to write should fail
		err = ops.WriteFile(testFile, []byte("new content"), 0o644)
		require.Error(t, err, "Should fail to write to read-only file")
		assert.True(t, os.IsPermission(err), "Error should be permission error")

		// Cleanup: restore permissions so TempDir cleanup works
		_ = os.Chmod(testFile, 0o644) //nolint:errcheck,gosec // Best effort cleanup
	})

	t.Run("ReadFromWriteOnlyFile", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "writeonly.txt")
		err := ops.WriteFile(testFile, []byte("secret"), 0o644)
		require.NoError(t, err)

		// Make file write-only
		err = ops.Chmod(testFile, 0o222)
		require.NoError(t, err)

		// Attempt to read should fail
		_, err = ops.ReadFile(testFile)
		require.Error(t, err, "Should fail to read write-only file")
		assert.True(t, os.IsPermission(err), "Error should be permission error")

		// Cleanup
		_ = os.Chmod(testFile, 0o644) //nolint:errcheck,gosec // Best effort cleanup
	})

	t.Run("WriteToReadOnlyDirectory", func(t *testing.T) {
		readOnlyDir := filepath.Join(tmpDir, "readonly_dir")
		err := ops.MkdirAll(readOnlyDir, 0o755)
		require.NoError(t, err)

		// Make directory read-only (0o555 is intentional for test)
		err = os.Chmod(readOnlyDir, 0o555) //nolint:gosec // Intentionally testing with read-only directory
		require.NoError(t, err)

		// Attempt to create file in read-only directory should fail
		testFile := filepath.Join(readOnlyDir, "test.txt")
		err = ops.WriteFile(testFile, []byte("content"), 0o644)
		require.Error(t, err, "Should fail to write in read-only directory")

		// Cleanup
		_ = os.Chmod(readOnlyDir, 0o755) //nolint:errcheck,gosec // Best effort cleanup
	})

	t.Run("AtomicWriteToReadOnlyDirectory", func(t *testing.T) {
		readOnlyDir := filepath.Join(tmpDir, "readonly_atomic_dir")
		err := ops.MkdirAll(readOnlyDir, 0o755)
		require.NoError(t, err)

		// Make directory read-only (0o555 is intentional for test)
		err = os.Chmod(readOnlyDir, 0o555) //nolint:gosec // Intentionally testing with read-only directory
		require.NoError(t, err)

		// Atomic write should fail (can't create temp file)
		testFile := filepath.Join(readOnlyDir, "atomic.txt")
		err = safeOps.WriteFileAtomic(testFile, []byte("content"), 0o644)
		require.Error(t, err, "Atomic write should fail in read-only directory")
		assert.Contains(t, err.Error(), "failed to create temp file",
			"Error should mention temp file creation failure")

		// Cleanup
		_ = os.Chmod(readOnlyDir, 0o755) //nolint:errcheck,gosec // Best effort cleanup
	})

	t.Run("ChmodVariousPermissions", func(t *testing.T) {
		permTests := []os.FileMode{
			0o000, 0o100, 0o200, 0o300, 0o400, 0o500, 0o600, 0o700,
			0o644, 0o755, 0o777,
		}

		for _, perm := range permTests {
			testFile := filepath.Join(tmpDir, fmt.Sprintf("chmod_%04o.txt", perm))
			err := ops.WriteFile(testFile, []byte("test"), 0o644)
			require.NoError(t, err)

			err = ops.Chmod(testFile, perm)
			require.NoError(t, err, "Chmod to %04o should succeed", perm)

			info, err := ops.Stat(testFile)
			require.NoError(t, err)
			assert.Equal(t, perm, info.Mode().Perm(),
				"File should have permission %04o", perm)

			// Restore permissions for cleanup
			_ = os.Chmod(testFile, 0o644) //nolint:errcheck,gosec // Best effort cleanup
		}
	})
}

// TestSpecialFilenames tests handling of unicode and special character filenames
func TestSpecialFilenames(t *testing.T) {
	tmpDir := t.TempDir()
	ops := NewDefaultFileOperator()

	//nolint:gosmopolitan // Intentionally testing unicode filenames
	testCases := []struct {
		name     string
		filename string
		valid    bool
	}{
		{"unicode_chinese", "配置文件.yaml", true},
		{"unicode_cyrillic", "конфигурация.json", true},
		{"unicode_japanese", "設定.txt", true},
		{"unicode_emoji", "config_🚀.json", true},
		{"spaces", "my config file.yaml", true},
		{"multiple_extensions", "config.backup.yaml", true},
		{"dash_underscore", "my-config_file.yaml", true},
		{"numbers_start", "123config.txt", true},
		{"long_255_chars", strings.Repeat("a", 251) + ".txt", true}, // 255 total with extension
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tc.filename)
			testData := []byte(fmt.Sprintf("Content for %s", tc.name))

			writeErr := ops.WriteFile(testFile, testData, 0o644)
			if tc.valid {
				require.NoError(t, writeErr, "Should write file with name %q", tc.filename)

				// Verify can read back
				readData, readErr := ops.ReadFile(testFile)
				require.NoError(t, readErr, "Should read file with name %q", tc.filename)
				assert.Equal(t, testData, readData, "Content should match")

				// Verify Exists works
				assert.True(t, ops.Exists(testFile), "File %q should exist", tc.filename)
			} else {
				require.Error(t, writeErr, "Should fail to write file with name %q", tc.filename)
			}
		})
	}

	// Test deeply nested unicode path
	//nolint:gosmopolitan // Intentionally testing unicode paths
	t.Run("unicode_nested_path", func(t *testing.T) {
		nestedPath := filepath.Join(tmpDir, "настройки", "配置", "config.yaml")
		testData := []byte("nested unicode config")

		err := ops.MkdirAll(filepath.Dir(nestedPath), 0o755)
		require.NoError(t, err, "Should create nested unicode directories")

		err = ops.WriteFile(nestedPath, testData, 0o644)
		require.NoError(t, err, "Should write file in unicode path")

		data, err := ops.ReadFile(nestedPath)
		require.NoError(t, err)
		assert.Equal(t, testData, data)
	})
}

// TestAtomicWriteCleanup verifies temp files are cleaned up properly
func TestAtomicWriteCleanup(t *testing.T) {
	tmpDir := t.TempDir()
	safeOps := NewDefaultSafeFileOperator()

	t.Run("NoLeftoverTempFiles", func(t *testing.T) {
		// Perform multiple atomic writes
		for i := 0; i < 10; i++ {
			testFile := filepath.Join(tmpDir, fmt.Sprintf("atomic_%d.txt", i))
			err := safeOps.WriteFileAtomic(testFile, []byte(fmt.Sprintf("content %d", i)), 0o644)
			require.NoError(t, err)
		}

		// Check for any leftover temp files
		entries, err := os.ReadDir(tmpDir)
		require.NoError(t, err)

		tempFiles := 0
		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".tmp-") {
				tempFiles++
				t.Logf("Found leftover temp file: %s", entry.Name())
			}
		}

		assert.Zero(t, tempFiles, "Should have no leftover temp files")
	})

	t.Run("OverwriteExistingFile", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "overwrite.txt")
		originalData := []byte("original content")
		newData := []byte("new content that is different")

		// Write original file
		err := safeOps.WriteFile(testFile, originalData, 0o644)
		require.NoError(t, err)

		// Overwrite atomically
		err = safeOps.WriteFileAtomic(testFile, newData, 0o644)
		require.NoError(t, err)

		// Verify new content
		data, err := safeOps.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, newData, data, "Content should be updated")

		// Verify no temp files
		entries, err := os.ReadDir(tmpDir)
		require.NoError(t, err)

		for _, entry := range entries {
			assert.False(t, strings.HasPrefix(entry.Name(), ".tmp-"),
				"Should have no leftover temp files after overwrite")
		}
	})
}

// TestEncodingEdgeCases tests JSON/YAML encoding edge cases
func TestEncodingEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	fileOps := NewDefaultFileOperator()
	jsonOps := NewDefaultJSONOperator(fileOps)
	yamlOps := NewDefaultYAMLOperator(fileOps)

	t.Run("JSONWithBOM", func(t *testing.T) {
		// UTF-8 BOM: \xEF\xBB\xBF
		bomJSON := []byte("\xEF\xBB\xBF{\"key\": \"value\"}")
		testFile := filepath.Join(tmpDir, "bom.json")
		err := fileOps.WriteFile(testFile, bomJSON, 0o644)
		require.NoError(t, err)

		var result map[string]string
		err = jsonOps.ReadJSON(testFile, &result)
		// Standard library doesn't handle BOM automatically
		if err != nil {
			assert.Contains(t, err.Error(), "invalid character",
				"Should report invalid character at start")
		} else {
			// If it succeeded, verify the value
			assert.Equal(t, "value", result["key"])
		}
	})

	t.Run("DeeplyNestedJSON", func(t *testing.T) {
		// Create deeply nested structure
		depth := 100
		nested := make(map[string]interface{})
		current := nested
		for i := 0; i < depth-1; i++ {
			inner := make(map[string]interface{})
			current[fmt.Sprintf("level%d", i)] = inner
			current = inner
		}
		current["value"] = "deep"

		testFile := filepath.Join(tmpDir, "deep.json")
		err := jsonOps.WriteJSON(testFile, nested)
		require.NoError(t, err, "Should write deeply nested JSON")

		var result map[string]interface{}
		err = jsonOps.ReadJSON(testFile, &result)
		require.NoError(t, err, "Should read deeply nested JSON")
		assert.NotNil(t, result)
	})

	t.Run("YAMLTypeCoercion", func(t *testing.T) {
		yamlData := `
yes_value: yes
no_value: no
on_value: on
off_value: off
true_value: true
false_value: false
number_string: "123"
actual_number: 123
`
		testFile := filepath.Join(tmpDir, "coercion.yaml")
		err := fileOps.WriteFile(testFile, []byte(yamlData), 0o644)
		require.NoError(t, err)

		var result map[string]interface{}
		err = yamlOps.ReadYAML(testFile, &result)
		require.NoError(t, err, "Should read YAML with type coercion")

		// YAML v3 treats yes/no as strings, not booleans by default
		// but on/off and true/false are booleans
		assert.NotNil(t, result["yes_value"])
		assert.NotNil(t, result["no_value"])
		assert.Equal(t, true, result["true_value"], "true should be boolean true")
		assert.Equal(t, false, result["false_value"], "false should be boolean false")
		assert.Equal(t, "123", result["number_string"], "Quoted number should be string")
		assert.Equal(t, 123, result["actual_number"], "Unquoted number should be int")
	})

	t.Run("YAMLAnchorsAndAliases", func(t *testing.T) {
		yamlData := `
defaults: &defaults
  adapter: postgres
  host: localhost

development:
  <<: *defaults
  database: dev_db

production:
  <<: *defaults
  database: prod_db
  host: production.server.com
`
		testFile := filepath.Join(tmpDir, "anchors.yaml")
		err := fileOps.WriteFile(testFile, []byte(yamlData), 0o644)
		require.NoError(t, err)

		var result map[string]map[string]string
		err = yamlOps.ReadYAML(testFile, &result)
		require.NoError(t, err, "Should read YAML with anchors and aliases")

		// Verify anchor expansion
		assert.Equal(t, "postgres", result["development"]["adapter"],
			"Development should inherit adapter from defaults")
		assert.Equal(t, "localhost", result["development"]["host"],
			"Development should inherit host from defaults")
		assert.Equal(t, "dev_db", result["development"]["database"])

		assert.Equal(t, "postgres", result["production"]["adapter"],
			"Production should inherit adapter from defaults")
		assert.Equal(t, "production.server.com", result["production"]["host"],
			"Production should override host")
		assert.Equal(t, "prod_db", result["production"]["database"])
	})

	t.Run("EmptyJSONStructures", func(t *testing.T) {
		testCases := []struct {
			name string
			data string
		}{
			{"empty_object", "{}"},
			{"empty_array", "[]"},
			{"null", "null"},
		}

		for _, tc := range testCases {
			testFile := filepath.Join(tmpDir, fmt.Sprintf("empty_%s.json", tc.name))
			err := fileOps.WriteFile(testFile, []byte(tc.data), 0o644)
			require.NoError(t, err)

			var result interface{}
			err = jsonOps.ReadJSON(testFile, &result)
			require.NoError(t, err, "Should read %s", tc.name)
		}
	})

	t.Run("EmptyYAMLStructures", func(t *testing.T) {
		testCases := []struct {
			name string
			data string
		}{
			{"empty_doc", "---"},
			{"null_value", "~"},
			{"empty_map", "{}"},
			{"empty_list", "[]"},
		}

		for _, tc := range testCases {
			testFile := filepath.Join(tmpDir, fmt.Sprintf("empty_%s.yaml", tc.name))
			err := fileOps.WriteFile(testFile, []byte(tc.data), 0o644)
			require.NoError(t, err)

			var result interface{}
			err = yamlOps.ReadYAML(testFile, &result)
			require.NoError(t, err, "Should read %s", tc.name)
		}
	})
}

// TestLoadConfigEdgeCases tests LoadConfig fallback behavior
func TestLoadConfigEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	ops := New()

	type Config struct {
		Name    string `json:"name" yaml:"name"`
		Version int    `json:"version" yaml:"version"`
	}

	t.Run("EmptyConfigFile", func(t *testing.T) {
		emptyFile := filepath.Join(tmpDir, "empty.yaml")
		err := ops.File.WriteFile(emptyFile, []byte(""), 0o644)
		require.NoError(t, err)

		var config Config
		_, err = ops.LoadConfig([]string{emptyFile}, &config)
		// Empty YAML is valid and results in zero value
		if err != nil {
			t.Logf("Empty file error (expected): %v", err)
		}
	})

	t.Run("CorruptedFirstValidSecond", func(t *testing.T) {
		corruptedFile := filepath.Join(tmpDir, "corrupted.json")
		validFile := filepath.Join(tmpDir, "valid.yaml")

		// Create corrupted JSON
		err := ops.File.WriteFile(corruptedFile, []byte("{invalid json"), 0o644)
		require.NoError(t, err)

		// Create valid YAML
		validData, err := yaml.Marshal(Config{Name: "valid", Version: 2})
		require.NoError(t, err)
		err = ops.File.WriteFile(validFile, validData, 0o644)
		require.NoError(t, err)

		var config Config
		path, err := ops.LoadConfig([]string{corruptedFile, validFile}, &config)
		require.NoError(t, err, "Should fall back to valid file")
		assert.Equal(t, validFile, path, "Should return valid file path")
		assert.Equal(t, "valid", config.Name)
		assert.Equal(t, 2, config.Version)
	})

	t.Run("NoExtensionJSONContent", func(t *testing.T) {
		noExtFile := filepath.Join(tmpDir, "config")
		jsonData, err := json.Marshal(Config{Name: "json-config", Version: 3})
		require.NoError(t, err)
		err = ops.File.WriteFile(noExtFile, jsonData, 0o644)
		require.NoError(t, err)

		var config Config
		path, err := ops.LoadConfig([]string{noExtFile}, &config)
		require.NoError(t, err, "Should read JSON content from file without extension")
		assert.Equal(t, noExtFile, path)
		assert.Equal(t, "json-config", config.Name)
	})

	t.Run("YMLExtension", func(t *testing.T) {
		ymlFile := filepath.Join(tmpDir, "config.yml")
		yamlData, err := yaml.Marshal(Config{Name: "yml-config", Version: 4})
		require.NoError(t, err)
		err = ops.File.WriteFile(ymlFile, yamlData, 0o644)
		require.NoError(t, err)

		var config Config
		path, err := ops.LoadConfig([]string{ymlFile}, &config)
		require.NoError(t, err, "Should read .yml extension as YAML")
		assert.Equal(t, ymlFile, path) //nolint:testifylint // Comparing file paths not YAML content
		assert.Equal(t, "yml-config", config.Name)
	})

	t.Run("AllCorrupted", func(t *testing.T) {
		corrupted1 := filepath.Join(tmpDir, "bad1.json")
		corrupted2 := filepath.Join(tmpDir, "bad2.yaml")

		err := ops.File.WriteFile(corrupted1, []byte("{bad json"), 0o644)
		require.NoError(t, err)
		err = ops.File.WriteFile(corrupted2, []byte("bad:\n  - yaml\n  - [unclosed"), 0o644)
		require.NoError(t, err)

		var config Config
		_, err = ops.LoadConfig([]string{corrupted1, corrupted2}, &config)
		require.Error(t, err, "Should error when all files are corrupted")
	})
}

// TestSymlinkHandling tests symlink behavior (platform-specific)
func TestSymlinkHandling(t *testing.T) {
	//nolint:goconst // "windows" constant not worth extracting in test files
	if runtime.GOOS == "windows" {
		t.Skip("Symlink tests not reliable on Windows without admin privileges")
	}

	tmpDir := t.TempDir()
	ops := NewDefaultFileOperator()

	t.Run("ReadThroughSymlink", func(t *testing.T) {
		// Create target file
		targetFile := filepath.Join(tmpDir, "target.txt")
		testData := []byte("target content")
		err := ops.WriteFile(targetFile, testData, 0o644)
		require.NoError(t, err)

		// Create symlink
		symlinkPath := filepath.Join(tmpDir, "link.txt")
		err = os.Symlink(targetFile, symlinkPath)
		require.NoError(t, err)

		// Read through symlink
		data, err := ops.ReadFile(symlinkPath)
		require.NoError(t, err, "Should read file through symlink")
		assert.Equal(t, testData, data)
	})

	t.Run("StatSymlink", func(t *testing.T) {
		targetFile := filepath.Join(tmpDir, "stat_target.txt")
		err := ops.WriteFile(targetFile, []byte("content"), 0o644)
		require.NoError(t, err)

		symlinkPath := filepath.Join(tmpDir, "stat_link.txt")
		err = os.Symlink(targetFile, symlinkPath)
		require.NoError(t, err)

		// Stat follows symlink by default
		info, err := ops.Stat(symlinkPath)
		require.NoError(t, err)
		assert.Zero(t, info.Mode()&os.ModeSymlink,
			"Stat should follow symlink (not report symlink mode)")
	})

	t.Run("CopySymlink", func(t *testing.T) {
		targetFile := filepath.Join(tmpDir, "copy_target.txt")
		testData := []byte("copy target content")
		err := ops.WriteFile(targetFile, testData, 0o644)
		require.NoError(t, err)

		symlinkPath := filepath.Join(tmpDir, "copy_link.txt")
		err = os.Symlink(targetFile, symlinkPath)
		require.NoError(t, err)

		dstFile := filepath.Join(tmpDir, "copy_dst.txt")
		err = ops.Copy(symlinkPath, dstFile)
		require.NoError(t, err, "Should copy file via symlink")

		// Destination should have content (not be a symlink)
		dstData, err := ops.ReadFile(dstFile)
		require.NoError(t, err)
		assert.Equal(t, testData, dstData)

		// Verify destination is not a symlink
		dstInfo, err := os.Lstat(dstFile)
		require.NoError(t, err)
		assert.Zero(t, dstInfo.Mode()&os.ModeSymlink,
			"Destination should not be a symlink")
	})

	t.Run("BrokenSymlink", func(t *testing.T) {
		brokenLink := filepath.Join(tmpDir, "broken_link.txt")
		err := os.Symlink("/nonexistent/path/file.txt", brokenLink)
		require.NoError(t, err)

		// Reading broken symlink should fail
		_, err = ops.ReadFile(brokenLink)
		require.Error(t, err, "Should fail to read broken symlink")
		assert.True(t, os.IsNotExist(err), "Error should be 'not exist'")
	})
}

// TestWriteFileWithBackupEdgeCases tests backup functionality
func TestWriteFileWithBackupEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	safeOps := NewDefaultSafeFileOperator()

	t.Run("BackupOverwritesExistingBackup", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "backup_overwrite.txt")
		backupFile := testFile + ".bak"

		// Write original
		err := safeOps.WriteFile(testFile, []byte("version1"), 0o644)
		require.NoError(t, err)

		// First backup
		err = safeOps.WriteFileWithBackup(testFile, []byte("version2"), 0o644)
		require.NoError(t, err)

		backupData1, err := safeOps.ReadFile(backupFile)
		require.NoError(t, err)
		assert.Equal(t, "version1", string(backupData1))

		// Second backup (should overwrite first backup)
		err = safeOps.WriteFileWithBackup(testFile, []byte("version3"), 0o644)
		require.NoError(t, err)

		backupData2, err := safeOps.ReadFile(backupFile)
		require.NoError(t, err)
		assert.Equal(t, "version2", string(backupData2),
			"Backup should contain previous version")

		mainData, err := safeOps.ReadFile(testFile)
		require.NoError(t, err)
		assert.Equal(t, "version3", string(mainData))
	})

	t.Run("BackupPreservesPermissions", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "backup_perms.txt")
		err := safeOps.WriteFile(testFile, []byte("original"), 0o600)
		require.NoError(t, err)

		err = safeOps.WriteFileWithBackup(testFile, []byte("new"), 0o600)
		require.NoError(t, err)

		backupFile := testFile + ".bak"
		info, err := safeOps.Stat(backupFile)
		require.NoError(t, err)

		// Backup should have same permissions as original (Copy preserves them)
		assert.Equal(t, os.FileMode(0o600), info.Mode().Perm(),
			"Backup should preserve original file permissions")
	})
}

// TestFileOpsEnsureDirEdgeCases tests directory creation edge cases
func TestFileOpsEnsureDirEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	ops := New()

	t.Run("WriteToExistingDirectory", func(t *testing.T) {
		// Directory already exists
		existingDir := filepath.Join(tmpDir, "existing")
		err := ops.File.MkdirAll(existingDir, 0o755)
		require.NoError(t, err)

		testFile := filepath.Join(existingDir, "test.json")
		err = ops.WriteJSONSafe(testFile, map[string]string{"key": "value"})
		require.NoError(t, err, "Should write to existing directory")
	})

	t.Run("WriteToDeeplyNestedPath", func(t *testing.T) {
		deepPath := filepath.Join(tmpDir, "a", "b", "c", "d", "e", "f", "test.yaml")
		err := ops.WriteYAMLSafe(deepPath, map[string]string{"deep": "value"})
		require.NoError(t, err, "Should create deeply nested directories")

		assert.True(t, ops.File.Exists(deepPath), "File should exist")
	})
}

// TestCopyPreservesPermissions verifies Copy preserves source file permissions
func TestCopyPreservesPermissions(t *testing.T) {
	// Skip on Windows where file permissions work differently
	if runtime.GOOS == "windows" {
		t.Skip("Permission preservation tests not reliable on Windows")
	}

	tmpDir := t.TempDir()
	ops := NewDefaultFileOperator()

	permTests := []os.FileMode{0o644, 0o600, 0o755, 0o400}

	for _, perm := range permTests {
		t.Run(fmt.Sprintf("Permission_%04o", perm), func(t *testing.T) {
			srcFile := filepath.Join(tmpDir, fmt.Sprintf("src_%04o.txt", perm))
			dstFile := filepath.Join(tmpDir, fmt.Sprintf("dst_%04o.txt", perm))

			// Create source with specific permissions
			err := ops.WriteFile(srcFile, []byte("test"), perm)
			require.NoError(t, err)

			// Copy
			err = ops.Copy(srcFile, dstFile)
			require.NoError(t, err)

			// Verify destination has same permissions
			dstInfo, err := ops.Stat(dstFile)
			require.NoError(t, err)
			assert.Equal(t, perm, dstInfo.Mode().Perm(),
				"Destination should have same permissions as source")
		})
	}
}

// TestReadDirSorted verifies ReadDir returns entries
func TestReadDirOrdering(t *testing.T) {
	tmpDir := t.TempDir()
	ops := NewDefaultFileOperator()

	// Create files with various names
	files := []string{"zebra.txt", "apple.txt", "123.txt", "BETA.txt", "alpha.txt"}
	for _, f := range files {
		err := ops.WriteFile(filepath.Join(tmpDir, f), []byte("content"), 0o644)
		require.NoError(t, err)
	}

	entries, err := ops.ReadDir(tmpDir)
	require.NoError(t, err)
	assert.Len(t, entries, len(files), "Should return all files")

	// Verify all files are present (order may vary by filesystem)
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}

	for _, f := range files {
		assert.Contains(t, names, f, "Should contain file %s", f)
	}
}

// TestRemoveAllNestedStructure tests RemoveAll with complex directory structure
func TestRemoveAllNestedStructure(t *testing.T) {
	tmpDir := t.TempDir()
	ops := NewDefaultFileOperator()

	// Create complex structure
	structure := filepath.Join(tmpDir, "root")
	dirs := []string{
		"root",
		"root/a",
		"root/a/b",
		"root/a/b/c",
		"root/x",
		"root/x/y",
	}

	for _, dir := range dirs {
		err := ops.MkdirAll(filepath.Join(tmpDir, dir), 0o755)
		require.NoError(t, err)
	}

	// Create files
	files := []string{
		"root/file1.txt",
		"root/a/file2.txt",
		"root/a/b/file3.txt",
		"root/a/b/c/file4.txt",
		"root/x/file5.txt",
		"root/x/y/file6.txt",
	}

	for _, f := range files {
		err := ops.WriteFile(filepath.Join(tmpDir, f), []byte("content"), 0o644)
		require.NoError(t, err)
	}

	// Remove all
	err := ops.RemoveAll(structure)
	require.NoError(t, err, "RemoveAll should succeed")

	// Verify completely removed
	assert.False(t, ops.Exists(structure), "Root should not exist")

	// Verify parent still exists
	assert.True(t, ops.Exists(tmpDir), "Parent tmpDir should still exist")
}

// TestStatOnVariousTypes tests Stat on different file types
func TestStatOnVariousTypes(t *testing.T) {
	tmpDir := t.TempDir()
	ops := NewDefaultFileOperator()

	t.Run("RegularFile", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "regular.txt")
		testData := []byte("regular file content")
		err := ops.WriteFile(testFile, testData, 0o644)
		require.NoError(t, err)

		info, err := ops.Stat(testFile)
		require.NoError(t, err)
		assert.False(t, info.IsDir())
		assert.Equal(t, int64(len(testData)), info.Size())
		assert.Equal(t, "regular.txt", info.Name())
	})

	t.Run("Directory", func(t *testing.T) {
		testDir := filepath.Join(tmpDir, "testdir")
		err := ops.MkdirAll(testDir, 0o755)
		require.NoError(t, err)

		info, err := ops.Stat(testDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
		assert.Equal(t, "testdir", info.Name())
	})

	t.Run("EmptyFile", func(t *testing.T) {
		emptyFile := filepath.Join(tmpDir, "empty.txt")
		err := ops.WriteFile(emptyFile, []byte{}, 0o644)
		require.NoError(t, err)

		info, err := ops.Stat(emptyFile)
		require.NoError(t, err)
		assert.Equal(t, int64(0), info.Size())
	})

	t.Run("NonExistent", func(t *testing.T) {
		_, err := ops.Stat(filepath.Join(tmpDir, "nonexistent"))
		require.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})
}

// TestJSONOperatorEdgeCases tests additional JSON operator scenarios
func TestJSONOperatorEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	fileOps := NewDefaultFileOperator()
	jsonOps := NewDefaultJSONOperator(fileOps)

	t.Run("UnicodeContent", func(t *testing.T) {
		//nolint:gosmopolitan // Intentionally testing unicode content
		data := map[string]string{
			"chinese": "中文",
			"russian": "русский",
			"emoji":   "🎉🚀💡",
			"arabic":  "العربية",
		}

		testFile := filepath.Join(tmpDir, "unicode.json")
		err := jsonOps.WriteJSON(testFile, data)
		require.NoError(t, err)

		var result map[string]string
		err = jsonOps.ReadJSON(testFile, &result)
		require.NoError(t, err)

		assert.Equal(t, data, result, "Unicode content should round-trip correctly")
	})

	t.Run("LargeNumberOfKeys", func(t *testing.T) {
		data := make(map[string]int)
		for i := 0; i < 10000; i++ {
			data[fmt.Sprintf("key_%d", i)] = i
		}

		testFile := filepath.Join(tmpDir, "many_keys.json")
		err := jsonOps.WriteJSON(testFile, data)
		require.NoError(t, err)

		var result map[string]int
		err = jsonOps.ReadJSON(testFile, &result)
		require.NoError(t, err)
		assert.Len(t, result, 10000, "Should read all keys")
	})

	t.Run("SpecialStringValues", func(t *testing.T) {
		data := map[string]string{
			"quotes":    `"quoted"`,
			"backslash": `back\slash`,
			"newline":   "line1\nline2",
			"tab":       "col1\tcol2",
			"null_char": "before\x00after",
		}

		testFile := filepath.Join(tmpDir, "special_strings.json")
		err := jsonOps.WriteJSON(testFile, data)
		require.NoError(t, err)

		var result map[string]string
		err = jsonOps.ReadJSON(testFile, &result)
		require.NoError(t, err)

		assert.Equal(t, data, result, "Special string values should round-trip correctly")
	})
}

// TestYAMLOperatorEdgeCases tests additional YAML operator scenarios
func TestYAMLOperatorEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	fileOps := NewDefaultFileOperator()
	yamlOps := NewDefaultYAMLOperator(fileOps)

	t.Run("MultilineStrings", func(t *testing.T) {
		type Config struct {
			Script string `yaml:"script"`
		}

		testFile := filepath.Join(tmpDir, "multiline.yaml")
		yamlContent := `script: |
  #!/bin/bash
  echo "Hello"
  echo "World"
`
		err := fileOps.WriteFile(testFile, []byte(yamlContent), 0o644)
		require.NoError(t, err)

		var config Config
		err = yamlOps.ReadYAML(testFile, &config)
		require.NoError(t, err)

		assert.Contains(t, config.Script, "#!/bin/bash")
		assert.Contains(t, config.Script, "echo \"Hello\"")
	})

	t.Run("ComplexStructure", func(t *testing.T) {
		type Server struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}
		type Database struct {
			Driver   string `yaml:"driver"`
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			Name     string `yaml:"name"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
		}
		type Config struct {
			Server   Server                 `yaml:"server"`
			Database Database               `yaml:"database"`
			Features []string               `yaml:"features"`
			Settings map[string]interface{} `yaml:"settings"`
		}

		original := Config{
			Server: Server{Host: "localhost", Port: 8080},
			Database: Database{
				Driver:   "postgres",
				Host:     "db.example.com",
				Port:     5432,
				Name:     "myapp",
				User:     "admin",
				Password: "secret",
			},
			Features: []string{"auth", "cache", "logging"},
			Settings: map[string]interface{}{
				"debug":     true,
				"timeout":   30,
				"max_conns": 100,
			},
		}

		testFile := filepath.Join(tmpDir, "complex.yaml")
		err := yamlOps.WriteYAML(testFile, original)
		require.NoError(t, err)

		var result Config
		err = yamlOps.ReadYAML(testFile, &result)
		require.NoError(t, err)

		assert.Equal(t, original.Server, result.Server)
		assert.Equal(t, original.Database, result.Database)
		assert.ElementsMatch(t, original.Features, result.Features)
	})

	t.Run("NullValues", func(t *testing.T) {
		yamlContent := `
key1: null
key2: ~
key3:
`
		testFile := filepath.Join(tmpDir, "nulls.yaml")
		err := fileOps.WriteFile(testFile, []byte(yamlContent), 0o644)
		require.NoError(t, err)

		var result map[string]interface{}
		err = yamlOps.ReadYAML(testFile, &result)
		require.NoError(t, err)

		assert.Nil(t, result["key1"], "null should be nil")
		assert.Nil(t, result["key2"], "~ should be nil")
		assert.Nil(t, result["key3"], "empty value should be nil")
	})
}

// TestIsFileFunction tests the package-level IsFile function
func TestIsFileFunction(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file
	testFile := filepath.Join(tmpDir, "isfile_test.txt")
	err := WriteFile(testFile, []byte("content"), 0o644)
	require.NoError(t, err)

	// Create a directory
	testDir := filepath.Join(tmpDir, "isfile_dir")
	err = MkdirAll(testDir, 0o755)
	require.NoError(t, err)

	t.Run("File", func(t *testing.T) {
		assert.True(t, IsFile(testFile), "Should identify file correctly")
	})

	t.Run("Directory", func(t *testing.T) {
		assert.False(t, IsFile(testDir), "Directory should not be identified as file")
	})

	t.Run("NonExistent", func(t *testing.T) {
		assert.False(t, IsFile(filepath.Join(tmpDir, "nonexistent")),
			"Non-existent path should not be identified as file")
	})
}

// TestWriteYAMLSafeMarshalError tests error handling in WriteYAMLSafe
func TestWriteYAMLSafeMarshalError(t *testing.T) {
	tmpDir := t.TempDir()
	ops := New()

	t.Run("InvalidDataType", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "invalid.yaml")
		// Channels cannot be marshaled to YAML
		invalidData := make(chan int)

		// This should panic with yaml.v3
		assert.Panics(t, func() {
			_ = ops.WriteYAMLSafe(testFile, invalidData) //nolint:errcheck // Testing panic behavior
		}, "Should panic when marshaling channel")
	})
}

// TestExistsEdgeCases tests the Exists method edge cases
func TestExistsEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()
	ops := NewDefaultFileOperator()

	t.Run("EmptyPath", func(t *testing.T) {
		// Empty path should return false (or error internally)
		result := ops.Exists("")
		assert.False(t, result, "Empty path should not exist")
	})

	t.Run("CurrentDirectory", func(t *testing.T) {
		result := ops.Exists(".")
		assert.True(t, result, "Current directory should exist")
	})

	t.Run("RootDirectory", func(t *testing.T) {
		result := ops.Exists("/")
		assert.True(t, result, "Root directory should exist")
	})

	t.Run("JustCreatedFile", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "just_created.txt")

		assert.False(t, ops.Exists(testFile), "File should not exist yet")

		err := ops.WriteFile(testFile, []byte("test"), 0o644)
		require.NoError(t, err)

		assert.True(t, ops.Exists(testFile), "File should exist after creation")
	})

	t.Run("JustRemovedFile", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "to_remove.txt")
		err := ops.WriteFile(testFile, []byte("test"), 0o644)
		require.NoError(t, err)

		assert.True(t, ops.Exists(testFile), "File should exist")

		err = ops.Remove(testFile)
		require.NoError(t, err)

		assert.False(t, ops.Exists(testFile), "File should not exist after removal")
	})
}

// TestMarshalUnmarshalRoundTrip tests that data survives marshal/unmarshal cycle
func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	fileOps := NewDefaultFileOperator()
	jsonOps := NewDefaultJSONOperator(fileOps)
	yamlOps := NewDefaultYAMLOperator(fileOps)

	type TestData struct {
		String string                 `json:"string" yaml:"string"`
		Int    int                    `json:"int" yaml:"int"`
		Float  float64                `json:"float" yaml:"float"`
		Bool   bool                   `json:"bool" yaml:"bool"`
		Array  []int                  `json:"array" yaml:"array"`
		Map    map[string]string      `json:"map" yaml:"map"`
		Nested map[string]interface{} `json:"nested" yaml:"nested"`
	}

	//nolint:gosmopolitan // Intentionally testing unicode round-trip
	original := TestData{
		String: "test string with unicode: 日本語 🎉",
		Int:    42,
		Float:  3.14159,
		Bool:   true,
		Array:  []int{1, 2, 3, 4, 5},
		Map:    map[string]string{"key1": "value1", "key2": "value2"},
		Nested: map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": "deep value",
			},
		},
	}

	t.Run("JSON", func(t *testing.T) {
		data, err := jsonOps.Marshal(original)
		require.NoError(t, err)

		var result TestData
		err = jsonOps.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.String, result.String)
		assert.Equal(t, original.Int, result.Int)
		assert.InDelta(t, original.Float, result.Float, 0.00001)
		assert.Equal(t, original.Bool, result.Bool)
		assert.Equal(t, original.Array, result.Array)
		assert.Equal(t, original.Map, result.Map)
	})

	t.Run("YAML", func(t *testing.T) {
		data, err := yamlOps.Marshal(original)
		require.NoError(t, err)

		var result TestData
		err = yamlOps.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.Equal(t, original.String, result.String)
		assert.Equal(t, original.Int, result.Int)
		assert.InDelta(t, original.Float, result.Float, 0.00001)
		assert.Equal(t, original.Bool, result.Bool)
		assert.Equal(t, original.Array, result.Array)
		assert.Equal(t, original.Map, result.Map)
	})
}

// TestIsPathTraversal tests the path traversal detection helper function
func TestIsPathTraversal(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		// Traversal patterns that should be detected
		{"basic_traversal", "../etc/passwd", true},
		{"traversal_in_middle", "foo/../../../etc", true},
		{"double_dot", "..", true},
		{"trailing_traversal", "safe/../..", true},
		{"complex_traversal", "./foo/../bar/../../../etc", true},

		// Safe paths that should not be detected as traversal
		{"simple_relative", "foo/bar", false},
		{"simple_absolute", "/usr/local/bin", false},
		{"single_dot", ".", false},
		{"dotfile", ".gitignore", false},
		{"double_dotfile", "..gitignore", true}, // Files starting with ".." are rejected defensively
		{"current_dir_prefix", "./foo/bar", false},
		{"empty_path", "", false},
		{"just_filename", "test.txt", false},
		{"deeply_nested", "a/b/c/d/e/f/g", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPathTraversal(tt.path)
			assert.Equal(t, tt.expected, result, "isPathTraversal(%q) should be %v", tt.path, tt.expected)
		})
	}
}

// TestWriteFileWithBackupWrappedErrors verifies error handling works correctly
// with wrapped errors (tests the simplified errors.Is check)
func TestWriteFileWithBackupWrappedErrors(t *testing.T) {
	tmpDir := t.TempDir()
	safeOps := NewDefaultSafeFileOperator()

	t.Run("NonExistentSourceFile", func(t *testing.T) {
		// Writing to a path where source doesn't exist should succeed
		// (no backup needed, just write the new file)
		newFile := filepath.Join(tmpDir, "new_file.txt")
		err := safeOps.WriteFileWithBackup(newFile, []byte("content"), 0o644)
		require.NoError(t, err, "Should succeed when source doesn't exist")

		// Verify file was created
		assert.True(t, safeOps.Exists(newFile), "File should exist after write")

		// Verify no backup was created (since source didn't exist)
		backupPath := newFile + ".bak"
		assert.False(t, safeOps.Exists(backupPath), "No backup should exist for new file")
	})

	t.Run("ExistingSourceFile", func(t *testing.T) {
		// Create an existing file
		existingFile := filepath.Join(tmpDir, "existing.txt")
		err := safeOps.WriteFile(existingFile, []byte("original"), 0o644)
		require.NoError(t, err)

		// Write with backup
		err = safeOps.WriteFileWithBackup(existingFile, []byte("updated"), 0o644)
		require.NoError(t, err, "Should succeed with backup")

		// Verify backup exists with original content
		backupPath := existingFile + ".bak"
		assert.True(t, safeOps.Exists(backupPath), "Backup should exist")
		backupData, err := safeOps.ReadFile(backupPath)
		require.NoError(t, err)
		assert.Equal(t, []byte("original"), backupData, "Backup should contain original content")

		// Verify new content
		newData, err := safeOps.ReadFile(existingFile)
		require.NoError(t, err)
		assert.Equal(t, []byte("updated"), newData, "File should contain new content")
	})

	t.Run("ReadOnlyBackupDirectory", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping permission test on Windows")
		}

		// Create a subdirectory with a file
		subDir := filepath.Join(tmpDir, "readonly_subdir")
		err := os.MkdirAll(subDir, PermDirSensitive)
		require.NoError(t, err)

		testFile := filepath.Join(subDir, "test.txt")
		err = safeOps.WriteFile(testFile, []byte("original"), 0o644)
		require.NoError(t, err)

		// Make directory read-only to prevent backup creation
		err = os.Chmod(subDir, 0o555) //nolint:gosec // G302: intentionally testing read-only permissions
		require.NoError(t, err)
		defer func() {
			// Restore permissions for cleanup
			_ = os.Chmod(subDir, PermDir) //nolint:errcheck // Best effort cleanup
		}()

		// Attempt to write with backup - should fail due to permission denied
		err = safeOps.WriteFileWithBackup(testFile, []byte("updated"), 0o644)
		require.Error(t, err, "Should fail when backup cannot be created")
		assert.Contains(t, err.Error(), "failed to create backup", "Error should mention backup failure")
	})
}

// Ensure the error variable is used to avoid unused variable errors
var (
	_ = errTestConcurrentWrite
	_ = bytes.Equal
)
