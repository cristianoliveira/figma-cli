package figma

import (
	"context"

	"github.com/cristianoliveira/figma-cli/internal/tokens"
)

// TokensAdapters bundles the Figma-backed fetcher ports for the token
// application. The token service never depends on *figma.Client; this
// adapter binds the client to the consumer-owned ports at the edge.
type TokensAdapters struct {
	Client *Client
}

// NewTokensAdapters builds a Figma-backed set of token fetchers.
func NewTokensAdapters(client *Client) *TokensAdapters {
	return &TokensAdapters{Client: client}
}

// VariablesFetcher returns the /variables/local meta fetcher.
func (a *TokensAdapters) VariablesFetcher() tokens.VariablesFetcher {
	return func(ctx context.Context, fileID string) (map[string]any, error) {
		return FetchVariables(a.Client.WithContext(ctx), fileID)
	}
}

// StylesFetcher returns the published styles fetcher.
func (a *TokensAdapters) StylesFetcher() tokens.StylesFetcher {
	return func(ctx context.Context, fileID string) ([]map[string]any, error) {
		return FetchStyles(a.Client.WithContext(ctx), fileID)
	}
}

// NodesFetcher returns the node-documents fetcher.
func (a *TokensAdapters) NodesFetcher() tokens.NodesFetcher {
	return func(ctx context.Context, fileID string, nodeIDs []string) (map[string]any, error) {
		return FetchNodes(a.Client.WithContext(ctx), fileID, nodeIDs)
	}
}

// DocumentFetcher returns the document-tree fetcher for the scan source.
func (a *TokensAdapters) DocumentFetcher() tokens.DocumentFetcher {
	return func(ctx context.Context, fileID string, nodeIDs []string) (any, error) {
		return FetchDocument(a.Client.WithContext(ctx), fileID, nodeIDs, "", "")
	}
}
