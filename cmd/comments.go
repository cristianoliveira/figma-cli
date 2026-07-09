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
  figma comments --id 20089:685897 <file-url>

A node ID from the URL or --id scopes comments to that node and its descendants.
Use --recursive=false to include comments attached only to the selected node.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input, err := figma.ParseInput(args[0])
		if err != nil {
			return err
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}
		nodeIDs := figma.ResolveNodeIDs(input, commentsNodeID)
		recursive, _ := cmd.Flags().GetBool("recursive")
		apiURL := figma.BuildCommentsURL(input.FileID, "")
		var response api.GetCommentsResponse
		if err := client.Fetch(apiURL, &response); err != nil {
			return err
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
		if len(nodeIDs) > 0 {
			documents, err := figma.FetchNodeDocuments(client, input.FileID, nodeIDs)
			if err != nil {
				return err
			}
			outputs = extract.FilterCommentsByNodeIDs(outputs, extract.CommentNodeIDs(documents, recursive))
		}
		if err := cli.NewPrinter(cmd).JSON(outputs); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	commentsCmd.Flags().StringVar(&commentsNodeID, "id", "", "filter comments to a specific node ID; overrides URL node-id")
	commentsCmd.Flags().Bool("recursive", true, "include comments from all descendant nodes")
	rootCmd.AddCommand(commentsCmd)
}
