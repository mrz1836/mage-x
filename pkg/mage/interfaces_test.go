package mage

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mrz1836/mage-x/pkg/testhelpers"
)

// Types are already defined in interfaces.go

// Mock implementations for interface testing

type MockBuilder struct {
	*testhelpers.MockBase
}

func NewMockBuilder(t *testing.T) *MockBuilder {
	return &MockBuilder{
		MockBase: testhelpers.NewMockBase(t),
	}
}

func (m *MockBuilder) Build(_ context.Context, _ *BuildOptions) error {
	err := m.ShouldReturnError("Build")
	m.RecordCall("Build", nil, nil, err)
	return err
}

func (m *MockBuilder) CrossCompile(_ context.Context, platforms []Platform) error {
	err := m.ShouldReturnError("CrossCompile")
	m.RecordCall("CrossCompile", []interface{}{platforms}, nil, err)
	return err
}

func (m *MockBuilder) Package(_ context.Context, format PackageFormat) error {
	err := m.ShouldReturnError("Package")
	m.RecordCall("Package", []interface{}{format}, nil, err)
	return err
}

func (m *MockBuilder) Clean(_ context.Context) error {
	err := m.ShouldReturnError("Clean")
	m.RecordCall("Clean", nil, nil, err)
	return err
}

func (m *MockBuilder) Install(_ context.Context, path string) error {
	err := m.ShouldReturnError("Install")
	m.RecordCall("Install", []interface{}{path}, nil, err)
	return err
}

type MockTester struct {
	*testhelpers.MockBase
}

func NewMockTester(t *testing.T) *MockTester {
	return &MockTester{
		MockBase: testhelpers.NewMockBase(t),
	}
}

func (m *MockTester) RunTests(_ context.Context, opts TestOptions) (*TestResults, error) {
	err := m.ShouldReturnError("RunTests")
	result := &TestResults{Passed: 10, Failed: 0, Duration: time.Second}
	if err != nil {
		result = nil
	}
	m.RecordCall("RunTests", []interface{}{opts}, []interface{}{result}, err)
	return result, err
}

func (m *MockTester) RunBenchmarks(_ context.Context, opts *IBenchmarkOptions) (*IBenchmarkResults, error) {
	err := m.ShouldReturnError("RunBenchmarks")
	result := &IBenchmarkResults{}
	if err != nil {
		result = nil
	}
	m.RecordCall("RunBenchmarks", []interface{}{opts}, []interface{}{result}, err)
	return result, err
}

func (m *MockTester) GenerateCoverage(_ context.Context, opts CoverageOptions) (*CoverageReport, error) {
	err := m.ShouldReturnError("GenerateCoverage")
	result := &CoverageReport{
		Overall: &CoverageInfo{Percentage: 85.5},
	}
	if err != nil {
		result = nil
	}
	m.RecordCall("GenerateCoverage", []interface{}{opts}, []interface{}{result}, err)
	return result, err
}

func (m *MockTester) RunUnit(_ context.Context, opts TestOptions) (*TestResults, error) {
	err := m.ShouldReturnError("RunUnit")
	result := &TestResults{Passed: 5, Failed: 0, Duration: time.Millisecond * 500}
	if err != nil {
		result = nil
	}
	m.RecordCall("RunUnit", []interface{}{opts}, []interface{}{result}, err)
	return result, err
}

func (m *MockTester) RunIntegration(_ context.Context, opts TestOptions) (*TestResults, error) {
	err := m.ShouldReturnError("RunIntegration")
	result := &TestResults{Passed: 5, Failed: 0, Duration: time.Millisecond * 1500}
	if err != nil {
		result = nil
	}
	m.RecordCall("RunIntegration", []interface{}{opts}, []interface{}{result}, err)
	return result, err
}

type MockDeployer struct {
	*testhelpers.MockBase
}

func NewMockDeployer(t *testing.T) *MockDeployer {
	return &MockDeployer{
		MockBase: testhelpers.NewMockBase(t),
	}
}

