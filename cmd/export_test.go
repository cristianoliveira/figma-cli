package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
)

func TestBuildExportURL(t *testing.T) {
	got, err := figma.BuildExportURL("file123", []string{"1:2"}, "png")
	if err != nil {
		t.Fatalf("BuildExportURL() error = %v", err)
	}

	expected := "https://api.figma.com/v1/images/file123?format=png&ids=1%3A2"
	if got != expected {
		t.Fatalf("BuildExportURL() = %v, expected %v", got, expected)
	}
}

func TestBuildExportURLRequiresNodeID(t *testing.T) {
	_, err := figma.BuildExportURL("file123", nil, "png")
	if err == nil {
		t.Fatal("BuildExportURL() error = nil, expected error")
	}
}

func TestBuildExportURLRequiresFormat(t *testing.T) {
	_, err := figma.BuildExportURL("file123", []string{"1:2"}, "")
	if err == nil {
		t.Fatal("BuildExportURL() error = nil, expected error")
	}
}
