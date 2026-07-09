package cli

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

// RunSimpleFetch is a shared runner for commands that fetch JSON from a Figma
// endpoint without needing node IDs or extra flags. It returns an error so
// cobra commands can use RunE and let Execute handle exit/printing.
func RunSimpleFetch(cmd *cobra.Command, args []string, buildURL func(fileID string) (string, error), label string) error {
	input, err := figma.ParseInput(args[0])
	if err != nil {
		return err
	}

	client, err := LoadClient()
	if err != nil {
		return err
	}

	apiURL, err := buildURL(input.FileID)
	if err != nil {
		return fmt.Errorf("building %s URL: %w", label, err)
	}

	result, err := client.FetchJSON(apiURL)
	if err != nil {
		return fmt.Errorf("fetching %s: %w", label, err)
	}

	if err := NewPrinter(cmd).JSON(result); err != nil {
		return err
	}
	return nil
}
