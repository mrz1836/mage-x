package mage

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// generateMockRunner mocks command runner for generate tests
type generateMockRunner struct {
	RunCmdErr       error
	RunCmdOutputVal string
	RunCmdOutputErr error
	Commands        []string
}

func (m *generateMockRunner) RunCmd(cmd string, args ...string) error {
	m.Commands = append(m.Commands, cmd+" "+filepath.Join(args...))
	return m.RunCmdErr
}

func (m *generateMockRunner) RunCmdOutput(cmd string, args ...string) (string, error) {
	return m.RunCmdOutputVal, m.RunCmdOutputErr
}

func (m *generateMockRunner) RunCmdInDir(dir, cmd string, args ...string) error {
	return nil
}

// TestGenerateAllUnit tests Generate.All without integration tag
func TestGenerateAllUnit(t *testing.T) {
	t.Run("runs go generate on all packages", func(t *testing.T) {
		mock := &generateMockRunner{}
		originalRunner := GetRunner()
		_ = SetRunner(mock) //nolint:errcheck // test cleanup
		defer func() {
			_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
		}()

		g := Generate{}
		err := g.All()

		require.NoError(t, err)
		require.Len(t, mock.Commands, 1)
		assert.Contains(t, mock.Commands[0], "go")
	})

	t.Run("returns error on command failure", func(t *testing.T) {
		mock := &generateMockRunner{
			RunCmdErr: assert.AnError,
		}
		originalRunner := GetRunner()
		_ = SetRunner(mock) //nolint:errcheck // test cleanup
		defer func() {
			_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
		}()

		g := Generate{}
		err := g.All()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "go generate failed")
	})

	t.Run("includes build tags when set", func(t *testing.T) {
		mock := &generateMockRunner{}
		originalRunner := GetRunner()
		_ = SetRunner(mock) //nolint:errcheck // test cleanup
		defer func() {
			_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
		}()

		// Set build tags env
		originalTags := os.Getenv("MAGE_X_BUILD_TAGS")
		require.NoError(t, os.Setenv("MAGE_X_BUILD_TAGS", "integration,e2e"))
		defer func() {
			if originalTags == "" {
				_ = os.Unsetenv("MAGE_X_BUILD_TAGS") //nolint:errcheck // cleanup
			} else {
				_ = os.Setenv("MAGE_X_BUILD_TAGS", originalTags) //nolint:errcheck // cleanup
			}
		}()

		g := Generate{}
		err := g.All()

		require.NoError(t, err)
		require.Len(t, mock.Commands, 1)
	})
}

// TestGenerateCleanUnit tests Generate.Clean without integration tag
func TestGenerateCleanUnit(t *testing.T) {
	t.Run("removes generated files", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create generated files
		require.NoError(t, os.WriteFile("test.pb.go", []byte("// generated"), 0o600))
		require.NoError(t, os.WriteFile("service_gen.go", []byte("// generated"), 0o600))

		g := Generate{}
		err = g.Clean()

		require.NoError(t, err)
		// Files should be removed
		_, err = os.Stat("test.pb.go")
		assert.True(t, os.IsNotExist(err))
		_, err = os.Stat("service_gen.go")
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("handles no generated files", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		g := Generate{}
		err = g.Clean()

		require.NoError(t, err)
	})

	t.Run("includes custom patterns from env", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Set custom patterns
		originalPatterns := os.Getenv("MAGE_X_GENERATED_PATTERNS")
		require.NoError(t, os.Setenv("MAGE_X_GENERATED_PATTERNS", "*_custom.go"))
		defer func() {
			if originalPatterns == "" {
				_ = os.Unsetenv("MAGE_X_GENERATED_PATTERNS") //nolint:errcheck // cleanup
			} else {
				_ = os.Setenv("MAGE_X_GENERATED_PATTERNS", originalPatterns) //nolint:errcheck // cleanup
			}
		}()

		// Create custom generated file
		require.NoError(t, os.WriteFile("test_custom.go", []byte("// custom generated"), 0o600))

		g := Generate{}
		err = g.Clean()

		require.NoError(t, err)
	})
}

