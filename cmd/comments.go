package cmd

import (
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/spf13/cobra"
)

var commentsNodeID string

var commentsCmd = &cobra.Command{
	Use:   "comments [file-id-or-url]",
	Short: "Fetch comments for a Figma file",
	Long: `Fetch comments for a Figma file via the Figma API.

Requires FIGMA_ACCESS_TOKEN environment variable set with a personal access token.
Examples:
  figma comments grnVU2vAihHXwYgHryu2xE
  figma comments https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=4-1082&p=f&m=dev
  figma comments --id 20089:685897 <file-url>`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		input, err := figma.ParseInput(args[0])
		if err != nil {
			cli.Die(err)
		}
		client, err := cli.LoadClient()
		if err != nil {
			cli.Die(err)
		}
		var nodeID string
		if commentsNodeID != "" {
			nodeID = figma.NormalizeNodeID(commentsNodeID)
		}
		apiURL := figma.BuildCommentsURL(input.FileID, nodeID)
		var response api.GetCommentsResponse
		if err := client.Fetch(apiURL, &response); err != nil {
			cli.Die(err)
		}
		outputs := make([]extract.CommentOutput, 0, len(response.Comments))
		for _, c := range response.Comments {
			out := extract.CommentOutput{
				ID:        c.Id,
				Message:   c.Message,
				CreatedAt: c.CreatedAt.String(),
				Resolved:  c.ResolvedAt != nil,
				User:      c.User.Handle,
			}
			if c.ParentId != nil {
				out.ParentID = *c.ParentId
			}
			out.NodeID = extract.ExtractNodeIDFromClientMeta(c.ClientMeta)
			outputs = append(outputs, out)
		}
		if err := cli.NewPrinter(cmd).JSON(outputs); err != nil {
			cli.Die(err)
		}
	},
}

func init() {
	commentsCmd.Flags().StringVar(&commentsNodeID, "id", "", "Filter comments to a specific node ID")
	rootCmd.AddCommand(commentsCmd)
}
