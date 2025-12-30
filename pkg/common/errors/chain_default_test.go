package errors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Static test errors for DefaultChainError tests (err113 compliant)
var (
	errDefaultChainTest1   = errors.New("default chain error 1")
	errDefaultChainTest2   = errors.New("default chain error 2")
	errDefaultChainTest3   = errors.New("default chain error 3")
	errDefaultContextError = errors.New("default context error")
	errDefaultFilterTest   = errors.New("filter test error")
	errDefaultForEachTest  = errors.New("foreach test error")
	errDefaultForEachStop  = errors.New("stop iteration")
)

// TestDefaultChainErrorAdd tests the Add method of DefaultChainError
func TestDefaultChainErrorAdd(t *testing.T) {
	t.Run("nil error is ignored", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		result := chain.Add(nil)

		// DefaultChainError creates a new chain, so we need to check the result
		assert.Equal(t, 0, result.Count())
	})

	t.Run("single error", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		result := chain.Add(errDefaultChainTest1)

		assert.Equal(t, 1, result.Count())
		assert.Equal(t, errDefaultChainTest1.Error(), result.First().Error())
	})

	t.Run("multiple errors creates new RealDefaultChainError", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		chain1 := chain.Add(errDefaultChainTest1)
		chain2 := chain1.Add(errDefaultChainTest2)
		chain3 := chain2.Add(errDefaultChainTest3)

		// Original DefaultChainError is unchanged
		assert.Equal(t, 0, chain.Count())

		// DefaultChainError.Add() returns RealDefaultChainError which mutates in place
		// So chain1, chain2, chain3 all point to the same RealDefaultChainError
		// after all mutations
		assert.Equal(t, 3, chain1.Count())
		assert.Equal(t, 3, chain2.Count())
		assert.Equal(t, 3, chain3.Count())
	})

	t.Run("preserves existing errors", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{errDefaultChainTest1}}
		result := chain.Add(errDefaultChainTest2)

		errs := result.Errors()
		require.Len(t, errs, 2)
		assert.Equal(t, errDefaultChainTest1.Error(), errs[0].Error())
		assert.Equal(t, errDefaultChainTest2.Error(), errs[1].Error())
	})
}

// TestDefaultChainErrorAddWithContext tests the AddWithContext method
func TestDefaultChainErrorAddWithContext(t *testing.T) {
	t.Run("nil error is ignored", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		ctx := &ErrorContext{Operation: "test"}
		result := chain.AddWithContext(nil, ctx)

		assert.Equal(t, 0, result.Count())
	})

	t.Run("standard error with context", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		ctx := &ErrorContext{
			Operation: "TestOperation",
			Resource:  "TestResource",
			User:      "TestUser",
		}
		result := chain.AddWithContext(errDefaultContextError, ctx)

		require.Equal(t, 1, result.Count())
		errs := result.Errors()
		require.Len(t, errs, 1)

		// Should be wrapped as MageError
		var mageErr MageError
		require.ErrorAs(t, errs[0], &mageErr)
		assert.Contains(t, mageErr.Error(), "default context error")
		assert.Equal(t, "TestOperation", mageErr.Context().Operation)
		assert.Equal(t, "TestResource", mageErr.Context().Resource)
		assert.Equal(t, "TestUser", mageErr.Context().User)
	})

	t.Run("nil context", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		result := chain.AddWithContext(errDefaultContextError, nil)

		assert.Equal(t, 1, result.Count())
	})

	t.Run("MageError preserves type", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		mageErr := NewErrorBuilder().
			WithCode(ErrInternal).
			WithMessage("mage test error").
			Build()

		ctx := &ErrorContext{Operation: "NewOp"}
		result := chain.AddWithContext(mageErr, ctx)

		require.Equal(t, 1, result.Count())
		errs := result.Errors()
		require.Len(t, errs, 1)

		var resultMageErr MageError
		require.ErrorAs(t, errs[0], &resultMageErr)
		assert.Equal(t, "NewOp", resultMageErr.Context().Operation)
	})

	t.Run("preserves existing errors immutably", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{errDefaultChainTest1}}
		ctx := &ErrorContext{Operation: "Op2"}
		result := chain.AddWithContext(errDefaultChainTest2, ctx)

		// Original unchanged
		assert.Equal(t, 1, chain.Count())

		// New chain has both
		assert.Equal(t, 2, result.Count())
	})
}

// TestDefaultChainErrorError tests the Error method
func TestDefaultChainErrorError(t *testing.T) {
	t.Run("empty chain", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		assert.Equal(t, "no errors", chain.Error())
	})

	t.Run("single error", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{errDefaultChainTest1}}
		assert.Equal(t, "default chain error 1", chain.Error())
	})

	t.Run("multiple errors", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{
			errDefaultChainTest1,
			errDefaultChainTest2,
			errDefaultChainTest3,
		}}

		errStr := chain.Error()
		assert.Contains(t, errStr, "3 errors occurred")
		assert.Contains(t, errStr, "[1] default chain error 1")
		assert.Contains(t, errStr, "[2] default chain error 2")
		assert.Contains(t, errStr, "[3] default chain error 3")
	})
}

