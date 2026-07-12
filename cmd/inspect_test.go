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

func TestInspectCommandRecursivelyEmitsImplementationSpecs(t *testing.T) {
	client := &figma.Client{HTTP: &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		body := `{"nodes":{"42:1":{"document":{"id":"42:1","name":"Button","type":"COMPONENT","absoluteBoundingBox":{"x":100,"y":200,"width":50,"height":40},"componentPropertyDefinitions":{"Disabled":{"type":"BOOLEAN","defaultValue":false}},"children":[{"id":"42:2","name":"Label","type":"TEXT","characters":"Save","absoluteBoundingBox":{"x":112.5,"y":205.25,"width":20,"height":10}}]}}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	result := executeCommand(newInspectCommand(func() (*figma.Client, error) { return client, nil }), "https://www.figma.com/design/abc/Name?node-id=42-1", "--recursive")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["42:1"]},"results":[{"id":"42:1","name":"Button","type":"COMPONENT","propertyDefinitions":{"Disabled":{"type":"BOOLEAN","defaultValue":false}},"bounds":{"x":100,"y":200,"width":50,"height":40},"relativeBounds":{"x":0,"y":0,"width":50,"height":40,"relativeTo":"42:1"},"layout":{},"typography":{}},{"id":"42:2","name":"Label","type":"TEXT","text":"Save","bounds":{"x":112.5,"y":205.25,"width":20,"height":10},"relativeBounds":{"x":12.5,"y":5.25,"width":20,"height":10,"relativeTo":"42:1"},"layout":{},"typography":{}}]}`, result.Stdout)
}

