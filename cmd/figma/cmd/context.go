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

// contextCmd represents the context command
var contextCmd = &cobra.Command{
	Use:   "context <figma-url>",
	Short: "Display node hierarchy and relationships in a tree format",
	Long: `Display node hierarchy and relationships in a tree format.

The command parses the Figma URL to extract the file key and node ID,
fetches the node hierarchy from the Figma API, and displays the tree
showing ancestors, current node, and children.

Flags allow controlling the number of ancestor levels, sibling inclusion,
and child levels to display.

Examples:
  figma context https://www.figma.com/file/ABC123/My-Design
  figma context https://www.figma.com/design/ABC123/My-Design?node-id=1:2
  figma context https://www.figma.com/file/ABC123/My-Design?node-id=1:2 --ancestors 2 --children 1`,
	Args: cobra.ExactArgs(1),
	RunE: runContext,
}

func init() {
	// Local flags for context command
	contextCmd.Flags().Int("ancestors", -1, "Number of parent levels to show (default all, -1 for all)")
	contextCmd.Flags().Bool("siblings", false, "Show nodes at same level")
	contextCmd.Flags().Int("children", -1, "Number of child levels to show (default all, -1 for all)")
	contextCmd.Flags().String("output-format", config.OutputFormatText, "Output format: json, yaml, text")
	contextCmd.Flags().StringP("output", "o", "", "Output format: json, yaml, text (overrides --output-format)")
}

// findNodePath performs DFS to find a node by ID and returns the path from root to node (including root and node).
// If node is not found, returns nil path and nil node.
func findNodePath(root *api.Node, targetID string) ([]*api.Node, *api.Node) {
	var path []*api.Node
	var found *api.Node

	var dfs func(node *api.Node, currentPath []*api.Node) bool
	dfs = func(node *api.Node, currentPath []*api.Node) bool {
		// Add current node to path
		currentPath = append(currentPath, node)
		if node.ID == targetID {
			path = make([]*api.Node, len(currentPath))
			copy(path, currentPath)
			found = node
			return true
		}
		for i := range node.Children {
			if dfs(&node.Children[i], currentPath) {
				return true
			}
		}
		return false
	}

	dfs(root, nil)
	return path, found
}

// collectSiblings returns all children of the parent of the target node (excluding the target node if excludeSelf true).
func collectSiblings(path []*api.Node) []*api.Node {
	if len(path) < 2 {
		// root node has no parent, no siblings
		return nil
	}
	parent := path[len(path)-2]
	// Return all children of parent
	siblings := make([]*api.Node, 0, len(parent.Children))
	for i := range parent.Children {
		siblings = append(siblings, &parent.Children[i])
	}
	return siblings
}

// limitDepth returns a copy of the node with children limited to the given depth.
// If depth < 0, returns the original node (no copy). If depth == 0, children are omitted.
// If depth > 0, children are limited to that depth (1 = immediate children only).
func limitDepth(node *api.Node, depth int) *api.Node {
	if node == nil {
		return nil
	}
	if depth < 0 {
		// unlimited depth, return original (no copy)
		return node
	}
	copyNode := *node
	if depth == 0 {
		copyNode.Children = nil
		return &copyNode
	}
	// depth > 0: limit children depth
	copyNode.Children = make([]api.Node, len(node.Children))
	for i := range node.Children {
		childCopy := limitDepth(&node.Children[i], depth-1)
		if childCopy != nil {
			copyNode.Children[i] = *childCopy
		}
	}
	return &copyNode
}

// printTreeNode prints a single node line with indentation and optional highlight.
func printTreeNode(cmd *cobra.Command, node *api.Node, indent int, highlight bool) {
	prefix := ""
	if indent > 0 {
		prefix = strings.Repeat("  ", indent-1) + "├─ "
	}
	// If this is the root (indent 0), no prefix
	if indent == 0 {
		prefix = ""
	}
	marker := ""
	if highlight {
		marker = "* "
	}
	line := fmt.Sprintf("%s%s%s (%s)", prefix, marker, node.Name, node.Type)
	cmd.Println(line)
}

// printTree prints a subtree starting from node, up to maxDepth (relative depth).
// If maxDepth < 0, unlimited depth.
// currentDepth is the depth relative to the starting node (0 for the node itself).
// highlight indicates whether the starting node should be highlighted.
func printTree(cmd *cobra.Command, node *api.Node, currentDepth, maxDepth int, highlight bool) {
	if maxDepth >= 0 && currentDepth > maxDepth {
		return
	}
	// Determine if this node should be highlighted (only the starting node)
	hl := highlight && currentDepth == 0
	printTreeNode(cmd, node, currentDepth, hl)

	// Recursively print children
	for i := range node.Children {
		printTree(cmd, &node.Children[i], currentDepth+1, maxDepth, false)
	}
}

