// Package utils provides utility functions for audit logging
package utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Static errors for audit operations
var (
	errAuditLoggingDisabled = errors.New("audit logging is disabled")
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

// AuditRegistry manages audit logger instances with thread-safe access
type AuditRegistry struct {
	mu       sync.RWMutex
	loggers  map[string]*AuditLogger
	default_ *AuditLogger
}

// NewAuditRegistry creates a new audit registry
func NewAuditRegistry() *AuditRegistry {
	return &AuditRegistry{
		loggers: make(map[string]*AuditLogger),
	}
}

// defaultRegistry is the package-level registry for backward compatibility
var defaultRegistry = NewAuditRegistry() //nolint:gochecknoglobals // Needed for backward compatibility

// GetOrCreateLogger gets or creates an audit logger with the given name
func (r *AuditRegistry) GetOrCreateLogger(name string, config *AuditConfig) *AuditLogger {
	r.mu.RLock()
	logger, exists := r.loggers[name]
	r.mu.RUnlock()

	if exists {
		return logger
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if existingLogger, exists := r.loggers[name]; exists {
		return existingLogger
	}

	// Create new logger
	logger = NewAuditLogger(config)
	r.loggers[name] = logger

	return logger
}

// GetLogger retrieves an existing logger by name
func (r *AuditRegistry) GetLogger(name string) (*AuditLogger, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	logger, exists := r.loggers[name]
	return logger, exists
}

// SetDefaultLogger sets the default logger for the registry
func (r *AuditRegistry) SetDefaultLogger(logger *AuditLogger) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.default_ = logger
}

// GetDefaultLogger returns the default logger, creating one if necessary
func (r *AuditRegistry) GetDefaultLogger() *AuditLogger {
	r.mu.RLock()
	if r.default_ != nil {
		defer r.mu.RUnlock()
		return r.default_
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if r.default_ != nil {
		return r.default_
	}

	// Create default logger with environment-based configuration
	config := DefaultAuditConfig()

	// Check if audit is enabled via environment variable
	if os.Getenv("MAGE_X_AUDIT_ENABLED") == TrueValue {
		config.Enabled = true
	}

	// Override database path if specified
	if dbPath := os.Getenv("MAGE_X_AUDIT_DB"); dbPath != "" {
		config.DatabasePath = dbPath
	}

	r.default_ = NewAuditLogger(&config)
	return r.default_
}

// CloseAll closes all registered loggers
func (r *AuditRegistry) CloseAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var firstErr error
	for name, logger := range r.loggers {
		if err := logger.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("failed to close logger %s: %w", name, err)
		}
	}

	if r.default_ != nil {
		if err := r.default_.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("failed to close default logger: %w", err)
		}
	}

	// Clear the maps
	r.loggers = make(map[string]*AuditLogger)
	r.default_ = nil

	return firstErr
}

// GetAuditLogger returns the default audit logger instance for backward compatibility
func GetAuditLogger() *AuditLogger {
	return defaultRegistry.GetDefaultLogger()
}

// GetAuditRegistry returns the default audit registry for advanced usage
func GetAuditRegistry() *AuditRegistry {
	return defaultRegistry
}

// NewAuditLoggerWithName creates a named audit logger that can be retrieved later
func NewAuditLoggerWithName(name string, config *AuditConfig) *AuditLogger {
	return defaultRegistry.GetOrCreateLogger(name, config)
}

// GetAuditLoggerByName retrieves an existing named audit logger
func GetAuditLoggerByName(name string) (*AuditLogger, bool) {
	return defaultRegistry.GetLogger(name)
}

// CloseAllAuditLoggers closes all audit loggers managed by the default registry
func CloseAllAuditLoggers() error {
	return defaultRegistry.CloseAll()
}

// NewAuditLogger creates a new audit logger with the given configuration
func NewAuditLogger(config *AuditConfig) *AuditLogger {
	logger := &AuditLogger{
		enabled: config.Enabled,
		config:  *config,
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
	if err := os.MkdirAll(dbDir, 0o750); err != nil {
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

	if _, err := db.ExecContext(context.Background(), createTableSQL); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			// Log close error but return original error
			log.Printf("failed to close database during error handling: %v", closeErr)
		}
		return fmt.Errorf("failed to create audit table: %w", err)
	}

	a.db = db
	return nil
}

