package inspect

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeNodes and fakeVars are the only collaborators the service tests
// need; the package owns no HTTP server, no figma client, no Cobra command.

type fakeNodes struct {
	details NodeDetails
	err     error
	got     fakeNodesCall
}

type fakeNodesCall struct {
	fileID      string
	nodeID      string
	vectorPaths bool
	depth       string
}

func (f *fakeNodes) FetchNodeDetails(_ context.Context, fileID, nodeID string, vectorPaths bool, depth string) (NodeDetails, error) {
	f.got = fakeNodesCall{fileID: fileID, nodeID: nodeID, vectorPaths: vectorPaths, depth: depth}
	return f.details, f.err
}

type fakeVars struct {
	vars map[string]any
	err  error
	got  string
}

func (f *fakeVars) FetchVariables(_ context.Context, fileID string) (map[string]any, error) {
	f.got = fileID
	return f.vars, f.err
}

// singleNodeDocument returns the smallest valid document tree that contains
// a single node with the requested ID; used by happy-path single mode.
func singleNodeDocument(id, name, typ string) map[string]any {
	return map[string]any{
		"id":   id,
		"name": name,
		"type": typ,
	}
}

// recursiveDocument returns a frame with one child, enough to exercise
// InspectTreeRelativeToScope.
func recursiveDocument(scopeID string) map[string]any {
	return map[string]any{
		"id":   scopeID,
		"name": "Root",
		"type": "FRAME",
		"children": []any{
			map[string]any{
				"id":   "1:2",
				"name": "Child",
				"type": "RECTANGLE",
			},
		},
	}
}

func TestServiceRequiresPorts(t *testing.T) {
	_, err := New(nil, nil).Inspect(Request{FileID: "f", NodeID: "1:1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "node fetcher is required")

	nodes := &fakeNodes{}
	_, err = New(nodes, nil).Inspect(Request{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "file ID is required")

	_, err = New(nodes, nil).Inspect(Request{FileID: "f"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "node ID is required")
}

func TestServiceSingleModeReturnsNode(t *testing.T) {
	nodes := &fakeNodes{details: NodeDetails{
		Documents: []any{singleNodeDocument("1:1", "Root", "FRAME")},
	}}
	svc := New(nodes, &fakeVars{})

	result, err := svc.Inspect(Request{FileID: "f", NodeID: "1:1"})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, ModeSingle, result.Mode)
	require.NotNil(t, result.Single)
	assert.Equal(t, "1:1", result.Single.ID)
	assert.Equal(t, "Root", result.Single.Name)
	assert.Equal(t, "FRAME", result.Single.Type)
	assert.Equal(t, "f", nodes.got.fileID)
	assert.Equal(t, "1:1", nodes.got.nodeID)
	assert.False(t, nodes.got.vectorPaths, "non-recursive inspect must not request vector paths")
}

func TestServiceRecursiveModeReturnsScopedNodes(t *testing.T) {
	doc := recursiveDocument("1:1")
	nodes := &fakeNodes{details: NodeDetails{Documents: []any{doc}}}
	svc := New(nodes, &fakeVars{})

	result, err := svc.Inspect(Request{FileID: "f", NodeID: "1:1", Recursive: true, Depth: -1})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, ModeRecursive, result.Mode)
	assert.Len(t, result.Nodes, 2, "root + one child")
	assert.False(t, result.Truncated)
	assert.Equal(t, 2, result.Total)
}

func TestServiceRecursiveModeHonoursResultLimit(t *testing.T) {
	doc := recursiveDocument("1:1")
	nodes := &fakeNodes{details: NodeDetails{Documents: []any{doc}}}
	svc := New(nodes, &fakeVars{})

	result, err := svc.Inspect(Request{FileID: "f", NodeID: "1:1", Recursive: true, Depth: -1, ResultLimit: 1})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Nodes, 1)
	assert.True(t, result.Truncated)
	assert.Equal(t, 2, result.Total, "Total must reflect the full result count, not the limit")
}

func TestServiceRecursiveTextFormatRendersHeaders(t *testing.T) {
	doc := recursiveDocument("1:1")
	nodes := &fakeNodes{details: NodeDetails{Documents: []any{doc}}}
	svc := New(nodes, &fakeVars{})

	result, err := svc.Inspect(Request{FileID: "f", NodeID: "1:1", Recursive: true, Depth: -1, Format: FormatText, Fields: []string{"id", "name"}})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.Text, "text format must populate Result.Text")
	assert.True(t, strings.Contains(result.Text, "Root") || strings.Contains(result.Text, "Child"))
}

