package extract

import (
	"strings"

	"github.com/cristianoliveira/figma-cli/internal/document"
)

// LayerMatch is a found layer, used by `figma find`.
type LayerMatch struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

// SearchCriteria filters nodes by name and/or type.
// Name is a case-insensitive substring; empty matches any name.
// Type is a case-insensitive exact node type; empty matches any type.
type SearchCriteria struct {
	Name string
	Type string
}

// Search walks the stable document model and returns every node
// matching the criteria. With both Name and Type empty it returns all
// nodes in the tree, in pre-order depth-first order. A nil document
// returns no matches.
//
// This function does not traverse `map[string]any`; it operates
// exclusively on the stable model produced by `figma.MapDocument`.
func Search(root *document.Node, criteria SearchCriteria) []LayerMatch {
	if root == nil {
		return nil
	}

	nameLower := strings.ToLower(criteria.Name)
	typeLower := strings.ToLower(criteria.Type)

	var matches []LayerMatch
	if nodeMatches(root, nameLower, typeLower) {
		matches = append(matches, LayerMatch{
			ID:   root.ID,
			Name: root.Name,
			Type: root.Type,
			Text: textCharacters(root),
		})
	}
	for _, child := range root.Children {
		matches = append(matches, Search(child, criteria)...)
	}
	return matches
}

func textCharacters(node *document.Node) string {
	if node.Type != textNodeType {
		return ""
	}
	return node.Text
}

func nodeMatches(node *document.Node, nameLower, typeLower string) bool {
	if nameLower != "" && !strings.Contains(strings.ToLower(node.Name), nameLower) {
		return false
	}
	if typeLower != "" && !strings.EqualFold(node.Type, typeLower) {
		return false
	}
	return true
}
