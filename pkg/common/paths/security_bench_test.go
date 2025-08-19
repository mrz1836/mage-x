package paths

import (
	"strings"
	"testing"
)

// BenchmarkSecurityValidationPerformance benchmarks various security validation scenarios
func BenchmarkSecurityValidationPerformance(b *testing.B) {
	// Test data representing different types of security concerns
	testCases := []struct {
		name string
		path string
	}{
		{"safe_short_path", "safe/file.txt"},
		{"safe_long_path", strings.Repeat("safe/", 50) + "file.txt"},
		{"traversal_simple", "../../../etc/passwd"},
		{"traversal_complex", "../../../" + strings.Repeat("../", 20) + "etc/passwd"},
		{"traversal_mixed", "safe/../unsafe/../../../etc/passwd"},
		{"url_encoded", "%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd"},
		{"unicode_encoded", "\u002e\u002e\u002f\u002e\u002e\u002f\u002e\u002e\u002fetc\u002fpasswd"},
		{"null_injection", "safe.txt\x00../../../etc/passwd"},
		{"control_chars", "safe.txt\n../../../etc/passwd"},
		{"windows_unc", "\\\\server\\share\\file.txt"},
		{"windows_drive", "C:\\Windows\\System32\\config\\SAM"},
		{"proc_filesystem", "/proc/self/fd/0"},
		{"dev_filesystem", "/dev/random"},
		{"very_long_path", strings.Repeat("A", 2000)},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			pb := NewPathBuilder(tc.path)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = pb.IsSafe()
			}
		})
	}
}

// BenchmarkPathValidatorRules benchmarks different validation rules
func BenchmarkPathValidatorRules(b *testing.B) {
	paths := []string{
		"/absolute/safe/path.txt",
		"relative/safe/path.txt",
		"../../../etc/passwd",
		"file.txt",
		"file_with_no_extension",
		".hidden_file",
		strings.Repeat("long/", 100) + "file.txt",
	}

	ruleTests := []struct {
		name      string
		setupFunc func() PathValidator
	}{
		{
			name: "require_absolute",
			setupFunc: func() PathValidator {
				return NewPathValidator().RequireAbsolute()
			},
		},
		{
			name: "require_relative",
			setupFunc: func() PathValidator {
				return NewPathValidator().RequireRelative()
			},
		},
		{
			name: "require_extension",
			setupFunc: func() PathValidator {
				return NewPathValidator().RequireExtension("txt", "go", "md")
			},
		},
		{
			name: "require_max_length",
			setupFunc: func() PathValidator {
				return NewPathValidator().RequireMaxLength(1000)
			},
		},
		{
			name: "forbid_pattern",
			setupFunc: func() PathValidator {
				return NewPathValidator().ForbidPattern(`\.\.`)
			},
		},
		{
			name: "require_pattern",
			setupFunc: func() PathValidator {
				return NewPathValidator().RequirePattern(`^[a-zA-Z0-9._/-]+$`)
			},
		},
		{
			name: "multiple_rules",
			setupFunc: func() PathValidator {
				return NewPathValidator().
					RequireAbsolute().
					RequireExtension("txt", "go").
					RequireMaxLength(500).
					ForbidPattern(`\.\.`)
			},
		},
	}

	for _, ruleTest := range ruleTests {
		b.Run(ruleTest.name, func(b *testing.B) {
			validator := ruleTest.setupFunc()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				for _, path := range paths {
					_ = validator.Validate(path)
				}
			}
		})
	}
}

// BenchmarkPathCleaning benchmarks path cleaning operations
func BenchmarkPathCleaning(b *testing.B) {
	testPaths := []struct {
		name string
		path string
	}{
		{"simple_clean", "./path/to/file.txt"},
		{"complex_clean", "./path/../to/./file/../other.txt"},
		{"traversal_clean", "../../../etc/passwd"},
		{"multiple_slashes", "path//to///file.txt"},
		{"mixed_separators", "path\\to/file.txt"},
		{"long_path_clean", strings.Repeat("./", 500) + "file.txt"},
		{"deep_traversal", strings.Repeat("../", 100) + "etc/passwd"},
	}

	for _, tp := range testPaths {
		b.Run(tp.name, func(b *testing.B) {
			pb := NewPathBuilder(tp.path)
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = pb.Clean()
			}
		})
	}
}

