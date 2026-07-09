package cmd

import (
	"time"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/figma/api"
	"github.com/spf13/cobra"
)

type fileOutput struct {
	Key          string  `json:"key"`
	Name         string  `json:"name"`
	LastModified string  `json:"lastModified"`
	ThumbnailURL *string `json:"thumbnailUrl,omitempty"`
}

type filesOutput struct {
	Project string       `json:"project"`
	Files   []fileOutput `json:"files"`
}

func newFilesOutput(response api.GetProjectFilesResponse) filesOutput {
	files := make([]fileOutput, 0, len(response.Files))
	for _, file := range response.Files {
		files = append(files, fileOutput{
			Key:          file.Key,
			Name:         file.Name,
			LastModified: file.LastModified.Format(time.RFC3339),
			ThumbnailURL: file.ThumbnailUrl,
		})
	}
	return filesOutput{Project: response.Name, Files: files}
}

var branches bool

var filesCmd = &cobra.Command{
	Use:   "files <project-url|project-id>",
	Short: "List files in a Figma project",
	Long: `List all files in a Figma project.

Pass a numeric project ID or a Figma project URL. Use --branches to request
branch data from Figma. Requires FIGMA_ACCESS_TOKEN with projects:read access.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectID, err := figma.ParseProjectInput(args[0])
		if err != nil {
			return err
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}
		response, err := figma.FetchProjectFiles(client, figma.BuildProjectFilesURL(projectID, branches))
		if err != nil {
			return err
		}
		return cli.NewPrinter(cmd).JSON(newFilesOutput(response))
	},
}

func init() {
	filesCmd.Flags().BoolVar(&branches, "branches", false, "include branch data")
	rootCmd.AddCommand(filesCmd)
}
