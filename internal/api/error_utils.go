package api

import (
	"net/http"
	"strconv"
	"time"
)

// NewAPIError creates a new APIError from a status code and message.
func NewAPIError(statusCode int, message string, details string) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Details:    details,
	}
}

// NewAPIErrorWithErr creates a new APIError wrapping an underlying error.
func NewAPIErrorWithErr(statusCode int, message string, details string, err error) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		Details:    details,
		Err:        err,
	}
}

// NewRateLimitError creates a new RateLimitError from rate limit headers.
func NewRateLimitError(statusCode int, message string, retryAfter time.Duration, limit, remaining int, reset time.Time) *RateLimitError {
	return &RateLimitError{
		APIError: APIError{
			StatusCode: statusCode,
			Message:    message,
			Details:    "rate limit exceeded",
		},
		RetryAfter: retryAfter,
		Limit:      limit,
		Remaining:  remaining,
		Reset:      reset,
	}
}

// NewRateLimitErrorFromHeaders creates a RateLimitError from HTTP headers.
// It parses Retry-After (seconds or timestamp), X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset.
func NewRateLimitErrorFromHeaders(statusCode int, message string, headers http.Header) *RateLimitError {
	retryAfter := parseRetryAfter(headers)
	limit := parseIntHeader(headers, "X-RateLimit-Limit", 0)
	remaining := parseIntHeader(headers, "X-RateLimit-Remaining", 0)
	reset := parseTimeHeader(headers, "X-RateLimit-Reset", time.Time{})

	return NewRateLimitError(statusCode, message, retryAfter, limit, remaining, reset)
}

// NewAuthError creates a new AuthError with a specific reason.
func NewAuthError(statusCode int, message string, reason string) *AuthError {
	return &AuthError{
		APIError: APIError{
			StatusCode: statusCode,
			Message:    message,
			Details:    reason,
		},
		Reason: reason,
	}
}

// NewParseError creates a new ParseError.
func NewParseError(err error, data string, location string) *ParseError {
	return &ParseError{
		Err:      err,
		Data:     data,
		Location: location,
	}
}

// ParseRetryAfter parses Retry-After header (seconds or HTTP-date).
func parseRetryAfter(headers http.Header) time.Duration {
	retryAfter := headers.Get("Retry-After")
	if retryAfter == "" {
		return 0
	}
	// Try parsing as seconds (integer)
	if seconds, err := strconv.Atoi(retryAfter); err == nil {
		return time.Duration(seconds) * time.Second
	}
	// Try parsing as HTTP-date (RFC 7231)
	if t, err := http.ParseTime(retryAfter); err == nil {
		now := time.Now()
		if t.After(now) {
			return t.Sub(now)
		}
		return 0
	}
	return 0
}

func parseIntHeader(headers http.Header, key string, defaultValue int) int {
	value := headers.Get(key)
	if value == "" {
		return defaultValue
	}
	if i, err := strconv.Atoi(value); err == nil {
		return i
	}
	return defaultValue
}

func parseTimeHeader(headers http.Header, key string, defaultValue time.Time) time.Time {
	value := headers.Get(key)
	if value == "" {
		return defaultValue
	}
	// Try parsing as Unix timestamp (seconds)
	if unix, err := strconv.ParseInt(value, 10, 64); err == nil {
		return time.Unix(unix, 0)
	}
	// Try parsing as RFC3339
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		return t
	}
	// Try parsing as HTTP-date
	if t, err := http.ParseTime(value); err == nil {
		return t
	}
	return defaultValue
}

// ErrorFromResponse creates an appropriate error type from an HTTP response.
// It reads the response body to extract error details (if any).
func ErrorFromResponse(resp *http.Response, body []byte) error {
	statusCode := resp.StatusCode
	message := http.StatusText(statusCode)
	details := string(body) // Could be JSON; caller may parse further

	switch statusCode {
	case http.StatusTooManyRequests:
		return NewRateLimitErrorFromHeaders(statusCode, message, resp.Header)
	case http.StatusUnauthorized, http.StatusForbidden:
		reason := "authentication_failed"
		if statusCode == http.StatusForbidden {
			reason = "insufficient_permissions"
		}
		return NewAuthError(statusCode, message, reason)
	default:
		return NewAPIError(statusCode, message, details)
	}
}
