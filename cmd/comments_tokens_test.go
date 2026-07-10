package cmd

import (
	"io"
	"net/http"
	"strings"
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

func TestCommentsCommandGroupsReviewThreads(t *testing.T) {
	client := fixtureClient(t, `{"comments":[
		{"id":"reply","message":"Fixed","created_at":"2026-01-02T00:00:00Z","file_key":"abc","parent_id":"root","client_meta":{},"reactions":[],"user":{"handle":"Cristian","id":"2","img_url":""}},
		{"id":"root","message":"Adjust spacing","created_at":"2026-01-01T00:00:00Z","file_key":"abc","client_meta":{},"reactions":[],"user":{"handle":"Ada","id":"1","img_url":""}}
	]}`)

	result := executeCommand(newCommentsCommand(func() (*figma.Client, error) { return client, nil }), "abc", "--state", "open", "--author", "cristian")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":[]},"results":[{"root":{"id":"root","message":"Adjust spacing","created_at":"2026-01-01 00:00:00 +0000 UTC","resolved":false,"user":"Ada","url":"https://www.figma.com/design/abc?m=dev#root"},"replies":[{"id":"reply","message":"Fixed","created_at":"2026-01-02 00:00:00 +0000 UTC","resolved":false,"user":"Cristian","parent_id":"root","url":"https://www.figma.com/design/abc?m=dev#reply"}]}]}`, result.Stdout)
}

func TestCommentsCommandSelectsNumericURLFragmentBeforeNodeScope(t *testing.T) {
	client := fixtureClient(t, `{"comments":[
		{"id":"12345","message":"Target","created_at":"2026-01-01T00:00:00Z","file_key":"abc","client_meta":{},"reactions":[],"user":{"handle":"Ada","id":"1","img_url":""}},
		{"id":"other","message":"Ignored","created_at":"2026-01-02T00:00:00Z","file_key":"abc","client_meta":{},"reactions":[],"user":{"handle":"Linus","id":"2","img_url":""}}
	]}`)

	result := executeCommand(newCommentsCommand(func() (*figma.Client, error) { return client, nil }), "https://www.figma.com/design/abc/Name?node-id=1-2#12345")

	require.NoError(t, result.Err)
	assert.Contains(t, result.Stdout, "Target")
	assert.NotContains(t, result.Stdout, "Ignored")
}

func TestCommentsCommandRejectsInvalidStateBeforeLoadingClient(t *testing.T) {
	loaded := false
	result := executeCommand(newCommentsCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}), "abc", "--state", "pending")

	assert.EqualError(t, result.Err, `invalid comment state "pending": expected all, open, or resolved`)
	assert.False(t, loaded)
}

func TestCommentsCommandIncludesAncestorsWithScopedFileRequest(t *testing.T) {
	client := &figma.Client{HTTP: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		var body string
		switch {
		case strings.HasSuffix(request.URL.Path, "/comments"):
			body = `{"comments":[{"id":"root","message":"Parent feedback","created_at":"2026-01-01T00:00:00Z","file_key":"abc","client_meta":{"node_id":"1:1","node_offset":{"x":0,"y":0}},"reactions":[],"user":{"handle":"Ada","id":"1","img_url":""}}]}`
		case strings.HasSuffix(request.URL.Path, "/nodes"):
			body = `{"nodes":{"1:2":{"document":{"id":"1:2","name":"Button","type":"FRAME"}}}}`
		default:
			assert.Equal(t, "1:2", request.URL.Query().Get("ids"))
			body = `{"document":{"id":"0:0","name":"Document","type":"DOCUMENT","children":[{"id":"1:1","name":"Screen","type":"FRAME","children":[{"id":"1:2","name":"Button","type":"FRAME"}]}]}}`
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}

	result := executeCommand(newCommentsCommand(func() (*figma.Client, error) { return client, nil }), "https://www.figma.com/design/abc/Name?node-id=1-2", "--include-ancestors")

	require.NoError(t, result.Err)
	assert.Contains(t, result.Stdout, "Parent feedback")
	assert.Contains(t, result.Stdout, "Screen")
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
