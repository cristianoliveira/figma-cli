package extract

// LayoutNode is a compact Figma tree for understanding layout and copy order.
type LayoutNode struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	Text       string         `json:"text,omitempty"`
	LayoutMode string         `json:"layoutMode,omitempty"`
	Gap        *float64       `json:"gap,omitempty"`
	Padding    *LayoutPadding `json:"padding,omitempty"`
	Children   []LayoutNode   `json:"children,omitempty"`
}

// LayoutPadding contains frame padding in Figma pixels.
type LayoutPadding struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

// ExtractLayout converts a Figma subtree into a compact ordered layout tree.
func ExtractLayout(value any) LayoutNode {
	object, ok := value.(map[string]any)
	if !ok {
		return LayoutNode{}
	}
	return extractLayoutNode(object)
}

func extractLayoutNode(object map[string]any) LayoutNode {
	node := LayoutNode{
		ID:         StringValue(object["id"]),
		Name:       StringValue(object["name"]),
		Type:       StringValue(object["type"]),
		LayoutMode: StringValue(object["layoutMode"]),
		Gap:        optionalNumber(object["itemSpacing"]),
		Padding:    layoutPadding(object),
	}
	if node.Type == textNodeType {
		node.Text = StringValue(object["characters"])
	}

	children, ok := object["children"].([]any)
	if !ok {
		return node
	}
	for _, child := range children {
		childObject, ok := child.(map[string]any)
		if !ok {
			continue
		}
		childNode := extractLayoutNode(childObject)
		if !meaningfulLayoutNode(childNode) {
			continue
		}
		node.Children = append(node.Children, childNode)
	}
	return node
}

func meaningfulLayoutNode(node LayoutNode) bool {
	return node.Type == textNodeType || node.LayoutMode != "" || node.Padding != nil || len(node.Children) > 0
}

func layoutPadding(object map[string]any) *LayoutPadding {
	top := numberValue(object["paddingTop"])
	right := numberValue(object["paddingRight"])
	bottom := numberValue(object["paddingBottom"])
	left := numberValue(object["paddingLeft"])
	if top == 0 && right == 0 && bottom == 0 && left == 0 {
		return nil
	}
	return &LayoutPadding{Top: top, Right: right, Bottom: bottom, Left: left}
}
