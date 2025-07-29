// Package utils provides utility functions for performance metrics and analytics
package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// MetricsCollector handles collection and storage of performance metrics
type MetricsCollector struct {
	mu      sync.RWMutex
	metrics map[string]*Metric
	config  MetricsConfig
	storage MetricsStorage
}

// MetricsConfig defines configuration for metrics collection
type MetricsConfig struct {
	Enabled       bool     `yaml:"enabled"`
	StoragePath   string   `yaml:"storage_path"`
	RetentionDays int      `yaml:"retention_days"`
	Collectors    []string `yaml:"collectors"`
	ExportFormat  string   `yaml:"export_format"`
	Aggregation   struct {
		Enabled  bool `yaml:"enabled"`
		Interval int  `yaml:"interval_minutes"`
	} `yaml:"aggregation"`
}

// DefaultMetricsConfig returns sensible defaults for metrics configuration
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		Enabled:       true,
		StoragePath:   ".mage/metrics",
		RetentionDays: 30,
		Collectors:    []string{"performance", "build", "test", "system"},
		ExportFormat:  "json",
		Aggregation: struct {
			Enabled  bool `yaml:"enabled"`
			Interval int  `yaml:"interval_minutes"`
		}{
			Enabled:  true,
			Interval: 60,
		},
	}
}

// Metric represents a single performance metric
type Metric struct {
	Name      string            `json:"name"`
	Type      MetricType        `json:"type"`
	Value     float64           `json:"value"`
	Unit      string            `json:"unit"`
	Timestamp time.Time         `json:"timestamp"`
	Tags      map[string]string `json:"tags"`
	Metadata  map[string]string `json:"metadata"`
	Duration  time.Duration     `json:"duration,omitempty"`
	Success   bool              `json:"success"`
	Error     string            `json:"error,omitempty"`
}

// MetricType defines the type of metric
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeTimer     MetricType = "timer"
	MetricTypeSummary   MetricType = "summary"
)

// PerformanceTimer tracks execution time and performance metrics
type PerformanceTimer struct {
	Name      string
	StartTime time.Time
	Tags      map[string]string
	collector *MetricsCollector
}

// BuildMetrics tracks build-specific performance metrics
type BuildMetrics struct {
	Duration        time.Duration     `json:"duration"`
	LinesOfCode     int64             `json:"lines_of_code"`
	PackagesBuilt   int               `json:"packages_built"`
	TestsRun        int               `json:"tests_run"`
	TestsPassed     int               `json:"tests_passed"`
	TestsFailed     int               `json:"tests_failed"`
	Coverage        float64           `json:"coverage"`
	BinarySize      int64             `json:"binary_size"`
	DependencyCount int               `json:"dependency_count"`
	Warnings        int               `json:"warnings"`
	Errors          int               `json:"errors"`
	Resources       ResourceMetrics   `json:"resources"`
	Timestamp       time.Time         `json:"timestamp"`
	Success         bool              `json:"success"`
	Tags            map[string]string `json:"tags"`
}

// ResourceMetrics tracks system resource usage
type ResourceMetrics struct {
	CPUUsage    float64 `json:"cpu_usage"`
	MemoryUsage int64   `json:"memory_usage"`
	DiskUsage   int64   `json:"disk_usage"`
	NetworkIO   int64   `json:"network_io"`
	FileHandles int     `json:"file_handles"`
	Goroutines  int     `json:"goroutines"`
}

// MetricsStorage interface for metrics persistence
type MetricsStorage interface {
	Store(metric *Metric) error
	Query(query MetricsQuery) ([]*Metric, error)
	Aggregate(query MetricsQuery) (*AggregatedMetrics, error)
	Cleanup(retentionDays int) error
}

// MetricsQuery defines query parameters for metrics
type MetricsQuery struct {
	StartTime time.Time
	EndTime   time.Time
	Names     []string
	Tags      map[string]string
	Limit     int
	OrderBy   string
}