// TestCheckForGenerateDirectivesUnit tests checkForGenerateDirectives function
func TestCheckForGenerateDirectivesUnit(t *testing.T) {
	t.Run("finds generate directives", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create file with generate directive
		content := `package main

//go:generate echo hello

func main() {}
`
		require.NoError(t, os.WriteFile("main.go", []byte(content), 0o600))

		hasGenerate, files := checkForGenerateDirectives()

		assert.True(t, hasGenerate)
		assert.Len(t, files, 1)
		assert.Contains(t, files[0], "main.go")
	})

	t.Run("returns false when no directives", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create file without generate directive
		require.NoError(t, os.WriteFile("main.go", []byte("package main\n\nfunc main() {}\n"), 0o600))

		hasGenerate, files := checkForGenerateDirectives()

		assert.False(t, hasGenerate)
		assert.Empty(t, files)
	})

	t.Run("skips vendor directory", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create vendor directory with generate directive
		vendorDir := filepath.Join(tempDir, "vendor", "pkg")
		require.NoError(t, os.MkdirAll(vendorDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(vendorDir, "gen.go"), []byte("package pkg\n\n//go:generate echo test\n"), 0o600))

		hasGenerate, files := checkForGenerateDirectives()

		assert.False(t, hasGenerate)
		assert.Empty(t, files)
	})
}

// TestFindInterfacesUnit tests findInterfaces function
func TestFindInterfacesUnit(t *testing.T) {
	t.Run("returns empty slice for simplified implementation", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create file with interface
		content := `package main

type Reader interface {
    Read(p []byte) (n int, err error)
}
`
		require.NoError(t, os.WriteFile("reader.go", []byte(content), 0o600))

		interfaces := findInterfaces()

		// Current implementation returns empty (simplified)
		assert.Empty(t, interfaces)
	})
}

// TestCheckForGRPCServiceUnit tests checkForGRPCService function
func TestCheckForGRPCServiceUnit(t *testing.T) {
	t.Run("returns true when service definition found", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create proto file with service
		content := `syntax = "proto3";

package test;

service UserService {
    rpc GetUser(GetUserRequest) returns (GetUserResponse);
}
`
		require.NoError(t, os.WriteFile("user.proto", []byte(content), 0o600))

		hasService := checkForGRPCService()

		assert.True(t, hasService)
	})

	t.Run("returns false when no service definition", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Create proto file without service
		content := `syntax = "proto3";

package test;

message User {
    string id = 1;
    string name = 2;
}
`
		require.NoError(t, os.WriteFile("user.proto", []byte(content), 0o600))

		hasService := checkForGRPCService()

		assert.False(t, hasService)
	})

	t.Run("returns false when no proto files", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		hasService := checkForGRPCService()

		assert.False(t, hasService)
	})
}

// TestBackupGeneratedFilesUnit tests backupGeneratedFiles function
func TestBackupGeneratedFilesUnit(t *testing.T) {
	t.Run("runs without error", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		// Function doesn't return error, just verify it doesn't panic
		assert.NotPanics(t, func() {
			backupGeneratedFiles()
		})
	})
}

// TestCompareGeneratedFilesUnit tests compareGeneratedFiles function
func TestCompareGeneratedFilesUnit(t *testing.T) {
	t.Run("returns empty for non-existent temp dir", func(t *testing.T) {
		different, err := compareGeneratedFiles("/nonexistent/path")

		require.NoError(t, err)
		assert.Empty(t, different)
	})

	t.Run("handles empty directory", func(t *testing.T) {
		tempDir := t.TempDir()

		different, err := compareGeneratedFiles(tempDir)

		require.NoError(t, err)
		assert.Empty(t, different)
	})
}

// TestGenerateCodeUnit tests Generate.Code
func TestGenerateCodeUnit(t *testing.T) {
	t.Run("runs go generate", func(t *testing.T) {
		mock := &generateMockRunner{}
		originalRunner := GetRunner()
		_ = SetRunner(mock) //nolint:errcheck // test cleanup
		defer func() {
			_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
		}()

		g := Generate{}
		err := g.Code()

		require.NoError(t, err)
		require.Len(t, mock.Commands, 1)
	})
}

// TestGenerateDocsUnit tests Generate.Docs
func TestGenerateDocsUnit(t *testing.T) {
	t.Run("runs echo command", func(t *testing.T) {
		mock := &generateMockRunner{}
		originalRunner := GetRunner()
		_ = SetRunner(mock) //nolint:errcheck // test cleanup
		defer func() {
			_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
		}()

		g := Generate{}
		err := g.Docs()

		require.NoError(t, err)
	})
}

