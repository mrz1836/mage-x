// Package testutil provides testing utilities and helpers for mage operations,
// including mock runners, test environments, and command matchers.
//
// # Test Environment
//
// Create isolated test environments with temporary directories:
//
//	func TestMyCommand(t *testing.T) {
//	    env := testutil.NewTestEnvironment(t)
//	    defer env.Cleanup()
//
//	    // Test runs in isolated temp directory
//	    // with mock runner configured
//	}
//
// # Mock Runner
//
// Mock external command execution:
//
//	runner, builder := testutil.NewMockRunner()
//
//	// Configure expected commands
//	builder.
//	    ForCommand("go", "build").
//	    Returns("", nil)
//
//	// Run code under test
//	err := myFunction()
//
//	// Verify expectations
//	runner.Verify(t)
//
// # Command Matcher
//
// Match and verify command invocations:
//
//	matcher := testutil.NewCommandMatcher()
//	matcher.Expect("go", "test", "-v")
//	matcher.Record("go", "test", "-v", "./...")
//	// Matcher checks that recorded commands match expectations
//
// # Mock OS Operations
//
// Mock file system operations for testing:
//
//	mockOS := testutil.NewMockOSOperations(ctrl)
//	mockOS.EXPECT().Stat(gomock.Any()).Return(nil, os.ErrNotExist)
//
// # Mock Go Operations
//
// Mock Go toolchain operations:
//
//	mockGo := testutil.NewMockGoOperations(ctrl)
//	mockGo.EXPECT().Build(gomock.Any()).Return(nil)
//
// # Testing Interface
//
// The TestingInterface allows use with both *testing.T and *testing.B:
//
//	type TestingInterface interface {
//	    TempDir() string
//	    Helper()
//	    Fatalf(format string, args ...interface{})
//	}
package testutil
