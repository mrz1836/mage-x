// Package mage provides reusable build tasks for Go projects using Mage
package mage

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/mrz1836/go-mage/pkg/common/fileops"
	"github.com/mrz1836/go-mage/pkg/utils"
)

// Audit namespace for audit logging and compliance operations
type Audit mg.Namespace

// Show displays audit events with optional filtering
func (Audit) Show() error {
	utils.Header("📊 MAGE-X Audit Log")

	auditLogger := utils.GetAuditLogger()

	// Parse command line arguments for filtering
	filter := utils.AuditFilter{
		Limit: 50, // Default limit
	}

	// Check for date range filter
	if startTime := utils.GetEnv("START_TIME", ""); startTime != "" {
		if t, err := time.Parse("2006-01-02", startTime); err == nil {
			filter.StartTime = t
		}
	}

	if endTime := utils.GetEnv("END_TIME", ""); endTime != "" {
		if t, err := time.Parse("2006-01-02", endTime); err == nil {
			filter.EndTime = t.Add(24 * time.Hour) // End of day
		}
	}

	// Check for user filter
	if userFilter := utils.GetEnv("USER", ""); userFilter != "" {
		filter.User = userFilter
	}

	// Check for command filter
	if commandFilter := utils.GetEnv("COMMAND", ""); commandFilter != "" {
		filter.Command = commandFilter
	}

	// Check for success filter
	if successFilter := utils.GetEnv("SUCCESS", ""); successFilter != "" {
		if success, err := strconv.ParseBool(successFilter); err == nil {
			filter.Success = &success
		}
	}

	// Check for limit
	if limitStr := utils.GetEnv("LIMIT", ""); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = limit
		}
	}

	events, err := auditLogger.GetEvents(filter)
	if err != nil {
		return fmt.Errorf("failed to get audit events: %w", err)
	}

	if len(events) == 0 {
		utils.Info("No audit events found")
		return nil
	}

	utils.Info("Found %d audit events", len(events))
	fmt.Println()

	// Display events in a table format
	fmt.Printf("%-20s %-12s %-20s %-15s %-8s %-10s\n", "TIMESTAMP", "USER", "COMMAND", "DURATION", "SUCCESS", "EXIT CODE")
	fmt.Println(strings.Repeat("-", 95))

	for _, event := range events {
		status := "✅"
		if !event.Success {
			status = "❌"
		}

		fmt.Printf("%-20s %-12s %-20s %-15s %-8s %-10d\n",
			event.Timestamp.Format("2006-01-02 15:04:05"),
			event.User,
			truncateString(event.Command, 20),
			event.Duration.Round(time.Millisecond),
			status,
			event.ExitCode,
		)
	}

	return nil
}

// Stats displays audit statistics
func (Audit) Stats() error {
	utils.Header("📈 MAGE-X Audit Statistics")

	auditLogger := utils.GetAuditLogger()

	stats, err := auditLogger.GetStats()
	if err != nil {
		return fmt.Errorf("failed to get audit stats: %w", err)
	}

	if stats.TotalEvents == 0 {
		utils.Info("No audit events found")
		return nil
	}

	utils.Info("Audit Statistics:")
	fmt.Printf("  Total Events: %d\n", stats.TotalEvents)
	fmt.Printf("  Successful:   %d (%.1f%%)\n", stats.SuccessfulEvents, float64(stats.SuccessfulEvents)/float64(stats.TotalEvents)*100)
	fmt.Printf("  Failed:       %d (%.1f%%)\n", stats.FailedEvents, float64(stats.FailedEvents)/float64(stats.TotalEvents)*100)
	fmt.Printf("  Date Range:   %s to %s\n", stats.EarliestEvent.Format("2006-01-02"), stats.LatestEvent.Format("2006-01-02"))

	if len(stats.TopUsers) > 0 {
		fmt.Println("\nTop Users:")
		for i, user := range stats.TopUsers {
			if i >= 5 { // Show top 5
				break
			}
			fmt.Printf("  %d. %s (%d events)\n", i+1, user.User, user.Count)
		}
	}

	if len(stats.TopCommands) > 0 {
		fmt.Println("\nTop Commands:")
		for i, cmd := range stats.TopCommands {
			if i >= 5 { // Show top 5
				break
			}
			fmt.Printf("  %d. %s (%d times)\n", i+1, cmd.Command, cmd.Count)
		}
	}

	return nil
}

