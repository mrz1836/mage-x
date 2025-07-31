package testhelpers

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewTestFixtures(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	require.NotNil(t, tf)
	require.Equal(t, t, tf.t)
	require.Equal(t, env, tf.env)
}

func TestTestFixtures_CreateGoProject(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	tf.CreateGoProject("testapp")

	// Check directories
	dirs := []string{
		"cmd/testapp",
		"pkg/testapp",
		"internal/testapp",
		"test",
		"docs",
		"scripts",
		".github/workflows",
	}

	for _, dir := range dirs {
		env.AssertDirExists(dir)
	}

	// Check files
	env.AssertFileExists("go.mod")
	env.AssertFileContains("go.mod", "module github.com/example/testapp")
	env.AssertFileContains("go.mod", "go 1.24")

	env.AssertFileExists("cmd/testapp/main.go")
	env.AssertFileContains("cmd/testapp/main.go", "package main")
	env.AssertFileContains("cmd/testapp/main.go", "func main()")

	env.AssertFileExists("pkg/testapp/testapp.go")
	env.AssertFileContains("pkg/testapp/testapp.go", "package testapp")
	env.AssertFileContains("pkg/testapp/testapp.go", "type App struct")

	env.AssertFileExists("pkg/testapp/testapp_test.go")
	env.AssertFileContains("pkg/testapp/testapp_test.go", "func TestNew(t *testing.T)")

	env.AssertFileExists("README.md")
	env.AssertFileContains("README.md", "# testapp")

	env.AssertFileExists(".gitignore")
	env.AssertFileContains(".gitignore", "*.exe")
	env.AssertFileContains(".gitignore", "vendor/")

	env.AssertFileExists("Makefile")
	env.AssertFileContains("Makefile", "build:")
	env.AssertFileContains("Makefile", "test:")
}

func TestTestFixtures_CreateMageProject(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	tf.CreateMageProject()

	// Check that Go project was created
	env.AssertFileExists("cmd/myapp/main.go")

	// Check magefile
	env.AssertFileExists("magefile.go")
	env.AssertFileContains("magefile.go", "// +build mage")
	env.AssertFileContains("magefile.go", "func Build() error")
	env.AssertFileContains("magefile.go", "func Test() error")
	env.AssertFileContains("magefile.go", "func Lint() error")
	env.AssertFileContains("magefile.go", "func Clean() error")
	env.AssertFileContains("magefile.go", "func All()")

	// Check mage configuration
	env.AssertFileExists(".mage.yaml")
	content := env.ReadFile(".mage.yaml")

	var config map[string]interface{}
	err := yaml.Unmarshal([]byte(content), &config)
	require.NoError(t, err)

	projectValue, ok := config["project"]
	require.True(t, ok, "project key should exist")
	project, ok := projectValue.(map[string]interface{})
	require.True(t, ok, "project should be a map")
	require.Equal(t, "myapp", project["name"])
	require.Equal(t, "1.0.0", project["version"])

	buildValue, ok := config["build"]
	require.True(t, ok, "build key should exist")
	build, ok := buildValue.(map[string]interface{})
	require.True(t, ok, "build should be a map")
	require.Equal(t, "bin/", build["output"])
	require.Equal(t, "myapp", build["binary"])
}

func TestTestFixtures_CreateDockerfile(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	tf.CreateDockerfile("testapp")

	env.AssertFileExists("Dockerfile")
	content := env.ReadFile("Dockerfile")

	// Check multi-stage build
	require.Contains(t, content, "FROM golang:1.24-alpine AS builder")
	require.Contains(t, content, "FROM alpine:latest")

	// Check build commands
	require.Contains(t, content, "go mod download")
	require.Contains(t, content, "go build -o testapp")

	// Check runtime
	require.Contains(t, content, "COPY --from=builder /app/testapp")
	require.Contains(t, content, "EXPOSE 8080")
	require.Contains(t, content, "CMD [\"./testapp\"]")
}

