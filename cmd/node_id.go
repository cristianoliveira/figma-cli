package cmd

import "strings"

func normalizeNodeID(nodeID string) string {
	return strings.ReplaceAll(nodeID, "-", ":")
}

func resolveNodeIDs(input *fileInput, explicitNodeID string) []string {
	if explicitNodeID == "" {
		return input.nodeIDs
	}

	parts := strings.Split(explicitNodeID, ",")
	nodeIDs := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		nodeIDs = append(nodeIDs, normalizeNodeID(part))
	}
	return nodeIDs
}
