package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

var componentsCmd = newComponentsCommand(cli.LoadClient)

func newComponentsCommand(loadClient func() (*figma.Client, error)) *cobra.Command {
	command := &cobra.Command{
		Use:   "components [figma-url-or-file-id]",
		Short: "List components, component sets, and instances as JSON",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeID, _ := cmd.Flags().GetString("id")
			nameFilter, _ := cmd.Flags().GetString("name")
			kind, _ := cmd.Flags().GetString("kind")
			raw, _ := cmd.Flags().GetBool("raw")
			usage, _ := cmd.Flags().GetBool("usage")
			if usage && raw {
				return fmt.Errorf("--usage and --raw cannot be used together")
			}
			if usage && kind != "" && kind != "instance" {
				return fmt.Errorf("--usage only supports --kind instance")
			}
			if kind != "" && kind != "component" && kind != "set" && kind != "instance" {
				return fmt.Errorf("invalid component kind %q: expected component, set, or instance", kind)
			}
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return err
			}
			nodeIDs, err := figma.ResolveRequiredNodeIDs(input, nodeID, "components")
			if err != nil {
				return err
			}
			client, err := loadClient()
			if err != nil {
				return err
			}
			documents, err := figma.FetchNodeDocuments(client, input.FileID, nodeIDs)
			if err != nil {
				return err
			}
			scope := output.Scope{FileKey: input.FileID, NodeIDs: nodeIDs}
			if raw {
				results := extract.ExtractRawComponentsFromDocuments(documents)
				results = extract.FilterRawComponentsByKind(results, kind)
				if nameFilter != "" {
					results = extract.FilterByName(results, nameFilter).([]map[string]any)
				}
				return cli.NewPrinter(cmd).JSON(output.NewQuery(scope, results))
			}

			results := extract.ExtractComponentsFromDocuments(documents)
			results = extract.FilterComponentsByKind(results, kind)
			if nameFilter != "" {
				results = extract.FilterByName(results, nameFilter).([]extract.ComponentOutput)
			}
			if usage {
				return cli.NewPrinter(cmd).JSON(output.NewQuery(scope, extract.AggregateComponentUsage(results)))
			}
			return cli.NewPrinter(cmd).JSON(output.NewQuery(scope, results))
		},
	}
	command.Flags().String("id", "", "node ID to inspect; defaults to URL node-id")
	command.Flags().String("name", "", "filter components by name (case-insensitive substring match)")
	command.Flags().String("kind", "", "filter by component, set, or instance")
	command.Flags().Bool("raw", false, "output raw Figma node JSON for jq power users")
	command.Flags().Bool("usage", false, "group component instances by exact component ID")
	return command
}

func init() {
	rootCmd.AddCommand(componentsCmd)
}