// TestDefaultChainErrorErrors tests the Errors method
func TestDefaultChainErrorErrors(t *testing.T) {
	t.Run("empty chain returns empty slice", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		errs := chain.Errors()

		assert.Empty(t, errs)
		assert.NotNil(t, errs) // Should be non-nil empty slice
	})

	t.Run("returns all errors", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{
			errDefaultChainTest1,
			errDefaultChainTest2,
		}}

		errs := chain.Errors()
		require.Len(t, errs, 2)
		assert.Equal(t, errDefaultChainTest1, errs[0])
		assert.Equal(t, errDefaultChainTest2, errs[1])
	})

	t.Run("returns reference to internal slice", func(t *testing.T) {
		// DefaultChainError.Errors() returns the internal slice directly
		originalErrs := []error{errDefaultChainTest1}
		chain := &DefaultChainError{errors: originalErrs}

		errs := chain.Errors()
		assert.Equal(t, originalErrs, errs)
	})
}

// TestDefaultChainErrorFirst tests the First method
func TestDefaultChainErrorFirst(t *testing.T) {
	t.Run("empty chain returns nil", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		assert.NoError(t, chain.First())
	})

	t.Run("single error", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{errDefaultChainTest1}}
		assert.Equal(t, errDefaultChainTest1, chain.First())
	})

	t.Run("multiple errors returns first", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{
			errDefaultChainTest1,
			errDefaultChainTest2,
			errDefaultChainTest3,
		}}
		assert.Equal(t, errDefaultChainTest1, chain.First())
	})
}

// TestDefaultChainErrorLast tests the Last method
func TestDefaultChainErrorLast(t *testing.T) {
	t.Run("empty chain returns nil", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		assert.NoError(t, chain.Last())
	})

	t.Run("single error", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{errDefaultChainTest1}}
		assert.Equal(t, errDefaultChainTest1, chain.Last())
	})

	t.Run("multiple errors returns last", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{
			errDefaultChainTest1,
			errDefaultChainTest2,
			errDefaultChainTest3,
		}}
		assert.Equal(t, errDefaultChainTest3, chain.Last())
	})
}

// TestDefaultChainErrorCount tests the Count method
func TestDefaultChainErrorCount(t *testing.T) {
	tests := []struct {
		name   string
		errors []error
		want   int
	}{
		{
			name:   "empty chain",
			errors: []error{},
			want:   0,
		},
		{
			name:   "one error",
			errors: []error{errDefaultChainTest1},
			want:   1,
		},
		{
			name:   "multiple errors",
			errors: []error{errDefaultChainTest1, errDefaultChainTest2, errDefaultChainTest3},
			want:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := &DefaultChainError{errors: tt.errors}
			assert.Equal(t, tt.want, chain.Count())
		})
	}
}

// TestDefaultChainErrorHasError tests the HasError method
func TestDefaultChainErrorHasError(t *testing.T) {
	t.Run("empty chain returns false", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		assert.False(t, chain.HasError(ErrInternal))
	})

	t.Run("chain with non-MageError returns false", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{errDefaultChainTest1}}
		assert.False(t, chain.HasError(ErrInternal))
	})

	t.Run("chain with MageError code found", func(t *testing.T) {
		mageErr := NewErrorBuilder().
			WithCode(ErrInternal).
			WithMessage("internal error").
			Build()
		chain := &DefaultChainError{errors: []error{mageErr}}

		assert.True(t, chain.HasError(ErrInternal))
	})

	t.Run("chain with MageError code not found", func(t *testing.T) {
		mageErr := NewErrorBuilder().
			WithCode(ErrInternal).
			WithMessage("internal error").
			Build()
		chain := &DefaultChainError{errors: []error{mageErr}}

		assert.False(t, chain.HasError(ErrNotFound))
	})

	t.Run("mixed errors finds MageError", func(t *testing.T) {
		mageErr := NewErrorBuilder().
			WithCode(ErrNotFound).
			WithMessage("not found").
			Build()
		chain := &DefaultChainError{errors: []error{
			errDefaultChainTest1,
			mageErr,
			errDefaultChainTest2,
		}}

		assert.True(t, chain.HasError(ErrNotFound))
		assert.False(t, chain.HasError(ErrInternal))
	})
}

