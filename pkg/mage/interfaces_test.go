package mage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Types are already defined in interfaces.go

// Mock implementations for interface testing

type MockBuilder struct {
	buildCalled        bool
	crossCompileCalled bool
	packageCalled      bool
	cleanCalled        bool
	installCalled      bool
	shouldError        bool
}

func (m *MockBuilder) Build(ctx context.Context, opts BuildOptions) error {
	m.buildCalled = true
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

func (m *MockBuilder) CrossCompile(ctx context.Context, platforms []Platform) error {
	m.crossCompileCalled = true
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

func (m *MockBuilder) Package(ctx context.Context, format PackageFormat) error {
	m.packageCalled = true
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

func (m *MockBuilder) Clean(ctx context.Context) error {
	m.cleanCalled = true
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

func (m *MockBuilder) Install(ctx context.Context, target string) error {
	m.installCalled = true
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

type MockTester struct {
	runTestsCalled       bool
	runBenchmarksCalled  bool
	generateCovCalled    bool
	runUnitCalled        bool
	runIntegrationCalled bool
	shouldError          bool
}

func (m *MockTester) RunTests(ctx context.Context, opts TestOptions) (*TestResults, error) {
	m.runTestsCalled = true
	if m.shouldError {
		return nil, assert.AnError
	}
	return &TestResults{Passed: 10, Failed: 0, Duration: time.Second}, nil
}

func (m *MockTester) RunBenchmarks(ctx context.Context, opts IBenchmarkOptions) (*IBenchmarkResults, error) {
	m.runBenchmarksCalled = true
	if m.shouldError {
		return nil, assert.AnError
	}
	return &IBenchmarkResults{}, nil
}

func (m *MockTester) GenerateCoverage(ctx context.Context, opts CoverageOptions) (*CoverageReport, error) {
	m.generateCovCalled = true
	if m.shouldError {
		return nil, assert.AnError
	}
	return &CoverageReport{
		Overall: &CoverageInfo{Percentage: 85.5},
	}, nil
}

func (m *MockTester) RunUnit(ctx context.Context, opts TestOptions) (*TestResults, error) {
	m.runUnitCalled = true
	if m.shouldError {
		return nil, assert.AnError
	}
	return &TestResults{Passed: 5, Failed: 0, Duration: time.Millisecond * 500}, nil
}

func (m *MockTester) RunIntegration(ctx context.Context, opts TestOptions) (*TestResults, error) {
	m.runIntegrationCalled = true
	if m.shouldError {
		return nil, assert.AnError
	}
	return &TestResults{Passed: 5, Failed: 0, Duration: time.Millisecond * 1500}, nil
}

type MockDeployer struct {
	deployCalled       bool
	validateCalled     bool
	rollbackCalled     bool
	getStatusCalled    bool
	scaleServiceCalled bool
	logsCalled         bool
	shouldError        bool
}

func (m *MockDeployer) Deploy(ctx context.Context, target DeployTarget, opts DeployOptions) error {
	m.deployCalled = true
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

func (m *MockDeployer) Validate(ctx context.Context, target DeployTarget) error {
	m.validateCalled = true
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

func (m *MockDeployer) Rollback(ctx context.Context, target DeployTarget, version string) error {
	m.rollbackCalled = true
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

func (m *MockDeployer) Status(ctx context.Context, target DeployTarget) (*DeployStatus, error) {
	m.getStatusCalled = true
	if m.shouldError {
		return nil, assert.AnError
	}
	return &DeployStatus{Version: "1.0.0", State: "running"}, nil
}

func (m *MockDeployer) Scale(ctx context.Context, target DeployTarget, replicas int) error {
	m.scaleServiceCalled = true
	if m.shouldError {
		return assert.AnError
	}
	return nil
}

func (m *MockDeployer) Logs(ctx context.Context, target DeployTarget, opts LogOptions) ([]string, error) {
	m.logsCalled = true
	if m.shouldError {
		return nil, assert.AnError
	}
	return []string{"log line 1", "log line 2"}, nil
}

// Test Builder interface
func TestBuilder_Interface(t *testing.T) {
	ctx := context.Background()
	builder := &MockBuilder{}

	t.Run("Build", func(t *testing.T) {
		opts := BuildOptions{Verbose: true}
		err := builder.Build(ctx, opts)
		require.NoError(t, err)
		assert.True(t, builder.buildCalled)
	})

	t.Run("CrossCompile", func(t *testing.T) {
		platforms := []Platform{{OS: "linux", Arch: "amd64"}}
		err := builder.CrossCompile(ctx, platforms)
		require.NoError(t, err)
		assert.True(t, builder.crossCompileCalled)
	})

	t.Run("Package", func(t *testing.T) {
		err := builder.Package(ctx, PackageFormatTarGz)
		require.NoError(t, err)
		assert.True(t, builder.packageCalled)
	})

	t.Run("Clean", func(t *testing.T) {
		err := builder.Clean(ctx)
		require.NoError(t, err)
		assert.True(t, builder.cleanCalled)
	})

	t.Run("Install", func(t *testing.T) {
		err := builder.Install(ctx, "/usr/local/bin")
		require.NoError(t, err)
		assert.True(t, builder.installCalled)
	})

	t.Run("Error handling", func(t *testing.T) {
		errorBuilder := &MockBuilder{shouldError: true}
		err := errorBuilder.Build(ctx, BuildOptions{})
		assert.Error(t, err)
	})
}

// Test Tester interface
func TestTester_Interface(t *testing.T) {
	ctx := context.Background()
	tester := &MockTester{}

	t.Run("RunTests", func(t *testing.T) {
		opts := TestOptions{Verbose: true}
		results, err := tester.RunTests(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.True(t, tester.runTestsCalled)
		assert.Equal(t, 10, results.Passed)
		assert.Equal(t, 0, results.Failed)
	})

	t.Run("RunBenchmarks", func(t *testing.T) {
		opts := IBenchmarkOptions{}
		results, err := tester.RunBenchmarks(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.True(t, tester.runBenchmarksCalled)
	})

	t.Run("GenerateCoverage", func(t *testing.T) {
		opts := CoverageOptions{}
		report, err := tester.GenerateCoverage(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, report)
		assert.True(t, tester.generateCovCalled)
		assert.Equal(t, 85.5, report.Overall.Percentage)
	})

	t.Run("RunUnit", func(t *testing.T) {
		opts := TestOptions{}
		results, err := tester.RunUnit(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.True(t, tester.runUnitCalled)
		assert.Equal(t, 5, results.Passed)
	})

	t.Run("RunIntegration", func(t *testing.T) {
		opts := TestOptions{}
		results, err := tester.RunIntegration(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, results)
		assert.True(t, tester.runIntegrationCalled)
		assert.Equal(t, 5, results.Passed)
	})

	t.Run("Error handling", func(t *testing.T) {
		errorTester := &MockTester{shouldError: true}
		_, err := errorTester.RunTests(ctx, TestOptions{})
		assert.Error(t, err)
	})
}

// Test Deployer interface
func TestDeployer_Interface(t *testing.T) {
	ctx := context.Background()
	deployer := &MockDeployer{}

	t.Run("Deploy", func(t *testing.T) {
		target := DeployTarget{Environment: "staging"}
		opts := DeployOptions{}
		err := deployer.Deploy(ctx, target, opts)
		require.NoError(t, err)
		assert.True(t, deployer.deployCalled)
	})

	t.Run("Validate", func(t *testing.T) {
		target := DeployTarget{Environment: "production"}
		err := deployer.Validate(ctx, target)
		require.NoError(t, err)
		assert.True(t, deployer.validateCalled)
	})

	t.Run("Rollback", func(t *testing.T) {
		target := DeployTarget{Environment: "production"}
		err := deployer.Rollback(ctx, target, "v1.0.0")
		require.NoError(t, err)
		assert.True(t, deployer.rollbackCalled)
	})

	t.Run("Status", func(t *testing.T) {
		target := DeployTarget{Environment: "production"}
		status, err := deployer.Status(ctx, target)
		require.NoError(t, err)
		require.NotNil(t, status)
		assert.True(t, deployer.getStatusCalled)
		assert.Equal(t, "1.0.0", status.Version)
		assert.Equal(t, "running", status.State)
	})

	t.Run("Scale", func(t *testing.T) {
		target := DeployTarget{Environment: "production"}
		err := deployer.Scale(ctx, target, 3)
		require.NoError(t, err)
		assert.True(t, deployer.scaleServiceCalled)
	})

	t.Run("Logs", func(t *testing.T) {
		target := DeployTarget{Environment: "production"}
		logs, err := deployer.Logs(ctx, target, LogOptions{})
		require.NoError(t, err)
		require.NotNil(t, logs)
		assert.True(t, deployer.logsCalled)
		assert.Len(t, logs, 2)
		assert.Equal(t, "log line 1", logs[0])
	})

	t.Run("Error handling", func(t *testing.T) {
		errorDeployer := &MockDeployer{shouldError: true}
		err := errorDeployer.Deploy(ctx, DeployTarget{}, DeployOptions{})
		assert.Error(t, err)
	})
}

// Test data structures and options
func TestBuildOptions(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		opts := BuildOptions{}
		assert.False(t, opts.Verbose)
		assert.Empty(t, opts.Tags)
		assert.Empty(t, opts.Output)
	})

	t.Run("with custom values", func(t *testing.T) {
		opts := BuildOptions{
			Verbose: true,
			Tags:    []string{"integration"},
			Output:  "/tmp/myapp",
			LDFlags: []string{"-s", "-w"},
		}

		assert.True(t, opts.Verbose)
		assert.Equal(t, []string{"integration"}, opts.Tags)
		assert.Equal(t, "/tmp/myapp", opts.Output)
		assert.Equal(t, []string{"-s", "-w"}, opts.LDFlags)
	})
}

func TestTestOptions(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		opts := TestOptions{}
		assert.False(t, opts.Verbose)
		assert.Empty(t, opts.Tags)
		assert.Empty(t, opts.Packages)
	})

	t.Run("with custom values", func(t *testing.T) {
		opts := TestOptions{
			Verbose:   true,
			Tags:      []string{"unit"},
			Packages:  []string{"./..."},
			Timeout:   time.Minute * 5,
			Race:      true,
			ShortMode: false,
			Cover:     true,
			Parallel:  4,
		}

		assert.True(t, opts.Verbose)
		assert.Equal(t, []string{"unit"}, opts.Tags)
		assert.Equal(t, []string{"./..."}, opts.Packages)
		assert.Equal(t, time.Minute*5, opts.Timeout)
		assert.True(t, opts.Race)
		assert.False(t, opts.ShortMode)
		assert.True(t, opts.Cover)
		assert.Equal(t, 4, opts.Parallel)
	})
}

func TestPlatform(t *testing.T) {
	t.Run("platform creation", func(t *testing.T) {
		platform := Platform{OS: "linux", Arch: "amd64"}
		assert.Equal(t, "linux", platform.OS)
		assert.Equal(t, "amd64", platform.Arch)
	})

	t.Run("empty platform", func(t *testing.T) {
		platform := Platform{}
		assert.Empty(t, platform.OS)
		assert.Empty(t, platform.Arch)
	})

	t.Run("common platforms", func(t *testing.T) {
		platforms := []Platform{
			{OS: "linux", Arch: "amd64"},
			{OS: "linux", Arch: "arm64"},
			{OS: "darwin", Arch: "amd64"},
			{OS: "darwin", Arch: "arm64"},
			{OS: "windows", Arch: "amd64"},
		}

		expectedOS := []string{"linux", "linux", "darwin", "darwin", "windows"}
		expectedArch := []string{"amd64", "arm64", "amd64", "arm64", "amd64"}

		for i, platform := range platforms {
			assert.Equal(t, expectedOS[i], platform.OS)
			assert.Equal(t, expectedArch[i], platform.Arch)
		}
	})
}

func TestPackageFormat(t *testing.T) {
	t.Run("package formats", func(t *testing.T) {
		formats := []PackageFormat{
			PackageFormatTarGz,
			PackageFormatZip,
			PackageFormatDeb,
			PackageFormatRPM,
			PackageFormatDMG,
		}

		expected := []string{"tar.gz", "zip", "deb", "rpm", "dmg"}

		for i, format := range formats {
			assert.Equal(t, expected[i], string(format))
		}
	})
}

func TestTestResults(t *testing.T) {
	t.Run("successful test results", func(t *testing.T) {
		results := TestResults{
			Passed:   10,
			Failed:   0,
			Skipped:  2,
			Duration: time.Second * 5,
		}

		assert.Equal(t, 10, results.Passed)
		assert.Equal(t, 0, results.Failed)
		assert.Equal(t, 2, results.Skipped)
		assert.Equal(t, time.Second*5, results.Duration)
		assert.Equal(t, 12, results.Passed+results.Failed+results.Skipped)
	})

	t.Run("failed test results", func(t *testing.T) {
		results := TestResults{
			Passed:  8,
			Failed:  2,
			Skipped: 1,
		}

		assert.Equal(t, 11, results.Passed+results.Failed+results.Skipped)
	})
}

func TestCoverageReport(t *testing.T) {
	t.Run("coverage report", func(t *testing.T) {
		report := CoverageReport{
			Overall: &CoverageInfo{
				Percentage: 92.5,
				Covered:    925,
				Uncovered:  75,
				Lines:      1000,
			},
		}

		assert.Equal(t, 92.5, report.Overall.Percentage)
		assert.Equal(t, 925, report.Overall.Covered)
		assert.Equal(t, 1000, report.Overall.Lines)
	})
}

func TestDeployTarget(t *testing.T) {
	t.Run("deploy target", func(t *testing.T) {
		target := DeployTarget{
			Environment: "production",
			Region:      "us-west-2",
			Name:        "prod-cluster",
			Provider:    "kubernetes",
		}

		assert.Equal(t, "production", target.Environment)
		assert.Equal(t, "us-west-2", target.Region)
		assert.Equal(t, "prod-cluster", target.Name)
		assert.Equal(t, "kubernetes", target.Provider)
	})
}

func TestDeployStatus(t *testing.T) {
	t.Run("deploy status", func(t *testing.T) {
		status := DeployStatus{
			Version:     "v1.2.3",
			State:       "running",
			Ready:       3,
			Total:       3,
			LastUpdated: time.Now(),
		}

		assert.Equal(t, "v1.2.3", status.Version)
		assert.Equal(t, "running", status.State)
		assert.Equal(t, 3, status.Ready)
		assert.Equal(t, 3, status.Total)
		assert.False(t, status.LastUpdated.IsZero())
	})
}

// Benchmark tests for interface methods
func BenchmarkBuilder_Build(b *testing.B) {
	ctx := context.Background()
	builder := &MockBuilder{}
	opts := BuildOptions{Verbose: false}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = builder.Build(ctx, opts)
	}
}

func BenchmarkTester_RunTests(b *testing.B) {
	ctx := context.Background()
	tester := &MockTester{}
	opts := TestOptions{Verbose: false}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = tester.RunTests(ctx, opts)
	}
}

func BenchmarkDeployer_Deploy(b *testing.B) {
	ctx := context.Background()
	deployer := &MockDeployer{}
	target := DeployTarget{Environment: "test"}
	opts := DeployOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = deployer.Deploy(ctx, target, opts)
	}
}

func BenchmarkPlatform_Access(b *testing.B) {
	platform := Platform{OS: "linux", Arch: "amd64"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = platform.OS + "/" + platform.Arch
	}
}
