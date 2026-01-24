package transport

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

func TestHTTPClient_GetFileVersions(t *testing.T) {
	t.Run("paginated response", func(t *testing.T) {
		mock := &mockTransport{}
		// Simulate a paginated response with versions array and pagination metadata
		body := `{"versions": [{"id": "1", "createdAt": "2023-01-01T00:00:00Z", "label": "v1", "description": "First version", "user": {"id": "user1", "handle": "alice", "imgUrl": ""}}], "pagination": {"before": null, "after": "cursor1"}}`
		mock.response = &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}
		rb := NewRequestBuilder("https://api.figma.com")
		hc := NewHTTPClient(mock, rb)

		ctx := context.Background()
		versions, err := hc.GetFileVersions(ctx, "file123")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(versions) != 1 {
			t.Errorf("expected 1 version, got %d", len(versions))
		}
		if versions[0].ID != "1" {
			t.Errorf("expected version ID '1', got %s", versions[0].ID)
		}
		// Check that request URL contains correct path
		if len(mock.requests) != 1 {
			t.Fatalf("expected 1 request, got %d", len(mock.requests))
		}
		req := mock.requests[0]
		if !strings.Contains(req.URL.Path, "/files/file123/versions") {
			t.Errorf("expected path to contain /files/file123/versions, got %s", req.URL.Path)
		}
	})

	t.Run("direct array response", func(t *testing.T) {
		mock := &mockTransport{}
		body := `[{"id": "2", "createdAt": "2023-01-02T00:00:00Z", "label": "v2", "description": "Second version", "user": {"id": "user2", "handle": "bob", "imgUrl": ""}}]`
		mock.response = &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}
		rb := NewRequestBuilder("https://api.figma.com")
		hc := NewHTTPClient(mock, rb)

		ctx := context.Background()
		versions, err := hc.GetFileVersions(ctx, "file456")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(versions) != 1 {
			t.Errorf("expected 1 version, got %d", len(versions))
		}
		if versions[0].ID != "2" {
			t.Errorf("expected version ID '2', got %s", versions[0].ID)
		}
	})

	t.Run("pagination parameters", func(t *testing.T) {
		mock := &mockTransport{}
		body := `{"versions": [], "pagination": {}}`
		mock.response = &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}
		rb := NewRequestBuilder("https://api.figma.com")
		hc := NewHTTPClient(mock, rb)

		ctx := context.Background()
		_, err := hc.GetFileVersions(ctx, "file789", api.WithPageSize(10), api.WithBefore("cursor1"), api.WithAfter("cursor2"), api.WithBranchForVersions("feature/branch"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if len(mock.requests) != 1 {
			t.Fatalf("expected 1 request, got %d", len(mock.requests))
		}
		req := mock.requests[0]
		query := req.URL.RawQuery
		if !strings.Contains(query, "page_size=10") {
			t.Errorf("expected query to contain page_size=10, got %s", query)
		}
		if !strings.Contains(query, "before=cursor1") {
			t.Errorf("expected query to contain before=cursor1, got %s", query)
		}
		if !strings.Contains(query, "after=cursor2") {
			t.Errorf("expected query to contain after=cursor2, got %s", query)
		}
		if !strings.Contains(query, "branch_data=feature%2Fbranch") {
			t.Errorf("expected query to contain branch_data=feature%%2Fbranch, got %s", query)
		}
	})
}

func TestHTTPClient_GetImage(t *testing.T) {
	mock := &mockTransport{}
	// Load fixture relative to project root
	fixturePath := filepath.Join("..", "..", "..", "testdata", "fixtures", "image_response.json")
	body, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to load fixture: %v", err)
	}
	mock.response = &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader(string(body))),
		Header:     make(http.Header),
	}
	rb := NewRequestBuilder("https://api.figma.com")
	hc := NewHTTPClient(mock, rb)

	ctx := context.Background()
	images, err := hc.GetImage(ctx, "file123", []string{"3:1"}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(images) != 1 {
		t.Errorf("expected 1 image, got %d", len(images))
	}
	url, ok := images["3:1"]
	if !ok {
		t.Fatal("expected image for node 3:1")
	}
	expectedURL := "https://s3-us-west-2.amazonaws.com/figma-alpha-api/img/abc/xyz/image.png"
	if url != expectedURL {
		t.Errorf("expected URL %s, got %s", expectedURL, url)
	}
	// Verify request path and query
	if len(mock.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(mock.requests))
	}
	req := mock.requests[0]
	if !strings.Contains(req.URL.Path, "/images/file123") {
		t.Errorf("expected path to contain /images/file123, got %s", req.URL.Path)
	}
	if !strings.Contains(req.URL.RawQuery, "ids=3:1") {
		t.Errorf("expected query to contain ids=3:1, got %s", req.URL.RawQuery)
	}
}
