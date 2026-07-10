package extract

// HandoffOptions controls the implementation-focused node selection.
type HandoffOptions struct {
	MaxDepth      int
	IncludeHidden bool
}

// ComponentUsage summarizes repeated component instances in a handoff.
type ComponentUsage struct {
	Name         string `json:"name"`
	ComponentID  string `json:"componentId,omitempty"`
	ComponentSet string `json:"componentSetId,omitempty"`
	Count        int    `json:"count"`
}

// HandoffOutput combines bounded implementation specs with component usage.
type HandoffOutput struct {
	Nodes      []InspectOutput  `json:"nodes"`
	Components []ComponentUsage `json:"components,omitempty"`
}

// ExtractHandoff creates bounded, implementation-focused output from a node tree.
func ExtractHandoff(value any, options HandoffOptions) HandoffOutput {
	if options.MaxDepth < 0 {
		options.MaxDepth = 0
	}
	nodes := make([]InspectOutput, 0)
	components := make([]ComponentUsage, 0)
	componentIndexes := make(map[string]int)

	var walk func(any, int)
	walk = func(value any, depth int) {
		object, ok := value.(map[string]any)
		if !ok || depth > options.MaxDepth {
			return
		}
		if depth > 0 && !options.IncludeHidden && object["visible"] == false {
			return
		}

		node := NodeToInspectOutput(object)
		nodes = append(nodes, node)
		if node.Type == componentTypeInstance {
			key := node.ComponentID
			if key == "" {
				key = node.Name
			}
			if index, exists := componentIndexes[key]; exists {
				components[index].Count++
			} else {
				componentIndexes[key] = len(components)
				components = append(components, ComponentUsage{
					Name:         node.Name,
					ComponentID:  node.ComponentID,
					ComponentSet: node.ComponentSetID,
					Count:        1,
				})
			}
		}

		children, _ := object["children"].([]any)
		for _, child := range children {
			walk(child, depth+1)
		}
	}

	walk(value, 0)
	return HandoffOutput{Nodes: nodes, Components: components}
}
