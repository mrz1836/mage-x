package mage

import (
	"sort"
	"testing"
)

func TestGetToolVersion_PrimaryEnv(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		primaryEnv string
		primaryVal string
		want       string
	}{
		{
			name:       "golangci-lint from primary env",
			toolName:   "golangci-lint",
			primaryEnv: "MAGE_X_GOLANGCI_LINT_VERSION",
			primaryVal: "v1.55.0",
			want:       "v1.55.0",
		},
		{
			name:       "gofumpt from primary env",
			toolName:   "gofumpt",
			primaryEnv: "MAGE_X_GOFUMPT_VERSION",
			primaryVal: "v0.5.0",
			want:       "v0.5.0",
		},
		{
			name:       "yamlfmt from primary env",
			toolName:   "yamlfmt",
			primaryEnv: "MAGE_X_YAMLFMT_VERSION",
			primaryVal: "v0.10.0",
			want:       "v0.10.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set primary env var
			t.Setenv(tt.primaryEnv, tt.primaryVal)

			got := GetToolVersion(tt.toolName)
			if got != tt.want {
				t.Errorf("GetToolVersion(%q) = %q, want %q", tt.toolName, got, tt.want)
			}
		})
	}
}

func TestGetToolVersion_LegacyFallback(t *testing.T) {
	tests := []struct {
		name      string
		toolName  string
		legacyEnv string
		legacyVal string
		want      string
	}{
		{
			name:      "golangci-lint from legacy env",
			toolName:  "golangci-lint",
			legacyEnv: "GOLANGCI_LINT_VERSION",
			legacyVal: "v1.54.0",
			want:      "v1.54.0",
		},
		{
			name:      "gofumpt from legacy env",
			toolName:  "gofumpt",
			legacyEnv: "GOFUMPT_VERSION",
			legacyVal: "v0.4.0",
			want:      "v0.4.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure primary is not set
			cfg, _ := GetToolConfig(tt.toolName)
			t.Setenv(cfg.PrimaryEnv, "")

			// Set legacy env var
			t.Setenv(tt.legacyEnv, tt.legacyVal)

			got := GetToolVersion(tt.toolName)
			if got != tt.want {
				t.Errorf("GetToolVersion(%q) = %q, want %q", tt.toolName, got, tt.want)
			}
		})
	}
}

func TestGetToolVersion_Default(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		want     string
	}{
		{
			name:     "golangci-lint defaults to latest",
			toolName: "golangci-lint",
			want:     VersionLatest,
		},
		{
			name:     "gofumpt defaults to latest",
			toolName: "gofumpt",
			want:     VersionLatest,
		},
		{
			name:     "go defaults to 1.24",
			toolName: "go",
			want:     "1.24",
		},
		{
			name:     "go-secondary defaults to 1.23",
			toolName: "go-secondary",
			want:     "1.23",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure no env vars are set
			cfg, exists := GetToolConfig(tt.toolName)
			if exists {
				t.Setenv(cfg.PrimaryEnv, "")
				if cfg.LegacyEnv != "" {
					t.Setenv(cfg.LegacyEnv, "")
				}
			}

			got := GetToolVersion(tt.toolName)
			if got != tt.want {
				t.Errorf("GetToolVersion(%q) = %q, want %q", tt.toolName, got, tt.want)
			}
		})
	}
}

func TestGetToolVersion_GoSuffixStrip(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		envVar   string
		envVal   string
		want     string
	}{
		{
			name:     "go version strips .x suffix",
			toolName: "go",
			envVar:   "MAGE_X_GO_VERSION",
			envVal:   "1.24.x",
			want:     "1.24",
		},
		{
			name:     "go-secondary version strips .x suffix",
			toolName: "go-secondary",
			envVar:   "MAGE_X_GO_SECONDARY_VERSION",
			envVal:   "1.23.x",
			want:     "1.23",
		},
		{
			name:     "go version without suffix unchanged",
			toolName: "go",
			envVar:   "MAGE_X_GO_VERSION",
			envVal:   "1.24",
			want:     "1.24",
		},
		{
			name:     "go version with patch number unchanged",
			toolName: "go",
			envVar:   "MAGE_X_GO_VERSION",
			envVal:   "1.24.1",
			want:     "1.24.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envVar, tt.envVal)

			got := GetToolVersion(tt.toolName)
			if got != tt.want {
				t.Errorf("GetToolVersion(%q) = %q, want %q", tt.toolName, got, tt.want)
			}
		})
	}
}

func TestGetToolVersion_UnknownTool(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		envVar   string
		envVal   string
		want     string
	}{
		{
			name:     "unknown tool from constructed env var",
			toolName: "custom-tool",
			envVar:   "MAGE_X_CUSTOM_TOOL_VERSION",
			envVal:   "v2.0.0",
			want:     "v2.0.0",
		},
		{
			name:     "unknown tool defaults to latest",
			toolName: "unknown-tool",
			want:     VersionLatest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVar != "" {
				t.Setenv(tt.envVar, tt.envVal)
			}

			got := GetToolVersion(tt.toolName)
			if got != tt.want {
				t.Errorf("GetToolVersion(%q) = %q, want %q", tt.toolName, got, tt.want)
			}
		})
	}
}

