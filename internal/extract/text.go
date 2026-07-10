package extract

import (
	"sort"
	"strings"
)

const textNodeType = "TEXT"

// TextNode represents a TEXT node in a Figma document.
type TextNode struct {
	ID   string
	Name string
	Text string
	Path []string
}

// TextNodeOutput is a JSON-serializable text node.
type TextNodeOutput struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	Text string   `json:"text"`
	Path []string `json:"path,omitempty"`
}

// TextLineOutput preserves Figma-provided line and list intent without inferring HTML semantics.
type TextLineOutput struct {
	Index       int     `json:"index"`
	Text        string  `json:"text"`
	ListType    string  `json:"listType,omitempty"`
	Indentation float64 `json:"indentation,omitempty"`
}

// OrderedTextOutput is copy from a selected frame in Figma tree order.
type OrderedTextOutput struct {
	ID               string                      `json:"id"`
	Name             string                      `json:"name"`
	Text             string                      `json:"text"`
	NodeKind         string                      `json:"nodeKind"`
	Depth            int                         `json:"depth"`
	Order            int                         `json:"order"`
	ParentName       string                      `json:"parentName"`
	Lines            []TextLineOutput            `json:"lines,omitempty"`
	StyleOverrideIDs []int                       `json:"styleOverrideIds,omitempty"`
	StyleOverrides   map[string]typographyOutput `json:"styleOverrides,omitempty"`
}

// ChangedTextOutput represents a changed text node in a diff.
type ChangedTextOutput struct {
	ID   string   `json:"id"`
	Name string   `json:"name"`
	From string   `json:"from"`
	To   string   `json:"to"`
	Path []string `json:"path,omitempty"`
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
	walkTextNodes(value, nil, &nodes)
	return nodes
}

// FindTextByLayerName finds all layers matching the given name and extracts their text.
// If recursive is true, all descendant TEXT nodes are included; otherwise only direct TEXT children.
func FindTextByLayerName(value any, layerName string, recursive bool) []LayerTextOutput {
	var matches []LayerTextOutput
	walkLayers(value, layerName, recursive, &matches)
	return matches
}

// OrderedTextForFrame extracts all descendant copy in Figma tree order.
func OrderedTextForFrame(value any) []OrderedTextOutput {
	root, ok := value.(map[string]any)
	if !ok {
		return nil
	}

	var outputs []OrderedTextOutput
	walkOrderedText(root, 0, "", &outputs)
	return outputs
}

func walkOrderedText(object map[string]any, depth int, parentName string, outputs *[]OrderedTextOutput) {
	if object["type"] == textNodeType {
		*outputs = append(*outputs, OrderedTextOutput{
			ID:               StringValue(object["id"]),
			Name:             StringValue(object["name"]),
			Text:             StringValue(object["characters"]),
			NodeKind:         "textBlock",
			Depth:            depth,
			Order:            len(*outputs),
			ParentName:       parentName,
			Lines:            textLinesFromObject(object),
			StyleOverrideIDs: styleOverrideIDs(object["characterStyleOverrides"]),
			StyleOverrides:   textStyleOverrides(object["styleOverrideTable"]),
		})
	}

	children, ok := object["children"].([]any)
	if !ok {
		return
	}
	for _, child := range children {
		childObject, ok := child.(map[string]any)
		if !ok {
			continue
		}
		walkOrderedText(childObject, depth+1, StringValue(object["name"]), outputs)
	}
}

func textLinesFromObject(object map[string]any) []TextLineOutput {
	lineTypes, hasLineTypes := object["lineTypes"].([]any)
	indentations, hasIndentations := object["lineIndentations"].([]any)
	if !hasLineTypes && !hasIndentations {
		return nil
	}
	hasListIntent := false
	for _, lineType := range lineTypes {
		if value := StringValue(lineType); value != "" && value != "NONE" {
			hasListIntent = true
			break
		}
	}
	if !hasListIntent {
		for _, indentation := range indentations {
			if numberValue(indentation) != 0 {
				hasListIntent = true
				break
			}
		}
	}
	if !hasListIntent {
		return nil
	}
	texts := strings.Split(StringValue(object["characters"]), "\n")
	lines := make([]TextLineOutput, 0, len(texts))
	for index, text := range texts {
		line := TextLineOutput{Index: index, Text: text}
		if index < len(lineTypes) {
			line.ListType = StringValue(lineTypes[index])
			if line.ListType == "NONE" {
				line.ListType = ""
			}
		}
		if index < len(indentations) {
			line.Indentation = numberValue(indentations[index])
		}
		lines = append(lines, line)
	}
	return lines
}

func textStyleOverrides(value any) map[string]typographyOutput {
	table, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	overrides := make(map[string]typographyOutput, len(table))
	for id, style := range table {
		overrides[id] = typographyFromValue(style)
	}
	return overrides
}

func styleOverrideIDs(value any) []int {
	values, ok := value.([]any)
	if !ok {
		return nil
	}
	seen := make(map[int]struct{})
	ids := make([]int, 0)
	for _, value := range values {
		id := int(numberValue(value))
		if id == 0 {
			continue
		}
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
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
			diff.Changed = append(diff.Changed, ChangedTextOutput{ID: id, Name: toNode.Name, From: fromNode.Text, To: toNode.Text, Path: toNode.Path})
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

func walkTextNodes(value any, parentPath []string, nodes *[]TextNode) {
	object, ok := value.(map[string]any)
	if !ok {
		return
	}

	name := StringValue(object["name"])
	if object["type"] == textNodeType {
		id, _ := object["id"].(string)
		text, _ := object["characters"].(string)
		if id != "" {
			*nodes = append(*nodes, TextNode{ID: id, Name: name, Text: text, Path: append([]string(nil), parentPath...)})
		}
	}

	path := append([]string(nil), parentPath...)
	if name != "" && object["type"] != textNodeType {
		path = append(path, name)
	}
	children, ok := object["children"].([]any)
	if !ok {
		return
	}
	for _, child := range children {
		walkTextNodes(child, path, nodes)
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
