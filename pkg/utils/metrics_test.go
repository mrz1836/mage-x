package utils

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultMetricsConfig(t *testing.T) {
	config := DefaultMetricsConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, ".mage/metrics", config.StoragePath)
	assert.Equal(t, 30, config.RetentionDays)
	assert.Equal(t, "json", config.ExportFormat)
	assert.Contains(t, config.Collectors, "performance")
	assert.Contains(t, config.Collectors, "build")
	assert.True(t, config.Aggregation.Enabled)
	assert.Equal(t, 60, config.Aggregation.Interval)
}

func TestMetricsCollector_Basic(t *testing.T) {
	tempDir := t.TempDir()
	config := MetricsConfig{
		Enabled:       true,
		StoragePath:   filepath.Join(tempDir, "metrics"),
		RetentionDays: 7,
		ExportFormat:  "json",
	}

	t.Run("NewMetricsCollector creates collector", func(t *testing.T) {
		collector := NewMetricsCollector(config)
		assert.NotNil(t, collector)
		assert.Equal(t, config, collector.config)
		assert.NotNil(t, collector.metrics)
	})

	t.Run("disabled collector", func(t *testing.T) {
		disabledConfig := config
		disabledConfig.Enabled = false

		collector := NewMetricsCollector(disabledConfig)
		assert.NotNil(t, collector)
		assert.False(t, collector.config.Enabled)
	})
}

func TestPerformanceTimer(t *testing.T) {
	tempDir := t.TempDir()
	config := MetricsConfig{
		Enabled:     true,
		StoragePath: filepath.Join(tempDir, "metrics"),
	}
	collector := NewMetricsCollector(config)

	t.Run("StartTimer creates timer", func(t *testing.T) {
		tags := map[string]string{"operation": "test"}
		timer := collector.StartTimer("test_operation", tags)

		assert.NotNil(t, timer)
		assert.Equal(t, "test_operation", timer.Name)
		assert.Equal(t, tags, timer.Tags)
		assert.Equal(t, collector, timer.collector)
		assert.False(t, timer.StartTime.IsZero())
	})

	t.Run("Stop records metric", func(t *testing.T) {
		timer := collector.StartTimer("test_stop", nil)
		time.Sleep(10 * time.Millisecond) // Ensure some duration

		duration := timer.Stop()

		assert.True(t, duration > 0)
		assert.True(t, duration >= 10*time.Millisecond)

		// Check that metric was recorded
		assert.True(t, len(collector.metrics) > 0)
	})

	t.Run("StopWithError records error metric", func(t *testing.T) {
		timer := collector.StartTimer("test_error", nil)
		time.Sleep(5 * time.Millisecond)

		testErr := assert.AnError
		duration := timer.StopWithError(testErr)

		assert.True(t, duration > 0)

		// Find the recorded metric
		var errorMetric *Metric
		for _, metric := range collector.metrics {
			if metric.Name == "test_error" {
				errorMetric = metric
				break
			}
		}

		require.NotNil(t, errorMetric)
		assert.False(t, errorMetric.Success)
		assert.Equal(t, testErr.Error(), errorMetric.Error)
	})

	t.Run("disabled collector timer", func(t *testing.T) {
		disabledConfig := config
		disabledConfig.Enabled = false
		disabledCollector := NewMetricsCollector(disabledConfig)

		timer := disabledCollector.StartTimer("disabled_test", nil)
		duration := timer.Stop()

		// Should work but not record metrics
		assert.True(t, duration >= 0)
		assert.Equal(t, 0, len(disabledCollector.metrics))
	})
}