// AggregatedMetrics contains aggregated metric data
type AggregatedMetrics struct {
	Count    int64   `json:"count"`
	Sum      float64 `json:"sum"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Average  float64 `json:"average"`
	Median   float64 `json:"median"`
	P95      float64 `json:"p95"`
	P99      float64 `json:"p99"`
	StdDev   float64 `json:"std_dev"`
	Variance float64 `json:"variance"`
	Period   string  `json:"period"`
}

// TrendData represents trend analysis data
type TrendData struct {
	Period    string    `json:"period"`
	Values    []float64 `json:"values"`
	Trend     string    `json:"trend"`
	Change    float64   `json:"change"`
	Timestamp time.Time `json:"timestamp"`
}

// PerformanceReport contains comprehensive performance analysis
type PerformanceReport struct {
	GeneratedAt     time.Time                        `json:"generated_at"`
	Period          string                           `json:"period"`
	TotalBuilds     int                              `json:"total_builds"`
	SuccessRate     float64                          `json:"success_rate"`
	AverageDuration time.Duration                    `json:"average_duration"`
	BuildTrends     map[string]TrendData             `json:"build_trends"`
	TestMetrics     map[string]AggregatedMetrics     `json:"test_metrics"`
	ResourceUsage   map[string]AggregatedMetrics     `json:"resource_usage"`
	Bottlenecks     []PerformanceBottleneck          `json:"bottlenecks"`
	Recommendations []PerformanceRecommendation      `json:"recommendations"`
	Comparisons     map[string]PerformanceComparison `json:"comparisons"`
}

// PerformanceBottleneck identifies performance issues
type PerformanceBottleneck struct {
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Impact      string  `json:"impact"`
	Severity    string  `json:"severity"`
	Frequency   int     `json:"frequency"`
	AvgDuration float64 `json:"avg_duration"`
	Description string  `json:"description"`
	Solution    string  `json:"solution"`
}

// PerformanceRecommendation provides optimization suggestions
type PerformanceRecommendation struct {
	Title       string  `json:"title"`
	Category    string  `json:"category"`
	Priority    string  `json:"priority"`
	Impact      string  `json:"impact"`
	Effort      string  `json:"effort"`
	Description string  `json:"description"`
	Action      string  `json:"action"`
	Confidence  float64 `json:"confidence"`
}

// PerformanceComparison compares performance across different periods
type PerformanceComparison struct {
	Metric      string  `json:"metric"`
	Current     float64 `json:"current"`
	Previous    float64 `json:"previous"`
	Change      float64 `json:"change"`
	ChangeType  string  `json:"change_type"`
	Significant bool    `json:"significant"`
}

var (
	// globalMetricsCollector is the singleton metrics collector
	globalMetricsCollector *MetricsCollector
	metricsOnce            sync.Once
)

// GetMetricsCollector returns the global metrics collector instance
func GetMetricsCollector() *MetricsCollector {
	metricsOnce.Do(func() {
		config := DefaultMetricsConfig()

		// Check if metrics are enabled via environment variable
		if os.Getenv("MAGE_METRICS_ENABLED") == "true" {
			config.Enabled = true
		}

		// Override storage path if specified
		if storagePath := os.Getenv("MAGE_METRICS_PATH"); storagePath != "" {
			config.StoragePath = storagePath
		}

		globalMetricsCollector = NewMetricsCollector(config)
	})

	return globalMetricsCollector
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(config MetricsConfig) *MetricsCollector {
	collector := &MetricsCollector{
		metrics: make(map[string]*Metric),
		config:  config,
	}

	if config.Enabled {
		// Initialize storage
		storage, err := NewJSONStorage(config.StoragePath)
		if err != nil {
			Error("Failed to initialize metrics storage: %v", err)
			collector.config.Enabled = false
		} else {
			collector.storage = storage
		}
	}

	return collector
}

// StartTimer starts a performance timer
func (mc *MetricsCollector) StartTimer(name string, tags map[string]string) *PerformanceTimer {
	if !mc.config.Enabled {
		return &PerformanceTimer{} // Return empty timer if disabled
	}

	return &PerformanceTimer{
		Name:      name,
		StartTime: time.Now(),
		Tags:      tags,
		collector: mc,
	}
}

// Stop stops the performance timer and records the metric
func (pt *PerformanceTimer) Stop() time.Duration {
	duration := time.Since(pt.StartTime)

	if pt.collector != nil && pt.collector.config.Enabled {
		metric := &Metric{
			Name:      pt.Name,
			Type:      MetricTypeTimer,
			Value:     float64(duration.Nanoseconds()),
			Unit:      "nanoseconds",
			Timestamp: time.Now(),
			Tags:      pt.Tags,
			Duration:  duration,
			Success:   true,
		}

		pt.collector.RecordMetric(metric)
	}

	return duration
}

// StopWithError stops the timer and records an error
func (pt *PerformanceTimer) StopWithError(err error) time.Duration {
	duration := time.Since(pt.StartTime)

	if pt.collector != nil && pt.collector.config.Enabled {
		metric := &Metric{
			Name:      pt.Name,
			Type:      MetricTypeTimer,
			Value:     float64(duration.Nanoseconds()),
			Unit:      "nanoseconds",
			Timestamp: time.Now(),
			Tags:      pt.Tags,
			Duration:  duration,
			Success:   false,
			Error:     err.Error(),
		}

		pt.collector.RecordMetric(metric)
	}

	return duration
}

// RecordMetric records a metric
func (mc *MetricsCollector) RecordMetric(metric *Metric) error {
	if !mc.config.Enabled {
		return nil
	}

	mc.mu.Lock()
	defer mc.mu.Unlock()

	// Store in memory
	key := fmt.Sprintf("%s_%d", metric.Name, metric.Timestamp.Unix())
	mc.metrics[key] = metric

	// Store persistently
	if mc.storage != nil {
		return mc.storage.Store(metric)
	}

	return nil
}

// RecordCounter records a counter metric
func (mc *MetricsCollector) RecordCounter(name string, value float64, tags map[string]string) error {
	metric := &Metric{
		Name:      name,
		Type:      MetricTypeCounter,
		Value:     value,
		Unit:      "count",
		Timestamp: time.Now(),
		Tags:      tags,
		Success:   true,
	}

	return mc.RecordMetric(metric)
}

// RecordGauge records a gauge metric
func (mc *MetricsCollector) RecordGauge(name string, value float64, unit string, tags map[string]string) error {
	metric := &Metric{
		Name:      name,
		Type:      MetricTypeGauge,
		Value:     value,
		Unit:      unit,
		Timestamp: time.Now(),
		Tags:      tags,
		Success:   true,
	}

	return mc.RecordMetric(metric)
}

// RecordBuildMetrics records comprehensive build metrics
func (mc *MetricsCollector) RecordBuildMetrics(buildMetrics BuildMetrics) error {
	if !mc.config.Enabled {
		return nil
	}

	// Record individual metrics
	metrics := []*Metric{
		{
			Name:      "build_duration",
			Type:      MetricTypeTimer,
			Value:     float64(buildMetrics.Duration.Nanoseconds()),
			Unit:      "nanoseconds",
			Timestamp: buildMetrics.Timestamp,
			Tags:      buildMetrics.Tags,
			Duration:  buildMetrics.Duration,
			Success:   buildMetrics.Success,
		},
		{
			Name:      "build_lines_of_code",
			Type:      MetricTypeGauge,
			Value:     float64(buildMetrics.LinesOfCode),
			Unit:      "lines",
			Timestamp: buildMetrics.Timestamp,
			Tags:      buildMetrics.Tags,
			Success:   buildMetrics.Success,
		},
		{
			Name:      "build_packages_built",
			Type:      MetricTypeCounter,
			Value:     float64(buildMetrics.PackagesBuilt),
			Unit:      "packages",
			Timestamp: buildMetrics.Timestamp,
			Tags:      buildMetrics.Tags,
			Success:   buildMetrics.Success,
		},
		{
			Name:      "build_tests_run",
			Type:      MetricTypeCounter,
			Value:     float64(buildMetrics.TestsRun),
			Unit:      "tests",
			Timestamp: buildMetrics.Timestamp,
			Tags:      buildMetrics.Tags,
			Success:   buildMetrics.Success,
		},
		{
			Name:      "build_coverage",
			Type:      MetricTypeGauge,
			Value:     buildMetrics.Coverage,
			Unit:      "percent",
			Timestamp: buildMetrics.Timestamp,
			Tags:      buildMetrics.Tags,
			Success:   buildMetrics.Success,
		},
		{
			Name:      "build_binary_size",
			Type:      MetricTypeGauge,
			Value:     float64(buildMetrics.BinarySize),
			Unit:      "bytes",
			Timestamp: buildMetrics.Timestamp,
			Tags:      buildMetrics.Tags,
			Success:   buildMetrics.Success,
		},
		{
			Name:      "build_cpu_usage",
			Type:      MetricTypeGauge,
			Value:     buildMetrics.Resources.CPUUsage,
			Unit:      "percent",
			Timestamp: buildMetrics.Timestamp,
			Tags:      buildMetrics.Tags,
			Success:   buildMetrics.Success,
		},
		{
			Name:      "build_memory_usage",
			Type:      MetricTypeGauge,
			Value:     float64(buildMetrics.Resources.MemoryUsage),
			Unit:      "bytes",
			Timestamp: buildMetrics.Timestamp,
			Tags:      buildMetrics.Tags,
			Success:   buildMetrics.Success,
		},
	}

	// Store all metrics
	for _, metric := range metrics {
		if err := mc.RecordMetric(metric); err != nil {
			return fmt.Errorf("failed to record metric %s: %w", metric.Name, err)
		}
	}

	return nil
}

// GetCurrentResourceMetrics returns current system resource usage
func (mc *MetricsCollector) GetCurrentResourceMetrics() ResourceMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return ResourceMetrics{
		CPUUsage:    getCPUUsage(),
		MemoryUsage: int64(m.Alloc),
		DiskUsage:   getDiskUsage(),
		NetworkIO:   getNetworkIO(),
		FileHandles: getFileHandles(),
		Goroutines:  runtime.NumGoroutine(),
	}
}

// QueryMetrics queries metrics based on criteria
func (mc *MetricsCollector) QueryMetrics(query MetricsQuery) ([]*Metric, error) {
	if !mc.config.Enabled || mc.storage == nil {
		return nil, fmt.Errorf("metrics collection is disabled")
	}

	return mc.storage.Query(query)
}

// GenerateReport generates a comprehensive performance report
func (mc *MetricsCollector) GenerateReport(period string) (*PerformanceReport, error) {
	if !mc.config.Enabled {
		return nil, fmt.Errorf("metrics collection is disabled")
	}

	// Define time period
	now := time.Now()
	var startTime time.Time

	switch period {
	case "day":
		startTime = now.AddDate(0, 0, -1)
	case "week":
		startTime = now.AddDate(0, 0, -7)
	case "month":
		startTime = now.AddDate(0, -1, 0)
	default:
		startTime = now.AddDate(0, 0, -7) // Default to week
		period = "week"
	}

	// Query metrics for the period
	query := MetricsQuery{
		StartTime: startTime,
		EndTime:   now,
		Limit:     10000,
		OrderBy:   "timestamp",
	}

	metrics, err := mc.QueryMetrics(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query metrics: %w", err)
	}

	// Analyze metrics
	report := &PerformanceReport{
		GeneratedAt:     now,
		Period:          period,
		BuildTrends:     make(map[string]TrendData),
		TestMetrics:     make(map[string]AggregatedMetrics),
		ResourceUsage:   make(map[string]AggregatedMetrics),
		Bottlenecks:     []PerformanceBottleneck{},
		Recommendations: []PerformanceRecommendation{},
		Comparisons:     make(map[string]PerformanceComparison),
	}

	// Analyze build metrics
	buildMetrics := filterMetrics(metrics, "build_")
	report.TotalBuilds = len(buildMetrics)

	// Calculate success rate
	successCount := 0
	for _, metric := range buildMetrics {
		if metric.Success {
			successCount++
		}
	}

	if len(buildMetrics) > 0 {
		report.SuccessRate = float64(successCount) / float64(len(buildMetrics)) * 100
	}

	// Calculate average duration
	if len(buildMetrics) > 0 {
		totalDuration := time.Duration(0)
		durationCount := 0

		for _, metric := range buildMetrics {
			if metric.Name == "build_duration" {
				totalDuration += metric.Duration
				durationCount++
			}
		}

		if durationCount > 0 {
			report.AverageDuration = totalDuration / time.Duration(durationCount)
		}
	}

	// Generate trends
	report.BuildTrends = mc.generateTrends(buildMetrics, period)

	// Aggregate test metrics
	testMetrics := filterMetrics(metrics, "test_")
	report.TestMetrics = mc.aggregateMetrics(testMetrics)

	// Aggregate resource usage
	resourceMetrics := filterMetrics(metrics, "resource_")
	report.ResourceUsage = mc.aggregateMetrics(resourceMetrics)

	// Identify bottlenecks
	report.Bottlenecks = mc.identifyBottlenecks(metrics)

	// Generate recommendations
	report.Recommendations = mc.generateRecommendations(report)

	// Generate comparisons
	report.Comparisons = mc.generateComparisons(metrics, period)

	return report, nil
}

// Cleanup removes old metrics based on retention policy
func (mc *MetricsCollector) Cleanup() error {
	if !mc.config.Enabled || mc.storage == nil {
		return nil
	}

	return mc.storage.Cleanup(mc.config.RetentionDays)
}

// Helper functions

func filterMetrics(metrics []*Metric, prefix string) []*Metric {
	var filtered []*Metric
	for _, metric := range metrics {
		if strings.HasPrefix(metric.Name, prefix) {
			filtered = append(filtered, metric)
		}
	}
	return filtered
}

func (mc *MetricsCollector) aggregateMetrics(metrics []*Metric) map[string]AggregatedMetrics {
	aggregated := make(map[string]AggregatedMetrics)

	// Group by metric name
	groups := make(map[string][]float64)
	for _, metric := range metrics {
		groups[metric.Name] = append(groups[metric.Name], metric.Value)
	}

	// Calculate aggregations
	for name, values := range groups {
		if len(values) == 0 {
			continue
		}

		sort.Float64s(values)

		sum := 0.0
		for _, v := range values {
			sum += v
		}

		avg := sum / float64(len(values))

		aggregated[name] = AggregatedMetrics{
			Count:   int64(len(values)),
			Sum:     sum,
			Min:     values[0],
			Max:     values[len(values)-1],
			Average: avg,
			Median:  percentile(values, 50),
			P95:     percentile(values, 95),
			P99:     percentile(values, 99),
		}
	}

	return aggregated
}

func (mc *MetricsCollector) generateTrends(metrics []*Metric, period string) map[string]TrendData {
	trends := make(map[string]TrendData)

	// Group by metric name
	groups := make(map[string][]float64)
	for _, metric := range metrics {
		groups[metric.Name] = append(groups[metric.Name], metric.Value)
	}

	for name, values := range groups {
		if len(values) < 2 {
			continue
		}

		// Calculate trend
		first := values[0]
		last := values[len(values)-1]
		change := ((last - first) / first) * 100

		trendDirection := "stable"
		if change > 5 {
			trendDirection = "increasing"
		} else if change < -5 {
			trendDirection = "decreasing"
		}

		trends[name] = TrendData{
			Period:    period,
			Values:    values,
			Trend:     trendDirection,
			Change:    change,
			Timestamp: time.Now(),
		}
	}

	return trends
}

func (mc *MetricsCollector) identifyBottlenecks(metrics []*Metric) []PerformanceBottleneck {
	bottlenecks := []PerformanceBottleneck{}

	// Analyze build duration bottlenecks
	buildDurations := filterMetrics(metrics, "build_duration")
	if len(buildDurations) > 0 {
		var durations []float64
		for _, metric := range buildDurations {
			durations = append(durations, metric.Value)
		}

		sort.Float64s(durations)
		p95 := percentile(durations, 95)
		avg := average(durations)

		if p95 > avg*2 {
			bottlenecks = append(bottlenecks, PerformanceBottleneck{
				Name:        "Slow Build Times",
				Category:    "Build Performance",
				Impact:      "High",
				Severity:    "Medium",
				Frequency:   len(buildDurations),
				AvgDuration: avg,
				Description: "Build times are inconsistent with some builds taking significantly longer",
				Solution:    "Consider parallel builds, dependency caching, or incremental builds",
			})
		}
	}

	// Analyze test bottlenecks
	testDurations := filterMetrics(metrics, "test_duration")
	if len(testDurations) > 0 {
		var durations []float64
		for _, metric := range testDurations {
			durations = append(durations, metric.Value)
		}

		sort.Float64s(durations)
		p95 := percentile(durations, 95)
		avg := average(durations)

		if p95 > avg*3 {
			bottlenecks = append(bottlenecks, PerformanceBottleneck{
				Name:        "Slow Test Execution",
				Category:    "Test Performance",
				Impact:      "Medium",
				Severity:    "Low",
				Frequency:   len(testDurations),
				AvgDuration: avg,
				Description: "Some tests are taking significantly longer than others",
				Solution:    "Optimize slow tests, use parallel test execution, or mock external dependencies",
			})
		}
	}

	return bottlenecks
}

func (mc *MetricsCollector) generateRecommendations(report *PerformanceReport) []PerformanceRecommendation {
	recommendations := []PerformanceRecommendation{}

	// Analyze success rate
	if report.SuccessRate < 95 {
		recommendations = append(recommendations, PerformanceRecommendation{
			Title:       "Improve Build Reliability",
			Category:    "Reliability",
			Priority:    "High",
			Impact:      "High",
			Effort:      "Medium",
			Description: fmt.Sprintf("Build success rate is %.1f%%, below the recommended 95%%", report.SuccessRate),
			Action:      "Investigate and fix failing builds, improve error handling",
			Confidence:  0.9,
		})
	}

	// Analyze build duration
	if report.AverageDuration > 5*time.Minute {
		recommendations = append(recommendations, PerformanceRecommendation{
			Title:       "Optimize Build Performance",
			Category:    "Performance",
			Priority:    "Medium",
			Impact:      "Medium",
			Effort:      "High",
			Description: fmt.Sprintf("Average build duration is %v, consider optimization", report.AverageDuration),
			Action:      "Implement build caching, parallel compilation, or incremental builds",
			Confidence:  0.8,
		})
	}

	// Analyze resource usage
	for metric, aggregated := range report.ResourceUsage {
		if metric == "build_memory_usage" && aggregated.Average > 1024*1024*1024 { // 1GB
			recommendations = append(recommendations, PerformanceRecommendation{
				Title:       "Optimize Memory Usage",
				Category:    "Resource Usage",
				Priority:    "Low",
				Impact:      "Medium",
				Effort:      "Medium",
				Description: fmt.Sprintf("Average memory usage is %.1f MB", aggregated.Average/(1024*1024)),
				Action:      "Profile memory usage and optimize high-memory operations",
				Confidence:  0.7,
			})
		}
	}

	return recommendations
}

func (mc *MetricsCollector) generateComparisons(metrics []*Metric, period string) map[string]PerformanceComparison {
	comparisons := make(map[string]PerformanceComparison)

	// This would compare current period with previous period
	// Implementation would depend on having historical data

	return comparisons
}

// System resource helper functions (placeholder implementations)
func getCPUUsage() float64 {
	// Implementation would use system calls to get CPU usage
	return 0.0
}

func getDiskUsage() int64 {
	// Implementation would check disk usage
	return 0
}

func getNetworkIO() int64 {
	// Implementation would check network I/O
	return 0
}

func getFileHandles() int {
	// Implementation would check open file handles
	return 0
}

func percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if len(values) == 1 {
		return values[0]
	}

	// Use linear interpolation for percentile calculation
	pos := p / 100 * float64(len(values)-1)
	if pos == float64(int(pos)) {
		// Exact position
		return values[int(pos)]
	}

	// Interpolate between two positions
	lower := int(pos)
	upper := lower + 1
	if upper >= len(values) {
		upper = len(values) - 1
	}

	fraction := pos - float64(lower)
	return values[lower] + fraction*(values[upper]-values[lower])
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// Package-level convenience functions

// StartTimer starts a performance timer using the global collector
func StartTimer(name string, tags map[string]string) *PerformanceTimer {
	return GetMetricsCollector().StartTimer(name, tags)
}

// RecordCounter records a counter metric using the global collector
func RecordCounter(name string, value float64, tags map[string]string) error {
	return GetMetricsCollector().RecordCounter(name, value, tags)
}

// RecordGauge records a gauge metric using the global collector
func RecordGauge(name string, value float64, unit string, tags map[string]string) error {
	return GetMetricsCollector().RecordGauge(name, value, unit, tags)
}

// RecordBuildMetrics records build metrics using the global collector
func RecordBuildMetrics(buildMetrics BuildMetrics) error {
	return GetMetricsCollector().RecordBuildMetrics(buildMetrics)
}

// GeneratePerformanceReport generates a performance report using the global collector
func GeneratePerformanceReport(period string) (*PerformanceReport, error) {
	return GetMetricsCollector().GenerateReport(period)
}

// CleanupMetrics removes old metrics using the global collector
func CleanupMetrics() error {
	return GetMetricsCollector().Cleanup()
}

// GetCurrentResourceMetrics returns current system resource metrics
func GetCurrentResourceMetrics() ResourceMetrics {
	return GetMetricsCollector().GetCurrentResourceMetrics()
}

// Simple JSON storage implementation
type JSONStorage struct {
	storagePath string
	mu          sync.RWMutex
}

// NewJSONStorage creates a new JSON storage instance
func NewJSONStorage(storagePath string) (*JSONStorage, error) {
	if err := os.MkdirAll(storagePath, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &JSONStorage{
		storagePath: storagePath,
	}, nil
}

// Store stores a metric in JSON format
func (js *JSONStorage) Store(metric *Metric) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	// Create filename based on date
	filename := fmt.Sprintf("metrics_%s.json", metric.Timestamp.Format("2006-01-02"))
	filePath := filepath.Join(js.storagePath, filename)

	// Read existing metrics
	var metrics []*Metric
	if data, err := os.ReadFile(filePath); err == nil {
		json.Unmarshal(data, &metrics)
	}

	// Append new metric
	metrics = append(metrics, metric)

	// Write back to file
	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metrics: %w", err)
	}

	return os.WriteFile(filePath, data, 0o644)
}

// Query queries metrics from JSON storage
func (js *JSONStorage) Query(query MetricsQuery) ([]*Metric, error) {
	js.mu.RLock()
	defer js.mu.RUnlock()

	var allMetrics []*Metric

	// Read all metric files in the date range (inclusive of end date)
	for d := query.StartTime; d.Before(query.EndTime.AddDate(0, 0, 1)); d = d.AddDate(0, 0, 1) {
		filename := fmt.Sprintf("metrics_%s.json", d.Format("2006-01-02"))
		filePath := filepath.Join(js.storagePath, filename)

		if data, err := os.ReadFile(filePath); err == nil {
			var metrics []*Metric
			if err := json.Unmarshal(data, &metrics); err == nil {
				allMetrics = append(allMetrics, metrics...)
			}
		}
	}

	// Filter metrics based on query
	var filteredMetrics []*Metric
	for _, metric := range allMetrics {
		if (metric.Timestamp.After(query.StartTime) || metric.Timestamp.Equal(query.StartTime)) &&
			(metric.Timestamp.Before(query.EndTime) || metric.Timestamp.Equal(query.EndTime)) {
			// Check names filter
			if len(query.Names) > 0 {
				found := false
				for _, name := range query.Names {
					if metric.Name == name {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}

			// Check tags filter
			if len(query.Tags) > 0 {
				match := true
				for key, value := range query.Tags {
					if metric.Tags[key] != value {
						match = false
						break
					}
				}
				if !match {
					continue
				}
			}

			filteredMetrics = append(filteredMetrics, metric)
		}
	}

	// Apply limit
	if query.Limit > 0 && len(filteredMetrics) > query.Limit {
		filteredMetrics = filteredMetrics[:query.Limit]
	}

	return filteredMetrics, nil
}

// Aggregate aggregates metrics from JSON storage
func (js *JSONStorage) Aggregate(query MetricsQuery) (*AggregatedMetrics, error) {
	metrics, err := js.Query(query)
	if err != nil {
		return nil, err
	}

	if len(metrics) == 0 {
		return &AggregatedMetrics{}, nil
	}

	values := make([]float64, len(metrics))
	for i, metric := range metrics {
		values[i] = metric.Value
	}

	sort.Float64s(values)

	sum := 0.0
	for _, v := range values {
		sum += v
	}

	return &AggregatedMetrics{
		Count:   int64(len(values)),
		Sum:     sum,
		Min:     values[0],
		Max:     values[len(values)-1],
		Average: sum / float64(len(values)),
		Median:  percentile(values, 50),
		P95:     percentile(values, 95),
		P99:     percentile(values, 99),
	}, nil
}

// Cleanup removes old metric files
func (js *JSONStorage) Cleanup(retentionDays int) error {
	js.mu.Lock()
	defer js.mu.Unlock()

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	entries, err := os.ReadDir(js.storagePath)
	if err != nil {
		return fmt.Errorf("failed to read storage directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Parse date from filename
		if strings.HasPrefix(entry.Name(), "metrics_") && strings.HasSuffix(entry.Name(), ".json") {
			dateStr := strings.TrimPrefix(entry.Name(), "metrics_")
			dateStr = strings.TrimSuffix(dateStr, ".json")

			if date, err := time.Parse("2006-01-02", dateStr); err == nil {
				if date.Before(cutoffDate) {
					filePath := filepath.Join(js.storagePath, entry.Name())
					if err := os.Remove(filePath); err != nil {
						Error("Failed to remove old metric file %s: %v", filePath, err)
					}
				}
			}
		}
	}

	return nil
}
