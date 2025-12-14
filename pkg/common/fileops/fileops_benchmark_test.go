package fileops

import (
	"fmt"
	"path/filepath"
	"testing"
)

// BenchmarkFileOperations benchmarks basic file operations at various sizes
func BenchmarkFileOperations(b *testing.B) {
	ops := NewDefaultFileOperator()

	sizes := []struct {
		name string
		size int
	}{
		{"1B", 1},
		{"1KB", 1024},
		{"64KB", 64 * 1024},
		{"1MB", 1024 * 1024},
	}

	for _, s := range sizes {
		data := make([]byte, s.size)
		for i := range data {
			data[i] = byte(i % 256)
		}

		b.Run(fmt.Sprintf("WriteFile_%s", s.name), func(b *testing.B) {
			tmpDir := b.TempDir()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				testFile := filepath.Join(tmpDir, fmt.Sprintf("bench_%d.txt", i))
				if err := ops.WriteFile(testFile, data, 0o644); err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("ReadFile_%s", s.name), func(b *testing.B) {
			tmpDir := b.TempDir()
			testFile := filepath.Join(tmpDir, "bench_read.txt")
			if err := ops.WriteFile(testFile, data, 0o644); err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := ops.ReadFile(testFile); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkAtomicWrite benchmarks atomic write operations
func BenchmarkAtomicWrite(b *testing.B) {
	safeOps := NewDefaultSafeFileOperator()

	sizes := []struct {
		name string
		size int
	}{
		{"1B", 1},
		{"1KB", 1024},
		{"64KB", 64 * 1024},
		{"1MB", 1024 * 1024},
	}

	for _, s := range sizes {
		data := make([]byte, s.size)
		for i := range data {
			data[i] = byte(i % 256)
		}

		b.Run(fmt.Sprintf("AtomicWrite_%s", s.name), func(b *testing.B) {
			tmpDir := b.TempDir()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				testFile := filepath.Join(tmpDir, fmt.Sprintf("atomic_%d.txt", i))
				if err := safeOps.WriteFileAtomic(testFile, data, 0o644); err != nil {
					b.Fatal(err)
				}
			}
		})

		b.Run(fmt.Sprintf("AtomicOverwrite_%s", s.name), func(b *testing.B) {
			tmpDir := b.TempDir()
			testFile := filepath.Join(tmpDir, "overwrite.txt")
			// Create initial file
			if err := safeOps.WriteFileAtomic(testFile, data, 0o644); err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := safeOps.WriteFileAtomic(testFile, data, 0o644); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkCopy benchmarks file copy operations
func BenchmarkCopy(b *testing.B) {
	ops := NewDefaultFileOperator()

	sizes := []struct {
		name string
		size int
	}{
		{"1KB", 1024},
		{"64KB", 64 * 1024},
		{"1MB", 1024 * 1024},
	}

	for _, s := range sizes {
		data := make([]byte, s.size)
		for i := range data {
			data[i] = byte(i % 256)
		}

		b.Run(fmt.Sprintf("Copy_%s", s.name), func(b *testing.B) {
			tmpDir := b.TempDir()
			srcFile := filepath.Join(tmpDir, "source.bin")
			if err := ops.WriteFile(srcFile, data, 0o644); err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				dstFile := filepath.Join(tmpDir, fmt.Sprintf("dest_%d.bin", i))
				if err := ops.Copy(srcFile, dstFile); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// BenchmarkJSONMarshal benchmarks JSON marshaling operations
func BenchmarkJSONMarshal(b *testing.B) {
	fileOps := NewDefaultFileOperator()
	jsonOps := NewDefaultJSONOperator(fileOps)

	type TestStruct struct {
		Name     string            `json:"name"`
		Value    int               `json:"value"`
		Tags     []string          `json:"tags"`
		Settings map[string]string `json:"settings"`
	}

	smallData := TestStruct{
		Name:     "small",
		Value:    42,
		Tags:     []string{"a", "b"},
		Settings: map[string]string{"key": "value"},
	}

	largeData := TestStruct{
		Name:  "large",
		Value: 12345,
		Tags:  make([]string, 100),
		Settings: func() map[string]string {
			m := make(map[string]string)
			for i := 0; i < 100; i++ {
				m[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
			}
			return m
		}(),
	}
	for i := range largeData.Tags {
		largeData.Tags[i] = fmt.Sprintf("tag_%d", i)
	}

	b.Run("Marshal_Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := jsonOps.Marshal(smallData); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Marshal_Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := jsonOps.Marshal(largeData); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MarshalIndent_Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := jsonOps.MarshalIndent(smallData, "", "  "); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MarshalIndent_Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := jsonOps.MarshalIndent(largeData, "", "  "); err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark unmarshal
	smallJSON, err := jsonOps.Marshal(smallData)
	if err != nil {
		b.Fatal(err)
	}
	largeJSON, err := jsonOps.Marshal(largeData)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("Unmarshal_Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result TestStruct
			if err := jsonOps.Unmarshal(smallJSON, &result); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Unmarshal_Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result TestStruct
			if err := jsonOps.Unmarshal(largeJSON, &result); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkYAMLMarshal benchmarks YAML marshaling operations
func BenchmarkYAMLMarshal(b *testing.B) {
	fileOps := NewDefaultFileOperator()
	yamlOps := NewDefaultYAMLOperator(fileOps)

	type TestStruct struct {
		Name     string            `yaml:"name"`
		Value    int               `yaml:"value"`
		Tags     []string          `yaml:"tags"`
		Settings map[string]string `yaml:"settings"`
	}

	smallData := TestStruct{
		Name:     "small",
		Value:    42,
		Tags:     []string{"a", "b"},
		Settings: map[string]string{"key": "value"},
	}

	largeData := TestStruct{
		Name:  "large",
		Value: 12345,
		Tags:  make([]string, 100),
		Settings: func() map[string]string {
			m := make(map[string]string)
			for i := 0; i < 100; i++ {
				m[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
			}
			return m
		}(),
	}
	for i := range largeData.Tags {
		largeData.Tags[i] = fmt.Sprintf("tag_%d", i)
	}

	b.Run("Marshal_Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := yamlOps.Marshal(smallData); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Marshal_Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := yamlOps.Marshal(largeData); err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark unmarshal
	smallYAML, err := yamlOps.Marshal(smallData)
	if err != nil {
		b.Fatal(err)
	}
	largeYAML, err := yamlOps.Marshal(largeData)
	if err != nil {
		b.Fatal(err)
	}

	b.Run("Unmarshal_Small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result TestStruct
			if err := yamlOps.Unmarshal(smallYAML, &result); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Unmarshal_Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var result TestStruct
			if err := yamlOps.Unmarshal(largeYAML, &result); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkWriteJSONSafe benchmarks the WriteJSONSafe facade method
func BenchmarkWriteJSONSafe(b *testing.B) {
	ops := New()

	data := map[string]interface{}{
		"name":    "benchmark",
		"value":   42,
		"tags":    []string{"a", "b", "c"},
		"enabled": true,
	}

	b.Run("WriteJSONSafe", func(b *testing.B) {
		tmpDir := b.TempDir()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			testFile := filepath.Join(tmpDir, fmt.Sprintf("safe_%d.json", i))
			if err := ops.WriteJSONSafe(testFile, data); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("WriteJSONSafe_WithDirCreation", func(b *testing.B) {
		tmpDir := b.TempDir()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			testFile := filepath.Join(tmpDir, fmt.Sprintf("dir_%d", i), "safe.json")
			if err := ops.WriteJSONSafe(testFile, data); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkWriteYAMLSafe benchmarks the WriteYAMLSafe facade method
func BenchmarkWriteYAMLSafe(b *testing.B) {
	ops := New()

	data := map[string]interface{}{
		"name":    "benchmark",
		"value":   42,
		"tags":    []string{"a", "b", "c"},
		"enabled": true,
	}

	b.Run("WriteYAMLSafe", func(b *testing.B) {
		tmpDir := b.TempDir()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			testFile := filepath.Join(tmpDir, fmt.Sprintf("safe_%d.yaml", i))
			if err := ops.WriteYAMLSafe(testFile, data); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkLoadConfig benchmarks config loading
func BenchmarkLoadConfig(b *testing.B) {
	ops := New()

	type Config struct {
		Name     string            `json:"name" yaml:"name"`
		Version  int               `json:"version" yaml:"version"`
		Settings map[string]string `json:"settings" yaml:"settings"`
	}

	config := Config{
		Name:    "benchmark",
		Version: 1,
		Settings: map[string]string{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		},
	}

	b.Run("LoadConfig_JSON", func(b *testing.B) {
		tmpDir := b.TempDir()
		configFile := filepath.Join(tmpDir, "config.json")
		if err := ops.WriteJSONSafe(configFile, config); err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result Config
			if _, err := ops.LoadConfig([]string{configFile}, &result); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("LoadConfig_YAML", func(b *testing.B) {
		tmpDir := b.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")
		if err := ops.WriteYAMLSafe(configFile, config); err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result Config
			if _, err := ops.LoadConfig([]string{configFile}, &result); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("LoadConfig_Fallback", func(b *testing.B) {
		tmpDir := b.TempDir()
		configFile := filepath.Join(tmpDir, "config.yaml")
		if err := ops.WriteYAMLSafe(configFile, config); err != nil {
			b.Fatal(err)
		}
		nonexistent := filepath.Join(tmpDir, "nonexistent.yaml")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var result Config
			if _, err := ops.LoadConfig([]string{nonexistent, configFile}, &result); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkExists benchmarks the Exists check
func BenchmarkExists(b *testing.B) {
	ops := NewDefaultFileOperator()

	b.Run("Exists_True", func(b *testing.B) {
		tmpDir := b.TempDir()
		testFile := filepath.Join(tmpDir, "exists.txt")
		if err := ops.WriteFile(testFile, []byte("test"), 0o644); err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ops.Exists(testFile)
		}
	})

	b.Run("Exists_False", func(b *testing.B) {
		tmpDir := b.TempDir()
		testFile := filepath.Join(tmpDir, "nonexistent.txt")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ops.Exists(testFile)
		}
	})
}

// BenchmarkStat benchmarks the Stat operation
func BenchmarkStat(b *testing.B) {
	ops := NewDefaultFileOperator()

	b.Run("Stat_File", func(b *testing.B) {
		tmpDir := b.TempDir()
		testFile := filepath.Join(tmpDir, "stat.txt")
		if err := ops.WriteFile(testFile, []byte("test content"), 0o644); err != nil {
			b.Fatal(err)
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := ops.Stat(testFile); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("Stat_Dir", func(b *testing.B) {
		tmpDir := b.TempDir()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := ops.Stat(tmpDir); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkMkdirAll benchmarks directory creation
func BenchmarkMkdirAll(b *testing.B) {
	ops := NewDefaultFileOperator()

	b.Run("MkdirAll_Single", func(b *testing.B) {
		tmpDir := b.TempDir()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dir := filepath.Join(tmpDir, fmt.Sprintf("dir_%d", i))
			if err := ops.MkdirAll(dir, 0o755); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("MkdirAll_Nested", func(b *testing.B) {
		tmpDir := b.TempDir()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			dir := filepath.Join(tmpDir, fmt.Sprintf("a_%d", i), "b", "c", "d")
			if err := ops.MkdirAll(dir, 0o755); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkReadDir benchmarks directory listing
func BenchmarkReadDir(b *testing.B) {
	ops := NewDefaultFileOperator()

	fileCounts := []int{10, 100, 1000}

	for _, count := range fileCounts {
		b.Run(fmt.Sprintf("ReadDir_%d_files", count), func(b *testing.B) {
			tmpDir := b.TempDir()
			// Create files
			for i := 0; i < count; i++ {
				f := filepath.Join(tmpDir, fmt.Sprintf("file_%04d.txt", i))
				if err := ops.WriteFile(f, []byte("x"), 0o644); err != nil {
					b.Fatal(err)
				}
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := ops.ReadDir(tmpDir); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
