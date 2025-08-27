package mage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseSafeTestArgs(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expected    []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid timeout with space separator",
			args:        []string{"-timeout", "5m"},
			expected:    []string{"-timeout", "5m"},
			expectError: false,
		},
		{
			name:        "valid timeout with equals",
			args:        []string{"-timeout=5m"},
			expected:    []string{"-timeout=5m"},
			expectError: false,
		},
		{
			name:        "multiple valid flags including timeout",
			args:        []string{"-v", "-timeout", "30s", "-json"},
			expected:    []string{"-v", "-timeout", "30s", "-json"},
			expectError: false,
		},
		{
			name:        "various duration formats",
			args:        []string{"-timeout", "1h30m45s"},
			expected:    []string{"-timeout", "1h30m45s"},
			expectError: false,
		},
		{
			name:        "invalid flag rejected",
			args:        []string{"-invalid"},
			expected:    nil,
			expectError: true,
			errorMsg:    "flag not allowed for security reasons: -invalid",
		},
		{
			name:        "mixed valid and invalid flags",
			args:        []string{"-v", "-invalid", "-timeout", "5m"},
			expected:    nil,
			expectError: true,
			errorMsg:    "flag not allowed for security reasons: -invalid",
		},
		{
			name:        "empty args",
			args:        []string{},
			expected:    nil,
			expectError: false,
		},
		{
			name:        "single valid flag",
			args:        []string{"-json"},
			expected:    []string{"-json"},
			expectError: false,
		},
		{
			name:        "timeout flag without value",
			args:        []string{"-timeout"},
			expected:    []string{"-timeout"},
			expectError: false,
		},
		{
			name:        "count flag with value",
			args:        []string{"-count", "3"},
			expected:    []string{"-count", "3"},
			expectError: false,
		},
		{
			name:        "count flag with equals",
			args:        []string{"-count=3"},
			expected:    []string{"-count=3"},
			expectError: false,
		},
		{
			name:        "non-flag arguments allowed",
			args:        []string{"-timeout", "5m", "./pkg/utils"},
			expected:    []string{"-timeout", "5m", "./pkg/utils"},
			expectError: false,
		},
		{
			name:        "benchtime flag",
			args:        []string{"-benchtime", "10s"},
			expected:    []string{"-benchtime", "10s"},
			expectError: false,
		},
		{
			name:        "run flag with pattern",
			args:        []string{"-run", "TestMyFunction"},
			expected:    []string{"-run", "TestMyFunction"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseSafeTestArgs(tt.args)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestHasTimeoutFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "timeout with space separator",
			args:     []string{"-timeout", "5m"},
			expected: true,
		},
		{
			name:     "timeout with equals",
			args:     []string{"-timeout=5m"},
			expected: true,
		},
		{
			name:     "no timeout flag",
			args:     []string{"-v", "-json"},
			expected: false,
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: false,
		},
		{
			name:     "timeout among other flags",
			args:     []string{"-v", "-timeout", "30s", "-json"},
			expected: true,
		},
		{
			name:     "timeout equals among other flags",
			args:     []string{"-v", "-timeout=30s", "-json"},
			expected: true,
		},
		{
			name:     "partial match should not match",
			args:     []string{"-timeouts", "5m"},
			expected: false,
		},
		{
			name:     "timeout flag without value",
			args:     []string{"-timeout"},
			expected: true,
		},
		{
			name:     "multiple timeout flags",
			args:     []string{"-timeout", "5m", "-timeout=10s"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasTimeoutFlag(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoveTimeoutFlag(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected []string
	}{
		{
			name:     "remove timeout with space separator",
			args:     []string{"-timeout", "5m", "-v"},
			expected: []string{"-v"},
		},
		{
			name:     "remove timeout with equals",
			args:     []string{"-v", "-timeout=5m", "-json"},
			expected: []string{"-v", "-json"},
		},
		{
			name:     "no timeout flag present",
			args:     []string{"-v", "-json"},
			expected: []string{"-v", "-json"},
		},
		{
			name:     "empty args",
			args:     []string{},
			expected: nil,
		},
		{
			name:     "only timeout flag",
			args:     []string{"-timeout", "5m"},
			expected: nil,
		},
		{
			name:     "only timeout equals flag",
			args:     []string{"-timeout=5m"},
			expected: nil,
		},
		{
			name:     "preserve flag order",
			args:     []string{"-v", "-timeout", "5m", "-json", "-race"},
			expected: []string{"-v", "-json", "-race"},
		},
		{
			name:     "timeout flag without value",
			args:     []string{"-timeout", "-v"},
			expected: []string{"-v"},
		},
		{
			name:     "multiple timeout flags",
			args:     []string{"-timeout", "5m", "-v", "-timeout=10s", "-json"},
			expected: []string{"-v", "-json"},
		},
		{
			name:     "timeout at end of args",
			args:     []string{"-v", "-json", "-timeout", "5m"},
			expected: []string{"-v", "-json"},
		},
		{
			name:     "timeout with path arguments",
			args:     []string{"-timeout", "5m", "./pkg/utils"},
			expected: []string{"./pkg/utils"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := removeTimeoutFlag(tt.args)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildTestArgsTimeoutHandling(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		race           bool
		cover          bool
		additionalArgs []string
		expectedArgs   []string
	}{
		{
			name: "config timeout used when no user timeout",
			config: &Config{
				Test: TestConfig{
					Timeout: "10m",
					Verbose: true,
				},
			},
			race:           false,
			cover:          false,
			additionalArgs: []string{"-json"},
			expectedArgs:   []string{"test", "-v", "-timeout", "10m", "-json"},
		},
		{
			name: "user timeout overrides config timeout",
			config: &Config{
				Test: TestConfig{
					Timeout: "10m",
					Verbose: true,
				},
			},
			race:           false,
			cover:          false,
			additionalArgs: []string{"-timeout", "5m", "-json"},
			expectedArgs:   []string{"test", "-v", "-timeout", "5m", "-json"},
		},
		{
			name: "user timeout equals format overrides config",
			config: &Config{
				Test: TestConfig{
					Timeout: "10m",
					Verbose: true,
				},
			},
			race:           false,
			cover:          false,
			additionalArgs: []string{"-timeout=5m", "-json"},
			expectedArgs:   []string{"test", "-v", "-timeout=5m", "-json"},
		},
		{
			name: "no config timeout, user provides timeout",
			config: &Config{
				Test: TestConfig{
					Verbose: true,
				},
			},
			race:           false,
			cover:          false,
			additionalArgs: []string{"-timeout", "5m"},
			expectedArgs:   []string{"test", "-v", "-timeout", "5m"},
		},
		{
			name: "race and cover flags with timeout handling",
			config: &Config{
				Test: TestConfig{
					Timeout: "10m",
				},
			},
			race:           true,
			cover:          true,
			additionalArgs: []string{"-timeout", "5m"},
			expectedArgs:   []string{"test", "-race", "-cover", "-timeout", "5m"},
		},
		{
			name: "empty additional args uses config defaults",
			config: &Config{
				Test: TestConfig{
					Timeout: "15m",
					Verbose: true,
					Tags:    "integration",
				},
			},
			race:         false,
			cover:        false,
			expectedArgs: []string{"test", "-v", "-timeout", "15m", "-tags", "integration"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildTestArgs(tt.config, tt.race, tt.cover, tt.additionalArgs...)
			assert.Equal(t, tt.expectedArgs, result)
		})
	}
}

func TestBuildTestArgsWithOverridesTimeoutHandling(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		raceOverride   *bool
		coverOverride  *bool
		additionalArgs []string
		expectedArgs   []string
	}{
		{
			name: "user timeout overrides config with race enabled",
			config: &Config{
				Test: TestConfig{
					Timeout: "10m",
					Race:    false,
				},
			},
			raceOverride:   boolPtr(true),
			coverOverride:  nil,
			additionalArgs: []string{"-timeout", "5m"},
			expectedArgs:   []string{"test", "-race", "-timeout", "5m"},
		},
		{
			name: "user timeout overrides config with cover enabled",
			config: &Config{
				Test: TestConfig{
					Timeout: "10m",
					Cover:   false,
				},
			},
			raceOverride:   nil,
			coverOverride:  boolPtr(true),
			additionalArgs: []string{"-timeout=30s"},
			expectedArgs:   []string{"test", "-cover", "-timeout=30s"},
		},
		{
			name: "combination of user flags and overrides",
			config: &Config{
				Test: TestConfig{
					Timeout: "10m",
					Verbose: true,
					Race:    true,
					Cover:   true,
				},
			},
			raceOverride:   boolPtr(false),
			coverOverride:  boolPtr(false),
			additionalArgs: []string{"-timeout", "2m", "-json"},
			expectedArgs:   []string{"test", "-v", "-timeout", "2m", "-json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildTestArgsWithOverrides(tt.config, tt.raceOverride, tt.coverOverride, tt.additionalArgs...)
			assert.Equal(t, tt.expectedArgs, result)
		})
	}
}

// Helper function to create bool pointers for tests
func boolPtr(b bool) *bool {
	return &b
}