// TestDefaultChainErrorFindByCode tests the FindByCode method
func TestDefaultChainErrorFindByCode(t *testing.T) {
	t.Run("empty chain returns nil", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		assert.Nil(t, chain.FindByCode(ErrInternal))
	})

	t.Run("code not found returns nil", func(t *testing.T) {
		mageErr := NewErrorBuilder().
			WithCode(ErrInternal).
			WithMessage("internal").
			Build()
		chain := &DefaultChainError{errors: []error{mageErr}}

		result := chain.FindByCode(ErrNotFound)
		assert.Nil(t, result)
	})

	t.Run("code found returns error", func(t *testing.T) {
		mageErr := NewErrorBuilder().
			WithCode(ErrInternal).
			WithMessage("internal error").
			Build()
		chain := &DefaultChainError{errors: []error{mageErr}}

		result := chain.FindByCode(ErrInternal)
		require.NotNil(t, result)
		assert.Equal(t, ErrInternal, result.Code())
		assert.Contains(t, result.Error(), "internal error")
	})

	t.Run("multiple matches returns first", func(t *testing.T) {
		mageErr1 := NewErrorBuilder().
			WithCode(ErrInternal).
			WithMessage("first internal").
			Build()
		mageErr2 := NewErrorBuilder().
			WithCode(ErrInternal).
			WithMessage("second internal").
			Build()
		chain := &DefaultChainError{errors: []error{mageErr1, mageErr2}}

		result := chain.FindByCode(ErrInternal)
		require.NotNil(t, result)
		assert.Contains(t, result.Error(), "first internal")
	})

	t.Run("skips non-MageError", func(t *testing.T) {
		mageErr := NewErrorBuilder().
			WithCode(ErrInternal).
			WithMessage("mage error").
			Build()
		chain := &DefaultChainError{errors: []error{
			errDefaultChainTest1,
			mageErr,
		}}

		result := chain.FindByCode(ErrInternal)
		require.NotNil(t, result)
		assert.Contains(t, result.Error(), "mage error")
	})
}

// TestDefaultChainErrorForEach tests the ForEach method
func TestDefaultChainErrorForEach(t *testing.T) {
	t.Run("empty chain does nothing", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		callCount := 0

		err := chain.ForEach(func(err error) error {
			callCount++
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, 0, callCount)
	})

	t.Run("iterates all errors when no error returned", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{
			errDefaultChainTest1,
			errDefaultChainTest2,
			errDefaultChainTest3,
		}}
		visited := make([]error, 0)

		err := chain.ForEach(func(e error) error {
			visited = append(visited, e)
			return nil
		})

		require.NoError(t, err)
		require.Len(t, visited, 3)
		assert.Equal(t, errDefaultChainTest1, visited[0])
		assert.Equal(t, errDefaultChainTest2, visited[1])
		assert.Equal(t, errDefaultChainTest3, visited[2])
	})

	t.Run("stops on first error from function", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{
			errDefaultChainTest1,
			errDefaultChainTest2,
			errDefaultChainTest3,
		}}
		callCount := 0

		err := chain.ForEach(func(e error) error {
			callCount++
			if callCount == 2 {
				return errDefaultForEachStop
			}
			return nil
		})

		require.ErrorIs(t, err, errDefaultForEachStop)
		assert.Equal(t, 2, callCount) // Should stop after second
	})

	t.Run("single error", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{errDefaultForEachTest}}
		var capturedErr error

		err := chain.ForEach(func(e error) error {
			capturedErr = e
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, errDefaultForEachTest, capturedErr)
	})
}

// TestDefaultChainErrorFilter tests the Filter method
func TestDefaultChainErrorFilter(t *testing.T) {
	t.Run("empty chain returns empty slice", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		result := chain.Filter(func(err error) bool { return true })

		assert.Empty(t, result)
	})

	t.Run("matches all", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{
			errDefaultChainTest1,
			errDefaultChainTest2,
		}}
		result := chain.Filter(func(err error) bool { return true })

		require.Len(t, result, 2)
		assert.Equal(t, errDefaultChainTest1, result[0])
		assert.Equal(t, errDefaultChainTest2, result[1])
	})

	t.Run("matches none", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{
			errDefaultChainTest1,
			errDefaultChainTest2,
		}}
		result := chain.Filter(func(err error) bool { return false })

		assert.Empty(t, result)
	})

	t.Run("matches some", func(t *testing.T) {
		mageErr := NewErrorBuilder().
			WithCode(ErrInternal).
			WithMessage("mage error").
			Build()
		chain := &DefaultChainError{errors: []error{
			errDefaultFilterTest,
			mageErr,
			errDefaultChainTest1,
		}}

		// Filter only MageErrors
		result := chain.Filter(func(err error) bool {
			var me MageError
			return errors.As(err, &me)
		})

		require.Len(t, result, 1)
		var me MageError
		require.ErrorAs(t, result[0], &me)
		assert.Equal(t, ErrInternal, me.Code())
	})

	t.Run("filter by error message", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{
			errDefaultChainTest1,
			errDefaultChainTest2,
			errDefaultChainTest3,
		}}

		result := chain.Filter(func(err error) bool {
			return err.Error() == "default chain error 2"
		})

		require.Len(t, result, 1)
		assert.Equal(t, errDefaultChainTest2, result[0])
	})
}

