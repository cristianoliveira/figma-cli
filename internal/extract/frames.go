package extract

// FrameMatch identifies a screen-level frame and its containing page or section.
type FrameMatch struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ParentID   string `json:"parentId"`
	ParentName string `json:"parentName"`
}

// DiscoverFrames returns frames directly contained by a canvas or section.
// It descends through nested sections but not through frames, avoiding layout
// frames inside each discovered screen.
func DiscoverFrames(value any) []FrameMatch {
	root, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	if isFrameContainer(root) {
		return discoverFramesInContainer(root)
	}
	if StringValue(root["type"]) != "DOCUMENT" {
		return nil
	}

	var frames []FrameMatch
	children, _ := root["children"].([]any)
	for _, childValue := range children {
		child, ok := childValue.(map[string]any)
		if !ok || StringValue(child["type"]) != "CANVAS" {
			continue
		}
		frames = append(frames, discoverFramesInContainer(child)...)
	}
	return frames
}

func discoverFramesInContainer(container map[string]any) []FrameMatch {
	children, ok := container["children"].([]any)
	if !ok {
		return nil
	}

	var frames []FrameMatch
	for _, childValue := range children {
		child, ok := childValue.(map[string]any)
		if !ok {
			continue
		}

		switch StringValue(child["type"]) {
		case "FRAME":
			frames = append(frames, FrameMatch{
				ID:         StringValue(child["id"]),
				Name:       StringValue(child["name"]),
				ParentID:   StringValue(container["id"]),
				ParentName: StringValue(container["name"]),
			})
		case "SECTION":
			frames = append(frames, discoverFramesInContainer(child)...)
		}
	}
	return frames
}

func isFrameContainer(object map[string]any) bool {
	nodeType := StringValue(object["type"])
	return nodeType == "CANVAS" || nodeType == "SECTION"
}
