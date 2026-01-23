package ratelimit

import (
	"context"
	"net/http"
	"time"
)

// Limiter defines an interface for rate limiting.
type Limiter interface {
	// Wait blocks until the request can proceed without exceeding rate limits.
	Wait(ctx context.Context) error
	// Update updates the limiter's state based on response headers.
	Update(headers http.Header)
}

// TokenBucketLimiter implements a token bucket rate limiter.
type TokenBucketLimiter struct {
	// implementation details omitted for design phase
}

// Wait implements Limiter.
func (l *TokenBucketLimiter) Wait(ctx context.Context) error {
	return nil // placeholder
}

// Update implements Limiter.
func (l *TokenBucketLimiter) Update(headers http.Header) {
	// parse X-RateLimit-* headers
}

// HeaderParser extracts rate limit information from HTTP headers.
type HeaderParser interface {
	Parse(headers http.Header) (limit, remaining int, reset time.Time, retryAfter time.Duration)
}

// DefaultHeaderParser implements HeaderParser for Figma API headers.
type DefaultHeaderParser struct{}

func (p *DefaultHeaderParser) Parse(headers http.Header) (limit, remaining int, reset time.Time, retryAfter time.Duration) {
	// placeholder
	return 0, 0, time.Time{}, 0
}
