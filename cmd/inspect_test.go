package cmd

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInspectCommandEmitsStableScopedContract(t *testing.T) {
	client := &figma.Client{HTTP: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Contains(t, request.URL.Path, "/v1/files/abc/nodes")
		body := `{"nodes":{"42:1":{"document":{"id":"42:1","name":"Button","type":"COMPONENT","styles":{"fill":"S:fill"}},"styles":{"S:fill":{"key":"key","name":"Brand/Primary","styleType":"FILL","remote":false,"description":""}}}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	command := newInspectCommand(func() (*figma.Client, error) { return client, nil })
	var stdout bytes.Buffer
	command.SetOut(&stdout)
	command.SetArgs([]string{"https://www.figma.com/design/abc/Name?node-id=42-1"})

	require.NoError(t, command.Execute())
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["42:1"]},"result":{"id":"42:1","name":"Button","type":"COMPONENT","bounds":{},"layout":{},"typography":{},"styleBindings":{"fill":"S:fill"},"resolvedStyles":{"fill":{"id":"S:fill","name":"Brand/Primary","type":"FILL"}}}}`, stdout.String())
}

func TestInspectCommandKeepsRawVariablesWhenMetadataIsUnavailable(t *testing.T) {
	client := &figma.Client{HTTP: &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		body := `{"nodes":{"42:1":{"document":{"id":"42:1","name":"Button","type":"COMPONENT","boundVariables":{"fills":[{"type":"VARIABLE_ALIAS","id":"V:brand"}]}}}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	command := newInspectCommandWithVariables(
		func() (*figma.Client, error) { return client, nil },
		func(*figma.Client, string) (map[string]any, error) { return nil, errors.New("forbidden") },
	)
	var stdout bytes.Buffer
	command.SetOut(&stdout)
	command.SetArgs([]string{"https://www.figma.com/design/abc/Name?node-id=42-1"})

	require.NoError(t, command.Execute())
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["42:1"]},"result":{"id":"42:1","name":"Button","type":"COMPONENT","bounds":{},"layout":{},"typography":{},"variableBindings":{"fills":["V:brand"]}}}`, stdout.String())
}

func TestInspectCommandRejectsMissingScopeWithoutLoadingClient(t *testing.T) {
	loaded := false
	command := newInspectCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	})
	command.SilenceErrors = true
	command.SilenceUsage = true
	command.SetArgs([]string{"abc"})

	err := command.Execute()

	assert.EqualError(t, err, "inspect requires a Figma URL with node-id or --id")
	assert.False(t, loaded)
}

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
