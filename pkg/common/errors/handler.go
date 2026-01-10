package errors

import (
	"context"
	"errors"
	"sync"
)

// RealDefaultErrorHandler is the actual implementation of ErrorHandler
type RealDefaultErrorHandler struct {
	mu               sync.RWMutex
	codeHandlers     map[ErrorCode]func(MageError) error
	severityHandlers map[Severity]func(MageError) error
	defaultHandler   func(error) error
	fallbackHandler  func(error) error
}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *RealDefaultErrorHandler {
	return &RealDefaultErrorHandler{
		codeHandlers:     make(map[ErrorCode]func(MageError) error),
		severityHandlers: make(map[Severity]func(MageError) error),
	}
}

// Handle handles an error based on registered handlers
func (h *RealDefaultErrorHandler) Handle(err error) error {
	if err == nil {
		return nil
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// Try to handle as MageError
	var mageErr MageError
	if !errors.As(err, &mageErr) {
		// Not a MageError, use default handler if available
		if h.defaultHandler != nil {
			if handlerErr := h.defaultHandler(err); handlerErr != nil {
				return h.handleFallback(handlerErr)
			}
			return nil
		}
		// No default handler, use fallback
		return h.handleFallback(err)
	}

	// Check code-specific handler first
	if handler, exists := h.codeHandlers[mageErr.Code()]; exists {
		if handlerErr := handler(mageErr); handlerErr != nil {
			return h.handleFallback(handlerErr)
		}
		return nil
	}

	// Check severity-specific handler
	if handler, exists := h.severityHandlers[mageErr.Severity()]; exists {
		if handlerErr := handler(mageErr); handlerErr != nil {
			return h.handleFallback(handlerErr)
		}
		return nil
	}

	// Use default handler
	if h.defaultHandler != nil {
		if handlerErr := h.defaultHandler(err); handlerErr != nil {
			return h.handleFallback(handlerErr)
		}
		return nil
	}

	// No handler found, use fallback
	return h.handleFallback(err)
}

// HandleWithContext handles an error with context
func (h *RealDefaultErrorHandler) HandleWithContext(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	// Check for context cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Enhance error with context information if possible
	var mageErr MageError
	if errors.As(err, &mageErr) {
		// Extract context values if available
		if requestID, ok := ctx.Value("requestID").(string); ok {
			mageErr = mageErr.WithField("requestID", requestID)
		}
		if userID, ok := ctx.Value("userID").(string); ok {
			mageErr = mageErr.WithField("userID", userID)
		}

		return h.Handle(mageErr)
	}

	return h.Handle(err)
}

// OnError registers a handler for a specific error code
func (h *RealDefaultErrorHandler) OnError(code ErrorCode, handler func(MageError) error) ErrorHandler {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.codeHandlers[code] = handler
	return h
}

// OnSeverity registers a handler for a specific severity
func (h *RealDefaultErrorHandler) OnSeverity(severity Severity, handler func(MageError) error) ErrorHandler {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.severityHandlers[severity] = handler
	return h
}

// SetDefault sets the default error handler
func (h *RealDefaultErrorHandler) SetDefault(handler func(error) error) ErrorHandler {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.defaultHandler = handler
	return h
}

// SetFallback sets the fallback error handler
func (h *RealDefaultErrorHandler) SetFallback(handler func(error) error) ErrorHandler {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.fallbackHandler = handler
	return h
}

// handleFallback handles errors using the fallback handler
func (h *RealDefaultErrorHandler) handleFallback(err error) error {
	if h.fallbackHandler != nil {
		return h.fallbackHandler(err)
	}
	return err
}

// NewDefaultErrorHandler creates a new DefaultErrorHandler
func NewDefaultErrorHandler() *DefaultErrorHandler {
	return &DefaultErrorHandler{
		handlers:         make(map[ErrorCode]func(MageError) error),
		severityHandlers: make(map[Severity]func(MageError) error),
	}
}

// Handle handles an error using the DefaultErrorHandler
func (h *DefaultErrorHandler) Handle(err error) error {
	handler := NewErrorHandler()
	handler.codeHandlers = h.handlers
	handler.severityHandlers = h.severityHandlers
	return handler.Handle(err)
}

// HandleWithContext handles an error with context using the DefaultErrorHandler
func (h *DefaultErrorHandler) HandleWithContext(ctx context.Context, err error) error {
	handler := NewErrorHandler()
	handler.codeHandlers = h.handlers
	handler.severityHandlers = h.severityHandlers
	return handler.HandleWithContext(ctx, err)
}

// OnError sets a handler for a specific error code
func (h *DefaultErrorHandler) OnError(code ErrorCode, handler func(MageError) error) ErrorHandler {
	h.handlers[code] = handler
	return h
}

// OnSeverity sets a handler for a specific error severity
func (h *DefaultErrorHandler) OnSeverity(severity Severity, handler func(MageError) error) ErrorHandler {
	h.severityHandlers[severity] = handler
	return h
}

// SetDefault sets the default error handler
func (h *DefaultErrorHandler) SetDefault(handler func(error) error) ErrorHandler {
	realHandler := NewErrorHandler()
	realHandler.SetDefault(handler)
	return h
}

// SetFallback sets the fallback error handler
func (h *DefaultErrorHandler) SetFallback(handler func(error) error) ErrorHandler {
	realHandler := NewErrorHandler()
	realHandler.SetFallback(handler)
	return h
}
