package cmd

import (
	"context"
	"errors"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/cristianoliveira/figma-cli/internal/tokens"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeTokenService records the request and returns a fixed result.
type fakeTokenService struct {
	gotRequest tokens.Request
	result     tokens.Result
	err        error
	calls      int
}

func (f *fakeTokenService) Run(_ context.Context, req tokens.Request) (tokens.Result, error) {
	f.calls++
	f.gotRequest = req
	if f.err != nil {
		return tokens.Result{}, f.err
	}
	return f.result, nil
}

func depsWithFakeTokens(fake *fakeTokenService) Deps {
	deps := Deps{
		LoadClient:   func() (*figma.Client, error) { return nil, errors.New("must not load client") },
		ResolveExec:  func() (string, error) { return "/tmp/figma", nil },
		GetEnv:       func(string) string { return "" },
		TokenService: func() (TokenService, error) { return fake, nil },
	}
	return deps
}

func TestTokensCommandMapsFlagsToRequest(t *testing.T) {
	fake := &fakeTokenService{result: tokens.Result{Formatted: ":root{}"}}
	result := executeDefaultCommand(newTokensCommand(depsWithFakeTokens(fake)),
		"abc", "--source", "variables", "--format", "css", "--prefix", "fig-", "--mode", "dark")

	require.NoError(t, result.Err)
	assert.Equal(t, "variables", fake.gotRequest.Source)
	assert.Equal(t, "css", fake.gotRequest.Format)
	assert.Equal(t, "fig-", fake.gotRequest.Prefix)
	assert.Equal(t, "dark", fake.gotRequest.Mode)
	assert.Equal(t, "abc", fake.gotRequest.FileID)
}

func TestTokensCommandRendersDiagnosticsToStderr(t *testing.T) {
	fake := &fakeTokenService{result: tokens.Result{
		Formatted: ":root{}",
		Diagnostics: []tokens.Diagnostic{
			{Severity: "note", Message: "no Styles/Variables found; scanning document nodes"},
		},
	}}
	result := executeDefaultCommand(newTokensCommand(depsWithFakeTokens(fake)), "abc", "--source", "auto")

	require.NoError(t, result.Err)
	assert.Contains(t, result.Stderr, "note: no Styles/Variables found; scanning document nodes")
	assert.Contains(t, result.Stdout, ":root{}")
}

func TestTokensCommandRendersFileMetadata(t *testing.T) {
	fake := &fakeTokenService{result: tokens.Result{
		Formatted:  ":root{}",
		OutputPath: "tokens.css",
		Bytes:      8,
	}}
	result := executeDefaultCommand(newTokensCommand(depsWithFakeTokens(fake)),
		"abc", "--source", "variables", "--format", "css", "--output", "tokens.css")

	require.NoError(t, result.Err)
	assert.Contains(t, result.Stdout, "tokens.css")
}

func TestTokensCommandForwardsServiceError(t *testing.T) {
	fake := &fakeTokenService{err: errors.New("variables source failed: forbidden")}
	result := executeDefaultCommand(newTokensCommand(depsWithFakeTokens(fake)), "abc", "--source", "variables")

	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "forbidden")
}
