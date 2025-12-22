package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/testhelpers"
)

// VulncheckTestSuite tests vulnerability checking functionality
type VulncheckTestSuite struct {
	testhelpers.BaseSuite
}

func TestVulncheckTestSuite(t *testing.T) {
	suite.Run(t, new(VulncheckTestSuite))
}

// Sample govulncheck JSON output for testing
const sampleGovulncheckJSON = `{"config":{"protocol_version":"v1.0.0","scanner_name":"govulncheck","scanner_version":"v1.1.4","db":"https://vuln.go.dev","go_version":"go1.21.0","scan_level":"symbol"}}
{"progress":{"message":"Scanning your code and 48 packages across 1 dependent module for known vulnerabilities..."}}
{"osv":{"id":"GO-2024-1234","aliases":["CVE-2024-38513"],"summary":"Test vulnerability 1","details":"This is a test vulnerability"}}
{"osv":{"id":"GO-2023-5678","aliases":["CVE-2023-45142","GHSA-xxxx-yyyy-zzzz"],"summary":"Test vulnerability 2","details":"Another test vulnerability"}}
{"osv":{"id":"GO-2024-9999","aliases":["CVE-2024-99999"],"summary":"Unexcluded vulnerability","details":"This should not be excluded"}}
{"finding":{"osv":"GO-2024-1234","fixed_version":"v1.2.3","trace":[{"module":"github.com/example/vuln","version":"v1.0.0","package":"github.com/example/vuln","function":"VulnFunc"}]}}
{"finding":{"osv":"GO-2023-5678","fixed_version":"v2.0.0","trace":[{"module":"github.com/another/vuln","version":"v0.9.0","package":"github.com/another/vuln","function":"BadFunc"}]}}
{"finding":{"osv":"GO-2024-9999","fixed_version":"v3.0.0","trace":[{"module":"github.com/third/vuln","version":"v0.5.0","package":"github.com/third/vuln","function":"ThirdFunc"}]}}`

// Sample JSON with no vulnerabilities
const sampleNoVulnsJSON = `{"config":{"protocol_version":"v1.0.0","scanner_name":"govulncheck","scanner_version":"v1.1.4","db":"https://vuln.go.dev","go_version":"go1.21.0","scan_level":"symbol"}}
{"progress":{"message":"Scanning your code..."}}`

// Sample JSON with empty output
const sampleEmptyJSON = ``

// TestParseGovulncheckJSON tests JSON parsing functionality
func (s *VulncheckTestSuite) TestParseGovulncheckJSON() {
	result, err := ParseGovulncheckJSON(sampleGovulncheckJSON)
	s.Require().NoError(err)

	// Check config was parsed
	s.NotNil(result.Config)
	s.Equal("v1.0.0", result.Config.ProtocolVersion)
	s.Equal("govulncheck", result.Config.ScannerName)

	// Check progress messages were captured
	s.Len(result.ProgressMsgs, 1)

	// Check OSV entries were parsed
	s.Len(result.OSVEntries, 3)
	s.Contains(result.OSVEntries, "GO-2024-1234")
	s.Contains(result.OSVEntries, "GO-2023-5678")
	s.Contains(result.OSVEntries, "GO-2024-9999")

	// Check aliases
	osv1 := result.OSVEntries["GO-2024-1234"]
	s.Contains(osv1.Aliases, "CVE-2024-38513")

	osv2 := result.OSVEntries["GO-2023-5678"]
	s.Contains(osv2.Aliases, "CVE-2023-45142")
	s.Contains(osv2.Aliases, "GHSA-xxxx-yyyy-zzzz")

	// Check findings were parsed
	s.Len(result.Findings, 3)
}

// TestParseGovulncheckJSON_NoVulns tests parsing output with no vulnerabilities
func (s *VulncheckTestSuite) TestParseGovulncheckJSON_NoVulns() {
	result, err := ParseGovulncheckJSON(sampleNoVulnsJSON)
	s.Require().NoError(err)

	s.NotNil(result.Config)
	s.Empty(result.OSVEntries)
	s.Empty(result.Findings)
}

// TestParseGovulncheckJSON_Empty tests parsing empty output
func (s *VulncheckTestSuite) TestParseGovulncheckJSON_Empty() {
	result, err := ParseGovulncheckJSON(sampleEmptyJSON)
	s.Require().NoError(err)

	s.Nil(result.Config)
	s.Empty(result.OSVEntries)
	s.Empty(result.Findings)
}

