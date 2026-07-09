package cli

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/env"
)

func TestLoadClient_FromEnv(t *testing.T) {
	t.Setenv("FIGMA_ACCESS_TOKEN", "secret-token")

	client, err := LoadClient()

	if err != nil {
		t.Fatalf("LoadClient() error = %v", err)
	}
	if client.Token != "secret-token" {
		t.Errorf("LoadClient() token = %q, want secret-token", client.Token)
	}
}

func TestLoadClient_MissingToken(t *testing.T) {
	t.Setenv("FIGMA_ACCESS_TOKEN", "")

	_, err := LoadClient()

	if err == nil {
		t.Fatal("LoadClient() expected error for missing token, got nil")
	}
	if _, ok := err.(*env.ErrTokenNotSet); !ok {
		t.Errorf("LoadClient() error = %T, want *env.ErrTokenNotSet", err)
	}
}
