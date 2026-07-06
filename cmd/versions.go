package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

func buildVersionsAPIURL(fileID string) string {
	return fmt.Sprintf("https://api.figma.com/v1/files/%s/versions", fileID)
}

var versionsCmd = &cobra.Command{
	Use:   "versions [file-id-or-url]",
	Short: "Fetch version history for a Figma file",
	Long: `Fetch version history for a Figma file via the Figma API.

Requires FIGMA_ACCESS_TOKEN environment variable set with a personal access token.
Examples:
  figma versions grnVU2vAihHXwYgHryu2xE
  figma versions https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=4-1082&p=f&m=dev`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input, err := parseInput(args[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		token, err := getFigmaToken()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		req, err := http.NewRequest("GET", buildVersionsAPIURL(input.fileID), nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error creating request: %v\n", err)
			os.Exit(1)
		}
		req.Header.Set("X-Figma-Token", token)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error making request: %v\n", err)
			os.Exit(1)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to close response body: %v\n", err)
			}
		}()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fmt.Fprintf(os.Stderr, "error: API returned status %d: %s\n", resp.StatusCode, body)
			os.Exit(1)
		}

		var result map[string]interface{}
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&result); err != nil {
			fmt.Fprintf(os.Stderr, "error decoding JSON: %v\n", err)
			os.Exit(1)
		}

		output, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error formatting output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(output))
	},
}

func init() {
	rootCmd.AddCommand(versionsCmd)
}
