package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const (
	// DefaultOAuthAuthURL is the Figma OAuth authorization endpoint.
	DefaultOAuthAuthURL = "https://www.figma.com/oauth"
	// DefaultOAuthTokenURL is the Figma OAuth token endpoint.
	DefaultOAuthTokenURL = "https://www.figma.com/api/oauth/token"
	// DefaultRedirectURL is the local redirect URL for CLI OAuth flow.
	DefaultRedirectURL = "http://localhost:8080/callback"
)

// OAuthConfig holds configuration for OAuth authentication.
type OAuthConfig struct {
	ClientID    string
	Scopes      []string
	RedirectURL string
}

// OAuthManager handles OAuth 2.0 authentication flow.
type OAuthManager struct {
	config *OAuthConfig
	oauth  *oauth2.Config
}

// NewOAuthManager creates a new OAuthManager with the given config.
func NewOAuthManager(cfg *OAuthConfig) *OAuthManager {
	oauth := &oauth2.Config{
		ClientID:    cfg.ClientID,
		Scopes:      cfg.Scopes,
		RedirectURL: cfg.RedirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  DefaultOAuthAuthURL,
			TokenURL: DefaultOAuthTokenURL,
		},
	}
	return &OAuthManager{
		config: cfg,
		oauth:  oauth,
	}
}

// GeneratePKCE generates a code verifier and code challenge using S256 method.
func GeneratePKCE() (verifier, challenge string, err error) {
	// Generate random bytes for verifier (length between 43 and 128)
	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}
	verifier = base64.RawURLEncoding.EncodeToString(bytes)
	// Compute SHA256 hash and encode as base64 URL without padding
	hash := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(hash[:])
	return verifier, challenge, nil
}

// AuthCodeURL returns the URL to redirect the user to for authorization.
func (m *OAuthManager) AuthCodeURL(state, codeChallenge string) string {
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	}
	return m.oauth.AuthCodeURL(state, opts...)
}

// ExchangeCode exchanges an authorization code for an access token.
func (m *OAuthManager) ExchangeCode(ctx context.Context, code, codeVerifier string) (*oauth2.Token, error) {
	opts := []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	}
	return m.oauth.Exchange(ctx, code, opts...)
}

// RefreshToken refreshes an access token using a refresh token.
func (m *OAuthManager) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	token := &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-1 * time.Hour), // force expired
	}
	src := m.oauth.TokenSource(ctx, token)
	newToken, err := src.Token()
	if err != nil {
		return nil, err
	}
	return newToken, nil
}

// TokenToAuthToken converts an oauth2.Token to our internal Token representation.

// parseScopes extracts scopes from oauth2.Token extra fields.
func parseScopes(oauthToken *oauth2.Token) []string {
	if scope, ok := oauthToken.Extra("scope").(string); ok && scope != "" {
		return strings.Split(scope, ",")
	}
	return nil
}
func TokenToAuthToken(oauthToken *oauth2.Token) *Token {
	token := &Token{
		Type:         TokenTypeOAuth,
		AccessToken:  oauthToken.AccessToken,
		RefreshToken: oauthToken.RefreshToken,
		Scopes:       parseScopes(oauthToken),
	}
	if !oauthToken.Expiry.IsZero() {
		token.ExpiresAt = &oauthToken.Expiry
	}
	return token
}

// AuthTokenToOAuthToken converts our Token to oauth2.Token.
func AuthTokenToOAuthToken(token *Token) *oauth2.Token {
	var expiry time.Time
	if token.ExpiresAt != nil {
		expiry = *token.ExpiresAt
	}
	return &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		Expiry:       expiry,
	}
}

// StartLocalServer starts a local HTTP server to handle OAuth callback.
// Returns the URL to redirect the user to and a channel that will receive the authorization code.
func StartLocalServer() (redirectURL string, codeChan chan string, err error) {
	// TODO: implement a local server on a random port
	return "", nil, fmt.Errorf("not implemented")
}
