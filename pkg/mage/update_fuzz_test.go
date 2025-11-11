package mage

import (
	"path/filepath"
	"strings"
	"testing"
)

// FuzzValidateExtractPath tests the validateExtractPath function for security vulnerabilities
// and edge cases that could lead to directory traversal attacks (Zip Slip).
//
// This fuzzer focuses on:
// - Directory traversal attempts (../, ../../, etc.)
// - Absolute paths (/etc/passwd, C:\Windows\, etc.)
// - Mixed path separators and tricks
// - Edge cases with empty strings, special characters
// - Unicode and encoded characters
// - Symlink-like patterns
func FuzzValidateExtractPath(f *testing.F) {
	// Seed corpus with known attack patterns and edge cases
	testCases := []struct {
		destDir string
		tarPath string
	}{
		// Normal valid cases
		{"/tmp/safe", "file.txt"},
		{"/tmp/safe", "subdir/file.txt"},
		{"/tmp/safe", "deep/nested/dir/file.txt"},

		// Classic directory traversal attacks
		{"/tmp/safe", "../etc/passwd"},
		{"/tmp/safe", "../../etc/passwd"},
		{"/tmp/safe", "../../../etc/passwd"},
		{"/tmp/safe", "../../../../etc/passwd"},

		// Traversal hidden in valid paths
		{"/tmp/safe", "valid/../../../etc/passwd"},
		{"/tmp/safe", "./valid/../../etc/passwd"},
		{"/tmp/safe", "a/b/c/../../../etc/passwd"},

		// Absolute paths (should be rejected)
		{"/tmp/safe", "/etc/passwd"},
		{"/tmp/safe", "/etc/shadow"},
		{"/tmp/safe", "/root/.ssh/id_rsa"},

		// Windows-specific attacks
		{"/tmp/safe", "C:\\Windows\\System32\\config\\sam"},
		{"/tmp/safe", "\\\\server\\share\\file"},
		{"/tmp/safe", "..\\..\\..\\Windows\\System32"},

		// Mixed separators
		{"/tmp/safe", "../\\../etc/passwd"},
		{"/tmp/safe", "..\\/../etc/passwd"},
		{"/tmp/safe", "valid/..\\/../../etc/passwd"},

		// Edge cases with dots
		{"/tmp/safe", "..."},
		{"/tmp/safe", "...."},
		{"/tmp/safe", ".../etc/passwd"},
		{"/tmp/safe", "..../etc/passwd"},

		// Empty and whitespace
		{"/tmp/safe", ""},
		{"/tmp/safe", " "},
		{"/tmp/safe", "\t"},
		{"/tmp/safe", "\n"},

		// Leading/trailing slashes
		{"/tmp/safe", "/file.txt"},
		{"/tmp/safe", "file.txt/"},
		{"/tmp/safe", "/subdir/file.txt"},

		// Multiple consecutive slashes
		{"/tmp/safe", "subdir//file.txt"},
		{"/tmp/safe", "subdir///file.txt"},
		{"/tmp/safe", "//file.txt"},

		// Null bytes (if not handled by filepath)
		{"/tmp/safe", "file\x00.txt"},
		{"/tmp/safe", "../\x00etc/passwd"},

		// Unicode tricks
		{"/tmp/safe", "\u002e\u002e\u002f\u002e\u002e\u002fetc/passwd"}, // Unicode dots and slash
		{"/tmp/safe", "valid\u002f\u002e\u002e\u002f\u002e\u002e"},

		// URL encoding patterns (should not be decoded by function)
		{"/tmp/safe", "%2e%2e%2f%2e%2e%2fetc/passwd"},
		{"/tmp/safe", "..%2f..%2fetc%2fpasswd"},

		// Very long paths
		{"/tmp/safe", strings.Repeat("a/", 100) + "file.txt"},
		{"/tmp/safe", strings.Repeat("../", 100)},

		// Paths with various destination directories
		{"/", "file.txt"},
		{"/tmp", "file.txt"},
		{".", "file.txt"},
		{"..", "file.txt"},
		{"/tmp/safe", "."},
		{"/tmp/safe", ".."},

		// Symlink-like patterns (actual symlinks would be different, but test strings)
		{"/tmp/safe", "link->../../../etc/passwd"},
		{"/tmp/safe", "link/../../../etc/passwd"},

		// Special filenames
		{"/tmp/safe", "..file.txt"},  // File starting with ..
		{"/tmp/safe", "...file.txt"}, // File starting with ...
		{"/tmp/safe", "file..txt"},   // File with .. in middle

		// Case variations
		{"/tmp/safe", "../ETC/PASSWD"},
		{"/tmp/safe", "..\\..\\ETC\\passwd"},
	}

	// Add all test cases to fuzzer corpus
	for _, tc := range testCases {
		f.Add(tc.destDir, tc.tarPath)
	}

	// Fuzzing function
	f.Fuzz(func(t *testing.T, destDir, tarPath string) {
		// Call the function under test
		result, err := validateExtractPath(destDir, tarPath)

		// Security-critical invariants that must always hold:

		// 1. If tarPath is absolute, it MUST be rejected
		if filepath.IsAbs(tarPath) && err == nil {
			t.Errorf("SECURITY: absolute path not rejected: tarPath=%q, result=%q", tarPath, result)
		}

		// 2. If no error, result MUST be within destDir
		if err == nil {
			cleanDest := filepath.Clean(destDir)
			cleanResult := filepath.Clean(result)

			// Use filepath.Rel to verify the result is within destDir
			relPath, relErr := filepath.Rel(cleanDest, cleanResult)
			if relErr != nil {
				t.Errorf("SECURITY: cannot determine relative path: destDir=%q, tarPath=%q, result=%q, err=%v",
					destDir, tarPath, result, relErr)
			} else {
				// Relative path must not start with .. or contain /..
				if strings.HasPrefix(relPath, "..") || strings.Contains(relPath, string(filepath.Separator)+"..") {
					t.Errorf("SECURITY: relative path contains traversal: destDir=%q, tarPath=%q, result=%q, relPath=%q",
						destDir, tarPath, result, relPath)
				}
			}
		}

		// 3. Common traversal patterns should always be rejected
		if err == nil {
			// Check for obvious traversal patterns in tarPath
			hasTraversal := strings.Contains(tarPath, "../") ||
				strings.Contains(tarPath, "..\\") ||
				strings.HasPrefix(tarPath, "..") ||
				(strings.Contains(tarPath, "/..") && !strings.Contains(tarPath, "/...")) // /.. but not /...

			if hasTraversal {
				// Verify the result is still safe
				cleanDest := filepath.Clean(destDir)
				cleanResult := filepath.Clean(result)
				relPath, relErr := filepath.Rel(cleanDest, cleanResult)

				if relErr == nil && (strings.HasPrefix(relPath, "..") ||
					strings.Contains(relPath, string(filepath.Separator)+"..")) {
					t.Errorf("SECURITY: traversal pattern not blocked: destDir=%q, tarPath=%q, result=%q",
						destDir, tarPath, result)
				}
			}
		}

		// 4. Error conditions should be consistent
		if err != nil {
			// Errors should be related to path traversal or path validation
			// No panics or unexpected error types
			_ = err.Error() // Should not panic
		}

		// 5. Function should never panic (this is implicitly tested by fuzzer)
		// If function panics, fuzzer will catch it
	})
}
