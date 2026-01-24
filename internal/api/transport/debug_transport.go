package transport

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/logging"
)

// DebugTransport wraps a Transport and logs request and response details.
type DebugTransport struct {
	transport Transport
	logger    logging.Logger
}

// NewDebugTransport creates a new DebugTransport.
func NewDebugTransport(wrapped Transport, logger logging.Logger) *DebugTransport {
	if logger == nil {
		logger = logging.Default()
	}
	return &DebugTransport{
		transport: wrapped,
		logger:    logger,
	}
}

// Do executes the request and logs details.
func (t *DebugTransport) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	// Log request details
	headers := make(map[string]string, len(req.Header))
	for key, values := range req.Header {
		// Redact sensitive headers
		switch strings.ToLower(key) {
		case "authorization", "x-figma-token":
			headers[key] = "[REDACTED]"
		default:
			headers[key] = strings.Join(values, ", ")
		}
	}
	t.logger.Debug(ctx, "HTTP request",
		logging.String("method", req.Method),
		logging.String("url", req.URL.String()),
		logging.Any("headers", headers),
	)

	// Execute request
	resp, err := t.transport.Do(ctx, req)
	if err != nil {
		t.logger.Error(ctx, "HTTP request failed",
			logging.String("method", req.Method),
			logging.String("url", req.URL.String()),
			logging.Err(err),
		)
		return resp, err
	}

	// Log response details
	respHeaders := make(map[string]string, len(resp.Header))
	for key, values := range resp.Header {
		respHeaders[key] = strings.Join(values, ", ")
	}
	t.logger.Debug(ctx, "HTTP response",
		logging.Int("status", resp.StatusCode),
		logging.String("status_text", resp.Status),
		logging.Any("headers", respHeaders),
	)

	// Log response body for error statuses (4xx, 5xx)
	if resp.StatusCode >= 400 {
		// Read the body but restore it for the caller
		bodyBytes, err := readAndRestoreBody(resp)
		if err != nil {
			t.logger.Warn(ctx, "Failed to read error response body", logging.Err(err))
		} else if len(bodyBytes) > 0 {
			// Truncate long bodies
			bodyStr := string(bodyBytes)
			if len(bodyStr) > 1024 {
				bodyStr = bodyStr[:1024] + "... (truncated)"
			}
			t.logger.Debug(ctx, "Error response body",
				logging.String("body", bodyStr),
			)
		}
	}
	return resp, nil
}

// readAndRestoreBody reads the response body and replaces it with a new reader.
func readAndRestoreBody(resp *http.Response) ([]byte, error) {
	if resp.Body == nil || resp.Body == http.NoBody {
		return nil, nil
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	return bodyBytes, nil
}
