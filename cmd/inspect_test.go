package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInspectNodeIDDefaultsToURLNode(t *testing.T) {
	input, err := figma.ParseInput("https://www.figma.com/design/abc/Name?node-id=42-1")
	require.NoError(t, err)

	nodeID, err := inspectNodeID(input, "")

	require.NoError(t, err)
	assert.Equal(t, "42:1", nodeID)
}

func TestInspectNodeIDRequiresScopeForBareFile(t *testing.T) {
	input, err := figma.ParseInput("abc")
	require.NoError(t, err)

	_, err = inspectNodeID(input, "")

	assert.EqualError(t, err, "inspect requires a Figma URL with node-id or --id")
}

func TestInspectNodeIDRejectsMultipleNodes(t *testing.T) {
	input, err := figma.ParseInput("https://www.figma.com/design/abc/Name?node-id=42-1,42-2")
	require.NoError(t, err)

	_, err = inspectNodeID(input, "")

	assert.EqualError(t, err, "inspect requires exactly one node ID")
}
