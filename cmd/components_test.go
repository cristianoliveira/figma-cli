package cmd

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComponentsCommandEmitsStableScopedResults(t *testing.T) {
	client := &figma.Client{HTTP: &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		body := `{"nodes":{"42:1":{"document":{"id":"42:1","name":"Screen","type":"FRAME","children":[{"id":"42:2","name":"Button","type":"INSTANCE","componentId":"1:1"}]}}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	command := newComponentsCommand(func() (*figma.Client, error) { return client, nil })
	var stdout bytes.Buffer
	command.SetOut(&stdout)
	command.SetArgs([]string{"https://www.figma.com/design/abc/Name?node-id=42-1", "--name", "button", "--kind", "instance"})

	require.NoError(t, command.Execute())
	assert.JSONEq(t, `{
		"scope":{"fileKey":"abc","nodeIds":["42:1"]},
		"results":[{"id":"42:2","name":"Button","type":"INSTANCE","componentId":"1:1","path":["Screen","Button"],"paints":{},"bounds":{},"layout":{},"typography":{}}]
	}`, stdout.String())
}

func TestComponentsCommandGroupsUsageByExactComponentID(t *testing.T) {
	client := &figma.Client{HTTP: &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		body := `{"nodes":{"42:1":{"document":{"id":"42:1","name":"Screen","type":"FRAME","children":[{"id":"42:2","name":"Primary","type":"INSTANCE","componentId":"1:1","variantProperties":{"Size":"Large"}},{"id":"42:3","name":"Secondary","type":"INSTANCE","componentId":"1:1"},{"id":"42:4","name":"Primary","type":"INSTANCE","componentId":"2:1"}]}}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	result := executeCommand(newComponentsCommand(func() (*figma.Client, error) { return client, nil }), "https://www.figma.com/design/abc/Name?node-id=42-1", "--usage")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["42:1"]},"results":[{"componentId":"1:1","name":"Primary","count":2,"instances":[{"id":"42:2","name":"Primary","path":["Screen","Primary"],"variantProperties":{"Size":"Large"}},{"id":"42:3","name":"Secondary","path":["Screen","Secondary"]}]},{"componentId":"2:1","name":"Primary","count":1,"instances":[{"id":"42:4","name":"Primary","path":["Screen","Primary"]}]}]}`, result.Stdout)
}

func TestComponentsCommandRejectsUsageWithRawWithoutLoadingClient(t *testing.T) {
	loaded := false
	command := newComponentsCommand(func() (*figma.Client, error) { loaded = true; return nil, nil })
	result := executeCommand(command, "https://www.figma.com/design/abc/Name?node-id=42-1", "--usage", "--raw")

	assert.EqualError(t, result.Err, "--usage and --raw cannot be used together")
	assert.False(t, loaded)
}

func TestComponentsCommandRejectsInvalidKindWithoutLoadingClient(t *testing.T) {
	loaded := false
	command := newComponentsCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	})
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs([]string{"--kind", "frame", "https://www.figma.com/design/abc/Name?node-id=42-1"})

	err := command.Execute()

	assert.EqualError(t, err, `invalid component kind "frame": expected component, set, or instance`)
	assert.False(t, loaded)
}

func TestComponentsCommandRejectsMissingScopeWithoutLoadingClient(t *testing.T) {
	loaded := false
	command := newComponentsCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	})
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs([]string{"abc"})

	err := command.Execute()

	assert.EqualError(t, err, "components requires a Figma URL with node-id or --id")
	assert.False(t, loaded)
}
