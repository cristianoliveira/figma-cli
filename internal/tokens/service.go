package tokens

import (
	"context"
	"errors"

	"github.com/cristianoliveira/figma-cli/internal/extract"
)

var (
	errFormatterMissing = errors.New("tokens service: formatter is required")
	errSinkMissing      = errors.New("tokens service: artifact sink is required for --output")
)

// Service orchestrates token extraction: resolve sources, format, and
// optionally persist. It has no Cobra, Figma, or filesystem dependency.
type Service struct {
	Policy    Policy
	Formatter Formatter
	Sink      ArtifactSink
}

// New builds a token Service with the given policy, formatter, and sink.
func New(policy Policy, formatter Formatter, sink ArtifactSink) *Service {
	return &Service{Policy: policy, Formatter: formatter, Sink: sink}
}

// Request is the full token workflow input. The command maps flags and
// normalized input into this shape; the service owns the rest.
type Request struct {
	FileID       string
	NodeIDs      []string
	Source       string
	Mode         string
	Format       string
	Prefix       string
	ScanFallback bool
	OutputPath   string
}

// Result is the token workflow outcome, including structured
// diagnostics and the formatted output.
type Result struct {
	Tokens      []extract.Token
	Diagnostics []Diagnostic
	Formatted   string
	OutputPath  string
	Bytes       int
}

// Run resolves tokens, formats them, and (when OutputPath is set)
// persists them. Formatter or sink failures stop cleanly and never
// claim a successful artifact.
func (s *Service) Run(ctx context.Context, req Request) (Result, error) {
	result := Result{}
	tokens, diags, err := s.Policy.Resolve(ctx, SourceRequest{
		FileID:  req.FileID,
		NodeIDs: req.NodeIDs,
		Mode:    req.Mode,
	}, req.Source, req.ScanFallback)
	if err != nil {
		return Result{}, err
	}
	result.Tokens = tokens
	result.Diagnostics = diags

	formatter := s.Formatter
	if formatter == nil {
		return Result{}, errFormatterMissing
	}
	formatted, err := formatter.Format(tokens, req.Format, req.Prefix)
	if err != nil {
		return Result{}, err
	}
	result.Formatted = formatted

	if req.OutputPath != "" {
		if s.Sink == nil {
			return Result{}, errSinkMissing
		}
		if err := s.Sink.Write(ctx, req.OutputPath, []byte(formatted)); err != nil {
			return Result{}, err
		}
		result.OutputPath = req.OutputPath
		result.Bytes = len(formatted)
	}
	return result, nil
}
