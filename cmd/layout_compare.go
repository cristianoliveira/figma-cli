package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

type layoutCompareOutput struct {
	Scope output.Scope `json:"scope"`
	extract.LayoutComparison
}

func newLayoutCompareCommand(loadClient func() (*figma.Client, error)) *cobra.Command {
	command := &cobra.Command{
		Use:   "compare [figma-url-or-file-id] (--id frame-id --id frame-id | --name frame --name frame)",
		Short: "Compare explicitly selected responsive frames",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			requestedIDs, _ := cmd.Flags().GetStringArray("id")
			requestedNames, _ := cmd.Flags().GetStringArray("name")
			if len(requestedIDs) > 0 && len(requestedNames) > 0 {
				return fmt.Errorf("layout compare accepts either --id or --name, not both")
			}
			if len(requestedIDs) < 2 && len(requestedNames) < 2 {
				return fmt.Errorf("layout compare requires at least two --id or --name values")
			}
			nodeIDs := make([]string, 0, len(requestedIDs))
			seen := make(map[string]struct{}, len(requestedIDs))
			for _, requestedID := range requestedIDs {
				nodeID := figma.NormalizeNodeID(requestedID)
				if _, exists := seen[nodeID]; exists {
					return fmt.Errorf("layout compare requires distinct frame IDs; %s is repeated", nodeID)
				}
				seen[nodeID] = struct{}{}
				nodeIDs = append(nodeIDs, nodeID)
			}
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return err
			}
			var sectionNodeID string
			if len(requestedNames) > 0 {
				sectionNodeID, err = figma.ResolveSingleNodeID(input, "", "layout compare by name")
				if err != nil {
					return err
				}
			}
			client, err := loadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			var documents []any
			if len(requestedNames) > 0 {
				sectionDocuments, fetchErr := figma.FetchNodeDocuments(client, input.FileID, []string{sectionNodeID})
				if fetchErr != nil {
					return fetchErr
				}
				documents, err = extract.FindLayoutVariantsByName(sectionDocuments[0], requestedNames)
				if err != nil {
					return err
				}
				nodeIDs = make([]string, 0, len(documents))
				for _, document := range documents {
					node, _ := document.(map[string]any)
					nodeIDs = append(nodeIDs, extract.StringValue(node["id"]))
				}
			} else {
				documents, err = figma.FetchNodeDocuments(client, input.FileID, nodeIDs)
				if err != nil {
					return err
				}
			}
			result := layoutCompareOutput{
				Scope:            output.Scope{FileKey: input.FileID, NodeIDs: nodeIDs},
				LayoutComparison: extract.CompareLayouts(documents),
			}
			return cli.NewPrinter(cmd).JSON(result)
		},
	}
	command.Flags().StringArray("id", nil, "frame ID to compare; repeat in responsive order")
	command.Flags().StringArray("name", nil, "exact frame name under URL scope; repeat in responsive order")
	return command
}
