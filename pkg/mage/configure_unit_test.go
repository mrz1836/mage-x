package mage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateConfiguration tests the validateConfiguration function
func TestValidateConfiguration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *Config
		wantErr error
	}{
		{
			name: "valid config with all required fields",
			config: &Config{
				Project: ProjectConfig{
					Name:   "test-project",
					Binary: "test-bin",
					Module: "github.com/test/project",
				},
			},
			wantErr: nil,
		},
		{
			name: "valid config with minimal fields",
			config: &Config{
				Project: ProjectConfig{
					Name:   "a",
					Binary: "b",
					Module: "c",
				},
			},
			wantErr: nil,
		},
		{
			name: "empty project name",
			config: &Config{
				Project: ProjectConfig{
					Name:   "",
					Binary: "test-bin",
					Module: "github.com/test/project",
				},
			},
			wantErr: ErrProjectNameRequired,
		},
		{
			name: "empty binary name",
			config: &Config{
				Project: ProjectConfig{
					Name:   "test-project",
					Binary: "",
					Module: "github.com/test/project",
				},
			},
			wantErr: ErrBinaryNameRequired,
		},
		{
			name: "empty module path",
			config: &Config{
				Project: ProjectConfig{
					Name:   "test-project",
					Binary: "test-bin",
					Module: "",
				},
			},
			wantErr: ErrModulePathRequired,
		},
		{
			name: "all required fields empty",
			config: &Config{
				Project: ProjectConfig{
					Name:   "",
					Binary: "",
					Module: "",
				},
			},
			wantErr: ErrProjectNameRequired, // First validation fails
		},
		{
			name: "whitespace-only name",
			config: &Config{
				Project: ProjectConfig{
					Name:   "   ",
					Binary: "test-bin",
					Module: "github.com/test/project",
				},
			},
			// Note: current implementation doesn't trim whitespace, so this passes
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateConfiguration(tt.config)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGenerateConfigurationSchema tests the generateConfigurationSchema function
func TestGenerateConfigurationSchema(t *testing.T) {
	t.Parallel()

	schema := generateConfigurationSchema()

	// Verify it's valid JSON
	var parsed map[string]interface{}
	err := json.Unmarshal([]byte(schema), &parsed)
	require.NoError(t, err, "Schema should be valid JSON")

	// Verify schema metadata
	assert.Equal(t, "http://json-schema.org/draft-07/schema#", parsed["$schema"], "Should have correct schema version")
	assert.Equal(t, "MAGE-X Configuration", parsed["title"], "Should have correct title")
	assert.Equal(t, "object", parsed["type"], "Root type should be object")

	// Verify required fields
	required, ok := parsed["required"].([]interface{})
	require.True(t, ok, "Should have required array")
	assert.Contains(t, required, "project")
	assert.Contains(t, required, "build")
	assert.Contains(t, required, "test")

	// Verify properties exist
	properties, ok := parsed["properties"].(map[string]interface{})
	require.True(t, ok, "Should have properties object")

	// Check project properties
	project, ok := properties["project"].(map[string]interface{})
	require.True(t, ok, "Should have project property")
	assert.Equal(t, "object", project["type"])

	projectProps, ok := project["properties"].(map[string]interface{})
	require.True(t, ok, "Project should have properties")
	assert.Contains(t, projectProps, "name")
	assert.Contains(t, projectProps, "binary")
	assert.Contains(t, projectProps, "module")

	// Check project required fields
	projectRequired, ok := project["required"].([]interface{})
	require.True(t, ok, "Project should have required fields")
	assert.Contains(t, projectRequired, "name")
	assert.Contains(t, projectRequired, "binary")
	assert.Contains(t, projectRequired, "module")

	// Check build properties
	build, ok := properties["build"].(map[string]interface{})
	require.True(t, ok, "Should have build property")
	assert.Equal(t, "object", build["type"])

	// Check test properties
	test, ok := properties["test"].(map[string]interface{})
	require.True(t, ok, "Should have test property")
	assert.Equal(t, "object", test["type"])
}

// TestMarshalJSON tests the marshalJSON function
func TestMarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "marshal config struct",
			input: &Config{
				Project: ProjectConfig{
					Name:   "test-project",
					Binary: "test-bin",
					Module: "github.com/test/project",
				},
				Build: BuildConfig{
					Output:   "bin",
					Parallel: 4,
				},
				Test: TestConfig{
					Timeout: "5m",
					Race:    true,
				},
			},
			wantErr: false,
		},
		{
			name:    "marshal empty config",
			input:   &Config{},
			wantErr: false,
		},
		{
			name: "marshal simple map",
			input: map[string]string{
				"key": "value",
			},
			wantErr: false,
		},
		{
			name:    "marshal nil",
			input:   nil,
			wantErr: false, // YAML marshal handles nil gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			data, err := marshalJSON(tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, data)
			}
		})
	}
}

