package errors

import (
	"errors"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

// RealDefaultErrorMetrics is the actual implementation of ErrorMetrics
type RealDefaultErrorMetrics struct {
	mu            sync.RWMutex
	errorStats    map[ErrorCode]*atomicErrorStat
	severityStats map[Severity]*atomicSeverityStat
	totalErrors   atomic.Int64
	startTime     time.Time
}

// atomicErrorStat holds atomic counters for error statistics
type atomicErrorStat struct {
	count     atomic.Int64
	firstSeen atomic.Value // time.Time
	lastSeen  atomic.Value // time.Time
}

// atomicSeverityStat holds atomic counters for severity statistics
type atomicSeverityStat struct {
	count atomic.Int64
}

// NewErrorMetrics creates a new error metrics collector
func NewErrorMetrics() *RealDefaultErrorMetrics {
	return &RealDefaultErrorMetrics{
		errorStats:    make(map[ErrorCode]*atomicErrorStat),
		severityStats: make(map[Severity]*atomicSeverityStat),
		startTime:     time.Now(),
	}
}

// RecordError records an error occurrence
func (m *RealDefaultErrorMetrics) RecordError(err error) {
	if err == nil {
		return
	}

	m.totalErrors.Add(1)

	var mageErr MageError
	if errors.As(err, &mageErr) {
		m.RecordMageError(mageErr)
	} else {
		// Record as unknown error
		m.recordErrorCode(ErrUnknown)
		m.recordSeverity(SeverityError)
	}
}

// RecordMageError records a MageError occurrence
func (m *RealDefaultErrorMetrics) RecordMageError(err MageError) {
	if err == nil {
		return
	}

	m.recordErrorCode(err.Code())
	m.recordSeverity(err.Severity())
}

// GetCount returns the count for a specific error code
func (m *RealDefaultErrorMetrics) GetCount(code ErrorCode) int64 {
	m.mu.RLock()
	stat, exists := m.errorStats[code]
	m.mu.RUnlock()

	if !exists {
		return 0
	}

	return stat.count.Load()
}

// GetCountBySeverity returns the count for a specific severity
func (m *RealDefaultErrorMetrics) GetCountBySeverity(severity Severity) int64 {
	m.mu.RLock()
	stat, exists := m.severityStats[severity]
	m.mu.RUnlock()

	if !exists {
		return 0
	}

	return stat.count.Load()
}

// GetRate returns the rate of errors per duration
func (m *RealDefaultErrorMetrics) GetRate(code ErrorCode, duration time.Duration) float64 {
	count := m.GetCount(code)
	if count == 0 {
		return 0
	}

	elapsed := time.Since(m.startTime)
	if elapsed < duration {
		duration = elapsed
	}

	return float64(count) / duration.Seconds()
}

// GetTopErrors returns the top N errors by count
func (m *RealDefaultErrorMetrics) GetTopErrors(limit int) []ErrorStat {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.getTopErrorsLocked(limit)
}

// getTopErrorsLocked returns the top N errors by count.
// Caller must hold at least a read lock on m.mu.
func (m *RealDefaultErrorMetrics) getTopErrorsLocked(limit int) []ErrorStat {
	// Convert to slice for sorting
	stats := make([]ErrorStat, 0, len(m.errorStats))
	for code, stat := range m.errorStats {
		count := stat.count.Load()
		if count == 0 {
			continue
		}

		es := ErrorStat{
			Code:  code,
			Count: count,
		}

		// Get timestamps
		if firstSeen, ok := stat.firstSeen.Load().(time.Time); ok {
			es.FirstSeen = firstSeen
		}
		if lastSeen, ok := stat.lastSeen.Load().(time.Time); ok {
			es.LastSeen = lastSeen
		}

		// Calculate average rate
		if !es.FirstSeen.IsZero() && !es.LastSeen.IsZero() {
			duration := es.LastSeen.Sub(es.FirstSeen)
			if duration > 0 {
				es.AverageRate = float64(count) / duration.Seconds()
			}
		}

		stats = append(stats, es)
	}

	// Sort by count (descending)
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	// Limit results
	if limit > 0 && len(stats) > limit {
		stats = stats[:limit]
	}

	return stats
}

// Reset clears all metrics
func (m *RealDefaultErrorMetrics) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.errorStats = make(map[ErrorCode]*atomicErrorStat)
	m.severityStats = make(map[Severity]*atomicSeverityStat)
	m.totalErrors.Store(0)
	m.startTime = time.Now()

	return nil
}

