package extract

import (
	"fmt"
	"math"
)

// numberValue coerces a Figma numeric field to float64, defaulting to 0.
func numberValue(value any) float64 {
	number, ok := value.(float64)
	if !ok {
		return 0
	}
	return number
}

// optionalNumber returns a pointer to a numeric field, or nil if absent.
func optionalNumber(value any) *float64 {
	number, ok := value.(float64)
	if !ok {
		return nil
	}
	return &number
}

func numberSlice(value any) []float64 {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	numbers := make([]float64, 0, len(items))
	for _, item := range items {
		numbers = append(numbers, numberValue(item))
	}
	return numbers
}

func mapValue(value any) map[string]any {
	object, _ := value.(map[string]any)
	return object
}

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
