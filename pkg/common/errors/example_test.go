package errors_test

import (
	"fmt"

	"github.com/mrz1836/mage-x/pkg/common/errors"
)

func ExampleNewErrorBuilder() {
	// Create a rich error with the builder pattern
	err := errors.NewErrorBuilder().
		WithCode(errors.ErrBuildFailed).
		WithMessage("compilation failed").
		WithSeverity(errors.SeverityError).
		WithField("file", "main.go").
		WithField("line", 42).
		Build()

	fmt.Println(err.Error())
	// Output: compilation failed
}

func ExampleNewErrorBuilder_withCause() {
	// Create an error that wraps another error
	cause := fmt.Errorf("file not found") //nolint:err113 // Example code
	err := errors.NewErrorBuilder().
		WithCode(errors.ErrFileNotFound).
		WithMessage("failed to load configuration").
		WithCause(cause).
		WithOperation("LoadConfig").
		Build()

	// Check the error message (includes cause)
	fmt.Println(err.Error())
	// Output: failed to load configuration: file not found
}

func ExampleErrorChain() {
	// Create and chain multiple errors
	chain := errors.NewErrorChain()

	//nolint:errcheck,err113 // Example code - error returns and dynamic errors are acceptable
	chain.Add(fmt.Errorf("first error"))
	chain.Add(fmt.Errorf("second error")) //nolint:errcheck,err113 // Example code
	chain.Add(fmt.Errorf("third error"))  //nolint:errcheck,err113 // Example code

	fmt.Printf("Chain has %d errors\n", chain.Count())
	fmt.Printf("Has errors: %v\n", chain.Count() > 0)
	// Output:
	// Chain has 3 errors
	// Has errors: true
}

func ExampleSafeExecute() {
	// Execute code that might panic, returning an error instead
	err := errors.SafeExecute(func() error {
		// This could panic, but SafeExecute catches it
		return nil
	})

	if err == nil {
		fmt.Println("Execution completed safely")
	}
	// Output: Execution completed safely
}

func ExampleSafeExecute_withPanic() {
	// SafeExecute catches panics and converts them to errors
	err := errors.SafeExecute(func() error {
		panic("something went wrong")
	})
	if err != nil {
		fmt.Println("Panic was caught and converted to error")
	}
	// Output: Panic was caught and converted to error
}

func ExampleIsCritical() {
	// Create a critical error
	criticalErr := errors.NewErrorBuilder().
		WithMessage("database connection lost").
		WithSeverity(errors.SeverityCritical).
		Build()

	// Create a non-critical error
	normalErr := errors.NewErrorBuilder().
		WithMessage("cache miss").
		WithSeverity(errors.SeverityWarning).
		Build()

	fmt.Printf("Critical error is critical: %v\n", errors.IsCritical(criticalErr))
	fmt.Printf("Normal error is critical: %v\n", errors.IsCritical(normalErr))
	// Output:
	// Critical error is critical: true
	// Normal error is critical: false
}

func ExampleIsTimeout() {
	// Create a timeout error
	timeoutErr := errors.NewErrorBuilder().
		WithCode(errors.ErrTimeout).
		WithMessage("operation timed out").
		Build()

	fmt.Printf("Timeout error is timeout: %v\n", errors.IsTimeout(timeoutErr))
	// Output: Timeout error is timeout: true
}
