package tokens

import (
	"context"
	"errors"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/extract"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSource records calls and returns a fixed token set / error.
type fakeSource struct {
	name   string
	tokens []extract.Token
	err    error
	calls  int
	gotCtx context.Context
	gotReq SourceRequest
}

func (f *fakeSource) Name() string { return f.name }
func (f *fakeSource) Resolve(ctx context.Context, req SourceRequest) ([]extract.Token, error) {
	f.calls++
	f.gotCtx = ctx
	f.gotReq = req
	return f.tokens, f.err
}

func tok(id string) extract.Token { return extract.Token{Path: []string{id}, Value: id} }

func tokensOf(ids ...string) []extract.Token {
	out := make([]extract.Token, 0, len(ids))
	for _, id := range ids {
		out = append(out, tok(id))
	}
	return out
}

func fakePolicy() (Policy, *fakeSource, *fakeSource, *fakeSource) {
	v := &fakeSource{name: "variables"}
	s := &fakeSource{name: "styles"}
	c := &fakeSource{name: "scan"}
	return Policy{Variables: v, Styles: s, Scan: c}, v, s, c
}

func TestExplicitVariablesOnly(t *testing.T) {
	p, v, s, c := fakePolicy()
	v.tokens = tokensOf("a")

	tokens, diags, err := p.Resolve(context.Background(), SourceRequest{FileID: "f"}, SourceVariables, false)

	require.NoError(t, err)
	assert.Equal(t, tokensOf("a"), tokens)
	assert.Empty(t, diags)
	assert.Equal(t, 1, v.calls)
	assert.Equal(t, 0, s.calls)
	assert.Equal(t, 0, c.calls)
}

func TestExplicitVariablesFailureIsError(t *testing.T) {
	p, v, _, _ := fakePolicy()
	v.err = errors.New("upstream 403")

	_, _, err := p.Resolve(context.Background(), SourceRequest{FileID: "f"}, SourceVariables, false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "upstream 403")
}

func TestExplicitStylesOnly(t *testing.T) {
	p, v, s, c := fakePolicy()
	s.tokens = tokensOf("b")

	tokens, diags, err := p.Resolve(context.Background(), SourceRequest{FileID: "f"}, SourceStyles, false)

	require.NoError(t, err)
	assert.Equal(t, tokensOf("b"), tokens)
	assert.Empty(t, diags)
	assert.Equal(t, 0, v.calls)
	assert.Equal(t, 1, s.calls)
	assert.Equal(t, 0, c.calls)
}

func TestExplicitScanOnly(t *testing.T) {
	p, v, s, c := fakePolicy()
	c.tokens = tokensOf("c")

	tokens, _, err := p.Resolve(context.Background(), SourceRequest{FileID: "f"}, SourceScan, false)

	require.NoError(t, err)
	assert.Equal(t, tokensOf("c"), tokens)
	assert.Equal(t, 0, v.calls)
	assert.Equal(t, 0, s.calls)
	assert.Equal(t, 1, c.calls)
}

func TestAutoVariablesNonEmpty(t *testing.T) {
	p, v, s, c := fakePolicy()
	v.tokens = tokensOf("a")

	tokens, diags, err := p.Resolve(context.Background(), SourceRequest{FileID: "f"}, SourceAuto, true)

	require.NoError(t, err)
	assert.Equal(t, tokensOf("a"), tokens)
	assert.Empty(t, diags)
	assert.Equal(t, 1, v.calls)
	assert.Equal(t, 0, s.calls)
	assert.Equal(t, 0, c.calls)
}

func TestAutoVariablesEmptyStylesNonEmpty(t *testing.T) {
	p, v, s, c := fakePolicy()
	v.tokens = nil
	s.tokens = tokensOf("b")

	tokens, diags, err := p.Resolve(context.Background(), SourceRequest{FileID: "f"}, SourceAuto, true)

	require.NoError(t, err)
	assert.Equal(t, tokensOf("b"), tokens)
	assert.Empty(t, diags, "empty variables is not a diagnostic")
	assert.Equal(t, 1, v.calls)
	assert.Equal(t, 1, s.calls)
	assert.Equal(t, 0, c.calls)
}

func TestAutoVariablesFailureStylesNonEmptyRecordsDiagnostic(t *testing.T) {
	p, v, s, c := fakePolicy()
	v.err = errors.New("variables down")
	s.tokens = tokensOf("b")

	tokens, diags, err := p.Resolve(context.Background(), SourceRequest{FileID: "f"}, SourceAuto, true)

	require.NoError(t, err)
	assert.Equal(t, tokensOf("b"), tokens)
	require.Len(t, diags, 1)
	assert.Equal(t, "warning", diags[0].Severity)
	assert.Contains(t, diags[0].Message, "variables source failed")
	assert.Equal(t, 1, v.calls)
	assert.Equal(t, 1, s.calls)
	assert.Equal(t, 0, c.calls)
}

func TestAutoBothEmptyScanEnabled(t *testing.T) {
	p, v, s, c := fakePolicy()
	v.tokens = nil
	s.tokens = nil
	c.tokens = tokensOf("c")

	tokens, diags, err := p.Resolve(context.Background(), SourceRequest{FileID: "f"}, SourceAuto, true)

	require.NoError(t, err)
	assert.Equal(t, tokensOf("c"), tokens)
	require.Len(t, diags, 1)
	assert.Equal(t, "note", diags[0].Severity)
	assert.Contains(t, diags[0].Message, "scanning document nodes")
	// call order: variables -> styles -> scan
	assert.Equal(t, 1, v.calls)
	assert.Equal(t, 1, s.calls)
	assert.Equal(t, 1, c.calls)
}

func TestAutoBothEmptyScanDisabled(t *testing.T) {
	p, v, s, c := fakePolicy()
	v.tokens = nil
	s.tokens = nil

	tokens, diags, err := p.Resolve(context.Background(), SourceRequest{FileID: "f"}, SourceAuto, false)

	require.NoError(t, err)
	assert.Nil(t, tokens)
	require.Len(t, diags, 1)
	assert.Equal(t, "note", diags[0].Severity)
	assert.Contains(t, diags[0].Message, "scan disabled")
	assert.Equal(t, 1, v.calls)
	assert.Equal(t, 1, s.calls)
	assert.Equal(t, 0, c.calls)
}

func TestAutoBothEmptyScanFailureIsError(t *testing.T) {
	p, v, s, c := fakePolicy()
	v.tokens = nil
	s.tokens = nil
	c.err = errors.New("scan failed")

	_, diags, err := p.Resolve(context.Background(), SourceRequest{FileID: "f"}, SourceAuto, true)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "scan failed")
	// The scan-fallback note is emitted before the failing scan.
	require.Len(t, diags, 1)
	assert.Equal(t, "note", diags[0].Severity)
}

