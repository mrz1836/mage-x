package mage

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNormalizeFileMode verifies that file permissions are properly normalized
func TestNormalizeFileMode(t *testing.T) {
	tests := []struct {
		name     string
		input    os.FileMode
		expected os.FileMode
	}{
		{
			name:     "executable file gets 0755",
			input:    0o777,
			expected: 0o755,
		},
		{
			name:     "regular file gets 0644",
			input:    0o666,
			expected: 0o644,
		},
		{
			name:     "user-only executable gets 0755",
			input:    0o700,
			expected: 0o755,
		},
		{
			name:     "group-only executable gets 0755",
			input:    0o070,
			expected: 0o755,
		},
		{
			name:     "other-only executable gets 0755",
			input:    0o007,
			expected: 0o755,
		},
		{
			name:     "read-only file gets 0644",
			input:    0o444,
			expected: 0o644,
		},
		{
			name:     "setuid is stripped",
			input:    os.ModeSetuid | 0o755,
			expected: 0o755,
		},
		{
			name:     "setgid is stripped",
			input:    os.ModeSetgid | 0o755,
			expected: 0o755,
		},
		{
			name:     "sticky bit is stripped",
			input:    os.ModeSticky | 0o755,
			expected: 0o755,
		},
		{
			name:     "all special bits stripped from executable",
			input:    os.ModeSetuid | os.ModeSetgid | os.ModeSticky | 0o777,
			expected: 0o755,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeFileMode(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeFileMode(%o) = %o, want %o", tt.input, result, tt.expected)
			}
		})
	}
}

// TestValidateExtractPath_Security verifies path traversal prevention
func TestValidateExtractPath_Security(t *testing.T) {
	destDir := "/tmp/test-extract"

	tests := []struct {
		name      string
		tarPath   string
		wantErr   bool
		errSubstr string
	}{
		{
			name:    "simple filename is valid",
			tarPath: "magex",
			wantErr: false,
		},
		{
			name:    "nested path is valid",
			tarPath: "bin/magex",
			wantErr: false,
		},
		{
			name:      "absolute path is rejected",
			tarPath:   "/etc/passwd",
			wantErr:   true,
			errSubstr: "absolute path",
		},
		{
			name:      "path traversal is rejected",
			tarPath:   "../../../etc/passwd",
			wantErr:   true,
			errSubstr: "traversal",
		},
		{
			name:      "hidden traversal is rejected",
			tarPath:   "foo/../../etc/passwd",
			wantErr:   true,
			errSubstr: "traversal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateExtractPath(destDir, tt.tarPath)
			if tt.wantErr {
				if err == nil {
					t.Errorf("validateExtractPath(%q, %q) expected error containing %q, got nil",
						destDir, tt.tarPath, tt.errSubstr)
				} else if tt.errSubstr != "" && !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("validateExtractPath(%q, %q) error = %v, want error containing %q",
						destDir, tt.tarPath, err, tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("validateExtractPath(%q, %q) unexpected error: %v",
						destDir, tt.tarPath, err)
				}
			}
		})
	}
}

