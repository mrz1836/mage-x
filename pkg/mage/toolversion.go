// Package mage provides tool version resolution utilities.
package mage

import (
	"strings"

	"github.com/mrz1836/mage-x/pkg/common/env"
	"github.com/mrz1836/mage-x/pkg/utils"
)

// ToolConfig defines configuration for a single tool's version resolution.
type ToolConfig struct {
	Name         string // Tool command name (e.g., "golangci-lint")
	PrimaryEnv   string // Primary MAGE_X env var (e.g., "MAGE_X_GOLANGCI_LINT_VERSION")
	LegacyEnv    string // Legacy env var for backward compat (e.g., "GOLANGCI_LINT_VERSION")
	DefaultValue string // Fallback if no env var set
	StripSuffix  string // Suffix to strip (e.g., ".x" for Go versions)
}

// toolRegistry is the canonical registry of all tool version configurations.
//
//nolint:gochecknoglobals // Required for package-level configuration registry
var toolRegistry = map[string]ToolConfig{
	// Core linting tools
	"golangci-lint": {
		Name:         "golangci-lint",
		PrimaryEnv:   "MAGE_X_GOLANGCI_LINT_VERSION",
		LegacyEnv:    "GOLANGCI_LINT_VERSION",
		DefaultValue: VersionLatest,
	},
	"gofumpt": {
		Name:         "gofumpt",
		PrimaryEnv:   "MAGE_X_GOFUMPT_VERSION",
		LegacyEnv:    "GOFUMPT_VERSION",
		DefaultValue: VersionLatest,
	},
	"yamlfmt": {
		Name:         "yamlfmt",
		PrimaryEnv:   "MAGE_X_YAMLFMT_VERSION",
		LegacyEnv:    "YAMLFMT_VERSION",
		DefaultValue: VersionLatest,
	},

	// Security scanning tools
	"govulncheck": {
		Name:         "govulncheck",
		PrimaryEnv:   "MAGE_X_GOVULNCHECK_VERSION",
		LegacyEnv:    "GOVULNCHECK_VERSION",
		DefaultValue: VersionLatest,
	},
	"nancy": {
		Name:         "nancy",
		PrimaryEnv:   "MAGE_X_NANCY_VERSION",
		LegacyEnv:    "NANCY_VERSION",
		DefaultValue: VersionLatest,
	},
	"gitleaks": {
		Name:         "gitleaks",
		PrimaryEnv:   "MAGE_X_GITLEAKS_VERSION",
		LegacyEnv:    "GITLEAKS_VERSION",
		DefaultValue: VersionLatest,
	},
	"staticcheck": {
		Name:         "staticcheck",
		PrimaryEnv:   "MAGE_X_STATICCHECK_VERSION",
		LegacyEnv:    "", // No legacy env var
		DefaultValue: VersionLatest,
	},

	// Code generation tools
	"mockgen": {
		Name:         "mockgen",
		PrimaryEnv:   "MAGE_X_MOCKGEN_VERSION",
		LegacyEnv:    "MOCKGEN_VERSION",
		DefaultValue: VersionLatest,
	},
	"swag": {
		Name:         "swag",
		PrimaryEnv:   "MAGE_X_SWAG_VERSION",
		LegacyEnv:    "SWAG_VERSION",
		DefaultValue: VersionLatest,
	},

	// Release tools
	"goreleaser": {
		Name:         "goreleaser",
		PrimaryEnv:   "MAGE_X_GORELEASER_VERSION",
		LegacyEnv:    "GORELEASER_VERSION",
		DefaultValue: VersionLatest,
	},

	// Go runtime versions (with special .x suffix handling)
	"go": {
		Name:         "go",
		PrimaryEnv:   "MAGE_X_GO_VERSION",
		LegacyEnv:    "GO_PRIMARY_VERSION",
		DefaultValue: "1.24",
		StripSuffix:  ".x",
	},
	"go-secondary": {
		Name:         "go-secondary",
		PrimaryEnv:   "MAGE_X_GO_SECONDARY_VERSION",
		LegacyEnv:    "GO_SECONDARY_VERSION",
		DefaultValue: "1.23",
		StripSuffix:  ".x",
	},
}

// ToolVersionOption allows customizing version lookup behavior.
type ToolVersionOption func(*toolVersionOptions)

type toolVersionOptions struct {
	warnOnMissing bool
}

// WithWarning enables warning output when version is not found in environment.
func WithWarning(enabled bool) ToolVersionOption {
	return func(o *toolVersionOptions) {
		o.warnOnMissing = enabled
	}
}

