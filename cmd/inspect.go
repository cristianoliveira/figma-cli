package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/inspect"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

const (
	inspectFormatJSON = "json"
	inspectFormatText = "text"
)

// InspectService is the application port that `figma inspect` drives.
// The factory accepts it through Deps so production composition can wire
// it against the Figma adapter and tests can drive it with fakes.
//
// We declare it as an interface so the command depends on the contract
// shape, not on the concrete *inspect.Service. This keeps cmd/inspect.go
// testable against a fake without reaching for the real implementation.
type InspectService interface {
	Inspect(inspect.Request) (*inspect.Result, error)
}

// newInspectCommand constructs `figma inspect`. The command is now
// responsible only for flag/argument parsing and rendering; fetch,
// scope resolution, extraction, enrichment, and limiting live in the
// inspect application service.
func newInspectCommand(deps Deps) *cobra.Command {
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
				return cli.NewUsageError(err)
			}
			if err := validateInspectFlags(cmd); err != nil {
				return cli.NewUsageError(err)
			}
			recursive, _ := cmd.Flags().GetBool("recursive")
			handoff, _ := cmd.Flags().GetBool("handoff")
			includeVectorPaths, _ := cmd.Flags().GetBool("include-vector-paths")
			depth, _ := cmd.Flags().GetInt("depth")
			includeHidden, _ := cmd.Flags().GetBool("include-hidden")
			inspectFormat, _ := cmd.Flags().GetString("format")
			fields, _ := cmd.Flags().GetStringSlice("fields")
			resultLimit, err := readResultLimit(cmd)
			if err != nil {
				return err
			}
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return cli.NewUsageError(err)
			}
			nodeID, err := inspectNodeID(input, explicitNodeID)
			if err != nil {
				return cli.NewUsageError(err)
			}

			service, err := deps.InspectService()
			if err != nil {
				return err
			}

			result, err := service.Inspect(inspect.Request{
				Context:            cmd.Context(),
				FileID:             input.FileID,
				NodeID:             nodeID,
				Recursive:          recursive,
				Handoff:            handoff,
				IncludeHidden:      includeHidden,
				Depth:              depth,
				IncludeVectorPaths: includeVectorPaths,
				Format:             inspect.Format(inspectFormat),
				Fields:             fields,
				ResultLimit:        resultLimit.limit(),
			})
			if err != nil {
				return err
			}

			return renderInspectResult(cmd, result)
		},
	}
	addNodeIDFlag(command, "node ID to inspect; defaults to URL node-id")
	command.Flags().Bool("recursive", false, "include implementation specs for all descendant nodes")
	command.Flags().Bool("handoff", false, "emit bounded implementation specs and component usage")
	command.Flags().Int("depth", 4, "maximum descendant depth for --handoff or --recursive; recursive stays unbounded unless set")
	command.Flags().Bool("include-hidden", false, "include invisible descendants in --handoff")
	command.Flags().Bool("include-vector-paths", false, "request and include exact Figma fill/stroke geometry; recursive use requires explicit --depth")
	command.Flags().String("format", inspectFormatJSON, "output format for recursive inspection: json or text")
	command.Flags().StringSlice("fields", nil, "comma-separated inspect fields for --format text; nested paths are supported")
	addResultLimitFlags(command)
	return command
}

// validateInspectFlags enforces the pre-call contract documented in the
// command's Long text. It returns nil for valid combinations and a
// descriptive error otherwise.
func validateInspectFlags(cmd *cobra.Command) error {
	inspectFormat, _ := cmd.Flags().GetString("format")
	if inspectFormat != inspectFormatJSON && inspectFormat != inspectFormatText {
		return fmt.Errorf("unknown inspect format %q (want json or text)", inspectFormat)
	}
	recursive, _ := cmd.Flags().GetBool("recursive")
	handoff, _ := cmd.Flags().GetBool("handoff")
	if inspectFormat == inspectFormatText && !recursive {
		return fmt.Errorf("--format text requires --recursive")
	}
	if cmd.Flags().Changed("fields") && inspectFormat != inspectFormatText {
		return fmt.Errorf("--fields requires --format text")
	}
	fields, _ := cmd.Flags().GetStringSlice("fields")
	if err := extract.ValidateInspectFields(fields); err != nil {
		return err
	}
	if handoff && recursive {
		return fmt.Errorf("--handoff and --recursive cannot be used together")
	}
	includeVectorPaths, _ := cmd.Flags().GetBool("include-vector-paths")
	depth, _ := cmd.Flags().GetInt("depth")
	if includeVectorPaths && recursive && !cmd.Flags().Changed("depth") {
		return fmt.Errorf("--include-vector-paths with --recursive requires explicit --depth")
	}
	if depth < 0 {
		return fmt.Errorf("--depth must be zero or greater")
	}
	if !recursive && (cmd.Flags().Changed("limit") || cmd.Flags().Changed("full")) {
		return fmt.Errorf("--limit and --full require --recursive")
	}
	return nil
}

// renderInspectResult turns the application service's Result into the
// command-layer rendering contract. The two output shapes
// (output.Detail, output.Query) are the existing CLI output contracts;
// the service has no knowledge of either.
func renderInspectResult(cmd *cobra.Command, result *inspect.Result) error {
	switch result.Mode {
	case inspect.ModeText:
		// ModeText is not actually emitted by the service today; kept as
		// an explicit guard so adding a text-only Result mode is safe.
		return cli.NewPrinter(cmd).Text("inspect", result.Text)
	case inspect.ModeRecursive:
		if result.Text != "" {
			return cli.NewPrinter(cmd).Text("inspect", result.Text)
		}
		return cli.NewPrinter(cmd).Structured(newLimitedQuery(cmd, result.Scope, nil, result.Total, result.Nodes))
	case inspect.ModeHandoff:
		if result.Handoff == nil {
			return fmt.Errorf("inspect service: handoff result missing payload")
		}
		return cli.NewPrinter(cmd).Structured(output.Detail[extract.HandoffOutput]{Scope: result.Scope, Result: *result.Handoff})
	case inspect.ModeSingle:
		if result.Single == nil {
			return fmt.Errorf("inspect service: single result missing payload")
		}
		return cli.NewPrinter(cmd).Structured(output.Detail[extract.InspectOutput]{Scope: result.Scope, Result: *result.Single})
	}
	return fmt.Errorf("inspect service: unknown mode %q", result.Mode)
}

func inspectNodeID(input *figma.FileInput, explicitNodeID string) (string, error) {
	return figma.ResolveSingleNodeID(input, explicitNodeID, "inspect")
}
