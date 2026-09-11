package cli

import (
	"bytes"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/env"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadClient_FromEnv(t *testing.T) {
	t.Setenv("FIGMA_ACCESS_TOKEN", "secret-token")

	client, err := LoadClient()

	require.NoError(t, err)
	assert.Equal(t, "secret-token", client.Token)
}

func TestNewPrinterUsesCommandOutput(t *testing.T) {
	var output bytes.Buffer
	cmd := &cobra.Command{}
	cmd.SetOut(&output)

	require.NoError(t, NewPrinter(cmd).Text("result", "captured"))

	assert.Equal(t, "captured", output.String())
}

func TestLoadClient_MissingToken(t *testing.T) {
	t.Setenv("FIGMA_ACCESS_TOKEN", "")

	_, err := LoadClient()

	require.Error(t, err)
	var tokenErr *env.ErrTokenNotSet
	assert.ErrorAs(t, err, &tokenErr, "underlying ErrTokenNotSet retained as cause")
}
