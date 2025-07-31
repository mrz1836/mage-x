package testhelpers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

// TestFixtures provides helpers for creating test data
type TestFixtures struct {
	t   *testing.T
	env *TestEnvironment
}

// NewTestFixtures creates a new test fixtures helper
func NewTestFixtures(t *testing.T, env *TestEnvironment) *TestFixtures {
	return &TestFixtures{t: t, env: env}
}

// CreateGoProject creates a complete Go project structure
func (tf *TestFixtures) CreateGoProject(name string) {
	tf.t.Helper()

	// Create directories
	dirs := []string{
		"cmd/" + name,
		"pkg/" + name,
		"internal/" + name,
		"test",
		"docs",
		"scripts",
		".github/workflows",
	}

	for _, dir := range dirs {
		tf.env.MkdirAll(dir)
	}

	// Create go.mod
	tf.env.WriteFile("go.mod", fmt.Sprintf(`module github.com/example/%s

go 1.24

require (
	github.com/magefile/mage v1.15.0
	github.com/stretchr/testify v1.8.4
)
`, name))

	// Create main.go
	tf.env.WriteFile("cmd/"+name+"/main.go", fmt.Sprintf(`package main

import (
	"fmt"
	"os"
	
	"%s/pkg/%s"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	app := %s.New()
	return app.Run()
}
`, "github.com/example/"+name, name, name))

	// Create package file
	tf.env.WriteFile("pkg/"+name+"/"+name+".go", fmt.Sprintf(`package %s

import "fmt"

// App represents the application
type App struct {
	version string
}

// New creates a new app instance
func New() *App {
	return &App{
		version: "1.0.0",
	}
}

// Run runs the application
func (a *App) Run() error {
	fmt.Printf("%%s v%%s\n", "%s", a.version)
	return nil
}
`, name, name))

	// Create test file
	tf.env.WriteFile("pkg/"+name+"/"+name+"_test.go", fmt.Sprintf(`package %s

import (
	"testing"
)

func TestNew(t *testing.T) {
	app := New()
	if app == nil {
		t.Fatal("expected app to be non-nil")
	}
}

func TestRun(t *testing.T) {
	app := New()
	if err := app.Run(); err != nil {
		t.Fatalf("unexpected error: %%v", err)
	}
}
`, name))

	// Create README
	tf.env.WriteFile("README.md", fmt.Sprintf(`# %s

A sample Go project for testing.

## Installation

	go install ./cmd/%s

## Usage

	%s

## Development

This project uses Mage for build automation:

	mage build
	mage test
	mage lint
`, name, name, name))

	// Create .gitignore
	tf.env.WriteFile(".gitignore", `# Binaries
*.exe
*.exe~
*.dll
*.so
*.dylib
bin/
dist/

# Test binary
*.test

# Output
*.out
coverage.html
coverage.txt

# Dependency directories
vendor/

# Go workspace
go.work

# IDE
.idea/
.vscode/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Project specific
.mage/
`)

	// Create Makefile
	tf.env.WriteFile("Makefile", fmt.Sprintf(`.PHONY: all build test clean

all: build

build:
	go build -o bin/%s ./cmd/%s

test:
	go test -v ./...

clean:
	rm -rf bin/ dist/
`, name, name))
}

// CreateMageProject creates a project with mage configuration
func (tf *TestFixtures) CreateMageProject() {
	tf.t.Helper()

	// Create basic project
	tf.CreateGoProject("myapp")

	// Create magefile
	tf.env.WriteFile("magefile.go", `// +build mage

package main

import (
	"fmt"
	"os"
	
	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Build builds the application
func Build() error {
	fmt.Println("Building...")
	return sh.Run("go", "build", "-o", "bin/myapp", "./cmd/myapp")
}

// Test runs tests
func Test() error {
	mg.Deps(Build)
	fmt.Println("Running tests...")
	return sh.Run("go", "test", "./...")
}

// Lint runs linting
func Lint() error {
	fmt.Println("Linting...")
	if err := sh.Run("go", "vet", "./..."); err != nil {
		return err
	}
	return sh.Run("go", "fmt", "./...")
}

// Clean cleans build artifacts
func Clean() error {
	fmt.Println("Cleaning...")
	return os.RemoveAll("bin/")
}

// All runs all targets
func All() {
	mg.Deps(Lint, Test, Build)
}
`)

	// Create mage configuration
	tf.env.WriteFile(".mage.yaml", `project:
  name: myapp
  version: 1.0.0
  description: A sample application
  
build:
  output: bin/
  binary: myapp
  flags:
    - -v
    - -trimpath
  ldflags:
    - -s
    - -w
    
test:
  coverage: true
  coverage_file: coverage.out
  coverage_html: coverage.html
  timeout: 5m
  
lint:
  tools:
    - name: golangci-lint
      command: golangci-lint
      args: [run]
    - name: go-vet
      command: go
      args: [vet, ./...]
      
docker:
  registry: docker.io/example
  image: myapp
  dockerfile: Dockerfile
  
release:
  changelog: true
  artifacts:
    - binary: myapp
      platforms:
        - linux/amd64
        - linux/arm64
        - darwin/amd64
        - darwin/arm64
        - windows/amd64
`)
}

