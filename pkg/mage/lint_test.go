package mage

import (
	"testing"

	"github.com/mrz1836/go-mage/pkg/mage/testutil"
	"github.com/stretchr/testify/assert"
)

func TestLintAll(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateProjectStructure()
	env.CreateGoMod("github.com/test/project")

	tests := []struct {
		name      string
		setupMock func()
		expectErr bool
	}{
		{
			name: "successful lint all",
			setupMock: func() {
				env.Builder.ExpectAnyCommand(nil) // golangci-lint run
			},
			expectErr: false,
		},
		{
			name: "lint issues found",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectAnyCommand(assert.AnError) // linting issues
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			lint := Lint{}
			err := env.WithMockRunner(
				func(r interface{}) { SetRunner(r.(CommandRunner)) },
				func() interface{} { return GetRunner() },
				func() error {
					return lint.All()
				},
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestLintGo(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateProjectStructure()
	env.CreateGoMod("github.com/test/project")

	// Create Go files
	env.CreateFile("main.go", `package main

import "fmt"

func main() {
	fmt.Println("Hello World")
}`)

	tests := []struct {
		name      string
		setupMock func()
		expectErr bool
	}{
		{
			name: "successful go lint",
			setupMock: func() {
				env.Builder.ExpectAnyCommand(nil) // golangci-lint run
			},
			expectErr: false,
		},
		{
			name: "go lint issues",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectAnyCommand(assert.AnError) // lint issues found
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			lint := Lint{}
			err := env.WithMockRunner(
				func(r interface{}) { SetRunner(r.(CommandRunner)) },
				func() interface{} { return GetRunner() },
				func() error {
					return lint.Go()
				},
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestLintDocker(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateProjectStructure()
	env.CreateGoMod("github.com/test/project")

	// Create Dockerfile
	env.CreateFile("Dockerfile", `FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o app .
CMD ["./app"]`)

	tests := []struct {
		name      string
		setupMock func()
		expectErr bool
	}{
		{
			name: "successful docker lint",
			setupMock: func() {
				env.Builder.ExpectAnyCommand(nil) // hadolint
			},
			expectErr: false,
		},
		{
			name: "docker lint issues",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectAnyCommand(assert.AnError) // lint issues found
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			lint := Lint{}
			err := env.WithMockRunner(
				func(r interface{}) { SetRunner(r.(CommandRunner)) },
				func() interface{} { return GetRunner() },
				func() error {
					return lint.Docker()
				},
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestLintYAML(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateProjectStructure()
	env.CreateGoMod("github.com/test/project")

	// Create YAML files
	env.CreateFile("config.yaml", `
name: testproject
version: 1.0.0
settings:
  debug: true
  port: 8080
`)

	tests := []struct {
		name      string
		setupMock func()
		expectErr bool
	}{
		{
			name: "successful yaml lint",
			setupMock: func() {
				env.Builder.ExpectAnyCommand(nil) // yamllint
			},
			expectErr: false,
		},
		{
			name: "yaml lint issues",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectAnyCommand(assert.AnError) // lint issues found
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			lint := Lint{}
			err := env.WithMockRunner(
				func(r interface{}) { SetRunner(r.(CommandRunner)) },
				func() interface{} { return GetRunner() },
				func() error {
					return lint.YAML()
				},
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestLintMarkdown(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateProjectStructure()
	env.CreateGoMod("github.com/test/project")

	// Create markdown files
	env.CreateFile("README.md", `# Test Project

This is a test project for demonstration purposes.

## Features

- Feature 1
- Feature 2

## Installation

` + "```bash" + `
go install
` + "```" + `
`)

	tests := []struct {
		name      string
		setupMock func()
		expectErr bool
	}{
		{
			name: "successful markdown lint",
			setupMock: func() {
				env.Builder.ExpectAnyCommand(nil) // markdownlint
			},
			expectErr: false,
		},
		{
			name: "markdown lint issues",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectAnyCommand(assert.AnError) // lint issues found
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			lint := Lint{}
			err := env.WithMockRunner(
				func(r interface{}) { SetRunner(r.(CommandRunner)) },
				func() interface{} { return GetRunner() },
				func() error {
					return lint.Markdown()
				},
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestLintShell(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateProjectStructure()
	env.CreateGoMod("github.com/test/project")

	// Create shell scripts
	env.CreateFile("scripts/build.sh", `#!/bin/bash
set -e

echo "Building application..."
go build -o app .
echo "Build complete!"`)

	tests := []struct {
		name      string
		setupMock func()
		expectErr bool
	}{
		{
			name: "successful shell lint",
			setupMock: func() {
				env.Builder.ExpectAnyCommand(nil) // shellcheck
			},
			expectErr: false,
		},
		{
			name: "shell lint issues",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectAnyCommand(assert.AnError) // lint issues found
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			lint := Lint{}
			err := env.WithMockRunner(
				func(r interface{}) { SetRunner(r.(CommandRunner)) },
				func() interface{} { return GetRunner() },
				func() error {
					return lint.Shell()
				},
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestLintJSON(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateProjectStructure()
	env.CreateGoMod("github.com/test/project")

	// Create JSON files
	env.CreateFile("package.json", `{
  "name": "test-project",
  "version": "1.0.0",
  "scripts": {
    "test": "echo \"test\""
  }
}`)

	tests := []struct {
		name      string
		setupMock func()
		expectErr bool
	}{
		{
			name: "successful json lint",
			setupMock: func() {
				env.Builder.ExpectAnyCommand(nil) // jsonlint
			},
			expectErr: false,
		},
		{
			name: "json lint issues",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectAnyCommand(assert.AnError) // lint issues found
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			lint := Lint{}
			err := env.WithMockRunner(
				func(r interface{}) { SetRunner(r.(CommandRunner)) },
				func() interface{} { return GetRunner() },
				func() error {
					return lint.JSON()
				},
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestLintSQL(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateProjectStructure()
	env.CreateGoMod("github.com/test/project")

	// Create SQL files
	env.CreateFile("migrations/001_create_users.sql", `CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);`)

	tests := []struct {
		name      string
		setupMock func()
		expectErr bool
	}{
		{
			name: "successful sql lint",
			setupMock: func() {
				env.Builder.ExpectAnyCommand(nil) // sqlfluff
			},
			expectErr: false,
		},
		{
			name: "sql lint issues",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectAnyCommand(assert.AnError) // lint issues found
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			lint := Lint{}
			err := env.WithMockRunner(
				func(r interface{}) { SetRunner(r.(CommandRunner)) },
				func() interface{} { return GetRunner() },
				func() error {
					return lint.SQL()
				},
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestLintConfig(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateProjectStructure()
	env.CreateGoMod("github.com/test/project")

	// Create config files
	env.CreateFile(".golangci.yml", `run:
  timeout: 5m
  tests: false

linters:
  enable:
    - gofmt
    - golint
    - govet`)

	tests := []struct {
		name      string
		setupMock func()
		expectErr bool
	}{
		{
			name: "successful config lint",
			setupMock: func() {
				env.Builder.ExpectAnyCommand(nil) // config validation
			},
			expectErr: false,
		},
		{
			name: "config lint issues",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectAnyCommand(assert.AnError) // config issues found
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			lint := Lint{}
			err := env.WithMockRunner(
				func(r interface{}) { SetRunner(r.(CommandRunner)) },
				func() interface{} { return GetRunner() },
				func() error {
					return lint.Config()
				},
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestLintFix(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateProjectStructure()
	env.CreateGoMod("github.com/test/project")

	tests := []struct {
		name      string
		setupMock func()
		expectErr bool
	}{
		{
			name: "successful lint fix",
			setupMock: func() {
				env.Builder.ExpectAnyCommand(nil) // golangci-lint run --fix
			},
			expectErr: false,
		},
		{
			name: "lint fix issues",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectAnyCommand(assert.AnError) // fix issues found
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			lint := Lint{}
			err := env.WithMockRunner(
				func(r interface{}) { SetRunner(r.(CommandRunner)) },
				func() interface{} { return GetRunner() },
				func() error {
					return lint.Fix()
				},
			)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}
