package cmd

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/extract"
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
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":[]},"from":"v1","to":"v2","total":1,"truncated":0,"changes":[{"id":"1:1","path":"Page/New","type":"modified","nodeType":"FRAME","changes":[{"property":"name","from":"Old","to":"New"}]}]}`, result.Stdout)
}

func TestPrepareStructuralChangesAppliesTerseAndLimitWithoutMutatingInput(t *testing.T) {
	changes := []extract.StructuralChange{
		{ID: "1", Changes: []extract.PropertyChange{{Property: "name", From: "A", To: "B"}}},
		{ID: "2", Changes: []extract.PropertyChange{{Property: "width", From: 10, To: 20}}},
		{ID: "3"},
	}

	prepared, total, truncated := prepareStructuralChanges(changes, true, 2)

	assert.Equal(t, 3, total)
	assert.Equal(t, 1, truncated)
	require.Len(t, prepared, 2)
	assert.Empty(t, prepared[0].Changes)
	assert.NotEmpty(t, changes[0].Changes)
}

func TestChangesCommandQuietSucceedsWhenChangesExist(t *testing.T) {
	client := &figma.Client{HTTP: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		name := request.URL.Query().Get("version")
		body := `{"document":{"id":"0:0","name":"` + name + `","type":"DOCUMENT"}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}

	result := executeCommand(newChangesCommand(func() (*figma.Client, error) { return client, nil }), "abc", "--from", "v1", "--to", "v2", "--quiet")

	assert.NoError(t, result.Err)
	assert.Empty(t, result.Stdout)
}

func TestChangesCommandQuietUsesGrepStyleExitCodes(t *testing.T) {
	client := fixtureClient(t, `{"document":{"id":"0:0","name":"Page","type":"DOCUMENT"}}`)

	result := executeCommand(newChangesCommand(func() (*figma.Client, error) { return client, nil }), "abc", "--from", "v1", "--to", "v2", "--quiet")

	var exitErr *cli.ExitCodeError
	require.ErrorAs(t, result.Err, &exitErr)
	assert.Equal(t, 1, exitErr.Code)
	assert.Empty(t, result.Stdout)
}

func TestChangesCommandRejectsInvalidLimitBeforeLoadingClient(t *testing.T) {
	loaded := false
	result := executeCommand(newChangesCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}), "abc", "--from", "v1", "--to", "v2", "--limit", "0")

	assert.EqualError(t, result.Err, "--limit must be greater than zero")
	assert.False(t, loaded)
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
