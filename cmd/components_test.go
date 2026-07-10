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
	command.SetArgs([]string{"https://www.figma.com/design/abc/Name?node-id=42-1", "--name", "button"})

	require.NoError(t, command.Execute())
	assert.JSONEq(t, `{
		"scope":{"fileKey":"abc","nodeIds":["42:1"]},
		"results":[{"id":"42:2","name":"Button","type":"INSTANCE","componentId":"1:1","paints":{},"bounds":{},"layout":{},"typography":{}}]
	}`, stdout.String())
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