func TestMetricsRecording(t *testing.T) {
	tempDir := t.TempDir()
	config := MetricsConfig{
		Enabled:     true,
		StoragePath: filepath.Join(tempDir, "metrics"),
	}
	collector := NewMetricsCollector(config)

	t.Run("RecordCounter", func(t *testing.T) {
		tags := map[string]string{"component": "test"}
		err := collector.RecordCounter("test_counter", 42.0, tags)
		require.NoError(t, err)

		// Find the recorded metric
		var counterMetric *Metric
		for _, metric := range collector.metrics {
			if metric.Name == "test_counter" {
				counterMetric = metric
				break
			}
		}

		require.NotNil(t, counterMetric)
		assert.Equal(t, MetricTypeCounter, counterMetric.Type)
		assert.Equal(t, 42.0, counterMetric.Value)
		assert.Equal(t, "count", counterMetric.Unit)
		assert.Equal(t, tags, counterMetric.Tags)
		assert.True(t, counterMetric.Success)
	})

	t.Run("RecordGauge", func(t *testing.T) {
		tags := map[string]string{"type": "memory"}
		err := collector.RecordGauge("memory_usage", 1024.0, "bytes", tags)
		require.NoError(t, err)

		// Find the recorded metric
		var gaugeMetric *Metric
		for _, metric := range collector.metrics {
			if metric.Name == "memory_usage" {
				gaugeMetric = metric
				break
			}
		}

		require.NotNil(t, gaugeMetric)
		assert.Equal(t, MetricTypeGauge, gaugeMetric.Type)
		assert.Equal(t, 1024.0, gaugeMetric.Value)
		assert.Equal(t, "bytes", gaugeMetric.Unit)
		assert.Equal(t, tags, gaugeMetric.Tags)
	})

	t.Run("RecordBuildMetrics", func(t *testing.T) {
		buildMetrics := BuildMetrics{
			Duration:        5 * time.Second,
			LinesOfCode:     1500,
			PackagesBuilt:   10,
			TestsRun:        25,
			TestsPassed:     23,
			TestsFailed:     2,
			Coverage:        85.5,
			BinarySize:      2048000,
			DependencyCount: 15,
			Warnings:        3,
			Errors:          0,
			Resources: ResourceMetrics{
				CPUUsage:    75.5,
				MemoryUsage: 512000000,
				DiskUsage:   1024000,
				NetworkIO:   10240,
				FileHandles: 50,
				Goroutines:  25,
			},
			Timestamp: time.Now(),
			Success:   true,
			Tags:      map[string]string{"platform": "linux"},
		}

		err := collector.RecordBuildMetrics(buildMetrics)
		require.NoError(t, err)

		// Check that multiple metrics were recorded
		buildMetricNames := []string{
			"build_duration",
			"build_lines_of_code",
			"build_packages_built",
			"build_tests_run",
			"build_coverage",
			"build_binary_size",
			"build_cpu_usage",
			"build_memory_usage",
		}

		for _, name := range buildMetricNames {
			found := false
			for _, metric := range collector.metrics {
				if metric.Name == name {
					found = true
					assert.Equal(t, buildMetrics.Tags, metric.Tags)
					assert.True(t, metric.Success)
					break
				}
			}
			assert.True(t, found, "Expected to find metric: %s", name)
		}
	})

	t.Run("disabled collector doesn't record", func(t *testing.T) {
		disabledConfig := config
		disabledConfig.Enabled = false
		disabledCollector := NewMetricsCollector(disabledConfig)

		err := disabledCollector.RecordCounter("disabled_counter", 1.0, nil)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(disabledCollector.metrics))
	})
}

func TestResourceMetrics(t *testing.T) {
	tempDir := t.TempDir()
	config := MetricsConfig{
		Enabled:     true,
		StoragePath: filepath.Join(tempDir, "metrics"),
	}
	collector := NewMetricsCollector(config)

	t.Run("GetCurrentResourceMetrics", func(t *testing.T) {
		metrics := collector.GetCurrentResourceMetrics()

		// Basic sanity checks - these values depend on the system
		assert.True(t, metrics.MemoryUsage >= 0)
		assert.True(t, metrics.Goroutines > 0) // There should be at least 1 goroutine running
		assert.True(t, metrics.CPUUsage >= 0)
		assert.True(t, metrics.DiskUsage >= 0)
		assert.True(t, metrics.NetworkIO >= 0)
		assert.True(t, metrics.FileHandles >= 0)
	})
}

