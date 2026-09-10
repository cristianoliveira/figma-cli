package assets

import (
	"fmt"
	"strings"
)

// DefaultExportOutputPath builds the default file path for an exported asset.
func DefaultExportOutputPath(fileID string, nodeID string, format string) string {
	return fmt.Sprintf("%s_%s.%s", fileID, strings.ReplaceAll(nodeID, ":", "-"), format)
}