// LogEvent logs an audit event
func (a *AuditLogger) LogEvent(event *AuditEvent) error {
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
	argsJSON, err := json.Marshal(event.Args)
	if err != nil {
		argsJSON = []byte("[]") // Fallback to empty array
	}
	envJSON, err := json.Marshal(event.Environment)
	if err != nil {
		envJSON = []byte("{}") // Fallback to empty object
	}
	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		metadataJSON = []byte("{}") // Fallback to empty object
	}

	// Insert into database
	query := `
	INSERT INTO audit_events (
		timestamp, user, command, args, working_dir,
		duration, exit_code, success, environment, metadata
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, execErr := a.db.ExecContext(context.Background(), query,
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

	return execErr
}

// GetEvents retrieves audit events with optional filtering
func (a *AuditLogger) GetEvents(filter *AuditFilter) ([]AuditEvent, error) {
	if !a.enabled {
		return nil, errAuditLoggingDisabled
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	query, args := a.buildEventsQuery(filter)

	rows, err := a.db.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit events: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// Log close error but don't fail the operation
			log.Printf("failed to close database rows: %v", closeErr)
		}
	}()

	events, err := a.scanAuditEvents(rows)
	if err != nil {
		return nil, err
	}

	return events, nil
}

// buildEventsQuery constructs the SQL query and args for event filtering
func (a *AuditLogger) buildEventsQuery(filter *AuditFilter) (query string, args []interface{}) {
	query = `
	SELECT id, timestamp, user, command, args, working_dir,
		   duration, exit_code, success, environment, metadata
	FROM audit_events
	WHERE 1=1`

	args = []interface{}{}

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

	return query, args
}

// scanAuditEvents processes database rows into AuditEvent structs
func (a *AuditLogger) scanAuditEvents(rows *sql.Rows) ([]AuditEvent, error) {
	var events []AuditEvent
	for rows.Next() {
		event, err := a.scanSingleAuditEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	// Check for errors during iteration
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return events, nil
}

// scanSingleAuditEvent scans a single row into an AuditEvent
func (a *AuditLogger) scanSingleAuditEvent(rows *sql.Rows) (AuditEvent, error) {
	var event AuditEvent
	var argsJSON, envJSON, metadataJSON string
	var durationNs int64

	scanErr := rows.Scan(
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
	if scanErr != nil {
		return event, fmt.Errorf("failed to scan audit event: %w", scanErr)
	}

	// Deserialize JSON fields
	event.Duration = time.Duration(durationNs)
	a.unmarshalJSONFields(&event, argsJSON, envJSON, metadataJSON)

	return event, nil
}

// unmarshalJSONFields deserializes JSON fields for an audit event
func (a *AuditLogger) unmarshalJSONFields(event *AuditEvent, argsJSON, envJSON, metadataJSON string) {
	if unmarshalErr := json.Unmarshal([]byte(argsJSON), &event.Args); unmarshalErr != nil {
		// Continue with empty args on unmarshal error
		log.Printf("failed to unmarshal args JSON: %v", unmarshalErr)
	}
	if unmarshalErr := json.Unmarshal([]byte(envJSON), &event.Environment); unmarshalErr != nil {
		// Continue with empty environment on unmarshal error
		log.Printf("failed to unmarshal environment JSON: %v", unmarshalErr)
	}
	if unmarshalErr := json.Unmarshal([]byte(metadataJSON), &event.Metadata); unmarshalErr != nil {
		// Continue with empty metadata on unmarshal error
		log.Printf("failed to unmarshal metadata JSON: %v", unmarshalErr)
	}
}

// GetStats returns audit statistics
func (a *AuditLogger) GetStats() (AuditStats, error) {
	if !a.enabled {
		return AuditStats{}, errAuditLoggingDisabled
	}

	a.mu.RLock()
	defer a.mu.RUnlock()

	var stats AuditStats

	if err := a.getBasicStats(&stats); err != nil {
		return stats, err
	}

	if err := a.getTopUsers(&stats); err != nil {
		return stats, err
	}

	if err := a.getTopCommands(&stats); err != nil {
		return stats, err
	}

	return stats, nil
}

// getBasicStats retrieves basic statistics (counts and date range)
func (a *AuditLogger) getBasicStats(stats *AuditStats) error {
	// Total events
	err := a.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM audit_events").Scan(&stats.TotalEvents)
	if err != nil {
		return fmt.Errorf("failed to get total events: %w", err)
	}

	// Success/failure counts
	err = a.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM audit_events WHERE success = 1").Scan(&stats.SuccessfulEvents)
	if err != nil {
		return fmt.Errorf("failed to get successful events: %w", err)
	}

	err = a.db.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM audit_events WHERE success = 0").Scan(&stats.FailedEvents)
	if err != nil {
		return fmt.Errorf("failed to get failed events: %w", err)
	}

	// Date range
	err = a.db.QueryRowContext(context.Background(), "SELECT MIN(timestamp), MAX(timestamp) FROM audit_events").Scan(&stats.EarliestEvent, &stats.LatestEvent)
	if err != nil {
		return fmt.Errorf("failed to get date range: %w", err)
	}

	return nil
}

// getTopUsers retrieves top users statistics
func (a *AuditLogger) getTopUsers(stats *AuditStats) error {
	rows, err := a.db.QueryContext(context.Background(), `
		SELECT user, COUNT(*) as count
		FROM audit_events
		GROUP BY user
		ORDER BY count DESC
		LIMIT 10
	`)
	if err != nil {
		return fmt.Errorf("failed to get top users: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// Log error but don't fail the operation
			log.Printf("failed to close rows for top users query: %v", closeErr)
		}
	}()

	for rows.Next() {
		var user string
		var count int
		if scanErr := rows.Scan(&user, &count); scanErr != nil {
			continue
		}
		stats.TopUsers = append(stats.TopUsers, UserStats{User: user, Count: count})
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		return fmt.Errorf("error during top users iteration: %w", err)
	}

	return nil
}

// getTopCommands retrieves top commands statistics
func (a *AuditLogger) getTopCommands(stats *AuditStats) error {
	rows, err := a.db.QueryContext(context.Background(), `
		SELECT command, COUNT(*) as count
		FROM audit_events
		GROUP BY command
		ORDER BY count DESC
		LIMIT 10
	`)
	if err != nil {
		return fmt.Errorf("failed to get top commands: %w", err)
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			// Log error but don't fail the operation
			log.Printf("failed to close rows for top commands query: %v", closeErr)
		}
	}()

	for rows.Next() {
		var command string
		var count int
		if scanErr := rows.Scan(&command, &count); scanErr != nil {
			continue
		}
		stats.TopCommands = append(stats.TopCommands, CommandStats{Command: command, Count: count})
	}

	// Check for errors during iteration
	if err = rows.Err(); err != nil {
		return fmt.Errorf("error during top commands iteration: %w", err)
	}

	return nil
}

// CleanupOldEvents removes old audit events based on retention policy
func (a *AuditLogger) CleanupOldEvents() error {
	if !a.enabled {
		return nil
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	cutoffTime := time.Now().AddDate(0, 0, -a.config.RetentionDays)

	result, err := a.db.ExecContext(context.Background(), "DELETE FROM audit_events WHERE timestamp < ?", cutoffTime)
	if err != nil {
		return fmt.Errorf("failed to cleanup old events: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// Continue if we can't get rows affected count
		rowsAffected = 0
	}
	if rowsAffected > 0 {
		Info("Cleaned up %d old audit events", rowsAffected)
	}

	return nil
}

// ExportEvents exports audit events to JSON
func (a *AuditLogger) ExportEvents(filter *AuditFilter) ([]byte, error) {
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
func LogAuditEvent(event *AuditEvent) error {
	return GetAuditLogger().LogEvent(event)
}

// GetAuditEvents retrieves audit events using the global logger
func GetAuditEvents(filter *AuditFilter) ([]AuditEvent, error) {
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
func ExportAuditEvents(filter *AuditFilter) ([]byte, error) {
	return GetAuditLogger().ExportEvents(filter)
}
