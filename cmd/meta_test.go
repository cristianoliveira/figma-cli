package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetaCommandUsesInjectedClient(t *testing.T) {
	client := fixtureClient(t, `{"name":"Checkout","document":{"id":"0:0","name":"Document","type":"DOCUMENT"}}`)

	command := newMetaCommand(func() (*figma.Client, error) { return client, nil })

	assert.Equal(t, "meta", command.Name())
	result := executeCommand(command, "abc")

	require.NoError(t, result.Err)
	assert.JSONEq(t, `{"name":"Checkout","document":{"id":"0:0","name":"Document","type":"DOCUMENT"}}`, result.Stdout)
}

func TestMetaCommandRejectsInvalidInputBeforeLoadingClient(t *testing.T) {
	loaded := false
	result := executeCommand(newMetaCommand(func() (*figma.Client, error) {
		loaded = true
		return nil, nil
	}), "https://www.figma.com/community/x")

	require.Error(t, result.Err)
	assert.False(t, loaded)
}
