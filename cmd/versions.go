package cmd

import (
	"fmt"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/spf13/cobra"
)

type versionUserOutput struct {
	Handle string `json:"handle"`
	ID     string `json:"id"`
}

type versionOutput struct {
	ID           string            `json:"id"`
	CreatedAt    string            `json:"createdAt"`
	Label        *string           `json:"label"`
	Description  *string           `json:"description"`
	Named        bool              `json:"named"`
	User         versionUserOutput `json:"user"`
	ThumbnailURL *string           `json:"thumbnailUrl,omitempty"`
}

// versionsPaginationOutput surfaces actionable pagination cursors. The Figma
// API returns opaque next_page/prev_page URLs; we translate those into the
// version ID a caller should pass back to --after / --before to page through
// history. Versions come back newest-first, so the next (older) page is reached
// via --after <oldest id in this page>, and the previous (newer) page via
// --before <newest id in this page>.
type versionsPaginationOutput struct {
	HasMore    bool   `json:"hasMore"`
	NextAfter  string `json:"nextAfter,omitempty"`
	HasPrev    bool   `json:"hasPrev"`
	PrevBefore string `json:"prevBefore,omitempty"`
}

type versionsOutput struct {
	Versions   []versionOutput          `json:"versions"`
	Pagination versionsPaginationOutput `json:"pagination"`
}

func newVersionsOutput(response api.GetFileVersionsResponse) versionsOutput {
	versions := make([]versionOutput, 0, len(response.Versions))
	for _, v := range response.Versions {
		versions = append(versions, versionOutput{
			ID:           v.Id,
			CreatedAt:    v.CreatedAt.Format(time.RFC3339),
			Label:        v.Label,
			Description:  v.Description,
			Named:        v.Label != nil || v.Description != nil,
			User:         versionUserOutput{Handle: v.User.Handle, ID: v.User.Id},
			ThumbnailURL: v.ThumbnailUrl,
		})
	}

	pagination := versionsPaginationOutput{
		HasMore: response.Pagination.NextPage != nil,
		HasPrev: response.Pagination.PrevPage != nil,
	}
	if len(versions) > 0 {
		if pagination.HasMore {
			pagination.NextAfter = versions[len(versions)-1].ID
		}
		if pagination.HasPrev {
			pagination.PrevBefore = versions[0].ID
		}
	}

	return versionsOutput{Versions: versions, Pagination: pagination}
}

// formatVersionsTable renders a human-readable view of version history:
// newest-first rows with author, named/auto type, and actionable pagination
// cursors. It mirrors the JSON shape from newVersionsOutput so the two stay
// in sync.
func formatVersionsTable(out versionsOutput) string {
	if len(out.Versions) == 0 {
		return "No versions found.\n"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Latest %d versions (newest first). Pass --after <id> for older pages.\n\n", len(out.Versions))

	tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, " #\tCreated (UTC)\tBy\tType\tVersion ID\tLabel")
	for i, v := range out.Versions {
		kind := "auto"
		if v.Named {
			kind = "named"
		}
		created, err := time.Parse(time.RFC3339, v.CreatedAt)
		if err != nil {
			created = time.Time{}
		}
		label := ""
		if v.Label != nil {
			label = *v.Label
		}
		_, _ = fmt.Fprintf(tw, "%2d\t%s\t%s\t%s\t%s\t%s\n",
			i+1, created.Format("2006-01-02 15:04:05"), v.User.Handle, kind, v.ID, label)
	}
	_ = tw.Flush()

	p := out.Pagination
	fmt.Fprintf(&b, "\nPagination: hasMore=%v", p.HasMore)
	if p.NextAfter != "" {
		fmt.Fprintf(&b, "  nextAfter=%s  (--after fetches older)", p.NextAfter)
	}
	fmt.Fprintf(&b, "  hasPrev=%v\n", p.HasPrev)
	return b.String()
}

var (
	versionsPageSize int
	versionsBefore   string
	versionsAfter    string
)

var versionsCmd = &cobra.Command{
	Use:   "versions [file-id-or-url]",
	Short: "Fetch version history for a Figma file",
	Long: `Fetch version history for a Figma file via the Figma API.

Versions are returned newest-first. Auto-save milestones have no label or
description; named versions surface both. Use the pagination flags to page
through history beyond the first page.

Requires FIGMA_ACCESS_TOKEN environment variable set with a personal access token.

Pagination:
  --page-size   Number of versions per page (1-50, API default 30).
  --before      Version ID: fetch versions newer than this one.
  --after       Version ID: fetch versions older than this one.

Each response includes a pagination block with nextAfter / prevBefore cursors;
pass nextAfter to --after to fetch the next (older) page.

Examples:
  figma versions grnVU2vAihHXwYgHryu2xE
  figma versions --page-size 50 https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=4-1082
  figma versions --after 2374505616843859677 grnVU2vAihHXwYgHryu2xE`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		input, err := figma.ParseInput(args[0])
		if err != nil {
			return err
		}
		apiURL, err := figma.BuildVersionsURL(input.FileID, figma.VersionsQuery{
			PageSize: versionsPageSize,
			Before:   versionsBefore,
			After:    versionsAfter,
		})
		if err != nil {
			return err
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}
		client = client.WithContext(cmd.Context())
		response, err := figma.FetchVersions(client, apiURL)
		if err != nil {
			return err
		}
		output := newVersionsOutput(response)
		return cli.NewPrinter(cmd).Render(output, formatVersionsTable(output))
	},
}

func init() {
	versionsCmd.Flags().IntVar(&versionsPageSize, "page-size", 0, "versions per page (1-50, API default 30)")
	versionsCmd.Flags().StringVar(&versionsBefore, "before", "", "version ID: fetch versions newer than this one")
	versionsCmd.Flags().StringVar(&versionsAfter, "after", "", "version ID: fetch versions older than this one")
	rootCmd.AddCommand(versionsCmd)
}
