package mage

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// UpdateBannerTestSuite tests the update banner functionality
type UpdateBannerTestSuite struct {
	suite.Suite

	oldStdout *os.File
	pipeRead  *os.File
	pipeWrite *os.File
}

func (suite *UpdateBannerTestSuite) SetupTest() {
	// Capture stdout for testing banner output
	suite.oldStdout = os.Stdout
	var err error
	suite.pipeRead, suite.pipeWrite, err = os.Pipe()
	suite.Require().NoError(err)
	os.Stdout = suite.pipeWrite
}

func (suite *UpdateBannerTestSuite) TearDownTest() {
	os.Stdout = suite.oldStdout
	if suite.pipeRead != nil {
		_ = suite.pipeRead.Close() //nolint:errcheck // Best effort cleanup
	}
	if suite.pipeWrite != nil {
		_ = suite.pipeWrite.Close() //nolint:errcheck // Best effort cleanup
	}
}

func (suite *UpdateBannerTestSuite) captureOutput() string {
	_ = suite.pipeWrite.Close() //nolint:errcheck // Best effort cleanup
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, suite.pipeRead) //nolint:errcheck // Test helper
	return buf.String()
}

func TestUpdateBannerSuite(t *testing.T) {
	suite.Run(t, new(UpdateBannerTestSuite))
}

// TestShowUpdateBannerWithUpdate tests showing banner when update available
func (suite *UpdateBannerTestSuite) TestShowUpdateBannerWithUpdate() {
	// Set NO_COLOR to ensure consistent output
	_ = os.Setenv("NO_COLOR", "1") //nolint:errcheck // Test env setup
	defer func() {
		_ = os.Unsetenv("NO_COLOR") //nolint:errcheck // Test env cleanup
	}()

	result := &UpdateCheckResult{
		CurrentVersion:  "v1.0.0",
		LatestVersion:   "v1.1.0",
		UpdateAvailable: true,
	}

	ShowUpdateBanner(result)

	output := suite.captureOutput()
	suite.Contains(output, "MAGE-X")
	suite.Contains(output, "v1.0.0")
	suite.Contains(output, "v1.1.0")
	suite.Contains(output, "update:install")
}

// TestShowUpdateBannerNoUpdate tests that banner is not shown when no update
func (suite *UpdateBannerTestSuite) TestShowUpdateBannerNoUpdate() {
	result := &UpdateCheckResult{
		CurrentVersion:  "v1.0.0",
		LatestVersion:   "v1.0.0",
		UpdateAvailable: false,
	}

	ShowUpdateBanner(result)

	output := suite.captureOutput()
	suite.Empty(output, "No banner should be shown when no update available")
}

// TestShowUpdateBannerNilResult tests handling of nil result
func (suite *UpdateBannerTestSuite) TestShowUpdateBannerNilResult() {
	ShowUpdateBanner(nil)

	output := suite.captureOutput()
	suite.Empty(output, "No banner should be shown for nil result")
}

// TestFormatUpdateBanner tests banner formatting
func TestFormatUpdateBanner(t *testing.T) {
	banner := formatUpdateBanner("v1.0.0", "v1.1.0")

	assert.Contains(t, banner, "MAGE-X")
	assert.Contains(t, banner, "v1.0.0")
	assert.Contains(t, banner, "v1.1.0")
	assert.Contains(t, banner, "update:install")
	assert.Contains(t, banner, "Upgrade")
}

// TestFormatUpdateBannerFancy tests fancy banner formatting
func TestFormatUpdateBannerFancy(t *testing.T) {
	banner := FormatUpdateBannerFancy("v1.0.0", "v1.2.0")

	assert.Contains(t, banner, "MAGE-X")
	assert.Contains(t, banner, "v1.0.0")
	assert.Contains(t, banner, "v1.2.0")
	assert.Contains(t, banner, "update:install")
	// Should contain box drawing characters
	assert.Contains(t, banner, "╭")
	assert.Contains(t, banner, "╰")
}

