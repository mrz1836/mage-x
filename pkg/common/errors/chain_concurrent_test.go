package errors

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Static test errors for concurrency tests (err113 compliant)
var (
	errConcurrentBase     = errors.New("concurrent base error")
	errConcurrentWorker   = errors.New("concurrent worker error")
	errConcurrentAdd      = errors.New("concurrent add error")
	errConcurrentFilter   = errors.New("filter target error")
	errConcurrentForEach  = errors.New("foreach error")
	errConcurrentStopIter = errors.New("stop iteration")
	errConcurrentFirst    = errors.New("first error")
)

// TestRealDefaultChainErrorConcurrentAdd tests that concurrent Add operations
// are thread-safe and don't cause data races
//
//nolint:errcheck // return values intentionally ignored in concurrency tests
func TestRealDefaultChainErrorConcurrentAdd(t *testing.T) {
	t.Run("concurrent adds from multiple goroutines", func(t *testing.T) {
		chain := NewErrorChain()
		const numGoroutines = 100
		const errorsPerGoroutine = 10

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < errorsPerGoroutine; j++ {
					_ = chain.Add(errConcurrentWorker)
				}
			}()
		}

		wg.Wait()

		// All errors should be added
		expectedCount := numGoroutines * errorsPerGoroutine
		assert.Equal(t, expectedCount, chain.Count())
	})

	t.Run("concurrent adds with nil errors", func(t *testing.T) {
		chain := NewErrorChain()
		const numGoroutines = 50

		var wg sync.WaitGroup
		wg.Add(numGoroutines * 2)

		// Half add real errors, half add nil
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				_ = chain.Add(errConcurrentAdd)
			}()
			go func() {
				defer wg.Done()
				_ = chain.Add(nil) // Should be ignored
			}()
		}

		wg.Wait()

		// Only real errors should be counted
		assert.Equal(t, numGoroutines, chain.Count())
	})

	t.Run("concurrent AddWithContext", func(t *testing.T) {
		chain := NewErrorChain()
		const numGoroutines = 50

		var wg sync.WaitGroup
		wg.Add(numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				ctx := &ErrorContext{
					Operation: "ConcurrentOp",
					User:      "worker",
				}
				_ = chain.AddWithContext(errConcurrentBase, ctx)
			}()
		}

		wg.Wait()

		assert.Equal(t, numGoroutines, chain.Count())

		// Verify all errors have context
		for _, err := range chain.Errors() {
			var mageErr MageError
			require.ErrorAs(t, err, &mageErr)
			assert.Equal(t, "ConcurrentOp", mageErr.Context().Operation)
		}
	})
}

// TestRealDefaultChainErrorConcurrentReadWrite tests that concurrent reads
// and writes don't cause data races
//
//nolint:errcheck // return values intentionally ignored in concurrency tests
func TestRealDefaultChainErrorConcurrentReadWrite(t *testing.T) {
	t.Run("concurrent reads during writes", func(t *testing.T) {
		chain := NewErrorChain()
		const numWriters = 20
		const numReaders = 20
		const operationsPerWorker = 50

		var wg sync.WaitGroup
		wg.Add(numWriters + numReaders)

		// Start writers
		for i := 0; i < numWriters; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerWorker; j++ {
					_ = chain.Add(errConcurrentWorker)
				}
			}()
		}

		// Start readers
		for i := 0; i < numReaders; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerWorker; j++ {
					_ = chain.Count()
					_ = chain.Errors()
					_ = chain.First()
					_ = chain.Last()
					_ = chain.Error()
				}
			}()
		}

		wg.Wait()

		// Verify final state
		expectedCount := numWriters * operationsPerWorker
		assert.Equal(t, expectedCount, chain.Count())
	})

	t.Run("concurrent HasError and FindByCode during writes", func(t *testing.T) {
		chain := NewErrorChain()
		const numWriters = 10
		const numReaders = 10
		const operationsPerWorker = 30

		var wg sync.WaitGroup
		wg.Add(numWriters + numReaders)

		// Start writers adding MageErrors
		for i := 0; i < numWriters; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerWorker; j++ {
					mageErr := NewErrorBuilder().
						WithCode(ErrInternal).
						WithMessage("concurrent mage error").
						Build()
					_ = chain.Add(mageErr)
				}
			}()
		}

		// Start readers checking for errors
		for i := 0; i < numReaders; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerWorker; j++ {
					_ = chain.HasError(ErrInternal)
					_ = chain.HasError(ErrNotFound)
					_ = chain.FindByCode(ErrInternal)
					_ = chain.FindByCode(ErrNotFound)
				}
			}()
		}

		wg.Wait()

		// Verify final state
		expectedCount := numWriters * operationsPerWorker
		assert.Equal(t, expectedCount, chain.Count())
		assert.True(t, chain.HasError(ErrInternal))
		assert.False(t, chain.HasError(ErrNotFound))
	})
}

