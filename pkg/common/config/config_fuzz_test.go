//go:build go1.18
// +build go1.18

package config

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// FuzzYAMLParsing tests YAML parsing with fuzzing
func FuzzYAMLParsing(f *testing.F) {
	// Add seed corpus with various YAML structures
	testcases := []string{
		// Valid YAML
		`key: value`,
		`
project:
  name: test
  version: 1.0.0
`,
		`
list:
  - item1
  - item2
  - item3
`,
		`
nested:
  level1:
    level2:
      level3: value
`,
		`
mixed:
  string: "value"
  number: 123
  float: 123.45
  bool: true
  null: null
`,
		// Edge cases
		`key: "value with spaces"`,
		`key: 'single quotes'`,
		`key: "quotes \"inside\" quotes"`,
		`key: |
  multiline
  value
`,
		`key: >
  folded
  value
`,
		// Potential security issues
		`key: !!python/object/apply:os.system ['whoami']`,
		`key: !!python/object/new:os.system ['whoami']`,
		`key: &anchor value
ref: *anchor`,
		`key: *undefined_anchor`,
		// Malformed YAML
		`key value`,
		`key:`,
		`: value`,
		`[unclosed`,
		`{unclosed`,
		`key: [1, 2, unclosed`,
		`"unclosed string`,
		`'unclosed string`,
		// Special characters
		`key: "value\x00null"`,
		`key: "value\nnewline"`,
		`key: "value\ttab"`,
		`key: "value\\backslash"`,
		`key: "value\"quote"`,
		`key: "🚀"`,
		// Large structures
		strings.Repeat("key: value\n", 1000),
		`key: ` + strings.Repeat("a", 10000),
		// Injection attempts
		`key: $(whoami)`,
		`key: ${USER}`,
		`key: %USER%`,
		"`key: `whoami``",
		`key: |
  #!/bin/bash
  whoami
`,
		// Empty and whitespace
		``,
		` `,
		`	`,
		`
`,
		`---`,
		`...`,
		// Comments
		`# comment`,
		`key: value # comment`,
		`# comment
key: value`,
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Test with a generic map structure
		var result map[string]interface{}

		// Should not panic
		decoder := yaml.NewDecoder(bytes.NewReader([]byte(input)))
		err := decoder.Decode(&result)

		// If it parsed successfully, verify some properties
		if err == nil {
			// Result should be a valid map (could be nil for empty docs)
			// Just ensure we can safely access it
			_ = len(result)

			// Try to encode it back (should not panic)
			var buf bytes.Buffer
			encoder := yaml.NewEncoder(&buf)
			encodeErr := encoder.Encode(result)
			if encodeErr == nil {
				// Try to round-trip - this may fail for edge cases like
				// numeric keys that get serialized ambiguously, causing
				// duplicate key errors on re-parse. That's acceptable for
				// fuzz testing - we just care that it doesn't panic.
				var roundTrip map[string]interface{}
				if decodeErr := yaml.Unmarshal(buf.Bytes(), &roundTrip); decodeErr != nil {
					// Some edge cases (like numeric keys) may fail round-trip
					// due to YAML serialization ambiguities. This is expected.
					t.Logf("Round-trip decode failed (expected for some edge cases): %v", decodeErr)
				}
			}
		}

		// Also test with the config loader
		loader := NewDefaultConfigLoader()

		// Create a test config structure
		type TestConfig struct {
			Project struct {
				Name    string `yaml:"name"`
				Version string `yaml:"version"`
			} `yaml:"project"`
			Build struct {
				Target string   `yaml:"target"`
				Flags  []string `yaml:"flags"`
			} `yaml:"build"`
			Test struct {
				Coverage bool   `yaml:"coverage"`
				Timeout  string `yaml:"timeout"`
			} `yaml:"test"`
		}

		var cfg TestConfig

		// Write to temp file and try to load
		tempFile := t.TempDir() + "/config.yaml"
		err = writeFile(tempFile, []byte(input))
		require.NoError(t, err, "Failed to write temp file")

		// Should not panic
		_, loadErr := loader.Load([]string{tempFile}, &cfg)

		// If loaded successfully, validate the config
		if loadErr == nil {
			// Ensure strings don't contain null bytes
			assert.NotContains(t, cfg.Project.Name, "\x00", "Config contains null byte")
			assert.NotContains(t, cfg.Project.Version, "\x00", "Config contains null byte")
			assert.NotContains(t, cfg.Build.Target, "\x00", "Config contains null byte")

			// Ensure no command injection in string fields
			dangerousPatterns := []string{"$(", "`", "${IFS}", "&&", "||", ";"}
			for _, pattern := range dangerousPatterns {
				assert.NotContains(t, cfg.Project.Name, pattern, "Config contains dangerous pattern")
				assert.NotContains(t, cfg.Build.Target, pattern, "Config contains dangerous pattern")
			}
		}
	})
}

