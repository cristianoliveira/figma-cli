package figma

import (
	"context"

	"github.com/cristianoliveira/figma-cli/internal/inspect"
)

// InspectAdapter implements inspect.NodeFetcher and inspect.VariableFetcher
// against the Figma API. It is the production wiring; tests use fakes.
type InspectAdapter struct {
	Client *Client
}

var (
	_ inspect.NodeFetcher     = (*InspectAdapter)(nil)
	_ inspect.VariableFetcher = (*InspectAdapter)(nil)
)

// NewInspectAdapter returns an InspectAdapter bound to the supplied client.
func NewInspectAdapter(client *Client) *InspectAdapter {
	return &InspectAdapter{Client: client}
}

// FetchNodeDetails satisfies inspect.NodeFetcher by routing through the
// existing FetchNodeDetails / FetchNodeDetailsWithVectorPaths helpers.
func (a *InspectAdapter) FetchNodeDetails(ctx context.Context, fileID, nodeID string, vectorPaths bool, depth string) (inspect.NodeDetails, error) {
	client := a.Client.WithContext(ctx)
	var (
		details NodeDetails
		err     error
	)
	if vectorPaths {
		details, err = FetchNodeDetailsWithVectorPaths(client, fileID, []string{nodeID}, depth)
	} else {
		details, err = FetchNodeDetails(client, fileID, []string{nodeID})
	}
	if err != nil {
		return inspect.NodeDetails{}, err
	}
	return inspect.NodeDetails{
		Documents: details.Documents,
		Styles:    details.Styles,
	}, nil
}

// FetchVariables satisfies inspect.VariableFetcher.
func (a *InspectAdapter) FetchVariables(ctx context.Context, fileID string) (map[string]any, error) {
	client := a.Client.WithContext(ctx)
	return FetchVariables(client, fileID)
}
