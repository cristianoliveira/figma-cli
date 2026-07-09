package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/env"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/spf13/cobra"
)

var commentsNodeID string

type commentOutput struct {
	ID        string `json:"id"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
	Resolved  bool   `json:"resolved"`
	NodeID    string `json:"node_id,omitempty"`
	User      string `json:"user"`
	ParentID  string `json:"parent_id,omitempty"`
}

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
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		token, err := env.GetFigmaToken()
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}

		var nodeID string
		if commentsNodeID != "" {
			nodeID = figma.NormalizeNodeID(commentsNodeID)
		}
		apiURL := figma.BuildCommentsURL(input.FileID, nodeID)

		client := figma.NewClient(token)
		var response api.GetCommentsResponse
		if err := client.Fetch(apiURL, &response); err != nil {
			fmt.Fprintf(os.Stderr, "error fetching comments: %v\n", err)
			os.Exit(1)
		}

		outputs := make([]commentOutput, 0, len(response.Comments))
		for _, c := range response.Comments {
			out := commentOutput{
				ID:        c.Id,
				Message:   c.Message,
				CreatedAt: c.CreatedAt.String(),
				Resolved:  c.ResolvedAt != nil,
				User:      c.User.Handle,
			}
			if c.ParentId != nil {
				out.ParentID = *c.ParentId
			}
			out.NodeID = extractNodeIDFromClientMeta(c.ClientMeta)
			outputs = append(outputs, out)
		}

		result, err := json.MarshalIndent(outputs, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error formatting output: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(result))
	},
}

// extractNodeIDFromClientMeta extracts the node_id from a ClientMeta union.
// ClientMeta is a discriminated union of Vector | FrameOffset | FrameOffsetRegion.
// Vectors have no node_id; FrameOffsets and FrameOffsetRegions do.
func extractNodeIDFromClientMeta(cm api.Comment_ClientMeta) string {
	b, err := cm.MarshalJSON()
	if err != nil {
		return ""
	}
	var raw struct {
		NodeID string `json:"node_id"`
	}
	if err := json.Unmarshal(b, &raw); err != nil {
		return ""
	}
	return raw.NodeID
}

func init() {
	commentsCmd.Flags().StringVar(&commentsNodeID, "id", "", "Filter comments to a specific node ID")
	rootCmd.AddCommand(commentsCmd)
}
