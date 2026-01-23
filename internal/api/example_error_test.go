package api_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/api"
)

func ExampleAPIError() {
	err := &api.APIError{
		StatusCode: http.StatusNotFound,
		Message:    "File not found",
		Details:    "The requested file 'abc123' does not exist",
	}
	fmt.Println(err.Error())
	fmt.Println(api.UserMessage(err))
	// Output:
	// API error (status 404): File not found
	// Resource not found: The requested file 'abc123' does not exist. Please verify the resource identifier.
}

func ExampleRateLimitError() {
	fixedTime := time.Date(2026, 1, 23, 14, 30, 0, 0, time.UTC)
	err := &api.RateLimitError{
		APIError: api.APIError{
			StatusCode: http.StatusTooManyRequests,
			Message:    "Too many requests",
		},
		RetryAfter: 30 * time.Second,
		Limit:      100,
		Remaining:  0,
		Reset:      fixedTime,
	}
	fmt.Println(err.Error())
	// Output:
	// rate limit exceeded (retry after 30s, remaining 0/100, reset at 2026-01-23 14:30:00 +0000 UTC): Too many requests
}

func ExampleAuthError() {
	err := &api.AuthError{
		APIError: api.APIError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Invalid token",
		},
		Reason: "token_expired",
	}
	fmt.Println(err.Error())
	fmt.Println(api.UserMessage(err))
	// Output:
	// authentication error (token_expired): Invalid token
	// Your authentication token has expired. Please re-authenticate using 'figma auth login'.
}

func ExampleWrapError() {
	baseErr := errors.New("network timeout")
	wrapped := api.WrapError(baseErr, "failed to fetch file")
	fmt.Println(wrapped.Error())
	fmt.Println(errors.Unwrap(wrapped) == baseErr)
	// Output:
	// failed to fetch file: network timeout
	// true
}

func ExampleWithRetry() {
	ctx := context.Background()
	config := api.DefaultRetryConfig()
	config.MaxAttempts = 3

	attempt := 0
	err := api.WithRetry(ctx, config, func(ctx context.Context) error {
		attempt++
		if attempt < 3 {
			return &api.RateLimitError{
				APIError: api.APIError{
					StatusCode: http.StatusTooManyRequests,
					Message:    "rate limit",
				},
				RetryAfter: 10 * time.Millisecond,
			}
		}
		return nil // success on third attempt
	})
	if err != nil {
		fmt.Printf("Failed after %d attempts: %v\n", attempt, err)
	} else {
		fmt.Printf("Succeeded on attempt %d\n", attempt)
	}
	// Output:
	// Succeeded on attempt 3
}
