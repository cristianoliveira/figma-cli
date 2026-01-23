package figma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *ParsedURL
		expectError bool
	}{
		{
			name:  "design URL with node-id",
			input: "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/-Cells--Drive?node-id=2270-190221",
			expected: &ParsedURL{
				FileKey:  "grnVU2vAihHXwYgHryu2xE",
				NodeID:   "2270-190221",
				FileName: "-Cells--Drive",
			},
		},
		{
			name:  "file URL with node-id",
			input: "https://www.figma.com/file/abc123/My-Design?node-id=123-456",
			expected: &ParsedURL{
				FileKey:  "abc123",
				NodeID:   "123-456",
				FileName: "My-Design",
			},
		},
		{
			name:  "URL with version and timestamp",
			input: "https://www.figma.com/design/xyz456/Another-Design?node-id=789-012&version-id=123&t=abc",
			expected: &ParsedURL{
				FileKey:   "xyz456",
				NodeID:    "789-012",
				FileName:  "Another-Design",
				Version:   "123",
				Timestamp: "abc",
			},
		},
		{
			name:        "invalid URL",
			input:       "not-a-url",
			expectError: true,
		},
		{
			name:        "unsupported path",
			input:       "https://www.figma.com/unknown/abc/name",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseURL(tt.input)
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
