package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommentsCommandEmitsScopedEmptyResults(t *testing.T) {
	client := fixtureClient(t, `{"comments":[]}`)

	result := executeCommand(newCommentsCommand(func() (*figma.Client, error) { return client, nil }), "abc")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":[]},"results":[]}`, result.Stdout)
}

func TestCommentsCommandReturnsAPIErrors(t *testing.T) {
	client := fixtureClientWithStatus(t, 403, `{"message":"forbidden"}`)

	result := executeCommand(newCommentsCommand(func() (*figma.Client, error) { return client, nil }), "abc")

	require.Error(t, result.Err)
	assert.ErrorContains(t, result.Err, "status 403")
}

func TestTokensCommandEmitsJSONWrappedArtifact(t *testing.T) {
	client := fixtureClient(t, `{"document":{"id":"0:0","name":"Document","type":"DOCUMENT","children":[{"id":"1:1","name":"Brand","type":"RECTANGLE","fills":[{"type":"SOLID","color":{"r":1,"g":0,"b":0,"a":1}}]}]}}`)

	result := executeCommand(newTokensCommand(func() (*figma.Client, error) { return client, nil }), "abc", "--source", "scan", "--format", "json", "--json")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"tokens":"{}\n"}`, result.Stdout)
}

func TestTokensCommandRejectsUnsupportedTeamBeforeLoadingClient(t *testing.T) {
	loaded := false
	result := executeCommand(newTokensCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}), "abc", "--team", "wire")

	assert.EqualError(t, result.Err, "--team is not supported yet; pass a file URL or file key")
	assert.False(t, loaded)
}
