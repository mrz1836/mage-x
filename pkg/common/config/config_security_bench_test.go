package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

// Internal security validation functions for benchmarks

// Security errors
var (
	ErrInvalidUTF8Bench            = errors.New("argument contains invalid UTF-8")
	ErrDangerousPipePatternBench   = errors.New("potentially dangerous pattern '|' detected")
	ErrDangerousPatternBench       = errors.New("potentially dangerous pattern detected")
	ErrPathTraversalBench          = errors.New("path traversal detected")
	ErrPathContainsNullBench       = errors.New("path contains null byte")
	ErrPathContainsControlBench    = errors.New("path contains control character")
	ErrAbsolutePathNotAllowedBench = errors.New("absolute paths not allowed outside of /tmp")
)

// validateCommandArgSecurityBench validates a command argument for security issues
func validateCommandArgSecurityBench(arg string) error {
	// Check for valid UTF-8
	if !utf8.ValidString(arg) {
		return ErrInvalidUTF8Bench
	}

	// Check for shell injection attempts
	dangerousPatterns := []string{
		"$(",     // Command substitution
		"`",      // Command substitution
		"&&",     // Command chaining
		"||",     // Command chaining
		";",      // Command separator
		">",      // Redirect
		"<",      // Redirect
		"$(echo", // Common injection pattern
		"${IFS}", // Shell variable manipulation
	}

	// Special cases where pipe is dangerous (not in regex or URLs)
	if strings.Contains(arg, "|") {
		// Allow pipe in regex patterns (contains regex metacharacters)
		if !strings.ContainsAny(arg, "^$[]()+*?.{}\\%") {
			// Allow pipe in URLs
			if !strings.HasPrefix(arg, "http://") && !strings.HasPrefix(arg, "https://") {
				return ErrDangerousPipePatternBench
			}
		}
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(arg, pattern) {
			return fmt.Errorf("%w: '%s'", ErrDangerousPatternBench, pattern)
		}
	}

	return nil
}

// validatePathSecurityBench validates a file path for security issues
func validatePathSecurityBench(path string) error {
	// Check for control characters and dangerous patterns first
	if strings.Contains(path, "\x00") {
		return ErrPathContainsNullBench
	}
	if strings.Contains(path, "\n") || strings.Contains(path, "\r") {
		return ErrPathContainsControlBench
	}

	// Check for path traversal BEFORE cleaning (Unix and Windows styles)
	if strings.Contains(path, "../") || strings.Contains(path, "..\\") ||
		strings.HasSuffix(path, "..") || path == ".." {
		return ErrPathTraversalBench
	}

	// Clean the path
	cleaned := filepath.Clean(path)

	// Check for Windows-style paths (which should be rejected on Unix systems)
	if strings.Contains(path, ":") && len(path) > 1 && path[1] == ':' {
		// This looks like a Windows drive path (C:, D:, etc.)
		return ErrAbsolutePathNotAllowedBench
	}

	// Check for UNC paths
	if strings.HasPrefix(path, "\\\\") {
		return ErrAbsolutePathNotAllowedBench
	}

	// Check if path is absolute when it shouldn't be
	if filepath.IsAbs(cleaned) && !strings.HasPrefix(cleaned, "/tmp") {
		// Allow absolute paths only in /tmp for now
		return ErrAbsolutePathNotAllowedBench
	}

	return nil
}

// BenchmarkConfigSecurityValidation benchmarks security validation performance
func BenchmarkConfigSecurityValidation(b *testing.B) {
	testConfigs := []struct {
		name   string
		config map[string]interface{}
	}{
		{
			name: "small_config",
			config: map[string]interface{}{
				"app": map[string]interface{}{
					"name": "test-app",
					"port": 8080,
				},
			},
		},
		{
			name:   "medium_config",
			config: generateTestConfig(50),
		},
		{
			name:   "large_config",
			config: generateTestConfig(500),
		},
		{
			name: "config_with_secrets",
			config: map[string]interface{}{
				"database": map[string]interface{}{
					"password": "secret123",
					"api_key":  "key-abc123",
				},
				"tokens": []string{
					"github_token_123",
					"gitlab_token_456",
				},
			},
		},
		{
			name: "config_with_commands",
			config: map[string]interface{}{
				"scripts": []string{
					"echo hello",
					"ls -la",
					"cat /etc/passwd",
					"$(whoami)",
					"`id`",
				},
				"commands": map[string]interface{}{
					"build": "go build -o app",
					"test":  "go test ./...",
				},
			},
		},
	}

	for _, tc := range testConfigs {
		b.Run(tc.name, func(b *testing.B) {
			loader := NewDefaultConfigLoader()
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				// Benchmark basic validation - ignore errors for performance measurement
				_ = loader.Validate(tc.config) //nolint:errcheck // performance benchmark ignores validation errors
			}
		})
	}
}

