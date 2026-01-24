package cmd

import (
	"fmt"
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/api"
	"github.com/cristianoliveira/figma-cli/internal/config"
	"github.com/cristianoliveira/figma-cli/internal/logging"
	"github.com/cristianoliveira/figma-cli/pkg/figma"
	"github.com/spf13/cobra"
)

// propsCmd represents the props command
var propsCmd = &cobra.Command{
	Use:   "props <figma-url>",
	Short: "Display node properties",
	Long: `Display node properties including position, size, constraints, rotation, and layout settings.

The command parses a Figma URL to extract the file key and optional node ID,
fetches the node (or file) from the Figma API, and outputs a structured
representation of its properties.

Examples:
  figma props https://www.figma.com/file/ABC123/My-Design
  figma props https://www.figma.com/design/ABC123/My-Design?node-id=1:2

Property groups can be filtered using the --fields flag.`,
	Args: cobra.ExactArgs(1),
	RunE: runProps,
}

func init() {
	propsCmd.Flags().String("fields", "", "Comma-separated list of property groups to show (e.g., \"geometry,constraints,layout\")")
	propsCmd.Flags().String("output-format", "", "Output format: json, pretty (pretty JSON), yaml, text (default: text)")
}

func runProps(cmd *cobra.Command, args []string) error {
	url := args[0]
	ctx := cmd.Context()

	// Get logger
	logger, err := GetLogger(cmd)
	if err != nil {
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

	// Determine output format
	outputFormat, _ := cmd.Flags().GetString("output-format")
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
		// treat as pretty JSON? We'll use text format (key-value)
		outputFormat = config.OutputFormatText
	case config.OutputFormatYAML:
		// keep as is
	default:
		// Default to text if not specified
		if outputFormat == "" {
			outputFormat = config.OutputFormatText
		} else {
			return fmt.Errorf("unsupported output format: %s (supported: %s, %s, %s, %s)", outputFormat, config.OutputFormatJSON, outputFormatPretty, config.OutputFormatYAML, config.OutputFormatText)
		}
	}
	logger.Debug(ctx, "Output format", logging.String("format", outputFormat))

	// Fetch file or node based on presence of node ID
	var node *api.Node
	if parsed.NodeID != "" {
		logger.Debug(ctx, "Fetching node", logging.String("file_key", parsed.FileKey), logging.String("node_id", parsed.NodeID))
		node, err = client.GetNode(ctx, parsed.FileKey, parsed.NodeID)
		if err != nil {
			logger.Error(ctx, "Failed to fetch node", logging.Err(err), logging.String("file_key", parsed.FileKey), logging.String("node_id", parsed.NodeID))
			return fmt.Errorf("failed to fetch node: %w", err)
		}
		logger.Debug(ctx, "Node fetched successfully", logging.String("node_name", node.Name), logging.String("node_type", node.Type))
	} else {
		logger.Debug(ctx, "Fetching file", logging.String("file_key", parsed.FileKey))
		file, err := client.GetFile(ctx, parsed.FileKey)
		if err != nil {
			logger.Error(ctx, "Failed to fetch file", logging.Err(err), logging.String("file_key", parsed.FileKey))
			return fmt.Errorf("failed to fetch file: %w", err)
		}
		logger.Debug(ctx, "File fetched successfully", logging.String("file_name", file.Name), logging.String("last_modified", file.LastModified))
		node = file.Document
		if node == nil {
			return fmt.Errorf("file has no document node")
		}
	}

	// Extract properties
	props := nodeToPropertyMap(node)

	// Filter fields if requested
	fieldsFlag, _ := cmd.Flags().GetString("fields")
	if fieldsFlag != "" {
		props = filterProperties(props, fieldsFlag)
	}

	// Output data
	return outputProps(cmd, props, outputFormat)
}

