// Package tokens owns the design-token extraction workflow and source
// selection policy. It has no dependency on Cobra, *figma.Client, or
// the filesystem; transport and persistence are injected through the
// narrow ports below.
package tokens

import (
	"context"

	"github.com/cristianoliveira/figma-cli/internal/extract"
)

// VariablesFetcher fetches the /variables/local meta object.
type VariablesFetcher func(ctx context.Context, fileID string) (map[string]any, error)

// StylesFetcher fetches published style metadata.
type StylesFetcher func(ctx context.Context, fileID string) ([]map[string]any, error)

// NodesFetcher fetches node documents keyed by node id (used to resolve
// published style values).
type NodesFetcher func(ctx context.Context, fileID string, nodeIDs []string) (map[string]any, error)

// DocumentFetcher fetches a document tree (used by the scan source).
type DocumentFetcher func(ctx context.Context, fileID string, nodeIDs []string) (any, error)

// Source resolves tokens from one backing store.
type Source interface {
	// Name returns the source identifier ("variables", "styles", "scan").
	Name() string
	// Resolve returns the tokens for the request. An empty result means
	// the source produced no tokens (not an error).
	Resolve(ctx context.Context, req SourceRequest) ([]extract.Token, error)
}

// SourceRequest is the subset of request state a Source needs.
type SourceRequest struct {
	FileID  string
	NodeIDs []string
	Mode    string
}

// Formatter formats tokens into an output string (CSS, Tailwind, JSON).
type Formatter interface {
	Format(tokens []extract.Token, format, prefix string) (string, error)
}

// FormatterFunc adapts a function to the Formatter interface.
type FormatterFunc func([]extract.Token, string, string) (string, error)

var _ Formatter = FormatterFunc(nil)

func (f FormatterFunc) Format(tokens []extract.Token, format, prefix string) (string, error) {
	return f(tokens, format, prefix)
}

// ArtifactSink persists formatted bytes to a path.
type ArtifactSink interface {
	Write(ctx context.Context, path string, data []byte) error
}

// ArtifactSinkFunc adapts a function to the ArtifactSink interface.
type ArtifactSinkFunc func(ctx context.Context, path string, data []byte) error

var _ ArtifactSink = ArtifactSinkFunc(nil)

func (f ArtifactSinkFunc) Write(ctx context.Context, path string, data []byte) error {
	return f(ctx, path, data)
}

// Diagnostic is a structured message produced by the source policy. It
// is returned as data so callers can render or log it; the application
// never writes it as a hidden side effect.
type Diagnostic struct {
	Severity string // "note" or "warning"
	Message  string
}