// TestDefaultChainErrorToSlice tests the ToSlice method
func TestDefaultChainErrorToSlice(t *testing.T) {
	t.Run("empty chain", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{}}
		result := chain.ToSlice()

		assert.Empty(t, result)
	})

	t.Run("delegates to Errors", func(t *testing.T) {
		chain := &DefaultChainError{errors: []error{
			errDefaultChainTest1,
			errDefaultChainTest2,
		}}

		toSlice := chain.ToSlice()
		errs := chain.Errors()

		assert.Equal(t, errs, toSlice)
	})
}

// TestDefaultChainErrorBehavior verifies that DefaultChainError operations
// create new RealDefaultChainError instances
func TestDefaultChainErrorBehavior(t *testing.T) {
	t.Run("Add creates new RealDefaultChainError", func(t *testing.T) {
		original := &DefaultChainError{errors: []error{errDefaultChainTest1}}
		modified := original.Add(errDefaultChainTest2)

		// Original DefaultChainError is unchanged
		assert.Equal(t, 1, original.Count())

		// Modified is a new RealDefaultChainError with both errors
		assert.Equal(t, 2, modified.Count())
	})

	t.Run("AddWithContext creates new RealDefaultChainError", func(t *testing.T) {
		original := &DefaultChainError{errors: []error{errDefaultChainTest1}}
		ctx := &ErrorContext{Operation: "test"}
		modified := original.AddWithContext(errDefaultChainTest2, ctx)

		// Original DefaultChainError is unchanged
		assert.Equal(t, 1, original.Count())

		// Modified is a new RealDefaultChainError with both errors
		assert.Equal(t, 2, modified.Count())
	})

	t.Run("chained operations behavior", func(t *testing.T) {
		// DefaultChainError.Add() returns RealDefaultChainError
		// Subsequent Add() calls on RealDefaultChainError mutate in place
		chain0 := &DefaultChainError{errors: []error{}}
		chain1 := chain0.Add(errDefaultChainTest1) // Returns RealDefaultChainError
		chain2 := chain1.Add(errDefaultChainTest2) // Mutates chain1 (same instance)
		chain3 := chain2.Add(errDefaultChainTest3) // Mutates chain1/2 (same instance)

		// Original DefaultChainError is unchanged
		assert.Equal(t, 0, chain0.Count())

		// chain1, chain2, chain3 are all the same RealDefaultChainError
		assert.Equal(t, 3, chain1.Count())
		assert.Equal(t, 3, chain2.Count())
		assert.Equal(t, 3, chain3.Count())
	})

	t.Run("multiple independent chains from DefaultChainError", func(t *testing.T) {
		original := &DefaultChainError{errors: []error{errDefaultChainTest1}}

		// Each call to original.Add creates a new RealDefaultChainError
		branch1 := original.Add(errDefaultChainTest2)
		branch2 := original.Add(errDefaultChainTest3)

		// Original is unchanged
		assert.Equal(t, 1, original.Count())

		// Each branch has original + their own error
		assert.Equal(t, 2, branch1.Count())
		assert.Equal(t, 2, branch2.Count())

		// Branches are independent
		assert.Equal(t, errDefaultChainTest2.Error(), branch1.Last().Error())
		assert.Equal(t, errDefaultChainTest3.Error(), branch2.Last().Error())
	})
}

// TestDefaultChainErrorImplementsInterface ensures DefaultChainError
// properly implements the ErrorChain interface
//
//nolint:errcheck // return values intentionally ignored in interface verification test
func TestDefaultChainErrorImplementsInterface(t *testing.T) {
	// This test verifies compile-time interface satisfaction
	var _ ErrorChain = (*DefaultChainError)(nil)

	// Runtime check
	chain := &DefaultChainError{errors: []error{}}
	var ec ErrorChain = chain

	// All methods should be callable through interface
	ec = ec.Add(errDefaultChainTest1)
	ec = ec.AddWithContext(errDefaultChainTest2, &ErrorContext{})
	_ = ec.Error()
	_ = ec.Errors()
	_ = ec.First()
	_ = ec.Last()
	_ = ec.Count()
	_ = ec.HasError(ErrInternal)
	_ = ec.FindByCode(ErrInternal)
	_ = ec.ForEach(func(err error) error { return nil })
	_ = ec.Filter(func(err error) bool { return true })
	_ = ec.ToSlice()
}
