package mage

import (
	"os"
	"testing"

	"github.com/mrz1836/go-mage/pkg/mage/testutil"
	"github.com/stretchr/testify/suite"
)

// GenerateTestSuite defines the test suite for Generate functions
type GenerateTestSuite struct {
	suite.Suite

	env      *testutil.TestEnvironment
	generate Generate
}

// SetupTest runs before each test
func (ts *GenerateTestSuite) SetupTest() {
	ts.env = testutil.NewTestEnvironment(ts.T())
	ts.env.CreateGoMod("test/module")
	ts.generate = Generate{}
}

// TearDownTest runs after each test
func (ts *GenerateTestSuite) TearDownTest() {
	ts.env.Cleanup()
}

// TestGenerateDefault tests the Default method
func (ts *GenerateTestSuite) TestGenerateDefault() {
	ts.Run("no generate directives found", func() {
		// Create a Go file without generate directives
		ts.env.CreateFile("main.go", "package main\n\nfunc main() {}\n")

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.Default()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("generate directives found", func() {
		// Create a Go file with generate directives
		ts.env.CreateFile("generate.go", "package main\n\n//go:generate echo test\n\nfunc main() {}\n")

		// Mock successful go generate command
		ts.env.Runner.On("RunCmd", "go", []string{"generate", "-v"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.Default()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("generate with build tags", func() {
		// Set environment variable for build tags
		originalTags := os.Getenv("GO_BUILD_TAGS")
		defer func() {
			if err := os.Setenv("GO_BUILD_TAGS", originalTags); err != nil {
				ts.T().Logf("Failed to restore GO_BUILD_TAGS: %v", err)
			}
		}()
		if err := os.Setenv("GO_BUILD_TAGS", "test,dev"); err != nil {
			ts.T().Fatalf("Failed to set GO_BUILD_TAGS: %v", err)
		}

		// Create a Go file with generate directives
		ts.env.CreateFile("generate.go", "package main\n\n//go:generate echo test\n\nfunc main() {}\n")

		// Mock successful go generate command with tags
		ts.env.Runner.On("RunCmd", "go", []string{"generate", "-v", "-tags", "test,dev"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.Default()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGenerateAll tests the All method
func (ts *GenerateTestSuite) TestGenerateAll() {
	ts.Run("successful generate all", func() {
		// Mock successful go generate command for all packages
		ts.env.Runner.On("RunCmd", "go", []string{"generate", "-v", "./..."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.All()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("generate all with build tags", func() {
		// Set environment variable for build tags
		originalTags := os.Getenv("GO_BUILD_TAGS")
		defer func() {
			if err := os.Setenv("GO_BUILD_TAGS", originalTags); err != nil {
				ts.T().Logf("Failed to restore GO_BUILD_TAGS: %v", err)
			}
		}()
		if err := os.Setenv("GO_BUILD_TAGS", "integration"); err != nil {
			ts.T().Fatalf("Failed to set GO_BUILD_TAGS: %v", err)
		}

		// Mock successful go generate command with tags
		ts.env.Runner.On("RunCmd", "go", []string{"generate", "-v", "./...", "-tags", "integration"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.All()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGenerateMocks tests the Mocks method - simplified test
func (ts *GenerateTestSuite) TestGenerateMocks() {
	ts.Run("successful mock setup", func() {
		// Mock the mockgen installation and execution
		ts.env.Runner.On("RunCmd", "go", []string{"install", "github.com/golang/mock/mockgen@latest"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.Mocks()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGenerateClean tests the Clean method
func (ts *GenerateTestSuite) TestGenerateClean() {
	ts.Run("no generated files to clean", func() {
		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.Clean()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("clean generated files", func() {
		// Create some mock generated files
		ts.env.CreateFile("test.pb.go", "// Generated file")
		ts.env.CreateFile("test_gen.go", "// Generated file")

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.Clean()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("clean with custom patterns", func() {
		// Set custom patterns environment variable
		originalPatterns := os.Getenv("GENERATED_PATTERNS")
		defer func() {
			if err := os.Setenv("GENERATED_PATTERNS", originalPatterns); err != nil {
				ts.T().Logf("Failed to restore GENERATED_PATTERNS: %v", err)
			}
		}()
		if err := os.Setenv("GENERATED_PATTERNS", "*.custom.go,*_test_gen.go"); err != nil {
			ts.T().Fatalf("Failed to set GENERATED_PATTERNS: %v", err)
		}

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.Clean()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGenerateCode tests the Code method
func (ts *GenerateTestSuite) TestGenerateCode() {
	ts.Run("successful code generation", func() {
		// Mock successful go generate command
		ts.env.Runner.On("RunCmd", "go", []string{"generate", "./..."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.Code()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGenerateDocs tests the Docs method
func (ts *GenerateTestSuite) TestGenerateDocs() {
	ts.Run("successful docs generation", func() {
		// Mock successful echo command
		ts.env.Runner.On("RunCmd", "echo", []string{"Generating documentation"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.Docs()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGenerateSwagger tests the Swagger method
func (ts *GenerateTestSuite) TestGenerateSwagger() {
	ts.Run("successful Swagger generation", func() {
		// Mock successful echo command
		ts.env.Runner.On("RunCmd", "echo", []string{"Generating Swagger docs"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.Swagger()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGenerateOpenAPI tests the OpenAPI method
func (ts *GenerateTestSuite) TestGenerateOpenAPI() {
	ts.Run("successful OpenAPI generation", func() {
		// Mock successful echo command
		ts.env.Runner.On("RunCmd", "echo", []string{"Generating OpenAPI spec"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.OpenAPI()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGenerateGraphQL tests the GraphQL method
func (ts *GenerateTestSuite) TestGenerateGraphQL() {
	ts.Run("successful GraphQL generation", func() {
		// Mock successful echo command
		ts.env.Runner.On("RunCmd", "echo", []string{"Generating GraphQL code"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.GraphQL()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGenerateSQL tests the SQL method
func (ts *GenerateTestSuite) TestGenerateSQL() {
	ts.Run("successful SQL generation", func() {
		// Mock successful echo command
		ts.env.Runner.On("RunCmd", "echo", []string{"Generating SQL files"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.SQL()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGenerateWire tests the Wire method
func (ts *GenerateTestSuite) TestGenerateWire() {
	ts.Run("successful Wire generation", func() {
		// Mock successful echo command
		ts.env.Runner.On("RunCmd", "echo", []string{"Generating wire files"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.Wire()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGenerateConfig tests the Config method
func (ts *GenerateTestSuite) TestGenerateConfig() {
	ts.Run("successful Config generation", func() {
		// Mock successful echo command
		ts.env.Runner.On("RunCmd", "echo", []string{"Generating config files"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				return ts.generate.Config()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGenerateHelperFunctions tests the helper functions
func (ts *GenerateTestSuite) TestGenerateHelperFunctions() {
	ts.Run("checkForGenerateDirectives function", func() {
		// Create files with and without generate directives
		ts.env.CreateFile("normal.go", "package main\n\nfunc main() {}\n")
		ts.env.CreateFile("with_generate.go", "package main\n\n//go:generate echo test\n\nfunc main() {}\n")

		hasGenerate, files := checkForGenerateDirectives()

		ts.Require().True(hasGenerate)
		ts.Require().GreaterOrEqual(len(files), 1)
	})

	ts.Run("findInterfaces function", func() {
		// Test the findInterfaces function (simplified implementation)
		interfaces := findInterfaces()

		// Should return empty slice since it's a simplified implementation
		ts.Require().Empty(interfaces)
	})

	ts.Run("checkForGRPCService function", func() {
		// Test without any proto files
		hasService := checkForGRPCService()
		ts.Require().False(hasService)

		// Create a proto file with service
		ts.env.CreateFile("test.proto", "syntax = \"proto3\";\n\nservice TestService {\n  rpc Test() returns ();\n}\n")

		hasService = checkForGRPCService()
		ts.Require().True(hasService)
	})

	ts.Run("backupGeneratedFiles function", func() {
		// Test backup function
		backupGeneratedFiles()
		// No error to check since function doesn't return error
	})

	ts.Run("compareGeneratedFiles function", func() {
		// Test compare function
		different, err := compareGeneratedFiles("/tmp/test")
		ts.Require().NoError(err)
		ts.Require().Empty(different)
	})
}

// TestGenerateIntegration tests integration scenarios
func (ts *GenerateTestSuite) TestGenerateIntegration() {
	ts.Run("complete generation workflow", func() {
		// Create a Go file with generate directives
		ts.env.CreateFile("generate.go", "package main\n\n//go:generate echo test\n\nfunc main() {}\n")

		// Mock successful commands
		ts.env.Runner.On("RunCmd", "go", []string{"generate", "-v"}).Return(nil)
		ts.env.Runner.On("RunCmd", "go", []string{"generate", "-v", "./..."}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				// Run default generation
				if err := ts.generate.Default(); err != nil {
					return err
				}
				// Run all packages generation
				return ts.generate.All()
			},
		)

		ts.Require().NoError(err)
	})

	ts.Run("generation with environment variables", func() {
		// Set environment variables
		originalTags := os.Getenv("GO_BUILD_TAGS")
		originalPatterns := os.Getenv("GENERATED_PATTERNS")
		defer func() {
			if err := os.Setenv("GO_BUILD_TAGS", originalTags); err != nil {
				ts.T().Logf("Failed to restore GO_BUILD_TAGS: %v", err)
			}
			if err := os.Setenv("GENERATED_PATTERNS", originalPatterns); err != nil {
				ts.T().Logf("Failed to restore GENERATED_PATTERNS: %v", err)
			}
		}()
		if err := os.Setenv("GO_BUILD_TAGS", "test"); err != nil {
			ts.T().Fatalf("Failed to set GO_BUILD_TAGS: %v", err)
		}
		if err := os.Setenv("GENERATED_PATTERNS", "*.custom.go"); err != nil {
			ts.T().Fatalf("Failed to set GENERATED_PATTERNS: %v", err)
		}

		// Create test files
		ts.env.CreateFile("generate.go", "package main\n\n//go:generate echo test\n\nfunc main() {}\n")
		ts.env.CreateFile("test.custom.go", "// Custom generated file")

		// Mock commands
		ts.env.Runner.On("RunCmd", "go", []string{"generate", "-v", "-tags", "test"}).Return(nil)

		err := ts.env.WithMockRunner(
			func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
			func() interface{} { return GetRunner() },
			func() error {
				// Run generation with tags
				if err := ts.generate.Default(); err != nil {
					return err
				}
				// Clean with custom patterns
				return ts.generate.Clean()
			},
		)

		ts.Require().NoError(err)
	})
}

// TestGenerateTestSuite runs the test suite
func TestGenerateTestSuite(t *testing.T) {
	suite.Run(t, new(GenerateTestSuite))
}
