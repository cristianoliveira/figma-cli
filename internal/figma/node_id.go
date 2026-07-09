package figma

import "strings"

// NormalizeNodeID converts a node ID from URL-style (hyphens) to API-style (colons).
func NormalizeNodeID(nodeID string) string {
	return strings.ReplaceAll(nodeID, "-", ":")
}

// ResolveNodeIDs resolves the effective node IDs from an explicit flag and URL-parsed IDs.
// If explicitNodeID is non-empty, it wins. Otherwise URL node IDs are used.
func ResolveNodeIDs(input *FileInput, explicitNodeID string) []string {
	if explicitNodeID == "" {
		return input.NodeIDs
	}

	parts := strings.Split(explicitNodeID, ",")
	nodeIDs := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		nodeIDs = append(nodeIDs, NormalizeNodeID(part))
	}
	return nodeIDs
}
