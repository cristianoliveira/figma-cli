package cmd

import "testing"

func TestBuildVersionsAPIURL(t *testing.T) {
	got := buildVersionsAPIURL("grnVU2vAihHXwYgHryu2xE")
	expected := "https://api.figma.com/v1/files/grnVU2vAihHXwYgHryu2xE/versions"

	if got != expected {
		t.Errorf("buildVersionsAPIURL() = %v, expected %v", got, expected)
	}
}
