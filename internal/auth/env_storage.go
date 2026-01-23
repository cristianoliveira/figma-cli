package auth

import (
	"context"
	"fmt"
	"os"
)

// envStorage implements Storage using environment variables.
// This is a fallback and only supports PAT tokens.
type envStorage struct{}

// NewEnvStorage creates a new environment variable-based token storage.
func NewEnvStorage() Storage {
	return &envStorage{}
}

// Store is not supported for environment variables (read-only).
func (s *envStorage) Store(ctx context.Context, token *Token) error {
	// We could set environment variable for the current process,
	// but that's not persistent. For simplicity, we treat env storage as read-only.
	return fmt.Errorf("cannot store token in environment variable storage (read-only)")
}

// Retrieve reads a PAT token from the FIGMA_ACCESS_TOKEN environment variable.
func (s *envStorage) Retrieve(ctx context.Context) (*Token, error) {
	token := os.Getenv("FIGMA_ACCESS_TOKEN")
	if token == "" {
		return nil, nil
	}
	return &Token{
		Type:         TokenTypePAT,
		AccessToken:  token,
		RefreshToken: "",
		ExpiresAt:    nil,
	}, nil
}

// Clear unsets the environment variable for the current process only.
func (s *envStorage) Clear(ctx context.Context) error {
	os.Unsetenv("FIGMA_ACCESS_TOKEN")
	return nil
}
