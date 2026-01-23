package logging

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	// requestIDKey is the key used to store request ID in context.
	requestIDKey contextKey = "request_id"
)

// NewContextWithRequestID creates a new context with a request ID.
// If no request ID is provided, a new UUID v4 is generated.
func NewContextWithRequestID(ctx context.Context, requestID ...string) context.Context {
	var id string
	if len(requestID) > 0 && requestID[0] != "" {
		id = requestID[0]
	} else {
		id = uuid.New().String()
	}
	return context.WithValue(ctx, requestIDKey, id)
}

// RequestIDFromContext extracts the request ID from the context.
// Returns the request ID and true if found, empty string and false otherwise.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	val := ctx.Value(requestIDKey)
	if val == nil {
		return "", false
	}
	id, ok := val.(string)
	return id, ok
}

// RequestID returns the request ID from context, or empty string if not present.
func RequestID(ctx context.Context) string {
	id, _ := RequestIDFromContext(ctx)
	return id
}
