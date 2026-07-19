package figma

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// defaultHTTPTimeout allows large Figma documents and exports to complete while
// still preventing a CLI request from waiting forever.
const defaultHTTPTimeout = 2 * time.Minute

// Client handles communication with the Figma API.
type Client struct {
	Token   string
	HTTP    *http.Client
	context context.Context
}

// ResponseError reports only status needed for domain-level recovery. Provider
// response bodies are intentionally not retained because they may contain
// implementation details or secrets.
type ResponseError struct {
	StatusCode int
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("Figma API returned status %d", e.StatusCode)
}

// NewClient creates a Figma API client with an explicit request timeout.
func NewClient(token string) *Client {
	return &Client{
		Token: token,
		HTTP:  &http.Client{Timeout: defaultHTTPTimeout},
	}
}

// WithContext returns a client copy that binds requests to ctx.
func (c *Client) WithContext(ctx context.Context) *Client {
	copy := *c
	copy.context = ctx
	return &copy
}

func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: defaultHTTPTimeout}
}

func (c *Client) requestContext() context.Context {
	if c.context != nil {
		return c.context
	}
	return context.Background()
}

// Fetch performs a GET request to a Figma API URL and decodes the JSON response into target.
func (c *Client) Fetch(url string, target any) error {
	req, err := http.NewRequestWithContext(c.requestContext(), "GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Figma-Token", c.Token)

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return &ResponseError{StatusCode: resp.StatusCode}
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decoding JSON: %w", err)
	}
	return nil
}

// FetchJSON fetches JSON from the given Figma API URL and returns it as a map.
func (c *Client) FetchJSON(apiURL string) (map[string]any, error) {
	var result map[string]any
	if err := c.Fetch(apiURL, &result); err != nil {
		return nil, err
	}
	return result, nil
}