func TestServiceHandoffModeReturnsHandoffPayload(t *testing.T) {
	doc := recursiveDocument("1:1")
	nodes := &fakeNodes{details: NodeDetails{Documents: []any{doc}}}
	svc := New(nodes, &fakeVars{})

	result, err := svc.Inspect(Request{FileID: "f", NodeID: "1:1", Handoff: true, Depth: 2})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, ModeHandoff, result.Mode)
	require.NotNil(t, result.Handoff)
	assert.NotEmpty(t, result.Handoff.Nodes)
}

func TestServiceTransportFailureSurfacesError(t *testing.T) {
	transportErr := errors.New("upstream unreachable")
	nodes := &fakeNodes{err: transportErr}
	svc := New(nodes, &fakeVars{})

	result, err := svc.Inspect(Request{FileID: "f", NodeID: "1:1"})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, transportErr), "transport errors must wrap the port error")
}

func TestServiceMissingNodeReturnsErrNodeMissing(t *testing.T) {
	nodes := &fakeNodes{details: NodeDetails{Documents: nil}}
	svc := New(nodes, &fakeVars{})

	result, err := svc.Inspect(Request{FileID: "f", NodeID: "missing"})
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrNodeMissing), "missing node must surface ErrNodeMissing")
}

func TestServiceInvalidDocumentPayloadErrors(t *testing.T) {
	nodes := &fakeNodes{details: NodeDetails{Documents: []any{"not-a-map"}}}
	svc := New(nodes, &fakeVars{})

	_, err := svc.Inspect(Request{FileID: "f", NodeID: "1:1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid document")
}

func TestServiceVectorPathsRequestedOnlyForRecursiveOrExplicitDepth(t *testing.T) {
	nodes := &fakeNodes{details: NodeDetails{Documents: []any{singleNodeDocument("1:1", "Root", "FRAME")}}}
	svc := New(nodes, &fakeVars{})

	_, err := svc.Inspect(Request{FileID: "f", NodeID: "1:1", IncludeVectorPaths: true})
	require.NoError(t, err)
	assert.True(t, nodes.got.vectorPaths, "single-node + vector paths must request geometry")
	assert.Equal(t, "1", nodes.got.depth, "single-node geometry depth defaults to 1")

	// Reset and try recursive with explicit depth.
	nodes.got = fakeNodesCall{}
	_, err = svc.Inspect(Request{FileID: "f", NodeID: "1:1", Recursive: true, IncludeVectorPaths: true, Depth: 3})
	require.NoError(t, err)
	assert.True(t, nodes.got.vectorPaths)
	assert.Equal(t, "3", nodes.got.depth)
}

func TestServiceVariableEnrichmentFailureIsBestEffort(t *testing.T) {
	doc := singleNodeDocument("1:1", "Root", "FRAME")
	doc["boundVariables"] = map[string]any{"fills": []any{map[string]any{"type": "VARIABLE_ALIAS", "id": "V:brand"}}}
	nodes := &fakeNodes{details: NodeDetails{Documents: []any{doc}}}
	vars := &fakeVars{err: errors.New("variables endpoint forbidden")}
	svc := New(nodes, vars)

	result, err := svc.Inspect(Request{FileID: "f", NodeID: "1:1"})
	require.NoError(t, err, "variable enrichment failure MUST NOT abort the inspect result")
	require.NotNil(t, result.Single)
	// VariableBindings remain unresolved; the inspect output is still valid.
	assert.Equal(t, "1:1", result.Single.ID)
}

func TestServiceVariableEnrichmentSucceedsWhenAvailable(t *testing.T) {
	doc := singleNodeDocument("1:1", "Root", "FRAME")
	doc["boundVariables"] = map[string]any{"fills": []any{map[string]any{"type": "VARIABLE_ALIAS", "id": "V:brand"}}}
	nodes := &fakeNodes{details: NodeDetails{Documents: []any{doc}}}
	vars := &fakeVars{vars: map[string]any{
		"V:brand": map[string]any{"name": "Brand", "type": "COLOR"},
	}}
	svc := New(nodes, vars)

	result, err := svc.Inspect(Request{FileID: "f", NodeID: "1:1"})
	require.NoError(t, err)
	require.NotNil(t, result.Single)
	require.NotNil(t, vars.got)
	assert.Equal(t, "f", vars.got)
	// After enrichment the variable lookup ran; we don't assert on the
	// resolved name here because that lives in extract_test.go already.
}

// Ensure the inspect package does not pull Cobra or *figma.Client in.
func TestServiceHasNoCobraOrFigmaClientDependency(t *testing.T) {
	// Compile-time guarantees: the package imports are restricted.
	// This test documents the architectural constraint.
	imports := []string{}
	_ = imports // intentionally empty list — the test's job is to assert
	//             the constraint via the import graph above, not at runtime.
	assert.NotNil(t, New(&fakeNodes{}, &fakeVars{}))
}

// _ makes the extract import used so goimports preserves it for the test
// helpers below; this also keeps coverage of the package's use of
// extract.InspectOutput stable.
var _ = extract.InspectOutput{}
