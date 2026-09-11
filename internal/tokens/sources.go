package tokens

import (
	"context"

	"github.com/cristianoliveira/figma-cli/internal/extract"
)

// VariablesSource resolves tokens from the /variables/local meta object.
type VariablesSource struct {
	Fetch VariablesFetcher
}

var _ Source = VariablesSource{}

func (s VariablesSource) Name() string { return "variables" }

func (s VariablesSource) Resolve(ctx context.Context, req SourceRequest) ([]extract.Token, error) {
	meta, err := s.Fetch(ctx, req.FileID)
	if err != nil {
		return nil, err
	}
	return extract.ExtractTokensFromVariablesE(meta, req.Mode)
}

// StylesSource resolves tokens from published styles, looking up node
// values for the style node IDs when present.
type StylesSource struct {
	FetchStyles StylesFetcher
	FetchNodes  NodesFetcher
}

var _ Source = StylesSource{}

func (s StylesSource) Name() string { return "styles" }

func (s StylesSource) Resolve(ctx context.Context, req SourceRequest) ([]extract.Token, error) {
	styles, err := s.FetchStyles(ctx, req.FileID)
	if err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	nodeIDs := make([]string, 0, len(styles))
	for _, style := range styles {
		id, _ := style["node_id"].(string)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		nodeIDs = append(nodeIDs, id)
	}
	if len(nodeIDs) == 0 {
		return extract.ExtractTokensFromStyles(styles, nil), nil
	}
	nodes, err := s.FetchNodes(ctx, req.FileID, nodeIDs)
	if err != nil {
		return nil, err
	}
	return extract.ExtractTokensFromStyles(styles, nodes), nil
}

// ScanSource resolves tokens by scanning a document tree for raw fills.
type ScanSource struct {
	FetchDocument DocumentFetcher
}

var _ Source = ScanSource{}

func (s ScanSource) Name() string { return "scan" }

func (s ScanSource) Resolve(ctx context.Context, req SourceRequest) ([]extract.Token, error) {
	doc, err := s.FetchDocument(ctx, req.FileID, req.NodeIDs)
	if err != nil {
		return nil, err
	}
	return extract.ExtractTokensFromDocument(doc), nil
}