// TestFilterExcludedVulns tests CVE filtering
func (s *VulncheckTestSuite) TestFilterExcludedVulns() {
	result, err := ParseGovulncheckJSON(sampleGovulncheckJSON)
	s.Require().NoError(err)

	// Test excluding CVE-2024-38513
	excludes := []string{"CVE-2024-38513"}
	filtered := FilterExcludedVulns(result, excludes)

	// Should have excluded one CVE
	s.Len(filtered.ExcludedCVEs, 1)
	s.Contains(filtered.ExcludedCVEs, "CVE-2024-38513")

	// Should have 2 remaining findings (for GO-2023-5678 and GO-2024-9999)
	s.Len(filtered.RemainingFindings, 2)
}

// TestFilterExcludedVulns_MultipleCVEs tests filtering multiple CVEs
func (s *VulncheckTestSuite) TestFilterExcludedVulns_MultipleCVEs() {
	result, err := ParseGovulncheckJSON(sampleGovulncheckJSON)
	s.Require().NoError(err)

	// Exclude both CVEs
	excludes := []string{"CVE-2024-38513", "CVE-2023-45142"}
	filtered := FilterExcludedVulns(result, excludes)

	// Should have excluded two CVEs
	s.Len(filtered.ExcludedCVEs, 2)

	// Should have 1 remaining finding (for GO-2024-9999)
	s.Len(filtered.RemainingFindings, 1)
	s.Equal("GO-2024-9999", filtered.RemainingFindings[0].OSV)
}

// TestFilterExcludedVulns_CaseInsensitive tests case-insensitive matching
func (s *VulncheckTestSuite) TestFilterExcludedVulns_CaseInsensitive() {
	result, err := ParseGovulncheckJSON(sampleGovulncheckJSON)
	s.Require().NoError(err)

	// Test with lowercase CVE
	excludes := []string{"cve-2024-38513"}
	filtered := FilterExcludedVulns(result, excludes)

	s.Len(filtered.ExcludedCVEs, 1)
	s.Len(filtered.RemainingFindings, 2)
}

// TestFilterExcludedVulns_AllExcluded tests excluding all vulnerabilities
func (s *VulncheckTestSuite) TestFilterExcludedVulns_AllExcluded() {
	result, err := ParseGovulncheckJSON(sampleGovulncheckJSON)
	s.Require().NoError(err)

	// Exclude all CVEs
	excludes := []string{"CVE-2024-38513", "CVE-2023-45142", "CVE-2024-99999"}
	filtered := FilterExcludedVulns(result, excludes)

	// All should be excluded
	s.Len(filtered.ExcludedCVEs, 3)
	s.Empty(filtered.RemainingFindings)
}

// TestFilterExcludedVulns_EmptyExcludes tests with no exclusions
func (s *VulncheckTestSuite) TestFilterExcludedVulns_EmptyExcludes() {
	result, err := ParseGovulncheckJSON(sampleGovulncheckJSON)
	s.Require().NoError(err)

	// No exclusions
	excludes := []string{}
	filtered := FilterExcludedVulns(result, excludes)

	// Nothing should be excluded
	s.Empty(filtered.ExcludedCVEs)
	s.Len(filtered.RemainingFindings, 3)
}

// TestFilterExcludedVulns_NilResult tests handling nil input
func (s *VulncheckTestSuite) TestFilterExcludedVulns_NilResult() {
	filtered := FilterExcludedVulns(nil, []string{"CVE-2024-38513"})

	s.Empty(filtered.ExcludedCVEs)
	s.Empty(filtered.RemainingFindings)
}

// TestFilterExcludedVulns_NonMatchingExcludes tests exclusions that don't match
func (s *VulncheckTestSuite) TestFilterExcludedVulns_NonMatchingExcludes() {
	result, err := ParseGovulncheckJSON(sampleGovulncheckJSON)
	s.Require().NoError(err)

	// Exclude CVEs that don't exist
	excludes := []string{"CVE-9999-99999", "CVE-0000-00000"}
	filtered := FilterExcludedVulns(result, excludes)

	// Nothing should be excluded
	s.Empty(filtered.ExcludedCVEs)
	s.Len(filtered.RemainingFindings, 3)
}

// TestParseCVEExclusions_FromEnv tests parsing from environment variable
func (s *VulncheckTestSuite) TestParseCVEExclusions_FromEnv() {
	// Set environment variable using T().Setenv which handles cleanup automatically
	s.T().Setenv("MAGE_X_CVE_EXCLUDES", "CVE-2024-38513,CVE-2023-45142")

	params := make(map[string]string)
	excludes := ParseCVEExclusions(params)

	s.Len(excludes, 2)
	s.Contains(excludes, "CVE-2024-38513")
	s.Contains(excludes, "CVE-2023-45142")
}

