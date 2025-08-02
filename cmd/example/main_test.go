package main

import (
	"bytes"
	"testing"

	"github.com/mrz1836/mage-x/pkg/utils"
	"github.com/stretchr/testify/suite"
)

type ExampleTestSuite struct {
	suite.Suite
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(ExampleTestSuite))
}

// Test main function execution
func (suite *ExampleTestSuite) TestMain() {
	// Capture output from default logger like in the utils test
	var buf bytes.Buffer
	originalLogger := utils.DefaultLogger
	utils.DefaultLogger = utils.NewLogger()
	utils.DefaultLogger.SetOutput(&buf)
	utils.DefaultLogger.SetColorEnabled(false)
	defer func() { utils.DefaultLogger = originalLogger }()

	// Execute main function
	main()

	// Verify output
	output := buf.String()
	suite.Contains(output, "mage-x example")
}

// Test main function with different scenarios
func (suite *ExampleTestSuite) TestMainOutput() {
	// This test verifies that main() function calls utils.Info correctly
	var buf bytes.Buffer
	originalLogger := utils.DefaultLogger
	utils.DefaultLogger = utils.NewLogger()
	utils.DefaultLogger.SetOutput(&buf)
	utils.DefaultLogger.SetColorEnabled(false)
	defer func() { utils.DefaultLogger = originalLogger }()

	// Execute main function
	main()

	// Verify the expected message is printed
	output := buf.String()
	suite.Contains(output, "mage-x example")
}

// Test that main function doesn't panic
func (suite *ExampleTestSuite) TestMainNoPanic() {
	suite.NotPanics(func() {
		main()
	})
}

// Test main function execution multiple times (idempotency)
func (suite *ExampleTestSuite) TestMainIdempotent() {
	// Test that calling main multiple times doesn't cause issues
	suite.NotPanics(func() {
		main()
		main()
		main()
	})
}

// Test package structure and imports
func (suite *ExampleTestSuite) TestPackageStructure() {
	// This is more of a documentation test to verify the package is correctly structured
	// We test that the expected functions exist and are callable

	// Test that main function exists (implicitly tested by calling it)
	suite.NotPanics(func() {
		main()
	})
}

// Test main with concurrent execution
func (suite *ExampleTestSuite) TestMainConcurrent() {
	// Test that main() can be called concurrently without issues
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					suite.Fail("main() panicked during concurrent execution")
				}
				done <- true
			}()
			main()
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}
}

// Test main function with logger capture
func (suite *ExampleTestSuite) TestMainLoggerCapture() {
	var buf bytes.Buffer
	originalLogger := utils.DefaultLogger

	// Create new logger with buffer output
	testLogger := utils.NewLogger()
	testLogger.SetOutput(&buf)
	testLogger.SetColorEnabled(false)
	testLogger.SetLevel(utils.LogLevelInfo)

	utils.DefaultLogger = testLogger
	defer func() { utils.DefaultLogger = originalLogger }()

	// Execute main function
	main()

	// Verify output contains expected message
	output := buf.String()
	suite.Contains(output, "mage-x example")
	suite.Contains(output, "INFO")
}

// Test that main function calls utils.Info exactly once
func (suite *ExampleTestSuite) TestMainSingleInfoCall() {
	var buf bytes.Buffer
	originalLogger := utils.DefaultLogger
	utils.DefaultLogger = utils.NewLogger()
	utils.DefaultLogger.SetOutput(&buf)
	utils.DefaultLogger.SetColorEnabled(false)
	defer func() { utils.DefaultLogger = originalLogger }()

	// Execute main function
	main()

	// Count occurrences of the message
	output := buf.String()
	suite.Contains(output, "mage-x example")

	// Simple check that it's called once (not multiple times)
	firstIndex := bytes.Index(buf.Bytes(), []byte("mage-x example"))
	suite.NotEqual(-1, firstIndex, "Should contain the message")

	// Check there's no second occurrence after the first
	remaining := buf.Bytes()[firstIndex+len("mage-x example"):]
	secondIndex := bytes.Index(remaining, []byte("mage-x example"))
	suite.Equal(-1, secondIndex, "Should only contain the message once")
}