// TestRealDefaultChainErrorConcurrentFilter tests that Filter is thread-safe
//
//nolint:errcheck // return values intentionally ignored in concurrency tests
func TestRealDefaultChainErrorConcurrentFilter(t *testing.T) {
	t.Run("concurrent filter during writes", func(t *testing.T) {
		chain := NewErrorChain()

		// Pre-populate with some errors
		for i := 0; i < 50; i++ {
			if i%2 == 0 {
				_ = chain.Add(errConcurrentFilter)
			} else {
				_ = chain.Add(errConcurrentBase)
			}
		}

		const numWriters = 10
		const numFilterers = 10
		const operationsPerWorker = 20

		var wg sync.WaitGroup
		wg.Add(numWriters + numFilterers)

		// Start writers
		for i := 0; i < numWriters; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerWorker; j++ {
					_ = chain.Add(errConcurrentWorker)
				}
			}()
		}

		// Start filterers
		filterResults := make([][]error, numFilterers)
		for i := 0; i < numFilterers; i++ {
			idx := i
			go func() {
				defer wg.Done()
				for j := 0; j < operationsPerWorker; j++ {
					result := chain.Filter(func(err error) bool {
						return err.Error() == errConcurrentFilter.Error()
					})
					// Store last result
					filterResults[idx] = result
				}
			}()
		}

		wg.Wait()

		// Verify filter results are consistent (all non-nil, contain expected errors)
		for _, result := range filterResults {
			require.NotNil(t, result)
			// All filtered errors should be errConcurrentFilter
			for _, err := range result {
				assert.Equal(t, errConcurrentFilter.Error(), err.Error())
			}
		}
	})

	t.Run("filter with MageError predicate", func(t *testing.T) {
		chain := NewErrorChain()

		// Add mixed errors
		for i := 0; i < 20; i++ {
			if i%2 == 0 {
				mageErr := NewErrorBuilder().
					WithCode(ErrInternal).
					WithMessage("internal").
					Build()
				_ = chain.Add(mageErr)
			} else {
				_ = chain.Add(errConcurrentBase)
			}
		}

		const numFilterers = 20
		var wg sync.WaitGroup
		wg.Add(numFilterers)

		for i := 0; i < numFilterers; i++ {
			go func() {
				defer wg.Done()
				result := chain.Filter(func(err error) bool {
					var me MageError
					return errors.As(err, &me)
				})
				// Should have 10 MageErrors
				assert.Len(t, result, 10)
			}()
		}

		wg.Wait()
	})
}

