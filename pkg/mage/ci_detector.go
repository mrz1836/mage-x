// Package mage provides CI environment detection for test output processing
package mage

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// CIDetector detects CI environment and determines mode
type CIDetector interface {
	// IsCI returns true if running in a CI environment
	IsCI() bool

	// Platform returns the detected CI platform (github, gitlab, etc.)
	Platform() string

	// GetConfig returns the effective CI configuration
	// Priority: explicit param > environment > config > default
	GetConfig(params map[string]string, cfg *Config) CIMode
}

// ciDetector implements CIDetector interface
type ciDetector struct {
	envGetter func(string) string
}

// NewCIDetector creates a new CI detector
func NewCIDetector() CIDetector {
	return &ciDetector{
		envGetter: os.Getenv,
	}
}

// newCIDetectorWithEnv creates a CI detector with custom environment getter (for testing)
func newCIDetectorWithEnv(envGetter func(string) string) CIDetector {
	return &ciDetector{
		envGetter: envGetter,
	}
}

// IsCI returns true if running in a CI environment
func (d *ciDetector) IsCI() bool {
	// Check generic CI environment variable first (covers most CI systems)
	if d.envGetter("CI") == trueValue {
		return true
	}

	// Check platform-specific variables
	ciVars := []string{
		"GITHUB_ACTIONS",   // GitHub Actions
		"GITLAB_CI",        // GitLab CI
		"CIRCLECI",         // CircleCI
		"TRAVIS",           // Travis CI
		"JENKINS_URL",      // Jenkins
		"TF_BUILD",         // Azure Pipelines
		"BUILDKITE",        // Buildkite
		"DRONE",            // Drone CI
		"CODEBUILD_CI",     // AWS CodeBuild
		"TEAMCITY_VERSION", // TeamCity
	}

	for _, v := range ciVars {
		if d.envGetter(v) != "" {
			return true
		}
	}

	// Check for Bitbucket Pipelines
	if d.envGetter("BITBUCKET_BUILD_NUMBER") != "" {
		return true
	}

	return false
}

// Platform returns the detected CI platform name
func (d *ciDetector) Platform() string {
	// Check in priority order (most specific first)
	if d.envGetter("GITHUB_ACTIONS") == trueValue {
		return string(CIFormatGitHub)
	}
	if d.envGetter("GITLAB_CI") == trueValue {
		return "gitlab"
	}
	if d.envGetter("CIRCLECI") == trueValue {
		return "circleci"
	}
	if d.envGetter("TRAVIS") == trueValue {
		return "travis"
	}
	if d.envGetter("JENKINS_URL") != "" {
		return "jenkins"
	}
	// Azure Pipelines sets TF_BUILD to "True" (capital T), unlike other CI systems
	// that typically use lowercase "true". This is documented behavior from Microsoft.
	if d.envGetter("TF_BUILD") == "True" {
		return "azure"
	}
	if d.envGetter("BUILDKITE") == trueValue {
		return "buildkite"
	}
	if d.envGetter("DRONE") == trueValue {
		return "drone"
	}
	if d.envGetter("CODEBUILD_CI") == trueValue {
		return "codebuild"
	}
	if d.envGetter("TEAMCITY_VERSION") != "" {
		return "teamcity"
	}
	if d.envGetter("BITBUCKET_BUILD_NUMBER") != "" {
		return "bitbucket"
	}

	// Generic CI detection
	if d.envGetter("CI") == trueValue {
		return "generic"
	}

	return "local"
}

// GetConfig returns the effective CI configuration
// Priority: explicit param > environment > config > default
func (d *ciDetector) GetConfig(params map[string]string, cfg *Config) CIMode {
	// Start with defaults
	mode := DefaultCIMode()

	// Apply config file settings if available
	if cfg != nil {
		mode = cfg.Test.CIMode
	}

	// Apply environment variable overrides
	d.applyEnvOverrides(&mode)

	// Apply explicit parameter overrides (highest priority)
	d.applyParamOverrides(&mode, params)

	// Auto-enable if CI is detected and not explicitly disabled
	if params == nil || params["ci"] == "" {
		if d.IsCI() && !mode.Enabled {
			mode.Enabled = true
		}
	}

	// Auto-select format based on platform if set to auto
	if mode.Format == CIFormatAuto {
		mode.Format = d.selectFormat()
	}

	return mode
}

// applyEnvOverrides applies environment variable overrides to CIMode
func (d *ciDetector) applyEnvOverrides(mode *CIMode) {
	// CI mode enable/disable
	if v := d.envGetter("MAGE_X_CI_MODE"); v != "" {
		switch strings.ToLower(v) {
		case trueValue, "on", "1", "auto":
			mode.Enabled = true
		case "false", "off", "0":
			mode.Enabled = false
		}
	}

	// Format override
	if v := d.envGetter("MAGE_X_CI_FORMAT"); v != "" {
		switch strings.ToLower(v) {
		case "github":
			mode.Format = CIFormatGitHub
		case "json":
			mode.Format = CIFormatJSON
		case "auto":
			mode.Format = CIFormatAuto
		}
	}

	// Context lines override
	if v := d.envGetter("MAGE_X_CI_CONTEXT"); v != "" {
		var lines int
		if err := parseIntEnv(v, &lines); err == nil && lines >= 0 && lines <= 100 {
			mode.ContextLines = lines
		}
	}

	// Max memory override
	if v := d.envGetter("MAGE_X_CI_MAX_MEMORY"); v != "" {
		// Support formats like "200MB", "200", "200mb"
		v = strings.ToUpper(strings.TrimSuffix(strings.TrimSuffix(v, "MB"), "mb"))
		var mb int
		if err := parseIntEnv(v, &mb); err == nil && mb >= 10 && mb <= 1000 {
			mode.MaxMemoryMB = mb
		}
	}

	// Output path override
	if v := d.envGetter("MAGE_X_CI_OUTPUT"); v != "" {
		mode.OutputPath = v
	}

	// Dedup override
	if v := d.envGetter("MAGE_X_CI_DEDUP"); v != "" {
		switch strings.ToLower(v) {
		case trueValue, "on", "1":
			mode.Dedup = true
		case "false", "off", "0":
			mode.Dedup = false
		}
	}
}