// BenchmarkConfigSecurityPathValidation benchmarks path validation performance
func BenchmarkConfigSecurityPathValidation(b *testing.B) {
	testPaths := []struct {
		name  string
		paths []string
	}{
		{
			name: "safe_paths",
			paths: []string{
				"config.yaml",
				"./config.yaml",
				"app/config.yaml",
				"/tmp/config.yaml",
			},
		},
		{
			name: "dangerous_paths",
			paths: []string{
				"../../../etc/passwd",
				"..\\..\\..\\windows\\system32\\config\\sam",
				"/etc/passwd",
				"C:\\Windows\\System32\\config\\SAM",
				"config\x00.yaml",
			},
		},
		{
			name: "long_paths",
			paths: []string{
				strings.Repeat("abcd/", 25) + "config.yaml", // Use different pattern to avoid recursive 'a' directories
				strings.Repeat("../", 100) + "etc/passwd",
				strings.Repeat("x", 1000) + ".yaml",
			},
		},
	}

	for _, tc := range testPaths {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				for _, path := range tc.paths {
					_ = validatePathSecurityBench(path) //nolint:errcheck // benchmark focuses on performance not error handling
				}
			}
		})
	}
}

// BenchmarkConfigSecurityEnvironmentFiltering benchmarks environment variable filtering
func BenchmarkConfigSecurityEnvironmentFiltering(b *testing.B) {
	// Create test environment with mix of safe and sensitive variables
	testEnv := []string{
		"HOME=/home/user",
		"PATH=/usr/bin:/bin",
		"USER=testuser",
		"LANG=en_US.UTF-8",
		"AWS_SECRET_ACCESS_KEY=secret123",
		"GITHUB_TOKEN=ghp_token123",
		"API_KEY=key123",
		"DATABASE_PASSWORD=password123",
		"PRIVATE_KEY=-----BEGIN PRIVATE KEY-----",
		"DEBUG=true",
		"PORT=8080",
		"APP_NAME=myapp",
	}

	// Extend with more variables for performance testing
	for i := 0; i < 100; i++ {
		testEnv = append(testEnv, fmt.Sprintf("VAR_%d=value_%d", i, i))
		testEnv = append(testEnv, fmt.Sprintf("SECRET_%d=secret_%d", i, i))
	}

	// Simulate SecureExecutor behavior for benchmarking
	// executor := security.NewSecureExecutor()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		// Use reflection to access the private filterEnvironment method
		// Since we can't access it directly, we'll simulate the filtering logic
		filteredCount := 0
		for _, env := range testEnv {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				varName := strings.ToUpper(parts[0])
				// Simulate filtering logic
				sensitivePrefixes := []string{
					"AWS_SECRET", "GITHUB_TOKEN", "API_KEY",
					"DATABASE_PASSWORD", "PRIVATE_KEY", "SECRET",
				}
				for _, prefix := range sensitivePrefixes {
					if strings.HasPrefix(varName, prefix) {
						filteredCount++
						break
					}
				}
			}
		}
		// Use filteredCount to prevent optimization
		_ = filteredCount
	}
}

// BenchmarkConfigSecurityCommandValidation benchmarks command argument validation
func BenchmarkConfigSecurityCommandValidation(b *testing.B) {
	testCommands := []struct {
		name     string
		commands []string
	}{
		{
			name: "safe_commands",
			commands: []string{
				"go build",
				"npm install",
				"docker build -t app .",
				"echo hello world",
				"ls -la /home/user",
			},
		},
		{
			name: "dangerous_commands",
			commands: []string{
				"$(whoami)",
				"`id`",
				"rm -rf / && echo done",
				"cat /etc/passwd | grep root",
				"echo $HOME > /tmp/leak",
				"command; malicious_command",
				"value\x00null",
			},
		},
		{
			name:     "mixed_commands",
			commands: generateMixedCommands(100),
		},
	}

	for _, tc := range testCommands {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				for _, cmd := range tc.commands {
					_ = validateCommandArgSecurityBench(cmd) //nolint:errcheck // benchmark focuses on performance not error handling
				}
			}
		})
	}
}

