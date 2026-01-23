package transport

import (
	"context"
	"net/http"
)

// Transport defines an interface for executing HTTP requests.
// This abstracts the underlying HTTP client implementation.
type Transport interface {
	// Do executes an HTTP request and returns the response.
	// The caller is responsible for closing the response body.
	Do(ctx context.Context, req *http.Request) (*http.Response, error)
}

// RoundTripper is a wrapper that can modify requests and responses.
type RoundTripper interface {
	http.RoundTripper
}

// RequestBuilder constructs HTTP requests for Figma API endpoints.
type RequestBuilder interface {
	// Build creates an HTTP request for the given method, path, and body.
	Build(ctx context.Context, method, path string, body interface{}) (*http.Request, error)
	// SetHeader sets a header on the request.
	SetHeader(key, value string)
}

// NewRequestBuilder returns a default RequestBuilder for Figma API.
func NewRequestBuilder(baseURL string) RequestBuilder {
	return &defaultRequestBuilder{
		baseURL: baseURL,
		headers: make(map[string]string),
	}
}

// defaultRequestBuilder implements RequestBuilder.
type defaultRequestBuilder struct {
	baseURL string
	headers map[string]string
}

func (b *defaultRequestBuilder) Build(ctx context.Context, method, path string, body interface{}) (*http.Request, error) {
	// Implementation omitted for design phase.
	return nil, nil
}

func (b *defaultRequestBuilder) SetHeader(key, value string) {
	b.headers[key] = value
}
