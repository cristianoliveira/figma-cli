package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/config"
)

// configStorage implements Storage using the config file.
type configStorage struct {
	cfg *config.Config
}

// NewConfigStorage creates a new config file-based token storage.
func NewConfigStorage(cfg *config.Config) Storage {
	return &configStorage{cfg: cfg}
}

// Store saves a token to the config file.
func (s *configStorage) Store(ctx context.Context, token *Token) error {
	// Serialize token to JSON
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}
	// Store in config
	s.cfg.AuthToken = string(data)
	// Also update Token and TokenType for backward compatibility
	s.cfg.Token = token.AccessToken
	s.cfg.TokenType = string(token.Type)
	// Save config to file
	if err := s.cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}

// Retrieve reads a token from the config file.
func (s *configStorage) Retrieve(ctx context.Context) (*Token, error) {
	// First try AuthToken (JSON)
	if s.cfg.AuthToken != "" {
		var token Token
		if err := json.Unmarshal([]byte(s.cfg.AuthToken), &token); err == nil {
			// Successfully unmarshaled
			return &token, nil
		}
		// If unmarshal fails, fall back to legacy fields
	}
	// Legacy PAT token stored in Token field
	if s.cfg.Token != "" {
		tokenType := TokenTypePAT
		if s.cfg.TokenType == "oauth" {
			tokenType = TokenTypeOAuth
		}
		return &Token{
			Type:         tokenType,
			AccessToken:  s.cfg.Token,
			RefreshToken: "",
			ExpiresAt:    nil,
			Scopes:       nil,
		}, nil
	}
	return nil, nil
}

// Clear removes the token from the config file.
func (s *configStorage) Clear(ctx context.Context) error {
	s.cfg.AuthToken = ""
	s.cfg.Token = ""
	s.cfg.TokenType = "pat"
	if err := s.cfg.Save(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}
