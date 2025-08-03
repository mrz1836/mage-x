package main

import (
	"errors"
	"fmt"
	"testing"

	"github.com/mrz1836/mage-x/pkg/mage"
	"github.com/mrz1836/mage-x/pkg/testhelpers"
)

// Static errors for mock testing
var (
	errMockBuildFailure  = errors.New("mock build failure")
	errMockTestFailure   = errors.New("test failure")
	errMockBuildError    = errors.New("build failure")
	errMockPreBuildError = errors.New("pre-build failure")
)

// MockBuild is a mock implementation of BuildNamespace using testhelpers.MockBase
type MockBuild struct {
	*testhelpers.MockBase

	platformCalled map[string]bool // Keep for specific platform tracking
}

// NewMockBuild creates a new mock build namespace
func NewMockBuild(t *testing.T) *MockBuild {
	return &MockBuild{
		MockBase:       testhelpers.NewMockBase(t),
		platformCalled: make(map[string]bool),
	}
}

func (m *MockBuild) Default() error {
	err := m.ShouldReturnError("Default")
	m.RecordCall("Default", nil, nil, err)
	return err
}

func (m *MockBuild) All() error {
	err := m.ShouldReturnError("All")
	m.RecordCall("All", nil, nil, err)
	return err
}

func (m *MockBuild) Platform(platform string) error {
	m.platformCalled[platform] = true
	err := m.ShouldReturnError("Platform")
	m.RecordCall("Platform", []interface{}{platform}, nil, err)
	return err
}

func (m *MockBuild) Linux() error {
	err := m.Platform("linux/amd64")
	m.RecordCall("Linux", nil, nil, err)
	return err
}

func (m *MockBuild) Darwin() error {
	err := m.Platform("darwin/amd64")
	m.RecordCall("Darwin", nil, nil, err)
	return err
}

func (m *MockBuild) Windows() error {
	err := m.Platform("windows/amd64")
	m.RecordCall("Windows", nil, nil, err)
	return err
}

func (m *MockBuild) Docker() error {
	err := m.ShouldReturnError("Docker")
	m.RecordCall("Docker", nil, nil, err)
	return err
}

func (m *MockBuild) Clean() error {
	err := m.ShouldReturnError("Clean")
	m.RecordCall("Clean", nil, nil, err)
	return err
}

func (m *MockBuild) Install() error {
	err := m.ShouldReturnError("Install")
	m.RecordCall("Install", nil, nil, err)
	return err
}

func (m *MockBuild) Generate() error {
	err := m.ShouldReturnError("Generate")
	m.RecordCall("Generate", nil, nil, err)
	return err
}

func (m *MockBuild) PreBuild() error {
	err := m.ShouldReturnError("PreBuild")
	m.RecordCall("PreBuild", nil, nil, err)
	return err
}

// MockTest is a mock implementation of TestNamespace using testhelpers.MockBase
type MockTest struct {
	*testhelpers.MockBase
	mage.Test // Embed the real Test implementation
}

func NewMockTest(t *testing.T) *MockTest {
	return &MockTest{
		MockBase: testhelpers.NewMockBase(t),
	}
}

// Only override the methods we want to track/mock
func (m *MockTest) Default() error {
	err := m.ShouldReturnError("Default")
	m.RecordCall("Default", nil, nil, err)
	return err
}

func (m *MockTest) Unit() error {
	err := m.ShouldReturnError("Unit")
	m.RecordCall("Unit", nil, nil, err)
	return err
}

func (m *MockTest) Integration() error {
	err := m.ShouldReturnError("Integration")
	m.RecordCall("Integration", nil, nil, err)
	return err
}

func (m *MockTest) Coverage(args ...string) error {
	err := m.ShouldReturnError("Coverage")
	m.RecordCall("Coverage", []interface{}{args}, nil, err)
	return err
}

func (m *MockTest) Race() error {
	err := m.ShouldReturnError("Race")
	m.RecordCall("Race", nil, nil, err)
	return err
}

// Functions to test

func deployApp(build mage.BuildNamespace, platforms []string) error {
	for _, platform := range platforms {
		if err := build.Platform(platform); err != nil {
			return fmt.Errorf("failed to build for %s: %w", platform, err)
		}
	}
	return nil
}