// TestParseCVEExclusions_FromParams tests parsing from command parameters
func (s *VulncheckTestSuite) TestParseCVEExclusions_FromParams() {
	// Ensure env var is not set by setting it to empty
	s.T().Setenv("MAGE_X_CVE_EXCLUDES", "")

	params := map[string]string{
		"exclude": "CVE-2024-38513,CVE-2023-45142",
	}
	excludes := ParseCVEExclusions(params)

	s.Len(excludes, 2)
	s.Contains(excludes, "CVE-2024-38513")
	s.Contains(excludes, "CVE-2023-45142")
}

// TestParseCVEExclusions_Combined tests merging env var and params
func (s *VulncheckTestSuite) TestParseCVEExclusions_Combined() {
	// Set environment variable using T().Setenv which handles cleanup automatically
	s.T().Setenv("MAGE_X_CVE_EXCLUDES", "CVE-2024-38513")

	params := map[string]string{
		"exclude": "CVE-2023-45142",
	}
	excludes := ParseCVEExclusions(params)

	s.Len(excludes, 2)
	s.Contains(excludes, "CVE-2024-38513")
	s.Contains(excludes, "CVE-2023-45142")
}

// TestParseCVEExclusions_Deduplication tests that duplicates are removed
func (s *VulncheckTestSuite) TestParseCVEExclusions_Deduplication() {
	// Set environment variable with duplicate using T().Setenv
	s.T().Setenv("MAGE_X_CVE_EXCLUDES", "CVE-2024-38513,CVE-2024-38513")

	params := map[string]string{
		"exclude": "CVE-2024-38513", // Same CVE again
	}
	excludes := ParseCVEExclusions(params)

	// Should be deduplicated to 1
	s.Len(excludes, 1)
	s.Contains(excludes, "CVE-2024-38513")
}

// TestParseCVEExclusions_Whitespace tests handling of whitespace
func (s *VulncheckTestSuite) TestParseCVEExclusions_Whitespace() {
	// Set environment variable with whitespace using T().Setenv
	s.T().Setenv("MAGE_X_CVE_EXCLUDES", "  CVE-2024-38513  ,  CVE-2023-45142  ")

	params := make(map[string]string)
	excludes := ParseCVEExclusions(params)

	s.Len(excludes, 2)
	s.Contains(excludes, "CVE-2024-38513")
	s.Contains(excludes, "CVE-2023-45142")
}

// TestParseCVEExclusions_Empty tests handling of empty values
func (s *VulncheckTestSuite) TestParseCVEExclusions_Empty() {
	s.T().Setenv("MAGE_X_CVE_EXCLUDES", "")

	params := make(map[string]string)
	excludes := ParseCVEExclusions(params)

	s.Empty(excludes)
}

// TestSplitAndTrimCVEs tests the split and trim helper
func (s *VulncheckTestSuite) TestSplitAndTrimCVEs() {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "simple list",
			input:    "CVE-2024-38513,CVE-2023-45142",
			expected: []string{"CVE-2024-38513", "CVE-2023-45142"},
		},
		{
			name:     "with whitespace",
			input:    "  CVE-2024-38513  ,  CVE-2023-45142  ",
			expected: []string{"CVE-2024-38513", "CVE-2023-45142"},
		},
		{
			name:     "single value",
			input:    "CVE-2024-38513",
			expected: []string{"CVE-2024-38513"},
		},
		{
			name:     "empty string",
			input:    "",
			expected: []string{},
		},
		{
			name:     "only commas",
			input:    ",,,,",
			expected: []string{},
		},
		{
			name:     "trailing comma",
			input:    "CVE-2024-38513,",
			expected: []string{"CVE-2024-38513"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := splitAndTrimCVEs(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

// TestDedupStrings tests the deduplication helper
func (s *VulncheckTestSuite) TestDedupStrings() {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "case insensitive",
			input:    []string{"CVE-2024-38513", "cve-2024-38513"},
			expected: []string{"CVE-2024-38513"},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := dedupStrings(tt.input)
			s.Equal(tt.expected, result)
		})
	}
}

// Standalone tests (not using suite)

func TestParseGovulncheckJSON_Basic(t *testing.T) {
	t.Parallel()

	result, err := ParseGovulncheckJSON(sampleGovulncheckJSON)
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.OSVEntries, 3)
	assert.Len(t, result.Findings, 3)
}

