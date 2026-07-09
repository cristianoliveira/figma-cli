package figma

import "testing"

func TestBuildNodesURLWithNodeIDs(t *testing.T) {
	got, err := BuildNodesURL("file123", []string{"1:2", "3:4"})
	if err != nil {
		t.Fatalf("BuildNodesURL() error = %v", err)
	}
	expected := "https://api.figma.com/v1/files/file123/nodes?ids=1%3A2%2C3%3A4"
	if got != expected {
		t.Fatalf("BuildNodesURL() = %v, expected %v", got, expected)
	}
}

func TestBuildFileURLWithoutNodeIDs(t *testing.T) {
	got, err := BuildFileURL("file123", nil, "", "")
	if err != nil {
		t.Fatalf("BuildFileURL() error = %v", err)
	}
	expected := "https://api.figma.com/v1/files/file123"
	if got != expected {
		t.Fatalf("BuildFileURL() = %v, expected %v", got, expected)
	}
}

func TestBuildVersionsURL(t *testing.T) {
	got := BuildVersionsURL("grnVU2vAihHXwYgHryu2xE")
	expected := "https://api.figma.com/v1/files/grnVU2vAihHXwYgHryu2xE/versions"
	if got != expected {
		t.Errorf("BuildVersionsURL() = %v, expected %v", got, expected)
	}
}

func TestBuildExportURL(t *testing.T) {
	got, err := BuildExportURL("file123", []string{"1:2"}, "png")
	if err != nil {
		t.Fatalf("BuildExportURL() error = %v", err)
	}
	expected := "https://api.figma.com/v1/images/file123?format=png&ids=1%3A2"
	if got != expected {
		t.Fatalf("BuildExportURL() = %v, expected %v", got, expected)
	}
}

func TestBuildExportURLRequiresNodeID(t *testing.T) {
	if _, err := BuildExportURL("file123", nil, "png"); err == nil {
		t.Fatal("BuildExportURL() error = nil, expected error")
	}
}

func TestBuildExportURLRequiresFormat(t *testing.T) {
	if _, err := BuildExportURL("file123", []string{"1:2"}, ""); err == nil {
		t.Fatal("BuildExportURL() error = nil, expected error")
	}
}

func TestBuildStylesURL(t *testing.T) {
	got := BuildStylesURL("file123")
	expected := "https://api.figma.com/v1/files/file123/styles"
	if got != expected {
		t.Errorf("BuildStylesURL() = %v, expected %v", got, expected)
	}
}

func TestBuildVariablesURL(t *testing.T) {
	got := BuildVariablesURL("file123")
	expected := "https://api.figma.com/v1/files/file123/variables/local"
	if got != expected {
		t.Errorf("BuildVariablesURL() = %v, expected %v", got, expected)
	}
}
