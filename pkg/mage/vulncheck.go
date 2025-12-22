package mage

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mrz1836/mage-x/pkg/utils"
)

// Govulncheck JSON output types
// Based on: https://github.com/golang/vuln/blob/master/internal/govulncheck/govulncheck.go

// VulnMessage represents a single message in govulncheck's streaming JSON output.
// Exactly one field will be populated per message.
type VulnMessage struct {
	Config   *VulnConfig   `json:"config,omitempty"`
	Progress *VulnProgress `json:"progress,omitempty"`
	OSV      *VulnOSV      `json:"osv,omitempty"`
	Finding  *VulnFinding  `json:"finding,omitempty"`
}

// VulnConfig holds configuration information about the govulncheck scan.
type VulnConfig struct {
	ProtocolVersion string `json:"protocol_version,omitempty"`
	ScannerName     string `json:"scanner_name,omitempty"`
	ScannerVersion  string `json:"scanner_version,omitempty"`
	DB              string `json:"db,omitempty"`
	DBLastModified  string `json:"db_last_modified,omitempty"`
	GoVersion       string `json:"go_version,omitempty"`
	ScanLevel       string `json:"scan_level,omitempty"`
	ScanMode        string `json:"scan_mode,omitempty"`
}

// VulnProgress represents a progress message during scanning.
type VulnProgress struct {
	Message string `json:"message,omitempty"`
}

// VulnOSV represents an OSV (Open Source Vulnerability) entry.
type VulnOSV struct {
	ID       string   `json:"id,omitempty"`      // GO-YYYY-NNNN format
	Aliases  []string `json:"aliases,omitempty"` // CVE-YYYY-NNNN entries
	Summary  string   `json:"summary,omitempty"`
	Details  string   `json:"details,omitempty"`
	Modified string   `json:"modified,omitempty"`
}

// VulnFinding represents a finding for a vulnerability.
type VulnFinding struct {
	OSV          string      `json:"osv,omitempty"` // References OSV ID
	FixedVersion string      `json:"fixed_version,omitempty"`
	Trace        []VulnFrame `json:"trace,omitempty"`
}

// VulnFrame represents a single frame in a vulnerability trace.
type VulnFrame struct {
	Module   string        `json:"module,omitempty"`
	Version  string        `json:"version,omitempty"`
	Package  string        `json:"package,omitempty"`
	Function string        `json:"function,omitempty"`
	Receiver string        `json:"receiver,omitempty"`
	Position *VulnPosition `json:"position,omitempty"`
}

