package extract

import "strings"

// LayerMatch is a found layer, used by `figma find`.
type LayerMatch struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// SearchCriteria filters nodes by name and/or type.
// Name is a case-insensitive substring; empty matches any name.
// Type is a case-insensitive exact node type; empty matches any type.
type SearchCriteria struct {
	Name string
	Type string
}

// Search walks a document tree and returns every node matching the criteria.
// With both Name and Type empty it returns all nodes in the tree.
func Search(value any, criteria SearchCriteria) []LayerMatch {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}

	nameLower := strings.ToLower(criteria.Name)
	typeLower := strings.ToLower(criteria.Type)

	var matches []LayerMatch
	if nodeMatches(object, nameLower, typeLower) {
		matches = append(matches, LayerMatch{
			ID:   StringValue(object["id"]),
			Name: StringValue(object["name"]),
			Type: StringValue(object["type"]),
		})
	}

	children, ok := object["children"].([]any)
	if !ok {
		return matches
	}
	for _, child := range children {
		matches = append(matches, Search(child, criteria)...)
	}
	return matches
}

// nodeMatches reports whether a node satisfies the lower-cased criteria.
func nodeMatches(object map[string]any, nameLower, typeLower string) bool {
	if nameLower != "" && !strings.Contains(strings.ToLower(StringValue(object["name"])), nameLower) {
		return false
	}
	if typeLower != "" && strings.ToLower(StringValue(object["type"])) != typeLower {
		return false
	}
	return true
}
