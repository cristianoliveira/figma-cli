package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/api"
	"github.com/cristianoliveira/figma-cli/internal/config"
	"github.com/cristianoliveira/figma-cli/internal/logging"
	"github.com/cristianoliveira/figma-cli/pkg/figma"
	"github.com/spf13/cobra"
)

var _ = context.Background()

// stylesCmd represents the styles command
var stylesCmd = &cobra.Command{
	Use:   "styles <figma-url>",
	Short: "Get styles applied to a node",
	Long: `Get styles applied to a Figma node, including text styles, fill styles,
effect styles, and grid styles. Show style references and their definitions.
Supports both applied and referenced styles.`,
	Args: cobra.ExactArgs(1),
	RunE: runStyles,
}

func init() {
	// No flags for now
}

// StyleInfo represents a style applied to or referenced by a node.
type StyleInfo struct {
	Type       string                 `json:"type"`                 // "text", "fill", "effect", "grid"
	StyleKey   string                 `json:"style_key,omitempty"`  // reference key (if referenced)
	Name       string                 `json:"name,omitempty"`       // style name (if referenced)
	Applied    map[string]interface{} `json:"applied,omitempty"`    // applied style properties
	Referenced map[string]interface{} `json:"referenced,omitempty"` // referenced style definition (if resolved)
}

// NodeStyles represents all styles for a node.
type NodeStyles struct {
	NodeID   string      `json:"node_id"`
	NodeName string      `json:"node_name"`
	NodeType string      `json:"node_type"`
	Styles   []StyleInfo `json:"styles"`
}

// runStyles implements the styles command.
func runStyles(cmd *cobra.Command, args []string) error {
	url := args[0]
	ctx := cmd.Context()

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

	// Fetch node (or file root)
	root, err := fetchRoot(ctx, client, parsed, logger)
	if err != nil {
		return err
	}

	// Collect styles
	nodeStyles := collectStyles(root, parsed.FileKey)

	// Determine output format
	outputFormat := cfg.OutputFormat
	if outputFormat == "" {
		outputFormat = config.OutputFormatText
	}

	// Output results
	switch outputFormat {
	case config.OutputFormatJSON:
		data, err := json.MarshalIndent(nodeStyles, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		cmd.Println(string(data))
	case config.OutputFormatYAML:
		// TODO: implement YAML output if needed
		fallthrough
	default: // text
		// Simple text output
		if len(nodeStyles.Styles) == 0 {
			cmd.Println("No styles found.")
			return nil
		}
		cmd.Printf("Styles for node %s (%s):\n\n", nodeStyles.NodeName, nodeStyles.NodeID)
		for _, s := range nodeStyles.Styles {
			cmd.Printf("- Type: %s", s.Type)
			if s.StyleKey != "" {
				cmd.Printf(" (key: %s)", s.StyleKey)
			}
			cmd.Println()
			if len(s.Applied) > 0 {
				cmd.Printf("  Applied: %v\n", s.Applied)
			}
			if len(s.Referenced) > 0 {
				cmd.Printf("  Referenced: %v\n", s.Referenced)
			}
		}
	}
	return nil
}

// collectStyles extracts styles from a node.
func collectStyles(node *api.Node, fileKey string) *NodeStyles {
	styles := &NodeStyles{
		NodeID:   node.ID,
		NodeName: node.Name,
		NodeType: node.Type,
		Styles:   []StyleInfo{},
	}

	// Text styles
	if node.Style != nil {
		styles.Styles = append(styles.Styles, StyleInfo{
			Type:    "text",
			Applied: textStyleToMap(node.Style),
		})
	}

	// Fill styles
	if len(node.Fills) > 0 {
		for i, fill := range node.Fills {
			styles.Styles = append(styles.Styles, StyleInfo{
				Type:    "fill",
				Applied: paintToMap(&fill, i),
			})
		}
	}

	// Effect styles
	if len(node.Effects) > 0 {
		for i, effect := range node.Effects {
			styles.Styles = append(styles.Styles, StyleInfo{
				Type:    "effect",
				Applied: effectToMap(&effect, i),
			})
		}
	}

	// Grid styles
	if len(node.LayoutGrids) > 0 {
		for i, grid := range node.LayoutGrids {
			styles.Styles = append(styles.Styles, StyleInfo{
				Type:    "grid",
				Applied: layoutGridToMap(&grid, i),
			})
		}
	}

	// Referenced styles (node.Styles map)
	for styleType, styleKey := range node.Styles {
		styles.Styles = append(styles.Styles, StyleInfo{
			Type:     styleType,
			StyleKey: styleKey,
			// Could attempt to resolve style definition here
		})
	}

	return styles
}

func textStyleToMap(ts *api.TextStyle) map[string]interface{} {
	m := map[string]interface{}{}
	if ts.FontFamily != "" {
		m["font_family"] = ts.FontFamily
	}
	if ts.FontWeight != 0 {
		m["font_weight"] = ts.FontWeight
	}
	if ts.FontSize != 0 {
		m["font_size"] = ts.FontSize
	}
	// Add more fields as needed
	return m
}

func paintToMap(paint *api.Paint, index int) map[string]interface{} {
	m := map[string]interface{}{
		"index": index,
		"type":  paint.Type,
	}
	if paint.Color != nil {
		m["color"] = fmt.Sprintf("rgba(%.0f,%.0f,%.0f,%.2f)", paint.Color.R*255, paint.Color.G*255, paint.Color.B*255, paint.Color.A)
	}
	if paint.Opacity != 0 {
		m["opacity"] = paint.Opacity
	}
	return m
}

func effectToMap(effect *api.Effect, index int) map[string]interface{} {
	m := map[string]interface{}{
		"index": index,
		"type":  effect.Type,
	}
	if effect.Color != nil {
		m["color"] = fmt.Sprintf("rgba(%.0f,%.0f,%.0f,%.2f)", effect.Color.R*255, effect.Color.G*255, effect.Color.B*255, effect.Color.A)
	}
	if effect.Radius != 0 {
		m["radius"] = effect.Radius
	}
	if effect.Spread != 0 {
		m["spread"] = effect.Spread
	}
	if effect.Visible {
		m["visible"] = effect.Visible
	}
	return m
}

func layoutGridToMap(grid *api.LayoutGrid, index int) map[string]interface{} {
	m := map[string]interface{}{
		"index": index,
		"type":  grid.Type,
	}
	if grid.Color != nil {
		m["color"] = fmt.Sprintf("rgba(%.0f,%.0f,%.0f,%.2f)", grid.Color.R*255, grid.Color.G*255, grid.Color.B*255, grid.Color.A)
	}
	if grid.Size != 0 {
		m["size"] = grid.Size
	}
	if grid.Gutter != 0 {
		m["gutter"] = grid.Gutter
	}
	if grid.Offset != 0 {
		m["offset"] = grid.Offset
	}
	if grid.Visible {
		m["visible"] = grid.Visible
	}
	return m
}
