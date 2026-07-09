package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
)

func TestBuildNodesURLWithNodeIDs(t *testing.T) {
	got, err := figma.BuildNodesURL("file123", []string{"1:2", "3:4"})
	if err != nil {
		t.Fatalf("BuildNodesURL() error = %v", err)
	}

	expected := "https://api.figma.com/v1/files/file123/nodes?ids=1%3A2%2C3%3A4"
	if got != expected {
		t.Fatalf("BuildNodesURL() = %v, expected %v", got, expected)
	}
}

func TestBuildFileURLWithoutNodeIDs(t *testing.T) {
	got, err := figma.BuildFileURL("file123", nil, "", "")
	if err != nil {
		t.Fatalf("BuildFileURL() error = %v", err)
	}

	expected := "https://api.figma.com/v1/files/file123"
	if got != expected {
		t.Fatalf("BuildFileURL() = %v, expected %v", got, expected)
	}
}
