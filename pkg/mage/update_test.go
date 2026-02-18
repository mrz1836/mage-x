package mage

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/common/fileops"
)

// TestGetUpdateChannel tests the getUpdateChannel function
func TestGetUpdateChannel(t *testing.T) {
	tests := []struct {
		name   string
		envVal string
		want   UpdateChannel
	}{
		{
			name:   "default to stable when no env set",
			envVal: "",
			want:   StableChannel,
		},
		{
			name:   "beta channel",
			envVal: "beta",
			want:   BetaChannel,
		},
		{
			name:   "Beta channel with capital B",
			envVal: "Beta",
			want:   BetaChannel,
		},
		{
			name:   "edge channel",
			envVal: "edge",
			want:   EdgeChannel,
		},
		{
			name:   "EDGE channel uppercase",
			envVal: "EDGE",
			want:   EdgeChannel,
		},
		{
			name:   "stable channel explicit",
			envVal: "stable",
			want:   StableChannel,
		},
		{
			name:   "invalid channel defaults to stable",
			envVal: "invalid",
			want:   StableChannel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envVal != "" {
				t.Setenv("UPDATE_CHANNEL", tt.envVal)
			}

			got := getUpdateChannel()
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestValidateExtractPath tests the security-critical path validation function
func TestValidateExtractPath(t *testing.T) {
	tests := []struct {
		name        string
		destDir     string
		tarPath     string
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid simple file path",
			destDir: "/tmp/extract",
			tarPath: "magex",
			wantErr: false,
		},
		{
			name:    "valid nested file path",
			destDir: "/tmp/extract",
			tarPath: "bin/magex",
			wantErr: false,
		},
		{
			name:    "valid file with dots in name",
			destDir: "/tmp/extract",
			tarPath: "magex.exe",
			wantErr: false,
		},
		{
			name:        "path traversal with ..",
			destDir:     "/tmp/extract",
			tarPath:     "../../../etc/passwd",
			wantErr:     true,
			errContains: "path traversal",
		},
		{
			name:        "path traversal in subdirectory",
			destDir:     "/tmp/extract",
			tarPath:     "subdir/../../etc/passwd",
			wantErr:     true,
			errContains: "path traversal",
		},
		{
			name:        "absolute path attempt",
			destDir:     "/tmp/extract",
			tarPath:     "/etc/passwd",
			wantErr:     true,
			errContains: "path traversal",
		},
		{
			name:    "multiple slashes normalized",
			destDir: "/tmp/extract",
			tarPath: "bin//magex",
			wantErr: false,
		},
		{
			name:    "trailing slash in tarPath",
			destDir: "/tmp/extract",
			tarPath: "bin/",
			wantErr: false,
		},
		{
			name:        "hidden parent traversal",
			destDir:     "/tmp/extract",
			tarPath:     "bin/../../../etc/passwd",
			wantErr:     true,
			errContains: "path traversal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := validateExtractPath(tt.destDir, tt.tarPath)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
				assert.Empty(t, got)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, got)
				// Verify the path is within destDir
				relPath, err := filepath.Rel(tt.destDir, got)
				require.NoError(t, err)
				assert.False(t, strings.HasPrefix(relPath, ".."), "path should be within destDir")
			}
		})
	}
}

