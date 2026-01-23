package cmd

import (
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"

	"github.com/cristianoliveira/figma-cli/internal/api"
	"github.com/cristianoliveira/figma-cli/internal/logging"
	"github.com/cristianoliveira/figma-cli/pkg/figma"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <figma-url>",
	Short: "Download a Figma file",
	Long: `Download a Figma file by its URL and output its details.

The command parses the Figma URL to extract the file key, fetches the file
from the Figma API, and outputs the file metadata and structure in JSON format.

Examples:
  figma get https://www.figma.com/file/ABC123/My-Design
  figma get https://www.figma.com/design/ABC123/My-Design?node-id=1:2`,
	Args: cobra.ExactArgs(1),
	RunE: runGet,
}

func init() {
	// Local flags for get command
	getCmd.Flags().String("output-format", "pretty", "Output format: json, pretty (pretty JSON), yaml")
}

func runGet(cmd *cobra.Command, args []string) error {
	url := args[0]
	ctx := cmd.Context()

	// Get logger
	logger, err := GetLogger(cmd)
	if err != nil {
		// Logging is optional for command execution
		logger = logging.NewNopLogger()
	}

	// Parse URL
	logger.Debug(ctx, "Parsing Figma URL", logging.String("url", url))
	parsed, err := figma.ParseURL(url)
	if err != nil {
		logger.Error(ctx, "Failed to parse Figma URL", logging.Err(err))
		return fmt.Errorf("invalid Figma URL: %w", err)
	}
	logger.Debug(ctx, "Parsed URL", logging.String("file_key", parsed.FileKey), logging.String("node_id", parsed.NodeID))

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

	// Fetch file
	logger.Debug(ctx, "Fetching file", logging.String("file_key", parsed.FileKey))
	file, err := client.GetFile(ctx, parsed.FileKey)
	if err != nil {
		logger.Error(ctx, "Failed to fetch file", logging.Err(err), logging.String("file_key", parsed.FileKey))
		return fmt.Errorf("failed to fetch file: %w", err)
	}
	logger.Debug(ctx, "File fetched successfully", logging.String("file_name", file.Name), logging.String("last_modified", file.LastModified))

	// Determine output format
	outputFormat, _ := cmd.Flags().GetString("output-format")
	if outputFormat == "" {
		// Fallback to global output flag
		outputFormat = cfg.OutputFormat
	}
	// Normalize output format
	switch outputFormat {
	case "json":
		// keep as is
	case "pretty":
		// keep as is
	case "text":
		// treat as pretty JSON
		outputFormat = "pretty"
	case "yaml":
		// keep as is
	default:
		return fmt.Errorf("unsupported output format: %s (supported: json, pretty, yaml)", outputFormat)
	}
	logger.Debug(ctx, "Output format", logging.String("format", outputFormat))

	// Output file
	return outputFile(cmd, file, outputFormat)
}

func outputFile(cmd *cobra.Command, file *api.File, format string) error {
	switch format {
	case "json":
		data, err := json.Marshal(file)
		if err != nil {
			return fmt.Errorf("failed to marshal file to JSON: %w", err)
		}
		cmd.Println(string(data))
	case "pretty":
		data, err := json.MarshalIndent(file, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal file to JSON: %w", err)
		}
		cmd.Println(string(data))
	case "yaml":
		data, err := yaml.Marshal(file)
		if err != nil {
			return fmt.Errorf("failed to marshal file to YAML: %w", err)
		}
		cmd.Println(string(data))
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
	return nil
}