func TestGetToolVersion_PrimaryTakesPrecedence(t *testing.T) {
	t.Setenv("MAGE_X_GOLANGCI_LINT_VERSION", "v1.55.0")
	t.Setenv("GOLANGCI_LINT_VERSION", "v1.50.0")

	got := GetToolVersion("golangci-lint")
	want := "v1.55.0"

	if got != want {
		t.Errorf("GetToolVersion(\"golangci-lint\") = %q, want %q (primary should take precedence)", got, want)
	}
}

func TestGetDefaultVersionFunctions_BackwardCompat(t *testing.T) {
	// Set up env vars for each tool
	t.Setenv("MAGE_X_GOLANGCI_LINT_VERSION", "v1.55.0")
	t.Setenv("MAGE_X_GOFUMPT_VERSION", "v0.5.0")
	t.Setenv("MAGE_X_YAMLFMT_VERSION", "v0.10.0")
	t.Setenv("MAGE_X_GOVULNCHECK_VERSION", "v1.0.0")
	t.Setenv("MAGE_X_MOCKGEN_VERSION", "v0.4.0")
	t.Setenv("MAGE_X_SWAG_VERSION", "v1.16.0")
	t.Setenv("MAGE_X_GO_VERSION", "1.24")
	t.Setenv("MAGE_X_GO_SECONDARY_VERSION", "1.23")

	tests := []struct {
		name     string
		fn       func() string
		toolName string
	}{
		{"GetDefaultGolangciLintVersion", GetDefaultGolangciLintVersion, "golangci-lint"},
		{"GetDefaultGofumptVersion", GetDefaultGofumptVersion, "gofumpt"},
		{"GetDefaultYamlfmtVersion", GetDefaultYamlfmtVersion, "yamlfmt"},
		{"GetDefaultGoVulnCheckVersion", GetDefaultGoVulnCheckVersion, "govulncheck"},
		{"GetDefaultMockgenVersion", GetDefaultMockgenVersion, "mockgen"},
		{"GetDefaultSwagVersion", GetDefaultSwagVersion, "swag"},
		{"GetDefaultGoVersion", GetDefaultGoVersion, "go"},
		{"GetSecondaryGoVersion", GetSecondaryGoVersion, "go-secondary"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapper := tt.fn()
			direct := GetToolVersion(tt.toolName)

			if wrapper != direct {
				t.Errorf("%s() = %q, GetToolVersion(%q) = %q; should be equal",
					tt.name, wrapper, tt.toolName, direct)
			}
		})
	}
}

func TestGetToolConfig(t *testing.T) {
	tests := []struct {
		name       string
		toolName   string
		wantExists bool
	}{
		{"known tool", "golangci-lint", true},
		{"known tool go", "go", true},
		{"unknown tool", "nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, exists := GetToolConfig(tt.toolName)
			if exists != tt.wantExists {
				t.Errorf("GetToolConfig(%q) exists = %v, want %v", tt.toolName, exists, tt.wantExists)
			}
			if exists && cfg.Name != tt.toolName {
				t.Errorf("GetToolConfig(%q).Name = %q, want %q", tt.toolName, cfg.Name, tt.toolName)
			}
		})
	}
}

func TestListRegisteredTools(t *testing.T) {
	tools := ListRegisteredTools()

	// Should have all expected tools
	expectedTools := []string{
		"golangci-lint", "gofumpt", "yamlfmt", "govulncheck",
		"nancy", "gitleaks", "staticcheck", "mockgen", "swag",
		"goreleaser", "go", "go-secondary",
	}

	if len(tools) != len(expectedTools) {
		t.Errorf("ListRegisteredTools() returned %d tools, want %d", len(tools), len(expectedTools))
	}

	// Sort both for comparison
	sort.Strings(tools)
	sort.Strings(expectedTools)

	for i, expected := range expectedTools {
		if i >= len(tools) || tools[i] != expected {
			t.Errorf("ListRegisteredTools() missing or wrong tool at index %d: got %v, want %s",
				i, tools, expected)
			break
		}
	}
}

func TestStripSuffix(t *testing.T) {
	tests := []struct {
		version string
		suffix  string
		want    string
	}{
		{"1.24.x", ".x", "1.24"},
		{"1.24", ".x", "1.24"},
		{"1.24.1", ".x", "1.24.1"},
		{"v1.0.0", "", "v1.0.0"},
		{"", ".x", ""},
	}

	for _, tt := range tests {
		t.Run(tt.version+"_"+tt.suffix, func(t *testing.T) {
			got := stripSuffix(tt.version, tt.suffix)
			if got != tt.want {
				t.Errorf("stripSuffix(%q, %q) = %q, want %q", tt.version, tt.suffix, got, tt.want)
			}
		})
	}
}

func TestGetToolVersion_CleanValue(t *testing.T) {
	// Test that inline comments are stripped (via env.MustGet)
	t.Setenv("MAGE_X_GOLANGCI_LINT_VERSION", "v1.55.0  # this is a comment")

	got := GetToolVersion("golangci-lint")
	want := "v1.55.0"

	if got != want {
		t.Errorf("GetToolVersion with comment = %q, want %q", got, want)
	}
}

func TestWithWarning(t *testing.T) {
	// Clear env vars to trigger warning path using t.Setenv for proper cleanup
	t.Setenv("MAGE_X_GOLANGCI_LINT_VERSION", "")
	t.Setenv("GOLANGCI_LINT_VERSION", "")

	// This should not panic and should return default
	got := GetToolVersion("golangci-lint", WithWarning(true))
	if got != VersionLatest {
		t.Errorf("GetToolVersion with warning = %q, want %q", got, VersionLatest)
	}
}
