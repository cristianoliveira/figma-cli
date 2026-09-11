package cli

import (
	"bytes"
	"errors"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/operr"
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

func TestErrorContractClassifiesNeutralCategories(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		category  string
		message   string
		recovery  string
		forbidden string
	}{
		{
			name:      "authentication",
			err:       operr.New(operr.CategoryAuthentication, "Figma authentication is not configured.", "Set FIGMA_ACCESS_TOKEN and retry.", errors.New("token=secret")),
			category:  "authentication",
			message:   "Figma authentication is not configured.",
			recovery:  "Set FIGMA_ACCESS_TOKEN and retry.",
			forbidden: "token=secret",
		},
		{
			name:      "authorization",
			err:       operr.New(operr.CategoryAuthorization, "Figma rejected authentication or access.", "Check FIGMA_ACCESS_TOKEN and file permissions, then retry.", errors.New("secret-provider-output")),
			category:  "authorization",
			message:   "Figma rejected authentication or access.",
			recovery:  "Check FIGMA_ACCESS_TOKEN and file permissions, then retry.",
			forbidden: "secret-provider-output",
		},
		{
			name:      "rate limit",
			err:       operr.New(operr.CategoryRateLimit, "Figma rate limit reached.", "Wait before retrying the Figma request.", errors.New("raw upstream")),
			category:  "rate_limit",
			message:   "Figma rate limit reached.",
			recovery:  "Wait before retrying the Figma request.",
			forbidden: "raw upstream",
		},
		{
			name:      "dependency unavailable",
			err:       operr.New(operr.CategoryDependencyUnavailable, "Figma request failed.", "Check Figma availability and retry.", errors.New("timeout 5xx")),
			category:  "dependency_unavailable",
			message:   "Figma request failed.",
			recovery:  "Check Figma availability and retry.",
			forbidden: "timeout 5xx",
		},
		{
			name:      "artifact access",
			err:       operr.New(operr.CategoryArtifactAccess, "Could not access a required file.", "Check the file path and permissions, then retry.", errors.New("/Users/private/design.png")),
			category:  "artifact_access",
			message:   "Could not access a required file.",
			recovery:  "Check the file path and permissions, then retry.",
			forbidden: "/Users/private",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contract, render := ErrorContract(test.err)
			require.True(t, render)
			assert.Equal(t, test.category, contract.Error.Category)
			assert.Equal(t, 1, contract.Error.ExitCode)
			assert.Equal(t, test.message, contract.Error.Message)
			assert.Equal(t, test.recovery, contract.Error.Recovery)
			assert.NotContains(t, contract.Error.Message+contract.Error.Recovery, test.forbidden)
		})
	}
}

func TestErrorContractUnknownFailureFallsBackToOperational(t *testing.T) {
	contract, render := ErrorContract(errors.New("provider stack trace with token=secret"))
	require.True(t, render)
	assert.Equal(t, "operational", contract.Error.Category)
	assert.Equal(t, 1, contract.Error.ExitCode)
	assert.Equal(t, "Command could not complete.", contract.Error.Message)
	assert.Equal(t, "Check inputs and dependencies, then retry.", contract.Error.Recovery)
	assert.NotContains(t, contract.Error.Message+contract.Error.Recovery, "token=secret")
}

func TestErrorContractInvalidInputIsUsage(t *testing.T) {
	err := operr.New(operr.CategoryInvalidInput, "malformed request.", "Check the input and retry.", errors.New("raw input"))
	contract, render := ErrorContract(err)
	require.True(t, render)
	assert.Equal(t, "usage", contract.Error.Category)
	assert.Equal(t, 2, contract.Error.ExitCode)
	assert.Equal(t, "malformed request.", contract.Error.Message)
}

func TestErrorContractPreservesSilentAndResultBearingErrors(t *testing.T) {
	for _, err := range []error{&ExitCodeError{Code: 1}, NewResultError(errors.New("gate failed"))} {
		_, render := ErrorContract(err)
		assert.False(t, render)
	}
}

func TestResultErrorRendersSafeMessageWithoutEnvelope(t *testing.T) {
	// A result-bearing error already emitted a result; RenderError prints
	// a safe message to stderr but must not emit a structured envelope or
	// the raw error text.
	command := &cobra.Command{Use: "tool"}
	var stderr bytes.Buffer
	command.SetErr(&stderr)

	require.NoError(t, RenderError(command, NewResultError(errors.New("gate failed"))))
	assert.Equal(t, "Command could not complete.\n", stderr.String())
	assert.NotContains(t, stderr.String(), "gate failed")
}

func TestResultErrorRedactsSensitiveCauses(t *testing.T) {
	command := &cobra.Command{Use: "tool"}
	var stderr bytes.Buffer
	command.SetErr(&stderr)

	err := NewResultError(errors.New("response-body token=SECRET123 Authorization: Bearer SECRET123 /Users/private/secret.png"))
	require.NoError(t, RenderError(command, err))

	for _, secret := range []string{"SECRET123", "Bearer", "/Users/private"} {
		assert.NotContains(t, stderr.String(), secret)
	}
	assert.Contains(t, stderr.String(), "Command could not complete.")
}

func TestResultErrorUsesClassifiedMessage(t *testing.T) {
	command := &cobra.Command{Use: "tool"}
	var stderr bytes.Buffer
	command.SetErr(&stderr)

	err := NewResultError(operr.New(operr.CategoryDependencyUnavailable,
		"Figma request failed.",
		"Check Figma availability and retry.",
		errors.New("token=SECRET123")))
	require.NoError(t, RenderError(command, err))

	assert.Equal(t, "Figma request failed.\n", stderr.String())
	assert.NotContains(t, stderr.String(), "SECRET123")
}