// TestExtractTarGz tests tar.gz extraction with security validation
func TestExtractTarGz(t *testing.T) {
	tests := []struct {
		name        string
		setupTar    func(t *testing.T, tarPath string)
		wantErr     bool
		errContains string
		checkFiles  []string // Files that should exist after extraction
	}{
		{
			name: "valid single file extraction",
			setupTar: func(t *testing.T, tarPath string) {
				createTestTarGz(t, tarPath, map[string]string{
					"test.txt": "hello world",
				})
			},
			wantErr:    false,
			checkFiles: []string{"test.txt"},
		},
		{
			name: "valid multiple files",
			setupTar: func(t *testing.T, tarPath string) {
				createTestTarGz(t, tarPath, map[string]string{
					"file1.txt":      "content1",
					"bin/magex":      "binary content",
					"docs/README.md": "readme",
				})
			},
			wantErr:    false,
			checkFiles: []string{"file1.txt", "bin/magex", "docs/README.md"},
		},
		{
			name: "malicious path traversal attempt skipped",
			setupTar: func(t *testing.T, tarPath string) {
				// Create a tar with path traversal attempt
				file, err := os.Create(tarPath) //nolint:gosec // test file
				require.NoError(t, err)
				defer file.Close() //nolint:errcheck // test file

				gzipWriter := gzip.NewWriter(file)
				defer gzipWriter.Close() //nolint:errcheck // test file

				tarWriter := tar.NewWriter(gzipWriter)
				defer tarWriter.Close() //nolint:errcheck // test file

				// Add a malicious entry
				content := []byte("malicious content")
				header := &tar.Header{
					Name:     "../../../etc/passwd",
					Mode:     0o644,
					Size:     int64(len(content)),
					Typeflag: tar.TypeReg,
				}
				require.NoError(t, tarWriter.WriteHeader(header))
				_, err = tarWriter.Write(content)
				require.NoError(t, err)
			},
			wantErr:    false,      // Should succeed but skip malicious files
			checkFiles: []string{}, // No files should be extracted
		},
		{
			name: "empty tar file",
			setupTar: func(t *testing.T, tarPath string) {
				createTestTarGz(t, tarPath, map[string]string{})
			},
			wantErr:    false,
			checkFiles: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directories
			tempDir := t.TempDir()
			tarPath := filepath.Join(tempDir, "test.tar.gz")
			extractDir := filepath.Join(tempDir, "extract")
			require.NoError(t, os.MkdirAll(extractDir, 0o750))

			// Setup test tar file
			tt.setupTar(t, tarPath)

			// Extract
			err := extractTarGz(tarPath, extractDir)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)

				// Check expected files exist
				for _, expectedFile := range tt.checkFiles {
					expectedPath := filepath.Join(extractDir, expectedFile)
					_, err := os.Stat(expectedPath)
					assert.NoError(t, err, "expected file %s should exist", expectedFile)
				}
			}
		})
	}
}

// TestIsNewer tests version comparison logic
func TestIsNewer(t *testing.T) {
	tests := []struct {
		name       string
		newVersion string
		oldVersion string
		want       bool
	}{
		{
			name:       "newer version",
			newVersion: "v1.2.0",
			oldVersion: "v1.1.0",
			want:       true,
		},
		{
			name:       "same version",
			newVersion: "v1.1.0",
			oldVersion: "v1.1.0",
			want:       false,
		},
		{
			name:       "older version",
			newVersion: "v1.0.0",
			oldVersion: "v1.1.0",
			want:       false,
		},
		{
			name:       "dev version always needs update",
			newVersion: "v1.0.0",
			oldVersion: "dev",
			want:       true,
		},
		{
			name:       "versions without v prefix",
			newVersion: "1.2.0",
			oldVersion: "1.1.0",
			want:       true,
		},
		{
			name:       "prerelease versions",
			newVersion: "v1.2.0-beta.1",
			oldVersion: "v1.1.0",
			want:       true,
		},
		{
			name:       "empty old version",
			newVersion: "v1.0.0",
			oldVersion: "",
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNewer(tt.newVersion, tt.oldVersion)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestGetReleaseForChannel tests channel-based release selection
func TestGetReleaseForChannel(t *testing.T) {
	tests := []struct {
		name    string
		channel UpdateChannel
		wantErr bool
	}{
		{
			name:    "stable channel",
			channel: StableChannel,
			wantErr: false, // Will fail in test env but testing code path
		},
		{
			name:    "beta channel",
			channel: BetaChannel,
			wantErr: false,
		},
		{
			name:    "edge channel",
			channel: EdgeChannel,
			wantErr: false,
		},
		{
			name:    "invalid channel defaults to stable",
			channel: UpdateChannel("invalid"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test will make actual network calls in integration testing
			// For unit tests, we just verify the function doesn't panic
			// and handles channels correctly
			_, err := getReleaseForChannel("mrz1836", "mage-x", tt.channel)
			// In test environment without network, we expect errors
			// but we're testing the channel logic paths exist
			if err != nil {
				// Verify it's a network/API error, not a panic or logic error
				assert.True(t,
					errors.Is(err, errGitHubAPIError) ||
						errors.Is(err, errNoReleasesFound) ||
						errors.Is(err, errNoBetaReleasesFound) ||
						strings.Contains(err.Error(), "failed") ||
						strings.Contains(err.Error(), "gh CLI") ||
						strings.Contains(err.Error(), "rate limit"),
					"unexpected error type: %v", err)
			}
		})
	}
}

// TestUpdateInfo tests the UpdateInfo structure
func TestUpdateInfo(t *testing.T) {
	tests := []struct {
		name string
		info UpdateInfo
	}{
		{
			name: "complete update info",
			info: UpdateInfo{
				Channel:         StableChannel,
				CurrentVersion:  "v1.0.0",
				LatestVersion:   "v1.1.0",
				UpdateAvailable: true,
				ReleaseNotes:    "New features",
				DownloadURL:     "https://github.com/example/releases/download/v1.1.0/binary.tar.gz",
			},
		},
		{
			name: "no update available",
			info: UpdateInfo{
				Channel:         BetaChannel,
				CurrentVersion:  "v1.1.0",
				LatestVersion:   "v1.1.0",
				UpdateAvailable: false,
			},
		},
		{
			name: "edge channel with prerelease",
			info: UpdateInfo{
				Channel:         EdgeChannel,
				CurrentVersion:  "v1.0.0",
				LatestVersion:   "v1.1.0-beta.1",
				UpdateAvailable: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify structure fields
			assert.NotEmpty(t, tt.info.Channel)
			assert.NotEmpty(t, tt.info.CurrentVersion)
			assert.NotEmpty(t, tt.info.LatestVersion)

			// Verify update logic
			if tt.info.UpdateAvailable {
				// When update is available, latest should differ from current
				// or current should be "dev"
				shouldUpdate := tt.info.CurrentVersion != tt.info.LatestVersion ||
					tt.info.CurrentVersion == "dev"
				assert.True(t, shouldUpdate, "UpdateAvailable should match version comparison")
			}
		})
	}
}

// TestCopyFile tests file copying functionality
func TestCopyFile(t *testing.T) {
	tests := []struct {
		name        string
		srcContent  string
		wantErr     bool
		errContains string
	}{
		{
			name:       "successful copy",
			srcContent: "test content",
			wantErr:    false,
		},
		{
			name:       "copy empty file",
			srcContent: "",
			wantErr:    false,
		},
		{
			name:       "copy large file",
			srcContent: strings.Repeat("abcd", 1000),
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			srcPath := filepath.Join(tempDir, "source.txt")
			dstPath := filepath.Join(tempDir, "dest.txt")

			// Create source file
			err := os.WriteFile(srcPath, []byte(tt.srcContent), 0o644) //nolint:gosec // test file
			require.NoError(t, err)

			// Copy file
			err = copyFile(srcPath, dstPath)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)

				// Verify destination file exists and has same content
				dstContent, err := os.ReadFile(dstPath) //nolint:gosec // test file
				require.NoError(t, err)
				assert.Equal(t, tt.srcContent, string(dstContent))
			}
		})
	}
}

