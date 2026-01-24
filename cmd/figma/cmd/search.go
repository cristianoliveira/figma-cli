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

const nodeTypeText = "TEXT"

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search <figma-url> <query>",
	Short: "Search nodes by text content, type, or name pattern",
	Long: `Search nodes within a Figma file by text content, node type, or name pattern.

Multiple search modes are supported:
- Text content search: matches text layers containing the query string
- Node type filter: filter nodes by type (FRAME, COMPONENT, INSTANCE, TEXT, etc.)
- Name pattern search: match node names using regular expressions

Flags control case sensitivity, exact matching, and output format.

Examples:
  figma search https://www.figma.com/file/ABC123/My-Design "hello"
  figma search https://www.figma.com/design/ABC123/My-Design "hello" --type TEXT
  figma search https://www.figma.com/file/ABC123/My-Design "button.*" --name
  figma search https://www.figma.com/file/ABC123/My-Design "frame" --type FRAME --case-sensitive`,
	Args: cobra.ExactArgs(2),
	RunE: runSearch,
}

func init() {
	// Add flags
	searchCmd.Flags().String("type", "", "Filter by node type (e.g., FRAME, COMPONENT, TEXT). If set to 'text', searches within text content.")
	searchCmd.Flags().String("name", "", "Regex pattern to match node names")
	searchCmd.Flags().Bool("case-sensitive", false, "Case-sensitive matching for text and name searches")
	searchCmd.Flags().Bool("exact", false, "Exact word matching (requires word boundaries) for text and name searches")
	// Output format flag inherited from global --output
}

// SearchOptions encapsulates search parameters
type SearchOptions struct {
	Query         string
	TypeFilter    string
	NamePattern   string
	CaseSensitive bool
	Exact         bool
}

// SearchResult represents a matching node with context
type SearchResult struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	FileKey string `json:"file_key"`
	NodeURL string `json:"node_url,omitempty"`
	Path    string `json:"path,omitempty"` // hierarchical path
	// Optional text content if matching text node
	Characters string `json:"characters,omitempty"`
}

// textMatches returns true if the text content matches the query.
func textMatches(text, query string, caseSensitive, exact bool) bool {
	if !caseSensitive {
		text = strings.ToLower(text)
		query = strings.ToLower(query)
	}
	if exact {
		// Use word boundary regex to match whole words
		// Escape query for regex special characters
		pattern := `\b` + regexp.QuoteMeta(query) + `\b`
		if !caseSensitive {
			pattern = "(?i)" + pattern
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			// fallback to equality
			return text == query
		}
		return re.MatchString(text)
	}
	return strings.Contains(text, query)
}

// matchesSearch returns true if the node matches the search options
func matchesSearch(node *api.Node, opts *SearchOptions, nameRegex *regexp.Regexp) bool {
	// Type filter (if set and not "text")
	if opts.TypeFilter != "" && opts.TypeFilter != "text" {
		if node.Type != opts.TypeFilter {
			return false
		}
	}

	// Name pattern filter
	if nameRegex != nil {
		if !nameRegex.MatchString(node.Name) {
			return false
		}
	}

	// Text content search (applies only to TEXT nodes)
	// If query is empty, we don't filter by text content.
	// If typeFilter is "text", we require node.Type == "TEXT" (implicitly).
	textSearch := opts.Query != ""
	typeFilterIsText := opts.TypeFilter == "text"

	// Determine if we need to consider text content
	if textSearch || typeFilterIsText {
		// Only TEXT nodes have characters
		if node.Type != nodeTypeText || node.Characters == "" {
			return false
		}
		// If query is empty, we match all TEXT nodes (when typeFilter is "text")
		if textSearch {
			if !textMatches(node.Characters, opts.Query, opts.CaseSensitive, opts.Exact) {
				return false
			}
		}
	}

	// If no filters at all, return false (no matches)
	hasFilter := opts.TypeFilter != "" || opts.NamePattern != "" || opts.Query != ""
	return hasFilter
}

// searchNodes performs DFS and returns matching nodes
func searchNodes(root *api.Node, fileKey string, opts *SearchOptions) []SearchResult {
	var results []SearchResult
	var nameRegex *regexp.Regexp
	if opts.NamePattern != "" {
		pattern := opts.NamePattern
		if !opts.CaseSensitive {
			pattern = "(?i)" + pattern
		}
		if opts.Exact {
			pattern = "^" + pattern + "$"
		}
		r, err := regexp.Compile(pattern)
		if err != nil {
			// Invalid regex, treat as no matches
			return results
		}
		nameRegex = r
	}

	var dfs func(node *api.Node, path string)
	dfs = func(node *api.Node, path string) {
		if node == nil {
			return
		}
		currentPath := path + "/" + node.Name
		if matchesSearch(node, opts, nameRegex) {
			result := SearchResult{
				ID:      node.ID,
				Name:    node.Name,
				Type:    node.Type,
				FileKey: fileKey,
				NodeURL: fmt.Sprintf("https://www.figma.com/design/%s?node-id=%s", fileKey, node.ID),
				Path:    currentPath,
			}
			if node.Type == nodeTypeText {
				result.Characters = node.Characters
			}
			results = append(results, result)
		}
		for i := range node.Children {
			dfs(&node.Children[i], currentPath)
		}
	}
	dfs(root, "")
	return results
}

// nolint:dupl
func runSearch(cmd *cobra.Command, args []string) error {
	url := args[0]
	query := args[1]
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

	// Fetch root node (file or specific node)
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

	// Parse flags
	typeFilter, _ := cmd.Flags().GetString("type")
	namePattern, _ := cmd.Flags().GetString("name")
	caseSensitive, _ := cmd.Flags().GetBool("case-sensitive")
	exact, _ := cmd.Flags().GetBool("exact")

	opts := &SearchOptions{
		Query:         query,
		TypeFilter:    typeFilter,
		NamePattern:   namePattern,
		CaseSensitive: caseSensitive,
		Exact:         exact,
	}

	// Perform search
	results := searchNodes(root, parsed.FileKey, opts)

	// Determine output format
	outputFormat := cfg.OutputFormat
	if outputFormat == "" {
		outputFormat = config.OutputFormatText
	}

	// Output results
	switch outputFormat {
	case config.OutputFormatJSON:
		data, err := json.MarshalIndent(results, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		cmd.Println(string(data))
	case config.OutputFormatYAML:
		data, err := yaml.Marshal(results)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %w", err)
		}
		cmd.Println(string(data))
	default: // text
		if len(results) == 0 {
			cmd.Println("No matches found.")
			return nil
		}
		for _, r := range results {
			cmd.Printf("%s (%s) [%s]\n", r.Name, r.Type, r.ID)
			if r.Path != "" {
				cmd.Printf("  Path: %s\n", r.Path)
			}
			if r.Characters != "" {
				cmd.Printf("  Text: %q\n", r.Characters)
			}
			cmd.Printf("  URL: %s\n\n", r.NodeURL)
		}
	}
	return nil
}