func TestInspectCommandBoundsRecursiveTraversalWithDepth(t *testing.T) {
	client := &figma.Client{HTTP: &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		body := `{"nodes":{"42:1":{"document":{"id":"42:1","name":"Root","type":"FRAME","children":[{"id":"42:2","name":"Child","type":"FRAME","children":[{"id":"42:3","name":"Grandchild","type":"TEXT"}]}]}}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}

	result := executeCommand(newInspectCommand(func() (*figma.Client, error) { return client, nil }), "https://www.figma.com/design/abc/Name?node-id=42-1", "--recursive", "--depth", "1")

	require.NoError(t, result.Err)
	assert.Contains(t, result.Stdout, `"id": "42:2"`)
	assert.NotContains(t, result.Stdout, `"id": "42:3"`)
}

func TestInspectCommandRecursiveEmitsComputedSiblingSpacing(t *testing.T) {
	client := &figma.Client{HTTP: &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		body := `{"nodes":{"42:1":{"document":{"id":"42:1","name":"Stack","type":"FRAME","layoutMode":"VERTICAL","itemSpacing":0,"absoluteBoundingBox":{"x":0,"y":0,"width":100,"height":100},"children":[{"id":"42:2","name":"Copy","type":"TEXT","absoluteBoundingBox":{"x":0,"y":10,"width":80,"height":48}},{"id":"42:3","name":"Link","type":"TEXT","absoluteBoundingBox":{"x":0,"y":58,"width":40,"height":24}}]}}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	result := executeCommand(newInspectCommand(func() (*figma.Client, error) { return client, nil }), "https://www.figma.com/design/abc/Name?node-id=42-1", "--recursive")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["42:1"]},"results":[{"id":"42:1","name":"Stack","type":"FRAME","bounds":{"width":100,"height":100},"relativeBounds":{"x":0,"y":0,"width":100,"height":100,"relativeTo":"42:1"},"layout":{"mode":"VERTICAL"},"typography":{}},{"id":"42:2","name":"Copy","type":"TEXT","bounds":{"y":10,"width":80,"height":48},"relativeBounds":{"x":0,"y":10,"width":80,"height":48,"relativeTo":"42:1"},"layout":{},"typography":{}},{"id":"42:3","name":"Link","type":"TEXT","bounds":{"y":58,"width":40,"height":24},"relativeBounds":{"x":0,"y":58,"width":40,"height":24,"relativeTo":"42:1"},"layout":{},"typography":{},"spacingFromPrevious":{"parentId":"42:1","previousId":"42:2","axis":"vertical","measured":0,"declared":0,"matchesDeclared":true}}]}`, result.Stdout)
}

func TestInspectCommandAcceptsNodeAliasForExplicitID(t *testing.T) {
	client := &figma.Client{HTTP: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Contains(t, request.URL.RawQuery, "ids=0%3A147")
		body := `{"nodes":{"0:147":{"document":{"id":"0:147","name":"Target","type":"FRAME"}}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	result := executeCommand(newInspectCommand(func() (*figma.Client, error) { return client, nil }), "grnVU2vAihHXwYgHryu2xE", "--node", "0:147")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"grnVU2vAihHXwYgHryu2xE","nodeIds":["0:147"]},"result":{"id":"0:147","name":"Target","type":"FRAME","bounds":{},"layout":{},"typography":{}}}`, result.Stdout)
}

func TestInspectCommandRejectsConflictingIDAndNodeAlias(t *testing.T) {
	result := executeCommand(newInspectCommand(func() (*figma.Client, error) { return nil, errors.New("must not load") }), "abc", "--id", "0:147", "--node", "0:148")

	assert.EqualError(t, result.Err, "--id and --node must match when both are provided")
}

func TestInspectCommandRecursiveExplicitIDEmitsRelativeBounds(t *testing.T) {
	client := &figma.Client{HTTP: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		assert.Contains(t, request.URL.RawQuery, "ids=42%3A1")
		body := `{"nodes":{"42:1":{"document":{"id":"42:1","name":"Button","type":"COMPONENT","absoluteBoundingBox":{"x":100,"y":200,"width":50,"height":40},"children":[{"id":"42:2","name":"Label","type":"TEXT","absoluteBoundingBox":{"x":112.5,"y":205.25,"width":20,"height":10}}]}}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	result := executeCommand(newInspectCommand(func() (*figma.Client, error) { return client, nil }), "abc", "--id", "42:1", "--recursive")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["42:1"]},"results":[{"id":"42:1","name":"Button","type":"COMPONENT","bounds":{"x":100,"y":200,"width":50,"height":40},"relativeBounds":{"x":0,"y":0,"width":50,"height":40,"relativeTo":"42:1"},"layout":{},"typography":{}},{"id":"42:2","name":"Label","type":"TEXT","bounds":{"x":112.5,"y":205.25,"width":20,"height":10},"relativeBounds":{"x":12.5,"y":5.25,"width":20,"height":10,"relativeTo":"42:1"},"layout":{},"typography":{}}]}`, result.Stdout)
}

func TestInspectCommandEmitsBoundedHandoff(t *testing.T) {
	client := &figma.Client{HTTP: &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		body := `{"nodes":{"42:1":{"document":{"id":"42:1","name":"Checkout","type":"FRAME","children":[{"id":"42:2","name":"Hidden","type":"TEXT","visible":false},{"id":"42:3","name":"Button","type":"INSTANCE","componentId":"9:1","styles":{"fill":"S:fill"},"children":[{"id":"42:4","name":"Label","type":"TEXT","characters":"Pay"}]},{"id":"42:5","name":"Button","type":"INSTANCE","componentId":"9:1"}]},"styles":{"S:fill":{"name":"Brand/Primary","styleType":"FILL"}}}}}`
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})}}
	result := executeCommand(newInspectCommand(func() (*figma.Client, error) { return client, nil }), "https://www.figma.com/design/abc/Name?node-id=42-1", "--handoff", "--depth", "1")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["42:1"]},"result":{"nodes":[{"id":"42:1","name":"Checkout","type":"FRAME","bounds":{},"layout":{},"typography":{}},{"id":"42:3","name":"Button","type":"INSTANCE","componentId":"9:1","bounds":{},"layout":{},"typography":{},"styleBindings":{"fill":"S:fill"},"resolvedStyles":{"fill":{"id":"S:fill","name":"Brand/Primary","type":"FILL"}}},{"id":"42:5","name":"Button","type":"INSTANCE","componentId":"9:1","bounds":{},"layout":{},"typography":{}}],"components":[{"name":"Button","componentId":"9:1","count":2}]}}`, result.Stdout)
}

func TestInspectCommandRejectsHandoffWithRecursive(t *testing.T) {
	command := newInspectCommand(func() (*figma.Client, error) { return nil, errors.New("must not load") })
	result := executeCommand(command, "https://www.figma.com/design/abc/Name?node-id=42-1", "--handoff", "--recursive")

	assert.EqualError(t, result.Err, "--handoff and --recursive cannot be used together")
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
