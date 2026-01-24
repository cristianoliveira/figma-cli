// Package api provides a client for the Figma API.
//
//go:generate go run ./gen -input client.go -output retry_client_gen.go -template gen/templates/retry.tmpl
package api

// RetryClient wraps a Client and adds retry logic for transient errors.
type RetryClient struct {
	client Client
	config RetryConfig
}

// NewRetryClient creates a new RetryClient with default retry configuration.
func NewRetryClient(client Client) *RetryClient {
	return &RetryClient{
		client: client,
		config: DefaultRetryConfig(),
	}
}

// NewRetryClientWithConfig creates a new RetryClient with custom configuration.
func NewRetryClientWithConfig(client Client, config RetryConfig) *RetryClient {
	return &RetryClient{
		client: client,
		config: config,
	}
}
