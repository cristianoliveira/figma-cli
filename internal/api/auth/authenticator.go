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
