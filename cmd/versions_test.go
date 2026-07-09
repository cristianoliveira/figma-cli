package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/figma"
)

func TestBuildVersionsURL(t *testing.T) {
	got := figma.BuildVersionsURL("grnVU2vAihHXwYgHryu2xE")
	expected := "https://api.figma.com/v1/files/grnVU2vAihHXwYgHryu2xE/versions"

	if got != expected {
		t.Errorf("BuildVersionsURL() = %v, expected %v", got, expected)
	}
}
