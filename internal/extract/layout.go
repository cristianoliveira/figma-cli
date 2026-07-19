package extract

import "math"

// LayoutNode is a compact Figma tree for understanding layout and copy order.
type LayoutNode struct {
	ID                  string         `json:"id"`
	Name                string         `json:"name"`
	Type                string         `json:"type"`
	Text                string         `json:"text,omitempty"`
	LayoutMode          string         `json:"layoutMode,omitempty"`
	Gap                 *float64       `json:"gap,omitempty"`
	Padding             *LayoutPadding `json:"padding,omitempty"`
	SpacingFromPrevious *LayoutSpacing `json:"spacingFromPrevious,omitempty"`
	Children            []LayoutNode   `json:"children,omitempty"`
}

// LayoutOptions controls optional layout calculations.
type LayoutOptions struct {
	MeasureSpacing bool
}

// LayoutTraversal describes how a depth bound changed a layout tree.
type LayoutTraversal struct {
	ReturnedNodes int  `json:"returnedNodes"`
	TotalNodes    int  `json:"totalNodes"`
	Truncated     bool `json:"truncated"`
	OmittedNodes  int  `json:"omittedNodes"`
}

// LayoutExtraction contains a layout tree and its depth-traversal evidence.
type LayoutExtraction struct {
	Result    LayoutNode
	Traversal LayoutTraversal
}

// LayoutSpacing compares measured sibling distance with declared auto-layout gap.
type LayoutSpacing struct {
	ParentID        string   `json:"parentId,omitempty"`
	PreviousID      string   `json:"previousId"`
	Axis            string   `json:"axis"`
	Measured        float64  `json:"measured"`
	Declared        *float64 `json:"declared,omitempty"`
	MatchesDeclared bool     `json:"matchesDeclared"`
}

// LayoutPadding contains frame padding in Figma pixels.
type LayoutPadding struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

// ExtractLayout converts a Figma subtree into a compact ordered layout tree.
func ExtractLayout(value any, options ...LayoutOptions) LayoutNode {
	object, ok := value.(map[string]any)
	if !ok {
		return LayoutNode{}
	}
	var option LayoutOptions
	if len(options) > 0 {
		option = options[0]
	}
	return extractLayoutNode(object, option)
}

// ExtractLayoutWithDepth converts a layout tree and optionally bounds descendants.
// A negative maxDepth preserves the complete tree. The selected root is depth zero.
func ExtractLayoutWithDepth(value any, maxDepth int, options ...LayoutOptions) LayoutExtraction {
	if _, ok := value.(map[string]any); !ok {
		return LayoutExtraction{}
	}

	full := ExtractLayout(value, options...)
	total := layoutNodeCount(full)
	result := full
	if maxDepth >= 0 {
		result = layoutTreeAtDepth(full, maxDepth)
	}
	returned := layoutNodeCount(result)
	return LayoutExtraction{
		Result: result,
		Traversal: LayoutTraversal{
			ReturnedNodes: returned,
			TotalNodes:    total,
			Truncated:     returned < total,
			OmittedNodes:  total - returned,
		},
	}
}

func layoutTreeAtDepth(node LayoutNode, maxDepth int) LayoutNode {
	if maxDepth == 0 {
		node.Children = nil
		return node
	}

	children := make([]LayoutNode, len(node.Children))
	for index, child := range node.Children {
		children[index] = layoutTreeAtDepth(child, maxDepth-1)
	}
	node.Children = children
	return node
}

func layoutNodeCount(node LayoutNode) int {
	count := 1
	for _, child := range node.Children {
		count += layoutNodeCount(child)
	}
	return count
}

func extractLayoutNode(object map[string]any, option LayoutOptions) LayoutNode {
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
	var previousObject map[string]any
	previousMeaningful := false
	for _, child := range children {
		childObject, ok := child.(map[string]any)
		if !ok {
			previousObject = nil
			previousMeaningful = false
			continue
		}
		childNode := extractLayoutNode(childObject, option)
		meaningful := meaningfulLayoutNode(childNode)
		if meaningful {
			if option.MeasureSpacing && previousMeaningful {
				childNode.SpacingFromPrevious = measureSiblingSpacing(object, previousObject, childObject)
			}
			node.Children = append(node.Children, childNode)
		}
		previousObject = childObject
		previousMeaningful = meaningful
	}
	return node
}

func measureSiblingSpacing(parent, previous, current map[string]any) *LayoutSpacing {
	if previous == nil || previous["visible"] == false || current["visible"] == false {
		return nil
	}
	if StringValue(previous["layoutPositioning"]) == layoutPositioningAbsolute || StringValue(current["layoutPositioning"]) == layoutPositioningAbsolute {
		return nil
	}
	previousBounds, previousOK := layoutBoundsFor(previous)
	currentBounds, currentOK := layoutBoundsFor(current)
	if !previousOK || !currentOK {
		return nil
	}

	var axis string
	var measured float64
	switch StringValue(parent["layoutMode"]) {
	case layoutModeVertical:
		axis = "vertical"
		measured = currentBounds.y - (previousBounds.y + previousBounds.height)
	case layoutModeHorizontal:
		axis = "horizontal"
		measured = currentBounds.x - (previousBounds.x + previousBounds.width)
	default:
		return nil
	}
	measured = normalizeMeasuredSpacing(measured)
	declared := optionalNumber(parent["itemSpacing"])
	if declared == nil {
		defaultGap := 0.0
		declared = &defaultGap
	}
	matches := math.Abs(measured-*declared) < 0.01
	return &LayoutSpacing{PreviousID: StringValue(previous["id"]), Axis: axis, Measured: measured, Declared: declared, MatchesDeclared: matches}
}

func normalizeMeasuredSpacing(value float64) float64 {
	if math.Abs(value) < 0.000001 {
		return 0
	}
	return value
}

type layoutBounds struct {
	x, y, width, height float64
}

func layoutBoundsFor(object map[string]any) (layoutBounds, bool) {
	bounds, ok := object["absoluteBoundingBox"].(map[string]any)
	if !ok {
		return layoutBounds{}, false
	}
	x, xOK := bounds["x"].(float64)
	y, yOK := bounds["y"].(float64)
	width, widthOK := bounds["width"].(float64)
	height, heightOK := bounds["height"].(float64)
	if !xOK || !yOK || !widthOK || !heightOK {
		return layoutBounds{}, false
	}
	return layoutBounds{x: x, y: y, width: width, height: height}, true
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
