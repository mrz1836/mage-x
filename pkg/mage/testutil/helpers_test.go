package testutil_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
)

// Static errors for testing
var (
	errSetterTest  = errors.New("setter error")
	errRestoreTest = errors.New("restore error")
)

// TestingInterfaceMock implements the TestingInterface for testing
type TestingInterfaceMock struct {
	tempDir      string
	fatalfCalled bool
	fatalfMsg    string
	fatalfArgs   []interface{}
}

func (m *TestingInterfaceMock) TempDir() string {
	if m.tempDir == "" {
		var err error
		m.tempDir, err = os.MkdirTemp("", "testutil_test")
		if err != nil {
			panic(err)
		}
	}
	return m.tempDir
}

func (m *TestingInterfaceMock) Helper() {
	// Mock implementation - does nothing
}

func (m *TestingInterfaceMock) Fatalf(format string, args ...interface{}) {
	m.fatalfCalled = true
	m.fatalfMsg = format
	m.fatalfArgs = args
	// Don't actually call panic/exit in tests
}

func (m *TestingInterfaceMock) Cleanup() {
	if m.tempDir != "" {
		_ = os.RemoveAll(m.tempDir) //nolint:errcheck // In test cleanup, we ignore the error as there's no meaningful recovery
	}
}

// TestUtilHelpersTestSuite tests the testutil helpers
type TestUtilHelpersTestSuite struct {
	suite.Suite

	mockT *TestingInterfaceMock
}

func (ts *TestUtilHelpersTestSuite) SetupTest() {
	ts.mockT = &TestingInterfaceMock{}
}

func (ts *TestUtilHelpersTestSuite) TearDownTest() {
	if ts.mockT != nil {
		ts.mockT.Cleanup()
	}
}

func (ts *TestUtilHelpersTestSuite) TestNewTestEnvironment() {
	// Save original directory
	origDir, err := os.Getwd()
	ts.Require().NoError(err)

	// Create separate mock for this test so cleanup doesn't interfere with other tests
	testMock := &TestingInterfaceMock{}

	env := testutil.NewTestEnvironment(testMock)
	ts.Require().NotNil(env)

	// Should have changed to temp directory
	currentDir, err := os.Getwd()
	ts.Require().NoError(err)

	// Resolve any symlinks for comparison (macOS has /var -> /private/var)
	expectedDir, err := filepath.EvalSymlinks(env.TempDir)
	ts.Require().NoError(err)
	actualDir, err := filepath.EvalSymlinks(currentDir)
	ts.Require().NoError(err)
	ts.Require().Equal(expectedDir, actualDir)

	// Should have original directory stored
	ts.Require().Equal(origDir, env.OrigDir)

	// Should have mock runner and builder
	ts.Require().NotNil(env.Runner)
	ts.Require().NotNil(env.Builder)
	ts.Require().NotNil(env.CommandMatcher)

	// Cleanup should restore original directory
	env.Cleanup()
	currentDir, err = os.Getwd()
	ts.Require().NoError(err)
	ts.Require().Equal(origDir, currentDir)

	// Clean up the test mock temp dir
	testMock.Cleanup()
}

func (ts *TestUtilHelpersTestSuite) TestNewTestEnvironmentDirectoryFailure() {
	// Create a mock that returns an invalid temp directory
	mockT := &TestingInterfaceMock{
		tempDir: "/nonexistent/path/that/cannot/be/created",
	}

	// This should trigger a Fatalf call due to chdir failure
	_ = testutil.NewTestEnvironment(mockT)
	ts.Require().True(mockT.fatalfCalled)
	ts.Require().Contains(mockT.fatalfMsg, "Failed to change directory")
}

func (ts *TestUtilHelpersTestSuite) TestWithMockRunner() {
	env := testutil.NewTestEnvironment(ts.mockT)
	defer env.Cleanup()

	// Create a simple runner getter/setter for testing
	var currentRunner interface{} = "original"

	getter := func() interface{} {
		return currentRunner
	}

	setter := func(runner interface{}) error {
		currentRunner = runner
		return nil
	}

	executed := false
	err := env.WithMockRunner(setter, getter, func() error {
		// Verify mock runner is set
		ts.Require().Equal(env.Runner, currentRunner)
		executed = true
		return nil
	})

	ts.Require().NoError(err)
	ts.Require().True(executed)

	// Verify original runner is restored
	ts.Require().Equal("original", currentRunner)
}

func (ts *TestUtilHelpersTestSuite) TestWithMockRunnerSetterError() {
	env := testutil.NewTestEnvironment(ts.mockT)
	defer env.Cleanup()

	setter := func(runner interface{}) error {
		return errSetterTest
	}

	getter := func() interface{} {
		return "original"
	}

	err := env.WithMockRunner(setter, getter, func() error {
		return nil
	})

	ts.Require().Error(err)
	ts.Require().Contains(err.Error(), "failed to set mock runner")
}

func (ts *TestUtilHelpersTestSuite) TestCreateGoMod() {
	env := testutil.NewTestEnvironment(ts.mockT)
	defer env.Cleanup()

	moduleName := "github.com/example/test"
	env.CreateGoMod(moduleName)

	// Verify go.mod file was created
	ts.Require().True(env.FileExists("go.mod"))

	// Verify content
	content := env.ReadFile("go.mod")
	ts.Require().Contains(content, "module "+moduleName)
	ts.Require().Contains(content, "go 1.24")
	ts.Require().Contains(content, "github.com/magefile/mage")
	ts.Require().Contains(content, "github.com/stretchr/testify")
}

