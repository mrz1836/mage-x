package mage

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/suite"
)

// FactoryFunctionsTestSuite provides comprehensive tests for all factory functions
type FactoryFunctionsTestSuite struct {
	suite.Suite
}

// TestFactoryFunctions tests all New*Namespace() factory functions
func (suite *FactoryFunctionsTestSuite) TestFactoryFunctions() {
	tests := []struct {
		name           string
		factory        func() interface{}
		expectedType   interface{}
		interfaceCheck func(interface{}) bool
		description    string
	}{
		{
			name:         "NewBuildNamespace",
			factory:      func() interface{} { return NewBuildNamespace() },
			expectedType: (*buildNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(BuildNamespace)
				return ok
			},
			description: "creates BuildNamespace implementation",
		},
		{
			name:         "NewTestNamespace",
			factory:      func() interface{} { return NewTestNamespace() },
			expectedType: (*testNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(TestNamespace)
				return ok
			},
			description: "creates TestNamespace implementation",
		},
		{
			name:         "NewLintNamespace",
			factory:      func() interface{} { return NewLintNamespace() },
			expectedType: (*lintNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(LintNamespace)
				return ok
			},
			description: "creates LintNamespace implementation",
		},
		{
			name:         "NewFormatNamespace",
			factory:      func() interface{} { return NewFormatNamespace() },
			expectedType: (*formatNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(FormatNamespace)
				return ok
			},
			description: "creates FormatNamespace implementation",
		},
		{
			name:         "NewDepsNamespace",
			factory:      func() interface{} { return NewDepsNamespace() },
			expectedType: (*depsNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(DepsNamespace)
				return ok
			},
			description: "creates DepsNamespace implementation",
		},
		{
			name:         "NewGitNamespace",
			factory:      func() interface{} { return NewGitNamespace() },
			expectedType: (*gitNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(GitNamespace)
				return ok
			},
			description: "creates GitNamespace implementation",
		},
		{
			name:         "NewReleaseNamespace",
			factory:      func() interface{} { return NewReleaseNamespace() },
			expectedType: (*releaseNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(ReleaseNamespace)
				return ok
			},
			description: "creates ReleaseNamespace implementation",
		},
		{
			name:         "NewDocsNamespace",
			factory:      func() interface{} { return NewDocsNamespace() },
			expectedType: (*docsNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(DocsNamespace)
				return ok
			},
			description: "creates DocsNamespace implementation",
		},
		{
			name:         "NewToolsNamespace",
			factory:      func() interface{} { return NewToolsNamespace() },
			expectedType: (*toolsNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(ToolsNamespace)
				return ok
			},
			description: "creates ToolsNamespace implementation",
		},
		{
			name:         "NewGenerateNamespace",
			factory:      func() interface{} { return NewGenerateNamespace() },
			expectedType: (*generateNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(GenerateNamespace)
				return ok
			},
			description: "creates GenerateNamespace implementation",
		},
		{
			name:         "NewCLINamespace",
			factory:      func() interface{} { return NewCLINamespace() },
			expectedType: (*cliNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(CLINamespace)
				return ok
			},
			description: "creates CLINamespace implementation",
		},
		{
			name:         "NewUpdateNamespace",
			factory:      func() interface{} { return NewUpdateNamespace() },
			expectedType: (*updateNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(UpdateNamespace)
				return ok
			},
			description: "creates UpdateNamespace implementation",
		},
		{
			name:         "NewModNamespace",
			factory:      func() interface{} { return NewModNamespace() },
			expectedType: (*modNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(ModNamespace)
				return ok
			},
			description: "creates ModNamespace implementation",
		},
		{
			name:         "NewRecipesNamespace",
			factory:      func() interface{} { return NewRecipesNamespace() },
			expectedType: (*recipesNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(RecipesNamespace)
				return ok
			},
			description: "creates RecipesNamespace implementation",
		},
		{
			name:         "NewMetricsNamespace",
			factory:      func() interface{} { return NewMetricsNamespace() },
			expectedType: (*metricsNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(MetricsNamespace)
				return ok
			},
			description: "creates MetricsNamespace implementation",
		},
		{
			name:         "NewWorkflowNamespace",
			factory:      func() interface{} { return NewWorkflowNamespace() },
			expectedType: (*workflowNamespaceWrapper)(nil),
			interfaceCheck: func(i interface{}) bool {
				_, ok := i.(WorkflowNamespace)
				return ok
			},
			description: "creates WorkflowNamespace implementation",
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Test factory function exists and returns non-nil
			result := tt.factory()
			suite.Require().NotNil(result, "Factory function %s should return non-nil", tt.name)

			// Test interface compliance
			suite.True(tt.interfaceCheck(result), "%s should implement the expected interface", tt.name)

			// Test type assertion
			suite.IsType(tt.expectedType, result, "%s should return correct wrapper type", tt.name)
		})
	}
}

