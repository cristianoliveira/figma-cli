package assetsedge

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/operr"
)

// classifyDownloadStatus returns a neutral classified error for a
// non-200 asset download, retaining only the safe status code. The
// response body is never included: it may contain provider internals
// or secrets.
func classifyDownloadStatus(statusCode int) error {
	return operr.New(operr.CategoryDependencyUnavailable,
		fmt.Sprintf("asset download returned status %d", statusCode),
		"Retry the asset download.",
		nil)
}
