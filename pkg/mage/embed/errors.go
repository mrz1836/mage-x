package embed

import "errors"

// Test-related sentinel errors
var (
	// ErrRegistrationIncomplete indicates that command registration did not complete successfully
	ErrRegistrationIncomplete = errors.New("registration not completed")

	// ErrInsufficientCommands indicates that fewer commands were registered than expected
	ErrInsufficientCommands = errors.New("insufficient commands registered")

	// ErrWorkerPanic indicates that a worker goroutine panicked during execution
	ErrWorkerPanic = errors.New("worker panicked")
)
