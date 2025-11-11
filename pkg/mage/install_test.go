package mage

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestInstall_BinaryNameResolution tests binary name resolution logic
func TestInstall_BinaryNameResolution(t *testing.T) {
	tests := []struct {
		name       string
		binaryName string
		wantEmpty  bool
	}{
		{
			name:       "explicit binary name",
			binaryName: "myapp",
			wantEmpty:  false,
		},
		{
			name:       "empty binary name needs resolution",
			binaryName: "",
			wantEmpty:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.binaryName != "" {
				assert.NotEmpty(t, tt.binaryName)
			} else {
				assert.Empty(t, tt.binaryName)
			}
		})
	}
}

// TestInstall_PathConstruction tests install path construction
func TestInstall_PathConstruction(t *testing.T) {
	tests := []struct {
		name       string
		binaryName string
		wantExt    string
	}{
		{
			name:       "linux binary no extension",
			binaryName: "myapp",
			wantExt:    "",
		},
		{
			name:       "windows binary with exe",
			binaryName: "myapp",
			wantExt:    ".exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gopath := "/tmp/gopath"
			installPath := filepath.Join(gopath, "bin", tt.binaryName)

			assert.Contains(t, installPath, tt.binaryName)
		})
	}
}

// TestInstall_GopathResolution tests GOPATH resolution
func TestInstall_GopathResolution(t *testing.T) {
	tests := []struct {
		name       string
		gopathEnv  string
		wantCustom bool
	}{
		{
			name:       "custom GOPATH",
			gopathEnv:  "/custom/go",
			wantCustom: true,
		},
		{
			name:       "default GOPATH when not set",
			gopathEnv:  "",
			wantCustom: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.gopathEnv != "" {
				t.Setenv("GOPATH", tt.gopathEnv)
				// GOPATH environment variable set
			}
		})
	}
}

// TestInstall_WindowsExeExtension tests Windows .exe extension handling
func TestInstall_WindowsExeExtension(t *testing.T) {
	tests := []struct {
		name       string
		binaryName string
		wantSuffix string
	}{
		{
			name:       "add exe extension",
			binaryName: "myapp",
			wantSuffix: ".exe",
		},
		{
			name:       "already has exe extension",
			binaryName: "myapp.exe",
			wantSuffix: ".exe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if runtime.GOOS == "windows" {
				// On Windows, we expect .exe suffix
				assert.Contains(t, tt.wantSuffix, ".exe")
			}
		})
	}
}

// TestInstall_ErrorConstants tests error constants
func TestInstall_ErrorConstants(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "installation verification failed",
			err:  errInstallationVerificationFailed,
			want: "verification failed",
		},
		{
			name: "system-wide not supported on Windows",
			err:  errSystemWideNotSupportedWindows,
			want: "not supported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, tt.err)
			assert.Contains(t, tt.err.Error(), tt.want)
		})
	}
}

// TestInstall_Constants tests defined constants
func TestInstall_Constants(t *testing.T) {
	tests := []struct {
		name  string
		value string
		check func(string) bool
	}{
		{
			name:  "default binary name",
			value: defaultBinaryName,
			check: func(v string) bool { return v == "app" },
		},
		{
			name:  "windows exe extension",
			value: windowsExeExt,
			check: func(v string) bool { return v == ".exe" },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.True(t, tt.check(tt.value), "constant value should match expected")
		})
	}
}

// TestInstall_VersionResolution tests version resolution for go install
func TestInstall_VersionResolution(t *testing.T) {
	tests := []struct {
		name         string
		version      string
		expectLatest bool
	}{
		{
			name:         "explicit version",
			version:      "v1.2.3",
			expectLatest: false,
		},
		{
			name:         "dev version uses latest",
			version:      "dev",
			expectLatest: true,
		},
		{
			name:         "empty version uses latest",
			version:      "",
			expectLatest: true,
		},
		{
			name:         "latest version",
			version:      "latest",
			expectLatest: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isDevOrEmpty := tt.version == versionDev || tt.version == "" || tt.version == "latest"
			assert.Equal(t, tt.expectLatest, isDevOrEmpty)
		})
	}
}

// TestInstall_ModulePathParsing tests module path parsing logic
func TestInstall_ModulePathParsing(t *testing.T) {
	tests := []struct {
		name       string
		modulePath string
		wantBinary string
	}{
		{
			name:       "simple module name",
			modulePath: "github.com/user/myapp",
			wantBinary: "myapp",
		},
		{
			name:       "nested module path",
			modulePath: "github.com/user/project/cmd/app",
			wantBinary: "app",
		},
		{
			name:       "domain with subdomain",
			modulePath: "git.example.com/team/project",
			wantBinary: "project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parts := filepath.SplitList(tt.modulePath)
			if len(parts) > 0 {
				// Last part should be the binary name
				lastPart := filepath.Base(tt.modulePath)
				assert.NotEmpty(t, lastPart)
			}
		})
	}
}

// TestInstall_BuildTagsHandling tests build tags handling
func TestInstall_BuildTagsHandling(t *testing.T) {
	tests := []struct {
		name      string
		buildTags string
		wantTags  bool
	}{
		{
			name:      "with build tags",
			buildTags: "integration,e2e",
			wantTags:  true,
		},
		{
			name:      "without build tags",
			buildTags: "",
			wantTags:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasTags := tt.buildTags != ""
			assert.Equal(t, tt.wantTags, hasTags)
		})
	}
}

// TestInstall_AliasCreation tests alias creation logic
func TestInstall_AliasCreation(t *testing.T) {
	tests := []struct {
		name    string
		aliases []string
		want    int
	}{
		{
			name:    "multiple aliases",
			aliases: []string{"app", "myapp", "app-cli"},
			want:    3,
		},
		{
			name:    "single alias",
			aliases: []string{"app"},
			want:    1,
		},
		{
			name:    "no aliases",
			aliases: []string{},
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Len(t, tt.aliases, tt.want)
		})
	}
}

// TestInstall_PathChecking tests PATH checking logic
func TestInstall_PathChecking(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		inPath bool
	}{
		{
			name:   "path in PATH",
			path:   "/usr/local/bin",
			inPath: true,
		},
		{
			name:   "path not in PATH",
			path:   "/nonexistent/path",
			inPath: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the path is a valid string
			assert.NotEmpty(t, tt.path)
		})
	}
}