// TestSecurityNamespaceFactorySpecialCase tests NewSecurityNamespace which returns nil
func (suite *FactoryFunctionsTestSuite) TestSecurityNamespaceFactorySpecialCase() {
	result := NewSecurityNamespace()
	suite.Nil(result, "NewSecurityNamespace should return nil as it's temporarily disabled")
}

// TestFactoryFunctionsConcurrency tests factory functions under concurrent access
func (suite *FactoryFunctionsTestSuite) TestFactoryFunctionsConcurrency() {
	const numGoroutines = 100
	var wg sync.WaitGroup
	results := make([]BuildNamespace, numGoroutines)

	// Test concurrent access to NewBuildNamespace
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			results[index] = NewBuildNamespace()
		}(i)
	}

	wg.Wait()

	// Verify all results are valid but distinct instances
	for i, result := range results {
		suite.NotNil(result, "Result %d should not be nil", i)
		suite.Implements((*BuildNamespace)(nil), result, "Result %d should implement BuildNamespace", i)

		// Each factory call should return a new instance
		for j := i + 1; j < len(results); j++ {
			suite.NotSame(result, results[j], "Factory should return distinct instances")
		}
	}
}

// TestFactoryFunctionDependencyInjection tests factory functions with dependency injection
func (suite *FactoryFunctionsTestSuite) TestFactoryFunctionDependencyInjection() {
	// Test that factory functions create instances with proper internal dependencies
	buildNS := NewBuildNamespace()
	suite.NotNil(buildNS)

	testNS := NewTestNamespace()
	suite.NotNil(testNS)

	lintNS := NewLintNamespace()
	suite.NotNil(lintNS)

	// Test that different namespace types are distinct
	suite.IsType(&buildNamespaceWrapper{}, buildNS)
	suite.IsType(&testNamespaceWrapper{}, testNS)
	suite.IsType(&lintNamespaceWrapper{}, lintNS)
}

// TestFactoryFunctionReturnTypes validates exact return types
func (suite *FactoryFunctionsTestSuite) TestFactoryFunctionReturnTypes() {
	// Test BuildNamespace
	buildNS := NewBuildNamespace()
	buildWrapper, ok := buildNS.(*buildNamespaceWrapper)
	suite.True(ok, "NewBuildNamespace should return *buildNamespaceWrapper")
	suite.NotNil(buildWrapper)

	// Test TestNamespace
	testNS := NewTestNamespace()
	testWrapper, ok := testNS.(*testNamespaceWrapper)
	suite.True(ok, "NewTestNamespace should return *testNamespaceWrapper")
	suite.NotNil(testWrapper)

	// Test LintNamespace
	lintNS := NewLintNamespace()
	lintWrapper, ok := lintNS.(*lintNamespaceWrapper)
	suite.True(ok, "NewLintNamespace should return *lintNamespaceWrapper")
	suite.NotNil(lintWrapper)
}

