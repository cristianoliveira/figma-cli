// Package cli holds shared command-line runtime helpers — error handling and
// dependency wiring used across every cobra command in cmd/. Keeping them here
// lets cmd/ contain only command definitions.
package cli

import (
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/env"
	"github.com/cristianoliveira/figma-cli/internal/figma"
)

// Die prints err to stderr and exits with status 1.
// It is the single error-exit path, so wording and exit code live in one place.
func Die(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

// LoadClient builds a Figma API client from the configured access token.
// Centralizing construction means HTTP config (timeouts, base URL, retries)
// has exactly one place to change.
func LoadClient() (*figma.Client, error) {
	token, err := env.GetFigmaToken()
	if err != nil {
		return nil, err
	}
	return figma.NewClient(token), nil
}
