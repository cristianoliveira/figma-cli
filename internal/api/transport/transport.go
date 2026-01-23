package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	// Construct full URL
	baseURL := strings.TrimSuffix(b.baseURL, "/")
	path = strings.TrimPrefix(path, "/")
	fullURL := fmt.Sprintf("%s/%s", baseURL, path)

	var reqBody io.Reader
	if body != nil {
		// Marshal JSON body
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// Set default headers
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	for key, value := range b.headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

func (b *defaultRequestBuilder) SetHeader(key, value string) {
	b.headers[key] = value
}