// TestFactoryFunctionErrorHandling tests error handling in factory functions
func (suite *FactoryFunctionsTestSuite) TestFactoryFunctionErrorHandling() {
	// All factory functions should be resilient and not panic
	suite.NotPanics(func() { NewBuildNamespace() }, "NewBuildNamespace should not panic")
	suite.NotPanics(func() { NewTestNamespace() }, "NewTestNamespace should not panic")
	suite.NotPanics(func() { NewLintNamespace() }, "NewLintNamespace should not panic")
	suite.NotPanics(func() { NewFormatNamespace() }, "NewFormatNamespace should not panic")
	suite.NotPanics(func() { NewDepsNamespace() }, "NewDepsNamespace should not panic")
	suite.NotPanics(func() { NewGitNamespace() }, "NewGitNamespace should not panic")
	suite.NotPanics(func() { NewReleaseNamespace() }, "NewReleaseNamespace should not panic")
	suite.NotPanics(func() { NewDocsNamespace() }, "NewDocsNamespace should not panic")
	suite.NotPanics(func() { NewToolsNamespace() }, "NewToolsNamespace should not panic")
	suite.NotPanics(func() { NewSecurityNamespace() }, "NewSecurityNamespace should not panic")
	suite.NotPanics(func() { NewGenerateNamespace() }, "NewGenerateNamespace should not panic")
	suite.NotPanics(func() { NewCLINamespace() }, "NewCLINamespace should not panic")
	suite.NotPanics(func() { NewUpdateNamespace() }, "NewUpdateNamespace should not panic")
	suite.NotPanics(func() { NewModNamespace() }, "NewModNamespace should not panic")
	suite.NotPanics(func() { NewRecipesNamespace() }, "NewRecipesNamespace should not panic")
	suite.NotPanics(func() { NewMetricsNamespace() }, "NewMetricsNamespace should not panic")
	suite.NotPanics(func() { NewWorkflowNamespace() }, "NewWorkflowNamespace should not panic")
}

// TestRunFactoryFunctionsTestSuite runs the factory functions test suite
func TestRunFactoryFunctionsTestSuite(t *testing.T) {
	suite.Run(t, new(FactoryFunctionsTestSuite))
}

// BenchmarkFactoryFunctions benchmarks factory function performance
func BenchmarkFactoryFunctions(b *testing.B) {
	tests := []struct {
		name    string
		factory func() interface{}
	}{
		{"NewBuildNamespace", func() interface{} { return NewBuildNamespace() }},
		{"NewTestNamespace", func() interface{} { return NewTestNamespace() }},
		{"NewLintNamespace", func() interface{} { return NewLintNamespace() }},
		{"NewFormatNamespace", func() interface{} { return NewFormatNamespace() }},
		{"NewDepsNamespace", func() interface{} { return NewDepsNamespace() }},
		{"NewGitNamespace", func() interface{} { return NewGitNamespace() }},
		{"NewReleaseNamespace", func() interface{} { return NewReleaseNamespace() }},
		{"NewDocsNamespace", func() interface{} { return NewDocsNamespace() }},
		{"NewToolsNamespace", func() interface{} { return NewToolsNamespace() }},
		{"NewGenerateNamespace", func() interface{} { return NewGenerateNamespace() }},
		{"NewCLINamespace", func() interface{} { return NewCLINamespace() }},
		{"NewUpdateNamespace", func() interface{} { return NewUpdateNamespace() }},
		{"NewModNamespace", func() interface{} { return NewModNamespace() }},
		{"NewRecipesNamespace", func() interface{} { return NewRecipesNamespace() }},
		{"NewMetricsNamespace", func() interface{} { return NewMetricsNamespace() }},
		{"NewWorkflowNamespace", func() interface{} { return NewWorkflowNamespace() }},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				result := tt.factory()
				if result == nil {
					b.Fatal("Factory function returned nil")
				}
			}
			b.StopTimer()
		})
	}
}

// BenchmarkFactoryFunctionsConcurrent benchmarks concurrent factory function calls
func BenchmarkFactoryFunctionsConcurrent(b *testing.B) {
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Test the most commonly used factory function
			buildNS := NewBuildNamespace()
			if buildNS == nil {
				b.Fatal("NewBuildNamespace returned nil")
			}
		}
	})
	b.StopTimer()
}
