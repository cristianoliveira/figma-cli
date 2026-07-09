package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
)

func TestParseInput(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedID  string
		expectedIDs []string
		wantErr     bool
	}{
		{
			name:        "plain file ID",
			input:       "grnVU2vAihHXwYgHryu2xE",
			expectedID:  "grnVU2vAihHXwYgHryu2xE",
			expectedIDs: nil,
			wantErr:     false,
		},
		{
			name:        "Figma design URL with node-id",
			input:       "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=339-27545&p=f&m=dev",
			expectedID:  "grnVU2vAihHXwYgHryu2xE",
			expectedIDs: []string{"339:27545"},
			wantErr:     false,
		},
		{
			name:        "Figma design URL without node-id",
			input:       "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-",
			expectedID:  "grnVU2vAihHXwYgHryu2xE",
			expectedIDs: nil,
			wantErr:     false,
		},
		{
			name:        "Figma file URL",
			input:       "https://www.figma.com/file/abc123/My-Design",
			expectedID:  "abc123",
			expectedIDs: nil,
			wantErr:     false,
		},
		{
			name:        "Figma URL without design or file",
			input:       "https://www.figma.com/community/abc",
			expectedID:  "",
			expectedIDs: nil,
			wantErr:     true,
		},
		{
			name:        "URL with multiple node-id parameters",
			input:       "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=339-27545&node-id=440-12345",
			expectedID:  "grnVU2vAihHXwYgHryu2xE",
			expectedIDs: []string{"339:27545", "440:12345"},
			wantErr:     false,
		},
		{
			name:        "node-id with multiple hyphens",
			input:       "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=339-27545-99",
			expectedID:  "grnVU2vAihHXwYgHryu2xE",
			expectedIDs: []string{"339:27545:99"},
			wantErr:     false,
		},
		{
			name:        "node-id with comma-separated values",
			input:       "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=339-27545,440-12345",
			expectedID:  "grnVU2vAihHXwYgHryu2xE",
			expectedIDs: []string{"339:27545", "440:12345"},
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := figma.ParseInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.FileID != tt.expectedID {
				t.Errorf("ParseInput() FileID = %v, expected %v", got.FileID, tt.expectedID)
			}
			if len(got.NodeIDs) != len(tt.expectedIDs) {
				t.Errorf("ParseInput() NodeIDs length = %v, expected %v", len(got.NodeIDs), len(tt.expectedIDs))
				return
			}
			for i := range got.NodeIDs {
				if got.NodeIDs[i] != tt.expectedIDs[i] {
					t.Errorf("ParseInput() NodeIDs[%d] = %v, expected %v", i, got.NodeIDs[i], tt.expectedIDs[i])
				}
			}
		})
	}
}