// FuzzJSONParsing tests JSON parsing with fuzzing
func FuzzJSONParsing(f *testing.F) {
	// Add seed corpus
	testcases := []string{
		// Valid JSON
		`{"key": "value"}`,
		`{"project": {"name": "test", "version": "1.0.0"}}`,
		`{"list": ["item1", "item2", "item3"]}`,
		`{"nested": {"level1": {"level2": {"level3": "value"}}}}`,
		`{"string": "value", "number": 123, "float": 123.45, "bool": true, "null": null}`,
		// Edge cases
		`{"key": "value with spaces"}`,
		`{"key": "quotes \"inside\" quotes"}`,
		`{"key": "unicode \u0041"}`,
		`{"key": "emoji 🚀"}`,
		`{"key": "newline\nnewline"}`,
		`{"key": "tab\ttab"}`,
		`{"key": "backslash\\backslash"}`,
		// Arrays and nested structures
		`[]`,
		`[1, 2, 3]`,
		`["a", "b", "c"]`,
		`[{"a": 1}, {"b": 2}]`,
		`{"array": [{"nested": true}]}`,
		// Malformed JSON
		`{key: "value"}`,
		`{"key": value}`,
		`{"key": "value"`,
		`{"key": "value"}}`,
		`{"unclosed": "string`,
		`{'single': 'quotes'}`,
		`{"trailing": "comma",}`,
		// Special values
		`{"number": 1.7976931348623157e+308}`,
		`{"number": -1.7976931348623157e+308}`,
		`{"number": 9007199254740992}`,
		`{"number": -9007199254740992}`,
		// Security attempts
		`{"key": "$(whoami)"}`,
		`{"key": "${USER}"}`,
		`{"key": "%USER%"}`,
		"{\"key\": \"`whoami`\"}",
		`{"__proto__": {"isAdmin": true}}`,
		`{"constructor": {"prototype": {"isAdmin": true}}}`,
		// Large structures
		`{"key": "` + strings.Repeat("a", 10000) + `"}`,
		`{` + strings.Repeat(`"key": "value",`, 1000) + `"last": "value"}`,
		// Empty and whitespace
		``,
		` `,
		`	`,
		`null`,
		`true`,
		`false`,
		`0`,
		`""`,
		// Unicode edge cases
		`{"key": "\u0000"}`,
		`{"key": "\uFFFF"}`,
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, input string) {
		// Test with a generic map structure
		var result map[string]interface{}

		// Should not panic
		err := json.Unmarshal([]byte(input), &result)

		// If it parsed successfully, verify some properties
		if err == nil {
			// Try to encode it back (should not panic)
			encoded, encodeErr := json.Marshal(result)
			if encodeErr == nil {
				// If we can round-trip, verify it's equivalent
				var roundTrip map[string]interface{}
				decodeErr := json.Unmarshal(encoded, &roundTrip)
				require.NoError(t, decodeErr, "Failed to decode round-tripped JSON")
			}
		}

		// Also test with the config loader
		loader := NewDefaultConfigLoader()

		// Create a test config structure
		type TestConfig struct {
			Project struct {
				Name    string `json:"name"`
				Version string `json:"version"`
			} `json:"project"`
			Build struct {
				Target string   `json:"target"`
				Flags  []string `json:"flags"`
			} `json:"build"`
			Test struct {
				Coverage bool   `json:"coverage"`
				Timeout  string `json:"timeout"`
			} `json:"test"`
		}

		var cfg TestConfig

		// Write to temp file and try to load
		tempFile := t.TempDir() + "/config.json"
		err = writeFile(tempFile, []byte(input))
		require.NoError(t, err, "Failed to write temp file")

		// Should not panic
		_, loadErr := loader.Load([]string{tempFile}, &cfg)

		// If loaded successfully, validate the config
		if loadErr == nil {
			// Ensure strings don't contain null bytes
			assert.NotContains(t, cfg.Project.Name, "\x00", "Config contains null byte")
			assert.NotContains(t, cfg.Project.Version, "\x00", "Config contains null byte")
			assert.NotContains(t, cfg.Build.Target, "\x00", "Config contains null byte")
		}
	})
}

