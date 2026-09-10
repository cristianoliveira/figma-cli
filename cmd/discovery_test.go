package cmd

import (
	"testing"
	"time"

	"github.com/cristianoliveira/figma-cli/internal/figma"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveTeamInput(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		env     string
		want    string
		wantErr bool
	}{
		{name: "explicit ID", args: []string{"123"}, env: "456", want: "123"},
		{name: "environment default", env: "456", want: "456"},
		{name: "missing both", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveTeamInput(tt.args, func(string) string { return tt.env })
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorContains(t, err, "FIGMA_TEAM_ID")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewProjectsOutput(t *testing.T) {
	response := figma.TeamProjects{Name: "Wire", Projects: []figma.TeamProject{{ID: "123", Name: "Design System"}}}

	got := newProjectsOutput(response)

	assert.Equal(t, projectsOutput{Team: "Wire", Total: 1, Projects: []projectOutput{{ID: "123", Name: "Design System"}}}, got)
}

func TestNewFilesOutput(t *testing.T) {
	modified, err := time.Parse(time.RFC3339, "2026-07-07T12:00:00Z")
	require.NoError(t, err)

	thumbnail := "https://example.com/thumb.png"
	response := figma.ProjectFiles{
		Name: "Design System",
		Files: []figma.ProjectFile{
			{Key: "abc", Name: "Primitives", LastModified: modified, ThumbnailURL: &thumbnail},
		},
	}

	got := newFilesOutput(response)

	require.Len(t, got.Files, 1)
	assert.Equal(t, "Design System", got.Project)
	assert.Equal(t, 1, got.Total)
	assert.Equal(t, "abc", got.Files[0].Key)
	assert.Equal(t, "2026-07-07T12:00:00Z", got.Files[0].LastModified)
	assert.Equal(t, "https://example.com/thumb.png", *got.Files[0].ThumbnailURL)
}

func TestNewDiscoveryOutputsUseEmptyArrays(t *testing.T) {
	assert.Empty(t, newProjectsOutput(figma.TeamProjects{}).Projects)
	assert.NotNil(t, newProjectsOutput(figma.TeamProjects{}).Projects)
	assert.Empty(t, newFilesOutput(figma.ProjectFiles{}).Files)
	assert.NotNil(t, newFilesOutput(figma.ProjectFiles{}).Files)
}
