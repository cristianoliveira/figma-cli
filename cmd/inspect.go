package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

var inspectCmd = newInspectCommand(cli.LoadClient)

func newInspectCommand(loadClient func() (*figma.Client, error)) *cobra.Command {
	return newInspectCommandWithVariables(loadClient, figma.FetchVariables)
}

func newInspectCommandWithVariables(
	loadClient func() (*figma.Client, error),
	fetchVariables func(*figma.Client, string) (map[string]any, error),
) *cobra.Command {
	command := &cobra.Command{
		Use:   "inspect [figma-url-or-file-id]",
		Short: "Show a curated summary of a specific Figma node",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			explicitNodeID, _ := cmd.Flags().GetString("id")
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return err
			}
			nodeID, err := inspectNodeID(input, explicitNodeID)
			if err != nil {
				return err
			}
			client, err := loadClient()
			if err != nil {
				return err
			}
			details, err := figma.FetchNodeDetails(client, input.FileID, []string{nodeID})
			if err != nil {
				return err
			}
			document, ok := details.Documents[0].(map[string]any)
			if !ok {
				return fmt.Errorf("node %s has an invalid document", nodeID)
			}
			node := extract.NodeToInspectOutput(document)
			extract.ResolveInspectStyleBindings(&node, details.Styles)
			if len(node.VariableBindings) > 0 {
				variables, fetchErr := fetchVariables(client, input.FileID)
				if fetchErr == nil {
					extract.ResolveInspectVariableBindings(&node, variables)
				}
			}
			result := output.Detail[extract.InspectOutput]{
				Scope:  output.Scope{FileKey: input.FileID, NodeIDs: []string{nodeID}},
				Result: node,
			}
			if err := cli.NewPrinter(cmd).JSON(result); err != nil {
				return err
			}
			return nil
		},
	}
	command.Flags().String("id", "", "node ID to inspect; defaults to URL node-id")
	return command
}

func inspectNodeID(input *figma.FileInput, explicitNodeID string) (string, error) {
	return figma.ResolveSingleNodeID(input, explicitNodeID, "inspect")
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
