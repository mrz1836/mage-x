package main

import (
	"fmt"
	"testing"

	"github.com/mrz1836/go-mage/pkg/mage"
)

// MockBuild is a mock implementation of BuildNamespace for testing
type MockBuild struct {
	defaultCalled  bool
	allCalled      bool
	platformCalled map[string]bool
	cleanCalled    bool
	shouldFail     bool
	failureMessage string
	callOrder      []string
}

// NewMockBuild creates a new mock build namespace
func NewMockBuild() *MockBuild {
	return &MockBuild{
		platformCalled: make(map[string]bool),
		callOrder:      []string{},
	}
}

func (m *MockBuild) Default() error {
	m.defaultCalled = true
	m.callOrder = append(m.callOrder, "Default")
	if m.shouldFail {
		return fmt.Errorf("%s", m.failureMessage)
	}
	return nil
}

func (m *MockBuild) All() error {
	m.allCalled = true
	m.callOrder = append(m.callOrder, "All")
	if m.shouldFail {
		return fmt.Errorf("%s", m.failureMessage)
	}
	return nil
}

func (m *MockBuild) Platform(platform string) error {
	m.platformCalled[platform] = true
	m.callOrder = append(m.callOrder, fmt.Sprintf("Platform(%s)", platform))
	if m.shouldFail {
		return fmt.Errorf("%s", m.failureMessage)
	}
	return nil
}

func (m *MockBuild) Linux() error {
	m.callOrder = append(m.callOrder, "Linux")
	return m.Platform("linux/amd64")
}

func (m *MockBuild) Darwin() error {
	m.callOrder = append(m.callOrder, "Darwin")
	return m.Platform("darwin/amd64")
}

func (m *MockBuild) Windows() error {
	m.callOrder = append(m.callOrder, "Windows")
	return m.Platform("windows/amd64")
}

func (m *MockBuild) Docker() error {
	m.callOrder = append(m.callOrder, "Docker")
	if m.shouldFail {
		return fmt.Errorf("%s", m.failureMessage)
	}
	return nil
}

func (m *MockBuild) Clean() error {
	m.cleanCalled = true
	m.callOrder = append(m.callOrder, "Clean")
	if m.shouldFail {
		return fmt.Errorf("%s", m.failureMessage)
	}
	return nil
}

func (m *MockBuild) Install() error {
	m.callOrder = append(m.callOrder, "Install")
	if m.shouldFail {
		return fmt.Errorf("%s", m.failureMessage)
	}
	return nil
}

func (m *MockBuild) Generate() error {
	m.callOrder = append(m.callOrder, "Generate")
	if m.shouldFail {
		return fmt.Errorf("%s", m.failureMessage)
	}
	return nil
}

func (m *MockBuild) PreBuild() error {
	m.callOrder = append(m.callOrder, "PreBuild")
	if m.shouldFail {
		return fmt.Errorf("%s", m.failureMessage)
	}
	return nil
}

// MockTest is a mock implementation of TestNamespace for testing
type MockTest struct {
	mage.Test         // Embed the real Test implementation
	defaultCalled     bool
	unitCalled        bool
	integrationCalled bool
	coverageCalled    bool
	raceCalled        bool
	shouldFail        bool
	failureMessage    string
	callOrder         []string
}

func NewMockTest() *MockTest {
	return &MockTest{
		callOrder: []string{},
	}
}

// Only override the methods we want to track/mock
func (m *MockTest) Default() error {
	m.defaultCalled = true
	m.callOrder = append(m.callOrder, "Default")
	if m.shouldFail {
		return fmt.Errorf("%s", m.failureMessage)
	}
	return nil
}

func (m *MockTest) Unit() error {
	m.unitCalled = true
	m.callOrder = append(m.callOrder, "Unit")
	if m.shouldFail {
		return fmt.Errorf("%s", m.failureMessage)
	}
	return nil
}

func (m *MockTest) Integration() error {
	m.integrationCalled = true
	m.callOrder = append(m.callOrder, "Integration")
	if m.shouldFail {
		return fmt.Errorf("%s", m.failureMessage)
	}
	return nil
}

func (m *MockTest) Coverage(args ...string) error {
	m.coverageCalled = true
	m.callOrder = append(m.callOrder, "Coverage")
	if m.shouldFail {
		return fmt.Errorf("%s", m.failureMessage)
	}
	return nil
}

