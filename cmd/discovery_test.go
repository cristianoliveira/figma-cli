package cmd

import (
	"encoding/json"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma/api"
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
	response := api.GetTeamProjectsResponse{Name: "Wire", Projects: []api.Project{{Id: "123", Name: "Design System"}}}

	got := newProjectsOutput(response)

	assert.Equal(t, projectsOutput{Team: "Wire", Projects: []projectOutput{{ID: "123", Name: "Design System"}}}, got)
}

func TestNewFilesOutput(t *testing.T) {
	var response api.GetProjectFilesResponse
	require.NoError(t, json.Unmarshal([]byte(`{"name":"Design System","files":[{"key":"abc","name":"Primitives","last_modified":"2026-07-07T12:00:00Z","thumbnail_url":"https://example.com/thumb.png"}]}`), &response))

	got := newFilesOutput(response)

	require.Len(t, got.Files, 1)
	assert.Equal(t, "Design System", got.Project)
	assert.Equal(t, "abc", got.Files[0].Key)
	assert.Equal(t, "2026-07-07T12:00:00Z", got.Files[0].LastModified)
	assert.Equal(t, "https://example.com/thumb.png", *got.Files[0].ThumbnailURL)
}

func TestNewDiscoveryOutputsUseEmptyArrays(t *testing.T) {
	assert.Empty(t, newProjectsOutput(api.GetTeamProjectsResponse{}).Projects)
	assert.NotNil(t, newProjectsOutput(api.GetTeamProjectsResponse{}).Projects)
	assert.Empty(t, newFilesOutput(api.GetProjectFilesResponse{}).Files)
	assert.NotNil(t, newFilesOutput(api.GetProjectFilesResponse{}).Files)
}
