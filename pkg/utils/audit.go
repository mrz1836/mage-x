// Package utils provides utility functions for audit logging
package utils

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// AuditEvent represents a single audit event
type AuditEvent struct {
	ID          int64             `json:"id"`
	Timestamp   time.Time         `json:"timestamp"`
	User        string            `json:"user"`
	Command     string            `json:"command"`
	Args        []string          `json:"args"`
	WorkingDir  string            `json:"working_dir"`
	Duration    time.Duration     `json:"duration"`
	ExitCode    int               `json:"exit_code"`
	Success     bool              `json:"success"`
	Environment map[string]string `json:"environment,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// AuditLogger handles audit logging operations
type AuditLogger struct {
	db      *sql.DB
	mu      sync.RWMutex
	enabled bool
	config  AuditConfig
}

// AuditConfig defines configuration for audit logging
type AuditConfig struct {
	Enabled         bool     `yaml:"enabled"`
	DatabasePath    string   `yaml:"database_path"`
	RetentionDays   int      `yaml:"retention_days"`
	LogEnvironment  bool     `yaml:"log_environment"`
	SensitiveEnvs   []string `yaml:"sensitive_envs"`
	ExcludeCommands []string `yaml:"exclude_commands"`
}

// DefaultAuditConfig returns sensible defaults for audit configuration
func DefaultAuditConfig() AuditConfig {
	return AuditConfig{
		Enabled:        false, // Opt-in for compliance
		DatabasePath:   ".mage/audit.db",
		RetentionDays:  90,
		LogEnvironment: false,
		SensitiveEnvs: []string{
			"AWS_SECRET_ACCESS_KEY",
			"GITHUB_TOKEN",
			"DATABASE_PASSWORD",
			"API_KEY",
			"SECRET",
			"PRIVATE_KEY",
			"TOKEN",
			"PASSWORD",
		},
		ExcludeCommands: []string{
			"help",
			"version",
			"audit:show",
		},
	}
}

var (
	// globalAuditLogger is the singleton audit logger instance
	globalAuditLogger *AuditLogger
	auditOnce         sync.Once
)

// GetAuditLogger returns the global audit logger instance
func GetAuditLogger() *AuditLogger {
	auditOnce.Do(func() {
		config := DefaultAuditConfig()

		// Check if audit is enabled via environment variable
		if os.Getenv("MAGE_AUDIT_ENABLED") == "true" {
			config.Enabled = true
		}

		// Override database path if specified
		if dbPath := os.Getenv("MAGE_AUDIT_DB"); dbPath != "" {
			config.DatabasePath = dbPath
		}

		globalAuditLogger = NewAuditLogger(config)
	})

	return globalAuditLogger
}

// NewAuditLogger creates a new audit logger with the given configuration
func NewAuditLogger(config AuditConfig) *AuditLogger {
	logger := &AuditLogger{
		enabled: config.Enabled,
		config:  config,
	}

	if config.Enabled {
		if err := logger.initDatabase(); err != nil {
			// Log error but don't fail - audit is optional
			Error("Failed to initialize audit database: %v", err)
			logger.enabled = false
		}
	}

	return logger
}

// initDatabase initializes the SQLite database for audit logging
func (a *AuditLogger) initDatabase() error {
	// Ensure directory exists
	dbDir := filepath.Dir(a.config.DatabasePath)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		return fmt.Errorf("failed to create audit database directory: %w", err)
	}

	// Open database
	db, err := sql.Open("sqlite3", a.config.DatabasePath)
	if err != nil {
		return fmt.Errorf("failed to open audit database: %w", err)
	}

	// Create table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS audit_events (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		user TEXT NOT NULL,
		command TEXT NOT NULL,
		args TEXT NOT NULL,
		working_dir TEXT NOT NULL,
		duration INTEGER NOT NULL,
		exit_code INTEGER NOT NULL,
		success BOOLEAN NOT NULL,
		environment TEXT,
		metadata TEXT
	);
	
	CREATE INDEX IF NOT EXISTS idx_timestamp ON audit_events(timestamp);
	CREATE INDEX IF NOT EXISTS idx_user ON audit_events(user);
	CREATE INDEX IF NOT EXISTS idx_command ON audit_events(command);
	CREATE INDEX IF NOT EXISTS idx_success ON audit_events(success);
	`

	if _, err := db.Exec(createTableSQL); err != nil {
		db.Close()
		return fmt.Errorf("failed to create audit table: %w", err)
	}

	a.db = db
	return nil
}

