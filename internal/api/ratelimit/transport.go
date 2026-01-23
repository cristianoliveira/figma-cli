package ratelimit

import (
	"context"
	"net/http"

	"github.com/cristianoliveira/figma-cli/internal/api/transport"
	"github.com/cristianoliveira/figma-cli/internal/logging"
)

// RateLimitTransport wraps a transport with rate limiting.
type RateLimitTransport struct {
	transport transport.Transport
	limiter   Limiter
	logger    logging.Logger
}

// NewRateLimitTransport creates a new RateLimitTransport.
func NewRateLimitTransport(wrapped transport.Transport, limiter Limiter, logger logging.Logger) *RateLimitTransport {
	if logger == nil {
		logger = logging.Default()
	}
	return &RateLimitTransport{
		transport: wrapped,
		limiter:   limiter,
		logger:    logger,
	}
}

// Do executes the request after waiting for rate limiter.
func (t *RateLimitTransport) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Wait for rate limiter
	if err := t.limiter.Wait(ctx); err != nil {
		t.logger.Warn(ctx, "Rate limiter wait failed", logging.NewField("error", err))
		return nil, err
	}

	// Execute request
	resp, err := t.transport.Do(ctx, req)
	if err != nil {
		return resp, err
	}

	// Update limiter with response headers
	t.limiter.Update(resp.Header)
	return resp, nil
}
