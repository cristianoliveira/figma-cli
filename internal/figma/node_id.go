package figma

import (
	"fmt"
	"strings"
)

// NormalizeNodeID converts a node ID from URL-style (hyphens) to API-style (colons).
func NormalizeNodeID(nodeID string) string {
	return strings.ReplaceAll(nodeID, "-", ":")
}

// ResolveNodeIDs resolves the effective node IDs from an explicit flag and URL-parsed IDs.
// If explicitNodeID is non-empty, it wins. Otherwise URL node IDs are used.
// ResolveSingleNodeID resolves one required node scope for a command.
func ResolveSingleNodeID(input *FileInput, explicitNodeID, command string) (string, error) {
	nodeIDs, err := ResolveRequiredNodeIDs(input, explicitNodeID, command)
	if err != nil {
		return "", err
	}
	if len(nodeIDs) != 1 {
		return "", fmt.Errorf("%s requires exactly one node ID", command)
	}
	return nodeIDs[0], nil
}

// ResolveRequiredNodeIDs resolves one or more required node IDs for a command.
func ResolveRequiredNodeIDs(input *FileInput, explicitNodeID, command string) ([]string, error) {
	nodeIDs := ResolveNodeIDs(input, explicitNodeID)
	if len(nodeIDs) == 0 {
		return nil, fmt.Errorf("%s requires a Figma URL with node-id or --id", command)
	}
	return nodeIDs, nil
}

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
