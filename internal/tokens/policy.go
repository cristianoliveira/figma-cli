package tokens

import (
	"context"
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/extract"
)

// Source identifiers.
const (
	SourceAuto      = "auto"
	SourceVariables = "variables"
	SourceStyles    = "styles"
	SourceScan      = "scan"
)

// Deterministic note messages (rendered with a "note: " prefix by the
// command; kept here so the policy owns the exact wording).
const (
	noteScanFallback  = "no Styles/Variables found; scanning document nodes (tokens named by value). Use --scan-fallback=false for named-only."
	noteScanDisabled  = "no Styles/Variables found and scan disabled; drop --scan-fallback=false (or use --source scan) to extract from raw fills"
	unavailableSource = "%s source failed"
)

// Policy is the explicit source selection and fallback policy. It owns
// Variables → Styles → optional scan ordering. Adding a new source
// means adding a field + wiring it here and in the command's policy
// construction — not editing transport, formatter, or Cobra code.
type Policy struct {
	Variables Source
	Styles    Source
	Scan      Source
}

// Resolve selects the requested source and returns tokens plus any
// structured diagnostics. Explicit sources never fall back: a provider
// failure is returned as an error. Auto falls back from Variables to
// Styles, then optionally to scan, recording diagnostics for provider
// failures and empty/disabled notes along the way.
func (p Policy) Resolve(ctx context.Context, req SourceRequest, source string, scanFallback bool) ([]extract.Token, []Diagnostic, error) {
	switch source {
	case "", SourceAuto:
		return p.resolveAuto(ctx, req, scanFallback)
	case SourceVariables:
		if p.Variables == nil {
			return nil, nil, fmt.Errorf("variables source is not configured")
		}
		tokens, err := p.Variables.Resolve(ctx, req)
		if err != nil {
			return nil, nil, err
		}
		return tokens, nil, nil
	case SourceStyles:
		if p.Styles == nil {
			return nil, nil, fmt.Errorf("styles source is not configured")
		}
		tokens, err := p.Styles.Resolve(ctx, req)
		if err != nil {
			return nil, nil, err
		}
		return tokens, nil, nil
	case SourceScan:
		if p.Scan == nil {
			return nil, nil, fmt.Errorf("scan source is not configured")
		}
		tokens, err := p.Scan.Resolve(ctx, req)
		if err != nil {
			return nil, nil, err
		}
		return tokens, nil, nil
	default:
		return nil, nil, fmt.Errorf("unknown --source %q (want variables, styles, scan, or auto)", source)
	}
}

// resolveAuto implements Variables → Styles → optional scan.
func (p Policy) resolveAuto(ctx context.Context, req SourceRequest, scanFallback bool) ([]extract.Token, []Diagnostic, error) {
	var diags []Diagnostic

	if p.Variables != nil {
		tokens, err := p.Variables.Resolve(ctx, req)
		if err != nil {
			diags = append(diags, Diagnostic{Severity: "warning", Message: fmt.Sprintf(unavailableSource, SourceVariables) + ": " + err.Error()})
		} else if len(tokens) > 0 {
			return tokens, diags, nil
		}
	}

	if p.Styles != nil {
		tokens, err := p.Styles.Resolve(ctx, req)
		if err != nil {
			diags = append(diags, Diagnostic{Severity: "warning", Message: fmt.Sprintf(unavailableSource, SourceStyles) + ": " + err.Error()})
		} else if len(tokens) > 0 {
			return tokens, diags, nil
		}
	}

	if scanFallback {
		diags = append(diags, Diagnostic{Severity: "note", Message: noteScanFallback})
		if p.Scan == nil {
			return nil, diags, fmt.Errorf("scan source is not configured")
		}
		tokens, err := p.Scan.Resolve(ctx, req)
		if err != nil {
			return nil, diags, err
		}
		return tokens, diags, nil
	}

	diags = append(diags, Diagnostic{Severity: "note", Message: noteScanDisabled})
	return nil, diags, nil
}
