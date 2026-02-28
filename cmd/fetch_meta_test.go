package cmd

import (
	"testing"
)

func TestExtractFileID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "plain file ID",
			input:    "grnVU2vAihHXwYgHryu2xE",
			expected: "grnVU2vAihHXwYgHryu2xE",
			wantErr:  false,
		},
		{
			name:     "Figma design URL with query",
			input:    "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Drive--Cells-?node-id=339-27545&p=f&m=dev",
			expected: "grnVU2vAihHXwYgHryu2xE",
			wantErr:  false,
		},
		{
			name:     "Figma file URL",
			input:    "https://www.figma.com/file/abc123/My-Design",
			expected: "abc123",
			wantErr:  false,
		},
		{
			name:     "Figma URL without design or file",
			input:    "https://www.figma.com/community/abc",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "URL with http prefix but invalid format",
			input:    "http:///example",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "URL missing file ID after design",
			input:    "https://www.figma.com/design/",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "HTTP URL",
			input:    "http://figma.com/design/abc123/Title",
			expected: "abc123",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractFileID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractFileID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.expected {
				t.Errorf("extractFileID() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

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
			got, err := parseInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}
			if got.fileID != tt.expectedID {
				t.Errorf("parseInput() fileID = %v, expected %v", got.fileID, tt.expectedID)
			}
			if len(got.nodeIDs) != len(tt.expectedIDs) {
				t.Errorf("parseInput() nodeIDs length = %v, expected %v", len(got.nodeIDs), len(tt.expectedIDs))
				return
			}
			for i := range got.nodeIDs {
				if got.nodeIDs[i] != tt.expectedIDs[i] {
					t.Errorf("parseInput() nodeIDs[%d] = %v, expected %v", i, got.nodeIDs[i], tt.expectedIDs[i])
				}
			}
		})
	}
}
