package cmd

import "testing"

func TestBuildExportAPIURL(t *testing.T) {
	got, err := buildExportAPIURL("file123", []string{"1:2"}, "png")
	if err != nil {
		t.Fatalf("buildExportAPIURL() error = %v", err)
	}

	expected := "https://api.figma.com/v1/images/file123?format=png&ids=1%3A2"
	if got != expected {
		t.Fatalf("buildExportAPIURL() = %v, expected %v", got, expected)
	}
}

func TestBuildExportAPIURLRequiresNodeID(t *testing.T) {
	_, err := buildExportAPIURL("file123", nil, "png")
	if err == nil {
		t.Fatal("buildExportAPIURL() error = nil, expected error")
	}
}

func TestBuildExportAPIURLRequiresFormat(t *testing.T) {
	_, err := buildExportAPIURL("file123", []string{"1:2"}, "")
	if err == nil {
		t.Fatal("buildExportAPIURL() error = nil, expected error")
	}
}

func TestDefaultExportOutputPath(t *testing.T) {
	got := defaultExportOutputPath("file123", "1:2", "png")
	expected := "file123_1-2.png"
	if got != expected {
		t.Fatalf("defaultExportOutputPath() = %v, expected %v", got, expected)
	}
}
