package extract

// Asset is an exportable image, component instance, or standalone vector.
type Asset struct {
	ID     string `json:"node_id"`
	Name   string `json:"name"`
	Kind   string `json:"kind"`
	Format string `json:"format"`
}

// ExtractAssets finds exportable assets in document order. Vectors inside an
// instance are omitted because exporting the instance preserves the full icon.
func ExtractAssets(documents []any) []Asset {
	seen := make(map[string]struct{})
	assets := make([]Asset, 0)
	for _, document := range documents {
		walkAssets(document, false, seen, &assets)
	}
	return assets
}

func walkAssets(value any, insideInstance bool, seen map[string]struct{}, assets *[]Asset) {
	node, ok := value.(map[string]any)
	if !ok {
		return
	}
	id := StringValue(node["id"])
	typeName := StringValue(node["type"])
	isInstance := typeName == "INSTANCE" || typeName == "COMPONENT"
	asset, exportable := assetForNode(node, insideInstance)
	if exportable && id != "" {
		if _, exists := seen[id]; !exists {
			seen[id] = struct{}{}
			*assets = append(*assets, asset)
		}
	}

	children, _ := node["children"].([]any)
	for _, child := range children {
		walkAssets(child, insideInstance || isInstance, seen, assets)
	}
}

func assetForNode(node map[string]any, insideInstance bool) (Asset, bool) {
	id := StringValue(node["id"])
	name := StringValue(node["name"])
	if hasImageFill(node) {
		return Asset{ID: id, Name: name, Kind: "image", Format: "png"}, true
	}
	switch StringValue(node["type"]) {
	case "INSTANCE", "COMPONENT":
		return Asset{ID: id, Name: name, Kind: "instance", Format: "svg"}, true
	case "VECTOR", "BOOLEAN_OPERATION":
		if !insideInstance {
			return Asset{ID: id, Name: name, Kind: "vector", Format: "svg"}, true
		}
	}
	return Asset{}, false
}

func hasImageFill(node map[string]any) bool {
	fills, _ := node["fills"].([]any)
	for _, fill := range fills {
		paint, ok := fill.(map[string]any)
		if ok && StringValue(paint["type"]) == "IMAGE" {
			return true
		}
	}
	return false
}
