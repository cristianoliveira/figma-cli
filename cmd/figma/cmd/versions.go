package cmd

import (
	"encoding/json"
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"

	"github.com/cristianoliveira/figma-cli/internal/api"
	"github.com/cristianoliveira/figma-cli/internal/config"
	"github.com/cristianoliveira/figma-cli/internal/logging"
	"github.com/cristianoliveira/figma-cli/pkg/figma"
	"github.com/spf13/cobra"
)

// versionsCmd represents the versions command
var versionsCmd = &cobra.Command{
	Use:   "versions <file-key>",
	Short: "List version history of a Figma file",
	Long: `List version history of a Figma file with pagination support.

The command retrieves the version history of a file, showing version labels,
timestamps, and user information.

Examples:
  figma versions ABC123
  figma versions ABC123 --page-size 10 --after "cursor"
  figma versions ABC123 --output-format json
  figma versions ABC123 --branch "feature/branch"`,
	Args: cobra.ExactArgs(1),
	RunE: runVersions,
}

func init() {
	// Add flags to versions command
	versionsCmd.Flags().Int("page-size", 0, "Number of versions per page (default: API default)")
	versionsCmd.Flags().String("before", "", "Cursor for pagination: get versions before this cursor")
	versionsCmd.Flags().String("after", "", "Cursor for pagination: get versions after this cursor")
	versionsCmd.Flags().String("branch", "", "Branch name for branch-specific versions")
	versionsCmd.Flags().String("output-format", outputFormatPretty, "Output format: json, pretty (pretty JSON), yaml")
	versionsCmd.Flags().StringP("output", "o", "", "Output format: json, pretty, yaml (overrides --output-format)")
}

// fileKeyRegex validates Figma file key format (alphanumeric)
var fileKeyRegex = regexp.MustCompile(`^[A-Za-z0-9]+$`)

func runVersions(cmd *cobra.Command, args []string) error {
	arg := args[0]
	ctx := cmd.Context()

	// Get logger
	logger, err := GetLogger(cmd)
	if err != nil {
		// Logging is optional for command execution
		logger = logging.NewNopLogger()
	}

	// Try to parse as Figma URL; if successful, extract file key.
	// Otherwise treat as raw file key.
	var fileKey string
	parsed, err := figma.ParseURL(arg)
	if err == nil {
		logger.Debug(ctx, "Parsed Figma URL", logging.String("file_key", parsed.FileKey), logging.String("node_id", parsed.NodeID))
		fileKey = parsed.FileKey
	} else {
		// Not a valid Figma URL; assume it's a raw file key.
		// Validate file key format (alphanumeric).
		if !fileKeyRegex.MatchString(arg) {
			return fmt.Errorf("invalid file key format: must be alphanumeric (or provide a valid Figma URL)")
		}
		fileKey = arg
		logger.Debug(ctx, "Using raw file key", logging.String("file_key", fileKey))
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

	// Get pagination flags
	pageSize, _ := cmd.Flags().GetInt("page-size")
	before, _ := cmd.Flags().GetString("before")
	after, _ := cmd.Flags().GetString("after")
	branch, _ := cmd.Flags().GetString("branch")

	// Build options
	var opts []api.GetFileVersionsOption
	if pageSize > 0 {
		opts = append(opts, api.WithPageSize(pageSize))
	}
	if before != "" {
		opts = append(opts, api.WithBefore(before))
	}
	if after != "" {
		opts = append(opts, api.WithAfter(after))
	}
	if branch != "" {
		opts = append(opts, api.WithBranchForVersions(branch))
	}

	// Fetch versions
	logger.Debug(ctx, "Fetching file versions", logging.String("file_key", fileKey), logging.Int("page_size", pageSize))
	versions, err := client.GetFileVersions(ctx, fileKey, opts...)
	if err != nil {
		logger.Error(ctx, "Failed to fetch file versions", logging.Err(err), logging.String("file_key", fileKey))
		return fmt.Errorf("failed to fetch file versions: %w", err)
	}
	logger.Debug(ctx, "Versions fetched successfully", logging.Int("count", len(versions)))

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

	// Output data
	return outputVersions(cmd, versions, outputFormat)
}

func outputVersions(cmd *cobra.Command, versions []*api.Version, format string) error {
	switch format {
	case config.OutputFormatJSON:
		bytes, err := json.Marshal(versions)
		if err != nil {
			return fmt.Errorf("failed to marshal to JSON: %w", err)
		}
		cmd.Println(string(bytes))
	case outputFormatPretty:
		bytes, err := json.MarshalIndent(versions, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal to JSON: %w", err)
		}
		cmd.Println(string(bytes))
	case config.OutputFormatYAML:
		bytes, err := yaml.Marshal(versions)
		if err != nil {
			return fmt.Errorf("failed to marshal to YAML: %w", err)
		}
		cmd.Println(string(bytes))
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
	return nil
}
