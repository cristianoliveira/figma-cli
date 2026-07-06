package cmd

import "testing"

func TestBuildTextsAPIURLWithNodeIDs(t *testing.T) {
	got, err := buildTextsAPIURL("file123", []string{"1:2", "3:4"})
	if err != nil {
		t.Fatalf("buildTextsAPIURL() error = %v", err)
	}

	expected := "https://api.figma.com/v1/files/file123/nodes?ids=1%3A2%2C3%3A4"
	if got != expected {
		t.Fatalf("buildTextsAPIURL() = %v, expected %v", got, expected)
	}
}

func TestBuildTextsAPIURLWithoutNodeIDs(t *testing.T) {
	got, err := buildTextsAPIURL("file123", nil)
	if err != nil {
		t.Fatalf("buildTextsAPIURL() error = %v", err)
	}

	expected := "https://api.figma.com/v1/files/file123"
	if got != expected {
		t.Fatalf("buildTextsAPIURL() = %v, expected %v", got, expected)
	}
}
