# Unit Testing with Namespace Interfaces

This example demonstrates how to write comprehensive unit tests for functions that use Go-Mage namespace interfaces. It shows how to create mock implementations and verify behavior without executing actual build operations.

## Overview

The example includes:
- Mock implementations of `BuildNamespace` and `TestNamespace` interfaces
- Unit tests for functions that use these interfaces
- Examples of testing error conditions and call ordering
- Interface compliance verification
- Performance benchmarks

## Key Features

- **Mock Implementations**: Complete mock namespaces for testing
- **Call Tracking**: Track which methods were called and in what order
- **Error Simulation**: Test error conditions without real failures
- **Interface Compliance**: Compile-time verification of mock interfaces
- **Comprehensive Coverage**: Test success and failure scenarios

## Files

- `namespace_test.go` - Complete test suite with mocks and tests
- `README.md` - This documentation

## Mock Features

### MockBuild

```go
type MockBuild struct {
    defaultCalled    bool
    allCalled        bool
    platformCalled   map[string]bool
    cleanCalled      bool
    shouldFail       bool
    failureMessage   string
    callOrder        []string
}
```

Features:
- Tracks all method calls
- Records call order
- Can simulate failures
- Platform-specific tracking
- Zero side effects

### MockTest

```go
type MockTest struct {
    defaultCalled     bool
    unitCalled        bool
    integrationCalled bool
    coverageCalled    bool
    raceCalled        bool
    shouldFail        bool
    failureMessage    string
    callOrder         []string
}
```

Features:
- Comprehensive test method tracking
- Configurable failure modes
- Call order verification
- Isolated test execution

## Example Tests

### Testing Function Behavior

```go
func TestDeployApp(t *testing.T) {
    mockBuild := NewMockBuild()
    platforms := []string{"linux/amd64", "darwin/amd64"}
    
    err := deployApp(mockBuild, platforms)
    
    // Verify no errors
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    
    // Verify all platforms were built
    for _, platform := range platforms {
        if !mockBuild.platformCalled[platform] {
            t.Errorf("Platform %s was not built", platform)
        }
    }
}
```

### Testing Error Conditions

```go
func TestRunCI_TestFailure(t *testing.T) {
    mockBuild := NewMockBuild()
    mockTest := NewMockTest()
    mockTest.shouldFail = true
    mockTest.failureMessage = "test failure"
    
    err := runCI(mockBuild, mockTest)
    
    if err == nil {
        t.Error("Expected error, got nil")
    }
    
    // Build should not have been called due to test failure
    if mockBuild.defaultCalled {
        t.Error("Build should not have been called when tests fail")
    }
}
```

### Testing Call Order

```go
func TestBuildPipeline(t *testing.T) {
    mockBuild := NewMockBuild()
    
    err := buildPipeline(mockBuild)
    
    // Verify correct order
    expectedOrder := []string{"PreBuild", "Clean", "Default"}
    
    for i, expected := range expectedOrder {
        if mockBuild.callOrder[i] != expected {
            t.Errorf("Expected call %d to be %s, got %s", 
                i, expected, mockBuild.callOrder[i])
        }
    }
}
```

## Running Tests

```bash
# Run all tests
go test -v

# Run specific test
go test -v -run TestDeployApp

# Run with coverage
go test -v -cover

# Run benchmarks
go test -v -bench=.

# Race condition detection
go test -v -race
```

## Test Categories

### 1. Success Path Tests
- Verify functions work correctly with valid inputs
- Check all expected methods are called
- Validate call ordering

### 2. Error Condition Tests
- Test behavior when namespaces return errors
- Verify error propagation and handling
- Check that execution stops appropriately

### 3. Interface Compliance Tests
- Compile-time verification that mocks implement interfaces
- Runtime type assertion tests
- Behavioral contract verification

### 4. Performance Tests
- Benchmark mock vs real namespace performance
- Measure interface call overhead
- Performance regression detection

## Best Practices Demonstrated

### 1. Mock Design
```go
// Good: Track what was called
type MockBuild struct {
    defaultCalled bool
    callOrder     []string
}

// Better: Track parameters too
func (m *MockBuild) Platform(platform string) error {
    m.platformCalled[platform] = true
    return nil
}
```

### 2. Error Testing
```go
// Test both success and failure
mockBuild.shouldFail = true
mockBuild.failureMessage = "specific error"
```

### 3. Interface Verification
```go
// Compile-time check
var _ mage.BuildNamespace = (*MockBuild)(nil)

// Runtime check
if _, ok := mock.(mage.BuildNamespace); !ok {
    t.Error("Mock doesn't implement interface")
}
```

### 4. Comprehensive Coverage
```go
func TestAllScenarios(t *testing.T) {
    testCases := []struct{
        name           string
        shouldFail     bool
        expectedError  string
        expectedCalls  []string
    }{
        {"success", false, "", []string{"PreBuild", "Build"}},
        {"failure", true, "build failed", []string{"PreBuild"}},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Testing Patterns

### 1. Arrange-Act-Assert
```go
func TestExample(t *testing.T) {
    // Arrange
    mock := NewMockBuild()
    
    // Act
    err := functionUnderTest(mock)
    
    // Assert
    if err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
}
```

### 2. Table-Driven Tests
```go
func TestMultipleScenarios(t *testing.T) {
    tests := []struct{
        name     string
        input    string
        expected bool
    }{
        {"case1", "input1", true},
        {"case2", "input2", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test logic
        })
    }
}
```

### 3. Subtests for Organization
```go
func TestDeployApp(t *testing.T) {
    t.Run("successful deployment", func(t *testing.T) {
        // Success test
    })
    
    t.Run("build failure", func(t *testing.T) {
        // Failure test
    })
}
```

## Benefits

### 1. Fast Tests
- No actual build operations
- Predictable execution time
- No external dependencies

### 2. Reliable Tests
- No flaky failures from environment
- Consistent results across machines
- Isolated test execution

### 3. Comprehensive Coverage
- Test error conditions safely
- Verify complex call patterns
- Test edge cases easily

### 4. Clear Intent
- Tests document expected behavior
- Mock calls show function requirements
- Error tests show failure handling

## Next Steps

After mastering unit testing:
- [Integration Testing](../integration-testing/) - Testing with real namespaces
- [Mock Implementations](../mock-implementations/) - Advanced mock patterns
- [Custom Implementations](../../custom/) - Building custom namespaces