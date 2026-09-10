package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFramesCommandContract(t *testing.T) {
	command := newFramesCommand(DepsForLoadClient(func() (*figma.Client, error) { return nil, nil }))

	assert.Equal(t, "frames [figma-url-or-file-id]", command.Use)
	require.NotNil(t, command.Flags().Lookup("id"))
	require.NotNil(t, command.Flags().Lookup("node"))
}
