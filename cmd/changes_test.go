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

func TestChangesCommandEmitsScopedStructuralDiff(t *testing.T) {
	client := &figma.Client{HTTP: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		name := "Old"
		if request.URL.Query().Get("version") == "v2" {
			name = "New"
		}
		body := `{"document":{"id":"0:0","name":"Page","type":"DOCUMENT","children":[{"id":"1:1","name":"` + name + `","type":"FRAME"}]}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}

	result := executeCommand(newChangesCommand(func() (*figma.Client, error) { return client, nil }), "abc", "--from", "v1", "--to", "v2")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":[]},"from":"v1","to":"v2","changes":[{"id":"1:1","path":"Page/New","type":"modified","nodeType":"FRAME","changes":[{"property":"name","from":"Old","to":"New"}]}]}`, result.Stdout)
}

func TestChangesCommandRequiresVersionsBeforeLoadingClient(t *testing.T) {
	loaded := false
	result := executeCommand(newChangesCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}), "abc", "--from", "v1")

	assert.EqualError(t, result.Err, "--from and --to are required")
	assert.False(t, loaded)
}
