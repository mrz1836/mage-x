package errors

import (
	"errors"
	"fmt"
	"strings"
	"sync"
)

// RealDefaultChainError is the actual implementation of ErrorChain
type RealDefaultChainError struct {
	mu     sync.RWMutex
	errors []error
}

// NewErrorChain creates a new error chain
func NewErrorChain() *RealDefaultChainError {
	return &RealDefaultChainError{
		errors: make([]error, 0),
	}
}

// Add adds an error to the chain
func (c *RealDefaultChainError) Add(err error) ErrorChain {
	if err == nil {
		return c
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.errors = append(c.errors, err)
	return c
}

// AddWithContext adds an error with context to the chain
func (c *RealDefaultChainError) AddWithContext(err error, ctx *ErrorContext) ErrorChain {
	if err == nil {
		return c
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Wrap the error with context if it's not already a MageError
	var mageErr MageError
	var me MageError
	if errors.As(err, &me) {
		mageErr = me.WithContext(ctx)
	} else {
		mageErr = NewBuilder().
			WithMessage("%s", err.Error()).
			WithContext(ctx).
			WithCause(err).
			Build()
	}

	c.errors = append(c.errors, mageErr)
	return c
}

// Error returns a string representation of all errors
func (c *RealDefaultChainError) Error() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.errors) == 0 {
		return "no errors"
	}

	if len(c.errors) == 1 {
		return c.errors[0].Error()
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%d errors occurred:\n", len(c.errors))

	for i, err := range c.errors {
		fmt.Fprintf(&sb, "  [%d] %v\n", i+1, err)
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// Errors returns all errors in the chain
func (c *RealDefaultChainError) Errors() []error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]error, len(c.errors))
	copy(result, c.errors)
	return result
}

// First returns the first error in the chain
func (c *RealDefaultChainError) First() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.errors) == 0 {
		return nil
	}
	return c.errors[0]
}

// Last returns the last error in the chain
func (c *RealDefaultChainError) Last() error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.errors) == 0 {
		return nil
	}
	return c.errors[len(c.errors)-1]
}

// Count returns the number of errors in the chain
func (c *RealDefaultChainError) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.errors)
}

// HasError returns true if the chain contains an error with the given code
func (c *RealDefaultChainError) HasError(code ErrorCode) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, err := range c.errors {
		var mageErr MageError
		if errors.As(err, &mageErr) {
			if mageErr.Code() == code {
				return true
			}
		}
	}
	return false
}

// FindByCode returns the first error with the given code
func (c *RealDefaultChainError) FindByCode(code ErrorCode) MageError {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, err := range c.errors {
		var mageErr MageError
		if errors.As(err, &mageErr) {
			if mageErr.Code() == code {
				return mageErr
			}
		}
	}
	return nil
}

// ForEach executes a function for each error in the chain
func (c *RealDefaultChainError) ForEach(fn func(error) error) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, err := range c.errors {
		if e := fn(err); e != nil {
			return e
		}
	}
	return nil
}

// Filter returns errors that match the predicate
func (c *RealDefaultChainError) Filter(predicate func(error) bool) []error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Pre-allocate with capacity to avoid reallocations
	result := make([]error, 0, len(c.errors))
	for _, err := range c.errors {
		if predicate(err) {
			result = append(result, err)
		}
	}
	return result
}

// ToSlice returns all errors as a slice
func (c *RealDefaultChainError) ToSlice() []error {
	return c.Errors()
}

// NewDefaultChainError creates a new DefaultChainError
func NewDefaultChainError() *DefaultChainError {
	return &DefaultChainError{errors: []error{}}
}

// Add adds an error to the chain (O(1) operation)
func (c *DefaultChainError) Add(err error) ErrorChain {
	if err != nil {
		c.errors = append(c.errors, err)
	}
	return c
}

// AddWithContext adds an error with context to the chain (O(1) operation)
func (c *DefaultChainError) AddWithContext(err error, ctx *ErrorContext) ErrorChain {
	if err == nil {
		return c
	}

	// Wrap the error with context if it's not already a MageError
	var mageErr MageError
	var me MageError
	if errors.As(err, &me) {
		mageErr = me.WithContext(ctx)
	} else {
		mageErr = NewBuilder().
			WithMessage("%s", err.Error()).
			WithContext(ctx).
			WithCause(err).
			Build()
	}
	c.errors = append(c.errors, mageErr)
	return c
}

// Error returns a string representation of all errors (O(n) operation)
func (c *DefaultChainError) Error() string {
	if len(c.errors) == 0 {
		return "no errors"
	}

	if len(c.errors) == 1 {
		return c.errors[0].Error()
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%d errors occurred:\n", len(c.errors))

	for i, err := range c.errors {
		fmt.Fprintf(&sb, "  [%d] %v\n", i+1, err)
	}

	return strings.TrimSuffix(sb.String(), "\n")
}

// Errors returns all errors in the chain
func (c *DefaultChainError) Errors() []error {
	return c.errors
}

// First returns the first error in the chain
func (c *DefaultChainError) First() error {
	if len(c.errors) == 0 {
		return nil
	}
	return c.errors[0]
}

// Last returns the last error in the chain
func (c *DefaultChainError) Last() error {
	if len(c.errors) == 0 {
		return nil
	}
	return c.errors[len(c.errors)-1]
}

// Count returns the number of errors in the chain
func (c *DefaultChainError) Count() int {
	return len(c.errors)
}

// HasError checks if the chain contains an error with the given code
func (c *DefaultChainError) HasError(code ErrorCode) bool {
	for _, err := range c.errors {
		var mageErr MageError
		if errors.As(err, &mageErr) && mageErr.Code() == code {
			return true
		}
	}
	return false
}

// FindByCode finds the first error in the chain with the given code
func (c *DefaultChainError) FindByCode(code ErrorCode) MageError {
	for _, err := range c.errors {
		var mageErr MageError
		if errors.As(err, &mageErr) && mageErr.Code() == code {
			return mageErr
		}
	}
	return nil
}

// ForEach executes a function for each error in the chain
func (c *DefaultChainError) ForEach(fn func(error) error) error {
	for _, err := range c.errors {
		if e := fn(err); e != nil {
			return e
		}
	}
	return nil
}

// Filter returns errors that match the given predicate
func (c *DefaultChainError) Filter(predicate func(error) bool) []error {
	// Pre-allocate with capacity to avoid reallocations
	result := make([]error, 0, len(c.errors))
	for _, err := range c.errors {
		if predicate(err) {
			result = append(result, err)
		}
	}
	return result
}

// ToSlice returns all errors in the chain as a slice.
func (c *DefaultChainError) ToSlice() []error {
	return c.Errors()
}
