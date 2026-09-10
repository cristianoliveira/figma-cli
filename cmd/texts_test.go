package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextsCommandEmitsScopedCopySemantics(t *testing.T) {
	client := fixtureClient(t, `{"nodes":{"1:1":{"document":{"id":"1:1","name":"List","type":"FRAME","children":[{"id":"1:2","name":"Items","type":"TEXT","characters":"One\nTwo","lineTypes":["ORDERED","ORDERED"],"lineIndentations":[0,1]}]}}}}`)

	result := executeCommand(newTextsCommand(DepsForLoadClient(func() (*figma.Client, error) { return client, nil })), "https://www.figma.com/design/abc/Name?node-id=1-1")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["1:1"]},"query":{"layer":"","recursive":false},"total":1,"results":[{"id":"1:2","name":"Items","text":"One\nTwo","nodeKind":"textBlock","depth":1,"order":0,"parentName":"List","lines":[{"index":0,"text":"One","listType":"ORDERED"},{"index":1,"text":"Two","listType":"ORDERED","indentation":1}]}]}`, result.Stdout)
}

func TestTextsCommandRejectsMissingScopeBeforeLoadingClient(t *testing.T) {
	loaded := false
	result := executeCommand(newTextsCommand(DepsForLoadClient(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	})), "abc")

	assert.EqualError(t, result.Err, "--layer or a node ID is required")
	assert.False(t, loaded)
}

func TestTextMatchesUsesSelectedURLNodeWithoutLayer(t *testing.T) {
	document := map[string]any{
		"id":   "4707:15504",
		"name": "Selected frame",
		"type": "FRAME",
		"children": []any{
			map[string]any{"id": "4707:15505", "name": "Title", "type": "TEXT", "characters": "Adminless groups"},
		},
	}
	input, err := figma.ParseInput("https://www.figma.com/design/QAhpkgySSOJ6gwJUTB0glb/Adminless-groups?node-id=4707-15504&m=dev")
	require.NoError(t, err)

	result, err := textResult(document, input, "", false)

	require.NoError(t, err)
	texts, ok := result.([]extract.OrderedTextOutput)
	require.True(t, ok)
	require.Len(t, texts, 1)
	assert.Equal(t, "4707:15505", texts[0].ID)
	assert.Equal(t, "Adminless groups", texts[0].Text)
	assert.Equal(t, 1, texts[0].Depth)
	assert.Equal(t, "Selected frame", texts[0].ParentName)
}

func TestTextMatchesRequiresLayerWhenInputHasNoNodeID(t *testing.T) {
	input, err := figma.ParseInput("QAhpkgySSOJ6gwJUTB0glb")
	require.NoError(t, err)

	_, err = textResult(map[string]any{}, input, "", false)

	assert.EqualError(t, err, "--layer or a node ID is required")
}
