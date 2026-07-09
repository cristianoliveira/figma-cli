// Package cli holds shared command-line runtime helpers — error handling and
// dependency wiring used across every cobra command in cmd/. Keeping them here
// lets cmd/ contain only command definitions.
package cli

import (
	"os"

	"github.com/cristianoliveira/figma-cli/internal/env"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

// NewPrinter builds an output.Printer bound to stdout, reading the global
// --json flag from cmd. Centralising this keeps the flag lookup in one place
// and stops every command from re-reading the same persistent flag.
func NewPrinter(cmd *cobra.Command) *output.Printer {
	asJSON, _ := cmd.Flags().GetBool("json")
	return output.New(os.Stdout, asJSON)
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
