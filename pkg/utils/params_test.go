package utils

import (
	"reflect"
	"testing"
)

func TestParseParams(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected map[string]string
	}{
		{
			name:     "empty args",
			args:     []string{},
			expected: map[string]string{},
		},
		{
			name:     "key=value pairs",
			args:     []string{"bump=patch", "output=json", "count=5"},
			expected: map[string]string{"bump": "patch", "output": "json", "count": "5"},
		},
		{
			name:     "boolean flags",
			args:     []string{"verbose", "push", "dry-run"},
			expected: map[string]string{"verbose": "true", "push": "true", "dry-run": "true"},
		},
		{
			name:     "mixed format",
			args:     []string{"bump=minor", "push", "timeout=30s", "verbose"},
			expected: map[string]string{"bump": "minor", "push": "true", "timeout": "30s", "verbose": "true"},
		},
		{
			name:     "explicit false values",
			args:     []string{"push=false", "verbose=no", "debug=0"},
			expected: map[string]string{"push": "false", "verbose": "no", "debug": "0"},
		},
		{
			name:     "values with spaces",
			args:     []string{"message=hello world", "path=/some/path"},
			expected: map[string]string{"message": "hello world", "path": "/some/path"},
		},
		{
			name:     "empty keys or values",
			args:     []string{"=value", "key=", ""},
			expected: map[string]string{"key": ""},
		},
		{
			name:     "multiple equals signs",
			args:     []string{"url=http://example.com/path?key=value"},
			expected: map[string]string{"url": "http://example.com/path?key=value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseParams(tt.args)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseParams() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetParam(t *testing.T) {
	params := map[string]string{
		"bump":    "patch",
		"timeout": "30s",
		"verbose": "true",
	}

	tests := []struct {
		name         string
		key          string
		defaultValue string
		expected     string
	}{
		{
			name:         "existing key",
			key:          "bump",
			defaultValue: "minor",
			expected:     "patch",
		},
		{
			name:         "non-existing key",
			key:          "missing",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "empty default",
			key:          "missing",
			defaultValue: "",
			expected:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetParam(params, tt.key, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("GetParam() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestHasParam(t *testing.T) {
	params := map[string]string{
		"bump":    "patch",
		"verbose": "true",
		"empty":   "",
	}

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{
			name:     "existing key with value",
			key:      "bump",
			expected: true,
		},
		{
			name:     "existing key with empty value",
			key:      "empty",
			expected: true,
		},
		{
			name:     "non-existing key",
			key:      "missing",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := HasParam(params, tt.key)
			if result != tt.expected {
				t.Errorf("HasParam() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsParamTrue(t *testing.T) {
	params := map[string]string{
		"true1":   "true",
		"true2":   "TRUE",
		"true3":   "1",
		"true4":   "yes",
		"true5":   "YES",
		"true6":   "on",
		"true7":   "ON",
		"false1":  "false",
		"false2":  "0",
		"false3":  "no",
		"false4":  "off",
		"invalid": "maybe",
		"empty":   "",
	}

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"true value", "true1", true},
		{"TRUE value", "true2", true},
		{"1 value", "true3", true},
		{"yes value", "true4", true},
		{"YES value", "true5", true},
		{"on value", "true6", true},
		{"ON value", "true7", true},
		{"false value", "false1", false},
		{"0 value", "false2", false},
		{"no value", "false3", false},
		{"off value", "false4", false},
		{"invalid value", "invalid", false},
		{"empty value", "empty", false},
		{"missing key", "missing", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsParamTrue(params, tt.key)
			if result != tt.expected {
				t.Errorf("IsParamTrue() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsParamFalse(t *testing.T) {
	params := map[string]string{
		"false1":  "false",
		"false2":  "FALSE",
		"false3":  "0",
		"false4":  "no",
		"false5":  "NO",
		"false6":  "off",
		"false7":  "OFF",
		"true1":   "true",
		"true2":   "1",
		"invalid": "maybe",
		"empty":   "",
	}

	tests := []struct {
		name     string
		key      string
		expected bool
	}{
		{"false value", "false1", true},
		{"FALSE value", "false2", true},
		{"0 value", "false3", true},
		{"no value", "false4", true},
		{"NO value", "false5", true},
		{"off value", "false6", true},
		{"OFF value", "false7", true},
		{"true value", "true1", false},
		{"1 value", "true2", false},
		{"invalid value", "invalid", false},
		{"empty value", "empty", false},
		{"missing key", "missing", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsParamFalse(params, tt.key)
			if result != tt.expected {
				t.Errorf("IsParamFalse() = %v, expected %v", result, tt.expected)
			}
		})
	}
}