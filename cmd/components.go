package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/components"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

var componentsCmd = newComponentsCommand(cli.LoadClient)

type componentsDiffOutput struct {
	FileKey    string `json:"fileKey"`
	FigmaCount int    `json:"figmaCount"`
	CodeCount  int    `json:"codeCount"`
	components.Comparison
}

func newComponentsCommand(loadClient func() (*figma.Client, error)) *cobra.Command {
	command := &cobra.Command{
		Use:   "components [figma-url-or-file-id]",
		Short: "List components, component sets, and instances as JSON",
		Example: `  figma components "<url>?node-id=42-1"
  figma components --id 42:1 --kind instance <file-key>
  figma components --id 42:1 --usage <file-key>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeID, err := explicitNodeIDFlag(cmd)
			if err != nil {
				return cli.NewUsageError(err)
			}
			nameFilter, _ := cmd.Flags().GetString("name")
			kind, _ := cmd.Flags().GetString("kind")
			raw, _ := cmd.Flags().GetBool("raw")
			usage, _ := cmd.Flags().GetBool("usage")
			diff, _ := cmd.Flags().GetBool("diff")
			codebase, _ := cmd.Flags().GetString("codebase")
			if diff && codebase == "" {
				return cli.NewUsageError(fmt.Errorf("--diff requires --codebase"))
			}
			if !diff && codebase != "" {
				return cli.NewUsageError(fmt.Errorf("--codebase requires --diff"))
			}
			if diff && (nodeID != "" || nameFilter != "" || kind != "" || raw || usage || cmd.Flags().Changed("limit") || cmd.Flags().Changed("full")) {
				return cli.NewUsageError(fmt.Errorf("--diff cannot be used with --id, --name, --kind, --raw, --usage, --limit, or --full"))
			}
			if usage && raw {
				return cli.NewUsageError(fmt.Errorf("--usage and --raw cannot be used together"))
			}
			if usage && kind != "" && kind != "instance" {
				return cli.NewUsageError(fmt.Errorf("--usage only supports --kind instance"))
			}
			if kind != "" && kind != "component" && kind != "set" && kind != "instance" {
				return cli.NewUsageError(fmt.Errorf("invalid component kind %q: expected component, set, or instance", kind))
			}
			resultLimit, err := readResultLimit(cmd)
			if err != nil {
				return err
			}
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return cli.NewUsageError(err)
			}
			if diff && len(input.NodeIDs) > 0 {
				return cli.NewUsageError(fmt.Errorf("--diff compares a whole Figma file; remove node-id from the URL"))
			}
			var nodeIDs []string
			var codeComponents []components.CodeComponent
			if diff {
				codeComponents, err = components.DiscoverCodeComponents(codebase)
				if err != nil {
					return err
				}
			} else {
				nodeIDs, err = figma.ResolveRequiredNodeIDs(input, nodeID, "components")
				if err != nil {
					return cli.NewUsageError(err)
				}
			}
			client, err := loadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			if diff {
				figmaComponents, err := figma.FetchPublishedComponents(client, input.FileID)
				if err != nil {
					return err
				}
				published := make([]components.FigmaComponent, 0, len(figmaComponents))
				for _, component := range figmaComponents {
					published = append(published, components.FigmaComponent{Name: component.Name, NodeID: component.NodeID})
				}
				comparison := components.Compare(published, codeComponents)
				return cli.NewPrinter(cmd).Structured(componentsDiffOutput{FileKey: input.FileID, FigmaCount: len(published), CodeCount: len(codeComponents), Comparison: comparison})
			}
			documents, err := figma.FetchNodeDocuments(client, input.FileID, nodeIDs)
			if err != nil {
				return err
			}
			scope := output.Scope{FileKey: input.FileID, NodeIDs: nodeIDs}
			query := map[string]any{"name": nameFilter, "kind": kind, "raw": raw, "usage": usage}
			if raw {
				results := extract.ExtractRawComponentsFromDocuments(documents)
				results = extract.FilterRawComponentsByKind(results, kind)
				if nameFilter != "" {
					results = extract.FilterByName(results, nameFilter).([]map[string]any)
				}
				results, total := limitResults(resultLimit, results)
				return cli.NewPrinter(cmd).Structured(output.NewLimitedQuery(scope, query, total, results))
			}

			results := extract.ExtractComponentsFromDocuments(documents)
			results = extract.FilterComponentsByKind(results, kind)
			if nameFilter != "" {
				results = extract.FilterByName(results, nameFilter).([]extract.ComponentOutput)
			}
			if usage {
				usageResults := extract.AggregateComponentUsage(results)
				usageResults, total := limitResults(resultLimit, usageResults)
				return cli.NewPrinter(cmd).Structured(output.NewLimitedQuery(scope, query, total, usageResults))
			}
			results, total := limitResults(resultLimit, results)
			return cli.NewPrinter(cmd).Structured(output.NewLimitedQuery(scope, query, total, results))
		},
	}
	addNodeIDFlag(command, "node ID to inspect; defaults to URL node-id")
	command.Flags().String("name", "", "filter components by name (case-insensitive substring match)")
	command.Flags().String("kind", "", "filter by component, set, or instance")
	command.Flags().Bool("raw", false, "output raw Figma node JSON for jq power users")
	command.Flags().Bool("usage", false, "group component instances by exact component ID")
	command.Flags().Bool("diff", false, "compare published Figma components with a frontend codebase")
	command.Flags().String("codebase", "", "frontend component source directory (required with --diff)")
	addResultLimitFlags(command)
	return command
}

func init() {
	rootCmd.AddCommand(componentsCmd)
}
