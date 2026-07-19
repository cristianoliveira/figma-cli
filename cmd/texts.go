package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

type textQuery struct {
	Scope     output.Scope   `json:"scope"`
	Query     map[string]any `json:"query"`
	Total     int            `json:"total"`
	Truncated bool           `json:"truncated,omitempty"`
	Hint      string         `json:"hint,omitempty"`
	Results   any            `json:"results"`
}

var textsCmd = newTextsCommand(cli.LoadClient)

func newTextsCommand(loadClient func() (*figma.Client, error)) *cobra.Command {
	command := &cobra.Command{
		Use:   "texts [file-id-or-url]",
		Short: "Extract ordered text and Figma-provided list semantics",
		Example: `  figma texts "<url>?node-id=42-1"
  figma texts --id 42:1 <file-key>
  figma texts --layer "Screen Name" --recursive <url>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			layerName, _ := cmd.Flags().GetString("layer")
			recursive, _ := cmd.Flags().GetBool("recursive")
			nodeID, err := explicitNodeIDFlag(cmd)
			if err != nil {
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

			nodeIDs := figma.ResolveNodeIDs(input, nodeID)
			if layerName == "" {
				resolvedNodeID, resolveErr := figma.ResolveSingleNodeID(input, nodeID, "texts")
				if resolveErr != nil {
					return cli.NewUsageError(fmt.Errorf("--layer or a node ID is required"))
				}
				nodeIDs = []string{resolvedNodeID}
			}

			client, err := loadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			var document any
			if len(nodeIDs) > 0 {
				documents, fetchErr := figma.FetchNodeDocuments(client, input.FileID, nodeIDs)
				if fetchErr != nil {
					return fetchErr
				}
				document = documents[0]
			} else {
				document, err = figma.FetchDocument(client, input.FileID, nil, "", "")
				if err != nil {
					return err
				}
			}
			input.NodeIDs = nodeIDs

			results, err := textResult(document, input, layerName, recursive)
			if err != nil {
				return err
			}
			results, total := limitTextResults(resultLimit, results)
			scope := output.Scope{FileKey: input.FileID, NodeIDs: nodeIDs}
			query := map[string]any{"layer": layerName, "recursive": recursive}
			contract := textQuery{Scope: scope, Query: query, Total: total, Truncated: textResultCount(results) < total, Results: results}
			if contract.Truncated {
				contract.Hint = cli.FullHint(cmd, args)
			}
			return cli.NewPrinter(cmd).Structured(contract)
		},
	}
	command.Flags().String("layer", "", "layer name to extract text from")
	addNodeIDFlag(command, "node ID to extract descendant text from; defaults to URL node-id")
	command.Flags().Bool("recursive", false, "include text from all descendant nodes")
	addResultLimitFlags(command)
	return command
}

func limitTextResults(options resultLimit, results any) (any, int) {
	switch values := results.(type) {
	case []extract.OrderedTextOutput:
		limited, total := limitResults(options, values)
		return limited, total
	case []extract.LayerTextOutput:
		limited, total := limitResults(options, values)
		return limited, total
	default:
		return results, 0
	}
}

func textResultCount(results any) int {
	switch values := results.(type) {
	case []extract.OrderedTextOutput:
		return len(values)
	case []extract.LayerTextOutput:
		return len(values)
	default:
		return 0
	}
}

func textResult(doc any, input *figma.FileInput, layerName string, recursive bool) (any, error) {
	if len(input.NodeIDs) > 0 && layerName == "" {
		return extract.OrderedTextForFrame(doc), nil
	}
	if layerName == "" {
		return nil, fmt.Errorf("--layer or a node ID is required")
	}
	return extract.FindTextByLayerName(doc, layerName, recursive), nil
}

func init() {
	rootCmd.AddCommand(textsCmd)
}