// BenchmarkConfigSecurityConfigLoad benchmarks secure configuration loading
func BenchmarkConfigSecurityConfigLoad(b *testing.B) {
	tmpDir := b.TempDir()

	// Create test configuration files
	configs := []struct {
		name    string
		format  string
		content string
		size    string
	}{
		{
			name:    "small_yaml",
			format:  "yaml",
			content: "app:\n  name: test\n  port: 8080",
			size:    "small",
		},
		{
			name:    "small_json",
			format:  "json",
			content: `{"app": {"name": "test", "port": 8080}}`,
			size:    "small",
		},
		{
			name:    "medium_yaml",
			format:  "yaml",
			content: generateYAMLConfig(100),
			size:    "medium",
		},
		{
			name:    "medium_json",
			format:  "json",
			content: generateJSONConfig(100),
			size:    "medium",
		},
		{
			name:    "large_yaml",
			format:  "yaml",
			content: generateYAMLConfig(1000),
			size:    "large",
		},
		{
			name:    "large_json",
			format:  "json",
			content: generateJSONConfig(1000),
			size:    "large",
		},
	}

	// Write test files
	for _, cfg := range configs {
		configPath := filepath.Join(tmpDir, cfg.name+"."+cfg.format)
		err := os.WriteFile(configPath, []byte(cfg.content), 0o600)
		if err != nil {
			b.Fatalf("Failed to write test config %s: %v", cfg.name, err)
		}
	}

	loader := NewDefaultConfigLoader()

	for _, cfg := range configs {
		b.Run(cfg.name, func(b *testing.B) {
			configPath := filepath.Join(tmpDir, cfg.name+"."+cfg.format)
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				var config map[string]interface{}
				err := loader.LoadFrom(configPath, &config)
				if err != nil {
					b.Fatalf("Failed to load config: %v", err)
				}
				// Basic security validation
				if err := loader.Validate(config); err != nil {
					// Validation errors are expected in security benchmarks
					b.Logf("Validation error (expected): %v", err)
				}
			}
		})
	}
}

// BenchmarkConfigSecurityExpansion benchmarks path expansion security
func BenchmarkConfigSecurityExpansion(b *testing.B) {
	testPaths := []struct {
		name  string
		paths []string
	}{
		{
			name: "simple_expansion",
			paths: []string{
				"$HOME/config.yaml",
				"${HOME}/config.yaml",
				"$USER/config.yaml",
				"./config.yaml",
			},
		},
		{
			name: "complex_expansion",
			paths: []string{
				"$HOME/$USER/$SHELL/config.yaml",
				"${HOME}/${USER}/${SHELL}/config.yaml",
				"$HOME/$(whoami)/config.yaml",
				"`echo $HOME`/config.yaml",
			},
		},
		{
			name: "malicious_expansion",
			paths: []string{
				"$(rm -rf /)/config.yaml",
				"`rm -rf /`/config.yaml",
				"${HOME:-$(whoami)}/config.yaml",
				"${IFS}malicious${IFS}command",
			},
		},
	}

	config := New()

	for _, tc := range testPaths {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				for _, path := range tc.paths {
					_ = config.expandPath(path)
				}
			}
		})
	}
}

