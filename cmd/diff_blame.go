package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/diff"
	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

type blameVersionOutput struct {
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`
	User      string `json:"user"`
}

type blameOutput struct {
	IntroducedIn    blameVersionOutput  `json:"introducedIn"`
	Previous        *blameVersionOutput `json:"previous,omitempty"`
	Changes         extract.TextOutput  `json:"changes"`
	PredatesHistory bool                `json:"predatesHistory"`
}

func newBlameOutput(r diff.BlameResult) blameOutput {
	out := blameOutput{
		IntroducedIn: blameVersionOutput{
			ID:        r.IntroducedIn.ID,
			CreatedAt: r.IntroducedIn.CreatedAt.Format(time.RFC3339),
			User:      r.IntroducedIn.User,
		},
		Changes:         r.Changes,
		PredatesHistory: r.PredatesHistory,
	}
	if r.Previous != nil {
		out.Previous = &blameVersionOutput{
			ID:        r.Previous.ID,
			CreatedAt: r.Previous.CreatedAt.Format(time.RFC3339),
			User:      r.Previous.User,
		}
	}
	return out
}

func formatBlame(out blameOutput) string {
	var b strings.Builder
	note := ""
	if out.PredatesHistory {
		note = "  (at or before the oldest searched version)"
	}
	intro := out.IntroducedIn
	fmt.Fprintf(&b, "Introduced in %s by %s on %s%s\n", intro.ID, intro.User, formatBlameDate(intro.CreatedAt), note)
	for _, c := range out.Changes.Changed {
		fmt.Fprintf(&b, "  %s%s %q: %q -> %q\n", c.ID, formatTextPath(c.Path), c.Name, c.From, c.To)
	}
	for _, a := range out.Changes.Added {
		fmt.Fprintf(&b, "  %s%s %q: added %q\n", a.ID, formatTextPath(a.Path), a.Name, a.Text)
	}
	for _, r := range out.Changes.Removed {
		fmt.Fprintf(&b, "  %s%s %q: removed %q\n", r.ID, formatTextPath(r.Path), r.Name, r.Text)
	}
	return b.String()
}

func formatTextPath(path []string) string {
	if len(path) == 0 {
		return ""
	}
	return " [" + strings.Join(path, "/") + "]"
}

func formatBlameDate(rfc3339 string) string {
	t, err := time.Parse(time.RFC3339, rfc3339)
	if err != nil {
		return rfc3339
	}
	return t.Format("2006-01-02 15:04:05")
}

var diffBlameCmd = &cobra.Command{
	Use:   "blame [file-id-or-url] --to version-id",
	Short: "Find the version that introduced the current text at a node",
	Long: `Find the version that introduced the current text at a node.

Searches the file's version history backward from --to, binary-searching for
the oldest version whose text matches --to at the given node. Prints the
introducing version (id, author, date) and the before/after text change.

The node is inferred from URL node-id or explicit --id/--node. --to is required;
--from optionally caps how far back to search (oldest version ID, default:
oldest available).

Fetching the full history may require several API calls. Use --json for a
machine-readable result.

Examples:
  figma diff blame https://www.figma.com/design/KEY/File?node-id=4707-15608 --to 2374505616843859677
  figma diff blame <url> --to <version-id> --from <older-version-id>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		toVersion, _ := cmd.Flags().GetString("to")
		fromVersion, _ := cmd.Flags().GetString("from")
		explicitNodeID, err := explicitNodeIDFlag(cmd)
		if err != nil {
			return err
		}
		if toVersion == "" {
			return fmt.Errorf("--to is required")
		}
		input, err := figma.ParseInput(args[0])
		if err != nil {
			return err
		}
		nodeID, err := figma.ResolveSingleNodeID(input, explicitNodeID, "diff blame")
		if err != nil {
			return err
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}
		client = client.WithContext(cmd.Context())
		result, err := diff.FindTextChange(figma.NewTextHistory(client, input.FileID, []string{nodeID}), toVersion, fromVersion)
		if err != nil {
			return err
		}
		out := newBlameOutput(result)
		return cli.NewPrinter(cmd).Render(out, formatBlame(out))
	},
}

func init() {
	addNodeIDFlag(diffBlameCmd, "node ID to inspect; defaults to URL node-id")
	diffBlameCmd.Flags().String("to", "", "target version ID whose text to explain (required)")
	diffBlameCmd.Flags().String("from", "", "oldest version ID to search back to (default: oldest available)")
	diffCmd.AddCommand(diffBlameCmd)
}