// TestUnmarshalJSON tests the unmarshalJSON function
func TestUnmarshalJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    []byte
		target  interface{}
		wantErr bool
	}{
		{
			name: "unmarshal valid YAML config",
			data: []byte(`
project:
  name: test-project
  binary: test-bin
  module: github.com/test/project
`),
			target:  &Config{},
			wantErr: false,
		},
		{
			name:    "unmarshal empty data",
			data:    []byte{},
			target:  &Config{},
			wantErr: false, // YAML unmarshal handles empty gracefully
		},
		{
			name:    "unmarshal JSON format",
			data:    []byte(`{"project": {"name": "test", "binary": "bin", "module": "mod"}}`),
			target:  &Config{},
			wantErr: false, // YAML can parse JSON
		},
		{
			name:    "unmarshal invalid YAML",
			data:    []byte("invalid: yaml: content: [broken"),
			target:  &Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := unmarshalJSON(tt.data, tt.target)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestMarshalUnmarshalRoundTrip tests that data survives marshal/unmarshal round-trip
func TestMarshalUnmarshalRoundTrip(t *testing.T) {
	t.Parallel()

	original := &Config{
		Project: ProjectConfig{
			Name:        "round-trip-test",
			Binary:      "test-binary",
			Module:      "github.com/test/roundtrip",
			Description: "A test project",
			Version:     "1.0.0",
			GitDomain:   "github.com",
			RepoOwner:   "test",
			RepoName:    "roundtrip",
		},
		Build: BuildConfig{
			Output:    "bin",
			Parallel:  8,
			TrimPath:  true,
			Verbose:   false,
			Platforms: []string{"linux/amd64", "darwin/amd64"},
			Tags:      []string{"production"},
		},
		Test: TestConfig{
			Timeout:   "10m",
			Race:      true,
			Cover:     true,
			CoverMode: "atomic",
			Parallel:  4,
			Verbose:   true,
		},
	}

	// Marshal
	data, err := marshalJSON(original)
	require.NoError(t, err, "Marshal should succeed")

	// Unmarshal
	var restored Config
	err = unmarshalJSON(data, &restored)
	require.NoError(t, err, "Unmarshal should succeed")

	// Compare key fields
	assert.Equal(t, original.Project.Name, restored.Project.Name)
	assert.Equal(t, original.Project.Binary, restored.Project.Binary)
	assert.Equal(t, original.Project.Module, restored.Project.Module)
	assert.Equal(t, original.Build.Output, restored.Build.Output)
	assert.Equal(t, original.Build.Parallel, restored.Build.Parallel)
	assert.Equal(t, original.Test.Timeout, restored.Test.Timeout)
	assert.Equal(t, original.Test.Race, restored.Test.Race)
}

// TestConfigureInit tests the Configure.Init method
func TestConfigureInit(t *testing.T) {
	// Not parallel - uses temp directory and changes cwd

	t.Run("init creates config file in empty directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			require.NoError(t, os.Chdir(originalDir))
		}()

		require.NoError(t, os.Chdir(tempDir))
		TestResetConfig()
		defer TestResetConfig()

		c := Configure{}
		err = c.Init()
		require.NoError(t, err)

		// Verify a config file was created
		configFound := false
		for _, cf := range MageConfigFiles() {
			if _, statErr := os.Stat(cf); statErr == nil {
				configFound = true
				break
			}
		}
		assert.True(t, configFound, "Config file should be created")
	})

	t.Run("init fails when config already exists", func(t *testing.T) {
		tempDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			require.NoError(t, os.Chdir(originalDir))
		}()

		require.NoError(t, os.Chdir(tempDir))
		TestResetConfig()
		defer TestResetConfig()

		// Create existing config file
		configFile := MageConfigFiles()[0]
		err = os.WriteFile(configFile, []byte("project:\n  name: existing\n"), os.FileMode(0o600))
		require.NoError(t, err)

		c := Configure{}
		err = c.Init()
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrConfigFileExists)
	})
}

// TestConfigureValidate tests the Configure.Validate method
func TestConfigureValidate(t *testing.T) {
	// Not parallel - modifies global config

	t.Run("validate succeeds with valid config", func(t *testing.T) {
		TestSetConfig(&Config{
			Project: ProjectConfig{
				Name:   "valid-project",
				Binary: "valid-bin",
				Module: "github.com/valid/module",
			},
		})
		defer TestResetConfig()

		c := Configure{}
		err := c.Validate()
		assert.NoError(t, err)
	})

	t.Run("validate fails with invalid config", func(t *testing.T) {
		TestSetConfig(&Config{
			Project: ProjectConfig{
				Name:   "", // Invalid: empty name
				Binary: "valid-bin",
				Module: "github.com/valid/module",
			},
		})
		defer TestResetConfig()

		c := Configure{}
		err := c.Validate()
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrProjectNameRequired)
	})
}