func TestTestFixtures_CreateKubernetesManifests(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	tf.CreateKubernetesManifests("testapp")

	// Check deployment
	env.AssertFileExists("k8s/deployment.yaml")
	deployment := env.ReadFile("k8s/deployment.yaml")
	require.Contains(t, deployment, "kind: Deployment")
	require.Contains(t, deployment, "name: testapp")
	require.Contains(t, deployment, "replicas: 3")
	require.Contains(t, deployment, "containerPort: 8080")

	// Check service
	env.AssertFileExists("k8s/service.yaml")
	service := env.ReadFile("k8s/service.yaml")
	require.Contains(t, service, "kind: Service")
	require.Contains(t, service, "name: testapp")
	require.Contains(t, service, "type: LoadBalancer")
}

func TestTestFixtures_CreateGitHubActions(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	tf.CreateGitHubActions()

	env.AssertFileExists(".github/workflows/ci.yml")
	workflow := env.ReadFile(".github/workflows/ci.yml")

	// Check workflow structure
	require.Contains(t, workflow, "name: CI")
	require.Contains(t, workflow, "on:")
	require.Contains(t, workflow, "push:")
	require.Contains(t, workflow, "pull_request:")

	// Check jobs
	require.Contains(t, workflow, "jobs:")
	require.Contains(t, workflow, "test:")
	require.Contains(t, workflow, "lint:")
	require.Contains(t, workflow, "build:")

	// Check steps
	require.Contains(t, workflow, "uses: actions/checkout@v3")
	require.Contains(t, workflow, "uses: actions/setup-go@v4")
	require.Contains(t, workflow, "go-version: '1.24'")
	require.Contains(t, workflow, "mage test")
	require.Contains(t, workflow, "mage lint")
	require.Contains(t, workflow, "mage build")
}

func TestTestFixtures_CreateTestData(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	tf.CreateTestData()

	// Check JSON file
	env.AssertFileExists("testdata/config.json")
	jsonContent := env.ReadFile("testdata/config.json")
	var jsonData map[string]interface{}
	err := json.Unmarshal([]byte(jsonContent), &jsonData)
	require.NoError(t, err)
	require.Equal(t, "test", jsonData["name"])
	require.Equal(t, "1.0.0", jsonData["version"])
	require.Equal(t, true, jsonData["debug"])

	// Check YAML file
	env.AssertFileExists("testdata/config.yaml")
	yamlContent := env.ReadFile("testdata/config.yaml")
	var yamlData map[string]interface{}
	err = yaml.Unmarshal([]byte(yamlContent), &yamlData)
	require.NoError(t, err)
	databaseValue, ok := yamlData["database"]
	require.True(t, ok, "database key should exist")
	database, ok := databaseValue.(map[string]interface{})
	require.True(t, ok, "database should be a map")
	require.Equal(t, "localhost", database["host"])
	require.Equal(t, 5432, database["port"])

	// Check CSV file
	env.AssertFileExists("testdata/data.csv")
	env.AssertFileContains("testdata/data.csv", "id,name,email,created_at")
	env.AssertFileContains("testdata/data.csv", "John Doe")

	// Check XML file
	env.AssertFileExists("testdata/data.xml")
	env.AssertFileContains("testdata/data.xml", "<?xml version=\"1.0\"")
	env.AssertFileContains("testdata/data.xml", "<users>")
	env.AssertFileContains("testdata/data.xml", "<name>John Doe</name>")
}

func TestTestFixtures_CreateJSONFile(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	data := map[string]interface{}{
		"string": "value",
		"number": 42,
		"bool":   true,
		"array":  []string{"a", "b", "c"},
		"nested": map[string]interface{}{
			"key": "value",
		},
	}

	tf.CreateJSONFile("test.json", data)

	env.AssertFileExists("test.json")
	content := env.ReadFile("test.json")

	var loaded map[string]interface{}
	err := json.Unmarshal([]byte(content), &loaded)
	require.NoError(t, err)

	require.Equal(t, "value", loaded["string"])
	require.InDelta(t, float64(42), loaded["number"], 0.001) // JSON numbers are float64
	require.Equal(t, true, loaded["bool"])
}

