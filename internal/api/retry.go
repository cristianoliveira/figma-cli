package api

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"time"
)

// RetryConfig holds configuration for retry behavior.
type RetryConfig struct {
	// MaxAttempts is the maximum number of attempts (including initial).
	MaxAttempts int
	// InitialDelay is the initial delay before first retry.
	InitialDelay time.Duration
	// MaxDelay is the maximum delay between retries.
	MaxDelay time.Duration
	// Multiplier is the factor by which delay increases each retry.
	Multiplier float64
	// JitterFactor adds random jitter to delays (0 = no jitter, 0.5 = up to 50% extra).
	JitterFactor float64
	// RetryableStatusCodes is a set of HTTP status codes that should be retried.
	RetryableStatusCodes map[int]bool
}

// DefaultRetryConfig returns a sensible default configuration for Figma API.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		JitterFactor: 0.2,
		RetryableStatusCodes: map[int]bool{
			http.StatusTooManyRequests:     true, // 429
			http.StatusInternalServerError: true, // 500
			http.StatusBadGateway:          true, // 502
			http.StatusServiceUnavailable:  true, // 503
			http.StatusGatewayTimeout:      true, // 504
		},
	}
}

// RetryableError is an error that indicates a request can be retried.
type RetryableError struct {
	Err        error
	RetryAfter time.Duration // Optional suggested wait time
}

func (e *RetryableError) Error() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("retryable error (retry after %v): %v", e.RetryAfter, e.Err)
	}
	return fmt.Sprintf("retryable error: %v", e.Err)
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// WithRetry executes a function with retry logic.
// The function should return an error; if the error is retryable, WithRetry will retry according to config.
func WithRetry(ctx context.Context, config RetryConfig, fn func(ctx context.Context) error) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Execute the function
		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		// Determine if error is retryable
		var retryAfter time.Duration
		if re, ok := err.(*RetryableError); ok {
			retryAfter = re.RetryAfter
		} else if !isRetryableError(err, config) {
			// Not retryable
			return err
		}

		// If this was the last attempt, break
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Wait before retry
		waitTime := delay
		if retryAfter > 0 {
			// Use Retry-After suggestion if present
			waitTime = retryAfter
		}

		// Add jitter
		if config.JitterFactor > 0 {
			jitter := 1 + config.JitterFactor*(2*rand.Float64()-1) // ±JitterFactor
			waitTime = time.Duration(float64(waitTime) * jitter)
			// Ensure waitTime is positive and not exceeding MaxDelay
			if waitTime < 0 {
				waitTime = 0
			}
			if waitTime > config.MaxDelay {
				waitTime = config.MaxDelay
			}
		}

		// Wait with context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		case <-time.After(waitTime):
			// Continue to next attempt
		}

		// Increase delay for next retry (unless we'll use Retry-After)
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}
	}

	return lastErr
}

// isRetryableError determines if an error is retryable based on config.
func isRetryableError(err error, config RetryConfig) bool {
	// Check if it's a RateLimitError (always retryable)
	if IsRateLimitError(err) {
		return true
	}

	// Check if it's an APIError with retryable status code
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		if config.RetryableStatusCodes[apiErr.StatusCode] {
			return true
		}
	}

	// Check for network errors (temporary, timeouts, connection failures)
	var netErr net.Error
	if errors.As(err, &netErr) {
		// Retry on temporary network errors and timeouts
		//nolint:staticcheck // Temporary is deprecated but we need to maintain behavior
		if netErr.Temporary() || netErr.Timeout() {
			return true
		}
	}

	return false
}

// NewRetryableError wraps an error as retryable.
func NewRetryableError(err error) *RetryableError {
	return &RetryableError{Err: err}
}

// NewRetryableErrorWithWait wraps an error as retryable with suggested wait time.
func NewRetryableErrorWithWait(err error, retryAfter time.Duration) *RetryableError {
	return &RetryableError{Err: err, RetryAfter: retryAfter}
}