// VulnPosition represents a source code position.
type VulnPosition struct {
	Filename string `json:"filename,omitempty"`
	Offset   int    `json:"offset,omitempty"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
}

// VulnScanResult holds the aggregated results of a vulnerability scan.
type VulnScanResult struct {
	Config       *VulnConfig
	OSVEntries   map[string]*VulnOSV // keyed by OSV ID
	Findings     []*VulnFinding
	ProgressMsgs []string
}

// VulnFilterResult holds the results after filtering excluded CVEs.
type VulnFilterResult struct {
	// RemainingFindings are findings for vulnerabilities NOT in the exclusion list
	RemainingFindings []*VulnFinding
	// ExcludedCVEs are the CVE IDs that were excluded
	ExcludedCVEs []string
	// AllOSVEntries is the original map of OSV entries for reference
	AllOSVEntries map[string]*VulnOSV
}

// ParseGovulncheckJSON parses govulncheck's JSON output.
// It handles both NDJSON (newline-delimited) and pretty-printed multi-line JSON formats.
func ParseGovulncheckJSON(jsonOutput string) (*VulnScanResult, error) {
	result := &VulnScanResult{
		OSVEntries: make(map[string]*VulnOSV),
		Findings:   make([]*VulnFinding, 0),
	}

	// Use json.Decoder to handle multiple JSON documents in the stream
	decoder := json.NewDecoder(strings.NewReader(jsonOutput))

	for {
		var msg VulnMessage
		if err := decoder.Decode(&msg); err != nil {
			if err.Error() == "EOF" {
				break
			}
			// Try to continue parsing even if one document fails
			continue
		}

		switch {
		case msg.Config != nil:
			result.Config = msg.Config
		case msg.Progress != nil:
			result.ProgressMsgs = append(result.ProgressMsgs, msg.Progress.Message)
		case msg.OSV != nil:
			result.OSVEntries[msg.OSV.ID] = msg.OSV
		case msg.Finding != nil:
			result.Findings = append(result.Findings, msg.Finding)
		}
	}

	return result, nil
}

// FilterExcludedVulns filters out vulnerabilities whose CVE aliases match the exclusion list.
// Returns the filtered result and the list of CVEs that were actually excluded.
func FilterExcludedVulns(result *VulnScanResult, excludes []string) *VulnFilterResult {
	if result == nil {
		return &VulnFilterResult{}
	}

	// Normalize exclusions to uppercase for case-insensitive matching
	excludeSet := make(map[string]bool)
	for _, cve := range excludes {
		excludeSet[strings.ToUpper(strings.TrimSpace(cve))] = true
	}

	// Build a set of OSV IDs that should be excluded based on CVE aliases
	excludedOSVIDs := make(map[string]bool)
	excludedCVEs := make([]string, 0)

	for osvID, osv := range result.OSVEntries {
		for _, alias := range osv.Aliases {
			normalizedAlias := strings.ToUpper(strings.TrimSpace(alias))
			if excludeSet[normalizedAlias] {
				excludedOSVIDs[osvID] = true
				excludedCVEs = append(excludedCVEs, alias)
				break
			}
		}
	}

	// Filter findings to only include non-excluded vulnerabilities
	remaining := make([]*VulnFinding, 0)
	for _, finding := range result.Findings {
		if !excludedOSVIDs[finding.OSV] {
			remaining = append(remaining, finding)
		}
	}

	return &VulnFilterResult{
		RemainingFindings: remaining,
		ExcludedCVEs:      dedupStrings(excludedCVEs),
		AllOSVEntries:     result.OSVEntries,
	}
}

// ParseCVEExclusions parses CVE exclusions from environment variable and params.
// Environment variable: MAGE_X_CVE_EXCLUDES (comma-separated)
// Param: exclude=CVE-X,CVE-Y
func ParseCVEExclusions(params map[string]string) []string {
	var excludes []string

	// From environment variable
	if envExcludes := os.Getenv("MAGE_X_CVE_EXCLUDES"); envExcludes != "" {
		excludes = append(excludes, splitAndTrimCVEs(envExcludes)...)
	}

	// From params (merged with env var)
	if paramExcludes, ok := params["exclude"]; ok && paramExcludes != "" {
		excludes = append(excludes, splitAndTrimCVEs(paramExcludes)...)
	}

	return dedupStrings(excludes)
}

// ReportVulnerabilities outputs found vulnerabilities in a readable format.
func ReportVulnerabilities(result *VulnFilterResult, allOSV map[string]*VulnOSV) {
	if result == nil || len(result.RemainingFindings) == 0 {
		return
	}

	// Group findings by OSV ID
	findingsByOSV := make(map[string][]*VulnFinding)
	for _, f := range result.RemainingFindings {
		findingsByOSV[f.OSV] = append(findingsByOSV[f.OSV], f)
	}

	// Report each unique vulnerability
	for osvID, findings := range findingsByOSV {
		osv := allOSV[osvID]
		if osv == nil {
			utils.Error("  %s: Unknown vulnerability", osvID)
			continue
		}

		// Get CVE alias if available
		cveID := osvID
		for _, alias := range osv.Aliases {
			if strings.HasPrefix(strings.ToUpper(alias), "CVE-") {
				cveID = alias
				break
			}
		}

		// Report vulnerability with summary
		summary := osv.Summary
		if summary == "" {
			summary = "No description available"
		}
		utils.Error("  %s: %s", cveID, summary)

		// Show fix version if available
		if len(findings) > 0 && findings[0].FixedVersion != "" {
			utils.Info("    Fixed in: %s", findings[0].FixedVersion)
		}

		// Show affected module from first trace
		if len(findings) > 0 && len(findings[0].Trace) > 0 {
			frame := findings[0].Trace[0]
			if frame.Module != "" {
				utils.Info("    Module: %s@%s", frame.Module, frame.Version)
			}
		}
	}
}

// splitAndTrimCVEs splits a comma-separated string and trims whitespace.
func splitAndTrimCVEs(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// dedupStrings removes duplicates from a string slice while preserving order.
func dedupStrings(strs []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(strs))
	for _, s := range strs {
		upper := strings.ToUpper(s)
		if !seen[upper] {
			seen[upper] = true
			result = append(result, s)
		}
	}
	return result
}

// DisplayScanConfig outputs the govulncheck scan configuration in a concise format.
func DisplayScanConfig(config *VulnConfig) {
	if config == nil {
		return
	}

	// Format: "govulncheck v1.1.4 | DB: 2025-12-20 | Go: 1.23.4 | Level: symbol"
	parts := []string{}

	if config.ScannerName != "" && config.ScannerVersion != "" {
		parts = append(parts, fmt.Sprintf("%s %s", config.ScannerName, config.ScannerVersion))
	} else if config.ScannerVersion != "" {
		parts = append(parts, fmt.Sprintf("govulncheck %s", config.ScannerVersion))
	}

	if config.DBLastModified != "" {
		// Try to format the date nicely (truncate time portion if present)
		dbDate := config.DBLastModified
		if len(dbDate) >= 10 {
			dbDate = dbDate[:10] // Take just YYYY-MM-DD
		}
		parts = append(parts, fmt.Sprintf("DB: %s", dbDate))
	}

	if config.GoVersion != "" {
		parts = append(parts, fmt.Sprintf("Go: %s", config.GoVersion))
	}

	if config.ScanLevel != "" {
		parts = append(parts, fmt.Sprintf("Level: %s", config.ScanLevel))
	}

	if len(parts) > 0 {
		utils.Info("%s", strings.Join(parts, " | "))
	}
}

// DisplayScannedModules outputs the list of scanned module dependencies with a summary.
func DisplayScannedModules(deps []ModuleDep) {
	if len(deps) == 0 {
		return
	}

	utils.Info("Scanned dependencies:")
	for _, dep := range deps {
		if dep.Version != "" {
			fmt.Printf("  %s %s\n", dep.Path, dep.Version)
		} else {
			fmt.Printf("  %s\n", dep.Path)
		}
	}
	utils.Info("Summary: %d modules scanned", len(deps))
}
