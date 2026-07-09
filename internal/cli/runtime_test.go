package cli

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadClient_FromEnv(t *testing.T) {
	t.Setenv("FIGMA_ACCESS_TOKEN", "secret-token")

	client, err := LoadClient()

	require.NoError(t, err)
	assert.Equal(t, "secret-token", client.Token)
}

func TestLoadClient_MissingToken(t *testing.T) {
	t.Setenv("FIGMA_ACCESS_TOKEN", "")

	_, err := LoadClient()

	require.Error(t, err)
	assert.IsType(t, &env.ErrTokenNotSet{}, err)
}
