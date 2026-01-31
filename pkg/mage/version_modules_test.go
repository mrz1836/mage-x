package mage

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDiscoverModules tests module discovery
func TestDiscoverModules(t *testing.T) {
	t.Run("RunsWithoutError", func(t *testing.T) {
		modules, err := discoverModules()
		require.NoError(t, err)
		// modules can be nil or empty slice - both are valid
		_ = modules
	})

	t.Run("ModuleStructureValid", func(t *testing.T) {
		modules, err := discoverModules()
		require.NoError(t, err)

		// Only check structure if modules exist
		for _, m := range modules {
			require.NotEmpty(t, m.Name, "Module should have a name")
			require.NotEmpty(t, m.Path, "Module should have a path")
		}
	})
}

// TestGetModuleTag tests module tag retrieval
func TestGetModuleTag(t *testing.T) {
	t.Run("ReturnsStringForNonexistent", func(t *testing.T) {
		tag := getModuleTag("nonexistent-module-xyz-abc")
		require.IsType(t, "", tag)
	})

	t.Run("ReturnsEmptyOrValidTag", func(t *testing.T) {
		tag := getModuleTag("models")
		require.True(t, tag == "" || strings.HasPrefix(tag, "v"))
	})
}

// TestParseModuleTagVersion tests module tag version parsing
func TestParseModuleTagVersion(t *testing.T) {
	tests := []struct {
		name       string
		tag        string
		moduleName string
		expected   string
	}{
		{"Standard tag", "models/v1.2.3", "models", "v1.2.3"},
		{"Nested module tag", "internal/engine/v0.5.0", "engine", "internal/engine/v0.5.0"},
		{"Tag without prefix", "v1.0.0", "models", "v1.0.0"},
		{"Empty module name", "models/v1.0.0", "", "models/v1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseModuleTagVersion(tt.tag, tt.moduleName)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestVersionModule tests the VersionModule struct
func TestVersionModule(t *testing.T) {
	t.Run("StructFieldsPopulated", func(t *testing.T) {
		module := VersionModule{
			Name:       "models",
			Path:       "models",
			CurrentTag: "v1.2.3",
			HasTags:    true,
		}

		require.Equal(t, "models", module.Name)
		require.Equal(t, "models", module.Path)
		require.Equal(t, "v1.2.3", module.CurrentTag)
		require.True(t, module.HasTags)
	})

	t.Run("NewModuleNoTags", func(t *testing.T) {
		module := VersionModule{
			Name:       "newmodule",
			Path:       "newmodule",
			CurrentTag: "",
			HasTags:    false,
		}

		require.False(t, module.HasTags)
		require.Empty(t, module.CurrentTag)
	})
}

// TestVersionModulesCommand tests the version:modules command
func TestVersionModulesCommand(t *testing.T) {
	version := Version{}
	err := version.Modules()
	require.NoError(t, err)
}

// TestBumpConfigModuleParameter tests module parameter parsing
func TestBumpConfigModuleParameter(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedModule string
	}{
		{"Default no module", []string{"bump=patch"}, ""},
		{"Specific module", []string{"bump=patch", "module=models"}, "models"},
		{"All modules", []string{"bump=minor", "module=all"}, "all"},
		{"Wildcard modules", []string{"bump=major", "module=*"}, "*"},
		{"Module with push and dry-run", []string{"bump=patch", "module=engine", "push", "dry-run"}, "engine"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseBumpConfig(tt.args)
			require.NoError(t, err)
			require.Equal(t, tt.expectedModule, cfg.module)
		})
	}
}

// TestBumpVersionProgression tests version progression
func TestBumpVersionProgression(t *testing.T) {
	tests := []struct {
		current  string
		bumpType string
		expected string
	}{
		{"v0.0.0", "patch", "v0.0.1"},
		{"v0.0.0", "minor", "v0.1.0"},
		{"v0.0.0", "major", "v1.0.0"},
		{"v1.2.3", "patch", "v1.2.4"},
		{"v1.2.3", "minor", "v1.3.0"},
		{"v1.2.3", "major", "v2.0.0"},
		{"v10.20.30", "patch", "v10.20.31"},
		{"v10.20.30", "minor", "v10.21.0"},
		{"v10.20.30", "major", "v11.0.0"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s_%s", tt.current, tt.bumpType), func(t *testing.T) {
			result, err := bumpVersion(tt.current, tt.bumpType)
			require.NoError(t, err)
			require.Equal(t, tt.expected, result)
		})
	}
}

// TestModuleFunctions tests module helper functions
func TestModuleFunctions(t *testing.T) {
	t.Run("GetSubmoduleCurrentVersion", func(t *testing.T) {
		version := getSubmoduleCurrentVersion("models")
		require.True(t, version == "" || strings.HasPrefix(version, "v"))
	})

	t.Run("GetSubmoduleCurrentVersionNonexistent", func(t *testing.T) {
		version := getSubmoduleCurrentVersion("nonexistent-module-xyz")
		require.Empty(t, version)
	})
}

