package transport

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/api"
)

// mockTransport implements Transport for testing.
type mockTransport struct {
	requests []*http.Request
	response *http.Response
	err      error
}

func (m *mockTransport) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	m.requests = append(m.requests, req)
	return m.response, m.err
}

// mockNetError implements net.Error for testing.
type mockNetError struct {
	timeout   bool
	temporary bool
	message   string
}

func (e *mockNetError) Error() string   { return e.message }
func (e *mockNetError) Timeout() bool   { return e.timeout }
func (e *mockNetError) Temporary() bool { return e.temporary }

func TestHTTPClient_RetryOnNetworkError(t *testing.T) {
	mock := &mockTransport{}
	// Simulate a temporary network error
	mock.err = &mockNetError{temporary: true, message: "temporary network error"}

	rb := NewRequestBuilder("https://api.figma.com")
	hc := NewHTTPClient(mock, rb)
	config := api.RetryConfig{
		MaxAttempts:          3,
		InitialDelay:         1 * time.Millisecond,
		MaxDelay:             10 * time.Millisecond,
		Multiplier:           2.0,
		JitterFactor:         0.0,
		RetryableStatusCodes: api.DefaultRetryConfig().RetryableStatusCodes,
	}
	retryClient := api.NewRetryClientWithConfig(hc, config)

	ctx := context.Background()
	_, err := retryClient.GetFile(ctx, "test")
	if err == nil {
		t.Error("expected error")
	}
	// Should have retried 3 times (first attempt + two retries)
	if len(mock.requests) != 3 {
		t.Errorf("expected 3 attempts, got %d", len(mock.requests))
	}
}

func TestHTTPClient_RetryOn5xx(t *testing.T) {
	mock := &mockTransport{}
	// Simulate 503 Service Unavailable
	mock.response = &http.Response{
		StatusCode: http.StatusServiceUnavailable,
		Status:     "503 Service Unavailable",
		Body:       http.NoBody,
		Header:     make(http.Header),
	}

	rb := NewRequestBuilder("https://api.figma.com")
	hc := NewHTTPClient(mock, rb)
	config := api.RetryConfig{
		MaxAttempts:          2,
		InitialDelay:         1 * time.Millisecond,
		MaxDelay:             10 * time.Millisecond,
		Multiplier:           2.0,
		JitterFactor:         0.0,
		RetryableStatusCodes: api.DefaultRetryConfig().RetryableStatusCodes,
	}
	retryClient := api.NewRetryClientWithConfig(hc, config)

	ctx := context.Background()
	_, err := retryClient.GetFile(ctx, "test")
	if err == nil {
		t.Error("expected error")
	}
	// Should have retried twice (first attempt + one retry)
	if len(mock.requests) != 2 {
		t.Errorf("expected 2 attempts, got %d", len(mock.requests))
	}
}

func TestHTTPClient_NoRetryOn4xx(t *testing.T) {
	mock := &mockTransport{}
	// Simulate 400 Bad Request (not retryable)
	mock.response = &http.Response{
		StatusCode: http.StatusBadRequest,
		Status:     "400 Bad Request",
		Body:       http.NoBody,
		Header:     make(http.Header),
	}

	rb := NewRequestBuilder("https://api.figma.com")
	hc := NewHTTPClient(mock, rb)
	config := api.RetryConfig{
		MaxAttempts:          3,
		InitialDelay:         1 * time.Millisecond,
		MaxDelay:             10 * time.Millisecond,
		Multiplier:           2.0,
		JitterFactor:         0.0,
		RetryableStatusCodes: api.DefaultRetryConfig().RetryableStatusCodes,
	}
	retryClient := api.NewRetryClientWithConfig(hc, config)

	ctx := context.Background()
	_, err := retryClient.GetFile(ctx, "test")
	if err == nil {
		t.Error("expected error")
	}
	// Should NOT retry, only one attempt
	if len(mock.requests) != 1 {
		t.Errorf("expected 1 attempt, got %d", len(mock.requests))
	}
}

func TestHTTPClient_RetryOn429(t *testing.T) {
	mock := &mockTransport{}
	// Simulate 429 Too Many Requests with Retry-After header
	header := make(http.Header)
	header.Set("Retry-After", "1") // 1 second
	mock.response = &http.Response{
		StatusCode: http.StatusTooManyRequests,
		Status:     "429 Too Many Requests",
		Body:       http.NoBody,
		Header:     header,
	}

	rb := NewRequestBuilder("https://api.figma.com")
	hc := NewHTTPClient(mock, rb)
	config := api.RetryConfig{
		MaxAttempts:          2,
		InitialDelay:         1 * time.Millisecond,
		MaxDelay:             10 * time.Millisecond,
		Multiplier:           2.0,
		JitterFactor:         0.0,
		RetryableStatusCodes: api.DefaultRetryConfig().RetryableStatusCodes,
	}
	retryClient := api.NewRetryClientWithConfig(hc, config)

	ctx := context.Background()
	_, err := retryClient.GetFile(ctx, "test")
	if err == nil {
		t.Error("expected error")
	}
	// Should retry twice (first attempt + one retry)
	if len(mock.requests) != 2 {
		t.Errorf("expected 2 attempts, got %d", len(mock.requests))
	}
}