func runCI(build mage.BuildNamespace, test mage.TestNamespace) error {
	// Run tests first
	if err := test.Coverage(); err != nil {
		return fmt.Errorf("tests failed: %w", err)
	}

	// Then build
	if err := build.Default(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	return nil
}

func buildPipeline(build mage.BuildNamespace) error {
	// Pre-build
	if err := build.PreBuild(); err != nil {
		return fmt.Errorf("pre-build failed: %w", err)
	}

	// Clean first
	if err := build.Clean(); err != nil {
		return fmt.Errorf("clean failed: %w", err)
	}

	// Then build
	if err := build.Default(); err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	return nil
}

// Test functions

func TestDeployApp(t *testing.T) {
	t.Run("successful deployment", func(t *testing.T) {
		mockBuild := NewMockBuild(t)
		platforms := []string{"linux/amd64", "darwin/amd64", "windows/amd64"}

		err := deployApp(mockBuild, platforms)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify all platforms were built
		for _, platform := range platforms {
			if !mockBuild.platformCalled[platform] {
				t.Errorf("Platform %s was not built", platform)
			}
		}

		// Verify call order using MockBase
		mockBuild.AssertCalled("Platform")
		mockBuild.AssertCalledTimes("Platform", 3)

		// Verify platform arguments
		for _, platform := range platforms {
			mockBuild.AssertCalledWith("Platform", platform)
		}
	})

	t.Run("build failure", func(t *testing.T) {
		mockBuild := NewMockBuild(t)
		mockBuild.SetMethodError("Platform", errMockBuildFailure)
		platforms := []string{"linux/amd64"}

		err := deployApp(mockBuild, platforms)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err.Error() != "failed to build for linux/amd64: mock build failure" {
			t.Errorf("Unexpected error message: %v", err)
		}
	})
}

func TestRunCI(t *testing.T) {
	t.Run("successful CI", func(t *testing.T) {
		mockBuild := NewMockBuild(t)
		mockTest := NewMockTest(t)

		err := runCI(mockBuild, mockTest)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify tests were run
		mockTest.AssertCalled("Coverage")

		// Verify build was run
		mockBuild.AssertCalled("Default")

		// Both operations should have been called once
		mockTest.AssertCalledTimes("Coverage", 1)
		mockBuild.AssertCalledTimes("Default", 1)
	})

	t.Run("test failure", func(t *testing.T) {
		mockBuild := NewMockBuild(t)
		mockTest := NewMockTest(t)
		mockTest.SetMethodError("Coverage", errMockTestFailure)

		err := runCI(mockBuild, mockTest)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err.Error() != "tests failed: test failure" {
			t.Errorf("Unexpected error message: %v", err)
		}

		// Build should not have been called due to test failure
		mockBuild.AssertNotCalled("Default")
	})

	t.Run("build failure", func(t *testing.T) {
		mockBuild := NewMockBuild(t)
		mockTest := NewMockTest(t)
		mockBuild.SetMethodError("Default", errMockBuildError)

		err := runCI(mockBuild, mockTest)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err.Error() != "build failed: build failure" {
			t.Errorf("Unexpected error message: %v", err)
		}

		// Tests should have been called
		mockTest.AssertCalled("Coverage")
	})
}

func TestBuildPipeline(t *testing.T) {
	t.Run("successful pipeline", func(t *testing.T) {
		mockBuild := NewMockBuild(t)

		err := buildPipeline(mockBuild)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify all operations were called
		mockBuild.AssertCalled("PreBuild")
		mockBuild.AssertCalled("Clean")
		mockBuild.AssertCalled("Default")

		// Verify call counts
		mockBuild.AssertCalledTimes("PreBuild", 1)
		mockBuild.AssertCalledTimes("Clean", 1)
		mockBuild.AssertCalledTimes("Default", 1)
	})

	t.Run("pre-build failure", func(t *testing.T) {
		mockBuild := NewMockBuild(t)
		mockBuild.SetMethodError("PreBuild", errMockPreBuildError)

		err := buildPipeline(mockBuild)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err.Error() != "pre-build failed: pre-build failure" {
			t.Errorf("Unexpected error message: %v", err)
		}

		// Only pre-build should have been called
		mockBuild.AssertCalled("PreBuild")
		mockBuild.AssertNotCalled("Clean")
		mockBuild.AssertNotCalled("Default")
	})
}

// Test interface compliance
func TestInterfaceCompliance(t *testing.T) {
	t.Run("mock build implements interface", func(t *testing.T) {
		var _ mage.BuildNamespace = (*MockBuild)(nil)
		t.Log("MockBuild implements BuildNamespace interface")
	})

	t.Run("mock test implements interface", func(t *testing.T) {
		var _ mage.TestNamespace = (*MockTest)(nil)
		t.Log("MockTest implements TestNamespace interface")
	})
}

// Benchmark tests
func BenchmarkMockBuild(b *testing.B) {
	mockBuild := NewMockBuild(nil) // No testing.T for benchmarks

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mockBuild.Default() //nolint:errcheck // Benchmark doesn't need error handling
	}
}

func BenchmarkRealNamespace(b *testing.B) {
	b.Skip("Skipping benchmark that calls actual build process - too slow for CI")

	build := mage.NewBuildNamespace()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Note: This may fail in test environment, but measures interface overhead
		_ = build.Default() //nolint:errcheck // Benchmark doesn't need error handling
	}
}
