package transport

import (
	"context"
	"net/http"
	"time"
)

// HTTPTransport implements Transport using a standard http.Client.
type HTTPTransport struct {
	client *http.Client
}

// NewHTTPTransport creates a new HTTPTransport with default settings.
// The underlying http.Client uses the provided timeout for requests.
func NewHTTPTransport(timeout time.Duration) *HTTPTransport {
	return &HTTPTransport{
		client: &http.Client{
			Timeout: timeout,
			// Use default transport; can be customized via SetClient.
		},
	}
}

// NewHTTPTransportWithClient creates a new HTTPTransport with a custom http.Client.
func NewHTTPTransportWithClient(client *http.Client) *HTTPTransport {
	return &HTTPTransport{
		client: client,
	}
}

// SetClient replaces the underlying http.Client.
func (t *HTTPTransport) SetClient(client *http.Client) {
	t.client = client
}

// Do executes an HTTP request and returns the response.
// The caller is responsible for closing the response body.
func (t *HTTPTransport) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Associate the context with the request
	req = req.WithContext(ctx)
	return t.client.Do(req)
}