// nodeToPropertyMap converts a node to a map of properties.
func nodeToPropertyMap(node *api.Node) map[string]interface{} {
	props := make(map[string]interface{})

	// Basic properties
	props["id"] = node.ID
	props["name"] = node.Name
	props["type"] = node.Type
	props["visible"] = node.Visible

	// Geometry
	if node.AbsoluteBoundingBox != nil {
		rect := node.AbsoluteBoundingBox
		props["x"] = rect.X
		props["y"] = rect.Y
		props["width"] = rect.Width
		props["height"] = rect.Height
	}

	// Transform
	if node.RelativeTransform != nil && len(node.RelativeTransform) > 0 {
		props["relativeTransform"] = node.RelativeTransform
	}

	// Rotation
	if node.Rotation != 0 {
		props["rotation"] = node.Rotation
	}

	// Constraints
	if node.Constraints != nil {
		constraints := make(map[string]interface{})
		constraints["horizontal"] = node.Constraints.Horizontal
		constraints["vertical"] = node.Constraints.Vertical
		props["constraints"] = constraints
	}
	props["isFixed"] = node.IsFixed

	// Layout properties
	if node.LayoutMode != "" {
		layout := make(map[string]interface{})
		layout["layoutMode"] = node.LayoutMode
		layout["primaryAxisAlignItems"] = node.PrimaryAxisAlignItems
		layout["counterAxisAlignItems"] = node.CounterAxisAlignItems
		layout["itemSpacing"] = node.ItemSpacing
		layout["paddingLeft"] = node.PaddingLeft
		layout["paddingRight"] = node.PaddingRight
		layout["paddingTop"] = node.PaddingTop
		layout["paddingBottom"] = node.PaddingBottom
		// Only include non-zero values
		if node.ItemSpacing != 0 || node.PaddingLeft != 0 || node.PaddingRight != 0 || node.PaddingTop != 0 || node.PaddingBottom != 0 {
			props["layout"] = layout
		} else if node.LayoutMode != "" {
			props["layout"] = layout
		}
	}

	// Additional properties that may be useful
	if node.Opacity != 0 && node.Opacity != 1 {
		props["opacity"] = node.Opacity
	}
	if node.BlendMode != "" {
		props["blendMode"] = node.BlendMode
	}
	if node.IsMask {
		props["isMask"] = node.IsMask
	}
	if node.Locked {
		props["locked"] = node.Locked
	}
	if node.StrokeWeight != 0 {
		props["strokeWeight"] = node.StrokeWeight
	}
	if node.CornerRadius != 0 {
		props["cornerRadius"] = node.CornerRadius
	}

	return props
}

// filterProperties filters the property map based on comma-separated field groups.
func filterProperties(props map[string]interface{}, fields string) map[string]interface{} {
	allowed := make(map[string]bool)
	for _, f := range strings.Split(fields, ",") {
		allowed[strings.TrimSpace(f)] = true
	}

	// Define field groups
	groups := map[string][]string{
		"geometry":    {"x", "y", "width", "height", "rotation", "relativeTransform"},
		"constraints": {"constraints", "isFixed"},
		"layout":      {"layout", "layoutMode", "primaryAxisAlignItems", "counterAxisAlignItems", "itemSpacing", "paddingLeft", "paddingRight", "paddingTop", "paddingBottom"},
		"basic":       {"id", "name", "type", "visible"},
		"style":       {"opacity", "blendMode", "strokeWeight", "cornerRadius"},
	}

	filtered := make(map[string]interface{})
	for key, value := range props {
		// Check if key is directly allowed
		if allowed[key] {
			filtered[key] = value
			continue
		}
		// Check if key belongs to an allowed group
		for group, fields := range groups {
			if allowed[group] {
				for _, field := range fields {
					if field == key {
						filtered[key] = value
						break
					}
				}
			}
		}
	}
	return filtered
}

// outputProps outputs the properties in the requested format.
func outputProps(cmd *cobra.Command, props map[string]interface{}, format string) error {
	switch format {
	case config.OutputFormatJSON:
		return outputData(cmd, props, config.OutputFormatJSON)
	case outputFormatPretty:
		return outputData(cmd, props, outputFormatPretty)
	case config.OutputFormatYAML:
		return outputData(cmd, props, config.OutputFormatYAML)
	case config.OutputFormatText:
		// Human-readable key-value list
		for key, value := range props {
			cmd.Printf("%s: %v\n", key, value)
		}
		return nil
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}
