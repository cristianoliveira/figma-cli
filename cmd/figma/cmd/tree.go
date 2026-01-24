package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/api"
	"github.com/cristianoliveira/figma-cli/internal/config"
	"github.com/cristianoliveira/figma-cli/internal/logging"
	"github.com/cristianoliveira/figma-cli/pkg/figma"
	"github.com/spf13/cobra"
)

// treeCmd represents the tree command
var treeCmd = &cobra.Command{
	Use:   "tree <file-key-or-url>",
	Short: "Display node hierarchy in tree format",
	Long: `Display complete node hierarchy from a starting point.

Supports max-depth parameter and filtering by node type.
Output can be ASCII tree or structured JSON/YAML.

Examples:
  figma tree ABC123
  figma tree https://www.figma.com/file/ABC123/My-Design --node-id 1:2
  figma tree ABC123 --type FRAME --max-depth 2
  figma tree ABC123 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: runTree,
}

func init() {
	treeCmd.Flags().String("node-id", "", "Start from specific node ID")
	treeCmd.Flags().String("type", "", "Filter by node type (e.g., FRAME, COMPONENT, TEXT)")
	treeCmd.Flags().Int("max-depth", -1, "Maximum depth to traverse (-1 for unlimited)")
	treeCmd.Flags().String("output", "", "Output format: json, yaml, text (default: text)")
}

// TreeOptions encapsulates tree parameters
type TreeOptions struct {
	NodeID       string
	TypeFilter   string
	MaxDepth     int
	OutputFormat string
}

// TreeNode represents a node in the filtered tree output
type TreeNode struct {
	ID       string      `json:"id"`
	Name     string      `json:"name"`
	Type     string      `json:"type"`
	Children []*TreeNode `json:"children,omitempty"`
}

// filterTree returns a filtered copy of the node tree.
// If typeFilter is empty, all nodes are included.
// maxDepth limits the depth relative to the starting node (0 = only root).
// Returns nil if node and its children do not match filter.
func filterTree(node *api.Node, typeFilter string, maxDepth int) *api.Node {
	if node == nil {
		return nil
	}
	if maxDepth == 0 {
		// Depth limit reached, include node but no children
		copyNode := *node
		copyNode.Children = nil
		return &copyNode
	}

	// Determine if this node matches the type filter
	matchesType := typeFilter == "" || node.Type == typeFilter

	// Filter children recursively
	var filteredChildren []api.Node
	for i := range node.Children {
		childDepth := maxDepth
		if maxDepth > 0 {
			childDepth = maxDepth - 1
		}
		filteredChild := filterTree(&node.Children[i], typeFilter, childDepth)
		if filteredChild != nil {
			filteredChildren = append(filteredChildren, *filteredChild)
		}
	}

	// If node matches type or has any filtered children, include it
	if matchesType || len(filteredChildren) > 0 {
		copyNode := *node
		copyNode.Children = filteredChildren
		return &copyNode
	}

	// Node does not match and has no matching children
	return nil
}

// toTreeNode converts an api.Node to a TreeNode (for JSON/YAML output).
func toTreeNode(node *api.Node) *TreeNode {
	if node == nil {
		return nil
	}
	tn := &TreeNode{
		ID:   node.ID,
		Name: node.Name,
		Type: node.Type,
	}
	if len(node.Children) > 0 {
		tn.Children = make([]*TreeNode, 0, len(node.Children))
		for i := range node.Children {
			childTN := toTreeNode(&node.Children[i])
			if childTN != nil {
				tn.Children = append(tn.Children, childTN)
			}
		}
	}
	return tn
}

// runTree implements the tree command.
func runTree(cmd *cobra.Command, args []string) error {
	arg := args[0]
	ctx := cmd.Context()

	logger, err := GetLogger(cmd)
	if err != nil {
		logger = logging.NewNopLogger()
	}

	// Try to parse as Figma URL; if successful, extract file key.
	// Otherwise treat as raw file key.
	var fileKey, nodeID string
	parsed, err := figma.ParseURL(arg)
	if err == nil {
		logger.Debug(ctx, "Parsed Figma URL", logging.String("file_key", parsed.FileKey), logging.String("node_id", parsed.NodeID))
		fileKey = parsed.FileKey
		nodeID = parsed.NodeID
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

	// Fetch root node (file or specific node)
	var root *api.Node
	if nodeID != "" {
		logger.Debug(ctx, "Fetching node", logging.String("file_key", fileKey), logging.String("node_id", nodeID))
		node, err := client.GetNode(ctx, fileKey, nodeID)
		if err != nil {
			logger.Error(ctx, "Failed to fetch node", logging.Err(err), logging.String("file_key", fileKey), logging.String("node_id", nodeID))
			return fmt.Errorf("failed to fetch node: %w", err)
		}
		logger.Debug(ctx, "Node fetched successfully", logging.String("node_name", node.Name), logging.String("node_type", node.Type))
		root = node
	} else {
		logger.Debug(ctx, "Fetching file", logging.String("file_key", fileKey))
		file, err := client.GetFile(ctx, fileKey)
		if err != nil {
			logger.Error(ctx, "Failed to fetch file", logging.Err(err), logging.String("file_key", fileKey))
			return fmt.Errorf("failed to fetch file: %w", err)
		}
		logger.Debug(ctx, "File fetched successfully", logging.String("file_name", file.Name), logging.String("last_modified", file.LastModified))
		root = file.Document
	}

	if root == nil {
		return fmt.Errorf("no node found")
	}

	// Parse flags
	nodeIDFlag, _ := cmd.Flags().GetString("node-id")
	typeFilter, _ := cmd.Flags().GetString("type")
	maxDepth, _ := cmd.Flags().GetInt("max-depth")
	outputFormat, _ := cmd.Flags().GetString("output")
	if outputFormat == "" {
		outputFormat = cfg.OutputFormat
	}
	if outputFormat == "" {
		outputFormat = config.OutputFormatText
	}

	// If node-id flag provided, find that node as root
	if nodeIDFlag != "" {
		_, targetNode := findNodePath(root, nodeIDFlag)
		if targetNode == nil {
			return fmt.Errorf("node %q not found in file", nodeIDFlag)
		}
		root = targetNode
		// Adjust maxDepth relative to new root? Keep as is.
	}

	// Apply filtering
	filteredRoot := filterTree(root, typeFilter, maxDepth)
	if filteredRoot == nil {
		// No nodes match the filter
		cmd.Println("No nodes found matching the filter.")
		return nil
	}

	// Output based on format
	switch outputFormat {
	case config.OutputFormatJSON, config.OutputFormatYAML:
		// Convert filtered tree to TreeNode for cleaner JSON/YAML output
		treeNode := toTreeNode(filteredRoot)
		return outputData(cmd, treeNode, outputFormat)
	case config.OutputFormatText:
		// ASCII tree output
		printTree(cmd, filteredRoot, 0, -1, false)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
	return nil
}
