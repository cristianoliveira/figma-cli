package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshalFileResponse(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "fixtures", "file_response.json"))
	require.NoError(t, err)

	var file File
	err = json.Unmarshal(data, &file)
	require.NoError(t, err)

	assert.Equal(t, "Sample Design", file.Name)
	assert.Equal(t, "2025-01-23T08:33:00Z", file.LastModified)
	assert.NotNil(t, file.Document)
	assert.Equal(t, "0:0", file.Document.ID)
	assert.Equal(t, "Document", file.Document.Name)
	assert.Equal(t, "DOCUMENT", file.Document.Type)
	require.Len(t, file.Document.Children, 1)
	canvas := file.Document.Children[0]
	assert.Equal(t, "1:1", canvas.ID)
	assert.Equal(t, "Page 1", canvas.Name)
	assert.Equal(t, "CANVAS", canvas.Type)
	require.Len(t, canvas.Children, 1)
	frame := canvas.Children[0]
	assert.Equal(t, "2:1", frame.ID)
	assert.Equal(t, "Frame 1", frame.Name)
	assert.Equal(t, "FRAME", frame.Type)
	assert.True(t, frame.Visible)
	require.Len(t, frame.Children, 1)
	text := frame.Children[0]
	assert.Equal(t, "3:1", text.ID)
	assert.Equal(t, "Text Layer", text.Name)
	assert.Equal(t, "TEXT", text.Type)
	assert.True(t, text.Visible)
	assert.Equal(t, "Hello, world!", text.Characters)
}

func TestUnmarshalNodeResponse(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "fixtures", "node_response.json"))
	require.NoError(t, err)

	var response struct {
		Nodes map[string]struct {
			Document Node `json:"document"`
		} `json:"nodes"`
		Name string `json:"name"`
	}
	err = json.Unmarshal(data, &response)
	require.NoError(t, err)

	require.Len(t, response.Nodes, 1)
	node, ok := response.Nodes["3:1"]
	require.True(t, ok)
	doc := node.Document
	assert.Equal(t, "3:1", doc.ID)
	assert.Equal(t, "Text Layer", doc.Name)
	assert.Equal(t, "TEXT", doc.Type)
	assert.True(t, doc.Visible)
	assert.Equal(t, "Hello, world!", doc.Characters)
	require.NotNil(t, doc.Style)
	assert.Equal(t, "Inter", doc.Style.FontFamily)
	assert.Equal(t, 600, doc.Style.FontWeight)
	assert.Equal(t, 16.0, doc.Style.FontSize)
	assert.Equal(t, 24.0, doc.Style.LineHeight)
}

func TestRoundTripSerialization(t *testing.T) {
	// Create a sample node and marshal/unmarshal
	node := Node{
		ID:      "test:1",
		Name:    "Test Node",
		Type:    "FRAME",
		Visible: true,
		Children: []Node{
			{
				ID:      "test:2",
				Name:    "Child",
				Type:    "RECTANGLE",
				Visible: true,
				Fills:   []Paint{{Type: "SOLID", Color: &Color{R: 1, G: 0, B: 0, A: 1}}},
				Opacity: 0.5,
			},
		},
		Characters: "text",
		Style: &TextStyle{
			FontFamily: "Roboto",
			FontWeight: 400,
			FontSize:   14,
		},
		Constraints: &LayoutConstraint{
			Vertical:   "TOP",
			Horizontal: "LEFT",
		},
	}
	data, err := json.Marshal(&node)
	require.NoError(t, err)

	var decoded Node
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, node.ID, decoded.ID)
	assert.Equal(t, node.Name, decoded.Name)
	assert.Equal(t, node.Type, decoded.Type)
	assert.Equal(t, node.Visible, decoded.Visible)
	assert.Equal(t, node.Characters, decoded.Characters)
	require.NotNil(t, decoded.Style)
	assert.Equal(t, node.Style.FontFamily, decoded.Style.FontFamily)
	require.NotNil(t, decoded.Constraints)
	assert.Equal(t, node.Constraints.Vertical, decoded.Constraints.Vertical)
	require.Len(t, decoded.Children, 1)
	child := decoded.Children[0]
	assert.Equal(t, "test:2", child.ID)
	require.Len(t, child.Fills, 1)
	assert.Equal(t, "SOLID", child.Fills[0].Type)
	require.NotNil(t, child.Fills[0].Color)
	assert.Equal(t, 1.0, child.Fills[0].Color.R)
}
