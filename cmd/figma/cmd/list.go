package cmd

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/cristianoliveira/figma-cli/internal/api"
	"github.com/cristianoliveira/figma-cli/internal/config"
	"github.com/cristianoliveira/figma-cli/internal/logging"
	"github.com/cristianoliveira/figma-cli/pkg/figma"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list <file-key-or-url>",
	Short: "Enumerate nodes by type within a file",
	Long: `Enumerate nodes by type within a Figma file, with filtering options.

Supports filtering by node type (FRAME, TEXT, COMPONENT, etc.), name patterns
(wildcards: *, ?, []) and pagination for large result sets.

Examples:
  figma list ABC123
  figma list https://www.figma.com/file/ABC123/My-Design --type FRAME
  figma list ABC123 --name "*button*" --limit 10
  figma list ABC123 --type TEXT --offset 20 --limit 5
  figma list ABC123 --output json`,
	Args: cobra.ExactArgs(1),
	RunE: runList,
}

func init() {
	// Add flags
	listCmd.Flags().String("type", "", "Filter by node type (e.g., FRAME, COMPONENT, TEXT)")
	listCmd.Flags().String("name", "", "Wildcard pattern to match node names (*, ?, [])")
	listCmd.Flags().Bool("case-sensitive", false, "Case-sensitive matching for name patterns")
	listCmd.Flags().Int("limit", 0, "Maximum number of nodes to return (0 for unlimited)")
	listCmd.Flags().Int("offset", 0, "Number of nodes to skip before returning results")
	listCmd.Flags().Bool("recursive", true, "Traverse node hierarchy recursively")
	listCmd.Flags().Int("depth", -1, "Maximum depth to traverse (-1 for unlimited)")
	// Output format flag inherited from global --output
}

// ListOptions encapsulates list parameters
type ListOptions struct {
	TypeFilter    string
	NamePattern   string
	CaseSensitive bool
	Limit         int
	Offset        int
	Recursive     bool
	MaxDepth      int
}

// ListNode represents a node in the list output
type ListNode struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	FileKey string `json:"file_key"`
	NodeURL string `json:"node_url,omitempty"`
	Path    string `json:"path,omitempty"` // hierarchical path
}

// wildcardToRegex converts a wildcard pattern to regex.
// Supports * (matches any sequence), ? (matches any single character),
// and keeps [abc] character classes as-is.
func wildcardToRegex(pattern string) string {
	// Escape regex special characters except *, ?, [
	// We'll replace * and ? after escaping.
	escaped := regexp.QuoteMeta(pattern)
	// Now unescape our wildcards
	escaped = strings.ReplaceAll(escaped, `\*`, `.*`)
	escaped = strings.ReplaceAll(escaped, `\?`, `.`)
	// Note: [ and ] are already escaped by QuoteMeta, we need to unescape them
	// but keep them as character class. We'll replace \[ with [ and \] with ]
	escaped = strings.ReplaceAll(escaped, `\[`, `[`)
	escaped = strings.ReplaceAll(escaped, `\]`, `]`)
	return "^" + escaped + "$"
}

// matchesFilter returns true if the node matches the list options.
func matchesFilter(node *api.Node, opts *ListOptions, nameRegex *regexp.Regexp) bool {
	// Type filter
	if opts.TypeFilter != "" && node.Type != opts.TypeFilter {
		return false
	}
	// Name pattern filter
	if nameRegex != nil && !nameRegex.MatchString(node.Name) {
		return false
	}
	return true
}

// collectNodes traverses the node tree and returns matching nodes.
func collectNodes(root *api.Node, fileKey string, opts *ListOptions) []ListNode {
	var results []ListNode
	var nameRegex *regexp.Regexp
	if opts.NamePattern != "" {
		pattern := wildcardToRegex(opts.NamePattern)
		if !opts.CaseSensitive {
			pattern = "(?i)" + pattern
		}
		r, err := regexp.Compile(pattern)
		if err != nil {
			// Invalid pattern, treat as no matches
			return results
		}
		nameRegex = r
	}

	maxDepth := opts.MaxDepth
	if !opts.Recursive {
		maxDepth = 1 // node itself + immediate children
	}

	var dfs func(node *api.Node, depth int, path string)
	dfs = func(node *api.Node, depth int, path string) {
		if node == nil {
			return
		}
		currentPath := path + "/" + node.Name
		if matchesFilter(node, opts, nameRegex) {
			results = append(results, ListNode{
				ID:      node.ID,
				Name:    node.Name,
				Type:    node.Type,
				FileKey: fileKey,
				NodeURL: fmt.Sprintf("https://www.figma.com/design/%s?node-id=%s", fileKey, node.ID),
				Path:    currentPath,
			})
		}
		// Stop traversing children if maxDepth reached
		if maxDepth >= 0 && depth >= maxDepth {
			return
		}
		for i := range node.Children {
			dfs(&node.Children[i], depth+1, currentPath)
		}
	}
	dfs(root, 0, "")
	return results
}

// runList implements the list command.
func runList(cmd *cobra.Command, args []string) error {
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
	typeFilter, _ := cmd.Flags().GetString("type")
	namePattern, _ := cmd.Flags().GetString("name")
	caseSensitive, _ := cmd.Flags().GetBool("case-sensitive")
	limit, _ := cmd.Flags().GetInt("limit")
	offset, _ := cmd.Flags().GetInt("offset")
	recursive, _ := cmd.Flags().GetBool("recursive")
	depth, _ := cmd.Flags().GetInt("depth")

	opts := &ListOptions{
		TypeFilter:    typeFilter,
		NamePattern:   namePattern,
		CaseSensitive: caseSensitive,
		Limit:         limit,
		Offset:        offset,
		Recursive:     recursive,
		MaxDepth:      depth,
	}

	// Collect matching nodes
	allNodes := collectNodes(root, fileKey, opts)
	total := len(allNodes)

	// Apply pagination
	start := opts.Offset
	if start < 0 {
		start = 0
	}
	if start > total {
		start = total
	}
	end := total
	if opts.Limit > 0 && start+opts.Limit < end {
		end = start + opts.Limit
	}
	nodes := allNodes[start:end]

	// Determine output format
	outputFormat := cfg.OutputFormat
	if outputFormat == "" {
		outputFormat = config.OutputFormatText
	}

	// Output results
	switch outputFormat {
	case config.OutputFormatJSON:
		data, err := json.MarshalIndent(nodes, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		cmd.Println(string(data))
	case config.OutputFormatYAML:
		data, err := yaml.Marshal(nodes)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		cmd.Println(string(data))
	default: // text
		if len(nodes) == 0 {
			cmd.Println("No nodes found.")
			return nil
		}
		cmd.Printf("Found %d nodes (showing %d):\n\n", total, len(nodes))
		for _, n := range nodes {
			cmd.Printf("%s (%s) [%s]\n", n.Name, n.Type, n.ID)
			if n.Path != "" {
				cmd.Printf("  Path: %s\n", n.Path)
			}
			cmd.Printf("  URL: %s\n\n", n.NodeURL)
		}
	}
	return nil
}
