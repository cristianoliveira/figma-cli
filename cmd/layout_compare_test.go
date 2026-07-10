package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLayoutCompareCommandUsesExplicitOrderedFrames(t *testing.T) {
	client := fixtureClient(t, `{"nodes":{
		"1:1":{"document":{"id":"1:1","name":"Desktop","type":"FRAME","layoutMode":"HORIZONTAL","itemSpacing":24,"absoluteBoundingBox":{"width":1440,"height":900}}},
		"2:1":{"document":{"id":"2:1","name":"Mobile","type":"FRAME","layoutMode":"VERTICAL","itemSpacing":12,"absoluteBoundingBox":{"width":375,"height":800}}}
	}}`)

	result := executeCommand(newLayoutCompareCommand(func() (*figma.Client, error) { return client, nil }), "abc", "--id", "1-1", "--id", "2:1")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["1:1","2:1"]},"variants":[{"id":"1:1","name":"Desktop","width":1440,"height":900,"mode":"HORIZONTAL","gap":24,"padding":{"top":0,"right":0,"bottom":0,"left":0},"css":{"display":"flex","flexDirection":"row","gap":24}},{"id":"2:1","name":"Mobile","width":375,"height":800,"mode":"VERTICAL","gap":12,"padding":{"top":0,"right":0,"bottom":0,"left":0},"css":{"display":"flex","flexDirection":"column","gap":12}}],"transitions":[{"fromId":"1:1","toId":"2:1","changes":[{"property":"width","from":1440,"to":375},{"property":"height","from":900,"to":800},{"property":"mode","from":"HORIZONTAL","to":"VERTICAL"},{"property":"gap","from":24,"to":12},{"property":"css.flexDirection","from":"row","to":"column"}]}]}`, result.Stdout)
}

func TestLayoutCompareCommandRequiresTwoFramesBeforeLoadingClient(t *testing.T) {
	loaded := false
	result := executeCommand(newLayoutCompareCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}), "abc", "--id", "1:1")

	assert.EqualError(t, result.Err, "layout compare requires at least two --id or --name values")
	assert.False(t, loaded)
}
