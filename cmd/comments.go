package cmd

import (
	"fmt"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

const (
	commentStateAll      = "all"
	commentStateOpen     = "open"
	commentStateResolved = "resolved"
)

var commentsCmd = newCommentsCommand(cli.LoadClient)

func newCommentsCommand(loadClient func() (*figma.Client, error)) *cobra.Command {
	var nodeID string
	var state string
	var author string
	var after string
	var before string
	var includeAncestors bool

	command := &cobra.Command{
		Use:   "comments [file-id-or-url]",
		Short: "Fetch comments for a Figma file",
		Long: `Fetch comments for a Figma file via the Figma API.

Requires FIGMA_ACCESS_TOKEN environment variable set with a personal access token.
Examples:
  figma comments grnVU2vAihHXwYgHryu2xE
  figma comments https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=4-1082&p=f&m=dev
  figma comments --id 20089:685897 <file-url>

A node ID from the URL or --id scopes comments to that node and its descendants.
Use --recursive=false to include comments attached only to the selected node.
A numeric URL fragment selects that exact comment regardless of node scope.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateCommentFilters(state, after, before); err != nil {
				return err
			}
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return err
			}
			client, err := loadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			nodeIDs := figma.ResolveNodeIDs(input, nodeID)
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
				out.URL = figma.BuildCommentWebURL(input.FileID, out.NodeID, out.ID)
				outputs = append(outputs, out)
			}
			if input.CommentID != "" {
				outputs = extract.FilterCommentsByID(outputs, input.CommentID)
				if len(outputs) == 0 {
					return fmt.Errorf("comment %s not found", input.CommentID)
				}
			} else if len(nodeIDs) > 0 {
				documents, err := figma.FetchNodeDocuments(client, input.FileID, nodeIDs)
				if err != nil {
					return err
				}
				extract.AttachCommentNodePaths(outputs, extract.CommentNodePaths(documents))
				scopeIDs := extract.CommentNodeIDs(documents, recursive)
				if includeAncestors {
					fileDocument, err := figma.FetchDocument(client, input.FileID, nodeIDs, "", "")
					if err != nil {
						return err
					}
					extract.AttachCommentNodePaths(outputs, extract.CommentNodePaths([]any{fileDocument}))
					for _, nodeID := range nodeIDs {
						for ancestorID := range extract.AncestorNodeIDs(fileDocument, nodeID) {
							scopeIDs[ancestorID] = struct{}{}
						}
					}
				}
				outputs = extract.FilterCommentsByNodeIDs(outputs, scopeIDs)
			}
			threads := extract.GroupCommentThreads(outputs)
			threads = extract.FilterCommentThreads(threads, state, author, after, before)
			result := output.NewQuery(output.Scope{FileKey: input.FileID, NodeIDs: nodeIDs}, threads)
			if err := cli.NewPrinter(cmd).JSON(result); err != nil {
				return err
			}
			return nil
		},
	}
	command.Flags().StringVar(&nodeID, "id", "", "filter comments to a specific node ID; overrides URL node-id")
	command.Flags().Bool("recursive", true, "include comments from all descendant nodes")
	command.Flags().StringVar(&state, "state", commentStateAll, "filter threads by all, open, or resolved")
	command.Flags().StringVar(&author, "author", "", "filter threads by root or reply author")
	command.Flags().StringVar(&after, "after", "", "include threads created at or after RFC3339 timestamp")
	command.Flags().StringVar(&before, "before", "", "include threads created at or before RFC3339 timestamp")
	command.Flags().BoolVar(&includeAncestors, "include-ancestors", false, "include comments attached to ancestor nodes")
	return command
}

func validateCommentFilters(state, after, before string) error {
	if state != commentStateAll && state != commentStateOpen && state != commentStateResolved {
		return fmt.Errorf("invalid comment state %q: expected all, open, or resolved", state)
	}
	timestamps := []struct{ flag, value string }{{"after", after}, {"before", before}}
	for _, timestamp := range timestamps {
		if timestamp.value == "" {
			continue
		}
		if _, err := time.Parse(time.RFC3339, timestamp.value); err != nil {
			return fmt.Errorf("invalid --%s timestamp %q: expected RFC3339", timestamp.flag, timestamp.value)
		}
	}
	if after != "" && before != "" && after > before {
		return fmt.Errorf("--after must not be later than --before")
	}
	return nil
}

func init() {
	rootCmd.AddCommand(commentsCmd)
}