// TestConfigureExport tests the Configure.Export method
func TestConfigureExport(t *testing.T) {
	// Not parallel - modifies global config and environment

	t.Run("export to stdout in yaml format", func(t *testing.T) {
		TestSetConfig(&Config{
			Project: ProjectConfig{
				Name:   "export-test",
				Binary: "export-bin",
				Module: "github.com/export/test",
			},
		})
		defer TestResetConfig()

		// Set FORMAT env var
		t.Setenv("FORMAT", "yaml")
		t.Setenv("OUTPUT", "")

		c := Configure{}
		err := c.Export()
		assert.NoError(t, err)
	})

	t.Run("export to stdout in json format", func(t *testing.T) {
		TestSetConfig(&Config{
			Project: ProjectConfig{
				Name:   "export-test",
				Binary: "export-bin",
				Module: "github.com/export/test",
			},
		})
		defer TestResetConfig()

		t.Setenv("FORMAT", "json")
		t.Setenv("OUTPUT", "")

		c := Configure{}
		err := c.Export()
		assert.NoError(t, err)
	})

	t.Run("export fails with unsupported format", func(t *testing.T) {
		TestSetConfig(&Config{
			Project: ProjectConfig{
				Name:   "export-test",
				Binary: "export-bin",
				Module: "github.com/export/test",
			},
		})
		defer TestResetConfig()

		t.Setenv("FORMAT", "xml") // Unsupported
		t.Setenv("OUTPUT", "")

		c := Configure{}
		err := c.Export()
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUnsupportedFormat)
	})

	t.Run("export to file", func(t *testing.T) {
		TestSetConfig(&Config{
			Project: ProjectConfig{
				Name:   "export-test",
				Binary: "export-bin",
				Module: "github.com/export/test",
			},
		})
		defer TestResetConfig()

		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "config")

		t.Setenv("FORMAT", "yaml")
		t.Setenv("OUTPUT", outputFile)

		c := Configure{}
		err := c.Export()
		require.NoError(t, err)

		// Verify file was created with .yaml extension
		_, err = os.Stat(outputFile + ".yaml")
		require.NoError(t, err)
	})
}

// TestConfigureImport tests the Configure.Import method
func TestConfigureImport(t *testing.T) {
	// Not parallel - modifies global config and environment

	t.Run("import fails without FILE env var", func(t *testing.T) {
		TestResetConfig()
		defer TestResetConfig()

		t.Setenv("FILE", "")

		c := Configure{}
		err := c.Import()
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrFileEnvRequired)
	})

	t.Run("import fails with nonexistent file", func(t *testing.T) {
		TestResetConfig()
		defer TestResetConfig()

		t.Setenv("FILE", "/nonexistent/path/config.yaml")

		c := Configure{}
		err := c.Import()
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrImportFileNotFound)
	})

	t.Run("import succeeds with valid yaml file", func(t *testing.T) {
		tempDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			require.NoError(t, os.Chdir(originalDir))
		}()
		require.NoError(t, os.Chdir(tempDir))

		TestResetConfig()
		defer TestResetConfig()

		// Create valid config file
		configContent := `
project:
  name: imported-project
  binary: imported-bin
  module: github.com/imported/module
build:
  output: bin
test:
  timeout: 5m
`
		configFile := filepath.Join(tempDir, "import-config.yaml")
		err = os.WriteFile(configFile, []byte(configContent), os.FileMode(0o600))
		require.NoError(t, err)

		t.Setenv("FILE", configFile)

		c := Configure{}
		err = c.Import()
		assert.NoError(t, err)
	})

	t.Run("import fails with invalid config", func(t *testing.T) {
		tempDir := t.TempDir()
		originalDir, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			require.NoError(t, os.Chdir(originalDir))
		}()
		require.NoError(t, os.Chdir(tempDir))

		TestResetConfig()
		defer TestResetConfig()

		// Create config file with missing required fields
		configContent := `
project:
  name: ""
  binary: ""
  module: ""
`
		configFile := filepath.Join(tempDir, "invalid-config.yaml")
		err = os.WriteFile(configFile, []byte(configContent), os.FileMode(0o600))
		require.NoError(t, err)

		t.Setenv("FILE", configFile)

		c := Configure{}
		err = c.Import()
		assert.Error(t, err)
	})

	t.Run("import fails with unsupported file format", func(t *testing.T) {
		tempDir := t.TempDir()

		TestResetConfig()
		defer TestResetConfig()

		// Create file with unsupported extension
		configFile := filepath.Join(tempDir, "config.xml")
		err := os.WriteFile(configFile, []byte("<config></config>"), os.FileMode(0o600))
		require.NoError(t, err)

		t.Setenv("FILE", configFile)

		c := Configure{}
		err = c.Import()
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrUnsupportedFileFormat)
	})
}

