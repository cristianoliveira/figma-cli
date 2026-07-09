package cmd

import (
	"fmt"
	"os"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/spf13/cobra"
)

type projectOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type projectsOutput struct {
	Team     string          `json:"team"`
	Projects []projectOutput `json:"projects"`
}

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

func newProjectsOutput(response api.GetTeamProjectsResponse) projectsOutput {
	projects := make([]projectOutput, 0, len(response.Projects))
	for _, project := range response.Projects {
		projects = append(projects, projectOutput{ID: project.Id, Name: project.Name})
	}
	return projectsOutput{Team: response.Name, Projects: projects}
}

var projectsCmd = &cobra.Command{
	Use:   "projects [team-url|team-id]",
	Short: "List projects in a Figma team",
	Long: `List all projects in a Figma team.

The team ID cannot be obtained from a Figma token. Pass a numeric team ID or
a Figma team URL, or set FIGMA_TEAM_ID. An explicit argument overrides the
environment default. Requires FIGMA_ACCESS_TOKEN with projects:read access.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		teamID, err := resolveTeamInput(args, os.Getenv)
		if err != nil {
			return err
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}
		response, err := figma.FetchTeamProjects(client, figma.BuildTeamProjectsURL(teamID))
		if err != nil {
			return err
		}
		return cli.NewPrinter(cmd).JSON(newProjectsOutput(response))
	},
}

func init() {
	rootCmd.AddCommand(projectsCmd)
}
