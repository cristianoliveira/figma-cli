package figma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInput(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedID        string
		expectedIDs       []string
		expectedCommentID string
		wantErr           bool
	}{
		{name: "plain file ID", input: "exampleFileKey123", expectedID: "exampleFileKey123"},
		{name: "design URL with node-id", input: "https://www.figma.com/design/exampleFileKey123/Example-Design?node-id=339-27545&p=f&m=dev", expectedID: "exampleFileKey123", expectedIDs: []string{"339:27545"}},
		{name: "comment URL", input: "https://www.figma.com/design/QAhpkgySSOJ6gwJUTB0glb?node-id=4707-15501&m=dev#1838610593", expectedID: "QAhpkgySSOJ6gwJUTB0glb", expectedIDs: []string{"4707:15501"}, expectedCommentID: "1838610593"},
		{name: "design URL without node-id", input: "https://www.figma.com/design/exampleFileKey123/Example-Design", expectedID: "exampleFileKey123"},
		{name: "file URL", input: "https://www.figma.com/file/abc123/My-Design", expectedID: "abc123"},
		{name: "URL without design or file", input: "https://www.figma.com/community/abc", wantErr: true},
		{name: "multiple node-id params", input: "https://www.figma.com/design/exampleFileKey123/Example-Design?node-id=339-27545&node-id=440-12345", expectedID: "exampleFileKey123", expectedIDs: []string{"339:27545", "440:12345"}},
		{name: "node-id with multiple hyphens", input: "https://www.figma.com/design/exampleFileKey123/Example-Design?node-id=339-27545-99", expectedID: "exampleFileKey123", expectedIDs: []string{"339:27545:99"}},
		{name: "comma-separated node-ids", input: "https://www.figma.com/design/exampleFileKey123/Example-Design?node-id=339-27545,440-12345", expectedID: "exampleFileKey123", expectedIDs: []string{"339:27545", "440:12345"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInput(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedID, got.FileID)
			assert.Equal(t, tt.expectedIDs, got.NodeIDs)
			assert.Equal(t, tt.expectedCommentID, got.CommentID)
		})
	}
}

func TestParseDiscoveryInput(t *testing.T) {
	tests := []struct {
		name    string
		parse   func(string) (string, error)
		input   string
		want    string
		wantErr bool
	}{
		{name: "bare team ID", parse: ParseTeamInput, input: "123456789", want: "123456789"},
		{name: "files team URL", parse: ParseTeamInput, input: "https://www.figma.com/files/team/123456789/Wire", want: "123456789"},
		{name: "short team URL", parse: ParseTeamInput, input: "https://figma.com/team/123456789/Wire", want: "123456789"},
		{name: "bare project ID", parse: ParseProjectInput, input: "987654321", want: "987654321"},
		{name: "files project URL", parse: ParseProjectInput, input: "https://www.figma.com/files/project/987654321/Design-System", want: "987654321"},
		{name: "short project URL", parse: ParseProjectInput, input: "https://figma.com/project/987654321/Design-System", want: "987654321"},
		{name: "non-numeric bare ID", parse: ParseTeamInput, input: "team-123", wantErr: true},
		{name: "wrong URL kind", parse: ParseTeamInput, input: "https://figma.com/files/project/987/Design", wantErr: true},
		{name: "foreign host", parse: ParseProjectInput, input: "https://example.com/files/project/987/Design", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.parse(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseInputNormalizesNodeIDs(t *testing.T) {
	// node-id uses URL hyphens; the API expects colons. Covered by the table above,
	// this pins the behaviour explicitly.
	got, err := ParseInput("https://www.figma.com/design/FILE/T?node-id=10-20")
	require.NoError(t, err)
	assert.Equal(t, []string{"10:20"}, got.NodeIDs)
}
