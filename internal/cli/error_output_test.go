package cli

import (
	"bytes"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/env"
	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	toon "github.com/toon-format/toon-go"
)

func TestRenderErrorWritesStructuredUsageErrorInSelectedFormat(t *testing.T) {
	command := &cobra.Command{Use: "tool"}
	command.Flags().Bool("json", false, "")
	var stdout bytes.Buffer
	command.SetOut(&stdout)

	err := NewUsageErrorWithRecovery(errors.New("missing value"), "Run `tool --help` for valid usage.")
	require.NoError(t, RenderError(command, err))

	var decoded map[string]any
	require.NoError(t, toon.Unmarshal(stdout.Bytes(), &decoded))
	assert.Equal(t, map[string]any{"error": map[string]any{
		"category": "usage", "message": "missing value", "exitCode": float64(2),
		"recovery": "Run `tool --help` for valid usage.",
	}}, decoded)

	stdout.Reset()
	require.NoError(t, command.Flags().Set("json", "true"))
	require.NoError(t, RenderError(command, err))
	assert.JSONEq(t, `{"error":{"category":"usage","message":"missing value","exitCode":2,"recovery":"Run `+"`tool --help`"+` for valid usage."}}`, stdout.String())
}

func TestErrorContractRedactsOperationalDependencies(t *testing.T) {
	tests := []struct {
		name              string
		err               error
		message, recovery string
		forbidden         string
	}{
		{name: "token", err: &env.ErrTokenNotSet{}, message: "Figma authentication is not configured.", recovery: "Set FIGMA_ACCESS_TOKEN and retry.", forbidden: "environment variable not set"},
		{name: "file", err: &os.PathError{Op: "open", Path: "/Users/private/design.png", Err: os.ErrNotExist}, message: "Could not access a required file.", recovery: "Check the file path and permissions, then retry.", forbidden: "/Users/private"},
		{name: "authorization", err: &figma.ResponseError{StatusCode: http.StatusUnauthorized}, message: "Figma rejected authentication or access.", recovery: "Check FIGMA_ACCESS_TOKEN and file permissions, then retry.", forbidden: "secret-provider-output"},
		{name: "rate limit", err: &figma.ResponseError{StatusCode: http.StatusTooManyRequests}, message: "Figma rate limit reached.", recovery: "Wait before retrying the Figma request.", forbidden: "raw upstream"},
		{name: "generic", err: errors.New("provider stack trace with token=secret"), message: "Command could not complete.", recovery: "Check inputs and dependencies, then retry.", forbidden: "token=secret"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contract, render := ErrorContract(test.err)
			require.True(t, render)
			assert.Equal(t, "operational", contract.Error.Category)
			assert.Equal(t, 1, contract.Error.ExitCode)
			assert.Equal(t, test.message, contract.Error.Message)
			assert.Equal(t, test.recovery, contract.Error.Recovery)
			assert.NotContains(t, contract.Error.Message+contract.Error.Recovery, test.forbidden)
		})
	}
}

func TestErrorContractPreservesSilentAndResultBearingErrors(t *testing.T) {
	for _, err := range []error{&ExitCodeError{Code: 1}, NewResultError(errors.New("gate failed"))} {
		_, render := ErrorContract(err)
		assert.False(t, render)
	}
}
