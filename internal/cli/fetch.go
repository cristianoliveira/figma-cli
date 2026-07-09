package cli

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

// RunSimpleFetch is a shared runner for commands that fetch JSON from a Figma
// endpoint without needing node IDs or extra flags.
func RunSimpleFetch(cmd *cobra.Command, args []string, buildURL func(fileID string) (string, error), label string) {
	input, err := figma.ParseInput(args[0])
	if err != nil {
		Die(err)
	}

	client, err := LoadClient()
	if err != nil {
		Die(err)
	}

	apiURL, err := buildURL(input.FileID)
	if err != nil {
		Die(fmt.Errorf("building %s URL: %w", label, err))
	}

	result, err := client.FetchJSON(apiURL)
	if err != nil {
		Die(fmt.Errorf("fetching %s: %w", label, err))
	}

	if err := NewPrinter(cmd).JSON(result); err != nil {
		Die(err)
	}
}