func TestMetricsQuery(t *testing.T) {
	tempDir := t.TempDir()
	config := MetricsConfig{
		Enabled:     true,
		StoragePath: filepath.Join(tempDir, "metrics"),
	}
	collector := NewMetricsCollector(config)

	// Record some test metrics
	now := time.Now()
	for i := 0; i < 5; i++ {
		metric := &Metric{
			Name:      "test_metric",
			Type:      MetricTypeCounter,
			Value:     float64(i),
			Unit:      "count",
			Timestamp: now.Add(time.Duration(i) * time.Minute),
			Success:   true,
		}
		collector.RecordMetric(metric)
	}

	t.Run("QueryMetrics with disabled collector", func(t *testing.T) {
		disabledConfig := config
		disabledConfig.Enabled = false
		disabledCollector := NewMetricsCollector(disabledConfig)

		query := MetricsQuery{
			StartTime: now.Add(-time.Hour),
			EndTime:   now.Add(time.Hour),
		}

		_, err := disabledCollector.QueryMetrics(query)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metrics collection is disabled")
	})
}

func TestPerformanceReport(t *testing.T) {
	tempDir := t.TempDir()
	config := MetricsConfig{
		Enabled:     true,
		StoragePath: filepath.Join(tempDir, "metrics"),
	}

	t.Run("GenerateReport with disabled collector", func(t *testing.T) {
		disabledConfig := config
		disabledConfig.Enabled = false
		disabledCollector := NewMetricsCollector(disabledConfig)

		_, err := disabledCollector.GenerateReport("day")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metrics collection is disabled")
	})

	t.Run("GenerateReport handles different periods", func(t *testing.T) {
		// This collector has storage properly initialized, so test should work
		storageCollector := NewMetricsCollector(MetricsConfig{
			Enabled:     true,
			StoragePath: filepath.Join(tempDir, "test_storage"),
		})

		periods := []string{"day", "week", "month", "invalid"}

		for _, period := range periods {
			t.Run(period, func(t *testing.T) {
				_, err := storageCollector.GenerateReport(period)
				// Should work since storage is properly initialized
				assert.NoError(t, err)
			})
		}
	})
}

func TestCleanup(t *testing.T) {
	tempDir := t.TempDir()
	config := MetricsConfig{
		Enabled:     true,
		StoragePath: filepath.Join(tempDir, "metrics"),
	}
	collector := NewMetricsCollector(config)

	t.Run("Cleanup with disabled collector", func(t *testing.T) {
		disabledConfig := config
		disabledConfig.Enabled = false
		disabledCollector := NewMetricsCollector(disabledConfig)

		err := disabledCollector.Cleanup()
		assert.NoError(t, err) // Should be no-op
	})

	t.Run("Cleanup with no storage", func(t *testing.T) {
		// Collector with enabled config but no storage initialized
		err := collector.Cleanup()
		assert.NoError(t, err) // Should handle nil storage gracefully
	})
}

func TestMetricTypes(t *testing.T) {
	// Test that all metric types are defined correctly
	types := []MetricType{
		MetricTypeCounter,
		MetricTypeGauge,
		MetricTypeHistogram,
		MetricTypeTimer,
		MetricTypeSummary,
	}

	expectedValues := []string{
		"counter",
		"gauge",
		"histogram",
		"timer",
		"summary",
	}

	for i, metricType := range types {
		assert.Equal(t, expectedValues[i], string(metricType))
	}
}

