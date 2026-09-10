package inspect

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/cristianoliveira/figma-cli/internal/output"
)

// Format selects the rendering mode for recursive inspection.
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// Mode identifies which inspect workflow produced the result.
type Mode string

const (
	ModeSingle    Mode = "single"
	ModeRecursive Mode = "recursive"
	ModeHandoff   Mode = "handoff"
	ModeText      Mode = "text"
)

// Request carries all inputs the inspect workflow needs. The consumer
// (Cobra command or test) fills the fields; the service is responsible
// for the orchestration.
type Request struct {
	Context context.Context

	FileID  string
	NodeID  string
	ScopeID string // optional, defaults to NodeID

	Recursive          bool
	Handoff            bool
	IncludeHidden      bool
	Depth              int // -1 means unbounded
	IncludeVectorPaths bool

	Format Format
	Fields []string

	ResultLimit int
}

// Result is the inspect workflow outcome. The consumer renders it; the
// service does not know about printers.
type Result struct {
	Mode Mode

	Scope     output.Scope
	Single    *extract.InspectOutput
	Nodes     []extract.InspectOutput
	Handoff   *extract.HandoffOutput
	Text      string // populated when Request.Format == FormatText
	Total     int
	Truncated bool
}

// ErrNodeMissing is returned when the inspected node ID is missing from
// the resolved document. It allows the consumer to map the failure to
// the appropriate exit class without leaking transport details.
var ErrNodeMissing = errors.New("inspected node was not returned by the transport")

// Service orchestrates the inspect workflow. It owns fetch, scope
// resolution, extraction, enrichment, limiting, and failure policy.
// Service has no dependency on Cobra or *figma.Client.
type Service struct {
	Nodes NodeFetcher
	Vars  VariableFetcher
}

// New builds a Service with the supplied ports. Both ports are
// required for the inspect workflow to be complete; nil ports will
// cause Inspect to return an error at the call site.
func New(nodes NodeFetcher, vars VariableFetcher) *Service {
	return &Service{Nodes: nodes, Vars: vars}
}

// Inspect drives the full workflow. It is the single entrypoint for
// the Cobra command and for tests.
func (s *Service) Inspect(req Request) (*Result, error) {
	if s.Nodes == nil {
		return nil, errors.New("inspect service: node fetcher is required")
	}
	if req.FileID == "" {
		return nil, errors.New("inspect service: file ID is required")
	}
	if req.NodeID == "" {
		return nil, errors.New("inspect service: node ID is required")
	}
	ctx := req.Context
	if ctx == nil {
		ctx = context.Background()
	}

	depth := ""
	if req.IncludeVectorPaths {
		if req.Recursive {
			depth = strconv.Itoa(req.Depth)
		} else {
			depth = "1"
		}
	}

	details, err := s.Nodes.FetchNodeDetails(ctx, req.FileID, req.NodeID, req.IncludeVectorPaths, depth)
	if err != nil {
		return nil, fmt.Errorf("fetching node details: %w", err)
	}
	if len(details.Documents) == 0 {
		return nil, ErrNodeMissing
	}
	document, ok := details.Documents[0].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("node %s has an invalid document", req.NodeID)
	}
	scope := output.Scope{FileKey: req.FileID, NodeIDs: []string{req.NodeID}}

	switch {
	case req.Recursive:
		return s.inspectRecursive(ctx, req, document, details.Styles, scope)
	case req.Handoff:
		return s.inspectHandoff(ctx, req, document, details.Styles, scope)
	default:
		return s.inspectSingle(ctx, req, document, details.Styles, scope)
	}
}

func (s *Service) inspectSingle(ctx context.Context, req Request, document map[string]any, styles map[string]map[string]any, scope output.Scope) (*Result, error) {
	node := extract.NodeToInspectOutput(document)
	s.enrichSingleNode(ctx, &node, styles, req.FileID)
	return &Result{Mode: ModeSingle, Scope: scope, Single: &node}, nil
}

func (s *Service) inspectRecursive(ctx context.Context, req Request, document map[string]any, styles map[string]map[string]any, scope output.Scope) (*Result, error) {
	var nodes []extract.InspectOutput
	if req.Depth >= 0 {
		nodes = extract.InspectTreeRelativeToScopeToDepth(document, req.NodeID, req.Depth)
	} else {
		nodes = extract.InspectTreeRelativeToScope(document, req.NodeID)
	}
	limited, total := limitResults(req.ResultLimit, nodes)
	s.enrichNodes(ctx, limited, styles, req.FileID)

	result := &Result{
		Mode:      ModeRecursive,
		Scope:     scope,
		Nodes:     limited,
		Total:     total,
		Truncated: len(limited) < total,
	}
	if req.Format == FormatText {
		text, err := extract.FormatInspectText(limited, req.Fields)
		if err != nil {
			return nil, err
		}
		if len(limited) < total {
			text = fmt.Sprintf("Showing %d of %d nodes. Run with --full for all nodes.\n\n%s", len(limited), total, text)
		}
		result.Text = text
	}
	return result, nil
}

func (s *Service) inspectHandoff(ctx context.Context, req Request, document map[string]any, styles map[string]map[string]any, scope output.Scope) (*Result, error) {
	handoff := extract.ExtractHandoff(document, extract.HandoffOptions{
		MaxDepth:      req.Depth,
		IncludeHidden: req.IncludeHidden,
	})
	s.enrichNodes(ctx, handoff.Nodes, styles, req.FileID)
	return &Result{Mode: ModeHandoff, Scope: scope, Nodes: handoff.Nodes, Handoff: &handoff}, nil
}

// enrichNodes resolves style and variable bindings for a slice of nodes.
// Variable enrichment is best-effort: a transport failure here MUST NOT
// abort the result, matching the previous policy.
func (s *Service) enrichNodes(ctx context.Context, nodes []extract.InspectOutput, styles map[string]map[string]any, fileID string) {
	needsVariables := false
	for index := range nodes {
		extract.ResolveInspectStyleBindings(&nodes[index], styles)
		if len(nodes[index].VariableBindings) > 0 {
			needsVariables = true
		}
	}
	if !needsVariables || s.Vars == nil {
		return
	}
	variables, err := s.Vars.FetchVariables(ctx, fileID)
	if err != nil {
		// Best-effort: variable bindings stay unresolved, the rest of
		// the result is still returned.
		return
	}
	for index := range nodes {
		extract.ResolveInspectVariableBindings(&nodes[index], variables)
	}
}

func (s *Service) enrichSingleNode(ctx context.Context, node *extract.InspectOutput, styles map[string]map[string]any, fileID string) {
	nodes := []extract.InspectOutput{*node}
	s.enrichNodes(ctx, nodes, styles, fileID)
	*node = nodes[0]
}

// limitResults caps the slice at limit and returns (capped, total). When
// limit is zero or negative, it returns the input untouched with the
// true total.
func limitResults(limit int, items []extract.InspectOutput) ([]extract.InspectOutput, int) {
	total := len(items)
	if limit <= 0 || total <= limit {
		return items, total
	}
	return items[:limit], total
}