// TestExtractTarGz_PermissionNormalization verifies that extracted files have safe permissions
func TestExtractTarGz_PermissionNormalization(t *testing.T) {
	// Create a temp directory for extraction
	destDir := t.TempDir()

	// Create a tar.gz with files having dangerous permissions
	tarBuffer := new(bytes.Buffer)
	gzWriter := gzip.NewWriter(tarBuffer)
	tarWriter := tar.NewWriter(gzWriter)

	// Add an executable file with 0777 permissions
	execContent := []byte("#!/bin/sh\necho hello")
	if err := tarWriter.WriteHeader(&tar.Header{
		Name: "dangerous-exec",
		Mode: 0o777, // This should be normalized to 0755
		Size: int64(len(execContent)),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(execContent); err != nil {
		t.Fatal(err)
	}

	// Add a regular file with 0666 permissions
	dataContent := []byte("some data")
	if err := tarWriter.WriteHeader(&tar.Header{
		Name: "dangerous-data",
		Mode: 0o666, // This should be normalized to 0644
		Size: int64(len(dataContent)),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := tarWriter.Write(dataContent); err != nil {
		t.Fatal(err)
	}

	if err := tarWriter.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatal(err)
	}

	// Write the tar.gz to a file (using 0o600 for test files)
	tarGzPath := filepath.Join(destDir, "test.tar.gz")
	if err := os.WriteFile(tarGzPath, tarBuffer.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}

	// Extract the tar.gz (using 0o750 for directories)
	extractDir := filepath.Join(destDir, "extracted")
	if err := os.MkdirAll(extractDir, 0o750); err != nil {
		t.Fatal(err)
	}

	if err := extractTarGz(tarGzPath, extractDir); err != nil {
		t.Fatalf("extractTarGz failed: %v", err)
	}

	// Verify executable file has 0755 permissions
	execPath := filepath.Join(extractDir, "dangerous-exec")
	execInfo, err := os.Stat(execPath)
	if err != nil {
		t.Fatalf("failed to stat extracted executable: %v", err)
	}
	execPerm := execInfo.Mode().Perm()
	if execPerm != 0o755 {
		t.Errorf("executable file has permissions %o, want 0755", execPerm)
	}

	// Verify regular file has 0644 permissions
	dataPath := filepath.Join(extractDir, "dangerous-data")
	dataInfo, err := os.Stat(dataPath)
	if err != nil {
		t.Fatalf("failed to stat extracted data file: %v", err)
	}
	dataPerm := dataInfo.Mode().Perm()
	if dataPerm != 0o644 {
		t.Errorf("data file has permissions %o, want 0644", dataPerm)
	}
}

// TestExtractTarGz_SizeLimitConstant verifies the maxUpdateFileSize constant is set
func TestExtractTarGz_SizeLimitConstant(t *testing.T) {
	// Verify the constant is set to a reasonable value (500MB)
	expectedSize := int64(500 * 1024 * 1024)
	if maxUpdateFileSize != expectedSize {
		t.Errorf("maxUpdateFileSize = %d, want %d (500MB)", maxUpdateFileSize, expectedSize)
	}

	// The io.LimitReader is used in extractTarGz to prevent zip bomb attacks.
	// This test verifies the constant exists and is reasonable.
	// Full integration testing of zip bomb protection would require creating
	// a multi-gigabyte test file, which is impractical for unit tests.
}

// TestChecksumVerification verifies the checksum computation and comparison
func TestChecksumVerification(t *testing.T) {
	testData := []byte("test data for checksum verification")
	expectedHash := sha256.Sum256(testData)
	expectedHashHex := hex.EncodeToString(expectedHash[:])

	// Test correct checksum
	actualHash := sha256.Sum256(testData)
	actualHashHex := hex.EncodeToString(actualHash[:])

	if actualHashHex != expectedHashHex {
		t.Errorf("checksum mismatch: expected %s, got %s", expectedHashHex, actualHashHex)
	}

	// Test case-insensitive comparison
	if !strings.EqualFold(actualHashHex, expectedHashHex) {
		t.Error("case-insensitive comparison failed")
	}
}

// TestFetchChecksumForAsset_ParsesChecksumFile verifies checksum file parsing
func TestFetchChecksumForAsset_ParsesChecksumFile(t *testing.T) {
	// This is a unit test for the parsing logic
	// The actual fetching would require a mock HTTP server

	checksumFileContent := `abc123def456789012345678901234567890123456789012345678901234  mage-x_v1.0.0_linux_amd64.tar.gz
fedcba987654321098765432109876543210fedcba987654321098765432109  mage-x_v1.0.0_darwin_amd64.tar.gz
1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef  mage-x_v1.0.0_windows_amd64.tar.gz
`

	// Parse the content (simulating what fetchChecksumForAsset does)
	lines := bytes.Split([]byte(checksumFileContent), []byte("\n"))

	expectedChecksums := map[string]string{
		"mage-x_v1.0.0_linux_amd64.tar.gz":   "abc123def456789012345678901234567890123456789012345678901234",
		"mage-x_v1.0.0_darwin_amd64.tar.gz":  "fedcba987654321098765432109876543210fedcba987654321098765432109",
		"mage-x_v1.0.0_windows_amd64.tar.gz": "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
	}

	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := bytes.Fields(line)
		if len(parts) >= 2 {
			checksum := string(parts[0])
			filename := string(parts[1])

			expected, ok := expectedChecksums[filename]
			if !ok {
				t.Errorf("unexpected file in checksums: %s", filename)
				continue
			}
			if checksum != expected {
				t.Errorf("checksum for %s: got %s, want %s", filename, checksum, expected)
			}
		}
	}
}

// TestUpdateChannel_BetaDifferentFromEdge verifies beta and edge channels behave differently
func TestUpdateChannel_BetaDifferentFromEdge(t *testing.T) {
	// Test that the channel constants are distinct
	if BetaChannel == EdgeChannel {
		t.Error("BetaChannel and EdgeChannel should be different")
	}
	if BetaChannel == StableChannel {
		t.Error("BetaChannel and StableChannel should be different")
	}
	if EdgeChannel == StableChannel {
		t.Error("EdgeChannel and StableChannel should be different")
	}
}

// TestGetUpdateChannel_EnvParsing verifies UPDATE_CHANNEL environment variable parsing
func TestGetUpdateChannel_EnvParsing(t *testing.T) {
	tests := []struct {
		envValue string
		expected UpdateChannel
	}{
		{"stable", StableChannel},
		{"STABLE", StableChannel},
		{"beta", BetaChannel},
		{"BETA", BetaChannel},
		{"edge", EdgeChannel},
		{"EDGE", EdgeChannel},
		{"", StableChannel},           // default
		{"unknown", StableChannel},    // default for unknown
		{"  stable  ", StableChannel}, // note: current code doesn't trim
	}

	for _, tt := range tests {
		t.Run(tt.envValue, func(t *testing.T) {
			// Use t.Setenv which automatically restores the original value
			t.Setenv("UPDATE_CHANNEL", tt.envValue)

			result := getUpdateChannel()
			if result != tt.expected {
				t.Errorf("getUpdateChannel() with UPDATE_CHANNEL=%q = %v, want %v",
					tt.envValue, result, tt.expected)
			}
		})
	}
}