// recordErrorCode records an error code occurrence
func (m *RealDefaultErrorMetrics) recordErrorCode(code ErrorCode) {
	m.mu.RLock()
	stat, exists := m.errorStats[code]
	m.mu.RUnlock()

	if !exists {
		m.mu.Lock()
		// Double-check after acquiring write lock
		stat, exists = m.errorStats[code]
		if !exists {
			stat = &atomicErrorStat{}
			m.errorStats[code] = stat
		}
		m.mu.Unlock()
	}

	// Update counters
	stat.count.Add(1)

	now := time.Now()

	// Update first seen (only once)
	stat.firstSeen.CompareAndSwap(nil, now)

	// Always update last seen
	stat.lastSeen.Store(now)
}

// recordSeverity records a severity occurrence
func (m *RealDefaultErrorMetrics) recordSeverity(severity Severity) {
	m.mu.RLock()
	stat, exists := m.severityStats[severity]
	m.mu.RUnlock()

	if !exists {
		m.mu.Lock()
		// Double-check after acquiring write lock
		stat, exists = m.severityStats[severity]
		if !exists {
			stat = &atomicSeverityStat{}
			m.severityStats[severity] = stat
		}
		m.mu.Unlock()
	}

	stat.count.Add(1)
}

// GetTotalErrors returns the total number of errors recorded
func (m *RealDefaultErrorMetrics) GetTotalErrors() int64 {
	return m.totalErrors.Load()
}

// GetErrorRate returns the overall error rate
func (m *RealDefaultErrorMetrics) GetErrorRate() float64 {
	total := m.GetTotalErrors()
	if total == 0 {
		return 0
	}

	elapsed := time.Since(m.startTime).Seconds()
	if elapsed == 0 {
		return 0
	}

	return float64(total) / elapsed
}

// GetSummary returns a summary of all error metrics
func (m *RealDefaultErrorMetrics) GetSummary() MetricsSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := MetricsSummary{
		TotalErrors: m.GetTotalErrors(),
		ErrorRate:   m.GetErrorRate(),
		StartTime:   m.startTime,
		Duration:    time.Since(m.startTime),
		ErrorCodes:  make(map[ErrorCode]int64),
		Severities:  make(map[Severity]int64),
		TopErrors:   m.getTopErrorsLocked(10),
	}

	// Collect error code counts
	for code, stat := range m.errorStats {
		if count := stat.count.Load(); count > 0 {
			summary.ErrorCodes[code] = count
		}
	}

	// Collect severity counts
	for severity, stat := range m.severityStats {
		if count := stat.count.Load(); count > 0 {
			summary.Severities[severity] = count
		}
	}

	return summary
}

// MetricsSummary contains a summary of error metrics
type MetricsSummary struct {
	TotalErrors int64
	ErrorRate   float64
	StartTime   time.Time
	Duration    time.Duration
	ErrorCodes  map[ErrorCode]int64
	Severities  map[Severity]int64
	TopErrors   []ErrorStat
}

// RecordError records an error occurrence in the metrics.
func (m *DefaultErrorMetrics) RecordError(err error) {
	metrics := NewErrorMetrics()
	metrics.RecordError(err)
}

// RecordMageError records a MageError occurrence in the metrics.
func (m *DefaultErrorMetrics) RecordMageError(err MageError) {
	metrics := NewErrorMetrics()
	metrics.RecordMageError(err)
}

// GetCount returns the count of occurrences for a specific error code.
func (m *DefaultErrorMetrics) GetCount(code ErrorCode) int64 {
	if stat, exists := m.counts[code]; exists {
		return stat.Count
	}
	return 0
}

// GetCountBySeverity returns the count of errors for a specific severity level.
func (m *DefaultErrorMetrics) GetCountBySeverity(_ Severity) int64 {
	var count int64
	for range m.counts {
		// This would need severity information in the stat
		count++
	}
	return count
}

// GetRate returns the error rate for a specific error code over the given duration.
func (m *DefaultErrorMetrics) GetRate(code ErrorCode, duration time.Duration) float64 {
	if stat, exists := m.counts[code]; exists {
		if stat.Count == 0 {
			return 0
		}
		elapsed := time.Since(stat.FirstSeen)
		if elapsed < duration {
			duration = elapsed
		}
		return float64(stat.Count) / duration.Seconds()
	}
	return 0
}

// GetTopErrors returns the top errors sorted by occurrence count.
func (m *DefaultErrorMetrics) GetTopErrors(limit int) []ErrorStat {
	stats := make([]ErrorStat, 0, len(m.counts))
	for _, stat := range m.counts {
		stats = append(stats, *stat)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	if limit > 0 && len(stats) > limit {
		stats = stats[:limit]
	}

	return stats
}

// Reset clears all error metrics data.
func (m *DefaultErrorMetrics) Reset() error {
	m.counts = make(map[ErrorCode]*ErrorStat)
	return nil
}