func TestPolicyContextPropagates(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "marker")

	p, v, s, c := fakePolicy()
	v.tokens = nil
	s.tokens = nil
	c.tokens = tokensOf("c")

	_, _, err := p.Resolve(ctx, SourceRequest{FileID: "f"}, SourceAuto, true)
	require.NoError(t, err)
	assert.Equal(t, "marker", v.gotCtx.Value(ctxKey{}))
	assert.Equal(t, "marker", s.gotCtx.Value(ctxKey{}))
	assert.Equal(t, "marker", c.gotCtx.Value(ctxKey{}))
}

// --- Service orchestration tests ---

type fakeFormatter struct {
	gotTokens []extract.Token
	gotFormat string
	gotPrefix string
	out       string
	err       error
}

func (f *fakeFormatter) Format(tokens []extract.Token, format, prefix string) (string, error) {
	f.gotTokens = tokens
	f.gotFormat = format
	f.gotPrefix = prefix
	if f.err != nil {
		return "", f.err
	}
	return f.out, nil
}

type fakeSink struct {
	gotPath string
	gotData []byte
	err     error
}

func (f *fakeSink) Write(_ context.Context, path string, data []byte) error {
	f.gotPath = path
	f.gotData = data
	return f.err
}

func TestServiceFormatsAndPersists(t *testing.T) {
	p, v, _, _ := fakePolicy()
	v.tokens = tokensOf("a")
	fm := &fakeFormatter{out: ":root{}"}
	sink := &fakeSink{}
	svc := New(p, fm, sink)

	result, err := svc.Run(context.Background(), Request{
		FileID: "f", Source: SourceVariables, Format: "css", Prefix: "fig-", OutputPath: "tokens.css",
	})

	require.NoError(t, err)
	assert.Equal(t, ":root{}", result.Formatted)
	assert.Equal(t, "tokens.css", result.OutputPath)
	assert.Equal(t, len(":root{}"), result.Bytes)
	assert.Equal(t, "fig-", fm.gotPrefix)
	assert.Equal(t, "tokens.css", sink.gotPath)
	assert.Equal(t, []byte(":root{}"), sink.gotData)
}

func TestServiceFormatterFailureStopsBeforeSink(t *testing.T) {
	p, v, _, _ := fakePolicy()
	v.tokens = tokensOf("a")
	fm := &fakeFormatter{err: errors.New("bad format")}
	sink := &fakeSink{}
	svc := New(p, fm, sink)

	_, err := svc.Run(context.Background(), Request{FileID: "f", Source: SourceVariables, Format: "bad", OutputPath: "x.css"})

	require.Error(t, err)
	assert.Nil(t, sink.gotData, "sink must not run when formatting fails")
}

func TestServiceSinkFailureStops(t *testing.T) {
	p, v, _, _ := fakePolicy()
	v.tokens = tokensOf("a")
	fm := &fakeFormatter{out: ":root{}"}
	sink := &fakeSink{err: errors.New("disk full")}
	svc := New(p, fm, sink)

	_, err := svc.Run(context.Background(), Request{FileID: "f", Source: SourceVariables, Format: "css", OutputPath: "x.css"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "disk full")
}

func TestServiceStdoutReturnsFormattedOnly(t *testing.T) {
	p, v, _, _ := fakePolicy()
	v.tokens = tokensOf("a")
	fm := &fakeFormatter{out: ":root{}"}
	svc := New(p, fm, &fakeSink{})

	result, err := svc.Run(context.Background(), Request{FileID: "f", Source: SourceVariables, Format: "css"})

	require.NoError(t, err)
	assert.Equal(t, ":root{}", result.Formatted)
	assert.Equal(t, "", result.OutputPath)
	assert.Equal(t, 0, result.Bytes)
}