// TestUpdateChannel validates UpdateChannel constants
func TestUpdateChannel(t *testing.T) {
	tests := []struct {
		name    string
		channel UpdateChannel
		want    string
	}{
		{
			name:    "stable channel constant",
			channel: StableChannel,
			want:    "stable",
		},
		{
			name:    "beta channel constant",
			channel: BetaChannel,
			want:    "beta",
		},
		{
			name:    "edge channel constant",
			channel: EdgeChannel,
			want:    "edge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, string(tt.channel))
		})
	}
}

// TestGetVersionInfoForUpdate tests version detection for updates
func TestGetVersionInfoForUpdate(t *testing.T) {
	t.Run("returns version string", func(t *testing.T) {
		version := getVersionInfoForUpdate()
		// Should return either "dev" or a version string
		assert.NotEmpty(t, version)
		// Version should be either "dev" or start with "v" or be a semantic version
		isValid := version == "dev" ||
			strings.HasPrefix(version, "v") ||
			strings.Contains(version, ".")
		assert.True(t, isValid, "version should be dev or valid semver: %s", version)
	})
}

// TestErrorConstants verifies error constants are defined
func TestErrorConstants(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "no releases found error",
			err:  errNoReleasesFound,
			want: "no releases found",
		},
		{
			name: "no beta releases found error",
			err:  errNoBetaReleasesFound,
			want: "no beta releases found",
		},
		{
			name: "no tar.gz found error",
			err:  errNoTarGzFound,
			want: "no tar.gz file found",
		},
		{
			name: "magex binary not found error",
			err:  errMagexBinaryNotFound,
			want: "magex binary not found",
		},
		{
			name: "path traversal error",
			err:  errPathTraversal,
			want: "path traversal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.err)
			assert.Contains(t, tt.err.Error(), tt.want)
		})
	}
}