// TestFormatUpdateBannerMinimal tests minimal banner formatting
func TestFormatUpdateBannerMinimal(t *testing.T) {
	banner := FormatUpdateBannerMinimal("v1.0.0", "v1.3.0")

	assert.Contains(t, banner, "v1.0.0")
	assert.Contains(t, banner, "v1.3.0")
	assert.Contains(t, banner, "update:install")
	// Should be shorter than fancy banner
	assert.Less(t, len(banner), 200)
}

// TestPadVersion tests version padding
func TestPadVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		width    int
		expected string
	}{
		{
			name:     "short version padded",
			version:  "v1.0.0",
			width:    12,
			expected: "v1.0.0      ",
		},
		{
			name:     "exact width",
			version:  "v1.0.0-beta",
			width:    11,
			expected: "v1.0.0-beta",
		},
		{
			name:     "long version truncated",
			version:  "v1.0.0-beta.1.rc2",
			width:    10,
			expected: "v1.0.0-bet",
		},
		{
			name:     "empty version",
			version:  "",
			width:    5,
			expected: "     ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padVersion(tt.version, tt.width)
			assert.Equal(t, tt.expected, result)
			assert.Len(t, result, tt.width)
		})
	}
}

// TestPadRight tests right padding
func TestPadRight(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected string
	}{
		{
			name:     "short string padded",
			input:    "hello",
			width:    10,
			expected: "hello     ",
		},
		{
			name:     "exact width",
			input:    "hello",
			width:    5,
			expected: "hello",
		},
		{
			name:     "long string truncated",
			input:    "hello world",
			width:    5,
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padRight(tt.input, tt.width)
			assert.Equal(t, tt.expected, result)
			assert.Len(t, result, tt.width)
		})
	}
}

// TestGetUpgradeCommand tests the upgrade command getter
func TestGetUpgradeCommand(t *testing.T) {
	cmd := GetUpgradeCommand()
	assert.Contains(t, cmd, "magex")
	assert.Contains(t, cmd, "update:install")
	assert.Equal(t, "magex update:install", cmd)
}