func (ts *TestUtilHelpersTestSuite) TestCreateMageConfig() {
	env := testutil.NewTestEnvironment(ts.mockT)
	defer env.Cleanup()

	config := `build:
  verbose: true
test:
  timeout: 10m`

	env.CreateMageConfig(config)

	// Verify config file was created
	ts.Require().True(env.FileExists(".mage.yaml"))

	// Verify content
	content := env.ReadFile(".mage.yaml")
	ts.Require().Equal(config, content)
}

func (ts *TestUtilHelpersTestSuite) TestCreateProjectStructure() {
	env := testutil.NewTestEnvironment(ts.mockT)
	defer env.Cleanup()

	env.CreateProjectStructure()

	// Verify directories were created
	expectedDirs := []string{
		"cmd/app",
		"pkg/utils",
		"bin",
		"docs",
	}

	for _, dir := range expectedDirs {
		info, err := os.Stat(dir)
		ts.Require().NoError(err)
		ts.Require().True(info.IsDir())
	}

	// Verify main.go was created
	ts.Require().True(env.FileExists("cmd/app/main.go"))
	content := env.ReadFile("cmd/app/main.go")
	ts.Require().Contains(content, "package main")
	ts.Require().Contains(content, `fmt.Println("Hello, World!")`)
}

func (ts *TestUtilHelpersTestSuite) TestCreateFile() {
	env := testutil.NewTestEnvironment(ts.mockT)
	defer env.Cleanup()

	// Test creating file in current directory
	env.CreateFile("test.txt", "test content")
	ts.Require().True(env.FileExists("test.txt"))
	ts.Require().Equal("test content", env.ReadFile("test.txt"))

	// Test creating file in subdirectory
	env.CreateFile("subdir/nested/file.txt", "nested content")
	ts.Require().True(env.FileExists("subdir/nested/file.txt"))
	ts.Require().Equal("nested content", env.ReadFile("subdir/nested/file.txt"))
}

func (ts *TestUtilHelpersTestSuite) TestFileExists() {
	env := testutil.NewTestEnvironment(ts.mockT)
	defer env.Cleanup()

	// File doesn't exist initially
	ts.Require().False(env.FileExists("nonexistent.txt"))

	// Create file and verify it exists
	env.CreateFile("exists.txt", "content")
	ts.Require().True(env.FileExists("exists.txt"))
}

func (ts *TestUtilHelpersTestSuite) TestReadFile() {
	env := testutil.NewTestEnvironment(ts.mockT)
	defer env.Cleanup()

	content := "test file content\nwith multiple lines"
	env.CreateFile("test.txt", content)

	readContent := env.ReadFile("test.txt")
	ts.Require().Equal(content, readContent)
}

func (ts *TestUtilHelpersTestSuite) TestStandardMocks() {
	env := testutil.NewTestEnvironment(ts.mockT)
	defer env.Cleanup()

	mocks := env.StandardMocks()
	ts.Require().NotNil(mocks)

	// Test build mocks
	buildMocks := mocks.ForBuild()
	ts.Require().NotNil(buildMocks)

	// Test test mocks
	testMocks := mocks.ForTest()
	ts.Require().NotNil(testMocks)

	// Test lint mocks
	lintMocks := mocks.ForLint()
	ts.Require().NotNil(lintMocks)

	// Test git mocks
	gitMocks := mocks.ForGit()
	ts.Require().NotNil(gitMocks)

	// Test all mocks
	allMocks := mocks.ForAll()
	ts.Require().NotNil(allMocks)
}

func (ts *TestUtilHelpersTestSuite) TestCleanupWithDirectoryError() {
	// Save original directory
	origDir, err := os.Getwd()
	ts.Require().NoError(err)

	env := testutil.NewTestEnvironment(ts.mockT)

	// Change to a directory that will be deleted
	tempSubDir := filepath.Join(env.TempDir, "subdir")
	err = os.MkdirAll(tempSubDir, 0o750)
	ts.Require().NoError(err)

	err = os.Chdir(tempSubDir)
	ts.Require().NoError(err)

	// Set an invalid original directory
	env.OrigDir = "/nonexistent/path"

	// Cleanup should not panic even with invalid directory
	env.Cleanup()

	// Restore to valid directory for test cleanup
	err = os.Chdir(origDir)
	ts.Require().NoError(err)
}

func (ts *TestUtilHelpersTestSuite) TestWithMockRunnerRestoreError() {
	env := testutil.NewTestEnvironment(ts.mockT)
	defer env.Cleanup()

	// Create a getter/setter where setter fails on restore
	var currentRunner interface{} = "original"
	callCount := 0

	getter := func() interface{} {
		return currentRunner
	}

	setter := func(runner interface{}) error {
		callCount++
		if callCount == 2 { // Fail on restore (second call)
			return errRestoreTest
		}
		currentRunner = runner
		return nil
	}

	executed := false
	err := env.WithMockRunner(setter, getter, func() error {
		executed = true
		return nil
	})

	// Should still succeed even if restore fails
	ts.Require().NoError(err)
	ts.Require().True(executed)
}

func TestTestUtilHelpersTestSuite(t *testing.T) {
	suite.Run(t, new(TestUtilHelpersTestSuite))
}
