package figma

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// Client handles communication with the Figma API.
type Client struct {
	Token string
	HTTP  *http.Client
}

// NewClient creates a Figma API client with the given token.
func NewClient(token string) *Client {
	return &Client{Token: token}
}

// httpClient returns the configured HTTP client, defaulting to http.DefaultClient.
func (c *Client) httpClient() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return http.DefaultClient
}

// Fetch performs a GET request to a Figma API URL and decodes the JSON response into target.
func (c *Client) Fetch(url string, target any) error {
	req, err := http.NewRequest("GET", url, nil)
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
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, body)
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
