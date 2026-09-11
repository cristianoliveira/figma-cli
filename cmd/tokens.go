package cmd

import (
	"context"
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/tokens"
	"github.com/spf13/cobra"
)

const tokenSourceAuto = "auto"

type tokensOptions struct {
	format       string
	source       string
	mode         string
	prefix       string
	output       string
	teamURL      string
	scanFallback bool
}

// TokenService is the consumer-owned port the `figma tokens` command
// drives. Production composition wires it against the Figma adapter;
// tests inject fakes.
type TokenService interface {
	Run(ctx context.Context, req tokens.Request) (tokens.Result, error)
}

// newTokensCommand constructs `figma tokens`. The command only validates
// flags, maps them into a request, and renders the result; source
// selection, fallback, diagnostics, formatting, and persistence live in
// the token application service.
func newTokensCommand(deps Deps) *cobra.Command {
	options := tokensOptions{}
	command := &cobra.Command{
		Use:   "tokens [figma-url-or-file-id]",
		Short: "Extract design tokens (colors, type, spacing, shadows) as CSS, Tailwind, or JSON",
		Example: `  figma tokens <file-key>
  figma tokens --format json --source styles <url>
  figma tokens --format css --prefix fig- --output tokens.css <url>`,
		Long: `Extract design tokens from a Figma file and emit ready-to-use CSS custom
properties, a Tailwind theme extend, or style-dictionary JSON.

Source resolution with --source auto (default): Variables first, then
published Styles (both semantically named). If neither exists, the command
falls back to scanning document nodes and naming tokens by value -- this is
on by default so raw-fill files still produce output. Pass --scan-fallback=false
for named tokens only. Pin --source in CI for deterministic output.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := validateTokensOptions(options); err != nil {
				return cli.NewUsageError(err)
			}
			if options.teamURL != "" {
				return cli.NewUsageError(fmt.Errorf("--team is not supported yet; pass a file URL or file key"))
			}
			explicitNodeID, err := explicitNodeIDFlag(cmd)
			if err != nil {
				return cli.NewUsageError(err)
			}
			input, err := figma.ParseInput(args[0])
			if err != nil {
				return cli.NewUsageError(err)
			}

			service, err := deps.TokenService()
			if err != nil {
				return err
			}

			result, err := service.Run(cmd.Context(), tokens.Request{
				FileID:       input.FileID,
				NodeIDs:      figma.ResolveNodeIDs(input, explicitNodeID),
				Source:       options.source,
				Mode:         options.mode,
				Format:       options.format,
				Prefix:       options.prefix,
				ScanFallback: options.scanFallback,
				OutputPath:   options.output,
			})
			if err != nil {
				return err
			}

			return renderTokensResult(cmd, result, options.format)
		},
	}
	addNodeIDFlag(command, "node ID to scan; defaults to URL node-id")
	command.Flags().StringVar(&options.format, "format", "css", "output format: css, tailwind, or json")
	command.Flags().StringVar(&options.source, "source", tokenSourceAuto, "token source: variables, styles, scan, or auto")
	command.Flags().BoolVar(&options.scanFallback, "scan-fallback", true, "in auto mode, fall back to a document node scan when no Styles/Variables exist (tokens named by value); use --scan-fallback=false for named-only")
	command.Flags().StringVar(&options.mode, "mode", "", "named mode to read (Variables only); default is the collection default")
	command.Flags().StringVar(&options.prefix, "prefix", "", "CSS custom property prefix, e.g. \"fig-\"")
	command.Flags().StringVar(&options.output, "output", "", "write to file instead of stdout")
	command.Flags().StringVar(&options.teamURL, "team", "", "pull from a team library (not yet supported)")
	return command
}

// renderTokensResult renders structured diagnostics to stderr and the
// formatted output to stdout or the artifact printer.
func renderTokensResult(cmd *cobra.Command, result tokens.Result, format string) error {
	for _, diag := range result.Diagnostics {
		switch diag.Severity {
		case "warning":
			cmd.PrintErrf("warning: %s\n", diag.Message)
		case "note":
			cmd.PrintErrf("note: %s\n", diag.Message)
		default:
			cmd.PrintErrf("%s\n", diag.Message)
		}
	}

	if result.OutputPath != "" {
		return cli.NewPrinter(cmd).File(result.OutputPath, map[string]any{"format": format, "bytes": result.Bytes})
	}
	return cli.NewPrinter(cmd).Text("tokens", result.Formatted)
}

func validateTokensOptions(options tokensOptions) error {
	if options.format != "css" && options.format != "tailwind" && options.format != "json" {
		return fmt.Errorf("unknown format %q (want css, tailwind, or json)", options.format)
	}
	if options.source != "" && options.source != tokenSourceAuto && options.source != "variables" && options.source != "styles" && options.source != "scan" {
		return fmt.Errorf("unknown --source %q (want variables, styles, scan, or auto)", options.source)
	}
	return nil
}
