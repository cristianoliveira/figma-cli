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
	Project   string       `json:"project"`
	Total     int          `json:"total"`
	Truncated bool         `json:"truncated,omitempty"`
	Hint      string       `json:"hint,omitempty"`
	Files     []fileOutput `json:"files"`
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
	return filesOutput{Project: response.Name, Total: len(files), Files: files}
}

var branches bool

var filesCmd = &cobra.Command{
	Use:   "files <project-url|project-id>",
	Short: "List files in a Figma project",
	Example: `  figma files <project-id>
  figma files https://www.figma.com/files/project/<project-id>/<name>`,
	Long: `List all files in a Figma project.

Pass a numeric project ID or a Figma project URL. Use --branches to request
branch data from Figma. Requires FIGMA_ACCESS_TOKEN with projects:read access.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resultLimit, err := readResultLimit(cmd)
		if err != nil {
			return err
		}
		projectID, err := figma.ParseProjectInput(args[0])
		if err != nil {
			return cli.NewUsageError(err)
		}
		client, err := cli.LoadClient()
		if err != nil {
			return err
		}
		client = client.WithContext(cmd.Context())
		response, err := figma.FetchProjectFiles(client, figma.BuildProjectFilesURL(projectID, branches))
		if err != nil {
			return err
		}
		result := newFilesOutput(response)
		result.Files, result.Total = limitResults(resultLimit, result.Files)
		result.Truncated = len(result.Files) < result.Total
		if result.Truncated {
			result.Hint = cli.FullHint(cmd, args)
		}
		return cli.NewPrinter(cmd).Structured(result)
	},
}

func init() {
	filesCmd.Flags().BoolVar(&branches, "branches", false, "include branch data")
	addResultLimitFlags(filesCmd)
	rootCmd.AddCommand(filesCmd)
}
