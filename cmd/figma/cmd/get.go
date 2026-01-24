package cmd

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/cristianoliveira/figma-cli/internal/api"
	"github.com/cristianoliveira/figma-cli/internal/config"
	"github.com/cristianoliveira/figma-cli/internal/logging"
	"github.com/cristianoliveira/figma-cli/pkg/figma"
	"github.com/spf13/cobra"
)

const (
	// outputFormatPretty is the pretty-printed JSON format.
	outputFormatPretty = "pretty"
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
	getCmd.Flags().String("output-format", outputFormatPretty, "Output format: json, pretty (pretty JSON), yaml")
	getCmd.Flags().StringP("output", "o", "", "Output format: json, pretty, yaml (overrides --output-format)")
	getCmd.Flags().Int("depth", 1, "Depth of children to fetch (1-10, default 1)")
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

	// Get depth flag
	depth, _ := cmd.Flags().GetInt("depth")
	if depth < 1 || depth > 10 {
		return fmt.Errorf("depth must be between 1 and 10, got %d", depth)
	}
	logger.Debug(ctx, "Depth setting", logging.Int("depth", depth))

	// Determine output format
	outputFormat, _ := cmd.Flags().GetString("output")
	if outputFormat == "" {
		outputFormat, _ = cmd.Flags().GetString("output-format")
	}
	if outputFormat == "" {
		// Fallback to global output flag
		outputFormat = cfg.OutputFormat
	}
	// Normalize output format
	switch outputFormat {
	case config.OutputFormatJSON:
		// keep as is
	case outputFormatPretty:
		// keep as is
	case config.OutputFormatText:
		// treat as pretty JSON
		outputFormat = outputFormatPretty
	case config.OutputFormatYAML:
		// keep as is
	default:
		return fmt.Errorf("unsupported output format: %s (supported: %s, %s, %s)", outputFormat, config.OutputFormatJSON, outputFormatPretty, config.OutputFormatYAML)
	}
	logger.Debug(ctx, "Output format", logging.String("format", outputFormat))

	// Fetch file or node based on presence of node ID
	var data interface{}
	if parsed.NodeID != "" {
		logger.Debug(ctx, "Fetching node", logging.String("file_key", parsed.FileKey), logging.String("node_id", parsed.NodeID), logging.Int("depth", depth))
		node, err := client.GetNode(ctx, parsed.FileKey, parsed.NodeID, api.WithDepthForNode(depth))
		if err != nil {
			logger.Error(ctx, "Failed to fetch node", logging.Err(err), logging.String("file_key", parsed.FileKey), logging.String("node_id", parsed.NodeID))
			return fmt.Errorf("failed to fetch node: %w", err)
		}
		logger.Debug(ctx, "Node fetched successfully", logging.String("node_name", node.Name), logging.String("node_type", node.Type))
		data = node
	} else {
		logger.Debug(ctx, "Fetching file", logging.String("file_key", parsed.FileKey))
		file, err := client.GetFile(ctx, parsed.FileKey)
		if err != nil {
			logger.Error(ctx, "Failed to fetch file", logging.Err(err), logging.String("file_key", parsed.FileKey))
			return fmt.Errorf("failed to fetch file: %w", err)
		}
		logger.Debug(ctx, "File fetched successfully", logging.String("file_name", file.Name), logging.String("last_modified", file.LastModified))
		data = file
	}

	// Output data
	return outputData(cmd, data, outputFormat)
}

func outputData(cmd *cobra.Command, data interface{}, format string) error {
	switch format {
	case config.OutputFormatJSON:
		bytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal to JSON: %w", err)
		}
		cmd.Println(string(bytes))
	case outputFormatPretty:
		bytes, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal to JSON: %w", err)
		}
		cmd.Println(string(bytes))
	case config.OutputFormatYAML:
		bytes, err := yaml.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal to YAML: %w", err)
		}
		cmd.Println(string(bytes))
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
	return nil
}