func TestTestFixtures_CreateYAMLFile(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	data := map[string]interface{}{
		"string": "value",
		"number": 42,
		"bool":   true,
		"list":   []string{"a", "b", "c"},
		"map": map[string]interface{}{
			"key": "value",
		},
	}

	tf.CreateYAMLFile("test.yaml", data)

	env.AssertFileExists("test.yaml")
	content := env.ReadFile("test.yaml")

	var loaded map[string]interface{}
	err := yaml.Unmarshal([]byte(content), &loaded)
	require.NoError(t, err)

	require.Equal(t, "value", loaded["string"])
	require.Equal(t, 42, loaded["number"])
	require.Equal(t, true, loaded["bool"])
}

func TestTestFixtures_CreateBinaryFile(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	tf.CreateBinaryFile("test.bin", 256)

	env.AssertFileExists("test.bin")
	data := env.ReadFile("test.bin")

	require.Len(t, data, 256)

	// Check pattern
	for i := 0; i < 256; i++ {
		require.Equal(t, byte(i%256), data[i])
	}
}

func TestTestFixtures_CreateLargeFile(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	tf.CreateLargeFile("large.txt", 100)

	env.AssertFileExists("large.txt")
	content := env.ReadFile("large.txt")

	lines := strings.Split(content, "\n")
	// 100 lines plus one empty line at the end
	require.Len(t, lines, 101)

	// Check first and last lines
	require.Contains(t, lines[0], "Line 1:")
	require.Contains(t, lines[99], "Line 100:")
}

func TestTestFixtures_CreateFileTree(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	tf.CreateFileTree(2, 2)

	// Check root level
	env.AssertFileExists("tree/file_0_0.txt")
	env.AssertFileExists("tree/file_0_1.txt")
	env.AssertDirExists("tree/dir_0_0")
	env.AssertDirExists("tree/dir_0_1")

	// Check second level
	env.AssertFileExists("tree/dir_0_0/file_1_0.txt")
	env.AssertFileExists("tree/dir_0_0/file_1_1.txt")
	env.AssertDirExists("tree/dir_0_0/dir_1_0")
	env.AssertDirExists("tree/dir_0_0/dir_1_1")

	// Verify content
	content := env.ReadFile("tree/dir_0_0/file_1_0.txt")
	require.Equal(t, "Content of tree/dir_0_0/file_1_0.txt", content)
}

func TestTestFixtures_CreateGitRepository(t *testing.T) {
	// Skip if git is not available
	if _, err := os.Stat("/usr/bin/git"); err != nil && os.IsNotExist(err) {
		t.Skip("git not available")
	}

	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	tf.CreateGitRepository()

	// Check git directory
	env.AssertDirExists(".git")

	// Check files
	env.AssertFileExists("README.md")
	env.AssertFileExists("main.go")
	env.AssertFileExists("main_test.go")
	env.AssertFileExists(".gitignore")
	env.AssertFileExists("docs/README.md")

	// Verify content
	env.AssertFileContains("README.md", "# Test Repository")
	env.AssertFileContains("main.go", "package main")
	env.AssertFileContains(".gitignore", "*.exe")
}

func TestTestFixtures_CreatePythonProject(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	tf.CreatePythonProject("pyapp")

	// Check directories
	env.AssertDirExists("pyapp")
	env.AssertDirExists("pyapp/tests")
	env.AssertDirExists("docs")
	env.AssertDirExists("scripts")

	// Check files
	env.AssertFileExists("setup.py")
	env.AssertFileContains("setup.py", "name=\"pyapp\"")
	env.AssertFileContains("setup.py", "find_packages()")

	env.AssertFileExists("pyapp/__init__.py")
	env.AssertFileContains("pyapp/__init__.py", "__version__ = \"1.0.0\"")

	env.AssertFileExists("pyapp/main.py")
	env.AssertFileContains("pyapp/main.py", "def hello(name=\"World\"):")

	env.AssertFileExists("pyapp/cli.py")
	env.AssertFileContains("pyapp/cli.py", "import click")
	env.AssertFileContains("pyapp/cli.py", "@click.command()")

	env.AssertFileExists("pyapp/tests/__init__.py")
	env.AssertFileExists("pyapp/tests/test_main.py")
	env.AssertFileContains("pyapp/tests/test_main.py", "class TestMain(unittest.TestCase):")

	env.AssertFileExists("requirements.txt")
	env.AssertFileContains("requirements.txt", "requests>=2.28.0")
	env.AssertFileContains("requirements.txt", "pytest>=7.0.0")

	env.AssertFileExists(".gitignore")
	env.AssertFileContains(".gitignore", "__pycache__/")
	env.AssertFileContains(".gitignore", "venv/")
}

