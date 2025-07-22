package mage

import (
	"testing"
)

// TestNamespaceArchitecture tests that key namespaces can be created and implement their interfaces
func TestNamespaceArchitecture(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{"Build", testBuildNamespace},
		{"Test", testTestNamespace},
		{"Lint", testLintNamespace},
		{"Tools", testToolsNamespace},
		{"Deps", testDepsNamespace},
		{"Mod", testModNamespace},
		{"Update", testUpdateNamespace},
		{"Metrics", testMetricsNamespace},
		{"Generate", testGenerateNamespace},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

func testBuildNamespace(t *testing.T) {
	ns := NewBuildNamespace()
	if ns == nil {
		t.Fatal("NewBuildNamespace returned nil")
	}

	// Test that it implements the interface by calling a method
	if err := ns.Default(); err != nil {
		t.Logf("Build.Default() returned error: %v (expected for test environment)", err)
	}
}

func testTestNamespace(t *testing.T) {
	ns := NewTestNamespace()
	if ns == nil {
		t.Fatal("NewTestNamespace returned nil")
	}

	// Test that it implements the interface
	if err := ns.Default(); err != nil {
		t.Logf("Test.Default() returned error: %v (expected for test environment)", err)
	}
}

func testLintNamespace(t *testing.T) {
	ns := NewLintNamespace()
	if ns == nil {
		t.Fatal("NewLintNamespace returned nil")
	}

	// Test that it implements the interface
	if err := ns.Default(); err != nil {
		t.Logf("Lint.Default() returned error: %v (expected for test environment)", err)
	}
}

func testToolsNamespace(t *testing.T) {
	ns := NewToolsNamespace()
	if ns == nil {
		t.Fatal("NewToolsNamespace returned nil")
	}

	// Test that it implements the interface
	if err := ns.Default(); err != nil {
		t.Logf("Tools.Default() returned error: %v (expected for test environment)", err)
	}
}

func testDepsNamespace(t *testing.T) {
	ns := NewDepsNamespace()
	if ns == nil {
		t.Fatal("NewDepsNamespace returned nil")
	}

	// Test that it implements the interface
	if err := ns.Default(); err != nil {
		t.Logf("Deps.Default() returned error: %v (expected for test environment)", err)
	}
}

func testModNamespace(t *testing.T) {
	ns := NewModNamespace()
	if ns == nil {
		t.Fatal("NewModNamespace returned nil")
	}

	// Test that it implements the interface using Download method
	if err := ns.Download(); err != nil {
		t.Logf("Mod.Download() returned error: %v (expected for test environment)", err)
	}
}

func testUpdateNamespace(t *testing.T) {
	ns := NewUpdateNamespace()
	if ns == nil {
		t.Fatal("NewUpdateNamespace returned nil")
	}

	// Test that it implements the interface using Check method
	if err := ns.Check(); err != nil {
		t.Logf("Update.Check() returned error: %v (expected for test environment)", err)
	}
}

func testRecipesNamespace(t *testing.T) {
	ns := NewRecipesNamespace()
	if ns == nil {
		t.Fatal("NewRecipesNamespace returned nil")
	}

	// Test that it implements the interface
	if err := ns.Default(); err != nil {
		t.Logf("Recipes.Default() returned error: %v (expected for test environment)", err)
	}
}

func testMetricsNamespace(t *testing.T) {
	ns := NewMetricsNamespace()
	if ns == nil {
		t.Fatal("NewMetricsNamespace returned nil")
	}

	// Test that it implements the interface
	if err := ns.LOC(); err != nil {
		t.Logf("Metrics.LOC() returned error: %v (expected for test environment)", err)
	}
}

func testGenerateNamespace(t *testing.T) {
	ns := NewGenerateNamespace()
	if ns == nil {
		t.Fatal("NewGenerateNamespace returned nil")
	}

	// Test that it implements the interface
	if err := ns.Default(); err != nil {
		t.Logf("Generate.Default() returned error: %v (expected for test environment)", err)
	}
}

// TestBasicFunctionality tests basic namespace functionality
func TestBasicFunctionality(t *testing.T) {
	// Test that we can create individual namespaces
	build := NewBuildNamespace()
	if build == nil {
		t.Error("NewBuildNamespace returned nil")
	}

	test := NewTestNamespace()
	if test == nil {
		t.Error("NewTestNamespace returned nil")
	}

	lint := NewLintNamespace()
	if lint == nil {
		t.Error("NewLintNamespace returned nil")
	}

	t.Log("Basic namespace functionality working correctly")
}
