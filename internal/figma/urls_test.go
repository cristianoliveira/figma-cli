package figma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildNodesURLWithNodeIDs(t *testing.T) {
	got, err := BuildNodesURL("file123", []string{"1:2", "3:4"})

	require.NoError(t, err)
	assert.Equal(t, "https://api.figma.com/v1/files/file123/nodes?ids=1%3A2%2C3%3A4", got)
}

func TestBuildFileURLWithoutNodeIDs(t *testing.T) {
	got, err := BuildFileURL("file123", nil, "", "")

	require.NoError(t, err)
	assert.Equal(t, "https://api.figma.com/v1/files/file123", got)
}

func TestBuildMeURL(t *testing.T) {
	got := BuildMeURL()

	assert.Equal(t, "https://api.figma.com/v1/me", got)
}

func TestBuildVersionsURL(t *testing.T) {
	got := BuildVersionsURL("grnVU2vAihHXwYgHryu2xE")

	assert.Equal(t, "https://api.figma.com/v1/files/grnVU2vAihHXwYgHryu2xE/versions", got)
}

func TestBuildTeamProjectsURL(t *testing.T) {
	assert.Equal(t, "https://api.figma.com/v1/teams/123/projects", BuildTeamProjectsURL("123"))
}

func TestBuildProjectFilesURL(t *testing.T) {
	assert.Equal(t, "https://api.figma.com/v1/projects/987/files", BuildProjectFilesURL("987", false))
	assert.Equal(t, "https://api.figma.com/v1/projects/987/files?branch_data=true", BuildProjectFilesURL("987", true))
}

func TestBuildExportURL(t *testing.T) {
	got, err := BuildExportURL("file123", []string{"1:2"}, "png")

	require.NoError(t, err)
	assert.Equal(t, "https://api.figma.com/v1/images/file123?format=png&ids=1%3A2", got)
}

func TestBuildExportURLRequiresNodeID(t *testing.T) {
	_, err := BuildExportURL("file123", nil, "png")

	assert.Error(t, err)
}

func TestBuildExportURLRequiresFormat(t *testing.T) {
	_, err := BuildExportURL("file123", []string{"1:2"}, "")

	assert.Error(t, err)
}

func TestBuildStylesURL(t *testing.T) {
	got := BuildStylesURL("file123")

	assert.Equal(t, "https://api.figma.com/v1/files/file123/styles", got)
}

func TestBuildVariablesURL(t *testing.T) {
	got := BuildVariablesURL("file123")

	assert.Equal(t, "https://api.figma.com/v1/files/file123/variables/local", got)
}
