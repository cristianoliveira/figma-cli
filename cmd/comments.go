package cmd

import (
	"fmt"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/comments"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
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
		Example: `  figma comments <file-key>
  figma comments --state open <url>
  figma comments --mine "<url>?node-id=42-1"`,
		Long: `Fetch comments for a Figma file via the Figma API.

Requires FIGMA_ACCESS_TOKEN environment variable set with a personal access token.
Examples:
  figma comments exampleFileKey123
  figma comments https://www.figma.com/design/exampleFileKey123/Example-Design?node-id=4-1082&p=f&m=dev
  figma comments --id 20089:685897 <file-url>

A node ID from the URL or --id scopes comments to that node and its descendants.
Use --recursive=false to include comments attached only to the selected node.
A numeric URL fragment selects that exact comment regardless of node scope.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateCommentFilters(state, after, before); err != nil {
				return cli.NewUsageError(err)
			}
			resultLimit, err := readResultLimit(cmd)
			if err != nil {
				return err
			}
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return cli.NewUsageError(err)
			}
			client, err := loadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			nodeIDs := figma.ResolveNodeIDs(input, nodeID)
			recursive, _ := cmd.Flags().GetBool("recursive")
			outputs, err := comments.Fetch(client, input.FileID)
			if err != nil {
				return err
			}
			if input.CommentID != "" {
				outputs = extract.FilterCommentsByID(outputs, input.CommentID)
				if len(outputs) == 0 {
					return fmt.Errorf("comment %s not found", input.CommentID)
				}
			} else if len(nodeIDs) > 0 {
				outputs, err = comments.Scope(client, input.FileID, nodeIDs, outputs, recursive, includeAncestors)
				if err != nil {
					return err
				}
			}
			threads := extract.GroupCommentThreads(outputs)
			threads = extract.FilterCommentThreads(threads, state, author, after, before)
			threads, total := limitResults(resultLimit, threads)
			query := map[string]any{
				"state":            state,
				"author":           author,
				"after":            after,
				"before":           before,
				"recursive":        recursive,
				"includeAncestors": includeAncestors,
				"commentId":        input.CommentID,
			}
			result := newLimitedQuery(cmd, output.Scope{FileKey: input.FileID, NodeIDs: nodeIDs}, query, total, threads)
			if err := cli.NewPrinter(cmd).Structured(result); err != nil {
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
	addResultLimitFlags(command)
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