// TestConfigureSchema tests the Configure.Schema method
func TestConfigureSchema(t *testing.T) {
	// Not parallel - modifies environment

	t.Run("schema outputs to stdout", func(t *testing.T) {
		t.Setenv("OUTPUT", "")

		c := Configure{}
		err := c.Schema()
		assert.NoError(t, err)
	})

	t.Run("schema outputs to file", func(t *testing.T) {
		tempDir := t.TempDir()
		outputFile := filepath.Join(tempDir, "schema.json")

		t.Setenv("OUTPUT", outputFile)

		c := Configure{}
		err := c.Schema()
		require.NoError(t, err)

		// Verify file was created
		data, err := os.ReadFile(outputFile) // #nosec G304 -- test file path is controlled
		require.NoError(t, err)

		// Verify it's valid JSON schema
		var schema map[string]interface{}
		err = json.Unmarshal(data, &schema)
		require.NoError(t, err)
		assert.Equal(t, "http://json-schema.org/draft-07/schema#", schema["$schema"])
	})
}

// TestConfigureShow tests the Configure.Show method
func TestConfigureShow(t *testing.T) {
	// Not parallel - modifies global config

	t.Run("show displays config without error", func(t *testing.T) {
		TestSetConfig(&Config{
			Project: ProjectConfig{
				Name:      "show-test",
				Binary:    "show-bin",
				Module:    "github.com/show/test",
				GitDomain: "github.com",
				RepoOwner: "show",
				RepoName:  "test",
			},
			Build: BuildConfig{
				Output:    "bin",
				Parallel:  4,
				Platforms: []string{"linux/amd64"},
				TrimPath:  true,
			},
			Test: TestConfig{
				Parallel:  4,
				Timeout:   "5m",
				CoverMode: "atomic",
				Race:      true,
			},
		})
		defer TestResetConfig()

		c := Configure{}
		err := c.Show()
		assert.NoError(t, err)
	})

	t.Run("show works with minimal config", func(t *testing.T) {
		TestSetConfig(&Config{
			Project: ProjectConfig{
				Name:   "minimal",
				Binary: "bin",
				Module: "mod",
			},
		})
		defer TestResetConfig()

		c := Configure{}
		err := c.Show()
		assert.NoError(t, err)
	})
}

// TestStaticErrors tests that static error variables are properly defined
func TestStaticErrors(t *testing.T) {
	t.Parallel()

	// Verify all errors are non-nil and have expected messages
	errors := map[string]error{
		"ErrConfigFileExists":      ErrConfigFileExists,
		"ErrUnsupportedFormat":     ErrUnsupportedFormat,
		"ErrFileEnvRequired":       ErrFileEnvRequired,
		"ErrImportFileNotFound":    ErrImportFileNotFound,
		"ErrUnsupportedFileFormat": ErrUnsupportedFileFormat,
		"ErrProjectNameRequired":   ErrProjectNameRequired,
		"ErrBinaryNameRequired":    ErrBinaryNameRequired,
		"ErrModulePathRequired":    ErrModulePathRequired,
	}

	for name, err := range errors {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			require.Error(t, err, "%s should not be nil", name)
			assert.NotEmpty(t, err.Error(), "%s should have an error message", name)
		})
	}
}

// Benchmark tests for performance validation

func BenchmarkValidateConfiguration(b *testing.B) {
	config := &Config{
		Project: ProjectConfig{
			Name:   "benchmark-project",
			Binary: "benchmark-bin",
			Module: "github.com/benchmark/module",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := validateConfiguration(config); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateConfigurationSchema(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateConfigurationSchema()
	}
}

func BenchmarkMarshalJSON(b *testing.B) {
	config := &Config{
		Project: ProjectConfig{
			Name:   "benchmark-project",
			Binary: "benchmark-bin",
			Module: "github.com/benchmark/module",
		},
		Build: BuildConfig{
			Output:   "bin",
			Parallel: 4,
		},
		Test: TestConfig{
			Timeout: "5m",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := marshalJSON(config); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshalJSON(b *testing.B) {
	data := []byte(`
project:
  name: benchmark-project
  binary: benchmark-bin
  module: github.com/benchmark/module
build:
  output: bin
  parallel: 4
test:
  timeout: 5m
`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var config Config
		if err := unmarshalJSON(data, &config); err != nil {
			b.Fatal(err)
		}
	}
}