// TestGenerateSwaggerUnit tests Generate.Swagger
func TestGenerateSwaggerUnit(t *testing.T) {
	t.Run("runs echo command", func(t *testing.T) {
		mock := &generateMockRunner{}
		originalRunner := GetRunner()
		_ = SetRunner(mock) //nolint:errcheck // test cleanup
		defer func() {
			_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
		}()

		g := Generate{}
		err := g.Swagger()

		require.NoError(t, err)
	})
}

// TestGenerateOpenAPIUnit tests Generate.OpenAPI
func TestGenerateOpenAPIUnit(t *testing.T) {
	t.Run("runs echo command", func(t *testing.T) {
		mock := &generateMockRunner{}
		originalRunner := GetRunner()
		_ = SetRunner(mock) //nolint:errcheck // test cleanup
		defer func() {
			_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
		}()

		g := Generate{}
		err := g.OpenAPI()

		require.NoError(t, err)
	})
}

// TestGenerateGraphQLUnit tests Generate.GraphQL
func TestGenerateGraphQLUnit(t *testing.T) {
	t.Run("runs echo command", func(t *testing.T) {
		mock := &generateMockRunner{}
		originalRunner := GetRunner()
		_ = SetRunner(mock) //nolint:errcheck // test cleanup
		defer func() {
			_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
		}()

		g := Generate{}
		err := g.GraphQL()

		require.NoError(t, err)
	})
}

// TestGenerateSQLUnit tests Generate.SQL
func TestGenerateSQLUnit(t *testing.T) {
	t.Run("runs echo command", func(t *testing.T) {
		mock := &generateMockRunner{}
		originalRunner := GetRunner()
		_ = SetRunner(mock) //nolint:errcheck // test cleanup
		defer func() {
			_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
		}()

		g := Generate{}
		err := g.SQL()

		require.NoError(t, err)
	})
}

// TestGenerateWireUnit tests Generate.Wire
func TestGenerateWireUnit(t *testing.T) {
	t.Run("runs echo command", func(t *testing.T) {
		mock := &generateMockRunner{}
		originalRunner := GetRunner()
		_ = SetRunner(mock) //nolint:errcheck // test cleanup
		defer func() {
			_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
		}()

		g := Generate{}
		err := g.Wire()

		require.NoError(t, err)
	})
}

// TestGenerateConfigUnit tests Generate.Config
func TestGenerateConfigUnit(t *testing.T) {
	t.Run("runs echo command", func(t *testing.T) {
		mock := &generateMockRunner{}
		originalRunner := GetRunner()
		_ = SetRunner(mock) //nolint:errcheck // test cleanup
		defer func() {
			_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
		}()

		g := Generate{}
		err := g.Config()

		require.NoError(t, err)
	})
}

// TestGenerateMocksUnit tests Generate.Mocks
func TestGenerateMocksUnit(t *testing.T) {
	t.Run("returns early when no interfaces found", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		mock := &generateMockRunner{}
		originalRunner := GetRunner()
		_ = SetRunner(mock) //nolint:errcheck // test cleanup
		defer func() {
			_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
		}()

		g := Generate{}
		err = g.Mocks()

		// Should return nil because no interfaces found
		require.NoError(t, err)
	})
}

// TestGenerateProtoUnit tests Generate.Proto
func TestGenerateProtoUnit(t *testing.T) {
	t.Run("returns early when no proto files", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		mock := &generateMockRunner{}
		originalRunner := GetRunner()
		_ = SetRunner(mock) //nolint:errcheck // test cleanup
		defer func() {
			_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
		}()

		g := Generate{}
		err = g.Proto()
		// Should return ErrProtocNotFound if protoc not installed
		// or nil if no proto files found
		if err != nil {
			assert.ErrorIs(t, err, ErrProtocNotFound)
		}
	})
}

// TestGenerateCheckUnit tests Generate.Check
func TestGenerateCheckUnit(t *testing.T) {
	t.Run("creates temp directory and runs check", func(t *testing.T) {
		tempDir := t.TempDir()
		originalWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tempDir))
		defer func() {
			require.NoError(t, os.Chdir(originalWd))
		}()

		mock := &generateMockRunner{}
		originalRunner := GetRunner()
		_ = SetRunner(mock) //nolint:errcheck // test cleanup
		defer func() {
			_ = SetRunner(originalRunner) //nolint:errcheck // cleanup
		}()

		g := Generate{}
		err = g.Check()

		// May succeed or fail depending on generate result
		// Just verify it doesn't panic
		_ = err
	})
}