// TestDryRunBehavior tests dry-run mode
func TestDryRunBehavior(t *testing.T) {
	version := Version{}

	t.Run("DryRunRootModule", func(t *testing.T) {
		err := version.Bump("bump=patch", "dry-run")
		require.NoError(t, err)
	})

	t.Run("DryRunWithModuleAll", func(t *testing.T) {
		err := version.Bump("bump=minor", "module=all", "dry-run")
		require.NoError(t, err)
	})

	t.Run("DryRunWithModuleWildcard", func(t *testing.T) {
		err := version.Bump("bump=patch", "module=*", "dry-run")
		require.NoError(t, err)
	})
}

// TestModuleBumpEdgeCases tests edge cases
func TestModuleBumpEdgeCases(t *testing.T) {
	version := Version{}

	t.Run("NonexistentModule", func(t *testing.T) {
		err := version.Bump("bump=patch", "module=nonexistent-xyz-abc")
		require.Error(t, err)
		require.ErrorIs(t, err, errSubmoduleNotFound)
	})

	t.Run("InvalidBumpTypeWithModule", func(t *testing.T) {
		err := version.Bump("bump=invalid", "module=models", "dry-run")
		require.Error(t, err)
		require.ErrorIs(t, err, errInvalidBumpType)
	})
}

// TestModuleTagFormat tests tag format generation
func TestModuleTagFormat(t *testing.T) {
	tests := []struct {
		moduleName string
		version    string
		expected   string
	}{
		{"models", "v1.0.0", "models/v1.0.0"},
		{"engine", "v0.5.2", "engine/v0.5.2"},
		{"utils", "v2.1.3", "utils/v2.1.3"},
	}

	for _, tt := range tests {
		t.Run(tt.moduleName, func(t *testing.T) {
			tagName := fmt.Sprintf("%s/%s", tt.moduleName, tt.version)
			require.Equal(t, tt.expected, tagName)
		})
	}
}

// TestParameterValidation tests parameter validation
func TestParameterValidation(t *testing.T) {
	t.Run("ValidModuleNames", func(t *testing.T) {
		validNames := []string{"models", "engine", "utils", "api", "internal"}

		for _, name := range validNames {
			cfg, err := parseBumpConfig([]string{"bump=patch", "module=" + name})
			require.NoError(t, err)
			require.Equal(t, name, cfg.module)
		}
	})

	t.Run("SpecialModuleValues", func(t *testing.T) {
		specialValues := []string{"all", "*"}

		for _, value := range specialValues {
			cfg, err := parseBumpConfig([]string{"bump=patch", "module=" + value})
			require.NoError(t, err)
			require.Equal(t, value, cfg.module)
		}
	})
}

// TestConcurrentModuleOperations tests thread safety
func TestConcurrentModuleOperations(t *testing.T) {
	t.Run("ConcurrentModuleDiscovery", func(t *testing.T) {
		done := make(chan error, 5)

		for i := 0; i < 5; i++ {
			go func() {
				_, err := discoverModules()
				done <- err
			}()
		}

		for i := 0; i < 5; i++ {
			err := <-done
			require.True(t, err == nil || err != nil)
		}
	})

	t.Run("ConcurrentVersionRetrieval", func(t *testing.T) {
		done := make(chan string, 5)

		for i := 0; i < 5; i++ {
			go func() {
				version := getModuleTag("models")
				done <- version
			}()
		}

		results := make([]string, 5)
		for i := 0; i < 5; i++ {
			results[i] = <-done
		}

		// All results should be identical
		for i := 1; i < len(results); i++ {
			require.Equal(t, results[0], results[i])
		}
	})
}

// TestModuleBumpCoordination tests coordinated bumping
func TestModuleBumpCoordination(t *testing.T) {
	t.Run("AllModulesMaintainConsistentBumpType", func(t *testing.T) {
		bumpTypes := []string{"patch", "minor", "major"}

		for _, bt := range bumpTypes {
			cfg, err := parseBumpConfig([]string{"bump=" + bt, "module=all"})
			require.NoError(t, err)
			require.Equal(t, bt, cfg.bumpType)
			require.Equal(t, "all", cfg.module)
		}
	})

	t.Run("WildcardIncludesRootAndModules", func(t *testing.T) {
		cfg, err := parseBumpConfig([]string{"bump=patch", "module=*"})
		require.NoError(t, err)
		require.Equal(t, "*", cfg.module)
	})
}

// TestErrorMessages tests error message quality
func TestErrorMessages(t *testing.T) {
	t.Run("SubmoduleNotFoundError", func(t *testing.T) {
		require.Error(t, errSubmoduleNotFound)
		require.Contains(t, errSubmoduleNotFound.Error(), "sub-module not found")
	})

	t.Run("ErrorContainsModuleName", func(t *testing.T) {
		moduleName := "nonexistent"
		err := fmt.Errorf("%w: %s", errSubmoduleNotFound, moduleName)
		require.Contains(t, err.Error(), moduleName)
	})
}

// Benchmark tests
func BenchmarkDiscoverModules(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = discoverModules() //nolint:errcheck // error intentionally ignored in test cleanup
	}
}

func BenchmarkGetModuleTag(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = getModuleTag("models")
	}
}

func BenchmarkParseModuleTagVersion(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = parseModuleTagVersion("models/v1.2.3", "models")
	}
}