func TestJSONStorage(t *testing.T) {
	tempDir := t.TempDir()
	storagePath := filepath.Join(tempDir, "test_storage")

	t.Run("NewJSONStorage creates storage", func(t *testing.T) {
		storage, err := NewJSONStorage(storagePath)
		require.NoError(t, err)
		assert.NotNil(t, storage)
		assert.Equal(t, storagePath, storage.storagePath)

		// Check that directory was created
		assert.DirExists(t, storagePath)
	})

	t.Run("Store and Query metrics", func(t *testing.T) {
		// Create unique storage path for this test
		testStoragePath := filepath.Join(tempDir, "test1")
		storage, err := NewJSONStorage(testStoragePath)
		require.NoError(t, err)

		// Store test metric
		now := time.Now()
		metric := &Metric{
			Name:      "test_storage_metric",
			Type:      MetricTypeCounter,
			Value:     100.0,
			Unit:      "count",
			Timestamp: now,
			Tags:      map[string]string{"test": "true"},
			Success:   true,
		}

		err = storage.Store(metric)
		require.NoError(t, err)

		// Query metrics
		query := MetricsQuery{
			StartTime: now.Add(-time.Hour),
			EndTime:   now.Add(time.Hour),
		}

		metrics, err := storage.Query(query)
		require.NoError(t, err)
		require.Len(t, metrics, 1)

		retrieved := metrics[0]
		assert.Equal(t, metric.Name, retrieved.Name)
		assert.Equal(t, metric.Type, retrieved.Type)
		assert.Equal(t, metric.Value, retrieved.Value)
		assert.Equal(t, metric.Tags, retrieved.Tags)
	})

	t.Run("Query with filters", func(t *testing.T) {
		// Create unique storage path for this test
		testStoragePath := filepath.Join(tempDir, "test2")
		storage, err := NewJSONStorage(testStoragePath)
		require.NoError(t, err)

		// Store multiple metrics
		now := time.Now()
		metrics := []*Metric{
			{
				Name:      "metric1",
				Type:      MetricTypeCounter,
				Value:     1.0,
				Timestamp: now,
				Tags:      map[string]string{"env": "test"},
			},
			{
				Name:      "metric2",
				Type:      MetricTypeGauge,
				Value:     2.0,
				Timestamp: now,
				Tags:      map[string]string{"env": "prod"},
			},
		}

		for _, metric := range metrics {
			err = storage.Store(metric)
			require.NoError(t, err)
		}

		// Query with name filter
		query := MetricsQuery{
			StartTime: now.Add(-time.Hour),
			EndTime:   now.Add(time.Hour),
			Names:     []string{"metric1"},
		}

		results, err := storage.Query(query)
		require.NoError(t, err)
		assert.True(t, len(results) >= 1)

		// Query with tags filter
		query = MetricsQuery{
			StartTime: now.Add(-time.Hour),
			EndTime:   now.Add(time.Hour),
			Tags:      map[string]string{"env": "test"},
		}

		results, err = storage.Query(query)
		require.NoError(t, err)
		assert.True(t, len(results) >= 1)
	})

	t.Run("Query with limit", func(t *testing.T) {
		// Create unique storage path for this test
		testStoragePath := filepath.Join(tempDir, "test3")
		storage, err := NewJSONStorage(testStoragePath)
		require.NoError(t, err)

		query := MetricsQuery{
			StartTime: time.Now().Add(-24 * time.Hour),
			EndTime:   time.Now().Add(time.Hour),
			Limit:     1,
		}

		results, err := storage.Query(query)
		require.NoError(t, err)
		assert.True(t, len(results) <= 1)
	})

	t.Run("Aggregate metrics", func(t *testing.T) {
		// Create unique storage path for this test
		testStoragePath := filepath.Join(tempDir, "test4")
		storage, err := NewJSONStorage(testStoragePath)
		require.NoError(t, err)

		query := MetricsQuery{
			StartTime: time.Now().Add(-24 * time.Hour),
			EndTime:   time.Now().Add(time.Hour),
		}

		aggregated, err := storage.Aggregate(query)
		require.NoError(t, err)
		assert.NotNil(t, aggregated)
		assert.True(t, aggregated.Count >= 0)
	})

	t.Run("Cleanup old files", func(t *testing.T) {
		// Create unique storage path for this test
		testStoragePath := filepath.Join(tempDir, "test5")
		storage, err := NewJSONStorage(testStoragePath)
		require.NoError(t, err)

		// Create an old metrics file
		oldDate := time.Now().AddDate(0, 0, -10)
		oldFilename := "metrics_" + oldDate.Format("2006-01-02") + ".json"
		oldFilePath := filepath.Join(testStoragePath, oldFilename)

		err = os.WriteFile(oldFilePath, []byte("[]"), 0o644)
		require.NoError(t, err)

		// Cleanup with 5 day retention
		err = storage.Cleanup(5)
		require.NoError(t, err)

		// File should be removed
		assert.NoFileExists(t, oldFilePath)
	})
}