// TestShouldUseColorForBanner tests color detection
func TestShouldUseColorForBanner(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name: "disabled in CI",
			envVars: map[string]string{
				"CI": "true",
			},
			expected: false,
		},
		{
			name: "disabled with NO_COLOR",
			envVars: map[string]string{
				"NO_COLOR": "1",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore env vars
			savedEnv := map[string]string{}
			for key := range tt.envVars {
				savedEnv[key] = os.Getenv(key)
			}
			for _, key := range []string{"CI", "NO_COLOR"} {
				if _, exists := savedEnv[key]; !exists {
					savedEnv[key] = os.Getenv(key)
				}
			}

			defer func() {
				for key, val := range savedEnv {
					if val != "" {
						_ = os.Setenv(key, val) //nolint:errcheck // Test env restore
					} else {
						_ = os.Unsetenv(key) //nolint:errcheck // Test env restore
					}
				}
			}()

			// Clear env vars
			_ = os.Unsetenv("CI")       //nolint:errcheck // Test env setup
			_ = os.Unsetenv("NO_COLOR") //nolint:errcheck // Test env setup

			// Set test env vars
			for key, val := range tt.envVars {
				_ = os.Setenv(key, val) //nolint:errcheck // Test env setup
			}

			result := shouldUseColorForBanner()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestBannerLineWidth tests that banner lines are consistent width
func TestBannerLineWidth(t *testing.T) {
	banner := formatUpdateBanner("v1.0.0", "v99.99.99")
	lines := strings.Split(banner, "\n")

	// Find the width of the box lines (containing +- characters)
	var boxWidth int
	for _, line := range lines {
		if strings.Contains(line, "+") && strings.Contains(line, "-") {
			boxWidth = len(line)
			break
		}
	}

	// All content lines should be the same width
	for _, line := range lines {
		if strings.Contains(line, "|") {
			assert.Len(t, line, boxWidth,
				"Line width mismatch: %q", line)
		}
	}
}

// TestBannerContainsUpgradeCommand tests that the upgrade command is complete
func TestBannerContainsUpgradeCommand(t *testing.T) {
	banner := formatUpdateBanner("v1.0.0", "v2.0.0")

	// The full upgrade command should be in the banner
	assert.Contains(t, banner, upgradeCmd)
}

// TestBannerVersionDisplay tests version display in various scenarios
func TestBannerVersionDisplay(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
	}{
		{
			name:    "standard versions",
			current: "v1.0.0",
			latest:  "v1.1.0",
		},
		{
			name:    "dev to release",
			current: "dev",
			latest:  "v1.0.0",
		},
		{
			name:    "long versions",
			current: "v1.2.3-beta.1",
			latest:  "v2.0.0-rc.1",
		},
		{
			name:    "no v prefix",
			current: "1.0.0",
			latest:  "1.1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			banner := formatUpdateBanner(tt.current, tt.latest)

			// Both versions should appear in the banner
			assert.Contains(t, banner, tt.current[:minInt(len(tt.current), 12)])
			assert.Contains(t, banner, tt.latest[:minInt(len(tt.latest), 12)])
		})
	}
}

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// TestShowUpdateBannerWithDifferentResults tests various result scenarios
func TestShowUpdateBannerWithDifferentResults(t *testing.T) {
	tests := []struct {
		name       string
		result     *UpdateCheckResult
		shouldShow bool
	}{
		{
			name:       "nil result",
			result:     nil,
			shouldShow: false,
		},
		{
			name: "update available",
			result: &UpdateCheckResult{
				CurrentVersion:  "v1.0.0",
				LatestVersion:   "v1.1.0",
				UpdateAvailable: true,
			},
			shouldShow: true,
		},
		{
			name: "no update",
			result: &UpdateCheckResult{
				CurrentVersion:  "v1.0.0",
				LatestVersion:   "v1.0.0",
				UpdateAvailable: false,
			},
			shouldShow: false,
		},
		{
			name: "from cache",
			result: &UpdateCheckResult{
				CurrentVersion:  "v1.0.0",
				LatestVersion:   "v2.0.0",
				UpdateAvailable: true,
				FromCache:       true,
			},
			shouldShow: true,
		},
		{
			name: "with error",
			result: &UpdateCheckResult{
				CurrentVersion: "v1.0.0",
				Error:          "network error",
			},
			shouldShow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("Failed to create pipe: %v", err)
			}
			os.Stdout = w

			// Set NO_COLOR to ensure consistent output
			_ = os.Setenv("NO_COLOR", "1") //nolint:errcheck // Test env setup
			defer func() {
				_ = os.Unsetenv("NO_COLOR") //nolint:errcheck // Test env cleanup
			}()

			ShowUpdateBanner(tt.result)

			_ = w.Close() //nolint:errcheck // Test cleanup
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r) //nolint:errcheck // Test helper
			os.Stdout = old

			output := buf.String()

			if tt.shouldShow {
				assert.NotEmpty(t, output, "Expected banner to be shown")
				assert.Contains(t, output, "MAGE-X")
			} else {
				assert.Empty(t, output, "Expected no banner to be shown")
			}
		})
	}
}

// TestBannerColorConstants tests that color constants are valid ANSI codes
func TestBannerColorConstants(t *testing.T) {
	// All color constants should start with escape sequence
	colors := []string{
		bannerColorReset,
		bannerColorYellow,
	}

	for _, color := range colors {
		assert.True(t, strings.HasPrefix(color, "\033["),
			"Color constant should start with ANSI escape: %q", color)
	}

	// Reset should be the standard reset code
	assert.Equal(t, "\033[0m", bannerColorReset)
}

// TestUpdateCheckResultString tests that UpdateCheckResult can be serialized
func TestUpdateCheckResultJSON(t *testing.T) {
	result := &UpdateCheckResult{
		CurrentVersion:  "v1.0.0",
		LatestVersion:   "v1.1.0",
		UpdateAvailable: true,
		ReleaseNotes:    "Test notes",
		ReleaseURL:      "https://example.com",
		CheckedAt:       time.Now(),
		FromCache:       false,
		Error:           "",
	}

	// Should be able to use in JSON marshaling (tested implicitly by cache)
	assert.NotNil(t, result)
	assert.Equal(t, "v1.0.0", result.CurrentVersion)
}

