package cmd

import (
	"github.com/cristianoliveira/figma-cli/internal/api"
	"github.com/cristianoliveira/figma-cli/internal/api/ratelimit"
	"github.com/cristianoliveira/figma-cli/internal/api/transport"
	"github.com/cristianoliveira/figma-cli/internal/config"
	"github.com/cristianoliveira/figma-cli/internal/logging"
)

// newAPIClient creates a new Figma API client from configuration.
// The client stack is: RateLimitTransport → HTTPClient → RetryClient.
func newAPIClient(cfg *config.Config, logger logging.Logger) (api.Client, error) {
	// Create transport with timeout
	httpTransport := transport.NewHTTPTransport(cfg.API.Timeout)

	// Create rate limiter based on tier and seat type
	limiter := ratelimit.NewTokenBucketLimiter(cfg.API.Tier, cfg.API.SeatType, logger)
	// Wrap transport with rate limiting
	rateLimitedTransport := ratelimit.NewRateLimitTransport(httpTransport, limiter, logger)

	// Create request builder with base URL
	requestBuilder := transport.NewRequestBuilder(cfg.API.BaseURL)

	// Set authorization header if token is present
	if cfg.Token != "" {
		switch cfg.TokenType {
		case "oauth":
			requestBuilder.SetHeader("Authorization", "Bearer "+cfg.Token)
		case "pat":
			fallthrough
		default:
			// PAT authentication uses X-Figma-Token header
			requestBuilder.SetHeader("X-Figma-Token", cfg.Token)
		}
	}

	// Create HTTP client with rate-limited transport
	httpClient := transport.NewHTTPClient(rateLimitedTransport, requestBuilder)

	// Wrap with retry client (configure retries based on cfg.API.MaxRetries)
	retryConfig := api.DefaultRetryConfig()
	if cfg.API.MaxRetries >= 0 {
		// MaxAttempts includes initial attempt, so add 1
		retryConfig.MaxAttempts = cfg.API.MaxRetries + 1
	}
	retryClient := api.NewRetryClientWithConfig(httpClient, retryConfig)

	return retryClient, nil
}
