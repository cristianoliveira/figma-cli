package figma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeNodeID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "api style", input: "20089:685897", expected: "20089:685897"},
		{name: "url style", input: "20089-685897", expected: "20089:685897"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, NormalizeNodeID(tt.input))
		})
	}
}

func TestResolveNodeIDsPrefersExplicitID(t *testing.T) {
	input := &FileInput{FileID: "file123", NodeIDs: []string{"1:2"}}

	got := ResolveNodeIDs(input, "3:4")

	assert.Equal(t, []string{"3:4"}, got)
}

func TestResolveNodeIDsFallsBackToURLNodeIDs(t *testing.T) {
	input := &FileInput{FileID: "file123", NodeIDs: []string{"1:2"}}

	got := ResolveNodeIDs(input, "")

	assert.Equal(t, []string{"1:2"}, got)
}

func TestResolveSingleNodeIDUsesURLOrExplicitScope(t *testing.T) {
	input := &FileInput{FileID: "file123", NodeIDs: []string{"1:2"}}

	fromURL, err := ResolveSingleNodeID(input, "", "inspect")
	require.NoError(t, err)
	assert.Equal(t, "1:2", fromURL)

	fromFlag, err := ResolveSingleNodeID(input, "3-4", "inspect")
	require.NoError(t, err)
	assert.Equal(t, "3:4", fromFlag)
}

func TestResolveSingleNodeIDRejectsMissingAndMultipleScope(t *testing.T) {
	_, err := ResolveSingleNodeID(&FileInput{}, "", "inspect")
	assert.EqualError(t, err, "inspect requires a Figma URL with node-id or --id")

	_, err = ResolveSingleNodeID(&FileInput{NodeIDs: []string{"1:2", "3:4"}}, "", "inspect")
	assert.EqualError(t, err, "inspect requires exactly one node ID")
}

func TestResolveRequiredNodeIDsPreservesMultipleScope(t *testing.T) {
	got, err := ResolveRequiredNodeIDs(&FileInput{NodeIDs: []string{"1:2", "3:4"}}, "", "assets")

	require.NoError(t, err)
	assert.Equal(t, []string{"1:2", "3:4"}, got)
}

func TestResolveNodeIDsAllowsCommaSeparatedExplicitIDs(t *testing.T) {
	input := &FileInput{FileID: "file123"}

	got := ResolveNodeIDs(input, "3-4,5:6")

	assert.Equal(t, []string{"3:4", "5:6"}, got)
}