func (m *MockTest) Race() error {
	m.raceCalled = true
	m.callOrder = append(m.callOrder, "Race")
	if m.shouldFail {
		return fmt.Errorf("%s", m.failureMessage)
	}
	return nil
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
		mockBuild := NewMockBuild()
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

		// Verify call order
		expectedCalls := []string{
			"Platform(linux/amd64)",
			"Platform(darwin/amd64)",
			"Platform(windows/amd64)",
		}

		for i, expected := range expectedCalls {
			if i >= len(mockBuild.callOrder) || mockBuild.callOrder[i] != expected {
				t.Errorf("Expected call %d to be %s, got %v", i, expected, mockBuild.callOrder)
			}
		}
	})

	t.Run("build failure", func(t *testing.T) {
		mockBuild := NewMockBuild()
		mockBuild.shouldFail = true
		mockBuild.failureMessage = "mock build failure"
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
		mockBuild := NewMockBuild()
		mockTest := NewMockTest()

		err := runCI(mockBuild, mockTest)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify tests were run
		if !mockTest.coverageCalled {
			t.Error("Coverage tests were not called")
		}

		// Verify build was run
		if !mockBuild.defaultCalled {
			t.Error("Default build was not called")
		}

		// Verify order: tests should run before build
		if len(mockTest.callOrder) == 0 || mockTest.callOrder[0] != "Coverage" {
			t.Error("Coverage tests should run first")
		}

		if len(mockBuild.callOrder) == 0 || mockBuild.callOrder[0] != "Default" {
			t.Error("Default build should run after tests")
		}
	})

	t.Run("test failure", func(t *testing.T) {
		mockBuild := NewMockBuild()
		mockTest := NewMockTest()
		mockTest.shouldFail = true
		mockTest.failureMessage = "test failure"

		err := runCI(mockBuild, mockTest)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err.Error() != "tests failed: test failure" {
			t.Errorf("Unexpected error message: %v", err)
		}

		// Build should not have been called due to test failure
		if mockBuild.defaultCalled {
			t.Error("Build should not have been called when tests fail")
		}
	})

	t.Run("build failure", func(t *testing.T) {
		mockBuild := NewMockBuild()
		mockTest := NewMockTest()
		mockBuild.shouldFail = true
		mockBuild.failureMessage = "build failure"

		err := runCI(mockBuild, mockTest)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err.Error() != "build failed: build failure" {
			t.Errorf("Unexpected error message: %v", err)
		}

		// Tests should have been called
		if !mockTest.coverageCalled {
			t.Error("Tests should have been called before build failure")
		}
	})
}

func TestBuildPipeline(t *testing.T) {
	t.Run("successful pipeline", func(t *testing.T) {
		mockBuild := NewMockBuild()

		err := buildPipeline(mockBuild)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		// Verify all operations were called
		if !mockBuild.cleanCalled {
			t.Error("Clean was not called")
		}

		if !mockBuild.defaultCalled {
			t.Error("Default build was not called")
		}

		// Verify correct order
		expectedOrder := []string{"PreBuild", "Clean", "Default"}

		if len(mockBuild.callOrder) != len(expectedOrder) {
			t.Errorf("Expected %d calls, got %d", len(expectedOrder), len(mockBuild.callOrder))
		}

		for i, expected := range expectedOrder {
			if i >= len(mockBuild.callOrder) || mockBuild.callOrder[i] != expected {
				t.Errorf("Expected call %d to be %s, got %v", i, expected, mockBuild.callOrder)
			}
		}
	})

	t.Run("pre-build failure", func(t *testing.T) {
		mockBuild := NewMockBuild()
		mockBuild.shouldFail = true
		mockBuild.failureMessage = "pre-build failure"

		err := buildPipeline(mockBuild)

		if err == nil {
			t.Error("Expected error, got nil")
		}

		if err.Error() != "pre-build failed: pre-build failure" {
			t.Errorf("Unexpected error message: %v", err)
		}

		// Only pre-build should have been called
		expectedCalls := []string{"PreBuild"}
		if len(mockBuild.callOrder) != len(expectedCalls) {
			t.Errorf("Expected %d calls, got %d: %v", len(expectedCalls), len(mockBuild.callOrder), mockBuild.callOrder)
		}
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
	mockBuild := NewMockBuild()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mockBuild.Default()
	}
}

func BenchmarkRealNamespace(b *testing.B) {
	build := mage.NewBuildNamespace()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Note: This may fail in test environment, but measures interface overhead
		_ = build.Default()
	}
}
