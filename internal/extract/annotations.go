package extract

import (
	"fmt"
	"math"

	"github.com/cristianoliveira/figma-cli/internal/annotations"
)

// ExtractCoordinateAnnotations maps a Figma subtree into screenshot-relative, pixel-covering bounds.
func ExtractCoordinateAnnotations(value any, scopeID string, maxDepth int, includeHidden bool) (annotations.Document, error) {
	root, ok := value.(map[string]any)
	if !ok {
		return annotations.Document{}, fmt.Errorf("annotation scope %s is invalid", scopeID)
	}
	scope := boundsFromValue(root["absoluteBoundingBox"])
	width, height := int(math.Ceil(scope.Width)), int(math.Ceil(scope.Height))
	if width < 1 || height < 1 {
		return annotations.Document{}, fmt.Errorf("annotation scope %s must have positive bounds", scopeID)
	}
	result := annotations.Document{Version: 1, CoordinateSpace: annotations.Size{Width: width, Height: height}}
	var walk func(any, int)
	walk = func(value any, depth int) {
		object, ok := value.(map[string]any)
		if !ok || (maxDepth >= 0 && depth > maxDepth) {
			return
		}
		if depth > 0 && !includeHidden && object["visible"] == false {
			return
		}
		boundsValue, hasBounds := object["absoluteBoundingBox"]
		if hasBounds {
			bounds := boundsFromValue(boundsValue)
			left := int(math.Floor(bounds.X - scope.X))
			top := int(math.Floor(bounds.Y - scope.Y))
			right := int(math.Ceil(bounds.X + bounds.Width - scope.X))
			bottom := int(math.Ceil(bounds.Y + bounds.Height - scope.Y))
			if right > left && bottom > top && left >= 0 && top >= 0 && right <= width && bottom <= height {
				result.Annotations = append(result.Annotations, annotations.Annotation{
					ID: StringValue(object["id"]), Label: StringValue(object["name"]),
					Bounds:   annotations.Bounds{X: left, Y: top, Width: right - left, Height: bottom - top},
					Metadata: map[string]any{"source": "figma", "nodeType": StringValue(object["type"])},
				})
			}
		}
		children, _ := object["children"].([]any)
		for _, child := range children {
			walk(child, depth+1)
		}
	}
	walk(root, 0)
	return result, nil
}
