package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
)

// runSimpleFetch is a shared runner for commands that fetch JSON from a Figma endpoint
// without needing node IDs or extra flags.
func runSimpleFetch(args []string, buildURL func(fileID string) (string, error), label string) {
	input, err := figma.ParseInput(args[0])
	if err != nil {
		cli.Die(err)
	}

	client, err := cli.LoadClient()
	if err != nil {
		cli.Die(err)
	}

	apiURL, err := buildURL(input.FileID)
	if err != nil {
		cli.Die(fmt.Errorf("building %s URL: %w", label, err))
	}

	result, err := client.FetchJSON(apiURL)
	if err != nil {
		cli.Die(fmt.Errorf("fetching %s: %w", label, err))
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		cli.Die(err)
	}
	fmt.Println(string(output))
}
