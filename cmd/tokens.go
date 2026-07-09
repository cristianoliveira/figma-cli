package cmd

import (
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

var (
	tokensFormat       string
	tokensSource       string
	tokensMode         string
	tokensPrefix       string
	tokensOutput       string
	tokensTeamURL      string
	tokensScanFallback bool
)

var tokensCmd = &cobra.Command{
	Use:   "tokens [figma-url-or-file-id]",
	Short: "Extract design tokens (colors, type, spacing, shadows) as CSS, Tailwind, or JSON",
	Long: `Extract design tokens from a Figma file and emit ready-to-use CSS custom
properties, a Tailwind theme extend, or style-dictionary JSON.

Source resolution with --source auto (default): Variables first, then
published Styles (both semantically named). If neither exists, the command
falls back to scanning document nodes and naming tokens by value -- this is
on by default so raw-fill files still produce output. Pass --scan-fallback=false
for named tokens only. Pin --source in CI for deterministic output.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if tokensTeamURL != "" {
			return fmt.Errorf("--team is not supported yet; pass a file URL or file key")
		}
		input, err := figma.ParseInput(args[0])
		if err != nil {
			return err
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}

		tokens, err := collectTokens(client, input.FileID, input.NodeIDs, tokensSource, tokensMode, tokensScanFallback)
		if err != nil {
			return err
		}

		out, err := extract.FormatTokens(tokens, tokensFormat, tokensPrefix)
		if err != nil {
			return err
		}

		if tokensOutput != "" {
			if err := os.WriteFile(tokensOutput, []byte(out), 0o644); err != nil {
				return err
			}
			if err := cli.NewPrinter(cmd).File(tokensOutput, map[string]any{"format": tokensFormat, "bytes": len(out)}); err != nil {
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
	case "", "auto":
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

func init() {
	tokensCmd.Flags().StringVar(&tokensFormat, "format", "css", "output format: css, tailwind, or json")
	tokensCmd.Flags().StringVar(&tokensSource, "source", "auto", "token source: variables, styles, scan, or auto")
	tokensCmd.Flags().BoolVar(&tokensScanFallback, "scan-fallback", true, "in auto mode, fall back to a document node scan when no Styles/Variables exist (tokens named by value); use --scan-fallback=false for named-only")
	tokensCmd.Flags().StringVar(&tokensMode, "mode", "", "named mode to read (Variables only); default is the collection default")
	tokensCmd.Flags().StringVar(&tokensPrefix, "prefix", "", "CSS custom property prefix, e.g. \"fig-\"")
	tokensCmd.Flags().StringVar(&tokensOutput, "output", "", "write to file instead of stdout")
	tokensCmd.Flags().StringVar(&tokensTeamURL, "team", "", "pull from a team library (not yet supported)")
	rootCmd.AddCommand(tokensCmd)
}
