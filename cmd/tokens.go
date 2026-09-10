package cmd

import (
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/output"
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

// newTokensCommand constructs `figma tokens` using the explicit loadClient
// dependency. All flag state binds to per-instance variables so two roots
// built independently cannot leak defaults.
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
			client, err := deps.LoadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())

			nodeIDs := figma.ResolveNodeIDs(input, explicitNodeID)
			tokens, err := collectTokens(client, input.FileID, nodeIDs, options.source, options.mode, options.scanFallback)
			if err != nil {
				return err
			}

			out, err := extract.FormatTokens(tokens, options.format, options.prefix)
			if err != nil {
				return err
			}

			if options.output != "" {
				if err := output.WriteFile(options.output, []byte(out), 0o644); err != nil {
					return err
				}
				if err := cli.NewPrinter(cmd).File(options.output, map[string]any{"format": options.format, "bytes": len(out)}); err != nil {
					return err
				}
				return nil
			}
			if err := cli.NewPrinter(cmd).Text("tokens", out); err != nil {
				return err
			}
			return nil
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

func validateTokensOptions(options tokensOptions) error {
	if options.format != "css" && options.format != "tailwind" && options.format != "json" {
		return fmt.Errorf("unknown format %q (want css, tailwind, or json)", options.format)
	}
	if options.source != "" && options.source != tokenSourceAuto && options.source != "variables" && options.source != "styles" && options.source != "scan" {
		return fmt.Errorf("unknown --source %q (want variables, styles, scan, or auto)", options.source)
	}
	return nil
}

// collectTokens resolves tokens for a file according to the requested source.
// "auto" tries Variables, then Styles. The document scan only runs when
// scanFallback is true (opt-in), so output stays deterministic per source.
func collectTokens(client *figma.Client, fileID string, nodeIDs []string, source, mode string, scanFallback bool) ([]extract.Token, error) {
	switch source {
	case "variables":
		return tokensFromVariables(client, fileID, mode)
	case "styles":
		return tokensFromStyles(client, fileID)
	case "scan":
		return tokensFromScan(client, fileID, nodeIDs)
	case "", tokenSourceAuto:
		if tokens, err := tokensFromVariables(client, fileID, mode); err == nil && len(tokens) > 0 {
			return tokens, nil
		}
		if tokens, err := tokensFromStyles(client, fileID); err == nil && len(tokens) > 0 {
			return tokens, nil
		}
		if scanFallback {
			fmt.Fprintln(os.Stderr, "note: no Styles/Variables found; scanning document nodes (tokens named by value). Use --scan-fallback=false for named-only.")
			return tokensFromScan(client, fileID, nodeIDs)
		}
		fmt.Fprintln(os.Stderr, "note: no Styles/Variables found and scan disabled; drop --scan-fallback=false (or use --source scan) to extract from raw fills")
		return nil, nil
	}
	return nil, fmt.Errorf("unknown --source %q (want variables, styles, scan, or auto)", source)
}

func tokensFromVariables(client *figma.Client, fileID, mode string) ([]extract.Token, error) {
	meta, err := figma.FetchVariables(client, fileID)
	if err != nil {
		return nil, err
	}
	return extract.ExtractTokensFromVariablesE(meta, mode)
}

func tokensFromScan(client *figma.Client, fileID string, nodeIDs []string) ([]extract.Token, error) {
	doc, err := figma.FetchDocument(client, fileID, nodeIDs, "", "")
	if err != nil {
		return nil, err
	}
	return extract.ExtractTokensFromDocument(doc), nil
}

func tokensFromStyles(client *figma.Client, fileID string) ([]extract.Token, error) {
	styles, err := figma.FetchStyles(client, fileID)
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	nodeIDs := make([]string, 0, len(styles))
	for _, s := range styles {
		id, _ := s["node_id"].(string)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		nodeIDs = append(nodeIDs, id)
	}
	if len(nodeIDs) == 0 {
		return extract.ExtractTokensFromStyles(styles, nil), nil
	}
	nodes, err := figma.FetchNodes(client, fileID, nodeIDs)
	if err != nil {
		return nil, err
	}
	return extract.ExtractTokensFromStyles(styles, nodes), nil
}
