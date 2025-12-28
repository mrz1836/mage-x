package env

import (
	"os"
	"testing"
)

// Test helper to set and restore environment variables
func withEnv(t *testing.T, key, value string) func() {
	t.Helper()
	old, existed := os.LookupEnv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("failed to set env %s: %v", key, err)
	}
	return func() {
		if existed {
			_ = os.Setenv(key, old) //nolint:errcheck // Cleanup in test, best effort
		} else {
			_ = os.Unsetenv(key) //nolint:errcheck // Cleanup in test, best effort
		}
	}
}

// Test helper to unset and restore environment variable
func withoutEnv(t *testing.T, key string) func() {
	t.Helper()
	old, existed := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("failed to unset env %s: %v", key, err)
	}
	return func() {
		if existed {
			_ = os.Setenv(key, old) //nolint:errcheck // Cleanup in test, best effort
		}
	}
}

func TestIsVerbose(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T) func()
		expected bool
	}{
		{
			name: "VERBOSE=true returns true",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "V")
				cleanup2 := withEnv(t, "VERBOSE", "true")
				return func() { cleanup2(); cleanup1() }
			},
			expected: true,
		},
		{
			name: "VERBOSE=1 returns true",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "V")
				cleanup2 := withEnv(t, "VERBOSE", "1")
				return func() { cleanup2(); cleanup1() }
			},
			expected: true,
		},
		{
			name: "V=true returns true",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "VERBOSE")
				cleanup2 := withEnv(t, "V", "true")
				return func() { cleanup2(); cleanup1() }
			},
			expected: true,
		},
		{
			name: "V=1 returns true",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "VERBOSE")
				cleanup2 := withEnv(t, "V", "1")
				return func() { cleanup2(); cleanup1() }
			},
			expected: true,
		},
		{
			name: "both VERBOSE and V unset returns false",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "VERBOSE")
				cleanup2 := withoutEnv(t, "V")
				return func() { cleanup2(); cleanup1() }
			},
			expected: false,
		},
		{
			name: "VERBOSE=false returns false",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "V")
				cleanup2 := withEnv(t, "VERBOSE", "false")
				return func() { cleanup2(); cleanup1() }
			},
			expected: false,
		},
		{
			name: "VERBOSE=0 returns false",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "V")
				cleanup2 := withEnv(t, "VERBOSE", "0")
				return func() { cleanup2(); cleanup1() }
			},
			expected: false,
		},
		{
			name: "empty VERBOSE returns false",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "V")
				cleanup2 := withEnv(t, "VERBOSE", "")
				return func() { cleanup2(); cleanup1() }
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(t)
			defer cleanup()

			result := IsVerbose()
			if result != tt.expected {
				t.Errorf("IsVerbose() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsCI(t *testing.T) {
	// List of all CI environment variables to test
	ciVars := []string{
		"CI",
		"CONTINUOUS_INTEGRATION",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"TRAVIS",
		"CIRCLECI",
		"JENKINS_URL",
		"CODEBUILD_BUILD_ID",
		"BUILDKITE",
		"AZURE_PIPELINES",
		"TEAMCITY_VERSION",
		"BITBUCKET_BUILD_NUMBER",
	}

	// Test that each CI var triggers detection
	for _, ciVar := range ciVars {
		t.Run("detects "+ciVar, func(t *testing.T) {
			// Clear all CI vars first
			cleanups := make([]func(), 0, len(ciVars))
			for _, v := range ciVars {
				cleanups = append(cleanups, withoutEnv(t, v))
			}
			defer func() {
				for _, cleanup := range cleanups {
					cleanup()
				}
			}()

			// Set just this one
			cleanup := withEnv(t, ciVar, "true")
			defer cleanup()

			if !IsCI() {
				t.Errorf("IsCI() should return true when %s is set", ciVar)
			}
		})
	}

	// Test that no CI vars returns false
	t.Run("returns false when no CI vars set", func(t *testing.T) {
		cleanups := make([]func(), 0, len(ciVars))
		for _, v := range ciVars {
			cleanups = append(cleanups, withoutEnv(t, v))
		}
		defer func() {
			for _, cleanup := range cleanups {
				cleanup()
			}
		}()

		if IsCI() {
			t.Error("IsCI() should return false when no CI environment variables are set")
		}
	})
}

func TestIsDebug(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T) func()
		expected bool
	}{
		{
			name: "DEBUG=true returns true",
			setup: func(t *testing.T) func() {
				return withEnv(t, "DEBUG", "true")
			},
			expected: true,
		},
		{
			name: "DEBUG=1 returns true",
			setup: func(t *testing.T) func() {
				return withEnv(t, "DEBUG", "1")
			},
			expected: true,
		},
		{
			name: "DEBUG unset returns false",
			setup: func(t *testing.T) func() {
				return withoutEnv(t, "DEBUG")
			},
			expected: false,
		},
		{
			name: "DEBUG=false returns false",
			setup: func(t *testing.T) func() {
				return withEnv(t, "DEBUG", "false")
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(t)
			defer cleanup()

			result := IsDebug()
			if result != tt.expected {
				t.Errorf("IsDebug() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsQuiet(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T) func()
		expected bool
	}{
		{
			name: "QUIET=true returns true",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "Q")
				cleanup2 := withEnv(t, "QUIET", "true")
				return func() { cleanup2(); cleanup1() }
			},
			expected: true,
		},
		{
			name: "Q=true returns true",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "QUIET")
				cleanup2 := withEnv(t, "Q", "true")
				return func() { cleanup2(); cleanup1() }
			},
			expected: true,
		},
		{
			name: "neither set returns false",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "QUIET")
				cleanup2 := withoutEnv(t, "Q")
				return func() { cleanup2(); cleanup1() }
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(t)
			defer cleanup()

			result := IsQuiet()
			if result != tt.expected {
				t.Errorf("IsQuiet() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsDryRun(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T) func()
		expected bool
	}{
		{
			name: "DRY_RUN=true returns true",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "DRYRUN")
				cleanup2 := withEnv(t, "DRY_RUN", "true")
				return func() { cleanup2(); cleanup1() }
			},
			expected: true,
		},
		{
			name: "DRYRUN=true returns true",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "DRY_RUN")
				cleanup2 := withEnv(t, "DRYRUN", "true")
				return func() { cleanup2(); cleanup1() }
			},
			expected: true,
		},
		{
			name: "neither set returns false",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "DRY_RUN")
				cleanup2 := withoutEnv(t, "DRYRUN")
				return func() { cleanup2(); cleanup1() }
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(t)
			defer cleanup()

			result := IsDryRun()
			if result != tt.expected {
				t.Errorf("IsDryRun() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestIsTest(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T) func()
		expected bool
	}{
		{
			name: "GO_TEST set returns true",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "TESTING")
				cleanup2 := withEnv(t, "GO_TEST", "1")
				return func() { cleanup2(); cleanup1() }
			},
			expected: true,
		},
		{
			name: "TESTING=true returns true",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "GO_TEST")
				cleanup2 := withEnv(t, "TESTING", "true")
				return func() { cleanup2(); cleanup1() }
			},
			expected: true,
		},
		{
			name: "neither set returns false",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "GO_TEST")
				cleanup2 := withoutEnv(t, "TESTING")
				return func() { cleanup2(); cleanup1() }
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(t)
			defer cleanup()

			result := IsTest()
			if result != tt.expected {
				t.Errorf("IsTest() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T) func()
		expected string
	}{
		{
			name: "LOG_LEVEL=debug returns debug",
			setup: func(t *testing.T) func() {
				return withEnv(t, "LOG_LEVEL", "debug")
			},
			expected: "debug",
		},
		{
			name: "LOG_LEVEL=error returns error",
			setup: func(t *testing.T) func() {
				return withEnv(t, "LOG_LEVEL", "error")
			},
			expected: "error",
		},
		{
			name: "LOG_LEVEL unset returns info",
			setup: func(t *testing.T) func() {
				return withoutEnv(t, "LOG_LEVEL")
			},
			expected: "info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(t)
			defer cleanup()

			result := GetLogLevel()
			if result != tt.expected {
				t.Errorf("GetLogLevel() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetParallelism(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T) func()
		expected int
	}{
		{
			name: "PARALLEL=4 returns 4",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "GOMAXPROCS")
				cleanup2 := withEnv(t, "PARALLEL", "4")
				return func() { cleanup2(); cleanup1() }
			},
			expected: 4,
		},
		{
			name: "GOMAXPROCS=8 returns 8 when PARALLEL unset",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "PARALLEL")
				cleanup2 := withEnv(t, "GOMAXPROCS", "8")
				return func() { cleanup2(); cleanup1() }
			},
			expected: 8,
		},
		{
			name: "PARALLEL takes precedence over GOMAXPROCS",
			setup: func(t *testing.T) func() {
				cleanup1 := withEnv(t, "GOMAXPROCS", "8")
				cleanup2 := withEnv(t, "PARALLEL", "2")
				return func() { cleanup2(); cleanup1() }
			},
			expected: 2,
		},
		{
			name: "neither set returns 0",
			setup: func(t *testing.T) func() {
				cleanup1 := withoutEnv(t, "PARALLEL")
				cleanup2 := withoutEnv(t, "GOMAXPROCS")
				return func() { cleanup2(); cleanup1() }
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(t)
			defer cleanup()

			result := GetParallelism()
			if result != tt.expected {
				t.Errorf("GetParallelism() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetTimeout(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*testing.T) func()
		expected string
	}{
		{
			name: "TIMEOUT=30s returns 30s",
			setup: func(t *testing.T) func() {
				return withEnv(t, "TIMEOUT", "30s")
			},
			expected: "30s",
		},
		{
			name: "TIMEOUT=5m returns 5m",
			setup: func(t *testing.T) func() {
				return withEnv(t, "TIMEOUT", "5m")
			},
			expected: "5m",
		},
		{
			name: "TIMEOUT unset returns empty string",
			setup: func(t *testing.T) func() {
				return withoutEnv(t, "TIMEOUT")
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup(t)
			defer cleanup()

			result := GetTimeout()
			if result != tt.expected {
				t.Errorf("GetTimeout() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
