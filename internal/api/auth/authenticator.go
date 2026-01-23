package auth

import (
	"net/http"
)

// Authenticator defines an interface for authenticating HTTP requests.
type Authenticator interface {
	// Authenticate adds authentication credentials to the request.
	Authenticate(req *http.Request) error
}

// TokenAuthenticator adds a Bearer token to the Authorization header.
type TokenAuthenticator struct {
	token string
}

// NewTokenAuthenticator creates a new TokenAuthenticator with the given token.
func NewTokenAuthenticator(token string) *TokenAuthenticator {
	return &TokenAuthenticator{token: token}
}

// Authenticate implements Authenticator.
func (a *TokenAuthenticator) Authenticate(req *http.Request) error {
	if a.token != "" {
		req.Header.Set("Authorization", "Bearer "+a.token)
	}
	return nil
}

// FigmaAuthenticator adds the appropriate authentication header based on token type.
type FigmaAuthenticator struct {
	tokenType string
	token     string
}

// NewFigmaAuthenticator creates a new FigmaAuthenticator.
func NewFigmaAuthenticator(tokenType, token string) *FigmaAuthenticator {
	return &FigmaAuthenticator{
		tokenType: tokenType,
		token:     token,
	}
}

// Authenticate implements Authenticator.
func (a *FigmaAuthenticator) Authenticate(req *http.Request) error {
	if a.token == "" {
		return nil
	}
	switch a.tokenType {
	case "pat":
		req.Header.Set("X-Figma-Token", a.token)
	case "oauth":
		req.Header.Set("Authorization", "Bearer "+a.token)
	default:
		// Default to PAT for backward compatibility
		req.Header.Set("X-Figma-Token", a.token)
	}
	return nil
}
