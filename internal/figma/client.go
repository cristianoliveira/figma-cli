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
	return &Client{Token: token, HTTP: http.DefaultClient}
}

// FetchJSON fetches JSON from the given Figma API URL and returns it as a map.
func (c *Client) FetchJSON(apiURL string) (map[string]any, error) {
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Figma-Token", c.Token)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, body)
	}

	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding JSON: %w", err)
	}
	return result, nil
}
