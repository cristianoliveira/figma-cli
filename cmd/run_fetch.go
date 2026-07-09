package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/env"
	"github.com/cristianoliveira/figma-cli/internal/figma"
)

// runSimpleFetch is a shared runner for commands that fetch JSON from a Figma endpoint
// without needing node IDs or extra flags.
func runSimpleFetch(args []string, buildURL func(fileID string) (string, error), label string) {
	input, err := figma.ParseInput(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	token, err := env.GetFigmaToken()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	apiURL, err := buildURL(input.FileID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error building %s URL: %v\n", label, err)
		os.Exit(1)
	}

	client := figma.NewClient(token)
	result, err := client.FetchJSON(apiURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching %s: %v\n", label, err)
		os.Exit(1)
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error formatting output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(output))
}
