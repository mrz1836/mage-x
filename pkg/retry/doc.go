// Package retry provides a flexible retry framework with configurable backoff strategies,
// error classification, and context-aware cancellation support.
//
// The package is designed to make error recovery elegant and efficient for network
// operations, command execution, and other transient failure scenarios.
//
// # Basic Usage
//
// The simplest way to use the retry package:
//
//	err := retry.Do(ctx, nil, func() error {
//	    return someNetworkCall()
//	})
//
// # Configuration
//
// Retry behavior is controlled through the Config struct:
//
//	cfg := &retry.Config{
//	    MaxAttempts: 5,
//	    Classifier:  retry.NewNetworkClassifier(),
//	    Backoff:     retry.DefaultBackoff(),
//	    OnRetry: func(attempt int, err error, delay time.Duration) {
//	        log.Printf("Retry %d: %v (waiting %v)", attempt, err, delay)
//	    },
//	}
//	err := retry.Do(ctx, cfg, operation)
//
// # Error Classification
//
// The Classifier interface determines which errors are retriable:
//
//   - NetworkClassifier: Retries common network errors (timeouts, connection refused)
//   - CommandClassifier: Retries transient command execution failures
//   - Custom classifiers can be implemented for domain-specific logic
//
// # Backoff Strategies
//
// Built-in backoff implementations:
//
//   - ExponentialBackoff: Increases delay exponentially with optional jitter
//   - FixedBackoff: Uses a constant delay between retries
//   - DefaultBackoff(): Returns a sensible exponential backoff configuration
//   - FastBackoff(): Returns a faster backoff for quick operations
//
// # Memory Efficiency
//
// The retry loop uses a reusable timer to avoid memory allocations on each
// retry attempt, making it suitable for high-frequency retry scenarios.
//
// # Context Support
//
// All retry operations support context cancellation. If the context is canceled
// during a retry wait, the operation returns immediately with the context error.
package retry