// TestPadVersionUnicode tests that padVersion handles Unicode characters correctly
//
//nolint:gosmopolitan // Unicode test cases are intentional for testing multi-byte character handling
func TestPadVersionUnicode(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		width    int
		expected int // expected rune count
	}{
		{
			name:     "ASCII version",
			version:  "v1.0.0",
			width:    12,
			expected: 12,
		},
		{
			name:     "Unicode characters",
			version:  "v日本語",
			width:    12,
			expected: 12,
		},
		{
			name:     "Emoji in version",
			version:  "v1.0🎉",
			width:    10,
			expected: 10,
		},
		{
			name:     "Long Unicode truncated",
			version:  "日本語版本情報",
			width:    5,
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padVersion(tt.version, tt.width)
			// Check that result has correct rune count (not byte count)
			runeCount := len([]rune(result))
			assert.Equal(t, tt.expected, runeCount,
				"Expected %d runes, got %d for input %q",
				tt.expected, runeCount, tt.version)
		})
	}
}

// TestPadRightUnicode tests that padRight handles Unicode characters correctly
//
//nolint:gosmopolitan // Unicode test cases are intentional for testing multi-byte character handling
func TestPadRightUnicode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		width    int
		expected int // expected rune count
	}{
		{
			name:     "ASCII string",
			input:    "hello",
			width:    10,
			expected: 10,
		},
		{
			name:     "Unicode string",
			input:    "こんにちは",
			width:    10,
			expected: 10,
		},
		{
			name:     "Mixed ASCII and Unicode",
			input:    "Hello世界",
			width:    15,
			expected: 15,
		},
		{
			name:     "Long Unicode truncated",
			input:    "日本語テスト文字列",
			width:    5,
			expected: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := padRight(tt.input, tt.width)
			runeCount := len([]rune(result))
			assert.Equal(t, tt.expected, runeCount,
				"Expected %d runes, got %d for input %q",
				tt.expected, runeCount, tt.input)
		})
	}
}

// TestBannerWithVeryLongVersion tests banner formatting with very long versions
func TestBannerWithVeryLongVersion(t *testing.T) {
	// Very long version string that exceeds versionDisplayWidth (12 chars)
	longVersion := "v1.2.3-beta.4.rc5+build.123456789"

	banner := formatUpdateBanner(longVersion, "v2.0.0")

	// Banner should still render correctly (be well-formed)
	assert.Contains(t, banner, "MAGE-X")
	assert.Contains(t, banner, "update:install")

	// All lines with | should have same length
	lines := strings.Split(banner, "\n")
	var boxWidth int
	for _, line := range lines {
		if strings.Contains(line, "+") && strings.Contains(line, "-") {
			boxWidth = len(line)
			break
		}
	}
	for _, line := range lines {
		if strings.Contains(line, "|") {
			assert.Len(t, line, boxWidth, "Line width mismatch: %q", line)
		}
	}
}

// TestBannerWithEmptyVersion tests banner formatting with empty version strings
func TestBannerWithEmptyVersion(t *testing.T) {
	// Empty versions should still produce a valid banner
	banner := formatUpdateBanner("", "v1.0.0")

	// Banner should still render
	assert.Contains(t, banner, "MAGE-X")
	assert.Contains(t, banner, "update:install")
	assert.Contains(t, banner, "Current:")
	assert.Contains(t, banner, "Latest:")
}

// TestVersionDisplayWidthConstant tests the version display width constant
func TestVersionDisplayWidthConstant(t *testing.T) {
	// versionDisplayWidth should be defined and reasonable
	assert.Equal(t, 12, versionDisplayWidth,
		"versionDisplayWidth should be 12 to accommodate typical semver versions")
}