func TestFilterExcludedVulns_Basic(t *testing.T) {
	t.Parallel()

	result, err := ParseGovulncheckJSON(sampleGovulncheckJSON)
	require.NoError(t, err)

	filtered := FilterExcludedVulns(result, []string{"CVE-2024-38513", "CVE-2023-45142"})
	assert.Len(t, filtered.ExcludedCVEs, 2)
	assert.Len(t, filtered.RemainingFindings, 1)
}

func TestParseCVEExclusions_EnvVar(t *testing.T) {
	// Don't run in parallel due to env var manipulation
	t.Setenv("MAGE_X_CVE_EXCLUDES", "CVE-2024-38513,CVE-2023-45142")

	params := make(map[string]string)
	excludes := ParseCVEExclusions(params)

	assert.Len(t, excludes, 2)
	assert.Contains(t, excludes, "CVE-2024-38513")
	assert.Contains(t, excludes, "CVE-2023-45142")
}

// TestDisplayScanConfig tests the scan config display function
func (s *VulncheckTestSuite) TestDisplayScanConfig() {
	tests := []struct {
		name   string
		config *VulnConfig
	}{
		{
			name:   "nil config",
			config: nil,
		},
		{
			name: "full config",
			config: &VulnConfig{
				ProtocolVersion: "v1.0.0",
				ScannerName:     "govulncheck",
				ScannerVersion:  "v1.1.4",
				DB:              "https://vuln.go.dev",
				DBLastModified:  "2025-12-20T15:04:05Z",
				GoVersion:       "go1.23.4",
				ScanLevel:       "symbol",
			},
		},
		{
			name: "minimal config",
			config: &VulnConfig{
				ScannerVersion: "v1.1.4",
			},
		},
		{
			name: "config with only scanner name",
			config: &VulnConfig{
				ScannerName:    "govulncheck",
				ScannerVersion: "v1.1.4",
			},
		},
		{
			name:   "empty config",
			config: &VulnConfig{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Should not panic
			s.NotPanics(func() {
				DisplayScanConfig(tt.config)
			})
		})
	}
}

// TestDisplayScanConfig_DateTruncation tests that long dates are truncated
func (s *VulncheckTestSuite) TestDisplayScanConfig_DateTruncation() {
	config := &VulnConfig{
		ScannerName:    "govulncheck",
		ScannerVersion: "v1.1.4",
		DBLastModified: "2025-12-20T15:04:05Z", // Full timestamp
	}

	// Should not panic when truncating date
	s.NotPanics(func() {
		DisplayScanConfig(config)
	})
}

// TestDisplayScannedModules tests the module display function
func (s *VulncheckTestSuite) TestDisplayScannedModules() {
	tests := []struct {
		name string
		deps []ModuleDep
	}{
		{
			name: "empty deps",
			deps: []ModuleDep{},
		},
		{
			name: "nil deps",
			deps: nil,
		},
		{
			name: "single dep with version",
			deps: []ModuleDep{
				{Path: "github.com/example/dep", Version: "v1.0.0"},
			},
		},
		{
			name: "single dep without version",
			deps: []ModuleDep{
				{Path: "github.com/example/dep"},
			},
		},
		{
			name: "multiple deps",
			deps: []ModuleDep{
				{Path: "github.com/example/dep1", Version: "v1.0.0"},
				{Path: "github.com/example/dep2", Version: "v2.0.0"},
				{Path: "github.com/example/dep3", Version: "v3.0.0"},
			},
		},
		{
			name: "mixed deps with and without versions",
			deps: []ModuleDep{
				{Path: "github.com/example/main"},
				{Path: "github.com/example/dep1", Version: "v1.0.0"},
				{Path: "github.com/example/dep2", Version: "v2.0.0"},
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			// Should not panic
			s.NotPanics(func() {
				DisplayScannedModules(tt.deps)
			})
		})
	}
}

// TestDisplayScanConfigFromParsedJSON tests displaying config parsed from JSON
func (s *VulncheckTestSuite) TestDisplayScanConfigFromParsedJSON() {
	result, err := ParseGovulncheckJSON(sampleGovulncheckJSON)
	s.Require().NoError(err)
	s.Require().NotNil(result.Config)

	// Should not panic when displaying parsed config
	s.NotPanics(func() {
		DisplayScanConfig(result.Config)
	})

	// Verify config fields were parsed correctly
	s.Equal("govulncheck", result.Config.ScannerName)
	s.Equal("v1.1.4", result.Config.ScannerVersion)
	s.Equal("go1.21.0", result.Config.GoVersion)
	s.Equal("symbol", result.Config.ScanLevel)
}