// GetToolVersion returns the version for a tool from environment variables.
// It checks primary env var, then legacy env var, then returns the default.
//
// Lookup order:
//  1. Primary env var (e.g., MAGE_X_GOLANGCI_LINT_VERSION)
//  2. Legacy env var (e.g., GOLANGCI_LINT_VERSION) - if defined
//  3. Default value (typically "latest")
//
// Example:
//
//	version := GetToolVersion("golangci-lint")
//	version := GetToolVersion("go", WithWarning(true))
func GetToolVersion(toolName string, opts ...ToolVersionOption) string {
	options := &toolVersionOptions{
		warnOnMissing: false,
	}
	for _, opt := range opts {
		opt(options)
	}

	cfg, exists := toolRegistry[toolName]
	if !exists {
		// For unknown tools, construct env var name dynamically
		return getUnknownToolVersion(toolName, options)
	}

	return resolveToolVersion(cfg, options)
}

// resolveToolVersion implements the lookup chain for a known tool.
func resolveToolVersion(cfg ToolConfig, options *toolVersionOptions) string {
	// Try primary env var
	if version := env.MustGet(cfg.PrimaryEnv); version != "" {
		return stripSuffix(version, cfg.StripSuffix)
	}

	// Try legacy env var (if defined)
	if cfg.LegacyEnv != "" {
		if version := env.MustGet(cfg.LegacyEnv); version != "" {
			return stripSuffix(version, cfg.StripSuffix)
		}
	}

	// Warn if configured and return default
	if options.warnOnMissing {
		utils.Warn("Tool version for %s not found in environment variables (%s)", cfg.Name, cfg.PrimaryEnv)
		utils.Warn("Consider sourcing .github/.env.base: source .github/.env.base")
		utils.Warn("Using fallback version: %s", cfg.DefaultValue)
	}

	return cfg.DefaultValue
}

// getUnknownToolVersion handles tools not in the registry.
func getUnknownToolVersion(toolName string, options *toolVersionOptions) string {
	envVar := "MAGE_X_" + strings.ToUpper(strings.ReplaceAll(toolName, "-", "_")) + "_VERSION"

	if version := env.MustGet(envVar); version != "" {
		return version
	}

	if options.warnOnMissing {
		utils.Warn("Tool version for %s not found in environment variables (%s)", toolName, envVar)
		utils.Warn("Consider sourcing .github/.env.base: source .github/.env.base")
	}

	return VersionLatest
}

// stripSuffix removes a suffix from version string if present.
func stripSuffix(version, suffix string) string {
	if suffix == "" {
		return version
	}
	if strings.HasSuffix(version, suffix) {
		return version[:len(version)-len(suffix)]
	}
	return version
}

// GetToolConfig returns the configuration for a tool (for advanced use cases).
func GetToolConfig(toolName string) (ToolConfig, bool) {
	cfg, exists := toolRegistry[toolName]
	return cfg, exists
}

// ListRegisteredTools returns all tools in the registry.
func ListRegisteredTools() []string {
	tools := make([]string, 0, len(toolRegistry))
	for name := range toolRegistry {
		tools = append(tools, name)
	}
	return tools
}

// --- Convenience functions for backward compatibility ---

// GetDefaultGolangciLintVersion returns the golangci-lint version.
func GetDefaultGolangciLintVersion() string { return GetToolVersion("golangci-lint") }

// GetDefaultGofumptVersion returns the gofumpt version.
func GetDefaultGofumptVersion() string { return GetToolVersion("gofumpt") }

// GetDefaultYamlfmtVersion returns the yamlfmt version.
func GetDefaultYamlfmtVersion() string { return GetToolVersion("yamlfmt") }

// GetDefaultGoVulnCheckVersion returns the govulncheck version.
func GetDefaultGoVulnCheckVersion() string { return GetToolVersion("govulncheck") }

// GetDefaultMockgenVersion returns the mockgen version.
func GetDefaultMockgenVersion() string { return GetToolVersion("mockgen") }

// GetDefaultSwagVersion returns the swag version.
func GetDefaultSwagVersion() string { return GetToolVersion("swag") }

// GetDefaultGoVersion returns the primary Go version.
func GetDefaultGoVersion() string { return GetToolVersion("go") }

// GetSecondaryGoVersion returns the secondary Go version.
func GetSecondaryGoVersion() string { return GetToolVersion("go-secondary") }

// GetToolVersionWithWarning is a convenience function that returns version with warning enabled.
// This matches the behavior of the old getToolVersionOrWarn function.
func GetToolVersionWithWarning(toolName string) string {
	return GetToolVersion(toolName, WithWarning(true))
}