// BenchmarkSecurityVsPerformance compares different security approaches
func BenchmarkSecurityVsPerformance(b *testing.B) {
	dangerousPaths := []string{
		"../../../etc/passwd",
		"..\\..\\..\\windows\\system32",
		"%2e%2e%2f%2e%2e%2f%2e%2e%2f",
		"file.txt\x00../../../etc/passwd",
		"/proc/self/fd/0",
		strings.Repeat("../", 50) + "etc/passwd",
	}

	b.Run("is_safe_check", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, path := range dangerousPaths {
				pb := NewPathBuilder(path)
				_ = pb.IsSafe()
			}
		}
	})

	b.Run("validator_check", func(b *testing.B) {
		validator := NewPathValidator().ForbidPattern(`\.\.`)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for _, path := range dangerousPaths {
				_ = validator.Validate(path)
			}
		}
	})

	b.Run("combined_check", func(b *testing.B) {
		validator := NewPathValidator().
			RequireMaxLength(1024).
			ForbidPattern(`\.\.`).
			RequirePattern(`^[^\x00\n\r\t]*$`) // No control characters
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for _, path := range dangerousPaths {
				pb := NewPathBuilder(path)
				_ = pb.IsSafe()
				_ = validator.ValidatePath(pb)
			}
		}
	})
}

// BenchmarkPathOperationsSecurity benchmarks various path operations under security load
func BenchmarkPathOperationsSecurity(b *testing.B) {
	paths := []string{
		"safe/file.txt",
		"../unsafe/file.txt",
		"very/deep/" + strings.Repeat("nested/", 20) + "file.txt",
		strings.Repeat("A", 255) + ".txt",
	}

	operations := []struct {
		name string
		fn   func(PathBuilder) PathBuilder
	}{
		{"clean", func(pb PathBuilder) PathBuilder { return pb.Clean() }},
		{"dir", func(pb PathBuilder) PathBuilder { return pb.Dir() }},
		{"join", func(pb PathBuilder) PathBuilder { return pb.Join("extra", "path") }},
		{"with_ext", func(pb PathBuilder) PathBuilder { return pb.WithExt(".new") }},
		{"without_ext", func(pb PathBuilder) PathBuilder { return pb.WithoutExt() }},
		{"append", func(pb PathBuilder) PathBuilder { return pb.Append("_suffix") }},
		{"prepend", func(pb PathBuilder) PathBuilder { return pb.Prepend("prefix_") }},
		{"clone", func(pb PathBuilder) PathBuilder { return pb.Clone() }},
	}

	for _, op := range operations {
		b.Run(op.name, func(b *testing.B) {
			pbs := make([]PathBuilder, len(paths))
			for i, path := range paths {
				pbs[i] = NewPathBuilder(path)
			}
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				for _, pb := range pbs {
					_ = op.fn(pb)
				}
			}
		})
	}
}

// BenchmarkMemoryEfficiency benchmarks memory usage of security operations
func BenchmarkMemoryEfficiency(b *testing.B) {
	largePaths := make([]string, 1000)
	for i := range largePaths {
		largePaths[i] = "path/to/file" + strings.Repeat("/subdir", i%100) + "/file.txt"
	}

	b.Run("batch_validation", func(b *testing.B) {
		validator := NewPathValidator().
			RequireMaxLength(2048).
			ForbidPattern(`\.\.`).
			RequirePattern(`^[a-zA-Z0-9._/-]+$`)
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			for _, path := range largePaths {
				_ = validator.Validate(path)
			}
		}
	})

	b.Run("batch_safety_check", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, path := range largePaths {
				pb := NewPathBuilder(path)
				_ = pb.IsSafe()
			}
		}
	})

	b.Run("batch_path_operations", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, path := range largePaths {
				pb := NewPathBuilder(path)
				_ = pb.Clean()
				_ = pb.Dir()
				_ = pb.Base()
				_ = pb.Ext()
			}
		}
	})
}
