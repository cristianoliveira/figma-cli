package cmd

import (
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

// newMetaCommand constructs `figma meta` using the explicit loadClient
// dependency.
func newMetaCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:   "meta [file-id-or-url]",
		Short: "Fetch metadata for a Figma file",
		Long:  "Fetch file metadata using FIGMA_ACCESS_TOKEN.",
		Example: `  figma meta <file-key>
  figma meta https://www.figma.com/design/<file-key>/<name>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return cli.NewUsageError(err)
			}
			apiURL, err := figma.BuildFileURL(input.FileID, input.NodeIDs, "", "1")
			if err != nil {
				return cli.NewUsageError(err)
			}
			client, err := deps.LoadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())

			result, err := client.FetchJSON(apiURL)
			if err != nil {
				return err
			}

			if err := cli.NewPrinter(cmd).Structured(result); err != nil {
				return err
			}
			return nil
		},
	}
}
