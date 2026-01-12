package retry_test

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/mrz1836/mage-x/pkg/retry"
)

func ExampleDo() {
	ctx := context.Background()

	// Counter to simulate transient failures
	attempts := 0

	// Create a classifier that considers our errors retryable
	cfg := &retry.Config{
		MaxAttempts: 3,
		Classifier:  retry.ClassifierFunc(func(err error) bool { return true }),
		Backoff:     retry.NoDelay(),
	}

	// Retry a function that fails a few times then succeeds
	err := retry.Do(ctx, cfg, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("temporary failure") //nolint:err113 // Example code
		}
		return nil
	})

	if err == nil {
		fmt.Printf("Succeeded after %d attempts\n", attempts)
	}
	// Output: Succeeded after 3 attempts
}

func ExampleDoWithData() {
	ctx := context.Background()

	// Retry a function that returns data
	result, err := retry.DoWithData(ctx, retry.DefaultConfig(), func() (string, error) {
		return "success", nil
	})

	if err == nil {
		fmt.Println(result)
	}
	// Output: success
}

func ExampleConfig_customBackoff() {
	ctx := context.Background()

	// Create custom configuration with exponential backoff
	cfg := &retry.Config{
		MaxAttempts: 5,
		Classifier:  retry.ClassifierFunc(func(err error) bool { return true }),
		Backoff: &retry.ExponentialBackoff{
			Initial:    100 * time.Millisecond,
			Multiplier: 2.0,
			Max:        2 * time.Second,
		},
		OnRetry: func(attempt int, err error, delay time.Duration) {
			fmt.Printf("Attempt %d failed, retrying in %v\n", attempt+1, delay)
		},
	}

	attempts := 0
	//nolint:errcheck,gosec // Example code - error handling is not the focus
	retry.Do(ctx, cfg, func() error {
		attempts++
		if attempts < 2 {
			return errors.New("failed") //nolint:err113 // Example code
		}
		return nil
	})

	fmt.Printf("Total attempts: %d\n", attempts)
	// Output:
	// Attempt 1 failed, retrying in 100ms
	// Total attempts: 2
}

func ExampleDefaultConfig() {
	cfg := retry.DefaultConfig()

	fmt.Printf("Max attempts: %d\n", cfg.MaxAttempts)
	// Output: Max attempts: 3
}

func ExampleCommandConfig() {
	cfg := retry.CommandConfig()

	fmt.Printf("Max attempts: %d\n", cfg.MaxAttempts)
	fmt.Println("Uses command classifier for retry decisions")
	// Output:
	// Max attempts: 3
	// Uses command classifier for retry decisions
}

func ExampleNewNetworkClassifier() {
	classifier := retry.NewNetworkClassifier()

	// Network errors are retryable
	networkErr := errors.New("connection refused") //nolint:err113 // Example code
	fmt.Printf("Network error retryable: %v\n", classifier.IsRetriable(networkErr))

	// Context canceled is not retryable
	cancelErr := context.Canceled
	fmt.Printf("Cancel error retryable: %v\n", classifier.IsRetriable(cancelErr))
	// Output:
	// Network error retryable: true
	// Cancel error retryable: false
}