func TestTestFixtures_CreateNodeProject(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	tf.CreateNodeProject("nodeapp")

	// Check directories
	env.AssertDirExists("src")
	env.AssertDirExists("test")
	env.AssertDirExists("public")
	env.AssertDirExists("docs")

	// Check package.json
	env.AssertFileExists("package.json")
	content := env.ReadFile("package.json")
	var pkg map[string]interface{}
	err := json.Unmarshal([]byte(content), &pkg)
	require.NoError(t, err)

	require.Equal(t, "nodeapp", pkg["name"])
	require.Equal(t, "1.0.0", pkg["version"])
	require.Equal(t, "src/index.js", pkg["main"])

	scriptsValue, ok := pkg["scripts"]
	require.True(t, ok, "scripts key should exist")
	scripts, ok := scriptsValue.(map[string]interface{})
	require.True(t, ok, "scripts should be a map")
	require.Equal(t, "node src/index.js", scripts["start"])
	require.Equal(t, "jest", scripts["test"])

	// Check index.js
	env.AssertFileExists("src/index.js")
	env.AssertFileContains("src/index.js", "const express = require('express')")
	env.AssertFileContains("src/index.js", "app.get('/'")
	env.AssertFileContains("src/index.js", "app.listen(port")

	// Check test
	env.AssertFileExists("test/app.test.js")
	env.AssertFileContains("test/app.test.js", "describe('GET /'")
	env.AssertFileContains("test/app.test.js", "expect(response.status).toBe(200)")

	// Check .gitignore
	env.AssertFileExists(".gitignore")
	env.AssertFileContains(".gitignore", "node_modules/")
	env.AssertFileContains(".gitignore", ".env")
}

func TestTestFixtures_ComplexScenario(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	// Create a complex project with multiple components
	tf.CreateMageProject()
	tf.CreateDockerfile("myapp")
	tf.CreateKubernetesManifests("myapp")
	tf.CreateGitHubActions()
	tf.CreateTestData()

	// Verify everything was created correctly
	env.AssertFileExists("magefile.go")
	env.AssertFileExists("Dockerfile")
	env.AssertFileExists("k8s/deployment.yaml")
	env.AssertFileExists("k8s/service.yaml")
	env.AssertFileExists(".github/workflows/ci.yml")
	env.AssertFileExists("testdata/config.json")
	env.AssertFileExists("testdata/config.yaml")

	// Create a file tree
	tf.CreateFileTree(3, 3)

	// Count files
	fileCount := 0
	err := filepath.Walk(env.AbsPath("tree"), func(_ string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !info.IsDir() {
			fileCount++
		}
		return nil
	})
	require.NoError(t, err)
	require.Greater(t, fileCount, 30) // Should have many files
}

func TestTestFixtures_EdgeCases(t *testing.T) {
	env := NewTestEnvironment(t)
	tf := NewTestFixtures(t, env)

	t.Run("empty project name", func(_ *testing.T) {
		// Should still work with empty name
		tf.CreateGoProject("")
		env.AssertFileExists("go.mod")
	})

	t.Run("special characters in name", func(_ *testing.T) {
		tf.CreatePythonProject("test-app_2023")
		env.AssertDirExists("test-app_2023")
	})

	t.Run("very deep file tree", func(_ *testing.T) {
		tf.CreateFileTree(5, 2)
		// Check deep file
		deepFile := "tree/dir_0_0/dir_1_0/dir_2_0/dir_3_0/file_4_0.txt"
		env.AssertFileExists(deepFile)
	})

	t.Run("large binary file", func(t *testing.T) {
		tf.CreateBinaryFile("large.bin", 1024*1024) // 1MB
		info, err := os.Stat(env.AbsPath("large.bin"))
		require.NoError(t, err)
		require.Equal(t, int64(1024*1024), info.Size())
	})
}