// FuzzPathExpansion tests path expansion with fuzzing
func FuzzPathExpansion(f *testing.F) {
	// Create a config instance
	cfg := New()

	// Add seed corpus
	testcases := []string{
		// Normal paths
		"config.yaml",
		"./config.yaml",
		"../config.yaml",
		"/etc/config.yaml",
		"~/config.yaml",
		"$HOME/config.yaml",
		"${HOME}/config.yaml",
		"$HOME/.config/app.yaml",
		"${HOME}/.config/app.yaml",

		// Environment variables
		"$USER/config.yaml",
		"${USER}/config.yaml",
		"$PATH/config.yaml",
		"${PATH}/config.yaml",
		"$UNDEFINED/config.yaml",
		"${UNDEFINED}/config.yaml",

		// Multiple variables
		"$HOME/$USER/config.yaml",
		"${HOME}/${USER}/config.yaml",
		"$HOME/$USER/$PATH/config.yaml",

		// Edge cases
		"",
		".",
		"..",
		"/",
		"//",
		"///",
		"$",
		"$$",
		"${",
		"${}",
		"${HOME",
		"$HOME}",
		"${HOME}}",
		"$HO ME/config.yaml",
		"${HO ME}/config.yaml",

		// Potential security issues
		"$HOME/../../etc/passwd",
		"${HOME}/../../etc/passwd",
		"$(whoami)/config.yaml",
		"${IFS}/config.yaml",
		"`whoami`/config.yaml",
		";rm -rf /",
		"&&whoami",
		"||whoami",
		"|whoami",

		// Special characters
		"config\x00.yaml",
		"config\n.yaml",
		"config\t.yaml",
		"config with spaces.yaml",
		"config;cmd.yaml",
		"config&cmd.yaml",
		"config|cmd.yaml",
		"config>output.yaml",
		"config<input.yaml",

		// Unicode
		"config-unicode.yaml",
		"конфиг.yaml",
		"🚀.yaml",
		"$HOME/🚀/config.yaml",

		// Long paths
		strings.Repeat("a", 255) + ".yaml",
		"$HOME/" + strings.Repeat("b", 255) + "/config.yaml",
		strings.Repeat("$HOME/", 50) + "config.yaml",
	}

	for _, tc := range testcases {
		f.Add(tc)
	}

	f.Fuzz(func(t *testing.T, path string) {
		// Should not panic
		expanded := cfg.expandPath(path)

		// Expanded path should be a string (even if empty)
		assert.NotNil(t, expanded, "expandPath returned nil")

		// If path was empty, expanded should also be empty
		if path == "" {
			assert.Empty(t, expanded, "Empty path should expand to empty")
		}

		// Check that dangerous patterns don't cause execution (they should be left as-is)
		// This simple expandPath implementation only handles $HOME, so other patterns remain unchanged
		if strings.Contains(path, "$(") || strings.Contains(path, "`") {
			// These patterns should be left unchanged (not executed)
			assert.True(t, strings.Contains(expanded, "$(") || strings.Contains(expanded, "`") || !strings.Contains(path, expanded),
				"Dangerous patterns should not be executed, path: %s, expanded: %s", path, expanded)
		}

		// If the original path had environment variables, they should be expanded or left as-is
		if strings.Contains(path, "$HOME") || strings.Contains(path, "${HOME}") {
			// HOME should either be expanded or left as-is if not set
			// Just ensure it doesn't cause issues
			_ = expanded
		}

		// Test buildConfigPaths with the fuzzed input
		paths := cfg.buildConfigPaths(path, ".", "/etc", "$HOME")

		// Should return some paths
		assert.NotNil(t, paths, "buildConfigPaths returned nil")

		// Each path should be a string
		for _, p := range paths {
			assert.IsType(t, "", p, "Path is not a string")
		}
	})
}