func (m *MockDeployer) Deploy(_ context.Context, target DeployTarget, opts *DeployOptions) error {
	err := m.ShouldReturnError("Deploy")
	m.RecordCall("Deploy", []interface{}{target, opts}, nil, err)
	return err
}

func (m *MockDeployer) Validate(_ context.Context, target DeployTarget) error {
	err := m.ShouldReturnError("Validate")
	m.RecordCall("Validate", []interface{}{target}, nil, err)
	return err
}

func (m *MockDeployer) Rollback(_ context.Context, target DeployTarget, version string) error {
	err := m.ShouldReturnError("Rollback")
	m.RecordCall("Rollback", []interface{}{target, version}, nil, err)
	return err
}

func (m *MockDeployer) Status(_ context.Context, target DeployTarget) (*DeployStatus, error) {
	err := m.ShouldReturnError("Status")
	result := &DeployStatus{Version: "1.0.0", State: "running"}
	if err != nil {
		result = nil
	}
	m.RecordCall("Status", []interface{}{target}, []interface{}{result}, err)
	return result, err
}

func (m *MockDeployer) Scale(_ context.Context, target DeployTarget, replicas int) error {
	err := m.ShouldReturnError("Scale")
	m.RecordCall("Scale", []interface{}{target, replicas}, nil, err)
	return err
}

func (m *MockDeployer) Logs(_ context.Context, target DeployTarget, opts LogOptions) ([]string, error) {
	err := m.ShouldReturnError("Logs")
	result := []string{"log line 1", "log line 2"}
	if err != nil {
		result = nil
	}
	m.RecordCall("Logs", []interface{}{target, opts}, []interface{}{result}, err)
	return result, err
}

// Test Builder interface
func TestBuilder_Interface(t *testing.T) {
	ctx := context.Background()
	builder := NewMockBuilder(t)

	t.Run("Build", func(t *testing.T) {
		opts := BuildOptions{Verbose: true}
		err := builder.Build(ctx, &opts)
		require.NoError(t, err)
		builder.AssertCalled("Build")
	})

	t.Run("CrossCompile", func(t *testing.T) {
		platforms := []Platform{{OS: "linux", Arch: "amd64"}}
		err := builder.CrossCompile(ctx, platforms)
		require.NoError(t, err)
		builder.AssertCalled("CrossCompile")
	})

	t.Run("Package", func(t *testing.T) {
		err := builder.Package(ctx, PackageFormatTarGz)
		require.NoError(t, err)
		builder.AssertCalled("Package")
	})

	t.Run("Clean", func(t *testing.T) {
		err := builder.Clean(ctx)
		require.NoError(t, err)
		builder.AssertCalled("Clean")
	})

	t.Run("Install", func(t *testing.T) {
		err := builder.Install(ctx, "/usr/local/bin")
		require.NoError(t, err)
		builder.AssertCalled("Install")
	})

	t.Run("Error handling", func(t *testing.T) {
		errorBuilder := NewMockBuilder(t)
		errorBuilder.SetMethodError("Build", assert.AnError)
		err := errorBuilder.Build(ctx, &BuildOptions{})
		assert.Error(t, err)
	})
}

// Test Tester interface
func TestTester_Interface(t *testing.T) {
	ctx := context.Background()
	tester := NewMockTester(t)

	t.Run("RunTests", func(t *testing.T) {
		opts := TestOptions{Verbose: true}
		results, err := tester.RunTests(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, results)
		tester.AssertCalled("RunTests")
		assert.Equal(t, 10, results.Passed)
		assert.Equal(t, 0, results.Failed)
	})

	t.Run("RunBenchmarks", func(t *testing.T) {
		opts := IBenchmarkOptions{}
		results, err := tester.RunBenchmarks(ctx, &opts)
		require.NoError(t, err)
		require.NotNil(t, results)
		tester.AssertCalled("RunBenchmarks")
	})

	t.Run("GenerateCoverage", func(t *testing.T) {
		opts := CoverageOptions{}
		report, err := tester.GenerateCoverage(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, report)
		tester.AssertCalled("GenerateCoverage")
		assert.InDelta(t, 85.5, report.Overall.Percentage, 0.001)
	})

	t.Run("RunUnit", func(t *testing.T) {
		opts := TestOptions{}
		results, err := tester.RunUnit(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, results)
		tester.AssertCalled("RunUnit")
		assert.Equal(t, 5, results.Passed)
	})

	t.Run("RunIntegration", func(t *testing.T) {
		opts := TestOptions{}
		results, err := tester.RunIntegration(ctx, opts)
		require.NoError(t, err)
		require.NotNil(t, results)
		tester.AssertCalled("RunIntegration")
		assert.Equal(t, 5, results.Passed)
	})

	t.Run("Error handling", func(t *testing.T) {
		errorTester := NewMockTester(t)
		errorTester.SetMethodError("RunTests", assert.AnError)
		_, err := errorTester.RunTests(ctx, TestOptions{})
		assert.Error(t, err)
	})
}

