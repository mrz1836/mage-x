package errors

import (
	"errors"
	"regexp"
	"sync"
)

// RealDefaultErrorMatcher is the actual implementation of ErrorMatcher
type RealDefaultErrorMatcher struct {
	mu       sync.RWMutex
	matchers []func(error) bool
	inverted bool
}

// NewErrorMatcher creates a new error matcher
func NewErrorMatcher() *RealDefaultErrorMatcher {
	return &RealDefaultErrorMatcher{
		matchers: make([]func(error) bool, 0),
		inverted: false,
	}
}

// Match returns true if the error matches all criteria
func (m *RealDefaultErrorMatcher) Match(err error) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if err == nil {
		return m.inverted
	}

	// If no matchers, match any error
	if len(m.matchers) == 0 {
		return !m.inverted
	}

	// All matchers must match
	for _, matcher := range m.matchers {
		if !matcher(err) {
			return m.inverted
		}
	}

	return !m.inverted
}

// MatchCode adds a code matcher
func (m *RealDefaultErrorMatcher) MatchCode(code ErrorCode) ErrorMatcher {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.matchers = append(m.matchers, func(err error) bool {
		var mageErr MageError
		if errors.As(err, &mageErr) {
			return mageErr.Code() == code
		}
		return false
	})

	return m
}

// MatchSeverity adds a severity matcher
func (m *RealDefaultErrorMatcher) MatchSeverity(severity Severity) ErrorMatcher {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.matchers = append(m.matchers, func(err error) bool {
		var mageErr MageError
		if errors.As(err, &mageErr) {
			return mageErr.Severity() == severity
		}
		return false
	})

	return m
}

// MatchMessage adds a message pattern matcher
func (m *RealDefaultErrorMatcher) MatchMessage(pattern string) ErrorMatcher {
	m.mu.Lock()
	defer m.mu.Unlock()

	re, err := regexp.Compile(pattern)
	if err != nil {
		// If pattern is invalid, match exact string
		m.matchers = append(m.matchers, func(err error) bool {
			return err.Error() == pattern
		})
	} else {
		m.matchers = append(m.matchers, func(err error) bool {
			return re.MatchString(err.Error())
		})
	}

	return m
}

// MatchType adds a type matcher
func (m *RealDefaultErrorMatcher) MatchType(target error) ErrorMatcher {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.matchers = append(m.matchers, func(err error) bool {
		return Is(err, target)
	})

	return m
}

// MatchField adds a field matcher
func (m *RealDefaultErrorMatcher) MatchField(key string, value interface{}) ErrorMatcher {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.matchers = append(m.matchers, func(err error) bool {
		var mageErr MageError
		if errors.As(err, &mageErr) {
			if fieldValue, exists := mageErr.Context().Fields[key]; exists {
				return fieldValue == value
			}
		}
		return false
	})

	return m
}

// MatchAny creates a matcher that matches if ANY of the provided matchers match
func (m *RealDefaultErrorMatcher) MatchAny(matchers ...ErrorMatcher) ErrorMatcher {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.matchers = append(m.matchers, func(err error) bool {
		for _, matcher := range matchers {
			if matcher.Match(err) {
				return true
			}
		}
		return false
	})

	return m
}

// MatchAll creates a matcher that matches if ALL of the provided matchers match
func (m *RealDefaultErrorMatcher) MatchAll(matchers ...ErrorMatcher) ErrorMatcher {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.matchers = append(m.matchers, func(err error) bool {
		for _, matcher := range matchers {
			if !matcher.Match(err) {
				return false
			}
		}
		return true
	})

	return m
}

// Not inverts the matcher logic
func (m *RealDefaultErrorMatcher) Not() ErrorMatcher {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a new matcher with inverted logic
	newMatcher := &RealDefaultErrorMatcher{
		matchers: make([]func(error) bool, len(m.matchers)),
		inverted: !m.inverted,
	}
	copy(newMatcher.matchers, m.matchers)

	return newMatcher
}

