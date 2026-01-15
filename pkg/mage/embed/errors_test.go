package embed

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestErrorConstants verifies that all error constants are defined correctly
func TestErrorConstants(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantMsg string
		wantNil bool
	}{
		{
			name:    "ErrRegistrationIncomplete",
			err:     ErrRegistrationIncomplete,
			wantMsg: "registration not completed",
			wantNil: false,
		},
		{
			name:    "ErrInsufficientCommands",
			err:     ErrInsufficientCommands,
			wantMsg: "insufficient commands registered",
			wantNil: false,
		},
		{
			name:    "ErrWorkerPanic",
			err:     ErrWorkerPanic,
			wantMsg: "worker panicked",
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantNil {
				assert.NoError(t, tt.err)
			} else {
				require.Error(t, tt.err, "error should not be nil")
				assert.Equal(t, tt.wantMsg, tt.err.Error(), "error message should match")
			}
		})
	}
}

// TestErrorInterface verifies that errors implement the error interface
func TestErrorInterface(t *testing.T) {
	_ = ErrRegistrationIncomplete
	_ = ErrInsufficientCommands
	_ = ErrWorkerPanic
}

// TestErrorEquality verifies that error constants can be compared with errors.Is
func TestErrorEquality(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		target error
		wantIs bool
	}{
		{
			name:   "ErrRegistrationIncomplete matches itself",
			err:    ErrRegistrationIncomplete,
			target: ErrRegistrationIncomplete,
			wantIs: true,
		},
		{
			name:   "ErrInsufficientCommands matches itself",
			err:    ErrInsufficientCommands,
			target: ErrInsufficientCommands,
			wantIs: true,
		},
		{
			name:   "ErrWorkerPanic matches itself",
			err:    ErrWorkerPanic,
			target: ErrWorkerPanic,
			wantIs: true,
		},
		{
			name:   "Different errors don't match",
			err:    ErrRegistrationIncomplete,
			target: ErrInsufficientCommands,
			wantIs: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := errors.Is(tt.err, tt.target)
			assert.Equal(t, tt.wantIs, result)
		})
	}
}

// TestErrorUniqueness verifies that each error constant is unique
func TestErrorUniqueness(t *testing.T) {
	// Collect all errors in a slice
	allErrors := []error{
		ErrRegistrationIncomplete,
		ErrInsufficientCommands,
		ErrWorkerPanic,
	}

	// Check that each error is unique
	for i, err1 := range allErrors {
		for j, err2 := range allErrors {
			if i != j {
				assert.NotEqual(t, err1, err2, "errors at index %d and %d should be different", i, j)
			}
		}
	}
}
