package cmd

import (
	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

// meOutput is the curated, user-facing identity shape. It keeps casing
// consistent with the rest of the CLI (camelCase) and shields the command's
// output contract from api.gen.go field renames such as img_url.
type meOutput struct {
	ID       string `json:"id"`
	Handle   string `json:"handle"`
	Email    string `json:"email"`
	ImageURL string `json:"imageUrl"`
}

// newMeCommand constructs the `figma me` command. It receives the explicit
// Deps so the loadClient invocation flows through the composition root;
// tests can inject a fake loader without touching cli.LoadClient.
func newMeCommand(deps Deps) *cobra.Command {
	return &cobra.Command{
		Use:     "me",
		Short:   "Show the authenticated user (also validates FIGMA_ACCESS_TOKEN)",
		Example: "  figma me",
		Long: `Call /v1/me to return the currently authenticated user.

Doubles as a token-validity check: a successful response means
FIGMA_ACCESS_TOKEN is configured and working.

Requires FIGMA_ACCESS_TOKEN environment variable.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			client, err := deps.LoadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			me, err := figma.FetchMe(client, figma.BuildMeURL())
			if err != nil {
				return err
			}
			return cli.NewPrinter(cmd).Structured(meOutput{
				ID:       me.Id,
				Handle:   me.Handle,
				Email:    me.Email,
				ImageURL: me.ImgUrl,
			})
		},
	}
}