// BenchmarkConfigSecurityTypeConversion benchmarks typed environment variable conversion
func BenchmarkConfigSecurityTypeConversion(b *testing.B) {
	env := NewDefaultEnvProvider()

	// Set up test environment variables
	testVars := map[string]string{
		"TEST_BOOL":     "true",
		"TEST_INT":      "42",
		"TEST_INT64":    "9223372036854775807",
		"TEST_FLOAT64":  "3.14159",
		"TEST_DURATION": "5m30s",
		"TEST_SLICE":    "a,b,c,d,e",
		"MALICIOUS_VAR": "$(whoami)",
		"LONG_VAR":      strings.Repeat("x", 1000),
	}

	for key, value := range testVars {
		if err := env.Set(key, value); err != nil {
			b.Fatalf("Failed to set environment variable %s: %v", key, err)
		}
	}

	// Clean up after benchmark
	b.Cleanup(func() {
		for key := range testVars {
			if err := env.Unset(key); err != nil {
				b.Logf("Failed to unset environment variable %s: %v", key, err)
			}
		}
	})

	typeTests := []struct {
		name string
		fn   func()
	}{
		{
			name: "GetBool",
			fn: func() {
				_ = env.GetBool("TEST_BOOL", false)
				_ = env.GetBool("MALICIOUS_VAR", false)
			},
		},
		{
			name: "GetInt",
			fn: func() {
				_ = env.GetInt("TEST_INT", 0)
				_ = env.GetInt("MALICIOUS_VAR", 0)
			},
		},
		{
			name: "GetInt64",
			fn: func() {
				_ = env.GetInt64("TEST_INT64", 0)
				_ = env.GetInt64("LONG_VAR", 0)
			},
		},
		{
			name: "GetFloat64",
			fn: func() {
				_ = env.GetFloat64("TEST_FLOAT64", 0.0)
				_ = env.GetFloat64("MALICIOUS_VAR", 0.0)
			},
		},
		{
			name: "GetDuration",
			fn: func() {
				_ = env.GetDuration("TEST_DURATION", time.Second)
				_ = env.GetDuration("MALICIOUS_VAR", time.Second)
			},
		},
		{
			name: "GetStringSlice",
			fn: func() {
				_ = env.GetStringSlice("TEST_SLICE", []string{})
				_ = env.GetStringSlice("LONG_VAR", []string{})
			},
		},
	}

	for _, tc := range typeTests {
		b.Run(tc.name, func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				tc.fn()
			}
		})
	}
}

// Helper functions for benchmarks

func generateTestConfig(size int) map[string]interface{} {
	config := make(map[string]interface{})

	for i := 0; i < size; i++ {
		config[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
		config[fmt.Sprintf("nested_%d", i)] = map[string]interface{}{
			"inner_key": fmt.Sprintf("inner_value_%d", i),
			"number":    i,
			"boolean":   i%2 == 0,
		}
		config[fmt.Sprintf("array_%d", i)] = []string{
			fmt.Sprintf("item_%d_1", i),
			fmt.Sprintf("item_%d_2", i),
			fmt.Sprintf("item_%d_3", i),
		}
	}

	return config
}

func generateMixedCommands(count int) []string {
	commands := make([]string, count)
	safeCommands := []string{
		"go build", "npm install", "docker build", "echo hello", "ls -la",
	}
	dangerousCommands := []string{
		"$(whoami)", "`id`", "rm -rf /", "cat /etc/passwd", "command; evil",
	}

	for i := 0; i < count; i++ {
		if i%3 == 0 {
			commands[i] = dangerousCommands[i%len(dangerousCommands)]
		} else {
			commands[i] = safeCommands[i%len(safeCommands)]
		}
	}

	return commands
}

func generateYAMLConfig(size int) string {
	var sb strings.Builder
	sb.WriteString("app:\n")
	sb.WriteString("  name: benchmark-app\n")
	sb.WriteString("  port: 8080\n")
	sb.WriteString("database:\n")
	sb.WriteString("  host: localhost\n")
	sb.WriteString("  port: 5432\n")
	sb.WriteString("config:\n")

	for i := 0; i < size; i++ {
		fmt.Fprintf(&sb, "  key_%d: value_%d\n", i, i)
		if i%10 == 0 {
			fmt.Fprintf(&sb, "  nested_%d:\n", i)
			fmt.Fprintf(&sb, "    inner_key: inner_value_%d\n", i)
			fmt.Fprintf(&sb, "    number: %d\n", i)
		}
	}

	return sb.String()
}

func generateJSONConfig(size int) string {
	config := map[string]interface{}{
		"app": map[string]interface{}{
			"name": "benchmark-app",
			"port": 8080,
		},
		"database": map[string]interface{}{
			"host": "localhost",
			"port": 5432,
		},
		"config": make(map[string]interface{}),
	}

	configMap, ok := config["config"].(map[string]interface{})
	if !ok {
		panic("config field is not a map[string]interface{}")
	}
	for i := 0; i < size; i++ {
		configMap[fmt.Sprintf("key_%d", i)] = fmt.Sprintf("value_%d", i)
		if i%10 == 0 {
			configMap[fmt.Sprintf("nested_%d", i)] = map[string]interface{}{
				"inner_key": fmt.Sprintf("inner_value_%d", i),
				"number":    i,
			}
		}
	}

	// Convert to JSON string
	jsonBytes, err := json.Marshal(config)
	if err != nil {
		panic(fmt.Sprintf("Failed to marshal config: %v", err))
	}
	return string(jsonBytes)
}