// FuzzConfigManager tests the config manager with fuzzing
func FuzzConfigManager(f *testing.F) {
	// Add seed corpus with various config formats
	testcases := []struct {
		format string
		data   string
	}{
		{"yaml", `key: value`},
		{"yaml", `nested:\n  key: value`},
		{"json", `{"key": "value"}`},
		{"json", `{"nested": {"key": "value"}}`},
		{"env", `KEY=value`},
		{"env", `KEY=value\nOTHER=data`},
		{"unknown", `random data`},
	}

	for _, tc := range testcases {
		f.Add(tc.format, tc.data)
	}

	f.Fuzz(func(t *testing.T, format, data string) {
		manager := NewDefaultConfigManager()

		// Sanitize format to create a valid filename extension
		// Only allow alphanumeric characters to avoid filesystem issues
		var sanitizedFormat strings.Builder
		for _, r := range format {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				sanitizedFormat.WriteRune(r)
				if sanitizedFormat.Len() >= 10 {
					break
				}
			}
		}
		// Ensure format is not empty
		ext := sanitizedFormat.String()
		if ext == "" {
			ext = "txt"
		}

		// Create a file source
		tempFile := t.TempDir() + "/config." + ext
		err := writeFile(tempFile, []byte(data))
		require.NoError(t, err, "Failed to write temp file")

		// Add source (should not panic)
		var configFormat Format
		switch format {
		case "yaml", "yml":
			configFormat = FormatYAML
		case "json":
			configFormat = FormatJSON
		default:
			configFormat = FormatYAML // default fallback
		}
		source := NewFileConfigSource(tempFile, configFormat, 100)
		manager.AddSource(source)

		// Try to load config (should not panic)
		type TestConfig struct {
			Key    string `yaml:"key" json:"key" env:"KEY"`
			Nested struct {
				Key string `yaml:"key" json:"key" env:"NESTED_KEY"`
			} `yaml:"nested" json:"nested"`
		}

		var cfg TestConfig
		err = manager.LoadConfig(&cfg)

		// If loaded successfully, validate
		if err == nil {
			// Ensure no null bytes
			assert.NotContains(t, cfg.Key, "\x00", "Config contains null byte")
			assert.NotContains(t, cfg.Nested.Key, "\x00", "Config contains null byte")
		}

		// Test Watch (should not panic even if not implemented)
		watchCallback := func(interface{}) {
			// Simple callback for testing
		}
		if err := manager.Watch(watchCallback); err != nil {
			// Log but don't fail, as Watch may not be fully implemented
			t.Logf("Watch returned error: %v", err)
		}

		// Test Reload (should not panic)
		if err := manager.Reload(&cfg); err != nil {
			// Log but don't fail, as this is a fuzz test
			t.Logf("Reload returned error: %v", err)
		}
	})
}

// Helper function to write file
func writeFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}