// applyParamOverrides applies explicit parameter overrides to CIMode
func (d *ciDetector) applyParamOverrides(mode *CIMode, params map[string]string) {
	if params == nil {
		return
	}

	// CI mode enable/disable
	if v, ok := params["ci"]; ok {
		switch strings.ToLower(v) {
		case trueValue, "on", "1", "":
			// Empty string means `ci` was passed without value (e.g., `magex test:unit ci`)
			mode.Enabled = true
		case "false", "off", "0":
			mode.Enabled = false
		}
	}

	// Format override
	if v, ok := params["ci_format"]; ok {
		switch strings.ToLower(v) {
		case "github":
			mode.Format = CIFormatGitHub
		case "json":
			mode.Format = CIFormatJSON
		case "auto":
			mode.Format = CIFormatAuto
		}
	}
}

// selectFormat returns the appropriate format based on detected platform
func (d *ciDetector) selectFormat() CIFormat {
	switch d.Platform() {
	case "github":
		return CIFormatGitHub
	default:
		return CIFormatJSON
	}
}

// GetMetadata returns CI metadata from environment
func (d *ciDetector) GetMetadata() CIMetadata {
	return CIMetadata{
		Branch:    d.getBranch(),
		Commit:    d.getCommit(),
		RunID:     d.getRunID(),
		Workflow:  d.getWorkflow(),
		Platform:  d.Platform(),
		GoVersion: runtime.Version(),
	}
}

// getBranch returns the current branch from CI environment
func (d *ciDetector) getBranch() string {
	// Try platform-specific variables first
	if v := d.envGetter("GITHUB_REF_NAME"); v != "" {
		return v
	}
	if v := d.envGetter("GITHUB_HEAD_REF"); v != "" {
		return v
	}
	if v := d.envGetter("CI_COMMIT_REF_NAME"); v != "" { // GitLab
		return v
	}
	if v := d.envGetter("CIRCLE_BRANCH"); v != "" { // CircleCI
		return v
	}
	if v := d.envGetter("TRAVIS_BRANCH"); v != "" { // Travis
		return v
	}
	if v := d.envGetter("GIT_BRANCH"); v != "" { // Jenkins
		return v
	}
	if v := d.envGetter("BUILD_SOURCEBRANCH"); v != "" { // Azure
		// Azure uses refs/heads/main format
		return strings.TrimPrefix(v, "refs/heads/")
	}
	if v := d.envGetter("BITBUCKET_BRANCH"); v != "" { // Bitbucket
		return v
	}
	return ""
}

// getCommit returns the current commit SHA from CI environment
func (d *ciDetector) getCommit() string {
	if v := d.envGetter("GITHUB_SHA"); v != "" {
		return v
	}
	if v := d.envGetter("CI_COMMIT_SHA"); v != "" { // GitLab
		return v
	}
	if v := d.envGetter("CIRCLE_SHA1"); v != "" { // CircleCI
		return v
	}
	if v := d.envGetter("TRAVIS_COMMIT"); v != "" { // Travis
		return v
	}
	if v := d.envGetter("GIT_COMMIT"); v != "" { // Jenkins
		return v
	}
	if v := d.envGetter("BUILD_SOURCEVERSION"); v != "" { // Azure
		return v
	}
	if v := d.envGetter("BITBUCKET_COMMIT"); v != "" { // Bitbucket
		return v
	}
	return ""
}

// getRunID returns the CI run ID from environment
func (d *ciDetector) getRunID() string {
	if v := d.envGetter("GITHUB_RUN_ID"); v != "" {
		return v
	}
	if v := d.envGetter("CI_PIPELINE_ID"); v != "" { // GitLab
		return v
	}
	if v := d.envGetter("CIRCLE_BUILD_NUM"); v != "" { // CircleCI
		return v
	}
	if v := d.envGetter("TRAVIS_BUILD_NUMBER"); v != "" { // Travis
		return v
	}
	if v := d.envGetter("BUILD_NUMBER"); v != "" { // Jenkins
		return v
	}
	if v := d.envGetter("BUILD_BUILDNUMBER"); v != "" { // Azure
		return v
	}
	if v := d.envGetter("BITBUCKET_BUILD_NUMBER"); v != "" { // Bitbucket
		return v
	}
	return ""
}

// getWorkflow returns the workflow/job name from CI environment
func (d *ciDetector) getWorkflow() string {
	if v := d.envGetter("GITHUB_WORKFLOW"); v != "" {
		return v
	}
	if v := d.envGetter("CI_JOB_NAME"); v != "" { // GitLab
		return v
	}
	if v := d.envGetter("CIRCLE_JOB"); v != "" { // CircleCI
		return v
	}
	if v := d.envGetter("TRAVIS_JOB_NAME"); v != "" { // Travis
		return v
	}
	if v := d.envGetter("JOB_NAME"); v != "" { // Jenkins
		return v
	}
	if v := d.envGetter("BUILD_DEFINITIONNAME"); v != "" { // Azure
		return v
	}
	return ""
}

// parseIntEnv parses an integer from string, returns error if invalid
func parseIntEnv(s string, result *int) error {
	var val int
	_, err := fmt.Sscanf(s, "%d", &val)
	if err == nil {
		*result = val
	}
	return err
}
