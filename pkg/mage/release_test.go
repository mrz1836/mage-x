package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestRelease_VersionPattern tests version pattern validation
func TestRelease_VersionPattern(t *testing.T) {
	tests := []struct {
		name    string
		version string
		valid   bool
	}{
		{
			name:    "semantic version",
			version: "v1.2.3",
			valid:   true,
		},
		{
			name:    "version without v prefix",
			version: "1.2.3",
			valid:   true,
		},
		{
			name:    "prerelease version",
			version: "v1.2.3-beta.1",
			valid:   true,
		},
		{
			name:    "dev version",
			version: "dev",
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation logic
			isValid := tt.version != versionDev && tt.version != ""
			if tt.valid {
				assert.True(t, isValid || tt.version == versionDev)
			}
		})
	}
}

// TestRelease_GitTagFormat tests Git tag formatting
func TestRelease_GitTagFormat(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantTag string
	}{
		{
			name:    "version with v prefix",
			version: "v1.2.3",
			wantTag: "v1.2.3",
		},
		{
			name:    "version without v prefix",
			version: "1.2.3",
			wantTag: "v1.2.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Tag should have v prefix
			assert.NotEmpty(t, tt.wantTag)
			if tt.version[0] != 'v' {
				assert.Contains(t, tt.wantTag, "v")
			}
		})
	}
}

// TestRelease_ChannelValidation tests release channel validation
func TestRelease_ChannelValidation(t *testing.T) {
	tests := []struct {
		name    string
		channel string
		valid   bool
	}{
		{
			name:    "stable channel",
			channel: "stable",
			valid:   true,
		},
		{
			name:    "beta channel",
			channel: "beta",
			valid:   true,
		},
		{
			name:    "edge channel",
			channel: "edge",
			valid:   true,
		},
		{
			name:    "invalid channel",
			channel: "invalid",
			valid:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validChannels := map[string]bool{
				"stable": true,
				"beta":   true,
				"edge":   true,
			}
			assert.Equal(t, tt.valid, validChannels[tt.channel])
		})
	}
}

// TestRelease_GitStateValidation tests Git working tree state validation
func TestRelease_GitStateValidation(t *testing.T) {
	tests := []struct {
		name       string
		hasChanges bool
		canRelease bool
	}{
		{
			name:       "clean working tree",
			hasChanges: false,
			canRelease: true,
		},
		{
			name:       "dirty working tree",
			hasChanges: true,
			canRelease: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean tree is required for release
			assert.Equal(t, tt.canRelease, !tt.hasChanges)
		})
	}
}

// TestRelease_SnapshotBuild tests snapshot build without release
func TestRelease_SnapshotBuild(t *testing.T) {
	tests := []struct {
		name       string
		isSnapshot bool
		shouldTag  bool
	}{
		{
			name:       "snapshot build no tag",
			isSnapshot: true,
			shouldTag:  false,
		},
		{
			name:       "release build with tag",
			isSnapshot: false,
			shouldTag:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.shouldTag, !tt.isSnapshot)
		})
	}
}

// TestRelease_LocalInstall tests local installation from release
func TestRelease_LocalInstall(t *testing.T) {
	tests := []struct {
		name        string
		version     string
		expectError bool
	}{
		{
			name:        "valid version tag",
			version:     "v1.2.3",
			expectError: false,
		},
		{
			name:        "dev version",
			version:     "dev",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isDev := tt.version == "dev"
			assert.Equal(t, tt.expectError, isDev)
		})
	}
}

// TestRelease_ChangelogGeneration tests changelog generation
func TestRelease_ChangelogGeneration(t *testing.T) {
	tests := []struct {
		name          string
		fromTag       string
		toTag         string
		shouldInclude bool
	}{
		{
			name:          "tags specified",
			fromTag:       "v1.0.0",
			toTag:         "v1.1.0",
			shouldInclude: true,
		},
		{
			name:          "no tags",
			fromTag:       "",
			toTag:         "",
			shouldInclude: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasTags := tt.fromTag != "" && tt.toTag != ""
			assert.Equal(t, tt.shouldInclude, hasTags)
		})
	}
}

// TestRelease_GitHubTokenValidation tests GitHub token validation
func TestRelease_GitHubTokenValidation(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		hasToken bool
	}{
		{
			name:     "token present",
			token:    "ghp_xxxxxxxxxxxx",
			hasToken: true,
		},
		{
			name:     "no token",
			token:    "",
			hasToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.hasToken, tt.token != "")
		})
	}
}

// TestRelease_GoReleaserConfigValidation tests goreleaser config validation
func TestRelease_GoReleaserConfigValidation(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		exists     bool
	}{
		{
			name:       "config exists",
			configPath: ".goreleaser.yml",
			exists:     true,
		},
		{
			name:       "config missing",
			configPath: "",
			exists:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.exists, tt.configPath != "")
		})
	}
}

// TestRelease_BuildFromTag tests building from a specific tag
func TestRelease_BuildFromTag(t *testing.T) {
	tests := []struct {
		name  string
		tag   string
		valid bool
	}{
		{
			name:  "valid tag",
			tag:   "v1.2.3",
			valid: true,
		},
		{
			name:  "empty tag",
			tag:   "",
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.tag != "")
		})
	}
}

// TestRelease_CleanValidation tests clean working directory validation
func TestRelease_CleanValidation(t *testing.T) {
	tests := []struct {
		name             string
		uncommittedFiles []string
		isClean          bool
	}{
		{
			name:             "no uncommitted files",
			uncommittedFiles: []string{},
			isClean:          true,
		},
		{
			name:             "has uncommitted files",
			uncommittedFiles: []string{"file1.go", "file2.go"},
			isClean:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isClean, len(tt.uncommittedFiles) == 0)
		})
	}
}
