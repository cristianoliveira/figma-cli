package figma

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/document"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapDocumentNestedRootChildrenGrandchild(t *testing.T) {
	raw := map[string]any{
		"id": "0:0", "name": "Root", "type": "FRAME",
		"children": []any{
			map[string]any{
				"id": "1:1", "name": "Child", "type": "FRAME",
				"children": []any{
					map[string]any{"id": "1:2", "name": "Grandchild", "type": "RECTANGLE"},
				},
			},
		},
	}
	root, err := MapDocument(raw)
	require.NoError(t, err)
	require.NotNil(t, root)
	assert.Equal(t, "0:0", root.ID)
	assert.Equal(t, "Root", root.Name)
	assert.Equal(t, "FRAME", root.Type)
	require.Len(t, root.Children, 1)
	assert.Equal(t, "1:1", root.Children[0].ID)
	require.Len(t, root.Children[0].Children, 1)
	assert.Equal(t, "1:2", root.Children[0].Children[0].ID)
	assert.Equal(t, "RECTANGLE", root.Children[0].Children[0].Type)
}

func TestMapDocumentTextNodeCarriesCharacters(t *testing.T) {
	raw := map[string]any{
		"id": "0:0", "name": "Page", "type": "FRAME",
		"children": []any{
			map[string]any{
				"id": "1:1", "name": "Title", "type": "TEXT",
				"characters": "Hello",
			},
		},
	}
	root, err := MapDocument(raw)
	require.NoError(t, err)
	require.Len(t, root.Children, 1)
	assert.Equal(t, "1:1", root.Children[0].ID)
	assert.Equal(t, "TEXT", root.Children[0].Type)
	assert.Equal(t, "Hello", root.Children[0].Text)
}

func TestMapDocumentNonTextNodeLeavesTextEmpty(t *testing.T) {
	raw := map[string]any{
		"id": "1:1", "name": "Vector", "type": "VECTOR",
	}
	root, err := MapDocument(raw)
	require.NoError(t, err)
	assert.Equal(t, "", root.Text)
}

func TestMapDocumentMissingOptionalChildrenNormalisesToEmpty(t *testing.T) {
	raw := map[string]any{
		"id": "1:1", "name": "Lonely", "type": "FRAME",
	}
	root, err := MapDocument(raw)
	require.NoError(t, err)
	require.NotNil(t, root.Children)
	assert.Empty(t, root.Children, "missing children must normalise to an empty (non-nil) slice")
}

func TestMapDocumentMissingOptionalCharactersForTextNodeStaysEmpty(t *testing.T) {
	raw := map[string]any{
		"id": "1:1", "name": "Empty", "type": "TEXT",
	}
	root, err := MapDocument(raw)
	require.NoError(t, err)
	assert.Equal(t, "", root.Text)
}

func TestMapDocumentPreservesSourceOrder(t *testing.T) {
	raw := map[string]any{
		"id": "0:0", "name": "Root", "type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:1", "name": "First", "type": "FRAME"},
			map[string]any{"id": "1:2", "name": "Second", "type": "FRAME"},
			map[string]any{"id": "1:3", "name": "Third", "type": "FRAME"},
		},
	}
	root, err := MapDocument(raw)
	require.NoError(t, err)
	require.Len(t, root.Children, 3)
	assert.Equal(t, []string{"1:1", "1:2", "1:3"}, []string{
		root.Children[0].ID, root.Children[1].ID, root.Children[2].ID,
	})
}

func TestMapDocumentUnknownFigmaTypeIsAcceptedAsString(t *testing.T) {
	// Future Figma types should not break the adapter; the stable model
	// carries Type as a free-form string so a stable consumer can still
	// search/inspect it.
	raw := map[string]any{
		"id": "1:1", "name": "Future", "type": "FUTURE_NODE_TYPE_9000",
	}
	root, err := MapDocument(raw)
	require.NoError(t, err)
	assert.Equal(t, "FUTURE_NODE_TYPE_9000", root.Type)
}

func TestMapDocumentMalformedMissingRequiredID(t *testing.T) {
	raw := map[string]any{
		"name": "Root", "type": "FRAME",
	}
	_, err := MapDocument(raw)
	require.Error(t, err)
	malformed, ok := err.(*document.MalformedError)
	require.True(t, ok, "expected *MalformedError, got %T", err)
	assert.Equal(t, "id", malformed.Field)
	assert.Equal(t, "missing", malformed.Reason)
	assert.Equal(t, "", malformed.Path, "root path is empty")
}

func TestMapDocumentMalformedRequiredIDWrongType(t *testing.T) {
	raw := map[string]any{
		"id": 42, "name": "Root", "type": "FRAME",
	}
	_, err := MapDocument(raw)
	require.Error(t, err)
	malformed, ok := err.(*document.MalformedError)
	require.True(t, ok, "expected *MalformedError")
	assert.Equal(t, "id", malformed.Field)
	assert.Contains(t, malformed.Reason, "wrong type")
}

func TestMapDocumentMalformedDescendantReportsPath(t *testing.T) {
	raw := map[string]any{
		"id": "0:0", "name": "Root", "type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:1", "name": "Child", "type": "FRAME"},
			map[string]any{
				"id": "1:2", "name": "Bad", "type": "FRAME",
				"children": []any{
					map[string]any{
						"id": "1:3",
						// missing name
						"type": "RECTANGLE",
					},
				},
			},
		},
	}
	_, err := MapDocument(raw)
	require.Error(t, err)
	malformed, ok := err.(*document.MalformedError)
	require.True(t, ok)
	assert.Equal(t, "1/0", malformed.Path, "root's 2nd child (index 1) -> its 1st grandchild (index 0)")
	assert.Equal(t, "name", malformed.Field)
}

func TestMapDocumentMalformedChildrenWrongType(t *testing.T) {
	raw := map[string]any{
		"id": "0:0", "name": "Root", "type": "FRAME",
		"children": "not-an-array",
	}
	_, err := MapDocument(raw)
	require.Error(t, err)
	malformed, ok := err.(*document.MalformedError)
	require.True(t, ok)
	assert.Equal(t, "children", malformed.Field)
	assert.Contains(t, malformed.Reason, "wrong type")
}

func TestMapDocumentMalformedChildNilEntry(t *testing.T) {
	raw := map[string]any{
		"id": "0:0", "name": "Root", "type": "FRAME",
		"children": []any{nil},
	}
	_, err := MapDocument(raw)
	require.Error(t, err)
	malformed, ok := err.(*document.MalformedError)
	require.True(t, ok)
	assert.Equal(t, "0", malformed.Path)
}

func TestMapDocumentNilRootErrors(t *testing.T) {
	_, err := MapDocument(nil)
	require.Error(t, err)
	malformed, ok := err.(*document.MalformedError)
	require.True(t, ok)
	assert.Equal(t, "document", malformed.Field)
}

func TestMapDocumentNonMapRootErrors(t *testing.T) {
	_, err := MapDocument("not-a-map")
	require.Error(t, err)
	malformed, ok := err.(*document.MalformedError)
	require.True(t, ok)
	assert.Equal(t, "node", malformed.Field)
}

func TestMapDocumentNullRequiredField(t *testing.T) {
	raw := map[string]any{
		"id": nil, "name": "Root", "type": "FRAME",
	}
	_, err := MapDocument(raw)
	require.Error(t, err)
	malformed, ok := err.(*document.MalformedError)
	require.True(t, ok)
	assert.Equal(t, "id", malformed.Field)
	assert.Equal(t, "null", malformed.Reason)
}
