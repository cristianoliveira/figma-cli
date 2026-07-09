package figma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseInput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedID  string
		expectedIDs []string
		wantErr     bool
	}{
		{name: "plain file ID", input: "grnVU2vAihHXwYgHryu2xE", expectedID: "grnVU2vAihHXwYgHryu2xE"},
		{name: "design URL with node-id", input: "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=339-27545&p=f&m=dev", expectedID: "grnVU2vAihHXwYgHryu2xE", expectedIDs: []string{"339:27545"}},
		{name: "design URL without node-id", input: "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-", expectedID: "grnVU2vAihHXwYgHryu2xE"},
		{name: "file URL", input: "https://www.figma.com/file/abc123/My-Design", expectedID: "abc123"},
		{name: "URL without design or file", input: "https://www.figma.com/community/abc", wantErr: true},
		{name: "multiple node-id params", input: "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=339-27545&node-id=440-12345", expectedID: "grnVU2vAihHXwYgHryu2xE", expectedIDs: []string{"339:27545", "440:12345"}},
		{name: "node-id with multiple hyphens", input: "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=339-27545-99", expectedID: "grnVU2vAihHXwYgHryu2xE", expectedIDs: []string{"339:27545:99"}},
		{name: "comma-separated node-ids", input: "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=339-27545,440-12345", expectedID: "grnVU2vAihHXwYgHryu2xE", expectedIDs: []string{"339:27545", "440:12345"}},
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
