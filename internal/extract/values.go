// Package extract owns the Figma document-to-output transforms.
// Generic document-tree value coercion lives in internal/document; this
// file retains only Figma-paint-specific helpers.
package extract

import (
	"fmt"
	"math"

	"github.com/cristianoliveira/figma-cli/internal/document"
)

// colorChannel converts a normalized Figma channel into an 8-bit value.
func colorChannel(value any) int {
	return int(math.Round(document.NumberValue(value) * 255))
}

// colorHexFromPaint renders a Figma color object to a #RRGGBB hex string.
func colorHexFromPaint(paint map[string]any) string {
	color, ok := paint["color"].(map[string]any)
	if !ok {
		return ""
	}
	return fmt.Sprintf("#%02X%02X%02X", colorChannel(color["r"]), colorChannel(color["g"]), colorChannel(color["b"]))
}

// colorsFromPaints collects hex colors from a paint array, skipping hidden paints.
func colorsFromPaints(value any) []string {
	paints, ok := value.([]any)
	if !ok {
		return nil
	}
	colors := make([]string, 0, len(paints))
	for _, paint := range paints {
		paintObject, ok := paint.(map[string]any)
		if !ok || paintObject["visible"] == false {
			continue
		}
		color := colorHexFromPaint(paintObject)
		if color != "" {
			colors = append(colors, color)
		}
	}
	return colors
}

// paintTypeSolid is a Figma-paint constant retained by the extract
// package; paint classification belongs here, not in the generic
// document tree.
const paintTypeSolid = "SOLID"
