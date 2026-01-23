package auth

import (
	"context"
	"time"
)

// TokenType represents the type of authentication token.
type TokenType string

const (
	// TokenTypePAT is a Personal Access Token.
	TokenTypePAT TokenType = "pat"
	// TokenTypeOAuth is an OAuth 2.0 access token.
	TokenTypeOAuth TokenType = "oauth"
)

// Token represents an authentication token for the Figma API.
type Token struct {
	// Type of token (PAT or OAuth).
	Type TokenType `json:"type"`
	// AccessToken is the token value.
	AccessToken string `json:"access_token"`
	// RefreshToken is used to refresh OAuth tokens (optional).
	RefreshToken string `json:"refresh_token,omitempty"`
	// ExpiresAt is the expiration time of the access token (optional).
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	// Scopes granted for OAuth tokens (optional).
	Scopes []string `json:"scopes,omitempty"`
}

// IsExpired returns true if the token has an expiration time and it has passed.
func (t *Token) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return t.ExpiresAt.Before(time.Now())
}

// NeedsRefresh returns true if the token is an OAuth token that can be refreshed.
func (t *Token) NeedsRefresh() bool {
	return t.Type == TokenTypeOAuth && t.RefreshToken != "" && t.IsExpired()
}

// Storage defines an interface for securely storing authentication tokens.
type Storage interface {
	// Store saves a token, overwriting any existing token.
	Store(ctx context.Context, token *Token) error
	// Retrieve returns the stored token, or nil if none exists.
	Retrieve(ctx context.Context) (*Token, error)
	// Clear removes any stored token.
	Clear(ctx context.Context) error
}
