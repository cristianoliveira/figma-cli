// Package inspect owns the inspect application workflow. It exposes a
// context-aware request/result API that the Cobra command (or any other
// consumer) can drive without depending on a concrete Figma client.
//
// The service depends on narrow ports defined here (NodeFetcher and
// VariableFetcher). The Figma adapter that lives in internal/figma
// implements those ports; tests inject fakes.
package inspect

import "context"

// NodeDetails is the minimal payload the inspect service needs from the
// transport layer. Documents is typically []map[string]any; Styles maps
// style IDs to their published metadata.
type NodeDetails struct {
	Documents []any
	Styles    map[string]map[string]any
}

// NodeFetcher abstracts fetching node documents and their style
// metadata. Implementations may translate this into HTTP calls against
// the Figma API or a stub for tests.
//
// depth is the bounded depth string Figma expects; pass "" for the
// default. vectorPaths enables exact fill/stroke geometry responses.
type NodeFetcher interface {
	FetchNodeDetails(ctx context.Context, fileID, nodeID string, vectorPaths bool, depth string) (NodeDetails, error)
}

// VariableFetcher abstracts fetching the Variables collection metadata
// used for variable-bindings enrichment. Variable enrichment is
// best-effort; the service tolerates a nil return without surfacing an
// error to the caller.
type VariableFetcher interface {
	FetchVariables(ctx context.Context, fileID string) (map[string]any, error)
}
