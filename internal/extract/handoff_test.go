package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractHandoffBoundsDepthExcludesHiddenAndSummarizesInstances(t *testing.T) {
	document := map[string]any{
		"id": "1:1", "name": "Checkout", "type": "FRAME",
		"children": []any{
			map[string]any{"id": "1:2", "name": "Hidden", "type": "TEXT", "visible": false, "characters": "Ignore"},
			map[string]any{"id": "1:3", "name": "Button", "type": "INSTANCE", "componentId": "9:1", "componentProperties": map[string]any{"Size": map[string]any{"value": "Large"}}, "children": []any{
				map[string]any{"id": "1:4", "name": "Label", "type": "TEXT", "characters": "Pay", "children": []any{
					map[string]any{"id": "1:5", "name": "Too deep", "type": "TEXT", "characters": "No"},
				}},
			}},
			map[string]any{"id": "1:6", "name": "Button", "type": "INSTANCE", "componentId": "9:1"},
		},
	}

	handoff := ExtractHandoff(document, HandoffOptions{MaxDepth: 2})

	assert.Equal(t, []string{"1:1", "1:3", "1:4", "1:6"}, inspectIDs(handoff.Nodes))
	assert.Equal(t, []ComponentUsage{{Name: "Button", ComponentID: "9:1", Count: 2}}, handoff.Components)
}

func TestExtractHandoffCanIncludeHiddenNodes(t *testing.T) {
	document := map[string]any{"id": "1:1", "name": "Root", "type": "FRAME", "children": []any{
		map[string]any{"id": "1:2", "name": "Hidden", "type": "TEXT", "visible": false},
	}}

	handoff := ExtractHandoff(document, HandoffOptions{MaxDepth: 1, IncludeHidden: true})

	assert.Equal(t, []string{"1:1", "1:2"}, inspectIDs(handoff.Nodes))
}

func inspectIDs(nodes []InspectOutput) []string {
	ids := make([]string, 0, len(nodes))
	for _, node := range nodes {
		ids = append(ids, node.ID)
	}
	return ids
}
