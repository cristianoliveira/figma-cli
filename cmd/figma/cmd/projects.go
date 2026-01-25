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
	rootCmd.AddCommand(projectsCmd)
	projectsCmd.AddCommand(projectsFilesCmd)
}

// projectsCmd represents the projects command
var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage Figma projects",
	Long: `List projects and manage project resources.

Use 'figma projects files <project_id>' to list files within a project.`,
}

// projectsFilesCmd represents the projects files command
var projectsFilesCmd = &cobra.Command{
	Use:   "files <project_id>",
	Short: "List files within a project",
	Long: `List all files within a given Figma project.

Output includes file key, name, last modified timestamp, and version.
Supports JSON, YAML, and text output formats.`,
	Args: cobra.ExactArgs(1),
	RunE: runProjectsFiles,
}

// FileOutput represents a file in the output
type FileOutput struct {
	Key          string `json:"key"`
	Name         string `json:"name"`
	LastModified string `json:"last_modified"`
	Version      string `json:"version,omitempty"`
	ThumbnailURL string `json:"thumbnail_url,omitempty"`
}

// runProjectsFiles implements the projects files command.
func runProjectsFiles(cmd *cobra.Command, args []string) error {
	projectID := args[0]
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

	// Fetch project files
	logger.Debug(ctx, "Fetching project files", logging.String("project_id", projectID))
	files, err := client.GetProjectFiles(ctx, projectID)
	if err != nil {
		logger.Error(ctx, "Failed to fetch project files", logging.Err(err), logging.String("project_id", projectID))
		return fmt.Errorf("failed to fetch project files: %w", err)
	}
	logger.Debug(ctx, "Project files fetched", logging.Int("count", len(files)))

	// Convert to output format
	outputFiles := make([]FileOutput, len(files))
	for i, f := range files {
		outputFiles[i] = FileOutput{
			Key:          f.Key,
			Name:         f.Name,
			LastModified: f.LastModified,
			Version:      f.Version,
			ThumbnailURL: f.ThumbnailURL,
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
		data, err := json.MarshalIndent(outputFiles, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		cmd.Println(string(data))
	case config.OutputFormatYAML:
		data, err := yaml.Marshal(outputFiles)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		cmd.Println(string(data))
	default: // text
		if len(outputFiles) == 0 {
			cmd.Println("No files found.")
			return nil
		}
		cmd.Printf("Found %d files in project %s:\n\n", len(outputFiles), projectID)
		for _, f := range outputFiles {
			cmd.Printf("%s (%s)\n", f.Name, f.Key)
			cmd.Printf("  Last modified: %s\n", f.LastModified)
			if f.Version != "" {
				cmd.Printf("  Version: %s\n", f.Version)
			}
			if f.ThumbnailURL != "" {
				cmd.Printf("  Thumbnail: %s\n", f.ThumbnailURL)
			}
			cmd.Println()
		}
	}
	return nil
}
