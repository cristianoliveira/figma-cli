package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

const defaultLayoutDepth = 4

type layoutOutput struct {
	Scope     output.Scope            `json:"scope"`
	Query     map[string]any          `json:"query"`
	Traversal extract.LayoutTraversal `json:"traversal"`
	Result    extract.LayoutNode      `json:"result"`
}

func readLayoutDepth(command *cobra.Command) (int, map[string]any, error) {
	depth, _ := command.Flags().GetInt("depth")
	full, _ := command.Flags().GetBool("full")
	if depth < 0 {
		return 0, nil, cli.NewUsageError(fmt.Errorf("--depth must be zero or greater"))
	}
	if full && command.Flags().Changed("depth") {
		return 0, nil, cli.NewUsageError(fmt.Errorf("--full cannot be combined with --depth"))
	}
	if full {
		return -1, map[string]any{"full": true}, nil
	}
	return depth, map[string]any{"maxDepth": depth}, nil
}

// newLayoutCommand constructs `figma layout` and attaches its nested
// `figma layout compare` subcommand. Both share the loadClient dependency.
func newLayoutCommand(deps Deps) *cobra.Command {
	command := &cobra.Command{
		Use:   "layout [figma-url-or-file-id]",
		Short: "Show an ordered frame tree with layout and copy",
		Example: `  figma layout "<url>?node-id=42-1"
  figma layout --depth 2 --id 42:1 <file-key>
  figma layout --full --measure-spacing "<url>?node-id=42-1"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			maxDepth, query, err := readLayoutDepth(cmd)
			if err != nil {
				return err
			}
			nodeID, err := explicitNodeIDFlag(cmd)
			if err != nil {
				return cli.NewUsageError(err)
			}
			measureSpacing, _ := cmd.Flags().GetBool("measure-spacing")
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return cli.NewUsageError(err)
			}
			resolvedNodeID, err := figma.ResolveSingleNodeID(input, nodeID, "layout")
			if err != nil {
				return cli.NewUsageError(err)
			}

			client, err := deps.LoadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			documents, err := figma.FetchNodeDocuments(client, input.FileID, []string{resolvedNodeID})
			if err != nil {
				return err
			}
			layout := extract.ExtractLayoutWithDepth(documents[0], maxDepth, extract.LayoutOptions{MeasureSpacing: measureSpacing})
			result := layoutOutput{
				Scope:     output.Scope{FileKey: input.FileID, NodeIDs: []string{resolvedNodeID}},
				Query:     query,
				Traversal: layout.Traversal,
				Result:    layout.Result,
			}
			return cli.NewPrinter(cmd).Structured(result)
		},
	}
	addNodeIDFlag(command, "node ID to inspect; defaults to URL node-id")
	command.Flags().Bool("measure-spacing", false, "measure geometric gaps between adjacent layout children")
	command.Flags().Int("depth", defaultLayoutDepth, "maximum descendant depth to return")
	command.Flags().Bool("full", false, "return every layout descendant")
	command.AddCommand(newLayoutCompareCommand(deps))
	return command
}