// Export exports audit events to JSON format
func (Audit) Export() error {
	utils.Header("📤 Exporting Audit Events")

	auditLogger := utils.GetAuditLogger()

	// Parse filtering options
	filter := utils.AuditFilter{}

	if startTime := utils.GetEnv("START_TIME", ""); startTime != "" {
		if t, err := time.Parse("2006-01-02", startTime); err == nil {
			filter.StartTime = t
		}
	}

	if endTime := utils.GetEnv("END_TIME", ""); endTime != "" {
		if t, err := time.Parse("2006-01-02", endTime); err == nil {
			filter.EndTime = t.Add(24 * time.Hour)
		}
	}

	if userFilter := utils.GetEnv("USER", ""); userFilter != "" {
		filter.User = userFilter
	}

	if commandFilter := utils.GetEnv("COMMAND", ""); commandFilter != "" {
		filter.Command = commandFilter
	}

	// Export events
	data, err := auditLogger.ExportEvents(filter)
	if err != nil {
		return fmt.Errorf("failed to export audit events: %w", err)
	}

	// Determine output file
	outputFile := utils.GetEnv("OUTPUT", "audit-export.json")

	// Write to file
	fileOps := fileops.New()
	if err := fileOps.File.WriteFile(outputFile, data, 0o644); err != nil {
		return fmt.Errorf("failed to write export file: %w", err)
	}

	utils.Success("Audit events exported to: %s", outputFile)
	return nil
}

// Cleanup removes old audit events based on retention policy
func (Audit) Cleanup() error {
	utils.Header("🧹 Cleaning Up Old Audit Events")

	auditLogger := utils.GetAuditLogger()

	if err := auditLogger.CleanupOldEvents(); err != nil {
		return fmt.Errorf("failed to cleanup old audit events: %w", err)
	}

	utils.Success("Old audit events cleaned up successfully")
	return nil
}

// Enable enables audit logging for the current project
func (Audit) Enable() error {
	utils.Header("🔍 Enabling Audit Logging")

	// Create .mage directory if it doesn't exist
	fileOps := fileops.New()
	if err := fileOps.File.MkdirAll(".mage", 0o755); err != nil {
		return fmt.Errorf("failed to create .mage directory: %w", err)
	}

	// Create or update .mage/config.yaml with audit settings
	configPath := ".mage/config.yaml"

	// Read existing config if it exists
	var config struct {
		Audit utils.AuditConfig `yaml:"audit"`
	}

	if fileOps.File.Exists(configPath) {
		// Config exists, read it
		if err := fileOps.JSON.ReadJSON(configPath, &config); err != nil {
			// If read fails, use defaults
			config.Audit = utils.DefaultAuditConfig()
		}
	} else {
		// Config doesn't exist, use defaults
		config.Audit = utils.DefaultAuditConfig()
	}

	// Enable audit logging
	config.Audit.Enabled = true

	// Write config back
	if err := fileOps.JSON.WriteJSONIndent(configPath, config, "", "  "); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	utils.Success("Audit logging enabled!")
	utils.Info("Configuration saved to: %s", configPath)
	utils.Info("Set MAGE_AUDIT_ENABLED=true environment variable to enable globally")

	return nil
}

// Disable disables audit logging for the current project
func (Audit) Disable() error {
	utils.Header("🚫 Disabling Audit Logging")

	configPath := ".mage/config.yaml"
	fileOps := fileops.New()

	// Read existing config if it exists
	var config struct {
		Audit utils.AuditConfig `yaml:"audit"`
	}

	if fileOps.File.Exists(configPath) {
		// Config exists, read it
		if err := fileOps.JSON.ReadJSON(configPath, &config); err != nil {
			// If read fails, use defaults
			config.Audit = utils.DefaultAuditConfig()
		}
	} else {
		// Config doesn't exist, use defaults
		config.Audit = utils.DefaultAuditConfig()
	}

	// Disable audit logging
	config.Audit.Enabled = false

	// Write config back
	if err := fileOps.JSON.WriteJSONIndent(configPath, config, "", "  "); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	utils.Success("Audit logging disabled!")
	utils.Info("Configuration saved to: %s", configPath)

	return nil
}

