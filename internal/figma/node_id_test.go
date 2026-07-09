package figma

import "testing"

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
			got := NormalizeNodeID(tt.input)
			if got != tt.expected {
				t.Fatalf("NormalizeNodeID() = %v, expected %v", got, tt.expected)
			}
		})
	}
}

func TestResolveNodeIDsPrefersExplicitID(t *testing.T) {
	input := &FileInput{FileID: "file123", NodeIDs: []string{"1:2"}}
	got := ResolveNodeIDs(input, "3:4")

	if len(got) != 1 || got[0] != "3:4" {
		t.Fatalf("ResolveNodeIDs() = %#v", got)
	}
}

func TestResolveNodeIDsFallsBackToURLNodeIDs(t *testing.T) {
	input := &FileInput{FileID: "file123", NodeIDs: []string{"1:2"}}
	got := ResolveNodeIDs(input, "")

	if len(got) != 1 || got[0] != "1:2" {
		t.Fatalf("ResolveNodeIDs() = %#v", got)
	}
}

func TestResolveNodeIDsAllowsCommaSeparatedExplicitIDs(t *testing.T) {
	input := &FileInput{FileID: "file123"}
	got := ResolveNodeIDs(input, "3-4,5:6")

	if len(got) != 2 || got[0] != "3:4" || got[1] != "5:6" {
		t.Fatalf("ResolveNodeIDs() = %#v", got)
	}
}
