package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
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
// It uses the provided OAuthConfig, state, and codeChallenge to generate the authorization URL.
// Returns the authorization URL to open in the browser and a channel that will receive the authorization code.
// The caller should open the authorization URL in the user's browser and wait for the code on the channel.
// The server will automatically shut down after receiving the callback or after a timeout.
func StartLocalServer(cfg *OAuthConfig, state, codeChallenge string) (authURL string, codeChan <-chan string, err error) {
	if cfg == nil {
		return "", nil, errors.New("OAuthConfig is required")
	}
	redirectURL := cfg.RedirectURL
	if redirectURL == "" {
		redirectURL = DefaultRedirectURL
	}
	parsed, err := url.Parse(redirectURL)
	if err != nil {
		return "", nil, fmt.Errorf("invalid redirect URL: %w", err)
	}
	// Ensure redirect URL includes a port
	if parsed.Port() == "" {
		defaultPort := "8080"
		if parsed.Scheme == "https" {
			defaultPort = "443"
		}
		parsed.Host = net.JoinHostPort(parsed.Hostname(), defaultPort)
		redirectURL = parsed.String()
		cfg.RedirectURL = redirectURL
	}
	path := parsed.Path
	if path == "" {
		path = "/"
	}
	// Create OAuth manager
	mgr := NewOAuthManager(cfg)
	authURL = mgr.AuthCodeURL(state, codeChallenge)

	// Channel for the authorization code (buffered size 1)
	ch := make(chan string, 1)
	// Channel to signal server shutdown
	shutdownCh := make(chan struct{})
	var server *http.Server
	var mu sync.Mutex
	serverExited := false

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			http.NotFound(w, r)
			return
		}
		code := r.URL.Query().Get("code")
		errorMsg := r.URL.Query().Get("error")
		if errorMsg != "" {
			errorDesc := r.URL.Query().Get("error_description")
			http.Error(w, fmt.Sprintf("OAuth error: %s - %s", errorMsg, errorDesc), http.StatusBadRequest)
			// Still shut down server
			go func() {
				mu.Lock()
				if !serverExited {
					serverExited = true
					mu.Unlock()
					_ = server.Shutdown(context.Background())
				} else {
					mu.Unlock()
				}
			}()
			return
		}
		if code == "" {
			http.Error(w, "Missing authorization code", http.StatusBadRequest)
			go func() {
				mu.Lock()
				if !serverExited {
					serverExited = true
					mu.Unlock()
					_ = server.Shutdown(context.Background())
				} else {
					mu.Unlock()
				}
			}()
			return
		}
		// Send code to channel (non-blocking)
		select {
		case ch <- code:
		default:
		}
		// Respond with success page
		w.Header().Set("Content-Type", "text/html")
		_, _ = io.WriteString(w, `<html><body><h1>Authentication successful!</h1><p>You can close this window and return to the CLI.</p></body></html>`)
		// Shutdown server after a short delay to allow response to be sent
		go func() {
			mu.Lock()
			if !serverExited {
				serverExited = true
				mu.Unlock()
				_ = server.Shutdown(context.Background())
			} else {
				mu.Unlock()
			}
		}()
	})

	server = &http.Server{
		Addr:    parsed.Host,
		Handler: handler,
	}
	// Start server in goroutine
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return "", nil, fmt.Errorf("failed to start local server: %w", err)
	}
	go func() {
		_ = server.Serve(listener)
		close(shutdownCh)
	}()

	// Set up timeout shutdown
	timeout := 5 * time.Minute
	timeoutTimer := time.AfterFunc(timeout, func() {
		mu.Lock()
		if !serverExited {
			serverExited = true
			mu.Unlock()
			_ = server.Shutdown(context.Background())
		} else {
			mu.Unlock()
		}
	})
	// Cleanup function to ensure server stops
	cleanup := func() {
		timeoutTimer.Stop()
		mu.Lock()
		if !serverExited {
			serverExited = true
			mu.Unlock()
			_ = server.Shutdown(context.Background())
			<-shutdownCh
		} else {
			mu.Unlock()
		}
	}
	// Return authURL and channel; caller must call cleanup after receiving code or on error
	// We'll attach cleanup to a goroutine that waits for either code or context cancellation?
	// Simpler: return a wrapper channel that triggers cleanup after receiving code.
	// We'll create a wrapper channel that forwards code and then cleans up.
	wrappedCh := make(chan string, 1)
	go func() {
		select {
		case code := <-ch:
			cleanup()
			wrappedCh <- code
			close(wrappedCh)
		case <-shutdownCh:
			cleanup()
			close(wrappedCh)
		}
	}()
	return authURL, wrappedCh, nil
}
