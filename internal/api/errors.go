package api

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// UserMessager is an interface for errors that provide user-friendly messages.
type UserMessager interface {
	UserMessage() string
}

// APIError represents a generic API error with status code and message.
type APIError struct {
	StatusCode int    // HTTP status code
	Message    string // Original error message
	Details    string // Additional details from API response
	Err        error  // Wrapped error, if any
}

func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("API error (status %d): %s: %v", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

func (e *APIError) Unwrap() error {
	return e.Err
}

// UserMessage returns a user-friendly error message with recovery suggestions.
func (e *APIError) UserMessage() string {
	switch e.StatusCode {
	case http.StatusBadRequest:
		return fmt.Sprintf("Invalid request: %s. Please check your input and try again.", e.Details)
	case http.StatusUnauthorized:
		return "Authentication failed. Please re-authenticate using 'figma auth login'."
	case http.StatusForbidden:
		return fmt.Sprintf("Access denied: %s. You may not have permission to access this resource.", e.Details)
	case http.StatusNotFound:
		return fmt.Sprintf("Resource not found: %s. Please verify the resource identifier.", e.Details)
	case http.StatusTooManyRequests:
		return fmt.Sprintf("Rate limit exceeded: %s. Please wait a moment and try again.", e.Details)
	case http.StatusInternalServerError:
		return fmt.Sprintf("Server error: %s. Please try again later.", e.Details)
	default:
		return fmt.Sprintf("An error occurred: %s (status %d).", e.Message, e.StatusCode)
	}
}

// RateLimitError represents a rate limiting error from the API.
type RateLimitError struct {
	APIError
	RetryAfter time.Duration // Duration to wait before retrying (from Retry-After header)
	Limit      int           // Rate limit limit
	Remaining  int           // Remaining requests
	Reset      time.Time     // When the rate limit resets
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded (retry after %v, remaining %d/%d, reset at %v): %s",
		e.RetryAfter, e.Remaining, e.Limit, e.Reset, e.Message)
}

// UserMessage returns a user-friendly message with retry suggestion.
func (e *RateLimitError) UserMessage() string {
	if e.RetryAfter > 0 {
		return fmt.Sprintf("Rate limit exceeded. Please wait %v before trying again.", e.RetryAfter)
	}
	return fmt.Sprintf("Rate limit exceeded. Please wait until %v before trying again.", e.Reset)
}

// AuthError represents an authentication error.
type AuthError struct {
	APIError
	Reason string // Specific reason, e.g., "token_expired", "invalid_token"
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("authentication error (%s): %s", e.Reason, e.Message)
}

// UserMessage returns a user-friendly message with authentication suggestion.
func (e *AuthError) UserMessage() string {
	switch e.Reason {
	case "token_expired":
		return "Your authentication token has expired. Please re-authenticate using 'figma auth login'."
	case "invalid_token":
		return "Invalid authentication token. Please re-authenticate using 'figma auth login'."
	case "missing_token":
		return "No authentication token found. Please authenticate using 'figma auth login'."
	default:
		return fmt.Sprintf("Authentication failed: %s. Please re-authenticate using 'figma auth login'.", e.Reason)
	}
}

// ParseError represents an error parsing API responses.
type ParseError struct {
	Err      error  // Wrapped error
	Data     string // Raw data that failed to parse (may be truncated)
	Location string // Location (line:column) if applicable
}

func (e *ParseError) Error() string {
	if e.Location != "" {
		return fmt.Sprintf("parse error at %s: %v", e.Location, e.Err)
	}
	return fmt.Sprintf("parse error: %v", e.Err)
}

func (e *ParseError) Unwrap() error {
	return e.Err
}

// UserMessage returns a user-friendly message.
func (e *ParseError) UserMessage() string {
	return "Failed to parse response from Figma API. This may be due to a temporary issue. Please try again later."
}

// WrapError wraps an error with additional context.
func WrapError(err error, context string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

// WrapErrorf wraps an error with formatted context.
func WrapErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

// IsAPIError checks if an error is an APIError (or its subtypes).
func IsAPIError(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr)
}

// IsRateLimitError checks if an error is a RateLimitError.
func IsRateLimitError(err error) bool {
	var rateLimitErr *RateLimitError
	return errors.As(err, &rateLimitErr)
}

// IsAuthError checks if an error is an AuthError.
func IsAuthError(err error) bool {
	var authErr *AuthError
	return errors.As(err, &authErr)
}

// IsParseError checks if an error is a ParseError.
func IsParseError(err error) bool {
	var parseErr *ParseError
	return errors.As(err, &parseErr)
}

// findUserMessager finds the first error in the chain that implements UserMessager.
func findUserMessager(err error) UserMessager {
	for e := err; e != nil; {
		if um, ok := e.(UserMessager); ok {
			return um
		}
		// Unwrap the error (support both standard Unwrap and wrapped errors)
		e = errors.Unwrap(e)
	}
	return nil
}

// UserMessage returns a user-friendly message for an error.
// If the error implements UserMessager, its UserMessage method is used.
// Otherwise, the error's Error() string is returned.
func UserMessage(err error) string {
	if um := findUserMessager(err); um != nil {
		return um.UserMessage()
	}
	return err.Error()
}

// Common errors (deprecated, but kept for compatibility)
var (
	// ErrNotFound is returned when a requested resource is not found.
	ErrNotFound = &APIError{
		StatusCode: http.StatusNotFound,
		Message:    "not found",
	}
	// ErrInvalidRequest is returned when the request is malformed.
	ErrInvalidRequest = &APIError{
		StatusCode: http.StatusBadRequest,
		Message:    "invalid request",
	}
)