// LogEvent logs an audit event
func (a *AuditLogger) LogEvent(event AuditEvent) error {
	if !a.enabled {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Check if command should be excluded
	for _, excluded := range a.config.ExcludeCommands {
		if event.Command == excluded {
			return nil
		}
	}

	// Filter sensitive environment variables
	if event.Environment != nil {
		event.Environment = a.filterSensitiveEnvs(event.Environment)
	}

	// Serialize JSON fields
	argsJSON, _ := json.Marshal(event.Args)
	envJSON, _ := json.Marshal(event.Environment)
	metadataJSON, _ := json.Marshal(event.Metadata)

	// Insert into database
	query := `
	INSERT INTO audit_events (
		timestamp, user, command, args, working_dir, 
		duration, exit_code, success, environment, metadata
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := a.db.Exec(query,
		event.Timestamp,
		event.User,
		event.Command,
		string(argsJSON),
		event.WorkingDir,
		event.Duration.Nanoseconds(),
		event.ExitCode,
		event.Success,
		string(envJSON),
		string(metadataJSON),
	)

	return err
}

// GetEvents retrieves audit events with optional filtering
func (a *AuditLogger) GetEvents(filter AuditFilter) ([]AuditEvent, error) {
	if !a.enabled {
		return nil, fmt.Errorf("audit logging is disabled")
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	query := `
	SELECT id, timestamp, user, command, args, working_dir, 
		   duration, exit_code, success, environment, metadata
	FROM audit_events
	WHERE 1=1`

	args := []interface{}{}

	// Add filters
	if !filter.StartTime.IsZero() {
		query += " AND timestamp >= ?"
		args = append(args, filter.StartTime)
	}

	if !filter.EndTime.IsZero() {
		query += " AND timestamp <= ?"
		args = append(args, filter.EndTime)
	}

	if filter.User != "" {
		query += " AND user = ?"
		args = append(args, filter.User)
	}

	if filter.Command != "" {
		query += " AND command = ?"
		args = append(args, filter.Command)
	}

	if filter.Success != nil {
		query += " AND success = ?"
		args = append(args, *filter.Success)
	}

	// Add ordering and limit
	query += " ORDER BY timestamp DESC"
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := a.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events: %w", err)
	}
	defer rows.Close()

	var events []AuditEvent
	for rows.Next() {
		var event AuditEvent
		var argsJSON, envJSON, metadataJSON string
		var durationNs int64

		err := rows.Scan(
			&event.ID,
			&event.Timestamp,
			&event.User,
			&event.Command,
			&argsJSON,
			&event.WorkingDir,
			&durationNs,
			&event.ExitCode,
			&event.Success,
			&envJSON,
			&metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit event: %w", err)
		}

		// Deserialize JSON fields
		event.Duration = time.Duration(durationNs)
		json.Unmarshal([]byte(argsJSON), &event.Args)
		json.Unmarshal([]byte(envJSON), &event.Environment)
		json.Unmarshal([]byte(metadataJSON), &event.Metadata)

		events = append(events, event)
	}

	return events, nil
}

// GetStats returns audit statistics
func (a *AuditLogger) GetStats() (AuditStats, error) {
	if !a.enabled {
		return AuditStats{}, fmt.Errorf("audit logging is disabled")
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	var stats AuditStats

	// Total events
	err := a.db.QueryRow("SELECT COUNT(*) FROM audit_events").Scan(&stats.TotalEvents)
	if err != nil {
		return stats, fmt.Errorf("failed to get total events: %w", err)
	}

	// Success/failure counts
	err = a.db.QueryRow("SELECT COUNT(*) FROM audit_events WHERE success = 1").Scan(&stats.SuccessfulEvents)
	if err != nil {
		return stats, fmt.Errorf("failed to get successful events: %w", err)
	}

	err = a.db.QueryRow("SELECT COUNT(*) FROM audit_events WHERE success = 0").Scan(&stats.FailedEvents)
	if err != nil {
		return stats, fmt.Errorf("failed to get failed events: %w", err)
	}

	// Date range
	err = a.db.QueryRow("SELECT MIN(timestamp), MAX(timestamp) FROM audit_events").Scan(&stats.EarliestEvent, &stats.LatestEvent)
	if err != nil {
		return stats, fmt.Errorf("failed to get date range: %w", err)
	}

	// Top users
	rows, err := a.db.Query(`
		SELECT user, COUNT(*) as count 
		FROM audit_events 
		GROUP BY user 
		ORDER BY count DESC 
		LIMIT 10
	`)
	if err != nil {
		return stats, fmt.Errorf("failed to get top users: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var user string
		var count int
		if err := rows.Scan(&user, &count); err != nil {
			continue
		}
		stats.TopUsers = append(stats.TopUsers, UserStats{User: user, Count: count})
	}

	// Top commands
	rows, err = a.db.Query(`
		SELECT command, COUNT(*) as count 
		FROM audit_events 
		GROUP BY command 
		ORDER BY count DESC 
		LIMIT 10
	`)
	if err != nil {
		return stats, fmt.Errorf("failed to get top commands: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var command string
		var count int
		if err := rows.Scan(&command, &count); err != nil {
			continue
		}
		stats.TopCommands = append(stats.TopCommands, CommandStats{Command: command, Count: count})
	}

	return stats, nil
}

// CleanupOldEvents removes old audit events based on retention policy
func (a *AuditLogger) CleanupOldEvents() error {
	if !a.enabled {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	cutoffTime := time.Now().AddDate(0, 0, -a.config.RetentionDays)

	result, err := a.db.Exec("DELETE FROM audit_events WHERE timestamp < ?", cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to cleanup old events: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		Info("Cleaned up %d old audit events", rowsAffected)
	}

	return nil
}

// ExportEvents exports audit events to JSON
func (a *AuditLogger) ExportEvents(filter AuditFilter) ([]byte, error) {
	events, err := a.GetEvents(filter)
	if err != nil {
		return nil, err
	}

	return json.MarshalIndent(events, "", "  ")
}

// Close closes the audit logger database connection
func (a *AuditLogger) Close() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.db != nil {
		return a.db.Close()
	}

	return nil
}

// filterSensitiveEnvs removes sensitive environment variables from the map
func (a *AuditLogger) filterSensitiveEnvs(env map[string]string) map[string]string {
	if !a.config.LogEnvironment {
		return nil
	}

	filtered := make(map[string]string)
	for key, value := range env {
		sensitive := false
		for _, sensitiveEnv := range a.config.SensitiveEnvs {
			if key == sensitiveEnv {
				sensitive = true
				break
			}
		}

		if sensitive {
			filtered[key] = "[REDACTED]"
		} else {
			filtered[key] = value
		}
	}

	return filtered
}

// AuditFilter defines filtering options for audit events
type AuditFilter struct {
	StartTime time.Time
	EndTime   time.Time
	User      string
	Command   string
	Success   *bool
	Limit     int
}

// AuditStats contains audit statistics
type AuditStats struct {
	TotalEvents      int            `json:"total_events"`
	SuccessfulEvents int            `json:"successful_events"`
	FailedEvents     int            `json:"failed_events"`
	EarliestEvent    time.Time      `json:"earliest_event"`
	LatestEvent      time.Time      `json:"latest_event"`
	TopUsers         []UserStats    `json:"top_users"`
	TopCommands      []CommandStats `json:"top_commands"`
}

// UserStats contains user statistics
type UserStats struct {
	User  string `json:"user"`
	Count int    `json:"count"`
}

// CommandStats contains command statistics
type CommandStats struct {
	Command string `json:"command"`
	Count   int    `json:"count"`
}

// Package-level convenience functions

// LogAuditEvent logs an audit event using the global logger
func LogAuditEvent(event AuditEvent) error {
	return GetAuditLogger().LogEvent(event)
}

// GetAuditEvents retrieves audit events using the global logger
func GetAuditEvents(filter AuditFilter) ([]AuditEvent, error) {
	return GetAuditLogger().GetEvents(filter)
}

// GetAuditStats returns audit statistics using the global logger
func GetAuditStats() (AuditStats, error) {
	return GetAuditLogger().GetStats()
}

// CleanupAuditEvents removes old audit events using the global logger
func CleanupAuditEvents() error {
	return GetAuditLogger().CleanupOldEvents()
}

// ExportAuditEvents exports audit events using the global logger
func ExportAuditEvents(filter AuditFilter) ([]byte, error) {
	return GetAuditLogger().ExportEvents(filter)
}
