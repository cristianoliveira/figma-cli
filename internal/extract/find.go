package extract

import "github.com/cristianoliveira/figma-cli/internal/figma"

// LayerMatch is a found layer, used by `figma find`.
type LayerMatch struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// FindLayersByName returns every node whose name matches layerName exactly.
func FindLayersByName(value any, layerName string) []LayerMatch {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}

	var matches []LayerMatch
	if object["name"] == layerName {
		matches = append(matches, LayerMatch{
			ID:   figma.StringValue(object["id"]),
			Name: figma.StringValue(object["name"]),
			Type: figma.StringValue(object["type"]),
		})
	}

	children, ok := object["children"].([]any)
	if !ok {
		return matches
	}
	for _, child := range children {
		matches = append(matches, FindLayersByName(child, layerName)...)
	}
	return matches
}