// TestCreateMagexAliases tests that createMagexAliases creates aliases
// even when no .mage.yaml is present (the normal user scenario for update:install)
func TestCreateMagexAliases(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink tests not supported on Windows")
	}

	t.Run("creates mgx alias without config file", func(t *testing.T) {
		// Set up a temp directory to act as GOPATH with a fake magex binary
		tempDir := t.TempDir()
		binDir := filepath.Join(tempDir, "bin")
		require.NoError(t, os.MkdirAll(binDir, fileops.PermDirSensitive))

		magexPath := filepath.Join(binDir, "magex")
		require.NoError(t, os.WriteFile(magexPath, []byte("fake binary"), 0o755)) //nolint:gosec // test file

		// Change to a directory without .mage.yaml to simulate running from any dir
		origDir, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(t.TempDir()))
		t.Cleanup(func() {
			require.NoError(t, os.Chdir(origDir))
		})

		// Call the function under test
		createMagexAliases(tempDir, magexPath)

		// Verify the mgx symlink was created
		mgxPath := filepath.Join(binDir, "mgx")
		info, err := os.Lstat(mgxPath)
		require.NoError(t, err, "mgx symlink should exist")
		assert.NotEqual(t, 0, info.Mode()&os.ModeSymlink, "mgx should be a symlink")

		// Verify it points to the magex binary
		target, err := os.Readlink(mgxPath)
		require.NoError(t, err)
		assert.Equal(t, magexPath, target)
	})

	t.Run("does not fail if alias already exists", func(t *testing.T) {
		tempDir := t.TempDir()
		binDir := filepath.Join(tempDir, "bin")
		require.NoError(t, os.MkdirAll(binDir, fileops.PermDirSensitive))

		magexPath := filepath.Join(binDir, "magex")
		require.NoError(t, os.WriteFile(magexPath, []byte("fake binary"), 0o755)) //nolint:gosec // test file

		// Pre-create the mgx symlink
		mgxPath := filepath.Join(binDir, "mgx")
		require.NoError(t, os.Symlink(magexPath, mgxPath))

		// Change to dir without .mage.yaml
		origDir, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(t.TempDir()))
		t.Cleanup(func() {
			require.NoError(t, os.Chdir(origDir))
		})

		// Should not panic or error
		createMagexAliases(tempDir, magexPath)

		// Symlink should still exist and point to magex
		target, err := os.Readlink(mgxPath)
		require.NoError(t, err)
		assert.Equal(t, magexPath, target)
	})
}

// Helper function to create a test tar.gz file
func createTestTarGz(t *testing.T, tarPath string, files map[string]string) {
	t.Helper()

	file, err := os.Create(tarPath) //nolint:gosec // test file
	require.NoError(t, err)
	defer file.Close() //nolint:errcheck // test file

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close() //nolint:errcheck // test file

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close() //nolint:errcheck // test file

	for name, content := range files {
		// Create directory entries if needed
		dir := filepath.Dir(name)
		if dir != "." && dir != "" {
			header := &tar.Header{
				Name:     dir + "/",
				Mode:     0o755,
				Typeflag: tar.TypeDir,
			}
			require.NoError(t, tarWriter.WriteHeader(header))
		}

		// Create file entry
		contentBytes := []byte(content)
		header := &tar.Header{
			Name:     name,
			Mode:     0o644,
			Size:     int64(len(contentBytes)),
			Typeflag: tar.TypeReg,
		}
		require.NoError(t, tarWriter.WriteHeader(header))
		_, err := tarWriter.Write(contentBytes)
		require.NoError(t, err)
	}
}

// TestDownloadUpdateAssetPattern tests the asset pattern matching for different platforms
func TestDownloadUpdateAssetPattern(t *testing.T) {
	tests := []struct {
		name        string
		assetName   string
		shouldMatch bool
	}{
		{
			name:        "matches current platform",
			assetName:   "mage-x_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz",
			shouldMatch: true,
		},
		{
			name:        "linux amd64",
			assetName:   "mage-x_linux_amd64.tar.gz",
			shouldMatch: runtime.GOOS == "linux" && runtime.GOARCH == "amd64",
		},
		{
			name:        "darwin arm64",
			assetName:   "mage-x_darwin_arm64.tar.gz",
			shouldMatch: runtime.GOOS == "darwin" && runtime.GOARCH == "arm64",
		},
		{
			name:        "windows amd64",
			assetName:   "mage-x_windows_amd64.tar.gz",
			shouldMatch: runtime.GOOS == "windows" && runtime.GOARCH == "amd64",
		},
		{
			name:        "wrong platform",
			assetName:   "mage-x_wrongos_wrongarch.tar.gz",
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
			matches := strings.Contains(tt.assetName, pattern)
			assert.Equal(t, tt.shouldMatch, matches)
		})
	}
}
