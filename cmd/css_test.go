package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCSSCommandNilWriterReturnsConfigError guards against a nil
// ArtifactWriter panic when --output is set without a wired writer.
func TestCSSCommandNilWriterReturnsConfigError(t *testing.T) {
	client := fixtureClient(t, `{"nodes":{"1:1":{"document":{"id":"1:1","name":"Frame","type":"FRAME","layoutMode":"HORIZONTAL"}}}}`)

	deps := DepsForLoadClient(func() (*figma.Client, error) { return client, nil })
	deps.ArtifactWriter = nil

	result := executeCommand(newCSSCommand(deps), "abc", "--id", "1:1", "--output", "out.css")

	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "artifact writer is not configured")
}
