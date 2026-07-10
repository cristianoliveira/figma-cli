package extract

import "sort"

const textNodeType = "TEXT"

// TextNode represents a TEXT node in a Figma document.
type TextNode struct {
	ID   string
	Name string
	Text string
}

// TextNodeOutput is a JSON-serializable text node.
type TextNodeOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Text string `json:"text"`
}

// ChangedTextOutput represents a changed text node in a diff.
type ChangedTextOutput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	From string `json:"from"`
	To   string `json:"to"`
}

// TextOutput is the result of a text diff between two versions.
type TextOutput struct {
	Added   []TextNodeOutput    `json:"added"`
	Removed []TextNodeOutput    `json:"removed"`
	Changed []ChangedTextOutput `json:"changed"`
}

// LayerTextOutput represents a layer with its text content.
type LayerTextOutput struct {
	ID    string           `json:"id"`
	Name  string           `json:"name"`
	Type  string           `json:"type"`
	Texts []TextNodeOutput `json:"texts"`
}

// ExtractTextNodes walks a document tree and collects all TEXT nodes.
func ExtractTextNodes(value any) []TextNode {
	var nodes []TextNode
	walkTextNodes(value, &nodes)
	return nodes
}

// FindTextByLayerName finds all layers matching the given name and extracts their text.
// If recursive is true, all descendant TEXT nodes are included; otherwise only direct TEXT children.
func FindTextByLayerName(value any, layerName string, recursive bool) []LayerTextOutput {
	var matches []LayerTextOutput
	walkLayers(value, layerName, recursive, &matches)
	return matches
}

// TextEqual reports whether two text-node sets carry identical text by node
// ID. It is the equality predicate used by blame's binary search: a version
// "has the new text" when its text set equals the target version's.
func TextEqual(a, b []TextNode) bool {
	d := DiffText(a, b)
	return len(d.Added) == 0 && len(d.Removed) == 0 && len(d.Changed) == 0
}

// DiffText computes added, removed, and changed text nodes between two sets.
func DiffText(from, to []TextNode) TextOutput {
	fromByID := map[string]TextNode{}
	toByID := map[string]TextNode{}

	for _, node := range from {
		fromByID[node.ID] = node
	}
	for _, node := range to {
		toByID[node.ID] = node
	}

	var diff TextOutput
	for id, fromNode := range fromByID {
		toNode, ok := toByID[id]
		if !ok {
			diff.Removed = append(diff.Removed, TextNodeOutput(fromNode))
			continue
		}
		if fromNode.Text != toNode.Text {
			diff.Changed = append(diff.Changed, ChangedTextOutput{ID: id, Name: toNode.Name, From: fromNode.Text, To: toNode.Text})
		}
	}
	for id, toNode := range toByID {
		if _, ok := fromByID[id]; !ok {
			diff.Added = append(diff.Added, TextNodeOutput(toNode))
		}
	}

	sort.Slice(diff.Added, func(i, j int) bool { return diff.Added[i].ID < diff.Added[j].ID })
	sort.Slice(diff.Removed, func(i, j int) bool { return diff.Removed[i].ID < diff.Removed[j].ID })
	sort.Slice(diff.Changed, func(i, j int) bool { return diff.Changed[i].ID < diff.Changed[j].ID })
	return diff
}

func walkTextNodes(value any, nodes *[]TextNode) {
	object, ok := value.(map[string]any)
	if !ok {
		return
	}

	if object["type"] == textNodeType {
		id, _ := object["id"].(string)
		name, _ := object["name"].(string)
		text, _ := object["characters"].(string)
		if id != "" {
			*nodes = append(*nodes, TextNode{ID: id, Name: name, Text: text})
		}
	}

	children, ok := object["children"].([]any)
	if !ok {
		return
	}
	for _, child := range children {
		walkTextNodes(child, nodes)
	}
}

func walkLayers(value any, layerName string, recursive bool, matches *[]LayerTextOutput) {
	object, ok := value.(map[string]any)
	if !ok {
		return
	}

	if object["name"] == layerName {
		*matches = append(*matches, LayerTextOutput{
			ID:    StringValue(object["id"]),
			Name:  StringValue(object["name"]),
			Type:  StringValue(object["type"]),
			Texts: textOutputsForLayer(object, recursive),
		})
	}

	children, ok := object["children"].([]any)
	if !ok {
		return
	}
	for _, child := range children {
		walkLayers(child, layerName, recursive, matches)
	}
}

func textOutputsForLayer(object map[string]any, recursive bool) []TextNodeOutput {
	if object["type"] == textNodeType {
		return []TextNodeOutput{{ID: StringValue(object["id"]), Name: StringValue(object["name"]), Text: StringValue(object["characters"])}}
	}
	if recursive {
		return textNodeOutputs(ExtractTextNodes(object))
	}

	var nodes []TextNode
	children, ok := object["children"].([]any)
	if !ok {
		return nil
	}
	for _, child := range children {
		childObject, ok := child.(map[string]any)
		if !ok || childObject["type"] != textNodeType {
			continue
		}
		nodes = append(nodes, TextNode{ID: StringValue(childObject["id"]), Name: StringValue(childObject["name"]), Text: StringValue(childObject["characters"])})
	}
	return textNodeOutputs(nodes)
}

func textNodeOutputs(nodes []TextNode) []TextNodeOutput {
	outputs := make([]TextNodeOutput, 0, len(nodes))
	for _, node := range nodes {
		outputs = append(outputs, TextNodeOutput(node))
	}
	return outputs
}
