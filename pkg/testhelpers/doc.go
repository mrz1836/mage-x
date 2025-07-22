// Package testhelpers provides utilities for testing mage-based projects.
//
// The package includes several key components:
//
// # TestEnvironment
//
// TestEnvironment provides an isolated environment for testing with automatic cleanup:
//
//	func TestMyTask(t *testing.T) {
//	    env := testhelpers.NewTestEnvironment(t)
//
//	    env.WriteFile("config.yaml", "key: value")
//	    env.SetEnv("MY_VAR", "test")
//
//	    output := env.CaptureOutput(func() {
//	        MyTask()
//	    })
//
//	    env.AssertFileContains("output.txt", "expected content")
//	}
//
// # MockRunner
//
// MockRunner provides a mock implementation of CommandRunner for testing command execution:
//
//	func TestBuildTask(t *testing.T) {
//	    runner := testhelpers.NewMockRunner()
//	    runner.SetOutputForCommand("go", "1.21.0")
//
//	    // Inject mock runner
//	    mage.SetRunner(runner)
//
//	    err := Build()
//	    assert.NoError(t, err)
//
//	    runner.AssertCalledWith(t, "go", "build", "-o", "app")
//	}
//
// # TestFixtures
//
// TestFixtures helps create common test data and project structures:
//
//	func TestProjectGeneration(t *testing.T) {
//	    env := testhelpers.NewTestEnvironment(t)
//	    fixtures := testhelpers.NewTestFixtures(t, env)
//
//	    fixtures.CreateGoProject("myapp")
//	    fixtures.CreateDockerfile("myapp")
//	    fixtures.CreateGitHubActions()
//
//	    // Test your project generation logic
//	}
//
// # TempWorkspace
//
// TempWorkspace provides a temporary file system workspace:
//
//	func TestFileOperations(t *testing.T) {
//	    ws := testhelpers.NewTempWorkspace(t, "test")
//
//	    ws.WriteTextFile("data.txt", "hello world")
//	    ws.CopyFile("data.txt", "backup/data.txt")
//
//	    ws.AssertFileEquals("backup/data.txt", "hello world")
//	}
//
// # Utilities
//
// The package also provides various utility functions:
//
//	// Skip tests based on conditions
//	testhelpers.RequireDocker(t)
//	testhelpers.RequireNetwork(t)
//	testhelpers.SkipIfShort(t)
//
//	// Assertions
//	testhelpers.AssertContains(t, output, "success")
//	testhelpers.EventuallyTrue(t, checkCondition, 5*time.Second, "condition failed")
//
//	// Golden files
//	golden := testhelpers.NewGolden(t, "testdata")
//	golden.CheckString("output", result)
package testhelpers
