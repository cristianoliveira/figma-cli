package figma

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/document"
)

// MapDocument converts the JSON-roundtripped Figma document tree (the
// `any` shape returned by FetchDocument / FetchNodeDocuments) into the
// stable document model. It validates required fields at every node and
// returns a *document.MalformedError with path context for any deviation.
//
// The mapper preserves source order for Children and never silently
// coerces missing required values to empty strings. Empty optional
// fields are normalised: missing/empty Children becomes a non-nil
// empty slice so JSON callers can rely on the array being present.
func MapDocument(value any) (*document.Node, error) {
	if value == nil {
		return nil, &document.MalformedError{Field: "document", Reason: "nil"}
	}
	root, err := mapNode(value, "")
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, &document.MalformedError{Field: "document", Reason: "root value is nil"}
	}
	return root, nil
}

// mapNode handles a single value. Nil values are errors at every level
// except the root optional Children slots. Required fields (id, name,
// type) are validated strictly.
func mapNode(value any, path string) (*document.Node, error) {
	if value == nil {
		return nil, &document.MalformedError{Path: path, Field: "node", Reason: "nil"}
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, &document.MalformedError{Path: path, Field: "node", Reason: fmt.Sprintf("wrong type: %T, expected map", value)}
	}

	id, err := requiredStringField(object, "id", path)
	if err != nil {
		return nil, err
	}
	name, err := requiredStringField(object, "name", path)
	if err != nil {
		return nil, err
	}
	typ, err := requiredStringField(object, "type", path)
	if err != nil {
		return nil, err
	}

	node := &document.Node{
		ID:       id,
		Name:     name,
		Type:     typ,
		Children: []*document.Node{},
	}
	if typ == "TEXT" {
		text, err := optionalStringField(object, "characters", path)
		if err != nil {
			return nil, err
		}
		node.Text = text
	}

	rawChildren, exists := object["children"]
	if exists && rawChildren != nil {
		children, ok := rawChildren.([]any)
		if !ok {
			return nil, &document.MalformedError{Path: path, Field: "children", Reason: fmt.Sprintf("wrong type: %T, expected array", rawChildren)}
		}
		for index, child := range children {
			childPath := joinPath(path, index)
			childNode, err := mapNode(child, childPath)
			if err != nil {
				return nil, err
			}
			if childNode == nil {
				// Defensive: malformed children should have produced an
				// error above. If we somehow get nil, fail loudly.
				return nil, &document.MalformedError{Path: childPath, Field: "node", Reason: "nil after mapping"}
			}
			node.Children = append(node.Children, childNode)
		}
	}

	return node, nil
}

func requiredStringField(object map[string]any, field, path string) (string, error) {
	raw, exists := object[field]
	if !exists {
		return "", &document.MalformedError{Path: path, Field: field, Reason: "missing"}
	}
	if raw == nil {
		return "", &document.MalformedError{Path: path, Field: field, Reason: "null"}
	}
	s, ok := raw.(string)
	if !ok {
		return "", &document.MalformedError{Path: path, Field: field, Reason: fmt.Sprintf("wrong type: %T, expected string", raw)}
	}
	return s, nil
}

func optionalStringField(object map[string]any, field, path string) (string, error) {
	raw, exists := object[field]
	if !exists || raw == nil {
		return "", nil
	}
	s, ok := raw.(string)
	if !ok {
		return "", &document.MalformedError{Path: path, Field: field, Reason: fmt.Sprintf("wrong type: %T, expected string", raw)}
	}
	return s, nil
}

func joinPath(parent string, index int) string {
	if parent == "" {
		return fmt.Sprintf("%d", index)
	}
	return fmt.Sprintf("%s/%d", parent, index)
}
