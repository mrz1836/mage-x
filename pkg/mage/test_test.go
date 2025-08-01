package mage

import (
	"testing"

	"github.com/mrz1836/mage-x/pkg/mage/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTestRun(t *testing.T) {
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
			name: "successful test run",
			setupMock: func() {
				env.Builder.ExpectGoCommand("test", nil)
			},
			expectErr: false,
		},
		{
			name: "test failures",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectGoCommand("test", assert.AnError)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			test := Test{}
			err := env.WithMockRunner(
				func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
				func() interface{} { return GetRunner() },
				test.Run,
			)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestTestRunWithCoverage(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateProjectStructure()
	env.CreateGoMod("github.com/test/project")

	tests := []struct {
		name        string
		coverageDir string
		setupMock   func()
		expectErr   bool
	}{
		{
			name:        "successful test with coverage",
			coverageDir: "coverage",
			setupMock: func() {
				env.Builder.ExpectGoCommand("test", nil)
			},
			expectErr: false,
		},
		{
			name:        "test with coverage to default location",
			coverageDir: "",
			setupMock: func() {
				env.Builder.ExpectGoCommand("test", nil)
			},
			expectErr: false,
		},
		{
			name:        "coverage test failure",
			coverageDir: "coverage",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectGoCommand("test", assert.AnError)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			test := Test{}
			err := env.WithMockRunner(
				func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
				func() interface{} { return GetRunner() },
				func() error {
					return test.Coverage(tt.coverageDir)
				},
			)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestTestRace(t *testing.T) {
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
			name: "successful race test",
			setupMock: func() {
				env.Builder.ExpectGoCommand("test", nil)
			},
			expectErr: false,
		},
		{
			name: "race condition detected",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectGoCommand("test", assert.AnError)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			test := Test{}
			err := env.WithMockRunner(
				func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
				func() interface{} { return GetRunner() },
				test.Race,
			)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestTestBench(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateProjectStructure()
	env.CreateGoMod("github.com/test/project")

	tests := []struct {
		name      string
		pattern   string
		setupMock func()
		expectErr bool
	}{
		{
			name:    "run all benchmarks",
			pattern: ".",
			setupMock: func() {
				env.Builder.ExpectGoCommand("test", nil)
			},
			expectErr: false,
		},
		{
			name:    "run specific benchmark",
			pattern: "BenchmarkMyFunction",
			setupMock: func() {
				env.Builder.ExpectGoCommand("test", nil)
			},
			expectErr: false,
		},
		{
			name:    "benchmark failure",
			pattern: ".",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectGoCommand("test", assert.AnError)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			test := Test{}
			err := env.WithMockRunner(
				func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
				func() interface{} { return GetRunner() },
				func() error {
					return test.Bench(tt.pattern)
				},
			)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestTestVet(t *testing.T) {
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
			name: "successful vet",
			setupMock: func() {
				env.Builder.ExpectGoCommand("vet", nil)
			},
			expectErr: false,
		},
		{
			name: "vet issues found",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectGoCommand("vet", assert.AnError)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			test := Test{}
			err := env.WithMockRunner(
				func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
				func() interface{} { return GetRunner() },
				test.Vet,
			)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestTestLint(t *testing.T) {
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
			name: "successful lint",
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
				env.Builder.ExpectAnyCommand(assert.AnError)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			test := Test{}
			err := env.WithMockRunner(
				func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
				func() interface{} { return GetRunner() },
				test.Lint,
			)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestTestClean(t *testing.T) {
	env := testutil.NewTestEnvironment(t)
	defer env.Cleanup()

	env.CreateProjectStructure()
	env.CreateGoMod("github.com/test/project")

	// Create some test artifacts to clean
	env.CreateFile("coverage.txt", "coverage data")
	env.CreateFile("coverage.html", "coverage html")

	tests := []struct {
		name      string
		setupMock func()
		expectErr bool
	}{
		{
			name: "successful clean",
			setupMock: func() {
				env.Builder.ExpectGoCommand("clean", nil)
			},
			expectErr: false,
		},
		{
			name: "clean failure",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectGoCommand("clean", assert.AnError)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			test := Test{}
			err := env.WithMockRunner(
				func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
				func() interface{} { return GetRunner() },
				test.Clean,
			)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestTestUnit(t *testing.T) {
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
			name: "successful unit tests",
			setupMock: func() {
				env.Builder.ExpectGoCommand("test", nil)
			},
			expectErr: false,
		},
		{
			name: "unit test failures",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectGoCommand("test", assert.AnError)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			test := Test{}
			err := env.WithMockRunner(
				func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
				func() interface{} { return GetRunner() },
				test.Unit,
			)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestTestIntegration(t *testing.T) {
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
			name: "successful integration tests",
			setupMock: func() {
				env.Builder.ExpectGoCommand("test", nil)
			},
			expectErr: false,
		},
		{
			name: "integration test failures",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectGoCommand("test", assert.AnError)
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			test := Test{}
			err := env.WithMockRunner(
				func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
				func() interface{} { return GetRunner() },
				test.Integration,
			)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}

func TestTestAll(t *testing.T) {
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
			name: "all tests pass",
			setupMock: func() {
				// Expect multiple test commands (unit, integration, vet, lint)
				env.Builder.ExpectGoCommand("test", nil). // unit tests
										ExpectGoCommand("test", nil). // integration tests
										ExpectGoCommand("vet", nil).  // vet
										ExpectAnyCommand(nil)         // lint
			},
			expectErr: false,
		},
		{
			name: "some tests fail",
			setupMock: func() {
				// Reset expectations to avoid conflicts with previous test
				env.Runner.ExpectedCalls = nil
				env.Builder.ExpectGoCommand("test", assert.AnError) // unit test failure
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			test := Test{}
			err := env.WithMockRunner(
				func(r interface{}) error { return SetRunner(r.(CommandRunner)) }, //nolint:errcheck // Test setup function returns error
				func() interface{} { return GetRunner() },
				test.All,
			)

			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			env.Runner.AssertExpectations(t)
		})
	}
}
