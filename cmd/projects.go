package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
)

type projectOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type projectsOutput struct {
	Team      string          `json:"team"`
	Total     int             `json:"total"`
	Truncated bool            `json:"truncated,omitempty"`
	Hint      string          `json:"hint,omitempty"`
	Projects  []projectOutput `json:"projects"`
}

// resolveTeamInput is exposed for testing via the deps GetEnv. Production
// callers inject the real env reader; tests can substitute a stub.
func resolveTeamInput(args []string, getenv func(string) string) (string, error) {
	input := getenv("FIGMA_TEAM_ID")
	if len(args) == 1 {
		input = args[0]
	}
	if input == "" {
		return "", fmt.Errorf("team URL or ID is required: pass one or set FIGMA_TEAM_ID")
	}
	return figma.ParseTeamInput(input)
}

// newProjectsOutput maps the stable figma.TeamProjects DTO into the
// command-layer output shape.
func newProjectsOutput(response figma.TeamProjects) projectsOutput {
	projects := make([]projectOutput, 0, len(response.Projects))
	for _, project := range response.Projects {
		projects = append(projects, projectOutput{ID: project.ID, Name: project.Name})
	}
	return projectsOutput{Team: response.Name, Total: len(projects), Projects: projects}
}

// newProjectsCommand constructs the `figma projects` command using the
// explicit loadClient dependency. Flag state lives on the returned command
// instance so two independently built roots cannot share defaults.
func newProjectsCommand(deps Deps) *cobra.Command {
	command := &cobra.Command{
		Use:   "projects [team-url|team-id]",
		Short: "List projects in a Figma team",
		Example: `  figma projects <team-id>
  figma projects https://www.figma.com/files/team/<team-id>/<name>`,
		Long: `List all projects in a Figma team.

The team ID cannot be obtained from a Figma token. Pass a numeric team ID or
a Figma team URL, or set FIGMA_TEAM_ID. An explicit argument overrides the
environment default. Requires FIGMA_ACCESS_TOKEN with projects:read access.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resultLimit, err := readResultLimit(cmd)
			if err != nil {
				return err
			}
			teamID, err := resolveTeamInput(args, deps.GetEnv)
			if err != nil {
				return cli.NewUsageError(err)
			}
			client, err := deps.LoadClient()
			if err != nil {
				return err
			}
			client = client.WithContext(cmd.Context())
			response, err := figma.FetchTeamProjects(client, figma.BuildTeamProjectsURL(teamID))
			if err != nil {
				return err
			}
			result := newProjectsOutput(response)
			result.Projects, result.Total = limitResults(resultLimit, result.Projects)
			result.Truncated = len(result.Projects) < result.Total
			if result.Truncated {
				result.Hint = cli.FullHint(cmd, args)
			}
			return cli.NewPrinter(cmd).Structured(result)
		},
	}
	addResultLimitFlags(command)
	return command
}