func TestPackageLevelMetricsFunctions(t *testing.T) {
	// Test global collector functions
	t.Run("global collector functions work", func(t *testing.T) {
		// These functions use the global collector which may or may not be enabled
		// We'll test that they don't panic

		assert.NotPanics(t, func() {
			timer := StartTimer("test_global", map[string]string{"global": "true"})
			timer.Stop()
		})

		assert.NotPanics(t, func() {
			RecordCounter("global_counter", 1.0, nil)
		})

		assert.NotPanics(t, func() {
			RecordGauge("global_gauge", 50.0, "percent", nil)
		})

		assert.NotPanics(t, func() {
			GetCurrentResourceMetrics()
		})

		assert.NotPanics(t, func() {
			CleanupMetrics()
		})
	})

	t.Run("GetMetricsCollector returns singleton", func(t *testing.T) {
		collector1 := GetMetricsCollector()
		collector2 := GetMetricsCollector()

		// Should return the same instance
		assert.Equal(t, collector1, collector2)
	})
}

func TestHelperFunctions(t *testing.T) {
	t.Run("percentile calculation", func(t *testing.T) {
		values := []float64{1.0, 2.0, 3.0, 4.0, 5.0, 6.0, 7.0, 8.0, 9.0, 10.0}

		p50 := percentile(values, 50)
		assert.InDelta(t, 5.5, p50, 0.1) // 50th percentile should be around 5.5

		p95 := percentile(values, 95)
		assert.InDelta(t, 9.5, p95, 0.1) // 95th percentile should be around 9.5

		// Edge cases
		assert.Equal(t, 0.0, percentile([]float64{}, 50))
		assert.Equal(t, 1.0, percentile([]float64{1.0}, 50))
	})

	t.Run("average calculation", func(t *testing.T) {
		values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
		avg := average(values)
		assert.Equal(t, 3.0, avg)

		// Edge cases
		assert.Equal(t, 0.0, average([]float64{}))
		assert.Equal(t, 5.0, average([]float64{5.0}))
	})

	t.Run("filterMetrics", func(t *testing.T) {
		metrics := []*Metric{
			{Name: "build_duration"},
			{Name: "test_count"},
			{Name: "build_success"},
			{Name: "deploy_time"},
		}

		buildMetrics := filterMetrics(metrics, "build_")
		assert.Len(t, buildMetrics, 2)
		assert.Equal(t, "build_duration", buildMetrics[0].Name)
		assert.Equal(t, "build_success", buildMetrics[1].Name)

		testMetrics := filterMetrics(metrics, "test_")
		assert.Len(t, testMetrics, 1)
		assert.Equal(t, "test_count", testMetrics[0].Name)

		noMatch := filterMetrics(metrics, "none_")
		assert.Len(t, noMatch, 0)
	})
}

// Benchmark tests
func BenchmarkMetricsCollector_RecordCounter(b *testing.B) {
	tempDir := b.TempDir()
	config := MetricsConfig{
		Enabled:     true,
		StoragePath: filepath.Join(tempDir, "metrics"),
	}
	collector := NewMetricsCollector(config)
	tags := map[string]string{"bench": "true"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		collector.RecordCounter("bench_counter", float64(i), tags)
	}
}

func BenchmarkPerformanceTimer_StartStop(b *testing.B) {
	tempDir := b.TempDir()
	config := MetricsConfig{
		Enabled:     true,
		StoragePath: filepath.Join(tempDir, "metrics"),
	}
	collector := NewMetricsCollector(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		timer := collector.StartTimer("bench_timer", nil)
		timer.Stop()
	}
}

func BenchmarkJSONStorage_Store(b *testing.B) {
	tempDir := b.TempDir()
	storage, err := NewJSONStorage(filepath.Join(tempDir, "bench_storage"))
	if err != nil {
		b.Fatal(err)
	}

	metric := &Metric{
		Name:      "bench_metric",
		Type:      MetricTypeCounter,
		Value:     1.0,
		Unit:      "count",
		Timestamp: time.Now(),
		Success:   true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.Store(metric)
	}
}