// CreateDockerfile creates a Dockerfile
func (tf *TestFixtures) CreateDockerfile(app string) {
	tf.t.Helper()

	tf.env.WriteFile("Dockerfile", fmt.Sprintf(`# Build stage
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o %s -ldflags="-w -s" ./cmd/%s

# Runtime stage
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/%s .

EXPOSE 8080

CMD ["./%s"]
`, app, app, app, app))
}

// CreateKubernetesManifests creates k8s manifests
func (tf *TestFixtures) CreateKubernetesManifests(app string) {
	tf.t.Helper()

	// Deployment
	tf.env.WriteFile("k8s/deployment.yaml", fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s
  labels:
    app: %s
spec:
  replicas: 3
  selector:
    matchLabels:
      app: %s
  template:
    metadata:
      labels:
        app: %s
    spec:
      containers:
      - name: %s
        image: %s:latest
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        resources:
          requests:
            memory: "64Mi"
            cpu: "250m"
          limits:
            memory: "128Mi"
            cpu: "500m"
`, app, app, app, app, app, app))

	// Service
	tf.env.WriteFile("k8s/service.yaml", fmt.Sprintf(`apiVersion: v1
kind: Service
metadata:
  name: %s
spec:
  selector:
    app: %s
  ports:
  - port: 80
    targetPort: 8080
  type: LoadBalancer
`, app, app))
}

// CreateGitHubActions creates GitHub Actions workflows
func (tf *TestFixtures) CreateGitHubActions() {
	tf.t.Helper()

	// CI workflow
	tf.env.WriteFile(".github/workflows/ci.yml", `name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.24'
    
    - name: Install Mage
      run: go install github.com/magefile/mage@latest
    
    - name: Run tests
      run: mage test
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.24'
    
    - name: Install Mage
      run: go install github.com/magefile/mage@latest
    
    - name: Run linting
      run: mage lint

  build:
    runs-on: ubuntu-latest
    needs: [test, lint]
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.24'
    
    - name: Install Mage
      run: go install github.com/magefile/mage@latest
    
    - name: Build
      run: mage build
    
    - name: Upload artifacts
      uses: actions/upload-artifact@v3
      with:
        name: binaries
        path: bin/
`)
}

// CreateTestData creates various test data files
func (tf *TestFixtures) CreateTestData() {
	tf.t.Helper()

	// JSON data
	tf.CreateJSONFile("testdata/config.json", map[string]interface{}{
		"name":    "test",
		"version": "1.0.0",
		"debug":   true,
		"servers": []map[string]interface{}{
			{"host": "localhost", "port": 8080},
			{"host": "0.0.0.0", "port": 9090},
		},
	})

	// YAML data
	tf.CreateYAMLFile("testdata/config.yaml", map[string]interface{}{
		"database": map[string]interface{}{
			"host":     "localhost",
			"port":     5432,
			"name":     "testdb",
			"user":     "testuser",
			"password": "testpass",
		},
		"cache": map[string]interface{}{
			"enabled": true,
			"ttl":     3600,
		},
	})

	// CSV data
	tf.env.WriteFile("testdata/data.csv", `id,name,email,created_at
1,John Doe,john@example.com,2023-01-01
2,Jane Smith,jane@example.com,2023-01-02
3,Bob Johnson,bob@example.com,2023-01-03
`)

	// XML data
	tf.env.WriteFile("testdata/data.xml", `<?xml version="1.0" encoding="UTF-8"?>
<users>
  <user id="1">
    <name>John Doe</name>
    <email>john@example.com</email>
  </user>
  <user id="2">
    <name>Jane Smith</name>
    <email>jane@example.com</email>
  </user>
</users>
`)
}

// CreateJSONFile creates a JSON file from data
func (tf *TestFixtures) CreateJSONFile(path string, data interface{}) {
	tf.t.Helper()

	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		tf.t.Fatalf("Failed to marshal JSON: %v", err)
	}

	tf.env.WriteFile(path, string(content))
}

// CreateYAMLFile creates a YAML file from data
func (tf *TestFixtures) CreateYAMLFile(path string, data interface{}) {
	tf.t.Helper()

	content, err := yaml.Marshal(data)
	if err != nil {
		tf.t.Fatalf("Failed to marshal YAML: %v", err)
	}

	tf.env.WriteFile(path, string(content))
}

// CreateBinaryFile creates a binary file with random data
func (tf *TestFixtures) CreateBinaryFile(path string, size int) {
	tf.t.Helper()

	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}

	fullPath := tf.env.AbsPath(path)
	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0o750); err != nil {
		tf.t.Fatalf("Failed to create directory: %v", err)
	}

	if err := os.WriteFile(fullPath, data, 0o600); err != nil {
		tf.t.Fatalf("Failed to write binary file: %v", err)
	}
}

// CreateLargeFile creates a large text file for testing
func (tf *TestFixtures) CreateLargeFile(path string, lines int) {
	tf.t.Helper()

	var content strings.Builder
	for i := 0; i < lines; i++ {
		content.WriteString(fmt.Sprintf("Line %d: This is a test line with some content to make it realistic.\n", i+1))
	}

	tf.env.WriteFile(path, content.String())
}

// CreateFileTree creates a complex file tree structure
func (tf *TestFixtures) CreateFileTree(depth, width int) {
	tf.t.Helper()
	tf.createFileTreeRecursive("tree", depth, width, 0)
}

func (tf *TestFixtures) createFileTreeRecursive(base string, maxDepth, width, currentDepth int) {
	if currentDepth >= maxDepth {
		return
	}

	for i := 0; i < width; i++ {
		// Create files
		fileName := fmt.Sprintf("%s/file_%d_%d.txt", base, currentDepth, i)
		tf.env.WriteFile(fileName, fmt.Sprintf("Content of %s", fileName))

		// Create subdirectories
		dirName := fmt.Sprintf("%s/dir_%d_%d", base, currentDepth, i)
		tf.env.MkdirAll(dirName)

		// Recurse
		tf.createFileTreeRecursive(dirName, maxDepth, width, currentDepth+1)
	}
}

// CreateGitRepository creates a git repository with history
func (tf *TestFixtures) CreateGitRepository() {
	tf.t.Helper()

	tf.env.SetupGitRepo()

	// Initial commit
	tf.env.WriteFile("README.md", "# Test Repository\n")
	tf.env.GitAdd("README.md")
	tf.env.GitCommit("Initial commit")

	// Add some history
	commits := []struct {
		file    string
		content string
		message string
	}{
		{"main.go", "package main\n\nfunc main() {}\n", "Add main.go"},
		{"main_test.go", "package main\n\nimport \"testing\"\n\nfunc TestMain(t *testing.T) {}\n", "Add tests"},
		{".gitignore", "*.exe\n*.test\n", "Add .gitignore"},
		{"docs/README.md", "# Documentation\n", "Add documentation"},
	}

	for _, commit := range commits {
		tf.env.WriteFile(commit.file, commit.content)
		tf.env.GitAdd(commit.file)
		tf.env.GitCommit(commit.message)

		// Add some delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)
	}
}

// CreatePythonProject creates a Python project structure
func (tf *TestFixtures) CreatePythonProject(name string) {
	tf.t.Helper()

	// Create directories
	dirs := []string{
		name,
		name + "/tests",
		"docs",
		"scripts",
	}

	for _, dir := range dirs {
		tf.env.MkdirAll(dir)
	}

	// Create setup.py
	tf.env.WriteFile("setup.py", fmt.Sprintf(`from setuptools import setup, find_packages

setup(
    name="%s",
    version="1.0.0",
    packages=find_packages(),
    install_requires=[
        "requests>=2.28.0",
        "click>=8.0.0",
    ],
    entry_points={
        'console_scripts': [
            '%s=%s.cli:main',
        ],
    },
)
`, name, name, name))

	// Create __init__.py
	tf.env.WriteFile(name+"/__init__.py", `"""A sample Python package."""

__version__ = "1.0.0"
`)

	// Create main module
	tf.env.WriteFile(name+"/main.py", fmt.Sprintf(`"""Main module for %s."""

def hello(name="World"):
    """Say hello to someone."""
    return f"Hello, {name}!"

if __name__ == "__main__":
    print(hello())
`, name))

	// Create CLI
	tf.env.WriteFile(name+"/cli.py", `"""Command-line interface."""

import click
from . import main

@click.command()
@click.option('--name', default='World', help='Name to greet.')
def greet(name):
    """Simple greeting program."""
    click.echo(main.hello(name))

def main():
    greet()

if __name__ == '__main__':
    main()
`)

	// Create tests
	tf.env.WriteFile(name+"/tests/__init__.py", "")
	tf.env.WriteFile(name+"/tests/test_main.py", `"""Tests for main module."""

import unittest
from .. import main

class TestMain(unittest.TestCase):
    def test_hello(self):
        self.assertEqual(main.hello(), "Hello, World!")
    
    def test_hello_with_name(self):
        self.assertEqual(main.hello("Alice"), "Hello, Alice!")

if __name__ == '__main__':
    unittest.main()
`)

	// Create requirements.txt
	tf.env.WriteFile("requirements.txt", `requests>=2.28.0
click>=8.0.0
pytest>=7.0.0
pytest-cov>=4.0.0
black>=22.0.0
flake8>=4.0.0
mypy>=0.990
`)

	// Create .gitignore
	tf.env.WriteFile(".gitignore", `# Python
__pycache__/
*.py[cod]
*$py.class
*.so
.Python
build/
develop-eggs/
dist/
downloads/
eggs/
.eggs/
lib/
lib64/
parts/
sdist/
var/
wheels/
*.egg-info/
.installed.cfg
*.egg
MANIFEST

# Virtual Environment
venv/
ENV/
env/

# Testing
.coverage
.pytest_cache/
htmlcov/

# IDE
.idea/
.vscode/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db
`)
}

// CreateNodeProject creates a Node.js project structure
func (tf *TestFixtures) CreateNodeProject(name string) {
	tf.t.Helper()

	// Create directories
	dirs := []string{
		"src",
		"test",
		"public",
		"docs",
	}

	for _, dir := range dirs {
		tf.env.MkdirAll(dir)
	}

	// Create package.json
	tf.CreateJSONFile("package.json", map[string]interface{}{
		"name":        name,
		"version":     "1.0.0",
		"description": "A sample Node.js project",
		"main":        "src/index.js",
		"scripts": map[string]string{
			"start": "node src/index.js",
			"test":  "jest",
			"lint":  "eslint src/",
			"dev":   "nodemon src/index.js",
		},
		"dependencies": map[string]string{
			"express": "^4.18.0",
			"dotenv":  "^16.0.0",
		},
		"devDependencies": map[string]string{
			"jest":    "^29.0.0",
			"eslint":  "^8.0.0",
			"nodemon": "^2.0.0",
		},
	})

	// Create index.js
	tf.env.WriteFile("src/index.js", `const express = require('express');
const app = express();
const port = process.env.PORT || 3000;

app.use(express.json());

app.get('/', (req, res) => {
    res.json({ message: 'Hello, World!' });
});

app.get('/health', (req, res) => {
    res.json({ status: 'ok' });
});

app.listen(port, () => {
    console.log(`+"`"+`Server running on port ${port}`+"`"+`);
});

module.exports = app;
`)

	// Create test
	tf.env.WriteFile("test/app.test.js", `const request = require('supertest');
const app = require('../src/index');

describe('GET /', () => {
    it('responds with hello message', async () => {
        const response = await request(app).get('/');
        expect(response.status).toBe(200);
        expect(response.body.message).toBe('Hello, World!');
    });
});

describe('GET /health', () => {
    it('responds with ok status', async () => {
        const response = await request(app).get('/health');
        expect(response.status).toBe(200);
        expect(response.body.status).toBe('ok');
    });
});
`)

	// Create .gitignore
	tf.env.WriteFile(".gitignore", `node_modules/
.env
.env.local
npm-debug.log*
yarn-debug.log*
yarn-error.log*
.DS_Store
coverage/
.nyc_output/
dist/
build/
`)
}
