// Package extract owns the Figma document-to-output transforms. The
// generic document-tree value coercion helpers (StringValue, number
// helpers, slice/map coercers) moved to internal/document (TASK-0006);
// the thin re-exports below keep the extract API stable while call sites
// migrate.
package extract

import (
	"fmt"
	"math"

	"github.com/cristianoliveira/figma-cli/internal/document"
)

// StringValue extracts a string from an any value, or returns "".
func StringValue(value any) string { return document.StringValue(value) }

// numberValue coerces a numeric field to float64, defaulting to 0.
func numberValue(value any) float64 { return document.NumberValue(value) }

// optionalNumber returns a pointer to a numeric field, or nil if absent.
func optionalNumber(value any) *float64 { return document.OptionalNumber(value) }

// numberSlice returns the numeric contents of an any-array, or nil.
func numberSlice(value any) []float64 { return document.NumberSlice(value) }

// mapValue returns the underlying map or nil.
func mapValue(value any) map[string]any { return document.MapValue(value) }

// colorChannel / colorHexFromPaint / colorsFromPaints are Figma-paint
// helpers that remain in extract; they describe Figma-specific shapes.
func colorChannel(value any) int {
	return int(math.Round(numberValue(value) * 255))
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