// Update DefaultErrorMatcher methods to use the real implementation
func (m *DefaultErrorMatcher) Match(err error) bool {
	matcher := NewErrorMatcher()
	for _, fn := range m.matchers {
		matcher.matchers = append(matcher.matchers, fn)
	}
	return matcher.Match(err)
}

func (m *DefaultErrorMatcher) MatchCode(code ErrorCode) ErrorMatcher {
	return NewErrorMatcher().MatchCode(code)
}

func (m *DefaultErrorMatcher) MatchSeverity(severity Severity) ErrorMatcher {
	return NewErrorMatcher().MatchSeverity(severity)
}

func (m *DefaultErrorMatcher) MatchMessage(pattern string) ErrorMatcher {
	return NewErrorMatcher().MatchMessage(pattern)
}

func (m *DefaultErrorMatcher) MatchType(target error) ErrorMatcher {
	return NewErrorMatcher().MatchType(target)
}

func (m *DefaultErrorMatcher) MatchField(key string, value interface{}) ErrorMatcher {
	return NewErrorMatcher().MatchField(key, value)
}

func (m *DefaultErrorMatcher) MatchAny(matchers ...ErrorMatcher) ErrorMatcher {
	return NewErrorMatcher().MatchAny(matchers...)
}

func (m *DefaultErrorMatcher) MatchAll(matchers ...ErrorMatcher) ErrorMatcher {
	return NewErrorMatcher().MatchAll(matchers...)
}

func (m *DefaultErrorMatcher) Not() ErrorMatcher {
	return NewErrorMatcher().Not()
}

// Convenience functions for creating matchers

// CodeMatcher creates a matcher for error codes
func CodeMatcher(codes ...ErrorCode) ErrorMatcher {
	matcher := NewMatcher()
	if len(codes) == 1 {
		return matcher.MatchCode(codes[0])
	}

	// Multiple codes - match any
	subMatchers := make([]ErrorMatcher, len(codes))
	for i, code := range codes {
		subMatchers[i] = NewMatcher().MatchCode(code)
	}
	return matcher.MatchAny(subMatchers...)
}

// SeverityMatcher creates a matcher for severities
func SeverityMatcher(severities ...Severity) ErrorMatcher {
	matcher := NewMatcher()
	if len(severities) == 1 {
		return matcher.MatchSeverity(severities[0])
	}

	// Multiple severities - match any
	subMatchers := make([]ErrorMatcher, len(severities))
	for i, severity := range severities {
		subMatchers[i] = NewMatcher().MatchSeverity(severity)
	}
	return matcher.MatchAny(subMatchers...)
}

// MessageMatcher creates a matcher for error messages
func MessageMatcher(pattern string) ErrorMatcher {
	return NewMatcher().MatchMessage(pattern)
}

// TypeMatcher creates a matcher for error types
func TypeMatcher(target error) ErrorMatcher {
	return NewMatcher().MatchType(target)
}

// FieldMatcher creates a matcher for error fields
func FieldMatcher(key string, value interface{}) ErrorMatcher {
	return NewMatcher().MatchField(key, value)
}

// CriticalMatcher matches critical and fatal errors
func CriticalMatcher() ErrorMatcher {
	return SeverityMatcher(SeverityCritical, SeverityFatal)
}

// RetryableMatcher matches retryable errors
func RetryableMatcher() ErrorMatcher {
	return NewMatcher().MatchAny(
		CodeMatcher(ErrInternal, ErrTimeout, ErrResourceExhausted, ErrUnavailable),
		CodeMatcher(ErrBuildFailed, ErrTestFailed, ErrCommandTimeout),
	)
}

// BuildErrorMatcher matches build-related errors
func BuildErrorMatcher() ErrorMatcher {
	return CodeMatcher(
		ErrBuildFailed,
		ErrCompileFailed,
		ErrTestFailed,
		ErrLintFailed,
		ErrPackageFailed,
		ErrDependencyError,
	)
}