func runContext(cmd *cobra.Command, args []string) error {
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
	ancestors, _ := cmd.Flags().GetInt("ancestors")
	siblings, _ := cmd.Flags().GetBool("siblings")
	children, _ := cmd.Flags().GetInt("children")

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
	case config.OutputFormatYAML:
		// keep as is
	case config.OutputFormatText:
		// keep as is
	default:
		return fmt.Errorf("unsupported output format: %s (supported: %s, %s, %s)", outputFormat, config.OutputFormatJSON, config.OutputFormatYAML, config.OutputFormatText)
	}
	logger.Debug(ctx, "Output format", logging.String("format", outputFormat))

	// Fetch the file (full tree)
	logger.Debug(ctx, "Fetching file", logging.String("file_key", parsed.FileKey))
	file, err := client.GetFile(ctx, parsed.FileKey)
	if err != nil {
		logger.Error(ctx, "Failed to fetch file", logging.Err(err), logging.String("file_key", parsed.FileKey))
		return fmt.Errorf("failed to fetch file: %w", err)
	}
	logger.Debug(ctx, "File fetched successfully", logging.String("file_name", file.Name), logging.String("last_modified", file.LastModified))

	root := file.Document
	if root == nil {
		return fmt.Errorf("file document is empty")
	}

	var targetNode *api.Node
	var path []*api.Node
	if parsed.NodeID == "" {
		// No node ID specified, target is the root
		targetNode = root
		path = []*api.Node{root}
	} else {
		// Search for node in tree
		logger.Debug(ctx, "Searching for node", logging.String("node_id", parsed.NodeID))
		path, targetNode = findNodePath(root, parsed.NodeID)
		if targetNode == nil {
			return fmt.Errorf("node %q not found in file", parsed.NodeID)
		}
		logger.Debug(ctx, "Node found", logging.String("node_name", targetNode.Name), logging.String("node_type", targetNode.Type))
	}

	// Output based on format
	switch outputFormat {
	case config.OutputFormatJSON, config.OutputFormatYAML:
		// Prepare structured output
		output := make(map[string]interface{})

		// Limit ancestors according to flag
		ancestorLevels := ancestors
		if ancestorLevels < 0 {
			ancestorLevels = len(path) - 1 // all ancestors
		}
		if ancestorLevels > len(path)-1 {
			ancestorLevels = len(path) - 1
		}
		if ancestorLevels > 0 {
			ancestorsList := make([]*api.Node, ancestorLevels)
			for i := 0; i < ancestorLevels; i++ {
				ancestorsList[i] = path[i]
			}
			output["ancestors"] = ancestorsList
		}

		// Siblings if requested
		if siblings && len(path) >= 2 {
			parent := path[len(path)-2]
			siblingsList := make([]*api.Node, 0, len(parent.Children))
			for i := range parent.Children {
				if parent.Children[i].ID != targetNode.ID {
					siblingsList = append(siblingsList, &parent.Children[i])
				}
			}
			if len(siblingsList) > 0 {
				output["siblings"] = siblingsList
			}
		}

		// Current node with limited children depth
		childDepth := children
		var currentNode interface{}
		if childDepth != 0 {
			currentNode = limitDepth(targetNode, childDepth)
		} else {
			// show no children
			currentNode = limitDepth(targetNode, 0)
		}
		output["current"] = currentNode

		// Output using the same function as get command
		var format string
		if outputFormat == config.OutputFormatJSON {
			format = config.OutputFormatJSON
		} else {
			format = config.OutputFormatYAML
		}
		return outputData(cmd, output, format)

	case config.OutputFormatText:
		// Print tree hierarchy according to flags

		// Determine ancestor levels to show
		ancestorLevels := ancestors
		if ancestorLevels < 0 {
			ancestorLevels = len(path) - 1 // all ancestors (excluding target node)
		}
		// Ensure not more than available
		if ancestorLevels > len(path)-1 {
			ancestorLevels = len(path) - 1
		}

		// Print ancestors from root up to ancestorLevels
		// We'll print each ancestor as a single line with appropriate indentation
		// The last ancestor is the parent of target node (if ancestorLevels > 0)
		for i := 0; i < ancestorLevels; i++ {
			node := path[i]
			printTreeNode(cmd, node, i, false)
		}

		// Determine if we need to show siblings
		if siblings && len(path) >= 2 {
			parent := path[len(path)-2]
			// Print all children of parent (siblings) with appropriate indentation
			siblingLevel := ancestorLevels
			for i := range parent.Children {
				sib := &parent.Children[i]
				// Skip the target node if it's among siblings (we'll print it separately)
				if sib.ID == targetNode.ID {
					continue
				}
				printTreeNode(cmd, sib, siblingLevel, false)
			}
		}

		// Print target node (highlighted)
		targetIndent := ancestorLevels
		printTreeNode(cmd, targetNode, targetIndent, true)

		// Print children
		childDepth := children
		maxDepth := childDepth // max depth relative to target node (0 = none, 1 = immediate children, etc.)
		if childDepth < 0 {
			// unlimited depth
			maxDepth = -1
		}
		if maxDepth != 0 {
			// Print children recursively (starting depth 1 relative to target)
			for i := range targetNode.Children {
				printTree(cmd, &targetNode.Children[i], targetIndent+1, maxDepth, false)
			}
		}
	default:
		return fmt.Errorf("output format %q not implemented for context command", outputFormat)
	}

	return nil
}