// Test Deployer interface
func TestDeployer_Interface(t *testing.T) {
	ctx := context.Background()
	deployer := NewMockDeployer(t)

	t.Run("Deploy", func(t *testing.T) {
		target := DeployTarget{Environment: "staging"}
		opts := DeployOptions{}
		err := deployer.Deploy(ctx, target, &opts)
		require.NoError(t, err)
		deployer.AssertCalled("Deploy")
	})

	t.Run("Validate", func(t *testing.T) {
		target := DeployTarget{Environment: "production"}
		err := deployer.Validate(ctx, target)
		require.NoError(t, err)
		deployer.AssertCalled("Validate")
	})

	t.Run("Rollback", func(t *testing.T) {
		target := DeployTarget{Environment: "production"}
		err := deployer.Rollback(ctx, target, "v1.0.0")
		require.NoError(t, err)
		deployer.AssertCalled("Rollback")
	})

	t.Run("Status", func(t *testing.T) {
		target := DeployTarget{Environment: "production"}
		status, err := deployer.Status(ctx, target)
		require.NoError(t, err)
		require.NotNil(t, status)
		deployer.AssertCalled("Status")
		assert.Equal(t, "1.0.0", status.Version)
		assert.Equal(t, "running", status.State)
	})

	t.Run("Scale", func(t *testing.T) {
		target := DeployTarget{Environment: "production"}
		err := deployer.Scale(ctx, target, 3)
		require.NoError(t, err)
		deployer.AssertCalled("Scale")
	})

	t.Run("Logs", func(t *testing.T) {
		target := DeployTarget{Environment: "production"}
		logs, err := deployer.Logs(ctx, target, LogOptions{})
		require.NoError(t, err)
		require.NotNil(t, logs)
		deployer.AssertCalled("Logs")
		assert.Len(t, logs, 2)
		assert.Equal(t, "log line 1", logs[0])
	})

	t.Run("Error handling", func(t *testing.T) {
		errorDeployer := NewMockDeployer(t)
		errorDeployer.SetMethodError("Deploy", assert.AnError)
		err := errorDeployer.Deploy(ctx, DeployTarget{}, &DeployOptions{})
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

		assert.InDelta(t, 92.5, report.Overall.Percentage, 0.001)
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
	builder := &MockBuilder{
		MockBase: testhelpers.NewMockBase(nil),
	}
	opts := BuildOptions{Verbose: false}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := builder.Build(ctx, &opts); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkTester_RunTests(b *testing.B) {
	ctx := context.Background()
	tester := &MockTester{
		MockBase: testhelpers.NewMockBase(nil),
	}
	opts := TestOptions{Verbose: false}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := tester.RunTests(ctx, opts); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDeployer_Deploy(b *testing.B) {
	ctx := context.Background()
	deployer := &MockDeployer{
		MockBase: testhelpers.NewMockBase(nil),
	}
	target := DeployTarget{Environment: "test"}
	opts := DeployOptions{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := deployer.Deploy(ctx, target, &opts); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPlatform_Access(b *testing.B) {
	platform := Platform{OS: "linux", Arch: "amd64"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = platform.OS + "/" + platform.Arch
	}
}
