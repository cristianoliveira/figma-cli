package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/config"
	"github.com/cristianoliveira/figma-cli/internal/logging"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	rootCmd.AddCommand(teamsCmd)
	teamsCmd.AddCommand(teamsProjectsCmd)
}

// teamsCmd represents the teams command
var teamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "Manage Figma teams",
	Long: `List teams and manage team resources.

Use 'figma teams projects <team_id>' to list projects for a team.`,
}

// teamsProjectsCmd represents the teams projects command
var teamsProjectsCmd = &cobra.Command{
	Use:   "projects <team_id>",
	Short: "List projects for a team",
	Long: `List all projects for a given Figma team.

Output includes project ID, name, creation date, and last modified date.
Supports JSON, YAML, and text output formats.`,
	Args: cobra.ExactArgs(1),
	RunE: runTeamsProjects,
}

// ProjectOutput represents a project in the output
type ProjectOutput struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CreatedAt  string `json:"created_at"`
	ModifiedAt string `json:"modified_at"`
}

// runTeamsProjects implements the teams projects command.
func runTeamsProjects(cmd *cobra.Command, args []string) error {
	teamID := args[0]
	ctx := cmd.Context()

	logger, err := GetLogger(cmd)
	if err != nil {
		logger = logging.NewNopLogger()
	}

	// Get configuration
	cfg, err := GetConfig(cmd)
	if err != nil {
		logger.Error(ctx, "Failed to get config", logging.Err(err))
		return err
	}

	// Validate required fields (token)
	if err := cfg.ValidateRequired(); err != nil {
		logger.Error(ctx, "Configuration validation failed", logging.Err(err))
		return err
	}

	// Build API client
	logger.Debug(ctx, "Creating API client", logging.String("base_url", cfg.API.BaseURL), logging.String("timeout", cfg.API.Timeout.String()))
	client, err := newAPIClient(cfg, logger)
	if err != nil {
		logger.Error(ctx, "Failed to create API client", logging.Err(err))
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Fetch projects
	logger.Debug(ctx, "Fetching team projects", logging.String("team_id", teamID))
	projects, err := client.GetTeamProjects(ctx, teamID)
	if err != nil {
		logger.Error(ctx, "Failed to fetch team projects", logging.Err(err), logging.String("team_id", teamID))
		return fmt.Errorf("failed to fetch team projects: %w", err)
	}
	logger.Debug(ctx, "Team projects fetched", logging.Int("count", len(projects)))

	// Convert to output format
	outputProjects := make([]ProjectOutput, len(projects))
	for i, p := range projects {
		outputProjects[i] = ProjectOutput{
			ID:         p.ID,
			Name:       p.Name,
			CreatedAt:  p.CreatedAt,
			ModifiedAt: p.ModifiedAt,
		}
	}

	// Determine output format
	outputFormat := cfg.OutputFormat
	if outputFormat == "" {
		outputFormat = config.OutputFormatText
	}

	// Output results
	switch outputFormat {
	case config.OutputFormatJSON:
		data, err := json.MarshalIndent(outputProjects, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		cmd.Println(string(data))
	case config.OutputFormatYAML:
		data, err := yaml.Marshal(outputProjects)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		cmd.Println(string(data))
	default: // text
		if len(outputProjects) == 0 {
			cmd.Println("No projects found.")
			return nil
		}
		cmd.Printf("Found %d projects for team %s:\n\n", len(outputProjects), teamID)
		for _, p := range outputProjects {
			cmd.Printf("%s (%s)\n", p.Name, p.ID)
			cmd.Printf("  Created:  %s\n", p.CreatedAt)
			cmd.Printf("  Modified: %s\n\n", p.ModifiedAt)
		}
	}
	return nil
}