// Report generates a compliance report
func (Audit) Report() error {
	utils.Header("📋 Generating Compliance Report")

	auditLogger := utils.GetAuditLogger()

	// Get audit statistics
	stats, err := auditLogger.GetStats()
	if err != nil {
		return fmt.Errorf("failed to get audit stats: %w", err)
	}

	// Generate compliance report
	report := AuditComplianceReport{
		GeneratedAt:      time.Now(),
		ReportPeriod:     fmt.Sprintf("%s to %s", stats.EarliestEvent.Format("2006-01-02"), stats.LatestEvent.Format("2006-01-02")),
		TotalEvents:      stats.TotalEvents,
		SuccessfulEvents: stats.SuccessfulEvents,
		FailedEvents:     stats.FailedEvents,
		SuccessRate:      float64(stats.SuccessfulEvents) / float64(stats.TotalEvents) * 100,
		TopUsers:         stats.TopUsers,
		TopCommands:      stats.TopCommands,
		AuditConfig:      utils.DefaultAuditConfig(),
	}

	// Get recent failed events for analysis
	failedFilter := utils.AuditFilter{
		Success: &[]bool{false}[0],
		Limit:   10,
	}

	failedEvents, err := auditLogger.GetEvents(failedFilter)
	if err == nil {
		report.RecentFailures = failedEvents
	}

	// Output file
	outputFile := utils.GetEnv("OUTPUT", "compliance-report.json")

	// Write report
	fileOps := fileops.New()
	if err := fileOps.JSON.WriteJSONIndent(outputFile, report, "", "  "); err != nil {
		return fmt.Errorf("failed to write report file: %w", err)
	}

	utils.Success("Compliance report generated: %s", outputFile)

	// Display summary
	fmt.Printf("\nCompliance Report Summary:\n")
	fmt.Printf("  Report Period: %s\n", report.ReportPeriod)
	fmt.Printf("  Total Events: %d\n", report.TotalEvents)
	fmt.Printf("  Success Rate: %.1f%%\n", report.SuccessRate)
	fmt.Printf("  Recent Failures: %d\n", len(report.RecentFailures))

	return nil
}

// AuditComplianceReport represents a compliance audit report
type AuditComplianceReport struct {
	GeneratedAt      time.Time            `json:"generated_at"`
	ReportPeriod     string               `json:"report_period"`
	TotalEvents      int                  `json:"total_events"`
	SuccessfulEvents int                  `json:"successful_events"`
	FailedEvents     int                  `json:"failed_events"`
	SuccessRate      float64              `json:"success_rate"`
	TopUsers         []utils.UserStats    `json:"top_users"`
	TopCommands      []utils.CommandStats `json:"top_commands"`
	RecentFailures   []utils.AuditEvent   `json:"recent_failures"`
	AuditConfig      utils.AuditConfig    `json:"audit_config"`
}

// Helper functions

// LogCommandExecution logs a command execution for audit purposes
func LogCommandExecution(command string, args []string, startTime time.Time, duration time.Duration, exitCode int, success bool) {
	auditLogger := utils.GetAuditLogger()

	// Get current user
	currentUser := "unknown"
	if usr, err := user.Current(); err == nil {
		currentUser = usr.Username
	}

	// Get current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		workingDir = "."
	}

	// Create audit event
	event := utils.AuditEvent{
		Timestamp:   startTime,
		User:        currentUser,
		Command:     command,
		Args:        args,
		WorkingDir:  workingDir,
		Duration:    duration,
		ExitCode:    exitCode,
		Success:     success,
		Environment: getFilteredEnvironment(),
		Metadata: map[string]string{
			"mage_version": getVersion(),
			"go_version":   getGoVersion(),
			"project":      filepath.Base(workingDir),
		},
	}

	// Log the event (errors are handled internally by the logger)
	if err := auditLogger.LogEvent(event); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: Failed to log audit event: %v\n", err)
	}
}

// getFilteredEnvironment returns a filtered environment map
func getFilteredEnvironment() map[string]string {
	env := make(map[string]string)

	// Only include relevant environment variables
	relevantEnvs := []string{
		"GO_VERSION",
		"GOPATH",
		"GOROOT",
		"GOOS",
		"GOARCH",
		"CGO_ENABLED",
		"MAGE_VERBOSE",
		"MAGE_AUDIT_ENABLED",
	}

	for _, envVar := range relevantEnvs {
		if value := os.Getenv(envVar); value != "" {
			env[envVar] = value
		}
	}

	return env
}

// getGoVersion returns the Go version
func getGoVersion() string {
	output, err := GetRunner().RunCmdOutput("go", "version")
	if err != nil {
		return "unknown"
	}

	// Parse "go version go1.24.0 linux/amd64" -> "1.24.0"
	parts := strings.Fields(output)
	if len(parts) >= 3 {
		version := parts[2]
		if strings.HasPrefix(version, "go") {
			return version[2:]
		}
	}

	return "unknown"
}
