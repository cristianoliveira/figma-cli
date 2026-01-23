package api

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

// mockNetError is a mock net.Error for testing
type mockNetError struct {
	timeout   bool
	temporary bool
	message   string
}

func (e *mockNetError) Error() string   { return e.message }
func (e *mockNetError) Timeout() bool   { return e.timeout }
func (e *mockNetError) Temporary() bool { return e.temporary }

func TestIsRetryableError_NetworkErrors(t *testing.T) {
	config := DefaultRetryConfig()

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "timeout error",
			err:      &mockNetError{timeout: true, message: "timeout"},
			expected: true,
		},
		{
			name:     "temporary error",
			err:      &mockNetError{temporary: true, message: "temporary"},
			expected: true,
		},
		{
			name:     "both timeout and temporary",
			err:      &mockNetError{timeout: true, temporary: true, message: "both"},
			expected: true,
		},
		{
			name:     "non-retryable net error",
			err:      &mockNetError{message: "permanent"},
			expected: false,
		},
		{
			name:     "rate limit error",
			err:      &RateLimitError{},
			expected: true,
		},
		{
			name:     "API error 429",
			err:      NewAPIError(http.StatusTooManyRequests, "rate limit", ""),
			expected: true,
		},
		{
			name:     "API error 500",
			err:      NewAPIError(http.StatusInternalServerError, "server error", ""),
			expected: true,
		},
		{
			name:     "API error 502",
			err:      NewAPIError(http.StatusBadGateway, "bad gateway", ""),
			expected: true,
		},
		{
			name:     "API error 503",
			err:      NewAPIError(http.StatusServiceUnavailable, "unavailable", ""),
			expected: true,
		},
		{
			name:     "API error 504",
			err:      NewAPIError(http.StatusGatewayTimeout, "gateway timeout", ""),
			expected: true,
		},
		{
			name:     "API error 400 (not retryable)",
			err:      NewAPIError(http.StatusBadRequest, "bad request", ""),
			expected: false,
		},
		{
			name:     "generic error",
			err:      errors.New("generic"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err, config)
			if result != tt.expected {
				t.Errorf("isRetryableError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestWithRetry_NetworkErrorRetry(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:          3,
		InitialDelay:         1 * time.Millisecond,
		MaxDelay:             10 * time.Millisecond,
		Multiplier:           2.0,
		JitterFactor:         0.1,
		RetryableStatusCodes: DefaultRetryConfig().RetryableStatusCodes,
	}

	attempt := 0
	errTemporary := &mockNetError{temporary: true, message: "temp"}

	err := WithRetry(context.Background(), config, func(ctx context.Context) error {
		attempt++
		if attempt < 3 {
			return errTemporary
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected success after retries, got error: %v", err)
	}
	if attempt != 3 {
		t.Errorf("expected 3 attempts, got %d", attempt)
	}
}

func TestWithRetry_MaxAttempts(t *testing.T) {
	config := RetryConfig{
		MaxAttempts:          2,
		InitialDelay:         1 * time.Millisecond,
		MaxDelay:             10 * time.Millisecond,
		Multiplier:           2.0,
		JitterFactor:         0.0,
		RetryableStatusCodes: DefaultRetryConfig().RetryableStatusCodes,
	}

	attempt := 0
	errExpected := &mockNetError{temporary: true, message: "temp"}

	err := WithRetry(context.Background(), config, func(ctx context.Context) error {
		attempt++
		return errExpected
	})

	if !errors.Is(err, errExpected) {
		t.Errorf("expected error %v, got %v", errExpected, err)
	}
	if attempt != 2 {
		t.Errorf("expected 2 attempts, got %d", attempt)
	}
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	config := DefaultRetryConfig()
	errExpected := errors.New("non-retryable")

	attempt := 0
	err := WithRetry(context.Background(), config, func(ctx context.Context) error {
		attempt++
		return errExpected
	})

	if !errors.Is(err, errExpected) {
		t.Errorf("expected error %v, got %v", errExpected, err)
	}
	if attempt != 1 {
		t.Errorf("expected 1 attempt, got %d", attempt)
	}
}
