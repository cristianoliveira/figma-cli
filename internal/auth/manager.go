package auth

import (
	"context"
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/config"
)

// Manager handles authentication tokens, including storage, retrieval, and refresh.
type Manager struct {
	cfg     *config.Config
	storage Storage
}

// NewManager creates a new token manager.
func NewManager(cfg *config.Config) (*Manager, error) {
	storage, err := NewKeyringStorage()
	if err != nil {
		// Fallback to environment variable storage
		storage = NewEnvStorage()
	}
	return &Manager{
		cfg:     cfg,
		storage: storage,
	}, nil
}

// GetToken returns a valid token, refreshing if necessary.
func (m *Manager) GetToken(ctx context.Context) (*Token, error) {
	// First try to retrieve from secure storage
	token, err := m.storage.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve token: %w", err)
	}
	if token == nil {
		// No token stored; maybe we have a PAT in config (legacy)
		if m.cfg.Token != "" && m.cfg.TokenType == "pat" {
			token = &Token{
				Type:         TokenTypePAT,
				AccessToken:  m.cfg.Token,
				RefreshToken: "",
				ExpiresAt:    nil,
			}
			// Optionally store in keyring for future use
			_ = m.storage.Store(ctx, token)
		} else {
			return nil, fmt.Errorf("no authentication token found")
		}
	}

	// Check if token needs refresh
	if token.NeedsRefresh() {
		newToken, err := m.refreshToken(ctx, token)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		token = newToken
		// Store the refreshed token
		if err := m.storage.Store(ctx, token); err != nil {
			// Log error but continue with new token
		}
	}

	return token, nil
}

// StoreToken saves a token to storage.
func (m *Manager) StoreToken(ctx context.Context, token *Token) error {
	return m.storage.Store(ctx, token)
}

// ClearToken removes the stored token.
func (m *Manager) ClearToken(ctx context.Context) error {
	return m.storage.Clear(ctx)
}

// refreshToken uses the refresh token to obtain a new access token.
func (m *Manager) refreshToken(ctx context.Context, token *Token) (*Token, error) {
	if token.Type != TokenTypeOAuth || token.RefreshToken == "" {
		return nil, fmt.Errorf("cannot refresh token: missing refresh token or wrong type")
	}
	// TODO: implement OAuth token refresh using Figma API
	// For now, return error indicating need to re-authenticate
	return nil, fmt.Errorf("token refresh not yet implemented")
}

// Authenticator returns an http.RoundTripper or similar that adds authentication headers.
// We'll implement later.
func (m *Manager) Authenticator() {}
