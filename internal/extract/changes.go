package extract

import (
	"reflect"
	"sort"
	"strings"
)

// PropertyChange is one curated design property changed between versions.
type PropertyChange struct {
	Property string `json:"property"`
	From     any    `json:"from,omitempty"`
	To       any    `json:"to,omitempty"`
}

// StructuralChange describes one added, removed, or modified Figma node.
type StructuralChange struct {
	ID       string           `json:"id"`
	Path     string           `json:"path"`
	Type     string           `json:"type"`
	NodeType string           `json:"nodeType"`
	Changes  []PropertyChange `json:"changes,omitempty"`
}

type indexedDocumentNode struct {
	object map[string]any
	path   string
}

// DiffDocuments compares stable node IDs and curated frontend-relevant properties.
func DiffDocuments(from, to any) []StructuralChange {
	fromNodes := indexDocumentNodes(from)
	toNodes := indexDocumentNodes(to)
	changes := make([]StructuralChange, 0)

	for id, toNode := range toNodes {
		fromNode, exists := fromNodes[id]
		if !exists {
			changes = append(changes, structuralNodeChange(id, toNode, "added"))
			continue
		}
		properties := changedNodeProperties(fromNode.object, toNode.object)
		if len(properties) > 0 {
			change := structuralNodeChange(id, toNode, "modified")
			change.Changes = properties
			changes = append(changes, change)
		}
	}
	for id, fromNode := range fromNodes {
		if _, exists := toNodes[id]; !exists {
			changes = append(changes, structuralNodeChange(id, fromNode, "removed"))
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Path == changes[j].Path {
			return changes[i].ID < changes[j].ID
		}
		return changes[i].Path < changes[j].Path
	})
	return changes
}

func structuralNodeChange(id string, node indexedDocumentNode, changeType string) StructuralChange {
	return StructuralChange{
		ID:       id,
		Path:     node.path,
		Type:     changeType,
		NodeType: StringValue(node.object["type"]),
	}
}

func indexDocumentNodes(value any) map[string]indexedDocumentNode {
	nodes := make(map[string]indexedDocumentNode)
	indexDocumentNode(value, nil, nodes)
	return nodes
}

func indexDocumentNode(value any, parentPath []string, nodes map[string]indexedDocumentNode) {
	object, ok := value.(map[string]any)
	if !ok {
		return
	}
	path := append([]string(nil), parentPath...)
	if name := StringValue(object["name"]); name != "" {
		path = append(path, name)
	}
	if id := StringValue(object["id"]); id != "" {
		nodes[id] = indexedDocumentNode{object: object, path: strings.Join(path, "/")}
	}
	children, _ := object["children"].([]any)
	for _, child := range children {
		indexDocumentNode(child, path, nodes)
	}
}

func changedNodeProperties(from, to map[string]any) []PropertyChange {
	properties := []struct {
		name string
		from any
		to   any
	}{
		{"name", StringValue(from["name"]), StringValue(to["name"])},
		{"type", StringValue(from["type"]), StringValue(to["type"])},
		{"componentId", StringValue(from["componentId"]), StringValue(to["componentId"])},
		{"componentSetId", StringValue(from["componentSetId"]), StringValue(to["componentSetId"])},
		{"layout.mode", StringValue(from["layoutMode"]), StringValue(to["layoutMode"])},
		{"layout.gap", numberValue(from["itemSpacing"]), numberValue(to["itemSpacing"])},
		{"layout.wrap", StringValue(from["layoutWrap"]), StringValue(to["layoutWrap"])},
		{"layout.paddingTop", numberValue(from["paddingTop"]), numberValue(to["paddingTop"])},
		{"layout.paddingRight", numberValue(from["paddingRight"]), numberValue(to["paddingRight"])},
		{"layout.paddingBottom", numberValue(from["paddingBottom"]), numberValue(to["paddingBottom"])},
		{"layout.paddingLeft", numberValue(from["paddingLeft"]), numberValue(to["paddingLeft"])},
		{"layout.sizingHorizontal", StringValue(from["layoutSizingHorizontal"]), StringValue(to["layoutSizingHorizontal"])},
		{"layout.sizingVertical", StringValue(from["layoutSizingVertical"]), StringValue(to["layoutSizingVertical"])},
		{"fills", colorsFromPaints(from["fills"]), colorsFromPaints(to["fills"])},
		{"strokes", colorsFromPaints(from["strokes"]), colorsFromPaints(to["strokes"])},
		{"opacity", numberValue(from["opacity"]), numberValue(to["opacity"])},
		{"cornerRadius", numberValue(from["cornerRadius"]), numberValue(to["cornerRadius"])},
	}
	changes := make([]PropertyChange, 0)
	for _, property := range properties {
		if !reflect.DeepEqual(property.from, property.to) {
			changes = append(changes, PropertyChange{Property: property.name, From: property.from, To: property.to})
		}
	}
	changes = append(changes, changedStyleProperties(from["styles"], to["styles"])...)
	changes = append(changes, changedBoundsProperties(from["absoluteBoundingBox"], to["absoluteBoundingBox"])...)
	fromChildren := documentChildIDs(from)
	toChildren := documentChildIDs(to)
	if !reflect.DeepEqual(fromChildren, toChildren) && sameStringSet(fromChildren, toChildren) {
		changes = append(changes, PropertyChange{Property: "childrenOrder", From: fromChildren, To: toChildren})
	}
	return changes
}

func changedStyleProperties(fromValue, toValue any) []PropertyChange {
	from, _ := fromValue.(map[string]any)
	to, _ := toValue.(map[string]any)
	keys := make(map[string]struct{}, len(from)+len(to))
	for key := range from {
		keys[key] = struct{}{}
	}
	for key := range to {
		keys[key] = struct{}{}
	}
	names := make([]string, 0, len(keys))
	for key := range keys {
		names = append(names, key)
	}
	sort.Strings(names)
	changes := make([]PropertyChange, 0)
	for _, name := range names {
		fromID := StringValue(from[name])
		toID := StringValue(to[name])
		if fromID != toID {
			changes = append(changes, PropertyChange{Property: "styles." + name, From: fromID, To: toID})
		}
	}
	return changes
}

func changedBoundsProperties(fromValue, toValue any) []PropertyChange {
	from, _ := fromValue.(map[string]any)
	to, _ := toValue.(map[string]any)
	changes := make([]PropertyChange, 0)
	for _, name := range []string{"x", "y", "width", "height"} {
		fromNumber := numberValue(from[name])
		toNumber := numberValue(to[name])
		if fromNumber != toNumber {
			changes = append(changes, PropertyChange{Property: "bounds." + name, From: fromNumber, To: toNumber})
		}
	}
	return changes
}

func documentChildIDs(object map[string]any) []string {
	children, _ := object["children"].([]any)
	ids := make([]string, 0, len(children))
	for _, child := range children {
		childObject, _ := child.(map[string]any)
		if id := StringValue(childObject["id"]); id != "" {
			ids = append(ids, id)
		}
	}
	return ids
}

func sameStringSet(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	counts := make(map[string]int, len(left))
	for _, value := range left {
		counts[value]++
	}
	for _, value := range right {
		counts[value]--
		if counts[value] < 0 {
			return false
		}
	}
	return true
}
