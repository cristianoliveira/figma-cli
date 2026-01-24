package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/api"
	"github.com/cristianoliveira/figma-cli/internal/logging"
	"github.com/cristianoliveira/figma-cli/pkg/figma"
	"github.com/spf13/cobra"
)

func init() {
	// Add flags to text command
	textCmd.Flags().Bool("recursive", false, "Extract text from all descendant nodes")
	textCmd.Flags().String("format", "plain", "Output format: plain, json, csv")
	textCmd.RunE = runText
}

// TextItem represents a text node with its metadata.
type TextItem struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Characters string `json:"characters"`
	Path       string `json:"path,omitempty"` // optional hierarchical path
}

// collectTexts traverses the node tree and returns all text items.
// If recursive is false, only direct children of the node are examined (depth 1).
// If recursive is true, all descendants are examined (unlimited depth).
func collectTexts(node *api.Node, recursive bool) []TextItem {
	var items []TextItem
	maxDepth := 1 // node itself + immediate children
	if recursive {
		maxDepth = -1 // unlimited
	}
	collectTextsDFS(node, maxDepth, "", &items)
	return items
}

// collectTextsDFS performs depth-first search up to maxDepth.
// maxDepth: 0 = node only, 1 = node + children, -1 = unlimited.
// path is the slash-separated path of node names from root.
func collectTextsDFS(node *api.Node, maxDepth int, path string, items *[]TextItem) {
	if node == nil {
		return
	}
	// Build current path
	currentPath := path + "/" + node.Name

	// If this is a TEXT node, add its characters
	if node.Type == "TEXT" && node.Characters != "" {
		*items = append(*items, TextItem{
			ID:         node.ID,
			Name:       node.Name,
			Type:       node.Type,
			Characters: node.Characters,
			Path:       currentPath,
		})
	}

	// If maxDepth == 0, stop traversing children.
	if maxDepth == 0 {
		return
	}
	// Determine depth for children
	childDepth := maxDepth - 1
	if maxDepth < 0 {
		childDepth = -1 // unlimited for children as well
	}
	// Recursively process children
	for i := range node.Children {
		collectTextsDFS(&node.Children[i], childDepth, currentPath, items)
	}
}

// runText implements the text command.
// nolint:dupl
func runText(cmd *cobra.Command, args []string) error {
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

	// Get flags
	recursive, _ := cmd.Flags().GetBool("recursive")
	format, _ := cmd.Flags().GetString("format")

	// Validate format
	switch format {
	case "plain", "json", "csv":
		// ok
	default:
		return fmt.Errorf("unsupported format %q (supported: plain, json, csv)", format)
	}

	// Fetch node or file
	var root *api.Node
	if parsed.NodeID != "" {
		logger.Debug(ctx, "Fetching node", logging.String("file_key", parsed.FileKey), logging.String("node_id", parsed.NodeID))
		node, err := client.GetNode(ctx, parsed.FileKey, parsed.NodeID)
		if err != nil {
			logger.Error(ctx, "Failed to fetch node", logging.Err(err), logging.String("file_key", parsed.FileKey), logging.String("node_id", parsed.NodeID))
			return fmt.Errorf("failed to fetch node: %w", err)
		}
		logger.Debug(ctx, "Node fetched successfully", logging.String("node_name", node.Name), logging.String("node_type", node.Type))
		root = node
	} else {
		logger.Debug(ctx, "Fetching file", logging.String("file_key", parsed.FileKey))
		file, err := client.GetFile(ctx, parsed.FileKey)
		if err != nil {
			logger.Error(ctx, "Failed to fetch file", logging.Err(err), logging.String("file_key", parsed.FileKey))
			return fmt.Errorf("failed to fetch file: %w", err)
		}
		logger.Debug(ctx, "File fetched successfully", logging.String("file_name", file.Name), logging.String("last_modified", file.LastModified))
		root = file.Document
	}

	if root == nil {
		return fmt.Errorf("no node found")
	}

	// Collect text items
	items := collectTexts(root, recursive)

	// Output based on format
	switch format {
	case "plain":
		for _, item := range items {
			cmd.Println(item.Characters)
		}
	case "json":
		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		cmd.Println(string(data))
	case "csv":
		writer := csv.NewWriter(cmd.OutOrStdout())
		defer writer.Flush()
		// Write header
		if err := writer.Write([]string{"id", "name", "type", "characters", "path"}); err != nil {
			return fmt.Errorf("failed to write CSV header: %w", err)
		}
		for _, item := range items {
			if err := writer.Write([]string{item.ID, item.Name, item.Type, item.Characters, item.Path}); err != nil {
				return fmt.Errorf("failed to write CSV row: %w", err)
			}
		}
	}

	return nil
}