// TestRealDefaultChainErrorConcurrentForEach tests that ForEach is thread-safe
//
//nolint:errcheck // return values intentionally ignored in concurrency tests
func TestRealDefaultChainErrorConcurrentForEach(t *testing.T) {
	t.Run("concurrent ForEach during writes", func(t *testing.T) {
		chain := NewErrorChain()

		// Pre-populate
		for i := 0; i < 20; i++ {
			_ = chain.Add(errConcurrentForEach)
		}

		const numWriters = 5
		const numIterators = 10

		var wg sync.WaitGroup
		wg.Add(numWriters + numIterators)

		// Start writers
		for i := 0; i < numWriters; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < 10; j++ {
					_ = chain.Add(errConcurrentWorker)
				}
			}()
		}

		// Start iterators
		iterationCounts := make([]int, numIterators)
		for i := 0; i < numIterators; i++ {
			idx := i
			go func() {
				defer wg.Done()
				count := 0
				_ = chain.ForEach(func(err error) error {
					count++
					return nil
				})
				iterationCounts[idx] = count
			}()
		}

		wg.Wait()

		// Verify all iterators completed without panic
		for _, count := range iterationCounts {
			assert.Positive(t, count)
		}
	})

	t.Run("concurrent ForEach with early termination", func(t *testing.T) {
		chain := NewErrorChain()

		// Pre-populate
		for i := 0; i < 100; i++ {
			_ = chain.Add(errConcurrentForEach)
		}

		const numIterators = 20
		var wg sync.WaitGroup
		wg.Add(numIterators)

		for i := 0; i < numIterators; i++ {
			go func() {
				defer wg.Done()
				count := 0
				err := chain.ForEach(func(e error) error {
					count++
					if count >= 10 {
						return errConcurrentStopIter
					}
					return nil
				})
				// Should terminate early (use assert in goroutine, not require)
				assert.ErrorIs(t, err, errConcurrentStopIter)
				assert.Equal(t, 10, count)
			}()
		}

		wg.Wait()
	})
}

// TestRealDefaultChainErrorConcurrentToSliceErrors tests ToSlice and Errors
// are thread-safe
//
//nolint:errcheck // return values intentionally ignored in concurrency tests
func TestRealDefaultChainErrorConcurrentToSliceErrors(t *testing.T) {
	chain := NewErrorChain()

	// Pre-populate
	for i := 0; i < 50; i++ {
		_ = chain.Add(errConcurrentBase)
	}

	const numReaders = 50
	var wg sync.WaitGroup
	wg.Add(numReaders * 2)

	// Half use ToSlice, half use Errors
	for i := 0; i < numReaders; i++ {
		go func() {
			defer wg.Done()
			result := chain.ToSlice()
			assert.Len(t, result, 50)
		}()
		go func() {
			defer wg.Done()
			result := chain.Errors()
			assert.Len(t, result, 50)
		}()
	}

	wg.Wait()
}

// TestRealDefaultChainErrorConcurrentFirstLast tests First and Last
// are thread-safe during concurrent operations
//
//nolint:errcheck // return values intentionally ignored in concurrency tests
func TestRealDefaultChainErrorConcurrentFirstLast(t *testing.T) {
	chain := NewErrorChain()
	_ = chain.Add(errConcurrentFirst)

	const numOperations = 100
	var wg sync.WaitGroup
	wg.Add(numOperations * 3)

	// Concurrent reads of First
	for i := 0; i < numOperations; i++ {
		go func() {
			defer wg.Done()
			first := chain.First()
			assert.Equal(t, errConcurrentFirst.Error(), first.Error())
		}()
	}

	// Concurrent reads of Last (while also adding)
	for i := 0; i < numOperations; i++ {
		go func() {
			defer wg.Done()
			_ = chain.Last() // Just verify no panic
		}()
	}

	// Concurrent adds
	for i := 0; i < numOperations; i++ {
		go func() {
			defer wg.Done()
			_ = chain.Add(errConcurrentWorker)
		}()
	}

	wg.Wait()

	// First should still be the first error
	assert.Equal(t, errConcurrentFirst.Error(), chain.First().Error())
}

// TestRealDefaultChainErrorConcurrentCount tests Count is thread-safe
//
//nolint:errcheck // return values intentionally ignored in concurrency tests
func TestRealDefaultChainErrorConcurrentCount(t *testing.T) {
	chain := NewErrorChain()

	const numAdders = 10
	const addsPerWorker = 100
	const numCounters = 20

	var wg sync.WaitGroup
	wg.Add(numAdders + numCounters)

	// Start adders
	for i := 0; i < numAdders; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < addsPerWorker; j++ {
				_ = chain.Add(errConcurrentWorker)
			}
		}()
	}

	// Start counters
	for i := 0; i < numCounters; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				count := chain.Count()
				// Count should be monotonically increasing (or at least non-negative)
				assert.GreaterOrEqual(t, count, 0)
			}
		}()
	}

	wg.Wait()

	// Final count should be exact
	assert.Equal(t, numAdders*addsPerWorker, chain.Count())
}
