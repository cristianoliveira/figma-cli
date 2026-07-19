package cmd

import (
	"fmt"
	"strconv"

	"github.com/cristianoliveira/figma-cli/internal/annotations"
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

var inspectCmd = newInspectCommand(cli.LoadClient)

const (
	inspectFormatJSON = "json"
	inspectFormatText = "text"
)

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
		Long:  "Show a curated summary of a specific Figma node. With --recursive, bounds stay absolute, relativeBounds are measured from the requested scope node, and spacingFromPrevious reports computed auto-layout sibling gaps. Use --format text with --fields to render a compact, selected implementation outline.",
		Example: `  figma inspect "<url>?node-id=42-1"
  figma inspect --id 42:1 <file-key>
  figma inspect --recursive --format text --fields id,name,type <url>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			explicitNodeID, err := explicitNodeIDFlag(cmd)
			if err != nil {
				return err
			}
			recursive, _ := cmd.Flags().GetBool("recursive")
			handoff, _ := cmd.Flags().GetBool("handoff")
			annotationsOutput, _ := cmd.Flags().GetString("annotations-output")
			includeVectorPaths, _ := cmd.Flags().GetBool("include-vector-paths")
			depth, _ := cmd.Flags().GetInt("depth")
			includeHidden, _ := cmd.Flags().GetBool("include-hidden")
			inspectFormat, _ := cmd.Flags().GetString("format")
			fields, _ := cmd.Flags().GetStringSlice("fields")
			if inspectFormat != inspectFormatJSON && inspectFormat != inspectFormatText {
				return fmt.Errorf("unknown inspect format %q (want json or text)", inspectFormat)
			}
			if inspectFormat == inspectFormatText && !recursive {
				return fmt.Errorf("--format text requires --recursive")
			}
			if cmd.Flags().Changed("fields") && inspectFormat != inspectFormatText {
				return fmt.Errorf("--fields requires --format text")
			}
			if err := extract.ValidateInspectFields(fields); err != nil {
				return err
			}
			if handoff && recursive {
				return fmt.Errorf("--handoff and --recursive cannot be used together")
			}
			if annotationsOutput != "" && !recursive {
				return fmt.Errorf("--annotations-output requires --recursive")
			}
			if includeVectorPaths && recursive && !cmd.Flags().Changed("depth") {
				return fmt.Errorf("--include-vector-paths with --recursive requires explicit --depth")
			}
			if depth < 0 {
				return fmt.Errorf("--depth must be zero or greater")
			}
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
			client = client.WithContext(cmd.Context())
			var details figma.NodeDetails
			if includeVectorPaths {
				geometryDepth := "1"
				if recursive {
					geometryDepth = strconv.Itoa(depth)
				}
				details, err = figma.FetchNodeDetailsWithVectorPaths(client, input.FileID, []string{nodeID}, geometryDepth)
			} else {
				details, err = figma.FetchNodeDetails(client, input.FileID, []string{nodeID})
			}
			if err != nil {
				return err
			}
			document, ok := details.Documents[0].(map[string]any)
			if !ok {
				return fmt.Errorf("node %s has an invalid document", nodeID)
			}
			scope := output.Scope{FileKey: input.FileID, NodeIDs: []string{nodeID}}
			if recursive {
				nodes := extract.InspectTreeRelativeToScope(document, nodeID)
				annotationDepth := -1
				if cmd.Flags().Changed("depth") {
					nodes = extract.InspectTreeRelativeToScopeToDepth(document, nodeID, depth)
					annotationDepth = depth
				}
				if annotationsOutput != "" {
					documentAnnotations, annotationsErr := extract.ExtractCoordinateAnnotations(document, nodeID, annotationDepth, includeHidden)
					if annotationsErr != nil {
						return annotationsErr
					}
					if err := annotations.Write(annotationsOutput, documentAnnotations); err != nil {
						return err
					}
				}
				enrichInspectNodes(nodes, details.Styles, client, input.FileID, fetchVariables)
				if inspectFormat == inspectFormatText {
					text, formatErr := extract.FormatInspectText(nodes, fields)
					if formatErr != nil {
						return formatErr
					}
					return cli.NewPrinter(cmd).Text("inspect", text)
				}
				return cli.NewPrinter(cmd).JSON(output.NewQuery(scope, nodes))
			}
			if handoff {
				result := extract.ExtractHandoff(document, extract.HandoffOptions{MaxDepth: depth, IncludeHidden: includeHidden})
				enrichInspectNodes(result.Nodes, details.Styles, client, input.FileID, fetchVariables)
				return cli.NewPrinter(cmd).JSON(output.Detail[extract.HandoffOutput]{Scope: scope, Result: result})
			}
			node := extract.NodeToInspectOutput(document)
			node = enrichInspectNode(node, details.Styles, client, input.FileID, fetchVariables)
			result := output.Detail[extract.InspectOutput]{Scope: scope, Result: node}
			if err := cli.NewPrinter(cmd).JSON(result); err != nil {
				return err
			}
			return nil
		},
	}
	addNodeIDFlag(command, "node ID to inspect; defaults to URL node-id")
	command.Flags().Bool("recursive", false, "include implementation specs for all descendant nodes")
	command.Flags().Bool("handoff", false, "emit bounded implementation specs and component usage")
	command.Flags().Int("depth", 4, "maximum descendant depth for --handoff or --recursive; recursive stays unbounded unless set")
	command.Flags().Bool("include-hidden", false, "include invisible descendants in --handoff or --annotations-output")
	command.Flags().String("annotations-output", "", "write generic screenshot-relative coordinate annotations; requires --recursive")
	command.Flags().Bool("include-vector-paths", false, "request and include exact Figma fill/stroke geometry; recursive use requires explicit --depth")
	command.Flags().String("format", inspectFormatJSON, "output format for recursive inspection: json or text")
	command.Flags().StringSlice("fields", nil, "comma-separated inspect fields for --format text; nested paths are supported")
	return command
}

func enrichInspectNode(
	node extract.InspectOutput,
	styles map[string]map[string]any,
	client *figma.Client,
	fileID string,
	fetchVariables func(*figma.Client, string) (map[string]any, error),
) extract.InspectOutput {
	nodes := []extract.InspectOutput{node}
	enrichInspectNodes(nodes, styles, client, fileID, fetchVariables)
	return nodes[0]
}

func enrichInspectNodes(
	nodes []extract.InspectOutput,
	styles map[string]map[string]any,
	client *figma.Client,
	fileID string,
	fetchVariables func(*figma.Client, string) (map[string]any, error),
) {
	needsVariables := false
	for index := range nodes {
		extract.ResolveInspectStyleBindings(&nodes[index], styles)
		needsVariables = needsVariables || len(nodes[index].VariableBindings) > 0
	}
	if !needsVariables {
		return
	}
	variables, err := fetchVariables(client, fileID)
	if err != nil {
		return
	}
	for index := range nodes {
		extract.ResolveInspectVariableBindings(&nodes[index], variables)
	}
}

func inspectNodeID(input *figma.FileInput, explicitNodeID string) (string, error) {
	return figma.ResolveSingleNodeID(input, explicitNodeID, "inspect")
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}
