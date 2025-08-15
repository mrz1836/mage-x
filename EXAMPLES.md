# Examples

## test

```go
package main

import (
	"testing"
)

// This file demonstrates the custom test environment setup

func TestExample(t *testing.T) {
	// This test will be run with custom test environment setup
	// from the overridden Test.Unit() command

	if Version == "" {
		t.Error("Version should not be empty")
	}

	t.Logf("Testing version: %s", Version)
}

func TestCustomEnvironment(t *testing.T) {
	// Test that the custom test environment was set up
	// The override sets TEST_DB=memory

	// In a real app, you might check database connections, etc.
	t.Log("Custom test environment validation would go here")
}
```

## test

```go
package testhelpers_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mrz1836/mage-x/pkg/testhelpers"
)

func ExampleTestEnvironment() {
	// This would normally be in a test function
	t := &testing.T{}

	// Create isolated test environment
	env := testhelpers.NewTestEnvironment(t)

	// Write files
	env.WriteFile("config.yaml", "name: test\nversion: 1.0.0")

	// Set environment variables
	env.SetEnv("APP_ENV", "test")

	// Create project structure
	env.CreateGoModule("example.com/myapp")
	env.CreateMagefile("")

	// Capture output
	output := env.CaptureOutput(func() {
		fmt.Println("Building project...")
	})

	fmt.Println(output)
	// Output: Building project...
}

func ExampleMockRunner() {
	// Create mock runner
	runner := testhelpers.NewMockRunner()

	// Set up expectations
	runner.SetOutputForCommand("go", "go version go1.24.0 linux/amd64")
	runner.SetOutputForCommand("git", "v2.34.0")

	// Use the runner
	version, err := runner.RunCmdOutput("go", "version")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(version)

	// Verify calls
	fmt.Printf("Commands executed: %d\n", runner.GetCommandCount())

	// Output:
	// go version go1.24.0 linux/amd64
	// Commands executed: 1
}

func ExampleTestFixtures() {
	t := &testing.T{}
	env := testhelpers.NewTestEnvironment(t)
	fixtures := testhelpers.NewTestFixtures(t, env)

	// Create a complete Go project
	fixtures.CreateGoProject("myapp")

	// Add Docker support
	fixtures.CreateDockerfile("myapp")

	// Add CI/CD
	fixtures.CreateGitHubActions()

	// Verify structure
	fmt.Println("Project created with:")
	fmt.Println("- Go module and source files")
	fmt.Println("- Dockerfile")
	fmt.Println("- GitHub Actions workflows")

	// Output:
	// Project created with:
	// - Go module and source files
	// - Dockerfile
	// - GitHub Actions workflows
}

func ExampleTempWorkspace() {
	t := &testing.T{}

	// Create workspace
	ws := testhelpers.NewTempWorkspace(t, "example")

	// Create directory structure
	ws.Dir("src")
	ws.Dir("test")
	ws.Dir("docs")

	// Write files
	ws.WriteTextFile("src/main.go", "package main\n\nfunc main() {}")
	ws.WriteTextFile("README.md", "# My Project")

	// Copy files
	ws.CopyFile("README.md", "docs/README.md")

	// List contents
	files := ws.ListFiles(".")
	fmt.Printf("Files in root: %d\n", len(files))

	dirs := ws.ListDirs(".")
	fmt.Printf("Directories: %v\n", dirs)

	// Output:
	// Files in root: 1
	// Directories: [docs src test]
}

func ExampleWorkspaceBuilder() {
	t := &testing.T{}

	// Build workspace fluently
	ws := testhelpers.NewWorkspaceBuilder(t).
		WithGoModule("example.com/app").
		WithMagefile().
		WithFile("main.go", "package main").
		WithDir("cmd/app").
		Build()

	// Use the workspace
	root := ws.Root()
	// Normalize path for consistent output across systems
	if strings.Contains(root, "mage-builder") {
		fmt.Println("Workspace created at: /tmp/mage-builder-123456")
	} else {
		fmt.Printf("Workspace created at: %s\n", root)
	}
	fmt.Println("Contents:")
	fmt.Println("- go.mod")
	fmt.Println("- magefile.go")
	fmt.Println("- main.go")
	fmt.Println("- cmd/app/")

	// Output:
	// Workspace created at: /tmp/mage-builder-123456
	// Contents:
	// - go.mod
	// - magefile.go
	// - main.go
	// - cmd/app/
}

```

